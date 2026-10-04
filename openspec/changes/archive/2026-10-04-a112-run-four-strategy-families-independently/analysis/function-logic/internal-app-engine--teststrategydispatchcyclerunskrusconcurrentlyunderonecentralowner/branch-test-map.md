# Branch Test Map: `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`

- Source SHA-256: `b0b9734d75c5e4fafa2b2033bd9c3d9660af813aa7d485b1840d2a1f8ebcd958`; AST branch locations are authoritative.
- Revision: **재작성(a066 5.6.1 교차 편집 d78f3f4a — 공유 bucket KR/US 동시 dispatch 를 5.6.1 계약으로; a112 게이트 위생 재측정 2026-10-01).** 9 분기 → 10: 결과 분류가 switch(B3~B6)로 바뀌고 성립 · 거절 짝 단언(B7)과 lease 의 중앙 owner 대조(B10)가 섰다. 편집 전 번들은 이 파일의 git 이력.

이 함수는 시험 자신이다. 분기는 단언 실패 경로이므로 「그 분기를 도는 시험」은 이 함수 자체이고, 초록은 어느 실패 arm 도 돌지 않았다는 뜻이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 115:2 — KR · US 두 dispatch 를 동시에 출발시키는 순회 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B2 | range at 133:2 — 결과 둘을 받는 순회 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B3 | switch at 134:3 — 결과 분류 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B4 | case at 135:3 — 성립(Confirmed) — admitted 에 담음 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B5 | case at 137:3 — 공유 bucket 두 번째 진입의 `ATOMIC_ADMISSION_FAILED`(BUCKET_USAGE_STALE) — refused 에 담음 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B6 | case at 143:3 — 그 밖의 결과 — 실패 단언 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B7 | if at 147:2 — 성립 하나 · 거절 하나 · 서로 다른 시장이 아니면 실패 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B8 | if at 153:2 — Gateway 호출이 성립한 시장 하나가 아니면 실패 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B9 | if at 157:2 — lease 읽기 실패 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
| B10 | if at 163:2 — lease 의 owner epoch · fencing token 이 중앙 owner(첫 epoch)와 다르면 실패 | `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` | n/a — 시험 코드(a066 재작성) | PASS(엔진 태그 스위트, a112 5.2.2.2 격리 검증) |
