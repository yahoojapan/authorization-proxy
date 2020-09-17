# Features to Debug

<a id="markdown-table-of-contents" name="table-of-contents"></a>
## Table of Contents

<!-- TOC depthFrom:2 -->

- [Features to Debug](#features-to-debug)
    - [Table of Contents](#table-of-contents)
    - [Get policy cache](#get-policy-cache)
        - [Configuration](#configuration)
        - [Example:](#example)
    - [Profiling](#profiling)
        - [Configuration](#configuration-1)

<!-- /TOC -->

<a id="markdown-get-policy-cache" name="get-policy-cache"></a>
## Get policy cache

- Only accepts HTTP `GET` request
- The endpoint is `/debug/cache/policy`
- Response body contains below information in JSON format.
- It will expose the entire policy cache to the client.

<a id="markdown-configuration" name="configuration"></a>
### Configuration

Example configuration for debug policy cache interface:

```yaml
version: v2.0.0
server:
  debug:
    enable: true
    port: 6083
    dump: true
...
```

The example configuration file is [here](../test/data/example_config.yaml). For more information, please refer to [config.go](../config/config.go).

<a id="markdown-example" name="example"></a>
### Example:

```bash
curl -X GET http://127.0.0.1:6083/debug/cache/policy
```

Output:

```json
{
    "domain1:role.role1":  [
        {
           "resource_domain": "resource_domain1",
            "effect": null,
            "action": "action_name1",
            "resource": "resource_name1",
            "action_regexp_string": "^action_name1$",
            "resource_regexp_string": "^resource_name1$"
        },
        {
           "resource_domain": "resource_domain2",
            "effect": null,
            "action": "*",
            "resource": "*",
            "action_regexp_string": "^.*$",
            "resource_regexp_string": "^.*$"
        },
    ],
    "domain2:role.role2":  [
        {
           "resource_domain": "resource_domain3",
            "effect": null,
            "action": "action_name3",
            "resource": "resource_name3",
            "action_regexp_string": "^action_name3$",
            "resource_regexp_string": "^resource_name3$"
        },
    ]
}
```

<a id="markdown-profiling" name="profiling"></a>
## Profiling

- Only accepts HTTP `GET` request
- The endpoint is `/debug/pprof`
- User can access this endpoint though web browser.

<a id="markdown-configuration-1" name="configuration-1"></a>
### Configuration

Example configuration for profiling interface:

```yaml
version: v2.0.0
server:
  debug:
    enable: true
    port: 6083
    profiling: true
...
```

The example configuration file is [here](../test/data/example_config.yaml). For more information, please refer to [config.go](../config/config.go).
