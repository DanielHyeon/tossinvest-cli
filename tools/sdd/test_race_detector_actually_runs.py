"""경합 검출기가 **도는 곳**이 있는지 묶는다.

a118 이 남긴 교훈의 세 번째 적용이다. 태그 뒤 테스트 71개는 "있었지만 어느
게이트에서도 돌지 않았"고, 그래서 하나가 한 달을 실패한 채 숨어 있었다. 2026-09-02
측정으로 `-race` 가 정확히 그 상태였다: Makefile 과 `.github/workflows/ci.yml`
어디에도 그 낱말이 없었다. 그동안 저장소에는 goroutine·채널·`sync.`·`atomic.`
을 쓰는 생산 패키지가 41개 있었다.

**이것이 실제 차이라는 증거.** a112 5.7 의 반증에서, 레인의 첫 실패 이유를 패키지
수준 슬롯에 쓰는 변이는 `-race` 없이 초록이었고 `-race` 로는 동시성 시험 하나가
빨개졌다. 즉 검출기 없이는 그 시험이 증명하는 것이 한 단계 약하다.

이 분리에는 하나의 실패 방식이 있다: 나중에 누가 `test-race` 를 게이트나 CI 에서
빼는 것. 그러면 시험은 남고 검출기는 사라지며, 그 사실은 아무도 못 본다.
아래 시험들이 그 편집을 실패로 만든다.
"""

import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MAKEFILE = ROOT / "Makefile"
WORKFLOW = ROOT / ".github" / "workflows" / "ci.yml"
GATE = ROOT / "tools" / "gate.sh"

TARGET = "test-race"


class RaceDetectorActuallyRunsTests(unittest.TestCase):
    def setUp(self):
        self.makefile = MAKEFILE.read_text(encoding="utf-8")
        self.workflow = WORKFLOW.read_text(encoding="utf-8")
        self.gate = GATE.read_text(encoding="utf-8")

    def test_the_target_exists_and_passes_the_race_flag(self):
        """`make test-race` 가 실제로 `-race` 를 넘긴다."""
        recipe = self._recipe(TARGET)
        self.assertTrue(recipe, "Makefile 에 test-race 타깃이 없습니다")
        self.assertTrue(
            any("-race" in line for line in recipe),
            f"test-race 의 recipe 에 -race 가 없습니다: {recipe}",
        )

    def test_the_target_names_at_least_one_package(self):
        """빈 패키지 목록은 아무것도 안 돌면서 초록이 된다."""
        match = re.search(r"^RACE_PACKAGES\s*=(.*?)(?=\n\n|\n[A-Za-z.])", self.makefile, re.S | re.M)
        self.assertIsNotNone(match, "Makefile 에 RACE_PACKAGES 선언이 없습니다")
        packages = [p for p in match.group(1).replace("\\", " ").split() if p.startswith("./")]
        self.assertGreaterEqual(
            len(packages), 1, "RACE_PACKAGES 가 비어 있습니다 — 검출기가 아무것도 안 봅니다"
        )
        for package in packages:
            self.assertTrue(
                (ROOT / package).is_dir(), f"RACE_PACKAGES 의 {package} 는 디렉터리가 아닙니다"
            )

    def test_the_completion_gate_runs_it(self):
        """`make gate` 의 타깃 목록에 test-race 가 들어 있다."""
        match = re.search(r"^for target in ([^;]+); do", self.gate, re.M)
        self.assertIsNotNone(match, "gate.sh 에서 타깃 순회 줄을 찾지 못했습니다")
        targets = match.group(1).split()
        self.assertIn(
            TARGET,
            targets,
            f"gate.sh 가 도는 타깃은 {targets} 이고 test-race 가 없습니다 — "
            "경합 검출기가 완료 게이트 밖입니다",
        )

    def test_the_gate_step_count_matches_the_targets_it_runs(self):
        """단계 수와 실제로 도는 타깃 수가 갈라지면 진행 표시가 거짓말을 한다."""
        total = re.search(r"^TOTAL_STEPS=(\d+)", self.gate, re.M)
        self.assertIsNotNone(total, "gate.sh 에 TOTAL_STEPS 가 없습니다")
        # 순회 **직전**의 STEP_NO 를 읽는다. 파일 앞쪽에도 STEP_NO 대입이 있으므로
        # 첫 번째를 집으면 다른 단계의 번호를 세게 된다.
        loop = re.search(r"^STEP_NO=(\d+)\nfor target in ([^;]+); do", self.gate, re.M)
        self.assertIsNotNone(loop, "gate.sh 에서 STEP_NO 초기값과 타깃 순회를 함께 찾지 못했습니다")
        first, targets = loop.group(1), loop.group(2).split()
        self.assertEqual(
            int(total.group(1)),
            int(first) - 1 + len(targets),
            "TOTAL_STEPS 가 실제 단계 수와 다릅니다",
        )

    def test_ci_runs_it(self):
        """CI 워크플로가 `make test-race` 를 실제로 돈다."""
        step = re.compile(r"^\s+run:\s*make test-race\s*$", re.MULTILINE)
        self.assertRegex(
            self.workflow,
            step,
            "CI 워크플로에 `run: make test-race` 단계가 없습니다 — "
            "검출기가 사람 손에서만 도는 상태로 돌아갔습니다",
        )

    def test_the_target_is_phony(self):
        """`test-race` 라는 파일이 생기면 타깃이 조용히 안 돌게 된다."""
        phony = re.search(r"^\.PHONY:(.*?)(?=\n[^\t ])", self.makefile, re.S | re.M)
        self.assertIsNotNone(phony, "Makefile 에 .PHONY 선언이 없습니다")
        self.assertIn(TARGET, phony.group(1).split())

    def test_every_engine_concurrency_test_is_named_in_the_race_filter(self):
        """이름으로 고른 목록은 **빠뜨릴 수 있다.** 그 빠뜨림을 실패로 만든다.

        `internal/app/engine` 은 통째로 검출기에 넣을 수 없다(2026-09-03 측정:
        14분 46초). 그래서 `test-race` 는 그 패키지에서 **이름으로** 고른다.
        고르는 방식에는 a118 이 이미 겪은 실패가 있다 — 나중에 늘어난 시험이
        어느 게이트에서도 안 도는 것. 아래는 대조다: 그 파일의 Test 함수
        전부가 목록에 있어야 한다.
        """
        files = re.search(r"^RACE_ENGINE_FILES\s*=(.*?)(?=\n\n|\n[A-Za-z.])", self.makefile, re.S | re.M)
        self.assertIsNotNone(files, "Makefile 에 RACE_ENGINE_FILES 선언이 없습니다")
        paths = [p for p in files.group(1).replace("\\", " ").split() if p.endswith("_test.go")]
        self.assertGreaterEqual(len(paths), 1, "RACE_ENGINE_FILES 가 비어 있습니다")
        declared = set()
        for relative in paths:
            source = ROOT / relative
            self.assertTrue(source.is_file(), f"{source} 가 없습니다 — 목록이 가리키는 파일이 사라졌습니다")
            declared |= set(re.findall(r"^func (Test\w+)\(", source.read_text(encoding="utf-8"), re.M))
        self.assertGreaterEqual(
            len(declared), 1, "그 파일들에 Test 함수가 없습니다 — 목록이 공허합니다"
        )
        match = re.search(r"^RACE_ENGINE_TESTS\s*=\s*(.+)$", self.makefile, re.M)
        self.assertIsNotNone(match, "Makefile 에 RACE_ENGINE_TESTS 선언이 없습니다")
        filtered = set(match.group(1).strip().split("|"))
        missing = sorted(declared - filtered)
        self.assertFalse(
            missing,
            f"이 동시성 시험들이 경합 검출기 밖입니다: {missing}. "
            "RACE_ENGINE_TESTS 에 더하거나, 왜 검출기가 필요 없는지 적으십시오.",
        )
        stale = sorted(filtered - declared)
        self.assertFalse(stale, f"RACE_ENGINE_TESTS 가 없는 시험을 가리킵니다: {stale}")

    def test_the_engine_race_step_actually_runs(self):
        """recipe 에 줄이 두 개여야 한다. 하나가 사라지면 절반이 조용히 빠진다."""
        recipe = self._recipe(TARGET)
        engine = [line for line in recipe if "RACE_ENGINE_TESTS" in line]
        self.assertEqual(
            len(engine), 1, f"test-race 가 엔진 동시성 시험을 돌지 않습니다: {recipe}"
        )
        self.assertIn("-race", engine[0])

    def _recipe(self, target):
        recipe = []
        capturing = False
        for line in self.makefile.splitlines():
            if line.startswith(f"{target}:"):
                capturing = True
                continue
            if not capturing:
                continue
            if not line.startswith("\t"):
                break
            stripped = line[1:].strip()
            if not stripped.startswith("#"):
                recipe.append(stripped)
        return recipe


if __name__ == "__main__":
    unittest.main()
