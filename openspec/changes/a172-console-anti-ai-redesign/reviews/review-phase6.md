# a172 6단계 독립 디자인 리뷰 (HEAD 156340d04, diff main...HEAD)

방법: Impeccable 4.6.0 critique/audit/optimize/craft-floor 를 직접 수행(서브에이전트 분리 없이 단일 컨텍스트 — `⚠️ DEGRADED: single-context`, 리뷰 전용 요청이라 분리 안 함). PRODUCT.md/DESIGN.md 없음 → 코드 기준, Operate 모드. 저장소 파일 수정 0.
측정 근거: 워크트리·main(3ac55d7c5) 콘솔 각각 빌드·GET 전용 기동(38196/38197), 5화면×light/dark 덤프, Playwright 로 계산·픽셀 측정. 스크립트 `r6dump.cjs r6contrast2.cjs r6contrast3.cjs r6glow.cjs r6m.cjs`(같은 폴더).

## (a) critique

| # | 휴리스틱 | 점 | 핵심 |
|---|---|---|---|
| 1 | 시스템 상태 가시성 | 3 | 톤 틴트는 상태 단어와 중복이고 약함(아래 ΔE). 해가 없음 |
| 2 | 현실 일치 | 3 | engine=stopped → warn 앰버가 휴장·주말의 상시 상태에 깔림 |
| 3 | 사용자 통제 | 4 | 장식은 pointer-events:none, reduced-motion/transparency 존중 |
| 4 | 일관성 | 3 | hover 조명이 nav·탭에선 이미 있는 translate+배경색 위에 세 번째 신호로 겹침 |
| 5 | 오류 방지 | 3 | 주문·승인 카드(tier-card·choice-tile)에 조명이 켜짐 → 문서 주장과 불일치(F2) |
| 6 | 인식 > 회상 | 3 | 톤 의미를 UI 어디에도 설명 안 함(발견 불가) |
| 7 | 유연성 | n/a | |
| 8 | 미니멀 | 2 | 정보 가치 대비 4 레이어 + 상시 애니메이션. 375px 에서는 세로 기둥(F1) |
| 9 | 오류 복구 | n/a | |
| 10 | 도움말 | n/a | |
합계(해당 8항목): 25/32. 디자인 특이성: 톤 연동은 제품 고유(정보를 나름), 나머지(떠다니는 radial blob)는 범용 "앰비언트 오로라" 문법.

판정(ops 콘솔에서 이 레이어가 자리값을 하는가): **조건부 통과.** 브리프(JS 0, CSP 불변, 정보 있는 그라디언트만, 저밀도)를 문자대로 충족. 단 실측으로 정보 전달력이 약하다 — 톤 틴트 vs 배경 ΔE(CIE76, 피크 10%): light 6.0~7.0 / dark 8.5~10.2 (중립과는 구분됨). 톤끼리 ΔE: light ok-warn 3.9, warn-bad 3.4, ok-bad 5.9 / dark 5.6~7.3. 라이트에서 ok↔warn↔bad 는 움직이는 10% 틴트로는 사실상 구분이 안 된다. → 이 그라디언트는 "정보를 나른다"기보다 "정보에 질감을 준다". 상태 단어·색 글자가 이미 판정을 한다. 해롭지는 않으므로 제거 권고는 아님(브리프 결정 존중), 문서가 "정보를 나른다"고 쓴 점만 과장.

주요 발견
1. (F1, P2) 375px 에서 blob 이 세로 기둥. header 536~598px 높이 × 170% = 911px, 폭 42%. 스크린샷 `r6-m375.png`: 좌우 세로 띠가 header 밖(+35% ≈ 190px)으로 흘러 h1 「설정 당일」·탭 위에 칠해짐. 1280 에서도 header 는 218px(status 가 header 안)이라 +76px 가 main 상단(h1)으로 번진다.
2. (F2, P2) 안전 불변식 4 의 문구와 구현 불일치. 문서(console-design.md:+1)·주석은 "주문·승인 영역 안에는 적용하지 않는다" 인데 제외 선택자는 `form:not([method=get]) *`(자손)뿐이다. `.tier-card`(POST form 의 부모)·`.choice-tile`(안에 POST form) 자체는 제외되지 않음. 실측: /settings/daily tier-card 5/5 가 POST form 포함 + ::after 켜짐, /optimization?category=exit-protection choice-tile 3/3 동일. 지연·가림은 없다(z-index -1, pointer-events none, opacity 만) → 즉시성 위반 아님. 그러나 위험 한도·익절 preset 을 적용하는 카드가 "눌릴 것 같은 조명"을 받고, 시험 (g) 은 선택자 문자열만 확인해 못 잡는다.
3. (F3, P3) 톤 기본 규칙: running&&ok→ok, stopped→warn, stale→bad(마지막이 이김), 그 외(on-request 화면·unwired)는 accent(파랑). 즉 녹색은 "실행 중 + 캐시 신선"일 때만, 휴장 정지 상태는 상시 앰버. 의미 있으나 문서에 한 줄도 없다.
4. (F4, P3) 호버 조명이 .settings-tabs a·.optimization-nav a 에선 translate + 배경색 + 조명 3중. 카드(tier/choice)에선 :focus-visible 분기가 죽은 코드(컨테이너는 포커스 안 받음).

## (b) audit / optimize

audit 점수: A11y 3 · Performance 3 · Theming 4 · Responsive 2 · Impl. integrity 3 = **15/20 (Good)**.
- A11y: 텍스트 대비 ≥5.13 (아래 표), reduced-motion 시 정지, reduced-transparency 시 제거, 장식 전부 pointer-events none. 빠진 것: reduced-transparency 에 hover 조명(12% accent)이 없음(F5), forced-colors 분기 없음(P3 — 그라디언트는 강제색에서 남을 수 있음, 영향 미미).
- Responsive: 375 에서 F1(기둥·번짐). hscroll 0(실측 3화면). 44px 대상 불변.
- Theming: 색 전부 토큰(`--tone` → --ok/--warn/--bad/--accent, color-mix). 다크 재정의 불필요하게 자동 따라감. 
- Performance(optimize): 키프레임은 translate/rotate/opacity 만(컴포지터). 레이어 수: 페이지당 의사요소 4개(header 2 + status 2) 상시. 래스터 면적 ≈ header blob 538×370 ×2 + status blob 538×190 ×2 ≈ 0.6M px ≈ 2.4MB 텍스처 추정(측정 아님, 크기로 계산). `:has()` 비용: body 의 5개 :has 는 문서 로드 시 1회 평가, 이 콘솔은 JS 가 없고 DOM 변이가 없어 재평가 없음(2초 새로고침은 전체 로드) → 무시 가능. 작성자 프레임 측정(phase6/after-*.json, 헤드리스 소프트웨어 렌더): on/static/none 모두 p95 16.8ms, >32ms 0, longtask 0 → 차이 없음. 단 헤드리스는 GPU 구동 전력을 못 잰다. 설정·최적화 등 비-새로고침 화면은 34/41초 무한 애니메이션이라 탭이 열려 있는 동안 프레임 커밋이 계속된다(P3, 하루 종일 켜 두는 콘솔).
- 선택자 폭: `body:has([data-reload=on]) :is(header, .console-status)::before` 의 `header` 는 비한정(내부 `header` 에도 매칭, 현재는 의사요소가 없어 무해) — 좁힐 것.
- 시험 공백: `filter` 정규식이 `backdrop-filter` 를 못 잡음(현재 사용 0, 실측 `backdrop-filter|blur(` 0건). (f)(reduced-motion) 정규식은 `@` 로 끊겨 올바른 블록에만 닿음을 확인(약하지만 정상).

craft-floor 점검: gradient text 없음, kicker 없음, glass/blur 없음, hard shadow 없음, 장식 이모지 없음 → 통과. "motion: one authored moment" — 앰비언트(상시) + hover 조명 + 기존 nudge/모달 → 이 콘솔의 규율(저밀도)엔 이미 충분, 추가 금지.

## (c) 검출기 원문

Impeccable `detect --json internal/console/templates.go internal/console/templates_optimization.go` (exit 0):
```json
[]
```
주의: 이 검출기는 .go 템플릿 문자열을 마크업으로 보지 못한다 — "0건"은 "검사 0"에 가깝다. 렌더 HTML 은 아래 avoid-ai-design 이 대신 봄.

avoid-ai-design 0.4.0, 렌더 덤프 5화면×2스킴 + console.css(11파일): P0 0 · P1 8 · P2 10 (총 18). 같은 덤프를 main 3ac55d7c5 에서 뜬 것과 (파일,규칙,노트) 집합 비교 → **신규 0, 소멸 0** (줄 번호만 +36 이동).
| 규칙 | 건수 | 대상 | 상태 |
|---|---|---|---|
| SD5 한 단어 강조 헤드라인 | P1 ×6 | 「시장별」「전략 파라미터」 등 h1/h2 의 `<span>` 강조 | 기존(main 동일) |
| SD4d 가운데점 메타 문자열 | P2 ×10 | 상단 nav·상태 문구 | 기존 |
| L4 "수치 스트립" | P1 ×2 | settings 줄 526 | 기존, 오탐(실데이터 한도 표) |
전체 JSON: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/89bda6e1-9160-4a44-b964-4d754f48371b/scratchpad/a172/r6-avoid.json`(워크트리), `r6m-avoid.json`(main). 6단계 CSS 는 아무 규칙에도 안 걸림. (SD5 는 설정 h1 이 의도적 이중 톤 제목이라 브리프상 phases 1~5 소관 — 재제기 안 함.)

## (d) 대비 측정 (텍스트 아래 틴트 픽셀 최악, 실측)
방법: 텍스트를 투명 처리한 스크린샷에서 각 요소 박스의 모든 배경 픽셀 vs 텍스트 색 WCAG 비. 키프레임 극단 4조합(a,b 각각 0/끝, opacity 최대 + 겹침 최대 포함) × 톤 ok/warn/bad/accent × light/dark, 전환 꺼서 중간색 오염 제거(첫 시도가 transition 중 투명 색을 읽어 4.4~4.6 이 나왔으나 아티팩트로 폐기·재측정). 설정 화면, 1280.

| 요소 | light 최소(톤 무관) | dark 최소 | 무틴트 기준(light/dark) |
|---|---|---|---|
| .console-status dt | 5.66 | 5.43 | 5.89 / 6.4+ |
| .console-status dd (엔진 정지, --warn 글자) | 5.67 | 7.48 | 6.39 / 8.74 |
| .console-status dd small | 5.34 | 5.10 | ≈5.9 / ≈6.3 |
| status 값 [data-data-tone] 자식 | 5.39 | 5.13 | 6.32 / 6.25 |
| header nav a | 7.86 | 9.72 | 8.87 / 10.99 |
| header nav a small | 5.22 | 6.24 | 5.89 / 7.04 |
| 브랜드 | 7.66 | 6.08 | – |
| main h1 (blob 번짐 지점) | 14.18 | 14.42 | 15.01 / 15.30 |
**전체 최소 5.10:1**(dark, advisory small, ok/accent). 모든 톤·스킴에서 AA(4.5) 통과, 틴트 비용은 약 −0.4~−1.1. 한계: 4가지 키프레임 극단 + 중간 표본이 아니라 연속 최악은 아님(겹침이 가장 큰 극단은 포함). 마지막 행은 blob 이 main 콘텐츠 위로 번진다는 증거(F1).

## (e) 수정 목록 (순위)

1. **`internal/console/templates.go:~395-397`(`body > header, .console-status {…}` 와 의사요소 규칙) · P2 · F1**
   문제: header 밖으로 blob 이 번져 main 위에 칠해지고(1280: +76px, 375: 세로 911px 기둥·+190px 번짐). 
   수정: `body > header { overflow: clip; }` 추가(`.console-status` 는 이미 hidden), 의사요소 `width: min(42%, 30rem); height: min(170%, 22rem); top: max(-35%, -7rem)` 로 상한. 포커스링 클립이 걱정되면 `overflow-clip-margin: 6px`.
   봉사 계약: 모바일 375·콘텐츠 가독(h1 대비 −0.8)·저밀도. 시험: 새 핀 필요 — `body > header` 규칙에 `overflow: clip` 포함 + 의사요소에 `min(`.
2. **`templates.go:~416-420`(hover 조명 규칙 3곳) · P2 · F2**
   문제: `.tier-card`(POST form 부모)·`.choice-tile`(POST form 포함)이 조명 받음, 문서·주석 주장과 불일치.
   수정(최소): 선택자에서 `.choice-tile, .tier-card` 제거(nav·탭만 남김, 어차피 mutate 카드에서 장식 이득 0). 대안: 두 곳에 `:not(:has(form:not([method=get])))` 추가. 문서 줄도 "카드" 표현 정정.
   계약: 안전 불변식 4(주문·승인 영역에 장식 없음), 확인 마찰 없음. 시험: 현재 (g) 는 문자열만 → 렌더 HTML 의 POST form 보유 요소가 조명 규칙 대상 selector 에 매칭 안 되는지를 보는 시험으로 교체/추가(핀 필요).
3. **`display_primitives_test.go` (g) · P2** — 위와 짝. 실제 템플릿을 렌더해 `.tier-card`/`.choice-tile` 이 POST form 을 갖는 사실과 규칙의 비매칭을 단언. 지금 시험은 문구가 있으면 통과(실제 위반을 통과시킴).
4. **`templates.go:~406-409` reduced-transparency · P3 · F5** — `@media (prefers-reduced-transparency: reduce)` 에 hover 조명 `::after { background: none }` 추가. 계약: 접근성. 핀: 기존 `prefers-reduced-transparency` 문자열 시험 확장.
5. **`templates.go` 의사요소 애니메이션 · P3** — `linear alternate` 라 끝점에서 속도가 꺾임(3px/s 라 거의 안 보임). `ease-in-out` 으로. 정지 시간대 전력은 `@media (prefers-reduced-motion)` 외에 선택지 없음(상시 컴포지터 커밋) — 수용 여부는 사람 판단.
6. **`templates.go:~402` · P3** — `body:has([data-reload=on]) :is(header, .console-status)` → `body > header, .console-status` 로 좁힘.
7. **`docs/console-design.md` 추가 줄 · P3** — "정보를 나른다"는 "엔진·데이터 톤을 약한 질감으로 따른다(상태 단어가 정본)"로 정정하고 톤 매핑(ok=실행+신선, warn=정지·지연, bad=stale, 그 외 accent)을 한 줄 명시.
8. **시험 (b) 정규식 · P3** — `backdrop-filter` 도 금지 목록에 포함(현재 사용 0).

do not fix (브리프와 충돌):
- 톤 틴트 농도 상향(ΔE 를 키우자는 제안): 저밀도·"장식은 조용히" 와 충돌, 상태는 이미 단어·색 글자가 판정.
- JS 로 포인터 추적·가시성 정지·스포트라이트: JS 0 결정 위반.
- CSP/스크립트 허용으로 WebGL 셰이더 복귀: 사용자 결정 2026-10-25 위반.
- blob 제거 권고: 브리프가 채택한 레이어, 실측상 위해 없음(AA 통과, 프레임 영향 없음).
- 새로고침 화면 애니메이션 정지는 그대로 유지(옳음).

## 정리 증거
콘솔 38196·38197 PID 로 종료, `pgrep -ax tossctl | grep -c 38196` = 0 (38197 도 0). 임시 main worktree `r6-main` 제거·prune 확인(`worktree list` 0건, 디렉터리 없음). 대상 워크트리 `git status --short` 0줄(수정 없음).
