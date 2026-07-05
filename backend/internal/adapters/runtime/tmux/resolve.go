package tmux

import (
	"fmt"
	"os"
	"os/exec"
)

// BinarySource describes how ResolveBinary picked the tmux executable.
type BinarySource string

const (
	BinarySourceExplicit BinarySource = "explicit" // AO_TMUX_BIN
	BinarySourcePath     BinarySource = "path"     // found on PATH
	BinarySourceBundled  BinarySource = "bundled"  // AO_BUNDLED_TMUX_BIN from the desktop app
	BinarySourceMissing  BinarySource = "missing"
)

// BinaryDeps injects filesystem lookup for ResolveBinary tests.
type BinaryDeps struct {
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)
}

// DefaultBinaryDeps returns production deps backed by exec.LookPath and os.Stat.
func DefaultBinaryDeps() BinaryDeps {
	return BinaryDeps{LookPath: exec.LookPath, Stat: os.Stat}
}

// ResolveBinary picks the tmux executable using this order:
//  1. AO_TMUX_BIN when set (explicit operator override)
//  2. tmux on PATH when present and executable (preserves a user's own install)
//  3. AO_BUNDLED_TMUX_BIN when set and executable (desktop app bundle fallback)
//
// An empty path with BinarySourceMissing means no candidate was found.
func ResolveBinary(explicit, bundled string, deps BinaryDeps) (path string, source BinarySource) {
	if deps.LookPath == nil {
		deps.LookPath = exec.LookPath
	}
	if deps.Stat == nil {
		deps.Stat = os.Stat
	}

	if explicit != "" {
		if isExecutable(explicit, deps) {
			return explicit, BinarySourceExplicit
		}
		return "", BinarySourceMissing
	}
	if path, err := deps.LookPath("tmux"); err == nil && isExecutable(path, deps) {
		return path, BinarySourcePath
	}
	if bundled != "" && isExecutable(bundled, deps) {
		return bundled, BinarySourceBundled
	}
	return "", BinarySourceMissing
}

// SourceLabel formats a resolved binary for doctor output.
func SourceLabel(source BinarySource) string {
	switch source {
	case BinarySourceExplicit:
		return "explicit AO_TMUX_BIN"
	case BinarySourcePath:
		return "PATH"
	case BinarySourceBundled:
		return "bundled"
	default:
		return ""
	}
}

func isExecutable(path string, deps BinaryDeps) bool {
	info, err := deps.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}

// ResolveBinaryMessage builds a doctor-style status line for a resolved binary.
func ResolveBinaryMessage(path, version string, source BinarySource) string {
	label := SourceLabel(source)
	if label == "" {
		return fmt.Sprintf("%s (%s)", path, version)
	}
	return fmt.Sprintf("%s (%s; %s)", path, version, label)
}
