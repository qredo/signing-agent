// Package classification Qredo Signing Agent V2
//
// The Qredo Signing Agent service interacts with [Partner API](https://developers.qredo.com/partner-api/api/swagger/) to register a Signing Agent to automate approvals according to your custody policy. <br/>
// Authentication and encryption are required; set up your [API key and secret in the Qredo Web App](https://developers.qredo.com/signing-agent/v2-signing-agent/get-started/). <br/>
//
// Version: 1.0.0
// Contact: Qredo API Services<support@qredo.com> https://www.qredo.com
// Schemes: http, https
// Host: localhost:8007
// BasePath: /api/v1
// License: APACHE 2.0 https://www.apache.org/licenses/LICENSE-2.0
// swagger:meta
package api

// swagger:parameters RegisterAgent
type DOCClientRegisterRequest struct {
	// in:body
	Body ClientRegisterRequest
}

// swagger:model GetClientResponse
type DOCGetClientResponse struct {
	// in:body
	Body GetClientResponse
}

// swagger:model ClientFeedResponse
type DOCClientFeedResponse struct {
	// The ID of the transaction
	// example: 2IXwq4klvWbnPf1YaAc1XD85jJX
	ID string `json:"id"`

	// The ID of the agent
	// example: 98cTMMSPrDdcDDVU8idhuJGK2U1P4vmQcsp8wnED8pPR
	CoreClientID string `json:"coreClientID"`

	// The type of the transaction
	// example: ApproveWithdraw
	Type string `json:"type"`

	// The status of the transaction
	// example: pending
	Status string `json:"status"`

	// The time that the transaction was started, utc unix time
	// example: 1670341423
	Timestamp int64 `json:"timestamp"`

	// The time that the transaction will expire, utc unix time
	// example: 1676184187
	ExpireTime int64 `json:"expireTime"`
}

// swagger:model GenericResponse
type DOCGenericResponse struct {
	// The result code of the request
	// example: 200
	Code int

	// The result message of the request
	// example: OK
	Msg string
}
