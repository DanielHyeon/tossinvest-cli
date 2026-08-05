# Function Logic Map: `quarantineExitSnapshotTx`

- Source: `internal/journal/exit_snapshot.go` (lines 507–539)
- AST evidence: `ast.json` (`source_sha256: b175980ad22e1eb5af7e6bef7185dcdaec7902ccdf6901ef6a1d322f8d8842ea`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 격리 생성.

## What it does

트랜잭션 안에서 격리를 만든다. a084는 활성 행이 다른 개정에서 왔으면 그것을 `SELECTOR_REVISED`로 닫고 현재 개정으로 각인한 새 행을 연다. 같은 개정이면 지금처럼 기존 행을 그대로 반환한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 활성 행 | 세대별 최대 1행 | `activeExitSnapshotQuarantineTx` | 읽기 실패는 오류 |
| `active.NeedsReJudgement()` | 개정 비교 | 격리 행 + `exitpolicy.RecoverySelectorRevision` | false면 기존 동작 그대로 |
| version | 세대 안에서 단조 증가 | MAX+1 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (509) `if` — if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation);… | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (511) `else` — } else if ok | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (511) `if` — } else if ok | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (512) `if` — if !active.NeedsReJudgement() | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (520) `if` — if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSe… | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (528) `if` — if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+… | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 509, 'column': 24}, 'text': 'activeExitSnapshotQuarantineTx'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 512, 'column': 7}, 'text': 'active.NeedsReJudgement'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 520, 'column': 13}, 'text': 'releaseExitSnapshotQuarantineTx'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 528, 'column': 12}, 'text': 'Scan'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 528, 'column': 12}, 'text': 'tx.QueryRowContext'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 535, 'column': 12}, 'text': 'tx.ExecContext'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 옛 행 UPDATE(해제) + 새 행 INSERT, 또는 새 행 INSERT만. 각인은 항상 현재 개정.

## Safety conclusion

- Safe edit boundary: 활성 행 조기 반환 분기와 INSERT의 열 하나. 사유·evidence·version 규칙은 그대로다.
- High-risk impact: yes — 같은 개정이 같은 세대를 두 번 재판정하지 못하게 하는 종결 조건이 여기 있다.
