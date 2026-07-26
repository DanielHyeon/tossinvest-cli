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

// Engine holds the automated-trading engine's settings.
//
// The CLI ignores this block entirely: it configures the unattended engine
// profile (internal/app/engine), not `tossctl`.
type Engine struct {
	AutomationGate AutomationGate `json:"automation_gate"`
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

	// LimitCurrency is the currency the money ceilings are expressed in
	// ("KRW"/"USD").
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
	AutomationGate *rawAutomationGate `json:"automation_gate"`
}

// mergeEngine copies a present engine block onto the defaults.
//
// An absent block, an absent automation_gate, or a nil receiver all leave the
// defaults alone — and the defaults are off.
func mergeEngine(cfg *Engine, raw *rawEngine) {
	if raw == nil || raw.AutomationGate == nil {
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
