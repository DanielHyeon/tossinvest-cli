package app

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestA166AppBudgetProviderFollowsRouting(t *testing.T) {
	for _, tc := range []struct {
		name, prefer string
		enabled      bool
		wantCall     bool
	}{{"wts", "wts", true, false}, {"disabled", "auto", false, false}, {"official", "auto", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			cfg := config.DefaultFile()
			cfg.OpenAPI = config.OpenAPI{Enabled: tc.enabled, Prefer: tc.prefer, Fallback: true}
			writeConfigFile(t, dir, cfg)
			writeCredentials(t, dir)
			calls := 0
			sentinel := errors.New("budget unavailable")
			_, e := New(Options{OutputFormat: "table", ConfigDir: dir, BudgetProvider: func() (*accountbudgetruntime.Factory, error) { calls++; return nil, sentinel }})
			if tc.wantCall {
				if calls != 1 || !errors.Is(e, sentinel) {
					t.Fatalf("wanted fail closed calls=%d err=%v", calls, e)
				}
			} else if calls != 0 || e != nil {
				t.Fatalf("unexpected budget calls=%d err=%v", calls, e)
			}
		})
	}
}
func TestA166AppNilFactoryIsError(t *testing.T) {
	dir := isolate(t)
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: "auto"}
	writeConfigFile(t, dir, cfg)
	writeCredentials(t, dir)
	c, e := New(Options{OutputFormat: "table", ConfigDir: dir, BudgetProvider: func() (*accountbudgetruntime.Factory, error) { return nil, nil }})
	if c != nil || e == nil {
		t.Fatal("nil factory accepted")
	}
}
