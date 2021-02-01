package tmux

import (
	. "github.com/ivaaaan/smug/pkg/commander"
	"os"
	"os/exec"
)

const (
	VSplit = "vertical"
	HSplit = "horizontal"
)

const (
	EvenHorizontal = "even-horizontal"
	EvenVertical   = "even-vertical"
	MainHorizontal = "main-horizontal"
	MainVertical   = "main-vertical"
	Tiled          = "tiled"
)

type Tmux struct {
	Commander Commander
}

func (tmux Tmux) NewSession(name string, root string, windowName string) (string, error) {
	cmd := exec.Command("tmux", "new", "-Pd", "-s", name, "-n", windowName, "-c", root)
	return tmux.Commander.Exec(cmd)
}

func (tmux Tmux) SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	res, err := tmux.Commander.Exec(cmd)
	return res == "" && err == nil
}

func (tmux Tmux) KillWindow(target string) error {
	cmd := exec.Command("tmux", "kill-window", "-t", target)
	_, err := tmux.Commander.Exec(cmd)
	return err
}

func (tmux Tmux) NewWindow(target string, name string, root string) (string, error) {
	cmd := exec.Command("tmux", "neww", "-Pd", "-t", target, "-n", name, "-c", root)

	return tmux.Commander.Exec(cmd)
}

func (tmux Tmux) SendKeys(target string, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, command, "Enter")
	return tmux.Commander.ExecSilently(cmd)
}

func (tmux Tmux) Attach(target string, stdin *os.File, stdout *os.File, stderr *os.File) error {
	cmd := exec.Command("tmux", "attach", "-d", "-t", target)

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return tmux.Commander.ExecSilently(cmd)
}

func (tmux Tmux) RenumberWindows(target string) error {
	cmd := exec.Command("tmux", "move-window", "-r", "-s", target, "-t", target)
	_, err := tmux.Commander.Exec(cmd)
	return err
}

func (tmux Tmux) SplitWindow(target string, splitType string, root string) (string, error) {
	args := []string{"split-window", "-Pd"}

	switch splitType {
	case VSplit:
		args = append(args, "-v")
	case HSplit:
		args = append(args, "-h")
	}

	args = append(args, []string{"-t", target, "-c", root, "-F", "#{pane_id}"}...)

	cmd := exec.Command("tmux", args...)

	pane, err := tmux.Commander.Exec(cmd)
	if err != nil {
		return "", err
	}

	return pane, nil
}

func (tmux Tmux) SelectLayout(target string, layoutType string) (string, error) {
	cmd := exec.Command("tmux", "select-layout", "-t", target, layoutType)
	return tmux.Commander.Exec(cmd)
}

func (tmux Tmux) StopSession(target string) (string, error) {
	cmd := exec.Command("tmux", "kill-session", "-t", target)
	return tmux.Commander.Exec(cmd)
}

func (tmux Tmux) SwitchClient(target string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", target)
	return tmux.Commander.ExecSilently(cmd)
}
