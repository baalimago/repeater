package filetools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type DuplicationError struct {
	errs []error
}

func (de DuplicationError) add(err error) {
	de.errs = append(de.errs, err)
}

func (de DuplicationError) Error() string {
	return fmt.Sprintf("got: %v errors, unwrap to check details", len(de.errs))
}

func (de DuplicationError) Unwrap() []error {
	return de.errs
}

// EoFDuplicator will check if file is currently being written to and await
// EoF before duplicating it
type EoFDuplicator struct {
	src *os.File
	// p is point in file of previous read. Slight optimization to skip having to rescan src
	// on each poll
	p        int64
	dst      []*os.File
	pollRate time.Duration
}

func NewEoFDuplicator(src *os.File, dst []*os.File, pollRate time.Duration) EoFDuplicator {
	return EoFDuplicator{
		src:      src,
		dst:      dst,
		pollRate: pollRate,
		p:        int64(0),
	}
}

func (ed EoFDuplicator) isReady() bool {
	stepSize := int64(1024)
	buf := make([]byte, stepSize)
	i := ed.p
	for {
		n, err := ed.src.ReadAt(buf, i*stepSize)
		if errors.Is(err, io.EOF) {
			return true
		}
		// Somehow, the file has ended, yet there is no EoF. File might still be writing
		// or some similar error in the file. Unsafe to duplicate from.
		if n != int(stepSize) && err == nil {
			return false
		}
		i++
	}
}

// Duplicate src to dst in a blocking manneer, awaiting EoF before copying
// all content from src to dst
func (ed EoFDuplicator) Duplicate(ctx context.Context) (int64, error) {
	dupErr := DuplicationError{
		errs: make([]error, 0),
	}
	for {
		// context cancelled, break
		if ctx.Err() != nil {
			return -1, ctx.Err()
		}
		if ed.isReady() {
			break
		}
		time.Sleep(ed.pollRate)
	}

	totAmB := int64(0)
	for _, dst := range ed.dst {
		amB, err := io.Copy(dst, ed.src)
		if err != nil {
			dupErr.add(err)
			continue
		}
		totAmB += amB
	}
	return totAmB, nil
}
