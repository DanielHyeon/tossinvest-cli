#!/usr/bin/env python3
"""a139 변이 원장 하네스 — 고정한 연결 워크트리(사본) 하나에서 변이를 하나씩 넣고 대상 시험을 돌린 뒤 되돌림.

사용: python3 a137_mutate.py <worktree> <expected-sha> <out.md>
- 시작 sha 를 단언하고, 사본이 깨끗한지(git status) 확인한 뒤, 무변이 대조군이 GREEN 이어야 시작함(기억: 대조군이 GREEN 이 아니면 멈출 것).
- 변이마다: 정확히 한 번 나타나는 문자열을 바꿈(아니면 NOT-REACHED) → 시험 → git checkout 으로 그 파일만 되돌림 → 사본이 다시 깨끗한지 확인.
"""
import os, subprocess, sys

WT, SHA, OUT = sys.argv[1], sys.argv[2], sys.argv[3]
ENV = dict(os.environ, GOFLAGS="-trimpath")

def run(args, **kw):
    return subprocess.run(args, cwd=WT, capture_output=True, text=True, env=ENV, **kw)

M, G, E = "internal/risk/", "internal/app/engine/", "internal/execgw/"
RB = M+"risk_budget.go"
GW = G+"guardian_wiring.go"
MUTANTS = [
 ("N1 나눗수 3 → 2", RB, "const PerTradeRiskBudgetDivisor = 3", "const PerTradeRiskBudgetDivisor = 2", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N2 내림 → 올림", RB, "floored := new(big.Int).Quo(minor.Num(), minor.Denom())", "floored := new(big.Int).Quo(new(big.Int).Add(minor.Num(), new(big.Int).Sub(minor.Denom(), big.NewInt(1))), minor.Denom())", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N3 USD 최소 단위 2 → 3", RB, "map[string]int{\"KRW\": 0, \"USD\": 2}", "map[string]int{\"KRW\": 0, \"USD\": 3}", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N4 KRW 최소 단위 0 → 2", RB, "map[string]int{\"KRW\": 0, \"USD\": 2}", "map[string]int{\"KRW\": 2, \"USD\": 2}", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N5 0 으로 내려가도 허용", RB, "\tif floored.Sign() <= 0 {", "\tif floored.Sign() < 0 {", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N6 모르는 통화를 KRW 처럼", RB, "\tscale, ok := riskBudgetMinorScale[currency]\n\tif !ok {", "\tscale, ok := riskBudgetMinorScale[currency]\n\tif false && !ok {", M, "TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit", ""),
 ("N7 생산 유도 우회(= dailyLoss)", GW, "\tif budget, err := risk.PerTradeRiskBudget(dailyLoss); err == nil {\n\t\tpolicy.RiskBudget = budget", "\tif budget, err := risk.PerTradeRiskBudget(dailyLoss); err == nil {\n\t\t_ = budget\n\t\tpolicy.RiskBudget = dailyLoss", G, "TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss", ""),
 ("N8 유도 실패 시 dailyLoss 로 폴백", GW, "\t\tpolicy.RiskBudget = riskcalc.Money{Currency: currency}", "\t\tpolicy.RiskBudget = dailyLoss", G, "TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName", ""),
 ("N9 기본 정책 값 되돌림", M+"input.go", "RiskBudget:        riskcalc.Money{Amount: \"33333\", Currency: \"KRW\"},", "RiskBudget:        riskcalc.Money{Amount: \"100000\", Currency: \"KRW\"},", M, "TestA139TheDefaultPolicyFollowsTheSameRule", ""),
 ("N10 audit 줄 삭제", G+"interlock.go", "{audit.ActionLimitChange, \"engine.automation_gate.per_trade_risk_budget\",", "{audit.ActionLimitChange, \"engine.automation_gate.per_trade_risk_budget_gone\",", G, "TestA139TheInterlockAuditsTheDerivedRiskBudget", ""),
 ("N11 audit 만 dailyLoss", GW, "value: budget.Amount, detail: \"derived by \" + perTradeRiskBudgetRule", "value: strconv.FormatFloat(gate.MaxDailyLossAmount, 'f', -1, 64), detail: \"derived by \" + perTradeRiskBudgetRule", G, "TestA139TheInterlockAuditsTheDerivedRiskBudget", ""),
 ("N12 factory 사전 검사 삭제", GW, "\tif audit := perTradeRiskBudgetAudit(gate); audit.value == \"\" {", "\tif audit := perTradeRiskBudgetAudit(gate); false && audit.value == \"\" {", G, "TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName", ""),
 ("N13 버전 v1 되돌림", GW, "const productionGuardianPolicyVersion = \"engine.automation_gate/risk-policy-v2\"", "const productionGuardianPolicyVersion = \"engine.automation_gate/risk-policy-v1\"", G, "TestA139TheProductionPolicyVersionNamesTheNewRule", ""),
 ("N14 유도 불가 사유 삭제", GW, "return riskBudgetAudit{detail: \"cannot be derived: \" + reason", "return riskBudgetAudit{detail: \"\" + reason", G, "TestA139AGateOffStartWithoutADailyLossStillStarts", ""),
 ("N15 acceptance 칸 삭제", G+"interlock.go", "\t\tstatus.PerTradeRiskBudget,\n", "\t\t\"\",\n", G, "TestA139TheInterlockAuditsTheDerivedRiskBudget", ""),
 ("N16 생성자에서 예산 되돌림(구현 리뷰 P1-1)", GW, "\tguardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{", "\tpolicy.RiskBudget = policy.MaxDailyLoss\n\tguardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{", G, "TestA139TheProductionGuardianSizesWithAThirdOfTheDailyLoss", ""),
 ("N17 유도 불가 audit 줄 건너뜀(구현 리뷰 P3-1)", G+"interlock.go", "\tfor _, c := range changes {\n", "\tfor _, c := range changes {\n\t\tif c.setting == \"engine.automation_gate.per_trade_risk_budget\" && c.value == \"\" {\n\t\t\tcontinue\n\t\t}\n", G, "TestA139AGateOffStartWithoutADailyLossStillStarts", ""),
 ("N18 account-base 경로가 일일 손실 전부", M+"account_base.go", "budget, err := parseDecimal(\"risk budget\", policy.RiskBudget.Amount)", "budget, err := parseDecimal(\"risk budget\", policy.MaxDailyLoss.Amount)", E, "TestA139TheAccountBasePathSizesWithAThirdOfTheDailyLoss", "tossos_testseams"),
 ("N19 매도가 예산을 읽음", M+"chain.go", "\tif in.Intent.Side == SideSell {\n\t\treturn at(StepReduction, evaluateReduction(in))", "\tif in.Intent.Side == SideSell {\n\t\tif in.Policy.RiskBudget.Amount == \"\" {\n\t\t\treturn at(StepReduction, refuse(ReasonInputUnavailable, \"budget\"))\n\t\t}\n\t\treturn at(StepReduction, evaluateReduction(in))", M, "TestA139ASellNeverReadsTheRiskBudget", ""),
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
lines = [f"# a139 변이 원장\n\n- 사본: 연결 워크트리 `{SHA[:8]}` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 {len(MUTANTS)} 시험 GREEN)\n",
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
