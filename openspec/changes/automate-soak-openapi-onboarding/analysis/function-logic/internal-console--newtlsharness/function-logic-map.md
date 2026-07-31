# Function Logic Map: `newTLSHarness`

- Source: `internal/console/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| option tweaks | zero or more credential-test callbacks | onboarding tests | setup fails test immediately |
| transport | real `httptest` TLS | `r.TLS` direct connection state | client trusts only generated test certificate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | apply option tweaks | test-only option mutation | continues | onboarding suite |
| B2 | console construction fails | none | fatal test | onboarding suite |
| B3 | cookie jar construction fails | none | fatal test | onboarding suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `New` | build real console handler | fatal on setup error | CodeGraph + AST |
| `httptest.NewTLSServer` | exercise direct TLS boundary | cleanup closes server | AST |
| `srv.Client` | trust generated certificate and preserve cookies | test-local only | AST |

## State mutations and fallbacks

- The TLS helper is isolated to secret-ingress tests; production remote peer, Host, and origin tests remain in `remote_test.go`.

## Safety conclusion

- Safe edit boundary: test-only direct TLS harness.
- High-risk impact: no, test infrastructure only.
