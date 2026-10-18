#!/usr/bin/env python3
"""a157 변이 하네스(design D6 · tasks 2.2).

규율(기억 「뮤테이션은 대상에 닿아야 한다」 · 「뮤테이션 원복은 baseline 이 맞아야 한다」 · 「병행 세션 게이트 간섭」):
  - 저장소 밖 사본에서만 변이함: <scratch>/a157-mut-<pid>/ (pid 접미 — 두 판이 한 사본을 쓰지 않게). 저장소 파일은 읽기만 함.
  - 시작 sha 단언: 사본의 대상 파일 sha == 저장소 HEAD blob sha(커밋된 GREEN). 다르면 멈춤.
  - 무변이 대조군이 GREEN 이어야 변이를 셈(아니면 멈춤).
  - 변이마다: 바꿀 문자열이 정확히 한 번 있는지 단언 → 적용 → 대상 함수 본문 sha 가 바뀌었는지(닿음 증거 — `_funcsha`) → 빌드 → 시험 →
    원래 바이트로 되돌리고 파일 sha 를 다시 단언(`git diff --quiet` 아님).
  - 한 번에 한 판(순차). GOFLAGS=-trimpath + 사본 전용 GOCACHE.
사용: python3 a157_mutate.py <scratch-dir> [mutant-id ...]   → 결과 JSON 을 stdout 마지막 줄에, 진행은 stderr 에.
"""
import hashlib, json, os, re, shutil, subprocess, sys, time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
FUNCSHA = "./openspec/changes/a157-first-leg-binding-before-sizing/analysis/harness/_funcsha"
FIRST = "internal/execgw/riskguardian_first_leg.go"
KIND = "internal/execgw/a157_daily_loss_kind.go"
ADMIT = "internal/app/engine/strategy_first_leg_admission.go"
SUP = "internal/app/engine/strategy_entry_supervisor.go"
TARGET_FILES = [FIRST, KIND, ADMIT, SUP]
FUNC = {FIRST: "RiskGuardian.PrecheckQFinalCampaignFirstLeg", KIND: "DailyLossRefusalKindOf",
        ADMIT: "strategyFirstLegAdmissionBridge.admit", SUP: "Context.NewPairedStrategyEntryProductionAssembly"}
TESTS = [
    ["go", "test", "-count=1", "-tags", "tossos_testseams", "./internal/execgw/",
     "-run", "TestA157|TestQFinalCampaignFirstLeg|TestA136QFinalCampaignFirstLeg|TestA138QFinalCampaignFirstLeg|TestA137"],
    ["go", "test", "-count=1", "-tags", "tossos_testseams", "./internal/app/engine/",
     "-run", "TestA157|TestA137|TestA146|TestStrategyFirstLegAdmission"],
]

# 결속 블록이 끝난 뒤(진입 판정 · 결속 결함 반환 다음) — 「사이징 뒤로 옮김」 변이가 판정을 끼워 넣는 자리.
AFTER = "\tif bindErr != nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, bindErr\n\t}\n"

B2_IN = ("\t\tversion, err := journal.QFinalPolicyVersion(g.policyVersion, request.Entry.Admission.TransactionID)\n"
         "\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tpolicyVersion = version\n")
B2_OUT = ("\tversion, versionErr := journal.QFinalPolicyVersion(g.policyVersion, request.Entry.Admission.TransactionID)\n"
          "\tif versionErr != nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, versionErr\n\t}\n\tpolicyVersion = version\n")
B3_IN = ("\t\tif strings.TrimSpace(owner.Key.ProspectiveGeneration) != \"\" {\n"
         "\t\t\treturn qFinalRefusal(riskbucket.RefusalOwnerConflict, \"prospective_generation\", \"first-leg token is journal-owned\", nil)\n\t\t}\n")
B3_OUT = ("\tif strings.TrimSpace(request.Entry.Admission.Owner.Key.ProspectiveGeneration) != \"\" {\n"
          "\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalOwnerConflict, \"prospective_generation\", \"first-leg token is journal-owned\", nil)\n\t}\n")
B4_IN = ("\t\tif !entryOK || !stopOK || !targetOK {\n\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
         "\t\t\t\t\"first_leg_price_unit\", \"sealed strategy execution price unit is invalid\", nil)\n\t\t}\n")
B4_OUT = ("\tif _, ok := terms.Entry().MajorDecimal(); !ok {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
          "\t\t\t\"first_leg_price_unit\", \"sealed strategy execution price unit is invalid\", nil)\n\t}\n")
B5_HEAD = "\t\tif request.Result.Quantity != request.Entry.QCandidate || lineage.AccountRef != g.accountRef ||\n"
B5_IN = (B5_HEAD +
         "\t\t\t!strings.EqualFold(string(lineage.Market), request.Entry.Market) || lineage.Symbol != request.Entry.Symbol ||\n"
         "\t\t\tentryMajor != request.Entry.EntryPrice || stopMajor != request.Entry.StopPrice ||\n"
         "\t\t\ttargetMajor != request.Entry.TargetPrice || lineage.LaneID != owner.LaneID ||\n"
         "\t\t\tlineage.CampaignID != owner.CampaignID || request.Campaign.CampaignID != owner.CampaignID {\n"
         "\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
         "\t\t\t\t\"first_leg_lineage\", \"sealed strategy result, q_candidate, owner and campaign are not exactly bound\", nil)\n\t\t}\n")
B5_OUT = ("\tentryMajor, _ := terms.Entry().MajorDecimal()\n\tstopMajor, _ := terms.EffectiveStop().MajorDecimal()\n"
          "\ttargetMajor, _ := terms.Target().MajorDecimal()\n\towner := request.Entry.Admission.Owner\n"
          "\tif request.Result.Quantity != request.Entry.QCandidate || lineage.AccountRef != g.accountRef ||\n"
          "\t\t!strings.EqualFold(string(lineage.Market), request.Entry.Market) || lineage.Symbol != request.Entry.Symbol ||\n"
          "\t\tentryMajor != request.Entry.EntryPrice || stopMajor != request.Entry.StopPrice ||\n"
          "\t\ttargetMajor != request.Entry.TargetPrice || lineage.LaneID != owner.LaneID ||\n"
          "\t\tlineage.CampaignID != owner.CampaignID || request.Campaign.CampaignID != owner.CampaignID {\n"
          "\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
          "\t\t\t\"first_leg_lineage\", \"sealed strategy result, q_candidate, owner and campaign are not exactly bound\", nil)\n\t}\n")
B6_IN = ("\t\tcurrentGeneration := request.Campaign.ExpectedPositionGeneration\n"
         "\t\tif currentGeneration < 0 || currentGeneration >= math.MaxInt64 || request.Campaign.ExpectedPositionVersion < 0 ||\n"
         "\t\t\tlineage.PositionGeneration != uint64(currentGeneration)+1 {\n\t\t\treturn qFinalRefusal(riskbucket.RefusalOwnerConflict,\n"
         "\t\t\t\t\"position_generation\", \"strategy prospective generation must be exactly campaign current generation plus one\", nil)\n\t\t}\n")
B6_OUT = ("\tcurrentGeneration := request.Campaign.ExpectedPositionGeneration\n"
          "\tif currentGeneration < 0 || currentGeneration >= math.MaxInt64 || request.Campaign.ExpectedPositionVersion < 0 ||\n"
          "\t\tlineage.PositionGeneration != uint64(currentGeneration)+1 {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalOwnerConflict,\n"
          "\t\t\t\"position_generation\", \"strategy prospective generation must be exactly campaign current generation plus one\", nil)\n\t}\n")
B7_IN = ("\t\tif strings.TrimSpace(request.AttemptID) == \"\" || request.AttemptID != strings.TrimSpace(request.AttemptID) || request.Revision < 1 ||\n"
         "\t\t\tstrings.TrimSpace(request.ActivationManifestDigest) == \"\" || request.ActivationManifestDigest != strings.TrimSpace(request.ActivationManifestDigest) {\n"
         "\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
         "\t\t\t\t\"first_leg_strategy\", \"complete strategy attempt and activation binding required\", nil)\n\t\t}\n")
B7_OUT = ("\tif strings.TrimSpace(request.AttemptID) == \"\" || request.AttemptID != strings.TrimSpace(request.AttemptID) || request.Revision < 1 ||\n"
          "\t\tstrings.TrimSpace(request.ActivationManifestDigest) == \"\" || request.ActivationManifestDigest != strings.TrimSpace(request.ActivationManifestDigest) {\n"
          "\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
          "\t\t\t\"first_leg_strategy\", \"complete strategy attempt and activation binding required\", nil)\n\t}\n")
B8_IN = ("\t\tbinding, err := guardianWeeklyFirstLegBinding(lineage.LaneID, string(lineage.Market), request.Weekly)\n"
         "\t\tif err != nil {\n\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
         "\t\t\t\t\"weekly_reservation\", \"weekly first-leg reservation authority is missing or does not match the sealed lane\", err)\n\t\t}\n"
         "\t\tweekly = binding\n")
B8_OUT = ("\tbinding, weeklyErr := guardianWeeklyFirstLegBinding(lineage.LaneID, string(lineage.Market), request.Weekly)\n"
          "\tif weeklyErr != nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n"
          "\t\t\t\"weekly_reservation\", \"weekly first-leg reservation authority is missing or does not match the sealed lane\", weeklyErr)\n\t}\n"
          "\tweekly = binding\n")
B5_KEEP_VARS = "\t\t_, _, _ = entryMajor, stopMajor, targetMajor\n"


def move(inside, outside, keep=""):
    """결속 판정 하나를 닫힌 함수에서 빼서 진입 판정 · 결속 결함 반환 뒤로 옮김(부분 이동)."""
    return [(inside, keep), (AFTER, AFTER + outside)]


MUTANTS = {
    # M1: 편집 전 순서 — 진입 판정을 결속 판정 앞에서 부르고 그 오류를 바로 반환(결속 블록이 사이징 뒤로 간 것과 같은 보고).
    "M1-sizing-first": (FIRST, [
        ("\tvar policyVersion string\n", "\tentry, err := g.PrecheckQFinalEntry(request.Entry)\n\tif err != nil {\n"
                                          "\t\treturn QFinalCampaignFirstLegPrecheck{}, err\n\t}\n\tvar policyVersion string\n"),
        ("\tentry, err := g.PrecheckQFinalEntry(request.Entry)\n\t// 진입 거절을", "\t// 진입 거절을"),
        ("\tif err != nil && (bindErr == nil || DailyLossRefusalKindOf(err) != DailyLossRefusalKindNone) {\n"
         "\t\treturn QFinalCampaignFirstLegPrecheck{}, err\n\t}\n", ""),
    ]),
    "M2-move-B2": (FIRST, move(B2_IN, B2_OUT)),
    "M3-move-B3": (FIRST, move(B3_IN, B3_OUT)),
    "M4-move-B4": (FIRST, move(B4_IN, B4_OUT, "\t\t_, _, _ = entryOK, stopOK, targetOK\n")),
    "M5-move-B5": (FIRST, move(B5_IN, B5_OUT, B5_KEEP_VARS)),
    "M6-move-B6": (FIRST, move(B6_IN, B6_OUT)),
    "M7-move-B7": (FIRST, move(B7_IN, B7_OUT)),
    "M8-move-B8": (FIRST, move(B8_IN, B8_OUT)),
    # M9: 판정 하나씩 삭제 — 그 갈래의 거절을 반환하지 않게(반환문을 버리는 대입으로). 조건식 · 변수 사용은 그대로라 컴파일됨.
    "M9-drop-B2": (FIRST, [("\t\t\treturn err\n\t\t}\n\t\tpolicyVersion = version", "\t\t\t_ = err\n\t\t}\n\t\tpolicyVersion = version")]),
    "M9-drop-B3": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalOwnerConflict, \"prospective_generation\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalOwnerConflict, \"prospective_generation\"")]),
    "M9-drop-B4": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_price_unit\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_price_unit\"")]),
    "M9-drop-B5": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_lineage\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_lineage\"")]),
    "M9-drop-B6": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalOwnerConflict,\n\t\t\t\t\"position_generation\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalOwnerConflict,\n\t\t\t\t\"position_generation\"")]),
    "M9-drop-B7": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_strategy\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"first_leg_strategy\"")]),
    "M9-drop-B8": (FIRST, [("\t\t\treturn qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"weekly_reservation\"",
                            "\t\t\t_ = qFinalRefusal(riskbucket.RefusalRiskCalculationInvalid,\n\t\t\t\t\"weekly_reservation\"")]),
    "M10-drop-nil-guard": (FIRST, [("\tif g == nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}", "\tif false && g == nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}")]),
    "M11-B5-qcandidate-self": (FIRST, [("request.Result.Quantity != request.Entry.QCandidate", "request.Result.Quantity != request.Result.Quantity")]),
    "M12a-B5-lane-self": (FIRST, [("lineage.LaneID != owner.LaneID", "lineage.LaneID != lineage.LaneID")]),
    "M12b-B5-campaign-self": (FIRST, [("lineage.CampaignID != owner.CampaignID ||", "lineage.CampaignID != lineage.CampaignID ||")]),
    "M13-drop-daily-loss-precedence": (FIRST, [("(bindErr == nil || DailyLossRefusalKindOf(err) != DailyLossRefusalKindNone)", "bindErr == nil")]),
    "M13b-invert-daily-loss-test": (FIRST, [("DailyLossRefusalKindOf(err) != DailyLossRefusalKindNone", "DailyLossRefusalKindOf(err) == DailyLossRefusalKindNone")]),
    "M14a-kind-drop-unmeasurable": (KIND, [("errors.As(err, &unmeasurable) {", "false && errors.As(err, &unmeasurable) {")]),
    "M14b-kind-drop-limit": (KIND, [("\t\tissue.Reason == string(risk.ReasonDailyLossLimitReached) {", "\t\tfalse && issue.Reason == string(risk.ReasonDailyLossLimitReached) {")]),
    "M14c-kind-unmeasurable-any-stage": (KIND, [("issue.Stage == StageChain &&\n\t\t\tissue.Reason == string(risk.ReasonInputUnavailable)", "issue.Reason == string(risk.ReasonInputUnavailable)")]),
    "M14d-kind-unmeasurable-without-marker": (KIND, [("errors.As(err, &unmeasurable) {", "true || errors.As(err, &unmeasurable) {")]),
    "M15-duplicate-B3-after-sizing": (FIRST, [(AFTER, AFTER + B3_OUT)]),
    # round2(구현 리뷰 P2): 리터럴은 그대로 두고 결속 판정의 **실행** 횟수 · 시점만 바꾸는 변이 — 행동이 같아 구조 시험만 잡을 수 있음.
    "M15b-named-closure-called-twice": (FIRST, [
        ("\tbindErr := func() error {\n", "\tbind := func() error {\n"),
        ("\t\tweekly = binding\n\t\treturn nil\n\t}()\n", "\t\tweekly = binding\n\t\treturn nil\n\t}\n\tbindErr := bind()\n"),
        ("\tissuedAt := g.clk.Now().UTC()\n", "\tif again := bind(); again != nil {\n\t\treturn QFinalCampaignFirstLegPrecheck{}, again\n\t}\n\tissuedAt := g.clk.Now().UTC()\n"),
    ]),
    "M15c-sole-invocation-after-sizing": (FIRST, [
        ("\tbindErr := func() error {\n", "\tbind := func() error {\n"),
        ("\t\tweekly = binding\n\t\treturn nil\n\t}()\n", "\t\tweekly = binding\n\t\treturn nil\n\t}\n"),
        ("\tentry, err := g.PrecheckQFinalEntry(request.Entry)\n", "\tentry, err := g.PrecheckQFinalEntry(request.Entry)\n\tbindErr := bind()\n"),
    ]),
    "M15d-invocation-inside-a-loop": (FIRST, [
        ("\tbindErr := func() error {\n", "\tvar bindErr error\n\tfor range 1 {\n\tbindErr = func() error {\n"),
        ("\t\tweekly = binding\n\t\treturn nil\n\t}()\n", "\t\tweekly = binding\n\t\treturn nil\n\t}()\n\t}\n"),
    ]),
    # round2(구현 리뷰 P3-4 · codex P3-2): 통과 봉인 값 — 결속 판정이 봉인에 넘기는 두 값의 대입 삭제.
    "M18-drop-policy-version-assignment": (FIRST, [("\t\tpolicyVersion = version\n", "\t\t_ = version\n")]),
    "M19-drop-weekly-assignment": (FIRST, [("\t\tweekly = binding\n", "\t\t_ = binding\n")]),
    "M16a-entrygate-no-nil-guard": (SUP, [("\tif c.Entry != nil { //", "\tif true { //")]),
    "M16b-drop-daily-loss-notice": (SUP, [("DailyLossNotice: c.strategyDailyLossNotice(), EntrySkipNotice: c.strategyEntrySkipNotice()}",
                                           "EntrySkipNotice: c.strategyEntrySkipNotice()}")]),
    "M16c-drop-entry-skip-notice": (SUP, [("DailyLossNotice: c.strategyDailyLossNotice(), EntrySkipNotice: c.strategyEntrySkipNotice()}",
                                           "DailyLossNotice: c.strategyDailyLossNotice()}")]),
    "M16d-drop-entrygate": (SUP, [("\t\topts.EntryGate = c.Entry\n", "\t\t_ = c.Entry\n")]),
    "M17-admit-ignores-classifier": (ADMIT, [("switch execgw.DailyLossRefusalKindOf(err) {", "switch execgw.DailyLossRefusalKindNone {")]),
}


def log(*a):
    print(*a, file=sys.stderr, flush=True)


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def run(cmd, cwd, env, timeout=1800):
    started = time.time()
    p = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
    return p.returncode, p.stdout + p.stderr, round(time.time() - started, 1)


def funcsha(path, fn, env):
    code, out, _ = run(["go", "run", FUNCSHA, str(path), fn], ROOT, env)
    if code != 0:
        raise SystemExit(f"funcsha failed: {out}")
    return out.split()[0]


def failures(out):
    names = sorted(set(re.findall(r"--- FAIL: (\S+)", out)))
    if "panic:" in out:
        names.append("panic")
    if "[build failed]" in out or "[setup failed]" in out:
        names.append("build-failed")
    return names


def main():
    scratch = Path(sys.argv[1]).resolve()
    only = sys.argv[2:]
    copy = scratch / f"a157-mut-{os.getpid()}"
    if copy.exists():
        raise SystemExit(f"copy already exists: {copy}")
    copy.mkdir(parents=True)
    for name in ["go.mod", "go.sum"]:
        shutil.copy2(ROOT / name, copy / name)
    for name in ["internal", "cmd", "tools", "fixtures", "schemas"]:
        if (ROOT / name).exists():
            shutil.copytree(ROOT / name, copy / name, symlinks=True)
    env = dict(os.environ, GOFLAGS="-trimpath", GOCACHE=str(scratch / "gocache-a157"), GOTOOLCHAIN="local")
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()
    start = {}
    for f in TARGET_FILES:
        blob = subprocess.run(["git", "-C", str(ROOT), "show", f"HEAD:{f}"], capture_output=True, check=True).stdout
        start[f] = hashlib.sha256(blob).hexdigest()
        if sha(copy / f) != start[f]:
            raise SystemExit(f"start sha mismatch for {f}: the copy is not the committed GREEN at {head}")
    log(f"copy={copy} head={head}")
    results = {"head": head, "copy": str(copy), "start_sha": start, "control": None, "mutants": []}
    outs = [run(cmd, copy, env) for cmd in TESTS]
    control = {"failed": sum((failures(o) for _, o, _ in outs), []), "codes": [c for c, _, _ in outs], "seconds": [s for _, _, s in outs]}
    results["control"] = control
    log("control", control)
    if any(control["codes"]) or control["failed"]:
        print(json.dumps(results))
        raise SystemExit("unmutated control is not GREEN — no mutant counts")
    for mid, (rel, pairs) in MUTANTS.items():
        if only and mid not in only:
            continue
        path = copy / rel
        original = path.read_bytes()
        text = original.decode()
        for old, new in pairs:
            if text.count(old) != 1:
                raise SystemExit(f"{mid}: expected exactly one occurrence of {old[:60]!r}, found {text.count(old)}")
            text = text.replace(old, new)
        before = funcsha(path, FUNC[rel], env)
        path.write_text(text)
        try:
            after = funcsha(path, FUNC[rel], env)
            reached = before != after
            code, out, secs = run(["go", "build", "./internal/...", "./cmd/..."], copy, env)
            if code != 0:
                entry = {"id": mid, "file": rel, "reached": reached, "verdict": "BUILD-FAIL", "detail": out[-400:]}
            else:
                outs = [run(cmd, copy, env) for cmd in TESTS]
                failed = sum((failures(o) for _, o, _ in outs), [])
                entry = {"id": mid, "file": rel, "reached": reached, "verdict": "CAUGHT" if (failed or any(c for c, _, _ in outs)) else "SURVIVED",
                         "failed": failed, "seconds": [s for _, _, s in outs]}
        finally:
            path.write_bytes(original)
            if sha(path) != start[rel]:
                raise SystemExit(f"{mid}: restore failed — {rel} sha differs from the start sha")
        log(mid, entry["verdict"], entry.get("reached"), entry.get("failed") or entry.get("detail", ""))
        results["mutants"].append(entry)
    for f in TARGET_FILES:
        if sha(copy / f) != start[f]:
            raise SystemExit(f"end sha mismatch for {f}")
    print(json.dumps(results, ensure_ascii=False))


if __name__ == "__main__":
    main()
