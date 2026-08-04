package console

// a076_notification_qr_test.go: the card draws the address, draws only the
// address, and draws nothing when there is nothing to draw.

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/qr"
)

var qrRectAttr = regexp.MustCompile(`<rect x="(\d+)" y="(\d+)" width="(\d+)" height="1"`)

// TestTheCardDrawsTheSubscribeAddress.
func TestTheCardDrawsTheSubscribeAddress(t *testing.T) {
	const channel = "tossos-mnzxcvbnmasdfghjklqwerty"
	h := notificationHarness(t, &fakeNotifications{
		block: config.Notifications{Enabled: true, BaseURL: "https://ntfy.sh", Topic: channel},
	})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]

	if !strings.Contains(card, "data-notification-qr") {
		t.Fatalf("the card draws no QR symbol:\n%s", card)
	}
	if !strings.Contains(card, "<svg") || !strings.Contains(card, "viewBox=") {
		t.Errorf("the QR is not an inline SVG:\n%s", card)
	}
	if n := len(qrRectAttr.FindAllString(card, -1)); n < 20 {
		t.Errorf("the symbol has %d module runs, which is too few to be a QR code", n)
	}
}

// TestTheDrawnSymbolIsTheAddress.
//
// The rectangles are read back into a matrix and compared to the symbol the
// encoder makes for the address the card displays. A card that drew the wrong
// channel would be worse than one that drew nothing — the operator would
// subscribe to something and hear silence. That the symbol itself is a correct
// QR code is internal/qr's own claim, checked there against the standard.
func TestTheDrawnSymbolIsTheAddress(t *testing.T) {
	const channel = "tossos-mnzxcvbnmasdfghjklqwerty"
	page := settingsPage{Notifications: config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.sh", Topic: channel,
	}}
	drawing := page.NotificationQR()
	if len(drawing.Rects) == 0 {
		t.Fatal("no symbol was drawn")
	}

	size := drawing.Extent - 2*qrQuietZone
	matrix := make(qr.Matrix, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
	}
	for _, r := range drawing.Rects {
		for dx := 0; dx < r.W; dx++ {
			x, y := r.X-qrQuietZone+dx, r.Y-qrQuietZone
			if x < 0 || y < 0 || x >= size || y >= size {
				t.Fatalf("a module run leaves the symbol: %+v", r)
			}
			matrix[y][x] = true
		}
	}
	want, err := qr.Encode(page.NotificationSubscribeURL())
	if err != nil {
		t.Fatalf("encoding the address the card shows: %v", err)
	}
	if want.Size() != size {
		t.Fatalf("the drawing is %d modules wide, the address encodes to %d", size, want.Size())
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if matrix[y][x] != want[y][x] {
				t.Fatalf("the drawn symbol differs from the address's symbol at (%d,%d); "+
					"the card is showing a code for something else", x, y)
			}
		}
	}
}

// TestTheQuietZoneIsThere. Without it a reader cannot find the finder patterns
// against the page around them.
//
// The margin is asserted as the number four and not as qrQuietZone. Written the
// other way the test measures itself: setting the constant to zero moves both
// sides of the comparison and passes, which is exactly what a mutation run found
// it doing. Four is the standard's requirement, so four is what is written here.
func TestTheQuietZoneIsThere(t *testing.T) {
	page := settingsPage{Notifications: config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-mnzxcvbnmasdfghjklqwerty",
	}}
	drawing := page.NotificationQR()
	if len(drawing.Rects) == 0 {
		t.Fatal("no symbol was drawn")
	}

	const required = 4
	left, top := drawing.Extent, drawing.Extent
	right, bottom := 0, 0
	for _, r := range drawing.Rects {
		if r.X < left {
			left = r.X
		}
		if r.Y < top {
			top = r.Y
		}
		if r.X+r.W > right {
			right = r.X + r.W
		}
		if r.Y+1 > bottom {
			bottom = r.Y + 1
		}
	}
	for _, margin := range []struct {
		side string
		got  int
	}{
		{"left", left},
		{"top", top},
		{"right", drawing.Extent - right},
		{"bottom", drawing.Extent - bottom},
	} {
		if margin.got < required {
			t.Errorf("the %s quiet zone is %d modules, want at least %d — a reader cannot "+
				"separate the finder patterns from the page without it", margin.side, margin.got, required)
		}
	}
	// The drawing has to be the symbol plus its two margins and nothing else; a
	// viewBox larger than the content would shrink the modules for no reason.
	if want := right + required; drawing.Extent != want {
		t.Errorf("the drawing is %d modules across, want %d (symbol + two margins)",
			drawing.Extent, want)
	}
}

// TestNoSymbolIsDrawnWithoutAnAddress.
func TestNoSymbolIsDrawnWithoutAnAddress(t *testing.T) {
	for _, block := range []config.Notifications{
		{},
		{Enabled: true, BaseURL: "https://ntfy.sh"},
		{Enabled: true, Topic: "tossos-abc"},
	} {
		page := settingsPage{Notifications: block}
		if drawing := page.NotificationQR(); len(drawing.Rects) != 0 || drawing.Extent != 0 {
			t.Errorf("%+v drew a symbol out of an incomplete address", block)
		}
	}
}

// TestTheOffCardDrawsNoSymbol.
//
// The address survives a mute so re-enabling keeps the subscription, but a card
// that showed a QR beside 꺼짐 would invite an operator to subscribe to a channel
// nothing is publishing to.
func TestTheOffCardDrawsNoSymbol(t *testing.T) {
	h := notificationHarness(t, &fakeNotifications{
		block: config.Notifications{
			Enabled: false, BaseURL: "https://ntfy.sh", Topic: "tossos-mnzxcvbnmasdfghjklqwerty",
		},
	})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]
	if strings.Contains(card, "data-notification-qr") {
		t.Errorf("a card with alerts off still draws the subscribe symbol:\n%s", card)
	}
}

// TestTheSVGCarriesNoRawMarkupFromTheServer.
//
// Everything inside the element is an integer this package computed. If a channel
// could reach the markup, a channel containing a quote would be an injection —
// so the assertion is that the drawn attributes are numeric, not that they are
// escaped.
func TestTheSVGCarriesNoRawMarkupFromTheServer(t *testing.T) {
	h := notificationHarness(t, &fakeNotifications{
		block: config.Notifications{
			Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-mnzxcvbnmasdfghjklqwerty",
		},
	})
	card := cardsOn(t, body(t, h.get(t, pathSettingsTools)))["notifications"]
	start := strings.Index(card, "<svg")
	end := strings.Index(card, "</svg>")
	if start < 0 || end < 0 {
		t.Fatal("no SVG element on the card")
	}
	svg := card[start : end+6]
	if strings.Contains(svg, "tossos-") {
		t.Errorf("the channel appears literally inside the SVG:\n%s", svg)
	}
	for _, m := range qrRectAttr.FindAllStringSubmatch(svg, -1) {
		for _, value := range m[1:] {
			if _, err := strconv.Atoi(value); err != nil {
				t.Errorf("a module run carries the non-numeric attribute %q", value)
			}
		}
	}
}
