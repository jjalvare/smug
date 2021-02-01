package smug

import (
	"errors"
	"fmt"
	. "github.com/ivaaaan/smug/pkg/commander"
	. "github.com/ivaaaan/smug/pkg/config"
	. "github.com/ivaaaan/smug/pkg/context"
	. "github.com/ivaaaan/smug/pkg/tmux"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

const (
	CommandStart = "start"
	CommandStop  = "stop"
	CommandNew   = "new"
	CommandEdit  = "edit"
)

var validCommands = []string{CommandStart, CommandStop, CommandNew, CommandEdit}

type Options struct {
	Command string
	Project string
	Config  string
	Windows []string
	Attach  bool
	Debug   bool
}

var ErrHelp = errors.New("help requested")

const (
	WindowsUsage = "List of windows to start. If session exists, those windows will be attached to current session."
	AttachUsage  = "Force switch client for a session"
	DebugUsage   = "Print all commands to ~/.config/smug/smug.log"
	FileUsage    = "A custom path to a config file"
)

// Creates a new FlagSet.
// Moved it to a variable to be able to override it in the tests.
var NewFlagSet = func(cmd string) *pflag.FlagSet {
	f := pflag.NewFlagSet(cmd, pflag.ContinueOnError)
	return f
}

func RunOptions(options Options) {
	userConfigDir := filepath.Join(ExpandPath("~/"), ".config/smug")

	var configPath string
	if options.Config != "" {
		configPath = options.Config
	} else {
		configPath = filepath.Join(userConfigDir, options.Project+".yml")
	}

	var logger *log.Logger
	if options.Debug {
		logFile, err := os.Create(filepath.Join(userConfigDir, "smug.log"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		logger = log.New(logFile, "", 0)
	}

	commander := DefaultCommander{Logger: logger}
	tmux := Tmux{Commander: commander}
	smug := Smug{Tmux: tmux, Commander: commander}
	context := CreateContext()

	switch options.Command {
	case CommandStart:
		if len(options.Windows) == 0 {
			fmt.Println("Starting a new session...")
		} else {
			fmt.Println("Starting new windows...")
		}
		config, err := GetConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = smug.Start(config, options, context)
		if err != nil {
			fmt.Println("Oops, an error occurred! Rolling back...")
			smug.Stop(config, options, context)
			os.Exit(1)
		}
	case CommandStop:
		if len(options.Windows) == 0 {
			fmt.Println("Terminating session...")
		} else {
			fmt.Println("Killing windows...")
		}
		config, err := GetConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = smug.Stop(config, options, context)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	case CommandNew:
	case CommandEdit:
		err := EditConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

}

func ParseOptions(argv []string, helpRequested func()) (Options, error) {
	if len(argv) == 0 {
		helpRequested()
		return Options{}, ErrHelp
	}

	if argv[0] == "--help" || argv[0] == "-h" {
		helpRequested()
		return Options{}, ErrHelp
	}

	cmd := argv[0]
	if !Contains(validCommands, cmd) {
		helpRequested()
		return Options{}, ErrHelp
	}

	flags := NewFlagSet(cmd)

	config := flags.StringP("file", "f", "", FileUsage)
	windows := flags.StringArrayP("windows", "w", []string{}, WindowsUsage)
	attach := flags.BoolP("attach", "a", false, AttachUsage)
	debug := flags.BoolP("debug", "d", false, DebugUsage)

	err := flags.Parse(argv)
	if err == pflag.ErrHelp {
		return Options{}, ErrHelp
	}

	if err != nil {
		return Options{}, err
	}

	if len(argv) < 2 && *config == "" {
		helpRequested()
		return Options{}, ErrHelp
	}

	var project string
	if *config == "" {
		project = argv[1]
	}

	if strings.Contains(project, ":") {
		parts := strings.Split(project, ":")
		project = parts[0]
		wl := strings.Split(parts[1], ",")
		windows = &wl
	}

	return Options{cmd, project, *config, *windows, *attach, *debug}, nil
}
