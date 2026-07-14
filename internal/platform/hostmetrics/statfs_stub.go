//go:build !linux

package hostmetrics

import "fmt"

func statfsCapacity(_ string) (StatFSResult, error) {
	return StatFSResult{}, fmt.Errorf("statfs: unsupported platform")
}
