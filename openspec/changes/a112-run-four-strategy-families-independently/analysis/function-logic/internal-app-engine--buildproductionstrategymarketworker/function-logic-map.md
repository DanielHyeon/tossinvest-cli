# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (411-471)
- Function: `buildProductionStrategyMarketWorker` in package `engine`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- File SHA-256: `22855de0f27de05c60c2b5ff8cf2d5c7e3ed50e78a9fa6f67fb81ec38decdbfa`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 6.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

worker 를 Effective 로 올릴지 정하는 함수다. 5.5 는 여기서 제안 개수를 세던 두 조각
(`!p.snapshot.Ready`, `len(p.entries) != 1`)을 지우고 `p.dispatchHandoff()` 하나로 바꿨고,
5.5-fix 는 그 값을 읽는 방법을 `Single()` 로 바꿨다 — 제안과 "건너가도 되는가"가 한 서명으로
같이 오므로, 거절을 안 보고 값을 읽는 판본은 쓸 수 없다. 다른 권한(schedule/candidate/route/
fx/risk/account)의 준비 상태 검사는 그대로다.

이 함수는 게이트웨이를 **읽기만** 한다. 그 약속은 주석이 아니라
`TestTheWorkerBuilderOnlyObservesThroughTheGateway` 가 지킨다 — 이 본문에서 `gateway.` 로
시작하는 호출은 전부 `Observe` 로 시작해야 한다.

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
- Per-test attribution set: `buildProductionStrategyMarketWorker` 를 부르는 시험은 트리 전체에서 둘
  (`strategy_dispatch_cycle_test.go:191`, `a112_dispatch_handoff_test.go:124`)이고,
  `ResultAuthority` 를 부르는 시험은 **둘**이다
  (`strategy_proposal_authority_test.go:44`, `a112_arbitration_test.go:214`). 그 각각을
  `-test.run '^<Test>$'` 로 따로 돌렸다. 앞선 판본이 "넷"이라고 적은 것은 **grep 히트 수를 시험
  수로 센 것**이다 — 44 행에 호출이 둘 있고, `a112_arbitration_test.go:215` 의 호출은 `t.Fatalf`
  안이라 실행되지 않는다.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 시험별 진입 수의 합이 스위트 전체
  진입 수와 정확히 같다. 이 집합 밖의 시험이 어느 arm 이든 들어갔다면 그 등식이 깨진다.
  깨진 행은 `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Exact AST return positions: 418:3, 433:3, 436:3, 442:3, 445:3, 466:3, 468:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 417:2 | arm entered 4x (engine tagged suite); arm entered 2x (engine untagged suite); `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B2 | if | 431:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedHandoffLeavesTheWorkerDormant` |
| B3 | if | 435:2 | arm entered 6x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedHandoffLeavesTheWorkerDormant`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` |
| B4 | if | 441:2 | arm entered 6x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedHandoffLeavesTheWorkerDormant`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` |
| B5 | if | 444:2 | arm entered 5x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedHandoffLeavesTheWorkerDormant`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` |
| B6 | if | 465:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

B3·B5·B6 은 **이 태스크가 바꾸지 않은** 분기이고 진입 0 은 커버리지 공백이지 통과가 아니다.
그 공백을 메우는 것은 태스크 5.7(fault injection)의 몫이다.


**태스크 8.7.1 이 분기 둘을 더했고 그래서 옛 B6 이 B8 이 되었다.**
B6(437:2)은 "이 시장에 검증된 4-가족 활성화가 있는가" 이고, B7(438:3)은 그 활성화가
이름 부른 **위험 번들 digest 와 ProtectionReady digest 가 살아 있는 값과 같은가** 다.
그 둘을 여기서 결속하는 이유는 두 사실이 제안 수집 단계에 **존재하지 않기** 때문이다
(둘 다 제안 뒤에 수집된다). 없는 사실을 결속하면 그 결속은 어떤 정상 입력으로도 참이
될 수 없고, 그것이 이 change 가 이미 한 번 만들었다 고친 문 없는 fail-closed 다.
활성화가 없는 시장에서는 B6 이 참이 되지 않으므로 아무것도 요구하지 않는다 —
그것이 오늘의 동작이고, 그 방향이 실제로 짐을 진다는 것은 반증 E17(활성화가 없어도
결속을 요구하게 만든 변이)이 CAUGHT 인 것으로 확인했다.
B8(446:2)은 옛 B6 그대로다: 증거 digest · desired revision · 권한 만료.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.forMarket` | 426:27 |
| `candidate.forMarket` | 426:55 |
| `route.forMarket` | 426:84 |
| `fx.forMarket` | 427:3 |
| `proposal.forMarket` | 427:25 |
| `riskAuthority.forMarket` | 427:53 |
| `account.forMarket` | 427:86 |
| `Single` | 430:23 |
| `p.dispatchHandoff` | 430:23 |
| `result.ValidProposal` | 435:6 |
| `gateway.ObserveStrategyProtection` | 441:15 |
| `strings.ToLower` | 441:54 |
| `string` | 441:70 |
| `gateway.ObserveStrategyEntryGate` | 444:15 |
| `strings.ToLower` | 444:53 |
| `string` | 444:69 |
| `strategyWorkerEvidenceDigest` | 462:12 |
| `validStrategyDigest` | 465:6 |
| `IsZero` | 465:64 |
| `a.authority.FreshUntil` | 465:64 |
| `a.authority.FreshUntil` | 469:23 |

## State mutations and fallbacks

- AST assignments: 10. Defers: 0. Goroutine statements: 0.
- `dormant` 지역 변수만 바꾼다. 이 함수는 원장·게이트웨이·활성화 어느 것도 쓰지 않는다.

## Safety conclusion

승격 조건이 **완화되지 않았다.** 경계는 이전 두 조각이 통과시키던 것과 정확히 같은 집합을
통과시킨다(준비된 시장 + 항목 정확히 하나). 상한을 푸는 것은 태스크 5.2 다.

5.5-fix 이후 이 자리의 관문은 우연에 기대지 않는다. 예전에는 `!handoff.Admitted()` 를 지워도
바로 아래 `ValidProposal()` 이 대신 걸러서 아무 시험도 깨지지 않았다. 지금 그 편집은
`TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 가 죽인다(M5 KILLED).
