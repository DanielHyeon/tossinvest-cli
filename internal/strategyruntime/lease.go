package strategyruntime

import (
	"crypto/sha256"
	"errors"
	"time"
)

type Lease struct {
	id                string
	operationID       string
	lineage           Lineage
	authority         authoritySnapshot
	owner             ownerFence
	issuedAt          time.Time
	expiresAt         time.Time
	revision          uint64
	state             LeaseState
	disposition       ReservationDisposition
	transportStarted  bool
	outcomeCode       string
	brokerOrderID     string
	queryDigest       string
	outcomeObservedAt time.Time
	seal              [32]byte
}

type leaseInput struct {
	ID          string
	OperationID string
	Lineage     Lineage
	Authority   authoritySnapshot
	Owner       ownerFence
	IssuedAt    time.Time
	ExpiresAt   time.Time
}

func newLease(input leaseInput) (Lease, error) {
	if !validIdentity(input.ID) || !validIdentity(input.OperationID) || !validLineage(input.Lineage) || !validAuthority(input.Authority) || !validOwnerFence(input.Owner) ||
		input.IssuedAt.IsZero() || !input.ExpiresAt.After(input.IssuedAt) {
		return Lease{}, errors.New("strategyruntime: invalid lease preimage")
	}
	lease := Lease{id: input.ID, operationID: input.OperationID, lineage: input.Lineage, authority: input.Authority, owner: input.Owner,
		issuedAt: input.IssuedAt.UTC(), expiresAt: input.ExpiresAt.UTC(), revision: 1, state: LeaseIssued, disposition: ReservationReserved}
	lease.seal = leaseSeal(lease)
	return lease, nil
}

func (lease Lease) State() LeaseState {
	if !validLease(lease) {
		return LeaseInvalid
	}
	return lease.state
}

func (lease Lease) Disposition() ReservationDisposition {
	if !validLease(lease) {
		return ReservationInvalid
	}
	return lease.disposition
}

func (lease Lease) OperationID() string {
	if !validLease(lease) {
		return ""
	}
	return lease.operationID
}

func (lease Lease) ExpiresAt() time.Time {
	if !validLease(lease) {
		return time.Time{}
	}
	return lease.expiresAt
}

func (lease Lease) TransportStarted() bool { return validLease(lease) && lease.transportStarted }

func (lease Lease) Revision() uint64 {
	if !validLease(lease) {
		return 0
	}
	return lease.revision
}

func leaseSeal(lease Lease) [32]byte {
	hash := sha256.New()
	for _, value := range []string{lease.id, lease.operationID, string(lease.lineage.seal[:]), string(lease.authority.seal[:]), string(lease.owner.seal[:]),
		formatTime(lease.issuedAt), formatTime(lease.expiresAt), string(lease.state), string(lease.disposition), lease.outcomeCode, lease.brokerOrderID,
		lease.queryDigest, formatTime(lease.outcomeObservedAt)} {
		writeString(hash, value)
	}
	if lease.transportStarted {
		writeString(hash, "transport-started")
	}
	writeUint64(hash, lease.revision)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validLease(lease Lease) bool {
	if !validIdentity(lease.id) || !validIdentity(lease.operationID) || !validLineage(lease.lineage) || !validAuthority(lease.authority) || !validOwnerFence(lease.owner) ||
		lease.issuedAt.IsZero() || !lease.expiresAt.After(lease.issuedAt) || lease.revision == 0 || lease.seal != leaseSeal(lease) {
		return false
	}
	switch lease.state {
	case LeaseIssued, LeaseClaimed:
		return lease.disposition == ReservationReserved && !lease.transportStarted
	case LeaseSubmitting:
		return lease.disposition == ReservationReserved && lease.transportStarted
	case LeaseSubmitted:
		return lease.disposition == ReservationTransferred && lease.transportStarted
	case LeaseRefused:
		return lease.disposition == ReservationReleased
	case LeaseAmbiguous:
		return (lease.disposition == ReservationHeld || lease.disposition == ReservationReleased || lease.disposition == ReservationTransferred) && lease.transportStarted
	default:
		return false
	}
}

func transitionLease(lease Lease, state LeaseState, disposition ReservationDisposition) Lease {
	next := lease
	next.revision++
	next.state, next.disposition = state, disposition
	if state == LeaseSubmitting || state == LeaseSubmitted || state == LeaseAmbiguous {
		next.transportStarted = true
	}
	next.seal = leaseSeal(next)
	return next
}

type retryReservation struct {
	id          string
	disposition ReservationDisposition
	seal        [32]byte
}

func newRetryReservation(id string, disposition ReservationDisposition) (retryReservation, error) {
	if !validIdentity(id) || disposition != ReservationHeld {
		return retryReservation{}, errors.New("strategyruntime: retry reservation must be exact HELD")
	}
	reservation := retryReservation{id: id, disposition: disposition}
	reservation.seal = hashStrings(reservation.id, string(reservation.disposition))
	return reservation, nil
}

func validRetryReservation(reservation retryReservation) bool {
	return validIdentity(reservation.id) && reservation.disposition == ReservationHeld && reservation.seal == hashStrings(reservation.id, string(reservation.disposition))
}

type claimInput struct {
	CurrentAuthority authoritySnapshot
	CurrentOwner     ownerFence
	Time             trustedTime
	RetryReservation retryReservation
}

type submittingInput struct {
	CurrentAuthority authoritySnapshot
	CurrentOwner     ownerFence
	Time             trustedTime
}

type AtomicResult struct {
	Lease                  Lease
	Code                   RefusalCode
	BrokerRequests         uint64
	TransportAuthorized    bool
	CommitAllowed          bool
	StateTransitions       uint64
	OriginalUnchanged      bool
	RetryDisposition       ReservationDisposition
	RetryReservationID     string
	ReservationTransitions uint64
	ExpectedRevision       uint64
	NextRevision           uint64
}

func Claim(lease Lease, input claimInput) AtomicResult {
	if !validLease(lease) {
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	if lease.state != LeaseIssued {
		return replayResult(lease, input.RetryReservation)
	}
	if code := pretransportValidation(lease, input.CurrentAuthority, input.CurrentOwner, input.Time); code != RefusalNone {
		return atomicTransition(lease, LeaseRefused, ReservationReleased, code, false)
	}
	return atomicTransition(lease, LeaseClaimed, ReservationReserved, RefusalNone, false)
}

func BeginSubmitting(lease Lease, input submittingInput) AtomicResult {
	if !validLease(lease) {
		return AtomicResult{Lease: lease, Code: RefusalInvalid}
	}
	if terminalLeaseState(lease.state) {
		return replayResult(lease, retryReservation{})
	}
	if lease.state != LeaseClaimed {
		return atomicTransition(lease, LeaseRefused, ReservationReleased, RefusalInvalidTransition, false)
	}
	if code := pretransportValidation(lease, input.CurrentAuthority, input.CurrentOwner, input.Time); code != RefusalNone {
		return atomicTransition(lease, LeaseRefused, ReservationReleased, code, false)
	}
	return atomicTransition(lease, LeaseSubmitting, ReservationReserved, RefusalNone, true)
}

func pretransportValidation(lease Lease, current authoritySnapshot, owner ownerFence, observed trustedTime) RefusalCode {
	if !validTrustedTime(observed) || !validAuthority(current) || !validOwnerFence(owner) {
		return RefusalInvalid
	}
	if observed.now.Before(lease.issuedAt) || !observed.now.Before(lease.expiresAt) {
		return RefusalExpired
	}
	if owner != lease.owner {
		return RefusalStaleOwner
	}
	if current.accountID != lease.authority.accountID || current.market != lease.authority.market || current.symbol != lease.authority.symbol {
		return RefusalScopeMismatch
	}
	if current != lease.authority {
		return RefusalAuthorityDrift
	}
	return RefusalNone
}

func atomicTransition(lease Lease, state LeaseState, disposition ReservationDisposition, code RefusalCode, transport bool) AtomicResult {
	next := transitionLease(lease, state, disposition)
	return AtomicResult{Lease: next, Code: code, TransportAuthorized: transport, CommitAllowed: true, StateTransitions: 1, ExpectedRevision: lease.revision, NextRevision: next.revision}
}

func replayResult(lease Lease, retry retryReservation) AtomicResult {
	result := AtomicResult{Lease: lease, Code: RefusalReplay, OriginalUnchanged: true, ExpectedRevision: lease.revision, NextRevision: lease.revision}
	if terminalLeaseState(lease.state) {
		result.Code = RefusalTerminalReplay
	}
	if validRetryReservation(retry) {
		result.RetryDisposition = ReservationReleased
		result.RetryReservationID = retry.id
		result.ReservationTransitions = 1
		result.CommitAllowed = true
	}
	return result
}
