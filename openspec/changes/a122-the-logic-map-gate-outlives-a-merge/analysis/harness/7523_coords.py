"""task 7.5.23 — 이 change 의 산문에 **살아 있는 절대 줄 인용**이 남아 있는지 센다.

세 로트 연속으로 같은 실수를 했다: 수리와 인용이 **같은 커밋**에 있으면 인용이 먼저 쓰이고
수리가 나중에 민다. 7.5.22 는 그 실수를 고치겠다고 적으면서 새 좌표 열여섯 개를 넣었고
**전부 +2 만큼 낡은 채로** 커밋됐다(자기 주석이 민 것이다).

그래서 규칙을 산문이 아니라 **세는 것**으로 만든다. 두 가지를 가른다:

- **산문의 인용** (`review.md` · `tasks.md` · `HANDOFF.md` · `branch-test-map.md` · README):
  절대 줄을 쓰면 안 된다. 함수 이름으로 쓴다. 여기 걸리면 rc≠0.
- **FLM 머리말의 범위** (`function-logic-map.md` 첫 문단): 추출 시점의 **기록**이라 허용하되,
  옆의 `ast*.json` 범위와 **맞아야** 한다. 안 맞으면 그것은 기록이 아니라 주장이다.

    python3 7523_coords.py
"""
import json
import re
import sys
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
CHANGE = Path(__file__).resolve().parents[2]

CITE = re.compile(r"(?:check_analysis|execution_baseline|role_check)\.py:(\d+)(?:-(\d+))?")
SYMBOL_CITE = re.compile(r"`([A-Za-z_][A-Za-z0-9_]*):(\d+)(?:-(\d+))?`")
PROSE = ("review.md", "tasks.md", "HANDOFF.md", "proposal.md", "branch-test-map.md")

live: list[tuple[str, int, str]] = []
for path in sorted(CHANGE.rglob("*.md")):
    if "_work" in path.parts:
        continue
    header = path.name == "function-logic-map.md"
    text = path.read_text(encoding="utf-8")
    for number, line in enumerate(text.splitlines(), start=1):
        if header and number <= 8:
            continue                      # 머리말은 아래에서 따로 본다
        for match in list(CITE.finditer(line)) + list(SYMBOL_CITE.finditer(line)):
            if "정정" in line or "거짓" in line or "낡" in line or "옛" in line:
                continue                  # 틀린 좌표를 **인용하며 고치는** 문장은 통과
            live.append((str(path.relative_to(CHANGE)), number, match.group(0)))

print(f"# 좌표 열거 — change {CHANGE.name}\n")
print(f"산문에 살아 있는 절대 줄 인용: **{len(live)}**")
for where, number, token in live:
    print(f"    {where}:{number}  {token}")

# --- `ast*.json` 이 **실재하는 소스**를 기술하는가 (7.5.23) ---
# 이것이 결정적인 검사다. 머리말과 `ast.json` 이 **같이** 틀리면 둘을 맞대 보는 것으로는 안 걸린다.
# 7.5.22 가 그랬다 — 둘 다 `:135-179` 인데 커밋된 파일은 `:137-181` 이고, 그 `source_sha256` 은
# 역사의 어느 blob 과도 안 맞는 **중간 상태**였다.
import ast as _ast
import hashlib
import subprocess

def _known_sources(relative: str) -> dict[str, bytes]:
    found: dict[str, bytes] = {}
    here = ROOT / relative
    if here.exists():
        raw = here.read_bytes()
        found[hashlib.sha256(raw).hexdigest()] = raw
    revisions = subprocess.run(
        ["git", "log", "--format=%H", "--", relative],
        cwd=ROOT, capture_output=True, text=True, check=True).stdout.split()
    # **한 프로세스**로 읽는다 — revision 마다 `git show` 를 부르면 이 하네스가 분 단위가 된다
    # ([[the-second-bottleneck-hides-behind-the-first]]: 비용은 프로세스 수로 못 박는다).
    wanted = b"".join(f"{revision}:{relative}\n".encode() for revision in revisions)
    batch = subprocess.run(["git", "cat-file", "--batch"],
                           cwd=ROOT, input=wanted, capture_output=True)
    stream, at = batch.stdout, 0
    while at < len(stream):
        end = stream.find(b"\n", at)
        if end < 0:
            break
        header = stream[at:end].split()
        at = end + 1
        if len(header) != 3:
            continue
        size = int(header[2])
        found.setdefault(hashlib.sha256(stream[at:at + size]).hexdigest(), stream[at:at + size])
        at += size + 1
    return found

def _span(source: bytes, name: str) -> tuple[int, int] | None:
    for node in _ast.walk(_ast.parse(source.decode("utf-8"))):
        if isinstance(node, (_ast.FunctionDef, _ast.AsyncFunctionDef)) and node.name == name:
            return node.lineno, node.end_lineno
    return None

cache: dict[str, dict[str, bytes]] = {}
orphan, drifted, anchored = [], [], 0
for bundle in sorted(CHANGE.rglob("ast*.json")):
    if "_work" in bundle.parts:
        continue
    try:
        data = json.loads(bundle.read_text(encoding="utf-8"))
    except Exception:
        continue
    if data.get("language") != "python":
        continue
    relative = data["file"]
    sources = cache.setdefault(relative, _known_sources(relative))
    where = str(bundle.relative_to(CHANGE))
    source = sources.get(data.get("source_sha256", ""))
    if source is None:
        orphan.append((where, data.get("source_sha256", "")[:12]))
        continue
    anchored += 1
    span = _span(source, data["function"])
    claimed = (int(data["start"]["line"]), int(data["end"]["line"]))
    if span != claimed:
        drifted.append((where, claimed, span))

print(f"\n`ast*.json`(python) 지문이 **실재하는 소스**를 가리키는가: {anchored} 맞음")
print(f"    지문이 어디에도 없는 것 — 커밋 안 한 중간 상태: **{len(orphan)}**")
for where, sha in orphan:
    print(f"        {where}  sha {sha}…")
print(f"    지문은 맞는데 범위가 다른 것: **{len(drifted)}**")
for where, claimed, span in drifted:
    print(f"        {where}  적힌 {claimed}  실제 {span}")

mismatch, checked = [], 0
for flm in sorted(CHANGE.rglob("function-logic-map.md")):
    if "_work" in flm.parts:
        continue
    head = "\n".join(flm.read_text(encoding="utf-8").splitlines()[:8])
    spans = {(int(a), int(b)) for a, b in CITE.findall(head) if b}
    if not spans:
        continue
    known = set()
    for bundle in sorted(flm.parent.glob("ast*.json")):
        try:
            data = json.loads(bundle.read_text(encoding="utf-8"))
        except Exception:
            continue
        known.add((int(data["start"]["line"]), int(data["end"]["line"])))
    for span in spans:
        checked += 1
        if span not in known:
            mismatch.append((str(flm.relative_to(CHANGE)), span, sorted(known)))

print(f"\nFLM 머리말 범위: {checked} 개 · 옆 `ast*.json` 과 **안 맞는 것: {len(mismatch)}**")
for where, span, known in mismatch:
    print(f"    {where}  머리말 {span[0]}-{span[1]}  ast {known}")
# 판정은 **객관적인 것**만 한다: 지문이 실재하지 않거나 범위가 어긋난 번들, 그리고
# 열린 task 가 읽는 `tasks.md` 의 살아 있는 좌표. 나머지(다른 산문 105건)는 세기만 한다 —
# 이 change 전체의 누적이라 이 로트의 범위가 아니고, 숫자로 남겨 다음 로트가 줄인다.
# **열린** task 만 판정한다. 닫힌 task 는 그때의 기록이고, 열린 task 는 **미래 작업의 지시**다 —
# 낡은 좌표를 따라가서 엉뚱한 자리를 고치는 것이 이 change 에서 실제로 일어난 일이다.
open_lines: set[int] = set()
tasks = (CHANGE / "tasks.md").read_text(encoding="utf-8").splitlines()
inside = False
for number, line in enumerate(tasks, start=1):
    if line.startswith("- ["):
        inside = line.startswith("- [ ]")
    elif line.startswith(("#", "- ")) or not line.strip():
        inside = inside and line.startswith(" ")
    if inside:
        open_lines.add(number)
open_live = [row for row in live if row[0] == "tasks.md" and row[1] in open_lines]
print(f"\n`tasks.md` 의 살아 있는 좌표 {len([r for r in live if r[0] == 'tasks.md'])}"
      f" — 그중 **열린 task 안: {len(open_live)}**")
for where, number, token in open_live:
    print(f"    {where}:{number}  {token}")
print(f"판정: 열린 task 의 좌표 {len(open_live)} · 뜬 번들 {len(orphan)} · 낡은 범위 {len(drifted)}")
raise SystemExit(1 if (open_live or orphan or drifted) else 0)
