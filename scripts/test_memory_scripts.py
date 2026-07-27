from __future__ import annotations

import os
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent


class MemoryScriptTests(unittest.TestCase):
    def test_recall_routes_gbrain_through_project_wrapper(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            log = directory / "calls.log"
            fake_python = directory / "python"
            fake_python.write_text(
                '#!/usr/bin/env sh\nprintf "%s\\n" "$*" >> "$CALL_LOG"\n',
                encoding="utf-8",
            )
            fake_gbrain = directory / "gbrain"
            fake_gbrain.write_text("#!/usr/bin/env sh\nexit 0\n", encoding="utf-8")
            fake_python.chmod(0o700)
            fake_gbrain.chmod(0o700)
            environment = os.environ.copy()
            environment.update(
                {
                    "PYTHON": str(fake_python),
                    "CALL_LOG": str(log),
                    "PATH": f"{directory}:{environment['PATH']}",
                }
            )
            subprocess.run(
                ["bash", str(ROOT / "scripts" / "memory-recall.sh"), "query"],
                cwd=ROOT,
                env=environment,
                check=True,
                capture_output=True,
                text=True,
            )
            calls = log.read_text(encoding="utf-8")
            self.assertIn("tools/sdd/gbrain_project.py search query", calls)


if __name__ == "__main__":
    unittest.main()
