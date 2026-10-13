package obs_test

// a133_alert_redact_test.go — 시스템 밖으로 나가는 알림(ntfy)에 계좌 정보가 실리지 않음(a133 design D6).
//
// 판정은 언제나 **httptest ntfy 서버가 받은 바이트**(본문 · 모든 헤더 값 · URL)로 함 — 가짜 Publisher 가 받은 구조체로 재면
// 경계(Ntfy.Publish) 안의 치환을 못 봄. 실 ntfy 전송은 없음(모든 서버가 httptest).

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

// 센티널 — 등록 원문 · 등록 숫자열 · 그 마스크 · 계좌 유도 id 둘. a133Unregistered 는 **등록하지 않은** 계좌 값이라
// 값 치환으로는 안 지워짐 — 필드 이름 버림(renderFields)만이 지움(변이 「필드 버림 삭제」를 잡는 표식).
const (
	a133Account      = "987-65-A133SENTINEL"
	a133Digits       = "98765432133"
	a133Mask         = "*******2133"
	a133PositionID   = "pos-00112233445566778899aabb"
	a133ModeID       = "mode-ffeeddccbbaa998877665544"
	a133Unregistered = "UNREGISTERED-ACCT-77"
	a133Topic        = "tossos-a133topic"
)

func init() { redact.Register(a133Account, a133Digits) }

// a133Request 는 서버가 받은 요청 하나의 바이트임.
type a133Request struct {
	url     string
	rawPath string
	query   string
	body    string
	headers http.Header
}

// all 은 요청이 운반한 모든 텍스트(URL · 본문 · 모든 헤더 이름과 값)를 한 문자열로 모음 — 누출 판정의 모집단.
func (r a133Request) all() string {
	var b strings.Builder
	b.WriteString(r.url + "\n" + r.body + "\n")
	for k, vs := range r.headers {
		b.WriteString(k + ": " + strings.Join(vs, ",") + "\n")
	}
	return b.String()
}

type a133Server struct {
	mu   sync.Mutex
	seen []a133Request
	hit  chan struct{}
}

// a133NtfyServer 는 status 를 돌려주는 httptest ntfy 와 그것을 겨눈 실제 *obs.Ntfy 를 만듦.
func a133NtfyServer(t *testing.T, status int) (*obs.Ntfy, *a133Server) {
	t.Helper()
	s := &a133Server{hit: make(chan struct{}, 64)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.seen = append(s.seen, a133Request{
			url: r.URL.String(), rawPath: r.URL.Path, query: r.URL.RawQuery,
			body: string(body), headers: r.Header.Clone(),
		})
		s.mu.Unlock()
		w.WriteHeader(status)
		s.hit <- struct{}{}
	}))
	t.Cleanup(srv.Close)
	return &obs.Ntfy{BaseURL: srv.URL, Topic: a133Topic, HTTPClient: srv.Client()}, s
}

func (s *a133Server) requests() []a133Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]a133Request(nil), s.seen...)
}

func (s *a133Server) wait(t *testing.T) {
	t.Helper()
	select {
	case <-s.hit:
	case <-time.After(5 * time.Second):
		t.Fatal("the ntfy server never received the notification")
	}
}

// a133Leaks 는 요청 바이트에 남은 센티널 목록임(등록 안 된 값 포함).
func a133Leaks(r a133Request) []string {
	all := r.all()
	var found []string
	for _, s := range []string{a133Account, a133Digits, a133Mask, a133PositionID, a133ModeID, a133Unregistered} {
		if strings.Contains(all, s) {
			found = append(found, s)
		}
	}
	return found
}

// a133MustBeClean 은 받은 요청마다 누출 0 · URL 이 /<topic> 이고 쿼리 없음을 단언함. 요청 0 이면 아무것도 안 잰 것이라 실패.
func a133MustBeClean(t *testing.T, reqs []a133Request) {
	t.Helper()
	if len(reqs) == 0 {
		t.Fatal("no request reached the server — the scan measured nothing")
	}
	for i, r := range reqs {
		if l := a133Leaks(r); len(l) != 0 {
			t.Fatalf("request %d leaked %v:\n%s", i, l, r.all())
		}
		if r.rawPath != "/"+a133Topic || r.query != "" {
			t.Fatalf("request %d URL carries more than the configured topic: path=%q query=%q", i, r.rawPath, r.query)
		}
		// 계좌 키 판정은 a131 규칙 그대로 — 어떤 줄의 키든 IsAccountKey 이면 필드 버림이 빠진 것(리뷰 F-4: 좁힌 버림 변이).
		for _, line := range strings.Split(r.body, "\n") {
			if key, _, ok := strings.Cut(line, ":"); ok && redact.IsAccountKey(key) {
				t.Fatalf("request %d body kept an account field line: %q", i, line)
			}
		}
	}
}

// a133Fields 는 계좌 키 필드 · 포지션 id · 운영 정보 필드를 함께 담은 문맥임.
func a133Fields() map[string]any {
	return map[string]any{
		obs.FieldAccount:   a133Unregistered,
		"account_ref":      a133Unregistered + "-ref",
		"account_no":       a133Unregistered + "-no",
		"masked_account":   a133Unregistered + "-masked",
		"account_seq":      a133Unregistered + "-seq",
		"position_id":      a133PositionID,
		obs.FieldSymbol:    "005930",
		obs.FieldReason:    "BROKER_STATE_UNKNOWN",
		obs.FieldAttemptID: "a1b2c3d4e5f60718293a4b5c6d7e8f90",
		obs.FieldQuantity:  7,
	}
}

// a133MustKeepMeaning 은 운영자가 조치에 쓰는 정보가 본문에 남았음을 단언함(design D0).
func a133MustKeepMeaning(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{
		"symbol: 005930", "reason: BROKER_STATE_UNKNOWN",
		"attempt_id: a1b2c3d4e5f60718293a4b5c6d7e8f90", "quantity: 7", "position_id: pos-[redacted]",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("operator context %q was lost:\n%s", want, body)
		}
	}
}

// TestA133NtfyBoundaryScrubsEveryByte — 경계 자체: Title · Body · Tags · Type 에 심은 센티널이 받은 바이트 어디에도 없음.
func TestA133NtfyBoundaryScrubsEveryByte(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	err := pub.Publish(context.Background(), obs.Notification{
		Type:     obs.EventType("order.in_doubt." + a133PositionID),
		Severity: obs.SeverityCritical,
		Title:    "IN_DOUBT on 005930 for " + a133Account,
		Body:     "attempt a1b2 for account " + a133Digits + " (" + a133Mask + ") position " + a133PositionID + " mode " + a133ModeID,
		Tags:     []string{"warning", "acct-" + a133Account},
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	reqs := srv.requests()
	a133MustBeClean(t, reqs)
	if !strings.Contains(reqs[0].headers.Get("Title"), "IN_DOUBT on 005930 for [account]") {
		t.Fatalf("title lost its non-account text: %q", reqs[0].headers.Get("Title"))
	}
	if !strings.Contains(reqs[0].body, "attempt a1b2") || !strings.Contains(reqs[0].headers.Get("Tags"), "warning") {
		t.Fatalf("non-account content changed: body=%q tags=%q", reqs[0].body, reqs[0].headers.Get("Tags"))
	}

	// 우선순위를 고정한 알림(Heartbeat 꼴, P6)도 같은 경계를 지남(구현 리뷰 I-5).
	if err := pub.Publish(context.Background(), obs.Notification{
		Type: obs.EventEngineHeartbeat, Severity: obs.SeverityNormal, Priority: 1,
		Title: "tossos engine alive " + a133Account, Body: "status " + a133Digits + " " + a133PositionID,
	}); err != nil {
		t.Fatalf("Publish(priority 1): %v", err)
	}
	reqs = srv.requests()
	a133MustBeClean(t, reqs)
	if len(reqs) != 2 || reqs[1].headers.Get("Priority") != "1" {
		t.Fatalf("the pinned-priority notification did not arrive as priority 1: %d requests", len(reqs))
	}
}

// TestA133CriticalFieldsLeaveNoAccount — P2(실시간 critical): 필드의 계좌 키 줄이 없고 포지션 id 는 가려지며 운영 정보는 남음.
func TestA133CriticalFieldsLeaveNoAccount(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	n, j, _ := newNotifier(t, pub)
	if err := n.Notify(context.Background(), obs.Event{
		Type:   obs.EventOrderInDoubt,
		Title:  "IN_DOUBT " + a133Account,
		Body:   "account " + a133Digits + " masked " + a133Mask,
		Fields: a133Fields(),
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	reqs := srv.requests()
	a133MustBeClean(t, reqs)
	a133MustKeepMeaning(t, reqs[0].body)
	if left, err := j.UndeliveredCount(context.Background()); err != nil || left != 0 {
		t.Fatalf("delivered alert still pending: %d %v", left, err)
	}
}

// TestA133BestEffortPathLeavesNoAccount — P1(일반 등급 동기 최선 발송).
func TestA133BestEffortPathLeavesNoAccount(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	n, _, _ := newNotifier(t, pub)
	if err := n.Notify(context.Background(), obs.Event{
		Type: obs.EventType("a133.normal_grade"), Title: "fill " + a133Account, Body: a133Mask, Fields: a133Fields(),
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	reqs := srv.requests()
	a133MustBeClean(t, reqs)
	a133MustKeepMeaning(t, reqs[0].body)
}

// TestA133NormalRelayLeavesNoAccount — P3(a092 일반 등급 이관).
func TestA133NormalRelayLeavesNoAccount(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	n, _, _ := newNotifier(t, pub)
	relay := obs.NewNormalRelay(n, 4)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = relay.Run(ctx); close(done) }()
	relay.Offer(obs.Event{Type: obs.EventType("a133.normal_grade"), Title: a133Account, Body: a133Digits, Fields: a133Fields()})
	srv.wait(t)
	cancel()
	<-done
	reqs := srv.requests()
	a133MustBeClean(t, reqs)
	a133MustKeepMeaning(t, reqs[0].body)
}

// TestA133FlushRedeliveryLeavesNoAccount — P5(저장 행 재전송): 저장된 Title · Body 에 계좌가 있어도 나갈 때 치환됨.
func TestA133FlushRedeliveryLeavesNoAccount(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	n, j, _ := newNotifier(t, pub)
	if _, err := j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: "a133-flush|" + a133Account, Type: string(obs.EventOrderUnresolved), Severity: string(obs.SeverityCritical),
		Title: "UNRESOLVED " + a133Account, Body: "detail " + a133Digits + " " + a133Mask + " " + a133PositionID,
		Payload: `{"account":"` + a133Account + `"}`,
	}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	delivered, remaining, err := n.Flush(context.Background())
	if err != nil || delivered != 1 || remaining != 0 {
		t.Fatalf("Flush = %d, %d, %v", delivered, remaining, err)
	}
	reqs := srv.requests()
	a133MustBeClean(t, reqs)
	// 비우지 않고 치환해서 보냈음(구현 리뷰 I-7) — 빈 제목 · 본문도 센티널 없음은 통과하므로 내용을 확인.
	if reqs[0].headers.Get("Title") != "UNRESOLVED [account]" || !strings.Contains(reqs[0].body, "pos-[redacted]") ||
		!strings.Contains(reqs[0].body, "detail [account] [account]") {
		t.Fatalf("the stored alert was not sent redacted: title=%q body=%q", reqs[0].headers.Get("Title"), reqs[0].body)
	}
}

// TestA133DeliveryJudgementIsUnchanged — 계좌를 심은 critical 알림도 거절 3회면 진입이 전달 실패로 잠기고 행은 PENDING 3 시도로 남음.
func TestA133DeliveryJudgementIsUnchanged(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusServiceUnavailable)
	n, j, gate := newNotifier(t, pub)
	if err := n.Notify(context.Background(), obs.Event{
		Type: obs.EventOrderUnresolved, Title: "UNRESOLVED " + a133Account, Body: a133Digits, Fields: a133Fields(),
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	reqs := srv.requests()
	if len(reqs) != 3 {
		t.Fatalf("publish attempts reaching the server = %d, want 3", len(reqs))
	}
	a133MustBeClean(t, reqs)
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonAlertUndelivered {
		t.Fatalf("entries must be blocked after the spent budget, got %v", rejected)
	}
	pending, err := j.PendingAlerts(context.Background(), 0)
	if err != nil || len(pending) != 1 || pending[0].Attempts != 3 {
		t.Fatalf("pending = %+v, %v", pending, err)
	}
}

// TestA133TheBoundaryCoversEveryTextField — 리뷰 F-3: Notification 에 문자열 칸이 늘면 이 시험이 실패해 경계 목록을 다시 보게 함.
// 지금의 문자열 칸 전부(Type · Title · Body · Tags)에 센티널을 심어 redactForTransport 가 지우는지 리플렉션으로 잼.
func TestA133TheBoundaryCoversEveryTextField(t *testing.T) {
	var textFields, allFields []string
	typ := reflect.TypeOf(obs.Notification{})
	for i := 0; i < typ.NumField(); i++ {
		allFields = append(allFields, typ.Field(i).Name)
	}
	// 칸 전체를 고정함 — map · 구조체 · []byte 칸이 늘어도 문자열 칸 목록만 보면 조용히 통과하므로(구현 리뷰 I-4).
	if got := strings.Join(allFields, ","); got != "Type,Severity,Title,Body,Tags,Priority" {
		t.Fatalf("Notification fields = %s — a new field may be sent by Ntfy.Publish; make redactForTransport cover it, then update this list", got)
	}
	msg := reflect.New(typ).Elem()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch {
		case f.Type.Kind() == reflect.String:
			msg.Field(i).SetString(a133Account + " " + a133PositionID)
			textFields = append(textFields, f.Name)
		case f.Type.Kind() == reflect.Slice && f.Type.Elem().Kind() == reflect.String:
			msg.Field(i).Set(reflect.ValueOf([]string{a133Account, a133Mask}))
			textFields = append(textFields, f.Name)
		}
	}
	if got := strings.Join(textFields, ","); got != "Type,Severity,Title,Body,Tags" {
		t.Fatalf("Notification text fields = %s — a new field may be sent by Ntfy.Publish; make redactForTransport cover it, then update this list", got)
	}
	out := reflect.ValueOf(obs.RedactForTransportForTest(msg.Interface().(obs.Notification)))
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if name == "Severity" {
			continue // 등급은 우선순위 판정 입력이고 헤더 문자열로 나가지 않음(priorityFor) — 치환 대상 아님
		}
		if v := fmt.Sprint(out.Field(i).Interface()); strings.Contains(v, a133Account) || strings.Contains(v, a133PositionID) || strings.Contains(v, a133Mask) {
			t.Fatalf("redactForTransport left account text in %s: %q", name, v)
		}
	}
}

// TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags — 리뷰 F-10: 계좌가 아닌 8자리 이상 숫자(주문 번호 · 원화 금액 · 유닉스 ms)는
// 그대로이고, 호출자가 넘긴 Tags 슬라이스는 고쳐 쓰지 않음(deliver 는 같은 msg 로 재시도함).
func TestA133TheBoundaryKeepsNonAccountNumbersAndTheCallersTags(t *testing.T) {
	pub, srv := a133NtfyServer(t, http.StatusOK)
	tags := []string{"acct-" + a133Account}
	body := "order 1234567890123 notional 150000000 KRW at 1760000000000"
	if err := pub.Publish(context.Background(), obs.Notification{Type: obs.EventOrderInDoubt, Title: "t", Body: body, Tags: tags}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if got := srv.requests()[0].body; got != body {
		t.Fatalf("non-account numbers changed: %q", got)
	}
	if tags[0] != "acct-"+a133Account {
		t.Fatalf("the caller's tags were rewritten in place: %q", tags[0])
	}
}

// TestA133PublishStillReportsBuildAndTransportFailures — Ntfy.Publish 의 B4(요청 구성 실패) · B8(전송 실패)가 경계 뒤에도 오류임
// (전달 판정 입력 불변, 리뷰 F-9 — BTM 에 이름 있는 시험).
func TestA133PublishStillReportsBuildAndTransportFailures(t *testing.T) {
	bad := &obs.Ntfy{BaseURL: "http://bad\x7fhost", Topic: a133Topic}
	if err := bad.Publish(context.Background(), obs.Notification{Title: a133Account, Body: a133Digits}); err == nil ||
		!strings.Contains(err.Error(), "building the ntfy request") {
		t.Fatalf("a request that cannot be built must be an error, got %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	gone := &obs.Ntfy{BaseURL: url, Topic: a133Topic, Timeout: time.Second}
	if err := gone.Publish(context.Background(), obs.Notification{Title: a133Account, Body: a133Digits}); err == nil ||
		!strings.Contains(err.Error(), "publishing to ntfy") {
		t.Fatalf("a transport failure must be an error, got %v", err)
	}
}
