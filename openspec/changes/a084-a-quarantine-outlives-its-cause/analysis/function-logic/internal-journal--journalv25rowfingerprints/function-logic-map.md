# Function Logic Map: `journalV25RowFingerprints`

- Source: `internal/journal/strategy_first_leg_v26_migration_test.go` (lines 181–261)
- AST evidence: `ast.json` (`source_sha256: 6809fca329344f92e85661a975b2e036aeb75edb0e8444bd6b08f647e8835d96`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 마이그레이션 불변식 테스트. 프로덕션 판정 경로가 아니다.

## What it does

마이그레이션 전후로 기존 테이블의 내용이 바뀌지 않았음을 해시로 비교한다. a084는 **열 이름을 해시에서 뺐다** — 열까지 얼리면 마이그레이션 규칙 2(nullable ADD COLUMN 허용)와 모순이고, 실제로 규칙 3이 금지하는 것은 열의 삭제·개명이다. 그 검사는 `assertColumnsOnlyAppended`로 분리했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이전 스키마 DB | `openJournalAtSchema`가 만든 버전 | 마이그레이션 계획 | 실패는 테스트 실패 |
| 현재 스키마 DB | `Open`이 올린 `SchemaVersion` | `schema.go` | 실패는 테스트 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (183) `if` — if tables == nil | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (187) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (190) `for` — for rows.Next() | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (192) `if` — if err := rows.Scan(&table); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (198) `if` — if err := rows.Close(); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (203) `range` — for _, table := range tables | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (206) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B8 | (210) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B9 | (226) `for` — for rows.Next() | 테스트 단언 | — | 아래 Branch Test Map |
| B10 | (229) `range` — for index := range values | 테스트 단언 | — | 아래 Branch Test Map |
| B11 | (232) `if` — if err := rows.Scan(targets...); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B12 | (237) `range` — for _, value := range values | 테스트 단언 | — | 아래 Branch Test Map |
| B13 | (238) `type-switch` — switch typed := value.(type) | 테스트 단언 | — | 아래 Branch Test Map |
| B14 | (239) `case` — case nil: | 테스트 단언 | — | 아래 Branch Test Map |
| B15 | (241) `case` — case int64: | 테스트 단언 | — | 아래 Branch Test Map |
| B16 | (243) `case` — case float64: | 테스트 단언 | — | 아래 Branch Test Map |
| B17 | (245) `case` — case string: | 테스트 단언 | — | 아래 Branch Test Map |
| B18 | (247) `case` — case []byte: | 테스트 단언 | — | 아래 Branch Test Map |
| B19 | (249) `case` — default: | 테스트 단언 | — | 아래 Branch Test Map |
| B20 | (255) `if` — if err := rows.Close(); err != nil | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 182, 'column': 2}, 'text': 't.Helper'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 184, 'column': 16}, 'text': 'j.db.Query'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 188, 'column': 4}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 190, 'column': 7}, 'text': 'rows.Next'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 192, 'column': 14}, 'text': 'rows.Scan'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 193, 'column': 5}, 'text': 'rows.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 194, 'column': 5}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 196, 'column': 13}, 'text': 'append'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 198, 'column': 13}, 'text': 'rows.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 199, 'column': 4}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 SQLite 파일만.

## Safety conclusion

- Safe edit boundary: 행 비교는 그대로 두고 열 비교를 분리했다. 마이그레이션 자체와 프로덕션 코드는 건드리지 않는다.
- High-risk impact: no — 다만 이 테스트들이 지키는 것은 원장 스키마 계약이므로, 약화가 아니라 **강화**임을 근거로 남긴다: 전에는 열 집합이 바뀌었다는 것만 알았고 이제는 무엇이 어떻게 바뀌었는지 말한다.
