# a174 design

전제: a173(공통 정리) 착지 뒤의 트리에서 시작한다. 아래 file:line 은 a173 이전 정찰(2026-10-25, `b27a1ea87`) 값이라 **편집 전에 다시 확인**한다. a173 이 `templates_settings.go` 의 알림 class·'적용 후' 표·체크박스 간격을 이미 바꿨다.

## 1. 지형

| 대상 | 위치 | 비고 |
|---|---|---|
| 설정 탭 템플릿 | `templates_settings.go` `settings`(:25) → `settings-standing`(:80)·`settings-daily`(:228)·`settings-strategy`(:382)·`settings-tools`(:394) | 탭 경로·라벨·기본 진입은 spec 851 로 고정 |
| 기동 인터록 판정 | `settings_operating.go:84-122` `(settingsPage).GatePreflight() []preflightClause{Name, OK, Detail, Judgeable}`, `GateBlockers()` :126 | `settingsView`(`settings.go:172-208`)가 네 탭 모두에 `Gate`·`Trading`·`TradingWired`·`GateWired` 를 채우므로 **당일 탭 템플릿도 핸들러 변경 없이 호출 가능** |
| 설정 머리 상태 요약 | `settings_card.go:350-381` `Tone` ok/warn (게이트·엔진·한도 세 칸) | 체크리스트가 이 자리를 확장한다(spec 851 「하위 화면 머리의 운영 상태」) |
| Guardian 프리셋 | `templates_settings.go:279` `/settings/limits/preset` 폼 ×5, `:313` `/settings/limits` 고급 폼 | 프리셋 정의는 Go 쪽(`settings_limits.go:275` 핸들러가 받는 preset id) — 표의 열은 그 목록에서 만든다 |
| 거래 정책 | `templates_settings.go:361` `/settings/trading` 폼, 현재 상태 표 그 위 | 네 필드 `place`·`sell`·`cancel`·`allow_live_order_actions` 이름 불변 |
| 검증 콘솔 엔진 버튼 | `templates.go:654`(`/engine/start`, `{{if .CanStart}}`)·`:660`(`/engine/stop`) | 체크리스트를 이 두 폼 **바로 위**에 둔다. 폼·CSRF·`CanStart/CanStop` 분기 불변 |
| 외부 편입 읽기 전용 | `templates_position_policy.go` 「외부 매수 자동편입」 표 | 설명 중복만 줄이고 표·링크는 유지 |

## 2. 설계

### 2.1 기동 준비 체크리스트 (읽기 전용 define `readiness`)
- 데이터: 새 함수 `startupReadiness(gate, trading, autostart, limits…) []readinessRow{Label, State(ok|warn|unknown), Detail, Href, Owner}` 를 **새 파일** `readiness.go` 에 둔다. 행: 자동화 게이트 · 엔진 자동 시작 · Guardian 한도 5개(`GatePreflight` 1번 절의 `Validate()` 결과를 옮김) · 거래 정책 4개(3번 절) · 능력 증명서(판정 불가, 2·6) · Guardian 주입 등(판정 불가, 4·5·7·8). 기존 `GatePreflight()` 의 판정을 **옮기고 재계산하지 않는다**(spec 델타 SHALL NOT). `GatePreflight()` 본문은 편집하지 않는다(FLM 요구 회피 — 기억 `flm-exemption-only-when-no-function-modified`).
- 검증 콘솔 페이지(`dashboardPage`?)에는 `Gate`·`Trading` 이 없을 수 있다 → 핸들러 본문 편집 없이 **뷰 모델 생성자 바깥**(새 함수 `readinessForConsole(c *Console) []readinessRow`)에서 만들어 템플릿 함수로 넘긴다(a173 의 `causes` 템플릿 함수 패턴). 읽는 값은 이미 설정 화면이 읽는 파일과 같다(브로커 호출 0).
- 렌더: `<section class="readiness" aria-labelledby=…><h2>기동 준비</h2><ol>…</ol></section>` — 줄마다 `상태 배지 · 라벨 · 한 줄 사유 · <a href=소유 탭#anchor>소유 탭</a>`. 컨트롤(체크박스·버튼) 0. `TestEachSettingControlAppearsOnExactlyOneTab`(`settings_cadence_test.go:279`)의 `formBlock` 스캔은 `<form` 만 보므로 통과해야 한다 — 확인.
- 자리: 상시·당일 머리(기존 세 칸 요약을 대체하거나 그 아래), 검증 콘솔 「엔진 런타임」의 버튼 바로 위. 기존 「기동 인터록 사전 판정」 표(상시 게이트 카드)는 **같은 데이터의 두 번째 표시**가 되므로 제거하고 체크리스트 링크 `#gate` 로 대체한다 — 단 `settings_operating_test.go:217-309` 가 그 표의 문구를 고정하는지 먼저 읽고, 고정하면 문구를 체크리스트 행으로 옮긴 뒤 시험을 그 자리로 옮긴다(기록).
- 기존 "지금 켜면 기동이 거부된다 — 미충족: …" 상자(상시 게이트·자동 시작 두 곳)는 체크리스트가 같은 사실을 말하므로 **한 곳(게이트 저장 버튼 옆 한 줄)** 만 남긴다.

### 2.2 Guardian 프리셋 비교표
- 데이터: 프리셋 목록을 통화별로 묶는 뷰 모델(새 함수 `presetMatrix(presets, current) []presetColumn`) — 행 = 한도 7(1회 수량·1회 금액·개방 노출·일손실 금액·일손실 비율·하루 매수 총액·최대 보유 수) + 통화, 열 = 현재 | 프리셋들. 값은 기존 프리셋 구조체에서 옮긴다.
- 렌더: `<table class="preset-matrix">` 통화당 1개(KRW: 스모크·소액 / USD: 스모크·소액·단일 종목). 열 머리 아래 기존 `/settings/limits/preset` 폼 1개씩(hidden preset id · csrf · 버튼). "현재" 열은 버튼 없음. FX 설명은 통화당 1회 `<p class="muted">`. 방향("최초 설정 — 이전 값이 없으므로 강화도 완화도 아니다") 은 표 아래 1줄.
- 모바일(≤720px): 표를 `table-scroll` 로 감싸거나 열 수 ≤4 이므로 글자 크기 유지하고 가로 스크롤 허용(`TestEveryTableIsResponsiveOrScrolls` 규칙 확인).
- 고급(개별 값) 폼은 `<details>` 안에 그대로.
- 시험 영향: `settings_limits_test.go` 의 `limitSection`/`limitCurrent` 가 HTML 마커 사이를 자른다(정찰 C 차점) · `settings_card_test.go:215` 프리셋 폼 존재 · `:536` 부재 케이스. 마커 문자열은 유지하거나 시험을 새 구조로 옮기고 기록.

### 2.3 거래 정책 한 줄 형식
- 줄: `<div class="toggle-row"><label><input type=checkbox name=place …> 주문 제출</label> <span class="status-pill ok|warn">켜짐|꺼짐</span> <p class="muted">설명</p></div>` × 4. 현재 상태 표(`토글 | 현재`) 제거 — 배지가 그 값. 필드 이름·순서·value·hidden 필드·버튼 문구 불변.
- `TestTheScreenRendersTheFourTogglesAndTheirState`(`settings_operating_test.go:151`)가 표의 "켜짐/꺼짐" 을 어떻게 찾는지 읽고, 배지 텍스트로 옮기되 단언 수는 줄이지 않는다.

### 2.4 외부 편입 읽기 전용 표(`/position-management`)
- 표는 유지(spec 851 "다른 탭 값의 읽기 전용 표시와 링크는 요구된다"). 표 위 설명 문단은 한 줄로 줄이고 "편집은 설정·상시" 링크를 둔다. 상시 탭의 편입 카드 설명 문단도 1회로.

## 3. 깨질 수 있는 시험 (편집 전 재확인)
`settings_cadence_test.go:54,279,320,332,338,359,390,418,460` · `settings_operating_test.go:151,157,217-309,291,313` · `settings_limits_test.go`(limitSection·limitCurrent·139·362·536) · `settings_card_test.go:105,215,275` · `settings_autostart_test.go:44,190` · `engineproc_test.go:85-356`(검증 콘솔 엔진 섹션) · `screen_paths_test.go:68` · a173 신규 `a173_common_cleanup_test.go`(상자 ≤3·원인 1회·식별자 0 — 체크리스트는 상자가 아니라 본문 섹션이어야 한다).

## 4. 측정 (a173 하네스 재사용: `openspec/changes/a173-console-ux-common-cleanup/measure/`)
- 기준선: a173 after 값(`/settings/daily` 1280 높이 5237px, 상자 3, 폼 7, 입력 24).
- 목표: `/settings/daily` ≤ 2500px, 상자 ≤ 3, 인지 부하 표 결정·기억 항목 "높음" 0, 모바일 390 에서 프리셋 표 가로 스크롤은 허용하되 잘림 0.

## 5. 안전
- 토글·한도 저장 폼의 action/method/field/csrf 를 **렌더 비교 시험**으로 전후 동일 단언(폼 바이트 diff 0 또는 필드 집합 동일).
- 확인 단계·확인 문구 추가 0. 클릭 수 불변(프리셋 버튼 1회, 토글 체크 + 저장 1회).
- 엔진 시작/정지 폼 마크업 불변(spec 634·662), 체크리스트는 그 위의 읽기 전용 섹션.
