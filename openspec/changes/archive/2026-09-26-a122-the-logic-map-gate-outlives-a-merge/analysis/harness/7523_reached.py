"""task 7.5.23 — 도달 계측기 `reached()` 를 **직접** 시험한다.

계측기는 이 change 의 모든 변이 증거의 바닥이다. 그런데 `reached()` 는 **SURVIVED 인 판에서만**
불리므로, 생존이 0 인 로트에서는 한 번도 안 돌고 증거에 안 나타난다 — 7.5.22 가 정확히 그랬다
(AC1~AC5 가 전부 CAUGHT 라 고친 계측기가 한 번도 안 돌았다).

여기서는 `run()` 을 가짜로 바꿔 여섯 경우를 만든다. 재는 것은 스위트가 아니라 **계측기의 판단**이다.

    python3 7523_reached.py
"""
import sys
import tempfile
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
H = Path(__file__).resolve().parent / "75_mut.py"

src = H.read_text(encoding="utf-8")
ns = {"__name__": "not_main", "__file__": str(H)}
exec(compile(src.split('if __name__ == "__main__":')[0], str(H), "exec"), ns)
reached = ns["reached"]

CONTROL = 42
BODY = '''def only_here():
    x = 1
    return x


def twice_a():
    y = 2
    return y


def twice_b():
    y = 2
    return y


def wrapped():
    return sum([
        1,
        2,
    ])
'''


def exercise(label, edits, stderr_of_marked, control=CONTROL):
    """`run()` 을 가짜로 바꾸고 `reached()` 를 부른다. 가짜는 **표식이 심긴 파일**을 읽어서
    그 파일이 문법적으로 성립하는지 보고, 호출자가 준 stderr 를 돌려준다."""
    holder = tempfile.TemporaryDirectory()
    target = Path(holder.name) / "target.py"
    target.write_text(BODY, encoding="utf-8")
    seen = {}

    def fake_run(_names):
        seen["marked"] = target.read_text(encoding="utf-8")
        return 0, stderr_of_marked(seen["marked"], control)

    ns["run"] = fake_run
    answer = reached(target, edits, BODY, control)
    restored = target.read_text(encoding="utf-8")
    holder.cleanup()
    return answer, seen.get("marked"), restored


def suite_ran(marked, control):
    return f"REACHED\nRan {control} tests in 0.1s\n\nOK\n" if "REACHED" in marked else f"Ran {control} tests in 0.1s\n\nOK\n"


def suite_quiet(marked, control):
    return f"Ran {control} tests in 0.1s\n\nOK\n"


def collection_died(marked, control):
    return "Traceback (most recent call last):\n  File \"t\", line 1\n    import sys as _s; print(\"REACHED\", file=_s.stderr)\nSyntaxError: invalid syntax\n"


CASES = [
    ("표식이 심기고 스위트가 그 자리를 돈다",
     [("    x = 1", "    x = 9")], suite_ran, "YES"),
    ("표식이 심기고 스위트가 그 자리를 **안** 돈다",
     [("    x = 1", "    x = 9")], suite_quiet, "no"),
    ("표식 head 가 파일에서 **유일하지 않다**",
     [("    y = 2", "    y = 9")], suite_ran, "?"),
    ("표식이 문법을 깬다 (문장이 아닌 자리)",
     [("        1,", "        7,")], suite_ran, "?"),
    ("심을 자리가 없다 (주석만 바꾸는 변이)",
     [("# nothing", "# something")], suite_ran, "?"),
    ("스위트가 대조군과 **다른 수**를 돌았다",
     [("    x = 1", "    x = 9")],
     lambda m, c: suite_ran(m, c - 1), "?"),
]

print("# `reached()` 직접 시험\n")
print("| 경우 | 기대 | 답 | |")
print("|---|---|---|---|")
worst = 0
for label, edits, stderr, want in CASES:
    answer, marked, restored = exercise(label, edits, stderr)
    ok = answer == want and restored == BODY
    worst |= 0 if ok else 1
    note = "" if restored == BODY else " · **원복 실패**"
    print(f"| {label} | `{want}` | `{answer}` | {'✅' if ok else '❌'}{note} |")
print("\n원복은 `finally` 안에 있다 — 스위트가 멎어도 표식 붙은 사본이 안 남는다.")
raise SystemExit(worst)
