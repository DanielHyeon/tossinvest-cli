# Function Logic Map (편집 전): `size`

- Source: `internal/breakoutlane/sizing.go`
- Source SHA-256: `61f517ed2fee9ddb3f7afea5753d6eb383956441b8df73f691ac014ee27bfff0`
- Signature: `size(params=4, results=1)`
- Source range: `33:1`–`99:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 3be54204 — 편집 없음).

## Inputs and invariants

- 편집 계획: 편집 없음 — 사이징 신탁 · 비보호 stop 시험의 근거 열거

## Branches and early returns

- Exact AST return nodes: `35:3`, `38:3`, `41:3`, `44:3`, `48:3`, `56:3`, `60:3`, `64:3`, `67:3`, `73:4`, `78:3`, `82:3`, `89:3`, `96:3`, `98:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 34:2 | `if !fxValid(f) {` |
| B2 | if | 37:2 | `if q.value.Currency != f.value.InstrumentCurrency {` |
| B3 | if | 40:2 | `if in.ProposedEntryMinor == 0 // q.value.AskMinor == 0 // in.StopMinor == 0 // in.StopMinor >= in.ProposedEntryMinor {` |
| B4 | if | 43:2 | `if in.TargetMinor <= in.ProposedEntryMinor {` |
| B5 | if | 47:2 | `if r != RefusalNone {` |
| B6 | if | 51:2 | `if q.value.AskMinor > base {` |
| B7 | if | 55:2 | `if ov {` |
| B8 | if | 59:2 | `if ov {` |
| B9 | if | 63:2 | `if ov {` |
| B10 | if | 66:2 | `if in.TargetMinor <= costExit {` |
| B11 | if | 70:2 | `if in.MinRiskRewardPPM > 0 {` |
| B12 | if | 72:3 | `if ov // netReward < need {` |
| B13 | if | 77:2 | `if r != RefusalNone {` |
| B14 | if | 81:2 | `if r != RefusalNone {` |
| B15 | if | 85:2 | `if n := notional / worst; n < candidate {` |
| B16 | if | 88:2 | `if candidate == 0 {` |
| B17 | if | 92:2 | `if in.FinalCap < final {` |
| B18 | if | 95:2 | `if final == 0 {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fxValid` | 34:6 |
| `convertCost` | 46:18 |
| `checkedAdd` | 54:15 |
| `checkedAdd` | 58:14 |
| `checkedAdd` | 62:18 |
| `mulDivCeil` | 71:15 |
| `convertCapacity` | 76:15 |
| `convertCapacity` | 80:17 |

## Safety conclusion

- 생산 코드 변경 0
