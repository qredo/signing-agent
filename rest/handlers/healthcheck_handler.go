package handlers

import (
	"net/http"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/rest/version"
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
// swagger:route GET /healthcheck/version healthcheck HealthcheckVersion
//
// # Check application version
//
// Responses:
//
//	200: VersionResponse
func (h *HealthCheckHandler) HealthCheckVersion(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.version
	return response, nil
}

// HealthCheckConfig
//
// swagger:route GET /healthcheck/config healthcheck HealthcheckConfig
//
// # Check application configuration
//
// Responses:
//
//	200: ConfigResponse
func (h *HealthCheckHandler) HealthCheckConfig(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.config
	return response, nil
}

// HealthCheckStatus
//
// swagger:route GET /healthcheck/status healthcheck HealthcheckStatus
//
// # Check application status
//
// Responses:
//
//	200: HealthCheckStatusResponse
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
