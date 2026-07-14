package netpresence

// InterfaceFlags captures the minimum metadata required for eligibility checks.
type InterfaceFlags struct {
	Loopback        bool
	Up              bool
	HasHardwareAddr bool
}

// IsEligible reports whether an interface satisfies the overview presence rule.
func IsEligible(iface InterfaceFlags) bool {
	return !iface.Loopback && iface.Up && iface.HasHardwareAddr
}

// Classify maps enumerated interface flags to the safe API network status.
func Classify(interfaces []InterfaceFlags) Network {
	for _, iface := range interfaces {
		if IsEligible(iface) {
			return Network{Status: StatusAvailable}
		}
	}
	return Network{Status: StatusUnknown}
}
