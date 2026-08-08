#!/usr/bin/env python3
"""`check_values.py`의 방어가 **실제로 시험되고 있는지**를 기계로 센다.

왜 이것이 있는가
================

7·8·9판이 **세 판 연속 같은 형태**로 거부됐다.

- 7판: 방어를 쓰고 안 돌렸다. 자기가 잡는다고 적은 `28.9`를 못 잡았다.
- 8판: 자체 시험을 만들었다. 그런데 케이스 8건이 **전부 이전 라운드 우회의 재생**이라
  8판이 새로 넣은 세 구성(`≥` 룩비하인드·낱말 면제·§7.5 수치)을 **0건** 덮었고,
  그 셋이 그대로 8라운드 차단 B2·B3·B4가 됐다.
- 9판: 그 진단을 문장으로 적고("검사에 줄을 추가하면 그 줄을 우회하는 케이스를 같은
  커밋에") 11건을 더했다. **그 11건도 전부 회귀 방지였고 신규 줄을 다시 0건 덮었다.**
  9라운드가 우회 34건을 만들어 34건 전부 통과시켰다.

세 번 다 사람이 "덮었다"고 판단했고 세 번 다 틀렸다. **그러므로 판단을 그만두고
센다.** 이 스크립트가 10판의 진입 조건이다.

두 축으로 센다
==============

**축 1 — `fail()` 자리 뮤테이션.** `check_values.py`가 실패를 보고할 수 있는 자리는
`fail()` 호출 하나하나다. 그 자리를 **하나씩 죽여** 자체 시험을 돌린다. 죽였는데도
자체 시험이 그대로 전부 통과하면, 그 실패 모드에는 **대응하는 케이스가 없다.**

  잡는 것: "검사에 줄을 추가했는데 그 줄을 밟는 케이스가 없다"
  못 잡는 것: 정규식이 **너무 좁아서** 애초에 못 보는 입력 (→ 축 2)

**축 2 — 섭동 행렬.** 알려진 위반 한 줄을 기계적으로 변형해 전 위치에 심고,
전부 잡히는지 본다. 9라운드 우회 34건은 전부 이 행렬의 한 칸이었다.

  변형: 단위 뒤 마침표 · 한국어 조사 · 정수 ms 재표기 · 부등호 접두 · 굵게 ·
        표 셀 · 낱말 게이트 회피 · 자릿수 구분 · 항 순서 뒤집기 · 열거 분할
  위치: proposal · design · tasks · spec · analysis · FLM

  잡는 것: "이 방어가 이 형태의 입력을 못 본다"

침묵한 면제 금지
================

덮이지 않아도 되는 자리가 있다(예: 0매치 가드는 문서가 정상인 한 안 밟힌다).
그런 자리는 `EXEMPT`에 **사유와 함께** 적는다. 사유 없는 미커버는 실패다.

실행
====

    python3 openspec/changes/a092-.../tools/coverage_gate.py
    python3 openspec/changes/a092-.../tools/coverage_gate.py --axis 1   # 뮤테이션만
    python3 openspec/changes/a092-.../tools/coverage_gate.py --axis 2   # 섭동만

저장소를 **읽기만 한다.** 전부 임시 사본에서 돈다. 종료 코드 0이면 통과.
"""

from __future__ import annotations

import argparse
import ast
import pathlib
import re
import shutil
import subprocess
import sys
import tempfile

CHANGE = pathlib.Path(__file__).resolve().parent.parent
CHANGE_ID = CHANGE.name

# ---------------------------------------------------------------------------
# 축 1 — `fail()` 자리 뮤테이션
# ---------------------------------------------------------------------------

# `fail()` 을 **스택 경로**로 선택적으로 무력화하는 shim. 임시 사본에만 넣는다.
#
# **10라운드 차단 4 — 줄 하나로 죽이면 래퍼의 호출자들이 라벨 하나로 붕괴한다.**
# 10판의 shim은 `stack()[1].lineno`, 즉 **직전** 호출자만 봤다. `_dwell_fail`은
# `fail()` 을 한 줄(`:787`)에서 부르고 그 함수를 두 곳(`:809` 요구 문단 ·
# `:827` Scenario)이 부른다. `:787`을 죽이면 둘이 동시에 죽으므로 "호출자 ①만
# 무력화된 상태"를 축 1이 표현할 수 없었다. `claims_adopted`(`fail@:311`,
# 호출자 `:339/:359/:379`)도 같다.
#
# 그래서 죽이는 단위를 **경로**로 바꾼다. `A092_MUTE_PATH`는 안쪽부터 바깥쪽으로
# 세는 줄 번호 목록이고, check_values.py 안의 프레임만 센다. `787,809`는
# "`:809`가 부른 `_dwell_fail`이 `:787`에서 낸 실패"만 죽인다.
FAIL_DEF = (
    "def fail(where: str, msg: str) -> None:\n"
    "    FAILURES.append(f\"{where}\\n      {msg}\")\n"
)
FAIL_SHIM = (
    "def fail(where: str, msg: str) -> None:\n"
    "    import inspect as _i, os as _o\n"
    "    _m = _o.environ.get(\"A092_MUTE_PATH\")\n"
    "    if _m:\n"
    "        _want = [int(_x) for _x in _m.split(\",\")]\n"
    "        _chain = [_s.lineno for _s in _i.stack()[1:]\n"
    "                  if _s.filename == __file__]\n"
    "        if _chain[:len(_want)] == _want:\n"
    "            return\n"
    "    FAILURES.append(f\"{where}\\n      {msg}\")\n"
)

# 덮이지 않아도 되는 `fail()` 자리 — 사유를 반드시 적는다.
#
# 키는 그 자리의 **진단 문구 조각**이다(줄 번호는 편집마다 바뀌므로 쓰지 않는다).
EXEMPT: dict[str, str] = {
    "0매치는 통과가 아니다(표 모양이 바뀌었으면":
        "0매치 가드 — 정상 문서에서는 정의상 안 밟힌다. 이 가드를 시험하려면 "
        "표를 통째로 지운 사본이 필요한데, 그것은 검사가 아니라 파일 존재 확인이다. "
        "**후보 표 행의 모양 변형은 덮이지 않는다** — 축 2의 '표 셀'은 일반 행에 "
        "값을 심을 뿐 후보 표의 행 모양을 바꾸지 않는다(10라운드 M3·B29). "
        "미탐 계열 U3에 적었다.",
    "design.md에서 후보 행을 하나도 못 찾았다":
        "위와 같은 0매치 가드.",
    "s가 analysis/ 어디에도 없다":
        "COLD_PUBLISH_S 선언과 코퍼스의 연결 — **축 1**의 자체 시험 케이스 "
        "'8라운드 M-8 — design이 말하는 냉 측정값이 검사 밖이다'가 덮는다. "
        "(10판은 이것을 '축 2의 냉 측정 변경 칸'이라 적었는데 축 2에 그런 칸은 "
        "없다 — 10라운드 M3의 축 오귀속.)",
    "BUILD OK/FAIL 행을 하나도 못 찾았다":
        "0매치 가드 — 표가 통째로 사라졌는지를 보는 검사이고, 그것은 값 검사가 "
        "아니라 파일 존재 확인이다. 표 안의 값이 틀리는 형태는 **축 1**의 "
        "'9라운드 H-2 — 산문이 D3 표와 다른 구성 수를 말한다'가 덮는다. "
        "**행을 더하는 형태는 덮이지 않는다** — 미탐 계열 U5에 적었다.",
    "체류 구성을 세는 문단을 하나도 못 찾았다":
        "0매치 가드. 열거 자체를 지우는 형태는 축 2의 '열거 제거' 칸이 덮는다.",
}


# ---------------------------------------------------------------------------
# 미탐 계열 — 이 두 축으로는 **원리적으로** 안 잡히는 것
# ---------------------------------------------------------------------------

# 10라운드가 우회 36건을 만들어 35건을 통과시켰다. 절반은 정규식으로 닫혔고
# (11판이 닫았다) 절반은 **닫히지 않는다.** 닫히지 않는 이유가 같다:
#
#   이 두 축이 세는 것은 **값의 일치**다. 문장의 참이 아니다.
#
# 그러므로 값이 하나도 안 틀린 채로 문장만 틀리면 두 축 다 조용하다. 여기까지는
# 스크립트 머리말이 이미 적어 둔 한계다. **11판이 더하는 것은 그 한계를 매 실행마다
# 소리 내어 말하는 것이다.** 못 세는 것을 안 적으면 "커버리지 게이트 통과"라는 줄이
# "전부 덮었다"로 읽힌다 — 그것이 7·8·9판이 세 번 연속 저지른 형태다.
#
# 이 목록은 리뷰어에게 **어디를 사람이 봐야 하는지**를 알려주는 것이 목적이다.
UNSEEABLE: list[tuple[str, str, str]] = [
    ("U1", "표지의 뜻이 문장과 어긋난다 (B13)",
     "`<!-- rejected-value -->`는 '이 줄은 기각을 *기록*한다'는 뜻인데, 두 기각값을 "
     "**근거로 삼는** 줄에 붙여도 통과한다. 표지에 사유 칸이 없어 기계가 뜻을 "
     "구별할 수 없다. 사유를 요구해도 사유의 참은 여전히 사람이 읽어야 한다."),
    ("U2", "두 delta가 서로 반대를 말한다 (B17)",
     "engine-safety가 'critical은 1회', exit-policy가 '체류 상한은 critical에만'이라 "
     "적어도 **값이 하나도 안 틀린다.** 등급 정책의 모순은 값 비교로 안 잡힌다."),
    ("U3", "후보 표의 채택 라벨이 다른 행으로 옮겨간다 (B18·B29)",
     "각 행의 산술은 그대로 맞으므로 행 검사는 통과한다. 어느 행이 '채택'인지는 "
     "라벨의 뜻이고, 엄격 매치가 하나도 없는 유령 표는 후보 표로 인식조차 안 된다."),
    ("U4", "코퍼스 원천을 넓혀 임의값을 인용 가능하게 만든다 (B19·B20)",
     "`analysis/`와 FLM은 코퍼스 원천이고 거기 실린 수는 정의상 인용 근거가 된다. "
     "무엇을 코퍼스에 싣는지는 문서 작성자의 자유이므로 고아 검사의 전제 자체를 "
     "옮길 수 있다. 코퍼스가 단위를 안 보는 것도 같은 뿌리다(B21)."),
    ("U5", "측정 표에 거짓 행을 **더한다** (B30·B31)",
     "행을 빼면 카운트 대조가 잡지만 더하면 카운트도 같이 는다. 열거에 항을 "
     "더하는 것도 같다 — `_dwell_hits`는 **빠진 항만** 센다."),
    ("U6", "`review.md` 안의 값 (10라운드 H5)",
     "리뷰 기록은 기각값과 우회 문자열을 인용하는 것이 일이라 검사 대상에서 뺐다. "
     "뺀 결과로 그 안의 값·판정 기록은 미검사다. 근거 인용이 review.md를 가리키는 "
     "문장이 있으면 그 인용은 사람이 확인해야 한다."),
    ("U7", "프로덕션 Go 소스와 문서의 대조 (6라운드 H2 형태)",
     "값은 안 틀리고 문장만 틀리는 결함. 이 스크립트는 문서끼리만 대조한다. "
     "소스 대조는 독립 리뷰의 일이고 10라운드가 172/172 인용을 그렇게 확인했다."),
]


def report_unseeable(report: list[str]) -> None:
    report.append("")
    report.append("## 이 게이트가 **못 세는 것** — 사람이 봐야 하는 자리")
    report.append("")
    for tag, title, why in UNSEEABLE:
        report.append(f"  {tag}  {title}")
        report.append(f"      {why}")
    report.append("")
    report.append("  두 축은 **값의 일치**를 센다. 문장의 참은 세지 않는다.")
    report.append("  위 계열이 비어 있다는 뜻이 아니라, 비었는지를 이 도구가 "
                  "모른다는 뜻이다.")


def fail_paths(src: str) -> list[tuple[tuple[int, ...], str, str]]:
    """`fail()` 이 도달할 수 있는 **경로**를 AST로 열거한다.

    반환은 `(경로, 라벨, 진단 문구 조각)`. 경로는 안쪽부터 바깥쪽으로 센 줄 번호다.

    **줄이 아니라 경로인 이유는 10라운드 차단 4다.** 래퍼를 거치는 호출은 `fail()`
    줄이 같으므로 줄로 세면 호출자들이 라벨 하나로 붕괴하고, 그러면 호출자 하나만
    죽은 상태를 축 1이 표현할 수 없다.

    **손으로 세지 않는다.** 정규식으로 줄을 훑던 10판의 방식은 래퍼를 볼 수 없었다.
    래퍼인지 아닌지는 호출 그래프의 성질이지 줄의 성질이 아니다.
    """
    tree = ast.parse(src)
    lines = src.splitlines()

    direct: dict[str, list[int]] = {}          # 함수 -> 직접 fail() 줄
    called_at: dict[str, list[int]] = {}       # 함수 -> 그 함수가 불린 줄

    class Walk(ast.NodeVisitor):
        def __init__(self) -> None:
            self.fn: str | None = None

        def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
            prev, self.fn = self.fn, node.name
            self.generic_visit(node)
            self.fn = prev

        def visit_Call(self, node: ast.Call) -> None:
            f = node.func
            name = (f.id if isinstance(f, ast.Name)
                    else f.attr if isinstance(f, ast.Attribute) else None)
            if name == "fail" and self.fn is not None:
                direct.setdefault(self.fn, []).append(node.lineno)
            elif name is not None and self.fn != name:
                called_at.setdefault(name, []).append(node.lineno)
            self.generic_visit(node)

    Walk().visit(tree)

    def blob_at(line: int) -> str:
        return " ".join(lines[line - 1:line + 4])

    out: list[tuple[tuple[int, ...], str, str]] = []
    for fn, fail_lines in sorted(direct.items(), key=lambda kv: kv[1][0]):
        callers = sorted(called_at.get(fn, []))
        for fl in sorted(fail_lines):
            if len(callers) >= 2:
                # 래퍼 — 호출자마다 별개 자리다.
                for cl in callers:
                    out.append(((fl, cl),
                                f"check_values.py:{fl} ← 호출자 :{cl}",
                                blob_at(fl) + " " + blob_at(cl)))
            else:
                out.append(((fl,), f"check_values.py:{fl}", blob_at(fl)))
    return out


def caught_names(out: str) -> set[str]:
    """자체 시험 출력에서 **잡힌 케이스 이름**을 뽑는다.

    축 1의 `want` 비교가 이것 위에 선다. 10판은 `rc != 0`이면 무엇이든 "죽었다"로
    셌고, 그러면 무력화가 **크래시**를 내도 커버리지로 세어졌다 — 10라운드 차단 4의
    후반부. 이제는 "어떤 케이스가 잡기를 그만뒀는가"를 묻는다.
    """
    return {l.split("잡았다 — ", 1)[1].strip()
            for l in out.splitlines() if l.startswith("잡았다 — ")}


def prepare_mutable(tmp: pathlib.Path) -> pathlib.Path:
    """change 디렉터리를 복사하고 `fail()` 에 shim을 심는다."""
    base = tmp / "openspec" / "changes"
    base.mkdir(parents=True, exist_ok=True)
    root = base / CHANGE_ID
    shutil.copytree(CHANGE, root,
                    ignore=shutil.ignore_patterns("__pycache__"))
    checker = root / "tools" / "check_values.py"
    src = checker.read_text(encoding="utf-8")
    if FAIL_DEF not in src:
        raise SystemExit(
            "축 1을 못 돌린다: check_values.py 의 fail() 정의가 예상과 다르다.\n"
            "이 하니스가 그 정의를 기준으로 shim을 심는다 — 정의가 바뀌었으면 "
            "FAIL_DEF 를 고친다. 앵커가 사라진 채로 '통과'하면 그것이 이 판이 "
            "고치려는 결함 그 자체다.")
    checker.write_text(src.replace(FAIL_DEF, FAIL_SHIM, 1), encoding="utf-8")
    return root


def run_selftest(root: pathlib.Path,
                 mute_path: tuple[int, ...] | None) -> tuple[int, str]:
    env = {"PATH": "/usr/bin:/bin", "LANG": "C.UTF-8"}
    if mute_path is not None:
        env["A092_MUTE_PATH"] = ",".join(str(n) for n in mute_path)
    proc = subprocess.run(
        [sys.executable, str(root / "tools" / "check_values_selftest.py")],
        capture_output=True, text=True, env=env)
    return proc.returncode, proc.stdout + proc.stderr


def exemption_for(blob: str) -> str | None:
    for frag, why in EXEMPT.items():
        if frag in blob:
            return why
    return None


def axis_one(report: list[str]) -> int:
    bad = 0
    with tempfile.TemporaryDirectory() as tmp:
        root = prepare_mutable(pathlib.Path(tmp))
        # **라벨은 원본 좌표로, 뮤테이션은 사본 좌표로.**
        #
        # shim이 `fail()` 정의를 길게 바꾸므로 사본의 줄 번호는 정의 아래로
        # 전부 밀린다. 10판은 사본을 그대로 파싱해서 라벨을 냈고, 그래서
        # 진단이 가리키는 줄이 실제 파일에 **없는** 줄이었다. 리뷰어가 그 줄을
        # 열면 다른 코드가 나온다 — 8라운드 B2가 "진단은 원문의 표기를
        # 가리킨다"로 요구한 것과 같은 규칙을 줄 번호에도 적용한다.
        src = (CHANGE / "tools" / "check_values.py").read_text(encoding="utf-8")
        sites = fail_paths(src)
        shift = FAIL_SHIM.count("\n") - FAIL_DEF.count("\n")
        def_line = next(i for i, l in enumerate(src.splitlines(), 1)
                        if l.startswith("def fail("))

        def to_copy(line: int) -> int:
            return line + shift if line > def_line else line

        rc, out = run_selftest(root, None)
        if rc != 0:
            report.append("축 1 기준선 실패 — shim을 심은 사본에서 자체 시험이 "
                          "이미 깨진다. 아래 결과는 무의미하다.\n" + out)
            return 1
        # 자체 시험의 케이스 수는 **센다.** 진단에 수를 박으면 케이스를 더한
        # 다음 판에서 그 문장이 조용히 거짓이 된다 — 이 저장소가 세 번 겪은 형태다.
        base_names = caught_names(out)
        nwrap = len([s for s in sites if len(s[0]) > 1])
        report.append(f"축 1 기준선 통과 — shim 사본에서 자체 시험 정상 "
                      f"(`fail()` 경로 {len(sites)}곳 — 그중 래퍼 경유 {nwrap}곳 · "
                      f"케이스 {len(base_names)}건)")

        for path, label, blob in sites:
            rc, out = run_selftest(root, tuple(to_copy(n) for n in path))
            lost = base_names - caught_names(out)
            if rc != 0 and lost:
                # **`want` 비교.** 죽은 것으로 세려면 "어떤 케이스가 잡기를
                # 그만뒀는지"를 말할 수 있어야 한다. 아무 이유로나 실패하는 것은
                # 커버리지가 아니다 — 자체 시험이 케이스마다 이미 쓰는 규칙을
                # 축 1에도 같은 모양으로 적용한다.
                report.append(f"  죽었다   {label} — 잡기를 그만둔 케이스 "
                              f"{len(lost)}건: {', '.join(sorted(lost))}")
                continue
            if rc != 0 and not lost:
                report.append(f"  **크래시** {label} — 무력화가 자체 시험을 "
                              f"비정상 종료시켰다. 잡기를 그만둔 케이스가 없으므로 "
                              f"이것은 커버리지가 아니다.\n{out[-400:]}")
                bad += 1
                continue
            why = exemption_for(blob)
            if why:
                report.append(f"  면제     {label} — {why}")
                continue
            snippet = re.sub(r"\s+", " ", blob)[:96]
            report.append(f"  **미커버** {label} — 이 자리를 죽여도 자체 시험이 "
                          f"{len(base_names)}건 그대로 통과한다. 대응 케이스가 없다.\n"
                          f"             {snippet}")
            bad += 1
    return bad


# ---------------------------------------------------------------------------
# 축 2 — 섭동 행렬
# ---------------------------------------------------------------------------

# 심는 위치. `analysis/`와 FLM은 **코퍼스 원천**이므로 여기에 표지 없는 기각값이
# 들어가면 blocklist를 피하면서 동시에 코퍼스로 재진입한다(9라운드 B-3).
SITES: dict[str, str] = {
    "proposal": "proposal.md",
    "design": "design.md",
    "tasks": "tasks.md",
    "spec": "specs/engine-safety/spec.md",
    "analysis": "analysis/delivery-latency.md",
    "FLM": ("analysis/function-logic/internal-obs--notifier.deliver/"
            "function-logic-map.md"),
}

# 인용 문서 — 측정을 **인용만** 하는 자리. 고아 검사의 대상이다.
CITING_SITES = ("proposal", "design", "tasks", "spec")

# 씨앗 — 잡혀야 하는 위반. `{v}` 자리에 값이 들어간다.
#
# 각 씨앗은 (이름, 문장 틀, 값, 출력 조각, 심을 위치).
#
# **고아 씨앗의 위치가 인용 문서뿐인 것은 오탐 회피가 아니라 정의다.** 측정치를
# `analysis/`에 싣는 것이 곧 고아가 아니게 만드는 방법이다. 반면 **기각값은 어디에
# 있어도 위반**이다 — `analysis/`와 FLM은 코퍼스 원천이므로 표지 없는 기각값이
# 거기 있으면 blocklist를 피하면서 동시에 코퍼스로 재진입한다(9라운드 B-3).
# 기대 문구는 **섭동이 만든 텍스트에서 뽑는다.** 씨앗에 박아 두면 정수 ms 재표기처럼
# 값의 표기를 바꾸는 섭동에서 기대가 틀려지고, 그러면 "잡았지만 이유가 다르다"가
# 오탐으로 쌓인다. 진단이 원문의 표기를 가리켜야 한다는 요구(8라운드 B2)를
# 하니스 쪽에서도 같은 규칙으로 적용하는 것이다.
SEEDS: list[tuple[str, str, str, str, tuple[str, ...]]] = [
    ("기각값", "이 회차의 초과분 최악은 {v}였다.", "28.9 ms",
     "기각된 값 ", tuple(SITES)),
    ("기각값2", "그 배수의 최악 체류는 {v}다.", "9.231 s",
     "기각된 값 ", tuple(SITES)),
    ("고아", "이 회차의 초과분 최악은 {v}였다.", "77.77 ms",
     "고아 측정치 ", CITING_SITES),
]

# 섭동이 만든 문장에서 실제로 심긴 수치 표기를 뽑는다.
#
# **자릿수 구분을 수의 일부로 읽어야 한다.** 첫 판은 `\d+`만 봐서 `9,231 ms`에서
# `231`을 뽑았고, 검사가 옳게 `9,231`을 가리켰는데도 하니스가 "이유가 다르다"로
# 12칸을 실패시켰다 — 하니스가 검사에 요구하는 규칙("진단은 원문의 표기를 가리킨다")을
# 하니스 자신이 안 지킨 것이다.
INJECTED = re.compile(
    r"(\d{1,3}(?:[,  ]\d{3})+(?:\.\d+)?|\d+(?:\.\d+)?)\s*(?:ms|s|밀리초|초)")

# 섭동 — 문장/값을 기계적으로 변형한다. 각각 (이름, 함수).
#
# 9라운드가 손으로 찾은 34건의 우회는 전부 이 표의 한 칸이다. 손으로 찾는 것을
# 그만두는 것이 이 표의 목적이다.
def _p_plain(sent: str, val: str) -> str:
    return sent.format(v=val)


def _p_period(sent: str, val: str) -> str:
    """단위 뒤 마침표 — 9라운드 B-1. 이 코퍼스의 native 형태다."""
    return "- 초과분 최악: " + val + "."


def _p_particle(sent: str, val: str) -> str:
    """한국어 조사가 단위에 붙는다 — 7라운드 M-1(c)."""
    return sent.format(v=val.replace(" ", ""))


def _p_int_ms(sent: str, val: str) -> str:
    """정수 ms 재표기 — 9라운드 B-2. 9.231 s → 9231 ms."""
    m = re.match(r"^(\d+(?:\.\d+)?)\s*(ms|s)$", val)
    if not m:
        return ""
    v, unit = float(m.group(1)), m.group(2)
    ms = v * 1000 if unit == "s" else v
    if abs(ms - round(ms)) > 1e-9:
        return ""
    return sent.format(v=f"{int(round(ms))} ms")


def _p_bound(sent: str, val: str) -> str:
    """부등호 접두 — 8라운드 B2."""
    return sent.format(v="≥ " + val)


def _p_bold(sent: str, val: str) -> str:
    return sent.format(v="**" + val + "**")


def _p_cell(sent: str, val: str) -> str:
    """표 셀 안."""
    return f"| 회차 | 초과분 | {val} |"


def _p_no_latency_word(sent: str, val: str) -> str:
    """지연 낱말 게이트 회피 — 8라운드 M-4 / 9라운드 B-2."""
    return f"운영자가 관측한 값은 {val}다."


def _p_out_of_scope(sent: str, val: str) -> str:
    """OUT_OF_SCOPE 낱말과 같은 줄 — 8라운드 M-4."""
    return sent.format(v=val) + " (timeout 설정 기준)"


def _p_comma(sent: str, val: str) -> str:
    """자릿수 구분 쉼표 — 9판은 `4,777ms`에서 `777`만 읽어 원문에 없는 수를 냈다."""
    t = _p_int_ms(sent, val)
    if not t:
        return ""
    return re.sub(r"(?<![\d.])(\d)(\d{3})(?=\s*ms)", r"\1,\2", t)


def _p_space(sent: str, val: str) -> str:
    """같은 것, 공백 구분."""
    t = _p_int_ms(sent, val)
    if not t:
        return ""
    return re.sub(r"(?<![\d.])(\d)(\d{3})(?=\s*ms)", r"\1 \2", t)


PERTURBATIONS = [
    ("원형", _p_plain),
    ("자릿수 쉼표", _p_comma),
    ("자릿수 공백", _p_space),
    ("단위 뒤 마침표", _p_period),
    ("조사 밀착", _p_particle),
    ("정수 ms 재표기", _p_int_ms),
    ("부등호 접두", _p_bound),
    ("굵게", _p_bold),
    ("표 셀", _p_cell),
    ("지연 낱말 없음", _p_no_latency_word),
    ("OUT_OF_SCOPE 동거", _p_out_of_scope),
]


def run_checker(root: pathlib.Path) -> tuple[int, str]:
    proc = subprocess.run(
        [sys.executable, str(root / "tools" / "check_values.py")],
        capture_output=True, text=True,
        env={"PATH": "/usr/bin:/bin", "LANG": "C.UTF-8"})
    return proc.returncode, proc.stdout + proc.stderr


# 열거 섭동 — 체류 구성 열거는 값이 아니라 구조라 위 행렬과 형태가 다르다.
#
# 두 delta의 문구가 글자까지 같지는 않다(exit-policy는 "같은 호출이 도는 운영 모드
# 승격 트랜잭션"으로 쓴다). 앵커를 문자열로 박으면 한쪽 파일이 조용히 안 시험된다 —
# 9판의 자체 시험이 정확히 그래서 exit-policy 쪽을 덜 덮었다.
ENUM_ANCHOR = re.compile(
    r"outbox 기록·시도별 실패 기록·게이트 래치·[^\n]*?구조화 로그 줄")

ENUM_PERTURBATIONS: list[tuple[str, str]] = [
    ("항 순서 뒤집기 — 승격이 앞",
     "운영 모드 승격 트랜잭션·outbox 기록"),
    ("두 줄 분할",
     "outbox 기록·시도별 실패 기록·게이트 래치와\n  운영 모드 승격 트랜잭션"),
    ("낱말 치환 — 아웃박스",
     "아웃박스 기록·게이트 래치·운영 모드 승격 트랜잭션"),
    ("의미 반전 — 제외하고",
     "outbox 기록·시도별 실패 기록·게이트 래치·운영 모드 승격 트랜잭션·"
     "그 호출이 쓰는 구조화 로그 줄은 **제외하고**"),
    ("열거 제거",
     "전송 관련 작업"),
]


def axis_two(report: list[str]) -> int:
    bad = 0
    with tempfile.TemporaryDirectory() as tmp:
        base = pathlib.Path(tmp) / "openspec" / "changes"
        base.mkdir(parents=True)
        clean = base / CHANGE_ID
        shutil.copytree(CHANGE, clean,
                        ignore=shutil.ignore_patterns("__pycache__"))

        rc, out = run_checker(clean)
        if rc != 0:
            report.append("축 2 기준선 실패 — 오염하지 않은 사본이 이미 실패한다."
                          "\n" + out)
            return 1
        report.append("축 2 기준선 통과 — 오염하지 않은 사본은 exit 0")

        n = 0
        for seed_name, sent, val, want_prefix, sites in SEEDS:
            for pert_name, pert in PERTURBATIONS:
                text = pert(sent, val)
                if not text:
                    continue
                tok = INJECTED.search(text)
                if tok is None:
                    report.append(f"  **하니스 결함** {seed_name} / {pert_name} — "
                                  f"섭동이 만든 문장에서 수치를 못 뽑았다: {text!r}")
                    bad += 1
                    continue
                want = want_prefix + tok.group(1)
                for site_name in sites:
                    relpath = SITES[site_name]
                    n += 1
                    root = base / f"{CHANGE_ID}-p{n}"
                    shutil.copytree(clean, root)
                    target = root / relpath
                    target.write_text(
                        target.read_text(encoding="utf-8") + "\n" + text + "\n",
                        encoding="utf-8")
                    rc, out = run_checker(root)
                    label = f"{seed_name} / {pert_name} / {site_name}"
                    if rc == 0:
                        report.append(f"  **빠져나감** {label}\n"
                                      f"               {text.strip()[:88]}")
                        bad += 1
                    elif want not in out:
                        report.append(f"  이유가 다르다 {label} — {want!r}가 "
                                      f"출력에 없다")
                        bad += 1
                    shutil.rmtree(root)

        # 열거 섭동. 앵커는 파일마다 문구가 조금씩 달라 정규식으로 찾는다 —
        # 두 delta의 열거가 글자까지 같아야 할 이유는 없고, 같기를 요구하면
        # 이 하니스가 문서 표현을 규정하게 된다.
        enum_runs = 0
        for i, (pert_name, replacement) in enumerate(ENUM_PERTURBATIONS):
            for j, relpath in enumerate(sorted(
                    str(p.relative_to(CHANGE))
                    for p in (CHANGE / "specs").rglob("spec.md"))):
                target_rel = relpath
                text = (clean / target_rel).read_text(encoding="utf-8")
                m = None
                for mm in ENUM_ANCHOR.finditer(text):
                    m = mm                      # 마지막 출현 = Scenario 쪽
                if m is None:
                    report.append(
                        f"  **앵커 없음** 열거 / {target_rel} — 이 파일에서 "
                        f"열거 줄을 못 찾았다. 0매치는 통과가 아니다")
                    bad += 1
                    continue
                root = base / f"{CHANGE_ID}-e{i}-{j}"
                shutil.copytree(clean, root)
                (root / target_rel).write_text(
                    text[:m.start()] + replacement + text[m.end():],
                    encoding="utf-8")
                enum_runs += 1
                rc, out = run_checker(root)
                label = (f"열거 / {pert_name} / "
                         f"{pathlib.Path(target_rel).parent.name}")
                if rc == 0:
                    report.append(f"  **빠져나감** {label}")
                    bad += 1
                elif "체류 구성 열거" not in out:
                    report.append(f"  이유가 다르다 {label}")
                    bad += 1
                shutil.rmtree(root)

        report.append(f"  섭동 {n}칸 + 열거 {enum_runs}칸 실행 (센 값이다)")
    return bad


# ---------------------------------------------------------------------------

def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--axis", type=int, choices=(1, 2), default=None)
    args = ap.parse_args()

    report: list[str] = []
    bad = 0

    if args.axis in (None, 1):
        report.append("## 축 1 — `fail()` 자리 뮤테이션")
        bad += axis_one(report)
        report.append("")
    if args.axis in (None, 2):
        report.append("## 축 2 — 섭동 행렬")
        bad += axis_two(report)
        report.append("")

    # **매 실행마다 못 세는 것을 말한다.** 통과 줄만 있으면 "전부 덮었다"로
    # 읽히고, 그 오독이 7·8·9판을 세 번 연속 통과시켰다.
    report_unseeable(report)

    print("\n".join(report))
    if bad:
        print(f"\n커버리지 게이트 실패 — {bad}건.\n"
              f"미커버 자리는 자체 시험에 케이스를 더하고, 빠져나간 섭동은 "
              f"검사를 고친 뒤 그 섭동을 케이스로 넣는다.")
        return 1
    print("\n커버리지 게이트 통과 — 모든 `fail()` 자리가 케이스에 걸리고 "
          "모든 섭동이 잡힌다.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
