#!/usr/bin/env python3
"""a168 3.1 변이 원장 하네스 — 사본 worktree 에서 변이를 한 번에 하나씩 넣고 선택 시험을 돌림.

사용(저장소 밖 사본 worktree 가 이미 있어야 함 — 실제 작업 worktree 에서는 돌지 않음):
    git -C <repo> worktree add --detach ~/tossos-work/a168-mut <GREEN sha>
    python3 -I run_mutations.py ~/tossos-work/a168-mut <GREEN sha> > mutation-ledger.txt

규칙: 시작 sha 단언 · 깨끗한 트리 단언 · 무변이 대조군 GREEN 확인 뒤 시작 · 변이마다 원문이 정확히 한 번 있는지 단언 ·
변이 뒤 `git checkout -- <file>` 로 되돌리고 깨끗한 트리를 다시 단언 · 한 번에 한 판(병렬 없음).
판정: 시험 rc != 0 이면 CAUGHT(실패한 시험 이름을 적음), rc == 0 이면 SURVIVED.
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

RESP = "internal/official/reconcile_response.go"
READS = "internal/official/reconcile_reads.go"
DETAIL = "internal/official/reconcile_detail_reads.go"

TEST = ["go", "test", "-count=1", "-run", "TestReconcile", "./internal/official/", "./internal/verifylive/"]

ACCEPT_COND = ("\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileConditionalPage{}, err",
               "\tif err := reconcileAccepted(attempts); err != nil && false {\n\t\treturn ReconcileConditionalPage{}, err")
ACCEPT_ORDER = ("\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileOrderPage{}, err",
                "\tif err := reconcileAccepted(attempts); err != nil && false {\n\t\treturn ReconcileOrderPage{}, err")
ACCEPT_INST = ("\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn \"\", err",
               "\tif err := reconcileAccepted(attempts); err != nil && false {\n\t\treturn \"\", err")
ACCEPT_DETAIL = ("\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileConditionalDetail{}, err",
                 "\tif err := reconcileAccepted(attempts); err != nil && false {\n\t\treturn ReconcileConditionalDetail{}, err")
DECODE_THEN_ACCEPT_ORDER = (
    "\tpage, err := DecodeReconcileOrderPage(raw)\n\tif err != nil {\n\t\treturn ReconcileOrderPage{}, err\n\t}\n"
    "\t// a168 — 기존 해독 뒤 판정(design D3)\n"
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileOrderPage{}, err\n\t}\n",
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileOrderPage{}, err\n\t}\n"
    "\tpage, err := DecodeReconcileOrderPage(raw)\n\tif err != nil {\n\t\treturn ReconcileOrderPage{}, err\n\t}\n",
)
DECODE_THEN_ACCEPT_DETAIL = (
    "\td, err := DecodeReconcileConditionalDetail(raw)\n\tif err != nil {\n\t\treturn ReconcileConditionalDetail{}, err\n\t}\n",
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileConditionalDetail{}, err\n\t}\n"
    "\td, err := DecodeReconcileConditionalDetail(raw)\n\tif err != nil {\n\t\treturn ReconcileConditionalDetail{}, err\n\t}\n",
)
DECODE_THEN_ACCEPT_INST = (
    "\tif len(raw) != 1 {\n",
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn \"\", err\n\t}\n\tif len(raw) != 1 {\n",
)
DECODE_THEN_ACCEPT = (
    "\tpage, err := DecodeReconcileConditionalPage(raw)\n\tif err != nil {\n\t\treturn ReconcileConditionalPage{}, err\n\t}\n"
    "\t// a168 — 기존 해독 뒤 판정(design D3): 200 · 모호하지 않은 본문일 때만 페이지를 돌려줌\n"
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileConditionalPage{}, err\n\t}\n",
    "\tif err := reconcileAccepted(attempts); err != nil {\n\t\treturn ReconcileConditionalPage{}, err\n\t}\n"
    "\tpage, err := DecodeReconcileConditionalPage(raw)\n\tif err != nil {\n\t\treturn ReconcileConditionalPage{}, err\n\t}\n",
)

MUTANTS = [
    ("M1", "200 판정을 2xx 판정으로", RESP, "last.StatusCode != http.StatusOK", "last.StatusCode/100 != http.StatusOK/100"),
    ("M2", "마지막 대신 첫 시도", RESP, "last := attempts[len(attempts)-1]", "last := attempts[0]"),
    ("M3", "D2 짝 비교 EqualFold → ==", RESP, "if strings.EqualFold(key, other) {", "if key == other {"),
    ("M3b", "D2 짝 비교 → ToLower", RESP, "if strings.EqualFold(key, other) {", "if strings.ToLower(key) == strings.ToLower(other) {"),
    ("M3d", "D6 비교 → ToLower", RESP, "key != contract && strings.EqualFold(key, contract)",
     "key != contract && strings.ToLower(key) == strings.ToLower(contract)"),
    ("M4", "배열 재귀 생략", RESP, "case len(trimmed) > 0 && trimmed[0] == '[':", "case false && len(trimmed) > 0 && trimmed[0] == '[':"),
    ("M5", "객체 재귀 생략", RESP, "if err := reconcileUnambiguous(fields[key], name+\".\"+strconv.Quote(key)); err != nil {",
     "if err := reconcileUnambiguous(nil, name+\".\"+strconv.Quote(key)); err != nil {"),
    ("M6a", "조건주문 목록 판정 호출 무력화", READS, *ACCEPT_COND),
    ("M6b", "일반 주문 목록 판정 호출 무력화", READS, *ACCEPT_ORDER),
    ("M6c", "종목 판정 호출 무력화", READS, *ACCEPT_INST),
    ("M6d", "by-id 판정 호출 무력화", DETAIL, *ACCEPT_DETAIL),
    ("M7", "조건주문 목록 판정을 해독 앞으로", READS, *DECODE_THEN_ACCEPT),
    ("M7b", "일반 주문 목록 판정을 해독 앞으로", READS, *DECODE_THEN_ACCEPT_ORDER),
    ("M7c", "by-id 판정을 해독 앞에 하나 더(앞 판정이 먼저 거절)", DETAIL, *DECODE_THEN_ACCEPT_DETAIL),
    ("M7d", "종목 판정을 항목 수 앞에 하나 더", READS, *DECODE_THEN_ACCEPT_INST),
    ("M12", "정확한 중복 판정(reconcileObject)을 map 해독으로", RESP,
     "\t\tfields, err := reconcileObject(trimmed, name)\n",
     "\t\tvar fields map[string]json.RawMessage\n\t\terr := json.Unmarshal(trimmed, &fields)\n"),
    ("M13", "경로 키 인용 삭제", RESP, "name+\".\"+strconv.Quote(key)", "name+\".\"+key"),
    ("M8", "관측 0 방어를 통과로", RESP, "\t\treturn reconcileSchema(\"no attempt was observed\")", "\t\treturn nil"),
    ("M9", "계약 키 아래로만 재귀", RESP,
     "\t\tfor _, key := range keys {\n\t\t\t// 경로의 키는",
     "\t\tfor _, key := range keys {\n\t\t\tknown := false\n\t\t\tfor _, c := range reconcileContractKeys {\n"
     "\t\t\t\tknown = known || c == key\n\t\t\t}\n\t\t\tif !known {\n\t\t\t\tcontinue\n\t\t\t}\n"
     "\t\t\t// 경로의 키는"),
    ("M10", "계약 키 목록에서 triggeredOrderId 삭제", RESP, ", \"triggeredOrderId\",\n", ",\n"),
    ("M11a", "D6 술어를 바이트 같음으로(규칙 꺼짐)", RESP, "key != contract && strings.EqualFold(key, contract)",
     "key != contract && key == contract"),
    ("M11b", "D6 의 key != contract 삭제", RESP, "key != contract && strings.EqualFold(key, contract)",
     "strings.EqualFold(key, contract)"),
]


def git(wt: Path, *args: str) -> str:
    return subprocess.run(["git", "-C", str(wt), *args], check=True, capture_output=True, text=True).stdout.strip()


def assert_clean(wt: Path) -> None:
    status = git(wt, "status", "--porcelain")
    if status:
        sys.exit(f"tree not clean: {status}")


def run_tests(wt: Path) -> tuple[int, list[str], bool]:
    proc = subprocess.run(TEST, cwd=wt, capture_output=True, text=True)
    out = proc.stdout + proc.stderr
    fails = sorted({ln.strip().split()[2] for ln in out.splitlines() if ln.strip().startswith("--- FAIL:")})
    # 컴파일 실패는 시험이 잡은 것이 아님 — 무효 변이로 따로 셈(1회차에서 M1 이 import 미사용으로 거짓 CAUGHT 였음)
    build_failed = "[build failed]" in out or "[setup failed]" in out
    return proc.returncode, fails, build_failed


def main() -> int:
    wt, start = Path(sys.argv[1]).resolve(), sys.argv[2]
    if git(wt, "rev-parse", "--show-toplevel") != str(wt):
        sys.exit("not a worktree root")
    if not git(wt, "rev-parse", "HEAD").startswith(start):
        sys.exit(f"HEAD is not {start}")
    if wt == Path("/home/daniel/tossos-work/a168-impl").resolve():
        sys.exit("refusing to mutate the working worktree")
    assert_clean(wt)
    rc, fails, _ = run_tests(wt)
    print(f"CONTROL (무변이) rc={rc} fails={fails}")
    if rc != 0:
        sys.exit("control is not GREEN — stop")
    caught = survived = 0
    for mid, desc, rel, old, new in MUTANTS:
        path = wt / rel
        src = path.read_text()
        if src.count(old) != 1:
            print(f"{mid} NOT-APPLIED ({src.count(old)} matches) — {desc}")
            continue
        path.write_text(src.replace(old, new))
        rc, fails, build_failed = run_tests(wt)
        git(wt, "checkout", "--", rel)
        assert_clean(wt)
        if build_failed:
            print(f"{mid} INVALID(build failed) — {desc}")
            continue
        verdict = "CAUGHT" if rc != 0 and fails else "SURVIVED"
        caught += verdict == "CAUGHT"
        survived += verdict == "SURVIVED"
        print(f"{mid} {verdict} rc={rc} — {desc} — fails={fails}")
    print(f"합계 CAUGHT {caught} · SURVIVED {survived} · 변이 {len(MUTANTS)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
