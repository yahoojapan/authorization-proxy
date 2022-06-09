# Athenz provider sidecar with gRPC support

## How do we implement this feature?

Since provider sidecar is implemented in go, we decided to find a library that supports like reverse proxy feature in gRPC interface, and we found [this library](https://github.com/mwitkow/grpc-proxy).

Unfortunately this library is quite old (last commit is 3 years ago), but this is the only library we can found to support our usecase, so we decided to give it a try.

## Making code changes

We implemented the feature with the below attention:

1. Match the current code structure
2. No breakable changes from the user side

For 2, it is mainly focusing on the provider sidecar configuration.

When the user uses the legacy provider sidecar configuration file with the new version, it will still work without any update on the configuration file.

### Implementation

File changes:

https://github.com/yahoojapan/authorization-proxy/pull/83/files

We have updated/changed the implementation in the following layers:

- usecase
- service
- handler

#### Usecase layer

In this layer, we create gRPC handler and pass it to service layer.

[Reference](https://github.com/yahoojapan/authorization-proxy/blob/1e14186eb1dd959e246a18be98c92d40a677a56e/usecase/authz_proxyd.go#L71-L84)

#### Service layer

In service layer, we implemented server startup logic. When the handler created from usecase layer is nil, the HTTP mode will be started like before.

#### Handler layer

In this layer, we implemented gRPC reverse proxy handler.

When the value of the configuration `proxy.scheme` is set to `grpc`, the gRPC handler will be created, and the server will start with gRPC mode.

[Reference](https://github.com/yahoojapan/authorization-proxy/blob/1e14186eb1dd959e246a18be98c92d40a677a56e/config/config.go#L133)

If it is not `grpc`, nil will be returned, and the service layer will start with HTTP mode.

[Reference](https://github.com/yahoojapan/authorization-proxy/blob/1e14186eb1dd959e246a18be98c92d40a677a56e/handler/grpc.go)

It retrieves the role token from the gRPC metadata, and authorize it using the athenz-authorizer.

If authorization succeeded, the gRPC request will proxy to the backend.

## Configuration

In handler layer, the gRPC call will be authenticated and authorized by athenz policy.

Setting the athenz policy is almost the same as before, other than the resource set on the resource.

The resource name is defined in the proto files [here](https://github.com/vdaas/vald/tree/master/apis/proto/v1/vald), following by the following scheme.

`/<package name>.<service name>/<rpc name>`

For example, Vald provides an interface for users to insert vector. Here is the proto file:

https://github.com/vdaas/vald/blob/master/apis/proto/v1/vald/insert.proto

```proto
syntax = "proto3";

package vald.v1;

import "apis/proto/v1/payload/payload.proto";
import "github.com/googleapis/googleapis/google/api/annotations.proto";

option go_package = "github.com/vdaas/vald/apis/grpc/v1/vald";
option java_multiple_files = true;
option java_package = "org.vdaas.vald.api.v1.vald";
option java_outer_classname = "ValdInsert";

service Insert {

  rpc Insert(payload.v1.Insert.Request) returns (payload.v1.Object.Location) {
    option (google.api.http) = {
      post : "/insert"
      body : "*"
    };
  }

  rpc StreamInsert(stream payload.v1.Insert.Request)
      returns (stream payload.v1.Object.StreamLocation) {}

  rpc MultiInsert(payload.v1.Insert.MultiRequest)
      returns (payload.v1.Object.Locations) {
    option (google.api.http) = {
      post : "/insert/multiple"
      body : "*"
    };
  }
}
```

Following the syntax, to configure the resource in policy should be `/vald.v1.insert/insert`.

For another gRPC interfaces, it should be the same.

The policy action is `grpc`, which is hardcoded in the source code.

## Design

### Athenz Provider Sidecar

Athenz provider sidecar can start with either gRPC mode and HTTP mode at the same time. The reasons are:

- We wanted to make minimal changes to it
- Supporting both gRPC and HTTP mode at the same time causes big changes on configuration file, and it may lead to breaking changes
- Also there are no such requirement from users

### Athenz Policy

To design Athenz policy configuration, there are 2 fields we need to think about:

- Action
- Resources

#### Policy Action

In the world of HTTP, different HTTP methods are supported, like `GET` and `POST`, and these value is used in action field.

But in gRPC, there are no such concept.
For each RPC endpoint, only 1 resource is supported.

But [gRPC supports 4 different types](https://grpc.io/docs/what-is-grpc/core-concepts/#rpc-life-cycle):

- Unary RPC
- Server streaming RPC
- Client streaming RPC
- Bidirectional streaming RPC

Due to the limitation of gRPC, each RPC endpoint support only 1 resource, a separate endpoint is required for each RPC type.

For the reasons above, currently Vald team decided to hardcode `grpc` in the action field and use the when performing authentication and authorization check.

[Reference](https://github.com/yahoojapan/authorization-proxy/blob/1e14186eb1dd959e246a18be98c92d40a677a56e/handler/grpc.go#L67)

#### Policy Resources

For HTTP mode of the provider sidecar, the HTTP url is set as the policy resources.

In gRPC, each RPC is differentiate with the gRPC method.

Like explained above, the gRPC method name is named by the following rule.

`/<package name>.<service name>/<rpc name>`

In Vald, each functionality is divided into different service. For example insert service, update service, delete service and etc.

For each services, each types (4 types explained above, e.g. unary, server streaming, etc.) are configured into different RPC.

We can easily control the authorization rule for each functionality by using wildcard resource in Vald.

For example we can easily enable or disable all insert resources for the user by configuring the Athenz policy like:

`ALLOW	*	<athenz.domain>:<role.name>	<athenz.domain>:/vald.v1.insert/*`

## Others

- The license of the library is apache, it should be fine :)

https://github.com/mwitkow/grpc-proxy/blob/master/LICENSE.txt

- Authorization proxy using a legacy version of go (1.14 vs 1.17.2), we tried to update it, but the unit test will be failed.
