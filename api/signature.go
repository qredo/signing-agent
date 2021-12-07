package api

import "errors"

type SignRequest struct {
	MessageHashHex string `json:"message_hash_hex"`
}

func (r *SignRequest) Validate() error {
	if r.MessageHashHex == "" {
		return errors.New("id")
	}

	return nil
}

type SignResponse struct {
	SignatureHex string `json:"signature_hex"`
	SignerID     string `json:"signer_id"`
}

type VerifyRequest struct {
	MessageHashHex string `json:"message_hash_hex"`
	SignatureHex   string `json:"signature_hex"`
	SignerID       string `json:"signer_id"`
}

func (r *VerifyRequest) Validate() error {
	if r.MessageHashHex == "" {
		return errors.New("id")
	}
	if r.SignatureHex == "" {
		return errors.New("id")
	}
	if r.SignerID == "" {
		return errors.New("id")
	}
	return nil
}
