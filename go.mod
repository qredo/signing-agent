module gitlab.qredo.com/qredo-server/core-client

go 1.18

replace github.com/btcsuite/btcd => github.com/qredo/btcd v0.21.2

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/go-playground/validator/v10 v10.10.1
	github.com/google/uuid v1.3.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/qredo/assets v1.8.1
	go.uber.org/zap v1.21.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/sys v0.0.0-20210816183151-1e6c022a8912 // indirect
	golang.org/x/text v0.3.7 // indirect
)
