#!/usr/bin/env python3
"""a138 변이 원장 하네스 — 고정한 연결 워크트리(사본) 하나에서 변이를 하나씩 넣고 대상 시험을 돌린 뒤 되돌림.

사용: python3 a137_mutate.py <worktree> <expected-sha> <out.md>
- 시작 sha 를 단언하고, 사본이 깨끗한지(git status) 확인한 뒤, 무변이 대조군이 GREEN 이어야 시작함(기억: 대조군이 GREEN 이 아니면 멈출 것).
- 변이마다: 정확히 한 번 나타나는 문자열을 바꿈(아니면 NOT-REACHED) → 시험 → git checkout 으로 그 파일만 되돌림 → 사본이 다시 깨끗한지 확인.
"""
import os, subprocess, sys

WT, SHA, OUT = sys.argv[1], sys.argv[2], sys.argv[3]
ENV = dict(os.environ, GOFLAGS="-trimpath")

def run(args, **kw):
    return subprocess.run(args, cwd=WT, capture_output=True, text=True, env=ENV, **kw)

E, J, C, G, K, M = "internal/execgw/", "internal/journal/", "internal/config/", "internal/app/engine/", "internal/console/", "internal/risk/"
OP = J+"open_positions.go"
MUTANTS = [
 ("M1 초과(>) → 도달(>=)", OP, "count > req.OpenPositionLimit.Max", "count >= req.OpenPositionLimit.Max", J, "TestA138TheSeventhSymbolIsRefused", ""),
 ("M2 든 종목 scale-in 통과 삭제", OP, "if _, already := held[this]; already {", "if _, already := held[this]; false && already {", J, "TestA138AnEntryOnAHeldSymbolIsAllowedAtTheMaximum", ""),
 ("M3 ② HELD 예약 종목 버림", OP, "\t\tset[v] = struct{}{}\n\t}\n\treturn set, nil", "\t\t_ = v\n\t}\n\treturn set, nil", J, "TestA138TheSeventhSymbolIsRefused", ""),
 ("M4 ① 외부 · 편입 보유 버림", OP, "FROM positions WHERE account_ref = ? AND state <> ?", "FROM positions WHERE account_ref = ? AND state <> ? AND entry_decision_id IS NOT NULL", J, "TestA138EveryKindOfLivePositionCounts", ""),
 ("M5 노출 미증가 조기 통과 반전", OP, "if safetyClass != string(SafetyClassExposureRaising) {", "if safetyClass == string(SafetyClassExposureRaising) {", J, "TestA138TheSeventhSymbolIsRefused", ""),
 ("M6 상한 없는 예약 통과", OP, "if req.OpenPositionLimit == nil || req.OpenPositionLimit.Max < 1 {", "if req.OpenPositionLimit == nil {\n\t\treturn nil\n\t}\n\tif req.OpenPositionLimit.Max < 1 {", J, "TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused", ""),
 ("M7 preimage 결함을 0 으로", OP, "\t\t\treturn nil, fmt.Errorf(\"journal: the open position count cannot be computed: %w\", err)", "\t\t\tcontinue", J, "TestA138AnUncountableHoldFailsClosed", ""),
 ("M8 IssueEntry 문 상한 누락", E+"riskguardian.go", "\t\t\t\tOpenPositionLimit: g.openPositionLimit(),\n", "", E, "TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum", ""),
 ("M9 인터록 7c 삭제", G+"interlock.go", "if err := checkGuardianOpenPositionLimit(facts.guardian, *status); err != nil {", "if err := error(nil); err != nil {", G, "TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum", ""),
 ("M10 digest 입력 삭제", E+"riskguardian.go", "\"\\nmax_open_positions=\" + strconv.Itoa(g.policy.MaxOpenPositions)", "\"\"", E, "TestA138TheLimitsDigestBindsTheOpenPositionMaximum", ""),
 ("M11 ③ 이번 종목 버림", OP, "count := len(held) + reconciled + 1", "count := len(held) + reconciled + 0", J, "TestA138TheSeventhSymbolIsRefused", ""),
 ("M12 CLOSED 필터 삭제", OP, "\"SELECT market, symbol FROM positions WHERE account_ref = ? AND state <> ?\", accountRef, PositionClosed)", "\"SELECT market, symbol FROM positions WHERE account_ref = ? AND ? <> ''\", accountRef, PositionClosed)", J, "TestA138PositionsAndHoldsOfOneSymbolCountOnceAndClosedOnesNotAtAll", ""),
 ("M13 시장 차원 무시", OP, "return positionVenue{market: normaliseMarket(market), symbol: normaliseSymbol(symbol)}", "return positionVenue{market: \"\", symbol: normaliseSymbol(symbol)}", J, "TestA138PositionsAndHoldsOfOneSymbolCountOnceAndClosedOnesNotAtAll", ""),
 ("M14 HELD 필터 삭제", OP, "AND r.state = ? AND d.safety_class = ?", "AND r.state <> '' AND ? <> '' AND d.safety_class = ?", J, "TestA138EveryHeldReservationKindCountsAndReleasedOnesDoNot", ""),
 ("M15 종류를 OPEN_EXPOSURE 만", OP, "r.kind IN (?, ?, ?)", "r.kind IN (?, ?, ?) AND r.kind = 'OPEN_EXPOSURE'", J, "TestA138EveryHeldReservationKindCountsAndReleasedOnesDoNot", ""),
 ("M16 ② safety_class 필터 삭제", OP, "AND r.state = ? AND d.safety_class = ?`,", "AND r.state = ? AND ? <> ''`,", J, "TestA138TheSeventhSymbolIsRefused", ""),
 ("M17 ④ RECONCILE 칸 버림", OP, "\tif count := len(held) + reconciled + 1;", "\treconciled = 0\n\tif count := len(held) + reconciled + 1;", J, "TestA138AReconciledSymbolWithoutAPositionRowCounts", ""),
 ("M18 정수 판정 IsInf 삭제", C+"open_positions.go", "math.IsNaN(v) || math.IsInf(v, 0) || v < 1", "math.IsNaN(v) || v < 1", C, "TestA138WholeOpenPositionsIsTheOneIntegerJudgement", ""),
 ("M19 정수 판정 소수 검사 삭제", C+"open_positions.go", " || v != math.Trunc(v) {", " {", C, "TestA138WholeOpenPositionsIsTheOneIntegerJudgement", ""),
 ("M20 factory 사전 거부 삭제", G+"guardian_wiring.go", "if v := gate.EffectiveMaxOpenPositions(); !wholeOpenPositions(v) {", "if v := gate.EffectiveMaxOpenPositions(); false && !wholeOpenPositions(v) {", G, "TestA138ProductionPolicyResolvesTheOpenPositionMaximum", ""),
 ("M21 Validate 1 미만 → 0 미만", M+"input.go", "if p.MaxOpenPositions < 1 {", "if p.MaxOpenPositions < 0 {", M, "TestA138PolicyValidationRefusesAnUnusableOpenPositionMaximum", ""),
 ("M22 콘솔 칸 없는 폼도 씀", K+"settings_limits.go", "if _, present := r.PostForm[\"max_open_positions\"]; present {", "if true {", K, "TestA138TheAdvancedFormSavesTheOpenPositionMaximum", ""),
 ("M23 편집기 검증 삭제", C+"limits.go", "\t\tif _, ok := WholeOpenPositions(*l.OpenPositions); !ok {", "\t\tif _, ok := WholeOpenPositions(*l.OpenPositions); false && !ok {", C, "TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked", ""),
 ("M24 기동 audit 줄 삭제", G+"interlock.go", "{audit.ActionLimitChange, \"engine.automation_gate.max_open_positions\",", "{audit.ActionLimitChange, \"engine.automation_gate.max_open_positions_gone\",", G, "TestA138TheInterlockAuditsTheOpenPositionMaximum", ""),
 ("M25 기본값 0", C+"open_positions.go", "\treturn DefaultMaxOpenPositions\n}", "\treturn 0\n}", C, "TestA138OpenPositionMaximumResolution", ""),
 ("M26 q_final 문 상한 누락", E+"riskguardian_qfinal.go", "\t\t\t\tOpenPositionLimit: g.openPositionLimit(),\n", "", E, "TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum", ""),
 ("M27 첫 다리 문 상한 누락", E+"riskguardian_first_leg.go", "\t\t\t\t\tOpenPositionLimit: g.openPositionLimit(),\n", "", E, "TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum", "tossos_testseams"),
 ("M28 콘솔 행 삭제", K+"settings_limits.go", "\t\t{\"최대 보유 포지션 수\", openPositionsText(p.Gate.MaxOpenPositions), openPositionsUnit(p.Gate.MaxOpenPositions)},\n", "", K, "TestA138TheScreenSaysWhatTheOpenPositionMaximumIs", ""),
 ("M29 콘솔 저장 audit 삭제", "cmd/tossctl/limitsettings.go", "\t\trecordOpenPositionsSave(openFrom, after.MaxOpenPositions)\n", "\t\t_ = openFrom\n", "cmd/tossctl/", "TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile", ""),
 ("M30 매도가 판정을 지남(IssueReduction 에 판정 삽입)", E+"riskguardian.go", "func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n", "func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n\tif g.policy.MaxOpenPositions > 0 {\n\t\treturn Issued{}, errors.New(\"MAX_OPEN_POSITIONS\")\n\t}\n", E, "TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum", ""),
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
lines = [f"# a138 변이 원장\n\n- 사본: 연결 워크트리 `{SHA[:8]}` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 {len(MUTANTS)} 시험 GREEN)\n",
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
