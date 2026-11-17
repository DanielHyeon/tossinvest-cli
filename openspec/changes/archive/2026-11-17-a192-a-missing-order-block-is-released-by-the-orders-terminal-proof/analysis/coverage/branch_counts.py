"""AST 분기 줄마다 그 줄에서 시작하는 cover 블록의 최대 count 를 뽑음(base 측정용)."""
import json, sys, re, collections
ast_path, profile = sys.argv[1], sys.argv[2]
d = json.load(open(ast_path))
f = d["file"]
counts = collections.defaultdict(int)
for line in open(profile):
    m = re.match(r'.*?/(internal/\S+|cmd/\S+):(\d+)\.(\d+),(\d+)\.(\d+) \d+ (\d+)$', line.strip())
    if not m or m.group(1) != f:
        continue
    s, e, n = int(m.group(2)), int(m.group(4)), int(m.group(6))
    for ln in range(s, e + 1):
        counts[ln] = max(counts[ln], n)
print(f'{d.get("receiver") or ""}.{d["function"]} {f} {d["start"]["line"]}-{d["end"]["line"]}')
for b in d.get("branches") or []:
    ln = b["at"]["line"]
    # 분기 본문은 다음 줄부터 시작하는 블록이 덮음 — 그 줄의 count 와 다음 줄의 count 를 함께 적음
    print(f'  {b["id"]}:{b["kind"]}@{ln} cond_line_count={counts.get(ln,0)} body_next_line_count={counts.get(ln+1,0)}')
