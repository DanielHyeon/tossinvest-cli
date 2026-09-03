# Function Logic Map: `strategyFamilyLaneStep`

- Source: `internal/app/engine/strategy_lane_runtime.go` (164-170)
- Function: `strategyFamilyLaneStep` in package `engine`
- Signature: `strategyFamilyLaneStep(params=2, results=1)`
- File SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

레인 하나가 사이클마다 실제로 도는 일을 만든다. **이 함수가 패키지 수준이고 인자가
값 둘뿐인 것이 태스크 5.1.2.1 의 요점이다.** 메서드로 두면 수신자가 무엇이든 될 수
있고, `*Context` 를 수신자로 두는 순간 원장과 게이트웨이가 전략군 평가 안으로
들어온다 — 오늘 생산의 시장 주기가 정확히 그 모양이다.

무엇을 만질 수 있는지가 주석이 아니라 **인자의 타입**으로 정해진다.
`*strategyworker.Lane` 이 사는 패키지는 자기 import 폐포에 broker mutator ·
writable journal · Guardian issuer 가 없다는 것을 `-deps`/`-deps-test` 로 훑어
시험으로 지킨다. 태스크 8.7.1 이 더한 둘째 인자
`strategyrouter.FamilyActivation` 도 능력이 아니라 값이다 — 필드가 전부 비공개라
이 패키지에서는 영값만 만들 수 있고 영값은 아무것도 승격하지 않는다.

두 인자의 **타입 목록**은 `TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion`
이 얼려 둔다. 앞 판본은 "인자는 하나" 라고 **개수**를 셌는데, 개수 검사는 정당한
인자가 늘 때 반드시 걸리고 걸렸을 때 통과시키는 유일한 방법이 숫자를 고치는 것이라
감시 대상과 같은 손짓이 된다.

오류를 절대 돌려주지 않는 이유: 평가 실패는 오류가 아니라 **거절**이고 거절은
`Cycle.Outcome` 에 담긴다. 둘을 한 값에 담으면 정당한 거절이 레인의 연속 실패
계수기를 올려 결국 진입을 잠근다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged / untagged suite: `go test -c [-tags tossos_testseams] -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/` 뒤 `-test.coverprofile`;
  스위트 실행은 `systemd-run --user --scope -p MemoryMax=12G -p MemorySwapMax=0` cgroup 안에서 돌렸다
  (이 패키지의 무제한 실행이 앞 로트에서 커널 OOM 을 냈다).
- Per-test attribution set: 두 바이너리의 `-test.list '.*'` **전체**(태그 491개 · 무태그 438개).

분기가 없다. 아래 한 줄은 행복 경로이며 이 함수의 몸통 진입 수다.

Exact AST return positions: 167:2, 168:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | branchless happy path | 164:1 | arm entered 170x (engine tagged suite); arm entered 170x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestALatchOnlyReopensForAStrictlyNewerSignedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |

위 수는 이 함수를 호출하는 `runLane` 의 `lane.RunBounded` 자리에서 잰 것이다
(`internal-app-engine--strategylaneruntime.runlane` 의 B1 과 같은 실행). 반환되는
클로저의 몸통(`lane.Run`)은 유계 사이클 안에서 도므로 여기서 따로 세지 않는다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Run` | 168:10 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

- Safe edit boundary: 인자 타입 둘이 이 함수가 닿을 수 있는 전부다.
- High-risk impact: yes — 레인 안에서 도는 값의 목록은
  `TestOnlyThePackageLevelStepEverRunsInsideALane` 이 패키지 전체에서 세어 한 줄로 얼린다.
