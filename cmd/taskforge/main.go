package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kakasuarez/taskforge/internal/config"
	"github.com/kakasuarez/taskforge/internal/runner"
)

func main() {

	filePath := flag.String("file", "", "Path to the config file")
	flag.Parse()
	if *filePath == "" {
		flag.Usage()
		os.Exit(2)
	}
	task, err := config.Load(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while loading TOML: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("running task %+v\n", task)
	result, err := runner.Run(task.Command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred while running task: %s: %s\n", task.Name, err)
		os.Exit(1)
	}
	fmt.Printf("task:      %s\n", task.Name)
	fmt.Printf("started:   %s\n", result.Start.Format(time.TimeOnly))
	fmt.Printf("duration:  %s\n", result.Duration)
	fmt.Printf("exit code: %d\n", result.ExitCode)
	fmt.Printf("stdout:    %q\n", result.Stdout)
	fmt.Printf("stderr:    %q\n", result.Stderr)
	os.Exit(result.ExitCode)
}
