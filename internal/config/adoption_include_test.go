package config

// adoption_include_test.go covers the per-symbol designation list (change
// console-adoption-controls task 1.1).
//
//	opt-in exists    include_symbols makes a symbol an adoption candidate with
//	                 enabled false — "체크한 종목만 편입" finally has a spelling.
//	meaningful       a block with only an include list still needs the stop
//	                 fraction: those symbols get the same synthetic stop, so the
//	                 validation that protects enabled=true protects them too.
//	zero value       an empty list is byte-for-byte today's behaviour (§0.2).

import (
	"strings"
	"testing"
)

func TestIncludeListParsedAndNormalised(t *testing.T) {
	svc := writeConfig(t, `{
	  "schema_version": 5,
	  "trading": {},
	  "engine": {"adoption": {"default_stop_pct": 0.05,
	    "include_symbols": [" 005930 ", "aapl", "AAPL", ""]}}
	}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	a := cfg.Engine.Adoption
	if a.Enabled {
		t.Error("an include list must not turn the global switch on")
	}
	if a.Rejected != "" {
		t.Fatalf("an include-only block with a valid fraction was refused: %q", a.Rejected)
	}
	if got := strings.Join(a.IncludeSymbols, ","); got != "005930,AAPL" {
		t.Errorf("include_symbols = %q, want \"005930,AAPL\" (trimmed, upper-cased, de-duplicated)", got)
	}
	if !a.Included("aapl") || !a.Included(" 005930 ") {
		t.Error("Included must match the way a symbol arrives from the broker, not only as written")
	}
	if a.Included("MSFT") || a.Included("") {
		t.Error("Included must not match a symbol nobody designated")
	}
}

// TestIncludeOnlyBlockNeedsTheStopFraction: the designation is a promise of a
// synthetic stop, and a promise with no fraction is refused — never silently
// clamped, never written off to run on a number nobody chose.
func TestIncludeOnlyBlockNeedsTheStopFraction(t *testing.T) {
	for _, pct := range []string{"0", "0.01", "1"} {
		t.Run(pct, func(t *testing.T) {
			svc := writeConfig(t, `{"schema_version":5,"trading":{},
			  "engine":{"adoption":{"include_symbols":["005930"],"default_stop_pct":`+pct+`}}}`)

			cfg, err := svc.Load(t.Context())
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			a := cfg.Engine.Adoption
			if a.Rejected == "" {
				t.Fatal("an include list with an out-of-band fraction must refuse the block: those " +
					"symbols would be adopted onto a stop nobody chose")
			}
			if len(a.IncludeSymbols) != 0 || a.Enabled {
				t.Errorf("a refused block must be zeroed, got %+v", a)
			}
		})
	}
}

// TestEmptyIncludeIsTheZeroValue keeps §0.2: every config written before this
// change has no include list, and every one of them must behave as before.
func TestEmptyIncludeIsTheZeroValue(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"enabled":false}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	a := cfg.Engine.Adoption
	if len(a.IncludeSymbols) != 0 || a.Rejected != "" || a.Included("005930") {
		t.Errorf("adoption = %+v, want the zero value with nothing included", a)
	}
}

// TestIncludeAndExcludeTogetherAreAccepted: the file may carry both — precedence
// (exclude wins) is the engine's judgement, not a parse error. Refusing the file
// would zero both lists and lose the exclusion protecting a holding.
func TestIncludeAndExcludeTogetherAreAccepted(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"default_stop_pct":0.05,
	    "include_symbols":["005930"],"exclude_symbols":["005930"]}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	a := cfg.Engine.Adoption
	if a.Rejected != "" {
		t.Fatalf("a both-lists block was refused: %q", a.Rejected)
	}
	if !a.Included("005930") || !a.Excludes("005930") {
		t.Error("both lists must survive the load; precedence belongs to the engine")
	}
}
