package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/baalimago/repeater/internal/output"
)

const incrementPlaceholder = "INC"

type configuredOper struct {
	am             int
	workers        int
	args           []string
	color          bool
	progress       output.Mode
	progressFormat string
	output         output.Mode
	reportFile     *os.File
	reportFileMu   *sync.Mutex
	reportFileMode string
	increment      bool
	processingTime time.Duration
	results        []result
}

type userQuitError string

func (uqe userQuitError) Error() string {
	return string(uqe)
}

const UserQuitError userQuitError = "user quit"

type incrementConfigError struct {
	args []string
}

func (ice incrementConfigError) Error() string {
	return fmt.Sprintf("increment is true, but args: %v, does not contain string '%s'", ice.args, incrementPlaceholder)
}

func New(am, workers int,
	args []string,
	pMode output.Mode,
	progressFormat string,
	oMode output.Mode,
	reportFile string,
	reportFileMode string,
	increment bool,
) (configuredOper, error) {
	shouldHaveReportFile := pMode == output.BOTH || pMode == output.REPORT_FILE ||
		oMode == output.BOTH || oMode == output.REPORT_FILE

	if shouldHaveReportFile && reportFile == "" {
		return configuredOper{}, fmt.Errorf("progress mode '%v', or output mode '%v', requires a report file but none is set. Use flag --reportFile <file_name>", pMode, oMode)
	}

	if increment && !containsIncrementPlaceholder(args) {
		return configuredOper{}, incrementConfigError{args: args}
	}

	if workers > am {
		return configuredOper{}, fmt.Errorf("please use less workers than repetitions. Am workers: %v, am repetitions: %v", workers, am)
	}

	c := configuredOper{
		am:             am,
		workers:        workers,
		args:           args,
		progress:       pMode,
		progressFormat: progressFormat,
		output:         oMode,
		reportFileMu:   &sync.Mutex{},
		increment:      increment,
	}

	file, err := c.getFile(reportFile, reportFileMode)
	if err != nil {
		if errors.Is(err, UserQuitError) {
			return c, err
		}
		return c, fmt.Errorf("failed to get file: %v", err)
	}
	c.reportFile = file
	return c, nil
}

// getFile a file. if one already exists, either consult the fileMode string, or query
// user how the file should be treated
func (c *configuredOper) getFile(s, fileMode string) (*os.File, error) {
	if s == "" {
		return nil, nil
	}

	if _, err := os.Stat(s); !errors.Is(err, os.ErrNotExist) {
		userResp := fileMode
		if fileMode == "" {
			printWarn(fmt.Sprintf("file: \"%v\", already exists. Would you like to [t]runcate, [a]ppend or [q]uit? [t/a/q]: ", s))
			fmt.Scanln(&userResp)
		}
		cleanedUserResp := strings.ToLower(strings.TrimSpace(userResp))
		c.reportFileMode = cleanedUserResp
		switch cleanedUserResp {
		case "t":
			// NOOP, fallthrough to os.Create below
		case "a":
			return os.OpenFile(s, os.O_APPEND|os.O_RDWR, 0o644)
		case "q":
			return nil, UserQuitError
		default:
			return nil, fmt.Errorf("unrecognized reply: \"%v\", valid options are [tT], [aA] or [qQ]", userResp)
		}
	}
	return os.Create(s)
}

func (c *configuredOper) String() string {
	reportFileName := "HIDDEN"
	if c.reportFile != nil {
		reportFileName = c.reportFile.Name()
	}
	return fmt.Sprintf(`am: %v
command: %v
increment: %v
workers: %v
color: %v
progress: %s
progress format: %q
output: %s
report file: %v
report file mode: %v`, c.am, c.args, c.increment, c.workers, c.color, c.progress, c.progressFormat, c.output, reportFileName, c.reportFileMode)
}

func (c *configuredOper) setupOutputStreams(toDo *exec.Cmd, res *result) {
	switch c.output {
	case output.STDOUT:
		toDo.Stdout = io.MultiWriter(os.Stdout, res)
		toDo.Stderr = io.MultiWriter(os.Stderr, res)
	case output.HIDDEN:
		toDo.Stdout = res
		toDo.Stderr = res
	case output.REPORT_FILE:
		toDo.Stdout = io.MultiWriter(c.reportFile, res)
		toDo.Stderr = io.MultiWriter(c.reportFile, res)
	case output.BOTH:
		toDo.Stdout = io.MultiWriter(c.reportFile, os.Stdout, res)
		toDo.Stderr = io.MultiWriter(c.reportFile, os.Stderr, res)
	}
}

func (c *configuredOper) setupProgressStreams() []io.Writer {
	progressStreams := make([]io.Writer, 0, 2)
	switch c.progress {
	case output.STDOUT:
		progressStreams = append(progressStreams, os.Stdout)
	case output.REPORT_FILE:
		progressStreams = append(progressStreams, c.reportFile)
	case output.BOTH:
		progressStreams = append(progressStreams, os.Stdout)
		progressStreams = append(progressStreams, c.reportFile)
	}
	return progressStreams
}

func containsIncrementPlaceholder(args []string) bool {
	for _, arg := range args {
		if strings.Contains(arg, incrementPlaceholder) {
			return true
		}
	}
	return false
}
