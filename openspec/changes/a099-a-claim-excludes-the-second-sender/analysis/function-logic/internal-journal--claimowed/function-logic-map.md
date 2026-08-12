# Function Logic Map: `claimOwed`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099의 반증은 이 함수의 이탈 한 줄이다.** B2 `:276` → 이탈 `:278`
> `return true, false`. 두 번째 값이 `false`이므로 호출자의 B6 `:197`이 거짓이 되고,
> **PENDING 행에는 UPDATE가 하나도 실행되지 않는다.**
>
> a099는 이 함수를 **편집하지 않는다.** 임차 판정은 상태 판정과 다른 질문이고
> (*"보내야 하는가"* vs *"내가 보내도 되는가"*), 이 함수는 앞의 것만 진다.
> 그러나 **분기를 근거로 쓰므로 산출물이 먼저다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | 셋 중 하나, **CHECK 제약 없음** | `schemaV3` `outbox.go:57` | 미지의 값은 B8 `:308` → 이탈 `:313` `true, true` |
| `deliveredAt`·`acknowledgedAt` | RFC3339 또는 NULL | 같은 테이블 | 파싱 실패는 `latestStamp`가 건너뛴다 → B5 `:284` |
| `now` | 호출자의 `j.clk.Now()` | `outbox.go:179` | 뒤로 간 시계는 B6 `:290`이 잡는다 |
| `remindAfter` | `<= 0`이면 재무장 없음 | 호출자 | B4 `:280` → 이탈 `:281` `false, false` |

**순수 함수다.** mutation도 I/O도 없다 — 호출 둘(`latestStamp@:283`, `now.Sub@:289`)뿐이다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 8 · 이탈 7 · 호출 2.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:275` | `switch state` | 없음 | — | — |
| **B2 `:276`** | **`case AlertPending`** | 없음 | **이탈 `:278` `true, false`** | **a099 R1 — 이것이 반증의 근거다** |
| B3 `:279` | `case AlertDelivered, AlertAcknowledged` | 없음 | (아래) | 기존 (a096) |
| B4 `:280` | `remindAfter <= 0` | 없음 | 이탈 `:281` `false, false` | 기존 |
| B5 `:284` | 타임스탬프를 못 읽는다 | 없음 | 이탈 `:287` `true, true` — **열린 쪽** | 기존 (a096 round 2) |
| B6 `:290` | `elapsed < 0` — 시계가 뒤로 갔다 | 없음 | 이탈 `:302` `true, true` — **열린 쪽** | 기존 (a096 round 2) |
| B7 `:304` | `elapsed < remindAfter` | 없음 | 이탈 `:305` `false, false` | 기존 |
| — 이탈 `:307` | 창이 지났다 | 없음 | `true, true` — 재무장 | 기존 |
| B8 `:308` | `default` — 미지의 상태 | 없음 | 이탈 `:313` `true, true` — **열린 쪽** | 기존 |

**이탈 일곱 중 셋이 `true, true`이고 전부 「모르면 보낸다」쪽이다.**
a099의 임차 만료도 같은 방향을 따른다 (design D4) — 만료는 재발송 쪽으로 연다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `latestStamp` `:283` | 두 타임스탬프 중 최신 | 파싱 실패는 조용히 건너뛴다 (`outbox.go:318-334`) → B5가 그것을 열린 쪽으로 받는다 | `ast.json` calls |
| `now.Sub` `:289` | 경과 시간 | 음수 가능 — B6이 그것을 명시적으로 다룬다 | 같음 |

**live binding**: 유일한 호출자는 `ClaimAlertForDelivery:196`이다.

## State mutations and fallbacks

- **mutation 없음.** 순수 함수다.
- **폴백은 셋 다 「보낸다」쪽**: 못 읽는 타임스탬프(B5), 뒤로 간 시계(B6),
  미지의 상태(B8). 세 주석이 각각 그 이유를 적는다 — 억제하는 쪽이 더 나쁜 오류라는
  같은 논거다.

## Safety conclusion

- **Safe edit boundary**: **a099는 이 함수를 편집하지 않는다.** 임차 판정은
  `ClaimAlertForDelivery`의 트랜잭션 안에서, 이 함수의 반환을 받은 **뒤에** 일어난다.
  편집 후 이 함수의 `source_sha256`이 그대로여야 하고, 그것이 §5.3의 확인 대상이다.
- **High-risk impact**: **yes** — 이 함수가 `owed`를 정하고 `owed`가 발송을 정한다.
  편집하지 않는다는 것이 위험이 없다는 뜻은 아니다.
- **덮이지 않은 것을 이름으로 적는다**: 이 함수는 **동시성을 전혀 모른다.**
  두 호출이 같은 행에 대해 같은 답을 주는 것이 오늘 정상 동작이고,
  **그 사실을 이 함수 밖에서 고치는 것이 a099다.** 이 함수에 그 책임을 옮기려는
  설계는 거부한다 — 순수 함수에 트랜잭션 상태를 넣는 것이기 때문이다.
