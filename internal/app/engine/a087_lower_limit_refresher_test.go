package engine

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type a087Admission struct {
	failAt int64
	calls  atomic.Int64
}

func (a *a087Admission) Admit(context.Context) error {
	call := a.calls.Add(1)
	if a.failAt > 0 && call == a.failAt {
		return errors.New("admission refused")
	}
	return nil
}

type a087LowerLimitReader struct {
	admission lowerLimitAdmission
	statuses  []int
	raw       official.ProtectivePriceLimits
	wires     atomic.Int64
	called    chan struct{}
}

func (r *a087LowerLimitReader) ProtectivePriceLimits(ctx context.Context, _ string) (official.ProtectivePriceLimits, error) {
	for _, status := range r.statuses {
		if err := r.admission.Admit(ctx); err != nil {
			return official.ProtectivePriceLimits{}, err
		}
		r.wires.Add(1)
		if r.called != nil {
			select {
			case r.called <- struct{}{}:
			default:
			}
		}
		if status == 200 {
			return r.raw, nil
		}
	}
	return official.ProtectivePriceLimits{}, errors.New("no successful wire")
}

type a087Positions []journal.Position

func (p a087Positions) Positions(context.Context, string) ([]journal.Position, error) { return p, nil }

type a087ExpiredPositions []journal.Position

func (p a087ExpiredPositions) Positions(ctx context.Context, _ string) ([]journal.Position, error) {
	<-ctx.Done()
	return p, nil
}

type a087Attempts bool

func (a a087Attempts) HasProtectiveExitAttempt(context.Context, string, string, string, string, []string) (bool, error) {
	return bool(a), nil
}

type a087FailingAttempts struct{ err error }

func (a a087FailingAttempts) HasProtectiveExitAttempt(context.Context, string, string, string, string, []string) (bool, error) {
	return false, a.err
}

type a087ExpiredAttempts struct{}

func (a087ExpiredAttempts) HasProtectiveExitAttempt(ctx context.Context, _ string, _ string, _ string, _ string, _ []string) (bool, error) {
	<-ctx.Done()
	return false, nil
}

func a087String(value string) *string { return &value }

func a087RefreshBoard(now time.Time, admission lowerLimitAdmission, reader lowerLimitReader) *exitLowerLimitBoard {
	return newExitLowerLimitBoard(clock.NewFake(now), reader,
		a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087Attempts(false), "acct-a087", admission)
}

func TestA087NilAdmissionMakesNoLowerLimitWire(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := a087RefreshBoard(now, nil, reader)
	if board.canRefresh() {
		t.Fatal("nil admission board가 갱신 가능 상태임")
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- board.Run(ctx) }()
	board.cycleComplete(5 * time.Second)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want canceled", err)
	}
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("nil admission wire GET = %d, want 0", got)
	}
	if _, ok := board.record("kr", "005930"); ok {
		t.Fatal("nil admission populated the lower-limit board")
	}
}

func TestA087AdmissionRejectionSkipsLowerLimitWire(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{failAt: 1}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := a087RefreshBoard(now, gate, reader)
	board.refresh(t.Context(), 5*time.Second)
	if got := gate.calls.Load(); got != 1 {
		t.Fatalf("admission calls = %d, want 1", got)
	}
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("rejected admission wire GET = %d, want 0", got)
	}
	if _, ok := board.record("kr", "005930"); ok {
		t.Fatal("rejected admission populated the lower-limit board")
	}
}

func TestA087Every401WireRequiresAdmission(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{401, 401, 200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := a087RefreshBoard(now, gate, reader)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 3 {
		t.Fatalf("401 retry wire GETs = %d, want 3", got)
	}
	if got := gate.calls.Load(); got != 3 {
		t.Fatalf("401 retry admission calls = %d, want 3", got)
	}
	if _, ok := board.record("kr", "005930"); !ok {
		t.Fatal("successful third wire did not populate the board")
	}
}

func TestA087Second401AdmissionRejectionPreventsThirdWire(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{failAt: 2}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{401, 401, 200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := a087RefreshBoard(now, gate, reader)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 1 {
		t.Fatalf("wire GETs after second admission refusal = %d, want 1", got)
	}
	if got := gate.calls.Load(); got != 2 {
		t.Fatalf("admission calls = %d, want 2", got)
	}
}

func TestA087RestartedProtectiveAttemptSkipsLowerLimitRead(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clock.NewFake(now), reader,
		a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087Attempts(true), "acct-a087", gate)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("재시작 금지 항목의 wire GET = %d, want 0", got)
	}
	if got := board.attemptedOn("kr", "005930", now); !got {
		t.Fatal("재시작 금지 항목을 메모리 판에 복원하지 못함")
	}
}

func TestA087JournalReadFailureLeavesLowerLimitBoardEmpty(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clock.NewFake(now), reader,
		a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087FailingAttempts{errors.New("journal unavailable")}, "acct-a087", gate)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("원장 오류 뒤 wire GET = %d, want 0", got)
	}
	if _, ok := board.record("kr", "005930"); ok {
		t.Fatal("원장 오류 뒤 하한가 판이 채워짐")
	}
}

func TestA087ExpiredPrefetchContextSkipsLowerLimitWire(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clock.NewFake(now), reader,
		a087ExpiredPositions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087Attempts(false), "acct-a087", gate)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	board.refresh(ctx, 5*time.Second)
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("만료된 사전 원장 읽기 뒤 하한가 wire GET = %d, want 0", got)
	}
}

func TestA087ExpiredAttemptReadSkipsLowerLimitWire(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clock.NewFake(now), reader,
		a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087ExpiredAttempts{}, "acct-a087", gate)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	board.refresh(ctx, 5*time.Second)
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("만료된 시도 원장 읽기 뒤 하한가 wire GET = %d, want 0", got)
	}
}

func TestA087LowerLimitRefresherSkipsOutsideKoreanLimitWindow(t *testing.T) {
	now := time.Date(2026, 3, 30, 7, 0, 0, 0, time.UTC) // KST 16:00 시각
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := a087RefreshBoard(now, gate, reader)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("장외 하한가 wire GET = %d, want 0", got)
	}
}

func TestA087LowerLimitRefresherReadsAtMostOneSymbolPerCycle(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200, 200}, raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clock.NewFake(now), reader, a087Positions{
		{Market: "kr", Symbol: "005930", State: "OPEN"},
		{Market: "kr", Symbol: "000660", State: "OPEN"},
	}, a087Attempts(false), "acct-a087", gate)
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 1 {
		t.Fatalf("한 주기 하한가 wire GET = %d, want 1", got)
	}
}

func TestA087LowerLimitRefresherStartsQuietForFiveMinutes(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	clk := clock.NewFake(now)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, called: make(chan struct{}, 2), raw: official.ProtectivePriceLimits{
		LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339)),
	}}
	board := newExitLowerLimitBoard(clk, reader,
		a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, a087Attempts(false), "acct-a087", gate)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- board.Run(ctx) }()

	board.cycleComplete(5 * time.Second)
	a087WaitFor(t, func() bool { return len(board.tokens) == 0 })
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("start wire GET = %d, want 0", got)
	}
	clk.Advance(exitLowerLimitQuiet - time.Second)
	board.cycleComplete(5 * time.Second)
	a087WaitFor(t, func() bool { return len(board.tokens) == 0 })
	if got := reader.wires.Load(); got != 0 {
		t.Fatalf("4m59s wire GET = %d, want 0", got)
	}
	clk.Advance(time.Second)
	board.cycleComplete(5 * time.Second)
	select {
	case <-reader.called:
	case <-time.After(time.Second):
		t.Fatal("5m token did not start a lower-limit read")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want canceled", err)
	}
}

func TestA087LowerLimitSnapshotReadWriteRace(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	board := &exitLowerLimitBoard{}
	m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "50000"}}
	proposal := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 1_000; i++ {
			board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
				lowerLimitKey("kr", "005930"): {price: "50400", timestamp: now},
			}})
		}
	}()
	for i := 0; i < 1_000; i++ {
		_ = exitLimitPrice(now, m, proposal, "50400", board)
	}
	<-done
}

func a087WaitFor(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !ready() {
		if time.Now().After(deadline) {
			t.Fatal("expected lower-limit refresher progress")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestA087PreviousDayRecordDoesNotBlockTodayRefresh(t *testing.T) {
	now := time.Date(2026, 3, 31, 1, 0, 0, 0, time.UTC)
	gate := &a087Admission{}
	reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339))}}
	board := a087RefreshBoard(now, gate, reader)
	board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{lowerLimitKey("kr", "005930"): {price: "50000", timestamp: now.AddDate(0, 0, -1)}}})
	board.refresh(t.Context(), 5*time.Second)
	if got := reader.wires.Load(); got != 1 {
		t.Fatalf("today wire GET = %d, want 1", got)
	}
	record, ok := board.record("kr", "005930")
	if !ok || record.price != "50400" {
		t.Fatalf("today record = %+v, exists %t", record, ok)
	}
}

func TestA087PreviousDayRecordFailsClosedOnAttemptOrJournalError(t *testing.T) {
	now := time.Date(2026, 3, 31, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name     string
		attempts lowerLimitAttempts
	}{
		{"today attempt", a087Attempts(true)},
		{"journal error", a087FailingAttempts{errors.New("journal unavailable")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gate := &a087Admission{}
			reader := &a087LowerLimitReader{admission: gate, statuses: []int{200}, raw: official.ProtectivePriceLimits{LowerLimitPrice: a087String("50400"), Timestamp: a087String(now.Format(time.RFC3339))}}
			board := newExitLowerLimitBoard(clock.NewFake(now), reader, a087Positions{{Market: "kr", Symbol: "005930", State: "OPEN"}}, tc.attempts, "acct-a087", gate)
			board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{lowerLimitKey("kr", "005930"): {price: "50000", timestamp: now.AddDate(0, 0, -1)}}})
			board.refresh(t.Context(), 5*time.Second)
			if got := reader.wires.Load(); got != 0 {
				t.Fatalf("wire GET = %d, want 0", got)
			}
			m := managed{position: journal.Position{Market: "kr", Symbol: "005930"}, state: journal.ExitState{Baseline: "49000"}}
			got := exitLimitPrice(now, m, exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}, "51000", board)
			if want := (exitLimitChoice{price: "51000", source: exitPriceSourceObserved}); got != want {
				t.Fatalf("choice = %#v, want %#v", got, want)
			}
		})
	}
}
