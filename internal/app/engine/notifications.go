package engine

// notifications.go joins the notification settings an operator writes in the
// config file with the secret they must not (change a074, engine-safety "알림
// 전송 경로는 설정 가능하고 그 설정은 audit된다").
//
// # Why this seam exists at all
//
// internal/config parses a file and reads no environment variable, so its tests
// do not depend on the process they run in. That leaves the join — file ⊕
// environment — with no home, and this is it: the one place that knows both
// halves, and the place that produces the audit summary of what was configured
// without producing the values.
//
// # The failure direction
//
// A notification setting this function cannot make sense of does *not* refuse
// the engine. Alert delivery is not a protection path, and an engine that will
// not start because of a typo in a topic name is an engine whose stop evaluation
// has also stopped — strictly worse than an engine with no alerts. So the answer
// to a bad setting is a nil publisher, a recorded reason, and today's behaviour:
// the outbox keeps the critical alert, the entry gate latches, and sustained
// failure escalates the operating mode. That is already the specified direction
// (risk-management), and it is what the build did before this change.

import (
	"os"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// The two environment variables. The token has no config-file equivalent by
// design; the topic does, and overrides it, because on ntfy.sh the topic name is
// the only access control there is.
const (
	envNtfyToken = "TOSSCTL_NTFY_TOKEN"
	envNtfyTopic = "TOSSCTL_NTFY_TOPIC"
)

// notificationResolution is what the audit trail and the startup log are told.
//
// It carries no topic and no token. §0.8 forbids writing a secret to a log, and
// §0.5's question — "what was this engine configured with, and when did that
// change" — is answered by whether a channel and a credential were present. The
// base URL is kept because a host address is not a secret and "which server were
// we sending to" is a question incident review actually asks.
type notificationResolution struct {
	Enabled         bool
	BaseURL         string
	TopicConfigured bool
	TokenConfigured bool
	// PublicService reports that this targets ntfy.sh, where the topic name is
	// the only access control. Warned about, never refused: an operator may be
	// running a private ntfy.sh topic deliberately.
	PublicService bool
	// Refused explains why no transport was built despite the block being on.
	Refused string
}

// resolveNotificationPublisher builds the alert transport, or explains why it
// did not.
//
// getenv is injectable so the tests do not mutate the process environment; nil
// means os.Getenv.
func resolveNotificationPublisher(cfg config.Notifications,
	getenv func(string) string) (obs.Publisher, notificationResolution) {
	if getenv == nil {
		getenv = os.Getenv
	}
	resolution := notificationResolution{
		Enabled: cfg.Enabled,
		BaseURL: strings.TrimSpace(cfg.BaseURL),
		Refused: strings.TrimSpace(cfg.Rejected),
	}
	if resolution.Refused != "" {
		// internal/config already zeroed the block and said why. Carrying its
		// reason forward rather than inventing a second one keeps one explanation
		// per failure.
		return nil, resolution
	}
	if !cfg.Enabled {
		// Not an error, and not a candidate for inference: a topic left behind in
		// the file by an operator who turned notifications off does not turn them
		// back on (§0.7).
		return nil, resolution
	}

	topic := strings.TrimSpace(getenv(envNtfyTopic))
	if topic == "" {
		topic = strings.TrimSpace(cfg.Topic)
	}
	if topic == "" {
		resolution.Refused = "notifications are enabled but no topic is configured, in the file or in " +
			envNtfyTopic
		return nil, resolution
	}
	token := strings.TrimSpace(getenv(envNtfyToken))

	ntfy := &obs.Ntfy{BaseURL: resolution.BaseURL, Topic: topic, Token: token}
	resolution.TopicConfigured = true
	resolution.TokenConfigured = token != ""
	resolution.PublicService = ntfy.UsesPublicService()
	return ntfy, resolution
}
