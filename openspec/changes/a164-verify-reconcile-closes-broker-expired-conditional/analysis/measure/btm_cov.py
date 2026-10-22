"""시험별 coverprofile 로 분기 본문 도달(taken)을 잰다 — BTM 행의 근거(손으로 고르지 않음).

분기 (L, C) 의 본문 블록 = 같은 줄 L 에서 열이 C 보다 뒤에 시작하는 블록(조건문 자체를 담은 블록은 열 C 에서 시작함).
"""
import json, os, re, sys
MOD = "github.com/JungHoonGhae/tossinvest-cli/"
pertest, fl = sys.argv[1], sys.argv[2]
hits = {}
for n in os.listdir(pertest):
    if not n.endswith(".out"):
        continue
    test = n.split("__", 1)[1][:-4].lstrip("_")
    for line in open(os.path.join(pertest, n)):
        m = re.match(r"(\S+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)", line)
        if not m or int(m.group(7)) == 0:
            continue
        hits.setdefault((m.group(1).replace(MOD, ""), int(m.group(2))), []).append((int(m.group(3)), test))
ROOT = "/mnt/D/Axipient/workspace/TossOS-a164-impl/"
def taken(f, line, col):
    # 조건이 여러 줄이면 본문 `{` 는 뒤 줄에 있음 — 그 줄의 마지막 `{` 뒤에서 시작하는 블록이 본문.
    src = open(ROOT + f).read().split("\n")
    for ln in range(line, line + 8):
        text = src[ln - 1].rstrip()
        if text.endswith("{"):
            brace = len(text)  # 1 기반 열의 `{`
            return sorted({t for c, t in hits.get((f, ln), []) if c >= brace})
    return []
out = {}
for d in sorted(os.listdir(fl)):
    a = json.load(open(os.path.join(fl, d, "ast.json")))
    rows = {}
    for i, b in enumerate(a.get("branches") or [], 1):
        rows[f"B{i}"] = taken(a["file"], b["at"]["line"], b["at"]["column"])
    if not a.get("branches"):
        rows["B1"] = sorted({t for c, t in hits.get((a["file"], a["start"]["line"]), [])})
    out[d] = rows
json.dump(out, sys.stdout, indent=1, ensure_ascii=False)
