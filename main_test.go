package main

import (
	"fmt"
	"testing"

	"github.com/baalimago/repeater/pkg/filetools"
)

func Test_do(t *testing.T) {

	t.Run("it should run command am amount of times", func(t *testing.T) {
		expectedCalls := 123
		testFilePath := fmt.Sprintf("%v/testFile", t.TempDir())
		// Add one line per run, anticipate a certain amount of lines in the test file...
		cOper := configuredOper{
			am:   expectedCalls,
			args: []string{"/bin/bash", "-c", fmt.Sprintf("echo 'line' >> %v", testFilePath)},
		}
		cOper.run()
		// ... anticipate a certain amount of lines in the file when it's done
		got, err := filetools.CheckAmLines(testFilePath)
		if err != nil {
			t.Fatalf("failed to check amount of lines: %v", err)
		}

		if got != expectedCalls {
			t.Fatalf("expected: %v, got: %v", expectedCalls, got)
		}
	})
}
