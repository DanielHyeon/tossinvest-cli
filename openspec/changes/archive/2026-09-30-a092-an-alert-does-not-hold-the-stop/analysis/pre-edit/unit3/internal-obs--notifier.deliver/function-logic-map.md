# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :420–574, 분기 27 · 반환 6 · 호출 30, source_sha256 `0bc75668ff17…`, 추출 HEAD `b3f14925`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(a092 21.2(a) · 21.4 · 22.3 C2/C27 · 23.3 K2/K4 · 24.3 M5): 착지 단위 ③ — `claimAndDeliver` 가 `deliver` 를 **배제 잠금 밖**에서 부르게 되므로,
  이 함수의 세 래치 자리(`:484` · `:520` · `:571`)가 운영자 승인(`Acknowledge`)의 「셈 ~ 해제」와 겹칠 수 있다. 세 자리를 원칙 E(근거 확정 순간 해제 세대 읽기 →
  `BlockUnlessClearedSince`)로 바꾸고, 승격이 포함된 두 자리(`:484` · `:571`)는 판정을 호출자에게 돌려 호출자가 「조건부 차단 → 승격 → 승격 실패면 무조건 차단」을 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | 이 발송자가 claim 한 행과 그 임차 | `claimAndDeliver` | 임차가 안 맞으면 정산이 거절(B8 · B16 · B25) |
| `n.mu` | **편집 전 전제: 호출자가 쥠**(주석 :405) — 편집 뒤 **쥐지 않음** | `claimAndDeliver` | 행 배제는 원장 임차가 진다(정본 「발송 권한은 원장이 준다」) |
| `n.Gate` 해제 세대 | `ClearEpoch(ReasonAlertUndelivered)` | `EntryGate`(a124) | 편집 뒤 근거 확정 직후 읽음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if `attempts <= 0` (:422) | 기본 시도 수 | — | (미실행) |
| B2 | for 시도 루프 (:428) | — | — | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` |
| B3 | if `n.Publisher == nil` (:429) | `lastErr` 설정 후 루프 탈출 → 소진 경로 | — | (미실행) |
| B4 | if 발행 성공(`err == nil`) (:434) | `MarkAlertDelivered` :435 | — | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` |
| B5 | if `markErr == nil` (:436) | 결과 분류 | — | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` |
| B6 | switch `settled.Outcome` (:437) | — | — | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` |
| B7 | case `SettleApplied` (:438) | **반환** `true,false` :439 | — | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` |
| B8 | case `LeaseLost · AlreadySettled` (:440) | `logLeaseLost` → **반환** `true,true` :451 — 선점(래치 없음) | — | (미실행) |
| B9 | case `SettleNotFound` (:452) | `markErr` 설정 | — | (미실행) |
| B10 | case 모르는 결과 (:454) | `markErr` 설정(fail-safe) | — | (미실행) |
| B11 | if `n.Log != nil` (:478) | 오류 로그 | — | `TestAPublishedButUnsettledRowKeepsItsLease`, `TestASendThatCannotBeRecordedLatchesTheGate` |
| B12 | if `n.Gate != nil` (:483) | **래치 `:484`** `Gate.Block` — 발행됐으나 정산 불가(임차 유지) → **반환** `false,false` :491 → 호출자 승격 | — | `TestAPublishedButUnsettledRowKeepsItsLease`, `TestASendThatCannotBeRecordedLatchesTheGate` |
| B13 | if `MarkAlertAttemptFailed` 오류 (:495) | — | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B14 | else 오류 없음 분기 (:499) | — | — | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` |
| B15 | if `n.Log != nil` (:496) | 오류 로그 | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B16 | if `failed.Outcome != SettleApplied` (:499) | 임차 상실 · 정산됨 · 행 없음 | — | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` |
| B17 | if `n.Log != nil` (:509) | 전송 실패 로그 | — | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` |
| B18 | if `failed.Outcome == SettleNotFound && n.Gate != nil` (:519) | **래치 `:520`** — 행 사라짐 → **반환** `false,true` :523(lost — 승격 없음) | — | `TestARowThatVanishedIsNotReportedAsContention` |
| B19 | if `attempt < attempts` (:525) | 대기 | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B20 | if `!n.wait(ctx)` (:526) | 문맥 종료 → 탈출 | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B21 | switch 반납 결과 (:543) | `ReleaseAlertClaim` :541(분리 문맥) | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B22 | case `relErr != nil` (:544) | 로그 후 래치로 | — | (미실행) |
| B23 | if `n.Log != nil` (:545) | — | — | (미실행) |
| B24 | case `SettleApplied` (:548) | 래치로 | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B25 | case 그 밖(선점) (:551) | `logLeaseLost` → **반환** `false,true` :561 | — | (미실행) |
| B26 | if `n.Log != nil` (:565) | 소진 로그 | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B27 | if `n.Gate != nil` (:570) | **래치 `:571`** 소진 → **반환** `false,false` :573 → 호출자 승격 | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Publisher.Publish` :433 | 원격 전송 — **편집 뒤 잠금 밖** | 실패는 시도 기록 | AST |
| `n.Journal.MarkAlertDelivered` :435 · `MarkAlertAttemptFailed` :494 · `ReleaseAlertClaim` :541 | 임차 토큰으로 정산 · 시도 기록 · 반납 = **근거 확정 지점** | 결과 분류는 위 표 | AST |
| `n.Gate.Block` :484 · :520 · :571 | 전달 실패 사유 래치 | 편집 뒤 `BlockUnlessClearedSince` / 판정 반환 | AST · a124 `alertdelivery.go` `judge` |

## Acknowledge 와의 겹침 표 (21.2(a) — 편집 뒤, `deliver` 가 잠금 밖일 때)

`Notifier.Acknowledge`(notifier.go :800 부근)는 `n.mu` 아래에서: (대상 없으면) PENDING 나열 → 각 행 `AcknowledgeAlert`(**임차 무시**) → `UndeliveredCount` → 0 이면 `Clear`(세대 +1).

| 승인이 끼는 순간 | `deliver` 가 보는 것 | 결과 |
|---|---|---|
| 발행 전 · 발행 중 | 정산 `MarkAlertDelivered` → `AlreadySettled`(B8) | 선점 — 래치 없음 · 승격 없음. 발행이 이미 나갔으면 사람은 같은 내용을 한 번 더 받음(무해) |
| 발행 실패 뒤 · 시도 기록 전 | `MarkAlertAttemptFailed` → `AlreadySettled`(B16, B18 아님) | 선점 — 래치 없음 |
| 마지막 시도 뒤 · 반납 전 | `ReleaseAlertClaim` → 그 밖(B25) | 선점 — 래치 없음 |
| **근거 확정 뒤 · 세대 읽기 전**(`:484` 정산 결과 · `:520` 시도 기록 결과 · `:571` 반납 결과가 돌아온 뒤) | 해제 세대가 이미 올라 있음 | 차단이 선다(「앞」으로 봄 — 정본 허용, 보수 방향). 승인이 행을 이미 정리했으면 다음 승인이 셈 0 으로 푼다 |
| **세대 읽기 뒤 · 적용 전** | `BlockUnlessClearedSince` 가 세대 불일치로 거절 | 차단 없음. `:484` · `:571` 은 승격은 적용(제때 된 승격은 해제로 풀리지 않음); 승격 쓰기 실패면 무조건 차단(K2) |
| 셈에 이 행이 든 채(PENDING · 임차 보유) 승인이 id 로 이 행을 빠뜨림 | 셈 ≥ 1 | 해제 없음 — 겹침 아님 |

- 이 표가 **주장하지 않는 것**: 승인의 대상 집합이 정해진 **뒤** 들어와 해제 전에 다른 발송자가 전달한 행(a124 D10 (i) 경계) — 이 change 의 범위 밖.

## State mutations and fallbacks

- `:484` 는 임차를 **유지**(폭풍 방지 — 주석 :486-490), `:571` 은 반납 뒤 래치, `:520` 은 행이 사라짐.
- 편집 뒤: 세 자리 모두 근거 확정 직후 `ClearEpoch` 를 읽는다. `:520` 은 자리에서 `BlockUnlessClearedSince`. `:484` · `:571` 은 (세대 · 사유) 판정을 반환하고 호출자가 적용.

## Safety conclusion

- Safe edit boundary: 시도 루프 · 정산 분류 · 임차 유지/반납 · 반환의 (sent, lost) 의미는 그대로. 바뀌는 것은 래치를 **언제 · 어떤 조건으로** 적용하는가뿐.
- High-risk impact: yes — 전달 실패 사유(진입 차단) 경로. 청산 경로 무관(게이트는 진입만 막음).
