# Function Logic Map: `AuthorityOrigin.Valid`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque origin token | private production bit set only by Client.AuthorityOrigin | official package | zero token returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | none | returns private production bit | authority-origin tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | field read only | pure | AST |

## State mutations and fallbacks

- Exposes validity without exposing a constructor or writable field.

## Safety conclusion

- Safe edit boundary: token must remain opaque outside the package.
- High-risk impact: yes, gates official FX evidence creation.
