package journal

// a066 task 5.5 — 진입 손실 잠금(entry loss lock)의 journal 쪽 계약.
//
// 판정 규칙은 refuseEntryUnderLossLock 하나이고 부르는 자리는 셋임(두 admission 트랜잭션과
// Gateway 재검증). 아래 표는 세 자리 각각에서 호출자가 넘기는 인자(계좌·시장·horizon)를 따로
// 흔드는 잠금을 걸어 봄 — 한 자리가 인자를 잘못 배관하면(상수·뒤바꿈·다른 값) 그 자리의 행이 빨개짐.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

var lossLockAt = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

// lossLockVariant 는 잠글 결정의 시장·horizon 한 쌍임. KR/SHORT 와 US/MEDIUM 둘을 써서
// 시장·horizon 인자가 상수로 굳어도 드러나게 함.
type lossLockVariant struct {
	market  riskbucket.Market
	horizon riskbucket.Horizon
}

// applyLossLockVariant 는 KR/SHORT 기본 fixture 를 변형으로 옮김(시장·종목·통화·horizon bucket 재봉인).
func applyLossLockVariant(t *testing.T, plan *RiskBucketAdmissionPlan, variant lossLockVariant) {
	t.Helper()
	if variant.market == riskbucket.MarketUS {
		plan.Owner.Key.Market = riskbucket.MarketUS
		plan.Owner.Key.Symbol = "AAPL"
		plan.Admission.Policy.QuoteCurrency = "USD"
		rebindRiskBucket(t, plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
		rebindRiskBucket(t, plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	}
	if variant.horizon != riskbucket.HorizonShort {
		rebindRiskBucket(t, plan, 0, riskbucket.BucketKey{Dimension: riskbucket.DimensionHorizon, Value: string(variant.horizon), PolicyVersion: "policy-v1"})
	}
}

func otherMarket(market riskbucket.Market) riskbucket.Market {
	if market == riskbucket.MarketKR {
		return riskbucket.MarketUS
	}
	return riskbucket.MarketKR
}

func otherHorizon(horizon riskbucket.Horizon) riskbucket.Horizon {
	if horizon == riskbucket.HorizonShort {
		return riskbucket.HorizonMedium
	}
	return riskbucket.HorizonShort
}

func activateLossLockForTest(t *testing.T, j *Journal, account string, market riskbucket.Market, horizon riskbucket.Horizon) {
	t.Helper()
	if _, _, err := j.ActivateEntryLossLock(context.Background(), EntryLossLock{
		AccountRef: account, Market: market, Horizon: horizon, Cause: "a066 5.5 test", ActivatedAt: lossLockAt,
	}); err != nil {
		t.Fatalf("activate %s/%s/%s: %v", account, market, horizon, err)
	}
}

func TestEntryLossLockRefusesEveryEntrySiteOnlyInItsOwnScope(t *testing.T) {
	type lockCase struct {
		name    string
		scope   func(v lossLockVariant) (string, riskbucket.Market, riskbucket.Horizon)
		refused bool
	}
	locks := []lockCase{
		{"no lock", nil, false},
		{"own scope", func(v lossLockVariant) (string, riskbucket.Market, riskbucket.Horizon) {
			return "acct-1", v.market, v.horizon
		}, true},
		{"other horizon", func(v lossLockVariant) (string, riskbucket.Market, riskbucket.Horizon) {
			return "acct-1", v.market, otherHorizon(v.horizon)
		}, false},
		{"other market", func(v lossLockVariant) (string, riskbucket.Market, riskbucket.Horizon) {
			return "acct-1", otherMarket(v.market), v.horizon
		}, false},
		{"other account", func(v lossLockVariant) (string, riskbucket.Market, riskbucket.Horizon) {
			return "acct-2", v.market, v.horizon
		}, false},
	}
	variants := []lossLockVariant{
		{riskbucket.MarketKR, riskbucket.HorizonShort},
		{riskbucket.MarketUS, riskbucket.HorizonMedium},
	}
	// 세 호출 자리. 각각 잠금을 건 뒤 진입을 시도하고 (거절 오류, 남은 결정 행 수)를 돌려줌.
	sites := []struct {
		name string
		run  func(t *testing.T, j *Journal, v lossLockVariant, lock func()) (error, int)
	}{
		{"CommitRiskBucketAdmission", func(t *testing.T, j *Journal, v lossLockVariant, lock func()) (error, int) {
			seedExistingRiskReservation(t, j, "existing-lossl", "acct-1")
			plan := riskBucketAdmissionFixture(t, "lossl", "acct-1", "lane-a", "campaign-lossl", "prospective-lossl", "100", "0")
			applyLossLockVariant(t, &plan, v)
			lock()
			_, err := j.CommitRiskBucketAdmission(context.Background(), plan)
			return err, countRiskBucketRows(t, j, "risk_bucket_final_decisions") + countRiskBucketRows(t, j, "risk_bucket_owners")
		}},
		{"RecordQFinalDecisionAndReserve", func(t *testing.T, j *Journal, v lossLockVariant, lock func()) (error, int) {
			request := qFinalIssueFixture(t, j, "lossl")
			applyLossLockVariant(t, &request.Admission, v)
			if v.market == riskbucket.MarketUS {
				intent := request.Issue.Decision.Preimage.(RiskIntent)
				intent.Market, intent.Symbol = "us", "AAPL"
				request.Issue.Decision.Preimage = intent
			}
			lock()
			_, err := j.RecordQFinalDecisionAndReserve(context.Background(), request)
			return err, countRiskBucketRows(t, j, "decisions") + countRiskBucketRows(t, j, "risk_reservations") +
				countRiskBucketRows(t, j, "risk_bucket_final_decisions") + countRiskBucketRows(t, j, "risk_bucket_owners")
		}},
		{"RevalidateQFinalAdmission", func(t *testing.T, j *Journal, v lossLockVariant, lock func()) (error, int) {
			request := qFinalIssueFixture(t, j, "lossl")
			applyLossLockVariant(t, &request.Admission, v)
			if v.market == riskbucket.MarketUS {
				intent := request.Issue.Decision.Preimage.(RiskIntent)
				intent.Market, intent.Symbol = "us", "AAPL"
				request.Issue.Decision.Preimage = intent
			}
			// 결정 ⑤: 잠금 **전**에 발급된 결정.
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), request); err != nil {
				t.Fatalf("pre-lock issuance: %v", err)
			}
			lock()
			required, err := j.RevalidateQFinalAdmission(context.Background(), request.Issue.Decision.ID)
			if !required {
				t.Fatalf("q_final decision must require revalidation (err=%v)", err)
			}
			return err, -1
		}},
	}
	for _, site := range sites {
		for _, v := range variants {
			for _, lc := range locks {
				t.Run(site.name+"/"+string(v.market)+"_"+string(v.horizon)+"/"+lc.name, func(t *testing.T) {
					j := openTestJournal(t)
					lock := func() {
						if lc.scope != nil {
							account, market, horizon := lc.scope(v)
							activateLossLockForTest(t, j, account, market, horizon)
						}
					}
					err, rows := site.run(t, j, v, lock)
					if !lc.refused {
						if err != nil {
							t.Fatalf("entry outside the locked scope was refused: %v", err)
						}
						return
					}
					if !errors.Is(err, ErrRiskBucketEntryLossLocked) || !errors.Is(err, ErrRiskBucketEntryBlocked) ||
						!riskbucket.IsRefusal(err, riskbucket.RefusalEntryLossLockActive) {
						t.Fatalf("entry inside the locked scope was not refused by the loss lock: %v", err)
					}
					// 거절된 admission 은 기록 권한을 남기지 않음(재검증 자리는 읽기 전용이라 -1).
					if rows > 0 {
						t.Fatalf("refused admission left %d decision/owner/reservation rows", rows)
					}
				})
			}
		}
	}
}

func TestEntryLossLockActivationKeepsTheFirstCauseAndIsImmutable(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first, changed, err := j.ActivateEntryLossLock(ctx, EntryLossLock{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "first", ActivatedAt: lossLockAt.Add(123456789 * time.Nanosecond)})
	if err != nil || !changed || first.Seq <= 0 {
		t.Fatalf("first activation: %+v changed=%v err=%v", first, changed, err)
	}
	// 반환된 기록은 다시 읽은 기록과 같아야 함(저장 정밀도는 초).
	if stored, found, err := activeEntryLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); err != nil || !found || stored != first {
		t.Fatalf("returned lock %+v differs from the stored one %+v (found=%v err=%v)", first, stored, found, err)
	}
	again, changed, err := j.ActivateEntryLossLock(ctx, EntryLossLock{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "second", ActivatedAt: lossLockAt.Add(time.Hour)})
	if err != nil || changed || again.Seq != first.Seq || again.Cause != "first" || !again.ActivatedAt.Equal(lossLockAt) {
		t.Fatalf("repeated activation must return the first lock unchanged: %+v changed=%v err=%v", again, changed, err)
	}
	if n := countRiskBucketRows(t, j, "risk_bucket_entry_loss_locks"); n != 1 {
		t.Fatalf("lock rows=%d want 1", n)
	}

	invalid := []EntryLossLock{
		{Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "c", ActivatedAt: lossLockAt},
		// 앞뒤 공백 계좌: 기록되면 trim 된 계좌의 진입과 일치하지 않는 효력 없는 잠금이 됨.
		{AccountRef: " acct-1 ", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonMedium, Cause: "c", ActivatedAt: lossLockAt},
		{AccountRef: "acct-1\t", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonMedium, Cause: "c", ActivatedAt: lossLockAt},
		{AccountRef: "acct-1", Market: "JP", Horizon: riskbucket.HorizonShort, Cause: "c", ActivatedAt: lossLockAt},
		{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: "WEEKLY", Cause: "c", ActivatedAt: lossLockAt},
		{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonMedium, Cause: " ", ActivatedAt: lossLockAt},
		{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonMedium, Cause: "c"},
	}
	for i, lock := range invalid {
		if _, _, err := j.ActivateEntryLossLock(ctx, lock); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("invalid lock %d accepted: err=%v", i, err)
		}
	}
	if n := countRiskBucketRows(t, j, "risk_bucket_entry_loss_locks"); n != 1 {
		t.Fatalf("invalid activations wrote rows: %d", n)
	}

	// 원장 쪽 규율 — API 를 거치지 않는 쓰기도 이력을 지우거나 덮지 못함.
	for _, statement := range []string{
		`UPDATE risk_bucket_entry_loss_locks SET cause='rewritten'`,
		`DELETE FROM risk_bucket_entry_loss_locks`,
		`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1','KR','SHORT','raw','2026-03-30T00:30:00Z')`,
	} {
		if _, err := j.db.Exec(statement); err == nil {
			t.Fatalf("statement was not refused: %s", statement)
		}
	}
	// 다른 범위의 raw 삽입은 받아야 함 — 트리거가 범위를 넘어 막으면 잠금이 다른 범위로 번진 것임.
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1','KR','MEDIUM','raw','2026-03-30T00:30:00Z')`); err != nil {
		t.Fatalf("a different scope was refused by the first-cause trigger: %v", err)
	}
}

func TestEntryLossLockSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	activateLossLockForTest(t, j, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openTestJournalAt(t, path)
	seedExistingRiskReservation(t, reopened, "existing-reopen", "acct-1")
	plan := riskBucketAdmissionFixture(t, "reopen", "acct-1", "lane-a", "campaign-reopen", "prospective-reopen", "100", "0")
	if _, err := reopened.CommitRiskBucketAdmission(context.Background(), plan); !errors.Is(err, ErrRiskBucketEntryLossLocked) {
		t.Fatalf("lock did not survive a restart: %v", err)
	}
}

func TestEntryLossLockConcurrentActivationsLeaveOneLock(t *testing.T) {
	j := openTestJournalWithBusy(t, filepath.Join(t.TempDir(), "journal.db"), 5*time.Second)
	const writers = 8
	var wg sync.WaitGroup
	seqs := make([]int64, writers)
	changes := make([]bool, writers)
	errs := make([]error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lock, changed, err := j.ActivateEntryLossLock(context.Background(), EntryLossLock{
				AccountRef: "acct-1", Market: riskbucket.MarketUS, Horizon: riskbucket.HorizonMedium,
				Cause: "writer-" + strings.Repeat("x", i+1), ActivatedAt: lossLockAt,
			})
			seqs[i], changes[i], errs[i] = lock.Seq, changed, err
		}(i)
	}
	wg.Wait()
	winners := 0
	for i := 0; i < writers; i++ {
		if errs[i] != nil {
			t.Fatalf("writer %d: %v", i, errs[i])
		}
		if seqs[i] != seqs[0] {
			t.Fatalf("writers disagree on the lock: %v", seqs)
		}
		if changes[i] {
			winners++
		}
	}
	if winners != 1 || countRiskBucketRows(t, j, "risk_bucket_entry_loss_locks") != 1 {
		t.Fatalf("winners=%d rows=%d, want exactly one", winners, countRiskBucketRows(t, j, "risk_bucket_entry_loss_locks"))
	}
}

func TestRefuseEntryUnderLossLockFailsClosedOnUnknownScopeAndReadError(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	for _, scope := range []struct {
		account string
		market  riskbucket.Market
		horizon riskbucket.Horizon
	}{{"", riskbucket.MarketKR, riskbucket.HorizonShort}, {"acct-1", "", riskbucket.HorizonShort}, {"acct-1", riskbucket.MarketKR, ""}} {
		err := refuseEntryUnderLossLock(ctx, j.db, scope.account, scope.market, scope.horizon)
		if !errors.Is(err, ErrRiskBucketEntryBlocked) || errors.Is(err, ErrRiskBucketEntryLossLocked) {
			t.Fatalf("unknown scope %+v must be refused as blocked, not as locked: %v", scope, err)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := refuseEntryUnderLossLock(cancelled, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("a lock read failure must refuse entry: %v", err)
	}
	if err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); err != nil {
		t.Fatalf("no lock must admit: %v", err)
	}
}

func TestMigrationV32ToV33StartsWithNoEntryLossLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 32)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	j := openJournalAtSchema(t, path, 33)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 33 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	// 업그레이드 직후 잠긴 범위가 있으면 아무도 잠근 적 없는 진입이 조용히 멈춤.
	if n := countRiskBucketRows(t, j, "risk_bucket_entry_loss_locks"); n != 0 {
		t.Fatalf("migrated journal starts with %d locks", n)
	}
	// 판정 코드는 현재 스키마의 것이므로(a066 5.5 v35: 해제 기록 표를 읽음) 현재 버전으로 다시 열어 잼.
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	j = openJournalAtSchema(t, path, SchemaVersion)
	defer j.Close()
	if err := refuseEntryUnderLossLock(context.Background(), j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); err != nil {
		t.Fatalf("migrated journal refuses entry with no lock: %v", err)
	}
	activateLossLockForTest(t, j, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort)
}

// TestEntryLossLockRefusesTheStrategyFirstLegAdmission 는 생산 전략 경로(첫 leg 원자 admission,
// strategy_first_leg_atomic.go 가 commitFreshRiskBucketAdmissionTx 를 부르는 자리)가 잠금에 막힘을 고정함.
// 적대 리뷰(2026-09-27)가 이 호출 자리만 잠금을 건너뛰는 변이가 모든 시험을 통과한다고 보였음.
func TestEntryLossLockRefusesTheStrategyFirstLegAdmission(t *testing.T) {
	cases := []struct {
		name    string
		market  string
		symbol  string
		lock    []lossLockVariant
		refused bool
	}{
		{"KR no lock", "KR", "005930", nil, false},
		{"KR SHORT lock refuses KR", "KR", "005930", []lossLockVariant{{riskbucket.MarketKR, riskbucket.HorizonShort}}, true},
		{"KR MEDIUM lock does not reach KR SHORT", "KR", "005930", []lossLockVariant{{riskbucket.MarketKR, riskbucket.HorizonMedium}}, false},
		{"US SHORT lock refuses US", "US", "AAPL", []lossLockVariant{{riskbucket.MarketUS, riskbucket.HorizonShort}}, true},
		{"KR locks do not reach US", "US", "AAPL", []lossLockVariant{{riskbucket.MarketKR, riskbucket.HorizonShort}, {riskbucket.MarketKR, riskbucket.HorizonMedium}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			account := "acct-" + strings.ToLower(tc.market)
			request := firstLegAtomicFixture(t, j, "lossl-first-leg", account, tc.market, tc.symbol)
			for _, lock := range tc.lock {
				activateLossLockForTest(t, j, account, lock.market, lock.horizon)
			}
			_, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if !tc.refused {
				if err != nil {
					t.Fatalf("first leg outside the locked scope was refused: %v", err)
				}
				return
			}
			if !errors.Is(err, ErrRiskBucketEntryLossLocked) || !riskbucket.IsRefusal(err, riskbucket.RefusalEntryLossLockActive) {
				t.Fatalf("first leg inside the locked scope was not refused by the loss lock: %v", err)
			}
			for _, table := range []string{"decisions", "risk_reservations", "risk_bucket_final_decisions", "risk_bucket_owners", "strategy_first_leg_bindings", "campaign_legs"} {
				if n := countRiskBucketRows(t, j, table); n != 0 {
					t.Fatalf("refused first leg left %d rows in %s", n, table)
				}
			}
		})
	}
}

// TestEntryLossLockSchemaRefusesValuesTheRuleCannotMatch 는 v33 CHECK 제약을 고정함: 소문자 시장·horizon 이나
// 모르는 값이 원장에 들어가면 대문자로 묻는 규칙과 영영 일치하지 않는 **효력 없는 잠금**이 됨.
func TestEntryLossLockSchemaRefusesValuesTheRuleCannotMatch(t *testing.T) {
	j := openTestJournal(t)
	for _, row := range [][2]string{{"kr", "SHORT"}, {"KR", "short"}, {"JP", "SHORT"}, {"KR", "WEEKLY"}} {
		if _, err := j.db.Exec(`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1',?,?,'raw','2026-03-30T00:30:00Z')`, row[0], row[1]); err == nil {
			t.Fatalf("schema accepted market=%q horizon=%q", row[0], row[1])
		}
	}
	for _, statement := range []string{
		`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('','KR','SHORT','raw','2026-03-30T00:30:00Z')`,
		`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1','KR','SHORT','','2026-03-30T00:30:00Z')`,
	} {
		if _, err := j.db.Exec(statement); err == nil {
			t.Fatalf("schema accepted an empty account or cause: %s", statement)
		}
	}
}

// TestEntryLossLockAccountFormIsTheDecisionBuildersForm 는 잠금의 계좌 정본형이 결정·예약 builder 의 형태(trim,
// 대소문자 보존)와 같고 비교가 정확 일치임을 고정함. 대소문자만 다른 id 는 원장에서 다른 계좌이므로 그 잠금은
// "acct-1" 진입을 막지 않고, "acct-1" 잠금은 막음 — 계약이 바뀌면(예: 한쪽만 소문자화) 여기서 빨개짐.
func TestEntryLossLockAccountFormIsTheDecisionBuildersForm(t *testing.T) {
	ctx := context.Background()
	built, err := DecisionRequest{ID: "d", AccountRef: "  Acct-1 ", SafetyClass: SafetyClassExposureRaising, Kind: KindPlace,
		Preimage:   RiskIntent{AccountRef: "Acct-1", Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "1", EntryPrice: "5", StopPrice: "4", TargetPrice: "7", PolicyVersion: "p"},
		LimitsJSON: `{}`, Nonce: "n", IssuedAt: lossLockAt, ExpiresAt: lossLockAt.Add(time.Minute)}.build()
	if err != nil || built.AccountRef != "Acct-1" {
		t.Fatalf("decision builder account form changed (want trim, case kept): %q err=%v", built.AccountRef, err)
	}
	for _, tc := range []struct {
		lockAccount string
		refused     bool
	}{{"ACCT-1", false}, {"Acct-1", false}, {"acct-1", true}} {
		j := openTestJournal(t)
		activateLossLockForTest(t, j, tc.lockAccount, riskbucket.MarketKR, riskbucket.HorizonShort)
		err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort)
		if got := errors.Is(err, ErrRiskBucketEntryLossLocked); got != tc.refused {
			t.Fatalf("lock on %q vs entry on \"acct-1\": refused=%v want %v (err=%v)", tc.lockAccount, got, tc.refused, err)
		}
	}
}
