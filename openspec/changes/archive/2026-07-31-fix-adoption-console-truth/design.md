## Context

The CLI has two profile-resolution contracts. Most console and engine state
follows an explicit `--config-dir`. The engine journal follows that contract in
`openEngineJournal`, but `runConsole` currently calls `journal.DefaultPath`
unconditionally. In the deployed container this produces two valid databases:
the engine writes `/var/lib/tossos/config/journal.db` at schema v9 while the
positions page reads `/var/lib/tossos/data/journal.db` at schema v8.

The adoption page renders a range input and an inline `oninput` handler. The
console deliberately sends `default-src 'none'` and no `script-src` exception,
so the browser refuses that handler. Weakening CSP to make one control work
would expand the attack surface of every console page.

This change touches journal selection and a protective-width setting, so it is
High-risk even though it changes neither the journal schema nor the
exit-policy calculation.

## Goals / Non-Goals

**Goals:**

- Make console and engine select the same journal for the same root profile.
- Preserve a type-level read-only journal boundary in the console.
- Make the current and edited adoption stop percentage unambiguous in browsers
  that enforce the existing CSP.
- Validate the console's 2% through 20% band on the server and persist the
  existing fractional JSON representation.
- Prove both fixes with isolated tests that cannot contact a broker or start an
  engine.

**Non-Goals:**

- No journal migration, copy, merge, deletion, or schema change.
- No rewrite of existing `position_adoptions`, `exit_states`, baselines, or
  policy identifiers.
- No CSP relaxation and no client-side script for the adoption percentage
  control. Unrelated legacy inline confirmation handlers are outside this
  change.
- No change to exit-policy math, automation gates, engine lifecycle, or order
  authority.
- No automatic engine restart after saving a percentage.

## Decisions

### D1. Use the engine profile rule as the journal source of truth

A small resolver in `cmd/tossctl` will return
`<config-dir>/journal.db` when `--config-dir` is set and otherwise delegate to
`journal.DefaultPath`. `runConsole` will inject that resolved path into the
existing read-only console option.

This mirrors `openEngineJournal` without handing the console a writable
`Journal`. Reusing the stale data-directory database or copying it into the
profile was rejected because either choice could hide or duplicate durable
trading history.

### D2. Fix identity and reject an incompatible selected journal before queries

The positions page will continue to use `journal.OpenReadOnly`. This change
does not migrate v8. The read-only schema check will additionally require the
v9 `exit_states.policy_id` column used by `LivePositionExits`; its absence is
classified as the existing typed too-old state before any account query runs.
That preserves the current operator instruction to start the engine once for
migration while preventing a raw `no such column` failure.

Adding `policy_id` fallbacks to queries was rejected because it would make an
incompatible journal appear readable while omitting the policy provenance the
current binary requires. Falling back to a different journal was rejected for
the same identity reason.

### D3. Submit a human-readable percentage without script

The form will use a native numeric percentage control named
`default_stop_percent`, with value `5`, min `2`, max `20`, and step `0.5`. The
server will reject empty, non-finite, non-numeric, out-of-band, and non-half-
percentage values before division, divide the accepted value by 100, and pass
the resulting fraction to the existing `config.Adoption` save seam.

Rendering uses deterministic decimal formatting with no unnecessary trailing
zeros. A legacy config fraction that is valid to the engine but not aligned to
the console's 0.5 percentage-point grid is displayed exactly with an
instruction to choose the nearest allowed value; submitting it unchanged is
refused rather than silently rounded.

This works with keyboard, touch, and native browser steppers and needs no
inline handler. Keeping the range plus a frozen output was rejected because it
continues to misreport the selected value; adding `unsafe-inline` script was
rejected because one control does not justify weakening console CSP.

### D4. Existing settings remain restart-bound

Saving changes only `engine.adoption` and its audit entry. A running engine
keeps the snapshot it started with, and the existing page notice continues to
say that a restart is required. The fix will not restart it automatically
because that would combine a protective-setting mutation with a process
lifecycle mutation.

## Risks / Trade-offs

- **Risk: Console and engine path rules drift again.** → Put the resolver behind
  a focused command-level test that compares explicit-profile and default
  behavior, and retain the existing read-only static guards.
- **Risk: Percent/fraction unit confusion widens or narrows protection by
  100×.** → Give the HTML field a percent-specific name, parse finite half-
  percentage ticks in one helper before division, cover the full 2..20 grid,
  and prove `7.5 → JSON 0.075 → config load 0.075`.
- **Risk: Existing managed positions are assumed to change immediately.** →
  Preserve the current next-engine-start notice and do not touch journal rows.
- **Trade-off: Native number controls look different across browsers.** → The
  value is explicit and accessible; correctness under CSP outweighs identical
  visual styling.

## Migration Plan

1. Build and test the replacement image without changing live configuration or
   process state.
2. In an isolated config, verify the positions reader selects the profile
   journal and the settings form round-trips a non-default percentage.
3. Production container recreation is a separate, explicitly human-authorized
   deployment step because container shutdown stops the engine and autostart
   may start it again.

Rollback is data-safe because config values remain fractions and no database
format changes. Rolling back the image also reintroduces both defects: the old
console reads the default data journal instead of the profile journal and its
adoption value display depends on CSP-blocked script.

## Open Questions

None. The deployed paths, schema versions, CSP response, rendered form, and
persisted config were all observed directly before design freeze.
