#!/usr/bin/env python3
"""a092 25.8 — 26라운드 수리 두 로트(d8769cfb · b910173a) 뒤 FLM/BTM 재작성.

분기 좌표 · 시험 목록 · 블록은 전부 측정값(ast.json · coverage-post-r26b-*.json)에서 읽는다. 저자가 쓰는 것은 분기의
뜻(시나리오)과 반증 변이 ID 뿐이다 — 「이 시험이 이 분기를 탄다」는 손으로 쓰지 않는다.

쓰기: python3 render_r26b_bundles.py (저장소 루트 어디서든 — 경로는 이 파일에서 유도)
"""
from __future__ import annotations

import json
import re
from pathlib import Path

ANALYSIS = Path(__file__).resolve().parents[1]
FL = ANALYSIS / "function-logic"
# 측정 커밋은 패키지별 — obs 는 게이트 준비 로트(15b64676)가 notifier.go 를 다시 만져 재측정, 나머지는 b910173a 뒤 불변.
COMMITS = {"obs": "15b64676", "engine": "b910173a", "execgw": "b910173a"}
COMMIT = "b910173a"  # 머리글 문구의 두 번째 로트 커밋
COV = {
    "obs": "coverage-post-r26d-obs.json",
    "engine": "coverage-post-r26b-engine.json",
    "execgw": "coverage-post-r26b-execgw.json",
}


def load_cov(name: str) -> dict:
    return json.loads((ANALYSIS / "harness" / COV[name]).read_text(encoding="utf-8"))


def old_scenarios(bundle: str) -> dict[str, str]:
    """기존 BTM 의 Scenario 칸(분기 ID 가 안 바뀐 번들만 씀)."""
    out = {}
    path = FL / bundle / "branch-test-map.md"
    for line in path.read_text(encoding="utf-8").splitlines():
        m = re.match(r"^\| (B\d+) \| [^|]+\| ([^|]+)\|", line)
        if m:
            out[m.group(1)] = m.group(2).strip()
    return out


def tests_cell(entered: list[str]) -> str:
    return ", ".join(f"`{t}`" for t in entered[:2]) if entered else "(미실행)"


def green_cell(row: dict, cov: dict) -> str:
    if row["block"] is None:
        return "블록 좌표 없음(조건이 여러 줄이거나 본문이 비어 하네스가 같은 줄 블록을 못 잡음) — 행동 증거는 RED 칸"
    n = len(row["entered_by"])
    if n == 0:
        return "측정 표본의 시험 0개"
    return f"블록 {row['block']}을 시험 {n}개가 실행, PASS"


def render(bundle: str, spec: dict) -> None:
    ast = json.loads((FL / bundle / "ast.json").read_text(encoding="utf-8"))
    cov = load_cov(spec["cov"])
    rows = cov["bundles"][bundle]
    scen = spec.get("scenario") or {}
    if spec.get("keep_scenarios"):
        prev = old_scenarios(bundle)
        scen = {**prev, **scen}
    red = spec.get("red", {})
    fn = ast.get("receiver") and f"{ast['receiver']}.{ast['function']}" or ast["function"]
    start, end = ast["start"]["line"], ast["end"]["line"]
    sha = ast["source_sha256"][:12]
    nb = len(ast.get("branches") or [])
    nr = len(ast.get("returns") or [])
    nc = len(ast.get("calls") or [])
    src = ast["file"]

    btm = [f"# Branch Test Map: `{fn}`", "",
           f"- Source: `{src}` (:{start}-{end}); **편집 뒤** 측정 — `analysis/harness/{COV[spec['cov']]}`"
           f"(연결 워크트리 `{COMMITS[spec['cov']]}`, `{cov['package'].split('tossinvest-cli/')[-1]}` 시험 {cov['tests_run']}개를 하나씩, "
           f"실패 {len(cov['failed_tests'])}). 편집 전 번들은 `{spec['pre']}`에 보존.",
           f"- 재번호: {spec.get('renumber', '없음.')}", "",
           "| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |",
           "|---|---|---|---|---|---|"]
    flm_rows = []
    for br in ast.get("branches") or []:
        bid = br["id"]
        r = rows[bid]
        s = scen.get(bid, "(시나리오 미기재)")
        btm.append(f"| {bid} | {br['kind']} at {br['at']['line']}:{br['at']['column']} | {s} | {tests_cell(r['entered_by'])} | "
                   f"{red.get(bid, '해당 없음(분기 불변)')} | {green_cell(r, cov)} |")
        flm_rows.append(f"| {bid} | {br['kind']} (:{br['at']['line']}) | {s} | — | {tests_cell(r['entered_by'])} |")
    (FL / bundle / "branch-test-map.md").write_text("\n".join(btm) + "\n", encoding="utf-8")

    flm = [f"# Function Logic Map: `{fn}`", "",
           f"- Source: `{src}`",
           f"- AST evidence: `ast.json` — **편집 뒤**, :{start}–{end}, 분기 {nb} · 반환 {nr} · 호출 {nc}, "
           f"source_sha256 `{sha}…`, 추출 커밋 `{COMMITS[spec['cov']]}`. 편집 전 번들은 `{spec['pre']}`에 보존.",
           "- Risk scan: `risk-pattern-report.md`",
           f"- 편집: {spec['edit']}", "",
           "## Inputs and invariants", "",
           "| Input/state | Valid range | Source of truth | Failure behavior |",
           "|---|---|---|---|"] + spec["inputs"] + ["",
           "## Branches and early returns", "",
           "| Branch | Condition | Mutation/side effect | Return/error | Required test |",
           "|---|---|---|---|---|"] + flm_rows + ["",
           "## Calls and live bindings", "",
           "| Callee | Why called | Error/timeout/retry contract | Evidence |",
           "|---|---|---|---|"] + spec["calls"] + ["",
           "## State mutations and fallbacks", ""] + spec["state"] + ["",
           "## Safety conclusion", ""] + spec["safety"]
    (FL / bundle / "function-logic-map.md").write_text("\n".join(flm) + "\n", encoding="utf-8")


PRE = "analysis/pre-edit/r26b/"

SPECS = {
    "internal-app-engine--alertdeliverer.recordfailedattempt": dict(
        cov="engine", pre=PRE + "internal-app-engine--alertdeliverer.recordfailedattempt/",
        renumber="편집 전(e55102f0) B1 → B1, B2 → B4, B3~B7 → B5~B9. 새 B2(반납 행 없음 · 모르는 결과 — d8769cfb) · B3(반납 선점 기록 — b910173a).",
        edit="(d8769cfb, codex P0 · #4) 반납 결과를 받아 행 없음 · 모르는 결과면 원칙 E 조건부 차단(승격 없음), 시도 기록 선점은 "
             "EventAlertClaimLost 로 기록. (b910173a, codex 재확인 R1) 반납 결과가 선점이어도 기록. 시도 한도 판정 · 승격 규칙 불변.",
        inputs=["| `id` · `token` | 이 실행자가 claim 한 행과 임차 | 배달 실행자 사이클 | 아래 분기 |",
                "| 반납 결과 | Applied · AlreadySettled · LeaseLost · NotFound · 모르는 값 | `ReleaseAlertClaim` | B2 · B3 |"],
        scenario={
            "B1": "시도 기록 오류 → 로그",
            "B2": "반납이 행 없음 · 모르는 결과 → 로그 + 조건부 차단(승격 없음)",
            "B3": "반납이 선점(승인 · 남의 임차) → 선점 기록",
            "B4": "시도 기록 오류 → 행별 연속 기록 실패 계수, 반환",
            "B5": "시도 기록 결과 분기",
            "B6": "Applied → 연속 기록 실패 끝",
            "B7": "한도 미만 → 반환",
            "B8": "시도 기록 선점 → 선점 기록",
            "B9": "행 없음 · 모르는 결과 → 조건부 차단(승격 없음)",
        },
        red={"B2": "X01 · X02 · X03 · X05 CAUGHT(`mutation-r26/ledger.tsv`)", "B3": "Z01 CAUGHT(`ledger-r26b.tsv`)",
             "B8": "X04 CAUGHT", "B7": "X05 CAUGHT(한도 판정 보존)"},
        calls=["| `MarkAlertAttemptFailed` | 시도 기록 | 오류면 결과를 읽지 않음(F10) | AST |",
               "| `release` → `ReleaseAlertClaim` | 반납 | (결과, ok) — 오류면 ok=false | AST |",
               "| `judge` · `readEpoch` | 원칙 E | 근거 확정 뒤 세대 | AST |",
               "| `logf(EventAlertClaimLost)` | 선점 기록 | 판정 없음 | AST |"],
        state=["- 원장 정산 · 반납 · 게이트 래치(judge) · 로그."],
        safety=["- Safe edit boundary: 반납 결과 분류와 선점 기록만 — 시도 한도 판정 · 승격 규칙 불변(X05 CAUGHT).",
                "- High-risk impact: yes — 배달 실행자의 진입 차단 판정(a124 정본 영역, a092 델타 「모든 발송자」)."],
    ),
    "internal-app-engine--alertdeliverer.release": dict(
        cov="engine", pre=PRE + "internal-app-engine--alertdeliverer.release/",
        edit="(d8769cfb, codex P0) 반납 결과 `(journal.SettleResult, bool)` 를 돌려줌 — 오류면 로그 후 `(SettleResult{}, false)`. 분류는 호출자.",
        inputs=["| `id` · `token` | 이 실행자가 claim 한 행과 임차 | 배달 실행자 사이클 | B1 |"],
        scenario={"B1": "반납 오류 → 로그, ok=false"},
        red={"B1": "해당 없음(분기 불변 — 반환값만 추가, 호출자 분류는 X01)"},
        calls=["| `ReleaseAlertClaim` | 반납 | 떨어진 ctx + 상한(`alertReleaseTimeout`) | AST |"],
        state=["- 원장 반납 하나."],
        safety=["- Safe edit boundary: 반환값 추가만 — 분기 · 기한 불변.",
                "- High-risk impact: yes — 결과가 배달 실행자의 차단 판정 입력이 됨."],
    ),
    "internal-obs--notifier.deliver": dict(
        cov="obs", pre=PRE + "internal-obs--notifier.deliver/", keep_scenarios=True,
        edit="(d8769cfb, codex #6) 시도 기록의 행 없음 갈래 조건 `failed.Outcome == SettleNotFound` → `!isPreemption(failed.Outcome)`, "
             "반납 갈래 조건 → `!isPreemption(released.Outcome)` — 한 판정(모르는 결과도 선점 아님). 분기 수 · 순서 불변(27).",
        inputs=["| `id` · `token` | claim 한 행과 임차 | `claimAndDeliver`(잠금 안 claim) | 토큰이 안 맞으면 정산 거절 |",
                "| `n.mu` | **쥐지 않음** | — | 배제는 임차 |",
                "| 해제 세대 | 근거 확정 직후 `readVerdict` 가 읽음 | `EntryGate.ClearEpoch` | 원칙 E |"],
        scenario={"B17": "**시도 기록이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 원칙 E 조건부 차단**",
                  "B24": "**반납이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 로그**"},
        red={"B17": "X06 CAUGHT(`isPreemption` 확대)", "B24": "X06 CAUGHT"},
        calls=["| `isPreemption` | 선점 판정 한 곳 | AlreadySettled · LeaseLost 만 참 | AST · `TestA092DeliverClassifiesThroughOneJudgement` |",
               "| `MarkAlertAttemptFailed` · `ReleaseAlertClaim` · `MarkAlertDelivered` | 정산 | 결과 분류 | AST |",
               "| `readVerdict` · `BlockUnlessClearedSince` | 원칙 E | — | AST |"],
        state=["- 원장 정산 · 게이트 조건부 래치 · 로그."],
        safety=["- Safe edit boundary: 두 조건식만 — 분기 구조 · 래치 순서 불변.",
                "- High-risk impact: yes — 동기 critical 발송의 진입 차단 판정."],
    ),
    "internal-obs--notifier.notifycritical": dict(
        cov="obs", pre=PRE + "internal-obs--notifier.notifycritical/", keep_scenarios=True,
        edit="(b910173a, 보이스 A#4 · B#4) 원장 없음 경고 줄의 원래 사건 유형 키 `FieldEvent` → `FieldTriggerEvent`(줄 자신의 event 를 가리지 않음). 분기 불변.",
        inputs=["| `n.Journal` | 없으면 B1 | 조립 | 경고 + 최선 발행 |"],
        red={"B2": "전용 변이 없음 — `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey` 가 이 줄의 event 키 수를 셈(편집 전 FAIL · `red-r26b-obs.log`)"},
        calls=["| `n.claimAndDeliver` | 기록 · 발송 | (sent, owed, verdict, err) | AST |", "| `n.judge` | 원칙 E | — | AST |"],
        state=["- 로그 키 하나만 바뀜."],
        safety=["- Safe edit boundary: 로그 키만.", "- High-risk impact: no(관측 줄) — 판정 불변."],
    ),
    "internal-obs--notifier.escalate": dict(
        cov="obs", pre=PRE + "internal-obs--notifier.escalate/", keep_scenarios=True,
        edit="(b910173a, 보이스 A#4 · B#4) 승격 실패 줄의 `FieldEvent` → `FieldTriggerEvent`. (15b64676, 게이트 준비 gstack 리뷰) 같은 줄의 error 칸을 "
             "`MaskAccount(err, n.AccountRef)` 로 — 원장 문구의 계좌 원문을 가림(불변식 8 (a), judge 게이트 설명 고정 문구의 짝). 분기 · 반환 불변. "
             "(같은 줄의 `FieldAccount` 는 base 관행 — 불변식 8 (b) 사람 결정 큐.)",
        inputs=["| `n.Journal` · `n.AccountRef` | 둘 다 있어야 승격 포함 | 조립 | B1 |"],
        red={"B3": "W02 CAUGHT(`TestA092TheEscalationFailureErrorMasksTheAccount` — error 칸 가림). 키 이름(`trigger_event`)은 행동 시험이 세지 않음(로그 키, 판정 불변 — 비례 원칙)"},
        calls=["| `EscalateOperatingMode(…, nil)` | 승격 · 통지 없음 | 오류 반환 | AST |"],
        state=["- 모드 행 하나(변화 시)."],
        safety=["- Safe edit boundary: 로그 키만.", "- High-risk impact: yes(모드 승격) — 이 편집은 판정 불변."],
    ),
    "internal-obs--notifier.logleaselost": dict(
        cov="obs", pre=PRE + "internal-obs--notifier.logleaselost.json(AST)",
        renumber="새 B6(`case res.Outcome != journal.SettleLeaseLost` — 모르는 결과), 편집 전 B6(default — 남의 임차) → B7.",
        edit="(b910173a, codex 재확인 R3) 결과 enum 으로 분류 — 모르는 결과는 `EventAlertUndelivered` 오류(선점 이름 금지). 기록만, 판정 없음.",
        inputs=["| `res.Outcome` | Applied 외 전부(호출자가 Applied 를 먼저 거름) | 원장 | B3~B7 |"],
        scenario={"B1": "로거 없음 → 반환", "B2": "결과 분기", "B3": "행 없음 → 오류 줄", "B4": "이미 정산 → 선점 기록",
                  "B5": "LeaseLost · 토큰 없음 → 자기 반납 재확인", "B6": "모르는 결과 → 원장 이상 오류 줄",
                  "B7": "LeaseLost · 남의 토큰 → 선점 경고"},
        red={"B6": "Z02 CAUGHT(`TestA092TheLeaseLossLogClassifiesByOutcome/unknown`)"},
        calls=["| `Log.Error` · `Log.Event` · `Log.Warn` | 기록 | 없음 | AST |"],
        state=["- 로그 줄 하나."],
        safety=["- Safe edit boundary: 로그 분류만 — 차단은 호출자가 이미 판정.", "- High-risk impact: no(관측)."],
    ),
    "internal-obs--notifier.publishbesteffort": dict(
        cov="obs", pre=PRE + "internal-obs--notifier.publishbesteffort.json(AST)",
        edit="(b910173a, 보이스 A#4 · B#4) 발행 실패 줄의 원래 사건 유형 · 등급 키를 `trigger_event` · `trigger_severity` 로 — emit 이 쓰는 줄 자신의 event · severity 를 가리지 않음. 분기 불변.",
        inputs=["| `n.Publisher` | nil 이면 B1 | 조립 | 반환 |"],
        scenario={"B1": "발행기 없음 → 반환", "B2": "발행 실패 → 경고 줄"},
        red={"B2": "Z13 CAUGHT"},
        calls=["| `Publisher.Publish` | 최선 발송 | 오류는 로그만 | AST |"],
        state=["- 로그 줄 하나."],
        safety=["- Safe edit boundary: 로그 키만.", "- High-risk impact: no(일반 등급 관측)."],
    ),
    "internal-app-engine--runtime.escalate": dict(
        cov="engine", pre=PRE + "internal-app-engine--runtime.escalate.json(AST)",
        renumber="새 B2(`ErrModeAnnouncementFailed` 갈래), 편집 전 B2(승격 오류) → B3.",
        edit="(b910173a, 보이스 B#7 잔재) 지속 실패 승격이 커밋되고 통지 기록만 실패하면 「승격됨 · 통지 기록 실패」로 기록 — 「재시작이 푸는 차단」 오보 제거. 승격 호출 · 알림 불변.",
        inputs=["| `r.opts.Escalate` · `AccountRef` | 둘 다 있어야 승격 | 조립 | B1 |"],
        scenario={"B1": "승격기 · 계좌 없음 → 반환", "B2": "커밋됨 · 통지 기록 실패 → 「승격됨」 경고", "B3": "승격 오류 → 「재시작이 푼다」 경고"},
        red={"B2": "Z18 CAUGHT(`TestA092AnUnannouncedSustainedTighteningIsReportedAsTightened`)"},
        calls=["| `Escalate.EscalateOperatingMode` | 지속 실패 → ENTRY_BLOCKED | 오류 분류 | AST |"],
        state=["- 모드 행 하나(변화 시) · 로그."],
        safety=["- Safe edit boundary: 로그 갈래 하나 — 승격 호출 · 루프 재시도 불변.", "- High-risk impact: yes(감독자 승격) — 판정 불변."],
    ),
    "internal-execgw--riskguardian.escalatefor": dict(
        cov="execgw", pre=PRE + "internal-execgw--riskguardian.escalatefor.json(AST)",
        renumber="새 B3(`ErrModeAnnouncementFailed` 갈래, B2 안).",
        edit="(b910173a, 보이스 B#7 잔재) 일일 손실 승격이 커밋되고 통지 기록만 실패하면 「승격됨 · 통지 기록 실패」 오류로 — 「재시작이 푸는 차단」 오보 제거. `errors.Is` 로 원 오류 보존. 진입 거절은 호출자가 이미 판정(불변).",
        inputs=["| `verdict.Reason` | DAILY_LOSS_LIMIT_REACHED 만 승격 | 체인 | B1 |"],
        scenario={"B1": "다른 거절 → 승격 없음", "B2": "승격 오류", "B3": "커밋됨 · 통지 기록 실패 → 「승격됨」 오류"},
        red={"B3": "Z19 CAUGHT(`TestA092AnUnannouncedDailyLossTighteningIsReportedAsTightened`)"},
        calls=["| `journal.EscalateOperatingMode` | 일일 손실 → ENTRY_BLOCKED | 오류는 거절에 덧붙음(errors.Join) | AST |"],
        state=["- 모드 행 하나(변화 시)."],
        safety=["- Safe edit boundary: 오류 문구 갈래 하나 — 거절 판정 · 승격 호출 불변(호출자 `errors.Join(chainRefusal, …)`).",
                "- High-risk impact: yes(Guardian) — 판정 불변, 진입은 어느 갈래든 거절."],
    ),
}

HEADER_ONLY = ["internal-app-engine--alertdeliverer.deliverone", "internal-obs--notifier.claimanddeliver",
               "internal-obs--notifier.logclaimheld"]


def refresh_header(bundle: str) -> None:
    ast = json.loads((FL / bundle / "ast.json").read_text(encoding="utf-8"))
    p = FL / bundle / "function-logic-map.md"
    lines = p.read_text(encoding="utf-8").split("\n")
    i = next(k for k, l in enumerate(lines) if l.startswith("- AST evidence:"))
    l = re.sub(r"source_sha256 `[0-9a-f]{12}…`", f"source_sha256 `{ast['source_sha256'][:12]}…`", lines[i])
    l = re.sub(r":\d+–\d+", f":{ast['start']['line']}–{ast['end']['line']}", l, count=1)
    l = re.sub(r" 26라운드 수리(\(`d8769cfb`\)| 두 로트\(`d8769cfb` · `[0-9a-f]+`\)) 뒤 재추출.*$", "", l)
    last = "15b64676" if "internal-obs--" in bundle else COMMIT
    l += (f" 26라운드 수리 두 로트(`d8769cfb` · `{COMMIT}`) 뒤 재추출(마지막 `{last}`) — 이 함수 본문 · 분기 수 불변(같은 파일의 다른 "
          "함수 편집으로 줄 이동 · 파일 해시만 바뀜, 분기 좌표는 `ast.json` 이 정본).")
    lines[i] = l
    p.write_text("\n".join(lines), encoding="utf-8")
    # BTM 앵커 · 범위를 ast.json 좌표로 다시 씀(분기 종류가 같아야 함 — 본문 불변의 확인)
    at = {x["id"]: (x["kind"], x["at"]["line"], x["at"]["column"]) for x in ast.get("branches") or []}
    b = FL / bundle / "branch-test-map.md"
    out = []
    for line in b.read_text(encoding="utf-8").split("\n"):
        m = re.match(r"^\| (B\d+) \| (\w+) at (\d+):(\d+) \|", line)
        if m:
            k, ln, col = at[m.group(1)]
            assert k == m.group(2), (bundle, m.group(1), k, m.group(2))
            line = f"| {m.group(1)} | {k} at {ln}:{col} |" + line[m.end():]
        out.append(line)
    text = re.sub(r"\(:(\d+)-(\d+)\)", f"(:{ast['start']['line']}-{ast['end']['line']})", "\n".join(out), count=1)
    b.write_text(text, encoding="utf-8")


if __name__ == "__main__":
    for b, spec in SPECS.items():
        render(b, spec)
    for b in HEADER_ONLY:
        refresh_header(b)
    print(f"rendered {len(SPECS)} · header-refreshed {len(HEADER_ONLY)}")
