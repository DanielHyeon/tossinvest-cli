//go:build !linux

package clock

type unsupportedBootSource struct{}

func (unsupportedBootSource) Snapshot() (BootStamp, error) { return BootStamp{}, ErrBootStamp }
func SystemBootSource() BootSource                         { return unsupportedBootSource{} }
