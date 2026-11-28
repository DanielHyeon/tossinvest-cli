#!/usr/bin/env python3
"""a208 변이 원장 하네스(design 8.4).

사본(연결 worktree)에서만 변이한다 — 주 체크아웃 · 작업 worktree 는 건드리지 않음.
순서: 시작 sha 단언 → 무변이 대조군 GREEN 확인 → 변이 하나씩(적용 → 대상 시험 → 원복 → 원복 확인).
쓰는 법: python3 a208_mutation.py <사본 worktree 경로> <기대 HEAD sha>
"""
import os, subprocess, sys

COPY, WANT = sys.argv[1], sys.argv[2]
ENV = dict(os.environ, GOFLAGS="-trimpath")

ENG = "./internal/app/engine/"
CFG = "./internal/config/"
CON = "./internal/console/"

# (id, 파일, 바꿀 원문, 변이, 패키지, -run 정규식)
MUTANTS = [
    ("M1", "internal/app/engine/a208_adoption_floor_only.go",
     "\tif opts.Adoption.FloorOnlyExit {\n\t\treturn exitpolicy.CommonLadderRunner",
     "\tif false && opts.Adoption.FloorOnlyExit {\n\t\treturn exitpolicy.CommonLadderRunner",
     ENG, "TestA208FloorOnlyAdoptionRecordsTheRunnerCopy|TestA208FloorOnlyAdoptionReplacesRatchet"),
    ("M2", "internal/app/engine/a208_adoption_floor_only.go",
     "\tif opts.Adoption.FloorOnlyExit {\n\t\treturn exitpolicy.CommonLadderRunner",
     "\tif true || opts.Adoption.FloorOnlyExit {\n\t\treturn exitpolicy.CommonLadderRunner",
     ENG, "TestA208OffAdoptionIsUnchanged"),
    ("M3", "internal/app/engine/a208_adoption_floor_only.go",
     "\tif floorOnly {", "\tif true || floorOnly {", ENG, "TestA208OffAdoptionIsUnchanged"),
    ("M4", "internal/app/engine/a208_adoption_floor_only.go",
     "\tif floorOnly {", "\tif false && floorOnly {", ENG, "TestA208FloorOnlyAlertSaysNoPartialTakeProfit"),
    ("M5", "internal/config/engine.go",
     "\tnext.FloorOnlyExit = raw.FloorOnlyExit\n", "\n", CFG, "TestA208FloorOnlyExitRoundTrips"),
    ("M6", "internal/config/adoption_io.go",
     "\tblock.FloorOnlyExit = raw.FloorOnlyExit\n", "\n", CFG, "TestA208"),
    ("M7", "internal/config/adoption_io.go",
     ", FloorOnlyExit: a.FloorOnlyExit}", "}", CFG, "TestA208FloorOnlyExitRoundTrips"),
    ("M8", "internal/config/engine.go",
     '`json:"floor_only_exit,omitempty"`', '`json:"floor_only_exit"`', CFG, "TestA208FloorOnlyExitRoundTrips"),
    ("M9", "internal/config/engine.go",
     "\t\tcfg.Adoption = Adoption{Rejected: why}\n",
     "\t\tcfg.Adoption = Adoption{Rejected: why, FloorOnlyExit: next.FloorOnlyExit}\n",
     CFG, "TestA208ARefusedBlockClearsFloorOnly"),
    ("M10", "internal/console/settings.go",
     "\t\tFloorOnlyExit: current.FloorOnlyExit,\n", "", CON, "TestA208ConsoleSaveKeepsFloorOnly"),
    ("M11", "internal/app/engine/interlock.go",
     '\t\t{audit.ActionAdoptionToggle, "engine.adoption.floor_only_exit", strconv.FormatBool(adoption.FloorOnlyExit), ""},\n',
     "", ENG, "TestA208FloorOnlyIsAuditedAtStartup"),
    ("M12", "internal/app/engine/interlock.go",
     'strconv.FormatBool(adoption.FloorOnlyExit), ""}', 'strconv.FormatBool(false), ""}',
     ENG, "TestA208FloorOnlyIsAuditedAtStartup"),
]


def git(*args):
    return subprocess.run(["git", "-C", COPY, *args], capture_output=True, text=True, check=True).stdout.strip()


def test(pkg, pattern):
    p = subprocess.run(["go", "test", pkg, "-run", pattern, "-count=1"], cwd=COPY, env=ENV,
                       capture_output=True, text=True)
    return p.returncode, (p.stdout + p.stderr).strip().splitlines()[-3:]


assert git("rev-parse", "HEAD") == WANT, "사본 HEAD 가 기대 sha 가 아님"
assert git("status", "--porcelain") == "", "사본이 깨끗하지 않음"

# 무변이 대조군 — 모든 대상 시험이 사본에서 GREEN 이어야 변이 결과를 믿음.
for pkg in (ENG, CFG, CON):
    rc, tail = test(pkg, "TestA208")
    assert rc == 0, f"대조군 실패 {pkg}: {tail}"
print("control GREEN: engine · config · console TestA208")

rows = []
for mid, path, old, new, pkg, pattern in MUTANTS:
    full = os.path.join(COPY, path)
    src = open(full, encoding="utf-8").read()
    assert src.count(old) == 1, f"{mid}: 변이 자리가 정확히 1 곳이 아님"
    open(full, "w", encoding="utf-8").write(src.replace(old, new))
    try:
        rc, tail = test(pkg, pattern)
    finally:
        git("checkout", "--", path)
    assert open(full, encoding="utf-8").read() == src, f"{mid}: 원복 실패"
    assert git("status", "--porcelain") == "", f"{mid}: 사본이 원복 뒤 깨끗하지 않음"
    verdict = "CAUGHT" if rc != 0 else "SURVIVED"
    rows.append((mid, path, verdict, " / ".join(tail)))
    print(f"{mid} {verdict} {path} :: {tail[-1] if tail else ''}")

print("SUMMARY", sum(r[2] == "CAUGHT" for r in rows), "/", len(rows))
