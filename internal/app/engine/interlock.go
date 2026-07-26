package engine

// interlock.go is the automation gate's startup interlock (harden-execution-base
// task 4.2, engine-safety "자동화 게이트 기동 인터록").
//
// # The rule
//
// The gate is off by default (internal/config/engine.go), and turning it on is
// not sufficient to make it work. With the gate on, the engine starts only if all
// of the following hold (engine-safety "자동화 게이트 기동 인터록", five clauses;
// task 7.5 brought the last three into existence):
//
//	 1. a Guardian is injected                     — nobody authorises orders otherwise
//	 2. every limit is set, positive, finite and
//	    carries a currency                         — a gate unlimited in one dimension is
//	                                                 not a partially authorised gate
//	 3. the trading policy permits selling and
//	    live order actions                         — "can buy, cannot exit" is not a
//	                                                 configuration the engine may run in
//	 4. a capability attestation exists, has not
//	    expired, names this account and covers
//	    every endpoint the engine calls            — see internal/attest
//	 5. the Guardian's EXPOSURE_RAISING limits are
//	    the limits the interlock just audited      — one source, or the audit is fiction
//	 6. an ExecutionGateway is constructed, with
//	    its order reader wired                     — see execgw.Gateway.Wiring
//	 7. the order transport can carry an
//	    idempotency key                            — the replay procedure assumes it did
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
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// ErrAutomationGateRefused is the startup refusal the interlock produces. The
// specific cause is wrapped: errors.Is(err, attest.ErrExpired) and friends still
// work, so a caller can tell an expired attestation from a missing Guardian.
var ErrAutomationGateRefused = errors.New("engine: the automation gate is enabled but its preconditions are not met")

// ErrGuardianRequired means the gate is on and no Guardian was injected.
var ErrGuardianRequired = errors.New(
	"engine: the automation gate is enabled but no Guardian was injected — " +
		"there is nothing to authorise orders, so the engine will not start")

// ErrLimitsRequired means the gate is on with a limit snapshot that is missing
// an item, or carries one that is not a usable number.
//
// It is per-field since task 7.5: the old check accepted a snapshot in which one
// of two ceilings was positive, which is exactly the "부분적으로 무제한인 게이트"
// the spec refuses. The specific field is named in the wrapped message.
var ErrLimitsRequired = errors.New(
	"engine: the automation gate is enabled but engine.automation_gate does not carry a complete " +
		"set of limits — an unlimited gate is not an authorised one")

// ErrTradingPolicyRefused means the gate is on with a trading policy that cannot
// exit a position.
//
// Buying with no way to sell is the one combination that makes the engine
// strictly more dangerous than not running it: it can open exposure it has no
// configured way to close (§0.3).
var ErrTradingPolicyRefused = errors.New(
	"engine: the automation gate is enabled but the trading policy does not permit both selling " +
		"and live order actions — an engine that can enter but not exit must not start")

// ErrGuardianLimitsMismatch means the injected Guardian would stamp
// EXPOSURE_RAISING decisions with limits other than the ones the interlock
// audited.
//
// The audit trail records what the config says. If the Guardian is authorising
// against something else, the audit describes a gate that does not exist. The
// equivalence is asked of EXPOSURE_RAISING only: RISK_REDUCING decisions carry no
// limit snapshot at all.
var ErrGuardianLimitsMismatch = errors.New(
	"engine: the injected Guardian's exposure limits are not the audited configuration limits — " +
		"the gate would be authorising against a source nobody audited")

// ErrGatewayRequired means the gate is on and no usable ExecutionGateway was
// constructed.
var ErrGatewayRequired = errors.New(
	"engine: the automation gate is enabled but the engine profile has no fully wired " +
		"ExecutionGateway — every mutation has to go through one")

// ErrKeylessTransport means the engine's order transport cannot carry the
// broker's idempotency key.
//
// The whole idempotent-replay procedure rests on the key having been sent
// (design D2): recovering a lost response by resending the stored body is only
// safe because the same key cannot create a second order inside its validity
// window. A transport that drops the key turns a replay into a duplicate order.
var ErrKeylessTransport = errors.New(
	"engine: the automation gate is enabled but the order transport cannot carry an idempotency " +
		"key — a place that cannot be replayed is not submittable")

// ExposureLimiter is the capability an injected Guardian has to implement so the
// interlock can prove clause 5.
//
// It is asked at startup rather than inferred from a decision, because inferring
// would mean issuing one: Guardian.Authorize has side effects (it persists a
// decision), and a startup check that writes a decision to inspect it is not a
// check, it is a first order.
//
// A Guardian that cannot state its limits fails the interlock. That is the
// fail-closed direction and it costs nothing today — no production Guardian
// exists before the issuer change — while making "same source" a property the
// next one has to satisfy rather than claim.
type ExposureLimiter interface {
	// ExposureLimits reports the snapshot this Guardian stamps on
	// EXPOSURE_RAISING decisions.
	ExposureLimits() execgw.Limits
}

// idempotencyKeyCarrier is implemented by an order transport whose place path
// sends the broker's clientOrderId.
//
// It is a marker rather than a runtime probe because the property is static: the
// engine's broker is the official client, whose place body carries the field
// (openapi: clientOrderId on OrderCreateRequest), and the web-session mutators
// this package cannot even import do not. Asking the type is how the interlock
// states a fact the build already guarantees.
type idempotencyKeyCarrier interface {
	// CarriesIdempotencyKey reports that a place sent through this transport
	// includes the decision's clientOrderId.
	CarriesIdempotencyKey() bool
}

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
// All five bounds execgw.Limits declares are config fields since task 7.5, so
// this is a straight transcription. It is also the single source clause 5 exists
// to protect: the snapshot the interlock audits, the snapshot the Guardian
// authorises against and the snapshot the gateway re-checks are the same value,
// derived here once.
func gateLimits(gate config.AutomationGate) execgw.Limits {
	return execgw.Limits{
		MaxQuantity:        boundIfPositive(gate.MaxOrderQuantity),
		MaxNotional:        boundIfPositive(gate.MaxOrderNotional),
		MaxTotalExposure:   boundIfPositive(gate.MaxTotalExposure),
		MaxDailyLossAmount: boundIfPositive(gate.MaxDailyLossAmount),
		MaxDailyLossRatio:  boundIfPositive(gate.MaxDailyLossRatio),
		Currency:           strings.ToUpper(strings.TrimSpace(gate.LimitCurrency)),
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

// gateFacts are what construction produced, for the interlock to verify.
//
// They are passed in rather than reached for, so that the checks are a function
// of values a test can supply — the structural clauses (a gateway that was not
// built, a transport that cannot carry a key) cannot be produced by running
// NewContext, because NewContext always builds them correctly.
type gateFacts struct {
	// guardian is the injected risk authority.
	guardian execgw.Guardian
	// trading is the user's configured order policy.
	trading config.Trading
	// gateway is what step 3 of construction produced.
	gateway *execgw.Gateway
	// transport is the order mutator underneath the trading service.
	transport trading.Broker
	// now is the instant the attestation's expiry is judged against.
	now time.Time
}

// runInterlock verifies the gate against what construction produced.
//
// The audit of the settings themselves happens earlier, in NewContext: it must
// precede the account read, which can itself refuse the start (task 7.1).
func runInterlock(status AutomationStatus, gate config.AutomationGate,
	log *audit.Log, logger *obs.Logger, facts gateFacts,
) (AutomationStatus, error) {
	if !gate.Enabled {
		// Off is the default and the whole of the behaviour: no attestation is
		// read, nothing is verified because nothing is permitted.
		logGateDecision(logger, status, nil)
		return status, nil
	}

	if err := verifyGate(&status, facts); err != nil {
		_ = refuseStartup(log, gate, err)
		logGateDecision(logger, status, err)
		return status, err
	}

	status.Verified = true
	if err := log.Record(audit.Entry{
		Action:  audit.ActionGateAccepted,
		Setting: "engine.automation_gate.enabled",
		Old:     "true",
		New:     "true",
		Detail:  acceptanceDetail(&status),
	}); err != nil {
		return status, fmt.Errorf("engine: recording the automation-gate acceptance: %w", err)
	}
	logGateDecision(logger, status, nil)
	return status, nil
}

// acceptanceDetail is what the audit records about a gate that was allowed to
// start.
//
// Every limit appears, not a summary: an operator reading the trail a month later
// is trying to answer "what was this engine authorised to do that day", and a
// detail line that names two of five ceilings cannot answer it.
func acceptanceDetail(status *AutomationStatus) string {
	l := status.Limits
	return fmt.Sprintf(
		"account=%s attestation_expires=%s currency=%s max_order_quantity=%s max_order_notional=%s "+
			"max_total_exposure=%s max_daily_loss_amount=%s max_daily_loss_ratio=%s",
		status.MaskedAccount(),
		status.AttestationExpiresAt.UTC().Format(time.RFC3339),
		l.Currency,
		limitString(l.MaxQuantity.Value),
		limitString(l.MaxNotional.Value),
		limitString(l.MaxTotalExposure.Value),
		limitString(l.MaxDailyLossAmount.Value),
		limitString(l.MaxDailyLossRatio.Value),
	)
}

// logGateDecision emits the one structured line the gate produces per start.
//
// It is one line and not three because the question an operator asks of a log is
// "what mode did this process come up in", and that is a single fact with fields
// hanging off it (obs: "셀 수 있는 이벤트 규약"). A refusal is the same event at
// warn level with a reason, so counting refusals is counting one event with one
// field, not correlating two.
func logGateDecision(logger *obs.Logger, status AutomationStatus, refusal error) {
	l := status.Limits
	fields := []any{
		obs.FieldAccount, status.MaskedAccount(),
		"gate_enabled", status.Enabled,
		"gate_verified", status.Verified,
		"limit_currency", l.Currency,
		"max_order_quantity", limitString(l.MaxQuantity.Value),
		"max_order_notional", limitString(l.MaxNotional.Value),
		"max_total_exposure", limitString(l.MaxTotalExposure.Value),
		"max_daily_loss_amount", limitString(l.MaxDailyLossAmount.Value),
		"max_daily_loss_ratio", limitString(l.MaxDailyLossRatio.Value),
	}
	if refusal != nil {
		logger.Warn(obs.EventOperatingMode, append(fields, obs.FieldReason, refusal.Error())...)
		return
	}
	logger.Event(obs.EventOperatingMode, fields...)
}

// verifyGate is the whole check. It stops at the first failure.
//
// The order is cheapest-and-most-local first: a configuration mistake should not
// cost a file read, and a file read should not cost a comparison against a
// Guardian nobody injected.
func verifyGate(status *AutomationStatus, facts gateFacts) error {
	// 1. The authority.
	if facts.guardian == nil {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, ErrGuardianRequired)
	}

	// 2. Every limit, per field. "Set, positive, finite, with a currency" is
	//    execgw.Limits.Validate, which is the same predicate the gateway applies
	//    to an EXPOSURE_RAISING decision — the interlock refuses at startup what
	//    the gateway would otherwise refuse one order at a time.
	if err := status.Limits.Validate(); err != nil {
		return fmt.Errorf("%w: %w: %w", ErrAutomationGateRefused, ErrLimitsRequired, err)
	}

	// 3. The trading policy. Selling and live order actions are both required:
	//    the failure being excluded is an engine that can open a position and has
	//    no configured way to close it (§0.3).
	if err := checkTradingPolicy(facts.trading); err != nil {
		return fmt.Errorf("%w: %w: %w", ErrAutomationGateRefused, ErrTradingPolicyRefused, err)
	}

	// 4. The attestation itself.
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
	if err := att.Verify(facts.now, accountRef, RequiredEndpoints()); err != nil {
		return fmt.Errorf("%w: %w", ErrAutomationGateRefused, err)
	}

	// 7. Single source: the Guardian authorises against the limits that were
	//    just audited, or the audit trail describes a gate that does not exist.
	if err := checkGuardianLimits(facts.guardian, status.Limits); err != nil {
		return fmt.Errorf("%w: %w: %w", ErrAutomationGateRefused, ErrGuardianLimitsMismatch, err)
	}

	// 8. The gateway, and the transport underneath it.
	if err := checkGateway(facts.gateway); err != nil {
		return fmt.Errorf("%w: %w: %w", ErrAutomationGateRefused, ErrGatewayRequired, err)
	}
	if err := checkKeyCapableTransport(facts.transport); err != nil {
		return fmt.Errorf("%w: %w: %w", ErrAutomationGateRefused, ErrKeylessTransport, err)
	}
	return nil
}

// checkTradingPolicy is clause 3.
func checkTradingPolicy(policy config.Trading) error {
	var missing []string
	if !policy.Sell {
		missing = append(missing, "trading.sell")
	}
	if !policy.AllowLiveOrderActions {
		missing = append(missing, "trading.allow_live_order_actions")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%s is off", strings.Join(missing, " and "))
}

// checkGuardianLimits is clause 4 of the spec's list: the injected Guardian and
// the audited configuration are one source.
func checkGuardianLimits(guardian execgw.Guardian, audited execgw.Limits) error {
	source, ok := guardian.(ExposureLimiter)
	if !ok {
		return fmt.Errorf("the injected Guardian (%T) does not report its exposure limits, so "+
			"equivalence with the audited configuration cannot be established", guardian)
	}
	declared := source.ExposureLimits()
	if !limitsEqual(declared, audited) {
		return fmt.Errorf("Guardian limits %s do not match the configured limits %s",
			limitsString(declared), limitsString(audited))
	}
	return nil
}

// checkGateway is clause 5: a gateway exists and is wired for what it promises.
//
// A non-nil pointer is not the question. Options.Orders is nil-able, and a
// gateway without it confirms a created order on the broker's acknowledgement
// alone — the identifier round trip never runs, and the spec's "생성 응답의
// 식별자는 상세조회 round-trip으로 실재를 확인한다" is unmet at runtime while
// looking satisfied in the code (issues.md 2026-07-26).
func checkGateway(gateway *execgw.Gateway) error {
	if gateway == nil {
		return errors.New("no gateway was constructed")
	}
	wiring := gateway.Wiring()
	var missing []string
	if !wiring.Orders {
		missing = append(missing, "the order reader (identifier round trip)")
	}
	if !wiring.Entry {
		missing = append(missing, "the entry gate")
	}
	if len(missing) > 0 {
		return fmt.Errorf("the gateway is missing %s", strings.Join(missing, " and "))
	}
	return nil
}

// checkKeyCapableTransport is the "키 미지원 transport" refusal.
func checkKeyCapableTransport(transport trading.Broker) error {
	if transport == nil {
		return errors.New("no order transport")
	}
	carrier, ok := transport.(idempotencyKeyCarrier)
	if !ok || !carrier.CarriesIdempotencyKey() {
		return fmt.Errorf("the order transport (%T) does not send the decision's clientOrderId",
			transport)
	}
	return nil
}

// limitsEqual compares two snapshots field by field, Set bits included.
//
// Plain equality on the struct would do the same thing today; it is written out
// so that adding a bound to execgw.Limits fails to compile here rather than
// silently going unchecked.
func limitsEqual(a, b execgw.Limits) bool {
	same := func(x, y execgw.Limit) bool { return x.Set == y.Set && x.Value == y.Value }
	return same(a.MaxQuantity, b.MaxQuantity) &&
		same(a.MaxNotional, b.MaxNotional) &&
		same(a.MaxTotalExposure, b.MaxTotalExposure) &&
		same(a.MaxDailyLossAmount, b.MaxDailyLossAmount) &&
		same(a.MaxDailyLossRatio, b.MaxDailyLossRatio) &&
		strings.EqualFold(strings.TrimSpace(a.Currency), strings.TrimSpace(b.Currency))
}

// limitsString renders a snapshot for a refusal message.
func limitsString(l execgw.Limits) string {
	return fmt.Sprintf("(qty=%s notional=%s exposure=%s daily_loss=%s daily_loss_ratio=%s %s)",
		limitString(l.MaxQuantity.Value), limitString(l.MaxNotional.Value),
		limitString(l.MaxTotalExposure.Value), limitString(l.MaxDailyLossAmount.Value),
		limitString(l.MaxDailyLossRatio.Value), l.Currency)
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
		{audit.ActionLimitChange, "engine.automation_gate.max_total_exposure", limitString(gate.MaxTotalExposure), ""},
		{audit.ActionLimitChange, "engine.automation_gate.max_daily_loss_amount", limitString(gate.MaxDailyLossAmount), ""},
		{audit.ActionLimitChange, "engine.automation_gate.max_daily_loss_ratio", limitString(gate.MaxDailyLossRatio), ""},
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
