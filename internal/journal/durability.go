package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AttemptState is a MutationAttempt's lifecycle state.
//
// The order-execution spec fixes the vocabulary:
//
//	RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT)
//	         → terminal: CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT
//
// The legality table for transitions lives in lifecycle.go; this file provides the
// durable primitives the table is enforced on top of.
type AttemptState string

const (
	// StateRecorded means the intent and the attempt are durably on disk and
	// nothing has been sent yet. A RECORDED attempt found after a restart is
	// safe to close as NOT_DISPATCHED.
	StateRecorded AttemptState = "RECORDED"
	// StateDispatchStarted means the request is about to go / has gone to the
	// broker. Anything found in this state after a restart is IN_DOUBT.
	StateDispatchStarted AttemptState = "DISPATCH_STARTED"
	// StateAcked means the broker accepted the mutation and named an order.
	StateAcked AttemptState = "ACKED"
	// StateInDoubt means we cannot tell whether the broker executed it.
	StateInDoubt AttemptState = "IN_DOUBT"

	// StateConfirmed is the terminal success: the mutation exists at the broker.
	StateConfirmed AttemptState = "CONFIRMED"
	// StateNotDispatched is the terminal "never left the process".
	StateNotDispatched AttemptState = "NOT_DISPATCHED"
	// StateFailedConfirmed is the terminal "broker definitively rejected it".
	StateFailedConfirmed AttemptState = "FAILED_CONFIRMED"
	// StateUnresolvedInDoubt is the terminal "we could not prove either way";
	// it blocks the symbol until an operator resolves it.
	StateUnresolvedInDoubt AttemptState = "UNRESOLVED_IN_DOUBT"
)

// IsTerminal reports whether the state ends the attempt's lifecycle.
func (s AttemptState) IsTerminal() bool {
	switch s {
	case StateConfirmed, StateNotDispatched, StateFailedConfirmed, StateUnresolvedInDoubt:
		return true
	default:
		return false
	}
}

// MutationKind is the kind of broker mutation an attempt performs.
type MutationKind string

const (
	KindPlace  MutationKind = "PLACE"
	KindCancel MutationKind = "CANCEL"
	KindAmend  MutationKind = "AMEND"
)

func (k MutationKind) valid() bool {
	switch k {
	case KindPlace, KindCancel, KindAmend:
		return true
	default:
		return false
	}
}

var (
	// ErrInvalidRequest means the caller's request is not recordable. It is
	// returned before anything is written.
	ErrInvalidRequest = errors.New("journal: invalid request")

	// ErrIntentMutated means an intent id was reused with different immutable
	// fields. Intents are the audit anchor; they never change.
	ErrIntentMutated = errors.New("journal: intent fields differ from the recorded intent")

	// ErrIntentNotFound / ErrAttemptNotFound are the lookup misses.
	ErrIntentNotFound  = errors.New("journal: intent not found")
	ErrAttemptNotFound = errors.New("journal: mutation attempt not found")

	// ErrUnexpectedState means the attempt was not in the state the transition
	// requires — the record and the caller's belief disagree, so we refuse.
	ErrUnexpectedState = errors.New("journal: attempt is not in the expected state")
)

// StateError carries the detail of a refused transition.
type StateError struct {
	AttemptID string
	Want      []AttemptState
	Got       AttemptState
	To        AttemptState
}

func (e *StateError) Error() string {
	want := make([]string, 0, len(e.Want))
	for _, s := range e.Want {
		want = append(want, string(s))
	}
	return fmt.Sprintf("journal: attempt %s is %s, cannot move to %s (expected one of %s)",
		e.AttemptID, e.Got, e.To, strings.Join(want, "|"))
}

func (e *StateError) Unwrap() error { return ErrUnexpectedState }

// Intent is the immutable record of what the engine decided to do. Every field is
// written once, at Prepare time, and never updated.
//
// Quantity and Price are decimal strings, not floats: a broker's decimal quantity
// has no exact binary float representation, and this record is the audit anchor.
// An empty Price means "no price" (market order) and is stored as NULL.
type Intent struct {
	ID          string
	Market      string // "kr" | "us" (see internal/clock.Market)
	TradingDay  string // market-local trading day, YYYY-MM-DD
	AccountRef  string
	Symbol      string
	Side        string // "BUY" | "SELL"
	OrderType   string // "LIMIT" | "MARKET" | …
	TimeInForce string
	Quantity    string // decimal string
	Price       string // decimal string; "" for market orders
	Currency    string
	Source      string // what produced the intent
	Fingerprint string // IN_DOUBT matching key
	Notes       string
}

// PrepareRequest asks the journal to durably record an intent and one attempt
// against it.
//
// The second block of fields is the decision binding (design D1/D2, task 1.3).
// They are the public-contract extension this change makes: an attempt records
// *which decision entitled it*, and the exact bytes and key that decision
// authorised, at RECORDED — before anything is sent. Written once and never
// updated: transition() only ever touches state, timestamps, reason and the
// broker order id, so these columns are immutable by construction.
type PrepareRequest struct {
	// Intent is the immutable intent. When the id already exists, every field
	// must match what was recorded or Prepare fails with ErrIntentMutated.
	Intent Intent
	// Kind is PLACE, CANCEL or AMEND.
	Kind MutationKind
	// AttemptID is a caller-supplied stable id. The caller owns id generation so
	// a retry after an ambiguous crash can be recognised rather than duplicated.
	AttemptID string
	// TargetOrderID is the broker order a CANCEL/AMEND acts on. Required for
	// those kinds.
	TargetOrderID string
	// Fingerprint overrides the intent's fingerprint for this attempt. Empty
	// inherits Intent.Fingerprint.
	Fingerprint string

	// AccountRef is the account the *decision* was issued for. It is recorded on
	// the attempt (the partial UNIQUE on the idempotency key needs the column on
	// this table — see issues.md) and must equal Intent.AccountRef: a decision
	// for one account may not authorise an attempt on another. Empty inherits
	// the intent's account.
	AccountRef string
	// DecisionID references the decisions row the issuer wrote before calling
	// the gateway. The row must already exist — the schema's foreign key is what
	// makes "a decision that was never persisted cannot be executed" structural
	// rather than a convention.
	DecisionID string
	// SafetyClass is the class that decision carries (EXPOSURE_RAISING,
	// RISK_REDUCING, PROTECTION_WEAKENING). Meaningless without a decision, so
	// it is refused without one: a class nothing backs is a caller's claim.
	SafetyClass string
	// Generation is the decision's reissue counter, always 0 in this change.
	Generation int
	// ClientOrderID is the broker idempotency key f(decision_id, generation).
	// PLACE only: the cancel and modify endpoints take no key (openapi —
	// clientOrderId exists on OrderCreateRequest and on the conditional-order
	// request, nowhere else).
	ClientOrderID string
	// WireBody is the canonical request body as it will be sent, stored verbatim.
	// A replay resends these bytes; it never rebuilds a body from structured
	// fields, because a change in the serialisation rules would then send
	// different bytes under the same idempotency key (design D2/D3).
	WireBody string
	// SerializerVersion names the rules WireBody was produced under, so a later
	// build can tell whether it still understands the stored bytes. Required
	// with a body and meaningless without one.
	SerializerVersion string
}

func (r PrepareRequest) validate() error {
	missing := func(field string) error {
		return fmt.Errorf("%w: %s is required", ErrInvalidRequest, field)
	}
	switch {
	case strings.TrimSpace(r.Intent.ID) == "":
		return missing("intent id")
	case strings.TrimSpace(r.AttemptID) == "":
		return missing("attempt id")
	case strings.TrimSpace(r.Intent.Market) == "":
		return missing("intent market")
	case strings.TrimSpace(r.Intent.TradingDay) == "":
		return missing("intent trading day")
	case strings.TrimSpace(r.Intent.AccountRef) == "":
		return missing("intent account ref")
	case strings.TrimSpace(r.Intent.Symbol) == "":
		return missing("intent symbol")
	case strings.TrimSpace(r.Intent.Side) == "":
		return missing("intent side")
	case strings.TrimSpace(r.Intent.OrderType) == "":
		return missing("intent order type")
	case strings.TrimSpace(r.Intent.Quantity) == "":
		return missing("intent quantity")
	case strings.TrimSpace(r.Intent.Currency) == "":
		return missing("intent currency")
	case strings.TrimSpace(r.Intent.Source) == "":
		return missing("intent source")
	case strings.TrimSpace(r.Intent.Fingerprint) == "":
		return missing("intent fingerprint")
	case !r.Kind.valid():
		return fmt.Errorf("%w: unknown mutation kind %q", ErrInvalidRequest, r.Kind)
	case r.Kind != KindPlace && strings.TrimSpace(r.TargetOrderID) == "":
		return fmt.Errorf("%w: %s requires a target order id", ErrInvalidRequest, r.Kind)
	}
	return r.validateDecisionBinding()
}

// validateDecisionBinding checks the decision half of the request.
//
// Every branch here refuses a *shape*, never a risk judgement: whether the
// decision itself authorises the mutation is the gateway's re-verification
// against the persisted row, and the journal has no business duplicating it.
// What the journal owns is that the row it writes cannot lie about which
// decision, which account and which key the attempt belongs to.
func (r PrepareRequest) validateDecisionBinding() error {
	decision := strings.TrimSpace(r.DecisionID)
	class := strings.TrimSpace(r.SafetyClass)
	key := strings.TrimSpace(r.ClientOrderID)

	if acct := strings.TrimSpace(r.AccountRef); acct != "" && acct != strings.TrimSpace(r.Intent.AccountRef) {
		return fmt.Errorf("%w: the decision was issued for account %q but the intent is on %q",
			ErrInvalidRequest, acct, r.Intent.AccountRef)
	}
	if decision == "" {
		switch {
		case class != "":
			return fmt.Errorf("%w: a safety class without a decision is not an authorisation",
				ErrInvalidRequest)
		case key != "":
			return fmt.Errorf("%w: an idempotency key without a decision is not derivable",
				ErrInvalidRequest)
		case r.Generation != 0:
			return fmt.Errorf("%w: a generation without a decision names nothing", ErrInvalidRequest)
		}
	} else {
		switch {
		case !validSafetyClass(class):
			return fmt.Errorf("%w: safety class %q is not one this build issues", ErrInvalidRequest, class)
		case r.Generation < 0:
			return fmt.Errorf("%w: generation %d is negative", ErrInvalidRequest, r.Generation)
		case key != "" && !ValidClientOrderID(key):
			return fmt.Errorf("%w: idempotency key %q is outside the broker's ^[a-zA-Z0-9\\-_]+$ (max 36)",
				ErrInvalidRequest, key)
		case key != "" && r.Kind != KindPlace:
			return fmt.Errorf("%w: %s carries no idempotency key — only order creation does",
				ErrInvalidRequest, r.Kind)
		}
	}

	body := strings.TrimSpace(r.WireBody)
	version := strings.TrimSpace(r.SerializerVersion)
	switch {
	case body != "" && version == "":
		return fmt.Errorf("%w: a stored wire body without a serializer version cannot be replayed safely",
			ErrInvalidRequest)
	case version != "" && body == "":
		return fmt.Errorf("%w: a serializer version without a wire body describes nothing",
			ErrInvalidRequest)
	}
	return nil
}

// validSafetyClass reports whether the class is one the schema's CHECK accepts.
// PROTECTION_WEAKENING is included because the column accepts it; no writer in
// this change produces one (design D4).
func validSafetyClass(class string) bool {
	switch class {
	case SafetyClassExposureRaising, SafetyClassRiskReducing, SafetyClassProtectionWeakening:
		return true
	default:
		return false
	}
}

// AttemptRecord is the stored form of a mutation attempt.
type AttemptRecord struct {
	ID                string
	IntentID          string
	Kind              MutationKind
	State             AttemptState
	AttemptNo         int
	TargetOrderID     string
	BrokerOrderID     string
	Fingerprint       string
	RecordedAt        string // RFC3339 UTC
	DispatchStartedAt string
	SettledAt         string
	ReasonCode        string
	Detail            string

	// The decision binding (task 1.3). Empty on every attempt recorded before
	// this schema, and on any path that submits without a decision.
	AccountRef        string
	DecisionID        string
	SafetyClass       string
	Generation        int
	ClientOrderID     string
	WireBody          string
	SerializerVersion string
	// ReplayCount and LastReplayAt belong to the idempotent replay procedure
	// (section 2). They are read here so the record is complete; nothing in this
	// change writes them.
	ReplayCount  int
	LastReplayAt string
}

// Attempt is a handle on a durably recorded attempt.
//
// It is deliberately the *only* thing Prepare returns, and Prepare only returns it
// after the RECORDED transaction committed. The execution gateway needs a handle to
// dispatch, so "submit before the journal commit" is not something a caller can
// express: there is nothing to dispatch with until the write is on disk.
type Attempt struct {
	j         *Journal
	id        string
	intentID  string
	kind      MutationKind
	attemptNo int

	mu    sync.Mutex
	state AttemptState
}

// ID returns the attempt id.
func (a *Attempt) ID() string { return a.id }

// IntentID returns the intent this attempt belongs to.
func (a *Attempt) IntentID() string { return a.intentID }

// Kind returns PLACE, CANCEL or AMEND.
func (a *Attempt) Kind() MutationKind { return a.kind }

// AttemptNo returns the 1-based attempt counter within the intent.
func (a *Attempt) AttemptNo() int { return a.attemptNo }

// State returns the last state this handle observed.
func (a *Attempt) State() AttemptState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

// Prepare records the intent and a RECORDED attempt in a single BEGIN IMMEDIATE
// transaction on a connection with synchronous=FULL, so the commit is fsynced
// before it returns.
//
// It returns an error, and no handle, if anything went wrong — which is the
// spec's "WHEN the journal transaction fails to commit THEN no broker submission
// happens" expressed as a type: without a handle there is nothing to dispatch.
func (j *Journal) Prepare(ctx context.Context, req PrepareRequest) (*Attempt, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	fingerprint := req.Fingerprint
	if fingerprint == "" {
		fingerprint = req.Intent.Fingerprint
	}
	now := j.nowString()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE (DSN _txlock=immediate)
	if err != nil {
		return nil, fmt.Errorf("journal: starting the intent transaction: %w", err)
	}
	defer tx.Rollback()

	existing, err := scanIntent(tx.QueryRowContext(ctx, intentSelect+" WHERE id = ?", req.Intent.ID))
	switch {
	case err == nil:
		if existing != req.Intent {
			return nil, fmt.Errorf("%w: intent %s", ErrIntentMutated, req.Intent.ID)
		}
	case errors.Is(err, ErrIntentNotFound):
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO intents
			   (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
			    time_in_force, quantity, price, currency, source, fingerprint, notes)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			req.Intent.ID, now, req.Intent.Market, req.Intent.TradingDay, req.Intent.AccountRef,
			req.Intent.Symbol, req.Intent.Side, req.Intent.OrderType, req.Intent.TimeInForce,
			req.Intent.Quantity, nullableString(req.Intent.Price), req.Intent.Currency,
			req.Intent.Source, req.Intent.Fingerprint, req.Intent.Notes); err != nil {
			return nil, fmt.Errorf("journal: recording intent %s: %w", req.Intent.ID, err)
		}
	default:
		return nil, err
	}

	// The binding is checked against the decision row inside this transaction.
	// The foreign key already refuses an attempt that names no decision at all;
	// this refuses one that names a decision saying something else, which is the
	// half a foreign key cannot see.
	if err := checkDecisionBinding(ctx, tx, req); err != nil {
		return nil, err
	}

	var attemptNo int
	if err := tx.QueryRowContext(ctx,
		"SELECT coalesce(max(attempt_no), 0) + 1 FROM mutation_attempts WHERE intent_id = ?",
		req.Intent.ID).Scan(&attemptNo); err != nil {
		return nil, fmt.Errorf("journal: numbering the attempt for intent %s: %w", req.Intent.ID, err)
	}

	// account_ref is always written, even for an attempt with no decision: it is
	// what the partial UNIQUE on the idempotency key is scoped by, and an attempt
	// whose account is only reachable through the intent's foreign key cannot be
	// covered by an index (issues.md, task 0.1 → 1.3).
	accountRef := strings.TrimSpace(req.AccountRef)
	if accountRef == "" {
		accountRef = strings.TrimSpace(req.Intent.AccountRef)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, target_order_id, broker_order_id,
		    fingerprint, recorded_at, account_ref, decision_id, safety_class, generation,
		    client_order_id, wire_body, serializer_version)
		 VALUES (?,?,?,?,?,?,'',?,?,?,?,?,?,?,?,?)`,
		req.AttemptID, req.Intent.ID, string(req.Kind), string(StateRecorded), attemptNo,
		req.TargetOrderID, fingerprint, now, accountRef,
		nullableString(strings.TrimSpace(req.DecisionID)),
		nullableString(strings.TrimSpace(req.SafetyClass)),
		nullableGeneration(req.DecisionID, req.Generation),
		nullableString(strings.TrimSpace(req.ClientOrderID)),
		nullableString(req.WireBody),
		nullableString(strings.TrimSpace(req.SerializerVersion))); err != nil {
		return nil, fmt.Errorf("journal: recording attempt %s: %w", req.AttemptID, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO attempt_transitions (attempt_id, from_state, to_state, at)
		 VALUES (?, '', ?, ?)`,
		req.AttemptID, string(StateRecorded), now); err != nil {
		return nil, fmt.Errorf("journal: recording the RECORDED transition for %s: %w", req.AttemptID, err)
	}

	// Everything below this line depends on the commit having returned without
	// error: that is the fsync the spec requires before dispatch.
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("journal: committing intent %s (nothing was submitted): %w",
			req.Intent.ID, err)
	}

	return &Attempt{
		j:         j,
		id:        req.AttemptID,
		intentID:  req.Intent.ID,
		kind:      req.Kind,
		attemptNo: attemptNo,
		state:     StateRecorded,
	}, nil
}

// MarkDispatchStarted records that the request is going out, and spends the
// decision's one-shot nonce in the same transaction.
//
// The two are one write because they are one fact: "this authorisation is now
// being used". A separate write would add a second commit to the hot path and,
// worse, would have a window in which the dispatch is recorded and the nonce is
// not — which is a decision that can be submitted twice after a crash
// (engine-safety: "소비 기록은 전송 시작 기록과 같은 트랜잭션에서 남긴다").
//
// It must commit before the broker call, so a crash between the two is
// discoverable as IN_DOUBT. An attempt with no decision spends nothing.
func (a *Attempt) MarkDispatchStarted(ctx context.Context) error {
	return a.transition(ctx, StateDispatchStarted, transitionOpts{
		from:               []AttemptState{StateRecorded},
		setDispatchStarted: true,
		consumeNonce:       true,
	})
}

// MarkAcked records the broker's acknowledgement and the order id it named.
func (a *Attempt) MarkAcked(ctx context.Context, brokerOrderID string) error {
	return a.transition(ctx, StateAcked, transitionOpts{
		from:          []AttemptState{StateDispatchStarted},
		brokerOrderID: brokerOrderID,
	})
}

// MarkInDoubt records that the outcome is unknown. It is not terminal: the
// resolution procedure has to finish before the attempt can be settled.
func (a *Attempt) MarkInDoubt(ctx context.Context, reasonCode, detail string) error {
	return a.transition(ctx, StateInDoubt, transitionOpts{
		from:       []AttemptState{StateDispatchStarted, StateAcked},
		reasonCode: reasonCode,
		detail:     detail,
	})
}

// Settle closes the attempt in a terminal state.
func (a *Attempt) Settle(ctx context.Context, state AttemptState, reasonCode, detail string) error {
	if !state.IsTerminal() {
		return fmt.Errorf("%w: %s is not a terminal state", ErrInvalidRequest, state)
	}
	return a.transition(ctx, state, transitionOpts{
		from:       []AttemptState{StateRecorded, StateDispatchStarted, StateAcked, StateInDoubt},
		reasonCode: reasonCode,
		detail:     detail,
		setSettled: true,
	})
}

type transitionOpts struct {
	// from is the set of states the transition is legal from. Empty means "any
	// non-terminal state".
	from               []AttemptState
	reasonCode         string
	detail             string
	brokerOrderID      string
	setDispatchStarted bool
	setSettled         bool
	// consumeNonce spends the decision's nonce in the same transaction. Only the
	// dispatch transition sets it.
	consumeNonce bool
}

// transition updates the attempt row and appends the transition history in one
// BEGIN IMMEDIATE transaction. The from-state is checked inside that transaction,
// so two writers cannot both believe they made the same move.
func (a *Attempt) transition(ctx context.Context, to AttemptState, o transitionOpts) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.j.nowString()
	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("journal: starting the transition transaction for %s: %w", a.id, err)
	}
	defer tx.Rollback()

	var current AttemptState
	if err := tx.QueryRowContext(ctx,
		"SELECT state FROM mutation_attempts WHERE id = ?", a.id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s", ErrAttemptNotFound, a.id)
		}
		return fmt.Errorf("journal: reading attempt %s: %w", a.id, err)
	}
	if err := checkTransitionAllowed(a.id, current, to, o.from); err != nil {
		return err
	}
	if o.consumeNonce {
		if err := consumeDecisionNonce(ctx, tx, a.id, now); err != nil {
			return err
		}
	}

	set := []string{"state = ?"}
	args := []any{string(to)}
	if o.setDispatchStarted {
		set = append(set, "dispatch_started_at = ?")
		args = append(args, now)
	}
	if o.setSettled {
		set = append(set, "settled_at = ?")
		args = append(args, now)
	}
	if o.brokerOrderID != "" {
		set = append(set, "broker_order_id = ?")
		args = append(args, o.brokerOrderID)
	}
	if o.reasonCode != "" {
		set = append(set, "reason_code = ?")
		args = append(args, o.reasonCode)
	}
	if o.detail != "" {
		set = append(set, "detail = ?")
		args = append(args, o.detail)
	}
	args = append(args, a.id, string(current))

	res, err := tx.ExecContext(ctx,
		"UPDATE mutation_attempts SET "+strings.Join(set, ", ")+" WHERE id = ? AND state = ?", args...)
	if err != nil {
		return fmt.Errorf("journal: updating attempt %s: %w", a.id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: updating attempt %s: %w", a.id, err)
	}
	if affected != 1 {
		// Someone changed the row between our read and our write inside the same
		// transaction — should be impossible with BEGIN IMMEDIATE, so treat it as
		// a hard state error rather than retrying.
		return &StateError{AttemptID: a.id, Want: o.from, Got: current, To: to}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO attempt_transitions (attempt_id, from_state, to_state, at, reason_code, detail)
		 VALUES (?,?,?,?,?,?)`,
		a.id, string(current), string(to), now, o.reasonCode, o.detail); err != nil {
		return fmt.Errorf("journal: recording the %s->%s transition for %s: %w", current, to, a.id, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: committing the %s->%s transition for %s: %w", current, to, a.id, err)
	}

	a.state = to
	return nil
}

// checkTransitionAllowed validates a move against the caller-declared from-set.
// lifecycle.go layers the spec's full legality table on top of this.
func checkTransitionAllowed(attemptID string, from, to AttemptState, allowed []AttemptState) error {
	if err := ValidateTransition(from, to); err != nil {
		return fmt.Errorf("%w (attempt %s)", err, attemptID)
	}
	if len(allowed) == 0 {
		return nil
	}
	for _, s := range allowed {
		if s == from {
			return nil
		}
	}
	return &StateError{AttemptID: attemptID, Want: allowed, Got: from, To: to}
}

// LookupIntent returns a recorded intent.
func (j *Journal) LookupIntent(ctx context.Context, id string) (Intent, error) {
	return scanIntent(j.db.QueryRowContext(ctx, intentSelect+" WHERE id = ?", id))
}

const intentSelect = `SELECT id, market, trading_day, account_ref, symbol, side, order_type,
	time_in_force, quantity, price, currency, source, fingerprint, notes FROM intents`

func scanIntent(row *sql.Row) (Intent, error) {
	var (
		in    Intent
		price sql.NullString
	)
	err := row.Scan(&in.ID, &in.Market, &in.TradingDay, &in.AccountRef, &in.Symbol, &in.Side,
		&in.OrderType, &in.TimeInForce, &in.Quantity, &price, &in.Currency, &in.Source,
		&in.Fingerprint, &in.Notes)
	if errors.Is(err, sql.ErrNoRows) {
		return Intent{}, ErrIntentNotFound
	}
	if err != nil {
		return Intent{}, fmt.Errorf("journal: reading intent: %w", err)
	}
	in.Price = price.String
	return in, nil
}

// LookupAttempt returns a recorded mutation attempt.
func (j *Journal) LookupAttempt(ctx context.Context, id string) (AttemptRecord, error) {
	return scanAttempt(j.db.QueryRowContext(ctx, attemptSelect+" WHERE id = ?", id))
}

const attemptSelect = `SELECT id, intent_id, kind, state, attempt_no, target_order_id,
	broker_order_id, fingerprint, recorded_at, dispatch_started_at, settled_at,
	reason_code, detail, account_ref, decision_id, safety_class, generation,
	client_order_id, wire_body, serializer_version, replay_count, last_replay_at
	FROM mutation_attempts`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttempt(row rowScanner) (AttemptRecord, error) {
	var (
		rec                        AttemptRecord
		dispatchStarted, settledAt sql.NullString
		accountRef, decisionID     sql.NullString
		safetyClass, clientOrderID sql.NullString
		wireBody, serializer       sql.NullString
		lastReplayAt               sql.NullString
		generation, replayCount    sql.NullInt64
	)
	err := row.Scan(&rec.ID, &rec.IntentID, &rec.Kind, &rec.State, &rec.AttemptNo,
		&rec.TargetOrderID, &rec.BrokerOrderID, &rec.Fingerprint, &rec.RecordedAt,
		&dispatchStarted, &settledAt, &rec.ReasonCode, &rec.Detail,
		&accountRef, &decisionID, &safetyClass, &generation,
		&clientOrderID, &wireBody, &serializer, &replayCount, &lastReplayAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AttemptRecord{}, ErrAttemptNotFound
	}
	if err != nil {
		return AttemptRecord{}, fmt.Errorf("journal: reading attempt: %w", err)
	}
	rec.DispatchStartedAt = dispatchStarted.String
	rec.SettledAt = settledAt.String
	rec.AccountRef = accountRef.String
	rec.DecisionID = decisionID.String
	rec.SafetyClass = safetyClass.String
	rec.Generation = int(generation.Int64)
	rec.ClientOrderID = clientOrderID.String
	rec.WireBody = wireBody.String
	rec.SerializerVersion = serializer.String
	rec.ReplayCount = int(replayCount.Int64)
	rec.LastReplayAt = lastReplayAt.String
	return rec, nil
}

// nowString formats the injected clock's instant the way the journal stores it.
func (j *Journal) nowString() string {
	return j.clk.Now().UTC().Format(time.RFC3339)
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullableGeneration stores the generation only for an attempt that has a
// decision. Zero is a meaningful generation (it is the only one this change
// issues), so "0 or absent" cannot be expressed by the value alone.
func nullableGeneration(decisionID string, generation int) any {
	if strings.TrimSpace(decisionID) == "" {
		return nil
	}
	return generation
}
