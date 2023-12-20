# repeater
A tool which repeats a command n amounts of time, with paralellisation and slight tweaks.

![repeatoopher](./img/repeatoopher.jpg)

### Installation
```bash
go install github.com/baalimago/repeater@latest
```

## Usage
Usecases: 
* CRUD state using curl as fast as you have network sockets
* Paralellize repetitive shell-scripts
* Ghetto benchmarking

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
# This curl will exit with non 0 exit code, which will error 10 times,
# once per worker. When a worker errors, it will commit soduku. 
# Repeater panics once all workers are dead. 
#
# The curl stdout and stderr will both be written to `-output`
# destination, which by default is stdout
repeater -n 100 -w 10 curl foobar.raboof

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
