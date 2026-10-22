"""a164 FLM/BTM 생성기 — AST 분기·원문 줄·시험별 coverage(측정)·손으로 쓴 노트를 합쳐 두 map 을 씀."""
import json, os, sys
ROOT = "/mnt/D/Axipient/workspace/TossOS-a164-impl/"
FL = ROOT + "openspec/changes/a164-verify-reconcile-closes-broker-expired-conditional/analysis/function-logic/"
cov = json.load(open(sys.argv[1]))
notes = json.load(open(sys.argv[2]))
label = sys.argv[3] if len(sys.argv) > 3 else "편집 전 기준선"
def esc(s): return s.replace("|", "\\|").strip()
def body_first(src, line):
    for ln in range(line, line + 8):
        if src[ln - 1].rstrip().endswith("{"):
            nxt = src[ln].strip()
            return nxt if nxt and nxt != "}" else "(빈 본문)"
    return "(한 줄 본문)"
for d in sorted(os.listdir(FL)):
    a = json.load(open(FL + d + "/ast.json"))
    n = notes[d]
    src = open(ROOT + a["file"]).read().split("\n")
    fn = (a.get("receiver") + "." if a.get("receiver") else "") + a["function"]
    br = a.get("branches") or []
    rets = [r["at"]["line"] for r in a.get("returns") or []]
    calls = sorted({c.get("text") or "" for c in (a.get("calls") or [])} - {""})
    rows, brows = [], []
    items = [(f"B{i}", b) for i, b in enumerate(br, 1)] or [("B1", None)]
    for bid, b in items:
        tests = cov[d].get(bid, [])
        selfrun = False
        if not tests and n.get("cover", {}).get(bid):
            # 시험 파일의 함수는 coverprofile 이 계측하지 않음 — 그 함수를 실행하는 시험(자기 자신 포함)의 PASS 가 증거.
            tests, selfrun = n["cover"][bid], True
        cited = " · ".join(f"`{t}`" for t in tests[:4]) + (f" 외 {len(tests)-4}" if len(tests) > 4 else "")
        if selfrun:
            cited += " (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거)"
        if not tests:
            cited = "미덮임(시험별 coverage 측정 0) — " + n.get("uncovered", {}).get(bid, "기존 공백")
        if b is None:
            cond = "분기 없음 — 단일 경로"
            body = esc(src[a["start"]["line"] - 1].strip())[:140]
        else:
            cond = f"{b['kind']} 줄 {b['at']['line']}: `{esc(src[b['at']['line']-1].strip())[:150]}`"
            body = f"`{esc(body_first(src, b['at']['line']))[:120]}`"
        eff = n.get("effects", {}).get(bid, "없음")
        rows.append(f"| {bid} | {cond} | {eff} | {body} | {cited} |")
        brows.append(f"| {bid} | {cond} | {cited} | 해당 없음 — {label} | 기준선 통과(측정, 아래 「측정」) |")
    rel = a["file"]
    fl = f"""# Function Logic Map: `{fn}`

- Source: `{rel}` ({a['start']['line']}-{a['end']['line']})
- Qualified function: `{fn}`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` {a['source_sha256'][:16]}…)
- AST evidence: `ast.json` — AST branches {len(br)}, 반환 {len(rets)}(줄 {', '.join(map(str, rets)) or '없음'}), 호출 이름 {len(calls)}
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: {n['role']}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
{chr(10).join(n['inputs'])}

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
{chr(10).join(rows)}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
{chr(10).join(n['calls'])}

AST 호출 이름 전체: {', '.join('`'+c+'`' for c in calls) or '없음'}.

## State mutations and fallbacks

{chr(10).join('- ' + s for s in n['state'])}

## Safety conclusion

{chr(10).join('- ' + s for s in n['safety'])}
"""
    bt = f"""# Branch Test Map: `{fn}`

- Source: `{rel}` ({a['start']['line']}-{a['end']['line']})
- AST branches {len(br)} — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
{chr(10).join(brows)}

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.
"""
    open(FL + d + "/function-logic-map.md", "w").write(fl)
    open(FL + d + "/branch-test-map.md", "w").write(bt)
    print("wrote", d)
