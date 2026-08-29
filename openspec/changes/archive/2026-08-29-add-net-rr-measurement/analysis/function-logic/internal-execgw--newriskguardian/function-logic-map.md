# Function Logic Map: `NewRiskGuardian`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Validates the Guardian's configuration and derives the limit snapshot. This change added one line: the optional `Observer` is copied onto the struct.

It is deliberately **not** validated. A Guardian with no observer issues exactly as it did before, which is what keeps the measurement from being load-bearing — refusing to construct one would make an analysis feature a precondition for trading.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Journal` | non-nil | wiring | refuses to construct |
| `opts.AccountRef` | non-empty | wiring | refuses to construct |
| `opts.PolicyVersion` | non-empty | wiring | refuses to construct |
| `opts.Policy` | `Validate()` clean | configuration | refuses to construct |
| `opts.Costs` | `Configured()` | configuration | refuses to construct |
| `opts.Observer` | nil or any `EntryObserver` | wiring | **nil is valid** — recording is optional |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`switch` @ internal/execgw/riskguardian.go:244) | `opts.Journal == nil` | none | error | riskguardian_test.go construction cases |
| B2 (`case` @ internal/execgw/riskguardian.go:245) | `account == ""` | none | error | riskguardian_test.go |
| B3 (`case` @ internal/execgw/riskguardian.go:247) | `opts.PolicyVersion` blank | none | error | riskguardian_test.go |
| B4 (`case` @ internal/execgw/riskguardian.go:249) | `opts.Policy.Validate()` error | none | error | riskguardian_test.go |
| B5 (`if` @ internal/execgw/riskguardian.go:253) | `!opts.Costs.Configured()` | none | error | riskguardian_test.go |
| B6 (`if` @ internal/execgw/riskguardian.go:256) | `ExposureLimitsFor` error | none | error | riskguardian_test.go |
| B7 (`if` @ internal/execgw/riskguardian.go:261) | `EncodeLimits` error | none | error | riskguardian_test.go |
| B8 (`if` @ internal/execgw/riskguardian.go:265) | `opts.Clock == nil` | defaults to `clock.System()` | continues | riskguardian_test.go |
| B9 (`if` @ internal/execgw/riskguardian.go:270) | `opts.TTL <= 0` | defaults to `DefaultDecisionTTL` (60s) | continues | riskguardian_test.go |
| B10 (`if` @ internal/execgw/riskguardian.go:274) | `opts.NewID == nil` | defaults to `randomID` | continues | riskguardian_test.go |
| B11 (`if` @ internal/execgw/riskguardian.go:278) | the constructed struct literal's field set | copies `opts.Observer` verbatim, unvalidated | the Guardian | `TestAGuardianWithNoObserverIssuesAsBefore`, `TestTheCostBasisIsRecordedWithEveryRow` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `opts.Policy.Validate` | a policy the chain cannot evaluate is a refusal to exist | pure | AST B4 |
| `ExposureLimitsFor` / `EncodeLimits` | derives and encodes the stamped limit snapshot | pure | AST B6/B7 |

## State mutations and fallbacks

- None outside the returned struct. No I/O, no clock read beyond defaulting.
- Live binding added: `observer: opts.Observer`. It is the only new field and it is nil-safe at every use site (`observeEntry` returns immediately on nil).

## Safety conclusion

- Safe edit boundary: one field copy in a struct literal, with no new validation and no new failure mode. Every existing refusal-to-construct is untouched.
- High-risk impact: **yes by path**, no by effect. A Guardian built by existing wiring (no Observer field) is byte-identical in behaviour to the pre-change one.
