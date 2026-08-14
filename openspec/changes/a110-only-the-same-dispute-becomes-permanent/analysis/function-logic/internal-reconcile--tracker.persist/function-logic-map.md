# Function Logic Map: `Tracker.persist`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| `Outcome.Added/Cleared` | deterministic pending additions and exact-cause releases | in-memory tracker + journal | additions stay fail-closed pending; releases publish only after commit |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | no journal | additions/release treated durable for isolated tests | unit tracker tests |
| B2–B3 | ordinary quantity block has blank symbol | remember the error, skip ambiguous write, continue representable additions and earned releases | blank-symbol journal/Restore and sibling-release regressions |
| B4 | Enter fails | return partial result/error | failed-enter tests |
| B5 | existing row owned by different cause | return authoritative replacement/error | cause-conflict test |
| B6 | earned releases are enumerated after additions | attempt every release reached before an unrepresentable blank error | sibling-release regression |
| B7 | release fails | return error without publishing that release | release failure test |
| B8 | exact release absent | return error | exact-cause release test |
| Return | release committed | append committed release and then return deferred blank error, if any | `TestA110BlankSymbolPendingDoesNotStarveValidSiblingRelease` |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `EnterReconcile` | durable fail-closed addition | idempotent; commit may be ambiguous | commit-then-timeout test |
| `ReleaseReconcile` | exact-cause release | journal confirmation required | B6–B8 |

## State mutations and fallbacks

- Returns classified durable/authoritative/released sets; caller mutates memory only from those confirmations.
- Blank-symbol ordinary evidence remains pending in memory because the schema cannot represent it distinctly from permanent account scope; its deferred error cannot starve representable writes or earned releases.

## Safety conclusion

- Safe boundary: skip only the ambiguous ordinary empty-symbol write, finish representable journal work, then surface its error; permanent account writes remain unchanged.
- High-risk impact: yes; journal is the restart authority.
