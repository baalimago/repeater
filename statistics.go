package main

import (
	"fmt"
	"math"
	"time"
)

type Result struct {
	WorkerID             int           `json:"workerID"`
	Idx                  int           `json:"taskIdx"`
	Runtime              time.Duration `json:"runtime"`
	RuntimeHumanReadable string        `json:"runtimeHumanReadable"`
	Output               string        `json:"output"`
	IsError              bool          `json:"isError"`
}

type statistics struct {
	max     Result
	min     Result
	total   time.Duration
	runtime time.Duration
	average time.Duration
	stdDev  time.Duration
	Results []Result `json:"results"`
}

// Write implements io.Writer to get the output of the command for
// both out and err
func (r *Result) Write(p []byte) (n int, err error) {
	r.Output += string(p)
	return len(p), nil
}

func (c *configuredOper) calcStats() statistics {
	tot := time.Duration(0)
	n := len(c.results)
	minDur := time.Duration(9223372036854775807)
	maxDur := time.Duration(-9223372036854775808)
	var min, max Result
	for _, r := range c.results {
		if r.Runtime < minDur {
			min = r
			minDur = r.Runtime
		}
		if r.Runtime > maxDur {
			max = r
			maxDur = r.Runtime
		}
		tot += r.Runtime
	}

	avr := int64(tot) / int64(n)
	varSum := 0.0
	for _, x := range c.results {
		varSum += math.Pow((float64(x.Runtime) - float64(avr)), 2.0)
	}
	variance := varSum / float64(n)
	stdDeviation := time.Duration(int64(math.Sqrt(variance)))
	return statistics{
		runtime: c.runtime,
		min:     min,
		max:     max,
		total:   tot,
		average: time.Duration(avr),
		stdDev:  stdDeviation,
		Results: c.results,
	}
}

func (s *statistics) String() string {
	return fmt.Sprintf(`
Runtime: %s, Total routine work time: %v,
Average time per task: %v, Std deviation: %v
Max time, index: %v, time: %v
Min time, index: %v, time: %v`,
		s.runtime, s.total,
		s.average, s.stdDev,
		s.max.Idx, s.max.Runtime,
		s.min.Idx, s.min.Runtime)
}
