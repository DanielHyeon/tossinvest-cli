#!/usr/bin/env python3
"""a200 2.5 — 변이 원장 하네스(design §3-6 + Manager 결정 3 의 추가 변이).

    python3 mutate.py <사본 worktree 절대경로> <기대 시작 sha>

- 사본은 이 change 브랜치의 커밋을 detached 로 뽑은 별도 worktree 여야 한다(작업 worktree 를 변이하지 않는다).
- 시작: 사본 HEAD == 기대 sha 단언, `git status --porcelain` 이 빈 것 단언.
- 무변이 대조군: 같은 시험이 GREEN 이 아니면 멈춘다.
- 변이마다: 원문이 정확히 1 회 있는지 단언 → 적용 → 시험 → `git checkout -- 파일` 로 되돌림 → 파일 sha256 이 원본과 같은지 단언.
- 시험: `go test ./internal/journal -run TestA200` 와 `go test ./internal/reconcile -run TestA200`(둘 다 -count=1).
  실패(빌드 실패 포함)면 CAUGHT. 빌드 실패는 따로 적는다(가짜 CAUGHT 를 가린다).
"""
import hashlib
import pathlib
import subprocess
import sys

F = "internal/journal/a200_adjustment_close_fill_recognition.go"
P = "internal/journal/position_projection.go"

MUTANTS = [
    # (가)는 도달 가능한 상태에서 (나)에 함의됨(CLOSED 에서 거절되는 매도는 SELL_ON_CLOSED 뿐) — 이 변이는 순수 판정 시험만 잡는다.
    ("M1", "(가) 거절 유형 검사 삭제(순수 판정)", F,
     "\tif !recognisableRefusal(f.Refusal) {\n\t\treturn false\n\t}\n", ""),
    ("M17", "원장 읽기 앞 거절 유형 조기 반환 삭제(impl-r1 P2-1)", F,
     "if !recognisableRefusal(outcome.Refusal) || !instance.Held", "if !instance.Held"),
    ("M18", "(라) 「주문 수량 > 0」 삭제(impl-r1 P2-2)", F,
     "if err != nil || positive <= 0 {", "if err != nil || (positive <= 0 && false) {"),
    # gstack r1 testing: (나)의 「마지막 행」 은 조정 행 정렬에 기댐 — 정렬을 뒤집으면 두 단계 수렴 원장 시험이 잡아야 함.
    ("M19", "조정 행 정렬 뒤집기(created_at, rowid → DESC)", F,
     "WHERE position_id = ? ORDER BY created_at, rowid`", "WHERE position_id = ? ORDER BY created_at DESC, rowid DESC`"),
    ("M2", "(나) 「0 조정 정확히 하나」 삭제", F,
     "if zeroRows != 1 || !lastZero ||", "if !lastZero ||"),
    ("M3", "(나) 「마지막 조정 행」 삭제", F,
     "if zeroRows != 1 || !lastZero ||", "if zeroRows != 1 || (!lastZero && false) ||"),
    ("M4", "(나) kind 검사 삭제", F,
     " || closing.Kind != convergenceAdjustmentKind {", " {"),
    ("M5", "(다) 발의 intent 비교 삭제", F,
     " || f.PendingIntentID != f.OriginIntentID {", " {"),
    ("M6", "(라) PLACE · AMEND 유일 검사 삭제", F,
     "if len(f.OrderAttempts) != 1 || f.OrderAttempts[0].Kind", "if len(f.OrderAttempts) == 0 || f.OrderAttempts[0].Kind"),
    ("M7", "(라) PLACE 종류 검사 삭제(AMEND 하나도 통과)", F,
     " f.OrderAttempts[0].Kind != string(KindPlace) ||\n", "\n"),
    ("M8", "(라) 접수 확정 상태 검사 삭제", F,
     " ||\n\t\tf.OrderAttempts[0].State != string(StateConfirmed) {", " {"),
    ("M9", "(라) 상한 비교 삭제", F,
     "\treturn cmp <= 0\n}", "\treturn cmp <= 0 || true\n}"),
    ("M10", "(라) 상한 비교를 < 로", F,
     "\treturn cmp <= 0\n}", "\treturn cmp < 0\n}"),
    ("M11", "(라) 감소분 대신 new_quantity", F,
     "riskcalc.SubDecimal(closing.PrevQuantity, closing.NewQuantity)", "riskcalc.SubDecimal(closing.NewQuantity, \"0\")"),
    ("M12", "(라) 감소분의 prev 열을 다른 열(prev_avg_price)로", F,
     "SELECT kind, prev_quantity, new_quantity FROM position_adjustments",
     "SELECT kind, prev_avg_price, new_quantity FROM position_adjustments"),
    ("M13", "인정 기록(표지) 쓰기 생략", F,
     "\tif err := appendExitEventFromHook(ctx, tx, exitEventRow{",
     "\tif err := func(context.Context, *ApplyTx, exitEventRow) error { return nil }(ctx, tx, exitEventRow{"),
    ("M14", "인정 뒤 수량 다시 빼기(투영 쓰기)", P,
     "err != nil || recognised {\n\t\t\treturn err\n\t\t}",
     "err != nil || recognised {\n\t\t\tif err == nil {\n\t\t\t\t_, err = tx.Exec(ctx, `UPDATE positions SET quantity = '-' || ? WHERE id = ?`, fill.Delta, instance.ID)\n\t\t\t}\n\t\t\treturn err\n\t\t}"),
    ("M15", "인정 뒤 차단도 쓰기", P,
     "err != nil || recognised {\n\t\t\treturn err\n\t\t}",
     "err != nil {\n\t\t\treturn err\n\t\t} else if recognised {\n\t\t\t_ = recognised\n\t\t}"),
    ("M16", "판정 호출 삭제(오늘 동작으로 되돌림)", P,
     "if recognised, err := recogniseAdjustmentCloseFill(ctx, tx, origin, instance, outcome); err != nil || recognised {",
     "if recognised, err := false, error(nil); err != nil || recognised {"),
]

TESTS = [
    ["go", "test", "./internal/journal", "-run", "TestA200", "-count=1"],
    ["go", "test", "./internal/reconcile", "-run", "TestA200", "-count=1"],
]


def git(copy: pathlib.Path, *args: str) -> str:
    return subprocess.check_output(["git", "-C", str(copy), *args], text=True).strip()


def run_tests(copy: pathlib.Path) -> tuple[bool, str]:
    out = []
    ok = True
    for cmd in TESTS:
        proc = subprocess.run(cmd, cwd=copy, capture_output=True, text=True)
        tail = "\n".join((proc.stdout + proc.stderr).strip().splitlines()[-4:])
        out.append(f"$ {' '.join(cmd)}\n{tail}")
        ok = ok and proc.returncode == 0
    return ok, "\n".join(out)


def sha(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> int:
    copy, expected = pathlib.Path(sys.argv[1]).resolve(), sys.argv[2]
    head = git(copy, "rev-parse", "HEAD")
    assert head.startswith(expected) or expected.startswith(head[:len(expected)]), (head, expected)
    assert git(copy, "status", "--porcelain") == "", "사본이 깨끗하지 않음"
    print(f"# 변이 원장 — 사본 {copy} HEAD {head}\n")
    ok, log = run_tests(copy)
    print(f"## 무변이 대조군: {'GREEN' if ok else 'RED — 멈춤'}\n```\n{log}\n```\n")
    if not ok:
        return 1
    print("| id | 변이 | 파일 | 결과 | 시험 꼬리 |\n|---|---|---|---|---|")
    survived = 0
    for mid, what, rel, old, new in MUTANTS:
        path = copy / rel
        original = path.read_text(encoding="utf-8")
        before = sha(path)
        assert original.count(old) == 1, f"{mid}: 원문 {original.count(old)} 회"
        path.write_text(original.replace(old, new), encoding="utf-8")
        ok, log = run_tests(copy)
        build_fail = "build failed" in log or "[setup failed]" in log
        result = "SURVIVED" if ok else ("CAUGHT(빌드 실패)" if build_fail else "CAUGHT")
        survived += ok
        git(copy, "checkout", "--", rel)
        assert sha(path) == before, f"{mid}: 되돌림 뒤 sha 다름"
        short = " / ".join(line for line in log.splitlines() if line.startswith(("--- FAIL", "FAIL", "ok")))[:300]
        print(f"| {mid} | {what} | {rel.split('/')[-1]} | {result} | {short} |")
    assert git(copy, "status", "--porcelain") == "", "끝난 뒤 사본이 깨끗하지 않음"
    print(f"\n살아남은 변이 {survived} / {len(MUTANTS)}. 끝난 뒤 사본 git status 깨끗함.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
