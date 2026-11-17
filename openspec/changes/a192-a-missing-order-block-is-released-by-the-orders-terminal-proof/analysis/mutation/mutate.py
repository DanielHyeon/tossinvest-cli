#!/usr/bin/env python3
"""a192 변이 하네스 — 사본 대상, 한 번에 한 판.

규율(기억 「뮤테이션 원복은 baseline 이 맞아야 한다」「뮤테이션은 대상에 닿아야 한다」):
- 저장소 경로는 이 파일 위치에서 유도함(<repo>/openspec/changes/<id>/analysis/mutation/mutate.py).
- 시작 때 HEAD 가 base-commit 과 같고, 대상 소스들의 sha256 을 적어 끝에 다시 대조함(원본 무변경 단언).
- 변이마다 **새 사본**(pid 를 이름에 넣은 임시 디렉터리)을 만들고, 끝나면 사본을 지움 — 원본을 고치거나
  git checkout 으로 되돌리지 않음.
- 무변이 대조군(m0)이 GREEN 이어야 변이를 돎. 변이는 바꿀 문장이 **정확히 한 번** 있어야 적용(닿음의 구조
  대조군)하고, 각 시험 안의 양성 대조(그 경로에 실제로 갔는지 단언)가 행동 쪽 닿음을 맡음.
- GOFLAGS=-trimpath -buildvcs=false, 전용 GOCACHE.
"""

from __future__ import annotations

import hashlib
import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path

HERE = Path(__file__).resolve()
CHANGE = HERE.parents[2]
REPO = HERE.parents[5]
COPY_ITEMS = ["go.mod", "go.sum", "internal", "cmd"]

RECONCILE = "./internal/reconcile/"
JOURNAL = "./internal/journal/"
ENGINE = "./internal/app/engine/"
A192 = "TestA192"

MISMATCH = "internal/reconcile/mismatch.go"
RELEASE = "internal/reconcile/order_terminal_release.go"
JRELEASE = "internal/journal/reconcile_order_terminal_release.go"
LOG = "internal/app/engine/a192_order_terminal_release_log.go"

# (id, 설명, 파일, 옛 문장, 새 문장, 시험 패키지, -run)
MUTANTS = [
    ("m1", "삽입 증명을 건너뜀(entered 무시)", RELEASE,
     "if err == nil && entered && bindable", "if err == nil && bindable", [RECONCILE], A192),
    ("m2", "재기동 뒤에도 결속이 있다고 침(결속 없는 missing-order 에 활성 행으로 결속을 새로 만듦)", RELEASE,
     """		t.terminal.seen[key] = true
	}
	t.updateOrderTerminalBindings(diff)""",
     """		t.terminal.seen[key] = true
	}
	for _, missing := range diff.MissingOrders {
		sym := strings.ToUpper(strings.TrimSpace(missing.Symbol))
		if t.terminal.bindings[sym] != nil {
			continue
		}
		if states, err := t.Journal.ActiveReconcileStates(context.Background()); err == nil {
			for _, st := range states {
				if strings.EqualFold(st.Symbol, sym) {
					if t.terminal.bindings == nil {
						t.terminal.bindings = map[string]*terminalBinding{}
					}
					t.terminal.bindings[sym] = &terminalBinding{rowID: st.ID, symbol: sym, orders: map[boundOrder]struct{}{}}
				}
			}
		}
	}
	t.updateOrderTerminalBindings(diff)""",
     [RECONCILE], A192),
    ("m3", "워터마크 비교를 뒤집음", JRELEASE,
     "fill_events WHERE id > ? AND", "fill_events WHERE id <= ? AND", [RECONCILE, JOURNAL], A192),
    ("m4", "무장 없이 첫 통과에서 해제", RELEASE,
     """	if !b.armed || b.armDigest != digest {
		verdict.outcome, verdict.reason = terminalArm, "armed"
		return verdict
	}""", "", [RECONCILE], A192),
    ("m5", "무장 digest 무효화 삭제", RELEASE,
     "if !b.armed || b.armDigest != digest {", "if !b.armed {", [RECONCILE], A192),
    ("m6", "C3 수량 흔적 무시", RELEASE, "	if b.tainted {", "	if false {", [RECONCILE], A192),
    ("m7", "C5 lineage 무시", RELEASE,
     "if order.identity.OrderID != order.original || snapshot.Amended {", "if false {", [RECONCILE], A192),
    ("m8", "C7 fail_closed 무시", RELEASE,
     "if !snapshot.Record.Terminal || snapshot.Record.FailClosed {", "if !snapshot.Record.Terminal {", [RECONCILE], A192),
    ("m9", "C8 엄격 → 이상", RELEASE,
     "if err != nil || !asOf.After(latest) {", "if err != nil || asOf.Before(latest) {", [RECONCILE], A192),
    ("m10", "C10 같은 번호 open 무시", RELEASE,
     "if cmp.externalIDs[order.identity.OrderID] || cmp.externalIDs[order.original] {", "if false && cmp.externalIDs[order.original] {", [RECONCILE], A192),
    ("m11", "해제 tx 안에서 공개 Journal 읽기 호출(연결 하나 — 자기 대기)", RELEASE,
     "			verdict := orderTerminalReleaseVerdict(binding, comparison, proof)",
     "			_ = t.Terminal.ReadOrderTerminalProof(ctx, release.request)\n"
     "			verdict := orderTerminalReleaseVerdict(binding, comparison, proof)",
     [RECONCILE], A192),
    ("m12", "tx 재판정 생략(1차 판정만 믿음)", RELEASE,
     "			return verdict.outcome == terminalReleaseOutcome",
     "			return true || verdict.outcome == terminalReleaseOutcome", [RECONCILE], A192),
    ("m13", "「전부」 → 「하나라도」(비종결 주문을 건너뜀)", RELEASE,
     """		if !snapshot.Record.Terminal || snapshot.Record.FailClosed {
			return hold("not_clean_terminal")
		}""", """		if !snapshot.Record.Terminal || snapshot.Record.FailClosed {
			continue
		}""", [RECONCILE], A192),
    ("m14", "빈 결속 표본 허용", RELEASE,
     """	if len(orders) == 0 {
		// 빈 표본의 전칭 판정은 자동 참이 되므로 따로 막음.
		return hold("identity_incomplete")
	}""", "", [RECONCILE], A192),
    ("m15", "해제 원인을 ADJUSTMENT_APPLIED 로", JRELEASE,
     "nowText, ReconcileReleaseOrderTerminalConfirmed, proof.State.Evidence",
     "nowText, ReconcileReleaseAdjustmentApplied, proof.State.Evidence", [RECONCILE, JOURNAL, ENGINE], A192),
    ("m16", "종결 증명 판정을 크레딧 판정 앞으로", MISMATCH,
     """			if !creditUsableBy(t.adjusted[symbol], diff.AsOf) {
				// Either nothing was written for this symbol, or what was written""",
     """			if t.orderTerminalCandidate(ctx, &block, diff, disputed) {
				out.Cleared = append(out.Cleared, block)
				continue
			}
			if !creditUsableBy(t.adjusted[symbol], diff.AsOf) {
				// Either nothing was written for this symbol, or what was written""", [RECONCILE], A192),
    ("m17", "Terminal == nil 에서도 판정", RELEASE,
     """	if t.Terminal == nil || t.Journal == nil {
		return false
	}
	symbol := strings.ToUpper(strings.TrimSpace(block.Symbol))
	b := t.terminal.bindings[symbol]""", """	symbol := strings.ToUpper(strings.TrimSpace(block.Symbol))
	b := t.terminal.bindings[symbol]""", [RECONCILE], A192),
    ("m18", "로그에 주문번호(Detail)를 실음", LOG,
     '"bound_orders", bound)', '"bound_orders", bound, "detail", block.Detail)', [ENGINE], A192),
    ("m19", "갈래 나누기 전 결속 갱신 삭제(흔적이 조기 반환 뒤로 밀림)", RELEASE,
     """		t.terminal.seen[key] = true
	}
	t.updateOrderTerminalBindings(diff)""", """		t.terminal.seen[key] = true
	}""", [RECONCILE], A192),
    ("m20", "재시도 삽입도 결속", RELEASE,
     "block.missing != nil && !t.terminal.seen[block.Key()]", "block.missing != nil", [RECONCILE], A192),
    ("m21", "Restore 의 결속 지우기 삭제", MISMATCH,
     "	t.terminal = terminalState{}\n", "", [RECONCILE], A192),
    ("m22", "배관: 판정에 disputed 대신 거짓을 넘김", RELEASE,
     "comparison := orderTerminalComparison(t.AccountRef, diff, disputed[symbol])",
     "comparison := orderTerminalComparison(t.AccountRef, diff, false)", [RECONCILE], A192),
    ("m23", "배관: 요청 워터마크를 0 으로", RELEASE,
     "StateID: b.rowID, Symbol: b.symbol, FillWatermark: b.fillWatermark,",
     "StateID: b.rowID, Symbol: b.symbol, FillWatermark: 0,", [RECONCILE], A192),
    ("m24", "배관: 결속 생성 때 워터마크를 버림", RELEASE,
     "rowID: state.ID, symbol: symbol, fillWatermark: watermark,", "rowID: state.ID, symbol: symbol, fillWatermark: watermark * 0,",
     [RECONCILE], A192),
    ("m25", "배관: 판정에 추적기 계좌 대신 비교 계좌를 넘김", RELEASE,
     "comparison := orderTerminalComparison(t.AccountRef, diff, disputed[symbol])",
     "comparison := orderTerminalComparison(diff.AccountRef, diff, disputed[symbol])", [RECONCILE], A192),
    ("m26", "배관: 해제 tx 에 다른 결속(무장 없는 복사)을 넘김", RELEASE,
     "	binding, comparison := release.binding, release.comparison",
     "	binding, comparison := release.binding, release.comparison\n	binding.armed = false", [RECONCILE], A192),
    ("m27", "RunOnce 의 로그 호출 삭제", "internal/app/engine/reconcileloop.go",
     "	d.logOrderTerminalReleases(outcome)\n", "", [ENGINE], A192),
    ("m28", "ValidReconcileReleaseCause 에서 새 원인 삭제", "internal/journal/reconcile_states.go",
     "ReconcileReleaseOperator, ReconcileReleaseOrderTerminalConfirmed:", "ReconcileReleaseOperator:",
     [RECONCILE, JOURNAL], A192),
    ("m29", "buildGateway 의 Terminal 배선 삭제", "internal/app/engine/gateway.go",
     "		Terminal: in.journal,\n", "", [ENGINE], A192),
    ("m30", "C12 의 종목 없는 체결 갈래 삭제", JRELEASE,
     "AND (UPPER(TRIM(symbol)) = ? OR TRIM(symbol) = ''))", "AND UPPER(TRIM(symbol)) = ?)", [JOURNAL], A192),
    ("m31", "1차 판정 사유를 남기지 않음", RELEASE,
     "	b.reason = verdict.reason\n", "", [ENGINE], A192),
    ("m32", "tx 재판정 사유를 남기지 않음", RELEASE,
     """		b.reason = "tx_" + reason""", """		_ = reason""", [RECONCILE], A192),
    ("m33", "사라진 블록의 결속을 버리지 않음", RELEASE,
     """		if !hasBlockFor(t.blocks, symbol) {
			delete(t.terminal.bindings, symbol)
		}""", """		_ = symbol""", [RECONCILE], A192),
    ('m34', 'Manager 표본 g1 — C2: 이미 해제된 행 검사 삭제', RELEASE,
     'if !proof.StateFound || !state.Active() || state.ID != b.rowID ||',
     'if !proof.StateFound || state.ID != b.rowID ||', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m35', 'Manager 표본 g2 — C2: 시장 범위 행 검사 삭제', RELEASE,
     'state.Cause != journal.ReconcileCauseQuantityMismatch || strings.TrimSpace(state.ScopeMarket) != "" ||',
     'state.Cause != journal.ReconcileCauseQuantityMismatch ||', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m36', 'Manager 표본 g3 — C2: 행 계좌 = 추적기 계좌 검사 삭제', RELEASE,
     'cmp.trackerAccount == "" || strings.TrimSpace(state.AccountRef) != cmp.trackerAccount {',
     'cmp.trackerAccount == "" {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m37', 'Manager 표본 g4 — C4: 비교 계좌 = 추적기 계좌 검사 삭제', RELEASE,
     'if id.AccountRef != cmp.trackerAccount || cmp.diffAccount != cmp.trackerAccount || id.Symbol != b.symbol {',
     'if id.AccountRef != cmp.trackerAccount || id.Symbol != b.symbol {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m38', 'Manager 표본 g5 — C11: 비교가 진입을 막는지 검사 삭제', RELEASE,
     'if cmp.blocksEntry {',
     'if false {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m39', 'Manager 표본 g6 — C13: 무장보다 나중 비교 검사 삭제', RELEASE,
     'if !asOf.After(b.armAsOf) {',
     'if false {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m40', 'Manager 표본 g7 — 읽기 실패 = 유지 삭제', RELEASE,
     'if proof.ReadErr != nil {',
     'if false {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m41', 'Manager 표본 g8 — 해제 UPDATE 의 released_at IS NULL 고정 삭제', JRELEASE,
     ' WHERE id = ? AND released_at IS NULL`,',
     ' WHERE id = ?`,', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m42', 'Manager 표본 g9 — RowsAffected == 1 검사 삭제', JRELEASE,
     'if changed != 1 {',
     'if false {', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m43', 'Manager 표본 g10 — persist: tx 재판정이 유지여도 추적기 기억에서 블록 삭제', MISMATCH,
     'if released {\n\t\t\t\tresult.released = append(result.released, block)\n\t\t\t}\n\t\t\tcontinue',
     'result.released = append(result.released, block)\n\t\t\t_ = released\n\t\t\tcontinue', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m44', 'Manager 표본 g11 — 수량 블록의 삽입도 결속', RELEASE,
     'bindable := t.Terminal != nil && block.missing != nil && !t.terminal.seen[block.Key()]',
     'bindable := t.Terminal != nil && !t.terminal.seen[block.Key()]', [RECONCILE, JOURNAL, ENGINE], A192),
    ('m45', 'Manager 표본 g12 — 수량 불일치 실격 흔적 삭제', RELEASE,
     '\t\t\tb.tainted = true\n\t\t}\n\t}\n\tfor _, external := range diff.ExternalPos {',
     '\t\t\t_ = b\n\t\t}\n\t}\n\tfor _, external := range diff.ExternalPos {', [RECONCILE, JOURNAL, ENGINE], A192),
]


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def git(*args: str) -> str:
    return subprocess.run(["git", "-C", str(REPO), *args], check=True, capture_output=True, text=True).stdout.strip()


def make_copy(tag: str) -> Path:
    root = Path(tempfile.mkdtemp(prefix=f"a192-mut-{os.getpid()}-{tag}-"))
    for item in COPY_ITEMS:
        src = REPO / item
        if src.is_dir():
            shutil.copytree(src, root / item, symlinks=True)
        else:
            shutil.copy2(src, root / item)
    return root


def run_tests(root: Path, pkgs: list[str], run: str, cache: Path) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath -buildvcs=false", GOCACHE=str(cache))
    cmd = ["go", "test", "-count=1", "-timeout", "120s", "-run", run, *pkgs]
    try:
        proc = subprocess.run(cmd, cwd=root, env=env, capture_output=True, text=True, timeout=300)
        return proc.returncode, proc.stdout + proc.stderr
    except subprocess.TimeoutExpired as exc:
        return 124, f"harness timeout: {exc}"


def main() -> int:
    only = set(sys.argv[1:])
    base = (CHANGE / "base-commit.txt").read_text().strip()
    head = git("rev-parse", "HEAD")
    if head != base:
        print(f"ABORT: HEAD {head} != base {base}")
        return 2
    targets = sorted({m[2] for m in MUTANTS})
    before = {t: sha(REPO / t) for t in targets}
    cache = Path(tempfile.gettempdir()) / f"a192-mut-gocache-{os.getuid()}"
    cache.mkdir(exist_ok=True)
    print(f"repo {REPO}\nhead {head}\ntargets " + " ".join(f"{t}={before[t][:12]}" for t in targets))

    # 무변이 대조군 — 모든 패키지·실행 조합이 사본에서 GREEN 이어야 함.
    combos = sorted({(tuple(m[5]), m[6]) for m in MUTANTS})
    root = make_copy("m0")
    try:
        for pkgs, run in combos:
            rc, out = run_tests(root, list(pkgs), run, cache)
            print(f"m0 control {' '.join(pkgs)} -run {run}: rc={rc}")
            if rc != 0:
                print(out[-3000:])
                print("ABORT: the unmutated control is not GREEN")
                return 3
    finally:
        shutil.rmtree(root, ignore_errors=True)

    tally = {"CAUGHT": 0, "SURVIVED": 0, "NOT_REACHED": 0}
    for mid, desc, path, old, new, pkgs, run in MUTANTS:
        if only and mid not in only:
            continue
        root = make_copy(mid)
        try:
            target = root / path
            text = target.read_text()
            count = text.count(old)
            if count != 1:
                verdict = "NOT_REACHED"
                detail = f"pattern occurs {count} time(s)"
            else:
                target.write_text(text.replace(old, new))
                start = time.time()
                rc, out = run_tests(root, pkgs, run, cache)
                if "[build failed]" in out or "build failed" in out:
                    # 컴파일이 안 되는 변이는 행동을 재지 못함 — 잡음으로 세지 않음.
                    verdict = "NOT_REACHED"
                else:
                    verdict = "CAUGHT" if rc != 0 else "SURVIVED"
                failing = [l.strip() for l in out.splitlines() if l.startswith("--- FAIL") or "panic:" in l
                           or "harness timeout" in l or "build failed" in l or l.startswith("FAIL\t")
                           or "_test.go:" in l]
                detail = f"rc={rc} {time.time() - start:.1f}s " + "; ".join(failing[:6])
            tally[verdict] += 1
            print(f"{mid} {verdict} — {desc} — {detail}")
        finally:
            shutil.rmtree(root, ignore_errors=True)

    after = {t: sha(REPO / t) for t in targets}
    if after != before or git("rev-parse", "HEAD") != head:
        print("ABORT: the original tree changed during the run")
        return 4
    print("tally " + " ".join(f"{k}={v}" for k, v in tally.items()))
    print("original tree unchanged (sha256 re-checked)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
