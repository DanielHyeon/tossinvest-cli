# Review: a054-console-status-shell

- Date: 2026-08-02
- Scope: 콘솔 공통 상태 표시줄, 화면 이름·경로 정정, 표시 프리미티브, 개요 상단 요약
- Review class: UI change — 경량 리뷰(validate + Manager 셀프리뷰 + 적대적 Eng 관점 + 기록)
- Voices: Manager 셀프리뷰, 적대적 Eng, DX(검사 가능성), QA/안전
- Status: **accepted with seven corrections applied below**

High-risk 경로(주문 실행·손절·사이징·Guardian·원장·reconciliation·인증) 변경 없음. 다만
루트 경로와 재시작 핸드오프 착지가 **인증 경로에 인접**하므로 적대적 Eng 관점을 포함했다.

## Findings

### F1. 375px 시나리오가 구현 불가능했다 — 수정

초안은 `documentElement.scrollWidth`를 시나리오에 썼다. **이 저장소에 그 측정을 하는 테스트는
없다** — `grep 'scrollWidth\|375' internal/console/*_test.go` → 0건. 기존 반응형 검증은
`strategy_runtime_test.go:195`처럼 렌더 결과에 `name="viewport"`·`@media (max-width: 720px)`·
`detail-grid`가 **문자열로 존재하는지**를 본다.

기존 스펙 문장이 `scrollWidth`를 말한다는 이유로 같은 문장을 복사하면, 검증할 수 없는 SHALL을
하나 더 만드는 것이다.

- **수정**: 시나리오를 렌더 결과에서 판정 가능한 형태로 다시 썼다 — viewport meta 존재, 모든
  표가 반응형 클래스 또는 스크롤 래퍼를 가짐, 고정 px 폭 부재. **실제 레이아웃 측정은
  자동 검사에 넣지 않고**, 브라우저 1회 수동 확인을 tasks의 증거 항목으로 분리했다.

### F2. 프리미티브 통합 방향이 반대였다 — 수정

초안은 "`.status-pill`을 없애고 `.state-badge`만 남긴다"였다. 실제 사용은 반대다.

| 클래스 | 사용 | 테스트 |
|---|---|---|
| `.status-pill` | `templates_optimization.go`에 **8곳** | `strategy_runtime_test.go:195`가 존재를 검사 |
| `.state-badge` | 2곳 | 없음 |

적은 쪽을 남기면 8곳을 고치고 기존 테스트를 깨는 작업이 된다.

- **수정**: `.status-pill`을 남기고 `.state-badge`를 흡수한다.

### F3. 핸드오프 쿼리 보존 요구가 틀렸다 — 수정

초안은 "리다이렉트는 쿼리스트링을 보존해야 한다 — `?handoff=`가 유실되면 재시작 인증이
끊긴다"라고 썼다. 코드는 정반대다.

- `session0`([console.go:801-841](../../../internal/console/console.go#L801-L841))이 핸들러보다
  **먼저** 돌고 `acceptHandoff`가 토큰을 소비한다. 토큰은 `handleDashboard`에 도달하지 않는다.
- `grantSession`([restart.go:249-251](../../../internal/console/restart.go#L249-L251))은
  `q.Del("handoff")`로 **의도적으로 파라미터를 지우고** 같은 경로로 리다이렉트한다.

즉 초안의 요구는 이미 잘못된 전제 위에 있었고, 그대로 구현하면 소비된 토큰을 URL에 남기게 된다.

- **수정**: 요구를 결과로 다시 썼다 — "토큰은 정확히 한 번 소비되고 브라우저는 렌더된 화면에
  착지한다". 보존해야 하는 쿼리는 `?notice=`다(아래 F4).

### F4. 재시작 안내의 착지 화면을 잘못 지정했다 — 수정

design 초안은 [restart.go:321](../../../internal/console/restart.go#L321)의 `target := "/"`를
"개요 경로로" 옮긴다고 적었다. 그 함수는 `redirectDashboard`이고 **재시작·soak 재기동 결과
안내**를 싣는다. 그 컨트롤은 검증 콘솔에 있다 — 개요로 보내면 운영자는 자기가 누른 버튼의
결과를 다른 화면에서 읽게 된다.

- **수정**: 재시작·soak 안내는 **검증 콘솔**로 돌아간다. `?notice=` 쿼리를 보존한다.

### F5. 신선도 톤이 "알려진 갱신 보류"를 오경보한다 — 수정

초안의 톤 규칙은 경과 시간만 본다. 그런데 이 콘솔에는 **갱신하지 않는 것이 정상인 상태**가
둘 있다.

1. 검증 run 진행 중 — `rate budget 보호` 요구사항이 갱신 보류를 **의무화**한다. 보류 중
   캐시는 필연적으로 늙는다.
2. 장 마감·discovery 미실행 — `/signals`의 스캔 시각은 tick이 돌지 않으면 자연히 늙는다.

경과만으로 판정하면 규정대로 동작하는 시스템에 경고를 붙인다. 경보가 늘 켜져 있으면 아무도
안 본다.

- **수정**: 보류 사유가 **알려진 경우** 톤 대신 그 사유를 표시한다. 톤은 사유 없이 늙었을
  때만 붙는다.

### F6. "정적 검사"라고 쓴 것이 정적 검사가 아니다 — 수정

이름 일치(`<h1>` ↔ nav 라벨 ↔ `Nav` 식별자) 검사를 정적 검사로 요구했다. 세 값은 Go 문자열
템플릿 안의 HTML과 핸들러의 문자열 필드에 흩어져 있어 정적 파싱 대상이 아니다.

- **수정**: **렌더 결과 비교**로 바꿨다 — 각 화면을 렌더해 `aria-current`가 붙은 nav 항목의
  텍스트와 `<h1>` 텍스트를 대조한다. 실제로 더 쉽고 더 강하다.

### F7. 승인 창 회귀 위험 — 요구사항 추가

`a055`은 검증 콘솔을 설정 하위 진입점으로 옮긴다. 검증 승인 창은 짧고, **이미 승인 창 소진
사고 기록이 있다**(measurements M11·M18·M22·M23). 발견성이 2클릭 깊어지면 그 사고가 다시
난다.

- **수정**: 상태 표시줄이 **승인 대기 중인 검증 run을 표시하고 그 화면으로 직접 링크**하도록
  요구사항을 추가했다. 표시줄은 모든 화면에 있으므로 발견성은 오히려 지금보다 좋아진다.
  이것이 `a055`의 이동을 안전하게 만드는 전제다 — 두 change의 의존 방향을 명시했다.

### F8. 임베딩이 `Refresh`도 가져간다 — tasks 보강

`RefreshSeconds()` shadowing만 테스트하기로 했는데, `Refresh bool`도 chrome으로 옮겨간다.
어떤 화면이 이 필드 설정을 빠뜨리면 **자동 재로드가 조용히 꺼진다** — 컴파일도 테스트도
실패하지 않는다.

- **수정**: 화면별로 "재로드가 걸린다/걸리지 않는다" 두 방향 모두 테스트하도록 tasks 2.1을
  고쳤다.

## 수용하지 않은 지적

- *"루트 경로를 그냥 두고 nav 라벨만 고치면 위험이 없다"* — 거부. 한 화면에 이름이 셋인 상태를
  남기면 상태 표시줄이 "지금 어느 화면인가"를 정직하게 말할 수 없고, `a055`의 nav 재편이 그
  모순 위에 쌓인다. 대신 F3·F4로 이동 경로의 실제 위험을 좁혔다.
- *"상태 표시줄에 kill switch 스위치를 함께 두면 편하다"* — 거부. 콘솔의 상태변경 행위 목록을
  늘리지 않는다는 것이 이 change의 전제다. 표시줄에는 폼이 없다.

## Verification evidence

- `openspec validate a054-console-status-shell --strict --no-interactive` → valid.
- 상태 표시줄이 읽는 값 전수 확인(design §1): 신규 데이터 소스 0개, 신규 브로커 호출 0건.
  `binstamp.Of`는 SHA-256이 아니라 `os.Stat`([binstamp.go:82-97](../../../internal/binstamp/binstamp.go#L82-L97)).
- `"/"` 참조 분류 확인: 라우트 6곳 / 쿠키 스코프 3곳 / URL 정규화 2곳. 쿠키 스코프 3곳은
  변경 대상이 아님을 design §4에 고정.
- 구현 착수 전이므로 테스트 실행 결과 없음. 이 리뷰는 proposal-freeze 리뷰다.

## Function Logic Map

기존 함수 내부 편집이 있다 — `handleDashboard`, `redirectDashboard`, 각 page 생성 handler,
`head` 렌더 경로. **not-applicable 아님.** tasks 1.5가 구현 전 Function Logic Map과
Branch Test Map 작성을 요구한다.

## Verdict

일곱 개 수정을 반영한 상태로 proposal을 freeze한다. 원격 노출·엔진 기동·LIVE 게이트 변경·
계좌 변경은 범위 밖이다. 구현은 `a055`보다 먼저 완료한다.

---

# 구현 후 추기 (tasks 6.4)

proposal-freeze 리뷰는 위까지다. 아래는 구현하며 **코드가 계약을 반박한 지점**과 그 처리다.

## 구현이 뒤집은 것 — 4건

### I1. 세션 advisory를 Go에서 읽으면 기존 안전 가드가 깨진다

초안 `chromeFor`는 `session.Outside`로 "open/closed" 문자열을 만들었다.
[static_test.go의 `TestTheMarketHoursAdvisoryCannotBlockAnything`](../../../internal/console/static_test.go)이
`data.go`·`templates.go` 외의 **모든** Go 파일에서 `.Outside`를 금지하고 있고, 이유가
살아 있다 — 브로커가 주문을 받는 실제 창은 미측정이므로 이 값을 읽는 Go 코드가 하나 늘 때마다
"advisory만"이라는 규칙이 참이어야 할 자리가 하나 늘어난다.

- **처리**: `statusStrip`이 `verifylive.SessionAdvisory` **값 전체**를 그대로 나른다.
  open/closed 판정은 템플릿이 한다(`templates.go`는 허용 목록에 있고, 기존 `hours` 템플릿이
  이미 같은 일을 한다). 가드에 파일을 추가해 통과시키는 쪽은 택하지 않았다.

### I2. `/signals`의 갱신 보류를 장 마감으로 판정하려 했다

design §2는 "스캔 주기 미도래"를 톤 억제 사유로 들었고, 그것을 세션 advisory로 판정하려면
I1을 다시 위반해야 한다.

- **처리**: `/signals`는 **TTL을 두지 않는다**. 이 화면의 값은 `tossctl candidate watch`의
  tick에만 전진하고 콘솔은 그 tick을 일으키지도 관측하지도 않으므로, 콘솔이 가지지 않은
  기준으로 나이를 채점하는 것 자체가 틀렸다. 톤 대신 "무엇이 이 값을 전진시키는가"를 항상
  표시한다. 장 마감 조건보다 단순하고 더 정직하며, advisory를 읽지 않는다.

### I3. 죽은 `onsubmit="return confirm(…)"` 3종 — 지우지 않고 고정했다

task 4.6은 렌더 결과에 인라인 핸들러가 없을 것을 요구한다. 설정 화면에 3종이 있고 **CSP
`default-src 'none'`(script-src 없음) 아래에서 실행되지 않는다** — 소스에만 있는 승인 단계다.

지우려 했으나 [settings_limits_test.go](../../../internal/console/settings_limits_test.go)의
`TestThePresetControlsAskForNoTyping`이 그 존재를 **단언하고 있다.**

- **처리**: 지우지 않았다. 둘은 Guardian 한도·시스템 업데이트 폼이라 High-risk 표면이고,
  승인처럼 생긴 것을 표시 변경의 부수 효과로 떼어내는 것은 이 change가 할 판단이 아니다.
  상속 테스트의 단언을 조용히 삭제하는 것도 마찬가지다. 대신 **줄어들기만 하는 재고**로
  고정했다 — 설정 외 화면에는 0개, 설정 화면의 것은 전부 같은 종류(`onsubmit="return
  confirm(`)여야 하며, 다른 화면에 하나라도 생기거나 종류가 바뀌면 FAIL한다. 응답 CSP에
  `script-src`가 없다는 것도 함께 고정했다 — 그것이 생기면 이 핸들러들이 죽은 것이 아니게 된다.
  제거는 `a055` issues.md I2 / task 4.6이 소유한다.

### I4. `/strategy-runtime`·`/market-schedule`은 nav 항목이 없다

두 화면은 `Nav: "optimization"`이라 **자기가 아닌 화면**에 `aria-current="page"`가 붙고 있었다.
이름 일치 검사가 이것을 잡았다 — 최적화라고 표시하면서 제목은 "한국 주식 전략 lane"이다.

- **처리**: nav 항목을 새로 만드는 것은 `a055`의 범위다. 두 화면의 nav 키를 어떤 항목과도
  맞지 않는 값으로 바꿔 `aria-current`를 떼었다. 두 화면 모두 최적화 화면의 자체 사이드바로
  도달하고 거기에 위치 표시가 있다. `a055`이 정규 진입점을 준다.

## 상속 테스트에 손댄 곳과 이유

| 테스트 | 변경 | 이유 |
|---|---|---|
| `TestTheRootScreenIsUnchangedAndTheTwoScreensAreNamedApart` | 이름·경로 갱신 + **단언 추가** | 이 change가 바꾸는 바로 그 계약. "승인 창을 보던 탭이 다른 화면으로 옮겨지면 안 된다"는 원래 취지는 버리지 않고, 옛 루트에서 리다이렉트된 탭이 **승인 대기와 직접 링크를 받는지**를 새로 단언했다(F7) |
| `TestTheOverviewReloadsAtTheCacheTTL` | 타입 단언 → 렌더 단언 | `Refresh`가 메서드에서 필드로 옮겨졌다. 렌더 검사가 더 강하다 — 핸들러가 설정을 빠뜨리면 잡히고, 메서드 형태는 잡지 못했다 |
| remote 테스트 10곳 | `GET /` → `GET /dashboard` | 루트가 항상 303이 되어 "인가됨(200) vs 미인가(303→/login)"을 구분할 수 없게 됐다. 그 구분이 이 테스트들의 전부다 |
| 검증 콘솔을 여는 테스트 8곳 | 경로만 이동 | 같은 화면, 새 경로 |
| `TestExternalPositionAutomaticManagementHasADiscoverableMenu` | `<h1>` 문자열 | 수식어가 `<span class="muted">`로 분리됐다 |

**약화한 단언은 없다.** 위 표에서 늘어난 것은 하나(`TestTheRootScreen…`), 나머지는 대상 경로와
문자열 갱신이다.

## 변이 검증 (task 6.1)

각 변이를 넣고 지정 테스트를 돌린 뒤 되돌렸다. 전부 FAIL해야 통과다.

| 변이 | 테스트 | 결과 |
|---|---|---|
| `/orders`의 `RefreshSeconds()` 삭제 | `TestEachScreenKeepsItsOwnReloadPeriod` | RED |
| `/positions`의 `page.Refresh = true` 삭제 | `TestEachScreenKeepsItsOwnReloadPeriod` | RED |
| 개요의 `page.Refresh = true` 삭제 | 〃 | RED |
| 검증 갱신 보류 사유 삭제 | `TestAKnownRefreshHoldGetsAReasonInsteadOfATone` | RED |
| 엔진 미배선을 "정지"로 표기 | `TestTheEngineCellSeparatesUnwiredFromStopped` | RED |
| 세션 쿠키 `Path`를 화면으로 축소 | `TestTheSessionCookieStaysScopedToTheWholeConsole` | RED |
| 루트를 404로 | `TestTheRootPathAnswersWithTheOverview` | RED |
| 재시작 안내를 개요로 | `TestARestartNoticeComesBackToTheScreenThatStartedIt` | RED |
| 검증 콘솔 제목을 "대시보드"로 되돌림 | `TestEveryScreenIsCalledOneThing` | RED |
| 표 하나의 스크롤 래퍼 제거 | `TestEveryTableInTheTemplatesIsResponsiveOrScrolls` | RED |
| `.state-badge` 부활 | `TestOneNameForOneStatusDisplay` | RED |
| h2를 본문 크기로 | `TestTheHeadingStepsAreDistinguishable` | RED |
| 개요에 인라인 핸들러 추가 | `TestNoScreenSmugglesInAScript` | RED |
| 요약 칸의 상세 링크 제거 | `TestTheOverviewAnswersFromTheTop` | RED |
| 미측정 보유를 0으로 표기 | `TestAnUnmeasuredSummaryCellSaysSo` | RED |

**약한 테스트 하나를 이 과정에서 찾았다.** `TestTheReloadCellAndTheMetaTagAreOneFact`는
표시줄과 meta tag의 **일치**를 보므로 둘이 함께 꺼지면 통과한다. 존재 여부는
`TestEachScreenKeepsItsOwnReloadPeriod`가 화면별로 본다 — 둘이 함께 F8의 위험을 덮는다.
일치 검사만으로 충분하다고 적지 않았다.

**렌더 검사만으로는 부족한 곳도 찾았다.** 표 검사를 렌더 결과로만 두면 seam이 배선되지 않아
렌더되지 않는 표(대부분)는 래퍼를 떼도 통과한다. 템플릿 소스를 함께 보는 정적 검사를 추가했다.

## 실행 결과

```
착수 전                go test ./...  → 5840 passed / 78 packages
구현 후                go test ./...  → 5870 passed / 78 packages   (신규 30, 회귀 0)
go vet ./...                          → 이슈 없음
openspec validate --strict            → valid
make sdd-check                        → CodeGraph 지문 일치, PM tracker 최신 (GBrain은 advisory·busy)
```

## 남은 것

- **task 4.7 — 사람이 해야 하는 1회 실측.** 브라우저 375px·1280px 레이아웃 확인은 자동
  검사가 아니고 이 저장소에 하니스가 없다(design §5b). **수행되지 않았다.** 자동 검사가 보는
  것은 viewport meta·표의 반응형 수단·viewport보다 넓은 고정 px 부재 넷이며, 실제 레이아웃은
  그 넷으로 증명되지 않는다.
- 미아카이브 `console-*` change의 아카이브 순서 충돌(`a055` issues.md I1)은 이 change가
  ADDED만 쓰므로 구현을 막지 않지만, 아카이브 전에 결론이 필요하다.

## task 4.7 — 브라우저 실측 (수행됨, 실제 결함 2건 발견)

design §5b는 이 실측을 "사람이 1회 수행하는 증거"로 분리했다. 하니스가 없다는 판단은
맞았지만 **측정 자체는 수행했다.** 방법:

1. 각 화면의 렌더 결과를 파일로 덤프했다(테스트 하니스가 만드는 바이트 그대로 — 이 콘솔의
   페이지는 인라인 CSS뿐이고 외부 asset이 없으므로 파일이 곧 같은 문서다).
2. loopback 정적 서버에 올리고 375px·1280px iframe에 넣어
   `documentElement.scrollWidth`와 `clientWidth`를 읽었다.

**첫 측정에서 2건이 실제로 넘쳤다.** 자동 검사 넷은 전부 통과한 상태였다 — 렌더 문자열로는
잡히지 않는 결함이고, design §5b가 이 작업을 남겨둔 이유가 바로 이것이다.

| 화면 | 측정 | 원인 |
|---|---|---|
| `/verify-console` | scrollWidth **606** / clientWidth 360 | soak·attestation·검증 기록의 **파일 경로가 산문 안에** 있고 줄바꿈되지 않았다. 스타일시트의 `overflow-wrap: anywhere`는 `code`에만 걸려 있었다 |
| `/orders` | scrollWidth **379** / clientWidth 360 | 좁은 화면 `dl`의 라벨 열이 `minmax(7rem, auto)`라 max-content로 커지고, 값 열의 최소 폭이 그만큼 밀렸다 |

- **수정**: `body`에 `overflow-wrap: anywhere`를 걸었다. 상속되므로 산문·`dd`·표 셀에 모두
  적용되고, `break-word`가 아니라 `anywhere`인 이유는 grid 열이 재는 intrinsic minimum까지
  줄여야 두 번째 결함도 함께 사라지기 때문이다.

**재측정 결과 (14화면 × 2폭 = 28건):**

```
375px   가로 넘침 0건
1280px  가로 넘침 0건
```

`.table-scroll` 안의 표는 여전히 607px지만 래퍼가 가두므로 문서는 넘치지 않는다 — 설계대로다.

> 남는 한계: 이 측정은 렌더된 정적 문서에 대한 것이고 실제 프로세스가 서비스하는 응답을
> 브라우저로 연 것이 아니다. 두 바이트열은 같지만, 응답 헤더가 관여하는 동작(CSP가 실제로
> 인라인 핸들러를 막는지 등)은 이 측정이 답하지 않는다. 그 쪽은 헤더 단언으로 고정돼 있다.
