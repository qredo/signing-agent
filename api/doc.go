package api

// swagger:parameters clientRegisterInit
type DOCClientRegisterRequest struct {
	// in:body
	Body ClientRegisterRequest
}

// swagger:parameters clientRegisterFinish
type DOCRegisterFinishRequest struct {
	// in:body
	Body ClientRegisterFinishRequest
}

// swagger:parameters payloadSign
type DOCSignRequest struct {
	// in:body
	Body SignRequest
}

// swagger:parameters signatureVerify
type DOCVerifyRequest struct {
	// in:body
	Body VerifyRequest
}
