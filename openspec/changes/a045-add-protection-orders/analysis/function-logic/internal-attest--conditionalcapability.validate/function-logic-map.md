# Function Logic Map: `ConditionalCapability.validate`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | One untrusted account/profile/market/session/type/trigger/quantity/persistence/reservation/idempotency/replace row. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 identity invalid; B2 unsupported enum/type; B3 unsafe quantity; B4 missing persistence/reservation/idempotency/replace guarantee; else valid. | No mutation. Existing gap: digit-stripping accepts arbitrary surrounding characters. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | Strict account canonicalizer plus enum/boolean checks only. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Existing gap: digit-stripping accepts arbitrary surrounding characters.

## Safety conclusion

- Safe edit boundary: Use a strict allowlisted account grammar and compare canonical values only.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
