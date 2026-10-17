# Function Logic Map: `FamilyWorker.Run`

- Source: `internal/strategyworker/worker.go` (193-203)
- Function: `FamilyWorker.Run` in package `strategyworker`
- Signature: `FamilyWorker.Run(params=2, results=1)`
- File SHA-256: `588c56f0a3100d0f1fa93e8fed4cf303b506ad88823c53e8f8690ebe84504335`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

봉인된 제안 하나를 조정자 봉투로 바꾼다. 태스크 8.7.1 이 인자를 하나 늘렸다:
`activation strategyrouter.FamilyActivation`.

**승격은 이 값의 필드가 아니다.** 앞 판본은 `desired`/`effective` 를 필드로 들고
있었고 `newWorker` 가 그 상태를 인자로 받았다. 그래서 켜진 worker 를 이 패키지의
시험이 직접 주조할 수 있었다. 이제 두 상태는 활성화의 함수이고
(`worker.Effective(activation)`), 활성화는 `internal/strategyrouter` 밖에서 영값
말고는 만들 수 없다(필드 전부 비공개). 영값은 아무것도 승격하지 않으므로 **영값이
안전한 값**이고, 켜려면 ed25519 서명 검증을 통과해야 한다.

왜 필드가 아니라 인자인가: 활성화에는 24시간 수명 상한이 있고 레인은 프로세스
수명이다(5.1.2.1). 승격을 값 안에 구우면 묵은 ON 이 생긴다. 묵은 OFF 는 안전한
방향이지만 묵은 ON 은 아니다.

이 함수는 아무것도 바꾸지 않는다. 그 약속은 이 자리에서 확인할 수 없고 패키지의
import 폐포가 확인한다(`dependency_closure_test.go`, `-deps`/`-deps-test`).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- strategyworker tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyworker ./internal/strategyworker/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행.
- strategyworker untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것.
- Per-test attribution set: 두 바이너리의 `-test.list '.*'` **전체**(태그 72개 · 무태그 54개)를
  `-test.run '^<Test>$'` 로 하나씩 돌린 프로파일.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이
  스위트 전체 진입 수와 정확히 같다. 어긋난 행은 `ATTRIBUTION MISMATCH` 로 표시되며
  아래에는 하나도 없다.

Exact AST return positions: 195:3, 198:3, 201:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 194:2 | arm entered 9x (strategyworker tagged suite); arm entered 1x (strategyworker untagged suite); `TestALatchedLaneReportsTheLatchRatherThanDormancy`, `TestEveryProductionWorkerIsBornDormantAndEmitsNothing` |
| B2 | if | 197:2 | arm entered 4x (strategyworker tagged suite); arm not entered (strategyworker untagged suite); `TestAWorkerOfTheOtherMarketRefuses`, `TestAWorkerRefusesAProposalFromAnotherLane`, `TestAWorkerRefusesAProposalWhoseSealNoLongerHolds`, `TestAWorkerRefusesWhenTheFamilyCannotBeDerivedFromTheSealedAuthority` |

B1 은 승격되지 않은 worker 다 — DORMANT 이고 거절 코드를 붙이지 않는다(붙이면 운영자가
원인을 중재 실패로 읽는다). B2 는 이 레인의 제안이 아니라는 거절이다. 둘을 통과하면
봉투가 나온다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `worker.Effective` | 194:5 |
| `worker.owns` | 197:6 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

- Safe edit boundary: 값 수신자이고 대입이 0 이다. 이 함수는 아무것도 기억하지 않는다.
- High-risk impact: yes — 이 판정이 진입 제안이 조정자에 닿는지를 정한다. B1 을 지운
  변이(E12)는 CAUGHT.
