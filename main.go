package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/internal/progress"
)

var amRunsFlag = flag.Int("n", 1, "Amount of times you wish to repeat the command.")
var colorFlag = flag.Bool("color", true, "Set to false to disable ANSI colors in the setup phase.")
var progressFlag = flag.String("progress", "stdout", "Options are: ['hidden', 'reportFile', 'stdout', 'both']")
var outputFlag = flag.String("output", "stdout", "Options are: ['hidden', 'reportFile', 'stdout', 'both']")
var reportFlag = flag.Bool("report", true, "Set to false to not get report.")
var reportFileFlag = flag.String("reportFile", "stdout", "Path to the file where the report will be saved. Options are: ['stdout', '<any file>']")

type colorCode int

const (
	RED colorCode = iota + 31
	GREEN
	YELLOW
)

type configuredOper struct {
	am         int
	args       []string
	color      bool
	progress   progress.Mode
	output     output.Mode
	reportFile *os.File
}

func coloredMessage(cc colorCode, msg string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", cc, msg)
}

func (c configuredOper) printStatus(out io.Writer, status, msg string, color colorCode) {
	if c.color {
		status = coloredMessage(color, status)
	}
	fmt.Fprintf(out, "%v: %v", status, msg)
}

func (c configuredOper) printErr(msg string) {
	c.printStatus(os.Stderr, "error: ", msg, RED)
}

// getFile by checking if it exists and querying user about how to treat the file
func getFile(s string) *os.File {
	panic("unimplemented")
}

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		if *colorFlag {
			fmt.Fprintf(os.Stderr, "%v: %v", coloredMessage(RED, "error"), "you need to supply at least 1 argument")
		} else {
			fmt.Fprintf(os.Stderr, "error: %v", "you need to supply at least 1 argument")
		}
		os.Exit(1)
	}

	c := configuredOper{
		am:    *amRunsFlag,
		args:  args,
		color: *colorFlag,
		// reportFile: getFile(*reportFileFlag),
	}
	c.run()
}

// run the configured command. Blocking operation, errors are handeled internally as the output
// depends on the configuration
func (c configuredOper) run() {
	for i := 0; i < c.am; i++ {
		do := exec.Command(c.args[0], c.args[1:]...)
		do.Stdout = os.Stdout
		do.Stderr = os.Stderr
		err := do.Run()
		if errors.Is(err, &exec.ExitError{}) {
			c.printErr(fmt.Sprintf("unexpected error encountered, aborting operations: %v", *err.(*exec.ExitError)))
		}
	}
}
