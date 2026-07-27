from __future__ import annotations

import subprocess
import unittest
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


if __name__ == "__main__":
    unittest.main()
