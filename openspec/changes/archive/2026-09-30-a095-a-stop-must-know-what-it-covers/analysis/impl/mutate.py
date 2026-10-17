#!/usr/bin/env python3
"""a095 구현 로트 변이 하네스.

규율(기억 「뮤테이션 원복은 baseline이 맞아야 한다」 · 「사본 하네스는 무변이 대조군 GREEN 선행」):
- 저장소 루트는 인자로 받음(절대경로를 박지 않음). 대상 파일의 시작 sha256 을 기록하고, 판마다 원문 바이트로 되쓰고 sha 를 다시 단언함.
- 무변이 대조군(control)이 GREEN 이 아니면 멈춤 — 눈먼 계측기가 CAUGHT 를 찍지 않게.
- 변이마다 치환 대상 문자열이 정확히 한 번 있어야 함(0 이면 계측기가 닿지 않은 것 — 오류로 멈춤).
- 한 번에 한 판(병렬 없음).

사용: python3 mutate.py <repo-root> [변이 id ...]
"""
import hashlib
import pathlib
import subprocess
import sys

PKG_ENGINE = "./internal/app/engine/"
RUN = "TestA095|TestTheUnmanagedAlertSurvivesAdoptionBeingOff|TestAFailedIncludeCycleSaysTriedNotOff|TestTheDefaultRowSaysOffAndUndesignated|TestAnExternalIncreaseAfterAdoptionIsReported"

ADOPTION = "internal/app/engine/adoption.go"
LOOP = "internal/app/engine/reconcileloop.go"
EVENT = "internal/obs/event.go"

# (id, 파일, 옛 문자열, 새 문자열, 설명)
MUTANTS = [
    ("G1", ADOPTION, "if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {",
     "if fact == factEnabledFailed || fact == factIncludeFailed {", "critical 판정이 알림 켜짐을 보지 않음"),
    ("G2", ADOPTION, "if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {",
     "if d.opts.NotificationsEnabled || (fact == factEnabledFailed || fact == factIncludeFailed) {", "&& → ||"),
    ("G16", ADOPTION, "if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {",
     "if d.opts.NotificationsEnabled && fact == factEnabledFailed {", "include 지정 시도 실패를 normal 로(Q2(b) 정정 되돌림)"),
    ("G17", ADOPTION, "adopted[c.position.ID] = adoptOutcome{result: adoptFailed, at: d.clk.Now()}",
     "adopted[c.position.ID] = adoptOutcome{result: adoptFailed}", "실패 시각을 싣지 않음 — 보고 순간으로 대체(codex P2 되돌림)"),
    ("G3", ADOPTION, """	if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {
		d.alertAdoptionFailed(ctx, p, fact, outcome.at)
		return
	}
	if d.unmanaged[p.ID][fact] {
		return
	}""", """	if d.unmanaged[p.ID][fact] {
		return
	}
	if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {
		if d.unmanaged[p.ID] == nil {
			d.unmanaged[p.ID] = map[string]bool{}
		}
		d.unmanaged[p.ID][fact] = true
		d.alertAdoptionFailed(ctx, p, fact, outcome.at)
		return
	}""", "critical 이 메모리 래치를 거침(5판 이전 모양)"),
    ("G4", ADOPTION, 'Key:   reconcileUnmanagedKey(obs.EventExitPositionUnmanaged, fact, p.ID),',
     'Key:   string(obs.EventExitPositionUnmanaged) + "|" + p.ID,', "normal key 를 옛 철자(exit 자리와 같음)로"),
    ("G5", ADOPTION, 'return string(t) + "|reconcile|" + fact + "|" + positionID',
     'return string(t) + "|reconcile|" + positionID', "key 에서 조건 칸 제거"),
    ("G6", ADOPTION, """		if result == adoptFailed {
			return factEnabledFailed""", """		if true {
			return factEnabledFailed""", "편입 켜짐 칸이 결과를 무시 — 연기도 시도 실패로"),
    ("G7", ADOPTION, """	case d.opts.Adoption.Enabled:
		if result == adoptFailed {""", """	case d.opts.Adoption.Included(p.Symbol) && false, d.opts.Adoption.Enabled && !d.opts.Adoption.Included(p.Symbol):
		if result == adoptFailed {""", "편입 켜짐 ∧ include 지정 종목이 켜짐 칸을 벗어남"),
    ("G8", ADOPTION, "	if d.unmanaged[p.ID][fact] {\n		return\n	}",
     "	if len(d.unmanaged[p.ID]) > 0 {\n		return\n	}", "normal 래치가 포지션 단위(사실 무시)"),
    ("G9", ADOPTION, "	if d.unmanaged[p.ID][fact] {\n		return\n	}", "", "normal 래치 제거"),
    ("G10", ADOPTION, "		adopted[c.position.ID] = adoptOutcome{result: adoptFailed, at: d.clk.Now()}\n", "", "시도 실패를 기록하지 않음(연기로 읽힘)"),
    ("G12", ADOPTION, 'Body: moment + "(UTC)에 편입 시도가 실패했다.',
     'Body: "편입 시도가 실패했다.', "critical 문장에서 시각을 뺌(Q8)"),
    ("G13", LOOP, "	opts.NotificationsEnabled = c.Config.Engine.Notifications.Enabled\n", "",
     "생산 조립이 로드된 켜짐을 넘기지 않음"),
    ("G14", LOOP, "	opts.NotificationsEnabled = c.Config.Engine.Notifications.Enabled\n",
     "	opts.NotificationsEnabled = opts.NotificationsEnabled || c.Config.Engine.Notifications.Enabled\n",
     "생산 조립이 호출자 값을 남김"),
    ("G15", EVENT, "	EventExitPositionAdoptionFailed: true,\n", "", "등급 표 등재 제거"),
    ("I1", ADOPTION, "			} else {\n				d.checkEngineOpenedIncrease(ctx, p)\n			}", "			}",
     "엔진 개설 포지션 검사 호출 제거"),
    ("I2", ADOPTION, "riskcalc.SubDecimal(a.NewQuantity, a.PrevQuantity)",
     "riskcalc.SubDecimal(a.PrevQuantity, a.NewQuantity)", "순증 부호 반대(열 바꿔치기)"),
    ("I3", ADOPTION, """	cmp, err := riskcalc.CompareDecimal(net, "0")
	if err != nil || cmp <= 0 {""", """	cmp, err := riskcalc.CompareDecimal(net, "0")
	if err != nil || cmp < 0 {""", "순증 0 을 증가로(off-by-one)"),
    ("I4", ADOPTION, "riskcalc.SubDecimal(a.NewQuantity, a.PrevQuantity)",
     "riskcalc.SubDecimal(a.NewQuantity, a.ExpectedPrevQuantity)", "prev 열을 expected_prev 로(동등 후보)"),
    ("I5", ADOPTION, """		cmp, err := riskcalc.CompareDecimal(p.Quantity, last)
		if err != nil || cmp <= 0 {""", """		cmp, err := riskcalc.CompareDecimal(p.Quantity, last)
		if err != nil || cmp < 0 {""", "최대 수량 래치 off-by-one(같은 수량 재보고)"),
    ("I6", ADOPTION, "	d.grown[p.ID] = p.Quantity\n", "", "최대 수량 래치를 옮기지 않음"),
    ("I7", ADOPTION, """	if !d.newGrowthMaximum(p) {
		return
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|grown|" + p.ID,
		Title: d.label(p.Symbol) + " 편입 후""", """	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|grown|" + p.ID,
		Title: d.label(p.Symbol) + " 편입 후""", "편입 포지션 자리의 래치 제거"),
    ("I8", ADOPTION, "	if err != nil || len(adjustments) == 0 {", "	if err != nil {", "조정 0 조기 반환 제거(동등 후보)"),
    ("B2", ADOPTION, """	adoption, err := d.opts.Journal.AdoptionOf(ctx, p.ID)
	if err != nil {
		return
	}""", """	adoption, err := d.opts.Journal.AdoptionOf(ctx, p.ID)
	if err != nil {
		d.alertUnmanaged(ctx, p, adoptOutcome{result: adoptFailed})
		return
	}""", "R2-B2 부활 — 조회 오류를 무관리 보고로(3.1 이 막아야 함)"),
]


def sha(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run_tests(root: pathlib.Path) -> bool:
    proc = subprocess.run(["go", "test", PKG_ENGINE, "-run", RUN, "-count=1", "-timeout", "900s"],
                          cwd=root, capture_output=True, text=True)
    return proc.returncode == 0


def main() -> int:
    root = pathlib.Path(sys.argv[1]).resolve()
    only = set(sys.argv[2:])
    files = {f: (root / f) for f in {ADOPTION, LOOP, EVENT}}
    originals = {f: p.read_bytes() for f, p in files.items()}
    start = {f: sha(p) for f, p in files.items()}
    for f in sorted(start):
        print(f"start {start[f][:16]} {f}")
    if not run_tests(root):
        print("CONTROL RED — 무변이 대조군이 GREEN 이 아님, 멈춤")
        return 2
    print("control GREEN")
    caught = survived = 0
    for mid, f, old, new, why in MUTANTS:
        if only and mid not in only:
            continue
        path = files[f]
        text = originals[f].decode()
        if text.count(old) != 1:
            print(f"{mid} UNREACHABLE — 치환 대상 {text.count(old)}회 ({why})")
            return 3
        path.write_text(text.replace(old, new, 1))
        try:
            ok = run_tests(root)
        finally:
            path.write_bytes(originals[f])
            if sha(path) != start[f]:
                print(f"{mid} RESTORE FAILED")
                return 4
        if ok:
            survived += 1
            print(f"{mid} SURVIVED — {why}")
        else:
            caught += 1
            print(f"{mid} CAUGHT — {why}")
    print(f"caught {caught} · survived {survived}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
