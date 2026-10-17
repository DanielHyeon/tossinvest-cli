package positionpolicyrpc

// attempt_thaw_client.go 는 a094 park 해동 route 의 tossctl 쪽임 — 같은 Dial 된 Client(같은 endpoint · 토큰 · descriptor
// 검증). 경로 문자열은 engine 패키지를 import 하지 않기 위해 여기 다시 적고, engine 쪽 시험이 두 값이 같음을 고정함.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
)

// AttemptThawPath 는 해동 route 임.
const AttemptThawPath = "/v1/attempt-thaw/resolve"

// ResolveParkedAttempt 는 엔진에 park 된 attempt 하나의 운영자 해소를 요청함.
func (c *Client) ResolveParkedAttempt(ctx context.Context, input attemptthaw.Request) (attemptthaw.Result, error) {
	var result attemptthaw.Result
	encoded, err := json.Marshal(input)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+AttemptThawPath, bytes.NewReader(encoded))
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote rpcError
		if err := json.NewDecoder(limited).Decode(&remote); err != nil {
			if response.StatusCode == http.StatusNotFound {
				return result, attemptthaw.ErrUnwired
			}
			return result, fmt.Errorf("attempt thaw control: HTTP %d", response.StatusCode)
		}
		base := map[string]error{
			"invalid":           attemptthaw.ErrInvalidRequest,
			"stale":             attemptthaw.ErrStale,
			"not_found":         attemptthaw.ErrNotFound,
			"audit_unavailable": attemptthaw.ErrAuditUnavailable,
			"unwired":           attemptthaw.ErrUnwired,
		}[remote.Code]
		if base == nil {
			base = errors.New("attempt thaw control: remote failure")
		}
		return result, fmt.Errorf("%w: %s", base, strings.TrimSpace(remote.Message))
	}
	if err := json.NewDecoder(limited).Decode(&result); err != nil {
		return result, fmt.Errorf("attempt thaw control: the resolution may have been recorded but its response could not be read: %w", err)
	}
	return result, nil
}
