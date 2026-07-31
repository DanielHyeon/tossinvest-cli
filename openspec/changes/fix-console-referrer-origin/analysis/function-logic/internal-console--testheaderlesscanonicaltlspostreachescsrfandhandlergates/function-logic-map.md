# Function Logic Map: `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates`

- Source: `internal/console/origin_fallback_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| remote test rig | canonical TLS/Host and process CSRF | `newRemoteTestRig` | test setup fails immediately |
| wrapped handler counter | starts at zero | test closure | detects unexpected gate traversal |
| headerless requests | valid or wrong CSRF | `remoteRequest` | assertions distinguish handler and CSRF paths |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | valid headerless request does not return 204 | none | fatal test failure | this test |
| B2 | valid headerless request does not call handler once | none | fatal test failure | this test |
| B3 | wrong-CSRF request is not the expected CSRF 403 | none | fatal test failure | this test |
| B4 | wrong-CSRF request changes handler count | none | fatal test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRemoteTestRig` | build isolated remote console | fatal on setup error | AST |
| `Console.mutating` | exercise real origin and CSRF gate ordering | synchronous response | AST |
| `remoteRequest` / `httptest.NewRecorder` | create direct TLS request and capture response | in-memory only | AST |

## State mutations and fallbacks

- Mutates only a local invocation counter and in-memory HTTP recorders.
- No operational handler, engine, order, setting, or external endpoint is used.

## Safety conclusion

- Safe edit boundary: unchanged inherited regression; map exists because adding
  the adjacent opaque-origin test changes the parsed function region.
- High-risk impact: yes, as evidence for the remote mutation security gate.
