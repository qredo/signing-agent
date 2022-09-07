package rest

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
	"gitlab.qredo.com/custody-engine/automated-approver/util"

	"gitlab.qredo.com/custody-engine/automated-approver/lib"

	"github.com/gorilla/context"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"go.uber.org/zap"
)

const (
	PathPrefix = "/api/v1"

	PathHealthcheckVersion = "/healthcheck/version"
	PathHealthCheckConfig  = "/healthcheck/config"
	PathClientFullRegister = "/register"
	PathClientsList        = "/client"
	PathAction             = "/client/action/{action_id}"
	PathClientFeed         = "/client/feed"
)

func WrapPathPrefix(uri string) string {
	return strings.Join([]string{PathPrefix, uri}, "")
}

type appHandlerFunc func(ctx *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error)

func (a appHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := &defs.RequestContext{}

	resp, err := a(ctx, w, r)

	if strings.ToLower(r.Header.Get("connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("upgrade")) == "websocket" {
		if err != nil {
			var apiErr *defs.APIError

			if !errors.As(err, &apiErr) {
				apiErr = defs.ErrInternal().Wrap(err)
			}

			context.Set(r, "error", apiErr)
		}
		return
	}

	FormatJSONResp(w, r, resp, err)
}

type route struct {
	path    string
	method  string
	handler appHandlerFunc
}

type Router struct {
	log        *zap.SugaredLogger
	config     *config.Config
	router     http.Handler
	handler    *handler
	middleware *Middleware
	version    *version.Version
}

func NewQRouter(log *zap.SugaredLogger, config *config.Config, version *version.Version) (*Router, error) {

	rt := &Router{
		log:        log,
		config:     config,
		router:     nil,
		handler:    &handler{},
		middleware: NewMiddleware(log, config.HTTP.ProxyForwardedHeader, config.HTTP.LogAllRequests),
		version:    version,
	}

	var err error
	store, err := util.NewFileStore(config.Base.StoreFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default file store")
	}

	core, err := lib.New(&config.Base, store)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init core")
	}

	rt.handler = &handler{
		core:    core,
		cfg:     *config,
		log:     log,
		version: version,
	}
	if err != nil {
		return nil, err
	}
	rt.router = rt.SetHandlers()

	return rt, nil
}

// set all handler
func (r *Router) SetHandlers() http.Handler {

	routes := []route{
		{PathHealthcheckVersion, http.MethodGet, r.handler.HealthCheckVersion},
		{PathHealthCheckConfig, http.MethodGet, r.handler.HealthCheckConfig},
		{PathClientFullRegister, http.MethodPost, r.handler.ClientFullRegister},
		{PathClientsList, http.MethodGet, r.handler.ClientsList},
		{PathAction, http.MethodPut, r.handler.ActionApprove},
		{PathAction, http.MethodDelete, r.handler.ActionReject},
		{PathClientFeed, defs.MethodWebsocket, r.handler.ClientFeed},
	}

	router := mux.NewRouter().PathPrefix(PathPrefix).Subrouter()
	for _, route := range routes {

		middle := r.middleware.notProtectedMiddleware

		if route.method == defs.MethodWebsocket {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler)))
		} else {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler))).Methods(route.method)
		}
	}

	router.Use(r.middleware.loggingMiddleware)

	r.printRoutes(router)

	return r.setupCORS(router)
}

// Start starts the service
func (r *Router) Start() error {
	errChan := make(chan error)
	r.StartHTTPListener(errChan)

	return <-errChan
}

// StartHTTPListener starts the HTTP listener
func (r *Router) StartHTTPListener(errChan chan error) {
	r.log.Infof("CORS policy: %s", strings.Join(r.config.HTTP.CORSAllowOrigins, ","))
	if r.config.HTTP.ProxyForwardedHeader != "" {
		r.log.Info("Use Proxy forwarded-for header: %s", r.config.HTTP.ProxyForwardedHeader)
	}
	r.log.Infof("Starting listener on %v", r.config.HTTP.Addr)

	err := r.handler.AutoApproval()
	if err != nil {
		r.log.Infof("Cannot start server. Error: %s", err.Error())
		os.Exit(1)
	}

	errChan <- http.ListenAndServe(r.config.HTTP.Addr, context.ClearHandler(r.router))
}

func (r *Router) setupCORS(h http.Handler) http.Handler {
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With"}),
		handlers.AllowedOrigins(r.config.HTTP.CORSAllowOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD"}),
		handlers.AllowCredentials(),
	)
	return cors(h)
}

func (r *Router) printRoutes(router *mux.Router) {
	if err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if tpl, err := route.GetPathTemplate(); err == nil {
			if met, err := route.GetMethods(); err == nil {
				for _, m := range met {
					r.log.Debugf("Registered handler %v %v", m, tpl)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

// WriteHTTPError writes the error response as JSON
func WriteHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *defs.APIError

	if !errors.As(err, &apiErr) {
		apiErr = defs.ErrInternal().Wrap(err)
	}
	context.Set(r, "error", apiErr)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(apiErr.Code())
	w.Write(apiErr.JSON())
}

// FormatJSONResp encodes response as JSON and handle errors
func FormatJSONResp(w http.ResponseWriter, r *http.Request, v interface{}, err error) {
	if err != nil {
		WriteHTTPError(w, r, err)
		return
	}

	if v == nil {
		v = &struct {
			Code int
			Msg  string
		}{
			Code: http.StatusOK,
			Msg:  http.StatusText(http.StatusOK),
		}
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		WriteHTTPError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
