# Debug Interface

1. `GET /debug/cache/policy`


## Get policy cache

- Only accept HTTP GET request
- Response body contains below information in JSON format.
- Please use it carefully!!

### Configuration

Example configuration for debug policy cache interface:

```yaml
version: v1.0.0
server:
  enable_debug: true
  debug_port: 6083
...
```

For more information, please refer to [config.go](./config/config.go).

### Example:

User can request to get the policy cache by the following command.

```bash
curl http://127.0.0.1:8082/debug/cache/policy
```

The output should something like:

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

