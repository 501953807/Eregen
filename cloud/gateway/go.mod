module eregen.dev/gateway

go 1.25

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/influxdata/influxdb-client-go/v2 v2.13.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/nats-io/nats.go v1.34.0
	github.com/redis/go-redis/v9 v9.5.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	eregen.dev/shared/validation v0.0.0
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oapi-codegen/runtime v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

replace eregen.dev/shared/validation => ../../shared/validation
replace eregen.dev/shared/crypto => ../../shared/crypto
replace eregen.dev/shared/sanitize => ../../shared/sanitize
replace eregen.dev/shared/ratelimit => ../../shared/ratelimit
