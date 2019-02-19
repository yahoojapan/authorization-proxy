# Authorization Proxy
[![License: Apache](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://opensource.org/licenses/Apache-2.0) [![release](https://img.shields.io/github/release/yahoojapan/authorization-proxy.svg?style=flat-square)](https://github.com/yahoojapan/authorization-proxy/releases/latest) [![CircleCI](https://circleci.com/gh/yahoojapan/authorization-proxy.svg)](https://circleci.com/gh/yahoojapan/authorization-proxy) [![codecov](https://codecov.io/gh/yahoojapan/authorization-proxy/branch/master/graph/badge.svg?token=2CzooNJtUu&style=flat-square)](https://codecov.io/gh/yahoojapan/authorization-proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/yahoojapan/authorization-proxy)](https://goreportcard.com/report/github.com/yahoojapan/authorization-proxy) [![GolangCI](https://golangci.com/badges/github.com/yahoojapan/authorization-proxy.svg?style=flat-square)](https://golangci.com/r/github.com/yahoojapan/authorization-proxy) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/32397d339f6c450a82af72c8a0c15e5f)](https://www.codacy.com/app/i.can.feel.gravity/authorization-proxy?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=yahoojapan/authorization-proxy&amp;utm_campaign=Badge_Grade) [![GoDoc](http://godoc.org/github.com/yahoojapan/authorization-proxy?status.svg)](http://godoc.org/github.com/yahoojapan/authorization-proxy) [![DepShield Badge](https://depshield.sonatype.org/badges/yahoojapan/authorization-proxy/depshield.svg)](https://depshield.github.io)


![logo](./images/logo.png)

---

## What is Authorization Proxy

Authorization Proxy is an implementation of [Kubernetes sidecar container](https://kubernetes.io/blog/2015/06/the-distributed-system-toolkit-patterns/) to provide a common interface for the user to do authentication and authorization check for specific URL resources. It caches the policies from [Athenz](https://github.com/yahoo/athenz), and provide a reverse proxy interface for the user to authenticate the role token written on the request header, to allow or reject user's specific URL request.

Requires go 1.9 or later.

## Use case

### Authorization and Authorization request

Authorization Proxy acts as a reverse proxy sitting in front of the user application. When the user request for specific URL resource of the user application, the request comes to authorization proxy first.

#### Policy updator

To authenticate the request, the authorization proxy should know which user can take an action to which resource, therefore the policy updator is introduced.

![Policy updator](./doc/assets/auth_proxy_policy_updator.png)

The policy updator periodically updates the Athenz config and Policy data from Athenz Server and validate and decode the policy data. The decoded result will store in the memory cache inside the policy updator.

#### Authorization success

![Auth success](./doc/assets/auth_proxy_use_case_auth_success.png)

The authorization proxy will verify and decode the role token written on the request header and check if the user can take an action to a specific resource. If the user is allowed to take an action the resource, the request will proxy to the user application.

#### Authorization failed

![Auth fail](./doc/assets/auth_proxy_use_case_auth_failed.png)

The authorization proxy will return unauthorized to the user whenever if the role token is invalid, or the role written on the role token has no privilege to take the action to the resource that user is requesting.

---

### Mapping rules

The mapping rules describe the elements using in the authorization proxy. The user can configure which Athenz domain's policies cached in the policy updator, and decide if the user is authorized to take the action to the resource.

The mapping rules were described below.

|          | Description                         | Map to (Athenz)  | Example   |   |
|----------|-------------------------------------|------------------|-----------|---|
| Role     | Role name written on the role token | Role             | admin     |   |
| Action   | HTTP/HTTPS request action           | Action           | POST      |   |
| Resource | HTTP/HTTPS request resource         | Resource         | /user/add |   |

All the HTTP/HTTPS methods and URI path will be automatically converted to lower case. 

## Configuration

- [config.go](./config/config.go)

---

## License

## TODO

## Authors

- [kpango](https://github.com/kpango)
- [kevindiu](https://github.com/kevindiu)
- [TakuyaMatsu](https://github.com/TakuyaMatsu)
- [tatyano](https://github.com/tatyano)
