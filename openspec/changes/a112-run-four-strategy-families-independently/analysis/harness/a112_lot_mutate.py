#!/usr/bin/env python3
"""a112 로트 5.6.2 · 5.2.2 변이 하네스 — a092 `mutate_unit2.py` 의 사본 방식(저장소 추적 파일을 pid 붙은 사본으로 복사 · 무변이
대조군이 GREEN 이 아니면 멈춤 · 변이는 한 번에 하나, 사본에서만 · 판정 CAUGHT/SURVIVED/BUILD-FAIL).

사용: python3 a112_lot_mutate.py --set 5.6.2.1 <scratch-dir> [미추적 새 파일 경로 …]
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
     "\t\tReasonStrategyCentralIntegrity,\n\t}", "\t}"),
    ("E11 reason missing from the latch order", RT,
     "\tReasonStrategyCentralIntegrity,\n\t// Appended, per the rule above. The operating mode", "\t// Appended, per the rule above. The operating mode"),
]
SET_5621_TESTS = [
    ["go", "test", "-count=1", "-run",
     "TestARefreshOnly|TestAnOrdinaryRefreshOnlyCycleErrorDoesNotBlockEntry|TestWithoutAnEntryGate|TestTheProductionStrategySupervisor|TestTheCompleteCensus|TestTheOnlyWorkerProduction|TestTheFourEscalations",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestReasonCodeEnumIsStable|TestTheStrategyCentralIntegrityLatch", "./internal/execgw"],
]
SETS = {"5.6.2.1": (SET_5621, SET_5621_TESTS)}
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
        if pristine.count(old) != 1:
            ledger.write(f"{ident}\tNOT-APPLIED\told occurs {pristine.count(old)} times\n")
            ledger.flush()
            continue
        target.write_text(pristine.replace(old, new, 1), encoding="utf-8")
        verdict, why = run_tests(copy, env)
        label = {"RED": "CAUGHT", "GREEN": "SURVIVED"}.get(verdict, verdict)
        ledger.write(f"{ident}\t{label}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
