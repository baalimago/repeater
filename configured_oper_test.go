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
			testFile.Sync()
			testFileName := testFile.Name()
			testFile.Close()

			testFileRead, err := os.Open(testFileName)
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(testFileRead)
			if err != nil {
				t.Fatal(err)
			}
			got := string(b)
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
			testFile.Sync()
			testFileName := testFile.Name()
			testFile.Close()

			testFileRead, err := os.Open(testFileName)
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(testFileRead)
			if err != nil {
				t.Fatal(err)
			}
			// It appends newline at end
			wantStr := fmt.Sprintf("%v\n", fmt.Sprintf(progFormat, 1, 1))
			got := string(b)
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
}
