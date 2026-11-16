#!/usr/bin/env python3
"""완료 게이트 11단계(openspec validate)의 판정을 잰다 (a190).

왜 이게 필요한가
----------------
11단계는 `make validate` = `openspec validate --all --strict` 였다. 2026-11-16 openspec 1.16.0 에서
그 명령은 **남의** change 와 main spec 의 strict 경고 51 항목 때문에 rc 1 이었고, 닫으려는 change 가
단독 strict 로 valid 여도 어떤 change 도 archive 할 수 없었다. a190 은 판정을 둘로 나눴다:
(가) 닫는 change 는 strict, (나) 저장소 전체는 strict 없이 오류만. 저장소 전체 strict 는 advisory 한 줄.

무엇을 관찰하는가
-----------------
- 판정 단위는 `tools/sdd/check_openspec_validate.py` 다. 1~10단계를 돌리지 않고 **그 단위를 직접** 부른다.
  거절은 「어디선가 rc 1」 이 아니라 **(가)/(나) 중 어느 쪽이, 어느 행(V1~V13, design.md D3)으로** 멈췄는지 본다.
- 진짜 openspec 으로 재는 갈래(R)는 CLI 가 없으면 사유를 대고 건너뛴다(CI 에는 openspec 이 없다).
  스텁 갈래(S)는 CI 에서도 돌며, 스텁은 받은 argv 를 기록해 명령 모양(`--strict` 유무, id)까지 단언한다.
- 배선은 실제 `tools/gate.sh` 를 임시 저장소에서 돌려 본다. 5단계와 6~10단계는 스텁으로 통과시키고,
  스텁 `Makefile` 의 `validate` 타깃은 **실패**하게 둔다 — 게이트가 옛 `make validate` 로 돌아가면 그 자리에서 빨개진다.
"""

from __future__ import annotations

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
GATE = ROOT / "tools" / "gate.sh"
SCRIPT = ROOT / "tools" / "sdd" / "check_openspec_validate.py"
OPENSPEC = shutil.which("openspec")
NEEDS_OPENSPEC = unittest.skipUnless(
    OPENSPEC, "openspec CLI 가 PATH 에 없다 — 진짜 CLI 갈래는 건너뛴다(스텁 갈래는 돈다)"
)

CLOSING = "a999-closing"
OTHER = "a998-other"

# openspec 1.16.0 은 기본으로 사용 통계를 보낸다. 시험은 끄고 돈다.
QUIET = {"OPENSPEC_TELEMETRY": "0", "DO_NOT_TRACK": "1"}

LONG = "위젯은 요청마다 응답을 기록해야 한다(SHALL). " + "이 문장은 strict 경고를 만들기 위해 길게 늘인 서술이다. " * 20

# 스텁 openspec. argv 로 어느 호출인지 가르고, 계획 파일의 rc·stdout·sleep 을 따른다.
STUB = '''#!{python}
import json, os, sys, time
argv = sys.argv[1:]
mode = "advisory" if ("--all" in argv and "--strict" in argv) else ("repo" if "--all" in argv else "closing")
with open(os.environ["STUB_OPENSPEC_LOG"], "a", encoding="utf-8") as log:
    log.write(json.dumps({{"mode": mode, "argv": argv, "cwd": os.getcwd()}}) + "\\n")
plan = json.load(open(os.environ["STUB_OPENSPEC_PLAN"], encoding="utf-8"))[mode]
time.sleep(plan.get("sleep", 0))
sys.stdout.write(plan["stdout"])
sys.exit(plan["rc"])
'''


def write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


# ---- 진짜 openspec 픽스처 --------------------------------------------------


def spec(repo: Path, name: str, *, warn: bool = False) -> None:
    """main spec 하나. warn 이면 Purpose 가 50자 미만이라 strict 에서만 경고가 난다."""
    purpose = "짧음" if warn else "이 능력은 시험 픽스처의 위젯이다. 이 문장은 자리표시가 아니라 사람이 쓴 실제 목적 서술이다."
    write(
        repo / "openspec" / "specs" / name / "spec.md",
        f"# {name} Specification\n\n## Purpose\n{purpose}\n\n## Requirements\n"
        "### Requirement: 위젯은 응답한다\n위젯은 요청에 응답해야 한다(SHALL).\n\n"
        "#### Scenario: 요청\n- **WHEN** 요청이 오면\n- **THEN** 응답한다\n",
    )


def change(repo: Path, rel: str, *, kind: str = "clean") -> None:
    """change 하나. clean = strict valid, warn = strict 에서만 경고(요구 문장 500자 초과), error = delta 없음(오류)."""
    base = repo / "openspec" / "changes" / rel
    write(base / "proposal.md", "## Why\n시험 픽스처의 change 다.\n\n## What Changes\n- 위젯이 응답을 기록한다.\n")
    write(base / "tasks.md", "## 1. 작업\n\n- [x] 1.1 끝났다\n")
    if kind == "error":
        return
    text = LONG if kind == "warn" else "위젯은 응답을 기록해야 한다(SHALL)."
    write(
        base / "specs" / "widget" / "spec.md",
        f"## ADDED Requirements\n### Requirement: 위젯은 응답을 기록한다 {rel}\n{text}\n\n"
        "#### Scenario: 기록\n- **WHEN** 응답하면\n- **THEN** 기록한다\n",
    )


# ---- 스텁 계획 ---------------------------------------------------------------


def report(root: Path, items: list[dict], *, root_path: str | None = None) -> str:
    """openspec 1.16.0 `--json` 모양(실측, design.md D3)을 흉내 낸다."""
    passed = sum(1 for i in items if i.get("valid") is True)
    return json.dumps(
        {
            "items": items,
            "summary": {"totals": {"items": len(items), "passed": passed, "failed": len(items) - passed}},
            "version": "1.0",
            "root": {"path": root_path if root_path is not None else str(root.resolve()), "source": "nearest"},
        }
    )


def item(id_: str, *, type_: str = "change", valid=True) -> dict:
    return {"id": id_, "type": type_, "valid": valid, "issues": [], "durationMs": 1}


class Sandbox:
    """임시 저장소 하나. 판정 스크립트(와 필요하면 gate.sh)를 복사해 그 저장소가 ROOT 가 되게 한다."""

    def __init__(self, case: unittest.TestCase) -> None:
        self._tmp = tempfile.TemporaryDirectory()
        case.addCleanup(self._tmp.cleanup)
        self.repo = Path(self._tmp.name) / "repo"
        self.bin = Path(self._tmp.name) / "bin"
        self.bin.mkdir(parents=True)
        (self.repo / "openspec" / "specs").mkdir(parents=True)
        (self.repo / "openspec" / "changes").mkdir(parents=True)
        (self.repo / "tools" / "sdd").mkdir(parents=True)
        shutil.copy2(SCRIPT, self.repo / "tools" / "sdd" / SCRIPT.name)
        self.log = Path(self._tmp.name) / "stub.log"
        self.plan = Path(self._tmp.name) / "plan.json"

    def stub(self, closing: tuple[int, str], repo: tuple[int, str], advisory: tuple[int, str], sleep: float = 0) -> None:
        stub = self.bin / "openspec"
        stub.write_text(STUB.format(python=sys.executable), encoding="utf-8")
        stub.chmod(0o755)
        self.plan.write_text(
            json.dumps(
                {
                    "closing": {"rc": closing[0], "stdout": closing[1], "sleep": sleep},
                    "repo": {"rc": repo[0], "stdout": repo[1]},
                    "advisory": {"rc": advisory[0], "stdout": advisory[1]},
                }
            ),
            encoding="utf-8",
        )

    def env(self, *, path: str) -> dict[str, str]:
        env = dict(os.environ, **QUIET)
        env.update(PATH=path, STUB_OPENSPEC_LOG=str(self.log), STUB_OPENSPEC_PLAN=str(self.plan))
        return env

    def stub_path(self) -> str:
        """스텁이 맨 앞. 진짜 openspec 이 뒤에 있어도 스텁이 이긴다."""
        return f"{self.bin}{os.pathsep}{os.environ.get('PATH', '')}"

    def calls(self) -> list[dict]:
        if not self.log.exists():
            return []
        return [json.loads(line) for line in self.log.read_text(encoding="utf-8").splitlines()]

    def check(self, change_id: str = CLOSING, *, path: str | None = None) -> subprocess.CompletedProcess[str]:
        """판정 단위를 직접 부른다. 인터프리터는 절대경로라 PATH 를 비워도 돈다."""
        return subprocess.run(
            [sys.executable, str(self.repo / "tools" / "sdd" / SCRIPT.name), change_id],
            cwd=self.repo,
            capture_output=True,
            text=True,
            check=False,
            env=self.env(path=path if path is not None else os.environ.get("PATH", "")),
            timeout=600,
        )


def out(result: subprocess.CompletedProcess[str]) -> str:
    return result.stdout + result.stderr


# ---- R: 진짜 openspec ----------------------------------------------------------


@NEEDS_OPENSPEC
class RealOpenSpecVerdicts(unittest.TestCase):
    def setUp(self) -> None:
        self.box = Sandbox(self)
        spec(self.box.repo, "widget")
        change(self.box.repo, CLOSING)

    def test_1_strict_warnings_elsewhere_do_not_fail_the_step(self) -> None:
        """① 남의 change·main spec 에만 strict 경고 → 통과, advisory 줄은 그 실패를 센다."""
        spec(self.box.repo, "gadget", warn=True)
        change(self.box.repo, OTHER, kind="warn")
        result = self.box.check()
        text = out(result)
        self.assertEqual(0, result.returncode, text)
        self.assertIn(f"(가) OK: {CLOSING}", text)
        self.assertIn("(나) OK:", text)
        # ⑦ advisory 는 판정이 아님을 말하고, 실제 strict 실패 수(2)를 센다 — 그리고 rc 는 0 이다.
        self.assertIn("advisory — 판정 아님: 저장소 전체 strict 2 passed, 2 failed (4 items)", text)

    def test_2_a_strict_warning_in_the_closing_change_fails_at_the_closing_check(self) -> None:
        """② 닫는 change 에 strict 경고 → (가)가 V11 로 FAIL, (나)는 OK."""
        shutil.rmtree(self.box.repo / "openspec" / "changes" / CLOSING)
        change(self.box.repo, CLOSING, kind="warn")
        result = self.box.check()
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn("(가) FAIL [V11]", text)
        self.assertIn("(나) OK:", text)

    def test_3_an_error_anywhere_fails_at_the_repository_check(self) -> None:
        """③ 다른 change 에 오류(delta 없음) → (나)가 V11 로 FAIL, (가)는 OK."""
        change(self.box.repo, OTHER, kind="error")
        result = self.box.check()
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn(f"(가) OK: {CLOSING}", text)
        self.assertIn("(나) FAIL [V11]", text)
        self.assertIn(OTHER, result.stderr)

    def test_5_a_change_openspec_cannot_find_fails(self) -> None:
        """⑤ CHANGE_ID 를 openspec 이 못 찾음 → (가) V4(unknown_item), (나) V10(그 change 를 본 적 없음)."""
        result = self.box.check("a997-nowhere")
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn("(가) FAIL [V4]", text)
        self.assertIn("unknown_item", text)
        self.assertIn("(나) FAIL [V10]", text)

    def test_an_archived_target_fails_and_says_why(self) -> None:
        """아카이브된 change 는 openspec 이 찾지 않는다 — 통과가 아니라 FAIL, 사유에 그 원인을 적는다."""
        shutil.rmtree(self.box.repo / "openspec" / "changes" / CLOSING)
        change(self.box.repo, f"archive/2026-11-01-{CLOSING}")
        result = self.box.check()
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn("(가) FAIL [V4]", text)
        self.assertIn("아카이브된 change 는 openspec 이 검증하지 않는다", text)

    def test_6_a_repository_check_over_zero_items_fails(self) -> None:
        """⑥ (나)가 0 항목을 검사하고 rc 0 → 그래도 FAIL(V9). 전칭 판정은 표본 0 이면 자동으로 참이 된다."""
        shutil.rmtree(self.box.repo / "openspec" / "changes" / CLOSING)
        shutil.rmtree(self.box.repo / "openspec" / "specs" / "widget")
        result = self.box.check()
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn("(나) FAIL [V9]", text)


# ---- S: 스텁 (CI 에서도 돈다) --------------------------------------------------


class StubbedVerdicts(unittest.TestCase):
    def setUp(self) -> None:
        self.box = Sandbox(self)
        root = self.box.repo
        self.ok_closing = (0, report(root, [item(CLOSING)]))
        self.ok_repo = (0, report(root, [item(CLOSING), item("widget", type_="spec")]))
        self.ok_advisory = (1, report(root, [item(CLOSING), item("widget", type_="spec", valid=False)]))

    def run_with(self, closing=None, repo=None, advisory=None) -> subprocess.CompletedProcess[str]:
        self.box.stub(closing or self.ok_closing, repo or self.ok_repo, advisory or self.ok_advisory)
        return self.box.check(path=self.box.stub_path())

    def test_all_ok_passes_and_the_three_commands_have_the_decided_shape(self) -> None:
        """(가)는 id + --strict, (나)는 --all 이고 --strict 없음, advisory 는 --all --strict. 셋 다 저장소 루트에서."""
        result = self.run_with()
        self.assertEqual(0, result.returncode, out(result))
        calls = {c["mode"]: c for c in self.box.calls()}
        self.assertEqual(["closing", "repo", "advisory"], [c["mode"] for c in self.box.calls()])
        self.assertEqual([("validate", CLOSING, "--strict", "--no-interactive", "--json")], [tuple(calls["closing"]["argv"])])
        self.assertEqual(("validate", "--all", "--no-interactive", "--json"), tuple(calls["repo"]["argv"]))
        self.assertEqual(("validate", "--all", "--strict", "--no-interactive", "--json"), tuple(calls["advisory"]["argv"]))
        for call in calls.values():
            self.assertEqual(str(self.box.repo.resolve()), str(Path(call["cwd"]).resolve()))
        self.assertIn("advisory — 판정 아님: 저장소 전체 strict 1 passed, 1 failed (2 items)", out(result))

    def test_4_openspec_absent_fails_both_checks(self) -> None:
        """④ PATH 에 openspec 이 없음 → (가)·(나) 둘 다 V1."""
        result = self.box.check(path=str(self.box.bin))  # 스텁을 만들지 않은 빈 디렉터리
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assertIn("(가) FAIL [V1]", text)
        self.assertIn("(나) FAIL [V1]", text)

    def test_7_an_advisory_that_crashes_never_changes_the_verdict(self) -> None:
        """⑦ advisory 명령이 죽어도(rc 2, JSON 아님) 판정은 그대로 통과, 사유 한 줄이 남는다."""
        result = self.run_with(advisory=(2, "boom"))
        text = out(result)
        self.assertEqual(0, result.returncode, text)
        self.assertIn("advisory — 판정 아님: 저장소 전체 strict 합계를 얻지 못함", text)

    def test_advisory_runs_even_when_the_verdict_fails(self) -> None:
        result = self.run_with(closing=(1, report(self.box.repo, [item(CLOSING, valid=False)])))
        self.assertEqual(1, result.returncode)
        self.assertIn("advisory — 판정 아님:", out(result))
        self.assertIn("advisory", [c["mode"] for c in self.box.calls()])

    def test_v3_output_that_is_not_json_fails(self) -> None:
        result = self.run_with(closing=(0, "Change is valid"), repo=(0, "Totals: 2 passed"))
        self.assertIn("(가) FAIL [V3]", out(result))
        self.assertIn("(나) FAIL [V3]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v5_an_error_status_alongside_items_fails(self) -> None:
        body = json.loads(self.ok_closing[1])
        body["status"] = [{"severity": "error", "code": "x", "message": "깨짐"}]
        result = self.run_with(closing=(0, json.dumps(body)))
        self.assertIn("(가) FAIL [V5]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v6_a_report_about_another_root_fails(self) -> None:
        other = report(self.box.repo, [item(CLOSING)], root_path="/somewhere/else")
        result = self.run_with(closing=(0, other))
        self.assertIn("(가) FAIL [V6]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v6_a_report_without_a_root_fails(self) -> None:
        body = json.loads(self.ok_repo[1])
        del body["root"]
        result = self.run_with(repo=(0, json.dumps(body)))
        self.assertIn("(나) FAIL [V6]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v7_valid_must_be_the_boolean_true(self) -> None:
        result = self.run_with(repo=(0, report(self.box.repo, [item(CLOSING), item("widget", type_="spec", valid="true")])))
        self.assertIn("(나) FAIL [V7]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v8_the_closing_report_must_name_exactly_this_change(self) -> None:
        for label, items in (
            ("다른 id", [item("a000-someone-else")]),
            ("spec 으로 해소", [item(CLOSING, type_="spec")]),
            ("둘", [item(CLOSING), item(CLOSING)]),
        ):
            with self.subTest(label):
                self.box.log.unlink(missing_ok=True)
                result = self.run_with(closing=(0, report(self.box.repo, items)))
                self.assertIn("(가) FAIL [V8]", out(result))
                self.assertEqual(1, result.returncode)

    def test_v10_a_repository_report_that_never_saw_the_closing_change_fails(self) -> None:
        result = self.run_with(repo=(0, report(self.box.repo, [item("widget", type_="spec")])))
        self.assertIn("(나) FAIL [V10]", out(result))
        self.assertEqual(1, result.returncode)

    def test_v12_rc_and_json_must_agree(self) -> None:
        for label, closing in (
            ("rc 0 인데 invalid", (0, report(self.box.repo, [item(CLOSING, valid=False)]))),
            ("rc 1 인데 전부 valid", (1, report(self.box.repo, [item(CLOSING)]))),
        ):
            with self.subTest(label):
                result = self.run_with(closing=closing)
                self.assertIn("(가) FAIL [V12]", out(result))
                self.assertEqual(1, result.returncode)


class TimeoutVerdict(unittest.TestCase):
    """V2 — 시한 초과. 300초를 기다릴 수 없으니 시한을 인자로 넣고 잠자는 스텁을 부른다."""

    def test_v2_a_command_that_outlives_its_deadline_fails(self) -> None:
        box = Sandbox(self)
        box.stub((0, report(box.repo, [item(CLOSING)])), (0, "{}"), (0, "{}"), sleep=5)
        spec_ = importlib.util.spec_from_file_location("check_openspec_validate_copy", box.repo / "tools" / "sdd" / SCRIPT.name)
        module = importlib.util.module_from_spec(spec_)
        spec_.loader.exec_module(module)
        with mock.patch.dict(os.environ, box.env(path=box.stub_path()), clear=True):
            verdict, reasons = module.closing_verdict(CLOSING, module.ROOT, timeout=0.5)
        self.assertFalse(verdict)
        self.assertTrue(any(r.startswith("[V2]") for r in reasons), reasons)


# ---- 배선: 실제 gate.sh -----------------------------------------------------------


class GateWiring(unittest.TestCase):
    """11단계가 판정 단위를 **이 change id 로** 부르고, 옛 `make validate` 를 부르지 않는지 실제 gate.sh 로 본다."""

    def setUp(self) -> None:
        self.box = Sandbox(self)
        repo = self.box.repo
        shutil.copy2(GATE, repo / "tools" / "gate.sh")
        write(repo / "tools" / "logic-map" / "check_analysis.py", "import sys\nsys.exit(0)\n")
        # 6~10단계 타깃은 통과, 옛 11단계 `validate` 는 실패 — 게이트가 그것을 부르면 여기서 빨개진다.
        write(
            repo / "Makefile",
            textwrap.dedent(
                """\
                sdd-check test test-seams test-race vet:
                \t@true
                validate:
                \t@echo "옛 make validate 가 불렸다" >&2; exit 1
                """
            ),
        )
        write(repo / "openspec" / "changes" / CLOSING / "review.md", "# 리뷰\n")
        spec(repo, "widget")
        change(repo, CLOSING)

    def gate(self, path: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["bash", str(self.box.repo / "tools" / "gate.sh"), CLOSING],
            capture_output=True,
            text=True,
            check=False,
            env=self.box.env(path=path),
            timeout=600,
        )

    def assert_reached_step_eleven(self, text: str) -> None:
        self.assertIn("OK: make vet", text)
        self.assertIn("==> 11/11", text)
        self.assertNotIn("옛 make validate 가 불렸다", text)

    def test_stubbed_pass_reaches_gate_pass_with_this_change_id(self) -> None:
        root = self.box.repo
        self.box.stub((0, report(root, [item(CLOSING)])), (0, report(root, [item(CLOSING)])), (1, report(root, [item(CLOSING, valid=False)])))
        result = self.gate(self.box.stub_path())
        text = out(result)
        self.assertEqual(0, result.returncode, text)
        self.assert_reached_step_eleven(text)
        self.assertIn(f"GATE PASS: {CLOSING}", text)
        self.assertIn("advisory — 판정 아님:", text)
        closing = [c for c in self.box.calls() if c["mode"] == "closing"]
        self.assertEqual(1, len(closing), self.box.calls())
        self.assertEqual(CLOSING, closing[0]["argv"][1])

    def test_stubbed_closing_failure_stops_the_gate_at_step_eleven(self) -> None:
        root = self.box.repo
        self.box.stub((1, report(root, [item(CLOSING, valid=False)])), (0, report(root, [item(CLOSING)])), (0, report(root, [item(CLOSING)])))
        result = self.gate(self.box.stub_path())
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assert_reached_step_eleven(text)
        self.assertIn("(가) FAIL [V11]", text)
        self.assertIn(f"GATE FAIL: {CLOSING} — openspec validate 실패", text)

    @NEEDS_OPENSPEC
    def test_real_openspec_passes_with_strict_warnings_elsewhere(self) -> None:
        spec(self.box.repo, "gadget", warn=True)
        result = self.gate(os.environ.get("PATH", ""))
        text = out(result)
        self.assertEqual(0, result.returncode, text)
        self.assert_reached_step_eleven(text)
        self.assertIn(f"GATE PASS: {CLOSING}", text)

    @NEEDS_OPENSPEC
    def test_real_openspec_fails_at_step_eleven_on_a_closing_warning(self) -> None:
        shutil.rmtree(self.box.repo / "openspec" / "changes" / CLOSING / "specs")
        change(self.box.repo, CLOSING, kind="warn")
        result = self.gate(os.environ.get("PATH", ""))
        text = out(result)
        self.assertEqual(1, result.returncode, text)
        self.assert_reached_step_eleven(text)
        self.assertIn("(가) FAIL [V11]", text)
        self.assertIn(f"GATE FAIL: {CLOSING} — openspec validate 실패", text)


if __name__ == "__main__":
    unittest.main()
