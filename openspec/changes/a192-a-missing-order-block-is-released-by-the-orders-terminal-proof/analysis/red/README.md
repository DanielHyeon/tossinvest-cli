# RED 관측 (a192 2단계, 2026-11-17)

순서: 새 파일(`internal/journal/reconcile_order_terminal_release.go`, `internal/reconcile/order_terminal_release.go` — 판정·결속 함수 본체)과
타입 필드(`Tracker.Terminal`·`Tracker.terminal`, `Block.missing`·`Block.terminalRelease`)만 넣고 **기존 함수 본문은 손대지 않은** 골격에서 시험을 돌림.
그래서 컴파일은 되고, 실패는 전부 **단언 실패**임(컴파일 실패 RED 0).

| 파일 | 명령 | rc | 실패 시험(단언) |
|---|---|---|---|
| `red-1-reconcile.txt` | `go test -count=1 -run TestA192 ./internal/reconcile/` | 1 | 6 — `…ArmsThenReleases…`(「second agreeing comparison … want the block released」, AwaitingAdjustment 로 남음), `…EveryBoundOrderMustBeTerminal`, `…AnOpenOrderWithTheSameIDKeepsTheBlock`(공백 다른 번호 대조 쪽), `…ABindingChangeRearms`, `…AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`(「release attempts = 0」), `…APermanentBlockStaysOperatorOnly`(종목 행 해제 없음) |
| `red-2-journal.txt` | `go test -count=1 -run TestA192 ./internal/journal/` | 1 | 3 — 새 원인이 `ValidReconcileReleaseCause` 에 없어 거절(「is not one this build writes」) |
| `red-3-engine.txt` | `go test -count=1 -run TestA192 ./internal/app/engine/` | 1 | 1 — `…TheReleaseLogCarries…`(「releasing cycle … Released:0 … want the block released」) |

RED 에서 **통과한** 시험(유지 쪽 시험 — 해제가 없는 골격에서는 당연히 통과)과 판정 함수 단위 시험(골격에 판정 본체가 이미 있음)은 RED 로 세지 않는다.
그것들의 가치는 변이(`../mutation/`)로 증명함. GREEN 뒤에 쓴 시험: `TestA192ReleaseWriteFailureKeepsTheGate`(커버 측정에서 persist B8 이 0 이라
추가), `TestA192AFillBeforeTheBlockDoesNotKeepIt`·`TestA192ACreditWinsOverTheTerminalProof`(변이 m23·m24·m16 을 잡으려고 추가),
`TestA192TheProductionTrackerReadsTerminalProofFromItsJournal`(m29), 재기동 시험의 순서 강화(m2) — RED 관측 없음, 변이로 대신 증명.
