package filetools_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/baalimago/repeater/pkg/filetools"
)

func Test_CheckAmLines(t *testing.T) {

	t.Run("given 10 lines, it should return 10", func(t *testing.T) {
		fileWithRows := `0
1
2
3
4
5
6
7
8
9
`
		testFilePath := fmt.Sprintf("%v/testFile", t.TempDir())
		file, err := os.Create(testFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Cleanup(func() { file.Close() })

		// Write the data to the file
		_, err = fmt.Fprint(file, fileWithRows)
		if err != nil {
			t.Fatalf("failed to write file with rows: %v", err)
		}

		got, err := filetools.CheckAmLines(testFilePath)
		if err != nil {
			t.Errorf("failed to check lines: %v", err)
		}

		if got != 10 {
			t.Errorf("expected: 10, got: %v", got)
		}
	})
}

func Test_WriteIfPossible(t *testing.T) {
	t.Run("it should write... if.. possible", func(t *testing.T) {
		testFilePath := fmt.Sprintf("%v/testFile", t.TempDir())
		goodFile, err := os.Create(testFilePath)
		if err != nil {
			t.Fatalf("failed to open test file: %v", err)
		}
		t.Cleanup(func() { goodFile.Close() })

		want := "writethis"
		_, err = filetools.WriteIfPossible(want, []*os.File{goodFile})
		if err != nil {
			t.Fatalf("should be possible to write, err: %v", err)
		}

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
