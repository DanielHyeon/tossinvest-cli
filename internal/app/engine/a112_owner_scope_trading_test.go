//go:build tossos_testseams

package engine

// a112 태스크 5.2.2.2 — 하류 권한(결과 · 위험 · 계좌 · 1차 레그)의 소유자 범위 전환과 시장 단위 개수 관문 제거.
//
// 이 파일의 fixture 는 생산 조립과 **같은 적재기**를 돈다: 제안 권한 쌍(서명 활성화 · 범위 둘) → `ResultAuthority()` → 위험 적재기
// `collect`(서명 위험 정책 — 두 종목) → 계좌 적재기 `collect`(적재 함수만 시험 권한으로 바꿈) → 1차 레그 권한 loader → dispatch 주기 →
// 생산 전달 몸통 `dispatchStrategyMarketHandoffs`(원장은 실제 journal, Gateway 는 스파이 — 실주문 아님).
//
// 종결 문장(Done): 「서명 활성화된 두 소유자 범위 시장이 범위마다 주문을 낸다」 — 첫 레그 둘, 각자 자기 범위의 위험 · 계좌 권한으로.

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

type a112TradingOptions struct {
	// riskSymbols 는 KR 서명 위험 정책에 더할 종목(기본: 둘째 범위 000660 을 같은 섹터로).
	riskSymbols []riskLoaderSymbol
	// maxOpenExposure 는 Guardian 계좌 노출 상한(KRW). 빈 값이면 기본 정책.
	maxOpenExposure string
	// failAccountFor 는 계좌 권한 적재를 실패시킬 종목. accountFailure 가 있으면 그 오류로(없으면 범위 국소 평문 오류).
	failAccountFor string
	accountFailure error
	// coordinatorOrder 면 둘째 범위(000660)를 앞에 둔다(조정자 사전순).
	coordinatorOrder bool
	// shortFreshFor 종목의 계좌 권한은 30초 뒤 만료(나머지는 1분).
	shortFreshFor string
	// a112 6.2: 둘째 범위(000660)의 레인(빈 값이면 첫 범위와 같은 continuation), 첫 레인 strategy 한도, 추가 strategy 항목.
	secondLane      *strategyflow.Descriptor
	strategyLimit   string
	extraStrategies []riskLoaderStrategy
	horizonShort    string
	// disallowFor 종목의 계좌 권한은 그 종목을 허용하지 않는다(Guardian precheck 의 버킷 고갈이 아닌 거절).
	disallowFor string
	// riskBudget 은 Guardian 정책의 거래당 위험 예산(KRW) — 아주 작으면 q_existing_guardian 0(버킷 고갈이 아닌 q_final 코드).
	riskBudget string
}

type a112TradingFixture struct {
	now        time.Time
	riskLoader strategyRiskAuthorityLoader
	// riskStub 은 축소 위험 원장 경로(a127 전에는 적재기가 늘 이것을 읽었음) — 손상 주입 시험만 useRiskStub 으로 씀.
	riskStub  string
	fx        strategyFXAuthorityPair
	guardian  *execgw.RiskGuardian
	clk       *clock.Fake
	journal   *journal.Journal
	proposals strategyProposalAuthorityPair
	risk      strategyRiskAuthorityPair
	accounts  strategyAccountAuthorityPair
	loader    *productionStrategyFirstLegAuthorityLoader
	cycle     *strategyDispatchCycle
	spy       *strategyDispatchGatewaySpy
	winner    strategyproposal.ProductionAuthority
	second    strategyproposal.ProductionAuthority
}

func newA112TradingFixture(t *testing.T, options a112TradingOptions) a112TradingFixture {
	t.Helper()
	symbols := options.riskSymbols
	if symbols == nil {
		symbols = []riskLoaderSymbol{{Symbol: "000660", Sector: "technology", SectorLimitMinor: "3000000", SymbolLimitMinor: "2000000"}}
	}
	riskFixture := newStrategyRiskLoaderFixtureFor(t, riskLoaderFixtureOptions{extraKR: symbols, generation: 1,
		krStrategyLimit: options.strategyLimit, extraKRStrategies: options.extraStrategies, krHorizonShort: options.horizonShort})
	now := riskFixture.results.observedAt
	// 제안 권한 쌍: KR 은 범위 둘(원래 005930 + 000660) · 서명 활성화, US 는 fixture 그대로 하나.
	proposals := strategyProposalAuthorityPair{observedAt: now}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		result := riskFixture.results.forMarket(market).result
		batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:trade-"+string(market), map[string]strategyflow.Result{result.Lineage.Symbol: result})
		authority, ok := batch.For(result.Lineage.Symbol)
		if !ok {
			t.Fatal("missing proposal test authority")
		}
		value := strategyProposalMarketAuthority{market: market, entries: []strategyProposalEntryAuthority{{authority: authority}},
			snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady}}
		value.snapshot.ProposalSetDigest = strategyProposalSetDigest(value.entries)
		if market == StrategyMarketKR {
			proposals.kr = value
		} else {
			proposals.us = value
		}
	}
	winner := proposals.kr.entries[0].authority
	secondLane := riskLoaderDescriptor(t, StrategyMarketKR)
	if options.secondLane != nil {
		secondLane = *options.secondLane
	}
	two := a112ExtraEntryKRWith(t, proposals.kr, now, "000660", options.coordinatorOrder, secondLane)
	two.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	proposals.kr = two
	second := two.entries[1].authority
	if options.coordinatorOrder {
		second = two.entries[0].authority
	}

	handle, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(t.TempDir(), journal.DBFileName),
		Clock: clock.NewFake(now), FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = handle.Close() })
	fakeClock := clock.NewFake(now)
	policy := risk.DefaultPolicy()
	if options.maxOpenExposure != "" {
		policy.MaxOpenExposure = riskcalc.Money{Amount: options.maxOpenExposure, Currency: "KRW"}
	}
	if options.riskBudget != "" {
		policy.RiskBudget = riskcalc.Money{Amount: options.riskBudget, Currency: "KRW"}
	}
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: handle, Clock: fakeClock, AccountRef: "acct-risk-loader",
		Policy: policy, Costs: costs.DefaultModel(), PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	// a127: 위험 적재기는 admission 과 **같은 실제 원장**을 읽는다(스키마 핀 수리 — 다리 `a112MirrorLedgerIntoRiskStub` 제거). 축소 stub 원장은
	// 실제 원장의 트리거 · 제약이 막는 손상 모양(GARBAGE 행 · latch 뷰)을 심는 시험만 `useRiskStub` 으로 고른다.
	riskLoader := *riskFixture.loader
	riskStub := riskLoader.journalPath
	riskLoader.journalPath = handle.Path()
	fixture := a112TradingFixture{now: now, riskLoader: riskLoader, riskStub: riskStub, fx: riskFixture.fx, guardian: guardian, clk: fakeClock, journal: handle,
		proposals: proposals, spy: &strategyDispatchGatewaySpy{observed: map[string]int{}}, winner: winner, second: second}
	accountLoader := a112AccountLoaderWith(t, now, options.failAccountFor, options.accountFailure, options.shortFreshFor)
	if options.disallowFor != "" {
		load := accountLoader.load
		accountLoader.load = func(ctx context.Context, config strategyaccount.ProductionConfig) (strategyaccount.Authority, error) {
			authority, err := load(ctx, config)
			if err != nil || config.Symbol != options.disallowFor {
				return authority, err
			}
			state := authority.AccountState()
			state.AllowedSymbols = []string{"999999"}
			return strategyaccount.AuthorityForTest(config.Market, authority.QuoteCurrency(), state, authority.ObservedAt(), authority.FreshUntil(),
				authority.Generation(), authority.ManifestDigest()), nil
		}
	}
	fixture.accounts = accountLoader.collect(context.Background(), proposals)
	fixture.wave(t)
	return fixture
}

// useRiskStub 은 위험 적재기를 축소 stub 원장으로 돌림 — 실제 원장의 트리거 · STRICT · FK 가 막는 손상 모양(GARBAGE 행 · latch 뷰)을 심어 결함
// 분류를 재는 시험 전용(a127: 기본은 실제 원장). 반환값은 stub 경로.
func (fixture *a112TradingFixture) useRiskStub() string {
	fixture.riskLoader.journalPath = fixture.riskStub
	return fixture.riskStub
}

// sqlOpenReadOnlyCount 는 원장 파일을 읽기 전용으로 열어 count 질의 하나를 읽음(시험 준비 단언용).
func sqlOpenReadOnlyCount(t *testing.T, path, query string, out *int) error {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	return db.QueryRow(query).Scan(out)
}

// wave 는 생산 조립처럼 결과 → 위험 권한을 **지금 원장으로** 다시 모으고 1차 레그 권한 · dispatch 주기를 새로 세운다(새 파도).
func (fixture *a112TradingFixture) wave(t *testing.T) {
	t.Helper()
	fixture.risk = fixture.riskLoader.collect(context.Background(), fixture.proposals.ResultAuthority(), fixture.fx)
	schedule := pairedDispatchSchedule(fixture.now)
	fixture.loader = newProductionStrategyFirstLegAuthorityLoader(fixture.clk, fixture.journal, fixture.guardian, schedule, fixture.proposals,
		fixture.risk, fixture.fx, fixture.accounts)
	fixture.cycle = newStrategyDispatchCycle(fixture.journal, fixture.spy, newStrategyFirstLegAdmissionBridge(fixture.guardian, fixture.loader),
		schedule, fixture.fx, fixture.risk, fixture.proposals, &strategyDispatchOwnerCoordinator{})
	fixture.cycle.revalidateSchedule = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error { return nil }
	fixture.cycle.now = fixture.clk.Now
}

// a112AccountLoader 는 생산 계좌 적재기를 세우되 적재 함수만 시험 권한으로 바꾼다 — 종목마다 그 종목을 허용한 계좌 권한.
func a112AccountLoaderWith(t *testing.T, now time.Time, failFor string, failure error, shortFreshFor string) *strategyAccountAuthorityLoader {
	t.Helper()
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	const digest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	env := map[string]string{strategyAccountPublicKeyEnv: base64.StdEncoding.EncodeToString(public), strategyAccountKeyIDEnv: "account-key-1",
		strategyAccountKRManifestDigestEnv: digest, strategyAccountUSManifestDigestEnv: digest}
	loader := newStrategyAccountAuthorityLoader(t.TempDir(), "acct-risk-loader", "KRW", now, func(key string) string { return env[key] })
	loader.load = func(_ context.Context, config strategyaccount.ProductionConfig) (strategyaccount.Authority, error) {
		if config.Symbol == failFor {
			if failure != nil {
				return strategyaccount.Authority{}, failure
			}
			return strategyaccount.Authority{}, errors.New("account authority unavailable for " + config.Symbol)
		}
		quote, cash := "KRW", "5000000"
		if config.Market == strategyaccount.MarketUS {
			quote, cash = "USD", "1000"
		}
		state := risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{config.Symbol}, HeldQuantity: "0",
			CashAvailable: riskcalc.Money{Amount: cash, Currency: quote}, OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"}, AccountEquity: riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
		fresh := now.Add(time.Minute)
		if config.Symbol == shortFreshFor {
			fresh = now.Add(30 * time.Second)
		}
		return strategyaccount.AuthorityForTest(config.Market, quote, state, now.Add(-time.Second), fresh, 1, digest), nil
	}
	return loader
}

func (fixture a112TradingFixture) placedSymbols() []string {
	fixture.spy.mu.Lock()
	defer fixture.spy.mu.Unlock()
	out := make([]string, 0, len(fixture.spy.calls))
	for _, call := range fixture.spy.calls {
		out = append(out, call.Intent.Symbol)
	}
	sort.Strings(out)
	return out
}

func (fixture a112TradingFixture) deliverKR(t *testing.T) error {
	t.Helper()
	return dispatchStrategyMarketHandoffs(context.Background(), fixture.journal, fixture.cycle, fixture.proposals.kr.dispatchHandoffs())
}

// Done: 서명 활성화된 두 소유자 범위 시장이 범위마다 첫 레그를 낸다 — 각자 자기 범위의 위험 · 계좌 권한으로.
//
// Done 문장의 확정된 의미(2026-10-01 Manager 판정 ④ — 약화가 아니라 정밀화): **이 fixture 에서** 범위별 발급은 파도 순차다(사이에 해제가
// 끼지 않는 원장 — 일반적인 안전은 원장 기준 합산과 스냅숏 하한 대조가 진다, codex #3 · 재확인 T). 두 범위는 horizon · 시장 ·
// 계좌 버킷을 공유하므로 첫 레그의 admission 이 공유 버킷 사용량을 올리고, 같은 파도에서 모은 둘째 범위의 버킷 스냅숏은 뒤처져 journal 이
// `BUCKET_USAGE_STALE` 로 거절한다(범위 거절이 아닌 원장 거절 — 주기가 멈춘다). 이것은 결함이 아니라 공유 버킷 이중 소비를 막는 설계다
// (관문 전수표 (e)). 둘째 파도가 첫 레그의 held 를 반영한 번들을 다시 모으면 둘째 범위가 발급된다 — 첫 범위는 캠페인이 이제 FLAT 이 아니라
// 전달 몸통이 건너뛴다(굶음 없음).
//
// 둘째 파도는 위험 적재기가 admission 과 같은 실제 원장을 다시 읽어 첫 레그의 held 를 본다(a127 — 예전 원장 → stub 복사 다리는 지워짐).
func TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	// J2 실측: 활성 두 범위 파도의 범위 항목 수 = 위험 N + 계좌 N(범위마다 적재 시도 하나 — 적재 호출은 각 적재기의 범위 순회 안 한 자리;
	// 항목 수는 적재 성공 수가 아니다 — 리뷰 B #6).
	if risk, account := len(fixture.risk.kr.scopes), len(fixture.accounts.kr.scopes); risk != 2 || account != 2 {
		t.Fatalf("reads per wave: risk=%d account=%d, want one per owner scope (2 + 2)", risk, account)
	}
	t.Logf("J2 measured: one activated two-scope KR wave holds risk=%d + account=%d owner-scope entries (one load attempt each)",
		len(fixture.risk.kr.scopes), len(fixture.accounts.kr.scopes))
	firstWave := fixture.deliverKR(t)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("first wave placed=%s err=%v, want the first scope in coordinator order only", got, firstWave)
	}
	if firstWave == nil || !strings.Contains(firstWave.Error(), "BUCKET_USAGE_STALE") {
		t.Fatalf("first wave err=%v, want the second scope refused by the shared-bucket usage CAS (stale snapshot) in the same wave", firstWave)
	}
	if scope := (*strategyScopeRefusal)(nil); errors.As(firstWave, &scope) {
		t.Fatalf("first wave err=%v — the journal's stale-usage refusal must not be typed as a scope refusal (J4)", firstWave)
	}
	var reservations int
	if err := sqlOpenReadOnlyCount(t, fixture.journal.Path(), `SELECT count(*) FROM risk_bucket_reservations`, &reservations); err != nil || reservations == 0 {
		t.Fatalf("arrangement: the first leg left no bucket reservation in the admission ledger (n=%d err=%v) — the second wave would prove nothing", reservations, err)
	}
	fixture.wave(t)
	if err := fixture.deliverKR(t); err != nil {
		t.Fatalf("second wave err=%v, want the second scope issued on a fresh bundle", err)
	}
	if got := strings.Join(fixture.placedSymbols(), ","); got != "000660,005930" {
		t.Fatalf("placed=%s, want one first leg per owner scope (000660, 005930)", got)
	}
	// 적재기가 범위마다 번들을 따로 만들었다(준비된 번들 digest 가 서로 다른 둘) — 이것은 적재기 출력의 개수이고, 「각 범위가 **자기** 번들로
	// 발급됐다」를 지키는 것은 1차 레그의 범위 대조(`production risk authority scope changed` — 변이 X08 이 그 대조로 CAUGHT, 리뷰 B #4)다.
	digests := map[string]bool{}
	for _, scope := range fixture.risk.kr.scopes {
		if scope.ready {
			digests[scope.bundle.Digest()] = true
		}
	}
	if len(digests) != 2 {
		t.Fatalf("per-scope risk authorities=%d distinct ready bundles, want 2", len(digests))
	}
}

// 두 레그 한 주기(관문 전수표 (b) · J5 ② — 리뷰 B #5 로 첫 파도 단언으로 고침): 같은 주기의 둘째 범위 admission 은 첫 레그의 **held** 예약을
// 센다. 한 레그(801)만 들어가는 계좌 노출 상한(1200)에서 둘째는 journal 합산(usage + held + new)으로 거절된다 — 브로커 스냅숏(계좌 권한의
// OpenExposure=0)은 첫 레그를 모르는데도. 이 합산 검사는 버킷 사용량 CAS(stale)보다 먼저 돈다(실측: 리뷰 B 사본). 예약 버전(CAS) 전진도 직접 잰다.
func TestTheSecondLegOfOneCycleCountsTheFirstLegsHeldReservation(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{maxOpenExposure: "1200"})
	versionBefore, err := fixture.journal.ReservationVersion(context.Background(), "acct-risk-loader")
	if err != nil {
		t.Fatal(err)
	}
	if fixture.accounts.kr.authority.OpenExposure().Amount != "0" {
		t.Fatal("arrangement: the broker snapshot must not know the first leg")
	}
	err = fixture.deliverKR(t) // 한 주기 — 첫 범위 발급, 둘째 범위 admission 은 첫 레그의 held 를 셈
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("placed=%s err=%v, want only the first scope under the exposure cap", got, err)
	}
	versionAfter, verr := fixture.journal.ReservationVersion(context.Background(), "acct-risk-loader")
	if verr != nil || versionAfter <= versionBefore {
		t.Fatalf("reservation version %d → %d (err=%v) — the first leg's reservation did not move the CAS version the next admission reads",
			versionBefore, versionAfter, verr)
	}
	if err == nil || !strings.Contains(err.Error(), "OPEN_EXPOSURE") || !strings.Contains(err.Error(), "already held 801") {
		t.Fatalf("err=%v, want the journal's aggregate refusal counting the first leg's held reservation (already held 801)", err)
	}
}

// 계좌 적재 실패는 결함이다(5.2.2.2 codex 재확인 P1 → Manager 판정 (A), 2026-10-01). 생산 계좌 매니페스트는 **시장 단위 파일 하나**
// (`strategyaccount.FileName(market)`)라 범위별 정책 거절 원인이 없다 — 적재 실패(파일 부재 · digest · 서명 · 창 · ctx)는 시장의 사유다. 범위마다
// 결과가 갈리는 것은 순차 읽기 사이 파일 교체 · 일시적 I/O 같은 경우뿐이고(codex 재확인 #2 T 정정), 그때도 결함이다. 그래서 한 범위의 계좌 적재
// 실패는 범위 거절(건너뛰기)이 아니라 주기를 멈추는 타입 없는 결함이고 원인을 남긴다. 앞 판(J3 계좌 절반 — 「그 범위만 거절, 다른 범위 거래」)은
// 범위별 정책 거절 원인이 있다는 틀린 전제를 재던 것이라 뒤집었다.
func TestAnAccountLoadFailureOnOneScopeIsAFaultThatStopsTheCycle(t *testing.T) {
	failure := errors.New("account manifest digest mismatch")
	for _, order := range []struct {
		name   string
		first  bool
		placed string
	}{{"fixture order (failing scope first)", false, ""}, {"coordinator order (000660 first)", true, "000660"}} {
		t.Run(order.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, a112TradingOptions{failAccountFor: "005930", accountFailure: failure, coordinatorOrder: order.first})
			err := fixture.deliverKR(t)
			if got := strings.Join(fixture.placedSymbols(), ","); got != order.placed {
				t.Fatalf("placed=%s err=%v, want %q — the account fault must stop the cycle, not skip to the next scope", got, err, order.placed)
			}
			if err == nil || a112ScopeRefusalOf(err) != nil || !errors.Is(err, failure) {
				t.Fatalf("err=%v — want an untyped fault carrying the account loader's cause", err)
			}
		})
	}
}

// 통화 재유도(A#3): 봉투의 통화가 아니라 조립 권한의 계보 시장에서 다시 유도한다.
func TestTheFirstLegCurrencyComesFromTheLineageNotTheEnvelope(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t) // 단일 범위 — 발급까지 가는 조립
	accepted := a112Accepted(t, fixture.proposals.kr.entries[0].authority)
	accepted.currency = "USD" // 위조: 검증기를 거치지 않은 봉투 통화
	issuance, err := fixture.loaderWith(fixture.proposals).collectStrategyFirstLegAuthority(context.Background(), accepted)
	if err != nil {
		t.Fatalf("arrangement: the single-scope assembly refused: %v", err)
	}
	if issuance.Entry.Currency != "KRW" {
		t.Fatalf("issued with envelope currency %q, want KRW re-derived from the KR lineage", issuance.Entry.Currency)
	}
}

// a127 D4 · S1 — 생산 위험 적재기는 admission 원장(`journal.Open`, 스키마 journal.SchemaVersion)을 주입된 현재 스키마로 읽는다. a112 의
// 트립와이어(핀 27 이 실제 원장을 거절함을 단언하던 시험)를 양성으로 뒤집은 것 — 동결 리터럴로 되돌리면(S1) 이 시험이 실패한다.
func TestTheRiskLoaderReadsTheRealJournal(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	loader := fixture.riskLoader
	if loader.journalPath != fixture.journal.Path() {
		t.Fatalf("arrangement: the risk loader reads %q, want the admission ledger %q", loader.journalPath, fixture.journal.Path())
	}
	scoped := fixture.proposals.ResultAuthority().kr.results()
	if len(scoped) != 2 {
		t.Fatalf("arrangement: KR results=%d, want the two owner scopes", len(scoped))
	}
	bundle, err := riskbucket.LoadProductionRiskSnapshotAuthority(context.Background(), riskbucket.ProductionRiskSnapshotConfig{
		ConfigDir: loader.configDir, JournalPath: loader.journalPath, Market: riskbucket.MarketKR, AccountID: loader.accountID,
		AccountCurrency: loader.accountCurrency, ManifestDigest: loader.digests[StrategyMarketKR], TrustedKeyID: loader.keyID,
		TrustedKey: loader.key, ObservedAt: loader.observedAt, JournalSchemaVersion: journal.SchemaVersion,
	}, riskbucket.ProductionRiskSnapshotInput{Result: scoped[0], FX: fixture.fx.kr.read.evidence})
	// 번들의 항목 배열은 길이 5 고정이라 길이 비교는 공허함(1.6 리뷰 P3-2) — digest 와 범위 준비 수로 잼.
	if err != nil || bundle.Digest() == "" {
		t.Fatalf("real-journal load (schema %d) err=%v digest=%q, want a sealed five-bucket bundle", journal.SchemaVersion, err, bundle.Digest())
	}
	collected := loader.collect(context.Background(), fixture.proposals.ResultAuthority(), fixture.fx)
	ready := 0
	for _, scope := range collected.kr.scopes {
		if !scope.ready {
			t.Fatalf("scope %v not ready against the real journal: %v", scope.key, scope.cause)
		}
		ready++
	}
	if ready != 2 {
		t.Fatalf("ready scopes=%d against the real journal, want both owner scopes", ready)
	}
}

// worker 승격(5.2.2.2): 서명 활성화된 두 범위 시장의 worker 는 시장 단위 handoff(상한 1 → OverCapacity)가 아니라 **범위별** handoff 를
// 본다 — 안 그러면 Done 의 시장은 worker 가 dormant 라 주기 자체가 돌지 않는다. 한 범위의 진입 관문 거절은 그 범위만 막는다(J3 — 다른
// 범위의 승격을 굶기지 않음). 모든 범위가 막히면 dormant, 활성화 없는 시장은 오늘 그대로(상한 거절 → dormant, 토글 OFF = upstream).
func TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	loader := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	now := loader.schedule.observedAt
	candidates := strategyCandidateAuthorityPair{observedAt: now, kr: readyCandidateAuthority(StrategyMarketKR), us: readyCandidateAuthority(StrategyMarketUS)}
	routes := strategyRouteAuthorityPair{observedAt: now, kr: readyRouteAuthority(StrategyMarketKR), us: readyRouteAuthority(StrategyMarketUS)}
	build := func(pair strategyProposalAuthorityPair) StrategyMarketWorker {
		return buildProductionStrategyMarketWorker(context.Background(), loader.clk, StrategyMarketKR, true, spy,
			loader.schedule, candidates, routes, loader.fx, pair, loader.risk, loader.accounts, func(context.Context) error { return nil })
	}
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)
	unactivated := proposals
	unactivated.kr = two
	if worker := build(unactivated); worker.Effective {
		t.Fatal("toggle OFF: an unactivated two-scope market promoted its worker past the market-wide capacity")
	}
	activated := unactivated
	activated.kr.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if worker := build(activated); !worker.Effective {
		t.Fatalf("activated two-scope market: worker=%+v, want Effective (per-scope handoffs)", worker)
	}
	// 읽기 전용 projection 도 같은 목록을 본다: 승격된 두 범위 시장은 「현재」이고 조정자 순서의 첫 승인 범위를 보인다(시장 단위
	// handoff 를 읽으면 상한 거절로 「현재」가 아니게 된다 — 리뷰 B 실측 상태 UNKNOWN).
	promoted := build(activated)
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Workers: []StrategyMarketWorker{promoted, {Market: StrategyMarketUS}},
		Clock: loader.clk})
	if err != nil {
		t.Fatal(err)
	}
	projected := strategyProjectionFromAssembly(StrategyEntryProductionAssembly{Supervisor: supervisor, proposals: activated})
	kr := projected.Markets[strategyprojection.Market(StrategyMarketKR)]
	if kr.Status != strategyprojection.StatusCurrent || kr.Campaign.ID == nil ||
		*kr.Campaign.ID != activated.kr.entries[0].authority.Proposal().Lineage.CampaignID {
		t.Fatalf("projection of the activated two-scope market: status=%q campaign=%v — want current, first admitted scope", kr.Status, kr.Campaign.ID)
	}
	// 이 fixture 의 위험 · 계좌 권한은 005930 범위에만 있다(000660 은 범위 권한 없음).
	spy.failEntryGateSymbol = map[string]error{"000660": errors.New("symbol entry blocked")}
	if worker := build(activated); !worker.Effective {
		t.Fatal("one scope's entry gate refusal starved the other scope's promotion")
	}
	// 리뷰 A #3: 승격 근거 범위는 자기 위험 · 계좌 권한이 준비돼 있어야 한다 — 005930(권한 있음)이 관문에 막히고 000660(관문 통과 · 권한 없음)만
	// 남으면 거래할 수 있는 범위가 0 이므로 dormant.
	spy.failEntryGateSymbol = map[string]error{"005930": errors.New("symbol entry blocked")}
	if worker := build(activated); worker.Effective {
		t.Fatal("promoted on a scope that has no risk or account authority of its own")
	}
	spy.failEntryGateSymbol["000660"] = errors.New("symbol entry blocked")
	if worker := build(activated); worker.Effective {
		t.Fatal("every scope refused by the entry gate, yet the worker was promoted")
	}
}

// 리뷰 A #3(만료 · digest): 두 범위가 모두 준비된 시장에서 worker 의 권한 만료는 준비된 계좌 범위들 중 **가장 이른** FreshUntil 이고, 계좌 식별
// 스냅숏은 범위 묶음이다(위험 BundleDigest 와 대칭).
func TestAWorkerOverTwoReadyScopesExpiresWithItsEarliestAccountScope(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{shortFreshFor: "000660"})
	now := fixture.now
	candidates := strategyCandidateAuthorityPair{observedAt: now, kr: readyCandidateAuthority(StrategyMarketKR), us: readyCandidateAuthority(StrategyMarketUS)}
	routes := strategyRouteAuthorityPair{observedAt: now, kr: readyRouteAuthority(StrategyMarketKR), us: readyRouteAuthority(StrategyMarketUS)}
	worker := buildProductionStrategyMarketWorker(context.Background(), fixture.clk, StrategyMarketKR, true, fixture.spy, fixture.loader.schedule,
		candidates, routes, fixture.fx, fixture.proposals, fixture.risk, fixture.accounts, func(context.Context) error { return nil })
	if !worker.Effective {
		t.Fatalf("arrangement: worker=%+v, want Effective", worker)
	}
	earliest := now.Add(30 * time.Second)
	if !worker.AuthorityExpiresAt.Equal(earliest) {
		t.Fatalf("AuthorityExpiresAt=%s, want the earliest ready account scope's FreshUntil %s", worker.AuthorityExpiresAt, earliest)
	}
	identities := make([]string, 0, 2)
	for _, scope := range fixture.accounts.kr.scopes {
		if scope.ready {
			identities = append(identities, scope.authority.Identity())
		}
	}
	if len(identities) != 2 || fixture.accounts.kr.snapshot.Identity != strategyWorkerEvidenceDigest(identities...) {
		t.Fatalf("account identity=%q over %d ready scopes, want the scope bundle digest", fixture.accounts.kr.snapshot.Identity, len(identities))
	}
}

// 위조는 주기를 멈춘다(J4 ② — 주문 경로 끝까지): 첫 handoff 가 같은 범위의 다른 캠페인(identity 불일치 — 위조 의심)이면 전달 몸통은
// 멈추고, 뒤의 **정상** 범위(000660)는 같은 주기에 나가지 않는다. 위조를 범위 거절로 오분류하는 변이(admit · dispatch · 수집 어느
// 자리든)는 000660 의 주문으로 드러난다.
func TestAForgedScopeStopsTheCycleBeforeTheNextValidScope(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	descriptor := riskLoaderDescriptor(t, StrategyMarketKR)
	forged, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, "acct-risk-loader", "005930", "campaign-forged-same-scope", 8,
		"100", "95", "120", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:forged-same-scope", map[string]strategyflow.Result{forged.Lineage.Symbol: forged})
	sealed, ok := batch.For(forged.Lineage.Symbol)
	if !ok {
		t.Fatal("arrangement: the forged proposal could not be sealed")
	}
	handoffs := strategyhandoff.AdmitEachOwnerScope(true, []strategyflow.Result{sealed.Proposal(), fixture.second.Proposal()})
	if len(handoffs) != 2 {
		t.Fatalf("arrangement: %d handoffs", len(handoffs))
	}
	err = dispatchStrategyMarketHandoffs(context.Background(), fixture.journal, fixture.cycle, handoffs)
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v err=%v — a forged scope must stop the cycle before the next scope", placed, err)
	}
	if scope := (*strategyScopeRefusal)(nil); err == nil || errors.As(err, &scope) || !strings.Contains(err.Error(), "production proposal identity changed") {
		t.Fatalf("err=%v — want the untyped identity refusal that stops the cycle", err)
	}
}

// ── a112 5.2.2.2 리뷰 수리(2026-10-01, J4 = (A)) ──────────────────────────────────────────────────────────────────────────
//
// 범위 거절은 **원천에서 온 신원**(riskbucket.ErrProductionRiskScopeRefused — 서명 정책 밖 종목 · scope latch)일 때만이다. 원장 결함 ·
// 무결성 · ctx 는 범위 칸이 비어 있어도 범위 거절이 아니라 주기를 멈추는 타입 없는 오류이고, 원인은 오류 사슬에 남는다.

func a112ScopeRefusalOf(err error) *strategyScopeRefusal {
	var refusal *strategyScopeRefusal
	if errors.As(err, &refusal) {
		return refusal
	}
	return nil
}

// 서명 위험 정책 밖의 종목(000660)은 그 범위만 거절되고 다른 범위는 거래한다 — 두 순서 모두(앞이어도 뒤 범위를 굶기지 않음).
func TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder(t *testing.T) {
	for _, order := range []struct {
		name  string
		first bool
	}{{"fixture order", false}, {"coordinator order (000660 first)", true}} {
		t.Run(order.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, a112TradingOptions{riskSymbols: []riskLoaderSymbol{}, coordinatorOrder: order.first})
			err := fixture.deliverKR(t)
			if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
				t.Fatalf("placed=%s err=%v, want only the scope inside the signed risk policy", got, err)
			}
			refusal := a112ScopeRefusalOf(err)
			if refusal == nil || refusal.scope.Symbol != "000660" || !errors.Is(err, riskbucket.ErrProductionRiskScopeRefused) {
				t.Fatalf("err=%v — want 000660's typed scope refusal carrying the policy's scope-refused identity", err)
			}
		})
	}
}

// E1(리뷰 A #1 · codex #2): 한 범위의 원장 버킷 행이 손상되면 그 범위의 위험 권한은 결함이다 — 범위 거절로 건너뛰고 다음 범위를 내지 않고,
// 주기를 멈추며(주문 0), 원인(원장 사용량 무효)을 오류 사슬에 남긴다.
func TestACorruptLedgerRowStopsTheCycleWithItsCause(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	stub, err := sql.Open("sqlite", "file:"+fixture.useRiskStub())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stub.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,account_ref,bucket_dimension,bucket_value,policy_version,snapshot_id,
		held_minor,filled_minor,state,risk_overage_latched,unknown_actual_latched) VALUES('corrupt-1','acct-risk-loader','symbol','005930','risk-policy-v1','',
		'-5','0','GARBAGE',0,0)`); err != nil {
		t.Fatal(err)
	}
	_ = stub.Close()
	fixture.wave(t)
	scopes := map[string]bool{}
	for _, scope := range fixture.risk.kr.scopes {
		scopes[scope.key.Symbol] = scope.ready
	}
	if scopes["005930"] || !scopes["000660"] {
		t.Fatalf("arrangement: risk scope readiness %v, want 005930 faulted and 000660 ready", scopes)
	}
	err = fixture.deliverKR(t)
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v err=%v — a ledger fault on one scope must stop the cycle, not skip to the next scope", placed, err)
	}
	if err == nil || a112ScopeRefusalOf(err) != nil {
		t.Fatalf("err=%v — want an untyped fault (not a scope refusal)", err)
	}
	if !errors.Is(err, riskbucket.ErrProductionRiskSnapshotUnavailable) || errors.Is(err, riskbucket.ErrProductionRiskScopeRefused) {
		t.Fatalf("err=%v — want the loader's cause preserved (snapshot unavailable, not scope-refused)", err)
	}
}

// scope latch 는 그 범위만 거절(신원 운반), latch 를 **읽지 못하는** 것은 결함(주기 멈춤) — 편집 전에는 둘이 한 오류였다.
func TestAScopeLatchIsRefusedAloneButALatchReadFaultStops(t *testing.T) {
	latched := newA112TradingFixture(t, a112TradingOptions{})
	stub, err := sql.Open("sqlite", "file:"+latched.useRiskStub())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stub.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation) VALUES('acct-risk-loader','KR','005930','1')`); err != nil {
		t.Fatal(err)
	}
	_ = stub.Close()
	latched.wave(t)
	err = latched.deliverKR(t)
	if got := strings.Join(latched.placedSymbols(), ","); got != "000660" {
		t.Fatalf("latched: placed=%s err=%v, want the unlatched scope only", got, err)
	}
	if refusal := a112ScopeRefusalOf(err); refusal == nil || refusal.scope.Symbol != "005930" || !errors.Is(err, riskbucket.ErrProductionRiskScopeRefused) {
		t.Fatalf("latched: err=%v — want 005930's typed scope refusal", err)
	}

	unreadable := newA112TradingFixture(t, a112TradingOptions{})
	stub, err = sql.Open("sqlite", "file:"+unreadable.useRiskStub())
	if err != nil {
		t.Fatal(err)
	}
	// 005930 의 latch 조회만 실패하게 함(정수 넘침으로 SELECT 가 오류) — 표를 지우면 두 범위가 함께 실패해 시장 전체가 준비 안 됨이 되어
	// 분류 갈래에 닿지 않는다(변이 Y03 이 그렇게 살아남았다).
	for _, statement := range []string{`DROP TABLE risk_bucket_scope_latches`,
		`CREATE VIEW risk_bucket_scope_latches AS SELECT 'acct-risk-loader' AS account_ref, 'KR' AS market, s.symbol AS symbol,
			'1' AS prospective_generation FROM (SELECT '005930' AS symbol UNION SELECT '000660') s
			WHERE CASE WHEN s.symbol = '005930' THEN abs(-9223372036854775808) ELSE 0 END`} {
		if _, err := stub.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	_ = stub.Close()
	unreadable.wave(t)
	readiness := map[string]bool{}
	for _, scope := range unreadable.risk.kr.scopes {
		readiness[scope.key.Symbol] = scope.ready
	}
	if readiness["005930"] || !readiness["000660"] {
		t.Fatalf("arrangement: risk readiness %v, want only 005930's latch read to fail", readiness)
	}
	err = unreadable.deliverKR(t)
	if placed := unreadable.placedSymbols(); len(placed) != 0 || err == nil || a112ScopeRefusalOf(err) != nil ||
		errors.Is(err, riskbucket.ErrProductionRiskScopeRefused) {
		t.Fatalf("latch unreadable: placed=%v err=%v — want no order and an untyped fault", placed, err)
	}
}

// 계좌 쪽 경계(조건 ④): 적재가 ctx 로 끝난 범위는 결함이다(주기 멈춤) — 평문 적재 실패(서명 매니페스트 부재)만 범위 거절.
func TestAnAccountLoadCancelledByItsContextStopsTheCycle(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{failAccountFor: "005930", accountFailure: context.Canceled})
	err := fixture.deliverKR(t)
	if placed := fixture.placedSymbols(); len(placed) != 0 || err == nil || a112ScopeRefusalOf(err) != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("placed=%v err=%v — want no order, an untyped fault and the context cause preserved", placed, err)
	}
}

// 리뷰 B #1(M20 격추): dispatch 가 lease 에 적는 위험 정책 세대는 **그 범위의 번들** 세대다 — 시장 칸 번들을 다른 세대(2)로 바꿔 놓아도 1.
func TestTheLeaseRiskGenerationComesFromTheScopesOwnBundle(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	other := newStrategyRiskLoaderFixtureGeneration(t, []riskLoaderSymbol{{Symbol: "000660", Sector: "technology", SectorLimitMinor: "3000000",
		SymbolLimitMinor: "2000000"}}, 2)
	otherPair := other.loader.collect(context.Background(), fixture.proposals.ResultAuthority(), other.fx)
	if !otherPair.kr.snapshot.Ready || otherPair.kr.bundle.Generation() != 2 {
		t.Fatalf("arrangement: generation-2 market bundle ready=%v generation=%d", otherPair.kr.snapshot.Ready, otherPair.kr.bundle.Generation())
	}
	fixture.risk.kr.bundle = otherPair.kr.bundle // 시장 칸만 세대 2 — 범위 번들은 세대 1 그대로
	fixture.cycle.risk.kr.bundle = otherPair.kr.bundle
	_ = fixture.deliverKR(t)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("arrangement: placed=%s, want the first scope issued", got)
	}
	db, err := sql.Open("sqlite", "file:"+fixture.journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var generation uint64
	if err := db.QueryRow(`SELECT risk_policy_generation FROM strategy_dispatch_market_authorities WHERE symbol='005930'`).Scan(&generation); err != nil {
		t.Fatal(err)
	}
	if generation != 1 {
		t.Fatalf("lease risk_policy_generation=%d, want 1 — the scope's own bundle, not the market slot", generation)
	}
}

// 계좌 적재기는 ctx 종료를 자기 오류로 접는다(생산 `strategyaccount.LoadProductionAuthority` 는 ctx.Err 를 ErrProductionAccountUnavailable 로
// 돌려줌) — 그래서 계좌 권한 수집은 실패 뒤 ctx 를 **직접** 보고 원인을 ctx 로 바꾼다(조건 ④). 취소된 ctx 로 모으면 실패한 범위의 원인은
// context.Canceled 다(원인 보존 — 1차 레그는 계좌 실패를 모두 결함으로 다룬다).
func TestAnAccountLoadThatFailsUnderACancelledContextIsAFault(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	accounts := a112AccountLoaderWith(t, fixture.now, "005930", strategyaccount.ErrProductionAccountUnavailable, "").collect(cancelled, fixture.proposals)
	for _, scope := range accounts.kr.scopes {
		if scope.key.Symbol != "005930" {
			continue
		}
		if scope.ready || !errors.Is(scope.cause, context.Canceled) {
			t.Fatalf("005930 account scope ready=%v cause=%v — want the context's cancellation as the cause", scope.ready, scope.cause)
		}
		return
	}
	t.Fatal("arrangement: no 005930 account scope")
}

// 리뷰 A #3(축별): 승격 근거 범위는 위험 · 계좌 권한을 **둘 다** 가져야 한다. 005930 이 관문에 막히고 000660 이 한쪽 권한만 가지면 dormant.
func TestAWorkerPromotesOnlyOnAScopeWithBothAuthorities(t *testing.T) {
	for _, missing := range []struct {
		name    string
		options a112TradingOptions
	}{
		{"000660 without risk authority", a112TradingOptions{riskSymbols: []riskLoaderSymbol{}}},
		{"000660 without account authority", a112TradingOptions{failAccountFor: "000660"}},
	} {
		t.Run(missing.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, missing.options)
			now := fixture.now
			candidates := strategyCandidateAuthorityPair{observedAt: now, kr: readyCandidateAuthority(StrategyMarketKR), us: readyCandidateAuthority(StrategyMarketUS)}
			routes := strategyRouteAuthorityPair{observedAt: now, kr: readyRouteAuthority(StrategyMarketKR), us: readyRouteAuthority(StrategyMarketUS)}
			build := func() StrategyMarketWorker {
				return buildProductionStrategyMarketWorker(context.Background(), fixture.clk, StrategyMarketKR, true, fixture.spy, fixture.loader.schedule,
					candidates, routes, fixture.fx, fixture.proposals, fixture.risk, fixture.accounts, func(context.Context) error { return nil })
			}
			if worker := build(); !worker.Effective {
				t.Fatalf("arrangement: worker=%+v, want Effective on 005930", worker)
			}
			fixture.spy.failEntryGateSymbol = map[string]error{"005930": errors.New("symbol entry blocked")}
			if worker := build(); worker.Effective {
				t.Fatal("promoted on a scope missing one of its own authorities")
			}
		})
	}
}
