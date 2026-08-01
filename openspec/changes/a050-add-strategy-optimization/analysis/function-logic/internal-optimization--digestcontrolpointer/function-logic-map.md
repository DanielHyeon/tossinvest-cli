# Function Logic Map: `digestControlPointer`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current version/snapshot digest | version 0 with empty digest, or positive version bound to its verified snapshot digest | control row and immutable snapshot | deterministic domain-separated SHA-256; mismatch rejected by readers/CAS |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless canonical JSON plus SHA-256 happy path | local allocation only | lowercase hex digest | control-pointer tamper/migration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| JSON marshal and SHA-256 | binds pointer version to exact snapshot digest with domain tag | fixed supported types, deterministic | tamper tests |

## State mutations and fallbacks

- Pure helper. It cannot repair a mismatched pointer and uses no wall-clock or database fallback.

## Safety conclusion

- Safe edit boundary: control-pointer integrity identity.
- High-risk impact: yes; rollback pointer tampering must be detected before read/preview/apply.
