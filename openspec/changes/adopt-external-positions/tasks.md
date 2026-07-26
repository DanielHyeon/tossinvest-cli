# Tasks: adopt-external-positions

> 선행: 적대적 리뷰(proposal-freeze) 통과 후 착수. 실효는 게이트 ON 이후(엔진 가동 중에만 편입·관측) — 코드는 2b·2c와 병행 landed 가능하나 **2c의 보호주문 계약과 충돌하지 않게** 편입 포지션도 ProtectionReady 경로를 동일하게 타야 한다. `internal/console`은 건드리지 않는다(1.8·대시보드와 충돌 방지 — 대시보드는 데이터 주도라 exit_states 행이 생기면 자동 표시된다).

## 1. 결정·원장 [T]

- [ ] 1.1 ADOPTION 결정 class + 전용 preimage(심볼·시장·수량·비용기준과 출처 표기·합성 손절·관측 시각) — 기존 decisions 스키마·preimage 재검증 규칙 준수, 원장 스키마 확장 규칙(guarded-column layout) 준수
- [ ] 1.2 포지션 결정 참조 부여: ADOPTION 결정 영속 → `entry_decision_id` 지정(수량·평단 무변경, 투영 규칙 위반 없음 — 조정 이벤트 경로 재사용 여부는 구현 판단, 편차 기록)
- [ ] 1.3 exit_state open(합성 t0): EntryPrice 출처 규칙(평단 우선, 부재 시 관측가), InitialStop = EntryPrice×(1−default_stop_pct), 크래시 복구(결정 영속 후 open 전 크래시 → 재시작이 완결, 중복 편입 금지)

## 2. 편입 파이프라인 [T]

- [ ] 2.1 무결정 보유 관측 → 편입 후보 판정: 인터록 활성 + RECONCILE 아님 + 신선한 보유 확인, `adoption.exclude_symbols`(기본 빈 목록) 제외, 제외·실패 시 기존 알림 유지
- [ ] 2.2 관측 루프 통합: 편입 포지션이 래칫·ladder·부분익절·pending·staleness 규칙을 엔진 진입 포지션과 동일 코드 경로로 받는 통합 테스트(합성 t0 하회 → RISK_REDUCING 발의 포함)
- [ ] 2.3 `adoption.default_stop_pct` 보수 기본값 + provenance(StockOS 대응 계약 확인, 없으면 산정 근거 문서화 — 임의 수치 금지), config 배선
- [ ] 2.4 편입 이벤트 기록·통지(편입가·합성 손절 포함) — 기존 "외부 포지션 발견" 알림 대체

## 3. 완료 게이트 [M]

- [ ] 3.1 §0 검토 기록: 보수 방향 확인(보호 추가만·사이징 무변경·감축 방향 매도만), flatten·kill switch가 편입 포지션을 덮는지 테스트
- [ ] 3.2 테스트 전수(-race)·`openspec validate adopt-external-positions --strict`
- [ ] 3.3 `make gate CHANGE=adopt-external-positions`
