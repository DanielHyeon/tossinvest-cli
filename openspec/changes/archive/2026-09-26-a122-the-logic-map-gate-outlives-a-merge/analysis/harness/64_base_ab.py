#!/usr/bin/env python3
"""task 6.4(b)(c) A/B — 해소기 · base 해소가 저장소의 **모든** change 디렉터리에서 같은 답을 내는가.

6.4 가 바꾼 것은 판정의 **입구** 둘뿐이다: id → 디렉터리(`ARCHIVED_CHANGE` 의 `re.ASCII`)와 디렉터리 → base
(`resolve_base` 의 모양 · HEAD 대조 · 태그 검사). 그 뒤의 판정은 한 줄도 안 바뀌었으므로 입구의 답이 같으면 판정도
같다 — 전수 `main()` A/B(창 9 번, 약 2,000 초) 대신 입구를 전수로 잰다. 예외도 답이다: 타입과 문장을 비교한다.

    python3 64_base_ab.py [<before-rev>]   # 기본 HEAD — 편집이 커밋되기 **전** 워킹트리와 비교할 때 쓴다
"""
import importlib.util
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
WORK = Path(__file__).resolve().parent / "_work" / "64_base_ab_before"
sys.path.insert(0, str(REPO / "tools" / "logic-map"))

import subprocess  # noqa: E402

import check_analysis as new  # noqa: E402

before = sys.argv[1] if len(sys.argv) > 1 else "HEAD"
WORK.mkdir(parents=True, exist_ok=True)
(WORK / "check_analysis_before.py").write_bytes(subprocess.run(
    ["git", "show", f"{before}:tools/logic-map/check_analysis.py"], cwd=REPO, capture_output=True, check=True,
).stdout)
spec = importlib.util.spec_from_file_location("ca_before", WORK / "check_analysis_before.py")
old = importlib.util.module_from_spec(spec)
sys.modules["ca_before"] = old
spec.loader.exec_module(old)
# 이 비교는 편집 전 코드가 `head` 인자를 모른다는 것을 안다 — 새 쪽에만 넘긴다.
assert "head" not in old.resolve_base.__code__.co_varnames[:old.resolve_base.__code__.co_argcount
                                                         + old.resolve_base.__code__.co_kwonlyargcount], \
    "before 쪽이 이미 6.4 판본이다 — 비교 기준이 오염됐다"


def answer(call):
    try:
        return ("ok", str(call()))
    except Exception as exc:  # 예외도 판정이다
        return (type(exc).__name__, str(exc))


root = REPO
head = new._head_commit(root)
changes = root / "openspec" / "changes"
names = sorted([p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
               + [p.name for p in (changes / "archive").iterdir() if p.is_dir()])
same = 0
different = []
for name in names:
    old_id, new_id = old._archived_change_id(name), new._archived_change_id(name)
    change = new_id or name
    resolved = (answer(lambda: old.resolve_referenced_change(root, change)),
                answer(lambda: new.resolve_referenced_change(root, change)))
    bases = (None, None)
    if resolved[1][0] == "ok":
        directory = Path(resolved[1][1])
        bases = (answer(lambda: old.resolve_base(directory, root, {}, change_id=change)),
                 answer(lambda: new.resolve_base(directory, root, {}, change_id=change, head=head)))
    if old_id == new_id and resolved[0] == resolved[1] and bases[0] == bases[1]:
        same += 1
    else:
        different.append((name, resolved, bases))
print(f"before {before} · HEAD {head[:12]} · change 디렉터리 {len(names)} · SAME {same} · DIFFERENT {len(different)}")
for row in different:
    print(row)
