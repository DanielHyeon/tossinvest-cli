package config

// engine.go adds the automated-trading engine's operating settings to the config
// file (harden-execution-base task 4.2, engine-safety "자동화 게이트 기동 인터록").
//
// # Why a new file in an upstream package
//
// The block has to live in config.File, because config.File is what both
// profiles load and what `tossctl config show` reports. Everything that can be
// additive is additive: the types, the parsing and the defaulting are all here,
// and service.go gains only the field declarations and the four lines that copy a
// present block across. No existing field changes meaning, and no existing
// parsing branch is touched.
//
// # The one rule this block must satisfy
//
// OFF must be indistinguishable from upstream (WORKFLOW §0.2). The zero value of
// every field is therefore the safe one: the gate is off, the limits are zero,
// and a config file written by any earlier schema version — which has no engine
// block at all — loads to exactly that. There is no "unset means on" anywhere in
// this file, and no default that opens anything.
//
// # Why the limits live in the config
//
// engine-safety requires the Guardian to be injected with non-zero limits before
// the gate may be on, and §0.5 requires limit changes to be auditable. A limit
// that lives in a config file is a limit an operator can see, diff, and have
// recorded when it changes; a limit compiled in or passed on a command line is
// neither. The engine reads these into the Guardian's limit snapshot at startup
// and refuses to start if they are zero while the gate is on.

import (
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// Engine holds the automated-trading engine's settings.
//
// The CLI ignores this block entirely: it configures the unattended engine
// profile (internal/app/engine), not `tossctl`.
type Engine struct {
	// Autostart records the operator's separate approval to attempt an engine
	// start whenever the console process starts. It does not grant order
	// capability: the automation gate and the engine's startup interlock remain
	// the only judges of that. Zero/missing is deliberately off.
	Autostart      bool           `json:"autostart"`
	AutomationGate AutomationGate `json:"automation_gate"`
	// Adoption configures whether externally acquired holdings are taken into
	// exit management. Zero value = off, which is what every pre-adoption config
	// produces.
	Adoption Adoption `json:"adoption"`
	// ExitPolicy selects the immutable common profile for newly managed
	// positions. Empty preserves the pre-change RATCHET behavior.
	ExitPolicy ExitPolicy `json:"exit_policy"`
	// Notifications configures where critical alerts are sent. Zero value = off,
	// which wires no transport and is exactly what every pre-a074 config
	// produces. See notifications.go.
	Notifications Notifications `json:"notifications"`
}

type ExitPolicy struct {
	CommonPolicy string `json:"common_policy,omitempty"`
	Rejected     string `json:"rejected,omitempty"`
}

func (p ExitPolicy) validate() string {
	id := strings.TrimSpace(p.CommonPolicy)
	if id == "" {
		return ""
	}
	if _, ok := exitpolicy.CommonPolicyByID(id); !ok {
		return "unknown common exit policy " + id
	}
	return ""
}

// Adoption is the external-position adoption feature's settings (change
// adopt-external-positions, design A3).
//
// # Off by default, and off is the landed behaviour
//
// §0.2 again: the zero value is the safe one. With `enabled` false the engine
// behaves exactly as it did before this change — including raising the unmanaged
// holding alert, which design A4 keeps regardless of this toggle. What the
// toggle turns on is the engine *acting* on that discovery.
//
// Turning it on is a §0.7 action: a human decision, recorded in the audit trail
// (recordGateSettings). Nothing in TossOS flips it.
//
// # Why the fraction has a floor and not just a ceiling
//
// `default_stop_pct` is the whole of the synthetic protection: the stop is
// `observed × (1 − pct)` and there is nothing else. Below
// exitpolicy.MinStopPct the band is narrower than the noise between two
// observations and than the round-trip cost of the exit itself, which makes it
// a device that liquidates on the first tick rather than a stop. A value outside
// [0.02, 1) is therefore refused, and a refused block leaves adoption entirely
// off rather than running on a number nobody chose.
type Adoption struct {
	// Enabled turns adoption on. Default false, and false is the only value any
	// pre-adoption config can produce.
	Enabled bool `json:"enabled"`

	// DefaultStopPct is the synthetic stop's distance below the adoption
	// observation, as a fraction. Must be in [0.02, 1) whenever the block is
	// meaningful; see Rejected.
	DefaultStopPct float64 `json:"default_stop_pct,omitempty"`

	// ExcludeSymbols are never adopted. It is the fine-grained control *inside*
	// enabled: an operator who wants the engine to manage everything except one
	// long-term holding names it here rather than turning the feature off.
	ExcludeSymbols []string `json:"exclude_symbols,omitempty"`

	// IncludeSymbols are adoption candidates even with Enabled false — the
	// per-symbol designation (change console-adoption-controls). Exclusion wins
	// when a symbol is on both lists; the engine owns that judgement. The list
	// is market-agnostic, like ExcludeSymbols. Empty is exactly the pre-change
	// behaviour (§0.2).
	IncludeSymbols []string `json:"include_symbols,omitempty"`

	// Rejected explains why the block was refused, and is empty when it was
	// accepted. A refused block is zeroed, so `Enabled` is already false — this
	// field exists so the engine can say *why* rather than silently ignoring what
	// an operator wrote.
	Rejected string `json:"rejected,omitempty"`
}

// Excludes reports whether a symbol is on the exclusion list.
func (a Adoption) Excludes(symbol string) bool {
	return onSymbolList(a.ExcludeSymbols, symbol)
}

// Included reports whether a symbol is on the per-symbol designation list.
func (a Adoption) Included(symbol string) bool {
	return onSymbolList(a.IncludeSymbols, symbol)
}

func onSymbolList(list []string, symbol string) bool {
	want := strings.ToUpper(strings.TrimSpace(symbol))
	if want == "" {
		return false
	}
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// validate reports why the block cannot be used, or "" when it can.
//
// A block that is off *and* carries no fraction is not validated at all: that is
// the absent block, and every config written before this change is one.
func (a Adoption) validate() string {
	// "Meaningful" widened by console-adoption-controls: an include list alone
	// promises those symbols a synthetic stop, so it needs the fraction exactly
	// as enabled does. Off with neither is the absent block everybody has.
	if !a.Enabled && len(a.IncludeSymbols) == 0 && a.DefaultStopPct == 0 {
		return ""
	}
	if err := exitpolicy.ValidateStopPct(a.DefaultStopPct); err != nil {
		return err.Error()
	}
	return ""
}

// rawAdoption is the parse shape. Enabled is a pointer for the same reason the
// gate's is: an explicit false and an absent key are distinguishable in tests,
// and both produce false.
type rawAdoption struct {
	Enabled        *bool    `json:"enabled"`
	DefaultStopPct float64  `json:"default_stop_pct"`
	ExcludeSymbols []string `json:"exclude_symbols"`
	IncludeSymbols []string `json:"include_symbols"`
}

// AutomationGate is the master switch for unattended order placement.
//
// Turning it on is a §0.7 action — a human decision, recorded in an audit log.
// Nothing in TossOS flips it automatically, and the engine refuses to start with
// it on until a capability attestation and non-zero limits are both in place.
type AutomationGate struct {
	// Enabled turns unattended order placement on. Default false, and false is
	// the only value any pre-v5 config can produce.
	Enabled bool `json:"enabled"`

	// AttestationFile overrides where the capability attestation is read from.
	// Empty uses the per-user default (<config dir>/capability-attestation.json).
	AttestationFile string `json:"attestation_file,omitempty"`

	// MaxOrderQuantity is the largest quantity a single order may carry. Zero
	// means "no quantity limit set", which with the gate on is a startup refusal
	// rather than permission.
	MaxOrderQuantity float64 `json:"max_order_quantity,omitempty"`

	// MaxOrderNotional is the largest quantity×price a single order may carry,
	// expressed in LimitCurrency.
	MaxOrderNotional float64 `json:"max_order_notional,omitempty"`

	// MaxTotalExposure is the account-wide ceiling on open exposure, expressed
	// in LimitCurrency. Zero means "not set", which with the gate on is a
	// startup refusal.
	MaxTotalExposure float64 `json:"max_total_exposure,omitempty"`

	// MaxDailyLossAmount is the absolute daily realised-loss ceiling, in
	// LimitCurrency.
	MaxDailyLossAmount float64 `json:"max_daily_loss_amount,omitempty"`

	// MaxDailyLossRatio is the daily loss ceiling as a fraction of capital, in
	// (0, 1]. 0.02 is two percent.
	MaxDailyLossRatio float64 `json:"max_daily_loss_ratio,omitempty"`

	// LimitCurrency is the account base currency in which all account-wide money
	// ceilings are expressed ("KRW"/"USD"). Market quote cash remains in its
	// quote currency and exposure-raising KR/US paths require frozen official
	// quote-to-base authority.
	LimitCurrency string `json:"limit_currency,omitempty"`
}

// LimitsSet reports whether the gate carries any usable limit at all.
//
// It is the weak question — "is anything set" — and it is deliberately not what
// the startup interlock asks. The interlock requires *every* limit to be present,
// positive and finite (execgw.Limits.Validate; engine-safety: "하나라도 누락·0·
// NaN·Inf이면 거부"), because a gate that is unlimited in one dimension is not a
// partially authorised gate, it is an unauthorised one. This helper stays as it
// was for the config surfaces that only need to say whether the block was filled
// in at all.
func (g AutomationGate) LimitsSet() bool {
	return g.MaxOrderQuantity > 0 || g.MaxOrderNotional > 0 ||
		g.MaxTotalExposure > 0 || g.MaxDailyLossAmount > 0 || g.MaxDailyLossRatio > 0
}

// rawAutomationGate is the parse shape. Enabled is a pointer only so that an
// explicit `"enabled": false` and an absent key are distinguishable in tests;
// both produce false, because there is no other safe answer.
type rawAutomationGate struct {
	Enabled            *bool   `json:"enabled"`
	AttestationFile    string  `json:"attestation_file"`
	MaxOrderQuantity   float64 `json:"max_order_quantity"`
	MaxOrderNotional   float64 `json:"max_order_notional"`
	MaxTotalExposure   float64 `json:"max_total_exposure"`
	MaxDailyLossAmount float64 `json:"max_daily_loss_amount"`
	MaxDailyLossRatio  float64 `json:"max_daily_loss_ratio"`
	LimitCurrency      string  `json:"limit_currency"`
}

type rawEngine struct {
	Autostart      *bool              `json:"autostart"`
	AutomationGate *rawAutomationGate `json:"automation_gate"`
	Adoption       *rawAdoption       `json:"adoption"`
	ExitPolicy     *rawExitPolicy     `json:"exit_policy"`
	Notifications  *rawNotifications  `json:"notifications"`
}

type rawExitPolicy struct {
	CommonPolicy string `json:"common_policy"`
}

// mergeAdoption copies a present adoption block onto the defaults, or refuses it.
//
// A refused block is *zeroed*, not partially kept: exit-policy's scenario is
// "설정이 거부되고 편입은 전면 비활성으로 남는다", and a block that kept `enabled`
// while dropping the fraction would be adoption running on a stop nobody chose.
func mergeAdoption(cfg *Engine, raw *rawAdoption) {
	if raw == nil {
		return
	}
	next := Adoption{DefaultStopPct: raw.DefaultStopPct}
	if raw.Enabled != nil {
		next.Enabled = *raw.Enabled
	}
	next.ExcludeSymbols = normaliseSymbols(raw.ExcludeSymbols)
	next.IncludeSymbols = normaliseSymbols(raw.IncludeSymbols)

	if why := next.validate(); why != "" {
		cfg.Adoption = Adoption{Rejected: why}
		return
	}
	cfg.Adoption = next
}

// normaliseSymbols upper-cases, trims, drops blanks and de-duplicates, so
// "005930" and " 005930 " are one exclusion rather than two and the list can be
// compared as text in the audit trail.
func normaliseSymbols(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		symbol := strings.ToUpper(strings.TrimSpace(s))
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	sort.Strings(out)
	return out
}

// mergeEngine copies a present engine block onto the defaults.
//
// An absent block, an absent automation_gate, or a nil receiver all leave the
// defaults alone — and the defaults are off.
func mergeEngine(cfg *Engine, raw *rawEngine) {
	if raw == nil {
		return
	}
	if raw.Autostart != nil {
		cfg.Autostart = *raw.Autostart
	}
	mergeAdoption(cfg, raw.Adoption)
	if raw.ExitPolicy != nil {
		next := ExitPolicy{CommonPolicy: strings.TrimSpace(raw.ExitPolicy.CommonPolicy)}
		next.Rejected = next.validate()
		cfg.ExitPolicy = next
	}
	// Before the automation-gate early return below, not after: an engine with the
	// gate off still runs a console and still has critical alerts to deliver, and
	// a notification block silently ignored because the file has no
	// `automation_gate` key would be the worst kind of quiet.
	mergeNotifications(cfg, raw.Notifications)
	if raw.AutomationGate == nil {
		return
	}
	gate := raw.AutomationGate
	if gate.Enabled != nil {
		cfg.AutomationGate.Enabled = *gate.Enabled
	}
	cfg.AutomationGate.AttestationFile = gate.AttestationFile
	cfg.AutomationGate.MaxOrderQuantity = gate.MaxOrderQuantity
	cfg.AutomationGate.MaxOrderNotional = gate.MaxOrderNotional
	cfg.AutomationGate.MaxTotalExposure = gate.MaxTotalExposure
	cfg.AutomationGate.MaxDailyLossAmount = gate.MaxDailyLossAmount
	cfg.AutomationGate.MaxDailyLossRatio = gate.MaxDailyLossRatio
	cfg.AutomationGate.LimitCurrency = gate.LimitCurrency
}
