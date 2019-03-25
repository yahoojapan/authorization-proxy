package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/kpango/glg"
	"github.com/yahoojapan/authorization-proxy/config"
)

// Server represents a authorization proxy server behavior
type Server interface {
	ListenAndServe(context.Context) chan []error
}

type server struct {
	// authorization proxy server
	srv        *http.Server
	srvRunning bool

	// Health Check server
	hcsrv     *http.Server
	hcrunning bool

	cfg config.Server

	// ProbeWaitTime
	pwt time.Duration

	// ShutdownDuration
	sddur time.Duration

	// mutext lock variable
	mu sync.RWMutex
}

const (
	// ContentType represents a HTTP header name "Content-Type"
	ContentType = "Content-Type"

	// TextPlain represents a HTTP content type "text/plain"
	TextPlain = "text/plain"

	// CharsetUTF8 represents a UTF-8 charset for HTTP response "charset=UTF-8"
	CharsetUTF8 = "charset=UTF-8"
)

var (
	// ErrContextClosed represents a error that the context is closed
	ErrContextClosed = errors.New("context Closed")
)

// NewServer returns a Server interface, which includes authorization proxy server and health check server structs.
// The authorization proxy server is a http.Server instance, which the port number is read from "config.Server.Port"
// , and set the handler as this function argument "handler".
//
// The health check server is a http.Server instance, which the port number is read from "config.Server.HealthzPort"
// , and the handler is as follow - Handle HTTP GET request and always return HTTP Status OK (200) response.
func NewServer(cfg config.Server, h http.Handler) Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: h,
	}
	srv.SetKeepAlivesEnabled(true)

	hcsrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HealthzPort),
		Handler: createHealthCheckServiceMux(cfg.HealthzPath),
	}
	hcsrv.SetKeepAlivesEnabled(true)

	dur, err := time.ParseDuration(cfg.ShutdownDuration)
	if err != nil {
		dur = time.Second * 5
	}

	pwt, err := time.ParseDuration(cfg.ProbeWaitTime)
	if err != nil {
		pwt = time.Second * 3
	}

	return &server{
		srv:   srv,
		hcsrv: hcsrv,
		cfg:   cfg,
		pwt:   pwt,
		sddur: dur,
	}
}

// ListenAndServe returns a error channel, which includes error returned from authorization proxy server.
// This function start both health check and authorization proxy server, and the server will close whenever the context receive a Done signal.
// Whenever the server closed, the authorization proxy server will shutdown after a defined duration (cfg.ProbeWaitTime), while the health check server will shutdown immediately
func (s *server) ListenAndServe(ctx context.Context) chan []error {
	echan := make(chan []error, 1)

	// error channels to keep track server status
	sech := make(chan error, 1)
	hech := make(chan error, 1)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// start both authorization proxy server and health check server
	go func() {
		s.mu.Lock()
		s.srvRunning = true
		s.mu.Unlock()
		wg.Done()

		glg.Info("authorization proxy api server starting")
		sech <- s.listenAndServeAPI()
		close(sech)

		s.mu.Lock()
		s.srvRunning = false
		s.mu.Unlock()
	}()

	go func() {
		s.mu.Lock()
		s.hcrunning = true
		s.mu.Unlock()
		wg.Done()

		glg.Info("authorization proxy health check server starting")
		hech <- s.hcsrv.ListenAndServe()
		close(hech)

		s.mu.Lock()
		s.hcrunning = false
		s.mu.Unlock()
	}()

	go func() {
		// wait for all server running
		wg.Wait()

		appendErr := func(errs []error, err error) []error {
			if err != nil {
				return append(errs, errors.Wrap(err, "server"))
			}
			return errs
		}

		errs := make([]error, 0, 3)
		for {
			select {
			case <-ctx.Done(): // when context receive done signal, close running servers and return any error
				s.mu.RLock()
				if s.hcrunning {
					glg.Info("authorization proxy health check server will shutdown")
					errs = appendErr(errs, s.hcShutdown(context.Background()))
				}
				if s.srvRunning {
					glg.Info("authorization proxy api server will shutdown")
					errs = appendErr(errs, s.apiShutdown(context.Background()))
				}
				s.mu.RUnlock()
				echan <- appendErr(errs, ctx.Err())
				return

			case err := <-sech: // when authorization proxy server returns, close running health check server and return any error
				if err != nil {
					errs = appendErr(errs, errors.Wrap(err, "close running health check server and return any error"))
				}

				s.mu.RLock()
				if s.hcrunning {
					glg.Info("authorization proxy health check server will shutdown")
					errs = appendErr(errs, s.hcShutdown(ctx))
				}
				s.mu.RUnlock()
				echan <- errs
				return

			case err := <-hech: // when health check server returns, close running authorization proxy server and return any error
				if err != nil {
					errs = append(errs, errors.Wrap(err, "close running authorization proxy server and return any error"))
				}

				s.mu.RLock()
				if s.srvRunning {
					glg.Info("authorization proxy api server will shutdown")
					errs = appendErr(errs, s.apiShutdown(ctx))
				}
				s.mu.RUnlock()
				echan <- errs
				return
			}
		}
	}()

	return echan
}

func (s *server) hcShutdown(ctx context.Context) error {
	hctx, hcancel := context.WithTimeout(ctx, s.sddur)
	defer hcancel()
	return s.hcsrv.Shutdown(hctx)
}

// apiShutdown returns any error when shutdown the authorization proxy server.
// Before shutdown the authorization proxy server, it will sleep config.ProbeWaitTime to prevent any issue from K8s
func (s *server) apiShutdown(ctx context.Context) error {
	time.Sleep(s.pwt)
	sctx, scancel := context.WithTimeout(ctx, s.sddur)
	defer scancel()
	return s.srv.Shutdown(sctx)
}

// createHealthCheckServiceMux return a *http.ServeMux object
// The function will register the health check server handler for given pattern, and return
func createHealthCheckServiceMux(pattern string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, handleHealthCheckRequest)
	return mux
}

// handleHealthCheckRequest is a handler function for and health check request, which always a HTTP Status OK (200) result
func handleHealthCheckRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(ContentType, fmt.Sprintf("%s;%s", TextPlain, CharsetUTF8))
		_, err := fmt.Fprint(w, http.StatusText(http.StatusOK))
		if err != nil {
			glg.Fatal(errors.Wrap(err, "cannot fmt.Fprint(w, http.StatusText(http.StatusOK))"))
		}
	}
}

// listenAndServeAPI return any error occurred when start a HTTPS server, including any error when loading TLS certificate
func (s *server) listenAndServeAPI() error {

	if !s.cfg.TLS.Enabled {
		return s.srv.ListenAndServe()
	}

	cfg, err := NewTLSConfig(s.cfg.TLS)
	if err == nil && cfg != nil {
		s.srv.TLSConfig = cfg
	}
	if err != nil {
		glg.Error(errors.Wrap(err, "cannot NewTLSConfig(s.cfg.TLS)"))
	}
	return s.srv.ListenAndServeTLS("", "")
}
