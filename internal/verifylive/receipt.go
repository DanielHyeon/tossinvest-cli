package verifylive

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// CausalReceipt is the M0-only, append-only evidence stream. It is deliberately
// separate from the operator record: the record may retain exact cleanup IDs,
// while this file is safe to share because it contains only tags and digests.
type CausalReceipt struct {
	mu          sync.Mutex
	f           *os.File
	dir         *os.File
	runID       string
	ready       bool
	activeLease *CausalReceiptLease
	terminalErr error
	seq         uint64
	start       time.Time
	sync        func(*os.File) error
}

// receiptHooks is deliberately private. It gives package tests deterministic
// crash points without exporting a receipt-authority or filesystem bypass.
type receiptHooks struct {
	random       func([]byte) (int, error)
	syncFile     func(*os.File) error
	syncDir      func(*os.File) error
	afterDirOpen func(string) error
	afterCreate  func(string) error
}

// ErrCausalReceiptLeased tells a caller that a whole M0 run is still using the
// receipt. Closing it would turn a completed-looking run into a partial one.
var ErrCausalReceiptLeased = errors.New("verifylive: causal receipt is leased by an active run")

// ErrCausalReceiptPoisoned is permanent for one receipt. Once a write or fsync
// fails, no later event may make the sequence look continuous again.
var ErrCausalReceiptPoisoned = errors.New("verifylive: causal receipt persistence failed permanently")

var errCausalReceiptNotReady = errors.New("verifylive: causal receipt is not ready")

// CausalReceiptLease keeps one ready receipt open for a complete Runner.Run.
// It carries no file access of its own; all writes still go through CausalReceipt.
type CausalReceiptLease struct {
	receipt *CausalReceipt
	runID   string
	once    sync.Once
}

// M0ReceiptStamp proves an event was durably appended. All values are relative
// to this receipt's run-local monotonic anchor; wall time is not causal.
type M0ReceiptStamp struct {
	RunID              string
	Sequence           uint64
	FsyncDoneElapsedNS int64
}

type receiptHeader struct {
	Version int    `json:"version"`
	RunID   string `json:"run_id"`
	Kind    string `json:"kind"`
}

// OpenCausalReceipt creates one fresh M0 receipt beneath an already secured
// owner-only directory. It never reuses a prior sequence or follows a symlink.
func OpenCausalReceipt(dir string) (*CausalReceipt, error) {
	return openCausalReceipt(dir, receiptHooks{})
}

// RunID returns the header-bound immutable receipt identity.
func (r *CausalReceipt) RunID() string {
	return r.runIDValue()
}

// AcquireRunLease reserves this receipt until Release. Runner must acquire it
// immediately before its M0 run and defer Release, so Close cannot race the
// causal chain between the header and the final child-fill receipt.
func (r *CausalReceipt) AcquireRunLease() (*CausalReceiptLease, error) {
	if r == nil {
		return nil, fmt.Errorf("verifylive: causal receipt is not ready")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.terminalErr != nil {
		return nil, r.terminalErr
	}
	if !r.ready || r.f == nil || r.dir == nil {
		return nil, errCausalReceiptNotReady
	}
	if r.activeLease != nil {
		return nil, ErrCausalReceiptLeased
	}
	lease := &CausalReceiptLease{receipt: r, runID: r.runID}
	r.activeLease = lease
	return lease, nil
}

// Release ends a whole-run lease. It is safe to defer and idempotent.
func (l *CausalReceiptLease) Release() {
	if l == nil || l.receipt == nil {
		return
	}
	l.once.Do(func() {
		l.receipt.mu.Lock()
		if l.receipt.activeLease == l {
			l.receipt.activeLease = nil
		}
		l.receipt.mu.Unlock()
	})
}

// RunID returns the immutable identity captured when the lease was acquired.
func (l *CausalReceiptLease) RunID() string {
	if l == nil {
		return ""
	}
	return l.runID
}

// usable is package-private so the Runner can reject a closed receipt without
// reading lifecycle fields directly.
func (r *CausalReceipt) usable() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.terminalErr == nil && r.ready && r.f != nil && r.dir != nil
}

// runIDValue is the single mutex-protected accessor for the immutable header
// identity. The value remains available after Close for diagnostics and leases,
// but cannot be changed after construction.
func (r *CausalReceipt) runIDValue() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runID
}

func (h receiptHooks) withDefaults() receiptHooks {
	if h.random == nil {
		h.random = rand.Read
	}
	if h.syncFile == nil {
		h.syncFile = func(file *os.File) error { return file.Sync() }
	}
	if h.syncDir == nil {
		h.syncDir = func(file *os.File) error { return file.Sync() }
	}
	return h
}

func newCausalReceipt(f, dir *os.File, runID string, syncFile func(*os.File) error) *CausalReceipt {
	return &CausalReceipt{f: f, dir: dir, runID: runID, start: time.Now(), sync: syncFile}
}

// receiptFileFromFD keeps the durable-file capability in this file. Platform
// code supplies only a verified descriptor, never a pathname-owned file.
func receiptFileFromFD(fd uintptr, name string) *os.File { return os.NewFile(fd, name) }

func currentReceiptUID() uint32 { return uint32(os.Getuid()) }

// finishOpen makes the receipt usable only after the header reaches both the
// leaf file and the same opened parent directory.
func (r *CausalReceipt) finishOpen(syncDir func(*os.File) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.appendLocked(receiptHeader{Version: 1, RunID: r.runID, Kind: "a100-m0-causal-receipt"}); err != nil {
		return err
	}
	if err := syncDir(r.dir); err != nil {
		return r.poisonLocked(fmt.Errorf("verifylive: syncing causal receipt directory: %w", err))
	}
	r.ready = true
	return nil
}

func (r *CausalReceipt) appendLocked(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return r.poisonLocked(fmt.Errorf("verifylive: encoding causal receipt: %w", err))
	}
	b = append(b, '\n')
	if _, err := r.f.Write(b); err != nil {
		return r.poisonLocked(fmt.Errorf("verifylive: writing causal receipt: %w", err))
	}
	syncFile := r.sync
	if syncFile == nil {
		syncFile = func(file *os.File) error { return file.Sync() }
	}
	if err := syncFile(r.f); err != nil {
		return r.poisonLocked(fmt.Errorf("verifylive: syncing causal receipt: %w", err))
	}
	return nil
}

func (r *CausalReceipt) poisonLocked(cause error) error {
	if r.terminalErr == nil {
		r.terminalErr = fmt.Errorf("%w: %v", ErrCausalReceiptPoisoned, cause)
	}
	r.ready = false
	if closeErr := r.closeResourcesLocked(); closeErr != nil {
		r.terminalErr = errors.Join(r.terminalErr, closeErr)
	}
	return r.terminalErr
}

// Close releases the receipt. The header and each append are individually fsynced.
func (r *CausalReceipt) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeLease != nil {
		return ErrCausalReceiptLeased
	}
	return r.closeResourcesLocked()
}

func (r *CausalReceipt) closeResourcesLocked() error {
	var err error
	if r.f != nil {
		err = r.f.Close()
		r.f = nil
	}
	if r.dir != nil {
		if closeErr := r.dir.Close(); err == nil {
			err = closeErr
		}
		r.dir = nil
	}
	r.ready = false
	return err
}

// M0ReceiptEvent is deliberately ID-free. Sequence and ElapsedNS are a single
// process-local monotonic ordering authority; wall/server time is not written.
type M0ReceiptEvent struct {
	Version   int    `json:"version"`
	Sequence  uint64 `json:"sequence"`
	Kind      string `json:"kind"`
	ElapsedNS int64  `json:"elapsed_ns"`
	Detail    string `json:"detail,omitempty"`
}

// RecordAttempt writes an ID-free digest at the transport body-read boundary.
// The lease receiver binds the write to the one active whole-run authority.
func (l *CausalReceiptLease) RecordAttempt(phase string, trace official.AttemptTrace) (M0ReceiptStamp, error) {
	if l == nil || l.receipt == nil {
		return M0ReceiptStamp{}, errCausalReceiptNotReady
	}
	r := l.receipt
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.validateLeaseLocked(l); err != nil {
		return M0ReceiptStamp{}, err
	}
	r.seq++
	payload := struct {
		M0ReceiptEvent
		StatusCode            int    `json:"status_code"`
		NoResponse            bool   `json:"no_response,omitempty"`
		ErrorClass            string `json:"error_class,omitempty"`
		DigestV1              string `json:"sha256_raw_bytes_v1,omitempty"`
		RequestStartElapsedNS int64  `json:"request_start_elapsed_ns"`
		BodyReadElapsedNS     int64  `json:"body_read_elapsed_ns"`
	}{M0ReceiptEvent: M0ReceiptEvent{Version: 1, Sequence: r.seq, Kind: "http-attempt", ElapsedNS: trace.BodyReadComplete.Sub(r.start).Nanoseconds(), Detail: phase}, StatusCode: trace.StatusCode, NoResponse: trace.StatusCode == 0, RequestStartElapsedNS: trace.RequestStart.Sub(r.start).Nanoseconds(), BodyReadElapsedNS: trace.BodyReadComplete.Sub(r.start).Nanoseconds()}
	if trace.Err != nil {
		payload.ErrorClass = "transport-or-body-read"
	}
	if trace.StatusCode != 0 && trace.Err == nil {
		body := trace.Body
		if trace.StatusCode >= 200 && trace.StatusCode < 300 {
			var env struct {
				Result json.RawMessage `json:"result"`
			}
			if err := json.Unmarshal(trace.Body, &env); err == nil && env.Result != nil {
				body = env.Result
			} else {
				payload.ErrorClass = "invalid-envelope"
			}
		}
		sum := sha256.Sum256(body)
		payload.DigestV1 = hex.EncodeToString(sum[:])
	}
	if err := r.appendLocked(payload); err != nil {
		return M0ReceiptStamp{}, err
	}
	stamp := r.stampLocked()
	if payload.ErrorClass == "invalid-envelope" {
		return stamp, fmt.Errorf("verifylive: causal receipt observed invalid 2xx envelope")
	}
	return stamp, nil
}

func (r *CausalReceipt) validateLeaseLocked(lease *CausalReceiptLease) error {
	if r.terminalErr != nil {
		return r.terminalErr
	}
	if r.activeLease != lease || !r.ready || r.f == nil || r.dir == nil {
		return errCausalReceiptNotReady
	}
	return nil
}

// m0CausalFieldsV1 is the versioned extracted identity. The only ID-shaped
// values are run-salted tags, never broker IDs.
type m0CausalFieldsV1 struct {
	Schema             int    `json:"schema"`
	ParentRequestTag   string `json:"parent_request_tag,omitempty"`
	ParentResponseTag  string `json:"parent_response_tag,omitempty"`
	PendingClientTag   string `json:"pending_client_tag,omitempty"`
	ParentClientTag    string `json:"parent_client_tag,omitempty"`
	ParentChildTag     string `json:"parent_child_tag,omitempty"`
	ChildCheckpointTag string `json:"child_checkpoint_tag,omitempty"`
	ChildRequestTag    string `json:"child_request_tag,omitempty"`
	ChildResponseTag   string `json:"child_response_tag,omitempty"`
	Symbol             string `json:"symbol,omitempty"`
	RequestedMarket    string `json:"requested_market,omitempty"`
	Market             string `json:"market,omitempty"`
	Type               string `json:"type,omitempty"`
	OrderType          string `json:"order_type,omitempty"`
	Quantity           string `json:"quantity,omitempty"`
	Side               string `json:"side,omitempty"`
	RootStatus         string `json:"root_status,omitempty"`
	FirstStatus        string `json:"first_status,omitempty"`
	Condition          string `json:"condition_type,omitempty"`
	Trigger            string `json:"trigger_price,omitempty"`
	Expiry             string `json:"expire_date,omitempty"`
	ChildStatus        string `json:"child_status,omitempty"`
	Currency           string `json:"currency,omitempty"`
	FilledQuantity     string `json:"filled_quantity,omitempty"`
	AverageFilledPrice string `json:"average_filled_price,omitempty"`
	FilledAmount       string `json:"filled_amount,omitempty"`
	FilledAt           string `json:"filled_at,omitempty"`
}

func (r *CausalReceipt) tag(id string) string {
	if r == nil || id == "" {
		return ""
	}
	r.mu.Lock()
	runID := r.runID
	r.mu.Unlock()
	s := sha256.Sum256([]byte(runID + "\x00" + id))
	return hex.EncodeToString(s[:])
}

// RecordCausal appends a sanitized durable barrier event under an active lease.
func (l *CausalReceiptLease) RecordCausal(kind string, fields m0CausalFieldsV1) (M0ReceiptStamp, error) {
	if l == nil || l.receipt == nil {
		return M0ReceiptStamp{}, errCausalReceiptNotReady
	}
	r := l.receipt
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.validateLeaseLocked(l); err != nil {
		return M0ReceiptStamp{}, err
	}
	r.seq++
	fields.Schema = 1
	payload := struct {
		M0ReceiptEvent
		Extracted m0CausalFieldsV1 `json:"extracted_v1"`
	}{M0ReceiptEvent: M0ReceiptEvent{Version: 1, Sequence: r.seq, Kind: kind, ElapsedNS: time.Since(r.start).Nanoseconds()}, Extracted: fields}
	if err := r.appendLocked(payload); err != nil {
		return M0ReceiptStamp{}, err
	}
	return r.stampLocked(), nil
}

func (r *CausalReceipt) stampLocked() M0ReceiptStamp {
	return M0ReceiptStamp{RunID: r.runID, Sequence: r.seq, FsyncDoneElapsedNS: time.Since(r.start).Nanoseconds()}
}
