# Branch Test Map: `ExitObserver.alertProposalRefused`

**편집 뒤 좌표**(AST 3 분기, base `064f59fd9` → 재고정 `b93ed133a`, `exitloop.go` `1927`–`1952`). 실행 0 — 덮는 시험은 이름 grep 으로만 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기본 본문(브로커 전 거절) | `TestARefusedProposalReleasesTheLevelAndAlerts` · `TestA197TheRefusalBodyNamesWhoRefused`(「refused in-engine」 — 본문 단언) | no | yes |
| B2 | 반대 미체결 거절 본문(토글 무관) | `TestA197TheToggleOffReadsAndCancelsNothing` | yes | yes |
| B3 | 그 밖의 브로커 거절 본문 | `TestA197TheRefusalBodyNamesWhoRefused`(「another broker no」 — gstack 리뷰 r2 testing 지적으로 추가) | no | yes |

2026-11-25 gstack 리뷰 r2 전에는 본문 문장을 단언하는 시험이 저장소에 0(grep `닿기 전에 거부` · `청산 주문이 제출되지 않았다` — 시험 파일의 제목 문자열은 `obs` 픽스처뿐). D9 의 RED 는 tasks 3.1(design §4-2 · §4-12).
