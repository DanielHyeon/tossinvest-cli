package clock

import (
	"errors"
	"testing"
	"time"
)

// 참여자마다 time namespace 신원이 다른 배치(호스트 · 컨테이너)를 나타내는 표본임.
// 각 표본은 Snapshot 이 offset 0 을 확인한 것이라 같은 boot 면 같은 CLOCK_BOOTTIME 시계임.
var (
	crossNSA = BootNamespace{Device: 4, Inode: 4026531834}
	crossNSB = BootNamespace{Device: 4, Inode: 4026533806}
	crossNSC = BootNamespace{Device: 4, Inode: 4026533867}
)

func TestBootCrossNamespaceSameBootSubtractsElapsed(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	now := BootStamp{BootID: testBootA, Namespace: crossNSB, Uptime: 13 * time.Second}
	r, e := RemainingBootHint(10*time.Second, saved, now, nil)
	if e != nil || r.Remaining != 7*time.Second || r.NeedsAnchor {
		t.Fatalf("cross namespace same boot %+v %v", r, e)
	}
}

func TestBootCrossNamespaceExpiredHintClears(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	now := BootStamp{BootID: testBootA, Namespace: crossNSB, Uptime: 41 * time.Second}
	r, e := RemainingBootHint(30*time.Second, saved, now, nil)
	if e != nil || r.Remaining != 0 || r.NeedsAnchor {
		t.Fatalf("expired cross namespace hint kept %+v %v", r, e)
	}
}

func TestBootCrossNamespaceAnchorRestart(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	// 다른 boot 의 첫 복원 anchor 를 ns B 가 남기고 ns C 가 재시작해 읽는 경우임.
	anchor := BootStamp{BootID: testBootB, Namespace: crossNSB, Uptime: 100 * time.Second}
	now := BootStamp{BootID: testBootB, Namespace: crossNSC, Uptime: 103 * time.Second}
	r, e := RemainingBootHint(10*time.Second, saved, now, &anchor)
	if e != nil || r.Remaining != 7*time.Second || r.NeedsAnchor {
		t.Fatalf("cross namespace anchor restart %+v %v", r, e)
	}
	// 같은 boot 로 돌아온 anchor(순서 맞음)를 다른 ns 가 남긴 경우도 anchor 기준으로 차감함.
	sameBootAnchor := BootStamp{BootID: testBootA, Namespace: crossNSB, Uptime: 12 * time.Second}
	now = BootStamp{BootID: testBootA, Namespace: crossNSC, Uptime: 15 * time.Second}
	r, e = RemainingBootHint(10*time.Second, saved, now, &sameBootAnchor)
	if e != nil || r.Remaining != 7*time.Second || r.NeedsAnchor {
		t.Fatalf("cross namespace same boot anchor %+v %v", r, e)
	}
}

func TestBootCrossNamespaceOrderingStillRejected(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	// 같은 boot 에서 저장 시점보다 과거인 현재 표본은 ns 와 무관하게 거절함(B3 uptime 항).
	now := BootStamp{BootID: testBootA, Namespace: crossNSB, Uptime: 9 * time.Second}
	if r, e := RemainingBootHint(10*time.Second, saved, now, nil); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("cross namespace reverse accepted %+v %v", r, e)
	}
	// anchor 보다 과거인 현재 표본은 거절함(B4 uptime 항).
	anchor := BootStamp{BootID: testBootB, Namespace: crossNSB, Uptime: 100 * time.Second}
	now = BootStamp{BootID: testBootB, Namespace: crossNSC, Uptime: 99 * time.Second}
	if r, e := RemainingBootHint(10*time.Second, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("cross namespace anchor reverse accepted %+v %v", r, e)
	}
	// duration 0 은 B8 에 닿기 전에 반환하므로 anchor 역행은 B4 만 거절함.
	if r, e := RemainingBootHint(0, saved, now, &anchor); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("zero duration hid cross namespace anchor reverse %+v %v", r, e)
	}
	// 같은 boot 에서 저장 시점보다 앞선 anchor 는 거절함(B2 uptime 항).
	early := BootStamp{BootID: testBootA, Namespace: crossNSB, Uptime: 9 * time.Second}
	now = BootStamp{BootID: testBootA, Namespace: crossNSC, Uptime: 15 * time.Second}
	if r, e := RemainingBootHint(10*time.Second, saved, now, &early); !errors.Is(e, ErrBootStamp) {
		t.Fatalf("cross namespace early anchor accepted %+v %v", r, e)
	}
}

func TestBootCrossNamespaceOtherBootNeedsAnchor(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	now := BootStamp{BootID: testBootB, Namespace: crossNSB, Uptime: 500 * time.Hour}
	// 다른 boot 는 ns 와 무관하게 uptime 을 빼지 않고 원래 duration 전체를 보류함.
	r, e := RemainingBootHint(10*time.Second, saved, now, nil)
	if e != nil || r.Remaining != 10*time.Second || !r.NeedsAnchor {
		t.Fatalf("other boot %+v %v", r, e)
	}
}

func TestBootCrossNamespaceStillRejectsInvalid(t *testing.T) {
	saved := BootStamp{BootID: testBootA, Namespace: crossNSA, Uptime: 10 * time.Second}
	for _, bad := range []BootStamp{
		{BootID: testBootA, Namespace: BootNamespace{Device: 4}, Uptime: 11 * time.Second},
		{BootID: testBootA, Namespace: BootNamespace{Inode: 4026533806}, Uptime: 11 * time.Second},
		{BootID: "not-a-boot-id", Namespace: crossNSB, Uptime: 11 * time.Second},
		{BootID: testBootA, Namespace: crossNSB, Uptime: -1},
	} {
		if r, e := RemainingBootHint(10*time.Second, saved, bad, nil); !errors.Is(e, ErrBootStamp) {
			t.Fatalf("invalid now %+v accepted %+v %v", bad, r, e)
		}
	}
}
