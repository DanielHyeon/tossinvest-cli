# Function Logic Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (422-475)
- Function: `Context.runProductionStrategyMarketCycle` in package `engine`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- File SHA-256: `51840da714b49651bee5292a3b51f2814f98b7e2ee5e6996088fb9cceba14d2a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 5.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

조정자가 고른 것이 공유 `strategyDispatchCycle.dispatch` 에 닿는 **유일한** 생산 자리다.
5.5 는 여기서 `!proposal.snapshot.Ready || len(proposal.entries) != 1` 를 지우고
`fresh.proposals.forMarket(market).dispatchHandoff()` 하나로 바꿨다. 5.5-fix 는 그 값을
`Single()` 로 받게 했다. **5.5-fix2 는 그 답을 이 함수에서 없앴다.**

적대 리뷰가 `Single()` 판본을 뚫었기 때문이다: 답을 받아 놓고 `if !handedOff { }` 처럼
아무것도 안 하는 조건에 넣으면 관문은 사라지는데 "답을 썼는지" 보는 검사는 통과했고,
두 스위트가 모두 초록이었다. 지금은 `Deliver` 를 쓴다 — **무시할 수 있는 boolean 이 이
함수에 없다.** 몸통이 도는지 마는지는 경계가 정한다.

**5.5-fix3 은 몸통이 받는 값의 타입을 바꿨다.** fix2 의 판본은 몸통이
`strategyflow.Result` 를 받았고, 그 값을 원본 목록의 것으로 바꿔치는 편집을 `entries`
라는 토큰 금지로 막았다. 3라운드 적대 리뷰어 셋이 각자 그 금지를 우회했다 —
`rawSelection()`, 새 파일의 `relay()`, 몸통 안의 `rawTailProposal()`. **토큰 금지는 네
함수의 본문 안에서만 셌기 때문에 헬퍼 하나를 한 다리 건너 두면 사라졌다.** 그리고 그
금지의 양성 대조 자체가 반례였다: 소비자 넷은 전원 `dispatchHandoff()` 를 통해 이미 한
다리 건너 `entries` 를 읽고 있었다.

지금 몸통이 받는 것은 `strategyhandoff.Delivered` 다. 값 필드가 비공개라서 경계 밖에서는
채울 수 없고, 경계를 지나지 않은 `strategyflow.Result` 를 `dispatch` 에 넘기는 편집은
**컴파일되지 않는다**(뮤테이션 M1 으로 확인). 봉투가 증명하지 못하는 나머지 —
엔진이 스스로 `Admit` 을 불러 봉투를 만드는 경로 — 는
`TestExactlyOneProductionSiteAdmitsIntoTheSeam`(패키지 전체의 `Admit` 호출을 센다)과
`strategyhandoff` 의 `TestOnlyTheEngineImportsThisSeam`(이 경계를 들여오는 패키지를
고정한다)이 함께 막는다.

**5.1.2 가 더한 것 — 여덟 전략군 레인의 사이클.** 이 함수는 이제 새로 고침 뒤,
handoff 앞에서 이 시장의 네 레인을 한 번씩 돌린다(`441:2`). 그 자리에서 도는 일은
`strategyFamilyLaneStep` 하나이고 그 함수는 인자가 `*strategyworker.Lane` 뿐이라
`*Context` 를 들 수 없다 — 오늘 이 함수의 몸통이 `c.Journal.CurrentPositionCampaignCAS`
(`462:15`)와 `fresh.dispatch.dispatch`(`469:12`)를 들고 있는 것과 정확히 대비되는
지점이다. **그 두 호출은 여전히 여기 있고, 그것이 맞다** — 스펙은 변경 권한을
시장 하나에 하나만 두라고 요구하고, 이 함수가 그 하나다. 5.1.2 가 옮긴 것은
*전략군의* 사이클이지 시장의 변경 권한이 아니다.

**반환값이 없는 호출인 이유.** `evaluate` 는 아무것도 돌려주지 않는다. 관측을
돌려주면 이 함수가 그것을 버릴 수 있고, 버린 것은 아무도 못 본다 — 위 fix2 가
`Single()` 의 boolean 에서 배운 것과 같은 답이다. 결과는 레인 런타임 안에 남고
`observations()` 가 읽는다.

**레인 사이클이 새로 고침 잠금 밖인 이유.** 레인은 마감 시한 감시견을 달고 돈다.
`c.strategyRefreshMu`(`481:2`) 안에서 돌리면 레인 하나의 느린 주기가 두 시장의 모든
권한 수집을 함께 세운다. `TestTheMarketCycleRunsItsLanesAndTheRefreshDoesNot` 가
`evaluate` 를 부르는 자리를 패키지 전체에서 세어 이 함수 하나로 고정한다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, **73.8~73.9% of statements** — 3회 측정에 2912·2909·2909 of 3941.
  태그 스위트도 흔들린다. 앞선 판본이 표본 하나(2912)를 안정값으로 적었고, 적대
  리뷰어는 같은 트리에서 2911 을 쟀다. 둘 다 맞고, 둘 다 **표본**이었다.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS,
  **63.5% of statements** — 같은 바이너리·같은 트리에서 다섯 번 재서 2498·2498·2499·2500·2501
  of 3936 이 나왔다. 이 수는 실행마다 흔들린다(무태그 쪽에 스케줄링 의존 시험이 있다).
  앞선 판본이 표본 하나(63.5%)를 안정된 값으로 적은 것을 정정한다.
- 분기의 arm 은 그 분기 위치 **다음에 처음 열리는** 커버리지 블록이다. 조건이 여러 줄이면
  여는 중괄호가 다음 줄에 있어서, 같은 줄만 보던 첫 판본은 이 태스크가 실제로 바꾼 분기를
  `null`(=측정 없음)로 보고했다. "자료 없음"은 "진입 0"과 다르고 그 차이가 요점이다.
- **측정 결과: 두 스위트 어느 쪽도 이 함수에 들어오지 않는다.** 모든 분기 arm 의 진입 수가 0 이다.
  트리 전체에서 이 함수를 부르는 시험이 없다(CodeGraph callers 와 커버리지 프로파일이 같은 답).
  그래서 이 함수에 대한 근거는 실행이 아니라 **소스에 무엇이 쓰여 있는지**뿐이고,
  아래 반증 표의 뮤테이션은 전부 AST 가드가 죽인 것이다.

Exact AST return positions: 425:3, 444:3, 460:2, 464:4, 467:4, 471:4, 473:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 424:2 | arm not entered (양쪽 스위트) — assembly refresh 실패 |
| B2 | if | 443:2 | **5.5-fix2 가 바꾼 분기** — `fresh.dispatch == nil`. handoff 거절은 더 이상 이 조건에 없다(경계 안으로 갔다). arm not entered (양쪽 스위트) |
| B3 | if | 463:3 | arm not entered (양쪽 스위트) — 캠페인 CAS 읽기 실패 |
| B4 | if | 466:3 | arm not entered (양쪽 스위트) — 이미 claim 되었거나 FLAT/CLOSED 아님 |
| B5 | if | 470:3 | arm not entered (양쪽 스위트) — lease 이미 소모 |

이 공백은 이 태스크가 만든 것이 아니다. 이 함수는 `*Context` 와 살아 있는 journal·gateway 를
요구하고, 그 배선은 태스크 5.7(fault injection·race)과 L6 의 몫이다. 여기서는 그 공백을
숨기지 않고 적는다 — 진입 0 은 통과가 아니다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.refreshPairedStrategyEntryProductionAssembly` | 423:16 |
| `evaluate` | 441:2 |
| `c.productionStrategyLanes` | 441:2 |
| `strategyLaneInputs` | 442:3 |
| `fresh.proposals.forMarket` | 442:36 |
| `Deliver` | 460:9 |
| `dispatchHandoff` | 460:9 |
| `fresh.proposals.forMarket` | 460:9 |
| `delivered.Result` | 461:14 |
| `c.Journal.CurrentPositionCampaignCAS` | 462:15 |
| `string` | 462:77 |
| `fresh.dispatch.dispatch` | 469:12 |
| `errors.Is` | 470:6 |

## State mutations and fallbacks

- AST assignments: 4. Defers: 0. Goroutine statements: 0.
- 이 함수 자신은 아무 상태도 쓰지 않는다. 쓰는 일은 전부 `dispatch` 안에서 일어난다.

## Safety conclusion

dispatch 호출 자리는 **정확히 하나**이고 그 자리에 넘어가는 값은 **경계가 묶어 준 이름**
(`Deliver` 몸통의 인자, 또는 `Single()` 의 첫 반환값)이어야 한다. 둘 다
`TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 가 AST 로 고정하고, 그 시험은
`dispatch` 를 메서드 값으로 꺼내 두는 우회(`send := fresh.dispatch.dispatch`)까지 철자 census
로 막는다. 5.5 판본은 인자가 `handoff.result` 인지만 봤으므로 `handoff` 라는 이름의 구조체
리터럴로 우회할 수 있었다. 관문이 완화되지 않았음은 `internal/strategyhandoff` 의 단위 시험이
값으로 지킨다.
