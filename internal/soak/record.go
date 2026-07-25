package soak

// record.go is the durable half.
//
// The soak's whole value is that it ran for days, so the record has to survive
// everything that happens over days: restarts, reboots, a laptop lid closing
// mid-append. JSON Lines is the format for exactly that reason — an append is one
// write, a torn write damages only the line it was writing, and every line before
// it is still evidence.
//
// The rules that follow from that are in LoadCycles: a torn *final* line is
// skipped (that is what a crash during an append looks like), a broken line
// anywhere else is an error (that is what corruption looks like, and silently
// dropping a day would shorten a streak or hide a failure).

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileName is the record's default name inside the data directory.
const FileName = "capability-soak.jsonl"

// RecordFormatVersion is the record format this build writes and understands.
const RecordFormatVersion = 1

// KindCycle marks a survey record. The field exists so a later change can append
// a different kind of line to the same file without every old reader mistaking
// it for a cycle.
const KindCycle = "cycle"

// EndpointResult is one endpoint's outcome in one cycle.
type EndpointResult struct {
	// Endpoint is the call, as "METHOD /path".
	Endpoint string `json:"endpoint"`
	// OK reports that the call succeeded.
	OK bool `json:"ok"`
	// Skipped reports that the call was not attempted, and SkipReason says why.
	// A skip is never a success: an endpoint that was never exercised cannot be
	// attested.
	Skipped bool `json:"skipped,omitempty"`
	// SkipReason is the human-readable reason for a skip.
	SkipReason string `json:"skip_reason,omitempty"`
	// Class is the outcome's classification.
	Class Class `json:"class,omitempty"`
	// LatencyMS is how long the call took, in milliseconds.
	LatencyMS int64 `json:"latency_ms"`
	// Requests is how many HTTP calls the probe made — more than one for the
	// paginated order walk.
	Requests int `json:"requests"`
	// Error is the failure message. It carries no credentials: the official
	// client's errors are status classifications and transport messages.
	Error string `json:"error,omitempty"`
}

// Credential is what one cycle observed about the unattended credential.
type Credential struct {
	// OK reports that the authenticated read succeeded.
	OK bool `json:"ok"`
	// Class classifies a failure; ClassAuth is the one that breaks the streak.
	Class Class `json:"class,omitempty"`
	// Observed reports that the token expiry could be read at all.
	Observed bool `json:"observed,omitempty"`
	// TokenExpiresAt is the cached access token's expiry. The token itself is
	// never recorded.
	TokenExpiresAt time.Time `json:"token_expires_at,omitempty"`
	// Refreshed reports that the expiry moved forward since the previous cycle —
	// the direct evidence that the credential renewed itself unattended.
	Refreshed bool `json:"refreshed,omitempty"`
}

// Completeness is what one cycle observed about the reads being whole.
type Completeness struct {
	// Evaluated reports that the check could be made at all. A cycle whose order
	// walk failed leaves this false, and Evaluate ignores it rather than counting
	// a network failure twice.
	Evaluated bool `json:"evaluated"`
	// OK reports that every completeness check passed.
	OK bool `json:"ok"`

	// OrderPages is how many pages the full order walk read.
	OrderPages int `json:"order_pages"`
	// OrderIDs is how many orders the full walk returned.
	OrderIDs int `json:"order_ids"`
	// DuplicateOrderIDs is how many identifiers appeared more than once.
	DuplicateOrderIDs int `json:"duplicate_order_ids,omitempty"`
	// CursorLoop reports a cursor that pointed back at a page already read.
	CursorLoop bool `json:"cursor_loop,omitempty"`
	// TruncatedPagination reports hasNext with no cursor to follow.
	TruncatedPagination bool `json:"truncated_pagination,omitempty"`
	// PageLimitReached reports that the walk stopped at its own page cap. It
	// suppresses the open-order coverage check rather than failing it.
	PageLimitReached bool `json:"page_limit_reached,omitempty"`

	// OpenOrders is how many orders the broker reported as working.
	OpenOrders int `json:"open_orders"`
	// OpenOrdersMissing is how many of those the full list did not contain — the
	// failure that would leave an engine blind to its own exposure.
	OpenOrdersMissing int `json:"open_orders_missing,omitempty"`

	// Positions is how many holdings the account reported.
	Positions int `json:"positions"`
	// QuotesRequested and QuotesReturned compare what was asked for with what
	// came back.
	QuotesRequested int `json:"quotes_requested"`
	QuotesReturned  int `json:"quotes_returned"`

	// Detail names the specific problems, for an operator reading the status.
	Detail string `json:"detail,omitempty"`
}

// Cycle is one line of the record.
type Cycle struct {
	// FormatVersion is the record format.
	FormatVersion int `json:"format_version"`
	// Kind is KindCycle.
	Kind string `json:"kind"`

	// StartedAt and FinishedAt bound the cycle.
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`

	// AccountRef is the account the reads resolved to.
	AccountRef string `json:"account_ref"`

	// Credential, Endpoints and Completeness are the three measurements.
	Credential   Credential       `json:"credential"`
	Endpoints    []EndpointResult `json:"endpoints"`
	Completeness Completeness     `json:"completeness"`
}

// Recorder appends cycles to the record.
//
// Every append is flushed and synced before it returns. A soak whose last day
// was still in a page cache when the machine lost power is a soak that has to
// start again.
type Recorder struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// OpenRecorder opens (or creates) the record for appending, owner-only.
func OpenRecorder(path string) (*Recorder, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("soak: creating %s: %w", dir, err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("soak: opening the record %s: %w", path, err)
	}
	return &Recorder{path: path, f: f}, nil
}

// Path returns the record's location.
func (r *Recorder) Path() string { return r.path }

// Append writes one cycle as a single line and syncs it.
func (r *Recorder) Append(c Cycle) error {
	if c.FormatVersion == 0 {
		c.FormatVersion = RecordFormatVersion
	}
	if c.Kind == "" {
		c.Kind = KindCycle
	}
	line, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("soak: encoding a cycle: %w", err)
	}
	line = append(line, '\n')

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f == nil {
		return errors.New("soak: the record is closed")
	}
	if _, err := r.f.Write(line); err != nil {
		return fmt.Errorf("soak: appending to %s: %w", r.path, err)
	}
	if err := r.f.Sync(); err != nil {
		return fmt.Errorf("soak: syncing %s: %w", r.path, err)
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

// LoadCycles reads the record.
//
// A missing file is an empty result and no error: "the soak has not been started"
// is the state every machine is in before it is, and every surface asks about it.
func LoadCycles(path string) ([]Cycle, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("soak: reading the record %s: %w", path, err)
	}
	defer f.Close()

	var (
		cycles []Cycle
		lineNo int
	)
	scanner := bufio.NewScanner(f)
	// Order pages make a cycle line longer than bufio's default 64KiB ceiling on
	// a busy account.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	// pending holds the previously scanned line so a parse failure on the *last*
	// line can be told apart from one in the middle: only the last line can be a
	// torn append.
	var pending []byte
	var pendingLine int

	flush := func() error {
		if pending == nil {
			return nil
		}
		c, err := decodeCycle(pending)
		if err != nil {
			return fmt.Errorf("soak: %s line %d: %w", path, pendingLine, err)
		}
		cycles = append(cycles, c)
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
		return nil, fmt.Errorf("soak: reading %s: %w", path, err)
	}

	if pending != nil {
		c, err := decodeCycle(pending)
		if err != nil {
			if errors.Is(err, errUnknownFormat) {
				return nil, fmt.Errorf("soak: %s line %d: %w", path, pendingLine, err)
			}
			// A torn final line is what a crash during an append leaves behind.
			// Everything before it is intact and is still evidence.
			return cycles, nil
		}
		cycles = append(cycles, c)
	}
	return cycles, nil
}

var errUnknownFormat = errors.New("the record was written by a newer TossOS")

func decodeCycle(line []byte) (Cycle, error) {
	var c Cycle
	if err := json.Unmarshal(line, &c); err != nil {
		return Cycle{}, fmt.Errorf("the line is not a cycle record: %w", err)
	}
	if c.FormatVersion > RecordFormatVersion {
		return Cycle{}, fmt.Errorf("%w: format %d, this build understands %d",
			errUnknownFormat, c.FormatVersion, RecordFormatVersion)
	}
	return c, nil
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
