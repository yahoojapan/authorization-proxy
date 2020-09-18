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

package service

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/v4/config"
)

// NewTLSConfig returns a *tls.Config struct or error.
// It reads TLS configuration and initializes *tls.Config struct.
// It initializes TLS configuration, for example the CA certificate and key to start TLS server.
// Server and CA Certificate, and private key will read from files from file paths defined in environment variables.
func NewTLSConfig(cfg config.TLS) (*tls.Config, error) {
	t := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
			tls.X25519,
		},
		SessionTicketsDisabled: true,
		// PreferServerCipherSuites: true,
		// CipherSuites: []uint16{
		// tls.TLS_RSA_WITH_RC4_128_SHA,
		// tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		// tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		// tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, // Maybe this is work on TLS 1.2
		// tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, // TLS1.3 Feature
		// tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, // TLS1.3 Feature
		// tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
		// tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
		// },
		ClientAuth: tls.NoClientCert,
	}

	cert := config.GetActualValue(cfg.CertPath)
	key := config.GetActualValue(cfg.KeyPath)
	ca := config.GetActualValue(cfg.CAPath)

	if cert != "" && key != "" {
		crt, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, errors.Wrap(err, "tls.LoadX509KeyPair(cert, key)")
		}
		t.Certificates = make([]tls.Certificate, 1)
		t.Certificates[0] = crt
	}

	if ca != "" {
		pool, err := NewX509CertPool(ca)
		if err != nil {
			return nil, errors.Wrap(err, "NewX509CertPool(ca)")
		}
		t.ClientCAs = pool
		t.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return t, nil
}

// NewX509CertPool returns *x509.CertPool struct or error.
// The CertPool will read the certificate from the path, and append the content to the system certificate pool.
func NewX509CertPool(path string) (*x509.CertPool, error) {
	var pool *x509.CertPool
	c, err := ioutil.ReadFile(path)
	if err == nil && c != nil {
		pool, err = x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(c) {
			err = errors.New("Certification Failed")
		}
	}
	return pool, errors.Wrap(err, "x509.SystemCertPool()")
}
