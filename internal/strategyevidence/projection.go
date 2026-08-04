package strategyevidence

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

type Refusal struct {
	Code   RefusalCode
	Kind   EvidenceKind
	Detail string
}

type Requirement struct {
	Kind        EvidenceKind
	Required    bool
	MaxAge      time.Duration
	Authorities []SourceAuthority
	Currency    string
	Unit        string
}

type ProjectionPolicy struct {
	Version  string
	Market   marketclock.Market
	Fatal    []Requirement
	Required []Requirement
	Optional []Requirement
}

type FatalAssessment struct {
	Blocked bool
	Reasons []Refusal
}

type LaneEvidenceSet struct {
	Eligible    bool
	Evidence    map[EvidenceKind]Envelope
	Unavailable map[EvidenceKind]Refusal
	Refusals    []Refusal
}

type ProjectionResult struct {
	PolicyVersion string
	SnapshotID    string
	Fatal         FatalAssessment
	Lane          LaneEvidenceSet
}

func Project(snapshot Snapshot, policy ProjectionPolicy) (ProjectionResult, error) {
	if strings.TrimSpace(policy.Version) == "" || policy.Market != snapshot.Market || strings.TrimSpace(snapshot.IssuerIdentity) == "" || strings.TrimSpace(snapshot.IssuerMappingVersion) == "" || snapshot.EvaluationAt.IsZero() {
		return ProjectionResult{}, fmt.Errorf("strategy evidence: invalid projection policy or snapshot scope")
	}
	result := ProjectionResult{
		PolicyVersion: policy.Version,
		SnapshotID:    snapshot.ID,
		Lane:          LaneEvidenceSet{Eligible: true, Evidence: make(map[EvidenceKind]Envelope), Unavailable: make(map[EvidenceKind]Refusal)},
	}
	for _, requirement := range policy.Fatal {
		evidence, refusal := choose(snapshot, requirement)
		if refusal.Code != "" {
			if requirement.Required {
				result.Fatal.Blocked = true
				result.Fatal.Reasons = append(result.Fatal.Reasons, refusal)
			}
			continue
		}
		var payload struct {
			Blocked *bool  `json:"blocked"`
			Code    string `json:"code"`
		}
		if err := json.Unmarshal(evidence.CanonicalPayload(), &payload); err != nil || payload.Blocked == nil || (*payload.Blocked && strings.TrimSpace(payload.Code) == "") {
			result.Fatal.Blocked = true
			result.Fatal.Reasons = append(result.Fatal.Reasons, Refusal{Code: RefusalPayloadInvalid, Kind: requirement.Kind, Detail: "invalid fatal payload"})
			continue
		}
		if *payload.Blocked {
			result.Fatal.Blocked = true
			result.Fatal.Reasons = append(result.Fatal.Reasons, Refusal{Code: RefusalFatalEvidence, Kind: requirement.Kind, Detail: payload.Code})
		}
	}
	for _, requirement := range append(append([]Requirement(nil), policy.Required...), policy.Optional...) {
		evidence, refusal := choose(snapshot, requirement)
		if refusal.Code != "" {
			result.Lane.Unavailable[requirement.Kind] = refusal
			if requirement.Required {
				result.Lane.Eligible = false
				result.Lane.Refusals = append(result.Lane.Refusals, refusal)
			}
			continue
		}
		result.Lane.Evidence[requirement.Kind] = evidence
	}
	sort.Slice(result.Fatal.Reasons, func(i, j int) bool {
		if result.Fatal.Reasons[i].Kind == result.Fatal.Reasons[j].Kind {
			return result.Fatal.Reasons[i].Code < result.Fatal.Reasons[j].Code
		}
		return result.Fatal.Reasons[i].Kind < result.Fatal.Reasons[j].Kind
	})
	sort.Slice(result.Lane.Refusals, func(i, j int) bool {
		if result.Lane.Refusals[i].Kind == result.Lane.Refusals[j].Kind {
			return result.Lane.Refusals[i].Code < result.Lane.Refusals[j].Code
		}
		return result.Lane.Refusals[i].Kind < result.Lane.Refusals[j].Kind
	})
	return result, nil
}

func choose(snapshot Snapshot, requirement Requirement) (Envelope, Refusal) {
	if !kindSupportsMarket(requirement.Kind, snapshot.Market) {
		return Envelope{}, Refusal{Code: RefusalUnsupportedEvidence, Kind: requirement.Kind, Detail: "kind is unsupported for market"}
	}
	byAuthority := make(map[SourceAuthority][]Envelope)
	identityMismatch := false
	for _, evidence := range snapshot.Items {
		h := evidence.Header()
		if h.Kind == requirement.Kind && (h.Market != snapshot.Market || h.Symbol != snapshot.Symbol || h.IssuerIdentity != snapshot.IssuerIdentity || h.IssuerMappingVersion != snapshot.IssuerMappingVersion) {
			identityMismatch = true
			continue
		}
		if h.Market == snapshot.Market && h.Symbol == snapshot.Symbol && h.IssuerIdentity == snapshot.IssuerIdentity && h.IssuerMappingVersion == snapshot.IssuerMappingVersion && h.Kind == requirement.Kind {
			byAuthority[h.Authority] = append(byAuthority[h.Authority], evidence)
		}
	}
	if identityMismatch {
		return Envelope{}, Refusal{Code: RefusalIdentityMismatch, Kind: requirement.Kind, Detail: "snapshot evidence identity does not match scope"}
	}
	authorities := requirement.Authorities
	if len(authorities) == 0 {
		return Envelope{}, Refusal{Code: RefusalEvidenceMissing, Kind: requirement.Kind, Detail: "versioned authority priority is required"}
	}
	for _, authority := range authorities {
		candidates := byAuthority[authority]
		if len(candidates) == 0 {
			continue
		}
		digests := make(map[string]struct{})
		for _, candidate := range candidates {
			digests[candidate.PayloadDigest()] = struct{}{}
		}
		if len(digests) > 1 {
			return Envelope{}, Refusal{Code: RefusalEvidenceConflict, Kind: requirement.Kind, Detail: "same-priority authoritative evidence conflicts"}
		}
		sort.Slice(candidates, func(i, j int) bool {
			left, right := candidates[i].Header(), candidates[j].Header()
			if left.SourceRecordID != right.SourceRecordID {
				return left.SourceRecordID < right.SourceRecordID
			}
			if left.RevisionIdentity != right.RevisionIdentity {
				return left.RevisionIdentity < right.RevisionIdentity
			}
			return left.EvidenceID < right.EvidenceID
		})
		selected := candidates[0]
		h := selected.Header()
		if h.Availability != AvailabilityAvailable {
			return Envelope{}, Refusal{Code: RefusalSourceUnavailable, Kind: requirement.Kind, Detail: "source marked evidence unavailable"}
		}
		if evidenceStale(snapshot, h, requirement.MaxAge) {
			return Envelope{}, Refusal{Code: RefusalEvidenceStale, Kind: requirement.Kind, Detail: "evidence exceeds versioned freshness"}
		}
		if currency := strings.ToUpper(strings.TrimSpace(requirement.Currency)); currency != "" && h.Currency != currency {
			return Envelope{}, Refusal{Code: RefusalCurrencyUnresolved, Kind: requirement.Kind, Detail: "evidence currency does not match policy"}
		}
		if unit := strings.TrimSpace(requirement.Unit); unit != "" && h.Unit != unit {
			return Envelope{}, Refusal{Code: RefusalUnitUnresolved, Kind: requirement.Kind, Detail: "evidence unit does not match policy"}
		}
		return selected, Refusal{}
	}
	return Envelope{}, Refusal{Code: RefusalEvidenceMissing, Kind: requirement.Kind, Detail: "no eligible authoritative evidence"}
}

func evidenceStale(snapshot Snapshot, header Header, maxAge time.Duration) bool {
	if maxAge <= 0 || header.ObservedAt.After(snapshot.EvaluationAt) || snapshot.EvaluationAt.Sub(header.ObservedAt) > maxAge {
		return true
	}
	location, err := snapshot.Market.Location()
	if err != nil {
		return true
	}
	effective, err := time.ParseInLocation("2006-01-02", header.MarketEffectiveDate, location)
	if err != nil {
		return true
	}
	evaluationLocal := snapshot.EvaluationAt.In(location)
	evaluationDate := time.Date(evaluationLocal.Year(), evaluationLocal.Month(), evaluationLocal.Day(), 0, 0, 0, 0, time.UTC)
	effectiveDate := time.Date(effective.Year(), effective.Month(), effective.Day(), 0, 0, 0, 0, time.UTC)
	effectiveAge := evaluationDate.Sub(effectiveDate)
	return effectiveAge < 0 || effectiveAge > maxAge
}
