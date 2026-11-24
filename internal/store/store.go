package store

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

//go:embed insert_record.sql
var insertRecord string

// Store is a store for the SQLite database
type Store struct {
	db *sql.DB
}

// Record is a specific row in the SQLite table for task execution runs
type Record struct {
	TaskName, TaskFile string
	Command            []string
	Start              time.Time
	Duration           time.Duration
	ExitCode           int
	Stdout, Stderr     string
	Error              string
}

// Open is the constructor function for Store
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", path, err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("create schema in %s: %w", path, errors.Join(err, closeErr))
		}
		return nil, fmt.Errorf("create schema in %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

// Save inserts a row and returns its new id
func (store *Store) Save(record Record) (int64, error) {
	startTime := record.Start.UnixMilli()
	commandBytes, err := json.Marshal(record.Command)
	if err != nil {
		return 0, fmt.Errorf("marshal command: %w", err)
	}
	var errorRun, durationRun, exitCodeRun any
	if record.Error != "" {
		errorRun = record.Error
	} else {
		durationRun = int64(record.Duration)
		exitCodeRun = record.ExitCode
	}
	result, err := store.db.Exec(insertRecord, record.TaskName, record.TaskFile, string(commandBytes), startTime, durationRun, exitCodeRun, record.Stdout, record.Stderr, errorRun)
	if err != nil {
		return 0, fmt.Errorf("insert run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted id: %w", err)
	}
	return id, nil
}

// Close closes the Store
func (store *Store) Close() error {
	return store.db.Close()
}
