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
// That read happens only when the gate is on. With the gate off — every machine
// today — engine.New makes no network call at all and behaves precisely as it did
// before this file existed (§0.2).
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
	// AccountRef is the account the credentials resolved to. Only populated when
	// the gate is on, because that is the only case where it is read.
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

// runInterlock records the audit trail and then verifies the gate.
//
// The audit comes first on purpose: an operator's change is worth recording
// whether or not the engine then agrees to start on it, and the refusal record
// that follows is what ties the two together.
func runInterlock(ctx context.Context, gate config.AutomationGate, paths config.Paths,
	reads OfficialReads, log *audit.Log, now time.Time, guardian execgw.Guardian,
) (AutomationStatus, error) {
	status := AutomationStatus{
		Enabled:         gate.Enabled,
		AttestationFile: attestationPath(gate, paths),
		Limits:          gateLimits(gate),
	}

	if err := recordGateSettings(log, gate, status.AttestationFile); err != nil {
		// A settings change we cannot record is a settings change nobody can
		// audit, and §0.5 makes the audit trail mandatory rather than
		// best-effort. Refusing here is the conservative direction: the engine
		// does not start, and no order is placed off the record.
		return status, fmt.Errorf("engine: recording the automation-gate audit trail: %w", err)
	}

	if !gate.Enabled {
		// Off is the default and the whole of the behaviour: no attestation is
		// read, no broker call is made, nothing is verified because nothing is
		// permitted.
		return status, nil
	}

	if err := verifyGate(ctx, &status, reads, guardian, now); err != nil {
		_ = log.Record(audit.Entry{
			Action:  audit.ActionGateRefused,
			Setting: "engine.automation_gate.enabled",
			Old:     "true",
			New:     "true",
			Detail:  err.Error(),
		})
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
func verifyGate(ctx context.Context, status *AutomationStatus, reads OfficialReads,
	guardian execgw.Guardian, now time.Time,
) error {
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

	// 5. The account the credentials actually reach.
	accountRef, err := resolveAccountRef(ctx, reads)
	if err != nil {
		return fmt.Errorf("%w: the account could not be resolved, so the attestation cannot be "+
			"matched against it: %w", ErrAutomationGateRefused, err)
	}
	status.AccountRef = accountRef

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
