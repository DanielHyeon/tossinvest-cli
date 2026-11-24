# CodeGraphContext supporting evidence — a197 (tasks 1.2)

- CodeGraphContext **0.5.1**(`codegraphcontext version`, falkordb).
- `codegraphcontext stats ~/tossos-work/a197-impl` → `Repository not found`(격리 worktree 미색인).
- `codegraphcontext stats /mnt/D/Axipient/workspace/TossOS` → Files 0 · Functions 0(주 저장소 색인도 비어 있음).
- 결론: **보조 문맥 0 — 「그 채널은 안 봤음」이 아니라 「그 채널에 볼 것이 없었음」**. 색인(`codegraphcontext index`)은 3,237 파일 규모라 1단계에서 돌리지 않았다(지시: `make sdd-sync` 금지, 전체 범위 명령 금지). advisory 이므로 판정에 쓰지 않는다.

## GBrain (advisory, 주 저장소 데이터 홈 읽기 전용 검색)

| 질의 | 결과 |
|---|---|
| `opposite pending order cancel external buy` | No results |
| `IssueReduction cancel` | No results |
| `LiveOrdersForSymbol` | `internal/journal/fills.go:1849-1927`(**낡음** — HEAD 는 `1854-1919`), a062 · a094(아카이브 전 경로) AST 산출물 |
| `EngineBrokerOrderIDs` | a136 아카이브 AST 산출물 |

GBrain 색인은 현재 HEAD 보다 낡았다(좌표 · 경로 불일치). 누락 후보로 쓴 것: a136 의 `EngineBrokerOrderIDs`(엔진 귀속 판정 후보 — `cited-ast/README.md` (c)). 그 밖의 결론은 CodeGraph · AST · 현재 소스로만 냈다.
