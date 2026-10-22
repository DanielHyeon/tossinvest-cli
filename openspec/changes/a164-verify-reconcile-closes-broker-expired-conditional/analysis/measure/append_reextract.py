"""각 번들 map 에 「a164 GREEN 재추출」 절을 덧붙임 — 구현 base aec9bde67 의 분기 열과 지금 분기 열을 difflib 으로 정렬."""
import difflib, json, os, subprocess, sys
S, FL = sys.argv[1], sys.argv[2]
ROOT = "/mnt/D/Axipient/workspace/TossOS-a164-impl/"
for d in sorted(os.listdir(FL)):
    new = json.load(open(f"{FL}/{d}/ast.json")); old = json.load(open(f"{S}/old-{d}.ast.json"))
    nsrc = open(ROOT + new["file"]).read().split("\n")
    osrc = subprocess.run(["git", "-C", ROOT, "show", "aec9bde67:" + old["file"]], capture_output=True, text=True, check=True).stdout.split("\n")
    seq = lambda a, src: [(b["kind"], src[b["at"]["line"] - 1].strip()) for b in a.get("branches") or []]
    o, n = seq(old, osrc), seq(new, nsrc)
    ops = []
    for tag, i1, i2, j1, j2 in difflib.SequenceMatcher(a=o, b=n, autojunk=False).get_opcodes():
        if tag == "equal":
            ops.append(f"equal 옛 B{i1+1}..B{i2} = 새 B{j1+1}..B{j2}")
        else:
            ops.append(f"{tag} 옛 B{i1+1}..B{i2}({i2-i1}) → 새 B{j1+1}..B{j2}({j2-j1})")
    sec = f"""
## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 {len(o)} 에서 편집 뒤 {len(n)} 로. `source_sha256` {old['source_sha256'][:12]}… → {new['source_sha256'][:12]}….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: {' · '.join(ops) if ops else '분기 없음(둘 다 0)'}.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
"""
    for name in ("function-logic-map.md", "branch-test-map.md"):
        open(f"{FL}/{d}/{name}", "a").write(sec)
print("appended")
