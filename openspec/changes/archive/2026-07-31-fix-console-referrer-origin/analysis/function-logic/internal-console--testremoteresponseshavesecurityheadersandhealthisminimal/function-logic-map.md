# Function Logic Map: `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal`

- Source: `internal/console/remote_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| remote login response | accepted peer/Host | remote test rig | exposes wrapper headers without authentication mutation |
| security headers | all present; referrer exactly `same-origin` | `remoteRuntime.security` | assertion failure |
| loopback health requests | GET and POST | `/healthz` wrapper exception | GET must be minimal; POST method-rejected |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate required security headers | read-only header access | continue through list | this test |
| B2 | a required header is empty | none | non-fatal test failure | this test |
| B3 | referrer policy differs from `same-origin` | none | non-fatal test failure | this test |
| B4 | loopback GET health differs from 200 `ok` | none | fatal test failure | this test |
| B5 | loopback POST health is not 405 | none | fatal test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRemoteTestRig` | build isolated remote console | fatal on setup error | AST |
| `serveRemote` | exercise complete remote wrapper | in-memory HTTP | AST |
| `Header.Get` | inspect effective response policy | pure read | AST |

## State mutations and fallbacks

- Mutates only in-memory request/response fixtures.
- Pins all existing wrapper headers while adding an exact referrer-policy
  contract; no operational route executes.

## Safety conclusion

- Safe edit boundary: add only the exact referrer header assertion.
- High-risk impact: yes, as regression evidence for a shared remote wrapper.
