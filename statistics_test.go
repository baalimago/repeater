package main

import (
	"math"
	"testing"
	"time"
)

func TestResultWrite(t *testing.T) {
	r := result{}
	input := []byte("test")
	n, err := r.Write(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(input) {
		t.Errorf("Expected number of bytes written: %d, got: %d", len(input), n)
	}
	if r.output != string(input) {
		t.Errorf("Expected output: %s, got: %s", string(input), r.output)
	}
}

func TestCalcStats(t *testing.T) {
	// Create some dummy results for testing
	results := []result{
		{workerID: 1, idx: 1, runtime: 10 * time.Second},
		{workerID: 2, idx: 2, runtime: 20 * time.Second},
		{workerID: 3, idx: 3, runtime: 30 * time.Second},
	}

	// Create a configuredOper with the dummy results
	c := configuredOper{results: results}

	// Calculate statistics
	stats := c.calcStats()

	want := 10 * time.Second
	got := stats.min.runtime
	if want != got {
		t.Fatalf("expected: %v, got: %v", want, got)
	}

	want = 30 * time.Second
	got = stats.max.runtime
	if want != got {
		t.Fatalf("expected: %v, got: %v", want, got)
	}

	// Check if the calculated statistics are as expected
	expectedTotal := 60 * time.Second
	if stats.total != expectedTotal {
		t.Errorf("Expected total time: %v, got: %v", expectedTotal, stats.total)
	}

	expectedAverage := (10*time.Second + 20*time.Second + 30*time.Second) / 3
	if stats.average != expectedAverage {
		t.Errorf("Expected average time: %v, got: %v", expectedAverage, stats.average)
	}

	expectedStdDev := time.Duration(math.Sqrt((math.Pow(float64(10*time.Second-expectedAverage), 2) +
		math.Pow(float64(20*time.Second-expectedAverage), 2) +
		math.Pow(float64(30*time.Second-expectedAverage), 2)) / 3))
	if stats.stdDev != expectedStdDev {
		t.Errorf("Expected standard deviation: %v, got: %v", expectedStdDev, stats.stdDev)
	}
}

func TestStatisticsString(t *testing.T) {
	stats := statistics{
		max:     result{idx: 1, runtime: 30 * time.Second},
		min:     result{idx: 2, runtime: 10 * time.Second},
		total:   60 * time.Second,
		average: 20 * time.Second,
		stdDev:  10 * time.Second,
	}

	expectedString := `
Total time: 1m0s, Average time per task: 20s, Std deviation: 10s
Max time, index: 1, time: 30s
Min time, index: 2, time: 10s`
	if stats.String() != expectedString {
		t.Errorf("Expected string representation: %s, got: %s", expectedString, stats.String())
	}
}
