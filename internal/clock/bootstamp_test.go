package clock

import (
	"errors"
	"testing"
	"time"
)

const testBootA = "12345678-1234-4234-8234-123456789abc"
const testBootB = "87654321-4321-4321-8321-cba987654321"

func TestBootRemainingConservative(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: 10 * time.Second}
	now := saved
	now.Uptime = 13 * time.Second
	r, e := RemainingBootHint(10*time.Second, saved, now, nil)
	if e != nil || r.Remaining != 7*time.Second || r.NeedsAnchor {
		t.Fatalf("same boot %+v %v", r, e)
	}
	now.BootID = testBootB
	now.Uptime = 200 * time.Second
	r, e = RemainingBootHint(10*time.Second, saved, now, nil)
	if e != nil || r.Remaining != 10*time.Second || !r.NeedsAnchor {
		t.Fatalf("other boot %+v %v", r, e)
	}
	anchor := now
	now.Uptime += 3 * time.Second
	r, e = RemainingBootHint(10*time.Second, saved, now, &anchor)
	if e != nil || r.Remaining != 7*time.Second || r.NeedsAnchor {
		t.Fatalf("anchor %+v %v", r, e)
	}
	now.Uptime = anchor.Uptime - 1
	r, e = RemainingBootHint(10*time.Second, saved, now, &anchor)
	if e == nil {
		t.Fatalf("reverse accepted %+v", r)
	}
	_, e = RemainingBootHint(31*time.Second, saved, now, nil)
	if !errors.Is(e, ErrBootStamp) {
		t.Fatalf("duration: %v", e)
	}
}

func TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: 10 * time.Second}
	anchor := BootStamp{BootID: testBootB, Namespace: BootNamespace{Device: 1, Inode: 9}, Uptime: time.Second}
	now := saved
	now.Namespace.Inode++
	if r, e := RemainingBootHint(10*time.Second, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("namespace drift accepted %+v %v", r, e)
	}
	now = saved
	now.Uptime--
	if r, e := RemainingBootHint(10*time.Second, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("uptime reverse accepted %+v %v", r, e)
	}
}

func TestBootZeroDurationDoesNotHideInvalidClock(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: 10 * time.Second}
	now := saved
	r, e := RemainingBootHint(0, saved, now, nil)
	if e != nil || r.Remaining != 0 || r.NeedsAnchor {
		t.Fatalf("zero: %+v %v", r, e)
	}
	now.Namespace.Inode++
	if _, e = RemainingBootHint(0, saved, now, nil); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("zero masked mismatch %v", e)
	}
}

func TestBootRemainingBoundariesAndRestarts(t *testing.T) {
	s := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: 20 * time.Second}
	same := s
	same.Uptime += 31 * time.Second
	if r, e := RemainingBootHint(30*time.Second, s, same, nil); e != nil || r.Remaining != 0 || r.NeedsAnchor {
		t.Fatalf("expired %+v %v", r, e)
	}
	reboot := BootStamp{BootID: testBootB, Namespace: BootNamespace{Device: 1, Inode: 3}, Uptime: 500 * time.Hour}
	if r, e := RemainingBootHint(30*time.Second, s, reboot, nil); e != nil || r.Remaining != 30*time.Second || !r.NeedsAnchor {
		t.Fatalf("reboot %+v %v", r, e)
	}
	anchor := reboot
	reboot.Uptime += 29 * time.Second
	if r, e := RemainingBootHint(30*time.Second, s, reboot, &anchor); e != nil || r.Remaining != time.Second || r.NeedsAnchor {
		t.Fatalf("restart %+v %v", r, e)
	}
	reboot.Uptime += time.Second
	if r, e := RemainingBootHint(30*time.Second, s, reboot, &anchor); e != nil || r.Remaining != 0 || r.NeedsAnchor {
		t.Fatalf("expiry %+v %v", r, e)
	}
	returned := s
	returned.Uptime += 30 * time.Second
	if r, e := RemainingBootHint(30*time.Second, s, returned, &anchor); e != nil || r.Remaining != 30*time.Second || !r.NeedsAnchor {
		t.Fatalf("return to prior boot %+v %v", r, e)
	}
	if r, e := RemainingBootHint(0, s, reboot, nil); e != nil || r.Remaining != 0 || r.NeedsAnchor {
		t.Fatalf("zero %+v %v", r, e)
	}
}
func TestBootRemainingInvalidStamps(t *testing.T) {
	s := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: time.Second}
	for _, bad := range []BootStamp{{}, {BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: -1}, {BootID: testBootA, Namespace: BootNamespace{Device: 1}, Uptime: time.Second}} {
		if _, e := RemainingBootHint(time.Second, s, bad, nil); !errors.Is(e, ErrBootStamp) {
			t.Fatal("invalid now accepted")
		}
		if _, e := RemainingBootHint(time.Second, bad, s, nil); !errors.Is(e, ErrBootStamp) {
			t.Fatal("invalid saved accepted")
		}
		if _, e := RemainingBootHint(time.Second, s, s, &bad); !errors.Is(e, ErrBootStamp) {
			t.Fatal("invalid anchor accepted")
		}
	}
	if _, e := RemainingBootHint(-1, s, s, nil); !errors.Is(e, ErrBootStamp) {
		t.Fatal("negative duration accepted")
	}
}

func TestBootRemainingRejectsAnchorBeforeSameBootProvenance(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: BootNamespace{Device: 1, Inode: 2}, Uptime: 10 * time.Second}
	anchor := saved
	anchor.Uptime = 0
	now := saved
	now.Uptime += time.Second
	if r, e := RemainingBootHint(10*time.Second, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("early anchor accepted %+v %v", r, e)
	}
	anchor = saved
	anchor.Namespace.Inode++
	if _, e := RemainingBootHint(10*time.Second, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatal("conflicting same boot anchor accepted")
	}
}
