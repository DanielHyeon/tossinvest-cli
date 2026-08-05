# Function Logic Map: `TestMigrationV26AddsPairedFirstLegAuthorityWithoutChangingV25Rows`

- Source: `internal/journal/strategy_first_leg_v26_migration_test.go` (lines 17–98)
- AST evidence: `ast.json` (`source_sha256: 6809fca329344f92e85661a975b2e036aeb75edb0e8444bd6b08f647e8835d96`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 마이그레이션 불변식 테스트. 프로덕션 판정 경로가 아니다.

## What it does

v25 → 현재 마이그레이션이 기존 행을 바꾸지 않는지 본다. a084는 열 단조성 단언을 추가했다. 행 단언은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이전 스키마 DB | `openJournalAtSchema`가 만든 버전 | 마이그레이션 계획 | 실패는 테스트 실패 |
| 현재 스키마 DB | `Open`이 올린 `SchemaVersion` | `schema.go` | 실패는 테스트 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (21) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (26) `if` — if err := insertStrategyDispatchLease(old, kr, StrategyDispatchLeaseClaimed, " | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (29) `if` — if err := insertStrategyDispatchLease(old, us, StrategyDispatchLeaseIssued, "" | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (32) `if` — if _, err := old.db.Exec(`INSERT INTO strategy_dispatch_outcomes( | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (41) `if` — if err := old.Close(); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (47) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (51) `if` — if version, err := current.SchemaVersion(context.Background()); err != nil ||  | 테스트 단언 | — | 아래 Branch Test Map |
| B8 | (55) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_market_a | 테스트 단언 | — | 아래 Branch Test Map |
| B9 | (61) `range` — for table, target := range map[string]*int | 테스트 단언 | — | 아래 Branch Test Map |
| B10 | (66) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(target); e | 테스트 단언 | — | 아래 Branch Test Map |
| B11 | (70) `if` — if leases != 2 || outcomes != 1 || qFinal != 2 || ownerEpochs != 1 || currentO | 테스트 단언 | — | 아래 Branch Test Map |
| B12 | (76) `if` — if err := current.db.QueryRow(`SELECT state,revision FROM strategy_dispatch_le | 테스트 단언 | — | 아래 Branch Test Map |
| B13 | (81) `range` — for table, want := range beforeRows | 테스트 단언 | — | 아래 Branch Test Map |
| B14 | (82) `if` — if got := afterRows[table]; got != want | 테스트 단언 | — | 아래 Branch Test Map |
| B15 | (86) `range` — for _, object := range []struct{ kind, name string } | 테스트 단언 | — | 아래 Branch Test Map |
| B16 | (94) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type=? | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 18, 'column': 10}, 'text': 'filepath.Join'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 18, 'column': 24}, 'text': 't.TempDir'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 19, 'column': 9}, 'text': 'openJournalAtSchema'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 20, 'column': 16}, 'text': 'old.AcquireStrategyDispatchOwner'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 20, 'column': 49}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 22, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 24, 'column': 8}, 'text': 'prepareStrategyDispatchLease'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 25, 'column': 8}, 'text': 'prepareStrategyDispatchLease'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 26, 'column': 12}, 'text': 'insertStrategyDispatchLease'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 27, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 SQLite 파일만.

## Safety conclusion

- Safe edit boundary: 행 비교는 그대로 두고 열 비교를 분리했다. 마이그레이션 자체와 프로덕션 코드는 건드리지 않는다.
- High-risk impact: no — 다만 이 테스트들이 지키는 것은 원장 스키마 계약이므로, 약화가 아니라 **강화**임을 근거로 남긴다: 전에는 열 집합이 바뀌었다는 것만 알았고 이제는 무엇이 어떻게 바뀌었는지 말한다.
