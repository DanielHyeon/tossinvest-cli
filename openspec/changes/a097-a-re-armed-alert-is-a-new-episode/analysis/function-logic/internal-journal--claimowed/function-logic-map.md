# Function Logic Map: `claimOwed`

- Source: `internal/journal/outbox.go` (L245-291)
- AST evidence: `ast.json` (8 branches, 7 returns)
- Risk scan: `risk-pattern-report.md`

**a097은 이 함수의 본문을 바꾸지 않는다.** 이 산출물이 존재하는 이유는 proposal R3이
이 함수의 분기 구조를 **근거로 삼기 때문**이다 — 분기를 주장하는 문서는 AST 열거를 먼저
만든다(`.claude/CLAUDE.md`).

## 이 Map이 결정하는 것

리뷰가 낸 P2 ②는 "`default` 분기가 `remindAfter <= 0`을 무시하고 재무장한다"였다.
AST가 그 관찰을 확인하면서 동시에 그것이 **버그가 아님**을 보인다:

`B4@256`(`remindAfter <= 0`)은 `B3@255`(`case AlertDelivered, AlertAcknowledged`)
**안에** 있고, `default`는 `B8@284`로 **형제**다. 형제 분기는 서로를 가드할 수 없다.
따라서 코드는 두 규칙을 표현하고 있다 — 인식된 settled 행의 *시한 재알림*과, 미지 상태의
*상태 복구*. 고칠 것은 동작이 아니라 그 둘을 한 문장으로 적은 문서다(design D3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED`, 또는 그 밖의 아무 값 | 원장 열 (CHECK 제약 없음) | 미지 값 → B8 fail-open |
| `deliveredAt`/`acknowledgedAt` | RFC3339 또는 NULL/공백 | 원장 열 | 둘 다 못 읽으면 B5 fail-open |
| `now` | 호출자의 clock | `j.clk.Now()` | 과거로 밀려도 B6 fail-open |
| `remindAfter` | 임의 `time.Duration` | 호출자 | `<= 0`이면 시한 재알림만 비활성 |

**불변식**: 순수 함수. 부작용도 오류도 없다. 반환은 `(owed, rearm)` 두 bool뿐이다.
`rearm == true`인 모든 경로에서 `owed == true`다 — 재무장은 보내기 위한 것이지 그 자체가
목적이 아니다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@251 | `switch state` | 없음 | — | 구조 분기 |
| B2@252 | `case AlertPending` | 없음 | `true, false` @254 | 기존 (a096) |
| B3@255 | `case AlertDelivered, AlertAcknowledged` | 없음 | — | 구조 분기 |
| B4@256 | `remindAfter <= 0` | 없음 | `false, false` @257 | **a097 2.5** (기록 전용 호출자) |
| B5@260 | 스탬프를 하나도 못 읽음 (`!ok`) | 없음 | `true, true` @263 — fail-open | 없음 — 원장 손상 주입 (`not-applicable`) |
| B6@266 | `elapsed < 0` (미래 스탬프) | 없음 | `true, true` @278 — fail-open | 기존 (a096b) |
| B7@280 | `elapsed < remindAfter` | 없음 | `false, false` @281 | 기존 (a096) |
| B8@284 | `default` (미지 상태) | 없음 | `true, true` @289 — fail-open | **a097 2.4** (`remindAfter=0`에서도 복구) |

정상 반환 하나 더: B3 안에서 창을 넘긴 경우 `true, true` @283.

**fail-open 세 곳(B5·B6·B8)이 이 함수의 성격이다.** 셋 다 "이 행을 신뢰할 수 없다"이고
셋 다 보내는 쪽으로 답한다. 억제하면 gate 래치도 승격도 돌지 않아 침묵이 길어진다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `latestStamp` | 두 스탬프 중 더 최근을 고른다 | 파싱 실패는 `ok=false`로만 표현 | AST calls |
| `now.Sub` | 경과 시간 | 오류 없음 | AST calls |

live config binding 없음.

## State mutations and fallbacks

- mutation 없음. 순수 함수다.
- fallback은 세 개의 fail-open 분기(B5·B6·B8)이며 전부 `(true, true)`다.

## Safety conclusion

- Safe edit boundary: **없음 — a097은 이 함수를 편집하지 않는다.** 문서(`ClaimAlertForDelivery`의
  doc comment)만 바꾼다.
- High-risk impact: yes (알림 전달 판단 전체)
- a097이 추가하는 테스트 2.4·2.5는 B8과 B4를 **고정**한다. 두 분기가 서로 다른 규칙임을
  주석이 아니라 테스트가 지키게 하는 것이 R3의 실질이다.
