package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/general"
	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/internal/progress"
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
	t.Run("it should print to report file when flagged to do so", func(t *testing.T) {
		runTest := func(outputMode output.Mode, wantOutputString bool) {
			testFile := general.CreateTestFile(t, "tFile")
			outputString := "test"
			co := configuredOper{
				am:         1,
				args:       []string{"printf", fmt.Sprintf("%v", outputString)},
				color:      false,
				progress:   progress.HIDDEN,
				output:     outputMode,
				reportFile: testFile,
			}

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
		runTest(output.REPORT_FILE, true)
	})

	t.Run("it should print progress to report file when flagged to do so", func(t *testing.T) {
		runTest := func(outputMode progress.Mode, wantProgress bool) {
			testFile := general.CreateTestFile(t, "tFile")
			outputString := "something"
			progFormat := "%v/%v"
			co := configuredOper{
				am:             1,
				args:           []string{"printf", fmt.Sprintf("%v", outputString)},
				color:          false,
				progressFormat: progFormat,
				progress:       outputMode,
				output:         output.HIDDEN,
				reportFile:     testFile,
			}

			co.run(context.Background())
			testFileName := testFile.Name()
			testFile.Close()

			got, err := checkReportFileContent(testFileName)
			if err != nil {
				t.Fatalf("failed to get report file: %v", err)
			}
			wantStr := fmt.Sprintf("%v\n", fmt.Sprintf(progFormat, 1, 1))
			if got != wantStr && wantProgress {
				t.Fatalf("for: %s, expected: %v, got: %v", outputMode, wantStr, got)
			} else if got == wantStr && !wantProgress {
				t.Fatalf("for: %s, expected empty string, got: %v", outputMode, got)
			}
		}
		runTest(progress.STDOUT, false)
		runTest(progress.HIDDEN, false)
		runTest(progress.BOTH, true)
		runTest(progress.REPORT_FILE, true)
	})

	t.Run("it should follow format set in outputFormat", func(t *testing.T) {
		wantFormat := "lol%vtest%v"
		testFile := general.CreateTestFile(t, "testFile")
		c := configuredOper{
			am:             1,
			args:           []string{"true"},
			color:          false,
			progress:       progress.REPORT_FILE,
			progressFormat: wantFormat,
			output:         output.HIDDEN,
			reportFile:     testFile,
		}

		c.run(context.Background())
		testFileName := testFile.Name()
		testFile.Close()
		got, err := checkReportFileContent(testFileName)
		if err != nil {
			t.Fatalf("failed to get report file: %v", err)
		}
		want := fmt.Sprintf("%v\n", fmt.Sprintf(wantFormat, 1, 1))
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}

func Test_results(t *testing.T) {
	t.Run("it should report output into results", func(t *testing.T) {
		// This should ouput "test"
		want := "test"
		c := configuredOper{
			am:   1,
			args: []string{"printf", want},
		}

		results := c.run(context.Background())
		gotLen := len(results.res)
		if gotLen != 1 {
			t.Fatalf("expected: 1, got: %v", gotLen)
		}

		got := results.res[0].output
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should report all output into results", func(t *testing.T) {
		wantAm := 10
		c := configuredOper{
			am: wantAm,
			// Date is most likely to exist in most OS's running this test
			args: []string{"date"},
		}
		results := c.run(context.Background())
		res := results.res
		gotLen := len(results.res)
		// ensure that the correc amount is output
		if gotLen != wantAm {
			t.Fatalf("expected: %v, got: %v", wantAm, gotLen)
		}

		uniqueSet := make(map[string]struct{})
		for _, k := range res {
			_, exists := uniqueSet[k.output]
			// Ensure that the output isn't copied for each one
			if exists {
				t.Fatalf("expected output to be different, this has shown twice: %v", exists)
			}
		}
	})
}
