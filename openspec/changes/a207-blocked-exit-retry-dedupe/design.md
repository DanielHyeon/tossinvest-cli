# a207 design — 막힌 청산 재시도의 반복 기록 · 알림 정리

> 기준: origin/main `e9266bfb6`, 2026-11-27. 수치는 운영 저널의 읽기 전용 사본에서 집계로만 잼
> (`analysis/a207_measure.py` → `a207_measure.out.txt`, 질의 비용 `analysis/a207_latency.py` → `a207_latency.out.txt`). 사본은 실행 끝에 지움.
> 분기 주장은 `analysis/ast-evidence/claims.md`(C1~C12, logic-map AST 21 개)를 근거로 함. 쓰기 · 읽기 · 알림 지점 열거는
> `analysis/code-context/evidence-reconciliation.md` 에 있음. 선행 a206 의 리더 열거(같은 HEAD)를 재사용함. 경량 리뷰 r1 의 수리 내역은 `review.md` 에 있음.

## 1. 사실(F)

| # | 사실 | 근거 |
|---|---|---|
| F1 | 무장 행 26,176 개 중 해제(PROPOSAL_REFUSED · PROPOSAL_CANCELLED 훅 있음)된 것 26,037 | measure.out |
| F2 | 해제된 무장 행의 월 · 사유(마지막 PLACE attempt): 2026-08 `<no attempt>`+CANCELLED 12,995 · `<no attempt>`+REFUSED 16 · `dispatch_rejected_by_broker` 3,119 · in_doubt 1 / 2026-10 dispatch_rejected 3,325 · `opposite_pending_order_exists` 6,494 / 2026-11 dispatch_rejected 1(11-25) · opposite_pending 86(11-16 10 · 11-25 64 · 11-27 12) | measure.out |
| F3 | **D1 · D2 술어 그대로의 압축 대상 25,946 행, 그 JSON 168,121,152 B**(운영 사본 JSON 전체 183.2 MB 의 91.8%). 그 가운데 그 무장 자신은 해제되지 않은 것 8(수용 확정 6 · 결과 미상 2) | measure.out(r1 수리 뒤) |
| F4 | 「막힘 에피소드」(키와 무관하게 연속한 해제된 무장) 11 개: 길이 1 · 4 · 5 · 8 · 10 · 13 · 64 · 288 · 2,614 · 9,531 · 13,499. 에피소드 평균 무장 간격 5.2~132.8 s(5 s 주기 시절). 에피소드 뒤 첫 비해제 무장의 사유: 수용 6 · 결과 미상 2. 다음 무장 없음 3 | measure.out |
| F5 | 키(세대 · action · 수준 · 사유) 구간 1,037 개, 길이 p50 2 · p90 11 · p99 60 · 최대 9,495. 구간 끝은 대부분 키 변경(1,026) — 사다리 해제가 rung 을 되돌리며(C4) 한 에피소드 안에서 action · 수준이 바뀜 | measure.out, C4 |
| F6 | 행 바이트(열 길이 합, 평균): 무장 행 JSON 6,478 · 평탄화 증거 376 · 기본 열 157, 해제 훅 행 141. effective JSON 평균 2,159(a206 F7) | measure.out |
| F7 | 주문 원장 bytes/row: intents 242 · decisions 514 · mutation_attempts 915 · attempt_transitions 166, 청산 PLACE attempt 당 전이 3.01 | measure.out |
| F8 | 거절 알림 outbox 행 13(10 포지션, 포지션당 key 최대 3), 전부 DELIVERED. engine.log 의 `exit.proposal_refused` 줄 9,906. 로그가 2026-10-21 부터만 남아 있음 | measure.out, `grep -c`(원문 출력 없음) |
| F9 | 압축 판정 질의 둘(직전 무장 행 · 해제 훅 존재)의 비용: 가장 큰 포지션(27,072 행)에서 p50 15.6 µs · p99 19.9 µs · max 134.1 µs. 두 질의 모두 인덱스 검색(`idx_exit_events_position` · `idx_exit_events_proposed_intent`). **단독 사본에서 잰 값이며 쓰기 잠금 경합은 재지 않음** | latency.out |
| F10 | 관측 주기는 main 에서 2 s(exitloop.go:101, a205). 10-29 에피소드(9,531 회 · 50,215 s)를 2 s 로 옮기면 상한은 ≈25,100 회(2 s 마다 한 번). 실제 간격은 a205 배포 뒤 측정 | exitloop.go, F4 |

## 2. 반복 경로(AST 요약)

| 사유 | 해제 지점(submit) | 해제 훅 | attempt 행 | 거절 알림 |
|---|---|---|---|---|
| 반대 미체결 `opposite_pending_order_exists` | B15 default :1634 | REFUSED | 있음(FailedConfirmed) | :1625 + a197 대기열(B17, 보호 손절만) |
| 브로커 확정 거절 `dispatch_rejected_by_broker` | B15 default :1634 | REFUSED | 있음 | :1625 |
| 다른 변경이 진행 중 `SymbolInFlight`(8월 라이브락형) | B12 :1607 | CANCELLED | 운영 데이터상 없음 | 없음 — `noteDelay` 30 s 경보(래치) |
| Guardian 거절 | B3 :1549 | REFUSED | 없음 | :1548 |
| 하한 0 주 | B2 :1530 | REFUSED | 없음 | 없음(주문 줄만) |
| intent 무효 | B5 :1566 | REFUSED | 없음 | `alertRefused`(래치) |

재무장 경로: 해제가 pending 을 비우고, 사다리면 rung 도 되돌림(C4). 다음 관측에서 `judgeLadder` B8(ratchet 은 `judgeRatchet` B6) →
`record` → `RecordExitJudgementResult` → 무장 → `submit` 순서로 감(C5). 그 사이에 시간이나 횟수를 보는 분기는 없음(C6).
쓰기 지점은 둘임: 판정 트랜잭션의 `appendExitEventTx`(C7, JSON 둘~셋)와 해제 트랜잭션의 훅 행(C4, 사유 없음).
읽기 지점도 둘임: 재생 출처 입증이 intent 마다 무장 행의 action 을 읽고(C9), 중복 판정 가드가 decision_id 를 읽음(C8).

## 3. 선택지

### 3.1 저장 — 재시도 한 번의 exit_events 바이트(열 길이 합, 페이지 · 인덱스 오버헤드 별도)

| 안 | 무엇 | 무장 + 훅 | 2 s 최악 1 일(43,200 회) | 스키마 | 리더 | 롤백 이미지(v36)에서 |
|---|---|---|---|---|---|---|
| 현행 | 매 무장 JSON 둘~셋 | 7,152 B | ≈309 MB | — | — | — |
| A | effective=recomputed 이면 effective 생략, 리더가 채움(a206 B) | 4,993 B | ≈216 MB | 없음 | `full()` 완화 | 그 행이 손상으로 읽힘 |
| **B2(권장)** | 직전 무장이 같은 키이고 해제됐으면 JSON 셋 생략 · 평탄화 증거(decision_id 포함) 유지 · effective_source 표지 | **674 B** | **≈29 MB** | 없음 | 모양 하나 추가 | 그 행이 손상으로 읽힘(§6, 엔진 무영향) |
| B1 | 같은 조건에서 증거 20 열 전부 NULL(legacy 모양) | 298 B | ≈13 MB | 없음 | 무변경 | 오류 없음 |
| C | 내용 주소 스냅숏 표(`exit_snapshot_blobs`) + 참조 열 | ≈3.0 KB | ≈130 MB | **v37** | 변경 | **DB 를 거부**(ErrSchemaTooNew) → 백업 복원 필요 |
| D-0 | 재시도 무장 행을 쓰지 않고 카운터만 | ≈0 | — | 표 추가 | — | 버림: 그 손절이 결과 미상이 되면 재생 출처를 입증 못 해 재생이 거절됨(C9) |

- 주문 원장은 어느 안에서도 그대로임. 브로커에 닿은 재시도 한 번이 ≈2.2~2.7 KB(F7, decision 1~2 행)이고, 2 s 최악이면 ≈94~116 MB/일임.
  실제 주문 시도의 기록이라 줄이지 않음(남은 위험 R1).
- 역사 재생: B2 를 8월부터 적용했다면 25,946 행(F3)의 JSON 168.1 MB 를 쓰지 않았음.
- **B1 을 권장하지 않는 이유**: decision_id 가 NULL 이 되면 부분 유일 인덱스(`WHERE decision_id IS NOT NULL`)가 그 행을 보지 않음.
  판정 트랜잭션의 중복 가드(:494)도 그 행을 못 찾음(C8). 그러면 같은 decision 의 판정이 같은 intent id(`exit_` + decision 해시,
  exitloop.go:1475-1481 `exitIntentID`)로 다시 무장됨. 지금은 `ErrProposalPending` 으로 멈추던 그 주기가 이미 쓴 intent id 로 Attach · Place 까지
  가서 게이트웨이의 intent 중복 처리에 맡겨짐. 그 처리는 이 단계에서 확인하지 않았음 — 동작이 바뀐다는 것만으로 버리기에 충분함.
  가드를 `proposed_intent_id` 질의로 옮기면 막을 수는 있으나, DB 수준 유일성을 잃고 「새 가드 발명」이 됨.
- **A 를 권장하지 않는 이유**: 절감이 1/3 뿐이라 2 s 최악에서도 ≈216 MB/일이 남음.
- **C 를 버리는 이유**: 롤백 이미지가 v37 저널을 열지 못함(schema.go:19-20 「The rollback is "run the previous binary", which refuses the newer
  user_version (ErrSchemaTooNew)」). 그러면 롤백이 엔진 정지(= 손절 없음) 중 백업 복원이 되어, 위험이 저장 절감보다 큼.

### 3.2 B2 의 손실과 대가

- **압축 행에서 잃는 것**: 그 재시도 판정의 스냅숏 JSON 셋.
- **압축 행에 남는 것**:
  - 평탄화 증거: 관측 id · 관측 시각 · snapshot id · decision id · 정책 id/버전/digest · 투영 수량 · 다음 목표/보호 · 비율 · effective_source(표지)
  - 기본 열: 관측가 · high water · baseline · 수준 · action · intent · 시각
  - 해제 훅 행과 주문 원장
  - 같은 키 구간의 첫 행이 같은 정책 · 같은 수준의 전체 스냅숏을 가짐.
- **거절 사유는 압축과 무관함**: 전체 증거 행에도 사유 열은 없음. 해제 훅도 사유를 싣지 않음(C2, apply_hook.go:832-834).
  사유는 attempt 가 있는 거절(반대 미체결 · 브로커 확정 거절)에서만 `mutation_attempts.reason_code` 로 남음. attempt 가 없는 거절의 구체 사유는
  지금도 로그와 알림에만 있음. 이 change 는 그 상태를 바꾸지 않음.
- **수용된 재시도도 압축될 수 있음**: 압축은 무장 시점, 즉 제출 전에 정함. 그래서 막힘이 풀려 수용된 그 무장 행도 직전 무장이 같은 키이고
  해제됐다면 압축됨. 역사 재생으로 8 행임(수용 6 · 결과 미상 2, F3).
  - 콘솔 orders 는 intent 당 연결 행 수를 먼저 판정함. 무장 행과 체결 훅 행이 둘이면 hydrate 전에 `ambiguous_exit_evidence` 로 끝나므로
    (account_views.go:545-578), 지금은 확정 주문 전부가 ambiguous 이고 orders 가 보이는 스냅숏은 이미 0 건임(a206 F10).
  - 그 행의 증거를 전체로 남기려면 「해제 때 직전 행을 압축하는」 런타임 UPDATE 가 필요함. 이는 exit_events append-only 계약
    (core_domain.go:31 · :244 주석)과 a206 의 제안 요구(「기존 행은 운영자 정비 명령으로만 바뀐다」)를 깸 → 채택하지 않음.
- **손상 탐지가 약해짐**: 압축 행은 평탄화 열과 스냅숏의 교차 검사(`validateExitEventFlattenedEvidence`)를 받지 못함. 리더는 모양만 검사함
  (필수 평탄화 열 Valid · JSON 셋 NULL · arm suppression NULL · 표지 값). 남은 위험 R3.
- **a206 정비와 겹치면 첫 행이 비워질 수 있음**: a206 정비는 닫힌 포지션의 미연결 행을 비우므로, 그 포지션의 구간 첫 행(거절된 무장이라 미연결)이
  비워지고 브로커에 연결된 압축 수용 행만 남을 수 있음. 그러면 「증거는 구간의 첫 행」이 그 포지션에서는 성립하지 않음. 그때 남는 증거는 압축 행의
  평탄화 열 + 주문 원장임. 사람 결정 6.

### 3.3 알림

| 안 | 무엇 | 정본 충돌 |
|---|---|---|
| 현행 | 거절 critical: key=포지션·action·수준, 전달 뒤 1 시간 창(C11), 관측마다 로그 한 줄(C10). 해소 알림 없음 | — |
| **N1(권장)** | 현행 유지 + 막힘 해소 일반 알림 1 회(수용 확정 뒤, 주기 끝, 엔진 메모리의 셈만 씀) | 없음(일반 등급 사건 추가) |
| N2 | N1 + 같은 막힘 에피소드 동안 재알림 억제 | **있음**. engine-safety Requirement 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다」는 억제가 창이어야 하고 영구적이면 안 된다고 함(SHALL NOT). 근거는 「반복 전송이 알림 경로가 살아 있다는 유일한 주기적 증거」라는 것임. 에피소드는 48 시간까지 갔음(F4) |
| N3 | key 를 포지션 단위로(action · 수준 제거) | 정본은 key 가 「조건」을 담으라고 함. 익절 거절 창 안에서 손절 거절이 새 push 를 못 냄 → 버림 |

### 3.4 결론

- 저장은 **B2**, 알림은 **N1** 로 함.
- 사용자 목표 (1)은 B2 가 이룸.
- 사용자 목표 (2) 가운데 「해소 1 회」는 N1 이 더함. 「시작 1 회」는 key 마다 이미 그러함(F8: 포지션당 key 최대 3).
  「막힘 중 0 회」는 정본과 부딪치므로 사람 결정 3 으로 올림.

## 4. 결정(D) — 2단계 B2 · N1 의 설계

- **D1 압축 판정의 자리**: `recordExitJudgementTx` 안에서만 함. 대상은 무장(`judgement.Proposal != nil`, B48)이고, B51(recomputed · effective 있음)이며,
  effective_source 가 `recomputed` 인 경우뿐임. 같은 트랜잭션(BEGIN IMMEDIATE, 이미 쥔 쓰기 잠금) 안에서 질의 둘을 함.
  ① 이 포지션의 직전 무장 행: `proposed_intent_id` 가 있고 action 이 훅이 아닌 행을 `ORDER BY id DESC LIMIT 1` 로 읽음.
  ② 그 intent 의 해제 훅(PROPOSAL_REFUSED · PROPOSAL_CANCELLED)이 있는지 봄.
  판정은 원장에서만 함 — 엔진 메모리 래치가 없으므로 재시작 뒤에도 같은 답이 나옴.
- **D1b 질의 오류를 판정 오류로 돌려주지 않음**: ① · ② 가운데 하나라도 오류면 그 무장 행을 **전체 증거로** 쓰고 판정을 계속함(리뷰 r1 P1).
  한계(리뷰 r2 P1): SQLite 는 일부 오류(SQLITE_IOERR 등)에서 읽기 문장만으로도 트랜잭션을 되돌리고, ctx 취소도 `sql.Tx` 를 되돌림
  (modernc.org/sqlite v1.54.0). 그 경우에는 뒤의 INSERT · commit 이 지금처럼 실패하고 그 주기는 제출하지 않음. 이는 같은 트랜잭션의 기존 읽기
  (진행 상태 :439 · 수명주기 :476 · 중복 질의 :494 · `appendExitEventTx` 의 세대 조회 :794, 하한 판정이면 :451 · :456 둘 더)가 이미 가진 실패 부류와 같음 — 새 부류는 없고, 같은 부류의 기회가 문장
  둘만큼 늘어남.
- **D2 같은 조건**: ①이 있고, ②가 참이고, ①의 `lifecycle_generation` · `action` · `level_after` 가 이 무장의 값(현재 세대 · 발의 action ·
  `levelAfter(level, ActiveRung)`)과 같아야 함. 거절 사유는 키에 넣지 않음 — 스냅숏 내용이 사유와 무관하고, 사유의 원장 기록은 압축과 상관없음(§3.2).
- **D3 압축 행의 모양**: 평탄화 증거 17 열은 지금과 같은 값으로 씀. decision_id 도 포함하므로 유일 인덱스와 :494 가드가 그대로 섬. JSON 셋은 NULL 로 둠.
  effective_source 에는 새 상수(예: `recomputed_repeat`, 이름은 구현 때 정함)를 씀. arm_suppressed_reason 은 NULL(무장 행이라 원래 없음).
  `appendExitEventTx` 에 「JSON 생략」 갈래 하나를 더함. 그 밖의 행(판정 전용 · saved_monotone · legacy · 훅 · 구간 첫 행)은 바이트가 같음.
- **D4 리더**: `hydrateExitEventEvidence` 에 압축 모양 하나를 더함. effective_source 가 표지 값이면 다음을 요구함: 필수 평탄화 열 Valid ·
  JSON 셋 NULL · arm suppression NULL. 통과하면 세 뷰의 UnknownReason 을 새 사유(예: `repeat_of_refused_proposal`)로 채우고 `nil` 을 돌려줌.
  그 밖의 부분 모양은 지금처럼 `ErrExitSnapshotCorrupt` 임(a206 C2 유지).
  표시 범위: 이 사유가 보이는 곳은 계좌 이력(`AccountExitEvents`)과 `Journal.ExitEvents` 임. orders 연결은 지금처럼 intent 당 행 수 판정이
  먼저이고, 단일 연결일 때만 이 사유가 보임(§3.2). 문구는 사람 결정 2.
- **D5 막힘 셈(엔진 메모리)**: `ExitObserver.release` 가 원장의 `released=true` 를 받았을 때만, 포지션별로 {거절 수, 처음 · 마지막 시각, 해제 종류별 수}를
  메모리에 더함. 지금은 그 bool 을 버리고 있음(`_, _, err` — C3). 원장 읽기는 0 임. 엔진이 재시작하면 0 부터 다시 셈.
  지속되는 감사 기록은 행 자체임: 재시도마다 무장 행 · 해제 훅 · attempt 행이 남으므로, 횟수와 처음/마지막 시각은 아래 SQL 로 언제든 다시 셀 수 있음.
  ```sql
  -- 포지션 하나의 연속 해제 무장(막힘 에피소드): 무장 행 중 해제 훅이 있는 것을 id 순으로 묶음
  SELECT e.id, e.created_at, e.action, e.level_after, a.reason_code
    FROM exit_events e LEFT JOIN mutation_attempts a ON a.intent_id = e.proposed_intent_id AND a.kind = 'PLACE'
   WHERE e.position_id = ? AND e.proposed_intent_id IS NOT NULL
     AND e.action NOT IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED','PROPOSAL_FILLED','OPENED','COMPLETED','ADJUSTMENT_CLOSED')
   ORDER BY e.id;
  ```
- **D6 해소 알림**: `submit` B9(확정 수용) 갈래에서는 그 포지션의 셈이 1 이상일 때 셈을 떼어 주기 대기열에 넣기만 함(I/O 0, 그 포지션의 Place 뒤).
  `endExitOrderCycle`(주기 끝, 모든 판정 · 제출 뒤)이 일반 등급 사건 하나를 `RecordOnly.Notify` → `NormalRelay.Offer` 로 넘김.
  Offer 는 막히지 않음 — 버퍼가 차거나 실행자가 멈췄으면 버리고 그 사실을 로그에 남김(normal_relay.go:42-57). 한 해소에 한 번인 이유는
  셈을 떼어 내는 자리가 확정 갈래 하나뿐이기 때문임(일반 등급에는 outbox 중복 제거가 없음). 엔진 재시작 · 포지션 닫힘 · 결과 미상(B10)은
  알리지 않음(사람 결정 4). 거절 critical(C10 · C11)은 그대로 둠.
- **D7 토글 없음**: 다음 두 경계를 시험으로 고정함(T5 · T6).
  - 막힘이 없는 경로(①이 없음 · ②가 거짓 · 키가 다름 · 질의 오류)는 지금과 행 바이트가 같음.
  - 알림기가 없으면(`opts.Alerts == nil`) 해소 알림도 없음.
  토글을 원하면 사람 결정 5.

## 5. 즉시성이 바뀌지 않는다는 근거와 그 한계

1. 압축은 이벤트 행의 **증거 열 값**만 바꿈. 같은 트랜잭션에서 손대지 않는 것은 셋임: exit_states 갱신(:584-603), 무장(:611-613),
   결과 판정(B55 :640-653 의 ArmOutcome · ArmedProposal). 그래서 `record` B19(:1466) → `submit`(:1471) 은 지금과 같은 주기에 같은 값으로 돎.
2. 압축 판정 질의의 오류는 판정 오류가 되지 않고 전체 행 기록으로 떨어짐(D1b, T8 이 고정함). 단, SQLite 가 그 오류로 트랜잭션을 되돌린 경우는
   같은 트랜잭션의 기존 읽기와 같은 부류의 실패로 남음(D1b 한계 — 새 부류 없음, 기회는 문장 둘만큼 늘어남).
3. submit 의 해제 5 곳(C1), 해제 판정(C3), 재무장 경로(C5)에는 새 분기가 없음. 대기 · 백오프 · 상한을 더하지 않음.
   D5 는 release 안의 메모리 갱신 하나이고, D6 은 확정 갈래의 대기열 추가 하나임. 둘 다 **그 포지션의** Place 뒤(해제 뒤 · 확정 뒤)에 있음.
   관측 루프는 포지션을 차례로 돌므로, 같은 주기에서 뒤에 판정되는 포지션의 Place 는 이 메모리 갱신(map 연산)만큼 늦어짐(리뷰 r2 P2).
4. 더해지는 비용:
   - 판정 트랜잭션 안의 인덱스 질의 둘: 단독 사본에서 p99 19.9 µs(F9). 이미 쥔 쓰기 잠금을 그만큼 더 오래 쥠. 콘솔 읽기(WAL)는 쓰기와 서로
     막지 않음. 같은 프로세스의 다른 쓰기와의 경합은 재지 않았음 — 2단계에서 시험 안 계측으로 잼(T1).
   - 주기 끝 해소 알림 하나: `RecordOnly.Notify` 가 구조화 로그 한 줄을 동기로 쓰고(record_only.go:51), `NormalRelay.Offer` 가 뮤텍스 아래에서
     채널에 넣음(가득 차거나 멈췄으면 버림 로그 한 줄 — normal_relay.go:42-57). `Run` 은 `ObserveOnce`(지연 실행되는 `endExitOrderCycle` 포함)가
     끝난 뒤 잠들므로(exitloop.go:432-440), 이 시간만큼 다음 주기 시작이 늦어짐. 해소 하나에 한 번뿐이고 원장 읽기는 0 임(리뷰 r2 P2).
5. 중복 판정 가드(:494)와 유일 인덱스가 압축 행에도 섬(D3). 그래서 같은 decision 판정의 「제출 안 함」(`ErrProposalPending`)도 지금과 같음.
6. 재시도마다 무장 행이 남으므로 재생 출처 입증(C9 — 손절은 `Critical`)도 같음.

## 6. 스키마 · 롤백 · 순서

- 스키마 변경 없음(`SchemaVersion = 36` 유지). 롤백 이미지(schema 36)는 저널을 그대로 열고, 압축 행을 「부분 증거 → 손상」으로 읽음.
  - 콘솔 계좌 이력은 그 행을 `invalid_event_evidence` 로 표시하고 목록은 계속 보임(a206 C3).
  - orders 연결은 행 수 판정이 먼저라 지금처럼 대부분 ambiguous 임. 단일 연결이면서 압축 행인 경우에만 `invalid_event_evidence` 로 보임(a206 C4).
  - `Journal.ExitEvents` 는 오류를 내지만 생산 호출자가 0 임.
  - 엔진은 exit_events 의 JSON 을 읽지 않으므로(a206 결론) 롤백 중 엔진 동작은 같음. 롤백 뒤 새로 쓰는 행은 다시 전체 증거임.
- a206 과의 관계:
  - a206 정비(닫힌 · 미연결 행의 증거 20 열 NULL)는 압축 행도 같은 방식으로 비움 → SQL 수준 충돌은 없음.
  - 다만 §3.2 의 「구간 첫 행이 비워짐」 문제가 있음(사람 결정 6).
  - a206 design 의 B 는 이 change 로 대체됨.
  - 두 브랜치가 `_registry.yaml` · `FEAT-TOS-009.yaml` 의 같은 목록 끝에 한 줄씩 더하므로, 늦게 병합하는 쪽이 목록을 맞춤.

## 7. 시험 계획(2단계)

### 7.1 RED(구현 전에 실패해야 함)

| # | 시험(이름은 구현 때 확정) | 못 박는 것 |
|---|---|---|
| T1 | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle`(engine, 실 journal + fake 브로커: 반대 미체결로 N=5 회 거절 → 6 번째 주기 수용) | **즉시성 회귀**: 주기 1~6 모두 Place 가 정확히 1 번 일어나고, 6 번째 주기의 Place 가 막힘이 풀린 첫 관측에서 남. 주기 사이 시계 진행은 관측 주기 하나뿐. 압축 전후의 주기당 판정 트랜잭션 시간도 같이 기록함(F9 의 경합 미측정 보강) |
| T2 | journal: 같은 키 반복 무장은 압축됨. 구간 첫 행, 직전이 미해제(수용 · 결과 미상)인 뒤, 세대 · action · 수준이 다른 경우는 전체 증거 | D1 · D2 술어의 각 항 |
| T3 | journal: 압축 행을 `Journal.ExitEvents` · `AccountExitEvents` 가 오류 없이 새 사유로 읽음. `BrokerOrderExitLinks` 는 단일 연결 압축 행에서 새 사유, 다중 연결이면 지금처럼 ambiguous. 표지 없는 부분 모양은 여전히 손상 | D4, a206 C2 유지 |
| T4 | journal: 압축 행과 같은 decision 의 판정이 `ErrProposalPending`. 기존 `TestExitSnapshotDuplicateDecisionIsNotRearmed` 를 압축 행 위에서 한 번 더 돌림 | D3 · §5.5 |
| T5 | journal: 막힘 없는 무장 · 판정 전용 · saved_monotone · legacy 행의 INSERT 값이 현행과 같음(골든) | D7 경계 |
| T6 | engine: 해소 알림 — 거절 N 회 뒤 수용이면 1 회, 거절 없는 수용이면 0, 같은 해소는 다시 안 냄, Place 뒤 주기 끝에 냄, `Alerts == nil` 이면 0 | D5 · D6 · D7 |
| T7 | execgw: 압축 무장 행의 intent 재생 출처가 손절이면 `Critical, true` | C9 · §5.6 |
| T8 | journal: 압축 판정 질의 ① · ② 에 오류를 주입하면(기존 `runExitWriteHook` 같은 시험 훅) 전체 증거 행이 기록되고 판정 결과가 무장 | D1b · §5.2 |

기존 회귀 앵커:
- `TestARefusalReArmsTheLevel`(journal/exit_state_test.go:375)
- `TestARefusedProposalIsReArmedAfterARestart`(:790)
- `TestARefusedProposalReleasesTheLevelAndAlerts`(engine/exitloop_test.go:1031)
- `TestExitSnapshotDuplicateDecisionIsNotRearmed`(journal/exit_snapshot_persistence_test.go:251)
- a197 · a094 의 거절 시험

### 7.2 변이 원장(High-risk — 비례 원칙상 의무)

| # | 변이 | 잡아야 할 시험 |
|---|---|---|
| M1 | ② 해제 확인 제거(직전이 수용이어도 압축) | T2 |
| M2 | ① 의 action 비교 제거 | T2 |
| M3 | ① 의 level_after 비교 제거 | T2 |
| M4 | ① 의 lifecycle_generation 비교 제거 | T2 |
| M5 | 압축 행에서 decision_id 도 NULL | T4(+ T3 모양 검사) |
| M6 | 압축을 판정 전용 행 · saved_monotone 에도 적용 | T5 |
| M7 | 리더가 표지 없는 부분 모양도 받아들임 | T3 |
| M8 | 해소 알림을 셈과 무관하게 모든 확정 수용에서 냄 | T6 |
| M9 | 해소 알림을 Place 앞(또는 판정 경로 안)에서 냄 | T6(순서) |
| M10 | 압축 판정 질의를 트랜잭션 밖(별도 연결)에서 읽음 | T2 의 동시 기록 사례(같은 트랜잭션 · BEGIN IMMEDIATE 고정) |
| M11 | 압축 판정 질의 오류를 판정 오류로 돌려줌 | T8 |
| M12 | release 가 `released=false` 에도 셈을 더함 | T6 |

### 7.3 리뷰

- 작성자는 Claude 계열이므로, 리뷰 보이스는 다른 계열 둘(예: Codex CLI 두 프롬프트)로 둠.
  ① 적대 Eng: 즉시성 · 중복 가드 · 재생 출처 입증. ② 원장 · 리더 계약 · 롤백. 2 회차가 상한이고, 새 가드 제안은 받지 않음.
- 선택 시험:
  - `go test ./internal/journal/ -run 'A207|ExitSnapshot|Refus|ExitEvent|BrokerOrderExitLinks|AccountExitEvents'`
  - `go test ./internal/app/engine/ -run 'A207|Refus|A197|A094'`
  - `go test ./internal/execgw/ -run 'A166|A207'`
  - `make lint`
  - 전체 시험 · `make gate` 는 사람이 명시 요청할 때만 돌림.

## 8. 위험

- R1 주문 원장은 그대로 쌓임: 2 s 최악이면 포지션당 ≈94~116 MB/일(브로커에 닿은 재시도만). 이 change 밖의 일임. 줄이려면 재시도 빈도를
  건드려야 해서 §0.3 과 부딪침.
- R2 수용된 재시도의 무장 행이 압축될 수 있음(§3.2, 역사 재생 8 행).
- R3 압축 행은 평탄화 ↔ 스냅숏 교차 검사를 받지 못함(§3.2).
- R4 롤백 중 콘솔이 압축 행을 손상으로 표시함(§6). 엔진에는 영향이 없음.
- R5 해소 알림이 사라지는 경우가 넷임: 일반 등급이라 원장에 남지 않음, Place 와 주기 끝 사이에 프로세스가 죽음, relay 버퍼가 가득 참,
  relay 실행자가 멈춤(뒤의 둘은 버렸다는 로그가 남음). 엔진 재시작 뒤의 해소는 셈이 0 이라 알리지 않음. 거절 critical 은 원장에 있으므로,
  잃는 것은 「풀렸다」는 말 하나뿐임.
- R6 a206 정비 뒤에는 「증거는 구간의 첫 행」이 깨질 수 있음(§3.2, 사람 결정 6).
- R7 측정의 한계: 운영 사본 한 시점(2026-11-27)뿐이고, engine.log 는 2026-10-21 부터만 있음. 2 s 주기의 실제 재시도 간격과 쓰기 잠금 경합은
  아직 재지 않았음(a205 배포 전). F4 · F9 는 5 s 시절 기록과 단독 사본에서 추정한 값임.
