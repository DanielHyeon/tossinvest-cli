package strategyruntime

import (
	"crypto/sha256"
	"errors"
	"time"
)

type outcomeEvidenceInput struct {
	Outcome             BrokerOutcome
	OperationID         string
	BrokerOrderID       string
	OutcomeCode         string
	QueryDigest         string
	LookupDigest        string
	ResponseDigest      string
	ObservedAt          time.Time
	Authoritative       bool
	LookupComplete      bool
	AcceptanceKnown     bool
	Accepted            bool
	DefinitiveRejection bool
	FillQuantity        uint64
	PendingOrderFound   bool
	TerminalOrderFound  bool
}

type outcomeEvidence struct {
	outcome             BrokerOutcome
	operationID         string
	brokerOrderID       string
	outcomeCode         string
	queryDigest         string
	lookupDigest        string
	responseDigest      string
	observedAt          time.Time
	authoritative       bool
	lookupComplete      bool
	acceptanceKnown     bool
	accepted            bool
	definitiveRejection bool
	fillQuantity        uint64
	pendingOrderFound   bool
	terminalOrderFound  bool
	seal                [32]byte
}

func newOutcomeEvidence(input outcomeEvidenceInput) (outcomeEvidence, error) {
	if !validIdentity(input.OperationID) || !validIdentity(input.OutcomeCode) || !validOptionalIdentity(input.BrokerOrderID) || !validDigest(input.QueryDigest) || !validDigest(input.LookupDigest) ||
		!validDigest(input.ResponseDigest) || input.ObservedAt.IsZero() {
		return outcomeEvidence{}, errors.New("strategyruntime: incomplete outcome evidence")
	}
	if !validOutcomeProof(input) {
		return outcomeEvidence{}, errors.New("strategyruntime: contradictory outcome proof")
	}
	evidence := outcomeEvidence{outcome: input.Outcome, operationID: input.OperationID, brokerOrderID: input.BrokerOrderID,
		outcomeCode: input.OutcomeCode, queryDigest: input.QueryDigest, lookupDigest: input.LookupDigest, responseDigest: input.ResponseDigest,
		observedAt: input.ObservedAt.UTC(), authoritative: input.Authoritative, lookupComplete: input.LookupComplete,
		acceptanceKnown: input.AcceptanceKnown, accepted: input.Accepted, definitiveRejection: input.DefinitiveRejection,
		fillQuantity: input.FillQuantity, pendingOrderFound: input.PendingOrderFound, terminalOrderFound: input.TerminalOrderFound}
	evidence.seal = outcomeEvidenceSeal(evidence)
	return evidence, nil
}

func validOutcomeProof(input outcomeEvidenceInput) bool {
	orderFound := input.PendingOrderFound != input.TerminalOrderFound
	switch input.Outcome {
	case OutcomeAccepted:
		return input.Authoritative && input.LookupComplete && input.AcceptanceKnown && input.Accepted && !input.DefinitiveRejection &&
			input.BrokerOrderID != "" && orderFound
	case OutcomeDefinitiveRejected:
		return input.Authoritative && input.LookupComplete && input.AcceptanceKnown && !input.Accepted && input.DefinitiveRejection &&
			input.BrokerOrderID == "" && input.FillQuantity == 0 && !input.PendingOrderFound && !input.TerminalOrderFound
	case OutcomeAuthoritativeNotSubmitted:
		return input.Authoritative && input.LookupComplete && input.AcceptanceKnown && !input.Accepted && !input.DefinitiveRejection &&
			input.BrokerOrderID == "" && input.FillQuantity == 0 && !input.PendingOrderFound && !input.TerminalOrderFound
	case OutcomeTransportUncertain:
		return !input.Authoritative && !input.LookupComplete && !input.AcceptanceKnown && !input.Accepted && !input.DefinitiveRejection &&
			input.BrokerOrderID == "" && input.FillQuantity == 0 && !input.PendingOrderFound && !input.TerminalOrderFound
	default:
		return false
	}
}

func outcomeEvidenceSeal(evidence outcomeEvidence) [32]byte {
	hash := sha256.New()
	for _, value := range []string{string(evidence.outcome), evidence.operationID, evidence.brokerOrderID, evidence.outcomeCode,
		evidence.queryDigest, evidence.lookupDigest, evidence.responseDigest, formatTime(evidence.observedAt)} {
		writeString(hash, value)
	}
	for _, value := range []bool{evidence.authoritative, evidence.lookupComplete, evidence.acceptanceKnown, evidence.accepted,
		evidence.definitiveRejection, evidence.pendingOrderFound, evidence.terminalOrderFound} {
		if value {
			writeString(hash, "1")
		} else {
			writeString(hash, "0")
		}
	}
	writeUint64(hash, evidence.fillQuantity)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validOutcomeEvidence(evidence outcomeEvidence, operationID string) bool {
	input := outcomeEvidenceInput{Outcome: evidence.outcome, OperationID: evidence.operationID, BrokerOrderID: evidence.brokerOrderID,
		OutcomeCode: evidence.outcomeCode, QueryDigest: evidence.queryDigest, LookupDigest: evidence.lookupDigest, ResponseDigest: evidence.responseDigest,
		ObservedAt: evidence.observedAt, Authoritative: evidence.authoritative, LookupComplete: evidence.lookupComplete,
		AcceptanceKnown: evidence.acceptanceKnown, Accepted: evidence.accepted, DefinitiveRejection: evidence.definitiveRejection,
		FillQuantity: evidence.fillQuantity, PendingOrderFound: evidence.pendingOrderFound, TerminalOrderFound: evidence.terminalOrderFound}
	return evidence.operationID == operationID && validIdentity(evidence.operationID) && validIdentity(evidence.outcomeCode) && validOptionalIdentity(evidence.brokerOrderID) && evidence.seal == outcomeEvidenceSeal(evidence) &&
		validDigest(evidence.queryDigest) && validDigest(evidence.lookupDigest) && validDigest(evidence.responseDigest) && !evidence.observedAt.IsZero() && validOutcomeProof(input)
}

func ClassifySubmitting(lease Lease, evidence outcomeEvidence) AtomicResult {
	if !validLease(lease) {
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	if terminalLeaseState(lease.state) {
		return replayResult(lease, retryReservation{})
	}
	if lease.state != LeaseSubmitting {
		return atomicTransition(lease, LeaseRefused, ReservationReleased, RefusalInvalidTransition, false)
	}
	if !validOutcomeEvidence(evidence, lease.operationID) || evidence.observedAt.Before(lease.issuedAt) {
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	lease.outcomeCode, lease.brokerOrderID, lease.queryDigest, lease.outcomeObservedAt = evidence.outcomeCode, evidence.brokerOrderID, evidence.queryDigest, evidence.observedAt
	switch evidence.outcome {
	case OutcomeAccepted:
		return atomicTransition(lease, LeaseSubmitted, ReservationTransferred, RefusalNone, false)
	case OutcomeDefinitiveRejected:
		return atomicTransition(lease, LeaseRefused, ReservationReleased, RefusalBrokerRejected, false)
	case OutcomeAuthoritativeNotSubmitted:
		return atomicTransition(lease, LeaseRefused, ReservationReleased, RefusalNotSubmitted, false)
	default:
		return atomicTransition(lease, LeaseAmbiguous, ReservationHeld, RefusalTransportUncertain, false)
	}
}

type recoveryInput struct {
	Outcome outcomeEvidence
}

func RecoverCrash(lease Lease, input recoveryInput) AtomicResult {
	if !validLease(lease) {
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	switch lease.state {
	case LeaseClaimed:
		return atomicTransition(lease, LeaseRefused, ReservationReleased, RefusalPretransportCrash, false)
	case LeaseSubmitting:
		return ClassifySubmitting(lease, input.Outcome)
	default:
		if terminalLeaseState(lease.state) {
			return replayResult(lease, retryReservation{})
		}
		return AtomicResult{Lease: lease, Code: RefusalInvalidTransition}
	}
}

func ReconcileReservation(lease Lease, evidence outcomeEvidence) AtomicResult {
	if !validLease(lease) || lease.state != LeaseAmbiguous || lease.disposition != ReservationHeld || !validOutcomeEvidence(evidence, lease.operationID) || evidence.observedAt.Before(lease.issuedAt) {
		if validLease(lease) && terminalLeaseState(lease.state) {
			return replayResult(lease, retryReservation{})
		}
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	next := lease
	next.revision++
	next.outcomeCode, next.brokerOrderID, next.queryDigest, next.outcomeObservedAt = evidence.outcomeCode, evidence.brokerOrderID, evidence.queryDigest, evidence.observedAt
	code := RefusalNone
	switch evidence.outcome {
	case OutcomeAccepted:
		next.disposition = ReservationTransferred
	case OutcomeAuthoritativeNotSubmitted:
		next.disposition = ReservationReleased
		code = RefusalNotSubmitted
	default:
		return replayResult(lease, retryReservation{})
	}
	next.seal = leaseSeal(next)
	return AtomicResult{Lease: next, Code: code, CommitAllowed: true, StateTransitions: 1, ExpectedRevision: lease.revision, NextRevision: next.revision}
}

type recoveryCapabilityInput struct {
	Market             Market
	OperationID        string
	Generation         uint64
	AttestationSerial  uint64
	Digest             string
	ExpiresAt          time.Time
	ClientKeyForwarded bool
	ClientKeyEchoed    bool
	ExactLookup        bool
	UniqueIdentity     bool
	PendingQuery       bool
	TerminalQuery      bool
	CancelResultQuery  bool
	Dedup              bool
	IdempotentSameKey  bool
	MaximumAttempts    uint64
}

type recoveryCapability struct {
	market             Market
	operationID        string
	generation         uint64
	attestationSerial  uint64
	digest             string
	expiresAt          time.Time
	clientKeyForwarded bool
	clientKeyEchoed    bool
	exactLookup        bool
	uniqueIdentity     bool
	pendingQuery       bool
	terminalQuery      bool
	cancelResultQuery  bool
	dedup              bool
	idempotentSameKey  bool
	maximumAttempts    uint64
	seal               [32]byte
}

func newRecoveryCapability(input recoveryCapabilityInput) (recoveryCapability, error) {
	capability := recoveryCapability{market: input.Market, operationID: input.OperationID, generation: input.Generation, attestationSerial: input.AttestationSerial, digest: input.Digest, expiresAt: input.ExpiresAt.UTC(),
		clientKeyForwarded: input.ClientKeyForwarded, clientKeyEchoed: input.ClientKeyEchoed, exactLookup: input.ExactLookup, uniqueIdentity: input.UniqueIdentity,
		pendingQuery: input.PendingQuery, terminalQuery: input.TerminalQuery, cancelResultQuery: input.CancelResultQuery, dedup: input.Dedup, idempotentSameKey: input.IdempotentSameKey, maximumAttempts: input.MaximumAttempts}
	if !completeRecoveryCapability(capability) {
		return recoveryCapability{}, errors.New("strategyruntime: incomplete recovery capability")
	}
	capability.seal = recoveryCapabilitySeal(capability)
	return capability, nil
}

func completeRecoveryCapability(capability recoveryCapability) bool {
	return validMarket(capability.market) && validIdentity(capability.operationID) && capability.generation > 0 && capability.attestationSerial > 0 && validDigest(capability.digest) && !capability.expiresAt.IsZero() &&
		capability.clientKeyForwarded && capability.clientKeyEchoed && capability.exactLookup && capability.uniqueIdentity && capability.pendingQuery && capability.terminalQuery && capability.cancelResultQuery &&
		capability.dedup && capability.idempotentSameKey && capability.maximumAttempts > 0
}

func recoveryCapabilitySeal(capability recoveryCapability) [32]byte {
	hash := sha256.New()
	writeString(hash, string(capability.market))
	writeString(hash, capability.operationID)
	writeUint64(hash, capability.generation)
	writeUint64(hash, capability.attestationSerial)
	writeString(hash, capability.digest)
	writeString(hash, formatTime(capability.expiresAt))
	for _, value := range []bool{capability.clientKeyForwarded, capability.clientKeyEchoed, capability.exactLookup, capability.uniqueIdentity, capability.pendingQuery, capability.terminalQuery, capability.cancelResultQuery, capability.dedup, capability.idempotentSameKey} {
		if value {
			writeString(hash, "1")
		} else {
			writeString(hash, "0")
		}
	}
	writeUint64(hash, capability.maximumAttempts)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type ResubmitDecision struct {
	Lease          Lease
	Allowed        bool
	Code           RefusalCode
	OperationID    string
	NewLease       bool
	BrokerRequests uint64
}

func AssessResubmit(lease Lease, capability recoveryCapability, current authoritySnapshot, owner ownerFence, observed trustedTime, attempts uint64) ResubmitDecision {
	result := ResubmitDecision{Lease: lease}
	if !validLease(lease) {
		result.Code = RefusalInvalid
		return result
	}
	if lease.state != LeaseAmbiguous || lease.disposition != ReservationHeld {
		result.Code = RefusalInvalidTransition
		return result
	}
	if !validTrustedTime(observed) || !validAuthority(current) || !validOwnerFence(owner) {
		result.Code = RefusalCapabilityUnattested
		return result
	}
	if owner != lease.owner {
		result.Code = RefusalStaleOwner
		return result
	}
	if current.accountID != lease.authority.accountID || current.market != lease.authority.market || current.symbol != lease.authority.symbol {
		result.Code = RefusalScopeMismatch
		return result
	}
	if current != lease.authority {
		result.Code = RefusalAuthorityDrift
		return result
	}
	if !completeRecoveryCapability(capability) || capability.seal != recoveryCapabilitySeal(capability) || capability.market != lease.authority.market ||
		capability.operationID != lease.operationID || capability.generation != lease.authority.protection.Generation ||
		capability.attestationSerial != lease.authority.protection.AttestationSerial || capability.digest != lease.authority.protection.Digest || !observed.now.Before(capability.expiresAt) {
		result.Code = RefusalCapabilityUnattested
		return result
	}
	if attempts >= capability.maximumAttempts {
		result.Code = RefusalRetryExhausted
		return result
	}
	result.Allowed, result.OperationID = true, lease.operationID
	return result
}
