# a075 · Review

## Pre-Edit Gate (편집 전 선언)

Base commit: `e540668fe56ea64ff1c3a3ae0b3629be3c149c77`

편집 계획은 **기존 함수 셋**이었다. 완료 후 도구가 요구한 증거는 여덟 개이고, 그 차이
자체가 기록할 가치가 있다 (issues I1).

| # | 함수 | 파일 | 분기 | 위험 | 편집 범위 |
|---|---|---|---|---|---|
| 1 | `runConsole` | `cmd/tossctl/console.go` | 36 | High | Options 리터럴에 필드 한 줄 |
| 2 | `Console.routes` | `internal/console/console.go` | 2 | High | 라우트 세 줄 |
| 3 | `Console.settingsView` | `internal/console/settings.go` | 11 | Normal | 읽기 블록 하나 |
| 4 | `consoleNotificationSeam` | `cmd/tossctl/console.go` | 1 | Normal | 신규 (기존 파일 안) |
| 5 | `consoleGateSwitchSeam` | `cmd/tossctl/console.go` | 1 | Normal | **무편집** — hunk 인접 |
| 6 | `fullSettingsHarness` | `…/settings_cadence_test.go` | 0 | Normal | seam 주입 한 줄 |
| 7 | `TestEveryRouteGoesThroughTheSessionGate` | `…/static_test.go` | 3 | High(guard) | 하한 27→30 |
| 8 | `TestEveryStateChangingRoute…CSRFGate` | `…/static_test.go` | 6 | High(guard) | 목록 세 항목 |

**선언 1 — 새 코드는 새 파일에만 둔다.** 새 함수를 기존 파일에 쓰면 그 함수도
"수정된 기존 함수"가 된다. 실제로는 `consoleNotificationSeam` 하나가 그렇게 됐고,
그것은 어쩔 수 없었다 — 이 저장소는 `internal/console`을 `cmd/tossctl/console.go`
하나만 import하도록 강제하므로 인터페이스를 이름으로 부르는 어댑터는 거기 살아야 한다.

**선언 2 — 판정·주문·원장 경로를 건드리지 않는다.** `exitpolicy`, `journal`,
`execgw`, `obs.Notifier`, outbox, entry gate, operating mode 어느 것도 편집 대상이
아니다. a075의 diff에 그 패키지들은 없다.

**선언 3 — 알림을 자동으로 켜지 않는다.** 켜는 것은 사람의 클릭이다 (§0.7). 이
change가 만드는 것은 그 사람이 누를 버튼이고, 누르기 전까지 파일에 한 바이트도
쓰이지 않는다.

**선언 4 — 화면이 값을 고르지 않는다.** `Enable()`은 인자를 받지 않는다. 채널 이름을
받는 입력란도, 토큰을 받는 입력란도 만들지 않는다 — 만들면 그것이 곧 붙여넣을 자리다.

**선언 5 — 채널 식별자를 audit·로그·리다이렉트 URL에 남기지 않는다** (§0.8).
남기는 곳은 설정 파일(0600, 원자적 rename)과 응답 본문 딱 둘이다.

**선언 6 — 정적 검사 목록과 스펙 문장을 같은 커밋에서 움직인다.** 상태변경 목록의
확장은 스펙이 그렇게 요구한다.

## 적대적 리뷰

### A1 — "콘솔이 알림을 켤 수 있게 하는 것은 §0.7 위반 아닌가"

아니다. §0.7은 "운영 토글 flip과 live 검증은 **사람이 직접 승인**한다"이고, 그것은
승인의 **주체**를 정하는 문장이지 승인의 **장소**를 정하는 문장이 아니다.
`console-owns-the-operating-toggles`가 자동화 게이트에 대해 같은 문장을 같은 방식으로
읽었고, 그 판단이 이미 랜딩되어 있다.

그리고 알림은 게이트보다 결과가 **작다**: 켜도 주문 능력이 생기지 않고, 끄면 오늘로
돌아온다. 게이트를 콘솔에 준 저장소가 알림을 손편집에 남겨 둘 이유가 없다.

### A2 — "채널을 기계가 만들면 운영자가 자기 topic을 못 쓴다"

못 쓰는 것이 맞고 그것이 설계다 (design D1). 다만 **자체 호스팅 경로는 막지 않았다**:
`base_url`은 파일에서 그대로 읽히고, a074의 `TOSSCTL_NTFY_TOKEN` 환경변수 경로는
손대지 않았다. 자체 호스팅 ntfy를 토큰 뒤에 두고 쓰는 운영자는 config.json의
`base_url`을 자기 서버로 바꾸면 되고, 그 편집은 이 버튼이 덮어쓰지 않는다 —
`Enable`은 기존 `base_url`이 있으면 유지한다.

바꾸지 못하는 것은 **topic 하나**다. 그리고 그것이 접근 제어인 유일한 값이다.

### A3 — "테스트 발송이 §0.3(청산 즉시성)을 침해하지 않나"

침해할 수 있는 위치에 있지 않다. 발송은 **콘솔 프로세스**에서, **운영자의 클릭에**
일어난다. 엔진의 exit 관측 사이클과 같은 goroutine도, 같은 프로세스도 아니다.
a074 리뷰 A5가 남긴 주의 — publisher가 켜지면 `alertRefused`의 동기 재시도가 사이클
안에서 최대 34초를 쓸 수 있다 — 는 **여전히 유효하고 a075가 바꾸지 않는다.** 그것은
엔진 쪽 관측 대상이며 a074 task 11.1에 남아 있다.

### A4 — "테스트가 실패했는데 켜진 채로 두는 것이 fail-open 아닌가"

아니다. 실패 시 도달하는 상태는 **알림이 켜져 있고 아직 도착을 확인하지 못한 상태**이며,
되돌렸을 때 도달하는 상태는 **알림이 꺼진 상태 — 오늘의 결함 그 자체**다. 둘 중
어느 쪽이 더 안전한지는 물어볼 것도 없다.

또한 발송 실패의 원인 대부분은 설정이 아니라 순간이다: DNS, 프록시, 방금 재시작한
컨테이너. 그런 것 때문에 방금 만든 채널을 버리면 운영자는 다시 만들고 다시 구독해야
한다. 화면은 실패 사유를 그대로 쓰고 [테스트 한 통 더]를 남긴다.

### A5 — "config.json에 채널을 쓰면 그 파일을 읽을 수 있는 무엇이든 알림 채널에 쓸 수 있다"

맞고, a074의 파일 헤더가 이미 그 이유로 환경변수 경로를 남겨 두었다. a075는 그
경로를 없애지 않았다 — `TOSSCTL_NTFY_TOPIC`은 여전히 파일을 덮어쓴다 (엔진 쪽).

파일에 쓰는 쪽을 택한 이유는 **버튼 하나**라는 요구 자체다. 환경변수는 컨테이너
재생성·systemd unit 편집을 요구하고, 그것이 이 change가 없애려는 마찰이다. 파일은
0600이고 원자적으로 rename되며, 같은 파일이 이미 계좌 편입 설정과 Guardian 한도를
담고 있다 — 채널을 거기 두는 것이 그 파일의 민감도를 올리지 않는다.

### A6 — "라우트 세 개는 너무 많다. 하나로 토글하면 되지 않나"

토글 하나면 "테스트만 다시 보내기"가 불가능해진다. 그리고 이 change의 핵심은
**확인**이다 — 운영자가 구독한 다음에 한 번 더 눌러볼 수 있어야 처음의 테스트가
구독 전에 발송되어 못 받은 경우를 구분할 수 있다.

세 라우트는 전부 이름이 무엇을 하는지 말하고, 전부 상태변경 목록에 열거되어 있으며,
전부 세션+CSRF 뒤에 있다.

### A7 — "`consoleGateSwitchSeam`의 증거는 형식주의 아닌가"

형식주의에 가깝지만 무해하지 않다. `check_analysis.py`의 `intersects`는 count==0인
삽입에 대해 `start <= line <= end + 1`을 쓰므로, 함수 바로 아래에 새 함수를 넣으면
위 함수가 "수정됨"으로 잡힌다. 증거를 만들면서 실제로 확인한 것은 하나다 —
`git diff`의 hunk가 `@@ -627,0 +633,10 @@`, 즉 **삭제 0줄**이라는 것. 그 함수의
바이트는 바뀌지 않았다. 도구를 우회하는 것보다 이 확인을 남기는 편이 싸다.

## 구현 후 확인

| 항목 | 결과 |
|---|---|
| `go build ./...` | OK |
| `go vet ./...` | clean |
| `go test ./... -count=1` | **6162 passed / 79 packages** (a075 이전 6136, 신규 26) |
| `openspec validate a075 --strict` | valid |
| `check_analysis.py` | evidence complete (8 함수) |
| 변이 검증 | 8건 전부 RED, 복원 후 5파일 바이트 동일 |

### 변이 표

| # | 변이 | 결과 |
|---|---|---|
| M1 | 채널 생성에서 난수를 제거 | `TestTwoChannelsAreNeverTheSame` FAIL |
| M2 | 끄기가 세 키를 다 쓰게 (채널 삭제) | `TestTurningNotificationsOffWritesOnlyTheSwitch` FAIL |
| M3 | audit에 채널 값 기록 | `TestTheAuditTrailCarriesNoChannelValue` FAIL |
| M4 | notice에 채널 주소 포함 | `TestTheNoticeNeverCarriesTheChannel` FAIL |
| M5 | 테스트 실패 시 설정 롤백 | `TestAFailedTestLeavesAlertsOn` FAIL |
| M6 | 다시 켤 때 채널 재생성 | `TestEnablingAgainKeepsTheSameChannel` FAIL |
| M7 | 카드에 토큰 입력란 추가 | `TestThereIsNoTokenInputOnTheAlertCard` FAIL |
| M8 | 켜기가 테스트를 보내지 않게 | `TestTurningAlertsOnAlsoSendsTheTest` FAIL |

M1의 첫 시도는 컴파일 오류였다(사용되지 않는 import). 두 번째 시도 —
`rand.Read(buf)`를 `rand.Read(nil)`로 — 가 정확히 원하는 결함을 만들었고,
모든 draw가 `tossos-aaaaaaaaaaaaaaaaaaaaaaaaaa`가 되며 실패했다.

### Pre-Edit 선언의 사후 검증

| 선언 | 확인 |
|---|---|
| 1. 새 코드는 새 파일 | 예외 하나(`consoleNotificationSeam`), 이유는 import 규칙 — 증거 작성됨 |
| 2. 판정·주문·원장 무편집 | `exitpolicy`·`journal`·`execgw`·`obs/notifier.go` diff 0줄 |
| 3. 자동으로 켜지 않음 | 기본값 off, 버튼 전에는 파일 무변경 |
| 4. 화면이 값을 고르지 않음 | `Enable()` 무인자 + `TestThereIsNoTokenInputOnTheAlertCard` (M7) |
| 5. 시크릿 미기록 | M3 + M4 + `TestTheAuditTrailCarriesNoChannelValue` |
| 6. 목록과 스펙 동시 이동 | `consoleStateChanging`·CSRF 목록·`operator-console` 문장이 이 커밋에 함께 있다 |

## 병합 후 재기준화 (2026-08-04)

이 change는 `main`과 동기화되지 않은 브랜치 위에서 작성됐다. `main`은 그동안
journal schema를 15에서 19로 올렸고, 그 브랜치로 빌드한 이미지는 콘솔만 뜨고
엔진이 journal 열기를 거부했다. 배포하려면 병합이 선행 조건이었다.

병합은 Function Logic Map 검사의 의미를 바꾼다. `check_analysis.py`는
`base-commit.txt`부터 작업 트리까지를 diff하므로, 병합 이후에는 `main`이 고친
함수까지 이 change가 고친 것으로 집계된다.

| 항목 | 값 |
|---|---|
| 원래 base | `e540668fe56ea64ff1c3a3ae0b3629be3c149c77` |
| 재기준화된 base | `6b47be2ff8ebd21a59dec682183e015cec8584da` (병합 커밋) |

**재기준화 전에 원래 base로 검사를 다시 돌렸다.** `e540668fe56e` 커밋을 detached
worktree에 체크아웃하고 `check_analysis.py --change`를 실행해
`evidence complete or diff-proven exempt`를 확인했다. 즉 이 change가 실제로 고친
함수의 증거는 완전했고, 병합이 그것을 지운 것이 아니다.

`main`이 고친 함수들의 logic map은 저장소 안에 그대로 있다 —
`openspec/changes/archive/2026-08-03-a061-show-history-instrument-names/`와
`.../2026-08-03-a062-reconcile-owned-orders/`의 analysis가 그것이다.

`revision: current`로 고정된 AST 증거는 병합된 파일 내용으로 다시 추출했다.
분기 ID 집합이 달라진 타깃은 자동으로 덮어쓰지 않고 따로 처리했다.

**게이트 결과의 의미는 좁아졌다.** 재기준화 이후 이 change의 logic-map 단계는
"병합 시점 이후로 이 change가 더 고친 함수가 없다"를 확인하는 것이고, "이 change가
고친 함수에 증거가 있다"는 위의 원래 base 재검사가 확인한다. 한 명령으로 둘 다
재도출되지는 않는다.
