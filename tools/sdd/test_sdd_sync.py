from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import sdd_sync


class SourceProbeTest(unittest.TestCase):
    """`gbrain sources list` can fail while exiting 0.

    Measured 2026-07-28: a wedged `gbrain serve` held the project lock and kept
    its heartbeat alive while its PGLite postmaster was dead. Every probe
    answered `gbrain sources: connect timed out.` on stdout with returncode 0.

    The old reading was `source not in stdout` -> "not registered" -> run
    `sources add`. That said `gbrain source registration` failed, which is the
    wrong sentence: the source was registered and had 816 pages. The remedy for
    an unreachable engine (find and kill the wedged serve) is nothing like the
    remedy for a missing source, so naming the wrong one costs a diagnosis.
    """

    def completed(self, stdout: str, returncode: int = 0):
        return subprocess.CompletedProcess(
            args=["gbrain", "sources", "list"],
            returncode=returncode,
            stdout=stdout,
            stderr="",
        )

    def test_a_timed_out_probe_is_not_read_as_a_missing_source(self) -> None:
        listed = self.completed("gbrain sources: connect timed out.\n")
        self.assertFalse(sdd_sync.source_registered(listed, "tossos-x"))
        self.assertTrue(sdd_sync.probe_unreachable(listed))

    def test_a_real_listing_reports_the_source_it_contains(self) -> None:
        listed = self.completed(
            "SOURCES\n"
            "  default                    federated    0 pages\n"
            "  tossos-x                   isolated   816 pages\n"
        )
        self.assertFalse(sdd_sync.probe_unreachable(listed))
        self.assertTrue(sdd_sync.source_registered(listed, "tossos-x"))
        self.assertFalse(sdd_sync.source_registered(listed, "tossos-other"))

    def test_a_reachable_engine_with_no_sources_still_registers(self) -> None:
        listed = self.completed("SOURCES\n")
        self.assertFalse(sdd_sync.probe_unreachable(listed))
        self.assertFalse(sdd_sync.source_registered(listed, "tossos-x"))

    def test_an_unreachable_engine_names_itself_and_skips_the_add(self) -> None:
        """The failure list must say the engine is unreachable, and the run must
        not go on to call `sources add`/`sync` — both would fail against the
        same dead engine and bury the one fact that matters."""
        with mock.patch.object(sdd_sync.shutil, "which", lambda name: name != "codegraphcontext"), \
             mock.patch.object(sdd_sync.subprocess, "run") as probe, \
             mock.patch.object(sdd_sync, "run") as run, \
             mock.patch.object(sdd_sync, "record_index_state"), \
             mock.patch.object(sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"), \
             mock.patch.object(sdd_sync.Path, "exists", lambda self: True):
            probe.return_value = self.completed("gbrain sources: connect timed out.\n")
            run.return_value = True
            failures = sdd_sync.sync()

        self.assertIn("gbrain engine unreachable", failures)
        self.assertNotIn("gbrain source registration", failures)
        self.assertNotIn("gbrain project sync", failures)
        gbrain_calls = [
            call.args[0] for call in run.call_args_list if call.args[0][0] == "gbrain"
        ]
        self.assertEqual(gbrain_calls, [], "a command ran against a dead engine")

    def test_nonzero_source_probe_cannot_be_overwritten_by_later_success(self) -> None:
        failed_probe = subprocess.CompletedProcess(
            args=["gbrain-project", "sources", "list"],
            returncode=17,
            stdout="",
            stderr="schema read failed\n",
        )
        with mock.patch.object(
            sdd_sync.shutil,
            "which",
            lambda name: name != "codegraphcontext",
        ), mock.patch.object(
            sdd_sync.subprocess, "run", return_value=failed_probe
        ), mock.patch.object(
            sdd_sync, "run", return_value=True
        ) as run, mock.patch.object(
            sdd_sync, "record_index_state"
        ) as record, mock.patch.object(
            sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"
        ), mock.patch.object(
            sdd_sync.Path, "exists", lambda self: True
        ):
            failures = sdd_sync.sync()

        self.assertIn("gbrain source probe", failures)
        gbrain_calls = [
            call.args[0]
            for call in run.call_args_list
            if "gbrain" in " ".join(call.args[0])
        ]
        self.assertEqual(gbrain_calls, [])
        record.assert_called_once_with({"codegraph"})

    def test_project_runner_executes_wrapper_and_preserves_nonbusy_failure(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "repo"
            wrapper = root / "tools" / "sdd" / "gbrain_project.py"
            wrapper.parent.mkdir(parents=True)
            shutil.copyfile(
                Path(__file__).with_name("gbrain_project.py"),
                wrapper,
            )
            binary = Path(temporary) / "bin" / "gbrain"
            binary.parent.mkdir()
            binary.write_text(
                "#!/usr/bin/env python3\n"
                "import sys\n"
                "print('schema read failed', file=sys.stderr)\n"
                "raise SystemExit(17)\n",
                encoding="utf-8",
            )
            binary.chmod(0o755)
            env = os.environ.copy()
            env["PATH"] = f"{binary.parent}{os.pathsep}{env['PATH']}"

            with mock.patch.object(sdd_sync, "GBRAIN_WRAPPER", wrapper):
                completed = sdd_sync.run_project_gbrain(
                    ["sources", "list"],
                    timeout=2,
                    env=env,
                )

        self.assertIsNotNone(completed)
        self.assertEqual(completed.returncode, 17)
        self.assertFalse(sdd_sync.project_gbrain_busy(completed))

    def test_active_project_owner_is_advisory_and_skips_all_gbrain_work(self) -> None:
        busy = subprocess.CompletedProcess(
            args=["gbrain-project", "sources", "list"],
            returncode=75,
            stdout="",
            stderr=(
                "[gbrain-project] busy: owner pid=123 command='gbrain serve' "
                "home=/tmp/project/.sdd/gbrain-home\n"
            ),
        )
        context_stats = subprocess.CompletedProcess(
            args=["codegraphcontext", "stats", "."],
            returncode=0,
            stdout="Repository: TossOS\n",
            stderr="",
        )
        with mock.patch.object(
            sdd_sync.shutil,
            "which",
            return_value="/usr/bin/tool",
        ), mock.patch.object(
            sdd_sync.subprocess, "run", side_effect=[context_stats, busy]
        ) as probe, mock.patch.object(
            sdd_sync, "run", return_value=True
        ) as run, mock.patch.object(
            sdd_sync, "record_index_state"
        ) as record, mock.patch.object(
            sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"
        ), mock.patch.object(
            sdd_sync.Path, "exists", lambda self: True
        ):
            failures = sdd_sync.sync()

        self.assertEqual(failures, [])
        probe_command = probe.call_args_list[-1].args[0]
        self.assertEqual(probe_command[0], sys.executable)
        self.assertEqual(probe_command[1], str(sdd_sync.GBRAIN_WRAPPER))
        self.assertEqual(probe_command[2:], ["sources", "list"])
        gbrain_calls = [
            call.args[0]
            for call in run.call_args_list
            if "gbrain" in " ".join(call.args[0])
        ]
        self.assertEqual(gbrain_calls, [])
        record.assert_called_once_with({"codegraph", "codegraphcontext"})

    def test_owner_that_starts_between_probe_and_sync_is_still_advisory(self) -> None:
        context_stats = subprocess.CompletedProcess(
            args=["codegraphcontext", "stats", "."],
            returncode=0,
            stdout="Repository: TossOS\n",
            stderr="",
        )
        listed = self.completed(
            "SOURCES\n"
            "  tossos-x                   isolated   816 pages\n"
        )
        busy = subprocess.CompletedProcess(
            args=["gbrain-project", "sync"],
            returncode=75,
            stdout="",
            stderr="[gbrain-project] busy: owner pid=456 command='gbrain serve'\n",
        )
        with mock.patch.object(
            sdd_sync.shutil, "which", return_value="/usr/bin/tool"
        ), mock.patch.object(
            sdd_sync.subprocess,
            "run",
            side_effect=[context_stats, listed, busy],
        ), mock.patch.object(
            sdd_sync, "run", return_value=True
        ), mock.patch.object(
            sdd_sync, "record_index_state"
        ) as record, mock.patch.object(
            sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"
        ), mock.patch.object(
            sdd_sync.Path, "exists", lambda self: True
        ):
            failures = sdd_sync.sync()

        self.assertEqual(failures, [])
        record.assert_called_once_with({"codegraph", "codegraphcontext"})

    def test_failed_source_registration_cannot_record_gbrain_freshness(self) -> None:
        context_stats = subprocess.CompletedProcess(
            args=["codegraphcontext", "stats", "."],
            returncode=0,
            stdout="Repository: TossOS\n",
            stderr="",
        )
        listed = self.completed("SOURCES\n")
        registration_failed = subprocess.CompletedProcess(
            args=["gbrain-project", "sources", "add"],
            returncode=19,
            stdout="",
            stderr="source registration failed\n",
        )
        sync_succeeded = subprocess.CompletedProcess(
            args=["gbrain-project", "sync"],
            returncode=0,
            stdout="sync complete\n",
            stderr="",
        )
        with mock.patch.object(
            sdd_sync.shutil, "which", return_value="/usr/bin/tool"
        ), mock.patch.object(
            sdd_sync.subprocess,
            "run",
            side_effect=[
                context_stats,
                listed,
                registration_failed,
                sync_succeeded,
            ],
        ), mock.patch.object(
            sdd_sync, "run", return_value=True
        ), mock.patch.object(
            sdd_sync, "record_index_state"
        ) as record, mock.patch.object(
            sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"
        ), mock.patch.object(
            sdd_sync.Path, "exists", lambda self: True
        ):
            failures = sdd_sync.sync()

        self.assertIn("gbrain source registration", failures)
        record.assert_called_once_with({"codegraph", "codegraphcontext"})


if __name__ == "__main__":
    unittest.main()
