# Function Logic Map: `pureApprovedCandidateBoundaryViolations`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| package source | all parsed production files for one direct ApprovedCandidate reader | `auditApprovedCandidateBoundaries` | parse/type errors become fail-closed findings |
| allowed value graph | exact candidate.ApprovedCandidate, immutable scalar, value struct/fixed array recursively containing only allowed values | 4th security re-review contract | any other `go/types.Type` becomes a finding |
| allowed execution | direct ApprovedCandidate accessor call or safe non-mutating builtin only | 4th security re-review contract | helper/package/method-value/asserted calls and mutation/concurrency constructs become findings |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | import is not exact candidate package | append finding | continue | external/internal import RED fixture |
| B2 | `go/types` check fails | append type-check finding | continue syntax scan | malformed/capability fixture |
| B3 | param/result/receiver/struct field/array/type argument recursively contains forbidden type | append typed-path finding | continue | alias/named/generic/candidate.Source RED table |
| B4 | method, receiver, init, function literal, go/defer/send | append execution finding | continue | syntax RED table |
| B5 | call is direct ApprovedCandidate accessor or safe builtin | none | allow | scalar/value-only success fixture |
| B6 | call is method value/asserted/free helper/package/other | append finding | continue | call RED table |
| B7 | assignment/inc-dec LHS dereferences, indexes, or selects | append mutation finding | continue | mutation RED fixture |
| B8 | no violation | none | empty findings | scalar/value-only success fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `types.Config.Check` | resolve aliases, named underlying types, generics and candidate selector types | custom importer admits candidate only; errors fail closed | RED fixtures |
| cycle-safe recursive type predicate | validate the complete reachable value-type graph | a per-walk visiting set prevents recursion loops; cycles reject | named/embedded/array fixtures |
| AST with `types.Info` | distinguish direct approved accessors and safe builtins from capability calls/mutations | conservative deny on unresolved use | call/mutation fixtures |

## State mutations and fallbacks

- Only test-local maps and type metadata are mutated. No production state, network, filesystem authority, or trading capability is invoked.
- There is no fallback from a failed type check to spelling-only acceptance.
- Top-level `var` is rejected regardless of scalar type; local scalar identifiers remain permitted because their value cannot escape through a capability-bearing API.

## Safety conclusion

- Safe edit boundary: replace the spelling-oriented pure-boundary audit with a type-checked allowlist; production candidate/order logic is untouched.
- High-risk impact: yes — a false negative could inject authority without a module-internal import edge.
