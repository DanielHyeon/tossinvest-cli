#!/usr/bin/env python3
"""task 7.5.14 · 7.5.18 · 6.4(h) — **시험 파일 쪽** 변이와 "그 시험 단독" 측정.

`75_mut.py` 는 `check_analysis.py` 만 변이한다. 이 로트의 가드 일부는 시험 파일에 산다(git 설정 고정 · 구조 시험의 원시 집합 ·
자리 수 대조) 그리고 7.5.18 은 "그 시험 **하나**가 반증 변이를 잡는가" 를 묻는다. 사본(pid 별 · `A122_HARNESS_WORK`)에서:

1. 7.5.18 — 생산 변이 MV1(호출부가 `ast.json` 을 다시 읽음) 아래에서 `test_the_bundle_text_is_built_from_the_bytes_the_command_read`
   **하나만** 돌린다: 기준 리비전의 시험 판(옛 픽스처) · 워킹트리의 시험 판(새 픽스처). 무변이 대조도 같이.
2. 시험 쪽 변이 — 변이마다 같은 시험 묶음으로 무변이 대조군 → 변이 → 원복 확인. 변이가 생산 코드도 같이 바꿀 수 있다(TM6+MU2).

**판정 규칙 (보수 — 독립 적대 리뷰 6(iii)).** `-v` 로 돌려 **기대한 시험 이름이 실제로 돌았는지** 본다. 첫 판은 rc 만 봐서, 없는 시험
이름(`_FailedTest`)으로도 `errors=1` 을 CAUGHT 로 적었다. 기대 시험이 대조군에서 안 돌았거나 변이 판이 시험을 못 돌렸으면 "못 쟀다" 다.

    A122_HARNESS_WORK=<ext4> python3 7514_test_mut.py [<before-rev>]     # 기본 3bb11b2d
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
SP = Path(os.environ.get("A122_HARNESS_WORK") or Path(__file__).resolve().parent / "_work")
WORK = SP / f"7514_test_mut.{os.getpid()}"
ENV = {**os.environ, "GOFLAGS": "-trimpath", "GOCACHE": str(SP / "gocache")}
C = "test_check_analysis"
R = f"{C}.TheRecheckReadsWhatTheVerdictRead"
G = f"{C}.TheSuiteDoesNotReadTheDevelopersGitConfig"
BUNDLE_TEST = f"{R}.test_the_bundle_text_is_built_from_the_bytes_the_command_read"
MV1 = ('                target, evidence.held.get(target / "ast.json"), names=names)',
       '                target, _read_regular(target / "ast.json"), names=names)')
MU2 = ('    if _kind(landing_file):\n', '    if landing_file.exists():\n')
TM6 = ('(isinstance(item.func, ast.Attribute) and item.func.attr in cls.PRIMITIVES)',
       '(isinstance(item.func, ast.Attribute) and item.func.attr in cls.PRIMITIVES - {"exists"})')
STRUCTURE = [f"{R}.test_no_read_primitive_lives_outside_the_funnels",
             f"{R}.test_the_census_sees_every_primitive_and_its_aliases",
             f"{R}.test_a_second_site_of_an_exempt_form_is_not_exempt",
             f"{R}.test_a_second_site_of_a_funnel_form_is_counted"]

# 이름 → (시험 파일 편집들, 생산 편집들, 돌릴 시험 이름들)
TEST_MUTATIONS = {
    "TM1_the_global_scope_is_not_pinned": (
        [('    environment["GIT_CONFIG_GLOBAL"] = os.devnull\n', "")], [], [G], "fixture_git_env.py"),
    "TM2_the_system_scope_is_not_pinned": (
        [('    environment["GIT_CONFIG_SYSTEM"] = os.devnull\n', "")], [], [G], "fixture_git_env.py"),
    "TM2b_config_variables_are_kept": (
        [("        if key in _FROM_THE_ENVIRONMENT or key.startswith(_NUMBERED):", "        if False:")], [], [G],
        "fixture_git_env.py"),
    "TM2c_the_baseline_suite_does_not_pin_itself": (
        [("fixture_git_env.isolate()\n", "")], [], [f"{G}.test_the_execution_baseline_suite_pins_itself"],
        "test_execution_baseline.py"),
    "TM3_the_census_drops_exists": (
        [('"stat", "lstat", "statvfs", "exists", "lexists",', '"stat", "lstat", "statvfs", "lexists",')], [], STRUCTURE,
        "test_check_analysis.py"),
    "TM4_the_exemption_count_is_not_checked": (
        [("                     for key, (expected, _) in cls.EXEMPT.items() if counts.get(key, 0) != expected]",
          "                     for key, (expected, _) in cls.EXEMPT.items() if False]")], [], STRUCTURE,
        "test_check_analysis.py"),
    "TM5_the_funnel_count_is_not_checked": (
        [("                     for key, expected in cls.FUNNELS.items() if counts.get(key, 0) != expected]",
          "                     for key, expected in cls.FUNNELS.items() if False]")], [], STRUCTURE,
        "test_check_analysis.py"),
    "TM6_the_census_skips_exists": ([TM6], [], STRUCTURE, "test_check_analysis.py"),
    "TM6+MU2_the_census_skips_exists_and_Y13_returns": ([TM6], [MU2], STRUCTURE, "test_check_analysis.py"),
    "TM7_aliases_are_not_followed": (
        [("        aliases = set(cls.NAME_PRIMITIVES)\n        while True:", "        aliases = set(cls.NAME_PRIMITIVES)\n        while False:")],
        [], STRUCTURE, "test_check_analysis.py"),
}


def copy(test_source: bytes | None = None) -> Path:
    if WORK.exists():
        shutil.rmtree(WORK)
    shutil.copytree(REPO / "tools" / "logic-map", WORK / "logic-map", ignore=shutil.ignore_patterns("__pycache__"))
    (WORK / "sdd").mkdir()
    shutil.copy(REPO / "tools" / "sdd" / "sdd_doctor.py", WORK / "sdd")
    shutil.copy(REPO / "tools" / "gate.sh", WORK / "gate.sh")
    if test_source is not None:
        (WORK / "logic-map" / "test_check_analysis.py").write_bytes(test_source)
    return WORK / "logic-map"


def run(where: Path, names: list[str]) -> tuple[int, set[str], str]:
    """(rc, 실제로 돈 시험 이름들, 꼬리). `_FailedTest`(없는 이름 · 수집 실패)는 돈 것으로 안 센다."""
    process = subprocess.run([sys.executable, "-m", "unittest", "-v", *names], cwd=where, capture_output=True, text=True,
                             timeout=3600, env=ENV)
    # 3.12 의 `-v` 줄: `test_x (모듈.클래스.test_x)` — 괄호 안이 전체 이름이다.
    ran = {match.group(2) for match in re.finditer(r"^(\w+) \(([\w.]+)\)\s", process.stderr, re.MULTILINE)
           if "_FailedTest" not in match.group(2) and "loader" not in match.group(2)}
    tail = " ".join(line for line in process.stderr.splitlines() if line.startswith(("Ran ", "OK", "FAILED")))
    return process.returncode, ran, tail


def covers(ran: set[str], names: list[str]) -> bool:
    """기대 이름마다 돈 시험이 하나 이상 있다(클래스 이름이면 그 클래스의 시험 하나 이상)."""
    return all(any(test == name or test.startswith(name + ".") for test in ran) for name in names)


def main() -> None:
    before = sys.argv[1] if len(sys.argv) > 1 else "3bb11b2d"
    old_test = subprocess.run(["git", "show", f"{before}:tools/logic-map/test_check_analysis.py"], cwd=REPO,
                              capture_output=True, check=True).stdout
    print("--- 7.5.18: MV1 아래 그 시험 하나")
    for label, source in ((f"{before} 시험 판", old_test), ("워킹트리 시험 판", None)):
        where = copy(source)
        target = where / "check_analysis.py"
        pristine = target.read_text(encoding="utf-8")
        assert pristine.count(MV1[0]) == 1
        rc, ran, tail = run(where, [BUNDLE_TEST])
        print(f"  {label} · 무변이: rc={rc} 돈 시험 {len(ran)} · {tail}")
        target.write_text(pristine.replace(*MV1), encoding="utf-8")
        rc, ran, tail = run(where, [BUNDLE_TEST])
        print(f"  {label} · MV1:   rc={rc} 돈 시험 {len(ran)} · {tail}")
    print("--- 시험 쪽 변이")
    where = copy()
    production = where / "check_analysis.py"
    production_pristine = production.read_text(encoding="utf-8")
    for name, (edits, production_edits, names, filename) in TEST_MUTATIONS.items():
        target = where / filename
        pristine = target.read_text(encoding="utf-8")
        control_rc, control_ran, control_tail = run(where, names)
        if control_rc or not covers(control_ran, names):
            print(f"  {name:48s} **못 쟀다** — 대조군 rc={control_rc} · {control_tail}")
            continue
        text = pristine
        for old, new in edits:
            assert text.count(old) == 1, f"{name}: 앵커 {old[:50]!r} 가 {text.count(old)}회"
            text = text.replace(old, new)
        produced = production_pristine
        for old, new in production_edits:
            assert produced.count(old) == 1, f"{name}: 생산 앵커 {old[:50]!r} 가 {produced.count(old)}회"
            produced = produced.replace(old, new)
        target.write_text(text, encoding="utf-8")
        production.write_text(produced, encoding="utf-8")
        rc, ran, tail = run(where, names)
        target.write_text(pristine, encoding="utf-8")
        production.write_text(production_pristine, encoding="utf-8")
        assert target.read_text(encoding="utf-8") == pristine and production.read_text(encoding="utf-8") == production_pristine
        verdict = "SURVIVED" if not rc else "CAUGHT" if covers(ran, names) else "**못 쟀다**"
        print(f"  {name:48s} 대조군 {control_tail} · 변이 {verdict} {tail}")
    shutil.rmtree(WORK, ignore_errors=True)


if __name__ == "__main__":
    main()
