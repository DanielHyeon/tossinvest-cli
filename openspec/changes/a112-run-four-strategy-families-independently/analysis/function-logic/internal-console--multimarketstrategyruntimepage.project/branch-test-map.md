# Branch Test Map: `multiMarketStrategyRuntimePage.project`

- Source SHA-256: `11b5fff5a3b5eb90a71c7cda8176666fc6ed583263c62292d3446082d79f3417`; AST branch locations are authoritative.
- 이 lot 의 편집은 루프 앞 직선 두 줄이며 분기 수를 바꾸지 않았다(편집 전후 2개).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 71:2 — 시장 카드 두 장 | `TestStrategyRuntimeMarketsRenderIndependently` | 아니오 — 기존 분기 | 예 |
| B2 | if at 83:3 — 시장 오류 코드 표시 | `TestStrategyRuntimeMarketsRenderIndependently` (US 가 RUNTIME_UNAVAILABLE) | 아니오 — 기존 분기 | 예 |

## 직선 편집의 반증 (분기가 아니라 화면 값)

| 시나리오 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 엔진이 보낸 digest 가 화면에 보인다 | `TestStrategyRuntimePageShowsTheDigestsTheOperatorMustWriteDown` | 예 — "Config digest"/"Build digest"/두 값 모두 페이지에 없음 | 예 |
| 엔진이 없으면 지어내지 않는다 | `TestStrategyRuntimePageDoesNotInventDigestsWhenTheEngineIsAbsent` | 예 — 두 항목의 자리 자체가 없음 | 예 |

## 뮤테이션 실측

| 뮤테이션 | 결과 |
|---|---|
| M3: `page.ConfigDigest` 를 `projectionValue(snapshot.Runtime.ConfigDigest)` 대신 상수로 채운다 (콘솔이 값을 지어내는 모양) | KILLED — 두 콘솔 테스트가 모두 실패. 부재 테스트는 렌더된 짝 `<dt>Config digest</dt><dd><code>not_observed</code></dd>` 를 그대로 찾으므로, 페이지 어딘가의 다른 `not_observed` 가 대신 통과시켜 주지 못한다. |
