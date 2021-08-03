package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/yahoojapan/authorization-proxy/v4/config"
)

func TestNewTLSConfig(t *testing.T) {
	type args struct {
		CertPath string
		KeyPath  string
		CAPath   string
		cfg      config.TLS
	}
	defaultArgs := args{
		cfg: config.TLS{
			CertPath: "../test/data/dummyServer.crt",
			KeyPath:  "../test/data/dummyServer.key",
			CAPath:   "../test/data/dummyCa.pem",
		},
	}

	tests := []struct {
		name       string
		args       args
		want       *tls.Config
		beforeFunc func(args args)
		checkFunc  func(*tls.Config, *tls.Config) error
		afterFunc  func(args args)
		wantErr    bool
	}{
		{
			name: "return value MinVersion test.",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.MinVersion != want.MinVersion {
					return fmt.Errorf("MinVersion not Matched :\tgot %d\twant %d", got.MinVersion, want.MinVersion)
				}
				return nil
			},
		},
		{
			name: "return value CurvePreferences test.",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if len(got.CurvePreferences) != len(want.CurvePreferences) {
					return fmt.Errorf("CurvePreferences not Matched length:\tgot %d\twant %d", len(got.CurvePreferences), len(want.CurvePreferences))
				}
				for _, actualValue := range got.CurvePreferences {
					var match bool
					for _, expectedValue := range want.CurvePreferences {
						if actualValue == expectedValue {
							match = true
							break
						}
					}

					if !match {
						return fmt.Errorf("CurvePreferences not Find :\twant %d", want.MinVersion)
					}
				}
				return nil
			},
		},
		{
			name: "return value SessionTicketsDisabled test.",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.SessionTicketsDisabled != want.SessionTicketsDisabled {
					return fmt.Errorf("SessionTicketsDisabled not matched :\tgot %v\twant %v", got.SessionTicketsDisabled, want.SessionTicketsDisabled)
				}
				return nil
			},
		},
		{
			name: "return value Certificates test.",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				for _, wantVal := range want.Certificates {
					var notExist = false
					for _, gotVal := range got.Certificates {
						if gotVal.PrivateKey == wantVal.PrivateKey {
							notExist = true
							break
						}
					}
					if notExist {
						return fmt.Errorf("Certificates PrivateKey not Matched :\twant %s", wantVal.PrivateKey)
					}
				}
				return nil
			},
		},
		{
			name: "return value ClientAuth test.",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.ClientAuth != want.ClientAuth {
					return fmt.Errorf("ClientAuth not Matched :\tgot %d \twant %d", got.ClientAuth, want.ClientAuth)
				}
				return nil
			},
		},
		{
			name: "cert file not found return value Certificates test.",
			args: args{
				cfg: config.TLS{
					CertPath: "",
				},
			},
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates:           nil,
				ClientAuth:             tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.Certificates != nil {
					return fmt.Errorf("Certificates not nil")
				}
				return nil
			},
		},
		{
			name: "cert file not found return value ClientAuth test.",
			args: args{
				cfg: config.TLS{
					CertPath: "",
				},
			},
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates:           nil,
				ClientAuth:             tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.Certificates != nil {
					return fmt.Errorf("Certificates not nil")
				}
				return nil
			},
		},
		{
			name: "CA file not found return value ClientAuth test.",
			args: args{
				cfg: config.TLS{
					CertPath: "",
					CAPath:   "",
				},
			},

			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
					tls.X25519,
				},
				SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertPath, defaultArgs.KeyPath)
					return []tls.Certificate{cert}
				}(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.ClientAuth != 0 {
					return fmt.Errorf("ClientAuth is :\t%d", got.ClientAuth)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt.args)
			}

			got, err := NewTLSConfig(tt.args.cfg)
			if err != nil {
				t.Errorf("NewTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if err != nil {
					t.Errorf("NewTLSConfig() error = %v", err)
					return
				}
			}

			if tt.afterFunc != nil {
				tt.afterFunc(tt.args)

			}
		})
	}
}

func TestNewX509CertPool(t *testing.T) {
	type args struct {
		path string
	}

	tests := []struct {
		name      string
		args      args
		want      *x509.CertPool
		checkFunc func(*x509.CertPool, *x509.CertPool) error
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "Check err if file is not exists",
			args: args{
				path: "",
			},
			want: &x509.CertPool{},
			checkFunc: func(*x509.CertPool, *x509.CertPool) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "Check Append CA is correct",
			args: args{
				path: "../test/data/dummyCa.pem",
			},
			want: func() *x509.CertPool {
				wantPool := x509.NewCertPool()
				c, err := ioutil.ReadFile("../test/data/dummyCa.pem")
				if err != nil {
					panic(err)
				}
				if !wantPool.AppendCertsFromPEM(c) {
					panic(errors.New("Error appending certs from PEM"))
				}
				return wantPool
			}(),
			checkFunc: func(want *x509.CertPool, got *x509.CertPool) error {
				for _, wantCert := range want.Subjects() {
					exists := false
					for _, gotCert := range got.Subjects() {
						if strings.EqualFold(string(wantCert), string(gotCert)) {
							exists = true
						}
					}
					if !exists {
						return fmt.Errorf("Error\twant\t%s\t not found", string(wantCert))
					}
				}
				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewX509CertPool(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewX509CertPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				err = tt.checkFunc(tt.want, got)
				if err != nil {
					t.Errorf("TestNewX509CertPool error = %s", err)
				}
			}
		})
	}
}
