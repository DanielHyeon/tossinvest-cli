# Function Logic Map: `loadRemoteCertificate`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| certificate/key paths and canonical host | non-empty readable PEM pair; host covered by current ServerAuth leaf | shared `networkboundary.LoadServerCertificate` | returns wrapped validation error; no listener is opened |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | shared loader rejects key pair, SAN, validity, or EKU | none | wrapped `console:` error | remote certificate mismatch tests |
| B2 | shared loader accepts the exact private certificate | returns parsed leaf in `tls.Certificate` | certificate, nil | remote console TLS construction tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `networkboundary.LoadServerCertificate` | share TLS identity validation with a051 daemon | fail closed; no retry or fallback to unverified TLS | AST + console remote tests |

## State mutations and fallbacks

- Reads certificate/key files through the shared loader and returns an in-memory certificate only.
- No fallback accepts an expired, wrong-host, or non-ServerAuth leaf.

## Safety conclusion

- Safe edit boundary: delegate existing console validation to the stricter shared network boundary helper.
- High-risk impact: yes; wrong acceptance would expose remote console/API identity, so full console TLS tests are required.
