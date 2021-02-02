package commander

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type ShellError struct {
	Command string
	Err     error
}

func (e *ShellError) Error() string {
	return fmt.Sprintf("Cannot run %q. Error %v", e.Command, e.Err)
}

type Commander interface {
	Exec(cmd *exec.Cmd) (string, error)
	ExecSilently(cmd *exec.Cmd) error
}

type DefaultCommander struct {
	Logger *log.Logger
}

func (c DefaultCommander) Exec(cmd *exec.Cmd) (string, error) {
	if c.Logger != nil {
		c.Logger.Println(strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println(err)
		}
		return "", &ShellError{strings.Join(cmd.Args, " "), err}
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}

func (c DefaultCommander) ExecSilently(cmd *exec.Cmd) error {
	if c.Logger != nil {
		c.Logger.Println(strings.Join(cmd.Args, " "))
	}

	err := cmd.Run()
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println(err)
		}
		return &ShellError{strings.Join(cmd.Args, " "), err}
	}
	return nil
}