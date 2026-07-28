package main

// candidatepanel.go builds the discovery source panel from whatever credentials
// exist. It is one function and a converter, and the reason it is not in
// candidate.go beside its two callers is the consumer guard.
//
// # Why the client is constructed here and not next to the command
//
// internal/candidate/consumer_guard_test.go fixes the list of files that may name a
// chase verdict, and then checks that none of those files can also name an order
// verb. The check reads what a file says, because inside cmd/tossctl — one Go
// package that already imports internal/execgw, internal/orderintent and
// internal/trading — an import graph says nothing at all.
//
// A selector scan sees `official.Place…`. It does not see
// `client.CreateConditionalOrder(…)`, because by then the package name is gone and
// what is left is a method on a value. So the thing the guard can actually forbid in
// a verdict-reading file is the construction: `official.New` and `tossclient.New`
// are the two lines that turn credentials into a value carrying every write the
// backend has. candidate.go renders the veto block, so it must not contain them, and
// this file — which names no verdict — is where they live.
//
// The narrowing that reaches this package's own callers is separate and older: what
// crosses into internal/candidate is candidatesrc's RankingReader and
// PopularityReader, each one method wide. This file is about what a *file* can say,
// not about what a package can be handed.

import (
	"context"
	"fmt"
	"os"

	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc"
)

// buildCandidatePanel assembles the market's sources from whatever credentials
// exist.
//
// The official rankings are required and the WTS popularity list is additive, which
// is spec Requirement 5's hard half: WTS sessions expire on ordinary days, so a
// configuration that stops without one stops routinely.
//
// The clock is handed in rather than taken inside candidatesrc because each source
// now dates the reading it remembers: the new-entrant fact is only an answer while
// the reading it is measured against is recent enough to be the previous one, and a
// source that cannot date its own memory cannot tell a tick from an outage.
func buildCandidatePanel(root *rootOptions, market string) ([]candidate.Source, error) {
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading the Open API credentials: %w", err)
	}
	if creds == nil {
		return nil, fmt.Errorf(
			"candidate: no Open API credentials. The official rankings are discovery's " +
				"required source — run `tossctl openapi login`")
	}
	client := official.New(*creds, tokenFile)

	// The WTS half is optional and its absence is ordinary. A missing or unreadable
	// session degrades the panel; it does not stop it.
	var wts *tossclient.Client
	if sess, serr := session.NewFileStore(resolveSessionFile(root)).Load(context.Background()); serr == nil && sess != nil {
		if cfg, cerr := loadConfig(root); cerr == nil {
			wts = tossclient.New(tossclient.Config{Session: sess, TradingPolicy: cfg.Trading})
		}
	}
	return candidatesrc.Panel(market, client, client, wtsPopularityReader(wts), clock.System()), nil
}

// wtsPopularityReader converts a possibly-nil client into a possibly-nil reader.
//
// A typed nil inside a non-nil interface is the classic way for "no WTS session" to
// become a nil dereference three layers down, and Panel's contract is that a nil
// reader means the source is absent from the panel.
func wtsPopularityReader(c *tossclient.Client) candidatesrc.PopularityReader {
	if c == nil {
		return nil
	}
	return c
}
