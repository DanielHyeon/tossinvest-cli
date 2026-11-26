"""정적 안전 시험 — 도구 소스는 표준 라이브러리의 허용 목록만 import 하고 브로커 · 네트워크 · 자격 증명 흔적이 없음."""

from __future__ import annotations

import ast
import unittest
from pathlib import Path

TOOL = Path(__file__).resolve().parent.parent
ALLOWED = {
    "__future__", "argparse", "collections", "dataclasses", "datetime", "enum", "hashlib", "json", "math",
    "pathlib", "re", "statistics", "sys", "typing",
    # 도구 자체 모듈과 복사본
    "aee_exits", "c1_entry", "sim_config", "verdict", "stockos_port",
}
FORBIDDEN_TEXT = ("tossctl", ".env", "socket", "urllib", "http.client", "requests", "subprocess", ".config/")


def tool_sources():
    return [p for p in TOOL.rglob("*.py") if "tests" not in p.relative_to(TOOL).parts]


class StaticTest(unittest.TestCase):
    def test_imports_are_allowlisted(self):
        for path in tool_sources():
            tree = ast.parse(path.read_text(encoding="utf-8"))
            for node in ast.walk(tree):
                if isinstance(node, ast.Import):
                    names = [a.name for a in node.names]
                elif isinstance(node, ast.ImportFrom) and node.level == 0:
                    names = [node.module]
                else:
                    continue
                for name in names:
                    with self.subTest(path=path.name, name=name):
                        self.assertIn(name.split(".")[0], ALLOWED)

    def test_no_forbidden_text_in_own_code(self):
        for path in tool_sources():
            if "stockos_port" in path.parts:
                continue  # 원문 복사본(docstring 에 `.env` 언급) — import 시험이 대신 묶음
            text = path.read_text(encoding="utf-8")
            for word in FORBIDDEN_TEXT:
                with self.subTest(path=path.name, word=word):
                    self.assertNotIn(word, text)


if __name__ == "__main__":
    unittest.main()
