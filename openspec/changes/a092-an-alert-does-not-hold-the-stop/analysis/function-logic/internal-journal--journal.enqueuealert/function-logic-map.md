# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go` (115-122)
- AST evidence: `ast.json` — branches 0, returns 1, calls 1, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다.

> **이 파일은 18라운드 B-P1 때문에 다시 썼다 — 그리고 그 이유가 이 change의 교훈이다.**
>
> 17판까지 이 지도는 분기 9개·이탈 9개짜리 함수를 서술했고 B1·B2·B5에 GREEN을 달았다.
> 그 함수는 **더 이상 없다.** a097(`285c7619`)이 중복 억제 로직을 전부
> `ClaimAlertForDelivery`로 옮겼고, 남은 것은 인자 하나를 0으로 고정해 위임하는
> **8줄짜리 래퍼**다. `ast.json`은 그때 다시 뽑혔고 `"branches": null`을 정확히 담고
> 있었다. **산문만 재고정되지 않았다.**
>
> `check_analysis.py`는 이것을 통과시켰다 — 파일이 넷 다 있고, 해시가 최신이고,
> AST가 요구하는 분기(0개)가 전부 덮였기 때문이다. 여분의 B1~B9 행도, `(111-151)`이라는
> 좌표도, `branches 9`라는 주장도 어떤 코드와도 대조되지 않았다.
> 그 세 가지를 검사하는 것이 이제 검사기에 있다.
>
> **그리고 이 파일이 a092의 나머지 13개 산출물을 쓸 때 쓴 템플릿이었다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 호출자의 것 | 호출자 | 그대로 `ClaimAlertForDelivery`로 전달 |
| `a` | `EventKey`·`Type`가 비어 있지 않아야 한다 | 호출자 | **이 함수는 검사하지 않는다** — 검사는 `ClaimAlertForDelivery` B1·B2가 한다 |
| `remindAfter` | **항상 0** | 이 함수가 고정 `:120` | 0은 "재알림 정책 없음"이고, 정착된 행을 남의 이름으로 재무장하지 않겠다는 선언이다(주석 `:116-119`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 (분기 없음) | — | 없음 — 위임뿐 | `return id, err` `:121` |

**분기가 0이라는 것이 이 지도의 결론이다.** 조건도 early return도 없다.
`:120`의 `j.ClaimAlertForDelivery(ctx, a, 0)` 한 줄이 함수 전체이고, 두 번째 반환값
(`owed` — 이 발송이 아직 빚인가)은 **버려진다**.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.ClaimAlertForDelivery` `:120` | 기록·중복 억제·상태 판정 전부 | 로컬 SQLite 트랜잭션 — 밀리초. 네트워크 없음 | AST calls + `internal-journal--journal.claimalertfordelivery/ast.json` |

**네트워크 없음.** 이 함수의 체류는 34초 예산에 포함되지 않는다.

### 중복 억제의 근거는 이제 여기가 아니다

`analysis/delivery-latency.md`의 측정 해석(같은 조건이 다시 관측되면 새 행이 생기지 않고
기존 id가 돌아온다)은 **`ClaimAlertForDelivery` B5**가 진다 —
`internal-journal--journal.claimalertfordelivery/function-logic-map.md`를 볼 것.
그 성질을 이 함수에서 읽으면 오늘은 우연히 맞고 다음 리팩터에서 틀린다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` | **이 함수 안이 아니다** | 전부 `ClaimAlertForDelivery`가 한다 |

- fallback 없음. 오류는 그대로 반환된다.
- **`owed`를 버리는 것이 유일한 정보 손실이다.** 그래서 이 함수의 호출자는
  "기록만 하면 되는" 자리여야 한다. 발송을 앞둔 호출자가 이것을 쓰면
  *이 발송이 아직 빚인가*를 알 수 없다 — 주석 `:112-114`가 그것을 계약으로 쓴다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: 간접. 내구성 자체는 `ClaimAlertForDelivery`가 진다.
- **이 함수가 a092에 대해 갖는 의미는 하나다**: `execgw.parkAlert`(`replay.go:551`)가
  이것을 쓰고 `owed`를 버리므로, **누가 그 행을 보낼 것인가가 이 함수 밖에서 정해져야
  한다.** 오늘은 아무도 정하지 않았고 그것이 design D0.1의 결함이다.
