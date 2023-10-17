package filetools

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// CheckAmLines by parsing the file and counting the amount of lines using
// a scanner. Will return error if the file fails to open or if the scanner fails somehow
// On error, am will be -1
func CheckAmLines(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return -1, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	// Solution joinked from: https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

type WriteErrors struct {
	errs []error
}

func (we *WriteErrors) Error() string {
	return fmt.Sprintf("got: %v write errors. unwrap for more details", len(we.errs))
}

func (we *WriteErrors) Unwrap() []error {
	return we.errs
}

func (we *WriteErrors) add(err WriteError) {
	we.errs = append(we.errs, err)
}

type WriteError struct {
	File *os.File
	Err  error
}

func (we WriteError) Error() string {
	return fmt.Sprintf("write error for: %v, err: %v", *we.File, we.Err)
}

// WriteIfPossible to the files, retuning the total amount of bytes written
// and a WriteErrors containing any potential errors
func WriteIfPossible(str string, files []*os.File) (int, error) {
	tot := 0
	var we error
	for _, f := range files {
		bytes, err := fmt.Fprint(f, str)
		tot += bytes
		if err != nil {
			// copy loop variable
			f := f
			if we == nil {
				we = &WriteErrors{
					errs: make([]error, 0),
				}
			}
			we.(*WriteErrors).add(WriteError{
				Err:  err,
				File: f,
			})
		}
	}
	return tot, we
}
