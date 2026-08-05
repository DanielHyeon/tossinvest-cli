# Function Logic Map: `TestTheOverviewReloadsAtTheCacheTTL`

- Source: `internal/console/overview_test.go`
- AST evidence: `ast.json` — **revision: base** (`source_sha256` 5fbdb9476eaa…)
- Risk scan: `risk-pattern-report.md`
- 이 함수는 **테스트**이고, a080이 이름과 판정 근거를 바꾸므로 diff에 잡혔다.
  base 리비전의 본문을 기록한다 — 무엇을 바꾸는지 알고 바꿨다는 증거다.

## base가 판정하던 것

```go
want := int(holdingsTTL / time.Second)
if got := (overviewPage{}).RefreshSeconds(); got != want {
	t.Errorf("the overview reloads every %ds, want %ds derived from holdingsTTL", got, want)
}
```

실패 문구가 "**derived from holdingsTTL**"이다. 이 테스트의 판정 대상은 값이
아니라 **파생 관계 자체**였다. a080이 끊으려는 것이 정확히 그 관계이므로 이
테스트는 갱신 대상이 될 수밖에 없다.

## Inputs and invariants (base)

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `holdingsTTL` | 30초 | `holdings.go` | 기대값의 출처 |
| `overviewPage{}.RefreshSeconds()` | `holdingsTTL`과 같아야 함 | `overview.go` | `t.Errorf` |
| 렌더된 `/dashboard` | meta refresh 존재·값 일치 | overview 핸들러 | `t.Error` |

## Branches and early returns (base)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `RefreshSeconds() != holdingsTTL/초` | 없음 | `t.Errorf` | 자체 |
| B2 | 렌더에 `http-equiv="refresh"` 없음 | 없음 | `t.Error` | 자체 |
| B3 | 렌더된 주기 값이 기대와 다름 | 없음 | `t.Errorf` | 자체 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newOverviewHarness`·`seedJournal`·`h.page` | 렌더 경로 | 테스트 하네스 | ast.json |

live config binding 없음. production 상태 mutation 없음.

## a080이 바꾸는 것

이름을 `TestTheOverviewReloadsAtTheEngineCadence`로 바꾸고, 기대값의 출처를
`holdingsTTL`에서 `lineRefreshSeconds()`로 옮긴다. B2·B3의 렌더 판정은 그대로
유지하고, **분기를 하나 더한다** — 새 상수가 `holdingsTTL`과 같아지면 실패한다.
값 일치만 보는 테스트는 누군가 다시 파생시켜도 통과하기 때문이다.

즉 판정이 약해지지 않고 **하나 늘어난다**.

## State mutations and fallbacks

- production 상태 mutation 없음. 테스트 하네스의 임시 디렉터리와 fake 시계만 쓴다.
- fallback 없음. 실패는 전부 `t.Error`/`t.Errorf`로 보고되고 조기 반환하지 않는다.

## Safety conclusion

- Safe edit boundary: 기대값의 출처와 분기 1개 추가. 렌더 판정은 무변경.
- High-risk impact: **no.** 테스트 함수다.
