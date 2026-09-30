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
	"fmt"
	"path/filepath"
	"reflect"
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
	// failAccountFor 는 계좌 권한 적재를 실패시킬 종목.
	failAccountFor string
}

type a112TradingFixture struct {
	now        time.Time
	riskLoader strategyRiskAuthorityLoader
	fx         strategyFXAuthorityPair
	guardian   *execgw.RiskGuardian
	clk        *clock.Fake
	journal    *journal.Journal
	proposals  strategyProposalAuthorityPair
	risk       strategyRiskAuthorityPair
	accounts   strategyAccountAuthorityPair
	loader     *productionStrategyFirstLegAuthorityLoader
	cycle      *strategyDispatchCycle
	spy        *strategyDispatchGatewaySpy
	winner     strategyproposal.ProductionAuthority
	second     strategyproposal.ProductionAuthority
}

func newA112TradingFixture(t *testing.T, options a112TradingOptions) a112TradingFixture {
	t.Helper()
	symbols := options.riskSymbols
	if symbols == nil {
		symbols = []riskLoaderSymbol{{Symbol: "000660", Sector: "technology", SectorLimitMinor: "3000000", SymbolLimitMinor: "2000000"}}
	}
	riskFixture := newStrategyRiskLoaderFixtureWith(t, symbols)
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
	two := a112ExtraEntryKR(t, proposals.kr, now, "000660", false)
	two.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	proposals.kr = two
	second := two.entries[1].authority

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
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: handle, Clock: fakeClock, AccountRef: "acct-risk-loader",
		Policy: policy, Costs: costs.DefaultModel(), PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	// 위험 적재기는 stub 원장(user_version 27)을 읽는다 — 실제 원장(v35)은 생산 적재기의 스키마 핀이 거절한다(결함, 아래 다리 참고).
	riskLoader := *riskFixture.loader
	fixture := a112TradingFixture{now: now, riskLoader: riskLoader, fx: riskFixture.fx, guardian: guardian, clk: fakeClock, journal: handle,
		proposals: proposals, spy: &strategyDispatchGatewaySpy{observed: map[string]int{}}, winner: winner, second: second}
	fixture.accounts = a112AccountLoader(t, now, options.failAccountFor).collect(context.Background(), proposals)
	fixture.wave(t)
	return fixture
}

// a112MirrorLedgerIntoRiskStub 는 admission 원장의 **실제 행**을 위험 적재기의 stub 원장으로 옮기고, 옮긴 행이 원본과 같음을 단언함.
//
// 다리(임시)의 사유: 생산 위험 적재기가 journal schema 27 에 고정돼(riskbucket productionRiskJournalSchema) 실제 원장(v35)을 거절하므로
// 적재기가 admission 원장을 직접 못 읽음 — 그래서 사용량이 stub 에서 늘지 않아 둘째 파도가 영원히 BUCKET_USAGE_STALE 로 남음.
// 이 복사는 행을 지어내지 않음: 원장에서 읽은 것만 넣고 표마다 행 집합이 원본과 같은지 대조함(복사기 자체의 시험).
// 옮기는 표와 열은 손으로 고르지 않고 **stub 의 스키마에서 유도**함(stub 이 적재기가 읽는 투영 — a126 이 표를 더해도 따라감).
// 제거 조건: 스키마 핀 수리 change 가 착지하면 적재기를 실제 원장으로 단일화하고 이 다리를 지움.
func a112MirrorLedgerIntoRiskStub(t *testing.T, ledgerPath, stubPath string) map[string]int {
	t.Helper()
	ledger, err := sql.Open("sqlite", "file:"+ledgerPath+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.Close()
	stub, err := sql.Open("sqlite", "file:"+stubPath)
	if err != nil {
		t.Fatal(err)
	}
	defer stub.Close()
	tables := a112StubTables(t, stub)
	if len(tables) == 0 {
		t.Fatal("arrangement: the risk stub has no tables to mirror")
	}
	counts := map[string]int{}
	for table, columns := range tables {
		source := a112ReadRows(t, ledger, table, columns)
		if _, err := stub.Exec("DELETE FROM " + table); err != nil {
			t.Fatal(err)
		}
		width := len(strings.Split(columns, ","))
		placeholders := strings.TrimSuffix(strings.Repeat("?,", width), ",")
		for _, row := range source {
			if _, err := stub.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, columns, placeholders), row...); err != nil {
				t.Fatal(err)
			}
		}
		// 복사기 시험: stub 의 행 집합이 원장의 것과 정확히 같아야 함(누락 · 변형 · 잔여 없음).
		if copied := a112ReadRows(t, stub, table, columns); !reflect.DeepEqual(copied, source) {
			t.Fatalf("mirror of %s diverged from the admission ledger: copied=%v source=%v", table, copied, source)
		}
		counts[table] = len(source)
	}
	return counts
}

// a112StubTables 는 stub 원장의 표 → 열 목록(쉼표, 선언 순서).
func a112StubTables(t *testing.T, stub *sql.DB) map[string]string {
	t.Helper()
	rows, err := stub.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	rows.Close()
	tables := map[string]string{}
	for _, name := range names {
		info, err := stub.Query(`SELECT name FROM pragma_table_info(?) ORDER BY cid`, name)
		if err != nil {
			t.Fatal(err)
		}
		var columns []string
		for info.Next() {
			var column string
			if err := info.Scan(&column); err != nil {
				t.Fatal(err)
			}
			columns = append(columns, column)
		}
		info.Close()
		tables[name] = strings.Join(columns, ",")
	}
	return tables
}

func a112ReadRows(t *testing.T, db *sql.DB, table, columns string) [][]any {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", columns, table, columns))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	width := len(strings.Split(columns, ","))
	var out [][]any
	for rows.Next() {
		values := make([]any, width)
		pointers := make([]any, width)
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatal(err)
		}
		out = append(out, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
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
func a112AccountLoader(t *testing.T, now time.Time, failFor string) *strategyAccountAuthorityLoader {
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
			return strategyaccount.Authority{}, errors.New("account authority unavailable for " + config.Symbol)
		}
		quote, cash := "KRW", "5000000"
		if config.Market == strategyaccount.MarketUS {
			quote, cash = "USD", "1000"
		}
		state := risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{config.Symbol}, HeldQuantity: "0",
			CashAvailable: riskcalc.Money{Amount: cash, Currency: quote}, OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"}, AccountEquity: riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
		return strategyaccount.AuthorityForTest(config.Market, quote, state, now.Add(-time.Second), now.Add(time.Minute), 1, digest), nil
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
// Done 문장의 확정된 의미(2026-10-01 Manager 판정 ④ — 약화가 아니라 정밀화): **범위별 발급은 파도 순차**다. 두 범위는 horizon · 시장 ·
// 계좌 버킷을 공유하므로 첫 레그의 admission 이 공유 버킷 사용량을 올리고, 같은 파도에서 모은 둘째 범위의 버킷 스냅숏은 뒤처져 journal 이
// `BUCKET_USAGE_STALE` 로 거절한다(범위 거절이 아닌 원장 거절 — 주기가 멈춘다). 이것은 결함이 아니라 공유 버킷 이중 소비를 막는 설계다
// (관문 전수표 (e)). 둘째 파도가 첫 레그의 held 를 반영한 번들을 다시 모으면 둘째 범위가 발급된다 — 첫 범위는 캠페인이 이제 FLAT 이 아니라
// 전달 몸통이 건너뛴다(굶음 없음).
//
// 둘째 파도 전의 원장 → stub 복사는 스키마 핀 결함의 다리다(a112MirrorLedgerIntoRiskStub 주석 — 핀 수리 change 착지 시 제거).
func TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	// J2 실측: 활성 두 범위 파도의 읽기 수 = 위험 N + 계좌 N(범위마다 적재 하나 — 적재 호출은 각 적재기의 범위 순회 안 한 자리).
	if risk, account := len(fixture.risk.kr.scopes), len(fixture.accounts.kr.scopes); risk != 2 || account != 2 {
		t.Fatalf("reads per wave: risk=%d account=%d, want one per owner scope (2 + 2)", risk, account)
	}
	t.Logf("J2 measured: one activated two-scope KR wave reads risk=%d + account=%d owner-scope authorities",
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
	if counts := a112MirrorLedgerIntoRiskStub(t, fixture.journal.Path(), fixture.riskLoader.journalPath); counts["risk_bucket_reservations"] == 0 {
		t.Fatal("arrangement: the first leg left no bucket reservation in the admission ledger — the second wave would prove nothing")
	}
	fixture.wave(t)
	if err := fixture.deliverKR(t); err != nil {
		t.Fatalf("second wave err=%v, want the second scope issued on a fresh bundle", err)
	}
	if got := strings.Join(fixture.placedSymbols(), ","); got != "000660,005930" {
		t.Fatalf("placed=%s, want one first leg per owner scope (000660, 005930)", got)
	}
	// 각 범위는 자기 범위의 위험 번들로 발급됐다 — 범위의 위험 권한 digest 가 서로 다르고, 발급된 수만큼이다.
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

// 두 레그 한 주기(관문 전수표 (b) · J5 ②): 둘째 admission 은 첫 레그의 **held** 예약을 센다. 한 레그만 들어가는 계좌 노출 상한에서
// 둘째는 journal 합산(usage + held + new)으로 거절된다 — 브로커 스냅숏(계좌 권한의 OpenExposure=0)은 첫 레그를 모르는데도.
//
// 버킷 스냅숏 CAS 를 떼어 놓고 노출 합산만 재려고 **둘째 파도**(번들을 지금 원장으로 다시 모음)에서 잰다: 계좌 권한의 브로커 스냅숏은 여전히
// OpenExposure=0 인데도 둘째 레그는 첫 레그의 held 예약 때문에 계좌 노출 상한에 걸린다.
func TestTheSecondLegOfOneCycleCountsTheFirstLegsHeldReservation(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{maxOpenExposure: "1200"})
	versionBefore, err := fixture.journal.ReservationVersion(context.Background(), "acct-risk-loader")
	if err != nil {
		t.Fatal(err)
	}
	_ = fixture.deliverKR(t) // 첫 파도: 첫 범위 발급(둘째는 공유 버킷 CAS 로 거절 — 위 Done 시험)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("first wave placed=%s, want the first scope", got)
	}
	versionAfter, verr := fixture.journal.ReservationVersion(context.Background(), "acct-risk-loader")
	if verr != nil || versionAfter <= versionBefore {
		t.Fatalf("reservation version %d → %d (err=%v) — the first leg's reservation did not move the CAS version the next admission reads",
			versionBefore, versionAfter, verr)
	}
	if fixture.accounts.kr.authority.OpenExposure().Amount != "0" {
		t.Fatal("arrangement: the broker snapshot must not know the first leg")
	}
	if counts := a112MirrorLedgerIntoRiskStub(t, fixture.journal.Path(), fixture.riskLoader.journalPath); counts["risk_bucket_reservations"] == 0 {
		t.Fatal("arrangement: the first leg left no bucket reservation to mirror")
	}
	fixture.wave(t)
	err = fixture.deliverKR(t)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("second wave placed=%s err=%v — the held reservation of the first leg must keep the second under the exposure cap", got, err)
	}
	if err == nil || !strings.Contains(err.Error(), "open_exposure") && !strings.Contains(strings.ToLower(err.Error()), "exposure") {
		t.Fatalf("second wave err=%v, want the aggregate open-exposure refusal (usage + held + new > limit)", err)
	}
}

// 범위별 거절(J3): 한 범위의 계좌 권한을 못 얻으면 그 범위만 거절되고(봉투 폴백 없음 · 기록됨) 다른 범위는 거래한다.
func TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{failAccountFor: "005930"})
	err := fixture.deliverKR(t)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "000660" {
		t.Fatalf("placed=%s err=%v, want only the scope with its own account authority", got, err)
	}
	var refusal *strategyScopeRefusal
	if err == nil || !errors.As(err, &refusal) {
		t.Fatalf("err=%v, want the skipped scope's typed refusal returned (recorded, not silent)", err)
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

// 다리의 제거 조건을 기계로 건다(결함 영수증 · Manager 판정 ②): 생산 위험 적재기는 오늘 admission 원장(journal.SchemaVersion)을 스키마 핀
// 27 로 거절한다. 핀 수리 change 가 착지하면 이 시험이 실패한다 — 그때 a112MirrorLedgerIntoRiskStub 를 지우고 적재기를 실제 원장으로 단일화할 것.
func TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{})
	loader := fixture.riskLoader
	loader.journalPath = fixture.journal.Path()
	scoped := fixture.proposals.ResultAuthority().kr.results()
	if len(scoped) != 2 {
		t.Fatalf("arrangement: KR results=%d, want the two owner scopes", len(scoped))
	}
	_, err := riskbucket.LoadProductionRiskSnapshotAuthority(context.Background(), riskbucket.ProductionRiskSnapshotConfig{
		ConfigDir: loader.configDir, JournalPath: loader.journalPath, Market: riskbucket.MarketKR, AccountID: loader.accountID,
		AccountCurrency: loader.accountCurrency, ManifestDigest: loader.digests[StrategyMarketKR], TrustedKeyID: loader.keyID,
		TrustedKey: loader.key, ObservedAt: loader.observedAt,
	}, riskbucket.ProductionRiskSnapshotInput{Result: scoped[0], FX: fixture.fx.kr.read.evidence})
	if err == nil || !strings.HasSuffix(err.Error(), "risk bucket: exact journal schema unavailable") {
		t.Fatalf("real-journal load err=%v (journal schema %d) — if the schema pin was repaired, remove the stub bridge and read the real ledger",
			err, journal.SchemaVersion)
	}
	t.Logf("receipt: journal.SchemaVersion=%d, real-journal load refused: %v", journal.SchemaVersion, err)
	collected := loader.collect(context.Background(), fixture.proposals.ResultAuthority(), fixture.fx)
	for _, scope := range collected.kr.scopes {
		if scope.ready {
			t.Fatalf("scope %v ready against the real journal — the pin was repaired; remove the stub bridge", scope.key)
		}
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
	// handoff 를 읽으면 상한 거절로 EvidenceStale 이 된다).
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
	spy.failEntryGateSymbol = map[string]error{"005930": errors.New("symbol entry blocked")}
	if worker := build(activated); !worker.Effective {
		t.Fatal("one scope's entry gate refusal starved the other scope's promotion")
	}
	spy.failEntryGateSymbol["000660"] = errors.New("symbol entry blocked")
	if worker := build(activated); worker.Effective {
		t.Fatal("every scope refused by the entry gate, yet the worker was promoted")
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
