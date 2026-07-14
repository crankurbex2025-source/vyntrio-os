//go:build linux

package netpresence

import "net"

type osInterfacesReader struct{}

func defaultInterfacesReader() InterfacesReader {
	return osInterfacesReader{}
}

func (osInterfacesReader) Interfaces() ([]InterfaceFlags, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	flags := make([]InterfaceFlags, 0, len(ifaces))
	for _, iface := range ifaces {
		flags = append(flags, InterfaceFlags{
			Loopback:        iface.Flags&net.FlagLoopback != 0,
			Up:              iface.Flags&net.FlagUp != 0,
			HasHardwareAddr: len(iface.HardwareAddr) > 0,
		})
	}
	return flags, nil
}
