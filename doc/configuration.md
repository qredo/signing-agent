# YAML config file template

```yaml
base:
  qredo_api_domain: play-api.qredo.network
  qredo_api_base_path: /api/v1/p
  http_scheme: https
  pin: 0
auto_approval:
  enabled: false
  retry_interval_max_sec: 300
  retry_interval_sec: 5
websocket:
  ws_scheme: wss
  reconnect_timeout_sec: 300
  reconnect_interval_sec: 5
  ping_period_sec: 5
	pong_wait_sec: 10
	write_wait_sec: 10
  read_buffer_size: 512
  write_buffer_size: 1024
http:
  addr: 0.0.0.0:8007
  cors_allow_origins:
    - '*'
  log_all_requests: false
logging:
  format: text
  level: debug
load_balancing:
  enable: false
  on_lock_error_timeout_ms: 300
  action_id_expiration_sec: 6
  redis:
    host: localhost
    port: 6379
    password: ""
    db: 0
store: 
  type: file 
  file: /volume/ccstore.db
  oci:
    compartment: ocid1.tenancy.oc1...
    vault: ocid1.vault.oc1...
    secret_encryption_key: ocid1.key.oc1...
    config_secret: automated_approver_config
  aws:
    region: aws-region-...
    config_secret: secrets_manager_secret...  
```

## Base

- **qredo_api_domain:** the domain of the api you want to use
- **qredo_api_base_path:** base path for the api urls, ex. scheme://domain/base_path
- **http_scheme:** the scheme to use for the api connection, ex. http or https
- **pin:** the pin number to use to provide a zero knowledge proof token for communication with the partner api

## Auto approval
- **enabled:** activate the automatic approval of every transaction that is received
- **retry_interval_max_sec:** the approve action maximum interval in seconds
- **retry_interval_sec:** the approve action retry interval in seconds

## Websocket
- **ws_scheme:** the scheme to use for the web socket feed connection, ex. ws or wss
- **reconnect_timeout_sec:** the reconnect timeout in seconds
- **reconnect_interval_sec:** the reconnect interval in seconds
- **ping_period_sec:** the ping period for the ping handler in seconds
- **pong_wait_sec:** the pong wait for the pong handler in seconds
- **write_wait_sec:** the write wait in seconds
- **read_buffer_size:** the websocket upgrader read buffer size in bytes
- **write_buffer_size:** the websocket upgrader write buffer size in bytes

## HTTP

- **addr:** the bind address and port the build in api endpoints
- **cors_allow_origins:** the value the the Access-Control-Allow-Origin of the responses of the build in api
- **log_all_requests:** log all incoming requests to the build in api

## Logging

- **format:** the format of the logging, ex. text or json
- **level:** the level of logging, ex. debug, info, warn, error, default is debug

## Load balancing

- **enable:** enables the load-balancing logic
- **on_lock_error_timeout_ms:** on lock timeout in milliseconds
- **action_id_expiration_sec:** expiration of action_id variable in Redis in seconds
- **redis:**
  - **host:** Redis host
  - **port:** Redis port
  - **password:** Redis password
  - **db:** Redis database name

## Store

- **type:** the type of store to use to store the private key information for the signing agent, ex. file, oci, aws
- **file:** the path to the storage file when file store is used
- **oci:** the oracle cloud configuration to store the private keys in an oracle vault
  - **compartment:** the oidc where the vault and encryption key reside
  - **vault:** the oidc of the vault where the secret will be stored
  - **secret_encryption_key:** the encryption key used for both the secret and the data inside the secret
  - **config_secret:** the name of secret that will be used to store the data
- **aws:** the amazon cloud configuration to store the private keys in amazon secrets manager
  - **region:** the AWS region where the secret is stored
  - **config_secret:** the name of the AWS Secrets Manager secret containing the encrypted data