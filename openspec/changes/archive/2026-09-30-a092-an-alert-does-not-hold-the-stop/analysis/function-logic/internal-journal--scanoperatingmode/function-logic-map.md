# Function Logic Map: `scanOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :713–732, 분기 3 · 반환 4 · 호출 5, source_sha256 `093341ae5773…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): `rowid` 를 `Seq` 로 함께 읽음.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 교체 B1 → B1; B2~B3 → B2~B3(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := row.Scan(&record.Seq, &record.ID, &record.AccountRef, &record.Mode,` (:718) | — | — | (미실행) |
| B2 | `if errors.Is(err, sql.ErrNoRows)` (:720) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B3 | `if err != nil` (:726) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.Is`, `fmt.Errorf`, `parseJournalTime`, `row.Scan` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: Low.
