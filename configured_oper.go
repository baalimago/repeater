package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/internal/progress"
	"github.com/baalimago/repeater/pkg/filetools"
)

type configuredOper struct {
	am             int
	workers        int
	args           []string
	color          bool
	progress       progress.Mode
	progressFormat string
	output         output.Mode
	reportFile     *os.File
}

func (c configuredOper) String() string {
	return fmt.Sprintf(`am: %v
command: %v
workers: %v
color: %v
progress: %s
progress format: %q
output: %s
report file: %v`, c.am, c.args, c.workers, c.color, c.progress, c.progressFormat, c.output, c.reportFile.Name())
}

func (c configuredOper) printStatus(out io.Writer, status, msg string, color colorCode) {
	if c.color {
		status = coloredMessage(color, status)
	}
	fmt.Fprintf(out, "%v: %v", status, msg)
}

func (c configuredOper) printErr(msg string) {
	c.printStatus(os.Stderr, "error", msg, RED)
}

func (c configuredOper) printOK(msg string) {
	c.printStatus(os.Stdout, "ok", msg, GREEN)
}

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c configuredOper) run(ctx context.Context) statistics {
	progressStreams := make([]io.Writer, 0, 2)
	switch c.progress {
	case progress.STDOUT:
		progressStreams = append(progressStreams, os.Stdout)
	case progress.REPORT_FILE:
		progressStreams = append(progressStreams, c.reportFile)
	case progress.BOTH:
		progressStreams = append(progressStreams, os.Stdout)
		progressStreams = append(progressStreams, c.reportFile)
	}

	defer filetools.WriteStringIfPossible("\n", progressStreams)

	ret := statistics{
		resMu: &sync.Mutex{},
	}
	minMaxMu := &sync.Mutex{}
	min := time.Duration(math.MaxInt64)
	max := time.Duration(-1)

	workChan := make(chan int)
	workCtx, workCtxCancel := context.WithCancel(ctx)
	workersDone := sync.WaitGroup{}
	workersDone.Add(c.workers)
	for i := 0; i < c.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case <-workCtx.Done():
					workersDone.Done()
					return
				case taskIdx := <-workChan:
					res := result{
						idx: taskIdx,
					}
					do := exec.Command(c.args[0], c.args[1:]...)
					switch c.output {
					case output.STDOUT:
						do.Stdout = io.MultiWriter(os.Stdout, &res)
						do.Stderr = io.MultiWriter(os.Stderr, &res)
					case output.HIDDEN:
						do.Stdout = &res
						do.Stderr = &res
					case output.REPORT_FILE:
						do.Stdout = io.MultiWriter(c.reportFile, &res)
						do.Stderr = io.MultiWriter(c.reportFile, &res)
					case output.BOTH:
						do.Stdout = io.MultiWriter(c.reportFile, os.Stdout, &res)
						do.Stderr = io.MultiWriter(c.reportFile, os.Stderr, &res)
					}

					t0 := time.Now()
					err := do.Run()
					res.runtime = time.Since(t0)

					minMaxMu.Lock()
					if res.runtime > max {
						ret.max = res
						max = res.runtime
					}
					if res.runtime < min {
						ret.min = res
						min = res.runtime
					}
					minMaxMu.Unlock()

					ret.resMu.Lock()
					ret.res = append(ret.res, res)
					ret.resMu.Unlock()
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) {
						c.printErr(fmt.Sprintf("unexpected error encountered for: %v, aborting operations: %v", workerID, exitErr))
						os.Exit(1)
					}
					filetools.WriteStringIfPossible(fmt.Sprintf(c.progressFormat, taskIdx+1, c.am), progressStreams)
				}
			}
		}(i)
	}

	confOperStart := time.Now()
	for i := 0; i < c.am; i++ {
		if ctx.Err() != nil {
			c.printErr(fmt.Sprintf("context error: %v", ctx.Err()))
			os.Exit(1)
		}
		workChan <- i
	}
	workCtxCancel()
	workersDone.Wait()
	ret.totalTime = time.Since(confOperStart)

	return ret
}
