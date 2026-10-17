"""task 7.5 MEASURE — 지렛대 넷의 값을 **재서** 고른다. 아직 아무것도 안 고친다.

후보 하나에 드는 일을 네 가지 방식으로 재고, a071 의 실패하는 walk(347 후보 · 35 번들)에
곱해서 예측을 낸다. 예측은 구현 뒤 실측으로 대조한다.
"""
import hashlib
import json
import statistics
import subprocess
import time
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
ANALYSIS = ROOT / "openspec/changes/a071-wire-kr-us-protection-readiness/analysis/function-logic"

sources, digests = [], {}
for ast_path in sorted(ANALYSIS.glob("*/ast.json")):
    value = json.loads(ast_path.read_text(encoding="utf-8"))
    if value.get("revision", "current") != "current":
        continue
    src, dig = str(value.get("file", "")), str(value.get("source_sha256", ""))
    if src and dig:
        sources.append(src)
        digests[src] = dig
sources = sorted(set(sources))
CAND = subprocess.run(["git", "rev-parse", "HEAD~40"], cwd=ROOT, capture_output=True,
                      text=True, check=True).stdout.strip()
print(f"번들 소스 {len(sources)} · 후보 {CAND[:12]}")


def timed(fn, n=5):
    got = [(lambda s: (fn(), time.monotonic() - s)[1])(time.monotonic()) for _ in range(n)]
    return statistics.median(got)


def show_each():
    """오늘 — 소스마다 `git show` 한 프로세스."""
    out = {}
    for src in sources:
        p = subprocess.run(["git", "show", f"{CAND}:{src}"], cwd=ROOT,
                           capture_output=True, timeout=30, check=False)
        out[src] = None if p.returncode else p.stdout
    return out


def cat_file_batch():
    """후보 하나에 `git cat-file --batch` 한 프로세스 — 파일 단위 fetch."""
    stdin = "".join(f"{CAND}:{src}\n" for src in sources).encode()
    p = subprocess.run(["git", "cat-file", "--batch"], cwd=ROOT, input=stdin,
                       capture_output=True, timeout=60, check=False)
    out, data, pos = {}, p.stdout, 0
    for src in sources:
        nl = data.index(b"\n", pos)
        header = data[pos:nl].decode()
        if header.endswith((" missing", " ambiguous")):
            out[src] = None
            pos = nl + 1
            continue
        size = int(header.split()[2])
        out[src] = data[nl + 1:nl + 1 + size]
        pos = nl + 1 + size + 1                       # 내용 뒤에 개행 하나
    return out


def ls_tree_ids():
    """blob id 만 — 내용을 안 읽는다. sha256 을 못 만드니 이것만으로는 판정이 안 된다."""
    p = subprocess.run(["git", "ls-tree", "-z", "--full-tree", CAND, "--", *sources],
                       cwd=ROOT, capture_output=True, timeout=30, check=False)
    return len(p.stdout)


def reparse_bundles():
    """`_pinning_bundles` 가 후보마다 세 번 하는 일 — glob + JSON 파싱."""
    n = 0
    for ast_path in sorted(ANALYSIS.glob("*/ast.json")):
        json.loads(ast_path.read_text(encoding="utf-8"))
        n += 1
    return n


rows = [("show 하나씩 (오늘)", timed(show_each)),
        ("cat-file --batch 한 번", timed(cat_file_batch)),
        ("ls-tree (id 만 — 판정 불가)", timed(ls_tree_ids)),
        ("번들 재파싱 ×3 (오늘)", timed(lambda: [reparse_bundles() for _ in range(3)]))]
for label, secs in rows:
    print(f"  {label:34s} {secs * 1000:8.1f} ms/후보 → 347 후보 {secs * 347:7.1f}s")

a, b = show_each(), cat_file_batch()
same = all(a[s] == b[s] for s in sources)
digest_ok = sum(1 for s in sources
                if b[s] is not None and hashlib.sha256(b[s]).hexdigest() == digests[s])
print(f"\n두 방식이 같은 바이트: {same} · 해시 일치 {digest_ok}/{len(sources)} · "
      f"없는 파일 {sum(1 for s in sources if a[s] is None)}")
