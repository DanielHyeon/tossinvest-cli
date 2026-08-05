# Function Logic Map: `TestReleasedV25BuildRefusesV26Journal`

- Source: `internal/journal/strategy_first_leg_v26_migration_test.go` (lines 126–136)
- AST evidence: `ast.json` (`source_sha256: 6809fca329344f92e85661a975b2e036aeb75edb0e8444bd6b08f647e8835d96`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 마이그레이션 불변식 테스트. 프로덕션 판정 경로가 아니다.

## What it does

낮은 버전 바이너리가 높은 스키마를 거부하는지 본다. a084는 같은 파일의 helper 변경에 인접해 diff에 걸렸을 뿐 내용은 바뀌지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이전 스키마 DB | `openJournalAtSchema`가 만든 버전 | 마이그레이션 계획 | 실패는 테스트 실패 |
| 현재 스키마 DB | `Open`이 올린 `SchemaVersion` | `schema.go` | 실패는 테스트 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (129) `if` — if err := current.Close(); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (133) `if` — if !errors.Is(err, ErrSchemaTooNew) | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 127, 'column': 10}, 'text': 'filepath.Join'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 127, 'column': 24}, 'text': 't.TempDir'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 128, 'column': 13}, 'text': 'openTestJournalAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 129, 'column': 12}, 'text': 'current.Close'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 130, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 113}, 'text': 'FixedFSProber'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 12}, 'text': 'Open'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 17}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 208}, 'text': 'migrationsThrough'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 132, 'column': 66}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 SQLite 파일만.

## Safety conclusion

- Safe edit boundary: 행 비교는 그대로 두고 열 비교를 분리했다. 마이그레이션 자체와 프로덕션 코드는 건드리지 않는다.
- High-risk impact: no — 다만 이 테스트들이 지키는 것은 원장 스키마 계약이므로, 약화가 아니라 **강화**임을 근거로 남긴다: 전에는 열 집합이 바뀌었다는 것만 알았고 이제는 무엇이 어떻게 바뀌었는지 말한다.
