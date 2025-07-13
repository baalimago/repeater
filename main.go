package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/baalimago/repeater/internal/output"
)

const DefaultProgressFormat = "\rProgress: (Success/Fail/Requested Am)(%v/%v/%v), Start at: %v, Remaining: %.1fs, Est. done at: %v"

var (
	amRunsFlag         = flag.Int("n", 1, "Amount of times you wish to repeat the command.")
	verboseFlag        = flag.Bool("v", false, "Set to display the configured operation before running")
	workersFlag        = flag.Int("w", 1, "Set the amout of workers to repeat the command with. Having more than 1 makes execution paralell. Expect performance diminishing returns when approaching CPU threads.")
	colorFlag          = flag.Bool("nocolor", false, "Set to true to disable ansi-colored output")
	progressFlag       = flag.String("progress", "STDOUT", "Options are: ['HIDDEN', 'FILE', 'STDOUT', 'BOTH']")
	progressFormatFlag = flag.String("progressFormat", DefaultProgressFormat, "Set the format for the output where 1st arg is the iteration and 2d arg is the amount of runs, 3d total, 4th start, 5th countdown, 6th est completion time.")
	outputFlag         = flag.String("output", "HIDDEN", "Options are: ['HIDDEN', 'FILE', 'STDOUT', 'BOTH']")
	fileFlag           = flag.String("file", "", "Path to the file where the report will be saved, configure file conflicts automatically with 'fileMode'")
	fileModeFlag       = flag.String("fileMode", "", "Configure how the report file should be treated. If a file exists, and this option isn't set, user will be queried. Options are: ['t'runcate, 'a'ppend] ")
	statisticsFlag     = flag.Bool("statistics", true, "Set to true if you don't wish to see statistics of the repeated command.")
	incrementFlag      = flag.Bool("increment", false, "Set to true and add an argument 'INC', to have 'INC' be replaced with the iteration. If increment == true && 'INC' is not set, repeater will panic.")
	resultFlag         = flag.String("result", "", "Set this to some filename and get a json-formated output of all the performed tasks. This output is the basis of the statistics.")
	retryOnFailFlag    = flag.Bool("retryOnFail", true, "Set to true to retry failed commands, effectively making repeate run until all commands are successful.")
)

func main() {
	flag.Parse()
	args := flag.Args()
	useColor = !(useColor && *colorFlag)

	if len(args) < 1 {
		printErr(fmt.Sprintf("error: %v", "you need to supply at least 1 argument\n"))
		os.Exit(1)
	}
	c, err := New(
		*amRunsFlag,
		*workersFlag,
		args, output.New(progressFlag),
		*progressFormatFlag,
		output.New(outputFlag),
		*fileFlag,
		*fileModeFlag,
		*incrementFlag,
		*resultFlag,
		*retryOnFailFlag)

	if *verboseFlag {
		fmt.Printf("Operation:\n%v\n------\n", &c)
	}
	if err != nil {
		printErr(fmt.Sprintf("configuration error: %v\n", err))
		os.Exit(1)
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
			fmt.Printf("%s\n", &stats)
		}

		if c.resultFile != nil {
			slices.SortFunc(stats.Results, func(a, b Result) int {
				return int(a.Runtime) - int(b.Runtime)
			})
			bytes, err := json.Marshal(stats.Results)
			if err != nil {
				printErr(fmt.Sprintf("failed to marshal results: %v", err))
			} else {
				printOK(fmt.Sprintf("printing results to file: %v\n", c.resultFile.Name()))
				fmt.Fprintf(c.resultFile, "%v", string(bytes))
			}
		}
		printOK("The repeat, has been done. Farewell.\n")
		os.Exit(0)
	case <-signalChannel:
	}
	ctxCancel()
	select {
	case <-isDone:
		if ctx.Err() != nil {
			printOK("graceful shutdown complete")
		}
	case <-signalChannel:
		printErr("aborting graceful shutdown")
	}
}
