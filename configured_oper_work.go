package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/threadsafe"
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
	do.Stdout = &res
	do.Stderr = &res
	t0 := time.Now()
	err := do.Run()
	timeSpent := time.Since(t0)
	res.Runtime = timeSpent
	res.RuntimeHumanReadable = timeSpent.String()
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
					c.workerWg.Done()
					return
				case taskIdx := <-workChan:
					c.workPlanMu.Lock()
					workingWorkrs := c.workers - c.amIdleWorkers
					requestedTasks := c.am
					// The current amount of workers is enough to reach the requested
					// amount of tasks in parallel so kill off this worker to not overshoot
					// the amount of repetitions
					if workingWorkrs+c.amSuccess >= requestedTasks || (!c.retryOnFail && taskIdx >= requestedTasks) {
						c.workerWg.Done()
						c.workPlanMu.Unlock()
						return
					}
					c.amIdleWorkers--
					c.workPlanMu.Unlock()
					res := c.doWork(workerID, taskIdx)
					c.workPlanMu.Lock()
					c.amIdleWorkers++
					if !res.IsError {
						c.amSuccess++
					}
					c.workPlanMu.Unlock()
					resultChan <- res
				}
			}
		}(i)
	}
}

func (c *configuredOper) runDelegator(ctx context.Context, workChan chan int) error {
	i := 0
	for {
		if ctx.Err() != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case workChan <- i:
			amSuccess := threadsafe.Read(c.workPlanMu, &c.amSuccess)
			if amSuccess >= c.am {
				return nil
			} else {
				i++
			}
		}
	}
}

func (c *configuredOper) getTimeStrings(startedAt time.Time, amComplete int) (doneIn time.Duration, doneAt time.Time) {
	tot := len(c.results)
	amSuccess := tot - amComplete
	if amSuccess == 0 {
		return
	}
	avgPerSuccess := time.Since(startedAt) / time.Duration(amSuccess)
	doneAt = startedAt.Add(avgPerSuccess * time.Duration(c.am-amSuccess))
	doneIn = time.Until(doneAt)
	return
}

func (c *configuredOper) runResultCollector(ctx context.Context, resultChan chan Result, progressStreams []io.Writer) {
	t0 := time.Now()
	handleRes := func(res Result) int {
		c.writeOutput(&res)
		c.results = append(c.results, res)
		amFails := 0
		for _, r := range c.results {
			if r.IsError {
				amFails++
			}
		}

		tot := len(c.results)
		amSuccess := tot - amFails
		timeLeft, estCompletion := c.getTimeStrings(t0, amSuccess)
		filetools.WriteStringIfPossible(
			fmt.Sprintf(c.progressFormat,
				amSuccess, amFails, c.am,
				t0.Format(time.RFC3339), timeLeft.Seconds(), estCompletion.Format(time.RFC3339)),
			progressStreams)
		return amSuccess
	}

	emptyResChan := func() {
		for len(resultChan) > 0 {
			res := <-resultChan
			handleRes(res)
		}
	}
	for {
		if ctx.Err() != nil {
			emptyResChan()
			return
		}
		select {
		case <-ctx.Done():
			emptyResChan()
			return
		case res := <-resultChan:
			amSuccess := handleRes(res)
			c.workPlanMu.Lock()
			// Escape condition so that all results are collected
			if amSuccess >= c.am && c.amIdleWorkers == c.workers {
				c.workPlanMu.Unlock()
				return
			}
			c.workPlanMu.Unlock()
		}
	}
}

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c *configuredOper) run(rootCtx context.Context) statistics {
	ctx, ctxCancel := context.WithCancel(rootCtx)
	progressStreams := c.setupProgressStreams()
	workChan := make(chan int)
	// Buffer the channel for each worker, so that the workers may leave a result and then quit
	resultChan := make(chan Result, c.am)
	workCtx, workCtxCancel := context.WithCancel(ctx)
	if c.workers < 1 {
		c.workers = 1
	}
	c.setupWorkers(workCtx, workChan, resultChan)

	go func() {
		c.workerWg.Wait()
		ctxCancel()
	}()
	confOperStart := time.Now()
	go func() {
		err := c.runDelegator(ctx, workChan)
		if err != nil {
			printErr(fmt.Sprintf("work delegator error: %v", err))
			ctxCancel()
		} else {
			workCtxCancel()
		}
	}()
	c.runResultCollector(ctx, resultChan, progressStreams)
	c.runtime = time.Since(confOperStart)

	return c.calcStats()
}
