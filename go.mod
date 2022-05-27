module github.com/yahoojapan/authorization-proxy/v4

go 1.16

require (
	github.com/kpango/glg v1.6.10
	github.com/mwitkow/grpc-proxy v0.0.0-20220126150247-db34e7bfee32
	github.com/pkg/errors v0.9.1
	github.com/yahoojapan/athenz-authorizer/v5 v5.3.2
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
)

replace golang.org/x/text => golang.org/x/text v0.3.7
