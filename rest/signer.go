package rest

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"

	"github.com/pkg/errors"
)

func signRequest(req *request) error {
	h := sha256.New()
	h.Write([]byte(req.timestamp))
	h.Write([]byte(req.uri))
	h.Write(req.body)
	dgst := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(nil, req.rsaKey, crypto.SHA256, dgst)
	if err != nil {
		return errors.Wrap(err, "cannot sign request")
	}

	req.signature = base64.RawURLEncoding.EncodeToString(signature)
	return nil
}
