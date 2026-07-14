package hostmetrics

import "fmt"

// DeriveUsedBytes returns total minus available with validation.
func DeriveUsedBytes(totalBytes, availableBytes uint64) (uint64, error) {
	if totalBytes == 0 {
		return 0, fmt.Errorf("used bytes: invalid total")
	}
	if availableBytes > totalBytes {
		return 0, fmt.Errorf("used bytes: available exceeds total")
	}
	return totalBytes - availableBytes, nil
}
