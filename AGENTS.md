# AGENTS.md

You're working on a project called `repeater`.

## Always read

- `./main.go` — this contains CLI flags and the primary execution flow.
- `./go.mod` — this shows which libraries are used; do not add additional third-party libraries unless explicitly requested.
- `./README.md` — this contains the project usage model, examples, and expected operator-facing behavior.
- `./configured_oper.go` — this defines the main runtime configuration and core invariants.
- `./configured_oper_work.go` — this contains worker orchestration, cancellation, and result collection logic.
- `./statistics.go` — this defines result aggregation and user-visible statistics output.

## Way of work

- Always write tests first, implementation second.
- When fixing a bug, validate the issue with a test, then fix the implementation.
- Prefer minimal diffs over broad refactors.
- Preserve CLI behavior unless the task explicitly changes it.
- For concurrency or cancellation changes, verify both correctness and shutdown behavior.

## Validation

Before you're done, ensure that these pass:

- `go test ./... -race -cover -timeout=10s`
- `go run honnef.co/go/tools/cmd/staticcheck@latest ./...`
- `go run mvdan.cc/gofumpt@latest -w .`

ALWAYS TEST WITH TIMEOUT. Otherwise you may deadlock debugging efforts.