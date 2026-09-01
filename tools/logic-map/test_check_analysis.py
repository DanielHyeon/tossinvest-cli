from __future__ import annotations

import json
import hashlib
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import check_analysis


def write_bundle(
    root: Path,
    *,
    branches: list | None,
    logic_source_line: str = "- Source: `internal/sample.go`",
    branch_source_line: str = "",
    branch_rows: str = "| B1 | leaf | test | yes | yes |",
) -> Path:
    """Write a minimal four-file bundle whose only variable is what it claims."""
    source = root / "internal" / "sample.go"
    source.parent.mkdir(parents=True, exist_ok=True)
    source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
    target = (
        root / "openspec" / "changes" / "change" / "analysis" / "function-logic" / "pkg--run"
    )
    target.mkdir(parents=True, exist_ok=True)
    (target / "ast.json").write_text(
        json.dumps(
            {
                "file": "internal/sample.go",
                "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                "package": "sample",
                "function": "Run",
                "signature": "Run(params=0, results=0)",
                "start": {"line": 2, "column": 1},
                "end": {"line": 2, "column": 14},
                "branches": branches,
            }
        ),
        encoding="utf-8",
    )
    (target / "function-logic-map.md").write_text(
        f"""# Function Logic Map: `Run`
{logic_source_line}
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
        encoding="utf-8",
    )
    (target / "branch-test-map.md").write_text(
        f"# Branch Test Map: `Run`\n{branch_source_line}\n{branch_rows}\n",
        encoding="utf-8",
    )
    (target / "risk-pattern-report.md").write_text(
        "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
        encoding="utf-8",
    )
    return target


def run_check(root: Path) -> list[str]:
    with mock.patch(
        "check_analysis.resolve_base",
        return_value="base",
    ), mock.patch(
        "check_analysis.changed_existing_functions",
        return_value={},
    ):
        return check_analysis.check("change", root)


class BundleTextCoversEveryProseFileInTheBundle(unittest.TestCase):
    """강제 판정이 읽는 범위가 **열거된 파일 목록**이면 그 목록 밖으로 옮기면 꺼진다.

    8차 적대 리뷰가 `risk-pattern-report.md` 로 그것을 보였다 — 번들 필수 파일이고
    표는 머리글도 행도 그대로 살아 있는데 감사만 꺼졌다. 네 라운드 동안 "남는
    회피는 X 뿐"이라고 적은 문장이 매번 코드가 보는 범위보다 한 칸 넓었다."""

    def test_every_md_in_the_bundle_directory_is_read(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            bundle = Path(raw) / "b"
            bundle.mkdir()
            # 확장자로 거르면 `.txt`·`.mdx` 로 빠져나간다(9차 적대 리뷰).
            names = ("function-logic-map.md", "branch-test-map.md",
                     "risk-pattern-report.md", "someone-elses-note.md",
                     "notes.txt", "notes.mdx", "notes.markdown")
            for index, name in enumerate(names):
                (bundle / name).write_text(f"mark-{index}", encoding="utf-8")
            (bundle / "ast.json").write_text("{}", encoding="utf-8")
            text = check_analysis._bundle_text(bundle / "function-logic-map.md")
        for index in range(len(names)):
            self.assertIn(f"mark-{index}", text)


class CheckAnalysisTests(unittest.TestCase):
    def test_explicit_exemption_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            change = Path(tmp) / "openspec" / "changes" / "docs-only"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("docs-only", Path(tmp)), [])

    def test_placeholders_are_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            target = (
                Path(tmp)
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text("{}\n", encoding="utf-8")
            for name in check_analysis.REQUIRED[1:]:
                (target / name).write_text("TODO\n", encoding="utf-8")
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertGreater(len(check_analysis.check("change", Path(tmp))), 0)

    def test_completed_bundle_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = root / "internal" / "sample.go"
            source.parent.mkdir(parents=True)
            source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
            target = (
                root
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text(
                json.dumps(
                    {
                        "file": "internal/sample.go",
                        "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                        "package": "sample",
                        "function": "Run",
                        "signature": "Run(params=0, results=0)",
                        "start": {"line": 2, "column": 1},
                        "end": {"line": 2, "column": 14},
                        "branches": [],
                    }
                ),
                encoding="utf-8",
            )
            (target / "function-logic-map.md").write_text(
                """# Function Logic Map: `Run`
- Source: `internal/sample.go`
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
                encoding="utf-8",
            )
            (target / "branch-test-map.md").write_text(
                "# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n",
                encoding="utf-8",
            )
            (target / "risk-pattern-report.md").write_text(
                "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("change", root), [])

    def test_null_branches_from_the_go_extractor_are_accepted(self) -> None:
        # The Go extractor marshals a nil slice as JSON null, so a branchless
        # function arrives as "branches": null rather than []. The bundle must
        # validate instead of crashing on the None.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = root / "internal" / "sample.go"
            source.parent.mkdir(parents=True)
            source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
            target = (
                root
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text(
                json.dumps(
                    {
                        "file": "internal/sample.go",
                        "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                        "package": "sample",
                        "function": "Run",
                        "signature": "Run(params=0, results=0)",
                        "start": {"line": 2, "column": 1},
                        "end": {"line": 2, "column": 14},
                        "branches": None,
                    }
                ),
                encoding="utf-8",
            )
            (target / "function-logic-map.md").write_text(
                """# Function Logic Map: `Run`
- Source: `internal/sample.go`
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
                encoding="utf-8",
            )
            (target / "branch-test-map.md").write_text(
                "# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n",
                encoding="utf-8",
            )
            (target / "risk-pattern-report.md").write_text(
                "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("change", root), [])

    def test_modified_function_cannot_use_exemption(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            required = {
                ("internal/order.go", "Engine.Place"): {
                    "file": "internal/order.go",
                    "function": "Engine.Place",
                    "current_hash": "abc",
                }
            }
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value=required,
            ):
                errors = check_analysis.check("change", root)
            self.assertTrue(any("Engine.Place" in error for error in errors))

    def test_invalid_base_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                side_effect=ValueError("invalid base"),
            ):
                errors = check_analysis.check("change", root)
            self.assertTrue(any("invalid base" in error for error in errors))

    def test_duplicate_branch_ids_do_not_satisfy_ast_coverage(self) -> None:
        text = "| B1 | one |\n| B1 | duplicate |\n"
        self.assertEqual(check_analysis.branch_ids(text), ["B1", "B1"])

    def test_git_diff_failure_is_not_treated_as_empty_change(self) -> None:
        failed = subprocess.CompletedProcess([], 128, "", "bad revision")
        with mock.patch(
            "check_analysis.subprocess.run",
            return_value=failed,
        ):
            with self.assertRaises(RuntimeError):
                check_analysis.changed_existing_functions(Path("/tmp"), "bad")

    def test_new_function_in_existing_file_is_not_reported_as_modified_existing(self) -> None:
        diff = subprocess.CompletedProcess(
            [],
            0,
            "\n".join(
                (
                    "diff --git a/internal/x.go b/internal/x.go",
                    "--- a/internal/x.go",
                    "+++ b/internal/x.go",
                    "@@ -8,0 +9,3 @@",
                )
            ),
            "",
        )
        old_functions = [
            {
                "function": "Existing",
                "start": {"line": 2},
                "end": {"line": 5},
                "source_sha256": "old-file",
            }
        ]
        current_functions = old_functions + [
            {
                "function": "Added",
                "start": {"line": 9},
                "end": {"line": 11},
                "source_sha256": "current-file",
            }
        ]
        with tempfile.TemporaryDirectory() as tmp, tempfile.NamedTemporaryFile(suffix=".go") as base_source:
            root = Path(tmp)
            current = root / "internal" / "x.go"
            current.parent.mkdir(parents=True)
            current.write_text("package x\n", encoding="utf-8")
            with mock.patch(
                "check_analysis.subprocess.run",
                return_value=diff,
            ), mock.patch(
                "check_analysis.base_file",
                return_value=Path(base_source.name),
            ), mock.patch(
                "check_analysis.go_functions",
                side_effect=(old_functions, current_functions),
            ):
                required = check_analysis.changed_existing_functions(root, "base")

        self.assertNotIn(("internal/x.go", "Added"), required)

    def test_base_file_load_failure_is_not_treated_as_new_file(self) -> None:
        diff = subprocess.CompletedProcess(
            [],
            0,
            "\n".join(
                (
                    "diff --git a/internal/x.go b/internal/x.go",
                    "--- a/internal/x.go",
                    "+++ b/internal/x.go",
                    "@@ -1 +1 @@",
                )
            ),
            "",
        )
        missing = subprocess.CompletedProcess([], 128, b"", b"missing")
        with mock.patch(
            "check_analysis.subprocess.run",
            side_effect=(diff, missing),
        ):
            with self.assertRaises(RuntimeError):
                check_analysis.changed_existing_functions(Path("/tmp"), "base")

    def test_environment_base_cannot_override_persisted_change_base(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text("persisted\n", encoding="utf-8")
            persisted = subprocess.CompletedProcess([], 0, "a" * 40 + "\n", "")
            override = subprocess.CompletedProcess([], 0, "b" * 40 + "\n", "")
            with mock.patch.dict(
                "os.environ",
                {"SDD_BASE_REF": "HEAD"},
                clear=False,
            ), mock.patch(
                "check_analysis.subprocess.run",
                side_effect=(persisted, override),
            ):
                with self.assertRaises(ValueError):
                    check_analysis.resolve_base(change, root)


    # The three checks below exist because ten of thirty-six a092 artifacts
    # asserted branch coverage the source no longer had. Existence, hash and
    # AST-branch coverage all passed; the prose had simply never been
    # re-anchored after the function moved. A checker that reads only the
    # bundle's shape cannot tell a current map from a stale one.

    def test_prose_line_range_must_match_the_ast_range(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[],
                logic_source_line="- Source: `internal/sample.go` (111-151)",
            )
            errors = run_check(root)
        self.assertTrue(any("line range" in error for error in errors), errors)

    def test_a_matching_prose_line_range_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[],
                logic_source_line="- Source: `internal/sample.go` (2-2)",
                branch_source_line="Source: `internal/sample.go` (2-2).",
            )
            self.assertEqual(run_check(root), [])

    def test_an_absent_prose_line_range_is_still_accepted(self) -> None:
        # Most maps in the corpus cite no line range at all. Requiring one
        # would fail evidence that never claimed a coordinate.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(root, branches=[])
            self.assertEqual(run_check(root), [])

    def test_prose_branch_count_must_match_the_ast_branch_count(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            target = write_bundle(root, branches=[])
            (target / "function-logic-map.md").write_text(
                (target / "function-logic-map.md")
                .read_text(encoding="utf-8")
                .replace(
                    "## Inputs and invariants",
                    "- AST evidence: `ast.json` — branches 9\n## Inputs and invariants",
                ),
                encoding="utf-8",
            )
            errors = run_check(root)
        self.assertTrue(any("branch count" in error for error in errors), errors)

    def test_a_branch_count_claim_without_an_ast_anchor_is_not_read_as_one(self) -> None:
        # "미테스트 분기 5개" is prose about coverage, not a claim about what
        # the extractor found. Only a claim anchored to AST is checked.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            target = write_bundle(root, branches=[])
            (target / "branch-test-map.md").write_text(
                (target / "branch-test-map.md").read_text(encoding="utf-8")
                + "\n미테스트 분기 5개는 이 change의 범위가 아니다.\n",
                encoding="utf-8",
            )
            self.assertEqual(run_check(root), [])

    def test_branch_ids_absent_from_the_ast_are_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[{"id": "B1", "kind": "if", "at": {"line": 2}}],
                branch_rows="| B1 | leaf | test | yes | yes |\n| B2 | ghost | test | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(any("B2" in error for error in errors), errors)

    def test_the_branchless_happy_path_row_is_not_read_as_an_extra_id(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(root, branches=None)
            self.assertEqual(run_check(root), [])


def write_test_file(root: Path, name: str, tests: dict[str, int]) -> Path:
    """Write a Go test file where each test's body is padded to a known length.

    The maps under check cite lines inside a test, not its declaration, so the
    fixture has to give each test a body with addressable interior lines.
    """
    lines: list[str] = ["package sample", ""]
    for test, body in sorted(tests.items(), key=lambda item: item[1]):
        lines.append(f"func {test}(t *testing.T) {{")
        lines.extend(["\t_ = t" for _ in range(body)])
        lines.append("}")
        lines.append("")
    path = root / "internal" / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines), encoding="utf-8")
    return path


class TestNamedTestsAreOpened(unittest.TestCase):
    """B-T1: a map that names a test must be answerable by the tree.

    Ten of a092's rows carried false coverage claims through eighteen rounds
    because the checker never opened the file the row pointed at. It verified
    that files existed, that hashes were current and that every AST branch had
    a row -- all of which a fabricated test name satisfies.
    """

    def test_a_cited_test_that_exists_nowhere_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestNobodyWroteThis` | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(
            any("TestNobodyWroteThis" in error for error in errors), errors
        )

    def test_a_cited_test_that_exists_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_row_naming_no_test_at_all_is_still_accepted(self) -> None:
        """Rows that honestly say "없음" are the point of the map, not a defect."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | **없음** — 이 change의 대상 | no | no |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_cited_line_past_the_end_of_the_test_file_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` (`sample_test.go:900`) | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(any("900" in error for error in errors), errors)

    def test_a_test_named_in_one_file_beside_call_sites_in_another_is_accepted(self) -> None:
        """a091's `severityof` row does this and is correct.

        19판 first required the cited line to belong to the test named beside
        it. This row is why that rule was withdrawn: it names a test with its
        own coordinate and then cites two independent call sites of the function
        under test, in a different file. Which coordinate answers for which
        claim lives in the prose, not in line co-occurrence.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_test_file(root, "other_test.go", {"TestSomewhereElse": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows=(
                    "| B1 | leaf | `TestRunLeaf` `sample_test.go:4` "
                    "· `other_test.go:4` | yes | yes |"
                ),
            )
            self.assertEqual(run_check(root), [])

    def test_a_qualified_path_resolves_to_that_file_not_the_local_one(self) -> None:
        """`replay_test.go` exists in two packages. The citation says which.

        a098 hit this for real: the bare name resolved against the package under
        test and answered for internal/journal's 228-line file when the row meant
        internal/execgw's. Qualifying the path has to actually redirect the
        lookup, or the advice to qualify it is empty.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            # Same basename, two packages. The local one is short; the cited one
            # is long enough that the line number only fits there.
            write_test_file(root, "shared_test.go", {"TestNear": 2})
            far = root / "other" / "shared_test.go"
            far.parent.mkdir(parents=True, exist_ok=True)
            far.write_text(
                "package other\n" + "\n".join("// pad" for _ in range(60)),
                encoding="utf-8",
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `other/shared_test.go:40` | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_qualified_path_that_does_not_exist_is_not_silently_skipped(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `nowhere/sample_test.go:4` | yes | yes |",
            )
            # Unresolvable is not an error today -- the corpus cites files in
            # trees this checker does not own. Pinned so a future change to that
            # policy is a decision, not a side effect.
            self.assertEqual(run_check(root), [])

    def test_a_line_in_the_named_tests_doc_comment_is_accepted(self) -> None:
        """a092 cites two doc comments. A comment attached to a test names it."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            path.write_text(
                "package sample\n\n// TestRunLeaf: why this test exists.\n"
                "func TestRunLeaf(t *testing.T) {\n\t_ = t\n}\n",
                encoding="utf-8",
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` (`sample_test.go:3`) | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_line_in_a_shared_harness_in_the_same_file_is_accepted(self) -> None:
        """`obs_test.go:338` is `newNotifier`, a helper -- and honest evidence."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            path.write_text(
                path.read_text(encoding="utf-8")
                + "\nfunc newHarness(t *testing.T) int {\n\treturn 1\n}\n",
                encoding="utf-8",
            )
            total = len(path.read_text(encoding="utf-8").splitlines())
            write_bundle(
                root,
                branches=None,
                branch_rows=f"| B1 | leaf | `TestRunLeaf` (`sample_test.go:{total - 1}`) | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_coordinate_without_a_name_in_the_same_row_is_accepted(self) -> None:
        """a092's `newnotifier` row cites two direct call sites and names neither.

        The row is true -- both lines call the function under test. Requiring a
        name here would reject correct evidence, so the name/line agreement is
        checked only where the row supplies both.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(
                root, "sample_test.go", {"TestRunLeaf": 3, "TestSomethingElse": 3}
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `sample_test.go:9` — 직접 부른다 | yes | yes |",
            )
            self.assertEqual(run_check(root), [])


if __name__ == "__main__":
    unittest.main()
