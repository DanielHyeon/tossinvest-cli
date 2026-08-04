package strategyrouter

type LegacyState struct {
	Enabled        bool
	Disabled       bool
	SelectedMarket Market
	Verified       bool
	Combined       bool
	Record         MarketRecord
}

type MigrationResult struct {
	State     SchedulerState
	Code      RefusalCode
	Duplicate bool
}

func MigrateLegacy(state SchedulerState, legacy LegacyState, migrationVersion string) MigrationResult {
	if migrationVersion == "" || !ValidSchedulerState(state) {
		return MigrationResult{State: cloneSchedulerState(state), Code: RefusalMigration}
	}
	if state.MigrationVersion == migrationVersion {
		return MigrationResult{State: cloneSchedulerState(state), Code: state.MigrationCode, Duplicate: true}
	}
	if state.MigrationVersion != "" {
		return MigrationResult{State: cloneSchedulerState(state), Code: RefusalMigration}
	}
	next := NewSchedulerState()
	next.MigrationVersion = migrationVersion
	code := RefusalNone
	switch {
	case legacy.Disabled && !legacy.Enabled:
		// Explicit legacy disabled state maps to two independently disabled records.
	case legacy.Enabled && legacy.Verified && !legacy.Combined && validMarket(legacy.SelectedMarket) && validMarketRecord(legacy.Record) && legacy.Record.Market == legacy.SelectedMarket:
		next.Records[legacy.SelectedMarket] = legacy.Record
	default:
		code = RefusalMigration
	}
	next.MigrationCode = code
	return MigrationResult{State: next, Code: code}
}
