#!/usr/bin/env python3
"""a194 MODIFIED delta 생성기 — 정본 Requirement 블록을 떠서 정확한 문자열 치환만 함(Manager 결정 2026-11-23).

손으로 옮겨 적지 않음: 각 치환은 원문이 블록 안에 정확히 1 회 나와야 하고, 생성 뒤 정본 대조 diff 를 남김.
실행: 저장소 루트에서 `python3 openspec/changes/<id>/analysis/spec-delta/make_modified.py`.
"""
import difflib
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parents[5]
CHANGE = pathlib.Path(__file__).resolve().parents[2]
OUT = pathlib.Path(__file__).resolve().parent


def block(spec: pathlib.Path, header: str) -> str:
    """정본에서 `### Requirement: <header>` 부터 다음 `### Requirement:` 앞까지를 그대로 떠옴."""
    lines = spec.read_text(encoding="utf-8").splitlines(keepends=True)
    want = f"### Requirement: {header}\n"
    starts = [i for i, l in enumerate(lines) if l == want]
    assert len(starts) == 1, f"{spec}: header {header!r} found {len(starts)} times"
    s = starts[0]
    e = next((i for i in range(s + 1, len(lines)) if lines[i].startswith("### Requirement:")), len(lines))
    # 블록 끝의 빈 줄과 다음 절 머리(## ...)는 블록이 아님
    while e > s and (lines[e - 1].strip() == "" or lines[e - 1].startswith("## ")):
        e -= 1
    return "".join(lines[s:e])


def apply(text: str, edits: list[tuple[str, str]]) -> str:
    for old, new in edits:
        n = text.count(old)
        assert n == 1, f"edit target occurs {n} times: {old[:60]!r}"
        text = text.replace(old, new)
    return text


EXIT_HEADER = "외부 취득 포지션의 자동 편입"
EXIT_EDITS = [
    (
        "전이 상태(RECONCILE 대기·인터록 미검증·Stabiliser 미수렴)만 무알림이다.",
        "전이 상태(RECONCILE 대기·인터록 미검증·Stabiliser 미수렴)만 무알림이다. "
        "RECONCILE 대기가 전이 상태인 범위는 「대사 차단으로 편입되지 못한 보유의 침묵은 전이 상태에 한한다」가 정한다.",
    ),
    (
        "- **WHEN** 심볼이 RECONCILE 상태일 때 무결정 보유가 관측되면",
        "- **WHEN** 심볼이 스스로 풀리는 RECONCILE 상태이고 유예(10분) 안일 때 무결정 보유가 관측되면",
    ),
]

SAFETY_HEADER = "무관리 보유 보고의 등급은 사실이 정한다"
SAFETY_COND = (
    "그 시도가 실패한 보유**는 critical 등급으로 보고해야 한다(SHALL — include 경유 편입은 정본 exit-policy에서 "
    "알림 규칙 전부가 enabled 경유와 같다)."
)
SAFETY_EDITS = [
    (
        SAFETY_COND,
        SAFETY_COND
        + " **알림이 켜진 엔진에서**, 같은 편입 후보(편입이 켜져 있거나 그 종목이 편입 지정되었고 제외 목록에 없는 무결정 보유)가 "
        "**전이 상태가 아닌 대사 차단**(운영자만 풀 수 있는 차단, 또는 같은 보유가 연속 10분 이상 편입 판정에서 빠진 스스로 풀리는 차단 — "
        "정본 exit-policy 「대사 차단으로 편입되지 못한 보유의 침묵은 전이 상태에 한한다」) 때문에 편입되지 못한 보고도 critical 등급이어야 한다"
        "(SHALL — 시도는 없었지만 엔진이 보호하기로 한 보유가 보호되지 않았다는 같은 사실이다. 본문은 그 차단을 관측한 대사 주기의 시각의 사건으로 적는다).",
    ),
]
SAFETY_EDITS.append((
    "critical 무관리 보고의 본문은 **편입 시도가 실패한 시각의 사건**으로 적어야 한다(SHALL — 그 시각을 문장에 담는다. 보고를 만든 순간이 아니라 실패를 관측한 순간이다).",
    "critical 무관리 보고의 본문은 편입 시도가 실패한 보고면 **편입 시도가 실패한 시각의 사건**으로, 대사 차단으로 편입되지 못한 보고면 **그 차단을 관측한 시각의 사건**으로 적어야 한다(SHALL — 그 시각을 문장에 담는다. 보고를 만든 순간이 아니라 실패 또는 차단을 관측한 순간이다).",
))  # E5 — Manager 결정 2026-11-23(R2-3(a))
SAFETY_SCENARIO = (
    "\n\n#### Scenario: 운영자만 풀 수 있는 대사 차단 아래의 편입 후보\n"
    "- **WHEN** `notifications.enabled`와 `adoption.enabled`가 모두 참인 엔진에서 귀속 실패 차단이 활성인 종목의 무결정 보유가 관측되면\n"
    "- **THEN** 편입 시도 없이 그 무관리 보고는 관측 시각을 담은 critical 등급으로 durable outbox에 기록된다"
)


def main() -> int:
    exit_spec = ROOT / "openspec/specs/exit-policy/spec.md"
    safety_spec = ROOT / "openspec/specs/engine-safety/spec.md"

    exit_orig = block(exit_spec, EXIT_HEADER)
    exit_new = apply(exit_orig, EXIT_EDITS)
    safety_orig = block(safety_spec, SAFETY_HEADER)
    safety_new = apply(safety_orig, SAFETY_EDITS).rstrip("\n") + SAFETY_SCENARIO + "\n"

    for name, orig, new in (("exit-policy", exit_orig, exit_new), ("engine-safety", safety_orig, safety_new)):
        (OUT / f"{name}.modified.md").write_text(new, encoding="utf-8")
        diff = difflib.unified_diff(orig.splitlines(keepends=True), new.splitlines(keepends=True),
                                    f"openspec/specs/{name}/spec.md (정본 블록)", f"{name}.modified.md")
        (OUT / f"{name}.diff").write_text("".join(diff), encoding="utf-8")
    print("ok", len(EXIT_EDITS), "exit-policy edits;", len(SAFETY_EDITS) + 1, "engine-safety edits (incl. 1 scenario)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
