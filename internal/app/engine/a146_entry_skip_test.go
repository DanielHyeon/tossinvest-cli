//go:build tossos_testseams

package engine

// a146: Guardian 상한 0(한 단위도 못 사는 진입)은 그 진입 하나의 건너뜀 — 레인 결함 · 잠금 · 삼킴이 아님(사용자 결정 2026-10-15).
// 분류는 타입 · 코드 · sentinel 셋의 곱(design D1), 전달은 건너뛰되 모으지 않음(D2), 흔적은 계수 · 첫 원인 · (일 키, 종목)마다 한 줄(D3).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a146Skip 은 전달 · 감독자 시험용 스킵 값 — 그 단계들은 원인 모양을 보지 않고 타입만 봄(원인 모양은 승인 시험이 실제 Guardian 거절로 잼).
func a146Skip(symbol string) *strategyEntrySkip {
	return &strategyEntrySkip{market: strategyrouter.MarketKR, symbol: symbol, entry: "100", stop: "95",
		cause: errors.New("execgw: q_final refused: EXISTING_GUARDIAN_CAP")}
}

// D1: 상한 0 은 GUARDIAN_CAP_ZERO + 원인 *strategyEntrySkip(시장 · 종목 · 원래 거절) — 발급 · 강화 없음, 기록자에 한 번.
func TestA146ACapZeroPrecheckIsATypedSkip(t *testing.T) {
	capZero := a146RealCapZeroRefusal(t)
	for _, market := range pairedStrategyFirstLegResults(t) {
		loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
		issuer := &firstLegGuardianSpy{precheckErr: capZero}
		var recorded []*strategyEntrySkip
		ctx := withStrategyEntrySkipRecorder(context.Background(), func(skip *strategyEntrySkip) { recorded = append(recorded, skip) })
		outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(ctx, market.result)
		if outcome.Code != StrategyFirstLegGuardianCapZero || issuer.issueCalls != 0 || issuer.escalations != 0 {
			t.Fatalf("outcome=%+v issuer=%+v — want GUARDIAN_CAP_ZERO with no issue and no escalation", outcome, issuer)
		}
		var skip *strategyEntrySkip
		var scope *strategyScopeRefusal
		if !errors.As(outcome.cause, &skip) || errors.As(outcome.cause, &scope) || isCentralStrategyIntegrity(outcome.cause) {
			t.Fatalf("cause=%v — want an entry skip, not a scope refusal or a central fault", outcome.cause)
		}
		lineage := market.result.Lineage
		if skip.market != lineage.Market || skip.symbol != lineage.Symbol || !errors.Is(skip, risk.ErrStrategyQuantityZero) {
			t.Fatalf("skip=%+v — want the lineage's market and symbol and the original refusal as its cause", skip)
		}
		if len(recorded) != 1 || recorded[0] != skip {
			t.Fatalf("recorded=%v — want the skip handed to the cycle recorder exactly once", recorded)
		}
		// 기록자가 없는 ctx(시험 · 감독자 밖)에서도 같은 분류.
		bare := newStrategyFirstLegAdmissionBridge(&firstLegGuardianSpy{precheckErr: capZero},
			&firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}).admit(context.Background(), market.result)
		if bare.Code != StrategyFirstLegGuardianCapZero {
			t.Fatalf("without a recorder: %+v", bare)
		}
	}
}

// D1 · D4: 같은 코드의 다른 원인 · sentinel 이 q_final 거절의 원인이 아니라 곁에 있는 모양 · 같은 sentinel 의 다른 코드 · 문구 흉내는
// 오늘처럼 결함(AUTHORITY_MISMATCH) — 기록자 호출 없음.
func TestA146OtherGuardianCapRefusalsStayFaults(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	capZero := a146RealCapZeroRefusal(t)
	capCode := func(cause ...error) error {
		return errors.Join(append([]error{&execgw.QFinalRefusal{Code: riskbucket.RefusalExistingGuardianCap, Field: "q_existing_guardian"}}, cause...)...)
	}
	cases := map[string]struct {
		err  error
		code StrategyFirstLegAdmissionCode
	}{
		"cap code with no cause":                  {&execgw.QFinalRefusal{Code: riskbucket.RefusalExistingGuardianCap}, StrategyFirstLegAuthorityMismatch},
		"cap code with an FX cause":               {capCode(risk.ErrAccountBaseFXUnavailable), StrategyFirstLegAuthorityMismatch},
		"cap code with a parse cause":             {capCode(errors.New("strconv.ParseUint: invalid syntax")), StrategyFirstLegAuthorityMismatch},
		"sentinel under another q_final code":     {errors.Join(&execgw.QFinalRefusal{Code: riskbucket.RefusalZeroQuantity}, risk.ErrStrategyQuantityZero), StrategyFirstLegAuthorityMismatch},
		"sentinel under a chain refusal":          {errors.Join(&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonInvalidOrderSize)}, risk.ErrStrategyQuantityZero), StrategyFirstLegAuthorityMismatch},
		"bare sentinel":                           {risk.ErrStrategyQuantityZero, StrategyFirstLegAuthorityMismatch},
		"sentinel beside, not under, the refusal": {capCode(risk.ErrStrategyQuantityZero), StrategyFirstLegAuthorityMismatch},
		"plain text mimicking the whole refusal":  {errors.New(capZero.Error()), StrategyFirstLegAuthorityMismatch},
		"plain text naming the skip code":         {errors.New(string(StrategyFirstLegGuardianCapZero)), StrategyFirstLegAuthorityMismatch},
	}
	for name, c := range cases {
		loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
		recorded := 0
		ctx := withStrategyEntrySkipRecorder(context.Background(), func(*strategyEntrySkip) { recorded++ })
		outcome := newStrategyFirstLegAdmissionBridge(&firstLegGuardianSpy{precheckErr: c.err}, loader).admit(ctx, market.result)
		var skip *strategyEntrySkip
		if outcome.Code != c.code || errors.As(outcome.cause, &skip) || recorded != 0 {
			t.Errorf("%s: outcome=%+v recorded=%d — want %s and no skip", name, outcome, recorded, c.code)
		}
	}
	// 버킷 고갈은 여전히 범위 거절(a112 6.2 — 무편집).
	loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	outcome := newStrategyFirstLegAdmissionBridge(&firstLegGuardianSpy{precheckErr: &execgw.QFinalRefusal{Code: riskbucket.RefusalBucketCapExhausted}}, loader).
		admit(context.Background(), market.result)
	var scope *strategyScopeRefusal
	if outcome.Code != StrategyFirstLegAuthorityMismatch || !errors.As(outcome.cause, &scope) {
		t.Fatalf("bucket exhaustion outcome=%+v — want today's scope refusal", outcome)
	}
}

// D2: 스킵은 그 범위만 건너뛰고 다음 범위로 — 주기 오류 없음(감싸도, dispatch 가 감싸는 모양으로도).
func TestA146ASkipContinuesToTheNextScopeWithoutAnError(t *testing.T) {
	for name, err := range map[string]error{
		"bare":                        a146Skip("005930"),
		"wrapped like dispatch wraps": fmt.Errorf("engine: first-leg admission %s: %w", StrategyFirstLegGuardianCapZero, a146Skip("005930")),
	} {
		seen, got := a137Deliver(t, map[string]error{"a": err})
		if got != nil || strings.Join(seen, ",") != "a,b,c" {
			t.Errorf("%s: delivered=%v err=%v — want every scope delivered and no cycle error", name, seen, got)
		}
	}
}

// D2 표: 스킵은 주기 오류에 합쳐지지 않고, 범위 거절 보고 · 멈추는 결함 · 일일 손실 표지의 처리는 그대로.
func TestA146ASkipIsNeverJoinedIntoTheCycleError(t *testing.T) {
	scope := &strategyScopeRefusal{scope: strategyrouter.OwnerKey{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930"},
		detail: "this owner scope's risk bucket is exhausted"}
	fault := errors.New("engine: first-leg admission AUTHORITY_MISMATCH: equity missing")

	seen, err := a137Deliver(t, map[string]error{"a": fmt.Errorf("x: %w", scope), "b": a146Skip("000660")})
	var skipped *strategyScopeRefusal
	var skip *strategyEntrySkip
	if strings.Join(seen, ",") != "a,b,c" || !errors.As(err, &skipped) || skipped != scope || errors.As(err, &skip) {
		t.Fatalf("scope refusal + skip: delivered=%v err=%v — want the scope refusal reported as today and the skip left out", seen, err)
	}

	seen, err = a137Deliver(t, map[string]error{"a": a146Skip("005930"), "b": fault})
	if strings.Join(seen, ",") != "a,b" || !errors.Is(err, fault) || errors.As(err, &skip) {
		t.Fatalf("skip + fault: delivered=%v err=%v — want the stopping fault alone, as today", seen, err)
	}

	seen, err = a137Deliver(t, map[string]error{"a": a146Skip("005930"), "b": a137Block()})
	if strings.Join(seen, ",") != "a,b" || err != nil {
		t.Fatalf("skip + daily-loss block: delivered=%v err=%v — want the quiet daily-loss end, as today", seen, err)
	}
}

// D2 · F6: 스킵만 있는 주기는 두 레인 종류 모두에서 잠금 0 · 삼킴 0(대조군: 표지 없는 결함은 a137 시험이 오늘처럼 잼).
func TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind(t *testing.T) {
	for _, effective := range []bool{true, false} {
		snapshot, calls := a137RunCycles(t, effective, fmt.Errorf("engine: first-leg admission %s: %w", StrategyFirstLegGuardianCapZero, a146Skip("005930")), 3)
		if calls < 3 || snapshot.Latched || snapshot.SwallowedCycleErrors != 0 {
			t.Errorf("effective=%v: cycles=%d snapshot=%+v — want ≥3 cycles, no latch and nothing swallowed", effective, calls, snapshot)
		}
	}
}

// D3: 감독자가 주기 ctx 에 실은 기록자로 스킵을 세고(주기마다), 첫 원인을 지키며, 알림은 (계좌 일 키, 종목)마다 한 번.
func TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol(t *testing.T) {
	for _, effective := range []bool{true, false} {
		var calls atomic.Int32
		var mu sync.Mutex
		var notices []string
		cycle := func(ctx context.Context) error {
			calls.Add(1)
			recordStrategyEntrySkip(ctx, a146Skip("005930"))
			recordStrategyEntrySkip(ctx, a146Skip("000660"))
			return nil
		}
		fake := clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)) // KST 10:00 — 일 키 10-17
		// US 는 아무것도 하지 않는 주기 — calls 가 KR 주기만 세게.
		idle := func(context.Context) error { return nil }
		worker := func(market StrategyMarket, cycle StrategyCycle) StrategyMarketWorker {
			if !effective {
				return StrategyMarketWorker{Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
			}
			return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
				AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
				EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				LatchRevision:      1, PollInterval: DefaultStrategyCycleLimit}
		}
		supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: fake, CycleLimit: MaximumStrategyCycleLimit,
			Workers: []StrategyMarketWorker{worker(StrategyMarketKR, cycle), worker(StrategyMarketUS, idle)},
			EntrySkipNotice: func(market StrategyMarket, day, symbol, cause string) {
				mu.Lock()
				defer mu.Unlock()
				if market == StrategyMarketKR {
					notices = append(notices, day+" "+symbol+" "+cause)
				}
			}})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- supervisor.Run(ctx) }()
		<-supervisor.Ready()
		advance := func(until int32) {
			deadline := time.Now().Add(3 * time.Second)
			for calls.Load() < until && time.Now().Before(deadline) {
				fake.Advance(DefaultStrategyCycleLimit)
				time.Sleep(time.Millisecond)
			}
			time.Sleep(20 * time.Millisecond)
		}
		advance(4)
		snapshot, _ := supervisor.Snapshot(StrategyMarketKR)
		mu.Lock()
		first := append([]string(nil), notices...)
		mu.Unlock()
		if snapshot.Latched || snapshot.SwallowedCycleErrors != 0 || snapshot.GuardianCapZeroSkips < 6 ||
			!strings.Contains(snapshot.FirstGuardianCapZeroSkip, "005930") || strings.Contains(snapshot.FirstGuardianCapZeroSkip, "000660") {
			t.Fatalf("effective=%v snapshot=%+v — want ≥6 skips counted, the first cause kept, no latch, nothing swallowed", effective, snapshot)
		}
		if len(first) != 2 || !strings.HasPrefix(first[0], "2026-10-17 005930 ") || !strings.HasPrefix(first[1], "2026-10-17 000660 ") {
			t.Fatalf("effective=%v notices=%v — want exactly one per symbol for account day 2026-10-17", effective, first)
		}
		// KST 08:00 을 넘겨 다음 일 키 → 종목마다 한 번 더.
		fake.Set(time.Date(2026, 10, 17, 23, 0, 0, 0, time.UTC))
		advance(calls.Load() + 3)
		mu.Lock()
		second := append([]string(nil), notices...)
		mu.Unlock()
		cancel()
		<-done
		if len(second) != 4 || !strings.HasPrefix(second[2], "2026-10-18 005930 ") || !strings.HasPrefix(second[3], "2026-10-18 000660 ") {
			t.Fatalf("effective=%v notices=%v — want one more per symbol for account day 2026-10-18", effective, second)
		}
	}
}

// D3: 생산 알림은 경고 줄 하나 — 이벤트 이름 · 시장 · 종목 · 계좌 일 키 · 매도 무관 문구. 로그가 없으면 알림 자리도 없음.
func TestA146TheProductionNoticeIsOneWarningLine(t *testing.T) {
	if (&Context{}).strategyEntrySkipNotice() != nil {
		t.Fatal("without a logger there must be no notice seam")
	}
	var buf strings.Builder
	c := &Context{Log: obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true})}
	c.strategyEntrySkipNotice()(StrategyMarketKR, "2026-10-17", "005930", a146Skip("005930").Error())
	line := buf.String()
	for _, want := range []string{string(obs.EventStrategyEntrySkipped), `"market":"KR"`, `"symbol":"005930"`, `"account_day":"2026-10-17"`,
		string(StrategyFirstLegGuardianCapZero), "at entry 100 stop 95", "sells, stops and exits are unaffected"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q lacks %q", line, want)
		}
	}
	if strings.Count(line, "\n") != 1 || !strings.Contains(line, `"level":"WARN"`) && !strings.Contains(line, `"level":"warn"`) {
		t.Fatalf("want exactly one warning line, got %q", line)
	}
}

// D3: 생산 조립이 스킵 알림 자리를 실제로 감독자에 넘김(배선 한 줄을 못 박음).
func TestA146TheProductionSupervisorCarriesTheSkipNotice(t *testing.T) {
	var buf strings.Builder
	c := &Context{Log: obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true}),
		Entry: execgw.NewEntryGate(clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)), nil)}
	supervisor, err := c.NewRefreshingPairedStrategyEntrySupervisor(clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	if supervisor.entrySkipNotice == nil {
		t.Fatal("the production supervisor has no entry-skip notice seam")
	}
}

// 실제 경로(F1 특성 시험의 뒤집힘): 예산 1 KRW 의 두 범위가 **둘 다** precheck 까지 평가되고, 주기 오류 없음, 브로커 0, 스킵 둘 기록.
func TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	var recorded []string
	ctx := withStrategyEntrySkipRecorder(context.Background(), func(skip *strategyEntrySkip) {
		recorded = append(recorded, string(skip.market)+":"+skip.symbol)
		if !errors.Is(skip, risk.ErrStrategyQuantityZero) {
			t.Errorf("skip %v lost the real Guardian refusal's sentinel", skip)
		}
		var qFinal *execgw.QFinalRefusal
		if !errors.As(skip, &qFinal) || qFinal.Code != riskbucket.RefusalExistingGuardianCap {
			t.Errorf("skip %v lost the real Guardian refusal", skip)
		}
	})
	err := dispatchStrategyMarketHandoffs(ctx, fixture.journal, fixture.cycle, fixture.proposals.kr.dispatchHandoffs())
	if err != nil {
		t.Fatalf("err=%v — a cycle of cap-0 skips must end without a cycle error", err)
	}
	if got := counter.prechecks.Load(); got != 2 {
		t.Fatalf("prechecks=%d — want both scopes evaluated", got)
	}
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v — a skipped entry must never reach the broker", placed)
	}
	if strings.Join(recorded, ",") != "KR:005930,KR:000660" {
		t.Fatalf("recorded=%v — want one skip per scope in coordinator order", recorded)
	}
}

// a146FixtureWithEquity 는 거래 fixture 의 계좌 권한 자본만 바꿈(체인 일일 손실 단이 입력 불가로 거절할 결함을 심는 데 씀).
func a146FixtureWithEquity(t *testing.T, riskBudget string, equity riskcalc.Money) a112TradingFixture {
	t.Helper()
	fixture := newA112TradingFixture(t, a112TradingOptions{riskBudget: riskBudget})
	loader := a112AccountLoaderWith(t, fixture.now, "", nil, "")
	load := loader.load
	loader.load = func(ctx context.Context, config strategyaccount.ProductionConfig) (strategyaccount.Authority, error) {
		authority, err := load(ctx, config)
		if err != nil {
			return authority, err
		}
		state := authority.AccountState()
		state.AccountEquity = equity
		return strategyaccount.AuthorityForTest(config.Market, authority.QuoteCurrency(), state, authority.ObservedAt(), authority.FreshUntil(),
			authority.Generation(), authority.ManifestDigest()), nil
	}
	fixture.accounts = loader.collect(context.Background(), fixture.proposals)
	fixture.wave(t)
	return fixture
}

// 명시 한계(design F10 · freeze 리뷰 P1-1): 상한 0 은 그 뒤의 판정(여기서는 체인의 입력 불가)을 가림 — 상한 0 이면 건너뜀(주문 0), 같은 결함은
// 상한이 한 단위 이상이 되는 주기에 결함으로 드러남. 양쪽을 못 박아 한계가 조용히 넓어지거나(다른 결함까지 스킵) 좁아지는 것을 봄.
func TestA146ACapZeroSkipHidesTheChecksAfterSizingUntilTheCapRises(t *testing.T) {
	broken := riskcalc.Money{Amount: "10000000", Currency: "USD"} // 한도 통화(KRW)와 다른 자본 — 체인 일일 손실 단의 입력 불가
	hidden := a146FixtureWithEquity(t, "1", broken)
	var skips int
	ctx := withStrategyEntrySkipRecorder(context.Background(), func(*strategyEntrySkip) { skips++ })
	err := dispatchStrategyMarketHandoffs(ctx, hidden.journal, hidden.cycle, hidden.proposals.kr.dispatchHandoffs())
	if err != nil || skips != 2 || len(hidden.placedSymbols()) != 0 {
		t.Fatalf("cap 0 + broken equity: err=%v skips=%d placed=%v — want the cap-0 skip to hide the fault, no order", err, skips, hidden.placedSymbols())
	}

	surfaced := a146FixtureWithEquity(t, "", broken) // 기본 예산 — 상한 ≥ 1
	err = surfaced.deliverKR(t)
	var issue *execgw.IssueRefusal
	if err == nil || !strings.Contains(err.Error(), string(StrategyFirstLegAuthorityMismatch)) || !errors.As(err, &issue) ||
		issue.Reason != string(risk.ReasonInputUnavailable) || len(surfaced.placedSymbols()) != 0 {
		t.Fatalf("cap ≥ 1 + broken equity: err=%v placed=%v — want the same fault surfaced as AUTHORITY_MISMATCH (INPUT_UNAVAILABLE), no order",
			err, surfaced.placedSymbols())
	}
}

// freeze 리뷰 P2-3: 계좌 일 키를 셀 수 없어도(시간대 자료 결함) 알림은 침묵하지 않음 — 빈 일 키로 종목마다 한 번.
func TestA146AnUncountableAccountDayStillAnnouncesTheSkip(t *testing.T) {
	saved := strategyEntrySkipAccountDay
	strategyEntrySkipAccountDay = func(time.Time) (string, error) { return "", errors.New("time zone data unavailable") }
	t.Cleanup(func() { strategyEntrySkipAccountDay = saved })
	var notices []string
	supervisor := &StrategyEntrySupervisor{clk: clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)),
		entrySkipNotice: func(market StrategyMarket, day, symbol, cause string) { notices = append(notices, day+"|"+symbol) }}
	worker := &strategyMarketRuntime{descriptor: StrategyMarketWorker{Market: StrategyMarketKR}}
	for range 3 {
		supervisor.recordGuardianCapZeroSkip(worker, a146Skip("005930"))
		supervisor.recordGuardianCapZeroSkip(worker, a146Skip("000660"))
	}
	if strings.Join(notices, ",") != "|005930,|000660" || worker.guardianCapZeroSkips != 6 {
		t.Fatalf("notices=%v skips=%d — want one notice per symbol with an empty day and every skip counted", notices, worker.guardianCapZeroSkips)
	}
}
