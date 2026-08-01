package strategydispatch

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

type AccountStateReader interface {
	ReadStrategyAccount(context.Context) (risk.AccountState, error)
}

// GuardianAdapter is the concrete dormant bridge to the existing Guardian and
// atomic journal issuance. Constructing it does not start a loop or enable LIVE.
type GuardianAdapter struct {
	Guardian *execgw.RiskGuardian
	Journal  *journal.Journal
	Account  AccountStateReader
	Collect  execgw.CollectExposure
}

func (a *GuardianAdapter) IssueAndPlan(ctx context.Context, request IssueRequest) (AtomicPlan, PlanReceipt, error) {
	if a == nil || a.Guardian == nil || a.Journal == nil || a.Account == nil || a.Collect == nil {
		return AtomicPlan{}, PlanReceipt{}, errors.New("strategy adapter: Guardian path not configured")
	}
	if request.Binding.AccountRef != a.Guardian.AccountRef() || request.Binding.GuardianVersion != a.Guardian.PolicyVersion() ||
		request.Binding.GuardianLimitsDigest != a.Guardian.LimitsDigest() || request.Order.SettingsDigest != request.Binding.SettingsDigest {
		return AtomicPlan{}, PlanReceipt{}, errors.New("strategy adapter: Guardian/manifest snapshot mismatch")
	}
	account, err := a.Account.ReadStrategyAccount(ctx)
	if err != nil {
		return AtomicPlan{}, PlanReceipt{}, err
	}
	issued, err := a.Guardian.IssueStrategyEntry(ctx, execgw.StrategyEntryIssuance{
		Decision: request.Decision, Account: account, Collect: a.Collect, AttemptID: request.AttemptID,
		ActivationManifestDigest: request.ManifestDigest, SettingsDigest: request.Order.SettingsDigest,
		ExpectedPolicyVersion: request.Binding.GuardianVersion, ExpectedLimitsDigest: request.Binding.GuardianLimitsDigest,
	})
	if err != nil {
		return AtomicPlan{}, PlanReceipt{}, err
	}
	receipt := issued.StrategyReceipt
	if receipt.AttemptID != request.AttemptID || receipt.AccountRef != request.Binding.AccountRef ||
		receipt.DecisionIdentity != request.Decision.Record().Identity || receipt.RiskIntentID != issued.Decision.ID ||
		receipt.ClientOrderID == "" || receipt.Quantity == "" || receipt.Revision != 1 || receipt.State != "PLANNED" {
		return AtomicPlan{}, PlanReceipt{}, errors.New("strategy adapter: atomic receipt binding mismatch")
	}
	planReceipt := PlanReceipt{
		AttemptID: receipt.AttemptID, AccountRef: receipt.AccountRef, DecisionIdentity: receipt.DecisionIdentity,
		RiskIntentID: receipt.RiskIntentID, ClientOrderID: receipt.ClientOrderID, Quantity: receipt.Quantity,
		Revision: uint64(receipt.Revision), State: receipt.State,
	}
	return AtomicPlan{
		AttemptID: receipt.AttemptID, Decision: request.Decision, RiskIntentID: issued.Decision.ID,
		GuardianDecisionID: issued.Decision.ID, GuardianGeneration: issued.Decision.Generation,
		ManifestDigest: request.ManifestDigest, ClientOrderID: receipt.ClientOrderID,
		Quantity: receipt.Quantity, Order: request.Order,
	}, planReceipt, nil
}

func (a *GuardianAdapter) RecordStrategyRefusal(ctx context.Context, receipt PlanReceipt, reason Reason) error {
	return a.Journal.RecordStrategyRefusal(ctx, journalReceipt(receipt), string(reason))
}

func (a *GuardianAdapter) RecordStrategyInDoubt(ctx context.Context, receipt PlanReceipt, reason Reason) error {
	return a.Journal.RecordStrategyInDoubt(ctx, journalReceipt(receipt), string(reason))
}

func (a *GuardianAdapter) RecordStrategyDispatched(ctx context.Context, receipt PlanReceipt, mutationAttemptID, brokerOrderID string) error {
	return a.Journal.RecordStrategyDispatched(ctx, journalReceipt(receipt), receipt.AccountRef, mutationAttemptID, brokerOrderID)
}

func journalReceipt(receipt PlanReceipt) journal.StrategyPlanReceipt {
	return journal.StrategyPlanReceipt{
		AttemptID: receipt.AttemptID, AccountRef: receipt.AccountRef, DecisionIdentity: receipt.DecisionIdentity,
		RiskIntentID: receipt.RiskIntentID, ClientOrderID: receipt.ClientOrderID, Quantity: receipt.Quantity,
		Revision: int(receipt.Revision), State: receipt.State,
	}
}

// GatewayAdapter is the only strategy mutation adapter. It delegates to
// execgw.Gateway, whose wrapped trading service uses the official Open API.
type GatewayAdapter struct{ Gateway *execgw.Gateway }

func (a GatewayAdapter) PlaceStrategyEntry(ctx context.Context, plan AtomicPlan) (execgw.Outcome, error) {
	if a.Gateway == nil || !plan.Decision.Valid() || plan.Order.OrderType != "LIMIT" || plan.Order.Currency != "KRW" {
		return execgw.Outcome{}, errors.New("strategy adapter: official gateway not configured")
	}
	record := plan.Decision.Record()
	quantity, err := exactFloat(plan.Quantity, true)
	if err != nil {
		return execgw.Outcome{}, fmt.Errorf("strategy adapter: invalid Guardian quantity: %w", err)
	}
	price, err := exactFloat(record.EntryPrice, false)
	if err != nil {
		return execgw.Outcome{}, fmt.Errorf("strategy adapter: invalid entry price: %w", err)
	}
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol: record.Symbol, Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: quantity, Price: price, CurrencyMode: "KRW",
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	outcome, err := a.Gateway.Place(ctx, execgw.PlaceRequest{
		Intent: intent, IntentID: plan.AttemptID,
		Decision: execgw.GuardianDecision{ID: plan.GuardianDecisionID, Generation: plan.GuardianGeneration},
	})
	return outcome, err
}

func exactFloat(raw string, whole bool) (float64, error) {
	normalized, ok := journal.NormalizeDecimal(raw)
	if !ok || normalized != strings.TrimSpace(raw) || (whole && strings.Contains(normalized, ".")) {
		return 0, errors.New("non-canonical decimal")
	}
	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil || value <= 0 || (whole && value > 1<<53) {
		return 0, errors.New("decimal outside exact float boundary")
	}
	roundTrip, ok := journal.NormalizeDecimal(strconv.FormatFloat(value, 'f', -1, 64))
	if !ok || roundTrip != normalized {
		return 0, errors.New("decimal loses precision in gateway type")
	}
	return value, nil
}

var _ StrategyIssuer = (*GuardianAdapter)(nil)
var _ OfficialGateway = GatewayAdapter{}
var _ interface {
	Place(context.Context, execgw.PlaceRequest) (execgw.Outcome, error)
} = (*execgw.Gateway)(nil)
