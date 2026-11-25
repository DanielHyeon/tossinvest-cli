//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 다른 time namespace 참여자의 표본을 흉내 내도록 fake 의 ns 신원만 바꿈.
func crossNSAt(boot *a166Boot, inode uint64) {
	boot.stamp.Namespace = clock.BootNamespace{Device: 4, Inode: inode}
}

// 429 hint 를 한 번 기록하고 Begin · Finish 오류를 그대로 돌려줌.
func crossNSSend(t *testing.T, b *Budget, s *ObservationStore, key ObservationKey, hint time.Duration) error {
	t.Helper()
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if e := lease.Release(); e != nil {
			t.Fatal(e)
		}
	}()
	n, err := s.Begin(ctx, lease, ObservationPoll, key)
	if err != nil {
		return err
	}
	return s.Finish(ctx, lease, n, ObservationRateLimited, &hint)
}

func TestA166ObservationCrossNamespaceHintRestore(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	var key ObservationKey
	key[0] = 9
	crossNSAt(boot, 4026531834)
	if err := crossNSSend(t, b, s, key, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	// 컨테이너 재생성처럼 다른 ns 의 참여자가 같은 boot 에서 복원하면 경과를 빼야 함.
	crossNSAt(boot, 4026533806)
	boot.stamp.Uptime += 3 * time.Second
	if got, err := s.PollRemaining(ctx, key); err != nil || got != 7*time.Second {
		t.Fatalf("cross namespace restore=%v %v", got, err)
	}
	// 다른 ns 의 429 갱신이 성공해야 pending 이 남지 않고 다음 Begin 이 slot 을 poison 하지 않음.
	crossNSAt(boot, 4026533867)
	if err := crossNSSend(t, b, s, key, 5*time.Second); err != nil {
		t.Fatalf("cross namespace finish: %v", err)
	}
	crossNSAt(boot, 4026531834)
	if got, err := s.PollRemaining(ctx, key); err != nil || got != 7*time.Second {
		t.Fatalf("after cross namespace finish=%v %v", got, err)
	}
	if err := crossNSSend(t, b, s, key, time.Second); err != nil {
		t.Fatalf("begin after cross namespace finish: %v", err)
	}
	if _, err := s.PollRemaining(ctx, key); err != nil {
		t.Fatalf("slot unusable after cross namespace finish (poisoned=%v): %v", errors.Is(err, ErrObservationPoisoned), err)
	}
	// 만료 뒤 또 다른 ns 에서 읽으면 slot 이 지워져야 함.
	crossNSAt(boot, 4026533806)
	boot.stamp.Uptime += 11 * time.Second
	if got, err := s.PollRemaining(ctx, key); err != nil || got != 0 {
		t.Fatalf("expired cross namespace=%v %v", got, err)
	}
	state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	if observationSlotIndex(state.Slots, key) >= 0 {
		t.Fatalf("expired cross namespace slot kept: %+v", state.Slots)
	}
}

func TestA166ObservationCrossNamespacePrune(t *testing.T) {
	b, s, boot := a166Store(t)
	crossNSAt(boot, 4026531834)
	state := observationState{Version: 1, Instance: "instance", Scope: "scope", Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: make([]observationSlot, observationSlots)}
	for i := range state.Slots {
		var key ObservationKey
		key[0] = byte(i)
		key[1] = 1
		state.Slots[i] = observationSlot{Key: key, Hint: &observationHint{Stamp: boot.stamp, Duration: time.Second}}
	}
	if err := s.write(state); err != nil {
		t.Fatal(err)
	}
	// 가득 찬 slot 을 다른 ns 참여자가 정리할 때 만료된 hint 를 지우고 새 hint 를 받아야 함.
	crossNSAt(boot, 4026533806)
	boot.stamp.Uptime += 2 * time.Second
	var key ObservationKey
	key[0] = 3
	key[1] = 2
	if err := crossNSSend(t, b, s, key, time.Second); err != nil {
		t.Fatalf("cross namespace prune: %v", err)
	}
	after, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Slots) != 1 || after.Slots[0].Key != key {
		t.Fatalf("cross namespace prune kept %d slots", len(after.Slots))
	}
}
