package engine

// interlock.go is the automation gate's startup interlock (harden-execution-base
// task 4.2, engine-safety "자동화 게이트 기동 인터록").
//
// # The rule
//
// The gate is off by default (internal/config/engine.go), and turning it on is
// not sufficient to make it work. With the gate on, the engine starts only if all
// of the following hold:
//
//	1. a Guardian is injected                      — nobody authorises orders otherwise
//	2. the limit snapshot is non-zero              — "no limit" is not an authorisation
//	3. a capability attestation exists             — see internal/attest
//	4. it has not expired
//	5. it is about the account the credentials resolve to
//	6. it covers every endpoint the engine will call
//
// Any failure is a startup refusal with a message that says what to do about it,
// not a warning. There is no degraded mode: an engine that trades with an
// unverified capability is the thing this interlock exists to prevent.
//
// # Why the account comes from the broker and not the config
//
// (5) compares the attestation against the account the *credentials* resolve to,
// which costs one GET /api/v1/accounts at startup. Comparing against a
// configured string would pass in exactly the case worth catching: an operator
// who copied a config between machines, or swapped credentials, and is now about
// to trade an account nobody verified.
//
// Since task 7.1 that read is no longer the interlock's: engine.NewContext
// performs it on every start, gate or no gate, because the journal and the
// gateway are scoped by the account too (design D8 step 1). The interlock
// receives the resolved reference and compares. The §0.2 promise it used to make
// — "the gate-off path makes no network call" — is therefore narrower now: the
// gate-off path performs no *gate* work, and the one account read it does perform
// is a read the engine profile needs regardless.
//
// # Audit (§0.5)
//
// The gate's state and each limit are compared against the last audited value on
// every startup, and a difference is appended to the audit log with its before and
// after values. The refusal or the acceptance is recorded too: a refused start is
// the only trace an attempt to enable an unverified gate leaves behind.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// ErrAutomationGateRefused is the startup refusal the interlock produces. The
// specific cause is wrapped: errors.Is(err, attest.ErrExpired) and friends still
// work, so a caller can tell an expired attestation from a missing Guardian.
var ErrAutomationGateRefused = errors.New("engine: the automation gate is enabled but its preconditions are not met")

// ErrGuardianRequired means the gate is on and no Guardian was injected.
var ErrGuardianRequired = errors.New(
	"engine: the automation gate is enabled but no Guardian was injected — " +
		"there is nothing to authorise orders, so the engine will not start")

// ErrLimitsRequired means the gate is on with a zero limit snapshot.
var ErrLimitsRequired = errors.New(
	"engine: the automation gate is enabled but engine.automation_gate carries no limit " +
		"(max_order_quantity and max_order_notional are both 0) — an unlimited gate is not an authorised one")

// RequiredEndpoints are the broker calls the engine makes and therefore the ones
// a capability attestation has to cover.
//
// The mutation endpoints are on the list because the attestation is a record of
// what was *proven to work*, and the one-off live verification in
// verify-execution-capability is what proves them. The soak tool itself contains
// no mutation transport (that is its own spec requirement); the endpoints get
// into the attestation from the supervised live check.
func RequiredEndpoints() []string {
	return []string{
		"GET /api/v1/accounts",
		"GET /api/v1/orders",
		"GET /api/v1/orders/{id}",
		"GET /api/v1/buying-power",
		"GET /api/v1/holdings",
		"POST /api/v1/orders",
		"POST /api/v1/orders/{id}/cancel",
	}
}

// AutomationStatus reports what the interlock decided.
//
// It is on Context so that an operator surface (and task 4.3's structured log)
// can state the gate's condition without re-deriving it, and so that a caller can
// tell "off" from "on and verified" — two states that both produce a working
// engine but permit entirely different behaviour.
type AutomationStatus struct {
	// Enabled mirrors the config toggle.
	Enabled bool
	// Verified reports that every interlock precondition held. It is false
	// whenever Enabled is false — an off gate is not a verified one.
	Verified bool
	// AccountRef is the account the credentials resolved to. Populated on every
	// start since task 7.1 — the read is no longer conditional on the gate.
	AccountRef string
	// AttestationFile is where the attestation was looked for.
	AttestationFile string
	// AttestationExpiresAt is the attestation's expiry, zero when there is none.
	AttestationExpiresAt time.Time
	// Limits is the snapshot the Guardian was injected with.
	Limits execgw.Limits
}

// MaskedAccount renders the account for logs and operator output.
func (s AutomationStatus) MaskedAccount() string { return attest.Mask(s.AccountRef) }

// attestationPath resolves where the attestation lives.
func attestationPath(gate config.AutomationGate, paths config.Paths) string {
	if p := strings.TrimSpace(gate.AttestationFile); p != "" {
		return p
	}
	return filepath.Join(paths.ConfigDir, attest.FileName)
}

// gateLimits turns the config's ceilings into the Guardian's limit snapshot.
//
// Only the two per-order ceilings exist in the config today, so only those two
// bits are set. execgw.Limits now requires five, and an EXPOSURE_RAISING
// decision carrying this snapshot would be refused by the gateway — which is the
// fail-closed direction ("부분적으로 무제한인 게이트는 허가된 게이트가 아니다")
// and is why the interlock keeps checking exactly what it checked before until
// the remaining three are added to the config (task 7.5; see issues.md).
func gateLimits(gate config.AutomationGate) execgw.Limits {
	return execgw.Limits{
		MaxQuantity: boundIfPositive(gate.MaxOrderQuantity),
		MaxNotional: boundIfPositive(gate.MaxOrderNotional),
		Currency:    strings.ToUpper(strings.TrimSpace(gate.LimitCurrency)),
	}
}

// boundIfPositive marks a limit as configured only when it carries a usable
// number, so an absent config value stays an absent limit rather than becoming
// "a limit of zero".
func boundIfPositive(v float64) execgw.Limit {
	if v > 0 {
		return execgw.Bound(v)
	}
	return execgw.Limit{}
}

// newAutomationStatus is the status before anything has been verified: the
// config's own account of itself.
func newAutomationStatus(gate config.AutomationGate, paths config.Paths) AutomationStatus {
	return AutomationStatus{
		Enabled:         gate.Enabled,
		AttestationFile: attestationPath(gate, paths),
		Limits:          gateLimits(gate),
	}
}

// refuseStartup records a refusal against the gate and returns the error
// unchanged, so a caller can `return nil, refuseStartup(...)`.
//
// It records only when the gate is on. A gate-off engine that fails to start has
// not attempted anything the gate audit is about, and writing a "gate refused"
// line for it would make the audit trail say something untrue.
func refuseStartup(log *audit.Log, gate config.AutomationGate, err error) error {
	if !gate.Enabled || log == nil {
		return err
	}
	_ = log.Record(audit.Entry{
		Action:  audit.ActionGateRefused,
		Setting: "engine.automation_gate.enabled",
		Old:     "true",
		New:     "true",
		Detail:  err.Error(),
	})
	return err
}

// runInterlock verifies the gate against what construction produced.
//
// The audit of the settings themselves happens earlier, in NewContext: it must
// precede the account read, which can itself refuse the start (task 7.1).
func runInterlock(status AutomationStatus, gate config.AutomationGate,
	log *audit.Log, now time.Time, guardian execgw.Guardian,
) (AutomationStatus, error) {
	if !gate.Enabled {
		// Off is the default and the whole of the behaviour: no attestation is
		// read, nothing is verified because nothing is permitted.
		return status, nil
	}

	if err := verifyGate(&status, guardian, now); err != nil {
		_ = refuseStartup(log, gate, err)
		return status, err
	}

	status.Verified = true
	if err := log.Record(audit.Entry{
		Action:  audit.ActionGateAccepted,
		Setting: "engine.automation_gate.enabled",
		Old:     "true",
		New:     "true",
		Detail: fmt.Sprintf("account=%s attestation expires %s limits qty=%s notional=%s %s",
			status.MaskedAccount(),
			status.AttestationExpiresAt.UTC().Format(time.RFC3339),
			limitString(status.Limits.MaxQuantity.Value),
			limitString(status.Limits.MaxNotional.Value),
			status.Limits.Currency),
	}); err != nil {
		return status, fmt.Errorf("engine: recording the automation-gate acceptance: %w", err)
	}
	return status, nil
}

// verifyGate is the six-step check. It stops at the first failure.
func verifyGate(status *AutomationStatus, guardian execgw.Guardian, now time.Time) error {
	// 1-2. The authority and its limits, before any I/O: these are configuration
	//      mistakes and there is no reason to spend a broker call discovering one.
	if guardian == nil {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, ErrGuardianRequired)
	}
	if !status.Limits.MaxQuantity.Set && !status.Limits.MaxNotional.Set {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, ErrLimitsRequired)
	}

	// 3-4. The attestation itself.
	att, err := attest.Load(status.AttestationFile)
	if err != nil {
		return fmt.Errorf("%w: %w — run the capability verification "+
			"(openspec change verify-execution-capability) and put its attestation at %s",
			ErrAutomationGateRefused, err, status.AttestationFile)
	}
	status.AttestationExpiresAt = att.ExpiresAt

	// 5. The account the credentials actually reach. Resolved during
	//    construction (task 7.1) rather than here, because it is needed whether
	//    or not the gate is on; an empty one means construction let a start
	//    through that it should have refused.
	accountRef := strings.TrimSpace(status.AccountRef)
	if accountRef == "" {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, ErrAccountUnresolved)
	}

	// 6. Expiry, account match and endpoint coverage, in one verification.
	if err := att.Verify(now, accountRef, RequiredEndpoints()); err != nil {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, err)
	}
	return nil
}

// resolveAccountRef reads the account the credentials belong to.
func resolveAccountRef(ctx context.Context, reads OfficialReads) (string, error) {
	if reads == nil {
		return "", errors.New("no official client")
	}
	accounts, err := reads.Accounts(ctx)
	if err != nil {
		return "", err
	}
	// DisplayName carries the official API's accountNo — the human account
	// number (internal/official/reads.go adaptAccounts). ID is accountSeq, an
	// internal key that means nothing to the operator reading an attestation.
	for _, a := range accounts {
		if strings.TrimSpace(a.DisplayName) != "" {
			return strings.TrimSpace(a.DisplayName), nil
		}
	}
	return "", errors.New("the broker returned no account number")
}

// recordGateSettings appends an audit entry for each setting that changed since
// the last recorded value.
func recordGateSettings(log *audit.Log, gate config.AutomationGate, attestationFile string) error {
	changes := []struct {
		action  string
		setting string
		value   string
		detail  string
	}{
		{audit.ActionGateToggle, "engine.automation_gate.enabled", strconv.FormatBool(gate.Enabled),
			"attestation file: " + attestationFile},
		{audit.ActionLimitChange, "engine.automation_gate.max_order_quantity", limitString(gate.MaxOrderQuantity), ""},
		{audit.ActionLimitChange, "engine.automation_gate.max_order_notional", limitString(gate.MaxOrderNotional), ""},
		{audit.ActionLimitChange, "engine.automation_gate.limit_currency",
			strings.ToUpper(strings.TrimSpace(gate.LimitCurrency)), ""},
	}
	for _, c := range changes {
		if _, err := log.RecordChange(c.action, c.setting, c.value, c.detail); err != nil {
			return err
		}
	}
	return nil
}

func limitString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// openAuditLog resolves the audit log's location.
//
// It lives in the data directory rather than the config directory, because the
// audit trail has to outlive the thing it audits: a config directory that is
// replaced, restored from backup or pointed elsewhere by --config-dir must not
// take the record of who changed the limits with it.
func openAuditLog(opts Options) (*audit.Log, error) {
	path := strings.TrimSpace(opts.AuditFile)
	if path == "" {
		dir, err := journal.DataDir()
		if err != nil {
			return nil, fmt.Errorf("engine: resolving the audit log location: %w", err)
		}
		path = filepath.Join(dir, audit.FileName)
	}
	log, err := audit.Open(audit.Options{Path: path, Clock: opts.Clock, Subject: opts.Operator})
	if err != nil {
		return nil, err
	}
	return log, nil
}
