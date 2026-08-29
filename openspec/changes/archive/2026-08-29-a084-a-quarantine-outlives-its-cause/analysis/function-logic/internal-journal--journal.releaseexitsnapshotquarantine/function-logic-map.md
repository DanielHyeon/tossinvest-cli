# Function Logic Map: `Journal.ReleaseExitSnapshotQuarantine`

- Source: `internal/journal/exit_snapshot.go` (lines 846–871)
- AST evidence: `ast.json` (`source_sha256: b175980ad22e1eb5af7e6bef7185dcdaec7902ccdf6901ef6a1d322f8d8842ea`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 운영자 해제 경로.

## What it does

운영자(a079) 해제. a084는 허용 kind 목록에 `SELECTOR_REVISED`를 더할 뿐, 검증·낙관적 잠금·stale 판정은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| kind | HUMAN_REPAIR / AUTHORITATIVE_RECONCILE / (a084) SELECTOR_REVISED | 호출자 | 그 밖은 `ErrInvalidRequest` |
| expectedVersion | 양수, 활성 행과 일치 | 호출자 | 불일치는 `ErrExitSnapshotReleaseStale` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (849) `if` — if strings.TrimSpace(positionID) == "" || generation < 0 || expectedVersion <=… | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (852) `if` — if kind != QuarantineReleaseHumanRepair && kind != QuarantineReleaseAuthoritat… | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (860) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (864) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (867) `if` — if changed != 1 | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 848, 'column': 9}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 849, 'column': 5}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 849, 'column': 86}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 850, 'column': 10}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 854, 'column': 10}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 856, 'column': 17}, 'text': 'j.db.ExecContext'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 859, 'column': 24}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 859, 'column': 3}, 'text': 'j.nowString'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 859, 'column': 53}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 863, 'column': 18}, 'text': 'result.RowsAffected'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 활성 행 1건 UPDATE.

## Safety conclusion

- Safe edit boundary: kind 검증 조건 하나.
- High-risk impact: no — 새 kind를 받아들일 뿐 해제 조건을 넓히지 않는다. 여전히 정확히 한 활성 행만 닫는다.
