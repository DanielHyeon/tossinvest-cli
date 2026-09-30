package obs

// a092 26라운드 2차 수리(Manager 판정 2026-09-30):
//   - codex 재확인 R3: 선점 로그의 분류는 결과 enum 으로 — 모르는 결과는 선점이 아니라 원장 이상.
//   - 보이스 A #4 · B #4: 줄 자신의 event 키를 가리지 않게 원래 사건 유형은 trigger_event 로.
//   - 보이스 B #1 · #2(불변식 8): 기록 전용 입구의 로그 줄에 필드(계좌) 없음, 게이트 설명에 원문 오류(계좌 담긴 키) 없음.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const a092Acct = "99887766554"

type a092SyncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *a092SyncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *a092SyncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func a092JSONLogger(buf *a092SyncBuffer) *Logger {
	return NewLogger(LogOptions{Writer: buf, JSON: true, Clock: clock.NewFake(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))})
}

// a092Lines 는 JSON 줄마다 **모든** 키-값 쌍을 순서대로 돌려줌 — 중복 키를 보려면 map 으로 풀면 안 됨.
func a092Lines(t *testing.T, s string) [][][2]string {
	t.Helper()
	var out [][][2]string
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if line == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		if _, err := dec.Token(); err != nil {
			t.Fatalf("not JSON: %q", line)
		}
		var pairs [][2]string
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				t.Fatal(err)
			}
			var v any
			if err := dec.Decode(&v); err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(v)
			pairs = append(pairs, [2]string{k.(string), strings.Trim(string(b), `"`)})
		}
		out = append(out, pairs)
	}
	return out
}

func a092Count(pairs [][2]string, key string) int {
	n := 0
	for _, p := range pairs {
		if p[0] == key {
			n++
		}
	}
	return n
}

func a092Value(pairs [][2]string, key string) string {
	for _, p := range pairs {
		if p[0] == key {
			return p[1]
		}
	}
	return ""
}

func a092OpenJournal(t *testing.T) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path: t.TempDir() + "/j.db", Clock: clock.System(),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	return j
}

// R3: 결과 enum 이 분류함 — 모르는 결과는 EventAlertUndelivered(원장 이상), 선점 이름을 쓰지 않음.
func TestA092TheLeaseLossLogClassifiesByOutcome(t *testing.T) {
	cases := []struct {
		name  string
		res   journal.SettleResult
		event EventType
		level string
		text  string
	}{
		{"unknown", journal.SettleResult{Outcome: journal.SettleOutcome(99)}, EventAlertUndelivered, "ERROR", "unknown"},
		{"own-release", journal.SettleResult{Outcome: journal.SettleLeaseLost}, EventAlertClaimLost, "INFO", "handed back"},
		{"stolen", journal.SettleResult{Outcome: journal.SettleLeaseLost, ClaimedBy: "other"}, EventAlertClaimLost, "WARN", ""},
		{"settled", journal.SettleResult{Outcome: journal.SettleAlreadySettled}, EventAlertClaimLost, "INFO", "settled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := &a092SyncBuffer{}
			n := &Notifier{Log: a092JSONLogger(buf), Clock: clock.System()}
			n.logLeaseLost(c.res, 7, Event{Type: EventOperatingMode})
			lines := a092Lines(t, buf.String())
			if len(lines) != 1 {
				t.Fatalf("lines = %d, want 1:\n%s", len(lines), buf.String())
			}
			if got := a092Value(lines[0], "msg"); got != string(c.event) {
				t.Errorf("event = %s, want %s", got, c.event)
			}
			if got := a092Value(lines[0], "level"); got != c.level {
				t.Errorf("level = %s, want %s", got, c.level)
			}
			if c.text != "" && !strings.Contains(buf.String(), c.text) {
				t.Errorf("line lacks %q:\n%s", c.text, buf.String())
			}
		})
	}
}

// A#4 · B#4: 이 패키지의 로그 줄은 "event" 키를 한 번만 씀 — 원래 사건 유형은 trigger_event.
func TestA092DropAndNoJournalLinesKeepTheirOwnEventKey(t *testing.T) {
	buf := &a092SyncBuffer{}
	n := &Notifier{Log: a092JSONLogger(buf), Clock: clock.System()}
	e := Event{Type: EventType("exit.proposal_capped"), Key: "k1"}
	n.logNormalDrop(e, "full")                       // 버림 기록
	_ = n.RecordCritical(context.Background(), e, 0) // 원장 없음 경고(RecordCritical)
	_ = n.recordCritical(context.Background(), e, 0) // 원장 없음 경고(recordCritical)
	_ = n.notifyCritical(context.Background(), e)    // 원장 없음 경고(notifyCritical)
	n.Publisher = a092FailPublisher{}
	n.publishBestEffort(context.Background(), e, SeverityNormal) // 발행 실패
	lines := a092Lines(t, buf.String())
	if len(lines) < 5 {
		t.Fatalf("lines = %d, want at least 5:\n%s", len(lines), buf.String())
	}
	for i, l := range lines {
		if a092Count(l, FieldEvent) != 1 {
			t.Errorf("line %d carries %d %q keys — a JSON reader keeps the last and loses the line's own event:\n%v",
				i, a092Count(l, FieldEvent), FieldEvent, l)
		}
		if a092Count(l, FieldSeverity) != 1 {
			t.Errorf("line %d carries %d %q keys:\n%v", i, a092Count(l, FieldSeverity), FieldSeverity, l)
		}
		if a092Value(l, FieldEvent) != a092Value(l, "msg") {
			t.Errorf("line %d event = %s, msg = %s", i, a092Value(l, FieldEvent), a092Value(l, "msg"))
		}
	}
}

type a092FailPublisher struct{}

func (a092FailPublisher) Publish(context.Context, Notification) error { return errors.New("503") }

// B#1: 기록 전용 입구 세 메서드의 로그 줄에 필드(계좌 원문)가 없음.
func TestA092RecordOnlyLinesCarryNoFields(t *testing.T) {
	j := a092OpenJournal(t)
	t.Cleanup(func() { _ = j.Close() })
	buf := &a092SyncBuffer{}
	n := &Notifier{Log: a092JSONLogger(buf), Journal: j, Clock: clock.System()}
	r := RecordOnly{N: n, Relay: NewNormalRelay(n, 4)}
	ctx := context.Background()
	fields := map[string]any{FieldAccount: a092Acct}
	_ = r.Notify(ctx, Event{Type: EventOperatingMode, Key: "c1", Fields: fields})                // critical
	_ = r.Notify(ctx, Event{Type: EventType("exit.proposal_capped"), Key: "n1", Fields: fields}) // normal
	_ = r.AnnounceOperatingMode(ctx, journal.ModeEntryBlocked, journal.OperatingModeRecord{
		ID: "t1", AccountRef: a092Acct, Mode: journal.ModeNormal, Actor: journal.ModeActorOperator, Cause: "checked"})
	if strings.Contains(buf.String(), a092Acct) {
		t.Errorf("a record-only log line carries the account number (invariant 8):\n%s", buf.String())
	}
	if len(a092Lines(t, buf.String())) < 3 {
		t.Fatalf("arrangement: expected three event lines:\n%s", buf.String())
	}
}

// B#2: 기록이 실패해도 게이트 설명에 원문 오류(계좌를 담은 사건 키)가 없음 — 상태 출력이 어디서나 읽는 칸.
func TestA092AFailedRecordKeepsTheKeyOutOfTheGate(t *testing.T) {
	j := a092OpenJournal(t)
	_ = j.Close() // 닫힌 원장 — 기록이 반드시 실패
	gate := execgw.NewEntryGate(clock.System(), map[execgw.RequiredQuery]time.Duration{})
	n := &Notifier{Journal: j, Gate: gate, Clock: clock.System()}
	err := RecordOnly{N: n}.AnnounceOperatingMode(context.Background(), journal.ModeEntryBlocked, journal.OperatingModeRecord{
		ID: "t2", AccountRef: a092Acct, Mode: journal.ModeNormal, Actor: journal.ModeActorOperator, Cause: "checked"})
	if err == nil {
		t.Fatal("arrangement: the record must fail on a closed journal")
	}
	detail, ok := gate.Blocks()[execgw.ReasonAlertUndelivered]
	if !ok {
		t.Fatal("a failed record did not latch")
	}
	if strings.Contains(detail, a092Acct) {
		t.Errorf("the gate description carries the account number: %q", detail)
	}
}
