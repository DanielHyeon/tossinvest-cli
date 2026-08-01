# Function Logic Map: `seedPolicyIdentity`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| kind/id/adoption/claimed identity | valid complete tuple or exact zero legacy seam | fixed compatibility registry | mismatch/unknown conflict |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | ratchet kind | resolve pinned ratchet identity | error propagates | seed tests |
| B2 | ladder kind | resolve pinned ladder/adopted identity | error propagates | seed tests |
| B3 | claimed tuple zero | preserve exact pinned compatibility | return expected | legacy tests |
| B4 | claimed invalid/mismatched | none | identity conflict | seed mismatch test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Legacy*PolicyIdentity` | obtain only provable pre-a042 meaning | unknown fails closed | AST |
| `PolicyIdentity.Validate` | reject partial/malformed tuple | error propagates | AST |

## State mutations and fallbacks

- Typed runtime identity is returned; no a042 column is written.

## Safety conclusion

- Safe edit boundary: pure seed validation.
- High-risk impact: yes — prevents state reinterpretation.
