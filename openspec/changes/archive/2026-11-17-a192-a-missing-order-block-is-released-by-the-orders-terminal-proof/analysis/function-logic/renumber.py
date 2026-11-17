"""옛(base)·새(worktree) ast.json 의 분기를 (종류, 그 줄의 소스 문장) 순서로 정렬해 재번호 표를 만듦.

손 재번호 금지(기억 「분기 번호는 위치다」). 종류만으로 맞추면 같은 종류가 이어질 때 삽입 위치가 모호해지므로
분기 줄의 소스 문장(앞뒤 공백 제거)을 같이 비교함. 옛 소스는 base 커밋에서, 새 소스는 작업 트리에서 읽음.
사용: renumber.py <old.ast.json> <new.ast.json> <base-commit> <repo-root>
"""
import difflib, json, subprocess, sys
old_ast, new_ast, base, root = sys.argv[1:5]
old = json.load(open(old_ast)); new = json.load(open(new_ast))
old_src = subprocess.run(["git", "-C", root, "show", f"{base}:{old['file']}"], check=True,
                         capture_output=True, text=True).stdout.splitlines()
new_src = open(f"{root}/{new['file']}", encoding="utf-8").read().splitlines()
def keyed(ast, src):
    return [(b["kind"], src[b["at"]["line"] - 1].strip()) for b in ast.get("branches") or []]
a, b = keyed(old, old_src), keyed(new, new_src)
ob, nb = old.get("branches") or [], new.get("branches") or []
sm = difflib.SequenceMatcher(a=a, b=b, autojunk=False)
for tag, i1, i2, j1, j2 in sm.get_opcodes():
    if tag == "equal":
        for k in range(i2 - i1):
            o, n = ob[i1 + k], nb[j1 + k]
            print(f'{o["id"]}:{o["kind"]}@{o["at"]["line"]} -> {n["id"]}:{n["kind"]}@{n["at"]["line"]}')
    else:
        for o in ob[i1:i2]:
            print(f'{o["id"]}:{o["kind"]}@{o["at"]["line"]} -> (removed/changed, {tag})  | {a[i1 + ob[i1:i2].index(o)][1]}')
        for n in nb[j1:j2]:
            print(f'(new/changed, {tag}) -> {n["id"]}:{n["kind"]}@{n["at"]["line"]}  | {b[j1 + nb[j1:j2].index(n)][1]}')
