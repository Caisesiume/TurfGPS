package main

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

const (
	// stopBudget bounds how long a test waits for serve to return after
	// cancellation. It is the service's own drain budget plus room for the
	// shutdown poll to notice quiescence, which is why it is a margin on
	// shutdownTimeout rather than a number of its own.
	stopBudget = shutdownTimeout + 2*time.Second

	// inFlightTimeout bounds how long a test waits for the server to reach the
	// point of delivering a response. It bounds the harness rather than the
	// service: reaching it means nothing was ever in flight, so the test that
	// follows would be measuring the drain against an idle connection and
	// proving nothing.
	inFlightTimeout = 10 * time.Second

	// drainWitness is how long a test watches for the drain to abandon a
	// response it has not delivered.
	//
	// It bounds a negative, which is the one thing that cannot be observed
	// directly: nothing can prove serve will never return early, only that it
	// did not while work was outstanding. The window is generous by orders of
	// magnitude rather than finely tuned — a server that drops its connections
	// or returns without draining does both in the same call, microseconds
	// after the cancellation, while a draining server cannot return until the
	// test releases the response.
	drainWitness = 250 * time.Millisecond
)

// TestServeAnswersRequestThenStopsOnCancel covers the two properties the
// process needs to be useful in a container: it answers a request while it is
// up, and it returns cleanly when it is asked to stop.
//
// It cancels with nothing in flight. That the drain finishes work already in
// progress is a separate property, and one this test cannot see: with an idle
// connection, dropping every connection outright is indistinguishable from
// draining. It is measured separately below.
func TestServeAnswersRequestThenStopsOnCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stopped := make(chan error, 1)
	go func() { stopped <- serve(ctx, ln) }()

	// The request carries the test's own context rather than ctx. Cancelling
	// ctx is the stop signal under test, so a request derived from it would be
	// aborted by the very event this test needs it to be independent of;
	// deriving it from t.Context() still keeps an abandoned request from
	// outliving the test.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String()+"/", nil)
	if err != nil {
		t.Fatalf("building the request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
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
	case <-time.After(stopBudget):
		t.Error("serve did not return after cancellation")
	}
}

// TestServeFinishesInFlightResponseOnCancel binds the drain itself: a response
// the service has begun delivering when it is asked to stop is delivered in
// full, serve does not return until it has been, and serve then returns nil.
//
// THE BUG IT CATCHES, and why the test above cannot catch it. Replace the drain
// with an abrupt close — srv.Close() in place of srv.Shutdown(drainCtx) — or
// drop the drain and return as soon as the context is done, and every assertion
// in the test above still passes. With nothing in flight at the moment of
// cancellation the three implementations are indistinguishable: all stop the
// server, all make serve return nil. Only outstanding work separates them, so
// this test constructs some.
//
// HOW THE WORK IS HELD OUTSTANDING, and why it is held at this exact point. The
// handler answers instantly and this test may not change it, so the response is
// held in the one place a test can reach from outside: the connection it is
// written to. The listener below hands the server a connection that pauses on
// its first write, which is the response leaving the handler.
//
// A half-sent request would be the obvious construction and it is wrong.
// net/http checks Server.shuttingDown() immediately after reading a request and
// abandons it — a request whose headers land after the shutdown began is never
// served, by design, and a test built that way goes red against a correct
// drain. What Shutdown promises is work already in progress, which is what this
// test holds.
func TestServeFinishesInFlightResponseOnCancel(t *testing.T) {
	base, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	held := newHeldResponse()
	defer held.resume()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stopped := make(chan error, 1)
	go func() { stopped <- serve(ctx, held.listener(base)) }()

	conn, err := net.Dial("tcp", base.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(inFlightTimeout + stopBudget)); err != nil {
		t.Fatalf("setting the connection deadline: %v", err)
	}

	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: turfgps.test\r\n\r\n")); err != nil {
		t.Fatalf("writing the request: %v", err)
	}

	select {
	case <-held.begun:
	case <-time.After(inFlightTimeout):
		t.Fatalf("the server had not begun delivering a response within %s, so nothing was in flight when it was asked to stop and the rest of this test would measure the drain against an idle connection",
			inFlightTimeout)
	}

	cancel()

	// The response is undelivered, so a drain has work outstanding and cannot
	// have finished. Both failures below are silent in every other test here.
	select {
	case <-held.dropped:
		t.Fatal("the connection carrying an undelivered response was closed at cancellation, want it kept open until the response had been delivered: this is dropping in-flight work, not draining it")
	case err := <-stopped:
		t.Fatalf("serve returned (%v) with a response still undelivered, want it to wait for work already in progress before returning", err)
	case <-time.After(drainWitness):
	}

	held.resume()

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("the response in flight when the service was asked to stop did not arrive: %v; want the drain to deliver it in full", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("the response in flight when the service was asked to stop was cut short: %v; want the drain to deliver it in full", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("the response delivered by the drain has status %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if len(body) == 0 {
		t.Error("the response delivered by the drain is empty, want a non-empty response")
	}

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("serve returned %v, want nil after draining an in-flight response", err)
		}
	case <-time.After(stopBudget):
		t.Errorf("serve did not return within %s of cancellation after the in-flight response was delivered", stopBudget)
	}
}

// heldResponse holds the server's first write to a connection until a test
// releases it, so the test can cancel while a response is genuinely in flight.
//
// It reports two things a test cannot otherwise see from the client side: that
// delivery has begun, and that the server closed the connection instead of
// finishing it.
type heldResponse struct {
	// begun receives once, when the server first writes to the connection.
	begun chan struct{}

	// dropped is closed if the server closes the connection.
	dropped chan struct{}

	release chan struct{}

	markDropped func()
	resumeOnce  func()
}

func newHeldResponse() *heldResponse {
	held := &heldResponse{
		begun:   make(chan struct{}, 1),
		dropped: make(chan struct{}),
		release: make(chan struct{}),
	}

	// Both are one-shot because the number of connections is not the test's to
	// control: a stray dial would otherwise close a closed channel and panic
	// the suite rather than fail a test.
	held.markDropped = sync.OnceFunc(func() { close(held.dropped) })
	held.resumeOnce = sync.OnceFunc(func() { close(held.release) })

	return held
}

// listener wraps ln so every connection it accepts is held.
func (h *heldResponse) listener(ln net.Listener) net.Listener {
	return &heldResponseListener{Listener: ln, held: h}
}

// resume lets the paused write through. It is idempotent, so a test can defer
// it against its own early exit and still call it at the point it means to:
// leaving it unreleased would strand the server's goroutine on a test that has
// already failed.
func (h *heldResponse) resume() { h.resumeOnce() }

type heldResponseListener struct {
	net.Listener

	held *heldResponse
}

func (l *heldResponseListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &heldResponseConn{Conn: conn, held: l.held}, nil
}

type heldResponseConn struct {
	net.Conn

	held *heldResponse
	once sync.Once
}

// Write pauses the first write that the server makes on this connection, which
// is the response leaving the handler.
//
// The signal is a non-blocking send, so a connection no test is watching cannot
// wedge the server's write loop on it; the pause that follows is deliberate and
// is always released, by the deferred resume if not before. net/http may write
// from more than one goroutine over a connection's life, so the pause is
// one-shot.
func (c *heldResponseConn) Write(p []byte) (int, error) {
	c.once.Do(func() {
		select {
		case c.held.begun <- struct{}{}:
		default:
		}

		<-c.held.release
	})

	return c.Conn.Write(p)
}

func (c *heldResponseConn) Close() error {
	c.held.markDropped()

	return c.Conn.Close()
}
