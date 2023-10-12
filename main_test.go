package main

import (
	"io"
	"testing"

	"github.com/baalimago/repeater/internal/output"
)

func Test_do(t *testing.T) {
	t.Run("output tests", func(t *testing.T) {
		c := configuredOper{
			output:   output.STDOUT
			printOut: io.Writer,
			errOut:   io.Writer,
		}

	})
}
