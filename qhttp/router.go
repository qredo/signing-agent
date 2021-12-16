package qhttp

import (
	"net/http"
	"strings"

	"gitlab.qredo.com/qredo-server/core-client/defs"

	qhandlers "gitlab.qredo.com/qredo-server/core-client/handlers"

	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"github.com/gorilla/context"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/qredo-core/qdefs"
	"gitlab.qredo.com/qredo-server/qredo-core/qnet"
	"go.uber.org/zap"
)

var (
	pathPrefix    = "/api/v1"
	authHeader    = "x-token"
	mfaAuthHeader = "x-zkp-token"
)

const (
	scopeNone                 = ""
	scopeRegister             = "register"
	scopeLogin                = "login"
	scopeLogged               = "logged"
	scopeMobile               = "mobile"
	scopeForgottenCredentials = "forgotten-pwd"
	scopeResetCredentials     = "reset-pwd"
	scopeResetMobile          = "reset-mobile"
	scopeSwitchAccount        = "switch-account"
	scopeWebsocket            = "websocket"
)

type appHandlerFunc func(ctx *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error)

func (a appHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := &defs.RequestContext{}

	resp, err := a(ctx, w, r)

	if strings.ToLower(r.Header.Get("connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("upgrade")) == "websocket" {
		if err != nil {
			var qErr *qerr.QErr
			if e, ok := err.(*qerr.QErr); ok {
				qErr = e
			} else {
				qErr = qerr.Internal().Wrap(err).WithMessage("unknown error")
			}
			context.Set(r, "error", qErr)
		}
		return
	}

	qnet.FormatJSONResp(w, r, resp, err)
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
	handler    *qhandlers.Handler
	middleware *Middleware
}

func NewQRouter(log *zap.SugaredLogger, config *config.Config) (*Router, error) {

	rt := &Router{
		log:        log,
		config:     config,
		middleware: NewMiddleware(log, config.HTTP.ProxyForwardedHeader, config.HTTP.LogAllRequests),
	}

	var err error
	rt.handler, err = qhandlers.New(rt.log, config)
	if err != nil {
		return nil, err
	}
	rt.router, _ = rt.setHandlers()

	return rt, nil
}

// set all handler
func (r *Router) setHandlers() (http.Handler, error) {

	routes := []route{

		{"/client", http.MethodPost, r.handler.ClientRegister},
		{"/client/{ref}", http.MethodPut, r.handler.ClientRegisterFinish},
		{"/client", http.MethodGet, r.handler.ClientsList},
		{"/client/{client_id}/action/{action_id}", http.MethodPut, r.handler.ActionApprove},
		{"/client/{client_id}/action/{action_id}", http.MethodDelete, r.handler.ActionReject},

		{"/client/{client_id}/sign", http.MethodPost, r.handler.Sign},
		{"/verify", http.MethodPost, r.handler.Verify},
	}

	router := mux.NewRouter().PathPrefix(pathPrefix).Subrouter()
	for _, route := range routes {

		middle := r.middleware.notProtectedMiddleware

		if route.method == qdefs.MethodWebsocket {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler)))
		} else {
			router.Handle(route.path, r.middleware.sessionMiddleware(middle(route.handler))).Methods(route.method)
		}
	}

	router.Use(r.middleware.loggingMiddleware)

	r.printRoutes(router)

	return r.setupCORS(router), nil
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
	r.log.Infof("Starting listener on %v for API url %v", r.config.HTTP.Addr, r.config.BaseURL)

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
					r.log.Infof("Registered handler %v %v", m, tpl)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
