# Function Logic Map: `orderedCounts`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L870–879, 분기 2개)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**이다. base 대비 byte 동일하고, 바로 아래에 `sortedSourceIDs`가 삽입되면서
diff hunk가 교차해 evidence가 요구되었다. `ast.json`은 base revision에서 뜬 것이다.

count 맵을 **키가 선언된 순서로** 렌더한다 — 아무도 넘지 않은 밴드도 자기 칸을 차지하도록.
0을 생략하면 "아무도 넘지 않았다"와 "아무도 세지 않았다"가 같은 화면이 되고, 그것이 이
저장소가 네 번째 그림자 밴드에서 한 번 지불한 실패다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `keys` | 선언 순서 | `candidate.BandsFor(code)` 등 | 빈 목록이면 'none' |
| `counts` | 키가 없어도 0 | tally | 맵 miss는 0으로 렌더된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, k := range keys` | `parts` 구성 | — | `cmd/tossctl/candidate_test.go`의 밴드 렌더 |
| B2 | `len(parts) == 0` | 없음 | `"none"` | **커버 없음** — 빈 keys를 넘기는 호출부가 없다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf("%s %d", k, counts[k])` | 칸 하나 | 순수 | ast.json calls (base) |
| `strings.Join(parts, " · ")` | 한 줄 | 순수 | ast.json calls (base) |

## State mutations and fallbacks

- 없음 — 순수 렌더, base와 byte 동일.
- fallback: 빈 목록은 `"none"`. 값을 지어내는 것이 아니라 '칸이 없다'를 말한다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입(`sortedSourceIDs`)만 존재한다.
- High-risk impact: no (렌더 전용).
