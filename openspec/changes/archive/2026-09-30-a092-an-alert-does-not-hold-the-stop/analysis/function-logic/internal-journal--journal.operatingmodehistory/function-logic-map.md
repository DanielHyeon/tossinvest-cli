# Function Logic Map: `Journal.OperatingModeHistory`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :609–630, 분기 4 · 반환 4 · 호출 9, source_sha256 `093341ae5773…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 이력 순서 `ORDER BY rowid`(K14).
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B4 → B1~B4(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:613) | — | — | (미실행) |
| B2 | `for rows.Next()` (:619) | — | — | `TestConcurrentEscalationsConvergeOnTheStrictestMode`, `TestConservativePrecedenceKeepsTheStricterMode` |
| B3 | `if err != nil` (:621) | — | — | (미실행) |
| B4 | `if err := rows.Err(); err != nil` (:626) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append`, `fmt.Errorf`, `j.db.QueryContext`, `rows.Close`, `rows.Err`, `rows.Next`, `scanOperatingMode`, `strings.TrimSpace` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: Low.
