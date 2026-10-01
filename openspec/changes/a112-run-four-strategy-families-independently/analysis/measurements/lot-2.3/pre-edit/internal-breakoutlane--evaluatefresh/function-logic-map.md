# Function Logic Map (편집 전): `evaluateFresh`

- Source: `internal/breakoutlane/machine.go`
- Source SHA-256: `3f2e013019481047008c2c7fbc6e365bab27467f4554f3e53ec0baef5e2488dc`
- Signature: `evaluateFresh(params=2, results=1)`
- Source range: `36:1`–`123:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD d54f4dca).

## Inputs and invariants

- 편집 계획: Manager 판정 2.3 (b): 입장하지 못한 돌파 봉 중 close buffer · wick 을 통과하고 1.2 <= RVOL 인 봉(=RVOL 만 1.5 미달)에 1.2 반사실 기록(p.RVOLAt1200000=true)을 더한다. B6 입장 갈래(57:3)와 그 안의 반사실 세 줄(60~62)은 손대지 않는다 — 새 if 는 B6 뒤(입장 시 break 로 건너뜀)에만 선다.

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

- High-risk(돌파 판정 · 증거). 기록 전용 — 상태 · 전이 · 거절 · 수량 · 봉인(decisionSeal 은 provenance 플래그를 포함하지 않음)에 닿지 않는다. 입장 경로(B6) 불변.
