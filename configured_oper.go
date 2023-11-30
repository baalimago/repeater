package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/general"
	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/pkg/filetools"
)

type configuredOper struct {
	am             int
	workers        int
	args           []string
	color          bool
	progress       output.Mode
	progressFormat string
	output         output.Mode
	reportFile     *os.File
	reportFileMu   *sync.Mutex
	increment      bool
}

// getFile by checking if it exists and querying user about how to treat the file
func getFile(s string) *os.File {
	f, err := os.Create(s)
	if err != nil {
		panic(fmt.Sprintf("not good: %v", err))
	}
	return f
}

type incrementConfigError struct {
	args []string
}

func (ice incrementConfigError) Error() string {
	return fmt.Sprintf("increment is true, but args: %v, does not contain string 'INC'", ice.args)
}

func New(am, workers int,
	args []string,
	color bool,
	pMode output.Mode,
	progressFormat string,
	oMode output.Mode,
	reportFile string,
	increment bool,
) (configuredOper, error) {
	shouldHaveReportFile := pMode == output.BOTH || pMode == output.REPORT_FILE ||
		oMode == output.BOTH || oMode == output.REPORT_FILE

	if shouldHaveReportFile && reportFile == "" {
		return configuredOper{}, fmt.Errorf("progress: %v, or output: %v, requires a report file, but none is specified", pMode, oMode)
	}

	if increment && !slices.Contains(args, "INC") {
		return configuredOper{
			color: *colorFlag,
		}, incrementConfigError{args: args}
	}

	if workers >= am {
		return configuredOper{
			color: *colorFlag,
		}, fmt.Errorf("please use less workers than repetitions. Am workes: %v, am repetitions: %v", workers, am)
	}

	return configuredOper{
		am:             am,
		workers:        workers,
		color:          color,
		args:           args,
		progress:       pMode,
		progressFormat: progressFormat,
		output:         oMode,
		reportFile:     getFile(reportFile),
		reportFileMu:   &sync.Mutex{},
		increment:      increment,
	}, nil
}

func (c configuredOper) String() string {
	reportFileName := "HIDDEN"
	if c.reportFile != nil {
		reportFileName = c.reportFile.Name()
	}
	return fmt.Sprintf(`am: %v
command: %v
increment: %v
workers: %v
color: %v
progress: %s
progress format: %q
output: %s
report file: %v`, c.am, c.args, c.increment, c.workers, c.color, c.progress, c.progressFormat, c.output, reportFileName)
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

func (c configuredOper) setupProgressStreams() []io.Writer {
	progressStreams := make([]io.Writer, 0, 2)
	switch c.progress {
	case output.STDOUT:
		progressStreams = append(progressStreams, os.Stdout)
	case output.REPORT_FILE:
		progressStreams = append(progressStreams, c.reportFile)
	case output.BOTH:
		progressStreams = append(progressStreams, os.Stdout)
		progressStreams = append(progressStreams, c.reportFile)
	}
	return progressStreams
}

func (c *configuredOper) setupOutputStreams(toDo *exec.Cmd, res *result) {
	switch c.output {
	case output.STDOUT:
		toDo.Stdout = io.MultiWriter(os.Stdout, res)
		toDo.Stderr = io.MultiWriter(os.Stderr, res)
	case output.HIDDEN:
		toDo.Stdout = res
		toDo.Stderr = res
	case output.REPORT_FILE:
		toDo.Stdout = io.MultiWriter(c.reportFile, res)
		toDo.Stderr = io.MultiWriter(c.reportFile, res)
	case output.BOTH:
		toDo.Stdout = io.MultiWriter(c.reportFile, os.Stdout, res)
		toDo.Stderr = io.MultiWriter(c.reportFile, os.Stderr, res)
	}
}

func (c configuredOper) replaceIncrement(args []string, i int) []string {
	if c.increment {
		var newArgs []string
		for _, arg := range args {
			if arg == "INC" {
				arg = fmt.Sprintf("%v", i)
			}
			newArgs = append(newArgs, arg)
		}
		return newArgs
	}
	return args
}

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c configuredOper) run(ctx context.Context) statistics {
	progressStreams := c.setupProgressStreams()
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
	if c.workers < 1 {
		c.workers = 1
	}
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
					args := c.replaceIncrement(c.args[1:], taskIdx)
					do := exec.Command(c.args[0], args...)
					c.setupOutputStreams(do, &res)
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
				// Used for debug
				// c.printOK(fmt.Sprintf("worker: %v exited\n", res.workerID))
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
