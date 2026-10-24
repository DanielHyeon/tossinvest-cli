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
- [ ] 4.1 스프링 느낌 모션(`transform`/`opacity` 만, `prefers-reduced-motion` 존중) · 4.2 컴포지터 속성만·`will-change` 최소 · 4.3 주문·토글·승인·청산 버튼에는 지연 전환 금지(안전 불변식 4)

## 5. 검증·포트
- [ ] 5.1 전체 콘솔 시험 · 렌더 · 경량 리뷰 · 포트

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
