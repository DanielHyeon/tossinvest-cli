# Function Logic Map: `TestExplicitOpaqueOriginCannotReachMutationHandler`

- Source: `internal/console/origin_fallback_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request transport/Host | canonical direct TLS and configured Host | `remoteRequest` | prevents unrelated refusal reasons |
| `Origin` | explicit opaque string `null` | test header | must be final invalid evidence |
| CSRF | current valid process token | test rig | proves CSRF cannot rescue opaque origin |
| handler counter | zero | test closure | detects security-gate bypass |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | response is not origin-specific HTTP 403 | none | fatal test failure | this test |
| B2 | wrapped handler was invoked | increments local counter | fatal test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRemoteTestRig` | build isolated remote console | fatal on setup error | AST |
| `Console.mutating` | exercise actual origin-before-CSRF ordering | synchronous response | AST |
| `remoteRequest` / `httptest.NewRecorder` | construct/capture in-memory request | no external I/O | AST |

## State mutations and fallbacks

- Mutates only a local invocation counter and in-memory response recorder.
- Uses a valid CSRF token deliberately so explicit opaque origin remains the
  deciding refusal and headerless TLS+Host fallback cannot apply.

## Safety conclusion

- Safe edit boundary: security regression test only; production origin
  evaluation remains unchanged.
- High-risk impact: yes, as direct evidence for the remote mutation boundary.
