package filetools

import (
	"fmt"
	"io"
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
	File io.Writer
	Err  error
}

func (we WriteError) Error() string {
	return fmt.Sprintf("write error for: %v, err: %v", we.File, we.Err)
}

// WriteStringIfPossible to the files, retuning the total amount of bytes written
// and a WriteErrors containing any potential errors
func WriteStringIfPossible(str string, files []io.Writer) (int, error) {
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
