package engine

// a133_redelivery_carries_no_account_internal_test.go — P4(엔진 생산의 outbox 재전송, alertDeliverer)가 경계를 지남(a133 design D1 · D6-2).
//
// 저장 행의 Title · Body 에 계좌를 심고, 전달기가 실제 *obs.Ntfy 로 httptest 서버에 보낸 바이트에 계좌가 없음을 잼.
// 전달 판정(성공 → DELIVERED · 거절 한도 → 진입 래치 + 모드 승격)은 a124 와 같음을 같은 행으로 잼.

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

const (
	a133EngAccount = "987-65-A133ENGSENTINEL"
	a133EngDigits  = "98765432134"
	a133EngMask    = "*******2134"
	a133EngPos     = "pos-aabbccddeeff001122334455"
	a133EngTopic   = "tossos-a133engine"
)

func init() { redact.Register(a133EngAccount, a133EngDigits) }

type a133EngServer struct {
	mu   sync.Mutex
	seen []string // 요청마다 URL · 본문 · 모든 헤더를 이은 바이트
	path []string
}

func a133EngNtfy(t *testing.T, status int) (*obs.Ntfy, *a133EngServer) {
	t.Helper()
	s := &a133EngServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var b strings.Builder
		b.WriteString(r.URL.String() + "\n" + string(body) + "\n")
		for k, vs := range r.Header {
			b.WriteString(k + ": " + strings.Join(vs, ",") + "\n")
		}
		s.mu.Lock()
		s.seen = append(s.seen, b.String())
		s.path = append(s.path, r.URL.Path+"?"+r.URL.RawQuery)
		s.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return &obs.Ntfy{BaseURL: srv.URL, Topic: a133EngTopic, HTTPClient: srv.Client()}, s
}

func (s *a133EngServer) mustBeClean(t *testing.T, want int) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.seen) != want {
		t.Fatalf("requests reaching the server = %d, want %d", len(s.seen), want)
	}
	for i, all := range s.seen {
		for _, sentinel := range []string{a133EngAccount, a133EngDigits, a133EngMask, a133EngPos} {
			if strings.Contains(all, sentinel) {
				t.Fatalf("request %d leaked %q:\n%s", i, sentinel, all)
			}
		}
		if s.path[i] != "/"+a133EngTopic+"?" {
			t.Fatalf("request %d URL = %q, want only the configured topic", i, s.path[i])
		}
	}
}

func a133EngRow(t *testing.T, f *a124Fixture) int64 {
	t.Helper()
	id, err := f.j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: "a133-redelivery", Type: "execgw.order_unresolved", Severity: "critical",
		Title:   "UNRESOLVED_IN_DOUBT " + a133EngAccount,
		Body:    "attempt for account " + a133EngDigits + " (" + a133EngMask + ") position " + a133EngPos,
		Payload: `{"account":"` + a133EngAccount + `"}`,
	})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

func TestA133TheDelivererRedeliversWithoutTheAccount(t *testing.T) {
	pub, srv := a133EngNtfy(t, http.StatusOK)
	f := a124Setup(t, pub)
	id := a133EngRow(t, f)
	f.cycles(t, 1)
	srv.mustBeClean(t, 1)
	if got := a124State(t, f, id); got != journal.AlertDelivered {
		t.Fatalf("state after a delivered redelivery = %s, want DELIVERED", got)
	}
	if f.latched() {
		t.Fatal("a delivered alert latched the entry gate")
	}
}

func TestA133TheDelivererStillLatchesOnRefusal(t *testing.T) {
	pub, srv := a133EngNtfy(t, http.StatusServiceUnavailable)
	f := a124Setup(t, pub)
	id := a133EngRow(t, f)
	f.cycles(t, alertAttemptLimit)
	srv.mustBeClean(t, alertAttemptLimit)
	if !f.latched() {
		t.Fatalf("not latched after %d refused redeliveries", alertAttemptLimit)
	}
	if got := f.mode(t); got != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s, want %s", got, journal.ModeEntryBlocked)
	}
	if got := f.attempts(t, id); got != alertAttemptLimit {
		t.Fatalf("attempts = %d, want %d", got, alertAttemptLimit)
	}
}
