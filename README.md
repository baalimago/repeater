# repeater
A tool which repeats a command n amounts of time, with paralellisation and slight tweaks.

Test coverage: 73.15% 😌👏

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
    -file ./run_output `        # with file ./run_output` \
    -output FILE `              # with command output written to some file` \
    -progress BOTH `            # with progress written to BOTH stdout and some file` \
    curl example.com `          # command to repeat`
```

```bash
# This will print "this is increment: 1\nthis is increment: 2\n..."
repeater -n 100 -output STDOUT -progress HIDDEN -increment echo "this is increment: INC"

# Show all available flags
repeater -h
```

## Roadmap
- [x] Synchronous executions with output to stdout (mimc a bash for loop)
- [x] Progress on multiline, and singleline
- [x] Parallelized execution
- [x] Execution reports when done
