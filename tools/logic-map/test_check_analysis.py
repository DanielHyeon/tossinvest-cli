from __future__ import annotations

import ast
import json
import hashlib
import io
import os
import subprocess
import sys
import tempfile
import traceback
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
    """git 을 mock 하고 번들 검증만 잰다. 루트는 **커밋 하나뿐인 저장소**다 (task 7.5.1 · 7.5.2.1).

    이 헬퍼는 `resolve_base` · `changed_existing_functions` 를 mock 해서 git 을 빼는데
    `_landing_record` 만 빠뜨렸다 — 저장소가 아닌 곳에서 git 이 rc≠0 으로 죽어도 예전에는
    조용히 `None`("기록 없음")이라 무해해 보였다. 이제 rc≠0 은 결함이다. 빈 저장소는 같은 전제를
    진짜로 만든다(실측: `HEAD:x missing` rc 0). 7.5.2.1 부터 명령이 먼저 `HEAD` 를 sha 로 풀므로
    빈 커밋 하나를 둔다 — 태어나지 않은 `HEAD` 는 결함이다."""
    _empty_repo(root)
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
            text = check_analysis._bundle_text(bundle, (bundle / "ast.json").read_bytes())
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
                ["function-logic reference base is invalid: change is neither open nor archived: reference"],
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
            # 문장은 **상수 하나**다 (task 7.6, 리뷰 I5). 기록 경로와 한 벌씩 들고 있던
            # 동안 이미 "ends" / "already ends" 로 갈렸다.
            self.assertEqual(errors, [check_analysis.ADOPTION_REFUSES_A_LANDING])

    def test_the_recorder_refuses_the_adoption_path_too(self) -> None:
        """`check` 의 거절만 있고 **기록 경로**의 거절을 재는 시험이 0 이었다(변이 M9).

        판정 둘이 서로를 덮는 모양이다 — 기록 경로가 값을 써 버리면 그 다음 5단계가
        거절하지만, 그때는 이미 감사 밖의 두 번째 손잡이가 파일로 존재한다."""
        raw, root, p, e = self._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            code, lines = check_analysis.record_landing(adoption.CHANGE, root)
            self.assertEqual(
                (code, lines),
                (1, [f"{adoption.CHANGE}: {check_analysis.ADOPTION_REFUSES_A_LANDING}"]),
            )
            self.assertFalse(
                (root / "openspec" / "changes" / adoption.CHANGE
                 / check_analysis.LANDING_FILE).exists(),
                "거절했으면 파일도 없어야 한다",
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
        """자리표시자 번들은 **그 가드의 문장**으로 거절된다 (task 7.5.2 — 재리뷰의 testing 전문가).

        이 시험은 저장소가 아닌 임시 디렉터리에서 `len(errors) > 0` 만 봤다. 7.5.1 이 git 실패를
        결함으로 올리자 `cannot derive … not a git repository` 한 줄이 그 단언을 만족해서,
        자리표시자 가드를 **지워도 초록**이었다(변이로 증명). 커밋 하나뿐인 저장소에서 돌리고, 결함 줄이
        없음과 가드마다의 문장을 단언한다 ([[passing-test-is-not-evidence]])."""
        with tempfile.TemporaryDirectory() as tmp:
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.2 · 7.5.2.1
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
                errors = check_analysis.check("change", Path(tmp))
            self.assertIn("pkg--run: ast.json is placeholder evidence", errors)
            self.assertTrue(any("pkg--run: function-logic-map.md still contains TODO" == error
                                for error in errors), errors)
            self.assertFalse(any("cannot derive" in error for error in errors), errors)

    def test_a_partly_filled_ast_is_still_a_placeholder(self) -> None:
        """자리표시자 판정은 **하나라도** 비면 선다(`any`). 전부 비어야 서는 `all` 로 바뀌어도
        `{}` 로만 재는 시험은 초록이다 — 채운 칸이 섞인 모양을 따로 둔다."""
        with tempfile.TemporaryDirectory() as tmp:
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.2 · 7.5.2.1
            target = Path(tmp) / "openspec" / "changes" / "change" / "analysis" / "function-logic" / "pkg--run"
            target.mkdir(parents=True)
            (target / "ast.json").write_text(json.dumps({
                "file": "pkg/run.go", "source_sha256": "0" * 64, "package": "pkg",
                "function": "Run", "signature": "", "start": {"line": 1}, "end": {"line": 2},
            }), encoding="utf-8")
            with mock.patch("check_analysis.resolve_base", return_value="base"), \
                    mock.patch("check_analysis.changed_existing_functions", return_value={}):
                errors = check_analysis.check("change", Path(tmp))
            self.assertIn("pkg--run: ast.json is placeholder evidence", errors)

    def test_completed_bundle_is_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
            root = Path(tmp)
            write_bundle(root, branches=[])
            self.assertEqual(run_check(root), [])

    def test_prose_branch_count_must_match_the_ast_branch_count(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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
            _empty_repo(tmp)  # 커밋 하나뿐인 저장소 — 7.5.1 · 7.5.2.1
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


def _commit_at(root: Path, subject: str, when: str) -> str:
    """커밋 시각을 **고정**해서 커밋한다. `git log -1` 은 커밋 시각 순으로 고르므로,
    시각이 같은 초에 겹치면 두 가지 중 어느 쪽이 하한인지가 실행마다 달라진다."""
    env = {**os.environ, "GIT_AUTHOR_DATE": when, "GIT_COMMITTER_DATE": when}
    subprocess.run(["git", "add", "."], cwd=root, check=True, env=env)
    subprocess.run(["git", "commit", "-qm", subject], cwd=root, check=True, env=env)
    return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()


def _commit_all_allowing_nothing(root: Path, subject: str) -> str:
    """추적 파일이 하나도 안 바뀌어도 커밋한다 — 갈림점 뒤에 **자리**만 만드는 커밋."""
    subprocess.run(["git", "add", "."], cwd=root, check=True)
    subprocess.run(["git", "commit", "-q", "--allow-empty", "-m", subject], cwd=root, check=True)
    return subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()


def _bundles(root: Path, analysis: Path) -> list:
    """생산 경로가 한 번 재서 넘기는 고정 번들 목록 (task 7.5). 시험이 사설 판정을
    직접 부를 때 같은 것을 넘긴다 — 여기서 다른 것을 만들면 시험이 생산과 다른 입력을 잰다."""
    return check_analysis._select_pinning(root, check_analysis._read_evidence(analysis))


def _head(root: Path) -> str:
    """명령이 푸는 것과 같은 `HEAD` 의 sha (task 7.5.2.1). 사설 판정을 직접 부르는 시험이 같은 끝을 넘긴다."""
    return check_analysis._head_commit(root)


def _inputs(root: Path, analysis: Path, **replaced) -> "check_analysis.LandingInputs":
    """생산 경로가 한 번 재서 넘기는 착지 입력 한 벌 (task 7.5.2.1 — 사설 판정에 스스로 읽는 기본값이
    없어졌다). 시험이 하한 · 수리 신호를 골라 넣을 때만 `replaced` 로 바꾼다 — 나머지는 생산과 같은
    읽기에서 온다."""
    return check_analysis._measure_landing_inputs(
        root, _head(root), check_analysis._read_evidence(analysis))._replace(**replaced)


def _empty_repo(root: Path | str) -> None:
    """커밋 **하나**(빈 것)만 있는 저장소 (task 7.5.2.1). 명령은 시작하자마자 `HEAD` 를 sha 로 푼다 —
    태어나지 않은 `HEAD` 는 결함이다(review.md 가 거부하는 정상 입력으로 적었다). git 을 mock 하는 시험의
    전제("HEAD 에 착지 기록이 없다")는 그대로다: 빈 커밋에도 기록은 없다."""
    subprocess.run(["git", "init", "-q"], cwd=root, check=True)
    subprocess.run(["git", "-c", "user.email=a122@example.invalid", "-c", "user.name=a122",
                    "commit", "-q", "--allow-empty", "-m", "empty"], cwd=root, check=True)


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


def _landed_work_fixture() -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
    """R → P(base) → L(작업) → 증거 커밋. 오늘 `check()` 가 `[]` 인 정직한 change.

    클래스 셋이 나눠 쓴다. `AForgedLandingPointIsRefusedByName` 의 사적 메서드였고
    다른 클래스 둘이 **그 클래스를 인스턴스로 만들어** 불렀다 (리뷰 I16). 픽스처를
    한 클래스가 소유하는 한 그 결합은 없앨 수 없으므로 소유자를 모듈로 올렸다.
    """
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


def _base_shaped_forgery_fixture() -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
    """R → P(base) → L(**진짜** Go 작업) → E(base 상태를 기술하는 증거).

    4.4 가 실측으로 재현한 위조 1번의 모양이다. 저자가 고른 것은 번들의 해시 하나뿐이고
    나머지는 전부 진짜 경로다. 선언 경로(`resolve_landing`)와 계산 경로
    (`compute_landing`·`record_landing`)가 **둘 다** 이 입력을 받는다.
    """
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

    def test_the_fixture_passes_before_any_landing_record(self) -> None:
        """양성 대조군. 이것이 초록이 아니면 아래 넷은 아무것도 재지 못한다."""
        raw, root, marks = _landed_work_fixture()
        with raw:
            self.assertEqual(check_analysis.check("mine", root), [])

    def _refuse(self, value: str, needle: str, prepare=None) -> None:
        """바늘은 그 가드의 **자기 문장**이다 (task 6.3).

        예전 바늘 `"landing"` 은 착지 관련 오류 문장 **전부**에 들어 있어서, 가드를 하나
        지워도 뒤의 다른 가드가 거절하고 시험은 초록으로 남았다 — 2026-09-11 변이 실측으로
        가드 넷(40-hex · 커밋 실재 · `landing ≤ HEAD` · `base ≤ landing`)이 전부 SURVIVED.
        이 파일은 같은 함정을 한 자리(`test_a_landing_the_evidence_does_not_describe`)에서
        이미 한 번 고쳤다. 나머지 넷에 같은 처방을 한다."""
        raw, root, marks = _landed_work_fixture()
        with raw:
            if prepare is not None:
                prepare(root, marks)
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
        self._refuse("f" * 40, "is not a commit in this repository")

    def test_a_landing_that_never_landed_on_this_history(self) -> None:
        """곁가지 커밋은 **실재**하지만 이 역사의 조상이 아니다 (task 6.3).

        이 가드(`landing ≤ HEAD`)에는 시험이 아예 없었다 — 곁가지 커밋을 만드는 픽스처가
        0 이었다. 지워도 뒤의 증거 판정이 대신 거절하므로 느슨한 바늘로는 영영 못 잰다."""
        def side_branch(root: Path, marks: dict[str, str]) -> None:
            home = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
            ).strip()
            subprocess.run(["git", "checkout", "-q", "-b", "side", marks["P"]], cwd=root, check=True)
            (root / "side.txt").write_text("side\n")
            marks["S"] = _commit_all(root, "S: a commit on a side branch")
            subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
        self._refuse("{S}", "never landed on this history", prepare=side_branch)

    def test_a_landing_before_the_base(self) -> None:
        self._refuse("{R}", "precedes the comparison base")

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
        self._refuse("HEAD", "must be a full 40-hex commit id")

    def test_an_empty_record_is_still_a_declaration(self) -> None:
        """빈 파일은 **선언이 없는 것**이 아니다.

        task 1.8 이 선언을 읽는 자리를 헬퍼로 뽑았다. 거기서 "파일 없음"과 "빈
        선언"을 안 가르면, 빈 파일을 커밋한 change 가 조용히 워킹트리를 대상으로
        삼고 아무 사유도 안 남는다."""
        self._refuse("", "must be a full 40-hex commit id")



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

    def test_the_forgery_is_red_without_a_landing_record(self) -> None:
        """음성 대조군. 선언이 없으면 같은 입력이 오늘도 빨갛다.

        이것이 빨갛지 않으면 아래 시험은 "선언이 길을 연다"를 재지 못하고 그냥
        깨진 픽스처를 재게 된다."""
        raw, root, _ = _base_shaped_forgery_fixture()
        with raw:
            errors = check_analysis.check("mine", root)
            # 바늘 없이 "빨갛다"만 보면 픽스처가 **다른 이유로** 깨져도 초록이다
            # (리뷰 I15). 음성 대조군이 재야 하는 것은 "오늘도 막힌다"가 아니라
            # "오늘은 **이 사유로** 막힌다"이다.
            self.assertTrue(
                any("AST source hash is stale" in error for error in errors), errors
            )

    def test_evidence_written_at_the_base_cannot_declare_the_base(self) -> None:
        raw, root, marks = _base_shaped_forgery_fixture()
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
        raw, root, marks = _landed_work_fixture()
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
                any("no ordinary commit on this history adds" in error for error in errors),
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
            before: dict[str, object] = {}
            self.assertEqual(
                check_analysis.check("mine", root, before), [], "아카이브 전에는 통과해야 한다"
            )
            self.assertEqual(before["landing"], landing, "아카이브 전에는 기록이 대상이다")
            archive = root / "openspec" / "changes" / "archive" / "2026-09-11-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            subprocess.run(["git", "mv", "openspec/changes/mine", archive.relative_to(root).as_posix()],
                           cwd=root, check=True)
            _commit_all(root, "archive the change")
            after: dict[str, object] = {}
            self.assertEqual(
                check_analysis.check("mine", root, after), [],
                "아카이브 뒤에도 같은 id 로 재검사돼야 한다",
            )
            # 판정이 `[]` 인 것만 보면 이 시험은 착지가 **안 읽혀도** 초록이다 — 이
            # 픽스처는 워킹트리를 대상으로 삼아도 통과하기 때문이다(리뷰 I12, 변이 M4).
            # 재는 것은 판정이 아니라 "아카이브 뒤에도 그 기록이 대상이었는가"다.
            self.assertEqual(
                after["landing"], landing, "아카이브가 기록을 안 보이게 만들면 안 된다"
            )


class TheJudgedBundleMustBeTheCommittedBundle(unittest.TestCase):
    """판정이 **읽은** 번들이 착지 커밋에 그대로 있어야 한다 (task 7.2, 리뷰 C3).

    하한(`_evidence_floor`)은 번들의 **경로**를 보고, 판정(`validate_target`·
    `_pinning_bundles`)은 **워킹트리의 내용**을 읽는다. 이 둘이 갈리는 자리마다
    저자가 하한을 안 움직이고 증거를 바꿀 수 있다 — 자리표시자를 일찍 커밋해 두고
    나중에 워킹트리에서 갈아 끼우거나, 번들 파일을 감시 밖 파일로 향하는 심링크로
    두거나.

    2026-09-12 전수 실측으로 고른 규칙이다: 활성·아카이브 76건에서 **거부 0**.
    범위를 번들의 산문 파일까지 넓히면 3건이 착지를 잃어서(a092·a043·a096 — 전부
    이웃 change 가 나중에 단 무효화 배너) `ast.json` 하나로 묶는다. 기록은 review.md
    `## MEASURE — task 7.2 후보 실측` · `## DECIDE — task 7.2`.
    """

    def _placeholder_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """자리표시자를 base 커밋에 넣고, 진짜 번들은 워킹트리에서만 갈아 끼운다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        bundle = change / "analysis" / "function-logic" / "internal--own"
        bundle.mkdir(parents=True)
        # scaffold 가 만드는 빈 자리. 이 커밋이 하한이 된다 — 내용은 아무것도 고정하지 않는데도.
        (bundle / "ast.json").write_text("{}\n", encoding="utf-8")
        marks = {"P": _commit_all(root, "P: base, with the analysis directory scaffolded")}
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        at_base = hashlib.sha256(own.read_bytes()).hexdigest()
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work lands")
        # 워킹트리에서만 갈아 끼운다. 커밋하면 하한이 여기로 올라와서 위조가 안 된다.
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=at_base,
        )
        return raw, root, marks

    def _declare_only_the_landing(self, root: Path, change: str, value: str) -> None:
        """`landed-commit.txt` **만** 커밋한다 — `git add .` 은 위조를 커밋해 버린다."""
        (root / "openspec" / "changes" / change / "landed-commit.txt").write_text(value + "\n")
        subprocess.run(
            ["git", "add", f"openspec/changes/{change}/landed-commit.txt"], cwd=root, check=True
        )
        subprocess.run(["git", "commit", "-qm", "declare the landing"], cwd=root, check=True)

    def test_the_replacement_is_red_without_a_landing_record(self) -> None:
        """음성 대조군. 선언이 없으면 같은 입력이 오늘도 빨갛다."""
        raw, root, _ = self._placeholder_fixture()
        with raw:
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("AST source hash is stale" in error for error in errors), errors
            )

    def test_a_placeholder_committed_early_cannot_pin_a_landing(self) -> None:
        """C3 — 커밋된 것은 `{}`, 판정이 읽은 것은 base 를 기술하는 번들."""
        raw, root, marks = self._placeholder_fixture()
        with raw:
            self._declare_only_the_landing(root, "mine", marks["P"])
            errors = check_analysis.check("mine", root)
            self.assertTrue(errors, "커밋 안 된 번들로 base 를 선언하면 거절해야 한다")
            self.assertTrue(
                any("does not hold the evidence this verdict read" in error for error in errors),
                f"사유를 이 가드의 자기 문장으로 말해야 한다: {errors}",
            )

    def _symlink_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, str]:
        """`ast.json` 이 감시 경로 **밖**을 가리키는 심링크. 대상은 한 번 고쳐진다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        at_base = hashlib.sha256(own.read_bytes()).hexdigest()
        bundle = _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=at_base,
        )
        target = change / "analysis" / "real-ast.json"
        shutil.move(str(bundle / "ast.json"), target)
        (bundle / "ast.json").symlink_to(Path("..") / ".." / "real-ast.json")
        base = _commit_all(root, "P: base, evidence symlinked out of its bundle")
        (change / "base-commit.txt").write_text(base + "\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        _commit_all(root, "W: this change's Go work lands")
        # 대상 파일만 고친다. 감시 경로(`…/internal--own/ast.json`)는 그대로다.
        value = json.loads(target.read_text(encoding="utf-8"))
        value["signature"] = "Own(params=0, results=1) // rewritten out of sight"
        target.write_text(json.dumps(value), encoding="utf-8")
        _commit_all(root, "T: rewrite the target the symlink points at")
        return raw, root, base

    def test_a_bundle_symlinked_out_of_the_watched_path_cannot_pin_a_landing(self) -> None:
        """C3 — 하한이 지켜보는 파일과 판정이 읽는 파일이 다르다.

        `ast.json` 이 감시 경로 **밖**을 가리키는 심링크면 대상 파일을 몇 번을 고쳐도
        하한은 안 움직인다."""
        raw, root, base = self._symlink_fixture()
        with raw:
            self._declare_only_the_landing(root, "mine", base)
            errors = check_analysis.check("mine", root)
            self.assertTrue(errors, "감시 밖 파일을 읽는 번들은 착지를 고정하지 못한다")
            self.assertTrue(
                any("does not hold the evidence this verdict read" in error for error in errors),
                f"사유를 이 가드의 자기 문장으로 말해야 한다: {errors}",
            )

    def test_the_gate_will_not_compute_a_landing_it_cannot_hold(self) -> None:
        """계산 경로에도 **같은** 조건이 선다.

        규칙이 `resolve_landing` 에만 있으면 도구가 자기가 거절할 값을 계산해서 기록한다.
        같은 규칙이 두 집에 살면 한쪽만 고쳐도 양쪽 시험이 초록이 되므로
        ([[two-judgements-cover-for-each-other]]) 계산 쪽도 따로 못 박는다. 워킹트리는
        깨끗하다 — `record_landing` 의 더러운 트리 거절이 이 자리를 가리지 않는다."""
        raw, root, _ = self._symlink_fixture()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, f"들고 있지 않은 증거로 착지를 계산하면 안 된다: {lines}")
            self.assertTrue(
                any("no commit holds the evidence this verdict read" in line for line in lines),
                f"사유를 이 가드의 자기 문장으로 말해야 한다: {lines}",
            )
            self.assertFalse((root / "openspec" / "changes" / "mine" / "landed-commit.txt").exists())

    def test_an_honest_bundle_still_pins_its_landing(self) -> None:
        """양성 대조군. 커밋된 번들과 워킹트리가 같으면 아무것도 안 막는다.

        이것이 빨개지면 규칙이 정상 입력을 죽인 것이고, 실측한 '거부 0' 이 거짓이 된다."""
        raw, root, marks = _landed_work_fixture()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            landing = subprocess.check_output(
                ["git", "rev-parse", "HEAD"], cwd=root, text=True
            ).strip()
            _declare_landing(change, landing, "record the landing the gate computed")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts["required_count"], 1)


class AnEvidenceRewriteHiddenInAMoveStillRaisesTheFloor(unittest.TestCase):
    """옮기면서 **고친** 증거는 하한을 올린다. 그대로 옮긴 것만 건너뛴다 (task 7.2, H1).

    하한이 rename 을 건너뛰는 이유는 하나다 — 아카이브가 번들을 통째로 옮기고, 그
    이동을 하한으로 세면 이미 유효했던 기록이 무효가 되기 때문이다. 그런데 `-M` 의
    기본 유사도는 50% 라 **내용을 고치면서** 옮긴 것도 같은 rename 으로 보고 건너뛰었고,
    `--diff-filter=MA` 는 정규 파일↔심링크 전환(`T`)을 아예 안 셌다. 둘 다 내용 교체인데
    하한이 안 움직인다.

    시험이 하한 함수 자체를 부르는 이유: 종단(`check`)에서는 같은 입력을 C3 등식
    (`_unheld_bundles`)이 **먼저** 거절한다. 그 위에서 재면 이 시험은 하한이 아니라 다른
    가드를 재게 된다([[first-failure-is-not-the-fix-scope]]). 하한의 계약은 "고정 번들이
    역사에 마지막으로 들어오거나 바뀐 커밋"이고, 그 계약을 그 층에서 못 박는다.

    2026-09-12 전수 실측: 하한이 움직이는 change 1건, 착지를 잃는 change 0건.
    """

    def _committed_change(self, root: Path) -> tuple[Path, str]:
        """base 커밋에 번들이 이미 있는 change. `(change 경로, base)`."""
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        base = _commit_all(root, "P: base, with the evidence already committed")
        (change / "base-commit.txt").write_text(base + "\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        _commit_all(root, "W: this change's Go work lands")
        return change, base

    def test_a_bundle_rewritten_where_it_stands_moves_the_floor(self) -> None:
        """`M` — 옮기지 않고 그 자리에서 고친 번들. 가장 흔한 모양인데 시험이 없었다.

        `--diff-filter` 에서 `M` 이 빠져도 아무 시험이 안 빨개졌다(리뷰 I11, 변이 M3).
        아래 두 시험은 **이동**(rename)을 재므로 `A`·`T` 만으로도 초록이다 — 수정
        하나만 세는 자리는 여기다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            change, _ = self._committed_change(root)
            analysis = change / "analysis" / "function-logic"
            added = check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root))
            ast = analysis / "internal--own" / "ast.json"
            value = json.loads(ast.read_text(encoding="utf-8"))
            value["source_sha256"] = hashlib.sha256(
                (root / "internal" / "own.go").read_bytes()
            ).hexdigest()
            ast.write_text(json.dumps(value), encoding="utf-8")
            rewritten = _commit_all(root, "rewrite the bundle where it stands")
            self.assertNotEqual(added, rewritten, "픽스처가 두 커밋을 갈라야 한다")
            self.assertEqual(
                check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)), rewritten,
                "그 자리에서 고친 증거가 하한이어야 한다",
            )

    def test_an_archive_that_also_rewrites_the_bundle_moves_the_floor(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            change, base = self._committed_change(root)
            archive = root / "openspec" / "changes" / "archive" / "2026-09-11-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            subprocess.run(
                ["git", "mv", "openspec/changes/mine", archive.relative_to(root).as_posix()],
                cwd=root, check=True,
            )
            # 옮기는 **같은 커밋**에서 번들을 다시 쓴다. 고정하는 리비전은 그대로다.
            ast = archive / "analysis" / "function-logic" / "internal--own" / "ast.json"
            value = json.loads(ast.read_text(encoding="utf-8"))
            value["signature"] = "Own(params=0, results=1) // rescaffolded"
            ast.write_text(json.dumps(value), encoding="utf-8")
            moved = _commit_all(root, "archive the change and rewrite its bundle in one commit")
            self.assertEqual(
                check_analysis._evidence_floor(root, _bundles(root, archive / "analysis" / "function-logic"), _head(root)),
                moved,
                "이동에 숨긴 재작성이 하한이어야 한다",
            )

    def test_a_pure_archive_move_still_does_not_move_the_floor(self) -> None:
        """양성 대조군. 그대로 옮긴 것까지 세면 a099 의 실물 기록이 무효가 된다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            change, base = self._committed_change(root)
            before = check_analysis._evidence_floor(root, _bundles(root, change / "analysis" / "function-logic"), _head(root))
            archive = root / "openspec" / "changes" / "archive" / "2026-09-11-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            subprocess.run(
                ["git", "mv", "openspec/changes/mine", archive.relative_to(root).as_posix()],
                cwd=root, check=True,
            )
            _commit_all(root, "archive the change, byte for byte")
            self.assertEqual(
                check_analysis._evidence_floor(root, _bundles(root, archive / "analysis" / "function-logic"), _head(root)),
                before,
                "그대로 옮긴 것은 하한이 아니다",
            )

    def test_turning_a_bundle_into_a_symlink_moves_the_floor(self) -> None:
        """`T` — 정규 파일을 심링크로 바꾸면 읽히는 내용이 통째로 바뀐다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            change, base = self._committed_change(root)
            bundle = change / "analysis" / "function-logic" / "internal--own"
            target = change / "analysis" / "real-ast.json"
            shutil.move(str(bundle / "ast.json"), target)
            (bundle / "ast.json").symlink_to(Path("..") / ".." / "real-ast.json")
            switched = _commit_all(root, "T: the bundle becomes a symlink")
            self.assertEqual(
                check_analysis._evidence_floor(root, _bundles(root, change / "analysis" / "function-logic"), _head(root)),
                switched,
                "정규 파일↔심링크 전환도 내용 교체다",
            )


class TheFloorDoesNotChangeWithTheDevelopersGitConfig(unittest.TestCase):
    """하한은 저장소의 함수여야 한다 — 그 기계의 git 설정의 함수가 아니라 (task 7.2, H2).

    `log.follow = true` 는 흔한 개인 설정이고, 경로가 **하나**일 때 `git log` 를
    `--follow` 로 만든다. 그러면 rename 이 짝지어져 `--diff-filter=MA` 에서 빠지고
    하한이 이동 **이전**으로 내려간다. 같은 저장소·같은 선언이 기계마다 다른 판정을
    받는다. 게이트는 자기가 부르는 git 의 설정을 지운다.

    2026-09-12 전수 실측: 거부 0.
    """

    def _renamed_bundle_fixture(
        self, *, describes: str = "base"
    ) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """번들이 한 번 **그대로** 이름을 바꾼 change. 하한 경로가 하나라 `log.follow` 가 켜진다.

        `describes` 는 번들이 어느 리비전을 기술하는가다. `"base"` 는 위조 — 증거가
        base 커밋에 이미 있고 base 상태를 기술하므로, 하한만 내려가면 창이 빈다.
        `"current"` 는 정직한 change 다. 둘 다 있어야 규칙이 무엇을 거부하고 무엇을
        통과시키는지 갈린다.
        """
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        at_base = hashlib.sha256(own.read_bytes()).hexdigest()
        if describes == "base":
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=at_base,
            )
        marks = {"P": _commit_all(root, "P: base")}
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work lands")
        if describes != "base":
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            marks["E"] = _commit_all(root, "E: evidence enters history under its first name")
        logic = change / "analysis" / "function-logic"

        def move(source: str, destination: str, subject: str) -> str:
            subprocess.run(
                ["git", "mv", (logic / source).relative_to(root).as_posix(),
                 (logic / destination).relative_to(root).as_posix()],
                cwd=root, check=True,
            )
            return _commit_all(root, subject)

        # 왕복이다. 한쪽 방향만 옮기면 오늘의 경로가 base 에 없어서 C3 등식이 **먼저**
        # 거절하고, 그러면 이 시험은 하한이 설정에 흔들리는 것을 못 재고 다른 가드를
        # 재게 된다. 제자리로 돌아와야 base 에 같은 경로·같은 바이트가 있고, 갈리는
        # 것이 하한 하나만 남는다 — 이 설정 의존성의 **최소 증인**이다.
        move("internal--own", "internal--own-moved", "R1: the bundle moves aside")
        marks["R"] = move("internal--own-moved", "internal--own", "R2: and moves back")
        return raw, root, marks

    def test_a_stale_landing_is_refused_on_a_log_follow_machine(self) -> None:
        raw, root, marks = self._renamed_bundle_fixture()
        with raw:
            subprocess.run(["git", "config", "log.follow", "true"], cwd=root, check=True)
            change = root / "openspec" / "changes" / "mine"
            _declare_landing(change, marks["P"], "declare the base as the landing")
            errors = check_analysis.check("mine", root)
            self.assertTrue(errors, "개인 git 설정이 하한을 내리면 안 된다")
            self.assertTrue(
                any("precedes the evidence that pins it" in error for error in errors),
                f"사유를 하한 가드의 자기 문장으로 말해야 한다: {errors}",
            )

    def test_the_same_repository_gets_the_same_verdict_either_way(self) -> None:
        """같은 저장소·같은 선언이 설정 하나로 갈리면 안 된다."""
        verdicts = []
        for follow in ("false", "true"):
            raw, root, marks = self._renamed_bundle_fixture()
            with raw:
                subprocess.run(["git", "config", "log.follow", follow], cwd=root, check=True)
                _declare_landing(
                    root / "openspec" / "changes" / "mine", marks["P"], "declare the base"
                )
                verdicts.append(bool(check_analysis.check("mine", root)))
        self.assertEqual(verdicts[0], verdicts[1], "판정이 개인 설정에 따라 갈렸다")

    def test_an_honest_landing_still_passes_on_a_log_follow_machine(self) -> None:
        """양성 대조군. 설정을 지우는 것이 정상 선언까지 막으면 여기가 빨개진다."""
        raw, root, marks = self._renamed_bundle_fixture(describes="current")
        with raw:
            subprocess.run(["git", "config", "log.follow", "true"], cwd=root, check=True)
            change = root / "openspec" / "changes" / "mine"
            _declare_landing(change, marks["R"], "declare the rename as the landing")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts["required_count"], 1)


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
            # task 7.2.2 부터는 E 도 기록하지 않는다 — 고정 소스가 base 와 같은 착지는 증거가
            # 작업이 어디 착지했는지 말하지 못한다(사람이 2026-09-14 에 고른 규칙, K2).
            self.assertEqual(code, 1, lines)
            self.assertFalse((root / "openspec" / "changes" / "mine" / "landed-commit.txt").exists())
            self.assertIn(
                f"landing point {marks['E'][:12]} changes none of the sources its evidence pins "
                f"since the comparison base {marks['P'][:12]}",
                lines[0],
            )

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

    def test_uncommitted_evidence_is_named_as_the_reason_nothing_was_recorded(self) -> None:
        """번들이 역사에 없으면 바닥이 없고, 계산 경로는 **자기 문장**으로 그것을 말한다.

        `resolve_landing` 에도 같은 뜻의 가드가 있어서 느슨한 바늘(`"never entered"`)로는
        갈리지 않는다. 그리고 이 가드를 지워도 아래의 "어느 커밋도 안 맞는다"가 대신
        rc 1 을 만들어 스위트가 초록으로 남았다(리뷰 I9, 변이 M7)."""
        raw, root, marks = self._fixture(base_also_matches=False)
        with raw:
            # E 를 되돌려 번들을 **추적되지 않은** 파일로 되돌린다. 추적 파일은 하나도
            # 안 바뀌므로 위의 dirty 거절에 걸리지 않는다 — 걸리면 이 시험은 바닥이
            # 아니라 그 가드를 재게 된다.
            subprocess.run(["git", "reset", "--hard", "-q", marks["W"]], cwd=root, check=True)
            _write_evidence(
                root / "openspec" / "changes" / "mine",
                package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256((root / "internal" / "own.go").read_bytes()).hexdigest(),
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any("commit the bundles in an ordinary commit" in line for line in lines), lines)

    def test_the_computation_names_what_stops_it_before_the_walk(self) -> None:
        """`compute_landing` 은 `record_landing` 말고도 불린다(`resolve_landing`·시험). 기록
        명령이 판정 함수로 **먼저** 멈추게 되면서(task 7.7) 그 경로로는 이 두 반환이 안 닿는다
        — 그러면 이 둘을 지워도 스위트가 초록이다([[two-judgements-cover-for-each-other]])."""
        raw, root, marks = self._fixture(base_also_matches=False)
        with raw:
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            subprocess.run(["git", "reset", "--hard", "-q", marks["W"]], cwd=root, check=True)
            _write_evidence(
                root / "openspec" / "changes" / "mine",
                package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256((root / "internal" / "own.go").read_bytes()).hexdigest(),
            )
            self.assertEqual(
                check_analysis.compute_landing(root, marks["P"], _inputs(root, analysis)),
                ("", "no ordinary commit on this history adds the pinning evidence — commit the bundles "
                     "in an ordinary commit (a merge commit's own changes are not read)"),
            )
            shutil.rmtree(analysis)
            self.assertEqual(
                check_analysis.compute_landing(root, marks["P"], _inputs(root, analysis)),
                ("", "no `revision: current` evidence pins a landing for this change"),
            )

    def test_a_base_it_cannot_resolve_is_named_as_the_base(self) -> None:
        """base 를 못 풀면 **base 라고** 말한다. 이 갈래를 재는 시험이 HEAD 에도 0 이었다
        (task 7.7 변이 R6 SURVIVED). 변이 아래서도 rc 는 1 이었다 — 바깥 결함 경계가
        `no landing recorded — …` 로 받아서다. 기록이 안 쓰이는 것은 우연이 지켰고
        빠진 것은 "무엇을" 못 풀었는지였다([[surviving-mutant-may-mean-accidental-safety]])."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            (root / "openspec" / "changes" / "mine" / "base-commit.txt").write_text("0" * 40 + "\n")
            _commit_all(root, "a base that names no commit")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertEqual(len(lines), 1, lines)
            self.assertTrue(lines[0].startswith("mine: cannot resolve the comparison base: "), lines)

    def test_the_recorder_refuses_the_forgery_this_class_is_named_after(self) -> None:
        """이 클래스는 "도구는 저자가 고를 값을 안 쓴다"를 주장하는데, 위조 픽스처를
        **기록 경로**에 준 적이 한 번도 없었다 (리뷰 I10).

        진짜 작업이 L 에 있고 증거가 base 상태를 기술하면 어느 커밋도 고정을 통과하지
        못한다. 도구는 그때 값을 만들지 않는다 — 만들면 그 값이 곧 위조다."""
        raw, root, marks = _base_shaped_forgery_fixture()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertEqual(len(lines), 1, lines)
            self.assertIn("is accepted as the landing", lines[0])
            self.assertTrue(
                lines[0].endswith("is not the revision this evidence describes: internal/own.go"),
                lines,
            )
            # 옛 꼬리 "the evidence does not describe any revision on this history" 는 **이
            # 픽스처에서 거짓**이었다 — 증거는 base P 를 정확히 기술한다(아래 한 줄). 순회는
            # 하한 이후만 걷고, 사유는 걸은 것만 말한다 (task 6.5).
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            self.assertEqual(check_analysis._pinning_at(root, marks["P"], _bundles(root, analysis)), (1, []))
            self.assertNotIn("any revision on this history", lines[0])

    def test_it_refuses_a_change_that_borrows_its_evidence(self) -> None:
        """빌리는 쪽은 값을 **복사**한다. 두 번째 값을 계산하면 창의 양끝이 갈린다.

        spec 이 자리를 그렇게 정했는데 그 거절을 재는 시험이 0 이었다(변이 M8)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _borrowed_fixture(raw)
            code, lines = check_analysis.record_landing("ordinary", root)
            self.assertEqual(code, 1, lines)
            # 5단계와 **같은 문장**이다 (task 7.2.3). 두 벌이면 갈린다(리뷰 I5 가 이관 문장에서 실측).
            self.assertEqual(lines, [f"ordinary: {check_analysis.BORROWED_REFUSES_A_LANDING}"])
            self.assertFalse(
                (root / "openspec" / "changes" / "ordinary" / "landed-commit.txt").exists(),
                "거절했으면 파일도 없어야 한다",
            )

    def test_it_stops_when_the_id_is_open_and_archived_at_once(self) -> None:
        """모호한 id 는 고르지 않고 멈춘다. 조용히 활성을 고르면 어느 change 에 기록한
        것인지가 아무 데도 안 남는다 (변이 M11).

        옛 판본은 "못 찾음"을 fallback 으로 흘리고 "모호함"만 타입으로 멈췄다. 7.6 이
        fallback 을 없앤 뒤로 둘 다 해소기의 문장으로 멈춘다 — 이 시험은 그 문장을 잰다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            archive = root / "openspec" / "changes" / "archive" / "2026-09-11-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            shutil.copytree(root / "openspec" / "changes" / "mine", archive)
            _commit_all(root, "the same id is open and archived at once")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any("open and archived at once" in line for line in lines), lines)

    def test_an_unknown_change_is_named_instead_of_given_a_base_to_capture(self) -> None:
        """오타 난 id 에 "base 를 capture 하라"고 권하면 **없는 change 를 만들라**는 말이 된다
        (리뷰 I4). 두 경로가 같은 해소기를 쓰므로 같은 문장으로 멈춘다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            code, lines = check_analysis.record_landing("mnie", root)
            self.assertEqual((code, lines), (1, ["change is neither open nor archived: mnie"]))
            self.assertEqual(
                check_analysis.check("mnie", root), ["change is neither open nor archived: mnie"]
            )

    def test_two_archived_copies_are_named_on_both_paths(self) -> None:
        """사본 둘도 해소기의 **자기 문장**으로 멈춘다. 옛 fallback 은 이것을 없는 경로로
        바꿔서 역시 "base 를 capture 하라"로 만들었다."""
        raw, root, _ = self._fixture(base_also_matches=False)
        with raw:
            archive = root / "openspec" / "changes" / "archive"
            archive.mkdir(parents=True, exist_ok=True)
            subprocess.run(
                ["git", "mv", "openspec/changes/mine", "openspec/changes/archive/2026-09-11-mine"],
                cwd=root, check=True,
            )
            shutil.copytree(archive / "2026-09-11-mine", archive / "2026-09-12-mine")
            _commit_all(root, "two archived copies of one id")
            expected = ["archive holds 2 copies of mine: 2026-09-11-mine, 2026-09-12-mine"]
            self.assertEqual(check_analysis.check("mine", root), expected)
            self.assertEqual(check_analysis.record_landing("mine", root), (1, expected))

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


class EvidenceAlreadyWrongInTheCommitThatHoldsItIsNamed(unittest.TestCase):
    """`revision: current` 번들이 **자기를 담은 커밋에서** 이미 틀린 모양 (task 6.5).

    증거를 뽑은 뒤 같은 세션에서 Go 를 한 번 더 고치고 둘을 한 커밋에 담으면 이 상태가
    된다. 실물은 a089(`internal/journal/outbox.go`) · a095(`internal/obs/notifier.go`) —
    번들이 base 에서는 맞고 번들을 담은 커밋 `a30eb35ae` 에서는 안 맞는다.

    도구는 이때 아무것도 기록하지 않는다(6.1.2). 고친 것은 **그 사유**다. 옛 꼬리
    "the evidence does not describe any revision on this history" 는 순회가 걷지 않은
    구간까지 주장했고, 걷기 실패 17건 중 4건(a089 · a095 · 아카이브 둘)에서는 증거가 하한
    **아래** 커밋을 전부 맞게 기술해 실측으로 거짓이었다. 이제 사유는 걸은 것만 말하고,
    무엇이 틀렸는지는 규칙이 **첫 후보**에 준 문장을 그대로 인용한다.
    """

    def _fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """P(base) → E(증거 + 그 뒤 한 번 더 고친 Go, **한 커밋**) → X(이웃이 다른 파일을 고침).

        번들 둘 중 `Own` 은 E 에서 이미 틀리고 `Other` 는 X 에서야 틀린다. 그래서 첫 후보 E
        와 마지막 후보 X 의 거절 문장이 다르다 — 사유가 **어느** 후보의 문장을 인용하는지가
        이 픽스처에서 갈린다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        other = root / "internal" / "other.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        other.write_text("package internal\nfunc Other() int { return 2 }\n")
        # 증거는 **이 순간**의 소스를 적는다 ...
        for function, source in (("Own", own), ("Other", other)):
            _write_evidence(
                change, package="internal", function=function,
                relative=f"internal/{source.name}",
                digest=hashlib.sha256(source.read_bytes()).hexdigest(),
            )
        # ... 그리고 같은 세션에서 Go 를 한 번 더 고친 뒤 둘을 한 커밋에 담는다.
        own.write_text("package internal\nfunc Own() int { return 3 }\n")
        marks["E"] = _commit_all(root, "E: evidence and one more Go edit in one commit")
        other.write_text("package internal\nfunc Other() int { return 3 }\n")
        marks["X"] = _commit_all(root, "X: a neighbour edits another pinned file")
        return raw, root, marks

    def test_the_recorder_names_what_already_differs_where_the_evidence_entered(self) -> None:
        """사유는 걸은 구간만 말하고, 첫 후보(증거가 들어온 커밋)에서 틀린 소스를 이름으로 댄다.

        마지막 후보 X 의 문장을 인용하면 `internal/other.go` 까지 붙는다 — 그 파일은 증거를
        담은 커밋에서는 맞았으므로 고칠 대상이 아니다."""
        raw, root, marks = self._fixture()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            entered = marks["E"][:12]
            self.assertEqual(lines, [
                f"mine: no landing recorded — no commit at or after the evidence ({entered}) "
                "is accepted as the landing — at the first commit walked, landing point "
                f"{entered} is not the revision this evidence describes: internal/own.go"
            ])
            self.assertFalse((root / "openspec" / "changes" / "mine" / "landed-commit.txt").exists())

    def test_the_gate_names_it_with_and_without_a_record(self) -> None:
        """6.5 는 "대상이 워킹트리라 아무 게이트도 못 본다"로 열렸다. 오늘 두 경로 모두 이름으로
        빨갛다 — 워킹트리 대상은 stale 로, 그 커밋을 선언하면 규칙의 불일치 문장으로."""
        raw, root, marks = self._fixture()
        with raw:
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("AST source hash is stale: internal/own.go" in error for error in errors), errors
            )
            change = root / "openspec" / "changes" / "mine"
            (change / "landed-commit.txt").write_text(marks["E"] + "\n")
            _commit_all(root, "declare the commit that holds the evidence")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(
                    f"landing point {marks['E'][:12]} is not the revision this evidence "
                    "describes: internal/own.go" in error
                    for error in errors
                ),
                errors,
            )


class ALandingMustChangeWhatItsEvidencePins(unittest.TestCase):
    """착지는 고정 번들의 소스 중 **하나 이상**을 base 와 다르게 가진 커밋이어야 한다 (task 7.2.2).

    사람이 2026-09-14 에 고른 규칙이다(리뷰 C1 · C2). 증거가 base 의 소스를 적은 채로 남으면
    그 증거가 맞는 커밋은 전부 그 소스가 아직 base 와 같은 자리이고, 거기서 좁힌 창은 그
    change 의 작업을 **하나도** 담지 못한다. 그런 착지는 두 모양에서 나온다:

    - **V1** — 저장소 규칙대로 FLM 을 먼저 커밋하고 Go 를 고친 뒤 번들을 안 갱신한다.
    - **V2** — Go 작업 뒤, 작업 이전에서 갈라진 곁가지에 base 상태 증거를 두고 병합한다.

    그리고 **정상**인 셋째 모양이 증거 내용으로는 둘과 같다 — 재기준화가 base 를 작업 **뒤로**
    옮긴 change(a074 · a077 · a079 …). 셋을 가를 정보가 저장소에 없어서, 사람이 셋째를 대가로
    치르고 앞의 둘을 막았다. 셋째는 이제 착지를 얻지 못하고 넓은 창으로 돌아간다.
    """

    OLD = "package internal\nfunc Own() int { return 1 }\n"
    NEW = "package internal\nfunc Own() int { return 2 }\n"

    def _base(self, *, read_only_file: bool = False) -> tuple[tempfile.TemporaryDirectory, Path, Path, Path, str]:
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text(self.OLD)
        if read_only_file:
            (root / "internal" / "other.go").write_text("package internal\nfunc Other() int { return 1 }\n")
        base = _commit_all(root, "P: base")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(base + "\n")
        (change / "review.md").write_text("mine\n")
        return raw, root, own, change, base

    def _flm_first(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """V1: P(base) → X(FLM 먼저 — 편집 전 소스를 적은 증거) → W(편집, 번들 안 갱신)."""
        raw, root, own, change, base = self._base()
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        marks = {"P": base, "X": _commit_all(root, "X: FLM first, as the repository asks")}
        own.write_text(self.NEW)
        marks["W"] = _commit_all(root, "W: the edit, bundle not refreshed")
        return raw, root, marks

    def _refusal(self, candidate: str, base: str) -> str:
        return (
            f"landing point {candidate[:12]} changes none of the sources its evidence pins "
            f"since the comparison base {base[:12]}: internal/own.go"
        )

    def test_the_recorder_refuses_evidence_written_before_the_edit(self) -> None:
        """V1 — 리뷰가 손으로 재현한 모양 그대로. 7.2.2 전에는 rc 0 으로 X 를 기록했고 5단계가
        `required 0` 으로 초록이었다(`internal/own.go` 가 바뀌었는데)."""
        raw, root, marks = self._flm_first()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse((root / "openspec" / "changes" / "mine" / "landed-commit.txt").exists())
            self.assertEqual(len(lines), 1, lines)
            self.assertIn(self._refusal(marks["X"], marks["P"]), lines[0])
            # X 는 증거와 **맞는다**. 계산 실패의 머리가 "맞는 커밋이 없다"고 말하면 거짓이다.
            self.assertNotIn("matches every pinning bundle", lines[0])
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("AST source hash is stale: internal/own.go" in error for error in errors), errors
            )

    def test_a_declared_landing_that_changes_none_of_its_pinned_sources_is_refused(self) -> None:
        """선언 경로도 같은 규칙에 묻는다 — 도구가 안 쓰는 값을 손으로 적어도 막힌다."""
        raw, root, marks = self._flm_first()
        with raw:
            (root / "openspec" / "changes" / "mine" / "landed-commit.txt").write_text(marks["X"] + "\n")
            _commit_all(root, "declare the commit before the edit")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(self._refusal(marks["X"], marks["P"]) in error for error in errors), errors
            )

    def test_the_recorder_refuses_side_branch_evidence_merged_after_the_work(self) -> None:
        """V2 — 곁가지 S1 은 작업 W 의 자손이 아니라서 `own.go` 가 base 와 같다. 7.2.2 전에는
        S1 을 기록했고 5단계가 `required 0` 이었다."""
        raw, root, own, change, base = self._base()
        with raw:
            c0 = _commit_all(root, "C0: the change directory, no evidence yet")
            at_base = hashlib.sha256(own.read_bytes()).hexdigest()
            own.write_text(self.NEW)
            _commit_all(root, "W: the Go work lands first")
            home = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
            ).strip()
            subprocess.run(["git", "checkout", "-q", "-b", "side", c0], cwd=root, check=True)
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go", digest=at_base,
            )
            s1 = _commit_all(root, "S1: base-state evidence on a side branch")
            subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
            subprocess.run(["git", "merge", "-q", "--no-edit", "side"], cwd=root, check=True)
            analysis = change / "analysis" / "function-logic"
            self.assertEqual(check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)), s1, "픽스처가 곁가지에 닿아야 한다")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertIn(self._refusal(s1, base), lines[0])

    def test_one_changed_pinned_source_is_enough(self) -> None:
        """규칙은 **하나 이상**이다 — 전부가 아니다. 번들 둘 중 하나는 이 change 가 안 고친 파일
        (근거로 삼은 함수)을 기술해도 착지는 받는다.

        2026-09-12 실측으로 "전부 바뀌어야 한다"(K2-all)는 착지를 얻는 76건 중 25건을 잃는다 —
        사람이 고른 규칙이 아니다. 변이 K-B(`all` → `any`)와 K-G(첫 번들만 본다)가 이 시험 없이
        살아남았다. 정렬 순서로 `internal/other.go` 가 먼저 오게 두어 K-G 가 닿는다."""
        raw, root, own, change, base = self._base(read_only_file=True)
        with raw:
            other = root / "internal" / "other.go"
            own.write_text(self.NEW)
            for function, source in (("Own", own), ("Other", other)):
                _write_evidence(
                    change, package="internal", function=function, relative=f"internal/{source.name}",
                    digest=hashlib.sha256(source.read_bytes()).hexdigest(),
                )
            landing = _commit_all(root, "W: own.go changes, other.go does not")
            # 닿는지 먼저 — `other.go` 가 base 에 **있고** 착지에서 같아야 이 시험이 "안 바뀐 고정
            # 소스"를 잰다. 첫 판은 그 파일을 base 뒤에 만들어서 base 에 없었고(= 바뀐 것으로 셈),
            # K-B · K-G 가 이 시험을 두고도 살아남았다([[mutation-must-reach-the-thing-under-test]]).
            self.assertIsNotNone(check_analysis._committed_bytes(root, base, "internal/other.go"))
            self.assertEqual(
                check_analysis._committed_bytes(root, base, "internal/other.go"),
                check_analysis._committed_bytes(root, landing, "internal/other.go"),
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            self.assertEqual(
                (change / "landed-commit.txt").read_text().strip(), landing
            )

    def test_a_change_whose_work_precedes_its_base_gets_no_landing(self) -> None:
        """사람이 치른 대가를 못 박는다. 재기준화가 base 를 작업 뒤로 옮긴 change 는 증거가 base 를
        정확히 적는다 — V1 과 같은 모양이라 착지를 얻지 못한다. 규칙이 누그러지면 여기가 빨개진다."""
        raw, root, own, change, base = self._base()
        with raw:
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            (root / "docs.md").write_text("post-deployment measurement\n")
            landing = _commit_all(root, "L: docs only — the work landed before the base")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertIn(self._refusal(landing, base), lines[0])
            (change / "landed-commit.txt").write_text(landing + "\n")
            _commit_all(root, "declare it by hand")
            errors = check_analysis.check("mine", root)
            self.assertTrue(any(self._refusal(landing, base) in error for error in errors), errors)


class EvidenceFirstCommittedInsideAMergeIsAKnownLimit(unittest.TestCase):
    """병합 커밋 **안에서** 처음 커밋한 증거는 하한을 못 세운다 (task 7.2.4, 리뷰 H7).

    하한은 `git log --diff-filter=MAT -- <번들>` 로 찾는데, `-m` 없는 `git log` 는 병합 커밋 자신의
    변경을 읽지 않는다. 그래서 병합을 마치며 번들을 처음 추가하면 그 change 는 착지를 얻지 못한다.
    막는 쪽으로 틀리는 문제라 사람이 2026-09-14 에 **한계로 두기로** 골랐다 — 기록이 없으면 5단계는
    워킹트리로 판정하므로 그 change 가 막히지는 않고, 창을 좁히지 못할 뿐이다.

    고친 것은 **문장**이다. 옛 문장 "the pinning evidence never entered this history (commit the bundles)"
    는 이 모양에서 거짓이었다 — 번들은 HEAD 에 커밋돼 있다(아래 도달 단언). 문장이 한계를 사실대로
    말해야 조언을 받은 사람이 이미 한 커밋을 또 하지 않는다([[a-not-found-reason-claims-only-what-was-searched]])."""

    WALK = "no ordinary commit on this history adds the pinning evidence — commit the bundles in an ordinary commit (a merge commit's own changes are not read)"

    def _fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        (root / "notes.md").write_text("a\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        _commit_all(root, "the change directory")
        home = subprocess.check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True).strip()
        subprocess.run(["git", "checkout", "-q", "-b", "side"], cwd=root, check=True)
        (root / "notes.md").write_text("side\n")
        _commit_all(root, "side: unrelated")
        subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: the work")
        subprocess.run(["git", "merge", "-q", "--no-ff", "--no-commit", "side"], cwd=root, check=True,
                       capture_output=True)
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
        marks["M"] = _commit_all(root, "M: the merge, and the evidence first committed inside it")
        return raw, root, marks

    def test_the_recorder_names_the_limit_instead_of_asking_for_a_commit_already_made(self) -> None:
        raw, root, marks = self._fixture()
        with raw:
            ast_json = "openspec/changes/mine/analysis/function-logic/internal--own/ast.json"
            # 닿는지 먼저 — 번들이 **정말** 커밋돼 있고 병합 커밋 M 이 그것을 처음 들였다.
            self.assertIsNotNone(check_analysis._committed_bytes(root, "HEAD", ast_json))
            self.assertIsNone(check_analysis._committed_bytes(root, f"{marks['M']}^1", ast_json))
            self.assertEqual(
                len(subprocess.check_output(["git", "rev-list", "--parents", "-1", marks["M"]], cwd=root, text=True).split()),
                3, "M 은 병합 커밋이어야 한다",
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertEqual(lines, [f"mine: no landing recorded — {self.WALK}"])

    def test_without_a_record_step_five_still_judges_the_working_tree(self) -> None:
        """한계의 크기: 막히는 것은 좁히기뿐이다. 증거가 맞으면 5단계는 통과한다."""
        raw, root, _ = self._fixture()
        with raw:
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), "")
            self.assertEqual(facts.get("required_count"), 1)

    def test_a_hand_written_record_at_the_merge_names_the_limit(self) -> None:
        raw, root, marks = self._fixture()
        with raw:
            _declare_landing(root / "openspec" / "changes" / "mine", marks["M"], "declare the merge by hand")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(
                    f"landing point {marks['M'][:12]} is pinned by evidence that no ordinary commit on this "
                    "history adds" in error and "non-merge" in error
                    for error in errors
                ),
                errors,
            )


def _own_work_fixture(raw: tempfile.TemporaryDirectory) -> tuple[Path, dict[str, str]]:
    """P(base) → W(이 change 의 Go 작업) → E(그 증거가 역사에 들어옴). 기록은 아직 없다.

    `TheGateRecordsTheLandingInsteadOfTheAuthor._fixture(base_also_matches=False)`
    와 같은 모양이고, 7.4 의 두 클래스가 이 **하나**를 나눠 쓴다 — 리뷰가 픽스처
    사본 다섯 벌을 이미 부채로 셌으므로(I15) 여섯째를 만들지 않는다.
    """
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
    marks["E"] = _commit_all(root, "E: its evidence enters the history")
    return root, marks


def _cli(root: Path, *extra: str, change: str = "mine") -> tuple[int, str]:
    """CLI 를 생산 경로 그대로 돌린다 — 게이트가 실제로 부르는 것이 이것뿐이다."""
    output = io.StringIO()
    argv = ["check_analysis.py", "--change", change, "--root", str(root), *extra]
    with mock.patch.object(sys, "argv", argv), redirect_stdout(output):
        code = check_analysis.main()
    return code, output.getvalue()


def _escaping_bundle(root: Path) -> None:
    """번들의 `file` 이 저장소 **밖**을 가리키게 만든다. 커밋까지 한다."""
    ast_path = (
        root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
        / "internal--own" / "ast.json"
    )
    value = json.loads(ast_path.read_text(encoding="utf-8"))
    value["file"] = "../outside/own.go"
    ast_path.write_text(json.dumps(value), encoding="utf-8")
    _commit_all(root, "the bundle names a source outside the repository")


class ARecordIsWrittenOnceAndNeverThroughASymlink(unittest.TestCase):
    """`--record-landing` 이 **자기 자리에 한 번만** 쓴다 (task 7.4, 리뷰 H4).

    오늘은 존재 확인과 쓰기 사이에 walk 하나가 통째로 들어간다 — 리뷰 I1 이 실측한
    시간이 133.7s·219.6s 다. 그 사이에 생긴 기록을 `write_text` 가 조용히 덮고,
    끊긴 심링크는 따라가서 저장소 **밖**에 쓴다. 같은 자리를 `execution_baseline.py`
    는 `open(…, "xb")` 로 이미 닫아 놨다 — 이 파일만 안 닫혀 있었다.

    양성 대조군은 따로 만들지 않는다. 정상 입력이 기록되는 것은
    `TheGateRecordsTheLandingInsteadOfTheAuthor` 의 왕복 시험이 이미 재고 있고,
    그 시험이 여기 고친 쓰기 경로를 그대로 지난다.
    """

    def test_a_dangling_symlink_record_is_refused_not_followed(self) -> None:
        """끊긴 심링크는 `exists()` 가 거짓이라 존재 확인을 그냥 통과한다."""
        raw = tempfile.TemporaryDirectory()
        elsewhere = tempfile.TemporaryDirectory()
        with raw, elsewhere:
            root, _ = _own_work_fixture(raw)
            stolen = Path(elsewhere.name) / "stolen.txt"
            (root / "openspec" / "changes" / "mine" / "landed-commit.txt").symlink_to(stolen)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse(stolen.exists(), "기록이 저장소 밖에 쓰였다")
            # 거절 **지점**을 단언한다 (task 6.3 이 같은 파일에서 배운 것). 배타 생성이
            # 아래에서 같은 값을 돌려주므로, 지점을 안 재면 이 거절을 지워도 초록이다
            # ([[surviving-mutant-may-mean-accidental-safety]] — 2026-09-13 M-A 로 실측).
            self.assertTrue(any("is a symlink" in line for line in lines), lines)

    def test_a_record_that_appears_while_computing_is_not_overwritten(self) -> None:
        """경합을 **결정적으로** 재는 자리 — 계산이 도는 동안 기록이 생기게 만든다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            record = root / "openspec" / "changes" / "mine" / "landed-commit.txt"

            def someone_records_while_we_walk(*args: object, **kwargs: object):
                record.write_text(marks["W"] + "\n", encoding="utf-8")
                return marks["E"], ""

            with mock.patch.object(
                check_analysis, "compute_landing", side_effect=someone_records_while_we_walk
            ):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertEqual(record.read_text(encoding="utf-8").strip(), marks["W"])


class AFaultInGitBecomesAVerdictNotATraceback(unittest.TestCase):
    """판정 도구는 판정을 못 내겠으면 **이름으로** 말한다 (task 7.4, 리뷰 H5·H6).

    두 자리가 열려 있었다. (H5) `compute_landing` 이 저장소 밖을 가리키는 번들 앞에서
    `ValueError` 를 내면 `record_landing` 이 스택으로 죽는다. (H6) `TimeoutExpired` 는
    `OSError` 가 **아니라서** `check` 의 예외 목록 넷을 그대로 빠져나간다 — git 이 한 번
    멎으면 게이트를 돌린 사람은 무엇이 왜 막혔는지도, 어느 창으로 쟀는지도 못 읽는다.
    """

    def test_record_landing_names_a_bundle_that_escapes_the_repository(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            _escaping_bundle(root)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any("escapes repository" in line for line in lines), lines)

    def test_the_cli_names_the_same_fault_and_still_prints_the_window(self) -> None:
        """창은 **잰 것**이므로 계속 찍는다. 이 결함은 착지를 잰 **뒤**에 터진다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            _escaping_bundle(root)
            code, output = _cli(root)
            self.assertEqual(code, 1, output)
            self.assertIn("escapes repository", output)
            self.assertIn("working tree (no landed-commit.txt in HEAD)", output)

    def test_the_cli_names_a_fault_the_record_path_cannot_answer(self) -> None:
        """기록 경로가 스스로 못 답하는 결함도 CLI 에서는 이름이 된다.

        `record_landing` 안에서 git 을 부르는 자리는 `compute_landing` 말고도 넷 더
        있다(`resolve_referenced_change` · `_landing_record` · `resolve_base` ·
        워킹트리 dirty 확인). 자리마다 목록을 베끼는 대신 **경계 하나**에 세웠으므로,
        그 경계가 실제로 서 있는지를 그중 하나로 잰다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            real_run = subprocess.run

            def the_dirty_check_hangs(args: object, **kwargs: object):
                if isinstance(args, list) and "--quiet" in args:
                    raise subprocess.TimeoutExpired(cmd=args, timeout=30)
                return real_run(args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", the_dirty_check_hangs):
                code, output = _cli(root, "--record-landing")
            self.assertEqual(code, 1, output)
            self.assertIn("timed out", output)
            self.assertFalse(
                (root / "openspec" / "changes" / "mine" / "landed-commit.txt").exists(),
                "결함 앞에서 기록이 만들어졌다",
            )

    def test_a_timed_out_git_call_is_named_instead_of_raised(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            _declare_landing(
                root / "openspec" / "changes" / "mine", marks["E"], "declare the landing"
            )
            real_run = subprocess.run

            def the_floor_call_hangs(args: object, **kwargs: object):
                if isinstance(args, list) and "-M100%" in args:
                    raise subprocess.TimeoutExpired(cmd=args, timeout=60)
                return real_run(args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", the_floor_call_hangs):
                code, output = _cli(root)
            self.assertEqual(code, 1, output)
            self.assertIn("timed out", output)


class TheRecordMustBeTheValueTheGateComputes(unittest.TestCase):
    """기록은 도구가 계산한 **그 값**이어야 한다 (task 7.3, 리뷰 H3).

    spec 은 이미 "저자가 선언하는 값이 아니라 도구가 계산해서 기록하는 값"을 SHALL 로
    적는데, 게이트는 기록이 **유효한지**만 보고 그 값인지는 안 봤다. 2026-09-13 전수
    실측: 착지를 얻는 76건 중 **65건**이 유효 조건을 통과하는 값을 둘 이상 갖고(최대
    516개), **44건**에서는 그 선택이 판정 입력을 바꾼다. 이탈이 오늘 이 역사에서
    창을 넓히는 방향뿐이라 해도(실측), 값을 만드는 주체를 도구로 옮긴 규칙은 게이트가
    그 값을 확인할 때만 규칙이다. 사람이 2026-09-13 에 그렇게 골랐다.
    """

    def _fixture(self, raw: tempfile.TemporaryDirectory) -> tuple[Path, dict[str, str]]:
        """P → W → E → L. `L` 은 **고정 소스를 안 건드리는** 나중 커밋이다.

        그래서 `L` 은 오늘의 가드를 전부 통과한다 — 고정 해시가 맞고, 하한(E) 뒤이고,
        판정이 읽은 번들을 들고 있다. 계산값은 `E` 이므로 저자에게 남은 선택이 정확히
        이 모양이다."""
        root, marks = _own_work_fixture(raw)
        (root / "internal" / "neighbour.go").write_text(
            "package internal\nfunc Neighbour() int { return 1 }\n"
        )
        marks["L"] = _commit_all(root, "L: a later commit that leaves the pinned source alone")
        return root, marks

    def test_a_later_commit_that_also_matches_is_refused(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._fixture(raw)
            _declare_landing(
                root / "openspec" / "changes" / "mine", marks["L"], "declare the later value"
            )
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("is not the landing this change's evidence computes" in e for e in errors),
                errors,
            )
            # 거절은 **두 값**을 다 말해야 한다. 하나만 말하면 저자는 무엇을 적어야
            # 하는지 모른 채 게이트만 빨간 것을 본다.
            self.assertTrue(any(marks["L"][:12] in e for e in errors), errors)
            self.assertTrue(any(marks["E"][:12] in e for e in errors), errors)
            # 그리고 **돌아가는 길**도 (task 7.2.6, spec 의 SHALL). 이 문장만 그것을
            # 빠뜨려서, `--record-landing` 을 권해 놓고 그 명령은 "이미 있다"로 거절했다
            # (2026-09-16 적대 리뷰 F3).
            self.assertTrue(
                any(check_analysis.LANDING_RECOVERY in e for e in errors), errors
            )

    def test_the_computed_value_is_still_accepted(self) -> None:
        """양성 대조군. 새 등식이 **모든** 기록을 거절하면 시험은 그래도 초록이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._fixture(raw)
            _declare_landing(
                root / "openspec" / "changes" / "mine", marks["E"], "declare the computed value"
            )
            self.assertEqual(check_analysis.check("mine", root), [])

    def test_the_landing_is_the_first_matching_commit_not_where_evidence_entered(self) -> None:
        """계산값은 **하한이 아니다**.

        둘이 갈리는 모양은 흔한 커밋 순서다 — 증거를 먼저 올리고 코드를 그 뒤에 올리면
        하한(ast.json 이 역사에 들어온 커밋)에서는 아직 소스가 안 맞고, 계산값은 소스가
        처음 맞는 커밋이다. 이 픽스처가 없으면 "하한과 비교" 변이가 스위트를 통과한다
        (2026-09-13 N-C 로 실측 — SURVIVED 였다)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = _init_fixture(raw)
            own = root / "internal" / "own.go"
            own.parent.mkdir(parents=True)
            own.write_text("package internal\nfunc Own() int { return 1 }\n")
            marks = {"P": _commit_all(root, "P: base")}
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(marks["P"] + "\n")
            (change / "review.md").write_text("mine\n")
            after = b"package internal\nfunc Own() int { return 2 }\n"
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(after).hexdigest(),
            )
            marks["E"] = _commit_all(root, "E: the evidence lands first, the source is still v1")
            own.write_bytes(after)
            marks["W2"] = _commit_all(root, "W2: the source the evidence describes lands")
            analysis = change / "analysis" / "function-logic"
            # 픽스처가 실제로 둘을 갈랐는지 **먼저** 단언한다. 안 갈렸으면 아래가 재는 것이 없다.
            self.assertEqual(check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)), marks["E"])
            self.assertEqual(
                check_analysis.compute_landing(root, marks["P"], _inputs(root, analysis))[0], marks["W2"]
            )
            _declare_landing(change, marks["W2"], "declare the computed landing")
            self.assertEqual(check_analysis.check("mine", root), [])

    def test_what_the_gate_records_is_what_the_gate_then_demands(self) -> None:
        """왕복. 도구가 쓴 값이 이 등식을 통과하지 않으면 기록 명령이 못 쓰게 된다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._fixture(raw)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "commit the recorded landing")
            self.assertEqual(check_analysis.check("mine", root), [])


class AComputedLandingIsAlwaysOneTheGateWillAccept(unittest.TestCase):
    """계산 경로가 내는 값은 선언 경로가 받는 값이어야 한다 (리뷰 I9, 변이 M6).

    `resolve_landing` 은 "착지가 base 의 자손인가"를 이미 묻는다. `compute_landing` 은
    같은 질문을 자기 순회 안에서 한 번 더 하는데, 그 절을 지워도 스위트가 초록이었다 —
    선형 픽스처만 있어서 `base..HEAD` 의 모든 커밋이 자동으로 base 의 자손이었기
    때문이다. `git rev-list base..HEAD` 는 **base 에서 안 보이는 커밋 전부**이고,
    병합이 있으면 거기에 base 를 한 번도 보지 못한 곁가지가 들어온다.

    두 판정이 갈리면 도구는 자기 게이트가 거절할 값을 쓴다
    ([[two-judgements-cover-for-each-other]]). 7.3.1 이 기록을 계산값과 **같게**
    묶은 뒤로 그 값은 고칠 수도 없다 — 그 change 는 영원히 통과하지 못한다.
    """

    def _merged_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """R(갈림) → 줄기 P(base) → 곁가지 S(작업+증거) → 병합 M.

        S 는 base 를 한 번도 보지 못했다. 증거는 S 에서 들어오므로 하한은 S 이고,
        `base ≤ floor` 가 거짓이라 순회는 base 에서 시작한다 — 그래서 S 가 후보
        목록에 **들어온다**. 정직한 답은 이 change 의 선 위에서 가장 낮은 M 이다."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        other = root / "internal" / "other.go"
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        (change / "base-commit.txt").write_text("pending\n")
        marks = {"R": _commit_all(root, "R: the fork point")}
        trunk = subprocess.check_output(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
        ).strip()
        marks["P"] = _commit_all_allowing_nothing(root, "P: base, on the trunk only")
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        _commit_all(root, "declare the base")
        subprocess.run(["git", "checkout", "-q", "-b", "side", marks["R"]], cwd=root, check=True)
        other.write_text("package internal\nfunc Other() int { return 2 }\n")
        _write_evidence(
            change, package="internal", function="Other", relative="internal/other.go",
            digest=hashlib.sha256(other.read_bytes()).hexdigest(),
        )
        marks["S"] = _commit_all(root, "S: work and evidence on a branch that never saw the base")
        subprocess.run(["git", "checkout", "-q", trunk], cwd=root, check=True)
        subprocess.run(["git", "merge", "-q", "--no-edit", "side"], cwd=root, check=True)
        marks["M"] = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=root, text=True
        ).strip()
        return raw, root, marks

    def test_the_fixture_reaches_the_walk_with_a_side_branch_in_it(self) -> None:
        """계측기 대조군. 이 셋이 성립하지 않으면 아래 두 시험은 아무것도 안 잰다
        ([[mutation-must-reach-the-thing-under-test]])."""
        raw, root, marks = self._merged_fixture()
        with raw:
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            self.assertEqual(check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)), marks["S"])
            self.assertFalse(check_analysis._is_ancestor(root, marks["P"], marks["S"]))
            walked = subprocess.check_output(
                ["git", "rev-list", f"{marks['P']}..HEAD"], cwd=root, text=True
            ).split()
            self.assertIn(marks["S"], walked, "곁가지가 순회 목록에 들어와야 한다")

    def test_a_commit_the_base_never_reached_is_not_the_landing(self) -> None:
        raw, root, marks = self._merged_fixture()
        with raw:
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            landing, why = check_analysis.compute_landing(root, marks["P"], _inputs(root, analysis))
            self.assertEqual(landing, marks["M"], why)
            self.assertNotEqual(landing, marks["S"], "곁가지는 이 change 의 선 위에 없다")

    def _parallel_evidence_fixture(self) -> tuple[tempfile.TemporaryDirectory, Path, dict[str, str]]:
        """P(base) → 곁가지 B(작업+증거) · 줄기 F(**같은 바이트**의 증거를 먼저, 더 늦게) → 병합 M.

        하한은 `-1` 이 고르는 늦은 쪽 F 다. F 에는 작업이 아직 없어 불일치로 떨어지고,
        그다음 후보 B 는 소스도 맞고 번들 바이트도 같은데 **F 의 자손이 아니다**. 이 후보를
        막는 것은 하한 순서 가드 **혼자**다 — 다른 픽스처는 전부 하한 앞의 후보가 불일치로
        먼저 떨어져서, 계산 경로가 규칙에 하한을 제대로 넘기는지를 아무도 안 쟀다
        (7.6 뮤테이션 R16 SURVIVED)."""
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "review.md").write_text("mine\n")
        (change / "base-commit.txt").write_text("pending\n")
        marks = {"P": _commit_at(root, "P: base", "2026-09-01T00:00:00")}
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        _commit_at(root, "declare the base", "2026-09-01T01:00:00")
        trunk = subprocess.check_output(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
        ).strip()
        worked = "package internal\nfunc Own() int { return 2 }\n"
        digest = hashlib.sha256(worked.encode()).hexdigest()
        subprocess.run(["git", "checkout", "-q", "-b", "side"], cwd=root, check=True)
        own.write_text(worked)
        _write_evidence(change, package="internal", function="Own", relative="internal/own.go", digest=digest)
        marks["B"] = _commit_at(root, "B: work and evidence on a branch", "2026-09-03T00:00:00")
        subprocess.run(["git", "checkout", "-q", trunk], cwd=root, check=True)
        _write_evidence(change, package="internal", function="Own", relative="internal/own.go", digest=digest)
        marks["F"] = _commit_at(root, "F: the same evidence first on the trunk", "2026-09-04T00:00:00")
        env = {**os.environ, "GIT_AUTHOR_DATE": "2026-09-05T00:00:00",
               "GIT_COMMITTER_DATE": "2026-09-05T00:00:00"}
        subprocess.run(["git", "merge", "-q", "--no-edit", "side"], cwd=root, check=True, env=env)
        marks["M"] = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
        return raw, root, marks

    def test_the_parallel_fixture_reaches_a_candidate_only_the_floor_refuses(self) -> None:
        """계측기 대조군 ([[mutation-must-reach-the-thing-under-test]])."""
        raw, root, marks = self._parallel_evidence_fixture()
        with raw:
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            self.assertEqual(check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)), marks["F"])
            self.assertFalse(check_analysis._is_ancestor(root, marks["F"], marks["B"]))
            self.assertIn("is not the revision this evidence describes",
                          check_analysis._landing_refusal(root, marks["P"], marks["F"], _inputs(root, analysis, floor=marks["F"], repairs=[]))[0])
            # B 를 막는 것이 하한 순서 **하나**뿐임을 직접 보인다: 하한을 base 로 주면 받는다.
            self.assertEqual(check_analysis._landing_refusal(root, marks["P"], marks["B"], _inputs(root, analysis, floor=marks["P"], repairs=[])), ("", []))
            self.assertIn("precedes the evidence that pins it",
                          check_analysis._landing_refusal(root, marks["P"], marks["B"], _inputs(root, analysis, floor=marks["F"], repairs=[]))[0])

    def test_a_candidate_older_than_its_evidence_is_not_the_landing(self) -> None:
        raw, root, marks = self._parallel_evidence_fixture()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "commit the recorded landing")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts["landing"], marks["M"], "곁가지 B 는 증거보다 앞선다")

    def test_what_it_records_on_a_merged_history_the_gate_still_accepts(self) -> None:
        """종단. 값 비교만 하면 "다른 값이 왜 나쁜가"가 시험에 안 남는다."""
        raw, root, marks = self._merged_fixture()
        with raw:
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "commit the recorded landing")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts["landing"], marks["M"])
            self.assertEqual(facts["required_count"], 1)


class TheLandingRuleLivesInOnePlace(unittest.TestCase):
    """착지 수락 규칙과 아카이브 이름 규칙은 각자 **한 함수**에 산다 (task 7.6, 리뷰 I2 · I6).

    두 벌이면 한쪽만 고쳐도 양쪽 시험이 초록이다([[two-judgements-cover-for-each-other]]).
    그리고 행동 시험은 **일치하는 두 사본**을 못 가른다 — 둘이 오늘 같은 답을 내면 초록이기
    때문이다. 그래서 이 클래스만 소스의 구조를 본다: 누가 누구에게 묻는가.
    """

    @staticmethod
    def _calls(function: str) -> set[str]:
        tree = ast.parse(Path(check_analysis.__file__).read_text(encoding="utf-8"))
        node = next(
            item for item in ast.walk(tree)
            if isinstance(item, ast.FunctionDef) and item.name == function
        )
        return {ast.unparse(call.func) for call in ast.walk(node) if isinstance(call, ast.Call)}

    def test_the_declared_and_the_computed_landing_ask_one_rule(self) -> None:
        for function in ("resolve_landing", "compute_landing"):
            calls = self._calls(function)
            self.assertIn("_landing_refusal", calls, function)
            # 규칙의 조각을 직접 부르는 자리가 곧 두 번째 사본이다.
            self.assertFalse(calls & {"_pinning_at", "_unheld_bundles"}, (function, sorted(calls)))

    def test_the_recorder_and_its_advice_ask_one_judge(self) -> None:
        """`--record-landing` 이 걷기 전에 멈추는 사유는 **한 함수**에 산다 (task 7.7, 리뷰 I8).

        기록 명령과 조언 줄이 각자 답하던 동안 조언은 명령의 거절 일곱 중 어느 것도 몰랐다.
        행동 시험은 **오늘 모양들**에서 둘이 같은지만 본다 — 명령에 여덟째 거절이 생기면
        조언 줄의 사본이 없는 한 여기서만 드러난다."""
        for function in ("record_landing", "main"):
            self.assertIn("_recording_refusal", self._calls(function), function)
        # 거절의 조각을 명령이 직접 부르면 그것이 두 번째 사본이다.
        self.assertFalse(
            self._calls("record_landing") & {"_landing_record", "resolve_base", "subprocess.run"},
            sorted(self._calls("record_landing")),
        )
        # 걷기 전의 하한은 계산과 판정이 **같은** 함수에 묻는다 — `_walk_floor` 한 곳. 둘 다 증거를
        # 스스로 읽지 않고 명령이 한 번 읽은 것을 받는다 (task 7.5.2.1).
        for function in ("_recording_refusal", "_measure_landing_inputs"):
            calls = self._calls(function)
            self.assertIn("_walk_floor", calls, function)
            self.assertFalse(calls & {"_read_evidence", "_evidence_floor"}, (function, sorted(calls)))
        # 계산은 받은 **한 벌**만 쓴다 (task 7.5.2.1) — 7.5.1 · 7.5.2 에서는 입력이 없으면 스스로 쟀다.
        calls = self._calls("compute_landing")
        self.assertFalse(calls & {"_measure_landing_inputs", "_walk_floor", "_read_evidence",
                                  "_select_pinning", "_evidence_floor", "_self_repair_commits"},
                         sorted(calls))

    def test_a_commits_blob_is_read_by_one_function(self) -> None:
        """"그 커밋의 blob 을 읽는다"는 철자는 **한 곳**에 산다 (task 7.5).

        행동으로는 못 가른다 — `_committed_bytes` 를 자기 `git show` 로 되돌려도 답이 같아서
        스위트가 초록이다(변이 T16 SURVIVED). 갈라지는 것은 **나중**이다: 한쪽에만 붙인
        `-Z` 프레이밍이나 blob 타입 검사가 다른 쪽에 없으면, 같은 질문에 두 답이 생긴다.
        그래서 의존을 구조로 못 박는다 ([[surviving-mutant-may-mean-accidental-safety]])."""
        readers = {"_committed_many"}
        for function in ("_committed_bytes", "_pinning_at", "_unheld_bundles"):
            calls = self._calls(function)
            self.assertTrue(readers & calls, (function, sorted(calls)))
            # 자기 프로세스를 띄우면 그것이 두 번째 사본이다.
            self.assertNotIn("subprocess.run", calls, function)
        # 배치로 읽는 자리가 자리마다 `_committed_bytes` 로 되돌아가지 않았는가.
        for function in ("_pinning_at", "_unheld_bundles"):
            self.assertNotIn("_committed_bytes", self._calls(function), function)

    def test_the_walk_measures_the_bundle_list_once(self) -> None:
        """후보 순회는 번들 목록을 **한 번** 재서 넘긴다 (task 7.5) — `floor` · `repairs` 와 같다.

        후보마다 다시 재도 답은 같으므로 행동 시험이 못 가른다. 값은 성능이 아니라 **일관성**
        이기도 하다: 두 번 재면 그 사이의 디스크 변화가 하한과 규칙을 갈라 놓는다."""
        # 걷는 쪽은 증거를 **읽지 않는다** — 명령이 한 번 읽은 것을 받는다 (task 7.5.2.1). 고르는 것은
        # 잴 때 한 번이다(`_measure_landing_inputs`).
        walking = ("_pinning_at", "_unheld_bundles", "_landing_refusal", "compute_landing", "resolve_landing")
        for function in walking + ("_measure_landing_inputs", "_walk_floor"):
            self.assertNotIn("_read_evidence", self._calls(function), function)
        for function in walking:
            self.assertNotIn("_select_pinning", self._calls(function), function)
        # 선언 경로도 기록 경로도 입력을 **같은 함수**에서 한 번 잰다 (task 7.5.1, F2).
        for function in ("resolve_landing", "record_landing"):
            calls = self._calls(function)
            self.assertIn("_measure_landing_inputs", calls, function)
            self.assertFalse(calls & {"_evidence_floor", "_self_repair_commits"}, (function, sorted(calls)))
        # 명령은 증거를 **한 자리**에서 읽는다 — 호출 자리를 센다(행동 시험은 한 번 실행만 본다).
        tree = ast.parse(Path(check_analysis.__file__).read_text(encoding="utf-8"))
        for function in ("check", "record_landing"):
            node = next(item for item in ast.walk(tree)
                        if isinstance(item, ast.FunctionDef) and item.name == function)
            sites = [call for call in ast.walk(node)
                     if isinstance(call, ast.Call) and ast.unparse(call.func) == "_read_evidence"]
            self.assertEqual(len(sites), 1, function)

    def test_archive_names_are_read_by_one_function(self) -> None:
        for function in ("resolve_referenced_change", "_pre_archive_path"):
            calls = self._calls(function)
            self.assertIn("_archived_change_id", calls, function)
            self.assertNotIn("ARCHIVED_CHANGE.fullmatch", calls, function)

    def test_an_archived_name_yields_its_id_and_nothing_else_does(self) -> None:
        for name, expected in (
            ("2026-09-11-mine", "mine"),
            ("2026-08-29-other-reference", "other-reference"),
            ("mine", ""),
            ("abcd-ef-gh-mine", ""),
            ("2026-09-11-", ""),
        ):
            self.assertEqual(check_analysis._archived_change_id(name), expected, name)


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

    def test_borrowed_evidence_no_longer_pins_a_landing(self) -> None:
        """3.2.3.1 은 빌린 번들로 착지를 고정하게 했다 — a073 의 유일한 수리 경로였다. 7.2.3 이 그 경로를
        닫았다(빌리는 change 는 좁히지 않는다). 빌린 번들이 그 착지를 **정확히** 기술해도 기록은 거절된다.

        양성 대조: 기록 전에는 통과한다 — 거절이 픽스처의 다른 결함이 아니라 기록 때문인지 가른다."""
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
            _declare_landing(change, marks["L"], "record the lender's landing on the borrower too")
            self.assertEqual(
                check_analysis.check("ordinary", root), [check_analysis.BORROWED_REFUSES_A_LANDING]
            )


class ABorrowedWindowIsNeverNarrowed(unittest.TestCase):
    """증거를 빌리는 change 는 비교 창을 **좁히지 않는다** (task 7.2.3, 리뷰 C4).

    1.8 은 창의 양쪽 끝을 빌려주는 change 와 공유하게 했다 — 빌리는 쪽은 빌려주는 쪽의 착지를
    **복사**한다. 그런데 빌려주는 쪽의 착지는 빌려주는 쪽 증거의 가장 낮은 값(6.1.2 · 7.3.1)이라,
    빌리는 쪽이 그 **뒤에** 한 Go 작업은 복사한 창 밖으로 나가고 어떤 판정도 못 본다. 빌리는 쪽은
    정의상 자기 번들이 0 이라 자기 작업이 어디 착지했는지 말할 증거를 **소유하지 않는다**.
    사람이 2026-09-14 에 "빌리는 change 는 좁히지 못한다"를 골랐다. base 공유는 그대로다.

    대가는 저장소의 빌리는 change 하나(a073, 아카이브 · 기록 없음)가 1.8 이 열어 둔 수리 경로
    (양쪽이 같은 착지를 기록해 336 오류 → 0)를 잃는 것이다. 그 경로는 쓰인 적이 없다."""

    def _borrower(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "ordinary"

    def _lender(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "reference"

    def test_a_lender_record_does_not_narrow_the_borrower(self) -> None:
        """C4 그 자체. 빌려주는 쪽이 L 을 기록하고 빌리는 쪽 작업이 L2 에 있으면, 빌리는 쪽의 창은
        워킹트리까지이고 L2 의 작업이 **요구된다**.

        먼저 양성 대조로 L 에서 좁힌 창이 그 작업을 **뺀다**는 것을 잰다 — 안 빼면 아래 단언이
        아무것도 재지 못한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw, later_work=True)
            borrower, lender = self._borrower(root), self._lender(root)
            base = (borrower / "base-commit.txt").read_text(encoding="utf-8").strip()
            self.assertNotIn(
                ("internal/other.go", "Other"),
                check_analysis.changed_existing_functions(root, base, marks["L"]),
                "양성 대조: L 에서 좁힌 창은 L2 의 작업을 빼야 한다",
            )
            (lender / "landed-commit.txt").write_text(marks["L"] + "\n")
            _commit_all(root, "the lender records its own computed landing")
            facts: dict[str, object] = {}
            errors = check_analysis.check("ordinary", root, facts)
            self.assertEqual(facts.get("landing"), "", f"빌리는 쪽이 빌려주는 쪽의 착지로 좁혀졌다: {errors}")
            self.assertTrue(
                any("internal/other.go:Other" in error for error in errors),
                f"빌리는 쪽의 뒤 작업이 요구에서 빠졌다: {errors}",
            )
            self.assertFalse(any("landing point" in error for error in errors), errors)

    def test_a_borrower_that_copies_the_lender_is_refused(self) -> None:
        """1.8 이 **받던** 모양 — 양쪽이 같은 값을 기록한다. 이제 빌리는 쪽의 기록은 이름으로 거절한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw, later_work=True)
            (self._lender(root) / "landed-commit.txt").write_text(marks["L"] + "\n")
            _declare_landing(self._borrower(root), marks["L"], "the borrower copies the lender")
            self.assertEqual(
                check_analysis.check("ordinary", root), [check_analysis.BORROWED_REFUSES_A_LANDING]
            )
            self.assertIn("a borrowed window is never narrowed", check_analysis.BORROWED_REFUSES_A_LANDING)

    def test_a_borrower_record_alone_is_refused(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _borrowed_fixture(raw)
            _declare_landing(self._borrower(root), marks["L"], "the borrower records alone")
            self.assertEqual(
                check_analysis.check("ordinary", root), [check_analysis.BORROWED_REFUSES_A_LANDING]
            )

    def test_an_undecodable_borrower_record_is_refused_not_raised(self) -> None:
        """"기록이 있는가"만 묻는다 — 해독하는 함수를 부르면 못 읽는 기록 앞에서 질문이 터진다
        (6.2.1 이 이관 경로에서 실측한 모양). 못 읽는 기록도 기록이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _borrowed_fixture(raw)
            (self._borrower(root) / "landed-commit.txt").write_bytes(b"\xff\xfe\n")
            _commit_all(root, "an undecodable borrower record")
            self.assertEqual(
                check_analysis.check("ordinary", root), [check_analysis.BORROWED_REFUSES_A_LANDING]
            )

    def test_neither_side_declaring_stays_accepted(self) -> None:
        """거부하면 안 되는 정상 입력: 양쪽 다 기록이 없는 모양. a073 · a072 의 오늘 모양이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _borrowed_fixture(raw)
            self.assertEqual(
                check_analysis.check("ordinary", root), [],
                "기록이 없는 빌림에 새 오류를 더했다",
            )


class AFailingStepFiveSaysWhichWindowRequiredThem(unittest.TestCase):
    """실패 출력이 "왜 이 함수가 요구되는가"를 말해야 한다 (task 3.3).

    2026-09-10 실측: a074 는 324줄 중 316줄이 `missing evidence for modified function`
    이고 **비교 창을 말하는 줄이 0** 이다. a076 은 이름 316개를 쉼표로 이은
    **21,838자짜리 한 줄**이고 역시 0 이다. 그 316개는 전부 남의 change 함수인데
    출력만 봐서는 알 길이 없다.

    기제는 `main` 의 early return 이다 — `if errors: … return 1` 이 착지·요구 수를
    찍는 자리를 건너뛴다. task 1.9 가 적은 "항상 출력한다"는 성공할 때만 참이었다."""

    def _failing_fixture(
        self, raw: tempfile.TemporaryDirectory, *, with_evidence: bool,
        evidence_at: str = "current", also_unchanged: bool = False,
        base_shaped_extra: int = 0,
    ):
        """base P → 내 작업 L → 이웃 change N (→ 증거 E). 착지 기록은 없다.

        `evidence_at` 은 번들이 **어느 시점의 소스**를 적는지다. `current` 는 오늘,
        `base` 는 base 의 것(FLM 을 먼저 쓰고 편집 뒤 갱신하지 않은 모양 = V1),
        `intermediate` 는 base 뒤이되 오늘은 아닌 것(stale 이지만 base 를 기술하지 않는다).
        """
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        other = root / "internal" / "other.go"
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        if also_unchanged:
            # base 이후 **한 번도 안 바뀌는** 파일. 그 번들은 신선한데도 내용이
            # base 와 같다 — 등식만 보면 base 를 기술하는 것과 구별되지 않는다.
            (root / "internal" / "still.go").write_text(
                "package internal\nfunc Still() int { return 1 }\n")
        extras = [root / "internal" / f"extra{n}.go" for n in range(1, base_shaped_extra + 1)]
        for n, extra in enumerate(extras, start=1):
            extra.write_text(f"package internal\nfunc Extra{n}() int {{ return 1 }}\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        for n, extra in enumerate(extras, start=1):
            extra.write_text(f"package internal\nfunc Extra{n}() int {{ return 2 }}\n")
        marks["L"] = _commit_all(root, "L: my work lands")
        other.write_text("package internal\nfunc Other() int { return 2 }\n")
        marks["N"] = _commit_all(root, "N: a different change lands")
        if with_evidence:
            # 해시를 손으로 옮겨 적지 않는다 — base 의 blob 을 git 에서 읽어서 쓴다.
            blob = subprocess.check_output(
                ["git", "show", f"{marks['P']}:internal/own.go"], cwd=root,
            ) if evidence_at == "base" else own.read_bytes()
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(blob).hexdigest(),
            )
            for n, extra in enumerate(extras, start=1):
                # 번들마다 base 의 blob 을 읽어서 적는다 — 전부 base 를 기술한다.
                _write_evidence(
                    change, package="internal", function=f"Extra{n}",
                    relative=f"internal/extra{n}.go",
                    digest=hashlib.sha256(subprocess.check_output(
                        ["git", "show", f"{marks['P']}:internal/extra{n}.go"], cwd=root,
                    )).hexdigest(),
                )
            if also_unchanged:
                still = root / "internal" / "still.go"
                _write_evidence(
                    change, package="internal", function="Still",
                    relative="internal/still.go",
                    digest=hashlib.sha256(still.read_bytes()).hexdigest(),
                )
            _commit_all(root, "E: evidence for my own function only")
            if evidence_at == "intermediate":
                # 번들이 적은 상태 뒤로 소스가 한 번 더 움직인다. 그러면 번들은
                # stale 이지만 base 를 기술하지는 않는다 — a066 · a071 의 모양이다.
                own.write_text("package internal\nfunc Own() int { return 3 }\n")
                marks["M"] = _commit_all(root, "M: my work moves on past the bundle")
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

    def test_a_bundle_that_records_the_base_is_not_told_to_record_a_landing(self) -> None:
        """조언 줄이 **자기가 못 잡게 될 증거**를 세탁하라고 권하면 안 된다 (task 7.1).

        V1(§6 로트 독립 리뷰가 낸 CRITICAL): 저장소 규칙대로 FLM 을 **먼저** 커밋하고
        코드를 편집한 뒤 번들을 갱신하지 않으면 5단계는 `AST source hash is stale` 로
        빨갛다. 그런데 **같은 출력이** `--record-landing` 을 권하고, 그 명령은 편집 전
        커밋을 착지로 계산해서 required 0 으로 통과시킨다. 게이트가 자기 조언 줄로
        자기 판정을 지우는 길이다.

        번들이 base 의 소스를 적었으면 도구가 받아들일 수 있는 착지는 **전부** 그
        함수가 아직 base 와 같은 지점이다. 그러므로 좁힌 창은 그 함수를 요구할 수
        없다 — 조언이 가리키는 곳에 얻을 것이 없다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(raw, with_evidence=True, evidence_at="base")
            code, printed = self._main_output(root)
            self.assertEqual(code, 1, printed)
            # 양성 대조 — 이 픽스처가 실제로 stale 경로여야 한다.
            self.assertIn("AST source hash is stale", printed, printed)
            self.assertNotIn(
                "--record-landing", printed,
                f"stale 증거 옆에서 착지 기록을 권하면 안 된다: {printed}",
            )
            advice = self._advice_line(printed)
            self.assertIn(
                "internal--own", advice,
                f"이름은 **조언 줄 안에** 있어야 한다 — 오류 줄에도 있으니 "
                f"출력 전체를 보면 바늘이 무뎌진다: {advice}",
            )
            # 사실 부분(창의 크기)은 그대로 남아야 한다 — 3.3 이 세운 줄이다.
            self.assertIn(
                "commit(s) that landed after the base", printed,
                f"창의 크기는 계속 말해야 한다: {printed}",
            )

    def test_a_stale_bundle_that_does_not_record_the_base_still_hears_the_advice(self) -> None:
        """stale 이라고 다 막지 않는다 — **base 를 기술하는** 것만 막는다 (task 7.1).

        2026-09-12 실측: 활성 11건이 워킹트리 모드에서 stale 로 보이고 그중 9건의
        stale 번들이 base 상태를 기술한다. 나머지 둘(a066 · a071)은 base 뒤의 상태를
        적었으므로 착지를 기록하면 창이 **실제로** 좁아진다 — 그쪽 조언은 살려 둔다.
        이 시험이 없으면 "stale 이면 무조건 막는다"가 위 시험을 그대로 통과한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(
                raw, with_evidence=True, evidence_at="intermediate",
            )
            code, printed = self._main_output(root)
            self.assertEqual(code, 1, printed)
            # 양성 대조 — 이쪽도 stale 이다. 다른 것은 **무엇을 기술하느냐** 하나다.
            self.assertIn("AST source hash is stale", printed, printed)
            self.assertIn(
                "--record-landing", printed,
                f"좁힐 것이 남은 change 는 조언을 잃으면 안 된다: {printed}",
            )

    def test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice(self) -> None:
        """거부할 **정상 입력**을 먼저 적는다 [[fail-closed-must-name-what-it-rejects]].

        base 이후 안 바뀐 파일의 번들은 신선한데도 `sha256(오늘) == sha256(base)` 다.
        등식만 보면 V1 과 똑같이 생겼다. 그래서 판정은 등식 하나가 아니라 **신선한
        번들을 먼저 건너뛰는** 것과 짝이어야 한다 — 그 `continue` 가 빠지면 멀쩡한
        change 가 조언을 잃는다. 이 시험이 그 자리를 못 박는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(
                raw, with_evidence=True, also_unchanged=True,
            )
            code, printed = self._main_output(root)
            self.assertEqual(code, 1, printed)
            # 양성 대조 — 그 번들이 신선해야(= stale 오류가 없어야) 이 시험이 뜻을 갖는다.
            self.assertNotIn("internal--still: AST source hash is stale", printed, printed)
            self.assertIn(
                "--record-landing", printed,
                f"신선한 번들은 세탁의 모양이 아니다 — 조언을 잃으면 안 된다: {printed}",
            )
            self.assertNotIn("record the source as it stood at the base", printed, printed)

    def _advice_line(self, printed: str) -> str:
        """조언 줄 **하나**를 집어낸다. 단언을 출력 전체에 걸면 오류 줄이 대신 만족시킨다."""
        lines = [line for line in printed.splitlines() if "so this window also holds" in line]
        self.assertEqual(1, len(lines), f"조언 줄이 정확히 하나여야 한다: {printed}")
        return lines[0]

    def _commands_refusing_shapes(self) -> dict[str, object]:
        """조언 줄이 `--record-landing` 을 권했는데 그 명령이 **거절하던** 모양들 (task 7.7).

        리뷰 I8 이 넷을 셌다(디스크에만 있는 기록 · staged 아카이브 이동 · 빌리는 쪽 ·
        번들 0). 수리 전에 다시 재니 **셋이 더** 있었다 — 더러운 트리 · 커밋 전 번들 ·
        끊긴 심링크 기록([[caller-count-is-not-fix-site-count]]). 일곱 다 명령이 걷기
        **전에** 멈추는 자리다. 걷고 나서야 아는 거절은 아래 따로 잰다.
        """
        def on_disk_only(raw):
            root, _ = self._failing_fixture(raw, with_evidence=True)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)  # 기록은 쓰였고 커밋만 안 됐다
            return root, "mine"

        def archive_move_staged(raw):
            root, _ = self._failing_fixture(raw, with_evidence=True)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            (root / "openspec" / "changes" / "archive").mkdir(parents=True)
            subprocess.run(
                ["git", "mv", "openspec/changes/mine", "openspec/changes/archive/2026-09-13-mine"],
                cwd=root, check=True,
            )
            return root, "mine"

        def borrower(raw):
            root, _ = _borrowed_fixture(raw, later_work=True)
            return root, "ordinary"

        def no_bundles(raw):
            root, _ = self._failing_fixture(raw, with_evidence=False)
            return root, "mine"

        def dirty_tree(raw):
            root, _ = self._failing_fixture(raw, with_evidence=True)
            (root / "openspec" / "changes" / "mine" / "review.md").write_text("mine, edited\n")
            return root, "mine"

        def bundles_not_committed(raw):
            root, _ = self._failing_fixture(raw, with_evidence=False)
            _write_evidence(
                root / "openspec" / "changes" / "mine", package="internal", function="Own",
                relative="internal/own.go",
                digest=hashlib.sha256((root / "internal" / "own.go").read_bytes()).hexdigest(),
            )
            return root, "mine"

        def dangling_symlink_record(raw):
            root, _ = self._failing_fixture(raw, with_evidence=True)
            (root / "openspec" / "changes" / "mine" / "landed-commit.txt").symlink_to(
                root / "nowhere.txt"
            )
            return root, "mine"

        return {
            "on_disk_only": on_disk_only,
            "archive_move_staged": archive_move_staged,
            "borrower": borrower,
            "no_bundles": no_bundles,
            "dirty_tree": dirty_tree,
            "bundles_not_committed": bundles_not_committed,
            "dangling_symlink_record": dangling_symlink_record,
        }

    def test_the_advice_never_names_a_command_that_would_refuse(self) -> None:
        """조언 줄은 명령이 **스스로 말할 거절**을 대신 말한다 (task 7.7, 리뷰 I8).

        옛 조언 줄은 명령의 거절 조건을 하나도 묻지 않고 명령을 권했다. 디스크에만 기록이
        있는 change 에서 게이트는 "기록이 없다, 기록하라"고 하고 명령은 "이미 있다"고
        했다 — 두 문장이 서로를 가리키며 돈다.

        단언은 **명령의 문장이 조언 줄 안에 있다**는 것이다. 문장을 시험에 옮겨 적지
        않는다: 옮겨 적으면 두 사본이 같이 바뀌어도 초록이다. 명령을 **실제로 돌려서**
        나온 문장을 바늘로 쓴다.
        """
        for name, build in self._commands_refusing_shapes().items():
            with self.subTest(shape=name):
                raw = tempfile.TemporaryDirectory()
                with raw:
                    root, change = build(raw)
                    _, printed = _cli(root, change=change)
                    advice = self._advice_line(printed)
                    code, recorded = _cli(root, "--record-landing", change=change)
                    # 양성 대조 — 이 모양에서 명령이 **실제로** 거절해야 이 시험이 뜻을 갖는다.
                    self.assertEqual(code, 1, recorded)
                    self.assertFalse(
                        (root / "nowhere.txt").exists(), "거절이 심링크를 따라 썼다",
                    )
                    reason = recorded.strip().split(f"{change}: ", 1)[1]
                    self.assertIn(reason, advice, f"명령의 거절을 조언이 말해야 한다: {printed}")
                    self.assertNotIn(
                        "to let the gate compute", advice,
                        f"거절할 명령을 권하면 안 된다: {advice}",
                    )
                    # 대상 텍스트도 **기록이 디스크에 있어도** 참이어야 한다. 옛 판본은
                    # 디스크에 파일이 있는데 "(no landed-commit.txt)" 라고 찍었다.
                    self.assertNotIn(f"(no {check_analysis.LANDING_FILE})", printed, printed)
                    self.assertIn(f"(no {check_analysis.LANDING_FILE} in HEAD)", printed, printed)

    def test_a_record_on_disk_is_named_as_not_committed(self) -> None:
        """디스크에만 있는 기록은 **무엇을 하면 되는지**를 말해야 한다 — "이미 있다"만으로는
        옛 두 문장 사이의 고리가 안 풀린다. 게이트가 기록을 커밋에서 읽는다는 것이 그 답이다.

        HEAD 에 있는 기록의 문장과 **갈라야** 한다. 한 문장이면 "커밋하라"가 이미 커밋된
        기록 앞에서 거짓말이 된다 — 아래 둘째 단언이 그 갈래를 잰다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._failing_fixture(raw, with_evidence=True)
            self.assertEqual(check_analysis.record_landing("mine", root)[0], 0)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertEqual(len(lines), 1, lines)
            self.assertIn("openspec/changes/mine/landed-commit.txt", lines[0])
            self.assertIn("not in HEAD", lines[0])
            _commit_all(root, "commit the record")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(
                (code, lines),
                (1, [f"mine: `landed-commit.txt` already exists — not overwritten; "
                     f"{check_analysis.LANDING_RECOVERY}"]),
            )

    def test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not(self) -> None:
        """조언 줄은 사유를 **하나**만 말한다 — 그러니 순서가 곧 조언이다 (task 7.7).

        추적 파일 수정은 커밋하면 사라진다. 그것을 먼저 말하면 영원히 기록할 수 없는
        change(빌리는 쪽 · 번들 0)가 "먼저 커밋하라"를 듣고, 커밋한 뒤에야 진짜 사유를
        듣는다. 이 저장소의 활성 번들 0 change 일곱은 tasks.md 한 줄만 고쳐도 그 상태다.
        첫 판에서 이 순서를 재는 시험이 0 이었다(변이 R10 SURVIVED)."""
        def borrower_in_a_dirty_tree(raw):
            root, _ = _borrowed_fixture(raw)
            (root / "openspec" / "changes" / "ordinary" / "review.md").write_text("edited\n")
            return root, "ordinary", "borrows its evidence"

        def no_bundles_in_a_dirty_tree(raw):
            root, _ = self._failing_fixture(raw, with_evidence=False)
            (root / "openspec" / "changes" / "mine" / "review.md").write_text("edited\n")
            return root, "mine", "no `revision: current` evidence pins a landing"

        for build in (borrower_in_a_dirty_tree, no_bundles_in_a_dirty_tree):
            with self.subTest(shape=build.__name__):
                raw = tempfile.TemporaryDirectory()
                with raw:
                    root, change, permanent = build(raw)
                    # 양성 대조 — 트리가 **실제로** dirty 여야 이 순서를 잰다.
                    self.assertEqual(
                        subprocess.run(["git", "diff", "--quiet", "HEAD"], cwd=root).returncode, 1,
                    )
                    advice = self._advice_line(_cli(root, change=change)[1])
                    self.assertIn(permanent, advice)
                    self.assertNotIn("uncommitted changes", advice)
                    code, lines = check_analysis.record_landing(change, root)
                    self.assertEqual(code, 1, lines)
                    self.assertIn(permanent, lines[0])

    def test_where_the_command_would_record_the_advice_still_names_it(self) -> None:
        """양성 대조군 — 거절이 없는 change 는 조언을 잃으면 안 된다
        [[fail-closed-must-name-what-it-rejects]]. 권한 명령을 **곧바로 돌려서** 기록되는지까지
        잰다: 조언 줄이 약속한 것을 명령이 지키는 왕복이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._failing_fixture(raw, with_evidence=True)
            advice = self._advice_line(_cli(root)[1])
            self.assertIn("--record-landing` to let the gate compute", advice)
            self.assertNotIn("cannot narrow it", advice)
            code, recorded = _cli(root, "--record-landing")
            self.assertEqual(code, 0, recorded)

    def test_a_refusal_only_the_walk_finds_is_not_promised_away(self) -> None:
        """걷고 나서야 아는 거절은 조언 줄이 **예측하지 않는다** — 대신 약속도 안 한다.

        어느 커밋도 번들과 안 맞는지는 후보를 전부 걸어야 안다. 그 walk 는 change 하나에
        133초까지 걸리고(리뷰 I1) 조언 줄은 워킹트리가 대상인 **모든** 실행에서 나간다.
        그래서 권하되 "기록된다"고 약속하지 않고, 못 찾으면 명령이 그렇게 말한다고 적는다.
        옛 문장("compute and record … which narrows it")은 이 모양에서 거짓이었다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._failing_fixture(raw, with_evidence=False)
            _write_evidence(
                root / "openspec" / "changes" / "mine", package="internal", function="Own",
                relative="internal/own.go", digest=hashlib.sha256(b"never committed").hexdigest(),
            )
            _commit_all(root, "E: evidence no commit matches")
            advice = self._advice_line(_cli(root)[1])
            code, recorded = _cli(root, "--record-landing")
            self.assertEqual(code, 1, recorded)
            self.assertIn("is accepted as the landing", recorded)  # 걷고 나서야 나오는 문장
            self.assertIn("--record-landing` to let the gate compute", advice)
            self.assertIn("says so instead of recording", advice, advice)
            # 조언이 말하는 조건도 명령과 같아야 한다 (task 7.2.2). 증거와 **맞는데** 규칙이 거절하는
            # 후보가 생겼으므로 "맞는 커밋이 없으면"은 명령이 거절하는 경우를 다 덮지 못한다.
            self.assertIn("if no commit on this history is accepted as the landing", advice, advice)

    def test_a_fault_while_asking_the_recorder_does_not_become_advice(self) -> None:
        """조언을 위해 부른 git 이 멎어도 **판정**은 그대로 찍히고 명령을 권하지 않는다.

        조언은 판정이 아니다. 조언 때문에 부른 것이 판정 줄을 스택으로 바꾸거나, 모르는
        채로 명령을 권하면 안 된다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._failing_fixture(raw, with_evidence=True)
            _, before = _cli(root)
            real_run = subprocess.run

            def the_dirty_check_hangs(args: object, **kwargs: object):
                if isinstance(args, list) and "--quiet" in args:
                    raise subprocess.TimeoutExpired(cmd=args, timeout=30)
                return real_run(args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", the_dirty_check_hangs):
                code, printed = _cli(root)
            self.assertEqual(code, 1, printed)
            advice = self._advice_line(printed)
            self.assertIn("timed out", advice)
            self.assertNotIn("to let the gate compute", advice)
            # 판정 줄은 조언과 무관하게 같아야 한다.
            verdict = [line for line in printed.splitlines() if "so this window also holds" not in line]
            self.assertEqual(
                verdict,
                [line for line in before.splitlines() if "so this window also holds" not in line],
            )

    def test_only_three_bundles_are_named_and_the_rest_are_counted(self) -> None:
        """이름을 다 쏟아내지 않는다 — a076 의 **21,838자 한 줄**이 그 이유다.

        2026-09-12 실측: a092 는 base 를 기술하는 번들이 **29개**다. 자르고 세는
        갈래가 실물 경로이므로 시험 없이 두지 않는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._failing_fixture(
                raw, with_evidence=True, evidence_at="base", base_shaped_extra=3,
            )
            code, printed = self._main_output(root)
            self.assertEqual(code, 1, printed)
            advice = self._advice_line(printed)
            self.assertIn("4 of this change's", advice, f"총 수는 세야 한다: {advice}")
            self.assertIn("and 1 more", advice, f"못 적은 것은 세어서 말해야 한다: {advice}")
            self.assertEqual(
                3, advice.count("internal--"), f"이름은 셋까지만 적어야 한다: {advice}",
            )

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
            own.write_text("package internal\n\nfunc Own() int { return 2 }\n\n// owner: nobody yet\n")
            base = _commit_all(root, "P: base")
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            (change / "base-commit.txt").write_text(base + "\n")
            (change / "review.md").write_text("mine\n")
            # 요구 집합은 diff 조각에 닿은 **기존** 함수만 센다. 함수에서 떨어진 주석 한 줄만 바꾸면
            # 고정 소스는 바뀌고(착지가 유효하다, task 7.2.2) 요구는 0 이다. 예전 픽스처는 작업이
            # base **앞**에 있는 a074 모양이었는데, 그 모양은 이제 착지를 얻지 못한다 —
            # `ALandingMustChangeWhatItsEvidencePins` 가 그것을 잰다. (함수를 파일 끝에 붙이면
            # `--unified=0` 조각의 경계가 `Own` 에 닿아 요구가 1 이 된다 — 첫 판에서 실측.)
            own.write_text("package internal\n\nfunc Own() int { return 2 }\n\n// owner: the operator team\n")
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            (root / "docs.md").write_text("post-deployment measurement\n")
            landing = _commit_all(root, "L: a comment away from any function, no existing function changes")
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
            # SHA 만 찾으면 라벨을 이관의 `audited source-commit` 으로 바꿔도 초록이다(task 6.3,
            # 변이 T1 실측 SURVIVED) — 그러면 출력이 있지도 않은 감사를 가리킨다. 라벨까지 본다.
            self.assertIn(f"landed-commit {landing}", printed, f"해소한 착지를 라벨과 함께 말해야 한다: {printed}")
            self.assertNotIn("audited source-commit", printed)


def _self_repair_fixture(
    raw: tempfile.TemporaryDirectory, *, repair: str = "pinned", record: bool = True
) -> tuple[Path, dict[str, str]]:
    """P(base) → W(작업) → E(증거) → 정직한 기록 → F(리뷰 수리). 기록은 **생산 경로**가 만든다.

    로트마다 도는 리뷰 수리 라운드의 모양이다 (task 7.2.6, H4). `repair` 가 F 의 모양을 고른다:

    - `pinned`     — 고정 파일과 change 디렉터리를 **같은 커밋**에서 (실물 a083 · a084 의 모양)
    - `other-go`   — 고정되지 않은 Go 와 change 디렉터리를 같은 커밋에서 (ANYGO 가 덮는 자리)
    - `docs-only`  — change 디렉터리만 (Go 없음)
    - `neighbour`  — Go 만, 남의 커밋 (거절하면 안 되는 자리)
    - `split`      — Go 커밋과 문서 커밋으로 **쪼갠다** (알려진 한계)
    - `pinned-refreshed` — 수리와 번들 갱신이 한 커밋에 (그 커밋 자신이 착지가 되는 자리)
    - `none`       — 수리 없음
    """
    root = _init_fixture(raw)
    own = root / "internal" / "own.go"
    own.parent.mkdir(parents=True)
    own.write_text("package internal\nfunc Own() int { return 1 }\n")
    other = root / "internal" / "other.go"
    other.write_text("package internal\nfunc Other() int { return 1 }\n")
    marks = {"P": _commit_all(root, "P: base")}
    change = root / "openspec" / "changes" / "mine"
    change.mkdir(parents=True)
    (change / "base-commit.txt").write_text(marks["P"] + "\n")
    (change / "review.md").write_text("mine\n")
    # change 디렉터리는 **따로** 커밋한다. 작업과 한 커밋에 넣으면 그 커밋부터 깃발이 서서
    # 시험이 재려는 자리(기록 **뒤**)가 아니라 그 앞을 재게 된다.
    marks["D"] = _commit_all(root, "D: the change directory")
    own.write_text("package internal\nfunc Own() int { return 2 }\n")
    marks["W"] = _commit_all(root, "W: this change's Go work lands")
    _write_evidence(
        change, package="internal", function="Own", relative="internal/own.go",
        digest=hashlib.sha256(own.read_bytes()).hexdigest(),
    )
    marks["E"] = _commit_all(root, "E: its evidence enters the history")
    # 값을 시험이 고르지 않는다 — 게이트가 계산한 것을 그대로 커밋한다.
    if record:
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        marks["record"] = (change / check_analysis.LANDING_FILE).read_text(encoding="utf-8").strip()
        marks["recorded"] = _commit_all(root, "the honest record")
    if repair == "none":
        return root, marks
    if repair == "split":
        own.write_text("package internal\nfunc Own() int { return 3 }\n")
        marks["F1"] = _commit_all(root, "F1: the fix alone")
        (change / "review.md").write_text("mine\nreview round 1\n")
        marks["F2"] = _commit_all(root, "F2: the notes alone")
        return root, marks
    if repair in ("pinned", "other-go", "docs-only", "pinned-refreshed"):
        (change / "review.md").write_text("mine\nreview round 1\n")
    if repair in ("pinned", "neighbour", "pinned-refreshed"):
        own.write_text("package internal\nfunc Own() int { return 3 }\n")
    if repair == "pinned-refreshed":
        # 수리와 번들 갱신이 **한 커밋**에 있는 정직한 로트. 이 커밋 자신이 착지가 된다.
        _write_evidence(
            change, package="internal", function="Own", relative="internal/own.go",
            digest=hashlib.sha256(own.read_bytes()).hexdigest(),
        )
    if repair == "other-go":
        other.write_text("package internal\nfunc Other() int { return 3 }\n")
    marks["F"] = _commit_all(root, f"F: review fix ({repair})")
    return root, marks


class WorkAfterTheRecordIsNotOutsideTheWindow(unittest.TestCase):
    """착지를 기록한 **뒤에** 이 change 자신이 Go 를 더 고치면 그 기록은 착지가 아니다.

    task 7.2.6 (H4) — 사람이 2026-09-16 에 변형 B-ANYGO 를 골랐다. 기록이 한 번 쓰이고 나면
    5단계는 `base..기록` 만 대조하므로, 기록 뒤의 리뷰 수리는 **고정 파일**을 고쳐도 창 밖에
    남아 초록이었다. 잊어버린 저자가 초록이고 번들을 갱신한 성실한 저자가 빨간 역전이었다.

    내용으로는 못 가른다 — 고정 소스가 착지와 지금 같아야 한다는 후보 규칙은 착지 있는 68건
    중 61(파일)·55(함수)를 거절했고 원인은 거의 전부 이웃이었다(review.md
    `## MEASURE — task 7.2.6`). 가르는 신호는 **같은 커밋이 이 change 의 디렉터리도 만졌는가**
    하나뿐이었다. 그 신원은 **거절에만** 쓴다(1.12).
    """

    def _analysis(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"

    def _only_this_rule_refuses(self, root: Path, marks: dict[str, str]) -> None:
        """닿음 단언: 깃발을 비우면 **같은 후보**가 받아들여진다.

        안 하면 다른 가드(불일치·미보유)가 이미 거절하는 입력에서도 이 시험이 초록이고,
        새 규칙은 아무것도 재지 않는다([[mutation-must-reach-the-thing-under-test]]).
        """
        analysis = self._analysis(root)
        self.assertEqual(
            check_analysis._landing_refusal(
                root, marks["P"], marks["record"],
                _inputs(root, analysis,
                        floor=check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)),
                        repairs=[]),
            ),
            ("", []),
        )

    def test_a_review_fix_to_a_pinned_file_after_the_record_is_refused(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            self._only_this_rule_refuses(root, marks)
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(
                    f"landing point {marks['record'][:12]} is followed by 1 later commit(s) "
                    "of this change's own Go work" in error and marks["F"][:12] in error
                    for error in errors
                ),
                errors,
            )

    def test_the_recorder_will_not_write_a_landing_its_own_later_work_outruns(self) -> None:
        """계산 경로도 **같은 함수**에 묻는다 (task 7.6, 리뷰 I2).

        기록이 아직 없는 상태에서 수리가 먼저 서면, 게이트는 그 앞 커밋을 착지로 쓰지 않는다.
        이 시험이 없으면 계산 경로에서만 규칙을 빼는 변이가 살아남는다 — 그러면 도구가
        자기 게이트가 거절할 값을 쓴다([[two-judgements-cover-for-each-other]]).
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _self_repair_fixture(raw, repair="pinned", record=False)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertIn("is followed by 1 later commit(s) of this change's own Go work", lines[0])

    def test_the_refusal_says_how_to_move_the_record(self) -> None:
        """거절만 하고 길을 안 말하면 성실한 저자가 두 문장 사이에 갇힌다 (task 7.2.6)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(check_analysis.LANDING_RECOVERY in error for error in errors), errors
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertIn(check_analysis.LANDING_RECOVERY, lines[0])

    def test_without_a_record_the_same_tree_was_already_red(self) -> None:
        """역전이 결함이었다는 근거: 잊은 저자만 기록으로 초록이 됐다.

        같은 트리에서 기록을 지우면 5단계는 워킹트리를 대상으로 삼아 stale 을 잡는다.
        규칙이 없애는 것은 그 **비대칭**이다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _self_repair_fixture(raw, repair="pinned")
            (root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE).unlink()
            _commit_all(root, "drop the record")
            errors = check_analysis.check("mine", root)
            self.assertTrue(any("AST source hash is stale" in error for error in errors), errors)

    def test_any_go_file_counts_not_only_the_pinned_one(self) -> None:
        """사람이 고른 것은 ANYGO 다 — 한 커밋 모양의 H2 까지 덮는다 (6.6)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="other-go")
            self._only_this_rule_refuses(root, marks)
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any("is followed by 1 later commit(s)" in error for error in errors), errors
            )

    def test_a_neighbours_later_go_edit_is_not_this_changes_own_repair(self) -> None:
        """거절할 정상 입력을 재는 자리다. 이웃이 나중에 같은 파일을 고쳐도 기록은 유효하다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="neighbour")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), marks["record"])

    def test_touching_the_change_without_editing_go_is_not_a_repair(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="docs-only")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), marks["record"])

    def test_splitting_the_fix_from_its_notes_evades_the_guard(self) -> None:
        """**알려진 한계** (task 7.2.6). 망각 가드이지 위조 가드가 아니다.

        Go 수리와 문서를 다른 커밋으로 쪼개면 어느 커밋도 둘 다 만지지 않아 깃발이 안 선다.
        spec 이 이 한계를 적는다. 한계를 시험으로 못 박아 두면 없어질 때 보인다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="split")
            self.assertEqual(check_analysis._self_repair_commits(root, self._analysis(root), _head(root)), [])
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), marks["record"])

    def test_refreshing_the_evidence_and_recording_again_lands_after_the_fix(self) -> None:
        """복구 경로가 **실제로** 돈다 — 문장이 가리키는 곳에 길이 있어야 한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            change = root / "openspec" / "changes" / "mine"
            (change / check_analysis.LANDING_FILE).unlink()
            _commit_all(root, "remove the record it outgrew")
            own = root / "internal" / "own.go"
            _write_evidence(
                change, package="internal", function="Own", relative="internal/own.go",
                digest=hashlib.sha256(own.read_bytes()).hexdigest(),
            )
            refreshed = _commit_all(root, "refresh the bundle")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            self.assertEqual((change / check_analysis.LANDING_FILE).read_text(encoding="utf-8").strip(), refreshed)
            _commit_all(root, "the record again")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), refreshed)
            # 착지가 수리 **뒤로** 왔다: 기록 시점의 값보다 나중이다.
            self.assertNotEqual(refreshed, marks["record"])

    def test_a_candidate_its_evidence_does_not_describe_keeps_that_sentence(self) -> None:
        """**자리가 못의 일부다** ([[a-new-guard-unpins-the-guards-behind-it]]).

        수리 커밋 F 는 두 가지가 동시에 참이다 — 번들이 F 의 소스를 기술하지 않고(불일치),
        그 뒤에 수리가 하나 더 있다. 진단은 **불일치**여야 한다: 저자가 고칠 것은 번들이고,
        "나중 작업이 앞선다"는 그 다음 이야기다. 규칙을 앞으로 옮기면 이 문장이 바뀐다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            change = root / "openspec" / "changes" / "mine"
            (root / "internal" / "own.go").write_text(
                "package internal\nfunc Own() int { return 4 }\n"
            )
            (change / "review.md").write_text("mine\nreview round 2\n")
            second = _commit_all(root, "F2: another review fix")
            analysis = self._analysis(root)
            repairs = check_analysis._self_repair_commits(root, analysis, _head(root))
            self.assertEqual(repairs, [marks["F"], second])
            refusal, _ = check_analysis._landing_refusal(
                root, marks["P"], marks["F"],
                _inputs(root, analysis,
                        floor=check_analysis._evidence_floor(root, _bundles(root, analysis), _head(root)),
                        repairs=repairs),
            )
            self.assertIn("is not the revision this evidence describes", refusal)
            self.assertNotIn("is followed by", refusal)

    def test_the_signal_survives_archiving_the_change(self) -> None:
        """아카이브는 디렉터리를 **옮긴다**. 옮긴 뒤 이름만 보면 활성 시절의 자기 수리가
        전부 안 보이고, 아카이브된 change 의 재검사가 조용히 초록이 된다 (6.2 가 연 경로).
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            archive = root / "openspec" / "changes" / "archive" / "2026-09-16-mine"
            archive.parent.mkdir(parents=True, exist_ok=True)
            subprocess.run(
                ["git", "mv", "openspec/changes/mine", archive.relative_to(root).as_posix()],
                cwd=root, check=True,
            )
            _commit_all(root, "archive the change")
            analysis = archive / "analysis" / "function-logic"
            self.assertEqual(check_analysis._self_repair_commits(root, analysis, _head(root)), [marks["F"]])
            errors = check_analysis.check("mine", root)
            self.assertTrue(any("is followed by 1 later commit(s)" in error for error in errors), errors)

    def test_the_repair_commit_itself_is_the_landing_when_it_refreshes_the_evidence(self) -> None:
        """자기 자신은 세지 않는다 — `_is_ancestor` 는 같은 커밋에서 참이다.

        수리와 번들 갱신을 **한 커밋**에 넣는 것이 이 저장소의 정직한 로트 모양이다. 자기
        자신까지 세면 그런 change 는 착지를 영영 못 얻고, 규칙이 성실한 쪽을 벌한다.

        닿음: 그 커밋이 깃발 목록에 **있는데도** 착지가 된다는 것을 같이 단언한다 —
        안 하면 깃발이 비어서 초록인 경우와 구분되지 않는다
        ([[mutation-must-reach-the-thing-under-test]]).
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned-refreshed", record=False)
            self.assertEqual(
                check_analysis._self_repair_commits(root, self._analysis(root), _head(root)), [marks["F"]]
            )
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            change = root / "openspec" / "changes" / "mine"
            self.assertEqual(
                (change / check_analysis.LANDING_FILE).read_text(encoding="utf-8").strip(),
                marks["F"],
            )
            _commit_all(root, "the record")
            facts: dict[str, object] = {}
            self.assertEqual(check_analysis.check("mine", root, facts), [])
            self.assertEqual(facts.get("landing"), marks["F"])


class TheRepairSignalIsMeasuredOnceAndNamesOldestFirst(unittest.TestCase):
    """깃발 집합은 **한 곳**에서 만든다 (`_self_repair_commits`).

    두 벌이면 선언 경로와 계산 경로가 갈리고, 갈리는 순간 도구는 자기 게이트가 거절할 값을
    쓴다([[two-judgements-cover-for-each-other]]).
    """

    def _analysis(self, root: Path) -> Path:
        return root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"

    def test_a_fix_on_a_side_branch_is_read(self) -> None:
        """곁가지의 **비병합** 수리는 읽힌다. 한계는 병합 커밋 **자신의** 변경이고, 그것은
        아래 `…_is_the_known_limit` 이 잰다 — 이름이 재지 않는 것을 약속하면 나중 독자가
        그 갈래를 덮인 것으로 센다(2026-09-16 적대 리뷰 F7)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="none")
            change = root / "openspec" / "changes" / "mine"
            home = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
            ).strip()
            subprocess.run(["git", "checkout", "-q", "-b", "side"], cwd=root, check=True)
            (root / "internal" / "own.go").write_text(
                "package internal\nfunc Own() int { return 3 }\n"
            )
            (change / "review.md").write_text("mine\nside\n")
            _commit_all(root, "side: the fix and its notes")
            subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
            subprocess.run(["git", "merge", "-q", "--no-ff", "side"], cwd=root, check=True,
                           capture_output=True)
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            # 곁가지의 그 커밋은 비병합이므로 **읽힌다**. 한계는 병합 커밋 자신의 변경이다.
            self.assertEqual(len(check_analysis._self_repair_commits(root, analysis, _head(root))), 1)
            self.assertNotEqual(check_analysis.check("mine", root), [])

    def test_a_fix_made_inside_the_merge_commit_itself_is_the_known_limit(self) -> None:
        """병합 커밋 **자신의** 변경(충돌 해소)은 안 읽힌다 — 7.2.4 의 H7 과 같은 부류.

        `git log` 는 `--diff-merges` 를 명시해야만 병합의 diff 를 낸다(git 2.43 실측: 어떤
        설정으로도 기본이 바뀌지 않는다). 그러므로 이 한계는 저장소 설정의 함수가 아니다.
        한계를 시험으로 못 박아 두면 없어질 때 보인다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _self_repair_fixture(raw, repair="none")
            change = root / "openspec" / "changes" / "mine"
            home = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
            ).strip()
            subprocess.run(["git", "checkout", "-q", "-b", "other"], cwd=root, check=True)
            (root / "internal" / "other.go").write_text(
                "package internal\nfunc Other() int { return 9 }\n"
            )
            _commit_all(root, "other: an unrelated branch")
            subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
            subprocess.run(["git", "merge", "-q", "--no-ff", "--no-commit", "other"],
                           cwd=root, check=True, capture_output=True)
            # 병합을 마치며 **그 커밋 안에서** Go 를 고치고 change 디렉터리도 만진다.
            (root / "internal" / "own.go").write_text(
                "package internal\nfunc Own() int { return 3 }\n"
            )
            (change / "review.md").write_text("mine\nfixed while merging\n")
            merge = _commit_all(root, "M: the merge, and a fix made inside it")
            self.assertEqual(
                len(subprocess.check_output(
                    ["git", "rev-list", "--parents", "-1", merge], cwd=root, text=True
                ).split()), 3, "M 은 병합 커밋이어야 한다",
            )
            analysis = change / "analysis" / "function-logic"
            self.assertEqual(check_analysis._self_repair_commits(root, analysis, _head(root)), [])

    def test_a_fix_the_merge_hides_is_still_read(self) -> None:
        """**가지치기가 일어나는** 병합 픽스처 (2026-09-16 적대 리뷰 F1, P0).

        경로 제한을 건 `git log` 는 기본으로 역사를 단순화한다 — 병합이 **그 경로에 대해** 한
        부모와 TREESAME 이면 반대편 가지를 통째로 버린다. 그러면 곁가지에서 한 자기 수리가
        목록에서 사라져, 이 규칙이 닫으려는 H4 가 바로 이 change 의 제목이 가리키는 상황(병합)
        에서 다시 열린다. 실측으로 재현했다: `check` 가 `[]` 초록인데 고정 파일은 기록 뒤에 바뀌어
        있었다. 선형 픽스처는 이 가드를 안 건드린다([[linear-fixtures-never-exercise-dag-guards]]).
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="none")
            change = root / "openspec" / "changes" / "mine"
            own = root / "internal" / "own.go"
            home = subprocess.check_output(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True
            ).strip()
            subprocess.run(["git", "checkout", "-q", "-b", "side"], cwd=root, check=True)
            own.write_text("package internal\nfunc Own() int { return 3 }\n")
            (change / "review.md").write_text("mine\nside fix\n")
            hidden = _commit_all(root, "C1: the fix and its notes, on a side branch")
            subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
            # 줄기가 **같은 편집**을 해서 병합이 그 경로에 대해 줄기와 TREESAME 이 된다.
            (change / "review.md").write_text("mine\nside fix\n")
            _commit_all(root, "main: the same note edit")
            subprocess.run(["git", "merge", "-q", "--no-ff", "-m", "merge", "side"],
                           cwd=root, check=True, capture_output=True)
            # 닿음: 단순화가 **실제로** 그 커밋을 버린다는 것을 먼저 보인다.
            simplified = subprocess.check_output(
                ["git", "log", "--no-merges", "--format=%H", "HEAD", "--",
                 "openspec/changes/mine/"], cwd=root, text=True).split()
            self.assertNotIn(hidden, simplified, "가지치기가 안 일어나면 이 시험은 F1 을 안 잰다")
            self.assertEqual(check_analysis._self_repair_commits(root, self._analysis(root), _head(root)), [hidden])
            errors = check_analysis.check("mine", root)
            self.assertTrue(any("is followed by 1 later commit(s)" in error for error in errors), errors)

    def test_a_rename_out_of_go_is_read_whatever_the_machine_configures(self) -> None:
        """판정은 사람의 git 설정의 함수가 아니어야 한다 (적대 리뷰 F4).

        `diff.renames` 가 켜져 있으면 `.go` 를 비-`.go` 이름으로 옮긴 커밋의 **옛 이름**이
        목록에서 사라져 깃발이 안 선다. `_evidence_floor` 가 `-M100% --no-follow` 로 지킨
        원칙과 같다 — 명령줄에서 못 박는다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _self_repair_fixture(raw, repair="none")
            subprocess.run(["git", "config", "diff.renames", "true"], cwd=root, check=True)
            change = root / "openspec" / "changes" / "mine"
            subprocess.run(["git", "mv", "internal/other.go", "internal/moved.txt"],
                           cwd=root, check=True)
            (change / "review.md").write_text("mine\nmoved it\n")
            moved = _commit_all(root, "F: move a Go file out of Go while noting it")
            self.assertEqual(
                check_analysis._self_repair_commits(root, self._analysis(root), _head(root)), [moved]
            )

    def test_a_non_ascii_go_name_is_read(self) -> None:
        """`core.quotePath` 기본값은 비ASCII 이름을 인용해서 `.go` 로 안 끝나게 만든다
        (적대 리뷰 F5). 안 보이면 **거절을 안 해서** 창이 좁아지는 쪽으로 틀린다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _self_repair_fixture(raw, repair="none")
            change = root / "openspec" / "changes" / "mine"
            (root / "internal" / "한글.go").write_text(
                "package internal\nfunc Hangul() int { return 1 }\n", encoding="utf-8"
            )
            (change / "review.md").write_text("mine\nadded it\n")
            added = _commit_all(root, "F: a Go file with a non-ASCII name")
            self.assertEqual(
                check_analysis._self_repair_commits(root, self._analysis(root), _head(root)), [added]
            )

    def test_a_git_failure_becomes_a_verdict_not_a_silent_pass(self) -> None:
        """**실패는 위반 0 이 아니다** ([[missing-tool-reports-clean]] · 적대 리뷰 F2).

        빈 목록으로 물러나면 git 이 멎은 것과 "만진 커밋이 없다"가 같은 말이 되고, 거절돼야 할
        입력이 초록으로 지나간다(주입 실험에서 실측). 옆의 `_evidence_floor` 는 실패하면
        거절로 간다 — 두 자리가 같은 방향이어야 한다.
        """
        real = subprocess.run
        for needle, sentence in (("--full-history", "cannot list the commits that touched"),
                                 ("--no-walk", "cannot read the files those commits changed"),
                                 ("..HEAD", "cannot walk the history after")):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root, _ = _self_repair_fixture(raw, repair="pinned")

                def broken(command, *args, _needle=needle, **kwargs):
                    # `..HEAD` 는 `_repairs_after` 의 `rev-list <후보>..HEAD` **하나**만 고른다
                    # — 순회의 `rev-list --reverse` 까지 깨면 무엇이 판정을 냈는지 안 갈린다.
                    # 7.5.2.1 부터 끝은 상징 `HEAD` 가 아니라 명령이 푼 sha 다 — 모양(`<후보>..<끝>` 하나)으로 고른다.
                    hit = (_needle in command if _needle != "..HEAD"
                           else command[:2] == ["git", "rev-list"] and len(command) == 3
                           and ".." in command[2])
                    if isinstance(command, list) and hit:
                        return subprocess.CompletedProcess(command, 1, "", "fatal: injected\n")
                    return real(command, *args, **kwargs)

                with mock.patch.object(check_analysis.subprocess, "run", broken):
                    errors = check_analysis.check("mine", root)
                self.assertTrue(any(sentence in error for error in errors), (needle, errors))

    def test_a_borrowing_change_never_measures_the_lenders_directory(self) -> None:
        """`_self_repair_commits` 는 `analysis` 에서 change 디렉터리를 **유도**하는데, 빌린
        증거 경로에서 `analysis` 는 **빌려주는 쪽**으로 재바인딩된다 (적대 리뷰 F6).

        오늘은 도달하지 않는다 — 빌리는 change 의 기록은 이름으로 거절되고 선언이 없으면
        `resolve_landing` 이 먼저 돌아간다. 그 사실을 **시험으로 못 박는다**: 1.8 의 "빌려주는
        쪽 착지를 복사" 규칙이 되살아나면 이 시험이 빨개져서, 누구의 디렉터리를 재야 하는지를
        그때 사람이 정하게 된다. 지금 조용히 남의 디렉터리를 재는 것보다 낫다.
        """
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _borrowed_fixture(raw)
            seen: list[Path] = []
            real = check_analysis._self_repair_commits

            def spy(root_arg, analysis_arg):
                seen.append(analysis_arg)
                return real(root_arg, analysis_arg)

            with mock.patch.object(check_analysis, "_self_repair_commits", spy):
                self.assertEqual(check_analysis.check("ordinary", root), [])
            self.assertEqual(seen, [], "빌리는 경로는 이 신호를 재지 않는다")

    def test_the_oldest_repair_is_the_one_named(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            change = root / "openspec" / "changes" / "mine"
            (root / "internal" / "own.go").write_text(
                "package internal\nfunc Own() int { return 4 }\n"
            )
            (change / "review.md").write_text("mine\nreview round 2\n")
            second = _commit_all(root, "F2: another review fix")
            analysis = change / "analysis" / "function-logic"
            self.assertEqual(
                check_analysis._self_repair_commits(root, analysis, _head(root)), [marks["F"], second]
            )
            errors = check_analysis.check("mine", root)
            self.assertTrue(
                any(
                    f"followed by 2 later commit(s)" in error and f"(first {marks['F'][:12]})" in error
                    for error in errors
                ),
                errors,
            )


class ABlobIsFetchedOncePerCommitNotOncePerBundle(unittest.TestCase):
    """한 커밋에서 번들 N 개를 판정하는 데 드는 git 프로세스는 **N 개가 아니라 하나**다 (task 7.5).

    성능 수리는 되돌렸을 때 빨개지는 시험이 없으면 근거가 아니다
    ([[passing-test-is-not-evidence]]). 그래서 시간이 아니라 **프로세스 수**를 센다 —
    시간은 기계마다 다르고 부하에 흔들리지만 spawn 수는 알고리즘의 함수다.

    2026-09-18 실측: 후보를 하나도 안 받는 walk 에서 spawn 의 97.1~97.3% 가 blob fetch 였고
    (a071 12,155 중 11,799 · 73.19s), 그 전부가 `git show` 한 프로세스에 파일 하나였다.
    """

    def _repo(self) -> tuple[tempfile.TemporaryDirectory, Path, str]:
        raw = tempfile.TemporaryDirectory()
        root = _init_fixture(raw)
        internal = root / "internal"
        internal.mkdir(parents=True, exist_ok=True)
        for index in range(4):
            (internal / f"own{index}.go").write_text(
                f"package internal\nfunc Own{index}() int {{ return {index} }}\n"
            )
        return raw, root, _commit_all(root, "P: sources")

    def _spawns(self, call) -> tuple[object, int]:
        """`check_analysis` 가 돌린 프로세스 수를 센다. 감싸는 것은 모듈의 이름 하나다."""
        seen = []
        real = check_analysis.subprocess.run

        def counting(*args, **kwargs):
            seen.append(args[0] if args else kwargs.get("args"))
            return real(*args, **kwargs)

        with mock.patch.object(check_analysis.subprocess, "run", counting):
            value = call()
        return value, len(seen)

    def test_many_reads_the_same_bytes_as_one(self) -> None:
        """새 복수 읽기와 옛 단수 읽기는 **같은 바이트**를 낸다 — 없는 파일까지."""
        raw, root, commit = self._repo()
        with raw:
            wanted = [f"internal/own{index}.go" for index in range(4)] + ["internal/gone.go"]
            many = check_analysis._committed_many(root, commit, wanted)
            self.assertEqual(sorted(many), sorted(wanted))
            for relative in wanted:
                self.assertEqual(
                    many[relative],
                    check_analysis._committed_bytes(root, commit, relative),
                    relative,
                )
            self.assertIsNone(many["internal/gone.go"])
            self.assertIn(b"func Own2()", many["internal/own2.go"])

    def test_a_tree_is_not_a_blob(self) -> None:
        """디렉터리를 물으면 `None` 이다. `git show <ref>:<dir>` 는 **목록을 찍는다** —
        그 바이트가 판정에 들어오면 파일 내용인 척한다. 막는 쪽으로 엄해진다.

        **뒤에 파일을 더 묻는다.** 트리만 물으면 응답이 하나뿐이라 내용 건너뛰기가 틀려도
        아무 데도 안 드러난다 ([[mutation-must-reach-the-thing-under-test]]) — 변이 T4(비-blob
        내용을 안 건너뜀)·T7(크기를 NUL 탐색으로 계산)이 그래서 첫 판에 살아남았다. 트리의
        raw 바이트에는 NUL 이 들어 있으므로 둘 다 여기서 경계가 밀린다."""
        raw, root, commit = self._repo()
        with raw:
            fetched = check_analysis._committed_many(
                root, commit, ["internal", "internal/own0.go", "internal/own1.go"]
            )
            self.assertIsNone(fetched["internal"])
            # 트리 **뒤**의 자리들이 여전히 제 파일이다 — 응답 경계가 안 밀렸다.
            self.assertIn(b"func Own0()", fetched["internal/own0.go"])
            self.assertIn(b"func Own1()", fetched["internal/own1.go"])

    def test_a_blob_holding_nul_bytes_is_read_whole(self) -> None:
        """내용에 NUL 이 있어도 **선언된 크기**로 자른다. 내용에서 NUL 을 찾아 자르면
        바이너리 파일이 잘리고 그 뒤 자리들이 통째로 밀린다."""
        raw, root, _ = self._repo()
        with raw:
            binary = b"package internal\n// \x00\x00 embedded\nfunc Nul() int { return 0 }\n"
            (root / "internal" / "nul.go").write_bytes(binary)
            commit = _commit_all(root, "P: a blob with NUL bytes")
            fetched = check_analysis._committed_many(
                root, commit, ["internal/nul.go", "internal/own3.go"]
            )
            self.assertEqual(fetched["internal/nul.go"], binary)
            self.assertIn(b"func Own3()", fetched["internal/own3.go"])

    def test_git_failing_is_a_fault_not_an_absence(self) -> None:
        """git 이 실패하면 **결함**이다 — 물은 경로를 `None` 으로 채우지 않는다 (task 7.5.1).

        1.4 수리 때 이것을 결함으로 올리려다 "시험 21개가 저장소 아닌 곳에서 돈다" 는 이유로
        되돌렸고, 그 되돌림이 가드 7 과 `_landing_record` 의 permissive 구멍을 남겼다. 다시 재니
        그 21개는 git 을 mock 하려는 픽스처가 `_landing_record` 만 빠뜨린 것이었다 — 빈 저장소로
        바꾸면 전제("기록 없음")는 그대로다(실측: 빈 저장소 `missing` rc 0 · 저장소 아님 rc 128).
        """
        with tempfile.TemporaryDirectory() as raw:
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(Path(raw), "HEAD", ["a.go", "b.go"])
            # rc 가드의 문장이어야 한다 — 그 가드를 지우면 빈 출력이 "잘렸다" 로 **다른 가드**에
            # 걸려 타입 단언만으로는 초록이다 (재리뷰 testing 전문가).
            self.assertIn("not a git repository", str(caught.exception))

    def test_the_minimum_git_version_is_written_down_once(self) -> None:
        """`-Z` 를 요구한다는 사실이 **어디에도 안 적혀 있었다** (2026-09-18 독립 리뷰 P1).

        값은 한 곳(`GIT_BATCH_MINIMUM`)에 살고 산문이 그 수를 인용한다 — 두 벌이면 갈린다
        ([[two-judgements-cover-for-each-other]])."""
        self.assertEqual(check_analysis.GIT_BATCH_MINIMUM, "2.42")
        tool = Path(check_analysis.__file__).resolve().parent
        # 도구 옆의 README 는 사본 하네스에도 따라온다 — 언제나 잰다.
        self.assertIn(check_analysis.GIT_BATCH_MINIMUM,
                      (tool / "README.md").read_text(encoding="utf-8"))
        workflow = tool.parent.parent / "docs" / "WORKFLOW.md"
        if not workflow.is_file():
            # 사본 하네스(`analysis/harness/75_mut.py`)는 `tools/logic-map` 만 복사한다.
            # 없는 파일을 "통과" 로 세지 않고 **건너뛴 사실을 남긴다** —
            # [[universal-check-passes-on-an-empty-sample]].
            self.skipTest("docs/WORKFLOW.md is outside this checkout (copy harness)")
        self.assertIn(check_analysis.GIT_BATCH_MINIMUM,
                      workflow.read_text(encoding="utf-8"))

    def test_a_truncated_response_is_a_verdict_not_a_partial_answer(self) -> None:
        """응답이 요청보다 짧으면 **부분 답을 쓰지 않는다** (2026-09-18 독립 리뷰 P0).

        예전에는 여기서 `break` 하고 "나머지는 `None` 이라 답이 같다"고 적었다. **거짓이었다**:
        마지막 머리가 멀쩡하고 내용만 잘린 응답에서 그 자리는 `None` 이 아니라 `b""` 가 되고,
        `None` 과 `b""` 는 `_unheld_bundles` 의 아카이브 대체 갈래를 여닫아 **판정을 바꾼다**.
        변이 T6 이 살아남은 것은 동등해서가 아니라 **이 갈래에 닿는 시험이 없어서**였다
        ([[mutation-must-reach-the-thing-under-test]]) — 그때 근거로 쓴 T18 은 `missing`
        갈래를 증명한 것이라 명제가 달랐다.
        """
        raw, root, commit = self._repo()
        with raw:
            real = check_analysis.subprocess.run

            # 첫 레코드는 **git 이 낸 그대로** 둔다 — 크기를 속이면 엄격한 파서가 한 단계 먼저
            # "NUL 종단 없음" 으로 잡아서 이 시험이 겨냥한 `end < 0` 갈래에 **닿지 않는다**
            # (7.5.1 에서 실제로 그렇게 됐다, [[mutation-must-reach-the-thing-under-test]]).
            first = real(["git", "cat-file", "--batch", "-Z"], cwd=root,
                         input=f"{commit}:internal/own0.go\0".encode(),
                         capture_output=True, check=True).stdout

            def answers_one_of_two(*args: object, **kwargs: object):
                process = real(*args, **kwargs)
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "cat-file" in argv:
                    # 둘째 머리가 NUL 없이 끊겼다 — 부분 답을 쓰면 첫 자리만 채운 dict 가 나간다.
                    process.stdout = first + b"cafebabecafebabecafebabecafebabecafebabe blob 5"
                return process

            with mock.patch.object(check_analysis.subprocess, "run", answers_one_of_two):
                with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                    check_analysis._committed_many(
                        root, commit, ["internal/own0.go", "internal/own1.go"]
                    )
            self.assertIn("truncated", str(caught.exception))

    def test_bytes_left_unread_are_a_verdict(self) -> None:
        """정상 응답은 **정확히** 소진된다(실측). 남은 바이트는 프레이밍을 잘못 읽었다는
        뜻이고, 밀린 프레이밍은 **다른 파일의 바이트**를 답에 넣는다."""
        raw, root, commit = self._repo()
        with raw:
            real = check_analysis.subprocess.run

            def adds_a_tail(*args: object, **kwargs: object):
                process = real(*args, **kwargs)
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "cat-file" in argv:
                    process.stdout = process.stdout + b"leftover"
                return process

            with mock.patch.object(check_analysis.subprocess, "run", adds_a_tail):
                with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                    check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn("byte(s) unread", str(caught.exception))

    def test_a_nul_in_a_path_is_refused_before_it_desyncs_the_batch(self) -> None:
        """요청이 NUL 로 끊기므로 경로 안의 NUL 은 레코드를 쪼갠다 — 물어보지 않는다.

        개행은 `-Z` 가 견디지만 NUL 은 프레이밍 문자 자체다. 오늘 실물에서 못 닿는 이유는
        `Path.resolve()` 가 먼저 `ValueError` 를 내기 때문뿐이고, 그것은 이 함수의 불변식이
        아니다 ([[fail-closed-must-name-what-it-rejects]])."""
        raw, root, commit = self._repo()
        with raw:
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, ["internal/o\0wn.go"])
            self.assertIn("NUL byte", str(caught.exception))

    def test_a_path_that_is_not_utf8_still_reaches_git(self) -> None:
        """파일 이름은 바이트다. 옛 판본은 경로를 argv 로 넘겨 `os.fsencode` 를 탔으므로
        디코딩 불가능한 이름도 그대로 갔다 — 엄격한 `utf-8` 인코딩은 그 앞에서 **판정 대신
        traceback** 이 된다.

        내용을 **되받아서** 잰다. `None` 만 단언하면 "git 에 물었는데 그 커밋에 없다" 와
        "아예 못 물었다" 가 같게 보인다 — 존재 검사는 역할 검사가 아니다
        ([[existence-check-is-not-a-role-check]])."""
        raw, root, _ = self._repo()
        with raw:
            odd = "internal/own\udcff.go"      # surrogateescape 로만 표현되는 이름
            body = b"package internal\nfunc Odd() int { return 9 }\n"
            try:
                (root / odd).write_bytes(body)
            except (OSError, UnicodeEncodeError) as exc:
                # UTF-8 파일명을 강제하는 파일시스템(APFS 등)은 이 이름을 못 만든다.
                # 없는 표본 위에서 통과시키지 않고 **건너뛴 사실을 남긴다**
                # ([[universal-check-passes-on-an-empty-sample]]).
                self.skipTest(f"this filesystem refuses non-UTF-8 names: {exc}")
            commit = _commit_all(root, "P: a name that is not UTF-8")
            self.assertEqual(check_analysis._committed_many(root, commit, [odd])[odd], body)

    def test_nothing_asked_means_no_git_at_all(self) -> None:
        """물은 것이 없으면 프로세스를 안 띄운다."""
        raw, root, commit = self._repo()
        with raw:
            value, spawns = self._spawns(lambda: check_analysis._committed_many(root, commit, []))
            self.assertEqual((value, spawns), ({}, 0))

    def test_a_failure_that_still_printed_is_not_parsed(self) -> None:
        """rc≠0 인데 **읽을 만한 출력이 있는** 경우가 실패 갈래의 진짜 모양이다.

        저장소 아닌 곳을 물으면 git 은 빈 출력으로 죽고, 그때는 파서도 같은 답을 낸다 —
        그래서 rc 가드를 지워도 아무 시험이 안 빨개졌다(변이 T5 · T17 이 **서로를 덮었다**,
        [[surviving-mutant-may-mean-accidental-safety]]). 부분 출력을 남기고 죽는 git 은 그
        둘을 가른다: 파서에 맡기면 **진짜 blob 을 읽었다고 답한다.**
        """
        raw, root, commit = self._repo()
        with raw:
            honest = check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn(b"func Own0()", honest["internal/own0.go"])
            real = check_analysis.subprocess.run

            def dies_after_printing(*args: object, **kwargs: object):
                process = real(*args, **kwargs)
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "cat-file" in argv:
                    process.returncode = 128      # 출력은 그대로 두고 실패만 시킨다
                return process

            with mock.patch.object(check_analysis.subprocess, "run", dies_after_printing):
                with self.assertRaises(check_analysis.GATE_FAULTS):
                    check_analysis._committed_many(root, commit, ["internal/own0.go"])

    def test_an_archived_bundle_held_at_the_candidate_is_not_called_unheld(self) -> None:
        """아카이브된 번들은 **옮긴 뒤 경로**로 들고 있으면 들고 있는 것이다.

        옮기기 전 경로는 못 찾을 때의 **대안**이지 대체가 아니다. `and` 를 `or` 로 바꾸면
        (변이 T14) 아카이브된 change 는 옮긴 뒤 커밋에서 전부 미보유가 된다."""
        raw, root, _ = self._repo()
        with raw:
            archive = (root / "openspec" / "changes" / "archive"
                       / "2026-09-18-mine")
            archive.mkdir(parents=True)
            relative = "internal/own0.go"
            _write_evidence(
                archive, package="internal", function="Own0", relative=relative,
                digest=hashlib.sha256((root / relative).read_bytes()).hexdigest(),
            )
            commit = _commit_all(root, "X: evidence, already archived")
            analysis = archive / "analysis" / "function-logic"
            bundles = _bundles(root, analysis)
            ast_path = bundles[0][0].relative_to(root).as_posix()
            # 픽스처가 실제로 그 갈래에 닿는지 먼저 단언한다.
            self.assertTrue(check_analysis._pre_archive_path(ast_path), ast_path)
            self.assertEqual(check_analysis._unheld_bundles(
                root, commit, bundles, check_analysis._read_evidence(analysis).held), [])

    def test_a_newline_in_a_path_still_names_one_file(self) -> None:
        """경로에 개행이 있어도 **그 파일**이다. 배치 입력을 줄 단위로 나누면 한 경로가
        둘로 쪼개져 엉뚱한 blob 이나 `missing` 이 된다 — NUL 구분이 그래서 있다."""
        raw, root, _ = self._repo()
        with raw:
            odd = "internal/two\nlines.go"
            (root / odd).write_bytes(b"package internal\nfunc Odd() int { return 7 }\n")
            commit = _commit_all(root, "P: a newline in the name")
            fetched = check_analysis._committed_many(root, commit, [odd, "internal/own0.go"])
            self.assertEqual(fetched[odd], (root / odd).read_bytes())
            self.assertIn(b"func Own0()", fetched["internal/own0.go"])

    def test_judging_one_commit_costs_one_git_process(self) -> None:
        """`_pinning_at` 은 번들 수와 무관하게 프로세스 **하나**를 쓴다."""
        raw, root, commit = self._repo()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            for index in range(4):
                relative = f"internal/own{index}.go"
                _write_evidence(
                    change, package="internal", function=f"Own{index}", relative=relative,
                    digest=hashlib.sha256((root / relative).read_bytes()).hexdigest(),
                )
            analysis = change / "analysis" / "function-logic"
            (count, mismatched), spawns = self._spawns(
                lambda: check_analysis._pinning_at(root, commit, _bundles(root, analysis))
            )
            self.assertEqual((count, mismatched), (4, []))
            self.assertEqual(spawns, 1, "번들 넷을 한 프로세스로 읽어야 한다")

    def test_the_unheld_check_also_costs_one_git_process(self) -> None:
        """`_unheld_bundles` 도 하나다 — 받는 후보는 두 함수를 **다** 통과하므로
        (2026-09-18 실측: a112 의 성공하는 walk 에서 spawn 이 47.5% · 47.5% 로 반반),
        한쪽만 고치면 성공 경로의 비용이 절반만 준다."""
        raw, root, _ = self._repo()
        with raw:
            change = root / "openspec" / "changes" / "mine"
            change.mkdir(parents=True)
            for index in range(4):
                relative = f"internal/own{index}.go"
                _write_evidence(
                    change, package="internal", function=f"Own{index}", relative=relative,
                    digest=hashlib.sha256((root / relative).read_bytes()).hexdigest(),
                )
            commit = _commit_all(root, "X: evidence")
            analysis = change / "analysis" / "function-logic"
            held = check_analysis._read_evidence(analysis).held
            unheld, spawns = self._spawns(
                lambda: check_analysis._unheld_bundles(root, commit, _bundles(root, analysis), held)
            )
            self.assertEqual(unheld, [])
            self.assertEqual(spawns, 1, "번들 넷을 한 프로세스로 읽어야 한다")



class AFailedGitReadIsAFaultNotAnAbsence(unittest.TestCase):
    """git 이 **못 돌았다** 와 객체가 **없다** 는 다른 답이다 (task 7.5.1, gstack 리뷰 2026-09-19).

    `_committed_many` 는 rc≠0 에서 전부 `None` 을 돌려줬고 docstring 은 "부르는 쪽이 `None` 을
    불일치로 세므로 판정이 느슨해지지 않는다" 고 **전칭으로** 적었다. 두 출처의 적대 리뷰가
    그 전칭을 두 자리에서 깼다 — 가드 7 은 한쪽만 실패한 `None != bytes` 를 "소스가 바뀌었다" 로
    읽어 **편집 전 커밋을 착지로 기록했고**, `_landing_record` 는 `None` 을 "기록 없음" 으로 읽어
    착지 검증을 **통째로 건너뛰었다.** 결함은 `subprocess.run` 층에서 주입한다 —
    `_committed_many` 의 반환을 흉내 내면 수리 자체를 건너뛴다
    ([[mutation-must-reach-the-thing-under-test]]).
    """

    OLD_GIT = b"error: unknown switch `Z'\n"

    @staticmethod
    def _cat_file(behaviour):
        """`git cat-file` 호출만 `behaviour(argv, request)` 에 맡긴다. `None` 을 돌려주면 진짜로 돈다."""
        real = subprocess.run

        def run(*args, **kwargs):
            argv = args[0] if args else kwargs.get("args")
            if isinstance(argv, list) and "cat-file" in argv:
                forged = behaviour(argv, kwargs.get("input") or b"")
                if forged is not None:
                    return forged
            return real(*args, **kwargs)

        return mock.patch.object(check_analysis.subprocess, "run", run)

    def _fail_reads_at(self, ref: str | None):
        """`ref` 를 묻는 읽기만 `-Z` 를 모르는 git 처럼 죽인다. `None` 이면 전부."""
        def behaviour(argv, request):
            if ref is None or request.startswith(ref.encode()):
                return subprocess.CompletedProcess(argv, 129, b"", self.OLD_GIT)
            return None
        return self._cat_file(behaviour)

    def _answers(self, stdout: bytes):
        return self._cat_file(lambda argv, request: subprocess.CompletedProcess(argv, 0, stdout, b""))

    def test_a_one_sided_git_failure_cannot_land_a_pre_edit_commit(self) -> None:
        """**P0 — 서브에이전트 F1 을 시험으로 굳힌다.** V1 픽스처에서 base 쪽 읽기**만** 죽이면
        옛 판본은 rc 0 으로 편집 전 커밋 X 를 착지로 **기록했다**(재현). 그 창에는 이 change 의
        Go 작업이 없다 — 가드 7 이 막으려던 바로 그것이다."""
        case = ALandingMustChangeWhatItsEvidencePins(
            "test_the_recorder_refuses_evidence_written_before_the_edit")
        raw, root, marks = case._flm_first()
        with raw:
            record = root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE
            with self._fail_reads_at(marks["P"]):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse(record.exists(), "편집 전 커밋을 착지로 기록했다")
            # 거절 사유는 **git 이 한 말**이다 — 저자의 증거를 탓하지 않는다.
            self.assertTrue(any("unknown switch" in line for line in lines), lines)

    def test_a_committed_record_is_not_read_as_absent_when_git_fails(self) -> None:
        """**Codex P1-1.** 커밋된 착지 기록이 있는데 git 이 못 읽으면 옛 판본은 "기록 없음" 으로
        읽고 7.3.1 의 등식과 거절 가드 여덟을 **건너뛰었다** — `check` 가 `[]` 를 냈다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            self.assertEqual(check_analysis.check("mine", root), [])    # 대조군: 정상 git
            with self._fail_reads_at(None):
                errors = check_analysis.check("mine", root)
            self.assertTrue(any("unknown switch" in error for error in errors), errors)

    def test_rc_failure_names_what_git_said(self) -> None:
        """rc≠0 은 "없다" 가 아니라 결함이고 **git 의 말**로 이름을 댄다 (F6). 옛 판본은 stderr 를
        받아 놓고 버렸고, 문서 세 곳이 그 오진을 피해 가는 법을 설명하고 있었다."""
        with tempfile.TemporaryDirectory() as raw:
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(Path(raw), "HEAD", ["a.go"])
            self.assertIn("not a git repository", str(caught.exception))

    def test_an_unknown_header_is_a_verdict_not_an_absence(self) -> None:
        """**Codex P2.** 알 수 없는 머리를 "없는 파일" 로 치면 결함이 부재로 둔갑한다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with self._answers(b"nonsense\0"), self.assertRaises(check_analysis.GATE_FAULTS):
                check_analysis._committed_many(root, commit, ["internal/own0.go"])

    def test_a_missing_answer_must_echo_the_spec_that_was_asked(self) -> None:
        """`missing` 은 **물은 spec 그대로** 돌아온다(실측). 다른 spec 의 `missing` 은 프레이밍이
        밀렸다는 뜻이고, 그대로 받으면 한 자리의 부재가 다른 자리에 붙는다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with self._answers(f"{commit}:internal/other.go missing\0".encode()), \
                    self.assertRaises(check_analysis.GATE_FAULTS):
                check_analysis._committed_many(root, commit, ["internal/own0.go"])

    def test_a_missing_payload_terminator_is_a_verdict(self) -> None:
        """**Codex P2.** 내용 뒤의 NUL 을 안 보면 `abcX` 가 `abc` 로 받아들여진다 — 총량이 맞는다고
        프레이밍이 맞는 것은 아니다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            oid = b"0" * 40
            with self._answers(oid + b" blob 3\0abcX"), self.assertRaises(check_analysis.GATE_FAULTS):
                check_analysis._committed_many(root, commit, ["internal/own0.go"])

    def test_a_nul_in_the_ref_is_refused_by_name(self) -> None:
        """NUL 검사가 경로만 봤다 (F7). 요청 문자열 **전체**가 프레이밍 대상이다."""
        raw, root, _ = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, "HE\0AD", ["internal/own0.go"])
            self.assertIn("NUL byte", str(caught.exception))

    def test_a_truncated_response_counts_records_not_blobs(self) -> None:
        """잘림 문장은 **읽은 레코드**를 센다 (F4). `missing` 레코드도 레코드다 — 옛 문장은
        blob 만 세서 "0 of 3" 이라고 했다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            asked = ["a.go", "b.go", "c.go"]
            stdout = b"".join(f"{commit}:{name} missing\0".encode() for name in asked[:2])
            with self._answers(stdout), self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, asked)
            self.assertIn("2 of 3", str(caught.exception))

    def test_a_short_last_payload_is_reported_as_truncated(self) -> None:
        """마지막 내용이 짧으면 "잘렸다" 고 말한다 (F5). 옛 문장은 `left -6 byte(s) unread` —
        음수이고, 짧았는데 **남았다** 고 했다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with self._answers(b"0" * 40 + b" blob 10\0abc"), \
                    self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn("truncated", str(caught.exception))
            self.assertNotIn("left -", str(caught.exception))

    def test_every_fault_handler_uses_the_one_list(self) -> None:
        """손 복사 예외 목록 넷이 `SubprocessError` 를 빠뜨려 새 60초 타임아웃이 창 줄을 삼켰다
        (서브에이전트). 목록은 **한 곳**에 산다 — [[a-fault-must-become-a-verdict]]."""
        tree = ast.parse(Path(check_analysis.__file__).read_text(encoding="utf-8"))
        copied = [
            node.lineno for node in ast.walk(tree)
            if isinstance(node, ast.ExceptHandler) and node.type is not None
            and "RuntimeError" in ast.unparse(node.type)
        ]
        self.assertEqual(copied, [], f"손 복사 예외 목록이 남은 줄: {copied}")

    def test_a_git_timeout_during_the_landing_is_a_verdict(self) -> None:
        """타임아웃은 `check` 밖으로 새지 않고 **판정 줄**이 된다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            self.assertEqual(check_analysis.check("mine", root), [])    # 대조군: 멎지 않는 git

            def hangs(argv, request):
                raise subprocess.TimeoutExpired(cmd=argv, timeout=60)

            with self._cat_file(hangs):
                errors = check_analysis.check("mine", root)
            self.assertTrue(any("timed out" in error for error in errors), errors)


class TheLandingIsJudgedOnOneInventory(unittest.TestCase):
    """착지 판정은 **한 번 잰** 번들 목록 위에서 선다 (task 7.5.1 — 서브에이전트 F2 · Codex P1-2).

    7.5 는 "번들 목록을 한 번 잰다" 고 주석을 달았는데 `resolve_landing` 은 목록을 재고 나서
    `compute_landing` 이 **또** 쟀다 — 선언한 착지를 목록 #1 로 판정하고 목록 #2 로 계산한 값과
    같기를 요구했다. 그리고 7.5 가 목록을 걷기 내내 얼렸으므로, 걷는 도중에 생긴 번들은
    `_unheld_bundles` 가 **보지 못했다**(7.5 이전 판본은 후보마다 다시 읽어서 봤다 — 회귀다).
    """

    def test_the_declared_path_measures_once(self) -> None:
        """하한과 수리 신호가 착지 판정 한 번에 **한 번씩**만 계산된다(옛 판본: 둘 다 두 번)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            with mock.patch.object(check_analysis, "_evidence_floor",
                                   wraps=check_analysis._evidence_floor) as floor, \
                    mock.patch.object(check_analysis, "_self_repair_commits",
                                      wraps=check_analysis._self_repair_commits) as repairs:
                self.assertEqual(check_analysis.check("mine", root), [])
            self.assertEqual((floor.call_count, repairs.call_count), (1, 1))

    def test_a_bundle_published_during_the_walk_is_not_left_unchecked(self) -> None:
        """**Codex P1-2.** 걷는 도중 커밋 안 된 번들 B 가 생기면 옛 판본은 얼린 목록에 B 가 없어
        "그 커밋이 판정이 읽은 증거를 들고 있나" 를 B 에 대해 **묻지 않고** 착지를 기록했다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            change = root / "openspec" / "changes" / "mine"
            other = root / "internal" / "other.go"
            other.write_text("package internal\nfunc Other() int { return 1 }\n")
            _commit_all(root, "an unrelated source the late bundle will pin")
            real = subprocess.run

            def publishes_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv:
                    _write_evidence(
                        change, package="internal", function="Other",
                        relative="internal/other.go",
                        digest=hashlib.sha256(other.read_bytes()).hexdigest(),
                    )
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", publishes_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse((change / check_analysis.LANDING_FILE).exists(), lines)
            # 거절은 **재확인의 문장**이어야 한다 — 뒤의 다른 거절도 rc 1 이라 그것만으로는 재확인을
            # 지워도 초록이다 (재리뷰 레드팀: 지문 변이 여섯이 살아남았다).
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)


    def test_a_bundle_published_while_the_floor_is_measured_is_not_left_unchecked(self) -> None:
        """지문은 하한 **앞에서** 잰다. 하한을 재는 `git log` 도중에 번들 B 가 생기면, 판정 목록은
        이미 B 없이 읽혔는데 뒤에 잰 지문에는 B 가 있어서 수락 직전 재확인이 "그대로" 라고 답한다
        (변이 V10 이 **닿고도** 살아남았다 — 이 틈에 넣는 시험이 없었다)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            change = root / "openspec" / "changes" / "mine"
            other = root / "internal" / "other.go"
            other.write_text("package internal\nfunc Other() int { return 1 }\n")
            _commit_all(root, "an unrelated source the late bundle will pin")
            real = subprocess.run

            import traceback

            def publishes_while_measuring_the_floor(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                # **착지 입력을 재는 도중**의 하한 `git log` 에서만 쓴다. `record_landing` 은 걷기 전에
                # `_recording_refusal` 로 하한을 한 번 먼저 재므로, 첫 `git log` 에 쓰면 두 판본이
                # 다 B 를 보고 이 시험이 순서를 **못 가른다** — 첫 판이 그래서 V10 을 못 잡았다
                # ([[mutation-must-reach-the-thing-under-test]]).
                measuring = any(frame.name == "_measure_landing_inputs"
                                for frame in traceback.extract_stack())
                if isinstance(argv, list) and "--diff-filter=MAT" in argv and measuring:
                    _write_evidence(
                        change, package="internal", function="Other",
                        relative="internal/other.go",
                        digest=hashlib.sha256(other.read_bytes()).hexdigest(),
                    )
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run",
                                   publishes_while_measuring_the_floor):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse((change / check_analysis.LANDING_FILE).exists(), lines)
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)


class AVerdictReadsPinnedSourcesInOneProcess(unittest.TestCase):
    """착지가 있는 판정은 고정 소스를 착지에서 **한 번** 읽는다 (task 7.5.1 — 서브에이전트 F3).

    7.5 는 "근본 원인은 fetch 단위였고 고쳤다" 고 적었는데 파일에서 가장 큰 남은 per-bundle
    루프 둘을 안 바꿨다 — `validate_target`(**매 게이트 실행의 본 판정 경로**, a112 147 프로세스)과
    가드 7(후보마다 2×소스 수). 그래서 번들 수와 무관해야 할 비용이 번들 수에 비례했다.
    """

    @staticmethod
    def _landed(raw: tempfile.TemporaryDirectory, sources: int) -> Path:
        """`_own_work_fixture` 와 같은 모양(P → W → E)에 고정 소스를 `sources` 개로 늘리고 기록까지 한다."""
        root = _init_fixture(raw)
        files = [root / "internal" / f"own{index}.go" for index in range(sources)]
        files[0].parent.mkdir(parents=True)
        for index, path in enumerate(files):
            path.write_text(f"package internal\nfunc Own{index}() int {{ return 1 }}\n")
        base = _commit_all(root, "P: base")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(base + "\n")
        (change / "review.md").write_text("mine\n")
        for index, path in enumerate(files):
            path.write_text(f"package internal\nfunc Own{index}() int {{ return 2 }}\n")
        _commit_all(root, "W: this change's Go work lands")
        for index, path in enumerate(files):
            _write_evidence(
                change, package="internal", function=f"Own{index}",
                relative=path.relative_to(root).as_posix(),
                digest=hashlib.sha256(path.read_bytes()).hexdigest(),
            )
        _commit_all(root, "E: its evidence enters the history")
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        _commit_all(root, "record the landing")
        return root

    @staticmethod
    def _cat_file_spawns(root: Path) -> tuple[list[str], int]:
        seen = []
        real = subprocess.run

        def counting(*args, **kwargs):
            argv = args[0] if args else kwargs.get("args")
            if isinstance(argv, list) and "cat-file" in argv:
                seen.append(argv)
            return real(*args, **kwargs)

        with mock.patch.object(check_analysis.subprocess, "run", counting):
            errors = check_analysis.check("mine", root)
        return errors, len(seen)

    def test_guard_seven_reads_each_side_once_even_when_it_refuses(self) -> None:
        """가드 7 은 **거절할 때** 모든 소스를 읽어야 한다(전부 base 와 같아서). 소스별로 읽는 판본은
        `all()` 이 첫 번째로 다른 소스에서 멈추므로 **받는** 경우엔 싸 보인다 — 그래서 변이 V13 이
        받는 픽스처에서 닿고도 살아남았다. V1 모양(FLM 먼저 커밋, 번들 안 갱신)은 가드 7 이 거절하는
        모양이고, 거기서 소스 1개와 4개의 blob 읽기 프로세스 수가 **같아야** 한다."""
        spawns = {}
        for sources in (1, 4):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root = _init_fixture(raw)
                files = [root / "internal" / f"own{index}.go" for index in range(sources)]
                files[0].parent.mkdir(parents=True)
                for index, path in enumerate(files):
                    path.write_text(f"package internal\nfunc Own{index}() int {{ return 1 }}\n")
                base = _commit_all(root, "P: base")
                change = root / "openspec" / "changes" / "mine"
                change.mkdir(parents=True)
                (change / "base-commit.txt").write_text(base + "\n")
                (change / "review.md").write_text("mine\n")
                for index, path in enumerate(files):        # 증거가 **base** 를 적는다 (V1)
                    _write_evidence(
                        change, package="internal", function=f"Own{index}",
                        relative=path.relative_to(root).as_posix(),
                        digest=hashlib.sha256(path.read_bytes()).hexdigest(),
                    )
                _commit_all(root, "X: FLM first, as the repository asks")
                for index, path in enumerate(files):
                    path.write_text(f"package internal\nfunc Own{index}() int {{ return 2 }}\n")
                _commit_all(root, "W: the edit, bundles not refreshed")
                seen = []
                real = subprocess.run

                def counting(*args, **kwargs):
                    argv = args[0] if args else kwargs.get("args")
                    if isinstance(argv, list) and "cat-file" in argv:
                        seen.append(argv)
                    return real(*args, **kwargs)

                with mock.patch.object(check_analysis.subprocess, "run", counting):
                    code, lines = check_analysis.record_landing("mine", root)
                self.assertEqual(code, 1, lines)            # 가드 7 이 X 를 거절한다
                self.assertTrue(any("changes none of the sources" in line for line in lines), lines)
                spawns[sources] = len(seen)
        self.assertEqual(spawns[1], spawns[4], spawns)

    def test_the_cost_does_not_grow_with_the_number_of_pinned_sources(self) -> None:
        """고정 소스 1개와 4개에서 blob 읽기 프로세스 수가 **같다**. 옛 판본은 가드 7(2×N)과
        `validate_target`(N) 때문에 번들 수에 비례했다."""
        spawns = {}
        for sources in (1, 4):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root = self._landed(raw, sources)
                errors, spawns[sources] = self._cat_file_spawns(root)
                self.assertEqual(errors, [], (sources, errors))
        self.assertEqual(spawns[1], spawns[4], spawns)


class TheVerdictReadsWhatTheLandingJudged(unittest.TestCase):
    """착지 판정과 그 뒤의 판정이 **같은 바이트**를 읽는다 (task 7.5.2 — 수리한 트리의 재리뷰).

    7.5.1 의 지문은 판정 목록과 **따로** 읽혔다. 한 `ast.json` 읽기가 잠깐 실패하면 그 번들이
    판정에서 빠진 채 두 지문이 같다고 답했다(Codex 재현). 그리고 수락 뒤 `check()` 가 `ast.json` 을
    **다시** 읽어서 그 사이 바뀐 증거로 판정했다(적대 서브에이전트 재현 — a099 에서 2.48s 창).
    결함은 **읽기 자체**(`Path.read_bytes`/`read_text`)나 `subprocess.run` 층에서 주입한다 —
    지문 함수의 반환을 흉내 내면 수리를 건너뛴다 ([[mutation-must-reach-the-thing-under-test]]).
    """

    MOVED = "changed while the landing was being judged"

    @staticmethod
    def _stack_has(*names: str) -> bool:
        frames = {frame.name for frame in traceback.extract_stack()}
        return all(name in frames for name in names)

    @staticmethod
    def _two_bundles(raw: tempfile.TemporaryDirectory) -> Path:
        """`_own_work_fixture` 에 고정 번들 하나(Other)를 더한다. 정상이면 E2 가 착지다."""
        root, _ = _own_work_fixture(raw)
        change = root / "openspec" / "changes" / "mine"
        other = root / "internal" / "other.go"
        other.write_text("package internal\nfunc Other() int { return 1 }\n")
        _write_evidence(
            change, package="internal", function="Other", relative="internal/other.go",
            digest=hashlib.sha256(other.read_bytes()).hexdigest(),
        )
        _commit_all(root, "E2: a second pinned bundle")
        return root

    @staticmethod
    def _reads(on_ast_json):
        """`ast.json` 을 읽는 **모든** 길(`read_bytes` · `read_text`)을 `on_ast_json(path)` 에 먼저 보인다."""
        real_bytes, real_text = Path.read_bytes, Path.read_text

        def read_bytes(path):
            if path.name == "ast.json":
                on_ast_json(path)
            return real_bytes(path)

        def read_text(path, *args, **kwargs):
            if path.name == "ast.json":
                on_ast_json(path)
            return real_text(path, *args, **kwargs)

        return mock.patch.multiple(Path, read_bytes=read_bytes, read_text=read_text)

    @staticmethod
    def _landed(raw: tempfile.TemporaryDirectory) -> tuple[Path, Path]:
        root, _ = _own_work_fixture(raw)
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        _commit_all(root, "record the landing")
        ast_path = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own" / "ast.json"
        return root, ast_path

    def test_the_landing_inputs_read_each_bundle_once(self) -> None:
        """**Codex P1 의 근본.** 착지 입력과 판정이 **한 번 읽은 바이트**에서 나와야 한다. 7.5.1 은
        번들마다 세 번 읽었다(지문의 목록 · 지문의 바이트 · 하한의 목록). 7.5.2.1 부터는 **명령 전체**에서
        한 번이다 — 수락 직전 재확인의 읽기는 대조만 하므로 따로 센다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = self._two_bundles(raw)
            reads: dict[str, int] = {}

            def count(path):
                if not self._stack_has("_raise_if_inputs_moved"):
                    reads[path.parent.name] = reads.get(path.parent.name, 0) + 1

            with self._reads(count):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            self.assertEqual(reads, {"internal--own": 1, "internal--other": 1})

    def test_a_bundle_whose_read_fails_once_is_never_judged_absent(self) -> None:
        """**Codex P1.** 기록하는 동안 번들 B 의 읽기 **하나**가 실패한다 — 몇 번째 읽기든.
        7.5.1 은 셋째 읽기(하한의 목록)에서 실패하면 B 를 빼고 판정했는데 두 지문은 B 를 담아 같았다
        → 착지를 **기록했다**. 이제 읽기는 둘뿐이다(판정 한 번 · 수락 직전 대조 한 번) — 어느 쪽이
        실패해도 두 읽기가 달라서 "움직였다" 다."""
        fired_any = False
        for nth in (1, 2, 3):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root = self._two_bundles(raw)
                record = root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE
                seen = {"count": 0, "fired": False}

                def flaky(path, nth=nth, seen=seen):
                    if path.parent.name == "internal--other" and self._stack_has("record_landing"):
                        seen["count"] += 1
                        if seen["count"] == nth:
                            seen["fired"] = True
                            raise PermissionError(13, "a flaky read", str(path))

                with self._reads(flaky):
                    code, lines = check_analysis.record_landing("mine", root)
                if not seen["fired"]:
                    continue                    # 이 판본에는 n 번째 읽기가 없다
                fired_any = True
                self.assertEqual(code, 1, (nth, lines))
                self.assertFalse(record.exists(), (nth, lines))
                self.assertTrue(any(self.MOVED in line for line in lines), (nth, lines))
        self.assertTrue(fired_any, "주입이 한 번도 닿지 않았다 — 계측기가 눈멀었다")

    def test_a_bundle_rewritten_during_the_walk_is_seen(self) -> None:
        """개수 · `file` · 해시가 그대로인 **재작성**도 움직임이다 (재리뷰 레드팀 — 지문에서 바이트를
        빼는 변이가 추가만 재던 시험을 통과했다)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            ast_path = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own" / "ast.json"
            real = subprocess.run

            def rewrites_at_guard_seven(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                request = kwargs.get("input") or b""
                if isinstance(argv, list) and "cat-file" in argv and isinstance(request, bytes) \
                        and request.startswith(marks["P"].encode()) and self._stack_has("compute_landing"):
                    value = json.loads(ast_path.read_text(encoding="utf-8"))
                    value["note"] = "rewritten mid-walk"
                    ast_path.write_text(json.dumps(value), encoding="utf-8")
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", rewrites_at_guard_seven):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertFalse((ast_path.parents[3] / check_analysis.LANDING_FILE).exists(), lines)
            self.assertTrue(any(self.MOVED in line for line in lines), lines)

    def test_a_declared_landing_refused_on_moving_evidence_says_run_again(self) -> None:
        """착지 입력을 **잴 때** 저자가 `ast.json` 을 쓰던 중이었으면(잰 것은 중간 상태 B'), 선언 경로의
        거절은 B' 에서 나온다. 그때 복구 조언(기록을 지우고 다시 기록)을 하면 다시 돌리면 통과할
        기록을 지우라고 권하게 된다(재리뷰 적대 F5 재현). 거절 앞에서 입력이 그대로인지 본다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, ast_path = self._landed(raw)
            original = ast_path.read_bytes()
            value = json.loads(original)
            value["note"] = "half-written"
            partial = json.dumps(value).encode("utf-8")
            real = subprocess.run

            def mid_edit_while_measuring(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list):
                    # 7.5.2.1 부터 증거는 `check` 가 역사를 푼 **뒤** 한 번 읽는다 — 그 앞뒤에 넣는다.
                    if argv == ["git", "rev-parse", "--verify", "HEAD^{commit}"]:
                        ast_path.write_bytes(partial)           # 증거를 읽기 직전: 쓰던 중
                    elif "--diff-filter=MAT" in argv and self._stack_has("_measure_landing_inputs"):
                        ast_path.write_bytes(original)          # 읽은 뒤: 저장을 마쳤다
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", mid_edit_while_measuring):
                errors = check_analysis.check("mine", root)
            self.assertTrue(any(self.MOVED in error for error in errors), errors)
            self.assertFalse(any("to move a record" in error for error in errors), errors)
            self.assertEqual(check_analysis.check("mine", root), [])   # 다시 돌리면 통과한다

    def test_evidence_swapped_after_the_verdict_read_it_is_not_read_again(self) -> None:
        """판정이 한 번 읽은 **뒤**에 갈아 끼운 증거는 판정에 안 들어온다 — 판정은 그 한 번의 바이트로
        선다. 옛 판본은 대상마다 디스크를 다시 읽어서 그 사이 통과하는 증거로 바꾸면 `[]` 였다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            ast_path = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own" / "ast.json"
            passing = ast_path.read_bytes()
            value = json.loads(passing)
            value["branches"] = [
                {"id": "B1", "kind": "if", "line": 2, "column": 1},
                {"id": "B2", "kind": "if", "line": 2, "column": 1},
            ]
            ast_path.write_text(json.dumps(value), encoding="utf-8")
            _commit_all(root, "E': evidence whose second branch the map does not cover")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            real_index = check_analysis.test_index

            def swaps_after_the_read(*args, **kwargs):
                ast_path.write_bytes(passing)
                return real_index(*args, **kwargs)

            with mock.patch.object(check_analysis, "test_index", swaps_after_the_read):
                errors = check_analysis.check("mine", root)
            self.assertTrue(any("missing AST branches" in error for error in errors), errors)

    def test_a_bundle_that_appears_after_the_verdict_read_is_missing(self) -> None:
        """한 번 읽을 때 없던 `ast.json` 은 **없는 것**이다. 대상 디렉터리는 있었는데 증거가 그 뒤에
        생기면, 옛 판본은 디스크에서 읽어 착지가 판정하지 않은 번들로 함수를 덮었다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, ast_path = self._landed(raw)
            late = ast_path.parent.parent / "internal--late"
            shutil.copytree(ast_path.parent, late)
            late_ast = (late / "ast.json").read_bytes()
            (late / "ast.json").unlink()
            real_index = check_analysis.test_index

            def publishes_after_the_read(*args, **kwargs):
                (late / "ast.json").write_bytes(late_ast)
                return real_index(*args, **kwargs)

            with mock.patch.object(check_analysis, "test_index", publishes_after_the_read):
                errors = check_analysis.check("mine", root)
            self.assertIn("internal--late: missing ast.json", errors)

    def test_evidence_swapped_and_restored_during_the_walk_is_not_judged(self) -> None:
        """바뀌었다 **돌아오는** 증거(ABA). 워킹트리의 `ast.json` 은 커밋 안 된 재작성 B 인데, 후보마다
        "그 커밋이 판정이 읽은 증거를 드는가" 를 볼 때만 커밋된 C 로 바꿨다가 곧 B 로 되돌린다.
        디스크를 후보마다 다시 읽던 판본은 그 순간의 C 로 "든다" 고 판정했고, 수락 직전 재확인은
        돌아온 B 를 보고 "그대로" 라고 답해 `[]` 를 냈다 — 착지가 B 를 들지 않는데. 판정은 잰 **그**
        바이트로 서야 한다 (task 7.5.2)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, ast_path = self._landed(raw)
            committed = ast_path.read_bytes()
            value = json.loads(committed)
            value["note"] = "an uncommitted rewrite"
            rewritten = json.dumps(value).encode("utf-8")
            ast_path.write_bytes(rewritten)
            honest = check_analysis.check("mine", root)
            self.assertTrue(any("does not hold the evidence" in error for error in honest), honest)  # 대조군
            real = subprocess.run

            def swaps_only_while_holding_is_judged(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "cat-file" in argv:
                    ast_path.write_bytes(committed if self._stack_has("_unheld_bundles") else rewritten)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", swaps_only_while_holding_is_judged):
                errors = check_analysis.check("mine", root)
            ast_path.write_bytes(rewritten)
            self.assertTrue(errors, "착지가 들지 않은 증거로 판정이 통과했다")
            self.assertTrue(
                any(self.MOVED in error or "does not hold the evidence" in error for error in errors), errors)

    def test_every_evidence_read_goes_through_the_one_reader(self) -> None:
        """판정의 어느 자리도 `ast.json` 을 디스크에서 **따로** 읽지 않는다 — 전부 `_read_evidence` 를
        거친다. 행동 시험은 자리마다 그 자리의 창에 주입해야 잡히는데 자리가 여럿이다(대상 판정 ·
        열거형 호출 판정 · 번들 글자 · "드는가" 판정). 그래서 불변식 자체를 잰다."""
        for landed in (True, False):
            raw = tempfile.TemporaryDirectory()
            with raw:
                if landed:
                    root, _ = self._landed(raw)
                else:
                    root, _ = _own_work_fixture(raw)
                elsewhere: list[str] = []

                def count(path):
                    if not self._stack_has("_read_evidence"):
                        elsewhere.append(path.parent.name)

                with self._reads(count):
                    errors = check_analysis.check("mine", root)
                self.assertEqual(errors, [], landed)
                self.assertEqual(elsewhere, [], landed)

    def test_bytes_read_once_decode_exactly_as_the_disk_read_did(self) -> None:
        """한 번 읽은 바이트를 글자로 바꾸는 규칙이 `Path.read_text` 와 **같아야** 한다 — 줄바꿈 변환까지.
        다르면 같은 증거가 디스크에서 읽을 때와 다른 글자가 된다."""
        with tempfile.TemporaryDirectory() as raw:
            path = Path(raw) / "ast.json"
            for body in (b'{"a": 1}\r\n', b"x\ry\r\nz\n", "\ud55c\uae00\r\n".encode("utf-8"), b""):
                path.write_bytes(body)
                self.assertEqual(check_analysis._decoded(body), path.read_text(encoding="utf-8"), body)
            path.write_bytes(b"\xff")
            with self.assertRaises(UnicodeDecodeError):
                path.read_text(encoding="utf-8")
            with self.assertRaises(UnicodeDecodeError):
                check_analysis._decoded(b"\xff")

    def test_evidence_that_could_not_be_read_is_named_not_skipped(self) -> None:
        """한 번 읽기에서 **못 읽은** `ast.json` 은 "없다" 도 아니고 판정에서 조용히 빠지지도 않는다 —
        그 대상의 판정 줄이 된다. 옛 판본은 대상 판정에서 다시 읽다가 `PermissionError` 로 판정 전체를
        결함 한 줄로 바꿨다(변이 W23 이 이 갈래에 닿는 시험이 없어서 살아남았다)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)

            def unreadable(path):
                if path.parent.name == "internal--own" and self._stack_has("_read_evidence"):
                    raise PermissionError(13, "not readable", str(path))

            with self._reads(unreadable):
                errors = check_analysis.check("mine", root)
            self.assertIn("internal--own: ast.json could not be read", errors)
            self.assertNotIn("internal--own: missing ast.json", errors)

    def test_a_head_that_cannot_be_read_is_a_fault_not_an_empty_head(self) -> None:
        """지문의 `HEAD` 를 못 읽으면 결함이다 — 빈 값이면 두 번 다 못 읽은 실행이 "그대로" 가 된다
        (7.5.1 이 `None` 으로 배운 것과 같다)."""
        with tempfile.TemporaryDirectory() as raw:
            subprocess.run(["git", "init", "-q"], cwd=raw, check=True)      # 커밋이 하나도 없다
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._head_commit(Path(raw))
            self.assertIn("cannot read HEAD", str(caught.exception))

    def test_the_evidence_is_read_again_only_to_accept(self) -> None:
        """재확인은 **받을 때만** 선다(레드팀 · testing 전문가 MC). 후보마다 다시 재면 7.5 가 걷어낸
        번들 순회가 돌아온다 — 프로세스를 안 띄우므로 spawn 수 시험은 못 본다. 거절하는 기록은 읽기 한 번
        (판정), 받는 기록은 두 번(판정 · 수락 직전 대조)이다 (task 7.5.2.1)."""
        case = ALandingMustChangeWhatItsEvidencePins(
            "test_the_recorder_refuses_evidence_written_before_the_edit")
        raw, root, _ = case._flm_first()
        with raw, mock.patch.object(check_analysis, "_read_evidence",
                                    wraps=check_analysis._read_evidence) as taken:
            code, lines = check_analysis.record_landing("mine", root)
        self.assertEqual((code, taken.call_count), (1, 1), lines)
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            with mock.patch.object(check_analysis, "_read_evidence",
                                   wraps=check_analysis._read_evidence) as taken:
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual((code, taken.call_count), (0, 2), lines)

    def test_an_adopted_changes_escaping_bundle_keeps_its_name(self) -> None:
        """**이관 change 회귀** (Codex P2 · 적대 F3). 착지 해소를 안 거치는 이관 경로에서 미리 읽기가
        번들을 **처음** 선별하다가 저장소 밖 소스에 멈춰, 대상 이름과 나머지 대상의 오류가 다 사라졌다."""
        case = CheckAnalysisTests("test_valid_adoption_uses_e_and_requires_complete_current_bundle")
        raw, root, p, e = case._adoption_with_complete_bundle()
        with raw, mock.patch.object(adoption, "P", p), mock.patch.object(adoption, "E", e):
            self.assertEqual(check_analysis.check(adoption.CHANGE, root), [])     # 대조군
            good = root / "openspec" / "changes" / adoption.CHANGE / "analysis" / "function-logic" / "internal-soak--attest"
            bad = good.parent / "bad-bundle"
            shutil.copytree(good, bad)
            value = json.loads((bad / "ast.json").read_text(encoding="utf-8"))
            value["file"] = "../outside/source.go"
            (bad / "ast.json").write_text(json.dumps(value), encoding="utf-8")
            case._commit(root, "an escaping bundle")
            subprocess.run(["git", "checkout", "--detach", "-q"], cwd=root, check=True)
            errors = check_analysis.check(adoption.CHANGE, root)
        self.assertIn("bad-bundle: AST source escapes repository", errors)

    def test_advice_that_cannot_be_computed_does_not_replace_the_verdict(self) -> None:
        """조언은 판정이 아니다 (적대 F6). 창이 워킹트리일 때 "번들이 base 를 적는가" 를 재다가 git 이
        죽으면 옛 판본은 판정 전체를 `cannot judge this change` 한 줄로 바꿨다 — 판정 자체는 git 없이
        디스크에서 서는데도."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            (root / "internal" / "own.go").write_text("package internal\nfunc Own() int { return 3 }\n")

            def fails_at_the_base(argv, request):
                if request.startswith(marks["P"].encode()):
                    return subprocess.CompletedProcess(argv, 129, b"", AFailedGitReadIsAFaultNotAnAbsence.OLD_GIT)
                return None

            with AFailedGitReadIsAFaultNotAnAbsence._cat_file(fails_at_the_base):
                code, output = _cli(root)
            self.assertEqual(code, 1, output)
            self.assertIn("internal--own: AST source hash is stale", output)
            self.assertIn("cannot tell whether", output)
            self.assertIn("unknown switch", output)
            self.assertNotIn("cannot judge this change", output)

    def test_the_base_shape_advice_reads_the_base_once(self) -> None:
        """조언의 base 읽기도 번들 수와 무관하게 한 프로세스다 (적대 F6 — 7.5 가 남긴 per-bundle 루프)."""
        spawns = {}
        for sources in (1, 3):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root = _init_fixture(raw)
                files = [root / "internal" / f"own{index}.go" for index in range(sources)]
                files[0].parent.mkdir(parents=True)
                for index, path in enumerate(files):
                    path.write_text(f"package internal\nfunc Own{index}() int {{ return 1 }}\n")
                base = _commit_all(root, "P: base")
                change = root / "openspec" / "changes" / "mine"
                change.mkdir(parents=True)
                (change / "base-commit.txt").write_text(base + "\n")
                (change / "review.md").write_text("mine\n")
                for index, path in enumerate(files):
                    path.write_text(f"package internal\nfunc Own{index}() int {{ return 2 }}\n")
                    _write_evidence(
                        change, package="internal", function=f"Own{index}",
                        relative=path.relative_to(root).as_posix(),
                        digest=hashlib.sha256(path.read_bytes()).hexdigest(),
                    )
                _commit_all(root, "W: work and evidence")
                for index, path in enumerate(files):        # 커밋 안 한 편집 — 번들이 낡았다
                    path.write_text(f"package internal\nfunc Own{index}() int {{ return 3 }}\n")
                seen = []
                real = subprocess.run

                def counting(*args, **kwargs):
                    argv = args[0] if args else kwargs.get("args")
                    request = kwargs.get("input") or b""
                    if isinstance(argv, list) and "cat-file" in argv and request.startswith(base.encode()):
                        seen.append(argv)
                    return real(*args, **kwargs)

                with mock.patch.object(check_analysis.subprocess, "run", counting):
                    check_analysis.check("mine", root)
                spawns[sources] = len(seen)
        self.assertEqual(spawns, {1: 1, 3: 1})

    def test_the_git_version_hint_is_given_only_for_a_usage_error(self) -> None:
        """"`-Z` 는 git 2.42 이상" 은 git 이 **옵션을 모른다**(rc 129)고 할 때만 맞는 말이다. 저장소가
        아닌 루트(rc 128)에도 붙이면 git 의 말 옆에 추측한 진단이 다시 붙는다 (재리뷰 · F6 의 취지)."""
        with tempfile.TemporaryDirectory() as raw:
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(Path(raw), "HEAD", ["a.go"])
        self.assertIn("not a git repository", str(caught.exception))
        self.assertNotIn("needs git", str(caught.exception))
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with AFailedGitReadIsAFaultNotAnAbsence()._fail_reads_at(None), \
                    self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn(f"needs git {check_analysis.GIT_BATCH_MINIMUM}", str(caught.exception))

    def test_a_submodule_answer_is_not_a_blob(self) -> None:
        """새 git 은 커밋이 없는 gitlink 를 `<oid> submodule` 로 답한다(영수증: `builtin/cat-file.c` 의
        `report_object_status(opt, NULL, &data->oid, "submodule")`). blob 이 아니므로 `None` 이다 —
        결함이 아니다 (Codex P2)."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with AFailedGitReadIsAFaultNotAnAbsence()._answers(b"a" * 40 + b" submodule\0"):
                self.assertEqual(check_analysis._committed_many(root, commit, ["vendor/sub"]),
                                 {"vendor/sub": None})

    def test_a_framing_fault_names_the_path_it_was_reading(self) -> None:
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            for stdout in (b"nonsense\0", b"0" * 40 + b" blob 3\0abcX"):
                with AFailedGitReadIsAFaultNotAnAbsence()._answers(stdout), \
                        self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                    check_analysis._committed_many(root, commit, ["internal/own0.go"])
                self.assertIn("internal/own0.go", str(caught.exception), stdout)

    def test_a_payload_missing_only_its_terminator_is_truncated(self) -> None:
        """내용은 다 왔고 **끝의 NUL 만** 없다 — 경계 그 자체. `>=` 를 `>` 로 바꾸면 `IndexError` 가
        `GATE_FAULTS` 밖으로 새어 판정 대신 traceback 이 된다 (레드팀 M20 · testing 전문가)."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with AFailedGitReadIsAFaultNotAnAbsence()._answers(b"0" * 40 + b" blob 3\0abc"), \
                    self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn("truncated", str(caught.exception))

    def test_a_source_missing_from_the_prefetch_is_read_at_the_landing(self) -> None:
        """미리 읽기에 없는 경로는 **착지에서** 읽는다(`HEAD` 가 아니다). 이 갈래를 부르는 시험이
        없어서 fallback 의 ref 를 바꾸는 변이가 살았다 (testing 전문가 MB · MB2)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            (root / "internal" / "own.go").write_text("package internal\nfunc Own() int { return 3 }\n")
            _commit_all(root, "N: HEAD now differs from the landing")
            target = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own"
            held = check_analysis._read_evidence(target.parent).held
            at_landing, _ = check_analysis.validate_target(target, root, None, False, marks["E"], {}, held=held)
            at_head, _ = check_analysis.validate_target(target, root, None, False, "HEAD", {}, held=held)
            self.assertFalse(any("AST source" in error for error in at_landing), at_landing)
            self.assertTrue(any("AST source hash is stale" in error for error in at_head), at_head)

    def test_a_fault_reading_the_landing_for_the_verdict_is_a_verdict(self) -> None:
        """착지 해소 **뒤**의 blob 읽기(미리 읽기)가 죽어도 판정 줄이 된다 — 창 줄과 git 의 말과 함께."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = self._landed(raw)
            landing = (root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE).read_text().strip()

            def fails_for_the_verdict(argv, request):
                if request.startswith(landing.encode()) and self._stack_has("check") \
                        and not self._stack_has("resolve_landing"):
                    return subprocess.CompletedProcess(argv, 129, b"", AFailedGitReadIsAFaultNotAnAbsence.OLD_GIT)
                return None

            with AFailedGitReadIsAFaultNotAnAbsence._cat_file(fails_for_the_verdict):
                code, output = _cli(root)
            self.assertEqual(code, 1, output)
            self.assertIn("cannot judge this change", output)
            self.assertIn("unknown switch", output)
            self.assertIn(f"landed-commit {landing}", output)

    def test_a_malformed_ast_is_a_verdict_not_a_traceback(self) -> None:
        """`start`·`end` 가 사전이 아니거나 `branches` 가 목록이 아닌 `ast.json` 은 판정 줄이 된다.
        옛 판본은 `AttributeError`·`TypeError` 로 `main()` 을 빠져나갔다(레드팀 — 7.5 이전부터).
        저장소 전수 3,048 번들 중 이 모양은 0 이다(review.md `## MEASURE · Pre-Edit Gate — task 7.5.2`)."""
        for key, bad in (("start", 5), ("end", "2"), ("branches", 5)):
            raw = tempfile.TemporaryDirectory()
            with raw:
                root, _ = _own_work_fixture(raw)
                ast_path = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own" / "ast.json"
                value = json.loads(ast_path.read_text(encoding="utf-8"))
                value[key] = bad
                ast_path.write_text(json.dumps(value), encoding="utf-8")
                errors = check_analysis.check("mine", root)
                self.assertIn("internal--own: ast.json is invalid", errors, key)

    def test_repairs_that_were_never_measured_are_not_read_as_none(self) -> None:
        """하한이 못 서면 수리 신호를 **안 잰다**. 그것을 `[]` 로 두면 "수리 커밋이 없다" 와 같은 말이
        된다(`_self_repair_commits` 가 빈 목록을 거절한 이유와 같다 — 재리뷰 maintainability).
        오늘은 하한 가드가 그 앞에서 거절해서 안 닿는다 — 그 **사슬**을 여기서 못 박는다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            with self.assertRaises(check_analysis.GATE_FAULTS):
                check_analysis._repairs_after(root, marks["E"], None, _head(root))
            empty = root / "openspec" / "changes" / "mine" / "analysis" / "nothing-here"
            inputs = check_analysis._measure_landing_inputs(
                root, _head(root), check_analysis._read_evidence(empty))
            self.assertTrue(inputs.why)
            self.assertIsNone(inputs.repairs)
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            refusal, _ = check_analysis._landing_refusal(
                root, marks["P"], marks["E"], _inputs(root, analysis, floor="", repairs=None))
            self.assertIn("no ordinary commit on this history adds", refusal)


class OneCommandJudgesOneHistoryAndOneRead(unittest.TestCase):
    """명령 하나는 역사 **하나**와 증거 읽기 **하나** 위에서 판정한다 (task 7.5.2.1 — 7.5.2 재리뷰).

    7.5.2 는 `ast.json` 바이트만 한 번 읽어 묶고 "묶었다" 고 적었다. 판정이 읽는 입력은 그것만이
    아니었다 — review.md `## MEASURE · Pre-Edit Gate — task 7.5.2.1` 의 목록(`7521_inputs.py` 가 AST 로
    셌다): 역사(`HEAD`, 열 자리에서 따로) · 증거 디렉터리의 존재 · 번들 목록 · 번들 안의 파일 목록 ·
    git 의 결함. 재리뷰가 그중 셋을 재현했다. 여기 시험은 입력마다 "움직여도 판정이 한 벌의 것인가" 를
    잰다. 결함은 `subprocess.run` · 파일시스템 층에서 넣는다 — 판정 함수를 흉내 내면 수리를 건너뛴다.
    """

    PIN = ["git", "rev-parse", "--verify", "HEAD^{commit}"]

    @staticmethod
    def _git_seen(call):
        """`call` 동안의 git 호출 전부 — `(argv, stdin 바이트)`."""
        seen = []
        real = subprocess.run

        def watching(*args, **kwargs):
            argv = args[0] if args else kwargs.get("args")
            request = kwargs.get("input") or b""
            seen.append((list(argv), request.encode() if isinstance(request, str) else request))
            return real(*args, **kwargs)

        with mock.patch.object(check_analysis.subprocess, "run", watching):
            result = call()
        return result, seen

    @staticmethod
    def _two_branches(raw: tempfile.TemporaryDirectory) -> tuple[Path, dict[str, str]]:
        """레드팀 `repro_b` 의 모양. 기본 가지: P → W → E → R(기록) → F(기록 뒤의 자기 수리 — 빨강).
        `other`: E → N(이웃 편집, 기록 없음 — 워킹트리 창이 이웃 함수까지 요구해서 빨강). **둘 다 각자로는
        빨갛다** — 섞였을 때만 초록이 나온다."""
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        neighbour = root / "internal" / "neighbour.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\nfunc Own() int { return 1 }\n")
        neighbour.write_text("package internal\nfunc Neighbour() int { return 1 }\n")
        marks = {"P": _commit_all(root, "P: base")}
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(marks["P"] + "\n")
        (change / "review.md").write_text("mine\n")
        own.write_text("package internal\nfunc Own() int { return 2 }\n")
        marks["W"] = _commit_all(root, "W: this change's Go work")
        _write_evidence(change, package="internal", function="Own", relative="internal/own.go",
                        digest=hashlib.sha256(own.read_bytes()).hexdigest())
        marks["E"] = _commit_all(root, "E: its evidence")
        home = subprocess.check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=root, text=True).strip()
        subprocess.run(["git", "checkout", "-q", "-b", "other"], cwd=root, check=True)
        neighbour.write_text("package internal\nfunc Neighbour() int { return 2 }\n")
        marks["N"] = _commit_all(root, "N: a neighbour edit, no record on this branch")
        subprocess.run(["git", "checkout", "-q", home], cwd=root, check=True)
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        marks["record"] = (change / check_analysis.LANDING_FILE).read_text(encoding="utf-8").strip()
        marks["R"] = _commit_all(root, "R: the record")
        own.write_text("package internal\nfunc Own() int { return 3 }\n")
        (change / "review.md").write_text("mine: review repair\n")
        marks["F"] = _commit_all(root, "F: this change's own repair after the record")
        return root, marks

    # --- 역사: 한 명령은 `HEAD` 를 **한 번** 푼다 ---

    def test_every_history_read_uses_the_one_resolved_commit(self) -> None:
        """**구조로** 잰다: 한 명령 안에서 상징 `HEAD` 를 읽는 git 호출은 그것을 sha 로 푸는 **한 번**
        뿐이다. 7.5.2 는 열 자리에서 따로 읽었다(착지 기록 · 조상 판정 · 하한 · 수리 신호 · 순회 · 수리 뒤 ·
        지문 표본 · 깨끗한가 · 창 줄). 표본을 하나 더 두는 것으로는 못 닫는다 — 표본 **전**의 읽기와
        떠났다 돌아온 `HEAD` 가 남는다(아래 두 시험)."""
        landed = tempfile.TemporaryDirectory()
        fresh = tempfile.TemporaryDirectory()
        with landed, fresh:
            recorded, _ = TheVerdictReadsWhatTheLandingJudged._landed(landed)
            working, _ = _own_work_fixture(fresh)
            for name, call in (
                ("check, landed", lambda: check_analysis.check("mine", recorded)),
                ("main, working tree + advice", lambda: _cli(working)),
                ("record", lambda: check_analysis.record_landing("mine", working)),
            ):
                result, seen = self._git_seen(call)
                reading_head = [argv for argv, request in seen
                                if any("HEAD" in word for word in argv) or b"HEAD" in request]
                self.assertEqual(reading_head, [self.PIN], (name, result))

    def test_a_branch_switch_mid_run_judges_the_history_it_started_on(self) -> None:
        """**레드팀 repro_b.** 병행 세션이 착지를 판정하는 도중 가지를 **한 번** 바꾼다(돌아오지 않는다).
        7.5.2 는 기록을 전환 **전** `HEAD` 에서 읽고, 지문 표본과 수리 신호 · 순회는 전환 **뒤** `HEAD` 에서
        읽어 rc 0 이었다 — 전환 전만 봐도, 후만 봐도 빨간데. 판정은 시작할 때의 역사 하나의 것이어야 한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = self._two_branches(raw)
            alone = check_analysis.check("mine", root)
            self.assertTrue(any("of this change's own Go work" in error for error in alone), alone)  # 대조군
            real = subprocess.run
            switched = []

            def switches_once(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                # 기록을 읽은 **뒤**, 옛 판본이 지문 표본을 뜨기 **전** — 선언값이 커밋인지 묻는 자리
                if isinstance(argv, list) and argv[:3] == ["git", "rev-parse", "--verify"] \
                        and argv[3] == marks["record"] + "^{commit}" and not switched:
                    switched.append(True)
                    real(["git", "checkout", "-q", "other"], cwd=root, check=True)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", switches_once):
                mixed = check_analysis.check("mine", root)
            self.assertTrue(switched, "주입이 닿지 않았다")
            self.assertEqual(mixed, alone)

    def test_a_head_that_leaves_and_returns_mid_run_is_never_read_twice(self) -> None:
        """**떠났다 돌아온 `HEAD`(ABA)** — 이 세션이 재현했다. 기록 뒤 자기 수리(F)가 선 역사에서, "기록
        뒤에 수리가 있나" 를 묻는 순간에만 `HEAD` 가 기록 커밋으로 물러났다가 곧 F 로 돌아온다. 7.5.2 는
        그 순간의 `HEAD` 로 "없다" 고 판정했고, 수락 직전 표본 비교는 돌아온 F 를 보고 "그대로" 라고 답해
        rc 0 이었다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _self_repair_fixture(raw, repair="pinned")
            alone = check_analysis.check("mine", root)
            self.assertTrue(any("of this change's own Go work" in error for error in alone), alone)  # 대조군
            real = subprocess.run
            moved = []

            def steps_back_while_asked(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and argv[:2] == ["git", "rev-list"] \
                        and "--reverse" not in argv and "--count" not in argv:
                    moved.append(True)
                    real(["git", "reset", "-q", "--soft", marks["recorded"]], cwd=root, check=True)
                    try:
                        return real(*args, **kwargs)
                    finally:
                        real(["git", "reset", "-q", "--soft", marks["F"]], cwd=root, check=True)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", steps_back_while_asked):
                errors = check_analysis.check("mine", root)
            self.assertTrue(moved, "주입이 닿지 않았다")
            self.assertEqual(errors, alone)

    def test_a_commit_landing_mid_judgment_does_not_change_the_verdict(self) -> None:
        """역사를 고정하면 실행 중의 커밋은 판정을 **멈추지도 바꾸지도** 않는다. 7.5.2 는 그것을
        `INPUTS_MOVED` 로 멈췄다 — 이 작업트리를 병행 세션이 같이 쓰므로 무관한 커밋 하나에 게이트가 멎었다.
        지키려던 것(두 역사를 섞은 판정)은 고정이 없앴다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = TheVerdictReadsWhatTheLandingJudged._landed(raw)
            real = subprocess.run

            def commits_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv:
                    (root / "notes.txt").write_text("a neighbour's commit\n")
                    real(["git", "add", "notes.txt"], cwd=root, check=True)
                    real(["git", "commit", "-qm", "a neighbour lands"], cwd=root, check=True)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", commits_mid_walk):
                errors = check_analysis.check("mine", root)
            self.assertEqual(errors, [])

    def test_a_commit_landing_mid_record_is_recorded_on_the_history_it_started_on(self) -> None:
        """기록 명령도 같다 — 시작할 때 푼 역사 위에서 계산한 값을 쓴다. 그 뒤 역사가 바뀌면 게이트가
        기록을 **다시 판정한다**(착지 뒤의 자기 수리면 거절한다) — 기록 명령이 미리 멈출 이유가 아니다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, marks = _own_work_fixture(raw)
            real = subprocess.run

            def commits_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv:
                    (root / "notes.txt").write_text("a neighbour's commit\n")
                    real(["git", "add", "notes.txt"], cwd=root, check=True)
                    real(["git", "commit", "-qm", "a neighbour lands"], cwd=root, check=True)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", commits_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            record = root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE
            self.assertEqual(record.read_text(encoding="utf-8").strip(), marks["E"])

    # --- git 의 결함은 판정이 아니다 ---

    def test_a_git_fault_is_never_advice_to_delete_the_record(self) -> None:
        """**레드팀 repro_a · repro_a3.** 7.5.1 은 `_committed_many` 의 결함만 결함으로 올렸다. 순회 ·
        하한 · 조상 판정의 rc 128 은 여전히 거절 사유가 되어 "기록을 지우고 다시 기록하라" 는 복구 조언이
        붙었다 — 다시 돌리면 통과할 **유효한 기록**을 지우라는 말이다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = TheVerdictReadsWhatTheLandingJudged._landed(raw)
            base = (root / "openspec" / "changes" / "mine" / "base-commit.txt").read_text().strip()
            self.assertEqual(check_analysis.check("mine", root), [])                  # 대조군
            real = subprocess.run
            for name, hit in (
                ("rev-list --reverse", lambda argv: "--reverse" in argv),
                ("floor git log", lambda argv: "--diff-filter=MAT" in argv),
                ("merge-base base", lambda argv: argv[:3] == ["git", "merge-base", "--is-ancestor"]
                 and argv[3] == base),
            ):
                def transient(*args, hit=hit, **kwargs):
                    argv = args[0] if args else kwargs.get("args")
                    if isinstance(argv, list) and hit(argv):
                        text = kwargs.get("text")
                        return subprocess.CompletedProcess(
                            argv, 128, "" if text else b"",
                            "fatal: transient\n" if text else b"fatal: transient\n")
                    return real(*args, **kwargs)

                with mock.patch.object(check_analysis.subprocess, "run", transient):
                    errors = check_analysis.check("mine", root)
                self.assertEqual(len(errors), 1, (name, errors))
                self.assertTrue(errors[0].startswith("cannot derive modified Go functions:"), (name, errors))
                self.assertIn("transient", errors[0], name)
                self.assertNotIn("to move a record", errors[0], name)

    def test_a_cleanliness_check_git_cannot_answer_is_not_called_dirty(self) -> None:
        """`git diff --quiet` 는 0(같다) · 1(다르다) 말고는 **답이 아니다**. 옛 판본은 rc 128 도 "추적
        파일이 바뀌었다" 로 읽어 "먼저 커밋하라" 고 했다 — 커밋할 것이 없는 저자에게."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            real = subprocess.run

            def transient(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and argv[:3] == ["git", "diff", "--quiet"]:
                    return subprocess.CompletedProcess(argv, 128, b"", b"fatal: transient\n")
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", transient):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any("transient" in line for line in lines), lines)
            self.assertFalse(any("uncommitted changes" in line for line in lines), lines)

    # --- 모양 검사는 파싱하는 자리 한 곳에 ---

    def test_every_malformed_ast_shape_is_a_verdict_not_a_traceback(self) -> None:
        """재리뷰가 셌다: 27 모양 중 **열**이 7.5.2 뒤에도 traceback 이었다(`calls`/`returns` 가 수 ·
        참거짓 · 최상위가 목록 · 글자 · 수 · 참거짓). 모양은 대상 판정이 아니라 **바이트를 값으로 푸는 자리**
        에서 봐야 한다 — 열거형 호출 판정은 대상 판정보다 **먼저** 같은 값을 쓴다. 규칙은 하나다: 사전 ·
        `start`/`end` 는 사전 · `branches`/`calls`/`returns` 는 목록(없거나 비어도 된다). 저장소 3,048 번들 0 건."""
        invalid = [("calls", 5), ("calls", 5.5), ("calls", True), ("returns", 5), ("returns", 5.5),
                   ("returns", True), ("calls", {"at": 1}), ("returns", "2:3"),
                   ("start", 5), ("end", "2"), ("branches", 5)]
        tops = [[{"file": "internal/own.go"}], "x", 5, True, [], None]
        for case in invalid + [("<top>", top) for top in tops]:
            raw = tempfile.TemporaryDirectory()
            with raw:
                root, _ = _own_work_fixture(raw)
                ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                            / "internal--own" / "ast.json")
                key, bad = case
                value = json.loads(ast_path.read_text(encoding="utf-8"))
                if key == "<top>":
                    value = bad
                else:
                    value[key] = bad
                ast_path.write_text(json.dumps(value), encoding="utf-8")
                errors = check_analysis.check("mine", root)
                expected = ("internal--own: ast.json is placeholder evidence" if key == "<top>"
                            else "internal--own: ast.json is invalid")
                self.assertIn(expected, errors, case)

    # --- 증거: 존재 · 목록 · 바이트를 한 번 ---

    @staticmethod
    def _landed_requiring_nothing(raw: tempfile.TemporaryDirectory) -> Path:
        """창이 함수를 **하나도** 요구하지 않는 착지(고정 파일의 주석만 바뀐다) + 면제 표지 + 지도에 없는
        분기 B2 를 가진 증거. 정직한 판정은 B2 로 빨갛다."""
        root = _init_fixture(raw)
        own = root / "internal" / "own.go"
        own.parent.mkdir(parents=True)
        own.write_text("package internal\n\n// rev 1\n\nfunc Own() int { return 1 }\n")
        base = _commit_all(root, "P: base")
        change = root / "openspec" / "changes" / "mine"
        change.mkdir(parents=True)
        (change / "base-commit.txt").write_text(base + "\n")
        (change / "review.md").write_text(f"mine\n{check_analysis.EXEMPTION}\n")
        own.write_text("package internal\n\n// rev 2\n\nfunc Own() int { return 1 }\n")
        _commit_all(root, "W: the file changes, no existing function does")
        bundle = _write_evidence(change, package="internal", function="Own", relative="internal/own.go",
                                 digest=hashlib.sha256(own.read_bytes()).hexdigest())
        value = json.loads((bundle / "ast.json").read_text(encoding="utf-8"))
        value["branches"] = [{"id": "B1", "kind": "if", "line": 5, "column": 1},
                             {"id": "B2", "kind": "if", "line": 5, "column": 1}]
        (bundle / "ast.json").write_text(json.dumps(value), encoding="utf-8")
        _commit_all(root, "E: evidence whose second branch the map does not cover")
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        _commit_all(root, "record the landing")
        return root

    def test_evidence_deleted_after_the_landing_is_judged_is_not_read_as_no_evidence(self) -> None:
        """**재리뷰(적대).** 증거 디렉터리의 **존재**도 입력이다. 7.5.2 는 착지를 판정한 뒤
        `analysis.exists()` 를 다시 물었고 그 조기 반환이 바이트 대조 **앞**에 있었다 — 착지가 판정한 뒤
        디렉터리를 지우면 요구 0 + 면제 표지로 `[]` 였다. 판정은 한 번 읽은 것(디렉터리가 **있었다**)으로 선다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = self._landed_requiring_nothing(raw)
            honest = check_analysis.check("mine", root)
            self.assertTrue(any("missing AST branches" in error for error in honest), honest)  # 대조군
            analysis = root / "openspec" / "changes" / "mine" / "analysis"
            real_changed = check_analysis.changed_existing_functions

            def deletes_after_acceptance(*args, **kwargs):
                shutil.rmtree(analysis)
                return real_changed(*args, **kwargs)

            with mock.patch.object(check_analysis, "changed_existing_functions", deletes_after_acceptance):
                errors = check_analysis.check("mine", root)
            self.assertNotEqual(errors, [])
            self.assertFalse(any(error.startswith("missing analysis") for error in errors), errors)

    def test_a_bundle_listing_that_misses_ast_json_keeps_the_bytes_judged(self) -> None:
        """**Codex P1 재현.** 열거형 호출 판정은 번들 안의 읽히는 파일을 **전부** 잇는데, 그 목록을 살아
        있는 `glob("*")` 로 만들었다 — 한 번 읽은 `ast.json` 이 목록에서 빠지면(한 번 읽은 뒤 지워짐) 그
        바이트도 빠졌다. 여기 `ast.json` 은 문자열 안에 U+2028 로 끊은 호출 좌표 표를 담아 열거를 켠다
        (`str.splitlines` 는 U+2028 에서도 끊는다). 정직한 판정은 "호출 표로 열어야 한다" 로 빨갛다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            value = json.loads(ast_path.read_text(encoding="utf-8"))
            value["calls"] = [{"at": {"line": 2, "column": 30}, "text": "f"}]
            value["note"] = "x | Callee | Position | | f | 2:30 | y"
            ast_path.write_text(json.dumps(value, ensure_ascii=False), encoding="utf-8")
            needle = "must open with a '| Callee expression | Position |' table"
            honest = check_analysis.check("mine", root)
            self.assertTrue(any(needle in error for error in honest), honest)          # 대조군
            real_index = check_analysis.test_index

            def deletes_after_the_read(*args, **kwargs):
                ast_path.unlink()
                return real_index(*args, **kwargs)

            with mock.patch.object(check_analysis, "test_index", deletes_after_the_read):
                errors = check_analysis.check("mine", root)
            self.assertTrue(any(needle in error for error in errors), errors)

    def test_a_bundle_directory_that_can_be_entered_but_not_listed_is_named(self) -> None:
        """**Codex P2.** 검색(x)은 되고 목록(r)은 안 되는 번들 디렉터리. 7.5.2 의 `glob("*/ast.json")` 은
        그 디렉터리를 조용히 건너뛰어 멀쩡히 읽히는 `ast.json` 을 "없다" 고 했다. 번들은 목록에서 세고
        `ast.json` 은 **직접** 연다. 산문을 다 이어야 하는 열거형 호출 판정은 목록 없이는 못 서므로 그것을
        **이름으로** 말한다 — 조용히 빈 글자로 판정하면 그 번들의 표가 판정에서 빠진다(permissive)."""
        if hasattr(os, "geteuid") and os.geteuid() == 0:
            self.skipTest("root 는 권한을 무시한다")
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            bundle = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--own"
            bundle.chmod(0o311)
            try:
                errors = check_analysis.check("mine", root)
            finally:
                bundle.chmod(0o755)
            self.assertNotIn("internal--own: missing ast.json", errors)
            self.assertTrue(any(error.startswith("internal--own: cannot list the bundle directory")
                                for error in errors), errors)

    def test_a_symlink_loop_in_a_bundle_source_is_named(self) -> None:
        """3.12 의 `Path.resolve()` 는 심링크 고리에서 `RuntimeError` 를 낸다(3.13 부터는 안 낸다). 대상
        판정은 `ValueError` 만 이름을 붙여 받아서, 고리는 판정 전체를 한 줄(또는 traceback)로 바꿨다 (재리뷰
        보안 전문가). 판정은 판본과 무관하게 **그 대상의 줄**이어야 한다 — 고리에는 읽을 소스가 없다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            os.symlink("loop.go", root / "internal" / "loop.go")
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            value = json.loads(ast_path.read_text(encoding="utf-8"))
            value["file"] = "internal/loop.go"
            ast_path.write_text(json.dumps(value), encoding="utf-8")
            errors = check_analysis.check("mine", root)
            self.assertIn("internal--own: AST source is missing: internal/loop.go", errors)

    def test_a_surrogate_in_a_path_prints_as_an_escape_not_a_traceback(self) -> None:
        """**레드팀.** JSON 의 `\\udcff` 는 합법이고 판정 줄은 그 경로를 이름으로 댄다. 출력이 엄격한
        UTF-8 이면(`en_US.UTF-8` 의 기본) `print` 가 `UnicodeEncodeError` 로 **판정 줄 없이** 끝났다 — 결과가
        로캘의 함수였다. `PYTHONIOENCODING` 으로 엄격함을 고정해 로캘과 무관하게 잰다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = TheVerdictReadsWhatTheLandingJudged._landed(raw)
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            value = json.loads(ast_path.read_text(encoding="utf-8"))
            value["file"] = "internal/\udcff.go"
            ast_path.write_text(json.dumps(value), encoding="utf-8")
            _commit_all(root, "the bundle names a source that is not UTF-8")
            process = subprocess.run(
                [sys.executable, str(Path(check_analysis.__file__).resolve()),
                 "--change", "mine", "--root", str(root)],
                capture_output=True, env={**os.environ, "PYTHONIOENCODING": "utf-8:strict"},
            )
            self.assertEqual(process.returncode, 1, process.stderr)
            self.assertNotIn(b"Traceback", process.stderr)
            self.assertIn(b"internal/\\udcff.go", process.stdout)

    def test_one_command_reads_the_evidence_once_to_judge(self) -> None:
        """중심 주장을 **직접** 잰다: 판정에 쓰는 증거 읽기는 명령마다 한 번이다(수락 · 거절 앞 재확인은
        **대조만** 한다 — 그 읽기의 바이트는 판정에 안 들어간다). 7.5.2 는 착지가 한 번, `check` 가 또 한 번
        읽고 둘을 대조했다 — 대조를 지우거나 둘째 읽기를 더하는 변이가 살았다(X1~X4)."""
        landed = tempfile.TemporaryDirectory()
        fresh = tempfile.TemporaryDirectory()
        with landed, fresh:
            recorded, _ = TheVerdictReadsWhatTheLandingJudged._landed(landed)
            working, _ = _own_work_fixture(fresh)
            for name, call in (
                ("check, landed", lambda: check_analysis.check("mine", recorded)),
                ("check, working tree", lambda: check_analysis.check("mine", working)),
                ("record", lambda: check_analysis.record_landing("mine", working)),
            ):
                judged = []
                real = check_analysis._read_evidence

                def counting(*args, **kwargs):
                    if not TheVerdictReadsWhatTheLandingJudged._stack_has("_raise_if_inputs_moved"):
                        judged.append(True)
                    return real(*args, **kwargs)

                with mock.patch.object(check_analysis, "_read_evidence", counting):
                    call()
                self.assertEqual(len(judged), 1, name)

    def test_evidence_swapped_after_the_landing_is_judged_is_not_what_the_verdict_reads(self) -> None:
        """수락 **뒤** 통과하는 증거로 갈아 끼워도 판정은 착지가 판정한 바이트(지도에 없는 B2)로 선다 —
        정직한 판정과 **같다**. 7.5.2 는 다시 읽어 대조하고 "바뀌었다" 로 멈췄다: 창은 없었지만 판정이 둘째
        읽기의 함수였다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            passing = ast_path.read_bytes()
            value = json.loads(passing)
            value["branches"] = [{"id": "B1", "kind": "if", "line": 2, "column": 1},
                                 {"id": "B2", "kind": "if", "line": 2, "column": 1}]
            ast_path.write_text(json.dumps(value), encoding="utf-8")
            _commit_all(root, "E': evidence whose second branch the map does not cover")
            code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 0, lines)
            _commit_all(root, "record the landing")
            honest = check_analysis.check("mine", root)
            self.assertTrue(any("missing AST branches" in error for error in honest), honest)  # 대조군
            real_changed = check_analysis.changed_existing_functions

            def swaps_after_acceptance(*args, **kwargs):
                ast_path.write_bytes(passing)
                return real_changed(*args, **kwargs)

            with mock.patch.object(check_analysis, "changed_existing_functions", swaps_after_acceptance):
                errors = check_analysis.check("mine", root)
            self.assertEqual(errors, honest)

    def test_a_bundle_removed_during_the_walk_is_seen(self) -> None:
        """재확인은 **빠진** 번들도 본다(추가만 재던 시험 옆의 나머지 절반 — 재리뷰 testing 전문가 X2)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = TheVerdictReadsWhatTheLandingJudged._two_bundles(raw)
            other = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--other"
            real = subprocess.run

            def removes_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv and other.exists():
                    shutil.rmtree(other)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", removes_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)

    def test_a_bundle_directory_added_during_the_walk_is_seen(self) -> None:
        """번들 **목록**도 한 번 읽은 것의 일부다. `ast.json` 이 아직 없는 번들 디렉터리가 걷는 도중 생기면
        판정의 대상 목록이 바뀐다 — 7.5.2 의 `glob("*/ast.json")` 지문은 그것을 못 봤다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            late = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic" / "internal--late"
            real = subprocess.run

            def adds_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv:
                    late.mkdir(exist_ok=True)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", adds_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)

    def test_a_bundle_rewritten_to_escape_mid_walk_is_a_move(self) -> None:
        """**재리뷰 레드팀.** 걷는 도중 번들이 저장소 **밖**을 가리키게 다시 쓰이면 그것은 움직임이다.
        7.5.2 의 재확인은 새 읽기에서 고정 목록부터 골라서, "움직였다" 대신 `escapes repository` 결함을
        냈다 — 다시 돌리라는 말 대신 저자의 증거를 탓했다. 바이트를 **먼저** 대조한다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            real = subprocess.run

            def escapes_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv:
                    value = json.loads(ast_path.read_text(encoding="utf-8"))
                    value["file"] = "../outside/own.go"
                    ast_path.write_text(json.dumps(value), encoding="utf-8")
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", escapes_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)
            self.assertFalse(any("escapes repository" in line for line in lines), lines)

    def test_a_source_path_that_starts_escaping_mid_walk_is_a_move(self) -> None:
        """바이트는 그대로인데 소스 경로의 **심링크**가 저장소 밖으로 바뀌면, 재확인의 고르기가
        `ValueError` 를 낸다. 그것도 움직임이다 — 잴 때는 골라졌던 것이 지금 안 골라진다."""
        raw = tempfile.TemporaryDirectory()
        outside = tempfile.TemporaryDirectory()
        with raw, outside:
            root, _ = _own_work_fixture(raw)
            (Path(outside.name) / "own.go").write_text("package internal\nfunc Own() int { return 2 }\n")
            link = root / "link"
            os.symlink("internal", link)
            ast_path = (root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
                        / "internal--own" / "ast.json")
            value = json.loads(ast_path.read_text(encoding="utf-8"))
            value["file"] = "link/own.go"               # 풀면 `internal/own.go` — 고정은 그대로 선다
            ast_path.write_text(json.dumps(value), encoding="utf-8")
            _commit_all(root, "E': the bundle names its source through a symlinked directory")
            self.assertEqual(check_analysis.record_landing("mine", root)[0], 0)       # 대조군
            (root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE).unlink()
            real = subprocess.run

            def retargets_mid_walk(*args, **kwargs):
                argv = args[0] if args else kwargs.get("args")
                if isinstance(argv, list) and "--reverse" in argv and link.resolve().parent == root:
                    link.unlink()
                    os.symlink(outside.name, link)
                return real(*args, **kwargs)

            with mock.patch.object(check_analysis.subprocess, "run", retargets_mid_walk):
                code, lines = check_analysis.record_landing("mine", root)
            self.assertEqual(code, 1, lines)
            self.assertTrue(any(TheVerdictReadsWhatTheLandingJudged.MOVED in line for line in lines), lines)

    def test_the_pins_are_compared_even_when_the_bytes_hold(self) -> None:
        """바이트가 그대로여도 고정 목록(소스 경로의 심링크 풀이)이 바뀌면 움직임이다 (재리뷰: 고정
        목록만 움직이는 시험이 없었다)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            head = check_analysis._head_commit(root)
            analysis = root / "openspec" / "changes" / "mine" / "analysis" / "function-logic"
            inputs = check_analysis._measure_landing_inputs(root, head, check_analysis._read_evidence(analysis))
            check_analysis._raise_if_inputs_moved(root, inputs)                      # 대조군: 그대로
            with self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._raise_if_inputs_moved(root, inputs._replace(bundles=[]))
            self.assertIn(TheVerdictReadsWhatTheLandingJudged.MOVED, str(caught.exception))

    def test_absent_and_unreadable_evidence_are_different_reads(self) -> None:
        """없는 `ast.json`(키 없음)과 못 읽는 `ast.json`(`None`)은 다른 읽기다 — 판정 줄도 다르고
        (`missing` / `could not be read`) 재확인에서도 다르다."""
        if hasattr(os, "geteuid") and os.geteuid() == 0:
            self.skipTest("root 는 권한을 무시한다")
        with tempfile.TemporaryDirectory() as raw:
            analysis = Path(raw)
            (analysis / "absent").mkdir()
            (analysis / "unreadable").mkdir()
            (analysis / "unreadable" / "ast.json").write_text("{}")
            (analysis / "unreadable" / "ast.json").chmod(0)
            try:
                evidence = check_analysis._read_evidence(analysis)
            finally:
                (analysis / "unreadable" / "ast.json").chmod(0o644)
            self.assertEqual(evidence.targets, (analysis / "absent", analysis / "unreadable"))
            self.assertNotIn(analysis / "absent" / "ast.json", evidence.held)
            self.assertIsNone(evidence.held[analysis / "unreadable" / "ast.json"])

    def test_a_walk_with_no_pinning_bundle_says_why(self) -> None:
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            self.assertEqual(
                check_analysis._walk_floor(root, [], check_analysis._head_commit(root)),
                ("", "no `revision: current` evidence pins a landing for this change"))

    def test_a_submodule_header_must_be_the_whole_answer(self) -> None:
        """`<oid> submodule` 은 **전체가** 그 모양일 때만 blob 아님이다 — 앞에 무엇이 붙으면 프레이밍이
        밀린 것이다(`fullmatch` 를 `search` 로 바꾸는 변이가 살았다)."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            with AFailedGitReadIsAFaultNotAnAbsence()._answers(b"x" + b"a" * 40 + b" submodule\0"), \
                    self.assertRaises(check_analysis.GATE_FAULTS):
                check_analysis._committed_many(root, commit, ["vendor/sub"])

    def test_the_version_hint_is_not_given_for_an_ordinary_failure(self) -> None:
        """rc 1 은 git 의 사용법 오류(129)가 아니다 — 버전 조언을 붙이지 않는다."""
        raw, root, commit = ABlobIsFetchedOncePerCommitNotOncePerBundle()._repo()
        with raw:
            def fails(argv, request):
                return subprocess.CompletedProcess(argv, 1, b"", b"error: something else\n")

            with AFailedGitReadIsAFaultNotAnAbsence._cat_file(fails), \
                    self.assertRaises(check_analysis.GATE_FAULTS) as caught:
                check_analysis._committed_many(root, commit, ["internal/own0.go"])
            self.assertIn("something else", str(caught.exception))
            self.assertNotIn("needs git", str(caught.exception))

    def test_a_hung_git_is_asked_once_for_the_verdict(self) -> None:
        """착지 뒤 판정의 미리 읽기에서 git 이 멎으면 **한 번** 묻고 판정 줄이 된다 — 대상마다 다시
        물으면 멎은 git 을 대상 수만큼 기다린다."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root = AVerdictReadsPinnedSourcesInOneProcess._landed(raw, 3)
            landing = (root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE).read_text().strip()
            asked = []

            def hangs_for_the_verdict(argv, request):
                if request.startswith(landing.encode()) and \
                        not TheVerdictReadsWhatTheLandingJudged._stack_has("resolve_landing"):
                    asked.append(True)
                    raise subprocess.TimeoutExpired(argv, 60)
                return None

            with AFailedGitReadIsAFaultNotAnAbsence._cat_file(hangs_for_the_verdict):
                code, output = _cli(root)
            self.assertEqual((code, len(asked)), (1, 1), output)
            self.assertIn("timed out", output)

    def test_a_landed_change_judges_a_base_bundle_without_pinning_it(self) -> None:
        """`revision: base` 번들은 착지를 고정하지 않는다 — 고정 목록에 넣으면 그 번들의 digest(base 의
        소스)가 착지에서 안 맞아 정직한 기록이 거절된다. 착지가 있는 change 에서 그런 번들이 판정되는
        경로를 부르는 시험이 없었다(재리뷰 testing 전문가)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = TheVerdictReadsWhatTheLandingJudged._landed(raw)
            base = (root / "openspec" / "changes" / "mine" / "base-commit.txt").read_text().strip()
            base_source = subprocess.run(["git", "show", f"{base}:internal/own.go"],
                                         cwd=root, capture_output=True, check=True).stdout
            bundle = _write_evidence(
                root / "openspec" / "changes" / "mine", package="internal", function="Base",
                relative="internal/own.go", digest=hashlib.sha256(base_source).hexdigest())
            value = json.loads((bundle / "ast.json").read_text(encoding="utf-8"))
            value["revision"] = "base"
            (bundle / "ast.json").write_text(json.dumps(value), encoding="utf-8")
            self.assertEqual(check_analysis.check("mine", root), [])

    def test_a_reused_context_carries_nothing_from_the_last_run(self) -> None:
        """`check` 가 채우는 사실은 **전부** 이번 실행의 것이어야 한다. 7.5.2 는 둘만 지웠다 — 앞선 이관
        실행의 `adoption_source` 나 조언의 번들 이름이 다음 실행의 창 줄로 새었다(재리뷰: pop 절반)."""
        raw = tempfile.TemporaryDirectory()
        with raw:
            root, _ = _own_work_fixture(raw)
            facts: dict[str, object] = {
                "adoption_source": "stale", "landing": "stale", "required_count": 99,
                "base_shaped_bundles": ["stale"], "base_shaped_fault": "stale", "head": "stale",
                "execution_baseline_adoption": True, "effective_base": "stale",
            }
            check_analysis.check("mine", root, facts)
            self.assertNotIn("adoption_source", facts)
            self.assertNotIn("base_shaped_fault", facts)
            self.assertEqual(facts["landing"], "")
            self.assertEqual(facts["base_shaped_bundles"], [])
            self.assertFalse(facts["execution_baseline_adoption"])
            self.assertEqual(facts["head"], check_analysis._head_commit(root))


if __name__ == "__main__":
    unittest.main()
