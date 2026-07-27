package main

// soak_test.go covers the CLI surface of `tossctl soak` (verify-execution-capability
// task 1.1).
//
// The survey's logic is tested in internal/soak against an interface. What is
// tested here is what only the wiring can get wrong, and it is tested against a
// fake Open API server so the assertions are about real HTTP:
//
//	the verbs      every request the soak issues is a GET, except the OAuth token
//	               exchange the API itself defines as a POST. This is the runtime
//	               half of the mutation exclusion; the source half is
//	               internal/soak/static_test.go.
//	the wiring     credentials are loaded, cycles are recorded, and the record
//	               ends up somewhere `soak status` can read it back from.
//	the refusal    `soak attest` writes nothing until the record earns it.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// --- a fake official API ----------------------------------------------------

// soakServer is a stand-in for the Open API that records every request it is
// asked to serve.
type soakServer struct {
	*httptest.Server

	mu       sync.Mutex
	requests []string // "METHOD /path"
	// orderListQueries is the query string of every GET /api/v1/orders, kept
	// apart from requests because the page size is a property of the query and
	// not of the path (task 1.11).
	orderListQueries []url.Values
}

func newSoakServer(t *testing.T) *soakServer {
	t.Helper()
	s := &soakServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/api/v1/orders" {
			s.orderListQueries = append(s.orderListQueries, r.URL.Query())
		}
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"soak-token","expires_in":3600,"token_type":"Bearer"}`)
		case r.URL.Path == "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		case r.URL.Path == "/api/v1/buying-power":
			fmt.Fprint(w, `{"result":{"cashBuyingPower":"1000000","currency":"KRW"}}`)
		case r.URL.Path == "/api/v1/holdings":
			fmt.Fprint(w, `{"result":{"items":[{"symbol":"005930","quantity":"10","lastPrice":"70000"}]}}`)
		case r.URL.Path == "/api/v1/orders":
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"ord-1","symbol":"005930","status":"OPEN",`+
				`"quantity":"1","price":"70000"}],"nextCursor":"","hasNext":false}}`)
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/"):
			fmt.Fprint(w, `{"result":{"orderId":"ord-1","symbol":"005930","status":"OPEN","quantity":"1"}}`)
		case r.URL.Path == "/api/v1/prices":
			fmt.Fprint(w, `{"result":[{"symbol":"005930","lastPrice":"70000","currency":"KRW"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func (s *soakServer) seen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

func (s *soakServer) orderLists() []url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]url.Values(nil), s.orderListQueries...)
}

// pointSoakAt wires the soak command at srv for the duration of the test. It
// replaces the factory rather than adding a --base-url flag: a flag that pointed
// the survey at an arbitrary server would let anybody write an attestation the
// live-trading gate accepts, against a server they wrote themselves.
func pointSoakAt(t *testing.T, srv *soakServer, tokenFile string) {
	t.Helper()
	previous := soakReadsFactory
	soakReadsFactory = func(*rootOptions) (soakWiring, error) {
		client := official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			tokenFile,
			official.WithBaseURL(srv.URL),
		)
		return soakWiring{
			reads:       soakReads{api: client},
			tokenExpiry: func() (time.Time, bool) { return soakTokenExpiry(tokenFile) },
			baseURL:     srv.URL,
		}, nil
	}
	t.Cleanup(func() { soakReadsFactory = previous })
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newRootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

// --- the mutation exclusion, at runtime -------------------------------------

// TestSoakIssuesNoMutatingRequest is the spy half of the guarantee in
// execution-verification: "soak 도구는 조회 전용이며 mutation transport를 포함해서는
// 안 된다".
//
// The OAuth exchange is the one POST, and it is the API's own definition of how
// a client obtains a token — it changes nothing about an account. Everything
// else must be a GET, and no request may reach an order-mutating path at all.
func TestSoakIssuesNoMutatingRequest(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "3", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	seen := srv.seen()
	if len(seen) == 0 {
		t.Fatal("the soak issued no request at all")
	}
	sawToken := false
	for _, req := range seen {
		method, path, _ := strings.Cut(req, " ")
		if method == http.MethodPost && path == "/oauth2/token" {
			sawToken = true
			continue
		}
		if method != http.MethodGet {
			t.Errorf("the soak issued %s — it is read-only and must issue nothing but GETs "+
				"(and the OAuth token exchange)", req)
		}
		if strings.Contains(path, "/cancel") || strings.Contains(path, "/modify") {
			t.Errorf("the soak reached %s; that is an order mutation path", req)
		}
	}
	if !sawToken {
		t.Error("no token exchange happened; the credential probe cannot have run")
	}
}

// TestSoakRunSurveysEveryRequiredEndpoint: the wiring has to reach each one, or
// the attestation could never cover it.
func TestSoakRunSurveysEveryRequiredEndpoint(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	cycles, err := soak.LoadCycles(filepath.Join(configDir, soak.FileName))
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("recorded %d cycle(s), want 1", len(cycles))
	}
	c := cycles[0]
	if c.AccountRef != "123-45-678901" {
		t.Errorf("AccountRef = %q, want the broker's account number", c.AccountRef)
	}
	for _, want := range soak.RequiredEndpoints() {
		found := false
		for _, e := range c.Endpoints {
			if e.Endpoint != want {
				continue
			}
			found = true
			if !e.OK {
				t.Errorf("%s: OK = false (%s / %s)", want, e.Class, e.Error)
			}
		}
		if !found {
			t.Errorf("the cycle recorded nothing for %s", want)
		}
	}
	if !c.Completeness.Evaluated || !c.Completeness.OK {
		t.Errorf("completeness = %+v, want an evaluated pass", c.Completeness)
	}
}

// TestSoakOrderWalkAsksForTheLargestPageTheAPIAllows (task 1.11).
//
// The walk used to take the API's default page size of twenty. On an account
// whose CLOSED history runs past a hundred orders that is five requests fired
// back to back where two would do, and it is what made every clean cycle end in a
// 429 on GET /api/v1/orders (measurements.md M8). The cheapest half of the fix is
// to stop asking for the small pages.
//
// Asserted against the real query string rather than the filter value, because
// the filter is only worth anything if the client actually puts it on the wire.
func TestSoakOrderWalkAsksForTheLargestPageTheAPIAllows(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	queries := srv.orderLists()
	if len(queries) == 0 {
		t.Fatal("the soak never asked for the order list")
	}
	for _, q := range queries {
		if got := q.Get("limit"); got != "100" {
			t.Errorf("GET /api/v1/orders?%s: limit = %q, want \"100\" (the openapi maximum; "+
				"an absent limit is the default 20 and is what burst into a 429)", q.Encode(), got)
		}
		// The page size is the only thing that changed. The status group is still
		// required and still walked as two groups (task 1.9).
		if status := q.Get("status"); status != "OPEN" && status != "CLOSED" {
			t.Errorf("GET /api/v1/orders?%s: status = %q, want OPEN or CLOSED", q.Encode(), status)
		}
	}
}

// TestSoakClassifiesTheBrokerSentinels is the wiring half of the 429 backoff.
//
// internal/soak decides what to retry from the Class its Options.Classify hands
// back — it cannot name official.ErrRateLimited itself, because that package is
// kept out of its import graph (internal/soak/static_test.go). So the retry is
// only scoped to a 429 if this function says so, and only this function can be
// asked.
func TestSoakClassifiesTheBrokerSentinels(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want soak.Class
	}{
		{nil, soak.ClassOK},
		{official.ErrRateLimited, soak.ClassRateLimited},
		{fmt.Errorf("reading the order list: %w", official.ErrRateLimited), soak.ClassRateLimited},
		{official.ErrAuth, soak.ClassAuth},
		{official.ErrIPNotAllowed, soak.ClassAuth},
		{official.ErrTransport, soak.ClassTransport},
		{official.ErrServer, soak.ClassServer},
		{errors.New(`official: 400 invalid-request {"field":"status"}`), soak.ClassOther},
	} {
		if got := classifySoakError(tc.err); got != tc.want {
			t.Errorf("classifySoakError(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

// TestSoakRunRefusesWithoutCredentials. Nothing can be measured without them,
// and a silent no-op run would look like a soak that had started.
func TestSoakRunRefusesWithoutCredentials(t *testing.T) {
	configDir := testenv.Isolate(t)

	_, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1")
	if err == nil {
		t.Fatal("soak run started without official credentials")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "credential") {
		t.Errorf("err = %v, want it to name the missing credentials", err)
	}
}

// --- status -----------------------------------------------------------------

func TestSoakStatusOnAnUnstartedSoak(t *testing.T) {
	configDir := testenv.Isolate(t)

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "status")
	if err != nil {
		t.Fatalf("soak status: %v", err)
	}
	if !strings.Contains(out, "soak run") {
		t.Errorf("status = %q, want it to point at `tossctl soak run`", out)
	}
}

// TestSoakStatusReportsProgressAndWhatIsMissing.
func TestSoakStatusReportsProgressAndWhatIsMissing(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "2", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "status")
	if err != nil {
		t.Fatalf("soak status: %v", err)
	}
	for _, want := range []string{"credential streak", "endpoints", "NOT READY"} {
		if !strings.Contains(out, want) {
			t.Errorf("status output has no %q section:\n%s", want, out)
		}
	}
	if strings.Contains(out, "678901") {
		t.Error("the status printed the full account number; it must be masked")
	}
}

// TestSoakStatusAsJSON, for the agent surface and for a cron job that wants to
// alert on the streak.
func TestSoakStatusAsJSON(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	out, _, err := runCLI(t, "--config-dir", configDir, "--output", "json", "soak", "status")
	if err != nil {
		t.Fatalf("soak status --output json: %v", err)
	}
	var report struct {
		Ready   bool     `json:"ready"`
		Reasons []string `json:"reasons"`
		Summary struct {
			StreakDays int `json:"streak_days"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("the JSON status is not valid JSON: %v\n%s", err, out)
	}
	if report.Ready {
		t.Error("a one-cycle soak reported itself ready")
	}
	if len(report.Reasons) == 0 {
		t.Error("a soak that is not ready reported no reason")
	}
}

// --- attestation ------------------------------------------------------------

// TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing.
func TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	_, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest")
	if err == nil {
		t.Fatal("soak attest wrote an attestation for a one-cycle soak")
	}
	if !strings.Contains(err.Error(), "consecutive") {
		t.Errorf("err = %v, want it to explain the missing days", err)
	}
	if _, statErr := os.Stat(filepath.Join(configDir, attest.FileName)); !os.IsNotExist(statErr) {
		t.Fatal("an attestation file exists although the soak was refused")
	}
}

// TestSoakAttestWritesAVerifiableAttestation drives the whole judgement from a
// seeded record: three consecutive days, every endpoint proven, the token seen
// to renew itself.
func TestSoakAttestWritesAVerifiableAttestation(t *testing.T) {
	configDir := testenv.Isolate(t)
	record := filepath.Join(configDir, soak.FileName)
	seedQualifyingRecord(t, record)

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest", "--verified-by", "tester")
	if err != nil {
		t.Fatalf("soak attest: %v", err)
	}

	path := filepath.Join(configDir, attest.FileName)
	a, err := attest.Load(path)
	if err != nil {
		t.Fatalf("the attestation the command wrote cannot be loaded: %v", err)
	}
	if err := a.Verify(time.Now(), "123-45-678901", soak.RequiredEndpoints()); err != nil {
		t.Fatalf("the attestation does not verify against the endpoints it claims: %v", err)
	}
	if a.SoakDays < 3 {
		t.Errorf("SoakDays = %d, want at least 3", a.SoakDays)
	}
	for _, e := range a.Endpoints {
		if !strings.HasPrefix(e, "GET ") {
			t.Errorf("the attestation claims %q; a read-only soak cannot have proven it", e)
		}
	}
	// The operator has to be told the gate will still refuse to start.
	for _, remaining := range soak.LiveOnlyEndpoints() {
		if !strings.Contains(out, remaining) {
			t.Errorf("the output does not mention %q, which the engine still requires:\n%s", remaining, out)
		}
	}
	if strings.Contains(out, "678901") {
		t.Error("the output printed the full account number; it must be masked")
	}
}

// TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn. This is the property
// the change depends on: a read-only soak cannot open the automation gate. The
// mutation endpoints have to come from the supervised live check (task 2.2).
func TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedQualifyingRecord(t, filepath.Join(configDir, soak.FileName))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err != nil {
		t.Fatalf("soak attest: %v", err)
	}
	a, err := attest.Load(filepath.Join(configDir, attest.FileName))
	if err != nil {
		t.Fatalf("attest.Load: %v", err)
	}

	missing := a.MissingEndpoints(engine.RequiredEndpoints())
	if len(missing) == 0 {
		t.Fatal("a soak-only attestation satisfied the engine's full endpoint set; the mutation " +
			"endpoints must still be outstanding")
	}
	for _, e := range missing {
		if strings.HasPrefix(e, "GET ") {
			t.Errorf("%s is a read and the soak should have proven it", e)
		}
	}
}

// TestSoakAndLiveEndpointsCoverTheEngineInterlock keeps the two lists from
// drifting apart: everything the engine requires is either something the soak
// proves or something it explicitly hands to the live check.
func TestSoakAndLiveEndpointsCoverTheEngineInterlock(t *testing.T) {
	covered := map[string]bool{}
	for _, e := range append(soak.RequiredEndpoints(), soak.LiveOnlyEndpoints()...) {
		covered[e] = true
	}
	for _, want := range engine.RequiredEndpoints() {
		if !covered[want] {
			t.Errorf("the engine requires %q and neither soak.RequiredEndpoints nor "+
				"soak.LiveOnlyEndpoints accounts for it", want)
		}
	}
}

// --- command conventions ----------------------------------------------------

func TestSoakCommandsAreRegisteredAndReadOnly(t *testing.T) {
	root := newRootCmd()
	parent := findCommand(root, "soak")
	if parent == nil {
		t.Fatal("soak is not registered on the root command")
	}

	want := map[string]string{"run": "official", "status": "local", "attest": "local"}
	for name, source := range want {
		sub := findCommand(parent, name)
		if sub == nil {
			t.Errorf("soak %s is not registered", name)
			continue
		}
		if got := sub.Annotations["source"]; got != source {
			t.Errorf("soak %s: source = %q, want %q", name, got, source)
		}
		if sub.Annotations["mutating"] == "true" {
			t.Errorf("soak %s declares itself mutating; the soak places no order", name)
		}
		if strings.TrimSpace(sub.Short) == "" {
			t.Errorf("soak %s has no Short description", name)
		}
	}
}

// TestSoakRunOffersTheOperatorTheKnobsTheSoakNeeds — and, being the command that
// feeds an attestation, no way to point it at a server of one's own choosing.
func TestSoakRunFlags(t *testing.T) {
	cmd := findCommand(findCommand(newRootCmd(), "soak"), "run")
	if cmd == nil {
		t.Fatal("soak run is not registered")
	}
	for _, name := range []string{"interval", "cycles", "symbols", "record"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("soak run has no --%s", name)
		}
	}
	for _, banned := range []string{"base-url", "api-base-url", "endpoint", "insecure"} {
		if cmd.Flags().Lookup(banned) != nil {
			t.Errorf("soak run offers --%s; the survey that feeds an attestation must not be "+
				"redirectable at an arbitrary server", banned)
		}
	}
}

// TestSoakRecordDefaultsToTheDataDirectory. Without --config-dir the record is
// durable state, not configuration, and belongs where the journal lives.
func TestSoakRecordDefaultsToTheDataDirectory(t *testing.T) {
	testenv.Isolate(t)

	got, err := resolveSoakRecord(&rootOptions{}, "")
	if err != nil {
		t.Fatalf("resolveSoakRecord: %v", err)
	}
	dataDir, err := journal.DataDir()
	if err != nil {
		t.Fatalf("journal.DataDir: %v", err)
	}
	if want := filepath.Join(dataDir, soak.FileName); got != want {
		t.Errorf("record path = %q, want %q", got, want)
	}
}

// --- fixtures ---------------------------------------------------------------

// seedQualifyingRecord writes a record describing a soak that has run to
// completion: three consecutive days, two cycles a day, every endpoint proven,
// and a token expiry that keeps moving forward.
func seedQualifyingRecord(t *testing.T, path string) {
	t.Helper()
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	start := time.Now().UTC().Add(-72 * time.Hour).Truncate(time.Hour)
	for day := 0; day < 3; day++ {
		for half := 0; half < 2; half++ {
			at := start.AddDate(0, 0, day).Add(time.Duration(half) * 8 * time.Hour)
			c := soak.Cycle{
				FormatVersion: soak.RecordFormatVersion,
				Kind:          "cycle",
				StartedAt:     at,
				FinishedAt:    at.Add(2 * time.Second),
				AccountRef:    "123-45-678901",
				Credential: soak.Credential{
					OK:             true,
					Class:          soak.ClassOK,
					Observed:       true,
					TokenExpiresAt: at.Add(time.Hour),
					Refreshed:      !(day == 0 && half == 0),
				},
				Completeness: soak.Completeness{Evaluated: true, OK: true, OrderPages: 1, OrderIDs: 1},
			}
			for _, e := range soak.RequiredEndpoints() {
				c.Endpoints = append(c.Endpoints, soak.EndpointResult{
					Endpoint: e, OK: true, Class: soak.ClassOK, Requests: 1, LatencyMS: 120,
				})
			}
			if err := rec.Append(c); err != nil {
				t.Fatalf("Append: %v", err)
			}
		}
	}
}
