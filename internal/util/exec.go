package util

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ExecuteCommand executes a command with a timeout
func ExecuteCommand(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// ExecuteCommandWithTimeout executes a command with a specific timeout duration
func ExecuteCommandWithTimeout(timeout time.Duration, name string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return ExecuteCommand(ctx, name, args...)
}

// ExecuteCommandWithStdin executes a command with stdin input
func ExecuteCommandWithStdin(ctx context.Context, input string, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewBufferString(input)

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// CheckCommandExists checks if a command exists in PATH
func CheckCommandExists(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("command '%s' not found in PATH", name)
	}
	return nil
}
