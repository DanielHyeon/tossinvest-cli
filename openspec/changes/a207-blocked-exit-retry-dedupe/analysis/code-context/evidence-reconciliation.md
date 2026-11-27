# a207 증거 조정 — 막힌 청산 재시도의 쓰기 · 읽기 · 알림 지점

- 기준 commit: `e9266bfb6`(origin/main), 2026-11-27. 질의는 같은 소스 트리를 가진 a206 worktree(`~/tossos-work/a206`, 그 브랜치의 유일한
  추가 커밋 `b18c85da2` 는 문서만)의 CodeGraph 색인으로 함 — a207 worktree 에 같은 색인을 다시 만들지 않음(57,792 nodes 재색인 생략).
- CodeGraph **1.6.0**. 시험 영향은 `codegraph affected … --filter '*_test.go'`(기본 판별식은 Go 시험을 못 찾음 — .claude/CLAUDE.md 하네스 주석).
- CodeGraphContext: **not-applicable** — 질문이 「어느 함수가 이 행을 쓰고 · 읽고 · 알리는가」라는 닫힌 열거이고, CodeGraph callers +
  `rg` 전수 + a206 의 리더 열거(`a206-exit-events-snapshot-shrink/analysis/code-context/evidence-reconciliation.md`, 같은 HEAD)가 정답 집합을 줌.
  2단계 구현 change 가 보조 문맥 후보 확장을 맡음.

## 질의와 결과

| 질의 | 결과 |
|---|---|
| `codegraph callers RecordExitJudgementResult` | 생산 1: `ExitObserver.record`(exitloop.go:1332). 나머지 `RecordExitJudgement`(같은 파일 래퍼) · 시험 1 |
| `codegraph callers recordExitJudgementTx` | `RecordExitJudgementResult` 1 — 판정 증거 행을 쓰는 트랜잭션은 이것 하나 |
| `codegraph callers appendExitEventTx` | 5: `recordExitJudgementTx`(판정 · 무장 행) · `clearExitProposalTx`(해제 훅 행) · `OpenExitState`(OPENED) · `resetExitStateForReadoptTx`(재편입) · `closeExitStateOnAdjustmentTx`(조정 닫힘). 반복 경로에 드는 것은 앞의 둘 |
| `codegraph callers clearExitProposalTx` | 2: `releaseExitProposal`(해제 판정) · `ResolveExitProposal`(체결 훅 — PROPOSAL_FILLED) |
| `codegraph callers ReleaseUnacceptedExitProposal` | 생산 3: `ExitObserver.release`(submit 의 해제 5 곳이 공유) · `noteHeldProposal`(판정 진입의 입증된 비수용 해제) · `attempt_thaw_command.go:42`(운영자) |
| `codegraph callers alertProposalRefused` | 생산 1: `submit`. 시험 1(`TestA197TheRefusalBodyNamesWhoRefused`) |
| `codegraph callers hydrateExitEventEvidence` | `AccountExitEvents` · `BrokerOrderExitLinks` · `scanExitEvent`(→ `Journal.ExitEvents`, 생산 호출자 0 — a206 재확인) |
| `codegraph callers HasExitEventForIntent` / `ExitProposalForIntent` | 생산 1: `Gateway.replayOrigin`(execgw/a166_replay_origin.go:14) — 재생 우선순위가 intent 당 exit_events 행의 action 에 기댐(claims C9) |
| `codegraph affected exit_state.go exit_snapshot.go exitloop.go --filter '*_test.go'` | 시험 파일 1,538(파일 단위 의존이라 넓음). 2단계 선택 시험은 design 「시험 계획」의 이름 목록으로 좁힘 |
| `rg -n "effective_source\|EffectiveSource" -g '!*_test.go'` | exit_events 의 effective_source 를 읽는 생산 코드는 journal 리더 셋(account_views.go:317 · :458, exit_state.go:879)과 열 존재 검사(readonly.go:92)뿐. operatorview · httpapi · console 의 같은 이름 필드는 exit_states 투영(「persisted effective snapshot」 고정 문구) |
| `rg -n "idx_exit_events"` | `idx_exit_events_position ON exit_events(position_id, id)`(core_domain.go:263) · `idx_exit_events_proposed_intent`(account_views.go:112) · 부분 유일 `idx_exit_events_decision … WHERE decision_id IS NOT NULL`(exit_snapshot.go:61) |
| `rg -n "func (o \*ExitObserver) (alertProposalRefused\|noteDelay\|alertRefused)"` + `rg "Notify(ctx"` | 엔진 exit 알림은 `o.alert` → `opts.Alerts.Notify` 하나로 모이고, 운영 조립은 `obs.RecordOnly`(기록 전용 — record_only.go:45) |
| 알림 정본 | engine-safety 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다」(spec.md:964-1000 — 영구 억제 SHALL NOT, 관측마다 구조화 로그 한 줄 SHALL) · :294(exit goroutine 의 기록이 창 지난 정착 행을 재무장 SHALL) |

## 불일치와 결론

- 불일치 없음. CodeGraph callers 와 `rg` 가 같은 쓰기 지점(판정 행 1 · 해제 훅 행 1)과 같은 리더 집합을 줌.
- **결론 1**: 막힌 청산 재시도 한 번은 exit_events 에 **판정 · 무장 행 1(세 JSON) + 해제 훅 행 1** 을 쓰고, 브로커에 닿은 거절이면
  주문 원장(intent · decision · attempt · 전이)도 씀. 줄일 수 있는 것은 판정 · 무장 행의 증거 바이트뿐임 — 행 자체는 재생 우선순위(C9)와
  중복 판정 가드(C8)가 읽음.
- **결론 2**: 거절 알림의 push 는 이미 key 당 재알림 창(1 시간)으로 묶여 있음(C11). 매 재시도마다 남는 것은 구조화 로그 한 줄과
  outbox 조회 한 번이며, 둘 다 정본 SHALL 임. 「막힘 시작 1 회 · 해소 1 회」 중 지금 없는 것은 **해소 알림**뿐임.
