# Branch Test Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` (cmd/tossctl)

변이 원문: `../../mutation-2/mutation2-run10.txt`. RED = 편집 전 판본이 a166 `3dc0c776b` 뒤 console_test.go:755 로 실패(`../../boundary/boundary.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 737 새로고침 세 번 | 자신 | yes | yes |
| B2 | 738 Orders 오류 | 자신(오류 없는 변이만 둠) | no | yes |
| B3 | 742 빌드 ≠ 1 | 자신 — o2·o3 FAIL | yes(o2·o3) | yes |
| B4 | 749 `official.New(` | 자신 — o6 FAIL | yes(o6) | yes |
| B5 | 761 파싱 실패 | 자신(정상 소스에선 안 탐) | no | yes |
| B6 | 765 선언 순회 | 자신 | yes | yes |
| B7 | 767 `(*lazyOrders).Orders` 고르기 | 자신 — 대조군(Holdings 를 세지 않음) | yes | yes |
| B8 | 776 첫 순회 type-switch | 자신 | yes | yes |
| B9 | 777 선택자 갈래 | 자신 | yes | yes |
| B10 | 778 참조 세기 | 자신 — o1·o7 FAIL(참조 2) | yes | yes |
| B11 | 781 대입문 갈래 | 자신 | yes | yes |
| B12 | 782 단일 RHS 직접 호출 | 자신 — o5 FAIL(메서드 값, 호출 0) | yes | yes |
| B13 | 783 LHS 식별자 `_` 아님 | 자신 — o8 FAIL(결과 버림) | yes | yes |
| B14 | 793 둘째 순회 type-switch | 자신 | yes | yes |
| B15 | 794 대입문 갈래 | 자신 | yes | yes |
| B16 | 795 LHS 순회 | 자신 | yes | yes |
| B17 | 796 LHS 이름이 broker | 자신 — o9 FAIL(다시 대입) | yes | yes |
| B18 | 800 단항식 갈래 | 자신 | yes | yes |
| B19 | 801 `&broker`(같은 객체) | 자신 — o14 FAIL(포인터 쓰기) | yes | yes |
| B20 | 804 ValueSpec 갈래 | 자신 | yes | yes |
| B21 | 805 이름 순회 | 자신 | yes | yes |
| B22 | 806 이름이 broker | 자신 — o10 FAIL(안쪽 var 가림) | yes | yes |
| B23 | 810 단언 갈래 | 자신 | yes | yes |
| B24 | 811 consoleOrdersReader 단언 아님 | 자신 | yes | yes |
| B25 | 815 broker 객체에 대한 첫 단언 | 자신 — 생산 소스 | yes | yes |
| B26 | 817 그 밖의 단언 | 자신 — o11·o12·o13 FAIL | yes | yes |
| B27 | 824 다섯 수 판정 | 자신 — o1·o2·o3·o4·o5·o7~o14 FAIL; S 판본은 o1·o4 PASS(못 잡음) | yes | yes |
