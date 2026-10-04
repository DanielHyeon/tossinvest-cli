"""`*_testseam.go` 는 전부 생산 빌드 밖에 있어야 한다(a112 8.8.4 항목 6 — Manager 판정 2026-10-04).

test seam(시험이 불투명 권한을 주조하는 함수 — 예: `strategyrouter.FamilyActivationForTest`)은 `//go:build tossos_testseams` 아래에만
있어서 생산 바이너리에 그 문이 아예 존재하지 않는다. 그 성질은 지금까지 파일마다 손으로 적은 빌드 줄 하나가 지켰고, 그 줄이 빠진
파일은 **조용히 생산 빌드에 들어간다**(컴파일 오류도 시험 실패도 없다). 이 가드가 그 빠짐을 실패로 만든다.

규칙(모르는 모양은 건너뛰지 않고 실패):
  - 저장소의 모든 `*_testseam.go`(`_test.go` 아님)의 첫 줄은 정확히 `//go:build tossos_testseams` 다.
  - 파일 안의 다른 `//go:build` · `// +build` 줄은 없다(첫 줄 하나뿐).
  - 전칭 판정이 빈 표본에서 통과하지 않게, 찾은 파일 수와 목록을 아래 EXPECTED 에 핀한다 — 0 개를 찾으면 실패이고, 하나가 늘거나 줄면
    목록 편집이 강제된다(그 편집이 리뷰에 보인다).
"""

import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TAG_LINE = "//go:build tossos_testseams"
SEARCH_ROOTS = ("internal", "cmd", "tools")

# 2026-10-04 실측(git ls-files '*_testseam.go' · 디스크 전수 같음): 24 파일.
# 2026-10-05 +1: a112 7.3.1 SHADOW 적재 훅 `strategy_lane_shadow_load_testseam.go`(4cbcfb36) — 첫 줄 `//go:build tossos_testseams` 확인, 25 파일.
EXPECTED = sorted([
    "internal/app/engine/strategy_lane_shadow_load_testseam.go",
    "internal/app/engine/strategy_lane_step_testseam.go",
    "internal/continuationlane/strategyflow_testseam.go",
    "internal/execgw/a124_entry_gate_testseam.go",
    "internal/execgw/strategy_authority_testseam.go",
    "internal/officialfx/authority_testseam.go",
    "internal/protectionreadiness/authority_testseam.go",
    "internal/reversallane/strategyflow_testseam.go",
    "internal/risk/account_base_testseam.go",
    "internal/scheduler/activation_testseam.go",
    "internal/strategyaccount/authority_testseam.go",
    "internal/strategy/approved_testseam.go",
    "internal/strategycandidate/approved_batch_testseam.go",
    "internal/strategyflow/authority_stop_provenance_testseam.go",
    "internal/strategyflow/authority_testseam.go",
    "internal/strategyflow/canonical_projection_testseam.go",
    "internal/strategyflow/lane_input_testseam.go",
    "internal/strategyproposal/multilane_testseam.go",
    "internal/strategyproposal/production_testseam.go",
    "internal/strategyrouter/arbitration_testseam.go",
    "internal/strategyrouter/production_family_activation_testseam.go",
    "internal/strategyrouter/production_route_manifest_testseam.go",
    "internal/strategyrouter/production_testseam.go",
    "internal/strategyrouter/strategyflow_testseam.go",
    "internal/weeklyvaluelane/strategyflow_testseam.go",
])


def found_testseam_files():
    files = []
    for top in SEARCH_ROOTS:
        base = ROOT / top
        if not base.is_dir():
            continue
        for path in base.rglob("*_testseam.go"):
            if path.name.endswith("_test.go"):
                continue
            files.append(path.relative_to(ROOT).as_posix())
    return sorted(files)


def violations(text):
    """빌드 줄 규칙을 어긴 사유 목록(빈 목록 = 통과)."""
    lines = text.splitlines()
    reasons = []
    if not lines or lines[0] != TAG_LINE:
        first = lines[0] if lines else "<empty file>"
        reasons.append(f"first line is {first!r}, want {TAG_LINE!r}")
    for number, line in enumerate(lines[1:], start=2):
        stripped = line.strip()
        if stripped.startswith("//go:build") or stripped.startswith("// +build"):
            reasons.append(f"line {number} is another build constraint: {stripped!r}")
    return reasons


class TestseamBuildTagTests(unittest.TestCase):
    def test_the_census_is_exactly_the_pinned_list(self):
        found = found_testseam_files()
        self.assertTrue(found, "no *_testseam.go found — the walk did not reach the repository, so every check below would pass empty")
        self.assertEqual(
            found, EXPECTED,
            "the *_testseam.go set changed — add or remove it in EXPECTED after checking its build line "
            f"(new: {sorted(set(found) - set(EXPECTED))}, gone: {sorted(set(EXPECTED) - set(found))})",
        )

    def test_every_testseam_file_is_outside_the_production_build(self):
        failures = {}
        for relative in found_testseam_files():
            reasons = violations((ROOT / relative).read_text(encoding="utf-8"))
            if reasons:
                failures[relative] = reasons
        self.assertFalse(failures, f"these seams would compile into the production binary: {failures}")

    def test_the_rule_refuses_every_unknown_shape(self):
        """판정 함수 자신의 반증 — 빌드 줄 없음 · 다른 태그 · 부정 태그 · 둘째 제약 · 빈 파일은 전부 위반이다."""
        good = TAG_LINE + "\n\npackage x\n"
        self.assertEqual(violations(good), [])
        for name, text in {
            "no build line": "package x\n",
            "another tag": "//go:build integration\n\npackage x\n",
            "negated tag": "//go:build !tossos_testseams\n\npackage x\n",
            "tag with an or": "//go:build tossos_testseams || linux\n\npackage x\n",
            "build line not first": "// comment\n" + good,
            "second constraint": TAG_LINE + "\n// +build linux\n\npackage x\n",
            "empty file": "",
        }.items():
            self.assertTrue(violations(text), f"{name}: accepted")


if __name__ == "__main__":
    unittest.main()
