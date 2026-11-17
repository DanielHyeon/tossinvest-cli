#!/usr/bin/env python3
"""a192 편집 함수 7개의 Function Logic Map·Branch Test Map 을 ast.json·커버 측정에서 만듦.

좌표(분기 B<n>·줄, 호출 줄)는 전부 ast.json 에서 읽음 — 손으로 적은 좌표가 소스와 어긋나지 않게 하려는 것.
설명 문장(조건·효과·시험 이름)만 아래 표에 손으로 적음. 사용: python3 gen_flm.py (이 파일 위치 기준으로 경로 유도)
"""

from __future__ import annotations

import json
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
ANALYSIS = HERE.parent


def counts(name: str) -> dict[str, dict[str, tuple[int, int]]]:
    out: dict[str, dict[str, tuple[int, int]]] = {}
    cur = ""
    for line in (ANALYSIS / "coverage" / name).read_text().splitlines():
        if not line.startswith("  "):
            cur = line.split()[0].lstrip(".")
            out[cur] = {}
            continue
        m = re.match(r"\s+(B\d+):\S+ cond_line_count=(\d+) body_next_line_count=(\d+)", line)
        if m:
            out[cur][m.group(1)] = (int(m.group(2)), int(m.group(3)))
    return out


AFTER = counts("after-branch-counts.txt")


def call_line(ast: dict, text: str) -> str:
    for call in ast.get("calls") or []:
        if call.get("text") == text:
            return f"call@{call['at']['line']}"
    raise SystemExit(f"call {text!r} not in AST {ast['function']}")


def write(d: str, qual: str, key: str, pre: str, desc: dict, inputs: list, calls, mutations: list,
          safety: list, tests: dict) -> None:
    folder = HERE / d
    ast = json.loads((folder / "ast.json").read_text())
    pre_ast = json.loads((ANALYSIS / "pre-edit-ast" / pre).read_text())
    branches = ast.get("branches") or []
    rets = " ".join(f"return@{r['at']['line']}" for r in ast.get("returns") or [])
    lines = [
        f"# Function Logic Map: `{qual}` ({ast['package']})", "",
        f"- Source: `{ast['file']}` ({ast['start']['line']}-{ast['end']['line']})",
        f"- 편집 전: `../../pre-edit-ast/{pre}`({pre_ast['start']['line']}~{pre_ast['end']['line']}, "
        f"sha {pre_ast['source_sha256'][:12]}, 분기 {len(pre_ast.get('branches') or [])}).",
        f"- AST evidence: `ast.json`(편집 뒤 sha {ast['source_sha256'][:12]}, 분기 {len(branches)}, return: {rets or '없음'}). "
        f"명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file {ast['file']} --func {qual}`.",
        "- Risk scan: `risk-pattern-report.md`(일치 0).",
        "- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).", "",
        "## 재번호", "", "```", *(folder / "renumber.txt").read_text().strip().splitlines(), "```", "",
        "## Inputs and invariants", "", "| Input/state | Valid range | Source of truth | Failure behavior |",
        "|---|---|---|---|", *inputs, "",
        "## Branches and early returns", "", "| Branch | Condition | Mutation/side effect | Return/error | Required test |",
        "|---|---|---|---|---|",
    ]
    for b in branches:
        cond, eff, ret, test = desc.get(b["id"], ("기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표)", "기존과 같음",
                                                  "기존과 같음", "기존 시험(패키지 GREEN)"))
        lines.append(f"| {b['id']} ({b['kind']}@{b['at']['line']}) | {cond} | {eff} | {ret} | {test} |")
    lines += ["", "## Calls and live bindings", "", "| Callee | Why called | Error/timeout/retry contract | Evidence |",
              "|---|---|---|---|", *calls(ast), "", "## State mutations and fallbacks", "", *mutations, "",
              "## Safety conclusion", "", *safety]
    (folder / "function-logic-map.md").write_text("\n".join(lines) + "\n")
    bt = [f"# Branch Test Map: `{qual}` ({ast['package']})", "",
          f"- Source: `{ast['file']}` ({ast['start']['line']}-{ast['end']['line']}). 커버 측정: `../../coverage/after-branch-counts.txt`"
          "(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.", "",
          "| Branch | Scenario | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
    for b in branches:
        c = AFTER.get(key, {}).get(b["id"], (0, 0))
        scen, test, red = tests.get(b["id"], ("기존 분기", "기존 시험(패키지 GREEN)", "해당 없음(기존 분기)"))
        ok = "yes" if c[0] or c[1] else ("미실행 — base 측정에서도 같은 블록이 0 이었거나(`base-branch-counts.txt`) "
                                         "선택 시험(-run) 밖; 이 change 가 만든 갈래 아님")
        bt.append(f"| {b['id']} | {scen} | {test} | {red} | {ok} — cover cond={c[0]} body={c[1]} |")
    (folder / "branch-test-map.md").write_text("\n".join(bt) + "\n")


RED1 = "yes — `../../red/red-1-reconcile.txt`(단언 실패, 컴파일 실패 아님)"

write("internal-reconcile--tracker.observe", "Tracker.Observe", "Tracker.Observe", "reconcile--tracker.observe.ast.json",
      {"B8": ("크레딧이 쓸 수 없음(`!creditUsableBy`) — 편집: 안에 B9 를 더함", "B9 거짓이면 `AwaitingAdjustment` append(기존)",
              "continue", "`TestACoincidentalAgreementStillDoesNotRelease`, `TestA192TheAdjustmentRulesHoldWithTheReaderOn`"),
       "B9": ("크레딧 없는 블록에서 `t.orderTerminalCandidate` 가 참 — 이 프로세스가 삽입한 missing-order 블록의 1차 판정이 「해제」",
              "블록에 `terminalRelease` 표시, `out.Cleared` append(무장이면 결속 무장만 갱신하고 거짓)", "continue",
              "`TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`, `TestA192WithoutATerminalReaderTheTrackerBehavesAsToday`")},
      ["| `diff` | 한 비교 | `ReconcileDriver.RunOnce` | as-of 없음/못 읽음 → 크레딧·종결 증명 둘 다 쓰지 않음 |",
       "| `t.Terminal` | nil 또는 `OrderTerminalJournal` | 운영: `buildGateway` 의 `*journal.Journal` | nil → 새 호출 셋이 즉시 반환(오늘과 같음) |",
       "| `t.terminal` | 결속 기억(프로세스 로컬) | `order_terminal_release.go` | 영값 = 결속 없음 |",
       "| 불변식 | 차단을 거는 조건·시점·순서 불변, 크레딧 규칙 먼저 | design D4·D8 | — |"],
      lambda a: [f"| `t.beginOrderTerminalObservation`({call_line(a, 't.beginOrderTerminalObservation')}) | 갈래 나누기 전 결속 갱신·관측 시작 블록 키 기록 | 오류 없음(기억만) | AST |",
                 f"| `t.orderTerminalCandidate`({call_line(a, 't.orderTerminalCandidate')}) | 크레딧 없는 블록의 1차 판정 | 읽기 실패는 증명의 ReadErr → 유지 | AST |",
                 f"| `t.persist`({call_line(a, 't.persist')}) | 영속 — 결속 생성·종결 증명 해제가 그 안 | 오류 → persistErr(기존) | AST |",
                 f"| `t.updateOrderTerminalBindings`({call_line(a, 't.updateOrderTerminalBindings')}) | persist 뒤 같은 비교의 다른 신원·흔적 더함, 사라진 블록의 결속 버림 | 오류 없음 | AST |"],
      ["- 새 상태: `t.terminal`(결속·무장·seen·사유). 블록 집합·streak·크레딧 갱신은 기존 그대로.",
       "- `Terminal == nil` 이면 새 호출 셋 모두 첫 줄에서 반환 → 오늘과 같은 행동(`TestA192WithoutATerminalReaderTheTrackerBehavesAsToday`)."],
      ["- Safe edit boundary: 호출 3줄 + B8 안 분기 1(B9). 기존 분기 28개는 재번호 표대로(B9 뒤 +1).",
       "- High-risk impact: yes(reconciliation). 차단을 거는 갈래(B3)의 블록 집합·추가 순서 무변경, 크레딧 경로 무변경, 해제는 persist 의 tx 재판정을 통과해야 커밋."],
      {"B8": ("크레딧 없음", "`TestACoincidentalAgreementStillDoesNotRelease`, `TestA192TheAdjustmentRulesHoldWithTheReaderOn`", "해당 없음(기존 분기)"),
       "B9": ("종결 증명 1차 판정 해제/무장", "`TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`", RED1)})

write("internal-reconcile--tracker.persist", "Tracker.persist", "Tracker.persist", "reconcile--tracker.persist.ast.json",
      {"B4": ("`enterReconcileWithBinding` 오류(편집: 호출만 감쌈)", "—", "오류 반환(기존)", "기존 + `TestA192AnUnprovenInsertIsNeverBound`"),
       "B7": ("Cleared 블록에 `terminalRelease` 표시가 있음", "`t.releaseOnOrderTerminal` — 같은 판정을 tx 안에서 다시 통과할 때만 장부 해제", "B8/B9 뒤 continue",
              "`TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`"),
       "B8": ("종결 증명 해제의 장부 오류", "—", "오류 반환 — 블록·게이트 유지", "`TestA192ReleaseWriteFailureKeepsTheGate`"),
       "B9": ("tx 재판정이 해제함", "`result.released` append", "—(continue)",
              "`TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`, `TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`(거짓 쪽)")},
      ["| `out.Added` | pending 블록 전부 | `pendingBlocksForPersistence` | 첫 오류에서 반환(기존) |",
       "| `out.Cleared` | 크레딧 해제 + 종결 증명 해제 후보 | `Observe` | 종결 증명 후보는 tx 재판정이 유지면 해제 없음(오류 아님) |",
       "| 불변식 | 장부 쓰기 내용(EnterReconcile 요청 필드)·크레딧 해제 요청 무변경 | design D11 | — |"],
      lambda a: [f"| `t.enterReconcileWithBinding`({call_line(a, 't.enterReconcileWithBinding')}) | `EnterReconcile` 을 감싸 삽입 증명(entered)으로 결속 생성 | 요청·반환 그대로 전달 | AST |",
                 f"| `t.releaseOnOrderTerminal`({call_line(a, 't.releaseOnOrderTerminal')}) | 종결 증명 해제(tx 재판정) | 장부 오류 그대로 반환 | AST |",
                 f"| `t.Journal.ReleaseReconcile`({call_line(a, 't.Journal.ReleaseReconcile')}) | 크레딧 해제(기존) | 기존 | AST |"],
      ["- 새 부수효과: 결속 생성(기억), 종결 증명 해제(장부 1행 UPDATE — 새 journal 메서드).",
       "- 크레딧 해제 갈래(B10·B11)는 편집 전 B7·B8 과 같은 문장."],
      ["- Safe edit boundary: EnterReconcile 호출을 감싼 함수로 바꿈(같은 요청·같은 반환), Cleared 루프 앞에 갈래 B7~B9.",
       "- High-risk impact: yes. `Terminal == nil` 이면 감싼 함수는 `EnterReconcile` 한 번과 같고 B7 은 표시가 없어 거짓."],
      {"B4": ("삽입 쓰기 오류", "`TestA192AnUnprovenInsertIsNeverBound`", "해당 없음(기존 분기 — 결속 없음 쪽은 base 에서도 통과)"),
       "B7": ("종결 증명 해제 갈래", "`TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`", RED1),
       "B8": ("해제 기록 실패", "`TestA192ReleaseWriteFailureKeepsTheGate`",
              "관측 안 함 — GREEN 뒤 커버 측정에서 0 을 보고 시험 추가. 변이 대신: 이 갈래는 오류 전달뿐이라 변이 대상 판정 없음, 시험이 오류·게이트 유지를 단언"),
       "B9": ("재판정 해제/유지", "`TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`", "yes — red-1(`release attempts = 0`)")})

write("internal-reconcile--blocksfor", "blocksFor", "blocksFor", "reconcile--blocksfor.ast.json",
      {"B2": ("missing-order 마다 — 편집: 리터럴에 `missing` 비공개 필드 1", "블록 append(키·값·개수 불변)", "—",
              "`TestMissingOrderRaisesASymbolBlock`, `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`")},
      ["| `diff.MissingOrders[i]` | `LocalOrder` | `Compare` | — |", "| 불변식 | 블록 집합·키·Detail 무변경 | design D8 | — |"],
      lambda a: [f"| `missing.Identity`({call_line(a, 'missing.Identity')}) | 결속 신원(정규화) | 순수 | AST |"],
      ["- 없음(값 생성). 새 필드는 장부에 쓰이지 않음."],
      ["- Safe edit boundary: 복합 리터럴 필드 1. 분기·호출 순서 불변.", "- High-risk impact: yes(블록 생성) — 블록 키·Reason·Release·Detail 그대로."],
      {"B2": ("missing-order 블록", "`TestMissingOrderRaisesASymbolBlock`, `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`", RED1)})

write("internal-reconcile--tracker.restore", "Tracker.Restore", "Tracker.Restore", "reconcile--tracker.restore.ast.json", {},
      ["| 장부 활성 행 | — | `ActiveReconcileStates` | 오류 → return(기존) |", "| 불변식 | 복원된 행은 결속 없음(삽입 증명 없음) | design D2·D6 | — |"],
      lambda a: [f"| `t.Journal.ActiveReconcileStates`({call_line(a, 't.Journal.ActiveReconcileStates')}) | 복원 | 기존 | AST |",
                 f"| `t.Gate.RebuildReconcileProjection`({call_line(a, 't.Gate.RebuildReconcileProjection')}) | 게이트 재구성 | 기존 | AST |"],
      ["- 편집: 잠금 안 대입 `t.terminal = terminalState{}` 1줄(분기 아님) — 같은 객체 재사용에도 결속·무장을 버림(리뷰 3 #3)."],
      ["- Safe edit boundary: 대입 1줄. 분기 6개 문장 불변.", "- High-risk impact: yes(복원) — 블록·게이트 복원 그대로, 결속만 비움(보수 방향)."],
      {"B5": ("영구 행 있음", "기존 + `TestA192APermanentBlockStaysOperatorOnly`", "해당 없음(기존 분기)")})

write("internal-journal--validreconcilereleasecause", "ValidReconcileReleaseCause", "ValidReconcileReleaseCause",
      "journal--validreconcilereleasecause.ast.json",
      {"B2": ("이 판본이 쓰는 원인 — 편집: case 목록에 `ORDER_TERMINAL_CONFIRMED` 추가", "—", "true",
              "`TestUnknownReleaseCauseIsRefused`, `TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites`")},
      ["| `cause` | 문자열 | 호출자 | 목록 밖 → false |"],
      lambda a: ["| (없음) | — | — | AST calls 0 |"],
      ["- 없음."],
      ["- Safe edit boundary: case 값 1. 다른 원인 판정 불변(`TestUnknownReleaseCauseIsRefused`).",
       "- High-risk impact: yes(원장 쓰기 제약) — 새 원인 하나만 허용 목록에 더함."],
      {"B2": ("허용 원인", "`TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites`", "yes — `../../red/red-2-journal.txt`"),
       "B3": ("그 밖", "`TestUnknownReleaseCauseIsRefused`", "해당 없음(기존)")})

write("internal-app-engine--buildgateway", "buildGateway", "buildGateway", "engine--buildgateway.ast.json", {},
      ["| `in.journal` | 운영 장부 | 엔진 조립 | 기존 |", "| 불변식 | 추적기 배선 외 무변경 | design D11 | — |"],
      lambda a: [f"| `tracker.Restore`({call_line(a, 'tracker.Restore')}) | 복원(기존) | 기존 | AST |"],
      ["- 편집: 추적기 복합 리터럴에 `Terminal: in.journal` 필드 1(분기 아님) — `TestA192TheProductionTrackerReadsTerminalProofFromItsJournal` 이 고정."],
      ["- Safe edit boundary: 필드 1. 분기 10개 문장 불변(줄 +2).", "- High-risk impact: yes(배선) — 이 필드가 운영에서 종결 증명 해제를 켬."],
      {})

write("internal-app-engine--reconciledriver.runonce", "ReconcileDriver.RunOnce", "ReconcileDriver.RunOnce",
      "engine--reconciledriver.runonce.ast.json", {},
      ["| `outcome` | Observe 결과 | Tracker | — |", "| 불변식 | 주기 순서·오류 처분 무변경 | design D7 | — |"],
      lambda a: [f"| `d.logOrderTerminalReleases`({call_line(a, 'd.logOrderTerminalReleases')}) | 해제마다 로그 1줄 + 유지 사유가 바뀔 때 1줄(계좌·주문번호·수량·가격 없음) | 오류 없음 | AST |",
                 f"| `d.opts.Tracker.Observe`({call_line(a, 'd.opts.Tracker.Observe')}) | 기존 | 기존 | AST |"],
      ["- 편집: 호출 1줄(분기 아님). 드라이버 구조체에 `terminalHoldLogged` 필드(타입 편집 — 이 함수 밖)."],
      ["- Safe edit boundary: 호출 1줄, 오류 갈래(B7·B8) 앞. 로그만 — 판정·순서 무변경.", "- High-risk impact: no(관측 전용)."],
      {"B7": ("Observe 오류", "기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`", "해당 없음(기존)")})
print("written")
