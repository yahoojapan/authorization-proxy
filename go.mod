module github.com/yahoojapan/authorization-proxy/v4

go 1.18

require (
	github.com/kpango/glg v1.6.12
	github.com/mwitkow/grpc-proxy v0.0.0-20220126150247-db34e7bfee32
	github.com/pkg/errors v0.9.1
	github.com/yahoojapan/athenz-authorizer/v5 v5.3.2
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/AthenZ/athenz v1.11.2 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/goccy/go-json v0.9.10 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/klauspost/cpuid/v2 v2.0.13 // indirect
	github.com/kpango/fastime v1.1.4 // indirect
	github.com/kpango/gache v1.2.8 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.2.25 // indirect
	github.com/lestrrat-go/option v1.0.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1 // indirect
)

replace golang.org/x/text => golang.org/x/text v0.3.7

replace github.com/yahoojapan/athenz-authorizer/v5 v5.3.2 => github.com/yahoojapan/athenz-authorizer/v5 v5.4.1-0.20221107001904-4d4db0a61e1f
