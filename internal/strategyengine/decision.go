package strategyengine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

var decisionDigestPattern = regexp.MustCompile(`^(?:sha256:)?[0-9a-f]{64}$`)

// DecisionRecord is serialization-only evidence. Constructing or modifying it
// grants no authority; only ParkerConservativeLane.Evaluate can mint Decision.
type DecisionRecord struct {
	CandidateLifeID     string
	CandidateState      string
	CandidateFirstSeen  int64
	CandidateLastSeen   int64
	CandidateValidUntil int64
	CandidateApprovedAt int64
	Market              string
	Symbol              string
	LaneID              string
	LaneVersion         string
	SourceCommit        string
	SourceDigest        string
	ConstantsDigest     string
	ThresholdVersion    string
	ThresholdSetDigest  string
	EvidenceDigest      string
	MarketInputVersion  string
	CalendarSource      string
	CalendarVersion     string
	ConfigSource        string
	ConfigVersion       string
	IndicatorSource     string
	IndicatorVersion    string
	IndicatorComputedAt int64
	TradingDay          bool
	SessionOpenAt       int64
	SessionCloseAt      int64
	NoEntryAfter        int64
	BarSource           string
	BarAdjusted         bool
	BarOpenAt           int64
	BarClosedAt         int64
	EvaluatedAt         int64
	ExpiresAt           int64
	Open                string
	High                string
	Low                 string
	Close               string
	Volume              string
	Currency            string
	VWAP                string
	VWAPSlopePct        string
	EMA9                string
	LVNSpacePct         string
	TangledPct          string
	Expansion           string
	HVNAboveDistancePct string
	StateSource         string
	StateAt             int64
	PositionSource      string
	PositionAt          int64
	EntryPrice          string
	LivePrice           string
	LivePriceObserved   bool
	EntryPriceDriftPct  string
	StopPrice           string
	TargetPrice         string
	ExpectedRR          string
	AcceptReasons       [7]string
	Identity            string
}

type Decision struct {
	record DecisionRecord
	valid  bool
}

func (d Decision) Valid() bool { return d.valid }

func (d Decision) Record() DecisionRecord {
	if !d.valid {
		return DecisionRecord{}
	}
	return d.record
}

func mintDecision(record DecisionRecord) (Decision, error) {
	if !strings.HasPrefix(record.CandidateLifeID, "candidate-life:v1:sha256:") || record.CandidateState != "active" ||
		record.CandidateFirstSeen <= 0 || record.CandidateLastSeen < record.CandidateFirstSeen ||
		record.CandidateValidUntil <= record.CandidateLastSeen || record.CandidateApprovedAt <= 0 || record.Market != "KR" ||
		strings.TrimSpace(record.Symbol) != record.Symbol || record.Symbol == "" ||
		record.LaneID != LaneID || record.LaneVersion != LaneVersion ||
		record.SourceCommit != SourceCommit || record.SourceDigest != FrozenSourceSetDigest ||
		record.ConstantsDigest != constantsDigest() || strings.TrimSpace(record.ThresholdVersion) == "" ||
		!decisionDigestPattern.MatchString(record.ThresholdSetDigest) || !decisionDigestPattern.MatchString(record.EvidenceDigest) ||
		record.MarketInputVersion != MarketInputVersion || record.CalendarSource != CalendarSource ||
		strings.TrimSpace(record.CalendarVersion) == "" || record.ConfigSource != ConfigSource || record.ConfigVersion != ConfigVersion ||
		record.IndicatorSource != IndicatorSource || record.IndicatorVersion != IndicatorVersion ||
		record.IndicatorComputedAt != record.EvaluatedAt || !record.TradingDay || record.BarSource != CalendarSource || record.BarAdjusted {
		return Decision{}, fmt.Errorf("strategy decision: provenance binding mismatch")
	}
	if record.BarOpenAt <= 0 || record.BarClosedAt-record.BarOpenAt != int64(5*60*1e9) ||
		record.CandidateApprovedAt > record.EvaluatedAt || record.EvaluatedAt < record.CandidateLastSeen ||
		record.EvaluatedAt >= record.CandidateValidUntil || record.EvaluatedAt < record.BarClosedAt ||
		record.ExpiresAt != record.BarClosedAt+int64(15*time.Second+time.Nanosecond) ||
		record.EvaluatedAt >= record.ExpiresAt || record.Currency != "KRW" || record.StateSource != "official-symbol-state" ||
		record.SessionOpenAt <= 0 || record.SessionCloseAt <= record.SessionOpenAt ||
		record.NoEntryAfter <= record.SessionOpenAt || record.NoEntryAfter > record.SessionCloseAt ||
		record.EvaluatedAt < record.SessionOpenAt+int64(10*time.Minute) || record.EvaluatedAt > record.NoEntryAfter ||
		record.StateAt <= 0 || record.EvaluatedAt-record.StateAt < 0 || record.EvaluatedAt-record.StateAt > int64(30*1e9) ||
		record.PositionSource != "official-position" || record.PositionAt <= 0 || record.EvaluatedAt-record.PositionAt < 0 ||
		record.EvaluatedAt-record.PositionAt > int64(30*1e9) {
		return Decision{}, fmt.Errorf("strategy decision: stale or invalid evidence clock")
	}
	positiveFields := []string{
		record.Open, record.High, record.Low, record.Close, record.Volume, record.VWAP,
		record.VWAPSlopePct, record.EMA9, record.LVNSpacePct, record.EntryPrice, record.LivePrice,
		record.StopPrice, record.TargetPrice, record.ExpectedRR,
	}
	for _, raw := range positiveFields {
		value, ok := decisionDecimal(raw)
		if !ok || value.Sign() <= 0 {
			return Decision{}, fmt.Errorf("strategy decision: invalid positive evidence")
		}
	}
	tangled, ok := decisionDecimal(record.TangledPct)
	if !ok || tangled.Cmp(rat("0.35")) < 0 {
		return Decision{}, fmt.Errorf("strategy decision: invalid tangled evidence")
	}
	for _, optional := range []string{record.Expansion, record.HVNAboveDistancePct} {
		if optional != "" {
			_, valid := decisionDecimal(optional)
			if !valid {
				return Decision{}, fmt.Errorf("strategy decision: invalid optional evidence")
			}
		}
	}
	entry, _ := decisionDecimal(record.EntryPrice)
	stop, _ := decisionDecimal(record.StopPrice)
	target, _ := decisionDecimal(record.TargetPrice)
	closePrice, _ := decisionDecimal(record.Close)
	livePrice, _ := decisionDecimal(record.LivePrice)
	lvn, _ := decisionDecimal(record.LVNSpacePct)
	wantStop := new(big.Rat).Mul(entry, rat("0.993"))
	risk := new(big.Rat).Sub(entry, wantStop)
	wantTarget := new(big.Rat).Add(entry, new(big.Rat).Mul(risk, rat("3.0")))
	wantRR := new(big.Rat).Quo(new(big.Rat).Quo(new(big.Rat).Mul(entry, lvn), rat("100")), risk)
	drift := new(big.Rat).Sub(livePrice, entry)
	if drift.Sign() < 0 {
		drift.Neg(drift)
	}
	drift.Mul(new(big.Rat).Quo(drift, entry), rat("100"))
	if entry.Cmp(closePrice) != 0 || stop.Cmp(wantStop) != 0 || target.Cmp(wantTarget) != 0 ||
		record.ExpectedRR != roundedDecimalString(wantRR) || record.EntryPriceDriftPct != roundedDecimalString(drift) {
		return Decision{}, fmt.Errorf("strategy decision: invalid prices")
	}
	if !record.LivePriceObserved && livePrice.Cmp(closePrice) != 0 {
		return Decision{}, fmt.Errorf("strategy decision: invalid live-price fallback")
	}
	if record.HVNAboveDistancePct != "" {
		hvn, _ := decisionDecimal(record.HVNAboveDistancePct)
		if hvn.Cmp(lvn) < 0 {
			return Decision{}, fmt.Errorf("strategy decision: invalid HVN forward space")
		}
	}
	wantReasons := [7]string{"VWAP_ABOVE", "VWAP_SLOPE_UP", "EMA9_PULLBACK_CONFIRMED", "VOLUME_PROFILE_SPACE_OK", "RR_GE_2", "NOT_TANGLED", "NOT_AFTER_ENTRY_CUTOFF"}
	if record.AcceptReasons != wantReasons {
		return Decision{}, fmt.Errorf("strategy decision: accept reason order mismatch")
	}
	want, err := decisionIdentity(record)
	if err != nil || record.Identity != want {
		return Decision{}, fmt.Errorf("strategy decision: identity mismatch")
	}
	return Decision{record: record, valid: true}, nil
}

func decisionIdentity(record DecisionRecord) (string, error) {
	copy := record
	copy.Identity = ""
	canonical, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "strategy-decision:v1:sha256:" + hex.EncodeToString(sum[:]), nil
}

func decisionDecimal(raw string) (*big.Rat, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "eE/") {
		return nil, false
	}
	value, ok := new(big.Rat).SetString(raw)
	return value, ok
}
