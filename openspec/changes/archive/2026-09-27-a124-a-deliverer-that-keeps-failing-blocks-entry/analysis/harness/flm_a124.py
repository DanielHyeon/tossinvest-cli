#!/usr/bin/env python3
"""a124 tasks 1.3 · 1.4 — 편집 대상 함수의 번들을 편집 뒤 소스로 다시 만든다.

손으로 옮기지 않는 것: ast.json(추출기) · 분기 표의 좌표와 그 줄의 소스 · 편집 전 → 편집 뒤 분기 대응
(difflib 로 두 판의 분기 줄을 정렬) · 분기 시험 표의 시험 칸(branch_coverage.py 측정 JSON).
손으로 쓰는 것: 입력·불변식 · 호출의 의미 · 상태 변화 · 안전 결론 — 아래 PROSE.

쓰기: python3 flm_a124.py --root <저장소> --coverage cov.json [--coverage …] --base 4798d399…
"""

from __future__ import annotations

import argparse
import difflib
import json
import os
import subprocess
from pathlib import Path

CHANGE = "openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry"

# (번들 디렉터리, 소스, 함수, 편집했는가)
TARGETS = [
    ("internal-app-engine--alertdeliverer.cycle", "internal/app/engine/alertdelivery.go", "alertDeliverer.cycle", True),
    ("internal-app-engine--alertdeliverer.deliverone", "internal/app/engine/alertdelivery.go", "alertDeliverer.deliverOne", True),
    ("internal-app-engine--alertdeliverer.release", "internal/app/engine/alertdelivery.go", "alertDeliverer.release", True),
    ("internal-app-engine--context.alertdeliverer", "internal/app/engine/auxiliary.go", "Context.AlertDeliverer", True),
    ("internal-journal--journal.settleunderclaim", "internal/journal/alert_claim.go", "Journal.settleUnderClaim", True),
    ("internal-execgw--entrygate.clear", "internal/execgw/retry.go", "EntryGate.Clear", True),
    ("internal-execgw--entrygate.block", "internal/execgw/retry.go", "EntryGate.Block", False),
]

PROSE = {
    "alertDeliverer.cycle": {
        "inputs": [
            ("`ctx`", "Run 의 수명 ctx", "Run", "행 사이에서 취소를 봄 — 끝나면 조용히 반환"),
            ("`d.led().PendingAlertsForDelivery(ctx, batch, alertAttemptLimit)`", "PENDING 행, 한도 아래 먼저 · 오래된 것 먼저",
             "원장(`outbox.go` 새 메서드)", "오류 → 나열 실패 계수(D8) 뒤 오류 반환"),
            ("`d.batch()`", "> 0 (기본 10)", "`alertDeliveryBatch`", "잘림 판정의 기준"),
        ],
        "calls": [
            ("`d.forgetLapsedHeld`", "만료된 「남이 쥠」 보고 기록 정리", "메모리"),
            ("`d.led().PendingAlertsForDelivery`", "굶주림 없는 선택(R2, AA4 — `PendingAlerts` 불변)", "오류 → `countListFailure`"),
            ("`d.countListFailure`", "나열 실패의 연속(D8 AA2)", "ctx 취소면 세지 않음"),
            ("`d.pruneRecordRuns`", "완전한 나열에 없는 행의 기록 실패 계수 삭제(Q5)", "잘린 나열에서는 부르지 않음"),
            ("`d.deliverOne`", "행 하나의 생애", "행 실패는 사이클 실패가 아님"),
        ],
        "state": [
            "나열이 성공하면 나열 실패 연속을 지움(`listRun`, `listSeen`).",
            "완전한 나열(`len < batch`)일 때만 기록 실패 계수를 거름 — 잘린 나열은 PENDING 이탈의 증거가 아님.",
            "어떤 잠금도 쥐지 않음. 원장 연산 · 발행은 전부 `deliverOne` 과 그 아래에서.",
        ],
        "safety": ("선택 순서와 나열 실패 계수만 더함 — 행을 버리는 길은 없음(한도 행도 선택에 남음).",
                   "yes — 진입 차단(전달 실패 사유)의 판정 입력을 만든다. 차단 방향으로만 틀림."),
    },
    "alertDeliverer.deliverOne": {
        "inputs": [
            ("`alert`", "나열 시점의 PENDING 행 — `Attempts` 는 판정에 쓰지 않음(D1)", "`cycle`", "나열 뒤 정착되면 임차가 「이미 정산됨」"),
            ("`d.led().ClaimAlertByID`", "임차 3 종 결과", "원장", "오류 → 행별 기록 실패 계수(D8)"),
            ("`d.Publisher`", "nil 허용 — nil 도 실패 시도(D3)", "배선", "nil → 고정 원인 문구로 실패 기록"),
        ],
        "calls": [
            ("`d.led().ClaimAlertByID`", "이 행의 발송 권한", "오류 → `countRecordFailure`"),
            ("`d.forgetHeld` · `d.reportHeld` · `d.forgetRecordRun`", "보고 · 계수 기록 정리", "메모리"),
            ("`d.Publisher.Publish`", "원격 전송 1 회", "잠금 없이 — 재시도는 다음 사이클"),
            ("`d.recordFailedAttempt`", "실패 기록 → 반납 → 판정(D1 실패 기록 표)", "판정은 원칙 E"),
            ("`d.recordDelivery`", "전달 정산 → 실패면 즉시 판정(D1 전달 정산 표)", "임차 유지"),
            ("`d.logf`", "기존 줄 — 토큰 없음", "관측만"),
        ],
        "state": [
            "원장 쓰기는 전부 하위 함수에 있다. 이 함수 자체는 메모리 기록(`heldReported`, `recordRuns`)만 바꾼다.",
            "publisher 가 없을 때도 이제 실패 시도가 기록된다(`alertNoPublisherCause`) — 편집 전에는 로그와 반납뿐이었다(D3).",
            "임차 결과가 「이미 정산됨」이면 그 행의 기록 실패 연속을 지운다(Y3).",
        ],
        "safety": ("판정은 하위 함수로 옮겼고 이 함수는 흐름만 가른다. 잠금을 쥐지 않는다.",
                   "yes — 전달 실패 → 진입 차단 · 모드 승격의 입구. 새 문은 없음(오늘 안 잠기던 자리가 잠긴다)."),
    },
    "alertDeliverer.release": {
        "inputs": [("`id` · `token`", "이 사이클이 끝낸 행의 임차", "`deliverOne` 하위", "오류는 로그")],
        "calls": [("`d.led().ReleaseAlertClaim`", "임차 반납(떼어 낸 ctx, 5 s)", "오류는 로그만 — 반납 Applied 는 기록 실패 계수를 지우지 않음(X3)")],
        "state": ["편집은 한 줄 — `d.Journal` → `d.led()`(시험 결함 주입용 원장 래퍼). 생산 값은 같은 `*journal.Journal`."],
        "safety": ("동작 불변(원장 대상만 인터페이스로).", "low — 임차 반납뿐."),
    },
    "Context.AlertDeliverer": {
        "inputs": [
            ("`c.Journal` · `c.Entry`", "둘 다 필수", "엔진 조립", "없으면 `ErrRuntimeUnavailable`"),
            ("`c.AccountRef`", "엔진이 기동 때 푼 계정", "`engine.go`", "빈 값이면 실행자는 승격하지 않음(Notifier.escalate 와 같은 규칙)"),
        ],
        "calls": [
            ("`c.clock`", "시계 기본값", "—"),
            ("`blockEntryOnDeliveryStop`", "실행자 정지 → `ReasonAlertSenderDown`(기존)", "OnStop"),
        ],
        "state": ["a124 배선: `Gate: c.Entry` · `AccountRef: c.AccountRef` 두 필드. Notifier 는 넘기지 않음(실행자는 n.mu 와 무관 — D5 · D7)."],
        "safety": ("생산 조립의 유일한 자리(`cmd/tossctl/engine.go:659`). 게이트와 계정을 넘기는 것 외 변화 없음.",
                   "yes — 실행자가 진입 게이트를 잠글 수 있게 된다(보수 방향)."),
    },
    "Journal.settleUnderClaim": {
        "inputs": [
            ("`token`", "비어 있지 않음", "임차", "빈 토큰 → 오류"),
            ("`stmt` · `args`", "id · PENDING · 토큰 CAS 로 끝나는 UPDATE", "세 호출자", "0 행이면 이유를 다시 읽음"),
        ],
        "calls": [
            ("`tx.ExecContext`", "정산 한 문장", "오류 → 롤백"),
            ("`readSettledAttemptsTx`", "**a124**: 적용된 정산의 attempts 를 같은 트랜잭션에서 읽음(D1)", "실패 → 롤백 · 오류(정산 없던 일, Z4)"),
            ("`tx.Commit`", "커밋", "실패 → 오류"),
            ("`explainSettleTx`", "0 행의 이유(LeaseLost · AlreadySettled · NotFound)", "—"),
        ],
        "state": [
            "적용 경로에서 커밋 **전에** 같은 트랜잭션으로 attempts 를 읽어 `SettleResult.Attempts` 에 담는다. 연결이 하나라 `j.db` 로 읽으면 자기 트랜잭션을 기다림.",
            "스키마 무변경. 세 호출자의 SQL · CAS 불변. 적용되지 않은 결과와 오류에서 Attempts 는 0.",
        ],
        "safety": ("가산 필드 + 트랜잭션 안 SELECT 한 줄. 실패하면 정산이 없던 일이 된다(원자성 시험).",
                   "yes — 원장(alert outbox) 정산 경로."),
    },
    "EntryGate.Clear": {
        "inputs": [("`reason`", "사유 코드", "호출자", "없는 래치면 삭제 없음")],
        "calls": [("`g.mu.Lock/Unlock`", "게이트 잠금", "map 연산뿐")],
        "state": [
            "**a124**: 해제 요청마다 `clearEpochs[reason]++` — 래치 유무와 무관(반례 ㉣). 지연 생성.",
            "`revision` 은 그대로 실제 삭제 때만 오른다(전략 봉인 의미 불변).",
        ],
        "safety": ("세대 증가 한 줄 + 지연 생성 두 줄. 잠금 안에서 밖을 부르지 않는다.",
                   "yes — 진입 게이트(High-risk). 해제 동작 자체는 불변."),
    },
    "EntryGate.Block": {
        "inputs": [("`reason` · `detail`", "사유 · 설명", "호출자", "이미 있으면 삽입 없음")],
        "calls": [("`g.mu.Lock/Unlock`", "게이트 잠금", "map 연산뿐")],
        "state": ["없을 때만 삽입하고 그때만 `revision++` — 처음 설명이 남는다(F9). a124 는 이 함수를 편집하지 않았다(위 필드 추가로 줄만 옮겨짐)."],
        "safety": ("대조 근거 — 편집 없음.", "yes — 진입 게이트."),
    },
}


# 편집하지 않은 대조 번들 — FLM 산문은 그대로 두고, 빠진 「Safety conclusion」 절만 더하며 BTM 은 측정으로 다시 만든다.
COMPARE = {
    "internal-app-engine--restorealertentrylatch": ("internal/app/engine/gateway.go", "restoreAlertEntryLatch",
        "a124 는 편집하지 않음 — 기동 때 PENDING 수로 알림 래치를 복원하는 기존 경로. 모드 투영은 복원하지 않음(D10).",
        "yes — 진입 게이트 복원. 대조 근거로만 씀."),
    "internal-obs--notifier.deliver": ("internal/obs/notifier.go", "Notifier.deliver",
        "a124 는 편집하지 않음(a092 표면) — 실행자의 판정 표(D1)가 대응시킨 동기 경로의 기준선.",
        "yes — 동기 알림 경로가 n.mu 아래에서 Gate.Block 을 부르는 기존 결함은 proposal Follow-ups."),
    "internal-obs--notifier.notifycritical": ("internal/obs/notifier.go", "Notifier.notifyCritical",
        "a124 는 편집하지 않음 — 승격(escalate) parity 의 기준선.",
        "yes — 운영 모드 승격 경로. 대조 근거로만 씀."),
}


def refresh_compare(root: Path, cov: dict) -> None:
    for bundle, (source, func, boundary, impact) in COMPARE.items():
        bdir = root / CHANGE / "analysis" / "function-logic" / bundle
        ast = json.loads((bdir / "ast.json").read_text(encoding="utf-8"))
        fl = (bdir / "function-logic-map.md").read_text(encoding="utf-8")
        if "## Safety conclusion" not in fl:
            fl = fl.rstrip("\n") + f"\n\n## Safety conclusion\n\n- Safe edit boundary: {boundary}\n- High-risk impact: {impact}\n"
            (bdir / "function-logic-map.md").write_text(fl, encoding="utf-8")
        src_now = (root / source).read_text(encoding="utf-8")
        cv = cov.get(bundle, {"rows": {}, "package": "?", "tests_run": 0})
        start, end = ast["start"]["line"], ast["end"]["line"]
        bt = [f"# Branch Test Map: `{func}`", "",
              f"- Source: `{source}` ({start}-{end}); current (a124 는 편집하지 않음 — 대조 번들)",
              f"- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `{cv['package']}` 의 시험 {cv['tests_run']} 개를 "
              "하나씩 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험.", "",
              "| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |",
              "|---|---|---|---|---|---|"]
        for b in ast.get("branches") or []:
            r = cv["rows"].get(b["id"], {})
            tests = r.get("entered_by", [])
            txt = line_at(src_now, b["at"]["line"]).replace("|", "\\|")
            if tests:
                shown = " · ".join(f"`{t}`" for t in tests[:3]) + (f" (+{len(tests) - 3})" if len(tests) > 3 else "")
                green = f"블록 {r.get('block')} 을 시험 {len(tests)} 개가 실행, 전부 PASS"
            elif r.get("union"):
                shown, green = "합집합에서만 진입", f"블록 {r.get('block')} — 패키지 합집합에서 실행"
            elif r:
                shown, green = "없음", f"블록 {r.get('block')} 을 어느 시험도 실행하지 않음"
            else:
                shown, green = "미측정", "이번 측정에 없음"
            bt.append(f"| {b['id']} | {b['kind']} at {b['at']['line']}:{b['at']['column']} | `{txt}` | {shown} | 대조(편집 없음) | {green} |")
        (bdir / "branch-test-map.md").write_text("\n".join(bt) + "\n", encoding="utf-8")
        print(f"{bundle}: compare BTM {len(ast.get('branches') or [])} rows")


def git_show(root: Path, rev: str, path: str) -> str:
    return subprocess.run(["git", "show", f"{rev}:{path}"], cwd=root, capture_output=True,
                          text=True, check=True).stdout


def extract(root: Path, source: str, func: str) -> dict:
    env = dict(os.environ, GOFLAGS="-trimpath")
    out = subprocess.run(["go", "run", "./tools/logic-map", "--file", source, "--func", func],
                         cwd=root, capture_output=True, text=True, check=True, env=env).stdout
    return json.loads(out)


def line_at(text: str, n: int) -> str:
    lines = text.splitlines()
    return lines[n - 1].strip() if 0 < n <= len(lines) else ""


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", required=True)
    ap.add_argument("--coverage", action="append", default=[])
    ap.add_argument("--base", required=True)
    args = ap.parse_args()
    root = Path(args.root).resolve()

    cov: dict[str, dict] = {}
    for c in args.coverage:
        data = json.loads(Path(c).read_text(encoding="utf-8"))
        for bundle, rows in data["bundles"].items():
            cov[bundle] = {"rows": rows, "package": data["package"], "tests_run": data["tests_run"],
                           "union_passed": data.get("union_passed")}

    refresh_compare(root, cov)
    for bundle, source, func, edited in TARGETS:
        bdir = root / CHANGE / "analysis" / "function-logic" / bundle
        bdir.mkdir(parents=True, exist_ok=True)
        old_ast = json.loads((bdir / "ast.json").read_text(encoding="utf-8")) if (bdir / "ast.json").exists() else None
        new_ast = extract(root, source, func)
        (bdir / "ast.json").write_text(json.dumps(new_ast, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

        src_now = (root / source).read_text(encoding="utf-8")
        src_base = git_show(root, args.base, source)
        branches = new_ast.get("branches") or []
        returns = [f"{r['at']['line']}:{r['at']['column']}" for r in (new_ast.get("returns") or [])]
        calls = new_ast.get("calls") or []
        start, end = new_ast["start"]["line"], new_ast["end"]["line"]
        prose = PROSE[func]

        # 편집 전 → 편집 뒤 대응(분기 줄의 소스 텍스트를 difflib 로 정렬)
        mapping = []
        if old_ast and edited:
            old_b = old_ast.get("branches") or []
            old_txt = [line_at(src_base, b["at"]["line"]) for b in old_b]
            new_txt = [line_at(src_now, b["at"]["line"]) for b in branches]
            sm = difflib.SequenceMatcher(a=old_txt, b=new_txt, autojunk=False)
            matched = {}
            for block in sm.get_matching_blocks():
                for k in range(block.size):
                    matched[block.a + k] = block.b + k
            for i, b in enumerate(old_b):
                j = matched.get(i)
                mapping.append((b["id"], f"{b['kind']} at {b['at']['line']}:{b['at']['column']}", old_txt[i],
                                branches[j]["id"] if j is not None else "— (없어짐 · 하위 함수로 옮김)"))

        cv = cov.get(bundle, {"rows": {}, "package": "?", "tests_run": 0})
        fl = [f"# Function Logic Map: `{func}`", "",
              f"- Source: `{source}` ({start}-{end})",
              f"- Revision: current — a124 구현 로트(2026-09-27) {'편집 뒤' if edited else '(편집 없음, 줄만 이동)'} 재추출; "
              f"source_sha256 `{new_ast['source_sha256']}`",
              "- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)",
              "- Risk scan: `risk-pattern-report.md`",
              f"- Extractor counts: AST branches {len(branches)} · returns {len(returns)} · calls {len(calls)}",
              f"- Exact AST return positions: {', '.join(returns) if returns else '(none)'}",
              f"- 편집 전 판(base `{args.base[:8]}`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.", "",
              "## Inputs and invariants", "",
              "| Input/state | Valid range | Source of truth | Failure behavior |", "|---|---|---|---|"]
        for row in prose["inputs"]:
            fl.append("| " + " | ".join(row) + " |")
        fl += ["", "## Branches and early returns", "",
               "| Branch | AST anchor | Source text at anchor |", "|---|---|---|"]
        for b in branches:
            txt = line_at(src_now, b["at"]["line"]).replace("|", "\\|")
            fl.append(f"| {b['id']} | {b['kind']} at {b['at']['line']}:{b['at']['column']} | `{txt}` |")
        if not branches:
            fl.append(f"| B1 | branchless happy path at {start}:1 | 본문 전체 |")
        if mapping:
            fl += ["", "### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)", "",
                   "| Base branch | Base anchor | Base source | Current branch |", "|---|---|---|---|"]
            for bid, anchor, txt, cur in mapping:
                fl.append(f"| {bid.replace('B', 'b')} | {anchor.replace(':', '·')} | `{txt.replace('|', '\\|')}` | {cur} |")
        fl += ["", "## Calls and live bindings", "",
               "| Callee | Why called | Error/timeout/retry contract |", "|---|---|---|"]
        for row in prose["calls"]:
            fl.append("| " + " | ".join(row) + " |")
        fl += ["", "## State mutations and fallbacks", ""] + [f"- {s}" for s in prose["state"]]
        fl += ["", "## Safety conclusion", "",
               f"- Safe edit boundary: {prose['safety'][0]}",
               f"- High-risk impact: {prose['safety'][1]}", ""]
        (bdir / "function-logic-map.md").write_text("\n".join(fl), encoding="utf-8")

        bt = [f"# Branch Test Map: `{func}`", "",
              f"- Source: `{source}` ({start}-{end}); current",
              f"- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `{cv['package']}` 의 시험 {cv['tests_run']} 개를 "
              "하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.",
              "", "| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |",
              "|---|---|---|---|---|---|"]
        old_ids_kept = {cur for _, _, _, cur in mapping}
        for b in branches:
            r = cv["rows"].get(b["id"], {})
            tests = r.get("entered_by", [])
            txt = line_at(src_now, b["at"]["line"]).replace("|", "\\|")
            if tests:
                shown = " · ".join(f"`{t}`" for t in tests[:3]) + (f" (+{len(tests) - 3})" if len(tests) > 3 else "")
                green = f"블록 {r.get('block')} 을 시험 {len(tests)} 개가 실행, 전부 PASS"
            elif r.get("union"):
                shown = "합집합에서만 진입(시험별 목록 밖)"
                green = f"블록 {r.get('block')} — 패키지 합집합에서 실행"
            elif r:
                shown = "없음"
                green = f"블록 {r.get('block')} 을 어느 시험도 실행하지 않음"
            else:
                shown = "미측정"
                green = "이 번들은 이번 측정에 없음"
            new_branch = edited and mapping and b["id"] not in old_ids_kept
            red = "a124 RED — 편집 전에 없던 분기" if new_branch else "기존 분기(회귀 핀)"
            bt.append(f"| {b['id']} | {b['kind']} at {b['at']['line']}:{b['at']['column']} | `{txt}` | {shown} | {red} | {green} |")
        if not branches:
            bt.append(f"| B1 | branchless happy path at {start}:1 | 본문 전체 | — | — | — |")
        (bdir / "branch-test-map.md").write_text("\n".join(bt) + "\n", encoding="utf-8")
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", source, "--output",
                        str(bdir / "risk-pattern-report.md")], cwd=root, check=True)
        print(f"{bundle}: {len(branches)} branches, mapping {len(mapping)} rows")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
