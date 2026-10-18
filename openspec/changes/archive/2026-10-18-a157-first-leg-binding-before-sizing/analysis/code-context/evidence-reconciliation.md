# a157 evidence reconciliation

기준: base `7b1748481`, 질의 HEAD `42cc0dff8`(internal/ 무변경). CodeGraph 1.6.0 · CodeGraphContext(advisory, 색인 범위 불일치) · HEAD grep(`raw/grep-*.txt`) · AST 8(`analysis/ast-evidence/`).

| 항목 | CodeGraph | HEAD 읽기 · AST · grep | 결론 |
|---|---|---|---|
| `PrecheckQFinalCampaignFirstLeg` 생산 호출자 | 인터페이스 **선언** 줄 `strategy_first_leg_admission.go:65` 을 호출자로 표시 | 실제 호출은 `:105 precheck, err := b.guardian.PrecheckQFinalCampaignFirstLeg(authority)` 하나(`raw/grep-precheck-first-leg-nontest.txt`) | **불일치(표기)** — CodeGraph 가 인터페이스 메서드 선언을 호출 자리로 적음. HEAD 가 권위: 생산 호출자 1(`admit` :105) |
| `PrecheckQFinalEntry` 호출자 | 첫 다리 :49 · `IssueQFinalEntry` :322 | grep `:50` · `:323`(`raw/grep-precheck-entry-nontest.txt`) | 일치(줄은 함수 머리 대 호출 줄 차이). 단독 발급 경로가 있으므로 판정을 `PrecheckQFinalEntry` 안에 넣지 않는다(design D1 기각 대안) |
| `admit` 생산 호출자 | 이름 충돌 — `strategyworker.RunBounded` 의 `admit` 등과 섞이고 `dispatch` 누락 | `strategy_dispatch_cycle.go:151 cycle.firstLeg.admit(ctx, result)` 하나(`raw/grep-bridge-admit-nontest.txt`) | **불일치** — 필드 경유 메서드 호출을 놓침(a146 과 같은 결함). HEAD 가 권위 |
| 「일일 손실 체인 거절」 판별 자리 | — | `admit` :110-123 두 조건이 유일한 분류 자리. `strategy_market_handoff_delivery.go:46 · :86` 은 `DailyLossUnmeasurable` 표지만 보는 **기록 · 주기 종료** 자리(분류 아님). 표지는 `dailyLossChainRefusal`(a137_daily_loss.go:203-209)이 `Step daily_loss` + `ReasonInputUnavailable` 일 때만 체인 거절에 붙이므로 표지가 있으면 체인 단계 · 입력 불가도 있다 | 판별 함수는 `admit` 의 두 조건을 옮긴다. 전달 루프의 표지 판정은 범위 밖 — a157 뒤 결속 결함이 이긴 요청은 표지 없는 거절을 내므로 전달 루프는 멈추는 결함으로 받음(design D2 (a)(b)(e)) |
| 조립 감독자의 옵션 | `NewStrategyEntrySupervisor` 호출자 3(조립 둘 + 생산 감독자) | AST `engine--context.newpairedstrategyentryproductionassembly` calls(394) — 옵션 `Clock · CycleLimit · Workers` 뿐. 생산 감독자 :433-434 는 `EntryGate · DailyLossNotice · EntrySkipNotice` | 일치 — design F9 |
| `c.Entry` 타입과 nil | — | `engine.go:181 Entry *execgw.EntryGate`(`raw/grep-context-entry-type.txt`) · `blockEntryOnCentralIntegrity` :1046 `s.entry == nil` · `(*EntryGate).Block` :543-545 수신자 가드 없음(`raw/grep-*.txt`) | 타입 있는 nil 이 인터페이스에 들어가면 panic — `if c.Entry != nil` 가드 필요(design F10 · D4) |
| 알림 함수의 nil | `strategyDailyLossNotice` · `strategyEntrySkipNotice` 호출자 = 생산 감독자 하나 | a137_daily_loss_notice.go:58 · a146_entry_skip.go:114 — `c.Log == nil` 이면 nil 반환 | 조립 감독자에 넘겨도 로그 없는 조립에서 안전 |
| 영향 시험 | 기본 판별식 1 파일(Python) | `--filter '*_test.go'` 1,153 파일 | 기본값 불신 — 대상 패키지 전체를 돌린다(tasks 2.1) |

편집을 막는 미해소 불일치: **없음**(두 불일치 모두 CodeGraph 표기 결함이고 HEAD grep 으로 해소).
