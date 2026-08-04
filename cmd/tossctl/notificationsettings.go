package main

// notificationsettings.go is the console's alert-delivery seam (change a065).
//
// # Everything the screen cannot do lives here
//
// `internal/console` names no config service, no random source, no transport and
// no audit log — three static guards in that package say so, and a065 keeps all
// three true by giving the screen a seam whose Enable takes no argument. The
// values it writes are decided here:
//
//	channel   obs.NewTopic — 128 bits of crypto/rand, never operator input
//	server    obs.DefaultNtfyBaseURL, written explicitly so the file says where
//	          alerts go rather than relying on an empty-means-public default
//	token     not written anywhere. Read from the engine's own environment by
//	          resolveNotificationPublisher (a064), and used HERE only to make the
//	          console's test message reach a self-hosted instance behind one.
//
// # The audit line carries no secret
//
// a064 fixed the contract for what an alert-settings audit entry may hold: the
// server address, and whether a channel and a credential exist. Not their values.
// The channel identifier is the access control on a public ntfy topic, so writing
// it to an append-only operations log would be writing the credential to the log.

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// envNtfyToken is the same variable the engine's assembly reads (a064). It is
// spelled again rather than exported from internal/app/engine: the two readers
// want the same name for different processes, and exporting it would make one
// package's private wiring another's API.
const envNtfyToken = "TOSSCTL_NTFY_TOKEN"

// errNoNotificationChannel is the test button's answer when nothing is set up.
var errNoNotificationChannel = errors.New("알림 채널이 아직 없다 — [알림 켜기]를 먼저 누른다")

// newNotificationSeam resolves the same per-profile config file every other
// settings seam resolves to.
func newNotificationSeam(root *rootOptions) *consoleNotifications {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleNotifications{svc: svc}
}

type consoleNotifications struct {
	svc *config.Service
	// publish is the transport the test button uses. Nil builds an obs.Ntfy from
	// the saved block; tests inject one so no test message leaves the machine.
	publish func(context.Context, config.Notifications, string) error
	// getenv reads the token. Nil means os.Getenv, matching the engine's own seam.
	getenv func(string) string
}

func (s *consoleNotifications) Load() (config.Notifications, error) {
	return s.svc.LoadRawNotifications()
}

// Enable turns delivery on, creating a channel only when there is none.
//
// Reusing an existing channel is the whole reason this reads before it writes:
// an operator who turns alerts off for an afternoon and back on again must not
// have to re-subscribe every device. The engine's assembly already guarantees the
// channel sitting under `enabled: false` did nothing while it was off.
func (s *consoleNotifications) Enable() (config.Notifications, error) {
	current, err := s.svc.LoadRawNotifications()
	if err != nil {
		return config.Notifications{}, err
	}
	next := config.Notifications{
		Enabled: true,
		BaseURL: strings.TrimSpace(current.BaseURL),
		Topic:   strings.TrimSpace(current.Topic),
	}
	if next.BaseURL == "" {
		next.BaseURL = obs.DefaultNtfyBaseURL
	}
	created := next.Topic == ""
	if created {
		topic, tErr := obs.NewTopic()
		if tErr != nil {
			// No fallback to a weaker source. A channel that could not be made
			// random must not exist: the operator gets a failed button, not account
			// events on a guessable public topic.
			return config.Notifications{}, tErr
		}
		next.Topic = topic
	}
	if err := s.svc.SaveNotifications(next); err != nil {
		return config.Notifications{}, err
	}
	// Best-effort, exactly as recordGateFlip is: the save is durable and a console
	// edit must not be rolled back because an audit disk write failed.
	recordNotificationChange(current, next, created)
	return next, nil
}

// Disable writes the switch and keeps the channel.
func (s *consoleNotifications) Disable() error {
	current, cErr := s.svc.LoadRawNotifications()
	if err := s.svc.SaveNotificationsEnabled(false); err != nil {
		return err
	}
	if cErr != nil {
		current = config.Notifications{}
	}
	next := current
	next.Enabled = false
	recordNotificationChange(current, next, false)
	return nil
}

// Test sends one message to the configured channel.
//
// It does not touch the outbox, the entry gate or the operating mode. The message
// is an ordinary publish through the same transport the engine would use, which
// is what makes it evidence: if this arrives, the address and the network are
// right, and what is left to be uncertain about is the engine's own environment.
func (s *consoleNotifications) Test(ctx context.Context) error {
	block, err := s.svc.LoadRawNotifications()
	if err != nil {
		return err
	}
	getenv := s.getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	token := strings.TrimSpace(getenv(envNtfyToken))
	if s.publish != nil {
		return s.publish(ctx, block, token)
	}
	return publishNotificationTest(ctx, block, token)
}

// publishNotificationTest is the real send.
func publishNotificationTest(ctx context.Context, block config.Notifications, token string) error {
	topic := strings.TrimSpace(block.Topic)
	if topic == "" {
		return errNoNotificationChannel
	}
	ntfy := &obs.Ntfy{BaseURL: strings.TrimSpace(block.BaseURL), Topic: topic, Token: token}
	return ntfy.Publish(ctx, obs.Notification{
		Type: obs.EventType("console.notification_test"),
		// Normal, not critical. Priority 5 bypasses a phone's do-not-disturb, and
		// a test message that did that would teach the operator to distrust the
		// one grade that means a live account is unprotected.
		Severity: obs.SeverityNormal,
		Title:    "TossOS 알림 테스트",
		Body: "이 메시지가 보이면 critical 알림이 이 채널로 도착한다. " +
			"보호가 멈춘 포지션·전달 실패·운영 모드 강등이 같은 경로로 온다.",
	})
}

// recordNotificationChange appends one entry per key that moved.
//
// Per key rather than one line for the block, matching recordPolicySave: a reader
// asking "when did alerts start going somewhere" should find that sentence rather
// than a diff they have to compute.
//
// `topic_configured` is a boolean and never the value. That is a064's contract and
// it is load-bearing here for the first time: before a065 nothing but a hand-edit
// could set a channel, and now a button does it several times a session.
func recordNotificationChange(before, after config.Notifications, created bool) {
	log := openAuditLog()
	if log == nil {
		return
	}
	detail := "operator console, 알림 설정 저장"
	if created {
		detail += " — 채널을 새로 만들었다 (128비트 난수, 값은 기록하지 않는다)"
	}
	if after.Enabled && usesPublicNotificationService(after.BaseURL) {
		detail += " — 공개 ntfy 서비스이며 채널 주소가 유일한 접근 제어다"
	}
	for _, field := range []struct {
		setting  string
		from, to string
	}{
		{"engine.notifications.enabled", boolText(before.Enabled), boolText(after.Enabled)},
		{"engine.notifications.base_url", strings.TrimSpace(before.BaseURL), strings.TrimSpace(after.BaseURL)},
		{"engine.notifications.topic_configured",
			boolText(strings.TrimSpace(before.Topic) != ""),
			boolText(strings.TrimSpace(after.Topic) != "")},
	} {
		if field.from == field.to {
			continue
		}
		_ = log.Record(audit.Entry{
			Action:  audit.ActionNotificationSetting,
			Setting: field.setting,
			Old:     field.from,
			New:     field.to,
			Detail:  detail,
		})
	}
}

// usesPublicNotificationService mirrors obs.Ntfy.UsesPublicService without
// building one, so the audit detail does not depend on a transport existing.
func usesPublicNotificationService(baseURL string) bool {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return base == "" || base == obs.DefaultNtfyBaseURL
}
