from __future__ import annotations

import json
import hashlib
import io
import os
import subprocess
import sys
import tempfile
import unittest
import shutil
from contextlib import redirect_stdout
from pathlib import Path
from unittest import mock

import check_analysis
import execution_baseline as adoption

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "sdd"))
import sdd_doctor


def write_bundle(
    root: Path,
    *,
    branches: list | None,
    logic_source_line: str = "- Source: `internal/sample.go`",
    branch_source_line: str = "",
    branch_rows: str = "| B1 | leaf | test | yes | yes |",
) -> Path:
    """Write a minimal four-file bundle whose only variable is what it claims."""
    source = root / "internal" / "sample.go"
    source.parent.mkdir(parents=True, exist_ok=True)
    source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
    target = (
        root / "openspec" / "changes" / "change" / "analysis" / "function-logic" / "pkg--run"
    )
    target.mkdir(parents=True, exist_ok=True)
    (target / "ast.json").write_text(
        json.dumps(
            {
                "file": "internal/sample.go",
                "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                "package": "sample",
                "function": "Run",
                "signature": "Run(params=0, results=0)",
                "start": {"line": 2, "column": 1},
                "end": {"line": 2, "column": 14},
                "branches": branches,
            }
        ),
        encoding="utf-8",
    )
    (target / "function-logic-map.md").write_text(
        f"""# Function Logic Map: `Run`
{logic_source_line}
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
        encoding="utf-8",
    )
    (target / "branch-test-map.md").write_text(
        f"# Branch Test Map: `Run`\n{branch_source_line}\n{branch_rows}\n",
        encoding="utf-8",
    )
    (target / "risk-pattern-report.md").write_text(
        "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
        encoding="utf-8",
    )
    return target


def run_check(root: Path) -> list[str]:
    with mock.patch(
        "check_analysis.resolve_base",
        return_value="base",
    ), mock.patch(
        "check_analysis.changed_existing_functions",
        return_value={},
    ):
        return check_analysis.check("change", root)


class BundleTextCoversEveryProseFileInTheBundle(unittest.TestCase):
    """강제 판정이 읽는 범위가 **열거된 파일 목록**이면 그 목록 밖으로 옮기면 꺼진다.

    8차 적대 리뷰가 `risk-pattern-report.md` 로 그것을 보였다 — 번들 필수 파일이고
    표는 머리글도 행도 그대로 살아 있는데 감사만 꺼졌다. 네 라운드 동안 "남는
    회피는 X 뿐"이라고 적은 문장이 매번 코드가 보는 범위보다 한 칸 넓었다."""

    def test_every_md_in_the_bundle_directory_is_read(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            bundle = Path(raw) / "b"
            bundle.mkdir()
            # 확장자로 거르면 `.txt`·`.mdx` 로 빠져나간다(9차 적대 리뷰).
            names = ("function-logic-map.md", "branch-test-map.md",
                     "risk-pattern-report.md", "someone-elses-note.md",
                     "notes.txt", "notes.mdx", "notes.markdown")
            for index, name in enumerate(names):
                (bundle / name).write_text(f"mark-{index}", encoding="utf-8")
            (bundle / "ast.json").write_text("{}", encoding="utf-8")
            text = check_analysis._bundle_text(bundle / "function-logic-map.md")
        for index in range(len(names)):
            self.assertIn(f"mark-{index}", text)


class CheckAnalysisTests(unittest.TestCase):
    def test_main_distinguishes_ordinary_adoption_and_invalid_results(self) -> None:
        def run_cli(errors: list[str], adopted: bool) -> tuple[int, str]:
            output = io.StringIO()
            def fake_check(change: str, root: Path, context: dict[str, object]) -> list[str]:
                context["execution_baseline_adoption"] = adopted
                return errors
            with mock.patch.object(sys, "argv", ["check_analysis.py", "--change", "fixture"]), \
                 mock.patch("check_analysis.check", side_effect=fake_check), \
                 redirect_stdout(output):
                status = check_analysis.main()
            return status, output.getvalue()

        ordinary_status, ordinary_output = run_cli([], False)
        self.assertEqual(ordinary_status, 0)
        self.assertIn("evidence complete or diff-proven exempt", ordinary_output)
        adoption_status, adoption_output = run_cli([], True)
        self.assertEqual(adoption_status, 0)
        self.assertIn("execution-baseline adoption exception evidence complete", adoption_output)
        invalid_status, invalid_output = run_cli(["invalid execution-baseline adoption: bad record"], False)
        self.assertEqual(invalid_status, 1)
        self.assertIn("invalid execution-baseline adoption: bad record", invalid_output)
        self.assertNotIn("adoption exception evidence complete", invalid_output)

    def test_main_real_adoption_prints_exception_label_only_after_validation(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            output = io.StringIO()
            with mock.patch.object(
                sys, "argv", ["check_analysis.py", "--change", adoption.CHANGE, "--root", str(root)]
            ), redirect_stdout(output):
                self.assertEqual(check_analysis.main(), 0)
            self.assertIn("execution-baseline adoption exception evidence complete", output.getvalue())
            record = root / "openspec" / "changes" / adoption.CHANGE / "execution-baseline.json"
            record.write_text("{}")
            self._commit(root, "invalid record"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            output = io.StringIO()
            with mock.patch.object(
                sys, "argv", ["check_analysis.py", "--change", adoption.CHANGE, "--root", str(root)]
            ), redirect_stdout(output):
                self.assertEqual(check_analysis.main(), 1)
            self.assertIn("invalid execution-baseline adoption", output.getvalue())
            self.assertNotIn("adoption exception evidence complete", output.getvalue())

    def test_ordinary_no_record_uses_p_and_allows_dirty_worktree(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            subprocess.run(["git", "init", "-q"], cwd=root, check=True)
            for key, value in (("user.email", "a120@example.invalid"), ("user.name", "a120")):
                subprocess.run(["git", "config", key, value], cwd=root, check=True)
            (root / "go.mod").write_text("module fixture\ngo 1.23\n")
            # `check()` invokes the real Go AST extractor from the fixture
            # checkout.  Keep this as a copied, committed tool rather than a
            # mock so the missing-map and green assertions exercise the same
            # P-to-dirty-worktree path as production.
            tool_source = Path(__file__).resolve().parent
            shutil.copytree(
                tool_source,
                root / "tools" / "logic-map",
                ignore=shutil.ignore_patterns("*.py", "__pycache__"),
            )
            source = root / "internal" / "ordinary.go"; source.parent.mkdir()
            source.write_text("package internal\nfunc Run() int { return 1 }\n")
            (root / "marker").write_text("P")
            p = self._commit(root, "P")
            change = root / "openspec" / "changes" / "ordinary"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(p + "\n")
            (change / "review.md").write_text("Function Logic Map: not-applicable\n")
            self._commit(root, "ordinary evidence")
            source.write_text("package internal\nfunc Run() int { return 2 }\n")
            self.assertEqual(check_analysis.resolve_base(change, root, change_id=change.name), p)
            errors = check_analysis.check("ordinary", root)
            self.assertTrue(any("missing Function Logic Map" in error or "missing evidence for modified function" in error for error in errors), errors)
            bundle = change / "analysis" / "function-logic" / "internal--run"; bundle.mkdir(parents=True)
            source_hash = hashlib.sha256(source.read_bytes()).hexdigest()
            ast = {"file":"internal/ordinary.go","source_sha256":source_hash,"package":"internal","function":"Run","signature":"Run(params=0, results=1)","start":{"line":2,"column":1},"end":{"line":2,"column":1},"branches":[]}
            (bundle / "ast.json").write_text(json.dumps(ast))
            (bundle / "function-logic-map.md").write_text("# Function Logic Map: `Run`\ninternal/ordinary.go\n## Inputs and invariants\ne\n## Branches and early returns\ne\n## Calls and live bindings\ne\n## State mutations and fallbacks\ne\n## Safety conclusion\ne\n")
            (bundle / "branch-test-map.md").write_text("# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n")
            (bundle / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/ordinary.go\n")
            with mock.patch.dict("os.environ", {"SDD_BASE_REF": p}, clear=False):
                self.assertEqual(check_analysis.check("ordinary", root), [])
            head = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
            for override in ("HEAD", head):
                with mock.patch.dict("os.environ", {"SDD_BASE_REF": override}, clear=False):
                    self.assertEqual(
                        check_analysis.check("ordinary", root),
                        ["cannot derive modified Go functions: SDD_BASE_REF must resolve to the selected effective comparison base"],
                    )

    def test_real_reference_requires_the_same_p_base(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            subprocess.run(["git", "init", "-q"], cwd=root, check=True)
            for key, value in (("user.email", "a120@example.invalid"), ("user.name", "a120")):
                subprocess.run(["git", "config", key, value], cwd=root, check=True)
            (root / "go.mod").write_text("module fixture\ngo 1.23\n")
            source = root / "internal" / "sample.go"; source.parent.mkdir()
            source.write_text("package sample\nfunc Run() {}\n")
            p = self._commit(root, "P")
            (root / "marker").write_text("Q")
            q = self._commit(root, "Q")
            change = root / "openspec" / "changes" / "ordinary"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(p + "\n")
            (change / "analysis" / "function-logic-reference.txt").parent.mkdir(parents=True)
            (change / "analysis" / "function-logic-reference.txt").write_text("reference\n")
            reference = root / "openspec" / "changes" / "reference"
            reference.mkdir(parents=True)
            (reference / "base-commit.txt").write_text(p + "\n")
            bundle = reference / "analysis" / "function-logic" / "internal--run"; bundle.mkdir(parents=True)
            source_hash = hashlib.sha256(source.read_bytes()).hexdigest()
            ast = {"file":"internal/sample.go","source_sha256":source_hash,"package":"sample","function":"Run","signature":"Run(params=0, results=0)","start":{"line":2,"column":1},"end":{"line":2,"column":1},"branches":[]}
            (bundle / "ast.json").write_text(json.dumps(ast))
            (bundle / "function-logic-map.md").write_text("# Function Logic Map: `Run`\ninternal/sample.go\n## Inputs and invariants\ne\n## Branches and early returns\ne\n## Calls and live bindings\ne\n## State mutations and fallbacks\ne\n## Safety conclusion\ne\n")
            (bundle / "branch-test-map.md").write_text("# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n")
            (bundle / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/sample.go\n")
            self.assertEqual(check_analysis.check("ordinary", root), [])
            (reference / "base-commit.txt").write_text(q + "\n")
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["function-logic reference must share the exact comparison base"],
            )
    def _reference_fixture(self, root: Path, reference: Path) -> str:
        """`ordinary` change 하나와, 그 change 가 가리키는 증거 보관처를 만든다.

        증거는 `reference` 디렉터리에 두고 `ordinary` 는 포인터만 갖는다 — a073 이
        a072 의 231개 번들을 그렇게 빌려 쓴다. 반환값은 공유 비교 base 다.
        """
        subprocess.run(["git", "init", "-q"], cwd=root, check=True)
        for key, value in (("user.email", "a073@example.invalid"), ("user.name", "a073")):
            subprocess.run(["git", "config", key, value], cwd=root, check=True)
        (root / "go.mod").write_text("module fixture\ngo 1.23\n")
        source = root / "internal" / "sample.go"; source.parent.mkdir()
        source.write_text("package sample\nfunc Run() {}\n")
        p = self._commit(root, "P")
        change = root / "openspec" / "changes" / "ordinary"
        (change / "analysis").mkdir(parents=True)
        (change / "base-commit.txt").write_text(p + "\n")
        (change / "analysis" / "function-logic-reference.txt").write_text("reference\n")
        reference.mkdir(parents=True, exist_ok=True)
        (reference / "base-commit.txt").write_text(p + "\n")
        bundle = reference / "analysis" / "function-logic" / "internal--run"; bundle.mkdir(parents=True)
        source_hash = hashlib.sha256(source.read_bytes()).hexdigest()
        ast = {"file":"internal/sample.go","source_sha256":source_hash,"package":"sample","function":"Run","signature":"Run(params=0, results=0)","start":{"line":2,"column":1},"end":{"line":2,"column":1},"branches":[]}
        (bundle / "ast.json").write_text(json.dumps(ast))
        (bundle / "function-logic-map.md").write_text("# Function Logic Map: `Run`\ninternal/sample.go\n## Inputs and invariants\ne\n## Branches and early returns\ne\n## Calls and live bindings\ne\n## State mutations and fallbacks\ne\n## Safety conclusion\ne\n")
        (bundle / "branch-test-map.md").write_text("# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n")
        (bundle / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/sample.go\n")
        return p

    def test_real_reference_resolves_an_archived_change(self) -> None:
        """참조된 change 를 아카이브해도 포인터가 고아가 되지 않는다.

        openspec archive 는 change 를 `archive/<YYYY-MM-DD>-<id>` 로 옮긴다. 옮긴
        뒤에도 그 증거를 빌려 쓰는 change 의 게이트는 통과해야 한다 — 안 그러면
        먼저 아카이브한 쪽이 나중 쪽의 완료를 영구히 막는다.
        """
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            archived = root / "openspec" / "changes" / "archive" / "2026-08-29-reference"
            self._reference_fixture(root, archived)
            self.assertEqual(check_analysis.check("ordinary", root), [])

    def test_real_archived_reference_still_requires_the_same_base(self) -> None:
        """아카이브 경로를 정말 읽는지 — base 를 어긋내면 반드시 빨개진다."""
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            archived = root / "openspec" / "changes" / "archive" / "2026-08-29-reference"
            self._reference_fixture(root, archived)
            (root / "marker").write_text("Q")
            q = self._commit(root, "Q")
            (archived / "base-commit.txt").write_text(q + "\n")
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["function-logic reference must share the exact comparison base"],
            )

    def test_real_archived_reference_rejects_a_suffix_collision(self) -> None:
        """`<날짜>-<id>` 를 정확히 맞춘다.

        거절 대상을 이름으로 적는다: `2026-08-29-other-reference` 는 이름이
        `reference` 로 끝날 뿐 다른 change 다. 접미사로 고르면 남의 증거를
        내 증거로 통과시킨다.
        """
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            other = root / "openspec" / "changes" / "archive" / "2026-08-29-other-reference"
            self._reference_fixture(root, other)
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["function-logic reference base is invalid: reference change is neither open nor archived: reference"],
            )

    def test_real_archived_reference_rejects_two_copies(self) -> None:
        """같은 id 의 아카이브가 둘이면 고르지 않고 멈춘다.

        고르면 어느 쪽 증거로 통과했는지 기록에 남지 않는다.
        """
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            archive = root / "openspec" / "changes" / "archive"
            self._reference_fixture(root, archive / "2026-08-29-reference")
            shutil.copytree(archive / "2026-08-29-reference", archive / "2026-09-01-reference")
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["function-logic reference base is invalid: archive holds 2 copies of reference: 2026-08-29-reference, 2026-09-01-reference"],
            )

    def test_real_reference_refuses_when_open_and_archived_collide(self) -> None:
        """빌린 증거의 id 가 활성과 아카이브에 동시에 있으면 고르지 않고 멈춘다.

        오늘은 활성이 조용히 이긴다 — `resolve_referenced_change` 가
        `if direct.is_dir(): return direct` 로 아카이브를 열어 보지도 않기 때문이다
        (`analysis/python-function-logic` 열거의 B1/L239-240). 고르면 어느 증거로
        게이트가 열렸는지 기록에 남지 않는다. 아카이브 **안**의 중복은 이미
        멈추는데 활성+아카이브만 안 멈추는 것은 세는 범위가 좁아서다.
        """
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            openspec = root / "openspec" / "changes"
            self._reference_fixture(root, openspec / "reference")
            self.assertEqual(check_analysis.check("ordinary", root), [], "양성 대조: 충돌 전에는 통과해야 한다")
            shutil.copytree(openspec / "reference", openspec / "archive" / "2026-08-29-reference")
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["function-logic reference base is invalid: reference is open and archived at once: "
                 "openspec/changes/reference, openspec/changes/archive/2026-08-29-reference"],
            )

    def test_real_gate_target_refuses_when_open_and_archived_collide(self) -> None:
        """게이트 **대상 자신**이 활성과 아카이브에 동시에 있어도 멈춘다.

        해소기만 고치면 이 경로는 안 바뀐다. `check` 가 `:689-692` 에서
        `except ValueError` 로 그 실패를 **삼키고** `openspec/changes/<id>` 로
        되돌아가기 때문이다 — 삼킨 결과가 정확히 "활성이 조용히 이긴다"이다.
        그래서 이 시험이 따로 있다.
        """
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            openspec = root / "openspec" / "changes"
            self._reference_fixture(root, openspec / "reference")
            self.assertEqual(check_analysis.check("ordinary", root), [], "양성 대조: 충돌 전에는 통과해야 한다")
            shutil.copytree(openspec / "ordinary", openspec / "archive" / "2026-08-29-ordinary")
            self.assertEqual(
                check_analysis.check("ordinary", root),
                ["ordinary is open and archived at once: "
                 "openspec/changes/ordinary, openspec/changes/archive/2026-08-29-ordinary"],
            )

    @staticmethod
    def _commit(root: Path, subject: str) -> str:
        subprocess.run(["git", "add", "."], cwd=root, check=True)
        subprocess.run(["git", "commit", "-qm", subject], cwd=root, check=True)
        return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()

    def _adoption_with_complete_bundle(self, deleted: bool = False) -> tuple[tempfile.TemporaryDirectory, Path, str, str]:
        raw = tempfile.TemporaryDirectory(); root = Path(raw.name)
        subprocess.run(["git", "init", "-q"], cwd=root, check=True)
        for key, value in (("user.email", "a120@example.invalid"), ("user.name", "a120")):
            subprocess.run(["git", "config", key, value], cwd=root, check=True)
        (root / "go.mod").write_text("module fixture\ngo 1.23\n")
        source = Path(__file__).resolve().parent
        shutil.copytree(source, root / "tools" / "logic-map", ignore=shutil.ignore_patterns("*.py", "__pycache__"))
        requirements = root / "tools" / "sdd" / "requirements.txt"; requirements.parent.mkdir(parents=True)
        requirements.write_text("typedb-driver==3.11.5\n", encoding="utf-8")
        change = root / "openspec" / "changes" / adoption.CHANGE; change.mkdir(parents=True)
        (change / "base-commit.txt").write_text("pending\n")
        target = root / "internal" / "soak" / "attest.go"; target.parent.mkdir(parents=True)
        target.write_text("package soak\nfunc Attest() int { return 1 }\n")
        p = self._commit(root, "P")
        (change / "base-commit.txt").write_text(p + "\n"); target.write_text("package soak\nfunc Attest() int { return 2 }\n")
        e = self._commit(root, "E"); ast = check_analysis.go_functions(target, root)[0]
        target.write_text("package soak\n" if deleted else "package soak\nfunc Attest() int { return 3 }\n")
        s = self._commit(root, "S")
        if not deleted: ast = check_analysis.go_functions(target, root)[0]
        with mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            ledger = change / "analysis" / "execution-baseline-ledger.json"; record = change / "execution-baseline.json"
            adoption.draft(root, adoption.CHANGE, s, ledger, record)
            for name, body in (("adversary.md", "adversary"), ("gstack.md", "gstack")):
                path = change / "analysis" / name; path.write_text(body)
            value = json.loads(record.read_text())
            for prefix, name in (("adversarial_review", "adversary.md"), ("gstack_review", "gstack.md")):
                path = change / "analysis" / name
                value[prefix + "_path"] = path.relative_to(root).as_posix()
                value[prefix + "_sha256"] = hashlib.sha256(path.read_bytes()).hexdigest()
            record.write_bytes(adoption.canonical(value))
        self._commit(root, "H")
        ast.update({"package": "soak", "signature": "Attest(params=0, results=1)", "branches": []})
        if deleted: ast["revision"] = "base"
        bundle = change / "analysis" / "function-logic" / "internal-soak--attest"; bundle.mkdir(parents=True)
        (bundle / "ast.json").write_text(json.dumps(ast))
        (bundle / "function-logic-map.md").write_text("# Function Logic Map: `Attest`\ninternal/soak/attest.go\n## Inputs and invariants\nevidence\n## Branches and early returns\nevidence\n## Calls and live bindings\nevidence\n## State mutations and fallbacks\nevidence\n## Safety conclusion\nevidence\n")
        (bundle / "branch-test-map.md").write_text("# Branch Test Map: `Attest`\n| B1 | leaf | test | yes | yes |\n")
        (bundle / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/soak/attest.go\n")
        self._commit(root, "map"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
        return raw, root, p, e

    @unittest.skipUnless(os.environ.get("SDD_PYTHON"), "requires an explicit external SDD_PYTHON integration environment")
    def test_real_adoption_without_local_venv_accepts_external_doctor_probe(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertFalse((root / ".sdd" / ".venv").exists())
            result = sdd_doctor.report(root)
            typedb = result["python_modules"]["typedb-driver"]
            self.assertTrue(typedb["ok"], typedb["detail"])
            self.assertIn("mode=external", typedb["detail"])
            self.assertEqual(adoption.validate(root / "openspec" / "changes" / adoption.CHANGE, root, p, adoption.CHANGE)["effective_base"], e)
            forbidden = root / ".sdd" / ".venv" / "forbidden.py"
            forbidden.parent.mkdir(parents=True)
            forbidden.write_text("forbidden\n", encoding="utf-8")
            with self.assertRaisesRegex(
                adoption.AdoptionError,
                r"untracked/ignored input is not allowed: \.sdd/\.venv/forbidden\.py",
            ):
                adoption.validate(root / "openspec" / "changes" / adoption.CHANGE, root, p, adoption.CHANGE)

    def test_valid_adoption_uses_e_and_requires_complete_current_bundle(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])
            bundle = root / "openspec" / "changes" / adoption.CHANGE / "analysis" / "function-logic" / "internal-soak--attest"
            shutil.rmtree(bundle); self._commit(root, "remove map"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)
            self.assertTrue(any("analysis directory has no targets" in error or ("missing" in error and "Attest" in error) for error in errors))

    def test_real_adoption_deleted_function_requires_base_revision(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle(deleted=True)
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])
            path = root / "openspec" / "changes" / adoption.CHANGE / "analysis" / "function-logic" / "internal-soak--attest" / "ast.json"
            value = json.loads(path.read_text()); value["revision"] = "current"; path.write_text(json.dumps(value))
            self._commit(root, "wrong deletion revision"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            self.assertTrue(any("AST revision must be base" in error for error in check_analysis.check(adoption.CHANGE, root)))

    def test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            with mock.patch.dict("os.environ", {"SDD_BASE_REF": e}, clear=False):
                self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])
            for bad in (p, "0" * 40, "HEAD"):
                with mock.patch.dict("os.environ", {"SDD_BASE_REF": bad}, clear=False):
                    self.assertTrue(any("cannot derive" in error for error in check_analysis.check(adoption.CHANGE, root)))
            record = root / "openspec" / "changes" / adoption.CHANGE / "execution-baseline.json"
            record.write_text("{}")
            self._commit(root, "invalid record"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            with mock.patch.dict("os.environ", {"SDD_BASE_REF": e}, clear=False):
                self.assertTrue(any("invalid execution-baseline adoption" in error for error in check_analysis.check(adoption.CHANGE, root)))

    def test_real_adoption_stale_current_ast_hash_fails(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            ast_path = root / "openspec" / "changes" / adoption.CHANGE / "analysis" / "function-logic" / "internal-soak--attest" / "ast.json"
            value = json.loads(ast_path.read_text()); value["source_sha256"] = "0" * 64; ast_path.write_text(json.dumps(value))
            self._commit(root, "stale map"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            self.assertTrue(any("AST source hash is stale" in error or "AST hash does not match" in error for error in check_analysis.check(adoption.CHANGE, root)))

    def test_real_adoption_rejects_local_maps_with_function_logic_reference(self) -> None:
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            analysis = root / "openspec" / "changes" / adoption.CHANGE / "analysis"
            (analysis / "function-logic-reference.txt").write_text("some-other-change\n")
            self._commit(root, "conflicting reference"); subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)
            self.assertIn("function-logic reference cannot coexist with local function-logic evidence", errors)

    def _adoption_output(self, root: Path) -> tuple[int, str]:
        output = io.StringIO()
        with mock.patch.object(
            sys, "argv",
            ["check_analysis.py", "--change", adoption.CHANGE, "--root", str(root)],
        ), redirect_stdout(output):
            code = check_analysis.main()
        return code, output.getvalue()

    def test_the_adoption_window_ends_at_the_audited_source_commit(self) -> None:
        """이관 예외의 창 끝은 워킹트리가 아니라 감사된 `source_commit` 이다 (task 1.12).

        `validate` 는 그 값을 이미 돌려준다(`{"effective_base", "source", "ledger"}`).
        아무도 읽지 않아서 대상이 워킹트리로 떨어졌고, 그 답이 오늘 맞는 이유는
        **다른 판정** 하나 때문이다 — source→head 에 `openspec/`·`docs/pm/` 밖 파일이
        들어오면 `source-to-evidence drift` 로 죽는다. 맞는 답을 우연으로 얻고 있다.
        실측으로 그 우연은 저장소에서 required **9 대 36** 짜리다(review.md
        §Pre-Edit 1.12)."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            change = root / "openspec" / "changes" / adoption.CHANGE
            # 기대값은 실행 중인 코드가 아니라 **감사된 영수증**에서 읽는다.
            source = json.loads((change / "execution-baseline.json").read_text())["source_commit"]
            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check(adoption.CHANGE, root, context), [])
            self.assertEqual(
                context.get("landing"), source,
                "이관 경로의 대상이 감사된 source_commit 이어야 한다",
            )
            code, output = self._adoption_output(root)
            self.assertEqual(code, 0, output)
            self.assertIn(f"audited source-commit {source}", output)
            self.assertNotIn("working tree", output)

    def test_the_adoption_path_does_not_advise_a_record_it_would_refuse(self) -> None:
        """3.3 의 안내가 이관 감사와 모순됐다 (task 1.12).

        실패·성공 양쪽에서 찍는 그 줄은 "`landed-commit.txt` 를 적어서 창을 좁혀라"
        라고 말한다. 이관 경로에서 그것은 **감사 어디에도 없는 두 번째 손잡이**를
        만들라는 말이고, 아래 시험이 거절하는 바로 그 파일이다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            code, output = self._adoption_output(root)
            self.assertEqual(code, 0, output)
            self.assertNotIn(check_analysis.LANDING_FILE, output)
            # 양성 대조: 이관 성공 줄은 그대로 있어야 한다. 없으면 이 시험은 출력이
            # 비었다는 이유로도 초록이 된다.
            self.assertIn("execution-baseline adoption exception evidence complete", output)

    def test_a_landing_record_is_refused_in_the_adoption_path(self) -> None:
        """이관 예외의 입력은 전부 열거되고 digest 로 묶인다 — 착지 기록만 빼고.

        `openspec/` 아래라 `source-to-evidence drift` 검사가 통과시키고, 추적 파일이라
        untracked 감사도 못 보고, 닫힌 키 집합(`execution_baseline.py:410`)에도 없다.
        그런데 비교 대상을 고른다. 손잡이는 하나여야 하고 그것은 감사된
        `source_commit` 이다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            change = root / "openspec" / "changes" / adoption.CHANGE
            source = json.loads((change / "execution-baseline.json").read_text())["source_commit"]
            # 감사된 값과 **같은** 값을 적어도 거절한다 — 값이 아니라 손잡이의 개수가
            # 문제다. 값으로 가르면 다른 값을 적는 순간 사유가 갈린다.
            (change / check_analysis.LANDING_FILE).write_text(source + "\n")
            self._commit(root, "declare a landing point inside the adoption path")
            subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)
            self.assertTrue(errors, "감사되지 않는 두 번째 손잡이가 그대로 통과했다")
            self.assertTrue(
                any("adoption does not accept" in error for error in errors),
                f"거절 사유를 이름으로 말해야 한다: {errors}",
            )

    def _archive_adoption(self, root: Path) -> Path:
        """`openspec archive` 가 하듯 a063 디렉터리를 **통째로** 옮기고 커밋한다."""
        archived = root / "openspec" / "changes" / "archive" / f"2026-09-11-{adoption.CHANGE}"
        archived.parent.mkdir(parents=True)
        subprocess.run(
            ["git", "mv", f"openspec/changes/{adoption.CHANGE}", archived.relative_to(root).as_posix()],
            cwd=root, check=True,
        )
        self._commit(root, "archive a063")
        subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
        return archived

    def test_an_archived_adoption_is_rechecked_by_its_id(self) -> None:
        """아카이브된 a063 도 그 id 로 다시 판정할 수 있어야 한다 (task 6.2).

        spec: "완료 게이트는 아카이브된 change 의 함수 분석도 그 id 로 재검사할 수 있어야
        한다(SHALL)". 이관 경로만 못 갔다. 4.4 는 이름 판정 하나를 쟀는데, 가드를 하나씩
        풀어 가며 재 보니 막는 자리가 **넷**이었다 — 이름 · 증거 경로의 접두사 · 원장
        읽기 · 리뷰 읽기. 기록은 옮기기 **전** 경로를 적고(`draft` 가 그렇게 쓴다),
        아카이브는 내용을 안 바꾸고 자리만 옮긴다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            archived = self._archive_adoption(root)
            # 기대값은 실행 중인 코드가 아니라 **옮겨진 자리의 영수증**에서 읽는다.
            source = json.loads((archived / "execution-baseline.json").read_text())["source_commit"]
            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check(adoption.CHANGE, root, context), [])
            self.assertTrue(context.get("execution_baseline_adoption"), "이관 경로로 판정해야 한다")
            self.assertEqual(context.get("landing"), source)
            code, output = self._adoption_output(root)
            self.assertEqual(code, 0, output)
            self.assertIn(f"audited source-commit {source}", output)
            self.assertIn("execution-baseline adoption exception evidence complete", output)

    def test_a_copied_adoption_record_does_not_make_another_change_a063(self) -> None:
        """이관 예외는 a063 하나의 것이다. 신원은 **게이트가 요청받은 id** 로 가른다 (task 6.2).

        증거를 지금 자리에서 읽게 되면 경로 판정은 더 이상 "통째로 복사한 디렉터리"를
        막지 못한다 — 복사본 안에서도 적힌 경로는 a063 의 것이고 digest 도 맞는다. 그 뒤로
        복사를 막는 것은 신원 판정 **하나**다. 활성 이름과 아카이브 이름 둘 다로 복사한다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            original = root / "openspec" / "changes" / adoption.CHANGE
            shutil.copytree(original, root / "openspec" / "changes" / "a130-copy")
            shutil.copytree(original, root / "openspec" / "changes" / "archive" / "2026-09-11-a131-copy")
            self._commit(root, "copy a063 whole under two other ids")
            subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            for other in ("a130-copy", "a131-copy"):
                errors = check_analysis.check(other, root)
                self.assertTrue(
                    any("adoption is not allowed for this change/base" in error for error in errors),
                    f"{other}: {errors}",
                )
            # 양성 대조: 원본은 그대로 통과해야 한다. 안 그러면 위 거절은 픽스처가 깨졌다는
            # 이유로도 성립한다.
            self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])

    def test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised(self) -> None:
        """이관 경로의 착지 기록 probe 는 **있느냐**만 묻는다 (task 6.2.1).

        옛 판본은 그 질문에 값을 **해독하는** 함수를 불렀고, 그 호출만 try 밖이었다.
        비-UTF-8 기록이면 `ValueError` 가 `check()` 를 뚫고 `main()` 이 traceback 으로
        죽었다. spec 은 이관 경로의 착지 기록에 "이관 경로가 그 기록을 받지 않는다는 것을
        이름으로 말한다"를 요구한다 — 못 읽는 기록도 기록이다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            change = root / "openspec" / "changes" / adoption.CHANGE
            (change / check_analysis.LANDING_FILE).write_bytes(b"\xff\xfe not utf-8\n")
            self._commit(root, "an undecodable landing record inside the adoption path")
            subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)
            self.assertTrue(
                any("adoption does not accept" in error for error in errors),
                f"거절 사유를 이름으로 말해야 한다: {errors}",
            )
            code, output = self._adoption_output(root)
            self.assertEqual(code, 1, output)
            self.assertIn("[logic-map] execution-baseline adoption does not accept", output)

    def test_explicit_exemption_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            change = Path(tmp) / "openspec" / "changes" / "docs-only"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("docs-only", Path(tmp)), [])

    def test_placeholders_are_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            target = (
                Path(tmp)
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text("{}\n", encoding="utf-8")
            for name in check_analysis.REQUIRED[1:]:
                (target / name).write_text("TODO\n", encoding="utf-8")
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertGreater(len(check_analysis.check("change", Path(tmp))), 0)

    def test_completed_bundle_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = root / "internal" / "sample.go"
            source.parent.mkdir(parents=True)
            source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
            target = (
                root
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text(
                json.dumps(
                    {
                        "file": "internal/sample.go",
                        "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                        "package": "sample",
                        "function": "Run",
                        "signature": "Run(params=0, results=0)",
                        "start": {"line": 2, "column": 1},
                        "end": {"line": 2, "column": 14},
                        "branches": [],
                    }
                ),
                encoding="utf-8",
            )
            (target / "function-logic-map.md").write_text(
                """# Function Logic Map: `Run`
- Source: `internal/sample.go`
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
                encoding="utf-8",
            )
            (target / "branch-test-map.md").write_text(
                "# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n",
                encoding="utf-8",
            )
            (target / "risk-pattern-report.md").write_text(
                "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("change", root), [])

    def test_null_branches_from_the_go_extractor_are_accepted(self) -> None:
        # The Go extractor marshals a nil slice as JSON null, so a branchless
        # function arrives as "branches": null rather than []. The bundle must
        # validate instead of crashing on the None.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = root / "internal" / "sample.go"
            source.parent.mkdir(parents=True)
            source.write_text("package sample\nfunc Run() {}\n", encoding="utf-8")
            target = (
                root
                / "openspec"
                / "changes"
                / "change"
                / "analysis"
                / "function-logic"
                / "pkg--run"
            )
            target.mkdir(parents=True)
            (target / "ast.json").write_text(
                json.dumps(
                    {
                        "file": "internal/sample.go",
                        "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                        "package": "sample",
                        "function": "Run",
                        "signature": "Run(params=0, results=0)",
                        "start": {"line": 2, "column": 1},
                        "end": {"line": 2, "column": 14},
                        "branches": None,
                    }
                ),
                encoding="utf-8",
            )
            (target / "function-logic-map.md").write_text(
                """# Function Logic Map: `Run`
- Source: `internal/sample.go`
## Inputs and invariants
evidence
## Branches and early returns
evidence
## Calls and live bindings
evidence
## State mutations and fallbacks
evidence
## Safety conclusion
evidence
""",
                encoding="utf-8",
            )
            (target / "branch-test-map.md").write_text(
                "# Branch Test Map: `Run`\n| B1 | leaf | test | yes | yes |\n",
                encoding="utf-8",
            )
            (target / "risk-pattern-report.md").write_text(
                "# Risk Pattern Report: `Run`\ninternal/sample.go\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value={},
            ):
                self.assertEqual(check_analysis.check("change", root), [])

    def test_modified_function_cannot_use_exemption(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            required = {
                ("internal/order.go", "Engine.Place"): {
                    "file": "internal/order.go",
                    "function": "Engine.Place",
                    "current_hash": "abc",
                }
            }
            with mock.patch(
                "check_analysis.resolve_base",
                return_value="base",
            ), mock.patch(
                "check_analysis.changed_existing_functions",
                return_value=required,
            ):
                errors = check_analysis.check("change", root)
            self.assertTrue(any("Engine.Place" in error for error in errors))

    def test_invalid_base_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "review.md").write_text(
                "Function Logic Map: not-applicable\n",
                encoding="utf-8",
            )
            with mock.patch(
                "check_analysis.resolve_base",
                side_effect=ValueError("invalid base"),
            ):
                errors = check_analysis.check("change", root)
            self.assertTrue(any("invalid base" in error for error in errors))

    def test_duplicate_branch_ids_do_not_satisfy_ast_coverage(self) -> None:
        text = "| B1 | one |\n| B1 | duplicate |\n"
        self.assertEqual(check_analysis.branch_ids(text), ["B1", "B1"])

    def test_git_diff_failure_is_not_treated_as_empty_change(self) -> None:
        failed = subprocess.CompletedProcess([], 128, "", "bad revision")
        with mock.patch(
            "check_analysis.subprocess.run",
            return_value=failed,
        ), mock.patch("check_analysis._safe_changed_go_paths"):
            with self.assertRaises(RuntimeError):
                check_analysis.changed_existing_functions(Path("/tmp"), "bad")

    def test_newline_changed_go_path_is_rejected_before_unified_diff_parsing(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            for args in (("init", "-q"), ("config", "user.email", "a120@example.invalid"), ("config", "user.name", "a120")):
                subprocess.run(["git", *args], cwd=root, check=True)
            path = root / "pkg" / "line\nbreak.go"; path.parent.mkdir()
            path.write_text("package pkg\nfunc X() {}\n", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=root, check=True); subprocess.run(["git", "commit", "-qm", "base"], cwd=root, check=True)
            base = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
            path.write_text("package pkg\nfunc X() { println(1) }\n", encoding="utf-8")
            with self.assertRaisesRegex(RuntimeError, "cannot be represented losslessly"):
                check_analysis._safe_changed_go_paths(root, base, "")

    def test_new_function_in_existing_file_is_not_reported_as_modified_existing(self) -> None:
        diff = subprocess.CompletedProcess(
            [],
            0,
            "\n".join(
                (
                    "diff --git a/internal/x.go b/internal/x.go",
                    "--- a/internal/x.go",
                    "+++ b/internal/x.go",
                    "@@ -8,0 +9,3 @@",
                )
            ),
            "",
        )
        old_functions = [
            {
                "function": "Existing",
                "start": {"line": 2},
                "end": {"line": 5},
                "source_sha256": "old-file",
            }
        ]
        current_functions = old_functions + [
            {
                "function": "Added",
                "start": {"line": 9},
                "end": {"line": 11},
                "source_sha256": "current-file",
            }
        ]
        with tempfile.TemporaryDirectory() as tmp, tempfile.NamedTemporaryFile(suffix=".go") as base_source:
            root = Path(tmp)
            current = root / "internal" / "x.go"
            current.parent.mkdir(parents=True)
            current.write_text("package x\n", encoding="utf-8")
            with mock.patch(
                "check_analysis.subprocess.run",
                return_value=diff,
            ), mock.patch(
                "check_analysis._safe_changed_go_paths",
            ), mock.patch(
                "check_analysis.base_file",
                return_value=Path(base_source.name),
            ), mock.patch(
                "check_analysis.go_functions",
                side_effect=(old_functions, current_functions),
            ):
                required = check_analysis.changed_existing_functions(root, "base")

        self.assertNotIn(("internal/x.go", "Added"), required)

    def test_base_file_load_failure_is_not_treated_as_new_file(self) -> None:
        diff = subprocess.CompletedProcess(
            [],
            0,
            "\n".join(
                (
                    "diff --git a/internal/x.go b/internal/x.go",
                    "--- a/internal/x.go",
                    "+++ b/internal/x.go",
                    "@@ -1 +1 @@",
                )
            ),
            "",
        )
        missing = subprocess.CompletedProcess([], 128, b"", b"missing")
        with mock.patch(
            "check_analysis.subprocess.run",
            side_effect=(diff, missing),
        ), mock.patch(
            "check_analysis._safe_changed_go_paths",
        ):
            with self.assertRaises(RuntimeError):
                check_analysis.changed_existing_functions(Path("/tmp"), "base")

    def test_environment_base_cannot_override_persisted_change_base(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            change = root / "openspec" / "changes" / "change"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text("persisted\n", encoding="utf-8")
            persisted = subprocess.CompletedProcess([], 0, "a" * 40 + "\n", "")
            override = subprocess.CompletedProcess([], 0, "b" * 40 + "\n", "")
            with mock.patch.dict(
                "os.environ",
                {"SDD_BASE_REF": "HEAD"},
                clear=False,
            ), mock.patch(
                "check_analysis.subprocess.run",
                side_effect=(persisted, override),
            ):
                with self.assertRaises(ValueError):
                    check_analysis.resolve_base(change, root, change_id=change.name)


    # The three checks below exist because ten of thirty-six a092 artifacts
    # asserted branch coverage the source no longer had. Existence, hash and
    # AST-branch coverage all passed; the prose had simply never been
    # re-anchored after the function moved. A checker that reads only the
    # bundle's shape cannot tell a current map from a stale one.

    def test_prose_line_range_must_match_the_ast_range(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[],
                logic_source_line="- Source: `internal/sample.go` (111-151)",
            )
            errors = run_check(root)
        self.assertTrue(any("line range" in error for error in errors), errors)

    def test_a_matching_prose_line_range_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[],
                logic_source_line="- Source: `internal/sample.go` (2-2)",
                branch_source_line="Source: `internal/sample.go` (2-2).",
            )
            self.assertEqual(run_check(root), [])

    def test_an_absent_prose_line_range_is_still_accepted(self) -> None:
        # Most maps in the corpus cite no line range at all. Requiring one
        # would fail evidence that never claimed a coordinate.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(root, branches=[])
            self.assertEqual(run_check(root), [])

    def test_prose_branch_count_must_match_the_ast_branch_count(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            target = write_bundle(root, branches=[])
            (target / "function-logic-map.md").write_text(
                (target / "function-logic-map.md")
                .read_text(encoding="utf-8")
                .replace(
                    "## Inputs and invariants",
                    "- AST evidence: `ast.json` — branches 9\n## Inputs and invariants",
                ),
                encoding="utf-8",
            )
            errors = run_check(root)
        self.assertTrue(any("branch count" in error for error in errors), errors)

    def test_a_branch_count_claim_without_an_ast_anchor_is_not_read_as_one(self) -> None:
        # "미테스트 분기 5개" is prose about coverage, not a claim about what
        # the extractor found. Only a claim anchored to AST is checked.
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            target = write_bundle(root, branches=[])
            (target / "branch-test-map.md").write_text(
                (target / "branch-test-map.md").read_text(encoding="utf-8")
                + "\n미테스트 분기 5개는 이 change의 범위가 아니다.\n",
                encoding="utf-8",
            )
            self.assertEqual(run_check(root), [])

    def test_branch_ids_absent_from_the_ast_are_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(
                root,
                branches=[{"id": "B1", "kind": "if", "at": {"line": 2}}],
                branch_rows="| B1 | leaf | test | yes | yes |\n| B2 | ghost | test | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(any("B2" in error for error in errors), errors)

    def test_the_branchless_happy_path_row_is_not_read_as_an_extra_id(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_bundle(root, branches=None)
            self.assertEqual(run_check(root), [])


def write_test_file(root: Path, name: str, tests: dict[str, int]) -> Path:
    """Write a Go test file where each test's body is padded to a known length.

    The maps under check cite lines inside a test, not its declaration, so the
    fixture has to give each test a body with addressable interior lines.
    """
    lines: list[str] = ["package sample", ""]
    for test, body in sorted(tests.items(), key=lambda item: item[1]):
        lines.append(f"func {test}(t *testing.T) {{")
        lines.extend(["\t_ = t" for _ in range(body)])
        lines.append("}")
        lines.append("")
    path = root / "internal" / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines), encoding="utf-8")
    return path


class TestNamedTestsAreOpened(unittest.TestCase):
    """B-T1: a map that names a test must be answerable by the tree.

    Ten of a092's rows carried false coverage claims through eighteen rounds
    because the checker never opened the file the row pointed at. It verified
    that files existed, that hashes were current and that every AST branch had
    a row -- all of which a fabricated test name satisfies.
    """

    def test_a_cited_test_that_exists_nowhere_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestNobodyWroteThis` | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(
            any("TestNobodyWroteThis" in error for error in errors), errors
        )

    def test_a_cited_test_that_exists_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_row_naming_no_test_at_all_is_still_accepted(self) -> None:
        """Rows that honestly say "없음" are the point of the map, not a defect."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | **없음** — 이 change의 대상 | no | no |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_cited_line_past_the_end_of_the_test_file_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` (`sample_test.go:900`) | yes | yes |",
            )
            errors = run_check(root)
        self.assertTrue(any("900" in error for error in errors), errors)

    def test_a_test_named_in_one_file_beside_call_sites_in_another_is_accepted(self) -> None:
        """a091's `severityof` row does this and is correct.

        19판 first required the cited line to belong to the test named beside
        it. This row is why that rule was withdrawn: it names a test with its
        own coordinate and then cites two independent call sites of the function
        under test, in a different file. Which coordinate answers for which
        claim lives in the prose, not in line co-occurrence.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_test_file(root, "other_test.go", {"TestSomewhereElse": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows=(
                    "| B1 | leaf | `TestRunLeaf` `sample_test.go:4` "
                    "· `other_test.go:4` | yes | yes |"
                ),
            )
            self.assertEqual(run_check(root), [])

    def test_a_qualified_path_resolves_to_that_file_not_the_local_one(self) -> None:
        """`replay_test.go` exists in two packages. The citation says which.

        a098 hit this for real: the bare name resolved against the package under
        test and answered for internal/journal's 228-line file when the row meant
        internal/execgw's. Qualifying the path has to actually redirect the
        lookup, or the advice to qualify it is empty.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            # Same basename, two packages. The local one is short; the cited one
            # is long enough that the line number only fits there.
            write_test_file(root, "shared_test.go", {"TestNear": 2})
            far = root / "other" / "shared_test.go"
            far.parent.mkdir(parents=True, exist_ok=True)
            far.write_text(
                "package other\n" + "\n".join("// pad" for _ in range(60)),
                encoding="utf-8",
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `other/shared_test.go:40` | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_qualified_path_that_does_not_exist_is_not_silently_skipped(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `nowhere/sample_test.go:4` | yes | yes |",
            )
            # Unresolvable is not an error today -- the corpus cites files in
            # trees this checker does not own. Pinned so a future change to that
            # policy is a decision, not a side effect.
            self.assertEqual(run_check(root), [])

    def test_a_line_in_the_named_tests_doc_comment_is_accepted(self) -> None:
        """a092 cites two doc comments. A comment attached to a test names it."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            path.write_text(
                "package sample\n\n// TestRunLeaf: why this test exists.\n"
                "func TestRunLeaf(t *testing.T) {\n\t_ = t\n}\n",
                encoding="utf-8",
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `TestRunLeaf` (`sample_test.go:3`) | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_line_in_a_shared_harness_in_the_same_file_is_accepted(self) -> None:
        """`obs_test.go:338` is `newNotifier`, a helper -- and honest evidence."""
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = write_test_file(root, "sample_test.go", {"TestRunLeaf": 3})
            path.write_text(
                path.read_text(encoding="utf-8")
                + "\nfunc newHarness(t *testing.T) int {\n\treturn 1\n}\n",
                encoding="utf-8",
            )
            total = len(path.read_text(encoding="utf-8").splitlines())
            write_bundle(
                root,
                branches=None,
                branch_rows=f"| B1 | leaf | `TestRunLeaf` (`sample_test.go:{total - 1}`) | yes | yes |",
            )
            self.assertEqual(run_check(root), [])

    def test_a_coordinate_without_a_name_in_the_same_row_is_accepted(self) -> None:
        """a092's `newnotifier` row cites two direct call sites and names neither.

        The row is true -- both lines call the function under test. Requiring a
        name here would reject correct evidence, so the name/line agreement is
        checked only where the row supplies both.
        """
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_test_file(
                root, "sample_test.go", {"TestRunLeaf": 3, "TestSomethingElse": 3}
            )
            write_bundle(
                root,
                branches=None,
                branch_rows="| B1 | leaf | `sample_test.go:9` — 직접 부른다 | yes | yes |",
            )
            self.assertEqual(run_check(root), [])



def _commit_all(root: Path, subject: str) -> str:
    subprocess.run(["git", "add", "."], cwd=root, check=True)
    subprocess.run(["git", "commit", "-qm", subject], cwd=root, check=True)
    return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()


def _init_fixture(raw: tempfile.TemporaryDirectory) -> Path:
    root = Path(raw.name)
    subprocess.run(["git", "init", "-q"], cwd=root, check=True)
    for key, value in (("user.email", "a122@example.invalid"), ("user.name", "a122")):
        subprocess.run(["git", "config", key, value], cwd=root, check=True)
    (root / "go.mod").write_text("module fixture\ngo 1.23\n")
    # 실제 Go AST 추출기를 픽스처 안에서 돌린다 — mock 을 쓰면 생산 경로가 아니라
    # 시험이 조종석에 앉는다.
    shutil.copytree(
        Path(__file__).resolve().parent,
        root / "tools" / "logic-map",
        ignore=shutil.ignore_patterns("*.py", "__pycache__"),
    )
    return root


def _write_evidence(change: Path, *, package: str, function: str, relative: str, digest: str) -> Path:
    """`revision: current` 번들 하나. 해시는 **호출자가 고른 리비전**의 것이다."""
    bundle = change / "analysis" / "function-logic" / f"{package}--{function.lower()}"
    bundle.mkdir(parents=True, exist_ok=True)
    (bundle / "ast.json").write_text(
        json.dumps(
            {
                "file": relative,
                "source_sha256": digest,
                "package": package,
                "function": function,
                "signature": f"{function}(params=0, results=1)",
                "start": {"line": 2, "column": 1},
                "end": {"line": 2, "column": 1},
                "branches": [],
            }
        ),
        encoding="utf-8",
    )
    (bundle / "function-logic-map.md").write_text(
        f"# Function Logic Map: `{function}`\n{relative}\n"
        "## Inputs and invariants\ne\n## Branches and early returns\ne\n"
        "## Calls and live bindings\ne\n## State mutations and fallbacks\ne\n"
        "## Safety conclusion\ne\n",
        encoding="utf-8",
    )
    (bundle / "branch-test-map.md").write_text(
        f"# Branch Test Map: `{function}`\n| B1 | leaf | test | yes | yes |\n", encoding="utf-8"
    )
    (bundle / "risk-pattern-report.md").write_text(
        f"# Risk Pattern Report\n{relative}\n", encoding="utf-8"
    )
    return bundle


def _borrowed_fixture(
    raw: tempfile.TemporaryDirectory, *, later_work: bool = False
) -> tuple[Path, dict[str, str]]:
    """`ordinary` 가 `reference` 의 번들을 빌린다. base P → 착지 L.

    a073 이 a072 의 231개 번들을 이렇게 빌려 쓴다. 자기 번들은 0 이다.

    `later_work` 를 켜면 L **뒤에** 커밋 L2 가 하나 더 서고, 거기서 빌린 증거가
    기술하지 않는 기존 Go 함수가 바뀐다. 빌린 번들은 L 에서도 L2 에서도 그대로
    고정되므로(그 파일은 안 바뀐다), 저자가 L 과 L2 중에서 **고를 수 있다**는
    것이 3.2.3.1 이 남긴 자유다. 그 자유가 요구 집합을 가르는 모양이다."""
    root = _init_fixture(raw)
    source = root / "internal" / "sample.go"
    source.parent.mkdir(parents=True)
    source.write_text("package sample\nfunc Run() int { return 1 }\n")
    other = root / "internal" / "other.go"
    if later_work:
        other.write_text("package sample\nfunc Other() int { return 1 }\n")
    marks = {"P": _commit_all(root, "P: base")}
    change = root / "openspec" / "changes" / "ordinary"
    (change / "analysis").mkdir(parents=True)
    (change / "base-commit.txt").write_text(marks["P"] + "\n")
    (change / "review.md").write_text("ordinary\n")
    (change / "analysis" / "function-logic-reference.txt").write_text("reference\n")
    reference = root / "openspec" / "changes" / "reference"
    reference.mkdir(parents=True)
    (reference / "base-commit.txt").write_text(marks["P"] + "\n")
    (reference / "review.md").write_text("reference\n")
    source.write_text("package sample\nfunc Run() int { return 2 }\n")
    _write_evidence(
        reference, package="sample", function="Run", relative="internal/sample.go",
        digest=hashlib.sha256(source.read_bytes()).hexdigest(),
    )
    marks["L"] = _commit_all(root, "L: the work and the borrowed evidence land")
    if later_work:
        other.write_text("package sample\nfunc Other() int { return 2 }\n")
        marks["L2"] = _commit_all(root, "L2: work the borrowed evidence does not describe")
    return root, marks


def _declare_landing(change: Path, value: str, subject: str) -> None:
    """착지 선언은 워킹트리가 아니라 **커밋**에서 읽히므로 반드시 커밋한다."""
    (change / "landed-commit.txt").write_text(value + "\n")
    _commit_all(_root_of(change), subject)


def _root_of(change: Path) -> Path:
    """`openspec/changes/<id>` 에서 저장소 루트로. 세지 않고 **유도**한다."""
    return next(parent for parent in change.parents if (parent / "go.mod").is_file())


class LandingPointBoundsTheComparisonTarget(unittest.TestCase):
    """5단계 비교의 **대상 쪽 끝**을 그 change 의 작업이 착지한 지점으로 묶는다.

    오늘은 대상이 워킹트리라, base 와 워킹트리 사이에 들어온 **다른 change 의 함수**가
    이 change 가 증거를 안 낸 함수로 집계된다. 배포 후 실측 태스크는 정의상 다른
    작업이 착지한 뒤에 닫히므로, 그런 태스크를 가진 change 는 전부 이 상태로 끝난다."""

    def _merge_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """실제 병합 모양: base P → Go 작업 → 증거 L → 이웃 change N → 나중 리팩터 M.

        증거는 작업이 착지한 **직후**에 들어온다. 옛 판본은 증거를 N·M 뒤에 커밋하고
        착지로는 그 앞을 선언했는데, 그 순서는 저장소에 없다 — a099(유일한 실물 기록)
        의 바닥은 `21a315d1` 이고 기록된 착지 `e6c4636a` **앞**이다. task 6.1.2.1 의
        하한이 그 비현실적 순서를 드러냈다. 단언은 하나도 안 바꿨다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        other = root / "internal" / "other.go"
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work lands")
        marks["own_at_L"] = hashlib.sha256(own.read_bytes()).hexdigest()
        _write_evidence(
            change, package="internal", function="Own",
            relative="internal/own.go", digest=marks["own_at_L"],
        )
        marks["L"] = _commit_all(root, "L: its evidence enters the history")
        other.write_text("package internal\nfunc Other() int { return 2 }\n")
        marks["N"] = _commit_all(root, "N: a different change lands")
        own.write_text("package internal\nfunc Own() int { return 3 }\n")
        marks["M"] = _commit_all(root, "M: someone else refactors the same file later")
        return raw, root, marks

    def test_recorded_landing_requires_only_this_changes_functions(self) -> None:
        """task 2.1 — 착지 지점이 기록되면 base..착지만 요구한다."""
        raw, root, marks = self._merge_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(marks["L"] + "\n")
            _commit_all(root, "record the landing point and the evidence")
            self.assertEqual(check_analysis.check("mine", root), [])

    def test_without_a_landing_record_the_target_is_still_the_worktree(self) -> None:
        """task 2.2 — 기록이 없으면 오늘 그대로다. 진행 중 change 의 판정은 안 바뀐다.

        이 시험만은 오늘도 초록이어야 한다. 초록에서 초록으로 남는 것이 이 시험의 일이다."""
        raw, root, marks = self._merge_fixture()
        with raw:
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("internal/other.go:Other" in error for error in errors),
                f"이웃 change 의 함수가 여전히 요구돼야 한다: {errors}",
            )
            self.assertTrue(
                any("stale" in error for error in errors),
                f"워킹트리 대조라 증거가 낡아야 한다: {errors}",
            )

    def test_an_uncommitted_landing_record_is_not_read(self) -> None:
        """task 1.5 — 기록은 워킹트리가 아니라 커밋에서 읽는다.

        워킹트리에서 읽으면 untracked 파일로 게이트를 통과한 뒤 지울 수 있고, 그러면
        어떤 대상으로 통과했는지가 아무 데도 안 남는다. 이 픽스처는 기록 없이는
        빨강이고 기록이 **먹히면** 초록이 되므로, 초록이 되면 워킹트리를 읽은 것이다."""
        raw, root, marks = self._merge_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(marks["L"] + "\n")  # untracked
            self.assertTrue(
                check_analysis.check("mine", root),
                "커밋되지 않은 착지 기록이 판정을 바꿨다 — 워킹트리를 읽고 있다",
            )
            _commit_all(root, "now commit the same record")
            self.assertEqual(
                check_analysis.check("mine", root), [],
                "같은 값을 커밋하면 이번엔 먹혀야 한다 — 양성 대조군",
            )

    def test_the_target_argument_actually_reaches_the_diff(self) -> None:
        """task 2.5 — 인자를 **빼는** 변이가 스위트를 통과하면 안 된다.

        기존 patch 자리 여섯은 전부 `return_value` 라 인자를 안 본다."""
        raw, root, marks = self._merge_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(marks["L"] + "\n")
            _commit_all(root, "record the landing point")
            seen: list[tuple] = []
            real = check_analysis.changed_existing_functions

            def spy(*args, **kwargs):
                seen.append((args, kwargs))
                return real(*args, **kwargs)

            with mock.patch("check_analysis.changed_existing_functions", side_effect=spy):
                check_analysis.check("mine", root)
            self.assertTrue(seen, "changed_existing_functions 가 불리지 않았다")
            args, kwargs = seen[0]
            passed = kwargs.get("target", args[2] if len(args) > 2 else "")
            self.assertEqual(
                passed, marks["L"],
                "착지 지점이 diff 의 target 으로 실제로 건너가야 한다",
            )


class AForgedLandingPointIsRefusedByName(unittest.TestCase):
    """착지 지점의 유효성은 그 change 의 **증거**로 판정한다.

    신원(`base-commit.txt` 가 그 커밋에 같은 내용으로 있는가)으로 판정하면 안 된다 —
    `base-commit.txt` 는 freeze 에 쓰이고 spec 이 불변을 요구하므로 그 검사는 freeze
    이후 **모든** 커밋에서 참이고 아무것도 가르지 못한다. a112 에 freeze 직후 커밋을
    주면 신원 판정 넷이 다 통과하면서 base→착지가 135→12 파일로 줄고 비테스트 Go
    48개(journal/schema.go · strategy_lane_latch_v32.go · positioncampaign/*)가 숨었다.

    **이 클래스의 모든 픽스처는 오늘 초록이다.** 착지 기록이 없으면 통과하고, 위조된
    기록을 넣었을 때만 빨개져야 한다. 그렇지 않으면 시험이 다른 이유로 초록/빨강이
    되어 판정을 재지 못한다."""

    def _clean_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """R → P(base) → L(HEAD). 증거는 워킹트리와 일치하므로 오늘 check() 는 [] 다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 0 }\n")
        marks = {"R": _commit_all(root, "R: before the base")}
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        marks["P"] = _commit_all(root, "P: base")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["L"] = _commit_all(root, "L: this change's work lands")
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        _commit_all(root, "evidence")
        return raw, root, marks

    def test_the_fixture_passes_before_any_landing_record(self) -> None:
        """양성 대조군. 이것이 초록이 아니면 아래 넷은 아무것도 재지 못한다."""
        raw, root, marks = self._clean_fixture()
        with raw:
            self.assertEqual(check_analysis.check("mine", root), [])

    def _refuse(self, value: str, needle: str) -> None:
        raw, root, marks = self._clean_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(value.format(**marks) + "\n")
            _commit_all(root, "record a landing point")
            errors = check_analysis.check("mine", root)
            self.assertTrue(errors, f"거절해야 한다: {value}")
            self.assertTrue(
                any(needle in error for error in errors),
                f"사유를 이름으로 말해야 한다 (기대: {needle!r}): {errors}",
            )

    def test_a_landing_that_is_not_a_commit(self) -> None:
        self._refuse("f" * 40, "landing")

    def test_a_landing_before_the_base(self) -> None:
        self._refuse("{R}", "landing")

    def test_a_landing_the_evidence_does_not_describe(self) -> None:
        """task 2.3·2.4 — `P` 는 신원 판정 넷을 전부 통과한다.

        실재하고, HEAD 의 조상이고, base 자신이며, 그 커밋에 이 change 의
        `base-commit.txt` 가 같은 내용으로 있다. 요구 집합은 ∅ 이 되어 "넣기만 하면
        통과" 구현이라면 초록이 된다. 증거는 `L` 을 기술하므로 거절돼야 한다."""
        # 사유를 **착지 판정의 문장**으로 못 박는다. `internal/own.go` 로 두면
        # `validate_target` 의 `AST source hash is stale` 도 그 바늘을 만족해서,
        # 착지 판정을 통째로 지우는 변이가 이 시험을 초록으로 통과한다(M4 로 실측).
        self._refuse("{P}", "is not the revision this evidence describes")

    def test_a_landing_written_as_a_revision_expression(self) -> None:
        """task 1.7 — `rev-parse` 는 `HEAD`·브랜치·태그를 받는다.

        `HEAD` 한 단어면 커밋 안 된 Go 편집이 통째로 요구에서 빠지고, 값의 뜻이
        리뷰 시점과 게이트 시점 사이에 바뀐다."""
        self._refuse("HEAD", "landing")

    def test_an_empty_record_is_still_a_declaration(self) -> None:
        """빈 파일은 **선언이 없는 것**이 아니다.

        task 1.8 이 선언을 읽는 자리를 헬퍼로 뽑았다. 거기서 "파일 없음"과 "빈
        선언"을 안 가르면, 빈 파일을 커밋한 change 가 조용히 워킹트리를 대상으로
        삼고 아무 사유도 안 남는다."""
        self._refuse("", "landing")



class ADeclaredLandingMustNotPrecedeItsOwnEvidence(unittest.TestCase):
    """착지는 그 증거가 역사에 들어온 지점보다 앞설 수 없다 (task 6.1.2.1).

    6.1.1 이 활성 13건 전부를 재서 얻은 축이다. 13건 **전부** 고정 번들이 자기 base
    뒤에 커밋됐다. 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없으므로 "번들이
    역사에 들어온 지점"은 저자가 고를 수 없는 유일한 하한이다.

    막는 것은 4.4 가 실측으로 재현한 위조 1번이다: 증거를 **base 상태**로 써 두고
    `landed-commit = base` 를 선언하면 창이 비어 required 0 이 된다. 같은 입력이
    선언 **없이는** `AST source hash is stale` 로 빨갛다 — a122 가 못 막은 것이
    아니라 없던 길을 연다.
    """

    def _forged_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """R → P(base) → L(진짜 Go 작업) → E(**base 상태**를 기술하는 증거)."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 0 }\n")
        marks = {"R": _commit_all(root, "R: before the base")}
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        marks["P"] = _commit_all(root, "P: base")
        at_base = hashlib.sha256(own.read_bytes()).hexdigest()
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["L"] = _commit_all(root, "L: this change's Go work lands")
        # 저자가 고른 것은 이 해시 하나뿐이다 — 나머지는 전부 진짜 경로다.
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=at_base,
        )
        marks["E"] = _commit_all(root, "E: evidence enters history")
        return raw, root, marks

    def test_the_forgery_is_red_without_a_landing_record(self) -> None:
        """음성 대조군. 선언이 없으면 같은 입력이 오늘도 빨갛다.

        이것이 빨갛지 않으면 아래 시험은 "선언이 길을 연다"를 재지 못하고 그냥
        깨진 픽스처를 재게 된다."""
        raw, root, _ = self._forged_fixture()
        with raw:
            self.assertTrue(check_analysis.check("mine", root))

    def test_evidence_written_at_the_base_cannot_declare_the_base(self) -> None:
        raw, root, marks = self._forged_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(marks["P"] + "\n")
            _commit_all(root, "declare the base as the landing")
            errors = check_analysis.check("mine", root)
            self.assertTrue(errors, "base 상태 증거로 base 를 선언하면 거절해야 한다")
            # 바늘은 이 가드의 **자기 문장**이다. `"landing"` 으로 두면 착지 판정
            # 어느 것을 지워도 다른 가드가 거절해서 초록으로 남는다(6.3 이 같은
            # 파일에서 실측한 함정).
            self.assertTrue(
                any("precedes the evidence that pins it" in error for error in errors),
                f"사유를 이 가드의 자기 문장으로 말해야 한다: {errors}",
            )

    def test_an_honest_landing_at_the_evidence_commit_passes(self) -> None:
        """양성 대조군. 하한이 정상 선언까지 막으면 여기가 빨개진다.

        정직한 change 는 증거가 자기 작업을 기술하고, 그 증거가 들어온 커밋을
        선언한다. 요구 집합은 비지 않는다(`Own` 하나) — 통과가 "검사를 안 했다"와
        구분되는 자리다."""
        raw, root, marks = AForgedLandingPointIsRefusedByName()._clean_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            evidence_commit = subprocess.check_output(
                ["git", "rev-parse", "HEAD"], cwd=root, text=True
            ).strip()
            (change / "landed-commit.txt").write_text(evidence_commit + "\n")
            _commit_all(root, "record the landing the gate computed")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts["landing"], evidence_commit)
            self.assertEqual(facts["required_count"], 1)


class TheEvidenceFloorSurvivesArchivingAndDemandsCommittedEvidence(unittest.TestCase):
    """하한을 만드는 두 조각을 각각 못 박는다 (task 6.1.2.1).

    M2·M3 변이가 살아남아서 쓴 시험이다 — 하한 판정(M1)만 시험이 있었고, "바닥이
    없으면 거절"과 "rename 은 바닥이 아니다"는 어느 시험도 안 봤다.
    [[surviving-mutant-may-mean-accidental-safety]]
    """

    def _fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work lands")
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        return raw, root, marks

    def test_uncommitted_evidence_cannot_pin_a_landing(self) -> None:
        """M2 — 번들이 역사에 없으면 바닥이 없고, 그러면 하한도 없다.

        디스크의 번들로 고정 순회는 돌지만 그 파일들은 게이트가 끝난 뒤 지울 수 있다.
        바닥 없이 통과시키면 저자는 다시 구간의 바닥을 고를 수 있다."""
        raw, root, marks = self._fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            # 해시가 **맞는** 지점을 선언한다. 안 맞는 지점을 쓰면 위의 불일치
            # 판정이 먼저 거절해서 이 시험은 바닥을 재지 못한다.
            (change / "landed-commit.txt").write_text(marks["W"] + "\n")
            subprocess.run(["git", "add", "openspec/changes/mine/landed-commit.txt"],
                           cwd=root, check=True)
            subprocess.run(["git", "commit", "-qm", "record a landing, evidence untracked"],
                           cwd=root, check=True)
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("never entered this history" in error for error in errors),
                f"커밋 안 된 증거는 착지를 고정하지 못한다: {errors}",
            )

    def test_archiving_the_change_does_not_invalidate_its_record(self) -> None:
        """M3 — 아카이브는 번들을 통째로 **옮긴다**. 그 rename 은 바닥이 아니다.

        rename 을 바닥으로 세면 아카이브하는 순간 이미 유효했던 기록이 무효가 되고,
        spec 의 "아카이브된 change 의 함수 분석도 그 id 로 재검사할 수 있어야
        한다(SHALL)"가 깨진다. 저장소의 유일한 실물 기록 a099 가 정확히 이 모양이다
        — 바닥 `21a315d1` 은 기록된 착지 `e6c4636a` 앞이고, rename 커밋은 뒤다."""
        raw, root, marks = self._fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            landing = _commit_all(root, "L: evidence enters the history")
            (change / "landed-commit.txt").write_text(landing + "\n")
            _commit_all(root, "record the landing the gate computed")
            self.assertEqual(
                check_analysis.check("mine", root), [], "아카이브 전에는 통과해야 한다"
            )
            archive = root / "openspec" / "changes" / "archive" / "2026-09-11-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            subprocess.run(["git", "mv", "openspec/changes/mine", archive.relative_to(root).as_posix()],
                           cwd=root, check=True)
            _commit_all(root, "archive the change")
            self.assertEqual(
                check_analysis.check("mine", root), [],
                "아카이브 뒤에도 같은 id 로 재검사돼야 한다",
            )


class TheGateRecordsTheLandingInsteadOfTheAuthor(unittest.TestCase):
    """착지 값을 만드는 주체가 저자에서 도구로 바뀐다 (task 6.1.2.2).

    사람이 2026-09-11 에 고른 방향이다. 6.1.1 이 활성 13건을 전수로 재서 얻은 것은
    "증거만으로는 위조 차단과 5.7 해제가 양립하지 않는다"였고, 네 축 중 값을 **만드는
    주체**를 바꾸는 것은 이것 하나뿐이었다.
    """

    def _fixture(self, *, base_also_matches: bool) -> tuple[
        tempfile.TemporaryDirectory, Path, dict[str, str]
    ]:
        """P(base) → W(Go 작업) → E(증거).

        `base_also_matches` 를 켜면 증거가 **base 상태**를 기술한다. 그러면 base 자신도
        고정을 통과하므로, 도구가 "가장 낮은 후보"를 고른다면 base 를 쓴다. 안 쓰는
        것이 이 클래스가 재는 것이다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        at_base = hashlib.sha256(own.read_bytes()).hexdigest()
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        if not base_also_matches:
            own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work lands")
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=at_base if base_also_matches else hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        marks["E"] = _commit_all(root, "E: its evidence enters the history")
        return raw, root, marks

    def test_the_recorded_value_is_one_the_gate_then_accepts(self) -> None:
        """왕복. 도구가 쓴 값으로 5단계가 통과하지 않으면 기록은 쓸모가 없다."""
        raw, root, marks = self._fixture(base_also_matches=False)
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            recorded = (root / "openspec" / "changes" / "mine" / "landed-commit.txt").read_text().strip()
            self.assertEqual(recorded, marks["E"])
            _commit_all(root, "commit the recorded landing")
            self.assertEqual(check_analysis.check("mine", root), [])

    def test_it_does_not_record_the_base_even_when_the_base_matches(self) -> None:
        """저자가 고를 수 있었던 가장 싼 값을 도구는 안 쓴다.

        4.4 가 실측으로 재현한 위조 1번이 정확히 이 값이다 — 증거를 base 상태로 써
        두고 base 를 선언하면 창이 비어 required 0 이 된다. 도구는 바닥 아래를 못
        고르므로 같은 입력에서 **증거가 들어온 지점**을 쓴다."""
        raw, root, marks = self._fixture(base_also_matches=True)
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            recorded = (root / "openspec" / "changes" / "mine" / "landed-commit.txt").read_text().strip()
            self.assertNotEqual(recorded, marks["P"], "base 를 기록하면 위조와 같은 값이다")
            self.assertEqual(recorded, marks["E"])

    def test_it_refuses_to_overwrite_an_existing_record(self) -> None:
        raw, root, marks = self._fixture(base_also_matches=False)
        with raw:
            (root / "openspec" / "changes" / "mine" / "landed-commit.txt").write_text(
                marks["W"] + "\n"
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1)
            self.assertTrue(any("already exists" in line for line in lines), lines)

    def test_it_refuses_while_tracked_files_are_modified(self) -> None:
        """기록은 **커밋된** 지점을 가리킨다. 그 상태의 Go 편집은 어느 커밋에도 없다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            (root / "internal" / "own.go").write_text(
                "package internal\nfunc Own() int { return 9 }\n"
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1)
            self.assertTrue(any("uncommitted changes" in line for line in lines), lines)

    def test_it_refuses_when_no_evidence_pins_a_landing(self) -> None:
        """번들이 0 인 change 는 기록할 값이 없다 — 쓰면 반드시 실패할 값을 쓰는 것이다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            shutil.rmtree(root / "openspec" / "changes" / "mine" / "analysis" / "function-logic")
            _commit_all(root, "drop the evidence")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1)
            self.assertTrue(any("no `revision: current` evidence" in line for line in lines), lines)

    def test_an_undecodable_committed_record_is_reported_not_raised(self) -> None:
        """"기록이 있는가"에 해독은 필요 없다 (task 6.2.1 — 호출 자리 열거가 찾은 둘째 자리).

        4.4 는 `check` 의 probe 하나를 셌다. 6.1.2 가 같은 모양의 probe 를 여기에 하나
        더 만들었고 역시 try 밖이다. 워킹트리에서 지운 기록이 HEAD 에 비-UTF-8 로 남아
        있으면 `exists()` 가 거짓이라 해독 호출까지 가서 터졌다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            record = root / "openspec" / "changes" / "mine" / "landed-commit.txt"
            record.write_bytes(b"\xff\xfe not utf-8\n")
            _commit_all(root, "an undecodable record")
            record.unlink()
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any("already exists" in line for line in lines), lines)

    def test_the_value_path_still_names_an_undecodable_record(self) -> None:
        """대조군. "있는가"를 해독에서 떼어 내도 **값을 읽는 쪽**은 여전히 해독 실패를
        이름으로 말해야 한다 (task 6.2.1). 저장소에 이 문장을 재는 시험이 0 이었다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            record = root / "openspec" / "changes" / "mine" / "landed-commit.txt"
            record.write_bytes(b"\xff\xfe not utf-8\n")
            _commit_all(root, "an undecodable record")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("landing point is not UTF-8" in error for error in errors), errors
            )


class ADeclaredLandingMustBePinnedByEvidence(unittest.TestCase):
    """착지 선언은 **그것을 고정할 증거가 있을 때만** 유효하다 (task 3.2.3.1).

    `analysis/landing-point.md` 는 "착지가 base 와 같아도 안전하다 — 번들이 고정하므로
    저자가 고를 수 없다"로 그 설계를 정당화한다. 그 문장에는 전제가 있다: **번들이
    있어야 한다.** 번들이 0 이면 고정 순회가 0회 돌고, `mismatched` 는 빈 채로 남고,
    저자가 구간의 바닥을 골라 요구 집합을 ∅ 로 만들 수 있다. 면제 표식까지 있으면
    그대로 통과한다 — 이 클래스는 그 전제를 코드에 세운다.

    거부하게 될 정상 입력을 먼저 적는다: **빌린 증거**(`function-logic-reference.txt`)
    를 쓰는 change 는 자기 번들이 0 이다(a073 이 그 모양이고, 오늘 빨갛고, 그 빨강을
    푸는 것이 바로 착지 선언이다). 그래서 고정은 그 change 가 실제로 딛는 증거로
    판정해야 한다. 아래 둘째·셋째 시험이 그것을 못 박는다."""

    def test_a_zero_bundle_change_cannot_declare_a_landing(self) -> None:
        """번들 0 + 면제 표식 + 바닥 착지 = 오늘은 통과한다. 그것이 이 구멍이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            own.write_text("package internal\nfunc Own() int { return 1 }\n")
            base = _commit_all(root, "P: base")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text(
                f"# review\n\n{check_analysis.EXEMPTION}\n", encoding="utf-8"
            )
            # 이 change 의 **진짜** Go 작업. base 뒤에 있으므로 숨길 것이 있다.
            own.write_text("package internal\nfunc Own() int { return 2 }\n")
            _commit_all(root, "L: this change really does change a Go function")

            # 양성 대조: 착지 기록이 없으면 그 작업이 요구로 잡혀야 한다. 이것이
            # 빨갛지 않으면 아래 단언은 아무것도 재지 못한다.
            control = check_analysis.check("mine", root)
            self.assertTrue(
                any("internal/own.go:Own" in error for error in control),
                f"숨길 Go 작업이 실제로 있어야 한다: {control}",
            )

            (change / "landed-commit.txt").write_text(base + "\n")
            _commit_all(root, "declare the freeze floor as the landing point")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                errors,
                "번들이 0 인데 바닥을 착지로 선언해 통과했다 — 고정할 것이 없는 선언이다",
            )
            self.assertTrue(
                any("pinned by no" in error for error in errors),
                f"거절 사유가 '고정할 증거가 없다'여야 한다: {errors}",
            )

    def test_base_revision_bundles_do_not_pin_a_landing(self) -> None:
        """`revision: base` 번들은 **base** 를 기술한다 — 착지에 대해 아무 말도 못 한다.

        `validate_target` 도 그 번들은 해싱하지 않는다(`revision != "current"`).
        고정할 수 없다는 것이 사실이므로 거절이 맞다. 저장소에 오늘 0건이고,
        생기면 그때 무엇으로 고정할지 정해야 한다 — review.md §Pre-Edit 3.2.3.1."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            own.write_text("package internal\nfunc Own() int { return 1 }\n")
            base = _commit_all(root, "P: base")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text("mine\n")
            bundle = _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            value = json.loads((bundle / "ast.json").read_text(encoding="utf-8"))
            value["revision"] = "base"
            (bundle / "ast.json").write_text(json.dumps(value), encoding="utf-8")
            _commit_all(root, "evidence that describes the base revision")
            self.assertEqual(
                check_analysis.check("mine", root), [],
                "양성 대조: 착지 기록 전에는 통과해야 한다",
            )
            (change / "landed-commit.txt").write_text(base + "\n")
            _commit_all(root, "record a landing point")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("pinned by no" in error for error in errors),
                f"base 리비전 번들이 착지를 고정한 것으로 셌다: {errors}",
            )

    def test_an_absolute_bundle_path_still_pins_a_landing(self) -> None:
        """거부하면 안 되는 정상 입력. 번들의 `file` 은 **절대경로일 수 있다**.

        `validate_target` 이 `normalized_source` 로 정규화하는 이유가 그것이다.
        고정 순회가 그것을 안 하면 `git show <sha>:/abs/path` 가 언제나 실패해서
        정상 입력이 위조로 몰린다 — 3.2.3.1 이 놓친 자리이고, a063 이관 픽스처가
        절대경로를 만들면서 드러났다. 저장소 번들 **3048개는 오늘 전부 상대경로**라
        실물 영향은 0 이다(2026-09-11 전수)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            own.write_text("package internal\nfunc Own() int { return 1 }\n")
            base = _commit_all(root, "P: base")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text("mine\n")
            own.write_text("package internal\nfunc Own() int { return 2 }\n")
            _write_evidence(
                change, package="internal", function="Own",
                relative=str(own),  # ← 절대경로. 추출기가 이렇게 적을 수 있다.
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            landing = _commit_all(root, "L: the work and the evidence land")
            self.assertEqual(
                check_analysis.check("mine", root), [],
                "양성 대조: 착지 기록 전에는 통과해야 한다",
            )
            _declare_landing(change, landing, "record the landing point")
            self.assertEqual(
                check_analysis.check("mine", root), [],
                "절대경로 번들이 고정하는 착지를 거절했다 — 정상 입력을 죽였다",
            )

    def test_borrowed_evidence_still_pins_a_landing(self) -> None:
        """거부하면 안 되는 정상 입력. 빌린 번들이 고정하므로 통과해야 한다.

        task 1.8 이 창의 **양쪽 끝**을 공유하게 만든 뒤로, 이 정상 입력은 빌려주는
        쪽에도 같은 선언이 있는 모양이다. 고정 판정이 재는 것은 그대로다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            change = root / "openspec" / "changes" / "ordinary"
            self.assertEqual(
                check_analysis.check("ordinary", root), [],
                "양성 대조: 착지 기록 전에는 통과해야 한다",
            )
            (root / "openspec" / "changes" / "reference" / "landed-commit.txt").write_text(
                marks["L"] + "\n"
            )
            _declare_landing(change, marks["L"], "record the shared landing point")
            self.assertEqual(
                check_analysis.check("ordinary", root), [],
                "빌린 증거로 고정되는 착지를 거절했다 — 정상 입력을 죽였다",
            )

    def test_borrowed_evidence_refuses_a_landing_it_does_not_describe(self) -> None:
        """고정이 **먹히는지**. 빌린 번들이 기술하지 않는 착지는 이름으로 거절한다.

        오늘은 `resolve_landing` 이 **지역** 번들만 보므로(빌린 change 는 0개)
        이 위조가 고정 판정을 그냥 통과하고, 뒤늦게 `validate_target` 이 다른
        사유로 빨개진다. 사유가 갈리면 무엇이 막았는지 기록에 안 남는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            change = root / "openspec" / "changes" / "ordinary"
            # 양쪽이 **같은** 바닥을 선언한다. 공유 규칙(task 1.8)에 먼저 걸리면
            # 고정 판정을 통째로 지워도 이 시험이 초록으로 남는다 — 판정 둘이
            # 서로를 가리는 자리라서 여기서 갈라 둔다.
            (root / "openspec" / "changes" / "reference" / "landed-commit.txt").write_text(
                marks["P"] + "\n"
            )
            _declare_landing(change, marks["P"], "record the freeze floor instead")
            errors = check_analysis.check("ordinary", root)
            self.assertTrue(errors, "증거가 기술하지 않는 착지가 통과했다")
            self.assertTrue(
                any("is not the revision this evidence describes" in error for error in errors),
                f"착지 판정이 이름으로 거절해야 한다: {errors}",
            )


class ABorrowedWindowIsSharedAtBothEnds(unittest.TestCase):
    """빌린 증거는 비교 창의 **양쪽 끝**을 다 공유해야 한다 (task 1.8).

    창의 **시작**에는 이미 규칙이 있다 — `referenced_base != base` 면 거절한다.
    **끝**에는 없었다. 3.2.3.1 이 넣은 고정("빌린 번들이 착지를 고정한다")은
    "착지가 같아야 한다"보다 약하다: 빌린 번들이 **안 바뀌는 구간 안에서는** 저자가
    여전히 고를 수 있고, 그 구간에 남의 Go 작업이 들어오면 요구 집합이 갈린다.

    선언 파일을 고른 근거는 `analysis/landing-point.md` 의 한 문장뿐이다 — "번들이
    고정하므로 저자가 고를 수 없다". 빌리는 change 에서는 그 번들이 **남의 것**이라
    그 문장이 끝까지 참이 되지 않는다. 그래서 착지는 그것을 고정하는 증거가 사는
    자리에 선언하고, 빌리는 쪽은 값을 **복사**한다.

    실측 (2026-09-10, a072): `revision: current` 번들 99개를 동시에 고정하는 커밋은
    base..HEAD 326개 중 **2개**이고 둘의 요구 집합은 같다. 오늘 이 규칙이 새로
    거절하는 저장소 change 는 **0건**이다 — review.md §Pre-Edit 1.8 의 표."""

    def _borrower(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "ordinary"

    def _lender(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "reference"

    def test_a_borrower_cannot_declare_a_landing_the_lender_did_not(self) -> None:
        """빌리는 쪽만 선언한 값은 저자가 고른 값이다.

        빌린 번들이 **고정하긴 한다** — 그래서 오늘 이 모양이 통과한다. 고정은
        구간을 남기고, 그 구간 안에서 값을 고른 것은 저자다. 빌려주는 쪽에 같은
        선언이 없으면 그 선택을 검증한 사람이 아무도 없다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            _declare_landing(self._borrower(root), marks["L"], "borrower declares alone")
            errors = check_analysis.check("ordinary", root)
            self.assertTrue(
                errors,
                "빌려주는 쪽이 선언하지 않은 착지를 빌리는 쪽이 혼자 선언해 통과했다",
            )
            self.assertTrue(
                any("must share the exact landing point" in error for error in errors),
                f"거절 사유가 창의 끝을 공유하지 않았다여야 한다: {errors}",
            )

    def test_a_borrower_cannot_pick_an_earlier_landing_than_the_lender(self) -> None:
        """빌린 번들이 고정하는 구간이 둘이면 저자가 좁은 쪽을 고를 수 있다.

        먼저 **양성 대조**로 L2 가 실제로 요구를 늘리는지 잰다. 안 늘면 아래 단언은
        아무것도 재지 못한다 — 숫자를 상수로 두면 변이가 살아남는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw, later_work=True)
            borrower, lender = self._borrower(root), self._lender(root)
            (lender / "landed-commit.txt").write_text(marks["L2"] + "\n")
            _declare_landing(borrower, marks["L2"], "both declare the later landing")
            control = check_analysis.check("ordinary", root)
            self.assertTrue(
                any("internal/other.go:Other" in error for error in control),
                f"양성 대조: L2 창에는 빌린 증거가 안 덮는 작업이 있어야 한다: {control}",
            )

            _declare_landing(borrower, marks["L"], "borrower narrows the window to L")
            errors = check_analysis.check("ordinary", root)
            self.assertTrue(
                any("must share the exact landing point" in error for error in errors),
                f"빌리는 쪽이 창을 저 혼자 좁혀 남의 작업을 뺐다: {errors}",
            )
            self.assertFalse(
                any("internal/other.go:Other" in error for error in errors),
                "이 시험은 좁힌 창이 실제로 그 작업을 뺀다는 것을 전제한다 — "
                f"안 빠졌다면 픽스처가 재는 것이 없다: {errors}",
            )

    def test_a_lender_only_declaration_is_not_inherited(self) -> None:
        """빌려주는 쪽만 선언한 것도 창이 안 맞는 것이다. 규칙은 한 가지 모양이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            (self._lender(root) / "landed-commit.txt").write_text(marks["L"] + "\n")
            _commit_all(root, "lender declares alone")
            errors = check_analysis.check("ordinary", root)
            self.assertTrue(
                any("must share the exact landing point" in error for error in errors),
                f"한쪽만 선언한 창을 통과시켰다: {errors}",
            )

    def test_a_shared_landing_is_accepted(self) -> None:
        """거부하면 안 되는 정상 입력 (1): 양쪽이 같은 값을 선언한 모양."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            (self._lender(root) / "landed-commit.txt").write_text(marks["L"] + "\n")
            _declare_landing(self._borrower(root), marks["L"], "both declare the landing")
            self.assertEqual(
                check_analysis.check("ordinary", root), [],
                "공유된 착지를 거절했다 — 규칙이 여는 바로 그 모양이다",
            )

    def test_neither_side_declaring_stays_accepted(self) -> None:
        """거부하면 안 되는 정상 입력 (2): 양쪽 다 선언이 없는 모양.

        a073·a072 의 **오늘 모양**이 이것이다. 이 규칙은 그 상태에 오류를 하나도
        더하지 않는다 — 저장소 실물 영향이 0 인 근거가 이 시험이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _borrowed_fixture(raw)
            self.assertEqual(
                check_analysis.check("ordinary", root), [],
                "선언이 없는 빌림에 새 오류를 더했다",
            )


class AFailingStepFiveSaysWhichWindowRequiredThem(unittest.TestCase):
    """실패 출력이 "왜 이 함수가 요구되는가"를 말해야 한다 (task 3.3).

    2026-09-10 실측: a074 는 324줄 중 316줄이 `missing evidence for modified function`
    이고 **비교 창을 말하는 줄이 0** 이다. a076 은 이름 316개를 쉼표로 이은
    **21,838자짜리 한 줄**이고 역시 0 이다. 그 316개는 전부 남의 change 함수인데
    출력만 봐서는 알 길이 없다.

    기제는 `main` 의 early return 이다 — `if errors: … return 1` 이 착지·요구 수를
    찍는 자리를 건너뛴다. task 1.9 가 적은 "항상 출력한다"는 성공할 때만 참이었다."""

    def _failing_fixture(self, raw: tempfile.TemporaryDirectory, *, with_evidence: bool):
        """base P → 내 작업 L → 이웃 change N (→ 증거 E). 착지 기록은 없다."""
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        other = root / "internal" / "other.go"
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["L"] = _commit_all(root, "L: my work lands")
        other.write_text("package internal\nfunc Other() int { return 2 }\n")
        marks["N"] = _commit_all(root, "N: a different change lands")
        if with_evidence:
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            _commit_all(root, "E: evidence for my own function only")
        return root, marks

    def test_a_window_that_could_not_be_derived_is_not_printed(self) -> None:
        """못 잰 창은 찍지 않는다 (task 4.1 에서 실물로 나왔다).

        착지 해소가 실패하면 `landing` 과 `required_count` 는 아예 채워지지 않는다.
        옛 판본은 그 빈칸을 기본값으로 찍어서 `working tree … required 0 function(s)`
        라고 **거짓말**을 했다 — 대상도 아니고 세지도 않은 값이다. 3.3 이 창을 찍게
        만든 이유가 "이름만 있고 이유가 없다"였는데, 지어낸 이유는 그보다 나쁘다.

        2026-09-11 실측: 아카이브된 a076(번들 0)에 착지를 주니 정확히 그 줄이 나왔다.
        실패 사유 줄은 그대로 있어야 하므로 아래에서 같이 단언한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(raw, with_evidence=False)
            change = root / "openspec" / "changes" / "mine"
            # 번들이 0 인 change 가 착지를 선언하면 3.2.3.1 이 거절한다 — 착지 해소가
            # 실패하는 실물 경로다.
            _declare_landing(change, marks["L"], "declare a landing nothing pins")
            code, output = self._main_output(root)
            self.assertEqual(code, 1, output)
            self.assertIn("pinned by no", output)
            self.assertNotIn("working tree", output)
            self.assertNotIn("required 0 function(s)", output)

    def _main_output(self, root: Path) -> tuple[int, str]:
        output = io.StringIO()
        with mock.patch.object(
            sys, "argv", ["check_analysis.py", "--change", "mine", "--root", str(root)]
        ), redirect_stdout(output):
            code = check_analysis.main()
        return code, output.getvalue()

    def test_a_failure_prints_the_comparison_window(self) -> None:
        """실패해도 base·대상·요구 수를 찍어야 한다. 오늘은 한 줄도 안 찍는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(raw, with_evidence=True)
            code, printed = self._main_output(root)
            self.assertEqual(code, 1, f"이 픽스처는 실패해야 한다: {printed}")
            self.assertIn(
                "missing evidence for modified function internal/other.go:Other", printed,
                "양성 대조: 남의 함수가 실제로 요구되고 있어야 한다",
            )
            self.assertIn(marks["P"][:12], printed, f"base 를 말해야 한다: {printed}")
            self.assertIn("working tree", printed, f"대상이 무엇인지 말해야 한다: {printed}")
            self.assertIn("required 2 function(s)", printed, f"요구 수를 말해야 한다: {printed}")

    def test_a_failure_says_how_much_of_the_window_is_not_this_change(self) -> None:
        """"남의 것"을 **숫자로** 말해야 한다 — base 뒤에 착지한 커밋 수.

        지어내지 않고 `git rev-list --count <base>..HEAD` 로 잰 값이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(raw, with_evidence=True)
            expected = subprocess.check_output(
                ["git", "rev-list", "--count", f"{marks['P']}..HEAD"], cwd=root, text=True
            ).strip()
            self.assertEqual(expected, "3", "픽스처가 L·N·E 셋을 담아야 한다")
            code, printed = self._main_output(root)
            self.assertIn("landed-commit.txt", printed, f"무엇을 하면 좁아지는지 말해야 한다: {printed}")
            self.assertIn(f"{expected} commit(s)", printed, f"창의 크기를 숫자로 말해야 한다: {printed}")
            # 상수를 죽인다. 커밋을 하나 더 얹으면 숫자도 하나 늘어야 한다 —
            # 안 그러면 "3" 이라고 적어 둔 구현이 위 단언을 통과한다.
            (root / "unrelated.md").write_text("someone else lands\n")
            _commit_all(root, "another change lands after the base")
            code, again = self._main_output(root)
            self.assertIn("4 commit(s)", again, f"재서 쓴 값이 아니다: {again}")

    def test_the_missing_map_message_carries_the_count_and_the_window(self) -> None:
        """번들이 아예 없는 경로의 메시지(이름 316개를 이어 붙이던 그 한 줄)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(raw, with_evidence=False)
            errors = check_analysis.check("mine", root)
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("missing Function Logic Map", errors[0])
            self.assertIn("2 function(s)", errors[0], f"몇 개인지 메시지 안에 있어야 한다: {errors[0][:200]}")
            self.assertIn(marks["P"][:12], errors[0], f"어느 base 인지 메시지 안에 있어야 한다: {errors[0][:200]}")
            self.assertIn("working tree", errors[0], f"어느 대상인지 메시지 안에 있어야 한다: {errors[0][:200]}")

    def test_the_success_line_is_unchanged(self) -> None:
        """기록 스무 곳이 인용하는 **성공** 줄은 글자 그대로 남아야 한다.

        a074·a077·a079·a092·a112 의 review 가 이 문자열을 인용한다. 설명을 늘리는
        태스크가 그 인용을 깨면 안 된다 — 회귀 핀이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            own.write_text("package internal\nfunc Own() int { return 1 }\n")
            base = _commit_all(root, "P: base")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text(
                f"# review\n\n{check_analysis.EXEMPTION}\n", encoding="utf-8"
            )
            _commit_all(root, "docs only, nothing to require")
            code, printed = self._main_output(root)
            self.assertEqual(code, 0, printed)
            self.assertIn("mine: evidence complete or diff-proven exempt", printed)


class AnEmptyRequiredSetIsAnnouncedNotSwallowed(unittest.TestCase):
    """요구 집합이 비었는데 번들이 있으면 그 사실을 말해야 한다(task 2.6·3.4).

    조용한 통과는 증거로 통과한 것과 구분되지 않는다. 그리고 그 출력은 `context`
    로 내야 한다 — `check` 의 반환 리스트에 붙이면 `main()` 이 1을 돌려 5단계가
    **실패**한다."""

    def test_empty_required_with_bundles_is_reported_and_still_passes(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            # 작업이 base **앞**에 있다 — 2026-08-04 재기준화가 a074~a079 를 그 모양으로
            # 만들었다. base..착지 의 Go diff 가 비고, 증거는 그 리비전을 기술한다.
            own.write_text("package internal\nfunc Own() int { return 2 }\n")
            base = _commit_all(root, "P: base already contains the work")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text("mine\n")
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            (root / "docs.md").write_text("post-deployment measurement\n")
            landing = _commit_all(root, "L: docs only, no Go change")
            (change / "landed-commit.txt").write_text(landing + "\n")
            _commit_all(root, "record the landing point")

            context: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, context), [])
            self.assertEqual(context.get("required_count"), 0)
            output = io.StringIO()
            with mock.patch.object(
                sys, "argv", ["check_analysis.py", "--change", "mine", "--root", str(root)]
            ), redirect_stdout(output):
                self.assertEqual(check_analysis.main(), 0)
            printed = output.getvalue()
            self.assertIn(landing, printed, f"해소한 착지 SHA 가 기록에 남아야 한다: {printed}")


if __name__ == "__main__":
    unittest.main()
