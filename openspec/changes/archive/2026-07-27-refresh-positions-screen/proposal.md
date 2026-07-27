# Change: refresh-positions-screen

## Why

사용자 보고(2026-07-27): 콘솔 `/positions` 화면에서 ① "엔진 원장에 이 종목의 포지션이 없다 —
…" 안내가 수동 보유 **종목 수만큼 반복**되어 읽기를 방해하고, ② 화면을 열어 둔 동안 **보유
종목이 갱신되지 않는다**. 둘 다 add-operator-dashboard가 의도한 설계였다 — 사유는 행마다
렌더되고(`positionRow.Reason`), 대시보드 화면은 자동 재로드를 두지 않았다
(`positionsPage.Refresh() == false`, "refreshing is a decision rather than a default").
실사용에서 그 결정이 틀렸음이 확인되었으므로 화면 동작을 바꾼다.

## What Changes

- **안내 dedup**: 보유 행 전체에 공통인 사유 2종 — 원장을 읽지 못함(관리 여부 불명), 원장에
  포지션 없음(관리 외) — 은 행 반복을 멈추고 **페이지 수준 1회 안내**로 옮긴다. 행별로 다른
  사유 2종 — 자격 기록 없는 원장 포지션, 자격은 있으나 exit 미개설 — 은 행에 남는다.
  "관리 외(미편입)" 라벨은 지금처럼 **행마다** 유지한다(스펙 라벨 SHALL 불변).
- **자동 갱신**: positions 화면에 `meta http-equiv="refresh"` 주기 재로드를 추가한다. 주기는
  서버 holdings 캐시 TTL(30초)과 같다. 서버는 여전히 요청 시 lazy 갱신 + TTL 상한이므로,
  열린 탭 하나의 브로커 비용 상한은 **holdings 1콜/TTL**이고, 검증 실행 중에는 기존 보류
  (hold)가 그대로 적용되어 재로드가 브로커 콜을 만들지 않는다. 서버측 백그라운드 폴러는
  여전히 없다. 기존 검증/대시보드 화면의 2초 재로드는 그대로 둔다(head 템플릿 주기
  파라미터화).

## Non-Goals

- 편입 버튼·종목별 opt-in 등 상태 변경 UI(콘솔 read-only 불변식 무접촉 — 편입은 엔진 대사
  루프의 몫)
- history 화면 자동 갱신(동결 값 표시라 요구가 없다)
- holdings TTL·rate budget 수치 변경, JavaScript 도입(무스크립트 규칙 유지)

## Capabilities

### Modified Capabilities

- `operator-console`: "포지션 가시성" — 공통 사유의 페이지 수준 1회 안내 허용(행별 반복
  비요구), "rate budget 보호" — TTL 이상 주기의 브라우저 재로드 지시 허용(비용 불변식 유지)

## Impact

- Affected code: `internal/console`만 — templates_portfolio.go(안내 위치), portfolio.go
  (`Reason` 분기 축소 + view 수준 공지 도우미), portfolio_pages.go(`Refresh` true),
  templates.go(head 주기 파라미터화), pages.go(기존 화면 `RefreshSeconds` 2 유지),
  holdings.go 주석 정합, 테스트
- High-risk 경로 무접촉: 주문·위험·원장·인증 코드 없음. 콘솔은 read-only이며 라우트 표 불변
- PM: 활성 product change의 기존 관례에 따라 registry `bootstrap_change_allowlist`에 등재
  (활성 스토리 체계는 SDD 이관 epic뿐이라 1:1 story가 없다 — archive 시 allowlist에서 제거)
