# Debug Interface

1. `GET /debug/policy-cache`

- Get policy cache from memory (Please use it carefully!!)

## Get policy cache

- Only accept HTTP GET request
- Response body contains below information in JSON format.

Example:

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

### Configuration

Example configuration for debug policy cache interface:

```yaml
version: v1.0.0
server:
  port: 8082
  enable:debug: true
  debug_port: 6083
... 
```
