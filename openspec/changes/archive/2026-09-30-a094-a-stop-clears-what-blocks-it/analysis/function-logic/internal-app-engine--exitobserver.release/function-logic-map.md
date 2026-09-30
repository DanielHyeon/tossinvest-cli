# Function Logic Map: `ExitObserver.release`

- Source: `internal/app/engine/exitloop.go` (`1486`–`1491`)
- Qualified: `ExitObserver.release`
- AST evidence: `ast.json` (`source_sha256` 0f943813a3efa423…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 1

**역할.** 제출하지 못했거나 비수용으로 끝난 발의를 푼다 — 판정은 원장 한 곳(`ReleaseUnacceptedExitProposal`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `intentID` | 무장된 발의의 intent | submit | 다르면 원장이 아무것도 안 바꿈 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1487` `if _, _, err := o.opts.Journal.ReleaseUnacceptedExitProposal(ctx, m.position.ID, intentID, how); err != nil {` | 예 |

## Calls and live bindings

`Journal.ReleaseUnacceptedExitProposal`(판정 읽기 + 해제 쓰기 한 트랜잭션).

## State mutations and fallbacks

발의 해제 · exit_events 한 행(판정이 허락할 때).

## Safety conclusion

- 해제 여부를 호출자가 정하지 않는다 — intent 의 attempt 가 전부 비수용이거나 없을 때만. High-risk: yes.
