// Package official provides a client for the Toss Open API (official credentials).
package official

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Credentials holds Toss Open API credentials.
// All fields are plain strings; secrets must never be logged or printed.
type Credentials struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Label     string `json:"label,omitempty"`
	SavedAt   string `json:"savedAt,omitempty"`
}

const (
	envKey    = "TOSSCTL_OPENAPI_KEY"
	envSecret = "TOSSCTL_OPENAPI_SECRET"
)

// LoadCredentials returns credentials from environment variables (if both
// TOSSCTL_OPENAPI_KEY and TOSSCTL_OPENAPI_SECRET are non-empty) or from file.
// If neither source is available, it returns (nil, nil).
// If the file exists but is unreadable or malformed, it returns an error.
func LoadCredentials(getenv func(string) string, file string) (*Credentials, error) {
	k := getenv(envKey)
	s := getenv(envSecret)
	if k != "" && s != "" {
		return &Credentials{APIKey: k, SecretKey: s}, nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveCredentials writes c to file as JSON with 0600 permissions.
// The parent directory is created with 0700 permissions if it does not exist.
func SaveCredentials(file string, c Credentials) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".openapi-credentials-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			_ = tmp.Close()
		}
		_ = os.Remove(tmpName)
	}()
	if err := tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Rename(tmpName, file); err != nil {
		return err
	}
	info, err := os.Lstat(file)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return errors.New("credential store did not produce a regular 0600 file")
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer parent.Close()
	return parent.Sync()
}

// DeleteCredentials removes the credentials file.
// If the file does not exist, it returns nil.
func DeleteCredentials(file string) error {
	err := os.Remove(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// MaskedKey returns a display-safe version of the API key:
// first 10 characters + "…" + last 4 characters.
// If the key is fewer than 14 characters, it returns "…".
func (c Credentials) MaskedKey() string {
	key := c.APIKey
	if len(key) < 14 {
		return "…"
	}
	return key[:10] + "…" + key[len(key)-4:]
}
