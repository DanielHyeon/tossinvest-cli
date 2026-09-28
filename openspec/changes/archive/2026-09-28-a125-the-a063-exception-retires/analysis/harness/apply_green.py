#!/usr/bin/env python3
"""a125 2.1 — check_analysis.py 에서 a063 이관 특례를 걷어낸다. 인자: 대상 check_analysis.py. 교체마다 정확히 한 번 맞아야 한다."""
import sys
from pathlib import Path

path = Path(sys.argv[1])
text = path.read_text(encoding="utf-8")


def swap(old: str, new: str) -> None:
    global text
    assert text.count(old) == 1, (text.count(old), old[:90])
    text = text.replace(old, new)


swap("from execution_baseline import AdoptionError, validate as validate_execution_baseline\n", "")
swap('    """이 change 의 비교 기준(창의 **시작**). a063 이관이면 `E`, 아니면 `base-commit.txt` 의 커밋이다.\n',
     '    """이 change 의 비교 기준(창의 **시작**). 모든 change 에서 `base-commit.txt` 의 커밋이다 (a125 — a063 이관 특례 폐기).\n')
swap('''    try:
        # 이관 신원은 디렉터리 이름이 아니라 요청받은 id 로 가른다 — 아카이브가 이름을
        # 바꾼다(task 6.2). 필수 인자인 이유: 기본값이 있으면 id 를 잊은 호출자가 조용히
        # 옛 판정(이름)으로 떨어지고, 그것이 아카이브된 a063 을 막던 바로 그 판정이다.
        adoption = validate_execution_baseline(change_dir, root, persisted, change_id)
    except AdoptionError as exc:
        raise ValueError(f"invalid execution-baseline adoption: {exc}") from exc
    effective = str(adoption["effective_base"]) if adoption else persisted
    if context is not None:
        context["execution_baseline_adoption"] = adoption is not None
        context["effective_base"] = effective
        if adoption:
            # `validate` 는 감사된 창의 **끝**을 이미 돌려준다. 옛 판본은 그 값을
            # 버리고 `landed-commit.txt` 를 찾았고, 없으니 대상이 워킹트리가 됐다.
            context["adoption_source"] = str(adoption["source"])
    override = os.environ.get("SDD_BASE_REF", "").strip()
    if override and resolve(override) != effective:
        raise ValueError("SDD_BASE_REF must resolve to the selected effective comparison base")
    return effective
''', '''    # 다른 기준을 고르는 입력은 없다 (a125). 변경 디렉터리에 남은 `execution-baseline.json` 은 읽는 코드가 없어
    # 판정으로 들어가는 문이 아니라 데이터다 — 그 모양을 거절하는 가드는 죽은 가드다(design D2).
    if context is not None:
        context["effective_base"] = persisted
    override = os.environ.get("SDD_BASE_REF", "").strip()
    if override and resolve(override) != persisted:
        raise ValueError("SDD_BASE_REF must resolve to the selected effective comparison base")
    return persisted
''')
swap('''# 이관 경로가 착지 기록을 거절하는 문장. 판정 경로(`check`)와 기록 경로(`record_landing`)가
# 한 벌씩 들고 있던 동안 이미 "ends" / "already ends" 로 갈렸다 (task 7.6, 리뷰 I5).
ADOPTION_REFUSES_A_LANDING = (
    f"execution-baseline adoption does not accept a `{LANDING_FILE}` record: "
    "the window ends at the audited source commit"
)
''', "")
swap("# **같은 문장**을 쓴다 — 두 벌이면 갈린다(위 이관 문장이 그랬다).\n",
     "# **같은 문장**을 쓴다 — 두 벌이면 갈린다(a122 7.6 리뷰 I5 에서 옛 이관 문장이 그랬다).\n")
swap('''def _target_text(landing: str, audited: bool = False) -> str:
    """비교 대상 쪽 끝을 사람이 읽는 말로. **무엇이** 그 끝을 고정했는지까지 말한다.

    이관 예외의 끝은 저자가 선언한 값이 아니라 `execution-baseline.json` 이 감사한
    `source_commit` 이다. 둘을 같은 말로 적으면 있지도 않은 파일을 가리키게 된다.
    """''', '''def _target_text(landing: str) -> str:
    """비교 대상 쪽 끝을 사람이 읽는 말로. **무엇이** 그 끝을 고정했는지까지 말한다(착지 기록 또는 워킹트리)."""''')
swap('    return f"audited source-commit {landing}" if audited else f"landed-commit {landing}"\n',
     '    return f"landed-commit {landing}"\n')
swap("    것에 해독하는 함수를 부르면 못 읽는 기록 앞에서 질문 자체가 터진다 — 이관 경로의\n    probe 와",
     "    것에 해독하는 함수를 부르면 못 읽는 기록 앞에서 질문 자체가 터진다 — 옛 이관 경로(a125 에서 폐기)의\n    probe 와")
swap("    # 들어온다(이관인가, 감사된 source 는 무엇인가). 호출자가 문맥을 줬으면 같은\n",
     "    # 들어온다(비교 기준 · 역사의 끝). 호출자가 문맥을 줬으면 같은\n")
swap('''    adopted = bool(facts.get("execution_baseline_adoption"))
    if adopted and _landing_record(change_dir, root, head) is not None:
        # 이관 예외의 정당성은 "판정에 들어가는 입력을 하나도 빠짐없이 열거하고
        # digest 로 묶었다"이다. `landed-commit.txt` 는 `openspec/` 아래라 drift 검사가
        # 통과시키고, 추적 파일이라 untracked 감사도 못 보고, 닫힌 키 집합에도 없다.
        # 그런데 비교 대상을 고른다 — 손잡이는 하나여야 하고 그것은 감사된 쪽이다.
        return [ADOPTION_REFUSES_A_LANDING], False
    try:
        # 조상 판정은 `base-commit.txt` 의 글자가 아니라 `resolve_base` 가 **반환한**
        # 값에 건다. a063 은 그 둘이 다르다(P → E).
        # 이관이면 착지를 **해소하지 않는다**. `resolve_landing` 의 판정들은 저자가
        # 고른 값을 위한 것이고, 감사된 source 는 고른 값이 아니다 — `validate` 가
        # ancestry(P,E)·ancestry(E,source)·ancestry(source,head,strict)·tree 대조·
        # digest 셋으로 이미 묶는다. 같은 판정을 두 번 하지 않는다.
        # 착지 판정 · 대상 판정 · base 모양 조언은 위에서 **한 번 읽은** 증거를 쓴다 (task 7.5.2 · 7.5.2.1).
        landing = str(facts.get("adoption_source", "")) if adopted \\
            else resolve_landing(change_dir, root, base, head, evidence)
''', '''    try:
        # 착지 판정 · 대상 판정 · base 모양 조언은 위에서 **한 번 읽은** 증거를 쓴다 (task 7.5.2 · 7.5.2.1).
        # a063 도 다른 change 와 같이 착지 규칙을 받는다 (a125 — 이관 특례 폐기).
        landing = resolve_landing(change_dir, root, base, head, evidence)
''')
swap("    return _verdict(root, base, landing, adopted, required, evidence, review_text), True\n",
     "    return _verdict(root, base, landing, required, evidence, review_text), True\n")
swap("    root: Path, base: str, landing: str, adopted: bool, required: dict[tuple[str, str], dict],\n",
     "    root: Path, base: str, landing: str, required: dict[tuple[str, str], dict],\n")
swap("f\"between base {base[:12]} and {_target_text(landing, adopted)}: {names}\"",
     "f\"between base {base[:12]} and {_target_text(landing)}: {names}\"")
swap('''    # 재리뷰 Codex P2 · 적대 F3). 닿는 것은 착지 판정을 안 거치는 이관 change 다 — 착지 경로는 같은 바이트를
    # 이미 골랐고(못 고르면 거기서 결함), 여기서 달라질 수 있는 것은 소스 경로의 심링크 풀이뿐이다.''',
     '''    # 재리뷰 Codex P2 · 적대 F3). 그 갈래가 닿던 곳은 착지 판정을 안 거치던 이관 change 였고 a125 에서 폐기됐다 —
    # 착지는 이제 언제나 같은 선별을 먼저 거치므로(못 고르면 거기서 결함) 여기서 달라질 수 있는 것은 소스 경로의
    # 심링크 풀이뿐이다. 방어로 둔다.''')
swap('''    if facts.get("execution_baseline_adoption"):
        return ADOPTION_REFUSES_A_LANDING, ""
''', "")
swap('''    - 실행 기준선 이관 change(a063). 그 경로의 창 끝은 감사된 source commit 이고
      spec 이 이 기록을 받지 않는다(SHALL NOT).
''', "")
swap('''        audited = bool(context.get("execution_baseline_adoption"))
''', "")
swap("f\"[logic-map] {args.change}: base {base[:12]} → {_target_text(landing, audited)} \"",
     "f\"[logic-map] {args.change}: base {base[:12]} → {_target_text(landing)} \"")
swap('''    if context.get("execution_baseline_adoption"):
        print(f"[logic-map] {args.change}: execution-baseline adoption exception evidence complete")
    else:
        print(f"[logic-map] {args.change}: evidence complete or diff-proven exempt")
''', '''    print(f"[logic-map] {args.change}: evidence complete or diff-proven exempt")
''')
path.write_text(text, encoding="utf-8")
print("ok")
