package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/general"
)

func Test_configuredOper_getFile(t *testing.T) {
	createFileWithContent := func(t *testing.T, fileName, content string) *os.File {
		f, err := os.Create(fileName)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		_, err = f.WriteString(content)
		if err != nil {
			t.Fatalf("failed to write data to file in pre test: %v", err)
		}
		return f
	}

	testCases := []struct {
		desc  string
		given struct {
			fileName string
			fileMode string
			// fileName intended to be injected, as it is generated via t.TempDir(). Note that
			// fileGenFunc runs before the test/the getFile call, but fileName is generated on
			// test start
			fileGenFunc func(*testing.T, string) *os.File
		}
		want struct {
			err error
			// validate file integrity by checking its content
			fileContent string
		}
	}{
		{
			desc: "given nonexistent file, no fileMode, some file path, it should return new file",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName:    fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode:    "anything",
				fileGenFunc: func(*testing.T, string) *os.File { return nil },
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         nil,
				fileContent: "",
			},
		},
		{
			desc: "given existent file, truncate fileMode, new file path, it should return new file",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName:    fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode:    "t",
				fileGenFunc: func(*testing.T, string) *os.File { return nil },
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         nil,
				fileContent: "",
			},
		},
		{
			desc: "given existent file, truncate fileMode, existing path, it should return new file",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName: fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode: "t",
				fileGenFunc: func(t *testing.T, fileName string) *os.File {
					return createFileWithContent(t, fileName, "SHOULD_GO_AWAY")
				},
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         nil,
				fileContent: "",
			},
		},
		{
			desc: "given existent file, append fileMode, existing file path, it should return existing file",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName: fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode: "a",
				fileGenFunc: func(t *testing.T, fileName string) *os.File {
					return createFileWithContent(t, fileName, "SHOULD_STAY")
				},
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         nil,
				fileContent: "SHOULD_STAY",
			},
		},
		{
			desc: "given nonexisting file, append fileMode, new file path, it should return new file",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName: fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode: "a",
				fileGenFunc: func(t *testing.T, _ string) *os.File {
					f := general.CreateTestFile(t, "someNewFile")
					_, err := f.WriteString("SHOULD_BE_IGNORED")
					if err != nil {
						t.Fatalf("failed to write data to file in pre test: %v", err)
					}
					return f
				},
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         nil,
				fileContent: "",
			},
		},
		{
			desc: "given existing file, quit fileMode, existing file path, it should error",
			given: struct {
				fileName    string
				fileMode    string
				fileGenFunc func(*testing.T, string) *os.File
			}{
				fileName: fmt.Sprintf("%v/someNewFile", t.TempDir()),
				fileMode: "q",
				fileGenFunc: func(t *testing.T, fileName string) *os.File {
					return createFileWithContent(t, fileName, "doesn't matter, should quit")
				},
			},
			want: struct {
				err         error
				fileContent string
			}{
				err:         UserQuitError,
				fileContent: "",
			},
		},
	}

	assertFileContent := func(t *testing.T, f *os.File, want string) {
		t.Helper()
		if f == nil {
			return
		}
		b, err := io.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		got := string(b)

		if string(b) != want {
			t.Fatalf("failed to verify contents of file, want: %v, got: %v", want, got)
		}
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.given.fileGenFunc(t, tC.given.fileName)
			c := configuredOper{}
			gotFile, gotErr := c.getFile(tC.given.fileName, tC.given.fileMode)
			if tC.want.err != nil && !errors.Is(gotErr, tC.want.err) {
				t.Fatalf("expected error: %v, got: %v", tC.want.err, gotErr)
			}

			assertFileContent(t, gotFile, tC.want.fileContent)
		})
	}
}
