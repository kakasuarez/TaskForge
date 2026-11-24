CREATE TABLE IF NOT EXISTS runs (
  id INTEGER PRIMARY KEY,
  task_name TEXT NOT NULL,
  task_file TEXT NOT NULL,
  command TEXT NOT NULL,
  start_unix_ms INTEGER NOT NULL,
  duration_ns INTEGER,
  exit_code INTEGER,
  stdout TEXT,
  stderr TEXT,
  error TEXT,
  CONSTRAINT started_xor_error CHECK ((exit_code IS NULL) != (error IS NULL) AND (error IS NULL) != (duration_ns IS NULL))
);
