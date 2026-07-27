from __future__ import annotations

import tempfile
import unittest
from pathlib import Path
from unittest import mock

import check_index_freshness as freshness


class IndexFreshnessTests(unittest.TestCase):
    def test_missing_physical_codegraph_database_is_unhealthy(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            ok, detail = freshness.codegraph_index_status(Path(tmp))
            self.assertFalse(ok)
            self.assertIn("missing or empty", detail)

    def test_missing_hard_index_fails_and_advisory_only_warns(self) -> None:
        with tempfile.TemporaryDirectory() as tmp, mock.patch.object(
            freshness,
            "repository_fingerprint",
            return_value="current",
        ):
            root = Path(tmp)
            errors, warnings = freshness.check(root)
            self.assertTrue(any("CodeGraph" in error for error in errors))
            self.assertEqual(len(warnings), 2)

    def test_recorded_hard_index_passes_with_stale_advisory_warning(self) -> None:
        with tempfile.TemporaryDirectory() as tmp, mock.patch.object(
            freshness,
            "repository_fingerprint",
            return_value="current",
        ), mock.patch.object(
            freshness,
            "codegraph_index_status",
            return_value=(True, "ok"),
        ), mock.patch(
            "check_index_freshness.subprocess.run",
        ) as process:
            process.return_value.stdout = "abc123\n"
            root = Path(tmp)
            state = root / ".sdd" / "index-state.json"
            freshness.record_index_state({"codegraph"}, root, state)
            errors, warnings = freshness.check(root, state)
            self.assertEqual(errors, [])
            self.assertEqual(len(warnings), 2)
