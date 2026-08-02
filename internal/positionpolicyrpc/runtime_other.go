//go:build !unix

package positionpolicyrpc

import (
	"context"
	"errors"
)

func DialRuntime(context.Context, string) (*RuntimeClient, error) {
	return nil, errors.New("position policy runtime: Unix sockets are unsupported on this platform")
}

func PrepareRuntimeSocket(string) error {
	return errors.New("position policy runtime: Unix sockets are unsupported on this platform")
}

func ValidateRuntimeSocket(string) error {
	return errors.New("position policy runtime: Unix sockets are unsupported on this platform")
}
