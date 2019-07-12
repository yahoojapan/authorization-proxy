/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	// currentVersion represents the configuration version.
	currentVersion = "v1.0.0"
)

// Config represents an application configuration content (config.yaml).
// In K8s environment, this configuration is stored in K8s ConfigMap.
type Config struct {
	// Version represents configuration file version.
	Version string `yaml:"version"`

	// Debug represents to print debug message or not.
	Debug bool `yaml:"debug"`

	// EnableColorLogging represents if user want to enable colorized logging.
	EnableColorLogging bool `yaml:"enable_log_color"`

	// Server represents the authorization proxy and health check server configuration.
	Server Server `yaml:"server"`

	// Athenz represents Athenz Data for authorization proxy to connect to Athenz server.
	Athenz Athenz `yaml:"athenz"`

	// Proxy represents the proxy destination of the authorization proxy.
	Proxy Proxy `yaml:"proxy"`

	// Authorization represents the detail configuration of the authorization proxy.
	Authorization Authorization `yaml:"provider"`
}

// Server represents authorization proxy server and health check server configuration.
type Server struct {
	// Port represents the server port.
	Port int `yaml:"port"`

	// EnableDebug represents if user want to enable debug funcationality.
	EnableDebug bool `yaml:"enable_debug"`

	// DebugPort represents debug server port.
	DebugPort int `yaml:"debug_port"`

	// DebugPolicyCachePath represents the policy cache debug path
	DebugPolicyCachePath string `yaml:"debug_policy_cache_path"`

	// HealthzPort represents health check server port.
	HealthzPort int `yaml:"health_check_port"`

	// HealthzPath represents the API path (pattern) for health check server.
	HealthzPath string `yaml:"health_check_path"`

	// Timeout represents the maximum authorization proxy server request handling duration.
	Timeout string `yaml:"timeout"`

	// ShutdownDuration represents the maximum shutdown duration.
	ShutdownDuration string `yaml:"shutdown_duration"`

	// ProbeWaitTime represents the pause duration before shutting down authorization proxy server after health check server shutdown.
	ProbeWaitTime string `yaml:"probe_wait_time"`

	// TLS represents the TLS configuration for authorization proxy server.
	TLS TLS `yaml:"tls"`
}

// TLS represents the TLS configuration for authorization proxy server.
type TLS struct {
	// Enable represents the authorization proxy server enable TLS or not.
	Enabled bool `yaml:"enabled"`

	// Cert represents the certificate file path of authorization proxy server.
	Cert string `yaml:"cert"`

	// Key represents the private key file path of authorization proxy server certificate.
	Key string `yaml:"key"`

	// CA represents the CA certificates file path for verifying clients connecting to authorization proxy server.
	CA string `yaml:"ca"`
}

// Athenz represents the configuration for authorization proxy server to connect to Athenz.
type Athenz struct {
	// URL represents the Athenz (ZMS and ZTS) URL handle authentication and authorization request.
	URL string `yaml:"url"`

	// Timeout represents the request timeout duration to Athenz server.
	Timeout string `yaml:"timeout"`

	// AthenzRootCA is the environment variable name having the Athenz root CA certificate file path for connecting to Athenz.
	AthenzRootCA string `yaml:"root_ca"`
}

// Proxy represent the proxy destination of the authorization proxy.
type Proxy struct {
	// Scheme represent the HTTP URL scheme of the proxy destination, default is http.
	Scheme string `yaml:"scheme"`

	// Host represent the proxy destination host, for example localhost.
	Host string `yaml:"host"`

	// Port represent the proxy destination port number.
	Port uint16 `yaml:"port"`

	// RoleHeader represent the HTTP header key name of the role token for Role token proxy request.
	RoleHeader string `yaml:"role_header_key"`

	// BufferSize represent the reverse proxy buffer size.
	BufferSize uint64 `yaml:"buffer_size"`
}

// Authorization represents the detail configuration of the authorization proxy.
type Authorization struct {
	// PubKeyRefreshDuration represents the refresh duration of Athenz PubKey.
	PubKeyRefreshDuration string `yaml:"pubKeyRefreshDuration"`

	// PubKeySysAuthDomain represents the system authenicate domain of Athenz.
	PubKeySysAuthDomain string `yaml:"pubKeySysAuthDomain"`

	// PubKeyEtagExpTime represents the Etag cache expiration time of Athenz PubKey.
	PubKeyEtagExpTime string `yaml:"pubKeyEtagExpTime"`

	// PubKeyEtagFlushDur represent the Etag cache expiration check duration.
	PubKeyEtagFlushDur string `yaml:"pubKeyEtagFlushDur"`

	// AthenzDomains represents the Athenz domains to fetch the policy.
	AthenzDomains []string `yaml:"athenzDomains"`

	// PolicyExpireMargin represents the policy expire margin to force refresh policies of Athenz domains.
	PolicyExpireMargin string `yaml:"policyExpireMargin"`

	// PolicyRefreshDuration represents the refresh duration of policy.
	PolicyRefreshDuration string `yaml:"policyRefreshDuration"`

	// PolicyEtagExpTime represents the Etag cache expiration time of policy.
	PolicyEtagExpTime string `yaml:"policyEtagExpTime"`

	// PolicyEtagFlushDur represent the Etag cache expiration check duration.
	PolicyEtagFlushDur string `yaml:"policyEtagFlushDur"`
}

// New returns the decoded configuration YAML file as *Config struct. Returns non-nil error if any.
func New(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "OpenFile failed")
	}
	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "decode file failed")
	}
	return cfg, nil
}

// GetVersion returns the current configuration version of Authorization Proxy.
func GetVersion() string {
	return currentVersion
}

// GetActualValue returns the environment variable value if the given val has "_" prefix and suffix, otherwise returns val directly.
func GetActualValue(val string) string {
	if checkPrefixAndSuffix(val, "_", "_") {
		return os.Getenv(strings.TrimPrefix(strings.TrimSuffix(val, "_"), "_"))
	}
	return val
}

// checkPrefixAndSuffix checks if the given string has given prefix and suffix.
func checkPrefixAndSuffix(str, pref, suf string) bool {
	return strings.HasPrefix(str, pref) && strings.HasSuffix(str, suf)
}
