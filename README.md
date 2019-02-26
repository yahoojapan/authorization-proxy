# Authorization Proxy

![logo](./images/logo.png)

---

## What is Authorization Proxy

Authorization Proxy is an implementation of [Kubernetes sidecar container](https://kubernetes.io/blog/2015/06/the-distributed-system-toolkit-patterns/) to provide a common interface for the user to do authentication and authorization check for specific URL resources. It caches the policies from [Athenz](https://github.com/yahoo/athenz), and provide a reverse proxy interface for the user to authenticate the role token written on the request header, to allow or reject user's specific URL request.

Requires go 1.9 or later.

## Use case

### Authenicate and Authorizate request

Authorization Proxy acts as a reverse proxy sitting in front of the user application. When the user request for specific URL resource of the user application, the request comes to authorization proxy first.

#### Policy updator

To authenticate the request, the authorization proxy should know which user can access which resource, therefore the policy updator is introduced.

![Policy updator](./doc/assets/auth_proxy_policy_updator.png)

The policy updator periodically update the Athenz Config and Policy data from Athenz Server and validate and decode the policy data. The decoded result will store in the memory cache inside the policy updator.

#### Authorizate success

![Auth success](./doc/assets/auth_proxy_use_case_auth_success.png)

The authorization proxy will verify and decode the role token written on the request header and check if the user can access a specific resource. If the user is allowed to access the resource, the request will proxy to the user application and return to the user.

#### Authorizate failed

![Auth fail](./doc/assets/auth_proxy_use_case_auth_failed.png)

The authorization proxy will return unauthorized to the user whenever if the role token is invalid, or the role written on the role token has no privilege to access to the resource user requesting.

---

### Mapping rules

The mapping rules describe the elements using in the authorization proxy. The user can configure which Athenz domain's policies cached in the policy updator, and decide if the user is authorized to access the resource.

The mapping rules were described below.

|          | Description                         | Map to (Athenz)  | Example   |   |
|----------|-------------------------------------|------------------|-----------|---|
| Role     | Role name written on the role token | Role             | admin     |   |
| Action   | HTTP/HTTPs request action           | Policy action    | POST      |   |
| Resource | HTTP/HTTPs request resource         | Policy resources | /user/add |   |

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