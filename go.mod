module github.com/hpcng/warewulf

go 1.15

require (
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/storage v1.23.9 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/etcd-io/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/umoci v0.4.6
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/vbauerster/mpb v3.4.0+incompatible // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/etcd-io/bbolt => go.etcd.io/bbolt v1.3.5
