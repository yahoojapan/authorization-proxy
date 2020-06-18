# Features to Debug

# Table of Contents

1. [Get policy cache](#get-policy-cache)

## Get policy cache

- Only accepts HTTP `GET` request
- The endpoint is `/debug/cache/policy`
- Response body contains below information in JSON format.
- It will expose the entire policy cache to the client.

### Configuration

Example configuration for debug policy cache interface:

```yaml
version: v1.0.0
server:
  debug_server:
    enable: true
    port: 6083
    enable_dump: true
...
```

The example configuration file is [here](../config/testdata/example_config.yaml). For more information, please refer to [config.go](./config/config.go).

### Example:

```bash
curl -X GET http://127.0.0.1:6083/debug/cache/policy
```

Output:

```json
{
    "domain1:role.role1": [
        {
            "resource_domain": "resource_domain1",
            "effect": null,
            "action": "action_name1",
            "resource": "role.role1",
            "regex_string": "^action_name1-role.role1$"
        },
        {
            "resource_domain": "resource_domain2",
            "effect": null,
            "action": "*",
            "resource": "*",
            "regex_string": "^.*-.*$"
        }
    ],
    "domain2:role.role2": [
        {
            "resource_domain": "resource_domain3",
            "effect": null,
            "action": "action_name2",
            "resource": "role.role2",
            "regex_string": "^action_name2-role.role2$"
        }
    ]
}
```

## Profiling

- Only accepts HTTP `GET` request
- The endpoint is `/debug/pprof`
- User can access this endpoint though web browser.

### Configuration

Example configuration for profiling interface:

```yaml
version: v1.0.0
server:
  debug_server:
    enable: true
    port: 6083
    enable_pprof: true
...
```

The example configuration file is [here](../config/testdata/example_config.yaml). For more information, please refer to [config.go](./config/config.go).
