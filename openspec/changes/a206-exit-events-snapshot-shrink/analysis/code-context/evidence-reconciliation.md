# a206 증거 조정 — exit_events 판정 증거 열을 읽는 곳

- 기준 commit: `e9266bfb6`(origin/main), worktree `~/tossos-work/a206`, 2026-11-27.
- CodeGraph **1.6.0**, worktree 에서 `codegraph init .` (57,792 nodes · 212,969 edges).
- CodeGraphContext: **not-applicable** — 이 단계는 코드를 편집하지 않고, 질문이 「어느 함수가 이 열을 읽는가」라는 닫힌 열거라
  CodeGraph callers + `rg` 문자열 전수가 정답 집합을 줌. 보조 문맥 후보 확장은 구현 change 의 몫으로 남김.

## 질의와 결과

| 질의 | 결과 |
|---|---|
| `rg -n "snapshot_json" -g '!*_test.go'`(openspec · docs 제외) | exit_events 의 세 JSON 을 SELECT 하는 생산 코드는 `exit_state.go:870 Journal.ExitEvents` · `account_views.go:303 ReadOnly.AccountExitEvents` · `account_views.go:370 ReadOnly.BrokerOrderExitLinks` 셋. 나머지 `effective_snapshot_json` 읽기(`exit_observation_refresh.go:239` · `exit_quarantine_list.go:43` · `exit_state.go:680`)는 **exit_states**(현재 상태 한 행) 의 열임. `readonly.go:88-91` 은 열 **존재** 검사(OpenReadOnly 가 열이 없으면 거절) |
| `rg -n "(FROM\|JOIN\|INTO\|UPDATE\|DELETE FROM) exit_events"` 생산 코드 | 위 셋 + 스냅숏을 읽지 않는 7 곳: `a087_protective_exit_attempt.go:26`(action) · `a166_replay_origin.go:18,59`(action · 존재) · `exit_state.go:494`(decision_id 중복) · `trade_outcomes.go:614`(proposed_intent_id) · `trade_outcome_backfill.go:209`(action=COMPLETED) · `provenance.go:240,284`(action · level_after · proposed_intent_id). UPDATE · DELETE 0 곳 |
| `rg -l exit_events tools scripts cmd auth-helper analysis` | 0 파일 — Python · 셸 도구 · CLI 는 이 테이블을 직접 읽지 않음 |
| `codegraph callers hydrateExitEventEvidence` | AccountExitEvents · BrokerOrderExitLinks · scanExitEvent (3) |
| `codegraph callers scanExitEvent` | Journal.ExitEvents (1) |
| `codegraph callers ExitEvents` | 20 — 전부 시험 함수 · 시험 헬퍼(`a111Events` · `readA200Ledger` · `Test…`). `rg "\.ExitEvents\(" -g '!*_test.go'` 도 0 → **생산 호출자 없음** |
| `codegraph callers AccountExitEvents` | 생산 2: `console overviewLedger`(overview.go:545) · `console history`(portfolio.go:698), 나머지 시험 6 |
| `codegraph callers BrokerOrderExitLinks` | 생산 1: `console ledgerOrderEvidence`(orders.go:702), 나머지 시험 15 |
| `codegraph affected exit_snapshot.go account_views.go exit_state.go --filter '*_test.go'` | 시험 파일 1,538(engine 259 · journal 169 · cmd/tossctl 121 · execgw 93 · console 84 …) — 파일 단위 의존이라 넓음. 기본 판별식은 Go 시험을 못 찾으므로(.claude/CLAUDE.md 하네스 주석) `--filter` 를 씀 |
| 엔진 · execgw 가 쓰는 exit_events 리더 | `HasProtectiveExitAttempt`(exit_lower_limit.go:213) · `HasExitEventForIntent` · `ExitProposalForIntent`(execgw/a166_replay_origin.go:41-47) — 셋 다 action · proposed_intent_id 만 읽음 |
| 스키마 트리거 | 운영 저널 사본 `sqlite_master` 에서 `tbl_name='exit_events'` 트리거 0. 코드의 `CREATE TRIGGER` 중 exit_events 대상 0. append-only 는 **주석 계약**(core_domain.go:31 · :244)이지 DB 강제가 아님 |

## 불일치와 결론

- 불일치 없음. CodeGraph callers 와 `rg` 전수가 같은 생산 리더 집합(3 + 스냅숏 안 읽는 7)을 줌.
- **결론: 엔진은 닫힌 포지션의 지난 exit_events 행에서 스냅숏 JSON 을 다시 읽지 않음.** 엔진 · execgw 의 리더는 action ·
  proposed_intent_id · decision_id 만 읽고, 스냅숏을 읽는 셋 중 `Journal.ExitEvents` 는 생산 호출자 0, 나머지 둘은 콘솔 읽기
  전용(`ReadOnly`). 엔진의 기동 · 복원 상태는 `exit_states` 한 행(`effective_snapshot_json`)에서 옴.
- 콘솔이 스냅숏을 **화면에 쓰는** 곳은 orders 의 연결된 1 행뿐이고(claims C5), 운영 사본에서 연결 134 건 전부가 ambiguous
  (무장 행 + PROPOSAL_FILLED 훅 행 2 행)라 지금은 스냅숏이 표시되는 주문이 0 임(`a206_reader_check.out.txt`).
