#!/usr/bin/env python3
"""task 7.5.15 변이 — `execution_baseline.py` 의 시한 다섯 자리마다 **실제로** 빨개지는 시험이 있는가.

`75_mut.py` 는 `check_analysis.py` 만 변이한다. 여기는 같은 규율(사본 대상 · pid 별 사본 · 무변이 대조군 창 양끝 ·
스위트 전체 · 원복 확인)을 그대로 빌려 **다른 파일**을 겨눈다 — 헬퍼는 `75_mut.py` 에서 불러온다.

    A122_HARNESS_WORK=<ext4 스크래치> python3 7515_mut.py
"""
import importlib.util
import shutil
from pathlib import Path

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("mut75", HERE / "75_mut.py")
mut = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mut)

MUTATIONS = {
    "EB1_the_git_helper_has_no_timeout": [(
        'capture_output=True, text=text, timeout=120, check=False)', 'capture_output=True, text=text, check=False)')],
    "EB2_ancestry_has_no_timeout": [(
        '    result = subprocess.run(["git", "merge-base", "--is-ancestor", older, newer], cwd=root, capture_output=True,\n'
        '                            timeout=10, check=False)',
        '    result = subprocess.run(["git", "merge-base", "--is-ancestor", older, newer], cwd=root, capture_output=True,\n'
        '                            check=False)')],
    "EB3_go_list_has_no_timeout": [(
        '    process = subprocess.run(command, cwd=root, capture_output=True, text=True, timeout=60, check=False)',
        '    process = subprocess.run(command, cwd=root, capture_output=True, text=True, check=False)')],
    "EB4_symbolic_ref_has_no_timeout": [(
        'capture_output=True, timeout=10,\n                      check=False).returncode == 0:',
        'capture_output=True,\n                      check=False).returncode == 0:')],
    "EB5_the_clean_tree_checks_have_no_timeout": [(
        'if subprocess.run(["git", *args], cwd=root, capture_output=True, timeout=30, check=False).returncode:',
        'if subprocess.run(["git", *args], cwd=root, capture_output=True, check=False).returncode:')],
}


def main() -> None:
    mut.setup()
    target = mut.WORK / "logic-map" / "execution_baseline.py"
    pristine = target.read_text(encoding="utf-8")
    origin = (mut.REPO / "tools" / "logic-map" / "execution_baseline.py").read_text(encoding="utf-8")
    assert pristine == origin, "사본이 원본과 다르다"
    code, output = mut.run(mut.SUITE)
    if code:
        print("STOP — 무변이 대조군이 빨갛다\n", output[-3000:])
        raise SystemExit(1)
    control_ran = mut.ran_count(output)
    print(f"control GREEN {output.strip().splitlines()[-1]} (Ran {control_ran})")
    survived = []
    for name, edits in MUTATIONS.items():
        text = pristine
        for old, new in edits:
            assert text.count(old) == 1, f"{name}: 앵커 {old[:60]!r} 가 {text.count(old)}회"
            text = text.replace(old, new)
        target.write_text(text, encoding="utf-8")
        code, output = mut.run(mut.SUITE)
        target.write_text(pristine, encoding="utf-8")
        assert target.read_text(encoding="utf-8") == pristine, "원복이 안 됐다"
        names = mut.failing(output)
        ran = mut.ran_count(output)
        suspect = "" if ran == control_ran else f"  · **환경 의심** (Ran {ran} ≠ {control_ran})"
        verdict = "CAUGHT" if code else "SURVIVED"
        if not code:
            survived.append(name)
        print(f"{name:44s} {verdict:9s}{suspect} {len(names):2d} {', '.join(n.split('.')[-1] for n in names[:3])}")
    code, output = mut.run(mut.SUITE)
    if code:
        print("\nSTOP — 창 **끝** 무변이 대조군이 빨갛다. 이 창의 결과는 전부 버린다.\n", output[-3000:])
        raise SystemExit(1)
    print(f"창 끝 control GREEN {output.strip().splitlines()[-1]} (Ran {mut.ran_count(output)})")
    print(f"\nSURVIVED {len(survived)}/{len(MUTATIONS)}" + (f": {survived}" if survived else ""))
    shutil.rmtree(mut.WORK, ignore_errors=True)


if __name__ == "__main__":
    main()
