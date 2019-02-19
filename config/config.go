package config

import (
	"io/ioutil"
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

	// Debug represents debug print function
	Debug bool `yaml:"debug"`

	// Server represents webhook server and health check server configuration.
	Server Server `yaml:"server"`

	// Athenz represents Athenz configuration for Provider Sidecar to connect to Athenz server.
	Athenz Athenz `yaml:"athenz"`

	Proxy Proxy `yaml:"proxy"`

	Provider Provider `yaml:"provider"`
}

// Server represents webhook server and health check server configuration.
type Server struct {
	// Port represents webhook server port.
	Port int `yaml:"port"`

	// HealthzPort represents health check server port.
	HealthzPort int `yaml:"health_check_port"`

	// HealthzPath represents the API path (pattern) for health check server.
	HealthzPath string `yaml:"health_check_path"`

	// Timeout represents the maximum webhook server request handling duration.
	Timeout string `yaml:"timeout"`

	// ShutdownDuration represents the maximum shutdown duration.
	ShutdownDuration string `yaml:"shutdown_duration"`

	// ProbeWaitTime represents the pause duration before shutting down webhook server after health check server shutdown.
	ProbeWaitTime string `yaml:"probe_wait_time"`

	// TLS represents the TLS configuration for webhook server.
	TLS TLS `yaml:"tls"`
}

// TLS represents the TLS configuration for webhook server.
type TLS struct {
	// Enable represents the webhook server enable TLS or not.
	Enabled bool `yaml:"enabled"`

	// CertKey represents the environment variable name having the certificate file path of webhook server.
	CertKey string `yaml:"cert_key"`

	// KeyKey represents the environment variable name having the private key file path of webhook server certificate.
	KeyKey string `yaml:"key_key"`

	// CAKey represents the environment variable name having the CA certificates file path for verifying clients connecting to webhook server.
	CAKey string `yaml:"ca_key"`
}

// Athenz represents the configuration for webhook server to connect to Athenz.
type Athenz struct {
	// URL represents the Athenz (ZMS and ZTS) URL handle authentication and authorization request.
	URL string `yaml:"url"`

	// Timeout represents the request timeout duration to Athenz server.
	Timeout string `yaml:"timeout"`

	// AthenzRootCA is the environment variable name having the Athenz root CA certificate file path for connecting to Athenz.
	AthenzRootCA string `yaml:"root_ca"`
}

// Proxy represent the reverse proxy configuration to connect to Athenz server
type Proxy struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   uint16 `yaml:"port"`

	// RoleHeader represent the HTTP header key name of the role token for Role token proxy request
	RoleHeader string `yaml:"role_header_key"`

	// BufferSize represent the reverse proxy buffer size
	BufferSize uint64 `yaml:"buffer_size"`
}

type Provider struct {
	// athenzConfd parameters
	AthenzConfRefreshDuration string `yaml:"athenzConfRefreshDuration"`
	AthenzConfSysAuthDomain   string `yaml:"athenzConfSysAuthDomain"`
	AthenzConfEtagExpTime     string `yaml:"athenzConfEtagExpTime"`
	AthenzConfEtagFlushDur    string `yaml:"athenzConfEtagFlushDur"`

	// policyd parameters
	AthenzDomains         []string `yaml:"athenzDomains"`
	PolicyExpireMargin    string   `yaml:"policyExpireMargin"`
	PolicyRefreshDuration string   `yaml:"policyRefreshDuration"`
	PolicyEtagFlushDur    string   `yaml:"policyEtagFlushDur"`
	PolicyEtagExpTime     string   `yaml:"policyEtagExpTime"`
}

// New returns the decoded configuration YAML file as *Config struct. Returns non-nil error if any.
func New(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "OpenFile failed")
	}
	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		b, _ := ioutil.ReadFile(path)
		return nil, errors.Wrap(err, "decode file failed \n\n"+string(b))
	}
	return cfg, nil
}

// GetVersion returns the current configuration version of Provider Sidecar.
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
