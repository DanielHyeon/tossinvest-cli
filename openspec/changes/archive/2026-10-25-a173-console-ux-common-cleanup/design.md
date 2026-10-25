# a173 design

정찰(2026-10-25, HEAD `b27a1ea87`, 읽기 전용) 결과를 근거로 한다. 줄 번호는 그 시점 값이고 편집 전에 다시 확인한다.

## 1. 지형 (어디를 만지나)

| 대상 | 위치 | 비고 |
|---|---|---|
| 공통 껍데기 | `templates.go:450` `head`(nav `:477-484`), `:502` `statusstrip`, `:526` `foot`, `:17` `style` | nav 6개 하드코딩, 상태 표시줄은 `chrome.go` `statusStrip`(`DataTone` 은 `freshness.tone()` `chrome.go:150`) |
| 미측정 사유 | `overview.go:103-141` `unmeasuredReason` 열거(8개), `:146-160` `unmeasuredSentences`, `reading.Why()` `:233` | 문장 정본. 신호 화면은 자기 문구(`signals.go:165`, `templates_signals.go:56,65,86`) |
| 미측정 출력 | `templates_overview.go:25` `unmeasured`, `:28` `cell` | 개요·주문이 공유. 주문 4곳(`templates_orders.go:42,45,50,68`) + 브로커 원문 2곳(`:105,109`); 개요 `:61,:67`(시장별), `:131,:175`(colspan), `:290`, 셀 `:138-142,:180-183,:261,:274` |
| 알림 상자 | `.notice` `templates.go:104`(108곳), `.danger` `:105`(38곳), 톤 고르는 Go 헬퍼 없음(템플릿이 class 를 글자로 고름) | 데이터 속성 톤: `[data-data-tone]`, `[data-state-tone]`(`settings_card.go:350-381`) |
| 설정 | `templates_settings.go` 탭 4개(`:80,:228,:382,:394`), 인터록 표 `:148-155`(`settings_operating.go:84-122` `GatePreflight()`), 거래 정책 폼 `:361` | 체크박스 간격만 (구조는 a174) |
| 최적화 | `templates_optimization.go:4`(카테고리 분기), 카테고리 메뉴 라벨의 change ID | 라벨 텍스트만 기술 정보로 (배치는 a175) |
| 측정 하네스 | `measure/shoot.cjs`, `measure/outline.py` | 경로는 절대경로 인자로만 받게 고친다(기억: 인용된 하네스는 저장소에) |

## 2. 설계

### 2.1 원인 1회 (뷰 모델 + 템플릿)
- 페이지 뷰 모델에 `Causes []causeBanner{Code unmeasuredReason, Sentence, Detail, Action, Href}` 를 둔다. 핸들러가 그 화면의 readings 를 훑어 **코드별 1개**로 모은다(개요·주문·포지션·이력). 순서는 열거 순서.
- `unmeasured` 템플릿은 두 모드: 페이지에 같은 코드의 배너가 있으면 `미측정 · <code>` 만, 없으면 종전처럼 문장까지(안전망 — 배너를 못 만든 화면에서 사유가 사라지지 않게). 템플릿 인자에 `$.Causes` 가 닿지 않는 곳은 `cell` 호출부에서 넘긴다.
- 브로커 원문(`.Broker.Error`)은 배너의 Detail 1회만. 상태 표시줄 데이터 시각 칸은 `읽지 못함 · <code>` + `title` 에 원문.
- 스펙 1374 의 여덟 사유 열거를 벗어나는 코드는 만들지 않는다. 시험: 모든 `consoleScreens` 를 렌더해 각 사유 **문장**(`unmeasuredSentences` 값)의 출현 횟수 ≤ 1, 그리고 `미측정` 이 보이는 모든 칸에 `<code>` 사유가 붙어 있음(사유 없는 미측정 0).

### 2.2 알림 3단계 (CSS 토큰 + 템플릿 class 교체)
- 의미: `.danger` = 차단, `.notice` = 결과 있음, 그 외 = 본문(`.muted` 또는 새 `.aside` 한 줄). class 이름은 유지(시험이 `class="notice"` 를 센다: `settings_card_test.go:105,275`, `console_prose_test.go:29`).
- 전환 규칙(템플릿 108+38 곳 전수 분류표를 tasks 산출물로 남긴다): "반영 시점"·"이 표의 한계"·"이 화면은 ~" 류 → 본문; "기동 거부·읽기 실패·설치 불가·브로커 조회 실패" → `.danger`; "되돌릴 수 없음·저장하면 ~가 된다" → `.notice`.
- `console_prose_test.go:29` 가 `<details>` 안의 `.notice` 를 금지하므로 접힌 블록 안에는 상자를 넣지 않는다(본문만).
- 시험: `consoleScreens` 전수 렌더에서 `.notice|.danger` 개수 ≤ 3 (빈 fixture 와 `seedJournal`/`newOrdersHarness` 채운 fixture 둘 다).

### 2.3 경계 배지 + 접힌 경계
- `head` 에 `<h1>` 옆 `<span class="status-pill scope">읽기 전용|편집 가능</span>`(페이지 struct 의 bool 하나, 폼이 있으면 편집 가능). `TestEveryScreenIsCalledOneThing` 은 `<h1>` 텍스트를 비교하므로 배지는 `<h1>` **밖**에 둔다.
- 머리 경계 문단은 `<details class="scope"><summary>이 화면의 경계</summary>…</details>` 로. 본문 첫 요소는 데이터.

### 2.4 상태 표시줄 압축
- 네 칸·`data-*` 속성·reload 초·`.notice.console-approval` 은 그대로(`status_strip_test.go:35-78,198-200`).
- 칸 설명(`<small>`)은 한 줄 상한: advisory 는 `advisory` 한 단어 + `title`; 데이터 시각 실패는 `읽지 못함 · <code>`; 자동 재로드는 `15초` / `없음`.
- CSS: 데스크톱 `.console-status` 높이 ≤ 72px(측정), 720px 이하는 네 칸을 `display:flex; flex-wrap` 칩으로. 앰비언트 의사요소 규칙(`TestAmbientMotionContract` 가 `.console-status` 선택자 모양을 고정)은 손대지 않는다.

### 2.5 기술 정보 접기
- `<details class="tech"><summary>기술 정보</summary><dl>…</dl></details>` 공용 define. 대상: 카테고리 라벨의 `a0NN`, settings digest, 파일 절대경로(개요 검증 기록·검증 콘솔 마커·저널 경로), `Golden field registry` 는 a175.
- 안전 상태 단어(UNWIRED·OFF·미충족·stale·정지)는 접지 않는다. 시험: 본문(`details.tech` 밖) 텍스트에 `/a0\d{3}/`·`^/` 절대경로·64자 hex 0건, 접은 뒤에도 전체 HTML 에는 그 값이 남아 있음(기존 문자열 시험 보호).

### 2.6 줄 길이·제목·'적용 후' 표
- `.page-intro`(52rem) 를 `p, .prose { max-inline-size: 42rem }` 계열로 확장(표·폼·코드 블록·그리드 제외). 측정은 하네스로 가장 긴 `<p>` 폭.
- 섹션 제목: 영문 식별자 제목(Shared projection contract·Exchange calendar·capability attestation·soak 등)은 `한글 (영문)` 또는 한글 + `<code>`. nav 라벨·h1 일치는 유지.
- "적용 후 / 실행되는 것" 2행 표(`settings_card` 류) → 저장 버튼 옆 `<p class="muted">` 한 줄 "저장하면 …; 실행되는 것은 없다". `settings_card_test.go` 가 마커 앞 60자 안의 `class="notice"` 를 요구하는 두 곳은 그 상자가 "결과 있음" 이면 유지, 아니면 시험을 고치고 기록.

### 2.7 거래 정책 체크박스 간격
- 라벨 블록을 `display:grid` 두 열(체크박스 | 라벨+설명)로 묶어 설명이 자기 체크박스 옆에 붙게 한다. 마크업 순서·name·value 불변(`settings_operating_test.go:151`).

## 3. 깨질 수 있는 시험 (정찰 C, 편집 전 재확인)
`screen_paths_test.go:68,210` · `settings_cadence_test.go:54,279,332,390,460` · `settings_operating_test.go:151,157,291,313` · `status_strip_test.go:35-78,198` · `display_primitives_test.go:287,399,427,506,662` · `console_prose_test.go:29,234` · `settings_card_test.go:105,275` · `a111_exit_freshness_test.go:180-182` · `performance_history_test.go:53` · `overview_test.go`·`orders_test.go`·`portfolio_test.go` 의 한글 복사 단언(269줄). 원칙: 문장은 **지우지 않고 자리를 옮긴다**. 고정 리터럴을 바꿔야 하면 tasks.md 「고정 리터럴 시험 수정 기록」에 전수 기록.

## 4. 측정 (전후, 같은 하네스)
- `measure/shoot.cjs <base> <token> <outdir>`: 22개 화면 × 1280/390 × light/dark 스크린샷 + HTML + `{h, forms, inputs, buttons, notices, words, stripH, maxP}`.
- `measure/outline.py <outdir>`: 제목·폼·입력 개요 + 사유 문장 중복 수 + `a0NN`/경로 노출 수 + `.notice|.danger` 수.
- 기준선(2026-10-25, 빈 fixture): `/orders` 문장 5회, `/settings/daily` 상자 11, 높이 5008, `/dashboard` 4507·경로 3, `/verify-console` 경로 4, `/strategy-runtime` `not_observed` 59.
- 채운 fixture: `seedJournal` + `newOrdersHarness` + `brokerRows()` 로 `/positions`·`/orders`·`/dashboard`·`/history` 를 Go 시험 안에서 렌더해 같은 수치를 센다(빈 표본 통과 방지).

## 5. 인지 부하 목표 (cognitive-load-assessment 6항목)
외재 부하만 줄인다. 완료 시 22개 화면 표에서 결정·기억·읽기·시각 항목 "높음" 0. 새 개념 항목은 내재 부하(업무 자체)라 제외하고 a174~a176 이 이어받는다.


### §4 결과 (2026-10-25 구현자 측정, 격리 콘솔 38321, 빈 fixture, 데스크톱 1280/모바일 390)
전 화면 표: `analysis/measure-table.md`. 요약: 같은 원인 문장 `/dashboard` seam_unwired 13→1·journal_unreadable 5→1·`/orders` 4→1(시험이 6 fixture 로 상시 확인) / 떠 있는 상자 최대 11→3 / 상태 표시줄 113~203px→≤71px, 모바일 229~374px→≤82px / 본문 change ID(최적화 5~10)·절대경로(`/dashboard` 3·`/verify-console` 4) 전부 0 / 가장 긴 p 1152→672px. 미달: `not_observed` 25(→a175), 설정 높이 +4%.
