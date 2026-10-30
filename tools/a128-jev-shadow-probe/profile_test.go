package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestInitialStateProjectionIsIndependent(t *testing.T) {
	pct := 3.0
	source := PublicState{Market: "KR", Symbol: "ABC", ChangePctVsReference: &pct, TopOfBook: &TopOfBookState{BestAsk: 10}, News: newsPresent, Intraday: &IntradayState{}, MissingInputs: []string{"stocks_meta: unavailable", "intraday: unavailable", "unknown: private"}}
	got := initialState(source)
	if got.Intraday != nil || got.News != newsAbsent || !reflect.DeepEqual(got.MissingInputs, []string{"stocks_meta: unavailable", "news: intentionally omitted by initial-fields-v1"}) {
		t.Fatalf("unexpected projection: %+v", got)
	}
	*got.ChangePctVsReference = 4
	got.TopOfBook.BestAsk = 20
	got.MissingInputs[0] = "changed"
	if pct != 3 || source.TopOfBook.BestAsk != 10 || source.MissingInputs[0] != "stocks_meta: unavailable" {
		t.Fatal("projection aliases source")
	}
}

func TestProfileOutputGuard(t *testing.T) {
	if err := validateProfileOutput("invalid", "x", true); err == nil {
		t.Fatal("invalid accepted")
	}
	if err := validateProfileOutput("compare", defaultLedgerDir, true); err == nil {
		t.Fatal("default accepted")
	}
	if err := validateProfileOutput("initial", "elsewhere", false); err == nil {
		t.Fatal("implicit output accepted")
	}
	if err := validateProfileOutput("compare", t.TempDir(), true); err != nil {
		t.Fatal(err)
	}
}

func TestProfileFlagsRejectedBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{{"--input-profile", "bogus"}, {"--input-profile", "initial", "--out", defaultLedgerDir}} {
		err := runCommand(context.Background(), args, func(string) string { return "" })
		if err == nil || strings.Contains(err.Error(), typeSafeAPIKeyEnv) {
			t.Fatalf("flags %v not rejected before credentials: %v", args, err)
		}
	}
}

func TestProfileOutputGuardRejectsSymlinkAlias(t *testing.T) {
	dir := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(defaultLedgerDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(defaultLedgerDir, "alias"); err != nil {
		t.Fatal(err)
	}
	if err := validateProfileOutput("compare", filepath.Join("alias", "child"), true); err == nil {
		t.Fatal("symlink alias accepted")
	}
}

func TestSelectionLotSeparatesProfiles(t *testing.T) {
	base := judgmentRow{StateRev: "same"}
	current := base
	current.InputProfile = "current"
	initial := base
	initial.InputProfile = "initial"
	if selectionLot(current) == selectionLot(initial) {
		t.Fatal("profiles mixed")
	}
	if selectionLot(base) != legacySelectionLot+"@same" {
		t.Fatalf("legacy changed: %s", selectionLot(base))
	}
}

func TestProfileOutputGuardRejectsLegacyLedger(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ledgerFileName), []byte("{\"kind\":\"judgment\",\"id\":\"old\"}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateProfileOutput("compare", dir, true); err == nil {
		t.Fatal("legacy ledger accepted")
	}
}

func TestProfileOutputGuardRejectsDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(filepath.Join(dir, "missing"), alias); err != nil {
		t.Fatal(err)
	}
	if err := validateProfileOutput("compare", filepath.Join(alias, "child"), true); err == nil {
		t.Fatal("dangling alias accepted")
	}
}

func TestAutomaticProfileOutputRouting(t *testing.T) {
	for _, test := range []struct{ profile, want string }{{"compare", defaultLedgerDir + "-compare"}, {"initial", defaultLedgerDir + "-initial"}, {"current", defaultLedgerDir}} {
		got, err := resolveProfileOutput(test.profile, "", false)
		if err != nil || got != test.want {
			t.Fatalf("%s => %s, %v", test.profile, got, err)
		}
	}
	if _, err := resolveProfileOutput("compare", defaultLedgerDir, true); err == nil {
		t.Fatal("explicit legacy output accepted")
	}
}

func TestManualSessionExperimentCanRestart(t *testing.T) {
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := newProbe(nil, nil, l, config{InputProfile: "compare", Cohort: "run", ManualWindow: &manualWindow{StartHour: 9, EndHour: 15, EndMin: 30, Raw: "09:00-15:30"}})
	p.now = func() time.Time { return at(10, 0) }
	if _, err := p.manualSession("KR"); err != nil {
		t.Fatal(err)
	}
	l.Close()
	if err := validateProfileOutput("compare", dir, true); err != nil {
		t.Fatalf("own session gap rejected after restart: %v", err)
	}
}
