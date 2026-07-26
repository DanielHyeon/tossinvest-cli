# Change: add-core-domain

> **[동결 — 재작성 대기]** 2026-07-26 2라운드 리뷰(25건, `review.md` 후반부)에서 착수 불가 판정. 이 스펙은 선행 계약(extend-execution-contract) 재작성 **이전** 기준으로 쓰여 있어 사실과 다르다: 2-클래스 어휘, 기각된 `min(로컬,계좌)` 상한, 빈 스냅샷 표지 등. 재작성은 2a 구현 완료 후 수행하며, 그때 반영할 확정 사항 — 구조적 RR·등급배수는 입력 생산자 부재로 P3 이관, HALT_ALL 어휘를 모드×클래스 표로, 비용 모델은 KIS 수치 이식 금지(2b 2.9 실측값), 최소 RR provenance 정정(1.5는 StockOS 최저값), journal 버전 단일화, 판정 체인과 예약 트랜잭션의 권위 관계 명시, flatten·발동 주문 방향 소유권은 2c와 협의.

## Why

P1이 만든 주문 실행 계층(Gateway·journal·체결 감지·reconcile)은 GuardianDecision 없이는 주문을 낼 수 없다 — 그런데 Guardian은 아직 초안 인터페이스뿐이다. 실전 직행(U2) 체제에서 자동매매를 켜는 마지막 조각은 위험 엔진(한도·손절·사이징)과 포지션·보호주문 체계다. StockOS에서 검증된 거래 불변조건(Guardian 판정 순서, No Stop = No Trade, 위험 기반 수량, kill switch BLOCK-ONLY)을 Toss 코드베이스에 이식한다.

이 change는 **판단 정책**을 다룬다. 그 판단을 강제하는 레일(조건주문의 Gateway 편입, 발동 주문 귀속, safety class, 한도 fail-closed, 위험 예약)은 선행 change `extend-execution-contract`가 만든다 — proposal-freeze 리뷰 45건(`review.md`)이 그 분리를 요구했다.

## What Changes

- `internal/costs`: KRW/USD 수수료·거래세 비용 모델 (StockOS 이식, provenance 주석, 보수 방향은 과대 추정)
- `internal/risk`: Guardian 판정 체인(StockOS 순서 보존, 이식 범위 명시), 구조적 손절·위험 기반 수량·최소 RR, 일일 손실·총 개방 노출의 **계산 계약**(권위 데이터·통화 정규화·거래일 경계·stale 시 fail-closed)
- GuardianDecision 발급자: 선행 change의 계약(주문 해시·RiskIntent 해시·한도 스냅샷·만료·nonce·예약)을 채우는 쪽. 위험 감소 의도는 빈 한도 스냅샷
- kill switch(BLOCK-ONLY) + 운영 모드 축(NORMAL/ENTRY_BLOCKED/EXIT_ONLY/HALT_ALL) — journal 영속·audit·알림. **보수 방향 전환은 자동·즉시, 완화·해제만 사람 승인**
- 자동화 게이트 활성화 배선: 선행 change가 강화한 인터록 전제조건을 실제로 충족시키는 엔진 측 Gateway 구성
- `internal/position` + journal v6: 포지션 상태기계(완전한 전이표), 단일 권위, reconcile 조정 이벤트, aggregate 경계 문서, provenance lineage
- 보호주문: 공식 조건주문(SINGLE/OCO) 네이티브 우선. 진입-보호 saga의 **완전한 상태 전이표**, 수량 정합(원자적 정정 우선), 재시작 복구, 폴백 조건
- 성과 원시 지표: 비용 차감 실현손익·R 배수·보유 시간 (MFE/MAE는 데이터 소스 부재로 P3 이관 — `review.md` E1)
- tracer slice: 종목 1개·limit·최소 수량 end-to-end 실행기 (httptest 검증, 실전 실행은 verify change 트랙)

## Capabilities

### New Capabilities

- `risk-management`: Guardian 판정 체인·한도 계산 계약·kill switch·운영 모드·게이트 활성화 배선
- `position-ledger`: 포지션 상태기계·aggregate 경계·조정 이벤트·provenance·원장 스키마
- `protection-execution`: 보호주문과 진입-보호 saga
- `trade-analytics`: 비용 모델과 성과 원시 지표

### Modified Capabilities

(없음 — 게이트 인터록과 실행 계약의 수정은 선행 change `extend-execution-contract`가 담당하고, 이 change는 그 계약을 구현한다)

## Impact

- Affected code: 신규 `internal/{risk,position,protection,costs}`, journal 스키마 v6(additive), engine 프로필의 Gateway·Guardian 배선. upstream 무수정 목표
- **선행 필수**: `extend-execution-contract` 완료. 이 change의 보호주문·게이트 태스크는 그 계약 없이는 안전하게 구현할 수 없다
- 병행: `verify-execution-capability`가 조건주문 능력 attestation을 생성. tracer 실전 실행과 자동 진입은 그 attestation + 사용자 승인 이후
- StockOS 이식 상수 규칙: 모든 수치에 출처·검증 상태 주석, Toss 검증 전 보수 기본값(운영 파라미터 미확정 시 small_live: 주문 100만 / 노출 1,000만 / 일손실 10만 KRW 또는 1%)
- **TossOS는 long-only**: SELL은 보유수량 이하 reduce-only이며 short 노출은 구조적으로 금지한다
