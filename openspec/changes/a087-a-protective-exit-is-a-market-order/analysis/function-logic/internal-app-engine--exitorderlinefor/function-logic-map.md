# Function Logic Map: `exitOrderLineFor`

- Source: `internal/app/engine/exit_order_log.go:103-112`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selected limit | lower-limit, observed, baseline, empty | `submit` single decision | sent price must equal line price |
| observed | original observed string | quote | never overwrite log field |
| baseline | persisted state string | exit state | fallback only if observed empty |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | 고른 price/source와 원 관측가를 사본에 함께 고정 | return line | lower_limit/observed/baseline golden |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출 없음 | — | — |

## State mutations and fallbacks

- signature gains selected price/source while `observedPrice` stays original.
- lower-limit source is legal only when selected price came from the immutable board.

## Safety conclusion

- Safe edit boundary: a129 diagnostic line only; broker write path is untouched.
- High-risk impact: yes — log must not assert a different price than the intent.
