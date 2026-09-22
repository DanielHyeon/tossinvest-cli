"""task 7.5.22 — 새 거절(`--numstat` 이 `-`/`-` 인 `*.go`)이 **오늘 거절할 정상 입력**을 센다.

[[fail-closed-must-name-what-it-rejects]]: 보수적 가드를 넣을 때 그것이 거부할 정상 입력을 **먼저**
센다. `0`/`0`(mode-only)은 거절하면 **안 되는** 쪽이라 같이 센다.

    python3 7522_numstat.py
"""
import subprocess
from collections import Counter
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
CH = ROOT / "openspec" / "changes"

bases = []
for d in sorted(CH.rglob("base-commit.txt")):
    bases.append((d.parent.name, d.read_text(encoding="utf-8").strip()))

tally = Counter()
suppressed = []
mode_only = []
for name, base in bases:
    r = subprocess.run(
        ["git", "diff", "--no-ext-diff", "--numstat", "-z", base, "--", "*.go"],
        cwd=ROOT, capture_output=True,
    )
    if r.returncode:
        tally["git 실패"] += 1
        continue
    tally["잰 change"] += 1
    chunks = [c for c in r.stdout.split(b"\0")]
    i = 0
    while i < len(chunks):
        c = chunks[i]
        if not c:
            i += 1
            continue
        parts = c.split(b"\t", 2)
        if len(parts) != 3:
            i += 1
            continue
        add, dele, path = parts
        if path == b"":          # rename: 다음 둘이 old·new
            path = chunks[i + 2] if i + 2 < len(chunks) else b"?"
            i += 3
        else:
            i += 1
        tally["*.go 파일 줄"] += 1
        if add == b"-" and dele == b"-":
            suppressed.append((name, path.decode("utf-8", "replace")))
        elif add == b"0" and dele == b"0":
            mode_only.append((name, path.decode("utf-8", "replace")))

for k, v in tally.items():
    print(f"  {k}: {v}")
print(f"\n본문 억제(`-`/`-`) — 새 거절이 걸 것: **{len(suppressed)}**")
for n, p in suppressed[:10]:
    print("   ", n, p)
print(f"\nmode-only(`0`/`0`) — 거절하면 **안** 되는 것: {len(mode_only)}")
for n, p in mode_only[:10]:
    print("   ", n, p)
