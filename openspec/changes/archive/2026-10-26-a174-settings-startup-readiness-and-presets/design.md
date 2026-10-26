# a174 design

전제: a173(공통 정리) 착지 뒤의 트리에서 시작한다. §1 표는 Manager 가 HEAD `4e20c5df5`(a173 착지 + 5.1 커밋, 2026-10-25 23:50, codegraph 1.6.0 원문 + grep)에서 다시 잰 값이다. 착수 시 base 가 바뀌면 1.1 에서 한 번 더 확인한다.

## 1. 지형 (HEAD `4e20c5df5` 재확인)

| 대상 | 위치 | 비고 |
|---|---|---|
| 설정 탭 템플릿 | `templates_settings.go` `settings`(:25) → `settings-standing`(:80)·`settings-daily`(:219)·`settings-strategy`(:370)·`settings-tools`(:382) | 탭 경로·라벨·기본 진입은 spec 851 로 고정 |
| 기동 인터록 판정 | `settings_operating.go:84-122` `(settingsPage).GatePreflight() []preflightClause{Name, OK, Detail, Judgeable}`, `GateBlockers()` :126 | 절은 네 개다: `1. 한도 다섯 개와 통화`(`Limits().Validate()` 오류 하나) · `3. 거래 정책`(`TradingPolicyOf(..).Complete()`, 꺼진 이름은 `Missing()`) · `2·6. 능력 증명서`(판정 불가) · `4·5·7·8. …`(판정 불가). `settingsView`(`settings.go:172-235`)가 네 탭 모두에 `Gate`·`Trading`·`Autostart` 를 채우므로 설정 탭은 핸들러 변경 없이 호출할 수 있다 |
| 거래 정책 표시 값 | `settings_operating.go:141` `(settingsPage).TradingRows() []toggleRow{Key, Label, On, Why}` | 토글 넷의 **상태 표시**. 판정이 아니다 |
| 설정 머리 상태 요약 | `settings_card.go:362-387` `(settingsPage).OperatingState() []stateChip`(게이트·엔진·한도 세 칸, `Tone` ok/warn) | 체크리스트가 이 자리를 확장한다(spec 851 「하위 화면 머리의 운영 상태」) |
| 기존 사전 판정 표 | `templates_settings.go:145-149` 「기동 인터록 사전 판정」 `{{range .GatePreflight}}` | 체크리스트가 같은 데이터를 말하므로 §2.1 에서 대체 |
| Guardian 프리셋 | `templates_settings.go:268-296` `{{range .LimitTierCards}}` → `.tier-card` 안 `/settings/limits/preset` 폼(hidden **`tier`** = `.ID`), `:304` `/settings/limits` 고급 폼 | 목록은 `config.GuardianTiers()`, 카드 뷰 모델은 `settings_limits.go:232` `LimitTierCards()`. 각 카드의 `Preview.Rows`(`settings_card.go:286` `previewAgainst`)가 이미 **한도 다섯 개**의 현재(`From`)→프리셋(`To`) 값을 갖는다 |
| 거래 정책 폼 | `templates_settings.go:347` 현재 상태 표(`{{range .TradingRows}}`), `:352` `/settings/trading` 폼 | 네 필드 `place`·`sell`·`cancel`·`allow_live_order_actions` 이름 불변 |
| 검증 콘솔 | `pages.go:80-91` `dashboardPage`(Gate·Trading·Autostart **없음**), `pages.go:111-126` `handleDashboard`, 템플릿 `templates.go` `dashboard` define(:638-801) 「엔진 런타임」 | 엔진 폼 `templates.go:693`(`/engine/start`, `{{if .CanStart}}`)·`:700`(`/engine/stop`). 폼·CSRF·`CanStart/CanStop` 분기 불변 |
| 템플릿 함수 | `pages.go:38` `pageFuncs` — **패키지 전역** FuncMap, `*Console` 없음. `pages.go:67` 에서 한 번 parse | 템플릿 함수로는 seam 을 읽을 수 없다(§2.1) |
| 이름 금지 시험 | `engineproc_test.go:194` `TestTheConsoleDecidesNothingAboutTheGate` | `Interlock`·`ProtectionReady` 는 모든 파일에서, `AutomationGate`·`automation_gate` 는 `settings.go`·`settings_limits.go`·`templates_settings.go` 밖에서 금지(주석 제외). 새 파일 `readiness.go` 와 `templates.go` 는 이 철자를 쓰면 안 된다 |
| 외부 편입 읽기 전용 | `templates_position_policy.go` `position-policy` define(:4), 표 캡션 「자동편입 설정의 기본값·저장값·실행값」(:24), 설명 문단(:40·:80) | 설명 중복만 줄이고 표·링크는 유지 |

## 2. 설계

### 2.1 기동 준비 체크리스트 (읽기 전용 define `readiness`)
- 데이터: 새 메서드 `(p settingsPage) StartupReadiness() []readinessRow{Label, State(ok|warn|unknown), Detail, Href}` 를 **새 파일** `readiness.go` 에 둔다. 행은 **절 단위 여섯**이다:
  1. 자동화 게이트 — `p.Gate.Enabled` 표시(켜짐/꺼짐), 링크 `/settings/standing#gate`
  2. 엔진 자동 시작 — `p.Autostart` 표시, 링크 상시 탭 자동 시작 카드
  3. Guardian 한도 다섯과 통화 — `GatePreflight()` 1번 절의 `OK`·`Detail` 그대로, 링크 `/settings/daily` 한도 카드
  4. 거래 정책 — `GatePreflight()` 3번 절의 `OK`·`Detail` 그대로(꺼진 토글 이름이 `Detail` 에 이미 있다), 링크 당일 탭 거래 정책
  5. 능력 증명서 — 판정 불가(2·6 절), 링크 없음
  6. Guardian 주입 · 게이트웨이 · 키 운반 — 판정 불가(4·5·7·8 절), 링크 없음
  행 3·4·5·6 은 `GatePreflight()` 를 **호출해 옮길 뿐** 판정하지 않는다(spec 델타 SHALL NOT 재계산). 한도를 다섯 행으로 쪼개지 않는다 — 쪼개려면 필드별 판정을 새로 해야 하고 그것이 재계산이다. 체크리스트 전체의 합산 판정("기동 가능")은 **만들지 않는다**: 이 패키지는 엔진이 돌아도 되는지 계산하지 않는다(`TestTheConsoleDecidesNothingAboutTheGate` 주석 :177-193). `GatePreflight()` 본문은 편집 0.
- 이름 규칙: `readiness.go`·`templates.go` 는 `AutomationGate`·`automation_gate`·`Interlock`·`ProtectionReady` 를 철자로 쓰지 않는다(§1 이름 금지 시험). 값은 `p.Gate` 필드로 받는다. `readiness` define 은 금지 면제 파일인 `templates_settings.go` 에 둔다.
- 검증 콘솔: `pageFuncs` 는 전역이라 `*Console` 이 없고 `dashboardPage` 에는 Gate·Trading·Autostart 가 없다. 그래서 **템플릿 함수 패턴은 쓸 수 없다.** 대신:
  - `dashboardPage` 에 필드 `Readiness []readinessRow` 를 더한다(구조체 필드 추가 — 함수 본문 아님).
  - 새 메서드 `(c *Console) startupReadiness() []readinessRow` 를 `readiness.go` 에 둔다. `c.opts.Limits`·`c.opts.TradingPolicy`·`c.opts.EngineBoot` 를 `settingsView` 와 같은 방식으로 읽어 `settingsPage{Gate: …, Trading: …, Autostart: …}.StartupReadiness()` 를 돌려준다. 읽기 실패나 미배선은 그 행을 `unknown` + 사유로 둔다(fail-closed 표시, 0값을 "통과"로 그리지 않는다 — a177 교훈).
  - `handleDashboard`(`pages.go:111-126`) 의 리터럴에 `Readiness: c.startupReadiness(),` **한 줄**을 더한다. **기존 함수 편집 1 → gate ⑤ FLM 번들 1**(tasks 3.1, 편집 전에 작성).
  - 비용: 검증 콘솔은 2초마다 다시 읽는다(`RefreshSeconds`). 읽기는 설정 파일 파싱 3회이고 브로커 호출 0 — `settingsView` 주석(:168-171)의 판단과 같다.
- 렌더: `<section class="readiness" aria-labelledby=…><h2>기동 준비</h2><ol>…</ol></section>` — 줄마다 `상태 배지 · 라벨 · 한 줄 사유 · <a href=소유 탭#anchor>소유 탭</a>`. 컨트롤(체크박스·버튼) 0. `TestEachSettingControlAppearsOnExactlyOneTab`(`settings_cadence_test.go:279`)의 `formBlock` 스캔은 `<form` 만 보므로 통과해야 한다 — 확인.
- 자리: 상시·당일 머리(기존 세 칸 요약을 대체하거나 그 아래), 검증 콘솔 「엔진 런타임」의 버튼 바로 위. 기존 「기동 인터록 사전 판정」 표(상시 게이트 카드)는 **같은 데이터의 두 번째 표시**가 되므로 제거하고 체크리스트 링크 `#gate` 로 대체한다 — 단 `settings_operating_test.go:217-309` 가 그 표의 문구를 고정하는지 먼저 읽고, 고정하면 문구를 체크리스트 행으로 옮긴 뒤 시험을 그 자리로 옮긴다(기록).
- 기존 "지금 켜면 기동이 거부된다 — 미충족: …" 상자는 settings_card.go:151·169 의 cardcautions 이고(템플릿 grep 0 은 Manager 오류 — 문구가 Go 문자열에 있었다), 저장 버튼 옆 결과 안내라서 남긴다(a173 결과 등급, 7개 카드 공유 정의). 체크리스트가 대체하는 것은 :145-149 사전 판정 표 하나다(Manager 4.5 결정 2026-10-26).

### 2.2 Guardian 프리셋 비교표
- 데이터: 새 메서드 `(p settingsPage) LimitTierMatrix() []tierMatrix`(새 파일, 통화별로 묶음)가 기존 `LimitTierCards()` 를 **그대로 받아** 통화(`Limits.Currency`)별로 나눈다. 행 = 한도 **다섯**(`previewAgainst` 의 `Rows` 와 같은 순서·라벨: 1회 주문 수량·1회 주문 금액·계좌 개방 노출·일일 손실 금액·일일 손실 비율) + 통화, 열 = 현재(`Rows[i].From`) | 프리셋들(`Rows[i].To`). 값·방향(`Axis`)·`FirstTime`·`CurrencyChange`·`Consequence` 는 카드 뷰 모델에서 옮기고 다시 계산하지 않는다. 열의 수와 이름은 `config.GuardianTiers()` 가 정한다 — 시험은 레지스트리에서 열을 세고 문서 값을 고정하지 않는다. base `28b438293` 의 실제 목록(`internal/config/limits.go:86-146`): KRW 표 = 현재 | `kr-smoke` | `kr-small-live`(권장), USD 표 = 현재 | `us-smoke` | `us-small-live` | `us-single-name`. 따라서 표 2개, 프리셋 폼 5개(지금과 같은 수).
- 렌더: `<table class="preset-matrix">` 통화당 1개. 열 머리 아래 기존 `/settings/limits/preset` 폼 1개씩 — hidden `csrf`·hidden **`tier`**(= 카드 `.ID`)·버튼 문구(`{{.Label}}{{if .Recommended}} · 권장{{end}} 기록`) 그대로. "현재" 열은 버튼 없음. 통화 결과 설명(`Consequence`)은 통화당 1회 `<p class="muted">`. 방향("최초 설정 — 이전 값이 없으므로 강화도 완화도 아니다")·통화 변경 문장은 표 아래 1줄. `data-tier`·`data-limit-axis`·`data-currency-consequence` 속성을 시험이 찾으면 옮겨 단다(1.1 에서 확인).
- 모바일(≤720px): 표를 `table-scroll` 로 감싸거나 열 수 ≤4 이므로 글자 크기 유지하고 가로 스크롤 허용(`TestEveryTableIsResponsiveOrScrolls` 규칙 확인).
- 고급(개별 값) 폼은 `<details>` 안에 그대로.
- 시험 영향: `settings_limits_test.go` 의 `limitSection`/`limitCurrent` 가 HTML 마커 사이를 자른다(정찰 C 차점) · `settings_card_test.go:215` 프리셋 폼 존재 · `:536` 부재 케이스. 마커 문자열은 유지하거나 시험을 새 구조로 옮기고 기록.

### 2.3 거래 정책 한 줄 형식
- 줄: `<div class="toggle-row"><label><input type=checkbox name=place …> 주문 제출</label> <span class="status-pill ok|warn">켜짐|꺼짐</span> <p class="muted">설명</p></div>` × 4. 현재 상태 표(`토글 | 현재`) 제거 — 배지가 그 값. 필드 이름·순서·value·hidden 필드·버튼 문구 불변.
- `TestTheScreenRendersTheFourTogglesAndTheirState`(`settings_operating_test.go:151`)가 표의 "켜짐/꺼짐" 을 어떻게 찾는지 읽고, 배지 텍스트로 옮기되 단언 수는 줄이지 않는다.

### 2.4 외부 편입 읽기 전용 표(`/position-management`)
- 표는 유지(spec 851 "다른 탭 값의 읽기 전용 표시와 링크는 요구된다"). 표 위 설명 문단은 한 줄로 줄이고 "편집은 설정·상시" 링크를 둔다. 상시 탭의 편입 카드 설명 문단도 1회로.

### 2.5 파일 겹침 (a175·a176 과 병행할 때)
- 콘솔 CSS 는 전부 `templates.go` 의 `style` define(:17-481) 한 곳에 있다. a174(`.readiness`·`.preset-matrix`·`.toggle-row`)·a175·a176 이 모두 이 블록을 고친다. 그리고 a174 는 `templates.go` 의 `dashboard` define, a176 은 같은 파일의 `verify` define 을 고친다. "파일이 겹치지 않는다"는 사실이 아니다.
- 규칙: 각 change 는 `style` **끝에** `/* a174 — 기동 준비·프리셋 비교표·거래 정책 줄 */` 로 시작하는 자기 블록 하나만 덧붙이고, 기존 규칙은 고치지 않는다(필요하면 더 구체적인 선택자로 덮는다). 그러면 rebase 충돌은 블록 끝 한 자리로 줄어든다.
- gate ⑤ 창 상속: a174 가 `handleDashboard` 를 고치면, a174 보다 나중에 착지하는 change 의 창에도 그 함수가 든다. 그쪽은 a174 의 번들을 소스 sha 를 대조한 뒤 복사한다(a173 이 a177 번들 3개를 복사한 선례).

## 3. 깨질 수 있는 시험 (편집 전 재확인)
`engineproc_test.go:194,226`(이름 금지·면제 목록 — 새 파일이 금지 철자를 쓰면 FAIL, 면제 목록은 넓히지 않는다) · `settings_cadence_test.go:54,279,320,332,338,359,390,418,460` · `settings_operating_test.go:151,157,217-309,291,313` · `settings_limits_test.go`(limitSection·limitCurrent·139·362·536) · `settings_card_test.go:105,215,275` · `settings_autostart_test.go:44,190` · `engineproc_test.go:85-356`(검증 콘솔 엔진 섹션) · `screen_paths_test.go:68` · a173 신규 `a173_common_cleanup_test.go`(상자 ≤3·원인 1회·식별자 0 — 체크리스트는 상자가 아니라 본문 섹션이어야 한다).

## 4. 측정 (a173 하네스 재사용: `openspec/changes/a173-console-ux-common-cleanup/measure/`)
- 기준선: a173 after 값(`/settings/daily` 1280 높이 5237px, 상자 3, 폼 7, 입력 24).
- 목표: `/settings/daily` ≤ 2800px(정의 때 2500 — proposal 완료 기준 조정 참조), 상자 ≤ 3, 인지 부하 표 결정·기억 항목 "높음" 0, 모바일 390 에서 프리셋 표 가로 스크롤은 허용하되 잘림 0.

## 5. 안전
- 토글·한도 저장 폼의 action/method/field/csrf 를 **렌더 비교 시험**으로 전후 동일 단언(폼 바이트 diff 0 또는 필드 집합 동일).
- 확인 단계·확인 문구 추가 0. 클릭 수 불변(프리셋 버튼 1회, 토글 체크 + 저장 1회).
- 엔진 시작/정지 폼 마크업 불변(spec 634·662), 체크리스트는 그 위의 읽기 전용 섹션.
