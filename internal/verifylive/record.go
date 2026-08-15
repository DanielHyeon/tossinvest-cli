package verifylive

// record.go is the evidence file.
//
// It follows internal/soak/record.go exactly, and for the same reason: an
// append-per-step JSON Lines file survives the machine dying between steps, and a
// torn *final* line is what a crash during an append looks like while a broken
// line anywhere else is corruption. A verification that has to be started over
// because the laptop slept is a verification that costs another live order.
//
// # What is and is not written here
//
// Digests, not bodies. The record is read by an operator, quoted into an
// attestation and — sooner or later — pasted into a support thread or an agent
// transcript. It carries request/response digests, latencies, error codes, status
// strings and counts. It does not carry order bodies, prices paid, balances, or
// an unmasked account number.
//
// The one exception is identifiers of live objects this tool created: an
// outstanding order id has to be printable, because "there is an order on your
// account and here is its id" is the whole point of Outstanding().

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileName is the evidence record's default name inside the data directory. It
// sits next to internal/soak's record: the two are read together when the
// attestation is built.
const FileName = "capability-verify.jsonl"

// USFileName is the US market's record.
//
// A verdict is about an account *and a market*: the same StepID means "the broker
// accepts a stop order" in one market and says nothing about the other. Sharing
// one file would let a US pass settle a KR step through settled/RedoSet, which
// would record a capability nobody measured. Separate files keep every existing
// mechanism — StepCount, RedoSet, resume, the plan digest, one-run-per-process —
// working unchanged, per market.
const USFileName = "capability-verify-us.jsonl"

// RecordFileName is the evidence file for a market. An unspecified market is KR,
// which is the file every existing record already lives in.
func RecordFileName(market string) string {
	if NormalizeMarket(market) == MarketUS {
		return USFileName
	}
	return FileName
}

// RecordFormatVersion is the format this build writes and understands.
const RecordFormatVersion = 1

// KindStep marks a step record. The field exists so a later change can append a
// different kind of line without every old reader mistaking it for a step.
const KindStep = "step"

// KindApproval marks the run-wide batch approval, granted or declined.
//
// It is a line of its own rather than a field on the first step because it is not
// about a step: it records that a person was shown a list of N live requests and
// answered, which is the fact the whole approval model turns on. Readers that walk
// steps skip it on the Kind, which is what the field was put there for.
const KindApproval = "approval"

// KindCleanup marks the prologue that removes what an earlier run left resting.
//
// It is a third kind for the reason KindStep's comment anticipated: the line
// records live requests and their artifacts, so Outstanding has to see it, but it
// measures no capability, so every reader that walks the procedure has to skip it.
// isStepEntry is the single place that decides which of those two it is.
const KindCleanup = "cleanup"

// KindM0Checkpoint is an owner-only recovery line. Unlike the sanitized causal
// receipt, it may retain exact broker identifiers so status/abort can direct a
// human to the only objects this verification owns.
const KindM0Checkpoint = "m0-checkpoint"

// Verdict is a step's outcome.
type Verdict string

// The verdicts. Only VerdictPass means the property was established.
const (
	// VerdictPass means the step's claim was observed to hold.
	VerdictPass Verdict = "pass"
	// VerdictFail means the step ran and its claim did not hold, or the broker
	// refused in a way the step could not interpret.
	VerdictFail Verdict = "fail"
	// VerdictSkipped means the step was not attempted, and Reason says why. A
	// skip is never a pass: an unexercised property cannot be attested.
	VerdictSkipped Verdict = "skipped"
	// VerdictRefused means a person declined the confirmation. It is recorded
	// rather than swallowed — a refusal is a decision and belongs in the audit
	// trail — and the run continues with the steps that do not depend on it.
	VerdictRefused Verdict = "refused"
	// VerdictDeferred means the step cannot be driven by this tool at all and is
	// recorded as unverified up front.
	VerdictDeferred Verdict = "deferred"
	// VerdictAwaitingRestart is the conditional-persistence step's halt: the
	// property is "it survived this process exiting", and no single process can
	// observe that about itself.
	VerdictAwaitingRestart Verdict = "awaiting-restart"
)

// Terminal reports whether a verdict ends a step. An awaiting-restart step is
// the only one a later run picks up again.
func (v Verdict) Terminal() bool { return v != VerdictAwaitingRestart }

// Process identifies the run's operating-system process.
//
// InstanceID is what makes conditional-persist enforceable. A PID is reused and a
// start time has second granularity on some systems; a random token minted once
// per process cannot collide, so "a different process read the conditional back"
// is a fact the record can prove rather than a claim the operator makes.
type Process struct {
	// PID is the process identifier, for the audit trail.
	PID int `json:"pid"`
	// InstanceID is unique to this process. Two entries sharing it were written
	// by the same run of the binary.
	InstanceID string `json:"instance_id"`
	// StartedAt is when this process built its runner.
	StartedAt time.Time `json:"started_at"`
}

// NewProcess mints the identity for this process.
func NewProcess(now time.Time) Process {
	return Process{PID: os.Getpid(), InstanceID: newToken("proc"), StartedAt: now.UTC()}
}

// Call is one HTTP call a step made.
type Call struct {
	// Endpoint is the call as "METHOD /path", spelled the way
	// internal/soak and internal/app/engine spell their endpoint sets so the
	// three compare without a translation table.
	Endpoint string `json:"endpoint"`
	// At is when the call started.
	At time.Time `json:"at"`
	// LatencyMS is the round trip. The idempotency step's safety margin is
	// derived from these, so they are recorded for every call and not only the
	// slow ones.
	LatencyMS int64 `json:"latency_ms"`
	// RequestDigest and ResponseDigest identify the bodies without carrying them.
	RequestDigest  string `json:"request_digest,omitempty"`
	ResponseDigest string `json:"response_digest,omitempty"`
	// Error is the failure, truncated. The official client's errors are status
	// classifications and the broker's own error envelope; neither carries a
	// credential.
	Error string `json:"error,omitempty"`
	// ErrorCode is the broker's error code when the body carried one — the
	// `idempotency-key-conflict` this whole exercise is looking for.
	ErrorCode string `json:"error_code,omitempty"`
}

// Observation is one measured property.
//
// Key is a dotted name that `verify report` groups on and that task 1.4 copies
// into the attestation's attribute set, so the names are part of the contract
// between this tool and the attestation: see report.go.
type Observation struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Detail string `json:"detail,omitempty"`
}

// Artifact is a live object this tool created.
//
// Every one of them must end the run with Cancelled true, with the single
// deliberate exception of the conditional order that has to outlive the process
// for conditional-persist to mean anything. Runner.Run refuses to report a clean
// finish while any other artifact is outstanding.
type Artifact struct {
	// Kind is "order" or "conditional-order".
	Kind string `json:"kind"`
	// ID is the broker's identifier. Printed in full on purpose: an operator who
	// has to go and cancel something by hand needs it.
	ID string `json:"id"`
	// Symbol is what it is against.
	Symbol string `json:"symbol"`
	// CreatedAt and CancelledAt bound its life.
	CreatedAt   time.Time `json:"created_at"`
	CancelledAt time.Time `json:"cancelled_at,omitempty"`
	// Cancelled reports that this tool confirmed the object is gone.
	Cancelled bool `json:"cancelled"`
	// Filled and FilledAt are the other way an object stops existing.
	//
	// Until 2026-07-30 there was only one, and "not cancelled" and "live" were the
	// same sentence. A conditional order that fires produces a child order that is
	// *meant* to fill, and a filled order is not a cancelled one — so without this
	// the record insists a successful measurement's child is live forever: filling
	// the exposure cap, printed as a leftover, and offered for cancellation by the
	// next run's prologue.
	//
	// It is a second field rather than a reuse of Cancelled because this file is
	// the document every entry in measurements.md is derived from. Writing a fill
	// as a cancellation would behave correctly in all three of those places and
	// make the trigger measurement's own conclusion read "we cancelled it".
	//
	// FilledAt is this tool's observation time, not the broker's. The broker
	// supplies no fill time at all — lastExecutedAt is null on every completed
	// order in both markets, and a US order's orderedAt carries only a date
	// (measurements.md M44, M45) — so the polling interval that produced it is the
	// error bound and is recorded alongside it as an observation.
	//
	// Absent is what every line written before 2026-07-30 carries, and such a line
	// keeps exactly the judgement it has today.
	Filled   bool      `json:"filled,omitempty"`
	FilledAt time.Time `json:"filled_at,omitempty"`
	// Deliberate marks the conditional order that is meant to survive the
	// process exit. It is still outstanding; it is just not a mistake.
	Deliberate bool `json:"deliberate,omitempty"`
	// HeldUntil names the step whose terminal verdict releases this object.
	//
	// Until that step decides — and decides *after* this line was written — the
	// cleanup prologue must not offer to cancel it. It exists because a
	// measurement that spans runs has to be able to say what it is waiting on:
	// the conditional a persistence step reads from a later process, and the
	// child order a trigger measurement has to watch fill. Without it, the only
	// object the prologue protected was a conditional order, and only ever
	// against one step's verdict (see holdGate).
	//
	// Empty is the record every line written before 2026-07-30 carries, and it
	// keeps exactly the judgement it has today.
	HeldUntil StepID `json:"held_until,omitempty"`
	// ChainID groups the objects one measurement created.
	//
	// A conditional modify issues a new identifier and invalidates the old one
	// immediately (measurements.md M19, M40 — both markets), so "these two are
	// one protection" is not derivable from the broker afterwards. Before this
	// field the link lived only in the prose of Note.
	ChainID string `json:"chain_id,omitempty"`
	// Note explains an artifact whose state needs one.
	Note string `json:"note,omitempty"`
}

// Entry is one line of the record: one step, one verdict.
type Entry struct {
	FormatVersion int    `json:"format_version"`
	Kind          string `json:"kind"`

	// RunID groups the steps of one invocation.
	RunID string `json:"run_id"`
	// Process is who wrote the line.
	Process Process `json:"process"`

	// StepID and Title identify the step; Tasks are the measurement tasks it
	// feeds, so a reader can go from a line back to tasks.md.
	StepID StepID   `json:"step_id"`
	Title  string   `json:"title"`
	Tasks  []string `json:"tasks,omitempty"`

	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`

	// AccountRef is masked. The unmasked number never enters this file.
	AccountRef string `json:"account_ref"`

	// Mutating repeats the catalogue's flag on the line itself, so a reader
	// grepping the record for "what did this touch" does not need the catalogue.
	Mutating bool `json:"mutating"`

	Verdict Verdict `json:"verdict"`
	Reason  string  `json:"reason,omitempty"`

	Calls        []Call        `json:"calls,omitempty"`
	Observations []Observation `json:"observations,omitempty"`
	Artifacts    []Artifact    `json:"artifacts,omitempty"`
	M0Checkpoint *M0Checkpoint `json:"m0_checkpoint,omitempty"`
}

// M0Checkpoint records pending/create/child recovery identity in the existing
// owner-only verify record. It is not a trading-journal record.
type M0Checkpoint struct {
	Kind                string `json:"kind"`
	ClientOrderID       string `json:"client_order_id,omitempty"`
	ParentConditionalID string `json:"parent_conditional_id,omitempty"`
	ChildOrderID        string `json:"child_order_id,omitempty"`
	Symbol              string `json:"symbol"`
	Market              string `json:"market"`
	Type                string `json:"type"`
	Quantity            string `json:"quantity"`
	Side                string `json:"side"`
	OrderType           string `json:"order_type"`
	ConditionType       string `json:"condition_type"`
	Trigger             string `json:"trigger"`
	ExpireDate          string `json:"expire_date"`
}

// Recorder appends entries to the record.
//
// Every append is flushed and synced before it returns, exactly as the soak's is:
// a step whose evidence was still in a page cache when the machine lost power is
// a live order nobody has a record of.
type Recorder struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// OpenRecorder opens (or creates) the record for appending, owner-only.
func OpenRecorder(path string) (*Recorder, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("verifylive: creating %s: %w", dir, err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("verifylive: opening the record %s: %w", path, err)
	}
	return &Recorder{path: path, f: f}, nil
}

// Path returns the record's location.
func (r *Recorder) Path() string { return r.path }

// Append writes one entry as a single line and syncs it.
func (r *Recorder) Append(e Entry) error {
	if e.FormatVersion == 0 {
		e.FormatVersion = RecordFormatVersion
	}
	if e.Kind == "" {
		e.Kind = KindStep
	}
	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("verifylive: encoding a step entry: %w", err)
	}
	line = append(line, '\n')

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f == nil {
		return errors.New("verifylive: the record is closed")
	}
	if _, err := r.f.Write(line); err != nil {
		return fmt.Errorf("verifylive: appending to %s: %w", r.path, err)
	}
	if err := r.f.Sync(); err != nil {
		return fmt.Errorf("verifylive: syncing %s: %w", r.path, err)
	}
	return nil
}

// Close releases the record. It is safe to call more than once.
func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f == nil {
		return nil
	}
	err := r.f.Close()
	r.f = nil
	return err
}

var errUnknownFormat = errors.New("the record was written by a newer TossOS")

// LoadEntries reads the record.
//
// A missing file is an empty result and no error: "the verification has not been
// started" is the state every machine is in before it is, and every surface asks
// about it.
func LoadEntries(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("verifylive: reading the record %s: %w", path, err)
	}
	defer f.Close()

	var (
		entries []Entry
		lineNo  int
	)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	// pending holds the previously scanned line so a parse failure on the *last*
	// line can be told apart from one in the middle: only the last can be torn.
	var pending []byte
	var pendingLine int

	flush := func() error {
		if pending == nil {
			return nil
		}
		e, err := decodeEntry(pending)
		if err != nil {
			return fmt.Errorf("verifylive: %s line %d: %w", path, pendingLine, err)
		}
		entries = append(entries, e)
		pending = nil
		return nil
	}

	for scanner.Scan() {
		lineNo++
		raw := append([]byte(nil), scanner.Bytes()...)
		if len(trimSpace(raw)) == 0 {
			continue
		}
		if err := flush(); err != nil {
			return nil, err
		}
		pending, pendingLine = raw, lineNo
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("verifylive: reading %s: %w", path, err)
	}

	if pending != nil {
		e, err := decodeEntry(pending)
		if err != nil {
			if errors.Is(err, errUnknownFormat) {
				return nil, fmt.Errorf("verifylive: %s line %d: %w", path, pendingLine, err)
			}
			// A torn final line is what a crash during an append leaves behind.
			return entries, nil
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func decodeEntry(line []byte) (Entry, error) {
	var e Entry
	if err := json.Unmarshal(line, &e); err != nil {
		return Entry{}, fmt.Errorf("the line is not a step record: %w", err)
	}
	if e.FormatVersion > RecordFormatVersion {
		return Entry{}, fmt.Errorf("%w: format %d, this build understands %d",
			errUnknownFormat, e.FormatVersion, RecordFormatVersion)
	}
	return e, nil
}

// --- derived state ------------------------------------------------------------

// isStepEntry reports a line that describes a step's outcome.
//
// An empty Kind counts: the field was added with the format and every line written
// before it existed was a step. Readers that walk steps use this rather than
// testing the field directly, so a third kind added later is skipped by all of them
// at once.
func isStepEntry(e Entry) bool { return e.Kind == "" || e.Kind == KindStep }

// StepCount reports how much of the procedure a record actually holds.
//
// It is not len(entries): a run whose batch approval was declined writes one line
// and touches nothing, and "the record already holds a verification" has to stay
// false afterwards — otherwise declining once would lock the operator out of
// starting, and the next run would introduce itself as a resumption of nothing.
func StepCount(entries []Entry) int {
	n := 0
	for _, e := range entries {
		if isStepEntry(e) {
			n++
		}
	}
	return n
}

// LastEntry returns the newest entry for a step.
func LastEntry(entries []Entry, id StepID) (Entry, bool) {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].StepID == id {
			return entries[i], true
		}
	}
	return Entry{}, false
}

// Settled reports whether a step has a terminal verdict on the record, so a
// resumed run does not place a second live order for work already done.
func Settled(entries []Entry, id StepID) bool {
	e, ok := LastEntry(entries, id)
	return ok && e.Verdict.Terminal()
}

// Passed reports whether a step's newest verdict is a pass. Dependency checks use
// it: a step whose prerequisite was skipped or refused must be skipped, not
// attempted against an object that was never created.
func Passed(entries []Entry, id StepID) bool {
	e, ok := LastEntry(entries, id)
	return ok && e.Verdict == VerdictPass
}

// Outstanding returns the live objects the record says this tool created and
// never cancelled.
//
// The whole record is walked and the *last* mention of each identifier wins, so
// an order created in one step and cancelled in another nets out. This is the
// function that decides whether the tool can report a clean finish.
func Outstanding(entries []Entry) []Artifact {
	var out []Artifact
	for _, l := range outstandingLines(entries) {
		out = append(out, l.Artifact)
	}
	return out
}

// ChainOf returns the chain identifier the record carries for an object, taking
// the newest line that names one.
//
// Empty means the record does not group this object with anything, which is what
// every line written before 2026-07-30 says.
func ChainOf(entries []Entry, kind, id string) string {
	chain := ""
	for _, e := range entries {
		for _, a := range e.Artifacts {
			if a.Kind == kind && a.ID == id && a.ChainID != "" {
				chain = a.ChainID
			}
		}
	}
	return chain
}

// outstandingLine is an outstanding artifact together with the index of the line
// Outstanding chose for it.
//
// The index is not decoration. The cleanup rule reads HeldUntil off this
// artifact, so the verdict that releases it has to be measured against the line
// that carried that field and no other (cleanup.go heldAfter).
type outstandingLine struct {
	Artifact
	at int
}

func outstandingLines(entries []Entry) []outstandingLine {
	order := []string{}
	latest := map[string]outstandingLine{}
	for i, e := range entries {
		for _, a := range e.Artifacts {
			key := a.Kind + "\x00" + a.ID
			if _, seen := latest[key]; !seen {
				order = append(order, key)
			}
			prev, seen := latest[key]
			if seen && prev.terminal() && !a.terminal() {
				// A later line that does not know how this ended must not
				// resurrect it. Both endings are monotone.
				continue
			}
			latest[key] = outstandingLine{Artifact: a, at: i}
		}
	}
	var out []outstandingLine
	for _, key := range order {
		if l := latest[key]; !l.terminal() {
			out = append(out, l)
		}
	}
	return out
}

// terminal reports that this line says the object has stopped existing.
//
// Two ways, and the difference between them is recorded rather than flattened:
// this tool cancelled it, or it filled. Every consumer of Outstanding asks the
// same question through this one predicate, so a third ending added later is
// honoured by all of them at once.
func (a Artifact) terminal() bool { return a.Cancelled || a.Filled }

// --- digests ------------------------------------------------------------------

// Digest is the record's fingerprint for a request or response body.
//
// Truncated to 16 bytes: it exists so two lines can be compared and so a claim
// can be tied to the bytes that supported it, not so anybody can reverse it.
func Digest(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "sha256:unencodable"
	}
	return DigestBytes(b)
}

// DigestBytes fingerprints raw bytes.
func DigestBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:16])
}

// newToken returns a short unambiguous random token with a prefix.
func newToken(prefix string) string {
	var buf [10]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("verifylive: crypto/rand is unavailable: " + err.Error())
	}
	return prefix + "-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf[:])
}

func trimSpace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && isSpaceByte(b[start]) {
		start++
	}
	for end > start && isSpaceByte(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
