module github.com/weblazy/metcd

go 1.14

replace (
	go.etcd.io/etcd/client/pkg/v3 => ./pkg
	go.etcd.io/etcd/server/v3 => ./server
)

require (
	go.etcd.io/etcd/client/pkg/v3 v3.5.0-alpha.0
	go.etcd.io/etcd/raft/v3 v3.5.0-alpha.0
	go.etcd.io/etcd/server/v3 v3.5.0-alpha.0
	go.uber.org/zap v1.16.1-0.20210329175301-c23abee72d19
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
