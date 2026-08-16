package httpapi

// a109_absence_is_a_state_not_a_nil_test.go — a109 §2.3 (design D4, freeze P1-4).
//
// 이 패키지의 소비자들은 「이 배포에 전략 화면이 있는가」를 **nil 검사**로 물어 왔다.
// a109 는 엔진이 돌아오면 다시 붙는 wrapper 를 그 자리에 꽂는데, wrapper 는 정의상
// non-nil 이라 그 검사가 영원히 거짓이 된다 — 전략 화면을 안 쓰는 배포의 REST 응답이
// dormant 스냅샷에서 **오류**로 바뀐다. a108 D4-2 가 금지한 접힘의 같은 모양이다.
//
// 설계가 열거한 nil 검사는 두 곳이었고(집계 스냅샷·SSE helper), production REST 경로인
// router.go 의 것이 빠져 있었다(issues.md T2-1). 여기서는 그 세 번째 자리를 잰다.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a109PresentReader 는 「나는 지금 부재다/아니다」를 스스로 말하는 reader 다.
// cmd/tossctl 의 재부착 wrapper 가 이 모양이다.
type a109PresentReader struct {
	configured bool
	snapshot   strategyprojection.Snapshot
	err        error
}

func (r a109PresentReader) Read(context.Context) (strategyprojection.Snapshot, error) {
	return r.snapshot, r.err
}

func (r a109PresentReader) StrategyRuntimeConfigured() bool { return r.configured }

func TestAbsenceIsAskedAsAStateNotANil(t *testing.T) {
	live := a109PresentReader{configured: true}
	for name, test := range map[string]struct {
		reader StrategyRuntimeReader
		want   bool
	}{
		"nil reader 는 부재다":                {reader: nil, want: true},
		"상태를 말하지 않는 reader 는 있다":          {reader: apiStrategyRuntimeStub{}, want: false},
		"부재라고 말하는 wrapper 는 부재다":          {reader: a109PresentReader{configured: false}, want: true},
		"붙었다고 말하는 wrapper 는 non-nil 이상이다": {reader: live, want: false},
	} {
		if got := StrategyRuntimeAbsent(test.reader); got != test.want {
			t.Errorf("%s: StrategyRuntimeAbsent = %v, want %v", name, got, test.want)
		}
	}
}

// TestTheRESTRouteStaysDormantForAnUnconfiguredWrapper 는 router.go 의 자리다.
//
// 전략 화면을 안 쓰는 배포에서 이 경로는 **200 + dormant 스냅샷**이어야 한다. wrapper 가
// 꽂혔다고 그 값이 오류로 바뀌면, 운영자는 「이 배포는 그 기능을 안 쓴다」와 「엔진이
// 죽었다」를 화면에서 구별할 수 없다.
func TestTheRESTRouteStaysDormantForAnUnconfiguredWrapper(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	for name, reader := range map[string]StrategyRuntimeReader{
		"reader 자리가 비었다":         nil,
		"wrapper 가 아직 못 붙었다(부재)": a109PresentReader{configured: false},
	} {
		handler, err := NewRouter(Options{
			Reader: contractReader{}, StrategyRuntime: reader,
			Now: func() time.Time { return now },
		})
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/strategy-runtime", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: status=%d body=%s — 부재가 오류가 됐다", name, recorder.Code, recorder.Body.String())
		}
		var envelope struct {
			Data strategyprojection.Snapshot `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, market := range []strategyprojection.Market{
			strategyprojection.MarketKR, strategyprojection.MarketUS,
		} {
			projection, ok := envelope.Data.Markets[market]
			if !ok || projection.Error == nil ||
				projection.Error.Code != strategyprojection.RefusalNotConfigured {
				t.Errorf("%s: %s 판정 = %+v, want %q", name, market, projection.Error,
					strategyprojection.RefusalNotConfigured)
			}
		}
	}
}

// TestTheStreamHelperRefusesAnUnconfiguredWrapper 는 SSE helper 의 같은 자리다.
//
// # 부재 fixture 가 **읽을 수 있는** 스냅샷을 드는 이유 (뮤테이션 M22 가 가르쳐 준 것)
//
// 처음에는 부재 fixture 를 빈 스냅샷으로 뒀다. 그러면 판정을 nil 검사로 되돌려도
// (부재 wrapper 는 non-nil 이므로 통과) 그 다음 줄의 `Validate` 가 빈 스냅샷을 거절해
// **오류가 나오고**, 테스트는 「거절했다」로 초록이었다 — 즉 다른 이유로 통과했다
// (뮤테이션 원장 M22, 생존). 그래서 부재 fixture 도 **유효한** dormant 스냅샷을 든다:
// 이제 오류를 낼 수 있는 것은 부재 판정 하나뿐이다.
func TestTheStreamHelperRefusesAnUnconfiguredWrapper(t *testing.T) {
	now := func() time.Time { return time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC) }
	if _, err := StrategyRuntimeSnapshotFunc(a109PresentReader{
		configured: false, snapshot: strategyprojection.DormantSnapshot(now()),
	}, now)(context.Background()); err == nil {
		t.Error("부재 wrapper 가 stream 스냅샷을 만들어 냈다 — nil 시절의 뜻과 달라졌다")
	}
	body, err := StrategyRuntimeSnapshotFunc(a109PresentReader{
		configured: true, snapshot: strategyprojection.DormantSnapshot(now()),
	}, now)(context.Background())
	if err != nil || len(body) == 0 {
		t.Errorf("붙은 wrapper 의 stream 스냅샷이 실패했다: %v", err)
	}
}
