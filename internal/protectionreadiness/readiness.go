package protectionreadiness

import "crypto/sha256"

type marketAssessmentInput struct {
	Scope      runtimeScope
	File       observedFile
	Supervisor supervisorBinding
}

type assessmentInput struct {
	Policy  pinnedTrustPolicy
	State   durableState
	Time    trustedTime
	Markets map[Market]marketAssessmentInput
}

type ReadinessSnapshot struct {
	release string
	kr      Verdict
	us      Verdict
	seal    [32]byte
}

func DefaultSnapshot() ReadinessSnapshot {
	snapshot := ReadinessSnapshot{release: ReadinessRelease,
		kr: Verdict{Market: MarketKR, State: Unwired, Code: RefusalMissingEvidence},
		us: Verdict{Market: MarketUS, State: Unwired, Code: RefusalMissingEvidence}}
	snapshot.seal = readinessSnapshotSeal(snapshot)
	return snapshot
}

func (snapshot ReadinessSnapshot) Verdict(market Market) Verdict {
	if snapshot.seal != readinessSnapshotSeal(snapshot) {
		return Verdict{Market: market, State: Unwired, Code: RefusalStateCorrupt}
	}
	switch market {
	case MarketKR:
		return snapshot.kr
	case MarketUS:
		return snapshot.us
	default:
		return Verdict{Market: market, State: Unwired, Code: RefusalInvalid}
	}
}

func (snapshot ReadinessSnapshot) Release() string {
	if snapshot.seal != readinessSnapshotSeal(snapshot) {
		return ""
	}
	return snapshot.release
}

func readinessSnapshotSeal(snapshot ReadinessSnapshot) [32]byte {
	hash := sha256.New()
	writeString(hash, snapshot.release)
	for _, verdict := range []Verdict{snapshot.kr, snapshot.us} {
		writeString(hash, string(verdict.Market))
		writeString(hash, string(verdict.State))
		writeString(hash, string(verdict.Code))
		writeString(hash, verdict.Provenance.KeyID)
		writeUint64(hash, verdict.Provenance.Serial)
		writeString(hash, verdict.Provenance.BodyDigest)
		writeString(hash, verdict.Provenance.BuildDigest)
		writeString(hash, verdict.Provenance.EvidenceDigest)
		writeString(hash, formatTime(verdict.Provenance.IssuedAt))
		writeString(hash, formatTime(verdict.Provenance.ExpiresAt))
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type AssessmentResult struct {
	Snapshot                   ReadinessSnapshot
	NextState                  durableState
	Mutations                  uint64
	ExternalMutations          uint64
	StateCommitAllowed         bool
	NoLaneAuthority            bool
	NoLiveAuthority            bool
	PreserveExistingProtection bool
	PreserveReduceOnlyExit     bool
}

func Assess(input assessmentInput) AssessmentResult {
	result := AssessmentResult{Snapshot: DefaultSnapshot(), NextState: input.State, NoLaneAuthority: true, NoLiveAuthority: true, PreserveExistingProtection: true, PreserveReduceOnlyExit: true}
	policyValid := input.Policy.seal == pinnedPolicySeal(input.Policy) && input.Policy.release == ReadinessRelease
	stateValid := validDurableState(input.State)
	timeValid := validTrustedTime(input.Time)
	timeRollback := timeValid && stateValid && !input.State.TrustedTimeFloor.IsZero() && input.Time.Now.Before(input.State.TrustedTimeFloor)
	if stateValid {
		result.NextState = cloneDurableState(input.State)
	}
	if stateValid && timeValid && !timeRollback {
		result.StateCommitAllowed = true
		if result.NextState.TrustedTimeFloor.IsZero() || input.Time.Now.After(result.NextState.TrustedTimeFloor) {
			result.NextState.TrustedTimeFloor = input.Time.Now.UTC()
		}
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		marketInput, present := input.Markets[market]
		if !present {
			continue
		}
		verdict := Verdict{Market: market, State: Unwired}
		switch {
		case !policyValid:
			verdict.Code = RefusalInvalid
		case !stateValid:
			verdict.Code = RefusalStateCorrupt
		case !timeValid:
			verdict.Code = RefusalTrustedTimeUnavailable
		case timeRollback:
			verdict.Code = RefusalTrustedTimeRollback
		default:
			verified, code := verifyAttestation(input.Policy, result.NextState, input.Time.Now, market, marketInput)
			verdict.Code = code
			if code == RefusalNone {
				verdict.State = Wired
				verdict.Provenance = Provenance{KeyID: verified.body.KeyID, Serial: verified.body.Serial, BodyDigest: verified.bodyDigest, BuildDigest: verified.body.BuildDigest,
					EvidenceDigest: verified.body.EvidenceDigest, IssuedAt: verified.issuedAt, ExpiresAt: verified.expiresAt}
				scope := serialScope{AccountID: verified.body.AccountID, ProfileID: verified.body.ProfileID, Market: verified.body.Market}
				result.NextState.Serials[scope] = verified.body.Serial
			}
		}
		if market == MarketKR {
			result.Snapshot.kr = verdict
		} else {
			result.Snapshot.us = verdict
		}
	}
	if result.StateCommitAllowed {
		result.NextState.seal = durableStateSeal(result.NextState)
		if result.NextState.seal != input.State.seal {
			result.Mutations = 1
		}
	} else {
		// Preserve the exact preimage. In particular, never repair or re-seal a
		// corrupt anti-rollback state into something a caller could commit.
		result.NextState = input.State
	}
	result.Snapshot.seal = readinessSnapshotSeal(result.Snapshot)
	return result
}
