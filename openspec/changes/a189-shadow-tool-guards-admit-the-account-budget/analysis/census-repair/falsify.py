#!/usr/bin/env python3
"""a189 census 갱신 반증 — 갱신된 corpus 가 예산을 거치지 않는 HTTP 호출을 여전히 잡는지 사본에서 잼.

HEAD 를 git archive 로 pid 붙은 임시 디렉터리에 풀고(작업 트리 무편집), 변이마다 대상 문자열 1회 단언 → 적용 →
그 패키지 컴파일 확인(go vet 대신 go build) → census 시험 → 기대 문구 단언. 무변이 대조군이 GREEN 이 아니면 멈춤.
사용: python3 -I falsify.py
"""
import os, shutil, subprocess, sys, tempfile
from pathlib import Path

MAIN = "tools/a128-jev-shadow-probe/main.go"
IMPORT = 'import (\n'
SRC = "\tvar src marketSource = client\n"
MUTANTS = [
    ("inside-pinned-function", [(IMPORT, IMPORT + '\t"net/http"\n'), (SRC, SRC + '\t_, _ = http.Get("https://example.invalid/")\n')],
     "TestA166WiringCensusBaseline", "reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions"),
    ("new-unpinned-function", [(IMPORT, IMPORT + '\t"net/http"\n'), ("\nfunc finishTool(", "\nfunc a189Unbudgeted() { _, _ = http.Get(\"https://example.invalid/\") }\n\nfunc finishTool(")],
     "TestA166WiringCensusBaseline", "callsite inventory count changed"),
]


class Fail(Exception):
    pass


def check(cond, msg):
    if not cond:
        raise Fail(msg)


def run(base, args, cache):
    env = {k: v for k, v in os.environ.items() if k != "GOFLAGS"}
    env["GOCACHE"] = str(cache)
    p = subprocess.run(["go", *args], cwd=base, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    return p.returncode, p.stdout


def main():
    root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
    head = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    dirty = subprocess.check_output(["git", "-C", str(root), "status", "--porcelain", "--", "internal", "tools", "cmd"], text=True)
    check(dirty == "", f"Go 소스·corpus 가 커밋되지 않음 — 커밋 뒤 돌릴 것:\n{dirty}")
    base = Path(tempfile.mkdtemp(prefix=f"a189-census-{os.getpid()}-"))
    cache = Path(tempfile.mkdtemp(prefix=f"a189-census-gocache-{os.getpid()}-"))
    try:
        arc = subprocess.run(["git", "-C", str(root), "archive", "HEAD", "go.mod", "go.sum", "cmd", "internal", "tools"], check=True, stdout=subprocess.PIPE).stdout
        subprocess.run(["tar", "-x", "-C", str(base)], input=arc, check=True)
        print(f"기준 HEAD {head}, 사본 {base.name}")
        test = ["test", "-count=1", "-v", "-run", "^TestA166WiringCensus", "./internal/accountbudgetruntime/"]
        rc, out = run(base, test, cache)
        if rc != 0:
            print(out)
            print("control: FAIL — 중단")
            return 1
        print(f"control: GREEN ({out.count('--- PASS')} PASS 줄)")
        target = base / MAIN
        original = target.read_text()
        issues = 0
        for mid, edits, expect_test, expect_msg in MUTANTS:
            text = original
            for old, new in edits:
                check(text.count(old) == 1, f"{mid}: TARGET-COUNT {text.count(old)} {old!r}")
                text = text.replace(old, new, 1)
                check(new in text, f"{mid}: 변이가 안 들어감")
            target.write_text(text)
            brc, bout = run(base, ["build", "-o", os.devnull, "./tools/a128-jev-shadow-probe/"], cache)
            rc, out = (1, "") if brc else run(base, test, cache)
            target.write_text(original)
            check(target.read_text() == original, "원복 실패")
            msgs = [l.strip() for l in out.splitlines() if "a166_wiring_census_test.go:" in l][:3]
            if brc:
                print(f"{mid}: INVALID(빌드 실패)\n    " + "\n    ".join(bout.splitlines()[:4])); issues += 1
            elif rc == 0:
                print(f"{mid}: SURVIVED"); issues += 1
            elif f"--- FAIL: {expect_test} (" in out and any(expect_msg in m for m in msgs):
                print(f"{mid}: CAUGHT — {expect_test}\n    " + "\n    ".join(msgs))
            else:
                print(f"{mid}: WRONG-CATCH\n    " + "\n    ".join(msgs)); issues += 1
        print(f"끝: 판정 이상 {issues}건")
        return 1 if issues else 0
    finally:
        shutil.rmtree(base, ignore_errors=True)
        shutil.rmtree(cache, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
