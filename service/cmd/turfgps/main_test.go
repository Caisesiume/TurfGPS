package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestServeAnswersRequestThenStopsOnCancel covers the two properties the
// process needs to be useful in a container: it answers a request while it is
// up, and it returns cleanly when it is asked to stop.
func TestServeAnswersRequestThenStopsOnCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- serve(ctx, ln) }()

	resp, err := http.Get("http://" + ln.Addr().String() + "/")
	if err != nil {
		cancel()
		t.Fatalf("get: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cancel()
		t.Fatalf("read body: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		cancel()
		t.Fatalf("close body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if len(body) == 0 {
		t.Error("body is empty, want a non-empty response")
	}

	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("serve returned %v, want nil after cancellation", err)
		}
	case <-time.After(shutdownTimeout + 2*time.Second):
		t.Error("serve did not return after cancellation")
	}
}
