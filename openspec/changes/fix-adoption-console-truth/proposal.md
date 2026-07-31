## Why

The deployed console can read a different journal from the engine whenever the
profile is selected with `--config-dir`. That path split makes managed holdings
appear unknown and lets a stale v8 journal fail on the v9 `policy_id` query.
The adoption stop-width slider also relies on an inline event handler that the
console's deny-by-default CSP blocks, so the displayed value remains 5%.

## What Changes

- Resolve the console journal from the same active profile rule used by the
  engine: an explicit `--config-dir` owns `journal.db`; otherwise use the
  normal data-directory journal.
- Keep the console journal strictly read-only and perform no migration or
  journal write; classify a selected pre-v9 journal before a query reaches the
  missing `exit_states.policy_id` column.
- Replace the script-dependent synthetic-stop slider interaction with a
  CSP-compatible percentage control whose submitted value is converted and
  validated server-side.
- Preserve the existing 2% through 20% console band, config fraction format,
  next-engine-start activation, audit trail, and exit-policy calculation.
- Add regression coverage proving stale unrelated journals cannot influence
  the positions screen, incompatible selected journals fail with an actionable
  state, and non-default percentages survive the settings form and config
  round trip.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `operator-console`: Require the positions view to follow the engine's active
  profile journal and require adoption stop-width editing to work without
  script execution.

## Impact

Affected code is limited to console assembly/path resolution, read-only journal
compatibility checks, the adoption settings form and save parser, and their
tests. No public API, database schema or migration, existing position baseline,
order path, engine loop, operating toggle, or dependency changes.
