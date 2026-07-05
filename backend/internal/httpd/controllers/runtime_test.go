package controllers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/config"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/controllers"
)

func TestRuntimeStatus(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(httpd.NewRouterWithControl(config.Config{}, log, nil, httpd.APIDeps{}, httpd.ControlDeps{}))
	defer srv.Close()

	body, status, _ := doRequest(t, srv, http.MethodGet, "/api/v1/runtime/status", "")
	if status != http.StatusOK {
		t.Fatalf("GET runtime/status = %d, want 200; body=%s", status, body)
	}

	var got controllers.RuntimeStatusResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Platform != runtime.GOOS {
		t.Fatalf("platform = %q, want %q", got.Platform, runtime.GOOS)
	}
	if runtime.GOOS == "windows" {
		if !got.Available || got.Runtime != "conpty" {
			t.Fatalf("response = %+v, want available conpty on Windows", got)
		}
		return
	}
	if got.Runtime != "tmux" {
		t.Fatalf("runtime = %q, want tmux", got.Runtime)
	}
	if got.Available && got.Path == "" {
		t.Fatalf("available tmux response missing path: %+v", got)
	}
	if !got.Available && got.InstallHint == "" {
		t.Fatalf("unavailable tmux response missing installHint: %+v", got)
	}
}
