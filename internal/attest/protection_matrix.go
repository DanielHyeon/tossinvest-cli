package attest

// This file defines a second, deliberately separate attestation contract for
// broker-resident protection. The existing capability-attestation.json remains
// format version 1 and keeps its historical semantics. A protection matrix can
// therefore be deployed dormant without making an older execution attestation
// appear to claim conditional-order behavior it never measured.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	ErrProtectionFile     = errors.New("attest: unsafe protection capability file")
	ErrProtectionInvalid  = errors.New("attest: invalid protection capability matrix")
	ErrProtectionExpired  = errors.New("attest: protection capability matrix expired")
	ErrProtectionScope    = errors.New("attest: protection capability does not cover the requested scope")
	ErrProtectionEvidence = errors.New("attest: protection evidence is absent or does not match")
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
	Tool             ProtectionTool `json:"tool"`
	Version          string         `json:"version"`
	Build            string         `json:"build"`
	Source           string         `json:"source"`
	Digest           string         `json:"evidence_digest"`
	CapabilityDigest string         `json:"capability_digest"`
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
	FormatVersion    int                     `json:"format_version"`
	IssuedAt         time.Time               `json:"issued_at"`
	ExpiresAt        time.Time               `json:"expires_at"`
	CapabilityDigest string                  `json:"capability_digest"`
	Evidence         []ProtectionEvidence    `json:"evidence"`
	Capabilities     []ConditionalCapability `json:"capabilities"`
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

// ParsedProtectionCapability has passed strict JSON and local file-integrity
// checks, but is deliberately not an authorization result. External evidence
// bytes and the requested runtime scope have not yet been verified.
type ParsedProtectionCapability struct{ matrix ProtectionCapabilityMatrix }

// VerifiedProtectionCapability is only constructed after external evidence
// bytes, time bounds, and exact runtime scope have all been checked.
type VerifiedProtectionCapability struct{ matrix ProtectionCapabilityMatrix }

func (v VerifiedProtectionCapability) Matrix() ProtectionCapabilityMatrix {
	m := v.matrix
	m.Evidence = append([]ProtectionEvidence(nil), m.Evidence...)
	m.Capabilities = append([]ConditionalCapability(nil), m.Capabilities...)
	return m
}

func ParseProtectionCapability(path string) (ParsedProtectionCapability, error) {
	uid := os.Geteuid()
	if uid < 0 {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: current owner cannot be determined", ErrProtectionFile)
	}
	return parseProtectionCapability(path, fileIdentity{UID: uint32(uid)})
}

func parseProtectionCapability(path string, owner fileIdentity) (ParsedProtectionCapability, error) {
	if filepath.Base(path) != ProtectionFileName {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: basename must be %s", ErrProtectionFile, ProtectionFileName)
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: inspecting parent %s: %v", ErrProtectionFile, parent, err)
	}
	if err := checkProtectionParentInfo(parentInfo, owner); err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: parent %s: %v", ErrProtectionFile, parent, err)
	}
	before, err := os.Lstat(path)
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: inspecting %s: %v", ErrProtectionFile, path, err)
	}
	if err := checkProtectionFileInfo(before, owner); err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: %s: %v", ErrProtectionFile, path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: opening %s: %v", ErrProtectionFile, path, err)
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: stat opened %s: %v", ErrProtectionFile, path, err)
	}
	if !os.SameFile(before, after) {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: %s changed while it was opened", ErrProtectionFile, path)
	}
	if err := checkProtectionFileInfo(after, owner); err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: opened %s: %v", ErrProtectionFile, path, err)
	}

	data, err := io.ReadAll(io.LimitReader(f, maxProtectionFileSize+1))
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: reading %s: %v", ErrProtectionFile, path, err)
	}
	if len(data) > maxProtectionFileSize {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: %s exceeds %d bytes", ErrProtectionFile, path, maxProtectionFileSize)
	}
	openedAfterRead, err := f.Stat()
	if err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: restat opened %s: %v", ErrProtectionFile, path, err)
	}
	pathAfterRead, err := os.Lstat(path)
	if err != nil || !sameProtectionSnapshot(after, openedAfterRead) || !sameProtectionSnapshot(after, pathAfterRead) {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: %s changed while it was read", ErrProtectionFile, path)
	}
	if err := checkProtectionFileInfo(openedAfterRead, owner); err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: restat opened %s: %v", ErrProtectionFile, path, err)
	}
	if err := checkProtectionFileInfo(pathAfterRead, owner); err != nil {
		return ParsedProtectionCapability{}, fmt.Errorf("%w: restat %s: %v", ErrProtectionFile, path, err)
	}
	m, err := decodeProtectionMatrix(data)
	if err != nil {
		return ParsedProtectionCapability{}, err
	}
	if err := m.validate(); err != nil {
		return ParsedProtectionCapability{}, err
	}
	return ParsedProtectionCapability{matrix: m}, nil
}

func sameProtectionSnapshot(before, after os.FileInfo) bool {
	return os.SameFile(before, after) && before.Size() == after.Size() && before.ModTime().Equal(after.ModTime()) && before.Mode() == after.Mode()
}

// VerifyProtectionCapability requires independent evidence bytes. A parsed
// matrix alone can never be promoted to a verified capability.
func VerifyProtectionCapability(parsed ParsedProtectionCapability, now time.Time, scope ProtectionScope, evidence map[string][]byte) (VerifiedProtectionCapability, error) {
	m := parsed.matrix
	if m.FormatVersion == 0 {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: empty parsed capability", ErrProtectionInvalid)
	}
	if m.IssuedAt.After(now) {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: issued_at is in the future", ErrProtectionInvalid)
	}
	if !now.Before(m.ExpiresAt) {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: expired at %s", ErrProtectionExpired, m.ExpiresAt.UTC().Format(time.RFC3339))
	}
	if err := m.verifyScope(scope); err != nil {
		return VerifiedProtectionCapability{}, err
	}
	if err := verifyProtectionEvidenceBytes(m.Evidence, evidence); err != nil {
		return VerifiedProtectionCapability{}, err
	}
	return VerifiedProtectionCapability{matrix: m}, nil
}

// LoadProtectionCapability is a convenience wrapper which preserves the hard
// Parse/Verify boundary by requiring the caller to provide evidence bytes.
func LoadProtectionCapability(path string, now time.Time, scope ProtectionScope, evidence map[string][]byte) (VerifiedProtectionCapability, error) {
	parsed, err := ParseProtectionCapability(path)
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	return VerifyProtectionCapability(parsed, now, scope, evidence)
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
	links, ok := fileLinkCount(info)
	if !ok || links != 1 {
		return fmt.Errorf("hard-link count is %d, want 1", links)
	}
	return nil
}

func checkProtectionParentInfo(info os.FileInfo, owner fileIdentity) error {
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("not a real directory")
	}
	if info.Mode().Perm() != 0o700 {
		return fmt.Errorf("mode is %04o, want 0700", info.Mode().Perm())
	}
	uid, ok := fileOwnerUID(info)
	if !ok || uid != owner.UID {
		return fmt.Errorf("owner uid is %d, want %d", uid, owner.UID)
	}
	return nil
}

func fileLinkCount(info os.FileInfo) (uint64, bool) {
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
	f := v.FieldByName("Nlink")
	if !f.IsValid() || !f.CanUint() {
		return 0, false
	}
	return f.Uint(), true
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

func (m ProtectionCapabilityMatrix) validate() error {
	if m.FormatVersion != ProtectionFormatVersion {
		return fmt.Errorf("%w: format_version is %d, require exactly %d", ErrProtectionInvalid, m.FormatVersion, ProtectionFormatVersion)
	}
	if m.IssuedAt.IsZero() || m.ExpiresAt.IsZero() || !m.ExpiresAt.After(m.IssuedAt) {
		return fmt.Errorf("%w: invalid issued/expiry window", ErrProtectionInvalid)
	}
	if err := validateProtectionEvidence(m.Evidence, m.CapabilityDigest); err != nil {
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
		account, _ := canonicalProtectionAccount(c.AccountRef)
		key := strings.Join([]string{account, c.Profile, string(c.Market), string(c.Session), string(c.ConditionalType), string(c.OrderType), string(c.Trigger.Source)}, "\x00")
		if seen[key] {
			return fmt.Errorf("%w: duplicate capability row %d", ErrProtectionInvalid, i)
		}
		seen[key] = true
	}
	canonical, err := canonicalProtectionMatrix(m)
	if err != nil {
		return fmt.Errorf("%w: canonicalizing capability matrix: %v", ErrProtectionInvalid, err)
	}
	if m.CapabilityDigest != protectionDigest(canonical) {
		return fmt.Errorf("%w: capability_digest does not bind canonical matrix", ErrProtectionInvalid)
	}
	return nil
}

func canonicalProtectionMatrix(m ProtectionCapabilityMatrix) ([]byte, error) {
	type canonicalEvidence struct {
		Tool    ProtectionTool `json:"tool"`
		Version string         `json:"version"`
		Build   string         `json:"build"`
		Source  string         `json:"source"`
		Digest  string         `json:"evidence_digest"`
	}
	type canonicalMatrix struct {
		FormatVersion int                     `json:"format_version"`
		IssuedAt      time.Time               `json:"issued_at"`
		ExpiresAt     time.Time               `json:"expires_at"`
		Evidence      []canonicalEvidence     `json:"evidence"`
		Capabilities  []ConditionalCapability `json:"capabilities"`
	}
	evidence := make([]canonicalEvidence, len(m.Evidence))
	for i, descriptor := range m.Evidence {
		evidence[i] = canonicalEvidence{Tool: descriptor.Tool, Version: descriptor.Version, Build: descriptor.Build, Source: descriptor.Source, Digest: descriptor.Digest}
	}
	return json.Marshal(canonicalMatrix{FormatVersion: m.FormatVersion, IssuedAt: m.IssuedAt, ExpiresAt: m.ExpiresAt, Evidence: evidence, Capabilities: m.Capabilities})
}

func validateProtectionEvidence(evidence []ProtectionEvidence, capabilityDigest string) error {
	want := map[ProtectionTool]bool{ToolVerifyExecutionCapability: true, ToolVerifyObservesTrigger: true}
	seen := map[ProtectionTool]bool{}
	for i, e := range evidence {
		if !want[e.Tool] || seen[e.Tool] {
			return fmt.Errorf("%w: evidence[%d] has unknown or duplicate tool %q", ErrProtectionInvalid, i, e.Tool)
		}
		if e.Version == "" || e.Version != strings.TrimSpace(e.Version) || !validSHA256(e.Build) || !validSHA256(e.Digest) || e.CapabilityDigest != capabilityDigest {
			return fmt.Errorf("%w: evidence[%d] lacks versioned build/digest", ErrProtectionInvalid, i)
		}
		if e.Source != expectedProtectionEvidenceSource(e.Tool) || filepath.Base(e.Source) != e.Source {
			return fmt.Errorf("%w: evidence[%d] source %q is not the exact tool evidence basename", ErrProtectionInvalid, i, e.Source)
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

func expectedProtectionEvidenceSource(tool ProtectionTool) string {
	switch tool {
	case ToolVerifyExecutionCapability:
		return "verify-execution-capability.evidence.jsonl"
	case ToolVerifyObservesTrigger:
		return "verify-observes-the-trigger.evidence.jsonl"
	default:
		return ""
	}
}

func verifyProtectionEvidenceBytes(descriptors []ProtectionEvidence, evidence map[string][]byte) error {
	if len(evidence) != len(descriptors) {
		return fmt.Errorf("%w: got %d evidence blobs, want %d", ErrProtectionEvidence, len(evidence), len(descriptors))
	}
	for _, descriptor := range descriptors {
		data, ok := evidence[descriptor.Source]
		if !ok || protectionDigest(data) != descriptor.Digest {
			return fmt.Errorf("%w: %s digest mismatch", ErrProtectionEvidence, descriptor.Source)
		}
	}
	return nil
}

func protectionDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
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
	account, accountErr := canonicalProtectionAccount(c.AccountRef)
	switch {
	case accountErr != nil || account == "":
		return errors.New("account_ref has invalid protection format")
	case c.Profile == "" || c.Profile != strings.TrimSpace(c.Profile):
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
	scopeAccount, err := canonicalProtectionAccount(scope.AccountRef)
	if err != nil || scope.Profile == "" || scope.Profile != strings.TrimSpace(scope.Profile) {
		return fmt.Errorf("%w: malformed runtime account/profile", ErrProtectionScope)
	}
	var matched *ConditionalCapability
	for i := range m.Capabilities {
		c := &m.Capabilities[i]
		account, _ := canonicalProtectionAccount(c.AccountRef)
		if account == scopeAccount && c.Profile == scope.Profile && c.Market == scope.Market && c.Session == scope.Session && c.ConditionalType == scope.ConditionalType && c.OrderType == scope.OrderType && c.Trigger.Source == scope.TriggerSource && scope.Quantity >= c.Quantity.Minimum && scope.Quantity <= c.Quantity.Maximum {
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

// canonicalProtectionAccount accepts only the deliberately narrow protection
// account grammar: 8-14 digits, optionally grouped by single hyphens. It does
// not inherit the legacy attestation parser's arbitrary-character removal.
func canonicalProtectionAccount(ref string) (string, error) {
	if ref == "" || ref != strings.TrimSpace(ref) {
		return "", errors.New("empty or padded account reference")
	}
	digits := make([]byte, 0, len(ref))
	previousHyphen := false
	for i := 0; i < len(ref); i++ {
		ch := ref[i]
		switch {
		case ch >= '0' && ch <= '9':
			digits = append(digits, ch)
			previousHyphen = false
		case ch == '-' && i > 0 && i < len(ref)-1 && !previousHyphen:
			previousHyphen = true
		default:
			return "", errors.New("invalid account character or separator")
		}
	}
	if len(digits) < 8 || len(digits) > 14 {
		return "", errors.New("account digit count outside 8-14")
	}
	return string(digits), nil
}
