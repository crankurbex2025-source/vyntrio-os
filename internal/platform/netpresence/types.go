package netpresence

// Network is the safe network-presence section for GET /api/v1/overview.
type Network struct {
	Status string `json:"status"`
}

// Unavailable returns the API shape when presence cannot be determined.
func Unavailable() Network {
	return Network{Status: StatusUnavailable}
}
