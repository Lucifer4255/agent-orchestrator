// Package probe checks host prerequisites for the selected terminal runtime.
package probe

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

const availabilityProbeTimeout = 2 * time.Second

// Prober resolves binaries and runs short version probes. Inject in tests.
type Prober struct {
	LookPath func(name string) (string, error)
	Run      func(ctx context.Context, path string, args ...string) ([]byte, error)
}

// Availability reports whether the selected terminal runtime can start sessions.
type Availability struct {
	Available   bool   `json:"available"`
	Runtime     string `json:"runtime"`
	Platform    string `json:"platform"`
	Message     string `json:"message,omitempty"`
	InstallHint string `json:"installHint,omitempty"`
	Path        string `json:"path,omitempty"`
	Version     string `json:"version,omitempty"`
}

// DefaultProber uses exec.LookPath and a short-lived exec probe.
func DefaultProber() Prober {
	return Prober{
		LookPath: exec.LookPath,
		Run: func(ctx context.Context, path string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, path, args...)
			return cmd.CombinedOutput()
		},
	}
}

// TmuxOnPath reports whether tmux resolves via LookPath. Spawn uses this fast
// check; AvailabilityStatus runs a version probe for doctor/API surfaces.
func TmuxOnPath(lookPath func(name string) (string, error)) bool {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	path, err := lookPath("tmux")
	return err == nil && path != ""
}

// AvailabilityStatus checks the platform runtime: tmux on Darwin/Linux, ConPTY on Windows.
func AvailabilityStatus(ctx context.Context, prober Prober) Availability {
	if runtime.GOOS == "windows" {
		return Availability{
			Available: true,
			Runtime:   "conpty",
			Platform:  runtime.GOOS,
			Message:   "ConPTY (built-in): no external terminal multiplexer required on Windows",
		}
	}
	return probeTmux(ctx, prober)
}

func probeTmux(ctx context.Context, prober Prober) Availability {
	hint := TmuxInstallHint()
	base := Availability{
		Runtime:     "tmux",
		Platform:    runtime.GOOS,
		InstallHint: hint,
	}
	if prober.LookPath == nil {
		prober.LookPath = exec.LookPath
	}
	path, err := prober.LookPath("tmux")
	if err != nil || path == "" {
		base.Available = false
		base.Message = "tmux not found in PATH; required on macOS/Linux to start agent sessions"
		return base
	}
	base.Path = path

	reqCtx, cancel := context.WithTimeout(ctx, availabilityProbeTimeout)
	defer cancel()
	out, err := proberRun(reqCtx, prober, path, "-V")
	if err != nil {
		base.Available = false
		base.Message = fmt.Sprintf("tmux found at %s but is not usable: %v", path, err)
		return base
	}
	version := firstOutputLine(out)
	if version == "" {
		version = "version unknown"
	}
	base.Available = true
	base.Version = version
	base.Message = fmt.Sprintf("%s (%s)", path, version)
	return base
}

func proberRun(ctx context.Context, prober Prober, path string, args ...string) ([]byte, error) {
	if prober.Run != nil {
		return prober.Run(ctx, path, args...)
	}
	cmd := exec.CommandContext(ctx, path, args...)
	return cmd.CombinedOutput()
}

// TmuxInstallHint returns platform-specific install guidance for tmux.
func TmuxInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install tmux"
	default:
		return "sudo apt install tmux    # Debian/Ubuntu\nsudo dnf install tmux    # Fedora/RHEL"
	}
}

// PrerequisiteError reports missing or unusable tmux with install guidance.
func PrerequisiteError() error {
	return fmt.Errorf(
		"%w: tmux required on macOS/Linux to run agent sessions but not in PATH; install with: %s",
		ports.ErrRuntimePrerequisite,
		strings.Split(TmuxInstallHint(), "\n")[0],
	)
}

// WrapExecError maps tmux exec failures (missing binary) to ErrRuntimePrerequisite.
func WrapExecError(err error) error {
	if err == nil {
		return nil
	}
	if isExecutableMissing(err) {
		return PrerequisiteError()
	}
	return err
}

func isExecutableMissing(err error) bool {
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "executable file not found") ||
		strings.Contains(msg, "no such file or directory")
}

func firstOutputLine(out []byte) string {
	line, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(line)
}
