# a172 3단계 디자인 리뷰 (Impeccable 4.6.0, Operate 모드)

Method: ⚠️ DEGRADED: single-context (서브에이전트 도구 있음, 그러나 요청이 in-thread 실행을 지정해 A·B를 한 문맥에서 순차 실행). 편집 0. PRODUCT.md/DESIGN.md 없음 → 기존 코드 기준 진행.
증거: 소스(templates.go 1~330), 스크린샷 dashboard/positions/optimization/settings, 라이브 측정(빌드 `review-tossctl`, 포트 38180, GET만, 종료·`ps` 빈 출력 확인). 측정 스크립트 `review-measure.cjs`, `review-probe.cjs`, `review-probe2.cjs`(같은 폴더).
한계: 데이터 없는 상태라 보유 표(a058)·주문/승인/토글 버튼·approval 화면은 렌더 안 됨 → 해당 부분은 소스 판독만, "미측정"으로 표기.

## (c) 탐지기 원본 JSON
```json
[
  {"antipattern":"side-tab","name":"Side-tab accent border","severity":"warning","category":"slop",
   "file":".../internal/console/templates.go","line":90,"snippet":"border-left: 3px solid var(--warn-line)"},
  {"antipattern":"side-tab","name":"Side-tab accent border","severity":"warning","category":"slop",
   "file":".../internal/console/templates.go","line":91,"snippet":"border-left: 3px solid var(--bad)"}
]
```
탐지기 누락(수동 판독): 같은 side-tab 패턴이 두 곳 더 있음 — `templates.go:215` `.card-preview{border-left:3px solid var(--accent)}`, `templates.go:238` `.optimization-nav a.on{box-shadow:inset 3px 0 var(--accent)}`(그림자로 쓴 측면 띠). 탐지기는 border-left 단독·반경 없는 요소·box-shadow 변형을 못 잡음. craft-floor 「1px 초과 colored border-left 금지」 및 규칙 20 위반.

## (a) 비평 — Nielsen 10 휴리스틱 (Operate, 0~4)
| # | 휴리스틱 | 점수 | 핵심 |
|---|---|---|---|
| 1 | 상태 가시성 | 3 | 상태 스트립·알약·글자 병기(색 단독 아님). 단 `meta refresh` 5초 재로드(스크롤·포커스 소실, 백엔드 계약이라 "do not fix") |
| 2 | 현실 일치 | 3 | 한글 도메인 어휘, 내비 부제가 질문형. verify 오류가 영어 원문 그대로 노출(백엔드) |
| 3 | 사용자 통제 | 3 | 읽기 전용 + 명시 승인. 되돌리기 UX는 범위 밖 |
| 4 | 일관성 | 2 | 알림 3종(.notice/.danger/.card-preview)이 서로 다른 띠 문법; 버튼 한 종류(검정 채움)뿐이라 위험 동작과 일반 동작이 같은 모양 |
| 5 | 오류 예방 | 3 | 타이핑 확인 없음(계약 준수), 3초 대기. 시각적 위험 구분은 약함(4번과 동일 원인) |
| 6 | 인식>회상 | 3 | 내비 부제, 표 caption. 대시보드 표 5개 중 5개 caption 없음 |
| 7 | 유연성 | 2 | 6개 화면 단일 경로. 키보드 건너뛰기(skip link) 없음, 내비 6개+상태 4칸을 매 화면 지나야 본문 |
| 8 | 미니멀 | 2 | 대시보드 1280 높이 4392px, 같은 "구조적으로 미측정…" 문단 3회 반복(복사). 저밀도 프리미엄 브리프와 어긋남(문구는 백엔드/카피 영역) |
| 9 | 오류 복구 | 3 | 사유+조치 문구 있음. `.danger`가 role=alert로 선언돼 있음 |
| 10 | 도움말 | 2 | `.explain` 펼침 링크 존재, 단 `summary` 높이 17~19px로 터치 목표 미달(아래) |
총 24/40 (Acceptable 상단). 

**디자인 특이성 판정**: 부분 통과. 웜 스톤 + 단일 액센트 + 비대칭 반경은 보라 그라디언트/Inter 기본값에서 벗어났다. 그러나 한 가지 시각 어휘(좌측 띠 알림, 같은 면 카드의 연속)가 화면 전체의 위계를 평평하게 만든다. 대시보드는 섹션 11개가 동일 면·동일 패딩·동일 반경으로 세로 나열 = "기계적 대칭"이 반경 층위에서만 깨지고 구조 층위에서는 남음.

**상위 이슈(비평)**
1. 측면 띠 알림 4곳(탐지기 2 + 수동 2) — 규칙 20·craft-floor 위반의 유일한 객관 지표.
2. 위험/일반 버튼 동형 — 규칙 20은 "항상 채워진 누르는 것"을 요구하므로 *채움 유지*는 맞으나, 채움 색이 한 가지(`--ink-1`)라 청산·승인과 "필터" 버튼이 구분되지 않음. 색을 바꾸는 대신 형태 힌트(반경 조합, 라벨)로 해결해야 함. 미측정(버튼 미렌더).
3. 대시보드 위계 없음: h2 11개가 모두 `--text-2`, 섹션 면이 모두 `--surface-1`. 읽는 순서는 내용이 아니라 위치가 정함.
4. 모바일 설정 화면 9216px: 가장 긴 화면이 같은 카드 반복(스크롤 비용).
5. 호버/active/disabled 상태 규칙 사실상 0개(`:hover` grep 0건; `button:disabled`는 `.optimization-content`·`.sticky-save` 한정). 모션 4단계 전에 상태 어휘가 비어 있음.

**페르소나 적신호**: 1인 운영자(모바일 경유) — 375px 설정 화면에서 "적용 후" 값이 글자 단위로 쪼개져 읽을 수 없음(아래 F1).
**감정 곡선**: 평온한 톤은 맞음. 오류 구간(빨간 블록)이 길어 verify 원문 전문이 박스를 채움.

## (b) 감사 — 점수 (0~4)
| # | 차원 | 점수 | 핵심 |
|---|---|---|---|
| 1 | 접근성 | 3 | 대비 측정 이상 없음(아래), focus-visible 3px, aria-label·caption 일부. 건너뛰기 링크 없음, 입력칸 높이 < 44 |
| 2 | 성능 | 4 | 외부 요청 0, 단일 style, 전환/애니메이션 0, `style=` 속성 1개. 레이아웃 속성 애니메이션 없음 |
| 3 | 테마 | 3 | 토큰 + 다크 `:root` 재정의 일관. 하드코딩 rgba 그림자 4곳, 다크 그림자(`.row-details[open]`)만 규칙 하나가 덮어씀(주석 "개별 덮어쓰기 없음"과 불일치) |
| 4 | 반응형 | 2 | hscroll 0(전 20조합). 단 375px `.card-preview` dd 폭 37.6px(F1), 입력 37px |
| 5 | 구현 무결성 | 2 | side-tab 4곳, `.card-preview`가 모바일 `dl` 규칙을 이김(특이도), 인라인 `style` 1개(`templates.go:372`) |
총 14/20 (Good 하단).

**라이브 측정 (Playwright/Chromium, 5화면 × 1280/375 × light/dark)**
- 가로 스크롤: 20조합 전부 `scrollWidth - innerWidth = 0`.
- 텍스트 대비: 모든 텍스트 노드 계산값 대비 < 4.5:1 인 것 0건(양 테마·양 폭). 알약/상태 포함. (비텍스트 3:1은 `TestTradingViewsThemeColorsMeetWCAGInBothThemes`가 담당; 재측정 안 함)
- `h1` 화면당 1개, `<main>` 있음, `lang="ko"`, viewport 메타 OK.
- 터치 목표 < 44px (측정, 1280 / 375):
  - 본문 속 인라인 링크 `A` 15~19px — WCAG 2.5.8 인라인 예외, 감점 없음.
  - 비인라인 위반: `summary.explain`("집계 기준 펼치기" 119x19 / 105x17), 설정 `summary`("고급 — 개별 값 직접 기…" 1102x28), 설정 `input` 352x40(1280) · 316x37(375, `.hypothetical-fields input{min-height:2.5rem}` = 40px 하한), `.optimization-nav a` 1280에서 높이 19(내부 `span`만 측정됨 — 링크 박스 `min-height:56px`라 무시), 설정 탭 `A` 52x19(실제 박스는 44 → 내부 텍스트 측정; 무시).
  - 실제 위반으로 확정: `.explain summary`(17~19px, `.row-details summary`만 44px 규칙이 있고 `.explain summary`는 없음), 설정 입력칸 37~40px.
- 폰트 크기 분포(1280): 17px 본문, 13.6px(`--text--1`), 14.4px(표 0.9rem 고정) — 14.4가 스케일 밖(계약: 시험이 고정, 건드리지 말 것). 375: 12.48px 요소 17~33개(`--text--1` 최소 0.78rem). 보조 설명 12.5px는 읽을 수 있으나 하한 13px 권장.
- 좁은 열(375, 설정): `dl.card-preview` 열 정의 `224px 37.6px` — 값 칸(dd)이 37.6px, 최대 11줄(스크린샷 `rv-narrow` 대신 측정값). 같은 규칙이 main에도 있었음(`git show main` 178행) → 이번 단계가 만든 결함은 아니나 3단계 "양 테마 대비·타입" 범위에서 드러난 미해결.
- 대시보드 375: `.table-scroll` 안 표의 첫 열이 28~41px(시장/검증 진행) — 같은 표의 마지막 열 507px. `min-width:36rem` 표 안에서 열 폭 배분이 자동 레이아웃에 맡겨져 앞 열이 글자 단위로 줄바꿈.

**다크 모드**: 토큰 재정의만으로 동작. 스크린샷 optimization-1280-dark에서 `.notice` 띠(warn-line) 대비 충분, 상태 스트립 `--lift` 그림자는 다크에서 거의 안 보임(면 구분은 `--surface-1`/`--surface-0` 차 `#1d1c19`/`#141311`=약 1.1:1) → 카드 경계가 사실상 면 색에만 의존. 위계는 읽히나 저시력자에게 면 경계는 3:1을 못 충족(장식 경계이므로 위반은 아님, 관찰).

**모션 제외 대상(4단계 대비)**: 주문·승인·청산·토글 `button`/`.button-link`/`form`, `.sticky-save`(승인 바), 승인 안내 `.console-approval`, `meta refresh` 대상 화면의 모든 상태 변화. 이들은 전환 0, 동작은 즉시. 모션 후보: 탭/내비 `a.on` 전환, `details` 열림(`.row-details[open]` 모달형 — `position:fixed` 열림은 transform으로 애니메이션 하지 말 것: 이미 `translate(-50%,-50%)`를 위치에 쓰고 있어 `transform` 애니메이션과 충돌).
`prefers-reduced-motion` 규칙 아직 없음(전환 0이라 현재는 무해).

## (d) 수정 순위표

| 순위 | 위치 | 심각도 | 문제 | 구체 수정 | 근거 계약 | 시험이 고정? |
|---|---|---|---|---|---|---|
| 1 | `templates.go:90,91` (`.notice`,`.danger`) | major | 3px `border-left` side-tab. 탐지기 2건, craft-floor 금지 | 측면 띠 제거. `border:1px solid var(--warn-line)`(danger는 `--bad-line`)로 전방 균일 테두리 + 기존 `--*-bg`. 의미는 이미 `<strong>`/텍스트가 운반 | 규칙 20, craft-floor "border-left>1px 금지" | `console_prose_test.go`는 `.notice` 클래스 위치만 봄(CSS 아님). 색 대비 시험이 `--warn-line`/`--bad` 3:1을 이미 잼. 신규: CSS에 `border-left:3px` 부재 시험 권장 |
| 2 | `templates.go:215` (`.card-preview`) | major | 3px accent `border-left`, 탐지기 누락. 더불어 `padding-left` 0.8rem로 값 칸이 줄어듦 | `border-left` 제거, 면(`--surface-2`)+`--radius-soft`로 구분(이미 `.tier-card` 어휘). 같은 요소 215행 padding을 `var(--space-2) var(--space-3)`로 | 규칙 20 | `settings_card_test.go:88`, `settings_limits_test.go:164`는 `class="card-preview"` 문자열만 단언 — 클래스 유지하면 안전 |
| 3 | `templates.go:238` (`.optimization-nav a.on`) | minor | `inset 3px 0` = 그림자 형 side-tab, 탐지기 누락 | `inset 0 -2px var(--accent)`(설정 탭 209행과 동일 어휘)로 통일 | 규칙 20·일관성 | 미고정 |
| 4 | `templates.go:214` (`.card-current, .card-preview`) + 모바일 블록 `:278~` | major | 375px에서 `grid-template-columns: minmax(8rem,14rem) 1fr`이 모바일 `dl` 규칙(특이도 낮음)을 이겨 dd = 37.6px, "적용 후" 값이 11줄로 쪼개짐. 승인 직전 값 비교 화면이라 읽기 실패는 안전 UX 문제 | 모바일 미디어 블록에 `.card-current,.card-preview,.entry-list{grid-template-columns:minmax(0,1fr)}`(dt 위·dd 아래 한 열) 또는 `minmax(6.5rem,40%) minmax(0,1fr)` | 스펙 `data-label` 카드 표 의도, 375 가독성 | 미고정(신규 375 측정 시험 권장) |
| 5 | `templates.go:231` `.hypothetical-fields input` + `:93` `input[type=text]` | major | 입력칸 높이 37~40px(< 44). 스펙 44px 터치 목표 위반(측정 316x37) | `min-height:44px`로 상향(2.5rem→44px), 텍스트 입력에도 `min-height:44px` | 스펙 44px | 대비 시험은 안 잼. 미고정 |
| 6 | `.explain summary` (`templates.go:~146`) | minor | 터치 목표 17~19px(측정). `.row-details summary`만 44px | `.explain summary{min-height:44px;display:flex;align-items:center}` | 스펙 44px | 미고정 |
| 7 | `templates.go:~240` `.table-scroll table{min-width:36rem}` + 대시보드 표 5개 | minor | 375px 스크롤 영역 안 첫 열 28~41px로 글자 단위 줄바꿈. 표에 `<caption>`·`data-label` 없음(스펙 모바일 카드 표 계약은 `.data-table` 한정이라 이 표들은 적용 대상 아님 — 관찰 유지) | 첫 열에 `white-space:nowrap`, 또는 `th:first-child{min-width:4rem}` | 가독성 | 미고정 |
| 8 | `templates.go:~105` `button` 전역 | minor | 위험 동작·일반 동작 동형(검정 채움 1종), `:hover`/`:active`/`button:disabled` 전역 규칙 없음 | 색을 바꾸지 말고(규칙 20 "항상 채워진 버튼") 상태 규칙만 추가: `button:disabled{opacity:.65;cursor:not-allowed}` 전역화, hover는 명도 단계만(`filter` 금지, 지연 0). **지연·전환 금지**(안전 불변식 4) | 규칙 20, 안전 불변식 4 | 미고정. 미측정(버튼 미렌더) |
| 9 | `templates.go:372` 인라인 `style="margin-left:auto"` | nit | CSS 정본 밖 스타일, 모바일 `!important`로 되돌려야 함(281행) | `.header-note{margin-left:auto}` 클래스로 이동하고 `!important` 제거 | 정본 CSS 단일화 | 템플릿 문자열 단언 시험 있으면 갱신 필요(미확인) |
| 10 | `templates.go:~302,321~` 다크 `.detail-cell .row-details[open]{box-shadow…}` | nit | "개별 규칙 색 덮어쓰기 없음" 주석과 불일치 | `--shadow-pop` 토큰을 `:root`/다크 `:root`에 정의하고 규칙은 변수만 읽게 | 규칙 20 토큰 원칙 | 미고정 |
| 11 | 다크 `--lift` | nit | `0 1px 2px rgba(0,0,0,.4)` — 면 차 1.1:1이라 카드 경계가 거의 안 보임 | 다크에서 `--surface-1`을 소폭 올리거나(`#1f1e1b`) `--line-soft` 1px 윤곽 추가. 대비 시험 재실행 필요 | 규칙 20 저밀도 프리미엄 | 대비 시험이 면 대비를 재므로 재실행 필수 |
| 12 | `.status-strip`, `.console-summary` 섹션 11개 위계 | minor | 모든 h2·섹션이 같은 크기·면. 위험 상태(엔진 정지·미측정)와 정보 섹션의 시각 차 없음 | 상태 스트립만 `--surface-1`, 정보 섹션은 `--surface-0`에 구분선만(면 층 2단계), 첫 섹션 h2 한 단계 키움 | 규칙 20 비대칭·저밀도 | 미고정 |

**do not fix (고정 계약과 충돌)**
- `.positions-table` 12/10/11px·8px 12px 패딩, `.line-item` 10~12px, `.number-column` — `a058_stockos_density_test.go`·스펙 146/158행이 고정. 12.48px 보조 글자 크기 지적도 이 표에는 적용 안 함.
- `table{font-size:0.9rem}`(14.4px), `.safety-line{font-size:1rem}` — 시험이 리터럴을 고정(console-design.md 명시).
- `meta http-equiv="refresh"`(5초) — 백엔드 계약. WCAG 2.2.1은 관찰만.
- 영어 verify 오류 원문, 반복되는 "구조적으로 미측정" 카피 — 백엔드/카피 범위(범위 밖).
- 버튼을 흐리게 하거나 색을 달리해 위험을 표시하는 방식 — 규칙 20·안전 불변식 4.

**4단계(모션) 주의**: `.detail-cell .row-details[open]`는 `transform`을 위치 지정에 이미 사용 → 열림 애니메이션은 `opacity`만, 또는 `translate`/`scale` 개별 속성으로. 승인·청산·토글 버튼 및 `.sticky-save`는 전환 0 유지. 상태 변화(데이터 갱신)에 애니메이션 금지.

사후: 측정용 콘솔(포트 38180) 종료, `ps -eo pid,args | grep -E '[c]onsole --port 38180'` 출력 없음 확인.
