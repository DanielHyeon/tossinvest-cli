#!/usr/bin/env python3
"""role_check 가 좌표의 **역할**을 실제로 보는지 반증한다.

여기 심는 오류는 전부 같은 차원 하나를 바꾼다: 좌표가 가리키는 것이 분기냐
반환이냐 호출이냐. 크기나 개수를 바꾸는 반증은 이 결함이 사는 차원이 아니다 —
개수는 이미 check_analysis 의 "AST branches N" 검사가 보고 있었고, a112 의
오류 넷은 그 검사를 전부 통과했다.

음성 대조도 같은 수만큼 있다. 맞는 문서에서 터지는 검사는 없는 것보다 나쁘고,
저장소에는 손으로 쓴 호출 분석 표가 226개 있다(그중 204개가 같은 헤더 철자를
쓴다 — 앞 판본은 그 최빈 철자 수를 전체 수로 적었다).

**4차 적대 리뷰가 이 파일의 범위를 반증했다.** 앞 판본은 비아카이브 404개에만
돌려 "오탐 0"을 확인했고, 결함이 사는 차원인 산문의 다양함은 아카이브 2610개에
있었다 — 대비율 `4.5:1`, 장 시작 `09:00`, 주소 `127.0.0.1:0` 셋이 좌표로 읽혔다.
아래 음성 대조에 그 셋을 그대로 심는다.

**같은 리뷰가 이 파일의 두 번째 범위도 반증했다.** 완전성 표지를 단 표 122 개 중
83 개만 실제로 대조되고 39 개는 좌표를 안 적어 조용히 빠져나갔다. 표지를 저자가
고르므로 그것은 감사 여부를 저자가 정하는 것이었다. 아래 두 클래스가 그 문이
닫혔다는 것과, 닫으면서 맞는 문서를 깨지 않았다는 것을 함께 심는다.
"""

from __future__ import annotations

import unittest

import role_check
from role_check import call_enumeration_in_use, role_errors

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


def errors(text, ast=AST, branch_map="", require_calls=False):
    return role_errors("t", text, ast, branch_map, require_calls)


def in_use(*texts, ast=AST):
    """강제 판정은 (본문, ast 값) 쌍을 본다 — 표지 철자가 아니라 표의 내용이다."""
    return call_enumeration_in_use((text, ast) for text in texts)


ENUMERATION = (
    "| Callee expression | Position |\n|---|---|\n| `helper` | 12:9 |\n| `other` | 16:4 |"
)


# branch-test-map 은 같은 분기 좌표를 다른 열에 다시 적는다.
BRANCH_MAP_GOOD = """# Branch Test Map: `f`

| Branch | Scenario anchor | Disposition | RED | GREEN |
|---|---|---|---|---|
| B1 | if at 10:2 | covered | yes | yes |
| B2 | if at 14:2 | covered | yes | yes |
"""


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


class RoleErrorsReadTheBranchTestMapToo(unittest.TestCase):
    """branch-test-map 의 앵커 좌표도 분기여야 한다.

    앞 판본은 function-logic-map 만 받았다. 그래서 ast.json 과 FLM 을
    재기준화하면서 branch-test-map 네 벌 38 줄을 옛 줄번호로 남겼는데 게이트가
    조용히 통과시켰다 — 같은 사실을 두 파일에 적으면서 한 파일만 본 것이다."""

    def test_a_correct_branch_test_map_is_silent(self):
        self.assertEqual(errors(GOOD, branch_map=BRANCH_MAP_GOOD), [])

    def test_a_stale_anchor_is_caught(self):
        stale = BRANCH_MAP_GOOD.replace("if at 10:2", "if at 8:2")
        found = errors(GOOD, branch_map=stale)
        self.assertEqual(len(found), 1, found)
        self.assertIn("branch test map", found[0])

    def test_an_anchor_carrying_trailing_prose_is_still_read(self):
        annotated = BRANCH_MAP_GOOD.replace(
            "if at 10:2", "if at 10:2 — the key token cannot be read"
        )
        self.assertEqual(errors(GOOD, branch_map=annotated), [])

    def test_a_branch_test_map_that_cites_nothing_is_not_a_claim(self):
        bare = """# Branch Test Map: `f`

| Branch | Scenario anchor | Disposition |
|---|---|---|
| B1 | the ledger refuses the claim | covered |
| B2 | the quote is older than the cap | covered |
"""
        self.assertEqual(errors(GOOD, branch_map=bare), [])


class RoleErrorsDoNotReadProseAsACoordinate(unittest.TestCase):
    """행 아무 데서나 숫자쌍을 집으면 맞는 문서가 게이트에서 터진다.

    아래 셋은 지어낸 것이 아니다. 4차 적대 리뷰가 저장소 전체 3014 번들에서
    찾아낸 실제 오탐이고, 앞 판본은 이 셋을 전부 좌표로 읽었다."""

    def test_a_contrast_ratio_is_not_a_coordinate(self):
        row = "| B1 | if | 10:2 | computed contrast is below 4.5:1 |"
        self.assertEqual(errors(GOOD.replace("| B1 | if | 10:2 | covered |", row)), [])

    def test_a_clock_time_is_not_a_coordinate(self):
        row = "| B1 | shifted from fixed 09:00 KST | 10:2 | covered |"
        self.assertEqual(errors(GOOD.replace("| B1 | if | 10:2 | covered |", row)), [])

    def test_a_socket_address_is_not_a_coordinate(self):
        row = '| B1 | `net.Listen("tcp4", "127.0.0.1:0") != nil` | 10:2 | covered |'
        self.assertEqual(errors(GOOD.replace("| B1 | if | 10:2 | covered |", row)), [])

    def test_prose_cannot_hide_a_wrong_coordinate(self):
        """오탐을 없애느라 진짜를 놓치면 안 된다."""
        row = "| B1 | shifted from fixed 09:00 KST | 11:3 | covered |"
        found = errors(GOOD.replace("| B1 | if | 10:2 | covered |", row))
        self.assertEqual(len(found), 1, found)
        self.assertIn("branch table", found[0])




class ACallTableCannotOptOutOfBeingChecked(unittest.TestCase):
    """표지를 달고 좌표를 빼는 것은 면제가 아니라 오류다.

    앞 판본은 행 하나라도 좌표가 없으면 표 전체를 조용히 건너뛰었다. 4차 적대
    리뷰가 재 보니 표지를 단 122개 중 **39개**가 그 문으로 나가 있었다 — 즉
    감사를 받을지를 문서 저자가 정하고 있었다. 여기 심는 반증은 그 문이 닫혔다는
    것과, 닫으면서 **맞는 문서를 깨지 않았다**는 것 둘 다이다."""

    def test_a_row_without_a_coordinate_is_now_an_error(self):
        text = GOOD.replace("| `other` | 16:4 |", "| `other` | 16 |")
        found = errors(text)
        self.assertEqual(len(found), 1, found)
        self.assertIn("carry no coordinate cell", found[0])

    def test_a_table_with_no_coordinate_at_all_is_caught_when_calls_exist(self):
        text = GOOD.replace(
            "| `helper` | 12:9 |\n| `other` | 16:4 |",
            "| (no call expressions in this function) | — |",
        )
        found = errors(text)
        self.assertEqual(len(found), 1, found)
        self.assertIn("cites no call position", found[0])

    def test_a_function_with_no_calls_may_say_so_without_a_coordinate(self):
        ast = dict(AST, calls=[])
        text = GOOD.replace(
            "| `helper` | 12:9 |\n| `other` | 16:4 |",
            "| (no call expressions in this function) | — |",
        )
        self.assertEqual(errors(text, ast=ast), [])

    def test_a_note_table_below_the_enumeration_does_not_pad_the_count(self):
        # 열거를 잘라 놓고 아래에 주석 표를 붙이면, 섹션의 `|` 줄을 다 모으는
        # 판본에서는 개수가 채워져 통과한다. 첫 표 한 벌만 읽으므로 터진다.
        text = GOOD.replace(
            ENUMERATION,
            "| Callee expression | Position |\n|---|---|\n| `helper` | 12:9 |\n"
            "\n### 손으로 쓴 주석\n\n"
            "| Callee (hand-written note) | Why called |\n|---|---|\n| `other` | 16 줄 |",
        )
        found = errors(text)
        self.assertEqual(len(found), 1, found)
        self.assertIn("enumerates 1 call(s) but ast.json has 2", found[0])


class TheEnumerationIsRequiredPerChangeNotPerTable(unittest.TestCase):
    """표지를 달지 말지를 번들마다 고를 수 있으면 그것이 곧 면제다.

    한 change 가 어디서든 열거형 표를 쓰면 그 change 의 모든 번들이 써야 한다.
    그래야 새 번들이 표를 빼고 들어오는 길이 막힌다."""

    HAND_WRITTEN = (
        "| Callee | Why called | Evidence |\n|---|---|---|\n"
        "| `helper` | 값 변환 | ast.json call at 12 |"
    )

    def test_a_missing_enumeration_is_an_error_when_the_change_uses_one(self):
        text = GOOD.replace(ENUMERATION, self.HAND_WRITTEN)
        found = errors(text, require_calls=True)
        self.assertEqual(len(found), 1, found)
        self.assertIn("must open with a", found[0])

    def test_a_change_that_never_enumerates_is_asked_for_nothing(self):
        text = GOOD.replace(ENUMERATION, self.HAND_WRITTEN)
        self.assertEqual(errors(text, require_calls=False), [])

    def test_the_mandate_is_measured_from_the_change_not_declared(self):
        hand = GOOD.replace(ENUMERATION, self.HAND_WRITTEN)
        self.assertFalse(in_use(hand, hand))
        self.assertTrue(in_use(hand, GOOD))

    def test_a_note_table_alone_does_not_turn_the_mandate_on(self):
        # 주석 표는 표지를 떼고 남는다. 그것이 다시 강제를 켜면, 강제를 끄려고
        # 주석까지 지우게 된다.
        text = GOOD.replace(
            ENUMERATION,
            self.HAND_WRITTEN + "\n\n### 주석\n\n| Callee (hand-written note) | Position |\n"
            "|---|---|\n| `helper` | 12 |",
        )
        self.assertFalse(in_use(text))


class TheCalleeColumnIsComparedNotJustItsCoordinate(unittest.TestCase):
    """완전성을 주장하는 열은 좌표 열이 아니라 **철자 열**이다.

    5차 적대 리뷰가 짚었다: 앞 판본은 좌표만 대조했으므로 행이 `os.Exit` 을
    적어도 조용했다. 좌표만 보는 검사는 "그 자리에 호출이 있다"까지만 말하고,
    그 행이 그 호출을 **가리킨다**는 것은 말하지 않는다 — 존재 검사와 역할
    검사의 그 차이가 이 change 가 다섯 라운드 동안 고쳐 온 결함이다."""

    def test_a_row_that_names_the_wrong_callee_is_caught(self):
        text = GOOD.replace("| `helper` | 12:9 |", "| `os.Exit` | 12:9 |")
        found = errors(text)
        self.assertEqual(len(found), 1, found)
        self.assertIn("names `os.Exit` but ast.json calls `helper`", found[0])

    def test_a_correct_table_is_not_broken_by_the_name_check(self):
        self.assertEqual(errors(GOOD), [])

    def test_a_call_node_without_text_reads_as_unnamed(self):
        # 저장소에 철자가 없는 호출 노드가 15개 있고 기존 문서는 `(unnamed)` 로
        # 적는다. 그 관례를 깨면 맞는 문서 15군데가 터진다.
        ast = dict(AST, calls=[
            {"kind": "call", "at": {"line": 12, "column": 9}},
            {"kind": "call", "at": {"line": 16, "column": 4}, "text": "other"},
        ])
        text = GOOD.replace("| `helper` | 12:9 |", "| `(unnamed)` | 12:9 |")
        self.assertEqual(errors(text, ast=ast), [])
        self.assertEqual(len(errors(GOOD, ast=ast)), 1)


class TheMandateIsNotTurnedOffByALineAboveTheEnumeration(unittest.TestCase):
    """표지를 **첫 표의 첫 줄에서만** 찾으면 한 줄로 강제를 끌 수 있다.

    5차 적대 리뷰가 133 번들 전부의 열거 위에 `|` 줄 하나씩을 얹어 강제 판정을
    `False` 로 만들었다. 표지는 그대로 서 있는데 감사만 꺼진다 — 표지를 세어
    "몇 개가 감사받는가"를 말하던 근거가 그 순간 무너진다."""

    ABOVE = "| note | 참고 |\n\n" + ENUMERATION

    def test_a_pipe_line_above_the_enumeration_does_not_turn_the_mandate_off(self):
        self.assertTrue(in_use(GOOD.replace(ENUMERATION, self.ABOVE)))

    def test_the_enumeration_must_still_open_the_section(self):
        # 강제는 켜지고, 그 번들은 "첫 표가 열거가 아니다"로 터진다. 두 규칙이
        # 서로 다른 자리에 있어야 한 줄 얹기가 어느 쪽으로도 이득이 아니다.
        found = errors(GOOD.replace(ENUMERATION, self.ABOVE), require_calls=True)
        self.assertEqual(len(found), 1, found)
        self.assertIn("must open with a", found[0])


class TheMandateIsAPropertyOfTheTableNotOfItsHeading(unittest.TestCase):
    """강제 판정이 머리글 철자면 한 글자로 꺼진다.

    6차 적대 리뷰가 공백 하나·앞 들여쓰기·굵게 셋으로 껐다. 셋 다 화면에서
    똑같이 보인다 — 보이지 않는 회피가 가장 나쁘다. 이제 판정은 표가 실제로
    호출 좌표를 빠짐없이 같은 순서로 적고 있는지를 본다. 남는 회피는 열거를
    정말 그만두는 것뿐이고, 그것은 숨길 수가 없다."""

    RENAMED = ENUMERATION.replace(
        "| Callee expression | Position |", "| Calls | Where |")

    def test_a_renamed_heading_still_turns_the_mandate_on(self):
        self.assertTrue(in_use(GOOD.replace(ENUMERATION, self.RENAMED)))

    def test_a_renamed_heading_then_fails_the_bundle_it_belongs_to(self):
        # 판정은 켜지고, 그 번들은 "첫 표가 열거가 아니다"로 터진다. 두 요구가
        # 서로 다른 자리에 있어야 이름 바꾸기가 어느 쪽으로도 이득이 아니다.
        found = errors(GOOD.replace(ENUMERATION, self.RENAMED), require_calls=True)
        self.assertEqual(len(found), 1, found)
        self.assertIn("must open with a", found[0])

    def test_moving_the_table_to_another_section_does_not_turn_it_off(self):
        # 7차 적대 리뷰: 절 이름에 판정을 걸면 표를 옮기는 것으로 꺼진다.
        # 표지도 행도 그대로 남으므로 완전성 주장은 멀쩡한 채 감사만 꺼졌다.
        moved = GOOD.replace(ENUMERATION, "") + "\n## 다른 절\n\n" + ENUMERATION + "\n"
        self.assertTrue(in_use(moved))

    def test_giving_up_the_enumeration_is_the_only_way_out_and_it_shows(self):
        # 좌표를 지우면 판정이 꺼진다. 그것은 의도한 잔여물이다 — 대신 열거가
        # 사라지므로 완전성 주장도 함께 사라진다.
        gone = GOOD.replace("| `helper` | 12:9 |\n| `other` | 16:4 |",
                            "| `helper` | 12 줄 |\n| `other` | 16 줄 |")
        self.assertFalse(in_use(gone))


class AHeadingThatLooksTheSameIsTreatedTheSame(unittest.TestCase):
    """화면에서 같아 보이는 머리글은 같게 읽는다.

    공백 둘·앞 들여쓰기·굵게는 마크다운이 똑같이 그린다. 그것들을 "표지가 아님"
    으로 읽으면 저자가 아무것도 안 바꾼 것처럼 보이면서 감사를 끌 수 있다."""

    def _truncated(self, header):
        return GOOD.replace(
            ENUMERATION, f"{header}\n|---|---|\n| `helper` | 12:9 |")

    def test_a_double_space_in_the_heading_is_still_the_marker(self):
        found = errors(self._truncated("| Callee  expression | Position |"))
        self.assertEqual(len(found), 1, found)
        self.assertIn("enumerates 1 call(s) but ast.json has 2", found[0])

    def test_a_bold_heading_is_still_the_marker(self):
        found = errors(self._truncated("| **Callee expression** | Position |"))
        self.assertEqual(len(found), 1, found)
        self.assertIn("enumerates 1 call(s) but ast.json has 2", found[0])

    def test_an_indented_table_is_still_read(self):
        found = errors(self._truncated("  | Callee expression | Position |"))
        self.assertEqual(len(found), 1, found)
        self.assertIn("enumerates 1 call(s) but ast.json has 2", found[0])

    def test_the_demoted_note_heading_is_still_not_the_marker(self):
        # 손으로 쓴 주석 표는 표지를 떼고 남는다. 정규화가 그것까지 삼키면
        # 주석이 완전성 주장이 되어 버린다.
        self.assertFalse(role_check.enumeration_header(
            "| Callee (hand-written note) | Source location |"))


class TheNodeListsMustAgreeAboutMalformedNodes(unittest.TestCase):
    """좌표 목록과 철자 목록이 다른 규칙으로 노드를 세면 **맞는 문서**가 터진다.

    6차 적대 리뷰가 dict 아닌 노드 하나를 ast.json 에 넣어 보였다: `_coords` 는
    건너뛰고 `_call_names` 는 자리를 채워, 뒤 행이 전부 한 칸씩 밀렸다."""

    def test_a_malformed_node_does_not_shift_the_name_list(self):
        for ghost in (
            "not-a-dict",
            {"kind": "call", "text": "ghost"},                   # at 이 없다
            {"kind": "call", "text": "ghost", "at": "nope"},      # at 이 dict 가 아니다
            {"kind": "call", "text": "ghost", "at": {"line": 9}},  # column 이 없다
        ):
            with self.subTest(ghost=ghost):
                ast = dict(AST, calls=[ghost] + AST["calls"])
                self.assertEqual(errors(GOOD, ast=ast), [])


if __name__ == "__main__":
    unittest.main()
