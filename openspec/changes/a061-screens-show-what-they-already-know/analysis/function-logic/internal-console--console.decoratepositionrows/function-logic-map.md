# Function Logic Map: `Console.decoratePositionRows`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a061-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | one HTTP request | handler | cancellation surfaces as a runtime/list read error, which stays unknown |
| `rows []positionRow` | request-local rows | `joinPositions` | mutated in place |
| `asOf` | the render instant | `Console.now` | passed through, never re-read |
| `c.opts.PositionPolicies` | nil means the legacy console with no lifecycle seam | wiring | nil -> lifecycle stays unknown, exit lines still decorated |
| `c.opts.Settings` | nil means no desired-config seam | wiring | nil -> `Designated`/`Excluded` stay false |
| `c.opts.EngineMarker` (a061, read via `protectionLiveness`) | path or empty | wiring | empty -> the pre-a061 observation-age bound |

**Invariant**: this is the single boundary where `/positions` and `/dashboard`
decorate rows, which is what stops the two screens from giving one holding two
different management or protection answers. a061 keeps the liveness read here for
the same reason.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.opts.PositionPolicies != nil` | sets `runtimeAttempted`, reads runtime | falls through | existing a052 tests |
| B2 | `List` returned no error | builds `policyByID` | falls through | existing a052 tests |
| B3 | `for _, state := range states` | fills the index | none | existing |
| B4 | `c.opts.Settings != nil` | attempts the desired-config read | falls through | existing include/exclude tests |
| B5 | `Load` returned no error | stamps both lists from one read | falls through | existing |
| B6 | `for i := range rows` | stamps `Designated`/`Excluded` | none | existing |
| B7 | `runtimeAttempted` | runs the lifecycle/management pass | falls through | existing a052 tests |
| B8 | `for i := range rows` | per-row lifecycle projection | none | existing |
| B9 | `row.InJournal` | sets `LifecycleProofRequired` | falls through | existing |
| B10 | policy row absent for this position | `journalKnown = false` | falls through | existing lifecycle-unknown test |
| B11 | policy row present | records status, generation, released | falls through | existing |
| B12 | `row.Management.Block != nil` | builds the reconcile block view | none | existing reconcile tests |

**a061 changes exactly one line**: the tail call now passes
`c.protectionLiveness(asOf)` as a fourth argument. No branch was added, removed
or reordered.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PositionPolicies.Runtime` | running engine settings | error deliberately discarded; `EffectiveKnown` stays false | AST |
| `PositionPolicies.List` | lifecycle proof per position | error leaves `policyByID` nil, which reads as unverified | AST |
| `Settings.Load` | desired include/exclude | error leaves both flags false | AST |
| `Console.protectionLiveness` (a061) | is anything maintaining the lines | reads a file modification time; a missing or unreadable marker reads as not running, never as running | `internal/console/protection_liveness.go` |
| `attachPositionExitLines` | the fail-closed exit projection | pure over the rows | AST |

No broker call, journal write, order or config save is reachable from this path.
`protectionLiveness` adds one `stat`, no network and no lock.

## State mutations and fallbacks

- Writes only to `rows[i]` display fields.
- Every seam failure keeps the row explicitly unknown rather than substituting a
  desired or default value.
- a061 adds no new failure mode: an unreadable marker is not-running, and an
  unwired marker takes the branch that behaves exactly as before.

## Safety conclusion

- Safe edit boundary: one added argument at one call site.
- High-risk impact: no. Display decoration only.
- Safety invariant 0.3 untouched.
