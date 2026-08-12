# Branch Test Map: `productionProtectionAssemblies`

> **측정 방법**: `go test -covermode=set -coverprofile`. 분기 *조건*이 아니라 **true 결과
> 본문의 실행 여부**를 측정했다. 담당 테스트는 테스트별 개별 프로파일로 특정했다.

| Branch | Scenario | Test | true 결과 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | (happy path) KR·US 두 assembly가 `Wired: false`로 조립된다 | `internal/app/engine` (435 tests) | **yes** (L39-44 covered) | 분기가 없으므로 경로도 하나뿐이다 |

분기가 0이므로 branch 단위 공백은 없다. a100이 `Wired`를 파라미터화하면 **그 시점에 분기가
생기고**, 그때 각 결과에 대한 RED 테스트가 필요해진다 — `Wired: true`를 낼 수 있는 조건과
낼 수 없는 조건 양쪽.
