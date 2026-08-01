package candidate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"
)

const SessionRegular = "regular"

const (
	SeenLateDefinition = "first-sighting rank percentile"
	ExtendedDefinition = "gain from stored first price"
	NearHighDefinition = "distance below intraday high"
)

type ThresholdScope struct {
	Market  string
	Session string
}

type thresholdMetric struct {
	Key        VetoCode `json:"key"`
	Definition string   `json:"definition"`
	Value      string   `json:"value"`
}

type thresholdWindow struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type thresholdDocument struct {
	Version        string            `json:"version"`
	Market         string            `json:"market"`
	Session        string            `json:"session"`
	Metrics        []thresholdMetric `json:"metrics"`
	SampleWindow   thresholdWindow   `json:"sample_window"`
	SampleCount    int               `json:"sample_count"`
	MissingRate    string            `json:"missing_rate"`
	EvidenceDigest string            `json:"evidence_digest"`
}

type activationDocument struct {
	Version        string    `json:"version"`
	Market         string    `json:"market"`
	Session        string    `json:"session"`
	SetDigest      string    `json:"set_digest"`
	EvidenceDigest string    `json:"evidence_digest"`
	ApprovedAt     time.Time `json:"approved_at"`
	ApprovedBy     string    `json:"approved_by"`
}

// ActivationRecord is a strict, immutable human approval reference. Its private
// fields ensure callers can obtain one only through LoadActivationRecord; a
// zero record is not an approval.
type ActivationRecord struct {
	valid          bool
	version        string
	scope          ThresholdScope
	setDigest      string
	evidenceDigest string
	approvedAt     time.Time
	approvedBy     string
}

func (r ActivationRecord) Version() string        { return r.version }
func (r ActivationRecord) Scope() ThresholdScope  { return r.scope }
func (r ActivationRecord) SetDigest() string      { return r.setDigest }
func (r ActivationRecord) EvidenceDigest() string { return r.evidenceDigest }
func (r ActivationRecord) ApprovedAt() time.Time  { return r.approvedAt }
func (r ActivationRecord) ApprovedBy() string     { return r.approvedBy }

// ThresholdSet is an immutable, validated threshold registry entry. All fields
// are private and accessors return values, so callers cannot partially mutate an
// approved version after loading it.
type ThresholdSet struct {
	valid          bool
	version        string
	scope          ThresholdScope
	metrics        [3]thresholdMetric
	sampleWindow   thresholdWindow
	sampleCount    int
	missingRate    string
	setDigest      string
	evidenceDigest string
	approvedAt     time.Time
	approvedBy     string
}

func (s ThresholdSet) Version() string        { return s.version }
func (s ThresholdSet) Scope() ThresholdScope  { return s.scope }
func (s ThresholdSet) SetDigest() string      { return s.setDigest }
func (s ThresholdSet) EvidenceDigest() string { return s.evidenceDigest }
func (s ThresholdSet) SampleCount() int       { return s.sampleCount }
func (s ThresholdSet) MissingRate() string    { return s.missingRate }
func (s ThresholdSet) ApprovedAt() time.Time  { return s.approvedAt }
func (s ThresholdSet) ApprovedBy() string     { return s.approvedBy }

func (s ThresholdSet) VetoThresholds() VetoThresholds {
	if !s.valid {
		return VetoThresholds{}
	}
	return VetoThresholds{
		SeenLatePercentilePct: s.metrics[0].Value,
		ExtendedGainPct:       s.metrics[1].Value,
		NearHighDistancePct:   s.metrics[2].Value,
	}
}

// LoadActivationRecord parses exactly one strict JSON approval record. It does
// not activate anything by itself; LoadThresholdSet binds it to a canonical set,
// opaque evidence bytes, an expected scope, and an injected time boundary.
func LoadActivationRecord(reader io.Reader) (ActivationRecord, error) {
	var document activationDocument
	if err := decodeOneStrictJSON(reader, &document, "activation record"); err != nil {
		return ActivationRecord{}, err
	}
	document.Version = strings.TrimSpace(document.Version)
	document.Market = strings.ToUpper(strings.TrimSpace(document.Market))
	document.Session = strings.ToLower(strings.TrimSpace(document.Session))
	document.SetDigest = strings.TrimSpace(document.SetDigest)
	document.EvidenceDigest = strings.TrimSpace(document.EvidenceDigest)
	document.ApprovedBy = strings.TrimSpace(document.ApprovedBy)
	switch {
	case document.Version == "":
		return ActivationRecord{}, fmt.Errorf("candidate activation record: version is required")
	case document.Market != MarketKR && document.Market != MarketUS:
		return ActivationRecord{}, fmt.Errorf("candidate activation record: market %q is unsupported", document.Market)
	case document.Session != SessionRegular:
		return ActivationRecord{}, fmt.Errorf("candidate activation record: session %q is unsupported", document.Session)
	case !digestPattern.MatchString(document.SetDigest):
		return ActivationRecord{}, fmt.Errorf("candidate activation record: set_digest must be sha256 plus 64 lowercase hex digits")
	case !digestPattern.MatchString(document.EvidenceDigest):
		return ActivationRecord{}, fmt.Errorf("candidate activation record: evidence_digest must be sha256 plus 64 lowercase hex digits")
	case document.ApprovedAt.IsZero():
		return ActivationRecord{}, fmt.Errorf("candidate activation record: approved_at is required")
	case document.ApprovedBy == "":
		return ActivationRecord{}, fmt.Errorf("candidate activation record: approved_by is required")
	}
	return ActivationRecord{
		valid: true, version: document.Version,
		scope:          ThresholdScope{Market: document.Market, Session: document.Session},
		setDigest:      document.SetDigest,
		evidenceDigest: document.EvidenceDigest,
		approvedAt:     document.ApprovedAt.UTC(),
		approvedBy:     document.ApprovedBy,
	}, nil
}

// DigestThresholdSetDocument validates and canonicalizes exactly one threshold
// document. Approval tooling uses this digest in a separate ActivationRecord;
// computing it never approves or activates numeric values.
func DigestThresholdSetDocument(reader io.Reader, expected ThresholdScope) (string, error) {
	set, err := parseThresholdDocument(reader, expected)
	if err != nil {
		return "", err
	}
	return canonicalThresholdSetDigest(set)
}

// LoadThresholdSet accepts a set only when opaque evidence bytes, the separate
// activation record, the canonical set digest, version, scope, evidence digest,
// and approval time all agree. asOf is injected by the caller; futureSkew is an
// explicit non-negative allowance. Every failure returns the zero set.
func LoadThresholdSet(reader io.Reader, evidence []byte, activation ActivationRecord, expected ThresholdScope,
	asOf time.Time, futureSkew time.Duration,
) (ThresholdSet, error) {
	if len(evidence) == 0 {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: opaque evidence bytes are required")
	}
	if !activation.valid {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: activation record is absent or invalid")
	}
	if asOf.IsZero() {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: injected asOf is required")
	}
	if futureSkew < 0 {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: future skew must be non-negative")
	}
	set, err := parseThresholdDocument(reader, expected)
	if err != nil {
		return ThresholdSet{}, err
	}
	setDigest, err := canonicalThresholdSetDigest(set)
	if err != nil {
		return ThresholdSet{}, err
	}
	evidenceDigest := DigestEvidence(evidence)
	switch {
	case set.evidenceDigest != evidenceDigest:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: evidence digest does not match opaque evidence bytes")
	case activation.evidenceDigest != evidenceDigest:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: activation evidence digest does not match opaque evidence bytes")
	case activation.setDigest != setDigest:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: activation set digest does not match canonical set digest")
	case activation.version != set.version:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: activation version %q does not match set version %q",
			activation.version, set.version)
	case activation.scope != set.scope:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: activation scope %s/%s does not match set scope %s/%s",
			activation.scope.Market, activation.scope.Session, set.scope.Market, set.scope.Session)
	case activation.approvedAt.Before(set.sampleWindow.To):
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: approved_at precedes sample_window.to")
	case activation.approvedAt.After(asOf.UTC().Add(futureSkew)):
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: approved_at exceeds injected asOf plus future skew")
	}
	set.valid = true
	set.setDigest = setDigest
	set.approvedAt = activation.approvedAt
	set.approvedBy = activation.approvedBy
	return set, nil
}

func parseThresholdDocument(reader io.Reader, expected ThresholdScope) (ThresholdSet, error) {
	var document thresholdDocument
	if err := decodeOneStrictJSON(reader, &document, "threshold set"); err != nil {
		return ThresholdSet{}, err
	}
	return validateThresholdDocument(document, expected)
}

func decodeOneStrictJSON(reader io.Reader, destination any, kind string) error {
	if reader == nil {
		return fmt.Errorf("candidate %s: document is required", kind)
	}
	const maximumDocumentBytes = 1 << 20
	raw, err := io.ReadAll(io.LimitReader(reader, maximumDocumentBytes+1))
	if err != nil {
		return fmt.Errorf("candidate %s: read: %w", kind, err)
	}
	if len(raw) > maximumDocumentBytes {
		return fmt.Errorf("candidate %s: document exceeds %d bytes", kind, maximumDocumentBytes)
	}
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return fmt.Errorf("candidate %s: %w", kind, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("candidate %s: decode: %w", kind, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("candidate %s: trailing JSON value", kind)
		}
		return fmt.Errorf("candidate %s: trailing data: %w", kind, err)
	}
	return nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var walkValue func() error
	walkValue = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, composite := token.(json.Delim)
		if !composite {
			return nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				if _, duplicate := seen[key]; duplicate {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = struct{}{}
				if err := walkValue(); err != nil {
					return err
				}
			}
		case '[':
			for decoder.More() {
				if err := walkValue(); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
		_, err = decoder.Token()
		return err
	}
	return walkValue()
}

func validateThresholdDocument(document thresholdDocument, expected ThresholdScope) (ThresholdSet, error) {
	document.Version = strings.TrimSpace(document.Version)
	document.Market = strings.ToUpper(strings.TrimSpace(document.Market))
	document.Session = strings.ToLower(strings.TrimSpace(document.Session))
	document.EvidenceDigest = strings.TrimSpace(document.EvidenceDigest)
	expected.Market = strings.ToUpper(strings.TrimSpace(expected.Market))
	expected.Session = strings.ToLower(strings.TrimSpace(expected.Session))

	switch {
	case expected.Market == "" || expected.Session == "":
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: expected scope is incomplete")
	case document.Market != expected.Market || document.Session != expected.Session:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: scope %s/%s does not match expected %s/%s",
			document.Market, document.Session, expected.Market, expected.Session)
	case document.Market != MarketKR && document.Market != MarketUS:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: market %q is unsupported", document.Market)
	case document.Session != SessionRegular:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: session %q is unsupported", document.Session)
	case document.Version == "":
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: version is required")
	case document.SampleWindow.From.IsZero() || document.SampleWindow.To.IsZero() ||
		!document.SampleWindow.From.Before(document.SampleWindow.To):
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: sample_window must have increasing from/to instants")
	case document.SampleCount <= 0:
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: sample_count must be positive")
	case !digestPattern.MatchString(document.EvidenceDigest):
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: evidence_digest must be sha256 plus 64 lowercase hex digits")
	}

	missing, err := decimalRatio(document.MissingRate)
	if err != nil || missing.Sign() < 0 || missing.Cmp(big.NewRat(1, 1)) > 0 {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: missing_rate must be a decimal from 0 through 1")
	}

	want := [...]struct {
		key        VetoCode
		definition string
	}{
		{VetoSeenLate, SeenLateDefinition},
		{VetoExtended, ExtendedDefinition},
		{VetoNearHigh, NearHighDefinition},
	}
	byKey := make(map[VetoCode]thresholdMetric, len(document.Metrics))
	for _, metric := range document.Metrics {
		metric.Definition = strings.TrimSpace(metric.Definition)
		metric.Value = strings.TrimSpace(metric.Value)
		if _, duplicate := byKey[metric.Key]; duplicate {
			return ThresholdSet{}, fmt.Errorf("candidate threshold set: duplicate metric %q", metric.Key)
		}
		byKey[metric.Key] = metric
	}
	var metrics [3]thresholdMetric
	for index, contract := range want {
		metric, found := byKey[contract.key]
		if !found {
			return ThresholdSet{}, fmt.Errorf("candidate threshold set: metric %s is required", contract.key)
		}
		if metric.Definition != contract.definition {
			return ThresholdSet{}, fmt.Errorf("candidate threshold set: metric %s definition is %q, want %q",
				contract.key, metric.Definition, contract.definition)
		}
		if why := thresholdReason(metric.Value); why != "" {
			return ThresholdSet{}, fmt.Errorf("candidate threshold set: metric %s value is invalid: %s", contract.key, why)
		}
		metrics[index] = metric
	}
	if len(byKey) != len(want) {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: metrics contain an unsupported key")
	}

	return ThresholdSet{
		version: document.Version,
		scope:   ThresholdScope{Market: document.Market, Session: document.Session},
		metrics: metrics,
		sampleWindow: thresholdWindow{
			From: document.SampleWindow.From.UTC(),
			To:   document.SampleWindow.To.UTC(),
		},
		sampleCount:    document.SampleCount,
		missingRate:    strings.TrimSpace(document.MissingRate),
		evidenceDigest: document.EvidenceDigest,
	}, nil
}

func canonicalThresholdSetDigest(set ThresholdSet) (string, error) {
	payload := struct {
		Version        string             `json:"version"`
		Market         string             `json:"market"`
		Session        string             `json:"session"`
		Metrics        [3]thresholdMetric `json:"metrics"`
		SampleWindow   thresholdWindow    `json:"sample_window"`
		SampleCount    int                `json:"sample_count"`
		MissingRate    string             `json:"missing_rate"`
		EvidenceDigest string             `json:"evidence_digest"`
	}{
		Version: set.version, Market: set.scope.Market, Session: set.scope.Session,
		Metrics: set.metrics, SampleWindow: set.sampleWindow, SampleCount: set.sampleCount,
		MissingRate: set.missingRate, EvidenceDigest: set.evidenceDigest,
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("candidate threshold set: canonical JSON: %w", err)
	}
	return DigestEvidence(canonical), nil
}

var (
	digestPattern       = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	plainDecimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)
)

func decimalRatio(raw string) (*big.Rat, error) {
	raw = strings.TrimSpace(raw)
	if !plainDecimalPattern.MatchString(raw) {
		return nil, fmt.Errorf("not a plain decimal")
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, fmt.Errorf("not a decimal")
	}
	return value, nil
}

func DigestEvidence(evidence []byte) string {
	sum := sha256.Sum256(evidence)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ThresholdRegistry rejects version aliasing: once a version identifies one
// canonical set digest, a different set cannot reuse that version.
type ThresholdRegistry struct {
	mu      sync.Mutex
	digests map[string]string
}

func NewThresholdRegistry() *ThresholdRegistry {
	return &ThresholdRegistry{digests: make(map[string]string)}
}

func (r *ThresholdRegistry) LoadThresholdSet(reader io.Reader, evidence []byte, activation ActivationRecord,
	expected ThresholdScope, asOf time.Time, futureSkew time.Duration,
) (ThresholdSet, error) {
	set, err := LoadThresholdSet(reader, evidence, activation, expected, asOf, futureSkew)
	if err != nil {
		return ThresholdSet{}, err
	}
	if err := r.register(set); err != nil {
		return ThresholdSet{}, err
	}
	return set, nil
}

func (r *ThresholdRegistry) register(set ThresholdSet) error {
	if r == nil {
		return fmt.Errorf("candidate threshold registry: registry is required")
	}
	if !set.valid {
		return fmt.Errorf("candidate threshold registry: unapproved or invalid set")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.digests == nil {
		r.digests = make(map[string]string)
	}
	if digest, found := r.digests[set.version]; found && digest != set.setDigest {
		return fmt.Errorf("candidate threshold registry: same version %q has different canonical set digest", set.version)
	}
	r.digests[set.version] = set.setDigest
	return nil
}

// CandidateLifeID is the opaque, deterministic identity of one candidate life.
// It is derived only from the normalized Key and FirstSeenAt and never exposes
// the raw symbol. A symbol that expires and crosses again gets a different ID.
type CandidateLifeID string

func (id CandidateLifeID) String() string { return string(id) }

// ApprovalErrorKind is the fail-closed reason an ApprovedCandidate was not
// minted. Callers can route refusals without parsing Error() text.
type ApprovalErrorKind string

const (
	ApprovalInvalidSet           ApprovalErrorKind = "invalid_set"
	ApprovalScopeMismatch        ApprovalErrorKind = "scope_mismatch"
	ApprovalInvalidCandidateLife ApprovalErrorKind = "invalid_candidate_life"
	ApprovalVetoRaised           ApprovalErrorKind = "veto_raised"
	ApprovalVetoUnmeasured       ApprovalErrorKind = "veto_unmeasured"
)

// ApprovalError is a typed refusal. Vetoes returns a copy in D3 order so an
// external caller cannot mutate the diagnostic held by the error.
type ApprovalError struct {
	kind      ApprovalErrorKind
	vetoes    [3]VetoCode
	vetoCount int
	detail    string
}

func (e *ApprovalError) Error() string {
	if e == nil {
		return "candidate approval: <nil>"
	}
	if e.detail == "" {
		return "candidate approval: " + string(e.kind)
	}
	return "candidate approval: " + string(e.kind) + ": " + e.detail
}

func (e *ApprovalError) Kind() ApprovalErrorKind {
	if e == nil {
		return ""
	}
	return e.kind
}

func (e *ApprovalError) Vetoes() []VetoCode {
	if e == nil || e.vetoCount == 0 {
		return nil
	}
	out := make([]VetoCode, e.vetoCount)
	copy(out, e.vetoes[:e.vetoCount])
	return out
}

func newApprovalError(kind ApprovalErrorKind, detail string, vetoes []VetoCode) *ApprovalError {
	err := &ApprovalError{kind: kind, detail: detail}
	err.vetoCount = min(len(vetoes), len(err.vetoes))
	copy(err.vetoes[:], vetoes[:err.vetoCount])
	return err
}

// ApprovedCandidate is an immutable, order-free, measured-and-clear verdict.
// All fields are private so callers cannot turn a refusal or a different Chase
// into an approved value. Accessors return value copies.
type ApprovedCandidate struct {
	valid            bool
	key              Key
	state            State
	firstSeenAt      time.Time
	lastSeenAt       time.Time
	validUntil       time.Time
	chase            Chase
	candidateLifeID  CandidateLifeID
	thresholdVersion string
	setDigest        string
	evidenceDigest   string
	approvedAt       time.Time
}

func (c ApprovedCandidate) Valid() bool                      { return c.valid }
func (c ApprovedCandidate) Key() Key                         { return c.key }
func (c ApprovedCandidate) State() State                     { return c.state }
func (c ApprovedCandidate) FirstSeenAt() time.Time           { return c.firstSeenAt }
func (c ApprovedCandidate) LastSeenAt() time.Time            { return c.lastSeenAt }
func (c ApprovedCandidate) ValidUntil() time.Time            { return c.validUntil }
func (c ApprovedCandidate) Chase() Chase                     { return c.chase }
func (c ApprovedCandidate) CandidateLifeID() CandidateLifeID { return c.candidateLifeID }
func (c ApprovedCandidate) ThresholdVersion() string         { return c.thresholdVersion }
func (c ApprovedCandidate) SetDigest() string                { return c.setDigest }
func (c ApprovedCandidate) EvidenceDigest() string           { return c.evidenceDigest }
func (c ApprovedCandidate) ApprovedAt() time.Time            { return c.approvedAt }
func (c ApprovedCandidate) MarketString() string             { return c.key.Market }
func (c ApprovedCandidate) SymbolString() string             { return c.key.Symbol }
func (c ApprovedCandidate) CandidateLifeIDString() string    { return c.candidateLifeID.String() }
func (c ApprovedCandidate) StateString() string              { return string(c.state) }
func (c ApprovedCandidate) FirstSeenUnixNano() int64         { return c.firstSeenAt.UTC().UnixNano() }
func (c ApprovedCandidate) LastSeenUnixNano() int64          { return c.lastSeenAt.UTC().UnixNano() }
func (c ApprovedCandidate) ValidUntilUnixNano() int64        { return c.validUntil.UTC().UnixNano() }
func (c ApprovedCandidate) ApprovedAtUnixNano() int64        { return c.approvedAt.UTC().UnixNano() }

func candidateLifeID(candidate Candidate) (CandidateLifeID, error) {
	market := strings.ToUpper(strings.TrimSpace(candidate.Market))
	symbol := strings.TrimSpace(candidate.Symbol)
	switch {
	case market != candidate.Market || (market != MarketKR && market != MarketUS):
		return "", fmt.Errorf("market is absent, unsupported, or non-canonical")
	case symbol == "" || symbol != candidate.Symbol:
		return "", fmt.Errorf("symbol is absent or non-canonical")
	case candidate.FirstSeenAt.IsZero():
		return "", fmt.Errorf("first_seen_at is required")
	}
	payload := "tossos:candidate-life:v1\x00" + market + "\x00" + symbol + "\x00" +
		candidate.FirstSeenAt.UTC().Format(time.RFC3339Nano)
	sum := sha256.Sum256([]byte(payload))
	return CandidateLifeID("candidate-life:v1:sha256:" + hex.EncodeToString(sum[:])), nil
}

func AssessApprovedCandidate(input VetoInputs, set ThresholdSet) (ApprovedCandidate, error) {
	if !set.valid {
		return ApprovedCandidate{}, newApprovalError(ApprovalInvalidSet,
			"candidate threshold set is unapproved or invalid", nil)
	}
	lifeID, err := candidateLifeID(input.Candidate)
	if err != nil {
		return ApprovedCandidate{}, newApprovalError(ApprovalInvalidCandidateLife, err.Error(), nil)
	}
	if input.Candidate.State != StateActive || input.Candidate.LastSeenAt.IsZero() ||
		input.Candidate.LastSeenAt.Before(input.Candidate.FirstSeenAt) ||
		!input.Candidate.CooledAt.IsZero() || input.At.Before(input.Candidate.LastSeenAt) ||
		input.At.Sub(input.Candidate.LastSeenAt) >= DefaultStalenessTTL {
		return ApprovedCandidate{}, newApprovalError(ApprovalInvalidCandidateLife,
			"candidate is not a current active life", nil)
	}
	if input.Candidate.Market != set.scope.Market {
		return ApprovedCandidate{}, newApprovalError(ApprovalScopeMismatch,
			fmt.Sprintf("candidate market %q does not match threshold market %q",
				input.Candidate.Market, set.scope.Market), nil)
	}
	input.Thresholds = set.VetoThresholds()
	chase := AssessChase(input)
	if raised := chase.Raised(); len(raised) != 0 {
		return ApprovedCandidate{}, newApprovalError(ApprovalVetoRaised,
			"one or more chase vetoes are dangerous", raised)
	}
	if unmeasured := chase.NotMeasured(); len(unmeasured) != 0 {
		return ApprovedCandidate{}, newApprovalError(ApprovalVetoUnmeasured,
			"one or more chase vetoes are not measured", unmeasured)
	}
	if !chase.Passed() {
		return ApprovedCandidate{}, newApprovalError(ApprovalVetoUnmeasured,
			"chase verdict did not establish measured-and-clear", nil)
	}
	return ApprovedCandidate{
		valid: true, key: input.Candidate.Key, state: input.Candidate.State,
		firstSeenAt: input.Candidate.FirstSeenAt, lastSeenAt: input.Candidate.LastSeenAt,
		validUntil: input.Candidate.LastSeenAt.Add(DefaultStalenessTTL),
		chase:      chase, candidateLifeID: lifeID,
		thresholdVersion: set.version, setDigest: set.setDigest,
		evidenceDigest: set.evidenceDigest, approvedAt: set.approvedAt,
	}, nil
}
