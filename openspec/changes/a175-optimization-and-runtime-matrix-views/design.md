# a175 design

전제: a173 착지 뒤 트리(a174 와 병행 가능 — 주 파일은 a174 가 `templates_settings.go`·`templates.go` `dashboard` define, a175 가 `templates_optimization.go`·`templates_signals.go`). file:line 은 정찰(2026-10-25) 값, 편집 전 재확인.
**겹침 정정(Manager, 2026-10-26)**: 콘솔 CSS 는 전부 `templates.go` `style` define(:17-481) 한 곳에 있어서 a174·a175·a176 이 모두 그 블록을 고친다(계산기 `.hypothetical-fields` 도 거기 있다). 규칙: `style` **끝에** `/* a175 — … */` 블록 하나만 덧붙이고 기존 규칙은 고치지 않는다(a174 design §2.5). a174 가 먼저 착지하면 a175 의 gate ⑤ 창에 `pages.go:Console.handleDashboard` 가 든다 — a174 번들을 sha 를 대조한 뒤 복사한다. 착수 시 `base-commit.txt` 를 그때 main HEAD 로 갱신하고, 작업 위치는 `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-optimization-and-runtime-matrix-views`(사람 결정 2026-10-26).

## 1. 지형 (Manager 재확인, HEAD `b7d34acad` 2026-10-26 — 정의본 대비 바뀐 곳은 ★)
| 대상 | 위치 | 비고 |
|---|---|---|
| 최적화 템플릿 | `templates_optimization.go` `optimization` define(:4-231). 머리 KPI `dl.status-strip`(:11-18, Desired·Effective·성과 근거·Effective entry·LIVE 권한·재시작) 은 **모든 카테고리**에서 렌더된다. 관련 링크 `nav.filter-bar`(:24-28, 3개). 왼쪽 `nav.optimization-nav`(:31-38). 개요 카드 그리드 `.category-summary`(:45-53) | 카테고리 분기 `{{if .Overview}}`(:41) `.ExitProtection`(:64) `.PositionManagement`(:175) `.CandidateFilters`(:184) `.StrategyRuntime`(:203) `.PerformanceHistory`(:215) |
| 카테고리 뷰 | `optimization_view.go` `optimizationCategoryViews()`(a173 이 change ID 를 `Tech` 로 옮김) | 메뉴 라벨·설명 |
| ★ 익절 프리셋 | `.tier-card` 가 **아니다.** `div.choice-grid`(:95) 안의 `article.choice-tile` × `.Policies`(`optimizationPolicyView{exitpolicy.CommonPolicy, Selected, Token}`). 각 타일은 `.Ladder.Rungs` 표 하나(단계·목표 수익률·진입가 대비 보호·잔량 기준 익절)와 폼 `/optimization/exit-policy` `class="choice-action"`(hidden `csrf`·`base_version`·`category`·`setting_key`·`option_id`) → **preview 로 가는 1단계**다(preview→apply 2단계 불변) | 사다리는 정책마다 4단(`internal/exitpolicy/common_policy.go:35-52`). ★ runner 의 T4 `TargetPct` 는 `"999.0"` sentinel 인데, 현재 표가 `{{$r.TargetPct}}%` 로 그대로 찍는다 — 타일 문구 "999% sentinel 을 입력값으로 노출하지 않는다"와 모순(기존 결함). 행렬은 sentinel 을 "고정 목표 없음" 으로 표시한다 |
| 호버 금지 시험 | `display_primitives_test.go:565-581` — POST 폼을 품은 카드 `.tier-card`·`.choice-tile` 에 hover 효과 금지 | `.choice-tile` 을 표 열로 바꾸면 폼을 품은 새 컨테이너 선택자를 이 금지 목록에 **추가**한다(기록). `.tier-card` CSS(`templates.go:264-265`)는 a174 뒤로 쓰이는 곳이 없지만 이 change 범위가 아니므로 남긴다 |
| 손절 비교 계산기 | `hypothetical-form` define(`templates_optimization.go:472`), CSS `templates.go:276-280` `.hypothetical-fields` | 입력 수는 기능 — 배치만, 수 불변 |
| ★ 중계 카테고리 | 종목별 관리(:175-182, 링크 1 + `tech a044`), 전략·실행(:203-213, **링크 2**: lane 상태·시장·일정), 성과·이력(:215-228, 링크 1 + **설정 변경 감사 이력 표 `.Audit`**) | 성과·이력의 감사 표는 중계가 아니라 이 카테고리 고유의 값이다 → **유지**한다. "요약 1 + 링크" 규칙은 링크 수를 줄이라는 뜻이 아니다 |
| ★ 후보 필터 | :184-201. 데이터 `optimizationPage.CandidateFilterMarkets()`(`optimization_candidate_filters.go:5`) → `candidate.CandidateFilterMarkets()`(`internal/candidate/threshold_descriptors.go:88`): KR·US 각각 같은 세 키 `seen_late`·`extended`·`near_high`(:45-55, 시장마다 복제) | 행렬은 **새 메서드(새 파일)** 로 키별로 다시 묶는다. 칸 필드는 Help·DefaultState·Unit·ValidRange·Direction·SampleState·EvidenceState·ApplyTiming·MissingEvidence·PreviewContract·LegacyValue·Provenance |
| strategy runtime | `strategy-runtime` define(:433-471): 공통 투영 계약 절(:442-447), `{{range .Markets}}` `article.detail-grid`(:448-469, 필드 15행), 필드 레지스트리 `golden`(:470-471) | `strategy-runtime` spec 「KR·US 독립 projection」 — 한 시장 실패가 다른 시장 표시를 막지 않는다 |
| 신호 | `templates_signals.go` `signals`(:29) 외 define 9개, 그림자 밴드 `signalsband`(:193) | ★ "기록만 하고 판정하지 않는다 6회" 는 **렌더 기준** 값이다(소스 grep 은 `signals.go` 1회 — Go 문자열을 시장마다 반복). 1.1 에서 렌더로 다시 센다 |
| 시험 | `optimization_ui_contract_test.go`(카테고리 순서·모바일·터치·CSP·deep link), `performance_history_test.go`, `signals_test.go`, `display_primitives_test.go:565-581`, a173 계약 시험, a174 계약 시험 | |

## 2. 설계
### 2.1 내비 하나
- 개요 카테고리의 카드 그리드 삭제. 개요 본문은 "무엇을 바꿀 수 있나" 문단 + 마지막 actor/reason/digest 표만.
- KPI 6칸(`Desired/Effective/성과 근거/Effective entry/LIVE 권한/재시작`)은 `{{if .Overview}}` 안으로. 다른 카테고리는 머리에 한 줄 요약(`Desired v1 · Effective v1 · entry OFF`)만.
- 보조 링크 줄(전략 lane · 시장·일정 · 레인 성과)은 왼쪽 메뉴 하단의 "관련 화면" 소제목 아래로 합친다(링크 수 불변, 목록 하나).
### 2.2 중계 카테고리 3개
- 종목별 관리·전략·실행·성과·이력: 설명 문단은 `<p>` 요약 1줄로 줄인다(대상 화면이 계산한 값을 옮김 — 전략·실행은 `lane desired OFF · autostart OFF · scheduler OFF` 한 줄이 이미 있음). 링크는 **지금 수 그대로**다(1·2·1). 머리 KPI(`status-strip`)는 §2.1 에 따라 이 카테고리들에서 빠진다. 성과·이력의 감사 이력 표는 유지한다. 높이 ≤ 700px 는 종목별 관리·전략·실행에 적용하고, 성과·이력은 감사 행 수에 따라 달라지므로 측정값만 적는다.
### 2.3 익절 프리셋 비교표
- `<table class="preset-compare">`: 행 = T1~T4(목표 수익률 / 진입가 대비 보호 / 잔량 기준 익절 은 셀 안 3줄 또는 행 ×3), 열 = `.Policies` 순서 그대로(이름·수는 레지스트리가 정한다 — 시험은 레지스트리에서 센다). 표 아래 열마다 기존 `/optimization/exit-policy` 폼(hidden 5개·버튼 문구·`disabled` 분기 바이트 불변) 1개. "추천 · 자동 저장 안 함" 배지·`현재 desired` 표시는 열 머리. 정책별 한 줄 설명(:107-109)은 열 머리 아래 1줄. runner 의 `999.0` sentinel 칸은 "고정 목표 없음" 으로 표시한다(새 표시 함수, 값 자체는 안 바꿈).
- `ExitPolicyWritable` 이 거짓이면 폼 대신 "읽기 전용 · owner descriptor 확인 필요" 를 열마다 그대로 둔다.
- 1280·390 잘림 0: 390 에서는 `table-scroll` 가로 스크롤 허용, 셀 최소 폭 4.5rem(a172 규칙).
- `.choice-tile` 을 없애면 `display_primitives_test.go:565-581` 의 hover 금지 목록이 그 컨테이너를 잃는다 → 폼을 품은 새 컨테이너 선택자를 금지 목록에 **추가**하고(시험 확장, 기록), 그 시험 함수 편집은 gate ⑤ 번들 대상이다.
### 2.4 후보 필터 행렬
- `<table class="veto-matrix">`: 행 = seen_late · extended · near_high, 열 = KR · US, 셀 = `desired / effective / 단위·범위 / 판정 방향` 4줄. 공통 행(표본·evidence · 적용 시점 · 누락 · preview/CAS 설명 · legacy provenance)은 **KR·US 값이 바이트 단위로 같을 때만** 표 아래 1회(`<dl>`)로 합친다. 다르면 그 행은 시장별 칸에 남긴다(차이를 숨기지 않는다 — 시험이 KR≠US fixture 로 확인). 새 메서드는 새 파일에 둔다(`candidate.CandidateFilterMarkets()` 결과를 키별로 다시 묶을 뿐 값은 안 바꿈). 높이 ≤ 1200px.
### 2.5 strategy runtime 두 열
- `<div class="runtime-grid">` 2열(KR | US), 720px 이하 1열. 필드 15행은 `<dl>` 그대로. 「Shared projection contract」·「Golden field registry」는 `details.tech`(a173 define) 안으로 — 값 전부 DOM 유지. `not_observed` 는 값이므로 그대로(숫자 25 는 줄지 않는다 — 목표는 세로 적층 제거).
### 2.6 신호
- KR/US 공통 설명("기록만 하고 판정하지 않는다…", 교차 임계 설명)은 시장 블록 위 1회. 그림자 밴드 seen_late·extended 절은 `<details>`(기본 접힘) — 단 후보 수·veto 수치 표는 접지 않음.

## 3. 측정
a173 하네스. 기준선 = a173 after(`candidate-filters` 3932px, `strategy-runtime` 3120px, 중계 1186px, `/optimization` 1609px, `/signals` 3006px). 목표: 중계 ≤700, 후보 필터 ≤1200, 프리셋 표 잘림 0(1280·390 스크린샷 + `scrollWidth <= clientWidth` 또는 `table-scroll` 안), 시각 항목 "높음" 0.

## 4. 안전
`/optimization/exit-policy`·`/optimization/hypothetical` 폼의 action/필드/CSRF 전후 동일(렌더 비교 시험). preview→apply 2단계 불변. 카테고리 deep link(`?category=`) 불변(`optimization_ui_contract_test.go`).
