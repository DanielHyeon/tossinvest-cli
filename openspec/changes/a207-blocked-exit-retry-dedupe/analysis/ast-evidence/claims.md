# a207 분기 주장의 AST 근거

proposal · design 이 근거로 쓰는 기존 함수 내부 분기 · early return · 호출은 이 디렉터리의 `tools/logic-map` 산출물로 열거함
(.claude/CLAUDE.md 필수 진입 5). 1단계는 기존 함수를 편집하지 않으므로 FLM 번들이 아니라 주장의 근거로 이 자리에 둠
(선례 a206 · a203 `analysis/ast-evidence/`). 2단계에서 고치는 함수는 `analysis/function-logic/` 번들을 따로 만듦.

- 기준: worktree `~/tossos-work/a207` HEAD `e9266bfb6`(origin/main), 2026-11-27.
- 생성: `GOFLAGS=-trimpath go build -o <scratch>/logicmap ./tools/logic-map` 뒤 `logicmap --file <파일> --func <함수> > <패키지>--<함수>.ast.json`.
- 줄 번호 ↔ 원문 대조는 산출물의 `branches[].at.line` · `returns[].at.line` · `calls[].at.line` 을 같은 HEAD 원문 줄과 맞춰 읽음.
  분기의 조건식 문구는 그 원문 줄을 인용한 것이고(logic-map 은 조건식을 싣지 않음), 분기 · return · 호출의 **존재와 위치**가 AST 근거임.

| 산출물 | 파일:줄 | source_sha256(앞 12) | 분기 | return |
|---|---|---|---|---|
| `engine--ExitObserver.submit` | internal/app/engine/exitloop.go:1513-1638 | 7626cf237058 | 18 | 10 |
| `engine--ExitObserver.record` | exitloop.go:1332-1473 | 7626cf237058 | 19 | 6 |
| `engine--ExitObserver.release` | exitloop.go:1645-1650 | 7626cf237058 | 1 | 2 |
| `engine--ExitObserver.judgeLadder` | exitloop.go:1123-1190 | 7626cf237058 | 8 | 7 |
| `engine--ExitObserver.judgeRatchet` | exitloop.go:1069-1121 | 7626cf237058 | 6 | 6 |
| `engine--ExitObserver.requiresFullJudgement` | exitloop.go:1195-1200 | 7626cf237058 | 1 | 2 |
| `engine--ExitObserver.alertProposalRefused` | exitloop.go:1931-1956 | 7626cf237058 | 3 | 0 |
| `engine--ExitObserver.alertRefused` | exitloop.go:1899-1921 | 7626cf237058 | 2 | 1 |
| `engine--ExitObserver.noteDelay` | exitloop.go:1960-1984 | 7626cf237058 | 2 | 2 |
| `engine--ExitObserver.noteExitOrder` | internal/app/engine/exit_order_log.go:198-231 | 6bedc542611a | 9 | 3 |
| `engine--ExitObserver.noteHeldProposal` | internal/app/engine/exit_held_proposal.go:126-173 | bc7a564eb8b1 | 11 | 6 |
| `journal--Journal.recordExitJudgementTx` | internal/journal/exit_state.go:396-654 | ea9f939d35d3 | 59 | 35 |
| `journal--appendExitEventTx` | exit_state.go:792-851 | ea9f939d35d3 | 7 | 8 |
| `journal--clearExitProposalTx` | internal/journal/apply_hook.go:868-908 | b39d23a138a0 | 9 | 8 |
| `journal--armExitProposalTx` | apply_hook.go:665-687 | b39d23a138a0 | 4 | 5 |
| `journal--Journal.releaseExitProposal` | internal/journal/exit_proposal_release.go:96-131 | 6f3fe938b9df | 8 | 9 |
| `journal--hydrateExitEventEvidence` | internal/journal/exit_snapshot.go:392-417 | 19ebe0eeb987 | 6 | 7 |
| `journal--Journal.recordAlertTx` | internal/journal/outbox.go:276-364 | ad74f6c9b295 | 8 | 7 |
| `journal--claimOwed` | outbox.go:372-418 | ad74f6c9b295 | 8 | 7 |
| `obs--RecordOnly.Notify` | internal/obs/record_only.go:45-61 | 465f09466857 | 3 | 4 |
| `execgw--Gateway.replayOrigin` | internal/execgw/a166_replay_origin.go:14-59 | 991f237a6004 | 15 | 13 |

## 주장

### 반복이 생기는 경로(해제 · 재무장)

- **C1 — submit 의 해제 호출은 5 곳이고, 각 해제 뒤 대기 · 횟수 상한 없이 return 함.** `ExitObserver.submit` 의 `o.release` 호출
  :1530(B2 `isZeroQuantity` — 하한 0주) · :1549(B3 `IssueReduction` 오류 — Guardian 거절) · :1566(B5 `sellIntent` 오류 — intent 무효) ·
  :1607(B12 `out.Reason == ReasonSymbolInFlight` → `ProposalCancelled`) · :1634(B15 default — 브로커 확정 거절 · 그 밖의 비수용 → `ProposalRefused`).
  각각 바로 뒤 return(:1532 · :1551 · :1568 · :1609 · :1636). 해제가 **없는** 갈래는 B1 하한 산술 실패(:1522) · B4 attach 실패(:1560) ·
  B9 확정 수용(:1594) · B10 결과 미상(:1604) · B13 attempt 는 있는데 입증된 비수용이 아님(:1618) — 이 다섯은 발의가 무장된 채 남음.
- **C2 — 사유와 갈래의 대응.** `opposite_pending_order_exists` · `dispatch_rejected_by_broker` 는 Place 가 돌아온 뒤의 확정 거절이라 B15
  default(:1620-1636)로 감: `alertProposalRefused`(:1625) → a197 대기열(B17 :1627, 보호 손절 ∧ FailedConfirmed ∧ 반대 미체결일 때만) →
  `release(ProposalRefused)`(:1634). 2026-08 의 `<no attempt>` + `PROPOSAL_CANCELLED` 는 B12 SymbolInFlight(:1605-1609, `noteDelay` 후
  `ProposalCancelled` 해제) 모양과 같음. Place 앞에서 끝나는 갈래(B1 · B2 · B3 · B4 · B5)는 attempt 를 만들지 않고, 그중 해제하는 것은
  B2 · B3 · B5(전부 `PROPOSAL_REFUSED`)임. Place 뒤의 B12 · B15 는 게이트웨이 결과에 따라 attempt 유무가 갈림 — 운영 데이터에서 CANCELLED 로 풀린 무장
  12,996 가운데 12,995 는 attempt 가 없었고 1 은 마지막 PLACE 가 `in_doubt_unresolved` 였음(measure.out). 해제 훅 행은 사유를 싣지 않음(apply_hook.go:832-834 주석 「The reason for the refusal is not recorded
  here」) — attempt 가 없는 거절의 구체 사유(Guardian · 하한 0 주 · intent 무효)는 원장에 없고 로그 · 알림에만 있음. 이 사실은 압축 전후가 같음
  (전체 증거 행에도 사유 열은 없음).
- **C3 — release 는 원장 판정 하나로 감.** `ExitObserver.release` B1(:1646) 이 `ReleaseUnacceptedExitProposal` 하나를 부르고 오류만
  감쌈. 원장 쪽 `Journal.releaseExitProposal` 은 `exitIntentAttemptsQ`(:113) → B5 `!allows(...)`(:117) 면 쓰기 없이 return(:118) →
  `clearExitProposalTx`(:120) → B7 `!released`(:124) 면 return(:125) → commit(:127). 그 밖의 해제 호출자는 CodeGraph 기준
  `noteHeldProposal`(exit_held_proposal.go:141, B4 `Verdict == ExitIntentUnaccepted`) · `attempt_thaw_command.go:42`(운영자) · 청소의
  `ReleaseClearedExitProposal`.
- **C4 — 해제는 훅 행 하나를 씀(증거 열 없음).** `clearExitProposalTx` B3(:881 발의 없음) · B4(:884 다른 intent) 면 쓰기 없이 return,
  아니면 pending 열을 비우고(:888) 사다리면 rung 을 되돌린 뒤(B6 :894 · B7 :895 · `rollBackRungTx` :896) `appendExitEventTx` 를
  `Evaluation` 없이 부름(:901) — `appendExitEventTx` B2(:801)가 거짓이라 증거 20 열이 전부 NULL 인 훅 행이 됨.
- **C5 — 재무장은 다음 관측의 전체 판정이 함.** `judgeLadder` 는 B8(:1185) `!requiresFullJudgement` 면 `refreshObservation`(:1186,
  행 없음)으로 가고 아니면 `record`(:1188). ratchet 도 같음: `judgeRatchet` B6(:1114) → `refreshObservation`(:1115) 아니면 `record`(:1118). `requiresFullJudgement` 는 B1(:1196) 이 아니면 `snapshot.Changed || !SameExitOperationalLine`
  (:1199). 해제가 pending 을 비우고(사다리는 rung 도 되돌림) 스냅숏이 다시 orderable 해지므로 `ChangedFromState`(:1175)의 `Orderable`
  항으로 Changed 가 참이 됨(손 읽기 — exitpolicy/snapshot.go:241-251 의 단일 대입식). `record` 는 B14(:1432) orderable 이면 발의를 싣고
  `RecordExitJudgementResult`(:1445) → B19(:1466) 가 무장(`ExitArmArmed`)이 아니면 return(:1467), 무장이면 `o.submit`(:1471).
- **C6 — 재무장 경로에 대기 · 백오프 분기는 없음.** `record` 의 return 여섯(:1336 · :1415 · :1451 · :1464 · :1467 · :1471) 중 제출을
  건너뛰는 것은 견적 불가(B1) · 청소 오류(B11) · `ErrProposalPending`(B17 :1447) · 기록 오류 · 무장 아님(B19) 뿐이고, 시간이나 횟수를
  보는 갈래는 없음. `submit` 도 같음(C1).

### exit_events 쓰기 지점

- **C7 — 판정 하나는 증거 행 하나를 씀.** `recordExitJudgementTx` B48(:611) 발의가 있으면 `armExitProposalTx`(:613) 로 무장하고,
  B51(:627 `recomputed != nil && effective != nil`) 이면 `evaluationForEvent`(:628) 로 스냅숏을 싣고, 조건 없이 `appendExitEventTx`
  (:630) 를 부름. `appendExitEventTx` 는 B2(:801) 안에서 saved · recomputed · effective 를 따로 `encode`(:809 · :812 · :815)해 한 INSERT
  (:830)로 씀. `encode` 는 스냅숏이 없으면 NULL(B3 :803-804)이라 saved 가 없는 첫 평가는 둘만, legacy 판정(평가 없음 — B51 거짓)은 0 개를 씀.
  즉 증거가 있는 무장 행은 **재무장마다** JSON 둘~셋을 새로 씀.
- **C8 — 중복 판정 가드는 decision_id 에 걸려 있음.** 같은 함수 B26(:492 `recomputed != nil`) 안의 질의(:494 `SELECT 1 FROM exit_events
  WHERE decision_id=?`) → B27(:496) 찾으면 `ErrProposalPending`(:501). DB 쪽 짝은 부분 유일 인덱스 `idx_exit_events_decision ON
  exit_events(decision_id) WHERE decision_id IS NOT NULL`(exit_snapshot.go:61-62, DDL 손 읽기). decision id 는 snapshot id 에서, snapshot id 는
  관측 id(가격 · fetched_at 포함)와 정책 · 평가 입력(baseline · high water · rung · pending 등)의 지문에서 유도됨(exitpolicy/snapshot.go:125-129
  입력 지문 · :253-256 `finishIDs`, exitloop.go:1246-1270 `stableObservationID` 손 읽기). 따라서 같은 견적이 다시 쓰여도 decision 이 늘 같지는
  않으나, **같은 decision 이면** 이 가드가 재무장 · 제출을 막음. 그래서 압축 행도 decision_id 를 남겨야 함(design D3).
- **C9 — 무장 행은 intent 마다 필요함.** `Gateway.replayOrigin` B11(:46 엔진 매도) 은 `ExitProposalForIntent`(:47) 가 돌려준 action 으로
  B14(:52 손절 → `Critical`) · B15(:54 익절 → `Normal`) 를 가르고, B12(:48) 못 찾으면 `Normal, false`(:49) — 「출처 미입증」임.
  예산이 걸린 재생 호출자는 미입증이면 재생을 거절하고 관측에 넘김(execgw/replay.go:309-318 `ReasonReplayIneligible`, 손 읽기).
  `ExitProposalForIntent` 는 `exit_events.proposed_intent_id = intent.ID` 행의 action 을 읽음(a166_replay_origin.go:17-19 SQL 손 읽기).
  따라서 재시도의 무장 행을 **아예 쓰지 않는** 설계는 결과 미상이 된 그 손절의 재생을 막음 → 버림(design 안 D-0).

### 알림 지점

- **C10 — 거절 알림은 매 재시도마다 기록 입구를 지남.** `submit` B3(:1548) · B15(:1625) 가 `alertProposalRefused` 를 부르고, 그 함수는
  분기(B1~B3 :1935-1939)로 본문만 고른 뒤 조건 없이 `o.alert`(:1942) 를 부름 — key 는 `exit.proposal_refused|포지션|action|level`(:1944-1945).
  엔진의 알림 부품 `RecordOnly.Notify` 는 B1(:47 알림기 없음) 이 아니면 `n.logEvent`(:51 구조화 로그 한 줄)를 먼저 하고, critical 이면
  `recordCritical`(:60).
- **C11 — 전송은 key 당 재알림 창(기본 1 시간)에 한 번.** `recordAlertTx` B3(:293 기존 행) → `claimOwed`(:294) → B4(:295 rearm) 일 때만
  UPDATE(:334). `claimOwed` 는 B2(:379 PENDING → 전송 필요) · B3(:382 DELIVERED/ACKNOWLEDGED) 안에서 B4(:383 창 0 → 재무장 없음) ·
  B5(:387 시각 없음 → 재무장) · B6(:393 음수 경과 → 재무장) · B7(:407 창 안 → 전송 없음, :408) · 창 지남 → 재무장(:410).
  창의 기본값은 `DefaultRemindAfter = time.Hour`(notifier.go:60, 손 읽기).
- **C12 — 래치 선례 셋.** `noteExitOrder` B2(:202 거절 줄) → B5(:209 같은 포지션의 직전 거절 서명과 같으면 return :210) — 연속 주기의 같은
  거절 줄은 한 번만(a074). `alertRefused` B1(:1900 이미 알림) → return(:1901). `noteDelay` B1(:1963 첫 관측이면 시각만) · B2(:1967 한도 전이거나
  이미 알림) → return. 즉 「같은 조건은 한 번」 래치는 판정 거부 · 지연 · 주문 줄에는 있고 **거절 알림(C10)에는 없음** — 거절 알림의
  중복 제거는 outbox key 와 재알림 창(C11)만 함.

## 손 읽기로 남긴 것(AST 아님, 표시)

- SQL 문자열 안의 조건(C8 의 유일 인덱스 DDL, C9 의 `proposed_intent_id` 질의, `alert_outbox` DDL).
- `ChangedFromState` · `finishIDs` · `stableObservationID` 의 단일 대입식 · 해시 입력 목록(C5 · C8).
- `DefaultRemindAfter` 상수(C11).
