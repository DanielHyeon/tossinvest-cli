# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :251–317, 분기 7 · 반환 5 · 호출 15, source_sha256 `0bc75668ff17…`, 추출 HEAD `b3f14925`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적: 착지 단위 ③(21.4 GREEN · 정본 요구 「exit 관측 goroutine이 기다리는 잠금은 원격 전송을 덮어서는 안 된다」). 오늘은 `n.mu.Lock` :254 · `defer n.mu.Unlock` :255 로 claim 부터 `deliver` :309(원격 전송)까지 잠금을 쥔다. 편집 뒤 claim · 그 판정까지만 잠금 안, `deliver` 는 잠금 밖. `deliver` 의 승격 포함 판정을 호출자에게 전달(반환값 하나 추가).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` · `e` | critical 사건 | `notifyCritical` | — |
| `n.mu` | 배제 잠금 — 기록 전용 입구 · `Acknowledge` 와 공유 | 알림기 | 편집 뒤 전송 동안 놓음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | claim 오류 (:263) | 로그 · `Gate.Block`(잠금 안, 무조건 — durable 기록 실패는 동기로) | `false,false,err` (:282) | `TestAClaimThatFailsBlocksNewEntries` |
| B2 | `n.Log != nil` (:276) | 로그 | — | 같음 |
| B3 | `n.Gate != nil` (:279) | `Block` | — | 같음 |
| B4 | `switch claim.Disposition` (:284) | — | — | 전부 |
| B5 | `ClaimSettled` (:285) | — | `false,false,nil` (:293) | `TestConcurrentObservationsOfOneConditionSendOnce` |
| B6 | `ClaimHeldElsewhere` (:294) | `logClaimHeld` | `false,false,nil` (:305) | `TestAHeldRowIsNotWhispered` |
| B7 | `deliver` 가 lost (:310) | — | `sent,false,nil` (:314) — 승격 없음 | `TestASenderThatLosesTheLeaseStopsAtOnce` |
| 종단 | — | `logClaimStolen` · `deliver` :309 | `sent,true,nil` (:316) | `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Journal.ClaimAlertForDelivery` :262 | 기록 + 임차(한 트랜잭션) | 오류는 B1 | AST |
| `n.deliver` :309 | 원격 전송 + 정산 — **편집 뒤 잠금 밖** | (sent, lost) + 편집 뒤 판정 | AST |

## State mutations and fallbacks

- claim 이 행을 기록하고 임차를 잡는다. B1 의 래치는 잠금 안에서 무조건 — 이 자리는 `Acknowledge` 와 같은 잠금 아래라 셈~해제와 겹치지 않는다(편집 뒤에도 그대로).

## Safety conclusion

- Safe edit boundary: claim · 분기 B1~B6 는 잠금 안 그대로. `deliver` 호출과 `logClaimStolen`만 잠금 밖으로. B6(남의 임차) 갈래는 편집 뒤 **동시 동기 발송**(같은 조건의 두 관측)에서도 도달 가능해진다 — `logClaimHeld` 등급 판정은 Manager 확인 대상(§24.6).
- High-risk impact: yes — critical 발송 경로의 잠금 범위.
