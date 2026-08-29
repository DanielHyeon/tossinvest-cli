# Branch Test Map: `Clone`

- Source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`; AST branch locations are authoritative.
- 이 lot 의 편집은 직선 코드 한 항목이며 분기 수를 바꾸지 않았다(편집 전후 1개).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 224:2 — 시장 두 개 복사 | `TestCloneCarriesRuntimeIdentityWithoutSharingIt`, `TestMarketFailureReplacesOnlyExactMarketWithoutFallback` | 예 — B1 을 도는 새 테스트가 "Clone 이 runtime identity 를 떨어뜨렸다"로 실패 | 예 — `go test ./internal/strategyprojection/ -count=1` ok |

## 직선 편집의 반증 (분기가 아니라 값)

| 뮤테이션 | 결과 |
|---|---|
| M1: `Runtime: cloneRuntimeIdentity(...)` 삭제 | KILLED — `TestStrategyRuntimeRESTCarriesTheDigestsTheOperatorMustWriteDown` 실패(REST 응답의 `runtime.configDigest` 가 null). 원복 후 심볼 1개 확인. |
