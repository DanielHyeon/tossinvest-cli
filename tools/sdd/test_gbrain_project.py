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

    # 디렉터리 이름이 바뀌면 config.json 의 절대경로가 **혼자 뒤처진다.**
    #
    # 2026-09-05 에 실제로 그랬다: 워크트리를 옮기자 래퍼는 `GBRAIN_HOME` 을
    # `__file__` 로 다시 계산해 옳게 넘겼는데, 그 홈 **안의** config.json 이 옛
    # 절대경로를 쥐고 있어서 gbrain 이 저장소 밖에 브레인을 새로 팠다. 42MB 짜리
    # 부분 재색인이 생겼고 진짜 459MB 브레인은 5시간 동안 아무도 안 썼다.
    # 조용한 이유는 이것이다 — 오류가 아니라 **다른 자리의 성공**이기 때문이다.
    #
    # 상대경로로는 못 고친다. `database_path` 는 PGLite 로 그대로 넘어가 프로세스
    # cwd 기준으로 풀리고(더 나쁜 같은 고장), 아예 빼면 in-memory 로 떨어져 조용히
    # 휘발한다. gbrain 은 `GBRAIN_HOME` 자체도 절대경로만 받는다.
    #
    # 그래서 이 값은 **저장하지 않고 유도한다**: 자기 위치를 이미 아는 래퍼가
    # 실행 직전에 홈과 어긋난 값을 제자리로 돌린다.
    def test_a_renamed_checkout_repoints_the_brain_at_its_own_home(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, _, env = self.make_project(temporary)
            home = script.parents[2] / ".sdd" / "gbrain-home" / ".gbrain"
            home.mkdir(parents=True)
            stale = "/mnt/D/project/axipient/TossOS/.sdd/gbrain-home/.gbrain/brain.pglite"
            (home / "config.json").write_text(
                json.dumps(
                    {
                        "engine": "pglite",
                        "database_path": stale,
                        "embedding_disabled": True,
                    },
                    indent=2,
                )
                + "\n",
                encoding="utf-8",
            )
            completed = subprocess.run(
                [sys.executable, str(script), "search", "anything"],
                env=env,
                capture_output=True,
                text=True,
                timeout=5,
                check=False,
            )
            self.assertEqual(completed.returncode, 0, completed.stderr)
            written = json.loads((home / "config.json").read_text(encoding="utf-8"))
            self.assertEqual(
                written["database_path"], str(home / "brain.pglite"),
                "래퍼가 이름이 바뀐 체크아웃의 브레인을 제 홈으로 돌려놓지 않았다",
            )
            # 나머지 칸은 래퍼가 손대지 않는다 — 이 도구가 소유한 값은 경로 하나다.
            self.assertEqual(written["engine"], "pglite")
            self.assertEqual(written["embedding_disabled"], True)

    # 없는 칸은 만들지 않는다.
    #
    # `database_path` 의 부재는 오타가 아니라 **in-memory 를 뜻하는 값**이다
    # (gbrain 의 pglite 엔진이 그렇게 읽는다). 래퍼가 친절하게 채우면 사람이 고른
    # 휘발성 브레인이 말없이 디스크 브레인으로 바뀐다 — 고치려던 것과 같은 종류의
    # 조용한 치환이다. 이 래퍼가 소유한 것은 "어긋난 경로를 제자리로"이지
    # "경로를 정하는 것"이 아니다.
    def test_a_config_without_a_brain_path_is_left_alone(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            script, _, env = self.make_project(temporary)
            home = script.parents[2] / ".sdd" / "gbrain-home" / ".gbrain"
            home.mkdir(parents=True)
            original = json.dumps({"engine": "pglite"}, indent=2) + "\n"
            (home / "config.json").write_text(original, encoding="utf-8")
            completed = subprocess.run(
                [sys.executable, str(script), "search", "anything"],
                env=env,
                capture_output=True,
                text=True,
                timeout=5,
                check=False,
            )
            self.assertEqual(completed.returncode, 0, completed.stderr)
            self.assertEqual(
                (home / "config.json").read_text(encoding="utf-8"), original,
                "래퍼가 사람이 고른 in-memory 구성에 디스크 경로를 지어 넣었다",
            )


if __name__ == "__main__":
    unittest.main()
