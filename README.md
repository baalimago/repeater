# repeater
A tool which repeats a command n amounts of time.

![repeatoopher](./img/repeatoopher.jpg)

### Installation
```bash
go install github.com/baalimago/repeater@latest
```

## Usage
Repeater is designed to perform repetitive tasks with slight modification.
Personally I've used it to CRUD state in parallel and fill login event logs as fast as I have network sockets.
It can probably also be used as a ghetto benchmarking tool.

```bash
repeater \
    -n 100 `# repeat 100 times` \
    -w 10 ` # using 10 workers` \
    -reportFile ./run_output `  # with reportfile ./run_output` \
    -output REPORT_FILE `       # with command output written to report file` \
    -progress BOTH `            # with progress written to BOTH stdout and report file` \
    curl example.com `          # command to repeat`
```

```bash
# This will panic since the curl will return non 0 exit code, the command's error file 
# will be written to -output, which by default is stdout
repeater -n 100 -w 10 curl wadiwaudbwadiubwada

# This will print "this is increment: 1\nthis is increment: 2\n..."
repeater -n 100 -increment echo "this is increment: INC"

# Show all available flags
repeater -h
```

## Roadmap
- [x] Synchronous executions with output to stdout (mimc a bash for loop)
- [x] Progress on multiline, and singleline
- [x] Parallelized execution
- [x] Execution reports when done
