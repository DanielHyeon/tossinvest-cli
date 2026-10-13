#!/usr/bin/env python3
"""a130 Branch Test Map 의 분기별 시험을 커버리지로 잼(구현 리뷰 P2-2).

편집한 함수 묶음(ast-base.json 이 있는 묶음)마다 ast.json 의 분기 좌표를 읽고, 그 분기의 **몸체 블록**(커버리지 블록 중 시작 줄이
분기 줄인 것)을 시험 하나씩 돌린 커버리지 프로필에서 찾아, 그 몸체를 실제로 실행한 시험 이름을 적음. 후보 시험이 아무도 안 덮으면
패키지 전체 시험 한 판의 프로필로 다시 봄(덮였으면 「패키지 전체」, 아니면 미도달로 남겨 사람이 사유를 적음).

쓰는 것: 이 change 의 branch-test-map.md 들(표만 다시 씀) + analysis/verify/btm-coverage.md(측정 영수증).
운영 원장 · 네트워크 없음. 사용: python3 a130_btm_coverage.py
"""
from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
CHANGE = HERE.parent.parent
ROOT = Path(subprocess.run(["git", "-C", str(HERE), "rev-parse", "--show-toplevel"],
                           capture_output=True, text=True, check=True).stdout.strip())
FLM = CHANGE / "analysis" / "function-logic"
MODULE = "github.com/JungHoonGhae/tossinvest-cli/"
# 패키지마다 후보 시험 패턴(변이 하네스의 RUNS 와 같은 묶음).
CANDIDATES = {
    "./internal/journal/": "TestA130|Adopt|Outcome|Backfill|TestRealized|TestAnAnalytics|TestAdoptionID|TestCampaignCore",
    "./internal/app/engine/": "TestA130|TestBinding|TestTheBoundExit|TestTheProjectionGuard|TestNewContext",
}


def go(args: list[str], env: dict[str, str]) -> subprocess.CompletedProcess:
    return subprocess.run(["go", *args], cwd=ROOT, capture_output=True, text=True, env=env)


def profile_blocks(path: Path) -> list[tuple[str, int, int, int, int]]:
    """(파일, 시작 줄, 시작 열, 끝 줄, 실행 횟수)."""
    out = []
    for line in path.read_text().splitlines()[1:]:
        m = re.match(r"(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$", line)
        if m:
            out.append((m.group(1).removeprefix(MODULE), int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(7))))
    return out


def taken(blocks, file: str, line: int, kind: str = "if") -> bool:
    # switch 문 자체엔 몸체 블록이 없음(갈래는 case 블록) — 그 줄을 품은 블록이 실행됐는지로 봄.
    if kind == "switch":
        return any(f == file and s <= line <= e and c > 0 for f, s, _, e, c in blocks)
    # 몸체 블록 = 그 줄에서 시작하는 블록 중 **가장 오른쪽** 것(`{` 바로 뒤). 같은 줄 왼쪽에서 시작하는 블록은 조건식을 평가하는 블록이라
    # 갈래가 열리지 않아도 실행됨(첫 판이 그것을 몸체로 잘못 세어 모든 오류 갈래를 「덮임」 으로 적었음 — 2026-10-13 원시 프로필로 확인).
    starts = [(col, c) for f, s, col, _, c in blocks if f == file and s == line]
    return bool(starts) and max(starts)[1] > 0


# 미도달 분기의 사유(사람이 원천을 읽고 적음 — 측정이 「미도달」 이라 한 것만 쓰임, 측정이 덮였다고 하면 무시됨).
DB = ("not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 "
      "결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄)")
REASONS = {
    "NewContext B2": DB, "NewContext B9": DB, "NewContext B15": DB,
    "adoptedRoundTripLegs B1": "not-applicable: 편입 기록 읽기 실패 — 편입 포지션은 adoption_id 가 가리키는 행이 FK 로 있음; 건강한 원장에서 도달 불가(design D5)",
    "adoptedRoundTripLegs B4": "not-applicable: engineSellOrders 읽기 실패 — 건강한 원장에서 도달 불가(design D5, TestA130EachReachableRefusalSiteNamesItsOwnReason 주석)",
    "ApplyExitFill B1": DB, "ApplyExitFill B6": DB, "ApplyExitFill B10": DB, "ApplyExitFill B12": DB, "ApplyExitFill B15": DB,
    "computeTradeOutcome B2": "not-applicable: positions 질의 실패 — 건강한 원장에서 도달 불가; a130 은 반환을 사유 코드로 바꿨을 뿐(사유 집합은 TestA130TheRefusalReasonSetIsClosed 가 고정)",
    "computeTradeOutcome B12": "not-applicable: 비용 추정 실패 — 앞 분기(B5 · B6)가 미구성 모델을 먼저 거르고 구성된 모델은 양의 금액에 실패하지 않음; 사유 코드만 a130",
    "computeTradeOutcome B13": "not-applicable: B12 와 같음(매도 쪽)",
    "freezeTradeOutcomeTx B2": "not-applicable: INSERT 실패 — 행 존재는 앞에서 확인했고 같은 트랜잭션이 쓰기 잠금을 쥠; 사유 코드만 a130",
    "BackfillTradeOutcome B6": "not-applicable: 앞의 TradeOutcomeOf 확인과 INSERT 사이에 다른 쓰기가 끼어야 열림(경합 전용); a130 무편집 갈래",
}
REASONS["NewContext B5"] = DB
REASONS["computeTradeOutcome B4"] = ("not-applicable: positions 행 Scan 실패 — 열 형식은 스키마가 고정(STRICT); 건강한 원장에서 도달 불가, "
                                     "사유 코드만 a130")
REASONS["BackfillTradeOutcome B5"] = DB
REASONS["RecordFill B35"] = DB
# 원시 프로필로 몸체 블록을 고쳐 잰 뒤(2026-10-13 다섯째 판) 새로 드러난 미도달 — 원천을 읽고 적음.
UNTESTED = ("not-applicable: 무편집 갈래이고 이 패키지 시험 전체가 도달하지 않음(측정) — a130 이전부터의 시험 공백, a130 범위 밖으로 기록")
REASONS.update({
    "NewContext B8": ("not-applicable: bindApplyHooks 실패 — 새로 연 원장에 훅을 한 번 다는 자리라 「이미 달림」 거절이 생산에서 열리지 않음; "
                      "a130 은 인자(싱크)만 더함, 거절 자체는 bindApplyHooks B1 · B2 가 잼"),
    "adoptedRoundTripLegs B6": "not-applicable: sumSellFills 읽기 실패 — 건강한 원장에서 도달 불가(design D5)",
    "ApplyExitFill B2": UNTESTED,
    "ApplyExitFill B4": DB,
    "ApplyExitFill B7": ("not-applicable: 저장된 taken_ratio_total 이 소수가 아닐 때만 — 쓰는 쪽이 늘 정규화 소수를 씀; 무편집 갈래"),
    "computeTradeOutcome B3": ("not-applicable: 포지션 행 없음 — 동결은 같은 트랜잭션이 막 닫은 포지션을, backfill 은 positions 에서 고른 모집단을 "
                               "넘기므로 생산에서 없음; 사유 코드만 a130"),
    "computeTradeOutcome B11": ("not-applicable: 매수 다리 수량 ≤ 0 — 편입 수량은 AdoptPosition 이 0 을 거절하고(adoption.go) 엔진 매수 다리는 "
                                "체결 수량의 합이라 건강한 원장에서 없음, 방어 갈래; 사유 코드만 a130"),
    "BackfillTradeOutcome B2": DB, "BackfillTradeOutcome B3": DB,
    "RecordFill B4": UNTESTED,
    "RecordFill B33": DB, "RecordFill B34": DB, "RecordFill B38": DB,
    "runApplyHooks B6": DB,
})
for _k in ("B5", "B9", "B10", "B13", "B16", "B17", "B18", "B21", "B22", "B24", "B26", "B27", "B29", "B32"):
    REASONS["RecordFill " + _k] = DB


def body_line(src: list[str], line: int, kind: str) -> int:
    """if · for · range 의 몸체 블록은 조건을 닫는 `{` 줄에서 시작함 — 여러 줄 조건이면 그 줄을 찾음."""
    if kind not in ("if", "for", "range"):
        return line
    for n in range(line, min(line + 12, len(src) + 1)):
        if src[n - 1].rstrip().endswith("{"):
            return n
    return line


def main() -> int:
    # -trimpath 를 쓰지 않음 — engine 의 원천 읽기 시험(runtime.Caller 경로)이 깨짐(첫 판 실측).
    env = dict(os.environ)
    scratch = Path(tempfile.mkdtemp(prefix=f"a130-btm-{os.getpid()}-"))
    bundles = []
    for d in sorted(FLM.iterdir()):
        if (d / "ast-base.json").exists():
            ast = json.loads((d / "ast.json").read_text())
            pkg = "./" + str(Path(ast["file"]).parent) + "/"
            bundles.append((d, ast, pkg))
    profiles: dict[str, dict[str, list]] = {}
    whole: dict[str, list] = {}
    for pkg in sorted({b[2] for b in bundles}):
        listed = go(["test", "-list", CANDIDATES[pkg], pkg], env).stdout.split()
        names = [n for n in listed if n.startswith("Test")]
        profiles[pkg] = {}
        for n in names:
            prof = scratch / f"{n}.out"
            r = go(["test", "-count=1", "-run", f"^{n}$", "-covermode=count", f"-coverprofile={prof}", pkg], env)
            if r.returncode != 0:
                print(f"FAIL {pkg} {n}: {r.stdout[-400:]}", file=sys.stderr)
                return 2
            profiles[pkg][n] = profile_blocks(prof)
        prof = scratch / "whole.out"
        # -skip TestA133: 병행 change a133 의 RED 시험(미커밋)이 같은 패키지에 있으면 전체 판이 그것으로 실패함 — a130 측정과 무관.
        r = go(["test", "-count=1", "-timeout", "45m", "-skip", "TestA133", "-covermode=count", f"-coverprofile={prof}", pkg], env)
        if r.returncode != 0:
            print(f"FAIL whole {pkg}: {r.stdout[-400:]}", file=sys.stderr)
            return 2
        whole[pkg] = profile_blocks(prof)
        print(f"# {pkg}: {len(names)} candidate tests measured", flush=True)

    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "-C", str(ROOT), "status", "--porcelain", "--", "internal/", "tools/"],
                           capture_output=True, text=True, check=True).stdout.strip()
    receipt = [f"# BTM 커버리지 측정 (a130 구현 리뷰 P2-2)", "",
               f"HEAD `{head}` + 워킹트리(미커밋 {len(dirty.splitlines())} 경로) · `python3 analysis/harness/a130_btm_coverage.py` · "
               "covermode=count · 시험 하나씩 프로필. 「몸체 실행」 = 조건을 닫는 `{` 줄에서 시작하는 커버리지 블록 중 가장 오른쪽 것(몸체)의 횟수 > 0 (switch 는 그 줄을 품은 블록). 미도달은 원천을 읽은 사유가 있어야 rc 0.", ""]
    unreached = 0
    for d, ast, pkg in bundles:
        lines = (d / "branch-test-map.md").read_text().splitlines()
        rows = []
        src = (ROOT / ast["file"]).read_text().splitlines()
        for b in ast.get("branches") or []:
            line = body_line(src, b["at"]["line"], b["kind"])
            names = [n for n, blk in profiles[pkg].items() if taken(blk, ast["file"], line, b["kind"])]
            names.sort(key=lambda n: (not n.startswith("TestA130"), n))
            if names:
                shown = " · ".join(f"`{n}`" for n in names[:4]) + (f" 외 {len(names) - 4}" if len(names) > 4 else "")
                green = "yes"
            elif taken(whole[pkg], ast["file"], line, b["kind"]):
                shown, green = "패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정)", "yes"
            elif f"{ast['function']} {b['id']}" in REASONS:
                shown, green = REASONS[f"{ast['function']} {b['id']}"], "n/a"
            else:
                shown, green = "미도달 — 사유 없음(사람이 REASONS 에 적어야 함)", "no"
                unreached += 1
            rows.append((b["id"], b["at"]["line"], shown, green))
        table_start = next(i for i, l in enumerate(lines) if l.startswith("| Branch"))
        old_rows = {l.split("|")[1].strip(): l for l in lines[table_start + 2:] if l.startswith("| B")}
        new = lines[:table_start + 2]
        for bid, line, shown, green in rows:
            scenario = old_rows[bid].split("|")[2].strip()
            new.append(f"| {bid} | {scenario} | {shown} | no | {green} |")
        tail = [l for l in lines[table_start + 2:] if not l.startswith("| B")]
        # 머리 문단의 시험 칸 설명을 측정 기준으로 바꿈.
        new = [re.sub(r"Test 칸은 .*$", "Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).", l)
               if l.startswith("편집 뒤 좌표") else l for l in new]
        (d / "branch-test-map.md").write_text("\n".join(new + tail).rstrip() + "\n")
        receipt.append(f"## {ast['function']} (`{ast['file']}`)")
        receipt.append("")
        for bid, line, shown, green in rows:
            receipt.append(f"- {bid} `:{line}` — {shown}")
        receipt.append("")
    receipt.append(f"사유 없는 미도달 분기: {unreached}")
    (CHANGE / "analysis" / "verify" / "btm-coverage.md").write_text("\n".join(receipt) + "\n")
    print(f"# unreached without a reason {unreached}")
    shutil.rmtree(scratch, ignore_errors=True)
    return 1 if unreached else 0


if __name__ == "__main__":
    sys.exit(main())
