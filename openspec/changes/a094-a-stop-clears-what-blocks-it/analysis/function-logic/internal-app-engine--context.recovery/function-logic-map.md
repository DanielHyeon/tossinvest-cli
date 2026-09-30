# Function Logic Map: `Context.Recovery`

- Source: `internal/app/engine/runtime_wiring.go` (`175`–`205`)
- Qualified: `Context.Recovery`
- AST evidence: `ast.json` (`source_sha256` 9906e617304dfdc8…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 5

**역할.** 재시작 복구를 조립한다. a094: Alerts(엔진 알림기)와 CatchUp(기동 따라잡기 클로저)을 호출자 값과 무관하게 덮는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Notifier` | 알림기 | Context | nil 이면 Alerts 없음 · 따라잡기 critical nil |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:176` `if c == nil \|\| c.Journal == nil \|\| c.Resolver == nil \|\| c.Entry == nil {` | 아니오 |
| B2 | if | `:179` `if opts.Clock == nil {` | 예 |
| B3 | if | `:189` `if c.Notifier != nil {` | 예 |
| B4 | if | `:195` `if notifier != nil {` | 아니오 |
| B5 | if | `:201` `if err != nil {` | 아니오 |

## Calls and live bindings

`reconcile.New` · `catchUpExitProposals`(클로저 — 실행은 복구 뒤).

## State mutations and fallbacks

없음(조립).

## Safety conclusion

- 배선 누락은 침묵이므로 역할 시험으로 고정(`CatchUpWired`). High-risk: no(배선).
