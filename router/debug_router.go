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

package router

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kpango/glg"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/handler"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

// NewDebugRouter return the ServeMux with debug endpoints
func NewDebugRouter(cfg config.Server, a service.Authorizationd) *http.ServeMux {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 32
	mux := http.NewServeMux()

	dur, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		dur = time.Second * 3
	}

	for _, route := range NewDebugRoutes(cfg.Debug, a) {
		mux.Handle(route.Pattern, routing(route.Methods, dur, route.HandlerFunc))
	}

	return mux
}

func routing(m []string, t time.Duration, h handler.Func) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, method := range m {
			if strings.EqualFold(r.Method, method) || method == "*" {

				ctx, cancel := context.WithTimeout(r.Context(), t)
				defer cancel()
				start := time.Now()
				ech := make(chan error)
				go func() {
					ech <- h(w, r.WithContext(ctx))
					close(ech)
				}()

				for {
					select {
					case err := <-ech:
						if err != nil {
							http.Error(w,
								fmt.Sprintf("Error: %s\t%s",
									err.Error(),
									http.StatusText(http.StatusInternalServerError)),
								http.StatusInternalServerError)
							glg.Error(err)
						}
						return
					case <-ctx.Done():
						glg.Errorf("Handler Time Out: %v", time.Since(start))
						return
					}
				}
			}
		}

		_, err := io.Copy(ioutil.Discard, r.Body)
		if err != nil {
			glg.Fatalln(err)
		}
		err = r.Body.Close()
		if err != nil {
			glg.Fatalln(err)
		}
		http.Error(w,
			fmt.Sprintf("Method: %s\t%s",
				r.Method,
				http.StatusText(http.StatusMethodNotAllowed)),
			http.StatusMethodNotAllowed)
	})
}
