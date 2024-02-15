# repeater
A tool which repeats a command n amounts of time, with paralellisation and slight tweaks.

Test coverage: 62.599999999999994% 😌👏

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
    -file ./run_output `  # with file ./run_output` \
    -output FILE `       # with command output written to some file` \
    -progress BOTH `            # with progress written to BOTH stdout and some file` \
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
