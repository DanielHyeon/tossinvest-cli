package execgw

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

const officialFXAuthoritySource = "official-fx"

// QFinalRefusal is a stable refusal from the market-generic q_final seam.
// QFinal is always zero: a refusal is never an order intent.
type QFinalRefusal struct {
	Code   riskbucket.RefusalCode
	Field  string
	Detail string
	cause  error
}

func (e *QFinalRefusal) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("execgw: q_final refused: %s (%s)", e.Code, e.Field)
	}
	return fmt.Sprintf("execgw: q_final refused: %s (%s): %s", e.Code, e.Field, e.Detail)
}

func (e *QFinalRefusal) Unwrap() error { return e.cause }

// QFinalEntryIssuance is the dormant market-generic Guardian seam consumed by
// the later durable KR/US dispatcher. It does not activate a lane or submit an
// order. The current legacy StrategyEntryIssuance remains the sealed KR Parker
// adapter and is intentionally not widened here.
type QFinalEntryIssuance struct {
	Market, Currency, Symbol string
	QCandidate               uint64
	EntryPrice, StopPrice    string
	TargetPrice              string
	Account                  risk.AccountState
	Collect                  CollectExposure
	Admission                journal.RiskBucketAdmissionPlan
	ExpectedPolicyVersion    string
	ExpectedLimitsDigest     string
}

// QFinalEntryPrecheck is opaque. Callers can inspect the result but cannot
// construct or change the exact request that final issuance consumes.
type QFinalEntryPrecheck struct {
	request     QFinalEntryIssuance
	market      costs.Market
	intent      risk.Intent
	admission   journal.RiskBucketAdmissionPlan
	decision    riskbucket.AdmissionDecision
	existingCap uint64
}

func (p QFinalEntryPrecheck) QCandidate() uint64 { return p.decision.QCandidate }
func (p QFinalEntryPrecheck) QFinal() uint64     { return p.decision.QFinal }

func (p QFinalEntryPrecheck) BindingCaps() []riskbucket.BucketKey {
	out := make([]riskbucket.BucketKey, len(p.decision.Binding))
	copy(out, p.decision.Binding)
	return out
}

// PrecheckQFinalEntry performs no journal mutation and no broker call. It
// calculates the existing Guardian cap, overwrites caller-supplied cap fields,
// calculates q_final, and runs the ordinary Guardian chain once on q_final.
func (g *RiskGuardian) PrecheckQFinalEntry(request QFinalEntryIssuance) (QFinalEntryPrecheck, error) {
	if g == nil {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid, "guardian", "Guardian is nil", nil)
	}
	if request.ExpectedPolicyVersion != g.policyVersion || request.ExpectedLimitsDigest != g.LimitsDigest() {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalStaleBucket, "guardian_snapshot", "policy or limits digest changed", nil)
	}
	marketName := strings.ToUpper(strings.TrimSpace(request.Market))
	currency := strings.ToUpper(strings.TrimSpace(request.Currency))
	market, exactCurrency, bucketMarket, err := qFinalMarketCurrency(marketName, currency)
	if err != nil {
		return QFinalEntryPrecheck{}, err
	}
	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	if symbol == "" || symbol != request.Symbol {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalUnknownSymbol, "symbol", "symbol is not canonical", nil)
	}
	if request.QCandidate == 0 {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalZeroQuantity, "q_candidate", "lane proposed zero", nil)
	}
	limitCurrency := g.policy.LimitCurrency()
	if limitCurrency != exactCurrency {
		// The existing Guardian chain does not convert its quantity/notional/cash
		// caps. Letting the monetary layer convert while that chain compares raw
		// USD prices to KRW amounts would fabricate headroom.
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved, "guardian_limit_currency", "existing Guardian limits are not in the market quote currency", nil)
	}
	// Seal every caller-owned slice before calculating authority. The opaque
	// precheck must remain a value snapshot even when its source DTO is reused or
	// mutated by another goroutine after this method returns.
	admission := cloneRiskBucketAdmissionPlan(request.Admission)
	admission.Admission.QCandidate = request.QCandidate
	// Freshness is judged at the Guardian clock, never at a caller-selected
	// evaluation instant that could make expired price/FX/snapshot evidence look
	// historical-but-valid.
	admission.Admission.Policy.EvaluatedAt = g.clk.Now().UTC()
	admission.Owner.Key.AccountID = strings.TrimSpace(admission.Owner.Key.AccountID)
	if admission.Owner.Key.AccountID != g.accountRef || admission.Owner.Key.Market != bucketMarket || admission.Owner.Key.Symbol != symbol {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalOwnerConflict, "owner", "owner does not match Guardian market scope", nil)
	}
	policy := admission.Admission.Policy
	if policy.AccountCurrency != limitCurrency || policy.QuoteCurrency != exactCurrency {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved, "reserve_currency", "reserve policy does not match Guardian/market currency", nil)
	}
	if bucketMarket == riskbucket.MarketUS && policy.AccountCurrency != policy.QuoteCurrency && policy.FX.Source != officialFXAuthoritySource {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalCurrencyUnresolved, "official_fx", "fresh official US FX authority is unresolved", nil)
	}
	guardianCapRaw, err := risk.StrategyEntryQuantity(g.policy, request.EntryPrice, request.StopPrice)
	if err != nil {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalExistingGuardianCap, "q_existing_guardian", err.Error(), err)
	}
	guardianCap, err := strconv.ParseUint(guardianCapRaw, 10, 64)
	if err != nil || guardianCap == 0 {
		return QFinalEntryPrecheck{}, qFinalRefusal(riskbucket.RefusalExistingGuardianCap, "q_existing_guardian", "Guardian cap is not a positive uint64", err)
	}
	admission.Admission.QExistingGuardian = guardianCap
	decision := riskbucket.CalculateAdmission(admission.Admission)
	if decision.Refusal != nil {
		return QFinalEntryPrecheck{}, qFinalRefusal(decision.Refusal.Code, decision.Refusal.Field, decision.Refusal.Error(), decision.Refusal)
	}
	intent := risk.Intent{
		AccountRef: g.accountRef, Market: market, Symbol: symbol, Side: risk.SideBuy,
		Quantity: strconv.FormatUint(decision.QFinal, 10), LimitPrice: request.EntryPrice,
		StopPrice: request.StopPrice, TargetPrice: request.TargetPrice,
	}
	in := risk.Input{Now: g.clk.Now().UTC(), Intent: intent, Account: request.Account, Policy: g.policy, Costs: g.costs}
	if verdict := evaluateChain(in); !verdict.Allowed {
		return QFinalEntryPrecheck{}, chainRefusal(verdict)
	}
	if _, verdict := risk.EntryExposureValue(in); !verdict.Allowed {
		return QFinalEntryPrecheck{}, chainRefusal(verdict)
	}
	request.Market, request.Currency, request.Symbol = marketName, currency, symbol
	request.Admission = admission
	return QFinalEntryPrecheck{request: request, market: market, intent: intent, admission: admission, decision: decision, existingCap: guardianCap}, nil
}

func cloneRiskBucketAdmissionPlan(plan journal.RiskBucketAdmissionPlan) journal.RiskBucketAdmissionPlan {
	plan.Admission.Buckets = append([]riskbucket.BucketSnapshot(nil), plan.Admission.Buckets...)
	plan.Snapshots = append([]journal.RiskBucketSnapshotReference(nil), plan.Snapshots...)
	return plan
}

// IssuePrecheckedQFinalEntry is the mutation half. The final decision is minted
// only here, after q_final is sealed in the opaque precheck.
func (g *RiskGuardian) IssuePrecheckedQFinalEntry(ctx context.Context, precheck QFinalEntryPrecheck) (Issued, error) {
	if g == nil || precheck.decision.QFinal == 0 || precheck.request.Collect == nil || precheck.admission.TransactionID == "" {
		return Issued{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid, "precheck", "complete sealed precheck required", nil)
	}
	now := g.clk.Now().UTC()
	policyVersion, err := journal.QFinalPolicyVersion(g.policyVersion, precheck.admission.TransactionID)
	if err != nil {
		return Issued{}, err
	}
	decision := journal.DecisionRequest{
		ID: g.newID(), AccountRef: g.accountRef, Generation: 0,
		SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: g.accountRef, Market: string(precheck.market), Symbol: precheck.intent.Symbol,
			Side: string(risk.SideBuy), Quantity: precheck.intent.Quantity, EntryPrice: precheck.intent.LimitPrice,
			StopPrice: precheck.intent.StopPrice, TargetPrice: precheck.intent.TargetPrice, PolicyVersion: policyVersion,
		},
		LimitsJSON: g.limitsJSON, Nonce: g.newID(), IssuedAt: now, ExpiresAt: now.Add(g.ttl),
	}
	reservationID := "res-" + decision.ID
	admission := precheck.admission
	// Re-evaluate freshness at final issuance too: an opaque precheck may have
	// waited while price/FX/snapshot evidence expired. The journal recalculates
	// q_final from this time and refuses before inserting the decision if it no
	// longer exactly matches the sealed quantity.
	admission.Admission.Policy.EvaluatedAt = now
	admission.DecisionID = decision.ID
	admission.ExistingReservationID = reservationID
	admission.CreatedAt = now

	collect := func(ctx context.Context, attempt int) (journal.QFinalIssueRequest, error) {
		snapshot, err := precheck.request.Collect(ctx, attempt)
		if err != nil {
			return journal.QFinalIssueRequest{}, err
		}
		exposure, verdict := risk.EntryExposureValue(risk.Input{Now: now, Intent: precheck.intent, Account: precheck.request.Account, Policy: g.policy, Costs: g.costs})
		if !verdict.Allowed {
			return journal.QFinalIssueRequest{}, chainRefusal(verdict)
		}
		usage, err := exposureUsage(snapshot.OpenExposure, exposure.Currency)
		if err != nil {
			return journal.QFinalIssueRequest{}, err
		}
		return journal.QFinalIssueRequest{
			Issue: journal.IssueRequest{Decision: decision, Reserve: journal.ReserveRequest{
				SnapshotAsOf: snapshot.AsOf, ObservedVersion: snapshot.Version,
				SnapshotUsage: []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: usage, Currency: exposure.Currency}},
				Limits:        []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: g.policy.MaxOpenExposure.Amount, Currency: exposure.Currency}},
				Reservations:  []journal.ReservationRequest{{ID: reservationID, Kind: journal.ReservationKindOpenExposure, Amount: exposure.Amount, Currency: exposure.Currency}},
			}}, Admission: admission,
		}, nil
	}
	out, err := g.journal.RecordQFinalDecisionAndReserveWithRecollection(ctx, collect, g.recollect)
	if err != nil {
		return Issued{}, issuanceRefusal(err)
	}
	return Issued{
		Decision:  GuardianDecision{ID: out.Issue.Decision.ID, Generation: out.Issue.Decision.Generation},
		ExpiresAt: out.Issue.Decision.ExpiresAt, Reservations: out.Issue.Reservations,
		Version: out.Issue.Version, RiskBucketReceipt: out.Admission,
	}, nil
}

func (g *RiskGuardian) IssueQFinalEntry(ctx context.Context, request QFinalEntryIssuance) (Issued, error) {
	precheck, err := g.PrecheckQFinalEntry(request)
	if err != nil {
		return Issued{}, err
	}
	return g.IssuePrecheckedQFinalEntry(ctx, precheck)
}

func qFinalMarketCurrency(market, currency string) (costs.Market, string, riskbucket.Market, error) {
	switch market {
	case "KR":
		if currency != "KRW" {
			return "", "", "", qFinalRefusal(riskbucket.RefusalCurrencyUnresolved, "market_currency", "KR requires KRW", nil)
		}
		return costs.MarketKR, "KRW", riskbucket.MarketKR, nil
	case "US":
		if currency != "USD" {
			return "", "", "", qFinalRefusal(riskbucket.RefusalCurrencyUnresolved, "market_currency", "US requires USD", nil)
		}
		return costs.MarketUS, "USD", riskbucket.MarketUS, nil
	default:
		return "", "", "", qFinalRefusal(riskbucket.RefusalUnknownMarket, "market", "only KR and US are supported", nil)
	}
}

func qFinalRefusal(code riskbucket.RefusalCode, field, detail string, cause error) error {
	return &QFinalRefusal{Code: code, Field: field, Detail: detail, cause: cause}
}
