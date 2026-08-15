# a109 뮤테이션 원장 — T2

> 규율(a108 승계): 각 뮤테이션은 **한 곳**만 죽이고, 그 뒤 **실제로 테스트를 돌려** 결과를
> 적는다. 원복은 `git checkout` 이 아니라 **역편집**이고, 원복 뒤 파일 크기·`func` 수·
> `if` 수를 세어 확인한다 — `git checkout` 은 커밋 전 GREEN 까지 지우고
> `git diff --quiet` 는 그것을 초록으로 통과시킨다(2026-08 실발생). 잡히지 않은
> 뮤테이션은 **생존**으로 사유와 함께 선언한다.

집행 도구: 스크립트가 (1) 앵커가 **정확히 1회** 나오는지 확인하고 (2) 치환 → 테스트 실행
→ 역치환 → 원복 검증을 자동으로 한다. 원복 검증은 세 수치(bytes·`func` 수·`if` 수)가
치환 전과 같고 파일 내용이 바이트 단위로 동일함을 확인한다. **28건 전부 `restored=true`.**

## §0 기준선 (뮤테이션 직전)

| 명령 | 결과 |
|---|---|
| `go test -race -count=1 ./cmd/tossctl/ ./internal/httpapi/ ./internal/console/` | ok 159.061s / ok 0.170s / ok 1.617s |
| `go vet ./cmd/... ./internal/httpapi/ ./internal/console/` | rc=0 |
| `gofmt -l ./cmd ./internal ./tools/logic-map` | 출력 없음 |

## §1 강등 경로 — 세 지점 각각 (D3)

| # | 뮤테이션 | 잡는 테스트 | 결과 |
|---|---|---|---|
| M1 | policy command 강등 → `return policyControlErr` (fatal 복원) | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | **CAUGHT** (rc=1) |
| M2 | policy runtime 강등 → fatal 복원 | 같음 | **CAUGHT** |
| M3 | alert control 강등 → fatal 복원 | 같음 | **CAUGHT** |
| M4 | 강등은 하되 **보고를 지운다**(`_ = policyControlErr`) | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestADegradedSiblingSaysWhichSurfaceIsMissing` | **CAUGHT** |
| M5 | policy command 강등이 **projection 좌표**를 쓴다(표면 오귀속) | `TestADegradedSiblingSaysWhichSurfaceIsMissing` | **CAUGHT** |
| M6 | 형제 이벤트 타입을 등급표에 **있는** 이름(`obs.EventAlertUndelivered`)으로 | `TestTheDegradationEventsAreNotOnTheCriticalRail` · `TestADegradedSiblingBootWritesNoUndeliveredOutboxRow` | **CAUGHT** |
| M7 | 강등 경로의 **nil 가드 제거**(`defer policyControl.Close()` 무조건) | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestSucceedingSiblingEndpointsAreStillServedAndClosed` | **생존** — 아래 §5 |

## §2 재부착 (D4)

| # | 뮤테이션 | 잡는 테스트 | 결과 |
|---|---|---|---|
| M8 | rate limit 제거(`tooSoon := false`) | `TestTheAttemptIsSingleFlightAndRateLimited`(구판) | CAUGHT — 단 §4 참조 |
| M8b | 같은 뮤테이션, **분리된** 테스트로 재측정 | `TestTheAttemptIsRateLimited` | **CAUGHT** |
| M9 | 운영 기본 간격 30s → 0 | `TestTheProductionRedialIntervalIsThirtySeconds` | **CAUGHT** |
| M10 | single-flight 제거(`a.trying` 검사 삭제) | `TestTheAttemptIsSingleFlightAndRateLimited`(구판) | **생존** → 테스트 결함 발견, §4 |
| M10b | 같은 뮤테이션, 분리·rate limit 을 끈 테스트로 재측정 | `TestTheAttemptIsSingleFlight` | **CAUGHT** |
| M11 | 실패한 시도가 현재 reader 를 **갈아끼운다** | `TestAFailedAttemptDoesNotClobberTheCurrentScreen` | **CAUGHT** |
| M12b | 시도가 성공해도 reader 를 갈아끼우지 않는다 | `TestTheDaemonAttachesWhenTheEngineComesUpLater` · `TestTheDaemonReattachesAfterTheEngineRestarts` | **CAUGHT** |
| M13 | 요청 goroutine 에서 **동기로** 시도한다(`go` 제거) | `TestTheRequestPathNeverWaitsForADial` | **CAUGHT** |
| M14 | 부재 상태에서 시도를 깨우지 않는다(`StrategyRuntimeConfigured` 의 wake 제거) | `TestTheDaemonAttachesWhenTheEngineComesUpLater` | **CAUGHT** |
| M15 | 전이가 아니라 **매번** 보고한다(`announce := true`) | `TestTheAttachmentReportsOnlyTransitions` | **CAUGHT** |

> M12 는 처음에 `a.reader` 대입을 통째로 지워 **컴파일 실패**로 끝났다 — 컴파일 실패는
> 뮤테이션 측정이 아니다. 컴파일되는 판(M12b, `_, a.failed = reader, false`)으로 다시 쟀다.

## §3 상태 신호 — nil 검사 세 곳의 대체 (P1-4 + issues.md T2-1)

| # | 뮤테이션 | 잡는 테스트 | 결과 |
|---|---|---|---|
| M16 | **집계 스냅샷**의 부재 판정을 `r.strategyRuntime != nil` 로 되돌린다 | `TestADialFailureRendersUnavailableRatherThanNotConfigured` 외 2 | **CAUGHT** |
| M21 | **REST 경로**(router.go)의 부재 판정을 `== nil` 로 되돌린다 | `TestTheRESTRouteStaysDormantForAnUnconfiguredWrapper` | **CAUGHT** |
| M22 | **SSE helper** 의 부재 판정을 `== nil` 로 되돌린다 | `TestTheStreamHelperRefusesAnUnconfiguredWrapper`(구판) | **생존** → 테스트 결함 발견, §4 |
| M22b | 같은 뮤테이션, fixture 를 고친 뒤 재측정 | 같은 테스트(수정판) | **CAUGHT** |
| M23 | 공유 판정(`StrategyRuntimeAbsent`)에서 **상태 신호 분기**를 지운다 | `TestAbsenceIsAskedAsAStateNotANil` 외 2 | **CAUGHT** |

## §4 문구·정적 핀

| # | 뮤테이션 | 잡는 테스트 | 결과 |
|---|---|---|---|
| M17b | alerts 문구를 옛 **단정**(「엔진이 없다」)으로 되돌린다 | `TestTheAlertsCLIDoesNotAssertTheEngineIsAbsent` · `TestTheCommandsRefuseWhenNoEngineIsRunning` | **CAUGHT** |
| M20 | 격리 해제 문구 **세 곳 중 한 곳만**(preview) 옛 단정으로 | `TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage` | **CAUGHT** — 값 단위 정정이 실제로 지켜진다 |
| M18 | `defer lock.Release()` 를 한 줄 아래로(여전히 endpoint 앞) | `TestTheJournalLockIsReleasedAfterEveryEndpointClose` | **생존 — 정상이다**: 순서 불변식이 여전히 참인 편집이므로 잡히면 안 된다(false positive 검사) |
| M18b | `defer lock.Release()` 를 **삭제**(모든 endpoint Close 뒤 = 해제 없음) | 같은 테스트 | **CAUGHT** |
| M19 | 금지 3종 주석의 `criticalEvents` 인용 **하나만** 지운다 | `TestTheDegradationCommentStillCitesTheForbiddenThree` | **생존 — 설계대로**: 이 핀은 「이유가 파일에 남아 있는가」를 재지 사본 수를 재지 않는다 |
| M19b | 그 인용을 **전부**(2곳) 지운다 | 같은 테스트 | **CAUGHT** |

## §5 뮤테이션이 **테스트의 결함**을 잡아냈다 (두 건)

이 원장의 실제 수확은 잡힌 26건이 아니라 **생존한 두 건**이다. 둘 다 「테스트가 다른
이유로 초록이었다」는 사례이고, 둘 다 테스트를 고친 뒤 다시 재서 CAUGHT 로 바꿨다.

### M10 — rate limit 이 single-flight 를 **가리고** 있었다

구판 `TestTheAttemptIsSingleFlightAndRateLimited` 는 창 안에서 20번 두드려 dial 이 1회임을
확인했다. 그런데 두 번째부터는 **rate limit** 이 막으므로 dial 은 어차피 1회다 —
`a.trying` 검사를 통째로 지워도 초록이었다. 두 성질을 한 테스트에서 재면 강한 쪽이 약한
쪽을 가린다.

**수정**: 테스트를 둘로 쪼갰다. `TestTheAttemptIsSingleFlight` 는 **간격을 0 으로 두어
rate limit 을 끄고**(겹침을 막는 것은 single-flight 뿐이다) 20개의 동시 요청을 건다.
`TestTheAttemptIsRateLimited` 는 시도가 **끝난 뒤** 다시 두드린다(그때 겹침은 없으므로
막는 것은 rate limit 뿐이다) + 창을 지나면 다시 시도함까지 잰다.

### M22 — 부재 fixture 가 **읽을 수 없는** 스냅샷을 들고 있었다

구판 `TestTheStreamHelperRefusesAnUnconfiguredWrapper` 의 부재 fixture 는 빈 스냅샷을
들었다. 판정을 `reader == nil` 로 되돌리면 부재 wrapper 는 non-nil 이라 통과하지만, 그
다음 줄의 `strategyprojection.Validate` 가 빈 스냅샷을 거절해 **어차피 오류**가 나온다 —
테스트는 「거절했다」로 초록이었다.

**수정**: 부재 fixture 도 **유효한** dormant 스냅샷을 들게 했다. 이제 오류를 낼 수 있는
것은 부재 판정 하나뿐이다.

## §6 생존 뮤테이션 선언 (측정 후 남긴 것)

### M7 — 강등 경로의 nil 가드 제거는 **테스트가 잡을 수 없다**

`defer policyControl.Close()` 를 무조건 등록해도 아무 테스트도 죽지 않는다. 이유는
측정 가능하다: 세 서버 타입의 `Close` 가 **모두 nil 수신자에 안전**하다
(`PositionPolicyCommandServer.Close`·`PositionPolicyRuntimeServer.Close`·
`AlertControlServer.Close` 각각 `if s == nil { return nil }`, internal/app/engine).
즉 nil 가드는 **동작을 바꾸지 않는다.**

그럼에도 가드를 두는 이유는 a108 이 B19 에서 세운 규율이다: 「강등 경로가 Close 를 부를
이유는 없고, 그 판단을 이 파일 안에 남긴다」. 이것은 **의도의 표현**이지 안전 장치가
아니므로, 잡히지 않는 것이 정상이다. 잡으려면 세 `Close` 의 nil 가드를 없애 패닉을
만들어야 하는데, 그것은 T1 표면을 **더 위험하게** 바꾸는 일이다 — 하지 않는다.

### M18·M19 — 「참인 편집」에 대한 false-positive 검사

둘은 생존이 **요구되는** 뮤테이션이다. M18 은 순서 불변식을 여전히 지키는 편집이고,
M19 는 이유가 파일에 남아 있는 편집이다. 각각의 **진짜** 위반(M18b·M19b)은 잡힌다.
정적 핀이 무해한 편집마다 우는 도구였다면 다음 사람이 그것을 지운다.

## §7 뮤테이션이 **닿지 못한** 표면 (정직한 공백)

- **`reportEngineEndpointDegraded` 의 B1 가드**(`ectx == nil || Notifier == nil`):
  이 경로를 타는 테스트가 없다. 조립이 실패한 배포에서만 나타나는 최소 표면이고,
  그 상태를 만들려면 `engine.Context` 를 반쪽으로 조립해야 한다 — 그 fixture 자체가
  프로덕션에 없는 모양이다. **미측정으로 선언한다.**
- **`runEngineRun` B14·B21**(in-process 조립 실패의 fatal 유지): 이 change 가 바꾸지
  않은 분기이고 기존 테스트도 없다. a109 밖의 공백으로 기록한다.
- **강등 goroutine 과 `ectx.Close()` 의 실제 경합**: a108 이 논증으로 닫았고(Normal
  이벤트는 journal handle 에 닿지 않는다) a109 도 같은 논증을 쓴다. 시간 경합을
  재현하는 테스트는 없다 — 논증이 근거이고, 그 논증의 전제(등급표 미등재)는 M6 이 잰다.
