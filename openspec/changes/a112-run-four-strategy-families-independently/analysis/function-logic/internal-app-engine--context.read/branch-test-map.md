# Branch Test Map: `Context.Read`

- Source SHA-256: `14ec90c888e64ccb7e45d5823f415cbf53a1b97b4a62adf9b476db478892f80a`; AST branch locations are authoritative.
- **이 lot 전까지 이 함수에는 어떤 테스트도 없었다.** 아래 네 테스트가 첫 실행이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 24:2 — nil receiver/context | `TestStrategyRuntimeReadWithoutAStoreStaysAnError` | 아니오 — 기존 동작을 처음으로 고정한 것이며 이 lot 이 바꾸지 않았다 | 예 |
| B2 | if at 30:2 — store 부재 | `TestStrategyRuntimeReadWithoutAStoreStaysAnError` | 아니오 — 같은 이유 | 예 |
| B3 | if at 34:2 — store 읽기 오류 (**조건이 좁아짐**) | `TestStrategyRuntimeReadOnAFailedStoreInventsNothing` | 아니오 — 편집 뒤에 쓴 테스트다. 대신 뮤테이션으로 반증한다(아래) | 예 |
| B4 | if at 50:2 — supervisor 부재 (**신규**) | `TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest` | 예 — "운영자가 적어야 하는 숫자가 여전히 밖으로 나오지 않는다"로 실패 | 예 |
| B5 | range at 53:2 — KR·US latch 검사 | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` | 아니오 — 편집 뒤에 쓴 테스트 | 예 |
| B6 | if at 55:3 — latch 안 된 시장 건너뛰기 | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` (US 는 latch 아님) | 아니오 — 같은 이유 | 예 |
| B7 | if at 59:3 — CURRENT 인 시장만 덮기 | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` (KR 이 CURRENT) | 아니오 — 같은 이유 | 예 |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M2: `WithRuntimeIdentity` 의 config/build 인자를 맞바꾼다 | KILLED — `TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest` 와 `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` 둘 다 실패. 원복 후 심볼 1개 확인. |
| M4: `projectionDigest(...)` 로 감싸 `sha256:` 접두사를 다시 벗긴다 (독립 리뷰가 찾은 그 결함) | KILLED — 같은 두 테스트가 실패. 형식 단언이 생산자의 변환을 되풀이하지 않고 소비자가 요구하는 모양을 직접 적기 때문이다. |
| M1(다른 번들): `Clone` 이 identity 를 떨어뜨린다 | KILLED — B5–B7 이 지나는 `WithMarketFailure` 가 `Clone` 을 거치므로 latch 경로도 이 뮤테이션에 죽는다. |

RED 없이 GREEN 만 있는 행(B1·B2·B3·B5·B6·B7)은 **이 lot 이 그 분기의 동작을 바꾸지 않았고**,
바뀐 두 곳(B3 의 조건 축소, B4 의 신설)은 위 뮤테이션으로 반증했다.
