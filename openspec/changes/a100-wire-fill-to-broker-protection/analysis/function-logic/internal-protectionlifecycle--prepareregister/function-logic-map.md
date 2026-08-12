# Function Logic Map: `prepareRegister`

- Source: `internal/protectionlifecycle/lifecycle.go` (L3-33)
- AST evidence: `ast.json` — 분기 6, return 7, 호출 11
- Risk scan: `risk-pattern-report.md`

> 설계 문서 §8.2의 2단계("체결 수량만큼 stop/conditional 보호 주문 제출")를 준비하는 함수다.
> 실제 제출은 반환된 `BrokerCommand`를 브로커 어댑터가 수행한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state`, `key` | position 존재, capability 유효 | `mutablePosition` | error 즉시 반환 |
| `key.Market` | entry가 열린 시장 | `state.marketEntryOpen` | `RefusalEntryLatched` |
| `position.Phase` | `Unprotected` 또는 `Terminal` | position 상태 | `RefusalEntryLatched` |
| `position.HasPending` | false | position 상태 | `RefusalOperationPending` |
| `position.Observed.Status` | `!= BrokerActive` | 브로커 관측 | `RefusalOperationPending` |
| `capability.exactOperationLookup` | true | 브로커 능력 매트릭스 | `RefusalInvalidObservation` |
| `quantity`, `trigger` | 둘 다 0 아님 **and** `quantity + OtherSellClaims == Holdings` | 호출부 + position | `RefusalSellClaimExceeded` |

**불변식**: 보호 수량은 **보유 수량을 정확히 채워야 한다** — 부족해도 넘쳐도 거부다(B6).
설계 문서 §8.2의 4단계가 제출 *이전*에 이미 강제된다.

**불변식 2**: 브로커가 정확한 operation 조회를 지원하지 못하면(B5) 보호를 시도조차 하지 않는다.
능력 매트릭스가 판정에 실제로 쓰이는 지점이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| B1 (L5) | `mutablePosition` 에러 | 없음 | `(state, {}, err)` |
| B2 (L8) | entry 미개방 ∨ `EntryLatch != ""` ∨ Phase가 Unprotected/Terminal 아님 | 없음 | `RefusalEntryLatched` |
| B3 (L11) | `position.HasPending` | 없음 | `RefusalOperationPending` |
| B4 (L14) | `Observed.Status == BrokerActive` | 없음 | `RefusalOperationPending` ("protection already active") |
| B5 (L17) | `!capability.exactOperationLookup` | 없음 | `RefusalInvalidObservation` |
| B6 (L20) | `quantity==0 ∨ trigger==0 ∨ quantity+OtherSellClaims != Holdings` | 없음 | `RefusalSellClaimExceeded` (수치 3개 포함) |
| fall-through (L32) | 정상 | 아래 mutation 전부 | `(next, command, nil)` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry 계약 | Evidence |
|---|---|---|---|
| `mutablePosition` | position + capability 검사 | error | AST L4 |
| `state.marketEntryOpen` | 시장 단위 entry latch | bool | AST L8 |
| `cloneState` / `next.reseal` | 불변 갱신 | 순수 | AST L23 / L31 |
| `operationIdentity` | 멱등 키 합성 (market·generation·revision·kind) | 순수 | AST L27 |
| `commandOrder` | Desired 주문 구성 | 순수 | AST L28 |

`operationIdentity`가 설계 문서 §8.1의 "client order id 멱등 생성"에 해당한다.

## State mutations and fallbacks

- `position.ProtectionRevision++`
- `command` 조립 후 `command.OperationKey = operationIdentity(...)`
- `position.Desired = commandOrder(command, "", BrokerUnknown)`
- `Pending, HasPending, Phase, EntryOpen, EntryLatch =
   command, true, SubmitPending, false, RefusalOperationPending`

즉 **보호 제출을 준비하는 순간 entry가 닫힌다**(`EntryOpen=false`, `EntryLatch=OperationPending`).
설계 문서 §8.2의 5단계보다 보수적이다 — 불일치가 아니라 *제출 중*에도 닫는다.

fallback 없음. 전부 fail-closed.

## Safety conclusion

- Safe edit boundary: 본문 편집 불필요. a100은 호출 경로와 브로커 어댑터 배선을 담당한다.
- High-risk impact: **yes** — 보호주문 제출의 유일한 준비 지점.
