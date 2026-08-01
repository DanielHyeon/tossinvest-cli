// Package positionpolicyrpc is the read/write-neutral client protocol for the
// engine-owned position-policy control plane. It imports neither journal nor
// engine; possession of this client can only reach the fixed authenticated
// loopback endpoint published by a running engine.
package positionpolicyrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

const (
	ControlDirectoryName = ".position-policy-control"
	DescriptorFileName   = "endpoint.json"
	maxDescriptorBytes   = 4096
)

type Descriptor struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	PID     int    `json:"pid"`
}

func DescriptorPath(engineDir string) string {
	return filepath.Join(ControlDirectory(engineDir), DescriptorFileName)
}

func ControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), ControlDirectoryName)
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func Dial(ctx context.Context, descriptorPath string) (*Client, error) {
	path := strings.TrimSpace(descriptorPath)
	if path == "" {
		return nil, errors.New("position policy control: descriptor path is empty")
	}
	f, err := openPrivateDescriptor(path, openDescriptorNoFollow)
	if err != nil {
		return nil, fmt.Errorf("position policy control: opening descriptor: %w", err)
	}
	defer f.Close()
	descriptor, err := decodeDescriptor(f)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(descriptor.Address)
	if err != nil || descriptor.PID <= 0 || len(strings.TrimSpace(descriptor.Token)) < 32 {
		return nil, errors.New("position policy control: descriptor fields are invalid")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return nil, errors.New("position policy control: endpoint is not loopback")
	}
	transport := &http.Transport{Proxy: nil}
	client := &Client{
		baseURL: "http://" + descriptor.Address, token: descriptor.Token,
		http: &http.Client{Transport: transport, Timeout: 5 * time.Second},
	}
	var health struct {
		OK bool `json:"ok"`
	}
	if err := client.call(ctx, http.MethodGet, "/v1/health", nil, &health); err != nil || !health.OK {
		if err == nil {
			err = errors.New("health response was not OK")
		}
		return nil, fmt.Errorf("position policy control: engine endpoint unavailable: %w", err)
	}
	return client, nil
}

// ValidateEngineDirectory verifies the existing parent in which the dedicated
// private control directory is created. Read/execute bits may be broader for
// compatibility with an existing engine directory, but no other principal may
// write it and no symlink traversal is accepted.
func ValidateEngineDirectory(path string) error {
	return validatePrivateDirectory(path, false)
}

// ValidateControlDirectory verifies the dedicated endpoint directory and its
// engine parent. The leaf must be exactly 0700.
func ValidateControlDirectory(path string) error {
	clean := strings.TrimSpace(path)
	if filepath.Base(clean) != ControlDirectoryName {
		return errors.New("position policy control: unexpected control directory name")
	}
	if err := ValidateEngineDirectory(filepath.Dir(clean)); err != nil {
		return err
	}
	return validatePrivateDirectory(clean, true)
}

// ValidatePrivateOpenFile verifies an already-open descriptor or staging file.
func ValidatePrivateOpenFile(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	return validateDescriptorInfo(info)
}

// ValidatePrivateFile verifies a staged descriptor without requiring its
// temporary filename to match the published endpoint filename.
func ValidatePrivateFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	return validateDescriptorInfo(info)
}

// ValidatePublishedDescriptor verifies a path after staging/rename.
func ValidatePublishedDescriptor(path string) error {
	if filepath.Base(path) != DescriptorFileName {
		return errors.New("position policy control: unexpected descriptor filename")
	}
	if err := ValidateControlDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	return ValidatePrivateFile(path)
}

func validatePrivateDirectory(path string, exact bool) error {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return errors.New("position policy control: private directory path is empty")
	}
	if err := rejectSymlinkTraversal(clean); err != nil {
		return err
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("position policy control: private directory is not a real directory")
	}
	if err := validateOwnerAndLinks(info, false); err != nil {
		return err
	}
	permissions := info.Mode().Perm()
	if exact && permissions != 0o700 {
		return fmt.Errorf("position policy control: control directory mode is %04o, want 0700", permissions)
	}
	if !exact && permissions&0o022 != 0 {
		return errors.New("position policy control: engine directory is writable by group or other")
	}
	return nil
}

func validateDescriptorInfo(info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return errors.New("position policy control: descriptor is not an exact 0600 regular file")
	}
	return validateOwnerAndLinks(info, true)
}

func openPrivateDescriptor(path string, opener func(string) (*os.File, error)) (*os.File, error) {
	if filepath.Base(path) != DescriptorFileName {
		return nil, errors.New("position policy control: unexpected descriptor filename")
	}
	if err := ValidateControlDirectory(filepath.Dir(path)); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err := validateDescriptorInfo(before); err != nil {
		return nil, err
	}
	f, err := opener(path)
	if err != nil {
		return nil, err
	}
	opened, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	current, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, opened) || !os.SameFile(opened, current) {
		f.Close()
		return nil, errors.New("position policy control: descriptor changed while it was opened")
	}
	if err := validateDescriptorInfo(opened); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func decodeDescriptor(r io.Reader) (Descriptor, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxDescriptorBytes+1))
	if err != nil || len(body) > maxDescriptorBytes {
		return Descriptor{}, errors.New("position policy control: descriptor is unreadable or oversized")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var descriptor Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return Descriptor{}, errors.New("position policy control: descriptor JSON is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Descriptor{}, errors.New("position policy control: descriptor must contain one JSON value")
	}
	return descriptor, nil
}

func (c *Client) List(ctx context.Context) ([]positionpolicy.State, error) {
	var result []positionpolicy.State
	return result, c.call(ctx, http.MethodGet, "/v1/positions", nil, &result)
}

func (c *Client) Preview(ctx context.Context, req positionpolicy.Request) (positionpolicy.Preview, error) {
	var result positionpolicy.Preview
	return result, c.call(ctx, http.MethodPost, "/v1/preview", req, &result)
}

func (c *Client) Apply(ctx context.Context, req positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	var result positionpolicy.State
	return result, c.call(ctx, http.MethodPost, "/v1/apply", req, &result)
}

type rpcError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) call(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote rpcError
		if err := json.NewDecoder(limited).Decode(&remote); err != nil {
			return fmt.Errorf("position policy control: HTTP %d", response.StatusCode)
		}
		return decodeRemoteError(remote)
	}
	if err := json.NewDecoder(limited).Decode(output); err != nil {
		return fmt.Errorf("position policy control: decoding response: %w", err)
	}
	return nil
}

func decodeRemoteError(remote rpcError) error {
	base := map[string]error{
		"invalid": positionpolicy.ErrInvalidRequest, "not_found": positionpolicy.ErrPositionNotFound,
		"stale": positionpolicy.ErrVersionMismatch, "exit_conflict": positionpolicy.ErrExitConflict,
		"unknown_state": positionpolicy.ErrUnknownState, "ineligible": positionpolicy.ErrIneligible,
		"capability_invalid":    positionpolicy.ErrCapabilityInvalid,
		"capability_too_early":  positionpolicy.ErrCapabilityTooEarly,
		"capability_expired":    positionpolicy.ErrCapabilityExpired,
		"capability_stale":      positionpolicy.ErrCapabilityStale,
		"confirmation_required": positionpolicy.ErrConfirmationRequired,
	}[remote.Code]
	if base == nil {
		base = errors.New("position policy control: remote failure")
	}
	return fmt.Errorf("%w: %s", base, strings.TrimSpace(remote.Message))
}

var _ interface {
	List(context.Context) ([]positionpolicy.State, error)
	Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error)
	Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error)
} = (*Client)(nil)
