package tmux

import (
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestResolveBinaryPrefersExplicitOverride(t *testing.T) {
	deps := BinaryDeps{
		LookPath: func(string) (string, error) { return "/usr/bin/tmux", nil },
		Stat:     func(string) (os.FileInfo, error) { return fileInfo{mode: 0o755}, nil },
	}
	path, source := ResolveBinary("/custom/tmux", "/bundle/tmux", deps)
	if path != "/custom/tmux" || source != BinarySourceExplicit {
		t.Fatalf("ResolveBinary = %q %q, want /custom/tmux explicit", path, source)
	}
}

func TestResolveBinaryPrefersPathOverBundled(t *testing.T) {
	deps := BinaryDeps{
		LookPath: func(string) (string, error) { return "/usr/local/bin/tmux", nil },
		Stat:     func(string) (os.FileInfo, error) { return fileInfo{mode: 0o755}, nil },
	}
	path, source := ResolveBinary("", "/Applications/Agent Orchestrator.app/Contents/Resources/tmux/tmux", deps)
	if path != "/usr/local/bin/tmux" || source != BinarySourcePath {
		t.Fatalf("ResolveBinary = %q %q, want PATH candidate", path, source)
	}
}

func TestResolveBinaryUsesBundledWhenPathMissing(t *testing.T) {
	bundled := "/Applications/Agent Orchestrator.app/Contents/Resources/tmux/tmux"
	deps := BinaryDeps{
		LookPath: func(string) (string, error) { return "", fs.ErrNotExist },
		Stat: func(path string) (os.FileInfo, error) {
			if path == bundled {
				return fileInfo{mode: 0o755}, nil
			}
			return nil, fs.ErrNotExist
		},
	}
	path, source := ResolveBinary("", bundled, deps)
	if path != bundled || source != BinarySourceBundled {
		t.Fatalf("ResolveBinary = %q %q, want bundled candidate", path, source)
	}
}

func TestResolveBinaryMissingWhenNothingExecutable(t *testing.T) {
	deps := BinaryDeps{
		LookPath: func(string) (string, error) { return "", fs.ErrNotExist },
		Stat:     func(string) (os.FileInfo, error) { return nil, fs.ErrNotExist },
	}
	path, source := ResolveBinary("", "/missing/tmux", deps)
	if path != "" || source != BinarySourceMissing {
		t.Fatalf("ResolveBinary = %q %q, want missing", path, source)
	}
}

func TestResolveBinaryRejectsNonExecutableExplicit(t *testing.T) {
	deps := BinaryDeps{
		LookPath: func(string) (string, error) { return "/usr/bin/tmux", nil },
		Stat: func(path string) (os.FileInfo, error) {
			if path == "/custom/tmux" {
				return fileInfo{mode: 0o644}, nil
			}
			return fileInfo{mode: 0o755}, nil
		},
	}
	path, source := ResolveBinary("/custom/tmux", "/bundle/tmux", deps)
	if path != "" || source != BinarySourceMissing {
		t.Fatalf("ResolveBinary = %q %q, want missing for bad explicit override", path, source)
	}
}

type fileInfo struct {
	mode fs.FileMode
}

func (f fileInfo) Name() string       { return "tmux" }
func (f fileInfo) Size() int64        { return 0 }
func (f fileInfo) Mode() fs.FileMode  { return f.mode }
func (f fileInfo) ModTime() time.Time { return time.Time{} }
func (f fileInfo) IsDir() bool        { return false }
func (f fileInfo) Sys() any           { return nil }
