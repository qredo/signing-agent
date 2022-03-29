package util

import (
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"gitlab.qredo.com/qredo-server/qredo-core/common"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"
)

func DecodeRequest(req interface{}, hr *http.Request) error {
	switch hr.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if err := common.DecodeJSON(hr, req); err != nil {
			if err != io.EOF {
				return qerr.New(qerr.ErrBadRequest).WithMessage("invalid json").Wrap(err)
			}
		}
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return qerr.New(qerr.ErrBadRequest).WithDetails("field", err.Error())
	}

	return nil
}
