//go:build unix

package main

// a109_the_reattach_wake_outlives_a_broken_aggregate_test.go — a109 §2-fix F3 (A2 P2-3).
//
// 재부착 시도를 깨우는 것은 **요청 경로**다. 그런데 요청이 없어도 데몬은 회복해야
// 하므로, 실제 상시 구동원은 하나뿐이다: 30초마다 도는 publisher 루프
// (`publishHTTPAPISnapshots`)가 집계 스냅샷을 새로 만들고, 그 집계가 부재 판정
// (`httpapi.StrategyRuntimeAbsent`)을 부르면서 시도가 깨어난다.
//
// 그 사슬은 **집계 성공에 얹혀 있었다.** 루프는 `Refresh` 가 오류면 `continue` 하고,
// 집계는 전략과 무관한 일곱 조회(engine·positions·orders·candidates·performance·
// settings·optimization)가 **하나라도** 실패하면 전략 블록에 닿기 전에 오류로 끝난다
// (`httpAPIReader.Snapshot` 의 B1–B7). 그래서 옵티마이저 DB 하나가 고장 나면
// 전략 화면의 재부착이 **영원히** 안 깨어난다 — 엔진이 멀쩡히 돌아와도.
//
// 여기서 재는 것은 그 결합의 부재다: 집계가 깨진 채로도 엔진 복귀가 재부착을 만든다.

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
)

// TestTheReattachWakeSurvivesABrokenAggregate 는 F3 그 자체다.
func TestTheReattachWakeSurvivesABrokenAggregate(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	live := &a109FakeReader{}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return live, true // 엔진은 이미 돌아와 있다 — 깨우기만 하면 붙는다
		})
	// 데몬은 엔진보다 먼저 떴다: 자리는 비어 있고 시도 대상이다.
	attachment.attach(nil, false)

	reader := a108SnapshotReader(t, nil)
	reader.strategyRuntime = attachment
	// 전략과 **무관한** 조회 하나가 고장 난다. 옵티마이저 채택 설정 읽기가 그 자리다.
	reader.adoptionDesired = func() (config.Adoption, string, error) {
		return config.Adoption{}, "", errors.New("optimization: 채택 설정을 읽을 수 없다")
	}

	snapshots, err := newHTTPAPISnapshotCache(reader)
	if err != nil {
		t.Fatal(err)
	}
	// 대조군: 집계가 **실제로** 실패하는지 먼저 본다. 안 그러면 이 테스트는 성공하는
	// 집계에서 초록이 되고 아무것도 재지 않는다.
	if _, err := snapshots.Refresh(context.Background()); err == nil {
		t.Fatal("집계가 성공했다 — 이 테스트가 재려는 상태가 아니다")
	}
	stream, err := httpapi.NewStream(httpapi.StreamOptions{}, snapshots.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stream.Close)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		publishHTTPAPISnapshots(ctx, stream, snapshots, 10*time.Millisecond)
	}()

	a109WaitFor(t, "집계가 깨진 채로의 재부착", func() bool {
		return attempts.Load() >= 1 && !httpapi.StrategyRuntimeAbsent(attachment)
	})
	cancel()
	<-done
}
