package config

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
)

// EnvListenAddr names the address the service's HTTP listener binds, as
// host:port (#125).
//
// THE DEFAULT IS LOOPBACK, AND IT IS NOT THIS PACKAGE'S. It is `defaultAddr` in
// the composition root, `cmd/turfgps`, beside the server's other budgets, and a
// test there fails if it widens. `DEPLOYMENT.md § What is deployed` puts the
// reverse proxy in front of the service over the loopback interface and makes
// it the only component that may face the network, so a service that bound
// every interface by omission would open the boundary that design rests on.
//
// Setting this variable is how a deployment says otherwise, and it is the only
// way: an empty host (`:8080`) or a wildcard address (`0.0.0.0:8080`, `[::]:8080`)
// binds every interface, and an operator writes one of them on purpose. The
// image is the one place in this repository that does, and `service/Dockerfile`
// says why.
const EnvListenAddr = "TURFGPS_LISTEN_ADDR"

// LoadListenAddr reads the address the HTTP listener binds, and returns
// fallback when the variable is unset or empty.
//
// A value it cannot hold to the form below is a refusal naming the variable,
// never a value quietly corrected, for the reason every value in this package
// is refused rather than repaired: an address rewritten here is an exposure
// boundary the operator believes they set.
//
// The host must be an IP address or empty. A NAME IS REFUSED, `localhost`
// included, because a name is resolved when the service starts: the interface
// it binds would then be decided by the resolver and the hosts file rather than
// by this setting, and Go binds only the first address a name resolves to. The
// port must be a decimal number; a service name is refused for the same reason.
// Port 0 is accepted — it asks the kernel for a free port, which the
// `listening` line reports — because it is well formed and a harness has a use
// for it, though a deployment has none.
func LoadListenAddr(lookup LookupFunc, fallback string) (string, error) {
	raw, ok := lookup(EnvListenAddr)
	if !ok || raw == "" {
		return fallback, nil
	}

	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		return "", fmt.Errorf("%s=%q is not host:port: %w", EnvListenAddr, raw, err)
	}

	if host != "" {
		if _, err := netip.ParseAddr(host); err != nil {
			return "", fmt.Errorf("%s=%q names the host %q, want an IP address: a name is resolved at start, so the resolver rather than this setting would decide the interface — write 127.0.0.1 or ::1 for loopback, or leave the host empty to listen on every interface", EnvListenAddr, raw, host)
		}
	}

	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return "", fmt.Errorf("%s=%q has the port %q, want a decimal number from 0 to 65535", EnvListenAddr, raw, port)
	}

	return raw, nil
}
