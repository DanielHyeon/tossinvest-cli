#!/usr/bin/env python3
"""a198 4d 변이 하네스(gstack review r2 G1 — 취소 실패 줄을 주기 끝에 씀). a198_mutate_4c.py 와 같은 절차:
사본 worktree · 시작 sha 단언 · 무변이 대조군 GREEN · 변이마다 닿음(coverprofile) · 원복 확인 · 끝 대조군 · 사본 제거.

사용: a198_mutate_4d.py <repo> <expected_sha> <outdir>
시험: engine -run 'A198|A129'(GOFLAGS 없음).
"""
import json, os, re, subprocess, sys

repo, sha, outdir = sys.argv[1], sys.argv[2], sys.argv[3]
copy = os.path.expanduser(f"~/tossos-work/a198-mut4d-{os.getpid()}")
env = dict(os.environ)
env.pop("GOFLAGS", None)
RUN = ["go", "test", "./internal/app/engine", "-run", "A198|A129", "-count=1"]
L = "internal/app/engine/a198_clear_cancel_log.go"
E = "internal/app/engine/exit_order_log.go"

MUTANTS = [
    ("X1 넣지 않고 바로 쓰기(이전 구현)", [(L,
        "	o.clearCancelPending = append(o.clearCancelPending, clearCancelLine{",
        "	_ = writeClearCancelLine(o.opts.ExitOrderLog, clearCancelLine{")]),
    ("X2 표시를 쓰기 앞에", [(L,
        "		if !writeClearCancelLine(logger, l) {\n			continue\n		}\n		if o.clearCancelLogged == nil {\n			o.clearCancelLogged = map[string]string{}\n		}\n		o.clearCancelLogged[l.key] = l.mark\n",
        "		if o.clearCancelLogged == nil {\n			o.clearCancelLogged = map[string]string{}\n		}\n		o.clearCancelLogged[l.key] = l.mark\n		writeClearCancelLine(logger, l)\n")]),
    ("X3 비우기 호출 삭제", [(E, "	o.flushClearCancelFailures()\n", "")]),
    ("X4 대기열을 비우지 않고 다시 씀", [(L, "	pending := o.clearCancelPending\n	o.clearCancelPending = nil\n",
        "	pending := o.clearCancelPending\n")]),
]


def sh(args, cwd=None, check=True):
    p = subprocess.run(args, cwd=cwd or copy, env=env, capture_output=True, text=True)
    if check and p.returncode != 0:
        raise SystemExit(f"command failed {args}: {p.stdout[-2000:]} {p.stderr[-2000:]}")
    return p


def emit(**kw):
    print(json.dumps(kw, ensure_ascii=False), flush=True)


subprocess.run(["git", "-C", repo, "worktree", "add", "--detach", copy, sha], check=True, capture_output=True)
try:
    assert sh(["git", "rev-parse", "--show-toplevel"]).stdout.strip() == copy
    head = sh(["git", "rev-parse", "HEAD"]).stdout.strip()
    assert head == sha, (head, sha)
    emit(step="copy", path=copy, sha=head)
    control = sh(RUN, check=False)
    emit(step="control", rc=control.returncode, tail=control.stdout.strip().splitlines()[-2:])
    if control.returncode != 0:
        raise SystemExit("control is not GREEN — stop")
    for name, edits in MUTANTS:
        originals, ok, first = {}, True, None
        for rel, old, new in edits:
            path = os.path.join(copy, rel)
            src = open(path).read()
            originals.setdefault(rel, src)
            if src.count(old) != 1:
                ok = False
                emit(mutant=name, applied=False, why=f"anchor count {src.count(old)} in {rel}")
                break
            if first is None:
                first = (rel, src[:src.index(old)].count("\n") + 1)
            open(path, "w").write(src.replace(old, new))
        if ok:
            rel, lineno = first
            prof = os.path.join(outdir, f"cover4d-{name.split()[0]}.out")
            res = sh(RUN + [f"-coverprofile={prof}"], check=False)
            reached = None
            if os.path.exists(prof):
                reached = 0
                for l in open(prof):
                    m = re.match(r".*/" + re.escape(os.path.basename(rel)) + r":(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)", l)
                    if m and int(m.group(1)) <= lineno <= int(m.group(2)):
                        reached = max(reached, int(m.group(3)))
            failed = [l.split()[2] for l in res.stdout.splitlines() if l.startswith("--- FAIL")]
            build = "[build failed]" in res.stdout or "[setup failed]" in res.stdout
            emit(mutant=name, applied=True, file=rel, line=lineno, reached=reached, rc=res.returncode,
                 caught=res.returncode != 0 and not build, build_error=build, failed=failed)
        for rel, src in originals.items():
            open(os.path.join(copy, rel), "w").write(src)
        restored = all(open(os.path.join(copy, rel)).read() == src for rel, src in originals.items()) and \
            sh(["git", "diff", "--quiet"], check=False).returncode == 0
        emit(mutant=name, restored=restored)
        if not restored:
            raise SystemExit("restore failed — stop")
    final = sh(RUN, check=False)
    emit(step="final-control", rc=final.returncode)
finally:
    subprocess.run(["git", "-C", repo, "worktree", "remove", "--force", copy], capture_output=True)
    emit(step="cleanup", removed=not os.path.exists(copy))
