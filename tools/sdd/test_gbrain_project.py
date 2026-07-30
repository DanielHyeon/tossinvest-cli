from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
import unittest
from pathlib import Path

from gbrain_project import GBRAIN_HOME, project_environment


class GBrainProjectTest(unittest.TestCase):
    def test_project_environment_overrides_global_brain_home(self) -> None:
        env = project_environment()
        self.assertEqual(env["GBRAIN_HOME"], str(GBRAIN_HOME))
        self.assertEqual(
            GBRAIN_HOME,
            GBRAIN_HOME.parents[1] / ".sdd" / "gbrain-home",
        )

    def make_project(self, temporary: str) -> tuple[Path, Path, dict[str, str]]:
        root = Path(temporary) / "repo"
        script = root / "tools" / "sdd" / "gbrain_project.py"
        script.parent.mkdir(parents=True)
        shutil.copyfile(Path(__file__).with_name("gbrain_project.py"), script)

        binary = Path(temporary) / "bin" / "gbrain"
        binary.parent.mkdir()
        binary.write_text(
            "#!/usr/bin/env python3\n"
            "import os, sys, time\n"
            "with open(os.environ['GBRAIN_TEST_CALLS'], 'a', encoding='utf-8') as f:\n"
            "    f.write(' '.join(sys.argv[1:]) + '\\n')\n"
            "ready = os.environ.get('GBRAIN_TEST_READY')\n"
            "if ready:\n"
            "    open(ready, 'w', encoding='utf-8').write(str(os.getpid()))\n"
            "if sys.argv[1:] == ['serve']:\n"
            "    time.sleep(60)\n"
            "else:\n"
            "    print('fake-gbrain ' + ' '.join(sys.argv[1:]))\n",
            encoding="utf-8",
        )
        binary.chmod(0o755)
        calls = Path(temporary) / "calls.txt"
        env = os.environ.copy()
        env["PATH"] = f"{binary.parent}{os.pathsep}{env['PATH']}"
        env["GBRAIN_TEST_CALLS"] = str(calls)
        return script, calls, env

    def wait_for(self, path: Path, timeout: float = 2.0) -> None:
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            if path.exists():
                return
            time.sleep(0.02)
        self.fail(f"timed out waiting for {path}")

    def test_duplicate_serve_exits_busy_before_starting_second_gbrain(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, calls, env = self.make_project(temporary)
            ready = Path(temporary) / "ready"
            holder_env = env | {"GBRAIN_TEST_READY": str(ready)}
            holder = subprocess.Popen(
                [sys.executable, str(script), "serve"],
                env=holder_env,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )
            try:
                self.wait_for(ready)
                started = time.monotonic()
                contender = subprocess.run(
                    [sys.executable, str(script), "serve"],
                    env=env,
                    capture_output=True,
                    text=True,
                    timeout=2,
                    check=False,
                )
                elapsed = time.monotonic() - started
                self.assertEqual(contender.returncode, 75)
                self.assertIn("[gbrain-project] busy:", contender.stderr)
                self.assertLess(elapsed, 1.0)
                self.assertEqual(calls.read_text(encoding="utf-8").splitlines(), ["serve"])
            finally:
                holder.terminate()
                holder.wait(timeout=2)

    def test_process_exit_releases_singleton_without_deleting_lock_file(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, _, env = self.make_project(temporary)
            ready = Path(temporary) / "ready"
            holder = subprocess.Popen(
                [sys.executable, str(script), "serve"],
                env=env | {"GBRAIN_TEST_READY": str(ready)},
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )
            self.wait_for(ready)
            holder.terminate()
            holder.wait(timeout=2)

            completed = subprocess.run(
                [sys.executable, str(script), "version"],
                env=env,
                capture_output=True,
                text=True,
                timeout=2,
                check=False,
            )
            self.assertEqual(completed.returncode, 0, completed.stderr)
            self.assertIn("fake-gbrain version", completed.stdout)

    def test_live_legacy_pglite_owner_is_busy_and_its_lock_is_preserved(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, calls, env = self.make_project(temporary)
            pglite_lock = (
                script.parents[2]
                / ".sdd"
                / "gbrain-home"
                / ".gbrain"
                / "brain.pglite"
                / ".gbrain-lock"
                / "lock"
            )
            pglite_lock.parent.mkdir(parents=True)
            metadata = {
                "pid": os.getpid(),
                "acquired_at": int(time.time() * 1000),
                "refreshed_at": int(time.time() * 1000),
                "command": "gbrain --token=must-not-leak serve",
            }
            pglite_lock.write_text(json.dumps(metadata), encoding="utf-8")

            completed = subprocess.run(
                [sys.executable, str(script), "version"],
                env=env,
                capture_output=True,
                text=True,
                timeout=2,
                check=False,
            )
            self.assertEqual(completed.returncode, 75)
            self.assertIn(f"legacy PGLite owner pid={os.getpid()}", completed.stderr)
            self.assertIn("command='gbrain serve'", completed.stderr)
            self.assertNotIn("must-not-leak", completed.stderr)
            self.assertEqual(json.loads(pglite_lock.read_text(encoding="utf-8")), metadata)
            self.assertFalse(calls.exists(), "the real gbrain executable was invoked")

    def test_stale_legacy_heartbeat_is_left_for_gbrain_recovery(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, _, env = self.make_project(temporary)
            pglite_lock = (
                script.parents[2]
                / ".sdd"
                / "gbrain-home"
                / ".gbrain"
                / "brain.pglite"
                / ".gbrain-lock"
                / "lock"
            )
            pglite_lock.parent.mkdir(parents=True)
            metadata = {
                "pid": os.getpid(),
                "acquired_at": int(time.time() * 1000) - 700_000,
                "refreshed_at": int(time.time() * 1000) - 700_000,
                "command": "gbrain serve",
            }
            pglite_lock.write_text(json.dumps(metadata), encoding="utf-8")

            completed = subprocess.run(
                [sys.executable, str(script), "version"],
                env=env,
                capture_output=True,
                text=True,
                timeout=2,
                check=False,
            )
            self.assertEqual(completed.returncode, 0, completed.stderr)
            self.assertIn("fake-gbrain version", completed.stdout)
            self.assertEqual(json.loads(pglite_lock.read_text(encoding="utf-8")), metadata)

    def test_lock_metadata_does_not_persist_command_arguments(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, _, env = self.make_project(temporary)
            completed = subprocess.run(
                [
                    sys.executable,
                    str(script),
                    "auth",
                    "create",
                    "--token",
                    "must-not-be-persisted",
                ],
                env=env,
                capture_output=True,
                text=True,
                timeout=2,
                check=False,
            )
            self.assertEqual(completed.returncode, 0, completed.stderr)
            metadata_path = (
                script.parents[2]
                / ".sdd"
                / "gbrain-home"
                / ".gbrain"
                / "tossos-process.lock"
            )
            metadata = metadata_path.read_text(encoding="utf-8")
            self.assertIn('"command": "gbrain auth"', metadata)
            self.assertNotIn("must-not-be-persisted", metadata)


if __name__ == "__main__":
    unittest.main()
