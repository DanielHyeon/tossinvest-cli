package engine_test

// a124 tasks 2.6 (b) · (c) — 판정하는 실행자가 손절 쪽(exit 관측 사이클)을 얼마나 늦추는가.
//
// a124 는 실행자에 원장 쓰기(정산의 가산 읽기 · 승격)와 선택 쿼리의 정렬을 더한다. 원장 연결은 하나라
// (`journal.go:174`) 그 몫은 exit 사이클과 같은 연결을 기다린다 — 「exit 체류 불변」은 주장하지 않는다
// (design D7, X4). 대신 **잰다**:
//
//	(b) 실행자의 트랜잭션 **안**에 지연을 넣어(연결을 쥔 채) exit 사이클 체류를 잰다. 수락선은 같은 일을 한
//	    무실행자 기준선 + a098 의 **고정** 여유 `a098ExitCycleDwellMargin`(AA3) — 측정값을 따라 늘어나는
//	    예산이 아니다. 여유보다 큰 지연을 넣은 대조에서는 **넘어야** 한다 — 계측기가 눈멀지 않았다는 증거.
//	(c) PENDING P = 10 · 1 000 · 10 000 에서 선택 쿼리의 연결 점유와, 그 backlog 위를 도는 실행자 옆의
//	    exit 체류(같은 고정 여유). 넘으면 멈추고 보고한다(대안 = 가산 색인 = 스키마 변경, Manager 결정 — AB1).
//
// 지연 주입은 원장 스키마 밖의 **시험 전용 트리거**다(둘째 연결로 같은 파일에 만든다). SQLite 트리거
// 본문에는 CTE 를 못 쓰므로 교차 조인의 행 수로 CPU 시간을 조절하고, 목표 지연에 맞게 먼저 잰다.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a124JudgedTransport 는 exit 사이클의 알림은 즉답하고, 실행자의 backlog 행은 즉시 실패시킨다 —
// 실행자가 매 행 실패 기록 · 반납 · 판정 트랜잭션을 연달아 돌게 한다.
type a124JudgedTransport struct {
	mu      sync.Mutex
	entered chan struct{}
	opened  bool
	other   int
}

func (p *a124JudgedTransport) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if string(n.Type) != a098BacklogEvent {
		p.other++
		return nil
	}
	if !p.opened {
		p.opened = true
		close(p.entered)
	}
	return errors.New("the backlog transport is down")
}

// a124SideHandle 은 같은 원장 파일에 여는 둘째 연결 — 시험 전용 트리거를 만들고 지연을 잰다.
func a124SideHandle(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(10000)")
	if err != nil {
		t.Fatalf("side handle: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// a124DelayRows 는 교차 조인 `count(*)` 가 target 에 가깝게 걸리는 행 수를 정한다. 비용은 행 수의 제곱에
// 비례하므로 한 번 재고 외삽한 뒤, 그 행 수로 한 번 더 재서 실제 값을 돌려준다(기록용).
func a124DelayRows(t *testing.T, db *sql.DB, target time.Duration) (int, time.Duration) {
	t.Helper()
	probe := func(rows int) time.Duration {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("delay fill: %v", err)
		}
		for _, stmt := range []string{`CREATE TABLE IF NOT EXISTS a124_delay (n INTEGER)`, `DELETE FROM a124_delay`} {
			if _, err := tx.Exec(stmt); err != nil {
				t.Fatalf("delay table: %v", err)
			}
		}
		ins, err := tx.Prepare(`INSERT INTO a124_delay VALUES (?)`)
		if err != nil {
			t.Fatalf("delay prepare: %v", err)
		}
		for i := 0; i < rows; i++ {
			if _, err := ins.Exec(i); err != nil {
				t.Fatalf("delay insert: %v", err)
			}
		}
		_ = ins.Close()
		if err := tx.Commit(); err != nil {
			t.Fatalf("delay commit: %v", err)
		}
		start := time.Now()
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM a124_delay a, a124_delay b`).Scan(&n); err != nil {
			t.Fatalf("delay probe: %v", err)
		}
		return time.Since(start)
	}
	// 공유 기계에서는 CPU 시간이 흔들린다(전체 스위트가 패키지를 병렬로 돈다). 그래서 한 번 외삽하고 끝내지 않고,
	// 세 번 잰 중앙값으로 목표의 0.7~1.5 배에 들 때까지 최대 여섯 번 다시 맞춘다(첫 게이트 판에서 한 번 재고 끝낸
	// 보정이 ½ 배 밖으로 떨어져 빨개졌다 — 2026-09-28).
	median := func(rows int) time.Duration {
		d := []time.Duration{probe(rows), probe(rows), probe(rows)}
		sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
		return d[1]
	}
	rows := 1000
	actual := median(rows)
	for i := 0; i < 6 && (actual < target*7/10 || actual > target*3/2); i++ {
		rows = int(float64(rows) * math.Sqrt(float64(target)/float64(actual)))
		if rows < 10 {
			rows = 10
		}
		actual = median(rows)
	}
	t.Logf("delay calibration: %d rows → %v per trigger (target %v)", rows, actual, target)
	return rows, actual
}

// a124SeedBacklogFast 는 backlog 행 n 개를 둘째 연결의 트랜잭션 하나로 넣는다 — EnqueueAlert 한 행마다
// fsync 하면 10 000 행이 분 단위라 측정이 게이트를 잡아먹는다. 행 모양은 a098SeedBacklog 와 같고, attempts 를
// 정할 수 있다 — 한도 − 1 로 넣으면 실행자의 **첫 실패가 곧 판정**이라 측정 창에 판정 · 승격이 든다(codex I2).
func a124SeedBacklogFast(t *testing.T, db *sql.DB, n, attempts int) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("seed begin: %v", err)
	}
	ins, err := tx.Prepare(`INSERT INTO alert_outbox (event_key, event_type, severity, title, body, state, attempts, created_at)
		VALUES (?, ?, 'critical', '밀린 알림', '실행자가 보낼 행', 'PENDING', ?, ?)`)
	if err != nil {
		t.Fatalf("seed prepare: %v", err)
	}
	now := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for i := 0; i < n; i++ {
		if _, err := ins.Exec(fmt.Sprintf("%s|fast|%d", a098BacklogEvent, i), a098BacklogEvent, attempts, now); err != nil {
			t.Fatalf("seed insert: %v", err)
		}
	}
	_ = ins.Close()
	if err := tx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
}

// a124InjectDelay 는 실행자의 쓰기에만 지연을 건다 — backlog 행의 UPDATE(실패 기록 · 반납)와 이 사유의 모드 승격.
// exit 사이클의 자기 알림 행은 event_type 이 달라 걸리지 않는다. 트리거는 지연 **뒤** 발화 시각을 남긴다 —
// 측정 창과 겹쳤는지를 사후에 확인한다(codex I2: 판정 · 승격이 측정 창 밖이면 그 몫을 안 잰 것이다).
func a124InjectDelay(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS a124_fired (kind TEXT, at REAL)`,
		`CREATE TRIGGER a124_delay_outbox BEFORE UPDATE ON alert_outbox WHEN OLD.event_type = '` + a098BacklogEvent + `'
		   BEGIN SELECT count(*) FROM a124_delay a, a124_delay b;
		         INSERT INTO a124_fired VALUES ('outbox', (julianday('now') - 2440587.5) * 86400.0); END`,
		`CREATE TRIGGER a124_delay_mode BEFORE INSERT ON operating_modes WHEN NEW.cause = '` +
			journal.ModeTriggerCriticalAlertUndelivered + `'
		   BEGIN SELECT count(*) FROM a124_delay a, a124_delay b;
		         INSERT INTO a124_fired VALUES ('mode', (julianday('now') - 2440587.5) * 86400.0); END`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("delay trigger: %v", err)
		}
	}
}

// a124Fires 는 [from, to] 창 안에서 발화한 트리거 수와 전체 수를 종류별로 센다.
func a124Fires(t *testing.T, db *sql.DB, kind string, from, to time.Time) (inside, total int) {
	t.Helper()
	lo, hi := float64(from.UnixNano())/1e9, float64(to.UnixNano())/1e9
	if err := db.QueryRow(`SELECT count(*) FROM a124_fired WHERE kind = ? AND at BETWEEN ? AND ?`, kind, lo, hi).Scan(&inside); err != nil {
		t.Fatalf("fires: %v", err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM a124_fired WHERE kind = ?`, kind).Scan(&total); err != nil {
		t.Fatalf("fires: %v", err)
	}
	return inside, total
}

type a124Measurement struct {
	dwell        time.Duration
	cycle        engine.ExitCycle
	outboxInside int
	outboxTotal  int
	modeInside   int
	modeTotal    int
}

// a124MeasureCycleBesideAJudgingExecutor 는 critical 알림을 한 번 내는 exit 사이클 하나를 잰다.
// withExecutor 면 그 옆에서 판정하는 실행자가 backlog 를 돈다(한도 − 1 로 시드 → 첫 실패부터 판정 · 승격,
// 판정 트랜잭션에 delay 만큼 지연). noPublisher 면 실행자에게 전송 수단이 없다(D3 변형).
func a124MeasureCycleBesideAJudgingExecutor(t *testing.T, withExecutor bool, delay time.Duration, backlog int, noPublisher bool) a124Measurement {
	t.Helper()
	notifier := &obs.Notifier{}
	ladder := exitpolicy.DefaultLadderPolicy()
	ladder.Rungs[0] = exitpolicy.Rung{TargetPct: "0.5", StopPct: "0", PartialRatio: "1"}
	ladder.PolicyDigest = ""
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		opts.Ladder = &ladder
		opts.Alerts = notifier
	})
	transport := &a124JudgedTransport{entered: make(chan struct{})}
	*notifier = *a098Notifier(h.journal, h.gate, transport, h.clk)

	p := h.entry("005930", "10", "10000", "9800", "10000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: "default_v1",
		EntryPrice: "10000", InitialStop: "9800",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 10100)

	var side *sql.DB
	stopExecutor := func() {}
	if withExecutor {
		side = a124SideHandle(t, h.dbPath)
		if delay > 0 {
			// 주입한 지연이 실제로 목표 근처인지 — 잰 값을 버리지 않는다(codex 2회차 R3).
			if _, actual := a124DelayRows(t, side, delay); actual < delay/2 || actual > 2*delay {
				t.Fatalf("calibrated delay %v is outside [%v, %v] for target %v — the injected load is not what the test claims",
					actual, delay/2, 2*delay, delay)
			}
		} else {
			a124DelayRows(t, side, time.Microsecond)
		}
		a124InjectDelay(t, side)
		a124SeedBacklogFast(t, side, backlog, obs.DefaultCriticalAttempts-1)
		execNotifier := notifier
		if noPublisher {
			execNotifier = &obs.Notifier{Journal: h.journal, Gate: h.gate, AccountRef: exitAccount, Clock: h.clk}
		}
		// 실제 시계 — 실행자가 측정 동안 스스로 돈다(주기 대기 2s 는 한 배치보다 길다).
		aux, err := (&engine.Context{Journal: h.journal, Entry: h.gate, Notifier: execNotifier, AccountRef: exitAccount}).
			AlertDeliverer(clock.System())
		if err != nil {
			t.Fatalf("AlertDeliverer: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { defer close(done); _ = aux.Run(ctx) }()
		// 표본마다 실행자를 **이 함수 안에서** 멈추고 기다린다 — 앞 표본의 실행자가 뒤 표본을 오염시키지 않게
		// (codex 2회차 R3). Cleanup 은 조기 종료(Fatal) 때의 안전망이다.
		stopExecutor = func() {
			cancel()
			select {
			case <-done:
			case <-time.After(60 * time.Second):
				t.Error("the executor did not stop")
			}
		}
		t.Cleanup(stopExecutor)
		if noPublisher {
			deadline := time.Now().Add(20 * time.Second)
			for {
				if _, total := a124Fires(t, side, "outbox", time.Time{}, time.Now().Add(time.Hour)); total > 0 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("the executor never wrote a backlog row")
				}
				time.Sleep(time.Millisecond)
			}
		} else {
			select {
			case <-transport.entered:
			case <-time.After(20 * time.Second):
				t.Fatal("the executor never reached a backlog publish")
			}
		}
	}

	type measured struct {
		cycle      engine.ExitCycle
		start, end time.Time
	}
	out := make(chan measured, 1)
	go func() {
		start := time.Now()
		cycle := h.observer.ObserveOnce(context.Background())
		out <- measured{cycle, start, time.Now()}
	}()
	var got measured
	select {
	case got = <-out:
	case <-time.After(60 * time.Second):
		t.Fatal("the exit cycle did not finish within 60s beside the judging executor")
	}
	if got.cycle.Err != nil {
		t.Fatalf("exit cycle: %v", got.cycle.Err)
	}
	transport.mu.Lock()
	other := transport.other
	transport.mu.Unlock()
	if other != 1 {
		t.Fatalf("the cycle published %d critical alerts, want 1 — it did not pass through the ledger-backed Notifier", other)
	}
	m := a124Measurement{dwell: got.end.Sub(got.start), cycle: got.cycle}
	if side != nil {
		m.outboxInside, m.outboxTotal = a124Fires(t, side, "outbox", got.start, got.end)
		m.modeInside, m.modeTotal = a124Fires(t, side, "mode", got.start, got.end)
	}
	stopExecutor()
	return m
}

// a124AcceptedDelay 는 수락 변형이 실행자 트랜잭션마다 **실제 비용 위에** 더 넣는 지연이다. 개발 디스크의 실제 비용은 실패 기록 평균
// 11.1 ms · p99 19.3 ms(BenchmarkA124JudgementTransactions, 2026-09-28)이므로 트랜잭션 하나는 합 ≈21 ms 가 된다. exit 사이클의 원장 연산
// 하나는 진행 중인 실행자 트랜잭션 하나까지 기다리므로 체류 증가는 대략 「트랜잭션 길이 × 겹친 실행자 트랜잭션 수」다: 15 ms 를 넣으면(합
// ≈26 ms) +231 ms 로 고정 여유 250 ms 에 거의 닿는다(review §1). 실행자 트랜잭션이 실제 비용 포함 ≈25 ms 를 넘는 환경은 이 여유 밖이다
// — design D7 「측정된 전제」, 배포 재측정은 tasks §6.
const a124AcceptedDelay = 10 * time.Millisecond

// TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin 는 2.6 (b) 다.
func TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin(t *testing.T) {
	if testing.Short() {
		t.Skip("wall-clock measurement beside a judging executor (tasks 2.6 (b))")
	}
	// 기준선도 세 번 재서 중앙값 — 측정값과 같은 저울(gstack /review testing · performance 전문가).
	bases := make([]a124Measurement, 0, 3)
	for i := 0; i < 3; i++ {
		bases = append(bases, a124MeasureCycleBesideAJudgingExecutor(t, false, 0, 0, false))
	}
	sort.Slice(bases, func(i, j int) bool { return bases[i].dwell < bases[j].dwell })
	base := bases[1]
	if base.cycle.Judged == 0 {
		t.Fatalf("baseline judged nothing (%+v)", base.cycle)
	}
	t.Logf("baseline exit cycle dwell (no executor): %v · %v · %v — median used", bases[0].dwell, bases[1].dwell, bases[2].dwell)
	bound := base.dwell + a098ExitCycleDwellMargin

	for _, noPublisher := range []bool{false, true} {
		name := map[bool]string{false: "transport fails", true: "no publisher"}[noPublisher]
		// 수락: 판정 트랜잭션마다 a124AcceptedDelay 를 연결을 쥔 채 쓴다. 세 번 재서 **중앙값**을 수락선에 댄다 —
		// 공유 기계의 한 번 튐으로 빨개지지 않게(Eng 리뷰 F7). 여유(고정)는 그대로다.
		runs := make([]a124Measurement, 0, 3)
		for i := 0; i < 3; i++ {
			m := a124MeasureCycleBesideAJudgingExecutor(t, true, a124AcceptedDelay, 10, noPublisher)
			// 표본마다 같은 일을 했는지 — 겹침이 없는 표본이 중앙값을 고르는 데 끼면 안 된다(codex 3회차 T2).
			if m.outboxInside == 0 || m.modeInside == 0 {
				t.Fatalf("%s sample %d: executor writes inside the window record/release=%d escalation=%d — a sample that measured nothing",
					name, i, m.outboxInside, m.modeInside)
			}
			runs = append(runs, m)
		}
		sort.Slice(runs, func(i, j int) bool { return runs[i].dwell < runs[j].dwell })
		acc := runs[1]
		t.Logf("%s: three runs %v · %v · %v — median used", name, runs[0].dwell, runs[1].dwell, runs[2].dwell)
		t.Logf("%s: exit cycle dwell beside a judging executor (target %v/tx) = %v; executor writes inside the window: record/release %d of %d, escalation %d of %d",
			name, a124AcceptedDelay, acc.dwell, acc.outboxInside, acc.outboxTotal, acc.modeInside, acc.modeTotal)
		if acc.cycle.Judged != base.cycle.Judged {
			t.Fatalf("%s: the two cycles did different work: baseline judged=%d, measured judged=%d", name, base.cycle.Judged, acc.cycle.Judged)
		}
		// 계측이 판정 쪽 쓰기와 **겹쳤는지** — 겹치지 않았으면 이 측정은 판정의 몫을 안 잰 것이다.
		if acc.outboxInside == 0 {
			t.Fatalf("%s: no executor record/release overlapped the measured exit cycle — nothing was measured", name)
		}
		// 승격 쓰기(모드 행 INSERT)는 모드가 처음 바뀔 때 한 번뿐이다 — 그 뒤의 승격은 같은 모드라 쓰지 않는다(F12).
		// 그 한 번이 측정 창 안에 있었는지를 단언한다(Eng 리뷰 F2).
		if acc.modeInside == 0 {
			t.Fatalf("%s: the escalation write did not overlap the measured exit cycle (fired %d) — the judgement's share was not measured", name, acc.modeTotal)
		}
		if acc.dwell > bound {
			t.Errorf("%s: exit cycle took %v beside a judging executor; bound %v (= baseline %v + fixed margin %v)",
				name, acc.dwell, bound, base.dwell, a098ExitCycleDwellMargin)
		}
	}

	// 대조: 여유보다 긴 지연 — 계측기가 경합을 **본다면** 넘어야 한다.
	control := a124MeasureCycleBesideAJudgingExecutor(t, true, 2*a098ExitCycleDwellMargin, 10, false)
	if control.outboxInside == 0 {
		t.Fatalf("control: no executor write overlapped the window — its long dwell would not be evidence of contention")
	}
	t.Logf("control: exit cycle dwell with %v per executor transaction = %v", 2*a098ExitCycleDwellMargin, control.dwell)
	if control.dwell <= bound {
		t.Errorf("the control stayed within the margin (%v ≤ %v) — the measurement cannot see ledger contention, so the acceptance above proves nothing",
			control.dwell, bound)
	}
}

// TestTheSelectionScalesWithinTheMarginUpToTenThousandPending 는 2.6 (c) 다.
func TestTheSelectionScalesWithinTheMarginUpToTenThousandPending(t *testing.T) {
	if testing.Short() {
		t.Skip("seeds 10 000 rows")
	}
	base := a124MeasureCycleBesideAJudgingExecutor(t, false, 0, 0, false)
	for _, pending := range []int{10, 1000, 10000} {
		pending := pending
		t.Run(fmt.Sprintf("P=%d", pending), func(t *testing.T) {
			// 선택 쿼리의 연결 점유(원장 하나 · 실행자 없음)
			clk := clock.NewFake(time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
			dir := t.TempDir()
			j, err := journal.Open(context.Background(), journal.Options{
				Path:     dir + "/" + journal.DBFileName,
				Clock:    clk,
				FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
			})
			if err != nil {
				t.Fatalf("journal.Open: %v", err)
			}
			t.Cleanup(func() { _ = j.Close() })
			a124SeedBacklogFast(t, a124SideHandle(t, dir+"/"+journal.DBFileName), pending, 0)
			const reps = 20
			start := time.Now()
			for i := 0; i < reps; i++ {
				rows, err := j.PendingAlertsForDelivery(context.Background(), 10, obs.DefaultCriticalAttempts)
				if err != nil || len(rows) != min(10, pending) {
					t.Fatalf("selection: %d rows, %v", len(rows), err)
				}
			}
			occupancy := time.Since(start) / reps
			t.Logf("P=%d: PendingAlertsForDelivery connection occupancy = %v per call", pending, occupancy)
			if occupancy > a098ExitCycleDwellMargin {
				t.Errorf("P=%d: one selection holds the ledger connection for %v, past the fixed margin %v — "+
					"STOP and report: the alternative is an additive index, a schema change (Manager decision, AB1)",
					pending, occupancy, a098ExitCycleDwellMargin)
			}

			// 그 backlog 위를 도는(판정하는) 실행자 옆의 exit 체류
			m := a124MeasureCycleBesideAJudgingExecutor(t, true, 0, pending, false)
			t.Logf("P=%d: exit cycle dwell beside the executor = %v (baseline %v)", pending, m.dwell, base.dwell)
			if m.cycle.Judged != base.cycle.Judged {
				t.Fatalf("different work: baseline judged=%d, measured=%d", base.cycle.Judged, m.cycle.Judged)
			}
			if m.outboxInside == 0 {
				t.Fatalf("P=%d: no executor write overlapped the measured exit cycle — nothing was measured", pending)
			}
			// (c) 의 주장은 선택 쿼리 · 기록 · 반납이다. 판정은 돌았는지만 확인한다(승격 쓰기의 창 겹침은 (b) 가 표본마다 단언).
			if m.modeTotal == 0 {
				t.Fatalf("P=%d: the executor never escalated — the backlog was not judged", pending)
			}
			if bound := base.dwell + a098ExitCycleDwellMargin; m.dwell > bound {
				t.Errorf("P=%d: exit cycle took %v; bound %v — STOP and report (AB1)", pending, m.dwell, bound)
			}
		})
	}
}

// BenchmarkA124JudgementTransactions 는 판정 경로의 원장 트랜잭션을 **하나씩** 잰다 — 실패 기록(가산 읽기 포함) ·
// 임차 반납 · 모드 승격(첫 번째만 쓰고 나머지는 같은 모드 확인). 평균과 p99 를 따로 보고한다. 「실행자 원장
// 트랜잭션 ≈16 ms 이내」 전제(design D7)를 운영 디스크에서 직접 확인하는 도구다(tasks §6, codex 3회차 T3):
//
//	TMPDIR=<운영 원장과 같은 파일시스템> go test ./internal/app/engine/ -run '^$' -bench A124JudgementTransactions -benchtime 200x
func BenchmarkA124JudgementTransactions(b *testing.B) {
	ctx := context.Background()
	j, err := journal.Open(ctx, journal.Options{
		Path:     b.TempDir() + "/" + journal.DBFileName,
		Clock:    clock.System(),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = j.Close() })
	ids := make([]int64, b.N)
	for i := range ids {
		id, err := j.EnqueueAlert(ctx, journal.Alert{EventKey: fmt.Sprintf("a124-bench-%d", i), Type: "x", Severity: "critical"})
		if err != nil {
			b.Fatal(err)
		}
		ids[i] = id
	}
	var record, release, escalate []time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		claim, err := j.ClaimAlertByID(ctx, ids[i], "a124-bench")
		if err != nil || claim.Disposition != journal.ClaimAcquired {
			b.Fatalf("claim: %+v %v", claim, err)
		}
		t0 := time.Now()
		if _, err := j.MarkAlertAttemptFailed(ctx, ids[i], claim.Token, "bench"); err != nil {
			b.Fatal(err)
		}
		t1 := time.Now()
		if _, err := j.ReleaseAlertClaim(ctx, ids[i], claim.Token); err != nil {
			b.Fatal(err)
		}
		t2 := time.Now()
		if _, _, err := j.EscalateOperatingMode(ctx, exitAccount, journal.ModeTriggerCriticalAlertUndelivered, nil); err != nil {
			b.Fatal(err)
		}
		t3 := time.Now()
		record, release, escalate = append(record, t1.Sub(t0)), append(release, t2.Sub(t1)), append(escalate, t3.Sub(t2))
	}
	b.StopTimer()
	for _, m := range []struct {
		name string
		d    []time.Duration
	}{{"record", record}, {"release", release}, {"escalate", escalate}} {
		sort.Slice(m.d, func(a, c int) bool { return m.d[a] < m.d[c] })
		var sum time.Duration
		for _, d := range m.d {
			sum += d
		}
		b.ReportMetric(float64(sum.Microseconds())/float64(len(m.d))/1000, m.name+"-avg-ms")
		b.ReportMetric(float64(m.d[len(m.d)*99/100].Microseconds())/1000, m.name+"-p99-ms")
	}
}
