module github.com/eclipse-kanto/container-management

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.11
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20200319173657-742aab907b54
	github.com/docker/docker => github.com/moby/moby v1.4.2-0.20200330121334-7f8b4b621b5d
	github.com/docker/libnetwork => github.com/moby/libnetwork v0.8.0-dev.2.0.20200318182600-ef149a924dfd
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc10
	github.com/vishvananda/netlink => github.com/vishvananda/netlink v1.1.0
)

go 1.17

require (
	github.com/containerd/containerd v1.3.0-beta.2.0.20190828155532-0293cbd26c69
	github.com/containerd/typeurl v0.0.0-20180627222232-a93fcdb778cd
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/libnetwork v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.3.5
	github.com/google/uuid v1.1.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc9.0.20200309161605-bbaba4c08118
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c
	google.golang.org/grpc v1.28.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/containerd/cgroups v1.0.2 // indirect
	github.com/containerd/continuity v0.0.0-20190426062206-aaeac12a7ffc // indirect
	github.com/containerd/fifo v0.0.0-20190226154929-a9fb20d87448 // indirect
	github.com/containerd/ttrpc v0.0.0-20190828154514-0e0f228740de // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/docker/distribution v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libkv v0.2.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.3 // indirect
	github.com/hashicorp/consul/api v1.4.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.3 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/memberlist v0.2.0 // indirect
	github.com/hashicorp/serf v0.9.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/ishidawataru/sctp v0.0.0-20191218070446-00ab2ac2db07 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-shellwords v1.0.10 // indirect
	github.com/miekg/dns v1.1.27 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/moby/ipvs v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/opencontainers/selinux v1.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/syndtr/gocapability v0.0.0-20170704070218-db04d3cc01c8 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200122045848-3419fae592fc // indirect
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	go.opencensus.io v0.22.0 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20200313141609-30c55424f95d // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
