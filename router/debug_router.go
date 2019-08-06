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
	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/service"
)

func NewDebugRouter(cfg config.Server, a service.Authorizationd) *http.ServeMux {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 32
	mux := http.NewServeMux()

	dur, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		dur = time.Second * 3
	}

	for _, route := range NewDebugRoutes(cfg.DebugServer, a) {
		//関数名取得
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
