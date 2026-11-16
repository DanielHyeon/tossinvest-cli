#!/usr/bin/env python3
"""a191 변이 하네스 — 워크트리 파일을 고치지 않고 `go test -overlay` 로 변이 사본을 끼움.

사용: python3 mutate.py <repo-root> <scratch-dir> [mutant-id ...]
- 사본은 scratch-dir/<pid>/ 아래에만 씀(판마다 pid — 두 판이 한 사본을 공유하지 않음).
- 무변이 대조군(m0)이 GREEN 이 아니면 멈춤.
- 시작·끝에 원본 파일 sha256 을 세어 같음을 단언(워크트리 무편집 증명).
"""
import hashlib, json, os, re, subprocess, sys
from pathlib import Path

ROOT = Path(sys.argv[1]).resolve()
SCRATCH = Path(sys.argv[2]).resolve() / str(os.getpid())
ONLY = set(sys.argv[3:])
CL = "internal/filldetect/closed_lookup.go"
DT = "internal/filldetect/detect.go"
RUN = "TestA191|TestJournalTrackedCarriesTheIntentID"

# (id, file, old, new, note) — old 는 그 파일에 정확히 한 번 있어야 함.
MUTANTS = [
 ("m1", DT, "\t\td.noteLeftOpen(cycle, key, order, len(snaps), snap)\n", "", "후보 등록 삭제(발동 안 함)"),
 ("m2", CL, "if snap.Derived.Terminal || snap.Derived.FailClosed {", "if snap.Derived.FailClosed {", "단건 종결도 후보"),
 ("m3", DT, "\t\tsnap.Derived = brokerstate.Derive(viewOf(snap, lineage))\n\t\tsnaps = append(snaps, snap)\n",
        "\t\tsnap.Derived = brokerstate.Derive(viewOf(snap, lineage))\n\t\tif complete { d.noteLeftOpen(cycle, key, TrackedOrder{OrderID: snap.OrderID}, len(snaps), snap) }\n\t\tsnaps = append(snaps, snap)\n", "OPEN 에 있는 주문도 후보"),
 ("m4", CL, "if !derived.Terminal || derived.FailClosed {", "if !derived.Terminal {", "행 파생의 !FailClosed 삭제"),
 ("m5", CL, "if !derived.Terminal || derived.FailClosed {", "if derived.FailClosed {", "행 파생의 Terminal 삭제"),
 ("m6", CL, "\t\t!closedQuantityEqual(rowSnap.FilledQuantity, single.FilledQuantity) {", "\t\tfalse {", "체결 수량 일치 삭제"),
 ("m7", CL, "if !closedQuantityEqual(rowSnap.Quantity, single.Quantity) ||", "if false ||", "주문 수량 일치 삭제"),
 ("m8", CL, "if actual, complete := snapshotOrderKey(rowSnap); !complete || actual != c.key {", "if _, complete := snapshotOrderKey(rowSnap); !complete {", "범위 일치 삭제"),
 ("m9", CL, "\t\tif !bytes.Equal(match, raw) {\n\t\t\treturn nil, UnresolvedClosedListAmbiguous\n\t\t}\n", "", "중복 중 첫 행 채택"),
 ("m10", DT, "\treturn d.settleLeftOpen(ctx, cycle, snaps), nil\n",
        "\tsnaps = d.settleLeftOpen(ctx, cycle, snaps)\n\tfor _, u := range cycle.Unresolved {\n\t\tif u.Reason == UnresolvedClosedListReadFailed {\n\t\t\treturn nil, errors.New(\"closed read failed\")\n\t\t}\n\t}\n\treturn snaps, nil\n", "조회 실패로 주기 실패"),
 ("m11", CL, "closedLookupInterval  = 30 * time.Second", "closedLookupInterval  = 0", "간격 제거"),
 ("m12", CL, "From:   day.AddDate(0, 0, -1).Format(closedTradingDay),", "From:   day.Format(closedTradingDay),", "창 from = 거래일"),
 ("m13", CL, "\t\tSymbol: c.key.Symbol,\n", "", "종목 필터 누락"),
 ("m14", CL, "unresolvedReportAfter = 30 * time.Second", "unresolvedReportAfter = 0", "즉시 보고"),
 ("m15", CL, "if *head.OrderID != orderID {", "if !strings.EqualFold(strings.TrimSpace(*head.OrderID), orderID) {", "id 비교 완화"),
 ("m16", CL, "\taccepted := single\n", "\taccepted := single\n\taccepted.FilledQuantity = rowSnap.FilledQuantity\n", "체결 수량을 행에서(m6 과 함께)"),
 ("m17", DT, "\t\tcycle.TrackedReads++\n", "\t\tcycle.TrackedReads++\n\t\td.noteLeftOpen(cycle, key, order, 0, Snapshot{})\n\t\t_ = d.settleLeftOpen(ctx, cycle, []Snapshot{{}})\n", "조회를 기존 범위 가드 앞으로"),
 ("m18", CL, "\tfor key := range d.leftOpen {\n\t\tif !current[key] {\n\t\t\tdelete(d.leftOpen, key)\n\t\t}\n\t}\n", "", "정리 삭제"),
 ("m19", CL, "\taccepted := single\n", "\taccepted := rowSnap\n", "행의 실행 필드 사용"),
 ("m20", CL, "\t\tif !bytes.Equal(match, raw) {", "\t\tif !a191SameScope(match, raw) {", "중복을 범위로만 판정"),
 ("m21", CL, "if snap.Derived.Terminal || snap.Derived.FailClosed {", "if snap.Derived.Terminal {", "UNKNOWN 단건도 후보(M2 위반)"),
 ("m22", CL, "func closedRowCarriesFill(raw json.RawMessage) bool {\n", "func closedRowCarriesFill(raw json.RawMessage) bool {\n\tif true {\n\t\treturn true\n\t}\n", "execution 존재 요구 삭제(M6)"),
 ("m23", CL, "\t\tif i == chosen {\n", "\t\t_, _ = chosen, i\n\t\tif last := track.lastLookup; last.IsZero() || now.Sub(last) >= closedLookupInterval {\n", "주기당 1건 상한 삭제"),
 ("m24", CL, "lookupCtx, cancel := context.WithTimeout(ctx, closedLookupDeadline)", "lookupCtx, cancel := context.WithCancel(ctx)", "조회 시한 삭제"),
 ("m25", CL, "\t\t\tobs.FieldIntentID:    c.order.IntentID,\n", "\t\t\tobs.FieldIntentID:    c.order.IntentID,\n\t\t\tobs.FieldOrderID:     c.order.OrderID,\n", "경보에 브로커 주문번호"),
 ("m26", CL, "\t\treason := UnresolvedClosedLookupDeferred\n", "\t\treason := UnresolvedClosedLookupDeferred\n\t\tif i != chosen {\n\t\t\ttrack.lastLookup = now\n\t\t}\n", "미뤄진 후보의 lastLookup 갱신"),
 ("m30", CL, "\tif len(d.leftOpenCycle) > 0 && d.leftOpenCycle[0].cycle != cycle {\n\t\td.leftOpenCycle = d.leftOpenCycle[:0]\n\t}\n", "", "실패 주기 후보 버림 삭제(리뷰 1 #1)"),
 ("m31", CL, "\t\t\t\tsnaps[c.index] = accepted\n\t\t\t\tcontinue\n", "\t\t\t\tsnaps[c.index] = accepted\n\t\t\t\tdelete(d.leftOpen, c.key)\n\t\t\t\tcontinue\n", "수용 즉시 간격 상태 삭제(리뷰 1 #2)"),
 ("m32", CL, "\t\tif !closedRowCarriesFill(row) {\n\t\t\treturn Snapshot{}, UnresolvedClosedListExecutionMissing\n\t\t}\n\t\treturn Snapshot{}, UnresolvedClosedListRowUnreadable\n", "\t\treturn Snapshot{}, UnresolvedClosedListRowUnreadable\n", "파싱 실패의 M6 사유 가름 삭제(리뷰 1 #3)"),
]
EXTRA = {"m9": "\nvar _ = bytes.Equal\n", "m20": "\nvar _ = bytes.Equal\nfunc a191SameScope(a, b json.RawMessage) bool {\n\tsa, _ := parseSnapshot(a, time.Time{}, time.Time{})\n\tsb, _ := parseSnapshot(b, time.Time{}, time.Time{})\n\tka, _ := snapshotOrderKey(sa)\n\tkb, _ := snapshotOrderKey(sb)\n\treturn ka == kb\n}\n"}
PAIR = {"m16": ("m6",)}  # m16 은 m6 의 일치 삭제와 함께 걸어야 의미가 있음

def sha(p): return hashlib.sha256((ROOT / p).read_bytes()).hexdigest()

def run(mid, files):
    SCRATCH.mkdir(parents=True, exist_ok=True)
    overlay = {"Replace": {}}
    for rel, text in files.items():
        dst = SCRATCH / mid / rel
        dst.parent.mkdir(parents=True, exist_ok=True)
        dst.write_text(text)
        overlay["Replace"][str(ROOT / rel)] = str(dst)
    ov = SCRATCH / mid / "overlay.json"
    ov.write_text(json.dumps(overlay))
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(["go", "test", "-overlay", str(ov), "./internal/filldetect/", "-count=1",
                        "-timeout", "45s", "-run", RUN, "-v"], cwd=ROOT, capture_output=True, text=True, env=env)
    out = p.stdout + p.stderr
    failed = sorted(set(re.findall(r"^--- FAIL: (\S+)", out, re.M)))
    if "panic: test timed out" in out:
        failed.append("TIMEOUT(" + ",".join(re.findall(r"^\s+(TestA191\S+) \(", out, re.M)[:3]) + ")")
    build_err = p.returncode != 0 and not failed
    return p.returncode, failed, out if build_err else ""

def main():
    before = {f: sha(f) for f in (CL, DT)}
    originals = {f: (ROOT / f).read_text() for f in (CL, DT)}
    rc, failed, err = run("m0", originals)
    print(f"m0 control rc={rc} failed={failed}")
    if rc != 0:
        print(err[:3000]); sys.exit("control is not GREEN — stop")
    for mid, rel, old, new, note in MUTANTS:
        if ONLY and mid not in ONLY:
            continue
        files = dict(originals)
        steps = [(rel, old, new)]
        for extra in PAIR.get(mid, ()):
            m = next(x for x in MUTANTS if x[0] == extra)
            steps.insert(0, (m[1], m[2], m[3]))
        for r, o, n in steps:
            assert files[r].count(o) == 1, (mid, r, o[:60])
            files[r] = files[r].replace(o, n)
        if mid in EXTRA:
            files[rel] = files[rel] + EXTRA[mid]
        rc, failed, err = run(mid, files)
        verdict = "CAUGHT" if failed else ("BUILD-FAIL" if rc != 0 else "SURVIVED")
        print(f"{mid}\t{verdict}\t{note}\t{' '.join(failed)}")
        if verdict == "BUILD-FAIL":
            print(err[:1500])
    after = {f: sha(f) for f in (CL, DT)}
    assert before == after, "worktree source changed during the run"
    print("worktree sha unchanged:", {k: v[:12] for k, v in after.items()})

main()
