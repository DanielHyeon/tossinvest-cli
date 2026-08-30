# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go` (97-156)
- Function: `strategyProjectionFromAssembly` in package `engine`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- File SHA-256: `14ec90c888e64ccb7e45d5823f415cbf53a1b97b4a62adf9b476db478892f80a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 8.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

읽기 전용 projection 을 만든다. 5.5 는 여기서 `len(authority.entries) != 1` 를 지우고
`assembly.proposals.forMarket(market).dispatchHandoff()` 로 바꿨다. 이유는 화면이 보는 제안과
dispatch 가 받는 제안이 **같은 판단**에서 나와야 하기 때문이다. 따로 세면 둘이 갈라지고,
갈라진 순간 화면은 틀린 것을 정상으로 보여 준다.

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

Exact AST return positions: 155:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 100:2 | arm not entered (양쪽 스위트) |
| B2 | if | 107:3 | arm not entered (양쪽 스위트) |
| B3 | switch | 109:4 | arm not entered (양쪽 스위트) |
| B4 | case | 110:4 | arm not entered (양쪽 스위트) |
| B5 | case | 112:4 | arm not entered (양쪽 스위트) |
| B6 | case | 114:4 | arm not entered (양쪽 스위트) |
| B7 | if | 123:3 | **이 태스크가 바꾼 분기.** arm not entered (양쪽 스위트) |
| B8 | if | 132:3 | arm not entered (양쪽 스위트) |

이 함수는 `StrategyEntryProductionAssembly` 전체를 요구하고, 트리 안에 그것을 만드는 시험이
하나도 없다. `TestStrategyRuntimeRead*` 들은 `Context.Read` 를 시험하지 이 함수를 지나지 않는다.
운영자 화면에 handoff 거절을 실제로 비추는 일은 태스크 7.3 이고, 그 태스크가 이 공백의 주인이다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `assembly.Schedule.ObservedAt.UTC` | 98:14 |
| `strategyprojection.DormantSnapshot` | 99:14 |
| `strategyprojection.Market` | 101:23 |
| `assembly.Schedule.For` | 102:15 |
| `assembly.Candidate.For` | 103:16 |
| `assembly.Proposal.For` | 104:15 |
| `assembly.Risk.For` | 105:11 |
| `assembly.Supervisor.Snapshot` | 106:17 |
| `strategyprojection.WithMarketFailure` | 117:15 |
| `dispatchHandoff` | 122:14 |
| `assembly.proposals.forMarket` | 122:14 |
| `handoff.Admitted` | 123:7 |
| `handoff.result.ValidProposal` | 123:30 |
| `strategyprojection.WithMarketFailure` | 124:15 |
| `projectionDigest` | 131:58 |
| `projectionDigest` | 133:21 |
| `strconv.Itoa` | 135:44 |
| `string` | 136:28 |
| `string` | 137:59 |
| `projectionDigest` | 138:23 |

## State mutations and fallbacks

- AST assignments: 26. Defers: 0. Goroutine statements: 0.
- 지역 `snapshot` 만 바꾼다. 활성화·주문·원장 어느 것도 쓰지 않는다.

## Safety conclusion

이 함수는 읽기 전용이고 5.5 는 그 성질을 바꾸지 않았다. 바뀐 것은 제안 하나를 고르는 판단의
**출처**뿐이며, 이제 dispatch 경로와 같은 seam 을 쓴다.
