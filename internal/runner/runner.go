package runner

import (
	"bytes"
	"errors"
	"os/exec"
	"time"
)

// Result is a struct for a result of any executed task
type Result struct {
	Start    time.Time
	Duration time.Duration
	ExitCode int
	Stdout   string
	Stderr   string
}

// Run takes a command and runs it and records the output and error along with execution time
func Run(command []string) (Result, error) {
	if len(command) == 0 {
		return Result{}, errors.New("command cannot be empty")
	}
	var result Result
	cmd := exec.Command(command[0], command[1:]...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	result.Start = time.Now()
	err := cmd.Run()
	result.Duration = time.Since(result.Start)
	result.Stdout = outBuf.String()
	result.Stderr = errBuf.String()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return Result{}, err
		}
	}
	return result, nil
}
