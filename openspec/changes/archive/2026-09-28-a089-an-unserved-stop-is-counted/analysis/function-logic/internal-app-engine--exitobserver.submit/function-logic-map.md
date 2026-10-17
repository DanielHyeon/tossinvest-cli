# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (1237-1312)
- AST evidence: `ast.json` — branches 11, returns 9, calls 24, assignments 7
- Risk scan: `risk-pattern-report.md`
- 진입점은 하나뿐이다: `record:1195`. `record`가 B6·B12·B13·B14로 끝나면 여기 도달하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `quantity` | 판정이 투영한 수량 | `snapshot.ProjectedQuantity` | 하한에 깎일 수 있다 |
| `o.opts.Floor` | nil 허용 | 주입 | nil이면 무캡(`:1404-1406`) |
| `observed` | 관측가 문자열 | 판정 | 빈 값이면 `m.state.Baseline`, 둘 다 비면 **거부**(`sellIntent`) |
| `o.opts.Submit` | non-nil | execgw | `Outcome.State`가 4갈래 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 주문? | 기존 테스트 |
|---|---|---|---|---|---|
| B1 `:1240` | `applyFloor` 오류 | 없음 | `return err` `:1241` | ❌ | **없음** |
| B2 `:1243` | `isZeroQuantity(submitQuantity)` | `release(Refused)` | `:1248` | ❌ | `:953`,`:982` ✅ |
| B3 `:1263` | `IssueReduction` 오류 | `alertProposalRefused` + `release` | `:1265` | ❌ | 간접 |
| B4 `:1272` | `AttachExitIntent` 오류 | 없음 | `return err` `:1273` | ❌ | **없음** |
| B5 `:1277` | `sellIntent` 오류 | `alertRefused` + `release` | `:1279` | ❌ | **없음** |
| B6 `:1287` | `switch` | — | — | — | — |
| B7 `:1288` | `State == StateConfirmed` | `log(OrderConfirmed)` | `return nil` `:1295` | ✅ **살아남** | 다수 |
| B8 `:1296` | `InDoubt \|\| UnresolvedInDoubt` | 없음 (**제안 유지**) | `return nil` `:1300` | ⚠ **미상** | `:1036` ✅ |
| B9 `:1301` | `Reason == ReasonSymbolInFlight` | **`noteDelay`** + `release(Cancelled)` | `:1303` | ❌ | 간접 |
| B10 `:1304` | default (브로커 거부) | `alertProposalRefused` + `release(Refused)` | `:1310` | ❌ | `:1013` ✅ |
| B11 `:1306` | `detail=="" && err!=nil` | `detail` 보정 | — | — | 간접 |

**이탈 9개 중 주문이 살아나는 것은 B7 하나**, 존재 미상이 B8 하나, 나머지 7개는 주문 없음.

## Calls and live bindings

§0.4 — 이 함수가 브로커에 닿는 횟수:

| 호출 | 대상 | 비고 |
|---|---|---|
| `applyFloor` → `Floor.ConfirmedFloor` `:1407` | 원장/캐시 | 브로커 직접 호출 아님 |
| `Issuer.IssueReduction` `:1251` | Guardian (로컬) | 브로커 아님 |
| `Journal.AttachExitIntent` `:1272` | 원장 | 브로커 아님 |
| **`Submit.Place` `:1281`** | **브로커 mutation** | **주기당 최대 1회** |

계측을 더해도 이 표는 바뀌지 않는다 — a089가 "요청 수 무변경"이라고 쓴 근거는 여기다.

## B2가 2026-08-02 사건의 경로다 (실측)

`applyFloor`가 `floor.Quantity`를 돌려주고(`:1446`) 그것이 0이면 B2가 조용히 `release`한다.
그 직전 `applyFloor`가 `EventExitProposalCapped`를 올리지만 **`criticalEvents`에 없다**
(`obs/event.go`) → `severity: normal` → **outbox 행 없음 · 게이트 무반응 · 재전달 없음**.

실측(`journal.db` / `engine.log`):

```text
2026-08-02 23:23:25 ~ 23:26:21  pos-522745e0 (042660)
  STOP_LOSS_LADDER × 13 → 전부 PROPOSAL_REFUSED
  exit.proposal_capped × 13, severity=normal
  detail: "the RECONCILE confirmed floor authorises 0 (broker sellable quantity)"
  alert_outbox 행: 0
  종료: ADJUSTMENT_CLOSED — 손절은 끝내 나가지 않았다
```

알림 제목은 "청산이 확정 하한에 걸려 **일부만 나갔다**"인데 나간 수량은 **0**이다.

## State mutations and fallbacks

- `release(...)`가 B2·B3·B5·B9·B10에서 레벨을 재발의 가능으로 되돌린다(`:1314-1321`)
- **B7(`StateConfirmed`)은 `release`를 부르지 않는다** — `pending_action`이 남고,
  그래서 다음 주기의 `record` B12(`ErrProposalPending`)는 **살아 있는 보호 주문**을
  가진 포지션에서도 나온다
- B8은 아무것도 바꾸지 않는다. 제안이 무장된 채 남고 대사기가 해소한다
- fallback: B11이 빈 `detail`을 `err`로 채운다. 그 외 fallback 없음

### B9만 시계를 건드린다

`noteDelay`의 저장소 전체 호출자는 `record:1146`과 여기 `:1302` 둘뿐이다(AST + grep).
브로커 거부(B10)에는 없다. **다만 `record` FLM에 적은 대로, 한 줄 추가로는 발화하지 않는다** —
`record:1150`의 `clearDelay`가 submit보다 먼저 무조건 돌기 때문이다.

## Safety conclusion

- **Safe edit boundary**: B7~B10의 기록·계수 추가는 `Submit.Place`(`:1281`) **뒤**이므로
  제출 시점·내용에 영향이 없다. B1~B5는 제출 **전**이지만 전부 이미 제출을 포기한 경로다.
- **High-risk impact**: **yes** — 손절이 브로커로 나가는 유일한 함수.
- **B8을 계수·시계에 넣으면 안 된다**: `:1297-1299` 주석이 "**The order may exist.**
  The proposal stays armed"라고 명시한다. 존재 미상을 미제출로 세면 살아 있는 보호 주문에
  대해 거짓 알림이 뜬다.
- **미테스트 이탈 3개**: B1(`applyFloor` 오류) · B4(`AttachExitIntent` 오류) ·
  B5(`sellIntent` 오류). B5는 a087이 바꾸려는 바로 그 함수다 — a087 착수 전 RED 필요.
