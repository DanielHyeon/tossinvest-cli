# 콘솔 디자인 시스템

정본은 `internal/console/templates.go` 의 `{{define "style"}}` 이다. 이 문서는 그 요약이며 어긋나면 코드가 맞다.
범위는 콘솔 화면(`internal/console/`)뿐이고 주문·전략 경로에는 쓰지 않는다. change `a172`, 규칙 20(`.claude/CLAUDE.md`).

## 원칙
- 색·여백·크기는 `:root` 변수에서만 읽고, 다크 테마는 `prefers-color-scheme: dark` 블록의 `:root` 변수 재정의로만 만든다.
- 웜 스톤 중성 + 액센트 하나(`--accent`) + 채도를 낮춘 의미색. 그라디언트·웹폰트·Tailwind/Material 기본 팔레트는 쓰지 않는다.
- 정보 섹션은 상단 구분선 하나만 두고, 떠 있는 면(`--surface-1` + `--lift`)은 먼저 읽어야 하는 상태 스트립·경고·승인 바에만 준다. 측면 3px 띠(side-tab)는 금지(시험 `TestNoSideTabAccentBorders`).
- 보유 표(a058: 12px/18px, 8px 12px)와 44px 터치 타깃은 고정이다. 그 밖의 여백은 `clamp()` 유동 값이다.

## 색 토큰 (라이트 / 다크)
| 토큰 | 라이트 | 다크 |
|---|---|---|
| `--surface-0` 페이지 | `#f5f3ee` | `#141311` |
| `--surface-1` 묶음 | `#fcfbf8` | `#201f1b` |
| `--surface-2` 안쪽 홈·입력칸 | `#ebe8e1` | `#11100e` |
| `--line` / `--line-soft` 장식 구분선 | `#d9d5ca` / `#e7e3da` | `#38362f` / `#2a2924` |
| `--line-strong` UI 경계(3:1) | `#857f73` | `#756f64` |
| `--ink-1` → `--ink-3` 본문 → 흐린 설명 | `#201e1a` `#47433b` `#625d54` | `#ece9e1` `#cbc7bb` `#a49f93` |
| `--accent` 링크·포커스 링 | `#2e5a86` | `#9fbbd8` |
| `--ok` / `-line` / `-bg` | `#2f6a40` `#5f8f6c` `#e6eee3` | `#93c7a0` `#5a9068` `#1b2a1f` |
| `--warn` / `-line` / `-bg` | `#7f540f` `#a9761e` `#f8eed9` | `#e0b765` `#a97b2d` `#2a2314` |
| `--bad` / `-line` / `-bg` | `#9a3228` `#b4503f` `#f6e6e1` | `#ea9a90` `#c25f52` `#2d1a17` |

대비 기준(시험 `TestTradingViewsThemeColorsMeetWCAGInBothThemes`가 두 테마를 모두 잰다): 텍스트 토큰은 세 면 위에서 4.5:1,
잉크는 의미색 바탕 위에서도 4.5:1, `--line-strong`·`--accent`·의미색 `-line` 은 3:1 이상이다.

## 타입 스케일 (모듈러, 좁은 화면 1.2 → 넓은 화면 1.25, 각 단계 clamp)
| 토큰 | 최소 → 최대 | 쓰임 | 줄 간격·자간 |
|---|---|---|---|
| `--text--1` | 0.78 → 0.85rem | 설명·레이블·알약 | 1.3~1.5, 0(한글) |
| `--text-0` | 0.9375 → 1.0625rem | 본문(최소 15px) | 1.65, 0 |
| `--text-1` | 1.125 → 1.33rem | h3 | 1.35, -0.004em |
| `--text-2` | 1.35 → 1.66rem | h2 | 1.28, -0.012em |
| `--text-3` | 1.62 → 2.07rem | h1 | 1.2, -0.02em |

숫자 열은 `font-variant-numeric: tabular-nums` 를 유지한다. 시험이 리터럴을 고정한 규칙(`table` 0.9rem, `.safety-line` 1rem, 보유 표 px)은 스케일 밖이다.

## 여백·반경
- 여백 `--space-1` … `--space-5` 는 clamp 유동 값이다.
- 반경은 역할별 세 조합이고 네 모서리가 다르다: `--radius-card`(0.9 0.35 0.8 0.3rem, 묶음·카드·모달·모바일 표 행), `--radius-control`(0.35 0.7 0.35 0.6rem, 버튼·입력칸·탭), `--radius-soft`(0.25 0.5 0.3 0.45rem, 알림·코드·안쪽 칸). 알약(`.status-pill`)만 999px.

## 금지 목록 (규칙 20)
- 보라 그라디언트, Inter 등 웹폰트 도배, 기계적 대칭(같은 반경·여백 복사), 고정 px 만으로 짠 스케일.
- 주문·토글·승인·청산 버튼을 장식으로 흐리거나 전환으로 늦추는 것(안전 불변식 4). 버튼은 항상 채워진 「누르는 것」으로 보인다.
- 접근성(포커스 링 `--accent` 3px·offset 2px, 44px 타깃)과 모션 성능 규칙이 미학보다 앞선다.

## 모션 (4단계, 순수 CSS)
- 토큰: `--dur-1/2/3` = 120/220/300ms, `--ease-spring`(14점 `linear()` 감쇠 스프링, 약 29% overshoot, nav·눌림), `--ease-pop`(12점, 모달 열림 전용, 약 12%), `--ease-settle`(10점, 약 5%). `linear()` 미지원이면 `@supports` 바깥 기본값 `ease-out`. 움직임 크기(`--nudge --press --pop-from`)도 토큰.
- 6단계 앰비언트(순수 CSS, JS·CSP 불변): `body > header`(`overflow: clip`)·`.console-status` 의 `::before/::after` 정적 radial-gradient 덩어리(크기 `min(42%, 30rem)`×`min(170%, 22rem)`, 색 `--tint-*` 전용 틴트를 테마별 `--tone-alpha` 라이트 19%·다크 15% 로 섞음)가 `translate/rotate/opacity` 로만 34/41초 루프. 엔진·데이터 톤을 따르는 **보조 신호 — 라이트·다크 모두 톤 구분 가능(ΔE≥12), 정본은 상태 단어**이고 색만으로 상태를 전하지 않는다(글자 대비 4.5 유지, `TestAmbientToneLegibility`). 톤 매핑: ok=실행+신선, warn=정지·지연, bad=stale, 그 외 `--accent`. `data-reload=on`(자동 재로드 화면)에서는 정지, reduced-motion 에서 정지, reduced-transparency 에서 덩어리·hover 조명 제거. hover 조명(`.optimization-nav a`·`.settings-tabs a` 의 `::after`, `opacity` 페이드)은 읽기 전용 이동 링크에만 걸고 POST 폼을 품은 카드(`.tier-card`·`.choice-tile`)에는 걸지 않는다(제외 선택자는 승인 영역 방어). 시험 `TestAmbientMotionContract`(렌더 화면 기준).
- 대상: 상단 nav·설정 탭·최적화 nav(hover/focus 시 `translate`), POST 폼 밖의 일반 `button`/`.button-link`(`:active` 에서 `scale`), 상세 모달 열림(`opacity` + 개별 `scale`, 위치용 `transform` 은 그대로)과 그 뒤 스크림(셀 `::before`, `opacity` 만, `pointer-events: none`).
- 제외(전환 0, 안전 불변식 4): `form:not([method=get])`·`form button`·`form input[type=submit]`·`[formmethod]`·`[formaction]` 로 잡히는 모든 버튼(주문·승인·청산·토글·엔진 시작/정지·재시작 등 모든 POST 폼), `.sticky-save`, `.approval-actions`, `.confirm-check`, `.console-approval`. 데이터 값이 바뀌는 자리에는 애니메이션 없음.
- 성능: 전환 속성은 `translate/scale/opacity` 와 hover 색(`background-color/color`, ≤ `--dur-1`)뿐. `all`·레이아웃 속성·`box-shadow`·blur 금지, `will-change` 는 열린 모달 1곳. 시험 `TestMotionContract`.
- `prefers-reduced-motion: reduce`: 시간 토큰을 0.01ms, 움직임 토큰을 0/1 로 리셋.

## 정보 구조 규칙 (change a173)

- **원인 1회**: 값을 읽지 못한 원인(`unmeasuredReason`)의 설명 문장은 화면당 한 번 `causes` 배너(상자 하나)에만 쓴다. 각 칸은 `미측정 · <사유 코드>`(`.muted.unmeasured` + `<code>`)만 남기고 구체 사유는 칸 `title`·배너 목록에 둔다. 새 칸을 만들면 뷰 모델에 `reading` 필드(내보내짐)로 두어야 배너가 훑는다.
- **알림 3단계**: `.danger` = 차단(읽기 실패·기동 거부), `.notice` = 결과가 따르는 사실(승인 대기·구버전 바이너리·저장 결과·비가역), `.aside` = 상자 없는 본문(경계·설명·빈 상태). 반영 시점은 저장 버튼 옆 `muted effect-timing` 한 줄. 한 화면의 떠 있는 상자 ≤ 3. 행 단위 표시(상태 알약·보조 지표)에는 `.notice`/`.danger` 를 쓰지 않는다(`status-pill warn`, `submetric bad|warn`).
- **경계 배지**: `<h1>` 바로 뒤 `<span class="status-pill scope">읽기 전용|편집 가능</span>`, "이 화면은 ~하지 않는다" 류 경계 문단은 `<details class="scope"><summary>이 화면의 경계</summary>` 하나. 배지는 `<h1>` 밖.
- **기술 정보**: change ID·절대경로·digest·SHA 는 `{{template "tech" 값}}`(`<details class="tech">`)로 접는다. 값은 DOM 에 남는다. 안전 상태(UNWIRED·OFF·미충족·stale)는 접지 않는다. `.notice`/`.danger` 는 `<details>` 안에 두지 않는다.
- **상태 표시줄**: 칸 설명은 한 줄(≤40자), 긴 사유는 `title`. 실패는 `읽지 못함 · <사유 코드>`. 데스크톱 ≤72px, 좁은 화면 칩 묶음 ≤120px.
- **산문 폭** `max-inline-size: 42rem`(표·폼·그리드 제외), 섹션 제목은 한글 우선(영문 식별자는 `<code>`), "적용 후/실행되는 것" 은 저장 버튼 옆 한 문단(`p.card-preview`), 비교 행이 있는 미리보기만 `dl.card-preview`.
