# Change: add-core-domain

> 2026-07-26 2판(동결 해제). 1판 리뷰 2라운드(70건)의 판정과 사용자 추가 요구(손익 극대화 exit 정책)를 반영해 재작성. 판 이력은 `review.md`, 설계 정본은 `design.md` D1~D8. 선행 2a `extend-execution-contract`는 GATE PASS·archive 완료 — 이 change는 그 메인 스펙 위의 **판단 정책**이다.

## Why

2a까지의 상태: 결정 없이는 mutation이 불가능한 봉인된 엔진, 예약·RECONCILE·계산 계약, 멱등 재생 골격 — 그러나 **결정을 발급하는 Guardian이 없어 자동매매는 여전히 구조적으로 불가능하다.** 이 change가 판단 계층을 채운다: Guardian 판정 체인과 발급자, 운영 모드, 포지션 투영, 성과, 그리고 사용자가 요구한 **손익 극대화 exit 정책**(수익 진행에 따라 보호 기준선을 단조 상승시키는 baseline ratchet + profit ladder — StockOS 검증 로직 이식).

1판이 착수 불가였던 이유를 2판이 해소한다: 신호 계층 입력을 요구해 ALLOW가 불가능했던 체인(구조적 RR·등급배수 → P3, 체인은 의도·스냅샷·journal 입력만), 2-클래스 어휘의 HALT_ALL(→ 모드×클래스 표), 달성 불가한 "같은 질의 투영" 주장(→ reconciliation 재배선을 MODIFIED로 정직하게 선언), KIS 수치가 실려 있던 비용 모델(→ 구조만 이식, 수치는 2b 실측), 기각된 단건 상한의 잔존(→ 2a 확정 하한 참조), StockOS 최저 티어 값이던 최소 RR 1.5(→ 잠금값 2.0).

## What Changes

- `internal/costs`: 비용 모델 **구조** 이식(KIS 수치 미이식 — 2b 실측 전 과대 추정 placeholder), 실질 본전, 청산 게이트 적용 금지(§0.3)
- `internal/risk`: Guardian 판정 체인(이식/제외/구조대체 열거·StockOS 매핑 표 산출물), No Stop=No Trade, 위험 기반 수량(배수 1.0 고정), 최소 RR 2.0(순수 산술), 심볼 allowlist(상품 분류 소스 부재의 구조 대체), 재진입 쿨다운. **체인은 사전 검사, 총계 권위는 2a 예약 트랜잭션**(RESERVATION_CONFLICT)
- Guardian 발급자: 2a 결정 계약 구현(RiskIntent/ReductionIntent preimage·멱등키·만료 5s), `ExposureLimiter`로 인터록 단일 출처 충족
- 운영 모드: **모드×클래스 표**(HALT_ALL=위험 감소 허용), 방향 비대칭 승인(보수 자동·완화 승인), 트리거 열거(분석 실패 비트리거), journal 영속
- `internal/position`: 체결+조정 이벤트의 투영(인스턴스·시장 차원·decimal), 상태기계 전이표, reconciliation의 로컬 상태를 이 투영으로 **재배선**
- `internal/exitpolicy` **(사용자 요구)**: baseline ratchet(R 트리거 0.4/0.8/1.0/1.2/2.0 — 기준선 단조 상승 불변식) + profit ladder(multi-rung, 판정/체결 필드 분리) + 실질 본전 결합. 발의는 ReductionIntent, 액추에이션·브로커측 보호는 2c
- 성과 원시 지표(`trade_outcomes` — 실현손익·R·보유시간·도달 exit 단계) + 분석 격리(모드 비강화)
- journal v6 단일 원자 마이그레이션(design D7 표), tracer slice 코드(실전 실행은 verify 트랙)

## Capabilities

### New Capabilities

- `risk-management`: 판정 체인·발급자·모드×클래스 표·provenance 수치 규칙
- `position-ledger`: 투영·상태기계·조정 이벤트·원자성·lineage·스키마 규칙
- `exit-policy`: baseline ratchet·profit ladder·판정/액추에이션 경계
- `trade-analytics`: 비용 모델(구조)·성과 원시 지표·분석 격리

### Modified Capabilities

- `reconciliation`: 로컬 상태 출처를 Position 투영으로 재배선, 불일치 해제 규칙(조정 반영 후 재조회 일치=자동, 영구 승격=운영자)

## Impact

- Affected code: 신규 `internal/{costs,risk,position,exitpolicy}`, journal v6(additive·단일 원자), `internal/reconcile/compare.go` 재배선, 엔진 Guardian 주입. **upstream 무수정 예정**(대상 파일 전부 TossOS 생성)
- 선행 완료: 2a(결정 계약·예약·RECONCILE·인터록). 병행: 2b 측정. 후행: 2c(브로커측 보호·PROTECTION_WEAKENING 발급·발동 주문 방향)
- **게이트 ON 금지 구간 명시**: 이 change 완료 후에도 브로커측 보호가 없으므로(2c 전) 로컬 기준선은 프로세스 사망 시 무력하다 — 자동 진입 게이트 ON은 2b attestation + 2c 완료 후에만. tracer 실전 실행도 동일 조건
- StockOS 이식 상수 규칙: 출처·검증 상태 주석, 미확정 시 small_live 보수 기본값. TossOS는 long-only
