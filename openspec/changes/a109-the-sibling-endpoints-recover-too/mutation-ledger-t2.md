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
| `go test -race -count=1 ./cmd/tossctl/ ./internal/httpapi/` | ok 159.061s / ok 0.170s (§2-fix 종료 시 172.199s / 1.624s) |
| `go test -race -count=1 -timeout 25m ./internal/console/` | §0a 참조 (§2-fix 종료 시 ok 693.315s) |
| `go vet ./cmd/... ./internal/httpapi/ ./internal/console/` | rc=0 |
| `$(go env GOROOT)/bin/gofmt -l ./cmd ./internal ./tools/logic-map` | 출력 없음 |

### §0a 정정 — 두 줄이 거짓이었다 (a109 §2-fix F9)

- **`gofmt` 는 PATH 에 없다.** 원래 적힌 `gofmt -l …` 은 이 환경에서 "command not
  found" 이고, 그 실패를 "출력 없음"으로 읽으면 **검사 0건이 위반 0건으로 보고된다**
  (없는 도구는 깨끗하다고 보고한다). 저장소의 진입점과 같은 절대경로
  `$(go env GOROOT)/bin/gofmt` 로 고쳤다 — `Makefile` 의 `GOFMT` 도 같은 식이다.
- **console 기준선 `ok 1.617s` 는 이 패키지의 수치가 아니다.** §2-fix 종료 시점 실측:

  | 명령 | 결과 |
  |---|---|
  | `go test -race -count=1 -timeout 25m ./internal/console/` | **ok 693.315s** (wall 11m39.363s) |
  | 같은 명령, **기본 timeout(600s)** | 그 전에 `panic: test timed out after 10m0s` — 693 > 600 이므로 완주할 수 없다 |
  | `go test -race -count=1 ./cmd/tossctl/` | ok 172.199s (wall 2m58.530s) |
  | `go test -race -count=1 ./internal/httpapi/` | ok 1.624s |

  1.6초로 적힌 줄은 다음 사람에게 "console 은 빠르다"고 말하고, 그 사람은 기본
  timeout 으로 돌린 뒤 **자기 변경이 무언가를 걸었다**고 읽는다. `internal/httpapi`
  의 1.624s 와 자릿수가 같은 것이 그 줄의 출처로 보인다 — 패키지를 헷갈린 값이다.

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

## §4a a109 §2-fix — A2 재현과 새 방어의 측정

집행 방식은 §0 과 같다(앵커 1회 확인 → 치환 → 실행 → 역치환 → 바이트 동일 확인).
아래 **전부 `restored=true`**.

### A2 가 보고한 것을 먼저 재현했다

| # | 뮤테이션 | 테스트 | 결과 |
|---|---|---|---|
| M24 | 금지 2번 줄(`engine.go:366`) **삭제** — 옛 단정(`"critical"`)으로 | `TestTheDegradationCommentStillCitesTheForbiddenThree` | **생존** = A2 P2-4 재현 (`"critical"` 은 `"criticalEvents"` 의 부분 문자열) |
| M25 | 순서 불변식의 flock 문장 삭제 — 옛 단정(`"flock"`)으로 | 같음 | **생존** (영문 헤더 `engine.go:14` 가 만족시킨다) |
| M32a | `defer lock.Release()` 삭제 | `TestTheJournalLockIsReleasedAfterEveryEndpointClose` | **CAUGHT** ("defer 가 없다") |
| M32b | **순서 역전** — `lock.Release` defer 를 endpoint Close 넷 **뒤로** 이동 | 같음 | **CAUGHT** (등록 #6 vs #2·#3·#4·#5, 네 줄 전부 보고) |

M32b 가 A2 가 요구한 "defer 순서 역전" 항목이다. 원장에 M18(한 줄 아래로 = 여전히 참인
편집, 생존이 정상)만 있고 **진짜 역전**이 없었다 — 그것이 이 핀의 유일한 사고 모양인데.

### 고친 뒤 다시 쟀다

| # | 뮤테이션 | 테스트 | 결과 |
|---|---|---|---|
| M24b | M24 와 같은 삭제, 고유 구절 단정으로 | `TestTheDegradationCommentStillCitesTheForbiddenThree` | **CAUGHT** |
| M25b | M25 와 같은 삭제, `"flock을 쥔 채로"` 로 | 같음 | **CAUGHT** |
| M26 | outbox 인용의 백틱 이름을 "그 함수"로 치환 | 같음 | **CAUGHT** |

### 새 방어 셋 (F1·F2·F3) — 각각 한 번씩 죽였다

| # | 뮤테이션 | 테스트 | 결과 |
|---|---|---|---|
| M29 | **F1 제거** — `requestCancelled` 를 `return false` 로 | `TestACancelledRequestDoesNotDetachAHealthyClient` | **CAUGHT** (시도 1회 + 탈착 보고 + live client 가 교체됨) |
| M30 | **F2 제거** — `observe` 의 `seat != a.seat` 검사 삭제 | `TestALateReadFailureDoesNotUnseatTheNewAttachment` | **CAUGHT** (탈착 보고 + 재-dial 2회) |
| M31 | **F3 제거** — publisher 의 `StrategyRuntimeAbsent` 한 줄 삭제 | `TestTheReattachWakeSurvivesABrokenAggregate` | **CAUGHT** (2초 안에 재부착 없음) |

### 두 문구·계약 항목

| # | 뮤테이션 | 테스트 | 결과 |
|---|---|---|---|
| M27b | alerts 부재 안내의 sentinel 을 날것의 오류(`opening the descriptor: %w`)로 | `TestTheCommandsRefuseWhenNoEngineIsRunning`(F6 수정판) | **CAUGHT** — a098 의 의도(운영자가 자기 경로를 의심하게 만들지 마라)가 **계속** 측정된다 |
| M28 | `AlertControlServer.Close` 의 nil 가드 제거 | `TestTheSiblingEndpointClosesAreSafeOnANilServer`(신규) | **CAUGHT** (nil 역참조 패닉) — M7 이 생존한 이유가 이제 계약이다 |

> M27 은 처음에 `if errors.Is(err, os.ErrNotExist)` 블록째 지워 **컴파일 실패**로
> 끝났다(`"os" imported and not used`). 컴파일 실패는 뮤테이션 측정이 아니다 —
> 컴파일되는 판(M27b)으로 다시 쟀다. M12 와 같은 실수이고 같은 처리다.

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

### M7 — 강등 경로의 nil 가드 제거는 **테스트가 잡을 수 없다** (§2-fix 로 절반 해소)

> **§2-fix F7 갱신**: 아래 결론(가드는 동작을 바꾸지 않으므로 잡히지 않는 것이 정상)은
> 그대로다. 다만 그 결론이 기대는 성질 — 세 `Close` 가 nil 수신자에 안전하다 — 은
> **T1 표면의 성질인데 아무도 계약으로 적지 않았다**. 그것을 소비자 쪽에서 재는 테스트를
> 넣었고(`TestTheSiblingEndpointClosesAreSafeOnANilServer`), M28 로 측정했다. 이제
> T1 이 Close 를 고치다 가드를 없애면 **강등 defer 가 패닉하기 전에** 그 테스트가 운다.


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

> **§2-fix F9 보강**: M18 의 짝으로 「이 핀이 실제로 막는 사고」 — `lock.Release` defer 를
> endpoint Close 넷 **뒤로** 옮기는 순서 역전 — 을 M32b 로 측정했다(CAUGHT). 생존이
> 요구되는 뮤테이션만 적어 두면 그 핀이 **무엇을 잡는지는** 원장에 없다.

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
