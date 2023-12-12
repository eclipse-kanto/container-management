module github.com/eclipse-kanto/container-management

go 1.19

replace github.com/docker/docker => github.com/moby/moby v23.0.9+incompatible

require (
	github.com/caarlos0/env/v6 v6.10.1
	github.com/containerd/cgroups v1.1.0
	github.com/containerd/containerd v1.6.28
	github.com/containerd/imgcrypt v1.1.9
	github.com/containerd/typeurl v1.0.2
	github.com/containers/ocicrypt v1.1.9
	github.com/docker/docker v23.0.9+incompatible
	github.com/eclipse-kanto/kanto/integration/util v0.0.0-20230103144956-911e45a2bf55
	github.com/eclipse-kanto/update-manager v0.0.0-20230628072101-b91f0c30e00f
	github.com/eclipse/ditto-clients-golang v0.0.0-20220225085802-cf3b306280d3
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.1
	github.com/notaryproject/notation-core-go v1.0.1
	github.com/notaryproject/notation-go v1.0.1
	github.com/opencontainers/go-digest v1.0.0
<<<<<<< HEAD
	github.com/opencontainers/image-spec v1.1.0-rc6
	github.com/opencontainers/runc v1.1.12
	github.com/opencontainers/runtime-spec v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil/v3 v3.22.7
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.2.1
=======
	github.com/opencontainers/image-spec v1.1.0-rc5
	github.com/opencontainers/runc v1.1.5
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil/v3 v3.22.7
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
>>>>>>> e361ff5 (Implement signed images verification (#215))
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	github.com/vishvananda/netlink v1.2.1-beta.2
<<<<<<< HEAD
	golang.org/x/crypto v0.18.0
	golang.org/x/net v0.18.0
	golang.org/x/sync v0.3.0
	golang.org/x/sys v0.16.0
=======
	golang.org/x/crypto v0.14.0
	golang.org/x/net v0.17.0
	golang.org/x/sync v0.4.0
	golang.org/x/sys v0.13.0
>>>>>>> e361ff5 (Implement signed images verification (#215))
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.1
<<<<<<< HEAD
	gotest.tools/v3 v3.5.0
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
=======
	gotest.tools/v3 v3.4.0
	oras.land/oras-go/v2 v2.3.1
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.6 // indirect
>>>>>>> e361ff5 (Implement signed images verification (#215))
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/containerd/continuity v0.4.2 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/libkv v0.2.1 // indirect
<<<<<<< HEAD
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
=======
	github.com/fxamacker/cbor/v2 v2.5.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/go-ldap/ldap/v3 v3.4.6 // indirect
>>>>>>> e361ff5 (Implement signed images verification (#215))
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/memberlist v0.3.1 // indirect
	github.com/hashicorp/serf v0.9.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ishidawataru/sctp v0.0.0-20210707070123-9a39160e9062 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/miekg/dns v1.1.46 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/moby/ipvs v1.0.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rootless-containers/rootlesskit v1.1.1 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20201008174630-78d3cae3a980 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20200128120323-432b2356ecb1 // indirect
<<<<<<< HEAD
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/term v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.10.0 // indirect
=======
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
>>>>>>> e361ff5 (Implement signed images verification (#215))
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
)
