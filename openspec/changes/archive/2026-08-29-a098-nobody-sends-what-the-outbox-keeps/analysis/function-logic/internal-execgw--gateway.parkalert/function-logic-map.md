# Function Logic Map: `Gateway.parkAlert`

- Source: `internal/execgw/replay.go` (534-559)
- AST evidence: `ast.json` — AST 기준 branches 2 / returns 0 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)

**a098은 이 함수를 편집하지 않는다.** 이 map이 있는 이유는 a098이 고치는 결함이
**여기서 시작하기 때문**이다 — 이 함수가 쓰는 행을 아무도 보내지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `g.entry` | nil 허용 | `gateway.go`의 조립 | B1 `:535` — nil이면 **래치를 안 세운다** |
| `rec` / `intent` | 이미 영속된 attempt | 호출자 `:358`이 `ResolveUnresolved` **성공 뒤에만** 부른다 | — |
| outbox 쓰기 | best-effort | `:551` `EnqueueAlert` | 반환값 **둘 다 버린다**(`_, _ =`) — 선언 주석이 이유를 적는다(`:530-533`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:535` | `g.entry != nil` | `:536` `entry.Block(ReasonUnresolvedInDoubt, …)` — **진입 게이트 래치** | 이탈 없음 | 기존 (아래) |
| B2 `:548` | `json.Marshal` 오류 | `:549` `payload = nil` | 이탈 없음 | 없음 — 아래 참조 |

이탈이 0이다. 함수는 void이고 **어떤 실패도 호출자에게 올라가지 않는다.**
그것이 설계다 — 선언 주석 `:530-533`이 *"a failure to enqueue does not un-park the
attempt — the park is already durable and is the safety property"*라고 적는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EntryGate.Block` | 진입 차단 | 오류 없음. 인메모리 맵 쓰기 | `:536` |
| `json.Marshal` | 운영자 문맥 | 실패하면 payload를 버린다 (B2) | `:540-547` |
| `Journal.EnqueueAlert` | outbox 기록 | **반환값 둘 다 버린다** | `:551` |

## State mutations and fallbacks

- **진입 게이트 래치**(인메모리)와 **outbox 행**(내구), 둘 다 만든다.
- `EnqueueAlert`(`outbox.go:115-122`)는 a097 이후 `ClaimAlertForDelivery(ctx, a, 0)`에
  위임하는 wrapper이고 `owed`를 버린다. **claim만 하고 발송하지 않는다.**
  그것이 이 자리에서는 옳다 — 주문 replay 경로 안에서 원격 전송을 기다릴 수 없다.
- **결함은 여기가 아니라 그 다음이다**: 이 행을 나중에 보내 줄 주체가 프로덕션에 없다.

## Safety conclusion

- Safe edit boundary: **a098은 이 함수를 안 건드린다.** 안전 성질(래치)은 이미 서 있다.
- High-risk impact: **yes** — 주문 경로. 그래서 a098은 발송을 **여기 넣지 않는다**(design D1 안 A 기각).
