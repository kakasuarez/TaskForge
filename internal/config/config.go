package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// Task describes a command-line task
type Task struct {
	Name    string   `toml:"name"`
	Command []string `toml:"command"`
}

// Load reads the path and loads the task in Task
func Load(path string) (Task, error) {
	var task Task
	metadata, err := toml.DecodeFile(path, &task)
	if err != nil {
		return Task{}, fmt.Errorf("failed to load config: %s: %w", path, err)
	}
	var unrecog []string
	for _, key := range metadata.Undecoded() {
		unrecog = append(unrecog, key.String())
	}
	if len(unrecog) > 0 {
		return Task{}, fmt.Errorf("unrecognized fields present in file %s: %s", path, strings.Join(unrecog, ", "))
	}
	if task.Name == "" {
		return Task{}, fmt.Errorf("task name cannot be empty in file %s", path)
	}
	if len(task.Command) == 0 || task.Command[0] == "" {
		return Task{}, fmt.Errorf("task command cannot be empty in file %s", path)
	}
	return task, nil
}
