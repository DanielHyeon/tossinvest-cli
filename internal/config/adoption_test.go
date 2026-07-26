package config

// adoption_test.go covers the external-position adoption block (change
// adopt-external-positions task 2.4).
//
// Two properties are being pinned and they matter for different reasons:
//
//	off by default   §0.2. Every config anyone has today has no adoption block,
//	                 and every one of them must load to "the engine does not
//	                 start managing shares its owner bought by hand".
//	refused, not     A stop fraction outside [0.02, 1) is not clamped to the
//	clamped          nearest legal value. Clamping would silently give the
//	                 operator protection they did not ask for; refusing the block
//	                 leaves adoption off, which is the direction §0 permits.

import (
	"strings"
	"testing"
)

func TestAdoptionDefaultsOff(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	a := cfg.Engine.Adoption
	if a.Enabled {
		t.Error("adoption must default to off: a config with no block is every config written before " +
			"this change, and none of their authors asked for their manual holdings to be managed")
	}
	if a.DefaultStopPct != 0 || len(a.ExcludeSymbols) != 0 || a.Rejected != "" {
		t.Errorf("the absent block must load as the zero value, got %+v", a)
	}
}

func TestDefaultFileHasAdoptionOff(t *testing.T) {
	if DefaultFile().Engine.Adoption.Enabled {
		t.Fatal("DefaultFile must not enable adoption")
	}
}

func TestAdoptionBlockParsed(t *testing.T) {
	svc := writeConfig(t, `{
	  "schema_version": 5,
	  "trading": {},
	  "engine": {
	    "adoption": {
	      "enabled": true,
	      "default_stop_pct": 0.05,
	      "exclude_symbols": [" 005930 ", "aapl", "AAPL", ""]
	    }
	  }
	}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	a := cfg.Engine.Adoption
	if !a.Enabled || a.DefaultStopPct != 0.05 || a.Rejected != "" {
		t.Fatalf("adoption = %+v, want the written block accepted", a)
	}
	// Normalised: trimmed, upper-cased, de-duplicated, blanks dropped, sorted.
	if got := strings.Join(a.ExcludeSymbols, ","); got != "005930,AAPL" {
		t.Errorf("exclude_symbols = %q, want \"005930,AAPL\"", got)
	}
	if !a.Excludes("aapl") || !a.Excludes(" AAPL ") {
		t.Error("Excludes must match the way a symbol arrives from the broker, not only as written")
	}
	if a.Excludes("MSFT") || a.Excludes("") {
		t.Error("Excludes must not match a symbol nobody listed")
	}
}

// TestAdoptionRefusesAStopFractionOutsideTheBand is the exit-policy scenario:
// out of range means the setting is refused and adoption stays entirely off.
func TestAdoptionRefusesAStopFractionOutsideTheBand(t *testing.T) {
	for _, pct := range []string{"0.01", "0.019", "0", "1", "1.5", "-0.05"} {
		t.Run(pct, func(t *testing.T) {
			svc := writeConfig(t, `{
			  "schema_version": 5,
			  "trading": {},
			  "engine": {"adoption": {"enabled": true, "default_stop_pct": `+pct+`,
			    "exclude_symbols": ["AAPL"]}}
			}`)

			cfg, err := svc.Load(t.Context())
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			a := cfg.Engine.Adoption
			if a.Enabled {
				t.Errorf("a %s stop fraction left adoption enabled; a refused block must leave the "+
					"feature entirely off rather than running on a number nobody chose", pct)
			}
			if a.DefaultStopPct != 0 || len(a.ExcludeSymbols) != 0 {
				t.Errorf("a refused block must be zeroed, got %+v", a)
			}
			if a.Rejected == "" {
				t.Error("a refused block must say why: an operator who finds adoption off has to be " +
					"able to tell a refusal from an oversight")
			}
			if !strings.Contains(a.Rejected, "0.02") {
				t.Errorf("the refusal must name the bound: %q", a.Rejected)
			}
		})
	}
}

// TestAdoptionAcceptsTheBandItself pins the endpoints, because "0.02 ≤ pct < 1"
// is a closed-open interval and an off-by-one at either end changes which
// configurations run.
func TestAdoptionAcceptsTheBandItself(t *testing.T) {
	for pct, want := range map[string]bool{
		"0.02":     true,
		"0.5":      true,
		"0.999999": true,
		"1":        false,
		"0.0199":   false,
	} {
		svc := writeConfig(t, `{"schema_version":5,"trading":{},
		  "engine":{"adoption":{"enabled":true,"default_stop_pct":`+pct+`}}}`)
		cfg, err := svc.Load(t.Context())
		if err != nil {
			t.Fatalf("Load(%s): %v", pct, err)
		}
		if got := cfg.Engine.Adoption.Enabled; got != want {
			t.Errorf("default_stop_pct %s: enabled = %v, want %v", pct, got, want)
		}
	}
}

// TestAdoptionOffWithNoFractionIsNotARefusal keeps the validation from firing on
// the configuration everybody already has: off, with no fraction at all.
func TestAdoptionOffWithNoFractionIsNotARefusal(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"enabled":false}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.Adoption.Rejected != "" {
		t.Errorf("an explicitly-off block with no fraction was refused: %q", cfg.Engine.Adoption.Rejected)
	}
}

// TestAdoptionDoesNotDisturbTheGate is the §0.2 isolation check: adding a block
// to the engine section must leave the automation gate exactly as it was.
func TestAdoptionDoesNotDisturbTheGate(t *testing.T) {
	svc := writeConfig(t, `{
	  "schema_version": 5,
	  "trading": {},
	  "engine": {
	    "automation_gate": {"enabled": true, "max_order_quantity": 5, "limit_currency": "KRW"},
	    "adoption": {"enabled": true, "default_stop_pct": 0.05}
	  }
	}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	gate := cfg.Engine.AutomationGate
	if !gate.Enabled || gate.MaxOrderQuantity != 5 || gate.LimitCurrency != "KRW" {
		t.Errorf("automation gate = %+v, want the written block untouched", gate)
	}
	if !cfg.Engine.Adoption.Enabled {
		t.Error("the adoption block was dropped when a gate block was present beside it")
	}
}

// TestAdoptionSurvivesAGateOnlyConfig is the mirror: a config with a gate and no
// adoption block must not acquire one.
func TestAdoptionSurvivesAGateOnlyConfig(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"automation_gate":{"enabled":true}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.Adoption.Enabled || cfg.Engine.Adoption.Rejected != "" {
		t.Errorf("adoption = %+v, want the zero value", cfg.Engine.Adoption)
	}
}
