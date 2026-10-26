# a174 tasks

Story: `STORY-TOS-a174` · Low~Medium(설정 폼 마크업 재배치, 설정 저장 핸들러·POST 불변). 변이 원장·다중 적대 리뷰 `not-applicable`(proposal). FLM/BTM 은 기존 함수 `pages.go:Console.handleDashboard` **한 개**에만 요구된다(한 줄 추가, gate ⑤). 나머지는 **새 함수·새 파일**로 구현하고 다른 기존 함수 편집은 0 이 목표다 — 더 필요하면 멈추고 Manager 에게 함수 이름을 보고한다. 착수 조건: a173 archive·push. 작업 위치: `/mnt/D/Axipient/workspace/TossOS.worktrees/a174-settings-startup-readiness-and-presets`(저장소 **밖** 형제 디렉터리, 사람 결정 2026-10-25. 브랜치 `feat/a174-settings-startup-readiness-and-presets`, base = `base-commit.txt`).

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · spec 델타(`specs/operator-console/spec.md` ADDED 3) · proposal-freeze 경량 리뷰(`review.md`, 착수 시 Manager 가 base 를 적는다)

## 1. 기준선·확인
- [x] 1.1 (Manager) `base-commit.txt` 를 a173 archive 뒤 main HEAD 로 갱신. design §1 표(HEAD `4e20c5df5` 값)를 그 base 에서 다시 확인하고 달라진 줄만 고친다. `config.GuardianTiers()` 의 통화별 실제 목록을 design §2.2 에 적는다. `GatePreflight` 의 `tools/logic-map` AST 를 증거로 둔다(편집 0, design §1 의 절 네 개 주장의 근거)
  - 증거(Manager Opus 5.5, 2026-10-26): base = `28b438293`(a173 archive, origin/main 과 같음). `git diff --stat 4e20c5df5 28b438293 -- internal/ cmd/` 빈 출력 → design §1 줄 번호 변경 없음. 프리셋 목록 = KRW `kr-smoke`·`kr-small-live`(권장), USD `us-smoke`·`us-small-live`·`us-single-name`(`internal/config/limits.go:86-146`) → design §2.2. AST = `analysis/evidence/GatePreflight.ast.json`(source sha `2a29caa23e01…`, :84-122): 분기 B1 if(:88)·B2 else(:92)·B3 switch(:100)·B4 case(:101)·B5 case(:106), return 1개, 판정 불가 절 두 개는 무조건 append → 절은 언제나 넷. FLM 번들 경로(`analysis/function-logic/`)가 아니라 `evidence/` 에 둔 이유: 이 함수는 편집 대상이 아니라서 번들을 만들 이유가 없고, 미완성 번들 디렉터리는 gate ⑤ 를 깬다
- [x] 1.2 a173 하네스로 기준선 촬영(`measure/baseline-<utc>.json`: 22화면 × 1280/390 × light/dark, 수치만) · `python3 tools/logic-map/check_analysis.py --change a174-…` 로 required 0 확인(편집 전)
  - 증거(Terra, 2026-10-26): `analysis/measure/baseline-20261025T230744Z.json`(96 metrics; `/settings/daily` light/1280 = 높이 5237px·상자 3·폼 7·입력 24), 격리 `shoot.cjs`는 빈 설정·자격증명 0·38321 포트 GET만으로 실행했고 PID 2377970 종료 뒤 포트 해제 확인. dashboard light/1280 screenshot 1장은 `Page.captureScreenshot` 프로토콜 오류로 누락했고 수치 JSON은 완성됨. 편집 전 gate 원문 `analysis/pre-edit-analysis.txt`에서 required 0 확인(분석 번들 미생성 표시는 사전 상태임).

## 2. RED
- [x] 2.1 `TestStartupReadinessAppearsOnThreeScreensWithoutControls`: 상시·당일·검증 콘솔 렌더에 `section.readiness` 1개, 그 안 `<form`·`<input`·`<button` 0, 행 = **절 단위 여섯**(게이트 · 자동 시작 · 한도 다섯과 통화 · 거래 정책 · 판정 불가 2 — design §2.1), 판정 가능한 줄에 소유 탭 링크, 합산 판정 문구("기동 가능" 류) 0, 검증 콘솔에서는 `/engine/start` 폼보다 **앞**에 위치. 한도·거래 정책 줄의 상태·사유가 같은 fixture 의 `GatePreflight()` 결과와 같음(옮김 단언). seam 미배선·읽기 실패 fixture 에서 해당 줄이 `unknown` + 사유(0값을 통과로 그리지 않음)
  - 증거(Terra, 2026-10-26): 새 시험은 production 편집 전 `analysis/red-output.txt`에서 section 부재로 FAIL, 구현 뒤 `go test ./internal/console -run 'Test(StartupReadinessAppearsOnThreeScreensWithoutControls|GuardianPresetsRenderAsOneMatrixPerCurrency|TradingPolicyRowsAppearOnce|SettingsFormsAreByteIdenticalAfterLayout)$' -count=1` PASS.
- [x] 2.2 `TestGuardianPresetsRenderAsOneMatrixPerCurrency`: 당일 렌더에 `table.preset-matrix` 가 `config.GuardianTiers()` 의 통화 수만큼, 행 = 한도 다섯 + 통화, 열 = 현재 + 그 통화 프리셋, 프리셋 열마다 `/settings/limits/preset` 폼 1개(hidden `tier` = 프리셋 ID), 통화 결과 설명 통화당 1회, 칸 값이 같은 fixture 의 `LimitTierCards()` `Preview.Rows` From/To 와 같음, 1280 에서 표 폭 ≤ 본문 폭(하네스로 측정 보조)
  - 증거(Terra, 2026-10-26): `internal/console/readiness_test.go`가 registry 카드에서 통화 수·프리셋 ID·Preview From/To를 읽어 비교함. RED/GREEN 명령은 2.1 증거와 같음.
- [x] 2.3 `TestTradingPolicyRowsAppearOnce`: `.toggle-row` 4개, 각 행 안에 체크박스·배지·설명이 함께, 현재 상태 표 0, 필드 이름 집합 == {place, sell, cancel, allow_live_order_actions}
  - 증거(Terra, 2026-10-26): `internal/console/readiness_test.go` RED/GREEN; 기존 상태 표 부재와 네 필드 집합을 단언함.
- [x] 2.4 `TestSettingsFormsAreByteIdenticalAfterLayout`: 폼 action/method/필드명/hidden 값 집합이 a173 착지 렌더와 동일(기준 스냅숏을 시험 fixture 로 고정)
  - 증거(Terra, 2026-10-26): `a173SettingsFormContracts` 고정 fixture가 `/settings/save`·gate·autostart·limits·trading·5 preset의 method/action/필드/hidden csrf·tier를 비교함. RED/GREEN 명령은 2.1 증거와 같음.
- [x] 2.5 a173 계약 시험(상자 ≤3·원인 1회·식별자 0·경계 배지) 가 설정 세 탭에서 계속 통과하는지 먼저 확인(RED 가 아니라 회귀 가드)
  - 증거(Terra, 2026-10-26): `go test ./internal/console -run 'Test(EachUnmeasuredCauseIsExplainedOnce|FloatingNoticesAreAtMostThree|OperatorBodyCarriesNoDeveloperIdentifiers|EveryScreenHeadHasScopeBadgeAndFoldedBoundary|StatusStripCellsAreOneLine|ProseWidthAndToggleSpacingRulesExist|DisclosureMarkerSurvivesTheMobileBlock)$' -count=1` PASS(9.538s).

## 3. GREEN (design §2 순서)
- [x] 3.1 **편집 전에** `pages.go:Console.handleDashboard` FLM 번들(`tools/logic-map`, `analysis/function-logic/`)을 만든다. 그다음 `readiness.go`(새 파일: `settingsPage.StartupReadiness`·`Console.startupReadiness`, 금지 철자 0 — design §1 이름 금지 시험) + `readiness` define(`templates_settings.go`) + `dashboardPage.Readiness` 필드 + `handleDashboard` 한 줄 + 세 자리 배치 + 기존 사전 판정 표(`templates_settings.go:145-149`) 정리. CSS 는 `templates.go` `style` 끝 a174 블록에만(design §2.5)
  - 증거(Terra, 2026-10-26): 편집 전 번들은 `analysis/function-logic/internal-console--console.handledashboard/`에 작성했고, 구현 후 `go run ./tools/logic-map --file internal/console/pages.go --func Console.handleDashboard`와 risk report로 AST SHA `d676ff005fdae0bfa6a4ef2b49415f7b7d99029b2bf68067a8b8776838244096`를 갱신함. `pages.go` 기존 함수 본문 변경은 `Readiness: c.startupReadiness(),` 한 줄뿐임.
- [x] 3.2 프리셋 비교표(뷰 모델 새 함수 + 템플릿 + CSS `.preset-matrix`, 모바일 규칙)
  - 증거(Terra, 2026-10-26): `LimitTierMatrices`는 `LimitTierCards()`의 Preview 값을 통화별로만 옮기고, 모바일 `table-scroll`·`min-width` 규칙을 a174 CSS 블록에 추가함.
- [x] 3.3 거래 정책 한 줄 형식
  - 증거(Terra, 2026-10-26): 기존 현재값 표를 제거하고 동일 `/settings/trading` form 안에서 `toggle-row` 네 줄에 checkbox·배지·설명을 함께 렌더함.
- [x] 3.4 외부 편입 설명 중복 정리(상시·`/position-management`)
  - 증거(Terra, 2026-10-26): 상시 탭은 기존 단일 편입 설명을 유지하고, `/position-management` 읽기 전용 표의 설명은 설정·상시 소유 링크 한 줄로 축소함.
- [x] 3.5 `gofmt -l`·`go vet ./internal/console/`·`go test ./internal/console/ -count=1`·`go test ./cmd/tossctl/ -run Console -count=1`. 고정 리터럴을 바꾼 시험은 아래에 전수 기록. `check_analysis.py` 결과(required N) 기록
  - 증거(Terra, 2026-10-26): `gofmt -l internal/console` 빈 출력, `go vet ./internal/console/` PASS, `go test ./internal/console/ -count=1` PASS, `go test ./cmd/tossctl/ -run Console -count=1` PASS. `check_analysis.py`는 required **1**(`Console.handleDashboard`) 및 `evidence complete or diff-proven exempt`를 출력함(착지 커밋 전이라 landing record 없음).

## 4. 측정·리뷰
- [x] 4.1 하네스 재촬영 → `measure/after-<utc>.json`, 전후 표(`/settings/daily` 높이·상자·폼·입력, 세 화면 체크리스트 유무)
  - 증거(Terra, 2026-10-26): `analysis/measure/after-20261026T013100Z.json`(96 metrics)·`after-20261026T013100Z-outline.json`. `/settings/daily` light/1280은 **5237→2445px**, 상자 3→3, form 7→7, input 24→24, button 7→7; `/settings/standing`·`/settings/daily`·`/verify-console`에 readiness를 렌더함. 격리 콘솔은 빈 설정·자격증명 0·38321 포트 GET만으로 실행했고 parent 2806550, child 2807015 종료 뒤 `port_38321_released=true` 확인.
- [x] 4.2 `cognitive-load-assessment` 설정 3탭 + 검증 콘솔 표(`analysis/cognitive-load-after.md`), 결정·기억 "높음" 0
  - 증거(Terra, 2026-10-26): `analysis/cognitive-load-after.md`; 전용 도구 부재를 기록하고 렌더 원문·격리 1280/390 화면·콘솔 규칙으로 수동 점검. 설정 3탭과 검증 콘솔 모두 결정·기억 "높음" 0.
- [x] 4.3 `design-critic` 재점수 + `fixing-accessibility`(체크박스 라벨·표 머리·링크 텍스트) 점검 → `analysis/review-after.md`
  - 증거(Terra, 2026-10-26): `analysis/review-after.md`; 전용 design-critic 미설치 사실과 수동 점검 범위를 기록함. `impeccable detect`는 변경 대상에서 `[]`, 접근성·모션·데스크톱/모바일 격리 렌더를 확인함.
- [x] 4.4 gstack `review`(경량) → `review.md` 「구현 리뷰」. 변이 1회: `StartupReadiness` 에서 거래 정책 줄 제거 → 2.1 FAIL 확인 후 원복(기록, 원복은 심볼 수를 세어 확인)
  - 증거(Terra, 2026-10-26): `review.md` 「구현 리뷰」. 거래 정책 행과 해당 지역 변수를 함께 제거한 변이에서 2.1이 행 수 5·거래 정책 부재로 FAIL했고, 원복 뒤 `거래 정책 행`·`tradingKnown` 심볼 수가 각각 1, 대상 시험 PASS를 확인함.
- [x] 4.6 Manager 리뷰 보정: 현재 계좌 통화 결과 notice와 프리셋 가정 결과를 분리하고, 프리셋 form의 가짜 preview 속성을 제거한다. USD·미설정 current notice RED 후 GREEN, 각 티어 실제 표 열의 다섯 값·방향과 표 아래 전환 문장을 단언한다. 정상 fixture의 네 preflight 절 이름 매핑 이탈도 회귀 시험으로 막는다.
  - 증거(Terra, 2026-10-26): `analysis/red-output.txt` 끝의 원문 RED(USD·미설정 current notice 0회, 가짜 form 속성, 표 아래 전환 문구 부재). `internal/console/settings_card_test.go:TestCurrentLimitCurrencyConsequenceMatchesConfiguredCurrency`, `settings_limits_test.go:TestThePresetControlsAskForNoTyping`, `readiness_test.go:TestGuardianPresetsRenderAsOneMatrixPerCurrency`, `readiness_test.go:TestStartupReadinessMapsEveryPreflightClause`의 집중 GREEN PASS. `check_analysis.py` required **5**(`Console.handleDashboard`, `TestEveryCardShowsWhatChangesAndWhen`, `TestTheCurrencyConsequenceIsRenderedAndNotFolded`, `TestThePresetControlsAskForNoTyping`, `limitCurrent`)와 `evidence complete or diff-proven exempt` 확인. 첫 보정 렌더 `analysis/measure/after-20261026T011645Z.json`은 2544px로 목표를 44px 초과해 프리셋 셀의 세로 여백만 a174 CSS 블록에서 제거함. 최종 `after-20261026T012343Z.json`·`after-20261026T012343Z-outline.json`은 `/settings/daily` light/1280 = **2499px**, 상자 **3**, form 7·input 24·button 7, 식별자·미측정 0을 기록함. 빈 설정·자격증명 0의 38321 GET 전용 격리 콘솔 parent 3135467·child 3144850은 종료됐고 `port_38321_released=true`을 재확인함. 최종 `gofmt -l internal/console` 빈 출력, `go vet ./internal/console/` PASS, `go test ./internal/console/ -count=1` PASS(59.845s), `go test ./cmd/tossctl/ -run Console -count=1` PASS(4.897s), strict OpenSpec PASS, `impeccable detect` `[]`을 확인함.
- [x] 4.5 Manager 검증 — 크롬으로 상시·당일·검증 콘솔 × 데스크톱/모바일 × 라이트/다크 직접 확인, 버튼 0
  - 증거(Manager 실측 2026-10-26 + 라운드 2 반영): 격리 콘솔 127.0.0.1:38322·빈 config-dir·임시 HOME·HEAD `aa2864abe`에서 클릭 0으로 확인함. `/settings/daily` 1280 라이트는 높이 2500·상자 3·form 7·input 24·button 7·체크리스트 6행·컨트롤 0, 현재 통화 notice는 "아직 없다" 문장. `/verify-console` 체크리스트는 시작 폼 앞, 390 다크는 가로 넘침 0·비교표는 table-scroll 안에서만 스크롤. PID 3277705 종료와 38322 포트 해제를 Manager가 확인함. Terra는 모바일 explain 펼침 표식·표 여백/버튼 폭을 라운드 2로 보정함. a173 하네스+추가 scrollWidth 측정 `analysis/measure/after-20261026T020606Z.json`·`-outline.json`·`-mobile-dark.json`에서 `/settings/daily` 1280 라이트 = 높이 **2786**·상자 3·form 7·input 24·button 7, 390 다크 = 높이 3713·상자 3, document scrollWidth/clientWidth = 390/390(페이지 가로 넘침 0), 표 region scrollWidth/clientWidth = 640/366·858/366(두 표 모두 region 안에서만 스크롤), HTTP 200·식별자/미측정 0을 기록함. 높이는 정보 여백 복원으로 2500을 넘었고 Manager 결정대로 억지로 다시 누르지 않음. 빈 설정·임시 HOME·자격증명 0의 38323 GET 전용 격리 콘솔 parent 3361555·child 3372130은 종료됐고 `port_38323_released=true`을 재확인함. 4.5 체크는 Manager 소유라 유지함.
  - Manager 판정(Opus 5.5, 2026-10-26): 라운드 1(aa2864abe)·2(54ce02572) diff 를 읽었다. 크롬 실측 두 번을 했다 — 1차 HEAD aa2864abe, 2차 HEAD 54ce02572, 둘 다 격리 콘솔 38322·빈 config-dir·임시 HOME·클릭 0. 확인 범위: `/settings/daily` 1280 라이트·다크(2차 다크 높이 2763·상자 3), 390 다크(가로 넘침 0, 비교표만 table-scroll), `/settings/standing` 1280(체크리스트 1·사전 판정 표 0), `/verify-console` 1280(체크리스트 6행·컨트롤 0·시작 폼 앞·시작/정지 폼 유지), 콘솔 오류 0. 두 번 다 PID 종료와 포트 해제를 확인했다. 모바일 라이트와 상시·검증 콘솔의 모바일은 직접 보지 않았다(Terra 하네스 96 측정으로 대신함). 완료 기준 2500→2800 조정은 proposal 에 적었다. 판정: 승인, 5.1 종결로 넘김.
  - 최종 검증(Terra, 2026-10-26): `gofmt -l internal/console`·`go vet ./internal/console/`·`git diff --check` 빈 출력, `go test ./internal/console/ -count=1` PASS(60.995s), `go test ./cmd/tossctl/ -run Console -count=1` PASS(3.924s), `check_analysis.py --change a174-settings-startup-readiness-and-presets` required **5**·증거 완비, strict OpenSpec PASS, `/home/daniel/.codex/skills/impeccable/scripts/impeccable detect --json --target internal/console/templates.go` `[]`.

## 5. 종결 (Manager 승인 뒤, Opus)
- [x] 5.1 rebase → FLM 번들(required 가 0 이 아니면) → tracker → sdd-sync → `make gate` → archive → push → retain
  - 사람 승인: 사용자가 a174 종결 1~6(Manager 편집 커밋·rebase/ff·전체 `make gate`·archive·push(로컬 docs 커밋 `8d20453c8`·`01e61a756` 포함)·정리·retain)을 "진행"으로 승인(2026-10-26, a174 한정). 운영 이미지 재빌드·배포는 범위 밖.
  - rebase: 불필요 — 브랜치가 main `01e61a756` 에서 분기했고 main 이 그 뒤로 움직이지 않음(ff 가능). Manager 편집 커밋 `11be30b70`.
  - gate ⑤ 사전 실행: `check_analysis.py --change a174-…` rc 0, required 5(`Console.handleDashboard` · `TestEveryCardShowsWhatChangesAndWhen` · `TestTheCurrencyConsequenceIsRenderedAndNotFolded` · `TestThePresetControlsAskForNoTyping` · `limitCurrent`) — `analysis/function-logic/` 번들 6(Terra 작성, `TestGuardianPresetsRenderAsOneMatrixPerCurrency` 포함) → `evidence complete or diff-proven exempt`.
  - 완료 게이트 결과는 review.md 「완료 게이트 실행 영수증」에 적음(관례: 체크 → tracker → gate).

## 고정 리터럴 시험 수정 기록

- `internal/console/settings_limits_test.go:TestThePresetControlsAskForNoTyping` — 프리셋 form 안의 가짜 `card-preview`·`data-limit-axis` 문자열 검사를 실제 대응 표 열의 다섯 한도 값·방향 및 표 아래 전환 문장 검사로 교체함. 단언 수를 줄이지 않고 가짜 속성 부재 단언을 추가함.
- `internal/console/settings_card_test.go:TestEveryCardShowsWhatChangesAndWhen` — 한도 카드는 form 안 preview가 아니라 비교표 열에서 적용 전후를 읽는 구조이므로 `preset-matrix` 존재를 단언하도록 정정함. 다른 카드의 preview 단언은 유지함.
- `internal/console/readiness_test.go:TestGuardianPresetsRenderAsOneMatrixPerCurrency` — `data-currency-consequence`는 현재 계좌 통화 notice 한 번만 가져야 하므로, 프리셋별 결과를 세던 단언을 현재 notice 1회 단언으로 정정함.
