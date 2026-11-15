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
- gstack 리뷰 반영(2026-11-15): 변이마다 「잡아야 할 시험」 이름을 두고, 그 시험이 `--- FAIL` 로 찍혀야만 CAUGHT 로 셈
  (다른 시험만 실패하면 WRONG-CATCH). 무결성 검사는 assert 대신 명시적 Fail 예외(python -O 에서도 유지).
  알 수 없는 ID 는 사본을 만들기 전에 거절. go 실행은 HOME·XDG_* 를 임시 디렉터리로 격리함 — 기존 a166 도구 시험이
  기본 토큰 캐시 경로(~/.cache/tossctl)를 읽으므로 운영자 토큰을 읽지 않게 함(모듈 캐시·go env 파일은 원래 값을 명시 전달).
사용: python3 -I mutation.py [ID…]
"""
from __future__ import annotations

import hashlib
import os
import re
import shutil
import subprocess
import sys
import json
import tempfile
from pathlib import Path


class Fail(Exception):
    """하네스 무결성 실패(assert 대신 — python -O 에서도 사라지지 않게)."""


def check(cond: bool, msg: str) -> None:
    if not cond:
        raise Fail(msg)

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


# 도구별 「잡아야 할 시험」: (이름 시험, import 시험)
GUARD_TESTS = {A128: ("TestNoMutatingOrAccountSurfaceIsNamed", "TestImportsStayInsideTheAllowlist"),
               A140: ("TestNoMutatingAccountOrWriteSurfaceIsNamed", "TestImportsStayInsideTheAllowlist")}


def tool_mutants(path: str, tag: str) -> list:
    pkg = TOOL_TESTS[path]
    names, imports = GUARD_TESTS[path]
    after = lambda extra: [("once", PRIO, PRIO + extra)]
    return [
        (f"{tag}-a-budget-name", path, after("\t_ = accountbudget.Critical\n"), pkg, "목록 밖 accountbudget 이름", names),
        (f"{tag}-a-runtime-name", path, after("\t_ = accountbudgetruntime.Config{}\n"), pkg, "목록 밖 accountbudgetruntime 이름", names),
        (f"{tag}-c-official-with", path, after('\t_ = official.WithBaseURL("http://mutant")\n'), pkg, "목록 밖 official.With… 호출(D5)", names),
        (f"{tag}-d-forbidden", path, after("\t_ = client.PlaceOrder\n"), pkg, "금지 이름 PlaceOrder 선택자", names),
        (f"{tag}-e-import", path, [("once", IMP_RT, IMP_RT + f'\t_ "{MOD}orderintent"\n')], pkg, "목록 밖 내부 패키지 import", imports),
        (f"{tag}-f-no-accountbudget", path, [("once", PRIO, ""), ("once", IMP_AB, "")], pkg, "accountbudget 사용 전부 제거(본 것 없음)", names),
        (f"{tag}-f-alias-runtime", path, [("once", IMP_RT, f'\tabr "{MOD}accountbudgetruntime"\n'), ("all", "accountbudgetruntime.", "abr.")], pkg,
         "accountbudgetruntime 를 별칭으로만 씀(본 것 없음)", names),
    ]


MUTANTS = tool_mutants(A128, "a128") + tool_mutants(A140, "a140") + [
    ("g1-planted-ref", RUNTIME_TEST, [("append", "", "\nfunc a189MutantRef() { _ = official.A112MBUSCandle }\n")], A112_TESTS,
     "허용 밖 시험 파일(runtime_test.go)에 M-B0 참조", "TestA112MBUSStaticNoProductCallerOrForbiddenPath"),
    ("g2-prefix-widen", GUARD, [("once", EXACT, 'strings.HasPrefix(path, "internal/accountbudgetruntime/")')], A112_TESTS,
     "정확 경로 비교를 디렉터리 접두사로 넓힘", "TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile"),
    ("g3-suffix-widen", GUARD, [("once", EXACT, 'strings.HasSuffix(path, "/v2_token_test.go")')], A112_TESTS,
     "정확 경로 비교를 파일 이름 접미사로 넓힘", "TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile"),
    ("g4-revert", GUARD, [("once", " || " + EXACT, "")], A112_TESTS, "허용 경로 추가를 되돌림(base 동작)",
     "TestA112MBUSStaticNoProductCallerOrForbiddenPath"),
]


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


_GO_ENV: dict[str, str] = {}


def isolated_env(cache: Path) -> dict[str, str]:
    # 운영자 HOME 을 감추되 모듈 캐시·go env 파일은 원래 자리를 명시함(오프라인 모듈 해석 유지).
    if not _GO_ENV:
        _GO_ENV.update(json.loads(subprocess.check_output(["go", "env", "-json", "GOPATH", "GOMODCACHE", "GOENV"], text=True)))
    home = cache / "home"
    home.mkdir(exist_ok=True)
    env = {k: v for k, v in os.environ.items() if k != "GOFLAGS" and not k.startswith("XDG_")}
    env.update(_GO_ENV)
    env.update(HOME=str(home), XDG_CACHE_HOME=str(home / ".cache"), XDG_CONFIG_HOME=str(home / ".config"),
               XDG_DATA_HOME=str(home / ".local/share"), GOCACHE=str(cache / "go-build"))
    return env


def go(base: Path, args: list[str], cache: Path) -> tuple[int, str]:
    env = isolated_env(cache)
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
            check(n == 1, f"TARGET-COUNT {n}: {old!r}")
            text = text.replace(old, new, 1)
        else:
            check(text.count(old) >= 1, f"TARGET-COUNT 0: {old!r}")
            text = text.replace(old, new)
    for kind, old, new in edits:
        if new:
            check(new in text, f"변이 문자열이 사본에 없음: {new!r}")
    return text


def main() -> int:
    root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
    head = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    wanted = set(sys.argv[1:])
    unknown = sorted(wanted - {m[0] for m in MUTANTS})
    if unknown:
        print(f"알 수 없는 변이 ID: {unknown}")
        return 2
    selected = [m for m in MUTANTS if not wanted or m[0] in wanted]
    base = Path(tempfile.mkdtemp(prefix=f"a189-mut-{os.getpid()}-"))
    cache = Path(tempfile.mkdtemp(prefix=f"a189-mut-gocache-{os.getpid()}-"))
    try:
        extra = subprocess.check_output(["git", "-C", str(root), "ls-files", "openspec/*.go"], text=True).split()
        archive = subprocess.run(["git", "-C", str(root), "archive", "HEAD", "go.mod", "go.sum", "cmd", "internal", "tools", "docs/migration", *extra],
                                 check=True, stdout=subprocess.PIPE).stdout
        subprocess.run(["tar", "-x", "-C", str(base)], input=archive, check=True)
        check((base / "go.mod").is_file() and not (base / ".git").exists(), "사본 모양이 다름")
        files = sorted({m[1] for m in selected})
        originals = {f: (base / f).read_text() for f in files}
        starts = {f: sha(base / f) for f in files}
        for f in files:
            check(starts[f] == sha(root / f), f"작업 트리 {f} 가 HEAD 와 다름 — 커밋 뒤 돌릴 것")
        print(f"기준 HEAD {head}, 사본 {base.name}(TMPDIR 아래), 전용 GOCACHE, HOME·XDG 격리")
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
        for mid, fname, edits, tests, label, expect in selected:
            target, original, start = base / fname, originals[fname], starts[fname]
            try:
                mutated = apply(original, edits)
            except Fail as exc:
                print(f"{mid}: TARGET — {label}: {exc}")
                issues += 1
                continue
            target.write_text(mutated)
            pkg_dir = "./" + str(Path(fname).parent) + "/"
            bstatus, bout = go(base, ["test", "-count=1", "-run", "^$", pkg_dir], cache)
            status, out = (1, "") if bstatus != 0 else run_tests(base, tests, cache)
            target.write_text(original)
            check(sha(target) == start, f"{fname} 원복 실패")
            if bstatus != 0:
                print(f"{mid}: INVALID(빌드 실패) — {label}")
                print("    " + "\n    ".join(bout.splitlines()[:6]))
                issues += 1
            elif status == 0:
                print(f"{mid}: SURVIVED — {label}")
                issues += 1
            elif f"--- FAIL: {expect} (" not in out:
                print(f"{mid}: WRONG-CATCH — {label}: 기대 시험 {expect} 가 실패 목록에 없음")
                print(summary(out))
                issues += 1
            else:
                print(f"{mid}: CAUGHT — {label}  (컴파일 {pkg_dir} ok, 기대 시험 {expect} FAIL 확인)")
                print(summary(out))
        for f in files:
            check(sha(base / f) == starts[f], f"{f} 끝 원복 sha 불일치")
        print(f"끝: 원복 sha 전부 일치, 판정 이상 {issues}건")
        return 1 if issues else 0
    finally:
        shutil.rmtree(base, ignore_errors=True)
        shutil.rmtree(cache, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
