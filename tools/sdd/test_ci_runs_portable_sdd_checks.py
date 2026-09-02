"""CI 가 도는 SDD 검사와 `make sdd-check` 가 도는 것이 갈라지지 않게 묶는다.

a118 이 남긴 교훈이 이 파일의 이유다. 태그 뒤 테스트 71개는 "있었지만 어느 게이트에서도
돌지 않았고", 그래서 그중 하나가 한 달을 실패한 채 숨어 있었다. 검사는 존재하는 것으로는
아무것도 막지 못하고, **도는 곳**이 있어야 막는다.

`make sdd-check` 의 검사 중 일부는 개발 워크스테이션에만 있는 것(codegraph 실행 파일,
gitignore 된 `.sdd/index-state.json`)을 필요로 해서 GitHub 러너에서는 돌 수 없다.
그래서 돌 수 있는 것만 `sdd-check-ci` 로 떼어 내고 CI 가 그것을 돈다.

이 분리에는 하나의 실패 방식이 있다: 나중에 누가 **옮길 수 있는** 검사를 `sdd-check` 에
직접 한 줄 더하는 것. 그러면 그 검사는 로컬에서만 돌고 CI 는 모른다 — a118 이 반복된다.
아래 세 시험이 그 한 줄을 실패로 만든다.
"""

import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MAKEFILE = ROOT / "Makefile"
WORKFLOW = ROOT / ".github" / "workflows" / "ci.yml"

# `sdd-check` 가 `sdd-check-ci` 위에 얹어도 되는 것 — 개발 워크스테이션에만 있는 것을
# 필요로 해서 러너로 옮길 수 없다고 **측정으로** 확인한 둘이다(2026-09-02, 외부 도구를
# 지운 PATH + depth-1 클론에서 각각 exit 1):
#
#   sdd-doctor              : rtk·openspec·codegraph·gbrain 등 로컬 설치 도구를 본다
#   check_index_freshness.py: gitignore 된 `.sdd/index-state.json` 과 codegraph 실행 파일을 본다
#
# 이 목록은 허용이 아니라 **예외**다. 여기 없는 줄이 `sdd-check` 에 붙으면 시험이 깨지고,
# 고치는 방법은 목록을 늘리는 것이 아니라 그 검사를 `sdd-check-ci` 에 넣는 것이다.
WORKSTATION_ONLY = (
    "sdd-doctor",
    "python3 tools/sdd/check_index_freshness.py",
)

DELEGATION = "$(MAKE) sdd-check-ci"


def recipe(target: str, text: str) -> tuple[list[str], list[str]]:
    """`target` 의 선행 조건과 recipe 줄을 돌려준다.

    make 의 recipe 는 TAB 으로 시작하는 줄이다. TAB 으로 시작하지 않는 첫 줄에서 끝난다 —
    빈 줄은 recipe 를 끝내지 않으므로 건너뛴다.
    """
    lines = text.splitlines()
    head = None
    for index, line in enumerate(lines):
        if line.startswith(f"{target}:"):
            head = index
            break
    if head is None:
        raise AssertionError(f"Makefile 에 {target} 타깃이 없습니다")

    prerequisites = lines[head].split(":", 1)[1].split()
    body: list[str] = []
    for line in lines[head + 1 :]:
        if not line.strip():
            continue
        if not line.startswith("\t"):
            break
        stripped = line[1:].strip()
        if stripped.startswith("#"):
            continue
        body.append(stripped)
    return prerequisites, body


class CIRunsPortableSDDChecksTests(unittest.TestCase):
    def setUp(self):
        self.makefile = MAKEFILE.read_text(encoding="utf-8")
        self.workflow = WORKFLOW.read_text(encoding="utf-8")

    def test_ci_workflow_runs_the_portable_subset(self):
        """CI 워크플로가 `make sdd-check-ci` 를 실제로 돈다."""
        step = re.compile(r"^\s+run:\s*make sdd-check-ci\s*$", re.MULTILINE)
        self.assertRegex(
            self.workflow,
            step,
            "ci.yml 에 `run: make sdd-check-ci` 단계가 없습니다 — "
            "옮길 수 있는 SDD 검사가 CI 에서 돌지 않습니다",
        )

    def test_ci_workflow_runs_on_pull_requests(self):
        """그 단계는 PR 에서 돌아야 의미가 있다."""
        self.assertRegex(
            self.workflow,
            re.compile(r"^  pull_request:\s*$", re.MULTILINE),
            "ci.yml 이 pull_request 에서 돌지 않습니다",
        )

    def test_sdd_check_adds_nothing_portable_on_top_of_the_ci_subset(self):
        """`sdd-check` 는 위임 + 워크스테이션 전용 예외로만 이루어진다."""
        prerequisites, body = recipe("sdd-check", self.makefile)
        self.assertIn(
            DELEGATION,
            body,
            "sdd-check 가 sdd-check-ci 에 위임하지 않습니다 — "
            "두 곳이 서로 다른 검사 목록을 말할 수 있게 됩니다",
        )

        surface = [*prerequisites, *(line for line in body if line != DELEGATION)]
        unexpected = [line for line in surface if line not in WORKSTATION_ONLY]
        self.assertEqual(
            unexpected,
            [],
            "sdd-check 에 sdd-check-ci 를 거치지 않는 검사가 있습니다: "
            f"{unexpected} — 러너에서 돌 수 있는 검사라면 sdd-check-ci 로 옮기세요. "
            "옮길 수 없다면 WORKSTATION_ONLY 에 이유와 함께 추가하세요.",
        )

    def test_sdd_check_ci_is_phony(self):
        """`sdd-check-ci` 라는 이름의 파일이 생기면 타깃이 조용히 안 돌게 된다."""
        phony = re.search(r"^\.PHONY:(.*?)(?=^\S)", self.makefile, re.MULTILINE | re.DOTALL)
        self.assertIsNotNone(phony, "Makefile 에 .PHONY 선언이 없습니다")
        declared = phony.group(1).replace("\\", " ").split()
        self.assertIn("sdd-check-ci", declared)


if __name__ == "__main__":
    unittest.main()
