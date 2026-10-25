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
