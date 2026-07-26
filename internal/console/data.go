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
	// AwaitingRestart names the step that is waiting for a new process.
	AwaitingRestart verifylive.StepID
	Outstanding     []verifylive.Artifact
}

// snapshot is the whole dashboard.
type snapshot struct {
	Now         time.Time
	Soak        soakView
	Attestation attestView
	Verify      verifyView
}

func (c *Console) snapshot() snapshot {
	now := c.now()
	return snapshot{
		Now:         now,
		Soak:        c.readSoak(now),
		Attestation: c.readAttestation(now),
		Verify:      c.readVerify(),
	}
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

// report renders the verify report the report page and its JSON download share.
func (c *Console) report() (verifylive.Report, error) {
	entries, err := verifylive.LoadEntries(c.opts.VerifyRecord)
	if err != nil {
		return verifylive.Report{}, err
	}
	return verifylive.BuildReport(c.opts.VerifyRecord, entries, c.now()), nil
}
