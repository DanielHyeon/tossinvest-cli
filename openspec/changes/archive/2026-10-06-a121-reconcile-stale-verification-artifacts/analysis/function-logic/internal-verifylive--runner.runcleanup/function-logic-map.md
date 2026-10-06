# Function Logic Map: `Runner.runCleanup`

- Source: `internal/verifylive/cleanup.go` (230-254)
- Qualified function: `Runner.runCleanup`
- Revision: `current` (구현 base `de147cc2` = 고정 사본 `2c6ef1ef` 의 같은 파일, `source_sha256` 5cdcedb9…)
- AST evidence: `ast.json` — AST branches 8, `return` 문 0(함수 끝 253:2 `sr.resolve` 로 종료), 호출 6
- Risk scan: `risk-pattern-report.md`
- 편집 예정: **없음.** freeze P2-11 — proposal 「a063 과의 관계」 가 "그 대상은 `runCleanup` 으로 가고 그것은
  `CancelConditionalOrder` 를 부를 수 있다 — 같은 DELETE 의 재시도다(이 분기는 이 개정에서 AST 로 열거하지 않았다)" 라고
  주장했다. 이 번들이 그 주장의 AST 근거다. 대사 경로는 이 함수를 **재사용하지 않는다**(design G2·contract-hard-evidence 3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | 호출자 | 취소·마감은 B8 로 루프 탈출 |
| `sr *stepRun` | 정리/중단 단계 | `Runner.cleanup`(`cleanup.go:315`) · `Runner.Abort`(`abort.go:129`) | `sr.resolve(first)` 로 판정 |
| `targets []Artifact` | `cleanupTargets()` 또는 abort 대상 | `cleanupFrom`·`withoutM0ManualReconcile`(`cleanup.go:103-108`) | 빈 목록이면 루프 0회 → pass |

불변식: (1) 모든 브로커 호출은 `r.cancelOrder`/`r.cancelConditional` 을 통하고, 그 둘은 `r.gate`(승인 목록 대조)를 먼저
통과해야 브로커를 부른다(`mutate.go:319`·`:624`). (2) 실패는 기록하고 계속한다 — 첫 오류만 판정에 쓴다(B7). (3) 성공한 취소만
`sr.cancelled` 로 artifact 를 종결한다(`mutate.go:637`) — **실패한 DELETE(404 포함)는 artifact 를 종결하지 않는다.**

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range at 232:2` | 대상 순회 | — | 없음 | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` |
| B2 | `switch at 234:3` | `a.Kind` 분기 머리 | — | 없음 | B3·B4 를 통해 평가(머리 자체엔 커버리지 블록 없음) |
| B3 | `case at 235:3` | `KindOrder` | `r.cancelOrder` → `broker.CancelOrder`(`mutate.go:333`) | `err` 대입 | `TestALeftoverOrderIsCancelledOnTheNextRun` |
| B4 | `case at 237:3` | `KindConditional` | **`r.cancelConditional` → `broker.CancelConditionalOrder`(`mutate.go:632`)** — proposal 이 주장한 분기 | `err` 대입 | `TestAbortEndsAHeldChain` |
| B5 | `case at 239:3` | 그 밖의 종류(default) | 없음 | `continue` | **기준선 미커버**(아래) |
| B6 | `if at 242:3` | `err == nil` | — | `continue` | `TestALeftoverOrderIsCancelledOnTheNextRun` |
| B7 | `if at 245:3` | `first == nil` | 첫 오류 보존 | 없음 | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` |
| B8 | `if at 248:3` | `ErrOutsidePlan` · `context.Canceled` · `DeadlineExceeded` | — | `break`(루프 탈출) | `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` |

`return` 문은 없다. 루프 뒤 `sr.resolve(first)`(253:2)가 판정을 한 번 정한다 — `first == nil` 이면 pass, 아니면
`resolve`(`runner.go:954-976`)의 분류(fail · refused · outside-plan · interrupted).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.cancelOrder` | 남은 일반 주문 취소 | `r.gate` 승인 대조 → `broker.CancelOrder`, 재시도 루프(`mutate.go:331`) | AST + HEAD |
| `r.cancelConditional` | 남은 조건주문 취소 | `r.gate` → `broker.CancelConditionalOrder` 1회, 실패 시 `logCall` 만 남기고 오류 반환(`mutate.go:631-636`) | AST + HEAD |
| `errors.Is` ×3 | 루프 중단 사유 | 순수 | AST |
| `sr.resolve` | 단계 판정 | 순수(상태 갱신) | AST + HEAD |

호출자(CodeGraph 1.6.0 + HEAD): `Runner.cleanup`(`cleanup.go:317`, 재개 prologue) · `Runner.Abort`(`abort.go:131`).

## State mutations and fallbacks

- `sr` 의 호출 기록(`logCall`)·취소 artifact(`cancelled`)·판정이 바뀐다. 기록 파일 추가는 호출자가 한다(`cleanup.go:321`).
- **a121 관련 결론**: 재개 prologue 가 대사되지 않은 stale 조건주문을 대상으로 고르면 B4 가 같은 `DELETE` 를 다시 보낸다
  (승인 목록에 올라간 뒤). a063 진단의 "Do **not** retry, cancel, resume" 이 막는 것이 바로 이 분기다. 대사는 `Outstanding`
  투영에서 그 artifact 를 빼서(`terminal()` 셋째 종결) 이 분기에 도달할 대상을 없애는 방식이고, **이 함수 자체는 편집하지
  않는다.**
- B5(default) 는 `Kind` 가 `order`·`conditional-order` 가 아닌 artifact 를 조용히 건너뛴다 — 현 기록 작성기는 두 종류만 쓰므로
  도달 경로가 없다(기준선 미커버). 대사 줄은 대상 artifact 의 `Kind` 를 그대로 재사용하므로(Q2 (a)) 이 분기에 새 값을 보내지
  않는다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** 대사 경로가 이 함수·`cleanup`·`Abort` 를 호출하면 안 된다 — 구조 시험이 대사 파일의
  쓰기 7이름 호출을 금지한다(`TestReconcileFilesCallNoWriteMethodAndAssertNoType`).
- High-risk impact: yes(실주문 취소 경로) — 그래서 재사용 금지가 결론이다.
