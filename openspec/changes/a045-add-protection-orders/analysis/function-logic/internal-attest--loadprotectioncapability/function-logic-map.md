# Function Logic Map: `LoadProtectionCapability`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | `path`, `now`, exact runtime scope, current effective UID; all are untrusted until every check succeeds. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 current UID unavailable → file error; B2 filesystem checks/open/read/decode fail → typed refusal; B3 matrix/window/scope mismatch → typed refusal; B4 all checks pass → capability returned. | Reads one local file only. Existing flaw: parsing and verification are fused and external evidence bytes are never rehashed. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `os.Geteuid`, `loadProtectionCapability`, matrix validation/scope verification; no network, retry, broker, or config binding. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Reads one local file only. Existing flaw: parsing and verification are fused and external evidence bytes are never rehashed.

## Safety conclusion

- Safe edit boundary: Split parse from verification; make verified output constructible only after evidence-byte digest and canonical matrix binding checks.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
