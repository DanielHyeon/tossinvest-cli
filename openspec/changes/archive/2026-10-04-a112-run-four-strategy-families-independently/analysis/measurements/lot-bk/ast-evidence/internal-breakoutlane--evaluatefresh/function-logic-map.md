# Function Logic Map (편집 전): `evaluateFresh`

- Source: `internal/breakoutlane/machine.go`
- Source SHA-256: `3f2e013019481047008c2c7fbc6e365bab27467f4554f3e53ec0baef5e2488dc`
- Signature: `evaluateFresh(params=2, results=1)`
- Source range: `36:1`–`123:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 3be54204 — 편집 없음(BK2 는 시험만)).

## Inputs and invariants

- 편집 계획: 편집 없음 — breakout 덮개 2차 시험 · census 가 근거로 쓰는 분기 열거

## Branches and early returns

- Exact AST return nodes: `75:3`, `86:4`, `89:4`, `97:4`, `100:4`, `107:3`, `110:3`, `114:3`, `122:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 40:2 | `for _, bar := range bars[:v1OpeningRangeBars] {` |
| B2 | if | 41:3 | `if bar.value.HighMinor > resistance {` |
| B3 | if | 44:3 | `if bar.value.LowMinor < low {` |
| B4 | range | 52:2 | `for i, bar := range bars[v1OpeningRangeBars:] {` |
| B5 | if | 54:3 | `if b.valueHighAbove(resistance) && !BreakoutCloseQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) {` |
| B6 | if | 57:3 | `if BreakoutCloseQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) && b.RVOLPPM >= v.Config.value.RVOLMinPPM && b.UpperWickRangePPM <= v.Config.value.UpperWickRangeMaxPPM {` |
| B7 | if | 69:2 | `if breakout < 0 {` |
| B8 | if | 71:3 | `if firstTouch {` |
| B9 | if | 79:2 | `if v.Market == MarketUS {` |
| B10 | for | 82:2 | `for i := breakout + 1; i < len(bars); i++ {` |
| B11 | if | 85:3 | `if b.CloseMinor < low {` |
| B12 | if | 88:3 | `if since > timeout {` |
| B13 | if | 91:3 | `if retest && b.CloseMinor >= resistance {` |
| B14 | if | 96:3 | `if retest && b.CloseMinor < resistance && b.VolumeExpanded {` |
| B15 | if | 99:3 | `if since >= timeout {` |
| B16 | if | 102:3 | `if !retest && RetestQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) {` |
| B17 | if | 106:2 | `if state != phaseArmed {` |
| B18 | if | 109:2 | `if refusal := validateQuote(v.Quote, v.EvaluatedAtMS, v.Sizing.ProposedEntryMinor, v.Config); refusal != RefusalNone {` |
| B19 | if | 113:2 | `if result.Refusal != RefusalNone {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `string` | 48:89 |
| `string` | 48:114 |
| `b.valueHighAbove` | 54:6 |
| `BreakoutCloseQualifies` | 54:39 |
| `BreakoutCloseQualifies` | 57:6 |
| `append` | 64:20 |
| `string` | 64:42 |
| `string` | 64:57 |
| `newDecision` | 70:8 |
| `decisionSeal` | 73:13 |
| `len` | 82:29 |
| `uint64` | 84:12 |
| `newDecision` | 86:11 |
| `appendTransition` | 86:71 |
| `string` | 86:91 |
| `newDecision` | 89:11 |
| `appendTransition` | 89:68 |
| `string` | 89:88 |
| `append` | 92:20 |
| `string` | 92:42 |
| `string` | 92:66 |
| `newDecision` | 97:11 |
| `appendTransition` | 97:71 |
| `string` | 97:91 |
| `newDecision` | 100:11 |
| `appendTransition` | 100:68 |
| `string` | 100:88 |
| `RetestQualifies` | 102:17 |
| `newDecision` | 107:10 |
| `validateQuote` | 109:16 |
| `newDecision` | 110:10 |
| `size` | 112:12 |
| `newDecision` | 114:10 |
| `append` | 116:18 |
| `string` | 116:40 |
| `newDecision` | 117:7 |
| `hashFields` | 120:17 |
| `v.Config.Digest` | 120:67 |
| `decisionSeal` | 121:11 |

## Safety conclusion

- 생산 코드 변경 0 — 근거 열거만
