// Package audit is the append-only record of operational setting changes.
//
// WORKFLOW §0.5: "운영 설정(위험 한도·레인 ON/OFF·운영 모드) 변경은 audit 로그로
// 추적 가능해야 한다", and engine-safety restates it for the automation gate:
// before-value, after-value, time, subject.
//
// # Why a file and not the journal database
//
// The journal is the order record and it refuses to open on a filesystem where
// fsync is not trustworthy — which is correct for orders and wrong for an audit
// trail, because the situation where you most want to know who changed a limit is
// the one where the engine refused to start. A plain append-only JSONL file has
// no such precondition, can be read with `tail`, and survives a journal that has
// to be moved or rebuilt.
//
// # What "append-only" means here
//
// Every write is O_APPEND plus an fsync, and nothing in this package can update
// or delete a line. Rewriting history is not an operation the type system offers.
// The file is 0600 because it names accounts and limits.
package audit

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// FileName is the audit log's default name inside the data directory.
const FileName = "audit.log"

// Action names a category of change. They are stable strings: an operator greps
// for them and Phase 2's reporting reads them.
const (
	// ActionGateToggle is the automation gate being turned on or off.
	ActionGateToggle = "automation_gate.toggle"
	// ActionLimitChange is a Guardian limit being changed.
	ActionLimitChange = "automation_gate.limit"
	// ActionGateRefused is a startup the interlock refused. Recorded because a
	// refused start with the gate on is exactly as interesting as a successful
	// one, and it is the only trace of an attempt that produced no engine.
	ActionGateRefused = "automation_gate.refused"
	// ActionGateAccepted is a startup the interlock allowed.
	ActionGateAccepted = "automation_gate.accepted"
	// ActionAdoptionToggle is external-position adoption being turned on or off
	// (change adopt-external-positions, §0.5/§0.7).
	//
	// It is a separate action from the gate's toggle because it authorises a
	// different thing: the gate decides whether the engine may place orders
	// unattended at all, and this decides whether it starts placing them for
	// positions its owner bought by hand. An operator auditing "when did the
	// engine start managing my long-term holdings" has to be able to find that
	// answer without reading it out of a gate entry.
	ActionAdoptionToggle = "adoption.toggle"
	// ActionAdoptionSetting is a change to how adoption behaves once on: the
	// synthetic stop's width, or the exclusion list.
	ActionAdoptionSetting = "adoption.setting"
)

// Entry is one recorded change.
//
// Old and New are strings rather than `any` so that a record written today reads
// identically in five years: a float that JSON round-trips through a different
// precision is a diff nobody can explain, and the audit log's whole value is that
// its lines mean exactly one thing.
type Entry struct {
	At      time.Time `json:"at"`
	Subject string    `json:"subject"`
	Action  string    `json:"action"`
	Setting string    `json:"setting"`
	Old     string    `json:"old"`
	New     string    `json:"new"`
	Detail  string    `json:"detail,omitempty"`
}

// Log is an open audit log.
type Log struct {
	path string
	clk  clock.Clock
	subj string

	mu sync.Mutex
}

// Options configures Open.
type Options struct {
	// Path is the log file. Empty is an error: the caller decides where it lives,
	// because the engine and the CLI resolve their data directories differently.
	Path string
	// Clock stamps entries. Defaults to clock.System().
	Clock clock.Clock
	// Subject names who is making the changes. Empty resolves the OS user.
	Subject string
}

// Open prepares the log, creating the directory and the file if needed.
func Open(opts Options) (*Log, error) {
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		return nil, errors.New("audit: a log path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("audit: creating %s: %w", filepath.Dir(path), err)
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	subject := strings.TrimSpace(opts.Subject)
	if subject == "" {
		subject = osSubject()
	}
	return &Log{path: path, clk: clk, subj: subject}, nil
}

// Path returns the log file path.
func (l *Log) Path() string { return l.path }

// Record appends one entry. At and Subject are filled in when empty.
//
// The fsync is not optional: an audit line that is still in the page cache when
// the machine loses power is an audit line that never existed, and the change it
// describes may well have reached the broker.
func (l *Log) Record(e Entry) error {
	if l == nil {
		return nil
	}
	if e.At.IsZero() {
		e.At = l.clk.Now().UTC()
	} else {
		e.At = e.At.UTC()
	}
	if strings.TrimSpace(e.Subject) == "" {
		e.Subject = l.subj
	}

	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: encoding an entry: %w", err)
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("audit: opening %s: %w", l.path, err)
	}
	defer f.Close()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("audit: appending to %s: %w", l.path, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("audit: flushing %s: %w", l.path, err)
	}
	return nil
}

// RecordChange appends an entry only when the value actually changed.
//
// It reports whether it wrote. Recording an unchanged setting on every startup
// would bury the one line that matters under thousands that do not, and an audit
// log nobody can read is not an audit log.
func (l *Log) RecordChange(action, setting, newValue, detail string) (bool, error) {
	if l == nil {
		return false, nil
	}
	previous, found, err := l.Latest(setting)
	if err != nil {
		return false, err
	}
	old := ""
	if found {
		old = previous.New
	}
	if found && old == newValue {
		return false, nil
	}
	return true, l.Record(Entry{
		Action:  action,
		Setting: setting,
		Old:     old,
		New:     newValue,
		Detail:  detail,
	})
}

// RecordAction appends an entry unconditionally.
//
// It is the counterpart to RecordChange for events rather than settings. A
// setting has a previous value worth deduplicating against; an action — an
// operator releasing a risk reservation, say — happened, and a second identical
// one is a second event, not a repeat of the first. Collapsing those would hide
// exactly the pattern (the same override, twice, minutes apart) that a reviewer
// is looking for.
func (l *Log) RecordAction(action, setting, value, detail string) error {
	if l == nil {
		return nil
	}
	return l.Record(Entry{
		Action:  action,
		Setting: setting,
		New:     value,
		Detail:  detail,
	})
}

// Latest returns the most recent entry for a setting.
func (l *Log) Latest(setting string) (Entry, bool, error) {
	entries, err := l.Entries()
	if err != nil {
		return Entry{}, false, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Setting == setting {
			return entries[i], true, nil
		}
	}
	return Entry{}, false, nil
}

// Entries reads the whole log, oldest first.
//
// A line that will not parse is skipped rather than failing the read: a truncated
// final line after a power loss must not make the preceding history unreadable.
func (l *Log) Entries() ([]Entry, error) {
	f, err := os.Open(l.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: reading %s: %w", l.path, err)
	}
	defer f.Close()

	var out []Entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		out = append(out, e)
	}
	if err := scanner.Err(); err != nil {
		return out, fmt.Errorf("audit: scanning %s: %w", l.path, err)
	}
	return out, nil
}

// osSubject identifies the operating-system user, so an unattended change still
// carries a name.
func osSubject() string {
	if u, err := user.Current(); err == nil && strings.TrimSpace(u.Username) != "" {
		return u.Username
	}
	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name
	}
	return "unknown"
}
