# a114 뮤테이션 원장

측정 2026-09-25, 비root, `analysis/harness/run.sh`(저장소가 아니라 `go.mod`·`cmd`·`internal` 사본에서 변이 —
병행 세션 보호, 시험 목록 `tests.txt`). 무변이 대조군 `none` rc=0 을 먼저 확인했다.

## a109 T2 원장의 콘솔 적용판

| id | 변이 | 결과 | 죽인 테스트 |
| --- | --- | --- | --- |
| M13 | 요청 goroutine 에서 동기로 시도(`go` 제거) | **사망** | `TestTheLifecycleRequestPathNeverWaitsForADial` |
| M10b | single-flight 제거 | **사망** | `TestTheLifecycleAttemptIsSingleFlight` |
| M8b | rate limit 제거 | **사망** | `TestTheLifecycleAttemptIsRateLimited` |
| M9 | 운영 간격 30s → 0 | **사망** | `TestTheProductionLifecycleRedialIntervalIsThirtySeconds` |
| M11 | 실패한 시도가 **붙어 있던** client 를 비운다 | **사망** | `TestAFailedLifecycleAttemptKeepsTheCurrentClient` · `…ReportsOnlyTransitions` |
| M12b | 성공한 시도가 자리를 갈아끼우지 않는다 | **사망** | `TestTheConsoleAttachesWhenTheEngineStartsLater` · `…ReattachesAfterTheEngineRestarts` |
| M14 | 비어 있는 자리에서 시도를 깨우지 않는다 | **사망** | `…AttachesWhenTheEngineStartsLater` · `…RequestPathNeverWaitsForADial` |
| M15 | 전이가 아니라 매번 보고 | **사망** | `TestTheLifecycleAttachmentReportsOnlyTransitions` |
| M29 | 취소 판정 제거 | **사망** | `TestACancelledLifecycleCallIsNotADetachment` |
| M30 | 자리 세대 비교 제거 | **사망** | `TestALateLifecycleFailureDoesNotUnseatTheNewAttachment` |
| M35 | 성공한 호출이 `attached` 를 복원하지 않음(a109 G3) | **사망** | `TestTheLifecycleAttachmentReportsOnlyTransitions`(회복 뒤 두 번째 사망도 한 줄) |
| M36 | 밀려난 client 를 닫지 않음 | **사망** | `TestALateLifecycleFailureDoesNotUnseatTheNewAttachment`(가짜 io.Closer) — 운영 `positionpolicyrpc.Client` 는 Close 가 없어 no-op(issues S2) |

not-applicable(침묵한 생략 아님): **M31**(publisher wake) — 콘솔에는 publisher 가 없다; 같은 역할을 하는 펌프는
C5 가 잰다. **M33**(빈 자리 sentinel) — lifecycle 에는 unavailable sentinel 이 없다(비어 있음 = detached 오류).
**M34**(무조건 wake) — 집계가 전략 읽기를 가리는 경로(httpapi B1–B7)가 콘솔 lifecycle 에는 없다: 모든 화면
호출이 wrapper 메서드를 거쳐 observe 한다. **M37·M38**(transport keep-alive·Close 본문) — `positionpolicyrpc`
는 파일 표면 밖.

## 콘솔 고유 (freeze 리뷰 반영)

| id | 변이 | 결과 | 죽인 테스트 |
| --- | --- | --- | --- |
| C1 | 모든 오류를 탈착으로(답 분류 제거) | **사망** | `TestAnEngineAnswerIsNotADetachment` · `TestAnInternalEngineErrorDoesNotFlapTheSeat` |
| C2 | 토큰 거절도 「답」으로 | **사망** | `TestATokenRejectionIsADetachment` |
| C3 | 부착 전 격리 호출을 `ErrUnwired` 로 감싼다 | **사망** | `TestTheQuarantineSurfaceRidesTheAttachment` |
| C4 | 부팅 해석이 `lastTry` 를 찍는다 | **사망** | `TestTheBootResolutionLeavesTheFirstWakeFree` |
| C5 | 펌프 제거 | **사망** | `TestTheConsoleAttachesWithoutAnyRender` |
| C6 | 실패한 Apply 를 한 번 더 보낸다 | **사망** | `TestTheLifecycleWrapperNeverResendsACommand` |
| C7 | runConsole 안에 부팅 1회 `positionpolicyrpc.Dial` 재도입 | **사망** | `TestRunConsoleNeverDialsTheLifecycleItself` |
| C8 | 격리 해제 전달 제거(항상 ErrUnwired) | **사망** | `…NeverResendsACommand` · `TestTheQuarantineSurfaceRidesTheAttachment` |

1판 합계 **20/20 사망**(648df8ef). 이 수는 내가 고른 변이 집합에 대한 진술이었다 — 구현 후 독립 리뷰가 새 변이
일곱을 냈고 그중 **둘이 생존**했다(아래 N2·N7).

## 구현 후 리뷰의 변이 (2판)

| id | 변이 | 1판 | 2판 | 죽인 테스트(2판) |
| --- | --- | --- | --- | --- |
| P2 (리뷰 N2) | observe 의 성공·답 경로가 `failed` 를 지우지 않음 — 같은 client 로 회복한 자리가 간격마다 재-dial·교체(로그 없이) | **생존** | **사망** | `TestARecoveredSeatStopsAsking` |
| P7 (리뷰 N7) | position policy decoder 의 remote failure 문구를 답에서 뺌 | **생존** | **사망** | `TestAnInternalListErrorDoesNotFlapTheSeat` |
| P21 | 사유 없는 remote failure 도 답으로(리뷰 P2-1) | — | **사망** | `TestAReasonlessRemoteFailureIsNotOurEngine` |
| 리뷰 N1·N3·N5·N6 | 격리 문구 제거·실패 경로 wake 제거·seat 증가 제거·토큰 검사 제거 | 사망 | — | 리뷰어 실측 |
| 리뷰 N4 | wake 가 `ctx.Err()` 를 무시 | 생존 | 생존(수용) | 종료 뒤 시도는 즉시 실패한다 — 등가에 가까운 무해 변이 |

2판 재측정(하네스 판정을 rc 가 아니라 FAIL/ok 줄로 바꾼 뒤 — 리뷰 P2-3): 대조군 green, P2·P7·P21·C1·C2 CAUGHT.
펌프의 절반 틱(리뷰 P2-2)은 시험으로 고정하지 않았다 — 지터 의존이라 결정적 시험이 없다(원장에 명기).
