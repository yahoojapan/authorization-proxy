package main

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/config"
)

func TestParseParams(t *testing.T) {
	type test struct {
		name       string
		beforeFunc func()
		checkFunc  func(*params) error
		checkErr   bool
	}
	tests := []test{
		func() test {
			return test{
				name: "check parseParams set default value",
				beforeFunc: func() {
					os.Args = []string{""}
				},
				checkFunc: func(p *params) error {
					if p.configFilePath != "/etc/athenz/provider/config.yaml" {
						return errors.Errorf("unexpected file path. got: %s, want: /etc/athenz/provider/config.yaml", p.configFilePath)
					}
					if p.showVersion != false {
						return errors.Errorf("unexpected showVersion flag. got: %v, want : false", p.showVersion)
					}
					return nil
				},
				checkErr: false,
			}
		}(),
		func() test {
			return test{
				name: "check parse error",
				checkFunc: func(p *params) error {
					return nil
				},
				beforeFunc: func() {
					os.Args = []string{"", "-="}
				},
				checkErr: true,
			}
		}(),
		func() test {
			return test{
				name: "check parseParams set user flags",
				beforeFunc: func() {
					os.Args = []string{"", "-f", "/dummy/path", "-version", "true"}
				},
				checkFunc: func(p *params) error {
					if p.configFilePath != "/dummy/path" {
						return errors.Errorf("unexpected file path. got: %s, want: /dummy/path", p.configFilePath)
					}
					if p.showVersion != true {
						return errors.Errorf("unexpected showVersion flag. got: %v, want: true", p.showVersion)
					}

					return nil
				},
				checkErr: false,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}

			got, err := parseParams()
			if err != nil && !tt.checkErr {
				t.Errorf("unexpected error: %v", err)
			}
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("checkFunc() error: %v", err)
			}
		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	type test struct {
		name      string
		args      args
		checkFunc func(config.Config) error
	}
	tests := []test{
		func() test {
			return test{
				name: "run error",
				args: args{
					cfg: config.Config{
						Authorization: config.Authorization{
							PubKeyRefreshDuration: "dummy",
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					got := run(cfg)
					want := "usecase returned error: cannot newAuthzD(cfg): error create pubkeyd: invalid refresh duration: time: invalid duration dummy"
					if len(got) != 1 {
						return errors.New("len(got) != 1")
					}
					if got[0].Error() != want {
						return errors.Errorf("got: %v, want: %v", got[0], want)
					}
					return nil
				},
			}
		}(),
		func() test {
			return test{
				name: "daemon init error",
				args: args{
					cfg: config.Config{
						Athenz: config.Athenz{
							URL: "127.0.0.1",
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					got := run(cfg)
					want := "daemon init error: error when processing pubkey: Error updating ZMS athenz pubkey: error fetch public key entries: error make http request: Get https://127.0.0.1/domain/sys.auth/service/zms: dial tcp 127.0.0.1:443: connect: connection refused"
					if len(got) != 1 {
						return errors.New("len(got) != 1")
					}
					if got[0].Error() != want {
						return errors.Errorf("got: %v, want: %v", got[0], want)
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.checkFunc(tt.args.cfg); err != nil {
				t.Errorf("run() error = %v", err)
			}
		})
	}
}
