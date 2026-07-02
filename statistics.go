package main

import (
	"fmt"
	"math"
	"time"
)

type OutputEvent struct {
	At   time.Time
	Text string
}

type Result struct {
	WorkerID             int           `json:"workerID"`
	Idx                  int           `json:"taskIdx"`
	Runtime              time.Duration `json:"runtime"`
	RuntimeHumanReadable string        `json:"runtimeHumanReadable"`
	Output               string        `json:"output"`
	Stdout               []OutputEvent `json:"stdout,omitempty"`
	Stderr               []OutputEvent `json:"stderr,omitempty"`
	IsError              bool          `json:"isError"`
	IsCancelled          bool          `json:"isCancelled"`
}

type statistics struct {
	am          int
	amDone      int
	amFails     int
	amCancelled int
	cancelled   bool
	max         Result
	min         Result
	total       time.Duration
	runtime     time.Duration
	average     time.Duration
	stdDev      time.Duration
	Results     []Result `json:"results"`
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
	if n == 0 {
		return statistics{}
	}
	minDur := time.Duration(9223372036854775807)
	maxDur := time.Duration(-9223372036854775808)
	amFails := 0
	amCancelled := 0
	var min, max Result
	for _, r := range c.results {
		if r.IsCancelled {
			amCancelled++
			continue
		}
		if r.IsError {
			amFails++
			continue
		}
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
		am:          c.am,
		amDone:      n,
		amFails:     amFails,
		amCancelled: amCancelled,
		cancelled:   c.wasCancelled,
		runtime:     c.runtime,
		min:         min,
		max:         max,
		total:       tot,
		average:     time.Duration(avr),
		stdDev:      stdDeviation,
		Results:     c.results,
	}
}

func (s *statistics) String() string {
	state := ""
	if s.cancelled {
		state = " (cancelled)"
	}
	return fmt.Sprintf(`
== Statistics ==
Amount of repitions: %v, completed: %v, amount of failures: %v, amount of cancelled: %v%s,
The following is calculated on successful attempts:
  Runtime: %s, Total routine work time: %v,
  Average time per task: %v, Std deviation: %v
  Max time, index: %v, time: %v
  Min time, index: %v, time: %v`,
		s.am, s.amDone, s.amFails, s.amCancelled, state,
		s.runtime, s.total,
		s.average, s.stdDev,
		s.max.Idx, s.max.Runtime,
		s.min.Idx, s.min.Runtime)
}
