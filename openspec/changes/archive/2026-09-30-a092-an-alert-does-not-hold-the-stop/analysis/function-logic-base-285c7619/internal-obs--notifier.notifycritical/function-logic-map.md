# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (170-225)
- AST evidence: `ast.json` — branches 4, returns 3, calls 11, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**이 함수가 이 change의 절단면이다.** 기록과 발송이 한 호출(`claimAndDeliver` `:194`)
안에서 연달아 일어나고, 앞의 것만 반환 전에 끝나야 한다.

> **18라운드 B-P1이 드러낸 것을 여기에도 적용했다.** 이 지도의 17판까지의 본문은
> `EnqueueAlert :177`과 `deliver :182`를 **두 개의 나란한 호출**로 서술했다. a097이
> 그 둘을 `claimAndDeliver` 하나로 합쳤고(`:194`), 산문은 재고정되지 않았다.
> 좌표만 옮기면 없는 절단선을 계속 가리키므로 **이 파일은 다시 썼다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 조립 (`exitwiring.go:76`은 항상 채운다) | nil이면 B1이 최선노력으로 강등 + 경고 |
| `n.Log` | nil 허용 | 조립 | B2가 nil을 흡수 |
| `e` | critical 등급이 확정된 이벤트 | `Notify:132` | — |
| `n.eventKey(e)` | 비어 있지 않은 문자열 | `eventKey:521` — `e.Key` 우선, 없으면 종류+attempt/order/symbol | 같은 조건이 같은 키를 만든다(중복 억제의 근거) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | durable? |
|---|---|---|---|---|
| B1 `:171` | `n.Journal == nil` | (B2) + `publishBestEffort(ctx, e, Critical)` `:180` | `return nil` `:181` | ❌ **강등** |
| B2 `:175` | B1 안 · `n.Log != nil` | `Log.Warn(EventAlertUndelivered, …)` `:176-178` | — | — |
| B3 `:195` | `claimAndDeliver`가 오류 | `n.escalate(ctx, e)` `:213` | `return fmt.Errorf(…)` `:214` | ❌ **행도 없다** |
| B4 `:217` | `owed && !sent` | `n.escalate(ctx, e)` `:222` | — | ✅ 행은 PENDING으로 남음 |
| — | 정상 | — | `return nil` `:224` | ✅ |

**이탈 3개 중 outbox 행이 남는 것은 B4 경로와 정상 경로 둘뿐이다.** B3은 claim 자체가
실패한 자리여서 게이트 래치도 행도 없다 — 주석 `:199-204`가 그것을 명시한다:
*"a claim that failed leaves no outbox row either, so a restart erases the block *and*
the evidence"*.

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `n.Log.Warn` `:176` | journal 미배선 경고 | 네트워크 없음 | AST calls |
| `n.publishBestEffort` `:180` | journal 없을 때의 강등 발송 | publish 1회, **10s** | AST calls + publishBestEffort FLM |
| `n.eventKey` `:185` | 중복 키 | 순수 문자열 조립 | AST calls |
| `encodeFields` `:190` | payload JSON | 실패 시 빈 문자열(`:546`) | AST calls |
| **`n.claimAndDeliver` `:194`** | **기록 + 발송을 한 번에** | **`n.mu`를 잡은 채 claim → publish → settle. 최대 34s.** 이 호출이 §0.3이 걸리는 자리다 | `internal-obs--notifier.claimanddeliver/ast.json` |
| `n.escalate` `:213`(claim 실패) · `:222`(전달 소진) | ENTRY_BLOCKED 강화 | 로컬 journal 쓰기. 주석 `:210-212`·`:218-221`이 "claimAndDeliver 밖"인 이유를 쓴다 — 그 함수가 `n.mu`를 잡고 있어 announcer가 재진입하면 교착 | AST calls |
| `fmt.Errorf` `:214` | 오류 포장 | — | AST calls |

### 절단선이 옮겨 갔다 — 그리고 그것이 a092의 문제다

17판까지 이 지도는 **`EnqueueAlert`와 `deliver` 사이**를 절단선으로 지목했다.
a097 이후 그 두 호출은 `claimAndDeliver` 하나이고, 그 안에서

| | 대상 | 비용 | 잠금 |
|---|---|---|---|
| `ClaimAlertForDelivery` (`claimAndDeliver:244`) | 로컬 SQLite | 밀리초 | `n.mu` 보유 |
| `n.deliver` (`claimAndDeliver:276`) | 원격 HTTPS | 최대 34s | **같은 `n.mu` 보유** |

가 **하나의 임계 구역 안에 있다.** 그래서 a092의 절단은 이 함수가 아니라
`claimAndDeliver`의 `:276` 한 줄에서 일어나야 한다 — design D0.3a가 그 자리를 정한다.

## State mutations and fallbacks

- `record`(`:184-191`)는 지역 값이고, 내구 기록은 `claimAndDeliver`가 만든다.
  중복 키에서 기존 행의 id가 돌아오는 성질은 `ClaimAlertForDelivery` FLM의 B5가 갖는다.
- fallback: B1이 journal 부재를 최선노력으로 강등하되 **로그로 알린다**(조용한 강등 금지).
- **goroutine 없음**(`go_statements: 0`).

## Safety conclusion

- **Safe edit boundary**: 이 함수 안이 아니다. 절단은 `claimAndDeliver` 안의
  claim(로컬)과 publish(원격) 사이이고, 이 함수는 **그 결과 두 값(`sent`·`owed`)의
  해석만** 담당한다.
- **High-risk impact**: **yes** — exit 관측 루프가 `:194`에서 최대 34초 머문다.
- **자르면 깨지는 것 셋**(이 change가 보존해야 할 계약):
  1. **B3과 B4의 escalate가 서로 다른 사건이다** — B3은 *기록 실패*(행 없음),
     B4는 *전달 실패*(행 PENDING). 하나로 합치면 재기동 후 복구 가능성이 다른 두
     상태가 같은 것으로 보고된다.
  2. **재진입 금지** — 주석 `:210-212`·`:218-221`. escalate가 `n.mu` 밖에 있는 것은
     정리가 아니라 교착 회피다.
  3. **B1의 강등 경로** — journal이 없으면 durable 큐 자체가 없으므로 비동기화의
     전제가 사라진다. 이 경로는 지금 형태를 유지해야 한다.
