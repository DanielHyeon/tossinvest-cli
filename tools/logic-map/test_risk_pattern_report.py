from __future__ import annotations

import json
import subprocess
import unittest
from unittest import mock

import risk_pattern_report as report


class RiskPatternReportTests(unittest.TestCase):
    def test_scan_accepts_findings_exit_and_parses_json(self) -> None:
        finding = {"ruleId": "go-panic", "range": {"start": {"line": 2}}}
        completed = subprocess.CompletedProcess([], 1, json.dumps([finding]), "")
        with mock.patch("risk_pattern_report.subprocess.run", return_value=completed):
            self.assertEqual(report.scan("sample.go"), [finding])

    def test_scan_rejects_tool_failure_and_malformed_json(self) -> None:
        failed = subprocess.CompletedProcess([], 2, "", "bad config")
        with mock.patch("risk_pattern_report.subprocess.run", return_value=failed):
            with self.assertRaises(RuntimeError):
                report.scan("sample.go")
        malformed = subprocess.CompletedProcess([], 0, "not-json", "")
        with mock.patch("risk_pattern_report.subprocess.run", return_value=malformed):
            with self.assertRaises(json.JSONDecodeError):
                report.scan("sample.go")

    def test_markdown_handles_empty_and_escapes_pipes(self) -> None:
        empty = report.markdown("sample.go", [])
        self.assertIn("No configured risk pattern matched", empty)
        rendered = report.markdown(
            "sample.go",
            [
                {
                    "ruleId": "rule",
                    "range": {"start": {"line": 0}},
                    "message": "left|right",
                }
            ],
        )
        self.assertIn("left\\|right", rendered)
        self.assertIn("sample.go:1", rendered)


if __name__ == "__main__":
    unittest.main()
