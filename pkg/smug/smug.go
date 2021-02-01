package smug

import (
	"fmt"
	. "github.com/ivaaaan/smug/pkg/commander"
	. "github.com/ivaaaan/smug/pkg/config"
	. "github.com/ivaaaan/smug/pkg/context"
	. "github.com/ivaaaan/smug/pkg/tmux"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultWindowName = "smug_def"

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return path
		}

		return strings.Replace(path, "~", userHome, 1)
	}

	return path
}

func Contains(slice []string, s string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}

	return false
}

type Smug struct {
	Tmux      Tmux
	Commander Commander
}

func (smug Smug) execShellCommands(commands []string, path string) error {
	for _, c := range commands {
		cmd := exec.Command("/bin/sh", "-c", c)
		cmd.Dir = path

		_, err := smug.Commander.Exec(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (smug Smug) switchOrAttach(target string, attach bool, insideTmuxSession bool) error {
	if insideTmuxSession && attach {
		return smug.Tmux.SwitchClient(target)
	} else if !insideTmuxSession {
		return smug.Tmux.Attach(target, os.Stdin, os.Stdout, os.Stderr)
	}
	return nil
}

func (smug Smug) Stop(config Config, options Options, context Context) error {
	windows := options.Windows
	if len(windows) == 0 {
		sessionRoot := ExpandPath(config.Root)

		err := smug.execShellCommands(config.Stop, sessionRoot)
		if err != nil {
			return err
		}
		_, err = smug.Tmux.StopSession(config.Session)
		return err
	}

	for _, w := range windows {
		err := smug.Tmux.KillWindow(config.Session + ":" + w)
		if err != nil {
			return err
		}
	}

	return nil
}

func (smug Smug) Start(config Config, options Options, context Context) error {
	sessionName := config.Session + ":"
	sessionExists := smug.Tmux.SessionExists(sessionName)
	sessionRoot := ExpandPath(config.Root)

	windows := options.Windows
	attach := options.Attach

	if !sessionExists {
		err := smug.execShellCommands(config.BeforeStart, sessionRoot)
		if err != nil {
			return err
		}

		_, err = smug.Tmux.NewSession(config.Session, sessionRoot, defaultWindowName)
		if err != nil {
			return err
		}
	} else if len(windows) == 0 {
		return smug.switchOrAttach(sessionName, attach, context.InsideTmuxSession)
	}

	for _, w := range config.Windows {
		if (len(windows) == 0 && w.Manual) || (len(windows) > 0 && !Contains(windows, w.Name)) {
			continue
		}

		windowRoot := ExpandPath(w.Root)
		if windowRoot == "" || !filepath.IsAbs(windowRoot) {
			windowRoot = filepath.Join(sessionRoot, w.Root)
		}

		window := sessionName + w.Name
		_, err := smug.Tmux.NewWindow(sessionName, w.Name, windowRoot)
		if err != nil {
			return err
		}

		for _, c := range w.Commands {
			err := smug.Tmux.SendKeys(window, c)
			if err != nil {
				return err
			}
		}

		layout := w.Layout
		if layout == "" {
			layout = EvenHorizontal
		}

		_, err = smug.Tmux.SelectLayout(sessionName+w.Name, layout)
		if err != nil {
			return err
		}

		for pIndex, p := range w.Panes {
			paneRoot := ExpandPath(p.Root)
			if paneRoot == "" || !filepath.IsAbs(p.Root) {
				paneRoot = filepath.Join(windowRoot, p.Root)
			}

			target := fmt.Sprintf("%s.%d", window, pIndex)
			newPane, err := smug.Tmux.SplitWindow(target, p.Type, paneRoot)
			if err != nil {
				return err
			}

			for _, c := range p.Commands {
				err = smug.Tmux.SendKeys(window+"."+newPane, c)
				if err != nil {
					return err
				}
			}
		}
	}

	smug.Tmux.KillWindow(sessionName + defaultWindowName)
	smug.Tmux.RenumberWindows(sessionName)

	if len(windows) == 0 && len(config.Windows) > 0 {
		return smug.switchOrAttach(sessionName+config.Windows[0].Name, attach, context.InsideTmuxSession)
	}

	return nil
}
