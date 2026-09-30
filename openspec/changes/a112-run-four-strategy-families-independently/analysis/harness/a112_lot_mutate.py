#!/usr/bin/env python3
"""a112 로트 5.6.2 · 5.2.2 변이 하네스 — a092 `mutate_unit2.py` 의 사본 방식(저장소 추적 파일을 pid 붙은 사본으로 복사 · 무변이
대조군이 GREEN 이 아니면 멈춤 · 변이는 한 번에 하나, 사본에서만 · 판정 CAUGHT/SURVIVED/BUILD-FAIL).

사용: python3 a112_lot_mutate.py --set 5.6.2.1|5.2.2.1|5.2.2.1-fix <scratch-dir> [로트 파일 경로 …]
(N 으로 시작하는 변이는 동작이 같은 리팩터 — GREEN-AS-EXPECTED 여야 한다.)
"""
from __future__ import annotations

import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())

SUP = "internal/app/engine/strategy_entry_supervisor.go"
FC = "internal/execgw/failclosed.go"
RT = "internal/execgw/retry.go"
SET_5621 = [
    ("E01 blocker claims success without latching", SUP,
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\ts.entry.Block(",
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\tif true {\n\t\treturn true\n\t}\n\ts.entry.Block("),
    ("E02 central fault swallowed again", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if false && isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {"),
    ("E03 central fault on refresh-only stops the engine", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if isCentralStrategyIntegrity(err) {"),
    ("E04 every refresh-only error blocks entry", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if !s.blockEntryOnCentralIntegrity(worker) || isCentralStrategyIntegrity(err) && false {"),
    ("E05 nil gate swallows instead of escalating", SUP,
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\ts.entry.Block(",
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn true\n\t}\n\ts.entry.Block("),
    ("E06 production supervisor not given the gate", SUP,
     "Workers: workers, EntryGate: c.Entry,", "Workers: workers,"),
    ("E07 production supervisor built without a gate", SUP,
     "\tif c.Entry == nil {\n\t\treturn nil, fmt.Errorf(\"%w: the strategy refresh supervisor needs the entry gate\"",
     "\tif false {\n\t\treturn nil, fmt.Errorf(\"%w: the strategy refresh supervisor needs the entry gate\""),
    ("E08 option dropped by the constructor", SUP,
     "\t\tentry: opts.EntryGate,\n", ""),
    ("E09 wrong reason code", SUP,
     "s.entry.Block(execgw.ReasonStrategyCentralIntegrity,", "s.entry.Block(execgw.ReasonStrategyDispatchFenced,"),
    ("E10 reason not registered in the vocabulary", FC,
     # 2026-09-30 리뷰 보이스 B #3: 옛 앵커 `…,\n\t}` 는 766a8456(a094)이 뒤에 줄을 덧붙인 뒤로 착지 트리에 없었다(NOT-APPLIED).
     "\t\tReasonStrategyCentralIntegrity,\n\n", "\n"),
    ("E11 reason missing from the latch order", RT,
     "\tReasonStrategyCentralIntegrity,\n\t// Appended, per the rule above. The operating mode", "\t// Appended, per the rule above. The operating mode"),
]
SET_5621_TESTS = [
    ["go", "test", "-count=1", "-run",
     "TestARefreshOnly|TestAnOrdinaryRefreshOnlyCycleErrorDoesNotBlockEntry|TestWithoutAnEntryGate|TestTheProductionStrategySupervisor|TestTheCompleteCensus|TestTheOnlyWorkerProduction|TestTheFourEscalations",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestReasonCodeEnumIsStable|TestTheStrategyCentralIntegrityLatch", "./internal/execgw"],
]
HO = "internal/strategyhandoff/handoff.go"
DH = "internal/app/engine/strategy_dispatch_handoff.go"
FL = "internal/app/engine/strategy_account_first_leg_authority.go"
SET_5221 = [
    ("F01 activation gate removed (every market per scope)", DH,
     "\tif !authority.familyActivation().Verified() {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}",
     "\tif false {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}"),
    ("F02 activated market stays market-wide", DH,
     "\tif !authority.familyActivation().Verified() {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}",
     "\tif true {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}"),
    ("F03 duplicate owner scope not refused", HO,
     "if _, duplicate := seen[scope]; duplicate {", "if _, duplicate := seen[scope]; duplicate && false {"),
    ("F04 owner scope ignores position generation", HO,
     "\t\tgeneration: result.Lineage.PositionGeneration,", "\t\tgeneration: 0,"),
    ("F05 owner scope ignores symbol", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),", "\t\tsymbol:     \"\","),
    ("F06 symbol spelling not normalised", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),", "\t\tsymbol:     result.Lineage.Symbol,"),
    ("F07 account spelling not normalised", HO,
     "\t\taccount:    strings.TrimSpace(result.Lineage.AccountRef),", "\t\taccount:    result.Lineage.AccountRef,"),
    ("F08 closed market split into scopes", HO,
     "\tif !ready || len(selected) == 0 {\n\t\treturn []Handoff{Admit(ready, selected)}",
     "\tif len(selected) == 0 {\n\t\treturn []Handoff{Admit(ready, selected)}"),
    ("F09 delivery loop stops after the first scope", DH,
     "\tfor _, handoff := range handoffs {\n\t\tif err := handoff.Deliver(body); err != nil {",
     "\tfor _, handoff := range handoffs[:1] {\n\t\tif err := handoff.Deliver(body); err != nil {"),
    ("F10 delivery loop continues past a fault", DH,
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\treturn err\n\t\t}",
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\t_ = err\n\t\t}"),
    ("F11 production cycle reverts to the market-wide handoff", SUP,
     "return deliverEachStrategyHandoff(fresh.proposals.forMarket(market).dispatchHandoffs(),",
     "return deliverEachStrategyHandoff([]strategyhandoff.Handoff{fresh.proposals.forMarket(market).dispatchHandoff()},"),
    ("F12 activated branch mints through the market-wide door", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\treturn []strategyhandoff.Handoff{strategyhandoff.Admit(authority.snapshot.Ready, selected)}"),
    ("F13 activated branch ignores readiness", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\treturn strategyhandoff.AdmitEachOwnerScope(true, selected)"),
    ("F14 first-leg count guard removed (today-equivalence pin must see it)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("F15 a second site mints per-scope handoffs", DH,
     "\treturn strategyhandoff.Admit(authority.snapshot.Ready, selected)\n}",
     "\t_ = strategyhandoff.AdmitEachOwnerScope\n\treturn strategyhandoff.Admit(authority.snapshot.Ready, selected)\n}"),
]
SET_5221_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestOnlyAnActivatedMarket|TestTwoOwnerScopesStill|TestEveryAdmittedOwnerScope|TestAScopeFault|TestTheProductionCycleDelivers"
     "|TestExactlyOneProduction|TestAdmitCensus|Handoff|Seam",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyhandoff"],
]
MD = "internal/app/engine/strategy_market_handoff_delivery.go"
RO = "internal/strategyrouter/router.go"
# 2026-09-30 리뷰 수리 로트(codex · 보이스 A/B/C). G = 잡혀야 하는 변이, N = 동작이 같은 리팩터(초록이어야 함 — 거짓 양성 대조).
SET_5221_FIX = [
    ("G01 owner scope keyed by horizon (B X03)", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),",
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)) + \"|\" + string(result.Lineage.Horizon),"),
    ("G02 owner scope keyed by lane (B X04)", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),",
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)) + \"|\" + result.Lineage.LaneID,"),
    ("G03 owner scope ignores account (codex P2-4)", HO,
     "\t\taccount:    strings.TrimSpace(result.Lineage.AccountRef),", "\t\taccount:    \"\","),
    ("G04 owner scope ignores market (codex P2-4 · B X02)", HO,
     "\t\tmarket:     string(result.Lineage.Market),", "\t\tmarket:     \"\","),
    ("G05 router OwnerKey normalises account differently (C #4a)", RO,
     "\t\tAccountRef:         strings.TrimSpace(account),", "\t\tAccountRef:         strings.ToUpper(strings.TrimSpace(account)),"),
    ("G06 production body drops the second scope onward (B X09)", MD,
     "\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\tlineage := delivered.Result().Lineage",
     "\tcalls := 0\n\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\tcalls++\n\t\tif calls > 1 {\n\t\t\treturn nil\n\t\t}\n\t\tlineage := delivered.Result().Lineage"),
    ("G07 cycle delivers through a same-named method of another type (B X08)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "a112MarketWideOnly{fresh.proposals.forMarket(market)}.dispatchHandoffs())\n}\n\ntype a112MarketWideOnly struct{ a strategyProposalMarketAuthority }\n\nfunc (m a112MarketWideOnly) dispatchHandoffs() []strategyhandoff.Handoff {\n\treturn []strategyhandoff.Handoff{m.a.dispatchHandoff()}")),
    ("G08 local shadow of the delivery function (B X10)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "\tdispatchStrategyMarketHandoffs := func(ctx context.Context, _ strategyCampaignCASReader, d strategyHandoffDispatcher, hs []strategyhandoff.Handoff) error {\n\t\treturn nil\n\t}\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())")),
    ("N01 same-behaviour refactor: handoffs through a local variable (B X11)", SUP,
     "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())",
     "\thandoffs := fresh.proposals.forMarket(market).dispatchHandoffs()\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, handoffs)"),
    ("G09 cycle passes the market-wide handoff (old F11)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "[]strategyhandoff.Handoff{fresh.proposals.forMarket(market).dispatchHandoff()})")),
    ("G10 B2 always refuses — only the control half should fail (codex P2-3 · F16)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif len(proposal.entries) >= 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("G11 second minting door via a public var's private type (codex P1-2)", HO,
     "var ErrNoDelivery = errors.New(\"strategyhandoff: delivery body is nil\")",
     "var ErrNoDelivery = deliveryError{errors.New(\"strategyhandoff: delivery body is nil\")}\n\ntype deliveryError struct{ error }\n\nfunc (deliveryError) Mint(r strategyflow.Result) Handoff {\n\treturn Handoff{selected: []strategyflow.Result{r}, pending: 1}\n}"),
    ("G12 engine admits each element separately (C #2 experiment C2)", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\tadmit := strategyhandoff.AdmitEachOwnerScope\n\tif !authority.snapshot.Ready || len(selected) == 0 {\n\t\treturn admit(authority.snapshot.Ready, selected)\n\t}\n\tvar out []strategyhandoff.Handoff\n\tfor _, one := range selected {\n\t\tout = append(out, admit(true, []strategyflow.Result{one})...)\n\t}\n\treturn out"),
    ("G13 a fourth Single() site with the new spelling (C #3 experiment D)", DH,
     "func deliverEachStrategyHandoff(",
     "func a112Peek(authority strategyProposalMarketAuthority) bool {\n\t_, ok := authority.dispatchHandoffs()[0].Single()\n\treturn ok\n}\n\nfunc deliverEachStrategyHandoff("),
    ("G14 Single() answer discarded via an intermediate binding (C #3 experiment D')", DH,
     "func deliverEachStrategyHandoff(",
     "func a112Peek(authority strategyProposalMarketAuthority) strategyflow.Result {\n\th := authority.dispatchHandoff()\n\tr, _ := h.Single()\n\treturn r\n}\n\nfunc deliverEachStrategyHandoff("),
    ("G15 new exported door the engine uses (C #1 experiment B shape)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "// AdmitVar 는 새 문.\nvar AdmitVar = AdmitEachOwnerScope\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("G16 delivery keeps going after a fault (old F10, now on the production body)", DH,
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\treturn err\n\t\t}",
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\t_ = err\n\t\t}"),
    ("G18 B2 count condition removed — both orders must fail (old F14; fixture order by orders, coordinator order by the message)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("G17 claimed scope stops the cycle instead of being skipped", MD,
     "\t\tif cas.Claimed || cas.State != \"FLAT\" && cas.State != \"CLOSED\" {\n\t\t\treturn nil\n\t\t}",
     "\t\tif cas.Claimed || cas.State != \"FLAT\" && cas.State != \"CLOSED\" {\n\t\t\treturn errors.New(\"claimed\")\n\t\t}"),
]
SET_5221_FIX_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "Handoff|Seam|Admit|OwnerScope|Delivery|Delivers|Dispatch|Classified|TwoOwner|SameOwner|Skipped|FaultInOne|Discards|OnlyAnActivated|EveryAdmitted|AScopeFault",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run",
     "Handoff|Seam|Admit|Delivery|Delivers|Dispatch|Classified|Skipped|FaultInOne|Discards",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyhandoff"],
]
SETS = {"5.6.2.1": (SET_5621, SET_5621_TESTS), "5.2.2.1": (SET_5221, SET_5221_TESTS),
        "5.2.2.1-fix": (SET_5221_FIX, SET_5221_FIX_TESTS)}
MUTANTS, TESTS = SET_5621, SET_5621_TESTS

def run_tests(copy: Path, env: dict) -> tuple[str, str]:
    failed, notes, build = [], [], False
    for command in TESTS:
        result = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if result.returncode != 0:
            out = result.stdout + result.stderr
            if "[build failed]" in out or "[setup failed]" in out:
                build = True
                notes.append(next((l for l in out.splitlines() if ".go:" in l), "build failed")[:200])
                continue
            names = [line.strip()[len("--- FAIL: "):].split(" ")[0] for line in out.splitlines() if line.strip().startswith("--- FAIL: ")]
            leaves = [n for n in names if not any(o != n and o.startswith(n + "/") for o in names)]
            failed.extend(leaves)
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    why = f"{len(failed)} failing: " + ", ".join(failed[:8]) + (" …" if len(failed) > 8 else "") + (" | " + " | ".join(notes) if notes else "")
    if build:
        return "BUILD-FAIL", why
    if failed or notes:
        return "RED", why
    return "GREEN", ""


def main() -> None:
    args = sys.argv[1:]
    only = None
    if "--only" in args:
        i = args.index("--only")
        only = re.compile(args[i + 1])
        args = args[:i] + args[i + 2:]
    global MUTANTS, TESTS
    if "--set" in args:
        i = args.index("--set")
        MUTANTS, TESTS = SETS[args[i + 1]]
        args = args[:i] + args[i + 2:]
    scratch, own = Path(args[0]), args[1:]
    copy = scratch / f"mut-a112-lot-{os.getpid()}"
    copy.mkdir(parents=True)
    # 사본은 HEAD 커밋 트리 + 이 로트가 넘긴 파일(own)뿐이다 — 병행 세션의 미커밋 편집이 사본에 섞이면 대조군부터
    # 깨지고(2026-09-30 실측: 남의 미커밋 journal 편집이 남의 미추적 파일을 참조), 섞인 채 GREEN 이면 무엇을 쟀는지 모른다.
    archive = subprocess.run(["git", "archive", "HEAD", "go.mod", "go.sum", "internal", "cmd", "tools"], cwd=ROOT,
                             capture_output=True, check=True).stdout
    subprocess.run(["tar", "-x", "-C", str(copy)], input=archive, check=True)
    for rel in own:
        source = ROOT / rel
        (copy / rel).parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, copy / rel)
    env = dict(os.environ, GOFLAGS="-trimpath")
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    dirty = []
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ledger.write(f"TREE\tHEAD {head} (git archive) + own: {','.join(own) or 'none'}\n")
    verdict, why = run_tests(copy, env)
    ledger.write(f"CONTROL\t{verdict}\t{why}\n")
    ledger.flush()
    if verdict != "GREEN":
        print("control not GREEN — stop:", verdict, why)
        sys.exit(2)
    for ident, rel, old, new in MUTANTS:
        if only and not only.search(ident):
            continue
        target = copy / rel
        pristine = target.read_text(encoding="utf-8")
        # 한 변이가 같은 파일의 여러 자리를 바꿀 수 있다(old · new 가 튜플이면 짝지어 차례로) — 각 앵커는 정확히 한 번.
        pairs = list(zip(old, new)) if isinstance(old, tuple) else [(old, new)]
        mutated, missing = pristine, [a for a, _ in pairs if pristine.count(a) != 1]
        if missing:
            ledger.write(f"{ident}\tNOT-APPLIED\tanchor count != 1: {missing[0][:60]!r}\n")
            ledger.flush()
            continue
        for a, b in pairs:
            mutated = mutated.replace(a, b, 1)
        target.write_text(mutated, encoding="utf-8")
        verdict, why = run_tests(copy, env)
        label = {"RED": "CAUGHT", "GREEN": "SURVIVED"}.get(verdict, verdict)
        if ident.startswith("N"):  # 동작이 같은 리팩터 — 초록이어야 한다(빨가면 거짓 양성)
            label = {"GREEN": "GREEN-AS-EXPECTED", "RED": "FALSE-POSITIVE"}.get(verdict, verdict)
        ledger.write(f"{ident}\t{label}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
