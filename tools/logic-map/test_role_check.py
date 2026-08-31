#!/usr/bin/env python3
"""role_check 가 좌표의 **역할**을 실제로 보는지 반증한다.

여기 심는 오류는 전부 같은 차원 하나를 바꾼다: 좌표가 가리키는 것이 분기냐
반환이냐 호출이냐. 크기나 개수를 바꾸는 반증은 이 결함이 사는 차원이 아니다 —
개수는 이미 check_analysis 의 "AST branches N" 검사가 보고 있었고, a112 의
오류 넷은 그 검사를 전부 통과했다.

음성 대조도 같은 수만큼 있다. 맞는 문서에서 터지는 검사는 없는 것보다 나쁘고,
저장소에는 손으로 쓴 호출 분석 표가 204개 있다.
"""

from __future__ import annotations

import unittest

from role_check import role_errors

# 분기 둘, 반환 셋, 호출 둘인 함수.
AST = {
    "branches": [
        {"id": "B1", "kind": "if", "at": {"line": 10, "column": 2}},
        {"id": "B2", "kind": "if", "at": {"line": 14, "column": 2}},
    ],
    "returns": [
        {"kind": "return", "at": {"line": 11, "column": 3}},
        {"kind": "return", "at": {"line": 15, "column": 3}},
        {"kind": "return", "at": {"line": 17, "column": 2}},
    ],
    "calls": [
        {"kind": "call", "at": {"line": 12, "column": 9}, "text": "helper"},
        {"kind": "call", "at": {"line": 16, "column": 4}, "text": "other"},
    ],
}

GOOD = """# Function Logic Map: `f`

## Branches and early returns

Exact AST return positions: 11:3, 15:3, 17:2.

| Branch | AST kind | Position | Disposition |
|---|---|---|---|
| B1 | if | 10:2 | covered |
| B2 | if | 14:2 | covered |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `helper` | 12:9 |
| `other` | 16:4 |
"""


def errors(text, ast=AST):
    return role_errors("t", text, ast)


class RoleErrorsCatchMisplacedCoordinates(unittest.TestCase):
    def test_a_correct_map_is_silent(self):
        self.assertEqual(errors(GOOD), [])

    def test_a_return_position_that_is_actually_a_branch_is_caught(self):
        # 17:2 를 분기 좌표 10:2 로 바꾼다. 개수는 그대로 셋이다.
        broken = GOOD.replace("11:3, 15:3, 17:2", "11:3, 15:3, 10:2")
        self.assertTrue(any("return positions" in e for e in errors(broken)), errors(broken))

    def test_a_branch_row_citing_a_return_is_caught(self):
        broken = GOOD.replace("| B2 | if | 14:2 |", "| B2 | if | 15:3 |")
        self.assertTrue(any("branch table cites" in e for e in errors(broken)), errors(broken))

    def test_a_branch_row_citing_a_call_is_caught(self):
        broken = GOOD.replace("| B1 | if | 10:2 |", "| B1 | if | 12:9 |")
        self.assertTrue(any("branch table cites" in e for e in errors(broken)), errors(broken))

    def test_a_call_row_citing_a_branch_is_caught(self):
        broken = GOOD.replace("| `other` | 16:4 |", "| `other` | 14:2 |")
        self.assertTrue(any("call positions" in e for e in errors(broken)), errors(broken))

    def test_a_truncated_enumerated_call_table_is_caught(self):
        broken = GOOD.replace("| `other` | 16:4 |\n", "")
        found = errors(broken)
        self.assertTrue(any("stops early" in e for e in found), found)

    def test_a_branchless_function_may_not_call_a_return_a_branch(self):
        ast = {"branches": None, "returns": AST["returns"], "calls": AST["calls"]}
        text = GOOD.replace("| B1 | if | 10:2 |", "| B1 | body | 11:3 |").replace(
            "| B2 | if | 14:2 |\n", ""
        ).replace("Exact AST return positions: 11:3, 15:3, 17:2.\n", "")
        found = role_errors("t", text, ast)
        self.assertTrue(any("no branches" in e for e in found), found)


class RoleErrorsStaySilentOnHonestMaps(unittest.TestCase):
    """맞는 줄에서 터지지 않는다는 것이 이 검사의 절반이다."""

    def test_a_branchless_body_row_at_the_function_start_is_fine(self):
        ast = {"branches": None, "returns": AST["returns"], "calls": AST["calls"]}
        text = "## Branches and early returns\n\n| B1 | body | 3:1 | ran 6x |\n"
        self.assertEqual(role_errors("t", text, ast), [])

    def test_prose_in_the_kind_column_is_not_compared(self):
        text = GOOD.replace(
            "| B2 | if | 14:2 |", "| B2 | case KindKRNetFlow → KR only | 14:2 |"
        )
        self.assertEqual(errors(text), [])

    def test_a_hand_written_call_analysis_table_is_not_enumerated(self):
        # 저장소에 204개 있는 모양. 한 줄이 호출 넷을 묶고 줄 번호만 적기도 한다.
        text = GOOD.replace(
            "| Callee expression | Position |\n|---|---|\n| `helper` | 12:9 |\n| `other` | 16:4 |",
            "| Callee | Why called | Contract | Evidence |\n|---|---|---|---|\n"
            "| `parseDecimal` (x4: 233:10, 236:12, 244:10, 247:12) | 값 변환 | fail-open | ast.json |",
        )
        self.assertEqual(errors(text), [])

    def test_a_map_that_claims_nothing_is_checked_for_nothing(self):
        self.assertEqual(errors("# Function Logic Map: `f`\n\n산문만 있다.\n"), [])

    def test_an_empty_ast_is_not_a_claim(self):
        self.assertEqual(role_errors("t", GOOD, {}), [])

    def test_a_node_without_a_column_is_not_compared(self):
        # 열 없이 줄만 있는 노드가 실제로 온다(check_analysis 의 고정 픽스처).
        # 없는 값을 0 으로 채워 비교하면 맞는 문서가 틀렸다고 나온다.
        ast = {
            "branches": [{"id": "B1", "kind": "if", "at": {"line": 10}}],
            "returns": None,
            "calls": None,
        }
        text = "## Branches and early returns\n\n| B1 | if | 10:2 | ok |\n"
        self.assertEqual(role_errors("t", text, ast), [])


if __name__ == "__main__":
    unittest.main()
