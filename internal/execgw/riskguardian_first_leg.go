package execgw

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// QFinalCampaignFirstLegIssuance is the strategy-only input to the atomic
// first-leg path. It intentionally carries no prospective token, router
// selector, dispatch lease or Gateway capability.
type QFinalCampaignFirstLegIssuance struct {
	Entry                    QFinalEntryIssuance
	Result                   strategyflow.Result
	ActivationManifestDigest string
	AttemptID                string
	Revision                 int
	Campaign                 journal.FirstLegCampaignRequest
	Weekly                   *journal.WeeklyFirstLegReservationBinding
}

// QFinalCampaignFirstLegPrecheck is opaque. It seals the exact q_final result
// together with the strategy/campaign values the journal will bind atomically.
type QFinalCampaignFirstLegPrecheck struct {
	entry    QFinalEntryPrecheck
	strategy journal.StrategyPlanRequest
	campaign journal.FirstLegCampaignRequest
	weekly   *journal.WeeklyFirstLegReservationBinding
	issuedAt time.Time
}

func (p QFinalCampaignFirstLegPrecheck) QCandidate() uint64 { return p.entry.QCandidate() }
func (p QFinalCampaignFirstLegPrecheck) QFinal() uint64     { return p.entry.QFinal() }

// PrecheckQFinalCampaignFirstLeg performs no journal mutation and no broker
// call. The journal remains the final exact lineage validator; these checks
// reject obvious cross-family or caller-token substitutions before collection.
//
// a157(사람 결정 2026-10-18 「1-1」): 첫 다리 고유의 결속 판정 일곱(B2~B8)은 q_final 을 쓰지 않으므로 진입 판정(사이징 · 버킷 · 체인 · 노출)
// **앞**에서 요청 값으로 함. 그래야 Guardian 상한 0 · 버킷 고갈 같은 진입 거절이 결속 결함을 가리지 않음. 보고 우선순위는
// 일일 손실 체인 거절 > 결속 결함 > 그 밖의 진입 거절 — 일일 손실이 이기는 것은 a137 강화(ENTRY_BLOCKED 영속)와 계좌 단위 진입 차단을
// 겹친 요청에서도 지키기 위함. 결속 판정식 · 코드 · field · 문구는 진입 판정과 독립이고, 받아들이는 요청은 두 판정을 모두 통과한 것
// (두 판정의 논리곱)이므로 판정 순서가 받아들이는 집합을 바꾸지 않음 — 순서는 둘 다 걸린 요청의 보고만 정함(design D2).
func (g *RiskGuardian) PrecheckQFinalCampaignFirstLeg(request QFinalCampaignFirstLegIssuance) (QFinalCampaignFirstLegPrecheck, error) {
	// 결속 판정이 g.policyVersion · g.accountRef 를 읽으므로 그 전에 막음 — PrecheckQFinalEntry 의 nil 거절과 같은 거절.
	if g == nil {
		return QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid, "guardian", "Guardian is nil", nil)
	}
	lineage := request.Result.Lineage
	terms := request.Result.ExecutionTerms
	// 결속 판정 B2~B8 — B2 → B8 순서로 보고 첫 결함에서 멈춤(닫힌 함수의 return). 이 닫힌 함수는 정의하는 자리에서 꼭 한 번 즉시
	// 호출되고 진입 판정보다 앞임(구조 시험이 고정). 입력은 요청 값: PrecheckQFinalEntry 는 TransactionID · Owner(AccountID 공백 제거 밖)와
	// q_candidate 를 바꾸지 않고 통과시키므로 진입 판정이 본 값과 같음(design F3 · F4). 시계 · 원장 · 브로커를 읽지 않는 순수 판정임(F2).
	var policyVersion string
	var weekly *journal.WeeklyFirstLegReservationBinding
	bindErr := func() error {
		version, err := journal.QFinalPolicyVersion(g.policyVersion, request.Entry.Admission.TransactionID)
		if err != nil {
			return err
		}
		policyVersion = version
		owner := request.Entry.Admission.Owner
		if strings.TrimSpace(owner.Key.ProspectiveGeneration) != "" {
			return qFinalRefusal(riskbucket.RefusalOwnerConflict, "prospective_generation", "first-leg token is journal-owned", nil)
		}
		entryMajor, entryOK := terms.Entry().MajorDecimal()
		stopMajor, stopOK := terms.EffectiveStop().MajorDecimal()
		targetMajor, targetOK := terms.Target().MajorDecimal()
		if !entryOK || !stopOK || !targetOK {
			return qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
				"first_leg_price_unit", "sealed strategy execution price unit is invalid", nil)
		}
		if request.Result.Quantity != request.Entry.QCandidate || lineage.AccountRef != g.accountRef ||
			!strings.EqualFold(string(lineage.Market), request.Entry.Market) || lineage.Symbol != request.Entry.Symbol ||
			entryMajor != request.Entry.EntryPrice || stopMajor != request.Entry.StopPrice ||
			targetMajor != request.Entry.TargetPrice || lineage.LaneID != owner.LaneID ||
			lineage.CampaignID != owner.CampaignID || request.Campaign.CampaignID != owner.CampaignID {
			return qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
				"first_leg_lineage", "sealed strategy result, q_candidate, owner and campaign are not exactly bound", nil)
		}
		currentGeneration := request.Campaign.ExpectedPositionGeneration
		if currentGeneration < 0 || currentGeneration >= math.MaxInt64 || request.Campaign.ExpectedPositionVersion < 0 ||
			lineage.PositionGeneration != uint64(currentGeneration)+1 {
			return qFinalRefusal(riskbucket.RefusalOwnerConflict,
				"position_generation", "strategy prospective generation must be exactly campaign current generation plus one", nil)
		}
		if strings.TrimSpace(request.AttemptID) == "" || request.AttemptID != strings.TrimSpace(request.AttemptID) || request.Revision < 1 ||
			strings.TrimSpace(request.ActivationManifestDigest) == "" || request.ActivationManifestDigest != strings.TrimSpace(request.ActivationManifestDigest) {
			return qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
				"first_leg_strategy", "complete strategy attempt and activation binding required", nil)
		}
		binding, err := guardianWeeklyFirstLegBinding(lineage.LaneID, string(lineage.Market), request.Weekly)
		if err != nil {
			return qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
				"weekly_reservation", "weekly first-leg reservation authority is missing or does not match the sealed lane", err)
		}
		weekly = binding
		return nil
	}()
	// 진입 판정은 결속 결과와 무관하게 정확히 한 번 — 일일 손실 측정 토큰 판정이 여기 있으므로 결속 결함이 있어도 돌려야 우선순위를 가를 수 있음.
	entry, err := g.PrecheckQFinalEntry(request.Entry)
	// 진입 거절을 보고하는 경우: 결속 결함이 없거나, 진입 거절이 일일 손실 체인 거절(한도 · 측정 불가)임(1-1 — 첫 다리 승인과 같은 판별).
	if err != nil && (bindErr == nil || DailyLossRefusalKindOf(err) != DailyLossRefusalKindNone) {
		return QFinalCampaignFirstLegPrecheck{}, err
	}
	// 그 밖의 진입 거절(상한 0 · 버킷 고갈 · 다른 체인 거절 …)과 겹치거나 진입 판정을 통과했으면 결속 결함을 보고함.
	if bindErr != nil {
		return QFinalCampaignFirstLegPrecheck{}, bindErr
	}
	issuedAt := g.clk.Now().UTC()
	riskIntent := journal.RiskIntent{
		AccountRef: g.accountRef, Market: entry.request.Market, Symbol: entry.intent.Symbol,
		Side: string(risk.SideBuy), Quantity: strconv.FormatUint(entry.decision.QFinal, 10),
		EntryPrice: entry.intent.LimitPrice, StopPrice: entry.intent.StopPrice,
		TargetPrice: entry.intent.TargetPrice, PolicyVersion: policyVersion,
	}
	projected, err := journal.ProjectAcceptedStrategyflowLineage(journal.StrategyflowLineageProjectionRequest{
		Result: request.Result, RiskIntent: riskIntent, ActivationManifestDigest: request.ActivationManifestDigest, CreatedAt: issuedAt,
	})
	if err != nil {
		return QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
			"first_leg_projection", "sealed strategy result cannot be bound to Guardian q_final", err)
	}
	strategy := journal.StrategyPlanRequest{Lineage: projected, AttemptID: request.AttemptID,
		ActivationManifestDigest: request.ActivationManifestDigest, Revision: request.Revision, CreatedAt: issuedAt}
	return QFinalCampaignFirstLegPrecheck{entry: entry, strategy: strategy, campaign: request.Campaign, weekly: weekly, issuedAt: issuedAt}, nil
}

// IssuePrecheckedQFinalCampaignFirstLeg calls the atomic journal first-leg API
// directly. It never calls standalone q_final issuance and cannot mint a lease
// or reach the Gateway.
func (g *RiskGuardian) IssuePrecheckedQFinalCampaignFirstLeg(ctx context.Context, precheck QFinalCampaignFirstLegPrecheck) (journal.QFinalCampaignFirstLegReceipt, error) {
	if g == nil || precheck.entry.decision.QFinal == 0 || precheck.entry.request.Collect == nil ||
		precheck.entry.admission.TransactionID == "" {
		return journal.QFinalCampaignFirstLegReceipt{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,
			"first_leg_precheck", "complete sealed first-leg precheck required", nil)
	}
	now := g.clk.Now().UTC()
	if precheck.issuedAt.IsZero() || now.Before(precheck.issuedAt) || !now.Before(precheck.issuedAt.Add(g.ttl)) {
		return journal.QFinalCampaignFirstLegReceipt{}, qFinalRefusal(riskbucket.RefusalStaleBucket,
			"first_leg_precheck", "sealed first-leg precheck expired or clock moved backwards", nil)
	}
	accountFX, err := qFinalAccountBaseFXAt(precheck.entry.request, now, precheck.entry.market, g.policy)
	legacyTestIdentity := err != nil && precheck.entry.request.testFXAuthoritySet &&
		precheck.entry.market == costs.MarketKR && precheck.entry.accountFX.Digest() == ""
	if (err != nil && !legacyTestIdentity) || (!legacyTestIdentity && !sameAccountBaseFX(precheck.entry.accountFX, accountFX)) {
		return journal.QFinalCampaignFirstLegReceipt{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved,
			"guardian_account_base_fx", "sealed account-base FX authority expired or changed", err)
	}
	policyVersion, err := journal.QFinalPolicyVersion(g.policyVersion, precheck.entry.admission.TransactionID)
	if err != nil {
		return journal.QFinalCampaignFirstLegReceipt{}, err
	}
	limitsJSON := g.limitsJSON
	if !legacyTestIdentity {
		limitsJSON, err = encodeQFinalLimits(g.limits, accountFX)
		if err != nil {
			return journal.QFinalCampaignFirstLegReceipt{}, err
		}
	}
	decision := journal.DecisionRequest{
		ID: g.newID(), AccountRef: g.accountRef, Generation: 0,
		SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: g.accountRef, Market: precheck.entry.request.Market, Symbol: precheck.entry.intent.Symbol,
			Side: string(risk.SideBuy), Quantity: precheck.entry.intent.Quantity, EntryPrice: precheck.entry.intent.LimitPrice,
			StopPrice: precheck.entry.intent.StopPrice, TargetPrice: precheck.entry.intent.TargetPrice, PolicyVersion: policyVersion,
		},
		LimitsJSON: limitsJSON, Nonce: g.newID(), IssuedAt: precheck.issuedAt, ExpiresAt: precheck.issuedAt.Add(g.ttl),
	}
	reservationID := "res-" + decision.ID
	admission := precheck.entry.admission
	admission.Admission.Policy.EvaluatedAt = now
	policy, err := qFinalPolicyAt(precheck.entry.request, now, admission.Admission.Policy)
	if err != nil {
		return journal.QFinalCampaignFirstLegReceipt{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved,
			"official_fx", "sealed current FX authority expired or changed", err)
	}
	if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {
		return journal.QFinalCampaignFirstLegReceipt{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved,
			"guardian_account_base_fx", "Guardian and monetary reservations no longer share one frozen FX authority", nil)
	}
	admission.Admission.Policy = policy
	admission.DecisionID = decision.ID
	admission.ExistingReservationID = reservationID
	admission.CreatedAt = precheck.issuedAt

	strategy := precheck.strategy
	collect := func(ctx context.Context, attempt int) (journal.QFinalCampaignFirstLegRequest, error) {
		snapshot, err := precheck.entry.request.Collect(ctx, attempt)
		if err != nil {
			return journal.QFinalCampaignFirstLegRequest{}, err
		}
		exposure, verdict := risk.EntryExposureValue(risk.Input{
			Now: now, Intent: precheck.entry.intent, Account: precheck.entry.request.Account,
			Policy: g.policy, Costs: g.costs, AccountBaseFX: accountFX,
		})
		if !verdict.Allowed {
			return journal.QFinalCampaignFirstLegRequest{}, chainRefusal(verdict)
		}
		usage, err := exposureUsage(snapshot.OpenExposure, exposure.Currency)
		if err != nil {
			return journal.QFinalCampaignFirstLegRequest{}, err
		}
		daily, err := g.dailyBuyLimit(now, qFinalManualBuyFX(precheck.entry.request, precheck.entry.market))
		if err != nil {
			return journal.QFinalCampaignFirstLegRequest{}, err
		}
		return journal.QFinalCampaignFirstLegRequest{
			Issue: journal.QFinalIssueRequest{
				Issue: journal.IssueRequest{Decision: decision, Reserve: journal.ReserveRequest{
					SnapshotAsOf: snapshot.AsOf, ObservedVersion: snapshot.Version,
					SnapshotUsage:     []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: usage, Currency: exposure.Currency}},
					Limits:            []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: g.policy.MaxOpenExposure.Amount, Currency: exposure.Currency}},
					Reservations:      []journal.ReservationRequest{{ID: reservationID, Kind: journal.ReservationKindOpenExposure, Amount: exposure.Amount, Currency: exposure.Currency}},
					DailyBuyLimit:     daily,
					OpenPositionLimit: g.openPositionLimit(),
				}}, Admission: admission,
			},
			Strategy: strategy, Campaign: precheck.campaign, Weekly: cloneGuardianWeeklyBinding(precheck.weekly),
			RouterID: strategyrouter.RouterID, RouterVersion: strategyrouter.RouterRelease,
		}, nil
	}
	receipt, err := g.journal.RecordQFinalCampaignFirstLegWithRecollection(ctx, collect, g.recollect)
	if err != nil {
		return journal.QFinalCampaignFirstLegReceipt{}, fmt.Errorf("execgw: issuing atomic q_final first leg: %w", err)
	}
	return receipt, nil
}

func guardianWeeklyFirstLegBinding(laneID, market string, binding *journal.WeeklyFirstLegReservationBinding) (*journal.WeeklyFirstLegReservationBinding, error) {
	wantWeekly := (market == "KR" && laneID == weeklyvaluelane.KRWeeklyLaneID) ||
		(market == "US" && laneID == weeklyvaluelane.USWeeklyLaneID)
	if !wantWeekly {
		if binding != nil {
			return nil, journal.ErrWeeklyReservationConflict
		}
		return nil, nil
	}
	if binding == nil || binding.PlannedOrdinal != 1 || binding.ScopeVersion == 0 {
		return nil, journal.ErrWeeklyReservationConflict
	}
	for _, value := range []string{binding.ReservationID, binding.StableWeek, binding.RequestDigest, binding.RecordDigest,
		binding.CalendarGeneration, binding.CalendarDigest} {
		if value == "" || value != strings.TrimSpace(value) || len(value) > 256 {
			return nil, journal.ErrWeeklyReservationConflict
		}
	}
	return cloneGuardianWeeklyBinding(binding), nil
}

func cloneGuardianWeeklyBinding(binding *journal.WeeklyFirstLegReservationBinding) *journal.WeeklyFirstLegReservationBinding {
	if binding == nil {
		return nil
	}
	copy := *binding
	return &copy
}
