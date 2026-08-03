// Package strategyprojection defines the single read-only operational truth
// shared by the console, private HTTP API, SSE snapshots and Unix transport.
// It contains data and validation only; it owns no order, activation, settings,
// journal or protection mutation capability.
package strategyprojection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const SchemaVersion = "tossos.strategy-runtime-projection/v1"

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type MarketStatus string

const (
	StatusCurrent MarketStatus = "CURRENT"
	StatusUnknown MarketStatus = "UNKNOWN"
)

type State string

const (
	StateOff State = "OFF"
	StateOn  State = "ON"
)

type Freshness string

const (
	FreshnessUnobserved Freshness = "UNOBSERVED"
	FreshnessCurrent    Freshness = "CURRENT"
	FreshnessStale      Freshness = "STALE"
	FreshnessUnknown    Freshness = "UNKNOWN"
)

type ComponentStatus string

const (
	ComponentCurrent ComponentStatus = "CURRENT"
	ComponentUnknown ComponentStatus = "UNKNOWN"
)

type ActivationStatus string

const (
	ActivationConfigured    ActivationStatus = "CONFIGURED"
	ActivationNotConfigured ActivationStatus = "NOT_CONFIGURED"
	ActivationUnknown       ActivationStatus = "UNKNOWN"
)

type ProtectionReadiness string

const (
	ProtectionWired   ProtectionReadiness = "WIRED"
	ProtectionUnwired ProtectionReadiness = "UNWIRED"
)

type ReconciliationStatus string

const (
	ReconciliationHealthy   ReconciliationStatus = "HEALTHY"
	ReconciliationUnhealthy ReconciliationStatus = "UNHEALTHY"
	ReconciliationUnknown   ReconciliationStatus = "UNKNOWN"
)

type RefusalCode string

const (
	RefusalNone                      RefusalCode = "NONE"
	RefusalNotConfigured             RefusalCode = "NOT_CONFIGURED"
	RefusalRuntimeUnavailable        RefusalCode = "RUNTIME_UNAVAILABLE"
	RefusalEvidenceStale             RefusalCode = "EVIDENCE_STALE"
	RefusalSchedulerBlocked          RefusalCode = "SCHEDULER_BLOCKED"
	RefusalActivationAbsent          RefusalCode = "ACTIVATION_ABSENT"
	RefusalProtectionUnwired         RefusalCode = "PROTECTION_UNWIRED"
	RefusalReconciliationUnavailable RefusalCode = "RECONCILIATION_UNAVAILABLE"
)

type Snapshot struct {
	SchemaVersion string                      `json:"schemaVersion"`
	GeneratedAt   time.Time                   `json:"generatedAt"`
	Markets       map[Market]MarketProjection `json:"markets"`
}

type MarketProjection struct {
	Market         Market                   `json:"market"`
	Status         MarketStatus             `json:"status"`
	Error          *MarketError             `json:"error"`
	Lane           LaneProjection           `json:"lane"`
	Evidence       EvidenceProjection       `json:"evidence"`
	Campaign       CampaignProjection       `json:"campaign"`
	HorizonRisk    HorizonRiskProjection    `json:"horizonRisk"`
	Scheduler      SchedulerProjection      `json:"scheduler"`
	Activation     ActivationProjection     `json:"activation"`
	Protection     ProtectionProjection     `json:"protection"`
	Reconciliation ReconciliationProjection `json:"reconciliation"`
	FirstRefusal   RefusalCode              `json:"firstRefusal"`
	ObservedAt     *time.Time               `json:"observedAt"`
}

type MarketError struct {
	Code       RefusalCode `json:"code"`
	ObservedAt time.Time   `json:"observedAt"`
}

type LaneProjection struct {
	ID        *string `json:"id"`
	Version   *string `json:"version"`
	Desired   State   `json:"desired"`
	Effective State   `json:"effective"`
}

type EvidenceProjection struct {
	ID        *string   `json:"id"`
	Digest    *string   `json:"digest"`
	Freshness Freshness `json:"freshness"`
}

type CampaignProjection struct {
	ID    *string `json:"id"`
	LegID *string `json:"legId"`
}

type HorizonRiskProjection struct {
	Bucket        *string         `json:"bucket"`
	PolicyVersion *string         `json:"policyVersion"`
	Status        ComponentStatus `json:"status"`
}

type SchedulerProjection struct {
	Desired           State     `json:"desired"`
	Effective         State     `json:"effective"`
	CalendarSource    *string   `json:"calendarSource"`
	CalendarVersion   *string   `json:"calendarVersion"`
	CalendarFreshness Freshness `json:"calendarFreshness"`
}

type ActivationProjection struct {
	Desired   State            `json:"desired"`
	Effective State            `json:"effective"`
	Digest    *string          `json:"digest"`
	Status    ActivationStatus `json:"status"`
}

type ProtectionProjection struct {
	Readiness ProtectionReadiness `json:"readiness"`
	Refusal   RefusalCode         `json:"refusal"`
}

type ReconciliationProjection struct {
	Status  ReconciliationStatus `json:"status"`
	Refusal RefusalCode          `json:"refusal"`
}

type Reader interface {
	Read(context.Context) (Snapshot, error)
}

func DormantSnapshot(generatedAt time.Time) Snapshot {
	generatedAt = generatedAt.UTC()
	return Snapshot{SchemaVersion: SchemaVersion, GeneratedAt: generatedAt, Markets: map[Market]MarketProjection{
		MarketKR: unknownMarket(MarketKR, RefusalNotConfigured, generatedAt),
		MarketUS: unknownMarket(MarketUS, RefusalNotConfigured, generatedAt),
	}}
}

func UnavailableSnapshot(generatedAt time.Time) Snapshot {
	generatedAt = generatedAt.UTC()
	return Snapshot{SchemaVersion: SchemaVersion, GeneratedAt: generatedAt, Markets: map[Market]MarketProjection{
		MarketKR: unknownMarket(MarketKR, RefusalRuntimeUnavailable, generatedAt),
		MarketUS: unknownMarket(MarketUS, RefusalRuntimeUnavailable, generatedAt),
	}}
}

func WithMarketFailure(snapshot Snapshot, market Market, code RefusalCode, observedAt time.Time) Snapshot {
	next := Clone(snapshot)
	if !validMarket(market) || code == RefusalNone || !validRefusal(code) || observedAt.IsZero() {
		return next
	}
	observedAt = observedAt.UTC()
	next.Markets[market] = unknownMarket(market, code, observedAt)
	if next.GeneratedAt.Before(observedAt) {
		next.GeneratedAt = observedAt
	}
	return next
}

func unknownMarket(market Market, code RefusalCode, observedAt time.Time) MarketProjection {
	freshness, activation := FreshnessUnknown, ActivationUnknown
	if code == RefusalNotConfigured {
		freshness, activation = FreshnessUnobserved, ActivationNotConfigured
	}
	return MarketProjection{Market: market, Status: StatusUnknown,
		Error:          &MarketError{Code: code, ObservedAt: observedAt.UTC()},
		Lane:           LaneProjection{Desired: StateOff, Effective: StateOff},
		Evidence:       EvidenceProjection{Freshness: freshness},
		HorizonRisk:    HorizonRiskProjection{Status: ComponentUnknown},
		Scheduler:      SchedulerProjection{Desired: StateOff, Effective: StateOff, CalendarFreshness: freshness},
		Activation:     ActivationProjection{Desired: StateOff, Effective: StateOff, Status: activation},
		Protection:     ProtectionProjection{Readiness: ProtectionUnwired, Refusal: code},
		Reconciliation: ReconciliationProjection{Status: ReconciliationUnknown, Refusal: code},
		FirstRefusal:   code}
}

func Clone(snapshot Snapshot) Snapshot {
	out := Snapshot{SchemaVersion: snapshot.SchemaVersion, GeneratedAt: snapshot.GeneratedAt, Markets: make(map[Market]MarketProjection, len(snapshot.Markets))}
	for market, item := range snapshot.Markets {
		out.Markets[market] = cloneMarket(item)
	}
	return out
}

func cloneMarket(item MarketProjection) MarketProjection {
	item.Error = cloneError(item.Error)
	item.Lane.ID, item.Lane.Version = cloneString(item.Lane.ID), cloneString(item.Lane.Version)
	item.Evidence.ID, item.Evidence.Digest = cloneString(item.Evidence.ID), cloneString(item.Evidence.Digest)
	item.Campaign.ID, item.Campaign.LegID = cloneString(item.Campaign.ID), cloneString(item.Campaign.LegID)
	item.HorizonRisk.Bucket, item.HorizonRisk.PolicyVersion = cloneString(item.HorizonRisk.Bucket), cloneString(item.HorizonRisk.PolicyVersion)
	item.Scheduler.CalendarSource, item.Scheduler.CalendarVersion = cloneString(item.Scheduler.CalendarSource), cloneString(item.Scheduler.CalendarVersion)
	item.Activation.Digest = cloneString(item.Activation.Digest)
	if item.ObservedAt != nil {
		value := item.ObservedAt.UTC()
		item.ObservedAt = &value
	}
	return item
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneError(value *MarketError) *MarketError {
	if value == nil {
		return nil
	}
	copyValue := *value
	copyValue.ObservedAt = copyValue.ObservedAt.UTC()
	return &copyValue
}

func OrderedMarkets(snapshot Snapshot) []MarketProjection {
	return []MarketProjection{cloneMarket(snapshot.Markets[MarketKR]), cloneMarket(snapshot.Markets[MarketUS])}
}

func Validate(snapshot Snapshot) error {
	if snapshot.SchemaVersion != SchemaVersion || snapshot.GeneratedAt.IsZero() || len(snapshot.Markets) != 2 {
		return errors.New("strategy projection: invalid envelope")
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		item, ok := snapshot.Markets[market]
		if !ok || item.Market != market {
			return fmt.Errorf("strategy projection: missing or cross-market %s", market)
		}
		if err := validateMarketProjection(item, snapshot.GeneratedAt); err != nil {
			return fmt.Errorf("strategy projection: market %s: %w", market, err)
		}
	}
	return nil
}

func validateMarketProjection(item MarketProjection, generatedAt time.Time) error {
	if !validState(item.Lane.Desired) || !validState(item.Lane.Effective) ||
		!validState(item.Scheduler.Desired) || !validState(item.Scheduler.Effective) ||
		!validState(item.Activation.Desired) || !validState(item.Activation.Effective) || !validRefusal(item.FirstRefusal) {
		return errors.New("invalid state or refusal")
	}
	if !pairedIdentity(item.Lane.ID, item.Lane.Version) || !pairedIdentity(item.Campaign.ID, item.Campaign.LegID) ||
		!pairedIdentity(item.HorizonRisk.Bucket, item.HorizonRisk.PolicyVersion) ||
		!pairedIdentity(item.Scheduler.CalendarSource, item.Scheduler.CalendarVersion) {
		return errors.New("partial identity pair")
	}
	if item.Lane.ID != nil && (!validIdentity(*item.Lane.ID) || !validIdentity(*item.Lane.Version)) ||
		item.Campaign.ID != nil && (!validIdentity(*item.Campaign.ID) || !validIdentity(*item.Campaign.LegID)) ||
		item.HorizonRisk.Bucket != nil && (!validIdentity(*item.HorizonRisk.Bucket) || !validIdentity(*item.HorizonRisk.PolicyVersion)) ||
		item.Scheduler.CalendarSource != nil && (!validIdentity(*item.Scheduler.CalendarSource) || !validIdentity(*item.Scheduler.CalendarVersion)) {
		return errors.New("noncanonical identity")
	}
	if err := validateEvidence(item.Evidence); err != nil {
		return err
	}
	if err := validateComponents(item); err != nil {
		return err
	}
	switch item.Status {
	case StatusCurrent:
		if item.Error != nil || item.ObservedAt == nil || item.ObservedAt.IsZero() || generatedAt.Before(*item.ObservedAt) ||
			item.Lane.ID == nil {
			return errors.New("current market lacks exact observation")
		}
	case StatusUnknown:
		if item.Error == nil || !validRefusal(item.Error.Code) || item.Error.Code == RefusalNone || item.Error.ObservedAt.IsZero() ||
			generatedAt.Before(item.Error.ObservedAt) || item.ObservedAt != nil || item.Lane.Desired != StateOff || item.Lane.Effective != StateOff ||
			item.Lane.ID != nil || item.Evidence.ID != nil || item.Campaign.ID != nil || item.HorizonRisk.Bucket != nil ||
			item.Scheduler.CalendarSource != nil || item.Activation.Digest != nil || item.Protection.Readiness != ProtectionUnwired ||
			item.Protection.Refusal != item.Error.Code || item.Reconciliation.Status != ReconciliationUnknown ||
			item.Reconciliation.Refusal != item.Error.Code || item.FirstRefusal != item.Error.Code {
			return errors.New("unknown market contains inferred or contradictory facts")
		}
		wantFreshness, wantActivation := FreshnessUnknown, ActivationUnknown
		if item.Error.Code == RefusalNotConfigured {
			wantFreshness, wantActivation = FreshnessUnobserved, ActivationNotConfigured
		}
		if item.Evidence.Freshness != wantFreshness || item.Scheduler.CalendarFreshness != wantFreshness || item.Activation.Status != wantActivation {
			return errors.New("unknown market freshness fallback")
		}
	default:
		return errors.New("invalid market status")
	}
	return nil
}

func validateEvidence(evidence EvidenceProjection) error {
	if !pairedIdentity(evidence.ID, evidence.Digest) {
		return errors.New("partial evidence")
	}
	switch evidence.Freshness {
	case FreshnessCurrent, FreshnessStale:
		if evidence.ID == nil || !validIdentity(*evidence.ID) || !validDigest(*evidence.Digest) {
			return errors.New("observed evidence lacks identity")
		}
	case FreshnessUnobserved, FreshnessUnknown:
		if evidence.ID != nil {
			return errors.New("unobserved evidence has inferred identity")
		}
	default:
		return errors.New("invalid evidence freshness")
	}
	return nil
}

func validateComponents(item MarketProjection) error {
	switch item.HorizonRisk.Status {
	case ComponentCurrent:
		if item.HorizonRisk.Bucket == nil {
			return errors.New("current risk lacks identity")
		}
	case ComponentUnknown:
		if item.HorizonRisk.Bucket != nil {
			return errors.New("unknown risk has inferred identity")
		}
	default:
		return errors.New("invalid risk status")
	}
	switch item.Scheduler.CalendarFreshness {
	case FreshnessCurrent, FreshnessStale:
		if item.Scheduler.CalendarSource == nil {
			return errors.New("observed calendar lacks identity")
		}
	case FreshnessUnobserved, FreshnessUnknown:
		if item.Scheduler.CalendarSource != nil {
			return errors.New("unknown calendar has inferred identity")
		}
	default:
		return errors.New("invalid calendar freshness")
	}
	switch item.Activation.Status {
	case ActivationConfigured:
		if item.Activation.Digest == nil || !validDigest(*item.Activation.Digest) {
			return errors.New("configured activation lacks digest")
		}
	case ActivationNotConfigured, ActivationUnknown:
		if item.Activation.Digest != nil {
			return errors.New("unknown activation has inferred digest")
		}
	default:
		return errors.New("invalid activation status")
	}
	if item.Protection.Readiness != ProtectionWired && item.Protection.Readiness != ProtectionUnwired {
		return errors.New("invalid protection readiness")
	}
	if item.Protection.Readiness == ProtectionWired && item.Protection.Refusal != RefusalNone ||
		item.Protection.Readiness == ProtectionUnwired && (item.Protection.Refusal == RefusalNone || !validRefusal(item.Protection.Refusal)) {
		return errors.New("contradictory protection readiness")
	}
	switch item.Reconciliation.Status {
	case ReconciliationHealthy:
		if item.Reconciliation.Refusal != RefusalNone {
			return errors.New("healthy reconciliation has refusal")
		}
	case ReconciliationUnhealthy, ReconciliationUnknown:
		if item.Reconciliation.Refusal == RefusalNone || !validRefusal(item.Reconciliation.Refusal) {
			return errors.New("blocked reconciliation lacks refusal")
		}
	default:
		return errors.New("invalid reconciliation status")
	}
	return nil
}

func validMarket(market Market) bool { return market == MarketKR || market == MarketUS }
func validState(state State) bool    { return state == StateOff || state == StateOn }

func validRefusal(code RefusalCode) bool {
	switch code {
	case RefusalNone, RefusalNotConfigured, RefusalRuntimeUnavailable, RefusalEvidenceStale, RefusalSchedulerBlocked,
		RefusalActivationAbsent, RefusalProtectionUnwired, RefusalReconciliationUnavailable:
		return true
	default:
		return false
	}
}

func pairedIdentity(left, right *string) bool { return (left == nil) == (right == nil) }

func validIdentity(value string) bool {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func digestString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
