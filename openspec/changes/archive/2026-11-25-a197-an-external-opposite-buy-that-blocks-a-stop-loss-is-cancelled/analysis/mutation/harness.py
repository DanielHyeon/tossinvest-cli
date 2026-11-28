#!/usr/bin/env python3
"""a197 변이 원장 하네스(tasks 4.1) — 사본 대상으로만, 한 번에 한 판.

사용: python3 harness.py --copy <빈 디렉터리 또는 없는 경로> --sha <시작 커밋> [--gocache <dir>] [--only id,...]
  - 저장소 루트는 이 파일 위치에서 유도함(절대경로를 박지 않음): analysis/mutation/harness.py → parents[5].
  - 사본은 `git archive <sha>` 를 풀어 만듦 — 실제 작업 트리 · .git 을 건드리지 않음. 시작 sha 를 단언함.
  - 무변이 대조군: 변이마다 쓰는 시험 묶음이 사본에서 전부 통과해야 시작함.
  - 닿음: 변이의 옛 문자열이 그 파일에 정확히 1 회 있어야 함(아니면 NOT_REACHED). 변이 뒤 `go vet` 이 실패하면 COMPILE(잡음 아님 — CAUGHT 로 세지 않음).
  - 복원: 변이마다 원래 바이트로 되돌리고 sha256 을 다시 대조함.
  - -trimpath 를 쓰지 않음: 엔진 구조 핀(a087 · a111)이 runtime.Caller 의 파일 경로를 열기 때문(-trimpath 면 경로가 모듈 경로가 되어 핀이 거짓 실패함).
"""
import argparse, hashlib, json, os, pathlib, subprocess, sys, tarfile, io

HERE = pathlib.Path(__file__).resolve()
ROOT = HERE.parents[5]
OUT = HERE.parent

E = "internal/app/engine/a197_external_opposite_buy.go"
L = "internal/app/engine/exit_order_log.go"
X = "internal/app/engine/exitloop.go"
J = "internal/journal/a197_external_cancel_issued.go"
N = "internal/app/engine/engine.go"
ENG = "./internal/app/engine"

MUTANTS = [
    ("M01", "토글 검사 삭제", [(E, "\tif !o.opts.ExternalBuyCancel {\n\t\treturn\n\t}", "\tif false && !o.opts.ExternalBuyCancel {\n\t\treturn\n\t}")], [(ENG, "A197TheToggleOff")]),
    ("M02", "보호 손절 조건 삭제", [(E, "return isProtective(p) && out.State", "return out.State")], [(ENG, "A197TheStageOpensOnly|A197ATakeProfit")]),
    ("M03", "사유 조건 삭제", [(E, "&& out.State == journal.StateFailedConfirmed && out.Reason == execgw.ReasonOppositePendingOrder", "&& out.State == journal.StateFailedConfirmed")], [(ENG, "A197TheStageOpensOnly")]),
    ("M04", "매수 조건 삭제(매도도 취소)", [(E, "if !row.buy || engineIDs[row.orderID] {", "if engineIDs[row.orderID] {")], [(ENG, "A197OneExternalBuyPerCycle")]),
    ("M05", "귀속 제외 삭제", [(E, "if !row.buy || engineIDs[row.orderID] {", "if !row.buy || (engineIDs[row.orderID] && false) {")], [(ENG, "A197AnEngineAttributed")]),
    ("M06", "미지 행 가드 삭제", [(E, "\t\tif !ok {\n\t\t\treturn nil, false\n\t\t}\n\t\trows = append", "\t\tif !ok {\n\t\t\tcontinue\n\t\t}\n\t\trows = append")], [(ENG, "A197AnUnreadableRow")]),
    ("M07", "「읽지 못함」을 「비어 있음」으로", [(E, "\t\tif err != nil {\n\t\t\treturn nil, false\n\t\t}\n\t\traws = append", "\t\tif err != nil {\n\t\t\treturn nil, true\n\t\t}\n\t\traws = append")], [(ENG, "A197AnIncompleteRead")]),
    ("M08", "커서 없는 다음 쪽 허용", [(E, "\t\tif !p.HasNext {", "\t\tif !p.HasNext || p.NextCursor == \"\" {")], [(ENG, "A197AnIncompleteRead")]),
    ("M09", "조회 시한 삭제", [(E, "listCtx, cancel := context.WithTimeout(ctx, externalBuyListTimeout)", "listCtx, cancel := context.WithCancel(ctx)")], [(ENG, "A197AnIncompleteRead")]),
    ("M10a", "「이미 냈다」 판독 삭제", [(E, "\t\tif issued {\n\t\t\tcontinue\n\t\t}", "\t\tif issued && false {\n\t\t\tcontinue\n\t\t}")], [(ENG, "A197ANumberIsCancelledOnce")]),
    ("M10b", "「이미 냈다」 상태를 CONFIRMED 로 좁힘", [(J, "WHERE a.kind = ? AND a.target_order_id = ? AND TRIM(i.source) = ?", "WHERE a.kind = ? AND a.target_order_id = ? AND TRIM(i.source) = ? AND a.state = 'CONFIRMED'")], [(ENG, "A197ANumberIsCancelledOnce"), ("./internal/journal", "A197")]),
    ("M11", "park 전이 삭제(IN_DOUBT 방치)", [(E, "\tcase out.State == journal.StateInDoubt && out.AttemptID != \"\":", "\tcase false && out.State == journal.StateInDoubt && out.AttemptID != \"\":")], [(ENG, "A197AnAmbiguousCancel")]),
    ("M13", "주기 끝 호출을 판정 경로로 옮김", [(L, "\to.runExternalBuyStage(ctx)\n\tif o.opts.ExitOrderLog == nil {", "\tif o.opts.ExitOrderLog == nil {"), (X, "\t\t\to.noteExternalBuyBlock(m)\n", "\t\t\to.noteExternalBuyBlock(m)\n\t\t\to.runExternalBuyStage(ctx)\n")], [(ENG, "A197TheStageRunsAfterEvery")]),
    ("M14", "B1 뒤로 옮김", [(L, "\to.runExternalBuyStage(ctx)\n\tif o.opts.ExitOrderLog == nil {\n\t\treturn\n\t}\n", "\tif o.opts.ExitOrderLog == nil {\n\t\treturn\n\t}\n\to.runExternalBuyStage(ctx)\n")], [(ENG, "A197TheStageRunsWithoutTheDedicatedLogger")]),
    ("M15", "주기당 상한 삭제", [(E, "\to.handleExternalBuy(ctx, item)\n}", "\tfor _, it := range queue {\n\t\to.handleExternalBuy(ctx, it)\n\t}\n\t_ = item\n}")], [(ENG, "A197OnePositionPerCycle")]),
    ("M16", "PLACE park 가드 삭제", [(E, "if err != nil || parked {", "if err != nil || (parked && false) {")], [(ENG, "A197AParkedEngineOrder")]),
    ("M17", "알림 key 에서 주문 번호 삭제", [(E, "A197AttemptKeyPrefix+item.positionID+\"|\"+target.orderID", "A197AttemptKeyPrefix+item.positionID")], [(ENG, "A197OneExternalBuyPerCycle")]),
    ("M18", "「시도」 알림을 취소 뒤로 옮김", [(E, "\to.recordExternalBuyCritical(ctx, A197AttemptKeyPrefix", "\tdefer o.recordExternalBuyCritical(ctx, A197AttemptKeyPrefix")], [(ENG, "A197AnExternalBuyIsCancelled")]),
    ("M19", "「이미 냈다」 키에 거래일 추가(M-1)", [(J, "WHERE a.kind = ? AND a.target_order_id = ? AND TRIM(i.source) = ?", "WHERE a.kind = ? AND a.target_order_id = ? AND TRIM(i.source) = ? AND TRIM(i.trading_day) = (SELECT MAX(TRIM(trading_day)) FROM intents)")], [(ENG, "A197ANumberIsCancelledOnce")]),
    ("M20", "실행 직전 재검증 삭제(M-3)", [(E, "\tif !externalBuyItemCurrent(item, positions) {", "\tif false && !externalBuyItemCurrent(item, positions) {")], [(ENG, "A197APositionClosed")]),
    ("M21", "재검증의 세대 비교 삭제(M-3)", [(E, "return p.InstanceSeq == item.instanceSeq && p.State", "return p.State")], [(ENG, "A197TheItemMustStill")]),
    ("M22", "토글 audit 실패를 삼킴(O-1)", [(N, "\t\treturn nil, fmt.Errorf(\"engine: recording the automation-gate audit trail: %w\", err)\n\t}\n\n\t// --- 1. account", "\t\t_ = err\n\t}\n\n\t// --- 1. account")], [(ENG, "A197Startup|A197AnUnrecordable")]),
    # 아래 셋은 design §4 의 22 개 밖 — gstack 리뷰 r2 수정(종결되지 않은 취소 attempt 의 알림 문장)을 잼.
    ("R2a", "종결 전 attempt 갈래 삭제(r2)", [(E, "\tcase externalBuyCancelHeld(out):", "\tcase false && externalBuyCancelHeld(out):")], [(ENG, "A197")]),
    ("R2b", "종결 전 판정을 늘 거짓으로(r2)", [(E, "return out.AttemptID != \"\" && !out.State.IsTerminal()", "return false")], [(ENG, "A197OnlyAnUnfinished|A197TheUncleared")]),
    ("R2c", "재시작 문장 선택 삭제(r2)", [(E, "\t\tif held {", "\t\tif false && held {")], [(ENG, "A197TheUncleared")]),
]
NOT_RUN = [("M12", "park 에 계정 래치 추가", "만들 수 없음 — 관측기에 진입 게이트 핸들이 없다(ExitObserverOptions 에 Gate 칸 없음). 계정 래치 부재는 시험 TestA197AnAmbiguousCancelIsParkedWithoutAnAccountLatch 가 엔진 조립의 진입 게이트로 직접 단언함")]


def run(cmd, cwd, env):
    p = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True)
    return p.returncode, p.stdout + p.stderr


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--copy", required=True)
    ap.add_argument("--sha", required=True)
    ap.add_argument("--gocache", required=True)
    ap.add_argument("--only", default="")
    a = ap.parse_args()
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", a.sha], capture_output=True, text=True, check=True).stdout.strip()
    copy = pathlib.Path(a.copy)
    if copy.exists() and any(copy.iterdir()):
        sys.exit(f"copy {copy} is not empty — refusing to reuse a copy another run may have mutated")
    copy.mkdir(parents=True, exist_ok=True)
    blob = subprocess.run(["git", "-C", str(ROOT), "archive", head], capture_output=True, check=True).stdout
    tarfile.open(fileobj=io.BytesIO(blob)).extractall(copy)
    (copy / ".a197-mutation-sha").write_text(head + "\n")
    assert (copy / ".a197-mutation-sha").read_text().strip() == head
    env = dict(os.environ, GOCACHE=a.gocache, GOFLAGS="")
    only = set(filter(None, a.only.split(",")))
    mutants = [m for m in MUTANTS if not only or m[0] in only]
    ledger = {"start_sha": head, "copy": "<scratch>", "control": [], "mutants": [], "not_run": NOT_RUN}
    # 무변이 대조군
    for pkg, pat in sorted({t for m in mutants for t in m[3]}):
        rc, out = run(["go", "test", pkg, "-run", pat, "-count=1"], copy, env)
        ledger["control"].append({"pkg": pkg, "run": pat, "rc": rc})
        (OUT / "raw").mkdir(exist_ok=True)
        (OUT / "raw" / f"control-{pkg.strip('./').replace('/', '_')}-{abs(hash(pat)) % 10**6}.txt").write_text(f"# go test {pkg} -run '{pat}' (무변이)\n" + out)
        if rc != 0:
            json.dump(ledger, open(OUT / "ledger.json", "w"), ensure_ascii=False, indent=1)
            sys.exit(f"control failed: {pkg} {pat}")
    for mid, title, edits, tests in mutants:
        saved = {}
        status = None
        for f, old, new in edits:
            p = copy / f
            if f not in saved:
                saved[f] = (p.read_bytes(), sha(p))
            src = p.read_text()
            if src.count(old) != 1:
                status = f"NOT_REACHED({f} anchor count {src.count(old)})"
                break
            p.write_text(src.replace(old, new, 1))
        outs = []
        if status is None:
            pkgs = sorted({t[0] for t in tests})
            rc, out = run(["go", "vet"] + pkgs, copy, env)
            if rc != 0:
                status = "COMPILE"
                outs.append(out)
            else:
                caught = False
                for pkg, pat in tests:
                    rc, out = run(["go", "test", pkg, "-run", pat, "-count=1"], copy, env)
                    outs.append(f"## go test {pkg} -run '{pat}' rc={rc}\n" + out)
                    caught = caught or rc != 0
                status = "CAUGHT" if caught else "SURVIVED"
        for f, (b, h) in saved.items():
            (copy / f).write_bytes(b)
            assert sha(copy / f) == h, f"restore failed for {f}"
        (OUT / "raw" / f"{mid}.txt").write_text(f"# {mid} {title} — {status}\n" + "\n".join(outs))
        ledger["mutants"].append({"id": mid, "title": title, "status": status, "tests": tests})
        print(mid, status, flush=True)
        json.dump(ledger, open(OUT / "ledger.json", "w"), ensure_ascii=False, indent=1)


if __name__ == "__main__":
    main()
