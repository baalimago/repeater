package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/baalimago/repeater/pkg/filetools"
)

func (c *configuredOper) replaceIncrement(args []string, i int) []string {
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

func (c *configuredOper) doWork(workerID, taskIdx int) Result {
	res := Result{
		WorkerID: workerID,
		Idx:      taskIdx,
	}
	args := c.replaceIncrement(c.args[1:], taskIdx)
	do := exec.Command(c.args[0], args...)
	c.setupOutputStreams(do, &res)
	t0 := time.Now()
	err := do.Run()
	timeSpent := time.Since(t0)
	res.Runtime = timeSpent
	res.RuntimeHumanReadable = fmt.Sprintf("%s", timeSpent)
	if err != nil {
		res.Output = err.Error()
		res.IsError = true
	}
	return res
}

// setupWorkes by starting one go routine for each worker that listens to workChan
func (c *configuredOper) setupWorkers(workCtx context.Context, workChan chan int, resultChan chan Result) {
	for i := 0; i < c.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case <-workCtx.Done():
					return
				case taskIdx := <-workChan:
					res := c.doWork(workerID, taskIdx)
					resultChan <- res
				}
			}
		}(i)
	}
}

// runDelegator in a blocking manner, will append data to stats
func (c *configuredOper) runDelegator(ctx context.Context, resultChan chan Result, workChan chan int, progressStreams []io.Writer) {
	i := 0
WORK_DELEGATOR:
	for {
		if ctx.Err() != nil {
			printErr(fmt.Sprintf("context error: %v", ctx.Err()))
			os.Exit(1)
		}
		select {
		case res := <-resultChan:
			if strings.Contains(res.Output, "ERROR:") {
				printErr(fmt.Sprintf("worker: %v received %v\n", res.WorkerID, res.Output))
			} else {
				// This is threadsafe snce only the delegator adds results
				c.results = append(c.results, res)
				filetools.WriteStringIfPossible(fmt.Sprintf(c.progressFormat, i, c.am), progressStreams)
			}

			if len(c.results) == c.am {
				break WORK_DELEGATOR
			}
		case workChan <- i:
			if i == c.am {
				continue
			}
			i++
		}
	}
}

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c *configuredOper) run(ctx context.Context) statistics {
	progressStreams := c.setupProgressStreams()
	defer filetools.WriteStringIfPossible("\n", progressStreams)

	workChan := make(chan int)
	// Buffer the channel for each worker, so that the workers may leave a result and then quit
	resultChan := make(chan Result, c.am)
	workCtx, workCtxCancel := context.WithCancel(ctx)
	if c.workers < 1 {
		c.workers = 1
	}
	c.setupWorkers(workCtx, workChan, resultChan)

	confOperStart := time.Now()
	c.runDelegator(ctx, resultChan, workChan, progressStreams)
	workCtxCancel()
	c.runtime = time.Since(confOperStart)

	return c.calcStats()
}
