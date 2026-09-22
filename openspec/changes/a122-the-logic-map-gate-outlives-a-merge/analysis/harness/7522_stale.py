"""task 7.5.22 — `revision: base` 번들의 `source_sha256` 을 base 커밋의 파일 바이트와 대조하고 change 로 귀속한다.

7.5.2.4 의 VERIFY 가 적은 귀속("a063 2 · a112 16")이 **거짓**이었다. 사람이 7.5.7(파일 수준 해시를
넣을 것인가)을 결정할 근거가 바로 이 내역이므로, 인용이 아니라 **다시 돌릴 수 있는 것**으로 남긴다
([[harness-cited-but-never-committed]]).

    python3 7522_stale.py
"""
import hashlib, json, subprocess, sys
from collections import Counter
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
CH = ROOT / "openspec" / "changes"


def change_dir(p: Path) -> Path | None:
    """번들 파일에서 그것이 속한 change 디렉터리까지 거슬러 올라간다."""
    for parent in p.parents:
        if parent.parent == CH or parent.parent == CH / "archive":
            return parent
    return None


total = base = current = other = 0
match = []
stale = []
missing = []
for p in CH.rglob("ast.json"):
    total += 1
    try:
        d = json.loads(p.read_text(encoding="utf-8"))
    except Exception:
        continue
    rev = d.get("revision")
    if rev == "current":
        current += 1
        continue
    if rev != "base":
        other += 1
        continue
    base += 1
    cd = change_dir(p)
    bc = cd / "base-commit.txt" if cd else None
    if bc is None or not bc.exists():
        missing.append((p, "no base-commit.txt"))
        continue
    commit = bc.read_text(encoding="utf-8").strip()
    r = subprocess.run(
        ["git", "show", f"{commit}:{d['file']}"],
        cwd=ROOT, capture_output=True,
    )
    if r.returncode != 0:
        missing.append((p, f"git show rc={r.returncode}"))
        continue
    got = hashlib.sha256(r.stdout).hexdigest()
    (match if got == d.get("source_sha256") else stale).append((cd.name, p, d["file"]))

print(f"ast.json 전수 {total} · base {base} · current {current} · 그 밖 {other}")
print(f"MATCH {len(match)} · STALE {len(stale)} · 못 잼 {len(missing)}")
print("\n=== STALE 귀속 ===")
arch = {d.name for d in (CH / "archive").iterdir() if d.is_dir()}
c = Counter()
for name, p, f in stale:
    c[("archive " if name in arch else "") + name] += 1
for k, v in sorted(c.items(), key=lambda kv: -kv[1]):
    print(f"  {v:3d}  {k}")
print("\n=== 못 잰 것 ===")
for p, why in missing[:10]:
    print(" ", why, p.relative_to(ROOT))
