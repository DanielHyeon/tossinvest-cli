#!/usr/bin/env python3
"""a125 1.2 — test_check_analysis.py 의 특례 시험을 반전한다. 인자: 대상 test_check_analysis.py 경로. AST 로 자리를 찾는다."""
import ast
import sys
from pathlib import Path

path = Path(sys.argv[1])
source = path.read_text(encoding="utf-8")
# `splitlines` 는 U+2028 · U+2029 · U+0085 에서도 자른다 — 이 파일에 그 글자가 있어 AST 줄 번호와 어긋난다. `\n` 으로만 자른다.
lines = [line + "\n" for line in source.split("\n")]
lines[-1] = lines[-1][:-1]
tree = ast.parse(source)
reversed_block = (Path(__file__).parent / "reversed_tests.py").read_text(encoding="utf-8")

DROP = {
    "test_main_distinguishes_ordinary_adoption_and_invalid_results",
    "test_main_real_adoption_prints_exception_label_only_after_validation",
    "_adoption_with_complete_bundle",
    "test_real_adoption_without_local_venv_accepts_external_doctor_probe",
    "test_valid_adoption_uses_e_and_requires_complete_current_bundle",
    "test_real_adoption_deleted_function_requires_base_revision",
    "test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back",
    "test_real_adoption_stale_current_ast_hash_fails",
    "test_real_adoption_rejects_local_maps_with_function_logic_reference",
    "_adoption_output",
    "test_the_adoption_window_ends_at_the_audited_source_commit",
    "test_the_adoption_path_does_not_advise_a_record_it_would_refuse",
    "test_a_landing_record_is_refused_in_the_adoption_path",
    "test_the_recorder_refuses_the_adoption_path_too",
    "_archive_adoption",
    "test_an_archived_adoption_is_rechecked_by_its_id",
    "test_a_copied_adoption_record_does_not_make_another_change_a063",
    "test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised",
    "test_the_execution_baseline_suite_pins_itself",
}
cuts: list[tuple[int, int]] = []          # 0-based [start, end)
insert_at = None
for node in tree.body:
    if isinstance(node, ast.ClassDef):
        if node.name == "CheckAnalysisTests":
            insert_at = node.lineno - 1 - len(node.decorator_list)
        for item in node.body:
            if isinstance(item, ast.FunctionDef) and item.name in DROP:
                start = min([item.lineno] + [d.lineno for d in item.decorator_list]) - 1
                cuts.append((start, item.end_lineno))
found = {lines[s].strip() for s, _ in cuts}
assert len(cuts) == len(DROP), (len(cuts), len(DROP))
out = []
skip = set()
for start, end in cuts:
    skip.update(range(start, end))
for index, line in enumerate(lines):
    if index == insert_at:
        out.append(reversed_block.lstrip("\n") + "\n\n")
    if index in skip:
        continue
    out.append(line)
text = "".join(out)


def swap(old: str, new: str) -> None:
    global text
    assert text.count(old) == 1, old[:80]
    text = text.replace(old, new)


swap("import execution_baseline as adoption\n", "")
swap('''        case = CheckAnalysisTests("test_valid_adoption_uses_e_and_requires_complete_current_bundle")
        raw, root, p, e = case._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])     # 대조군
            good = root / "openspec" / "changes" / adoption.CHANGE / "analysis" / "function-logic" / "internal-soak--attest"''',
     '''        # a125: 이관 경로가 없어졌다 — 같은 모양의 a063 픽스처를 일반 경로에서 잰다(반전 D3).
        raw, root, _, _ = _a063_fixture()
        with raw:
            self.assertEqual(check_analysis.check(A063, root), [])     # 대조군
            good = root / "openspec" / "changes" / A063 / "analysis" / "function-logic" / "internal-soak--attest"''')
swap('''            case._commit(root, "an escaping bundle")
            subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)''',
     '''            _recommit_detached(root, "an escaping bundle")
            errors = check_analysis.check(A063, root)''')
swap('            self.assertFalse(facts["execution_baseline_adoption"])\n',
     '            self.assertNotIn("execution_baseline_adoption", facts)      # a125: 옛 키는 채워지지도 남지도 않는다\n')
swap("        for module in (check_analysis, adoption):\n",
     "        # a125(freeze F9): 모듈 목록은 손으로 고른 것이 아니라 디렉터리의 `subprocess` 를 쓰는 비시험 모듈 **전부**와 같아야 한다.\n"
     "        spawning = sorted(path.name for path in Path(check_analysis.__file__).resolve().parent.glob(\"*.py\")\n"
     "                          if not path.name.startswith(\"test_\") and \"subprocess.run(\" in path.read_text(encoding=\"utf-8\"))\n"
     "        # `risk_pattern_report.py` 는 번들 저작 도구다 — 판정이 import 하지 않는다(판정 모듈 목록 밖).\n"
     "        self.assertEqual(spawning, [\"check_analysis.py\", \"risk_pattern_report.py\"],\n"
     "                         \"새로 자식 프로세스를 띄우는 모듈은 판정 모듈인지 가려 아래 목록에 넣어야 한다\")\n"
     "        for module in (check_analysis,):\n")
swap('''            errors = check_analysis._verdict(
                root, "b" * 40, "l" * 40, False,
''', '''            errors = check_analysis._verdict(      # a125: `adopted` 인자가 없어졌다
                root, "b" * 40, "l" * 40,
''')
path.write_text(text, encoding="utf-8")
print("dropped", len(cuts), "inserted at", insert_at + 1)
