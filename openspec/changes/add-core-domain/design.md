# Design: add-core-domain

## Context

P1 완료 상태: Gateway는 GuardianDecision을 요구하지만 발급자가 없어 자동 주문은 구조적으로 불가능. journal·filldetect·reconcile·flatten이 가동 가능. 선행 change `extend-execution-contract`가 조건주문의 Gateway 편입·발동 주문 귀속·safety class·한도 fail-closed·위험 예약을 완성한다. 이 change는 그 레일 위의 **판단**을 구현한다. StockOS 인벤토리(docs/stockos-inventory.md)의 순수 로직·테스트 케이스가 이식 소스다.

## Goals / Non-Goals

**Goals**: Guardian 실구현과 게이트 활성화 배선, 포지션·보호·비용·성과의 도메인 코어. tracer slice로 코드 경로 검증.

**Non-Goals**: 실행 계약 자체(선행 change), 전략·후보·스케줄러(P3), 웹 API(P4), MFE/MAE(P3 — 시세 스트림 도착 후), StockOS의 LLM 게이트·capital stage 퀘스트(미채택), short 지원(구조적 금지).

## Decisions

### D1. 원장 = journal DB 확장 (별도 DB 없음)

Position·saga·성과 테이블을 journal DB v6 additive 마이그레이션으로 추가한다(v5는 선행 change). 단일 writer 락·내구성 계약·백업을 하나로 유지한다.

마이그레이션은 **버전별 immutable**로 분리한다 — 하나의 태스크에 모든 테이블을 넣지 않는다. 각 버전은 키·FK·unique 제약·append-only 여부를 명시하고 스키마 계약 테스트를 동반한다.

**롤백은 구버전 바이너리 실행이 아니다** — `ErrSchemaTooNew`로 기동이 거부되므로 그것은 롤백이 아니라 정지다. 실제 복구 경로는 마이그레이션 직전 자동 백업으로의 복원이며, forward-fix가 기본 정책이다.

### D2. 보호주문은 네이티브 조건주문 우선

토스 공식 API의 SINGLE/OCO 조건주문은 브로커에 상주하므로 무인 운영(프로세스 사망·네트워크 단절)에서 결정적으로 안전하다. 단 그 안전성은 **검증된 속성에 한해** 주장할 수 있다: 프로세스 사망 후 존속, 시장별 지원, 트리거 기준가, 정규장 밖 동작, 만료, OCO sibling 취소, 부분체결 잔량 처리, 정정 원자성. 이 속성들은 `verify-execution-capability`의 실계좌 attestation 항목이며, **검증되지 않은 시장·주문 유형에서는 자동 진입을 하지 않는다.**

SINGLE(손절 단독)로 시작할지 OCO(손절+익절)로 시작할지는 능력 검증 결과로 정한다 — 구현 완료 뒤로 미룰 수 없는 결정이다.

### D3. Guardian은 순수 판정 + 발급자

`internal/risk`: 판정 체인은 순수 함수(입력: 의도·포지션·한도·당일 손익·시장 상태 스냅샷), 발급자는 체인 ALLOW를 선행 change의 결정 계약(주문 해시·RiskIntent 해시·한도 스냅샷·만료·nonce·예약)으로 변환한다.

StockOS `guardian.py`(714줄) 이식 범위를 명시한다:

- **이식**: kill switch·모드, 게이트 상태 latch, 주문 크기 한도, 구조적 손절 계약, 최소 RR, 현금·비용 검증, 중복·재진입 규칙, 총 개방 노출, 일일 손실
- **제외**: KIS 고유 항목, LLM 게이트, capital stage 퀘스트, 미국장 진입 시간창(KR 운영 기준 — 미국 시장 활성화 시 재검토)
- **판단 보류(구현 시 Manager 확인)**: 레버리지/인버스 심볼 클래스, ETF/ETN 서술자 클래스, 당일 재진입 쿨다운 — 세 항목은 Toss 상품 분류 체계 확인 후 결정

### D4. 운영 모드는 journal 영속 상태 + 방향 비대칭 승인

모드 축은 journal 테이블(현재 모드·전환 이력·주체)로 영속한다. 승인 규칙은 **방향 비대칭**이다: 보수 방향 전환(NORMAL→ENTRY_BLOCKED→EXIT_ONLY→HALT_ALL)은 자동·즉시·durable하게 일어나고, 완화·해제만 사람 승인과 audit를 요구한다. 손실 한도 도달·자격증명 실패·outbox 전달 실패가 승인을 기다리는 동안 계속 진입하는 것은 §0의 취지에 반한다.

HALT_ALL은 "모든 자동 제출 중단"이 아니라 "모든 **노출 증가** 중단"이다. 선행 change의 RISK_REDUCING 클래스(보호 생성·증량, reduce-only 청산)는 HALT_ALL에서도 통과한다 — 그렇지 않으면 재시작 시 발견된 미보호 포지션에 손절을 걸 수 없어 "No Stop = No Trade"와 정면 충돌한다.

kill switch와 모드의 우선순위: 둘 중 더 보수적인 쪽이 이긴다. 조합표를 스펙 산출물로 만든다.

### D5. Position의 단일 권위와 조정 이벤트

`reconcile.LocalStateFromJournal`이 이미 순 보유수량을 파생하고 있으므로, `internal/position`은 **같은 질의 위의 투영**으로 만든다 — 포지션 진실을 두 개 만들지 않는다. 체결 delta는 부호가 없으므로(`Applied.Delta >= 0`) 상태기계가 intent side에서 부호를 재도출한다.

"계좌 값 우선"과 "체결에서만 파생"의 모순은 **조정 이벤트**로 푼다: reconcile 불일치는 Position을 직접 덮어쓰지 않고 append-only 조정 이벤트를 발행하며, Position은 (체결 이벤트 + 조정 이벤트)의 투영이 된다. 직접 쓰기 API는 여전히 없고 provenance도 끊기지 않는다. 청산 수량은 선행 change의 상한 규칙(`min(로컬, 계좌)`)을 따른다.

### D6. 보호 saga는 완전한 상태 전이표를 갖는다

`ACTIVE` 하나로는 crash 테스트가 무엇을 증명해야 하는지 알 수 없다. 상태: DESIRED → RECORDED → DISPATCHED → (ACTIVE | AMBIGUOUS) → (REPLACING | CANCELING | DEGRADED) → CLOSED. 각 상태의 허용 mutation·timeout·재시작 행동과, 각 journal 커밋 전후의 crash point를 fault-injection 수용 매트릭스로 만든다.

수량 정정은 **원자적 정정 우선**이다. `official.ModifyConditionalOrder`(internal/official/conditional_writes.go:63)가 존재하므로, 취소-후-재등록을 일괄 강제하면 네이티브 조건주문을 택한 이유인 무보호 창을 스스로 만든다. 정정 원자성은 D2의 능력 검증 항목이고, 검증되지 않으면 그때 취소-후-재등록으로 폴백한다.

"보호 수량은 항상 체결 수량과 일치"는 동시 체결 환경에서 달성 불가능한 서술이므로, **측정 가능한 transient invariant + 최대 허용 시간**으로 대체한다.

### D7. tracer slice는 별도 게이트 뒤

tracer는 코드 완성과 무관하게 attestation + 게이트 ON 사람 승인 후에만 실전 실행된다. change 완료 게이트는 httptest 통합 검증만 요구한다. **live 검증이 끝나기 전에는 자동화 게이트가 계속 OFF**라는 것이 명시적 산출물이다.

## Risks / Trade-offs

- [Guardian 수치가 Toss 시장 미검증] → 보수 기본값 + provenance 주석 + audit 잠금(§0.9). 비용 bps는 **과대 추정**이 보수 방향. "검증됨" 전환은 verify/tracer 결과에 결속
- [조건주문 실동작 미실측] → D2의 속성별 attestation, 미검증 시장·유형은 자동 진입 금지
- [D5 조정 이벤트가 reconcile 동작을 바꾼다] → 기존 reconcile 테스트 회귀 금지
- [journal DB 비대화] → 성과 테이블에 보존 기간 정책(기본 180일), 삭제·집계는 주문 경로 밖 비동기

## Migration Plan

journal v6 additive, 버전별 분리. 마이그레이션 전 자동 백업 → 실패 시 복원. forward-fix 기본.

## Open Questions

- 운영 파라미터 수치(사용자 확정 대기 — 미확정 시 small_live 기본값)
- D3의 판단 보류 3항목(레버리지/인버스, ETF/ETN, 재진입 쿨다운) — Toss 상품 분류 확인 후
- SINGLE 단독 vs OCO — 능력 검증 결과로 결정(구현 뒤로 미루지 않음)
