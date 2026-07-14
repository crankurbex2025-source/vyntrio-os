package netpresence

import "context"

// InterfacesReader returns minimal interface metadata for eligibility checks.
type InterfacesReader interface {
	Interfaces() ([]InterfaceFlags, error)
}

// CollectorDeps configures injectable interface readers for tests.
type CollectorDeps struct {
	Reader InterfacesReader
}

// Collector assembles read-only network presence for the overview DTO.
type Collector struct {
	reader InterfacesReader
}

// NewCollector creates a network presence collector.
func NewCollector(deps CollectorDeps) Collector {
	reader := deps.Reader
	if reader == nil {
		reader = defaultInterfacesReader()
	}
	return Collector{reader: reader}
}

// Collect returns network presence with unavailable degradation on failure.
func (c Collector) Collect(_ context.Context) Network {
	interfaces, err := c.reader.Interfaces()
	if err != nil {
		return Unavailable()
	}
	return Classify(interfaces)
}
