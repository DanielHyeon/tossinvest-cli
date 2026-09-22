"""task 7.5.22 — 도달 계측기의 표식이 **문법을 깨는** 변이를 센다 (C1 검증).

옛 계측기는 표식을 문장이 아닌 자리 앞에 심어 `SyntaxError` 를 냈고, 그 트레이스백이 표식 줄을 그대로
인쇄해서 **시험 0개 돈 판**을 "도달함" 으로 기록했다. 이 하네스는 스위트를 돌리지 않고 `compile()` 만
해 보므로 싸다 — 새 `reached()` 의 gate 가 거를 자리 수를 그대로 센다.

    python3 7522_marker.py
"""
import importlib.util, sys
from pathlib import Path
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
H = Path(__file__).resolve().parent / "75_mut.py"

# MUTATIONS 만 꺼내 쓴다 — `__main__` 을 안 돌린다.
src = H.read_text(encoding="utf-8")
ns = {"__name__": "not_main", "__file__": str(H)}
exec(compile(src.split('if __name__ == "__main__":')[0], str(H), "exec"), ns)
MUT = ns["MUTATIONS"]
pristine = (ROOT / "tools" / "logic-map" / "check_analysis.py").read_text(encoding="utf-8")

broke, ok_plant, no_anchor, moved = [], [], [], []
for name, edits in MUT.items():
    if edits == "MOVE_FIRST":
        moved.append(name); continue
    marked = pristine
    planted = any_broken = False
    for old_line, _ in edits:
        head = old_line.splitlines()[0]
        if head.strip().startswith(("#", '"')) or not head.strip():
            continue
        indent = head[: len(head) - len(head.lstrip())]
        cand = marked.replace(head, f'{indent}import sys as _s; print("REACHED", file=_s.stderr)\n{head}', 1)
        if cand == marked:
            continue
        try:
            compile(cand, "t", "exec")
        except SyntaxError:
            any_broken = True
            continue
        marked = cand
        planted = True
    if any_broken and not planted:
        broke.append(name)
    elif any_broken:
        broke.append(name + " (일부)")
    elif planted:
        ok_plant.append(name)
    else:
        no_anchor.append(name)

print(f"변이 전수 {len(MUT)}")
print(f"  표식이 **문법을 깬다**(옛 계측기가 거짓 '도달함' 을 낼 자리): **{len(broke)}**")
for n in broke: print("     ", n)
print(f"  정상 심음 {len(ok_plant)} · 심을 앵커 없음 {len(no_anchor)} · MOVE_FIRST {len(moved)}")
for n in no_anchor: print("      앵커없음:", n)
