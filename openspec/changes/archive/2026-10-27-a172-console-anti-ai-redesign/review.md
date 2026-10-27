# a172 리뷰 기록

change: `a172-console-anti-ai-redesign` · Story `STORY-TOS-a172` · 위험 Low(콘솔 CSS · 장식 마크업 · 시험).
이 파일은 종결 시점(2026-10-28, 종결 에이전트 Opus 5.5)에 썼다. 구현 중에는 review.md 가 없었고(gate ④ 실패 원인),
리뷰 결과는 tasks.md 와 세션 스크래치패드의 리뷰 원문에만 있었다. 아래는 그 기록을 모은 것이다.

## proposal-freeze

- 등급: UI · 문서 · 시험 change 라 경량 리뷰(docs/WORKFLOW.md 「리뷰 게이트」 — validate + Manager 셀프리뷰 + 기록).
  주문 · 위험관리 · 원장 · 대사 경로 미접촉이라 적대적 Eng 보이스 의무 없음.
- proposal-freeze 에 대한 별도 gstack(autoplan) 리뷰 원문은 **없음** — 저장소 · 스크래치패드 어디에서도 찾지 못함.
  proposal 의 범위 · 비목표 · 위험 · `not-applicable` 사유가 freeze 시점의 판단 기록이다.
- 6단계 범위 정정(사람 결정 2026-10-25): JS asset · `/static/` 라우트 · `script-src 'self'` 안(`bbe176f5e`)을
  operator-console 스펙의 SHALL NOT(클라이언트 스크립트 · CSP 완화 금지)을 근거로 되돌리고 CSS-only 로 확정(`376b8a706`).
  requirement 수준 변경이 아니라 범위 축소다.
- spec 델타(`794f9e95a`, 테마 토큰 · 모션 계약 요구사항)는 strict validate 통과용으로 구현 뒤 추가됨.

## 구현 리뷰 (단계별 Impeccable 리뷰 — 사용자 지시 "단계마다 Impeccable 리뷰")

원문은 세션 스크래치패드(`/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/89bda6e1-…/scratchpad/a172/`)에 남아 있었고,
종결 때 찾아 `reviews/` 로 복사해 커밋했다(원본은 그대로 둠).

| 리뷰 | 원문 | 요약 | 처리(tasks.md 표) |
|---|---|---|---|
| 1단계 징후 측정 | `reviews/stage1-scan.md` | AI 티 징후 측정 | 1.2 제거 |
| 3단계 Impeccable | `reviews/review-stage3.md` | 12건(side-tab 테두리 · 모바일 한 열 · 44px · 버튼 상태 · 다크 경계 등) | 반영 10 · 부분 2(7 스크롤 표 · 12 위계) — 사유 tasks 「Impeccable 3단계 리뷰 반영」 |
| 4단계 Impeccable | `reviews/review-stage4.md` | R1~R17 | 반영 9(R4 는 부분) · 거절 1(R3 — 규칙 20 스프링 지시 우선) · 보류 2(R12 · R13) · 해당 없음 1(R5) · do-not-fix 유지 4(R14~R17) — tasks 「Impeccable 4단계 리뷰 반영」 |
| 6단계 Impeccable | `reviews/review-phase6.md` | critique 25/32 · audit 15/20 · avoid-ai-design 신규 0 · 대비 최저 5.10:1 | 반영 7 · 거절 1(키프레임 `ease-in-out` — `TestMotionContract` 금지) · do-not-fix 유지 — tasks 「6단계 Impeccable 리뷰 반영」 |

- 원문 한계: `review-stage3.md` 는 스스로 "DEGRADED: single-context"(두 보이스를 한 문맥에서 순차)라고 적는다.
  `review-phase6.md` 가 가리키는 `r6-avoid.json` 등 측정 산출물은 옮기지 않았다(스크래치패드에만 있음).
- 5.1 경량 리뷰 · 포트(2026-10-25): 전체 콘솔 시험(비캐시 console 48.3s · cmd/tossctl 63.3s), 검출기 `[]`, main ff `79d32f9af`.
- 사람 결정 2건(5.1): POST 폼 버튼 전환 전부 제외 유지, 스프링 유지 · 모달 오버슈트 12%. 6.5b 라이트 테마 톤 구분은
  리뷰어의 "강도 올리지 말 것"을 사용자가 번복(2026-10-25).
- 변이 확인(tasks 기록): `TestNoSideTabAccentBorders`(border-left 재삽입 → FAIL), `TestMotionContract`(변이 FAIL),
  `TestAmbientMotionContract`(2회 CAUGHT), 렌더 기반 단언 (g)(`.tier-card` 재진입 → FAIL), `TestAmbientToneLegibility`(라이트 alpha 10% → FAIL).

## gate ⑤ — Function Logic Map (종결 때 사후 작성)

- proposal 의 `not-applicable` 사유("Go 함수 내부 로직을 편집하지 않는다")는 **사실과 달랐다**: a172 는 시험 함수
  본문을 고쳤다(`display_primitives_test.go` 의 `TestTheHeadingStepsAreDistinguishable` · `declaredFontSize` ·
  `declaredBodyFontSize`, `trading_views_test.go` 의 `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA` 삭제 · 대체).
  gate ⑤ 는 시험 파일도 세고, 면제 표지는 required 0 일 때만 받는다. 제품 Go 코드(`templates.go`)는 CSS 문자열만 바뀌어 함수 본문 변경 0.
- 사전 실행(2026-10-28, HEAD `eaf713ff9`): `check_analysis.py --change a172-console-anti-ai-redesign` → base `fd6afe23b` →
  워킹트리 창, required 26 · 번들 0, rc 1. 창에 base 뒤 착지한 커밋 79 개가 들어 있어 a173 · a174 · a175 · a176(a178) 이
  고친 함수까지 요구됨. `--record-landing` 은 `revision: current` 번들이 없어 좁히지 못함. base 재고정은 사용자가 기각(2번 안).
- 처분: **복사 18 · 새로 작성 8**. 재실행 → `evidence complete or diff-proven exempt`, rc 0.
  - 복사 18(현재 소스 sha256 = 가장 나중 아카이브 번들의 `source_sha256`, 출처 주석 한 줄 추가). 복사 뒤 18 개 전부
    `go run ./tools/logic-map` 로 현재 소스에서 다시 뽑아 ast.json 이 바이트 단위로 같음을 확인함(`revision` 키 제외).
    - a173(8): `Console.chromeFor` · `holdingsSnapshot.freshness` · `ordersSnapshot.freshness` · `signalsView.freshness` ·
      `Console.handleOptimization` · `optimizationCategoryViews` · `Console.handlePerformanceHistory` · `fakePerformanceReader.Dashboard`
      (`handleOptimization` · `handlePerformanceHistory` · `Dashboard` 는 a177 번들도 같은 sha 지만 a173 이 나중에 아카이브됨)
    - a175(4): `TestCandidateFilterCardsContainExactMarketMetricMatrix` · `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` ·
      `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions` · `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`
    - a174(3): `Console.handleDashboard` · `TestEveryCardShowsWhatChangesAndWhen` · `TestThePresetControlsAskForNoTyping`
    - a176(3, a178 보다 나중 아카이브): `verifyPrescribedBinary` · `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` · `fakeBuildRunner.rebuild`
  - 새로 작성 8:
    - a172 자기 편집 4: `TestTheHeadingStepsAreDistinguishable` · `declaredFontSize` · `declaredBodyFontSize`(current),
      `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`(삭제 → `revision: base`, 대체 시험 `TestTradingViewsThemeColorsMeetWCAGInBothThemes`).
    - 본문 무변경인데 0줄 hunk 가 함수 끝에 닿아 요구된 3(`revision: base`, ast 는 `git show fd6afe23b:<file>` 에서 뽑음):
      `TestTheSummaryCostsNoBrokerCall`(a172 가 바로 뒤에 새 시험을 붙임, `@@ -371,0 …`) ·
      `TestTheCurrencyConsequenceIsRenderedAndNotFolded` · `limitCurrent`(a174 `aa2864abe` 가 바로 뒤에 붙임).
    - 파일 sha 가 아카이브 번들과 달라진 1: `Console.handleReport` — a173 번들 sha `f219fc44…` vs 현재 `d676ff00…`
      (a174 가 같은 파일의 `handleDashboard` 를 고침). 함수 본문은 a173 판과 같음(AST 상대 좌표 비교 일치) → 분석은 옮기고 ast 는 현재로 다시 뽑음.
- 분기 · 시험 대응은 실제 시험 이름으로 적었고 2026-10-28 `go test ./internal/console -count=1 -run '^(…9개…)$' -v` 전부 PASS 로 확인함.
  `handleReport` 는 `-coverprofile` 로 쟀다: 콘솔 패키지 전체 실행에서 B2 블록 count>0, **B1(`c.report` 오류 갈래) count 0** —
  시험 없는 기존 공백(a173 측정과 같음, a172 무관)이라 지어내지 않고 「측정상 없음」으로 적음.
  시험 함수 번들의 실패 갈래(`t.Fatal`/`t.Errorf`)는 통과하는 동안 설계상 실행되지 않음 — 「미실행(설계상)」으로 적음. RED 관측은 사후 작성이라 전부 `no`.
- 발견(a174 아카이브 증거, a172 를 막지 않음): a174 의 `internal-console--limitcurrent/ast.json` 은 `revision: base`, sha `5524583f…` 인데
  좌표가 598-609(분기 601 · 605)다. 그 sha 판의 실제 좌표는 597-608(분기 600 · 604)이다. sha 가 같아도 이 번들은 복사하지 않고 base 판에서 다시 뽑았다.
  gate 는 md 산문 좌표와 ast.json 만 대조하고 ast.json 좌표와 소스는 대조하지 않아 통과했던 것으로 보인다.
- FLM 은 구현 판단에 쓰이지 않았다(사후 작성). 고위험 경로 영향 없음(콘솔 CSS · 시험 · 읽기 전용 렌더).

## 6.6 main 병합

- a172 커밋은 이미 main 의 조상이다: `git merge-base --is-ancestor b27a1ea87 eaf713ff9` 참. 5.1 포트(main ff `79d32f9af`, 2026-10-25),
  6단계(`156340d04` · `4f3b94e28` · `b27a1ea87`) 와 spec 델타(`794f9e95a`) 모두 main 역사에 있다. 2026-10-25 배포 · 운영 중(Manager 실측).

## 완료 게이트 실행 영수증

- 격리 worktree `/mnt/D/Axipient/workspace/TossOS.worktrees/a172-close`(`/tmp` 밖), detached `7cbbbc378`(종결 증거 커밋), 2026-10-28 KST.
- 준비: `make sdd-infra` rc 0 → `python3 tools/sdd/sdd_sync.py --no-gbrain` rc 1 — CodeGraph 색인은 됨(3,021 파일, 13.7s),
  **advisory CodeGraphContext 색인만 300초 시한 초과**(`codegraphcontext index .` timed out) → `make sdd-check` rc 0 →
  `codegraph status .` 37.0s · 3.29s · 3.00s · 2.71s · 2.75s(연속 2회 15초 미만 충족). `/tmp/.git` 은 있었지만 worktree 가 `/tmp` 밖이라 무관.
- 조용함: 시작 직전 1분 load 3.47, 다른 gate · go test 없음.
- `rtk proxy make -C <격리> gate CHANGE=a172-console-anti-ai-redesign` 1 회차(06:39~07:31) → **GATE PASS 11/11, rc 0**, `repo:` 줄 = 격리 worktree.
  ①~④ OK · ⑤ `OK: Function Logic Map` · ⑥ sdd-check · ⑦ test · ⑧ test-seams · ⑨ test-race · ⑩ vet · ⑪ validate 전부 OK. 재시도 없음
  (`deadline exceeded` · VCS status 128 0건). 원문 로그는 `reviews/gate-run1.log`(세션 스크래치패드에서 복사).
- archive: `openspec archive a172-console-anti-ai-redesign --yes` → `operator-console: + 1 added`, `2026-10-27-a172-console-anti-ai-redesign`(openspec 은 UTC 날짜).
