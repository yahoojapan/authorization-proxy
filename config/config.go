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

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	// currentVersion represents the current configuration version.
	currentVersion = "v2.0.0"
)

// Config represents the configuration (config.yaml) of authorization proxy.
type Config struct {
	// Version represents the configuration file version.
	Version string `yaml:"version"`

	// Server represents the authorization proxy and the health check server configuration.
	Server Server `yaml:"server"`

	// Athenz represents the Athenz server connection configuration.
	Athenz Athenz `yaml:"athenz"`

	// Proxy represents the proxy destination configuration.
	Proxy Proxy `yaml:"proxy"`

	// Authorization represents the detail authorization configuration.
	Authorization Authorization `yaml:"authorization"`

	// Log represents the logger configuration.
	Log Log `yaml:"log"`
}

// Server represents the authorization proxy and the health check server configuration.
type Server struct {
	// Port represents the server listening port.
	Port int `yaml:"port"`

	// Timeout represents the maximum request handling duration.
	Timeout string `yaml:"timeout"`

	// ShutdownTimeout represents the duration before force shutdown.
	ShutdownTimeout string `yaml:"shutdownTimeout"`

	// ShutdownDelay represents the delay duration between the health check server shutdown and the client sidecar server shutdown.
	ShutdownDelay string `yaml:"shutdownDelay"`

	// TLS represents the TLS configuration of the authorization proxy.
	TLS TLS `yaml:"tls"`

	// HealthCheck represents the health check server configuration.
	HealthCheck HealthCheck `yaml:"healthCheck"`

	// Debug represents the debug server configuration.
	Debug Debug `yaml:"debug"`
}

// TLS represents the TLS configuration of the authorization proxy.
type TLS struct {
	// Enable represents whether to enable TLS.
	Enable bool `yaml:"enable"`

	// CertPath represents the server certificate file path.
	CertPath string `yaml:"certPath"`

	// KeyPath represents the private key file path of the server certificate.
	KeyPath string `yaml:"keyPath"`

	// CAPath represents the CA certificate chain file path for verifying client certificates.
	CAPath string `yaml:"caPath"`
}

// HealthCheck represents the health check server configuration.
type HealthCheck struct {
	// Port represents the server listening port.
	Port int `yaml:"port"`

	// Endpoint represents the health check endpoint (pattern).
	Endpoint string `yaml:"endpoint"`
}

// Debug represents the debug server configuration.
type Debug struct {
	// Enable represents if user want to enable debug server functionality.
	Enable bool `yaml:"enable"`

	// Port represents debug server port.
	Port int `yaml:"port"`

	// Dump represents whether to enable memory dump functionality.
	Dump bool `yaml:"dump"`

	// Profiling represents whether to enable profiling functionality.
	Profiling bool `yaml:"profiling"`
}

// Athenz represents the Athenz server connection configuration.
type Athenz struct {
	// URL represents the Athenz (ZMS or ZTS) API URL.
	URL string `yaml:"url"`

	// Timeout represents the request timeout duration to Athenz server.
	Timeout string `yaml:"timeout"`

	// CAPath represents the CA certificate chain file path for verifying Athenz server certificate.
	CAPath string `yaml:"caPath"`
}

// Proxy represents the proxy destination configuration.
type Proxy struct {
	// Scheme represents the HTTP URL scheme of the proxy destination, default is http.
	Scheme string `yaml:"scheme"`

	// Host represents the proxy destination host, for example, localhost.
	Host string `yaml:"host"`

	// Port represents the proxy destination port number.
	Port uint16 `yaml:"port"`

	// BufferSize represents the reverse proxy buffer size.
	BufferSize uint64 `yaml:"bufferSize"`

	// OriginHealthCheckPaths represents health check paths of your origin application.
	// WARNING!!! Setting this configuration may introduce security hole in your system. ONLY set this configuration as the application's health check endpoint.
	// Tips for performance: define your health check endpoint with a different length from the most frequently used endpoint, for example, use `/healthcheck` (len: 12) when `/most_used` (len: 10), instead of `/healthccc` (len: 10)
	OriginHealthCheckPaths []string `yaml:"originHealthCheckPaths"`
}

// Authorization represents the detail authorization configuration.
type Authorization struct {
	// AthenzDomains represents Athenz domains containing the RBAC policies.
	AthenzDomains []string `yaml:"athenzDomains"`

	// PublicKey represents the configuration to fetch Athenz public keys.
	PublicKey PublicKey `yaml:"publicKey"`

	// Policy represents the configuration to fetch Athenz policies.
	Policy Policy `yaml:"policy"`

	// JWK represents the configuration to fetch Athenz JWK.
	JWK JWK `yaml:"jwk"`

	// AccessToken represents the configuration to control access token verification.
	AccessToken AccessToken `yaml:"accessToken"`

	// RoleToken represents the configuration to control role token verification.
	RoleToken RoleToken `yaml:"roleToken"`
}

// PublicKey represents the configuration to fetch Athenz public keys.
type PublicKey struct {
	// SysAuthDomain represents the system authentication domain of Athenz.
	SysAuthDomain string `yaml:"sysAuthDomain"`

	// RefreshPeriod represents the duration of the refresh period.
	RefreshPeriod string `yaml:"refreshPeriod"`

	// RetryDelay represents the duration between each retry.
	RetryDelay string `yaml:"retryDelay"`

	// ETagExpiry represents the duration before Etag expires.
	ETagExpiry string `yaml:"eTagExpiry"`

	// ETagPurgePeriod represents the duration of purging expired items in the ETag cache.
	ETagPurgePeriod string `yaml:"eTagPurgePeriod"`
}

// Policy represents the configuration to fetch Athenz policies.
type Policy struct {
	// Disable decides whether to check the policy.
	Disable bool `yaml:"disable"`

	// ExpiryMargin represents the policy expiry margin to force refresh policies beforehand.
	ExpiryMargin string `yaml:"expiryMargin"`

	// RefreshPeriod represents the duration of the refresh period.
	RefreshPeriod string `yaml:"refreshPeriod"`

	// PurgePeriod represents the duration of purging expired items in the cache.
	PurgePeriod string `yaml:"purgePeriod"`

	// RetryDelay represents the duration between each retry.
	RetryDelay string `yaml:"retryDelay"`

	// RetryAttempts represents number of attempts to retry.
	RetryAttempts int `yaml:"retryAttempts"`

	// MappingRules represents translation rules for determining action and resource.
	MappingRules map[string][]authorizerd.Rule `yaml:"mappingRules"`
}

// JWK represents the configuration to fetch Athenz JWK.
type JWK struct {
	// RefreshPeriod represents the duration of the refresh period.
	RefreshPeriod string `yaml:"refreshPeriod"`

	// RetryDelay represents the duration between each retry.
	RetryDelay string `yaml:"retryDelay"`

	// URLs represents URLs that delivers JWK Set excluding athenz.
	URLs []string `yaml:"urls"`
}

// AccessToken represents the configuration to control access token verification.
type AccessToken struct {
	// Enable decides whether to verify access token.
	Enable bool `yaml:"enable"`

	// VerifyCertThumbprint represents whether to enforce certificate thumbprint verification.
	VerifyCertThumbprint bool `yaml:"verifyCertThumbprint"`

	// VerifyClientID represents whether to enforce certificate common name and client_id verification.
	VerifyClientID bool `yaml:"verifyClientID"`

	// AuthorizedClientIDs represents list of allowed client_id and common name.
	AuthorizedClientIDs map[string][]string `yaml:"authorizedClientIDs"`

	// CertBackdateDuration represents the certificate issue time backdating duration. (for usecase: new cert + old token)
	CertBackdateDuration string `yaml:"certBackdateDuration"`

	// CertOffsetDuration represents the certificate issue time offset when comparing with the issue time of the access token. (for usecase: new cert + old token)
	CertOffsetDuration string `yaml:"certOffsetDuration"`
}

// RoleToken represents the configuration to control role token verification.
type RoleToken struct {
	// Enable decides whether to verify role token.
	Enable bool `yaml:"enable"`

	// RoleAuthHeader represents the HTTP header for extracting the role token.
	RoleAuthHeader string `yaml:"roleAuthHeader"`
}

// Log represents the logger configuration.
type Log struct {
	// Level represents the logger output level. Values: "debug", "info", "warn", "error", "fatal".
	Level string `yaml:"level"`

	// Color represents whether to print ANSI escape code.
	Color bool `yaml:"color"`
}

// New returns the decoded configuration YAML file as *Config struct. Returns non-nil error if any.
func New(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o600)
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

// GetVersion returns the current configuration version of the authorization proxy.
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
