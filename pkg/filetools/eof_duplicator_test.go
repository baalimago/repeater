package filetools

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/context_tests"
	"github.com/baalimago/go_away_boilerplate/pkg/general"
)

func Test_EofDuplicator(t *testing.T) {
	testFile := general.CreateTestFile(t, "testFile")
	testDuration := time.Millisecond * 10
	eofDuplicator := NewEoFDuplicator(testFile, nil, testDuration)
	context_tests.ReturnsOnContextCancel(t,
		func(ctx context.Context) {
			eofDuplicator.Duplicate(ctx)
		},
		testDuration,
	)

	t.Run("isReay", func(t *testing.T) {
		t.Run("it should return false if file lacks EOF", func(t *testing.T) {
			openFile, err := os.OpenFile(fmt.Sprintf("%v/testFile", t.TempDir()), os.O_WRONLY|os.O_CREATE, 0666)
			t.Cleanup(func() { openFile.Close() })
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			// Start a routine which continiously writes to file
			go func() {
				writeTicks := time.NewTicker(testDuration / 10)
				doneWriting := time.After(testDuration / 2)
				for {
					select {
					case <-writeTicks.C:
						openFile.Write([]byte("a"))
					case <-doneWriting:
						break
					}
				}
			}()
			ed := NewEoFDuplicator(openFile, nil, time.Millisecond)
			got := ed.isReady()
			want := false
			if got != want {
				t.Fatal("expected isReady to return false when src is open")
			}
		})
	})
}
