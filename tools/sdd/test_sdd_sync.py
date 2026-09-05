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


class SourcePathDriftTest(unittest.TestCase):
    """이름만 보는 등록 검사는 경로 표류를 영원히 못 본다.

    2026-09-05 실측: 체크아웃을 옮기자 source 는 `tossos-…` 라는 **이름으로는**
    그대로 등록돼 있었고 `source_registered` 가 True 를 냈다. 그래서 `sources add`
    는 건너뛰어졌고, 등록된 경로는 옛 자리를 계속 가리켰다. 그 다음에 도는 sync 가
    낸 말은 `Not a git repository: /mnt/D/project/axipient/TossOS` 였다 — 참이지만
    **무엇을 고쳐야 하는지는 안 말한다.** 진단 하나를 통째로 사람이 다시 해야 했고,
    되돌리는 값은 페이지 5126개 재색인이었다.

    그래서 이름이 아니라 **경로**를 본다. 고치는 것은 자동으로 하지 않는다 —
    gbrain 이 요구하는 복구가 파괴적(remove 가 페이지까지 지운다)이라 사람의 승인
    없이 부를 수 없다. 이 검사가 소유한 것은 판정과 처방을 **말하는 것**까지다.
    """

    def completed(self, stdout: str, returncode: int = 0):
        return subprocess.CompletedProcess(
            args=["gbrain", "sources", "list"],
            returncode=returncode,
            stdout=stdout,
            stderr="",
        )

    def listing(self, path: str) -> str:
        return (
            "SOURCES\n"
            "───────\n"
            "  default               federated          0 pages  never synced\n"
            "  tossos-x              isolated        4261 pages  never synced\n"
            f"                        {path}\n"
        )

    def test_the_listed_path_is_read_back_for_the_named_source(self) -> None:
        listed = self.completed(self.listing("/new/place/TossOS"))
        self.assertEqual(
            sdd_sync.registered_path(listed, "tossos-x"), "/new/place/TossOS"
        )
        # 경로가 없는 항목(`default`)에 남의 경로를 붙여 읽으면 안 된다.
        self.assertIsNone(sdd_sync.registered_path(listed, "default"))

    def test_a_listing_without_paths_claims_nothing(self) -> None:
        """옛 gbrain 은 경로를 안 찍는다. 그때는 '표류 아님'이 아니라 '모름'이다."""
        listed = self.completed(
            "SOURCES\n  tossos-x                   isolated   816 pages\n"
        )
        self.assertIsNone(sdd_sync.registered_path(listed, "tossos-x"))

    def test_a_drifted_registration_is_named_and_stops_the_run(self) -> None:
        with mock.patch.object(sdd_sync.shutil, "which", lambda name: name != "codegraphcontext"), \
             mock.patch.object(sdd_sync.subprocess, "run") as probe, \
             mock.patch.object(sdd_sync, "run") as run, \
             mock.patch.object(sdd_sync, "record_index_state"), \
             mock.patch.object(sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"), \
             mock.patch.object(sdd_sync.Path, "exists", lambda self: True):
            probe.return_value = self.completed(
                self.listing("/mnt/D/project/axipient/TossOS")
            )
            run.return_value = True
            failures = sdd_sync.sync()

        self.assertIn("gbrain source path drifted", failures)
        gbrain_calls = [
            call.args[0]
            for call in run.call_args_list
            if "gbrain" in " ".join(call.args[0])
        ]
        self.assertEqual(
            gbrain_calls, [], "표류한 등록에 대고 add/sync 를 불렀다"
        )

    def test_a_registration_that_still_points_here_runs_the_sync(self) -> None:
        """참인 쪽도 잰다 — 아니면 '언제나 표류'가 위 시험을 통과한다."""
        with mock.patch.object(sdd_sync.shutil, "which", lambda name: name != "codegraphcontext"), \
             mock.patch.object(sdd_sync.subprocess, "run") as probe, \
             mock.patch.object(sdd_sync, "run") as run, \
             mock.patch.object(sdd_sync, "record_index_state"), \
             mock.patch.object(sdd_sync.Path, "read_text", lambda self, encoding=None: "tossos-x\n"), \
             mock.patch.object(sdd_sync.Path, "exists", lambda self: True):
            probe.return_value = self.completed(self.listing(str(sdd_sync.ROOT)))
            run.return_value = True
            failures = sdd_sync.sync()

        self.assertNotIn("gbrain source path drifted", failures)
        synced = [
            call.args[0]
            for call in run.call_args_list
            if "sync" in " ".join(call.args[0])
        ]
        self.assertTrue(synced, "제자리인 등록인데 sync 를 안 불렀다")


if __name__ == "__main__":
    unittest.main()
