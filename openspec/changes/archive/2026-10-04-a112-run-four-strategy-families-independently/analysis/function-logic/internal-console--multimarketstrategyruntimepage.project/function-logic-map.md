# Function Logic Map: `multiMarketStrategyRuntimePage.project`

- Source: `internal/console/strategy_runtime_multimarket.go`
- Current source SHA-256: `dd84f3074d2fd14f67c369050bb42e78c55403b044d36c1aa4e45168acb155f5`
- Signature: `multiMarketStrategyRuntimePage.project(params=1, results=0)`
- Source range: `66:1`–`92:2`
- AST evidence: `ast.json`, regenerated from the post-edit worktree; AST 분기 2개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 54의 `ConfigDigest`/`BuildDigest` 노출.

## 이 lot 이 무엇을 바꿨는가

루프 **앞**의 직선 구간에 두 줄이 늘었다: `page.ConfigDigest`, `page.BuildDigest` 를
`projectionValue(snapshot.Runtime.…)` 로 채운다. 분기는 편집 전후 모두 2개다.

## 왜 시장 카드가 아니라 페이지 상단인가

두 digest 는 시장별 사실이 아니라 **엔진 프로세스 하나의 사실**이다. 그리고 운영자가 이 숫자를
찾는 순간은 두 시장이 다 UNKNOWN 일 때다 — 시장 카드 안에 뒀다면 정확히 그때 사라진다.
그래서 `SchemaVersion`/`GeneratedAt` 과 같은 자리, 같은 층에 둔다.

## 콘솔이 스스로 계산하면 안 되는 이유

같은 저장소의 상수로 만드는 값이라 콘솔 바이너리도 같은 값을 계산할 수 있다. 하지만 콘솔과
엔진의 build 가 다르면 그렇게 만든 숫자는 **엔진이 거절할 매니페스트**를 낳는다. 그래서 이
함수는 `snapshot` 이 실어 온 값만 쓰고, 없으면 `projectionValue` 가 `not_observed` 를 준다.

## Inputs and invariants

- 입력은 AST signature 그대로다.
- 불변식: 이 함수는 화면 문자열만 만든다. 설정·주문·활성화 능력이 없다.

## Branches and early returns

- Exact AST return nodes: 없음(값을 돌려주지 않는 함수).

| Branch | AST kind | Source location | Edited by this lot | Disposition |
|---|---|---|---|---|
| B1 | range | 75:2 | 아니오 | 시장 카드 루프. 기존 그대로. |
| B2 | if | 87:3 | 아니오 | 시장 오류 코드 표시. 기존 그대로. |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| runtimeProjectionTime | 68:21 |
| projectionValue | 71:22 |
| projectionValue | 72:21 |
| strategyprojection.Registry | 73:16 |
| make | 74:17 |
| strategyprojection.OrderedMarkets | 75:23 |
| string | 76:50 |
| string | 76:79 |
| runtimeProjectionClass | 76:113 |
| string | 76:136 |
| projectionValue | 77:12 |
| projectionValue | 77:56 |
| string | 77:105 |
| string | 77:147 |
| projectionValue | 78:16 |
| projectionValue | 78:67 |
| string | 78:125 |
| projectionValue | 79:16 |
| projectionValue | 79:58 |
| projectionValue | 79:108 |
| projectionValue | 80:23 |
| string | 80:84 |
| string | 81:22 |
| string | 81:74 |
| projectionValue | 82:20 |
| projectionValue | 82:85 |
| string | 82:153 |
| string | 83:23 |
| string | 83:77 |
| projectionValue | 83:130 |
| string | 84:22 |
| string | 84:75 |
| string | 84:129 |
| string | 85:26 |
| string | 85:85 |
| string | 86:18 |
| runtimeProjectionTimePointer | 86:57 |
| string | 88:21 |
| append | 90:18 |

## State mutations and fallbacks

- page 구조체 필드만 채운다. 없는 관측은 `not_observed` 로 표시하며 값을 지어내지 않는다.

## Safety conclusion

- 읽기 전용 화면이다. `TestStrategyRuntimePageShowsTheDigestsTheOperatorMustWriteDown` 이 같은
  요청에서 broker mutation 0을 함께 잰다.

a112 게이트 위생(2026-10-01): STALE 재측정 — 줄 이동(본문 불변)
