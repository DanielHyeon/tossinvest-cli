#!/usr/bin/env python3
"""a137 변이 원장 하네스 — 고정한 연결 워크트리(사본) 하나에서 변이를 하나씩 넣고 대상 시험을 돌린 뒤 되돌림.

사용: python3 a137_mutate.py <worktree> <expected-sha> <out.md>
- 시작 sha 를 단언하고, 사본이 깨끗한지(git status) 확인한 뒤, 무변이 대조군이 GREEN 이어야 시작함(기억: 대조군이 GREEN 이 아니면 멈출 것).
- 변이마다: 정확히 한 번 나타나는 문자열을 바꿈(아니면 NOT-REACHED) → 시험 → git checkout 으로 그 파일만 되돌림 → 사본이 다시 깨끗한지 확인.
"""
import os, subprocess, sys

WT, SHA, OUT = sys.argv[1], sys.argv[2], sys.argv[3]
ENV = dict(os.environ, GOFLAGS="-trimpath")

def run(args, **kw):
    return subprocess.run(args, cwd=WT, capture_output=True, text=True, env=ENV, **kw)

E, J, D, G = "internal/execgw/", "internal/journal/", "internal/dailyloss/", "internal/app/engine/"
MUTANTS = [
 ("M1 max → 호출자만", E+"a137_daily_loss.go", "\t\taccount.DailyRealizedLoss = riskcalc.Money{Amount: measured.Amount, Currency: measured.Currency}\n", "", E, "TestA137TheLedgerLossRefusesAZeroCallerAndEscalates", ""),
 ("M2 측정 실패 → 호출자 값 낙하", E+"a137_daily_loss.go", "\t\taccount.DailyRealizedLoss = riskcalc.Money{}\n\t\treturn account, err\n", "\t\treturn account, nil\n", E, "TestA137AMeasurementFailureNeverFallsBackToTheCaller", ""),
 ("M3 무효 호출자 값 대체", E+"a137_daily_loss.go", "\t\treturn account, nil // 무효한", "\t\taccount.DailyRealizedLoss = measured\n\t\treturn account, nil // 무효한", E, "TestA137AnInvalidCallerValueIsNotReplaced", ""),
 ("M4 이익 상쇄", D+"measure.go", "\t\tif !negative {\n\t\t\tcontinue // Q12", "\t\tif false && !negative {\n\t\t\tcontinue // Q12", D, "TestA137ProfitsDoNotOffsetLosses", ""),
 ("M5 시장 현지일 귀속", D+"measure.go", "day, err := clock.AccountTradingDay(closedAt)", "day, err := clock.MarketKR.TradingDay(closedAt)", D, "TestA137AttributionIsTheOneAccountDay", ""),
 ("M6 조정 종결 무조건 제외", D+"measure.go", "\t\t\tif row.EngineSellDispatched {", "\t\t\tif false && row.EngineSellDispatched {", D, "TestA137AnAdjustmentCloseWithADispatchedEngineSellIsUnmeasurable", ""),
 ("M7 성과 부재 무시", D+"measure.go", "\t\tif row.Outcome == nil {\n\t\t\treturn Losses{}, unmeasurable(row, \"closed today without a frozen outcome row\")", "\t\tif row.Outcome == nil {\n\t\t\tcontinue\n\t\t\treturn Losses{}, unmeasurable(row, \"closed today without a frozen outcome row\")", D, "TestA137AManagedCloseWithoutAnOutcomeIsUnmeasurable", ""),
 ("M8 토큰 신선도 무시", E+"a137_daily_loss.go", "if age := now.Sub(m.at); age < 0 || age > riskcalc.AccountSnapshotStaleness {", "if age := now.Sub(m.at); false && (age < 0 || age > riskcalc.AccountSnapshotStaleness) {", E, "TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne", ""),
 ("M9 일 키 확인 삭제", E+"a137_daily_loss.go", "if err != nil || day != m.dayKey {", "if err != nil {", E, "TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne", ""),
 ("M10 다른 Guardian 토큰 허용", E+"a137_daily_loss.go", "if m.guardian == nil || m.guardian != g {", "if m.guardian == nil {", E, "TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne", ""),
 ("M11 위험 축소가 측정", E+"riskguardian.go", "func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n", "func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n\t_ = g.MeasureDailyLoss(ctx)\n", E, "TestA137RiskReductionNeverMeasures", ""),
 ("M12 IssueEntry 바닥 미적용", E+"riskguardian.go", "account, measureErr := g.dailyLossFloor(req.Account, measurement, now, nil)", "account, measureErr := req.Account, error(nil)\n\t_ = measurement", E, "TestA137TheLedgerLossRefusesAZeroCallerAndEscalates", ""),
 ("M13 측정 실패 상세 미교체", E+"a137_daily_loss.go", "if !verdict.Allowed && measureErr != nil && verdict.Step == \"daily_loss\"", "if false && !verdict.Allowed && measureErr != nil && verdict.Step == \"daily_loss\"", E, "TestA137AMeasurementFailureNeverFallsBackToTheCaller", ""),
 ("M14 강화 생략", G+"strategy_first_leg_admission.go", "refusal.cause = b.guardian.EscalateDailyLossLimit(ctx, err)", "refusal.cause = &execgw.DailyLossEntryBlock{Cause: err}", G, "TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault", "tossos_testseams"),
 ("M15 Stage 검사 삭제", G+"strategy_first_leg_admission.go", "errors.As(err, &issue) && issue.Stage == execgw.StageChain &&\n\t\t\tissue.Reason == string(risk.ReasonDailyLossLimitReached)", "errors.As(err, &issue) &&\n\t\t\tissue.Reason == string(risk.ReasonDailyLossLimitReached)", G, "TestA137ClassificationIsUnchangedForEveryPrecheckRefusal", "tossos_testseams"),
 ("M16 precheck 원인 버림", G+"strategy_first_leg_admission.go", "\t\t// a137 F8: 원인을 실음 — 하류가 errors.As 로 거절 타입을 볼 수 있게(문구 · 분류는 그대로).\n\t\trefusal.cause = err\n", "", G, "TestA137APrecheckRefusalKeepsItsCause", "tossos_testseams"),
 ("M17 차단을 오류로", G+"strategy_market_handoff_delivery.go", "if errors.As(err, &blocked) || errors.As(err, &unmeasurable) {", "if errors.As(err, &unmeasurable) {\n\t\t\t_ = blocked", G, "TestA137ADailyLossModeBlockEndsTheCycleQuietly", "tossos_testseams"),
 ("M18 건너뛴 범위 보고 삭제", G+"strategy_market_handoff_delivery.go", "\t\t\treturn errors.Join(skipped...)\n\t\t}\n\t\t// 멈추는", "\t\t\treturn nil\n\t\t}\n\t\t// 멈추는", G, "TestA137AMixedCycleStillReportsItsSkippedScopes", "tossos_testseams"),
 ("M19 다른 trigger 도 조용히", E+"modegate.go", "rec.Mode == journal.ModeEntryBlocked && rec.Actor == journal.ModeActorAuto && rec.Cause == journal.ModeTriggerDailyLossLimit", "rec.Mode == journal.ModeEntryBlocked", E, "TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock", ""),
 ("M20 다른 latch 와 함께여도 조용히", E+"a137_daily_loss.go", "if len(g.latches) != 1 || g.modeDailyLossDetail == \"\" {", "if g.modeDailyLossDetail == \"\" {", E, "TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock", ""),
 ("M21 펜스 원인 버림", E+"a137_daily_loss.go", "\tif errors.As(err, &block) {\n\t\trejected.cause = block", "\tif false && errors.As(err, &block) {\n\t\trejected.cause = block", E, "TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange", ""),
 ("M22 checkEntry 표지 삭제", E+"gateway.go", "\t\treturn g.entry.markStrategyRefusal(rejected, plan.market, plan.symbol)", "\t\treturn rejected", E, "TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock", "tossos_testseams"),
 ("M23 관측 표지 삭제", E+"strategy_entry_gate_authority.go", "return StrategyEntryGateAuthority{}, g.entry.markStrategyRefusal(refusal, market, symbol)", "return StrategyEntryGateAuthority{}, refusal", E, "TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock", ""),
 ("M24 원장 보호 주문 무시", J+"a137_daily_loss_rows.go", "\tfor _, span := range protections {\n\t\tif overlaps(span[0], span[1]) {", "\tfor _, span := range protections {\n\t\tif false && overlaps(span[0], span[1]) {", J, "TestA137AnOverlappingProtectionOrderIsSeen", ""),
 ("M25 원장 형식 결함 무시", J+"a137_daily_loss_rows.go", "\tif malformed > 0 {", "\tif false && malformed > 0 {", J, "TestA137AMalformedClosedAtMakesTheReadFail", ""),
 ("M26 원장 NOT_DISPATCHED 포함", J+"a137_daily_loss_rows.go", "AND ma.state NOT IN (?, ?)", "AND ma.state NOT IN (?, ?, 'x') OR 1=1", J, "TestA137AnOverlappingDispatchedEngineSellIsSeen", ""),
 ("M27 원장 조정 종결 decimal 0 → 문자 비교", J+"a137_daily_loss_rows.go", "r.AdjustmentClosed = lastAdjQ != nil && isZeroDecimal(*lastAdjQ)", "r.AdjustmentClosed = lastAdjQ != nil && *lastAdjQ == \"0\"", J, "TestA137AnAdjustmentToDecimalZeroIsAnAdjustmentClose", ""),
 ("M28 환산 권한 없이 1:1", E+"a137_daily_loss.go", "\t\tauthority, ok := fx[market]\n\t\tif !ok {\n\t\t\treturn \"\", fmt.Errorf", "\t\tauthority, ok := fx[market]\n\t\tif !ok {\n\t\t\treturn magnitude, nil\n\t\t\treturn \"\", fmt.Errorf", E, "TestA137AForeignLossConvertsOnlyThroughThePairedAuthority", "tossos_testseams"),
 ("M2c 살아 있는 문 바닥 미적용", E+"riskguardian_qfinal.go", "Intent: intent, Account: account, Policy: g.policy, Costs: g.costs, AccountBaseFX: accountFX}", "Intent: intent, Account: request.Account, Policy: g.policy, Costs: g.costs, AccountBaseFX: accountFX}\n\t_ = account", E, "TestA137ThePrecheckDoorRefusesAMeasuredLossAtTheLimit", "tossos_testseams"),
 ("M1b 창 롤오버 삭제", E+"a137_daily_loss.go", "from := midnight.Add(clock.AccountDayRollover).UTC()", "from := midnight.UTC()", E, "TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead", "tossos_testseams"),
 ("M1c 읽기 여유 음수", E+"a137_daily_loss.go", "const dailyLossReadMargin = time.Hour", "const dailyLossReadMargin = -time.Hour", E, "TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead", "tossos_testseams"),
 ("M3b 펜스 단일 latch 무시", E+"a137_daily_loss.go", "if len(g.latches) != 1 || g.modeDailyLossDetail == \"\" {", "if len(g.latches) > 2 || g.modeDailyLossDetail == \"\" {", E, "TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange", ""),
 ("M29 살아 있는 문 FX 없음", E+"riskguardian_qfinal.go", "g.dailyLossFloor(request.Account, request.DailyLoss, now, qFinalManualBuyFX(request, market))", "g.dailyLossFloor(request.Account, request.DailyLoss, now, nil)", E, "TestA137ThePrecheckDoorConvertsAUSLossOnlyWithThePairedAuthority", "tossos_testseams"),
 ("M30 측정 불가 표지 미부착", E+"a137_daily_loss.go", "\t\trefusal.cause = &DailyLossUnmeasurable{Cause: measureErr}", "\t\t_ = measureErr", E, "TestA137ThePrecheckDoorRefusesAnUnmeasurableDayAsTyped", "tossos_testseams"),
 ("M31 루프가 측정 불가를 결함으로", G+"strategy_market_handoff_delivery.go", "if errors.As(err, &blocked) || errors.As(err, &unmeasurable) {", "if errors.As(err, &blocked) {\n\t\t\t_ = unmeasurable", G, "TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind", "tossos_testseams"),
 ("M32 승인의 측정 불가 갈래 삭제", G+"strategy_first_leg_admission.go", "if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {", "if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); false && errors.As(err, &unmeasurable) {", G, "TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault", "tossos_testseams"),
 ("M33 승인이 토큰을 버림", G+"strategy_first_leg_admission.go", "authority.Entry.DailyLoss = b.guardian.MeasureDailyLoss(ctx)", "_ = b.guardian.MeasureDailyLoss(ctx)", G, "TestA137AdmissionCarriesTheMeasuredTokenIntoThePrecheck", "tossos_testseams"),
 ("M38 측정 불가 계수 미증가", G+"a137_daily_loss_notice.go", "\t\tworker.dailyLossUnmeasurable++", "\t\t_ = worker.dailyLossUnmeasurable", G, "TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay", "tossos_testseams"),
 ("M39 알림이 주기마다", G+"a137_daily_loss_notice.go", "announce := day != \"\" && worker.dailyLossNoticeDay != day", "announce := day != \"\"", G, "TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay", "tossos_testseams"),
 ("M40 전달 자리 기록 삭제", G+"strategy_market_handoff_delivery.go", "\t\t\trecordStrategyDailyLossUnmeasurable(ctx, unmeasurable)", "\t\t\t_ = unmeasurable", G, "TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay", "tossos_testseams"),
 ("M41 원인 판정 재실행 삭제", E+"a137_daily_loss.go", "if again := evaluateChain(probe); !again.Allowed", "if again := evaluateChain(probe); false && !again.Allowed", E, "TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked", ""),
 ("M42 주기 ctx 기록자 미부착", G+"strategy_entry_supervisor.go", "cycleCtx := withStrategyDailyLossRecorder(ctx, func(cause error) { s.recordDailyLossUnmeasurable(worker, cause) })", "cycleCtx := ctx", G, "TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay", "tossos_testseams"),
 ("M43 생산 알림 배선 삭제", G+"strategy_entry_supervisor.go", "DailyLossNotice: c.strategyDailyLossNotice(),", "DailyLossNotice: nil,", G, "TestA137TheProductionSupervisorCarriesTheNotice", "tossos_testseams"),
 ("M44 표지를 다른 파일에서 지어냄", G+"strategy_first_leg_admission.go", "\t\t// a137 F8: 원인을 실음 — 하류가 errors.As 로 거절 타입을 볼 수 있게(문구 · 분류는 그대로).\n\t\trefusal.cause = err\n", "\t\trefusal.cause = errors.Join(err, &execgw.DailyLossEntryBlock{})\n", E, "TestA137TheQuietMarkersAreMadeOnlyInTheirOwnFile", ""),
]

def test(pkg, name, tags):
    args = ["go", "test", "-count=1", "-run", "^(" + name + ")$"]
    if tags:
        args += ["-tags", tags]
    return run(args + ["./" + pkg])

head = run(["git", "rev-parse", "HEAD"]).stdout.strip()
assert head == SHA, f"worktree HEAD {head} != expected {SHA}"
assert run(["git", "status", "--porcelain"]).stdout.strip() == "", "worktree not clean"
for _, _, _, _, pkg, name, tags in MUTANTS:
    r = test(pkg, name, tags)
    assert r.returncode == 0 and "no tests to run" not in r.stdout, f"control not GREEN: {name}\n{r.stdout[-800:]}"
lines = [f"# a137 변이 원장\n\n- 사본: 연결 워크트리 `{SHA[:8]}` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 {len(MUTANTS)} 시험 GREEN)\n",
         "| 변이 | 파일 | 시험 | 결과 |", "|---|---|---|---|"]
for label, path, old, new, pkg, name, tags in MUTANTS:
    full = os.path.join(WT, path)
    src = open(full).read()
    if src.count(old) != 1:
        lines.append(f"| {label} | `{path}` | `{name}` | NOT-REACHED (문자열 {src.count(old)}회) |")
        continue
    open(full, "w").write(src.replace(old, new))
    r = test(pkg, name, tags)
    verdict = "CAUGHT" if r.returncode != 0 else "SURVIVED"
    if "build failed" in r.stdout or "[build failed]" in r.stdout:
        verdict = "CAUGHT(build)"
    run(["git", "checkout", "--", path])
    assert run(["git", "status", "--porcelain"]).stdout.strip() == "", f"revert failed after {label}"
    lines.append(f"| {label} | `{path}` | `{name}` | {verdict} |")
    print(label, verdict, flush=True)
open(OUT, "w").write("\n".join(lines) + "\n")
