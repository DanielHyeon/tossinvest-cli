# Change: add-operator-dashboard

## Why

사용자 결정(2026-07-26): 운영 콘솔에 StockOS 수준의 매매 대시보드가 없다 — 보유 종목 상세, 익절·손절 라인, 기준선(래칫)·워터마크, 거래 이력을 웹 화면에서 봐야 자동매매를 신뢰하고 운용할 수 있다. 현재 콘솔은 검증(2b) 전용 화면뿐이다. 필요한 데이터는 이미 존재한다: 브로커 조회 API(보유·매도가능·시세)와 2d가 landed한 journal 투영(positions·exit_states·exit_events·trade_outcomes).

## What Changes

- 콘솔에 **포지션 화면** 추가: 브로커 보유 스냅샷(수량·평단·현재가·평가손익·매도가능)과 journal 투영을 심볼로 조인 — 엔진 관리 포지션은 exit 라인(t0 손절 기준선·워터마크·래칫 단계·ladder rung·부분익절 여부·pending exit)을 함께 표시, `EntryDecisionID`가 빈 보유(수동·외부 취득)는 **관리 외**로 구분 표시(exit 라인 없음이 정상임을 명시)
- 콘솔에 **거래 이력 화면** 추가: trade_outcomes(동결된 왕복 결과)와 exit_events 시간순
- journal **읽기 전용 접근**: RO open, 콘솔은 어떤 journal 쓰기도 하지 않는다
- 브로커 스냅샷 **캐시**: TTL 기반, 화면 새로고침이 rate budget을 소모하지 않게

## Non-Goals

- 주문·게이트·설정 조작 라우트 (콘솔의 부재 보장 유지)
- exit 정책·엔진 동작 변경, 수동 보유에 exit 라인 생성(엔진 관리 밖)
- 원격 접근(127.0.0.1 전용 유지)

## Capabilities

### New Capabilities

- `operator-console`: 로컬 웹 콘솔의 read-only 운영 가시성 계약 — 포지션·exit 상태·거래 이력 표시와 그 안전 불변식(무주문·무게이트·RO journal·rate budget)

### Modified Capabilities

(없음 — execution-verification의 검증 화면 계약은 그대로)

## Impact

- Affected code: `internal/console/`(화면·라우트 추가), journal RO 리더(콘솔용 조회 함수), 브로커 스냅샷 캐시
- 엔진 미가동(게이트 OFF) 동안 journal은 비어 있을 수 있다 — 빈 상태를 정직하게 표시하고 브로커 보유만으로도 유용해야 한다
- 2b·2c와 독립: 측정을 막지 않고, 측정 결과를 요구하지 않는다
