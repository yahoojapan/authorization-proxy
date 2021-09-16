module github.com/yahoojapan/authorization-proxy/v4

go 1.14

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210915214749-c084706c2272
	golang.org/x/net => golang.org/x/net v0.0.0-20210916014120-12bc252f5db8
	golang.org/x/text => golang.org/x/text v0.3.7
	google.golang.org/grpc => google.golang.org/grpc v1.40.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.27.1
)

require (
	github.com/kpango/glg v1.6.4
	github.com/mwitkow/grpc-proxy v0.0.0-20181017164139-0f1106ef9c76
	github.com/pkg/errors v0.9.1
	github.com/yahoojapan/athenz-authorizer/v5 v5.2.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)
