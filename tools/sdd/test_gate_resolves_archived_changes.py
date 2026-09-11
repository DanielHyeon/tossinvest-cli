#!/usr/bin/env python3
"""`make gate` 가 아카이브된 change 를 찾는지 본다 (a122 task 2.7).

왜 이게 필요한가
----------------
`openspec archive` 는 끝난 change 를 `openspec/changes/archive/<YYYY-MM-DD>-<id>` 로
옮긴다. `tools/gate.sh` 는 `openspec/changes/<id>` 만 봤다. 그래서 **짝이 먼저
아카이브되면 남은 쪽의 완료 게이트가 영구히 막힌다.**

2026-09-09 실물: `a099` 의 짝 `a098` 이 2026-08-29 에 아카이브됐고,
`make gate CHANGE=a099-…` 는 3단계에서 죽었다. 같은 change 의 5단계는 통과하는데
3단계가 거기 도달하는 것을 막고 있었다.

같은 모양의 결함을 이 저장소는 이미 두 번 고쳤다(`f6965ebb`, a122 task 3.2.2).
둘 다 Python 이라 `resolve_referenced_change` 한 함수를 공유했다. gate.sh 는 shell 이라
그 함수를 부를 수 없어 규칙을 **옮겨 적는다**. 그러면 정본이 둘이 되므로 여기서
shell 쪽을, `test_check_analysis.py` 가 Python 쪽을 각각 못 박는다.

무엇을 관찰하는가
-----------------
3단계까지만 본다. 픽스처는 `review.md` 를 두지 않아 **4단계에서 멈춘다** — 그래서
`make test` 같은 무거운 단계가 돌지 않고, 3단계를 넘었다는 사실만 정확히 잰다.
"""

from __future__ import annotations

import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GATE = ROOT / "tools" / "gate.sh"


def write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def make_change(repo: Path, rel: str, *, pairs: list[str] | None = None) -> None:
    """완료된 change 하나를 만든다. 미완료 task 는 0건이다."""
    base = repo / "openspec" / "changes" / rel
    write(base / "tasks.md", "## 1. 작업\n\n- [x] 1.1 끝났다\n")
    if pairs is not None:
        write(base / "deploy-pair.txt", "".join(f"{p}\n" for p in pairs))


def run_gate(repo: Path, change_id: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["bash", str(repo / "tools" / "gate.sh"), change_id],
        capture_output=True,
        text=True,
        check=False,
    )


class GateArchiveResolutionTest(unittest.TestCase):
    def setUp(self) -> None:
        self._tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self._tmp.name)
        # gate.sh 는 자기 위치에서 repo-root 를 계산한다. 복사만 하면 임시 저장소가 root 가 된다.
        (self.repo / "tools").mkdir(parents=True)
        shutil.copy2(GATE, self.repo / "tools" / "gate.sh")
        self.addCleanup(self._tmp.cleanup)

    def assert_passed_step_three(self, result: subprocess.CompletedProcess[str]) -> None:
        """3단계를 넘었는가 = 4단계(review.md 없음)에서 죽었는가."""
        out = result.stdout + result.stderr
        self.assertIn("4/", out, f"3단계를 못 넘었다:\n{out}")
        self.assertNotIn("짝 change 의 tasks.md 가 없습니다", out)

    def assert_stopped_before_step_four(
        self, result: subprocess.CompletedProcess[str]
    ) -> None:
        """거절이 **여기서** 일어났는가.

        픽스처에 `review.md` 가 없어 게이트는 어차피 4단계에서 죽는다. 그래서
        `returncode != 0` 만 보면 무엇이든 통과한다 — 해소기가 완전히 망가져도
        초록이다. 그러니 4단계에 **도달하지 못했다**는 것까지 본다.
        """
        out = result.stdout + result.stderr
        self.assertNotEqual(0, result.returncode, out)
        self.assertNotIn("4/", out, f"거절되어야 하는데 3단계를 넘었다:\n{out}")

    # (d) 새로 되어야 하는 것 — 아카이브된 짝
    def test_archived_pair_is_found(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirs"])
        make_change(self.repo, "archive/2026-08-29-a998-theirs", pairs=["a999-mine"])
        self.assert_passed_step_three(run_gate(self.repo, "a999-mine"))

    # (a) 죽이면 안 되는 것 — 아직 활성인 짝
    def test_active_pair_still_works(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirs"])
        make_change(self.repo, "a998-theirs", pairs=["a999-mine"])
        self.assert_passed_step_three(run_gate(self.repo, "a999-mine"))

    # (b) 죽이면 안 되는 것 — 오타 id 는 계속 실패해야 한다
    def test_typo_pair_still_fails(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirz"])
        result = run_gate(self.repo, "a999-mine")
        self.assert_stopped_before_step_four(result)
        self.assertIn("a998-theirz", result.stdout + result.stderr)

    # (b') 접미사만 같은 아카이브본을 오타로 통과시키면 안 된다
    def test_archive_suffix_match_is_not_enough(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["theirs"])
        make_change(self.repo, "archive/2026-08-29-a998-theirs", pairs=["a999-mine"])
        result = run_gate(self.repo, "a999-mine")
        self.assert_stopped_before_step_four(result)

    # (c) 활성과 아카이브에 같은 id 가 동시에 있으면 fail-closed
    def test_active_and_archived_duplicate_fails_closed(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirs"])
        make_change(self.repo, "a998-theirs", pairs=["a999-mine"])
        make_change(self.repo, "archive/2026-08-29-a998-theirs", pairs=["a999-mine"])
        result = run_gate(self.repo, "a999-mine")
        self.assert_stopped_before_step_four(result)

    # (c') 아카이브 안에 같은 id 사본이 둘이면 fail-closed
    def test_two_archived_copies_fail_closed(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirs"])
        make_change(self.repo, "archive/2026-08-29-a998-theirs", pairs=["a999-mine"])
        make_change(self.repo, "archive/2026-09-01-a998-theirs", pairs=["a999-mine"])
        result = run_gate(self.repo, "a999-mine")
        self.assert_stopped_before_step_four(result)

    # `:116` — gate 대상 자신이 아카이브된 경우 1단계에서 죽지 않아야 한다
    def test_gate_target_itself_may_be_archived(self) -> None:
        make_change(self.repo, "archive/2026-08-29-a997-done")
        result = run_gate(self.repo, "a997-done")
        out = result.stdout + result.stderr
        self.assertNotIn("tasks.md 가 없습니다", out, out)
        self.assertIn("4/", out, out)

    # 없는 id 는 여전히 1단계에서 실패한다
    def test_unknown_change_still_fails_at_step_one(self) -> None:
        result = run_gate(self.repo, "a996-nowhere")
        self.assert_stopped_before_step_four(result)

    # a122 6.3 — gate 대상 **자신**이 활성과 아카이브에 동시에 있으면 1단계가 멈춘다.
    # 그 거절은 1단계의 `RESOLVE_ERROR` 블록 하나가 한다. 블록을 지워도 기존 시험은 전부
    # 초록이었다 — 대체 경로 `openspec/changes/<id>` 가 **실재**해서 게이트가 활성
    # 디렉터리로 그대로 진행하고 4단계에서 죽기 때문이다. 그래서 **무엇이라고** 멈추는지
    # 1단계의 문장으로 본다.
    def test_gate_target_open_and_archived_at_once_stops_at_step_one(self) -> None:
        make_change(self.repo, "a995-twice")
        make_change(self.repo, "archive/2026-08-29-a995-twice")
        result = run_gate(self.repo, "a995-twice")
        self.assert_stopped_before_step_four(result)
        self.assertIn("확정할 수 없는 change-id", result.stdout + result.stderr)

    # a122 6.3 — 날짜 모양이 아닌 이름은 아카이브가 아니다. `case` 가드를 지우면
    # `${name#????-??-??-}` 의 `?` 가 아무 글자나 먹어 `abcd-ef-gh-<id>` 가 통과한다.
    def test_archive_name_without_a_date_is_not_an_archive(self) -> None:
        make_change(self.repo, "a999-mine", pairs=["a998-theirs"])
        make_change(self.repo, "archive/abcd-ef-gh-a998-theirs", pairs=["a999-mine"])
        result = run_gate(self.repo, "a999-mine")
        self.assert_stopped_before_step_four(result)
        self.assertIn("짝 change 를 찾을 수 없습니다", result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
