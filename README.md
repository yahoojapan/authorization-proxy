# Provider Sidecar

<img src="https://github.com/yahoojapan/authorization-proxy/raw/master/images/logo.png" width="200">

---

## What is Provider Sidecar
Provider Sidecar is API for a Kubernetes authentication and authorization webhook that integrates with
[Athenz](https://github.com/yahoo/athenz) for access checks. It allows flexible resource
mapping from K8s resources to Athenz ones.

You can also use just the authorization hook without also using the authentication hook.
Use of the authentication hook requires Athenz to be able to sign tokens for users.

Requires go 1.9 or later.

---

## Use case
### Authorization
![Use case](./doc/assets/use-case.png)

1. K8s webhook request (SubjectAccessReview) ([Webhook Mode - Kubernetes](https://kubernetes.io/docs/reference/access-authn-authz/webhook/))
	- the K8s API server wants to know if the user is allowed to do the requested action
2. Athenz RBAC request ([Athenz](http://www.athenz.io/))
	- Athenz server contains the user authorization information for access control
	- ask Athenz server is the user action is allowed based on pre-configurated policy

Provider Sidecar convert the K8s request to Athenz request based on the mapping rules in `config.yaml` ([example](./config/example_config.yaml)).
- [conversion logic](./doc/authorization-proxy-functional-overview.md)
- [config details](./doc/config-detail.md)

P.S. It is just a sample deployment solution above. Provider Sidecar can work on any environment as long as it can access both the API server and the Athenz server.

---

## Specification
```
Under construction...
```

### Configuration
- [config.go](./config/config.go)
- [config details](./doc/config-detail.md)

---

## Contact
```
Under construction...
```

## License
```
Under construction...
```

