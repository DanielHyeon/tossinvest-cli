# a175 design

전제: a173 착지 뒤 트리(a174 와 병행 가능 — 주 파일은 a174 가 `templates_settings.go`·`templates.go` `dashboard` define, a175 가 `templates_optimization.go`·`templates_signals.go`). file:line 은 정찰(2026-10-25) 값, 편집 전 재확인.
**겹침 정정(Manager, 2026-10-26)**: 콘솔 CSS 는 전부 `templates.go` `style` define(:17-481) 한 곳에 있어서 a174·a175·a176 이 모두 그 블록을 고친다(계산기 `.hypothetical-fields` 도 거기 있다). 규칙: `style` **끝에** `/* a175 — … */` 블록 하나만 덧붙이고 기존 규칙은 고치지 않는다(a174 design §2.5). a174 가 먼저 착지하면 a175 의 gate ⑤ 창에 `pages.go:Console.handleDashboard` 가 든다 — a174 번들을 sha 를 대조한 뒤 복사한다. 착수 시 `base-commit.txt` 를 그때 main HEAD 로 갱신하고, 작업 위치는 `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-optimization-and-runtime-matrix-views`(사람 결정 2026-10-26).

## 1. 지형
| 대상 | 위치 | 비고 |
|---|---|---|
| 최적화 템플릿 | `templates_optimization.go:4` 단일 `optimization` 템플릿, 카테고리 분기 `{{if .ExitProtection}}…{{if .CandidateFilters}}…{{if .PerformanceHistory}}(:214)` | 왼쪽 카테고리 메뉴 `optimization-nav`, 개요 카드 그리드, KPI 6칸 히어로 |
| 카테고리 뷰 | `optimization_view.go` `optimizationCategoryViews()`(a173 이 라벨의 change ID 를 기술 정보로 옮김) | 메뉴 라벨·설명은 여기서 |
| 익절 프리셋 | `templates_optimization.go` 「청산/보호 · 익절 정책」 `.tier-card` ×3 + `/optimization/exit-policy` 폼 ×3 | `TestAmbientMotionContract` 가 `.tier-card` 를 hover 제외 목록에 둔 채 고정 — class 를 유지하거나 시험 갱신 기록 |
| 손절 비교 계산기 | `templates.go:~254` `.hypothetical-fields`, `/optimization/hypothetical` POST | 입력 10개는 기능(a154 spec) — 배치만, 수 불변 |
| 후보 필터 | `templates_optimization.go` 「후보 필터」 KR/US × 사유 3 블록 | `candidate-veto-thresholds` spec 의 표시 요구(값·사유·미측정) 확인 |
| strategy runtime | `templates_optimization.go:430` `strategy-runtime` 템플릿, `:462` Golden field registry | `strategy-runtime` spec 1726 「KR·US 독립 projection」 — 두 열로 나란히 두되 독립성(한 시장 실패가 다른 시장 표시를 막지 않음) 유지 |
| 신호 | `templates_signals.go:29` `signals` + 하위 define, `signals.go:165` 자기 문구 | a173 이 `causes` 배너를 붙였음. 그림자 밴드 두 절은 `details` 로 |
| 시험 | `optimization_ui_contract_test.go:80`(카테고리 순서·모바일·터치·CSP), `performance_history_test.go`, `signals_test.go`(한글 복사 18줄), `display_primitives_test.go:506`(hover 선택자·`.tier-card`), a173 계약 시험 | |

## 2. 설계
### 2.1 내비 하나
- 개요 카테고리의 카드 그리드 삭제. 개요 본문은 "무엇을 바꿀 수 있나" 문단 + 마지막 actor/reason/digest 표만.
- KPI 6칸(`Desired/Effective/성과 근거/Effective entry/LIVE 권한/재시작`)은 `{{if .Overview}}` 안으로. 다른 카테고리는 머리에 한 줄 요약(`Desired v1 · Effective v1 · entry OFF`)만.
- 보조 링크 줄(전략 lane · 시장·일정 · 레인 성과)은 왼쪽 메뉴 하단의 "관련 화면" 소제목 아래로 합친다(링크 수 불변, 목록 하나).
### 2.2 중계 카테고리 3개
- 종목별 관리·전략 실행·성과 이력: `<p>` 요약 1줄(대상 화면이 계산한 값을 옮김 — 전략 실행은 `lane desired OFF · autostart OFF · scheduler OFF` 한 줄이 이미 있음) + 버튼 링크 1개. KPI·카드 0. 높이 ≤ 700px.
### 2.3 익절 프리셋 비교표
- `<table class="preset-compare">`: 행 = T1~T4(목표 수익률 / 진입가 대비 보호 / 잔량 기준 익절 은 셀 안 3줄 또는 행 ×3), 열 = 균형형·러너형·하이브리드 50. 표 아래 열마다 기존 `/optimization/exit-policy` 폼 버튼 1개(열 정렬). "추천 · 자동 저장 안 함" 배지는 열 머리.
- 1280·390 잘림 0: 390 에서는 `table-scroll` 가로 스크롤 허용, 셀 최소 폭 4.5rem(a172 규칙).
- `.tier-card` 를 없애면 `TestAmbientMotionContract` 의 hover 제외 목록 단언이 바뀐다 → 폼을 품은 새 컨테이너 class 를 제외 목록에 넣고 시험을 **확장**(기록).
### 2.4 후보 필터 행렬
- `<table class="veto-matrix">`: 행 = seen_late · extended · near_high, 열 = KR · US, 셀 = `desired / effective / 단위·범위 / 판정 방향` 4줄. 공통 행(표본·evidence · 적용 시점 · 누락 · preview/CAS 설명 · legacy provenance)은 표 아래 1회(`<dl>`). 높이 ≤ 1200px.
### 2.5 strategy runtime 두 열
- `<div class="runtime-grid">` 2열(KR | US), 720px 이하 1열. 필드 15행은 `<dl>` 그대로. 「Shared projection contract」·「Golden field registry」는 `details.tech`(a173 define) 안으로 — 값 전부 DOM 유지. `not_observed` 는 값이므로 그대로(숫자 25 는 줄지 않는다 — 목표는 세로 적층 제거).
### 2.6 신호
- KR/US 공통 설명("기록만 하고 판정하지 않는다…", 교차 임계 설명)은 시장 블록 위 1회. 그림자 밴드 seen_late·extended 절은 `<details>`(기본 접힘) — 단 후보 수·veto 수치 표는 접지 않음.

## 3. 측정
a173 하네스. 기준선 = a173 after(`candidate-filters` 3932px, `strategy-runtime` 3120px, 중계 1186px, `/optimization` 1609px, `/signals` 3006px). 목표: 중계 ≤700, 후보 필터 ≤1200, 프리셋 표 잘림 0(1280·390 스크린샷 + `scrollWidth <= clientWidth` 또는 `table-scroll` 안), 시각 항목 "높음" 0.

## 4. 안전
`/optimization/exit-policy`·`/optimization/hypothetical` 폼의 action/필드/CSRF 전후 동일(렌더 비교 시험). preview→apply 2단계 불변. 카테고리 deep link(`?category=`) 불변(`optimization_ui_contract_test.go`).
