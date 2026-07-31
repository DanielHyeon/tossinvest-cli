package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
)

const (
	SchedulerVersion = "scheduler-v1"
	DesiredFileName  = "scheduler-desired.json"
)

type MarketScope string

const (
	MarketScopeNone MarketScope = "none"
	MarketScopeKR   MarketScope = "KR"
	MarketScopeUS   MarketScope = "US"
)

type SessionScope string

const SessionRegular SessionScope = "regular"

type DesiredState struct {
	Revision        uint64       `json:"revision"`
	Version         string       `json:"version"`
	Enabled         bool         `json:"enabled"`
	AutoStart       bool         `json:"autoStart"`
	Market          MarketScope  `json:"market"`
	Session         SessionScope `json:"session"`
	Actor           string       `json:"actor,omitempty"`
	ApprovedAt      time.Time    `json:"approvedAt,omitempty"`
	CalendarVersion string       `json:"calendarVersion,omitempty"`
	ConfigVersion   string       `json:"configVersion,omitempty"`
}

func DefaultDesiredState() DesiredState {
	return DesiredState{Version: SchedulerVersion, Market: MarketScopeNone, Session: SessionRegular}
}

func (d DesiredState) validate() error {
	if d.Version != SchedulerVersion {
		return fmt.Errorf("scheduler version %q is unsupported", d.Version)
	}
	if d.Session != SessionRegular {
		return fmt.Errorf("session %q is unsupported", d.Session)
	}
	switch d.Market {
	case MarketScopeNone, MarketScopeKR, MarketScopeUS:
	default:
		return fmt.Errorf("market scope %q is unsupported", d.Market)
	}
	if d.Enabled || d.AutoStart {
		if !d.Enabled {
			return errors.New("auto-start requires scheduler enabled")
		}
		if d.Actor == "" || d.ApprovedAt.IsZero() || d.Market == MarketScopeNone || d.CalendarVersion == "" || d.ConfigVersion == "" {
			return errors.New("enabled scheduler requires actor, approval time, market, calendar version, and config version")
		}
	}
	return nil
}

func (d DesiredState) validateAt(now time.Time) error {
	if err := d.validate(); err != nil {
		return err
	}
	if !d.ApprovedAt.IsZero() && (now.IsZero() || d.ApprovedAt.After(now)) {
		return errors.New("scheduler approval time is in the future")
	}
	return nil
}

type DesiredStore struct{ path string }

func NewDesiredStore(path string) *DesiredStore { return &DesiredStore{path: path} }

func (s *DesiredStore) Load(ctx context.Context) (DesiredState, error) {
	if err := ctx.Err(); err != nil {
		return DesiredState{}, err
	}
	return s.loadAt(time.Now().UTC())
}

func (s *DesiredStore) loadAt(now time.Time) (DesiredState, error) {
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultDesiredState(), nil
	}
	if err != nil {
		return DesiredState{}, err
	}
	if err := rejectDuplicateDesiredKeys(raw); err != nil {
		return DesiredState{}, fmt.Errorf("decode scheduler desired state: %w", err)
	}
	var out DesiredState
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		return DesiredState{}, fmt.Errorf("decode scheduler desired state: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return DesiredState{}, errors.New("decode scheduler desired state: trailing JSON value")
	}
	if err := out.validateAt(now); err != nil {
		return DesiredState{}, err
	}
	return out, nil
}

func (s *DesiredStore) Save(ctx context.Context, desired DesiredState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := desired.validateAt(time.Now().UTC()); err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	lock, err := acquireDesiredLock(ctx, s.path+".lock")
	if err != nil {
		return err
	}
	defer lock.release()
	if err := ctx.Err(); err != nil {
		return err
	}
	current, err := s.loadAt(time.Now().UTC())
	if err != nil {
		return err
	}
	if current.Revision != desired.Revision {
		return fmt.Errorf("%w: expected %d, current %d", ErrDesiredRevisionConflict, desired.Revision, current.Revision)
	}
	if desired.Revision == math.MaxUint64 {
		return errors.New("scheduler desired revision exhausted")
	}
	desired.Revision++
	raw, err := json.MarshalIndent(desired, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".scheduler-desired-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

type ActivationBinding struct {
	SchedulerVersion string       `json:"schedulerVersion"`
	DesiredRevision  uint64       `json:"desiredRevision"`
	CalendarVersion  string       `json:"calendarVersion"`
	Market           MarketScope  `json:"market"`
	Session          SessionScope `json:"session"`
	ConfigVersion    string       `json:"configVersion"`
	BuildDigest      string       `json:"buildDigest"`
	Actor            string       `json:"actor"`
	ApprovedAt       time.Time    `json:"approvedAt"`
}

func (d DesiredState) ActivationBinding(buildDigest string) ActivationBinding {
	return ActivationBinding{
		SchedulerVersion: d.Version, DesiredRevision: d.Revision, CalendarVersion: d.CalendarVersion, Market: d.Market,
		Session: d.Session, ConfigVersion: d.ConfigVersion, BuildDigest: buildDigest,
		Actor: d.Actor, ApprovedAt: d.ApprovedAt,
	}
}

type CurrentBinding struct {
	SchedulerVersion string
	CalendarVersion  string
	Market           MarketScope
	Session          SessionScope
	ConfigVersion    string
	BuildDigest      string
}

type ActivationVerifier interface {
	verifyActivation(context.Context, ActivationBinding, time.Time) error
}

var (
	ErrManifestUnavailable     = errors.New("activation manifest verifier unavailable")
	ErrManifestMismatch        = errors.New("activation manifest mismatch")
	ErrManifestExpired         = errors.New("activation manifest expired")
	ErrManifestRevoked         = errors.New("activation manifest revoked")
	ErrDesiredRevisionConflict = errors.New("scheduler desired revision conflict")
)

// Activation is an opaque capability issued only after an exact manifest
// verification. Callers cannot forge it with a bool or public struct literal.
type Activation struct{ binding ActivationBinding }

type ResumeReason string

const (
	ResumeExactManifest       ResumeReason = "EXACT_MANIFEST"
	ResumeDesiredMismatch     ResumeReason = "DESIRED_MISMATCH"
	ResumeManifestMismatch    ResumeReason = "MANIFEST_MISMATCH"
	ResumeManifestExpired     ResumeReason = "MANIFEST_EXPIRED"
	ResumeManifestRevoked     ResumeReason = "MANIFEST_REVOKED"
	ResumeManifestUnavailable ResumeReason = "MANIFEST_UNAVAILABLE"
	ResumeAutoStartOff        ResumeReason = "AUTO_START_OFF"
	ResumeVerificationFailed  ResumeReason = "VERIFICATION_FAILED"
)

type RestoreResult struct {
	Restored   bool
	Reason     ResumeReason
	Err        error
	Activation *Activation
}

// Restore is the explicit a047 integration seam. A nil verifier is the current
// production posture and always refuses restoration.
func Restore(ctx context.Context, desired DesiredState, current CurrentBinding, verifier ActivationVerifier, now time.Time) RestoreResult {
	if err := ctx.Err(); err != nil {
		return RestoreResult{Reason: ResumeVerificationFailed, Err: err}
	}
	if !desired.AutoStart {
		return RestoreResult{Reason: ResumeAutoStartOff}
	}
	if err := desired.validate(); err != nil {
		return RestoreResult{Reason: ResumeDesiredMismatch, Err: err}
	}
	if now.IsZero() || now.Before(desired.ApprovedAt) {
		return RestoreResult{Reason: ResumeDesiredMismatch, Err: errors.New("approval time is in the future")}
	}
	if current.SchedulerVersion != desired.Version || current.CalendarVersion != desired.CalendarVersion ||
		current.Market != desired.Market || current.Session != desired.Session || current.ConfigVersion != desired.ConfigVersion || current.BuildDigest == "" {
		return RestoreResult{Reason: ResumeDesiredMismatch}
	}
	if verifier == nil {
		return RestoreResult{Reason: ResumeManifestUnavailable, Err: ErrManifestUnavailable}
	}
	binding := desired.ActivationBinding(current.BuildDigest)
	err := verifier.verifyActivation(ctx, binding, now)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return RestoreResult{Reason: ResumeVerificationFailed, Err: ctxErr}
	}
	if err == nil {
		return RestoreResult{Restored: true, Reason: ResumeExactManifest, Activation: &Activation{binding: binding}}
	}
	switch {
	case errors.Is(err, ErrManifestExpired):
		return RestoreResult{Reason: ResumeManifestExpired, Err: err}
	case errors.Is(err, ErrManifestRevoked):
		return RestoreResult{Reason: ResumeManifestRevoked, Err: err}
	default:
		return RestoreResult{Reason: ResumeManifestMismatch, Err: err}
	}
}

func rejectDuplicateDesiredKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("desired state must be a JSON object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("desired state contains a non-string key")
		}
		if seen[key] {
			return fmt.Errorf("duplicate field %q", key)
		}
		seen[key] = true
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
