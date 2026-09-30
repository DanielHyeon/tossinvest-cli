# Function Logic Map: `Journal.CurrentOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :568–577, 분기 1 · 반환 2 · 호출 4, source_sha256 `093341ae5773…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 본문 불변 — `modeLatestOrder` 상수(rowid 내림차순)와 `operatingModeSelect`(rowid 열 추가) 변경으로 순서 · 스냅숏 Seq 가 바뀜.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B1 → B1~B1(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if account == ""` (:570) | — | — | `TestCurrentOperatingModeNeedsAnAccount` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currentModeFromRow`, `fmt.Errorf`, `j.db.QueryRowContext`, `strings.TrimSpace` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(fail-open 방향 제거).
