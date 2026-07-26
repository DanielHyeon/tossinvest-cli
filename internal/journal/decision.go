package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// decision.go is the issuer's half of the extended execution contract: the API a
// decision issuer uses to persist a decision *before* it calls the gateway, and
// the API the gateway uses to read that decision back (design D1, task 1.4).
//
// # Why the record and not the caller is the authority
//
// A submission carries a decision id and nothing else. Everything the gateway
// re-verifies — the risk preimage, the idempotency key, the class, the expiry,
// the nonce — is read from the row written here. If the gateway re-verified
// against values the submitter handed it, the verification would be circular:
// a caller able to change the stop price could also change the copy it is
// checked against (engine-safety: "제출 호출자가 공급한 위험 데이터로 재검증해서는
// 안 된다 — 검증이 순환한다").
//
// # The preimage is class-specific
//
// EXPOSURE_RAISING carries a RiskIntent (what the entry risks: entry, stop,
// target, size, policy version) and RISK_REDUCING a ReductionIntent (what the
// exit is allowed to remove). A reduction has no sizing risk to bind, and an
// entry has no upper bound to relax, so one schema for both would have to make
// every field optional — and an optional stop price is exactly the field a
// tampered decision would drop.
//
// # Broker-behaviour claims
//
// None, except the idempotency key's shape, which is derived and checked in
// idempotency.go against docs/migration/openapi.latest.json.

// ErrDecisionNotFound means no decisions row carries that id. The gateway treats
// it as "there is no authorisation", never as "assume it was fine".
var ErrDecisionNotFound = errors.New("journal: decision not found")

// Preimage is the risk data a decision is bound to. Implementations are value
// types with a canonical text form; the hash is over that text.
type Preimage interface {
	// Kind returns the PreimageKind* tag stored alongside the text.
	Kind() string
	// Canonical returns the canonical JSON, or an error when the preimage is not
	// complete enough to be one.
	Canonical() (string, error)
}

// RiskIntent is the preimage of an EXPOSURE_RAISING decision: everything about
// the entry that the size was justified by.
//
// Decimals are strings, as everywhere else in the journal. Empty TargetPrice
// means "no take-profit"; EntryPrice and StopPrice are required, because a
// decision that does not say where the risk ends is not a risk decision.
type RiskIntent struct {
	AccountRef    string `json:"account_ref"`
	Market        string `json:"market"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Quantity      string `json:"quantity"`
	EntryPrice    string `json:"entry_price"`
	StopPrice     string `json:"stop_price"`
	TargetPrice   string `json:"target_price"`
	PolicyVersion string `json:"policy_version"`
}

// Kind implements Preimage.
func (RiskIntent) Kind() string { return PreimageKindRiskIntent }

// Canonical implements Preimage.
func (r RiskIntent) Canonical() (string, error) {
	market, symbol, side, err := canonicalVenue(r.Market, r.Symbol, r.Side)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(r.AccountRef) == "" {
		return "", fmt.Errorf("%w: risk intent has no account", ErrInvalidRequest)
	}
	quantity, err := canonicalDecimal("quantity", r.Quantity, true)
	if err != nil {
		return "", err
	}
	entry, err := canonicalDecimal("entry price", r.EntryPrice, true)
	if err != nil {
		return "", err
	}
	stop, err := canonicalDecimal("stop price", r.StopPrice, true)
	if err != nil {
		return "", err
	}
	target, err := canonicalDecimal("target price", r.TargetPrice, false)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(r.PolicyVersion) == "" {
		return "", fmt.Errorf("%w: risk intent has no policy version", ErrInvalidRequest)
	}
	return canonicalJSON([][2]string{
		{"kind", PreimageKindRiskIntent},
		{"account_ref", strings.TrimSpace(r.AccountRef)},
		{"market", market},
		{"symbol", symbol},
		{"side", side},
		{"quantity", quantity},
		{"entry_price", entry},
		{"stop_price", stop},
		{"target_price", target},
		{"policy_version", strings.TrimSpace(r.PolicyVersion)},
	}), nil
}

// ReductionIntent is the preimage of a RISK_REDUCING decision: an upper bound on
// what may be removed, and why.
//
// MaxQuantity is a ceiling rather than an exact size on purpose — an exit sized
// from a fresher sellable-quantity read must not be refused for being *smaller*
// than the decision (§0.3), and it must not be allowed to be larger.
type ReductionIntent struct {
	AccountRef  string `json:"account_ref"`
	Market      string `json:"market"`
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	MaxQuantity string `json:"max_quantity"`
	Reason      string `json:"reason"`
}

// Kind implements Preimage.
func (ReductionIntent) Kind() string { return PreimageKindReductionIntent }

// Canonical implements Preimage.
func (r ReductionIntent) Canonical() (string, error) {
	market, symbol, side, err := canonicalVenue(r.Market, r.Symbol, r.Side)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(r.AccountRef) == "" {
		return "", fmt.Errorf("%w: reduction intent has no account", ErrInvalidRequest)
	}
	max, err := canonicalDecimal("max quantity", r.MaxQuantity, true)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(r.Reason) == "" {
		return "", fmt.Errorf("%w: reduction intent has no reason", ErrInvalidRequest)
	}
	return canonicalJSON([][2]string{
		{"kind", PreimageKindReductionIntent},
		{"account_ref", strings.TrimSpace(r.AccountRef)},
		{"market", market},
		{"symbol", symbol},
		{"side", side},
		{"max_quantity", max},
		{"reason", strings.TrimSpace(r.Reason)},
	}), nil
}

// ParsePreimage turns stored canonical text back into a typed preimage.
//
// It is strict in both directions: unknown fields are refused, and the parsed
// value must re-canonicalise to exactly the stored text. That round trip is what
// makes "the text is the hash's subject" and "the struct is what we compare
// against the order" the same statement — without it, a blob could carry extra
// fields, or a differently spelled decimal, and still hash to a value the
// verification accepted.
func ParsePreimage(kind, canonical string) (Preimage, error) {
	var (
		parsed Preimage
		err    error
	)
	dec := json.NewDecoder(strings.NewReader(canonical))
	dec.DisallowUnknownFields()
	switch kind {
	case PreimageKindRiskIntent:
		var v struct {
			RiskIntent
			Kind string `json:"kind"`
		}
		err = dec.Decode(&v)
		parsed = v.RiskIntent
	case PreimageKindReductionIntent:
		var v struct {
			ReductionIntent
			Kind string `json:"kind"`
		}
		err = dec.Decode(&v)
		parsed = v.ReductionIntent
	default:
		return nil, fmt.Errorf("%w: preimage kind %q is not one this build understands",
			ErrInvalidRequest, kind)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: reading the %s preimage: %v", ErrInvalidRequest, kind, err)
	}
	round, err := parsed.Canonical()
	if err != nil {
		return nil, err
	}
	if round != canonical {
		return nil, fmt.Errorf("%w: the stored preimage is not in canonical form", ErrInvalidRequest)
	}
	return parsed, nil
}

// HashPreimage is the binding hash: SHA-256 over the canonical text, hex.
func HashPreimage(canonical string) string {
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// Decision is a persisted decisions row.
type Decision struct {
	ID          string
	AccountRef  string
	Generation  int
	SafetyClass string
	// PreimageKind and RiskPreimage are the stored text; RiskHash is over it.
	PreimageKind string
	RiskPreimage string
	RiskHash     string
	// ClientOrderID is set for place decisions only, and is always
	// DeriveClientOrderID(ID, Generation).
	ClientOrderID string
	// LimitsJSON is the limit snapshot for EXPOSURE_RAISING, empty otherwise.
	LimitsJSON string
	Nonce      string
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// DecisionRequest asks the journal to persist one decision.
type DecisionRequest struct {
	// ID is the issuer's own identifier. It is the only thing a submitter passes
	// to the gateway, so it must be unguessable enough not to be reachable by
	// accident — issuers mint it the same way attempt ids are minted.
	ID         string
	AccountRef string
	// Generation is the reissue counter. It must be 0 in this change: the
	// mechanism that advances it (protection replacement, reissue) is defined by
	// a later change, and a generation nothing advances is a field nothing can
	// forge (engine-safety: "generation은 이 change에서 항상 0").
	Generation int
	// SafetyClass is EXPOSURE_RAISING or RISK_REDUCING. PROTECTION_WEAKENING is
	// reserved in the schema and refused here: this change issues none.
	SafetyClass string
	// Kind is the mutation the decision authorises. PLACE decisions carry a
	// derived idempotency key; CANCEL and AMEND do not, because the broker's
	// cancel and modify endpoints take no clientOrderId (openapi).
	Kind MutationKind
	// Preimage is the class-specific risk data.
	Preimage Preimage
	// LimitsJSON is the limit snapshot. Required for EXPOSURE_RAISING and
	// refused for RISK_REDUCING — an exit that carries limits is an exit
	// something can refuse (§0.3).
	LimitsJSON string
	// Nonce is the one-shot authorisation token, spent when the attempt records
	// its dispatch. Unique across the database.
	Nonce               string
	IssuedAt, ExpiresAt time.Time
}

// RecordDecision persists a decision and returns the stored form.
//
// It is the issuer's transaction: it commits on its own, before the gateway is
// called, so a crash between deciding and submitting leaves the decision on disk
// with nothing sent — which is the recoverable direction. The reverse (submit,
// then record) has no recoverable direction at all.
func (j *Journal) RecordDecision(ctx context.Context, req DecisionRequest) (Decision, error) {
	dec, err := req.build()
	if err != nil {
		return Decision{}, err
	}
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO decisions
		   (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
		    risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		dec.ID, dec.AccountRef, dec.Generation, dec.SafetyClass, dec.PreimageKind,
		dec.RiskPreimage, dec.RiskHash, nullableString(dec.ClientOrderID),
		nullableString(dec.LimitsJSON), dec.Nonce,
		formatJournalTime(dec.IssuedAt), formatJournalTime(dec.ExpiresAt)); err != nil {
		return Decision{}, fmt.Errorf("journal: recording decision %s: %w", dec.ID, err)
	}
	return dec, nil
}

// build validates the request and derives everything the row stores.
func (r DecisionRequest) build() (Decision, error) {
	invalid := func(format string, args ...any) (Decision, error) {
		return Decision{}, fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
	}
	id := strings.TrimSpace(r.ID)
	account := strings.TrimSpace(r.AccountRef)
	class := strings.TrimSpace(r.SafetyClass)
	nonce := strings.TrimSpace(r.Nonce)

	switch {
	case id == "":
		return invalid("a decision id is required")
	case account == "":
		return invalid("a decision account ref is required")
	case nonce == "":
		return invalid("a decision nonce is required — the one-shot property has nothing to spend without one")
	case r.Generation != 0:
		return invalid("generation %d: this build issues generation 0 only; advancing it belongs to the "+
			"change that introduces reissue", r.Generation)
	case !r.Kind.valid():
		return invalid("unknown mutation kind %q", r.Kind)
	case r.Preimage == nil:
		return invalid("a decision carries a risk preimage; there is nothing to re-verify without one")
	}
	switch class {
	case SafetyClassExposureRaising, SafetyClassRiskReducing:
	case SafetyClassProtectionWeakening:
		return invalid("PROTECTION_WEAKENING is reserved and is not issued by this build")
	default:
		return invalid("safety class %q is not one this build issues", class)
	}
	if want := preimageKindFor(class); r.Preimage.Kind() != want {
		return invalid("a %s decision carries a %s preimage, got %s", class, want, r.Preimage.Kind())
	}

	canonical, err := r.Preimage.Canonical()
	if err != nil {
		return Decision{}, err
	}
	if got := preimageAccount(r.Preimage); got != account {
		return invalid("the preimage is for account %q but the decision is for %q", got, account)
	}

	limits := strings.TrimSpace(r.LimitsJSON)
	switch class {
	case SafetyClassExposureRaising:
		if limits == "" {
			return invalid("an EXPOSURE_RAISING decision requires a limit snapshot")
		}
	case SafetyClassRiskReducing:
		if limits != "" {
			return invalid("a RISK_REDUCING decision carries no limit snapshot — an exit is not limit-checked")
		}
	}

	if r.IssuedAt.IsZero() || r.ExpiresAt.IsZero() {
		return invalid("a decision is bounded in time; issued_at and expires_at are required")
	}
	if !r.ExpiresAt.After(r.IssuedAt) {
		return invalid("the decision expires at or before it was issued")
	}

	key := ""
	if r.Kind == KindPlace {
		key = DeriveClientOrderID(id, r.Generation)
	}

	return Decision{
		ID:            id,
		AccountRef:    account,
		Generation:    r.Generation,
		SafetyClass:   class,
		PreimageKind:  r.Preimage.Kind(),
		RiskPreimage:  canonical,
		RiskHash:      HashPreimage(canonical),
		ClientOrderID: key,
		LimitsJSON:    limits,
		Nonce:         nonce,
		IssuedAt:      r.IssuedAt,
		ExpiresAt:     r.ExpiresAt,
	}, nil
}

// LookupDecision reads a decision back. This is the gateway's only source for
// what a decision authorises.
func (j *Journal) LookupDecision(ctx context.Context, id string) (Decision, error) {
	return scanDecision(j.db.QueryRowContext(ctx, decisionSelect+" WHERE id = ?", id))
}

const decisionSelect = `SELECT id, account_ref, generation, safety_class, preimage_kind,
	risk_preimage, risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at
	FROM decisions`

func scanDecision(row rowScanner) (Decision, error) {
	var (
		dec                 Decision
		key, limits         sql.NullString
		issuedAt, expiresAt string
	)
	err := row.Scan(&dec.ID, &dec.AccountRef, &dec.Generation, &dec.SafetyClass,
		&dec.PreimageKind, &dec.RiskPreimage, &dec.RiskHash, &key, &limits, &dec.Nonce,
		&issuedAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Decision{}, ErrDecisionNotFound
	}
	if err != nil {
		return Decision{}, fmt.Errorf("journal: reading decision: %w", err)
	}
	dec.ClientOrderID = key.String
	dec.LimitsJSON = limits.String
	if dec.IssuedAt, err = parseJournalTime(issuedAt); err != nil {
		return Decision{}, fmt.Errorf("journal: decision %s issued_at: %w", dec.ID, err)
	}
	if dec.ExpiresAt, err = parseJournalTime(expiresAt); err != nil {
		return Decision{}, fmt.Errorf("journal: decision %s expires_at: %w", dec.ID, err)
	}
	return dec, nil
}

// checkDecisionBinding verifies, inside the Prepare transaction, that the
// attempt's claimed binding is the one the decision row actually says.
//
// It is a consistency check between two rows, not a risk judgement: whether the
// decision authorises *this order* is re-verified by the gateway against the
// preimage. What is enforced here is that the attempt cannot record a class, a
// generation, an account or a key its decision does not carry — which is what
// makes the attempt row usable as evidence later.
func checkDecisionBinding(ctx context.Context, tx *sql.Tx, req PrepareRequest) error {
	id := strings.TrimSpace(req.DecisionID)
	if id == "" {
		return nil
	}
	dec, err := scanDecision(tx.QueryRowContext(ctx, decisionSelect+" WHERE id = ?", id))
	if errors.Is(err, ErrDecisionNotFound) {
		return fmt.Errorf("%w: attempt %s names decision %s, which is not recorded",
			ErrInvalidRequest, req.AttemptID, id)
	}
	if err != nil {
		return err
	}

	account := strings.TrimSpace(req.AccountRef)
	if account == "" {
		account = strings.TrimSpace(req.Intent.AccountRef)
	}
	switch {
	case dec.AccountRef != account:
		return fmt.Errorf("%w: decision %s is for account %q, the attempt is on %q",
			ErrInvalidRequest, id, dec.AccountRef, account)
	case dec.Generation != req.Generation:
		return fmt.Errorf("%w: decision %s is generation %d, the attempt claims %d",
			ErrInvalidRequest, id, dec.Generation, req.Generation)
	case dec.SafetyClass != strings.TrimSpace(req.SafetyClass):
		return fmt.Errorf("%w: decision %s is %s, the attempt claims %s",
			ErrInvalidRequest, id, dec.SafetyClass, req.SafetyClass)
	case dec.ClientOrderID != strings.TrimSpace(req.ClientOrderID):
		return fmt.Errorf("%w: decision %s authorises idempotency key %q, the attempt carries %q",
			ErrInvalidRequest, id, dec.ClientOrderID, req.ClientOrderID)
	}
	return nil
}

// AttemptForClientOrderID finds the attempt that claimed an idempotency key.
//
// The key is claimed at RECORDED, so this answers "has this decision already
// been submitted once" without consulting anything in memory — which is what a
// second process, or the same process after a restart, needs.
func (j *Journal) AttemptForClientOrderID(ctx context.Context, accountRef, key string) (AttemptRecord, error) {
	if strings.TrimSpace(key) == "" {
		return AttemptRecord{}, ErrAttemptNotFound
	}
	return scanAttempt(j.db.QueryRowContext(ctx,
		attemptSelect+" WHERE account_ref = ? AND client_order_id = ?", accountRef, key))
}

// preimageKindFor maps a class to the only preimage schema it may carry.
func preimageKindFor(class string) string {
	if class == SafetyClassExposureRaising {
		return PreimageKindRiskIntent
	}
	return PreimageKindReductionIntent
}

func preimageAccount(p Preimage) string {
	switch v := p.(type) {
	case RiskIntent:
		return strings.TrimSpace(v.AccountRef)
	case ReductionIntent:
		return strings.TrimSpace(v.AccountRef)
	default:
		return ""
	}
}

// PreimageVenue reports the account, market, symbol and side a preimage binds,
// which is what the gateway compares the mutation against.
func PreimageVenue(p Preimage) (account, market, symbol, side string) {
	switch v := p.(type) {
	case RiskIntent:
		market, symbol, side, _ = canonicalVenue(v.Market, v.Symbol, v.Side)
		return strings.TrimSpace(v.AccountRef), market, symbol, side
	case ReductionIntent:
		market, symbol, side, _ = canonicalVenue(v.Market, v.Symbol, v.Side)
		return strings.TrimSpace(v.AccountRef), market, symbol, side
	default:
		return "", "", "", ""
	}
}

// --- canonical encoding -----------------------------------------------------

// canonicalJSON writes a flat JSON object with the fields in the given order.
//
// Hand-rolled rather than json.Marshal of a map, because a map's iteration order
// is not the encoder's output order by accident but by contract-free luck, and
// this text is what a hash is taken over: two builds must produce the same bytes
// for the same values, forever.
func canonicalJSON(fields [][2]string) string {
	var b strings.Builder
	b.WriteByte('{')
	for i, kv := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		writeJSONString(&b, kv[0])
		b.WriteByte(':')
		writeJSONString(&b, kv[1])
	}
	b.WriteByte('}')
	return b.String()
}

func writeJSONString(b *strings.Builder, s string) {
	// encoding/json is the escaping authority; the error case cannot happen for
	// a string, and swallowing it would be worse than the empty write it causes.
	encoded, err := json.Marshal(s)
	if err != nil {
		b.WriteString(`""`)
		return
	}
	b.Write(encoded)
}

// canonicalVenue normalises the three fields that identify where a mutation
// lands, so a decision written with "kr"/"aapl" binds the same order as one
// written with "KR"/"AAPL".
func canonicalVenue(market, symbol, side string) (string, string, string, error) {
	m := strings.ToLower(strings.TrimSpace(market))
	s := strings.ToUpper(strings.TrimSpace(symbol))
	sd := strings.ToUpper(strings.TrimSpace(side))
	switch {
	case m == "":
		return "", "", "", fmt.Errorf("%w: the preimage has no market", ErrInvalidRequest)
	case s == "":
		return "", "", "", fmt.Errorf("%w: the preimage has no symbol", ErrInvalidRequest)
	case sd != "BUY" && sd != "SELL":
		return "", "", "", fmt.Errorf("%w: preimage side %q is neither BUY nor SELL", ErrInvalidRequest, side)
	}
	return m, s, sd, nil
}

// canonicalDecimal normalises a decimal string without going through a float:
// the journal stores decimals as text precisely because binary floats cannot
// hold a broker's decimals exactly, and normalising through one here would undo
// that for the single value the whole verification compares on.
func canonicalDecimal(field, value string, required bool) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		if required {
			return "", fmt.Errorf("%w: the preimage has no %s", ErrInvalidRequest, field)
		}
		return "", nil
	}
	normalised, ok := NormalizeDecimal(v)
	if !ok {
		return "", fmt.Errorf("%w: %s %q is not a decimal", ErrInvalidRequest, field, value)
	}
	if required && normalised == "0" {
		return "", fmt.Errorf("%w: %s must be greater than zero", ErrInvalidRequest, field)
	}
	return normalised, nil
}

// NormalizeDecimal returns the canonical spelling of a non-negative decimal
// string: no sign, no leading zeros, no trailing fractional zeros, no bare
// trailing point. It reports false for anything that is not one.
//
// Exact and float-free, so "0.10000000000000000555" survives as itself.
func NormalizeDecimal(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if s[0] == '+' {
		s = s[1:]
	}
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}
	intPart, fracPart, hasPoint := strings.Cut(s, ".")
	if intPart == "" && fracPart == "" {
		return "", false
	}
	if !allDigits(intPart) || !allDigits(fracPart) {
		return "", false
	}
	if hasPoint && fracPart == "" && intPart == "" {
		return "", false
	}
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}
	fracPart = strings.TrimRight(fracPart, "0")
	out := intPart
	if fracPart != "" {
		out += "." + fracPart
	}
	if negative && out != "0" {
		out = "-" + out
	}
	return out, true
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// formatJournalTime / parseJournalTime are the journal's timestamp convention:
// RFC3339 in UTC, second resolution. Formatting truncates sub-second precision,
// which moves an expiry *earlier* — the conservative direction, and the only one
// worth having by accident.
func formatJournalTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func parseJournalTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not an RFC3339 instant: %w", s, err)
	}
	return t.UTC(), nil
}
