# Function Logic Map: `activeExitSnapshotQuarantineTx`

- Source: `internal/journal/exit_snapshot.go` (lines 785–800)
- AST evidence: `ast.json` (`source_sha256: b175980ad22e1eb5af7e6bef7185dcdaec7902ccdf6901ef6a1d322f8d8842ea`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 재판정 여부의 입력.

## What it does

세대의 활성 격리를 읽는다. a084는 `selector_revision`을 함께 읽고 NULL을 0(미기록)으로 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selector_revision | INTEGER NULL | `exit_snapshot_quarantines` | NULL은 0 — 현재 개정으로 만들어내지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (793) `if` — if errors.Is(err, sql.ErrNoRows) | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 790, 'column': 9}, 'text': 'Scan'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 790, 'column': 9}, 'text': 'tx.QueryRowContext'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 793, 'column': 5}, 'text': 'errors.Is'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 없음 (읽기).

## Safety conclusion

- Safe edit boundary: SELECT 열 하나와 NULL 처리.
- High-risk impact: yes — NULL을 현재 개정으로 읽으면 이 change가 고치려는 세 행이 영원히 재판정되지 않는다.
