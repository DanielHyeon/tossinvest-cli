//go:build tossos_testseams

package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

func TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver(t *testing.T) {
	dir := testenv.Isolate(t)
	now := time.Date(2026, 7, 30, 6, 52, 33, 0, time.UTC)
	cfg := config.DefaultFile()
	cfg.Trading = config.Trading{
		Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
	}
	cfg.Engine.AutomationGate = config.AutomationGate{
		Enabled:            true,
		MaxOrderQuantity:   100,
		MaxOrderNotional:   300,
		MaxTotalExposure:   1000,
		MaxDailyLossAmount: 50,
		MaxDailyLossRatio:  0.01,
		LimitCurrency:      "USD",
	}
	writeConfigFile(t, dir, cfg)
	writeCredentials(t, dir)
	if err := attest.Save(filepath.Join(dir, attest.FileName), attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    "123-45",
		IssuedAt:      now.Add(-time.Hour),
		ExpiresAt:     now.Add(24 * time.Hour),
		SoakDays:      3,
		Endpoints:     engine.RequiredEndpoints(),
		VerifiedBy:    "test operator",
	}); err != nil {
		t.Fatalf("save attestation: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = io.WriteString(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			_, _ = io.WriteString(w, `{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	var constructions int
	clk := clock.NewFake(now)
	ectx, err := assembleEngine(context.Background(), &rootOptions{configDir: dir}, clk,
		obs.NewLogger(obs.LogOptions{Writer: io.Discard, Clock: clk}),
		func(opts *engine.Options) {
			engine.ConfigureCLIRegressionForTest(opts,
				journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
				func() { constructions++ },
			)
		},
		official.WithBaseURL(srv.URL),
		official.WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("assembleEngine: %v", err)
	}
	if constructions != 1 {
		t.Fatalf("production Guardian constructions = %d, want exactly 1", constructions)
	}
	guardian, ok := ectx.Guardian.(*execgw.RiskGuardian)
	if !ok {
		t.Fatalf("Guardian = %T, want *execgw.RiskGuardian", ectx.Guardian)
	}
	wantLimits := execgw.Limits{
		MaxQuantity:        execgw.Bound(100),
		MaxNotional:        execgw.Bound(300),
		MaxTotalExposure:   execgw.Bound(1000),
		MaxDailyLossAmount: execgw.Bound(50),
		MaxDailyLossRatio:  execgw.Bound(0.01),
		Currency:           "USD",
	}
	if got := guardian.ExposureLimits(); !reflect.DeepEqual(got, wantLimits) {
		t.Fatalf("Guardian limits = %+v, want %+v", got, wantLimits)
	}
	if !reflect.DeepEqual(ectx.Automation.Limits, wantLimits) {
		t.Fatalf("interlock limits = %+v, want %+v", ectx.Automation.Limits, wantLimits)
	}

	issued, err := guardian.IssueReduction(context.Background(), execgw.ReductionIssuance{
		Intent: risk.Intent{
			AccountRef: ectx.AccountRef,
			Market:     costs.MarketUS,
			Symbol:     "AAPL",
			Side:       risk.SideSell,
			Quantity:   "1",
		},
		Account: risk.AccountState{HeldQuantity: "2"},
		Reason:  "CLI production wiring regression",
	})
	if err != nil {
		t.Fatalf("IssueReduction: %v", err)
	}
	sharedJournal := ectx.Journal
	decision, err := sharedJournal.LookupDecision(context.Background(), issued.Decision.ID)
	if err != nil {
		t.Fatalf("decision is not visible through Context.Journal: %v", err)
	}
	if decision.SafetyClass != journal.SafetyClassRiskReducing {
		t.Fatalf("decision safety class = %s, want %s",
			decision.SafetyClass, journal.SafetyClassRiskReducing)
	}
	if _, err := ectx.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()}); err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	if err := ectx.Close(); err != nil {
		t.Fatalf("Context.Close: %v", err)
	}
	if _, err := sharedJournal.MaxDecisionTTL(context.Background()); err == nil {
		t.Fatal("engine-owned journal remained usable after Context.Close")
	}
}
