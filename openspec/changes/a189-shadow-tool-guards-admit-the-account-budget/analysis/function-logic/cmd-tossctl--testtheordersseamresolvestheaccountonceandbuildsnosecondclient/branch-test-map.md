# Branch Test Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` (cmd/tossctl)

변이 원문: `../../mutation-2/mutation2-run3.txt`. RED = 편집 전 판본이 a166 `3dc0c776b` 뒤 console_test.go:755 로 실패(`../../boundary/boundary.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 737 새로고침 세 번 | 자신 | yes | yes |
| B2 | 738 Orders 오류 | 자신(변이에서 안 탐 — 오류 없는 변이만 둠) | no | yes |
| B3 | 742 빌드 ≠ 1 | 자신 — 변이 o2·o3 FAIL(S 판본에서도 o2 FAIL) | yes(o2·o3) | yes |
| B4 | 749 `official.New(` | 자신 — 변이 o6 FAIL | yes(o6) | yes |
| B5 | 758 파싱 실패 | 자신(정상 소스에선 안 탐) | no | yes |
| B6 | 762 선언 순회 | 자신 | yes | yes |
| B7 | 764 `(*lazyOrders).Orders` 고르기 | 자신 — 대조군(Holdings 의 같은 철자를 세지 않음) | yes | yes |
| B8 | 771 공유 resolver 참조 세기 | 자신 — o1(2)·o4(0) FAIL, o5(메서드 값, 1) PASS | yes(o1·o4) | yes(o5) |
| B9 | 777 선언 1·참조 1 아님 | 자신 — o1·o2·o3·o4 FAIL. 비교: S 판본은 o1·o4 PASS(못 잡음) | yes | yes |
