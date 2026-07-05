package probe

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestAvailabilityStatusWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ConPTY availability is only reported on Windows")
	}
	avail := AvailabilityStatus(context.Background(), Prober{})
	if !avail.Available || avail.Runtime != "conpty" {
		t.Fatalf("availability = %+v, want available conpty", avail)
	}
}

func TestAvailabilityStatusTmuxMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux probe applies on Darwin/Linux only")
	}
	avail := AvailabilityStatus(context.Background(), Prober{
		LookPath: func(string) (string, error) { return "", exec.ErrNotFound },
	})
	if avail.Available || avail.Runtime != "tmux" || avail.InstallHint == "" {
		t.Fatalf("availability = %+v, want unavailable tmux with install hint", avail)
	}
	if !strings.Contains(avail.Message, "not found in PATH") {
		t.Fatalf("message = %q, want PATH guidance", avail.Message)
	}
}

func TestAvailabilityStatusTmuxOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux probe applies on Darwin/Linux only")
	}
	avail := AvailabilityStatus(context.Background(), Prober{
		LookPath: func(string) (string, error) { return "/bin/tmux", nil },
		Run: func(_ context.Context, name string, args ...string) ([]byte, error) {
			if name != "/bin/tmux" || len(args) != 1 || args[0] != "-V" {
				t.Fatalf("unexpected probe: %s %v", name, args)
			}
			return []byte("tmux 3.4\n"), nil
		},
	})
	if !avail.Available || avail.Path != "/bin/tmux" || avail.Version != "tmux 3.4" {
		t.Fatalf("availability = %+v, want available tmux 3.4", avail)
	}
}

func TestAvailabilityStatusTmuxVersionFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux probe applies on Darwin/Linux only")
	}
	avail := AvailabilityStatus(context.Background(), Prober{
		LookPath: func(string) (string, error) { return "/bin/tmux", nil },
		Run: func(context.Context, string, ...string) ([]byte, error) {
			return nil, errors.New("exec: tmux: not found")
		},
	})
	if avail.Available || !strings.Contains(avail.Message, "not usable") {
		t.Fatalf("availability = %+v, want unavailable on version probe failure", avail)
	}
}

func TestWrapExecError(t *testing.T) {
	wrapped := WrapExecError(errors.New(`exec: "tmux": executable file not found in $PATH`))
	if !errors.Is(wrapped, ports.ErrRuntimePrerequisite) {
		t.Fatalf("WrapExecError = %v, want ErrRuntimePrerequisite", wrapped)
	}
	if !strings.Contains(wrapped.Error(), "brew install tmux") && !strings.Contains(wrapped.Error(), "apt install tmux") {
		t.Fatalf("wrapped = %q, want install hint", wrapped.Error())
	}
}
