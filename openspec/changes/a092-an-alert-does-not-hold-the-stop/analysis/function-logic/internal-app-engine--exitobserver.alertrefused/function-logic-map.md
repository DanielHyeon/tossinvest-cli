# Function Logic Map: `ExitObserver.alertRefused`

- Source: `internal/app/engine/exitloop.go` (1516-1538)
- AST evidence: `ast.json` — branches 2, returns 1, calls 5, assignments 3,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

`alertProposalRefused`와 **짝으로만 의미가 있다.** 둘은 이름이 비슷하고 같은 판정 경로에
있지만, 하나는 래치가 있고 하나는 없다. a092는 편집하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m managed` | 관리 포지션 | 호출자 `judge` | — |
| `cause` | 판정 거부 오류 | `exitpolicy` | `errors.As`가 실패하면 `field`는 빈 문자열 |
| `o.refused` | `map[string]bool` | 이 함수(`:1520`)와 `clearRefused`(`:1540`) | 이미 true면 B1이 반환 |

## Branches and early returns

| Branch | 위치 | 조건 | Mutation/side effect | Return | `Notify` 도달 |
|---|---|---|---|---|---|
| B1 | `:1517` | `o.refused[m.position.ID]` — 이미 알림 | 없음 | `return` `:1518` | ❌ |
| — | `:1520` | — | `o.refused[ID] = true` | — | — |
| B2 | `:1523` | `errors.As(cause, &refusal)` | `field = refusal.Field` | — | — |
| — | `:1526` | — | **`o.alert(...)` — `Notify` 도달 (P4)** | 암묵 `:1538` | ✅ |

**B1이 래치다.** 포지션당 1회이고, 해제는 `clearRefused`(`:1540`)뿐이다.
그래서 **이 알림은 사이클마다 반복되지 않는다.**

## Calls and live bindings

| Callee | 위치 | 계약 | Evidence |
|---|---|---|---|
| `errors.As` | `:1523` | 즉시 | AST calls |
| `o.alert` | `:1526` | **동기·기한 없음.** critical(`EventExitJudgementRefused`) | `internal-app-engine--exitobserver.alert` FLM |
| `string`·`o.label`·`cause.Error` | `:1528`·`:1529`·`:1531` | 즉시 | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `o.refused[ID]` | `:1520` | **래치** — 포지션당 1회 |
| `field` | `:1521`·`:1524` | 로컬 |

- goroutine 없음, defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: **yes** — §0.5. "손절도 평가되지 않는다"를 사람에게 알리는 알림이다.
- **a092가 여기서 얻는 사실**: 래치가 있는 알림과 없는 알림이 같은 사이클에 섞여 있다.
  래치가 있는 쪽만 보고 "사이클당 1회"라고 쓰면 거짓이 된다 — 2라운드 H2가 그것이었다.
  실측 표본 6개 중 4개가 이 경로(`exit.proposal_refused`)와 이 함수의 이벤트
  (`exit.judgement_refused`)에서 나왔다(`analysis/delivery-latency.md` §2).
