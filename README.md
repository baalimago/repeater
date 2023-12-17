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
# This command will:
# * repeat 'curl example.com' 100 times using 10 workers
# * print output of the curl to the REPORT_FILE, which is './run_output'
# * print progress to both stdout and the REPORT_FILE 
# * it will create 'run_output' if it doesn't exist, and consult user on file conflict
repeater -n 100 -w 10 -reportFile ./run_output -output REPORT_FILE -progress BOTH curl example.com

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
