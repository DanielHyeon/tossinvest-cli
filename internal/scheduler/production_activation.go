package scheduler

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const (
	productionActivationSchema       = "strategy-activation-manifest:v1"
	productionActivationDomain       = "TossOS/strategy-scheduler-activation/ed25519/v1"
	productionActivationAlgorithm    = "Ed25519"
	productionActivationMaximumBytes = 64 << 10
	productionActivationMaximumLife  = 24 * time.Hour
)

var ErrPairedActivationInvalid = errors.New("scheduler: paired production activation requires exact KR and US requests")

// ProductionActivationConfig contains only the immutable location and trust
// pins used to consume an activation created by a separate human process. It
// has no desired-state writer, signing key, toggle, journal or broker handle.
type ProductionActivationConfig struct {
	ConfigDir      string
	Market         MarketScope
	ManifestDigest string
	TrustedKeyID   string
	TrustedKey     ed25519.PublicKey
}

type productionActivationVerifier struct {
	config ProductionActivationConfig
	path   string
	owner  uint64
}

type productionActivationManifestBody struct {
	SchemaVersion      string       `json:"schema_version"`
	Domain             string       `json:"domain"`
	SignatureAlgorithm string       `json:"signature_algorithm"`
	KeyID              string       `json:"key_id"`
	Generation         uint64       `json:"generation"`
	SchedulerVersion   string       `json:"scheduler_version"`
	DesiredRevision    uint64       `json:"desired_revision"`
	CalendarVersion    string       `json:"calendar_version"`
	Market             MarketScope  `json:"market"`
	Session            SessionScope `json:"session"`
	ConfigVersion      string       `json:"config_version"`
	BuildDigest        string       `json:"build_digest"`
	Actor              string       `json:"actor"`
	ApprovedAt         string       `json:"approved_at"`
	IssuedAt           string       `json:"issued_at"`
	ExpiresAt          string       `json:"expires_at"`
	Revoked            bool         `json:"revoked"`
}

type productionActivationManifest struct {
	productionActivationManifestBody
	Signature string `json:"signature"`
}

// ProductionActivationFileName is a closed mapping. It never accepts a path
// component from the manifest or caller.
func ProductionActivationFileName(market MarketScope) string {
	switch market {
	case MarketScopeKR:
		return "strategy-activation-KR.json"
	case MarketScopeUS:
		return "strategy-activation-US.json"
	default:
		return ""
	}
}

// ProductionDesiredFileName separates the two human-controlled desired states.
// The released single-market DesiredFileName remains unchanged for its existing
// console workflow.
func ProductionDesiredFileName(market MarketScope) string {
	switch market {
	case MarketScopeKR:
		return "scheduler-desired-KR.json"
	case MarketScopeUS:
		return "scheduler-desired-US.json"
	default:
		return ""
	}
}

// LoadAt is the production paired-loader boundary. It preserves Load's exact
// parser and validation while using the engine's one frozen clock observation.
func (s *DesiredStore) LoadAt(ctx context.Context, now time.Time) (DesiredState, error) {
	if s == nil || ctx == nil || now.IsZero() {
		return DesiredState{}, errors.New("scheduler: desired-state load time is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return DesiredState{}, err
	}
	return s.loadAt(now.UTC())
}

// NewProductionActivationVerifier creates a read-only verifier. Returning the
// package-private interface keeps Activation construction inside scheduler.
func NewProductionActivationVerifier(config ProductionActivationConfig) (ActivationVerifier, error) {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	name := ProductionActivationFileName(config.Market)
	owner, ok := productionActivationOwnerUID()
	if !ok || name == "" || !filepath.IsAbs(config.ConfigDir) || !canonicalActivationDigest(config.ManifestDigest) ||
		!boundedActivationIdentity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize {
		return nil, ErrManifestUnavailable
	}
	return &productionActivationVerifier{config: config, path: filepath.Join(config.ConfigDir, name), owner: owner}, nil
}

func (verifier *productionActivationVerifier) verifyActivation(ctx context.Context, binding ActivationBinding, now time.Time) (ActivationEvidence, error) {
	if verifier == nil || ctx == nil || verifier.path == "" || now.IsZero() {
		return ActivationEvidence{}, ErrManifestMismatch
	}
	if err := ctx.Err(); err != nil {
		return ActivationEvidence{}, err
	}
	data, err := readProductionActivationFile(verifier.path, verifier.owner, 0o400, productionActivationMaximumBytes)
	if err != nil || activationDigest(data) != verifier.config.ManifestDigest {
		return ActivationEvidence{}, ErrManifestMismatch
	}
	manifest, err := decodeProductionActivationManifest(data)
	if err != nil {
		return ActivationEvidence{}, ErrManifestMismatch
	}
	canonical, err := json.Marshal(manifest.productionActivationManifestBody)
	if err != nil {
		return ActivationEvidence{}, ErrManifestMismatch
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	if err != nil || base64.StdEncoding.EncodeToString(signature) != manifest.Signature || len(signature) != ed25519.SignatureSize ||
		!ed25519.Verify(verifier.config.TrustedKey, canonical, signature) {
		return ActivationEvidence{}, ErrManifestMismatch
	}
	if manifest.Revoked {
		return ActivationEvidence{}, ErrManifestRevoked
	}
	if err := validateProductionActivationManifest(manifest.productionActivationManifestBody, verifier.config, binding, now.UTC()); err != nil {
		return ActivationEvidence{}, err
	}
	if err := ctx.Err(); err != nil {
		return ActivationEvidence{}, err
	}
	expires, _ := canonicalActivationTime(manifest.ExpiresAt)
	return ActivationEvidence{Generation: manifest.Generation, ExpiresAt: expires}, nil
}

func decodeProductionActivationManifest(data []byte) (productionActivationManifest, error) {
	if len(data) == 0 || len(data) > productionActivationMaximumBytes {
		return productionActivationManifest{}, ErrManifestMismatch
	}
	if err := rejectDuplicateDesiredKeys(data); err != nil {
		return productionActivationManifest{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionActivationManifest
	if err := decoder.Decode(&manifest); err != nil {
		return productionActivationManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionActivationManifest{}, errors.New("scheduler: activation manifest has trailing JSON")
	}
	return manifest, nil
}

func validateProductionActivationManifest(body productionActivationManifestBody, config ProductionActivationConfig, binding ActivationBinding, now time.Time) error {
	if body.SchemaVersion != productionActivationSchema || body.Domain != productionActivationDomain ||
		body.SignatureAlgorithm != productionActivationAlgorithm || body.KeyID != config.TrustedKeyID || body.Generation == 0 ||
		body.SchedulerVersion != SchedulerVersion || body.SchedulerVersion != binding.SchedulerVersion ||
		body.DesiredRevision != binding.DesiredRevision || body.CalendarVersion != binding.CalendarVersion ||
		body.Market != config.Market || body.Market != binding.Market || body.Session != SessionRegular || body.Session != binding.Session ||
		body.ConfigVersion != binding.ConfigVersion || body.BuildDigest != binding.BuildDigest || body.Actor != binding.Actor ||
		!boundedActivationIdentity(body.CalendarVersion) || !boundedActivationIdentity(body.ConfigVersion) ||
		!boundedActivationIdentity(body.BuildDigest) || !boundedActivationIdentity(body.Actor) {
		return ErrManifestMismatch
	}
	approved, okApproved := canonicalActivationTime(body.ApprovedAt)
	issued, okIssued := canonicalActivationTime(body.IssuedAt)
	expires, okExpires := canonicalActivationTime(body.ExpiresAt)
	if !okApproved || !okIssued || !okExpires || !approved.Equal(binding.ApprovedAt.UTC()) || issued.Before(approved) ||
		issued.After(now) || !issued.Before(expires) || expires.Sub(issued) > productionActivationMaximumLife {
		return ErrManifestMismatch
	}
	if !now.Before(expires) {
		return ErrManifestExpired
	}
	return nil
}

func canonicalActivationTime(raw string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || parsed.IsZero() || parsed.Location() != time.UTC || parsed.UTC().Format(time.RFC3339Nano) != raw {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func canonicalActivationDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func activationDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func boundedActivationIdentity(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n")
}

// PairedRestoreRequest is one market's read-only restore input. It contains no
// writer and is accepted only as part of an exact KR/US pair.
type PairedRestoreRequest struct {
	Market   MarketScope
	Desired  DesiredState
	Current  CurrentBinding
	Verifier ActivationVerifier
}

// PairedRestoreResult preserves two independent classifications.
type PairedRestoreResult struct {
	KR RestoreResult
	US RestoreResult
}

func (result PairedRestoreResult) For(market MarketScope) RestoreResult {
	switch market {
	case MarketScopeKR:
		return result.KR
	case MarketScopeUS:
		return result.US
	default:
		return RestoreResult{Reason: ResumeVerificationFailed, Err: ErrPairedActivationInvalid}
	}
}

// RestorePairedProduction freezes the clock once, then verifies KR and US in
// separate goroutines. It waits for both classifications but never cancels a
// peer merely because the other market refused.
func RestorePairedProduction(ctx context.Context, requests [2]PairedRestoreRequest, now func() time.Time) (PairedRestoreResult, error) {
	failed := pairedActivationFailure(ErrPairedActivationInvalid)
	if ctx == nil || now == nil || !exactPairedRestoreRequests(requests) {
		return failed, ErrPairedActivationInvalid
	}
	frozen := now()
	if frozen.IsZero() {
		return failed, ErrPairedActivationInvalid
	}
	frozen = frozen.UTC()
	type outcome struct {
		market MarketScope
		result RestoreResult
	}
	outcomes := make(chan outcome, 2)
	for _, request := range requests {
		request := request
		go func() {
			result := RestoreResult{Reason: ResumeVerificationFailed}
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						result.Err = fmt.Errorf("scheduler: %s activation verifier panic", request.Market)
					}
				}()
				result = Restore(ctx, request.Desired, request.Current, request.Verifier, frozen)
			}()
			outcomes <- outcome{market: request.Market, result: result}
		}()
	}
	result := PairedRestoreResult{}
	for index := 0; index < 2; index++ {
		outcome := <-outcomes
		if outcome.market == MarketScopeKR {
			result.KR = outcome.result
		} else {
			result.US = outcome.result
		}
	}
	return result, nil
}

func exactPairedRestoreRequests(requests [2]PairedRestoreRequest) bool {
	seen := map[MarketScope]bool{}
	for _, request := range requests {
		if (request.Market != MarketScopeKR && request.Market != MarketScopeUS) || seen[request.Market] ||
			(request.Desired.AutoStart && request.Desired.Market != request.Market) || request.Current.Market != request.Market {
			return false
		}
		seen[request.Market] = true
	}
	return seen[MarketScopeKR] && seen[MarketScopeUS]
}

func pairedActivationFailure(err error) PairedRestoreResult {
	value := RestoreResult{Reason: ResumeVerificationFailed, Err: err}
	return PairedRestoreResult{KR: value, US: value}
}
