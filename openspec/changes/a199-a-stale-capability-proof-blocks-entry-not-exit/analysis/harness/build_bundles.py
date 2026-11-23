#!/usr/bin/env python3
"""a199 1.9 — 편집 전(pre-edit) FLM/BTM 번들 생성기.

분기 표는 손으로 읽지 않고 `tools/logic-map` AST 를 열거해 만든다(분기 id · 종류 · 줄 · 그 줄의 소스 문장).
GREEN observed 칸은 base 의 선택 실행 coverprofile(`analysis/implementation/base-cover-*.out`, set 모드)에서
분기 줄에서 시작하는 블록의 count 를 읽는다 — 시험별 귀속이 아니라 선택 집합 전체의 실측이다(시험별 귀속은 2단계 RED).
사람이 쓴 것은 아래 FUNCS 의 의도 · 불변식 · a199 편집 칸뿐이다.

사용(저장소 루트): python3 <이 파일> --bin <logic-map 바이너리>
"""

from __future__ import annotations

import argparse
import json
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
ANALYSIS = HERE.parent
CHANGE = ANALYSIS.parent.name
IMPL = ANALYSIS / "implementation"

ENGINE_SEL = "`analysis/implementation/base-selected-engine.txt`(rc 0, `-run` 목록은 그 파일 머리가 아니라 review.md 1.9 에)"

FUNCS = [
    dict(file="internal/app/engine/interlock.go", func="verifyGate", cover="base-cover-engine.out",
         purpose="게이트 ON 기동의 조항 1~8 검사. 첫 실패에서 반환(AST 11 갈래가 전부 오류 반환).",
         inputs=[("`status.Limits`·`AttestationFile`·`AccountRef`", "newAutomationStatus + NewContext 가 채움", "NewContext"),
                 ("`facts`(guardian · trading · gateway · transport · now)", "NewContext 가 조립", "gateFacts"),
                 ("불변식", "하나라도 실패하면 `ErrAutomationGateRefused` 로 감싼 오류(조항 sentinel `%w`)", "engine-safety 「자동화 게이트 기동 인터록」")],
         mutation="`status.AttestationExpiresAt`(524) 한 칸만 씀.",
         a199="B6@536 의 낡음(만료 · endpoint 누락)을 반환하지 않고 표지로 남긴 채 B7~B11 을 끝까지 검사(D2), 만료일 때 계좌를 따로 확인(D1a). 첫 실패 반환 구조를 바꾸는 편집 — claim-check C2.",
         tests="`TestGateOnRefusals` · `TestNothingElseRefusesTheOperatorConfiguration` · `TestRefusedStartIsAudited`"),
    dict(file="internal/app/engine/interlock.go", func="runInterlock", cover="base-cover-engine.out",
         purpose="게이트 OFF 면 즉시 반환, ON 이면 verifyGate → 거절(감사 · 로그) 또는 수락(Verified · EntryPermitted · 감사 · 로그).",
         inputs=[("`gate.Enabled`", "설정", "config"), ("`log`·`logger`", "NewContext", "감사 · 구조 로그"),
                 ("불변식", "수락 감사 기록 실패면 기동 거절(B3)", "§0.5")],
         mutation="`status.Verified`(406) · `status.EntryPermitted`(409) 씀, 감사 1줄.",
         a199="세 번째 결과(청산 전용): Verified=false 유지 · 새 표지 · EntryPermitted=false · 새 감사 action · 진입 래치(D3·D4·D5·D8).",
         tests="`TestGateOnRefusals` · `TestGateOffStartsAndDoesNoGateWork` · `TestGateDecisionIsLoggedStructurally`"),
    dict(file="internal/app/engine/interlock.go", func="logGateDecision", cover="base-cover-engine.out",
         purpose="기동마다 구조 로그 한 줄(`engine.operating_mode`) — 거절이면 Warn + reason.",
         inputs=[("`status`", "runInterlock", "AutomationStatus"), ("`refusal`", "verifyGate 오류 또는 nil", "—")],
         mutation="로그 한 줄.",
         a199="(조건부 — Q) B1@479 의 「보호는 프로세스 수명」 문장을 청산 전용 기동에도, 새 표지 필드 하나.",
         tests="`TestGateDecisionIsLoggedStructurally`"),
    dict(file="internal/app/engine/engine.go", func="NewContext", cover="base-cover-engine.out",
         purpose="엔진 프로필 조립: 계좌 → journal → 게이트웨이 → Guardian → 인터록 → Context.",
         inputs=[("`opts`", "assembleEngine · 시험", "Options"), ("불변식", "인터록이 거절하면 Context 없음(B15@588)", "engine-safety 「엔진 프로필은 다음 순서로 구성한다」"),
                 ("a166 census", "본문 전체가 `shapes.json` 키로 동결", "internal/accountbudgetruntime/testdata/a166-wiring")],
         mutation="journal 열기 · 감사 · 게이트웨이 조립(기동 복원 래치) · Guardian 생성.",
         a199="B16@592 Guardian 유지 조건을 청산 쪽 표지로(D3). 래치를 runInterlock 에 두면 `gateFacts` 에 진입 게이트 한 칸. census 키 1 갱신.",
         tests="`TestGateOnRefusals` · `TestNothingElseRefusesTheOperatorConfiguration` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver`"),
    dict(file="cmd/tossctl/engine.go", func="runEngineRun", cover="base-cover-cmd.out",
         purpose="`tossctl engine run` — 잠금 → 조립/인터록 → 게이트 OFF 거절 → verify 잠금 → 마커 → 루프 → 실행.",
         inputs=[("`ectx.Automation.Verified`", "NewContext", "B7@232"), ("불변식", "defer 등록 순서(a109 정적 핀)", "a109_the_lock_outlives_every_endpoint_close_test.go")],
         mutation="잠금 · 마커 · endpoint 기동 · 출력.",
         a199="B7@232 를 청산 쪽 표지로 + 청산 전용 기동 출력 한 줄(D8). 새 defer 없음.",
         tests="`TestAGateOffEngineRefusesWithoutEnumeratingClauses` · `TestAnUnmetInterlockIsEnumerated`"),
    dict(file="internal/app/engine/exitwiring.go", func="Context.ExitObserver", cover="base-cover-engine.out",
         purpose="exit 관측 루프 조립 — 미검증 게이트면 거절, Guardian 이 `ReductionIssuer` 여야 함.",
         inputs=[("`c.Automation.Verified`", "NewContext", "B2@323"), ("`c.Guardian`", "NewContext B16", "B3@327")],
         mutation="opts 덮어쓰기(Journal · Prices · Retrier · Issuer · Submit …).",
         a199="B2@323 를 청산 쪽 표지로(D3). a198 이 같은 함수를 편집(Guardian 을 CancelIssuer 로 단언) — a198 착지 뒤 순차.",
         tests="`TestA092ExitObserverGetsRecordOnlyAlertPaths` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver`"),
    dict(file="internal/app/engine/reconcileloop.go", func="Context.ReconcileDriver", cover="base-cover-engine.out",
         purpose="대사 + 편입 루프 조립 — 미검증 게이트면 거절.",
         inputs=[("`c.Automation.Verified`", "NewContext", "B2@368"), ("`opts.Prices`", "호출자 또는 c.Official", "B5@385 poll 결속")],
         mutation="opts 덮어쓰기.",
         a199="B2@368 를 청산 쪽 표지로(D3).",
         tests="`TestTheDriverRefusesAnUnverifiedGate`"),
    dict(file="internal/app/engine/attestation_watch.go", func="judgeAttestation", cover="base-cover-engine.out",
         purpose="감시의 판정 — Load 오류 → unreadable, Verify 첫 실패 → 단계, 통과면 기한까지 날수.",
         inputs=[("`att`·`loadErr`", "attest.Load", "B1@101"), ("`accountRef`", "엔진 자격 증명의 계좌", "Verify")],
         mutation="없음(순수).",
         a199="어느 결함이 청산 전용 / 거부인지를 인터록과 같은 함수로(D7) — 지금은 만료 + 다른 계좌를 expired 로 판정(B4@108), 인터록(D1a)은 거부.",
         tests="`TestA135StageBoundaries` · `TestA135ScopeLossIsCritical` · `TestA135UnusableAttestationsAreCritical` · `TestA196DeadlineBoundaries`"),
    dict(file="internal/app/engine/attestation_watch.go", func="attestationWatch.eventFor", cover="base-cover-engine.out",
         purpose="판정을 알림 하나로 — 제목 · 본문 한국어 고정 문구, 계좌 · 경로 · 원문 오류 없음.",
         inputs=[("`v`", "judgeAttestation + prepare", "단계"), ("`w.critical`", "notifications.enabled", "B1@385")],
         mutation="없음(문구 조립).",
         a199="expired · scope_lost 의 귀결 문장(「청산 전용으로 뜬다」), 청산 전용으로 돌고 있으면 현재 사실(D7). 날수 · 지남 단계 문장은 Q.",
         tests="`TestA196RefusalTellsTheConsequenceFirst` · `TestA196ScopeLossWordingFollowsTheMissingKind` · `TestA196ImminentProofNoticeTellsTheProofDeadline`"),
    dict(file="internal/app/engine/attestation_watch.go", func="newAttestationWatch", cover="base-cover-engine.out",
         purpose="엔진 Context 에서 감시를 만듦(게이트 · 파일 · 계좌 · 알림기).",
         inputs=[("`c.Automation`", "NewContext(기동 시점 값)", "227-228")],
         mutation="감시 구조체 채움.",
         a199="기동 때의 청산 전용 표지를 감시에 옮김 — 「지금 청산 전용으로 돌고 있다」의 출처(D6 로 프로세스 수명 동안 불변).",
         tests="`TestA135ProductionExecutorShape`"),
    dict(file="internal/execgw/failclosed.go", func="AllReasonCodes", cover="base-cover-execgw.out",
         purpose="이 빌드가 낼 수 있는 사유 코드 전부(정렬) — golden 과 a098 길이 단언의 입력.",
         inputs=[("없음", "—", "—")],
         mutation="없음.",
         a199="새 사유 코드 한 줄 등록(D4). golden · a098 등록 변수 갱신.",
         tests="`TestReasonCodeEnumIsStable` · `TestTheSenderDownReasonIsRegisteredInTheEnumeration`"),
    # Manager 결정(2026-11-24)으로 더한 편집 대상
    dict(file="internal/app/engine/attestation_watch.go", func="attestationWatch.prepare", cover="base-cover-engine.out",
         purpose="읽기 · 판정 · 중복 확인 · 문구 — healthy 면 알림 없이 반환(B8@345).",
         inputs=[("`w.load`·`w.clock`·`w.accountRef`", "newAttestationWatch", "판정 입력"), ("`w.sent`·`w.episode`", "감시 메모리", "하루 한 번 · 에피소드")],
         mutation="`w.identity`·`w.sent`·`w.episode`·`w.episodeAt`.",
         a199="결정 7: 기동 표지가 청산 전용이면 healthy 여도 하루 한 번 알림 — 조건을 파일이 아니라 기동 표지에.",
         tests="`TestA135WatchAnnouncesEachStageOnce` · `TestA196ScopeLossOncePerDayAcrossReissues`"),
    dict(file="cmd/tossctl/soak.go", func="runSoakAttest", cover=None,
         purpose="`tossctl soak attest` — attestation 발급 · 저장 · 범위 상실 안내 출력.",
         inputs=[("발급 결과 · 빠진 endpoint", "soak.BuildAttestation", "출력 문구")],
         mutation="attestation 파일 저장 · 출력.",
         a199="결정 11: 565 「The automation gate will refuse to start」 · 577 「clause 9 … still refuse」 문구를 새 귀결로(문구만).",
         tests="선택 실행 미실시(2단계)"),
    dict(file="cmd/tossctl/soak.go", func="newSoakAttestCmd", cover=None,
         purpose="`soak attest` 명령 정의 — Long 도움말.",
         inputs=[("도움말 문자열", "상수", "220")],
         mutation="없음.",
         a199="결정 11: 도움말 「the gate will still refuse」 문구를 새 귀결로.",
         tests="선택 실행 미실시(2단계)"),
    # 뒤집히는 기존 시험 함수
    dict(file="internal/app/engine/interlock_test.go", func="TestGateOnRefusals", cover=None,
         purpose="게이트 ON 거절 사례 표(조항별 sentinel · 문구).",
         inputs=[("사례 표", "시험 안", "342-556")],
         mutation="임시 디렉터리 · httptest.",
         a199="낡음 사례 셋(「attestation expired」·「misses an endpoint the engine calls」·「predates the price read」)이 거절 기대에서 청산 전용 기대로.",
         tests="`TestGateOnRefusals`"),
    dict(file="internal/app/engine/interlock_test.go", func="TestRefusedStartIsAudited", cover=None,
         purpose="거절된 기동의 감사 1줄.",
         inputs=[("픽스처", "만료 attestation(653)", "—")],
         mutation="임시 디렉터리.",
         a199="픽스처를 거절이 유지되는 결함으로 교체(만료는 청산 전용이 됨).",
         tests="`TestRefusedStartIsAudited`"),
    dict(file="internal/app/engine/interlock_test.go", func="TestGateDecisionIsLoggedStructurally", cover=None,
         purpose="수락 · 거절 · 게이트 OFF 의 구조 로그 한 줄.",
         inputs=[("하위 「refused」 픽스처", "만료 attestation(804)", "—")],
         mutation="임시 디렉터리.",
         a199="「refused」 픽스처 교체 + (조건부) 청산 전용 하위 시험.",
         tests="`TestGateDecisionIsLoggedStructurally`"),
    dict(file="internal/app/engine/runtime_wiring_test.go", func="TestEachInterlockClauseHasALine", cover=None,
         purpose="거절 조항마다 열거 한 줄.",
         inputs=[("「an expired attestation」 사례", "95", "—")],
         mutation="임시 디렉터리.",
         a199="만료 사례를 만료 + 다른 계좌(D1a) 등 거절 유지 모양으로.",
         tests="`TestEachInterlockClauseHasALine`"),
    dict(file="internal/app/engine/a196_attestation_deadline_internal_test.go", func="TestA196RefusalTellsTheConsequenceFirst", cover=None,
         purpose="거절 단계 알림의 제목 · 본문 첫 문장 = 「기동이 거부된다」.",
         inputs=[("사례 다섯", "scope_lost · expired · incomplete · account_mismatch · unreadable", "—")],
         mutation="없음.",
         a199="scope_lost · expired 의 기대 문장을 청산 전용 귀결로, 나머지 셋 무변.",
         tests="`TestA196RefusalTellsTheConsequenceFirst`"),
    dict(file="internal/app/engine/a196_attestation_deadline_internal_test.go", func="TestA196ScopeLossWordingFollowsTheMissingKind", cover=None,
         purpose="범위 상실 문장의 원인 · 할 일이 누락 종류를 따름.",
         inputs=[("사례 넷", "POST 둘 · POST 하나 · GET 하나 · POST+GET", "—")],
         mutation="없음.",
         a199="접두 단언(「기동이 거부된다」)을 청산 전용 귀결로.",
         tests="`TestA196ScopeLossWordingFollowsTheMissingKind`"),
]


def run(cmd: list[str]) -> str:
    return subprocess.run(cmd, check=True, capture_output=True, text=True).stdout


def bundle_dir(src: str, func: str) -> Path:
    return Path(run(["python3", "tools/logic-map/scaffold_analysis.py", "--change", CHANGE, "--file", src, "--func", func]).strip())


def cover_blocks(profile: str | None, src: str) -> list[tuple[int, int, int, str]]:
    if not profile:
        return []
    out = []
    for line in (IMPL / profile).read_text(encoding="utf-8").splitlines():
        if line.startswith("mode:"):
            continue
        loc, _n, count = line.rsplit(" ", 2)
        path, span = loc.rsplit(":", 1)
        if not path.endswith("/" + src) and not path.endswith(src):
            continue
        a, _b = span.split(",")
        out.append((int(a.split(".")[0]), int(a.split(".")[1]), int(count), span))
    return out


def body_count(blocks, line: int, col: int) -> str:
    # 분기 줄에서 시작하고 분기 열보다 뒤에서 여는 블록 = 그 분기의 몸통(if 본문 · case 본문)
    cands = [b for b in blocks if b[0] == line and b[1] > col]
    if not cands:
        cands = [b for b in blocks if b[0] == line + 1 or (b[0] == line)]
    if not cands:
        return "미측정(블록 없음)"
    b = min(cands, key=lambda x: x[1])
    return f"{'yes' if b[2] > 0 else 'no'} — {b[3]} count {b[2]}"


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--bin", required=True)
    args = ap.parse_args()
    for f in FUNCS:
        src, func = f["file"], f["func"]
        out = bundle_dir(src, func)
        ast_text = run([args.bin, "--file", src, "--func", func])
        (out / "ast.json").write_text(ast_text, encoding="utf-8")
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", src, "--output", str(out / "risk-pattern-report.md")], check=True)
        ast = json.loads(ast_text)
        lines = Path(src).read_text(encoding="utf-8").splitlines()
        branches = ast.get("branches") or []
        returns = ast.get("returns") or []
        calls = ast.get("calls") or []
        name = (ast.get("receiver") + "." if ast.get("receiver") else "") + ast["function"]
        start, end = ast["start"]["line"], ast["end"]["line"]
        blocks = cover_blocks(f["cover"], src)
        is_test = src.endswith("_test.go")

        brow = []
        for b in branches:
            ln = b["at"]["line"]
            text = lines[ln - 1].strip().replace("|", "\\|")
            if len(text) > 110:
                text = text[:107] + "…"
            brow.append(f"| {b['id']} ({b['kind']}@{ln}) | `{text}` | — | AST 열거 | 2단계 RED 에서 이름 확정 |")
        if not brow:
            brow.append(f"| (분기 없음) | 한 갈래(happy path {start}-{end}) | — | — | 아래 BTM B1 |")

        flm = [
            f"# Function Logic Map: `{name}`",
            "",
            f"> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.",
            "",
            f"- Source: `{src}` ({start}-{end})",
            f"- AST evidence: `ast.json` — branches **{len(branches)}** / returns **{len(returns)}** / calls **{len(calls)}**, source SHA-256 `{ast['source_sha256']}`(파일 단위).",
            f"  명령 `go run ./tools/logic-map --file {src} --func {func}`.",
            "- Risk scan: `risk-pattern-report.md`.",
            f"- 의도: {f['purpose']}",
            "",
            "## Inputs and invariants",
            "",
            "| Input/state | Valid range | Source of truth | Failure behavior |",
            "|---|---|---|---|",
        ]
        for a, b_, c in f["inputs"]:
            flm.append(f"| {a} | {b_} | {c} | 아래 분기 표 |")
        flm += [
            "",
            "## Branches and early returns",
            "",
            "| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |",
            "|---|---|---|---|---|",
            *brow,
            "",
            "## Calls and live bindings",
            "",
            "| Callee | Why called | Error/timeout/retry contract | Evidence |",
            "|---|---|---|---|",
            f"| AST calls {len(calls)}개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |",
            "",
            "## State mutations and fallbacks",
            "",
            f"- {f['mutation']}",
            "",
            "## Safety conclusion",
            "",
            f"- a199 편집(예정): {f['a199']}",
            "- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.",
            f"- High-risk impact: {'no(시험 함수)' if is_test else 'yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)'}",
            "",
        ]
        (out / "function-logic-map.md").write_text("\n".join(flm), encoding="utf-8")

        if is_test:
            head = f"- 편집 전. 이 시험은 base 에서 rc 0({ENGINE_SEL}). 시험 함수라 커버 대상이 아님 — 분기는 시험 자신의 사례 · 단언 갈래."
        else:
            head = f"- 편집 전. 커버: `../../implementation/{f['cover']}`(set 모드, 선택 실행 rc 0). 칸의 count 는 선택 집합 전체 실측 — 시험별 귀속 아님. 대표 시험: {f['tests']}."
        btm = [f"# Branch Test Map: `{name}`", "", head, "",
               "| Branch | Scenario | Test | RED observed | GREEN observed |",
               "|---|---|---|---|---|"]
        for b in branches:
            ln = b["at"]["line"]
            green = "yes — 시험 rc 0(base)" if is_test else body_count(blocks, ln, b["at"]["column"])
            btm.append(f"| {b['id']} | {b['kind']}@{ln} | 선택 집합(머리) | 해당 없음(기존 분기) | {green} |")
        if not branches:
            green = "yes — 시험 rc 0(base)" if is_test else body_count(blocks, start, 0)
            btm.append(f"| B1 | branchless happy path {start}-{end} | 선택 집합(머리) | 해당 없음 | {green} |")
        (out / "branch-test-map.md").write_text("\n".join(btm) + "\n", encoding="utf-8")
        print(f"{out.name}: branches {len(branches)}")


if __name__ == "__main__":
    main()
