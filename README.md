# TaskForge
A lightweight alternative to a combination of cron, a task runner and a task-execution monitor

## Status

Early development (v1 in progress).

- [x] Read a task definition from a TOML file
- [x] Run the task and capture stdout, stderr, exit code, start time and duration
- [ ] Record each run in SQLite
- [ ] Automated tests

## Requirements

- Go 1.27+

## Build and run

```sh
go build -o taskforge ./cmd/taskforge
./taskforge -file examples/ls.toml
```

Or without building a binary:

```sh
go run ./cmd/taskforge -file examples/ls.toml
```

Note: `go run` always exits with 1 when the program fails, whatever code the program itself returned. Use the built binary to see real exit codes.

## Task file format

```toml
name = "List files"         # required, non-empty
command = ["ls", "-la"]     # required; first element is the program, the rest are its arguments
```

- The command is run directly, **not through a shell**. That means shell features such as `~`, `*`, `$VAR`, pipes and redirects don't work unless you call a shell explicitly:
  ```toml
  command = ["sh", "-c", "echo out; echo err >&2; exit 3"]
  ```
- Unknown keys (for example the typo `comand`) are rejected with an error listing them.

## Output

```
task:      List DNE
started:   18:02:43
duration:  4.283224ms
exit code: 2
stdout:    ""
stderr:    "ls: cannot access '/does-not-exist': No such file or directory\n"
```

## Exit codes

| Situation | TaskForge exits with |
|---|---|
| Task ran | the task's own exit code |
| `-file` flag missing | `2` (usage error) |
| Invalid task file (missing, bad TOML, unknown keys, empty name or command) | `1` |
| Task could not be started (e.g. program not found) | `1` |

## Known limitations

- **Tasks killed by a signal exit as 255.** When a task is terminated by a signal (e.g. `kill -9`), Go reports its exit code as `-1`. TaskForge passes that to `os.Exit(-1)`, and since Unix keeps only the lowest 8 bits of an exit code, the shell sees **255**. The usual shell convention is `128 + signal number` (e.g. `137` for SIGKILL), which TaskForge does not follow yet.
- **Exit code 1 is ambiguous.** TaskForge's own errors and a task that genuinely exits with 1 both produce `1`.
- **Output is buffered in memory.** stdout and stderr are captured in full and printed after the task finishes. There is no live streaming and no size limit.
- **No timeouts.** A task that hangs makes TaskForge hang.

## Examples

`examples/` contains task files that exercise each case:

| File | Case |
|---|---|
| `ls.toml` | succeeds (exit 0) |
| `sleep.toml` | measures a ~1s duration |
| `lsdne.toml` | task fails with a non-zero exit code (2) |
| `test.toml` | separate stdout/stderr capture, exit 3 |
| `error.toml` | program not found |
| `nocomm.toml` | missing `command` |
| `noname.toml` | missing `name` |
| `unrecog.toml` | unknown key |

## Project layout

```
cmd/taskforge/      entry point: flag parsing, wires the packages together
internal/config/    loads and validates a task from TOML
internal/runner/    runs a command and returns its Result
internal/store/     SQLite persistence (not implemented yet)
examples/           sample task files
```
