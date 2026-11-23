#!/usr/bin/env python3
"""a198 4.1 변이 하네스 — 사본(git worktree, 경로에 pid)에서 한 번에 한 판.

순서: 사본 생성 → 시작 sha 단언 → 무변이 대조군 GREEN → 변이마다 (적용 → 심볼 세기로 적용 확인 → 닿음(coverprofile 로
변이 줄 블록 계수 > 0) → 시험 → 원복 → 심볼 세기로 원복 확인). 결과는 표준 출력 JSON 줄.
사용: a198_mutate.py <repo> <expected_sha> <outdir>
"""
import json, os, re, subprocess, sys

repo, sha, outdir = sys.argv[1], sys.argv[2], sys.argv[3]
copy = os.path.expanduser(f"~/tossos-work/a198-mut-{os.getpid()}")
env = dict(os.environ)
env.pop("GOFLAGS", None)  # -trimpath 는 runtime.Caller 경로 시험을 깨뜨림(4.2 관찰) — 사본에서는 끔
RUN = ["go", "test", "./internal/app/engine", "-run", os.environ.get("A198_RUN", "A198"), "-count=1"]
EXIT = "internal/app/engine/exitloop.go"
LOG = "internal/app/engine/a198_clear_cancel_log.go"

ONLY = os.environ.get("A198_ONLY", "")
MUTANTS = [
    ("M1 결함 복원(RiskGuardian 매도·PLACE 결정)", EXIT,
     """		decision, err := (&execgw.Issuer{Journal: o.opts.Journal, Clock: o.clk, AccountRef: o.opts.AccountRef}).
			IssueReduction(ctx, execgw.ReductionRequest{
				Kind:        journal.KindCancel,
				Market:      strings.ToLower(strings.TrimSpace(order.Market)),
				Symbol:      order.Symbol,
				Side:        strings.ToUpper(strings.TrimSpace(order.Side)),
				MaxQuantity: quantity,
				Reason: fmt.Sprintf("cancelling the working %s order %s before liquidating position %s",
					strings.ToUpper(strings.TrimSpace(order.Side)), order.OrderID, m.position.ID),
			})""",
     """		issuedM, err := o.opts.Issuer.IssueReduction(ctx, execgw.ReductionIssuance{
			Intent: risk.Intent{AccountRef: o.opts.AccountRef, Market: costs.Market(strings.ToLower(strings.TrimSpace(order.Market))),
				Symbol: order.Symbol, Side: risk.SideSell, Quantity: order.Quantity},
			Account: risk.AccountState{HeldQuantity: order.Quantity},
			Reason:  "mutant",
		})
		decision := issuedM.Decision""", "issuedM"),
    ("M2 결정 방향을 매도로 고정", EXIT,
     "				Side:        strings.ToUpper(strings.TrimSpace(order.Side)),\n				MaxQuantity: quantity,",
     "				Side:        \"SELL\",\n				MaxQuantity: quantity,", "Side:        \"SELL\""),
    ("M3 결정에 멱등 키가 붙는 발행(KindPlace)", EXIT,
     "				Kind:        journal.KindCancel,", "				Kind:        journal.KindPlace,", "Kind:        journal.KindPlace"),
    ("M4 실패 기록 삭제(두 호출)", EXIT,
     "			o.noteClearCancelFailure(m, order, out.State, string(out.Reason))\n", "", None),
    ("M4b 발행 실패 기록 삭제", EXIT,
     "			o.noteClearCancelFailure(m, order, \"\", exitClearCancelDecisionFailed)\n", "", None),
    ("M5 실패 기록의 중복 억제 삭제(매 주기 한 줄)", LOG,
     "	if o.clearCancelLogged[key] == mark {\n		return\n	}\n", "", None),
    ("M6 실패 줄에 원문 싣기(LiveOrder 전체)", LOG,
     "		obs.FieldReason, reason,\n	)", "		obs.FieldReason, reason,\n		\"detail\", fmt.Sprint(order),\n	)", "fmt.Sprint(order)"),
    ("M7 매도 취소 접수를 치움으로 판정", EXIT,
     "			res.cleared, res.awaitingClose = false, true\n		}\n	}\n	if !res.cleared {",
     "		}\n	}\n	if !res.cleared {", None),
]


def sh(args, cwd=copy, check=True):
    p = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True)
    if check and p.returncode != 0:
        raise SystemExit(f"command failed {args}: {p.stdout[-2000:]} {p.stderr[-2000:]}")
    return p


def emit(**kw):
    print(json.dumps(kw, ensure_ascii=False), flush=True)


subprocess.run(["git", "-C", repo, "worktree", "add", "--detach", copy, sha], check=True, capture_output=True)
try:
    top = sh(["git", "rev-parse", "--show-toplevel"]).stdout.strip()
    assert top == copy, (top, copy)
    head = sh(["git", "rev-parse", "HEAD"]).stdout.strip()
    assert head == sha, (head, sha)
    emit(step="copy", path=copy, sha=head)
    control = sh(RUN, check=False)
    emit(step="control", rc=control.returncode, tail=control.stdout.strip().splitlines()[-1:])
    if control.returncode != 0:
        raise SystemExit("control is not GREEN — stop")
    if "fmt\"" not in open(os.path.join(copy, LOG)).read():
        pass
    for name, rel, old, new, marker in MUTANTS:
        if ONLY and not name.startswith(ONLY):
            continue
        path = os.path.join(copy, rel)
        src = open(path).read()
        if src.count(old) != 1:
            emit(mutant=name, applied=False, why=f"anchor count {src.count(old)}")
            continue
        mutated = src.replace(old, new)
        extra_import = ""
        if rel == LOG and "fmt.Sprint" in new:
            mutated = mutated.replace('import (\n\t"strings"', 'import (\n\t"fmt"\n\t"strings"')
        open(path, "w").write(mutated)
        applied = (marker is None and mutated.count(old) == 0) or (marker is not None and marker in mutated)
        # 닿음: 변이 줄이 시험 실행에서 실행되는지(coverprofile 블록 계수)
        # 닿음 줄 = 원문에서 변이가 앉는 자리의 첫 줄(삭제형이면 그 바로 앞 줄). 사본의 원문 줄 번호로 원문 coverprofile 대신
        # 변이 판 coverprofile 을 읽으므로, 변이 앞 줄 수가 같은 자리(변이 위쪽)만 씀.
        idx = src.index(old)
        lineno = src[:idx].count("\n") + (1 if new.strip() else 0)
        prof = os.path.join(outdir, f"cover-{name.split()[0]}.out")
        res = sh(RUN + [f"-coverprofile={prof}"], check=False)
        reached = None
        if os.path.exists(prof):
            reached = 0
            for l in open(prof):
                m = re.match(r".*/" + re.escape(os.path.basename(rel)) + r":(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)", l)
                if m and int(m.group(1)) <= lineno <= int(m.group(2)):
                    reached = max(reached, int(m.group(3)))
        failed = [l for l in res.stdout.splitlines() if l.startswith("--- FAIL")]
        emit(mutant=name, applied=applied, line=lineno, reached=reached,
             rc=res.returncode, caught=res.returncode != 0, failed=failed,
             build_error=("[build failed]" in res.stdout) or ("cannot" in res.stdout and res.returncode == 2))
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
