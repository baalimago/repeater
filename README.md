# repeater
A tool which repeats a command n amounts of time. Includes progress and error handling.

### Installation
```bash
go install github.com/baalimago/repeater@latest
```

### Usage
```bash
repeater -n 1000 <cmd>
```

## Roadmap
- [x] Synchronous executions with output to stdout (mimc a bash for loop)
- [x] Progress on multiline, and singleline
- [ ] Parallelized execution
- [ ] Execution reports when done
