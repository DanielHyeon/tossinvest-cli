package reconcile_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a102 D1b — 수집 오류가 브로커의 429였다는 사실은 오류를 타고 밖으로 나가야 한다.
//
// 왜 이것이 먼저인가: 복구가 429를 알아보려면 `errors.Is(err, official.ErrRateLimited)`가
// 성립해야 하는데, `Collect`는 원인을 `%v`로 감싸 정체를 문자열로 만든다. 그 상태에서는
// 복구 쪽에 무슨 코드를 넣어도 429를 볼 수 없다. 그래서 겹1의 첫 테스트는 복구가 아니라
// 여기다.

// TestCollectPreservesTheRateLimitIdentity: 세 읽기 중 어디서 429가 나든 그 정체가
// 호출자에게 도달한다.
//
// 세 갈래를 따로 재는 이유는 wrap 지점이 세 곳이기 때문이다(snapshot.go의 주문·보유·현금).
// 한 곳만 고치고 나머지를 잊으면, 그 엔드포인트에서 시작된 부팅 429는 여전히 복구를
// 죽인다 — 2026-08-13 02:03:30.545Z에 실제로 일어난 일의 모양이다.
func TestCollectPreservesTheRateLimitIdentity(t *testing.T) {
	t.Run("open-order list", func(t *testing.T) {
		c, _, orders, _, _ := newCollector(t)
		orders.err = official.ErrRateLimited
		assertRateLimitSurvives(t, c)
	})

	t.Run("holdings", func(t *testing.T) {
		c, _, _, holdings, _ := newCollector(t)
		holdings.err = official.ErrRateLimited
		assertRateLimitSurvives(t, c)
	})

	t.Run("buying power", func(t *testing.T) {
		c, _, _, holdings, balance := newCollector(t)
		holdings.positions = []domain.Position{{Symbol: "AAPL", Quantity: 10}}
		balance.err = official.ErrRateLimited
		assertRateLimitSurvives(t, c)
	})
}

// assertRateLimitSurvives는 두 판정이 **동시에** 성립하는지를 본다.
//
// 둘 다 필요하다. 429 판정만 보면 `ErrPartialSnapshot` 소비자(대사 드라이버의 mismatch
// 경로)를 깨뜨리고도 초록일 수 있고, 부분 스냅샷 판정만 보면 오늘의 코드가 그대로 통과한다.
func assertRateLimitSurvives(t *testing.T, c *reconcile.Collector) {
	t.Helper()
	snap, err := c.Collect(context.Background())
	if err == nil {
		t.Fatal("a refused read must not produce a snapshot")
	}
	if !snap.Empty() {
		t.Fatalf("a failed collect must return no snapshot, got %+v", snap)
	}
	if !errors.Is(err, reconcile.ErrPartialSnapshot) {
		t.Errorf("errors.Is(err, ErrPartialSnapshot) = false; err = %v", err)
	}
	if !errors.Is(err, official.ErrRateLimited) {
		t.Errorf("errors.Is(err, official.ErrRateLimited) = false; err = %v\n"+
			"the cause is wrapped with %%v, so the 429 is a string by the time recovery reads it", err)
	}
}

// TestCollectStillReportsAPartialSnapshot: wrap 동사를 바꿔도 기존 소비자의 판정은
// 그대로여야 한다. 그리고 429가 아닌 오류가 429로 보여서도 안 된다.
func TestCollectStillReportsAPartialSnapshot(t *testing.T) {
	boom := errors.New("broker said no")

	t.Run("an ordinary failure is still a partial snapshot", func(t *testing.T) {
		c, _, orders, _, _ := newCollector(t)
		orders.err = boom
		_, err := c.Collect(context.Background())
		if !errors.Is(err, reconcile.ErrPartialSnapshot) {
			t.Fatalf("err = %v, want ErrPartialSnapshot", err)
		}
		if errors.Is(err, official.ErrRateLimited) {
			t.Fatalf("an ordinary failure must not read as a rate limit: %v", err)
		}
		if !errors.Is(err, boom) {
			t.Fatalf("the cause must survive: %v", err)
		}
	})

	t.Run("an unreadable order is still a partial snapshot", func(t *testing.T) {
		c, _, orders, _, _ := newCollector(t)
		orders.pages = map[string]execgw.OrderPage{
			"": {Orders: []json.RawMessage{json.RawMessage(`{"symbol":"AAPL"}`)}},
		}
		_, err := c.Collect(context.Background())
		if !errors.Is(err, reconcile.ErrPartialSnapshot) {
			t.Fatalf("err = %v, want ErrPartialSnapshot", err)
		}
		if errors.Is(err, official.ErrRateLimited) {
			t.Fatalf("a parse failure must not read as a rate limit: %v", err)
		}
	})
}
