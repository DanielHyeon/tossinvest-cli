(prompt-common.md 전문을 먼저 읽는다 — 아래는 보이스 C 의 초점.)

## 보이스 C — 시험 적합성 · 변이

- 새 시험과 수정된 교차 시험의 단언이 **공허하게 참**이 될 수 있는 자리(표본이 비는 전칭, 우연히 채워지는 문자열 포함 단언, 타임아웃으로 대신한 판정, 훅이 부르지 않으면 통과하는 단언)를 찾아라.
- 변이 원장(`analysis/mutation-unit2/` · `mutation-25.6/` · `mutation-unit3/`)이 덮지 않는 생산 코드 변경을 찾고, 스위트를 통과할 것으로 보이는 **구체적 변이**(파일:줄 · 바꿀 문자열)를 제시하라. 가능하면 실제로 확인하라: 대상 트리를 `/tmp/claude-1000/a092-r25-C-$$` 로 복사하고(만들기 전에 `set -euo pipefail`, 사본에서 `git rev-parse` 가 실패함을 단언 — 실제 저장소가 아님), 사본만 변이한 뒤 관련 패키지 `go test` 를 돌려 SURVIVED/CAUGHT 를 적는다(무변이 대조군 먼저).
- 교차 change 시험 편집 둘(`TestAcknowledgeCannotClearTheGateMidSend` · `TestAHeldRowIsNotWhispered`)이 원래 주제를 보존했는지 — 약해졌다면 어느 변이가 옛 시험에서는 잡히고 새 시험에서는 사는지.
