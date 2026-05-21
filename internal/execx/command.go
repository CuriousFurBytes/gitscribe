package execx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Result struct {
	Command  []string
	Stdout   string
	Stderr   string
	ExitCode int
}

func (r Result) Output() string {
	switch {
	case r.Stdout == "":
		return r.Stderr
	case r.Stderr == "":
		return r.Stdout
	default:
		return r.Stdout + "\n" + r.Stderr
	}
}

func Run(ctx context.Context, dir string, stdin []byte, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "GH_PAGER=cat", "PAGER=cat")

	if len(stdin) > 0 {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{
		Command:  append([]string{name}, args...),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("%s exited with code %d", name, result.ExitCode)
	}

	result.ExitCode = -1
	return result, err
}
