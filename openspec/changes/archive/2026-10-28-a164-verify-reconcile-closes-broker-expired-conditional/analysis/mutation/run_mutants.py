"""a164 변이 원장 하네스(tasks 4.x) — 사본 디렉터리에서만 변이를 걸고, 무변이 대조군이 GREEN 일 때만 진행한다.

사용: python3 -I run_mutants.py <사본 루트> <결과 jsonl>
- 사본은 호출자가 만든다(rsync, .git 제외). 이 스크립트는 사본 밖을 쓰지 않는다.
- 변이마다 정확한 문자열 치환(개수 1 단언) → 대상 시험 집합 실행(-json) → 실패한 최상위 시험 이름 기록 → 원본 바이트로 되돌림
  → sha256 대조로 되돌림 확인. 한 번에 한 변이(사본 하나를 두 판이 같이 쓰지 않음).
"""
import hashlib, json, subprocess, sys, time
from pathlib import Path

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
VL = ("./internal/verifylive/", [], "Reconcile|Expiry|Expired|AnOlder|OldBinary|ADelete|AGeneric|Unreconciled|WithoutM0|LockedAppend|Progress|Report|TheRealRecords|Filled")
OF = ("./internal/official/", [], "Reconcile")
CMD = ("./cmd/tossctl/", ["-tags", "tossos_testseams"], "Reconcile|VerifyReconcile|M0Trigger|Expiry")
VLF = "internal/verifylive/"

RE = VLF + "reconcile_expired.go"
RC = VLF + "reconcile.go"
OD = "internal/official/reconcile_detail_reads.go"
MUTANTS = [
    ("4.1", "by-id 404 를 statement 로 취급", OD, [("return ReconcileConditionalDetail{Silent: true, RequestID: requestID}, nil", "return ReconcileConditionalDetail{Silent: false, RequestID: requestID}, nil")], [OF]),
    ("4.2", "A3 by-id 둘 다 statement 요구 삭제", RE, [("if byID[0].Silent || byID[1].Silent {", "if false && (byID[0].Silent || byID[1].Silent) {")], [VL]),
    ("4.3a", "A4 EXPIRED 비교를 COMPLETED 허용으로 넓힘", RE, [('case status == "COMPLETED" || triggered != "":', 'case triggered != "":'),
        ('case status != "EXPIRED" || hasSecond', 'case (status != "EXPIRED" && status != "COMPLETED") || hasSecond')], [VL]),
    ("4.3b", "A4 triggeredOrderId 검사 삭제", RE, [('case status == "COMPLETED" || triggered != "":', 'case status == "COMPLETED":')], [VL]),
    ("4.3c", "A4 second 검사 삭제", RE, [('case status != "EXPIRED" || hasSecond || !reconcileExpiryLegVocabulary[firstStatus]:', 'case status != "EXPIRED" || !reconcileExpiryLegVocabulary[firstStatus]:')], [VL]),
    ("4.4a", "by-id 둘째 읽기 삭제", RC, [("if byID[1], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {", "if false {")], [VL]),
    ("4.4b", "by-id 두 응답 비교 삭제", RC, [("if err := checkReconcileByIDPair(byID); err != nil {", "if err := checkReconcileByIDPair(byID); false && err != nil {")], [VL]),
    ("4.5a", "A5 OPEN 조건 판정 삭제(만료 모드)", RE, [("if err := checkReconcileNoLiveOrder(first); err != nil {", "if err := checkReconcileNoLiveOrder(reconcileRead{plainOpen: first.plainOpen}); err != nil {")], [VL]),
    ("4.5b", "A5 OPEN 일반 판정 삭제(만료 모드)", RE, [("if err := checkReconcileNoLiveOrder(first); err != nil {", "if err := checkReconcileNoLiveOrder(reconcileRead{condOpen: first.condOpen}); err != nil {")], [VL]),
    ("4.5c", "A5 발동 흔적 판정 삭제(만료 모드)", RE, [("if err := checkReconcileNoFiredTrace(first); err != nil {", "if err := checkReconcileNoFiredTrace(reconcileRead{}); err != nil {")], [VL]),
    ("4.6", "A7 체인 검사 삭제", RE, [('if chain := ChainOf(entries, target.Kind, target.ID); chain != "" {', 'if chain := ChainOf(entries, target.Kind, target.ID); false && chain != "" {')], [VL]),
    ("4.7a", "--expired 에서 Q1 재검을 다시 부름(창)", RC, [("if err := checkReconcileWindow(target, start, now(), p.Expired); err != nil {", "if err := checkReconcileWindow(target, start, now(), false); err != nil {")], [VL]),
    ("4.7b", "--expired 에서 정적 Q1 을 다시 부름", RC, [("\t\tif err := checkReconcileExpiryVocabulary(); err != nil {", "\t\tif err := checkReconcileRetention(target, now()); err != nil {")], [VL]),
    ("4.7c", "플래그 없음에서 Q1 재검 건너뜀(창 둘)", RC, [("checkReconcileWindow(target, start, now(), p.Expired)", "checkReconcileWindow(target, start, now(), true)"),
        ("checkReconcileWindow(target, start, at, p.Expired)", "checkReconcileWindow(target, start, at, true)")], [VL]),
    ("4.7d", "플래그 없음에서 정적 Q1 건너뜀", RC, [("\tif p.Expired {\n\t\tif err := checkReconcileExpiryVocabulary(); err != nil {", "\tif true {\n\t\tif err := checkReconcileExpiryVocabulary(); err != nil {")], [VL]),
    ("4.8", "terminal() 에서 ReconciledExpired 삭제", VLF + "record.go", [("return a.Cancelled || a.Filled || a.ReconciledAbsent || a.ReconciledExpired", "return a.Cancelled || a.Filled || a.ReconciledAbsent")], [VL, CMD]),
    ("4.9a", "by-id id 일치 검사 삭제", RE, [("if d.ID != target.ID || d.Symbol", "if d.Symbol")], [VL]),
    ("4.9b", "by-id 심볼 일치 검사 삭제", RE, [("|| d.Symbol != target.Symbol ", "")], [VL]),
    ("4.9c", "by-id 시장 일치 검사 삭제", RE, [("|| d.Market != market ", "")], [VL]),
    ("4.10a", "CLOSED 어휘를 영수증 없는 값으로 넓힘", RE, [('var reconcileExpiryClosedVocabulary = map[string]bool{"EXPIRED": true}', 'var reconcileExpiryClosedVocabulary = map[string]bool{"EXPIRED": true, "COMPLETED": true}')], [VL]),
    ("4.10b", "leg 어휘를 영수증 없는 값으로 넓힘", RE, [('var reconcileExpiryLegVocabulary = map[string]bool{"EXPIRED": true}', 'var reconcileExpiryLegVocabulary = map[string]bool{"EXPIRED": true, "ORDERED": true}')], [VL]),
    ("4.12", "silent + CLOSED 를 수락으로 되돌림", RE, [("\t\tif len(closed) > 0 {\n\t\t\treturn refuse(RefuseExpirySourcesDisagree", "\t\tif len(closed) > 0 {\n\t\t\treturn nil\n\t\t\treturn refuse(RefuseExpirySourcesDisagree")], [VL]),
    ("4.13a", "by-id first.status 어휘 검사 삭제", RE, [('checkReconcileExpiryStatement(fmt.Sprintf("by-id read %d", i+1), d.Status, d.FirstStatus,', 'checkReconcileExpiryStatement(fmt.Sprintf("by-id read %d", i+1), d.Status, "EXPIRED",')], [VL]),
    ("4.13b", "CLOSED 행 first.status 어휘 검사 삭제", RE, [('checkReconcileExpiryStatement("the CLOSED row", row.Status, row.FirstStatus,', 'checkReconcileExpiryStatement("the CLOSED row", row.Status, "EXPIRED",')], [VL]),
    ("4.13c", "type == SINGLE 검사 삭제", RE, [('|| d.Type != "SINGLE" ', '')], [VL]),
    ("4.14a", "silent 를 「404 면 전부」로(code 비교 삭제)", OD, [("body.Error == nil || body.Error.Code != reconcileDetailNotFoundCode", "body.Error == nil")], [OF]),
    ("4.14b", "본문 해독 불능 404 를 silent 로", OD, [("if json.Unmarshal([]byte(apiErr.Body), &body) != nil || body.Error == nil || body.Error.Code != reconcileDetailNotFoundCode {",
        "if json.Unmarshal([]byte(apiErr.Body), &body) == nil && (body.Error == nil || body.Error.Code != reconcileDetailNotFoundCode) {")], [OF]),
    ("4.15", "429·5xx·ErrAuth·ErrIPNotAllowed 를 silent 로 접음", OD, [("if !errors.As(err, &apiErr) || apiErr.Code != http.StatusNotFound {\n\t\treturn \"\", false\n\t}",
        "if !errors.As(err, &apiErr) {\n\t\treturn \"\", true\n\t}\n\tif apiErr.Code != http.StatusNotFound {\n\t\treturn \"\", false\n\t}")], [OF]),
    ("4.16a", "by-id 비교에서 status 삭제", RE, [("a.Status != b.Status || ", "")], [VL]),
    ("4.16b", "by-id 비교에서 first.status 삭제", RE, [("a.FirstStatus != b.FirstStatus || ", "")], [VL]),
    ("4.16c", "by-id 비교에서 triggeredOrderId 삭제", RE, [("a.TriggeredOrderID != b.TriggeredOrderID || ", "")], [VL]),
    ("4.16d", "by-id 비교에서 second 삭제", RE, [(" || a.HasSecond != b.HasSecond", "")], [VL]),
    ("4.17", "읽기 집합에서 일반 OPEN 읽기 생략(공용 readReconcileSet)", VLF + "reconcile_check.go", [("\tif out.plainOpen, err = readReconcilePlainGroup(ctx, r, symbol); err != nil {", "\tif false {")], [VL]),
    ("4.18", "만료 모드 A6 CLOSED 어휘 검사 삭제", RE, [('if row.Status == "" || reconcileOpenFamily[row.Status] || !reconcileExpiryClosedVocabulary[row.Status] {', 'if false && row.Status == "" {')], [VL]),
    ("4.19", "ReconciledArtifacts 필터에 넷째 종결 미반영", VLF + "projection_reconciled.go", [("if a := latest[key]; a.ReconciledAbsent || a.ReconciledExpired {", "if a := latest[key]; a.ReconciledAbsent {")], [VL]),
    ("4.20a", "by-id ① 을 승인 앞으로", RC, [("\tif err := approval.WriteText(out); err != nil {", "\tif p.Expired {\n\t\tif _, err := readReconcileByID(ctx, p.Reader, target.ID); err != nil {\n\t\t\treturn ReconcileResult{}, err\n\t\t}\n\t}\n\tif err := approval.WriteText(out); err != nil {")], [VL]),
    ("4.20b", "by-id 메서드의 ensureIdentity 생략", "cmd/tossctl/verify_reconcile_expired.go", [("\tif err := r.ensureIdentity(ctx); err != nil {", "\tif err := error(nil); err != nil {")], [CMD]),
    ("4.20c", "창 시작을 첫 목록 읽기 직전으로 되돌림", RC, [("\tstart := now()\n\tvar byID [2]official.ReconcileConditionalDetail\n\tif p.Expired {\n\t\tif byID[0], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {\n\t\t\treturn ReconcileResult{}, err\n\t\t}\n\t}\n",
        "\tvar byID [2]official.ReconcileConditionalDetail\n\tif p.Expired {\n\t\tif byID[0], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {\n\t\t\treturn ReconcileResult{}, err\n\t\t}\n\t}\n\tstart := now()\n")], [VL]),
    ("4.21", "줄에 by-id Call 을 실음", RE, [("\t\tAccountRef:    accountMask,\n\t\tObservations: []Observation{\n\t\t\t{Key: ObservationReconcileBasis, Value: basis,\n\t\t\t\tDetail: \"two by-id",
        "\t\tAccountRef:    accountMask,\n\t\tCalls:         []Call{{Endpoint: \"GET /api/v1/conditional-orders/{id}\", At: start}},\n\t\tObservations: []Observation{\n\t\t\t{Key: ObservationReconcileBasis, Value: basis,\n\t\t\t\tDetail: \"two by-id")], [VL, CMD]),
    ("4.22", "by-id 해독의 중복 키 거절 삭제(리뷰 Codex P1 수리)", OD, [("\t\tif _, dup := fields[key]; dup {", "\t\tif _, dup := fields[key]; false && dup {")], [OF]),
    ("4.23", "부재 모드 신선도 문구를 만료 문구로(D1 문구 불변)", VLF + "reconcile_check.go", [('\t\tsince := "the first list read"', '\t\tsince := "the first by-id read"')], [VL]),
]


def run(sets):
    failed, build = [], False
    for pkg, tags, regex in sets:
        p = subprocess.run(["go", "test", "-count=1", *tags, pkg, "-run", regex, "-json"], cwd=ROOT, capture_output=True, text=True)
        for line in p.stdout.splitlines():
            try:
                e = json.loads(line)
            except ValueError:
                continue
            # 표 시험은 하위 시험 이름까지 남김 — 어느 가드 행이 변이를 잡았는지가 원장의 요점.
            if e.get("Action") == "fail" and e.get("Test"):
                failed.append(e["Test"])
            if e.get("Action") == "output" and "[build failed]" in e.get("Output", ""):
                build = True
        if p.returncode != 0 and not failed:
            build = True
            failed.append("(no test failed — rc %d: %s)" % (p.returncode, (p.stderr or p.stdout)[-300:]))
    return sorted(set(failed)), build


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    only = set(sys.argv[3:])
    with OUT.open("a") as out:
        ctrl_failed, ctrl_build = run([VL, OF, CMD])
        out.write(json.dumps({"id": "4.11-control", "what": "무변이 대조군", "failed": ctrl_failed, "build_error": ctrl_build,
                              "verdict": "GREEN" if not ctrl_failed else "RED"}, ensure_ascii=False) + "\n")
        out.flush()
        if ctrl_failed:
            sys.exit("control is not GREEN — stop")
        for mid, what, rel, edits, sets in MUTANTS:
            if only and mid not in only:
                continue
            path = ROOT / rel
            original = path.read_bytes()
            before = sha(path)
            text = original.decode()
            for old, new in edits:
                n = text.count(old)
                if n != 1:
                    sys.exit(f"{mid}: expected one occurrence, found {n}: {old[:60]!r}")
                text = text.replace(old, new)
            path.write_bytes(text.encode())
            t0 = time.time()
            failed, build = run(sets)
            path.write_bytes(original)
            assert sha(path) == before, f"{mid}: revert failed"
            verdict = "BUILD-ERROR" if build and not [f for f in failed if not f.startswith("(")] else ("CAUGHT" if failed else "SURVIVED")
            out.write(json.dumps({"id": mid, "what": what, "file": rel, "verdict": verdict, "failed": failed,
                                  "seconds": round(time.time() - t0, 1)}, ensure_ascii=False) + "\n")
            out.flush()


main()
