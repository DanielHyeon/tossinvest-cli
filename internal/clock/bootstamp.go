package clock

import (
	"errors"
	"time"
)

// ErrBootStamp denotes an invalid or unsupported boot-time observation.
var ErrBootStamp = errors.New("boot time observation unavailable")

// BootNamespace identifies a time namespace without storing account data.
type BootNamespace struct {
	Device uint64
	Inode  uint64
}

// BootStamp is a validated boot/namespace and elapsed boot time observation.
type BootStamp struct {
	BootID    string
	Namespace BootNamespace
	Uptime    time.Duration
}

// BootSource samples a validated boot clock without changing Clock's contract.
type BootSource interface{ Snapshot() (BootStamp, error) }

// HintRemaining requires a durable restore anchor before a new boot may proceed.
type HintRemaining struct {
	Remaining   time.Duration
	NeedsAnchor bool
}

func validBootID(id string) bool {
	if len(id) != 36 {
		return false
	}
	if id[14] < '1' || id[14] > '8' || (id[19] != '8' && id[19] != '9' && id[19] != 'a' && id[19] != 'b') {
		return false
	}
	allZero := true
	for i, c := range []byte(id) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
		if c != '0' {
			allZero = false
		}
	}
	return !allZero
}
func validStamp(s BootStamp) bool {
	return validBootID(s.BootID) && s.Namespace.Device != 0 && s.Namespace.Inode != 0 && s.Uptime >= 0
}

// RemainingBootHint never interprets a new boot's uptime as elapsed provenance.
// NeedsAnchor requires the caller to persist now before permitting the poll.
func RemainingBootHint(duration time.Duration, saved, now BootStamp, anchor *BootStamp) (HintRemaining, error) {
	if duration < 0 || duration > 30*time.Second || !validStamp(saved) || !validStamp(now) || (anchor != nil && !validStamp(*anchor)) {
		return HintRemaining{}, ErrBootStamp
	}
	if anchor != nil && anchor.BootID == saved.BootID && (anchor.Namespace != saved.Namespace || anchor.Uptime < saved.Uptime) {
		return HintRemaining{}, ErrBootStamp
	}
	if now.BootID == saved.BootID && (now.Namespace != saved.Namespace || now.Uptime < saved.Uptime) {
		return HintRemaining{}, ErrBootStamp
	}
	if anchor != nil && now.BootID == anchor.BootID && (now.Namespace != anchor.Namespace || now.Uptime < anchor.Uptime) {
		return HintRemaining{}, ErrBootStamp
	}
	if duration == 0 {
		return HintRemaining{}, nil
	}
	base := saved
	if anchor != nil {
		base = *anchor
	}
	if now.BootID != base.BootID {
		return HintRemaining{Remaining: duration, NeedsAnchor: true}, nil
	}
	if now.Namespace != base.Namespace || now.Uptime < base.Uptime {
		return HintRemaining{}, ErrBootStamp
	}
	elapsed := now.Uptime - base.Uptime
	if elapsed >= duration {
		return HintRemaining{}, nil
	}
	return HintRemaining{Remaining: duration - elapsed}, nil
}
