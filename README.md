# repeater
A tool which repeats a command n amounts of time.
Includes progress and error handling.

![repeatoopher](./img/repeatoopher.jpg)

### Installation
```bash
go install github.com/baalimago/repeater@latest
```

## Usage
Repeater is designed to perform repetitive tasks with slight modification.
Personally I've used it to CRUD state in paralell and fill login event logs as fast as I have network sockets.

```bash
# This command will repeat 'curl example.com' 100 times, using 10 workers and report progress to stdout
# The output of curl will then be printed to the report file
repeater -n 100 -w 10 -reportFile ./run_output -output REPORT_FILE -progress BOTH curl example.com

# This will print "this is increment 1234..."
repeater -n 100 -increment echo "this is increment: " INC

# Show all avaliable flags 
repeater -h
```

## Roadmap
- [x] Synchronous executions with output to stdout (mimc a bash for loop)
- [x] Progress on multiline, and singleline
- [x] Parallelized execution
- [x] Execution reports when done
