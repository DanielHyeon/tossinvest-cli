# Branch Test Map: `Validate`

- Source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`; AST branch locations are authoritative.
- B2 만 이 lot 이 추가했다. 나머지 넷은 기존 분기이며 기존 테스트가 계속 덮는다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 267:2 — envelope 3항 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` ("missing US") | 아니오 — 기존 분기, 이 lot 이 안 바꿨다 | 예 |
| B2 | if at 270:2 — runtime identity (**신규**) | `TestValidationRejectsPartialOrNoncanonicalRuntimeIdentity` | 예 — 4개 하위 케이스(config only/build only/noncanonical config/short build)가 "반쪽이거나 비정규인 runtime identity 가 통과했다"로 실패 | 예 |
| B3 | range at 273:2 — KR·US | `TestDormantSnapshotContainsExactPairedHonestMarkets` | 아니오 — 기존 분기 | 예 |
| B4 | if at 275:3 — 시장 부재/교차 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` ("cross-market record") | 아니오 — 기존 분기 | 예 |
| B5 | if at 278:3 — 시장 레코드 판정 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` ("zero market fallback", "third readiness") | 아니오 — 기존 분기 | 예 |

## B2 가 기존 스냅샷을 새로 거절하지 않음의 실측

`TestDormantSnapshotHasNoRuntimeIdentity` 가 dormant·unavailable 두 스냅샷(두 필드 모두 nil)이
계속 유효함을 잰다. 즉 이 분기는 **거절 집합을 늘리기만** 하고 기존 통과 집합을 줄이지 않는다.
