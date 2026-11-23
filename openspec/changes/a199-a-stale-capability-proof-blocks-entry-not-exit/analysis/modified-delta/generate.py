#!/usr/bin/env python3
"""a199 MODIFIED delta 생성기 — 정본 Requirement 전문에 exact-string 치환을 적용함.

규칙(analysis/spec-delta-intent.md 머리):
- 정본은 base-commit.txt 의 커밋에서 읽음(없으면 HEAD) — 정본이 아카이브로 바뀐 뒤에도 같은 결과를 재현하기 위함.
- 치환마다 정확히 1회 적중을 단언함.
- 역치환(새 문자열 → 옛 문자열, 역순)으로 정본 전문이 바이트 그대로 돌아오는지 단언함 — 의도 외 차이 0.
- Scenario 이름은 정본의 것을 전부 남기는지 단언함.

사용: python3 generate.py [--check]   (--check 는 쓰지 않고 디스크의 spec.md 와 비교만 함)
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
CHANGE_DIR = HERE.parent.parent
OUT = CHANGE_DIR / "specs" / "engine-safety" / "spec.md"
SPEC_PATH = "openspec/specs/engine-safety/spec.md"


def git(*args: str) -> str:
    return subprocess.run(["git", "-C", str(HERE), *args], check=True, capture_output=True, text=True).stdout


def canonical_text() -> str:
    base_file = CHANGE_DIR / "base-commit.txt"
    rev = base_file.read_text(encoding="utf-8").strip() if base_file.is_file() else "HEAD"
    return git("show", f"{rev}:{SPEC_PATH}")


def requirement_block(text: str, title: str) -> str:
    """「### Requirement: <title>」부터 다음 ### Requirement 또는 ## 머리 직전까지(끝 공백 제거)."""
    head = f"### Requirement: {title}\n"
    assert text.count(head) == 1, f"Requirement 머리 적중 {text.count(head)}회: {title}"
    start = text.index(head)
    nxt = re.compile(r"^(### Requirement: |## )", re.M)
    m = nxt.search(text, start + len(head))
    end = m.start() if m else len(text)
    return text[start:end].rstrip("\n") + "\n"


def scenario_names(block: str) -> list[str]:
    return re.findall(r"^#### Scenario: (.+)$", block, re.M)


# ── A. 「자동화 게이트 기동 인터록」 ──────────────────────────────────────────

A_TITLE = "자동화 게이트 기동 인터록"

A2_PARAGRAPH = """**capability attestation 의 낡음(만료 · 필요 endpoint 누락)은 기동 조건이 아니라 진입 허가 조건이다**(SHALL).
근거는 보호 배선과 같은 실패의 비대칭이다 — 증명이 낡은 경로로 진입하면 손절 없는 새 포지션이 생기지만,
그 경로로 청산을 시도하다 실패하면 남는 것은 런타임을 거부했을 때와 같은 상태다.
attestation 이 읽히고 온전하며 해석된 계좌를 가리키는데 낡았으면, 그리고 나머지 조항이 전부 충족되면:

- 런타임은 **청산 전용으로 기동한다**(SHALL). 루프 집합은 변하지 않는다.
- 노출을 **증가**시키는 mutation 은 거부된다(SHALL). 집행은 보호 배선과 같은 mutation chokepoint 한 곳이고
  판정은 mutation 의 형태에서 계산한다.
- 이 진입 차단은 매 기동마다 attestation 에서 다시 유도된다(SHALL) — 재시작으로 풀리지 않고,
  attestation 이 유효해진 뒤의 기동에서는 사람의 해제 절차 없이 사라진다.
- 낡음 밖의 attestation 결함(없음 · 읽을 수 없음 · 불완전 · 알 수 없는 형식 · 계좌 불일치)은 기동을 거부한다(SHALL).
  만료된 attestation 의 계좌 일치도 따로 확인해야 한다(SHALL — 만료가 계좌 불일치를 가려서는 안 된다).
- 낡음과 다른 조항의 결함이 함께 있으면 기동을 거부한다(SHALL).
- 청산 전용 기동은 수락도 거부도 아닌 결과로 audit 에 기록되고, 운영자에게 기동 시점에 전달된다(SHALL —
  읽는 문장이며 타이핑 확인·추가 승인 마찰이어서는 안 된다(SHALL NOT)).
- 실행 중인 엔진은 attestation 이 낡아져도 동작이 바뀌지 않는다(SHALL NOT) — 판정은 기동 시점뿐이다."""

A_LAST_SCENARIO = """#### Scenario: 정수가 아닌 최대 보유 포지션 수

- **WHEN** 게이트 ON 설정의 `max_open_positions` 가 음수이거나 정수가 아니면
- **THEN** 인터록이 기동을 거부하고 게이트 OFF 설정의 읽기는 영향을 받지 않는다
"""

A_NEW_SCENARIOS = """
#### Scenario: 필요 endpoint 가 빠진 attestation 으로 기동
- **WHEN** 게이트 ON 이고 attestation 이 만료 전이며 계좌가 일치하지만 주문 접수 endpoint 를 덮지 않고, 나머지 조항은 충족된 채 엔진이 기동한다
- **THEN** 런타임은 청산 전용으로 기동하고 매수는 거부되며 보유의 손절 매도는 제출된다

#### Scenario: 만료됐고 다른 계좌를 가리키는 attestation
- **WHEN** attestation 이 만료됐고 그 계좌가 자격 증명이 해석한 계좌와 다르다
- **THEN** 기동은 거부된다

#### Scenario: 낡은 attestation 과 한도 출처 불일치
- **WHEN** attestation 이 낡았고 Guardian 의 한도가 인터록이 감사한 한도와 다르다
- **THEN** 기동은 한도 출처 불일치로 거부된다

#### Scenario: attestation 을 갱신한 뒤의 재기동
- **WHEN** 청산 전용으로 돌던 엔진을 유효한 attestation 으로 다시 기동한다
- **THEN** 진입 차단은 사람의 해제 절차 없이 사라지고 audit 에는 수락이 기록된다

#### Scenario: 낡은 채 재기동
- **WHEN** 청산 전용으로 돌던 엔진을 attestation 을 고치지 않고 다시 기동한다
- **THEN** 런타임은 다시 청산 전용으로 기동한다
"""

A_EDITS: list[tuple[str, str, str]] = [
    (
        "A1 조항 2",
        "2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치.",
        "2. capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성)이 존재하고 읽히며 온전하고 계좌가 일치. 만료와 endpoint 누락(낡음)은 기동을 거부하지 않고 아래 「낡은 증명」 절이 정한다.",
    ),
    (
        "A2 낡은 증명 절 삽입",
        "게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).\n\n**브로커측",
        "게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).\n\n" + A2_PARAGRAPH + "\n\n**브로커측",
    ),
    (
        "A3 만료 Scenario WHEN",
        "- **WHEN** 게이트 ON + attestation 만료 상태로 기동하면\n",
        "- **WHEN** 게이트 ON + attestation 만료 상태이고 그 attestation 이 해석된 계좌를 가리킬 때 기동하면\n",
    ),
    (
        "A3 만료 Scenario THEN",
        "- **THEN** 기동이 거부되고 재검증 안내가 출력된다\n",
        "- **THEN** 런타임은 청산 전용으로 기동하고, 노출 증가 mutation 은 거부되며, exit 관측·대사·체결 감지는 돈다\n",
    ),
    (
        "A 추가 Scenario 5(Requirement 끝)",
        A_LAST_SCENARIO,
        A_LAST_SCENARIO + A_NEW_SCENARIOS,
    ),
]

# ── B. 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 ─────────

B_TITLE = "capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다"

B_LAST_SCENARIO_THEN_OLD = "- **THEN** 제목과 본문의 첫 문장은 지금 엔진을 재시작하면 기동이 거부된다는 것을 말하고, 본문은 빠진 endpoint 의 이름과 자동 손절이 돌지 않게 된다는 것과 감독하 실주문 검증의 재실행을 말한다\n"
B_LAST_SCENARIO_THEN_NEW = "- **THEN** 제목과 본문의 첫 문장은 지금 엔진을 재시작하면 새 진입이 막힌 청산 전용으로 뜬다는 것을 말하고, 본문은 빠진 endpoint 의 이름과 손절·청산은 돌고 새 진입만 막힌다는 것과 감독하 실주문 검증의 재실행을 말한다\n"

B_NEW_SCENARIOS = """
#### Scenario: 청산 전용으로 돌고 있는 엔진
- **WHEN** 낡은 attestation 으로 청산 전용 기동한 엔진에서 판정이 돈다
- **THEN** critical 알림은 엔진이 지금 청산 전용으로 돌고 있고 새 진입이 막혀 있음을 현재의 사실로 말한다

#### Scenario: 거부가 유지되는 결함
- **WHEN** 실행 중인 엔진이 읽을 attestation 이 다른 계좌를 가리킨다
- **THEN** critical 알림은 지금 재시작하면 기동이 거부되고 그동안 자동 손절이 돌지 않는다고 말한다
"""

B_EDITS: list[tuple[str, str, str]] = [
    ("B1 첫 문단의 질문", "「지금 재시작하면 무장하는가」", "「지금 재시작하면 무엇으로 뜨는가(정상 · 청산 전용 · 거부)」"),
    (
        "B2 판정 동일성 문장 잇기",
        "판정 규칙을 인터록과 따로 두어서는 안 된다 (SHALL NOT).\n",
        "판정 규칙을 인터록과 따로 두어서는 안 된다 (SHALL NOT). 어느 결함이 기동을 거부하고 어느 결함이 청산 전용 기동이 되는지도 같다 (SHALL).\n",
    ),
    (
        "B3-1 기한의 이름(「재시작이 거절되기 시작할 시각」)",
        "**재시작이 거절되기 시작할 시각(기한)**",
        "**재시작의 결과가 정상이 아니게 되는 시각(기한)**",
    ),
    (
        "B2 귀결 문장(B3-2 「재시작이 거절될 상태의 알림」 포함)",
        "재시작이 거절될 상태의 알림은 제목과 본문의 첫 문장에서 그 결과 — 지금 엔진을 재시작하면 기동이 거부된다 — 를 조건문이 아닌 사실로 말해야 하고 (SHALL), 본문에서 그동안 자동 손절이 돌지 않는다는 것과 운영자가 할 일을 말해야 한다 (SHALL).",
        "재시작의 결과가 정상이 아닌 상태의 알림은 제목과 본문의 첫 문장에서 그 결과를 조건문이 아닌 사실로 말해야 한다 (SHALL) — 낡음(만료 · 범위 상실)이면 「지금 엔진을 재시작하면 새 진입이 막힌 청산 전용으로 뜬다」, 그 밖(다른 계좌 · 없음 · 읽을 수 없음 · 불완전)이면 「지금 엔진을 재시작하면 기동이 거부된다」. 거부되는 상태의 본문은 그동안 자동 손절이 돌지 않는다는 것을, 청산 전용인 상태의 본문은 손절·청산은 돌고 새 진입만 막힌다는 것을 말해야 하며 (SHALL), 둘 다 운영자가 할 일을 말해야 한다 (SHALL). 엔진이 이미 청산 전용으로 돌고 있으면 알림은 그것을 현재의 사실로 말해야 한다 (SHALL).",
    ),
    (
        "B3-3 「재시작이 거절될 상태(…)는 고쳐질 때까지」",
        "재시작이 거절될 상태(만료 · 범위 상실 · 계좌 불일치 · 무효)는 고쳐질 때까지",
        "재시작의 결과가 정상이 아닌 상태(만료 · 범위 상실 · 계좌 불일치 · 무효)는 고쳐질 때까지",
    ),
    ("B3-4 「거절 상태가 같은 한」", "거절 상태가 같은 한(", "그 상태가 같은 한("),
    ("B3-5 「다시 거절 상태가 되면」", "다시 거절 상태가 되면", "다시 그 상태가 되면"),
    ("B3-6 Scenario 「범위 상실 알림은 결과와 할 일을 말한다」 THEN", B_LAST_SCENARIO_THEN_OLD, B_LAST_SCENARIO_THEN_NEW),
    ("B 추가 Scenario 2(Requirement 끝)", B_LAST_SCENARIO_THEN_NEW, B_LAST_SCENARIO_THEN_NEW + B_NEW_SCENARIOS),
]


def apply(block: str, edits: list[tuple[str, str, str]], report: list[str]) -> str:
    out = block
    for name, old, new in edits:
        hits = out.count(old)
        assert hits == 1, f"{name}: 적중 {hits}회(1회여야 함)"
        out = out.replace(old, new)
        report.append(f"{name}: 적중 1")
    # 역치환 — 새 문자열이 결과에 정확히 1회 있어야 되돌림이 유일하다
    back = out
    for name, old, new in reversed(edits):
        hits = back.count(new)
        assert hits == 1, f"{name}: 역치환 적중 {hits}회"
        back = back.replace(new, old)
    assert back == block, "역치환 결과가 정본과 다름 — 의도 외 차이 있음"
    report.append("역치환: 정본과 바이트 동일(의도 외 차이 0)")
    missing = [s for s in scenario_names(block) if s not in scenario_names(out)]
    assert not missing, f"정본 Scenario 이름 소실: {missing}"
    report.append(f"Scenario: 정본 {len(scenario_names(block))} 유지, 결과 {len(scenario_names(out))}")
    return out


def build() -> tuple[str, list[str]]:
    text = canonical_text()
    report: list[str] = []
    report.append(f"[A] {A_TITLE}")
    a = apply(requirement_block(text, A_TITLE), A_EDITS, report)
    report.append(f"[B] {B_TITLE}")
    b = apply(requirement_block(text, B_TITLE), B_EDITS, report)
    body = "## MODIFIED Requirements\n\n" + a + "\n" + b
    return body, report


def main() -> int:
    body, report = build()
    print("\n".join(report))
    if "--check" in sys.argv[1:]:
        same = OUT.read_text(encoding="utf-8") == body
        print("디스크 spec.md 일치" if same else "디스크 spec.md 불일치")
        return 0 if same else 1
    OUT.write_text(body, encoding="utf-8")
    print(f"썼음: {OUT.relative_to(CHANGE_DIR)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
