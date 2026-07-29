# Function Logic Map: `mutationNote`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:mutationNote`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change가 추가한 leaf. 운영자가 게이트를 켜도 되는지 판단할 때 읽는 문장 — 어떤 mutation이 덮였고 어떤 것이 남았는지.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| accepted []attest.Proof | 실린 증거 | acceptSupervised | 비어 있으면 '덮이지 않음' 문장 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 303) — `for _, p := range accepted {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `range` (line 307) — `for _, e := range LiveOnlyEndpoints() {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 308) — `if have[normaliseEndpoint(e)] {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `else` (line 310) — `} else {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `switch` (line 314) — `switch {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `case` (line 315) — `case len(missing) == 0:` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `case` (line 318) — `case len(covered) == 0:` | 없음 | 루프/분기 계속 | 아래 참조 |
| B8 | `case` (line 321) — `default:` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseEndpoint` | ast.json calls (line 304) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `LiveOnlyEndpoints` | ast.json calls (line 307) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `append` | ast.json calls (line 309) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `len` | ast.json calls (line 315) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Sprintf` | ast.json calls (line 316) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Join` | ast.json calls (line 317) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수.

## Safety conclusion

- Safe edit boundary: 신규 함수. 문자열만 만든다.
- High-risk impact: no — 문서화. 단 이 문장이 틀리면 운영자가 오판한다.
