# Change: harden-execution-base

## Why

TossOS는 실계좌 자동매매를 실전 직행으로 운영한다(사용자 결정 U2). 그 전제는 주문 실행 계층이 "전략 손실"과 "시스템 결함"을 구분할 수 있을 만큼 신뢰 가능해야 한다는 것이다. 현재 upstream 코드에는 주문 상태기계·멱등성·체결 감지 상태 반영·재시도 정책·위험 인터록·관측성이 전무하고(P0 정찰·베이스라인 확인), 엔진이 호출할 배선은 CLI 내부에 갇혀 있다.

## What Changes

- `internal/app` 신설: CLI 배선(`newAppContext`)을 엔진·CLI 공용으로 승격(위임 shim), 엔진 프로필은 official-only 브로커 강제
- `trading.MutationResult` → `internal/domain` 이동(type alias로 upstream 호환), 조건주문 게이트 `internal/trading` 이동
- 주문 상태기계 코드화(공식 API status 스키마 정본, IN_DOUBT/AMEND_IN_DOUBT 포함) + durable intent journal(제출 전 영속 기록, ext4 경로)
- 체결 감지: 공식 API 주기 폴링 권위 + 신선도 SLO, SSE는 coalescing 힌트
- retry matrix(주문 mutation 자동 재시도 금지), 재시작 reconciliation 계약, FX·interactive auth fail-closed
- 자동화 게이트 설계(기본 OFF, Guardian 미주입 시 기동 거부 인터록 — 활성화는 Phase 2)
- 관측성(구조화 로깅·셀 수 있는 이벤트)과 등급화 알림(critical outbox + heartbeat), `tossctl` flatten-all 비상 saga(--dry-run 포함)
- (분리) 공식 API capability soak·실계좌 검증·약관 검토는 wall-clock·사용자 의존이므로 별도 change `verify-execution-capability`로 분리 — 본 change의 gate를 막지 않는다. 단, 자동화 게이트 ON은 그 change의 attestation 없이는 불가(기동 인터록)

## Capabilities

### New Capabilities

- `order-execution`: 주문 상태기계, intent journal, 멱등성, retry matrix, fail-closed 분기
- `fill-detection`: 체결 감지 권위(폴링)와 SSE 힌트 규약
- `reconciliation`: 계좌 권위 대사 계약과 재시작 복구
- `engine-safety`: 엔진 배선 안전(official-only 증명, 기동 인터록, flatten-all, 관측성·알림)

### Modified Capabilities

- `sdd-workflow`: StockOS SDD 규칙 이식(사용자 지시 2026-07-26) — 최상위 안전 불변식, High-risk Pre-Edit 선언, 완료 보고 조건 요구사항 추가

## Impact

- Affected code: `internal/app`(신규), `internal/domain`, `internal/trading`, `cmd/tossctl`(shim·flatten-all), `internal/push`(소비자 신규), 신규 패키지 `internal/journal`·`internal/filldetect`·`internal/reconcile`·`internal/obs`
- upstream 테스트 650개 회귀 금지(shim·type alias 전략), 신규 코드는 httptest 계약 테스트 동반
- 후속 의존: Phase 2 위험 엔진(T2.4)이 이 change의 인터록에 연결되어야 자동 주문이 활성화됨
- 알려진 잔존 리스크(리뷰 기록): 계좌 단일 writer lease와 CLI/MCP 주문 중재는 Phase 4 데몬과 함께 구현 — 그 전까지 MCP 표면은 Gateway를 우회할 수 있으며 사용자 수동 영역으로 취급, reconciliation이 external provenance로 격리
