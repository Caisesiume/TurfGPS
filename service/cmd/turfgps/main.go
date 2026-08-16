// Command turfgps runs the TurfGPS service as one long-lived server process,
// built as a single self-contained executable per NFR-003.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// defaultAddr is where the server listens. All interfaces, because the
	// process is expected to be the only thing in its container.
	defaultAddr = ":8080"

	// shutdownTimeout bounds the drain of in-flight requests once the process
	// has been asked to stop. It stays well inside a container runtime's usual
	// grace period, which kills the process outright when that period expires.
	shutdownTimeout = 5 * time.Second

	// readHeaderTimeout bounds how long a client may take to send its request
	// headers, so an idle connection cannot hold a server goroutine open.
	readHeaderTimeout = 10 * time.Second

	// readTimeout bounds the request as a whole, headers and body together.
	// readHeaderTimeout above stops applying once the headers are in, so
	// without this a client may send a well-formed header block and then
	// trickle a body for as long as it likes.
	readTimeout = 30 * time.Second

	// writeTimeout bounds the response. Go arms the write deadline as soon as
	// the request headers have been read, so this budget covers reading the
	// body and running the handler as well as writing the reply — which is why
	// it sits above readTimeout rather than beside it. A handler that
	// legitimately needs longer extends its own deadline through
	// http.ResponseController; raising this ceiling for one slow route would
	// relax it for every route.
	writeTimeout = 60 * time.Second

	// idleTimeout bounds how long a kept-alive connection may wait between
	// requests. Go falls back on readTimeout when this is unset, which is the
	// wrong budget for it: a connection waiting to be reused is not a slow
	// client, and closing healthy connections every 30 seconds would cost a
	// handshake per client per half-minute for nothing.
	idleTimeout = 120 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		slog.Error("cannot listen", "addr", defaultAddr, "error", err)
		os.Exit(1)
	}

	if err := serve(ctx, ln); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

// serve runs the HTTP server on ln until ctx is cancelled, then drains
// in-flight requests and returns. It takes ownership of ln and closes it.
func serve(ctx context.Context, ln net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("TurfGPS service\n"))
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	slog.Info("listening", "addr", ln.Addr().String())

	select {
	case err := <-served:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	slog.Info("stopping", "drain", shutdownTimeout)

	drainCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(drainCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
