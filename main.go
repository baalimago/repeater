package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baalimago/repeater/internal/output"
	"github.com/baalimago/repeater/internal/progress"
)

const DEFAULT_PROGRESS_FORMAT = "\rProgress: (%v/%v)"

var amRunsFlag = flag.Int("n", 1, "Amount of times you wish to repeat the command.")
var verboseFlag = flag.Bool("v", false, "Set to display the configured operation before running")
var colorFlag = flag.Bool("color", true, "Set to false to disable ANSI colors in the setup phase.")
var progressFlag = flag.String("progress", "stdout", "Options are: ['HIDDEN', 'REPORT_FILE', 'STDOUT', 'BOTH']")
var progressFormatFlag = flag.String("progressFormat", DEFAULT_PROGRESS_FORMAT, "Set the format for the output where first argument is the iteration and second argument is the amount of runs.")
var outputFlag = flag.String("output", "stdout", "Options are: ['HIDDEN', 'REPORT_FILE', 'STDOUT', 'BOTH']")
var reportFlag = flag.Bool("report", true, "Set to false to not get report.")
var reportFileFlag = flag.String("reportFile", "stdout", "Path to the file where the report will be saved. Options are: ['stdout', '<any file>']")

type colorCode int

const (
	RED colorCode = iota + 31
	GREEN
	YELLOW
)

func coloredMessage(cc colorCode, msg string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", cc, msg)
}

// getFile by checking if it exists and querying user about how to treat the file
func getFile(s string) *os.File {
	f, err := os.Create(s)
	if err != nil {
		panic(fmt.Sprintf("not good: %v", err))
	}
	return f
}

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		errMsg := "you need to supply at least 1 argument\n"
		if *colorFlag {
			fmt.Fprintf(os.Stderr, "%v: %v", coloredMessage(RED, "error"), errMsg)
		} else {
			fmt.Fprintf(os.Stderr, "error: %v", errMsg)
		}
		os.Exit(1)
	}

	c := configuredOper{
		am:             *amRunsFlag,
		args:           args,
		color:          *colorFlag,
		progress:       progress.New(progressFlag),
		progressFormat: *progressFormatFlag,
		output:         output.New(outputFlag),
		reportFile:     getFile(*reportFileFlag),
	}

	if *verboseFlag {
		fmt.Printf("%v\n", c)
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	isDone := make(chan struct{})
	go func() {
		c.run(ctx)
		close(isDone)
	}()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Listening for termination signals. Press Ctrl+C to exit.")

	// Block until a termination signal is received, or if all commands are done
	select {
	case <-isDone:
		c.printOK("command repeated\n")
		os.Exit(0)
	case <-signalChannel:
	}
	ctxCancel()
	select {
	case <-isDone:
		if ctx.Err() != nil {
			c.printOK("graceful shutdown complete")
		}
	case <-signalChannel:
		c.printErr("aborting graceful shutdown")
	}
}
