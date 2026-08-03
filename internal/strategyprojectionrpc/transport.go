// Package strategyprojectionrpc exposes only the bounded authenticated read
// projection over a private runtime Unix socket.
package strategyprojectionrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

const (
	ControlDirectoryName = ".strategy-runtime-read"
	DescriptorFileName   = "endpoint.json"
	SocketFileName       = "runtime.sock"
	ProjectionPath       = "/v1/strategy-runtime"
	MaxProjectionBytes   = 1 << 20
	maxDescriptorBytes   = 4096
	descriptorSchema     = "tossos.strategy-runtime-unix/v1"
)

type Descriptor struct {
	SchemaVersion string `json:"schemaVersion"`
	Socket        string `json:"socket"`
	Token         string `json:"token"`
	PID           int    `json:"pid"`
}

func ControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), ControlDirectoryName)
}

func DescriptorPath(engineDir string) string {
	return filepath.Join(ControlDirectory(engineDir), DescriptorFileName)
}

func SocketPath(engineDir string) string {
	return filepath.Join(ControlDirectory(engineDir), SocketFileName)
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *Client) Read(ctx context.Context) (strategyprojection.Snapshot, error) {
	var result strategyprojection.Snapshot
	if c == nil || c.http == nil || len(c.token) < 32 {
		return result, errors.New("strategy projection runtime: invalid client")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+ProjectionPath, nil)
	if err != nil {
		return result, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.http.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxProjectionBytes+1))
	if err != nil || len(body) > MaxProjectionBytes {
		return result, errors.New("strategy projection runtime: response unreadable or oversized")
	}
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("strategy projection runtime: HTTP %d", response.StatusCode)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("strategy projection runtime: invalid response: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return result, errors.New("strategy projection runtime: response must contain one value")
	}
	if err := strategyprojection.Validate(result); err != nil {
		return strategyprojection.Snapshot{}, err
	}
	return strategyprojection.Clone(result), nil
}

func readDescriptor(path string) (Descriptor, error) {
	if filepath.Base(path) != DescriptorFileName || filepath.Base(filepath.Dir(path)) != ControlDirectoryName {
		return Descriptor{}, errors.New("strategy projection runtime: unexpected descriptor path")
	}
	directory, err := os.Lstat(filepath.Dir(path))
	if err != nil || !directory.IsDir() || directory.Mode()&os.ModeSymlink != 0 || directory.Mode().Perm() != 0o700 {
		return Descriptor{}, errors.New("strategy projection runtime: control directory is not exact 0700")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor is not exact 0600 regular file")
	}
	file, err := openDescriptorNoFollow(path)
	if err != nil {
		return Descriptor{}, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return Descriptor{}, err
	}
	current, currentErr := os.Lstat(path)
	if currentErr != nil || !os.SameFile(info, opened) || !os.SameFile(opened, current) ||
		!opened.Mode().IsRegular() || opened.Mode().Perm() != 0o600 {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor changed while it was opened")
	}
	body, err := io.ReadAll(io.LimitReader(file, maxDescriptorBytes+1))
	if err != nil || len(body) > maxDescriptorBytes {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor unreadable or oversized")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var descriptor Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor JSON invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor must contain one value")
	}
	if descriptor.SchemaVersion != descriptorSchema || descriptor.Socket != SocketFileName || len(descriptor.Token) < 32 || descriptor.PID <= 0 {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor fields invalid")
	}
	return descriptor, nil
}

func hasRequestBody(request *http.Request) bool {
	return request.ContentLength > 0 || len(request.TransferEncoding) > 0
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

var _ strategyprojection.Reader = (*Client)(nil)
