package obs_test

// a102_event_values_are_unique_test.go closes the hole a102's A2 review found
// (F7): every assertion about an event's name compared the constant to itself.
//
//	if got["event"] != string(obs.EventRecoveryRateLimited) { … }
//
// That is true no matter what the constant says. A2 changed
// EventRecoveryRateLimited's value to an existing event's value and the whole
// suite stayed green — which would have made a new operator line indistinguishable
// from an old one on the dashboard, the exact defect a099 named ("a line that
// means three things cannot be alerted on").
//
// The enum has no runtime registry to walk, so the source is the enumeration.
// Reading it is the only way to make the check cover every member rather than the
// one member whose test happened to be written.

import (
	"os"
	"regexp"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

var eventDeclaration = regexp.MustCompile(`(?m)^\s*(Event[A-Za-z0-9_]+)\s+EventType\s*=\s*"([^"]+)"`)

func eventConstants(t *testing.T) map[string]string {
	t.Helper()
	body, err := os.ReadFile("event.go")
	if err != nil {
		t.Fatalf("reading event.go: %v", err)
	}
	found := map[string]string{}
	for _, match := range eventDeclaration.FindAllStringSubmatch(string(body), -1) {
		found[match[1]] = match[2]
	}
	if len(found) < 20 {
		t.Fatalf("only %d event constants were found; the declaration shape moved and this "+
			"test would silently stop checking", len(found))
	}
	return found
}

// TestNoTwoEventsShareAValue. The event name is the counting key: two constants
// with one value are two conditions an operator cannot separate.
func TestNoTwoEventsShareAValue(t *testing.T) {
	owner := map[string]string{}
	for name, value := range eventConstants(t) {
		if first, taken := owner[value]; taken {
			t.Errorf("%s and %s both publish %q; a dashboard cannot tell them apart",
				first, name, value)
			continue
		}
		owner[value] = name
	}
}

// TestTheRateLimitedRecoveryEventKeepsItsOwnName pins the literal, because a
// comparison against the constant proves nothing about what the constant is.
func TestTheRateLimitedRecoveryEventKeepsItsOwnName(t *testing.T) {
	const want = "reconcile.recovery_rate_limited"
	if string(obs.EventRecoveryRateLimited) != want {
		t.Fatalf("EventRecoveryRateLimited = %q, want %q", obs.EventRecoveryRateLimited, want)
	}
	if obs.EventRecoveryRateLimited == obs.EventRecoveryComplete {
		t.Fatal("the throttled recovery and the finished recovery publish the same line")
	}
	if obs.EventRecoveryRateLimited.Subject() != "reconcile" {
		t.Errorf("subject = %q, want reconcile", obs.EventRecoveryRateLimited.Subject())
	}
	// 이 줄은 이상 상태의 표지이지 CRITICAL 알림이 아니다 — 게이트는 §1이 닫는다.
	if obs.SeverityOf(obs.EventRecoveryRateLimited) != obs.SeverityNormal {
		t.Errorf("severity = %v, want normal", obs.SeverityOf(obs.EventRecoveryRateLimited))
	}
}
