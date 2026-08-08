# Function Logic Map: `ExitObserver.alertProposalRefused`

- Source: `internal/app/engine/exitloop.go` (1548-1565)
- AST evidence: `ast.json` — **branches 0, returns 0**, calls 6, assignments 0,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

**a092의 "알림당 예산"이 "사이클당 예산"이 될 수 없는 이유가 이 함수의 AST에 있다.**
2라운드 H2가 이 자리다.

## 이 산출물이 손으로 읽은 증거와 다른 점

"래치가 없다"는 **부재에 관한 주장**이다. 손으로 읽으면 볼 곳을 골라서 못 본 것인지
정말 없는 것인지 구별되지 않는다. `ast.json`은 그 자리에 **`"branches": null`과
`"returns": null`** 을 내놓는다(빈 목록의 직렬화가 `[]`가 아니라 `null`이다 — 파일에
있는 리터럴 그대로다). 조건문이 0개이므로 조기 반환도, 중복 억제도, 어떤 형태의
게이트도 **없다.**
`.claude/CLAUDE.md`가 "손으로 읽은 증거는 볼 곳을 고르므로 선택적"이라고 말하는 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m managed` | 관리 포지션 | 호출자 `judge` | — |
| `proposal` | `exitpolicy.Proposal` | 정책 판정 | — |
| `detail` | 거부 사유 문자열 | 호출자 | — |
| **억제 상태** | **없음** | — | **없음 — 이 함수는 상태를 보지 않는다** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | `Notify` 도달 |
|---|---|---|---|---|
| — | **없음(AST branches 0)** | — | **없음(AST returns 0)** | — |
| — `:1550` | 무조건 | **`o.alert(...)` — `Notify` 도달 (P4)** | 암묵 `:1565` | ✅ **항상** |

**호출 1회 = `Notify` 1회.** 조건이 없으므로 예외가 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.alert` `:1550` | 제출되지 않은 청산 통지 | **동기·기한 없음.** critical(`event.go:295` `EventExitProposalRefused: true`) | `internal-app-engine--exitobserver.alert` FLM |
| `string` `:1552`·`:1553`·`:1560` | Key/Fields 조립 | 즉시 | AST calls |
| `o.label` `:1554` | 종목 표시명 | 즉시 | AST calls |
| `fmt.Sprintf` `:1555` | Body 조립 | 즉시 | AST calls |

## 사이클 총 체류가 유계가 아닌 이유

| 함수 | 억제 | AST 근거 |
|---|---|---|
| `alertRefused` `:1516` | **래치** `o.refused[ID]` | branches 2, returns 1 |
| `alertUnmanaged` `:1495` | **래치** `o.unmanaged[ID]` | 해당 FLM |
| `announceQuarantine` `:59` | **래치** `o.quarantineAnnounced[key]` | branches 2, returns 1 |
| `noteDelay` `:1569` | **래치** `o.delayAlerted[ID]` | 해당 FLM |
| **`alertProposalRefused` `:1548`** | **없음** | **branches 0, returns 0** |

`ObserveOnce`는 `range states`(B5 `:453`)로 포지션마다 `judge`를 돌린다. `judge`가
포지션마다 이 함수를 부를 수 있으므로 **한 사이클의 `Notify` 횟수 상한은 포지션 수다.**
그래서 a092의 계약은 **알림 하나당** 예산이고, 사이클 총 체류는 유계가 되지 않는다.
사이클 총 상한은 a093이 배달을 루프 밖으로 빼면서 다룬다.

## State mutations and fallbacks

- **상태 변경 없음.** AST assignments 0.
- fallback 없음. goroutine 없음, defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: **yes** — §0.5. 이 알림이 늦으면 "청산 주문이 브로커에 닿지
  못했다"는 사실이 늦게 알려진다.
- **a092가 여기서 얻는 사실**: spec은 "사이클 총 체류"를 약속하면 **안 된다.**
  약속할 수 있는 것은 알림 하나의 체류뿐이다.
