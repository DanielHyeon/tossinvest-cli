# Branch Test Map: `ExitObserver.alertProposalRefused`

**편집 전 좌표**(AST 0 분기, base `064f59fd9`, `exitloop.go` `1901`–`1918`). 실행 0 — 덮는 시험은 이름 grep 으로만 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거절된 보호 발의 → `exit.proposal_refused` critical 1 | `TestARefusedProposalReleasesTheLevelAndAlerts` | 미측정 | 미측정 |

본문 문장을 단언하는 시험은 저장소에 0(grep `닿기 전에 거부` · `청산 주문이 제출되지 않았다` — 시험 파일의 제목 문자열은 `obs` 픽스처뿐). D9 의 RED 는 tasks 3.1(design §4-2 · §4-12).
