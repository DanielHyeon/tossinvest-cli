# Design: add-core-domain

## Context

P1 완료 상태: Gateway는 GuardianDecision을 요구하지만 발급자가 없어 자동 주문은 구조적으로 불가능. journal(v4)·filldetect·reconcile·flatten이 가동 가능. 공식 조건주문 API(SINGLE/OCO/OTO)는 P1의 `trading.Service` Conditional 메서드로 접근 가능. StockOS 인벤토리(docs/stockos-inventory.md)의 순수 로직·테스트 케이스가 이식 소스다.

## Goals / Non-Goals

**Goals**: Guardian 실구현과 게이트 활성화 경로, 포지션·보호·비용·성과의 도메인 코어. tracer slice로 실전 최소 검증.
**Non-Goals**: 전략·후보·스케줄러(P3), 웹 API(P4), 레인 개념(P3 — 심볼 단위 차단으로 대체 중), StockOS의 LLM 게이트·capital stage 퀘스트(미채택).

## Decisions

### D1. 원장 = journal DB 확장 (별도 DB 없음)
Position·saga·provenance·성과 테이블을 journal DB v5+ additive 마이그레이션으로 추가. 단일 writer 락·내구성 계약·백업을 하나로 유지하고 P1이 정의한 import 계약을 스키마 확장으로 대체한다(이관 없음 — 같은 DB).

### D2. 보호주문은 네이티브 조건주문 우선
StockOS는 KIS에 조건주문이 없어 synthetic OCO를 만들었지만, 토스 공식 API는 SINGLE/OCO 조건주문을 제공한다. 브로커 상주 보호는 무인 운영(프로세스 사망·WTS 만료·네트워크 단절)에서 결정적으로 안전하다. synthetic은 폴백으로만. 검증 항목(조건주문 게이트 계약·트리거 동작)은 tracer slice와 verify change의 실계좌 검증에 포함.

### D3. Guardian은 순수 판정 + 발급자
`internal/risk`: 판정 체인은 순수 함수(입력: 의도·포지션·한도·당일 손익·시장 상태 스냅샷), 발급자는 체인 ALLOW를 execgw.GuardianDecision(nonce·만료·한도 스냅샷)으로 변환. StockOS guardian.py의 판정 순서를 표로 이식하되 KIS·LLM·capital stage 항목은 제외하고 P1 reason-code enum에 통합. 테스트는 StockOS test_guardian.py(20)·test_target_stop_contract.py(29)·test_a090(36) 케이스를 Go로 이식.

### D4. 운영 모드는 journal 영속 상태
모드 축은 journal 테이블(현재 모드·전환 이력·주체)로 영속. Gateway·Guardian·flatten이 같은 모드 스냅샷을 읽는다. HALT_ALL도 수동 flatten-all은 통과(§0.3).

### D5. MFE/MAE는 관측 해상도 명시형
StockOS에 정리된 구현이 없어 신규 설계: filldetect의 가격 관측 스냅샷(보유 심볼 한정)을 saga 기간에 기록하고 최대 유리/불리 이동을 계산하되, 관측 간격 필드를 함께 저장해 저해상도 구간의 과대 해석을 막는다. 고빈도 시세 스트림 도입은 P3+ 결정.

### D6. tracer slice는 별도 게이트 뒤
tracer slice(태스크 5.1)는 코드 완성과 무관하게 attestation(verify change) + 게이트 ON 사람 승인 후에만 실행된다. change 완료 게이트(5.3)는 tracer 실행을 요구하지 않는다 — tracer는 verify change 2.x와 함께 사용자 협조 트랙으로 이동 가능.

## Risks / Trade-offs

- [조건주문 실동작(트리거 정확성·부분체결 시 잔량 처리) 미실측] → verify change 실계좌 검증 항목에 조건주문 시나리오 추가, 검증 전 폴백 준비
- [Guardian 수치가 Toss 시장 미검증] → 보수 기본값 + provenance 주석 + audit 잠금(§0.9)
- [MFE/MAE 저해상도] → 해상도 필드로 정직하게 기록, 지표 소비자(P3 배분)가 감안
- [journal DB 비대화] → 성과·가격 관측 테이블에 보존 기간 정책(기본 180일) 포함

## Migration Plan

journal v5+ additive. 롤백 = revert + 구버전 바이너리 ErrSchemaTooNew(데이터 손실 없음).

## Open Questions

- 운영 파라미터 수치(사용자 확정 대기 — 미확정 시 small_live 기본값)
- 조건주문 OCO의 익절 leg 사용 여부(손절 단독 SINGLE로 시작할지) — tracer slice 결과로 결정
