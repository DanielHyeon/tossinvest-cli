# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (1077-1197)
- AST evidence: `ast.json` — branches 14, returns 5, calls 6, assignments 2
- Risk scan: `risk-pattern-report.md`
- **이름 정정**: a089 초안의 P표는 이 함수의 줄을 `judge:1145`·`judge:1171`·`judge:1190`으로
  적었다. `judge`는 **807-840**의 다른 함수다. 아래 줄들은 전부 `record` 안에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.Orderable` | bool | `exitpolicy` 순수 함수 (`snapshot.go:199-210`) | false면 `orderable=false`로 시작 |
| `snapshot.CancelPendingFirst` | bool | 판정 | true면 B3 진입 |
| `proposal.Action` | enum | 판정 | `isFullExit`가 **익절 포함**(`:1208-1211`) |
| `m.reJudge` | bool | 격리 재판정 | B4의 유일한 가드 |
| `o.opts.Journal` | non-nil | 주입 | `RecordExitJudgementResult` 오류가 B11 계열 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | 기존 테스트 |
|---|---|---|---|---|
| B1/B2 `:1095,:1097` | `quote.FetchedAt.IsZero()` | `judgement.ObservationSource` | (통과) | 간접 |
| B3 `:1117` | `orderable && (CancelPendingFirst \|\| isFullExit)` | — | (진입) | `:861`,`:883` |
| B4 `:1118` | `m.reJudge && !isProtective` | `orderable=false`, `ArmSuppressedReJudge` | (통과) | `a084b_rejudge_bound_test.go` |
| B5 `:1140` | else (평시) | `clearTheSymbol` 호출 | — | `:861` |
| B6 `:1142` | `clearTheSymbol` 오류 | 없음 | **`return err` `:1143`** | **없음** |
| B7 `:1145` | `!cleared` | **`noteDelay`**, `orderable=false`, `ArmSuppressedWorkingOrder` | (통과) | **`:883` ✅** |
| B8 `:1149` | `cleared` | **`clearDelay` `:1150`** | (통과) | `:914-919` |
| B9/B10 `:1156,:1158` | `orderable`, `intentID==""` | `judgement.Proposal` 설정 | (통과) | 간접 |
| B11 `:1170` | `RecordExitJudgementResult` 오류 | — | (분기) | 간접 |
| B12 `:1171` | `ErrProposalPending` | 없음 | **`return nil` `:1175`** | **없음** |
| B13 `:1177` | `ErrExitSnapshotQuarantined` | `announceQuarantineFromLedger` | **`return err` `:1188`** | a074 계열 |
| B14 `:1190` | `ArmedProposal==nil \|\| ArmOutcome != Armed` | 없음 | **`return nil` `:1191`** | **없음** |
| — `:1195` | 정상 | — | `return o.submit(...)` | 다수 |

## B7/B8이 지연 시계의 전부다 — 그리고 원인별로 갈린다

`noteDelay`(B7)와 `clearDelay`(B8)는 **같은 조건의 양변**이다: `clearTheSymbol` 성공 여부.

| 미제출 원인 | `cleared` | 시계 |
|---|---|---|
| working order를 못 치웠다 (B7) | false | **시작·누적** — `:883` 테스트가 31초 후 critical을 확인 |
| 브로커가 거부했다 (`submit` B10) | **true** (거부된 주문은 살아 있지 않다) | **B8이 매 주기 초기화** |

즉 시계는 **고장난 것이 아니라 다른 원인을 재고 있다.** 거부 경로의 시계가 안 뜨는 이유는
`noteDelay` 호출자가 없어서가 아니라 **`clearDelay`가 submit보다 먼저 무조건 돌기 때문**이다.
a089 1라운드 C1의 진단은 여기서 확인된다.

**따라서 `:1150`을 그냥 제거하면 안 된다** — B7이 시작한 시계의 유일한 해제점이다.
제거하려면 그 해제를 `submit`의 `StateConfirmed`가 대신해야 하고, `:914-919`
("취소가 성공하면 청산이 나간다")가 그 대체를 검증하는 회귀 테스트가 된다.

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `o.clearTheSymbol` `:1141` | 충돌 주문 정리 | 오류 → `return err`(B6) | AST |
| `o.noteDelay` `:1146` | 지연 시계 시작 | 오류 없음 | AST |
| `o.clearDelay` `:1150` | 지연 시계 해제 | 오류 없음 | AST |
| `Journal.RecordExitJudgementResult` `:1169` | 판정 커밋 + 무장 | 3종 오류 분기 | AST |
| `o.announceQuarantineFromLedger` `:1186` | 격리 생성 알림 | 오류 없음 | AST |
| `o.submit` `:1195` | 브로커 제출 | 오류 전파 | AST |

## State mutations and fallbacks

- `orderable`은 B4·B7에서 false로 꺾인다. 두 경우 모두 `ArmSuppressedReason`이 원장에 남는다
- **실측**: `arm_suppressed_reason`이 `exit_events` **671행 전부 비어 있다** — B4·B7은
  이 배포에서 한 번도 발화한 적이 없다 `[미측정]`
- **실측**: `effective_source`에 `saved` **0건** → B14의 `ExitArmSuppressedSavedMonotone`
  (`exit_state.go:613-615`)도 미발화 `[미측정]`

## Safety conclusion

- **Safe edit boundary**: B8의 `clearDelay` 위치 변경은 §0.3 경로 밖(기록만)이지만
  **B7의 해제를 함께 옮겨야** 한다. 단독 제거는 시계를 latch시킨다.
- **High-risk impact**: **yes**
- **B12·B14는 "보호 주문이 없다"를 뜻하지 않는다**: `pending_action`은 접수 성공 후에도
  남고(`submit`의 `StateConfirmed`가 `release`를 부르지 않는다), `:1109-1111`이 "the arming
  refuses a second proposal while the first is outstanding"라고 명시한다. 이 둘을 미제출로
  계수하면 **정상 보호 중인 포지션에 거짓 critical**이 뜬다. a089 초안 spec의
  "어느 경로로 멈췄든 보호 주문이 없다는 사실은 같다"는 **거짓**이다.
