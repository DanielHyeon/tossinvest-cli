# a048 · 시장 시간 인지 스케줄러 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-004`
- **Feature**: `FEAT-TOS-010`
- **Story**: `STORY-TOS-a048`

## Why

TossOS에는 KR/US 시간 계산 primitive가 있지만 전략 진입을 세션·휴장일·DST와 API 예산에 맞춰 기동하는 scheduler가 없다. 자동 재시작과 시장 전환을 진입 권한과 분리해 명시해야 한다.

## What Changes

- KR/US 정규장, 휴장일, 조기폐장과 미국 DST를 반영하는 scheduler를 추가한다.
- 장 닫힘은 신규 진입만 차단하며 reconcile·exit·filldetect를 중지하지 않는다.
- polling cadence는 공식 API rate budget 안에서 결정한다.
- 재시작은 이전에 운영자가 승인한 desired state만 복원한다.
- a050의 `strategy-runtime` 카테고리 안에 `시장·일정` section을 두고 scheduler desired/effective, 시장 범위, 세션, 자동 기동, calendar 상태와 대기 사유를 설명한다.
- 기본값은 scheduler OFF, auto-start OFF, 선택 시장 없음, 정규장만 허용이다. exchange calendar는 자동 관리되는 read-only 근거이며 사용자가 임의 날짜 기본값을 편집하지 않는다.
- **비목표**: 전략 판단, 청산 기준선, 자동 토글 승인.

## Capabilities

### New Capabilities

- `market-aware-scheduler`: 시장 세션·desired state·rate budget 기반 진입 스케줄링.

### Modified Capabilities

- `operator-console`: 시장·일정 설정과 실효 상태 설명을 추가한다.

## Impact

- 신규 `internal/scheduler`, `internal/clock`, engine runtime와 autostart desired-state 설정.
- 장 시간 오류가 LIVE 주문에 영향하므로 시간 경계·DST·휴장 회귀 테스트가 필요하다.
