# a064 · Review

## Pre-Edit Gate (편집 전 선언)

Base commit: `0c563c6cca3f636a680fc1b76ae440f9eee7706d`

편집할 **기존** 함수는 여섯 개이고, 여섯 개 모두 편집 전에 Function Logic Map과
Branch Test Map을 만들었다.

| # | 함수 | 파일 | 분기 | 위험 | 편집 범위 |
|---|---|---|---|---|---|
| 1 | `ExitObserver.Run` | `internal/app/engine/exitloop.go` | 3 | High | 사이클 결과를 읽어 실패를 로그 |
| 2 | `ExitObserver.workingSet` | `internal/app/engine/exitloop.go` | 20 | High | 격리 생성 2경로에 이벤트 발행 |
| 3 | `ExitObserver.record` | `internal/app/engine/exitloop.go` | 11 | High | 격리 sentinel 인지 + 되읽기 발행 |
| 4 | `mergeEngine` | `internal/config/engine.go` | 5 | Normal | 알림 블록 병합 분기 1개 |
| 5 | `recordGateSettings` | `internal/app/engine/interlock.go` | 3 | High | audit 항목 4개 append |
| 6 | `NewContext` | `internal/app/engine/engine.go` | 13 | High | 인자 2개 교체 |

**선언 1 — 판정 로직을 바꾸지 않는다.** `exitpolicy` 패키지, `SelectRecoverySnapshot`,
`ValidateRecoveryDerivation`, 손절·익절·사이징 수식에 한 글자도 손대지 않는다. a062가
고친 첫 rung 격리 경로도 그대로 둔다.

**선언 2 — 원장 write 함수를 바꾸지 않는다.** `QuarantineExitSnapshot`,
`quarantineExitSnapshotTx`, `RecordExitJudgementResult`, `ReleaseExitSnapshotQuarantine`
는 편집 대상이 아니다. 격리 생성 사실은 이미 커밋된 행을 **읽어서** 얻는다 (design D3).

**선언 3 — 어떤 포지션도 판정 대상 집합에서 빼거나 넣지 않는다.** `workingSet`의
반환값은 편집 전후로 동일해야 하고, 그것을 테스트로 고정한다
(`TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet`).

**선언 4 — 알림을 켜지 않는다.** 기본값은 off이고, off는 오늘과 구별 불가능하다.
전송을 실제로 켜는 것은 §0.7 사람 판정이며 에이전트가 하지 않는다.

**선언 5 — 시크릿을 기록하지 않는다.** 토큰은 설정 파일에 자리를 만들지 않고 환경에서만
읽는다. 토큰과 topic 값은 audit·구조화 로그·알림 본문 어디에도 남기지 않는다.

**선언 6 — §0.3을 지킨다.** 새 알림 호출은 전부 (a) 격리되어 이번 사이클에 발의하지
않는 포지션에 대해서만, (b) 그 사실이 확정된 뒤에만 실행된다. `workingSet`이 격리 행을
`append(out, refused...)`로 마지막에 두는 기존 순서를 유지한다 — publisher가 실제로
배선되면 그 순서가 처음으로 의미를 갖는다.

## 적대적 리뷰

### A1 — "격리 알림을 critical로 만들면 계좌가 더 자주 멈추지 않나"

멈추지 않는다. 같은 포지션에 대해 `exit.judgement_refused`가 **이미** critical이고
**이미** 발행된다(운영 원장의 4·5·6번 행이 그것이다). 새 이벤트는 그것을 한 사이클
앞당기고 식별자를 실을 뿐, 이전에 알림이 없던 상황을 알림이 있는 상황으로 바꾸지
않는다. 새 차단 조건은 0개다.

### A2 — "그럼 사이클 실패는 왜 critical이 아닌가. 일관성이 없다"

일관성의 축이 다르기 때문이다. 등급은 "이 조건이 포지션이 보호받지 못한다는 뜻인가"로
정해지고(`event.go`), 격리는 예, 사이클 실패는 **경우에 따라 다르다**. 후자를 critical로
만들면 SQLite busy 하나가 outbox → 전달 실패 → gate latch → ENTRY_BLOCKED가 된다.
`event.go`가 measurement 이벤트에 대해 이미 같은 판단을 명시적으로 적어 두었다.

### A3 — "되읽기가 race를 만들지 않나. 다른 writer가 그 사이에 해제하면?"

해제는 사람이 콘솔에서 하는 동작이고 격리 생성과 밀리초 단위로 겹칠 확률은 실질적으로
0이지만, 겹쳐도 안전하다. `announceQuarantineFromLedger`는 `active == false`면 **조용히
반환한다.** 알림이 하나 덜 나가고, 판정 경로의 error는 원래대로 반환된다. 반대 방향
(다른 원인으로 새 격리가 이미 있는 경우)이면 그 격리를 알린다 — 그것도 참인 사실이다.

### A4 — "`workingSet` B11에서 발행하면 corrupt 행은 매 사이클 알리게 된다"

**맞았고, 그것이 변이 검증이 잡은 결함이다(I4).** corruption 검사는 활성 격리 검사보다
앞에 있어서 corrupt 행은 매 사이클 `QuarantineExitSnapshot`을 호출하고 같은 행을
돌려받는다. latch가 그것을 흡수하며, latch를 지우면 3사이클에서 3회 발행으로 RED가
된다. 리뷰 전 초안의 테스트는 1사이클만 관측해 이 하중을 보지 못했다.

### A5 — "publisher가 배선되면 `alertRefused`가 청산을 지연시키지 않나"

이것이 `workingSet`의 `return append(out, refused...)` 순서가 존재하는 이유이고,
a064는 그 순서를 유지한다. 함수 주석이 명시한다 — "alertRefused may synchronously
retry a publisher, so every valid position—including an emergency breach—is
recorded/armed/submitted before alert delivery can wait." 새 발행도 같은 위치, 즉
분류 단계에서만 일어나고 격리된 포지션에 대해서만 일어난다. 격리된 포지션은 정의상
이번 사이클에 발의하지 않는다.

**주의로 남긴다**: publisher가 실제로 켜지면 이 순서 가정이 처음으로 실전에 걸린다.
`obs.Ntfy`의 기본 timeout은 10초이고 critical 경로는 3회 재시도(간격 2초)이므로
최악 34초를 사이클 안에서 소비할 수 있다. 유효 포지션의 발의가 그보다 먼저 끝난다는
것이 안전 논거의 전부다. 배포 후 실측(11.1)에서 확인 대상.

### A6 — "audit에 topic을 안 남기면 무엇이 바뀌었는지 추적이 되나"

§0.5가 요구하는 것은 "운영 설정 변경이 시각·주체와 함께 추적 가능"이고, 그 질문은
값 없이 답해진다: 알림이 켜졌다/꺼졌다, 어느 서버로 보낸다, 채널과 자격증명이
있다/없다가 시각·주체와 함께 남는다. topic 값 자체를 남기면 §0.8 위반이며 — ntfy.sh
구성에서 그것은 bearer secret과 같다 — 변이 M6이 그 줄을 그대로 인용하며 실패한다.

### A7 — "설정 오류로 기동을 거부하지 않는 것은 fail-open 아닌가"

fail-open이 아니다. 실패 시 도달하는 상태가 **오늘의 상태**다: publisher 없음 →
critical 알림 outbox 보존 → entry gate latch → 지속되면 ENTRY_BLOCKED. 이것은
risk-management가 지정한 방향이고 지금 이미 벌어지고 있다. 반대로 기동을 거부하면
손절 평가 자체가 멈춘다 — 그것이 fail-open이다.

## 구현 후 확인

| 항목 | 결과 |
|---|---|
| `go build ./...` | OK |
| `go vet ./...` | clean |
| `go test ./... -count=1` | **6136 passed / 79 packages** (a064 이전 6095, 신규 41) |
| `openspec validate --all --strict` | 61 passed, 0 failed |
| `check_analysis.py` | evidence complete (7 함수) |
| 변이 검증 | 7건 전부 RED, 복원 후 6파일 바이트 동일 |

### 변이 표

| # | 변이 | 결과 |
|---|---|---|
| M1 | `reportCycle`의 로그 호출 제거 | `TestRunReportsAFailedCycle` FAIL |
| M2 | 격리 이벤트를 normal 등급으로 | `TestAQuarantineCreationIsCritical` FAIL |
| M3 | in-process latch 제거 | `TestACorruptSnapshotQuarantine…` FAIL (3회 발행) |
| M4 | latch 키에서 version 제거 | `TestANewQuarantineVersionIsAnnouncedAgain` FAIL |
| M5 | 거부된 알림 블록을 부분 적용 | `TestNotificationsRefuseANonHTTPBaseURL` FAIL |
| M6 | audit에 topic 값 기록 | `TestTheAuditTrailCarriesNoNotificationSecret` FAIL |
| M7 | 설정 off에서도 publisher 조립 | `TestADisabledBlockWithAChannelStaysOff` FAIL |

M3은 첫 시도에서 **아무 테스트도 깨뜨리지 못했다.** 그 조사가 latch의 실제 하중
지점을 드러냈고 테스트를 고쳤다 — `issues.md` I4.

### Pre-Edit 선언의 사후 검증

| 선언 | 확인 |
|---|---|
| 1. 판정 로직 불변 | `exitpolicy` 패키지 diff 0줄 |
| 2. 원장 write 함수 불변 | `exit_snapshot.go`·`exit_state.go` diff 0줄 |
| 3. 판정 대상 집합 불변 | `TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet` — 알림 유/무로 `ExitCycle`과 저장 상태 동일 |
| 4. 알림을 켜지 않음 | 기본값 off, `TestStartupWiresNoPublisherWhenNotificationsAreOff` |
| 5. 시크릿 미기록 | M6 + `TestTheNotificationBodyCarriesNoToken` + config에 토큰 필드 부재 |
| 6. §0.3 순서 유지 | `append(out, refused...)` 그대로; A5의 주의는 11.1로 이월 |
