#!/usr/bin/env python3
"""a092 값 단위 교차문서 검사.

왜 이것이 있는가
================

1~6판이 여섯 번 연속 freeze에서 거부됐고, 6라운드가 네 결함의 **형태가 하나**임을
확정했다: **정정이 리뷰어가 가리킨 그 문서에만 착지하고, 같은 값이 사는 다른 문서에는
안 간다.** 정정의 단위가 "값"이 아니라 "리뷰어가 준 file:line"이었기 때문이다.

그래서 7판의 첫 산출물은 문서 편집이 아니라 이 스크립트다. 단위를 값으로 바꾼다.

무엇이 손으로 박혀 있고 무엇이 유도되는가 — 이 구별이 이 스크립트의 설계다
=============================================================================

6판이 만든 `tasks.md` 7.4.2는 기대값 **16**을 손으로 박았고, 6판이 자기 산출물에 절을
하나 더하는 순간 그 숫자가 거짓이 됐다(6라운드 M1/H-4, 실제 출력 17).
**기대값을 박은 검사는 검사가 아니라 또 하나의 주장이다.**

구별은 이렇다.

- **선택한 상수**(T · D · Attempts · budget)는 유도할 수 없다. 사람이 골랐다.
  그러므로 **한 자리에 한 번 선언하고**(아래 `ADOPTED`) 모든 문서를 그것에 맞춘다.
  선언이 하나이므로 이것은 중복이 아니다.
- **셀 수 있는 것**(산출물 수 · 표 행 수 · 비표준 절 수 · 등식의 값)은 **절대
  선언하지 않는다.** 파일 시스템과 문서 자신에서 유도해 서로 대조한다.

검사 목록
=========

1. **산술** — 문서에 적힌 등식을 평가한다. `3 × 1.3s + 2 × 0.2s = 4.2s`는
   상수가 틀린 것보다 먼저 **산술이 안 닫힌다**(6라운드 B1).
2. **예산 등식의 입력** — 결과가 a092의 채택량인 등식은 채택 상수만 쓸 수 있다.
3. **후보 훑기 표** — design D2의 각 후보 행이 자기 산술을 지킨다.
4. **고아 측정치** — proposal·design·tasks·specs가 인용하는 측정 수치는
   `analysis/`에 실린 것이어야 한다. `28.9`가 이 검사에 걸린다(6라운드 B2).
5. **명명값의 유일성** — 같은 양이 문서마다 다른 값으로 적히면 실패한다.
6. **판 번호** — 전 문서가 같은 판이어야 한다(6라운드 M6/M-4).
7. **유도 카운트** — 산출물 수 = 훑기 표 행 수, 비표준 절은 전부 표에 인용이 있어야
   한다. 어느 쪽도 기대값을 박지 않는다.

이 검사가 못 하는 것 — 침묵한 생략 금지
=======================================

- **문장의 참을 검사하지 않는다.** `design.md:314`의 Go 주석이 거짓이었던 것
  (`EntryGate.Block`은 journal에 줄 서지 않는다, 6라운드 H2)은 소스를 읽어야 나온다.
  이 스크립트는 **값의 불일치**만 잡는다.
- **`review.md`는 검사 대상이 아니다.** 그 파일은 판정 기록이고 **기각된 값을
  일부러 그대로 인용한다**(예: 후보 #2의 9.231초). 대상에 넣으면 모든 라운드가 실패한다.
- **5번의 정규식은 열거이지 전칭이 아니다.** 새 양을 문서에 도입하면 여기 추가해야
  한다. 그것을 강제하는 기계는 없다.

실행
====

    python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/check_values.py

저장소를 **읽기만 한다.** 종료 코드 0이면 통과, 1이면 위반 목록을 낸다.
"""

from __future__ import annotations

import pathlib
import re
import sys

CHANGE = pathlib.Path(__file__).resolve().parent.parent
FLM_ROOT = CHANGE / "analysis" / "function-logic"

# ---------------------------------------------------------------------------
# 선언 — 사람이 고른 값. 이 파일이 유일한 선언 자리다.
# ---------------------------------------------------------------------------

DRAFT = 11  # 판 번호

# 기각된 값 — 이것도 사람의 판단이므로 선언한다.
#
# 7라운드 B2: 코퍼스를 `analysis/`의 **산문**에서 긁으면, 어떤 값을 *기각하려고*
# 인용한 문장이 그 값을 코퍼스에 넣어 **영구히 면역시킨다.** 아래 둘이 실제로 그렇게
# 됐다. 그래서 두 겹으로 막는다 — (1) 이 blocklist가 코퍼스와 무관하게 항상 잡고,
# (2) `measurement_corpus()`가 REJECT_MARK가 붙은 줄을 코퍼스에서 뺀다.
# **키는 ms로 정규화한 값이다.** 9판은 `9.231`을 그대로 키로 썼는데, 그러면 같은 양의
# 다른 표기(`9231 ms`)가 조회를 통과한다 — 10라운드 축 2가 그 칸을 잡았다. 값이
# 같으면 표기가 달라도 같은 값이다.
REJECTED_MS: dict[float, str] = {
    28.9: "6라운드 B2 — 어떤 측정 표에도 근거가 없던 왕복 지연",
    9231.0: "5라운드 H1 — 채택하지 않은 후보 #2(T=1.4s·D=200ms)에서 잰 `n.mu` 배수",
}


def rejected_reason(value_ms: float) -> str | None:
    for k, why in REJECTED_MS.items():
        if abs(k - value_ms) < 1e-6:
            return why
    return None

# `analysis/` 안에서 "이 수는 기각됐다"를 말하는 줄에 붙인다. 이 줄의 숫자는
# 측정 코퍼스에 들어가지 않는다 — 기각 문장이 기각된 값을 부활시키지 못하게.
REJECT_MARK = "<!-- rejected-value -->"

# 절 번호는 측정이 아니다 — 코퍼스에서 뺀다(8라운드 M-7). 두 모양이 있다:
# 본문 안의 `§7.5`, 그리고 **`§` 없는 마크다운 제목**(`### 7.5 8판 재측정 …`).
# 후자가 실제로 `7.5`를 코퍼스에 넣어 "체류 7.5s"라는 없는 측정을 통과시켰다.
# 10라운드 축 2: 9판은 `§7.5`와 `### 7.5 …` 두 모양만 지웠다. `7.9 절`, 목록 항목
# `- 8.4 재측정 절차`, 본문의 `7.5절에서`는 그대로 `\d+\.\d+`로 코퍼스에 들어가
# "체류 7.9s" 같은 없는 측정을 통과시켰다. 절 참조의 모양을 전부 센다.
SECTION_NUMBER = re.compile(
    r"§\s*\d+(?:\.\d+)*"                       # §7.5
    r"|\d+(?:\.\d+)+\s*절"                     # 7.9 절 · 7.5절
)
HEADING_NUMBER = re.compile(
    r"^(#{1,6}\s+)\d+(?:\.\d+)*(?=[\s.])"      # ### 7.5 …
    r"|^(\s*[-*]\s+)\d+(?:\.\d+)+(?=[\s.])"    # - 8.4 …
)

ADOPTED_MS = {
    "alertPublishTimeout": 1300.0,
    "alertRetryDelay": 150.0,
    "alertTransportBudget": 4200.0,
    "alertOverheadReserve": 800.0,
    "alertBudget": 5000.0,
}
ADOPTED_ATTEMPTS = 3

# D2 후보 표의 "냉 여유" 열이 나누는 값 — 창 안의 유일한 냉연결 publish 측정.
# 이것은 고른 값이 아니라 잰 값이므로 선언이 아니라 **인용**이다. main()이
# `analysis/`의 코퍼스 안에 실제로 있는지 확인한다 — 측정이 바뀌면 이 줄도 깨진다.
COLD_PUBLISH_S = 0.754

# 등식이 합법적으로 쓸 수 있는 지속시간(ms). 채택 상수 + 정상 등급 1회(=T).
LEGAL_EQ_MS = set(ADOPTED_MS.values())

# 상수라서 고아 검사에서 면제되는 수치 표기(ms 아닌 초 표기 포함).
#
# 채택 상수만 손으로 적는다. **후보 표(D2)의 값은 적지 않는다** — 그것은 표에서
# 유도할 수 있고, 유도할 수 있는 것을 선언하면 표가 바뀔 때 거짓이 된다
# (이 파일 머리말의 구별). `candidate_tokens()` 가 그 일을 한다.
CONSTANT_TOKENS = {
    "1.3", "0.15", "4.2", "4.200", "0.8", "5.0", "5.000", "1300", "150", "800",
    # 같은 상수의 ms 표기 — 9라운드 M-2가 오탐으로 짚었다. `4200 ms`는 고아가 아니다.
    "4200", "5000",
}


def candidate_tokens() -> set[str]:
    """design D2 후보 표의 각 셀에 실제로 적힌 수치 표기를 모은다.

    후보 #2의 reserve 400ms처럼 **기각된 후보의 상수**도 산문에서 논의된다.
    그것은 측정이 아니라 표에서 유도되는 값이므로 고아 검사에서 면제되어야 하고,
    면제 목록은 **표를 읽어서** 만든다. 표에 없는 수는 여기 안 들어온다.

    9판은 reserve 열이 자유 텍스트(`[^|]*?`)였고 셀에서 **모든 수**를 긁었다.
    그래서 표에 행을 하나 더하며 그 셀에 아무 수나 적으면 전역 면제가 됐고,
    무차원 비(냉 여유 `2.65×`)가 지속시간 인용(`2.65s`)까지 면제했다 —
    10라운드 축 2. 이제 **지속시간 셀에서만** 긁는다.
    """
    tokens: set[str] = set()
    design = CHANGE / "design.md"
    if not design.exists():
        return tokens
    for line in design.read_text(encoding="utf-8").splitlines():
        m = SWEEP_ROW.match(line)
        if not m:
            continue
        # 그룹: #, attempts, Timeout, RetryDelay, transport, reserve, 냉 여유.
        # 지속시간 넷만 쓴다 — 냉 여유는 배수라 지속시간 면제 근거가 될 수 없다.
        for cell in m.groups()[2:6]:
            if as_ms(cell or "") is None:
                continue
            for num in re.findall(r"\d+(?:\.\d+)?", cell):
                tokens.add(num)
    return tokens

# 산문이 D3 실측 표의 구성 수를 말하는 자리. 한글 수사도 읽는다 — 9판까지 여섯 곳이
# "여덟 구성"이었고 아라비아 숫자만 보는 검사였다면 그 여섯 곳을 못 봤다.
CONFIG_COUNT = re.compile(
    r"(\d+|하나|둘|셋|넷|다섯|여섯|일곱|여덟|아홉|열|열하나|열둘|열셋)\s*구성")
KO_NUM = {"하나": 1, "둘": 2, "셋": 3, "넷": 4, "다섯": 5, "여섯": 6, "일곱": 7,
          "여덟": 8, "아홉": 9, "열": 10, "열하나": 11, "열둘": 12, "열셋": 13}

STANDARD_SECTIONS = {
    "Inputs and invariants",
    "Branches and early returns",
    "Calls and live bindings",
    "State mutations and fallbacks",
    "Safety conclusion",
}

# 검사 대상. review.md는 위 주석의 이유로 제외한다.
def target_files() -> list[pathlib.Path]:
    # **목록이 아니라 범위로 잡는다 — 11판, 10라운드 차단 3(B27·B28).**
    #
    # 10판은 세 파일 + `specs/**/spec.md` + `analysis/*.md` + FLM만 봤다. 그래서
    # change 루트에 `addendum.md`를 하나 더하고 거기에 **채택 상수를 전부 다르게
    # 적고 기각값까지 인용해도** 통과했다. 새 spec delta 디렉터리도 같았다.
    # 문서를 더하는 것은 판마다 일어나는 일이고, 그때마다 이 목록을 고치는 것을
    # 기억해야 한다면 그것은 검사가 아니라 관습이다.
    #
    # `review.md`만 뺀다 — 리뷰 기록은 **기각값과 우회 문자열을 인용하는 것이 일**
    # 이라 대상으로 삼으면 정의상 실패한다. 뺀다는 사실을 여기 적는 것이
    # 침묵한 면제와 다른 점이고, 그 결과 review.md 안의 값은 미검사로 남는다
    # (10라운드 H5 — 미탐 계열 U6에 적었다).
    skip = {"review.md"}
    files = sorted(f for f in CHANGE.rglob("*.md")
                   if f.name not in skip and "__pycache__" not in f.parts)
    return [f for f in files if f.exists()]


def rel(p: pathlib.Path) -> str:
    return str(p.relative_to(CHANGE.parent.parent.parent))


FAILURES: list[str] = []
NOTES: list[str] = []
CANDIDATE_TOKENS: set[str] = set()


def fail(where: str, msg: str) -> None:
    FAILURES.append(f"{where}\n      {msg}")


# ---------------------------------------------------------------------------
# 지속시간 파싱
# ---------------------------------------------------------------------------

_DUR = re.compile(r"^(\d+(?:\.\d+)?)\s*(ms|밀리초|s|초)$")


def as_ms(token: str) -> float | None:
    """'1.3s' / '150ms' / '1.3초' → ms. 지속시간이 아니면 None."""
    m = _DUR.match(token.strip())
    if not m:
        return None
    value = float(m.group(1))
    return value if m.group(2) in ("ms", "밀리초") else value * 1000.0


def as_scalar(token: str) -> float | None:
    token = token.strip()
    return float(token) if re.fullmatch(r"\d+(?:\.\d+)?", token) else None


# ---------------------------------------------------------------------------
# 검사 1·2 — 산술과 예산 등식의 입력
# ---------------------------------------------------------------------------

# 형태 A:  <라벨> = n × dur + n × dur = dur
EQ_TWO_TERM = re.compile(
    r"([^\n=|]{0,40}?)=\s*"
    r"(\d+)\s*[×*]\s*(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))\s*\+\s*"
    r"(\d+)\s*[×*]\s*(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))\s*=\s*"
    r"\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))"
)

# 형태 B:  <라벨> = dur − dur = dur
EQ_SUB = re.compile(
    r"([^\n=|]{0,40}?)=\s*"
    r"(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))\s*[−-]\s*"
    r"(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))\s*=\s*"
    r"\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))"
)

# 형태 C:  <라벨> = n × dur = dur
EQ_ONE_TERM = re.compile(
    r"([^\n=|]{0,40}?)=\s*"
    r"(\d+)\s*[×*]\s*(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))\s*=\s*"
    r"\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))"
)

BUDGET_LABEL = re.compile(
    r"transport|reserve|alertTransportBudget|alertOverheadReserve|alertBudget|예산"
)

# 7라운드 M-1(a): 라벨로만 걸면 라벨을 안 쓰는 문장이 그대로 빠져나간다.
# 그래서 **결과값으로도** 건다 — 결과가 a092의 채택량인 등식은, 라벨이 무엇이든,
# 채택 상수만 쓸 수 있다. 라벨 검사는 남긴다(둘은 서로 다른 것을 잡는다).
#
# 반증·후보 서술은 이 규칙에서 빠져야 한다. 빠지는 방법은 침묵이 아니라 **표기**다.
#
# 8라운드 B3: 8판은 이것을 산문 낱말 목록(`후보|반증|기각|가정|만약|바꾸면|대안|~~`)으로
# 썼는데, 그 낱말들은 D2 산문에 상시 등장하므로 **면제가 사실상 기본값**이었다.
# `기각 이력을 정리하면, 채택된 transport 예산 = 3 × 1.5s + 2 × 250ms = 5.0s 이다.`
# 한 줄이 통과했다 — 3판이 기각당한 후보 #1 그 자체다.
# 표기는 **낱말이 아니라 마커**여야 한다. `REJECT_MARK`와 같은 이유다.
#
# 9판은 인자 없는 `<!-- counterfactual -->`을 썼는데 10라운드가 구멍 넷을 짚었다:
# (1) 렌더된 문서에서 **안 보인다** — 독자에게는 반증이라는 표시가 없다,
# (2) 출력에 **흔적이 없다**(`REJECT_MARK`는 NOTE를 남긴다) — 침묵한 예외다,
# (3) 마커 하나가 **줄 전체**를 덮어 한 줄에 등식 둘을 놓으면 둘 다 면제된다,
# (4) 실 문서 사용 횟수가 **0**이었다 — 한 번도 안 쓰인 기제였다.
#
# 그래서 마커가 **값을 지목**하게 한다. `<!-- counterfactual: 5.0s -->`는 그 줄의
# 5.0s짜리 등식 하나만 면제하고, 면제했다는 사실이 출력에 남는다.
# 콜론을 **선택으로** 둔다. 필수로 두면 9판의 값 없는 형태(`<!-- counterfactual -->`)가
# 정규식에 아예 안 걸려서, "값을 지목하지 않는다"는 진단이 영원히 안 나온다 —
# 잡히기는 하나 다른 이유로 잡히고, 그러면 표지 문법의 결함이 안 보인다.
COUNTERFACTUAL_MARK = re.compile(r"<!--\s*counterfactual:?\s*([^>]*?)\s*-->")

TOL_MS = 0.5


def claims_adopted(label: str, result_ms: float | None, line: str) -> bool:
    """이 등식이 a092의 **채택량**을 주장하는가.

    두 갈래로 건다.

    - 라벨이 예산 계열이다(7판까지의 규칙).
    - 또는 **결과가 채택량 자체다** — 라벨이 무엇이든. 7라운드 M-1(a)가
      라벨만 보는 규칙을 우회했고, 우회에 필요한 것은 라벨을 안 쓰는 것뿐이었다.

    반증·후보 서술은 표지가 있어야 면제된다. 표지 없는 예외는 침묵한 생략이다.
    표지는 **그 등식의 결과값을 지목**해야 한다 — 줄 단위 면제는 같은 줄의 다른
    등식까지 덮는다(10라운드).
    """
    for mark in COUNTERFACTUAL_MARK.finditer(line):
        named = as_ms(mark.group(1).strip())
        if named is None:
            fail("[표지] counterfactual",
                 f"반증 표지가 값을 지목하지 않는다: {mark.group(0)} — "
                 f"`<!-- counterfactual: 5.0s -->`처럼 그 등식의 결과를 적는다")
            continue
        if result_ms is not None and abs(result_ms - named) < TOL_MS:
            NOTES.append(f"  [반증 표지] 결과 {mark.group(1).strip()}의 등식을 "
                         f"면제한다 — 채택량이 아니라 반증 서술")
            return False
    if BUDGET_LABEL.search(label):
        return True
    return result_ms is not None and any(
        abs(result_ms - v) < TOL_MS for v in ADOPTED_MS.values()
    )


def check_arithmetic(path: pathlib.Path, lines: list[str]) -> None:
    for no, line in enumerate(lines, 1):
        where = f"{rel(path)}:{no}"

        for m in EQ_TWO_TERM.finditer(line):
            label, n1, d1, n2, d2, res = m.groups()
            a, b, r = as_ms(d1), as_ms(d2), as_ms(res)
            if None in (a, b, r):
                continue
            lhs = int(n1) * a + int(n2) * b
            if abs(lhs - r) > TOL_MS:
                fail(where, f"산술이 안 닫힌다: {m.group(0).strip()}  "
                            f"→ 좌변 {lhs/1000:.4g}s ≠ 우변 {r/1000:.4g}s")
            elif claims_adopted(label, r, line):
                bad = [d for d in (d1, d2) if as_ms(d) not in LEGAL_EQ_MS]
                # 시도 수는 **transport 예산**의 등식에만 걸린다. 정상 등급은
                # 정의상 1회이고(`정상 = 1 × 1.3s = 1.3s`), 그 등식의 결과도
                # 채택 상수라 무조건 걸면 옳은 문장이 실패한다.
                if (int(n1) != ADOPTED_ATTEMPTS and r is not None
                        and abs(r - ADOPTED_MS["alertTransportBudget"]) < TOL_MS):
                    bad.append(f"시도 {n1}")
                if bad:
                    fail(where, f"채택량('{label.strip()}')의 등식이 채택하지 않은 값을 "
                                f"쓴다: {', '.join(bad)}  — {m.group(0).strip()}")

        for m in EQ_SUB.finditer(line):
            label, d1, d2, res = m.groups()
            a, b, r = as_ms(d1), as_ms(d2), as_ms(res)
            if None in (a, b, r):
                continue
            if abs((a - b) - r) > TOL_MS:
                fail(where, f"산술이 안 닫힌다: {m.group(0).strip()}  "
                            f"→ 좌변 {(a-b)/1000:.4g}s ≠ 우변 {r/1000:.4g}s")
            elif claims_adopted(label, r, line):
                bad = [d for d in (d1, d2, res) if as_ms(d) not in LEGAL_EQ_MS]
                if bad:
                    fail(where, f"채택량('{label.strip()}')의 등식이 채택하지 않은 값을 "
                                f"쓴다: {', '.join(bad)}  — {m.group(0).strip()}")

        for m in EQ_ONE_TERM.finditer(line):
            if EQ_TWO_TERM.search(line):
                continue
            label, n1, d1, res = m.groups()
            a, r = as_ms(d1), as_ms(res)
            if None in (a, r):
                continue
            if abs(int(n1) * a - r) > TOL_MS:
                fail(where, f"산술이 안 닫힌다: {m.group(0).strip()}  "
                            f"→ 좌변 {int(n1)*a/1000:.4g}s ≠ 우변 {r/1000:.4g}s")
            # 10라운드 축 2 / 9라운드 B-6: 형태 C에는 채택량 검사가 없었다.
            # 7라운드 M-1(a)가 만든 규칙("결과가 채택량이면 라벨이 무엇이든
            # 채택 상수만 쓸 수 있다")이 형태 A·B에만 구현돼 있었고,
            # `alertTransportBudget = 3 × 1.4s = 4.2s` 한 줄이 그대로 통과했다.
            elif claims_adopted(label, r, line):
                bad = [d for d in (d1,) if as_ms(d) not in LEGAL_EQ_MS]
                # 시도 수는 **transport 예산**의 등식에만 걸린다. 정상 등급은
                # 정의상 1회이고(`정상 = 1 × 1.3s = 1.3s`), 그 등식의 결과도
                # 채택 상수라 무조건 걸면 옳은 문장이 실패한다.
                if (int(n1) != ADOPTED_ATTEMPTS and r is not None
                        and abs(r - ADOPTED_MS["alertTransportBudget"]) < TOL_MS):
                    bad.append(f"시도 {n1}")
                if bad:
                    fail(where, f"채택량('{label.strip()}')의 등식이 채택하지 않은 "
                                f"값을 쓴다: {', '.join(bad)}  — {m.group(0).strip()}")


# ---------------------------------------------------------------------------
# 검사 3 — design D2 후보 훑기 표의 각 행이 자기 산술을 지킨다
# ---------------------------------------------------------------------------

SWEEP_ROW = re.compile(
    r"^\|\s*\*{0,2}(\d+)[^|]*\|\s*\*{0,2}(\d+)\*{0,2}\s*\|"       # #, Attempts
    r"\s*\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|s|초))\*{0,2}\s*\|"        # Timeout
    r"\s*\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|s|초))\*{0,2}\s*\|"        # RetryDelay
    r"\s*\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|s|초))\*{0,2}\s*\|"        # transport
    # reserve — **값이 먼저**, 주석은 뒤에 자유롭게. 9판은 셀 전체가 자유 텍스트라
    # 아무 수나 면제됐고(9라운드 B-7), 값만 받게 좁히자 `**0 — 채택 불가**` 행이
    # 통째로 검사 밖으로 나갔다. 값은 엄격히, 꼬리말은 관대히.
    r"\s*\*{0,2}(\d+(?:\.\d+)?\s*(?:ms|s|초))\*{0,2}[^|]*\|"      # reserve
    r"\s*\*{0,2}(\d+\.\d+)\s*×[^|]*\|"                            # 냉 여유
)

# 위 정규식이 **놓치는** 후보 행을 찾기 위한 느슨한 짝. 10라운드 축 2가 세 가지
# 모양 변형으로 행을 검사 밖으로 빼냈다 — `×`를 `배`로, 배수를 정수로, reserve 셀에
# 꼬리말 붙이기. 행이 조용히 사라지면 그 행의 산술은 아무도 안 본다.
#
# 9판에는 표 전체의 0매치 가드만 있었고 **행 단위 가드가 없었다.** 7행이 6행이 돼도
# 아무 말이 없었다. 느슨한 수와 엄격한 수를 대조하는 것이 그 가드다.
SWEEP_ROW_LOOSE = re.compile(r"^\|\s*\*{0,2}\d+[^|]*\|(?:[^|]*\|){5,}")


def _tables(lines: list[str]) -> list[tuple[int, list[tuple[int, str]]]]:
    """연속하는 `|` 줄을 한 표로 묶는다. (첫 줄 번호, [(줄 번호, 줄)])."""
    out: list[tuple[int, list[tuple[int, str]]]] = []
    buf: list[tuple[int, str]] = []
    for no, line in enumerate(lines, 1):
        if line.startswith("|"):
            buf.append((no, line))
        elif buf:
            out.append((buf[0][0], buf))
            buf = []
    if buf:
        out.append((buf[0][0], buf))
    return out


def check_candidate_sweep(path: pathlib.Path, lines: list[str]) -> int:
    rows = 0
    # 느슨한 짝으로 "후보 표 안인데 엄격 정규식에 안 맞는 행"을 찾는다.
    #
    # **표를 먼저 나누고 나서 본다.** 9판의 첫 시도는 "엄격 매치가 한 번이라도 난
    # 뒤부터"였는데, 그러면 **표의 첫 행**이 빠져나갈 때 가드가 안 걸린다 —
    # 실제로 후보 #1 행(`0 — 채택 불가`)이 그렇게 조용히 사라졌다.
    loose_missed: list[int] = []
    for _, block in _tables(lines):
        if not any(SWEEP_ROW.match(l) for _, l in block):
            continue                     # 후보 표가 아니다
        for no, l in block:
            if SWEEP_ROW_LOOSE.match(l) and not SWEEP_ROW.match(l):
                loose_missed.append(no)

    for no, line in enumerate(lines, 1):
        m = SWEEP_ROW.match(line)
        if not m:
            continue
        idx, attempts, timeout, delay, transport, reserve, cold = m.groups()
        t, d, tr = as_ms(timeout), as_ms(delay), as_ms(transport)
        if None in (t, d, tr):
            continue
        rows += 1
        where = f"{rel(path)}:{no}"
        want = int(attempts) * t + (int(attempts) - 1) * d
        if abs(want - tr) > TOL_MS:
            fail(where, f"후보 #{idx} 행의 transport가 자기 입력과 다르다: "
                        f"{attempts}×{timeout} + {int(attempts)-1}×{delay} = "
                        f"{want/1000:.4g}s, 표는 {tr/1000:.4g}s")
        r = as_ms(reserve.strip()) if as_ms(reserve.strip()) is not None else None
        if r is not None and abs((ADOPTED_MS["alertBudget"] - tr) - r) > TOL_MS:
            fail(where, f"후보 #{idx} 행의 reserve가 budget − transport와 다르다: "
                        f"5s − {tr/1000:.4g}s = "
                        f"{(ADOPTED_MS['alertBudget']-tr)/1000:.4g}s, 표는 {reserve.strip()}")
        # 냉 여유 — 7라운드 M-1(b). 이 열이 이 change의 **유일한 판단**(1회 상한을
        # 냉 측정 대비 얼마나 남길 것인가)인데 6·7판의 검사는 이 열을 안 봤다.
        want_cold = round(t / 1000.0 / COLD_PUBLISH_S, 2)
        if abs(want_cold - float(cold)) > 0.005:
            fail(where, f"후보 #{idx} 행의 냉 여유가 Timeout ÷ 냉 측정과 다르다: "
                        f"{t/1000:.4g}s ÷ {COLD_PUBLISH_S}s = {want_cold}×, "
                        f"표는 {cold}×")
    if loose_missed:
        fail("[검사 3] 후보 표에서 빠져나간 행",
             f"{len(loose_missed)}행이 후보 표 안에 있는데 엄격 정규식에 안 맞는다 "
             f"(줄 {', '.join(str(n) for n in loose_missed)}) — 그 행의 산술은 "
             f"아무도 안 본다. 모양이 바뀌었으면 SWEEP_ROW를 고친다")
    return rows


# ---------------------------------------------------------------------------
# 검사 4 — 고아 측정치
# ---------------------------------------------------------------------------

# 뒤 경계는 `\b`가 아니다 — 한국어 조사가 붙으면(`28.9ms였다`) `s`와 `였` 사이에
# 단어 경계가 없어서 `\b`가 실패한다. 7라운드 M-1(c)가 그 우회를 실증했다.
# 필요한 것은 "단위 뒤에 영숫자가 더 붙지 않는다"이지 "단어가 끝난다"가 아니다.
#
# **왼쪽은 부정 룩비하인드로 막지 않는다** — 8라운드 B2. `(?<![≥≤<>]\s)`는 면제가
# 아니라 **첫 자리 삭제**였다: 룩비하인드가 첫 자리에서 실패하면 엔진이 다음 위치에서
# 재시도하므로 `≥ 28.9 ms`가 `8.9`로 잡혔다. 기각값이 그대로 빠져나가고, 잡히는
# 경우에도 진단이 **원문에 없는 수**를 출력했다. 왼쪽에 필요한 것은 "숫자 중간에서
# 시작하지 않는다"뿐이고, 부등호 판정은 매치 **뒤에** 왼쪽 문맥을 보고 한다.
# **오른쪽 경계에서 마침표를 빼면 안 된다** — 10라운드 축 2. 9판의
# `(?![A-Za-z0-9.])`는 `- 초과분 최악: 28.9 ms.` 를 **매치 0개**로 만들었다:
# 문장 끝 마침표가 경계를 깨고, 새로 넣은 왼쪽 앵커가 `8.9`로의 재시작을 막았기
# 때문이다. 그래서 기각값 조회도 고아 검사도 그 줄을 **아예 못 봤다.**
# 이 표기는 이 코퍼스의 native 형태다 — `design.md`의 `→ reserve ≥ 712ms.` 처럼.
#
# 필요한 것은 "단위 뒤에 수가 이어지지 않는다"이지 "마침표가 없다"가 아니다.
_UNIT_END = r"(?![A-Za-z0-9])(?!\.\d)"
# 자릿수 구분 쉼표를 수의 일부로 읽는다. 9판은 `4,777ms`의 `777`만 매치해
# **원문에 없는 수**를 진단으로 냈다(8라운드 B2가 이름 붙인 병의 재발).
_INT = r"\d{1,3}(?:[,  ]\d{3})+|\d+"
# **단위의 철자를 넓힌다 — 10라운드 차단 3(B08·B09·B22).**
#
# 10판의 알파벳은 `ms|밀리초|s|초`뿐이었고 `_UNIT_END`가 `(?![A-Za-z0-9])`라
# `9.231 sec`은 `s` 뒤의 `e`에서 매치가 **깨졌다**. 깨진 것은 "잡지 못했다"가
# 아니라 **아예 못 봤다**이다 — 기각값 조회도 고아 검사도 그 줄을 안 봤다.
# 긴 철자를 먼저 놓아야 `msec`이 `ms`로 잘리지 않는다.
_MS_UNITS = r"msec|ms|밀리세컨드|밀리초"
_S_UNITS = r"seconds|secs|sec|s|초"
_US_UNITS = r"µs|μs|us|마이크로초"
_ANY_UNIT = "|".join((_MS_UNITS, _US_UNITS, _S_UNITS))

MEASURE = re.compile(r"(?<![\d.])(\d+\.\d+)\s*(" + _ANY_UNIT + r")" + _UNIT_END)

# 정수 측정치도 본다(8라운드 M-7). 단위는 ms 계열로 한정한다 — 초 단위 정수는
# 이 문서군에서 거의 전부 설계 상수(10s·34s·5초·60초·90초)이고, 측정치는 ms로 적힌다.
MEASURE_INT = re.compile(
    r"(?<![\d.,])(" + _INT + r")\s*(" + _MS_UNITS + r")" + _UNIT_END)

# 기각값 조회 전용 — 정수 초도 본다. `9.231 s`를 `9231 ms`로 다시 쓰는 우회
# (10라운드 축 2)를 막으려면 단위를 정규화해야 하고, 정규화하려면 초도 읽어야 한다.
MEASURE_ANY = re.compile(
    r"(?<![\d.,])(" + _INT + r"(?:\.\d+)?)\s*(" + _ANY_UNIT + r")" + _UNIT_END)


# **보이지 않는 문자와 합자 단위를 접는다 — 10라운드 차단 3(B10·B11·B12).**
#
# `28.9​ms`(제로폭 공백)와 `28.9 ㎳`(U+33B3)는 **렌더된 문서에서 `28.9 ms`와
# 구별되지 않는다.** 읽는 사람에게 같아 보이는 두 문자열이 검사에게 다르면,
# 그 차이는 검사의 것이지 문서의 것이 아니다. 스캔 전에 한 번 접는다.
_INVISIBLE = dict.fromkeys(
    (0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF, 0x00AD), None)
_LIGATURE = {"㎳": "ms", "㎲": "us", "㎱": "ns"}


def fold_notation(line: str) -> str:
    """같아 **보이는** 표기를 같게 만든다. 스캔하는 모든 줄이 이것을 거친다."""
    line = line.translate(_INVISIBLE)
    for lig, plain in _LIGATURE.items():
        line = line.replace(lig, plain)
    # 소수점 쉼표(`28,9 ms`). 자릿수 구분은 **정확히 세 자리**를 묶으므로
    # 그 규칙이 둘을 가른다: `9,231`은 자릿수 구분이고 `28,9`는 소수점이다.
    return re.sub(r"(?<=\d),(\d{1,2})(?!\d)", r".\1", line)


def norm_num(value: str) -> str:
    """자릿수 구분(쉼표·공백·비분리 공백)을 뗀 표기. 비교와 변환은 전부 이것을 거친다.

    **한 곳으로 모으는 이유**: 10판이 자릿수 구분을 읽게 만들면서 `float()`를
    부르는 자리 하나를 안 고쳐 `9,231`에서 크래시가 났다. 검사가 죽는 것은
    통과도 실패도 아니라 **검사가 없는 것**이다.
    """
    return value.replace(",", "").replace("\u00a0", "").replace(" ", "")


_MS_SET = frozenset(("ms", "msec", "밀리초", "밀리세컨드"))
_US_SET = frozenset(("µs", "μs", "us", "마이크로초"))


def token_ms(value: str, unit: str) -> float:
    """표기를 ms로 옮긴다. **철자를 넓혔으면 환산도 넓혀야 한다** — 안 그러면
    `9.231 sec`을 읽고도 9.231ms로 재서 기각값 조회가 조용히 빗나간다."""
    n = float(norm_num(value))
    if unit in _MS_SET:
        return n
    if unit in _US_SET:
        return n / 1000.0
    return n * 1000.0

CITING = ("proposal.md", "design.md", "tasks.md", "spec.md")

# 9판까지 이 검사에는 게이트가 셋 있었다 — `LATENCY_WORDS`가 **있어야** 검사하고,
# `OUT_OF_SCOPE`가 **없어야** 검사하고, 부등호 앞이면 건너뛰었다. 10라운드 축 2가
# 셋 다 우회를 실증했다. 형태가 하나다: **게이트의 열쇠를 문장 쓰는 사람이 쥐고
# 있다.** 낱말을 빼면(지연 낱말) 검사가 꺼지고, 낱말을 더하면(`timeout`) 검사가
# 꺼지고, `≥`를 붙이면 검사가 꺼진다.
#
# 그래서 게이트를 없앤다. 없애고 재 보니 걸리는 줄은 **14개**였고 전부 정당했다 —
# 반증 산술의 중간값·테스트 실행 시간·후보 재배분. 즉 게이트는 이 14개를 위해
# **모든 측정치의 검사를 끄고 있었다.**
#
# 대신 값이 설명되는 길을 넷으로 정한다. 앞의 셋은 유도이고 넷째만 선언이다.
#
#   1. `analysis/`의 코퍼스에 있다               (측정이다)
#   2. 채택 상수이거나 후보 표의 값이다           (표에서 유도)
#   3. **같은 줄의 등식이 그 값을 낸다**          (산술에서 유도)
#   4. `<!-- not-a-measurement: 사유 -->` 표지    (선언 — 사유를 적는다)
#
# 3번이 게이트를 대신한다. `3 × 3s + 2 × 150ms = 9.3s`의 `9.3`은 측정이 아니라
# 그 줄이 스스로 유도한 값이고, 그것은 낱말이 아니라 **산술**로 판정할 수 있다.
NOT_MEASUREMENT = re.compile(r"<!--\s*not-a-measurement:\s*([^>]*?)-->")


def derived_on_line(line: str) -> set[float]:
    """이 줄의 등식들이 스스로 내는 값(ms). 산술로 유도되므로 측정이 아니다.

    좌변의 항과 우변의 결과를 모두 넣는다 — `3 × 1.0s + 2 × 150ms = 3.3s`에서
    `1.0`·`150`·`3.3` 전부가 그 줄이 설명하는 값이다.
    """
    out: set[float] = set()
    for pat in (EXPR_TWO, EXPR_SUB, EXPR_ONE):
        for m in pat.finditer(line):
            for tok in m.groups():
                v = as_ms(tok or "")
                if v is not None:
                    out.add(v)
    return out


# 검사 1·2의 등식은 **라벨을 요구한다**(`라벨 = …`). 산문은 라벨 없이 쓴다 —
# `누가 1.3s를 3s로 바꾸면 transport가 3 × 3s + 2 × 150ms = 9.3s가 되어`.
# 유도 판정은 라벨과 무관하므로 라벨 없는 짝을 따로 둔다.
_D = r"(\d+(?:\.\d+)?\s*(?:ms|밀리초|s|초))"
EXPR_TWO = re.compile(
    r"(?:\d+)\s*[×*]\s*" + _D + r"\s*\+\s*(?:\d+)\s*[×*]\s*" + _D +
    r"\s*=\s*\*{0,2}" + _D)
EXPR_ONE = re.compile(r"(?:\d+)\s*[×*]\s*" + _D + r"\s*=\s*\*{0,2}" + _D)
EXPR_SUB = re.compile(_D + r"\s*[−-]\s*" + _D + r"\s*=\s*\*{0,2}" + _D)


def measurement_corpus() -> list[float]:
    """`analysis/`가 실제로 **싣고 있는** 수의 집합.

    7라운드 B2: 이 함수가 줄을 가리지 않고 긁었기 때문에, 어떤 값을 *기각하려고*
    쓴 문장이 그 값을 코퍼스에 넣었다. `REJECT_MARK`가 붙은 줄은 여기 안 들어온다.
    표에서만 긁지 않는 이유는 실제 측정치의 일부가 산문에 있기 때문이고
    (예: `n.mu` 배수 8.458초), 그것을 표로 옮기는 것은 이 스크립트의 일이 아니다.
    """
    corpus: list[float] = []
    sources = list(sorted((CHANGE / "analysis").glob("*.md")))
    sources += list(sorted(FLM_ROOT.rglob("*.md")))
    for f in sources:
        for line in f.read_text(encoding="utf-8").splitlines():
            line = fold_notation(line)
            if REJECT_MARK in line:
                continue
            # 절 번호는 측정이 아니다(8라운드 M-7). `§7.5`가 코퍼스에 `7.5`로
            # 들어가면 "체류 7.5s" 같은 없는 측정이 인용될 수 있다.
            line = SECTION_NUMBER.sub(" ", line)
            # 두 가지 제목 모양 중 어느 쪽이 맞았는지에 따라 살릴 접두가 다르다.
            # 참여하지 않은 그룹을 `\1`로 되쓰면 예외가 나므로 함수로 쓴다.
            line = HEADING_NUMBER.sub(
                lambda m: m.group(1) or m.group(2) or "", line)
            for m in re.finditer(r"\d+\.\d+", line):
                corpus.append(float(m.group(0)))
            # 정수 측정치는 **단위를 달고 있을 때만** 코퍼스에 넣는다. 맨 정수를
            # 다 넣으면 로그 줄 번호(6866·6896…)까지 들어와 검사가 무의미해진다.
            for m in MEASURE_INT.finditer(line):
                corpus.append(float(norm_num(m.group(1))))
    return corpus


def cited_ok(tok: str, corpus: list[float]) -> bool:
    """반올림도 잘라 쓰기도 허용한다 — 0.2819→0.282, 4.234955→4.234.

    정수 토큰(8라운드 M-7의 `4777 ms`)도 받는다 — 소수 자리 0으로 다룬다.
    """
    tok = norm_num(tok)
    digits = len(tok.split(".")[1]) if "." in tok else 0
    want = float(tok)
    scale = 10 ** digits
    for value in corpus:
        if abs(round(value, digits) - want) < 1e-9:
            return True
        if abs(int(value * scale) / scale - want) < 1e-9:
            return True
    return False


def check_rejected_values(path: pathlib.Path, lines: list[str]) -> None:
    """기각값 조회 — **어떤 게이트보다도 먼저, 모든 줄에서.**

    8라운드 M-4: 8판은 이 조회를 `check_orphan_measurements` 안에 두었고, 그
    함수는 `LATENCY_WORDS`가 없거나 `OUT_OF_SCOPE`가 있으면 줄을 통째로
    건너뛴다. 그래서 `운영자가 관측한 값은 28.9 ms다`도 `timeout 설정 탓에
    9.231초였다`도 통과했다. **`review.md` 8.1이 "코퍼스와 무관하게 항상 잡고"라고
    쓴 문장은 그 시점에 거짓이었다.** 이제 별도 함수이고 게이트 밖이다.

    부등호 앞이든 뒤든 상관없다 — 기각된 값은 하한으로도 인용할 수 없다.

    **모든 파일에서 돈다** — 10라운드 축 2. 9판은 `CITING` 네 종류로 제한했는데,
    `analysis/`와 FLM은 **코퍼스 원천**이다. 표지 없는 기각값이 거기 있으면
    blocklist를 피하면서 **동시에 코퍼스로 재진입**한다 — 7라운드 B2가 이름 붙인
    세탁이 파일 종류만 바꿔 되살아난다. 이 docstring이 "모든 줄에서"라고 쓴 것을
    이제 파일에 대해서도 참으로 만든다.

    단위는 **정규화해서** 본다 — `9.231 s`와 `9231 ms`는 같은 값이다.
    """
    for no, line in enumerate(lines, 1):
        for m in MEASURE_ANY.finditer(line):
            tok, unit = m.group(1), m.group(2)
            why = rejected_reason(token_ms(tok, unit))
            if why is None:
                continue
            if REJECT_MARK in line:
                # 기각을 **기록하는** 문장은 허용한다. 다만 조용히는 아니다 —
                # 표지를 달아야 하고, 달면 출력에 남는다. 침묵한 예외 금지.
                NOTES.append(f"  [기각 인용] {rel(path)}:{no} — {tok} ({why})")
            else:
                # 진단은 **원문에 있는 표기**를 가리켜야 한다. 정규화한 값을
                # 출력하면 리뷰어가 문서에 없는 수를 찾게 된다(8라운드 B2).
                fail(f"{rel(path)}:{no}",
                     f"기각된 값 {tok}{unit} — {why}. 근거로 쓸 수 없다. "
                     f"기각을 기록하는 문장이면 그 줄에 {REJECT_MARK} 를 단다")


def check_orphan_measurements(path: pathlib.Path, lines: list[str],
                              corpus: list[float]) -> None:
    if path.name not in CITING:
        return
    for no, line in enumerate(lines, 1):
        if SWEEP_ROW.match(line):          # 후보 표는 검사 3이 산술로 본다
            continue
        exempt = NOT_MEASUREMENT.search(line)
        if exempt:
            why = exempt.group(1).strip()
            # **빈 사유는 사유가 아니다 — 10라운드 M4(B24).** 10판은
            # `<!-- not-a-measurement: -->` 를 받아 사유 칸이 빈 채로 통과시켰고,
            # 출력에도 빈 칸이 남았다. 설계가 금지한 "침묵한 예외"가 문법적으로
            # 허용된 것이다. 표지의 값은 표지가 있다는 사실이 아니라 **사유**다.
            if len(why) < 8:
                fail(f"{rel(path)}:{no}",
                     f"not-a-measurement 표지의 사유가 비었거나 너무 짧다: {why!r} — "
                     f"이 값이 왜 측정이 아닌지를 적는다. 표지만 달고 사유를 비우면 "
                     f"그것은 선언한 예외가 아니라 침묵한 예외다")
                continue
            # 선언한 예외는 조용하지 않다 — 표지가 있어야 하고, 달면 출력에 남는다.
            NOTES.append(f"  [측정 아님] {rel(path)}:{no} — {why}")
            continue
        derived = derived_on_line(line)
        for pat in (MEASURE, MEASURE_INT):
            for m in pat.finditer(line):
                tok, unit = m.group(1), m.group(2)
                ms = token_ms(tok, unit)
                if rejected_reason(ms) is not None:
                    continue               # check_rejected_values 가 이미 봤다
                if any(abs(ms - d) < TOL_MS for d in derived):
                    continue               # 이 줄의 등식이 스스로 내는 값
                if (norm_num(tok) in CONSTANT_TOKENS
                        or norm_num(tok) in CANDIDATE_TOKENS
                        or cited_ok(tok, corpus)):
                    continue
                fail(f"{rel(path)}:{no}",
                     f"고아 측정치 {tok} — analysis/ 어디에도 없고 이 줄의 "
                     f"등식으로도 유도되지 않는다. 측정이면 측정 표에 싣고, "
                     f"측정이 아니면 그 줄에 "
                     f"<!-- not-a-measurement: 사유 --> 를 단다")


# ---------------------------------------------------------------------------
# 검사 4.5 — 체류 구성 열거의 일관성 (8라운드 B1)
# ---------------------------------------------------------------------------

# 8판은 두 delta의 **요구 문단**에 "구조화 로그 줄"을 더하고 "이 열거는 완전해야
# 한다(SHALL)"까지 새로 걸었는데, **같은 파일의 Scenario는 안 고쳤다.** 정정이
# 리뷰어가 가리킨 자리에만 착지하는 형태가, 완전성을 SHALL로 선언한 그 파일 안에서
# 재발한 것이다.
#
# 9판의 트리거는 `outbox[^\n]*승격` 하나였다. 10라운드 축 2가 그 트리거를 네 가지로
# 우회했다 — **항 순서를 뒤집으면**(`승격`이 앞) 매치가 사라지고, 두 줄로 쪼개도,
# `아웃박스`로 바꿔도, 아예 지워도 그 줄이 검사 대상에서 **사라진다.** 트리거가
# 사라지면 "빠진 항"을 셀 대상 자체가 없어진다는 것이 이 설계의 결함이었다.
#
# 그래서 트리거를 **개수**로 바꾼다: 다섯 항 중 **둘 이상**을 말하는 줄은 체류 구성을
# 열거하는 줄이고, 그러면 다섯을 전부 말해야 한다. 순서에 무관하고, 두 줄로 쪼개도
# 앞줄이 걸리고, 별칭도 같은 항으로 센다.
# 별칭은 **좁아야 한다.** `게이트`를 `진입 차단`으로, `로그 줄`을 `로그`로 넓혔더니
# 소진의 **결과** 열거(`outbox 미전달 보존·신규 진입 차단·… 일반은 로그 한 줄`)가
# 체류 구성으로 잡혔다. 넓은 별칭은 오탐을 만들고, 오탐은 검사를 끄게 만든다.
DWELL_TERMS: tuple[tuple[str, ...], ...] = (
    ("outbox", "아웃박스", "EnqueueAlert"),
    ("시도별 실패 기록", "시도별 실패", "MarkAlertAttemptFailed"),
    ("게이트 래치", "Gate.Block", "gate latch"),
    ("승격 트랜잭션", "EscalateOperatingMode"),
    ("로그 줄", "구조화 로그", "structured log"),
)
# 둘로 잡으면 **결과 보존 열거**(`outbox 미전달 보존·신규 진입 차단·운영 모드 강화`)가
# 걸린다. 그것은 체류 구성이 아니라 소진의 **결과**이고 항이 셋뿐인 다른 열거다.
# 실측으로 셋이 두 열거를 가른다.
DWELL_TRIGGER_MIN = 3

# 낱말이 다 있어도 의미가 뒤집히면 열거가 아니다 — 10라운드 축 2의 '의미 반전'.
DWELL_NEGATION = re.compile(r"제외하고|제외한|빼고|아닌 것|들지 않는다")

# 이 열거를 말하는 문서. 9판은 `spec.md`만 봤고, 그래서 **유도 정본(design D2)과
# 제안서와 측정 정본이 네 항으로 남은 것을 못 봤다**(9라운드 차단 2).
DWELL_DOCS = ("spec.md", "design.md", "proposal.md", "delivery-latency.md")

# **0매치 가드를 거는 파일은 따로 정한다.** 이름만으로 걸면 이 change에 무관한
# delta(예: observability)를 하나 더할 때 그 파일이 체류 열거를 안 담았다는 이유로
# 실패한다 — 9라운드 M-3이 오탐으로 짚었다. 열거를 담아야 하는 자리는 유도할 수
# 없으므로 선언하되, 선언한 자리가 비면 실패한다(침묵한 면제 금지).
DWELL_REQUIRED = (
    "specs/engine-safety/spec.md", "specs/exit-policy/spec.md",
    "design.md", "proposal.md", "analysis/delivery-latency.md",
)

# **요구 열거가 사는 자리를 이름으로 선언한다 — 11판, 10라운드 차단 3(B14~B16).**
#
# 10판은 문단이 `DWELL_TRIGGER_MIN`(3항) 이상을 담을 때만 그 문단을 열거로 봤다.
# 그래서 다섯 항을 **두 항으로 줄이면** 트리거 미만이라 문단이 통째로 안 보였고,
# 같은 파일의 다른 문단(Scenario)이 `seen > 0`을 채워 0매치 가드도 안 걸렸다.
# **9라운드 차단 2와 8라운드 B1이 글자 그대로 재현되는데 검사는 초록이었다.**
#
# 트리거는 **발견**을 위한 것이고 발견은 축소에 약하다 — 항을 지울수록 안 보인다.
# 그래서 자리를 **선언**한다. 아래 앵커를 담은 문단은 트리거와 무관하게 완전한
# 열거여야 한다. 앵커를 지우면 0매치로 실패하므로 지워서 피할 수도 없다.
# Scenario 쪽에 이미 쓰던 규칙(`DWELL_SCENARIO_TITLES`)을 요구 문단에도 건 것이다.
DWELL_ANCHORS: dict[str, str] = {
    "specs/engine-safety/spec.md":
        '"자기 차례"는 전송에 쓴 시간과 같은 호출 안의 비전송 작업',
    "specs/exit-policy/spec.md":
        "그 시간은 원격 전송에 쓴 시간만이 아니라",
    "design.md":
        "`EnqueueAlert` 1회(outbox 기록) + `MarkAlertAttemptFailed`",
    "proposal.md":
        "같은 `Notify` 호출 안에 전송 말고도",
    "analysis/delivery-latency.md":
        "초과분은 `EnqueueAlert` 1회(outbox 기록)",
}

# 9판은 열거를 **지우면** 잡지 못했다 — 지워진 줄은 트리거가 없기 때문이다.
# 낱말 개수로는 원리적으로 못 잡는다. 그래서 Scenario 쪽은 **자리를 이름으로 선언**한다.
#
# 이것은 선언이지만 침묵한 선언이 아니다: 제목이 파일에서 사라져도 실패한다
# (0매치는 통과가 아니다). 열거를 지우든 Scenario를 지우든 이름을 바꾸든 소리가 난다.
SCENARIO_HEAD = re.compile(r"^#{2,6}\s*Scenario:\s*(.+?)\s*$")
DWELL_SCENARIO_TITLES: dict[str, str] = {
    "한 알림의 체류가 exit 관측 주기를 넘지 않는다": "engine-safety",
    "응답하지 않는 알림 전송 중의 관측": "exit-policy",
}


def _paragraphs(lines: list[str]) -> list[tuple[int, str]]:
    """빈 줄로 나눈 문단. (첫 줄 번호, 본문)."""
    out: list[tuple[int, str]] = []
    start, buf = 0, []
    for no, line in enumerate(lines, 1):
        if line.strip():
            if not buf:
                start = no
            buf.append(line)
        elif buf:
            out.append((start, "\n".join(buf)))
            buf = []
    if buf:
        out.append((start, "\n".join(buf)))
    return out


def _dwell_hits(text: str) -> tuple[int, list[str]]:
    hit, missing = 0, []
    for aliases in DWELL_TERMS:
        if any(a in text for a in aliases):
            hit += 1
        else:
            missing.append(aliases[0])
    return hit, missing


def _dwell_fail(where: str, missing: list[str]) -> None:
    fail(where,
         f"체류 구성 열거에 빠진 항이 있다: {', '.join(missing)} — "
         f"두 delta가 '열거는 소진 경로에 대해 완전해야 한다'를 SHALL로 걸었다. "
         f"요구 문단·Scenario·유도 정본·제안서·측정 정본이 같은 열거를 말해야 한다")


def check_dwell_enumeration(path: pathlib.Path, lines: list[str]) -> None:
    if path.name not in DWELL_DOCS:
        return
    # **문단 단위로 본다.** 줄 단위로 보면 여러 줄에 걸친 열거를 원리적으로 못
    # 읽는다 — proposal과 측정 정본의 열거가 실제로 두세 줄에 걸쳐 있어서, 9판의
    # 줄 단위 검사는 그 둘을 **한 번도 안 봤다**(9라운드 차단 2가 거기 살아 있었다).
    # 완전한 열거를 두 줄로 쪼개는 것은 위반이 아니고, 쪼갠 채로 항이 빠진 것이
    # 위반이다. 문단이 그 둘을 가르는 단위다.
    seen = 0
    for start, para in _paragraphs(lines):
        hit, missing = _dwell_hits(para)
        if hit < DWELL_TRIGGER_MIN:
            continue
        seen += 1
        where = f"{rel(path)}:{start}"
        if missing:
            _dwell_fail(where, missing)
        elif DWELL_NEGATION.search(para):
            fail(where,
                 "체류 구성 열거의 낱말은 다 있으나 의미가 뒤집혀 있다 — "
                 "이 문단은 열거가 아니라 배제다")

    # Scenario 쪽은 이름으로 건다 — 열거를 통째로 지워도 이름은 남는다.
    if path.name == "spec.md":
        want = {t for t, owner in DWELL_SCENARIO_TITLES.items()
                if owner == path.parent.name}
        found: set[str] = set()
        in_scen, head_no, title, body = False, 0, "", []

        def _close() -> None:
            if in_scen and title in want:
                found.add(title)
                _, miss = _dwell_hits("\n".join(body))
                if miss:
                    _dwell_fail(f"{rel(path)}:{head_no}", miss)

        for no, line in enumerate(lines, 1):
            m = SCENARIO_HEAD.match(line)
            if m:
                _close()
                in_scen, head_no, title, body = True, no, m.group(1), []
            elif in_scen and line.startswith("#"):
                _close()
                in_scen = False
            elif in_scen:
                body.append(line)
        _close()
        for missing_title in sorted(want - found):
            fail(f"[열거] {rel(path)}",
                 f"열거를 담아야 하는 Scenario를 못 찾았다: "
                 f"'{missing_title}' — 이름이 바뀌었으면 "
                 f"DWELL_SCENARIO_TITLES를 고친다. 0매치는 통과가 아니다")

    # **선언한 자리는 트리거와 무관하게 완전해야 한다** (11판 — 10라운드 차단 3).
    # 위의 트리거 루프는 항이 줄수록 눈이 어두워진다. 이 블록은 그 반대다:
    # 자리를 먼저 정하고 그 자리가 완전한지만 묻는다.
    anchor = DWELL_ANCHORS.get(str(path.relative_to(CHANGE)))
    if anchor:
        hit_para = [(start, para) for start, para in _paragraphs(lines)
                    if anchor in para]
        if not hit_para:
            fail(f"[열거] {rel(path)}",
                 f"요구 열거가 사는 문단을 못 찾았다: {anchor!r} — 문서가 바뀌었으면 "
                 f"DWELL_ANCHORS를 고친다. 0매치는 통과가 아니다")
        for start, para in hit_para:
            _, miss = _dwell_hits(para)
            if miss:
                _dwell_fail(f"{rel(path)}:{start}", miss)

    if seen == 0 and str(path.relative_to(CHANGE)) in DWELL_REQUIRED:
        fail(f"[열거] {rel(path)}",
             "체류 구성을 세는 문단을 하나도 못 찾았다 — 0매치는 통과가 아니다")


# ---------------------------------------------------------------------------
# 검사 5 — 명명값의 유일성
# ---------------------------------------------------------------------------

NAMED_QUANTITIES: list[tuple[str, list[str]]] = [
    ("전 회차 최악 초과분(ms)",
     [r"전 회차 최악[은는]?\s*\*{0,2}(\d+(?:\.\d+)?)\s*ms",
      r"관측 최댓값[^\n]{0,20}?\*{0,2}(\d+(?:\.\d+)?)\s*ms"]),
    ("초과분 관측 범위의 하한(ms)",
     [r"관측 범위\s*\*{0,2}(\d+(?:\.\d+)?)\s*ms",
      r"관측 범위[^\n]{0,12}?\*{0,2}(\d+(?:\.\d+)?)\s*~"]),
    ("reserve ÷ 전 회차 최악(배)",
     [r"800\s*/\s*\d+(?:\.\d+)?\s*=\s*\*{0,2}(\d+(?:\.\d+)?)배",
      r"reserve 800\s*ms는 그것의 \*{0,2}(\d+(?:\.\d+)?)배",
      r"전 회차 최악[^\n]{0,20}?reserve 800ms\s*=\s*\*{0,2}(\d+(?:\.\d+)?)배"]),
    # 8라운드 M-5: 8판의 정규식은 `(\d+)배`라 `11.2배`에서 `11`만 읽었고,
    # 같은 양의 두 표기가 나란히 살아 있어도 통과했다. 소수를 포함시킨다.
    ("초과분 산포 배수",
     [r"(\d+(?:\.\d+)?)배가?\s*(?:넘게\s*)?(?:흩어진다|갈린다|갈리므로|흩어졌다)",
      r"산포 배수 = \*{0,2}(\d+(?:\.\d+)?)배",
      r"\*{0,2}(\d+(?:\.\d+)?)배 산포"]),

    # 8라운드 M-8 — design이 말하는 냉 측정값이 어느 검사에도 안 걸렸다.
    # 그 값이 D2 냉 여유 열 전체의 제수인데.
    ("냉 publish 측정(초)",
     [r"냉 측정은 1건, (\d+\.\d+)초",
      r"냉 측정\((\d+\.\d+)s\)",
      r"최댓값 (\d+\.\d+)초"]),

    # 8라운드 B4 — 8판의 대표 산출물(§7.5)의 수치가 어느 검사에도 안 걸렸다.
    # design과 analysis가 서로 다른 값을 말해도 통과했다.
    ("8판 재측정 최악 초과분 — 로거 nil(ms)",
     [r"로거 nil 셀의 최악 초과분 \*{0,2}(\d+\.\d+)\s*ms",
      r"^\| 8판 재측정 — \*\*로거 nil\*\*[^|]*\|[^|]*\|[^|]*?~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"^\| C \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"이번 회차 최악 (\d+\.\d+)\s*ms",
      r"8판 재측정의 최악 (\d+\.\d+)\s*ms"]),
    ("8판 재측정 최악 초과분 — 로거 실제 전체(ms)",
     [r"로거 실제 셀의 최악 초과분 \*{0,2}(\d+\.\d+)\s*ms",
      r"^\| E \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"D=50ms 셀이 [\d.]+~(\d+\.\d+)\s*ms"]),
    ("8판 재측정 최악 초과분 — 로거 실제·채택 상수(ms)",
     [r"^\| 8판 재측정 — \*\*로거 실제\*\*[^|]*\|[^|]*?~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"^\| D \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"셀 D의 (\d+\.\d+)\s*ms",
      r"D=150ms 셀\(셀 D\)이 [\d.]+~(\d+\.\d+)\s*ms"]),
    # 10판이 실행 가능한 프로브로 다시 잰 회차(§7.6). 8판 회차와 셀 이름이 달라야
    # 정규식이 안 겹친다 — 그래서 문서가 프라임을 쓴다. 산문이 이 값을 말하는 자리가
    # 표와 갈리면 8라운드 B4가 §7.5에서 낸 것과 같은 결함이 §7.6에서 재발한다.
    ("10판 재측정 최악 초과분 — 셀 D′(ms)",
     [r"^\| D′ \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"이번 회차 최악은 셀 D′의 (\d+\.\d+)\s*ms"]),
    ("10판 재측정 — 셀 A′ 최댓값(ms)",
     [r"^\| A′ \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"\(59\.2 < (\d+\.\d+)\)"]),
    ("10판 재측정 — 셀 B′ 최댓값(ms)",
     [r"^\| B′ \|[^|]*\|[^|]*\|[^|]*\|[^|~]*~\s*\*{0,2}(\d+\.\d+)\s*ms",
      r"\((\d+\.\d+) < 75\.9\)"]),

    ("줄 간격(무효) 표본 수",
     [r"\(표본\s*(\d+)\)", r"무효 간격 표본[^\n|]{0,60}?n=(\d+)",
      r"줄 간격 표본\s*(\d+)건"]),
    # 7라운드 M-1(d): 7판의 창은 {0,30}이었고 design.md:209의 실제 간격은 31자다.
    # **한 글자 차이로 0매치였고, 0매치가 조용히 통과했다.** 둘 다 고친다.
    ("유효 왕복 표본 수",
     [r"유효[^\n]{0,20}?표본[^\n]{0,80}?n=(\d+)",
      r"^## \d+\. 표본 (\d+)건"]),
    ("냉 표본 수",
     [r"냉 표본\s*\*{0,2}n=(\d+)", r"냉연결 표본은 (\d+)건"]),
    ("`n.mu` 배수 실측(채택 상수, 초)",
     [r"채택 상수[^\n]{0,60}?실측\s*\*{0,2}(\d+\.\d+)초",
      r"loopB\s*=\s*\*{0,2}(\d+\.\d+)초"]),
    ("publish 유발 이벤트 종류(프로덕션)",
     [r"실제 publish 유발[^\n]{0,20}?\*{0,2}(\d+)종",
      r"프로덕션 publish\s*\*{0,2}(\d+)종"]),
    # **백분율을 정규식에 박지 않는다 — 10라운드 M/H4(B26).**
    #
    # 10판은 `\(46%\)`를 **리터럴로** 박았다. 그러면 비율은 절대 평가되지 않고
    # `39개 중 30개(46%)`가 통과한다(실값 77%). 검사가 보는 것이 자기가 미리
    # 적어 둔 문자열이면 그것은 검사가 아니라 그 문자열의 사본이다.
    # 이제 백분율도 캡처해서 **아래 CHECK_RATIOS가 산술로 판정한다.**
]

# **비율은 유일성이 아니라 산술로 본다 — 10라운드 H4(B26).**
#
# 10판은 이것을 명명값으로 두고 정규식에 `\(46%\)`를 **리터럴로** 박았다. 결과가
# 둘이다. 첫째, 비율이 절대 평가되지 않아 `39개 중 30개(46%)`가 통과했다(실값 77%).
# 둘째, **박아 둔 46%가 아닌 비율은 아예 안 보였다** — 이 문서군에는 `47개 중
# 24개(51%)`라는 두 번째 비율이 있고 10판은 그것을 한 번도 검사하지 않았다.
#
# 유일성은 여기서 틀린 도구다. 서로 다른 창의 비율이 서로 다른 것은 정상이다.
# 물어야 할 것은 "같은가"가 아니라 **"닫히는가"** 이고, 그것은 산술이다.
RATIO = re.compile(r"(\d+)\s*[개건] 중 (\d+)\s*[개건]?\((\d+)%\)")


def check_ratios(files: list[pathlib.Path]) -> None:
    seen = 0
    for path in files:
        for no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            for m in RATIO.finditer(line):
                den, num, pct = (int(g) for g in m.groups())
                seen += 1
                if not den:
                    continue
                real = num / den * 100
                # 반올림 표기를 허용하되 그 이상은 허용하지 않는다.
                if abs(real - pct) > 1.0:
                    fail(f"{rel(path)}:{no}",
                         f"비율이 안 닫힌다: {num}/{den} = {real:.1f}% 인데 "
                         f"{pct}%라고 적혀 있다 — {m.group(0)}")
    if seen == 0:
        fail("[비율]", "`N개 중 M개(P%)` 꼴을 문서 어디에서도 못 찾았다 — "
                       "0매치는 통과가 아니다. 표기가 바뀌었으면 정규식을 고친다")
    else:
        NOTES.append(f"  비율 {seen}곳 산술 통과 (유도 — 손으로 박지 않는다)")


def check_named(files: list[pathlib.Path]) -> None:
    for name, patterns in NAMED_QUANTITIES:
        seen: dict[str, list[str]] = {}
        for path in files:
            for no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
                for pat in patterns:
                    for m in re.finditer(pat, line):
                        key = "/".join(m.groups())
                        seen.setdefault(key, []).append(f"{rel(path)}:{no}")
        if len(seen) > 1:
            detail = "; ".join(f"{k} ({', '.join(v)})" for k, v in sorted(seen.items()))
            fail(f"[명명값] {name}", f"두 개 이상의 값이 있다 — {detail}")
        elif seen:
            NOTES.append(f"  {name} = {next(iter(seen))} "
                         f"({len(next(iter(seen.values())))}곳)")
        else:
            # 7라운드 M-1(d). 0매치는 "일치한다"가 아니라 "안 봤다"이다.
            # 검사 목록에 이름만 남고 아무 데도 안 걸리는 항목은 검사가 아니라 장식이다.
            fail(f"[명명값] {name}",
                 "이 양을 문서 어디에서도 못 찾았다 — 0매치는 통과가 아니다. "
                 "양이 문서에서 사라졌으면 이 항목을 지우고, 표기가 바뀌었으면 "
                 "정규식을 고친다")


# ---------------------------------------------------------------------------
# 검사 6 — 판 번호
# ---------------------------------------------------------------------------

# **한 자리가 아니라 여러 자리다** — 10라운드 축 2. 9판의 `\((\d)판\)`은 `(10판)`을
# 아예 못 읽었다. 다음 판이 10판이므로 이 검사는 **다음 커밋에서 전 문서 0매치로
# 조용히 죽었을 것**이다. 그리고 `check_named`와 달리 0매치 가드가 없었다 —
# 7라운드 M-1(d)가 "0매치는 통과가 아니다"라고 이름 붙인 결함이 이 검사만 비껴갔다.
DRAFT_MARK = re.compile(r"\((\d+)판\)|\*\*(\d+)판이다\*\*")


def check_draft(files: list[pathlib.Path]) -> None:
    seen = 0
    for path in files:
        if path.name not in ("proposal.md", "design.md", "tasks.md") and \
           path.parent.name != "analysis":
            continue
        for no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            for m in DRAFT_MARK.finditer(line):
                got = int(m.group(1) or m.group(2))
                seen += 1
                # 이력 서술("3판은 …")은 괄호나 '판이다' 형태가 아니다.
                if got != DRAFT:
                    fail(f"{rel(path)}:{no}",
                         f"판 번호가 {got}판이다 — 이 판은 {DRAFT}판이다")
    if seen == 0:
        fail("[판 번호]",
             f"어느 문서에서도 판 표기를 못 찾았다 — 0매치는 통과가 아니다. "
             f"표기가 바뀌었으면 DRAFT_MARK를 고친다")
    else:
        NOTES.append(f"  판 표기 {seen}곳 — 전부 {DRAFT}판")


# ---------------------------------------------------------------------------
# 검사 7 — 유도 카운트 (기대값을 박지 않는다)
# ---------------------------------------------------------------------------

def check_derived_counts() -> None:
    dirs = sorted(d for d in FLM_ROOT.iterdir() if (d / "ast.json").exists())
    n_artifacts = len(dirs)

    tasks = (CHANGE / "tasks.md").read_text(encoding="utf-8").splitlines()
    sweep_rows = [
        (no, l) for no, l in enumerate(tasks, 1)
        if re.match(r"^\|\s*\d+\s*\|\s*`", l)
    ]
    if len(sweep_rows) != n_artifacts:
        fail("[유도] 훑기 표 행 수",
             f"산출물 {n_artifacts}개인데 tasks.md 훑기 표는 {len(sweep_rows)}행이다")

    for claim_no, line in enumerate(tasks, 1):
        for m in re.finditer(r"함수 \*\*(\d+)개\*\*", line):
            if int(m.group(1)) != n_artifacts:
                fail(f"{rel(CHANGE / 'tasks.md')}:{claim_no}",
                     f"'함수 {m.group(1)}개'가 산출물 수 {n_artifacts}와 다르다")

    # D3 컴파일 실측 표의 행 수 — 산문이 이 수를 말하면 표와 같아야 한다.
    #
    # 9라운드 H-2: 8라운드 H5가 한 행을 타입별 둘로 가르면서, 여섯 곳에 손으로 박힌
    # "여덟 구성"이 낡았다. 표는 아홉 행이었다. **기대값을 박은 검사는 검사가 아니라
    # 또 하나의 주장이다**가 이 파일의 원칙인데, 그 원칙이 검사 코드에만 적용되고
    # 산문에는 안 적용되고 있었다. 세 번째 재발이라 이제 센다.
    design_lines = (CHANGE / "design.md").read_text(encoding="utf-8").splitlines()
    build_rows = [l for l in design_lines
                  if re.match(r"^\|.*\|\s*\*{0,2}(?:BUILD OK|FAIL)\*{0,2}\s*\|", l)]
    n_configs = len(build_rows)
    if n_configs == 0:
        fail("[유도] D3 컴파일 실측 표",
             "BUILD OK/FAIL 행을 하나도 못 찾았다 — 0매치는 통과가 아니다")
    else:
        NOTES.append(f"  D3 컴파일 실측 {n_configs}구성 (유도 — 손으로 박지 않는다)")
        for path in (CHANGE / "design.md", CHANGE / "proposal.md",
                     CHANGE / "tasks.md"):
            for no, line in enumerate(
                    path.read_text(encoding="utf-8").splitlines(), 1):
                for m in CONFIG_COUNT.finditer(line):
                    got = KO_NUM.get(m.group(1), m.group(1))
                    if str(got) != str(n_configs):
                        fail(f"{rel(path)}:{no}",
                             f"'{m.group(0)}'이 D3 실측 표의 {n_configs}행과 "
                             f"다르다 — 수를 산문에 박지 말고 표를 가리킨다")

    # 비표준 절: 전부 훑기 표에 인용이 있어야 한다.
    table_text = "\n".join(l for _, l in sweep_rows)
    uncited: list[str] = []
    total = 0
    for d in dirs:
        flm = d / "function-logic-map.md"
        if not flm.exists():
            continue
        for line in flm.read_text(encoding="utf-8").splitlines():
            if not re.match(r"^#{2,4} ", line):
                continue
            title = line.lstrip("# ").strip()
            if title in STANDARD_SECTIONS:
                continue
            total += 1
            plain = re.sub(r"[`*]", "", title)
            windows = [plain[i:i + 8] for i in range(max(1, len(plain) - 7))]
            if not any(w in table_text for w in windows):
                uncited.append(f"{d.name} · {title}")
    NOTES.append(f"  비표준 절 {total}개 (유도 — 손으로 박지 않는다)")
    if uncited:
        fail("[유도] 표가 인용하지 않은 비표준 절",
             f"{len(uncited)}개 — " + "; ".join(uncited))


# ---------------------------------------------------------------------------

def main() -> int:
    global CANDIDATE_TOKENS
    files = target_files()
    corpus = measurement_corpus()
    CANDIDATE_TOKENS = candidate_tokens()
    if not CANDIDATE_TOKENS:
        fail("[유도] 후보 표 토큰",
             "design D2의 후보 표에서 수치를 하나도 못 읽었다 — "
             "0매치는 통과가 아니다(표 모양이 바뀌었으면 SWEEP_ROW를 고친다)")

    # 인용한 냉 측정이 실제로 analysis/에 있는가. 없으면 D2의 냉 여유 열 전체가
    # 근거 없는 나눗셈이 된다.
    if not cited_ok(f"{COLD_PUBLISH_S}", corpus):
        fail("[선언] COLD_PUBLISH_S",
             f"{COLD_PUBLISH_S}s가 analysis/ 어디에도 없다 — "
             f"D2의 냉 여유 열이 나누는 값이므로 측정에 실려 있어야 한다")

    sweep_rows_total = 0
    for path in files:
        # **접은 뒤에 스캔한다.** 제로폭 공백·합자 단위·소수점 쉼표는 렌더에서
        # 보이지 않으므로, 접기 전에 스캔하면 검사와 독자가 다른 문서를 읽는다.
        lines = [fold_notation(l)
                 for l in path.read_text(encoding="utf-8").splitlines()]
        check_arithmetic(path, lines)
        check_rejected_values(path, lines)
        check_orphan_measurements(path, lines, corpus)
        check_dwell_enumeration(path, lines)
        if path.name == "design.md":
            sweep_rows_total = check_candidate_sweep(path, lines)

    if sweep_rows_total == 0:
        fail("[검사 3] 후보 훑기 표", "design.md에서 후보 행을 하나도 못 찾았다 — "
                                     "표 모양이 바뀌었으면 정규식을 고친다")
    else:
        NOTES.append(f"  후보 훑기 표 {sweep_rows_total}행 산술 통과")

    check_named(files)
    check_ratios(files)
    check_draft(files)
    check_derived_counts()

    print(f"a092 값 단위 교차문서 검사 — 대상 {len(files)}개 문서 "
          f"(review.md 제외: 기각된 값을 일부러 인용한다)")
    for note in NOTES:
        print(note)

    if FAILURES:
        print(f"\n실패 {len(FAILURES)}건\n")
        for i, f in enumerate(FAILURES, 1):
            print(f"  {i:2d}. {f}")
        return 1

    print("\n통과 — 검사한 값 전부가 문서마다 하나의 표기를 갖는다.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
