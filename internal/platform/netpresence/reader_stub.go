//go:build !linux

package netpresence

import "os"

type stubInterfacesReader struct{}

func defaultInterfacesReader() InterfacesReader {
	return stubInterfacesReader{}
}

func (stubInterfacesReader) Interfaces() ([]InterfaceFlags, error) {
	return nil, os.ErrInvalid
}
