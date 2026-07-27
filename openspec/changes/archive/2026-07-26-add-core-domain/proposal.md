# Change: add-core-domain

> 2026-07-26 3판. 1판 70건 + 2판 38건 리뷰의 판정과 사용자 요구(손익 극대화 exit 정책)를 반영. 판 이력은 `review.md`, 설계 정본은 `design.md` D1~D8. 선행 2a `extend-execution-contract`는 GATE PASS·archive 완료 — 이 change는 그 메인 스펙 위의 **판단 정책**이다.

## Why

2a까지의 상태: 결정 없이는 mutation이 불가능한 봉인된 엔진, 예약·RECONCILE·계산 계약, 멱등 재생 골격 — 그러나 **결정을 발급하는 Guardian이 없어 자동매매는 여전히 구조적으로 불가능하다.** 이 change가 판단 계층을 채운다: Guardian 판정 체인과 발급자, 운영 모드, 포지션 투영, 성과, 그리고 사용자가 요구한 **손익 극대화 exit 정책**(수익 진행에 따라 보호 기준선을 단조 상승시키는 baseline ratchet + profit ladder — StockOS 검증 로직 이식).

1판이 착수 불가였던 이유를 2판이 해소한다: 신호 계층 입력을 요구해 ALLOW가 불가능했던 체인(구조적 RR·등급배수 → P3, 체인은 의도·스냅샷·journal 입력만), 2-클래스 어휘의 HALT_ALL(→ 모드×클래스 표), 달성 불가한 "같은 질의 투영" 주장(→ reconciliation 재배선을 MODIFIED로 정직하게 선언), KIS 수치가 실려 있던 비용 모델(→ 구조만 이식, 수치는 2b 실측), 기각된 단건 상한의 잔존(→ 2a 확정 하한 참조), StockOS 최저 티어 값이던 최소 RR 1.5(→ 잠금값 2.0).

## What Changes

- `internal/costs`: 비용 모델 **구조** 이식(KIS 수치 미이식 — 2b 실측 전 과대 추정 placeholder), 실질 본전, 청산 게이트 적용 금지(§0.3)
- `internal/risk`: Guardian 판정 체인(이식/제외/구조대체 열거·StockOS 매핑 표 산출물), No Stop=No Trade, 위험 기반 수량(배수 1.0 고정), 최소 RR 2.0(순수 산술), 심볼 allowlist(상품 분류 소스 부재의 구조 대체), 재진입 쿨다운. **체인은 사전 검사, 총계 권위는 2a 예약 트랜잭션**(RESERVATION_CONFLICT)
- Guardian 발급자: 2a 결정 계약 구현(RiskIntent/ReductionIntent preimage·멱등키·만료 60s=실장 상수), **원자 발급**(`RecordDecisionAndReserve` — 결정+예약 한 트랜잭션, 거부 시 고아 결정 없음), `ExposureLimiter` 단일 출처, 예약 실패 reason 분화
- 운영 모드: **3모드×클래스 표**(EXIT_ONLY 삭제 — ENTRY_BLOCKED와 행동 동일; HALT_ALL=위험 감소 허용), EntryGate 투영으로 강제, 방향 비대칭 승인, 트리거→목적 상태 열거(전부 ENTRY_BLOCKED·분석 실패 비트리거)
- `internal/position`: 체결+조정 이벤트의 투영(인스턴스·시장 차원·decimal), 상태기계 전이표, reconciliation의 로컬 상태를 이 투영으로 **재배선**
- `internal/exitpolicy` **(사용자 요구)**: **t0 기준선=진입 손절가**(체결 직후부터 손절 판정 유효), baseline ratchet(관측 최고가 워터마크 프로브·후보 합성·기준선/레벨/워터마크 3중 단조) + profit ladder(rung 세트·분모 규칙·판정/체결 분리) + **pending 수명주기**(레벨당 1회·미해소 억제·크래시 복원) + 포지션당 정책 하나(RATCHET|LADDER) + 실질 본전(MAX_RATE 상한). 관측 두절 60초→ENTRY_BLOCKED 자동 강화. 발의는 ReductionIntent, 액추에이션·브로커측 보호는 2c
- 성과 원시 지표(`trade_outcomes` — 실현손익·R·보유시간·도달 exit 단계) + 분석 격리(모드 비강화)
- journal v6 단일 원자 마이그레이션(design D7 표), tracer slice 코드(실전 실행은 verify 트랙)

## Capabilities

### New Capabilities

- `risk-management`: 판정 체인·발급자·모드×클래스 표·provenance 수치 규칙
- `position-ledger`: 투영·상태기계·조정 이벤트·원자성·lineage·스키마 규칙
- `exit-policy`: baseline ratchet·profit ladder·판정/액추에이션 경계
- `trade-analytics`: 비용 모델(구조)·성과 원시 지표·분석 격리

### Modified Capabilities

- `reconciliation`: 메인 스펙 SHALL 전문 보존 위에 — 로컬 상태 출처를 Position 투영으로, 조정 이벤트 compare-and-append, 해제 규칙(조정 반영 후 일치=자동+ADJUSTMENT_APPLIED, 영구=운영자)
- `engine-safety`: 결정 영속(원자 발급·EXPOSURE_RAISING의 HELD 예약 제출 검증), 인터록 조항 6(**ProtectionReady — 브로커측 보호 미배선 시 게이트 ON 거부, 산문이 아니라 기계**) + 가격 조회 endpoint drift guard 편입

## Impact

- Affected code: 신규 `internal/{costs,risk,position,exitpolicy}`, journal v6(additive·단일 원자), `internal/reconcile/compare.go` 재배선, 엔진 Guardian 주입. **upstream 무수정 예정**(대상 파일 전부 TossOS 생성)
- 선행 완료: 2a(결정 계약·예약·RECONCILE·인터록). 병행: 2b 측정. 후행: 2c(브로커측 보호·PROTECTION_WEAKENING 발급·발동 주문 방향)
- **게이트 ON 금지의 기계화**: 인터록 조항 6(ProtectionReady)이 2c 전 게이트 ON을 구조적으로 거부한다 — 로컬 기준선은 프로세스 사망 시 무력하므로. tracer 실전 실행도 동일 조건 + 관측 두절 자동 강화가 무기한 무손절을 차단
- StockOS 이식 상수 규칙: 출처·검증 상태 주석, 미확정 시 small_live 보수 기본값. TossOS는 long-only
