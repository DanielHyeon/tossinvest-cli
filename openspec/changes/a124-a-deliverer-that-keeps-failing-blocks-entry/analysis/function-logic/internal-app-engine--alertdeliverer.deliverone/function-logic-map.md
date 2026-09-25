# Function Logic Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 전**(base 논리), :167–242, 분기 12 · 반환 6 · 호출 22.
  source_sha256 `89a491c9c259…`, 추출 HEAD `463cc895` (2026-09-25).
- Risk scan: `risk-pattern-report.md`

**이 번들은 proposal 이 이 함수의 분기를 근거로 쓰기 때문에 문서보다 먼저 만들었다.** a124 는 이 함수를
편집한다(한도 판정) — 편집 뒤 `revision: current` 로 재추출한다(tasks 1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `alert` | `PENDING` 행 (나열 시점) | `cycle` :150 | 나열 뒤 정착되면 B6 |
| `d.Journal.ClaimAlertByID` | 임차 3 종 결과 | `internal/journal` | 오류 → B1 |
| `d.Publisher` | nil 허용 | 배선 | B8 |
| `claim.Token` | 임차 토큰 — 원장 밖으로 안 나간다 | 임차 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 임차 요청 오류 (:169) | `logf(EventAlertUndelivered)` :170 | return (:171) | a098 (임차 실패) |
| B2 | `switch claim.Disposition` (:173) | — | — | — |
| B3 | `ClaimAcquired` (:174) | `forgetHeld` :177 | 계속 | `a098_the_outbox_gets_emptied_test.go` |
| B4 | `ClaimHeldElsewhere` (:178) | B5 참이면 `logf(EventAlertClaimHeld)` :188 — 임차당 한 번 | return (:192) | `a098_two_senders_one_row_test.go` |
| B5 | `d.reportHeld(alert.ID, claim.ExpiresAt)` (:187) | `heldReported` 기록 | — | 같음 (R21) |
| B6 | default — 나열과 임차 사이에 정착됨 (:193) | `forgetHeld` :195 | return (:196) | a098 |
| B7 | `claim.Stole` (:198) | `logf(EventAlertClaimHeld, "expired lease taken over")` :201 | 계속 | `a098_two_senders_one_row_test.go` |
| B8 | `d.Publisher == nil` (:205) | `logf(EventAlertUndelivered)` :209 · `release` :211 (임차 반납, 행은 `PENDING`) | return (:212) | a098 (무설정 publisher) |
| B9 | `Publish` 실패 `perr != nil` (:221) | `MarkAlertAttemptFailed` :222 (`attempts+1`, `last_error`; `outbox.go:471`) · `release` :225 | return (:226) | `a098_the_backlog_does_not_delay_protection_test.go` |
| B10 | `MarkAlertAttemptFailed` 자체 오류 (:222) | `logf(EventAlertUndelivered, "recording a failed attempt failed")` :223 | B9 로 합류 → :226 | 1.4 에서 확인 |
| B11 | `MarkAlertDelivered` 오류 (:230) | `logf` :233, **임차 유지** | return (:234) | a098 (정착 실패) |
| B12 | `settled.Outcome != SettleApplied` (:236) | `logf("settled by somebody else")` :239, 임차 유지 | 종단 | `a098_the_operator_reads_and_acknowledges_test.go` |
| 종단 | 발행·정착 성공 | — | (:242) | `a098_the_outbox_gets_emptied_test.go` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.Journal.ClaimAlertByID` :168 | 이 행의 발송 권한 | 오류 → B1; `Held`/settled 는 결과값 | AST |
| `d.Publisher.Publish` :215 | 원격 전송 1회 | ctx 기한 — 재시도는 **다음 사이클** | AST |
| `d.Journal.MarkAlertAttemptFailed` :222 | 시도 수·마지막 오류 기록 | 오류 → B10 | AST + `outbox.go:471` |
| `d.release` :211 · :225 | 임차 반납 (detached ctx, `alertReleaseTimeout` 5s) | 오류는 로그 | AST |
| `d.Journal.MarkAlertDelivered` :229 | 정착 | 오류 → B11 | AST |

## State mutations and fallbacks

- **B8 의 결과는 로그 한 줄과 임차 반납뿐이고(`attempts` 불변), B9 의 결과는 `attempts+1` · 반납뿐이다.** (2026-09-26 교정 — freeze 7회차 X5) `execgw.EntryGate` 도 `EscalateOperatingMode` 도
  이 함수에 없다 — `alertdelivery.go` 전체에서 `Gate`·`Escalate` 는 0 회(grep, HEAD `463cc895`).
  동기 경로 `Notifier.deliver` 의 같은 실패는 세 자리에서 게이트를 잠근다(`notifier.go:484·:520·:571`).
  이 비대칭이 a124 R1 의 근거다.
- 행은 실패해도 `PENDING` 으로 남는다(내구성 — 정본 「critical 알림 전달 실패 지속」의 보존 요구). a124 는
  이것을 바꾸지 않는다.
- 이 함수는 `attempts` 를 **읽지 않는다**(나열 행의 `attempts` 는 `alert.Attempts` 로 와 있으나 어느 분기도
  보지 않는다). 한도 판정은 새 분기다.
