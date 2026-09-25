# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :420–574, 분기 27 · 반환 6 · 호출 30.
  source_sha256 `0bc75668ff17…`, 추출 HEAD `463cc895` (2026-09-25).
- Risk scan: `risk-pattern-report.md`

**a124 는 이 함수를 편집하지 않는다.** 이 번들은 proposal 이 이 함수의 분기를 **대조 근거**로 쓰기 때문에
만들었다 — 오늘 전달 실패로 게이트를 잠그는 자리가 여기(와 `claimAndDeliver` :280) 뿐이고, a092 가 이 경로를
엔진 루프에서 뺀다. 편집 주체는 a092 다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Attempts` | ≤0 이면 `DefaultCriticalAttempts` | B1 :422 | — |
| `n.Publisher` | nil 허용 | 배선 | B3 — 시도 없이 `break` |
| `id`, `token` | `claimAndDeliver` 가 얻은 임차 | 호출자 | 정착 결과로 판정 |
| `n.Gate` | nil 허용 | 배선 | nil 이면 잠금 자리 4 곳이 모두 조용히 통과 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `attempts <= 0` (:422) | 기본값 | — | 기본값 시험 |
| B2 | `for attempt := 1..attempts` (:428) | — | — | — |
| B3 | `Publisher == nil` (:429) | `lastErr` 설정, `break` | → 루프 뒤 | `TestAFailedClaimWithNothingWiredStillReports` 계열 |
| B4 | `Publish` 성공 (:434) | `MarkAlertDelivered` :435 | — | — |
| B5·B6 | 정착 성공, `switch settled.Outcome` (:436-437) | — | — | — |
| B7 | `SettleApplied` (:438) | — | `true,false` (:439) | `TestOneConditionIsOneSend` |
| B8 | `LeaseLost`/`AlreadySettled` (:440) | `logLeaseLost` :450 | `true,true` (:451) | `TestAcknowledgeCannotClearTheGateMidSend` |
| B9 | `SettleNotFound` (:452) | `markErr` 설정 → :475 | — | 1.4 |
| B10 | 미지 outcome (:454) | `markErr` 설정 (fail-safe) | — | a099 회귀 핀 |
| B11·B12 | 발행됐으나 기록 못 함 (:478·:483) | `Log.Error` · **`Gate.Block(ReasonAlertUndelivered)` :484** · 임차 유지 | `false,false` (:491) | `TestASendThatCannotBeRecordedLatchesTheGate` |
| B13·B15 | `MarkAlertAttemptFailed` 오류 (:495-496) | `Log.Error` | 계속 (다음 시도) | 1.4 |
| B14·B16 | 실패 기록은 됐는데 `Outcome != Applied` (:499) | B17 로그 · `logLeaseLost` :514 · B18 `NotFound && Gate` → **`Gate.Block` :520** | `false,true` (:523) | a099 회귀 핀 |
| B19·B20 | 남은 시도 있음 → `wait` (:525-526) | 대기 (`n.wait`) | `!wait` 면 `break` | `TestAnUndeliveredConditionIsStillRetried` |
| B21 | 루프 뒤 `ReleaseAlertClaim` 결과 `switch` (:543) | — | — | — |
| B22·B23 | 반납 오류 (:544-545) | `Log.Error` | 계속 → 잠금 | 1.4 |
| B24 | 반납 `Applied` (:548) | — | 계속 → 잠금 | — |
| B25 | 그 밖 — 행이 옮겨 감 (:551) | `logLeaseLost` :560 | `false,true` (:561) | a099 |
| B26·B27 | 시도 소진 (:565·:570) | `Log.Error` · **`Gate.Block(ReasonAlertUndelivered)` :571** | `false,false` (:573) | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Publisher.Publish` :433 | 원격 전송 | 시도당 `DefaultPublishTimeout`, 시도 간 `n.wait` | AST |
| `n.Journal.MarkAlertDelivered` :435 | 정착 | 오류·비 Applied → :475 잠금 | AST |
| `n.Journal.MarkAlertAttemptFailed` :494 | 실패 기록 | 오류 → B13 | AST |
| `n.Journal.ReleaseAlertClaim` :541 | 소진 뒤 반납 | `releaseCtx` 기한 | AST |
| `n.Gate.Block` :484·:520·:571 | 진입 차단 | 메모리 래치 — 재시작은 `restoreAlertEntryLatch` 가 복원 | AST + grep |

## State mutations and fallbacks

- 게이트를 잠그는 자리는 이 함수에 **셋**(:484·:520·:571), `claimAndDeliver` 에 하나(:280)다 — `notifier.go` 의
  비테스트 `Gate.Block(ReasonAlertUndelivered)` 는 이 넷이 전부다(grep, HEAD `463cc895`). 운영 모드 승격은
  여기 없고 호출자 `notifyCritical` B4 에 있다(`owed && !sent`).
- 이 함수는 `claimAndDeliver` 가 `n.mu` 를 쥔 채 부른다(a092 D0.3d 2번). a092 가 이 호출을 엔진 루프 밖으로 빼면
  위 세 잠금 자리는 **루프 경로에서 도달 불가**가 된다 — 그 뒤에 지속 실패로 잠그는 자리는
  `restoreAlertEntryLatch` (기동 시 1회) 뿐이다. a124 는 그 빈자리를 배달 실행자에 만든다.
