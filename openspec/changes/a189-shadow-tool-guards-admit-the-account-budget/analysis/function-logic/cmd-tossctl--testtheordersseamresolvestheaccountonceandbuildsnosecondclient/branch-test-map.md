# Branch Test Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` (cmd/tossctl)

변이 원문: `../../mutation-2/mutation2-run5.txt`. RED = 편집 전 판본이 a166 `3dc0c776b` 뒤 console_test.go:755 로 실패(`../../boundary/boundary.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 737 새로고침 세 번 | 자신 | yes | yes |
| B2 | 738 Orders 오류 | 자신(오류 없는 변이만 둠) | no | yes |
| B3 | 742 빌드 ≠ 1 | 자신 — o2·o3 FAIL | yes(o2·o3) | yes |
| B4 | 749 `official.New(` | 자신 — o6 FAIL | yes(o6) | yes |
| B5 | 760 파싱 실패 | 자신(정상 소스에선 안 탐) | no | yes |
| B6 | 764 선언 순회 | 자신 | yes | yes |
| B7 | 766 `(*lazyOrders).Orders` 고르기 | 자신 — 대조군(Holdings 를 세지 않음) | yes | yes |
| B8 | 774 첫 순회 type-switch | 자신 | yes | yes |
| B9 | 775 선택자 갈래 | 자신 | yes | yes |
| B10 | 776 참조 세기 | 자신 — o1·o7 FAIL(참조 2) | yes | yes |
| B11 | 779 대입문 갈래 | 자신 | yes | yes |
| B12 | 780 단일 RHS 직접 호출 | 자신 — o5 FAIL(메서드 값, 호출 0) | yes | yes |
| B13 | 781 LHS 식별자 `_` 아님 | 자신 — o8 FAIL(결과 버림) | yes | yes |
| B14 | 791 둘째 순회 type-switch | 자신 | yes | yes |
| B15 | 792 대입문 갈래 | 자신 | yes | yes |
| B16 | 793 LHS 순회 | 자신 | yes | yes |
| B17 | 794 LHS 가 broker | 자신 — o9 FAIL(다시 대입) | yes | yes |
| B18 | 798 ValueSpec 갈래 | 자신 | yes | yes |
| B19 | 799 이름 순회 | 자신 | yes | yes |
| B20 | 800 이름이 broker | 자신 — o10 FAIL(안쪽 var 가림) | yes | yes |
| B21 | 804 단언 갈래 | 자신 | yes | yes |
| B22 | 805 `broker.(consoleOrdersReader)` | 자신 — o4·o7 FAIL 의 asserted 경로, 생산 소스에서 1 | yes | yes |
| B23 | 813 다섯 수 판정 | 자신 — o1·o2·o3·o4·o5·o7·o8·o9·o10 FAIL; S 판본은 o1·o4 PASS(못 잡음) | yes | yes |
