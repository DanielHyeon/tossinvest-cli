#!/usr/bin/env python3
"""a189 tasks 3.1 사본 변이 하네스(양성 대조군).

HEAD 를 git archive 로 pid 붙은 임시 디렉터리에 풀고 그 사본에서만 변이함(작업 트리는 건드리지 않음).
사본에는 go.mod·go.sum·cmd·internal·tools·docs/migration 과 openspec 아래 추적된 .go 파일을 넣음(A112 가드의 저장소 순회 대상과 같게).
docs/migration 은 a140 TestA169SecurityTypeOpenAPIEnum 이 ../../docs/migration/openapi.latest.json 을 읽기 때문임
(1차 대조군이 그 파일 없음으로 FAIL — history-1-control-missing-docs-migration.txt).
- 작업 트리의 변이 대상 파일이 HEAD 와 같은지 먼저 단언함(커밋 뒤 돌릴 것).
- 무변이 대조군이 GREEN 이 아니면 멈춤.
- 변이마다: 대상 문자열이 정확히 1회인지 단언 → 적용 → 변이 문자열이 사본에 들어갔는지 단언 →
  변이 파일 패키지를 시험 파일까지 컴파일(-run '^$')해 빌드 실패를 CAUGHT 로 세지 않음 → 시험 → 원복 후 SHA-256 단언.
- -trimpath 는 쓰지 않음: 두 도구의 nonTestSources 가 runtime.Caller 경로로 소스를 읽으므로 -trimpath 에서는
  사본 소스를 못 찾음(a187 history-1-control-trimpath.txt 의 같은 기전). 대신 전용 GOCACHE 를 쓰고 끝에 지움.
사용: python3 -I mutation.py [ID…]
"""
from __future__ import annotations

import hashlib
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

MOD = "github.com/JungHoonGhae/tossinvest-cli/internal/"
A128 = "tools/a128-jev-shadow-probe/main.go"
A140 = "tools/a140-breakout-signal-shadow/main.go"
GUARD = "internal/official/a112_mbus_static_test.go"
RUNTIME_TEST = "internal/accountbudgetruntime/runtime_test.go"
PRIO = "\tctx = accountbudget.WithPriority(ctx, accountbudget.Auxiliary)\n"
IMP_AB = f'\t"{MOD}accountbudget"\n'
IMP_RT = f'\t"{MOD}accountbudgetruntime"\n'
EXACT = 'path == "internal/accountbudgetruntime/v2_token_test.go"'

# 시험 묶음: (패키지, -run 정규식 또는 None=전부)
TOOL_TESTS = {A128: ("./tools/a128-jev-shadow-probe/", None), A140: ("./tools/a140-breakout-signal-shadow/", None)}
A112_TESTS = ("./internal/official/", "^(TestA112MBUS|TestA189)")


def tool_mutants(path: str, tag: str) -> list:
    pkg = TOOL_TESTS[path]
    after = lambda extra: [("once", PRIO, PRIO + extra)]
    return [
        (f"{tag}-a-budget-name", path, after("\t_ = accountbudget.Critical\n"), pkg, "목록 밖 accountbudget 이름"),
        (f"{tag}-a-runtime-name", path, after("\t_ = accountbudgetruntime.Config{}\n"), pkg, "목록 밖 accountbudgetruntime 이름"),
        (f"{tag}-c-official-with", path, after('\t_ = official.WithBaseURL("http://mutant")\n'), pkg, "목록 밖 official.With… 호출(D5)"),
        (f"{tag}-d-forbidden", path, after("\t_ = client.PlaceOrder\n"), pkg, "금지 이름 PlaceOrder 선택자"),
        (f"{tag}-e-import", path, [("once", IMP_RT, IMP_RT + f'\t_ "{MOD}orderintent"\n')], pkg, "목록 밖 내부 패키지 import"),
        (f"{tag}-f-no-accountbudget", path, [("once", PRIO, ""), ("once", IMP_AB, "")], pkg, "accountbudget 사용 전부 제거(본 것 없음)"),
        (f"{tag}-f-alias-runtime", path, [("once", IMP_RT, f'\tabr "{MOD}accountbudgetruntime"\n'), ("all", "accountbudgetruntime.", "abr.")], pkg,
         "accountbudgetruntime 를 별칭으로만 씀(본 것 없음)"),
    ]


MUTANTS = tool_mutants(A128, "a128") + tool_mutants(A140, "a140") + [
    ("g1-planted-ref", RUNTIME_TEST, [("append", "", "\nfunc a189MutantRef() { _ = official.A112MBUSCandle }\n")], A112_TESTS,
     "허용 밖 시험 파일(runtime_test.go)에 M-B0 참조"),
    ("g2-prefix-widen", GUARD, [("once", EXACT, 'strings.HasPrefix(path, "internal/accountbudgetruntime/")')], A112_TESTS,
     "정확 경로 비교를 디렉터리 접두사로 넓힘"),
    ("g3-suffix-widen", GUARD, [("once", EXACT, 'strings.HasSuffix(path, "/v2_token_test.go")')], A112_TESTS,
     "정확 경로 비교를 파일 이름 접미사로 넓힘"),
    ("g4-revert", GUARD, [("once", " || " + EXACT, "")], A112_TESTS, "허용 경로 추가를 되돌림(base 동작)"),
]


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def go(base: Path, args: list[str], cache: Path) -> tuple[int, str]:
    env = {k: v for k, v in os.environ.items() if k != "GOFLAGS"}
    env["GOCACHE"] = str(cache)
    proc = subprocess.run(["go", *args], cwd=base, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    return proc.returncode, proc.stdout


def run_tests(base: Path, tests, cache: Path, verbose: bool = False) -> tuple[int, str]:
    pkg, run = tests
    args = ["test", "-count=1", *(["-v"] if verbose else []), *(["-run", run] if run else []), pkg]
    return go(base, args, cache)


def summary(out: str) -> str:
    fails = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL")]
    msgs = [l.strip() for l in out.splitlines() if re.match(r"^\s+\S+_test\.go:\d+: ", l)][:6]
    return "\n".join(["  잡은 시험:"] + ["    " + f for f in fails] + ["  실패 원문(앞 6줄):"] + ["    " + m for m in msgs])


def apply(original: str, edits) -> str:
    text = original
    for kind, old, new in edits:
        if kind == "append":
            text = text + new
        elif kind == "once":
            n = text.count(old)
            assert n == 1, f"TARGET-COUNT {n}: {old!r}"
            text = text.replace(old, new, 1)
        else:
            assert text.count(old) >= 1, f"TARGET-COUNT 0: {old!r}"
            text = text.replace(old, new)
    for kind, old, new in edits:
        if new:
            assert new in text, f"변이 문자열이 사본에 없음: {new!r}"
    return text


def main() -> int:
    root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
    head = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    wanted = set(sys.argv[1:])
    selected = [m for m in MUTANTS if not wanted or m[0] in wanted]
    base = Path(tempfile.mkdtemp(prefix=f"a189-mut-{os.getpid()}-"))
    cache = Path(tempfile.mkdtemp(prefix=f"a189-mut-gocache-{os.getpid()}-"))
    try:
        extra = subprocess.check_output(["git", "-C", str(root), "ls-files", "openspec/*.go"], text=True).split()
        archive = subprocess.run(["git", "-C", str(root), "archive", "HEAD", "go.mod", "go.sum", "cmd", "internal", "tools", "docs/migration", *extra],
                                 check=True, stdout=subprocess.PIPE).stdout
        subprocess.run(["tar", "-x", "-C", str(base)], input=archive, check=True)
        assert (base / "go.mod").is_file() and not (base / ".git").exists()
        files = sorted({m[1] for m in selected})
        originals = {f: (base / f).read_text() for f in files}
        starts = {f: sha(base / f) for f in files}
        for f in files:
            assert starts[f] == sha(root / f), f"작업 트리 {f} 가 HEAD 와 다름 — 커밋 뒤 돌릴 것"
        print(f"기준 HEAD {head}, 사본 {base.name}(TMPDIR 아래), 전용 GOCACHE")
        for f in files:
            print(f"  {f} sha256={starts[f]}")
        groups = sorted({m[3] for m in selected})
        for tests in groups:
            status, out = run_tests(base, tests, cache, verbose=True)
            if status != 0:
                print(out)
                print(f"control {tests}: FAIL — 변이 판정 중단")
                return 1
            print(f"control {tests}: GREEN ({out.count('--- PASS')} PASS 줄, FAIL 0)")
        issues = 0
        for mid, fname, edits, tests, label in selected:
            target, original, start = base / fname, originals[fname], starts[fname]
            try:
                mutated = apply(original, edits)
            except AssertionError as exc:
                print(f"{mid}: TARGET — {label}: {exc}")
                issues += 1
                continue
            target.write_text(mutated)
            pkg_dir = "./" + str(Path(fname).parent) + "/"
            bstatus, bout = go(base, ["test", "-count=1", "-run", "^$", pkg_dir], cache)
            status, out = (1, "") if bstatus != 0 else run_tests(base, tests, cache)
            target.write_text(original)
            assert sha(target) == start, f"{fname} 원복 실패"
            if bstatus != 0:
                print(f"{mid}: INVALID(빌드 실패) — {label}")
                print("    " + "\n    ".join(bout.splitlines()[:6]))
                issues += 1
            elif status == 0:
                print(f"{mid}: SURVIVED — {label}")
                issues += 1
            else:
                print(f"{mid}: CAUGHT — {label}  (컴파일 {pkg_dir} ok)")
                print(summary(out))
        for f in files:
            assert sha(base / f) == starts[f]
        print(f"끝: 원복 sha 전부 일치, 판정 이상 {issues}건")
        return 1 if issues else 0
    finally:
        shutil.rmtree(base, ignore_errors=True)
        shutil.rmtree(cache, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
