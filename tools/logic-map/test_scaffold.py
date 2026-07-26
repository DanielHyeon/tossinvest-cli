import tempfile
import unittest
from pathlib import Path

from scaffold_analysis import artifact_dir, scaffold


class ScaffoldTests(unittest.TestCase):
    def test_scaffold_is_idempotent_and_scoped_to_change(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            out = scaffold("add-example", "internal/foo/bar.go", "Runner.Run", root)
            self.assertEqual(
                out,
                artifact_dir("add-example", "internal/foo/bar.go", "Runner.Run", root),
            )
            expected = out / "function-logic-map.md"
            self.assertTrue(expected.exists())
            self.assertTrue((out / "ast.json").exists())
            expected.write_text("custom", encoding="utf-8")
            scaffold("add-example", "internal/foo/bar.go", "Runner.Run", root)
            self.assertEqual(expected.read_text(encoding="utf-8"), "custom")

    def test_change_id_cannot_escape_openspec_directory(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            for value in ("../escape", "/tmp/escape", "UPPER"):
                with self.assertRaises(ValueError):
                    artifact_dir(value, "internal/foo.go", "Run", root)


if __name__ == "__main__":
    unittest.main()
