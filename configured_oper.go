package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/internal/progress"
	"github.com/baalimago/repeater/pkg/filetools"
)

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c configuredOper) run(ctx context.Context) {
	progressStreams := make([]io.Writer, 0, 2)
	switch c.progress {
	case progress.STDOUT:
		progressStreams = append(progressStreams, os.Stdout)
	case progress.REPORT_FILE:
		progressStreams = append(progressStreams, c.reportFile)
	case progress.BOTH:
		progressStreams = append(progressStreams, os.Stdout)
		progressStreams = append(progressStreams, c.reportFile)
	}

	defer filetools.WriteStringIfPossible("\n", progressStreams)

	for i := 0; i < c.am; i++ {
		if ctx.Err() != nil {
			c.printErr(fmt.Sprintf("context error: %v", ctx.Err()))
			os.Exit(1)
		}
		do := exec.Command(c.args[0], c.args[1:]...)
		switch c.output {
		case output.STDOUT:
			do.Stdout = os.Stdout
			do.Stderr = os.Stderr
		case output.HIDDEN:
		case output.REPORT_FILE:
			do.Stdout = c.reportFile
			do.Stderr = c.reportFile
		case output.BOTH:
			do.Stdout = io.MultiWriter(c.reportFile, os.Stdout)
			do.Stderr = io.MultiWriter(c.reportFile, os.Stderr)
		}
		err := do.Run()
		if errors.Is(err, &exec.ExitError{}) {
			c.printErr(fmt.Sprintf("unexpected error encountered, aborting operations: %v", *err.(*exec.ExitError)))
			os.Exit(1)
		}
		filetools.WriteStringIfPossible(fmt.Sprintf(c.progressFormat, i+1, c.am), progressStreams)
	}
}
