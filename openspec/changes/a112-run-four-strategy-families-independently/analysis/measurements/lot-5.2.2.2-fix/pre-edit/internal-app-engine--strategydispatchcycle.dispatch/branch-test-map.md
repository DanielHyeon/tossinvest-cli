# Branch Test Map: `dispatch`

- Source SHA-256: `d9d29dfcc61759b835bce75e3f500885704f2f27ee9879056cdccbf4f50a5b3b`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 19 분기 → 22: 편집 전 B12(admission 거절 → 오류)안에 범위 거절 타입을 `%w` 로 싣는 B13 을 더했고, 편집 전 B14 앞의 위험 세대 읽기를 시장 번들에서 **그 범위의 번들**로 옮겼다(B15 범위 키 · B16 범위 번들). 편집 전 B13 → B14, B14 → B17, B15~B19 → B18~B22.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategydispatchcycle.dispatch/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 78:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 81:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 86:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 93:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 97:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 114:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 115:3 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B8 | if at 136:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | if at 140:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 144:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 148:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 152:2 — admission 거절 → 주기 오류(편집 전 B12) | `a112_owner_scope_trading_test.go` `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` | no — 조건 불변 | yes |
| B13 | if at 154:3 — **(새)** 거절이 범위 거절이면 타입을 `%w` 로 싣는다 — 문구는 같음(J4 ①: 분류는 타입으로) | `a112_owner_scope_handoff_test.go` `TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority` · `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` | yes — 변이 X06(모든 거절에 타입) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B14 | if at 160:2 — Guardian 결정 세대 없음(편집 전 B13) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B15 | if at 168:2 — **(새)** 계보의 범위 키 정규화 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` | no — 경로 추가 | yes |
| B16 | if at 169:3 — **(새)** 그 범위의 준비된 위험 번들에서 세대를 읽음(범위 번들이 없으면 세대 0 → B17 거절) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` | M20(시장 번들 세대) SURVIVED — **예상**: 모든 범위 번들이 같은 서명 시장 매니페스트에서 와서 세대가 구조상 시장 단위다 | yes |
| B17 | if at 173:2 — 서명 위험 정책 세대 없음(편집 전 B14) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B18 | if at 179:2 — 편집 전과 같은 분기(좌표만, 편집 전 B15) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B19 | if at 192:2 — 편집 전과 같은 분기(좌표만, 편집 전 B16) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B20 | if at 198:2 — 편집 전과 같은 분기(좌표만, 편집 전 B17) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B21 | if at 202:2 — 편집 전과 같은 분기(좌표만, 편집 전 B18) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B22 | if at 215:4 — 편집 전과 같은 분기(좌표만, 편집 전 B19) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
