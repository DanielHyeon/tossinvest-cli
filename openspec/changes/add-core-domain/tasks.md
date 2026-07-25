# Tasks: add-core-domain

> [M]=Manager, [T]=Teammate. TDD, 체크박스는 산출물 커밋과 동일 커밋. Pre-Edit 전문은 upstream 파일 수정 시에만(현재 예정 없음 — 발생 시 보고).
> 아키텍처(design.md): 원장=journal DB v5+ 확장, 보호=네이티브 조건주문 우선, Guardian=순수 판정+발급자. 기존 패키지 함수는 additive 원칙, 불가피한 수정은 issues.md 기록.

## 1. 비용·수량·손절 (StockOS 순수 로직 이식)

- [ ] 1.1 [T] `internal/costs`: KRW/USD 수수료·거래세 비용 모델 (StockOS costs·cost_model 이식, provenance 주석, test_costs 케이스)
- [ ] 1.2 [T][High-risk] `internal/risk` 사이징·손절 계약: No Stop = No Trade, 위험 기반 수량(floor, fail-closed), 최소 RR(불가 시 거부) — test_target_stop_contract(29)·test_a090(36) 케이스 이식
- [ ] 1.3 [T] 구조적 RR 계산(measured-move, cap 6.0, 계산 불가 시 None) — test_structural_rr(14) 이식

## 2. Guardian·운영 모드

- [ ] 2.1 [T][High-risk] Guardian 판정 체인(순수 함수): 고정 순서·첫 실패 정지·reason-code 통합 — test_guardian(20) 이식, 순서 표 문서화
- [ ] 2.2 [T][High-risk] 한도 판정: 주문 크기·총 개방 노출·일일 손실(절대액+%, equity≤0 즉시 차단)·중복/재진입 규칙 — 보수 기본값 + provenance 주석
- [ ] 2.3 [T][High-risk] 운영 모드 축: journal 영속(v5 마이그레이션 시작)·전환 audit·critical 알림, HALT_ALL의 수동 flatten 예외, 재시작 유지 테스트
- [ ] 2.4 [T][High-risk] GuardianDecision 발급자: 체인 ALLOW→execgw 계약(nonce·만료 기본 5s·한도 스냅샷), 청산·보호 의도의 진입 판정 면제 규칙
- [ ] 2.5 [T][High-risk] 게이트 활성화 통합: engine 배선에 Guardian 주입, attestation+한도+모드+사람 승인 미충족 전 조합 기동 거부 통합 테스트

## 3. 포지션·Provenance

- [ ] 3.1 [T] journal v5+: Position·운영 모드·provenance·성과·가격 관측 테이블(additive, 보존 기간 180일 정책 포함)
- [ ] 3.2 [T][High-risk] `internal/position` 상태기계: FLAT→OPENING→OPEN→(SCALING|CLOSING)→CLOSED, 체결 delta에서만 파생, reconcile 계좌 우선 — StockOS in_flight 11상태 계약 참조
- [ ] 3.3 [T] `docs/aggregates.md`: Order/Fill/Position/ProtectionSaga 경계·이벤트 흐름 문서
- [ ] 3.4 [T] provenance lineage: GuardianDecision 스냅샷→intent→attempt→fill→position 참조 연결, 단일 질의 재구성 테스트

## 4. 보호주문

- [ ] 4.1 [T][High-risk] 조건주문 어댑터: trading.Service Conditional 경유(upstream 직접 호출 금지), stop 단독(SINGLE)·stop+target(OCO) 등록·취소·정정, httptest 계약 테스트
- [ ] 4.2 [T][High-risk] 진입-보호 saga: 체결 감지→stop-first 제출→확인→ACTIVE, journal 영속, 무보호 노출 SLO(기본 10s) 측정·critical 알림, 보호 IN_DOUBT 시 심볼 진입 차단
- [ ] 4.3 [T][High-risk] 재시작 복구: 미완 saga 감지→보호 재확인·재제출, 미보호 포지션 critical 알림 — crash/restart 테스트
- [ ] 4.4 [T][High-risk] 부분체결 정합: 체결분만 보호, 취소-후-재등록 정정, CLOSED 시 보호 취소, oversell 방지 테스트
- [ ] 4.5 [T][High-risk] synthetic 폴백: 조건주문 거부 케이스 감지→로컬 감시 폴백 활성·critical 알림 (폴백 자체는 최소 구현 — 감시 루프+시장가 아닌 공격적 limit 청산)
- [ ] 4.6 [T] 손절 즉시성 보존 테스트: 진입 차단 latch·EXIT_ONLY·영구 불일치 상태에서 보호·청산 경로 무영향 검증 (§0.3)

## 5. 성과·검증

- [ ] 5.1 [T] 성과 원시 지표: 청산 완결 시 비용 차감 실현손익·R 배수·보유 시간·MFE/MAE(해상도 필드) 기록, 집계는 파생 계산
- [ ] 5.2 [T] tracer slice 코드: 하드코딩 심볼 1개·limit·최소 수량의 진입→보호→청산 end-to-end 실행기(httptest 통합 테스트로 검증) — **실전 실행은 attestation+게이트 ON+사용자 승인 후 별도 수행**

## 6. 완료 게이트 [M]

- [ ] 6.1 diff 리뷰(upstream 무수정 확인)·Pre-Edit/race/crash 확인
- [ ] 6.2 `go test ./...` 독립 재실행 green
- 6.3 (게이트 명령 자체) `make gate CHANGE=add-core-domain` 통과 후 완료 선언
- 6.4 (사용자 확인 후) archive · tracer 실전 실행은 verify change와 함께
