# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go` (78-397)
- Function: `TestSchemaTablesAndColumns` in package `journal`
- Signature: `TestSchemaTablesAndColumns(params=1, results=0)`
- File SHA-256: `73f2ca58ca5f0e263ed3db208f35318607dc45a74b67149911237a8abf1504e0`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

이 시험은 원장 스키마의 **얼린 열거표**다. 테이블 이름 전부, 그리고 고른 테이블의
칼럼 전부를 목록으로 들고 있으면서 실제 `sqlite_master` 와 대조한다.

L5 5.3.3 이 이 함수의 **본문**을 편집했다 — 테이블 목록에 두 줄이 늘었다
(`strategy_lane_latches`, `strategy_lane_latch_recoveries`). 목록 편집이 곧 본문 편집이므로
FLM 이 필요하다.

**목록을 고친 것이 이 시험을 약화시키지 않는 이유.** 이 열거표의 일은 "새 테이블이
조용히 생기지 않게" 하는 것이고, 실제로 그렇게 작동했다: v32 마이그레이션을 넣자마자
이 시험이 빨개졌고, 그래서 새 테이블 둘이 사람 손으로 선언됐다. 목록을 고치지 않고
통과시키는 방법은 없다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` 는 `_test.go` 를 계측하지 않으므로
  이 함수를 말해 줄 커버리지 프로파일이 없다. 아래는 소스 텍스트로 분류한 것이다.
- 실행 근거: `go test -count=1 -run TestSchemaTablesAndColumns ./internal/journal/` — PASS
  (2026-09-03). 편집 **전**에는 같은 명령이 실패했고, 실패 문구가 새 테이블 둘을 이름으로
  가리켰다.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 165:2 | 시험이 PASS 이므로 진입하지 않는 arm(질의 실패) |
| B2 | for | 169:2 | 진입 — 실제 테이블 이름을 읽는다 |
| B3 | if | 171:3 | 시험이 PASS 이므로 진입하지 않는 arm(스캔 실패) |
| B4 | if | 176:2 | 시험이 PASS 이므로 진입하지 않는 arm(rows.Err) |
| B5 | if | 180:2 | **이 태스크가 재는 분기** — 얼린 목록과 실제가 다르면 실패. 편집 전 진입(RED), 편집 후 미진입(GREEN) |
| B6 | for | 390:2 | 진입 — 칼럼 대조 |
| B7 | if | 393:3 | 시험이 PASS 이므로 진입하지 않는 arm |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `openTestJournal` | 79:7 |
| `context.Background` | 80:9 |
| `j.db.QueryContext` | 163:15 |
| `t.Fatal` | 166:3 |
| `rows.Next` | 169:6 |
| `rows.Scan` | 171:13 |
| `t.Fatal` | 172:4 |
| `append` | 174:15 |
| `rows.Err` | 176:12 |
| `t.Fatal` | 177:3 |
| `rows.Close` | 179:2 |
| `strings.Join` | 180:5 |
| `strings.Join` | 180:37 |
| `t.Fatalf` | 181:3 |
| `tableColumns` | 391:10 |
| `sort.Strings` | 392:3 |
| `strings.Join` | 393:6 |
| `strings.Join` | 393:32 |
| `t.Errorf` | 394:4 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.
- 이 함수는 아무 생산 상태도 쓰지 않는다. 읽기 전용 대조다.

## Safety conclusion

- 이 편집은 열거표에 두 줄을 더한 것뿐이고, 그 둘은 이 change 가 만든 append-only
  테이블이다. 기존 테이블·칼럼 목록은 한 줄도 바뀌지 않았다.
- 이 시험이 빨개진 것이 v32 마이그레이션의 첫 외부 신호였다. 열거표가 일한 것이다.
