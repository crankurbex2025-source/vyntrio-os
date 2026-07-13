package peer

import (
	"net"
	"net/netip"
)

// IsLoopback reports whether remoteAddr is a direct TCP peer on loopback.
func IsLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if !addr.IsValid() || !addr.IsLoopback() {
		return false
	}

	return addr.Is4() || addr.Is6()
}
