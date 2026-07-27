package journal

// issuance.go is atomic issuance: writing a decision and the reservations it
// takes in one transaction (design D1, task 0.2).
//
// # Why the two cannot be separate transactions
//
// The reservation's `decision_id` is a NOT NULL foreign key and
// checkDecisionReservable reads a persisted decision, so "reserve, then decide"
// is structurally impossible. That leaves "decide, then reserve" — and that
// ordering has a crash window: a decision is on disk, it is submittable, and
// the limit headroom it was supposed to consume was never taken. A process that
// dies in that window comes back with an authorisation nobody paid for.
//
// RecordDecisionAndReserve closes the window by making both writes one
// BEGIN IMMEDIATE. A reservation refusal rolls the decision back with it, so
// there is no state in which a submittable decision exists without its hold
// (risk-management: 예약이 거부되면 결정도 같은 트랜잭션에서 함께 롤백되어
// 제출 가능한 결정이 남지 않는다). The Gateway's HELD-reservation check on
// EXPOSURE_RAISING submissions enforces the same statement from the other side;
// this is the half that makes the state unreachable rather than merely refused.
//
// # The refusal reasons are part of the contract
//
// A caller has to be able to record *why* an issuance failed, and the four
// reasons mean different things operationally: LIMIT_REACHED is the account
// being full, SNAPSHOT_RECOLLECTION_EXHAUSTED is the broker snapshot never
// settling, VERSION_CONFLICT is a race with another decision, DECISION_EXPIRED
// is the 60-second TTL running out. IssueRefusalReason is the single mapping
// from the sentinel errors to those strings, so two call sites cannot classify
// the same failure differently.
//
// # What is not here
//
// No policy. The journal does not decide whether a decision *should* be issued
// — that is the Guardian chain (task 4.1), which calls this after it allows.
// This file's whole job is that the two rows arrive together or not at all.
//
// # Broker-behaviour claims
//
// None. The broker snapshot arrives as data the caller collected, outside the
// transaction, exactly as it does for Reserve.

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Issuance refusal reasons (risk-management: 예약 실패는 원인별 안정
// reason-code로 기록된다). Stable strings: they reach the alerting path and the
// operator's record of why an entry did not happen.
const (
	// IssueReasonLimitReached — an aggregate limit was reached. Reaching counts:
	// the reservation transaction refuses at >=, so this is "the account is
	// full", not "the account is nearly full".
	IssueReasonLimitReached = "LIMIT_REACHED"
	// IssueReasonRecollectionExhausted — the re-collection cap or budget was
	// spent without a snapshot that held. Every stale/superseded refusal that the
	// loop retried converges here at the end.
	IssueReasonRecollectionExhausted = "SNAPSHOT_RECOLLECTION_EXHAUSTED"
	// IssueReasonVersionConflict — a single-shot issuance found the snapshot
	// stale or the reservation ledger moved under it. With the re-collection loop
	// this is a retry; without one it is the caller's answer.
	IssueReasonVersionConflict = "VERSION_CONFLICT"
	// IssueReasonDecisionExpired — the decision's TTL passed before the
	// reservation was taken, which is what bounds the freshness of every chain
	// input the re-collection loop does not re-derive.
	IssueReasonDecisionExpired = "DECISION_EXPIRED"
)

// IssueRequest is one atomic issuance: the decision to persist and the
// reservations it takes against the account's aggregate limits.
//
// Reserve.DecisionID and Reserve.AccountRef may be left empty; they are filled
// from the decision. Setting them to anything else is refused rather than
// silently overridden — a reservation pointing at another decision is a bug
// worth stopping on, not a field worth correcting.
type IssueRequest struct {
	Decision DecisionRequest
	Reserve  ReserveRequest
}

// IssueResult is what one successful issuance wrote.
type IssueResult struct {
	Decision Decision
	// Version is the reservation ledger version after the commit, usable as the
	// next ObservedVersion without a re-read.
	Version      int64
	Reservations []Reservation
}

// RecordDecisionAndReserve persists a decision and its reservations in one
// BEGIN IMMEDIATE transaction.
//
// Every refusal — a stale snapshot, a superseded ledger, an expired decision, a
// reached limit — rolls back both writes. Nothing partial is ever committed and
// nothing is submitted: after a refusal the decision does not exist, so there
// is nothing for a later process to find and send.
//
// The broker snapshot behind the request is collected by the caller *before*
// this is called, for the reason Reserve documents: the journal holds a single
// connection, so a network round trip inside the transaction would block every
// mutation record in the process.
func (j *Journal) RecordDecisionAndReserve(ctx context.Context, req IssueRequest) (IssueResult, error) {
	dec, reserve, plan, err := req.build()
	if err != nil {
		return IssueResult{}, err
	}

	now := j.clk.Now().UTC()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE (DSN _txlock=immediate)
	if err != nil {
		return IssueResult{}, fmt.Errorf("journal: starting the issuance transaction for decision %s: %w",
			dec.ID, err)
	}
	// The rollback that makes the whole file true. It runs on every path that
	// does not reach Commit, including the limit refusal below.
	defer tx.Rollback()

	// The snapshot's age and the ledger version first: both are refusals that do
	// not need the decision to exist, and there is no reason to write a row this
	// transaction has already decided to throw away.
	if err := reservePrecheck(ctx, tx, reserve, plan, now); err != nil {
		return IssueResult{}, err
	}
	// The decision, inside the transaction. From here on the reservation's
	// foreign key is satisfiable, and every later refusal takes this row with it.
	if err := insertDecisionRow(ctx, tx, dec); err != nil {
		return IssueResult{}, err
	}
	out, err := reserveRows(ctx, tx, reserve, plan, now)
	if err != nil {
		return IssueResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return IssueResult{}, fmt.Errorf("journal: committing the issuance of decision %s: %w",
			dec.ID, err)
	}
	return IssueResult{Decision: dec, Version: out.Version, Reservations: out.Reservations}, nil
}

// CollectIssue produces one IssueRequest from a freshly collected broker
// snapshot. Attempt counts from 1.
//
// It is the caller's function and it is where the network round trip happens —
// outside the transaction, by construction, because this signature is the only
// way a snapshot reaches RecordDecisionAndReserveWithRecollection.
type CollectIssue func(ctx context.Context, attempt int) (IssueRequest, error)

// RecordDecisionAndReserveWithRecollection runs the issuance under the bounded
// re-collection loop.
//
// It re-collects only on the two errors that mean "the snapshot is no longer a
// description of the account". The chain is *not* re-run between attempts
// (design D1): mode and latch changes are caught by the EntryGate re-check at
// submission, and the decision's own expiry is the freshness bound on
// everything else — which is why a loop that outlives the TTL surfaces as
// DECISION_EXPIRED rather than as a successful issuance of stale reasoning. The
// default budget (10s) is under the implemented TTL (60s) for that reason.
//
// Each attempt is a complete transaction: a refused attempt leaves no decision
// behind, so the next attempt is not competing with the corpse of the last one.
func (j *Journal) RecordDecisionAndReserveWithRecollection(ctx context.Context, collect CollectIssue, policy RecollectPolicy) (IssueResult, error) {
	if collect == nil {
		return IssueResult{}, fmt.Errorf("%w: issuing needs a snapshot collector", ErrInvalidRequest)
	}
	return recollectLoop(j.clk, policy, func(attempt int) (IssueResult, error) {
		req, err := collect(ctx, attempt)
		if err != nil {
			return IssueResult{}, fmt.Errorf("journal: collecting the broker snapshot: %w", err)
		}
		return j.RecordDecisionAndReserve(ctx, req)
	})
}

// IssueRefusalReason maps an issuance failure to the stable reason-code
// risk-management names, reporting false for anything that is not one of the
// four (a malformed request, a database error — those are bugs or outages, and
// classifying them as a risk refusal would hide them).
//
// The order is not arbitrary. ErrRecollectionExhausted wraps the last
// stale/superseded refusal it retried, so it has to be tested before the
// version conflict it contains; otherwise a loop that ran out of attempts would
// report a race it recovered from three times.
func IssueRefusalReason(err error) (string, bool) {
	switch {
	case err == nil:
		return "", false
	case errors.Is(err, ErrReservationLimitExceeded):
		return IssueReasonLimitReached, true
	case errors.Is(err, ErrDecisionExpired):
		return IssueReasonDecisionExpired, true
	case errors.Is(err, ErrRecollectionExhausted):
		return IssueReasonRecollectionExhausted, true
	case errors.Is(err, ErrSnapshotStale), errors.Is(err, ErrSnapshotSuperseded):
		return IssueReasonVersionConflict, true
	default:
		return "", false
	}
}

// build validates the pairing and produces everything the transaction needs.
// Both halves are validated before the transaction opens, so a malformed
// request never takes the write lock.
func (r IssueRequest) build() (Decision, ReserveRequest, reservePlan, error) {
	dec, err := r.Decision.build()
	if err != nil {
		return Decision{}, ReserveRequest{}, reservePlan{}, err
	}

	reserve := r.Reserve
	if strings.TrimSpace(reserve.DecisionID) == "" {
		reserve.DecisionID = dec.ID
	}
	if strings.TrimSpace(reserve.AccountRef) == "" {
		reserve.AccountRef = dec.AccountRef
	}

	plan, err := reserve.build()
	if err != nil {
		return Decision{}, ReserveRequest{}, reservePlan{}, err
	}
	if plan.decisionID != dec.ID {
		return Decision{}, ReserveRequest{}, reservePlan{}, fmt.Errorf(
			"%w: the issuance writes decision %s but reserves against %s; one transaction issues one decision",
			ErrInvalidRequest, dec.ID, plan.decisionID)
	}
	if plan.accountRef != dec.AccountRef {
		return Decision{}, ReserveRequest{}, reservePlan{}, fmt.Errorf(
			"%w: decision %s is for account %q, the reservation is on %q",
			ErrInvalidRequest, dec.ID, dec.AccountRef, plan.accountRef)
	}
	return dec, reserve, plan, nil
}
