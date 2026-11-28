# CodeGraphContext / GBrain supporting evidence (a192, 1단계)

## CodeGraphContext — not-applicable(이 worktree)

- `codegraphcontext version` → 0.5.1. `codegraphcontext stats .` → `Repository not found: /home/daniel/tossos-work/a192-missing-order-block-release`.
- 설정은 전역 falkordb 다(a191 기록). 이 worktree 를 색인하면 공유 그래프 DB 에 저장소가 하나 더 쓰인다. 1단계 지시(공유 상태를 건드리지 않음)와
  advisory 등급을 함께 보고 **색인하지 않았다**. 공유 체크아웃의 CGC 색인도 보지 않았다(안 봄 — 없음이 아님).

## GBrain — not-applicable(이 worktree)

- `python3 tools/sdd/gbrain_project.py search "missing order reconcile block release"` → `No brain configured. Run: gbrain init`. 데이터 홈은 worktree
  로컬이고 새 worktree 에는 없다. `make sdd-sync` 는 GBrain 단계가 worktree HEAD 를 전진시킬 수 있어(기억 `sdd-sync-gbrain-fast-forwards-worktree`) 돌리지 않았다.

## 대신 한 보조 탐색(손으로 — 모두 현재 HEAD 로 확인)

| 후보 | 결과 |
|---|---|
| `reconcile_states` 를 읽거나 쓰는 비시험 파일 | `rg -l 'reconcile_states\|ActiveReconcileStates\|ReconcileState\b'` → reconcile/mismatch.go·recovery.go, journal 6, execgw 4, engine 1, cmd 2 — design Q5 표 |
| release cause 소비자 | `risk_bucket_owner.go` 의 broker-zero(Q5 ⑧) — 새 cause 를 받지 않음 |
| 차단이 청산을 바꾸는 자리 | `exitwiring.go` `reconcileFloor`, `exitloop.go` `applyFloor`(Q5 ③) |
| 결속에 쓸 종결 근거 | `scoped_fill_snapshots`(6요소 키·`terminal`·`fail_closed`·`committed_at`), `Journal.LookupFillScoped` |
| 같은 cause 를 쓰는 다른 생산자 | `position_projection.go` `enterReconcileInTx`·`reconcileCauseFor`(과매도 → `QUANTITY_MISMATCH`) |
| 선례 | a083(크레딧·as-of 엄격 나중), a110(같은 분쟁만 영구), a191(종결 목록 확인), risk_bucket 의 `reconcile_states` 열 추가(v33 계열 ALTER) |
