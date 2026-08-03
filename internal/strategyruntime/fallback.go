package strategyruntime

import (
	"errors"
	"time"
)

type CentralFaultKind string

const (
	FaultJournalCorrupt    CentralFaultKind = "JOURNAL_CORRUPT"
	FaultGatewayInvariant  CentralFaultKind = "GATEWAY_INVARIANT"
	FaultOwnerFenceCorrupt CentralFaultKind = "OWNER_FENCE_CORRUPT"
	FaultMultipleOwners    CentralFaultKind = "MULTIPLE_CURRENT_OWNERS"
)

func validCentralFaultKind(kind CentralFaultKind) bool {
	switch kind {
	case FaultJournalCorrupt, FaultGatewayInvariant, FaultOwnerFenceCorrupt, FaultMultipleOwners:
		return true
	default:
		return false
	}
}

type safetyFallbackManifestInput struct {
	Release string
	RTO     time.Duration
}

type safetyFallbackManifest struct {
	release string
	rto     time.Duration
	seal    [32]byte
}

func newSafetyFallbackManifest(input safetyFallbackManifestInput) (safetyFallbackManifest, error) {
	if input.Release != RuntimeRelease || input.RTO <= 0 || input.RTO > 60*time.Second {
		return safetyFallbackManifest{}, errors.New("strategyruntime: fallback RTO must be within 60 seconds")
	}
	manifest := safetyFallbackManifest{release: input.Release, rto: input.RTO}
	manifest.seal = safetyFallbackManifestSeal(manifest)
	return manifest, nil
}

func safetyFallbackManifestSeal(manifest safetyFallbackManifest) [32]byte {
	return hashStrings(manifest.release, manifest.rto.String())
}

type centralFaultInput struct {
	Kind           CentralFaultKind
	DetectedAt     time.Time
	EvidenceDigest string
	CurrentOwner   ownerFence
}

type centralFault struct {
	kind           CentralFaultKind
	detectedAt     time.Time
	evidenceDigest string
	currentOwner   ownerFence
	seal           [32]byte
}

func newCentralFault(input centralFaultInput) (centralFault, error) {
	if !validCentralFaultKind(input.Kind) || input.DetectedAt.IsZero() || !validDigest(input.EvidenceDigest) || !validOwnerFence(input.CurrentOwner) {
		return centralFault{}, errors.New("strategyruntime: invalid central fault evidence")
	}
	fault := centralFault{kind: input.Kind, detectedAt: input.DetectedAt.UTC(), evidenceDigest: input.EvidenceDigest, currentOwner: input.CurrentOwner}
	fault.seal = centralFaultSeal(fault)
	return fault, nil
}

func centralFaultSeal(fault centralFault) [32]byte {
	return hashStrings(string(fault.kind), formatTime(fault.detectedAt), fault.evidenceDigest, string(fault.currentOwner.seal[:]))
}

type FallbackStatus string

const (
	FallbackStarted     FallbackStatus = "SAFETY_FALLBACK_STARTED"
	FallbackUnavailable FallbackStatus = "SAFETY_FALLBACK_UNAVAILABLE"
)

type FallbackPlan struct {
	Status                   FallbackStatus
	EntryAllowed             bool
	LeaseIssuanceAllowed     bool
	Safety                   SafetyState
	RTO                      time.Duration
	OldOwnerFenced           bool
	CriticalAlert            bool
	PreserveBrokerProtection bool
}

func PlanSafetyFallback(manifest safetyFallbackManifest, fault centralFault, replacement ownerFence, observed trustedTime) FallbackPlan {
	plan := FallbackPlan{Status: FallbackUnavailable, PreserveBrokerProtection: true, CriticalAlert: true}
	if manifest.seal != safetyFallbackManifestSeal(manifest) || manifest.release != RuntimeRelease || manifest.rto <= 0 || manifest.rto > 60*time.Second {
		return plan
	}
	plan.RTO = manifest.rto
	if fault.seal != centralFaultSeal(fault) || !validCentralFaultKind(fault.kind) || !validTrustedTime(observed) || !validOwnerFence(replacement) {
		return plan
	}
	if replacement.epoch <= fault.currentOwner.epoch || replacement.token == fault.currentOwner.token {
		return plan
	}
	if observed.now.Before(fault.detectedAt) || observed.now.After(fault.detectedAt.Add(manifest.rto)) {
		return plan
	}
	plan.Status, plan.OldOwnerFenced, plan.CriticalAlert, plan.Safety = FallbackStarted, true, false, fullSafetyState()
	return plan
}
