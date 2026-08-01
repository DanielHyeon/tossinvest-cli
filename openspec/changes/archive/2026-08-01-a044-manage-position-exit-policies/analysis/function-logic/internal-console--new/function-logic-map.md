# Function Logic Map: `New`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Options` | required verification seam; optional narrow capabilities | console API | constructor refuses missing verifier/invalid remote |
| generated tokens | cryptographically random, process local | console auth contract | constructor returns error if future signing setup fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `StartVerify == nil` | none | `ErrNoVerifyWiring` | console constructor tests |
| B2 | nil clock | install UTC system clock | continue | console tests |
| B3 | remote invalid | none | error | remote tests |
| B4-B6 | nil output/binary and boot note | local defaults/display state only | continue | console tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRemoteRuntime` | validate remote boundary | fail closed | remote tests |
| `newToken` | mint session/CSRF/action signing material | crypto/rand | static/token tests |
| `c.routes` | freeze authenticated route graph | no network I/O | route static tests |

## State mutations and fallbacks

- Initializes process-local auth/signing state and read caches; no journal or broker write is opened here.

## Safety conclusion

- Safe edit boundary: add only opaque-action signing material; never accept token material in `Options`.
- High-risk impact: yes — console approval boundary; negative token/CSRF tests required.
