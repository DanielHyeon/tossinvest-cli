# Function Logic Map: `Console.snapshotFor`

- Source: `internal/console/data.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-us-market`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| run/화면의 시장 | KR 또는 US(zero value = KR) | verifylive.NormalizeMarket | 미지정은 KR로 해석 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 (internal/console/data.go) | 아래 State mutations 참조 | 정상 반환 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 콘솔 판독이 시장 범위를 갖는다 — 스냅샷·기록 경로·재측정 집합·리포트가 시장별이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 콘솔 판독이 시장 범위를 갖는다 — 스냅샷·기록 경로·재측정 집합·리포트가 시장별이다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 한 시장의 기록을 다른 시장 이름으로 읽으면 화면이 측정하지 않은 능력을 측정한 것으로 보여준다. US 경로가 배선되지 않았으면 빈 기록으로 읽는다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
