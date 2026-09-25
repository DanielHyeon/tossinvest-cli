# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :176–231, 분기 4 · 반환 3 · 호출 11.
  source_sha256 `0bc75668ff17…`, 추출 HEAD `463cc895` (2026-09-25).
- Risk scan: `risk-pattern-report.md`

**a124 는 이 함수를 편집하지 않는다**(편집 주체는 a092). 이 번들은 proposal 이 「운영 모드 승격은 이 함수의
B4 에서만 일어난다」를 근거로 쓰기 때문에 만들었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 배선 | B1 — 내구 기록 없이 best-effort |
| `e` | critical 이벤트 | `Notify` 의 등급 분기 | — |
| `claimAndDeliver` 결과 `(sent, owed, err)` | — | :200 | B3 · B4 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n.Journal == nil` (:177) | B2 `Log.Warn` :182 · `publishBestEffort` :186 | `nil` (:187) | `TestAFailedClaimWithNothingWiredStillReports` |
| B2 | `n.Log != nil` (:181) | 경고 로그 | — | 같음 |
| B3 | `claimAndDeliver` 오류 (:201) — 기록 실패 | **`n.escalate` :219** (게이트는 `claimAndDeliver` :280 이 이미 잠갔다) | wrapped error (:220) | `TestAClaimThatFailsAttemptsTheDurableBlock` |
| B4 | `owed && !sent` (:223) — 시도했으나 못 보냄 | **`n.escalate` :228** (`n.mu` 밖 — 재진입 교착 회피) | — | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` 계열 |
| 종단 | 보냈거나 빚지지 않음 | — | `nil` (:230) | `TestOneConditionIsOneSend` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.claimAndDeliver(ctx, record, e)` :200 | 기록·임차·**동기 전송** | `n.mu` 아래에서 `deliver` 까지 | AST + a092 D0.3d |
| `n.escalate(ctx, e)` :219 · :228 | `EscalateOperatingMode(…, ModeTriggerCriticalAlertUndelivered)` (`notifier.go:383`) | 실패는 error 로그 | AST + grep |
| `n.publishBestEffort` :186 | 원장 없는 구성의 전송 | best-effort | AST |

## State mutations and fallbacks

- `ModeTriggerCriticalAlertUndelivered` 로 승격하는 비테스트 호출은 `escalate` :383 하나이고, `escalate` 의
  호출자는 이 함수의 :219·:228 둘뿐이다(grep, HEAD `463cc895`). 즉 **「전달 실패 지속 → 모드 승격」은
  오늘 동기 경로에서만 일어난다.**
- a092 가 `claimAndDeliver` 에서 `deliver` 를 빼면 B4 의 `owed && !sent` 는 「시도했으나 못 보냄」이 아니라
  「시도하지 않음」이 되고, 승격은 B3(기록 실패)에서만 남는다. 전송 실패의 승격 주체가 사라지는 자리가 여기다.
