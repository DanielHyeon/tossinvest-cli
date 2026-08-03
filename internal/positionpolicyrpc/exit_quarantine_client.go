package positionpolicyrpc

// exit_quarantine_client.go is the console side of change a063's three routes.
//
// It rides the same dialled Client — same socket, same bearer token, same
// descriptor validation — and carries its own request helper for one reason: the
// error vocabulary. Client.call translates remote codes into positionpolicy
// errors, and a release refused for "not_quarantined" has no honest positionpolicy
// equivalent. Reusing that translation would hand the console a
// position-not-found where the truth is "this generation is not quarantined".

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

func (c *Client) Quarantines(ctx context.Context) ([]exitquarantine.Row, error) {
	var result []exitquarantine.Row
	return result, c.callQuarantine(ctx, http.MethodGet, "/v1/quarantines", nil, &result)
}

func (c *Client) PreviewQuarantineRelease(ctx context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	var result exitquarantine.Preview
	return result, c.callQuarantine(ctx, http.MethodPost, "/v1/quarantine/preview", req, &result)
}

func (c *Client) ReleaseQuarantine(ctx context.Context,
	req exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	var result exitquarantine.Result
	return result, c.callQuarantine(ctx, http.MethodPost, "/v1/quarantine/release", req, &result)
}

func (c *Client) callQuarantine(ctx context.Context, method, path string, input, output any) error {
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
			// An engine that predates a063 has no such route, so its mux answers
			// with the catch-all rather than a coded error. Saying "unwired"
			// here is what lets the console draw a screen instead of a 500.
			if response.StatusCode == http.StatusNotFound {
				return exitquarantine.ErrUnwired
			}
			return fmt.Errorf("exit quarantine control: HTTP %d", response.StatusCode)
		}
		return decodeRemoteQuarantineError(remote)
	}
	if err := json.NewDecoder(limited).Decode(output); err != nil {
		return fmt.Errorf("exit quarantine control: decoding response: %w", err)
	}
	return nil
}

func decodeRemoteQuarantineError(remote rpcError) error {
	base := map[string]error{
		"invalid":               exitquarantine.ErrInvalidRequest,
		"not_quarantined":       exitquarantine.ErrNotQuarantined,
		"stale":                 exitquarantine.ErrVersionMismatch,
		"capability_invalid":    exitquarantine.ErrCapabilityInvalid,
		"capability_too_early":  exitquarantine.ErrCapabilityTooEarly,
		"capability_expired":    exitquarantine.ErrCapabilityExpired,
		"confirmation_required": exitquarantine.ErrConfirmationRequired,
		"unwired":               exitquarantine.ErrUnwired,
	}[remote.Code]
	if base == nil {
		base = errors.New("exit quarantine control: remote failure")
	}
	return fmt.Errorf("%w: %s", base, strings.TrimSpace(remote.Message))
}

var _ interface {
	Quarantines(context.Context) ([]exitquarantine.Row, error)
	PreviewQuarantineRelease(context.Context, exitquarantine.Request) (exitquarantine.Preview, error)
	ReleaseQuarantine(context.Context, exitquarantine.ApplyRequest) (exitquarantine.Result, error)
} = (*Client)(nil)
