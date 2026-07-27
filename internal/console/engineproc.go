package console

// engineproc.go is the engine's status line and its two buttons (openspec change
// add-engine-runtime, task 2.1).
//
// # What the console can and cannot do about the engine
//
// It can see whether one is running, and it can start and stop the process. It
// cannot make the engine able to trade. Whether the engine has order capability
// is decided by two things this package cannot reach: the automation gate in the
// operator's config, whose flip is a §0.7 human approval performed outside any
// browser, and the engine's own startup interlock, which re-checks every clause
// on every start. A [엔진 시작] pressed against an unmet interlock produces a
// process that refuses and exits — and this file's whole contribution to that
// case is to show the operator the refusal the engine itself printed.
//
// That is the difference between a console that drives the engine and a console
// that bypasses it, and it is asserted rather than described: the start seam
// returns the engine process's own reason, and the test that matters checks the
// reason reached the page.
//
// # Nothing here starts a process
//
// StartEngine and StopEngine are functions cmd/tossctl supplies, exactly as
// Relaunch and RestartSoak are. This package imports no process API at all
// (static_test.go), which is what lets the whole surface be tested with httptest
// and no fork.
//
// # The status is read from the advisory marker, not from the lock
//
// internal/enginelock has both. The lock is the exclusion and it is invisible —
// the only way to learn whether a flock is held is to try to take it, and a
// dashboard that did that several times a minute would be fighting the engine for
// its own lock. The marker is a timestamp, refreshed every minute, and reading it
// costs a stat. A crashed engine therefore shows as running for at most five
// minutes, which is the cost internal/runlock's package comment already accepted
// for the same trade.

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

// StartEngine spawns the engine process and reports what happened, in one line
// an operator reads.
//
// The error is the engine's own refusal where there is one — a gate that is off,
// an interlock clause that is unmet — because the operator's question after
// pressing the button is "why is it not running", and any sentence this package
// invented would be a second answer to it.
type StartEngine func() (string, error)

// StopEngine signals the running engine and waits for it to close its journal.
//
// It is the signal discipline and not a kill: the engine's loops are cancelled,
// allowed to finish the cycle they are in, and the journal is closed cleanly.
type StopEngine func() (string, error)

// Errors the engine routes produce before anything happens.
var (
	errNoEngineControl = errors.New(
		"이 빌드에는 엔진 기동/정지 배선이 없다. 터미널에서 `tossctl engine run`으로 직접 띄워라")
	errNoEngineStop = errors.New("이 빌드에는 엔진 정지 배선이 없다")
)

// engineView is the dashboard's engine section.
type engineView struct {
	// Marker is where the advisory marker was looked for. Empty means the status
	// half is unwired, which the page says rather than reporting "stopped".
	Marker string
	// Wired reports that the marker path was supplied at all.
	Wired bool
	// Running reports a marker touched inside the staleness window.
	Running bool
	// PID is the engine process the marker names, when it named one.
	PID int
	// StartedAt and RefreshedAt are the marker's own timestamps.
	StartedAt   time.Time
	RefreshedAt time.Time
	// Stale reports that the running engine was started from a different
	// executable than the one installed now. An unknown build is never stale.
	Stale bool
	// BuildAt renders the running engine's build timestamp.
	BuildAt string
	// StaleAfter is the freshness window, so the page can say how long a crashed
	// engine keeps reading as running.
	StaleAfter time.Duration

	// CanStart and CanStop report that the caller wired the seams.
	CanStart bool
	CanStop  bool

	// Note is the last thing a start or a stop said — including, for a refused
	// start, the interlock clauses the engine enumerated.
	Note   string
	NoteAt time.Time
}

// StartedAtText and RefreshedAtText render the marker's timestamps for the page.
func (v engineView) StartedAtText() string   { return renderInstant(v.StartedAt) }
func (v engineView) RefreshedAtText() string { return renderInstant(v.RefreshedAt) }
func (v engineView) NoteAtText() string      { return renderInstant(v.NoteAt) }

func renderInstant(at time.Time) string {
	if at.IsZero() {
		return "(알 수 없음)"
	}
	return at.UTC().Format("2006-01-02 15:04:05Z")
}

// readEngine reports what the marker says.
func (c *Console) readEngine(now time.Time, installed binstamp.Stamp) engineView {
	v := engineView{
		Marker:     c.opts.EngineMarker,
		Wired:      strings.TrimSpace(c.opts.EngineMarker) != "",
		StaleAfter: enginelock.StaleAfter,
		CanStart:   c.opts.StartEngine != nil,
		CanStop:    c.opts.StopEngine != nil,
	}
	v.Note, v.NoteAt = c.engineNoteNow()
	if !v.Wired {
		return v
	}
	status := enginelock.Read(c.opts.EngineMarker, now)
	v.Running = status.Running
	v.RefreshedAt = status.RefreshedAt
	v.PID = status.Marker.PID
	v.StartedAt = status.Marker.StartedAt
	v.Stale = status.Stale(installed)
	if status.Marker.Binary.Known() {
		v.BuildAt = status.Marker.Binary.ModTime.UTC().Format("2006-01-02 15:04:05Z")
	} else {
		v.BuildAt = "(알 수 없음)"
	}
	return v
}

// --- the two acts ------------------------------------------------------------------

// handleEngineStart asks cmd/tossctl to spawn the engine.
//
// A refusal is not an error page: the engine refusing to start is the *normal*
// outcome in a build with no broker-resident protective execution, and an
// operator who pressed the button is owed the reason on the screen they are
// already looking at.
func (c *Console) handleEngineStart(w http.ResponseWriter, r *http.Request) {
	if c.opts.StartEngine == nil {
		c.noteEngine(errNoEngineControl.Error())
		c.redirectDashboard(w, r, errNoEngineControl.Error())
		return
	}
	note, err := c.opts.StartEngine()
	if err != nil {
		reason := "엔진이 기동을 거부했다: " + err.Error()
		if strings.TrimSpace(note) != "" {
			reason += "\n" + note
		}
		c.noteEngine(reason)
		c.redirectDashboard(w, r, "엔진이 기동을 거부했다 — 사유는 아래 엔진 섹션에 있다")
		return
	}
	if strings.TrimSpace(note) == "" {
		note = "엔진을 시작했다."
	}
	c.noteEngine(note)
	c.redirectDashboard(w, r, note)
}

// handleEngineStop signals the running engine.
func (c *Console) handleEngineStop(w http.ResponseWriter, r *http.Request) {
	if c.opts.StopEngine == nil {
		c.noteEngine(errNoEngineStop.Error())
		c.redirectDashboard(w, r, errNoEngineStop.Error())
		return
	}
	note, err := c.opts.StopEngine()
	if err != nil {
		reason := "엔진 정지 실패: " + err.Error()
		c.noteEngine(reason)
		c.redirectDashboard(w, r, reason)
		return
	}
	if strings.TrimSpace(note) == "" {
		note = "엔진을 정지시켰다."
	}
	c.noteEngine(note)
	c.redirectDashboard(w, r, note)
}

// noteEngine records what the last act said, for the dashboard's engine section.
func (c *Console) noteEngine(text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.engineNote = text
	c.engineNoteAt = c.now()
}

func (c *Console) engineNoteNow() (string, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.engineNote, c.engineNoteAt
}
