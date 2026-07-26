package console

// data.go reads the three local files the dashboard reports on.
//
// Every read is per-request and every failure is a state rather than an error:
// "there is no soak record yet" and "the attestation is missing" are what a
// machine looks like before the work is done, and they are exactly what an
// operator opens this console to find out. A page that 500'd on a missing file
// would be a page that only works once it is no longer needed.
//
// Nothing here writes. Nothing here reaches the network. The account references
// are masked on the way through, because a console page is the kind of thing that
// ends up in a screenshot.

import (
	"errors"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// soakView is the soak record, read.
type soakView struct {
	Record  string
	Present bool
	Error   string

	AccountRef string
	Cycles     int
	StreakDays int
	MinDays    int
	FirstAt    time.Time
	LastAt     time.Time
	Ready      bool
	Reasons    []string
	Days       []soak.Day
	// Binary is the executable the survey said it was running, as of its last
	// recorded cycle. The zero stamp means the record predates the survey
	// stamping itself, which reads as unknown and never as stale.
	Binary binstamp.Stamp
}

// attestView is the capability attestation, read.
type attestView struct {
	Path    string
	Present bool
	// Usable reports that the file exists, parses, has not expired and names
	// everything an attestation has to name. It is deliberately NOT "the gate
	// would open": that answer needs the live account the credentials resolve to,
	// which this console never asks for.
	Usable  bool
	Reasons []string

	AccountRef string
	SoakDays   int
	IssuedAt   time.Time
	ExpiresAt  time.Time
	Endpoints  []string
	// Missing are the engine's required endpoints this attestation does not cover.
	Missing []string
}

// verifyView is the live verification's evidence record, read.
type verifyView struct {
	Record  string
	Present bool
	Error   string

	AccountRef string
	Steps      []verifylive.Outcome
	Pending    []verifylive.StepID
	Done       int
	Total      int
	// Resume reports that a verification is already on the record, so a run
	// started now continues it rather than beginning one.
	Resume bool
	// Redo are the steps whose newest verdict is fail or skipped: the ones a
	// re-measurement could still change. It is verifylive.RedoSet's answer, read
	// from the record on every request rather than remembered, because the record
	// is what the runner will act on.
	Redo []verifylive.StepID
	// AwaitingRestart names the step that is waiting for a new process.
	AwaitingRestart verifylive.StepID
	Outstanding     []verifylive.Artifact
}

// RedoCount is what the start screen prints before anything is sent.
func (v verifyView) RedoCount() int { return len(v.Redo) }

// RedoList is the set as one readable line.
func (v verifyView) RedoList() string {
	parts := make([]string, 0, len(v.Redo))
	for _, id := range v.Redo {
		parts = append(parts, string(id))
	}
	return strings.Join(parts, ", ")
}

// binaryView is the stale-binary reading for both long-running processes.
//
// The question is the same one twice — "is the process still the installed build?"
// — and the two halves answer it from different evidence because the console can
// see itself and cannot see the soak:
//
//	console   its own fingerprint, taken once at startup, against the file at that
//	          path now. Nothing to ask anybody: this process knows what it loaded.
//	soak      the fingerprint the survey stamps on every cycle it records. It is
//	          the survey's own statement about itself as of its last cycle, which
//	          is the simplest honest source: no process table to scrape, no pid to
//	          guess at, no second definition of "the soak" to keep in step.
//
// A soak record written by a build from before this existed carries no stamp, and
// that reads as unknown rather than as stale — the same direction binstamp takes
// everywhere.
type binaryView struct {
	// Installed is the executable on disk right now.
	Installed binstamp.Stamp
	// Console is what this process was loaded from.
	Console binstamp.Stamp
	// Soak is what the survey said it was running, at its last recorded cycle.
	Soak binstamp.Stamp
	// SoakAt is when that cycle was.
	SoakAt time.Time
}

// ConsoleStale reports that this console is running an older build than the one
// installed.
func (v binaryView) ConsoleStale() bool { return !v.Console.Same(v.Installed) }

// SoakStale reports the same about the survey.
func (v binaryView) SoakStale() bool { return v.Soak.Known() && !v.Soak.Same(v.Installed) }

// InstalledAt renders the installed build's timestamp for the page.
func (v binaryView) InstalledAt() string {
	if !v.Installed.Known() {
		return "(알 수 없음)"
	}
	return v.Installed.ModTime.Format("2006-01-02 15:04:05Z")
}

// ConsoleAt renders this process's build timestamp.
func (v binaryView) ConsoleAt() string {
	if !v.Console.Known() {
		return "(알 수 없음)"
	}
	return v.Console.ModTime.Format("2006-01-02 15:04:05Z")
}

// SoakBuildAt renders the survey's build timestamp.
func (v binaryView) SoakBuildAt() string {
	if !v.Soak.Known() {
		return "(알 수 없음)"
	}
	return v.Soak.ModTime.Format("2006-01-02 15:04:05Z")
}

// snapshot is the whole dashboard.
type snapshot struct {
	Now         time.Time
	Soak        soakView
	Attestation attestView
	Verify      verifyView
	Binary      binaryView
	// Session is the KR market-hours reading. It is advisory: nothing on this
	// console consults it before starting anything.
	Session verifylive.SessionAdvisory
}

func (c *Console) snapshot() snapshot {
	now := c.now()
	soakView := c.readSoak(now)
	return snapshot{
		Now:         now,
		Soak:        soakView,
		Attestation: c.readAttestation(now),
		Verify:      c.readVerify(),
		Binary:      c.readBinary(soakView),
		Session:     verifylive.KRSessionAdvisory(now),
	}
}

// readBinary compares the installed executable against the two processes that are
// meant to be it.
func (c *Console) readBinary(sv soakView) binaryView {
	v := binaryView{Console: c.startedWith, Soak: sv.Binary, SoakAt: sv.LastAt}
	if c.opts.Binary != nil {
		v.Installed, _ = c.opts.Binary()
	}
	return v
}

func (c *Console) readSoak(now time.Time) soakView {
	v := soakView{Record: c.opts.SoakRecord, MinDays: c.opts.MinSoakDays}
	if strings.TrimSpace(v.Record) == "" {
		return v
	}
	cycles, err := soak.LoadCycles(v.Record)
	if err != nil {
		v.Error = err.Error()
		return v
	}
	if len(cycles) == 0 {
		return v
	}

	criteria := soak.DefaultCriteria()
	if v.MinDays > 0 {
		criteria.MinConsecutiveDays = v.MinDays
	}
	v.MinDays = criteria.MinConsecutiveDays

	summary := soak.Summarize(cycles)
	v.Present = true
	v.AccountRef = attest.Mask(summary.AccountRef)
	v.Cycles = summary.All.Cycles
	v.StreakDays = summary.StreakDays
	v.FirstAt, v.LastAt = summary.FirstAt, summary.LastAt
	v.Days = summary.Days
	v.Binary = summary.Binary
	v.Ready, v.Reasons = summary.Evaluate(now, criteria)
	return v
}

func (c *Console) readAttestation(now time.Time) attestView {
	v := attestView{Path: c.opts.Attestation}
	if strings.TrimSpace(v.Path) == "" {
		return v
	}
	a, err := attest.Load(v.Path)
	if err != nil {
		if !errors.Is(err, attest.ErrMissing) {
			v.Reasons = append(v.Reasons, err.Error())
		}
		return v
	}

	v.Present = true
	v.AccountRef = attest.Mask(a.AccountRef)
	v.SoakDays = a.SoakDays
	v.IssuedAt, v.ExpiresAt = a.IssuedAt, a.ExpiresAt
	v.Endpoints = a.Endpoints
	v.Missing = a.MissingEndpoints(c.opts.RequiredEndpoints)

	switch {
	case strings.TrimSpace(a.AccountRef) == "":
		v.Reasons = append(v.Reasons, "계좌를 명시하지 않았다")
	case len(a.Endpoints) == 0:
		v.Reasons = append(v.Reasons, "검증된 endpoint가 하나도 없다")
	}
	if a.Expired(now) {
		v.Reasons = append(v.Reasons, "만료되었다 — soak을 다시 돌리고 `tossctl soak attest`를 다시 실행하라")
	}
	for _, missing := range v.Missing {
		v.Reasons = append(v.Reasons, "미검증 endpoint: "+missing)
	}
	v.Usable = len(v.Reasons) == 0
	return v
}

func (c *Console) readVerify() verifyView {
	v := verifyView{Record: c.opts.VerifyRecord, Total: len(verifylive.Steps())}
	if strings.TrimSpace(v.Record) == "" {
		return v
	}
	entries, err := verifylive.LoadEntries(v.Record)
	if err != nil {
		v.Error = err.Error()
		return v
	}

	progress := verifylive.BuildProgress(v.Record, entries)
	v.Present = verifylive.StepCount(entries) > 0
	v.Resume = v.Present
	v.Redo = verifylive.RedoSet(entries)
	v.AccountRef = progress.AccountRef
	v.Steps = progress.Steps
	v.Pending = progress.Pending
	v.AwaitingRestart = progress.AwaitingRestart
	v.Outstanding = progress.Outstanding
	for _, s := range progress.Steps {
		if s.Verdict.Terminal() {
			v.Done++
		}
	}
	return v
}

// redoSet re-reads the record and returns the steps a re-measurement may touch.
//
// It is read here, from the file, and never taken from the request. A step list
// that arrived in a form field would let anything that can reach this socket name
// a step to re-run — including one that already passed, which is the single thing
// the redo path must not do. The form says "redo" and nothing else; what that
// means is decided by verifylive.RedoSet against the evidence on disk.
func (c *Console) redoSet() ([]verifylive.StepID, error) {
	entries, err := verifylive.LoadEntries(c.opts.VerifyRecord)
	if err != nil {
		return nil, err
	}
	return verifylive.RedoSet(entries), nil
}

// report renders the verify report the report page and its JSON download share.
func (c *Console) report() (verifylive.Report, error) {
	entries, err := verifylive.LoadEntries(c.opts.VerifyRecord)
	if err != nil {
		return verifylive.Report{}, err
	}
	return verifylive.BuildReport(c.opts.VerifyRecord, entries, c.now()), nil
}
