package main

// verify_test.go covers the CLI surface of `tossctl verify` (task 1.5).
//
// The procedure itself is tested in internal/verifylive, against a fake broker
// and against an httptest server. What is tested here is what only the command
// can get wrong:
//
//	the flags it must not offer   nothing that could answer a confirmation
//	the terminal requirement      with no TTY, not one mutating request is issued,
//	                              watched at the HTTP layer against a fake API
//	the refusal without creds     orders go through the official API or not at all
//	the record's location         next to the soak's, and it follows --config-dir
//	--list                        reads nothing, sends nothing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- a fake official API -------------------------------------------------------

type verifyServer struct {
	*httptest.Server

	mu       sync.Mutex
	requests []string
}

func newVerifyServer(t *testing.T) *verifyServer {
	t.Helper()
	s := &verifyServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r.Method+" "+r.URL.Path)
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"verify-token","expires_in":3600,"token_type":"Bearer"}`)
		case r.URL.Path == "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		case r.URL.Path == "/api/v1/holdings":
			fmt.Fprint(w, `{"result":{"items":[{"symbol":"005930","quantity":"2","lastPrice":"70000"}]}}`)
		case r.URL.Path == "/api/v1/sellable-quantity":
			fmt.Fprint(w, `{"result":{"sellableQuantity":"2"}}`)
		case r.URL.Path == "/api/v1/prices":
			fmt.Fprint(w, `{"result":[{"symbol":"005930","lastPrice":"70000","currency":"KRW"}]}`)
		case r.URL.Path == "/api/v1/price-limits":
			fmt.Fprint(w, `{"result":{"upperLimitPrice":"91000","lowerLimitPrice":"49000","currency":"KRW"}}`)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"result":{"orders":[],"nextCursor":"","hasNext":false}}`)
		case r.URL.Path == "/api/v1/conditional-orders" && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"result":{"conditionalOrders":[],"nextCursor":"","hasNext":false}}`)
		default:
			// Anything else is a mutation this test never wants to see; answering
			// it at all would hide the bug rather than surface it.
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func (s *verifyServer) seen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

// pointVerifyAt wires the command at srv for the duration of the test.
//
// It replaces the factory rather than adding a --base-url flag: the record this
// command writes is the evidence an attestation is built from, and evidence
// produced against a server of one's own choosing is not evidence.
func pointVerifyAt(t *testing.T, srv *verifyServer, tokenFile string) {
	t.Helper()
	previous := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		client := official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			tokenFile,
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
			official.WithAccountSeq(7),
		)
		return client, "123-45-678901", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })
}

// --- the flags that must not exist ----------------------------------------------

// TestVerifyRunOffersNoAutomationFlag is tasks.md's requirement as a test: the
// mutations are gated on "flatten-all 패턴의 TTY typed-confirmation(자동화 플래그
// 금지)".
func TestVerifyRunOffersNoAutomationFlag(t *testing.T) {
	cmd := findCommand(findCommand(newRootCmd(), "verify"), "run")
	if cmd == nil {
		t.Fatal("verify run is not registered")
	}
	banned := []string{
		"yes", "y", "force", "f", "confirm", "assume-yes", "no-confirm",
		"non-interactive", "batch", "auto-approve", "skip-confirmation", "unattended",
	}
	for _, name := range banned {
		if cmd.Flags().Lookup(name) != nil {
			t.Errorf("verify run offers --%s; every mutation must be typed by a person", name)
		}
		if len(name) == 1 && cmd.Flags().ShorthandLookup(name) != nil {
			t.Errorf("verify run offers -%s; every mutation must be typed by a person", name)
		}
	}
	for _, banned := range []string{"base-url", "api-base-url", "endpoint", "insecure"} {
		if cmd.Flags().Lookup(banned) != nil {
			t.Errorf("verify run offers --%s; the tool that produces the evidence must not be "+
				"redirectable at an arbitrary server", banned)
		}
	}
}

// TestVerifyRunOffersTheKnobsTheProcedureNeeds.
func TestVerifyRunOffersTheKnobsTheProcedureNeeds(t *testing.T) {
	cmd := findCommand(findCommand(newRootCmd(), "verify"), "run")
	if cmd == nil {
		t.Fatal("verify run is not registered")
	}
	for _, name := range []string{
		"list", "resume", "symbol", "holding-symbol", "offset-pct",
		"max-sell-quantity", "include-ttl-edge", "confirm-each", "ttl-wait", "redo", "record",
	} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("verify run has no --%s", name)
		}
	}
	if f := cmd.Flags().Lookup("include-ttl-edge"); f != nil && f.DefValue != "false" {
		t.Errorf("--include-ttl-edge defaults to %q; the step it unlocks creates a second live order",
			f.DefValue)
	}
	// --confirm-each opts INTO more prompting, never out of it, so the default has
	// to be the batch approval rather than the finer gate.
	if f := cmd.Flags().Lookup("confirm-each"); f != nil && f.DefValue != "false" {
		t.Errorf("--confirm-each defaults to %q, want false — it is an opt-in to per-mutation prompts",
			f.DefValue)
	}
	if f := cmd.Flags().Lookup("offset-pct"); f != nil && f.DefValue != "20" {
		t.Errorf("--offset-pct defaults to %q, want the conservative 20", f.DefValue)
	}
}

func TestVerifyCommandsAreRegisteredAndAnnotated(t *testing.T) {
	parent := findCommand(newRootCmd(), "verify")
	if parent == nil {
		t.Fatal("verify is not registered on the root command")
	}
	want := map[string]struct {
		source   string
		mutating bool
	}{
		"run":    {"official", true},
		"status": {"local", false},
		"report": {"local", false},
	}
	for name, expect := range want {
		sub := findCommand(parent, name)
		if sub == nil {
			t.Errorf("verify %s is not registered", name)
			continue
		}
		if got := sub.Annotations["source"]; got != expect.source {
			t.Errorf("verify %s: source = %q, want %q", name, got, expect.source)
		}
		if got := sub.Annotations["mutating"] == "true"; got != expect.mutating {
			t.Errorf("verify %s: mutating = %v, want %v", name, got, expect.mutating)
		}
		if strings.TrimSpace(sub.Short) == "" {
			t.Errorf("verify %s has no Short description", name)
		}
	}
}

// --- the terminal requirement ----------------------------------------------------

// TestVerifyRunIssuesNoMutatingRequestWithoutATerminal.
//
// This is the safety property watched at the HTTP layer. The test binary's stdin
// is not a terminal, so every confirmation is refused before anything is sent —
// and the fake API must therefore see nothing but GETs and the OAuth exchange.
func TestVerifyRunIssuesNoMutatingRequestWithoutATerminal(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	out, _, err := runCLI(t, "--config-dir", configDir, "verify", "run")
	if err == nil {
		t.Log("the run returned no error; the assertions below are about what it sent, not what it returned")
	}

	seen := srv.seen()
	if len(seen) == 0 {
		t.Fatal("the verification issued no request at all; it cannot have started")
	}
	sawToken := false
	for _, req := range seen {
		method, path, _ := strings.Cut(req, " ")
		if method == http.MethodPost && path == "/oauth2/token" {
			sawToken = true
			continue
		}
		if method != http.MethodGet {
			t.Errorf("the verification issued %s with no terminal to confirm it", req)
		}
	}
	if !sawToken {
		t.Error("no token exchange happened; the run cannot have reached the API")
	}
	if strings.Contains(out, "123-45-678901") {
		t.Error("the output printed the unmasked account number")
	}
}

// TestVerifyRunRefusesWithoutOfficialCredentials. Orders go through the official
// Open API or not at all; the web session is never used to place one.
func TestVerifyRunRefusesWithoutOfficialCredentials(t *testing.T) {
	configDir := testenv.Isolate(t)

	_, _, err := runCLI(t, "--config-dir", configDir, "verify", "run")
	if err == nil {
		t.Fatal("verify run started without official credentials")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "credential") {
		t.Errorf("err = %v, want it to name the missing credentials", err)
	}
}

// TestVerifyRunRefusesToRestartOverAnExistingRecord. Starting over silently would
// place live orders for measurements already made.
func TestVerifyRunRefusesToRestartOverAnExistingRecord(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))
	seedVerifyRecord(t, filepath.Join(configDir, verifylive.FileName))

	_, _, err := runCLI(t, "--config-dir", configDir, "verify", "run")
	if err == nil {
		t.Fatal("verify run started over an existing record without --resume")
	}
	for _, want := range []string{"--resume", "verify status"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to mention %q", err, want)
		}
	}
	for _, req := range srv.seen() {
		if strings.HasPrefix(req, "POST /api/v1/orders") {
			t.Errorf("an order was placed before the record check: %s", req)
		}
	}
}

// --- --list ----------------------------------------------------------------------

// TestVerifyRunListNeedsNoCredentialsAndSendsNothing. An operator has to be able
// to read the whole procedure before deciding to run any of it.
func TestVerifyRunListNeedsNoCredentialsAndSendsNothing(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	out, _, err := runCLI(t, "--config-dir", configDir, "verify", "run", "--list")
	if err != nil {
		t.Fatalf("verify run --list: %v", err)
	}
	if len(srv.seen()) != 0 {
		t.Errorf("--list issued requests: %v", srv.seen())
	}
	for _, s := range verifylive.Steps() {
		if !strings.Contains(out, string(s.ID)) {
			t.Errorf("--list omits %s", s.ID)
		}
	}
	for _, want := range []string{"one share", "typed confirmation", "--include-ttl-edge"} {
		if !strings.Contains(out, want) {
			t.Errorf("--list does not mention %q:\n%s", want, out)
		}
	}
	// The approval model, which is what an operator decides on before running
	// anything: one string for the whole run, and a list that is a limit.
	for _, want := range []string{
		"ONE typed, expiring confirmation for the whole run",
		"--confirm-each",
		"stops and asks for a fresh approval",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("--list does not describe the batch approval (%q missing):\n%s", want, out)
		}
	}
}

// TestVerifyRunAsksForTheBatchApprovalBeforeAnythingIsSent.
//
// The test binary has no terminal, so the approval cannot be typed — and the
// account must therefore see no mutating request at all, not even from the steps
// that would have run before the first one. Checked at the HTTP layer, which is
// the only place the claim is worth anything.
func TestVerifyRunAsksForTheBatchApprovalBeforeAnythingIsSent(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	out, _, _ := runCLI(t, "--config-dir", configDir, "verify", "run")

	for _, req := range srv.seen() {
		method, path, _ := strings.Cut(req, " ")
		if method == http.MethodPost && path == "/oauth2/token" {
			continue
		}
		if method != http.MethodGet {
			t.Errorf("the run issued %s although the batch was never approved", req)
		}
	}
	if !strings.Contains(out, "ONE expiring typed string for the whole run") {
		t.Errorf("the run banner does not describe the batch model:\n%s", out)
	}
	if !strings.Contains(out, "--confirm-each") {
		t.Errorf("the run banner does not offer the per-mutation alternative:\n%s", out)
	}
}

// TestAnUnapprovedRunDoesNotBlockTheNextAttempt.
//
// The declined approval is on the record — it is an audit fact — but it is not a
// step, so the guard that stops somebody silently re-running measurements already
// made must not mistake it for one.
func TestAnUnapprovedRunDoesNotBlockTheNextAttempt(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	runCLI(t, "--config-dir", configDir, "verify", "run")

	entries, err := verifylive.LoadEntries(filepath.Join(configDir, verifylive.FileName))
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("the declined approval was not recorded at all")
	}
	if n := verifylive.StepCount(entries); n != 0 {
		t.Fatalf("a run that was never approved recorded %d step(s)", n)
	}

	_, _, err = runCLI(t, "--config-dir", configDir, "verify", "run")
	if err != nil && strings.Contains(err.Error(), "already holds") {
		t.Errorf("the second attempt was refused as a restart: %v", err)
	}
}

// TestVerifyRunUnderConfirmEachStillSendsNothingWithoutATerminal — the finer gate
// is kept, and it is kept with the same rail.
func TestVerifyRunUnderConfirmEachStillSendsNothingWithoutATerminal(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	out, _, _ := runCLI(t, "--config-dir", configDir, "verify", "run", "--confirm-each")

	for _, req := range srv.seen() {
		method, path, _ := strings.Cut(req, " ")
		if method == http.MethodPost && path == "/oauth2/token" {
			continue
		}
		if method != http.MethodGet {
			t.Errorf("--confirm-each issued %s with no terminal to confirm it", req)
		}
	}
	if !strings.Contains(out, "--confirm-each:") {
		t.Errorf("the banner does not say which approval model is in force:\n%s", out)
	}
}

// --- status and report -------------------------------------------------------------

func TestVerifyStatusOnAnUnstartedVerification(t *testing.T) {
	configDir := testenv.Isolate(t)

	out, _, err := runCLI(t, "--config-dir", configDir, "verify", "status")
	if err != nil {
		t.Fatalf("verify status: %v", err)
	}
	if !strings.Contains(out, "verify run") {
		t.Errorf("status = %q, want it to point at `tossctl verify run`", out)
	}
}

// TestVerifyStatusNamesWhatIsStillLive — the thing an operator needs after an
// interrupted run.
func TestVerifyStatusNamesWhatIsStillLive(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedVerifyRecord(t, filepath.Join(configDir, verifylive.FileName))

	out, _, err := runCLI(t, "--config-dir", configDir, "verify", "status")
	if err != nil {
		t.Fatalf("verify status: %v", err)
	}
	if !strings.Contains(out, "co-1") {
		t.Errorf("the status does not print the live conditional's id:\n%s", out)
	}
	if !strings.Contains(out, "--resume") {
		t.Errorf("the status does not tell the operator how to finish:\n%s", out)
	}
	if strings.Contains(out, "123-45-678901") {
		t.Error("the status printed the unmasked account number")
	}
}

// TestVerifyReportKeepsReplayDisabledWithoutEvidence is the spec's rule at the
// command surface: "검증되지 않은 상태에서 멱등 재생을 해소 절차로 사용해서는 안
// 된다".
func TestVerifyReportKeepsReplayDisabledWithoutEvidence(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedVerifyRecord(t, filepath.Join(configDir, verifylive.FileName))

	out, _, err := runCLI(t, "--config-dir", configDir, "verify", "report")
	if err != nil {
		t.Fatalf("verify report: %v", err)
	}
	if !strings.Contains(out, "DISABLED") {
		t.Errorf("the report does not report idempotent replay as disabled:\n%s", out)
	}
	if !strings.Contains(out, "no-automatic-entry list") {
		t.Errorf("the report does not produce task 2.6's list:\n%s", out)
	}
}

func TestVerifyReportAsJSON(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedVerifyRecord(t, filepath.Join(configDir, verifylive.FileName))

	out, _, err := runCLI(t, "--config-dir", configDir, "--output", "json", "verify", "report")
	if err != nil {
		t.Fatalf("verify report --output json: %v", err)
	}
	if !strings.Contains(out, `"replay_enabled": false`) && !strings.Contains(out, `"replay_enabled":false`) {
		t.Errorf("the JSON report does not carry replay_enabled=false:\n%s", out)
	}
	if !strings.Contains(out, `"unverified"`) {
		t.Errorf("the JSON report carries no unverified list:\n%s", out)
	}
}

// --- the record's location ---------------------------------------------------------

// TestVerifyRecordDefaultsToTheDataDirectory, next to the soak's: without
// --config-dir it is durable state, not configuration.
func TestVerifyRecordDefaultsToTheDataDirectory(t *testing.T) {
	testenv.Isolate(t)

	got, err := resolveVerifyRecord(&rootOptions{}, "")
	if err != nil {
		t.Fatalf("resolveVerifyRecord: %v", err)
	}
	dataDir, err := journal.DataDir()
	if err != nil {
		t.Fatalf("journal.DataDir: %v", err)
	}
	if want := filepath.Join(dataDir, verifylive.FileName); got != want {
		t.Errorf("record path = %q, want %q", got, want)
	}
}

func TestVerifyRecordFollowsTheConfigDir(t *testing.T) {
	got, err := resolveVerifyRecord(&rootOptions{configDir: "/tmp/profile"}, "")
	if err != nil {
		t.Fatalf("resolveVerifyRecord: %v", err)
	}
	if want := filepath.Join("/tmp/profile", verifylive.FileName); got != want {
		t.Errorf("record path = %q, want %q", got, want)
	}
}

// --- fixtures --------------------------------------------------------------------

// seedVerifyRecord writes a record describing a run that stopped at the
// persistence check with a conditional order still registered.
func seedVerifyRecord(t *testing.T, path string) {
	t.Helper()
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	now := time.Now().UTC()
	entries := []verifylive.Entry{
		{
			StepID: verifylive.StepReadFixtures, Title: "fixtures", Verdict: verifylive.VerdictPass,
			AccountRef: "********8901", StartedAt: now, FinishedAt: now,
			Observations: []verifylive.Observation{{Key: "order.status.observed", Value: "FILLED"}},
		},
		{
			StepID: verifylive.StepConditionalRegister, Title: "register", Verdict: verifylive.VerdictPass,
			AccountRef: "********8901", StartedAt: now, FinishedAt: now, Mutating: true,
			Artifacts: []verifylive.Artifact{{
				Kind: "conditional-order", ID: "co-1", Symbol: "005930",
				CreatedAt: now, Deliberate: true,
			}},
		},
		{
			StepID: verifylive.StepConditionalPersist, Title: "persist",
			Verdict: verifylive.VerdictAwaitingRestart, Reason: "run `tossctl verify run --resume`",
			AccountRef: "********8901", StartedAt: now, FinishedAt: now,
		},
	}
	for _, e := range entries {
		if err := rec.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}
