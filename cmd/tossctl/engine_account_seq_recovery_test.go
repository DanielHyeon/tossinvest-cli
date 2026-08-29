//go:build tossos_testseams

package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func assembleAccountSequenceTestEngine(
	t *testing.T,
	handler http.Handler,
	officialOptions ...official.Option,
) (*engine.Context, string, *httptest.Server, error) {
	t.Helper()
	dir := isolate(t)
	writeConfigFile(t, dir, config.DefaultFile())
	writeCredentials(t, dir)
	srv := httptest.NewServer(handler)
	options := []official.Option{
		official.WithBaseURL(srv.URL),
		official.WithHTTPClient(srv.Client()),
	}
	options = append(options, officialOptions...)
	ectx, err := assembleEngine(
		context.Background(),
		&rootOptions{configDir: dir},
		clock.System(),
		obs.NewLogger(obs.LogOptions{Writer: io.Discard}),
		func(opts *engine.Options) {
			engine.ConfigureCLIRegressionForTest(
				opts,
				journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
				nil,
			)
		},
		options...,
	)
	return ectx, dir, srv, err
}

func writeEngineAccountSequenceToken(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
}

func writeEngineAccountSequenceAccounts(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"result":`+body+`}`)
}

func writeEmptyEngineSnapshotResponse(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/json")
	switch path {
	case "/api/v1/orders":
		_, _ = io.WriteString(w, `{"result":{"orders":[],"nextCursor":"","hasNext":false}}`)
	case "/api/v1/holdings":
		_, _ = io.WriteString(w, `{"result":{"items":[]}}`)
	case "/api/v1/buying-power":
		_, _ = io.WriteString(w, `{"result":{"cashBuyingPower":"1000","currency":"KRW"}}`)
	default:
		http.Error(w, "unexpected path", http.StatusNotFound)
	}
}

func assertActualEngineRecoveryReusesTheStartupAccountSequence(
	t *testing.T,
	officialOptions ...official.Option,
) {
	t.Helper()
	var accountCalls atomic.Int32
	var ordersCalls atomic.Int32
	var holdingsCalls atomic.Int32
	var balanceCalls atomic.Int32
	var wrongHeaders atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeEngineAccountSequenceToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) != 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			writeEngineAccountSequenceAccounts(w,
				`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
		case "/api/v1/orders", "/api/v1/holdings", "/api/v1/buying-power":
			if r.Header.Get("X-Tossinvest-Account") != "7" {
				wrongHeaders.Add(1)
			}
			switch r.URL.Path {
			case "/api/v1/orders":
				ordersCalls.Add(1)
			case "/api/v1/holdings":
				holdingsCalls.Add(1)
			case "/api/v1/buying-power":
				balanceCalls.Add(1)
			}
			writeEmptyEngineSnapshotResponse(w, r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	})

	ectx, _, srv, err := assembleAccountSequenceTestEngine(t, handler, officialOptions...)
	t.Cleanup(srv.Close)
	if err != nil {
		t.Fatalf("assembleEngine: %v", err)
	}
	t.Cleanup(func() { _ = ectx.Close() })
	if ectx.AccountRef != "123-45" {
		t.Fatalf("AccountRef = %q, want 123-45", ectx.AccountRef)
	}

	recovery, err := ectx.Recovery(reconcile.Options{
		Clock: clock.System(),
		Stabilise: reconcile.Stabilisation{
			Interval:    time.Millisecond,
			Required:    2,
			MaxAttempts: 2,
		},
	})
	if err != nil {
		t.Fatalf("Context.Recovery: %v", err)
	}
	report, err := recovery.Run(context.Background())
	if err != nil {
		t.Fatalf("Recovery.Run: %v", err)
	}
	if !report.Complete || !recovery.Complete() {
		t.Fatal("actual engine recovery did not complete")
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want startup discovery only", calls)
	}
	if calls := wrongHeaders.Load(); calls != 0 {
		t.Fatalf("snapshot requests with wrong account header = %d", calls)
	}
	if got := []int32{ordersCalls.Load(), holdingsCalls.Load(), balanceCalls.Load()}; got[0] != 2 || got[1] != 2 || got[2] != 2 {
		t.Fatalf("snapshot endpoint calls = %v, want [2 2 2]", got)
	}
}

func TestActualEngineRecoveryReusesTheStartupAccountSequence(t *testing.T) {
	assertActualEngineRecoveryReusesTheStartupAccountSequence(t)
}

func TestActualEngineRecoveryAcceptsAMatchingExplicitAccountSequence(t *testing.T) {
	assertActualEngineRecoveryReusesTheStartupAccountSequence(t, official.WithAccountSeq(7))
}

func TestActualEngineRecoveryStillFailsClosedOnASnapshot429(t *testing.T) {
	var accountCalls atomic.Int32
	var ordersCalls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeEngineAccountSequenceToken(w)
		case "/api/v1/accounts":
			accountCalls.Add(1)
			writeEngineAccountSequenceAccounts(w,
				`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
		case "/api/v1/orders":
			ordersCalls.Add(1)
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			http.NotFound(w, r)
		}
	})

	ectx, _, srv, err := assembleAccountSequenceTestEngine(t, handler)
	t.Cleanup(srv.Close)
	if err != nil {
		t.Fatalf("assembleEngine: %v", err)
	}
	t.Cleanup(func() { _ = ectx.Close() })
	// 예산과 백오프를 테스트가 명시한다. 기본값(5분 / 15초)을 그대로 두면 한 번 도는 데
	// 벽시계 300초가 들고, 그 300초는 정보가 아니라 비용이다 — a118이 이 스위트를 게이트에
	// 배선한 뒤로는 매 실행마다 든다.
	const backoff = reconcile.DefaultRateLimitBackoff
	// 일부러 백오프의 배수가 아닌 예산을 쓴다. 배수를 쓰면 "예산"과 "실제 소진액"이
	// 우연히 같아져서 둘을 혼동한 단언이 통과해 버린다 — 리뷰가 잡은 그 구멍이다.
	const budget = 3*backoff + 5*time.Second
	// 기대 읽기 횟수는 계약에서 유도한다. 리터럴로 굳히면 상수를 바꾼 change 가 아무
	// 신호도 받지 못한다 — 이 단언이 한 달 동안 틀린 채로 있었던 이유가 정확히 그것이다.
	// 거절된 읽기는 attempt 를 소모하지 않으므로(internal/reconcile/recovery.go 의
	// "Deliberately no attempt++") 루프 상한이 아니라 예산이 멈춘다: 예산을 다 쓸 때까지
	// budget/backoff 번 기다리고, 그다음 읽기가 초과로 실패한다.
	wantWaits := int(budget / backoff)
	wantReads := wantWaits + 1

	// 잠들지 않고 요청된 대기를 기록하는 시계. 상한을 지운 뮤테이션이 이 테스트를
	// timeout 까지 매달지 않고 오류로 끝내 준다(a102Clock 의 runaway 가드).
	clk := newA102Clock()
	recovery, err := ectx.Recovery(reconcile.Options{
		Clock: clk,
		Stabilise: reconcile.Stabilisation{
			Interval:         time.Millisecond,
			Required:         2,
			MaxAttempts:      2,
			RateLimitBackoff: backoff,
			MaxRateLimitWait: budget,
		},
	})
	if err != nil {
		t.Fatalf("Context.Recovery: %v", err)
	}
	report, err := recovery.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("Recovery.Run error = %v, want ErrRecoveryIncomplete", err)
	}
	if report.Complete || recovery.Complete() {
		t.Fatal("recovery accepted a partial snapshot after orders 429")
	}
	if !report.Snapshot.Empty() {
		t.Fatalf("partial snapshot was retained: %+v", report.Snapshot)
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want 1", calls)
	}
	if calls := int(ordersCalls.Load()); calls != wantReads {
		t.Fatalf("/api/v1/orders calls = %d, want %d (예산 %s ÷ 백오프 %s + 1)",
			calls, wantReads, budget, backoff)
	}
	// 호출 수 하나로는 "예산이 멈췄다"와 "다른 무엇이 멈췄다"가 갈리지 않는다.
	// 기대 소진액은 예산이 아니라 wantWaits × backoff 다. 마지막 한 번은 예산을 넘기므로
	// 실행되지 않고, 예산이 백오프의 배수가 아닐 때 두 값이 갈린다.
	wantWaited := time.Duration(wantWaits) * backoff
	if report.RateLimitWaits != wantWaits || report.RateLimitWaited != wantWaited {
		t.Fatalf("rate-limit 보고 = %d회 / %s, want %d회 / %s",
			report.RateLimitWaits, report.RateLimitWaited, wantWaits, wantWaited)
	}
	if got := len(clk.waits); got != wantWaits {
		t.Fatalf("실제 대기 = %d회, want %d — 예산을 넘겨 기다려서는 안 된다", got, wantWaits)
	}
	// 횟수만 세면 "더 짧게 자 놓고 장부에는 백오프만큼 적는" 버그가 통과한다.
	for i, d := range clk.waits {
		if d != backoff {
			t.Fatalf("%d번째 대기 = %s, want %s", i+1, d, backoff)
		}
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("err = %q, want 사유가 rate limit 을 지목할 것", err)
	}
}

func TestEngineRefusesAnIncompleteFirstAccountRecordBeforeOpeningTheJournal(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "blank account number",
			body: `[{"accountNo":"","accountSeq":7},{"accountNo":"later","accountSeq":8}]`,
		},
		{
			name: "zero sequence",
			body: `[{"accountNo":"first","accountSeq":0},{"accountNo":"later","accountSeq":8}]`,
		},
		{
			name: "negative sequence",
			body: `[{"accountNo":"first","accountSeq":-7},{"accountNo":"later","accountSeq":8}]`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeEngineAccountSequenceToken(w)
				case "/api/v1/accounts":
					writeEngineAccountSequenceAccounts(w, tc.body)
				default:
					http.NotFound(w, r)
				}
			})
			ectx, dir, srv, err := assembleAccountSequenceTestEngine(t, handler)
			t.Cleanup(srv.Close)
			if ectx != nil {
				_ = ectx.Close()
			}
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("assembleEngine error = %v, want ErrAccountUnresolved", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !os.IsNotExist(statErr) {
				t.Fatalf("journal was created before account refusal: %v", statErr)
			}
		})
	}
}

func TestEngineRefusesAnExplicitSequenceThatDoesNotMatchTheFirstRecord(t *testing.T) {
	for _, seq := range []int{99, -1} {
		t.Run(strconv.Itoa(seq), func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeEngineAccountSequenceToken(w)
				case "/api/v1/accounts":
					writeEngineAccountSequenceAccounts(w,
						`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
				default:
					http.NotFound(w, r)
				}
			})
			ectx, dir, srv, err := assembleAccountSequenceTestEngine(
				t, handler, official.WithAccountSeq(seq),
			)
			t.Cleanup(srv.Close)
			if ectx != nil {
				_ = ectx.Close()
			}
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("assembleEngine error = %v, want ErrAccountUnresolved", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !os.IsNotExist(statErr) {
				t.Fatalf("journal was created before sequence mismatch refusal: %v", statErr)
			}
		})
	}
}
