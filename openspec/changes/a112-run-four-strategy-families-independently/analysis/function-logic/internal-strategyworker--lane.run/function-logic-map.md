# Function Logic Map: `Lane.Run`

- Source: `internal/strategyworker/lane.go` (205-213)
- Function: `Lane.Run` in package `strategyworker`
- Signature: `Lane.Run(params=2, results=1)`
- File SHA-256: `b6919f2ac3ce70c08631286b8c879bd3b3ab273228d21246b359fa5031e594c5`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

잠금을 먼저 보고 worker 사이클로 넘긴다. 태스크 8.7.1 이 인자를 하나 늘렸다:
`activation strategyrouter.FamilyActivation` — 그 값은 그대로 worker 에게 넘어간다.

**순서가 계약이다: 잠금이 활성화보다 먼저다.** 잠긴 레인을 DORMANT 로 보고하면
운영자는 "아직 안 켰다" 로 읽는다. 실제로는 고장으로 닫힌 것이고 그 둘은 필요한
조치가 다르다(하나는 켜는 일, 다른 하나는 복구 증거를 찾는 일). 뒤집으면 승격이
사라진 순간 잠긴 레인의 진단이 "안 켰다" 로 바뀌어 복구가 필요한 상태가 시야에서
사라진다.

잠금 상태는 `lane.mu` 아래에서 **복사해 나온다**(206-208). 잠금을 들고 worker 를
부르지 않는 이유: worker 사이클은 이 레인의 상태를 만지지 않지만, 잠금을 들고
바깥 코드를 부르면 그 안에서 같은 잠금을 다시 잡는 날 교착이 된다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- strategyworker tagged / untagged suite: `go test -c [-tags tossos_testseams] -covermode=count -coverpkg=./internal/strategyworker ./internal/strategyworker/` 뒤 `-test.coverprofile`.
- Per-test attribution set: 두 바이너리의 테스트 **전체**(태그 72 · 무태그 54).
- **귀속 완전성은 측정이다.** 아래 분기에서 테스트별 진입 수의 합이 스위트 전체 진입
  수와 같다. 어긋난 행은 `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 없다.

Exact AST return positions: 210:3, 212:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 209:2 | arm entered 4x (strategyworker tagged suite); arm entered 3x (strategyworker untagged suite); `TestALatchedLaneEmitsNothingEvenWhenItIsEffective`, `TestALatchedLaneReportsTheLatchRatherThanDormancy`, `TestAProductionLaneBornFromADurableRecordIsLatched`, `TestARestoredLaneStillCannotUnlatchItself` |

B1 은 잠긴 레인이다 — LATCHED 를 돌려주고 worker 를 부르지 않는다. 통과하면 승격
판정은 worker 가 한다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.mu.Lock` | 206:2 |
| `lane.mu.Unlock` | 208:2 |
| `lane.worker.Run` | 212:9 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.
- 그 하나(`latched := lane.latched`)는 지역 사본이다. 이 함수는 레인 상태를 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: 잠금 안에서 하는 일은 한 필드를 읽는 것뿐이다.
- High-risk impact: yes — 잠긴 레인의 가족이 조정자에 닿지 않는 것이 이 change 가
  사려던 가족 단위 고장 격리다. 잠금 검사를 지운 변이(E5)는 CAUGHT.
