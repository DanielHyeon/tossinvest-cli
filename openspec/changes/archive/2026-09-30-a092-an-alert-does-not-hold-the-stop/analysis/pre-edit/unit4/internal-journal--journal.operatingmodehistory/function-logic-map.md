# Function Logic Map: `Journal.OperatingModeHistory`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :593–614, 분기 4 · 반환 4 · 호출 9, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 이력 순서를 rowid 오름차순으로(K14 — 델타 ADDED 「같은 순서 하나」).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:597) | — | — | (미실행) |
| B2 | `for rows.Next()` (:603) | — | — | `TestConcurrentEscalationsConvergeOnTheStrictestMode`, `TestConservativePrecedenceKeepsTheStricterMode` |
| B3 | `if err != nil` (:605) | — | — | (미실행) |
| B4 | `if err := rows.Err(); err != nil` (:610) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append`, `fmt.Errorf`, `j.db.QueryContext`, `rows.Close`, `rows.Err`, `rows.Next`, `scanOperatingMode`, `strings.TrimSpace` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: Low — 읽기 순서.
