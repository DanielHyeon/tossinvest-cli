# Function Logic Map — a055-console-settings-cadence

작성 시점: 구현 전 (WORKFLOW §6, tasks 1.7). base commit `b331f664`.

기존 함수 **내부**를 편집하는 것은 아래 넷뿐이다. 나머지 변경은 신설 함수와 template
상수이며, 기존 handler 45곳은 `redirectSettings` 한 곳을 통해서만 목적지가 바뀐다 —
호출부를 45번 고치는 대신 호출되는 함수가 요청 경로에서 소속 탭을 유도한다.

---

## F1. `(*Console).redirectSettings` — `internal/console/settings.go:390`

### 현재 로직

```
redirectSettings(w, r, notice):
  B1: 무조건  → 303 "/settings?notice=" + escape(notice)
```

분기 0개. 45개 호출부 전부가 같은 한 화면으로 돌아간다.

### 변경 후 로직

```
redirectSettings(w, r, notice):
  B1: origin ← settingsOriginFor(r.URL.Path)
  B2: origin 미등록  → tab=pathSettingsDaily, form=""   (안전 기본값)
  B3: form == ""     → 303 tab + "?notice=…"
  B4: form != ""     → 303 tab + "?notice=…&form=" + form + "#" + form
```

### 왜 요청 경로에서 유도하는가

저장 결과를 만든 폼이 어느 것인지 아는 것은 **그 POST 라우트 자신**이다. notice 문자열을
파싱해 추측하면 문구가 바뀔 때마다 조용히 어긋나고, 45개 호출부에 인자를 추가하면
새 handler가 인자를 빠뜨려도 컴파일된다. 라우트 → (탭, 폼) 표는 정적 검사로 완전성을
강제할 수 있다(아래 B-map T4).

### 위험

- **High-risk 아님.** 저장 판정·검증·audit·CSRF 게이트를 건드리지 않는다. 이 함수는
  저장이 끝난 뒤의 응답 헤더 하나만 쓴다.
- 회귀 위험은 "폼 결과가 잘못된 탭에 뜬다" 하나이며 T1~T4가 덮는다.

### Branch Test Map

| Bn | 조건 | 테스트 |
|---|---|---|
| B1 | 등록된 POST 경로 | `TestASaveResultComesBackToTheFormThatCausedIt` |
| B2 | 미등록 경로 | `TestAnUnmappedSettingsPostStillLandsOnASettingsTab` |
| B3 | form 없는 origin | 위 같은 테스트의 기본값 절 |
| B4 | form 있는 origin | `TestASaveResultComesBackToTheFormThatCausedIt` |
| T4 | 표 완전성 | `TestEverySettingsPostNamesTheFormItReturnsTo` (라우트 표 대조) |

---

## F2. `(*Console).handleSettings` — `internal/console/settings.go:140`

### 현재 로직

```
handleSettings(w, r):
  B1: page ← settingsPage{chrome, CSRF, Notice, EngineRunning}
  B2: opts.Settings != nil      → Wired=true; Load(); LoadErr/Block/Verdict
  B3: opts.Limits != nil        → LimitsWired=true; Load(); LimitsLoadErr/Gate
  B4: opts.TradingPolicy != nil → TradingWired=true; Load(); TradingLoadErr/Trading
  B5: GateWired ← opts.Gate != nil
  B6: opts.EngineBoot != nil    → AutostartWired=true; Load(); AutostartLoadErr/Autostart
  B7: AutostartNote ← engineNoteNow()
  B8: opts.SystemUpdater != nil → UpdateWired=true; Inspect(); signedReleaseReceipt()
  B9: ReleaseDownloadWired ← Downloader != nil && Stager != nil
  B10: render "settings"
```

### 변경 후 로직

B1~B9의 **읽기는 한 글자도 바뀌지 않는다.** 바뀌는 것은 셋이다.

```
handleSettings(w, r):                      ← /settings, 이제 리다이렉트 전용
  B1': r.URL.Path != "/settings" → 404 refuse   (mux 정확 매치이므로 방어적)
  B2': fragment는 서버에 오지 않는다 → 질의 파라미터로 온 앵커만 옮긴다
  B3': 303 pathSettingsDaily (기본 진입 = 가역·일 단위)

settingsView(r) → settingsPage:            ← 신설. 위 B1~B9를 그대로 옮긴 것
handleSettingsTab(tab) → http.HandlerFunc: ← 신설. settingsView + Tab 지정 + render
```

`/settings#adoption`의 fragment는 **HTTP 요청에 실리지 않는다.** 브라우저가 fragment를
보존한 채 리다이렉트를 따라가므로 `/settings/daily#adoption`이 된다 — 이것은 계약이
요구하는 `/settings/standing#adoption`이 아니다. 따라서 `#adoption`은 서버가 알 수 있는
형태(`?section=adoption`)로도 받고, **daily 탭에 `#adoption` 앵커를 두어** 브라우저가
가져온 fragment가 상시 탭으로 다시 보내지도록 한다(§3.2 테스트가 두 경로를 모두 덮는다).

### 위험

- 읽기 seam 호출 순서와 개수가 그대로임을 테스트로 고정한다(브로커 호출 0건은 a052가 이미 고정).
- `/settings`를 여는 상속 테스트 60여 곳이 리다이렉트를 따라 daily에 도착한다 — 각 테스트를
  그 섹션을 소유한 탭으로 옮긴다. 이것이 이 change에서 가장 넓은 기계적 편집이다.

### Branch Test Map

| Bn | 조건 | 테스트 |
|---|---|---|
| B1' | 잘못된 경로 | mux 정확 매치로 도달 불가 — 방어 코드, 테스트 면제 |
| B2' | `?section=adoption` | `TestTheOldAdoptionLinkLandsOnTheStandingTab` |
| B3' | 기본 진입 | `TestOpeningSettingsLandsOnTheDailyTab` |
| B2~B9 (이동분) | 각 seam 미배선/배선 | 상속 테스트 전량 + `TestEachTabHeaderStatesTheOperatingState` |

---

## F3. `(*Console).routes` — `internal/console/console.go:710`

### 현재 로직

라우트 40건을 `mux.HandleFunc` 리터럴로 등록한다. 분기 2개(`c.opts.Remote != nil`,
`c.opts.OpenAPI != nil` 계열).

### 변경 후 로직

GET 4건 추가. 리터럴 경로만 쓴다 — `registeredRoutes`가 비리터럴 경로를 거부한다.

```
+ mux.HandleFunc("/settings/standing", c.session0(c.handleSettingsTab(tabStanding)))
+ mux.HandleFunc("/settings/daily",    c.session0(c.handleSettingsTab(tabDaily)))
+ mux.HandleFunc("/settings/strategy", c.session0(c.handleSettingsTab(tabStrategy)))
+ mux.HandleFunc("/settings/tools",    c.session0(c.handleSettingsTab(tabTools)))
```

### 위험 — 정적 검사 4종을 동시에 만족해야 한다

| 검사 | 요구 | 신설 4경로의 판정 |
|---|---|---|
| `TestEveryRouteGoesThroughTheSessionGate` | `session0` 필수 | 통과 |
| `TestNoRouteIsRegisteredWithAMethodPattern` | 경로에 공백 금지 | 통과 |
| `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | 읽기는 CSRF 금지 | `mutating` 미적용 → 통과 |
| `TestNoRouteNamesAnAccountMutation` | `actVerbs`·`accountVerbs` 금지 | **검증함**: standing·daily·strategy·tools 넷 중 어느 것도 `start/stop/save/limit/preset/include/exclude/enable/config/reset/delete/gate/adopt/token/order/…`를 부분문자열로 포함하지 않는다 |

`opaqueHandler`가 `*ast.CallExpr`를 통과시키므로 `c.handleSettingsTab(tabDaily)`처럼
handler를 반환하는 호출도 등록 형태로 읽힌다 — 다만 그 인자가 `*ast.Ident`(`tabDaily`)라
`opaqueHandler`가 **true**를 반환해 실패한다. 따라서 tab은 **문자열 리터럴**로 넘긴다:
`c.handleSettingsTab("standing")`. (`*ast.BasicLit`는 명시적으로 건너뛴다.)

### Branch Test Map

| Bn | 조건 | 테스트 |
|---|---|---|
| B1 | 신설 4라우트 등록 | `TestTheFourSettingsTabsAreRegisteredGetRoutes` |
| B2 | 전부 GET·CSRF 밖 | 같은 테스트 + 기존 `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B3 | 상태변경 행위 목록 무증가 | `TestTheTabsAddNoStateChangingRoute` |

---

## F4. `currentNavLabel` — `internal/console/screen_paths_test.go:231` (테스트 헬퍼)

### 현재 로직

```
currentNavLabel(page):
  B1: aria-current="page" 없음 → ("", false)
  B2: 그 뒤 '>' 와 '</a>' 사이 텍스트를 이름으로 반환
```

### 변경 후 로직

nav 항목이 라벨과 **설명 한 줄**을 함께 렌더하게 되므로(계약 `콘솔 내비게이션은 운영
흐름을 따른다`) B2가 라벨 대신 "라벨+설명"을 반환한다. `documentTitle`이 `<h1>`의
`<span class="muted">` 수식어를 떼는 것과 같은 이유로, 설명 요소를 뗀다.

```
  B2': 텍스트에서 <small>…</small> 구간을 제거한 뒤 trim
```

설명은 이름이 아니다 — a054 요구사항(`화면 이름·경로·제목은 한 화면을 가리킨다`)은
그대로 성립하고, 검사 대상만 정확해진다.

### Branch Test Map

| Bn | 조건 | 테스트 |
|---|---|---|
| B1 | aria-current 없음 | `TestEveryScreenIsCalledOneThing`의 skip 절 |
| B2' | 라벨+설명 | `TestTheNavigationSaysWhatEachScreenAnswers` + 기존 이름 일치 검사 |

---

## 면제 선언

- **신설 함수**(탭 handler, 카드 표준 view 함수, explain 파서)는 기존 로직 편집이 아니므로
  이 문서의 Branch Test Map 대상이 아니다. 각 신설 함수의 분기는 그 함수를 도입하는
  RED 테스트가 직접 덮는다.
- **template 상수**는 Go 함수가 아니다. 렌더 결과에 대한 검사로 덮는다.
- **High-risk 함수 편집 없음.** 주문·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결
  경로의 함수는 이 change에서 하나도 편집하지 않는다. Guardian 한도는 **표시**만 바뀌고
  저장 검증(`GuardianLimits.Validate`, `CeilingViolations`, writer)은 무변경이다.
