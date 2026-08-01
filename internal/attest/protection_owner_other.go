//go:build !unix

package attest

// Protection artifact ownership cannot be established portably on this
// platform, so the verifier fails closed instead of weakening the check.
func currentProtectionOwnerUID() (uint32, bool) { return 0, false }
