# Function Logic Map: `Clone`

- Source: `internal/strategyprojection/model.go`
- Current source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`
- Signature: `Clone(params=1, results=1)`
- Source range: `221:1`–`228:2`
- AST evidence: `ast.json`, regenerated from the post-edit worktree; AST 분기 1개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 54의 `ConfigDigest`/`BuildDigest` 노출.

## 이 lot 이 무엇을 바꿨는가

`out := Snapshot{...}` 리터럴에 `Runtime: cloneRuntimeIdentity(snapshot.Runtime)` 한 항목이 늘었다.
분기가 아니라 **직선 코드**의 추가다 — AST 분기 수는 편집 전후 모두 1개다.

## 왜 이 편집이 필요한가

`Clone` 은 이 패키지의 **모든 출구**가 지나는 자리다. store 의 `Read`/`Replace`, httpapi router 의
응답, `WithMarketFailure`, 콘솔이 모두 여기를 지난다. 필드를 추가하고 이 함수를 안 고치면 값은
envelope 에 담겼다가 **경계를 넘는 순간 조용히 사라진다.**

## Inputs and invariants

- 입력은 AST signature 그대로다.
- 불변식: 복사본은 원본과 **포인터를 공유하지 않는다.** 소비자가 받은 문자열을 바꿔도 store 안의
  진실이 흔들리면 안 된다 (`cloneString` 과 같은 규칙).
- 이 함수는 값을 판정하지 않는다. 판정은 `Validate` 의 일이다.

## Branches and early returns

- Exact AST return nodes: `227:2`.

| Branch | AST kind | Source location | Edited by this lot | Disposition |
|---|---|---|---|---|
| B1 | range | 224:2 | 아니오 | 시장 두 개를 도는 기존 루프. 이 lot 은 손대지 않았다. |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| cloneRuntimeIdentity | 223:12 |
| make | 223:61 |
| len | 223:95 |
| cloneMarket | 225:25 |

## State mutations and fallbacks

- 새 지도를 만들고 시장 레코드를 복사할 뿐, 입력을 수정하지 않는다. fallback 없음.

## Safety conclusion

- 읽기 전용 관측 자료구조다. 주문·손절·사이징·Guardian·원장 경로에 닿지 않는다.
- 이 편집의 반증: `Runtime:` 항목을 지우면 `TestCloneCarriesRuntimeIdentityWithoutSharingIt` 과
  `TestStrategyRuntimeRESTCarriesTheDigestsTheOperatorMustWriteDown` 이 실패한다 (뮤테이션 M1, 실측).
