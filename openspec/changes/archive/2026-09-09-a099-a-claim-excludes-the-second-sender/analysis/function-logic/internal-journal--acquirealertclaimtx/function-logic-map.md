# Function Logic Map: `acquireAlertClaimTx`

- Source: `internal/journal/alert_claim.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **`check_analysis.py`가 이 번들을 요구하지 않는다** — 새 파일의 새 함수다.
> 그런데도 만든 이유는 **a099의 배제가 전부 이 함수 안에 있기 때문**이다.
> design C2(청구 술어), C3(세 갈래 처분), D3/R14(토큰), R12(만료 저장)는 전부
> 이 함수의 분기를 근거로 쓴다. **근거로 쓰는 문서는 AST 산출물이 먼저다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 호출자가 연 쓰기 트랜잭션 | `ClaimAlertForDelivery` `outbox.go:259` · `ClaimAlertByID` `:295` | 롤백은 호출자의 defer |
| `id` | 존재하는 행 | 호출자 | B2 → `ErrAlertNotFound` |
| `claimant` | **공백 아님** | **호출자가 이미 검증했다** — 두 진입점 모두 | 여기서는 `TrimSpace`만 하고 안 거절한다 |
| `now` | 호출자의 시계 | `j.clk.Now()` | 술어의 두 인자를 만든다 |
| `lease` | `> 0` | `j.alertLease` (기본 81s) | `expires = now + lease` |
| `alertClaimSkew` | `2s` 상수 | `:47` | 역행 시계 판정의 여유 |

**불변식 — a099의 전부**: *"청구 가능한지는 UPDATE의 WHERE 절만 판정한다."*
doc comment `:210-213`이 그것을 적는다 — 행을 미리 읽는 이유는 **훔친 상대를
이름 붙이기 위해서**이고, 0행일 때 다시 보는 이유는 **거절을 설명하기 위해서**다.
**Go에 판정 사본을 두면 술어가 둘이 되고 갈라진다.**

## Branches and early returns

AST 열거 — 분기 8 · 이탈 8 · 호출 23 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:218` | `readAlertLeaseTx` 실패 | 없음 | (아래 두 갈래) | 없음 |
| B2 `:219` | `errors.Is(err, sql.ErrNoRows)` | 없음 | `:220` `ErrAlertNotFound` | 없음 |
| — 이탈 `:222` | 그 밖의 읽기 오류 | 없음 | 포장 오류 | 없음 |
| B3 `:226` | `mintAlertClaimToken` 실패 | 없음 | `:227` err | 없음 (`crypto/rand` 실패) |
| B4 `:237` | UPDATE 실패 | 없음 | `:238` 포장 오류 | 없음 |
| B5 `:241` | `RowsAffected` 실패 | 없음 | `:242` 포장 오류 | 없음 |
| **B6 `:244`** | **`n == 1` — 잡았다** | **임차 열 넷을 쓴다** (`:232-236`) | `:258` `ClaimAcquired` + 토큰 | `TestClaimingAFreshAlertOwesDelivery` |
| **B7 `:253`** | **`before.token != ""` — 남의 것을 뺏었다** | 없음 (표시만) | `Stole`·`StoleBy`·`StoleAt` | `TestAnExpiredClaimIsPickedUpByAnotherSender` |
| **B8 `:263`** | **0행 + 행이 PENDING이 아니다** | 없음 | `:264` `ClaimSettled` | 없음 |
| — 이탈 `:266` | 0행 + PENDING | 없음 | **`ClaimHeldElsewhere`** + 현 보유자 정보 | `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` |

### 판정은 `:234`의 `alertClaimable` 안에 있다

```sql
WHERE id = ? AND state = ? AND (claim_token = ''
     OR claim_expires_at IS NULL
     OR claim_expires_at <= ?      -- nowStamp: 만료는 포함(inclusive)
     OR claimed_at > ?)            -- now+skew: 미래 발급은 엄격 초과
```

**네 갈래 전부 「행이 말하는 것」만 읽는다.** 경합자의 lease 길이가 안 들어간다 —
들어가면 짧은 임차를 든 자가 남의 유효한 임차를 자기 시계로 재해석해
**아무도 버리지 않은 행을 훔친다**(R12, `TestALeaseHolderWithAShorterLeaseCannotStealALiveClaim`).

**만료가 저장 값인 이유가 그것이다.** 발급자가 `now+lease`를 계산해 행에 쓰고,
판정은 그 저장 값만 본다.

### B6과 이탈 `:266`이 「하나만 성공」의 전부다

`n == 1`은 SQLite가 그 UPDATE를 실제로 적용했다는 뜻이다. 같은 순간 다른
트랜잭션이 같은 술어로 들어오면 **행이 이미 임차되어 술어가 거짓**이 되고 0행을 받는다.
`n`은 Go가 세는 값이 아니라 **드라이버가 돌려주는 값**이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readAlertLeaseTx` `:217` | **판정이 아니라 설명용 선행 읽기** | 같은 트랜잭션 — UPDATE가 볼 상태와 같다 | `ast.json` calls |
| `errors.Is` `:219` | `ErrNoRows` 판별 | — | 같음 |
| `mintAlertClaimToken` `:225` | **토큰 발급** | `crypto/rand`. 실패는 오류로 올린다 | 같음 |
| `RFC3339` `:229`, `:235`, `:236` | 저장 형식 — **초 단위로 자른다** | 그 절단이 `alertClaimSkew` 하한 2초의 이유다 | 같음 |
| `now.Add(lease)` `:230` · `now.Add(alertClaimSkew)` `:236` | 만료·역행 경계 | — | 같음 |
| **`tx.ExecContext` `:231`** | **판정 + 취득을 한 문장으로** | 실패면 B4 | 같음 |
| `res.RowsAffected` `:240` | 하나 잡았나 | 실패면 B5 | 같음 |
| `strings.TrimSpace` `:235`, `:249` | 저장 값과 반환 값을 같은 모양으로 | — | 같음 |
| `claimStamp` `:256`, `:270`, `:271` | nullable 타임스탬프 읽기 | 없음/공백/파싱 실패 전부 zero time | 같음 |

**live bindings — 호출자 둘, 둘 다 같은 파일**:
`ClaimAlertForDelivery` (`outbox.go:259`, `owed`일 때만) ·
`ClaimAlertByID` (`alert_claim.go:295`, `Flush`가 쓴다).

## State mutations and fallbacks

- **UPDATE 하나.** `claim_token`·`claimed_by`·`claimed_at`·`claim_expires_at`.
  **다른 열은 안 건드린다** — state도, attempts도.
- **0행일 때 아무것도 안 쓴다.** 두 갈래(`ClaimSettled`/`ClaimHeldElsewhere`)는
  **선행 읽기로 설명만 한다.** 그 읽기는 같은 트랜잭션 안이라 UPDATE가 본 것과 같다.
- **폴백은 「열린 쪽」이다**: 만료·`NULL`·역행 시계는 전부 **다시 청구 가능**으로 간다.
  죽은 발송자가 행을 영원히 잠그지 않는다.
- **토큰은 반환에만 담기고 로그에 안 나간다.** `ClaimResult.Token`의 주석 `:84-87`이
  이유를 적는다 — 읽을 수 있는 자는 남의 발송을 정산할 수 있다.

## Safety conclusion

- **Safe edit boundary**: 판정을 Go로 옮기는 편집이 금지선이다.
  `before`를 읽어 Go에서 만료를 비교하고 UPDATE의 술어를 줄이는 형태가 그것인데,
  **그 순간 술어가 둘이 되고 두 사본은 반드시 갈라진다.**
  `alertClaimable`에 `lease`를 인자로 넣는 편집도 금지다(R12).
- **High-risk impact**: **yes** — 원장이고, 이 함수가 잘못 잡으면
  critical 알림이 두 번 나가거나 아무도 안 보낸다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B1·B2·B3·B4·B5에 테스트가 없다.** 드라이버·`crypto/rand` 오류 경로다.
    **`not-applicable`: 이 change는 다섯을 근거로 아무것도 주장하지 않는다.**
    **B2(`ErrAlertNotFound`)는 그중 유일하게 도달 가능한 것이다** —
    `ClaimAlertByID`에 없는 id를 주면 온다. `Flush`가 그 오류를
    `notifier.go:693`에서 이름으로 걸러 내므로 **경로는 살아 있고 테스트만 없다.**
  - **B8(`ClaimSettled`)에 테스트가 없다.** 목록을 읽은 뒤 청구하기 전에 다른
    누군가가 정산한 경우다. `ClaimAlertForDelivery`는 `!owed`로 먼저 걸러 여기
    안 오고, `ClaimAlertByID`로는 올 수 있다. **관측되지 않았다.**
  - **이 함수는 발송을 안 막는다.** 자기 임차를 넘겨 멈춘 발송자의 push는
    만료 뒤에도 도착한다. 보장은 **누가 정산하는가**에 대한 것이다.
