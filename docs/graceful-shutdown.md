<a id="markdown-graceful-shutdown" name="graceful-shutdown"></a>
# Graceful shutdown

The sidecar server supports graceful shutdown.
To enable it, set `shutdown_duration` and `probe_wait_time` to value > 0 in the `config.yaml`.

<!-- TOC -->

- [Graceful shutdown](#graceful-shutdown)
    - [Rolling update in K8s with graceful shutdown](#rolling-update-in-k8s-with-graceful-shutdown)

<!-- /TOC -->

<a id="markdown-rolling-update-in-k8s-with-graceful-shutdown" name="rolling-update-in-k8s-with-graceful-shutdown"></a>
## Rolling update in K8s with graceful shutdown

1. make sure the `strategy` is set in the deployment
    - sample
    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    spec:
    strategy:
        rollingUpdate:
            maxSurge: 25%
            maxUnavailable: 25%
        type: RollingUpdate
    ```
1. make sure the `readinessProbe` for sidecar is set
    - sample
    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    spec:
        containers:
        -   name: sidecar
            readinessProbe:
                httpGet:
                    path: /healthz
                    port: 8081
                initialDelaySeconds: 3
                timeoutSeconds: 2
                successThreshold: 1
                failureThreshold: 2
                periodSeconds: 3
    ```
1. make sure the `config.yaml` has the correct value
    - `probe_wait_time = failureThreshold * periodSeconds + timeoutSeconds` (add 1s for buffer)
    - `0 < shutdown_duration < terminationGracePeriodSeconds - probe_wait_time`
    - sample
    ```yaml
    version: "v1.0.0"
    server:
        health_check_port: 8081
        health_check_path: /healthz
        shutdown_duration: 10s
        probe_wait_time: 9s
    ```
1. make sure your application can still handle new requests after shutdown for `probe_wait_time` seconds
