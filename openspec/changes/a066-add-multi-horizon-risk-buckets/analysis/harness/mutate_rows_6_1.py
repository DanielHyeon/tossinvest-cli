#!/usr/bin/env python3
"""a066 6.1 (B)(1) — Branch Test Map 의 "covered 인데 RED 기록 없음" 행마다 그 분기를 끄는 변이를 사본에서 돌림.

사용: python3 mutate_rows_6_1.py <scratch-dir> [--only REGEX]

- 행 목록: analysis/function-logic/*/branch-test-map.md 에서 RED 칸이 "no RED recorded" 이고 GREEN 칸이 "NOT covered"
  가 아닌 행(2026-08-04 파도 행 — 그 표기는 역사적 사실로 두고, 이 원장이 오늘의 RED→GREEN 을 따로 적음).
- 위치: 행의 "<kind> at L:C" 를 그 번들 FLM 의 Source 파일에서 찾고, 그 줄에 kind 키워드가 없으면 NOT-APPLIED(표류).
- 변이(컴파일되는 모양만):
    if    → 조건 끝에 `&& false` (init 문 변수는 그대로 쓰임)          case → `case (X) && false:`
    for   → `for cond && false {`                                        range → 본문 첫 줄에 `if true { continue }`
    else  → `} else if false {` (`} else if err := F(…` 는 F 를 no-op 함수로)
- 사본: <scratch-dir>/mutrows-<pid>/ (git 추적 파일 = 작업 트리 내용). 원장 첫 줄 TREE, 둘째 줄 무변이 대조군 — RED 면 멈춤.
- 판정: 초점 시험(아래 FOCUS)에서 실패하면 CAUGHT. 초점에서 살아남으면 전체 패키지로 다시 돌려 CAUGHT(full)/SURVIVED.
- 한 번에 한 판. 변이마다 원본으로 되돌리고 되돌림을 바이트 비교로 확인함.
"""
import glob
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
CHANGE = ROOT / "openspec/changes/a066-add-multi-horizon-risk-buckets"
FOCUS = {
    "internal/journal": [["go", "test", "-count=1", "-run",
                          "RiskBucket|QFinal|A066|FirstLeg|Fill|Owner|Strategy|Reconcile|Migration|Schema", "./internal/journal"]],
    "internal/riskbucket": [["go", "test", "-count=1", "./internal/riskbucket"],
                            ["go", "test", "-count=1", "-tags", "tossos_testseams", "./internal/riskbucket"]],
}
FULL = {
    "internal/journal": [["go", "test", "-count=1", "./internal/journal"]],
    "internal/riskbucket": [["go", "test", "-count=1", "-run", "RiskBucket|QFinal|A066|Fill|Owner", "./internal/journal"]],
}


def rows():
    out = []
    for bmap in sorted(glob.glob(str(CHANGE / "analysis/function-logic/*/branch-test-map.md"))):
        flm = Path(bmap).with_name("function-logic-map.md").read_text(encoding="utf-8")
        src = re.search(r"Source: `([^`]+)`", flm).group(1)
        for line in Path(bmap).read_text(encoding="utf-8").splitlines():
            if not re.match(r"\|\s*B\d+", line):
                continue
            cells = [c.strip() for c in line.strip().strip("|").split("|")]
            if len(cells) < 5 or "no RED recorded" not in cells[3] or "NOT covered" in cells[4]:
                continue
            m = re.match(r"(if|case|for|range|else) at (\d+):(\d+)", cells[1])
            out.append((Path(bmap).parent.name, cells[0], src, m.group(1) if m else "?", int(m.group(2)) if m else 0))
    return out


def mutate(text: str, kind: str, lineno: int):
    lines = text.split("\n")
    i = lineno - 1
    line = lines[i]
    stripped = line.lstrip()
    if kind == "if":
        if not re.search(r"\bif\b", line) or not line.rstrip().endswith("{"):
            return None
        head = line.rstrip()[:-1].rstrip()
        k = head.index("if ") + 3
        prefix, cond = head[:k], head[k:]
        if "; " in cond:  # init 문은 그대로 두고 조건만 끔
            init, cond = cond.rsplit("; ", 1)
            prefix += init + "; "
        lines[i] = f"{prefix}({cond}) && false {{"
    elif kind == "case":
        if not stripped.startswith("case ") or not line.rstrip().endswith(":"):
            return None
        indent = line[: len(line) - len(stripped)]
        lines[i] = f"{indent}case ({stripped[5:].rstrip()[:-1]}) && false:"
    elif kind == "for":
        if not stripped.startswith("for ") or " range " in stripped or not line.rstrip().endswith("{"):
            return None
        indent = line[: len(line) - len(stripped)]
        lines[i] = f"{indent}for ({stripped[4:].rstrip()[:-1].strip()}) && false {{"
    elif kind == "range":
        if not stripped.startswith("for ") or " range " not in stripped or not line.rstrip().endswith("{"):
            return None
        lines[i] = line + " if true { continue };"
    elif kind == "else":
        # `} else if err := F(…); err != nil {` 가 여러 줄에 걸치면 조건을 못 끄므로 F 를 아무것도 안 하는 함수로 바꿈.
        m = re.search(r"\} else if err := ([A-Za-z_][A-Za-z0-9_.]*)\(", line)
        if m:
            lines[i] = line.replace(m.group(1) + "(", "func(...any) error { return nil }(", 1)
        elif "} else {" in line:
            lines[i] = line.replace("} else {", "} else if false {", 1)
        else:
            return None
    else:
        return None
    return "\n".join(lines)


def run(copy: Path, commands, env):
    failed, notes = [], []
    for command in commands:
        r = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if r.returncode != 0:
            out = r.stdout + r.stderr
            names = [l.strip()[len("--- FAIL: "):].split(" ")[0] for l in out.splitlines() if l.strip().startswith("--- FAIL: ")]
            failed += [n for n in names if not any(o != n and o.startswith(n + "/") for o in names)]
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    if failed or notes:
        return False, f"{len(failed)} failing: " + ", ".join(failed[:6]) + (" …" if len(failed) > 6 else "") + (" | " + " | ".join(notes) if notes else "")
    return True, ""


def main():
    args = sys.argv[1:]
    only = None
    if "--only" in args:
        i = args.index("--only")
        only = re.compile(args[i + 1])
        args = args[:i] + args[i + 2:]
    scratch = Path(args[0])
    copy = scratch / f"mutrows-{os.getpid()}"
    copy.mkdir(parents=True)
    tracked = subprocess.run(["git", "ls-files", "--", "go.mod", "go.sum", "internal", "cmd", "tools"], cwd=ROOT,
                             capture_output=True, text=True, check=True).stdout.split()
    for rel in tracked:
        if (ROOT / rel).exists():
            (copy / rel).parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(ROOT / rel, copy / rel)
    env = dict(os.environ, GOFLAGS="-trimpath")
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "diff", "--name-only", "HEAD", "--", "internal", "cmd", "tools", "go.mod", "go.sum"],
                           cwd=ROOT, capture_output=True, text=True, check=True).stdout.split()
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ledger.write(f"TREE\tHEAD {head}\tuncommitted tracked: {','.join(dirty) or 'none'}\n")
    ok, why = run(copy, FOCUS["internal/journal"] + FOCUS["internal/riskbucket"], env)
    ledger.write(f"CONTROL\t{'GREEN' if ok else 'RED'}\t{why}\n")
    ledger.flush()
    if not ok:
        print("control RED — stop:", why)
        sys.exit(2)
    for bundle, branch, src, kind, lineno in rows():
        ident = f"{bundle} {branch} {kind}@{lineno}"
        if only and not only.search(ident):
            continue
        target = copy / src
        pristine = target.read_text(encoding="utf-8")
        mutated = mutate(pristine, kind, lineno)
        if mutated is None:
            ledger.write(f"{ident}\tNOT-APPLIED\tline {lineno} of {src} does not carry a {kind} header: {pristine.split(chr(10))[lineno - 1].strip()[:100]}\n")
            ledger.flush()
            continue
        pkg = "internal/riskbucket" if src.startswith("internal/riskbucket") else "internal/journal"
        target.write_text(mutated, encoding="utf-8")
        green, why = run(copy, FOCUS[pkg], env)
        verdict = "CAUGHT"
        if green:
            green, why = run(copy, FULL[pkg], env)
            verdict = "SURVIVED" if green else "CAUGHT(full)"
        ledger.write(f"{ident}\t{verdict}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
        if target.read_text(encoding="utf-8") != pristine:
            ledger.write(f"{ident}\tREVERT-FAILED\n")
            ledger.close()
            sys.exit(3)
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
