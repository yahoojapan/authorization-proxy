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

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"

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
				path: "../test/data/not_valid_config.yaml",
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

				err = f.Chmod(0o000)
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
				path: "../test/data/example_config.yaml",
			},
			want: &Config{
				Version: "v2.0.0",
				Server: Server{
					Port:            8082,
					Timeout:         "10s",
					ShutdownTimeout: "10s",
					ShutdownDelay:   "9s",
					TLS: TLS{
						Enable:   true,
						CertPath: "test/data/dummyServer.crt",
						KeyPath:  "test/data/dummyServer.key",
						CAPath:   "test/data/dummyCa.pem",
					},
					HealthCheck: HealthCheck{
						Port:     6082,
						Endpoint: "/healthz",
					},
					Debug: Debug{
						Enable:    false,
						Port:      6083,
						Dump:      true,
						Profiling: true,
					},
				},
				Athenz: Athenz{
					URL:     "https://athenz.io:4443/zts/v1",
					Timeout: "30s",
					CAPath:  "_athenz_root_ca_",
				},
				Proxy: Proxy{
					Scheme:                 "http",
					Host:                   "localhost",
					Port:                   80,
					BufferSize:             4096,
					OriginHealthCheckPaths: []string{},
				},
				Authorization: Authorization{
					PublicKey: PublicKey{
						SysAuthDomain:   "sys.auth",
						RefreshPeriod:   "24h",
						ETagExpiry:      "168h",
						ETagPurgePeriod: "84h",
					},
					AthenzDomains: []string{
						"provider-domain1",
						"provider-domain2",
					},
					Policy: Policy{
						ExpiryMargin:  "48h",
						RefreshPeriod: "1h",
						PurgePeriod:   "24h",
						RetryDelay:    "",
						RetryAttempts: 0,
						MappingRules: map[string][]authorizerd.Rule{
							"domain1": {
								authorizerd.Rule{
									Method:   "get",
									Path:     "/path1/{path2}",
									Action:   "action",
									Resource: "path1.{path2}",
								},
								authorizerd.Rule{
									Method:   "get",
									Path:     "/path?param={value}",
									Action:   "action",
									Resource: "path.{value}",
								},
							},
							"domain2": {
								authorizerd.Rule{
									Method:   "get",
									Path:     "/path1/{path2}?param={value}",
									Action:   "action",
									Resource: "{path2}.{value}",
								},
							},
						},
					},
					JWK: JWK{
						RefreshPeriod: "",
						RetryDelay:    "",
						URLs:          []string{"http://your-jwk-set-url1", "https://your-jwk-set-url2"},
					},
					AccessToken: AccessToken{
						Enable:               true,
						VerifyCertThumbprint: true,
						VerifyClientID:       true,
						AuthorizedClientIDs: map[string][]string{
							"common_name1": {"client_id1", "client_id2"},
							"common_name2": {"client_id1", "client_id2"},
						},
						CertBackdateDuration: "1h",
						CertOffsetDuration:   "1h",
					},
					RoleToken: RoleToken{
						Enable:         true,
						RoleAuthHeader: "Athenz-Role-Auth",
					},
				},
				Log: Log{
					Level: "debug",
					Color: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				if err := tt.beforeFunc(); err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
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
			want: "v2.0.0",
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
				if err := tt.beforeFunc(); err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
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
