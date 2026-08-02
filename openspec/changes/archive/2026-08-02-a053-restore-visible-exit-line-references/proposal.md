## Why

`/positions` used to show the journal baseline and initial stop directly. a042/a043 correctly changed actionable prices to require a canonical effective snapshot, but the additive migration intentionally left historical rows without that snapshot. Existing managed KR positions therefore look as if their protection information disappeared, while adoption-pending or reconcile-blocked US positions show only dashes even though an effective initial-stop policy is known.

## What Changes

- Keep canonical `ExitLine` authority and freshness behavior unchanged.
- Promote stored legacy entry/baseline/initial-stop/high-water evidence into the primary visual hierarchy whenever no fresh actionable snapshot can be shown, with an explicit non-effective label.
- For KR and US candidate positions that do not yet have an exit state, show `기준선 미생성`, why it is not established, and the running engine's effective initial-stop percentage when known. Do not synthesize a price from broker average or current price.
- Reject raw or canonical evidence from a different known lifecycle generation instead of showing an earlier position lifecycle's prices.
- Keep `/positions` read-only, responsive, and free of visible input controls.
- Preserve the API's existing separation between `exitLine` and `storedExitEvidence`, and add a read-only `exitLineReference` projection for the same legacy/plan explanation; no journal schema, order, reconcile, or toggle mutation is added.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `operator-console`: restores visible KR/US exit-line reference information while retaining canonical-vs-reference truth labels.
- `http-api-service`: exposes the same typed non-effective reference and lifecycle-generation refusal as the console.

## Impact

- `internal/operatorview`, console/API projections, OpenAPI documentation, templates, and regression tests change.
- Journal schema, exit evaluator, order path, broker calls, reconcile state, and configuration do not change. The public API receives one additive nullable read field; no route, mutation authority, or existing field semantics change.
