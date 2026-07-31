# Function Logic Map: `RegisteredCommonPolicyDescriptors`

- Source: `internal/exitpolicy/common_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| registered common policies | deep-copied, valid immutable profiles | package registry built by `buildCommonPolicies` | impossible shipped identity fails at initialization/descriptor construction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate registered profiles | appends a value descriptor | returns a detached descriptor slice | common policy descriptor contract test |
| B2 | profile identity fails | no invalid descriptor is returned | panic for invalid shipped configuration | policy identity validation tests |
| B3 | iterate ladder rungs | appends stable T1..Tn rows with exact strings | descriptor keeps exact policy values | common policy descriptor contract test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RegisteredCommonPolicies` | obtains deep copies of registered profiles | no runtime failure | CodeGraph + AST |
| `LadderPolicy.Identity` | binds metadata to immutable semantics | invalid shipped configuration panics | CodeGraph + AST |

## State mutations and fallbacks

- Allocates fresh descriptor and rung slices; it does not expose mutable registry slices.

## Safety conclusion

- Safe edit boundary: transport-neutral read model construction only.
- High-risk impact: yes, because displayed parameters and executable identity must remain identical.
