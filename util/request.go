package util

import (
	"io"
	"net/http"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"
	"gitlab.qredo.com/qredo-server/qredo-core/qnet"
)

func DecodeRequest(req api.Validator, hr *http.Request) error {
	switch hr.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if err := qnet.DecodeJSON(hr, req); err != nil {
			if err != io.EOF {
				return qerr.New(qerr.ErrBadRequest).WithMessage("invalid json").Wrap(err)
			}
		}
	}

	if err := req.Validate(); err != nil {
		return qerr.New(qerr.ErrBadRequest).WithDetails("field", err.Error())
	}

	return nil
}
