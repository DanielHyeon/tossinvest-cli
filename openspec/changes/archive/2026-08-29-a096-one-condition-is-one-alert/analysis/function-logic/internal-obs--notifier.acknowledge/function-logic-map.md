# Function Logic Map: `Notifier.Acknowledge`

- Source: `internal/obs/notifier.go` (416–454)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 10분기, 반환 6곳)
- Risk scan: `risk-pattern-report.md`
a096이 이 함수에 더한 것은 **`n.mu` 하나**다. 표시도 해제 조건도 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `operator` | 공백 아님 | 사람 | B1: 오류. 감사 흔적이 목적이므로 익명 해제는 없다 |
| `n.Journal` | nil 가능 | 조립 시점 | B2: gate만 풀고 반환 |
| `ids` | 비었으면 PENDING 전부 | `PendingAlerts` | B4: 조회 오류를 그대로 올린다 |
| `n.mu` | **이 함수가 잡는다** | a096 | 조회~해제 전체 동안 보유 |

불변식: **"남은 게 0일 때만 gate를 푼다"가 하나의 원자적 판정이어야 한다.**
이 함수는 읽고-판정하고-푼다. 그 사이에 PENDING 행이 새로 생기면, 세었을 때는 0이었지만
풀 때는 0이 아닌 상태에서 gate가 열린다.

a096 이전에는 그 "새로 생기는 PENDING"이 새 조건의 첫 알림뿐이었다. a096은 **재무장**이라는
두 번째 원천을 만든다 — 창을 넘긴 종결 행이 PENDING으로 돌아간다. 그래서 같은 잠금 안에
넣는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@417 | `operator`가 공백 | 없음 | `errors.New(…)` @418 | 기존 `TestAcknowledgeRequiresAnIdentity` |
| B2@420 | `n.Journal == nil` | — | — | 기존 |
| B3@421 | `n.Gate != nil` | `Gate.Clear` | `nil` @424 | 기존 |
| B4@430 | `len(ids) == 0` → 조회 | 없음 | 오류면 @433 | 기존 |
| B5@432 | `PendingAlerts` 오류 | 없음 | 그 오류 | 없음(미진입) |
| B6@435 | 각 id 순회 | `AcknowledgeAlert` | — | 기존 |
| B7@439 | ack 오류가 `ErrAlertNotFound`가 아님 | 없음 | 그 오류 @442 | 없음(미진입) |
| B8@440 | (같은 조건의 두 번째 항) | — | — | — |
| B9@447 | `UndeliveredCount` 오류 | 없음 | 그 오류 @448 | 없음(미진입) |
| B10@450 | `remaining == 0 && n.Gate != nil` | `Gate.Clear` | `nil` @453 | 기존 `TestAcknowledgeWhileStillPendingKeepsTheBlock` |

**B10이 잠금이 지키는 것이다.** `UndeliveredCount`(B9 위)와 `Gate.Clear`(B10) 사이에
재무장이 끼면, 미전달 critical 알림이 있는 채로 진입이 열린다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Journal.PendingAlerts` | 대상 목록 | B5 | AST :431 |
| `n.Journal.AcknowledgeAlert` | 상태 전이 + 감사 | `ErrAlertNotFound`는 흡수 | AST :436 |
| `n.Journal.UndeliveredCount` | 해제 조건 | B9 | AST :446 |
| `n.Gate.Clear` | 진입 차단 해제 | 없음 | AST :423,:451 |

`execgw.EntryGate`는 자기 뮤텍스만 쓰고 이 패키지로 되돌아오지 않는다. 따라서 `n.mu`를
잡은 채 `Gate.Clear`를 불러도 교착이 없다.

## State mutations and fallbacks

- PENDING 행 → ACKNOWLEDGED, `acknowledged_at`·`acknowledged_by` 기록.
- 재무장된 행도 PENDING이므로 이 함수의 대상이 된다. 운영자가 "리마인더 대기 중인 행"을
  확인한 것으로 표시되는데, 그것이 옳다 — 그 행이 말하는 조건은 여전히 참이고, 운영자는
  그것을 보고 있다.
- 되돌림 없음.

## Safety conclusion

- Safe edit boundary: 함수 진입부의 `n.mu.Lock()` 한 쌍. 표시 로직도 해제 조건도 그대로다.
- High-risk impact: **no** — 주문·손절·익절·사이징·체결 경로에 닿지 않는다.
- 비용: 전송이 진행 중이면 운영자의 해제가 최대 재시도 예산(3회×`RetryDelay`)만큼 기다린다.
  그것이 목적이다 — 전송을 시도하는 중에 gate를 푸는 것은 결과를 모르고 푸는 것이다.
