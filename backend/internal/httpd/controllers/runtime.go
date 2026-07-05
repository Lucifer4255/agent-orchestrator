package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/runtime/probe"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
)

// RuntimeController owns the /runtime routes.
type RuntimeController struct{}

// Register mounts runtime REST routes on the supplied router.
func (c *RuntimeController) Register(r chi.Router) {
	r.Get("/runtime/status", c.status)
}

func (c *RuntimeController) status(w http.ResponseWriter, r *http.Request) {
	avail := probe.AvailabilityStatus(r.Context(), probe.DefaultProber())
	envelope.WriteJSON(w, http.StatusOK, RuntimeStatusResponse{
		Available:   avail.Available,
		Runtime:     avail.Runtime,
		Platform:    avail.Platform,
		Message:     avail.Message,
		InstallHint: avail.InstallHint,
		Path:        avail.Path,
		Version:     avail.Version,
	})
}
