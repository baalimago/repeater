package filetools_test

import (
	"io"
	"os"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/general"
	"github.com/baalimago/repeater/pkg/filetools"
)

func Test_WriteIfPossible(t *testing.T) {
	t.Run("it should write... if.. possible", func(t *testing.T) {
		goodFile := general.CreateTestFile(t, "testFile")
		t.Cleanup(func() { goodFile.Close() })

		want := "writethis"
		_, err := filetools.WriteIfPossible(want, []*os.File{goodFile})
		if err != nil {
			t.Fatalf("should be possible to write, err: %v", err)
		}

		testFilePath := goodFile.Name()
		readFile, err := os.Open(testFilePath)
		if err != nil {
			t.Fatalf("failed to open testFile: %v", err)
		}

		got, err := io.ReadAll(readFile)
		t.Cleanup(func() { readFile.Close() })
		if err != nil {
			t.Fatalf("failed to read test file: %v", err)
		}
		if string(got) != want {
			t.Fatalf("expected: %v, got: %s", want, got)
		}
	})
}
