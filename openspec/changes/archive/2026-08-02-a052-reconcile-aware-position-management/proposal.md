## Why

`/positions`와 `/position-management`는 entry/adoption provenance만으로 미국 보유분을 `편입 불가`로 설명하지만, 실제 운영 상태에서는 include 지정과 자동편입 설정이 유효하고 account-wide 영구 대사 차단 때문에 편입이 보류되고 있다. 같은 화면의 자동편입 요약과 a051 API도 저장된 ON·3% 설정 대신 registry 기본 OFF·5%를 보여 운영자가 원인과 실제 정책을 잘못 판단하게 한다.

## What Changes

- 저장된 adoption 설정(desired)과 실행 중 엔진이 로드한 설정(effective)을 공통 read model로 노출한다.
- `journal unknown > already managed > operator released > excluded > adoption candidate and covering reconcile block > adoption candidate pending > unselected` 순서의 순수 projector로 `/positions`, `/position-management`, `/api/v1/positions`, `/api/v1/optimization` 상태를 통일한다.
- running engine의 adoption-blocking tracker projection 범위·사유·시작 시각을 읽기 전용으로 표시해 미국 시장 미지원과 대사 대기를 구분한다.
- canonical effective exit snapshot이 없을 때 actionable 보호선은 만들지 않는다. 원장 t0/initial-stop/baseline이 있으면 `원장 기록 · 실효 미확인` 증거로 별도 표시하고, 현재 실효 보호선으로 승격하지 않는다.
- 미국 include-only 보유분의 정상 fold→adopt→exit t0/provenance와 reconcile-blocked 경로를 실제 engine boundary 회귀 테스트로 고정한다.
- 대사 차단 해제, LIVE 주문, gate/lane/kill-switch 변경은 추가하지 않는다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `operator-console`: actual desired/effective 설정, reconcile-aware 보조 상태와 raw-vs-effective exit 근거를 추가한다.
- `http-api-service`: positions와 optimization이 웹과 같은 actual adoption/reconcile 의미를 반환한다.
- `exit-policy`: 미국 include-only 편입의 정상 경로와 대사 차단 경로를 명시적으로 회귀 검증한다.

## Impact

- `internal/positionpolicy`, engine adoption quote validation, engine-owned position policy command/RPC의 GET read와 sidecar용 runtime-only Unix read transport, console position/position-management view, a051 read adapter와 계약 테스트가 영향을 받는다.
- journal schema와 broker API 호출은 추가하지 않는다. adoption driver와 동일한 `reconcile.Tracker.Blocks()`와 config read seam만 사용한다.
- console/API에는 reconcile resolution capability나 신규 mutation route를 추가하지 않는다. 영구 차단 해제는 authoritative stable re-query와 atomic journal/tracker/gate 전이를 별도 변경에서 설계하기 전까지 기존 operator-only 불변식을 유지한다.
