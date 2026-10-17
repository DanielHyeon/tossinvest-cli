#!/usr/bin/env python3
"""a125 3.2 — 변이. 사본(pid 붙은 디렉터리)에만 변이를 넣고, **무변이 대조군이 초록이 아니면 멈춘다**.

    python3 mutate.py --src <tools/logic-map> --work <작업 디렉터리> --expect-sha <check_analysis.py sha256>

변이마다 `TheA063ExceptionIsRetired` 와 관련 시험을 돌려 RED(CAUGHT)/GREEN(SURVIVED)을 적는다. 변이가 **닿았는지**는 치환
대상 문자열이 정확히 한 번 있었는지로 단언한다(안 닿은 변이를 CAUGHT 로 세지 않는다).
"""
import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("--src", required=True)
parser.add_argument("--work", required=True)
parser.add_argument("--expect-sha", required=True)
args = parser.parse_args()
src = Path(args.src).resolve()
assert hashlib.sha256((src / "check_analysis.py").read_bytes()).hexdigest() == args.expect_sha, "source moved"

A063 = "a063-align-attestation-renewal-profile"
TESTS = ["test_check_analysis.TheA063ExceptionIsRetired",
         "test_check_analysis.OneCommandJudgesOneHistoryAndOneRead.test_a_reused_context_carries_nothing_from_the_last_run",
         "test_check_analysis.EveryGateSubprocessHasATimeout"]

MUTANTS = {
    "M1 기록의 execution_base 를 base 로 읽기": (
        '    if context is not None:\n        context["effective_base"] = persisted\n',
        '    _record = change_dir / "execution-baseline.json"\n'
        '    if _record.is_file():\n'
        '        try:\n'
        '            persisted = json.loads(_record.read_text())["execution_base"]\n'
        '        except Exception:\n'
        '            pass\n'
        '    if context is not None:\n        context["effective_base"] = persisted\n'),
    "M2 a063 착지 기록 거절 복원(판정 경로)": (
        '        landing = resolve_landing(change_dir, root, base, head, evidence)\n',
        f'        if change == "{A063}" and _landing_record(change_dir, root, head) is not None:\n'
        '            return ["execution-baseline adoption does not accept a landing"], False\n'
        '        landing = resolve_landing(change_dir, root, base, head, evidence)\n'),
    "M3 a063 기록 명령 거절 복원(_recording_refusal)": (
        '    facts: dict[str, object] = {}\n    try:\n        base = resolve_base(change_dir, root, facts, change_id=change, head=head)\n',
        f'    if change == "{A063}":\n        return "execution-baseline adoption does not accept a landing", ""\n'
        '    facts: dict[str, object] = {}\n    try:\n        base = resolve_base(change_dir, root, facts, change_id=change, head=head)\n'),
    "M4 id 로 착지 우회(검증 없이 수락)": (
        '    candidate = _declared_landing(change_dir, root, head)\n',
        '    candidate = _declared_landing(change_dir, root, head)\n'
        f'    if candidate and change_dir.name.endswith("{A063}"):\n        return candidate\n'),
    "M5 비정규 기록 거절 복원": (
        '    if context is not None:\n        context["effective_base"] = persisted\n',
        '    _record = change_dir / "execution-baseline.json"\n'
        '    if os.path.lexists(_record) and not _record.is_file() or _record.is_symlink():\n'
        '        raise ValueError("execution-baseline record is not a regular file")\n'
        '    if context is not None:\n        context["effective_base"] = persisted\n'),
    "M6 SDD_BASE_REF 대조 끔": (
        '    if override and resolve(override) != persisted:\n',
        '    if override and resolve(override) != persisted and False:\n'),
    "M7 이관 라벨 복원(출력)": (
        '    print(f"[logic-map] {args.change}: evidence complete or diff-proven exempt")\n',
        f'    print(f"[logic-map] {{args.change}}: " + ("execution-baseline adoption exception evidence complete" if args.change == "{A063}" else "evidence complete or diff-proven exempt"))\n'),
    "M8 옛 문맥 키 되살림": (
        '    if context is not None:\n        context["effective_base"] = persisted\n',
        '    if context is not None:\n        context["effective_base"] = persisted\n        context["execution_baseline_adoption"] = False\n'),
}


def run(copy: Path) -> tuple[bool, str]:
    environment = {**os.environ, "GOFLAGS": "-trimpath", "GOCACHE": str(Path(args.work) / "gocache-mut")}
    process = subprocess.run([sys.executable, "-m", "unittest", *TESTS], cwd=copy, capture_output=True, text=True,
                             timeout=1800, env=environment)
    tail = [line for line in process.stderr.splitlines() if line.startswith(("FAIL:", "ERROR:", "Ran ", "OK", "FAILED"))]
    return process.returncode == 0, "\n".join(tail)


work = Path(args.work) / f"mut-{os.getpid()}"
results = {}
copy = work / "control" / "tools" / "logic-map"
shutil.copytree(src, copy, ignore=shutil.ignore_patterns("__pycache__"))
shutil.copytree(src.parent / "sdd", copy.parent / "sdd", ignore=shutil.ignore_patterns("__pycache__", ".venv"))
green, tail = run(copy)
results["control"] = {"green": green, "tail": tail}
print("control", "GREEN" if green else "RED", flush=True)
assert green, "무변이 대조군이 초록이 아니다 — 하네스 결함, 멈춘다\n" + tail
for name, (old, new) in MUTANTS.items():
    copy = work / f"m{list(MUTANTS).index(name) + 1}" / "tools" / "logic-map"
    shutil.copytree(src, copy, ignore=shutil.ignore_patterns("__pycache__"))
    shutil.copytree(src.parent / "sdd", copy.parent / "sdd", ignore=shutil.ignore_patterns("__pycache__", ".venv"))
    target = copy / "check_analysis.py"
    text = target.read_text(encoding="utf-8")
    assert text.count(old) == 1, (name, text.count(old))       # 닿았는가
    target.write_text(text.replace(old, new), encoding="utf-8")
    green, tail = run(copy)
    results[name] = {"verdict": "SURVIVED" if green else "CAUGHT", "tail": tail}
    print(name, results[name]["verdict"], flush=True)
    shutil.rmtree(copy.parent.parent)
shutil.rmtree(work / "control")
(Path(args.work) / f"mutate-{os.getpid()}.json").write_text(json.dumps(results, ensure_ascii=False, indent=1))
print("done", sum(r.get("verdict") == "CAUGHT" for r in results.values()), "/", len(MUTANTS))
