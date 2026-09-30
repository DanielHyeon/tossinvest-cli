# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (`319`–`358`)
- Qualified: `Context.ExitObserver`
- AST evidence: `ast.json` (`source_sha256` 7638a84b5cddf820…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 6

**역할.** exit 관측 루프를 조립한다. a094: `opts.Critical = c.Notifier`(새 critical 의 단일 입구, 호출자 값과 무관하게 덮음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Notifier` | 알림기 | Context | nil 이면 Critical 도 nil(경고 로그만) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:320` `if c == nil {` | 아니오 |
| B2 | if | `:323` `if !c.Automation.Verified {` | 예 |
| B3 | if | `:327` `if !ok {` | 아니오 |
| B4 | if | `:343` `if opts.Names == nil {` | 예 |
| B5 | if | `:348` `if c.Notifier != nil {` | 예 |
| B6 | if | `:354` `if opts.Floor == nil {` | 예 |

## Calls and live bindings

`obs.RecordOnly` · `exitSideRetrier` · `exitSideFloor` · `NewExitObserver`.

## State mutations and fallbacks

없음(조립).

## Safety conclusion

- B5 안에서 Alerts · Announcer · Critical 셋을 같은 알림기로 덮는다. High-risk: no(배선) — 단 배선 누락은 알림 침묵이므로 시험으로 고정.
