# CodeGraphContext / GBrain supporting evidence (a191, 1단계)

## CodeGraphContext — not-applicable(이 worktree)

- `codegraphcontext version` → 0.5.1. `codegraphcontext stats .` → `Repository not found: /home/daniel/tossos-work/a191-filldetect-closed-list`.
- 설정은 전역 falkordb(`~/.codegraphcontext/.env`)다. 이 worktree 를 색인하면 공유 그래프 DB 에 저장소가 하나 더 쓰인다. 1단계 지시(공유 상태를
  건드리지 않음)와 advisory 등급을 함께 보고 **색인하지 않았다**. 공유 체크아웃의 CGC 색인도 보지 않았다(안 봄 — 없음이 아님).

## GBrain — not-applicable(이 worktree)

- `python3 tools/sdd/gbrain_project.py search "open list single order read PENDING cancelled closed list"` → `No brain configured. Run: gbrain init`
  (rc 1). 데이터 홈 `.sdd/gbrain-home/` 은 worktree 로컬이고 새 worktree 에는 없다. `make sdd-sync` 는 GBrain 단계가 worktree HEAD 를
  전진시킬 수 있어(기억 `sdd-sync-gbrain-fast-forwards-worktree`) 돌리지 않았다.

## 대신 한 보조 탐색(손으로 — 모두 현재 HEAD 로 확인)

| 후보 | 결과 |
|---|---|
| CLOSED 목록 조회 선례 | `execgw.Resolver.scanBoth`(indoubt.go:409), `amend_indoubt.go:415 findSuccessors` — 둘 다 OPEN+CLOSED, 바이트 키 |
| 같은 결함 계열 기록 | `verify-execution-capability/measurements.md` M25·M28·M31·M32, openapi `paths./api/v1/orders.get` |
| 감지기 요청의 예산 등급 | `internal/accountbudget` `Priority`, exitloop 의 `Critical` 표시 3곳, 감지기 표시 0 |
| 콘솔의 CLOSED 읽기 | `cmd/tossctl/console.go:1059~1086`(OPEN 전량 + CLOSED 100) — Q9 대안 |
| 경보 통로 후보 | `obs.Notifier`·`obs.Logger`, `EntryGate.BlockSymbol` |
