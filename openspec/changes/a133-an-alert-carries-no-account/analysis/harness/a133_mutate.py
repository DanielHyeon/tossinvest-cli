#!/usr/bin/env python3
"""a133 변이 하네스 — 사본에서만 변이하고, 무변이 대조군이 GREEN 이어야 판을 돎(a131 하네스와 같은 틀).

사용: python3 -I a133_mutate.py [--keep]
- 저장소 루트는 이 파일 위치에서 유도함. 사본 = `git archive HEAD` + 워킹트리의 a133 표면 파일 덮어쓰기, 사본 경로에 pid.
- 각 변이: 원문 조각이 사본 파일에 **정확히 한 번** 있어야 함(닿음 단언) → 바꿈 → 대상 시험 → 시험 실패면 CAUGHT → 원복.
- 컴파일 실패는 BUILD-FAILED 로 따로 셈(시험이 잡은 것이 아님).
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
PACKAGES = ["./internal/obs/", "./internal/app/engine/", "./cmd/tossctl/"]
RUN = "TestA133|TestNtfy"
# a133 표면 — 병행 a130 의 파일(journal · 콘솔 · 도구)은 사본에 넣지 않음.
SURFACE = (
    "internal/obs/ntfy.go",
    "internal/obs/notifier.go",
    "internal/obs/a133_",
    "internal/app/engine/a133_",
    "cmd/tossctl/a133_",
)


def changed_files() -> list[str]:
    out = subprocess.run(["git", "-C", str(ROOT), "status", "--porcelain", "--untracked-files=all"],
                         capture_output=True, text=True, check=True).stdout
    return [line[3:] for line in out.splitlines() if line[3:].endswith(".go") and line[3:].startswith(SURFACE)]


MUTANTS = [
    ("M1 the boundary is bypassed (Publish does not redact)", "internal/obs/ntfy.go",
     "\tmsg = redactForTransport(msg)\n", ""),
    ("M2 one send path is missed (redaction moved from the boundary into notificationFor)", "internal/obs/ntfy.go",
     "\tmsg = redactForTransport(msg)\n", ""),
    ("M3 account-key fields are rendered (field drop removed)", "internal/obs/notifier.go",
     "\t\tif redact.IsAccountKey(k) {\n\t\t\tcontinue\n\t\t}\n", ""),
    ("M4 the title is sent unredacted", "internal/obs/a133_transport_redact.go",
     "\tmsg.Title = redact.String(msg.Title)\n", ""),
    ("M5 tags and the type tag are sent unredacted", "internal/obs/a133_transport_redact.go",
     "\tmsg.Type = EventType(redact.String(string(msg.Type)))\n", ""),
    ("M6 the tags list is sent unredacted", "internal/obs/a133_transport_redact.go",
     "\t\t\ttags[i] = redact.String(tag)\n", "\t\t\ttags[i] = tag\n"),
    ("M7 the body is sent unredacted", "internal/obs/a133_transport_redact.go",
     "\tmsg.Body = redact.String(msg.Body)\n", ""),
    ("M8 a second production transport without the boundary", "internal/obs/ntfy.go",
     "// priorityFor maps a severity",
     "type a133Bypass struct{ inner *Ntfy }\n\n"
     "func (b a133Bypass) Publish(ctx context.Context, msg Notification) error { return b.inner.Publish(ctx, msg) }\n\n"
     "// priorityFor maps a severity"),
    ("M9 the field drop is narrowed to the `account` key only", "internal/obs/notifier.go",
     "\t\tif redact.IsAccountKey(k) {\n", "\t\tif redact.IsAccountKey(k) && k == FieldAccount {\n"),
    ("M10 a second obs file reaches the network", "internal/obs/heartbeat.go",
     "import (\n", "import (\n\t_ \"net/http\"\n"),
    ("M11 Notification gains a text field the boundary does not cover", "internal/obs/ntfy.go",
     "\t// Priority is ntfy's 1-5 scale.", "\t// Click is a new header field (mutation).\n\tClick string\n\t// Priority is ntfy's 1-5 scale."),
]

# M2 는 경계의 치환을 지우고 notificationFor 에 같은 치환을 넣음 — 실시간 경로(P1~P3)는 덮이고 재전송(P4 · P5)만 샘.
EXTRA = {
    "M2": ("internal/obs/notifier.go",
           "\treturn Notification{Type: e.Type, Severity: severity, Title: title, Body: body}\n",
           "\treturn redactForTransport(Notification{Type: e.Type, Severity: severity, Title: title, Body: body})\n"),
}


def run(cmd: list[str], cwd: Path) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, env=env)
    return p.returncode, p.stdout + p.stderr


def tests(copy: Path) -> tuple[str, str]:
    rc, out = run(["go", "test", "-count=1", "-run", RUN, *PACKAGES], copy)
    failed = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "panic:" in l]
    if "[build failed]" in out or "[setup failed]" in out:
        return "BUILD", "; ".join([l for l in out.splitlines() if ".go:" in l][:4])
    return ("GREEN" if rc == 0 else "RED"), "; ".join(failed[:6])


def apply(repo: Path, rel: str, old: str, new: str) -> tuple[Path, str] | None:
    path = repo / rel
    text = path.read_text()
    if text.count(old) != 1:
        return None
    path.write_text(text.replace(old, new, 1))
    return path, text


def main() -> int:
    keep = "--keep" in sys.argv
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    files = changed_files()
    copy = Path(tempfile.mkdtemp(prefix=f"a133-mutate-{os.getpid()}-"))
    try:
        archive = copy / "head.tar"
        subprocess.run(["git", "-C", str(ROOT), "archive", "--format=tar", "-o", str(archive), "HEAD"], check=True)
        with tarfile.open(archive) as tf:
            tf.extractall(copy / "repo", filter="data")
        archive.unlink()
        repo = copy / "repo"
        for rel in files:
            src = ROOT / rel
            if src.exists():
                (repo / rel).parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src, repo / rel)
        print(f"# a133 mutation — HEAD {head} + {len(files)} working-tree a133 files · copy {repo}")
        state, why = tests(repo)
        print(f"CONTROL {state} {why}")
        if state != "GREEN":
            print("# control is not GREEN — no mutant is judged")
            return 2
        caught = 0
        for name, rel, old, new in MUTANTS:
            undo: list[tuple[Path, str]] = []
            first = apply(repo, rel, old, new)
            if first is None:
                print(f"UNREACHED {name} — fragment not exactly once in {rel}")
                continue
            undo.append(first)
            extra = EXTRA.get(name.split()[0])
            if extra:
                second = apply(repo, *extra)
                if second is None:
                    for p, t in undo:
                        p.write_text(t)
                    print(f"UNREACHED {name} — extra fragment not exactly once in {extra[0]}")
                    continue
                undo.append(second)
            try:
                state, why = tests(repo)
            finally:
                for p, t in reversed(undo):
                    p.write_text(t)
            verdict = {"GREEN": "SURVIVED", "RED": "CAUGHT", "BUILD": "BUILD-FAILED"}[state]
            caught += 1 if state == "RED" else 0
            print(f"{verdict} {name} — {why}", flush=True)
        print(f"# {caught}/{len(MUTANTS)} caught")
        return 0 if caught == len(MUTANTS) else 1
    finally:
        if not keep:
            shutil.rmtree(copy, ignore_errors=True)


if __name__ == "__main__":
    raise SystemExit(main())
