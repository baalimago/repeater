package main

import (
	"math"
	"testing"
	"time"
)

func TestResultWrite(t *testing.T) {
	r := Result{}
	input := []byte("test")
	n, err := r.Write(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(input) {
		t.Errorf("Expected number of bytes written: %d, got: %d", len(input), n)
	}
	if r.Output != string(input) {
		t.Errorf("Expected output: %s, got: %s", string(input), r.Output)
	}
}

func TestCalcStats(t *testing.T) {
	// Create some dummy results for testing
	results := []Result{
		{WorkerID: 1, Idx: 1, Runtime: 10 * time.Second},
		{WorkerID: 2, Idx: 2, Runtime: 20 * time.Second},
		{WorkerID: 3, Idx: 3, Runtime: 30 * time.Second},
	}

	// Create a configuredOper with the dummy results
	c := configuredOper{results: results}

	// Calculate statistics
	stats := c.calcStats()

	want := 10 * time.Second
	got := stats.min.Runtime
	if want != got {
		t.Fatalf("expected: %v, got: %v", want, got)
	}

	want = 30 * time.Second
	got = stats.max.Runtime
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
