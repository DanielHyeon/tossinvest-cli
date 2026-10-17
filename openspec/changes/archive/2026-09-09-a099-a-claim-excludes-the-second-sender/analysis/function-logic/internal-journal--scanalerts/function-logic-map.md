# Function Logic Map: `scanAlerts`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 함수가 「토큰은 원장을 안 떠난다」를 지키는 자리다.**
>
> a099는 `alertSelect`에 열 셋을 더했다 — `claimed_by`, `claimed_at`,
> `claim_expires_at`. **`claim_token`은 안 더했다.** 운영자의 백로그는
> *"아무도 안 보내는 행"*과 *"누가 보내고 있는 행"*을 갈라야 하지만,
> 토큰을 읽을 수 있는 자는 **남의 발송을 정산할 수 있다.**
>
> 이 함수는 그 SELECT의 결과를 받는 쪽이므로, **`Scan`의 인자 목록이 곧
> 그 약속의 증거다.** `:565-567`에 `claim_token`을 받는 변수가 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rows` | `alertSelect`가 만든 커서 | 세 호출자 | 열 순서가 어긋나면 B2가 오류 |
| 열 개수 | **18** | `alertSelect` `:497-501` | `Scan`의 인자 수와 정확히 맞아야 한다 |
| nullable 넷 → 셋 | `last_attempt_at`·`delivered_at`·`acknowledged_at` + **`claimed_at`·`claim_expires_at`** | 스키마 | `parseOptionalStamp`가 nil을 준다 |

**불변식**: *"`Alert`에 토큰이 없다."* 구조체에도 없고 이 함수도 안 읽는다.

## Branches and early returns

AST 열거 — 분기 3 · 이탈 3 · 호출 12 · 대입 12 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:552` | `rows.Next()` — 행마다 | `out`에 append `:579` | — | 기존 |
| B2 `:565` | `rows.Scan` 실패 | 없음 | 이탈 `:568` 포장 오류 | 없음 |
| B3 `:581` | `rows.Err()` | 없음 | 이탈 `:582` 포장 오류 | 없음 |
| — 이탈 `:584` | 정상 | — | `out, nil` | 기존 다수 |

**a099가 이 함수에 더한 것은 분기가 아니라 대입 셋이다** (`:576`, `:577`, `:578`).
`ClaimedBy`·`ClaimedAt`·`ClaimExpiresAt`. **제어 흐름은 안 바뀌었다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rows.Next` `:552` | 커서 전진 | — | `ast.json` calls |
| **`rows.Scan` `:565`** | **열 18개를 구조체로** | 열 순서·타입이 `alertSelect`와 맞아야 한다 | 같음 |
| `parseStamp` `:570` | `created_at` — NOT NULL | 파싱 실패는 zero time | 같음 |
| `parseOptionalStamp` `:571`, `:572`, `:573`, **`:577`, `:578`** | nullable 다섯 | 없음/공백/파싱 실패 전부 nil | 같음 |
| `rows.Err` `:581` | 커서 오류 | — | 같음 |

**live bindings — 호출자 셋, 전부 같은 파일**:
`PendingAlerts` `:521` · `LookupAlert` `:544` · (`alertSelect`를 쓰는 read 경로).

**a098의 배달 루프와 콘솔이 이 함수의 결과를 본다.**
`Alert.ClaimedBy`가 채워지므로 백로그 화면이 「누가 보내는 중」을 보일 수 있다.

## State mutations and fallbacks

- **읽기만 한다.** 원장을 안 바꾼다.
- **`out`이 nil로 시작한다** (`:551`). 행이 없으면 `nil, nil`을 돌려주고,
  `LookupAlert`가 `len(alerts) == 0`으로 `ErrAlertNotFound`를 만든다.
- **폴백**: 파싱 못 하는 타임스탬프는 조용히 nil/zero가 된다.
  `claimStamp`(`alert_claim.go:200`)가 같은 규약을 임차 쪽에서 쓴다 —
  *"there is no lease to speak of"*.

## Safety conclusion

- **Safe edit boundary**: **`Scan`의 인자 목록과 `alertSelect`의 열 목록이 한 쌍이다.**
  한쪽만 바꾸면 B2가 런타임에 오류를 낸다 — 컴파일러가 안 잡는다.
  **`claim_token`을 이 목록에 더하는 편집이 금지선이다.**
- **High-risk impact**: **yes (읽기)** — 진입 게이트가 반응하는 `UndeliveredCount`는
  다른 함수지만, 백로그 화면과 a098의 루프가 이 결과를 본다.
  **잘못 읽으면 운영자가 잘못 본다.**
- **덮이지 않은 것을 이름으로 적는다**:
  - **B2·B3에 테스트가 없다.** 드라이버 오류 경로다.
    **`not-applicable`: 이 change는 둘을 근거로 아무것도 주장하지 않는다.**
  - **열 목록의 짝을 검사하는 것은 컴파일러가 아니라 테스트다.**
    `TestPendingAlertsProjectTheLeaseButNeverTheToken`
    `a099_lease_lifecycle_test.go:391`이 그 짝과 **토큰 부재**를 같이 본다.
  - **`Alert` 구조체가 커졌다.** 열 셋이 늘었고, 이 구조체를 JSON으로 내보내는
    경로가 있으면 그 표면도 커진다. **이 change에서 확인 안 했다** —
    `claim_token`이 없으므로 유출 위험은 없지만, **표면이 늘었다는 사실은 적어 둔다.**
