package obs

// ntfy.go is the concrete notification transport (design D5: "알림은 ntfy 구체
// 구현(추상화 없음)").
//
// # Why ntfy and why no abstraction layer
//
// The design decision was to pick one and implement it rather than build a
// provider interface with one implementation. A notification abstraction that has
// never had a second provider is an abstraction whose shape is an accident, and
// its cost is paid on every read of the alerting path — which is the path you
// read while something is going wrong.
//
// The one seam that does exist is Publisher, and it exists because the *outbox*
// needs to be testable against a transport that fails on demand, not because a
// second provider is planned.
//
// # Self-hosted or ntfy.sh
//
// BaseURL defaults to https://ntfy.sh but is meant to be pointed at a self-hosted
// instance: the topic name is the only access control ntfy.sh has, and an engine
// publishing account events to a guessable public topic is an information leak.
// The engine warns about that at construction rather than refusing, because the
// operator may well be running a private ntfy.sh topic deliberately.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultNtfyBaseURL is the public ntfy service.
const DefaultNtfyBaseURL = "https://ntfy.sh"

// Publisher sends one notification.
//
// It exists so the outbox can be driven against a transport that fails when the
// test says so. Ntfy is the only production implementation.
type Publisher interface {
	Publish(ctx context.Context, n Notification) error
}

// Notification is one message.
type Notification struct {
	Type     EventType
	Severity Severity
	Title    string
	Body     string
	// Tags are ntfy's emoji/label tags.
	Tags []string
	// Priority is ntfy's 1-5 scale. Zero uses the severity's default.
	Priority int
}

// Ntfy publishes to an ntfy topic.
type Ntfy struct {
	// BaseURL is the ntfy server. Empty uses DefaultNtfyBaseURL.
	BaseURL string
	// Topic is the topic to publish to. Required.
	Topic string
	// Token is an optional bearer token for a protected topic.
	Token string
	// HTTPClient is the transport. Tests inject an httptest client. Nil uses a
	// client with Timeout.
	HTTPClient *http.Client
	// Timeout bounds one publish. Zero uses 10s.
	Timeout time.Duration
}

// ErrNtfyNotConfigured means no topic was set.
var ErrNtfyNotConfigured = errors.New("obs: ntfy has no topic configured")

// Publish sends one notification.
//
// A non-2xx response is an error, and so is a transport failure. The caller —
// the outbox for a critical alert, a best-effort send for an ordinary one —
// decides what that means; this function's only job is to be honest about
// whether the message left.
func (n *Ntfy) Publish(ctx context.Context, msg Notification) error {
	topic := strings.TrimSpace(n.Topic)
	if topic == "" {
		return ErrNtfyNotConfigured
	}
	base := strings.TrimRight(strings.TrimSpace(n.BaseURL), "/")
	if base == "" {
		base = DefaultNtfyBaseURL
	}

	timeout := n.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base+"/"+topic, bytes.NewReader([]byte(msg.Body)))
	if err != nil {
		return fmt.Errorf("obs: building the ntfy request: %w", err)
	}

	// ntfy takes the message metadata as headers so the body stays the human
	// text. Titles are ASCII-folded by the server; nothing here depends on that.
	if title := strings.TrimSpace(msg.Title); title != "" {
		req.Header.Set("Title", title)
	}
	tags := append([]string{string(msg.Type)}, msg.Tags...)
	req.Header.Set("Tags", strings.Join(tags, ","))
	req.Header.Set("Priority", strconv.Itoa(priorityFor(msg)))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if token := strings.TrimSpace(n.Token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := n.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("obs: publishing to ntfy: %w", err)
	}
	defer resp.Body.Close()
	// Draining matters for connection reuse, and the body is the server's error
	// text when there is one.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("obs: ntfy refused the message: %s: %s",
			resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// priorityFor maps a severity onto ntfy's 1-5 scale unless the caller pinned one.
func priorityFor(msg Notification) int {
	if msg.Priority > 0 {
		return msg.Priority
	}
	if msg.Severity == SeverityCritical {
		// 5 = "max": bypasses the client's do-not-disturb. A critical event here
		// means a live account is in a state nobody has accounted for.
		return 5
	}
	return 3
}

// UsesPublicService reports whether this publisher targets ntfy.sh, where the
// topic name is the only access control. Wiring code warns on it.
func (n *Ntfy) UsesPublicService() bool {
	base := strings.TrimRight(strings.TrimSpace(n.BaseURL), "/")
	return base == "" || base == DefaultNtfyBaseURL
}
