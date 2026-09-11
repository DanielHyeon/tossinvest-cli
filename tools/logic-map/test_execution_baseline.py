from __future__ import annotations

import subprocess
import tempfile
import unittest
import hashlib
import json
import shutil
from pathlib import Path
from unittest import mock

import execution_baseline as adoption


class ExecutionBaselineUnitTests(unittest.TestCase):
    def _valid_adoption(self):
        raw, root = self._fixture(); change = root / "openspec" / "changes" / adoption.CHANGE
        change.mkdir(parents=True); (change / "base-commit.txt").write_text("pending\n")
        source = root / "internal" / "soak" / "attest.go"; source.parent.mkdir(parents=True)
        source.write_text("package soak\nfunc Attest() int { return 1 }\n")
        p = self._commit(root, "P")
        (change / "base-commit.txt").write_text(p + "\n"); source.write_text("package soak\nfunc Attest() int { return 2 }\n")
        e = self._commit(root, "E"); source.write_text("package soak\nfunc Attest() int { return 3 }\n")
        s = self._commit(root, "S")
        with mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            ledger_path = change / "analysis" / "execution-baseline-ledger.json"; record_path = change / "execution-baseline.json"
            adoption.draft(root, adoption.CHANGE, s, ledger_path, record_path)
            adversary = change / "analysis" / "adversary.md"; gstack = change / "analysis" / "gstack.md"
            adversary.write_text("adversary"); gstack.write_text("gstack")
            record = json.loads(record_path.read_text())
            record.update({"adversarial_review_path": adversary.relative_to(root).as_posix(), "adversarial_review_sha256": hashlib.sha256(adversary.read_bytes()).hexdigest(), "gstack_review_path": gstack.relative_to(root).as_posix(), "gstack_review_sha256": hashlib.sha256(gstack.read_bytes()).hexdigest()})
            record_path.write_bytes(adoption.canonical(record))
        self._commit(root, "H"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
        return raw, root, change, p, e

    def test_valid_adoption_then_only_inventory_mutation_fails_inventory_error(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertEqual(adoption.validate(change, root, p, change.name)["effective_base"], e)
            # A committed H mutation is required so the record remains byte-bound.
            ledger = json.loads((change / "analysis/execution-baseline-ledger.json").read_text())
            ledger["execution_to_source"]["function_inventory"] = []
            ledger_raw = adoption.canonical(ledger); (change / "analysis/execution-baseline-ledger.json").write_bytes(ledger_raw)
            record = json.loads((change / "execution-baseline.json").read_text()); record["ledger_sha256"] = hashlib.sha256(ledger_raw).hexdigest(); (change / "execution-baseline.json").write_bytes(adoption.canonical(record))
            self._commit(root, "bad inventory H"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            with self.assertRaisesRegex(adoption.AdoptionError, "execution-to-source function inventory mismatch"):
                adoption.validate(change, root, p, change.name)

    def test_valid_fixture_evidence_binding_mutations_have_specific_errors(self) -> None:
        for field, value, error in (
            ("gstack_review_path", "openspec/changes/a063-align-attestation-renewal-profile/analysis/adversary.md", "review paths must be distinct"),
            ("ledger_path", "outside.json", "outside current change analysis"),
            ("inherited_history_disposition", "completed", "disposition is invalid"),
        ):
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                record_path = change / "execution-baseline.json"; record = json.loads(record_path.read_text()); record[field] = value
                if field == "gstack_review_path": record["gstack_review_sha256"] = record["adversarial_review_sha256"]
                record_path.write_bytes(adoption.canonical(record)); self._commit(root, "mutate " + field); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
                with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.validate(change, root, p, change.name)

    def test_valid_fixture_record_schema_and_envelope_mutations_fail_closed(self) -> None:
        cases = (
            ("schema", True, "schema must be exact integer 1"),
            ("schema", "1", "schema must be exact integer 1"),
            ("ledger_sha256", 1, "record field types are invalid"),
            ("planning_base", "0" * 40, "invalid execution-baseline record"),
            ("execution_base", "0" * 40, "invalid execution-baseline record"),
            ("source_commit", "not-a-commit", "commit must be a full lowercase SHA-1"),
            ("source_tree", "0" * 40, "source tree mismatch"),
            ("gstack_review_sha256", "0" * 64, "gstack_review digest mismatch"),
        )
        for field, value, error in cases:
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                record_path = change / "execution-baseline.json"
                record = json.loads(record_path.read_text()); record[field] = value
                record_path.write_bytes(adoption.canonical(record)); self._commit(root, "mutate " + field)
                subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
                with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.validate(change, root, p, change.name)

    def test_valid_fixture_unknown_and_wrong_ledger_schema_fail_closed(self) -> None:
        for mutate, error in (
            (lambda ledger: ledger.update({"unexpected": True}), "invalid execution-baseline ledger"),
            (lambda ledger: ledger.update({"schema": True}), "schema must be exact integer 1"),
            (lambda ledger: ledger.update({"inherited_history_disposition": "completed"}), "ledger inherited history disposition is invalid"),
        ):
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                ledger_path = change / "analysis/execution-baseline-ledger.json"
                ledger = json.loads(ledger_path.read_text()); mutate(ledger)
                ledger_raw = adoption.canonical(ledger); ledger_path.write_bytes(ledger_raw)
                record_path = change / "execution-baseline.json"; record = json.loads(record_path.read_text())
                record["ledger_sha256"] = hashlib.sha256(ledger_raw).hexdigest(); record_path.write_bytes(adoption.canonical(record))
                self._commit(root, "mutate ledger"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
                with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.validate(change, root, p, change.name)

    def test_valid_fixture_base_head_and_source_mutations_have_specific_errors(self) -> None:
        cases = (
            ("base-full", "base", "0" * 40 + "\n", "planning base file is substituted"),
            ("base-abbreviated", "base", "abc123\n", "planning base file is substituted"),
            ("source-moved", "record", ("source_commit", "execution_base"), "source tree mismatch"),
        )
        for _, kind, value, error in cases:
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                if kind == "base":
                    (change / "base-commit.txt").write_text(value, encoding="ascii")
                else:
                    record_path = change / "execution-baseline.json"; record = json.loads(record_path.read_text())
                    record[value[0]] = record[value[1]]
                    record_path.write_bytes(adoption.canonical(record))
                self._commit(root, "mutate source envelope"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
                with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.validate(change, root, p, change.name)

    def test_valid_fixture_refuses_symlink_base_attached_head_and_dirty_tracked_source(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            base = change / "base-commit.txt"; base.unlink(); base.symlink_to("execution-baseline.json")
            with self.assertRaisesRegex(adoption.AdoptionError, "evidence path contains symlink"):
                adoption.validate(change, root, p, change.name)
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            subprocess.run(["git", "checkout", "-q", "master"], cwd=root, check=True)
            with self.assertRaisesRegex(adoption.AdoptionError, "requires detached HEAD"):
                adoption.validate(change, root, p, change.name)
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            (root / "internal/soak/attest.go").write_text("package soak\nfunc Attest() int { return 99 }\n")
            with self.assertRaisesRegex(adoption.AdoptionError, "Go worktree bytes mismatch|evidence differs from HEAD"):
                adoption.validate(change, root, p, change.name)

    def test_source_go_lock_rejects_mode_change_when_git_ignores_filemode(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            subprocess.run(["git", "config", "core.filemode", "false"], cwd=root, check=True)
            target = root / "internal/soak/attest.go"; target.chmod(0o755)
            self.assertEqual(subprocess.run(["git", "diff", "--quiet"], cwd=root).returncode, 0)
            with self.assertRaisesRegex(adoption.AdoptionError, "Go worktree mode mismatch"):
                adoption.validate(change, root, p, change.name)

    def test_source_go_lock_rejects_source_tree_and_go_symlink_substitution(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            target = root / "internal/soak/attest.go"; target.unlink(); target.symlink_to(root / "go.mod")
            with self.assertRaisesRegex(adoption.AdoptionError, "evidence path contains symlink"):
                adoption.validate(change, root, p, change.name)
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            record_path = change / "execution-baseline.json"; record = json.loads(record_path.read_text())
            record["source_tree"] = "f" * 40; record_path.write_bytes(adoption.canonical(record))
            self._commit(root, "source tree substitution"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            with self.assertRaisesRegex(adoption.AdoptionError, "source tree mismatch"):
                adoption.validate(change, root, p, change.name)

    def test_tracked_source_go_symlink_and_non_utf8_git_path_fail_closed(self) -> None:
        raw, root = self._fixture()
        with raw:
            target = root / "internal" / "soak" / "attest.go"; target.parent.mkdir(parents=True)
            target.symlink_to(root / "go.mod"); commit = self._commit(root, "symlink source")
            with self.assertRaisesRegex(adoption.AdoptionError, "non-regular Go file"):
                adoption.go_tree_entries(root, commit)
        raw, root = self._fixture()
        with raw, mock.patch.object(adoption, "_git", return_value=subprocess.CompletedProcess([], 0, b"bad\xff\0", b"")):
            with self.assertRaisesRegex(adoption.AdoptionError, "filename is not UTF-8"):
                adoption.nul_paths(root, "diff", "--name-only")

    def test_filesystem_scan_metadata_and_source_mutations(self) -> None:
        cases = ((".codegraph/cache.json", b"{}", None), (".sdd/index-state.json", b"{}", None), (".sdd/gbrain-home/state.json", b"{}", None), (".sdd/history/checkpoints/one", b"{}", None), (".sdd/history/refresh-queue/one", b"{}", None), (".sdd/history/index-refresh.lock", b"{}", None), (".sdd/history/events/one.jsonl", b"{}", None), ("tools/x/__pycache__/foo.pyc", b"x", None), ("auth-helper/x/__pycache__/foo.pyc", b"x", None), ("ordinary.c", b"int x;", "untracked/ignored input is not allowed"), ("ordinary.s", b"", "untracked/ignored input is not allowed"), ("ordinary.h", b"", "untracked/ignored input is not allowed"), ("ordinary.go", b"package ordinary", "untracked/ignored input is not allowed"), (".codegraph/link", None, "untracked/ignored input is not allowed"), (".codegraph/run", b"x", "untracked/ignored input is not allowed"), ("tools/x/__pycache__/nested/foo.pyc", b"x", "untracked/ignored input is not allowed"))
        for relative, content, error in cases:
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                path=root/relative; path.parent.mkdir(parents=True, exist_ok=True)
                if content is None: path.symlink_to(root / "go.mod")
                else:
                    path.write_bytes(content)
                    if relative.endswith("/run"): path.chmod(0o755)
                if error is None: self.assertEqual(adoption.validate(change, root, p, change.name)["effective_base"], e)
                else:
                    with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.validate(change, root, p, change.name)

    def test_ignored_build_inputs_are_not_hidden_by_git_ignore_rules(self) -> None:
        for relative in ("ignored.c", "ignored.s", "ignored.h", "ignored.go", "embed.txt"):
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                (root / ".git/info/exclude").write_text(relative + "\n", encoding="utf-8")
                (root / relative).write_text("x", encoding="utf-8")
                with self.assertRaisesRegex(adoption.AdoptionError, "untracked/ignored input is not allowed"):
                    adoption.validate(change, root, p, change.name)

    def test_go_input_enumeration_accepts_documented_shapes_and_rejects_bad_output(self) -> None:
        raw, root = self._fixture()
        with raw:
            output = json.dumps({"Dir": str(root), "GoFiles": ["x.go"], "SFiles": ["x.s"], "EmbedFiles": ["asset.txt"]})
            ok = subprocess.CompletedProcess([], 0, output, "")
            with mock.patch.object(adoption.subprocess, "run", return_value=ok) as run:
                self.assertEqual(adoption.go_inputs(root), {"x.go", "x.s", "asset.txt"})
                self.assertEqual(adoption.go_inputs(root, "tossos_testseams"), {"x.go", "x.s", "asset.txt"})
                self.assertIn("tossos_testseams", run.call_args_list[-1].args[0])
            for output, error in (("{", "invalid go package enumeration JSON"), ("[]", "entry is not an object"), (json.dumps({"Dir": str(root), "GoFiles": "x.go"}), "GoFiles is malformed"), (json.dumps({"Dir": str(root), "EmbedFiles": ["../escape"]}), "path escapes package")):
                with mock.patch.object(adoption.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, output, "")):
                    with self.assertRaisesRegex(adoption.AdoptionError, error): adoption.go_inputs(root)
            with mock.patch.object(adoption.subprocess, "run", return_value=subprocess.CompletedProcess([], 1, "", "go failed")):
                with self.assertRaisesRegex(adoption.AdoptionError, "go failed"): adoption.go_inputs(root)

    def test_allowed_metadata_becomes_fatal_when_go_list_reports_it_as_input(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            path = root / ".codegraph/cache.json"; path.parent.mkdir(); path.write_text("{}", encoding="utf-8")
            with mock.patch.object(adoption, "go_inputs", return_value={".codegraph/cache.json"}):
                with self.assertRaisesRegex(adoption.AdoptionError, "untracked/ignored Go build input"):
                    adoption.validate(change, root, p, change.name)

    def test_validate_unions_default_and_seams_inputs_and_blocks_each_profile_failure(self) -> None:
        raw, root, change, p, e = self._valid_adoption()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            path = root / ".codegraph/cache.json"; path.parent.mkdir(); path.write_text("{}")
            with mock.patch.object(adoption, "go_inputs", side_effect=(set(), {".codegraph/cache.json"})):
                with self.assertRaisesRegex(adoption.AdoptionError, "Go build input"):
                    adoption.validate(change, root, p, change.name)
        for failure in (adoption.AdoptionError("default failed"), adoption.AdoptionError("seams failed")):
            raw, root, change, p, e = self._valid_adoption()
            with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                side = (failure, set()) if "default" in str(failure) else (set(), failure)
                with mock.patch.object(adoption, "go_inputs", side_effect=side):
                    with self.assertRaisesRegex(adoption.AdoptionError, str(failure)):
                        adoption.validate(change, root, p, change.name)
    def _fixture(self) -> tuple[tempfile.TemporaryDirectory, Path]:
        raw = tempfile.TemporaryDirectory(); root = Path(raw.name)
        subprocess.run(["git", "init", "-q", "-b", "master"], cwd=root, check=True)
        subprocess.run(["git", "config", "user.email", "a120@example.invalid"], cwd=root, check=True)
        subprocess.run(["git", "config", "user.name", "a120"], cwd=root, check=True)
        (root / "go.mod").write_text("module fixture\ngo 1.23\n", encoding="utf-8")
        source = Path(__file__).resolve().parent
        shutil.copytree(source, root / "tools" / "logic-map", ignore=shutil.ignore_patterns("*.py", "__pycache__"))
        return raw, root

    @staticmethod
    def _commit(root: Path, subject: str) -> str:
        subprocess.run(["git", "add", "."], cwd=root, check=True)
        subprocess.run(["git", "commit", "-qm", subject], cwd=root, check=True)
        return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()

    def test_real_git_inventory_detects_missing_duplicate_hash_and_revision_rows(self) -> None:
        raw, root = self._fixture()
        with raw:
            target = root / "pkg" / "x.go"; target.parent.mkdir()
            target.write_text("package pkg\nfunc Keep() int { return 1 }\n", encoding="utf-8")
            p = self._commit(root, "P")
            target.write_text("package pkg\nfunc Keep() int { return 2 }\n", encoding="utf-8")
            e = self._commit(root, "E")
            rows = adoption.function_inventory(root, p, e)
            self.assertEqual(len(rows), 1); self.assertEqual(rows[0]["revision"], "current")
            with self.assertRaisesRegex(adoption.AdoptionError, "differs from immutable"): adoption.exact_inventory([], rows)
            with self.assertRaisesRegex(adoption.AdoptionError, "duplicate rows"): adoption.exact_inventory(rows + rows, rows)
            wrong = [dict(rows[0], source_sha256="0" * 64)]
            with self.assertRaisesRegex(adoption.AdoptionError, "differs from immutable"): adoption.exact_inventory(wrong, rows)
            wrong = [dict(rows[0], revision="base")]
            with self.assertRaisesRegex(adoption.AdoptionError, "differs from immutable"): adoption.exact_inventory(wrong, rows)

    def test_real_git_deleted_existing_function_uses_base_revision(self) -> None:
        raw, root = self._fixture()
        with raw:
            target = root / "pkg" / "x.go"; target.parent.mkdir()
            target.write_text("package pkg\nfunc Gone() int { return 1 }\n", encoding="utf-8")
            e = self._commit(root, "E")
            target.write_text("package pkg\n", encoding="utf-8")
            s = self._commit(root, "S")
            rows = adoption.function_inventory(root, e, s)
            self.assertEqual(rows[0]["function"], "Gone")
            self.assertEqual(rows[0]["revision"], "base")

    def test_real_git_history_keeps_merge_parent_and_reverted_path(self) -> None:
        raw, root = self._fixture()
        with raw:
            (root / "base.txt").write_text("base", encoding="utf-8"); p = self._commit(root, "P")
            subprocess.run(["git", "checkout", "-qb", "side"], cwd=root, check=True)
            (root / "reverted.txt").write_text("side", encoding="utf-8"); self._commit(root, "side")
            subprocess.run(["git", "checkout", "-q", "master"], cwd=root, check=True)
            (root / "main.txt").write_text("main", encoding="utf-8"); self._commit(root, "main")
            subprocess.run(["git", "merge", "--no-ff", "side", "-qm", "merge"], cwd=root, check=True)
            (root / "reverted.txt").unlink(); e = self._commit(root, "revert path")
            history = adoption.history(root, p, e)
            flattened = [change["path"] for row in history for parent in row["parent_diffs"] for change in parent["changes"]]
            self.assertIn("reverted.txt", flattened)
            self.assertTrue(any(len(row["parents"]) == 2 for row in history))

    def test_generator_refuses_overwrite_after_real_git_draft(self) -> None:
        raw, root = self._fixture()
        with raw:
            (root / "pkg").mkdir(); (root / "pkg" / "x.go").write_text("package pkg\nfunc X() {}\n", encoding="utf-8")
            p = self._commit(root, "P")
            (root / "pkg" / "x.go").write_text("package pkg\nfunc X() { println(1) }\n", encoding="utf-8")
            e = self._commit(root, "E")
            s = e
            change = root / "openspec" / "changes" / adoption.CHANGE
            out = change / "analysis" / "execution-baseline-ledger.json"; record = change / "execution-baseline.json"
            with mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                adoption.draft(root, adoption.CHANGE, s, out, record)
                self.assertTrue(out.is_file()); self.assertTrue(record.is_file())
                with self.assertRaises(adoption.AdoptionError): adoption.draft(root, adoption.CHANGE, s, out, record)

    def test_generator_rejects_outside_or_symlinked_outputs_before_writing(self) -> None:
        raw, root = self._fixture()
        with raw:
            (root / "marker").write_text("P"); p = self._commit(root, "P")
            (root / "marker").write_text("E"); e = self._commit(root, "E")
            change = root / "openspec" / "changes" / adoption.CHANGE
            ledger = change / "analysis" / "execution-baseline-ledger.json"; record = change / "execution-baseline.json"
            with mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                with self.assertRaisesRegex(adoption.AdoptionError, "fixed change record"):
                    adoption.draft(root, adoption.CHANGE, e, root / "outside.json", record)
                self.assertFalse((root / "outside.json").exists())
                change.mkdir(parents=True); (change / "analysis").symlink_to(root)
                with self.assertRaisesRegex(adoption.AdoptionError, "parent is unsafe"):
                    adoption.draft(root, adoption.CHANGE, e, ledger, record)
                self.assertFalse(record.exists())

    def test_generator_preserves_existing_broken_target_and_cleans_only_its_partial_ledger(self) -> None:
        raw, root = self._fixture()
        with raw:
            (root / "marker").write_text("P"); p = self._commit(root, "P")
            (root / "marker").write_text("E"); e = self._commit(root, "E")
            change = root / "openspec" / "changes" / adoption.CHANGE
            ledger = change / "analysis" / "execution-baseline-ledger.json"; record = change / "execution-baseline.json"
            with mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
                change.mkdir(parents=True); record.symlink_to("missing-record.json")
                with self.assertRaisesRegex(adoption.AdoptionError, "already exists"):
                    adoption.draft(root, adoption.CHANGE, e, ledger, record)
                self.assertTrue(record.is_symlink()); self.assertFalse(ledger.exists())
                record.unlink()
                original = adoption._exclusive_write
                def fail_record(path: Path, contents: bytes) -> None:
                    if path == record:
                        raise OSError("record failure")
                    original(path, contents)
                with mock.patch.object(adoption, "_exclusive_write", side_effect=fail_record):
                    with self.assertRaisesRegex(OSError, "record failure"):
                        adoption.draft(root, adoption.CHANGE, e, ledger, record)
                self.assertFalse(ledger.exists()); self.assertFalse(record.exists())
    def test_canonical_is_key_sorted_compact_ascii(self) -> None:
        self.assertEqual(adoption.canonical({"z": "한", "a": 1}), b'{"a":1,"z":"\\ud55c"}')

    def test_duplicate_json_key_fails_closed(self) -> None:
        with self.assertRaises(adoption.AdoptionError):
            adoption.strict_json(b'{"schema":1,"schema":1}')

    def test_non_full_commit_is_rejected_without_git(self) -> None:
        with self.assertRaises(adoption.AdoptionError):
            adoption.full_commit(Path("/unused"), "HEAD")

    def test_strict_source_before_head_rejects_source_equal_to_head(self) -> None:
        raw, root = self._fixture()
        with raw:
            commit = self._commit(root, "only commit")
            with self.assertRaisesRegex(adoption.AdoptionError, "required commit ancestry is absent"):
                adoption.ancestry(root, commit, commit, strict=True)

    def test_history_preserves_each_parent_and_sorts_paths(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            subprocess.run(["git", "init", "-q"], cwd=root, check=True)
            subprocess.run(["git", "config", "user.email", "a120@example.invalid"], cwd=root, check=True)
            subprocess.run(["git", "config", "user.name", "a120"], cwd=root, check=True)
            (root / "z.txt").write_text("0", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=root, check=True)
            subprocess.run(["git", "commit", "-qm", "P"], cwd=root, check=True)
            p = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
            (root / "z.txt").write_text("1", encoding="utf-8")
            (root / "a.txt").write_text("1", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=root, check=True)
            subprocess.run(["git", "commit", "-qm", "E"], cwd=root, check=True)
            e = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
            rows = adoption.history(root, p, e)
            self.assertEqual(len(rows), 1)
            self.assertEqual(rows[0]["commit"], e)
            changes = rows[0]["parent_diffs"][0]["changes"]
            self.assertEqual([row["path"] for row in changes], ["a.txt", "z.txt"])

    def test_absent_record_keeps_ordinary_behavior(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            change = root / "openspec" / "changes" / "ordinary"
            change.mkdir(parents=True)
            self.assertIsNone(adoption.validate(change, root, "a" * 40, change.name))

    def test_present_record_for_wrong_change_fails_before_git(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            change = root / "openspec" / "changes" / "ordinary"
            change.mkdir(parents=True)
            (change / "execution-baseline.json").write_text("{}", encoding="utf-8")
            with self.assertRaises(adoption.AdoptionError):
                adoption.validate(change, root, adoption.P, change.name)


if __name__ == "__main__":
    unittest.main()
