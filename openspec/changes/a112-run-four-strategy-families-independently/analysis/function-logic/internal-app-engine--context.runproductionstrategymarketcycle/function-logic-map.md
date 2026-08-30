# Function Logic Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (422-445)
- Function: `Context.runProductionStrategyMarketCycle` in package `engine`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- File SHA-256: `17ad4c0c684b74686dd1e80b256a06971802afa26bcfd300dbeac9bd5f7e0496`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 5.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

조정자가 고른 것이 공유 `strategyDispatchCycle.dispatch` 에 닿는 **유일한** 생산 자리다.
5.5 는 여기서 `!proposal.snapshot.Ready || len(proposal.entries) != 1` 를 지우고
`fresh.proposals.forMarket(market).dispatchHandoff()` 하나로 바꿨다. dispatch 에 넘기는 값도
`handoff.result` 로 바뀌어, 넘어가는 것과 관문을 통과한 것이 같은 값임이 형태로 드러난다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.9% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 63.5% of statements.
- 분기의 arm 은 그 분기 위치 **다음에 처음 열리는** 커버리지 블록이다. 조건이 여러 줄이면
  여는 중괄호가 다음 줄에 있어서, 같은 줄만 보던 첫 판본은 이 태스크가 실제로 바꾼 분기를
  `null`(=측정 없음)로 보고했다. "자료 없음"은 "진입 0"과 다르고 그 차이가 요점이다.
- **측정 결과: 두 스위트 어느 쪽도 이 함수에 들어오지 않는다.** 모든 분기 arm 의 진입 수가 0 이다.
  트리 전체에서 이 함수를 부르는 시험이 없다(CodeGraph callers 와 커버리지 프로파일이 같은 답).
  그래서 이 함수에 대한 근거는 실행이 아니라 **소스에 무엇이 쓰여 있는지**뿐이고,
  아래 반증 표의 뮤테이션은 전부 AST 가드가 죽인 것이다.

Exact AST return positions: 425:3, 430:3, 435:3, 438:3, 442:3, 444:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 424:2 | arm not entered (양쪽 스위트) — assembly refresh 실패 |
| B2 | if | 429:2 | **이 태스크가 바꾼 분기.** arm not entered (양쪽 스위트) |
| B3 | if | 434:2 | arm not entered (양쪽 스위트) — 캠페인 CAS 읽기 실패 |
| B4 | if | 437:2 | arm not entered (양쪽 스위트) — 이미 claim 되었거나 FLAT/CLOSED 아님 |
| B5 | if | 441:2 | arm not entered (양쪽 스위트) — lease 이미 소모 |

이 공백은 이 태스크가 만든 것이 아니다. 이 함수는 `*Context` 와 살아 있는 journal·gateway 를
요구하고, 그 배선은 태스크 5.7(fault injection·race)과 L6 의 몫이다. 여기서는 그 공백을
숨기지 않고 적는다 — 진입 0 은 통과가 아니다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.refreshPairedStrategyEntryProductionAssembly` | 423:16 |
| `dispatchHandoff` | 428:13 |
| `fresh.proposals.forMarket` | 428:13 |
| `handoff.Admitted` | 429:6 |
| `c.Journal.CurrentPositionCampaignCAS` | 433:14 |
| `string` | 433:76 |
| `fresh.dispatch.dispatch` | 440:11 |
| `errors.Is` | 441:5 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- 이 함수 자신은 아무 상태도 쓰지 않는다. 쓰는 일은 전부 `dispatch` 안에서 일어난다.

## Safety conclusion

dispatch 호출 자리는 **정확히 하나**이고 그 자리에 넘어가는 값은 `handoff.result` 다.
둘 다 `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 가 AST 로 고정한다.
관문이 완화되지 않았음은 handoff 의 단위 시험이 지킨다.
