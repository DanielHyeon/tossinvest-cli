# Branch Test Map: `admit`

- Source SHA-256: `c31f12fd07855ab32d29c815c8b7b21e14c83add0cd3e1c46503bf01c15eda22`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 6 분기 → 7: 편집 전 B3(권한 수집 실패 → AuthorityCollectionFailed) 안에 범위 거절 타입을 결과에 싣는 B4 를 더했다. 편집 전 B4~B6 → B5~B7. 편집 전 번들은 base 016da624 기준(`pre-edit/internal-app-engine--strategyfirstlegadmissionbridge.admit`).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyfirstlegadmissionbridge.admit/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 75:2 — 결과 검증 거절 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 78:2 — bridge · loader · Guardian 부재 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 82:2 — 1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`(문구 = 수집 오류) | `a112_owner_scope_trading_test.go` `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` | no — 조건 불변 | yes |
| B4 | if at 84:3 — **(새)** 수집 오류가 `*strategyScopeRefusal` 이면(errors.As — 타입) 결과에 싣는다; 그 밖의 수집 오류는 싣지 않음(J4) | `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` · `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` · `a112_scope_refusal_census_test.go` census | yes — 변이 X05(모든 수집 오류에 타입) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B5 | if at 89:2 — 권한 불일치(편집 전 B4) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 93:2 — Guardian precheck 실패(편집 전 B5) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 97:2 — 원자 admission 실패(편집 전 B6) — 원장 `BUCKET_USAGE_STALE` 은 여기서 **타입 없이** 올라감 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(같은 파도 둘째 범위 — 타입 아님 단언) | no — 조건 불변 | yes |
