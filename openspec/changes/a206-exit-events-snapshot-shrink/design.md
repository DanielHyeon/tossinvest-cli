# a206 design — exit_events 판정 증거 정비

> 기준: origin/main `e9266bfb6`, 2026-11-27. 수치는 모두 운영 저널의 읽기 전용 사본에서 집계로만 잼
> (`analysis/a206_measure.py` → `a206_measure.out.txt`, 리더 확인 `a206_reader_check.go.txt` → `.out.txt`).
> 분기 주장은 `analysis/ast-evidence/claims.md`(C1~C7), 리더 열거는 `analysis/code-context/evidence-reconciliation.md`.

## 1. 사실(F)

| # | 사실 | 근거 |
|---|---|---|
| F1 | 파일 293,179,392 B, page 4096 × 71,577, freelist 0. dbstat: exit_events 231.8 MB, decisions 13.4, mutation_attempts 12.1, attempt_transitions 6.6, exit_events 인덱스 셋 9.7 MB | measure.out |
| F2 | exit_events 는 주석으로만 append-only(core_domain.go:31 · :244). 트리거 0, 생산 코드의 UPDATE · DELETE 0 | reconciliation |
| F3 | 세 스냅숏을 읽는 생산 리더는 3 개: `Journal.ExitEvents`(생산 호출자 0) · `ReadOnly.AccountExitEvents`(콘솔 개요 · 포트폴리오 — Evaluation 안 씀) · `ReadOnly.BrokerOrderExitLinks`(콘솔 orders) | reconciliation, C5 |
| F4 | 엔진 · execgw 의 exit_events 리더(HasProtectiveExitAttempt · HasExitEventForIntent · ExitProposalForIntent · decision_id 중복 검사 · trade outcome · backfill · provenance)는 action · proposed_intent_id · decision_id · level_after 만 읽음. **엔진은 닫힌 포지션 행의 스냅숏을 다시 읽지 않음** | reconciliation |
| F5 | 증거 열 20 개가 전부 NULL 이면 `legacy_event`(오류 없음), 일부만 있으면 `ErrExitSnapshotCorrupt` | C1 · C2 |
| F6 | 읽기 전용 두 리더는 행 손상을 행 단위 `invalid_event_evidence` 로 바꾸고 목록은 계속 돌려줌 | C3 · C4 |
| F7 | 판정 행 하나가 세 JSON 을 따로 직렬화(평균 각 2,159 B). effective = recomputed 100%(28,326/28,326), saved = 직전 행 effective 96.2%(27,242) | C6, measure.out |
| F8 | 닫힌 포지션 행이 스냅숏 바이트의 99.4%(182.06 / 183.2 MB; 2026-11-27 05:3x 사본 기준 CLOSED 129 · OPEN 1). CLOSED 인데 exit_states 가 completed 0 인 포지션 0 | measure.out |
| F9 | 닫힌 포지션 행 중 broker_order_id 를 가진 attempt 에 연결된 행 258 개(그 attempt 는 전부 확정 PLACE) | measure.out |
| F10 | 확정 PLACE intent 134 개 전부가 exit_events 2 행(무장 + PROPOSAL_FILLED)을 가져 orders 연결이 모두 ambiguous — 지금 스냅숏이 화면에 나오는 주문 0 | reader_check.out |
| F11 | 거절된 발의는 해제 뒤 다음 관측에서 재무장(대기 · 상한 없음). 관측 주기는 main 에서 2 s(a205, exitloop.go:101) | C7 |
| F12 | `~/.config/tossctl` 의 journal `.bak` 17 파일 1,509,904,032 B(디렉터리 1.8 GB) | `ls -l` 크기 합 |

## 2. 반복 기록의 원인과 현재 상태

| 유형 | 기간 · 규모 | 판정 | 현재 |
|---|---|---|---|
| R-a 얼어붙은 attempt 라이브락(무장 → attempt 없이 CANCELLED, 5 s 주기) | 2026-08: 무장 13,011 · CANCELLED 12,996 | a094 proposal.md:211 이 같은 모양을 진단(「STOP_LOSS_LADDER → PROPOSAL_CANCELLED 5초 주기 — 라이브락」) | 2026-09 이후 attempt 없는 무장 0 건 — 이 유형은 멈춤(수리 커밋 귀속은 a094 archive 까지만 확인, 커밋 단위 미추적) |
| R-b 반대 미체결(`opposite_pending_order_exists`) | 무장 행 기준 10월 6,494 · 11월 86(11-16 10 · 11-25 64 · 11-27 12) | a197 이 외부 반대 매수를 취소해 손절을 통과시킴 | a197 운영 반영(2026-11-25) 날과 그 뒤(11-27)에도 있음. 줄었으나 0 아님, 원인 · 반영 시각 대비 전후 미분석 |
| R-c 브로커 확정 거절(`dispatch_rejected_by_broker`) | 무장 행 기준 08월 3,119 · 10월 3,325 · 11월 1 | 설계 동작(§0.3 재발의) | **재발 가능**. 2 s 주기에서 하루 최대 43,200 무장 행 × 6,479 B ≈ 280 MB/일/포지션 |

**결론**: 「루프 버그」는 R-a 하나였고 고쳐짐. R-b · R-c 는 손절을 계속 다시 내는 정상 경로이며, 용량 문제의 근본은
**재시도 한 번마다 같은 내용의 스냅숏 세 벌을 새 행에 쓰는 저장 방식**(F7)임. 재시도 빈도는 §0.3 때문에 건드리지 않음.

## 3. 선택지와 사본 실측

사본 = 운영 저널 `?mode=ro` → SQLite backup API. 각 시나리오는 새 사본에 UPDATE → VACUUM → 크기 → 삭제.

| 시나리오 | 바뀐 행 | VACUUM 뒤 파일 | exit_events 표 | 리더 결과(reader_check) |
|---|---|---|---|---|
| S0 VACUUM 만 | 0 | 290,603,008 | 231,841,792 | 기준 |
| S1 닫힌 · 미연결 행의 JSON 셋 NULL | 54,442 | 83,640,320 | 24,879,104 | `Journal.ExitEvents` 가 닫힌 포지션 124/129 에서 **오류**, 계좌 이력 28,025 행 `invalid_event_evidence` |
| **S2 닫힌 · 미연결 행의 증거 20 열 NULL** | 54,442 | **68,952,064** | **12,402,688** | 오류 0. 계좌 이력은 `legacy_event` 54,582 · 스냅숏 있음 301. orders 연결 지문 S0 과 동일 |
| S3 닫힌 행 전부(연결 포함) JSON NULL | 54,700 | 82,722,816 | 23,961,600 | (S1 과 같은 손상 + 연결 행도 손상) |
| S4 중복 제거 모의(saved=직전 effective 를 먼저, 그다음 effective=recomputed 를 NULL) | 55,568 | 175,517,696 | 116,756,480 | 모의 — 리더 계약 위반이라 그대로 쓸 수 없음 |

- S1 은 버림: 정비가 원장을 「손상」으로 읽히게 만듦(F5). a084 개정 2 의 교훈(자기 리더가 거절하는 값을 쓰는 것)과 같은 결함.
- S2 가 가장 작고 리더 무변경임. 대가: 그 행의 decision_id · policy · observation 같은 평탄화 열도 사라짐. 이 열의 생산 독자는
  decision_id 중복 검사(exit_state.go:494) 하나인데, 그 검사는 **판정 중인 열린 포지션의** decision id 를 찾으므로
  CLOSED(최종 — core_domain.go positions 주석) 포지션의 행에서는 일치할 일이 없음(decision id 는 snapshot id 에서,
  snapshot id 는 position id · generation · observation id 를 담은 입력 지문에서 유도 — exitpolicy/snapshot.go:254-269, 손 읽기).
- S4 는 「앞으로 덜 쓰기」의 상한 추정: 과거 행까지 다시 쓰면 −115 MB(S0 대비). B 가 제안하는 effective 생략만이면 새 증거 행 바이트의 1/3,
  saved 까지 직전 행 참조로 바꾸면 2/3 이지만 후자는 한 행만 읽는 리더(orders 연결)가 직전 행을 찾아야 해서 B 범위에서 뺌.

## 4. 결정(D) — 2단계 A 의 설계

- **D1 대상 행**: `positions.state='CLOSED'` 이고 `exit_states.completed=1` 인 포지션의 exit_events 행 가운데,
  `proposed_intent_id` 의 mutation_attempts 에 `broker_order_id` 가 비지 않은 행이 **하나도 없는** 행. (+ 사람 결정 1 의 경과 일수 조건)
  **a207 압축 행 · 앵커 제외(2026-11-27 사람 결정, a207 design D8 과 같은 술어)**: 그 가운데 (i) `effective_source = 'recomputed_repeat'` 인
  압축 행 자신과 (ii) 어떤 압축 행의 앵커인 행은 대상에서 뺌. 앵커 = 같은 포지션에서 압축 행보다 id 가 작은 무장 행(`proposed_intent_id` 가 있고
  action 이 OPENED · READOPTED · PROPOSAL_REFUSED · PROPOSAL_CANCELLED · PROPOSAL_FILLED · COMPLETED · ADJUSTMENT_CLOSED 가 아님) 중
  `recomputed_snapshot_json IS NOT NULL` 이고 `effective_source <> 'recomputed_repeat'` 인 가장 큰 id 의 행. 압축 행을 남기는 이유는 표지가
  사라지면 다음 실행이 앵커를 못 찾아 비우므로 D6(3) 멱등이 깨지기 때문임. 제외 집합은 UPDATE 전에 같은 트랜잭션에서 계산함:
  ```sql
  SELECT c.id FROM exit_events c WHERE c.effective_source = 'recomputed_repeat'
  UNION
  SELECT (SELECT max(a.id) FROM exit_events a
           WHERE a.position_id = c.position_id AND a.id < c.id
             AND coalesce(a.proposed_intent_id, '') <> ''
             AND coalesce(a.action, '') NOT IN ('OPENED','READOPTED','PROPOSAL_REFUSED','PROPOSAL_CANCELLED',
                                                'PROPOSAL_FILLED','COMPLETED','ADJUSTMENT_CLOSED')
             AND a.recomputed_snapshot_json IS NOT NULL
             AND coalesce(a.effective_source, '') <> 'recomputed_repeat')
    FROM exit_events c WHERE c.effective_source = 'recomputed_repeat';
  ```
  2026-11-27 운영 저널에는 압축 행이 없으므로(a207 배포 전) 이 제외는 0 행이고 §3 의 사본 실측 수치는 그대로임.
- **D2 바꾸는 값**: 증거 20 열(`exit_snapshot.go` schemaV10 의 ADD COLUMN 중 exit_events 몫, lifecycle_generation 제외) 전부 NULL.
  나머지 열(id · position_id · 가격 셋 · level_after · action · proposed_intent_id · created_at · lifecycle_generation)은 그대로.
  한 트랜잭션. 이미 20 열이 NULL 인 행은 세지 않음(멱등).
- **D3 표면**: `tossctl engine exit-evidence-compact`(이름은 구현 때 기존 명령 가족에 맞춤), `mutating: true`, 운영자 · 승인 참조 ·
  note 필수(a066 · a092 완화 명령 가족 계약), 엔진 잠금(enginelock)을 잡지 못하면 거절. 기본은 `--dry-run` 과 같은 집계만,
  `--apply` 일 때만 씀.
- **D4 순서**: 잠금 → `VACUUM INTO <journal>.pre-a206-<UTC>.bak`(backup.go 와 같은 방식) → UPDATE 한 트랜잭션 → `PRAGMA integrity_check`
  → `VACUUM` → 집계(전 · 후 행 수 · 바이트 · 파일 크기)를 audit 에 기록. 실패하면 UPDATE 는 롤백되고 백업은 남김.
- **D5 새 코드만**: journal 정비 함수와 명령은 새 파일에 둠. 기존 리더 · 쓰기 함수는 고치지 않음(gate ⑤ required 0 목표).
- **D6 시험**: (1) 대상 술어 — 열린 포지션 · 연결 행 · 훅 행 · a207 압축 행과 그 앵커가 안 바뀜, (2) 정비 뒤 `ExitEvents` · `AccountExitEvents` ·
  `BrokerOrderExitLinks` 가 오류 없이 같은 연결 결과, (3) 멱등(두 번째 실행 0 행), (4) 엔진 잠금 보유 중 거절, (5) `--apply` 없이는 쓰기 0.

## 5. 위험과 되돌리기

- 되돌리기: D4 의 백업 파일을 journal.db 자리에 복사(backup.go:27 절차). 백업은 사람이 검증 뒤 지움 — 지우기 전까지는 디스크가 오히려
  293 MB 늘어남.
- 엔진 정지 = 손절 없음. 실행은 두 시장이 닫힌 창에서만(사람 결정 3). VACUUM 자체는 사본에서 시나리오당 약 1 초.
- 콘솔(OpenReadOnly)이 열려 있으면 VACUUM 이 busy 로 실패할 수 있음 → 콘솔도 정지하거나 busy 시 롤백 없이 재시도 안내.
- 정보 손실: 닫힌 포지션의 판정별 스냅숏은 복구 불가(백업 제외). 그 포지션의 trade_outcomes · 주문 원장 · exit_events 의
  action · 가격 · 시각은 남음.
- 재발: R-c 가 다시 길게 돌면 A 만으로는 다음 정비까지 다시 쌓임 → B(후속)가 근본.

## 6. 후속 후보(이 change 밖)

- B: effective_source=`recomputed` 일 때 effective JSON 생략 + 리더 복원(쓰기 `appendExitEventTx` · 읽기 `hydrateExitEventEvidence`
  계약 변경, High-risk, FLM 필수).
- orders 연결 ambiguous(F10): `event_count` 가 PROPOSAL_FILLED 훅 행을 셈.
- R-b 잔여 원인 분석(a197 뒤 11-25 · 11-27).
