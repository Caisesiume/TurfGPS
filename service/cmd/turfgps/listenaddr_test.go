package main

import (
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/Caisesiume/TurfGPS/service/internal/config"
)

// TestDefaultListenAddrIsLoopback binds #125: with nothing configured, the
// service listens on loopback and on nothing wider.
//
// It resolves the address through listenAddr, the call main makes, with an
// environment that sets nothing, and requires its host to be a loopback IP
// address. An empty host and a wildcard address both fail it, and they are the
// two ways the default could widen: `:8080`, the value this replaced, binds
// every interface on both address families.
//
// It parses rather than binds, deliberately. Binding the default would hold
// port 8080 on the host running the suite, so the test would fail on a machine
// already serving there and race any other test that starts the service.
func TestDefaultListenAddrIsLoopback(t *testing.T) {
	t.Parallel()

	addr, err := listenAddr(func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatalf("listenAddr with nothing configured: %v, want the default", err)
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("the default listen address %q is not host:port: %v", addr, err)
	}

	ip, err := netip.ParseAddr(host)
	if err != nil || !ip.IsLoopback() {
		t.Fatalf("with %s unset the service listens on %q, want a loopback address: the reverse proxy is the only component that may face the network, so an interface wider than loopback is one a deployment must open by setting %s, never one it inherits",
			config.EnvListenAddr, addr, config.EnvListenAddr)
	}
}

// TestStartupRefusesAMalformedListenAddr binds #125's refusal at the point main
// acts on it: a malformed address yields no address to bind, and the error
// names the setting, which main logs before exiting — before anything listens.
func TestStartupRefusesAMalformedListenAddr(t *testing.T) {
	t.Parallel()

	addr, err := listenAddr(func(name string) (string, bool) {
		if name == config.EnvListenAddr {
			return "localhost:8080", true
		}

		return "", false
	})
	if err == nil {
		t.Fatalf("listenAddr accepted %s=localhost:8080 as %q, want a refusal", config.EnvListenAddr, addr)
	}

	if !strings.Contains(err.Error(), config.EnvListenAddr) {
		t.Errorf("the refusal %q does not name %s", err, config.EnvListenAddr)
	}
}
