package main

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

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
							AthenzConfRefreshDuration: "dummy",
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					got := run(cfg)
					want := "usecase returned error: cannot newAuthorizationd(cfg): error create athenzConfd: invalid refresh druation: time: invalid duration dummy"
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
				name: "run success",
				args: args{
					cfg: config.Config{
						Debug: true,
						Athenz: config.Athenz{
							URL: "athenz.com",
						},
						Authorization: config.Authorization{
							AthenzConfRefreshDuration: "10s",
							AthenzConfSysAuthDomain:   "dummy.sys.auth",
							AthenzConfEtagExpTime:     "10s",
							AthenzConfEtagFlushDur:    "10s",
							AthenzDomains:             []string{"dummyDom1", "dummyDom2"},
							PolicyExpireMargin:        "10s",
							PolicyRefreshDuration:     "10s",
							PolicyEtagExpTime:         "10s",
							PolicyEtagFlushDur:        "10s",
						},
						Server: config.Server{
							HealthzPath: "/dummy",
						},
						Proxy: config.Proxy{
							BufferSize: 512,
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					var got []error
					mux := &sync.Mutex{}
					go func() {
						mux.Lock()
						got = run(cfg)
						mux.Unlock()
					}()
					time.Sleep(time.Second)
					syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

					time.Sleep(time.Second)
					mux.Lock()
					defer mux.Unlock()
					if len(got) != 1 {
						return errors.Errorf("got: %v", got)
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
