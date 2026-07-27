# Change: add-operator-dashboard

## Why

사용자 결정(2026-07-26): 운영 콘솔에 StockOS 수준의 매매 대시보드가 없다 — 보유 종목 상세, 익절·손절 라인, 기준선(래칫)·워터마크, 거래 이력을 웹 화면에서 봐야 자동매매를 신뢰하고 운용할 수 있다. 필요한 데이터는 이미 존재한다: 브로커 조회 API(보유 — 현재가·평단 포함)와 2d가 landed한 journal 투영(positions·exit_states·exit_events·trade_outcomes). 부수 효과: 콘솔의 안전 불변식(루프백·세션 토큰·핸드오프 토큰·CSRF)은 지금까지 코드·테스트에만 있고 어떤 스펙도 소유하지 않는다 — 이 change의 `operator-console` capability가 성문화한다.

## What Changes

- 콘솔에 **포지션 화면**: 브로커 보유 스냅샷(수량·평단·현재가·평가손익)과 journal 투영을 심볼로 조인 — exit 관리 포지션(이 change 시점: entry 결정 — 편입 표시는 adopt-external-positions가 확장)은 exit 라인(t0·손절 기준선·워터마크·래칫 단계·ladder rung·부분익절·pending exit) 표시, 자격 없는 보유는 "관리 외(미편입)" 구분(라벨 단일화)
- 콘솔에 **거래 이력 화면**: trade_outcomes 동결 값 + positions·exit_states 조인으로 심볼·진입가 표시(스키마에 없는 청산가는 표시하지 않는다 — 재계산 금지)
- **journal RO 접근 신설**: `journal.OpenReadOnly`(additive — 디렉터리·파일 생성 없음, 마이그레이션 없음, `mode=ro`) + 계좌 단위 질의(positions⟕exit_states, 계좌 exit_events, trade_outcomes 조인). **이 change가 `internal/journal` 추가분을 소유**하고, adopt-external-positions는 이 조각 landed 후 착수한다(동시 작업 금지)
- **브로커 스냅샷**: 요청 시 lazy 갱신·백그라운드 폴러 없음, 갱신 1회 = holdings **1콜**(현재가는 응답의 lastPrice 사용 — 심볼별 시세 fan-out 금지), TTL ≥ 15s 캐시
- 콘솔 안전 불변식(루프백 리스너·세션 토큰+재시작 핸드오프 토큰의 2경로 인증·CSRF·프로세스 재기동 외 상태변경 부재·게이트 라우트 부재 — 1.6~1.8 구현의 성문화)을 `operator-console` capability로 성문화

## Non-Goals

- 주문·게이트·설정 조작 라우트 / exit 정책·엔진 동작 변경 / 원격 접근
- 수동 보유에 exit 라인 생성(그것은 adopt-external-positions의 소관)

## Capabilities

### New Capabilities

- `operator-console`: 로컬 웹 콘솔의 안전 불변식(루프백·세션·CSRF·무주문)과 read-only 운영 가시성(포지션·exit 상태·거래 이력·rate budget 보호) 계약

### Modified Capabilities

(없음)

## Impact

- Affected code: `internal/console`(화면·라우트), `internal/journal`(RO open·계좌 질의 — additive만, 스키마 무변경), 브로커 스냅샷용 조회 전용 인터페이스(mutation 메서드 0 — 정적 고정)
- 선행: 2b task 1.8 **커밋 완료** 후 착수(`internal/console` 충돌 방지). 2b 측정·2c와 독립
- 콘솔 브로커 호출은 이 change가 최초 도입 — §0.4 계상을 스펙에 명시(갱신당 1콜·lazy·TTL)
