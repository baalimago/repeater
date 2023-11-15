package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/general"
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
	reportFileMu   *sync.Mutex
}

func (c configuredOper) String() string {
	reportFileName := "HIDDEN"
	if c.reportFile != nil {
		reportFileName = c.reportFile.Name()
	}
	return fmt.Sprintf(`am: %v
command: %v
workers: %v
color: %v
progress: %s
progress format: %q
output: %s
report file: %v`, c.am, c.args, c.workers, c.color, c.progress, c.progressFormat, c.output, reportFileName)
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

	ret := statistics{}
	minMaxMu := &sync.Mutex{}
	min := time.Duration(math.MaxInt64)
	max := time.Duration(-1)

	workChan := make(chan int)
	// Buffer the channel for each worker, so that the workers may leave a result and then quit
	resultChan := make(chan result, c.am)
	workCtx, workCtxCancel := context.WithCancel(ctx)
	runningWorkers := 0
	runningWorkersMu := &sync.Mutex{}
	for i := 0; i < c.workers; i++ {
		go func(workerID, amTasks int) {
			general.RaceSafeWrite(runningWorkersMu, general.RaceSafeRead(runningWorkersMu, &runningWorkers)+1, &runningWorkers)
			defer func() {
				general.RaceSafeWrite(runningWorkersMu, general.RaceSafeRead(runningWorkersMu, &runningWorkers)-1, &runningWorkers)
			}()
			for {
				select {
				case <-workCtx.Done():
					return
				case taskIdx := <-workChan:
					res := result{
						workerID: workerID,
						idx:      taskIdx,
					}
					// Kill the worker here if the sought after amount of repetitions has been performed
					if taskIdx == amTasks {
						// Send result here to unbock delegator
						resultChan <- res
						return
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

					if err != nil {
						res.output = fmt.Sprintf("ERROR: %v, check output for more info", err)
						resultChan <- res
						return
					} else {
						resultChan <- res
					}
				}
			}
		}(i, c.am)
	}

	confOperStart := time.Now()
	i := 0
WORK_DELEGATOR:
	for {
		if ctx.Err() != nil {
			c.printErr(fmt.Sprintf("context error: %v", ctx.Err()))
			os.Exit(1)
		}
		select {
		case res := <-resultChan:
			if strings.Contains(res.output, "ERROR:") {
				c.printErr(fmt.Sprintf("worker: %v received %v\n", res.workerID, res.output))
			} else if res.idx == c.am {
				c.printOK(fmt.Sprintf("worker: %v exited", res.workerID))
			} else {
				ret.res = append(ret.res, res)
				filetools.WriteStringIfPossible(fmt.Sprintf(c.progressFormat, i, c.am), progressStreams)
			}

			if general.RaceSafeRead(runningWorkersMu, &runningWorkers) == 0 && len(resultChan) == 0 {
				break WORK_DELEGATOR
			}
		case workChan <- i:
			if i == c.am {
				continue
			}
			i++
		}
	}
	workCtxCancel()
	ret.totalTime = time.Since(confOperStart)

	return ret
}
