"""원문 복사 시험 — 머리 주석 1줄을 뺀 바이트의 sha256 이 StockOS 기록 커밋의 원본과 같음."""

from __future__ import annotations

import hashlib
import json
import subprocess
import unittest
from pathlib import Path

PORT = Path(__file__).resolve().parent.parent / "stockos_port"
MANIFEST = json.loads((PORT / "MANIFEST.json").read_text(encoding="utf-8"))


def body_without_header(path: Path) -> bytes:
    raw = path.read_bytes()
    header, _, rest = raw.partition(b"\n")
    return header, rest


class PortTest(unittest.TestCase):
    def test_every_copy_matches_manifest(self):
        for rel, sha in MANIFEST["files"].items():
            with self.subTest(rel=rel):
                header, body = body_without_header(PORT / rel)
                self.assertEqual(hashlib.sha256(body).hexdigest(), sha)
                expected = (f"# stockos-port: source={MANIFEST['source_root']}/{rel} "
                            f"commit={MANIFEST['source_commit']} — verbatim copy, do not edit").encode()
                self.assertEqual(header, expected)

    def test_no_other_python_in_port(self):
        own = {"__init__.py", "exit/__init__.py", "ladder/__init__.py"}
        found = {p.relative_to(PORT).as_posix() for p in PORT.rglob("*.py")}
        self.assertEqual(found, set(MANIFEST["files"]) | own)
        for rel in own:
            self.assertLess(len((PORT / rel).read_text(encoding="utf-8").splitlines()), 3)

    def test_manifest_matches_stockos_repo_when_present(self):
        repo = Path(MANIFEST["source_repo"])
        if not (repo / ".git").exists():
            self.skipTest("StockOS 저장소 없음 — MANIFEST 대조만 함")
        for rel, sha in MANIFEST["files"].items():
            with self.subTest(rel=rel):
                blob = subprocess.run(
                    ["git", "-C", str(repo), "show", f"{MANIFEST['source_commit']}:{MANIFEST['source_root']}/{rel}"],
                    check=True, capture_output=True, timeout=60,
                ).stdout
                self.assertEqual(hashlib.sha256(blob).hexdigest(), sha)


if __name__ == "__main__":
    unittest.main()
