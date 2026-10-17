package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutsignal"
)

// effectivePolicy is the complete non-secret policy that determines a shadow lot.
// It deliberately excludes paths, credentials, tokens, and other runtime locations.
const (
	policyOpeningRangeMinutes  = 15
	policyBreakoutBufferPPM    = uint64(100_000)
	policyTimeoutKR            = 8
	policyTimeoutUS            = 10
	policyRVOLMinPPM           = uint64(1_500_000)
	policyUpperWickRangeMaxPPM = uint64(350_000)
	policyFinalCap             = uint64(math.MaxUint64)
)

type effectivePolicy struct {
	Version              string       `json:"version"`
	Market               string       `json:"market"`
	SessionSource        string       `json:"session_source"`
	Top                  int          `json:"top"`
	Params               shadowParams `json:"params"`
	ReadGapNS            int64        `json:"read_gap_ns"`
	RetryDelayNS         []int64      `json:"retry_delay_ns"`
	QuoteSuspendNS       int64        `json:"quote_suspend_ns"`
	ArmedAttempts        int          `json:"armed_attempts"`
	ArmedWindowNS        int64        `json:"armed_window_ns"`
	GapPersistNS         int64        `json:"gap_persist_ns"`
	LabelGraceNS         int64        `json:"label_grace_ns"`
	BudgetEveryNS        int64        `json:"budget_every_ns"`
	PollLookbackNS       int64        `json:"poll_lookback_ns"`
	MetricsVersion       string       `json:"metrics_version"`
	LaneVersion          string       `json:"lane_version"`
	OpeningRangeMinutes  int          `json:"opening_range_minutes"`
	BreakoutBufferPPM    uint64       `json:"breakout_buffer_ppm"`
	TimeoutKR            int          `json:"timeout_kr"`
	TimeoutUS            int          `json:"timeout_us"`
	RVOLMinPPM           uint64       `json:"rvol_min_ppm"`
	UpperWickRangeMaxPPM uint64       `json:"upper_wick_range_max_ppm"`
	FinalCap             uint64       `json:"final_cap"`
	ReportMinSamples     int          `json:"report_min_samples"`
	LabelTiePolicy       string       `json:"label_tie_policy"`
	LabelFinalPolicy     string       `json:"label_final_policy"`
}

func effectivePolicyFor(cfg probeConfig, source string) effectivePolicy {
	retry := make([]int64, len(cfg.RetryDelays))
	for i, delay := range cfg.RetryDelays {
		retry[i] = int64(delay)
	}
	return effectivePolicy{
		Version: "a140-effective-policy/v1", Market: "KR", SessionSource: source, Top: cfg.Top, Params: cfg.Params,
		ReadGapNS: int64(cfg.ReadGap), RetryDelayNS: retry, QuoteSuspendNS: int64(cfg.QuoteSuspend),
		ArmedAttempts: cfg.ArmedAttempts, ArmedWindowNS: int64(cfg.ArmedWindow), GapPersistNS: int64(cfg.GapPersist),
		LabelGraceNS: int64(cfg.LabelGrace), BudgetEveryNS: int64(cfg.BudgetEvery), PollLookbackNS: int64(cfg.PollLookback),
		MetricsVersion: breakoutsignal.Version, LaneVersion: breakoutlane.LaneVersionV1,
		OpeningRangeMinutes: policyOpeningRangeMinutes, BreakoutBufferPPM: policyBreakoutBufferPPM, TimeoutKR: policyTimeoutKR, TimeoutUS: policyTimeoutUS,
		RVOLMinPPM: policyRVOLMinPPM, UpperWickRangeMaxPPM: policyUpperWickRangeMaxPPM, FinalCap: policyFinalCap,
		ReportMinSamples: minVerdictSamples, LabelTiePolicy: "stop", LabelFinalPolicy: "session-close",
	}
}

func (p effectivePolicy) digest() string {
	raw, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func lotForPolicy(digest string) string { return "a140-effective-policy/v1@" + digest }

// sessionIdentity is per-session provenance. It is intentionally outside the lot,
// which aggregates compatible policy across dates.
type sessionIdentity struct {
	SessionID       string `json:"session_id"`
	CalendarVersion string `json:"calendar_version"`
	Open            string `json:"open"`
	Close           string `json:"close"`
}

func sessionIdentityFor(session sessionInfo) sessionIdentity {
	return sessionIdentity{SessionID: session.ID, CalendarVersion: session.CalendarVersion,
		Open:  session.Open.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		Close: session.Close.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")}
}
