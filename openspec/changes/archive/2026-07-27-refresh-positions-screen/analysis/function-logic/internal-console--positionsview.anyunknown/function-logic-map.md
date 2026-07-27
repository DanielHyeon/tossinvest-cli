# Function Logic Map: `positionsView.AnyUnknown`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

change refresh-positions-screen에서 신설된 무상태 leaf 함수다(diff 교차로 evidence 요구). 원장 미판독 상태의 페이지 공지 1회 렌더 여부를 결정한다(design D1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Rows` | []positionRow | `joinPositions` 산출 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 range | `v.Rows` 순회 | 없음 | — | `TestTheUnknownStateNoticeAppearsOncePerPage` |
| B2 if | `r.Unknown()` 참이면 early return true | 없음 | `true`/순회 종료 후 `false` | 같은 테스트 (공지 1회) + journal 정상 케이스(공지 부재) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Unknown()` | 행의 미판독 판정 재사용(정의 1곳) | 순수 | ast.json calls + HEAD |

## State mutations and fallbacks

- 없음(순수 판독). 공지 문장은 템플릿이 소유한다.

## Safety conclusion

- Safe edit boundary: 신규 leaf — 판정 로직은 기존 `Unknown()` 재사용
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)
