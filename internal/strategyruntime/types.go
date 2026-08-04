package strategyruntime

import "time"

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

func validMarket(market Market) bool { return market == MarketKR || market == MarketUS }

type EntryState string

const (
	EntryOff EntryState = "OFF"
	EntryOn  EntryState = "ON"
)

type RuntimeState string

const (
	RuntimeUnobserved RuntimeState = "UNOBSERVED"
	RuntimeObserved   RuntimeState = "OBSERVED"
)

type RefusalCode string

const (
	RefusalNone                 RefusalCode = ""
	RefusalInvalid              RefusalCode = "INVALID"
	RefusalDisabled             RefusalCode = "DISABLED"
	RefusalWaitMarket           RefusalCode = "WAIT_MARKET"
	RefusalBudgetDeferred       RefusalCode = "BUDGET_DEFERRED"
	RefusalWorkerAbnormal       RefusalCode = "WORKER_ABNORMAL"
	RefusalWorkerFailure        RefusalCode = "WORKER_FAILURE"
	RefusalAuthorityDrift       RefusalCode = "AUTHORITY_DRIFT"
	RefusalScopeMismatch        RefusalCode = "SCOPE_MISMATCH"
	RefusalExpired              RefusalCode = "LEASE_EXPIRED"
	RefusalStaleOwner           RefusalCode = "STALE_OWNER_FENCE"
	RefusalReplay               RefusalCode = "LEASE_REPLAY"
	RefusalTerminalReplay       RefusalCode = "TERMINAL_LEASE_REPLAY"
	RefusalInvalidTransition    RefusalCode = "INVALID_LEASE_TRANSITION"
	RefusalPretransportCrash    RefusalCode = "PRETRANSPORT_CRASH"
	RefusalBrokerRejected       RefusalCode = "BROKER_REJECTED"
	RefusalNotSubmitted         RefusalCode = "AUTHORITATIVE_NOT_SUBMITTED"
	RefusalTransportUncertain   RefusalCode = "TRANSPORT_UNCERTAIN"
	RefusalCapabilityUnattested RefusalCode = "RECOVERY_CAPABILITY_UNATTESTED"
	RefusalRetryExhausted       RefusalCode = "RECOVERY_RETRY_EXHAUSTED"
)

func validRefusalCode(code RefusalCode) bool {
	switch code {
	case RefusalNone, RefusalInvalid, RefusalDisabled, RefusalWaitMarket, RefusalBudgetDeferred, RefusalWorkerAbnormal,
		RefusalWorkerFailure, RefusalAuthorityDrift, RefusalScopeMismatch, RefusalExpired, RefusalStaleOwner,
		RefusalReplay, RefusalTerminalReplay, RefusalInvalidTransition, RefusalPretransportCrash, RefusalBrokerRejected,
		RefusalNotSubmitted, RefusalTransportUncertain, RefusalCapabilityUnattested, RefusalRetryExhausted:
		return true
	default:
		return false
	}
}

type LeaseState string

const (
	LeaseInvalid    LeaseState = ""
	LeaseIssued     LeaseState = "ISSUED"
	LeaseClaimed    LeaseState = "CLAIMED"
	LeaseSubmitting LeaseState = "SUBMITTING"
	LeaseSubmitted  LeaseState = "SUBMITTED"
	LeaseAmbiguous  LeaseState = "AMBIGUOUS"
	LeaseRefused    LeaseState = "REFUSED"
)

func terminalLeaseState(state LeaseState) bool {
	return state == LeaseSubmitted || state == LeaseAmbiguous || state == LeaseRefused
}

type ReservationDisposition string

const (
	ReservationInvalid     ReservationDisposition = ""
	ReservationReserved    ReservationDisposition = "RESERVED"
	ReservationReleased    ReservationDisposition = "RELEASED"
	ReservationTransferred ReservationDisposition = "TRANSFERRED"
	ReservationHeld        ReservationDisposition = "HELD"
)

type BrokerOutcome string

const (
	OutcomeAccepted                  BrokerOutcome = "ACCEPTED"
	OutcomeDefinitiveRejected        BrokerOutcome = "DEFINITIVE_REJECTED"
	OutcomeAuthoritativeNotSubmitted BrokerOutcome = "AUTHORITATIVE_NOT_SUBMITTED"
	OutcomeTransportUncertain        BrokerOutcome = "TRANSPORT_UNCERTAIN"
)

const (
	RuntimeRelease       = "a072-paired-strategy-runtime-v1"
	MaximumWorkerBackoff = 30 * time.Second
)

type trustedTime struct {
	now  time.Time
	seal [32]byte
}

func newTrustedTime(now time.Time) trustedTime {
	if now.IsZero() {
		return trustedTime{}
	}
	value := trustedTime{now: now.UTC()}
	value.seal = hashStrings(value.now.Format(time.RFC3339Nano))
	return value
}

func validTrustedTime(value trustedTime) bool {
	return !value.now.IsZero() && value.seal == hashStrings(value.now.UTC().Format(time.RFC3339Nano))
}
