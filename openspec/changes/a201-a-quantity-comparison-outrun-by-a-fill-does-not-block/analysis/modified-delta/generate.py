#!/usr/bin/env python3
"""a201 v3 정본 delta 생성 — ADDED(added.md) + MODIFIED 3(정본 원문에서 exact-string 치환).

단언: 치환마다 그 Requirement 블록 안에서 정확히 1회 적중, 역치환하면 원문과 같음(의도 외 차이 0),
Scenario 이름 집합 보존. 실행: python3 generate.py <repo-root>
"""

import pathlib
import re
import sys

ROOT = pathlib.Path(sys.argv[1])
SPEC = ROOT / "openspec/specs/reconciliation/spec.md"
HERE = pathlib.Path(__file__).resolve().parent
CHANGE = HERE.parent.parent
OUT = CHANGE / "specs/reconciliation/spec.md"

NEW_REQ = "「체결이 조정을 대신한 수량 불일치 차단은 체결 credit 뒤의 재조회 일치로 풀린다」"

EDITS = {
    "불일치 시 진입 차단": [
        (
            "**결속된 주문 전부의 종결 기록 뒤의 재조회 일치**에만 근거하며, 각각 구분되는 release cause"
            "(ADJUSTMENT_APPLIED 계열, ORDER_TERMINAL_CONFIRMED)",
            "**결속된 주문 전부의 종결 기록 뒤의 재조회 일치**, 또는 " + NEW_REQ + " Requirement 가 정한 "
            "**체결 credit 뒤의 재조회 일치**에만 근거하며, 각각 구분되는 release cause"
            "(ADJUSTMENT_APPLIED 계열, ORDER_TERMINAL_CONFIRMED, FILL_APPLIED)",
        ),
        (
            "조정도 종결 증명도 없이 우연히 일치한 단발 관측은",
            "조정도 종결 증명도 체결 credit 도 없이 우연히 일치한 단발 관측은",
        ),
        # V4b(Manager 결정 1, codex v3 r1 V3R1-1): 깨끗한 비교의 외부 포지션 분쟁도 권위 오류 전 반박.
        (
            "그 현재 blocking comparison이 adjustment credit보다 later이고 credited symbol을 계속 disputed로 "
            "보고하면 authority error 반환 전에 credit을 refuted로 만료해야 한다(SHALL).",
            "그 현재 blocking comparison이 adjustment credit보다 later이고 credited symbol을 계속 disputed로 "
            "보고하면 authority error 반환 전에 credit을 refuted로 만료해야 한다(SHALL). 진입을 막지 않는 "
            "comparison이 credited symbol을 외부 포지션으로 disputed로 보고 authority read가 실패해도 같다 — "
            "adjustment credit과 체결 credit 모두 error 반환 전에 refuted로 만료해야 한다(SHALL).",
        ),
    ],
    "조정 credit은 그것을 만든 비교가 아니라 그 다음 재조회가 쓴다": [
        (
            "Requirement 의\n조건을 전부 갖춘 경우뿐이다(SHALL).",
            "Requirement 의\n조건을 전부 갖춘 경우, 또는 " + NEW_REQ
            + " Requirement 의 체결 credit 뒤의 재조회 일치뿐이다(SHALL).",
        ),
        (
            "missing-order 종결 증명 해제의 조건도 갖추지 않은 심볼의",
            "missing-order 종결 증명 해제의 조건도 체결 credit 도 갖추지 않은 심볼의",
        ),
        # V3R1-3(Manager 결정 3): 반박 뒤의 해제가 조정만을 요구한다는 THEN.
        (
            "해제는 새 조정과 그 뒤의 일치를 다시 요구한다",
            "해제는 새 조정 또는 새 체결 credit 과 그 뒤의 일치를 다시 요구한다",
        ),
    ],
    "missing-order 종목 차단은 그것을 만든 주문들의 종결 증명과 그 뒤의 일치로 풀린다": [
        (
            "이 경로로는 해제되지 않고 조정 이벤트 뒤의 재조회 일치만 그 차단을 풀 수 있다",
            "이 경로로는 해제되지 않고 조정 이벤트 또는 체결 credit 뒤의 재조회 일치만 그 차단을 풀 수 있다",
        ),
    ],
}


def blocks(text):
    """Requirement 이름 → 블록 원문(머리줄부터 다음 Requirement 직전까지, 끝 공백 제거)."""
    parts = re.split(r"(?m)^(?=### Requirement: )", text)
    out = {}
    for part in parts:
        m = re.match(r"### Requirement: (.+)\n", part)
        if m:
            out[m.group(1).strip()] = part.rstrip() + "\n"
    return out


def scenarios(block):
    return re.findall(r"(?m)^#### Scenario: (.+)$", block)


def main():
    spec = SPEC.read_text(encoding="utf-8")
    reqs = blocks(spec)
    modified = []
    for name, edits in EDITS.items():
        assert name in reqs, f"requirement not found: {name}"
        original = reqs[name]
        block = original
        for old, new in edits:
            n = block.count(old)
            assert n == 1, f"{name}: expected exactly one hit for {old!r}, got {n}"
            block = block.replace(old, new)
        back = block
        for old, new in reversed(edits):
            assert back.count(new) == 1, f"{name}: reverse hit count for {new!r} != 1"
            back = back.replace(new, old)
        assert back == original, f"{name}: reverse substitution does not restore the original"
        assert scenarios(block) == scenarios(original), f"{name}: scenario names changed"
        modified.append(block)
        print(f"[ok] {name}: {len(edits)} edit(s), scenarios {len(scenarios(block))} kept")
    added = (HERE / "added.md").read_text(encoding="utf-8").rstrip() + "\n"
    OUT.write_text(added + "\n## MODIFIED Requirements\n\n" + "\n".join(modified), encoding="utf-8")
    print(f"[ok] wrote {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
