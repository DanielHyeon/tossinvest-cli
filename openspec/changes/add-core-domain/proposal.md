# Change: add-core-domain

## Why

P1이 만든 주문 실행 계층(Gateway·journal·체결 감지·reconcile)은 GuardianDecision 없이는 주문을 낼 수 없다 — 그런데 Guardian은 아직 초안 인터페이스뿐이다. 실전 직행(U2) 체제에서 자동매매를 켜는 마지막 조각은 위험 엔진(한도·손절·사이징)과 포지션·보호주문 체계다. StockOS에서 검증된 거래 불변조건(Guardian 판정 순서, No Stop = No Trade, 위험 기반 수량, kill switch BLOCK-ONLY)을 Toss 코드베이스에 이식한다.

## What Changes

- `internal/risk`: Guardian 판정 체인(StockOS 순서·수치 보존, provenance 주석, env 결합 제거), 일일 손실·총 노출·수량 한도, 구조적 손절·위험 기반 수량·최소 RR, GuardianDecision 발급(execgw one-shot nonce 계약 구현)
- kill switch(BLOCK-ONLY) + 운영 모드 축(NORMAL/ENTRY_BLOCKED/EXIT_ONLY/HALT_ALL) — journal 영속·audit·알림 연동
- 자동화 게이트 활성화 경로 완성: attestation + Guardian 주입 검증(P1 인터록)을 통과하는 실제 배선
- `internal/position` + journal 스키마 확장: 포지션 상태기계, aggregate 경계(Order/Fill/Position/ProtectionSaga) 문서화, provenance lineage
- 보호주문: **공식 조건주문(SINGLE/OCO) 네이티브 우선** — 프로세스 사망에도 브로커측 보호 유지, 불가 케이스만 synthetic saga 폴백. 진입 체결→보호 완료 노출 SLA, 부분체결 수량 조정, 재시작 복구
- `internal/costs` + 성과 원시 지표: 거래 비용 모델(KRW/USD 수수료·세금), R 배수·PF·MDD·승률, MFE/MAE 신규 설계(filldetect 스냅샷 기반)
- tracer slice: 레인 1개·종목 1개·limit·최소 수량 실전 end-to-end (attestation·게이트 ON·사용자 승인 선행)

## Capabilities

### New Capabilities

- `risk-management`: Guardian 판정 체인·한도·kill switch·운영 모드·게이트 활성화
- `position-ledger`: 포지션 상태기계·aggregate 경계·provenance·원장 스키마
- `protection-execution`: 보호주문(네이티브 조건주문 우선)과 진입-보호 saga
- `trade-analytics`: 비용 모델과 성과 원시 지표

### Modified Capabilities

(없음 — engine-safety의 인터록 Requirement는 그대로이고 이 change가 그 계약을 구현)

## Impact

- Affected code: 신규 `internal/{risk,position,protection,costs,analytics}`, journal 스키마 v5+(additive), execgw Guardian 배선, engine 프로필 배선. upstream 무수정 목표(조건주문은 P1의 trading.Conditional 경유)
- 선행: P1 archive 완료(됨). tracer slice 실행은 verify-execution-capability의 attestation + 사용자 승인 후
- StockOS 이식 상수 규칙: 모든 수치에 출처·검증 상태 주석, Toss 검증 전 보수 기본값(운영 파라미터 미확정 시 small_live: 주문 100만/노출 1,000만/일손실 10만 KRW 또는 1%)
