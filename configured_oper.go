package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/threadsafe"
	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/pkg/filetools"
)

const incrementPlaceholder = "INC"

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
	reportFileMode string
	increment      bool
}

type userQuitError string

func (uqe userQuitError) Error() string {
	return string(uqe)
}

const UserQuitError userQuitError = "user quit"

// getFile a file. if one already exists, either consult the fileMode string, or query
// user how the file should be treated
func (co *configuredOper) getFile(s, fileMode string) (*os.File, error) {
	if s == "" {
		return nil, nil
	}

	if _, err := os.Stat(s); !errors.Is(err, os.ErrNotExist) {
		userResp := fileMode
		if fileMode == "" {
			co.printWarn(fmt.Sprintf("file: \"%v\", already exists. Would you like to [t]runcate, [a]ppend or [q]uit? [tT/aA/qQ]: ", s))
			fmt.Scanln(&userResp)
		}
		cleanedUserResp := strings.ToLower(strings.TrimSpace(userResp))
		co.reportFileMode = cleanedUserResp
		switch cleanedUserResp {
		case "t":
			// NOOP, fallthrough to os.Create below
		case "a":
			return os.OpenFile(s, os.O_APPEND|os.O_RDWR, 0644)
		case "q":
			return nil, UserQuitError
		default:
			return nil, fmt.Errorf("unrecognized reply: \"%v\", valid options are [tT], [aA] or [qQ]", userResp)
		}
	}
	return os.Create(s)
}

type incrementConfigError struct {
	args []string
}

func (ice incrementConfigError) Error() string {
	return fmt.Sprintf("increment is true, but args: %v, does not contain string '%s'", ice.args, incrementPlaceholder)
}

func New(am, workers int,
	args []string,
	color bool,
	pMode output.Mode,
	progressFormat string,
	oMode output.Mode,
	reportFile string,
	reportFileMode string,
	increment bool,
) (configuredOper, error) {
	shouldHaveReportFile := pMode == output.BOTH || pMode == output.REPORT_FILE ||
		oMode == output.BOTH || oMode == output.REPORT_FILE

	if shouldHaveReportFile && reportFile == "" {
		return configuredOper{}, fmt.Errorf("progress: %v, or output: %v, requires a report file, but none is specified", pMode, oMode)
	}

	if increment && !containsIncrementPlaceholder(args) {
		return configuredOper{
			color: *colorFlag,
		}, incrementConfigError{args: args}
	}

	if workers > am {
		return configuredOper{
			color: *colorFlag,
		}, fmt.Errorf("please use less workers than repetitions. Am workers: %v, am repetitions: %v", workers, am)
	}

	c := configuredOper{
		am:             am,
		workers:        workers,
		color:          color,
		args:           args,
		progress:       pMode,
		progressFormat: progressFormat,
		output:         oMode,
		reportFileMu:   &sync.Mutex{},
		increment:      increment,
	}

	file, err := c.getFile(reportFile, reportFileMode)
	if err != nil {
		if errors.Is(err, UserQuitError) {
			return c, err
		}
		return c, fmt.Errorf("failed to get file: %v", err)
	}
	c.reportFile = file
	return c, nil
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
report file: %v
report file mode: %v`, c.am, c.args, c.increment, c.workers, c.color, c.progress, c.progressFormat, c.output, reportFileName, c.reportFileMode)
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

func (c configuredOper) printWarn(msg string) {
	c.printStatus(os.Stdout, "warning", msg, YELLOW)
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
			if strings.Contains(arg, incrementPlaceholder) {
				arg = strings.ReplaceAll(arg, incrementPlaceholder, strconv.Itoa(i))
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
			threadsafe.Write(runningWorkersMu, threadsafe.Read(runningWorkersMu, &runningWorkers)+1, &runningWorkers)
			defer func() {
				threadsafe.Write(runningWorkersMu, threadsafe.Read(runningWorkersMu, &runningWorkers)-1, &runningWorkers)
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

			if threadsafe.Read(runningWorkersMu, &runningWorkers) == 0 && len(resultChan) == 0 {
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

func containsIncrementPlaceholder(args []string) bool {
	for _, arg := range args {
		if strings.Contains(arg, incrementPlaceholder) {
			return true
		}
	}
	return false
}
