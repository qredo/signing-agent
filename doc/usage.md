[TOC]

# Introduction

The Automated Approver is an agent that can operate as a standalone service (exposing a RESTful API to 3rd party applications), or as a library integrated into an application. We recommend deploying the service on premise, on the customer’s infrastructure. We also recommend that every automated approver instance be used to manage a unique agent ID, and that multiple instances be deployed (preferably on different cloud infrastructures) in order to meet a multiple signer threshold such as 2 or 3 out of 5. The automated approver uses a dedicated subset of the Qredo Server APIs, called the Partner API, to perform its functions. It can also be used to create a programmable approver service that listens to incoming approval requests over Websockets, and is then able to perform automated custody.

In a nutshell, it works just like the phone app but without the human element. The server acts just like a human approver, which means that it *approves* all transaction types that move assets to and from a Qredo wallet:

- **transfer** - (also transfer out) a transaction between wallets that both reside on the Qredo Network: a L2 to L2 transaction.
- **withdrawal** - a transaction where assets in a Qredo wallet move to a wallet outside the Qredo Network (BTC, ETH, etc.): a L2 to L1 transaction).
- **atomic swap** - a transfer out transaction where you offer a certain amount of an asset in exchange for a (transfer in) certain amount of another asset. (e.g. exchange 100000000 ETH qweis for 735601 satoshis). Both parties that participate have a transfer out transaction that undergoes custody with their approvers. This transaction type is discussed in more detail in the [Atomic swaps](https://developers.qredo.com/partner-api/how-tos/atomic-swap/) section of the Qredo documentation portal.

# Using Automated Approver as a Service

As mentioned above, the Automated Approver is a standalone component of the Qredo ecosystem. Everyone who intends to run an Automated Approver must first register it on the Qredo network. Below is a step-by-step explanation of the registration process, which involves the *PartnerAPP* (e.g. Banco Hipotecario), the *automated-approver-service* (e.g. Automated Approver running on Banco Hipotecario’s infrastructure), and *QredoBE* (e.g. our Qredo back-end).

```mermaid
sequenceDiagram
autonumber
  PartnerAPP->>automated-approver-service:POST /register {"name":"...","APIKey":"...","Base64PrivateKey":"..."}
  automated-approver-service->>automated-approver-service: Generate BLS and EC Keys
  automated-approver-service->>QredoBE: Register agent with Qredo Chain
  QredoBE->>automated-approver-service: {agentId, feedURL, other data} 
  automated-approver-service->>PartnerAPP: {agentId, feedURL}
```

1. The *PartnerApp* triggers the registration process by providing its client name, parther APIKey and Base64PrivateKey  to the *automated-approver-service*.
2. *automated-approver-service* generates BLS and EC keys. 
3. The *automated-approver-service* can now register itself to the partner API on the *QredoBE*, by sending the `client name`, `BLS`, and `EC` public keys. The *QredoBE* is returning ClientID, CLientSecret that will be responsible for authentication.
4. The `agentId` and a `feedURL` is returned by the *QredoBE* to the *automated-approver-service*. This feed is used by the *automated-approver-service* to keep a communication channel open with the *QredoBE*.
5. The `agentId` and a `feedURL` is also passed along to the *PartnerApp* so that the latter can monitor for new actions that need to be approved in case the service is not configured for auto-approval.

All the data above is currently stored on premises in a file by the automated-approver-service, and since some of it (ClientSecret, EC & BLS private keys) is quite sensitive it needs to be running in a secure environment.

**Note:** an always up to date API documentation can be accessed within the (private) [Gitlab repo](https://gitlab.qredo.com/custody-engine/automated-approver/-/blob/master/doc/swagger/swagger.yaml).

## API

### POST /api/v1/register

Request:

```json
{
  "name": "string", 
  "APIKey": "string",
  "Base64PrivateKey": "string"
}
```

Response (clientRegisterResponse):

```json
{
  "agentId": "string",
  "feedUrl": "string"
}
```



# Using Automated Approver as a Library

There are times when the Automated Approver benefits from being tightly coupled with an application or a service. In this case, it can be imported as a Go package directly into that application.

 An example of the automated approver onboarding process using the automated approver library in a Go app would look like:

```mermaid
sequenceDiagram
  autonumber
  PartnerAPP->>PartnerAPP:ClientRegister('client_name')

  rect rgb(200, 150, 255)
  note right of PartnerAPP: inside the automated approver lib
  PartnerAPP->>PartnerAPP: Generate BLS and EC Keys
  end
  PartnerAPP->>QredoBE: ClientInit(reqDataInit, RefID, APIKey, Base64PrivateKey) reqDataInit: {name, BLS & EC PubKeys}
  QredoBE->>QredoBE: Create New MFA ID, New IDDoc
  QredoBE->>PartnerAPP:{ClientSecret, ClientID, unsigned IDDoc, AccountCode}
  PartnerAPP->>PartnerAPP: ClientRegisterFinish(ClientSercert, ID, unsigned IDDoc, AccountCode)
  rect rgb(200, 150, 255)
  note right of PartnerAPP: inside the automated approver lib
  PartnerAPP->>PartnerAPP: Store ClientSercert, ID, AccountCode
  PartnerAPP->>PartnerAPP: Sign IDDoc
  PartnerAPP->>QredoBE: POST /p/coreclient/finish { IDdoc Signature }
  end
  QredoBE->>QredoBE: Save IDDoc in qredochain
  QredoBE->>PartnerAPP: {feedURL}
```

# Approving a transaction

Prerequisites: 

- a automated approver service instance has been installed and configured
- a automated-approver has been created with id `agentID`

Steps:

1. A websocket connection to the *Qredo BE* is opened for said `agentID`
2. *PartnerAPP* is constantly monitoring for new actions to be handled
3. A new transfer is initiated
4. The *Qredo BE* returns the transaction id: `tx_id`
5. Shortly after, a new action is received through the websocket with `action_id` equal to the `tx_id` for the transfer.
6. Initiate new action
7. The *PartnerAPP* requests from the *Qredo BE* details for the action
8. *Qredo BE* returns action details incl. the payload (list of messages)
9. Sign payload (for the new action)
10. The *PartnerAPP* decides to approve the transactions, thus sending the payload to the automated-approver with a `PUT` request. (`DELETE` is for reject)

After that sequence, the transaction should be complete.

### Using the library to approve a transaction

```mermaid
sequenceDiagram
  autonumber

  par
  PartnerAPP->>QredoBE: WEBSOCKET /coreclient/feed
  PartnerAPP->>PartnerAPP: monitor for new actions
  end
  PartnerAPP->>QredoBE:POST /company/{company_id}/transfer
  QredoBE->>PartnerAPP: tx_id
  QredoBE->>PartnerAPP: {via websocket } action_id(==tx_id), type, status
  PartnerAPP->>PartnerAPP: ActionApprove(actionID)
  rect rgb(200, 150, 255)
  note right of PartnerAPP: inside the automated approver lib
  PartnerAPP->>QredoBE: GET /coreclient/action/{action_id}
  QredoBE->>PartnerAPP: action details incl. list of messages
  PartnerAPP->>PartnerAPP: signing those messages
  PartnerAPP->>QredoBE: PUT /coreclient/action/{action_id} send signed messages to BE

  end
```



### Data Models

```Go
ClientRegisterFinishRequest {
    accountCode    string
    clientID   string
    clientSecret   string
    id  string
    idDoc  string
}
```

```Go
ClientRegisterRequest {
    name             string
    apikey           string
    base64privatekey string
}
```

```Go
SignRequest {
    message_hash_hex    string
}
```

```Go
VerifyRequest {
    message_hash_hex    string
    signature_hex   string
    signer_id   string
}
```

```Go
clientRegisterFinishResponse {
    feed_url    string
}
```

```Go
clientRegisterResponse {
    bls_public_key  string
    ec_public_key   string
    ref_id  string
}
```

```Go
signResponse {
    signature_hex   string
    signer_id   string
}
```

