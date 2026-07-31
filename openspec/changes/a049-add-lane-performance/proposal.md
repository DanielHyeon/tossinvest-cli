# a049 · 레인 성과 귀속 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-004`
- **Feature**: `FEAT-TOS-011`
- **Story**: `STORY-TOS-a049`

## Why

현재 거래 outcome 집계는 승률·PF·MDD·R을 제공하지만 어떤 후보·레인·설정이 성과를 만들었는지 결정적으로 연결하지 못한다. 최적화 전에 provenance 기반 lane attribution과 시계열 지표를 마련해야 한다.

## What Changes

- candidate→lane→decision→order→fill→close의 결정적 식별자 연결을 정의한다.
- 비용 후 P&L, realized R, slippage, 5/15/30분 markout과 데이터가 있는 MFE/MAE를 집계한다.
- 링크 누락은 `link_missing`, 시계열 누락은 `not_measured`로 표시하고 0으로 추정하지 않는다.
- 수집·조회 경로는 주문·설정·LIVE 토글 권한을 갖지 않는다.
- a050의 `performance-history` 카테고리에 읽기 전용 성과·변경 이력을 두고 metric 정의, 표본, 기간, provenance와 누락 사유를 설명한다.
- 화면 조회 기본값은 최근 30일, 전체 시장, 전체 lane, 완전한 lineage만 포함이다. 이 필터 기본값은 거래 정책 기본값이 아니며 저장·LIVE 권한을 만들지 않는다.
- **비목표**: 설정 추천·적용, 자본 배분, 자동 승격.

## Capabilities

### New Capabilities

- `lane-performance`: 결정적 레인 귀속, markout, MFE/MAE와 비용 후 성과 계약.

### Modified Capabilities

- `trade-analytics`: 포트폴리오 집계를 lane·policy 버전별로 확장한다.
- `operator-console`: 성과·이력의 읽기 전용 category와 filter 의미를 추가한다.

## Impact

- 신규 `internal/performance`, journal analytics schema/read models, console read surface.
- 보존 기간·시계열 수집 API 예산과 migration 비용을 design에서 확정한다.
