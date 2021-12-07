module gitlab.qredo.com/qredo-server/core-client

go 1.16

replace (
	github.com/btcsuite/btcd => github.com/qredo/btcd v0.21.2
	github.com/qredo/assets => /Users/osn/go/src/github.com/qredo/assets
	gitlab.qredo.com/qredo-server/qredo-core => /Users/osn/qredo/qredo-server/qredo-core
	gitlab.qredo.com/qredo-server/qredo-libs => /Users/osn/qredo/qredo-server/qredo-libs
	gitlab.qredo.com/qredo-server/qredo-libs/aws/dynamodb => /Users/osn/qredo/qredo-server/qredo-libs/aws/dynamodb
	gitlab.qredo.com/qredo-server/qredo-libs/keystore => /Users/osn/qredo/qredo-server/qredo-libs/keystore
	gitlab.qredo.com/qredo-server/qredo-libs/kvstore => /Users/osn/qredo/qredo-server/qredo-libs/kvstore
)

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/corpix/uarand v0.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.2.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/nats-io/nats.go v1.11.0 // indirect
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/qredo/assets v1.6.18
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/urfave/cli/v2 v2.3.0
	gitlab.qredo.com/qredo-server/qredo-core v0.0.0-00010101000000-000000000000
	gitlab.qredo.com/qredo-server/qredo-libs/keystore v0.0.0-00010101000000-000000000000 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.uber.org/zap v1.17.0
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.33.2 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
