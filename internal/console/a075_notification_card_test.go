package console

// a075_notification_card_test.go is the screen half of "알림은 버튼 하나로
// 켜진다": one press writes, one press proves, and the channel never leaves the
// response body.

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// fakeNotifications records what the screen asked for. It generates a channel
// the way the real seam does — the screen never supplies one — so a test that
// found the screen posting a topic would fail on the seam's signature first.
type fakeNotifications struct {
	block    config.Notifications
	loadErr  error
	saveErr  error
	testErr  error
	enables  int
	disables int
	tests    int
}

func (f *fakeNotifications) Load() (config.Notifications, error) {
	return f.block, f.loadErr
}

func (f *fakeNotifications) Enable() (config.Notifications, error) {
	f.enables++
	if f.saveErr != nil {
		return config.Notifications{}, f.saveErr
	}
	if f.block.Topic == "" {
		f.block.Topic = "tossos-mnzxcvbnmasdfghjklqwerty"
	}
	if f.block.BaseURL == "" {
		f.block.BaseURL = "https://ntfy.sh"
	}
	f.block.Enabled = true
	return f.block, nil
}

func (f *fakeNotifications) Disable() error {
	f.disables++
	if f.saveErr != nil {
		return f.saveErr
	}
	// The channel survives, which is the property the off path exists to have.
	f.block.Enabled = false
	return nil
}

func (f *fakeNotifications) Test(context.Context) error {
	f.tests++
	return f.testErr
}

func notificationHarness(t *testing.T, seam *fakeNotifications) *harness {
	t.Helper()
	h := fullSettingsHarness(t, func(o *Options) { o.Notifications = seam })
	h.authenticate(t)
	return h
}

// TestTheAlertCardOffersOneButtonWhenAlertsAreOff.
//
// One button, no fields. The card is the answer to "engine.notifications와
// TOSSCTL_NTFY_TOKEN 이런거 수동 설정하도록 하지 말고" — a screen that asked for a
// channel name would have moved the hand-edit rather than removed it.
func TestTheAlertCardOffersOneButtonWhenAlertsAreOff(t *testing.T) {
	h := notificationHarness(t, &fakeNotifications{})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]

	if !strings.Contains(card, `action="/settings/notifications/on"`) {
		t.Errorf("the alert card offers no way to turn alerts on:\n%s", card)
	}
	if !strings.Contains(card, `<button type="submit">알림 켜기</button>`) {
		t.Errorf("the alert card has no 알림 켜기 button:\n%s", card)
	}
	if strings.Contains(card, `<input type="text"`) || strings.Contains(card, `<input type="password"`) {
		t.Errorf("the alert card asks the operator to type something; the channel is "+
			"generated and the token is never asked for:\n%s", card)
	}
}

// TestThereIsNoTokenInputOnTheAlertCard.
//
// a074 kept a token out of the config struct so it had nowhere to land. The same
// argument applies to a form field, and it is stated here as a property of the
// rendered page rather than of the handler: a field is enough to make an operator
// paste a secret, whether or not anything reads it.
func TestThereIsNoTokenInputOnTheAlertCard(t *testing.T) {
	h := notificationHarness(t, &fakeNotifications{
		block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-abc"},
	})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]
	for _, form := range formBlock.FindAllStringSubmatch(card, -1) {
		for _, banned := range []string{"token", "password", "secret"} {
			if strings.Contains(strings.ToLower(form[2]), `name="`+banned) {
				t.Errorf("the form posting to %s carries a %q field:\n%s", form[1], banned, form[2])
			}
		}
	}
}

// TestTurningAlertsOnAlsoSendsTheTest.
//
// The two acts are one press because the console and the engine are different
// processes: a save alone cannot be confirmed until the engine restarts, and an
// operator who cannot tell whether it worked will not trust it when it matters.
func TestTurningAlertsOnAlsoSendsTheTest(t *testing.T) {
	seam := &fakeNotifications{}
	h := notificationHarness(t, seam)
	h.post(t, "/settings/notifications/on", url.Values{"csrf": {h.csrf}})

	if seam.enables != 1 {
		t.Errorf("Enable was called %d time(s), want 1", seam.enables)
	}
	if seam.tests != 1 {
		t.Errorf("Test was called %d time(s) in the same press; the operator has no other "+
			"way to find out whether delivery works before the engine restarts", seam.tests)
	}
}

// TestAFailedTestLeavesAlertsOn (design D8, the change's most deliberate call).
//
// Rolling back on a failed send would make whether alerts are configured depend
// on whether the network was up for one second — and the state it would roll back
// to is the one this whole change exists to leave.
func TestAFailedTestLeavesAlertsOn(t *testing.T) {
	seam := &fakeNotifications{testErr: errors.New("dial tcp: no route to host")}
	h := notificationHarness(t, seam)
	resp := h.postNoRedirect(t, "/settings/notifications/on", url.Values{"csrf": {h.csrf}})

	if !seam.block.Enabled {
		t.Error("a failed test message turned alerts back off")
	}
	if seam.disables != 0 {
		t.Errorf("a failed test message called Disable %d time(s)", seam.disables)
	}
	notice := resp.Header.Get("Location")
	if !strings.Contains(notice, url.QueryEscape("no route to host")) {
		t.Errorf("the failure reason did not reach the operator: %s", notice)
	}
}

// TestTheNoticeNeverCarriesTheChannel (design D5).
//
// redirectSettings answers through a query string, and query strings live in
// browser history and in whatever sits in front of this console. The channel is
// the access control on a public ntfy topic, so it is rendered into the body and
// never into a Location header.
func TestTheNoticeNeverCarriesTheChannel(t *testing.T) {
	const channel = "tossos-mnzxcvbnmasdfghjklqwerty"
	for _, route := range []string{
		"/settings/notifications/on",
		"/settings/notifications/test",
		"/settings/notifications/off",
	} {
		seam := &fakeNotifications{
			block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: channel},
		}
		h := notificationHarness(t, seam)
		resp := h.postNoRedirect(t, route, url.Values{"csrf": {h.csrf}})
		location := resp.Header.Get("Location")
		decoded, err := url.QueryUnescape(location)
		if err != nil {
			decoded = location
		}
		if strings.Contains(decoded, channel) {
			t.Errorf("%s put the channel in the redirect URL, where browser history and "+
				"proxy logs keep it: %s", route, decoded)
		}
	}
}

// TestTheCardShowsTheSubscribeAddress. The operator has to be able to read it —
// it is not a secret from them, it is a secret from everyone else.
func TestTheCardShowsTheSubscribeAddress(t *testing.T) {
	const channel = "tossos-mnzxcvbnmasdfghjklqwerty"
	h := notificationHarness(t, &fakeNotifications{
		block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: channel},
	})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]
	if !strings.Contains(card, "https://ntfy.sh/"+channel) {
		t.Errorf("the card does not show the subscribe address; the operator cannot "+
			"subscribe to a channel they cannot read:\n%s", card)
	}
	if !strings.Contains(card, "data-notification-subscribe") {
		t.Errorf("the subscribe address is not marked as such:\n%s", card)
	}
}

// TestTurningAlertsOffKeepsTheChannel.
func TestTurningAlertsOffKeepsTheChannel(t *testing.T) {
	const channel = "tossos-mnzxcvbnmasdfghjklqwerty"
	seam := &fakeNotifications{
		block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: channel},
	}
	h := notificationHarness(t, seam)
	h.post(t, "/settings/notifications/off", url.Values{"csrf": {h.csrf}})

	if seam.block.Enabled {
		t.Error("the off button did not turn alerts off")
	}
	if seam.block.Topic != channel {
		t.Errorf("the off button changed the channel to %q; a mute must not cost a "+
			"re-subscription on every device", seam.block.Topic)
	}
	if seam.tests != 0 {
		t.Errorf("turning alerts off sent %d test message(s)", seam.tests)
	}
}

// TestTheTestButtonChangesNothing.
func TestTheTestButtonChangesNothing(t *testing.T) {
	seam := &fakeNotifications{
		block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-abc"},
	}
	h := notificationHarness(t, seam)
	h.post(t, "/settings/notifications/test", url.Values{"csrf": {h.csrf}})

	if seam.enables != 0 || seam.disables != 0 {
		t.Errorf("the test button wrote settings: %d enable(s), %d disable(s)",
			seam.enables, seam.disables)
	}
	if seam.tests != 1 {
		t.Errorf("the test button sent %d message(s), want 1", seam.tests)
	}
}

// TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton.
func TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton(t *testing.T) {
	h := fullSettingsHarness(t, func(o *Options) { o.Notifications = nil })
	h.authenticate(t)
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]
	if strings.Contains(card, `action="/settings/notifications/on"`) {
		t.Errorf("an unwired build still renders the button:\n%s", card)
	}
	if !strings.Contains(card, "data-save-block=") {
		t.Errorf("an unwired build renders neither a button nor a reason:\n%s", card)
	}
}

// TestAnUnreadableAlertConfigDoesNotTakeTheTabDown.
func TestAnUnreadableAlertConfigDoesNotTakeTheTabDown(t *testing.T) {
	h := notificationHarness(t, &fakeNotifications{loadErr: errors.New("config.json: unexpected }")})
	page := body(t, h.get(t, pathSettingsTools))
	if !strings.Contains(page, "unexpected }") {
		t.Errorf("the read failure is not stated on the page:\n%s", page)
	}
	if !strings.Contains(page, `data-settings-card="system-update"`) {
		t.Error("an unreadable alert block took the rest of the 도구 tab down with it")
	}
}
