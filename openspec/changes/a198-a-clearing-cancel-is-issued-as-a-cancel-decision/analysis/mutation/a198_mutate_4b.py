#!/usr/bin/env python3
"""a198 4b.8 보강 변이 하네스 — a198_mutate.py 와 같은 절차(사본 · 시작 sha 단언 · 무변이 대조군 GREEN · 닿음 · 원복 확인 · 한 번에 한 판).

사용: a198_mutate_4b.py <repo> <expected_sha> <outdir>
시험: engine(A198|A094) + execgw(A198|A165) — 둘 다 이 change 의 선택 시험 묶음 안.
"""
import json, os, re, subprocess, sys

repo, sha, outdir = sys.argv[1], sys.argv[2], sys.argv[3]
copy = os.path.expanduser(f"~/tossos-work/a198-mut4b-{os.getpid()}")
env = dict(os.environ)
env.pop("GOFLAGS", None)
RUN = ["go", "test", "./internal/app/engine", "./internal/execgw", "-run", "A198|A094|A165", "-count=1"]
EXIT = "internal/app/engine/exitloop.go"
LOG = "internal/app/engine/a198_clear_cancel_log.go"
CR = "internal/execgw/cancel_refusal.go"
SET = "internal/execgw/a198_settled_cancel_refusal.go"

MUTANTS = [
    ("V1 D7 을 PLACE·AMEND 에도(취소 정책 kind 검사 삭제)", CR,
     "	if g.cancelRefusal == nil || plan.kind != journal.KindCancel {",
     "	if g.cancelRefusal == nil {", None),
    ("V2 부분문자열 일치로 완화(already-f 접두)", SET,
     "	if !code.present || !settledOrderCancelCodes[strings.ToLower(code.value)] {",
     "	if !code.present || (!settledOrderCancelCodes[strings.ToLower(code.value)] && !strings.HasPrefix(strings.ToLower(code.value), \"already-f\")) {", None),
    ("V3 status 409 전체를 확정 거절로", SET,
     "	top, nested, ok := refusalCodes(apiErr.Body)",
     "	if apiErr.Code == 409 {\n		out.Class = journal.DispatchRejected\n		out.ReasonCode = journal.ReasonDispatchRejected\n		return out, true\n	}\n	top, nested, ok := refusalCodes(apiErr.Body)", None),
    ("V4 already-modified 를 D7 에", SET,
     "	\"already-rejected\": true,\n}",
     "	\"already-rejected\": true,\n	\"already-modified\": true,\n}", None),
    ("V5 D7 을 flatten 정책에도", CR,
     "		if policy.Definitive {\n			if settled, isSettled := settledCancelRefusal(first); isSettled {",
     "		if true {\n			if settled, isSettled := settledCancelRefusal(first); isSettled {", None),
    ("V6 D7 삭제", CR,
     "			if settled, isSettled := settledCancelRefusal(first); isSettled {\n				return settled\n			}\n",
     "", None),
    ("V7 D8 삭제(접수 확정 매수도 다시 취소)", EXIT,
     "		if accepted {\n			if !buy {",
     "		if accepted && !buy {\n			if !buy {", None),
    ("V8 D8 을 매도에도 「치움」으로 확대(초과 매도 방향)", EXIT,
     "		if accepted {\n			if !buy {\n				res.cleared, res.awaitingClose = false, true\n			}\n			continue\n		}",
     "		if accepted {\n			continue\n		}", None),
    ("N10 발행 실패에도 cleared 유지", EXIT,
     "			o.noteClearCancelFailure(m, order, \"\", exitClearCancelDecisionFailed)\n			res.cleared = false\n",
     "			o.noteClearCancelFailure(m, order, \"\", exitClearCancelDecisionFailed)\n", None),
    ("N2 결정 최대 수량 ×2", EXIT,
     "				MaxQuantity: quantity,", "				MaxQuantity: quantity * 2,", None),
    ("N6 접수 확정 뒤 표시 정리 삭제", EXIT,
     "		o.noteClearCancelConfirmed(m, order)\n", "", None),
    ("N9 해석 실패 줄 삭제", EXIT,
     "			o.noteClearCancelFailure(m, order, \"\", exitClearCancelOrderUnreadable)\n", "", None),
    ("N11 표시를 쓰기 앞에", LOG,
     "	// 쓰기 대상이 패닉해도 손절 루프는 계속 돎",
     "	if o.clearCancelLogged == nil {\n		o.clearCancelLogged = map[string]string{}\n	}\n	o.clearCancelLogged[key] = mark\n	// 쓰기 대상이 패닉해도 손절 루프는 계속 돎", None),
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
    for name, rel, old, new, marker in MUTANTS:
        path = os.path.join(copy, rel)
        src = open(path).read()
        if src.count(old) != 1:
            emit(mutant=name, applied=False, why=f"anchor count {src.count(old)}")
            continue
        mutated = src.replace(old, new)
        open(path, "w").write(mutated)
        idx = src.index(old)
        lineno = src[:idx].count("\n") + (1 if new.strip() else 0)
        prof = os.path.join(outdir, f"cover4b-{name.split()[0]}.out")
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
        emit(mutant=name, applied=True, line=lineno, reached=reached, rc=res.returncode,
             caught=res.returncode != 0 and not build, build_error=build, failed=failed)
        open(path, "w").write(src)
        restored = open(path).read() == src and sh(["git", "diff", "--quiet", "--", rel], check=False).returncode == 0
        emit(mutant=name, restored=restored)
        if not restored:
            raise SystemExit("restore failed — stop")
    final = sh(RUN, check=False)
    emit(step="final-control", rc=final.returncode)
finally:
    subprocess.run(["git", "-C", repo, "worktree", "remove", "--force", copy], capture_output=True)
    emit(step="cleanup", removed=not os.path.exists(copy))
