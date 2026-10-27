# a172 4단계(모션) 디자인 리뷰 — HEAD 6f1546f93

방법: Impeccable 4.6.0 Operate(critique·audit·optimize·craft-floor·operate 참조 읽음, 단일 문맥 순차 실행 — 하위 에이전트 미사용) + avoid-ai-design 0.4.0 detect-only + hallmark sloplint audit. 편집 0(`git status` 깨끗). PRODUCT.md/DESIGN.md 없음 → 코드 기준.
증거: 소스 `templates.go` 17~395, 라이브 콘솔(빌드 `review4-tossctl`, 포트 38190, GET만, 5화면 × 1280/375 × light/dark 측정), 스크린샷 `review4/shots/*.png`, `stage4/modal-open-1280-light.png`. 측정 스크립트 `review4-measure.cjs`, `review4-dump.cjs`. 덤프: `review4/html/*.html`, `review4/console.css`.
한계(미측정): 데이터 없는 콘솔이라 보유 표(a058)·주문/승인/청산 버튼·`.sticky-save`·`.approval-actions`·실제 상세 모달은 렌더 안 됨 → 소스 판독 + 4단계 주입 DOM 스크린샷만. 프레임 수치는 이번에 재측정하지 않음(`frames.cjs`는 읽기만). hallmark `--render` 생략(puppeteer 없음). 터치 합성 제스처 미시험.

## 3단계 수정 12건 중 무작위 3건 검증 (반복 지적 금지)
- #1/#2/#3 side-tab: `console.css`에 `border-left`는 `.console-status > div` 1px 구분선뿐(179~180), `inset 3px` 0건. 탐지기도 side-tab 0건 → 닫힘 확인.
- #4 375px `.card-preview`: 측정 `cardPrev = 299.8px` 한 열(1280은 `224px 840px`) → 닫힘 확인.
- #6 `summary` 44px: 전역 `summary{min-height:44px}`(templates.go:163) 존재, 375·1280 전 화면 `small` 목록에 summary 0건 → 닫힘 확인. (#5 입력칸도 text input 44px로 측정에서 사라짐.)

## (a) Impeccable

### 탐지기 원본 JSON (`detect --json templates.go templates_optimization.go`, rc=0)
```json
[
  {
    "antipattern": "bounce-easing",
    "name": "Bounce or elastic easing",
    "description": "Bounce and elastic easing feel dated and tacky. Real objects decelerate smoothly — use exponential easing (ease-out-quart/quint/expo) instead.",
    "severity": "advisory",
    "category": "slop",
    "file": "/mnt/D/Axipient/workspace/TossOS/worktree/internal/console/templates.go",
    "line": 375,
    "snippet": "animation: var(--ease-spring)",
    "advisory": true
  }
]
```
판정: 규칙 20 "스프링 탄성은 사용자 요구"와 충돌하는 advisory → **do not fix**(제거 요구 안 함). 대신 튜닝 수치를 아래 M1에 둠.

### critique — Nielsen 휴리스틱 (Operate, 0~4)
| # | 휴리스틱 | 점수 | 핵심 |
|---|---|---|---|
| 1 | 상태 가시성 | 3 | 상태 스트립·알약·글자 병기. `meta refresh`는 백엔드 계약 |
| 2 | 현실 일치 | 3 | 한글 도메인 어휘. 영어 verify 원문 노출(백엔드, 범위 밖) |
| 3 | 사용자 통제 | 3 | 읽기 전용 + 명시 승인 |
| 4 | 일관성 | 3 | 알림 두 종(`.notice`/`.danger`)이 1px 전방 테두리+배경으로 통일됨(3단계 대비 +1). 단 `.card-preview`/`.tier-card`/`.choice-tile`이 같은 면(surface-2)이라 역할 구분은 여전히 위치 의존 |
| 5 | 오류 예방 | 3 | 타이핑 확인 없음(계약 준수). 설정 토글 체크박스 13px 표적(아래 F1) |
| 6 | 인식>회상 | 3 | 내비 부제. 대시보드 표 caption 없음 |
| 7 | 유연성 | 2 | skip link 없음. 내비 6개+상태 4칸을 매 화면 지나야 본문 |
| 8 | 미니멀 | 2 | 대시보드 1280에서 4651px, 375에서 6365px. 같은 "구조적으로 미측정…" 문단 3회 반복(카피, 범위 밖). 저밀도 브리프와 여전히 어긋남 |
| 9 | 오류 복구 | 3 | 사유+조치 있음 |
| 10 | 도움말 | 3 | summary 44px로 해소 |
총 **25/40** (3단계 24 → +1). 미니멀·유연성이 발목.

디자인 특이성(Operate 기준): 통과. 웜 스톤 + 단일 청색 액센트 + 네 모서리 다른 반경 + 면 층(0/1/2) + 시스템 글꼴. 보라 그라디언트·Inter·글래스 없음. 약점: 대시보드 11개 섹션이 같은 `h2`(--text-2)·같은 구분선 → 위계는 위치가 정함(3단계 #12 미반영, 아래 R9).

모션 비평(브리프 기준):
- 스프링 근사는 의도대로 작동. `--ease-spring` 정점 1.287(21%) = **28.7% 오버슈트**. 적용처가 `translate:-2px`(내비 hover, 오버슈트 0.57px), `scale .97→1`(버튼 press, 정점 1.009), `pop-in scale .96→1`(모달, 정점 1.0115)라 **절대 크기가 작아 눈에 띄는 튕김은 모달 열림 한 곳뿐**. 28.7%는 "탄성이 느껴짐" 하한을 넉넉히 넘기고, 모달(읽기 전용 상세)에는 과함. 권고 M1: 모달만 `--ease-settle`(정점 1.045, 4.5%) 계열로 바꾸거나 `--ease-spring` 정점을 1.287→1.10~1.15(약 10~15%)로 낮춤. nav·버튼은 현행 유지 가능. 제거 요구 아님.
- `--dur-3: 360ms`는 operate.md "150~250ms" 권고 초과. 읽기 전용 모달 열림이라 안전 불변식 4 무관하나, 닫기·승인 흐름은 아님. 권고: 모달 열림 280ms(`--dur-3`을 280~300ms).
- 제외 목록 구성 양호: 승인·청산·토글 POST 폼, `.sticky-save`, `.approval-actions`, `.confirm-check`, `.console-approval` 전부 `transition:none`(367~368)으로 못 박힘. 사이트 전체에서 `form`이 `method=post`인 7개 폼(설정)이 제외 선택자에 걸림 → 측정에서 전환 속성 보유 요소는 `nav a`(.on 포함)뿐(trans 목록 `A.`, `A.on` 2종). 즉 버튼 전환은 `method=get` 폼/폼 밖 `button.secondary`류에만 걸리고 이 화면들엔 렌더 0건 — 미측정.

### audit (코드 수준) — 점수 0~4
| # | 차원 | 점수 | 핵심 |
|---|---|---|---|
| 1 | 접근성 | 3 | focus-visible 3px(296~298). `prefers-reduced-motion`은 시간·거리 토큰을 0으로 리셋(378) — "0.01ms 전체 kill"이 아니라 상태 전환(색)은 0.01ms로 즉시, 정적 위계 유지. 허용 범위. 결함: 설정 토글 라벨/체크박스 표적 13~19px (F1), skip link 없음 |
| 2 | 성능 | 4 | 외부 요청 0, CSS 24.6KB 단일, 전환 속성은 translate/scale/opacity/색뿐(레이아웃 속성·all·blur 0), `will-change`는 열린 모달 한 곳(상시 아님) |
| 3 | 테마 | 4 | 토큰 + 다크 `:root` 재정의 일관. 모달/바 그림자 토큰화(`--shadow-pop/-bar`) 완료. 하드코딩 rgba는 토큰 정의 안에만 |
| 4 | 반응형 | 3 | hscroll 0 (20조합 전부). 375에서 12.48px 요소 13~33개(`--text--1` 하한 0.78rem). 표적 위반은 F1만 |
| 5 | 구현 무결성 | 3 | side-tab 0, 인라인 style 0, 클래스 일관. 잔여: `.choice-tile{transition}`(370)은 서버가 정적으로 그린 상태라 전환이 일어날 값 변화가 없음 → 죽은 규칙(R10) |
총 **17/20** (Good 상단; 3단계 14 → +3).

측정(Playwright/Chromium, 5화면 × 1280/375 × light/dark): hscroll 0; 본문 폰트 17px/15.02px, 보조 13.6/12.48px, 표 14.4px(시험 고정), h1 33.1/26.5px; 비인라인 44px 미만 요소는 설정 화면의 체크박스 라벨만(아래 F1).

### optimize (UI 성능)
- 병목 없음(측정 근거: 외부 요청 0, 단일 인라인 style, 이미지·JS 0, 레이아웃 속성 애니메이션 0). 최적화 대상으로 확정할 항목 0 → "손대지 말 것".
- 관찰(미측정, 하향 권고): ① `body *`가 아닌 `:where(...)` 복합 선택자 7개 목록(360~365)이 모든 `button` 매칭 시 `:where(form:not(...) *)`를 평가 — 렌더 버튼 수십 개 규모라 비용 무시 가능. ② 대시보드 1280에서 4651px 단일 문서 — `content-visibility:auto`를 `section`에 주면 오프스크린 렌더 절감이 가능하나 `meta refresh` 5초 재로드·스크롤 위치 소실과 `contain-intrinsic-size` 점프 위험이 있어 이득 불확실 → **측정 전 채택 금지**. ③ 모달 `will-change`는 열린 동안만 적용되고 `animation-fill-mode:both`라 종료 후에도 `will-change`가 남음(열려 있는 동안이므로 의도와 일치, 닫으면 요소 자체가 `position` 규칙에서 빠져 해제) — 문제 없음.

## (b) avoid-ai-design 0.4.0 (detect-only)
원본 JSON: `review4/aad.json` (scanned 6, P0 0 / P1 4 / P2 5). 요약(행 = 파일·규칙·행 번호·노트; html 5개는 모두 같은 인라인 CSS를 포함):
```json
{"tool":"avoid-ai-design","version":"0.4.0","scanned":6,"summary":{"P0":0,"P1":4,"P2":5,"total":9},
 "findings":[
  {"file":"dashboard.html","id":"SD5","severity":"P1","name":"One word accented in the headline","lines":[471],"notes":["시장별"]},
  {"file":"dashboard.html","id":"SD4d","severity":"P2","name":"Template chrome: meta strings joined with middle dots","lines":[381,383,407,555,556]},
  {"file":"optimization.html","id":"SD5","severity":"P1","lines":[404],"notes":["전략 파라미터"]},
  {"file":"optimization.html","id":"SD4d","severity":"P2","lines":[381,383,410,412,446,450,489,495,504]},
  {"file":"orders.html","id":"SD4d","severity":"P2","lines":[380,382]},
  {"file":"positions.html","id":"SD4d","severity":"P2","lines":[381,383]},
  {"file":"settings.html","id":"L4","severity":"P1","name":"Round-number stat strip","lines":[488]},
  {"file":"settings.html","id":"SD5","severity":"P1","lines":[402],"notes":["당일"]},
  {"file":"settings.html","id":"SD4d","severity":"P2","lines":[381,383,403,497,529,530,563,596,629,659,660]}
 ]}
```
수동 판정(헤더 마크업 실물 대조):
- SD5 ×3(`<h1>설정 <span class="muted">당일</span></h1>`, `<h1>최적화 <span class="muted">전략 파라미터</span></h1>`, `<h2>계좌 보유 <span class="muted">시장별</span></h2>`): **진짜**. 한 단어/한 구절만 옅은 색(--ink-3)으로 낮추는 부제 처리는 표지 문법. 단 "강조"가 아니라 보조 한정어를 *낮추는* 방향이고, 같은 마크업을 모든 화면 제목이 공유(일관성)하며 규칙 20이 금지한 악센트 단어 아님. 심각도 nit. 수정: 부제를 `<small class="page-sub">`로 줄 분리 or 유지(R12).
- L4(`미설정 → <strong>1%</strong>`): **오탐**. 임계값 표시이지 지표 띠 아님.
- SD4d(`상시 · 당일 · 전략 · 도구`, `안전 검증 상태 · 잔여물 · Guardian 한도`): 내비 부제·섹션 제목의 가운뎃점 결합은 **카피(정보 구조 이름) 범위** — 화면 이름 조합이라 "stock chrome"으로 볼 소지는 있으나 콘솔 정보 모델의 일부 → 수정 요구 안 함(nit 관찰).

시각 tell 리뷰(스크린샷: dashboard 1280 light/dark, modal-open, optimization):
1. 첫 시야가 상태 스트립(4칸) → 같은 면 카드 반복. "카드 + 카드" 구조는 면 층 없이 같은 `--surface-1`/`--lift`라 둔탁. 3단계 #12 위계 제안(섹션 면 2단계)이 여전히 미반영. minor.
2. `.optimization-nav a` 목록(5개 동형 카드 + 소제목): 같은 크기 카드 연속 = 규칙 "Same-size cards" 경향. 그러나 내비라는 기능이 있어 허용. nit.
3. `.optimization-steps` 3칸 동일 타일, `.choice-grid` 3칸 동일 타일: 같은 크기 3열. 데이터 비교 UI라 허용하나 `.choice-tile` 면=`--surface-1`, `.optimization-steps li`=`--surface-2`로 색이 갈려 있고 타일 내부 구조도 다름 → 판정 통과.
4. 다크 `--surface-1 #201f1b` vs `--surface-0 #141311`: 카드 윤곽은 `--lift` 안 `0 0 0 1px var(--line-soft)`로 보강됨(3단계 #11 반영) → 다크에서 카드 경계 보임. 통과.
5. 모달: 스크린샷에서 **스크림(배경 흐림/어둠) 없음**. 모달 뒤 본문과 같은 밝기라 "떠 있음"이 그림자 한 겹에만 의존. 읽기 전용이라 안전 무관. 미려함·집중 측면에서 minor(R6).
6. 모달 헤더: `상세· 닫기` — `::after{content:" · 닫기"}`가 `summary` 텍스트("상세")에 공백 없이 붙고 앞 가운뎃점이 다른 글자와 달라붙어 보임(스크린샷 `상세· 닫기`). nit(R11).
7. 글래스·그라디언트 텍스트·이모지·hard offset shadow·보라: 0건. 모노스페이스는 `code`/입력/`pre`/오류 원문에 한정(허용).

## (c) hallmark audit (`sloplint.mjs <dir> --scope app --json`, `--tier floor`)
원본 JSON: `review4/hm-app.json`(rc=1; fails 45, reflex 18, warns 21), `review4/hm-floor.json`(rc=1; fails 45, reflex 0, warns 0). 렌더 검사 off(puppeteer 없음). 중복(html 5개 + css 1개) 접어 그룹화한 원본 발췌:
```
19 FAIL floor  dash "—" in copy                             dashboard:391 optimization:391 orders:390 positions:391 settings:391
20 FAIL floor  no Hallmark stamp on first CSS line          console.css:1 + html 5
34 FAIL floor  width:min(42rem,calc(100vw - 2re…) on .detail-cell .row-details[open] (scrollbar overflow)   console.css:148
34 FAIL floor  width:calc(100vw - 1.5rem) on .detail-cell .row-details[open] (scrollbar overflow)           console.css:297
34 FAIL floor  no overflow-x:clip on html and body          html 5
63 FAIL floor  middle-dot meta string (nav 부제·섹션·설정 라벨 6종)
67 FAIL floor  "종결 주문 CLOSED 그룹 한 페이지" typed in capitals at 12px   orders:439 ; settings:497 "[KRW] 주문 100주 …"
24b REFLEX     gap:3px 5px (.holding-verdicts) console.css:117 ; gap:5px (.line-item) :121 ; padding:1px 6px (.positions-table .status-pill) :127
41 WARN floor  dark surface (L 17%) on .card-preview / .tier-card sets no text colour   console.css:205, 207
49 WARN floor  long clickable label without nowrap (settings:496,529,595,628)
F23 WARN finish no meta description (5 pages)
```
판정(Hallmark 다이얼은 마케팅·신규 설계용이라 operate 콘솔에 맞게 걸러 읽음. 재설계 안 함):
- 19 em dash ×5(`advisory — 요일과 시각만…`): **카피**(백엔드 문구). 범위 밖, 수정 요구 안 함.
- 20 Hallmark stamp: **무관**(Hallmark 산출물 표지). do not fix.
- 34 `100vw` ×2: `.row-details[open]`는 `position:fixed`라 문서 스크롤 너비에 안 걸림 → **오탐**(측정 hscroll 0 전 조합, 모달은 주입 DOM 스크린샷으로 확인). 그러나 데스크톱에서 스크롤바가 있을 때 `100vw`는 스크롤바 폭을 포함해 모달 중심이 `left:50%` 기준 ~8px 어긋날 수 있음 → minor, `width:min(42rem, calc(100% - 2rem))`는 fixed 기준이 뷰포트라 `100%`가 동일 효과(R7).
- 34 `overflow-x:clip` on html/body: 현재 hscroll 0 + `body{overflow-wrap:anywhere}`로 이미 안전. 방어선 추가일 뿐. 저비용 권고 nit(R13).
- 63 middle-dot ×6: 정보 이름 결합(카피). nit, 수정 요구 안 함(avoid-ai-design SD4d와 동일 건).
- 67 대문자 12px: `CLOSED`/`KRW`는 도메인 약어(한글 환경). 오탐 성격. do not fix.
- 24b REFLEX ×3: `.holding-verdicts`·`.line-item`·`.positions-table .status-pill`은 `a058_stockos_density_test.go`·스펙 146/158 고정 리터럴 영역 → **do not fix**(계약).
- 41 dark 면 텍스트 색 미설정: `.card-preview`/`.tier-card`는 `body{color:var(--ink-1)}` 상속이 다크 토큰으로 바뀜(측정 대비 이상 없음, 3단계) → **오탐**.
- 49 긴 클릭 라벨: 설정 카드 라벨이 한글로 길지만 줄바꿈 허용이 더 안전(375 hscroll 0). do not fix.
- F23 meta description: 로그인 후 로컬 콘솔, 검색 비노출. do not fix(범위 밖).
floor 결론: 실질 결함 0, 방어선 권고 2(R7·R13). 마케팅성 floor 항목(19·20·63·67·F23) 전부 operate 콘솔에서 해당 없음/카피 범위.

## (d) 병합 수정 순위 (단일 목록)
기준 라인 = `internal/console/templates.go`(CSS 줄은 위 읽은 판) 및 `templates_settings.go`. 번호 R = 순위.

| 순위 | 위치 | 심각도 | 문제 | 구체 수정 | 근거 계약·브리프 | 시험 고정? | 제기 도구 |
|---|---|---|---|---|---|---|---|
| R1 | `templates_settings.go:~106` 주변 설정 토글 라벨(`name="place"/"sell"/"cancel"/"enabled"` 등)·`templates.go` 신규 규칙 | **major** | 설정 화면 체크박스 13x13, 라벨 높이 17~19px(실측 1280 `LABEL 92x19`, 375 `84x17`; 총 8개). 스펙 44px 터치 목표 위반. 토글 상태를 바꾸는 입력이라 오탭 위험보다 *맞추기 어려움*이 문제(안전 즉시성 영향 아님) | `label:has(> input[type=checkbox]){display:flex;align-items:center;gap:var(--space-2);min-height:44px}`, `input[type=checkbox]{width:1.25rem;height:1.25rem}` (`.confirm-check input`과 같은 값). 제출 동작·문구·타이핑 확인 추가 금지 | 스펙 44px, 안전 불변식 4(즉시성 불변), "확인 문구 금지" 준수 | 미고정(대비 시험은 안 잼). `settings_*_test.go`가 `<label>` 마크업 문자열을 단언하는지 확인 후 CSS만 바꿀 것 — 마크업 불변이면 안전 | Impeccable audit(측정), hallmark(간접) |
| R2 | `templates.go:49~53` `--ease-spring`(348) 와 `pop-in`(374~376) | minor | 모달 열림에 정점 1.287(28.7% 오버슈트) + 360ms. 읽기 전용 상세가 "튕기며" 열리는 건 operate 톤에 과함. 나머지(nav·버튼 press)는 절대 변위가 작아 무해 | 권고안 A: `.detail-cell .row-details[open]`의 `animation`을 `var(--ease-settle)`로(정점 1.045, 4.5%) 바꾸고 `--dur-3`을 300ms로. 권고안 B: `--ease-spring`을 `linear(0, 0.55 8%, 1.08 18%, 1.14 26%, 1.04 36%, 0.99 48%, 1.01 62%, 1 78%, 1)` 류(정점 ~1.14)로 낮춤. 스프링 느낌은 유지(오버슈트 ≥10%). **제거 요구 아님** | 브리프(스프링 요구 유지) vs operate.md 150~250ms·"motion conveys state" | `display_primitives_test.go`/`trading_views_test.go`가 `linear(`·`--ease-spring`·`--dur-3` 리터럴을 고정할 가능성 — 고정돼 있으면 시험 동시 수정 필요(미확인) | Impeccable critique(모션) |
| R3 | `templates.go:375` bounce-easing 탐지 | — | 탐지기 advisory 1건 | **do not fix** — 규칙 20 "스프링 탄성은 사용자 요구"가 탐지기 취향보다 우선(brief wins). 위 R2의 오버슈트 수치 조정만 권고 | 규칙 20 | — | Impeccable detect |
| R4 | `templates.go:360~368` 제외 선택자 | minor(방어) | 제외 목록이 `form:not([method=get]) *`·`.sticky-save *` 등 **경로 의존**. 승인/청산 버튼이 이 다섯 컨테이너 밖에 새로 생기면(예: `method` 속성 없는 `<form>` = 기본 GET이지만 JS가 POST로 바꿀 때, 또는 `formmethod="post"` 속성이 달린 버튼) 전환이 붙음. 현재 렌더 화면 측정에선 위반 0이나 화면별 승인 화면(approval·청산)은 미렌더 | 제외 판정을 CSS가 아니라 **마크업 표식**으로: 상태 변경 버튼에 `data-mutating`(또는 클래스)을 서버가 붙이고, 전환 규칙을 `button:not([data-mutating])`으로. 최소안: `button[formmethod]`, `[formaction]`도 제외 선택자에 추가 | 안전 불변식 4, 규칙 20 단서("즉시성은 전환으로 늦추지 않는다") | 구조 시험(`display_primitives_test.go`)이 제외 선택자 문자열을 이미 고정한 것으로 보임(4단계 커밋) — 새 표식은 템플릿 전수 시험 추가 필요 | Impeccable audit(안전 관점 판독) — 미측정 |
| R5 | `templates.go:378` reduced-motion | nit(통과) | 토큰 리셋 방식: `--dur-*`를 0.01ms로, `--nudge:0`,`--press:1`,`--pop-from:1`. 상태(색) 변화는 즉시 유지, 변위만 제거 → audit.md 기준 통과. 단 `@keyframes pop-in`이 `from{opacity:0}`는 남아 `0.01ms`로 즉시 완료 | 변경 불필요(정보 표기) | 접근성 | — | Impeccable audit |
| R6 | `templates.go:172~177` 모달 | minor | 모달 뒤 스크림 없음(스크린샷 `modal-open-1280-light.png`: 본문 위에 그림자만). 시선 집중·"프리미엄 머티리얼" 약함. 읽기 전용 상세라 안전 무관. `<details>` 모달이라 배경 비활성화·포커스 트랩은 구조적으로 불가 | `.detail-cell .row-details[open]::before`는 부모 박스 밖이라 불가 → 대안: `details[open]`에 `outline`/`box-shadow` 확산을 키우거나(`--shadow-pop` 0 18px 54px → 0 0 0 100vmax rgba(…,.18) 금지: 클리핑·성능), 또는 **그대로 둠**. 권고: 그대로 두고 문서화(비목표) | 규칙 20 프리미엄, operate.md "modal as first thought" 경고 | 미고정 | avoid-ai-design 시각 tell |
| R7 | `templates.go:173,326` 모달 폭 `100vw` | nit | `position:fixed`+`left:50%`+`100vw`는 스크롤바가 있는 데스크톱에서 스크롤바 폭만큼 중심이 어긋남(≈8~15px) | `width:min(42rem, calc(100% - 2rem))`(fixed의 `100%`는 뷰포트) / 모바일 `width:calc(100% - 1.5rem)` | 레이아웃 정확성 | 시험이 `calc(100vw - 1.5rem)` 리터럴 고정 가능(미확인) | hallmark 34 |
| R8 | `templates.go:97~98` | nit | `letter-spacing:0.02em`을 `dt`·`small`류 한글에 적용. 한글 소제목 자간 확대는 가독성 이득 근거 약함(라틴 소문자 레이블 관행) | 한글이 대부분이므로 `0.02em` → `0`(또는 라틴 전용 요소만) | 규칙 20 프리미엄 | 미고정 | Impeccable critique(typeset) |
| R9 | 대시보드 `section` 11개 (`templates.go:88`) | minor | 모든 섹션이 같은 `h2` 크기·같은 위 구분선. 3단계 #12(면 2단계) 미반영. 대시보드 4651px | 상태 요약만 `--surface-1`+`--lift`(이미 있음), 나머지 섹션은 구분선 유지하되 `section + section` 위 여백을 `--space-5`로 올려 호흡 확보. 내용 삭제 아님 | 규칙 20 저밀도·프리미엄 | 미고정 | Impeccable critique, avoid-ai-design |
| R10 | `templates.go:370~371` `.choice-tile` | nit | `scale` 전환이 서버가 정적으로 그린 `data-selected`에는 일어날 전환 변화가 없음(새로고침 시 전환 불발) → 죽은 규칙 + `scale:1.012` 상시 확대로 타일 정렬이 미세하게 어긋남(1.2%) | `.choice-tile{transition:…}`와 `[data-selected=true]{scale:…}` 삭제, 선택 강조는 이미 `--ok-bg`+`inset 1px` 링으로 충분. YAGNI | KISS/YAGNI(규칙 1·2) | `trading_views_test.go`가 `--tile-on` 토큰 존재를 고정했을 수 있음(미확인) | Impeccable audit(구현 무결성) |
| R11 | `templates.go:180` `::after{content:" · 닫기"}` | nit | 스크린샷에서 `상세· 닫기`로 붙어 보임. `content`의 선행 공백이 `display:flex` 아이템화(summary가 flex)로 붕괴 | `content:"닫기"`, 앞에 `margin-inline-start:.4em`로 간격(또는 `gap`) | 가독성 | 시험이 ` · 닫기` 문자열 고정 가능성(미확인) | 스크린샷 시각 확인 |
| R12 | `h1/h2`의 `<span class="muted">` 부제 (마크업) | nit | avoid-ai-design SD5 ×3: 제목 뒤 한 구절 옅게 처리 | 유지 가능. 바꾸려면 `<small>`+줄 분리. 오탐 성격 높아 변경 비권고 | 일관성 | 마크업 문자열 시험이 있을 수 있음 | avoid-ai-design |
| R13 | `templates.go:58~72` `body` | nit | `overflow-x` 방어선 없음(현재 hscroll 0 20조합) | `html,body{overflow-x:clip}` 추가는 `position:sticky`(`.sticky-save`, `.optimization-nav`)를 깨지 않는 `clip`이라 안전하나 현재 문제가 없어 **YAGNI** — 보류 | 규칙 1·2 | — | hallmark 34 |
| R14 | `.positions-table`(126~137, 152~153), `.line-item`(143~146) | — | 3/5/1·6px, 10~12px 리터럴·4px 스케일 이탈 | **do not fix** — `a058_stockos_density_test.go` 리터럴·스펙 146/158 고정 | 고정 계약 | 고정됨 | hallmark 24b |
| R15 | `table{font-size:.9rem}`(94), `.safety-line{font-size:1rem}`(158) | — | 14.4px가 스케일 밖 | **do not fix** — 시험 리터럴 고정(`console-design.md` 명시) | 고정 계약 | 고정됨 | Impeccable(관찰) |
| R16 | 카피: em dash·`·` 결합·영어 verify 원문·"구조적으로 미측정" 반복 | — | 문구/백엔드 | **do not fix**(스타일시트·마크업 구조 범위 밖) | 범위 | — | hallmark 19·63, avoid-ai-design SD4d |
| R17 | `meta refresh` 5초 | — | 스크롤·포커스·모달 상태 소실 | **do not fix** — 백엔드 계약(3단계에서 이미 do-not-fix) | 백엔드 계약 | — | Impeccable(관찰) |

## 상위 5 (영향 순)
1. R1 — 설정 토글 체크박스·라벨 44px (측정된 유일한 표적 위반, 8개).
2. R4 — 상태 변경 버튼의 전환 제외를 경로 의존에서 표식 의존으로(안전 불변식 4 방어, 승인·청산 화면은 미렌더라 미측정).
3. R2 — 모달 스프링 오버슈트 28.7%→10~15% / 360→300ms (스프링 유지, 제거 아님).
4. R9 — 대시보드 섹션 위계·여백(저밀도 브리프).
5. R10 — 죽은 `.choice-tile` 전환·`scale:1.012` 제거(YAGNI, 정렬 오차 제거).

사후: 측정용 콘솔(포트 38190) 종료. `ps -eo pid,args | grep -E '[c]onsole --port 38190'` 출력 없음 확인. 리포 편집 0(`git status --short` 빈 출력).
