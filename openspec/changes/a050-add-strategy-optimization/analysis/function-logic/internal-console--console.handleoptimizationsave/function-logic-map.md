# Function Logic Map: `Console.handleOptimizationSave`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| lifecycle commander | required for preview/apply | `Console.Options` | 501 without side effect |
| action token/capability | opaque server-issued value only | command service | invalid/expired/replayed refused |
| option ID | finite owner registry option | `settingmeta.FieldDescriptor` | invented value refused |
| version | bound into server token, never trusted raw | command service snapshot | stale base returns 412 |
| risk confirmation | checkbox + server reason code after 3 seconds | preview capability | missing/early returns refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | commander absent | refusal only | 501 | unwired command test |
| B2 | preview request | durable candidate + opaque preview | render preview | option-only preview test |
| B3 | invalid/invented option | no mutation | 400 | invented option test |
| B4 | apply without valid capability/confirmation | no mutation | 4xx | capability/confirmation tests |
| B5 | stale CAS | no mutation | 412 | conflict test |
| B6 | successful apply or rollback-as-new-candidate | lifecycle DB/audit only | redirect with version/restart | apply/history/rollback tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OptimizationCommander.Preview` | validates changed subset and mints one-shot token | no retry; error rendered | CodeGraph + AST |
| `OptimizationCommander.Apply` | atomic CAS, history/audit and token consumption | idempotent command ID; stale base 412 | CodeGraph + AST |
| `Console.refuse/render/http.Redirect` | explicit result without client script | centralized CSP/session/CSRF wrappers | CodeGraph + AST |

## State mutations and fallbacks

- Only optimization desired/effective lifecycle state is mutable.
- Types expose no broker, journal, lane, gate, kill-switch, or LIVE setters.
- Rollback creates a new candidate/version; it never rewrites history.

## Safety conclusion

- Safe edit boundary: finite option selection → preview capability → CAS apply.
- High-risk impact: yes; fail-closed validation, one-shot token, 3-second risk delay, confirmation and atomic audit are mandatory.
