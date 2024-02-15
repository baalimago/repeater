package filetools_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
	"github.com/baalimago/repeater/pkg/filetools"
)

type errorWriter struct{}

func (ew errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("here i go erroring again!")
}

func Test_WriteStirngIfPossible(t *testing.T) {
	t.Run("it should write... if.. possible", func(t *testing.T) {
		goodFile := testboil.CreateTestFile(t, "testFile")
		t.Cleanup(func() { goodFile.Close() })

		want := "writethis"
		_, err := filetools.WriteStringIfPossible(want, []io.Writer{goodFile})
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

	t.Run("it should error if not possible", func(t *testing.T) {
		want := "writethis"
		_, got := filetools.WriteStringIfPossible(want, []io.Writer{&errorWriter{}})
		var wrErr filetools.WriteError
		if !errors.As(got, &wrErr) {
			t.Fatal("expected WriteError, got nil")
		}
	})

}
