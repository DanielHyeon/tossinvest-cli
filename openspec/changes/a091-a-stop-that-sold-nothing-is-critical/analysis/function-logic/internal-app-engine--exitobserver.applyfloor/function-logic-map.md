# Function Logic Map: `ExitObserver.applyFloor`

- Source: `internal/app/engine/exitloop.go` (1403-1447)
- AST evidence: `ast.json` — branches 6, returns 7, defers 0
- Risk scan: `risk-pattern-report.md`
- 호출자: **`ExitObserver.submit` 하나뿐** (codegraph callers + grep 일치)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.Floor` | nil 허용 | 주입 | nil이면 무캡(B1) |
| `floor.Quantity` | 십진 문자열, **0 가능** | `ConfirmedFloor` (RECONCILE 확정 하한) | 0이면 아무것도 제출되지 않는다 |
| `floor.Bound` | 사유 문자열 | 같음 | 알림 `floor_bound` 필드 |
| `quantity` | 판정이 투영한 수량 | `snapshot.ProjectedQuantity` | — |
| **호출 맥락(보호/익절)** | — | **전달되지 않는다** | ⚠ 이 함수는 제안이 손절인지 익절인지 모른다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 제출 수량 |
|---|---|---|---|---|
| B1 `:1404` | `Floor == nil` | 없음 | `:1405` | 원안 |
| B2 `:1408` | `ConfirmedFloor` 오류 | **`logErr`만** (알림 없음) | `:1414` `"0", true` | **0** |
| B3 `:1416` | `!applies` | 없음 | `:1417` | 원안 |
| B4 `:1420` | `CompareDecimal` 오류 | 없음 | `:1421` `"", false, err` | — |
| B5 `:1423` | `cmp >= 0` (하한이 충분) | 없음 | `:1424` | 원안 |
| B6 `:1427` | `SubDecimal` 오류 | 없음 | `:1428` `"", false, err` | — |
| — `:1446` | 하한이 원안보다 작다 | **`alert(EventExitProposalCapped)`** | `floor.Quantity, true` | **0일 수 있다** |

### 0주가 되는 경로는 둘이고, 등급이 서로 다르다

| 경로 | 조건 | 남는 것 |
|---|---|---|
| B2 `:1414` | 하한을 **계산할 수 없다** | `logErr` 한 줄. **알림조차 없다** |
| `:1446` | 하한이 **0을 허용한다** | `EventExitProposalCapped` — **`severity: normal`** |

둘 다 `submit`의 `isZeroQuantity` 분기(`:1243`)로 가서 조용히 `release`된다.

### 알림 본문이 사실과 다르다

```go
Title: "… 청산이 확정 하한에 걸려 일부만 나갔다"
Body:  "%s 제안 · … 허용하는 수량 %s (%s) · 잔여 %s는 매도되지 않는다."
```

`floor.Quantity == 0`이면 나간 수량은 **0**인데 제목은 "일부만 나갔다"이다.
`remainder == quantity`이므로 본문의 수치는 맞지만 제목이 반대를 말한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Floor.ConfirmedFloor` `:1407` | RECONCILE 확정 하한 | 오류 → B2(0주, fail-closed) | AST |
| `o.logErr` `:1412` | B2의 유일한 기록 | 삼킴 | AST |
| `riskcalc.CompareDecimal` `:1419` | 하한 vs 원안 | 오류 → B4 | AST |
| `riskcalc.SubDecimal` `:1426` | 잔여 계산 | 오류 → B6 | AST |
| `o.alert` `:1430` | 캡 알림 | 삼킴. **등급은 `SeverityOf`가 정한다** | AST |

브로커 접촉 없음.

## State mutations and fallbacks

- 이 함수는 상태를 바꾸지 않는다. 반환값과 알림이 전부다
- **fallback이 fail-closed다**: 하한을 못 구하면 0으로 본다(`:1409-1411` 주석).
  방향은 옳다 — 문제는 그 사실이 **critical로 보고되지 않는 것**이다

## Safety conclusion

- **Safe edit boundary**: 등급·문구·이벤트 종류 변경은 반환값에 영향이 없다.
  제출 수량 계산(B1~B6, `:1446`)은 **건드리지 않는다** → §0.9 무관
- **High-risk impact**: **yes** — 손절이 0주로 깎이는 유일한 지점
- **이 함수는 보호/익절을 모른다**: 등급을 여기서 나누려면 호출자가 맥락을 넘겨야 한다.
  `submit`은 `proposal`을 갖고 있으므로(`:1237` 시그니처) 전달 가능하다
