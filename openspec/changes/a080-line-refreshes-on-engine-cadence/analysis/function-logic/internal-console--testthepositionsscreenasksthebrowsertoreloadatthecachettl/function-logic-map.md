# Function Logic Map: `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL`

- Source: `internal/console/portfolio_refresh_test.go`
- AST evidence: `ast.json` — **revision: base** (`source_sha256` 6ef7211c1859…)
- Risk scan: `risk-pattern-report.md`
- 이 함수는 **테스트**이고, a080이 이름과 판정 근거를 바꾸므로 diff에 잡혔다.

## base가 판정하던 것

```go
if page := h.page(t, "/positions"); !strings.Contains(page,
	`<meta http-equiv="refresh" content="30">`) {
	t.Error("the positions screen does not ask the browser to reload at the cache TTL")
}
```

주석이 근거를 spec에 두고 있었다 — "Spec `rate budget 보호`: the reload
instruction's period must not undercut the holdings cache TTL". 즉 이 테스트는
**예산 요구사항의 대리 판정**이었고, 리터럴 `30`이 그 대리물이었다.

## Inputs and invariants (base)

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 렌더된 `/positions` | meta refresh = 30 (리터럴) | positions 핸들러 | `t.Error` |
| 렌더된 `/history` | meta refresh 부재 | history 핸들러 | `t.Error` |

## Branches and early returns (base)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `/positions`에 `content="30"` meta refresh가 없다 | 없음 | `t.Error` | 자체 |
| B2 | `/history`가 meta refresh를 갖는다 | 없음 | `t.Error` | 자체 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`seedJournal`·`h.page` | 렌더 경로 | 테스트 하네스 | ast.json |

브로커 호출은 하네스의 counting fake가 받는다. production 상태 mutation 없음.

## a080이 바꾸는 것 — 대리 판정을 직접 판정으로 옮긴다

이 편집의 핵심은 **판정을 지우는 것이 아니라 더 정확한 자리로 옮기는 것**이다.

| 판정 | base | a080 이후 |
|---|---|---|
| 재로드 주기가 존재하고 특정 값이다 | B1 (리터럴 30) | 같은 함수, 기대값이 `lineRefreshSeconds()` |
| **열린 탭의 브로커 비용 상한** | B1이 주기로 대리 | `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`이 호출 수를 직접 센다 |
| `/history`는 자동 재로드가 없다 | B2 | 무변경 |

대리 판정이 부정확했다는 것이 이 change의 근거다 — 상한을 지키는 것은 주기가
아니라 `holdingsCache.get`의 TTL gate이고, 주기를 TTL 아래로 내려도 상한은
유지된다. 변이 6.2에서 그 gate를 제거하자 새 테스트가 5회 허용에 24회 관측으로
RED가 되었다. base의 B1은 그 변이를 잡지 못했을 것이다.

## State mutations and fallbacks

- production 상태 mutation 없음. 브로커 호출은 하네스의 counting fake가 받는다.
- fallback 없음. 두 분기 모두 `t.Error`로 보고하고 계속 진행한다.

## Safety conclusion

- Safe edit boundary: `/positions` 기대 주기의 출처. B2는 무변경.
- High-risk impact: **no.** 테스트 함수다.
- 예산 판정이 무방비로 남지 않았다는 증거: 변이 6.2.
