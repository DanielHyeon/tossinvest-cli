# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 산출물이 a099의 설계 하나를 뒤집었다.**
>
> a099의 초안은 `MarkAlertAttemptFailed`가 임차를 푼다고 했다. 이 함수의 AST가
> 그것을 반증한다: B8 `:384`가 실패를 기록하고, **B9 `:387`이 예산이 남았는지 보고
> B10 `:388`이 대기한 뒤 B2 `:349`의 루프가 `:354`로 되돌아간다.**
>
> **실패 기록은 발송자가 끝났다는 뜻이 아니다.** 거기서 임차를 풀면 `:388`의 대기 동안
> 두 번째 발송자가 그 행을 집고, 원래 발송자는 임차 없이 `:354`를 다시 실행한다.
> **a099가 막으려는 이중 발송을 a099가 만든다.**
>
> 그래서 해제는 **루프 밖**(`:394` 이후)이다 — design D3의 ⚠⚠.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | claim에서 받은 행 | 호출자 (`claimAndDeliver:276`) | — |
| `n.Attempts` | `<= 0`이면 기본값 | B1 `:343` → `DefaultCriticalAttempts` = 3 (`notifier.go:45`) | — |
| `n.Publisher` | nil 허용 | 배선 | B3 `:350` → `lastErr` 설정 후 `break` |
| `n.RetryDelay` | `<= 0`이면 기본값 | `wait` `:411-413` → `DefaultRetryDelay` = 2초 (`:48`) | — |
| **`n.mu`** | **호출자가 잡고 있어야 한다** | **doc `:336-340`이 PRECONDITION으로 명시** | 안 잡으면 이중 발송 — 그 doc이 그렇게 적는다 |
| **임차** | **오늘 없다** | — | a099가 이 함수 밖에서 취득하고 이 함수 끝에서 푼다 |

**doc comment `:336-340`이 이 change의 가장 직접적인 증거다**:
*"PRECONDITION: the caller holds n.mu, and holds it across the claim that produced id as
well as this send. … or two observations of the same condition each conclude the send is
owed and each publish."* — **배제를 뮤텍스에 맡긴다고 코드가 스스로 적는다.**

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 12 · 이탈 3 · 호출 16 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:343` | `attempts <= 0` | 기본값 3 | — | 기존 |
| B2 `:349` | `for attempt := 1; attempt <= attempts` | **재시도 루프** | — | **a099 R6** |
| B3 `:350` | `n.Publisher == nil` | `lastErr` · `break` | — | 기존 |
| B4 `:355` | publish 성공 | `MarkAlertDelivered` `:356` | — | 기존 |
| B5 `:357` | 정산 성공 | — | **이탈 `:358` `true`** | 기존 |
| B6 `:373` | `n.Log != nil` (정산 실패 경로) | 로그 | — | 기존 (a096 r2) |
| B7 `:378` | `n.Gate != nil` (같음) | **게이트 래치** | 이탈 `:381` `false` | 기존 |
| **B8 `:384`** | **`MarkAlertAttemptFailed`가 오류** | 로그만 | — | **a099 R6 — 이 자리에서 임차를 풀면 안 된다** |
| **B9 `:387`** | **`attempt < attempts`** | — | — | **a099 R6 — 예산이 남았다** |
| **B10 `:388`** | **`!n.wait(ctx)`** | 대기 | `break` | **a099 R6 — 이 대기 동안 행이 무주공산이면 안 된다** |
| B11 `:398` | `n.Log != nil` (예산 소진) | 로그 | — | 기존 |
| B12 `:403` | `n.Gate != nil` (같음) | **게이트 래치** | — | 기존 |
| — 이탈 `:406` | 예산 소진 | 행은 PENDING으로 남는다 (`:394-395`의 주석) | `false` | **a099 R7 — 여기 직전에 임차를 푼다** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Publisher.Publish` `:354` | **네트워크 발송** | 기한은 Publisher가 정한다 — `Ntfy.Timeout` 기본 10초 (`ntfy.go:72`) | `ast.json` calls |
| `n.Journal.MarkAlertDelivered` `:356` | 정산 | 실패면 B6·B7 — **게이트를 잠그고 `false`** | 같음 |
| `n.Journal.MarkAlertAttemptFailed` `:384` | 시도 기록 | 오류는 로그만 (B8) | 같음 |
| `n.wait` `:388` | 시도 사이 대기 | ctx 취소면 `false` → `break` | 같음 |
| `n.Gate.Block` `:379`, `:404` | 진입 래치 | 메모리 맵 (`execgw/retry.go:495-510`) | 같음 |

**한 사이클의 상한** — 세 값의 곱과 합이다:

| 항목 | 값 | 근거 |
|---|---|---|
| 시도 | 3 | `DefaultCriticalAttempts` `notifier.go:45` |
| 한 시도 | 10초 | `Ntfy.Timeout` 기본 `ntfy.go:72` |
| 대기 | 2초 | `DefaultRetryDelay` `notifier.go:48` |
| **실패 기록의 쓰기 대기** | **5초 × 3** | `defaultBusyTimeout` `journal.go:32` |
| **해제의 쓰기 대기** | **5초** | 같음 |
| **상한** | **3×(10+5) + 2×2 + 5 = 54초** | 1라운드 B-P2 |

**임차는 이 값보다 길어야 한다** (design D4·C4). 한 시도(10초)가 아니다.

> **⚠ 이 표는 3판까지 `3×10 + 2×2 = 34초`라고 적혀 있었다.** SQLite 쓰기 대기를
> 빠뜨린 값이고, design D4가 1라운드 B-P2로 이미 고친 것을 **이 산출물이 안 따라왔다.**
> 34초 임차를 쓰면 발송자가 **자기 문서화된 예산을 쓰는 도중에** 임차를 잃는다.

## State mutations and fallbacks

- 성공: 행이 DELIVERED로 간다 (`:356`).
- 정산만 실패: **게이트를 잠그고 `false`** — a096 round 2가 만든 경로. 푸시는 나갔는데
  기록이 없으므로 「계산 안 되는 발송은 미정산」으로 다룬다 (`:360-369`의 주석).
- 예산 소진: 행은 **PENDING으로 남고** 게이트가 잠긴다 (`:394-395`).
- **폴백 없음.** 세 이탈 중 둘이 게이트를 잠근다 — 닫힌 쪽으로 실패한다.

## Safety conclusion

- **Safe edit boundary**: a099는 **이탈 `:406` 직전에 `ReleaseAlertClaim` 한 번**을
  더하고 `:356`·`:384`의 호출에 claim token을 넘긴다. **루프 안(B8~B10)에서는 아무것도
  풀지 않는다.** 열두 분기의 조건과 세 이탈의 반환값은 안 바꾼다.
  편집 후 AST의 branches가 12 그대로면 제어 흐름 무변화다.

  > **⚠⚠ 이탈 `:381`은 해제 자리가 **아니다** — 1라운드 B-P3이 뒤집었다.**
  > 이 산출물의 1판은 *"publish 성공 + 정산 실패에서도 풀어야 한다"*고 적었다.
  > **그것이 a096 폭풍을 되살린다**: 푸시는 나갔는데 정산이 안 됐고 행이 PENDING인데
  > 임차까지 없으면, 다음 관측(a096에서 5.6초마다)이 다시 claim해서 **다시 보낸다.**
  > 한 행이 다시 예순 번이 된다.
  >
  > 이 경로는 **「보냈는지 원장이 모른다」는 상태**이고, 그때 임차는
  > **억제 표시로 남아 있어야 한다.** 만료가 그것을 푼다.

  > **⚠⚠ B8~B10이 「아무것도 안 한다」로는 부족하다 — 1라운드 A-P4·B-P4.**
  > `:384`가 오류를 로그로만 다루고 루프가 계속 돈다. 소유자를 잃은 발송자가
  > **그대로 `:354`를 다시 실행한다.** 임차를 잃었다는 것은
  > **남은 publish를 즉시 중단시키는 결과**여야 하고, 그러려면 B8의 처리가 바뀐다 —
  > 즉 **제어 흐름이 바뀐다.** 위 「branches 12 그대로」는 §4 확정 후 다시 판정한다.
- **High-risk impact**: **yes** — 손절 경로가 `claimAndDeliver`를 지나 이 함수로 온다.
  a092의 전체 근거가 이 함수의 체류 시간이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B5 `:357`이 거짓인 경로**(publish 성공 + 정산 실패)에서 a099가 임차를 어떻게
    다루는지. 행은 PENDING으로 남고 게이트가 잠긴다. **임차는 안 푼다** — 위 ⚠⚠가
    이유를 적는다. 이탈 `:381` 직전은 **해제 자리가 아니고**, 임차는 억제 표시로
    남아 만료가 푼다. **§4.5b의 「두 자리」가 맞다.**

    > **⚠⚠ 3판까지 이 항목은 정확히 반대로 적혀 있었다** — *"임차는 풀어야 한다 …
    > §4.5b가 두 자리라고 적었는데 실제로는 셋이다"*. **같은 파일 스무 줄 위의 ⚠⚠가
    > 그것을 이미 반증하고 있었다.** 구현자가 이 줄을 따르면 a096 폭풍이 성공 경로로
    > 돌아온다. 3라운드가 잡았다 — **정본 절을 만들어도 사본이 살아남는다는 증거다.**
  - **`n.mu` 구간** — a099는 안 건드린다. **`not-applicable`: a092의 표면이다.**
  - **`Publish`의 실제 기한** — Publisher 구현이 정하고 이 함수는 모른다.
    `Ntfy` 말고 다른 Publisher가 배선되면 위 54초 유도가 깨진다. §3.4가 실측한다.
