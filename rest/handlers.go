package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/defs"
	"gitlab.qredo.com/qredo-server/core-client/lib"
)

type handlers struct {
	core lib.CoreClient
}

// ClientRegister
//
// swagger:route POST /client browser clientRegister clientRegisterInit
//
// Initiate client registration procedure
//
// Responses:
//      200: clientRegisterResponse
func (h *handlers) ClientRegister(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	return h.core.ClientRegister(req)
}

// ClientRegisterFinish
//
// swagger:route POST /client{ref}  browser clientRegister clientRegisterFinish
//
// Finish client registration procedure
//
// Responses:
//      200: clientRegisterFinishResponse
func (h *handlers) ClientRegisterFinish(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	ref := mux.Vars(r)["ref"]
	req := &api.ClientRegisterFinishRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	h.core.ClientRegisterFinish(req)
}
