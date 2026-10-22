# Function Logic Map: `Journal.UnsettledCancelSymbols`

- Source: `internal/journal/flatten.go` (324-346)
- Qualified function: `Journal.UnsettledCancelSymbols`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 3556eecde6278e1c…)
- AST evidence: `ast.json` — AST branches 4, 반환 4, 호출 9
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

flatten 취소 step 중 Pending·InDoubt·Held 인 종목만 센다(쿼리 조건). Failed·Done 은 정착으로 본다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sagaID | string | planLiquidation | B1 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 330:3, 338:4, 343:3, 345:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 329:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 329 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `for at 335:2` | for rows.Next() { | 아래 「State mutations」 | 소스 줄 335 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 337:3` | if err := rows.Scan(&symbol, &state); err != nil { | 아래 「State mutations」 | 소스 줄 337 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 342:2` | if err := rows.Err(); err != nil { | 아래 「State mutations」 | 소스 줄 342 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`db.QueryContext`, `rows.Scan`.

AST 호출 이름 전체: `fmt.Errorf`, `j.db.QueryContext`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `strings.ToUpper`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
