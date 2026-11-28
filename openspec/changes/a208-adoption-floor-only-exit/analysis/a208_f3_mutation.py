#!/usr/bin/env python3
"""a208 F3 변이 원장 하네스(design 8.4 M13~M23).

사본(연결 worktree)에서만 변이한다 — 주 체크아웃 · 작업 worktree 는 건드리지 않음.
순서: 시작 sha 단언 → 무변이 대조군 GREEN 확인 → 변이 하나씩(적용 → 대상 시험 → 원복 → 원복 확인). 한 번에 한 판.
쓰는 법: python3 a208_f3_mutation.py <사본 worktree 경로> <기대 HEAD sha>
변이 하나는 (파일, 원문, 바꿀 글) 쌍 여럿 — 의미 층을 재려고 digest 고정까지 같이 푸는 변이(…b)가 있음.
"""
import os, subprocess, sys

COPY, WANT = sys.argv[1], sys.argv[2]
ENV = dict(os.environ, GOFLAGS="-trimpath")

ENG, CFG, EXP, JRN = "./internal/app/engine/", "./internal/config/", "./internal/exitpolicy/", "./internal/journal/"
FLOOR = "internal/exitpolicy/a208_adopted_floor_only.go"
ENGF = "internal/app/engine/a208_adoption_floor_only.go"
LOOP = '\tfor i := range ladder.Rungs {\n\t\tladder.Rungs[i].PartialRatio = "0"\n\t}\n'
PIN = "\tladder.PolicyID, ladder.PolicyVersion, ladder.PolicyDigest = identity.ID, identity.Version, identity.Digest\n"
NOPIN = "\tladder.PolicyID, ladder.PolicyVersion = identity.ID, identity.Version\n"

# (id, 설명, [(파일, 원문, 변이)], [(패키지, -run)])
MUTANTS = [
    ("M13", "사본 부분익절 원복(digest 고정 유지)", [(FLOOR, LOOP, "")],
     [(EXP, "TestA208FloorOnlyCopyKeepsTheHybridFloors"), (ENG, "TestA208FloorOnlyAdoptedPositionNeverTakesProfit")]),
    ("M13b", "사본 부분익절 원복 + digest 주장 제거(의미 층)", [(FLOOR, LOOP, ""), (FLOOR, PIN, NOPIN)],
     [(EXP, "TestA208FloorOnlyCopyKeepsTheHybridFloors"), (ENG, "TestA208FloorOnlyAdoptedPositionNeverTakesProfit")]),
    ("M14", "OFF 갈래가 사본 ID 를 적음", [(ENGF, "\tif !opts.Adoption.FloorOnlyExit {\n", "\tif false && !opts.Adoption.FloorOnlyExit {\n")],
     [(ENG, "TestA208OffAdoptionIsUnchanged")]),
    ("M15", "기존 편입 갈래를 HYBRID_50 까지 넓힘(표)",
     [("internal/exitpolicy/common_policy.go", "\tif adopted && policy.ID == CommonLadderRunner {\n",
       "\tif adopted && policy.ID == CommonLadderHybrid50 {\n\t\tfor i := range policy.Ladder.Rungs {\n"
       "\t\t\tpolicy.Ladder.Rungs[i].PartialRatio = \"0\"\n\t\t}\n\t}\n\tif adopted && policy.ID == CommonLadderRunner {\n")],
     [(EXP, "TestA208AdoptedHybridRowsKeepTheirMeaning")]),
    ("M15c", "기존 편입 갈래를 HYBRID_50 까지 넓힘(정체성)",
     [("internal/exitpolicy/identity.go", "\tcase CommonLadderHybrid50:\n\t\tdigest = hybrid50PolicyDigest\n",
       "\tcase CommonLadderHybrid50:\n\t\tdigest = hybrid50PolicyDigest\n\t\tif adopted {\n\t\t\treturn adoptedFloorOnlyIdentity(adopted)\n\t\t}\n")],
     [(EXP, "TestA208AdoptedHybridRowsKeepTheirMeaning")]),
    ("M16", "사본 편입 검사 삭제", [(FLOOR, "\tif !adopted {\n", "\tif false && !adopted {\n")],
     [(EXP, "TestA208FloorOnlyCopyIsAdoptionOnly"), (JRN, "TestA208FloorOnlyCopyNeedsAnAdoptedPosition")]),
    ("M17", "원장 record 가 사본 ID 거부",
     [("internal/journal/adoption.go", " && !exitpolicy.IsAdoptedFloorOnlyPolicyID(a.ExitPolicyID)", "")],
     [(JRN, "TestA208JournalRefusesOtherUnregisteredPolicies"), (ENG, "TestA208FloorOnlyAdoptionRecordsTheHybridCopy")]),
    ("M17b", "원장 OpenExitState 가 사본 ID 거부",
     [("internal/journal/exit_state.go", ' &&\n\t\t\t!exitpolicy.IsAdoptedFloorOnlyPolicyID(policyID) {', " {")],
     [(JRN, "TestA208JournalRefusesOtherUnregisteredPolicies"), (ENG, "TestA208FloorOnlyAdoptionRecordsTheHybridCopy")]),
    ("M18", "BALANCED 를 RUNNER 사본으로 대응(개정 1 동작)",
     [(FLOOR, "\tcase CommonLadderRunner:\n\t\treturn CommonLadderRunner, true\n",
       "\tcase CommonLadderRunner, CommonLadderBalanced:\n\t\treturn CommonLadderRunner, true\n")],
     [(ENG, "TestA208FloorOnlyIsIgnoredWithoutAFloorCopy"), (CFG, "TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy")]),
    ("M19", "사본이 러너 트레일을 버림(digest 주장 제거)",
     [(FLOOR, PIN, NOPIN + '\tladder.RunnerTrailPct = ""\n')],
     [(EXP, "TestA208FloorOnlyCopyKeepsTheHybridFloors"), (ENG, "TestA208FloorOnlyAdoptedPositionNeverTakesProfit")]),
    ("M20", "audit 값이 무시돼도 FormatBool", [(ENGF, '\tif adoption.FloorOnlyIgnored != "" {\n', '\tif false {\n')],
     [(ENG, "TestA208IgnoredFloorOnlyIsAuditedAtStartup")]),
    ("M21", "로더 무시 사유가 늘 빈 값",
     [("internal/config/a208_floor_only_ignored.go", "AdoptedFloorOnlyPolicyID(commonPolicy); ok {", "AdoptedFloorOnlyPolicyID(commonPolicy); ok || true {")],
     [(CFG, "TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy")]),
    ("M22", "mergeEngine 파생 줄 누락",
     [("internal/config/engine.go", "\tcfg.Adoption.FloorOnlyIgnored = cfg.Adoption.floorOnlyIgnoredReason(cfg.ExitPolicy.CommonPolicy)\n", "")],
     [(CFG, "TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy")]),
    ("M23", "알림이 토글 값만 보고 무시 여부를 안 봄",
     [(ENGF, "\tif _, ok := adoptionFloorOnlyPolicy(opts); ok {\n", "\tif opts.Adoption.FloorOnlyExit {\n")],
     [(ENG, "TestA208FloorOnlyIsIgnoredWithoutAFloorCopy")]),
]


def git(*args):
    return subprocess.run(["git", "-C", COPY, *args], capture_output=True, text=True, check=True).stdout.strip()


def test(pkg, pattern):
    p = subprocess.run(["go", "test", pkg, "-run", pattern, "-count=1"], cwd=COPY, env=ENV,
                       capture_output=True, text=True)
    return p.returncode, (p.stdout + p.stderr).strip().splitlines()[-3:]


assert git("rev-parse", "HEAD") == WANT, "사본 HEAD 가 기대 sha 가 아님"
assert git("status", "--porcelain") == "", "사본이 깨끗하지 않음"

# 무변이 대조군 — 대상 시험이 사본에서 GREEN 이어야 변이 결과를 믿음.
for pkg in (ENG, CFG, EXP, JRN):
    rc, tail = test(pkg, "TestA208")
    assert rc == 0, f"대조군 실패 {pkg}: {tail}"
print("control GREEN: engine · config · exitpolicy · journal TestA208")

caught = 0
for mid, what, edits, checks in MUTANTS:
    originals = {}
    for path, old, new in edits:
        full = os.path.join(COPY, path)
        src = originals.setdefault(path, open(full, encoding="utf-8").read())
        cur = open(full, encoding="utf-8").read()
        assert cur.count(old) == 1, f"{mid}: 변이 자리가 정확히 1 곳이 아님 ({path})"
        open(full, "w", encoding="utf-8").write(cur.replace(old, new))
    results = []
    try:
        for pkg, pattern in checks:
            rc, tail = test(pkg, pattern)
            results.append((pkg, pattern, rc, tail[-1] if tail else ""))
    finally:
        for path in originals:
            git("checkout", "--", path)
    for path, src in originals.items():
        assert open(os.path.join(COPY, path), encoding="utf-8").read() == src, f"{mid}: 원복 실패"
    assert git("status", "--porcelain") == "", f"{mid}: 사본이 원복 뒤 깨끗하지 않음"
    hit = [r for r in results if r[2] != 0]
    verdict = "CAUGHT" if hit else "SURVIVED"
    caught += bool(hit)
    print(f"{mid} {verdict} — {what}")
    for pkg, pattern, rc, last in results:
        print(f"    {'FAIL' if rc else 'pass'} {pkg} {pattern} :: {last}")

print("SUMMARY", caught, "/", len(MUTANTS))
