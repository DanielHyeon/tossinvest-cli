# CodeGraph evidence

Captured after `make sdd-sync` on 2026-07-30.

## Definitions and flow

- `cmd/tossctl.runConsole` constructs `localupdate.Updater`, execution-lock
  seams, and `console.Options`; it is the production composition root.
- `internal/console.Console.routes` registers every fixed HTTP route. Its
  state-changing routes are wrapped by `session0(mutating(...))`.
- `internal/console.Console.handleSettings` builds the settings model and calls
  `SystemUpdater.Inspect`; GET does not start account or update work.
- `internal/console.handleSystemUpdateInstall` holds `activityMu`, validates the
  reviewed candidate digest, takes engine/verification exclusions, calls
  `Install`, then commits relaunch.
- `internal/localupdate.Updater` fixes current, candidate, and rollback paths at
  construction. `Inspect` and `Install` use no-follow descriptors and
  `debug/buildinfo`; candidate code is never executed.

## Impact

- `runConsole` affects only `newConsoleCmd` and the `cmd/tossctl` composition
  unit at CodeGraph depth 3.
- `Console.routes` affects console construction/server/session wrappers and
  `runConsole`; 19 symbols were reported at depth 3.
- `Console.handleSettings` affects its `/settings` route and source unit only.
- No order, Guardian, gateway, journal, reconciliation, exit-observer, or engine
  loop symbol appeared in the definition/caller/impact closure.

## Edit boundary

Existing functions permitted to change are mapped under
`analysis/function-logic/`: `runConsole`, `newUpdateCmd`, `Console.routes`,
`Console.handleSettings`, `Updater.Inspect`, `Updater.Install`, and
`inspectOpen`. New release client/verifier/extractor/handler functions are
introduced separately. Any additional existing-function edit requires a new
AST, logic map, risk report, and branch-test map before modification.
