package candidate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
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
	ApprovedAt     time.Time         `json:"approved_at"`
	ApprovedBy     string            `json:"approved_by"`
}

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
	evidenceDigest string
	approvedAt     time.Time
	approvedBy     string
}

func (s ThresholdSet) Version() string        { return s.version }
func (s ThresholdSet) Scope() ThresholdScope  { return s.scope }
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

// LoadThresholdSet parses exactly one strict JSON document and accepts it only
// when every evidence, approval, metric, and expected scope field is complete.
// Every failure returns the zero set; no partial field fallback is possible.
func LoadThresholdSet(reader io.Reader, expected ThresholdScope) (ThresholdSet, error) {
	var document thresholdDocument
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return ThresholdSet{}, fmt.Errorf("candidate threshold set: trailing JSON value")
		}
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: trailing data: %w", err)
	}
	set, err := validateThresholdDocument(document, expected)
	if err != nil {
		return ThresholdSet{}, err
	}
	return set, nil
}

func validateThresholdDocument(document thresholdDocument, expected ThresholdScope) (ThresholdSet, error) {
	document.Version = strings.TrimSpace(document.Version)
	document.Market = strings.ToUpper(strings.TrimSpace(document.Market))
	document.Session = strings.ToLower(strings.TrimSpace(document.Session))
	document.EvidenceDigest = strings.TrimSpace(document.EvidenceDigest)
	document.ApprovedBy = strings.TrimSpace(document.ApprovedBy)
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
	case document.ApprovedAt.IsZero():
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: approved_at is required")
	case document.ApprovedBy == "":
		return ThresholdSet{}, fmt.Errorf("candidate threshold set: approved_by is required")
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
		valid: true, version: document.Version,
		scope:   ThresholdScope{Market: document.Market, Session: document.Session},
		metrics: metrics, sampleWindow: document.SampleWindow, sampleCount: document.SampleCount,
		missingRate: strings.TrimSpace(document.MissingRate), evidenceDigest: document.EvidenceDigest,
		approvedAt: document.ApprovedAt.UTC(), approvedBy: document.ApprovedBy,
	}, nil
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

// ApprovedCandidate is the order-free verdict exposed after a complete set has
// been loaded. ThresholdVersion makes later evidence attribution explicit.
type ApprovedCandidate struct {
	Key
	Chase            Chase
	ThresholdVersion string
}

func AssessApprovedCandidate(input VetoInputs, set ThresholdSet) (ApprovedCandidate, error) {
	if !set.valid {
		return ApprovedCandidate{}, fmt.Errorf("candidate threshold set: unapproved or invalid set")
	}
	if input.Candidate.Market != set.scope.Market {
		return ApprovedCandidate{}, fmt.Errorf("candidate threshold set: candidate market %q does not match %q",
			input.Candidate.Market, set.scope.Market)
	}
	input.Thresholds = set.VetoThresholds()
	return ApprovedCandidate{
		Key: input.Candidate.Key, Chase: AssessChase(input), ThresholdVersion: set.version,
	}, nil
}
