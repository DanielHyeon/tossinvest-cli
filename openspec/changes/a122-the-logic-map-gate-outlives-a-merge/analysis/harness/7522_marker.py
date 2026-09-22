"""task 7.5.22 · 7.5.23 — 도달 계측기가 **거짓말할 수 있는 자리**를 정적으로 전수한다.

스위트를 안 돌리므로 싸다. 세 갈래를 가른다 (7.5.23 정정 — 7.5.22 는 셋을 "30" 하나로 뭉갰다):

1. **거짓 "도달함"**: 표식이 문법을 깨고, `SyntaxError` 트레이스백이 **표식 줄**을 인쇄한다.
   그 줄에 `REACHED` 가 있으므로 시험을 0개 돌린 판이 "도달함" 이 된다.
2. **거짓 "안 닿음"**: 표식이 문법을 깨지만 트레이스백이 **다른 줄**을 인쇄한다. 눈이 먼 것인데
   옛 계측기는 이것을 진짜 음성과 같게 적었다.
3. **엉뚱한 줄**: 표식 head 가 파일에서 유일하지 않아 `str.replace(…, 1)` 이 **첫** 자리에 심는다.
   문법은 멀쩡하므로 `compile()` 이 못 본다 — 계측기가 **안 본 줄**에 대해 자신 있게 답한다.

셋 다 새 `reached()` 가 `"?"`(못 쟀다)로 낸다. 이 하네스는 그 수가 유지되는지 보는 자다.

    python3 7522_marker.py [<revision>]     # 기본은 워킹트리
"""
import ast
import subprocess
import sys
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
H = Path(__file__).resolve().parent / "75_mut.py"
REV = sys.argv[1] if len(sys.argv) > 1 else ""

src = H.read_text(encoding="utf-8")
ns = {"__name__": "not_main", "__file__": str(H)}
exec(compile(src.split('if __name__ == "__main__":')[0], str(H), "exec"), ns)
MUT = ns["MUTATIONS"]

TARGET = ROOT / "tools" / "logic-map" / "check_analysis.py"
if REV:
    pristine = subprocess.run(["git", "show", f"{REV}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout.decode("utf-8")
else:
    pristine = TARGET.read_text(encoding="utf-8")

tree = ast.parse(pristine)
spans = sorted(((n.lineno, n.end_lineno, n.name) for n in ast.walk(tree)
                if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef))),
               key=lambda row: row[1] - row[0])


def owner(line: int) -> str:
    for start, end, name in spans:
        if start <= line <= end:
            return name
    return "<module>"


def line_of(index: int) -> int:
    return pristine.count("\n", 0, index) + 1


false_yes, false_no, wrong_line, no_anchor, partial, fine = [], [], [], [], [], []
for name, edits in MUT.items():
    if edits == "MOVE_FIRST":
        continue
    eligible = planted = 0
    verdicts = []
    for old_line, _ in edits:
        head = old_line.splitlines()[0]
        if head.strip().startswith(("#", '"')) or not head.strip():
            continue
        eligible += 1
        count = pristine.count(head)
        if count == 0:
            verdicts.append(("앵커없음", ""))
            continue
        if count != 1:
            here = owner(line_of(pristine.index(head)))
            there = owner(line_of(pristine.index(edits[0][0]))) if pristine.count(edits[0][0]) == 1 else "?"
            verdicts.append(("엉뚱한줄" if here != there else "유일하지않음", f"표식@{here} 변이@{there}"))
            continue
        indent = head[: len(head) - len(head.lstrip())]
        candidate = pristine.replace(
            head, f'{indent}import sys as _s; print("REACHED", file=_s.stderr)\n{head}', 1)
        try:
            compile(candidate, "target", "exec")
        except SyntaxError as error:
            kind = "거짓도달함" if (error.text and "REACHED" in error.text) else "거짓안닿음"
            verdicts.append((kind, error.msg[:34]))
            continue
        planted += 1
    kinds = {kind for kind, _ in verdicts}
    if "거짓도달함" in kinds:
        false_yes.append((name, verdicts))
    elif "거짓안닿음" in kinds:
        false_no.append((name, verdicts))
    elif "엉뚱한줄" in kinds or "유일하지않음" in kinds:
        wrong_line.append((name, verdicts))
    elif "앵커없음" in kinds and planted == 0:
        no_anchor.append((name, verdicts))
    elif eligible == 0:
        # 주석·문자열만 바꾸는 변이다 — 심을 자리가 없다. 옛 계측기는 이것도 "안 닿음" 으로 적었다.
        no_anchor.append((name, [("심을자리없음", "")]))
    elif planted != eligible:
        partial.append((name, f"{planted}/{eligible}"))
    else:
        fine.append(name)

print(f"# 도달 계측기 정적 전수 — 변이 {len(MUT)} · 대상 {REV or '워킹트리'}\n")
for label, rows in (("거짓 '도달함' (트레이스백이 표식 줄을 인쇄)", false_yes),
                    ("거짓 '안 닿음' (문법은 깨지는데 다른 줄을 인쇄 — 눈멂)", false_no),
                    ("엉뚱한 줄 (표식 head 가 유일하지 않다)", wrong_line),
                    ("앵커 없음 (심을 자리가 없다)", no_anchor),
                    ("부분 심기", partial)):
    print(f"{label}: **{len(rows)}**")
    for row in rows:
        print(f"    {row[0][:46]:48s} {row[1] if isinstance(row[1], str) else row[1][:2]}")
print(f"\n정상 심음: {len(fine)}")
print("새 `reached()` 는 위 다섯을 전부 `\"?\"`(못 쟀다)로 낸다 — 자신 있는 거짓을 안 낸다.")
