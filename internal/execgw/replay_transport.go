package execgw

// replay_transport.go puts the stored bytes on the wire
// (extend-execution-contract task 2.2).
//
// # Why this is not internal/official
//
// Every method on the official client *builds* a body from structured fields.
// That is exactly what a replay must not do: a change in the serialisation rules
// would send different bytes under the same idempotency key, and the broker's
// answer to that is `422 idempotency-key-conflict` (openapi), which proves
// nothing about the original order. So the replay path has a transport of its
// own whose only input is bytes it cannot have produced.
//
// # What is deliberately missing
//
//   - No retry. The replay is counted in the journal *before* it is sent, so a
//     transport-level retry would send an uncounted request. The 401-refresh
//     retry the official client performs is right for a read and wrong here.
//   - No body construction, no field access, no key. HTTPReplay cannot name the
//     symbol, the side or the quantity of what it is sending.
//   - No status interpretation. It reports the status and the broker's error
//     code; the rules that turn those into an outcome live in replay.go, where
//     they can be read next to the spec that fixes them.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// replayPlacePath is the order-creation endpoint, the only one that accepts an
// idempotency key (openapi: clientOrderId is on OrderCreateRequest and on the
// conditional-order request, nowhere else — and this build has no conditional
// orders).
const replayPlacePath = "/api/v1/orders"

// maxReplayResponseBytes bounds what is read back. A replay answer is a small
// JSON object; anything larger is a proxy error page, and reading it in full
// helps nobody.
const maxReplayResponseBytes = 1 << 20

// HTTPReplay resends a stored body to the order-creation endpoint.
type HTTPReplay struct {
	// BaseURL is the API root, e.g. https://openapi.tossinvest.com.
	BaseURL string
	// HTTP performs the request. Required — there is no default, because a
	// replay with no timeout is a replay that can outlive the key it is using.
	HTTP *http.Client
	// Headers supplies the per-request headers: Authorization and
	// X-Tossinvest-Account. It is a function so the caller's token manager
	// stays the caller's, and so a token refresh is never something this file
	// decides to do in the middle of a replay.
	Headers func(ctx context.Context) (map[string]string, error)
}

// ReplayPlace sends the stored bytes and reports what came back.
//
// A returned error means "no usable answer" — the caller treats that as a lost
// response, which is the same thing it treated the original outcome as. A
// non-2xx status is *not* an error: it is an answer, and the rules in replay.go
// need to see it.
func (h HTTPReplay) ReplayPlace(ctx context.Context, body ReplayBody) (ReplayResponse, error) {
	if body.Empty() {
		return ReplayResponse{}, errors.New("execgw: there is no stored body to replay")
	}
	if h.HTTP == nil {
		return ReplayResponse{}, errors.New("execgw: the replay transport has no http client")
	}
	if strings.TrimSpace(h.BaseURL) == "" {
		return ReplayResponse{}, errors.New("execgw: the replay transport has no base url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(h.BaseURL, "/")+replayPlacePath, bytes.NewReader(body.Bytes()))
	if err != nil {
		return ReplayResponse{}, fmt.Errorf("execgw: building the replay request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if h.Headers != nil {
		headers, err := h.Headers(ctx)
		if err != nil {
			return ReplayResponse{}, fmt.Errorf("execgw: resolving the replay headers: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := h.HTTP.Do(req)
	if err != nil {
		return ReplayResponse{}, fmt.Errorf("execgw: the replay request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayResponseBytes))
	if err != nil {
		// The status arrived but the body did not. Reporting the status alone
		// would let a 2xx with an unread body look like "no identifier", which is
		// inconclusive anyway — but saying so explicitly is more honest.
		return ReplayResponse{Status: resp.StatusCode},
			fmt.Errorf("execgw: reading the replay response (HTTP %d): %w", resp.StatusCode, err)
	}
	return parseReplayResponse(resp.StatusCode, raw), nil
}

// replayEnvelope is the documented response shape on both sides: `{"result": …}`
// for a success and `{"error": {"code", "message"}}` for a refusal (openapi).
type replayEnvelope struct {
	Result struct {
		OrderID string `json:"orderId"`
		// A pointer distinguishes the documented null ("미전달 시 null") from an
		// echo of a different key, which is a response about a different order.
		ClientOrderID *string `json:"clientOrderId"`
	} `json:"result"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// parseReplayResponse reads the envelope without interpreting it.
//
// Identifiers come out verbatim. `orderId` and `clientOrderId` are opaque
// tokens — openapi contracts no shape for either — so trimming or case-folding
// on the way in would produce a value the broker never sent, and the byte-exact
// comparisons that follow would silently stop matching.
func parseReplayResponse(status int, raw []byte) ReplayResponse {
	out := ReplayResponse{Status: status}
	var env replayEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		// An unparseable body is not evidence of anything. The status still is.
		return out
	}
	out.OrderID = env.Result.OrderID
	if env.Result.ClientOrderID != nil {
		out.ClientOrderID = *env.Result.ClientOrderID
	}
	out.ErrorCode = strings.TrimSpace(env.Error.Code)
	out.ErrorMessage = strings.TrimSpace(env.Error.Message)
	return out
}

// Compile-time proof that the HTTP transport is a ReplayTransport, and that a
// ReplayTransport is all the gateway asks for.
var _ ReplayTransport = HTTPReplay{}

// budgetedHTTPReplay는 검증된 인증 소유자와 단일 POST를 같은 예산에 결속함.
type budgetedHTTPReplay struct {
	client *official.Client
	budget *accountbudget.Budget
	replay HTTPReplay
}

// NewBudgetedHTTPReplay는 기존 공개 HTTPReplay의 OFF 동작을 변경하지 않음.
func NewBudgetedHTTPReplay(baseURL string, client *http.Client, auth *official.Client, budget *accountbudget.Budget, cleanup func(error)) (ReplayTransport, error) {
	if client == nil || client.Timeout <= 0 || auth == nil || budget == nil || cleanup == nil || !auth.UsesAccountBudget(budget) || baseURL != auth.BaseURL() {
		return nil, accountbudget.ErrUnavailable
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	template, ok := base.(*http.Transport)
	if !ok || template == nil {
		return nil, accountbudget.ErrUnavailable
	}
	guard, err := accountbudget.NewTransport(budget, template, cleanup)
	if err != nil {
		return nil, err
	}
	copyClient := *client
	copyClient.Transport = guard
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &budgetedHTTPReplay{client: auth, budget: budget, replay: HTTPReplay{BaseURL: baseURL, HTTP: &copyClient}}, nil
}

func (r *budgetedHTTPReplay) ReplayPlace(ctx context.Context, body ReplayBody) (ReplayResponse, error) {
	if r == nil || r.client == nil || r.budget == nil || !r.client.UsesAccountBudget(r.budget) {
		return ReplayResponse{}, errors.Join(accountbudget.ErrUnavailable, accountbudget.ErrNotSent)
	}
	copyReplay := r.replay
	copyReplay.Headers = func(ctx context.Context) (map[string]string, error) {
		if !r.client.UsesAccountBudget(r.budget) {
			return nil, errors.Join(accountbudget.ErrUnavailable, accountbudget.ErrNotSent)
		}
		headers, err := r.client.AuthHeaders(ctx)
		if err != nil {
			return nil, err
		}
		if !r.client.UsesAccountBudget(r.budget) {
			return nil, errors.Join(accountbudget.ErrUnavailable, accountbudget.ErrNotSent)
		}
		return headers, nil
	}
	return copyReplay.ReplayPlace(ctx, body)
}
