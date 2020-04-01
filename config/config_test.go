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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/kpango/glg"
)

func TestNew(t *testing.T) {
	type args struct {
		path string
	}
	type test struct {
		name       string
		args       args
		beforeFunc func() error
		afterFunc  func() error
		want       *Config
		wantErr    error
	}
	tests := []test{
		{
			name: "Test file content not valid",
			args: args{
				path: "./testdata/not_valid_config.yaml",
			},
			wantErr: fmt.Errorf("decode file failed: yaml: line 11: could not find expected ':'"),
		},
		{
			name: "Open file error",
			args: args{
				path: "./tmp",
			},
			beforeFunc: func() error {
				f, err := os.Create("./tmp")
				if err != nil {
					return err
				}
				defer f.Close()

				err = f.Chmod(0000)
				if err != nil {
					return err
				}

				return nil
			},
			afterFunc: func() error {
				return os.Remove("./tmp")
			},
			wantErr: fmt.Errorf("OpenFile failed: open ./tmp: permission denied"),
		},
		{
			name: "Test file content valid",
			args: args{
				path: "./testdata/example_config.yaml",
			},
			want: &Config{
				Version:            "v1.0.0",
				Debug:              false,
				EnableColorLogging: false,
				Server: Server{
					Port:             8082,
					HealthzPort:      6082,
					HealthzPath:      "/healthz",
					Timeout:          "10s",
					ShutdownDuration: "10s",
					ProbeWaitTime:    "9s",
					TLS: TLS{
						Enabled: false,
						Cert:    "/etc/athenz/provider/keys/server.crt",
						Key:     "/etc/athenz/provider/keys/private.key",
						CA:      "/etc/athenz/provider/keys/ca.crt",
					},
					DebugServer: DebugServer{
						Enable:          false,
						Port:            6083,
						EnableDump:      true,
						EnableProfiling: true,
					},
				},
				Athenz: Athenz{
					URL:          "https://www.athenz.com:4443/zts/v1",
					Timeout:      "30s",
					AthenzRootCA: "",
				},
				Proxy: Proxy{
					Scheme:         "http",
					Host:           "localhost",
					Port:           80,
					RoleHeader:     "Athenz-Role-Auth",
					BufferSize:     4096,
					BypassURLPaths: []string{},
				},
				Authorization: Authorization{
					PubKeyRefreshDuration: "24h",
					PubKeySysAuthDomain:   "sys.auth",
					PubKeyEtagExpTime:     "168h",
					PubKeyEtagFlushDur:    "84h",
					AthenzDomains: []string{
						"domain1",
					},
					PolicyExpireMargin:    "48h",
					PolicyRefreshDuration: "1h",
					PolicyEtagExpTime:     "48h",
					PolicyEtagFlushDur:    "24h",
					Role: Role{
						Enable: true,
					},
					Access: []Access{
						Access{
							Enable:               true,
							VerifyCertThumbprint: true,
							CertBackdateDur:      "1h",
							CertOffsetDur:        "1h",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			got, err := New(tt.args.path)

			if tt.wantErr == nil && err != nil {
				t.Errorf("New() unexpected error, got: %v", err)
				return
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("want error: %v, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("New() error, got: %v, want: %v", err, tt.wantErr)
					return
				}
			}
			glg.Debugf("%v want: %v", got, tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New()= %v, want= %v", got, tt.want)
				return
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test get version return sidecar version",
			want: "v1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVersion(); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetActualValue(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name       string
		args       args
		beforeFunc func() error
		afterFunc  func() error
		want       string
	}{
		{
			name: "GetActualValue without env var",
			args: args{
				val: "test_env",
			},
			want: "test_env",
		},
		{
			name: "GetActualValue with env var",
			args: args{
				val: "_dummy_key_",
			},
			beforeFunc: func() error {
				return os.Setenv("dummy_key", "dummy_value")
			},
			afterFunc: func() error {
				return os.Unsetenv("dummy_key")
			},
			want: "dummy_value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			if got := GetActualValue(tt.args.val); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPrefixAndSuffix(t *testing.T) {
	type args struct {
		str  string
		pref string
		suf  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Check true prefix and suffix",
			args: args{
				str:  "_dummy_",
				pref: "_",
				suf:  "_",
			},
			want: true,
		},
		{
			name: "Check false prefix and suffix",
			args: args{
				str:  "dummy",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
		{
			name: "Check true prefix but false suffix",
			args: args{
				str:  "_dummy",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
		{
			name: "Check false prefix but true suffix",
			args: args{
				str:  "dummy_",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPrefixAndSuffix(tt.args.str, tt.args.pref, tt.args.suf); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
