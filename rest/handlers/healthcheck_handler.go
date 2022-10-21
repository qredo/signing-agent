package handlers

import (
	"net/http"

	"gitlab.qredo.com/custody-engine/signing-agent/api"
	"gitlab.qredo.com/custody-engine/signing-agent/config"
	"gitlab.qredo.com/custody-engine/signing-agent/defs"
	"gitlab.qredo.com/custody-engine/signing-agent/hub"
	"gitlab.qredo.com/custody-engine/signing-agent/rest/version"
)

type HealthCheckHandler struct {
	version      *version.Version
	config       *config.Config
	source       hub.SourceStats
	feedClients  hub.ConnectedClients
	localFeedUrl string
}

func NewHealthCheckHandler(source hub.SourceStats, version *version.Version, config *config.Config, feedHub hub.ConnectedClients, localFeed string) *HealthCheckHandler {
	return &HealthCheckHandler{
		source:       source,
		version:      version,
		config:       config,
		feedClients:  feedHub,
		localFeedUrl: localFeed,
	}
}

// HealthCheckVersion
//
// swagger:route GET /healthcheck/version
//
// Check application version.
func (h *HealthCheckHandler) HealthCheckVersion(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.version
	return response, nil
}

// HealthCheckConfig
//
// swagger:route GET /healthcheck/config
//
// Check application configuration.
func (h *HealthCheckHandler) HealthCheckConfig(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.config
	return response, nil
}

// HealthCheckStatus
//
// swagger:route GET /healthcheck/status
//
// Check application status.
func (h *HealthCheckHandler) HealthCheckStatus(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")

	readyState := h.source.GetReadyState()
	sourceFeedUrl := h.source.GetFeedUrl()
	connectedFeedClients := h.feedClients.GetExternalFeedClients()

	response := api.HealthCheckStatusResponse{
		WebsocketStatus: api.NewWebsocketStatus(readyState, sourceFeedUrl, h.localFeedUrl, connectedFeedClients),
	}

	return response, nil
}
