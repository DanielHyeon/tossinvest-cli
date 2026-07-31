package attest

// This file defines a second, deliberately separate attestation contract for
// broker-resident protection. The existing capability-attestation.json remains
// format version 1 and keeps its historical semantics. A protection matrix can
// therefore be deployed dormant without making an older execution attestation
// appear to claim conditional-order behavior it never measured.

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	ProtectionFileName      = "protection-capability-attestation.json"
	ProtectionFormatVersion = 1
	maxProtectionFileSize   = 1 << 20
)

var (
	ErrProtectionFile    = errors.New("attest: unsafe protection capability file")
	ErrProtectionInvalid = errors.New("attest: invalid protection capability matrix")
	ErrProtectionExpired = errors.New("attest: protection capability matrix expired")
	ErrProtectionScope   = errors.New("attest: protection capability does not cover the requested scope")
)

type ProtectionTool string

const (
	ToolVerifyExecutionCapability ProtectionTool = "verify-execution-capability"
	ToolVerifyObservesTrigger     ProtectionTool = "verify-observes-the-trigger"
)

type CapabilityMarket string

const (
	MarketKR CapabilityMarket = "KR"
	MarketUS CapabilityMarket = "US"
)

type CapabilitySession string

const (
	SessionRegular  CapabilitySession = "REGULAR"
	SessionExtended CapabilitySession = "EXTENDED"
)

type CapabilityConditionalType string

const (
	ConditionalSingle CapabilityConditionalType = "SINGLE"
	ConditionalOCO    CapabilityConditionalType = "OCO"
)

type CapabilityOrderType string

const (
	OrderMarket CapabilityOrderType = "MARKET"
	OrderLimit  CapabilityOrderType = "LIMIT"
)

type CapabilityTriggerSource string

const (
	TriggerLastTrade CapabilityTriggerSource = "LAST_TRADE"
	TriggerMarkPrice CapabilityTriggerSource = "MARK_PRICE"
)

type CapabilityTriggerDirection string

const TriggerFallsToOrBelow CapabilityTriggerDirection = "FALLS_TO_OR_BELOW"

type ReplaceMode string

const (
	ReplaceAtomic     ReplaceMode = "ATOMIC"
	ReplaceContinuous ReplaceMode = "CONTINUOUS"
)

type ToolBuild struct {
	Version string `json:"version"`
	Build   string `json:"build"`
}

type ProtectionEvidence struct {
	Tool    ProtectionTool `json:"tool"`
	Version string         `json:"version"`
	Build   string         `json:"build"`
	Digest  string         `json:"evidence_digest"`
}

type TriggerCapability struct {
	Source    CapabilityTriggerSource    `json:"source"`
	Direction CapabilityTriggerDirection `json:"direction"`
}

type QuantityCapability struct {
	Minimum     int64 `json:"minimum"`
	Maximum     int64 `json:"maximum"`
	PartialFill bool  `json:"partial_fill"`
}

type PersistenceCapability struct {
	SurvivesProcessExit bool `json:"survives_process_exit"`
	SurvivesRestart     bool `json:"survives_restart"`
}

type ReservationCapability struct {
	ReservesSellableQuantity bool `json:"reserves_sellable_quantity"`
}

type IdempotencyCapability struct {
	Create        bool `json:"create"`
	ClientOrderID bool `json:"client_order_id"`
}

type ReplaceCapability struct {
	Mode                  ReplaceMode `json:"mode"`
	ContinuousCoverage    bool        `json:"continuous_coverage"`
	NewIdentifierRecorded bool        `json:"new_identifier_recorded"`
}

type ConditionalCapability struct {
	AccountRef      string                    `json:"account_ref"`
	Profile         string                    `json:"profile"`
	Market          CapabilityMarket          `json:"market"`
	Session         CapabilitySession         `json:"session"`
	ConditionalType CapabilityConditionalType `json:"conditional_type"`
	OrderType       CapabilityOrderType       `json:"order_type"`
	Trigger         TriggerCapability         `json:"trigger"`
	Quantity        QuantityCapability        `json:"quantity"`
	Persistence     PersistenceCapability     `json:"persistence"`
	Reservation     ReservationCapability     `json:"reservation"`
	Idempotency     IdempotencyCapability     `json:"idempotency"`
	Replace         ReplaceCapability         `json:"replace"`
}

type ProtectionCapabilityMatrix struct {
	FormatVersion int                     `json:"format_version"`
	IssuedAt      time.Time               `json:"issued_at"`
	ExpiresAt     time.Time               `json:"expires_at"`
	Evidence      []ProtectionEvidence    `json:"evidence"`
	Capabilities  []ConditionalCapability `json:"capabilities"`
}

// ProtectionScope is the exact runtime identity a future wiring change would
// ask the dormant matrix to cover. Nothing in this change calls it from engine
// startup or turns the result into ProtectionReady=WIRED.
type ProtectionScope struct {
	AccountRef      string
	Profile         string
	Market          CapabilityMarket
	Session         CapabilitySession
	ConditionalType CapabilityConditionalType
	OrderType       CapabilityOrderType
	TriggerSource   CapabilityTriggerSource
	Quantity        int64
	Tools           map[ProtectionTool]ToolBuild
}

type fileIdentity struct{ UID uint32 }

// LoadProtectionCapability checks the file itself, strictly decodes the entire
// matrix, validates every claim, and finally requires one exact scope match.
func LoadProtectionCapability(path string, now time.Time, scope ProtectionScope) (ProtectionCapabilityMatrix, error) {
	uid := os.Geteuid()
	if uid < 0 {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: current owner cannot be determined", ErrProtectionFile)
	}
	return loadProtectionCapability(path, now, scope, fileIdentity{UID: uint32(uid)})
}

func loadProtectionCapability(path string, now time.Time, scope ProtectionScope, owner fileIdentity) (ProtectionCapabilityMatrix, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: inspecting %s: %v", ErrProtectionFile, path, err)
	}
	if err := checkProtectionFileInfo(before, owner); err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: %s: %v", ErrProtectionFile, path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: opening %s: %v", ErrProtectionFile, path, err)
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: stat opened %s: %v", ErrProtectionFile, path, err)
	}
	if !os.SameFile(before, after) {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: %s changed while it was opened", ErrProtectionFile, path)
	}
	if err := checkProtectionFileInfo(after, owner); err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: opened %s: %v", ErrProtectionFile, path, err)
	}

	data, err := io.ReadAll(io.LimitReader(f, maxProtectionFileSize+1))
	if err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: reading %s: %v", ErrProtectionFile, path, err)
	}
	if len(data) > maxProtectionFileSize {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: %s exceeds %d bytes", ErrProtectionFile, path, maxProtectionFileSize)
	}
	m, err := decodeProtectionMatrix(data)
	if err != nil {
		return ProtectionCapabilityMatrix{}, err
	}
	if err := m.validate(now); err != nil {
		return ProtectionCapabilityMatrix{}, err
	}
	if err := m.verifyScope(scope); err != nil {
		return ProtectionCapabilityMatrix{}, err
	}
	return m, nil
}

func checkProtectionFileInfo(info os.FileInfo, owner fileIdentity) error {
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("not a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		return fmt.Errorf("mode is %04o, want 0600", info.Mode().Perm())
	}
	uid, ok := fileOwnerUID(info)
	if !ok {
		return errors.New("owner cannot be determined")
	}
	if uid != owner.UID {
		return fmt.Errorf("owner uid is %d, want %d", uid, owner.UID)
	}
	return nil
}

// fileOwnerUID uses the FileInfo's platform stat object without baking a
// platform-specific syscall.Stat_t into this otherwise portable parser.
func fileOwnerUID(info os.FileInfo) (uint32, bool) {
	v := reflect.ValueOf(info.Sys())
	if !v.IsValid() {
		return 0, false
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}
	f := v.FieldByName("Uid")
	if !f.IsValid() || !f.CanUint() {
		return 0, false
	}
	return uint32(f.Uint()), true
}

func decodeProtectionMatrix(data []byte) (ProtectionCapabilityMatrix, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var m ProtectionCapabilityMatrix
	if err := dec.Decode(&m); err != nil {
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: %v", ErrProtectionInvalid, err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing JSON value")
		}
		return ProtectionCapabilityMatrix{}, fmt.Errorf("%w: %v", ErrProtectionInvalid, err)
	}
	return m, nil
}

func (m ProtectionCapabilityMatrix) validate(now time.Time) error {
	if m.FormatVersion != ProtectionFormatVersion {
		return fmt.Errorf("%w: format_version is %d, require exactly %d", ErrProtectionInvalid, m.FormatVersion, ProtectionFormatVersion)
	}
	if m.IssuedAt.IsZero() || m.ExpiresAt.IsZero() || m.IssuedAt.After(now) || !m.ExpiresAt.After(m.IssuedAt) {
		return fmt.Errorf("%w: invalid issued/expiry window", ErrProtectionInvalid)
	}
	if !now.Before(m.ExpiresAt) {
		return fmt.Errorf("%w: expired at %s", ErrProtectionExpired, m.ExpiresAt.UTC().Format(time.RFC3339))
	}
	if err := validateProtectionEvidence(m.Evidence); err != nil {
		return err
	}
	if len(m.Capabilities) == 0 {
		return fmt.Errorf("%w: no capability rows", ErrProtectionInvalid)
	}
	seen := map[string]bool{}
	for i, c := range m.Capabilities {
		if err := c.validate(); err != nil {
			return fmt.Errorf("%w: capability[%d]: %v", ErrProtectionInvalid, i, err)
		}
		key := strings.Join([]string{accountDigits(c.AccountRef), strings.TrimSpace(c.Profile), string(c.Market), string(c.Session), string(c.ConditionalType), string(c.OrderType), string(c.Trigger.Source)}, "\x00")
		if seen[key] {
			return fmt.Errorf("%w: duplicate capability row %d", ErrProtectionInvalid, i)
		}
		seen[key] = true
	}
	return nil
}

func validateProtectionEvidence(evidence []ProtectionEvidence) error {
	want := map[ProtectionTool]bool{ToolVerifyExecutionCapability: true, ToolVerifyObservesTrigger: true}
	seen := map[ProtectionTool]bool{}
	for i, e := range evidence {
		if !want[e.Tool] || seen[e.Tool] {
			return fmt.Errorf("%w: evidence[%d] has unknown or duplicate tool %q", ErrProtectionInvalid, i, e.Tool)
		}
		if strings.TrimSpace(e.Version) == "" || !validSHA256(e.Build) || !validSHA256(e.Digest) {
			return fmt.Errorf("%w: evidence[%d] lacks versioned build/digest", ErrProtectionInvalid, i)
		}
		seen[e.Tool] = true
	}
	for tool := range want {
		if !seen[tool] {
			return fmt.Errorf("%w: missing evidence from %s", ErrProtectionInvalid, tool)
		}
	}
	return nil
}

func validSHA256(s string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(s, prefix) || len(s) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(s, prefix))
	return err == nil
}

func (c ConditionalCapability) validate() error {
	switch {
	case accountDigits(c.AccountRef) == "":
		return errors.New("account_ref is empty")
	case strings.TrimSpace(c.Profile) == "":
		return errors.New("profile is empty")
	case c.Market != MarketKR && c.Market != MarketUS:
		return fmt.Errorf("market %q is unknown", c.Market)
	case c.Session != SessionRegular:
		return fmt.Errorf("session %q is not attested for dormant SINGLE+MARKET protection", c.Session)
	case c.ConditionalType != ConditionalSingle:
		return fmt.Errorf("conditional_type %q is not SINGLE", c.ConditionalType)
	case c.OrderType != OrderMarket:
		return fmt.Errorf("order_type %q is not MARKET", c.OrderType)
	case c.Trigger.Source != TriggerLastTrade:
		return fmt.Errorf("trigger source %q is not LAST_TRADE", c.Trigger.Source)
	case c.Trigger.Direction != TriggerFallsToOrBelow:
		return fmt.Errorf("trigger direction %q is unknown", c.Trigger.Direction)
	case c.Quantity.Minimum < 1 || c.Quantity.Maximum < c.Quantity.Minimum || !c.Quantity.PartialFill:
		return errors.New("quantity range/partial-fill claim is unsafe")
	case !c.Persistence.SurvivesProcessExit || !c.Persistence.SurvivesRestart:
		return errors.New("persistence claim is incomplete")
	case !c.Reservation.ReservesSellableQuantity:
		return errors.New("sellable reservation claim is absent")
	case !c.Idempotency.Create || !c.Idempotency.ClientOrderID:
		return errors.New("idempotency claim is incomplete")
	case c.Replace.Mode != ReplaceAtomic && c.Replace.Mode != ReplaceContinuous:
		return fmt.Errorf("replace mode %q is unknown", c.Replace.Mode)
	case !c.Replace.ContinuousCoverage || !c.Replace.NewIdentifierRecorded:
		return errors.New("replace continuity/identity claim is incomplete")
	}
	return nil
}

func (m ProtectionCapabilityMatrix) verifyScope(scope ProtectionScope) error {
	var matched *ConditionalCapability
	for i := range m.Capabilities {
		c := &m.Capabilities[i]
		if sameAccount(c.AccountRef, scope.AccountRef) && c.Profile == strings.TrimSpace(scope.Profile) && c.Market == scope.Market && c.Session == scope.Session && c.ConditionalType == scope.ConditionalType && c.OrderType == scope.OrderType && c.Trigger.Source == scope.TriggerSource && scope.Quantity >= c.Quantity.Minimum && scope.Quantity <= c.Quantity.Maximum {
			matched = c
			break
		}
	}
	if matched == nil {
		return fmt.Errorf("%w: account/profile/market/session/type/trigger/quantity did not match", ErrProtectionScope)
	}
	if len(scope.Tools) != 2 {
		return fmt.Errorf("%w: both verifier tool builds are required", ErrProtectionScope)
	}
	for _, e := range m.Evidence {
		want, ok := scope.Tools[e.Tool]
		if !ok || want.Version != e.Version || want.Build != e.Build {
			return fmt.Errorf("%w: %s tool/build mismatch", ErrProtectionScope, e.Tool)
		}
	}
	return nil
}
