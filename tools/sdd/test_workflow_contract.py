#!/usr/bin/env python3
"""docs/WORKFLOW.md 가 통째로 사라지는 것을 막는 가드.

왜 이게 필요한가
----------------
2026-07-31, 승인된 change `align-full-sdd-pm-contract` 의 task 4.1 산출물이
`docs/WORKFLOW.md` 에 랜딩했다(`c0619279`, +122/-33). 9시간 22분 뒤 **무관한
콘솔 커밋** `6d61e988` 이 같은 파일에서 138줄을 지웠다. 두 커밋 사이에 이 파일을
만진 커밋은 없고 병합 커밋도 아니다 — 콘솔 change 의 diff 에 워크플로 계약이
섞여 들어가 옛 판으로 덮인 것이다.

그 뒤 6주 동안 아무도 몰랐다. `make gate` 도 `make sdd-check` 도 이 파일을 보지
않기 때문이다. 그리고 그동안 5단계(증거 조정) 산출물은 한 건도 생산되지 않았다.

그래서 여기서 검사한다. 이 가드는 문장의 품질을 보지 않는다. **정본에서 사라지면
안 되는 제목과 계약 문구가 아직 거기 있는지만** 본다. 그것만으로 위 사고는 잡힌다.

이 파일은 `make sdd-test` 가 돌리고, 그것을 `make sdd-check-ci`(CI) 와
`make gate` 6단계가 부른다.
"""

from __future__ import annotations

import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / "docs" / "WORKFLOW.md"

# 사라지면 안 되는 절 제목. 줄 전체가 정확히 일치해야 한다 —
# 제목이 좁아지는 것(예: "## Pre-Edit 선언" → "## Pre-Edit 선언 (High-risk 전용)")도
# 실제로 일어난 사고라서, 부분 일치로 봐주면 그 축소를 놓친다.
REQUIRED_SECTIONS = (
    "## 0. 최상위 안전 불변식",
    "## 권위 경계",
    "## Full SDD 도구 계층",
    "### SDD 4계층 앵커",
    "## 역할 분리",
    "## SDD 사이클",
    "### READY 판정",
    "## 코드 증거 절차",
    "## Function Logic Map",
    "## 리뷰 게이트 (등급제)",
    "## 완료 게이트 (자동화)",
    "### PM 계층",
    "## 불변 규칙",
    "## OpenSpec 적용 범위",
    "## 예외 경로",
    "## 위험도 분류",
    "## Pre-Edit 선언",
    "## 완료 보고 금지 조건",
    "## 에이전트 실행 순서",
)

# 절 제목만 보면 제목은 남기고 알맹이만 지우는 편집을 놓친다.
# 그래서 각 단계를 실제로 지탱하는 문구도 함께 본다.
REQUIRED_PHRASES = (
    # 5단계 증거 조정의 산출물 — 2026-08-02 이후 실제로 생산이 멈춘 자리다.
    "analysis/code-context/",
    "evidence-reconciliation.md",
    # 2단계에서 구현 기준 commit 을 고정하는 도구.
    "capture_change_base.py",
    # 6단계 면제는 gate 가 문자열로 검사한다. 형식이 바뀌면 면제가 조용히 죽는다.
    "Function Logic Map: not-applicable",
    # PM 1:1 계약의 fail-closed 항목.
    "bootstrap allowlist",
    # 침묵한 생략 금지 — 건너뛰려면 사유를 남기라는 규칙.
    "not-applicable",
    # 어느 단계가 기계 검사이고 어느 단계가 규율인지 적은 표.
    "단계별 강제 지점",
)


def read_workflow() -> str:
    return WORKFLOW.read_text(encoding="utf-8")


def missing_items(text: str) -> list[str]:
    """정본에서 빠진 제목·문구를 모아 돌려준다. 비어 있으면 통과다."""
    lines = {line.rstrip() for line in text.splitlines()}
    gone = [name for name in REQUIRED_SECTIONS if name not in lines]
    gone += [phrase for phrase in REQUIRED_PHRASES if phrase not in text]
    return gone


class WorkflowContractTest(unittest.TestCase):
    def test_workflow_file_exists(self) -> None:
        self.assertTrue(WORKFLOW.is_file(), f"{WORKFLOW} 가 없다")

    def test_every_required_section_and_phrase_is_present(self) -> None:
        gone = missing_items(read_workflow())
        self.assertEqual(
            [],
            gone,
            "docs/WORKFLOW.md 에서 다음이 사라졌다. 무관한 커밋이 이 파일을 덮지 "
            f"않았는지 `git log -p -- docs/WORKFLOW.md` 로 확인할 것: {gone}",
        )

    def test_guard_notices_a_deleted_section(self) -> None:
        """양성 대조군 — 지우면 정말 잡히는지 각 항목마다 확인한다.

        이걸 안 하면 검사기가 눈이 멀어도 스위트는 초록이다.
        """
        text = read_workflow()
        for name in REQUIRED_SECTIONS:
            with self.subTest(section=name):
                mutated = "\n".join(
                    line for line in text.splitlines() if line.rstrip() != name
                )
                self.assertIn(name, missing_items(mutated))

    def test_guard_notices_a_deleted_phrase(self) -> None:
        text = read_workflow()
        for phrase in REQUIRED_PHRASES:
            with self.subTest(phrase=phrase):
                self.assertIn(phrase, missing_items(text.replace(phrase, "")))


if __name__ == "__main__":
    unittest.main()
