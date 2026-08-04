# Function Logic Map: `TestApprovedCandidateTaintRejectsAuthorityInterface`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| repository Go sources | candidate authority consumers must accept only approved snapshot | frozen source tree | test fails on tainted interface |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | repository walk/read fails | none | test failure | self |
| B2 | forbidden tainted consumer signature is found | none | test failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Walk`, `os.ReadFile`, `strings.Contains` | static authority-boundary inspection | local deterministic scan | AST |

## State mutations and fallbacks

- Test-only reads; no production mutation or fallback.

## Safety conclusion

- Safe edit boundary: extend forbidden consumer patterns with newly introduced proposal boundary.
- High-risk impact: yes; a raw candidate value must not become execution authority.
