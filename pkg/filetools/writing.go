package filetools

import (
	"context"
	"fmt"
	"os"
)

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

// Duplicator interface describing some struct which attempts to duplicate something,
// returning the amount of bytes that has been duplicated and and error on failures
type Duplicator interface {
	// Duplicate some data, return amount of bytes duplicated and/or an error
	Duplicate(context.Context) (int64, error)
}
