# repeater
A tool which repeats a command n amounts of time, with paralellisation and slight tweaks.

Test coverage: 60.480% 😌👏

![repeatoopher](./img/repeatoopher.jpg)

### Installation
```bash
go install github.com/baalimago/repeater@latest
```

You may also use the setup script:
```bash
curl -fsSL https://raw.githubusercontent.com/baalimago/repeater/main/setup.sh | sh
```

## Usage
Since v1.2.1, repeater will re-attempt the command until successful.
A success is a command returning exit code 0.
Set flag `-retryOnFail=false` if you simply wish to repeat the CMD `-n` amount of times, regardless of outcome.

Usecases: 
* CRUD state using curl as fast as you have network sockets
* Paralellize repetitive shell-scripts
* Ghetto benchmarking

```bash
repeater \
    -n 100 `# repeat 100 times` \
    -w 10 ` # using 10 workers` \
    -output FILE `              # with command output written to FILE` \
    -progress BOTH `            # with progress written to BOTH STDOUT and FILE` \
    -file ./run_output `        # with FILE ./run_output` \
    -result ./run_result `      # with result (output + time taken) for each command` \
    curl example.com `          # command to repeat`
```

```bash
# This will print "this is increment: 1\nthis is increment: 2\n..."
repeater -n 100 -output STDOUT -progress HIDDEN -increment echo "this is increment: INC"

# Show all available flags
repeater -h
```

## Benchmarks
`repeater` outperforms many other parallizers, including GNU parallel and xargs. 

Run `./benchmark.sh` to try out repeaters performance vs similar parallelization tools.
You may benchmark any command that you want, simply run `./benchmark.sh <MAX RUNS> <YOUR COMMAND>`.
It will run the command with `repeater`, `parallel` and `xargs` and print the time taken for each by starting to repeat the command 10 times then 100, 1000, etc up until `<MAX_RUNS>`.
Note that `./benchmark.sh` will break if it detects and diffs in the output of the commands, so ensure the commands output is deterministic.

