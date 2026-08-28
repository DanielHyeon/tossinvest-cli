// Package strategyproposal reconstructs sealed, cap-free q_candidate proposals
// from externally signed KR/US manifests and immutable local evidence.
package strategyproposal

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

const (
	productionSchema       = "strategy-proposal-input:v1"
	productionDomain       = "TossOS/strategy-proposal-input/ed25519/v1"
	productionAlgorithm    = "Ed25519"
	productionMaximumBytes = 1 << 20
)

var ErrProductionProposalUnavailable = errors.New("strategyproposal: production proposal authority unavailable")

// ErrBreakoutEvidenceUnavailable 는 돌파-되돌림 레인이 아직 생산 입력을 만들 수 없다는 뜻이다.
// 순수 코어는 봉마다 RVOL·윗꼬리 ppm·거래량 확장 여부를, 스냅샷마다 ATR 을 요구하는데(결정 49),
// L1 이 저장하는 닫힌 1분봉 증거에는 그 네 값이 하나도 없고 만들어 주는 생산자도 없다.
// 그래서 여기서는 값을 지어내지 않고 닫아서 거절한다.
var ErrBreakoutEvidenceUnavailable = errors.New("strategyproposal: breakout derived-metric evidence (atr, rvol ppm, upper-wick ppm, volume-expanded) is neither stored nor produced, so no breakout lane input can be built")

type ProductionConfig struct {
	ConfigDir, EvidencePath, JournalPath       string
	AccountRef                                 string
	Market                                     strategyrouter.Market
	ManifestDigest, TrustedKeyID               string
	TrustedKey                                 ed25519.PublicKey
	ObservedAt                                 time.Time
	RouteManifestDigest, ActivationDigest      string
	CalendarGeneration, CalendarDigest         string
	SchedulerConfigVersion, EvidenceDBIdentity string
}

type ProductionTarget struct {
	Approved strategy.ApprovedSnapshot
	Router   strategyrouter.RouteRequest
}

type ProductionAuthority struct {
	proposal       strategyflow.Result
	snapshotID     string
	snapshotDigest string
	weekly         *journal.WeeklyFirstLegReservationBinding
}

func (authority ProductionAuthority) Proposal() strategyflow.Result { return authority.proposal }
func (authority ProductionAuthority) SnapshotID() string            { return authority.snapshotID }
func (authority ProductionAuthority) SnapshotDigest() string        { return authority.snapshotDigest }
func (authority ProductionAuthority) WeeklyBinding() *journal.WeeklyFirstLegReservationBinding {
	if authority.weekly == nil {
		return nil
	}
	copy := *authority.weekly
	return &copy
}

type ProductionBatchAuthority struct {
	values         map[string]ProductionAuthority
	manifestDigest string
}

func (authority ProductionBatchAuthority) Len() int               { return len(authority.values) }
func (authority ProductionBatchAuthority) ManifestDigest() string { return authority.manifestDigest }

// batchKey 는 담는 단위다. 종목 하나가 여러 가족을 동시에 제안할 수 있으므로
// 종목만으로 담으면 나중 것이 앞의 것을 조용히 덮어쓴다.
func batchKey(symbol, laneID string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "\x00" + laneID
}

// For 는 그 종목의 제안이 정확히 하나일 때만 돌려준다.
// 둘 이상이면 여기서 고르지 않고 닫아서 거절한다 — 고르는 일은 조정자의 몫이고,
// 임의로 하나를 집으면 그게 곧 사전 선택이기 때문이다.
func (authority ProductionBatchAuthority) For(symbol string) (ProductionAuthority, bool) {
	values := authority.LanesFor(symbol)
	if len(values) != 1 {
		return ProductionAuthority{}, false
	}
	return values[0], true
}

// LanesFor 는 그 종목의 모든 가족 제안을 레인 이름 순서로 돌려준다.
func (authority ProductionBatchAuthority) LanesFor(symbol string) []ProductionAuthority {
	prefix := strings.ToUpper(strings.TrimSpace(symbol)) + "\x00"
	keys := make([]string, 0, len(authority.values))
	for key := range authority.values {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	values := make([]ProductionAuthority, 0, len(keys))
	for _, key := range keys {
		values = append(values, authority.values[key])
	}
	return values
}

type productionStop struct {
	PriceMinor string `json:"price_minor"`
	Source     string `json:"source"`
	Policy     string `json:"policy"`
	Version    string `json:"version"`
	Digest     string `json:"digest"`
	ObservedAt string `json:"observed_at"`
	FreshUntil string `json:"fresh_until"`
}

type productionStructureEvent struct {
	Kind       string `json:"kind"`
	RecordID   string `json:"record_id"`
	Digest     string `json:"digest"`
	At         string `json:"at"`
	FreshUntil string `json:"fresh_until"`
}

type productionStructure struct {
	Sweep   productionStructureEvent `json:"sweep"`
	Break   productionStructureEvent `json:"break"`
	Reclaim productionStructureEvent `json:"reclaim"`
}

type productionScope struct {
	Symbol                   string                 `json:"symbol"`
	PositionGeneration       uint64                 `json:"position_generation"`
	CandidateID              string                 `json:"candidate_id"`
	CampaignID               string                 `json:"campaign_id"`
	Horizon                  strategyrouter.Horizon `json:"horizon"`
	LaneID                   string                 `json:"lane_id"`
	LaneVersion              string                 `json:"lane_version"`
	SnapshotID               string                 `json:"snapshot_id"`
	SnapshotDigest           string                 `json:"snapshot_digest"`
	RiskBudgetMinor          string                 `json:"risk_budget_minor"`
	PerShareRiskMinor        string                 `json:"per_share_risk_minor"`
	PlannedQuantity          uint64                 `json:"planned_quantity"`
	PolicyDigest             string                 `json:"policy_digest"`
	AccountCurrency          string                 `json:"account_currency"`
	QuoteCurrency            string                 `json:"quote_currency"`
	LegOrdinal               int                    `json:"leg_ordinal"`
	FilledQuantity           uint64                 `json:"filled_quantity"`
	SavedEffectiveStopMinor  string                 `json:"saved_effective_stop_minor"`
	Stop                     productionStop         `json:"stop"`
	EntryPriceMinor          string                 `json:"entry_price_minor"`
	TargetPriceMinor         string                 `json:"target_price_minor"`
	FreshUntil               string                 `json:"fresh_until"`
	ConfigVersion            string                 `json:"config_version"`
	ConfigDigest             string                 `json:"config_digest"`
	ThresholdSet             string                 `json:"threshold_set"`
	MinimumFlowPressurePPM   int64                  `json:"minimum_flow_pressure_ppm"`
	MinimumParticipationPPM  int64                  `json:"minimum_participation_ppm"`
	MinimumPriceChangePPM    int64                  `json:"minimum_price_change_ppm"`
	MinimumAbsorptionPPM     uint64                 `json:"minimum_absorption_ppm"`
	MinimumDrawdownPPM       uint64                 `json:"minimum_drawdown_ppm"`
	MinimumRelativeVolumePPM uint64                 `json:"minimum_relative_volume_ppm"`
	StructuralWindowNS       int64                  `json:"structural_window_ns"`
	Structure                productionStructure    `json:"structure"`
	ModelVersion             string                 `json:"model_version"`
	ThresholdDigest          string                 `json:"threshold_digest"`
	StagedTargetMinor        string                 `json:"staged_target_minor"`
	EntryCostsMinor          string                 `json:"entry_costs_minor"`
	EstimatedExitCostsMinor  string                 `json:"estimated_exit_costs_minor"`
	MinimumRRPPM             uint64                 `json:"minimum_rr_ppm"`
	WeeklyReservationID      string                 `json:"weekly_reservation_id"`
	StableWeek               string                 `json:"stable_week"`
}

type productionBody struct {
	SchemaVersion          string                `json:"schema_version"`
	Domain                 string                `json:"domain"`
	SignatureAlgorithm     string                `json:"signature_algorithm"`
	KeyID                  string                `json:"key_id"`
	Generation             uint64                `json:"generation"`
	AccountRef             string                `json:"account_ref"`
	Market                 strategyrouter.Market `json:"market"`
	RouteManifestDigest    string                `json:"route_manifest_digest"`
	ActivationDigest       string                `json:"activation_digest"`
	CalendarGeneration     string                `json:"calendar_generation"`
	CalendarDigest         string                `json:"calendar_digest"`
	SchedulerConfigVersion string                `json:"scheduler_config_version"`
	EvidenceDBIdentity     string                `json:"evidence_db_identity"`
	Actor                  string                `json:"actor"`
	ObservedAt             string                `json:"observed_at"`
	FreshUntil             string                `json:"fresh_until"`
	Revoked                bool                  `json:"revoked"`
	Scopes                 []productionScope     `json:"scopes"`
}

type productionManifest struct {
	productionBody
	Signature string `json:"signature"`
}

func ProductionFileName(market strategyrouter.Market) string {
	if market == strategyrouter.MarketKR {
		return "strategy-proposal-input-KR.json"
	}
	if market == strategyrouter.MarketUS {
		return "strategy-proposal-input-US.json"
	}
	return ""
}

func LoadProductionAuthorityBatch(ctx context.Context, config ProductionConfig, targets []ProductionTarget, fx officialfx.Evidence) (ProductionBatchAuthority, error) {
	config = canonicalConfig(config)
	owner, ownerOK := productionOwnerUID()
	name := ProductionFileName(config.Market)
	if ctx == nil || !ownerOK || name == "" || !filepath.IsAbs(config.ConfigDir) || !filepath.IsAbs(config.EvidencePath) || !filepath.IsAbs(config.JournalPath) ||
		config.AccountRef == "" || config.ObservedAt.IsZero() || !digestValid(config.ManifestDigest) || !identity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize ||
		!identity(config.RouteManifestDigest) || !identity(config.ActivationDigest) || !identity(config.CalendarGeneration) || !identity(config.CalendarDigest) ||
		!identity(config.SchedulerConfigVersion) || !identity(config.EvidenceDBIdentity) || len(targets) == 0 || len(targets) > 10_000 {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	if err := ctx.Err(); err != nil {
		return ProductionBatchAuthority{}, err
	}
	data, err := readProductionFile(filepath.Join(config.ConfigDir, name), owner, 0o400, productionMaximumBytes)
	if err != nil || digest(data) != config.ManifestDigest {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	manifest, err := decodeManifest(data)
	if err != nil || !verifyManifest(manifest, config) {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	evidenceStore, err := strategyevidence.OpenReadOnly(ctx, strategyevidence.Options{Path: config.EvidencePath, Clock: marketclock.NewFake(config.ObservedAt)})
	if err != nil {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	defer evidenceStore.Close()
	port := strategyevidence.NewDormantSnapshotReadPort(evidenceStore)
	var journalRO *journal.ReadOnly
	for _, scope := range manifest.Scopes {
		if isWeeklyLane(scope.LaneID) {
			journalRO, err = journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: config.JournalPath})
			break
		}
	}
	if err != nil {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	if journalRO != nil {
		defer journalRO.Close()
	}
	targetBySymbol, ok := canonicalTargets(config.Market, targets)
	if !ok {
		return ProductionBatchAuthority{}, ErrProductionProposalUnavailable
	}
	values := make(map[string]ProductionAuthority, len(targetBySymbol))
	for _, scope := range manifest.Scopes {
		target, wanted := targetBySymbol[scope.Symbol]
		if !wanted {
			continue
		}
		// 평가 전에 승자를 고르지 않는다. 자격 있는 가족을 전부 받아서
		// 이 스코프의 레인이 그 안에 들어 있는지만 확인한다(태스크 4.3.1).
		routed := strategyrouter.RouteSet(target.Router)
		if routed.Code != strategyrouter.RefusalNone || !routed.Valid() || target.Approved.CandidateLifeID() != scope.CandidateID {
			continue
		}
		if !routeSetAdmitsScope(routed, config, scope) {
			continue
		}
		clockMarket := marketclock.MarketKR
		if config.Market == strategyrouter.MarketUS {
			clockMarket = marketclock.MarketUS
		}
		snapshot, err := port.Replay(ctx, clockMarket, strategyevidence.SnapshotReference{ID: scope.SnapshotID, Digest: scope.SnapshotDigest})
		if err != nil {
			continue
		}
		lane, weekly, err := buildLaneInput(ctx, config, scope, snapshot, fx, journalRO)
		if err != nil {
			continue
		}
		proposal := strategyflow.Propose(strategyflow.Request{Approved: target.Approved, Router: target.Router, Lane: lane})
		if !proposal.ValidProposal() {
			continue
		}
		values[batchKey(scope.Symbol, scope.LaneID)] = ProductionAuthority{proposal: proposal, snapshotID: snapshot.ID, snapshotDigest: snapshot.Digest, weekly: weekly}
	}
	return ProductionBatchAuthority{values: values, manifestDigest: config.ManifestDigest}, nil
}

// routeSetAdmitsScope 는 자격 집합에 이 스코프와 정확히 같은 신원이 들어 있는지만 확인한다.
// 하나도 없으면 그 스코프는 제안을 만들 수 없다. 느슨하게 맞추지 않는다.
func routeSetAdmitsScope(routed strategyrouter.RouteSetResult, config ProductionConfig, scope productionScope) bool {
	for _, decision := range routed.Decisions {
		if decision.Key.AccountRef != config.AccountRef || decision.Key.Market != config.Market ||
			decision.Key.Symbol != scope.Symbol || decision.Key.PositionGeneration != scope.PositionGeneration ||
			decision.LaneID != scope.LaneID || decision.LaneVersion != scope.LaneVersion || decision.Horizon != scope.Horizon {
			continue
		}
		// 자격은 맞지만 이미 다른 캠페인이 잡고 있으면 거절한다.
		if decision.ExistingOwner && decision.CampaignID != scope.CampaignID {
			return false
		}
		return true
	}
	return false
}

func buildLaneInput(ctx context.Context, config ProductionConfig, scope productionScope, snapshot strategyevidence.Snapshot, fx officialfx.Evidence, journalRO *journal.ReadOnly) (strategyflow.LaneInput, *journal.WeeklyFirstLegReservationBinding, error) {
	if scope.LaneID == breakoutlane.KRLaneID || scope.LaneID == breakoutlane.USLaneID {
		return strategyflow.LaneInput{}, nil, ErrBreakoutEvidenceUnavailable
	}
	stopObserved, ok1 := parseTime(scope.Stop.ObservedAt)
	stopFresh, ok2 := parseTime(scope.Stop.FreshUntil)
	fresh, ok3 := parseTime(scope.FreshUntil)
	if !ok1 || !ok2 || scope.CampaignID == "" || scope.SnapshotDigest != snapshot.Digest || scope.SnapshotID != snapshot.ID {
		return strategyflow.LaneInput{}, nil, ErrProductionProposalUnavailable
	}
	if scope.LaneID == continuationlane.KRContinuationLaneID || scope.LaneID == continuationlane.USContinuationLaneID {
		if !ok3 {
			return strategyflow.LaneInput{}, nil, ErrProductionProposalUnavailable
		}
		input := continuationlane.ProductionProposalInput{Snapshot: snapshot, FX: fx, AccountRef: config.AccountRef, Symbol: scope.Symbol,
			CandidateID: scope.CandidateID, CampaignID: scope.CampaignID, PositionGeneration: int64(scope.PositionGeneration), RiskBudgetMinor: scope.RiskBudgetMinor,
			PerShareRiskMinor: scope.PerShareRiskMinor, PlannedQuantity: scope.PlannedQuantity, PolicyDigest: scope.PolicyDigest, ConfigDigest: scope.ConfigDigest,
			AccountCurrency: scope.AccountCurrency, QuoteCurrency: scope.QuoteCurrency, ThresholdSetID: scope.ThresholdSet,
			MinimumFlowPressurePPM: scope.MinimumFlowPressurePPM, MinimumParticipationPPM: scope.MinimumParticipationPPM,
			MinimumPriceChangePPM: scope.MinimumPriceChangePPM, Leg: continuationlane.LegProgress{Ordinal: scope.LegOrdinal, FilledQuantity: scope.FilledQuantity},
			SavedEffectiveStopMinor: scope.SavedEffectiveStopMinor, Stop: continuationlane.ProductionStopInput{PriceMinor: scope.Stop.PriceMinor,
				Source: scope.Stop.Source, Policy: scope.Stop.Policy, Version: scope.Stop.Version, Digest: scope.Stop.Digest, ObservedAt: stopObserved, FreshUntil: stopFresh},
			EntryPriceMinor: scope.EntryPriceMinor, TargetPriceMinor: scope.TargetPriceMinor, FreshUntil: fresh}
		if config.Market == strategyrouter.MarketKR {
			input.Market = continuationlane.MarketKR
			a, err := continuationlane.BuildProductionKRProposalAuthority(input)
			if err != nil {
				return strategyflow.LaneInput{}, nil, err
			}
			return strategyflow.ContinuationKR(a.Request()), nil, nil
		}
		input.Market = continuationlane.MarketUS
		a, err := continuationlane.BuildProductionUSProposalAuthority(input)
		if err != nil {
			return strategyflow.LaneInput{}, nil, err
		}
		return strategyflow.ContinuationUS(a.Request()), nil, nil
	}
	if scope.LaneID == reversallane.KRReversalLaneID || scope.LaneID == reversallane.USReversalLaneID {
		if !ok3 {
			return strategyflow.LaneInput{}, nil, ErrProductionProposalUnavailable
		}
		market := reversallane.MarketKR
		if config.Market == strategyrouter.MarketUS {
			market = reversallane.MarketUS
		}
		evidenceSchema := "kr-absorption-v1"
		if market == reversallane.MarketUS {
			evidenceSchema = "us-dislocation-v1"
		}
		structure, err := reversalStructure(scope.Structure, market, config.AccountRef, scope.Symbol, scope.PositionGeneration, evidenceSchema)
		if err != nil {
			return strategyflow.LaneInput{}, nil, err
		}
		input := reversallane.ProductionProposalInput{Snapshot: snapshot, FX: fx, Market: market, AccountRef: config.AccountRef, Symbol: scope.Symbol,
			CandidateID: scope.CandidateID, CampaignID: scope.CampaignID, PositionGeneration: scope.PositionGeneration, RiskBudgetMinor: scope.RiskBudgetMinor,
			PerShareRiskMinor: scope.PerShareRiskMinor, PlannedQuantity: scope.PlannedQuantity, PolicyDigest: scope.PolicyDigest, ConfigDigest: scope.ConfigDigest,
			AccountCurrency: scope.AccountCurrency, QuoteCurrency: scope.QuoteCurrency, ConfigVersion: scope.ConfigVersion, ThresholdSet: scope.ThresholdSet,
			MinimumAbsorptionPPM: scope.MinimumAbsorptionPPM, MinimumDrawdownPPM: scope.MinimumDrawdownPPM, MinimumRelativeVolumePPM: scope.MinimumRelativeVolumePPM,
			StructuralWindow: time.Duration(scope.StructuralWindowNS), Leg: reversallane.LegProgress{Ordinal: scope.LegOrdinal, FilledQuantity: scope.FilledQuantity},
			SavedEffectiveStopMinor: scope.SavedEffectiveStopMinor, Stop: reversallane.ProductionStopInput{PriceMinor: scope.Stop.PriceMinor, Source: scope.Stop.Source,
				Policy: scope.Stop.Policy, Version: scope.Stop.Version, Digest: scope.Stop.Digest, ObservedAt: stopObserved, FreshUntil: stopFresh}, Structure: structure,
			EntryPriceMinor: scope.EntryPriceMinor, TargetPriceMinor: scope.TargetPriceMinor, FreshUntil: fresh}
		if market == reversallane.MarketKR {
			a, err := reversallane.BuildProductionKRProposalAuthority(input)
			if err != nil {
				return strategyflow.LaneInput{}, nil, err
			}
			return strategyflow.ReversalKR(a.Request()), nil, nil
		}
		a, err := reversallane.BuildProductionUSProposalAuthority(input)
		if err != nil {
			return strategyflow.LaneInput{}, nil, err
		}
		return strategyflow.ReversalUS(a.Request()), nil, nil
	}
	if !isWeeklyLane(scope.LaneID) || journalRO == nil {
		return strategyflow.LaneInput{}, nil, ErrProductionProposalUnavailable
	}
	reservation, err := journalRO.WeeklyMarketReservation(ctx, scope.CampaignID, string(config.Market), scope.StableWeek)
	if err != nil || reservation.ReservationID != scope.WeeklyReservationID {
		return strategyflow.LaneInput{}, nil, ErrProductionProposalUnavailable
	}
	market := weeklyvaluelane.MarketKR
	if config.Market == strategyrouter.MarketUS {
		market = weeklyvaluelane.MarketUS
	}
	input := weeklyvaluelane.ProductionProposalInput{Snapshot: snapshot, FX: fx, Market: market, AccountRef: config.AccountRef, Symbol: scope.Symbol,
		CandidateID: scope.CandidateID, CampaignID: scope.CampaignID, PositionGeneration: scope.PositionGeneration, RiskBudgetMinor: scope.RiskBudgetMinor,
		PerShareRiskMinor: scope.PerShareRiskMinor, PlannedQuantity: scope.PlannedQuantity, PolicyDigest: scope.PolicyDigest, AccountCurrency: scope.AccountCurrency,
		QuoteCurrency: scope.QuoteCurrency, ModelVersion: scope.ModelVersion, ModelConfigDigest: scope.ConfigDigest, ThresholdDigest: scope.ThresholdDigest,
		Reservation: weeklyvaluelane.ProductionReservationInput{ReservationID: reservation.ReservationID, CampaignID: reservation.CampaignID, Market: reservation.Market,
			StableWeek: reservation.StableWeek, Provider: reservation.Provider, TimeZone: reservation.TimeZone, SessionDate: reservation.SessionDate,
			CalendarGeneration: reservation.CalendarGeneration, CalendarDigest: reservation.CalendarDigest, Status: reservation.Status,
			RequestDigest: reservation.RequestDigest, RecordDigest: reservation.RecordDigest, PlannedOrdinal: reservation.PlannedOrdinal,
			ScopeVersion: reservation.ScopeVersion, PositiveLegCount: reservation.PositiveLegCount, ObservedAt: reservation.ObservedAt,
			FreshUntil: reservation.FreshUntil, EvaluatedAt: reservation.EvaluatedAt}, Leg: weeklyvaluelane.LegProgress{Ordinal: scope.LegOrdinal, FilledQuantity: scope.FilledQuantity},
		Stop: weeklyvaluelane.ProductionStopInput{PriceMinor: scope.Stop.PriceMinor, Version: scope.Stop.Version, Source: scope.Stop.Source, Policy: scope.Stop.Policy,
			Digest: scope.Stop.Digest, ObservedAt: stopObserved, FreshUntil: stopFresh}, SavedEffectiveStopMinor: scope.SavedEffectiveStopMinor,
		EntryPriceMinor: scope.EntryPriceMinor, StagedTargetMinor: scope.StagedTargetMinor, EntryCostsMinor: scope.EntryCostsMinor,
		EstimatedExitCostsMinor: scope.EstimatedExitCostsMinor, MinimumRRPPM: scope.MinimumRRPPM}
	weekly := &journal.WeeklyFirstLegReservationBinding{ReservationID: reservation.ReservationID, StableWeek: reservation.StableWeek,
		PlannedOrdinal: reservation.PlannedOrdinal, ScopeVersion: reservation.ScopeVersion, RequestDigest: reservation.RequestDigest,
		RecordDigest: reservation.RecordDigest, CalendarGeneration: reservation.CalendarGeneration, CalendarDigest: reservation.CalendarDigest}
	if market == weeklyvaluelane.MarketKR {
		a, err := weeklyvaluelane.BuildProductionKRProposalAuthority(input)
		if err != nil {
			return strategyflow.LaneInput{}, nil, err
		}
		return strategyflow.WeeklyKR(a.Request()), weekly, nil
	}
	a, err := weeklyvaluelane.BuildProductionUSProposalAuthority(input)
	if err != nil {
		return strategyflow.LaneInput{}, nil, err
	}
	return strategyflow.WeeklyUS(a.Request()), weekly, nil
}

func reversalStructure(value productionStructure, market reversallane.Market, account, symbol string, generation uint64, schema string) (reversallane.StructuralConfirmation, error) {
	convert := func(item productionStructureEvent, want reversallane.StructuralEventKind) (reversallane.StructuralEvent, error) {
		if item.Kind == "" && item.RecordID == "" && item.Digest == "" && item.At == "" && item.FreshUntil == "" {
			return reversallane.StructuralEvent{}, nil
		}
		at, ok1 := parseTime(item.At)
		fresh, ok2 := parseTime(item.FreshUntil)
		if !ok1 || !ok2 || item.Kind != string(want) {
			return reversallane.StructuralEvent{}, ErrProductionProposalUnavailable
		}
		return reversallane.StructuralEvent{Kind: want, AccountRef: account, Market: market, Symbol: symbol, PositionGeneration: generation,
			EvidenceVersion: schema, RecordID: item.RecordID, Digest: item.Digest, At: at, FreshUntil: fresh}, nil
	}
	sweep, e1 := convert(value.Sweep, reversallane.EventSweep)
	broken, e2 := convert(value.Break, reversallane.EventBreak)
	reclaim, e3 := convert(value.Reclaim, reversallane.EventReclaim)
	if e1 != nil || e2 != nil || e3 != nil {
		return reversallane.StructuralConfirmation{}, ErrProductionProposalUnavailable
	}
	return reversallane.StructuralConfirmation{Sweep: sweep, Break: broken, Reclaim: reclaim}, nil
}

func canonicalTargets(market strategyrouter.Market, targets []ProductionTarget) (map[string]ProductionTarget, bool) {
	result := make(map[string]ProductionTarget, len(targets))
	for _, target := range targets {
		if !target.Approved.Valid() || target.Approved.Market() != string(market) || target.Router.Key.Market != market || target.Router.Key.Symbol != target.Approved.Symbol() || result[target.Approved.Symbol()].Approved.Valid() {
			return nil, false
		}
		result[target.Approved.Symbol()] = target
	}
	return result, len(result) == len(targets)
}

func canonicalConfig(config ProductionConfig) ProductionConfig {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.EvidencePath = filepath.Clean(strings.TrimSpace(config.EvidencePath))
	config.JournalPath = filepath.Clean(strings.TrimSpace(config.JournalPath))
	config.AccountRef = strings.TrimSpace(config.AccountRef)
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	config.ObservedAt = config.ObservedAt.UTC()
	return config
}

func decodeManifest(data []byte) (productionManifest, error) {
	var manifest productionManifest
	if len(data) == 0 || len(data) > productionMaximumBytes || strictDecode(data, &manifest) != nil {
		return productionManifest{}, ErrProductionProposalUnavailable
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionManifest{}, ErrProductionProposalUnavailable
	}
	return manifest, nil
}

func strictDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrProductionProposalUnavailable
	}
	return nil
}

func verifyManifest(manifest productionManifest, config ProductionConfig) bool {
	body := manifest.productionBody
	observed, ok1 := parseTime(body.ObservedAt)
	fresh, ok2 := parseTime(body.FreshUntil)
	if body.SchemaVersion != productionSchema || body.Domain != productionDomain || body.SignatureAlgorithm != productionAlgorithm || body.KeyID != config.TrustedKeyID || body.Generation == 0 || body.AccountRef != config.AccountRef || body.Market != config.Market || body.RouteManifestDigest != config.RouteManifestDigest || body.ActivationDigest != config.ActivationDigest || body.CalendarGeneration != config.CalendarGeneration || body.CalendarDigest != config.CalendarDigest || body.SchedulerConfigVersion != config.SchedulerConfigVersion || body.EvidenceDBIdentity != config.EvidenceDBIdentity || !identity(body.Actor) || body.Revoked || !ok1 || !ok2 || observed.After(config.ObservedAt) || !config.ObservedAt.Before(fresh) || observed.After(fresh) || !validScopes(body.Market, body.Scopes) {
		return false
	}
	canonical, err := json.Marshal(body)
	if err != nil {
		return false
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	return err == nil && base64.StdEncoding.EncodeToString(signature) == manifest.Signature && len(signature) == ed25519.SignatureSize && ed25519.Verify(config.TrustedKey, canonical, signature)
}

func validScopes(market strategyrouter.Market, scopes []productionScope) bool {
	if len(scopes) == 0 || len(scopes) > 10_000 {
		return false
	}
	previous := ""
	for _, scope := range scopes {
		// 한 종목이 여러 가족을 동시에 제안할 수 있으므로 순서와 유일성의 단위는
		// 종목이 아니라 (종목, 레인) 쌍이다. 종목만으로 접으면 뒤 스코프가 앞을 덮어쓴다.
		key := scope.Symbol + "\x00" + scope.LaneID
		if scope.Symbol == "" || scope.Symbol != strings.ToUpper(strings.TrimSpace(scope.Symbol)) || key <= previous || scope.PositionGeneration == 0 || scope.CandidateID == "" || scope.CampaignID == "" || scope.SnapshotID == "" || scope.SnapshotDigest == "" || scope.PlannedQuantity == 0 || scope.LegOrdinal < 1 || scope.LegOrdinal > 7 || !laneMatchesMarket(market, scope.LaneID, scope.LaneVersion, scope.Horizon) {
			return false
		}
		previous = key
	}
	return true
}

func laneMatchesMarket(market strategyrouter.Market, lane, version string, horizon strategyrouter.Horizon) bool {
	if market == strategyrouter.MarketKR {
		return (lane == continuationlane.KRContinuationLaneID && version == continuationlane.LaneVersionV1 && horizon == strategyrouter.HorizonShort) || (lane == reversallane.KRReversalLaneID && version == reversallane.LaneVersionV1 && horizon == strategyrouter.HorizonShort) || (lane == weeklyvaluelane.KRWeeklyLaneID && version == weeklyvaluelane.LaneVersionV1 && horizon == strategyrouter.HorizonWeekly) || (lane == breakoutlane.KRLaneID && version == breakoutlane.LaneVersionV1 && horizon == strategyrouter.HorizonShort)
	}
	if market == strategyrouter.MarketUS {
		return (lane == continuationlane.USContinuationLaneID && version == continuationlane.LaneVersionV1 && horizon == strategyrouter.HorizonShort) || (lane == reversallane.USReversalLaneID && version == reversallane.LaneVersionV1 && horizon == strategyrouter.HorizonShort) || (lane == weeklyvaluelane.USWeeklyLaneID && version == weeklyvaluelane.LaneVersionV1 && horizon == strategyrouter.HorizonWeekly) || (lane == breakoutlane.USLaneID && version == breakoutlane.LaneVersionV1 && horizon == strategyrouter.HorizonShort)
	}
	return false
}

func isWeeklyLane(lane string) bool {
	return lane == weeklyvaluelane.KRWeeklyLaneID || lane == weeklyvaluelane.USWeeklyLaneID
}
func parseTime(raw string) (time.Time, bool) {
	value, err := time.Parse(time.RFC3339Nano, raw)
	return value.UTC(), err == nil && !value.IsZero() && value.Location() == time.UTC && value.UTC().Format(time.RFC3339Nano) == raw
}
func identity(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n")
}
func digestValid(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") || value != strings.ToLower(value) {
		return false
	}
	raw, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(raw) == sha256.Size
}
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
