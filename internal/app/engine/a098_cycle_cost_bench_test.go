package engine_test

// a098 task 3.2 — 루프 주기를 **재서** 정한다.
//
// 이 파일은 테스트가 아니라 **측정 도구**다. 벤치마크이므로 `go test`가 기본으로
// 안 돌리고 `make test`를 느리게 하지 않는다. 돌리는 법:
//
//	go test ./internal/app/engine/ -run '^$' -bench 'A098' -benchtime 200x
//
// 왜 벤치마크인가: 3.2 는 *"a092 의 값을 인용하지 않는다"*를 요구한다. 인용하지 않으려면
// 이 저장소에서 실제로 재야 하고, 잰 것을 **다시 잴 수 있어야** 그 값이 근거가 된다.
// 숫자를 tasks.md 에 적고 재현 명령을 안 남기면 그것도 결국 인용이다.
//
// 무엇을 재는가 — 주기를 가르는 것은 두 비용이다.
//
//	빈 주기      밀린 행이 없을 때 한 주기가 원장에 지우는 비용. 이것이 주기의 하한을
//	             정한다: 이 비용 × (1초/주기) 가 무시할 만해야 한다.
//	배치 읽기    밀린 행이 있을 때 batch 크기별 읽기 비용. 4.0b 의 배치를 정한다.
//	claim→settle 행 하나를 실제로 처리하는 원장 왕복(쓰기 둘). 한 주기가 batch 행을
//	             처리하는 데 드는 시간의 하한이다.
//
// 재지 **않는** 것: publish 자체. 그것은 네트워크이고 `DefaultPublishTimeout`(10s)이
// 이미 상한을 준다. 주기를 정하는 데 필요한 것은 **원장 쪽 비용**이다 — publish 는
// 어차피 주기보다 오래 걸릴 수 있고, 그래서 4.0 이 잠금을 안 쥐는 것이다(D1.1).

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a098BenchJournal opens a throwaway journal on the benchmark's temp dir.
func a098BenchJournal(b *testing.B) *journal.Journal {
	b.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(b.TempDir(), journal.DBFileName),
		Clock:    clock.NewFake(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = j.Close() })
	return j
}

// a098SeedPending writes n PENDING critical alert rows.
func a098SeedPending(b *testing.B, j *journal.Journal, n int) {
	b.Helper()
	ctx := context.Background()
	for i := range n {
		if _, err := j.EnqueueAlert(ctx, journal.Alert{
			EventKey: fmt.Sprintf("a098-bench-%d", i),
			Type:     "engine.alert_bench",
			Severity: "critical",
			Title:    "bench",
			Body:     "bench body",
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkA098EmptyCycleRead is the floor: what one cycle costs when there is
// nothing to send. Every tick of the delivery executor pays at least this.
func BenchmarkA098EmptyCycleRead(b *testing.B) {
	j := a098BenchJournal(b)
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		rows, err := j.PendingAlerts(ctx, 50)
		if err != nil {
			b.Fatal(err)
		}
		if len(rows) != 0 {
			b.Fatalf("empty outbox returned %d rows", len(rows))
		}
	}
}

// BenchmarkA098BatchRead measures the batch knob (PendingAlerts B1's true arm —
// the branch nothing in the tree has ever executed; see the pendingalerts
// branch-test map). backlog is held at 1000 so the batch is what varies.
func BenchmarkA098BatchRead(b *testing.B) {
	for _, batch := range []int{1, 10, 50, 100, 0 /* 0 = all, what Flush does */} {
		name := fmt.Sprintf("batch=%d", batch)
		if batch == 0 {
			name = "batch=all"
		}
		b.Run(name, func(b *testing.B) {
			j := a098BenchJournal(b)
			a098SeedPending(b, j, 1000)
			ctx := context.Background()
			want := batch
			if batch == 0 || batch > 1000 {
				want = 1000
			}
			b.ResetTimer()
			for range b.N {
				rows, err := j.PendingAlerts(ctx, batch)
				if err != nil {
					b.Fatal(err)
				}
				if len(rows) != want {
					b.Fatalf("limit %d returned %d rows, want %d", batch, len(rows), want)
				}
			}
		})
	}
}

// BenchmarkA098ClaimSettleRoundTrip is what one row costs once the loop decides
// to send it: two ledger writes around the publish. It bounds how many rows a
// cycle can retire, independent of the network.
func BenchmarkA098ClaimSettleRoundTrip(b *testing.B) {
	j := a098BenchJournal(b)
	ctx := context.Background()
	a098SeedPending(b, j, b.N)
	rows, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		b.Fatal(err)
	}
	if len(rows) < b.N {
		b.Fatalf("seeded %d rows, want %d", len(rows), b.N)
	}
	b.ResetTimer()
	for i := range b.N {
		claim, err := j.ClaimAlertByID(ctx, rows[i].ID, "a098-bench")
		if err != nil {
			b.Fatal(err)
		}
		if claim.Disposition != journal.ClaimAcquired {
			b.Fatalf("row %d: disposition %v, want acquired", rows[i].ID, claim.Disposition)
		}
		if _, err := j.MarkAlertDelivered(ctx, rows[i].ID, claim.Token); err != nil {
			b.Fatal(err)
		}
	}
}
