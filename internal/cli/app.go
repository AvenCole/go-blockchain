package cli

import (
	"fmt"
	"io"
	"strings"

	"go-blockchain/internal/config"
)

const version = "0.1.0"

// App provides the minimal CLI skeleton approved in Plan 1.
type App struct {
	cfg    config.Config
	stdout io.Writer
	stderr io.Writer
}

// NewApp creates the initialization-phase CLI entrypoint.
func NewApp(cfg config.Config, stdout io.Writer, stderr io.Writer) App {
	return App{
		cfg:    cfg,
		stdout: stdout,
		stderr: stderr,
	}
}

// Run executes one command and returns a process exit code.
func (a App) Run(args []string) int {
	if len(args) == 0 || isHelpArg(args[0]) {
		a.printHelp()
		return 0
	}

	switch strings.ToLower(args[0]) {
	case "help":
		a.printHelp()
		return 0
	case "version", "about":
		a.printVersion()
		return 0
	case "doctor":
		a.printDoctor()
		return 0
	default:
		fmt.Fprintf(a.stderr, "unknown command: %s\n\n", args[0])
		a.printHelp()
		return 1
	}
}

func isHelpArg(arg string) bool {
	switch strings.ToLower(arg) {
	case "-h", "--help", "/?":
		return true
	default:
		return false
	}
}

func (a App) printHelp() {
	fmt.Fprintf(a.stdout, "%s initialization CLI\n\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout, "Usage:")
	fmt.Fprintf(a.stdout, "  %s [command]\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	fmt.Fprintln(a.stdout, "  help      Show command help")
	fmt.Fprintln(a.stdout, "  version   Show project version information")
	fmt.Fprintln(a.stdout, "  about     Show project version information")
	fmt.Fprintln(a.stdout, "  doctor    Check initialization-stage readiness")
}

func (a App) printVersion() {
	fmt.Fprintf(a.stdout, "%s version %s\n", a.cfg.ProjectName, version)
	fmt.Fprintf(a.stdout, "default data dir: %s\n", a.cfg.DataDir)
	fmt.Fprintf(a.stdout, "default port: %d\n", a.cfg.DefaultPort)
}

func (a App) printDoctor() {
	fmt.Fprintf(a.stdout, "doctor: initialization skeleton is ready\n")
	fmt.Fprintf(a.stdout, "project=%s\n", a.cfg.ProjectName)
	fmt.Fprintf(a.stdout, "data_dir=%s\n", a.cfg.DataDir)
	fmt.Fprintf(a.stdout, "default_port=%d\n", a.cfg.DefaultPort)
	fmt.Fprintf(a.stdout, "log_level=%s\n", a.cfg.LogLevel)
	fmt.Fprintf(a.stdout, "network_mode=%s\n", a.cfg.NetworkMode)
	fmt.Fprintln(a.stdout, "next_step=implement basic blockchain")
}
