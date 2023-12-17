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
}

// Write implements io.Writer to get the output of the command for
// both out and err
func (r *result) Write(p []byte) (n int, err error) {
	r.output += string(p)
	return len(p), nil
}

type statistics struct {
	max       result
	min       result
	totalTime time.Duration
	res       []result
}

func (s statistics) String() string {
	tot := time.Duration(0)
	n := len(s.res)
	if n == 0 {
		return "\nNo results were caught, cannot produce statisitcs"
	}
	for _, r := range s.res {
		tot += r.runtime
	}

	avr := int64(tot) / int64(n)
	varSum := 0.0
	for _, x := range s.res {
		varSum += math.Pow((float64(x.runtime) - float64(avr)), 2.0)
	}
	variance := varSum / float64(n)
	stdDeviation := time.Duration(int64(math.Sqrt(variance)))

	return fmt.Sprintf(`
Total time: %v, Average time per task: %v, Std deviation: %v
Max time, index: %v, time: %v
Min time, index: %v, time: %v`, s.totalTime, time.Duration(avr), time.Duration(stdDeviation),
		s.max.idx, s.max.runtime,
		s.min.idx, s.min.runtime)
}
