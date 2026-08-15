//go:build !linux

package verifylive

import "fmt"

// M0 is a Linux-supported measurement tool. Refusing unsupported platforms is
// safer than silently replacing dirfd/no-follow guarantees with pathname checks.
func openCausalReceipt(_ string, _ receiptHooks) (*CausalReceipt, error) {
	return nil, fmt.Errorf("verifylive: M0 causal receipts require Linux dirfd no-follow support")
}
