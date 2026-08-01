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
| parsed POST form | at most 4096 bytes; action-specific allowlist; exactly one value per field | route middleware + handler validation | oversized is 413; unknown/duplicate is 400 before commander call |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | strict action-specific form validation fails | refusal only | 400 before commander dispatch | unexpected/duplicate request tests |
| B2 | commander absent | refusal only | 501 | unwired command test |
| B3-B5 | apply request succeeds or maps capability/CAS error | lifecycle DB/audit only on service success | redirect or explicit 4xx/412 | capability/confirmation/conflict tests |
| B6-B8 | rollback preview parses exact versions/category and calls service | durable candidate only; settings unchanged | preview or explicit refusal | rollback lifecycle tests |
| B9-B11 | finite server-option preview parses exact fields and calls service | durable candidate only; settings unchanged | preview or 400/error mapping | option-only/invented-option tests |

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
- Request parsing is bounded before CSRF evaluation, and action dispatch accepts only its exact server-rendered field set.

## Safety conclusion

- Safe edit boundary: finite option selection → preview capability → CAS apply.
- High-risk impact: yes; fail-closed validation, one-shot token, 3-second risk delay, confirmation and atomic audit are mandatory.
