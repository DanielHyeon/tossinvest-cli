"""a195 변이 원장 하네스 — 사본에서 한 번에 한 변이씩 돌림.

사용: python3 mutate.py <repo-worktree> <expected-sha> <out-ledger>
- 사본: git archive <sha> 를 ~/tossos-work/a195-mut-<pid>/ 에 풀어 씀(worktree 는 안 건드림).
- 시작 sha 단언: worktree HEAD 가 expected-sha 와 같아야 함.
- 무변이 대조군이 GREEN 이 아니면 멈춤. 변이마다 앵커가 정확히 1회인지 단언, 원복 뒤 sha256 이 원본과 같은지 단언.
"""
import hashlib, os, shutil, subprocess, sys, tarfile, io

repo, expected, ledger = sys.argv[1:4]
head = subprocess.check_output(["git", "-C", repo, "rev-parse", "HEAD"], text=True).strip()
assert head == expected, f"start sha {head} != {expected}"
copy = os.path.expanduser(f"~/tossos-work/a195-mut-{os.getpid()}")
assert not os.path.exists(copy)
os.makedirs(copy)
data = subprocess.check_output(["git", "-C", repo, "archive", expected])
tarfile.open(fileobj=io.BytesIO(data)).extractall(copy)

env = dict(os.environ, GOFLAGS="-trimpath -buildvcs=false")
SUITES = [
    ["go", "test", "./cmd/tossctl", "-run", "TestA166EngineReconcile|TestEngineReconcileResolve|A195", "-count=1"],
    ["go", "test", "./internal/accountbudgetruntime", "-run", "TestA166WiringCensus", "-count=1"],
]

def run_suites():
    results = []
    for cmd in SUITES:
        p = subprocess.run(cmd, cwd=copy, env=env, capture_output=True, text=True)
        out = p.stdout + p.stderr
        fails = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL:")]
        results.append((" ".join(cmd[2:3]), p.returncode, fails, out))
    return results

CAUSES = "cmd/tossctl/reconcile_resolve_causes.go"
BUILD = "cmd/tossctl/engine_reconcile.go"
MUTANTS = [
    ("M1 원인 판정 삭제(헬퍼가 항상 통과)", CAUSES,
     "func reconcileResolveRefusedCause(blocks []reconcile.Block) error {\n",
     "func reconcileResolveRefusedCause(blocks []reconcile.Block) error {\n\treturn nil\n"),
    ("M2 허용 목록에 IDENTIFIER_CONFLICT 추가", CAUSES,
     "case block.Cause == journal.ReconcileCauseQuantityMismatch:",
     "case block.Cause == journal.ReconcileCauseQuantityMismatch || block.Cause == journal.ReconcileCauseIdentifierConflict:"),
    ("M3 종목 범위 조건 삭제(계좌 전체 AF 허용)", CAUSES,
     "block.Scope == reconcile.ScopeSymbol && strings.TrimSpace(block.Symbol) != \"\":",
     "(strings.TrimSpace(block.Symbol) != \"\" || block.Scope != reconcile.ScopeSymbol):"),
    ("M4 ATTRIBUTION_FAILED 를 허용 목록에서 빼기", CAUSES,
     "case block.Cause == journal.ReconcileCauseAttributionFailed &&",
     "case false && block.Cause == journal.ReconcileCauseAttributionFailed &&"),
    ("M5 첫 허용 원인에서 즉시 통과 반환", CAUSES,
     "\t\t\tcontinue\n", "\t\t\treturn nil\n"),
    ("M6 validate 클로저가 헬퍼 결과를 무시", BUILD,
     "\t\t\treturn reconcileResolveRefusedCause(blocks)\n",
     "\t\t\t_ = reconcileResolveRefusedCause(blocks)\n\t\t\treturn nil\n"),
]

lines = [f"# a195 변이 원장 — 시작 sha {expected}(단언 통과), 사본 {copy}, GOFLAGS='-trimpath -buildvcs=false'",
         "# 스위트: " + " ; ".join(" ".join(c) for c in SUITES)]
control = run_suites()
lines.append("## 무변이 대조군: " + ", ".join(f"{name} rc={rc}" for name, rc, _, _ in control))
if any(rc != 0 for _, rc, _, _ in control):
    lines.append("대조군 GREEN 아님 — 중단")
    for *_, out in control: lines.append(out)
    open(ledger, "w").write("\n".join(lines) + "\n"); sys.exit(1)

for name, rel, old, new in MUTANTS:
    path = os.path.join(copy, rel)
    original = open(path, encoding="utf-8").read()
    digest = hashlib.sha256(original.encode()).hexdigest()
    count = original.count(old)
    # M5 는 continue 두 자리 모두(QM·AF 갈래)를 바꿈 — 그 밖의 변이는 앵커가 정확히 1회여야 함.
    want = 2 if name.startswith("M5") else 1
    assert count == want, f"{name}: anchor count {count} != {want}"
    open(path, "w", encoding="utf-8").write(original.replace(old, new))
    results = run_suites()
    open(path, "w", encoding="utf-8").write(original)
    assert hashlib.sha256(open(path, encoding="utf-8").read().encode()).hexdigest() == digest, f"{name}: revert mismatch"
    killed = [f for _, rc, fails, _ in results for f in fails]
    build_fail = [n for n, rc, fails, out in results if rc != 0 and not fails]
    verdict = "CAUGHT" if any(rc != 0 for _, rc, _, _ in results) else "SURVIVED"
    lines.append(f"## {name} — {verdict}")
    lines.append(f"- 파일 {rel}, 앵커 {count}회, 원복 sha256 일치")
    for suite, rc, fails, out in results:
        lines.append(f"- {suite}: rc={rc}, 실패 시험 {len(fails)}: " + "; ".join(fails))
        if rc != 0 and not fails:
            lines.append("  (시험 실패 줄 없음 — 빌드/vet 실패 원문)\n" + "\n".join("    " + l for l in out.splitlines()[:15]))

final = run_suites()
lines.append("## 원복 뒤 대조군: " + ", ".join(f"{n} rc={rc}" for n, rc, _, _ in final))
open(ledger, "w").write("\n".join(lines) + "\n")
shutil.rmtree(copy)
print("\n".join(lines))
