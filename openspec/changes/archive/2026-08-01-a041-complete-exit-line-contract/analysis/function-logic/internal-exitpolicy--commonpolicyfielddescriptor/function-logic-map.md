# Function Logic Map: `CommonPolicyFieldDescriptor`

- Source: `internal/exitpolicy/common_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| common policy descriptors | finite server-owned option set with stable IDs | `RegisteredCommonPolicyDescriptors` | field validation rejects missing or unregistered options |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate common policies | appends one radio-tile option per immutable policy | returns finite option field | metadata descriptor contract tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RegisteredCommonPolicyDescriptors` | obtains immutable policy identities and descriptions | shipped configuration is fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Builds a fresh option slice; default/effective remain explicitly unapproved while HYBRID is only recommended.

## Safety conclusion

- Safe edit boundary: transport-neutral metadata only; no UI control or settings mutation.
- High-risk impact: yes, because arbitrary text must never select an exit policy.
