# Function Logic Map: `Journal.QuarantineExitSnapshot`

- Source: `internal/journal/exit_snapshot.go` (lines 732–783)
- AST evidence: `ast.json` (`source_sha256: b175980ad22e1eb5af7e6bef7185dcdaec7902ccdf6901ef6a1d322f8d8842ea`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 격리 생성(공개 진입점).

## What it does

`quarantineExitSnapshotTx`와 같은 규칙의 공개 진입점. exit loop의 corruption·identity 격리가 여기로 온다. a084는 같은 close-and-reopen 규칙과 각인을 더한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position/generation | 현재 세대와 일치해야 한다 | `positions.instance_seq` | 불일치는 `ErrInvalidRequest` |
| 활성 행 | 세대별 최대 1행 | 같은 트랜잭션 | 현재 개정이면 그대로 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (735) `if` — if id == "" || generation < 0 || why == "" || proof == "" | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (739) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (744) `if` — if err := tx.QueryRowContext(ctx, `SELECT instance_seq FROM positions WHERE id… | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (747) `if` — if actual != generation | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (750) `if` — if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation);… | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (752) `else` — } else if ok | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (752) `if` — } else if ok | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (753) `if` — if !active.NeedsReJudgement() | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (759) `if` — if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSe… | 본문 참조 | — | 아래 Branch Test Map |
| B10 | (767) `if` — if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+… | 본문 참조 | — | 아래 Branch Test Map |
| B11 | (773) `if` — if _, err := tx.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines | 본문 참조 | — | 아래 Branch Test Map |
| B12 | (779) `if` — if err := tx.Commit(); err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 734, 'column': 20}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 734, 'column': 51}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 734, 'column': 78}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 736, 'column': 36}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 738, 'column': 13}, 'text': 'j.db.BeginTx'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 742, 'column': 8}, 'text': 'tx.Rollback'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 744, 'column': 12}, 'text': 'Scan'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 744, 'column': 12}, 'text': 'tx.QueryRowContext'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 748, 'column': 36}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 750, 'column': 24}, 'text': 'activeExitSnapshotQuarantineTx'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 753, 'column': 7}, 'text': 'active.NeedsReJudgement'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 759, 'column': 13}, 'text': 'releaseExitSnapshotQuarantineTx'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 762, 'column': 4}, 'text': 'j.nowString'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 767, 'column': 12}, 'text': 'Scan'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 옛 행 UPDATE(해제) + 새 행 INSERT, 또는 새 행 INSERT만.

## Safety conclusion

- Safe edit boundary: 활성 행 분기와 INSERT의 열 하나.
- High-risk impact: yes — 두 생성 경로가 각인 규칙에서 갈리면 한쪽만 재판정되므로 대칭이어야 한다.
