# Function Logic Map: `TestOnlyTheListedFilesCanNameTheChaseVerdict`

- Source: `internal/candidate/consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| repository Go files and explicit `verdictReaders` map | every Chase verdict consumer is listed with a review reason | candidate consumer guard | test fails on an unlisted consumer or stale allowlist entry |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | repository walk/parser fails or finds too few files | none | test fatal | this guard itself |
| B2 | file does not name a verdict | records nothing | continue | this guard itself |
| B3 | verdict consumer is not allowlisted | reports file and safety explanation | test error | RED when `httpapi_reader.go` first added |
| B4 | allowlist entry no longer consumes verdict | reports stale permission | test error | this guard itself |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `goFilesUnder`, `parser.ParseFile`, `namesVerdict` | mechanically find all verdict consumers | parser failures are fatal; no skipped production file | AST + focused candidate tests |

## State mutations and fallbacks

- Test-only state records seen paths. It grants no runtime capability.
- a051 adds `httpapi_reader.go` solely as a read-only projection; the paired order-verb guard proves that file cannot submit orders.

## Safety conclusion

- Safe edit boundary: one allowlist entry with a narrow read-only reason.
- High-risk impact: yes; Chase is buy eligibility, so both consumer and order-verb guards must pass.
