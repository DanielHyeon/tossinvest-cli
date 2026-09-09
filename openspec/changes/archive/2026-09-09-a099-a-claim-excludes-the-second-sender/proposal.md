# a099 — claim이 두 번째 발송자를 배제한다

base: `285c7619110d2f8c53a1d9ddfbadd16ad0e9e53e` (`base-commit.txt`)

## 한 줄

`ClaimAlertForDelivery`는 **claim을 하지 않는다.** PENDING 행에 대해 아무 UPDATE도 하지
않고 `owed=true`를 돌려준다. 두 발송자를 가르는 것은 원장이 아니라 `obs.Notifier`의
**Go 뮤텍스 하나**다. a099는 이름이 이미 약속한 것을 실제로 만든다.

## 왜 이것이 별도 change인가 — 그리고 왜 a092·a098보다 먼저인가

**사용자 결정 (2026-08-10)**: a099 신설 + 4단계(a099 → a098 → a092 → contract).
**사용자 결정 3 (2026-08-11)**: 그 4단계가 **3 배포 단위**가 됐다 —
**a099와 a098이 함께 나간다**(design D9). 2라운드 A-P10 = B-P6이 *"a099 단독 배포는
오늘보다 나쁘다"*를 증명했기 때문이다.

19라운드는 a092와 a098이 **어느 순서로 착지해도 안전하지 않다**는 것을 확정했다
(A-P3 = B-P1, a092 `review.md` §19). 세 가지 착지 순서를 놓고 고른 것이 잘못이었다.
셋 다 **두 change의 범위를 고정한 채 순서만 바꾸는** 형태였고, 정석은 범위를 다시 나누는
것이다 — expand → migrate → contract.

expand 단계를 실제로 그리자 질문 하나가 남았다. **두 발송자가 공존하는 동안 무엇이
가르는가?** 답이 없었다. 그래서 a099가 먼저다.

### 이 change가 여는 것

| 단계 | change | 이 change 없이는 |
|---|---|---|
| expand | **a099** — 원장이 배제를 진다 | — |
| migrate 1 | a098 — 배달 루프 신설 (**`Flush`를 안 부르는 형태**) | 루프와 `claimAndDeliver`가 같은 행을 동시에 publish한다 |
| migrate 2 | a092 — 정지 경로에서 동기 전송 제거 | 보낼 주체가 없는 창이 생긴다 |
| contract | 죽은 기계 제거 | 지울 수 없다 — `n.mu`가 유일한 배제이므로 |

### ⚠⚠ 「a098이 배제 없는 발송자를 만든다」는 **그대로는 거짓이다** (1라운드 A-P1)

리뷰가 확인시켜 준 것:

| | |
|---|---|
| `&obs.Notifier{}` 생성처 | `exitwiring.go:73` **하나**, 호출은 `gateway.go:280` 하나 |
| 두 번째 엔진 프로세스 | `internal/enginelock`의 flock이 **거부한다** |

**즉 a098의 루프가 `Notifier.Flush`를 부르면 같은 `n.mu`가 가른다.** 그 형태에서는
a099가 필요 없다.

**그런데 그 형태는 정지를 붙잡는다** (A-P2). `Flush`는 `n.mu`를 쥔 채 밀린 행 전부를
publish하고 동기 정지 알림이 같은 뮤텍스를 쓴다 — `N × publish timeout`, N에 상한 없음.

| a098의 형태 | 배제 | 정지 지연 |
|---|---|---|
| (i) `Flush`를 부른다 | `n.mu`가 한다 | **`N × publish timeout`** — 불변식 4 위반 |
| **(ii) `n.mu` 밖에서 보낸다** | **아무것도 안 한다** | 없음 |

**사용자 결정 (2026-08-10, 안 1)**: a098을 **(ii)로 고정**한다.
a098 design D1.1·D1.2가 그 정본이고, a098 proposal §범위 1이 그렇게 바뀌었다.

**그러므로 a099가 필요한 정확한 이유는 이것이다** — *"a098이 발송자를 만드니까"*가 아니라
***"a098이 (ii)여야 하고, (ii)는 원장 배제를 요구하니까"***.

## 무엇이 실제로 일어나고 있는가 — 전부 AST가 열거한 값이다

### `ClaimAlertForDelivery`가 PENDING 행에 하는 일: 없다

AST가 열거한다(`analysis/function-logic/internal-journal--journal.claimalertfordelivery`,
분기 11 / 이탈 10 / defer 1, `outbox.go:169-261`).

기존 행 경로는 B5 `:195`(`case err == nil`)에서 시작한다. 그 아래에서
**UPDATE는 정확히 하나뿐이고 B6 `:197`(`if rearm`) 안에 있다.** rearm이 거짓이면 분기
전체를 건너뛰고 이탈 `:240` `return existing, owed, tx.Commit()`로 나간다.

rearm이 언제 거짓인가는 `claimOwed`의 AST가 열거한다
(`internal-journal--claimowed`, 분기 8 / 이탈 7, `outbox.go:269-315`).
B2 `:276` `case AlertPending` → 이탈 `:278` **`return true, false`**.

| 상태 | owed | rearm | B6 `:197` | 행에 일어나는 일 |
|---|---|---|---|---|
| **PENDING** | **true** | **false** | **거짓** | **아무것도 없다** |
| DELIVERED/ACKNOWLEDGED + 창 경과 | true | true | 참 | PENDING으로 재무장 |
| 미지의 상태 | true | true | 참 | PENDING으로 복구 |
| 행 없음 | true | — | (INSERT 경로) | INSERT — `:248`이 PENDING을 쓴다 |

**첫 행이 이 change의 전부다.** 발송을 앞둔 가장 흔한 경우에 원장은 아무 표시도 남기지
않는다. 두 번째 발송자가 같은 순간에 들어오면 같은 행을 읽고 같은 `owed=true`를 받는다.

### 지금 배제를 지는 것은 원장이 아니다 — 코드가 그렇게 적어 뒀다

세 곳이 같은 사실을 서로 다른 자리에서 진술한다.

| 자리 | 문장 |
|---|---|
| `outbox.go:166-168` | *"Exclusion against a concurrent claimer is the caller's: obs.Notifier holds its delivery mutex across the claim and the send."* |
| `notifier.go:230-234` | *"Claiming outside it lets two observations … each conclude the send is owed, and each publish"* |
| `notifier.go:431-433` | *"The same mutex the notify path holds. Without it a flush and an observation can publish the same row at the same moment"* |

`claimAndDeliver`의 AST가 그 뮤텍스의 구간을 열거한다
(`internal-obs--notifier.claimanddeliver`, 분기 4 / 이탈 3 / defer 1,
`notifier.go:238-277`). defer `:242`가 `n.mu.Unlock()`이고, claim 호출은 `:244`,
발송 이탈은 `:276` `return n.deliver(ctx, id, e), true, nil`이다.
**뮤텍스가 claim과 send를 함께 덮는다.**

`Flush`는 claim을 아예 부르지 않는다 — `PendingAlerts`를 읽고 곧장 publish한다
(a098이 잰 `internal-obs--notifier.flush`, 분기 6 / 이탈 4, `notifier.go:427-462`).
**두 발송 경로 중 하나는 claim 자체를 거치지 않는다.**

### CAS는 이중 발송을 막지 않는다

`MarkAlertDelivered`의 AST(`internal-journal--journal.markalertdelivered`,
분기 1 / 이탈 2, `outbox.go:337-348`)는 UPDATE 하나와 오류 검사 하나뿐임을 열거한다.
그 UPDATE의 술어가 `WHERE id = ? AND state = ?`(PENDING)이다.

**그 CAS는 네트워크 호출이 끝난 뒤에 돈다.** `claimAndDeliver`가 `:276`에서 `deliver`를
부르고, `deliver`가 publish에 성공한 뒤에 `MarkAlertDelivered`를 부른다
(`notifier.go:356`). 즉 CAS가 거르는 것은 **이중 정산**이지 이중 발송이 아니다.
두 번째 발송자의 푸시는 이미 운영자의 전화기에 도착해 있다.

`MarkAlertAttemptFailed`도 같은 모양이다(분기 1 / 이탈 2, `outbox.go:352-363`).

### 그래서 오늘 사고가 안 나는 이유는 하나뿐이다

**프로덕션 발송자가 하나다.**

| 함수 | 프로덕션 호출자 |
|---|---|
| `ClaimAlertForDelivery` | `notifier.go:244` — `claimAndDeliver` 하나 |
| `EnqueueAlert` (claim에 위임, owed를 버린다) | `replay.go:551` 하나 |
| `Notifier.Flush` | **0** (a098이 측정) |

경합할 상대가 없다. **a098이 두 번째 발송자를 만드는 순간 이 사실이 끝난다.**

## 범위

**하는 것 — 이것뿐이다.**

**값은 전부 design D-CORE가 정본이다.** 여기서 다시 적지 않는다.

1. `alert_outbox`에 임차 열 **넷**을 더하는 additive 스키마 하나 (`schemaV31` · C1).
   만료는 **발급자가 계산해서 저장한다** — 판정자의 설정이 남의 임차를 재해석하지
   못하게 하는 것이 그 열의 존재 이유다 (사용자 결정 6-1).
2. `ClaimAlertForDelivery`가 발송을 넘겨줄 때 그 임차를 **CAS로 취득**하고
   **`ClaimResult`를 돌려준다**(C3). **취득 실패는 `owed=false`가 아니다** —
   `ClaimHeldElsewhere`라는 **셋째 결과**다.
3. 임차 **만료**로 죽은 발송자의 행이 자동으로 다시 claim 가능해진다 (C2).
4. 정산(`MarkAlertDelivered`)과 **포기**(`ReleaseAlertClaim` 신설)가 임차를 **푼다**.
   **실패 기록(`MarkAlertAttemptFailed`)은 안 푼다** — 같은 발송자가 아직 재시도한다.
5. claim 없이 발송하는 경로를 **하나 남기지 않는다** — `Flush`도 claim을 거친다.
6. **진입 게이트의 트리거는 오늘 그대로 둔다**(C6 · 사용자 결정 5-1).
   바뀌는 것은 **경합에 아무것도 안 한다**는 것뿐이다 — 잠그지도, 풀지도 않는다.
   기동 시 원장에서 `Block`만 복원하는 것은 **a098**이 진다.

**안 하는 것.**

- `n.mu`의 구간을 **줄이지 않는다.** 오늘 그대로 둔다 — 그것이 a092의 일이다.
- 배달 루프를 만들지 않는다 — a098의 일이다.
- 상태 셋(`PENDING`/`DELIVERED`/`ACKNOWLEDGED`)을 **늘리지 않는다.** 이유는 design D1.
- `UndeliveredCount`·`PendingAlerts`·`AcknowledgeAlert`의 술어를 바꾸지 않는다.
- 운영자 표면(`tossctl`)·알림 등급·예산 상수를 건드리지 않는다 — a098·a092다.
- **정확히 한 번 발송을 보장하지 않는다.** 유계 실행 밖에서는 at-least-once다(design C7).

## 무엇이 오늘과 같고, 무엇이 다른가

> **⚠⚠ 1판은 「동작 변화 0」이라고 적었다. 1라운드 A-P7 = B-P1이 그것을 반증했다.**
> 참인 것은 더 좁다: **발송자가 하나이고 아무도 안 죽었을 때 발송 결정이 같다.**

**같은 것**: 발송자가 하나이므로 임차 CAS는 경합 상대가 없어 항상 성공한다.
오늘 보내는 것을 그대로 보내고, 오늘 안 보내는 것을 그대로 안 보낸다.
읽기 술어 다섯도 그대로다(design D1).

**다른 것 — 전부 이름을 적는다:**

| 무엇 | 오늘 | a099 |
|---|---|---|
| **발송자가 claim 후 죽는다** | 다음 관측이 그냥 보낸다 | 임차가 살아 있는 동안 **후속 발송자가 못 보낸다.** 만료 뒤 **a098의 배달 루프**가 집고, 그 발송이 실패하면 **오늘의 자리**(`deliver:403-404`)가 게이트를 잠근다. **경합 자체는 게이트를 안 건드린다**(C6) |
| **claim 경합으로 발송을 건너뛴다** | 그런 결과가 없다 | 건너뛰고 **구조화 로그 한 줄**을 남긴다. 게이트는 **안 건드린다** — 잠그지도 풀지도 않는다(C6) |
| **아직 시도조차 안 한 미전달 critical 행** | 진입을 안 막는다 | **여전히 안 막는다.** 3판은 막는다고 적었고 **사용자 결정 5-1이 그것을 뺐다** |
| PENDING 경로의 원장 쓰기 | 없음 | **UPDATE 하나 + fsync** (design D2) |
| 그 쓰기가 실패하면 | 그런 실패가 없다 | `claimAndDeliver` B1이 **게이트를 잠그고 발송을 포기한다** — 실패 경로가 하나 는다 |
| 기록 전용 호출자(`replay.go:551`) | 그냥 기록 | **임차를 안 잡는 모드로 위임한다**(design D13). 안 그러면 갓 쓴 행이 자기 임차 뒤에 갇힌다 |

**되돌리기**: 임차 열을 안 읽는 것은 롤백이 아니라 **새 바이너리**다.
`SchemaVersion` 31 DB는 v30이 거부하므로 실제 복구는 **migration 직전 백업 복원**이다
(design D11).

## 안전 불변식 대조

| # | 불변식 | a099의 관계 |
|---|---|---|
| 1 | 승인 없는 LIVE 주문 side effect 금지 | 무관 — 주문 경로를 안 건드린다 |
| 3 | 토글 OFF = upstream 동작 | 토글 없음. **「동작 변화 0」은 거짓이다** — 발송자가 죽은 뒤의 재발송 시점이 바뀐다(아래 표 첫 행). **진입 게이트의 트리거는 안 바꾼다**(C6 · 결정 5-1) |
| 4 | **손절·비상 청산의 즉시성** | **위험 자리다.** 임차 취득은 `claimAndDeliver`가 이미 잡은 트랜잭션 **안**에서 일어나고 **새 연결·새 뮤텍스를 만들지 않는다.** 대신 **SQL UPDATE 한 문장과 WAL fsync가 는다** — design D2가 이것을 지는 요구이고, §3.5가 재기 전에는 불변식 4를 만족한다고 적지 않는다 |
| 5 | 원장 스키마는 High-risk | **예.** Pre-Edit 선언은 design D5 |
| 6 | 보수 방향만 | 배제를 **더 강하게** 만든다. 임차 실패 시 발송을 건너뛰는 쪽이지 더 보내는 쪽이 아니다. **진입 차단을 푸는 조건은 하나도 안 완화한다** — 3판 C6이 그것을 완화했고 3라운드 A-P1이 잡았다 |
| 8 | 시크릿·개인정보 로그 금지 | **임차 로그에 토큰 원문·계좌·알림 본문을 안 넣는다** (task 4.5c) |
| 7 | 운영 토글 flip은 사람이 승인 | 무관 |

**진입 게이트와의 결합이 이 change의 가장 위험한 표면이다.** `UndeliveredCount`가 세는
값이 진입 차단을 좌우한다(`outbox.go:407`의 주석 *"the number the entry gate reacts
to"*). 상태를 하나 더하는 설계였다면 그 카운트에서 행이 조용히 빠져나가 **차단이 스스로
풀렸을 것이다.** design D1이 상태를 안 늘리는 이유가 이것이고, **R21이 그것만 관측한다.**

> **이 표면에서 이미 두 번 미끄러졌다.** 2판은 경합에 게이트를 **잠갔고**(2라운드 A-P1),
> 3판은 게이트를 **유도로 바꾸면서 푸는 쪽까지 자동화했다**(3라운드 A-P1).
> **4판은 다섯 자리 전부를 안 건드린다**(C6의 표). §5.3b가 그 diff가 비어 있음을 확인한다.

## 배포 제약 — 침묵한 생략이 아니다

`SchemaVersion`은 지금 30이다(`schema.go:6`). a099는 31을 쓴다.
**낮은 버전의 바이너리는 이 DB를 거부한다** — schema.go의 규칙이 그렇게 설계돼 있고,
그것이 옳다. 배포 시 콘솔과 엔진이 같은 빌드여야 한다.

**⛔ 그리고 a099를 혼자 배포하지 않는다** (사용자 결정 3 · 2026-08-11).
`Flush`의 프로덕션 호출자가 0이므로 **claim한 뒤 죽은 행을 다시 집을 주체가 없다.**
오늘은 다음 관측이 그냥 보내는데 a099 단독은 살아 있는 임차 동안 억제한다 —
**오늘보다 나쁘다.** a098과 **같은 배포 단위**로 나간다 (design D9).

## 열려 있는 것

| 무엇 | 왜 아직 |
|---|---|
| 임차 기간의 `margin`과 `:skew` 값 | 측정 전. §3.4·§3.4b가 정한다. 유도식과 하한은 design C4·C2 |
| `claim_token`의 생성 방식 | **재사용 없는 값**이라는 것만 정해져 있다(C1). 형태는 §4.3 |
| **CLI 승인이 실행 중 엔진의 게이트에 닿는 방법** | **a098이 진다** (a098 design D7.1). `EntryGate`는 엔진 메모리 안에 있고 CLI는 다른 프로세스다. a099는 이 표면을 안 만든다 |
| `Flush`를 claim 경로로 옮길 때 그 함수의 잠금 구간 | a099는 claim만 더하고 **잠금은 안 건드린다.** 구간 재설계는 a092 |
