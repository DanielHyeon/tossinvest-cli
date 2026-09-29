# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :471–626, 분기 25 · 반환 6 · 호출 34, source_sha256 `e790b278b3e6…`, 추출 커밋 `fbc6df5f`. 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ③ — `fbc6df5f`): `deliver` 는 이제 `n.mu` 를 쥐지 않은 채 돈다. 세 래치 자리가 원칙 E 로 바뀌었다 — 편집 전 B12(`:484` `Gate.Block`)와 B27(`:571` `Gate.Block`)은 사라지고 판정(`latchVerdict` — 세대 · 사유)을 반환하며, 편집 전 B18(`:520`)은 B17 에서 `BlockUnlessClearedSince` 로 조건부 차단한다. 세대는 새 함수 `readVerdict` 가 근거 확정 직후 읽는다(시험 훅 `evidence:` · `epoch:` 단계).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | claim 한 행과 임차 | `claimAndDeliver`(잠금 안 claim) | 토큰이 안 맞으면 정산 거절 |
| `n.mu` | **쥐지 않음** | — | 배제는 임차(A-4 귀속: 임차 무시 변이 L19 가 배제 시험 8개를 깸) |
| 해제 세대 | 근거 확정 직후 `readVerdict` 가 읽음 | `EntryGate.ClearEpoch` | 읽기 뒤 해제 → 차단 생략, 사이 해제 → 다시 차단 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:473) | 기본 시도 수 | — | (미실행) |
| B2 | for (:479) | 시도 루프 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |
| B3 | if (:480) | 발행기 없음 | — | (미실행) |
| B4 | if (:485) | 발행 성공 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |
| B5 | if (:487) | 정산 오류 없음 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |
| B6 | switch (:488) | 정산 결과 분기 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |
| B7 | case (:489) | 정산됨 | — | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092BothAnnouncersBuildTheSameEvent` |
| B8 | case (:491) | 선점(승인 · 남의 임차) — 래치 없음 | — | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` |
| B9 | case (:503) | 행 없음 → 미정산 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |
| B10 | case (:505) | 모르는 결과 → 미정산 | — | (미실행) |
| B11 | if (:531) | **unrecorded 판정**(편집 전 B11 로그 + B12 래치 → 로그만, 래치는 판정 반환) | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |
| B12 | if (:545) | 시도 기록 오류(편집 전 B13) | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B13 | else (:549) | 오류 없음 분기(편집 전 B14) | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B14 | if (:546) | 로그(편집 전 B15) | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B15 | if (:549) | 임차 상실 · 정산됨 · 행 없음(편집 전 B16) | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B16 | if (:559) | 전송 실패 로그(편집 전 B17) | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B17 | if (:569) | **vanished — 자리에서 원칙 E 조건부 차단**(편집 전 B18) | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B18 | if (:578) | 대기(편집 전 B19) | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B19 | if (:579) | 문맥 종료(편집 전 B20) | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B20 | switch (:596) | 반납 결과 분기(편집 전 B21) | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |
| B21 | case (:597) | 반납 오류(편집 전 B22) | — | (미실행) |
| B22 | if (:598) | 로그(편집 전 B23) | — | (미실행) |
| B23 | case (:601) | 반납 적용(편집 전 B24) | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |
| B24 | case (:604) | 반납 선점(편집 전 B25) | — | (미실행) |
| B25 | if (:620) | **exhausted 판정** 로그(편집 전 B26; 편집 전 B27 래치는 판정 반환으로) | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Publisher.Publish` | 원격 전송 — 잠금 밖 | 실패는 시도 기록 | AST |
| `MarkAlertDelivered` · `MarkAlertAttemptFailed` · `ReleaseAlertClaim` | 근거 확정 지점 | 결과 분류 | AST |
| `n.readVerdict` · `n.hook` | 세대 읽기 · 시험 단계 | 생산에서 훅 nil | AST |
| `n.Gate.BlockUnlessClearedSince`(B17) | vanished 조건부 차단 | — | AST |

## State mutations and fallbacks

- unrecorded: 임차 유지(폭풍 방지), 판정 반환. exhausted: 반납 뒤 판정 반환. vanished: 자리에서 조건부 차단, lost 반환(승격 없음).
- 반환 (sent, lost) 의 의미는 편집 전과 같다. 셋째 반환(판정)만 늘었다.

## Safety conclusion

- Safe edit boundary: 시도 루프 · 정산 분류 · 임차 유지/반납 불변. 래치의 적용 조건만 원칙 E.
- High-risk impact: yes — 진입 차단 경로. 청산 무관.
