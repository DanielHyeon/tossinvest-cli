//go:build !unix

package ratebudget

func tryLock(string) (lockHandle, error) { return nil, ErrUnsupported }
