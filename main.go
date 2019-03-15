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
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"path/filepath"
	"runtime"
	"syscall"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/usecase"
)

// params is the data model for Authorization Proxy command line arguments.
type params struct {
	configFilePath string
	showVersion    bool
}

// parseParams parses command line arguments to params object.
func parseParams() (*params, error) {
	p := new(params)
	f := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	f.StringVar(&p.configFilePath,
		"f",
		"/etc/athenz/provider/config.yaml",
		"authorization-proxy config yaml file path")
	f.BoolVar(&p.showVersion,
		"version",
		false,
		"show authorization-proxy version")

	err := f.Parse(os.Args[1:])
	if err != nil {
		return nil, errors.Wrap(err, "Parse Failed")
	}

	return p, nil
}

// run starts the daemon and listens for OS signal.
func run(cfg config.Config) []error {

	g := glg.Get().
		SetLevelMode(glg.LOG, glg.NONE).
		SetLevelMode(glg.PRINT, glg.NONE).
		SetLevelMode(glg.DEBG, glg.NONE)

	if cfg.Debug {
		g.SetLevelMode(glg.DEBG, glg.STD)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	daemon, err := usecase.New(cfg)
	if err != nil {
		return []error{errors.Wrap(err, "usecase returned error")}
	}

	ech := daemon.Start(ctx)
	sigCh := make(chan os.Signal, 1)

	defer func() {
		close(sigCh)
		close(ech)
	}()

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-sigCh:
			cancel()
			glg.Warn("authorization-proxy server shutdown...")
		case errs := <-ech:
			return errs
		}
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(runtime.Error); ok {
				panic(err)
			}
			glg.Error(err)
		}
	}()

	p, err := parseParams()
	if err != nil {
		glg.Fatal(errors.Wrap(err, "parseParams returned error"))
		return
	}

	if p.showVersion {
		glg.Infof("authorization-proxy version -> %s", config.GetVersion())
		return
	}

	cfg, err := config.New(p.configFilePath)
	if err != nil {
		glg.Fatal(errors.Wrap(err, "config instance create error"))
		return
	}

	// check versions between configuration file and config.go
	if cfg.Version != config.GetVersion() {
		glg.Fatal(errors.New("invalid sidecar configuration version"))
		return
	}

	errs := run(*cfg)
	if errs != nil && len(errs) > 0 {
		var emsg string
		for _, err = range errs {
			emsg += "\n" + err.Error()
		}
		glg.Fatal(emsg)
		return
	}
}
