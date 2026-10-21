# a163 CodeGraphContext 보조 문맥 (SDD 4단계)

- 질의: `codegraphcontext find name adaptPrices --no-fuzzy` · `find name quoteUsable --no-fuzzy` (2026-10-21, 각 rc=0). 원문 `cgc-find.txt`.
- 결과: 각 20건(표시 상한). 전부 **다른 워크트리·임시 사본**(`/tmp/tossos-*`, `TossOS-worktrees/*`, 옛 경로 `/mnt/D/project/axipient/TossOS`)이고
  이 체크아웃 경로(`/mnt/D/Axipient/workspace/TossOS`)의 항목은 표시된 20건 안에 없다. `quoteUsable` 줄 번호도 1068~1137 로 갈려 현재 HEAD(1179)와 다르다.
- 판정: CGC 는 이 저장소의 현재 상태를 대표하지 않는다 → 누락 후보 발견용으로도 쓸 결과가 없다. hard evidence 는 CodeGraph + AST + 현재 HEAD 원문.
