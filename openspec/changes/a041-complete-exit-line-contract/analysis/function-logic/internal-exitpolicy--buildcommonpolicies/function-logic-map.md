# Function Logic Map: `buildCommonPolicies`

- Source: `internal/exitpolicy/common_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| static common policy table | stable IDs, semantic versions, validated exact decimal/rational parameters | a041 OpenSpec and existing StockOS defaults | process initialization panics if a shipped policy cannot produce an identity |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate every static profile | verifies canonical identity against the declared literal | returns registered profiles | pinned digest contract test |
| B2 | identity construction fails | no partially usable registry escapes initialization | panic during process initialization | policy identity validation tests |
| B3 | canonical digest differs from fixed declaration | no registry escapes initialization | panic | pinned digest drift test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `LadderPolicy.Identity` | canonicalizes and hashes exact semantics | returns validation error; static configuration panics | CodeGraph + AST |

## State mutations and fallbacks

- The local table already declares fixed digests; construction verifies rather than invents them.
- Existing target, stop, partial, and runner numbers are preserved.

## Safety conclusion

- Safe edit boundary: static registry construction; no runtime position or order mutation.
- High-risk impact: yes, because a digest mismatch must fail closed before exit evaluation.
