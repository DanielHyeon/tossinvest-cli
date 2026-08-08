# Function Logic Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go` (`799`–`834`)
- Qualified: `Gateway.checkSymbolFree`
- AST evidence: `ast.json` (`source_sha256` 9601d6562e363a2a…)
- Risk scan: `risk-pattern-report.md`
- 분기 9 · return 8 · 호출 8

**역할.** 이 종목에 미정산 mutation이나 UNRESOLVED attempt가 있는지 보고, 있으면 거절한다. **두 차단의 범위가 다르다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `plan.symbol/market` | 대상 종목 | `mutationPlan` | B4·B9의 매칭 대상 |
| `plan.raisesExposure` | 노출을 늘리는가 (`side == "buy"`, `gateway.go:377`) | `mutationPlan` | **B5가 이것으로 UNRESOLVED 차단을 면제한다** |
| `PendingAttempts` | 미종결 attempt 전수 | 원장 | B2. UNRESOLVED는 제외된다(`resolution.go:89`) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:801` `if err != nil {` | `fmt.Errorf` | :802 | 예 |
| B2 | range | `:804` `for _, rec := range pending {` | `g.attemptTargets` | — | 예 |
| B3 | if | `:806` `if err != nil {` | — | :807 | 아니오 |
| B4 | if | `:809` `if same {` | `reject` | :810 | 예 |
| B5 | if | `:815` `if !plan.raisesExposure {` | `g.journal.UnresolvedAttempts` | :816 | 예 |
| B6 | if | `:819` `if err != nil {` | `fmt.Errorf` | :820 | 아니오 |
| B7 | range | `:822` `for _, rec := range unresolved {` | `g.attemptTargets` | — | 예 |
| B8 | if | `:824` `if err != nil {` | — | :825 | 아니오 |
| B9 | if | `:827` `if same {` | `reject` | :828, :833 | 아니오 |

## Calls and live bindings

`journal.PendingAttempts`(B1 앞) · `attemptTargets`(B2·B7 안) · `journal.UnresolvedAttempts`(B6 앞) · `reject`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 판정만 한다. 게이트 latch는 `Resolver.park`가 따로 건다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** B2–B4에 위험 비증가 면제를 주는 것은 spec의 SHALL(*심볼당 in-flight mutation 1개 제한은 모든 safety class에*)을 깨고, archive `2026-07-26-extend-execution-contract/design.md:63`이 그 carve-out을 이미 검토·폐기했다. a094는 attempt를 **종결시켜서**(R1) B2가 애초에 매칭하지 않게 한다.
- **High-risk impact**: yes — 이 함수가 손절의 통과 여부를 정한다.
