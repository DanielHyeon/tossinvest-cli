# Function Logic Map: `readinessFixture.marketInputForBody`

- Source: `internal/protectionreadiness/attestation_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| attestation body and private key | fixture-owned exact signed data | test fixture | fatal test on canonical/file/supervisor construction error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | canonicalization or JSON/file/binding construction fails | test-only allocation | fatal test | fixture consumers |
| B2 | construction succeeds | creates sealed in-memory fixture | return exact scope/file/supervisor | attestation suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| canonical/sign/file/binding helpers | create cryptographically valid fixture | fatal on error | CodeGraph + AST |

## State mutations and fallbacks

- Test-only in-memory/file value construction; exact quantity bounds copied from signed body.

## Safety conclusion

- Safe edit boundary: fixture scope mirrors body exactly
- High-risk impact: no (test helper)
