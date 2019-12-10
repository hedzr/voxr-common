module github.com/hedzr/voxr-common

go 1.12

// replace github.com/hedzr/cmdr v0.0.0 => ../cmdr

// replace github.com/hedzr/cmdr v0.2.25 => ../cmdr

// replace github.com/hedzr/logex v0.0.0 => ../logex

// exclude github.com/coreos/etcd v3.3.10+incompatible // indirect

require (
	github.com/bgentry/speakeasy v0.1.0
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/couchbase/go-couchbase v0.0.0-20190916184909-f83e63d76bc4
	github.com/couchbase/gomemcached v0.0.0-20191004160342-7b5da2ec40b2 // indirect
	github.com/couchbase/goutils v0.0.0-20190315194238-f9d42b11473b // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis v6.15.5+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gocql/gocql v0.0.0-20190927095247-bd5f930c6137
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.3 // indirect
	github.com/hashicorp/consul/api v1.2.0
	github.com/hedzr/cmdr v1.6.9
	github.com/influxdata/influxdb v1.7.8
	github.com/jinzhu/gorm v1.9.11
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/magiconair/properties v1.8.1
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pbnjay/memory v0.0.0-20190104145345-974d429e7ae4
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.etcd.io/etcd v3.3.15+incompatible
	go.uber.org/multierr v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823
	google.golang.org/grpc v1.24.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/yaml.v2 v2.2.4
	sigs.k8s.io/yaml v1.1.0 // indirect
)
