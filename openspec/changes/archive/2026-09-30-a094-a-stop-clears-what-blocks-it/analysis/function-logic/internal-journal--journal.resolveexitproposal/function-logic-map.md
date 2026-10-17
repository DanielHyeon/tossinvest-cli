# Function Logic Map: `Journal.ResolveExitProposal`

- Source: `internal/journal/apply_hook.go` (`829`–`852`)
- Qualified: `Journal.ResolveExitProposal`
- AST evidence: `ast.json` (`source_sha256` 8daed316f087979f…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 5

**역할.** 판정 없이 발의를 비우는 쓰기 — 이제 기대 intent 필수, 쓰기는 `clearExitProposalTx` 한 곳.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expectedIntentID` | 무장된 발의의 intent | 호출자 | 비면 ErrInvalidRequest(B1), 다르면 무변화 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:832` `if strings.TrimSpace(expectedIntentID) == "" {` | 예 |
| B2 | if | `:836` `if err != nil {` | 아니오 |
| B3 | if | `:840` `if err != nil {` | 아니오 |
| B4 | if | `:845` `if err != nil \|\| !released {` | 예 |
| B5 | if | `:848` `if err := tx.Commit(); err != nil {` | 예 |

## Calls and live bindings

`proposalResolutionAction` · `clearExitProposalTx`(기대 intent 대조 · rung 되돌림 · exit_events).

## State mutations and fallbacks

exit_states 발의 컬럼 NULL · exit_events 한 행 · (사다리) active_rung.

## Safety conclusion

- 늦게 도착한 해제가 다른 발의를 지우지 않는다(4.3a). 손절 가격 컬럼 무접촉(4.5). High-risk: yes.
