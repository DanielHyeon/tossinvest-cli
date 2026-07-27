from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import capture_change_base


class CaptureChangeBaseTests(unittest.TestCase):
    def test_capture_is_full_hash_and_write_once(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "add-example"
            change.mkdir(parents=True)
            (change / "proposal.md").write_text("# proposal\n", encoding="utf-8")
            completed = subprocess.CompletedProcess(
                [],
                0,
                "a" * 40 + "\n",
                "",
            )
            with mock.patch(
                "capture_change_base.subprocess.run",
                return_value=completed,
            ):
                target = capture_change_base.capture("add-example", root)
                self.assertEqual(target.read_text(encoding="utf-8"), "a" * 40 + "\n")
                with self.assertRaises(ValueError):
                    capture_change_base.capture("add-example", root)
