# Function Logic Map: `exitOrderLine.args`

- Source: `internal/app/engine/exit_order_log.go:140-169`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| price_source | `lower_limit`, `observed`, `baseline` | `exitOrderLineFor` | invalid value cannot be constructed |
| whitelist fields | fixed scalar values | exit order line | raw broker payload never emitted |
| fill values | present only for follow-up fill | journal fill record | omitted otherwise |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fill follow-up | append fill fields | return args | a129 fill source golden |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitOrderToken` | sanitize state/reason | malformed token becomes fixed fallback | AST base args |
| `append` | emit fill-only fields | no error | AST B1 |

## State mutations and fallbacks

- `price_source` is a required fixed key in every line before optional fill fields.
- existing field ordering remains stable except the additive source key adjacent to price.

## Safety conclusion

- Safe edit boundary: whitelist serialization only.
- High-risk impact: yes — evidence must identify the actual submitted price origin.
