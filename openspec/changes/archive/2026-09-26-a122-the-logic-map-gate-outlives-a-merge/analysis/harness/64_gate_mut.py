#!/usr/bin/env python3
"""task 6.4(f) 변이 — `tools/gate.sh` 의 `resolve_change_dir` 를 **한쪽만** 고치면 무엇이 빨개지는가.

`75_mut.py` 는 `check_analysis.py` 만 변이한다. 여기는 shell 사본을 변이하고, 그 사본을 부르는 스위트 **둘**을
돈다: `test_check_analysis`(전체 — 두 해소기를 한 표로 도는 `TheTwoChangeResolversAgree` 가 여기 있다)와
`tools/sdd/test_gate_resolves_archived_changes`(shell 쪽 자기 시험). `gate.sh` 를 읽는 시험 파일은 저장소 전수로
셋이고(`rg -l 'gate\\.sh' -g 'test_*.py'` — ripgrep 14.1.0 실측; 첫 판이 적은 `--include` 는 rg 에 없는 플래그라 rc 2 다), 셋째 `test_race_detector_actually_runs` 는 게이트 **단계 목록**을
읽을 뿐 해소기를 안 부른다 — 그리고 `Makefile` · CI 파일을 읽어서 사본에서는 못 돈다.

규율은 `75_mut.py` 와 같다: 사본 대상 · 사본이 원본과 같은지 단언 · 무변이 대조군이 창 **앞뒤**로 GREEN ·
한 번에 한 판 · 판마다 `Ran N` 이 대조군과 같은지.

    python3 64_gate_mut.py
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
# 작업 자리는 기본 `_work/`(무시되는 디렉터리)다. `A122_HARNESS_WORK` 로 다른 파일시스템을 줄 수 있다 (task 6.4 보수) —
# 이 저장소는 /mnt/D(ntfs-3g) 위라 사본과 전용 GOCACHE 가 거기 있으면 스위트 한 판이 8~10 분, ext4 스크래치면 직접 실행과
# 비슷하다. 산출물의 뜻은 같다 — 사본 · 캐시 자리만 바뀐다.
SP = Path(os.environ.get("A122_HARNESS_WORK") or Path(__file__).resolve().parent / "_work")
SP.mkdir(parents=True, exist_ok=True)
WORK = SP / f"64_gate_mut_work.{os.getpid()}"
GO_ENV = {"GOFLAGS": "-trimpath", "GOCACHE": str(SP / "gocache")}

MUTATIONS = {
    # 날짜의 숫자 제한을 지운다 — 아무 글자 넷·둘·둘이 날짜가 된다.
    "G1_the_date_takes_any_character": [(
        "\t\t[0123456789][0123456789][0123456789][0123456789]-[0123456789][0123456789]-[0123456789][0123456789]-*) ;;",
        "\t\t????-??-??-*) ;;")],
    # 나열을 범위로 되돌린다(6.4 보수, P0-F1) — `en_US.UTF-8` 의 콜레이션이 9 없는 유니코드 날짜를 먹는다.
    "G5_the_date_is_a_locale_range": [(
        "\t\t[0123456789][0123456789][0123456789][0123456789]-[0123456789][0123456789]-[0123456789][0123456789]-*) ;;",
        "\t\t[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-*) ;;")],
    # 사본 둘을 하나로 친다.
    "G2_two_copies_pass": [(
        '\tif [ "$resolve_hits" -gt 1 ]; then', '\tif [ "$resolve_hits" -gt 2 ]; then')],
    # 활성 디렉터리를 안 본다.
    "G3_the_open_change_is_not_seen": [(
        '\tif [ -d "openspec/changes/$resolve_id" ]; then', '\tif false; then')],
    # 날짜를 벗긴 나머지가 아니라 **접미사**로 고른다(Python 쪽이 한때 한 실수).
    "G4_a_suffix_is_enough": [(
        '\t\t[ "${resolve_name#????-??-??-}" = "$resolve_id" ] || continue',
        '\t\tcase "$resolve_name" in *"$resolve_id") ;; *) continue ;; esac')],
}


def setup() -> Path:
    if WORK.exists():
        shutil.rmtree(WORK)
    # 저장소 모양을 그대로 둔다 — `test_gate_resolves_archived_changes` 는 `parents[2]` 를 루트로 쓰고,
    # `test_check_analysis` 는 `check_analysis.py` 옆 디렉터리의 `gate.sh` 를 쓴다.
    tools = WORK / "tools"
    shutil.copytree(REPO / "tools" / "logic-map", tools / "logic-map",
                    ignore=shutil.ignore_patterns("__pycache__"))
    (tools / "sdd").mkdir(parents=True)
    for name in ("sdd_doctor.py", "test_gate_resolves_archived_changes.py"):
        shutil.copy(REPO / "tools" / "sdd" / name, tools / "sdd")
    shutil.copy(REPO / "tools" / "gate.sh", tools / "gate.sh")
    return tools / "gate.sh"


def run() -> tuple[int, str, int]:
    """두 스위트를 돌고 `(rc, 출력, 돈 시험 수 합)` 을 낸다."""
    code, output, ran = 0, "", 0
    for cwd, module in ((WORK / "tools" / "logic-map", "test_check_analysis"),
                        (WORK / "tools" / "sdd", "test_gate_resolves_archived_changes")):
        process = subprocess.run([sys.executable, "-m", "unittest", module], cwd=cwd,
                                 capture_output=True, text=True, timeout=3600,
                                 env={**os.environ, **GO_ENV})
        code |= process.returncode
        output += process.stderr
        match = re.search(r"^Ran (\d+) tests", process.stderr, re.MULTILINE)
        ran += int(match.group(1)) if match else -10_000
    return code, output, ran


def failing(output: str) -> list[str]:
    return sorted({line.split(" ")[1] for line in output.splitlines()
                   if line.startswith(("FAIL: ", "ERROR: "))})


if __name__ == "__main__":
    target = setup()
    pristine = target.read_text(encoding="utf-8")
    assert pristine == (REPO / "tools" / "gate.sh").read_text(encoding="utf-8"), "사본이 원본과 다르다"
    code, output, control_ran = run()
    if code:
        print("STOP — 무변이 대조군이 빨갛다\n", output[-3000:])
        raise SystemExit(1)
    print(f"control GREEN (Ran {control_ran})")
    survived = []
    for name, edits in MUTATIONS.items():
        text = pristine
        for old, new in edits:
            assert text.count(old) == 1, f"{name}: 앵커 {old[:60]!r} 가 {text.count(old)}회"
            text = text.replace(old, new)
        target.write_text(text, encoding="utf-8")
        try:
            code, output, ran = run()
        finally:
            target.write_text(pristine, encoding="utf-8")
        names = failing(output)
        if not code:
            survived.append(name)
        suspect = "" if ran == control_ran else f"  · **환경 의심** (Ran {ran} ≠ {control_ran})"
        print(f"{name:36s} {'CAUGHT' if code else 'SURVIVED':9s}{suspect} {len(names):2d} "
              + ", ".join(n.split(".")[-1] for n in names))
    code, output, ran = run()
    if code:
        print("\nSTOP — 창 끝 무변이 대조군이 빨갛다. 결과를 버린다.\n", output[-3000:])
        raise SystemExit(1)
    # 창 끝 대조군도 **같은 수**를 돌아야 한다 (6.4 보수, P2) — 수가 다르면 그 창 안에서 환경이 바뀐 것이다.
    if ran != control_ran:
        print(f"\nSTOP — 창 끝 대조군이 Ran {ran} ≠ 시작 {control_ran}. 결과를 버린다.")
        raise SystemExit(1)
    print(f"창 끝 control GREEN (Ran {ran})")
    print(f"\nSURVIVED {len(survived)}/{len(MUTATIONS)}" + (f": {survived}" if survived else ""))
    shutil.rmtree(WORK, ignore_errors=True)
