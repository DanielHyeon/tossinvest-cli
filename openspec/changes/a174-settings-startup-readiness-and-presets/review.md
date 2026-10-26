# a174 review

## proposal-freeze (2026-10-26, 경량: validate + Manager 셀프리뷰, docs/WORKFLOW.md 「위험 등급 가중」 UI change)

- 보이스: Manager(Opus 5.5) 단독. 적대적 Eng 보이스는 해당 없음. 주문·위험관리·원장·reconciliation 경로를 건드리지 않고, 기동 인터록 판정(`GatePreflight`)은 호출만 하고 편집하지 않는다.
- base: `28b438293`(a173 archive 커밋, origin/main 과 같음).
- 입력: Fable 5.1 의 정의(proposal·design·tasks.draft·spec 델타·HANDOFF), HEAD `4e20c5df5`/`28b438293` 의 codegraph 1.6.0 원문과 grep, `analysis/evidence/GatePreflight.ast.json`.
- 사람 결정(2026-10-26, HANDOFF §3): worktree 는 형제 디렉터리, WORKFLOW 문구는 게이트에 맞춘다, a177 교훈 durable page 작성(a173 종결자가 함), 구현자는 GPT-5.6 Terra.
- 발견과 결정(정의본 대비 정정, 전부 design·proposal·tasks 에 반영):
  1. **검증 콘솔 데이터 경로가 성립하지 않았다.** 정의본은 템플릿 함수 `readinessForConsole(c *Console)` 을 제안했다. 하지만 `pageFuncs`(`pages.go:38`)는 패키지 전역이라 `*Console` 이 없고, `dashboardPage`(`pages.go:80-91`)에는 Gate·Trading·Autostart 가 없다. → `dashboardPage.Readiness` 필드 + 새 메서드 `Console.startupReadiness` + `handleDashboard` 한 줄로 바꿨다. 그래서 FLM 면제가 아니라 번들 1개(편집 전 작성)가 필요하다(design §2.1, proposal not-applicable, tasks 3.1).
  2. **행 모델이 재계산 금지와 충돌했다.** 정의본은 "한도 5행·거래 정책 4행"이었다. AST 상 `GatePreflight` 는 한도·정책을 절 하나씩으로 낸다(B1/B2, B3~B5). 다섯 행으로 쪼개려면 필드별 판정을 새로 해야 하고, 그것이 spec 델타가 금지한 재계산이다. → 절 단위 여섯 행, 합산 판정 0(tasks 2.1).
  3. **이름 금지 시험.** `engineproc_test.go:194` 는 `settings.go`·`settings_limits.go`·`templates_settings.go` 밖에서 `AutomationGate`·`automation_gate` 를, 모든 파일에서 `Interlock`·`ProtectionReady` 를 금지한다. → `readiness.go`·`templates.go` 는 그 철자를 쓰지 않고, `readiness` define 은 `templates_settings.go` 에 둔다(design §1·§2.1).
  4. **프리셋 사실 정정.** 한도는 일곱이 아니라 다섯 + 통화이고(`previewAgainst` 의 `Rows`), 폼 hidden 필드 이름은 `tier` 다. 값·방향은 `LimitTierCards()` 에서 옮긴다(design §2.2).
  5. **파일 겹침.** CSS 는 전부 `templates.go` `style` 한 곳에 있다. a174·a175·a176 이 모두 고친다 → `style` 끝에 change 마다 블록 하나만 덧붙인다. a176 은 a174 착지 뒤에 착수한다(design §2.5, a175·a176 design 머리말).
  6. 0값 유출 방지: seam 미배선·읽기 실패인 줄은 `unknown` + 사유로 그린다(a177 교훈, tasks 2.1).
- `openspec validate a174-settings-startup-readiness-and-presets --strict --no-interactive`: `Change 'a174-settings-startup-readiness-and-presets' is valid`(2026-10-26).
- 남은 위험: 검증 콘솔은 2초마다 다시 읽으므로 설정 파일 파싱이 3회 늘어난다(브로커 호출 0). `Validate()` 내부(config 패키지)는 읽지 않았다 — 그 오류 문장을 그대로 옮길 뿐이다.
- 결론: proposal-freeze 승인. 구현 착수 가능. 위치 `/mnt/D/Axipient/workspace/TossOS.worktrees/a174-settings-startup-readiness-and-presets`, 구현자 GPT-5.6 Terra. 단 정정본이 **커밋된 뒤에** worktree 를 만든다(미커밋 편집은 worktree 에 들어가지 않는다).

## Pre-Edit Gate — a174 3.1

- 대상 심볼: `console.Console.handleDashboard`
- CodeGraph: 정의·호출자·callee·impact는 `analysis/code-context/codegraph-baseline.md`에 기록함. 호출자는 `/verify-console` 하나이며 영향은 handler와 route로 한정됨.
- CodeGraphContext: 새 worktree 미등록으로 미가용. 현재 HEAD·CodeGraph·현재 소스·기존 시험으로 조정한 근거는 `analysis/code-context/` 세 파일에 기록함.
- 기존 동작: dashboard 리터럴은 chrome·CSRF·notice·snapshot·재시작 seam을 조립하고, B1은 현재 run이 있을 때만 `Run`·`Refresh`를 채운 뒤 한 번 렌더함.
- Function Logic Map / Branch Test Map: `analysis/function-logic/internal-console--console.handledashboard/`.
- upstream 상속 시험 영향: 있음. 설정·검증 콘솔 렌더와 폼 리터럴을 새 RED 2.1~2.4 및 기존 console 패키지 회귀로 확인함.
- 실패 시험 선행: RED 2.1~2.4를 production 편집 전에 작성·실행함.
- 설정·DB·journal 변경과 rollback: 없음. 읽기 전용 뷰 모델과 템플릿 배치만 추가하며, 되돌림은 이 change의 코드·템플릿 커밋을 되돌리는 것으로 충분함.
- 안전 불변식: 통과. live 주문·운영 토글·엔진 제어 handler와 POST 계약을 변경하지 않음.

## 구현 리뷰 (Terra, 2026-10-26, 경량)

- 연구: `search-first` 빠른 점검에서 현 저장소의 `GatePreflight`·`LimitTierCards()`·`config.GuardianTiers()`와 기존 console 계약 시험을 확인함. 한도 값·프리셋·POST 계약은 새 의존성이나 외부 구현 없이 이 정본을 재사용하는 것으로 결정함. 외부 문서·레지스트리 검색은 로컬 정본이 충분해 수행하지 않았고, 새 의존성·라이선스 영향은 없음. 호환성은 기존 form 계약 시험과 console 회귀 시험으로 확인함.
- 적용 범위: UI 변경이므로 `impeccable`·접근성·모션 지침을 적용함. `gstack review` 경량 절차로 독립 specialist 네 개를 실행했고, 전용 design-critic은 미설치여서 `analysis/review-after.md`의 수동 검토로 대체함.
- 리뷰 기준: Manager가 지정한 구현 base `01e61a756`을 사용함. `origin/main` diff에는 선행 a173/Manager 문서가 섞여 있어 a174 구현 범위를 판정하는 기준으로 쓰지 않음.
- 범위 검사: `gstack-diff-scope 01e61a756`은 backend 범위로 분류함(Go 템플릿 포함). 변경 대상 `impeccable detect` 결과는 `[]`임.
- specialist 결과: security·performance는 `NO FINDINGS`. testing은 각 seam 단독 실패가 다른 unknown 행에 가려질 수 있음을 정보성으로 지적했고, 여섯 단독 nil/error 케이스를 상시·당일·검증 콘솔에 각각 단언하는 새 시험으로 보강함. maintainability는 `GatePreflight` 절 이름의 문자열 매핑 드리프트 위험을 정보성으로 지적함.
- maintainability 처분: 미해결·정보성. 안정 ID 추가는 `preflightClause`/`GatePreflight` 계약 변경을 요구하고, 이 change는 본문·판정 구조 편집 금지이므로 수행하지 않음. 절 이름이 바뀌면 해당 행은 통과가 아니라 `unknown`과 사유를 표시함.
- 변이: `StartupReadiness`에서 거래 정책 행과 그 지역 변수를 함께 제거했을 때 `TestStartupReadinessAppearsOnThreeScreensWithoutControls`가 행 수 5, 거래 정책 행 부재로 FAIL함. 원복 뒤 `거래 정책 행`·`tradingKnown` 심볼 수가 각각 1이고 대상 시험 PASS임.
- 미실행: Claude/Codex 적대적 전체 리뷰와 구조화 리뷰는 Manager가 요청한 "경량" 범위를 넘고 종결자 검증과 중복되어 실행하지 않음. gstack 업그레이드는 사용자 범위 밖 전역 변경이므로 하지 않음.
- 결론: 차단·고위험 발견 없음. 위 문자열 매핑 정보성 위험을 Manager handoff에 남김.

## Manager 리뷰 보정 1 (Terra, 2026-10-26)

- P1-1 처분: `LimitCurrencyConsequence`를 한도 카드의 `data-currency-consequence="current"` notice 한 번으로 복원함. 표별 결과는 "이 통화의 프리셋을 고르면"이라는 가정 문맥의 muted 문장으로 남기고, 현재 통화 속성을 갖지 않음. USD·미설정 fixture에서 각 현재 문장과 KRW 오독 부재를 RED→GREEN으로 확인함.
- P1-2 처분: preset form에서 `card-preview`, `data-limit-axis="비교표"`, 가짜 `aria-label`을 제거함. 고정 리터럴 시험은 동일 form의 대응 표 열에서 `LimitTierCards()` preview 다섯 값·실제 방향을 확인하고, FirstTime·CurrencyChange 문장이 표 아래에 있는지 확인하도록 보강함. 수정 함수와 이유는 `tasks.md` 고정 리터럴 시험 수정 기록에 전수 기재함.
- P3 처분: 정상 fixture의 네 `GatePreflight` 절 모두가 `사전 판정 항목을 찾을 수 없다`를 렌더하지 않는 회귀 단언을 추가함. 안정 ID 도입은 기존 `GatePreflight` 계약 변경 금지 범위를 넘으므로 하지 않음.
- 분석 번들: 기존 함수 편집·삽입 경계 기준 required 5개(`Console.handleDashboard`, `TestEveryCardShowsWhatChangesAndWhen`, `TestTheCurrencyConsequenceIsRenderedAndNotFolded`, `TestThePresetControlsAskForNoTyping`, `limitCurrent`) 모두 `analysis/function-logic/`에 AST·FLM·BTM·risk report TODO 0으로 작성함. 두 인접 기존 함수는 본문 의미를 바꾸지 않아 base revision AST로 기록함.
- UI·접근성: `impeccable` operate surface 기준으로 표 머리·scope, 버튼의 기존 이름과 즉시성을 유지함. 숨은 form 속성에 의존하지 않고 읽기 순서에서 현재 통화 notice→비교표 caption→표 아래 전환 문장을 제공함. 첫 렌더 2544px가 목표를 44px 넘어서 a174 CSS 블록에서 프리셋 셀의 세로 여백만 제거했고, 최종 격리 렌더는 2499px·상자 3으로 통과함. `impeccable detect --json internal/console/templates.go internal/console/templates_settings.go`는 `[]`임.
- 연구: 현 저장소의 `LimitTierCards()`·`LimitCurrencyConsequence()`·`config.GuardianTiers()`와 기존 form 계약을 다시 검색해 재사용함. 외부 구현·새 의존성은 불필요하며 라이선스·호환성 영향 없음.
- P3 기록(Manager 4.5, 2026-10-26): `/verify-console`의 체크리스트 `<h2>`가 「엔진 런타임」 `<h2>` 절 안에 중첩됨. define을 매개변수로 나누는 것은 현 범위에서 YAGNI라 보류함.
- P3 기록(Manager 4.5, 2026-10-26): 한도 줄 사유는 `Validate()`의 영문 원문을 그대로 옮김. 재계산 금지 규칙 때문에 번역·재구성하지 않음.
- P2 처분(Manager 4.5, 2026-10-26): 모바일 `details.explain > summary`의 펼침 표식은 a174 모바일 CSS에서 `display: list-item`으로 복원하고, a173 복원 계약과 같은 순서 단언을 새 독립 시험으로 RED→GREEN 확인함.
- P2 처분(Manager 4.5, 2026-10-26): 비교표 셀은 기본 줄 간격과 `0.3rem` 세로 여백을 복원함. 열 머리 버튼은 줄바꿈과 최대 13rem 폭을 허용해 390 다크 표 폭을 640px·858px로 줄였음. 1280 높이는 2786px가 되어 2500 목표를 넘지만, 정보 여백을 다시 줄이지 않고 측정값을 기록함.
