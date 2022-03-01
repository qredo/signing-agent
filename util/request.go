package util

import (
	"io"
	"net/http"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/qredo-core/common"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"
)

func DecodeRequest(req api.Validator, hr *http.Request) error {
	switch hr.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if err := common.DecodeJSON(hr, req); err != nil {
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
