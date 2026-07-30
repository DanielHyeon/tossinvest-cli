# Function Logic Map: `Updater.Inspect`

- Source: `internal/localupdate/updater.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| updater fixed paths | absolute current + sibling candidate | constructor | constructor error |
| candidate state | no-follow regular executable or absent | filesystem | non-installable reason |
| updater mutex | one inspection/stage/install at a time | `Updater.mu` | waits |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | candidate absent | none | non-installable/없음 | existing inspect test |
| B2 | candidate invalid | none; never executes | non-installable/reason | existing refusal tests |
| B3 | valid candidate/platform | none | metadata/installable | existing metadata test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `inspectPath` | validate fixed candidate | no-follow/buildinfo/hash | CodeGraph + AST |
| `replacementSupported` | platform install capability | Unix-only | CodeGraph + AST |

## State mutations and fallbacks

- Add mutex only; inspection semantics remain unchanged.

## Safety conclusion

- Safe edit boundary: lock then delegate to an unexported locked helper if needed.
- High-risk impact: yes; serialization is the candidate TOCTOU boundary.
