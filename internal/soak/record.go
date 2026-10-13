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

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

// FileName is the record's default name inside the data directory.
const FileName = "capability-soak.jsonl"

// RecordFormatVersion is the record format this build writes and understands.
// a131: 2 = 계좌를 봉인해 쓰는 형식(redact.SealedFormatVersion 과 같아야 함 — 시험이 고정). 형식 1 까지만 아는 옛 읽기는
// 새 줄을 「newer format」으로 거절함(평문 옛 줄 + 봉인 새 줄 기록으로 옛 판본이 판정하는 길을 닫음).
const RecordFormatVersion = 2

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

	// OrderPages is how many pages the CLOSED and OPEN walks read between them.
	OrderPages int `json:"order_pages"`
	// OrderIDs is how many order identifiers those walks returned between them.
	//
	// The two groups overlap — the spec puts PARTIAL_FILLED in both — so an order
	// the broker reports in each of them is counted once per walk rather than
	// silently folded into one. OrdersInBothStatuses is how many those are, and
	// OrderIDs minus it is the number of distinct orders seen.
	OrderIDs int `json:"order_ids"`
	// OrdersInBothStatuses is how many identifiers came back from both the CLOSED
	// and the OPEN walk. It is an observation, not a fault.
	OrdersInBothStatuses int `json:"orders_in_both_statuses,omitempty"`
	// DuplicateOrderIDs is how many identifiers appeared more than once inside a
	// single walk — a list that cannot be paged through exactly once.
	DuplicateOrderIDs int `json:"duplicate_order_ids,omitempty"`
	// BlankOrderIDs is how many entries the broker returned with no identifier on
	// them. Nothing can reconcile against an order it cannot name.
	BlankOrderIDs int `json:"blank_order_ids,omitempty"`
	// CursorLoop reports a cursor that pointed back at a page already read.
	CursorLoop bool `json:"cursor_loop,omitempty"`
	// TruncatedPagination reports hasNext with no cursor to follow.
	TruncatedPagination bool `json:"truncated_pagination,omitempty"`
	// PageLimitReached reports that a walk stopped at its own page cap. It is
	// noted rather than failed: the pagination contract still held over every page
	// that was read, and what is beyond the cap is unread, not wrong.
	PageLimitReached bool `json:"page_limit_reached,omitempty"`

	// OpenOrders is how many orders the broker reported as working.
	OpenOrders int `json:"open_orders"`

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
	// AccountRefSealed 는 a131 부터 기록 줄에 실리는 봉인된 계좌임(redact.SealRecordLine). 쓰기 때는 비어 있고 줄 변환이
	// 채움; 읽을 때 LoadCycles 가 열어 AccountRef 를 채움 — 결속 판정은 종전과 같은 값을 봄.
	AccountRefSealed string `json:"account_ref_sealed,omitempty"`

	// Credential, Endpoints and Completeness are the three measurements.
	Credential   Credential       `json:"credential"`
	Endpoints    []EndpointResult `json:"endpoints"`
	Completeness Completeness     `json:"completeness"`

	// Binary is the executable the process that wrote this cycle was running.
	//
	// It carries no measurement. It is here so a reader — the operator console's
	// dashboard, in practice — can tell whether the survey appending to this file
	// is the build that is installed, without a process table and without a pid
	// that means nothing an hour later. It is additive and optional: every line
	// written before this field existed parses exactly as it did.
	Binary binstamp.Stamp `json:"binary,omitzero"`
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
	// key 는 OpenRecorder 가 브로커 호출 전에 준비한 봉인 키임(a131) — Append 는 이 메모리 키만 씀.
	key *redact.SealKey
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
	// a131: 봉인 키 준비는 여기서 — 측정·실주문보다 먼저 실패해야 함(Append 에서 나면 실주문 뒤의 기록 실패가 됨).
	key, err := redact.PrepareSealKey(path)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("soak: preparing the record's seal key: %w", err)
	}
	return &Recorder{path: path, f: f, key: key}, nil
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
	// a131: 계좌는 봉인해 씀(같은 규칙 — redact.SealRecordLine, 키는 OpenRecorder 가 준비).
	line, err = redact.SealRecordLine(line, r.key)
	if err != nil {
		return fmt.Errorf("soak: sealing the record line: %w", err)
	}
	line = append(line, '\n')

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f == nil {
		return errors.New("soak: the record is closed")
	}
	// a131: 정리 명령이 경로를 원자 교체했으면 다시 엶 — 열린 fd 는 사라진 inode 를 가리켜 그 뒤 줄이 소리 없이 사라짐.
	if err := reopenIfReplaced(r); err != nil {
		return err
	}
	if _, err := r.f.Write(line); err != nil {
		return fmt.Errorf("soak: appending to %s: %w", r.path, err)
	}
	if err := r.f.Sync(); err != nil {
		return fmt.Errorf("soak: syncing %s: %w", r.path, err)
	}
	return nil
}

// reopenIfReplaced 는 기록 경로의 파일이 열린 fd 와 다른 inode 면(정리 명령의 원자 교체) 경로를 다시 엶. 호출자가 r.mu 를 쥠.
func reopenIfReplaced(r *Recorder) error {
	open, err := r.f.Stat()
	if err != nil {
		return fmt.Errorf("soak: stat the open record %s: %w", r.path, err)
	}
	now, err := os.Stat(r.path)
	if err == nil && os.SameFile(open, now) {
		return nil
	}
	f, oerr := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if oerr != nil {
		return fmt.Errorf("soak: reopening the replaced record %s: %w", r.path, oerr)
	}
	_ = r.f.Close()
	r.f = f
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
	// a131: 봉인된 계좌는 이 기록 옆의 키로 엶(키는 한 번만 읽음, 만들지 않음).
	opener := redact.NewOpener(filepath.Dir(path))
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
		if c, err = openCycleAccount(c, opener); err != nil {
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
		// a131: 열 수 없는 봉인은 찢긴 줄이 아님 — 관용 없이 오류.
		if c, err = openCycleAccount(c, opener); err != nil {
			return nil, fmt.Errorf("soak: %s line %d: %w", path, pendingLine, err)
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
