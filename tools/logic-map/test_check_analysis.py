from __future__ import annotations

import json
import hashlib
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import check_analysis


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


if __name__ == "__main__":
    unittest.main()
