#!/usr/bin/env python3
"""a121 GREEN 변이 원장 하네스 — 사본 트리에서만 변이, 무변이 대조군 GREEN 확인 후 진행, TREE 스탬프 기록."""
import hashlib, json, os, shutil, subprocess, sys, time
from pathlib import Path

# 저장소 뿌리는 이 파일 위치에서 유도함(절대경로 고정 금지). 사용: python3 mutate.py <사본 디렉터리> <ledger.json> <mutants.json>
SRC = str(Path(__file__).resolve().parents[6])
WORK = sys.argv[1]
LEDGER = sys.argv[2]

PKG = {
    "v": (["./internal/verifylive"], [], "", ""),
    "o": (["./internal/official"], [], "", "Reconcile|Decode|TheExisting"),
    "c": (["./cmd/tossctl"], ["-tags", "tossos_testseams"], "", "Reconcile|VerifyReconcile|Mutating|LeafCommands"),
}

def tree_stamp(root):
    h = hashlib.sha256()
    for d, _, files in sorted(os.walk(root)):
        if "/.git" in d:
            continue
        for f in sorted(files):
            if f.endswith(".go") or f in ("go.mod", "go.sum"):
                p = os.path.join(d, f)
                h.update(os.path.relpath(p, root).encode())
                h.update(open(p, "rb").read())
    return h.hexdigest()

def run(pkg):
    pkgs, tags, skip, only = PKG[pkg]
    cmd = ["go", "test", "-count=1"] + tags + pkgs
    if skip:
        cmd += ["-skip", skip]
    if only:
        cmd += ["-run", only]
    r = subprocess.run(cmd, cwd=WORK, capture_output=True, text=True, env=dict(os.environ, GOFLAGS="-trimpath"))
    out = r.stdout + r.stderr
    if "build failed" in out or "[setup failed]" in out:
        return "BUILD-FAIL", out
    fails = [l.split()[2] for l in out.splitlines() if l.startswith("--- FAIL:")]
    return ("FAIL" if r.returncode != 0 else "PASS"), ",".join(fails) if fails else out[-300:]

def main():
    mutants = json.load(open(sys.argv[3]))
    if os.path.exists(WORK):
        shutil.rmtree(WORK)
    shutil.copytree(SRC, WORK, ignore=shutil.ignore_patterns(".git"))
    stamp = tree_stamp(WORK)
    assert stamp == tree_stamp(SRC), "copy differs from source"
    rows = []
    for pkg in PKG:
        status, detail = run(pkg)
        rows.append({"id": "CONTROL-" + pkg, "pkg": pkg, "result": status, "detail": detail[:200]})
        if status != "PASS":
            json.dump({"tree": stamp, "rows": rows}, open(LEDGER, "w"), indent=1, ensure_ascii=False)
            sys.exit(f"control {pkg} not GREEN: {detail[:500]}")
    for m in mutants:
        path = os.path.join(WORK, m["file"])
        orig = open(path).read()
        edits = m.get("edits") or [[m["old"], m["new"]]]
        mutated = orig
        if any(mutated.count(o) != 1 for o, _ in edits):
            rows.append({"id": m["id"], "result": "NOT-APPLIED", "detail": "anchor missing"})
            continue
        for o, n in edits:
            mutated = mutated.replace(o, n, 1)
        open(path, "w").write(mutated)
        status, detail = run(m["pkg"])
        open(path, "w").write(orig)
        verdict = {"FAIL": "CAUGHT", "PASS": "SURVIVED", "BUILD-FAIL": "BUILD-FAIL"}[status]
        rows.append({"id": m["id"], "pkg": m["pkg"], "file": m["file"], "result": verdict, "why": m.get("why", ""),
                     "detail": detail[:240]})
        print(m["id"], verdict, detail[:120], flush=True)
    assert tree_stamp(WORK) == stamp, "copy not restored"
    json.dump({"tree": stamp, "rows": rows}, open(LEDGER, "w"), indent=1, ensure_ascii=False)

main()
