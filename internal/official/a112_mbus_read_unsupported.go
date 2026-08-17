//go:build !unix

package official

import (
	"fmt"
	"time"
)

func a112MBUSCachedToken(_ string, _ time.Time) (string, error) {
	return "", fmt.Errorf("A112 M-B0 cached-token measurement is unsupported on this platform")
}
