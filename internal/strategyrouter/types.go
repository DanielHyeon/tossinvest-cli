package strategyrouter

import "fmt"

// Market is an explicit trading-market scope. Only KR and US are supported.
type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

func validMarket(market Market) bool {
	return market == MarketKR || market == MarketUS
}

// Horizon identifies a strategy cadence. It is deliberately absent from OwnerKey.
type Horizon string

const (
	HorizonShort  Horizon = "SHORT"
	HorizonWeekly Horizon = "WEEKLY"
)

func validHorizon(horizon Horizon) bool {
	return horizon == HorizonShort || horizon == HorizonWeekly
}

type DesiredState string

const (
	StateOff DesiredState = "OFF"
	StateOn  DesiredState = "ON"
)

func validDesiredState(state DesiredState) bool {
	return state == StateOff || state == StateOn
}

type RuntimeState string

const RuntimeUnobserved RuntimeState = "UNOBSERVED"

type RefusalCode string

const (
	RefusalNone                   RefusalCode = ""
	RefusalInvalid                RefusalCode = "INVALID"
	RefusalDisabled               RefusalCode = "DISABLED"
	RefusalAmbiguous              RefusalCode = "AMBIGUOUS"
	RefusalReconstructionMismatch RefusalCode = "RECONSTRUCTION_MISMATCH"
	RefusalOwnerSnapshotStale     RefusalCode = "OWNER_SNAPSHOT_STALE"
	RefusalScopeMismatch          RefusalCode = "SCOPE_MISMATCH"
	RefusalVersionConflict        RefusalCode = "VERSION_CONFLICT"
	RefusalMigration              RefusalCode = "MIGRATION_REFUSED"
	RefusalBudgetDeferred         RefusalCode = "BUDGET_DEFERRED"
	RefusalReplay                 RefusalCode = "REPLAY"
	RefusalDuplicate              RefusalCode = "DUPLICATE"
)

func boundedNonEmpty(name, value string) error {
	if value == "" {
		return fmt.Errorf("strategyrouter: %s is empty", name)
	}
	if len(value) > 256 {
		return fmt.Errorf("strategyrouter: %s exceeds 256 bytes", name)
	}
	return nil
}
