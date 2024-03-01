package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/baalimago/repeater/pkg/filetools"
)

func Test_do(t *testing.T) {
	t.Run("it should run command am amount of times, one workers", func(t *testing.T) {
		expectedCalls := 123
		testFilePath := fmt.Sprintf("%v/testFile", t.TempDir())
		// Add one line per run, anticipate a certain amount of lines in the test file...
		cOper := configuredOper{
			am:            expectedCalls,
			args:          []string{"/bin/bash", "-c", fmt.Sprintf("echo 'line' >> %v", testFilePath)},
			amIdleWorkers: 1,
			workerWg:      &sync.WaitGroup{},
			workPlanMu:    &sync.Mutex{},
		}
		cOper.workerWg.Add(1)
		cOper.run(context.Background())
		// ... anticipate a certain amount of lines in the file when it's done
		got, err := filetools.CheckAmLines(testFilePath)
		if err != nil {
			t.Fatalf("failed to check amount of lines: %v", err)
		}

		if got != expectedCalls {
			t.Fatalf("expected: %v, got: %v", expectedCalls, got)
		}
	})

	for i := 0; i < 1000; i++ {
		t.Run("it should run command am amount of times, 10 workers", func(t *testing.T) {
			expectedCalls := 123
			amWorkers := 10
			testFilePath := fmt.Sprintf("%v/testFile", t.TempDir())
			// Add one line per run, anticipate a certain amount of lines in the test file...
			cOper := configuredOper{
				am:            expectedCalls,
				workers:       amWorkers,
				amIdleWorkers: amWorkers,
				args:          []string{"/bin/bash", "-c", fmt.Sprintf("echo 'line' >> %v", testFilePath)},
				workerWg:      &sync.WaitGroup{},
				workPlanMu:    &sync.Mutex{},
			}
			cOper.workerWg.Add(amWorkers)
			cOper.run(context.Background())
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

}
