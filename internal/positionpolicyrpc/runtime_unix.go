//go:build unix

package positionpolicyrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DialRuntime(ctx context.Context, descriptorPath string) (*RuntimeClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := strings.TrimSpace(descriptorPath)
	if path == "" {
		return nil, errors.New("position policy runtime: descriptor path is empty")
	}
	f, err := openRuntimeDescriptor(path)
	if err != nil {
		return nil, fmt.Errorf("position policy runtime: opening descriptor: %w", err)
	}
	defer f.Close()
	descriptor, err := decodeRuntimeDescriptor(f)
	if err != nil {
		return nil, err
	}
	if descriptor.Socket != RuntimeSocketFileName || descriptor.PID <= 0 ||
		len(strings.TrimSpace(descriptor.Token)) < 32 {
		return nil, errors.New("position policy runtime: descriptor fields are invalid")
	}
	socketPath := filepath.Join(filepath.Dir(path), descriptor.Socket)
	if err := ValidateRuntimeSocket(socketPath); err != nil {
		return nil, fmt.Errorf("position policy runtime: validating socket: %w", err)
	}
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}
	return &RuntimeClient{
		baseURL: "http://runtime", token: descriptor.Token,
		http: &http.Client{Transport: transport, Timeout: 5 * time.Second},
	}, nil
}

func PrepareRuntimeSocket(path string) error {
	if err := validateRuntimeSocketPath(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSocket == 0 || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return errors.New("position policy runtime: stale endpoint is not an exact 0600 Unix socket")
	}
	if err := validateOwnerAndLinks(info, false); err != nil {
		return err
	}
	return os.Remove(path)
}

func ValidateRuntimeSocket(path string) error {
	if err := validateRuntimeSocketPath(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSocket == 0 || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return errors.New("position policy runtime: endpoint is not an exact 0600 Unix socket")
	}
	return validateOwnerAndLinks(info, false)
}

func validateRuntimeSocketPath(path string) error {
	clean := strings.TrimSpace(path)
	if filepath.Base(clean) != RuntimeSocketFileName {
		return errors.New("position policy runtime: unexpected socket filename")
	}
	return ValidateRuntimeControlDirectory(filepath.Dir(clean))
}
