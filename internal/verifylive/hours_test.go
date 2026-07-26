package verifylive

// hours_test.go covers the advisory's two claims: it reads the clock in
// Asia/Seoul whatever zone it is handed, and it is the same window the evidence
// line already records.

import (
	"strings"
	"testing"
	"time"
)

func TestKRSessionAdvisoryWindow(t *testing.T) {
	kst := time.FixedZone("KST", 9*60*60)
	cases := []struct {
		name    string
		at      time.Time
		outside bool
		label   string
	}{
		// 2026-07-27 is a Monday.
		{"just before the open", time.Date(2026, 7, 27, 8, 59, 0, 0, kst), true, "outside KR regular hours"},
		{"the open", time.Date(2026, 7, 27, 9, 0, 0, 0, kst), false, "KR regular hours"},
		{"midday", time.Date(2026, 7, 27, 12, 30, 0, 0, kst), false, "KR regular hours"},
		{"one minute before the close", time.Date(2026, 7, 27, 15, 29, 0, 0, kst), false, "KR regular hours"},
		{"the close", time.Date(2026, 7, 27, 15, 30, 0, 0, kst), true, "outside KR regular hours"},
		{"the evening", time.Date(2026, 7, 27, 22, 0, 0, 0, kst), true, "outside KR regular hours"},
		{"friday midday", time.Date(2026, 7, 31, 12, 0, 0, 0, kst), false, "KR regular hours"},
		{"saturday midday", time.Date(2026, 8, 1, 12, 0, 0, 0, kst), true, "weekend"},
		// The Sunday the 422 was measured on.
		{"the sunday of the first run", time.Date(2026, 7, 26, 21, 31, 0, 0, kst), true, "weekend"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := KRSessionAdvisory(tc.at)
			if got.Outside != tc.outside {
				t.Errorf("Outside = %v, want %v (%s)", got.Outside, tc.outside, tc.at)
			}
			if got.Label != tc.label {
				t.Errorf("Label = %q, want %q", got.Label, tc.label)
			}
			if strings.TrimSpace(got.Detail) == "" {
				t.Error("the advisory has no sentence for an operator to read")
			}
		})
	}
}

// TestKRSessionAdvisoryConvertsFromAnyZone. The console's clock reports UTC; a
// window judged in the process timezone would be wrong on every machine that is
// not already in Seoul.
func TestKRSessionAdvisoryConvertsFromAnyZone(t *testing.T) {
	// 2026-07-27 03:00 UTC is 12:00 KST on a Monday — inside the session.
	inside := KRSessionAdvisory(time.Date(2026, 7, 27, 3, 0, 0, 0, time.UTC))
	if inside.Outside {
		t.Errorf("03:00 UTC Monday judged outside the session: %+v", inside)
	}
	if h := inside.At.Hour(); h != 12 {
		t.Errorf("At.Hour() = %d, want 12 — the reading is not in Asia/Seoul", h)
	}
	// 2026-07-27 08:00 UTC is 17:00 KST — after the close.
	after := KRSessionAdvisory(time.Date(2026, 7, 27, 8, 0, 0, 0, time.UTC))
	if !after.Outside {
		t.Errorf("08:00 UTC Monday judged inside the session: %+v", after)
	}
}

// TestTheAdvisoryNamesTheMeasuredErrorCode. The warning is only worth reading if
// it points at the evidence; a vague "orders may fail" is a warning people learn
// to click past.
func TestTheAdvisoryNamesTheMeasuredErrorCode(t *testing.T) {
	kst := time.FixedZone("KST", 9*60*60)
	for _, at := range []time.Time{
		time.Date(2026, 8, 1, 12, 0, 0, 0, kst),  // Saturday
		time.Date(2026, 7, 27, 22, 0, 0, 0, kst), // Monday evening
	} {
		got := KRSessionAdvisory(at)
		if !strings.Contains(got.Detail, "order-hours-closed") {
			t.Errorf("the advisory for %s does not name the measured code: %q", at, got.Detail)
		}
	}
}

// TestSessionLabelStillReadsTheSameWindow: the evidence line and the advisory
// must not drift into disagreeing about what "regular hours" means.
func TestSessionLabelStillReadsTheSameWindow(t *testing.T) {
	kst := time.FixedZone("KST", 9*60*60)
	cases := map[string]time.Time{
		"KR regular hours 12:00 KST":         time.Date(2026, 7, 27, 12, 0, 0, 0, kst),
		"outside KR regular hours 16:00 KST": time.Date(2026, 7, 27, 16, 0, 0, 0, kst),
		"weekend 2026-08-01 12:00 KST":       time.Date(2026, 8, 1, 12, 0, 0, 0, kst),
		"outside KR regular hours 08:59 KST": time.Date(2026, 7, 27, 8, 59, 0, 0, kst),
		"KR regular hours 09:00 KST":         time.Date(2026, 7, 27, 9, 0, 0, 0, kst),
		"outside KR regular hours 15:30 KST": time.Date(2026, 7, 27, 15, 30, 0, 0, kst),
		"weekend 2026-07-26 21:31 KST":       time.Date(2026, 7, 26, 21, 31, 0, 0, kst),
	}
	for want, at := range cases {
		r := &Runner{now: func() time.Time { return at }}
		if got := r.sessionLabel(); got != want {
			t.Errorf("sessionLabel() = %q, want %q", got, want)
		}
	}
}
