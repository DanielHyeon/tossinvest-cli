# Function Logic Map: `canonicalProtectionMatrix`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| validated matrix | UTC time and deterministic evidence/capability order | signed schema | noncanonical semantics rejected before digest acceptance |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | copy evidence without self-referential capability digest and marshal stable field order | none | bytes/error | reorder/digest tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | deterministic wire bytes after semantic order validation | error propagates | CodeGraph + AST |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: require one semantic order before hashing/signing.
- High-risk impact: yes.
