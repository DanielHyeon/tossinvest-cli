from __future__ import annotations

import os
import stat
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import sdd_doctor
import sdd_sync


class SDDDoctorTests(unittest.TestCase):
    def _root(self, directory: Path, requirements: str = "typedb-driver==3.11.5\n") -> Path:
        root = directory / "repo"
        (root / "tools" / "sdd").mkdir(parents=True)
        (root / "tools" / "sdd" / "requirements.txt").write_text(requirements, encoding="utf-8")
        return root

    @staticmethod
    def _executable(path: Path) -> Path:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("#!/bin/sh\n", encoding="utf-8")
        path.chmod(path.stat().st_mode | stat.S_IXUSR)
        return path

    @staticmethod
    def _report_status(name: str, command: tuple[str, ...]) -> dict:
        if name == "typedb-driver":
            return {"ok": True, "detail": "typedb-driver 3.11.5"}
        return {"ok": True, "detail": "available"}

    def _report(self, root: Path, environment: dict[str, str]) -> tuple[dict, mock.Mock]:
        with mock.patch.dict(os.environ, environment, clear=True), mock.patch(
            "sdd_doctor.command_status", side_effect=self._report_status
        ) as command_status, mock.patch("sdd_doctor.skill_status", return_value={}), mock.patch(
            "sdd_doctor.service_status", return_value={"ok": True, "detail": "available", "required": False}
        ):
            return sdd_doctor.report(root), command_status

    def test_absent_override_preserves_local_venv_command(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = self._root(Path(tmp))
            local = self._executable(root / ".sdd" / ".venv" / "bin" / "python")
            result, command_status = self._report(root, {})
        typedb = result["python_modules"]["typedb-driver"]
        self.assertTrue(typedb["ok"])
        self.assertIn("mode=local", typedb["detail"])
        typedb_call = next(call for call in command_status.call_args_list if call.args[0] == "typedb-driver")
        self.assertEqual(typedb_call.args[1][0], str(local))

    def test_absent_override_keeps_local_setup_hint(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            result, command_status = self._report(self._root(Path(tmp)), {})
        typedb = result["python_modules"]["typedb-driver"]
        self.assertFalse(typedb["ok"])
        self.assertIn("run `make sdd-infra`", typedb["detail"])
        self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

    def test_explicit_invalid_values_fail_without_local_fallback(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            root = self._root(directory)
            local = self._executable(root / ".sdd" / ".venv" / "bin" / "python")
            external = directory / "external"
            directory_path = external / "directory"; directory_path.mkdir(parents=True)
            non_executable = external / "not-executable"; non_executable.parent.mkdir(parents=True, exist_ok=True); non_executable.write_text("x", encoding="utf-8")
            values = ("", "relative/python", str(external / "missing"), str(directory_path), str(non_executable))
            for value in values:
                with self.subTest(value=value):
                    result, command_status = self._report(root, {"SDD_PYTHON": value})
                    typedb = result["python_modules"]["typedb-driver"]
                    self.assertFalse(typedb["ok"])
                    self.assertIn("mode=external", typedb["detail"])
                    self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))
                    self.assertNotIn(str(local), typedb["detail"])

    def test_both_path_boundaries_and_symlink_directions(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            root = self._root(directory)
            external_target = self._executable(directory / "external" / "python")
            lexical_inside = root / "bin" / "external-link"; lexical_inside.parent.mkdir()
            lexical_inside.symlink_to(external_target)
            resolved_inside = directory / "outside-link-to-repo"
            internal_target = self._executable(root / "bin" / "internal-python")
            resolved_inside.symlink_to(internal_target)
            external_link = directory / "external" / "linked-python"; external_link.symlink_to(external_target)
            sibling = self._executable(directory / "repo-sibling" / "python")
            expected = {
                lexical_inside: False,
                resolved_inside: False,
                external_link: True,
                sibling: True,
            }
            for candidate, ok in expected.items():
                with self.subTest(candidate=candidate):
                    result, command_status = self._report(root, {"SDD_PYTHON": str(candidate)})
                    typedb = result["python_modules"]["typedb-driver"]
                    self.assertEqual(typedb["ok"], ok)
                    if ok:
                        call = next(call for call in command_status.call_args_list if call.args[0] == "typedb-driver")
                        self.assertEqual(call.args[1][0], str(candidate))
                        self.assertIn(f"resolved={candidate.resolve()}", typedb["detail"])
                    else:
                        self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

    def test_lexical_boundary_normalizes_alias_and_dot_segments_before_resolving_links(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            real_root = self._root(directory)
            alias_root = directory / "alias-root"
            alias_root.symlink_to(real_root, target_is_directory=True)
            external = self._executable(directory / "external" / "python")
            local_link = real_root / ".sdd" / ".venv" / "bin" / "python"
            local_link.parent.mkdir(parents=True)
            local_link.symlink_to(external)

            (directory / "outside").mkdir()
            escaped_candidate = directory / "outside" / ".." / "repo" / ".sdd" / ".venv" / "bin" / "python"
            dot_root = directory / "repo" / ".." / "repo"
            cases = (
                (alias_root, local_link, "alias root"),
                (dot_root, local_link, "root dot segments"),
                (real_root, escaped_candidate, "candidate dot segments"),
            )
            for supplied_root, candidate, label in cases:
                with self.subTest(label=label):
                    result, command_status = self._report(supplied_root, {"SDD_PYTHON": str(candidate)})
                    typedb = result["python_modules"]["typedb-driver"]
                    self.assertFalse(typedb["ok"])
                    self.assertIn("invalid selection: selection is lexically inside repository", typedb["detail"])
                    self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

            result, command_status = self._report(alias_root, {"SDD_PYTHON": str(external)})
            typedb = result["python_modules"]["typedb-driver"]
            self.assertTrue(typedb["ok"])
            typedb_call = next(call for call in command_status.call_args_list if call.args[0] == "typedb-driver")
            self.assertEqual(typedb_call.args[1][0], str(external))

    def test_symlink_loop_is_an_invalid_external_selection_and_safe_local_diagnostic(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            root = self._root(directory)
            external_a = directory / "external-a"
            external_b = directory / "external-b"
            external_a.symlink_to(external_b)
            external_b.symlink_to(external_a)

            result, command_status = self._report(root, {"SDD_PYTHON": str(external_a)})
            typedb = result["python_modules"]["typedb-driver"]
            self.assertFalse(typedb["ok"])
            self.assertIn("mode=external", typedb["detail"])
            self.assertIn("invalid selection", typedb["detail"])
            self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

            local_a = root / ".sdd" / ".venv" / "bin" / "python"
            local_b = root / ".sdd" / ".venv" / "bin" / "python-b"
            local_a.parent.mkdir(parents=True)
            local_a.symlink_to(local_b)
            local_b.symlink_to(local_a)
            local_result, local_command_status = self._report(root, {})
            local_typedb = local_result["python_modules"]["typedb-driver"]
            self.assertFalse(local_typedb["ok"])
            self.assertIn("mode=local", local_typedb["detail"])
            self.assertIn("unresolved", local_typedb["detail"])
            self.assertFalse(any(call.args[0] == "typedb-driver" for call in local_command_status.call_args_list))

    def test_explicit_mode_requires_exact_single_pinned_driver(self) -> None:
        cases = {
            "missing": "# no driver\n",
            "malformed": "typedb-driver>=3.11.5\n",
            "multiple": "typedb-driver==3.11.5\ntypedb-driver==3.11.5\n",
            "nonexact": "typedb-driver==3.11.5 # comment\n",
        }
        for name, requirements in cases.items():
            with self.subTest(name=name), tempfile.TemporaryDirectory() as tmp:
                directory = Path(tmp)
                root = self._root(directory, requirements)
                external = self._executable(directory / "external" / "python")
                result, command_status = self._report(root, {"SDD_PYTHON": str(external)})
                typedb = result["python_modules"]["typedb-driver"]
                self.assertFalse(typedb["ok"])
                self.assertIn("requirements", typedb["detail"])
                self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

    def test_explicit_mode_malformed_utf8_requirements_fails_without_driver_probe(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            root = self._root(directory)
            (root / "tools" / "sdd" / "requirements.txt").write_bytes(b"typedb-driver==3.11.5\n\xff")
            external = self._executable(directory / "external" / "python")
            result, command_status = self._report(root, {"SDD_PYTHON": str(external)})
        typedb = result["python_modules"]["typedb-driver"]
        self.assertFalse(typedb["ok"])
        self.assertIn("cannot read requirements", typedb["detail"])
        self.assertFalse(any(call.args[0] == "typedb-driver" for call in command_status.call_args_list))

    def test_nul_path_is_defensive_direct_helper_input_not_an_os_environment_case(self) -> None:
        """NUL cannot occur in real os.environ; this only covers direct helper defense."""
        with tempfile.TemporaryDirectory() as tmp:
            root = self._root(Path(tmp))
            selected, resolved, error = sdd_doctor._external_python(root, "/tmp/external\x00python")
        self.assertIsNone(selected)
        self.assertIn("unresolved", resolved)
        self.assertEqual(error, "selection cannot be resolved")

    def test_explicit_mode_rejects_missing_and_mismatched_driver(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            root = self._root(directory)
            local = self._executable(root / ".sdd" / ".venv" / "bin" / "python")
            external = self._executable(directory / "external" / "python")
            for detail, process_ok in (("typedb-driver missing", False), ("typedb-driver 0.0.0", True)):
                with self.subTest(detail=detail), mock.patch.dict(os.environ, {"SDD_PYTHON": str(external)}, clear=True), mock.patch(
                    "sdd_doctor.command_status", side_effect=lambda name, command: {"ok": process_ok, "detail": detail} if name == "typedb-driver" else {"ok": True, "detail": "available"}
                ) as command_status, mock.patch("sdd_doctor.skill_status", return_value={}), mock.patch("sdd_doctor.service_status", return_value={}):
                    result = sdd_doctor.report(root)
                    typedb = result["python_modules"]["typedb-driver"]
                    self.assertFalse(typedb["ok"])
                    self.assertIn("expected typedb-driver==3.11.5", typedb["detail"])
                    self.assertIn(f"observed {detail}", typedb["detail"])
                    driver_commands = [call.args[1] for call in command_status.call_args_list if call.args[0] == "typedb-driver"]
                    self.assertEqual([command[0] for command in driver_commands], [str(external)])
                    self.assertNotIn(str(local), [command[0] for command in driver_commands])

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
