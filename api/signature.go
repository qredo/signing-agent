package api

type SignRequest struct {
	MessageHashHex string `json:"message_hash_hex" validate:"required"`
}

// swagger:model signResponse
type SignResponse struct {
	SignatureHex string `json:"signature_hex"`
	SignerID     string `json:"signer_id"`
}

type VerifyRequest struct {
	MessageHashHex string `json:"message_hash_hex" validate:"required"`
	SignatureHex   string `json:"signature_hex" validate:"required"`
	SignerID       string `json:"signer_id" validate:"required"`
}
