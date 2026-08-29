# Function Logic Map: `TestMigrationV27AddsPairedWeeklyAuthorityWithoutChangingV26Rows`

- Source: `internal/journal/strategy_weekly_reservation_v27_migration_test.go` (lines 13–42)
- AST evidence: `ast.json` (`source_sha256: a32c5b2f62357467f0f60142b9693f0f11290a46edf19f7f70664af939ffcc0d`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 마이그레이션 불변식 테스트. 프로덕션 판정 경로가 아니다.

## What it does

v26 → 현재 마이그레이션이 기존 행을 바꾸지 않는지 본다. a084 이전에는 열 이름까지 해시해서 v26 이전 테이블에 대한 모든 additive 열 추가를 거부했다. 행 비교와 열 단조성 비교로 나눴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이전 스키마 DB | `openJournalAtSchema`가 만든 버전 | 마이그레이션 계획 | 실패는 테스트 실패 |
| 현재 스키마 DB | `Open`이 올린 `SchemaVersion` | `schema.go` | 실패는 테스트 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (18) `if` — if err := old.Close(); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (22) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (26) `if` — if version, err := current.SchemaVersion(context.Background()); err != nil ||  | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (30) `range` — for table, want := range before | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (31) `if` — if after[table] != want | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (36) `range` — for _, name := range []string{"strategy_weekly_reservation_scopes", "strategy_ | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (38) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type=' | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 14, 'column': 10}, 'text': 'filepath.Join'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 14, 'column': 24}, 'text': 't.TempDir'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 15, 'column': 9}, 'text': 'openJournalAtSchema'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 16, 'column': 12}, 'text': 'journalV25RowFingerprints'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 17, 'column': 19}, 'text': 'journalTableColumns'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 17, 'column': 47}, 'text': 'mapKeys'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 18, 'column': 12}, 'text': 'old.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 19, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 21, 'column': 119}, 'text': 'FixedFSProber'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 21, 'column': 18}, 'text': 'Open'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 SQLite 파일만.

## Safety conclusion

- Safe edit boundary: 행 비교는 그대로 두고 열 비교를 분리했다. 마이그레이션 자체와 프로덕션 코드는 건드리지 않는다.
- High-risk impact: no — 다만 이 테스트들이 지키는 것은 원장 스키마 계약이므로, 약화가 아니라 **강화**임을 근거로 남긴다: 전에는 열 집합이 바뀌었다는 것만 알았고 이제는 무엇이 어떻게 바뀌었는지 말한다.
