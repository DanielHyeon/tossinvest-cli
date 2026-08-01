// Package strategyengine contains the dormant, authority-free strategy domain
// and orchestration seams. No production runtime constructs an activation or
// wires these interfaces in a047.
package strategyengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
)

const (
	LaneID                = "krx_parker_vwap_conservative_v1"
	LaneVersion           = "1"
	SourceCommit          = "d75113d3c338148606d86c8aedbbeb7ed446c0b8"
	FrozenSourceSetDigest = "09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd"
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
	RefusalZeroVolume       Refusal = "zero_volume"
	RefusalIndicator        Refusal = "indicator_incomplete"
	RefusalVWAPAbove        Refusal = "vwap_above"
	RefusalVWAPSlope        Refusal = "vwap_slope"
	RefusalEMA9Pullback     Refusal = "ema9_bullish_pullback"
	RefusalLVNSpace         Refusal = "lvn_forward_space"
	RefusalTangledBand      Refusal = "tangled_band"
	RefusalBandExpansion    Refusal = "band_expansion"
	RefusalRR               Refusal = "expected_rr"
	RefusalHVNCeiling       Refusal = "hvn_ceiling"
	RefusalAge              Refusal = "signal_age"
	RefusalDrift            Refusal = "entry_price_drift"
	RefusalSource           Refusal = "source_not_configured"
	RefusalLaneOff          Refusal = "lane_off"
	RefusalKillSwitch       Refusal = "kill_switch"
	RefusalProtection       Refusal = "protection_unwired"
	RefusalGate             Refusal = "gate_closed"
	RefusalManifest         Refusal = "activation_manifest_invalid"
)

// LaneInput contains only versioned, already-observed values. Exact decimal
// strings remain strings until the pure evaluator parses them.
type LaneInput struct {
	Approved                                             strategy.ApprovedCandidate
	Market, Session                                      string
	SourceVerified, RegularSession, BarsClosedContiguous bool
	SymbolStateNormal, NoExistingPosition                bool
	Volume, Price, VWAP, VWAPSlopePct, EMA9, EMA20       string
	LVNForwardSpacePct, BandExpansionRate, ExpectedRR    string
	Untangled, HVNCeilingClear                           bool
	SignalAgeSeconds                                     int64
	EntryPriceDriftPct                                   string
}

type EntryDecision struct {
	Accepted                                                              bool
	Reason                                                                Refusal
	CandidateLifeID, ThresholdVersion, ThresholdSetDigest, EvidenceDigest string
	LaneID, LaneVersion, SourceCommit, SourceSetDigest, ConstantsDigest   string
	Market, Symbol, EntryPrice, StopPrice, TargetPrice, ExpectedRR        string
	Identity                                                              string
}

type EntryLane interface{ Evaluate(LaneInput) EntryDecision }

// SourceBlob is one reproducible source-manifest row. Paths must be clean,
// relative slash paths and blob digests lowercase SHA-256 hex.
type SourceBlob struct{ Path, BlobSHA256 string }

func VerifyFrozenSource(blobs []SourceBlob) (string, error) {
	if len(blobs) == 0 {
		return "", fmt.Errorf("strategy source manifest: unavailable")
	}
	previous := ""
	h := sha256.New()
	for _, blob := range blobs {
		path := strings.TrimSpace(blob.Path)
		digest := strings.TrimSpace(blob.BlobSHA256)
		if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") || strings.Contains(path, "..") {
			return "", fmt.Errorf("strategy source manifest: invalid relative path")
		}
		if previous != "" && path <= previous {
			return "", fmt.Errorf("strategy source manifest: paths are not strictly sorted")
		}
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size || digest != strings.ToLower(digest) {
			return "", fmt.Errorf("strategy source manifest: invalid blob digest")
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(digest))
		_, _ = h.Write([]byte{'\n'})
		previous = path
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != FrozenSourceSetDigest {
		return got, fmt.Errorf("strategy source manifest: frozen digest mismatch")
	}
	return got, nil
}

// ManifestBinding is the complete expected authority snapshot. Every field is
// compared exactly at decision and again immediately before durable planning.
type ManifestBinding struct {
	AccountRef, Profile, BuildDigest, CommitDigest                                 string
	LaneID, LaneVersion, LaneSourceDigest, LaneConstantsDigest                     string
	ThresholdVersion, ThresholdSetDigest, EvidenceDigest, SettingsDigest           string
	AttestationDigest                                                              string
	AttestationExpiresAt                                                           time.Time
	GuardianVersion, GuardianLimitsDigest, ReconciliationWatermark                 string
	ProtectionProfile, ProtectionState, OperatingPolicy                            string
	SchedulerScope, CalendarVersion                                                string
	LaneApproved, SchedulerApproved, AutoStartApproved, GateApproved, LiveApproved bool
	Actor, AuditID                                                                 string
	IssuedAt, ExpiresAt                                                            time.Time
	Generation                                                                     uint64
}

type activationManifest struct {
	binding ManifestBinding
	digest  string
	revoked bool
}

// Activation is deliberately opaque. No exported function in this change
// installs or mints one; the repository shipped to production is empty.
type Activation struct {
	digest     string
	generation uint64
	valid      bool
}

func (a Activation) Digest() string {
	if !a.valid {
		return ""
	}
	return a.digest
}

type ManifestRepository struct {
	mu      sync.RWMutex
	current activationManifest
}

func NewDormantManifestRepository() *ManifestRepository { return &ManifestRepository{} }

var ErrActivationNotConfigured = errors.New("strategy activation: not configured")

func (r *ManifestRepository) Verify(expected ManifestBinding, now time.Time) (Activation, error) {
	if r == nil || now.IsZero() {
		return Activation{}, ErrActivationNotConfigured
	}
	r.mu.RLock()
	manifest := r.current
	r.mu.RUnlock()
	if manifest.digest == "" {
		return Activation{}, ErrActivationNotConfigured
	}
	if manifest.revoked {
		return Activation{}, fmt.Errorf("strategy activation: revoked")
	}
	canonicalDigest, err := manifestDigest(manifest.binding)
	if err != nil || manifest.digest != canonicalDigest {
		return Activation{}, fmt.Errorf("strategy activation: manifest digest mismatch")
	}
	if !sameBinding(manifest.binding, expected) {
		return Activation{}, fmt.Errorf("strategy activation: binding mismatch")
	}
	if expected.Generation == 0 || expected.IssuedAt.IsZero() || expected.ExpiresAt.IsZero() || !expected.IssuedAt.Before(expected.ExpiresAt) {
		return Activation{}, fmt.Errorf("strategy activation: invalid validity window")
	}
	now = now.UTC()
	if now.Before(expected.IssuedAt.UTC()) || !now.Before(expected.ExpiresAt.UTC()) || !now.Before(expected.AttestationExpiresAt.UTC()) {
		return Activation{}, fmt.Errorf("strategy activation: expired or not yet valid")
	}
	if !allApprovals(expected) || expected.ProtectionState != "WIRED" {
		return Activation{}, fmt.Errorf("strategy activation: approvals incomplete")
	}
	return Activation{digest: manifest.digest, generation: expected.Generation, valid: true}, nil
}

func manifestDigest(binding ManifestBinding) (string, error) {
	canonical, err := json.Marshal(binding)
	if err != nil {
		return "", fmt.Errorf("strategy activation: canonical manifest: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func sameBinding(a, b ManifestBinding) bool { return a == b }
func allApprovals(b ManifestBinding) bool {
	return b.LaneApproved && b.SchedulerApproved && b.AutoStartApproved && b.GateApproved && b.LiveApproved
}

type GateState struct {
	LaneDesired, LaneEffective, GateOpen, ProtectionWired bool
	KillSwitch                                            bool
}
type Guardian interface {
	Authorize(context.Context, EntryDecision) (string, error)
}
type AttemptStore interface {
	Plan(context.Context, PlannedAttempt) error
}
type OfficialGateway interface {
	Place(context.Context, PlannedAttempt) (string, error)
}

type PlannedAttempt struct {
	Decision                                          EntryDecision
	GuardianDecisionID, ManifestDigest, ClientOrderID string
}
type Dependencies struct {
	Manifest *ManifestRepository
	Guardian Guardian
	Attempts AttemptStore
	Gateway  OfficialGateway
	Now      func() time.Time
}

// Dispatch is a dormant orchestrator. With the only production-constructible
// repository (empty) it stops before Guardian, journal, or broker. Tests exercise
// the order with a package-private manifest fixture; no runtime wires it.
func Dispatch(ctx context.Context, decision EntryDecision, binding ManifestBinding, gates GateState, deps Dependencies) error {
	if !decision.Accepted {
		return fmt.Errorf("strategy dispatch: decision refused: %s", decision.Reason)
	}
	switch {
	case !gates.LaneDesired || !gates.LaneEffective:
		return fmt.Errorf("strategy dispatch: %s", RefusalLaneOff)
	case gates.KillSwitch:
		return fmt.Errorf("strategy dispatch: %s", RefusalKillSwitch)
	case !gates.ProtectionWired:
		return fmt.Errorf("strategy dispatch: %s", RefusalProtection)
	case !gates.GateOpen:
		return fmt.Errorf("strategy dispatch: %s", RefusalGate)
	}
	if deps.Manifest == nil || deps.Guardian == nil || deps.Attempts == nil || deps.Gateway == nil || deps.Now == nil {
		return fmt.Errorf("strategy dispatch: dependencies unavailable")
	}
	if _, err := deps.Manifest.Verify(binding, deps.Now()); err != nil {
		return fmt.Errorf("strategy dispatch: pre-guardian manifest: %w", err)
	}
	guardianID, err := deps.Guardian.Authorize(ctx, decision)
	if err != nil {
		return fmt.Errorf("strategy dispatch: guardian: %w", err)
	}
	activation, err := deps.Manifest.Verify(binding, deps.Now())
	if err != nil {
		return fmt.Errorf("strategy dispatch: submit-time manifest: %w", err)
	}
	attempt := PlannedAttempt{Decision: decision, GuardianDecisionID: guardianID, ManifestDigest: activation.Digest(), ClientOrderID: decision.Identity}
	if err := deps.Attempts.Plan(ctx, attempt); err != nil {
		return fmt.Errorf("strategy dispatch: durable plan: %w", err)
	}
	if _, err := deps.Gateway.Place(ctx, attempt); err != nil {
		return fmt.Errorf("strategy dispatch: official gateway: %w", err)
	}
	return nil
}
