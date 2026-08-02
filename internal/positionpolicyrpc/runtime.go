package positionpolicyrpc

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

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

const (
	RuntimeControlDirectoryName = ".position-policy-runtime"
	RuntimeDescriptorFileName   = "endpoint.json"
	RuntimeSocketFileName       = "runtime.sock"
)

// RuntimeDescriptor grants access only to the separate runtime-read socket.
// Socket is a fixed basename rather than an absolute engine-namespace path, so
// another container may resolve it through its own mount of the shared engine
// directory without accepting descriptor path traversal.
type RuntimeDescriptor struct {
	Socket string `json:"socket"`
	Token  string `json:"token"`
	PID    int    `json:"pid"`
}

func RuntimeControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), RuntimeControlDirectoryName)
}

func RuntimeDescriptorPath(engineDir string) string {
	return filepath.Join(RuntimeControlDirectory(engineDir), RuntimeDescriptorFileName)
}

func RuntimeSocketPath(engineDir string) string {
	return filepath.Join(RuntimeControlDirectory(engineDir), RuntimeSocketFileName)
}

func ValidateRuntimeControlDirectory(path string) error {
	clean := strings.TrimSpace(path)
	if filepath.Base(clean) != RuntimeControlDirectoryName {
		return errors.New("position policy runtime: unexpected control directory name")
	}
	if err := ValidateEngineDirectory(filepath.Dir(clean)); err != nil {
		return err
	}
	return validatePrivateDirectory(clean, true)
}

func ValidateRuntimePublishedDescriptor(path string) error {
	if filepath.Base(path) != RuntimeDescriptorFileName {
		return errors.New("position policy runtime: unexpected descriptor filename")
	}
	if err := ValidateRuntimeControlDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	return ValidatePrivateFile(path)
}

func openRuntimeDescriptor(path string) (*os.File, error) {
	if filepath.Base(path) != RuntimeDescriptorFileName {
		return nil, errors.New("position policy runtime: unexpected descriptor filename")
	}
	if err := ValidateRuntimeControlDirectory(filepath.Dir(path)); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err := validateDescriptorInfo(before); err != nil {
		return nil, err
	}
	f, err := openDescriptorNoFollow(path)
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
		return nil, errors.New("position policy runtime: descriptor changed while it was opened")
	}
	if err := validateDescriptorInfo(opened); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func decodeRuntimeDescriptor(r io.Reader) (RuntimeDescriptor, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxDescriptorBytes+1))
	if err != nil || len(body) > maxDescriptorBytes {
		return RuntimeDescriptor{}, errors.New("position policy runtime: descriptor is unreadable or oversized")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var descriptor RuntimeDescriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return RuntimeDescriptor{}, errors.New("position policy runtime: descriptor JSON is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return RuntimeDescriptor{}, errors.New("position policy runtime: descriptor must contain one JSON value")
	}
	return descriptor, nil
}

type RuntimeClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *RuntimeClient) Runtime(ctx context.Context) (positionpolicy.ManagementRuntime, error) {
	var result positionpolicy.ManagementRuntime
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/runtime", nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.http.Do(req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote rpcError
		if err := json.NewDecoder(limited).Decode(&remote); err != nil {
			return result, fmt.Errorf("position policy runtime: HTTP %d", response.StatusCode)
		}
		return result, fmt.Errorf("position policy runtime: %s: %s", remote.Code,
			strings.TrimSpace(remote.Message))
	}
	if err := json.NewDecoder(limited).Decode(&result); err != nil {
		return result, fmt.Errorf("position policy runtime: decoding response: %w", err)
	}
	return result, nil
}

var _ positionpolicy.RuntimeReader = (*RuntimeClient)(nil)
