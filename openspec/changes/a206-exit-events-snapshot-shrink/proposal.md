# a206 — 닫힌 포지션의 exit_events 판정 증거를 비워 운영 저널 용량을 줄인다

> 단계: **1단계(조사 + 제안)**. 이 커밋에는 생산 코드 변경이 없음 — 문서 · 측정 하네스 · AST 근거만 있음.
> 구현(2단계)은 아래 「사람이 결정할 것」이 정해지고 READY 가 된 뒤에 시작함. 저널은 원장 = **High-risk**.

## 왜

운영 저널 `~/.config/tossctl/journal.db` 가 293,179,392 B(사본 기준, user_version 36, WAL, auto_vacuum 0, freelist 0)이고
그중 `exit_events` 테이블이 231.8 MB(+ 인덱스 9.7 MB)로 79% 를 차지함. 사용자 목표는 파일 크기 축소이며, 주문 원장
(intents · decisions · mutation_attempts · attempt_transitions · spent_nonces 등)은 삭제 대상이 아님.

실측(2026-11-27, 읽기 전용 사본, `analysis/a206_measure.out.txt`):

| 사실 | 값 |
|---|---|
| exit_events 행 · 포지션 | 54,883 행 · 130 포지션. 상위 3 포지션이 27,072 + 19,079 + 5,272 = 51,423 행(93.7%) |
| 세 스냅숏 JSON(saved · recomputed · effective) | 183.2 MB — 2026-08 110.0 · 2026-10 66.3 · 2026-11 6.6 MB |
| 닫힌 포지션(CLOSED, exit_states.completed=1) 의 행 | 54,700 행 · 182.06 MB(**99.4%**). 열린 포지션(사본 시점 1 개) 183 행 · 1.14 MB |
| effective = recomputed(바이트 동일) | 28,326 / 28,326 증거 행(**100%**, effective_source 전부 `recomputed`) — 61.2 MB |
| saved = 같은 포지션 직전 행의 effective | 27,242 / 28,326(96.2%) — 58.8 MB |
| 반복 기록 폭증 날 | 2026-08-07 26,103 · 08-08 999 · 08-24 5,236 · 10-21 683 · 10-29 19,081 행 |

폭증의 정체는 **거절된 청산 발의의 주기마다 재무장**이다. 무장 행(LADDER_PARTIAL · STOP_LOSS_LADDER) 하나가 스냅숏
평균 6,479 B 를 쓰고, 해제 훅 행(PROPOSAL_REFUSED · PROPOSAL_CANCELLED)이 뒤따름. 재무장은 손절 즉시성(§0.3)을 위한
**설계 동작**이고(`ExitObserver.submit` 의 해제 5 곳, claims C7), 대기 · 횟수 상한이 없음.

- 8월형(무장 → 브로커 attempt 없이 CANCELLED, 2026-08 13,011 건): a094(2026-09-30 archive)가 진단한 얼어붙은 attempt
  라이브락. 2026-09 이후 0 건.
- 10월형(`opposite_pending_order_exists`, 무장 행 6,494 건): a197(반대 미체결 외부 매수 취소, 운영 반영 2026-11-25)이 다룸.
  11월에도 무장 행 86 건(11-16 10 · 11-25 64 · 11-27 12) — 줄었지만 0 은 아님(원인 · a197 반영 시각 전후 미분석).
- 브로커 확정 거절(`dispatch_rejected_by_broker`, 무장 행 08월 3,119 · 10월 3,325 건): **지금 코드에서도 그대로 가능**.
  a205(관측 주기 5s → 2s, main 착지)로 같은 거절이 계속되면 포지션 하나당 하루 최대 43,200 무장 행 ≈ 스냅숏 280 MB/일
  (5s 일 때 112 MB/일). a205 design R3 「저널 쓰기 2.5배 — 미측정」의 크기가 이것임.

## 무엇을 바꾸나(제안 — 사람 결정 전)

1. **A. 일회성 정비 명령(권장, 지금 크기)** — 엔진이 멈춘 창에서 사람이 실행하는 `mutating: true` 명령.
   닫힌 포지션의 exit_events 행 가운데 **브로커 주문에 연결되지 않은 행**(그 intent 의 attempt 에 broker_order_id 가 없는 행)의
   판정 증거 열 20 개를 전부 NULL 로 바꾸고 VACUUM. 행 자체 · action · proposed_intent_id · 가격 · 시각은 남김.
   사본 실측 **293.2 MB → 69.0 MB**(exit_events 표 231.8 → 12.4 MB). 생산 리더 셋의 결과는 바뀌지 않거나(orders 연결 지문 동일)
   `legacy_event` 로 읽힘(오류 0) — `analysis/a206_reader_check.out.txt`.
2. **B. 앞으로의 중복 저장 줄이기(권장, 성장 원인 — 별도 change)** — effective_source 가 `recomputed` 이면
   `effective_snapshot_json` 을 쓰지 않고 리더가 recomputed 에서 채움. 새 증거 행 바이트의 1/3(평균 2,159 B) 절감.
   쓰기 · 읽기 계약 변경이라 별도 Story/change 로 분리함.
3. **C. 보존 기간 정책** — 지금은 자동화하지 않음. A 명령을 사람이 필요할 때(예: 분기마다) 다시 돌리는 것으로 충분하고,
   엔진 안 자동 정리는 VACUUM 없이는 파일을 줄이지 못함(auto_vacuum 0 — NULL 로 비운 페이지는 재사용만 됨).

이 change(a206)의 2단계 범위는 **A 하나**임. B 는 후속 Story 후보로 기록만 함.

## 비목표

- exit_events 행 삭제, 주문 원장 테이블 변경, 열 삭제(OpenReadOnly 가 세 JSON 열의 **존재**를 요구함 — readonly.go:88-91).
- 청산 재무장 빈도 변경(§0.3 손절 즉시성 — 거절된 손절은 다음 주기에 다시 나가야 함).
- 엔진 안 자동 정리 · 스케줄러.
- orders 화면 연결이 지금 모든 체결 주문에서 ambiguous 인 결함(무장 행 + PROPOSAL_FILLED 훅 행 = 2 행) 수리 — 별도 Story 후보.
- `~/.config/tossctl/*.bak` 정리(17 파일 1.51 GB — 저널보다 5배 큼). 사람이 결정.
- 20 MB 목표: 불가능. A 뒤에도 exit_events 밖 원장 · 인덱스가 ~56 MB 임.

## 사람이 결정할 것

1. A 를 구현할지, 대상 범위를 「닫힌 포지션 전부(연결 행 제외)」로 할지 「닫힌 지 N 일 지난 것만」으로 할지.
2. 비운 행을 지금 리더대로 `legacy_event` 로 읽히게 둘지(리더 무변경, 권장), 아니면 「정비로 비움」을 따로 표시할 열을
   추가할지(스키마 v37 + 리더 변경 — 더 정직하지만 High-risk 코드가 늘어남).
3. 실행 창(엔진 · 콘솔 정지, 두 시장 닫힌 KST 05:00~08:00)과 실행 전 백업을 어디에 둘지.
4. B 를 후속 change 로 열지.
5. `.bak` 1.51 GB 를 어떻게 할지.

## 영향

- 코드(2단계, 새 파일만 목표): `internal/journal/` 정비 함수 1 · `cmd/tossctl` 명령 1 · 시험. 기존 함수 편집 0 을 목표로 함
  (gate ⑤ required 0).
- 스펙: `position-ledger` ADDED Requirement 1(닫힌 포지션 판정 증거의 운영자 정비).
- 운영: 실행은 사람만. 엔진 정지 = 손절 없음이므로 두 시장 닫힌 창에서만.
