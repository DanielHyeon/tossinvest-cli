# CodeGraph baseline

- Base commit: `2dd1fe5ab419946b0ef75c770afe717b533c5383`
- Sync: `make sdd-sync` completed before queries.
- Task query: adoption percentage fixed at 5% under CSP plus console/engine
  journal identity and v9 compatibility.

## Hard evidence

- `cmd/tossctl.runConsole` unconditionally calls `journal.DefaultPath` and
  injects that value as `console.Options.JournalPath`.
- `internal/app/engine.openEngineJournal` joins `ConfigDir` with
  `journal.DBFileName`; without an override it delegates through `journal.Open`
  to the normal default.
- `cmd/tossctl.engineJournalDir` already implements the same profile split for
  the engine lock/marker.
- `internal/console.(*Console).handleSettingsSave` reads
  `default_stop_pct` as a fraction with `strconv.ParseFloat` and delegates to
  the injected settings seam.
- `settingsPage.StopPctSlider` currently returns a fraction and
  `StopPctPercent` multiplies it by 100.
- `internal/journal.(*ReadOnly).checkSchema` rejects too-new versions and
  missing tables but does not require the v9 `exit_states.policy_id` column.
- `LivePositionExits` calls `accountExitStates`, whose select uses that v9
  column before querying live positions.

## Callers / impact

- `runConsole` caller: `newConsoleCmd`; impact: console assembly only.
- `handleSettingsSave` caller: `/settings/save`; impact: route plus
  `AdoptionSettings.Load/Save`.
- `checkSchema` caller: `OpenReadOnly`; impact: read-only positions/history
  opening and its tests, not writable journal paths.

