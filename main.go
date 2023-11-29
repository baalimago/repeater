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

var (
	amRunsFlag         = flag.Int("n", 1, "Amount of times you wish to repeat the command.")
	verboseFlag        = flag.Bool("v", false, "Set to display the configured operation before running")
	workersFlag        = flag.Int("w", 1, "Set the amout of workers to repeat the command with. Having more than 1 makes execution paralell. Expect performance diminishing returns when approaching CPU threads.")
	colorFlag          = flag.Bool("color", true, "Set to false to disable ANSI colors in the setup phase.")
	progressFlag       = flag.String("progress", "STDOUT", "Options are: ['HIDDEN', 'REPORT_FILE', 'STDOUT', 'BOTH']")
	progressFormatFlag = flag.String("progressFormat", DEFAULT_PROGRESS_FORMAT, "Set the format for the output where first argument is the iteration and second argument is the amount of runs.")
	outputFlag         = flag.String("output", "STDOUT", "Options are: ['HIDDEN', 'REPORT_FILE', 'STDOUT', 'BOTH']")
	reportFlag         = flag.Bool("report", true, "Set to false to not get report.")
	reportFileFlag     = flag.String("reportFile", "STDOUT", "Path to the file where the report will be saved. Options are: ['STDOUT', '<any file>']")
	statisticsFlag     = flag.Bool("statistics", true, "Set to false if you don't wish to see statistics of the repeated command.")
	incrementFlag      = flag.Bool("increment", false, "Set to true and add an argument 'INC', to have 'INC' be replaced with the iteration. If increment == true && 'INC' is not set, repeater will panic.")
)

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
	// If it's stdout, it shouldn't create file as report should be written to stdout
	if s == "STDOUT" {
		return nil
	}
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
		workers:        *workersFlag,
		color:          *colorFlag,
		progress:       progress.New(progressFlag),
		progressFormat: *progressFormatFlag,
		output:         output.New(outputFlag),
		reportFile:     getFile(*reportFileFlag),
		increment:      *incrementFlag,
	}

	c, err := New(*amRunsFlag, *workersFlag, args, *colorFlag, progress.New(progressFlag), *progressFormatFlag, output.New(outputFlag), getFile(*reportFileFlag), *incrementFlag)
	if err != nil {
		c.printErr(fmt.Sprintf("configuration error: %v\n", err))
		os.Exit(1)
	}

	if *verboseFlag {
		fmt.Printf("%s\n", c)
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	isDone := make(chan statistics)
	go func() {
		isDone <- c.run(ctx)
	}()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Listening for termination signals. Press Ctrl+C to exit.")

	// Block until a termination signal is received, or if all commands are done
	select {
	case stats := <-isDone:
		if *statisticsFlag {
			fmt.Printf("== Statistics ==%s\n", stats)
		}
		c.printOK("i have done the repeat.\n")
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
