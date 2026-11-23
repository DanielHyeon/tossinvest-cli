#!/usr/bin/env python3
"""a198 4c.8 보강 2 변이 하네스 — a198_mutate_4b.py 와 같은 절차(사본 · 시작 sha 단언 · 무변이 대조군 GREEN · 닿음 · 원복 확인 ·
한 번에 한 판). 변이 하나가 여러 자리를 고칠 수 있음(첫 자리로 닿음을 잼).

사용: a198_mutate_4c.py <repo> <expected_sha> <outdir>   (A198_ONLY="W1,W2" 로 거를 수 있음)
시험: engine(A198|A193|A094) · journal(A198) · execgw(A198) · reconcile(A094|A198) — 이 change 의 선택 시험 묶음 안.
"""
import json, os, re, subprocess, sys

repo, sha, outdir = sys.argv[1], sys.argv[2], sys.argv[3]
copy = os.path.expanduser(f"~/tossos-work/a198-mut4c-{os.getpid()}")
env = dict(os.environ)
env.pop("GOFLAGS", None)
RUN = ["go", "test", "./internal/app/engine", "./internal/journal", "./internal/execgw", "./internal/reconcile",
       "-run", "A198|A193|A094", "-count=1"]
J = "internal/journal/a198_cancel_target_close.go"
S = "internal/app/engine/in_doubt_settle.go"
C = "internal/app/engine/a198_in_doubt_cancel_close.go"
A = "internal/reconcile/acked_boot.go"

MATCH = "ON `+confirmedOrderTerminalMatch, args...)"
SW_CANCEL = "		case rec.State == journal.StateInDoubt && rec.Kind == journal.KindCancel:\n"
CLOSE_LOOP = ("	for _, rec := range cancels {\n		if err := s.closeCancel(ctx, rec); err != nil {\n			return err\n		}\n	}\n")
TAIL = "		if err := s.resolve(ctx, rec); err != nil {\n			return err\n		}\n	}\n	return nil\n}"

MUTANTS = [
    ("W1 종결 아닌 대상에도 닫기(terminal 조건 삭제)", [(J, MATCH,
        "ON `+strings.Replace(confirmedOrderTerminalMatch, \"AND f.terminal = 1\", \"\", 1), args...)")]),
    ("W2 범위 결속 삭제(거래일)", [(J, MATCH,
        "ON `+strings.Replace(confirmedOrderTerminalMatch, \"AND TRIM(f.trading_day) = TRIM(c.trading_day)\", \"\", 1), args...)")]),
    ("W3 소유 범위 삭제(계좌)", [(J, "AND TRIM(i.account_ref) = ? AND", "AND (TRIM(i.account_ref) = ? OR 1=1) AND")]),
    ("W3b 소유 범위 삭제(방향)", [(J, "AND UPPER(TRIM(i.side)) = ?\n", "AND (UPPER(TRIM(i.side)) = ? OR 1=1)\n")]),
    ("W4 소유 유일 조건 삭제", [(J, "	if owners != 1 {", "	if owners < 1 {")]),
    ("W5 판정 표 뒤집기(취소로 끝남 → 확정 거절)", [(J, "		return StateConfirmed, ReasonResolvedFound, true",
        "		return StateFailedConfirmed, ReasonResolvedAbsent, true")]),
    ("W6 fail_closed 행도 닫기", [(J, "		if failClosed != 0 || !ok ||", "		if !ok ||")]),
    ("W7 REPLACED 를 닫기", [(J, "	case brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled:",
        "	case brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled, brokerstate.StateReplaced:")]),
    ("W8 세션에서 ACKED 도 닫기", [
        (S, SW_CANCEL, "		case (rec.State == journal.StateInDoubt || rec.State == journal.StateAcked) && rec.Kind == journal.KindCancel:\n"),
        (J, "	if AttemptState(state) != StateInDoubt || MutationKind(kind) != KindCancel ||",
            "	if (AttemptState(state) != StateInDoubt && AttemptState(state) != StateAcked) || MutationKind(kind) != KindCancel ||"),
        (J, "	vals = append(vals, detail, attemptID, string(StateInDoubt))", "	vals = append(vals, detail, attemptID, state)")]),
    ("W9 D9 를 AMEND 에도", [(A, "	if rec.Kind == journal.KindCancel && strings.TrimSpace(rec.BrokerOrderID) != \"\" {",
        "	if (rec.Kind == journal.KindCancel || rec.Kind == journal.KindAmend) && strings.TrimSpace(rec.BrokerOrderID) != \"\" {")]),
    ("W10 D10 이 PLACE 를 건드리기", [
        (S, "		case rec.State == journal.StateInDoubt && rec.Kind == journal.KindPlace:\n			targets = append(targets, rec)\n" + SW_CANCEL,
            "		case rec.State == journal.StateInDoubt:\n"),
        (J, "	if AttemptState(state) != StateInDoubt || MutationKind(kind) != KindCancel ||",
            "	if AttemptState(state) != StateInDoubt ||")]),
    ("W11 닫기를 발주 해소 뒤로", [(S, CLOSE_LOOP, ""), (S, TAIL, TAIL.replace("	return nil\n}", CLOSE_LOOP + "	return nil\n}"))]),
    ("W12 source 한정 삭제", [(J, "strings.TrimSpace(target) == \"\" || producer != strings.TrimSpace(source) {",
        "strings.TrimSpace(target) == \"\" || false {")]),
    ("W13 D9 쓰기 실패를 true 로", [(A,
        "the ledger write failed: %v\", err))\n			return false\n		}\n		return true",
        "the ledger write failed: %v\", err))\n			return true\n		}\n		return true")]),
    ("W14 D9 삭제", [(A, "	if rec.Kind == journal.KindCancel && strings.TrimSpace(rec.BrokerOrderID) != \"\" {",
        "	if false && rec.Kind == journal.KindCancel && strings.TrimSpace(rec.BrokerOrderID) != \"\" {")]),
    ("W15 스윕이 브로커(해소기)를 부르기", [(C,
        "	res, err := s.deps.Ledger.CloseInDoubtCancelOnTargetClose(ctx, rec.ID, engineSource)",
        "	_, _ = s.deps.Resolver.Resolve(ctx, rec.ID)\n	res, err := s.deps.Ledger.CloseInDoubtCancelOnTargetClose(ctx, rec.ID, engineSource)")]),
    ("W16 CANCEL 에 처음 봄 알림", [(S, CLOSE_LOOP, "	for _, rec := range cancels {\n		s.noteFirstSight(ctx, rec)\n	}\n" + CLOSE_LOOP)]),
    ("W17 D10 삭제(취소를 고르지 않음)", [(S, SW_CANCEL + "			cancels = append(cancels, rec)\n", "")]),
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
    emit(step="control", rc=control.returncode, tail=control.stdout.strip().splitlines()[-4:])
    if control.returncode != 0:
        raise SystemExit("control is not GREEN — stop")
    only = os.environ.get("A198_ONLY", "")
    for name, edits in MUTANTS:
        if only and not any(name.startswith(p + " ") for p in only.split(",")):
            continue
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
            prof = os.path.join(outdir, f"cover4c-{name.split()[0]}.out")
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
