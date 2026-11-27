#!/usr/bin/env python3
"""a207 변이 원장 하네스 — 별도 worktree(사본)에서만 변이를 걸고 시험을 돌린 뒤 되돌림.

사용:
  python3 -I a207_mutation.py <사본 worktree> <기대 HEAD sha>

지키는 것:
  - 사본 worktree 가 깨끗하고 HEAD 가 기대 sha 일 때만 시작함(원본 worktree 를 건드리지 않음).
  - 무변이 대조군이 GREEN 이 아니면 멈춤(계측기가 눈먼 경우를 가름).
  - 변이마다 원문 치환이 정확히 1 회 일어났는지 단언하고, 끝나면 `git checkout -- <파일>` 로 되돌린 뒤 사본이 다시 깨끗한지 확인함.
결과: CAUGHT(시험 실패) / SURVIVED(시험 통과) / NA(구조상 불가 — 표 밖에서 사유 기록).
"""
import subprocess
import sys

JOURNAL = ["go", "test", "./internal/journal/", "-count=1", "-run",
           "A207|TestExitSnapshotDuplicateDecisionIsNotRearmed|TestARefusalReArmsTheLevel"]
ENGINE = ["go", "test", "./internal/app/engine/", "-count=1", "-run", "A207|TestARefusedProposalReleasesTheLevelAndAlerts"]

PRED = "internal/journal/a207_compact_evidence.go"
APPEND = "internal/journal/exit_state.go"
HYDRATE = "internal/journal/exit_snapshot.go"
NOTICE = "internal/app/engine/a207_unblock_notice.go"
LOOP = "internal/app/engine/exitloop.go"
CYCLE = "internal/app/engine/exit_order_log.go"

MUTANTS = [
    ("M1", "해제 훅 확인 제거", PRED, "h.action IN (?, ?)", "(h.action IN (?, ?) OR 1)", JOURNAL),
    ("M2", "직전 무장 action 비교 제거", PRED, "prev.action = ?", "(prev.action = ? OR 1)", JOURNAL),
    ("M3", "직전 무장 level_after 비교 제거", PRED, "prev.level = ?", "(prev.level = ? OR 1)", JOURNAL),
    ("M4", "직전 무장 lifecycle_generation 비교 제거", PRED, "prev.generation = ?", "(prev.generation = ? OR 1)", JOURNAL),
    ("M5", "압축 행에서 decision_id 도 NULL", APPEND,
     "effectiveSource = EffectiveSourceRecomputedRepeat",
     "effectiveSource = EffectiveSourceRecomputedRepeat\n\t\t\t\tdecisionID = nil", JOURNAL),
    ("M6", "압축 대상 조건에서 action 확인 제거(판정 전용 행에도 적용)", APPEND,
     'row.Evaluation.EffectiveSource == EffectiveSourceRecomputed && row.Action != "" && ',
     'row.Evaluation.EffectiveSource == EffectiveSourceRecomputed && ', JOURNAL),
    ("M6b", "압축 대상 조건에서 출처 recomputed 확인 제거(saved_monotone 에도 적용)", APPEND,
     'row.Evaluation.EffectiveSource == EffectiveSourceRecomputed && row.Action != "" && ',
     'row.Action != "" && ', JOURNAL),
    ("M7", "리더가 표지 없는 JSON 없는 행도 압축으로 읽음", HYDRATE,
     "if evidence.EffectiveSource.String == EffectiveSourceRecomputedRepeat {",
     "if evidence.EffectiveSource.String == EffectiveSourceRecomputedRepeat || !evidence.RecomputedJSON.Valid {", JOURNAL),
    ("M7b", "압축 모양 검사에서 JSON NULL 요구 제거", PRED,
     "!evidence.SavedJSON.Valid && !evidence.RecomputedJSON.Valid && !evidence.EffectiveJSON.Valid &&",
     "!evidence.SavedJSON.Valid && !evidence.EffectiveJSON.Valid &&", JOURNAL),
    ("M8", "해소 알림을 셈과 무관하게 냄", NOTICE,
     "\tif streak == nil || streak.count == 0 {\n\t\treturn\n\t}",
     "\tif streak == nil {\n\t\tstreak = &refusalStreak{}\n\t}", ENGINE),
    ("M9", "해소 알림을 주기 끝이 아니라 확정 갈래에서 바로 냄", NOTICE,
     "o.unblockNotices = append(o.unblockNotices, obs.Event{", "o.alert(context.Background(), obs.Event{", ENGINE),
    ("M11", "술어 조회 오류를 판정 오류로 돌려줌", APPEND,
     "lookupErr == nil && repeat {", "lookupErr != nil || repeat {\n\t\t\t\tif lookupErr != nil {\n\t\t\t\t\treturn lookupErr\n\t\t\t\t}", JOURNAL),
    ("M12", "원장이 해제를 거절해도 셈을 더함", LOOP, "\tif released {", "\tif released || true {", ENGINE),
    ("M13", "주기 끝의 해소 알림 넘기기 제거", CYCLE, "\to.flushUnblockNotices(ctx)\n", "", ENGINE),
    ("M14", "확정 갈래의 해소 대기열 넣기 제거", LOOP,
     "\t\to.noteUnblocked(m, proposal, intentID) // a207 D6: Place 뒤 — 대기열에 넣기만 함\n", "", ENGINE),
]


def sh(cmd, cwd):
    return subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    tree, want = sys.argv[1], sys.argv[2]
    head = sh(["git", "rev-parse", "HEAD"], tree).stdout.strip()
    if head != want:
        sys.exit(f"refuse: copy HEAD {head} is not {want}")
    if sh(["git", "status", "--porcelain"], tree).stdout.strip():
        sys.exit("refuse: copy worktree is not clean")
    for name, cmd in (("journal", JOURNAL), ("engine", ENGINE)):
        r = sh(cmd, tree)
        print(f"control {name}: rc={r.returncode}")
        if r.returncode != 0:
            print(r.stdout[-2000:], r.stderr[-2000:])
            sys.exit("refuse: the unmutated control is not GREEN")
    print("| # | 변이 | 파일 | 결과 | 첫 실패 |")
    print("|---|---|---|---|---|")
    for mid, what, path, old, new, cmd in MUTANTS:
        full = f"{tree}/{path}"
        text = open(full).read()
        if text.count(old) != 1:
            sys.exit(f"{mid}: expected exactly one site in {path}, found {text.count(old)}")
        open(full, "w").write(text.replace(old, new))
        try:
            r = sh(cmd, tree)
            out = r.stdout + r.stderr
            first = next((line.strip() for line in out.splitlines() if line.startswith("--- FAIL")), "")
            if not first and r.returncode != 0:
                first = next((line.strip() for line in out.splitlines() if "error" in line or "FAIL" in line), "build/test failure")
            verdict = "CAUGHT" if r.returncode != 0 else "SURVIVED"
            print(f"| {mid} | {what} | `{path}` | {verdict} | {first[:110]} |")
        finally:
            sh(["git", "checkout", "--", path], tree)
        if sh(["git", "status", "--porcelain"], tree).stdout.strip():
            sys.exit(f"{mid}: copy not clean after restore")
    print("final: copy clean, HEAD", sh(["git", "rev-parse", "HEAD"], tree).stdout.strip())


if __name__ == "__main__":
    main()
