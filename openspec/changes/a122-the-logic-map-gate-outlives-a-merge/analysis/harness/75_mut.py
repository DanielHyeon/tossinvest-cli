#!/usr/bin/env python3
"""task 7.5 변이 — 새 배치 fetch 의 갈래마다 **실제로** 빨개지는 시험이 있는가.

사본 대상 + **무변이 대조군**이 먼저다 ([[mutation-revert-needs-the-right-baseline]] ·
[[mutation-must-reach-the-thing-under-test]]). 대조군이 초록이 아니면 멈춘다.
"""
import shutil
import subprocess
import sys
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
WORK = SP / "75_mut_work"
SUITE = [
    "test_check_analysis.ABlobIsFetchedOncePerCommitNotOncePerBundle",
    "test_check_analysis.ALandingMustChangeWhatItsEvidencePins",
    "test_check_analysis.AComputedLandingIsAlwaysOneTheGateWillAccept",
    "test_check_analysis.TheRecordMustBeTheValueTheGateComputes",
    "test_check_analysis.TheLandingRuleLivesInOnePlace",
    "test_check_analysis.ADeclaredLandingMustBePinnedByEvidence",
    "test_check_analysis.ABorrowedWindowIsNeverNarrowed",
    "test_check_analysis.AFailingStepFiveSaysWhichWindowRequiredThem",
    "test_check_analysis.WorkAfterTheRecordIsNotOutsideTheWindow",
    "test_check_analysis.EvidenceFirstCommittedInsideAMergeIsAKnownLimit",
    "test_check_analysis.EvidenceAlreadyWrongInTheCommitThatHoldsItIsNamed",
    "test_check_analysis.TheRepairSignalIsMeasuredOnceAndNamesOldestFirst",
    "test_check_analysis.ARecordIsWrittenOnceAndNeverThroughASymlink",
    "test_check_analysis.AFaultInGitBecomesAVerdictNotATraceback",
]

MUTATIONS = {
    # --- 배치 프레이밍 ---
    "T1_output_not_nul_framed": [(
        '["git", "cat-file", "--batch", "-Z"],', '["git", "cat-file", "--batch", "-z"],')],
    "T2_line_framed_both_ways": [
        ('["git", "cat-file", "--batch", "-Z"],', '["git", "cat-file", "--batch"],'),
        ('f"{ref}:{relative}\\0".encode("utf-8")', 'f"{ref}:{relative}\\n".encode("utf-8")'),
        ('        end = data.find(b"\\0", position)', '        end = data.find(b"\\n", position)'),
    ],
    "T3_any_type_counts_as_content": [(
        '        if fields[1] == b"blob":', '        if True:')],
    "T4_non_blob_content_not_skipped": [(
        '        if fields[1] == b"blob":\n'
        '            found[relative] = data[position:position + size]\n'
        '        position += size + 1',
        '        if fields[1] == b"blob":\n'
        '            found[relative] = data[position:position + size]\n'
        '            position += size + 1')],
    "T5_failure_is_not_all_none": [(
        '    if process.returncode:\n        return found',
        '    if process.returncode:\n        pass')],
    "T6_short_response_not_stopped": [(
        '        if end < 0:\n            break', '        if end < 0:\n            end = len(data)')],
    # T6 가 살아남는 **이유**를 가설이 아니라 변이로 확인한다: 사전 채움이 `break` 와
    # `continue` 를 같게 만든다. 그 사전 채움이 못 박혀 있으면 T6 는 동등 변이다.
    "T18_no_prefill_of_unanswered_paths": [(
        '    found: dict[str, bytes | None] = {relative: None for relative in wanted}',
        '    found: dict[str, bytes | None] = {}')],
    "T7_size_from_content_scan": [(
        '        size = int(fields[2])', '        size = data.find(b"\\0", position) - position')],
    # --- 배치가 실제로 배치인가 ---
    "T8_pinning_at_fetches_one_by_one": [(
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles])\n'
        '    for _, source, digest in bundles:\n'
        '        blob = blobs[source]',
        '    for _, source, digest in bundles:\n'
        '        blob = _committed_bytes(root, candidate, source)')],
    "T9_unheld_fetches_one_by_one": [(
        '        committed = blobs[relative]\n'
        '        if committed is None and before:\n'
        '            committed = blobs[before]',
        '        committed = _committed_bytes(root, candidate, relative)\n'
        '        if committed is None and before:\n'
        '            committed = _committed_bytes(root, candidate, before)')],
    "T10_only_the_first_source_is_asked": [(
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles])',
        '    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles][:1])')],
    "T11_pre_archive_path_not_batched": [(
        'root, candidate, [path for _, relative, before in watched\n'
        '                          for path in ((relative, before) if before else (relative,))],',
        'root, candidate, [relative for _, relative, _ in watched],')],
    # 살아남은 둘이 서로를 덮는지 본다 ([[surviving-mutant-may-mean-accidental-safety]]).
    "T17_failure_and_short_response_both_open": [
        ('    if process.returncode:\n        return found',
         '    if process.returncode:\n        pass'),
        ('        if end < 0:\n            break',
         '        if end < 0:\n            end = len(data)'),
    ],
    # --- 한 번 재서 넘기는 목록이 **그 목록**인가 ---
    "T12_walk_passes_no_bundles": [(
        '        refusal, names = _landing_refusal(root, base, candidate, bundles, floor, repairs)',
        '        refusal, names = _landing_refusal(root, base, candidate, [], floor, repairs)')],
    "T13_declared_path_reads_a_second_list": [(
        '    bundles = _pinning_bundles(root, analysis)\n'
        '    refusal, _ = _landing_refusal(\n'
        '        root, base, candidate, bundles, _evidence_floor(root, bundles),',
        '    bundles = _pinning_bundles(root, analysis)\n'
        '    refusal, _ = _landing_refusal(\n'
        '        root, base, candidate, bundles, _evidence_floor(root, []),')],
    "T14_unheld_merged_guard_flipped": [(
        '        if committed is None and before:', '        if committed is None or before:')],
    # --- resolve 중복 제거가 같은 닻을 쓰는가 ---
    "T15_anchor_is_not_the_root": [(
        '    anchor = root.resolve()', '    anchor = (root / "openspec").resolve()')],
    "T16_committed_bytes_is_a_second_spelling": [(
        '    return _committed_many(root, ref, [relative])[relative]',
        '    process = subprocess.run(\n'
        '        ["git", "show", f"{ref}:{relative}"],\n'
        '        cwd=root, capture_output=True, timeout=30, check=False,\n'
        '    )\n'
        '    return None if process.returncode else process.stdout')],
}


def setup() -> Path:
    if WORK.exists():
        shutil.rmtree(WORK)
    shutil.copytree(REPO / "tools" / "logic-map", WORK / "logic-map",
                    ignore=shutil.ignore_patterns("__pycache__"))
    (WORK / "sdd").mkdir(parents=True)
    shutil.copy(REPO / "tools" / "sdd" / "sdd_doctor.py", WORK / "sdd")
    return WORK / "logic-map" / "check_analysis.py"


def run(names: list[str]) -> tuple[int, str]:
    process = subprocess.run([sys.executable, "-m", "unittest", *names],
                             cwd=WORK / "logic-map", capture_output=True, text=True, timeout=5400)
    return process.returncode, process.stderr


def failing(output: str) -> list[str]:
    return sorted({line.split(" ")[1] for line in output.splitlines()
                   if line.startswith(("FAIL: ", "ERROR: "))})


if __name__ == "__main__":
    target = setup()
    pristine = target.read_text(encoding="utf-8")
    assert "_committed_many" in pristine, "대상이 사본에 없다"
    code, output = run(SUITE)
    if code != 0:
        print("STOP — 무변이 대조군이 빨갛다\n", output[-3000:])
        raise SystemExit(1)
    print("control GREEN", output.strip().splitlines()[-1])
    survived = []
    for name, edits in MUTATIONS.items():
        text = pristine
        for old, new in edits:
            assert text.count(old) == 1, f"{name}: 앵커 {old[:60]!r} 가 {text.count(old)}회"
            text = text.replace(old, new)
        target.write_text(text, encoding="utf-8")
        code, output = run(SUITE)
        names = failing(output)
        verdict = "CAUGHT" if code else "SURVIVED"
        if not code:
            survived.append(name)
        print(f"{name:36s} {verdict:9s} {len(names):2d} "
              f"{', '.join(n.split('.')[-1] for n in names[:3])}"
              + (f" 외 {len(names) - 3}" if len(names) > 3 else ""))
        target.write_text(pristine, encoding="utf-8")
    print(f"\nSURVIVED {len(survived)}/{len(MUTATIONS)}" + (f": {survived}" if survived else ""))
