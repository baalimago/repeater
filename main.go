package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/exp/slog"
)

var amRuns = flag.Int("n", 1, "Amount of times you wish to repeat the command.")
var noColor = flag.Bool("nocolor", false, "Set to true to disable ANSI colors.")
var progress = flag.String("progress", "hidden", "Current options are: ['oneline', 'multiline', 'hidden']. Note that 'oneline' will print on multiple lines if stdout yields a newline.")
var output = flag.String("output", "stdout", "Set how the output of the commands will be shown. Current options are: ['stdout', 'hidden']")

type color int

const (
	RED    color = 31
	GREEN  color = 32
	YELLOW color = 33
)

func coloredMessage(c color, msg string) string {
	if *noColor {
		return msg
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, msg)
}

func ok(msg string) {
	fmt.Fprintf(os.Stderr, "%v: %v", coloredMessage(GREEN, "ok"), msg)
}

func error(msg string) {
	fmt.Fprintf(os.Stderr, "%v: %v", coloredMessage(RED, "error"), msg)
}

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		error("you need to supply a command\n")
		os.Exit(1)
	}

	for i := 0; i < *amRuns; i++ {
		cmd := exec.Command(args[0], args[1:]...)
		switch *output {
		case "stdout":
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		case "hidden":
		}
		if *progress != "" {
			var resultPadding string
			if *output == "stdout" {
				resultPadding = ", result: "
			}
			toPrint := fmt.Sprintf("run: (%v/%v)%v", i, *amRuns, resultPadding)
			switch *progress {
			case "multiline":
				slog.Info(toPrint)
			case "oneline":
				fmt.Printf("\r%v", toPrint)
			}
		}
		err := cmd.Run()
		if errors.Is(err, &exec.ExitError{}) {
			error(fmt.Sprintf("exit error found, aborting operations: %v", *err.(*exec.ExitError)))
			os.Exit(1)
		}

	}
}
