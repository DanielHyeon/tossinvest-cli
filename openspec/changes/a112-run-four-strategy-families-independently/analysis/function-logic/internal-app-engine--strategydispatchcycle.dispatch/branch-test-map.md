# Branch Test Map: `dispatch`

- Source SHA-256: `72021ef372fe658540d68292cf0556181ad8a981ce81d807eb0903e3b55c67a7`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리: B13 의 조건을 `admitted.scope != nil` 에서 `admitted.cause != nil` 로 — 수집 오류를 사슬째(`%w`) 나른다(범위 거절 타입 · 결함 원인 모두). 분기 수 · 위치 불변(22).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--strategydispatchcycle.dispatch/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 78:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 81:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 86:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 93:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 97:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 114:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 115:3 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B8 | if at 136:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | if at 140:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 144:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 148:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 152:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B13 | if at 155:3 — 수집 오류가 있으면 사슬째 `%w`(문구 = Detail), 없으면 문구만 | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(STALE 은 타입 없음) | yes — 변이 Y20(%v) · Y21(원인 없는 거절에 범위 타입) CAUGHT | yes |
| B14 | if at 161:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B15 | if at 169:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B16 | if at 170:3 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B17 | if at 174:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B18 | if at 180:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B19 | if at 193:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B20 | if at 199:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B21 | if at 203:2 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B22 | if at 216:4 — 80ae96a5 와 같은 분기 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
