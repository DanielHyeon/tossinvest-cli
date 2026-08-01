# Function Logic Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability allowlist | valid test/domain fixture | closed enumeration | fail test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | exact method/exemption set | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleCapabilities` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- exact method/exemption set.

## Safety conclusion

- Safe edit boundary: not-applicable test evidence.
- High-risk impact: no.
