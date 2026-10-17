#!/usr/bin/env python3
"""task 5.1 하네스: a071 번들 40개의 ast.json · 두 map · 위험 보고서를 HEAD(또는 base) 에서 다시 만듦.

왜 필요한가 (2026-09-25 실측):
- 번들 40개 중 15개의 ast.json 이 추출기 출력이 아니라 손으로 만든 뼈대였음
  (`branches` 에 `id` 만 있고 `kind`·`at`·`calls` 가 없음) — 좌표 역할 검사가 조용히 건너뜀.
- current 번들 7개는 base 뒤 남의 편집으로 source_sha256 이 낡았음.
- 분기 표의 행 번호가 AST 분기와 어긋나거나(행 누락·없는 행) 시험 칸이 "affected package regression" 같은
  이름 없는 주장이었음, 위험 보고서도 손으로 쓴 "none" 이었음.

그래서 손으로 옮기지 않고 다음만 씀:
- ast.json: `tools/logic-map` 추출기 출력 그대로(base 번들은 base blob 을 같은 상대 경로에 풀어 추출하고
  `revision: base` 만 덧붙임).
- 분기 표: AST 분기 순서·좌표 그대로, 조건 칸은 그 좌표의 소스 줄 원문.
- 시험 칸: `51_matrix.sh` 의 시험별 커버 프로필에서 그 분기 본문 블록을 실행한 시험을 **측정으로** 고름.
  _test.go 안 함수는 커버 계측 밖이므로 그 시험 자체(또는 호출하는 시험)와 그 PASS 를 적고 계측 밖이라고 밝힘.
- 위험 보고서: `risk_pattern_report.py` 와 같은 ast-grep 설정으로 파일을 스캔하고 함수 범위 안 결과를 따로 표시.

산문 절(입력·호출·상태·안전)은 a071 저자 기록(2026-08-04)을 그대로 두고 그 사실을 머리에 밝힘.

사용: 51_refresh.py --matrix <51_matrix.sh 출력> --basesrc <base blob 을 푼 디렉터리> --logicmap <추출기 바이너리>
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
CHANGE = HERE.parent.parent
REPO = Path(subprocess.run(["git", "-C", str(HERE), "rev-parse", "--show-toplevel"],
                           capture_output=True, text=True, check=True).stdout.strip())
BUNDLES = CHANGE / "analysis" / "function-logic"
BASE = (CHANGE / "base-commit.txt").read_text().strip()
MODULE = "github.com/JungHoonGhae/tossinvest-cli/"
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
from risk_pattern_report import markdown as risk_markdown  # noqa: E402

# base 번들의 함수는 HEAD 에 없음. 171739a4 의 diff 가 무엇으로 바꿨는지를 적은 대응표(손으로 읽은 diff 근거,
# 인용 시험 이름은 검사기가 트리에 있는지 확인함).
BASE_REPLACEMENT = {
    "internal-app-engine--testthetracerdrivesentryratchetandexit": (
        ["TestScalarStartupReadinessCannotAuthorizeTheTracer"],
        "171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함",
    ),
    "internal-execgw--gateway.protection": (
        ["TestARaisingMutationIsRefusedWhileProtectionIsUnwired", "TestProductionDefaultRefusesKRAndUSBuyBeforeBroker"],
        "171739a4 가 스칼라 판독기 `Gateway.protection` 을 지우고 `checkProtection` 이 봉인 어댑터를 읽게 함",
    ),
    "internal-execgw--isclaim": (
        ["TestNoShippedFileClaimsProtection"],
        "171739a4 가 스칼라 위조 탐지 보조 함수를 지우고 그 시험을 은퇴한 공개 스칼라 위조 봉쇄로 다시 씀",
    ),
    "internal-execgw--options.setprotectionreadyfortest": (
        ["TestNoShippedFileClaimsProtection", "TestReductionNeverReadsReadinessProvider"],
        "171739a4 가 스칼라 WIRED 세터를 지우고 export_test.go `init` 의 비스칼라 시험 하네스로 대체함",
    ),
    "internal-reconcile--testgatewayrefusesnewordersuntilrecoverycompletes": (
        ["TestEntryGateRefusesEntriesUntilRecoveryCompletes", "TestProductionDefaultRefusesKRAndUSBuyBeforeBroker"],
        "171739a4 가 같은 자리에서 이 시험을 복구 래치만 격리하는 시험으로 바꾸고 게이트웨이 준비 상태는 별도 경계로 뺌",
    ),
}
SECTIONS = ("Inputs and invariants", "Calls and live bindings", "State mutations and fallbacks", "Safety conclusion")


def qualified(ast: dict) -> str:
    return f"{ast['receiver']}.{ast['function']}" if ast.get("receiver") else ast["function"]


def cell(text: str) -> str:
    """표 칸에 넣을 소스 원문 — 파이프·백틱을 다치지 않게 바꾸고 길이를 자름."""
    text = " ".join(text.split()).replace("|", "\\|").replace("`", "'")
    text = text.replace("TODO", "T-O-D-O")
    return text if len(text) <= 110 else text[:107] + "..."


def sections(text: str) -> dict[str, list[str]]:
    out: dict[str, list[str]] = {}
    current = None
    for line in text.splitlines():
        if line.startswith("## "):
            current = line[3:].strip()
            out.setdefault(current, [])
        elif current is not None:
            out[current].append(line)
    return out


# ---- 커버 프로필 ----------------------------------------------------------------------------------

def load_matrix(matrix: Path, package: str) -> tuple[dict[str, set], dict[str, str], list]:
    """시험 이름 → 실행된 블록 집합, 시험 이름 → PASS/FAIL, 그리고 모든 블록 목록."""
    base = matrix / package.replace("/", "_")
    results = {}
    for line in (base / "results.txt").read_text().splitlines():
        verdict, name = line.split(" ", 1)
        results[name] = verdict
    executed: dict[str, set] = {}
    blocks: set = set()
    pattern = re.compile(r"^(.*):(\d+)\.(\d+),(\d+)\.(\d+) \d+ (\d+)$")
    for profile in (base / "prof").glob("*.out"):
        hit = set()
        for line in profile.read_text().splitlines()[1:]:
            m = pattern.match(line)
            if not m:
                continue
            path = m.group(1).removeprefix(MODULE)
            block = (path, int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(5)))
            blocks.add(block)
            if int(m.group(6)) > 0:
                hit.add(block)
        executed[profile.stem] = hit
    return executed, results, sorted(blocks)


def body_block(blocks: list, path: str, kind: str, line: int, column: int, end_line: int, following=None):
    """분기 좌표에 대응하는 커버 블록. switch 는 그 문장을 담은 블록(평가됨), 나머지는 좌표 바로 뒤 블록(들어감).

    `following` 은 이 분기 다음 AST 분기의 좌표 — 뒤 블록이 그보다 늦게 시작하면 이 분기에는 본문 블록이 없는
    것이므로 대응을 주장하지 않음(여러 줄 조건은 허용하되 남의 분기 본문을 가져오지 않게)."""
    mine = [b for b in blocks if b[0] == path]
    if kind == "switch":
        around = [b for b in mine if (b[1], b[2]) <= (line, column) <= (b[3], b[4])]
        return (min(around, key=lambda b: (b[3] - b[1], b[4])), "evaluated") if around else (None, "")
    # else 의 AST 좌표는 `{` 자리이고 커버 블록은 그 한 칸 앞에서 시작함(2026-09-25 실측: else at 161:9 ↔ 블록 161.8).
    floor = (line, column - 1) if kind == "else" else (line, column + 1)
    after = [b for b in mine if (b[1], b[2]) >= floor and b[1] <= end_line]
    if not after:
        return None, ""
    first = min(after, key=lambda b: (b[1], b[2]))
    if following is not None and (first[1], first[2]) > following:
        return None, ""
    return first, "entered"


def test_files(package: str) -> dict[str, str]:
    names = {}
    for path in sorted((REPO / package).glob("*_test.go")):
        for m in re.finditer(r"^func (Test\w+)\(", path.read_text(), re.M):
            names[m.group(1)] = path.name
    return names


def pick(tests: list[str], owners: dict[str, str]) -> list[str]:
    """인용할 시험 고르기: a071 시험 파일의 시험을 먼저, 그다음 이름순."""
    return sorted(tests, key=lambda t: (not owners.get(t, "").startswith("a071_"), t))


# ---- _test.go 함수의 호출자 ------------------------------------------------------------------------

def callers(package: str, needle: str) -> list[str]:
    """이 패키지 시험 파일에서 `needle` 을 부르는 줄을 담은 Test 함수(텍스트 탐색, 괄호 깊이로 범위)."""
    found = []
    for path in sorted((REPO / package).glob("*_test.go")):
        lines = path.read_text().splitlines()
        for index, line in enumerate(lines):
            m = re.match(r"^func (Test\w+)\(", line)
            if not m:
                continue
            depth, offset = 0, index
            for offset in range(index, len(lines)):
                depth += lines[offset].count("{") - lines[offset].count("}")
                if depth <= 0 and offset > index:
                    break
            if any(needle in body for body in lines[index + 1:offset + 1]):
                found.append(m.group(1))
    return sorted(set(found))


# ---- 생성 -----------------------------------------------------------------------------------------

def extract(logicmap: str, cwd: Path, file: str, function: str) -> tuple[dict, bytes]:
    run = subprocess.run([logicmap, "--file", file, "--func", function], cwd=cwd, capture_output=True, check=True)
    return json.loads(run.stdout), run.stdout


def risk(path: str, cwd: Path, start: int, end: int) -> str:
    scan = subprocess.run(["ast-grep", "scan", "--json=compact", "-c", str(REPO / "tools/logic-map/sgconfig.yml"), path],
                          cwd=cwd, capture_output=True, text=True, check=False)
    if scan.returncode not in (0, 1):
        raise RuntimeError(scan.stderr)
    findings = json.loads(scan.stdout) if scan.stdout.strip() else []
    findings = findings if isinstance(findings, list) else [findings]
    inside = [f for f in findings if start <= int(f["range"]["start"]["line"]) + 1 <= end]
    text = risk_markdown(path, findings).rstrip("\n")
    text += f"\n\n## 함수 범위 대조 (task 5.1, {start}-{end})\n\n"
    if inside:
        text += "\n".join(f"- 범위 안: `{f['ruleId']}` at {path}:{int(f['range']['start']['line']) + 1}" for f in inside)
    else:
        text += "- 범위 안 결과 0건 — 위 결과는 모두 같은 파일의 다른 함수임."
    return text + "\n"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--matrix", required=True, type=Path)
    parser.add_argument("--basesrc", required=True, type=Path)
    parser.add_argument("--logicmap", required=True)
    parser.add_argument("--head", required=True)
    args = parser.parse_args()
    matrices: dict[str, tuple] = {}
    report = []
    for bundle in sorted(p for p in BUNDLES.iterdir() if p.is_dir()):
        old = json.loads((bundle / "ast.json").read_text())
        base = old.get("revision") == "base"
        file, function = old["file"], qualified(old)
        package = str(Path(file).parent)
        cwd = args.basesrc if base else REPO
        ast, raw = extract(args.logicmap, cwd, file, function)
        if base:
            ast["revision"] = "base"
            raw = (json.dumps(ast, indent=2, ensure_ascii=False) + "\n").encode()
        (bundle / "ast.json").write_bytes(raw)
        source_lines = (cwd / file).read_text().splitlines()
        start, end = ast["start"]["line"], ast["end"]["line"]
        branches = ast.get("branches") or []
        returns = ast.get("returns") or []
        in_test_file = file.endswith("_test.go")
        rows_flm, rows_btm = [], []
        if not base and not in_test_file:
            if package not in matrices:
                matrices[package] = load_matrix(args.matrix, package) + (test_files(package),)
            executed, results, blocks, owners = matrices[package]
            population = len(results)
        anchors = [(b["at"]["line"], b["at"]["column"]) for b in branches]
        for position, branch in enumerate(branches or [None]):
            if branch is None:
                bid, anchor, text = "B1", f"branchless happy path at {start}:1", cell(source_lines[start - 1])
            else:
                bid = branch["id"]
                line, column = branch["at"]["line"], branch["at"]["column"]
                anchor = f"{branch['kind']} at {line}:{column}"
                text = cell(source_lines[line - 1])
            if base:
                names, why = BASE_REPLACEMENT[bundle.name]
                cited = " · ".join(f"`{n}`" for n in names)
                coverage = "base 소스의 분기 — HEAD 에 함수 없음"
                green = f"대체 시험 PASS at HEAD {args.head[:8]} (시험별 실행). {why}"
            elif in_test_file:
                if function.startswith("Test"):
                    names = [function]
                    note = "시험 자신"
                elif function == "init":
                    names = []
                    note = "패키지 시험 바이너리 초기화 — 이 패키지의 모든 시험이 지나감"
                else:
                    needle = ("." + ast["function"] + "(") if ast.get("receiver") else (ast["function"] + "(")
                    names = callers(package, needle)[:3]
                    note = "호출하는 시험(텍스트 탐색)"
                verdicts = matrices_verdicts(args.matrix, package, names)
                cited = " · ".join(f"`{n}`" for n in names) if names else "패키지 전체 시험"
                coverage = "_test.go — 커버 계측 밖(분기 진입 여부 미측정)"
                green = f"{note}; {verdicts}"
            else:
                if branch is not None:
                    # 다음 분기는 목록 순서(AST 순회)가 아니라 **위치**로 고름 — else 는 안쪽 if 들보다 뒤에 있음.
                    later = [a for a in anchors if a > (line, column)]
                    following = min(later) if later else None
                    twin = anchors.count((line, column)) > 1
                    block, mode = (None, "") if twin else body_block(blocks, file, branch["kind"], line, column, end, following)
                else:
                    # 분기 없는 함수: 함수 본문 첫 블록이 곧 "호출됨" 임.
                    inside = [b for b in blocks if b[0] == file and start <= b[1] <= end]
                    block, mode = (min(inside, key=lambda b: (b[1], b[2])), "called") if inside else (None, "")
                if block is None and branch is not None and anchors.count((line, column)) > 1:
                    cited, coverage = "—", "else-if: 본문 블록 없음"
                    green = "같은 좌표의 if 분기(다음 행)가 평가된다는 것이 곧 이 else 로 들어옴 — 그 행의 측정을 볼 것"
                elif block is None:
                    cited, coverage, green = "—", "커버 블록 대응 없음", "측정 불가(블록 미대응)"
                else:
                    hits = pick([t for t, s in executed.items() if block in s], owners)
                    if hits:
                        cited = " · ".join(f"`{t}`" for t in hits[:2]) + (f" (+{len(hits) - 2})" if len(hits) > 2 else "")
                        coverage = f"{mode} {len(hits)}/{population}"
                        green = f"블록 {block[1]}.{block[2]}-{block[3]}.{block[4]} 을 시험 {len(hits)}개가 실행, 전부 PASS"
                    else:
                        cited = "없음"
                        coverage = f"NOT {mode} 0/{population}"
                        green = f"블록 {block[1]}.{block[2]}-{block[3]}.{block[4]} 을 무태그 패키지 시험 어느 것도 실행하지 않음"
            rows_flm.append(f"| {bid} | {anchor} | `{text}` | {coverage} |")
            rows_btm.append(f"| {bid} | {anchor} | `{text}` | {cited} | 5.1 에서 재실행 안 함 | {green} |")
            report.append((bundle.name, bid, coverage))
        write_maps(bundle, ast, file, function, base, rows_flm, rows_btm, returns, args.head, len(branches))
        (bundle / "risk-pattern-report.md").write_text(risk(file, cwd, start, end))
    json.dump(report, sys.stdout, ensure_ascii=False, indent=0)
    return 0


_VERDICTS: dict[str, dict[str, str]] = {}


def matrices_verdicts(matrix: Path, package: str, names: list[str]) -> str:
    if package not in _VERDICTS:
        path = matrix / package.replace("/", "_") / "results.txt"
        _VERDICTS[package] = dict(reversed(l.split(" ", 1)) for l in path.read_text().splitlines())
    verdicts = _VERDICTS[package]
    if not names:
        passed = sum(1 for v in verdicts.values() if v == "PASS")
        return f"패키지 시험 {passed}/{len(verdicts)} PASS (시험별 실행)"
    return ", ".join(f"{n} {verdicts.get(n, '미실행')}" for n in names) + " (시험별 실행)"


def write_maps(bundle: Path, ast: dict, file: str, function: str, base: bool, rows_flm: list[str],
               rows_btm: list[str], returns: list, head: str, count: int) -> None:
    start, end = ast["start"]["line"], ast["end"]["line"]
    old = sections((bundle / "function-logic-map.md").read_text())
    revision = f"base `{BASE[:8]}` (HEAD 에 함수 없음)" if base else f"current — HEAD `{head[:8]}`"
    head_lines = [
        f"# Function Logic Map: `{function}`",
        "",
        f"- Source: `{file}` ({start}-{end})",
        f"- Revision: {revision}; source_sha256 `{ast['source_sha256']}`",
        "- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)",
        "- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)",
        f"- Extractor counts: AST branches {count} · returns {len(returns)} · calls {len(ast.get('calls') or [])}",
    ]
    if returns:
        head_lines.append("- Exact AST return positions: " + ", ".join(f"{r['at']['line']}:{r['at']['column']}" for r in returns))
    head_lines += [
        "- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 "
        "a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.",
        "",
    ]
    body = head_lines[:]
    for name in SECTIONS[:1]:
        body += [f"## {name}", *old.get(name, ["", "- (저자 기록 없음)", ""])]
    body += ["## Branches and early returns", "",
             "| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |",
             "|---|---|---|---|", *rows_flm, ""]
    author = old.get("Branches and early returns", [])
    concept = [re.sub(r"^\|\s*B(\d+)\s*\|", r"| N\1 |", l) for l in author if l.strip()]
    if concept:
        body += ["## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)", "", *concept, ""]
    for name in SECTIONS[1:]:
        body += [f"## {name}", *old.get(name, ["", "- (저자 기록 없음)", ""])]
    (bundle / "function-logic-map.md").write_text("\n".join(body).rstrip("\n") + "\n")
    btm = [
        f"# Branch Test Map: {function}",
        "",
        f"- Source: `{file}` ({start}-{end}); {revision}",
        "- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 "
        "실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.",
        "",
        "| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |",
        "|---|---|---|---|---|---|",
        *rows_btm,
        "",
    ]
    (bundle / "branch-test-map.md").write_text("\n".join(btm))


if __name__ == "__main__":
    raise SystemExit(main())
