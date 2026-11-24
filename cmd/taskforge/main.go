package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kakasuarez/taskforge/internal/config"
	"github.com/kakasuarez/taskforge/internal/runner"
	"github.com/kakasuarez/taskforge/internal/store"
)

func prepareDBPath() (string, error) {
	var dbDirectory string
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dbDirectory = filepath.Join(homeDir, ".local", "share")
	} else {
		dbDirectory = xdgDataHome
	}
	dbPath := filepath.Join(dbDirectory, "taskforge", "taskforge.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return "", fmt.Errorf("create db directory: %w", err)
	}
	return dbPath, nil
}
func run() int {
	filePath := flag.String("file", "", "Path to the config file")
	flag.Parse()
	if *filePath == "" {
		flag.Usage()
		return 2
	}
	absFilePath, err := filepath.Abs(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while converting to absolute path: %s\n", err)
		return 1
	}
	task, err := config.Load(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while loading TOML: %s\n", err)
		return 1
	}
	dbPath, err := prepareDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while building DB path: %s\n", err)
		return 1
	}
	runs, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while opening DB: %s\n", err)
		return 1
	}
	defer runs.Close()
	attemptStart := time.Now()
	result, err := runner.Run(task.Command)
	var recordError string
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while running task: %s: %s\n", task.Name, err)
		recordError = err.Error()
		result.Start = attemptStart
	} else {
		fmt.Printf("task:      %s\n", task.Name)
		fmt.Printf("started:   %s\n", result.Start.Format(time.TimeOnly))
		fmt.Printf("duration:  %s\n", result.Duration)
		fmt.Printf("exit code: %d\n", result.ExitCode)
		fmt.Printf("stdout:    %q\n", result.Stdout)
		fmt.Printf("stderr:    %q\n", result.Stderr)
	}

	taskRecord := store.Record{
		TaskName: task.Name,
		TaskFile: absFilePath,
		Command:  task.Command,
		Start:    result.Start,
		Duration: result.Duration,
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Error:    recordError,
	}
	_, err = runs.Save(taskRecord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while saving record: %s\n", err)
		return 1
	}
	if recordError != "" {
		return 1
	}
	return result.ExitCode
}
func main() {
	os.Exit(run())
}
