# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (377-419)
- Function: `buildProductionStrategyMarketWorker` in package `engine`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- File SHA-256: `12586e3cf90b708e66988931ad424d7312593bf518f0987a0893bf4f6f4b6fb9`
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

Exact AST return positions: 384:3, 399:3, 402:3, 405:3, 408:3, 414:3, 416:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 383:2 | arm entered 2x (tagged); 0 (untagged); `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` (`wiringReady=false` 두 시장) |
| B2 | if | 397:2 | **이 태스크가 바꾼 분기.** arm entered 1x (tagged); 0 (untagged); `TestARefusedHandoffLeavesTheWorkerDormant` (KR 이 상한 초과로 거절) |
| B3 | if | 401:2 | arm not entered (양쪽 스위트). 봉인이 깨진 제안으로 여기까지 오는 시험이 없다 |
| B4 | if | 404:2 | arm entered 1x (tagged); 0 (untagged); `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` (`spy.failProtection` KR) |
| B5 | if | 407:2 | arm not entered (양쪽 스위트). 진입 게이트 관측 실패를 만드는 시험이 없다 |
| B6 | if | 413:2 | arm not entered (양쪽 스위트). 잘못된 digest·revision·만료를 만드는 시험이 없다 |

B3·B5·B6 은 **이 태스크가 바꾸지 않은** 분기이고 진입 0 은 커버리지 공백이지 통과가 아니다.
그 공백을 메우는 것은 태스크 5.7(fault injection)의 몫이다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.forMarket` | 392:27 |
| `candidate.forMarket` | 392:55 |
| `route.forMarket` | 392:84 |
| `fx.forMarket` | 393:3 |
| `proposal.forMarket` | 393:25 |
| `riskAuthority.forMarket` | 393:53 |
| `account.forMarket` | 393:86 |
| `Single` | 396:23 |
| `p.dispatchHandoff` | 396:23 |
| `result.ValidProposal` | 401:6 |
| `gateway.ObserveStrategyProtection` | 404:15 |
| `strings.ToLower` | 404:54 |
| `string` | 404:70 |
| `gateway.ObserveStrategyEntryGate` | 407:15 |
| `strings.ToLower` | 407:53 |
| `string` | 407:69 |
| `strategyWorkerEvidenceDigest` | 410:12 |
| `validStrategyDigest` | 413:6 |
| `IsZero` | 413:64 |
| `a.authority.FreshUntil` | 413:64 |
| `a.authority.FreshUntil` | 417:23 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.
- `dormant` 지역 변수만 바꾼다. 이 함수는 원장·게이트웨이·활성화 어느 것도 쓰지 않는다.

## Safety conclusion

승격 조건이 **완화되지 않았다.** 경계는 이전 두 조각이 통과시키던 것과 정확히 같은 집합을
통과시킨다(준비된 시장 + 항목 정확히 하나). 상한을 푸는 것은 태스크 5.2 다.

5.5-fix 이후 이 자리의 관문은 우연에 기대지 않는다. 예전에는 `!handoff.Admitted()` 를 지워도
바로 아래 `ValidProposal()` 이 대신 걸러서 아무 시험도 깨지지 않았다. 지금 그 편집은
`TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 가 죽인다(M5 KILLED).
