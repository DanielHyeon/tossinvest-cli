package strategyflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// AcceptedProjectionSchemaVersion is the only canonical persisted schema this
// build emits or accepts for a pure strategyflow result.
const AcceptedProjectionSchemaVersion = "strategyflow-accepted:v1"

// AcceptedProjection is read-only canonical evidence. It neither grants
// execution authority nor exposes a way to populate sealed execution terms.
type AcceptedProjection struct {
	payload       string
	payloadDigest string
	lineage       Lineage
	terms         ExecutionTerms
}

func (p AcceptedProjection) Payload() string                { return p.payload }
func (p AcceptedProjection) PayloadDigest() string          { return p.payloadDigest }
func (p AcceptedProjection) Lineage() Lineage               { return p.lineage }
func (p AcceptedProjection) ExecutionTerms() ExecutionTerms { return p.terms }

type acceptedProjectionPayload struct {
	SchemaVersion           string                 `json:"schema_version"`
	Quantity                uint64                 `json:"quantity"`
	Lineage                 acceptedLineagePayload `json:"lineage"`
	ExecutionTerms          acceptedTermsPayload   `json:"execution_terms"`
	CommonSafetyIndependent bool                   `json:"common_safety_independent"`
	GuardianCalls           uint64                 `json:"guardian_calls"`
	BrokerCalls             uint64                 `json:"broker_calls"`
	Mutations               uint64                 `json:"mutations"`
}

type acceptedLineagePayload struct {
	AccountRef              string `json:"account_ref"`
	Market                  string `json:"market"`
	Symbol                  string `json:"symbol"`
	PositionGeneration      uint64 `json:"position_generation"`
	CandidateState          string `json:"candidate_state"`
	CandidateLifeID         string `json:"candidate_life_id"`
	CandidateFirstSeenNS    int64  `json:"candidate_first_seen_ns"`
	CandidateLastSeenNS     int64  `json:"candidate_last_seen_ns"`
	CandidateValidUntilNS   int64  `json:"candidate_valid_until_ns"`
	CandidateApprovedAtNS   int64  `json:"candidate_approved_at_ns"`
	ThresholdVersion        string `json:"threshold_version"`
	ThresholdSetDigest      string `json:"threshold_set_digest"`
	CandidateEvidenceDigest string `json:"candidate_evidence_digest"`
	RouterEvidenceDigest    string `json:"router_evidence_digest"`
	LaneEvidenceDigest      string `json:"lane_evidence_digest"`
	RouterID                string `json:"router_id"`
	RouterRelease           string `json:"router_release"`
	Horizon                 string `json:"horizon"`
	LaneID                  string `json:"lane_id"`
	LaneVersion             string `json:"lane_version"`
	LaneRelease             string `json:"lane_release"`
	ConfigDigest            string `json:"config_digest"`
	CampaignID              string `json:"campaign_id"`
	LegOrdinal              int    `json:"leg_ordinal"`
	PlannedCeiling          uint64 `json:"planned_ceiling"`
	RiskBudgetDigest        string `json:"risk_budget_digest"`
	ExecutionPolicyDigest   string `json:"execution_policy_digest"`
	Complete                bool   `json:"complete"`
	Identity                string `json:"identity"`
}

type acceptedTermsPayload struct {
	AccountRef      string                `json:"account_ref"`
	Market          string                `json:"market"`
	Symbol          string                `json:"symbol"`
	CampaignID      string                `json:"campaign_id"`
	LegOrdinal      int                   `json:"leg_ordinal"`
	Quantity        uint64                `json:"quantity"`
	Entry           acceptedPricePayload  `json:"entry"`
	EffectiveStop   acceptedPricePayload  `json:"effective_stop"`
	Target          acceptedPricePayload  `json:"target"`
	Policy          acceptedPolicyPayload `json:"policy"`
	LineageIdentity string                `json:"lineage_identity"`
	Identity        string                `json:"identity"`
}

type acceptedPricePayload struct {
	PriceMinor  string `json:"price_minor"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Digest      string `json:"digest"`
	AsOf        string `json:"as_of"`
	Currency    string `json:"currency"`
	MinorScale  int    `json:"minor_scale"`
	UnitVersion string `json:"unit_version"`
}

type acceptedPolicyPayload struct {
	StagedTargetMinor string `json:"staged_target_minor"`
	FairValueMinor    string `json:"fair_value_minor"`
	EntryCostsMinor   string `json:"entry_costs_minor"`
	ExitCostsMinor    string `json:"exit_costs_minor"`
	MinimumRRPPM      uint64 `json:"minimum_rr_ppm"`
	DecisionDigest    string `json:"decision_digest"`
	CalendarDigest    string `json:"calendar_digest"`
	CapSnapshotID     string `json:"cap_snapshot_id"`
	Identity          string `json:"identity"`
}

// ProjectAccepted creates deterministic evidence from a sealed, pure accepted
// result. It performs no I/O and cannot call Guardian, journal, Gateway or broker.
func ProjectAccepted(result Result) (AcceptedProjection, error) {
	if err := validateAcceptedResult(result); err != nil {
		return AcceptedProjection{}, err
	}
	record := projectionPayloadFromResult(result)
	payload, err := json.Marshal(record)
	if err != nil {
		return AcceptedProjection{}, fmt.Errorf("strategyflow: canonical accepted projection encoding: %w", err)
	}
	return acceptedProjection(string(payload), result.Lineage, result.ExecutionTerms), nil
}

// VerifyAcceptedProjection strictly verifies canonical persisted evidence. It
// returns another opaque read-only value and does not grant execution authority.
func VerifyAcceptedProjection(payload string) (AcceptedProjection, error) {
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	var record acceptedProjectionPayload
	if err := decoder.Decode(&record); err != nil {
		return AcceptedProjection{}, fmt.Errorf("strategyflow: canonical accepted projection decoding: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return AcceptedProjection{}, errors.New("strategyflow: canonical accepted projection has trailing data")
	}
	canonical, err := json.Marshal(record)
	if err != nil || string(canonical) != payload {
		return AcceptedProjection{}, errors.New("strategyflow: canonical accepted projection is not canonical")
	}
	if record.SchemaVersion != AcceptedProjectionSchemaVersion {
		return AcceptedProjection{}, errors.New("strategyflow: canonical accepted projection schema unsupported")
	}
	result := resultFromProjectionPayload(record)
	if err := validateAcceptedResult(result); err != nil {
		return AcceptedProjection{}, err
	}
	return acceptedProjection(payload, result.Lineage, result.ExecutionTerms), nil
}

func acceptedProjection(payload string, lineage Lineage, terms ExecutionTerms) AcceptedProjection {
	sum := sha256.Sum256([]byte(payload))
	return AcceptedProjection{payload: payload, payloadDigest: "sha256:" + hex.EncodeToString(sum[:]), lineage: lineage, terms: terms}
}

func validateAcceptedResult(result Result) error {
	if result.Code != RefusalNone || result.NativeCode != "" || result.Quantity == 0 || !result.CommonSafetyIndependent ||
		result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 {
		return errors.New("strategyflow: canonical accepted projection requires a pure accepted result")
	}
	lineage, terms := result.Lineage, result.ExecutionTerms
	if !lineage.Complete || !lineage.Valid() || !terms.Valid() {
		return errors.New("strategyflow: canonical accepted projection requires valid sealed lineage and execution terms")
	}
	if lineage.RouterID != strategyrouter.RouterID || lineage.RouterRelease != strategyrouter.RouterRelease || !registeredProjectionDescriptor(lineage) {
		return errors.New("strategyflow: canonical accepted projection has unsupported production binding")
	}
	if terms.AccountRef() != lineage.AccountRef || terms.Market() != lineage.Market || terms.Symbol() != lineage.Symbol ||
		terms.CampaignID() != lineage.CampaignID || terms.LegOrdinal() != lineage.LegOrdinal || terms.Quantity() != result.Quantity ||
		terms.LineageIdentity() != lineage.Identity || terms.Policy().Identity() != lineage.ExecutionPolicyDigest {
		return errors.New("strategyflow: canonical accepted projection lineage/terms mismatch")
	}
	wantCurrency, wantScale := map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[lineage.Market], map[strategyrouter.Market]int{strategyrouter.MarketKR: 0, strategyrouter.MarketUS: 2}[lineage.Market]
	for _, price := range []PriceProvenance{terms.Entry(), terms.EffectiveStop(), terms.Target()} {
		if price.Currency() != wantCurrency || price.MinorScale() != wantScale {
			return errors.New("strategyflow: canonical accepted projection price unit mismatch")
		}
	}
	return nil
}

func registeredProjectionDescriptor(lineage Lineage) bool {
	for _, descriptor := range pairedDescriptors {
		if lineage.Market == descriptor.Market && lineage.Horizon == descriptor.Horizon && lineage.LaneID == descriptor.LaneID &&
			lineage.LaneVersion == descriptor.LaneVersion && lineage.LaneRelease == descriptor.Release {
			return true
		}
	}
	return false
}

func projectionPayloadFromResult(result Result) acceptedProjectionPayload {
	lineage, terms := result.Lineage, result.ExecutionTerms
	return acceptedProjectionPayload{SchemaVersion: AcceptedProjectionSchemaVersion, Quantity: result.Quantity,
		Lineage: lineageToPayload(lineage), ExecutionTerms: termsToPayload(terms), CommonSafetyIndependent: result.CommonSafetyIndependent,
		GuardianCalls: result.GuardianCalls, BrokerCalls: result.BrokerCalls, Mutations: result.Mutations}
}

func lineageToPayload(v Lineage) acceptedLineagePayload {
	return acceptedLineagePayload{AccountRef: v.AccountRef, Market: string(v.Market), Symbol: v.Symbol, PositionGeneration: v.PositionGeneration,
		CandidateState: v.CandidateState, CandidateLifeID: v.CandidateLifeID, CandidateFirstSeenNS: v.CandidateFirstSeenNS,
		CandidateLastSeenNS: v.CandidateLastSeenNS, CandidateValidUntilNS: v.CandidateValidUntilNS, CandidateApprovedAtNS: v.CandidateApprovedAtNS,
		ThresholdVersion: v.ThresholdVersion, ThresholdSetDigest: v.ThresholdSetDigest, CandidateEvidenceDigest: v.CandidateEvidenceDigest,
		RouterEvidenceDigest: v.RouterEvidenceDigest, LaneEvidenceDigest: v.LaneEvidenceDigest, RouterID: v.RouterID, RouterRelease: v.RouterRelease,
		Horizon: string(v.Horizon), LaneID: v.LaneID, LaneVersion: v.LaneVersion, LaneRelease: v.LaneRelease, ConfigDigest: v.ConfigDigest,
		CampaignID: v.CampaignID, LegOrdinal: v.LegOrdinal, PlannedCeiling: v.PlannedCeiling, RiskBudgetDigest: v.RiskBudgetDigest,
		ExecutionPolicyDigest: v.ExecutionPolicyDigest, Complete: v.Complete, Identity: v.Identity}
}

func termsToPayload(v ExecutionTerms) acceptedTermsPayload {
	return acceptedTermsPayload{AccountRef: v.AccountRef(), Market: string(v.Market()), Symbol: v.Symbol(), CampaignID: v.CampaignID(),
		LegOrdinal: v.LegOrdinal(), Quantity: v.Quantity(), Entry: priceToPayload(v.Entry()), EffectiveStop: priceToPayload(v.EffectiveStop()),
		Target: priceToPayload(v.Target()), Policy: policyToPayload(v.Policy()), LineageIdentity: v.LineageIdentity(), Identity: v.Identity()}
}

func priceToPayload(v PriceProvenance) acceptedPricePayload {
	return acceptedPricePayload{PriceMinor: v.PriceMinor(), Source: v.Source(), Version: v.Version(), Digest: v.Digest(), AsOf: v.AsOf(),
		Currency: v.Currency(), MinorScale: v.MinorScale(), UnitVersion: v.UnitVersion()}
}

func policyToPayload(v ExecutionPolicy) acceptedPolicyPayload {
	return acceptedPolicyPayload{StagedTargetMinor: v.StagedTargetMinor(), FairValueMinor: v.FairValueMinor(), EntryCostsMinor: v.EntryCostsMinor(),
		ExitCostsMinor: v.ExitCostsMinor(), MinimumRRPPM: v.MinimumRRPPM(), DecisionDigest: v.DecisionDigest(), CalendarDigest: v.CalendarDigest(),
		CapSnapshotID: v.CapSnapshotID(), Identity: v.Identity()}
}

func resultFromProjectionPayload(v acceptedProjectionPayload) Result {
	lineage := Lineage{AccountRef: v.Lineage.AccountRef, Market: strategyrouter.Market(v.Lineage.Market), Symbol: v.Lineage.Symbol,
		PositionGeneration: v.Lineage.PositionGeneration, CandidateState: v.Lineage.CandidateState, CandidateLifeID: v.Lineage.CandidateLifeID,
		CandidateFirstSeenNS: v.Lineage.CandidateFirstSeenNS, CandidateLastSeenNS: v.Lineage.CandidateLastSeenNS,
		CandidateValidUntilNS: v.Lineage.CandidateValidUntilNS, CandidateApprovedAtNS: v.Lineage.CandidateApprovedAtNS,
		ThresholdVersion: v.Lineage.ThresholdVersion, ThresholdSetDigest: v.Lineage.ThresholdSetDigest,
		CandidateEvidenceDigest: v.Lineage.CandidateEvidenceDigest, RouterEvidenceDigest: v.Lineage.RouterEvidenceDigest,
		LaneEvidenceDigest: v.Lineage.LaneEvidenceDigest, RouterID: v.Lineage.RouterID, RouterRelease: v.Lineage.RouterRelease,
		Horizon: strategyrouter.Horizon(v.Lineage.Horizon), LaneID: v.Lineage.LaneID, LaneVersion: v.Lineage.LaneVersion,
		LaneRelease: v.Lineage.LaneRelease, ConfigDigest: v.Lineage.ConfigDigest, CampaignID: v.Lineage.CampaignID,
		LegOrdinal: v.Lineage.LegOrdinal, PlannedCeiling: v.Lineage.PlannedCeiling, RiskBudgetDigest: v.Lineage.RiskBudgetDigest,
		ExecutionPolicyDigest: v.Lineage.ExecutionPolicyDigest, Complete: v.Lineage.Complete, Identity: v.Lineage.Identity}
	policy := ExecutionPolicy{stagedTargetMinor: v.ExecutionTerms.Policy.StagedTargetMinor, fairValueMinor: v.ExecutionTerms.Policy.FairValueMinor,
		entryCostsMinor: v.ExecutionTerms.Policy.EntryCostsMinor, exitCostsMinor: v.ExecutionTerms.Policy.ExitCostsMinor,
		minimumRRPPM: v.ExecutionTerms.Policy.MinimumRRPPM, decisionDigest: v.ExecutionTerms.Policy.DecisionDigest,
		calendarDigest: v.ExecutionTerms.Policy.CalendarDigest, capSnapshotID: v.ExecutionTerms.Policy.CapSnapshotID, identity: v.ExecutionTerms.Policy.Identity}
	terms := ExecutionTerms{accountRef: v.ExecutionTerms.AccountRef, market: strategyrouter.Market(v.ExecutionTerms.Market), symbol: v.ExecutionTerms.Symbol,
		campaignID: v.ExecutionTerms.CampaignID, legOrdinal: v.ExecutionTerms.LegOrdinal, quantity: v.ExecutionTerms.Quantity,
		entry: priceFromPayload(v.ExecutionTerms.Entry), stop: priceFromPayload(v.ExecutionTerms.EffectiveStop), target: priceFromPayload(v.ExecutionTerms.Target),
		policy: policy, lineageIdentity: v.ExecutionTerms.LineageIdentity, identity: v.ExecutionTerms.Identity}
	return Result{Quantity: v.Quantity, Lineage: lineage, ExecutionTerms: terms, CommonSafetyIndependent: v.CommonSafetyIndependent,
		GuardianCalls: v.GuardianCalls, BrokerCalls: v.BrokerCalls, Mutations: v.Mutations}
}

func priceFromPayload(v acceptedPricePayload) PriceProvenance {
	return PriceProvenance{priceMinor: v.PriceMinor, source: v.Source, version: v.Version, digest: v.Digest, asOf: v.AsOf,
		currency: v.Currency, minorScale: v.MinorScale, unitVersion: v.UnitVersion}
}
