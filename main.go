package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baalimago/repeater/internal/output"
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
	reportFileFlag     = flag.String("reportFile", "", "Path to the file where the report will be saved, configure file conflicts automatically with 'reportFileMode'")
	reportFileModeFlag = flag.String("reportFileMode", "", "Configure how the report file should be treated. If a reportFile exists, and this option isn't set, user will be queried. Options are: ['r'ecreate, 'a'ppend] ")
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
	c, err := New(*amRunsFlag, *workersFlag, args, *colorFlag, output.New(progressFlag), *progressFormatFlag, output.New(outputFlag), *reportFileFlag, *reportFileModeFlag, *incrementFlag)
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
		c.printOK("The repeat, has been done. Farewell.\n")
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
