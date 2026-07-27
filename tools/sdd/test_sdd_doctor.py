from __future__ import annotations

import subprocess
import unittest
from unittest import mock

import sdd_doctor
import sdd_sync


class SDDDoctorTests(unittest.TestCase):
    def test_required_cli_missing_has_install_hint(self) -> None:
        with mock.patch("sdd_doctor.shutil.which", return_value=None):
            status = sdd_doctor.command_status("ast-grep", ("ast-grep", "--version"))
        self.assertFalse(status["ok"])
        self.assertIn("install ast-grep", status["detail"])

    def test_command_timeout_fails_closed(self) -> None:
        with mock.patch("sdd_doctor.shutil.which", return_value="/bin/tool"), mock.patch(
            "sdd_doctor.subprocess.run",
            side_effect=subprocess.TimeoutExpired(["tool"], 8),
        ):
            status = sdd_doctor.command_status("tool", ("tool", "--version"))
        self.assertFalse(status["ok"])
        self.assertIn("timed out", status["detail"])

    def test_missing_service_is_advisory(self) -> None:
        completed = subprocess.CompletedProcess([], 1, "", "missing")
        with mock.patch("sdd_doctor.shutil.which", return_value="/bin/docker"), mock.patch(
            "sdd_doctor.subprocess.run",
            return_value=completed,
        ):
            status = sdd_doctor.service_status("missing")
        self.assertFalse(status["ok"])
        self.assertFalse(status["required"])

    def test_service_timeout_is_advisory(self) -> None:
        with mock.patch("sdd_doctor.shutil.which", return_value="/bin/docker"), mock.patch(
            "sdd_doctor.subprocess.run",
            side_effect=subprocess.TimeoutExpired(["docker"], 8),
        ):
            status = sdd_doctor.service_status("slow")
        self.assertFalse(status["ok"])
        self.assertFalse(status["required"])
        self.assertIn("probe unavailable", status["detail"])

    def test_sync_reports_all_missing_indexers(self) -> None:
        with mock.patch("sdd_sync.shutil.which", return_value=None):
            failures = sdd_sync.sync(include_gbrain=True)
        self.assertIn("codegraph missing", failures)
        self.assertIn("codegraphcontext missing", failures)
        self.assertIn("gbrain missing", failures)


if __name__ == "__main__":
    unittest.main()
