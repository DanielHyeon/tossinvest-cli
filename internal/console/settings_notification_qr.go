package console

// settings_notification_qr.go draws the subscribe address as something a phone
// camera can read (change a076).
//
// # Why the card needs this and not just the link
//
// a075's card shows the address and links it, which is enough at the machine. It
// is not enough for the case critical alerts exist for: the operator is somewhere
// else. Getting the channel onto a phone meant typing 26 random characters, which
// is the friction a075 removed from the config file and left on the phone.
//
// # Why the SVG is built from numbers rather than from a string
//
// This package renders with html/template, and handing it pre-built markup means
// template.HTML — a hole in the escaping that has to be argued about every time
// someone reads it. There is no need: a QR symbol is a set of rectangles, so the
// template ranges over integer coordinates and writes the elements itself. No raw
// HTML crosses the boundary and no new escaping seam exists to get wrong.
//
// # Why one rectangle per run
//
// A version-3 symbol is 29x29 with about 400 dark modules. One element each is
// 400 elements on a settings page. Merging horizontal runs cuts it to roughly a
// quarter with no change to what is drawn.

import "github.com/JungHoonGhae/tossinvest-cli/internal/qr"

// qrQuietZone is the light margin the standard requires around a symbol. Without
// it a reader cannot find the finder patterns against a page.
const qrQuietZone = 4

// qrRect is one horizontal run of dark modules, in module coordinates.
type qrRect struct {
	X, Y, W int
}

// notificationQR is the card's symbol: the drawing plus the size its viewBox needs.
type notificationQR struct {
	// Extent is the width and height in modules, quiet zone included.
	Extent int
	Rects  []qrRect
}

// NotificationQR is the subscribe address as a scannable symbol, or the zero
// value when there is no address to draw.
//
// A symbol that cannot be built is not an error the operator can act on — the
// address is still on the card as text and as a link — so this returns nothing
// and the card renders without it.
func (p settingsPage) NotificationQR() notificationQR {
	address := p.NotificationSubscribeURL()
	if address == "" {
		return notificationQR{}
	}
	matrix, err := qr.Encode(address)
	if err != nil {
		return notificationQR{}
	}
	return notificationQR{
		Extent: matrix.Size() + 2*qrQuietZone,
		Rects:  qrRuns(matrix),
	}
}

// qrRuns merges each row's dark modules into horizontal rectangles, offset by the
// quiet zone.
func qrRuns(matrix qr.Matrix) []qrRect {
	out := make([]qrRect, 0, matrix.Size()*2)
	for y, row := range matrix {
		start := -1
		for x := 0; x <= len(row); x++ {
			dark := x < len(row) && row[x]
			switch {
			case dark && start < 0:
				start = x
			case !dark && start >= 0:
				out = append(out, qrRect{
					X: start + qrQuietZone,
					Y: y + qrQuietZone,
					W: x - start,
				})
				start = -1
			}
		}
	}
	return out
}
