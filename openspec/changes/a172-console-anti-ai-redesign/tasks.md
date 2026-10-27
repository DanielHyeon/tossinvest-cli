# a172 tasks

Story: `STORY-TOS-a172` · Low risk(CSS·장식 마크업만). FLM/BTM·변이·다중 적대 리뷰 `not-applicable`(proposal).

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · base `fd6afe23b`

## 1. 1단계 — 날것의 담백한 구조
- [x] 1.1 징후 측정 → scratchpad a172/stage1-scan.md
- [x] 1.2 제거(eyebrow·균일 반경 중립화·평면 카드 배경)
- [x] 1.3 gofmt·vet·console/cmd 시험, 렌더 확인

## 2. 2단계 — 프리미엄 미니멀 테마(hallmark 대체)
- [x] 2.1 `:root` 토큰(`--space-*` clamp·`--radius-*` 역할별 3조합·`--surface-0/1/2`·`--line`·`--ink-*`·`--accent`·`--ok/--warn/--bad`)과 다크는 변수만 재정의
- [x] 2.2 비대칭 반경(card·control·soft, 알약 유지), 중첩 테두리를 면·약한 그림자·구분선 하나로 대체(테두리 선언 25 → 15)
- [x] 2.3 본문·제목 clamp 스케일, `.eyebrow/.section-kicker` 셀렉터 삭제, 기본 링크 색 `a { color: var(--accent) }`
- [x] 2.4 gofmt·vet·console/cmd 시험, 20장 렌더(`scratchpad/a172/stage2/`)

## 3. 3단계 — 디자인 시스템 무결성(타입 스케일·팔레트·대비·CSS 변수 재조정)
- [x] 3.1 토큰 감사: 31 → 33개, 전부 사용 ≥1회. `--bad-soft`·`--state-warn`·`--state-off`·`--error-line`·`--btn-*`·`--field` 폐지, 잉크 1→3 단조, 의미색 fg·line·bg 세 역할
- [x] 3.2 모듈러 타입 스케일 `--text--1..3`(clamp, 단계별 줄 간격·자간)
- [x] 3.3 팔레트 재조정: Tailwind/Material 기본값·순수 무채색 0건, 모든 hex 는 `:root` 정의 38곳에만 존재
- [x] 3.4 라이트·다크 양 테마 대비 시험(텍스트 4.5:1, UI 경계 3:1), `docs/console-design.md`

## 4. 4단계 — 스프링 물리 인터랙션 + GPU 가속 최적화(`ui-ux-pro-max` + `fixing-motion-performance`)
- [x] 4.1 `linear()` 스프링 이징·시간 토큰, nav·탭·링크 버튼 `:active`·선택 타일·상세 모달 열림(`opacity`+`scale`)
- [x] 4.2 전환 속성 compositor 전용(hover 색 ≤ `--dur-1`), `will-change` 1곳, blur 0
- [x] 4.3 POST 폼·`.sticky-save`·`.approval-actions`·`.confirm-check`·`.console-approval` 전환 0(안전 불변식 4), 시험 `TestMotionContract`(변이로 FAIL 확인)
- [x] 4.4 reduced-motion(토큰 리셋), 프레임 측정(`scratchpad/a172/frames.cjs`: 모달 10회·nav hover 20회, p95 16.8ms, 32ms 초과 0)

## 5. 검증·포트
- [x] 5.1 전체 콘솔 시험 · 렌더 · 경량 리뷰 · 포트 — 2026-10-25 main ff `79d32f9af`(비캐시 console 48.3s · cmd/tossctl 63.3s, 검출기 `[]`); 사람 결정 2건은 권장안 확정(POST 폼 버튼 전환 전부 제외 유지, 스프링 유지·모달 오버슈트 12%)

## 6. 6단계 — CSS-only 셰이더풍 인터랙션(사용자 지시 2026-10-25, 사람 결정: JS·CSP 완화 불채택)
- [x] 6.1 정찰 — JS·WebGL 안 검토 후 불채택(design.md 「6.1 정찰: 검토했으나 채택 안 함」), 범위 CSS-only 로 정정
- [x] 6.2 구현: 앰비언트 톤 배경·새로고침 화면 정지(`data-reload=on`, 마크업 변경 0)·카드 hover 조명·방어. 신규 `TestAmbientMotionContract`(RED→GREEN), 변이 2회 CAUGHT(hover 제외 규칙 `.sticky-save *` 삭제 / 새로고침 `animation: none` 삭제)
- [x] 6.3 측정(`frames6.cjs`·`trace6.cjs`, light·dark) · 6.4 `fixing-motion-performance` 점검(수정 0) · 6.5 `avoid-ai-design` detect P0 0 · Impeccable `[]` · 20장+hover 2+톤 6장(`scratchpad/a172/phase6/shots/`)
- [x] 6.5b 라이트 테마 톤 구분(사용자 지시 2026-10-25, 리뷰어의 "강도 올리지 말 것" 번복): 전용 `--tint-*` 토큰+테마별 `--tone-alpha`(라이트 19%·다크 15%), 톤 쌍 ΔE76 렌더 실측 ≥14.1·글자 대비 ≥4.65. 신규 `TestAmbientToneLegibility`(변이: 라이트 alpha 10% → FAIL). 고정 리터럴 변경 1건: `TestAmbientMotionContract` (e) 의 `--tone: var(--ok|warn|bad)` → `--tone: var(--tint-*)`
- [x] 6.6 main 병합(사람 승인 뒤)
  - 증거(2026-10-28 종결): `git merge-base --is-ancestor b27a1ea87 eaf713ff9` 참 — a172 커밋(5.1 포트 main ff `79d32f9af`, 6단계 `156340d04`·`4f3b94e28`·`b27a1ea87`, spec 델타 `794f9e95a`)이 main 의 조상, 2026-10-25 배포·운영 중.
  - 종결 승인: 사용자 "1"(2026-10-28) — FLM 번들 작성·복사, review.md 작성, 커밋, `/tmp` 밖 격리 worktree 워밍업 뒤 전체 `make gate`, archive, Story 경로, tracker, push, 정리, retain. base 재고정 금지.
  - gate ⑤ 사전 실행: required 26 → 번들 복사 18 · 새로 작성 8 → rc 0 (review.md 「gate ⑤」). 리뷰 원문 3 + 징후 측정 1 은 스크래치패드에서 찾아 `reviews/` 로 복사.
  - 완료 게이트 결과는 review.md 「완료 게이트 실행 영수증」에 적음(관례: 체크 → tracker → gate).

## 고정 리터럴 시험 수정 기록
- 2단계 `display_primitives_test.go` `TestTheHeadingStepsAreDistinguishable` + 헬퍼(`declaredFontSize`·`declaredBodyFontSize`): 제목·본문 크기가 `clamp()` 라 `Nrem`/`Npx/` 리터럴 파서가 읽지 못함. 파서가 clamp 의 최소·최대를 읽도록 바꾸고, 제목 단계 순서(h1>h2>h3>본문)를 가장 좁은 뷰포트와 가장 넓은 뷰포트 양쪽에서 검사하며 본문 최소 15px 기준은 유지(본문 clamp 최소 0.9375rem = 15px). 약화 아님: 검사 지점이 1곳에서 2곳으로 늘어남.
- 2단계 `trading_views_test.go` `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`: 다크 색이 개별 규칙(`.ok { color: #22c55e; }`)에서 `:root` 변수로 옮겨져 규칙 리터럴이 사라짐. 다크 블록의 변수 값을 읽어 `--ok/--warn/--bad/--bad-soft/--accent/--ink-1~3` 이 `--surface-0/1/2` 세 면 위에서 모두 4.5:1 이상인지 잼(기존: 두 색 × 한 면 `#1d1d22`). 약화 아님: 대상 색 2 → 8, 면 1 → 3.
- 3단계 `display_primitives_test.go` 헬퍼: 크기가 `--text-N` 토큰(`font-size: var(--text-N)`, 본문은 `body` 규칙)으로 바뀌어 `tokenSize` 가 `:root` 의 clamp 를 읽도록 확장. 검사 내용(제목 단계 순서 양끝 뷰포트, 본문 최소 15px)은 동일.
- 3단계 `trading_views_test.go`: `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA` → `TestTradingViewsThemeColorsMeetWCAGInBothThemes`. 한 함수가 라이트·다크를 돌며 텍스트 8종 × 세 면 4.5:1, 잉크 × 의미색 바탕 4.5:1, 의미색 fg × 자기 bg 4.5:1, UI 경계(`--line-strong`·`--accent`·의미색 line) × 세 면 3:1, 잉크 단조성. 약화 아님(라이트 신규, 대상 증가).
- 3단계 시험 고정으로 스케일 밖에 남긴 규칙: `.safety-line { ... font-size: 1rem; }`(a111 시험), `table` 0.9rem·모바일 `.line-item` 0.9rem(a057 시험), 보유 표 px(a058).

## Impeccable 3단계 리뷰 반영 (review-stage3.md, 사용자 지시: 단계마다 Impeccable 리뷰)
번호 · 위치 · 처리
1. `.notice`·`.danger` 3px border-left · 반영: 1px 전방 균일 테두리(`--warn-line`/`--bad-line`) + `--lift`. 시험 `TestNoSideTabAccentBorders` 추가(2·3 포함), 검출기 JSON `[]`, 시험을 변이(border-left: 3px 재삽입)로 FAIL 확인
2. `.card-preview` 3px border-left · 반영: `--surface-2` 면 + `--radius-soft`, padding `--space-2/3`
3. `.optimization-nav a.on` `inset 3px 0` · 반영: `inset 0 -2px var(--accent)`(설정 탭과 같은 어휘)
4. 모바일 `.card-preview` 한 열 · 반영: 720px 이하 `.card-current, .card-preview, .entry-list` 를 `minmax(0, 1fr)`. 375 측정 dd 37.6px → 300px
5. 입력칸 44px · 반영: `input[type=text]`·`.hypothetical-fields input/select` `min-height: 44px`(보유 표 안에는 입력칸 없음)
6. `.explain summary` 44px · 반영: 전역 `summary { min-height: 44px; align-content: center; }`(`.row-details summary` 규칙은 a057 리터럴이라 그대로). 375·1280 5화면 44px 미만 0건
7. 스크롤 표 글자 단위 줄바꿈 · 부분: 첫 열 `nowrap` 은 둘째 열("검증 진행" 41px)을 못 구해 측정 후 폐기, `.table-scroll th, td { min-width: 4.5rem }` 채택(375 대시보드 첫·둘째 열 36~41px → 72px). 표 마크업·`<caption>` 은 건드리지 않음
8. 버튼 상태 · 반영: `button`·`.button-link` `:hover`(`--ink-2`)·`:active`(`--ink-3`)·secondary 변형·`button:disabled` 전역. 채움 유지, transition 0
9. 인라인 style · 반영: `.header-note` 클래스, 모바일 `!important` 제거(남은 `!important` 는 a057 시험 고정 1건)
10. 다크 그림자 덮어쓰기 · 반영: `--shadow-pop`·`--shadow-bar` 토큰, 다크 `:root` 에서 값만 재정의, 개별 규칙 덮어쓰기 삭제
11. 다크 카드 경계 1.1:1 · 반영: 다크 `--surface-1` `#1d1c19` → `#201f1b` + 다크 `--lift` 에 `--line-soft` 1px 윤곽. 대비 시험 재실행 통과
12. 위계 · 부분: 정보 섹션은 상단 구분선만(면·그림자 제거), 떠 있는 면은 상태 스트립·`.notice`/`.danger`·`.sticky-save`·`.order-summary` 에만. 「첫 섹션 h2 한 단계 키움」은 h2 가 이미 `--text-2` 이고 텍스트·순서 변경 없이 키울 근거 측정이 없어 미반영(사유: 효과 미측정)

### 6단계 Impeccable 리뷰 반영 (review-phase6.md: critique 25/32, audit 15/20, avoid-ai-design 신규 0, 대비 최저 5.10:1)
번호 · 처리 · 사유
- 1 · 반영 · `body > header { overflow: clip; overflow-clip-margin: 6px }`(nav 포커스 링 여유) + 의사요소 `width: min(42%, 30rem); height: min(170%, 22rem); top: max(-35%, -7rem)`. 헤더 밖 번짐·h1 대비는 아래 실측
- 2 · 반영 · `.tier-card`(POST 폼 부모)·`.choice-tile`(POST 폼 포함)을 hover 조명 목록에서 뺌(`:has()` 부정보다 단순·확실). 주석·design.md·console-design.md 문구를 실제와 맞춤
- 3 · 반영 · 시험 (g) 렌더 기반: 설정 wired 하네스로 `consoleScreens` 전부를 렌더·파싱해 조명 대상(`.optimization-nav a`·`.settings-tabs a`) 하위 `<form` 0 단언 + 목록에 `.tier-card`/`.choice-tile` 재진입 시 실패 + 매칭 0건이면 실패(빈 표본 통과 방지). 변이 1회(`.tier-card` 를 목록에 되돌림 → FAIL, 렌더 기반 단언이 `/settings/daily` 에서 `<form>` 3건을 잡음)
- 4 · 반영 · `prefers-reduced-transparency` 가 hover `::after` 도 `background: none`, 시험이 앰비언트·조명 둘 다 요구
- 5 · 거절 · 키프레임 `ease-in-out`: `TestMotionContract` 가 `ease-in-out` 을 금지(4단계 계약, 단일 곡선 이징 금지). `linear alternate` 유지, 끝점 속도 꺾임은 약 3px/s 라 지각 불가
- 6 · 반영 · 새로고침 정지 선택자를 `:is(body > header, .console-status)` 로 좁힘(`.choice-tile header` 등 중첩 header 제외)
- 7 · 반영 · 문서 「정보를 나른다」→「약한 질감, 상태 단어가 정본」 + 톤 매핑 한 줄(ok=실행+신선, warn=정지·지연, bad=stale, 그 외 accent)
- 8 · 반영 · 시험 (b) 금지 목록에 `backdrop-filter` 추가
- do not fix 유지 · 톤 농도 상향·JS 포인터 추적·CSP/WebGL·blob 전면 제거 (사람 결정 / 브리프 충돌)

### 6단계 고정 시험 수정 기록
- `TestMotionContract` 허용 속성 맵에 `rotate` 1개 추가(앰비언트 키프레임이 개별 `rotate` 속성을 씀). 이 밖의 기존 시험·CSP 핀 수정 0, `<script`·`script-src` 부재 핀 전부 그대로 통과.

## Impeccable 4단계 리뷰 반영 (review-stage4.md)
번호 · 처리 · 사유
- R1 · 반영 · 설정 토글 라벨 17~19px → 1280 46px·375 44px(실측 8개, 체크박스 20px). `label:not(.confirm-check):has(> input[type=checkbox]) { display: block; min-height: 44px }` — flex 가 아니라 block 이라 라벨 안 `<strong>`+문장 인라인 흐름이 그대로. 마크업·문구·제출 동작 불변
- R2 · 반영 · 모달 전용 `--ease-pop`(정점 약 1.11~1.13)·`--dur-3` 300ms. nav·눌림은 `--ease-spring` 유지. 시험은 `--ease-pop: linear(` 단언 추가
- R3 · 거절(리뷰어 do-not-fix) · 규칙 20 의 스프링 지시가 탐지기 취향보다 우선. 검출기 advisory 1건 그대로 보고
- R4 · 반영(부분) · 제외 목록에 `[formmethod]`, `[formaction]`, `form button`, `form input[type=submit]` 추가(세 :not 목록 + transition: none 규칙), `TestEveryPostFormButtonIsExcludedFromMotion` 이 네 선택자의 존재를 단언. 서버 `data-mutating` 표식은 Go 변경이라 이번 범위 밖 — 후속 후보
- R5 · 해당 없음 · 변경 불필요(리뷰어도 통과)
- R6 · 반영 · `.detail-cell:has(.row-details[open])::before` 고정 스크림(`--scrim`, opacity 애니메이션, `pointer-events: none`). 모달 박스 안 `::before` 는 `overflow: auto` 에 잘려 셀 쪽에 둠. `backdrop-filter` 0. 주입 DOM 스크린샷 1장
- R7 · 반영 · `width: min(42rem, calc(100% - 2rem))`, 모바일 `calc(100% - 1.5rem)`·`max-height: calc(100% - 3rem)`(두 곳)
- R8 · 반영 · 한글 레이블 `letter-spacing: 0.02em` 그룹 규칙 삭제. 라틴 전용 `header strong`(tossctl console) 만 유지
- R9 · 반영 · `section { margin: var(--space-5) 0 }`. 면 층(상태 스트립·경고·승인만 떠 있음)은 3단계에서 이미 적용, 텍스트 불변
- R10 · 반영 · `.choice-tile` 전환·`scale` 상시 확대·`--tile-on` 삭제
- R11 · 반영 · `content: "닫기"` + `margin-inline-start: 0.5em`
- R12 · 보류 · 오탐 성격(제목 뒤 muted 부제). 기록만
- R13 · 보류 · `overflow-x: clip` 은 측정 hscroll 0 이라 YAGNI. 기록만
- R14~R17 · 유지 · do not fix(a058·스펙 리터럴, 시험 고정 폰트 크기, 카피·백엔드, meta refresh)
