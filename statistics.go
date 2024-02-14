package main

import (
	"fmt"
	"math"
	"time"
)

type result struct {
	workerID int
	idx      int
	runtime  time.Duration
	output   string
	isError  bool
}

type statistics struct {
	max     result
	min     result
	total   time.Duration
	average time.Duration
	stdDev  time.Duration
}

// Write implements io.Writer to get the output of the command for
// both out and err
func (r *result) Write(p []byte) (n int, err error) {
	r.output += string(p)
	return len(p), nil
}

func (c *configuredOper) calcStats() statistics {
	tot := time.Duration(0)
	n := len(c.results)
	minDur := time.Duration(9223372036854775807)
	maxDur := time.Duration(-9223372036854775808)
	var min, max result
	for _, r := range c.results {
		if r.runtime < minDur {
			min = r
			minDur = r.runtime
		}
		if r.runtime > maxDur {
			max = r
			maxDur = r.runtime
		}
		tot += r.runtime
	}

	avr := int64(tot) / int64(n)
	varSum := 0.0
	for _, x := range c.results {
		varSum += math.Pow((float64(x.runtime) - float64(avr)), 2.0)
	}
	variance := varSum / float64(n)
	stdDeviation := time.Duration(int64(math.Sqrt(variance)))
	return statistics{
		min:     min,
		max:     max,
		total:   tot,
		average: time.Duration(avr),
		stdDev:  stdDeviation,
	}
}

func (s *statistics) String() string {
	return fmt.Sprintf(`
Total time: %v, Average time per task: %v, Std deviation: %v
Max time, index: %v, time: %v
Min time, index: %v, time: %v`, s.total, s.average, s.stdDev,
		s.max.idx, s.max.runtime,
		s.min.idx, s.min.runtime)
}
