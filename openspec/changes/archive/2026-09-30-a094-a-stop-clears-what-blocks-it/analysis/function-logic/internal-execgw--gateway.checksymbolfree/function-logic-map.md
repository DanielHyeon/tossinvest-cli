# Function Logic Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go` (`799`–`830`)
- Qualified: `Gateway.checkSymbolFree`
- AST evidence: `ast.json` (`source_sha256` 96264a796aa63f4a…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 7

**역할.** 같은 종목의 미종결 · park attempt 가 이 mutation 을 막는지. a094: 미종결 판정을 `unsettledFor` 로 추출(exit 청소와 공유).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `plan` | 시장 · 종목 · 노출 증가 여부 | Gateway.submit |  |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:802` `if err != nil {` | 예 |
| B2 | if | `:805` `if len(unsettled) > 0 {` | 예 |
| B3 | if | `:811` `if !plan.raisesExposure {` | 예 |
| B4 | if | `:815` `if err != nil {` | 아니오 |
| B5 | range | `:818` `for _, rec := range unresolved {` | 예 |
| B6 | if | `:820` `if err != nil {` | 아니오 |
| B7 | if | `:823` `if same {` | 아니오 |

## Calls and live bindings

`unsettledFor`(PendingAttempts + attemptTargets) · `UnresolvedAttempts` · `attemptTargets`.

## State mutations and fallbacks

없음(판정).

## Safety conclusion

- 옛 판본은 첫 일치에서 거절했고 새 판본은 끝까지 모은 뒤 첫 것을 이름으로 거절한다 — 뒤 행의 intent 읽기 실패가 거절 대신 오류가 될 수 있다(둘 다 발송 안 함). High-risk: yes.
