# Function Logic Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 뒤**, :270–342, 분기 11 · 반환 4 · 호출 20, source_sha256 `df8a4171e8dc…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전). 26라운드 수리 두 로트(`d8769cfb` · `b910173a`) 뒤 재추출(마지막 `b910173a`) — 이 함수 본문 · 분기 수 불변(같은 파일의 다른 함수 편집으로 줄 이동 · 파일 해시만 바뀜, 분기 좌표는 `ast.json` 이 정본).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): 인수 줄 이벤트 `EventAlertClaimHeld` → `EventAlertClaimStolen`(A#2). 분기 불변.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): B1~B11 → B1~B11(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:272) | — | — | (미실행) |
| B2 | `switch claim.Disposition` (:278) | — | — | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B3 | `case journal.ClaimAcquired:` (:279) | — | — | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B4 | `case journal.ClaimHeldElsewhere:` (:283) | — | — | (미실행) |
| B5 | `if d.reportHeld(alert.ID, claim.ExpiresAt)` (:292) | — | — | (미실행) |
| B6 | `default:` (:298) | — | — | (미실행) |
| B7 | `if claim.Stole` (:305) | — | — | (미실행) |
| B8 | `if d.Publisher == nil` (:316) | — | — | (미실행) |
| B9 | `} else` (:326) | — | — | (미실행) |
| B10 | `if perr != nil` (:333) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B11 | `if perr != nil` (:337) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ClaimAlertByID`, `Format`, `claim.ExpiresAt.UTC`, `d.Publisher.Publish`, `d.claimant`, `d.countRecordFailure`, `d.forgetHeld`, `d.forgetRecordRun`, `d.led`, `d.logf`, `d.recordDelivery`, `d.recordFailedAttempt`, `d.reportHeld`, `errors.New`, `obs.EventType`, `perr.Error` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: 낮음 — 관측 이름.
