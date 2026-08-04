package strategyrouter

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"time"
)

// OwnerKey is the only position ownership scope. Horizon must never be added.
type OwnerKey struct {
	AccountRef         string
	Market             Market
	Symbol             string
	PositionGeneration uint64
}

func NewOwnerKey(account string, market Market, symbol string, generation uint64) (OwnerKey, error) {
	key := OwnerKey{
		AccountRef:         strings.TrimSpace(account),
		Market:             market,
		Symbol:             strings.ToUpper(strings.TrimSpace(symbol)),
		PositionGeneration: generation,
	}
	if err := boundedNonEmpty("account", key.AccountRef); err != nil {
		return OwnerKey{}, err
	}
	if !validMarket(key.Market) {
		return OwnerKey{}, errors.New("strategyrouter: unsupported market")
	}
	if err := boundedNonEmpty("symbol", key.Symbol); err != nil {
		return OwnerKey{}, err
	}
	if key.PositionGeneration == 0 {
		return OwnerKey{}, errors.New("strategyrouter: position generation is zero")
	}
	return key, nil
}

func validOwnerKey(key OwnerKey) bool {
	canonical, err := NewOwnerKey(key.AccountRef, key.Market, key.Symbol, key.PositionGeneration)
	return err == nil && canonical == key
}

type Owner struct {
	Key         OwnerKey
	Horizon     Horizon
	LaneID      string
	LaneVersion string
	CampaignID  string
	Active      bool
	Desired     DesiredState
	Effective   DesiredState
}

type OwnerSnapshot struct {
	Key        OwnerKey
	Revision   uint64
	Digest     string
	ObservedAt time.Time
	FreshUntil time.Time
	Owners     []Owner
	seal       [32]byte
}

func newOwnerSnapshot(key OwnerKey, revision uint64, digest string, observedAt, freshUntil time.Time, owners []Owner) (OwnerSnapshot, error) {
	if !validOwnerKey(key) || revision == 0 || boundedNonEmpty("owner digest", digest) != nil || observedAt.IsZero() || !freshUntil.After(observedAt) {
		return OwnerSnapshot{}, errors.New("strategyrouter: invalid owner snapshot")
	}
	copyOwners := append([]Owner(nil), owners...)
	sort.Slice(copyOwners, func(i, j int) bool {
		if copyOwners[i].Horizon != copyOwners[j].Horizon {
			return copyOwners[i].Horizon < copyOwners[j].Horizon
		}
		if copyOwners[i].LaneID != copyOwners[j].LaneID {
			return copyOwners[i].LaneID < copyOwners[j].LaneID
		}
		return copyOwners[i].CampaignID < copyOwners[j].CampaignID
	})
	for _, owner := range copyOwners {
		if !validOwner(owner) {
			return OwnerSnapshot{}, errors.New("strategyrouter: invalid owner")
		}
	}
	snapshot := OwnerSnapshot{Key: key, Revision: revision, Digest: digest, ObservedAt: observedAt.UTC(), FreshUntil: freshUntil.UTC(), Owners: copyOwners}
	snapshot.seal = ownerSnapshotSeal(snapshot)
	return snapshot, nil
}

func validOwner(owner Owner) bool {
	return validOwnerKey(owner.Key) && validHorizon(owner.Horizon) &&
		boundedNonEmpty("lane id", owner.LaneID) == nil && boundedNonEmpty("lane version", owner.LaneVersion) == nil &&
		boundedNonEmpty("campaign id", owner.CampaignID) == nil && validDesiredState(owner.Desired) && validDesiredState(owner.Effective)
}

func ownerSnapshotSeal(snapshot OwnerSnapshot) [32]byte {
	h := sha256.New()
	writeOwnerKey(h, snapshot.Key)
	writeUint64(h, snapshot.Revision)
	writeString(h, snapshot.Digest)
	writeString(h, snapshot.ObservedAt.UTC().Format(time.RFC3339Nano))
	writeString(h, snapshot.FreshUntil.UTC().Format(time.RFC3339Nano))
	writeUint64(h, uint64(len(snapshot.Owners)))
	for _, owner := range snapshot.Owners {
		writeOwnerKey(h, owner.Key)
		writeString(h, string(owner.Horizon))
		writeString(h, owner.LaneID)
		writeString(h, owner.LaneVersion)
		writeString(h, owner.CampaignID)
		if owner.Active {
			writeString(h, "1")
		} else {
			writeString(h, "0")
		}
		writeString(h, string(owner.Desired))
		writeString(h, string(owner.Effective))
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

type hashWriter interface {
	Write([]byte) (int, error)
}

func writeString(h hashWriter, value string) {
	writeUint64(h, uint64(len(value)))
	_, _ = h.Write([]byte(value))
}

func writeUint64(h hashWriter, value uint64) {
	var buffer [8]byte
	binary.BigEndian.PutUint64(buffer[:], value)
	_, _ = h.Write(buffer[:])
}

func writeOwnerKey(h hashWriter, key OwnerKey) {
	writeString(h, key.AccountRef)
	writeString(h, string(key.Market))
	writeString(h, key.Symbol)
	writeUint64(h, key.PositionGeneration)
}

type Candidate struct {
	Key            OwnerKey
	Horizon        Horizon
	LaneID         string
	LaneVersion    string
	Score          int64
	Eligible       bool
	Desired        DesiredState
	Effective      DesiredState
	EvidenceDigest string
	ConfigDigest   string
}

func validCandidateValue(candidate Candidate) bool {
	return validOwnerKey(candidate.Key) && validHorizon(candidate.Horizon) &&
		boundedNonEmpty("lane id", candidate.LaneID) == nil && boundedNonEmpty("lane version", candidate.LaneVersion) == nil &&
		boundedNonEmpty("evidence digest", candidate.EvidenceDigest) == nil && boundedNonEmpty("config digest", candidate.ConfigDigest) == nil &&
		validDesiredState(candidate.Desired) && validDesiredState(candidate.Effective)
}

type RouteRequest struct {
	Key                    OwnerKey
	ExpectedOwnerRevision  uint64
	ExpectedMarketRevision uint64
	EvaluatedAt            time.Time
	Snapshot               OwnerSnapshot
	MarketRecord           MarketRecord
	Candidates             []Candidate
}

type RouteDecision struct {
	Key            OwnerKey
	Horizon        Horizon
	LaneID         string
	LaneVersion    string
	CampaignID     string
	EvidenceDigest string
	ConfigDigest   string
	ExistingOwner  bool
}

type RouteResult struct {
	Code                    RefusalCode
	Decision                RouteDecision
	CommonSafetyIndependent bool
	Mutations               uint64
}

func Route(request RouteRequest) RouteResult {
	result := RouteResult{CommonSafetyIndependent: true}
	if !validOwnerKey(request.Key) || request.ExpectedOwnerRevision == 0 || request.EvaluatedAt.IsZero() {
		result.Code = RefusalInvalid
		return result
	}
	snapshot := request.Snapshot
	if snapshot.seal != ownerSnapshotSeal(snapshot) || !validOwnerKey(snapshot.Key) {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if snapshot.Key != request.Key {
		result.Code = RefusalScopeMismatch
		return result
	}
	if snapshot.Revision != request.ExpectedOwnerRevision {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if request.EvaluatedAt.Before(snapshot.ObservedAt) || !request.EvaluatedAt.Before(snapshot.FreshUntil) {
		result.Code = RefusalOwnerSnapshotStale
		return result
	}
	active := make([]Owner, 0, 1)
	for _, owner := range snapshot.Owners {
		if owner.Key != request.Key {
			result.Code = RefusalScopeMismatch
			return result
		}
		if !validOwner(owner) {
			result.Code = RefusalReconstructionMismatch
			return result
		}
		if owner.Active {
			active = append(active, owner)
		}
	}
	if len(active) > 1 {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if !validMarketRecord(request.MarketRecord) {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if request.MarketRecord.Market != request.Key.Market {
		result.Code = RefusalScopeMismatch
		return result
	}
	if request.ExpectedMarketRevision == 0 || request.MarketRecord.Revision != request.ExpectedMarketRevision {
		result.Code = RefusalVersionConflict
		return result
	}
	if EvaluateMarketLifecycle(request.MarketRecord, request.EvaluatedAt) != LifecycleReady {
		result.Code = RefusalDisabled
		return result
	}
	if len(active) == 1 {
		owner := active[0]
		if owner.Desired != StateOn || owner.Effective != StateOn {
			result.Code = RefusalDisabled
			return result
		}
		result.Decision = RouteDecision{Key: owner.Key, Horizon: owner.Horizon, LaneID: owner.LaneID, LaneVersion: owner.LaneVersion, CampaignID: owner.CampaignID, ExistingOwner: true}
		return result
	}

	var selected Candidate
	found, tie := false, false
	for _, candidate := range request.Candidates {
		if candidate.Key != request.Key {
			result.Code = RefusalScopeMismatch
			return result
		}
		if !validCandidateValue(candidate) {
			result.Code = RefusalInvalid
			return result
		}
		if !candidate.Eligible || candidate.Desired != StateOn || candidate.Effective != StateOn {
			continue
		}
		if !found || candidate.Score > selected.Score {
			selected, found, tie = candidate, true, false
		} else if candidate.Score == selected.Score {
			tie = true
		}
	}
	if tie {
		result.Code = RefusalAmbiguous
		return result
	}
	if !found {
		result.Code = RefusalDisabled
		return result
	}
	result.Decision = RouteDecision{Key: selected.Key, Horizon: selected.Horizon, LaneID: selected.LaneID, LaneVersion: selected.LaneVersion, EvidenceDigest: selected.EvidenceDigest, ConfigDigest: selected.ConfigDigest}
	return result
}
