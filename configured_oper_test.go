package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
	"github.com/baalimago/repeater/internal/output"
)

func checkReportFileContent(reportFile string) (string, error) {
	testFileRead, err := os.Open(reportFile)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(testFileRead)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Test_configuredOper(t *testing.T) {
	t.Run("it should print to file when flagged to do so", func(t *testing.T) {
		runTest := func(outputMode output.Mode, wantOutputString bool) {
			testFile := testboil.CreateTestFile(t, "tFile")
			outputString := "test"
			co := configuredOper{
				am:            1,
				args:          []string{"printf", fmt.Sprintf("%v", outputString)},
				progress:      output.HIDDEN,
				output:        outputMode,
				outputFile:    testFile,
				amIdleWorkers: 1,
				workPlanMu:    &sync.Mutex{},
				workerWg:      &sync.WaitGroup{},
			}
			co.workerWg.Add(1)

			co.run(context.Background())
			testFileName := testFile.Name()
			testFile.Close()
			got, err := checkReportFileContent(testFileName)
			if err != nil {
				t.Fatalf("failed to get report file: %v", err)
			}
			if got != outputString && wantOutputString {
				t.Fatalf("for: %v, expected: %v, got: %v", outputMode, outputString, got)
			} else if got == outputString && !wantOutputString {
				t.Fatalf("for: %v, expected empty string, got: %v", outputMode, got)
			}
		}
		runTest(output.STDOUT, false)
		runTest(output.HIDDEN, false)
		runTest(output.BOTH, true)
		runTest(output.FILE, true)
	})

	t.Run("it should print progress to report file when flagged to do so", func(t *testing.T) {
		runTest := func(outputMode output.Mode, wantProgress bool) {
			testFile := testboil.CreateTestFile(t, "tFile")
			outputString := "something"
			progFormat := "%v/%v/%v/%v/%v/%v"
			co := configuredOper{
				am:             1,
				args:           []string{"printf", fmt.Sprintf("%v", outputString)},
				progressFormat: progFormat,
				progress:       outputMode,
				amIdleWorkers:  1,
				output:         output.HIDDEN,
				outputFile:     testFile,
				workPlanMu:     &sync.Mutex{},
				workerWg:       &sync.WaitGroup{},
			}
			co.workerWg.Add(1)

			co.run(context.Background())
			testFileName := testFile.Name()
			testFile.Close()

			got, err := checkReportFileContent(testFileName)
			if err != nil {
				t.Fatalf("failed to get report file: %v", err)
			}
			wantStr := fmt.Sprintf(progFormat, 1, 0, 1, 0.1, time.Now(), 0.1)
			if strings.Contains(got, wantStr) && wantProgress {
				t.Fatalf("for: %s, expected: %v, got: %v", outputMode, wantStr, got)
			} else if got == wantStr && !wantProgress {
				t.Fatalf("for: %s, expected empty string, got: %v", outputMode, got)
			}
		}
		runTest(output.STDOUT, false)
		runTest(output.HIDDEN, false)
		runTest(output.BOTH, true)
		runTest(output.FILE, true)
	})

	t.Run("it should follow format set in progressFormat", func(t *testing.T) {
		wantFormat := "lol%vtest%vhere%vmore%vfields%vnow%v"
		testFile := testboil.CreateTestFile(t, "testFile")
		c := configuredOper{
			am:             1,
			args:           []string{"true"},
			amIdleWorkers:  1,
			progress:       output.FILE,
			progressFormat: wantFormat,
			output:         output.HIDDEN,
			outputFile:     testFile,
			workPlanMu:     &sync.Mutex{},
			workerWg:       &sync.WaitGroup{},
		}
		c.workerWg.Add(1)

		c.run(context.Background())
		testFileName := testFile.Name()
		testFile.Close()
		got, err := checkReportFileContent(testFileName)
		if err != nil {
			t.Fatalf("failed to get report file: %v", err)
		}
		want := fmt.Sprintf(wantFormat, 1, 0, 1, 1.0, time.Now(), 1.0)
		if strings.Contains(got, want) {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}

func Test_results(t *testing.T) {
	t.Run("it should report output into results", func(t *testing.T) {
		// This should ouput "test"
		want := "test"
		c := configuredOper{
			am:            1,
			args:          []string{"printf", want},
			workPlanMu:    &sync.Mutex{},
			amIdleWorkers: 1,
			workerWg:      &sync.WaitGroup{},
		}
		c.workerWg.Add(1)

		c.run(context.Background())
		gotLen := len(c.results)
		if gotLen != 1 {
			t.Fatalf("expected: 1, got: %v", gotLen)
		}

		got := c.results[0].Output
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should report all output into results", func(t *testing.T) {
		wantAm := 10
		c := configuredOper{
			am: wantAm,
			// Date is most likely to exist in most OS's running this test
			args:          []string{"date"},
			workerWg:      &sync.WaitGroup{},
			workPlanMu:    &sync.Mutex{},
			amIdleWorkers: 1,
		}
		c.workerWg.Add(1)
		c.run(context.Background())
		time.Sleep(time.Millisecond)
		gotLen := len(c.results)
		// ensure that the correc amount is output
		if gotLen != wantAm {
			t.Fatalf("expected: %v, got: %v", wantAm, gotLen)
		}

		uniqueSet := make(map[string]struct{})
		for _, k := range c.results {
			_, exists := uniqueSet[k.Output]
			// Ensure that the output isn't copied for each one
			if exists {
				t.Fatalf("expected output to be different, this has shown twice: %v", exists)
			}
		}
	})
}

func Test_configuredOper_New(t *testing.T) {
	t.Run("it should return incrementConfigError if increment is true and no args contains 'INC'", func(t *testing.T) {
		args := []string{"test", "abc"}
		_, gotErr := New(0, 0, args, output.HIDDEN, "testing", output.HIDDEN, "", "", true, "", false)
		if gotErr == nil {
			t.Fatal("expected to get error, got nil")
		}

		var got incrementConfigError
		if !errors.As(gotErr, &got) {
			t.Fatalf("expected to get incrementConfigError, got: %v", gotErr)
		}

		for _, want := range args {
			if !strings.Contains(got.Error(), want) {
				t.Fatalf("error: %v, does not contain: %v", got, want)
			}
		}
	})

	t.Run("it should not return an error if increment is true and one argument is 'INC'", func(t *testing.T) {
		args := []string{"test", "abc", "INC"}
		_, gotErr := New(0, 0, args, output.HIDDEN, "testing", output.HIDDEN, "", "", true, "", false)
		if gotErr != nil {
			t.Fatalf("expected nil, got: %v", gotErr)
		}
	})

	t.Run("it should not return an error if increment is true and one argument contains 'INC'", func(t *testing.T) {
		args := []string{"test", "abc", "another-argument/INC"}
		_, gotErr := New(0, 0, args, output.HIDDEN, "testing", output.HIDDEN, "", "", true, "", false)
		if gotErr != nil {
			t.Fatalf("expected nil, got: %v", gotErr)
		}
	})

	t.Run("it should return incrementConfigError if the number of workers is greater than the number of times to repeat the command", func(t *testing.T) {
		am := 1
		workers := 2
		args := []string{"test", "abc"}
		_, gotErr := New(am, workers, args, output.HIDDEN, "testing", output.HIDDEN, "", "", false, "", false)
		if gotErr == nil {
			t.Fatal("expected to get error, got nil")
		}

		want := fmt.Errorf("please use less workers than repetitions. Am workers: %v, am repetitions: %v", workers, am)

		if gotErr.Error() != want.Error() {
			t.Fatalf("got: %v, want: %v", gotErr, want)
		}
	})

	t.Run("it should not return an error if the number of workers is lower than the number of times to repeat the command", func(t *testing.T) {
		args := []string{"test", "abc"}
		_, gotErr := New(2, 1, args, output.HIDDEN, "testing", output.HIDDEN, "", "", false, "", false)
		if gotErr != nil {
			t.Fatalf("expected nil, got: %v", gotErr)
		}
	})
	t.Run("it should not return an error if the number of workers is equal to the number of times to repeat the command", func(t *testing.T) {
		args := []string{"test", "abc"}
		_, gotErr := New(2, 2, args, output.HIDDEN, "testing", output.HIDDEN, "", "", false, "", false)
		if gotErr != nil {
			t.Fatalf("expected nil, got: %v", gotErr)
		}
	})
}
