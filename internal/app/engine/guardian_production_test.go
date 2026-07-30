package engine_test

import (
	"context"
	"io"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func openGateWithoutProductionGuardian(t *testing.T, dir string, srv *httptest.Server) (*engine.Context, error) {
	t.Helper()
	return openGateWithGuardianOptions(t, dir, srv, false, func(opts *engine.Options) {
		opts.DisableProductionGuardianForTest()
	})
}

func openGateWithGuardianOptions(
	t *testing.T,
	dir string,
	srv *httptest.Server,
	protectionReady bool,
	mutate func(*engine.Options),
) (*engine.Context, error) {
	t.Helper()
	opts := engine.Options{
		ConfigDir: dir,
		Clock:     clock.NewFake(interlockNow),
		AuditFile: filepath.Join(dir, "audit.log"),
		Operator:  "test-operator",
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		},
		Logger: obs.NewLogger(obs.LogOptions{Writer: io.Discard, Clock: clock.NewFake(interlockNow)}),
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "ext4", Magic: journal.MagicExt,
	}))
	if protectionReady {
		opts.SetProtectionReadyForTest()
	}
	if mutate != nil {
		mutate(&opts)
	}
	eng, err := engine.New(opts)
	if eng != nil {
		t.Cleanup(func() { _ = eng.Close() })
	}
	return eng, err
}

func TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver(t *testing.T) {
	dir := isolate(t)
	gate := smallLiveGate()
	gate.MaxOrderNotional = 300
	gate.MaxTotalExposure = 1000
	gate.MaxDailyLossAmount = 50
	gate.LimitCurrency = "usd"
	writeGateConfig(t, dir, gate)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	guardian, ok := eng.Guardian.(*execgw.RiskGuardian)
	if !ok {
		t.Fatalf("Guardian = %T, want *execgw.RiskGuardian", eng.Guardian)
	}
	if guardian.AccountRef() != eng.AccountRef {
		t.Fatalf("Guardian account = %q, context account = %q",
			guardian.AccountRef(), eng.AccountRef)
	}
	want := execgw.Limits{
		MaxQuantity:        execgw.Bound(100),
		MaxNotional:        execgw.Bound(300),
		MaxTotalExposure:   execgw.Bound(1000),
		MaxDailyLossAmount: execgw.Bound(50),
		MaxDailyLossRatio:  execgw.Bound(0.01),
		Currency:           "USD",
	}
	if got := guardian.ExposureLimits(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Guardian limits = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(eng.Automation.Limits, want) {
		t.Fatalf("interlock limits = %+v, want %+v", eng.Automation.Limits, want)
	}
	if _, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()}); err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
}

func TestProductionGuardianConstructionFailureClosesTheEngineJournal(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	var captured *journal.Journal
	eng, err := openGateWithGuardianOptions(t, dir, srv, true, func(opts *engine.Options) {
		opts.FailProductionGuardianForTest(func(j *journal.Journal) { captured = j })
	})
	if err == nil {
		t.Fatal("construction failure must refuse startup")
	}
	if eng != nil {
		t.Fatal("construction failure returned a partial engine")
	}
	if captured == nil {
		t.Fatal("test factory did not receive the engine journal")
	}
	if _, qerr := captured.MaxDecisionTTL(context.Background()); qerr == nil {
		t.Fatal("journal query succeeded after startup failure; handle was not closed")
	}
}
