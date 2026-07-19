package storageinventory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func stableDeviceID(kernelName string) (string, bool) {
	if kernelName == "" {
		return "", false
	}
	sum := sha256.Sum256([]byte(kernelName))
	return fmt.Sprintf("disk-%s", hex.EncodeToString(sum[:6])), true
}
