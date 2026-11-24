# TaskForge
A lightweight alternative to a combination of cron, a task runner and a task-execution monitor

## Status

Early development (v1 in progress).

- [x] Read a task definition from a TOML file
- [x] Run the task and capture stdout, stderr, exit code, start time and duration
- [x] Record each run in SQLite, including runs that failed to start
- [ ] Read runs back from SQLite (`Recent`)
- [ ] Override the database location (`TASKFORGE_DB`)
- [ ] Automated tests

## Roadmap

### Next

1. **`Recent(n)` in `store`**: read the latest runs back, using `sql.Null[T]` for the nullable columns (`duration_ns`, `exit_code`, `error`) and converting them back into a `Record`.
2. **Tests**:
   - `store_test.go`: open a database in `t.TempDir()`, save runs (including a failed start), read them back with `Recent` and compare.
   - `config` and `runner` tests: turn the files in `examples/` into test cases under each package's `testdata/` directory.
3. **`TASKFORGE_DB` override**: an environment variable that sets the database path, so tests and cron jobs don't write to the real history. Priority: `TASKFORGE_DB` > XDG default.

### Later

- **Timeouts**: stop tasks that run too long (`exec.CommandContext`).
- **Signal exit codes**: exit with `128 + signal number` for signal-killed tasks instead of 255.
- **Live output**: stream stdout/stderr to the terminal while still capturing them (`io.MultiWriter`).
- **Readable command JSON**: store `>&` as-is instead of escaped (`json.Encoder` with `SetEscapeHTML(false)`).
- **History command**: show past runs from the database.
- **Schema migrations**: upgrade an existing database when the schema changes.
- **Scheduling**: run tasks on a schedule, the cron part of TaskForge.

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

## Run history

Every run is saved to a SQLite database:

```
$XDG_DATA_HOME/taskforge/taskforge.db     (if XDG_DATA_HOME is set)
~/.local/share/taskforge/taskforge.db     (otherwise)
```

The directory is created automatically (permissions `0700`) the first time a task runs. Nothing is created when the flags or the task file are invalid.

Each row in the `runs` table holds:

| Column | Meaning |
|---|---|
| `id` | assigned by SQLite |
| `task_name` | `name` from the task file |
| `task_file` | absolute path of the task file |
| `command` | the command as a JSON array, e.g. `["ls","-la"]` |
| `start_unix_ms` | start time, Unix milliseconds |
| `duration_ns` | duration in nanoseconds, `NULL` if the task never started |
| `exit_code` | task's exit code, `NULL` if the task never started |
| `stdout`, `stderr` | captured output |
| `error` | why the task could not be started, `NULL` if it started |

A task that **could not be started** (e.g. program not found) is still recorded: `error` is set, and `exit_code` and `duration_ns` are `NULL`. A `CHECK` constraint enforces that every row is one or the other.

To look at the history with the `sqlite3` command-line tool:

```sh
sqlite3 -header ~/.local/share/taskforge/taskforge.db \
  'SELECT id, task_name, exit_code, duration_ns, error FROM runs'
```

## Exit codes

| Situation | TaskForge exits with |
|---|---|
| Task ran | the task's own exit code |
| `-file` flag missing | `2` (usage error) |
| Invalid task file (missing, bad TOML, unknown keys, empty name or command) | `1` |
| Task could not be started (e.g. program not found). The run is still recorded. | `1` |
| Database could not be created, opened or written | `1` |

## Known limitations

- **Tasks killed by a signal exit as 255.** When a task is terminated by a signal (e.g. `kill -9`), Go reports its exit code as `-1`. TaskForge passes that to `os.Exit(-1)`, and since Unix keeps only the lowest 8 bits of an exit code, the shell sees **255**. The usual shell convention is `128 + signal number` (e.g. `137` for SIGKILL), which TaskForge does not follow yet.
  The run history stores the `-1` as the exit code.
- **Exit code 1 is ambiguous.** TaskForge's own errors and a task that genuinely exits with 1 both produce `1`.
- **A failed save hides the task's exit code.** If the run can't be written to the database, TaskForge exits with `1` even if the task itself succeeded.
- **The database location can't be changed yet**, other than through `XDG_DATA_HOME`. A `TASKFORGE_DB` override is planned.
- **No schema migrations.** The table is created with `CREATE TABLE IF NOT EXISTS`, so an existing database keeps its old schema when the schema changes. During development, delete the database file and let TaskForge create it again.
- **`<`, `>` and `&` in commands are escaped in the stored JSON** (e.g. `>&2` is stored as `>&2`). This is valid JSON and decodes back correctly, but it is harder to read in `sqlite3`.
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
| `error.toml` | program not found (recorded with `error` set) |
| `nocomm.toml` | missing `command` |
| `noname.toml` | missing `name` |
| `unrecog.toml` | unknown key |

## Project layout

```
cmd/taskforge/      entry point: flags, database location, wires the packages together
internal/config/    loads and validates a task from TOML
internal/runner/    runs a command and returns its Result
internal/store/     SQLite persistence: opens the database, creates the schema, saves runs
  schema.sql          the runs table (embedded into the binary)
  insert_record.sql   the INSERT statement (embedded into the binary)
examples/           sample task files
```
