module github.com/yahoojapan/authorization-proxy/v4

go 1.14

require (
	github.com/kpango/glg v1.6.4
	github.com/pkg/errors v0.9.1
	github.com/yahoojapan/athenz-authorizer/v5 v5.2.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/yahoojapan/athenz-authorizer/v5 v5.2.0 => github.com/yahoojapan/athenz-authorizer/v5 v5.2.1-0.20210802054819-60faa177ce41
