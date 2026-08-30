# Function Logic Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go` (144-156)
- Function: `strategyProposalAuthorityPair.ResultAuthority` in package `engine`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- File SHA-256: `1ce0765cff483524cfbb428959be1d1b83da126f6533d07f708bce06b7fe1e7c`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

이 함수는 5.5 에서 바뀌었다. 이전에는 `len(value.entries) != 1` 을 직접 세었고, 이제는
`value.dispatchHandoff().Single()` 이 정한 결과만 읽는다. 세는 규칙이 사라진 것이 아니라
패키지 밖 `internal/strategyhandoff` 한 곳으로 옮겨졌다 — 엔진 안에서는 "여기서는 주문을 낼
수 없다"를 import 로 증명할 수 없기 때문이다.

새로 생긴 조건이 하나 있다: handoff 는 `snapshot.Ready` 도 본다. 이전 이 함수는 보지 않았다.
그 검사가 거부하게 될 **정상 입력**은 없다 — 항목을 담은 시장 권한은 모두 준비된 시장이며,
그 사실은 사람이 읽은 것이 아니라 `TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady` 가
생산 소스를 파싱해 확인한다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.9% of statements (2912/3941).
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS,
  **63.5~63.6% of statements** — 같은 바이너리·같은 트리에서 다섯 번 재서 2498·2499·2501·2502·2502
  of 3936 이 나왔다. 이 수는 실행마다 흔들린다(무태그 쪽에 스케줄링 의존 시험이 있다).
  앞선 판본이 표본 하나(63.5%)를 안정된 값으로 적은 것을 정정한다. 태그 스위트는 안정적이다.
- 분기의 arm 은 그 분기 위치 **다음에 처음 열리는** 커버리지 블록이다. 조건이 여러 줄이면
  여는 중괄호가 다음 줄에 있어서, 같은 줄만 보던 첫 판본은 이 태스크가 실제로 바꾼 분기를
  `null`(=측정 없음)로 보고했다. "자료 없음"은 "진입 0"과 다르고 그 차이가 요점이다.
- Per-test attribution set: `buildProductionStrategyMarketWorker` 를 부르는 시험은 트리 전체에서 둘
  (`strategy_dispatch_cycle_test.go:191`, `a112_dispatch_handoff_test.go:112`)이고,
  `ResultAuthority` 를 부르는 시험은 **둘**이다
  (`strategy_proposal_authority_test.go:44`, `a112_arbitration_test.go:214`). 그 각각을
  `-test.run '^<Test>$'` 로 따로 돌렸다. 앞선 판본이 "넷"이라고 적은 것은 **grep 히트 수를 시험
  수로 센 것**이다 — 44 행에 호출이 둘 있고, `a112_arbitration_test.go:215` 의 호출은 `t.Fatalf`
  안이라 실행되지 않는다.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 시험별 진입 수의 합이 스위트 전체
  진입 수와 정확히 같다. 이 집합 밖의 시험이 어느 arm 이든 들어갔다면 그 등식이 깨진다.
  깨진 행은 `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Exact AST return positions: 151:4, 153:3, 155:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 150:3 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` |

중재가 거절한 시장은 `Ready=false` 로 닫히므로 handoff 는 `HANDOFF_MARKET_CLOSED` 를 돌려주고
이 arm 이 그것을 받는다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Single` | 149:24 |
| `value.dispatchHandoff` | 149:24 |
| `result.ValidProposal` | 150:21 |
| `convert` | 155:70 |
| `convert` | 155:110 |

## State mutations and fallbacks

- AST assignments: 2. Defers: 0. Goroutine statements: 0.
- 이 함수는 아무것도 바꾸지 않는다. 값을 만들어 돌려줄 뿐이다.

## Safety conclusion

시장 단위 단일 제안 가정의 소비자 **여섯 함수** 중 하나였고, 5.5 에서 경계를 쓰도록 바뀌었다.
가정 자체를 없애는 것은 태스크 5.2 의 몫이며 이 태스크는 상한을 풀지 않았다.
