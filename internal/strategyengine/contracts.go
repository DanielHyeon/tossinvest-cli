package strategyengine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
)

const (
	LaneID                = "krx_parker_vwap_conservative_v1"
	LaneVersion           = "1"
	SourceCommit          = "d75113d3c338148606d86c8aedbbeb7ed446c0b8"
	FrozenSourceSetDigest = "09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd"
	MarketInputVersion    = "parker-market-input:v1"
	CalendarSource        = "official-open-api"
	ConfigSource          = "stockos-frozen-config"
	ConfigVersion         = "parker-vwap-conservative:v1"
	IndicatorSource       = "stockos-parker-vwap"
	IndicatorVersion      = SourceCommit
)

type Refusal string

const (
	RefusalNone             Refusal = ""
	RefusalCandidate        Refusal = "candidate_not_approved"
	RefusalUnsupportedScope Refusal = "unsupported_scope"
	RefusalSession          Refusal = "session_closed"
	RefusalBarIntegrity     Refusal = "bar_integrity"
	RefusalSymbolState      Refusal = "symbol_state_unavailable"
	RefusalExistingPosition Refusal = "existing_position"
	RefusalInvalidBar       Refusal = "invalid_bar_values"
	RefusalIlliquidBar      Refusal = "illiquid_bar"
	RefusalIndicator        Refusal = "indicator_incomplete"
	RefusalVWAPAbove        Refusal = "vwap_above"
	RefusalVWAPSlope        Refusal = "vwap_slope"
	RefusalEMA9Pullback     Refusal = "ema9_bullish_pullback"
	RefusalFakeBreakout     Refusal = "fake_breakout"
	RefusalLVNSpace         Refusal = "lvn_forward_space"
	RefusalTangledBand      Refusal = "tangled_band"
	RefusalBandExpansion    Refusal = "band_expansion"
	RefusalRR               Refusal = "expected_rr"
	RefusalHVNCeiling       Refusal = "hvn_ceiling"
	RefusalAge              Refusal = "signal_age"
	RefusalDrift            Refusal = "entry_price_drift"
	RefusalSource           Refusal = "source_not_configured"
	RefusalDecision         Refusal = "decision_invalid"
)

const (
	SourceRejectProfileDisabled      = "REJECT_PROFILE_DISABLED"
	SourceRejectScopeFrozen          = "REJECT_SCOPE_FROZEN"
	SourceRejectNonTradingDay        = "REJECT_NON_TRADING_DAY"
	SourceRejectAfterHours           = "REJECT_AFTER_HOURS"
	SourceRejectOpeningWindow        = "REJECT_OPENING_WINDOW"
	SourceRejectAfterEntryCutoff     = "REJECT_AFTER_ENTRY_CUTOFF"
	SourceRejectBarNotClosed         = "REJECT_BAR_NOT_CLOSED"
	SourceRejectSymbolStateStale     = "REJECT_SYMBOL_STATE_STALE"
	SourceRejectPositionAlreadyOpen  = "REJECT_POSITION_ALREADY_OPEN"
	SourceRejectIlliquidBar          = "REJECT_ILLIQUID_BAR"
	SourceRejectIndicatorUnavailable = "REJECT_INDICATOR_UNAVAILABLE"
	SourceRejectVWAPBelow            = "REJECT_VWAP_BELOW"
	SourceRejectVWAPSlopeDown        = "REJECT_VWAP_SLOPE_DOWN"
	SourceRejectEMA9PullbackMissing  = "REJECT_EMA9_PULLBACK_MISSING"
	SourceRejectFakeBreakout         = "REJECT_FAKE_BREAKOUT"
	SourceRejectHVNBlock             = "REJECT_HVN_BLOCK"
	SourceRejectTangled              = "REJECT_TANGLED"
	SourceRejectVolatilityExpansion  = "REJECT_VOLATILITY_EXPANSION"
	SourceRejectRRTooLow             = "REJECT_RR_TOO_LOW"
	SourceRejectStaleSignal          = "REJECT_STALE_SIGNAL"
	SourceRejectPriceDrift           = "REJECT_PRICE_DRIFT"
)

type SourceBlob struct {
	Path       string
	BlobSHA256 string
}

// SourceProof has no exported fields. A caller cannot turn a boolean or a
// caller-selected digest into provenance; only VerifyFrozenSource can mint it.
type SourceProof struct {
	digest string
	valid  bool
}

func (p SourceProof) Valid() bool {
	return p.valid && p.digest == FrozenSourceSetDigest
}

func (p SourceProof) Digest() string {
	if !p.Valid() {
		return ""
	}
	return p.digest
}

func VerifyFrozenSource(blobs []SourceBlob) (SourceProof, error) {
	if len(blobs) == 0 {
		return SourceProof{}, fmt.Errorf("strategy source manifest: unavailable")
	}
	previous := ""
	hash := sha256.New()
	for _, blob := range blobs {
		path := strings.TrimSpace(blob.Path)
		digest := strings.TrimSpace(blob.BlobSHA256)
		if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") || strings.Contains(path, "..") {
			return SourceProof{}, fmt.Errorf("strategy source manifest: invalid relative path")
		}
		if previous != "" && path <= previous {
			return SourceProof{}, fmt.Errorf("strategy source manifest: paths are not strictly sorted")
		}
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size || digest != strings.ToLower(digest) {
			return SourceProof{}, fmt.Errorf("strategy source manifest: invalid blob digest")
		}
		_, _ = hash.Write([]byte(path))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(digest))
		_, _ = hash.Write([]byte{'\n'})
		previous = path
	}
	got := hex.EncodeToString(hash.Sum(nil))
	if got != FrozenSourceSetDigest {
		return SourceProof{}, fmt.Errorf("strategy source manifest: frozen digest mismatch")
	}
	return SourceProof{digest: got, valid: true}, nil
}

type LaneInput struct {
	Approved strategy.ApprovedSnapshot
	Source   SourceProof
	Bar      strategymarket.VerifiedBar
	State    strategymarket.FreshNormalState
	Position strategymarket.NoPositionProof
	Market   VersionedMarketInput
}

// MarketInputFields is the scalar DTO accepted by the sealer. Expected RR,
// signal age, price drift, and an HVN ceiling price are deliberately absent:
// those values are derived inside the lane from frozen observations.
type MarketInputFields struct {
	Version                                string
	Market, Symbol                         string
	CalendarSource, CalendarVersion        string
	ConfigSource, ConfigVersion            string
	IndicatorSource, IndicatorVersion      string
	EvaluatedAt, IndicatorComputedAt       time.Time
	TradingDay                             bool
	SessionOpenAt, SessionCloseAt          time.Time
	NoEntryAfter                           time.Time
	VWAP, VWAPSlopePct, EMA9               string
	LVNForwardSpacePct, TangledScorePct    string
	BandExpansionRate, HVNAboveDistancePct string
	CurrentPrice                           string
}

// VersionedMarketInput is an immutable, provenance-bearing market/config/
// indicator bundle. A zero value is always refused.
type VersionedMarketInput struct {
	valid                                  bool
	version                                string
	market, symbol                         string
	calendarSource, calendarVersion        string
	configSource, configVersion            string
	indicatorSource, indicatorVersion      string
	evaluatedAt, indicatorComputedAt       time.Time
	tradingDay                             bool
	sessionOpenAt, sessionCloseAt          time.Time
	noEntryAfter                           time.Time
	vwap, vwapSlopePct, ema9               string
	lvnForwardSpacePct, tangledScorePct    string
	bandExpansionRate, hvnAboveDistancePct string
	currentPrice                           string
}

func SealVersionedMarketInput(fields MarketInputFields) (VersionedMarketInput, error) {
	if fields.Version != MarketInputVersion || fields.Market != "KR" || fields.Symbol == "" ||
		strings.TrimSpace(fields.Symbol) != fields.Symbol || fields.CalendarSource != CalendarSource ||
		strings.TrimSpace(fields.CalendarVersion) == "" || fields.ConfigSource != ConfigSource ||
		fields.ConfigVersion != ConfigVersion || fields.IndicatorSource != IndicatorSource ||
		fields.IndicatorVersion != IndicatorVersion || fields.EvaluatedAt.IsZero() ||
		!fields.IndicatorComputedAt.Equal(fields.EvaluatedAt) {
		return VersionedMarketInput{}, fmt.Errorf("strategy market input: provenance mismatch")
	}
	if fields.TradingDay && (fields.SessionOpenAt.IsZero() || fields.SessionCloseAt.IsZero() ||
		fields.NoEntryAfter.IsZero() || !fields.SessionOpenAt.Before(fields.NoEntryAfter) ||
		fields.NoEntryAfter.After(fields.SessionCloseAt)) {
		return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid calendar window")
	}
	for _, raw := range []string{fields.VWAP, fields.EMA9} {
		if _, ok := positive(raw); !ok {
			return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid required decimal")
		}
	}
	if _, ok := decimal(fields.VWAPSlopePct); !ok {
		return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid slope decimal")
	}
	if _, ok := decimal(fields.LVNForwardSpacePct); !ok {
		return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid LVN decimal")
	}
	if _, ok := nonnegative(fields.TangledScorePct); !ok {
		return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid tangled decimal")
	}
	if fields.CurrentPrice != "" {
		if _, ok := decimal(fields.CurrentPrice); !ok {
			return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid live-price decimal")
		}
	}
	for _, raw := range []string{fields.BandExpansionRate, fields.HVNAboveDistancePct} {
		if raw != "" {
			if _, ok := decimal(raw); !ok {
				return VersionedMarketInput{}, fmt.Errorf("strategy market input: invalid optional decimal")
			}
		}
	}
	return VersionedMarketInput{
		valid: true, version: fields.Version, market: fields.Market, symbol: fields.Symbol,
		calendarSource: fields.CalendarSource, calendarVersion: fields.CalendarVersion,
		configSource: fields.ConfigSource, configVersion: fields.ConfigVersion,
		indicatorSource: fields.IndicatorSource, indicatorVersion: fields.IndicatorVersion,
		evaluatedAt: fields.EvaluatedAt.UTC(), indicatorComputedAt: fields.IndicatorComputedAt.UTC(),
		tradingDay: fields.TradingDay, sessionOpenAt: fields.SessionOpenAt.UTC(),
		sessionCloseAt: fields.SessionCloseAt.UTC(), noEntryAfter: fields.NoEntryAfter.UTC(),
		vwap: fields.VWAP, vwapSlopePct: fields.VWAPSlopePct, ema9: fields.EMA9,
		lvnForwardSpacePct: fields.LVNForwardSpacePct, tangledScorePct: fields.TangledScorePct,
		bandExpansionRate: fields.BandExpansionRate, hvnAboveDistancePct: fields.HVNAboveDistancePct,
		currentPrice: fields.CurrentPrice,
	}, nil
}

type Evaluation struct {
	Decision     Decision
	Reason       Refusal
	SourceReason string
}

type EntryLane interface {
	Evaluate(LaneInput) Evaluation
}
