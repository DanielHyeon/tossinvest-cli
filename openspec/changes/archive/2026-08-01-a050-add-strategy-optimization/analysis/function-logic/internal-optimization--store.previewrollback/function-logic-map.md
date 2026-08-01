# Function Logic Map: `Store.previewRollback`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| base/target/category/actor | base equals current; target immutable snapshot exists; only active registry keys in requested category | append-only snapshots and current registry | conflict/invalid/historical-key error; no candidate inserted before validation |
| rollback changes | exact desired-value delta between target and current | persisted snapshots | unchanged/wrong-category fields omitted; no client-authored values |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | current snapshot read fails | none | error | corruption tests |
| B2 | base version stale | none | `ErrVersionConflict` | rollback conflict test |
| B3 | target snapshot missing/corrupt | none | invalid candidate | invalid target test |
| B4-B5 | enumerate current and target desired keys | local set only | continue | added/removed key rollback test |
| B6 | inspect union of keys | local change map only | continue | rollback diff test |
| B7 | historical key no longer active | none | `ErrHistoricalKeyInactive` | inactive key test |
| B8 | field outside category or already equal | none | skip | category/no-op test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| current/target snapshot reads | binds rollback to immutable history | any corruption fails closed | snapshot corruption tests |
| `Registry.Field` | prevents restoring retired/unowned writable key | exact key lookup | historical-key test |
| `Store.preview` | applies ordinary candidate validation/MAC/actor lifecycle to derived delta | one call/no retry | rollback lifecycle tests |

## State mutations and fallbacks

- Builds only local key/change maps until the fully derived rollback request is delegated to normal preview. History is never rewritten.
- No unknown historical key, cross-category value, or no-op is invented into a candidate.

## Safety conclusion

- Safe edit boundary: append-only rollback candidate derivation.
- High-risk impact: yes; rollback changes settings and must use the same capability/actor/CAS gates.
