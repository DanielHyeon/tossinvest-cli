# a099 design — claim이 두 번째 발송자를 배제한다

base: `285c7619110d2f8c53a1d9ddfbadd16ad0e9e53e` (`base-commit.txt`)

## D0 — 이 change가 반증하는 것

> **⚠⚠ 1라운드 A-T1이 이 절의 인용을 정정했다.**
> 아래 문장은 a092의 **현재 D0.3이 아니다.** a092는 18판(D0.3a)과 19판(D0.3b)에서
> **잠금 유지**로 이미 물러섰다. 저 문장은 a092의 **17판 잔재**이고,
> `a092/analysis/function-logic/internal-journal--journal.markalertdelivered/`의
> function-logic-map에 아직 살아 있다.
>
> **a099가 반증하는 것은 a092의 현재 결론이 아니라 그 잔재다.**
> a092 쪽 잔재 정리는 §7.1이 진다.

17판이 쓴 문장:

> 배제의 근거가 Go 뮤텍스에서 SQL 술어로 옮겨간다.

**오늘 그 SQL 술어는 없다.** `ClaimAlertForDelivery`의 AST가 그것을 열거한다
(`analysis/function-logic/internal-journal--journal.claimalertfordelivery`,
분기 11 / 이탈 10, `outbox.go:169-261`): 기존 PENDING 행 경로에서 B6 `:197`이 거짓이면
UPDATE가 하나도 실행되지 않고 이탈 `:240`으로 나간다. `claimOwed`의 B2 `:276` →
이탈 `:278` `return true, false`가 그 조건을 확정한다.

배제하는 술어로 오해되기 쉬운 것이 `MarkAlertDelivered`의 CAS인데, 그것은
**publish 이후에 돈다**. `claimAndDeliver` 이탈 `:276`이 `deliver`를 부르고, `deliver`가
publish 성공 뒤 `notifier.go:356`에서 `MarkAlertDelivered`를 부른다. 그 CAS가 거르는 것은
이중 **정산**이다.

### 이 반증이 어디까지 번지는가

| 자리 | 무엇을 D0.3 위에 세웠나 | a099 이후 |
|---|---|---|
| a092 D0.3a·D0.3b | 잠금 구간 축소의 정당화 | **성립한다** — a099가 술어를 만든 뒤에 |
| a092 §8.7 · D7 | `claim → release → publish → 재취득 → 정산` | 그 `claim`이 진짜여야 한다. **a099가 선행 조건이다** |
| a092 R17-3 | 두 claimer의 배제 관측 | a099의 §3이 그 관측을 먼저 만든다 |
| a098 공존 이야기 | 루프와 `claimAndDeliver`가 함께 산다 | **a099 없이는 두 발송자가 같은 행을 publish한다** |

**a092가 틀린 것이 아니라 순서가 틀렸다.** D0.3이 서술한 최종 상태는 옳고,
그 상태로 가는 첫 걸음이 a092 안에 없었다.

## D-CORE — 정본: 하나의 상태기, 하나의 API, 하나의 스키마

> **2라운드가 BLOCK한 이유는 개별 결함이 아니라 패턴이었다** (review §3.7).
> 1라운드가 준 자리를 고쳤는데 **같은 값의 사본**이 다른 절·tasks·map에 살아남았다.
> 두 보이스가 같은 처방을 냈다 — **API와 상태기를 먼저 하나로 확정하고 나머지를 유도하라.**
>
> **이 절이 그 정본이다. 아래 값은 이 절에만 적는다.**
> 다른 절과 tasks·spec·map은 이 절을 **가리키고, 다시 적지 않는다.**
> 값이 다른 곳에 다시 나타나면 그것은 사본이고, 사본은 이 change에서 두 번 사고를 냈다.

### C1 — 열은 **넷**이다 (사용자 결정 6-1)

```sql
-- schemaV31 (SchemaVersion 30 → 31)
ALTER TABLE alert_outbox ADD COLUMN claim_token      TEXT NOT NULL DEFAULT '';
ALTER TABLE alert_outbox ADD COLUMN claimed_by       TEXT NOT NULL DEFAULT '';
ALTER TABLE alert_outbox ADD COLUMN claimed_at       TEXT;
ALTER TABLE alert_outbox ADD COLUMN claim_expires_at TEXT;
```

| 열 | 무엇 | 비어 있음의 뜻 | **누가 읽는가** |
|---|---|---|---|
| `claim_token` | **취득마다 새로 만드는, 재사용 없는 값** | `''` = **임차 없음** | **모든 정산 CAS** |
| `claimed_by` | 운영자·디버그용 이름 (어느 루프인지) | — | 로그 · 운영자 목록 (D12) |
| `claimed_at` | **발급 시각** (UTC RFC3339) | `NULL` = 임차 없음 | **시계 역행 판정** |
| `claim_expires_at` | **발급자가 계산한 만료** = `claimed_at + lease` | `NULL` = 임차 없음 | **만료 판정** |

> **⛔⛔ 3판은 이 열을 없앴다. 그것이 틀렸다 — 3라운드 A-P5 = B-R12.**
> 진단은 맞았다 — 3판까지 `claim_expires_at`은 **쓰이기만 하고 읽히지 않는 열**이었다.
> **처방이 틀렸다.** 없앨 것이 아니라 **읽는 쪽을 그 열로 바꾸는 것**이 답이었다.
>
> 임차 기간은 `Options.AlertLease`로 **인스턴스마다 주입된다**(C4). 만료를 저장하지
> 않고 매번 유도하면 **판정하는 쪽이 자기 lease로 남의 임차를 재해석한다**:
>
> ```text
> A가 lease=90s로 claim            → claimed_at = T
> B가 lease=60s로 연 핸들에서 판정  → T+60에 "만료"로 보고 훔친다
> 시계는 멀쩡한데 둘이 동시에 publish한다
> ```
>
> **1라운드 B-P4가 깬 것과 같은 모양이다** — 경합자가 들고 온 값으로 남의 임차를
> 판정한다. 3판은 그 값을 **인자에서** 없앴고, `Options`를 통해 **뒷문으로**
> 되돌아왔다. 만료를 **발급자가 계산해서 저장**하면 그 문이 닫힌다.

**두 시각 열이 **둘 다** 읽힌다. 쓰이기만 하는 열은 이제 없다.**

| 판정 | 읽는 열 | 왜 그 열이어야 하는가 |
|---|---|---|
| 만료 | `claim_expires_at` | **발급자의 lease가 권위다.** 판정자의 설정은 안 본다 |
| 시계 역행 | `claimed_at` | 역행에서는 만료 시각**도** 미래다. 발급 **시각** 자체를 봐야 갈린다 |

`RFC3339`(`fills.go:1955-1960`)는 **UTC로 정규화**하므로 문자열 비교가 시각 비교와
같다. 대신 **초 단위로 절삭**한다 — C2의 skew가 그것을 흡수한다.

### C2 — claim 가능 술어

`:now`와 `:skew`는 **원장이 만든다.** **`:lease`는 이 술어에 안 들어간다** —
만료는 이미 행에 저장돼 있다(C1). 호출자는 어느 값도 안 들고 온다.

```sql
state = 'PENDING'
AND ( claim_token = ''                     -- 임차 없음
   OR claim_expires_at IS NULL             -- 같음 (방어)
   OR claim_expires_at <= :now             -- 만료 — 저장값으로만 판정
   OR claimed_at       >  :now_plus_skew ) -- 발급 시각이 미래 → 시계 역행, 열어 준다
```

| 경계 | 부등호 | 왜 |
|---|---|---|
| 만료 | `<=` | 여는 방향이다. 1초 일찍 열려도 재발송 쪽이다 (D4) |
| 역행 | `>` **엄격** | 방금 발급한 임차가 자기 규칙에 걸리면 안 된다 (1라운드 A-P6) |

취득에 성공하면 **같은 UPDATE가 넷을 함께 쓴다** —
`claim_token`(새 토큰) · `claimed_by`(호출자가 준 이름) · `claimed_at`(`:now`) ·
`claim_expires_at`(`:now + lease`). **`lease`는 이 자리에서만 쓰인다.**

`:skew`의 **하한은 2초**다 — RFC3339 절삭 1초 + 여유 1초. 값은 §3.4b가 정한다.

### C3 — 타입 있는 API

문자열 오류로 갈래를 구분하지 않는다. **호출자가 분기해야 하는 것은 전부 타입이다.**

```go
// 취득 결과 — 셋은 서로 다른 뜻이고, 호출자의 행동이 다르다.
type ClaimDisposition int

const (
    ClaimAcquired      ClaimDisposition = iota // 보내라. Token을 들고 있다
    ClaimSettled                               // 보낼 것이 없다 (전달·승인 완료, 창 안)
    ClaimHeldElsewhere                         // 유효한 임차를 다른 발송자가 들고 있다
)

type ClaimResult struct {
    Disposition ClaimDisposition
    ID          int64
    Token       string    // Acquired일 때만 비어 있지 않다
    ClaimedBy   string    // 지금 보유자 — HeldElsewhere면 남, Acquired면 나
    ClaimedAt   time.Time // 지금 임차의 발급 시각
    ExpiresAt   time.Time // 지금 임차의 만료 (C1의 저장값 그대로)

    // 만료 탈취를 관측 가능하게 만드는 셋. Acquired인데 이전 임차가 있었을 때만 찬다.
    Stole    bool      // 남의 만료 임차를 가져왔다
    StoleBy  string    // 이전 보유자
    StoleAt  time.Time // 이전 임차의 발급 시각 — 나이를 로그에 남긴다 (D12)
}

func (j *Journal) ClaimAlertForDelivery(ctx context.Context, a Alert,
    remindAfter time.Duration, claimant string) (ClaimResult, error)
func (j *Journal) ClaimAlertByID(ctx context.Context, id int64,
    claimant string) (ClaimResult, error)   // ← lease를 인자로 받지 않는다
func (j *Journal) EnqueueAlert(ctx context.Context, a Alert) (int64, error) // 임차를 안 잡는다
```

> **`claimant`는 3라운드에 들어왔다.** 3판의 `ClaimAlertForDelivery`는 `claimed_by`에
> 쓸 이름을 받을 자리가 없었는데 C1은 그 열을 요구했다. **`ClaimAlertByID`만 받고
> 다른 하나는 안 받는 것은 같은 열의 두 정본이다.** `EnqueueAlert`는 임차를 안 잡으므로
> 안 받는다 — 그 비대칭은 D13이 뜻하는 바 그대로다.
>
> **`Stole*` 셋도 3라운드에 들어왔다.** R18(만료 탈취를 별도 이벤트로)은 3판 API로는
> **작성 불가**였다 — `Journal`에 logger도 event sink도 없고 `ClaimResult`가 이전
> 보유자를 안 돌려줬다. **원장은 사실을 돌려주고 이벤트는 obs가 낸다**(D12).

정산 쪽도 같다. **0행이 왜 0행인지가 호출자의 행동을 가른다.**

```go
type SettleOutcome int

const (
    SettleApplied        SettleOutcome = iota // 내 임차였고 적용됐다
    SettleLeaseLost                           // 행은 PENDING인데 토큰이 다르다
    SettleAlreadySettled                      // 행이 PENDING이 아니다 (운영자 승인 또는 전달됨)
    SettleNotFound                            // 행이 없다
)

// 정산 결과도 사실을 함께 돌려준다. enum 하나로는 obs가 이벤트를 못 만든다.
type SettleResult struct {
    Outcome   SettleOutcome
    ClaimedBy string    // LeaseLost일 때 **새 보유자** — 내가 잃은 상대다
    ClaimedAt time.Time // 그 임차의 발급 시각 (나이를 로그에 남긴다)
    ExpiresAt time.Time // 그 임차의 만료
}

func (j *Journal) MarkAlertDelivered(ctx, id int64, token string) (SettleResult, error)
func (j *Journal) MarkAlertAttemptFailed(ctx, id int64, token, cause string) (SettleResult, error)
func (j *Journal) ReleaseAlertClaim(ctx, id int64, token string) (SettleResult, error)
```

> **`SettleResult`는 4라운드에 들어왔다.** `SettleOutcome` enum 하나로는 D12가 요구하는
> 「토큰 불일치」 이벤트에 **새 보유자와 임차 나이를 넣을 수 없다** — 4라운드 B가
> R18을 *"RED는 참이지만 작성 불완전"*으로 판정한 이유다. 0행일 때 어차피 같은
> 트랜잭션에서 한 번 더 읽으므로 **그 읽기의 결과를 버리지 않고 돌려주는 것**이 답이다.

**넷의 계약 — 발송자가 무엇을 하는지가 넷을 가르는 유일한 이유다.**

| 값 | 뜻 | 발송자 | 오류인가 |
|---|---|---|---|
| `SettleApplied` | 내 임차였고 적용됐다 | 계속한다 | 아니오 |
| `SettleLeaseLost` | 남이 들고 있다 | **남은 전송을 즉시 멈춘다** + 별도 이벤트 (D12) | 아니오 — 정상 경합이다 |
| `SettleAlreadySettled` | 운영자가 승인했거나 이미 전달됐다 | **남은 전송을 멈춘다** — 보낼 이유가 사라졌다 | **아니오** |
| `SettleNotFound` | 행이 없다 | **멈추고 오류로 올린다** — 원장이 행을 잃었다 | **예** |

> **`SettleAlreadySettled`가 오류가 아닌 이유**: `AcknowledgeAlert`가 발송 중인 행을
> 승인할 수 있고(C5), 그때 발송자의 정산이 이 값을 받는다. 그것은 **정상 경로**다.
> 오류로 다루면 사람이 알림을 승인할 때마다 로그에 오류가 찍힌다.

넷을 가르려면 **0행일 때 같은 트랜잭션 안에서 한 번 더 읽어야 한다.** 그것이 이 API의
비용이고, 그 비용을 내는 이유는 위 표의 **행동이 넷 다 다르기 때문**이다(C5).

### C4 — 임차 기간은 유도한다. 상수를 고르지 않는다

```text
bound  = 한 발송 사이클의 상한 (D4의 유도 — 오늘 기본값으로 54초)
lease  = ceil(bound × margin)          margin은 §3.4가 실측으로 정한다 (하한 1.5)
```

| 누가 | 무엇 |
|---|---|
| `obs` | `bound`를 **살아 있는 설정에서** 계산한다 — 그 설정을 아는 패키지가 여기다 |
| `journal` | `Options.AlertLease`로 받는다. 안 주면 `Open`이 **패키지 기본값**을 채운다 (`journal.go:139-142`의 `BusyTimeout` 선례와 같은 모양) |
| **a099의 테스트** | `DefaultAlertLease > bound(기본 설정)`을 **단언한다** ← 설정이 바뀌면 이 테스트가 깨진다 |
| **a098** | 살아 있는 설정에서 계산한 값을 engine에서 **주입한다** |

**`bound`는 공개 값이어야 한다.** 오늘 유도에 들어가는 셋 중 `defaultBusyTimeout`은
비공개이고(`journal.go:32`) `Ntfy`의 10초는 **상수도 아닌 리터럴**이다(`ntfy.go:95`).
그러면 R20이 *"설정이 바뀌면 깨지는 테스트"*가 아니라 **숫자를 베껴 쓴 테스트**가 된다
— 기억 「증거는 측정해서 쓸 것」이 금지하는 모양이다. 그래서 **`obs`가
`DefaultAlertDeliveryBound()`를 공개하고 R20이 그것을 읽는다** (4라운드 B).

**lease는 `claim`하는 UPDATE 한 자리에서만 쓰인다**(C2). 거기서 계산한 만료가
행에 저장되므로, **인스턴스마다 lease가 달라도 남의 유효한 임차를 재해석하지
못한다.** 그것이 사용자 결정 6-1이 사는 값이고, **R12가 그것을 관측한다** —
6-1 이전에는 그 테스트를 쓸 수 없었다(3라운드 B).

**a099는 `internal/app/engine`을 안 건드린다.** 그래서 a099만으로는 유도가
「기본값에 대한 단언」까지다. **살아 있는 주입은 a098이 진다** — 결정 3으로 둘이
같은 배포에 나가므로 그 사이 상태가 배포되지 않는다.

### C5 — 상태기

`state`와 임차는 **직교한다.** 상태는 전달 여부, 임차는 소유권이다 (D1).

| 원장 상태 | `state` | `claim_token` | 뜻 | `UndeliveredCount` | `PendingAlerts` |
|---|---|---|---|---|---|
| 미전달·무임차 | `PENDING` | `''` | 아무도 안 보내는 중 | **센다** | **보인다** |
| 미전달·임차중 | `PENDING` | 유효 | 누가 보내는 중 | **센다** | **보인다** |
| 미전달·만료임차 | `PENDING` | 만료 | 죽은 발송자의 잔재 | **센다** | **보인다** |
| 전달됨 | `DELIVERED` | `''` | 정산 완료 | 안 센다 | 안 보인다 |
| 승인됨 | `ACKNOWLEDGED` | **남을 수 있다** | 운영자가 봤다 | 안 센다 | 안 보인다 |

> **`ACKNOWLEDGED`에 임차가 남는 경로가 실재한다.** `AcknowledgeAlert`의 술어는
> `WHERE id = ? AND state = 'PENDING'`이고 **임차를 안 본다**
> (`internal-journal--journal.acknowledgealert` 번들). 발송 중인 행을 운영자가 승인하면
> 그 UPDATE가 성공하고, 발송자의 정산은 `SettleAlreadySettled`가 된다. C3이 그 갈래를
> 만든 이유가 이것이다. **재무장은 그 잔재를 반드시 지운다**(전이 표의 마지막 행).
>
> **승인 술어에 임차를 넣지 않는다.** 운영자의 승인은 발송자의 임차보다 우선한다 —
> 사람이 봤다는 사실이 기계의 진행 중 상태에 막히면 안 된다.

전이는 이것뿐이다.

| 전이 | 함수 | 술어 | 임차에 하는 일 |
|---|---|---|---|
| 취득 | `ClaimAlertForDelivery` · `ClaimAlertByID` | C2 | **네 열을 함께 쓴다** (C1) |
| 성공 정산 | `MarkAlertDelivered` | `id + state=PENDING + token 일치` | **푼다** — 네 열 초기화 |
| 시도 실패 기록 | `MarkAlertAttemptFailed` | 같음 | **유지한다** ← 재시도가 남아 있다 (D3) |
| 포기 | `ReleaseAlertClaim` | 같음 | **푼다** — 네 열 초기화 |
| 운영자 승인 | `AcknowledgeAlert` | `id + state=PENDING` (**임차 무시**) | **안 건드린다** |
| 재무장 | `ClaimAlertForDelivery`의 rearm UPDATE | `claimOwed`가 참 | **반드시 지운다** — 새 episode다 |

> **「네 열 초기화」의 정본은 하나다** — `claim_token=''` · `claimed_by=''` ·
> `claimed_at=NULL` · `claim_expires_at=NULL`. 이 문장을 다른 절·tasks·map에서
> 다시 적지 않는다. 이 change는 **같은 값의 사본으로 세 라운드 연속 사고를 냈다.**

### C6 — 진입 게이트: **트리거는 오늘 그대로, 경합에는 안 잠근다** (사용자 결정 5-1)

**오늘의 게이트 자리는 다섯이고, a099는 다섯 다 안 건드린다.**

| # | 자리 | 오늘 | a099 + a098 |
|---|---|---|---|
| 1 | `claimAndDeliver:261-262` — 행을 **기록조차 못 했다** | `Block` | **그대로** |
| 2 | `deliver:378-379` — publish 성공 + **정산 실패** | `Block` | **그대로** |
| 3 | `deliver:403-404` — **재시도 예산 소진** | `Block` | **그대로** |
| 4 | `Acknowledge:481-482` — `Journal`이 nil | `Clear` | **그대로** |
| 5 | `Acknowledge:510-511` — 승인 뒤 남은 수가 0 | `Clear` | **그대로** |
| — | claim 경합으로 발송을 건너뜀 | 없음 | **아무것도 안 한다** — 래치도, 유도도 |
| — | 기동 직후 | 메모리 래치라 **비어 있다** | **원장을 읽어 `Block`만 복원한다** ← a098 |
| **6** | **배달 실행자가 죽었다** | **없다** | **`Block` — 오늘 없던 자리다** ← **a098** (사용자 결정 8-1) |

> **⚠ 여섯째 줄은 a099가 아니라 a098이 진다 — 그래도 여기 적는다.**
> 위 다섯은 **a099가 한 줄도 안 건드리고**, 결정 5-1이 요구한 것이 그것이다.
> 그러나 이 표의 마지막 열은 **`a099 + a098` 배포 단위**를 말하고,
> 그 단위는 **오늘 없던 차단을 하나 만든다**(사용자 결정 8-1, 2026-08-11).
>
> **여섯째를 이 표에서 빼면 이 표가 거짓말을 한다** — 배포 단위가 다섯만
> 그대로 둔다는 뜻으로 읽힌다. 정본은 a098 design **D8.2**이고,
> 그 자리는 **자기 ReasonCode를 쓴다**: 같은 코드를 쓰면 `Acknowledge`가
> 미전달 0에서 5번 줄을 풀 때 **여섯째까지 함께 풀린다**(`notifier.go:507-512`).
>
> 조건의 종류가 다르다 — 위 다섯은 *"보내려다 실패했다"*이고 여섯째는
> *"보낼 주체가 없다"*다.

다섯을 그대로 두기로 한 경위는 아래에 있다.

> **⛔⛔ 3판의 C6은 승인된 규범을 뒤집었다 — 3라운드 A-P1. 내가 만든 결함이다.**
>
> 사용자 결정 2는 *"경합으로 래치하지 말고 원장에서 유도하라"*였는데, 3판은
> **차단과 해제를 둘 다** 유도로 바꿨다. 해제 쪽이 문제다.
>
> | | 오늘 | 3판 C6 |
> |---|---|---|
> | 차단 | 시도했다 실패한 행만 | 미전달 행이 있으면 — **더 보수적** |
> | 해제 | **운영자 승인 + 0건** | 전달만 되면 자동 — **덜 보수적** |
>
> 해제 쪽은 승인된 정본 `openspec/specs/engine-safety/spec.md:147-152`
> (*"전달 복구 후 **수동 확인**"*)을 `MODIFIED` 없이 어긴다. 불변식 6은 명확한 근거가
> 있는 **보수 방향만** 허용한다. 게다가 성공 정산 뒤 `Clear`를 부르는 **자리가 없어서**
> 2판 D10이 만든 영구 래치가 형태만 바꿔 되살아났다 (3라운드 A-P2).
>
> **사용자 결정 5-1(최소)로 되돌린다.** 규범 변경 0, `MODIFIED` 불필요.

바뀌는 것은 **둘뿐이다.**

1. **경합자는 게이트를 안 건드린다.** `ClaimHeldElsewhere`를 받으면 발송을 건너뛰고
   구조화 로그 한 줄만 남긴다(D12). `Gate.Block`도 `Gate.Clear`도 부르지 않는다.
   `EntryGate.Block`은 reason별 **메모리 래치이고 자동 해제가 없다**
   (`execgw/retry.go:498-505`) — 정상 발송 중의 경합에 그것을 걸면 성공한 발송이
   **아무도 못 푸는 잠금**을 남긴다 (2라운드 A-P1).
2. **기동 복원.** 오늘 재시작은 게이트를 지운다 — 래치가 메모리에 있기 때문이다.
   미전달 critical 행이 원장에 남아 있는데 진입이 열리는 것은 **재시작이 차단을 푸는
   우회로**다. a098이 기동 시 미전달 수가 0보다 크면 `Block`을 복원한다.
   **`Clear`는 복원하지 않는다** — 방향이 보수적인 쪽만 한다.

> **비대칭을 정직하게 적는다.** 런타임 차단의 근거는 *시도했다 실패한 행*이고
> 기동 복원의 근거는 *미전달 수*다. 그래서 **한 번도 시도되지 않은 미전달 행**은
> 런타임에는 진입을 안 막고 재시작 뒤에는 막는다.
>
> 두 근거를 합치는 것이 결정 5의 (b)·(c)였고 **사용자가 안 골랐다.** 합치면 critical
> 알림 하나마다 사람 승인이 있어야 진입이 열린다 — 제품 동작이 크게 바뀐다.
> **그래서 「아직 시도 안 한 행이 진입을 막는다」는 이 change가 하지 않는다**
> (3판의 R5는 등록부에서 빠졌다).

### C7 — 무엇을 보장하고 무엇을 보장하지 않는가 (사용자 결정 1)

| 조건 | 보장 |
|---|---|
| 두 발송자가 **유계 실행** 아래에서 경합한다 | **정확히 하나만** 보낸다 |
| 발송자가 **정지·긴 GC·VM 일시정지**로 임차를 넘겨 잃는다 | 이미 나간 publish는 못 되돌린다. **정산은 반드시 거부된다** (`SettleLeaseLost`) |
| 그 결과 | 알림이 **두 번 보일 수 있다.** 한 번도 안 보이는 것보다 낫다 (불변식 4·6) |

> **⛔ 「at-least-once」라고 부르지 않는다 — 4라운드 A-P8.**
> 그 이름은 **적어도 한 번은 도달한다**는 뜻인데 이 change는 그것을 못 지킨다.
> 전송 수단이 계속 죽어 있거나(ntfy 불통) 발송 주체가 없으면(오늘의 `Flush`)
> 도달 횟수는 **0**이다. 이 change가 정하는 것은 **누가 보내고 누가 정산하는가**이고,
> **몇 번 도달하는가**는 a098의 배달 주체와 진입 차단이 진다.

> **첫 행은 조건부다 — 그 조건을 적는다.** *"유계 실행"*이 뜻하는 유계는
> **`lease > bound`**다(C4). 임차가 한 발송 사이클보다 짧으면 발송자는 아무도 안
> 멈췄는데도 자기 예산을 쓰는 도중에 임차를 잃고, 그때 둘이 보낸다.
> **그래서 C4의 단언(R20)이 이 보장의 일부다.** 기본 설정에 대해서는 a099가 단언하고,
> 살아 있는 설정에 대해서는 **a098의 주입이 진다** — 둘이 한 배포 단위인 이유가
> 여기에도 있다(D9).

**만료 임차는 원격 publish를 fence하지 못한다.** A가 SIGSTOP에 걸리고, 임차가 만료되고,
B가 탈취해 보내고, A가 깨어나 자기 publish를 실행한다 — 전화기에 두 번 도착한 **뒤에야**
A의 정산이 거부된다. **a099가 기존 CAS를 반증한 논리(publish 뒤에 도는 술어는 이중
발송을 못 막는다)가 자기 토큰 CAS에도 그대로 걸린다** (2라운드 B-P2).

그래서 규범을 **참이 되는 자리까지 낮춘다**: 원장은 이중 **정산**을 절대 허용하지 않고,
이중 **발송**은 유계 실행 밖에서 허용한다. 1·2판은 둘을 한 문장에 넣어 **거짓인 규범**을
만들었다. 정확히 한 번을 원하면 전송 수단에 idempotency 키가 필요한데 ntfy에는 없다
(D8).

## D1 — 상태를 늘리지 않는다. 임차 열을 더한다

두 안을 놓고 골랐다.

**안 A — `SENDING` 상태 추가.** 발송 중인 행을 상태로 표시한다.
**안 B — 임차 열 추가**(C1의 넷). 상태는 셋 그대로 두고 행은 발송 중에도 PENDING이다.

### 안 A를 버린 이유는 진입 게이트다

`UndeliveredCount`의 AST(`internal-journal--journal.undeliveredcount`,
분기 1 / 이탈 2, `outbox.go:408-415`)는 이 함수가 질의 하나와 오류 검사 하나뿐임을
열거한다. 그 질의의 술어는 `WHERE state = ?`(PENDING)이고, `outbox.go:407`의 주석이
그 값의 용도를 적는다 — *"the number the entry gate reacts to"*.

안 A에서 발송 중인 행은 PENDING이 아니다. 그러면:

- `UndeliveredCount`에서 **빠진다**
- `PendingAlerts`에서 **빠진다** → 운영자의 밀린 목록에서 사라진다
- 발송자가 죽으면 그 행은 **어느 읽기 경로에도 안 나타난 채** 영원히 SENDING이다

> **⚠ 1판은 여기서 「진입 차단이 스스로 풀린다」고 썼다. 기전을 안 적었다 — 1라운드 A-P9.**
>
> `UndeliveredCount`가 스스로 게이트를 움직이지 않는다. 실제 배선은 하나다:
> **`Acknowledge`가 그 값을 읽고 0이면 `Gate.Clear`를 부른다**(`notifier.go:506`).
> 그러니 정확한 문장은 *"발송 중인 행이 카운트에서 빠지면, 운영자가 **다른** 알림을
> 승인하는 순간 게이트가 풀린다"*이다. `EntryGate.Block` 자체는 메모리 래치일 뿐이다
> (`execgw/retry.go:495-510`).
>
> **그리고 안 A가 반드시 그렇게 되는 것도 아니다** — 네 읽기 질의를
> `state IN ('PENDING','SENDING')`으로 고치면 안 A도 이 성질을 지킬 수 있다.
> 1판은 **일부러 망가뜨린 4상태 대안**과 비교했다. 정직한 비교는 아래다.

| | 안 A (질의를 제대로 고친 4상태) | **안 B (임차 열)** |
|---|---|---|
| 읽기 질의 넷 | **전부 고쳐야 한다** | **한 줄도 안 고친다** |
| 고치는 것을 빠뜨리면 | **미전달인데 게이트가 풀린다** | 그런 실패 모드가 없다 |
| 임차 논리를 통째로 무시하면 | 행이 SENDING에 갇힌다 | **오늘과 같다** |
| 상태의 의미 | *"보내는 중"*은 전달 여부가 아니다 — 상태 열에 성질이 둘 섞인다 | 전달 여부는 상태, 소유권은 임차 |

**안 B를 고르는 이유는 「안 A가 반드시 깨진다」가 아니라 「안 B는 깨질 자리가 없다」다.**
안전 불변식 4·5에 닿는 결정이므로 실수 여지가 적은 쪽을 고른다.

### 안 B가 지키는 것 — 읽기 경로 다섯이 그대로다

| 함수 | 술어 | AST 근거 | a099 |
|---|---|---|---|
| `UndeliveredCount` `:411` | `WHERE state = ?` | 분기 1 / 이탈 2 | **그대로** |
| `PendingAlerts` `:393` | `WHERE state = ? ORDER BY id` | 분기 2 / 이탈 2 (`internal-journal--journal.pendingalerts`) | **그대로** |
| `AcknowledgeAlert` `:378` | `WHERE id = ? AND state = ?` | (편집 없음) | **그대로** |
| `MarkAlertDelivered` `:342` | `WHERE id = ? AND state = ?` | 분기 1 / 이탈 2 | **+ 토큰 술어**, 성공 시 **임차를 푼다** (C5) |
| `MarkAlertAttemptFailed` `:357` | `WHERE id = ? AND state = ?` | 분기 1 / 이탈 2 | **+ 토큰 술어**, 임차는 **유지한다** (C5 · D3의 ⚠⚠) |

> **⚠ 이 표의 마지막 행이 3라운드까지 *"임차 해제"*라고 적혀 있었다.** 정본(C5)은
> 그 반대를 말하고, D3의 ⚠⚠ 절 전체가 그 반대인 이유를 진다. **정본 절을 만들어도
> 같은 문서의 요약표가 안 따라오면 사본은 살아남는다** — 3라운드가 그것을 잡았다.

### 그리고 안 B는 실패 방향이 옳다

임차 논리를 **통째로 무시해도** 행은 PENDING이고 모든 술어가 오늘과 같다.
즉 안 B의 최악은 **오늘**이다. 안 A의 최악은 오늘보다 나쁘다.

## D2 — claim은 이미 열린 트랜잭션 안에서, 새 연결 없이

`ClaimAlertForDelivery`는 이미 `:182`에서 트랜잭션을 열고 `:186`에서 rollback을 defer하며
`:191`에서 SELECT를 한다. 임차 CAS는 **그 트랜잭션 안의 UPDATE 한 문장**이다.

새 연결도, 새 뮤텍스도 없다. **그러나 「비용이 없다」는 뜻이 아니다** —
같은 트랜잭션 안에 **UPDATE 한 문장**이 늘고 그 커밋은 **WAL fsync**를 한다.
아래 표가 그 값을 적는다. 4라운드 A-T1이 *"새 왕복 없음"*이라는 옛 표현이
바로 아래 비용표와 모순된다고 지적했다.

**정직하게 적는다: 이것은 동작 변화 0이 아니다.** PENDING 경로는 오늘 UPDATE를 하나도
안 하고, a099 이후에는 하나 한다.

> **⚠⚠ 1판이 그 비용을 틀리게 적었다 — 1라운드 A-P7.**
> *"트랜잭션이 읽기에서 쓰기로 승격된다"*고 썼는데 **틀렸다.**
> `_txlock=immediate`(`journal.go:202`)이므로 `BeginTx`가 **이미 쓰기 잠금을 잡는다.**
>
> 실제로 바뀌는 것은 그것보다 **크다**: 지금까지 깨끗하던 트랜잭션이 **WAL을 더럽히고**,
> `synchronous(full)`(`:199`)이라 **커밋마다 fsync**하며, 그 연결은
> `SetMaxOpenConns(1)`(`:151`)로 **하나뿐**이다.

| | 오늘 | a099 |
|---|---|---|
| PENDING 경로의 문장 수 | SELECT 1 | SELECT 1 + UPDATE 1 |
| 그 커밋이 하는 일 | WAL 무변화 | **WAL 쓰기 + fsync** |
| 연결 | 하나 (`:151`) | 같음 — **다른 쓰기와 직렬화된다** |
| 그 구간을 덮는 잠금 | `n.mu` (`notifier.go:242` defer) | **같음 — a099는 안 줄인다** |
| 같은 잠금 안에서 이미 일어나는 것 | **네트워크 publish** (`:276` → `deliver`) | 같음 |

### 합격 기준을 **재기 전에** 정한다 (1라운드 A-P8)

1판은 *"네트워크 publish보다 지배적일 수 없다"*를 근거로 삼았다. **그것은 안전 기준이
아니다** — 비교 대상이 이미 큰 값이라는 말일 뿐이고, 얼마까지 허용하는지를 안 정한다.

**§3.5가 재기 전에 다음을 선언한다.**

| 무엇 | 값 |
|---|---|
| 대상 | `claimAndDeliver` 진입부터 `deliver` 첫 `Publish` 호출까지의 체류 |
| 기준선 | a099 **이전** HEAD에서 같은 지점을 잰 값 |
| 허용 회귀 | **§3.5가 숫자로 선언하고 그 뒤에 잰다** — 넘으면 이 설계는 기각이고, 임차 취득을 정지 임계 경로 **밖으로** 옮긴다 |
| 측정 조건 | 다른 쓰기가 같은 연결을 쓰는 상태를 포함한다 (`SetMaxOpenConns(1)`) |

**측정 결과를 기록만 하고 넘어가지 않는다.** 기준을 넘으면 설계가 바뀐다.

## D3 — 배제의 근거는 **매 취득마다 새로 만드는 토큰**이다

> **⚠⚠ 1판은 `claimed_by`(프로세스·루프 식별자)를 CAS 술어로 썼다.**
> **1라운드 A-P4 = B-P5가 그것을 ABA로 깼다.** 같은 루프가 같은 문자열로 재취득하면
> 옛 발송자의 CAS가 **새 보유자의 임차에 그대로 맞는다.**
>
> ```text
> A가 "delivery-loop"로 claim → 만료 → A가 아직 살아 있음
> B(같은 루프)가 "delivery-loop"로 재claim
> A가 늦게 도착해 WHERE claimed_by='delivery-loop' → B의 임차를 친다
> ```
>
> **안정된 이름은 fencing token이 아니다.**

그래서 이름 하나를 **열 넷으로 나눈다** — **열과 술어와 시그니처는 C1·C2·C3이 정본이다.**
여기서 다시 적지 않는다. 이 절이 지는 것은 **왜 이름이 아니라 토큰인가**뿐이다.

`claim_token`은 `ClaimAlertForDelivery`·`ClaimAlertByID`가 **돌려주고**, 이후 모든
변경(`MarkAlertDelivered`·`MarkAlertAttemptFailed`·`ReleaseAlertClaim`)이 그것을 받는다.

취득할 때는 토큰을 **비교하지 않는다.** 만료된 임차는 누구나 가져갈 수 있다.
그 외 모든 변경은 **비교한다.**

### 임차를 잃은 발송자는 **남은 전송을 즉시 멈춘다**

토큰 불일치는 오류 문자열이 아니라 **타입 있는 결과**다(C3의 `SettleLeaseLost`).
그래야 `deliver`가 그것을 받고 루프를 끊는다.

**1판은 이것을 안 적었고**, 그 상태로는 `deliver:384`가 오류를 로그로만 다루고
`:387-391`이 계속 돈다 — **임차를 잃은 발송자가 계속 publish한다**(1라운드 A-P4).
그것은 a099의 요구 자체를 무효로 만든다.

### ⚠⚠ 「실패 기록」과 「해제」는 같은 것이 아니다

a099의 첫 초안은 `MarkAlertAttemptFailed`가 임차를 풀게 했다. **틀렸다.**

`deliver`의 재시도 루프를 읽으면 그 이유가 나온다(`notifier.go:341-392`).

| 줄 | 무슨 일 |
|---|---|
| `:354` | publish 시도 |
| `:384` | 실패하면 **`MarkAlertAttemptFailed`** |
| `:387-391` | **예산이 남았으면 대기하고 다시 `:354`로 간다** |

**실패 기록은 발송자가 끝났다는 뜻이 아니다.** 같은 호출 안에서 또 보낸다.
여기서 임차를 풀면 `:388`의 대기 동안 두 번째 발송자가 그 행을 집고,
원래 발송자는 임차 없이 `:354`를 다시 실행한다 — **a099가 막으려던 바로 그 이중 발송을
a099가 만든다.**

그래서 셋으로 가른다.

```text
MarkAlertDelivered(ctx, id, token)             // 성공 — 임차를 푼다 (state가 떠난다)
MarkAlertAttemptFailed(ctx, id, token, cause)  // 시도 기록 — 임차를 유지한다
ReleaseAlertClaim(ctx, id, token)              // 포기 — 임차만 푼다  ← 신설
```

셋 다 토큰을 CAS 술어에 넣는다. 임차를 잃은 발송자는 셋 중 무엇도 못 하고,
**그 사실을 타입 있는 결과로 받아 남은 전송을 멈춘다.**

**`ReleaseAlertClaim`의 호출 자리 둘** — 내가 셋으로 늘렸고 1라운드 B-P3이 되돌렸다:

| 자리 | 푸는가 | 왜 |
|---|---|---|
| `deliver`의 루프 **뒤** (`notifier.go:394` 이후) | **푼다** | 예산을 다 썼다. 이 발송자는 끝났고 행은 PENDING으로 남는다 |
| `Flush`의 루프 안, 행마다 (D7) | **푼다** | 그 행에 대한 이 발송자의 일이 끝났다 |
| `deliver` 이탈 `:381` (publish 성공 + **정산 실패**) | **⛔ 안 푼다** | 푸시는 나갔는데 원장이 그것을 모른다. 여기서 풀면 다음 관측이 **다시 보낸다** — 2026-08-08의 폭풍이 성공 경로로 돌아온다(a096 round 2가 같은 자리를 이미 겪었다). **임차가 억제 표시로 남고 만료가 푼다** |

기존 호출 넷은 전부 `internal/obs/notifier.go`다 — `:356`, `:384`, `:452`, `:455`.

## D4 — 만료는 침묵이 아니라 재발송 쪽으로 연다

**claim 가능 술어는 C2가 정본이다.** 이 절은 그 술어가 **왜 그 모양인가**를 진다.

만료된 임차는 **행을 다시 보낼 수 있게** 만든다. 죽은 발송자가 알림을 무덤까지
가져가지 않는다.

이 방향이 보수적이다(불변식 6). 알림 시스템에서 **at-least-once가 at-most-once보다
안전하다** — 운영자가 두 번 보는 것과 한 번도 못 보는 것 중 후자가 사고다.

### 시계가 뒤로 가면 — 이 저장소는 이미 그 경우를 겪었다

`claimOwed`의 B6 `outbox.go:290-302`가 그 선례다. NTP 되돌림·복원된 스냅숏·부팅 시
앞서 있던 RTC를 열거하고, **억제하는 쪽이 더 나쁜 오류**라고 판정해서 열린 쪽으로
실패한다.

임차에 같은 판정을 적용한다.

| 시계가 | `claim_expires_at <= now`가 | 결과 | 판정 |
|---|---|---|---|
| 앞으로 튄다 | 일찍 참이 된다 | 임차가 일찍 풀린다 → **이중 발송 가능** | 받아들인다 — at-least-once (D4) |
| **뒤로 간다** | **오래 거짓** | **임차가 스큐만큼 길어진다** → 발송자가 죽었으면 **그동안 아무도 안 보낸다** | **받아들이지 않는다** |

> **역행 판정에 만료 열을 쓸 수 없는 이유가 이 표에 있다.** 시계가 뒤로 갔을 때
> `claim_expires_at`은 그냥 **미래**이고, 그것은 살아 있는 정상 임차와 구별되지 않는다.
> 갈리는 것은 **발급 시각이 미래인가**이고, 그래서 역행 판정만 `claimed_at`을 읽는다
> (C1의 두 번째 표).

뒤로 간 시계는 억제 방향이므로 막아야 한다. 1판은 규칙을 이렇게 썼다.

```text
저장된 만료 시각이 now 보다 임차 기간 이상 미래면, 그 임차는 만료된 것으로 본다.
```

> **⚠⚠ 두 보이스가 이 규칙을 각각 다른 방식으로 깼다.**
>
> **A-P6 (등호 경계)**: 정상 임차의 만료는 `now + lease`다. 그 값은
> `now + lease` **이상**이므로 방금 발급한 임차가 자기 규칙에 걸린다. 게다가
> 시각은 RFC3339 문자열이라 **초 단위로 잘린다** — 경계가 더 흐려진다.
>
> **B-P4 (임차 길이가 다르면 유효한 임차를 훔친다)** — 이쪽이 더 나쁘다:
> ```text
> A가 60초 임차로 claim  → expires = T+60
> B가 34초 임차로 접근    → T+60 > B.now + 34  → "시계 역행"으로 판정하고 훔친다
> 시계는 멀쩡한데 둘이 동시에 publish한다
> ```
> 규칙이 **경합자가 들고 온 값**으로 남의 임차를 판정한다. 그 값은 권위가 없다.

> **⚠⚠ 3판은 B-P4를 고쳤다고 믿었고, 뒷문이 열려 있었다 — 3라운드 A-P5 = B-R12.**
> 3판은 `lease`를 **인자에서** 없앴지만 만료를 저장하지 않고 매번 유도하게 만들었고,
> `lease`는 `Options.AlertLease`로 **인스턴스마다** 들어온다. 그러면 위 시나리오가
> `60초 인자` → `60초 설정으로 연 핸들`로 **이름만 바꿔** 그대로 성립한다.
> **사용자 결정 6-1이 만료를 다시 저장하게 해서 그 문을 닫았다**(C1).

### 고친 규칙

1. **임차 기간의 권위는 발급 시점에 있다.** 호출자가 값을 들고 오지 않고(C3),
   판정자의 설정도 남의 임차에 못 닿는다 — 만료가 **행에 저장돼 있기 때문**이다(C1).
2. **만료 시각을 저장한다.** 발급자가 `claimed_at + lease`를 계산해
   `claim_expires_at`에 쓰고, **만료 판정은 그 저장값만 읽는다.**
   역행 판정만 `claimed_at`을 읽는다 — 이유는 위 표 아래에 적었다.
3. 역행 경계는 **엄격 부등호**, 만료 경계는 **등호 포함**이고, RFC3339 초 절삭을
   흡수할 **명시된 skew 여유**를 둔다(C2). 그 값은 §3.4b가 정한다.
4. **앞으로 튄 시계**는 임차를 일찍 만료시킨다 — 그것은 재발송 방향이므로 받아들이되,
   **탈취가 일어났다는 사실은 관측 가능해야 한다**(1라운드 B-P6). 조용히 넘어가지 않는다.
   `ClaimResult`의 `Stole*` 셋이 그 사실을 obs에 넘긴다(C3).

§3이 이것을 관측한다 (R9~R12). **규범 문서에도 넣는다** — 1판은 design에만 있었고
spec에 없었다(B-P4).

### 만료 값이 넘어야 하는 것은 **한 시도가 아니라 한 사이클 전체**다

첫 초안은 *"한 번의 발송 시도가 끝나기 전에 만료되면 안 된다"*라고 썼다. **부족하다.**
발송자는 한 번 claim하고 **재시도 예산 전체**를 그 임차 아래에서 쓴다(D3의 표).

오늘 기본값으로 그 사이클의 상한을 계산하면:

| 항목 | 값 | 근거 |
|---|---|---|
| 시도 횟수 | 3 | `DefaultCriticalAttempts` `notifier.go:45` |
| 한 시도의 상한 | 10초 | `Ntfy.Timeout` 기본값 `ntfy.go:72` |
| 시도 사이 대기 | 2초 | `DefaultRetryDelay` `notifier.go:48` |
| **실패 기록의 쓰기 대기** | **5초 × 3** | `defaultBusyTimeout` `journal.go:32` — **1판이 빠뜨렸다** |
| **해제의 쓰기 대기** | **5초** | 같음 |
| **한 사이클 상한** | **3×(10+5) + 2×2 + 5 = 54초** | 1라운드 B-P2 |

**임차는 54초보다 길어야 한다.** 34초를 쓰면 발송자가 **자기 문서화된 예산을 쓰는
도중에** 임차를 잃고, 그때 다른 발송자가 같은 행을 집는다.

**그리고 이 값은 네 설정에 매여 있다** — `Notifier.Attempts`·`Notifier.RetryDelay`·
`Ntfy.Timeout`·`Journal.Options.BusyTimeout`, 그리고 a092가 도입할 발송 상수들.
**상수를 하나 고르는 것이 아니라 유도식을 코드에 둔다** — 누가 계산하고 누가 주입하고
무엇이 그것을 지키는지는 **C4가 정본이다.** 그래야 설정이 바뀔 때 임차가 조용히
짧아지지 않는다.

**이 값은 a092에 매여 있다.** a092가 `alertPublishAttempts`·`alertPublishTimeout`을
도입하면 위 셋이 바뀌고 이 상한도 바뀐다. **a099는 그 상수를 인용하지 않지만,
a092가 착지할 때 이 유도를 다시 하는 것을 §7이 넘긴다.**

§3.4가 실측으로 이 유도를 확인한다. **a092의 `alertDeliveryBound`도 a098의 주기도
인용하지 않는다** — 위 세 값은 이 change가 직접 읽은 것이다.

## D5 — Pre-Edit 선언 (High-risk)

```text
Pre-Edit Gate:
- change id / task id: a099 / §4의 GREEN task 전부
- 대상 심볼 (**편집한다**) — 1라운드 A-P10이 둘, 3라운드가 둘을 더 채웠다:
    journal.Journal.ClaimAlertForDelivery   — 임차 CAS + 재무장 UPDATE에 **열 넷** (C1)
                                              반환이 ClaimResult로, 인자에 claimant (C3)
    journal.Journal.MarkAlertDelivered      — 토큰 술어 + SettleOutcome, 성공 시 해제 (C3·C5)
    journal.Journal.MarkAlertAttemptFailed  — **임차 유지.** 토큰 술어만 더한다 (D3)
    journal.Journal.EnqueueAlert            — **← 누락됐었다.** 기록 전용이므로
                                              **임차를 안 잡는 경로**로 위임한다 (A-P5·D13)
    journal.Open                            — **← 3라운드에 추가됐다.** `Options.AlertLease`가
                                              비면 패키지 기본값을 채운다 (C4).
                                              `journal.go:139-142`의 BusyTimeout 자리 바로 옆이다
    journal.scanAlerts                      — **← 3라운드에 추가됐다.** 임차 상태를 투영한다 (D12).
                                              `alertSelect`(`outbox.go:386`, 상수)와 `Alert`(구조체)도
                                              같이 바뀌지만 **함수 내부 로직은 이쪽뿐**이다
    obs.Notifier.deliver                    — **← 누락됐었다.** 포기 시 해제 + token 전파
    obs.Notifier.Flush                      — 행마다 claim, 못 잡으면 건너뛴다
    obs.Notifier.claimAndDeliver            — **doc + claim 경합 결과 분리** (B-P1).
                                              1판은 「doc만」이라고 적었고 그것은
                                              **crash 경로를 오늘보다 나쁘게 만든다**
- 대상 심볼 (**편집하지 않는다 — 그러나 분기를 근거로 쓴다**):
    journal.claimOwed                — B2 :276이 반증의 근거
    journal.Journal.UndeliveredCount — 술어 불변이 D1의 안전 논거
    journal.Journal.PendingAlerts    — 같음
    journal.Journal.AcknowledgeAlert — 술어가 임차를 안 본다는 것이 C5·D10의 근거
    obs.Notifier.Acknowledge         — **← 2라운드에 「편집」으로 넣었다가 결정 5-1로
                                       「인용만」이 됐다.** C6이 `:510-511`의 분기를
                                       *"오늘 이렇게 동작하고 a099가 안 건드린다"*의
                                       근거로 쓰므로 **산출물은 그대로 필요하다**
- 기존 동작 파악 근거: analysis/function-logic/ **13개 번들**(AST + Map + Branch Test Map),
  internal/journal/outbox_test.go,
  internal/obs/{obs_test.go, a096_one_send_per_condition_test.go, a096b_round2_test.go,
                a097_claim_failure_blocks_entry_test.go, a097_exclusion_is_an_event_test.go,
                escalation_test.go}
  ※ 1판은 여기에 `internal/obs/notifier_test.go`를 적었다. **그 파일은 없다** —
    1라운드 B-T1이 잡았다. 있을 법한 이름을 근거로 쓴 것이다.
- upstream 상속 테스트 영향: no — alert outbox는 TossOS 고유(schemaV3, task 4.3)
- 실패 테스트 선행 작성: yes (§3 전부가 §4보다 먼저다)
- 안전 불변식 §0 위반 여부 검토: 미판정 — §3의 진입 게이트 관측과 지연 측정 뒤에 판정한다
```

> **편집하지 않는 셋에도 산출물이 있는 이유**: `.claude/CLAUDE.md`의 규칙은
> *"함수 내부의 분기를 **근거로 삼는** 문서"*에 걸린다. 편집 여부가 아니다.
> D0의 반증도 D1의 안전 논거도 이 셋의 내부를 근거로 쓴다.

**High-risk 경로다**: 원장 스키마(불변식 5). 적대적 Eng 리뷰가 필수다.

## D6 — 스키마

**DDL은 C1이 정본이다.** 이 절이 지는 것은 `schema.go:8-22`의 additive 규칙 대조뿐이다.

| 규칙 | 대조 |
|---|---|
| 1. 출시된 step을 편집하지 않는다 | `schemaV3`(`outbox.go:41-68`)을 안 건드린다 |
| 2. 새 열은 nullable이거나 DEFAULT를 가진다 | `claim_token`·`claimed_by`는 DEFAULT `''`, `claimed_at`·`claim_expires_at`은 nullable |
| 3. 열을 지우거나 이름을 바꾸지 않는다 | 없음 |
| 4. 과거 행을 migration에서 다시 쓰지 않는다 | 없음 — 기존 행은 `claim_token=''`·`claim_expires_at=NULL`, 즉 **claim 가능**이고 그것이 오늘의 의미다 |

`ALTER TABLE ... ADD COLUMN ... TEXT NOT NULL DEFAULT ''`는 이 저장소의 선례가 있다
(`fills.go:2012-2014`, schemaV16). STRICT 테이블에서 동작한다.

인덱스는 안 더한다. `idx_outbox_state(state, id)`가 PENDING 집합을 이미 고르고,
임차 술어는 그 집합 안에서만 판정한다. 그 집합은 **미전달 critical 알림**이므로
구조적으로 작다.

## D7 — claim을 안 거치는 발송 경로를 남기지 않는다

`Flush`는 오늘 claim을 아예 안 부른다 — `PendingAlerts`를 읽고 곧장 publish한다
(a098이 잰 `internal-obs--notifier.flush`, 분기 6 / 이탈 4, `notifier.go:427-462`).

**임차가 있는데 우회 경로가 하나 남아 있으면 그것은 임차가 아니다.**

그래서 journal에 id로 거는 claim을 하나 더한다 — 시그니처는 **C3이 정본**이고
`lease`를 인자로 받지 않는다.

`Flush`는 행마다 이것을 부르고, `ClaimHeldElsewhere`를 받은 행은 **건너뛴다**
(오류가 아니다 — 다른 발송자가 보내는 중이라는 뜻이다).

오늘 `Flush`는 `n.mu`를 루프 전체에 걸쳐 잡는다(`:434-435`). a099는 **그 잠금을 안
건드린다.** 그 안에서 claim은 항상 성공한다. 구간 재설계는 a092다.

## D8 — a099가 하지 **않는** 것 — 정직한 한계

| 한계 | 왜 |
|---|---|
| **정확히 한 번은 보장하지 않는다** | **C7이 그 경계를 규범 수준으로 적는다.** 만료 임차는 원격 publish를 fence하지 못한다 — 이중 **정산**은 절대 안 되고 이중 **발송**은 유계 실행 밖에서 가능하다. 정확히 한 번을 원하면 전송 수단에 idempotency 키가 필요한데 **ntfy에는 없다** |
| **지연 특성을 개선하지 않는다** | `n.mu`가 그대로다. 그것이 a092 |
| **발송 주체를 만들지 않는다** | `Flush`의 프로덕션 호출자는 여전히 0이다. 그것이 a098이고, **그래서 둘이 한 배포 단위다**(D9) |
| **운영자 표면을 만들지 않는다** | `Acknowledge`의 프로덕션 호출자도 0이다. `tossctl` 하위 명령은 a098이다 (사용자 결정 4) |
| **`deliver` 안의 재시도 예산·타임아웃** | 안 건드린다 — a092의 상수 표면이다 |
| **`SetModeProjector` 미배선** | a092가 발견한 미배정 후속 그대로 |

## D9 — 왜 이 순서인가 (3 배포 단위)

> **⚠⚠ 2라운드 A-P10 = B-P6이 이 표의 1단계를 깼다. 사용자 결정 3으로 고쳤다.**
> 1·2판은 a099를 **혼자 배포할 수 있다**고 적었다. **거짓이다** —
> `Flush`의 프로덕션 호출자가 0이므로 claim한 뒤 죽은 행을 **다시 집을 주체가 없다.**
> 오늘은 다음 관측이 그냥 보내는데, a099 단독은 살아 있는 임차 동안 그것을 억제한다.
> **a099만 올리면 오늘보다 나쁘다.**

| # | 배포 단위 | 이 단위가 끝나면 참인 것 | 되돌리기 |
|---|---|---|---|
| 1 | **a099 + a098 — 함께 나간다** | 원장이 배제를 지고, **재claim 주체가 존재한다.** 두 발송자가 **임차로** 갈린다 — `Flush`를 안 부르므로 잠금으로는 안 갈린다 | 백업 복원 + v30 (D11) |
| 2 | a092 | 정지 경로가 동기 전송을 안 한다 | 되돌리면 1단위 상태 |
| 3 | contract | 죽은 기계가 없다 (`Flush`·인라인 재시도·`wait`·죽은 상수) | — |

**각 배포 단위의 상태가 혼자서 맞다.** 단위 **안**의 두 change는 각자 자기 gate를
통과하지만 **중간 상태로 배포하지 않는다.** 그것이 착지 순서를 고르는 것과 다른 점이고,
19라운드 A-P3 = B-P1이 *"어느 순서도 안전하지 않다"*고 판정한 이유를 없앤다.

> **⛔ 3라운드가 이 묶음을 「문장뿐」이라고 잡았다 (P-4, 두 보이스 공통).**
> `gate.sh`는 change ID 하나만 받고 그 change의 `tasks.md`만 검사한다
> (`gate.sh:67`·`:83`·`:103`). **사람이 잊으면 a099만 나간다.**
>
> **사용자 결정 7-1**: 짝을 `openspec/changes/<id>/deploy-pair.txt`로 **선언**하고
> `gate.sh`가 그것을 검사한다 — 짝 디렉터리 존재 · 짝의 미완료 task 0건 ·
> **상호 선언**(짝도 나를 적었는가). 세 검사가 다 통과해야 gate가 계속된다.
> 검사 방법과 교착을 푸는 예외는 §6.7이 진다.

| a099가 혼자 못 하는 것 | a098이 그것을 어떻게 지는가 |
|---|---|
| 만료된 임차를 다시 집기 | 배달 루프가 주기마다 PENDING을 훑는다 |
| 기동 직후 게이트 복원 (C6) | **`Runtime`의 `Recover` 콜백**이 원장을 읽고 **`Block`만** 건다 — 루프보다 먼저 완주한다 |
| 살아 있는 설정에서 임차 기간 주입 (C4) | engine 배선이 a098에 있다 |
| 운영자 승인 표면 (사용자 결정 4) | `tossctl` 하위 명령 |

> **⚠ 2단계에 조건이 붙었다 — 1라운드 A-P2.**
> a098이 `Notifier.Flush`를 부르는 형태였다면 **이 표의 2단계가 거짓이다**:
> `Flush`가 `n.mu`를 쥔 채 밀린 행 전부를 보내므로 정지가 `N × publish timeout`만큼
> 밀린다. **사용자 결정(2026-08-10, 안 1)으로 a098을 `Flush`를 안 부르는 형태로
> 고정했고**(a098 design D1.1), 그 결정이 이 표의 2단계를 참으로 만든다.

## D10 — 임차 경합은 「이미 전달됨」이 아니다 ⛔⛔

**1라운드 B-P1이 찾은 것이고, 이 change에서 가장 무거운 정정이다.**

1판은 *"취득 실패 = `owed=false`"*라고 썼다. 그 값은 오늘 **하나의 뜻**만 갖는다 —
*"운영자가 이미 받았고 창이 안 지났다"*(`notifier.go:267-268`의 주석). 그래서
`claimAndDeliver` 이탈 `:274`가 게이트를 안 잠그고 `:217`의 `owed && !sent`가
escalate를 안 하는 것이 **옳다.**

거기에 *"다른 발송자가 들고 있다"*를 같은 값으로 넣으면:

| 단계 | 일어나는 일 |
|---|---|
| 1 | 발송자 A가 claim하고 publish 전에 죽는다 |
| 2 | 행은 PENDING, 임차는 살아 있다 |
| 3 | 재시작한 B가 같은 조건을 관측 → **`owed=false`** |
| 4 | `:274`가 조용히 반환 — **게이트 래치 없음** |
| 5 | `:217`이 거짓 — **escalate 없음** |
| 6 | **미전달 critical 알림이 있는데 신규 진입이 열린다** |

**오늘은 이렇게 안 된다.** 임차가 없으므로 3에서 `owed=true`가 나오고 그냥 보낸다.
**즉 1판 설계는 crash 경로를 오늘보다 나쁘게 만든다** — 안전 불변식 4·5.

### 결정

1. `ClaimAlertForDelivery`의 결과를 **셋으로 가른다** — C3의 `ClaimDisposition`이다.
   `ClaimAcquired`·`ClaimSettled`가 오늘의 `owed`이고 `ClaimHeldElsewhere`가 새로 생긴다.
2. `ClaimHeldElsewhere`를 받은 호출자는 **발송을 건너뛰고 게이트를 안 건드린다** — C6.
   **경합을 근거로 래치하지도, 유도하지도 않는다** (사용자 결정 5-1).
3. 그 사실은 **조용하지 않다** — 구조화 로그 한 줄 (D12).
4. **회복 주체는 a098의 배달 루프다.** 만료된 뒤에도 누군가 다시 claim해야 보내지는데
   a099 시점에는 `Flush`의 프로덕션 호출자가 0이다. **그래서 둘이 한 배포 단위다**
   (D9 · 사용자 결정 3).

> **⛔⛔ 2판의 2번이 틀렸다 — 2라운드 A-P1.**
> 2판은 *"셋째를 받으면 게이트를 잠근다"*라고 적었다. `EntryGate.Block`은 reason별
> **메모리 래치이고 자동 해제가 없다**(`execgw/retry.go:498-505`). 그러면:
>
> | 단계 | 무슨 일 |
> |---|---|
> | 1 | **정상 발송 중**인 행을 다른 관측이 스친다 → `ClaimHeldElsewhere` |
> | 2 | 게이트가 `ReasonAlertUndelivered`로 잠긴다 |
> | 3 | 원래 발송자가 **성공적으로 정산한다** |
> | 4 | **푸는 경로가 없다** — `Acknowledge`를 부를 알림이 애초에 없다 |
> | 5 | **신규 진입이 사람 개입까지 영구 차단된다** |
>
> 원장은 「살아서 보내는 중」과 「죽었다」를 **구분해 주지 않는다.** 그래서 경합자가
> 추측하면 안 된다.
>
> **3판은 여기서 한 걸음 더 갔고, 그것이 3라운드 A-P1이다.** 「잠그지 마라」를
> 「원장에서 유도하라」로 읽고 **해제까지** 유도로 바꿨다 — 승인된 규범을 뒤집는
> 변경이었다. **사용자 결정 5-1은 걸음을 하나만 허용한다: 경합자는 게이트를
> 아무 방향으로도 안 건드린다**(C6).
>
> 그러면 죽은 발송자의 행은 어떻게 되는가 — **만료가 풀고 a098의 배달 루프가 다시
> 집는다.** 그 루프의 발송이 실패하면 오늘의 자리(`deliver:403-404`)가 게이트를
> 잠근다. **새 자리를 만들 필요가 없다.**
>
> `TestConcurrentObservationsOfOneConditionSendOnce`
> (`a096_one_send_per_condition_test.go:249`)는 **발송 횟수만 본다.** 이 회귀를
> 통과시킨다 — §3의 R23이 게이트 상태를 따로 본다.

## D11 — 롤백은 「임차 열을 안 읽는다」가 아니다

**1라운드 A-P11.** 1판은 *"임차 열을 안 읽으면 정확히 오늘로 돌아온다"*를 롤백이라고
불렀다. 그것은 **새 바이너리를 하나 더 만드는 일**이지 롤백이 아니다.

`SchemaVersion` 31로 올라간 DB는 **v30 바이너리가 거부한다**(설계이고 옳다).
이 저장소의 실제 복구 수단은 **migration 직전 자동 백업**이다
(`internal/journal/journal.go:255-260`이 부르고 `internal/journal/backup.go:64-146`이 구현한다).

> **2라운드 A-T3 = B-T3.** 1판은 `journal.go:280-319`를 인용했는데 그 범위는
> `applyMigration` **트랜잭션**이지 백업이 아니다. 실패한 migration의 복원 힌트는
> `journal.go:268`의 `withRestoreHint`가 붙인다. 백업은 `VACUUM INTO`이고
> `current == 0`(방금 만든 DB)일 때는 건너뛴다 — v30→v31에는 해당 없다.

**배포·롤백 절차:**

| # | 무엇 |
|---|---|
| 1 | 엔진 정지 |
| 2 | 콘솔과 엔진을 **같은 빌드**로 올린다 — 한쪽만 v31이면 다른 쪽이 DB를 거부한다 |
| 3 | migration이 v30→v31 백업을 남긴다. **그 파일 경로를 기록한다** |
| 4 | 되돌리려면: 엔진 정지 → **백업 복원** → v30 바이너리 기동 |

§4가 **부분 실패 주입 migration 테스트**를 요구한다 — `ALTER TABLE` **넷**과
`user_version`이 함께 롤백되는지.

## D12 — 임차의 움직임은 관측 가능해야 한다

**1라운드 A-P4 = B-P6.** 1판은 claim 경합·만료 탈취·소유자 불일치를 전부
기존 `EventAlertUndelivered`로 흘려보냈고, `Flush:452`는 아예 오류를 버린다.
**저장소 규칙상 침묵한 생략은 금지다.**

| 사건 | 오늘 | a099 |
|---|---|---|
| claim 경합으로 발송을 건너뜀 | `!owed`와 구분 안 됨 | **별도 이벤트** |
| 만료로 남의 임차를 가져감 | 없음 | **별도 이벤트** — 이전 보유자·임차 나이 |
| 토큰 불일치로 변경 거부 | 전송 실패와 같은 로그 | **별도 이벤트** + 남은 전송 중단 |
| 해제 실패 | `Flush:452`가 **버린다** | **버리지 않는다** |
| 운영자 목록의 임차 상태 | `Alert`·`alertSelect`에 열이 없다 | **보인다** — 누가 언제까지 들고 있는지 |

**이벤트를 내는 것은 `obs`이고 `journal`이 아니다.** `Journal`에는 logger도 event
sink도 없다 — 3라운드 B가 *"R18은 이 API로 쓸 수 없다"*고 판정한 이유가 그것이다.
**원장은 사실을 돌려주고**(C3의 `ClaimResult.Stole*`·`SettleOutcome`) **obs가 그것을
이벤트로 만든다.** 이 방향이면 R18이 관측 가능해진다.

**로그에 넣지 않는 것**: 토큰 원문 · 계좌 · 알림 본문 · payload (불변식 8).
토큰은 **배제의 근거**이고, 로그로 새면 그것을 아는 프로세스가 남의 임차를 정산할 수 있다.

## D13 — 기록만 하는 호출자는 임차를 잡으면 안 된다

**1라운드 A-P5.** `EnqueueAlert`는 `ClaimAlertForDelivery(ctx, a, 0)`에 위임하고
`owed`를 버린다(`outbox.go:120`). 그 함수가 임차를 잡게 되면:

| | |
|---|---|
| `replay.go:551`이 `parkAlert`에서 행을 쓴다 | 그 호출이 임차를 잡는다 |
| 그 호출자는 **발송하지 않는다** | 그러므로 **아무도 안 푼다** |
| a098의 배달 루프가 그 행을 보려 한다 | **만료까지 못 집는다** |

**갓 기록된 critical 알림이 자기 임차 뒤에 갇힌다.** 1판은 이 경로를 안 다뤘다.

### 결정

「내구 행이 있게 한다」와 「발송 권한을 얻는다」를 **분리한다.**

**플래그 인자로 가르지 않는다.** 공통 부분을 내부 함수로 빼고 위를 둘로 만든다.

```text
recordAlert(...)              // 비공개 — dedup · 상태 복구 · 재무장. 임차를 안 잡는다
EnqueueAlert                  = recordAlert                      // 기록만
ClaimAlertForDelivery         = recordAlert + C2의 임차 CAS      // 기록 + 발송 권한
```

> **AST가 이 결정을 강제했다.** `EnqueueAlert`의 `ast.json`은 **분기 0 · 이탈 1 ·
> 호출 1**이다 — 이 함수 안에는 **임차를 거절할 자리가 없다.** 그러므로 D13은
> `EnqueueAlert`가 아니라 **위임받는 쪽에서** 구현된다. 손으로 읽었으면
> *"EnqueueAlert에 조건을 하나 더한다"*로 끝냈을 자리다.

`EnqueueAlert`는 `ClaimResult`를 안 돌려준다(C3) — 발송 권한을 안 얻으므로 돌려줄
토큰이 없고, 그 사실이 시그니처에 드러난다.

§3이 관측한다 (R10): `replay.go` 경로로 만들어진 행을 **배달 루프가 즉시 집을 수 있다.**

## D14 — 4라운드 리뷰가 더한 것 (2026-08-12)

7 리뷰어(구현자 pass + 특화 6 + codex 교차모델)가 BLOCK 했다. 판정을 뒤집은 것은
**증명의 공백**이지 새 기능 요구가 아니었다.

### D14.1 — 제목의 주장이 자기 테스트로 증명되지 않았다

R1 테스트(`TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`)는 goroutine
둘이 **핸들 하나**를 공유했다. `Open`이 `db.SetMaxOpenConns(1)`을 걸므로
(`journal.go:174`, 주석이 스스로 *"serialising every statement onto one connection"*
이라고 적는다) 두 번째 goroutine 은 첫 번째가 **커밋하고 커넥션을 반납할 때까지**
`BeginTx`에서 막힌다. 그 테스트가 증명한 것은 순차 속성이다.

`-race` exit 0 도 이것을 메우지 않는다. race detector 는 Go 메모리 경합을 보지,
직렬화된 SQL 두 개가 배제되는지를 보지 않는다.

**고침**: `TestSeparateHandlesRaceForOnePendingRow` — 발송자 수만큼 핸들을 같은
파일에 연다. 커넥션이 달라야 `BEGIN IMMEDIATE` 둘이 진짜로 겹친다.

**그 테스트가 새 경로를 실제로 짚는다는 것을 뮤테이션으로 쟀다.** `busy_timeout`을
0으로 낮추면:

| 테스트 | 결과 |
|---|---|
| 옛 R1 (핸들 하나) | **PASS** — Go 커넥션 풀이 줄 세우므로 SQLITE_BUSY 가 안 난다 |
| 새 테스트 (핸들 여덟) | **FAIL** — `database is locked (5) (SQLITE_BUSY)` |

옛 테스트가 **구조적으로 도달할 수 없는** 실패를 새 테스트가 잡는다. 커버리지
주장이 주장이 아니라 측정값이다.

### D14.2 — 임차를 **놓는** 경로가 일관되지 않았다

| 자리 | 무엇이 틀렸나 | 고침 |
|---|---|---|
| `deliver` 반납 | 루프를 빠져나오는 두 길 중 하나가 ctx 취소인데, 반납을 **그 죽은 ctx**로 냈다. `BeginTx`가 즉시 실패하므로 종료 중에는 반납이 **항상** 실패했다 | `releaseCtx` = `WithoutCancel` + `DefaultBusyTimeout`. 정리는 그것을 유발한 취소를 물려받지 않는다 |
| `Flush` 반납 | 같은 노출 + `_, _ =`로 결과 전부 폐기 | 같은 `releaseCtx`, 결과를 읽고 `logLeaseLost` |
| `AcknowledgeAlert` | PENDING 을 떠나는 네 경로 중 **유일하게** `alertClaimCleared`를 안 썼다. 배제는 안 깨지지만(술어가 전부 `state = PENDING` 안) 토큰이 정산된 행에 무기한 남고 `backup.go`의 `VACUUM INTO`가 복사본마다 그것을 실어 나른다 | 넷째 경로도 같은 상수를 쓴다 |
| 반납 시점의 상실 | 루프 **안**에서 알면 에스컬레이션 안 함, **반납 때** 알면 그대로 `Gate.Block` — 운영자가 승인한 직후 엔진이 진입을 다시 막는다 | 두 발견 시점이 같은 결론을 낸다 |

### D14.3 — 한 사실에 두 결론

`SettleNotFound`는 원장이 *"That one **is** an error"*라고 못박은 유일한 값인데,
실패-기록 경로가 그것을 `!= SettleApplied` 한 덩어리로 삼켜 `claimed_by`가 빈
`engine.alert_claim_lost`(= 평범한 경합)로 보고했다. 성공 경로에서는 같은 값이
게이트를 잠그는 오류였다.

`logLeaseLost`가 이제 넷을 가른다: 정산됨 / 경합(이름 있음) / **이름 없는 상실**
(내가 이미 반납한 것 — 아무도 안 가져갔다) / 소실(오류·게이트).

`deliver`의 `if markErr == nil { return true, false }`는 오늘 도달 불가였지만 방향이
fail-open 이었다 — 다섯째 outcome 이 생기면 조용히 "배달 성공"이 된다. `default` arm 이
그것을 fail-safe 로 뒤집는다.

### D14.4 — `ClaimHeldElsewhere`는 오늘 배선에서 「배제가 도는 중」이 아니다

D10 이 이 분기를 *"경합은 이미 전달됨이 아니다"*로 세운 것은 옳다. 4라운드가 더한 것은
**오늘 그 분기에 닿는 원인이 하나뿐**이라는 사실이다: 프로덕션 Notifier 는
`exitwiring.go:73` 하나이고 `Notifier.Flush`는 프로덕션 호출자가 0이다. 그러므로 이
분기의 실제 유일한 원인은 **죽은 발송자가 남긴 임차**이고, 그때 info 한 줄은 틀린 답이다.

고침은 등급과 정보량까지다 — `Warn` + `claim_expires_at`. **회수는 여전히 a098 이다**
(D9). a099 는 그 상태를 *만들지 않는 것*(D14.2)과 *보이게 하는 것*까지 진다.

### D14.5 — lease > 예산은 기본값에서만 참이었다

`Notifier.Attempts`·`RetryDelay`·`Ntfy.Timeout`·`Options.AlertLease`는 두 패키지에
흩어진 **독립 설정 넷**인데 둘을 비교하는 코드가 없었다. 예산이 임차보다 길면 발송자는
이 코드가 쓰라고 시킨 예산을 쓰는 도중 행을 잃고 두 번째 발송자가 발행한다 —
**아무 데도 잘못된 것이 없는 채로**, 그래서 아무도 못 찾는다.

`Journal.AlertLease()` + `Notifier.checkAlertLease()`(`sync.Once`)가 첫 claim 전에
`EventAlertLeaseTooShort`를 한 번 낸다. 배선의 성질이므로 알림마다 반복하지 않는다.

### D14.6 — 배포 제약 둘 (코드로 못 막는다)

**(a) 이미 열려 있는 v30 프로세스는 임차를 통째로 우회한다.** 마이그레이션이
additive 라 옛 바이너리의 outbox SQL 이 그대로 돈다 — 그 빌드는 claim 열을 읽지도
쓰지도 않으므로 v31 발송자가 쥔 행을 그냥 발행한다. `ErrSchemaTooNew`는 **그 뒤에
여는** 프로세스만 막지, **이미 열려 있는** 엔진을 fence 하지 못한다.
⇒ 롤링 교체 금지. 옛 컨테이너가 **완전히 내려간 뒤** 새 것을 올린다.

**(b) 스키마 31 은 저널 단위로 단방향이다.** 첫 a099 바이너리가 운영자의 저널을
31 로 올리는 순간 현재 main 빌드(30)는 `ErrSchemaTooNew`로 그 파일을 거부한다.
거부는 안전한 방향이지만(옛 코드가 새 모양을 오독하지 않는다) **이 제품에서
"엔진이 못 뜬다"는 손절이 없다는 뜻이다.** 바이너리만 되돌리는 것은 롤백이 아니라
장애다. 복구는 `backup.go`의 절차이고 v30 사본 경로는
`schema_meta.pre_migration_backup_path`(`journal.db.v30-pre-v31.<stamp>.bak`)에 있다.
