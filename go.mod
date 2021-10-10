module github.com/google/gnostic-grpc

go 1.17

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.2
	github.com/google/gnostic v0.5.5
	github.com/google/go-cmp v0.5.6
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jhump/protoreflect v1.10.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	google.golang.org/genproto v0.0.0-20201019141844-1ed22bb0c154
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

// until google/gnostic release new version
replace github.com/google/gnostic v0.5.5 => github.com/google/gnostic v0.5.6-0.20210930170106-c7a5e4fe8b37
