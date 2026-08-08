#!/usr/bin/env python3
"""`check_values.py`가 자기가 잡는다고 적은 것을 실제로 잡는지 돌려 본다.

왜 이것이 있는가
================

7라운드 차단 5건 중 4건이 **7판이 새로 만든 방어** 안에 있었다. 형태가 하나였다 —
방어를 쓰고 **돌려 보지 않았다.** `check_values.py`는 자기가 잡는다고 적은 `28.9`를
실제로는 못 잡았고(7라운드 B2), 우회 네 가지가 실증됐다(M-1).

그러므로 8판의 규칙은 이렇다: **검사를 고쳤으면 그 검사를 우회하는 문서를 만들어
실패하는 것을 본다.** 하나라도 통과하면 이 자체 시험이 실패한다.

**8판은 그 규칙을 자기가 고친 부분에는 안 지켰다.** 케이스 8건이 전부 7라운드가
이미 이름 붙인 우회의 재생이었고, 8판이 **새로 넣은** 구성 — `≥` 룩비하인드,
`COUNTERFACTUAL` 낱말 면제, §7.5 수치 — 은 한 건도 안 덮였다. 그 셋이 그대로
8라운드 차단 B2·B3·B4가 됐다. 그래서 9판의 규칙은 한 문장 더 길다:

    검사에 줄을 추가하면 **그 줄을 우회하는 케이스를 같은 커밋에** 넣는다.
    회귀 방지가 아니라 **신규 방어 검증**이 목적이다.

이 시험은 저장소를 **읽기만 한다.** change 디렉터리를 임시 사본으로 복제하고 그
사본만 오염시킨다.

    python3 openspec/changes/a092-.../tools/check_values_selftest.py

종료 코드 0이면 모든 방어가 살아 있다.
"""

from __future__ import annotations

import pathlib
import re
import shutil
import subprocess
import sys
import tempfile

CHANGE = pathlib.Path(__file__).resolve().parent.parent
CHANGE_ID = CHANGE.name


def run_in(copy_root: pathlib.Path) -> tuple[int, str]:
    script = copy_root / "tools" / "check_values.py"
    proc = subprocess.run([sys.executable, str(script)],
                          capture_output=True, text=True)
    return proc.returncode, proc.stdout + proc.stderr


def edit(root: pathlib.Path, rel_path: str, old: str, new: str) -> None:
    p = root / rel_path
    text = p.read_text(encoding="utf-8")
    if old not in text:
        raise SystemExit(
            f"자체 시험을 못 돌린다: {rel_path} 안에 앵커가 없다 — {old!r}\n"
            f"문서가 바뀌었으면 이 시험의 앵커를 고친다. 앵커가 사라진 채로 "
            f"'통과'하면 그것이 바로 7라운드가 잡은 결함이다.")
    p.write_text(text.replace(old, new, 1), encoding="utf-8")


def append(root: pathlib.Path, rel_path: str, text: str) -> None:
    p = root / rel_path
    p.write_text(p.read_text(encoding="utf-8") + text + "\n", encoding="utf-8")


def edit_all(root: pathlib.Path, old: str, new: str) -> None:
    """change 안의 모든 `.md`에서 바꾼다. **전 문서에서 표기가 사라지는** 우회는
    한 파일만 고쳐서는 재현되지 않는다 — 0매치 가드가 그것을 위해 있다."""
    hit = 0
    for p in sorted(root.rglob("*.md")):
        t = p.read_text(encoding="utf-8")
        if old in t:
            p.write_text(t.replace(old, new), encoding="utf-8")
            hit += 1
    if not hit:
        raise SystemExit(
            f"자체 시험을 못 돌린다: 어느 문서에도 앵커가 없다 — {old!r}")


# **판 표기를 손으로 박지 않는다 — 11판.** 10판은 앵커에 `(10판)`을 박았고,
# 그래서 판을 올리는 순간 자체 시험이 "앵커가 없다"로 **죽었다.** 차단은 되지만
# 진단이 "우회를 잡았다"가 아니라 "못 돌린다"이다. 판 번호는 `check_values.py`가
# 이미 하나 갖고 있으므로 그것을 읽는다 — 값의 사본을 만들지 않는다.
_CV = (pathlib.Path(__file__).resolve().parent / "check_values.py").read_text(
    encoding="utf-8")
DRAFT_NO = int(re.search(r"^DRAFT = (\d+)", _CV, re.M).group(1))
DRAFT_TAG = f"({DRAFT_NO}판)"
DRAFT_ALT = f"({DRAFT_NO}차)"


# 각 항목: (이름, 오염 함수, 출력에 있어야 할 조각)
#
# 조각까지 보는 이유: 아무 이유로나 실패하는 것은 "잡았다"가 아니다.
CASES: list[tuple[str, object, str]] = [
    (
        "7라운드 B2 — 기각값 재인용 (표지 없음)",
        lambda r: append(r, "proposal.md",
                         "\n측정된 초과분은 28.9 ms다.\n"),
        "기각된 값 28.9",
    ),
    (
        "7라운드 M-1(c) — 단위 뒤에 한국어 조사",
        lambda r: append(r, "proposal.md",
                         "\n그 회차의 초과분은 28.9ms였다.\n"),
        "기각된 값 28.9",
    ),
    (
        "7라운드 B2 — 기각 문장이 새 값을 코퍼스에 세탁",
        lambda r: (
            append(r, "analysis/delivery-latency.md",
                   "\n6판이 적은 초과분 77.77 ms는 근거가 없었다. "
                   "<!-- rejected-value -->\n"),
            append(r, "proposal.md",
                   "\n채택 상수의 초과분은 77.77 ms다.\n"),
        ),
        "고아 측정치 77.77",
    ),
    (
        "7라운드 M-1(a) — 라벨을 안 쓰면 채택량 등식이 빠져나간다",
        lambda r: append(r, "design.md",
                         "\n한 알림이 쓰는 시간 = 3 × 1.5s + 2 × 250ms = 5.0s\n"),
        "채택하지 않은 값",
    ),
    (
        "7라운드 M-1(b) — 냉 여유 열은 아무도 안 봤다",
        lambda r: edit(r, "design.md",
                       "| **4.20s** | **800ms** | **1.72×** |",
                       "| **4.20s** | **800ms** | **9.99×** |"),
        "냉 여유가 Timeout ÷ 냉 측정과 다르다",
    ),
    (
        "7라운드 M-1(d) — 0매치가 조용히 통과한다",
        lambda r: (
            edit(r, "design.md", "유효 왕복 표본 (§3)", "유효 왕복 샘플 (§3)"),
            edit(r, "analysis/delivery-latency.md",
                 "## 3. 표본 6건", "## 3. 표본 여섯 건"),
        ),
        "0매치는 통과가 아니다",
    ),
    (
        "6라운드 M6 — 판 번호가 문서마다 다르다",
        lambda r: edit(r, "design.md", DRAFT_TAG, "(8판)"),
        "판 번호가 8판이다",
    ),
    (
        "6라운드 B1 — 산술이 안 닫힌다",
        lambda r: append(r, "design.md",
                         "\ntransport 예산 = 3 × 1.3s + 2 × 150ms = 4.9s\n"),
        "산술이 안 닫힌다",
    ),

    # --- 9판이 더한 케이스: 8판이 *새로 넣은* 구성을 우회한다 -----------------
    #
    # 8라운드 차단 4건 중 3건이 여기 있었다. 8판의 자체 시험 8건은 전부 7라운드가
    # 이미 이름 붙인 우회의 재생이었고, 8판이 새로 넣은 룩비하인드·낱말 면제·
    # §7.5 수치는 **한 건도 안 덮였다.** 검사에 줄을 추가하면 그 줄을 우회하는
    # 케이스를 같은 커밋에 넣는다 — 회귀 방지가 아니라 신규 방어 검증이다.

    (
        "8라운드 B2 — 부등호 앞에 붙은 기각값 (룩비하인드가 첫 자리를 먹는다)",
        lambda r: append(r, "proposal.md",
                         "\n그 배수의 최악 체류는 ≥ 9.231초다.\n"),
        "기각된 값 9.231",
    ),
    (
        "8라운드 B2 — 같은 것, 진단이 없는 수를 말하면 안 된다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 ≥ 28.9 ms였다.\n"),
        "기각된 값 28.9",
    ),
    (
        "8라운드 M-4 — 기각값이 지연 낱말 게이트 밖에 있다",
        lambda r: append(r, "proposal.md",
                         "\n운영자가 관측한 값은 28.9 ms다.\n"),
        "기각된 값 28.9",
    ),
    (
        "8라운드 M-4 — 기각값이 OUT_OF_SCOPE 낱말과 같은 줄에 있다",
        lambda r: append(r, "proposal.md",
                         "\n그 체류는 timeout 설정 탓에 9.231초였다.\n"),
        "기각된 값 9.231",
    ),
    (
        "8라운드 B3 — 반증 낱말 하나로 채택량 등식이 면제된다",
        lambda r: append(r, "design.md",
                         "\n기각 이력을 정리하면, 채택된 transport 예산 = "
                         "3 × 1.5s + 2 × 250ms = 5.0s 이다.\n"),
        "채택하지 않은 값",
    ),
    (
        "8라운드 B4 — §7.5 수치가 design과 analysis에서 갈린다",
        lambda r: edit(r, "design.md",
                       "79.9 ~ **319.3 ms**", "79.9 ~ **919.3 ms**"),
        "8판 재측정 최악",
    ),
    (
        "8라운드 M-8 — design이 말하는 냉 측정값이 검사 밖이다",
        lambda r: edit(r, "design.md",
                       "냉 측정은 1건, 0.754초다", "냉 측정은 1건, 0.900초다"),
        "냉 publish 측정",
    ),
    (
        "8라운드 M-7 — 정수 측정치는 고아 검사가 못 본다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 4777 ms였다.\n"),
        "고아 측정치 4777",
    ),
    (
        "8라운드 M-7 — 절 번호가 측정 코퍼스를 오염시킨다",
        lambda r: append(r, "proposal.md",
                         "\n한 알림의 체류 실측은 7.5s였다.\n"),
        "고아 측정치 7.5",
    ),
    (
        "8라운드 B1 — 요구 문단만 고치고 Scenario 열거는 그대로 둔다",
        lambda r: edit(r, "specs/engine-safety/spec.md",
                       "outbox 기록·시도별 실패 기록·게이트 래치·"
                       "운영 모드 승격 트랜잭션·그 호출이 쓰는 구조화 로그 줄을 합쳐",
                       "outbox·게이트·승격 작업을 합쳐"),
        "체류 구성 열거에 빠진 항이 있다",
    ),
    (
        "8라운드 B1 — 같은 것, exit-policy 쪽",
        lambda r: edit(r, "specs/exit-policy/spec.md",
                       "·그 호출이 쓰는 구조화 로그 줄까지 마친 시점이",
                       "까지 마친 시점이"),
        "체류 구성 열거에 빠진 항이 있다",
    ),

    # --- 10판이 더한 케이스 --------------------------------------------------
    #
    # 아래는 **손으로 고른 것이 아니다.** `tools/coverage_gate.py` 축 1이
    # `fail()` 이 도달할 수 있는 경로를 하나씩 죽여 보고 "죽여도 자체 시험이
    # 그대로인 자리"를 셌고, 그 목록이 그대로 여기 케이스가 됐다. 9판까지 세 판
    # 연속으로 틀린 것이 바로 **사람이 '덮었다'고 판단한 것**이므로 판단하지 않고 센다.
    #
    # 축 2(섭동 행렬)가 잡은 구멍은 케이스가 아니라 **검사 자신**을 고쳐서 막았다 —
    # 게이트 셋 제거·단위 정규화·문단 단위 열거·마침표 경계. 축 2는 그 수정이
    # 유지되는지를 매 실행마다 다시 훑는다.
    #
    # **여기에 수를 적지 않는다 — 10라운드 M7.** 10판은 「14건」·「134칸」이라
    # 적었고 그 시점에 이미 15건·146칸이었다. 이 파일이 마지막 줄에 자기 케이스
    # 수를 세어 출력하고 `coverage_gate.py`가 자기 칸 수를 세어 출력한다.
    # *"기대값을 박은 검사는 검사가 아니라 또 하나의 주장이다"* 는 이 change의
    # 원칙이고, 그 원칙은 **주석에도** 적용된다.

    (
        "축1 — 반증 표지가 값을 지목하지 않는다",
        lambda r: append(r, "design.md",
                         "\n채택된 transport 예산 = 3 × 1.5s + 2 × 250ms = 5.0s "
                         "<!-- counterfactual -->\n"),
        "반증 표지가 값을 지목하지 않는다",
    ),
    # --- 11판이 더한 케이스 ------------------------------------------------
    #
    # 10판의 축 1은 `fail()` 을 **줄 번호**로 죽였다. 래퍼를 거치는 호출은 `fail()`
    # 줄이 같으므로 호출자들이 라벨 하나로 붕괴했고, 아래 셋은 그 붕괴 뒤에 숨어
    # 있었다(10라운드 차단 4). 11판이 죽이는 단위를 **스택 경로**로 바꾸자
    # 축 1이 곧바로 이 셋을 이름으로 불렀다 — 손으로 고른 케이스가 아니다.
    (
        "11판 축1 — 뺄셈 등식 줄의 반증 표지가 값을 지목하지 않는다",
        lambda r: append(r, "design.md",
                         "\nreserve = 5.0s − 4.2s = 800ms "
                         "<!-- counterfactual -->\n"),
        "반증 표지가 값을 지목하지 않는다",
    ),
    (
        "11판 축1 — 한 항 등식 줄의 반증 표지가 값을 지목하지 않는다",
        lambda r: append(r, "design.md",
                         "\n정상 = 1 × 1.3s = 1.3s "
                         "<!-- counterfactual -->\n"),
        "반증 표지가 값을 지목하지 않는다",
    ),
    (
        # 8라운드 B1의 **거울상**. 그 케이스는 Scenario 쪽(:827)을 밟고,
        # 요구 문단 쪽(:809)은 아무도 안 밟고 있었다.
        "11판 축1 — 요구 문단 열거만 항을 잃고 Scenario는 멀쩡하다",
        lambda r: edit(r, "specs/engine-safety/spec.md",
                       "(outbox 기록·시도별 실패 기록·게이트 래치·"
                       "운영 모드 승격 트랜잭션·**그 호출이 쓰는 구조화 로그 줄**)",
                       "(outbox 기록·게이트 래치·"
                       "운영 모드 승격 트랜잭션·**그 호출이 쓰는 구조화 로그 줄**)"),
        "체류 구성 열거에 빠진 항이 있다",
    ),
    # 10라운드 차단 3 — 우회 36건 중 안전에 직접 닿는 것들. 검사를 고치고
    # 그 수정이 유지되는지를 케이스로 고정한다.
    (
        # B14. 다섯 항을 **둘로** 줄이면 DWELL_TRIGGER_MIN(3) 미만이라 10판의
        # 트리거 루프에는 문단이 아예 안 보였고, 같은 파일 Scenario가 seen>0을
        # 채워 0매치 가드도 안 걸렸다. 9라운드 차단 2·8라운드 B1의 재현이다.
        "10라운드 B14 — 요구 열거를 트리거 미만(두 항)으로 줄인다",
        lambda r: edit(r, "specs/engine-safety/spec.md",
                       "(outbox 기록·시도별 실패 기록·게이트 래치·"
                       "운영 모드 승격 트랜잭션·**그 호출이 쓰는 구조화 로그 줄**)",
                       "(outbox 기록과 게이트 래치 **둘뿐이다**)"),
        "체류 구성 열거에 빠진 항이 있다",
    ),
    (
        # B16. 같은 축소를 proposal에 하고, 파일 끝에 완전한 열거를 담은 미끼
        # 문단을 붙여 **파일 단위** 0매치 가드를 끈다. 자리를 선언하면 미끼가
        # 소용없다 — 검사가 보는 것은 파일이 아니라 앵커가 있는 문단이다.
        "10라운드 B16 — 열거를 줄이고 미끼 문단으로 0매치 가드를 끈다",
        lambda r: (
            edit(r, "proposal.md",
                 "`MarkAlertAttemptFailed` 시도 수만큼(시도별 실패 기록), "
                 "`Gate.Block`(게이트 래치),\n"
                 "**소진 시 `n.escalate`의 `EscalateOperatingMode` "
                 "승격 트랜잭션**(`notifier.go:187`),\n"
                 "그리고 **그 호출이 쓰는 구조화 로그 줄**"
                 "(`notifier.go:131`·`:279` 항상,\n"
                 "`:228`은 승격이 실제로 일어났을 때). ",
                 "`Gate.Block`(게이트 래치)뿐이다. "),
            append(r, "proposal.md",
                   "\n> outbox 기록·시도별 실패 기록·게이트 래치·"
                   "운영 모드 승격 트랜잭션·그 호출이 쓰는 구조화 로그 줄\n"),
        ),
        "체류 구성 열거에 빠진 항이 있다",
    ),
    (
        # B27. change 루트에 문서를 하나 더하면 10판의 `target_files()` 밖이라
        # 채택 상수를 전부 다르게 적고 기각값을 인용해도 통과했다.
        "10라운드 B27 — change 루트에 새 문서를 더해 검사 밖으로 나간다",
        lambda r: (r / "addendum.md").write_text(
            "# a092 부록\n\n이 회차의 초과분 최악은 28.9 ms였다.\n",
            encoding="utf-8"),
        "기각된 값 28.9",
    ),
    # 표기 우회 계열 — 10라운드 B08~B12·B22. 전부 렌더된 문서에서는 `28.9 ms`·
    # `9.231 s`와 구별되지 않는데, 10판의 정규식에게는 다른 문자열이었다.
    (
        "10라운드 B08 — 기각값의 단위를 `sec`/`msec`으로 적는다",
        lambda r: append(r, "proposal.md",
                         "\n그 배수의 최악 체류는 9.231 sec다.\n"),
        "기각된 값",
    ),
    (
        "10라운드 B09 — 기각값을 마이크로초로 적는다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 28900 µs였다.\n"),
        "기각된 값",
    ),
    (
        "10라운드 B10 — 기각값의 단위를 합자 문자(㎳)로 적는다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 28.9 ㎳였다.\n"),
        "기각된 값",
    ),
    (
        "10라운드 B11 — 기각값과 단위 사이에 제로폭 공백을 넣는다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 28.9​ms였다.\n"),
        "기각된 값",
    ),
    (
        "10라운드 B12 — 기각값을 소수점 쉼표로 적는다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 28,9 ms였다.\n"),
        "기각된 값",
    ),
    (
        "10라운드 B22 — 기각값의 단위를 `밀리세컨드`로 적는다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악은 28.9 밀리세컨드였다.\n"),
        "기각된 값",
    ),
    # --- 11판이 더한 줄을 우회하는 케이스 ------------------------------------
    #
    # 축 1이 이름을 부른 자리다. 11판이 검사에 줄을 더했고 그중 셋이 **자기 케이스
    # 없이** 있었다 — 8·9판이 세 판 연속 저지른 것과 똑같은 형태다. 사람이 아니라
    # 기계가 그것을 말해 줬다는 것만 다르다.
    (
        # 앵커가 없는 문서의 열거는 여전히 트리거 루프가 본다. 새 spec delta가
        # 정확히 그 경우다(10라운드 B28도 이 형태였다).
        "11판 축1 — 앵커 없는 새 spec delta의 열거에 항이 빠진다",
        lambda r: ((r / "specs" / "observability").mkdir(exist_ok=True),
                   (r / "specs" / "observability" / "spec.md").write_text(
                       "# observability delta\n\n"
                       "## MODIFIED Requirements\n\n"
                       "### Requirement: 알림 체류 관측\n\n"
                       "그 시간은 같은 호출 안의 outbox 기록·시도별 실패 기록·"
                       "게이트 래치를 포함한다(SHALL).\n",
                       encoding="utf-8")),
        "체류 구성 열거에 빠진 항이 있다",
    ),
    (
        "11판 축1 — 요구 열거의 앵커 문장 자체를 지운다",
        lambda r: edit(r, "specs/engine-safety/spec.md",
                       '"자기 차례"는 전송에 쓴 시간과 같은 호출 안의 비전송 작업',
                       "그 시간은 아래 작업"),
        "요구 열거가 사는 문단을 못 찾았다",
    ),
    (
        "11판 축1 — 비율 표기가 전 문서에서 사라진다",
        # 두 표기를 **다** 지워야 0매치다. 한쪽만 지우면 다른 쪽이 남아
        # 가드가 안 걸린다 — 첫 시도가 정확히 그렇게 통과했다.
        lambda r: (edit_all(r, "개 중 ", "개중 "),
                   edit_all(r, "건 중 ", "건중 ")),
        "꼴을 문서 어디에서도 못 찾았다",
    ),
    (
        # 10판은 `\(46%\)`를 리터럴로 박아 비율을 절대 평가하지 않았다.
        "10라운드 B26 — 비율의 분자를 바꾼다 (30/39을 46%라 적는다)",
        lambda r: edit(r, "design.md",
                       "**39개 중 18개(46%)**", "**39개 중 30개(46%)**"),
        "비율이 안 닫힌다",
    ),
    (
        "10라운드 B24 — not-a-measurement 표지의 사유를 비운다",
        lambda r: append(r, "proposal.md",
                         "\n이 회차의 초과분 최악 실측은 77.77 ms였다. "
                         "<!-- not-a-measurement: -->\n"),
        "사유가 비었거나 너무 짧다",
    ),
    (
        "축1 — 뺄셈 등식의 산술이 안 닫힌다",
        lambda r: append(r, "design.md",
                         "\nreserve = 5.0s − 4.2s = 900ms\n"),
        "산술이 안 닫힌다",
    ),
    (
        "축1 — 뺄셈 등식이 채택하지 않은 값을 쓴다",
        lambda r: append(r, "design.md",
                         "\nreserve = 5.0s − 4.0s = 1.0s\n"),
        "채택하지 않은 값",
    ),
    (
        "축1 — 한 항 등식의 산술이 안 닫힌다",
        lambda r: append(r, "design.md",
                         "\n한 알림이 쓰는 시간 = 3 × 1.3s = 4.0s\n"),
        "산술이 안 닫힌다",
    ),
    (
        "9라운드 B-6 — 한 항 등식에 채택량 검사가 없다",
        lambda r: append(r, "design.md",
                         "\nalertTransportBudget = 3 × 1.4s = 4.2s\n"),
        "채택하지 않은 값",
    ),
    (
        "축1 — 후보 행의 transport가 자기 입력과 다르다",
        lambda r: edit(r, "design.md",
                       "| 4 | 2 | 2.0s | 250ms | 4.25s |",
                       "| 4 | 2 | 2.0s | 250ms | 4.50s |"),
        "transport가 자기 입력과 다르다",
    ),
    (
        "축1 — 후보 행의 reserve가 budget − transport와 다르다",
        lambda r: edit(r, "design.md",
                       "| 4.25s | 750ms |", "| 4.25s | 700ms |"),
        "reserve가 budget − transport와 다르다",
    ),
    (
        "9라운드 B-8 — 후보 행이 모양만 바꿔 검사 밖으로 나간다",
        lambda r: edit(r, "design.md", "| 1.33× |", "| 1.33배 |"),
        "후보 표에서 빠져나간 행",
    ),
    (
        "9라운드 축2 — 열거의 낱말은 다 있으나 의미가 뒤집힌다",
        lambda r: append(r, "specs/engine-safety/spec.md",
                         "\n이 상한은 outbox 기록·시도별 실패 기록·게이트 래치·"
                         "운영 모드 승격 트랜잭션·그 호출이 쓰는 구조화 로그 줄은 "
                         "**제외하고** 잰다.\n"),
        "의미가 뒤집혀",
    ),
    (
        "9라운드 축2 — 열거를 담아야 할 Scenario가 사라진다",
        lambda r: edit(r, "specs/engine-safety/spec.md",
                       "#### Scenario: 한 알림의 체류가 exit 관측 주기를 넘지 않는다",
                       "#### Scenario: 한 알림의 체류"),
        "열거를 담아야 하는 Scenario를 못 찾았다",
    ),
    (
        "9라운드 B-9 — 판 표기가 전 문서에서 사라진다",
        lambda r: (
            edit(r, "design.md", DRAFT_TAG, DRAFT_ALT),
            edit(r, "tasks.md", DRAFT_TAG, DRAFT_ALT),
            edit(r, "analysis/delivery-latency.md", DRAFT_TAG, DRAFT_ALT),
            edit(r, "analysis/notify-reach.md", DRAFT_TAG, DRAFT_ALT),
        ),
        "판 표기를 못 찾았다",
    ),
    (
        "축1 — 훑기 표 행 수가 산출물 수와 어긋난다",
        lambda r: append(r, "tasks.md",
                         "\n| 99 | `없는함수` | 표에만 있는 행 |\n"),
        "훑기 표는",
    ),
    (
        "축1 — '함수 N개'가 산출물 수와 다르다",
        lambda r: edit(r, "tasks.md", "함수 **23개**", "함수 **24개**"),
        "가 산출물 수",
    ),
    (
        "9라운드 H-2 — 산문이 D3 표와 다른 구성 수를 말한다",
        lambda r: append(r, "design.md",
                         "\n앞 절에서 아홉 구성을 실측했다.\n"),
        "D3 실측 표의",
    ),
    (
        "축1 — 표가 인용하지 않는 비표준 절이 생긴다",
        lambda r: edit(r,
                       "analysis/function-logic/internal-obs--notifier.deliver/"
                       "function-logic-map.md",
                       "### 최악 예산의 산술",
                       "### 아무도 인용하지 않는 새 절 제목"),
        "표가 인용하지 않은 비표준 절",
    ),
]


def main() -> int:
    failures: list[str] = []

    with tempfile.TemporaryDirectory() as tmp:
        # 실제 배치를 흉내 낸다 — check_values.py 의 rel() 이 세 단계 위를 쓴다.
        base = pathlib.Path(tmp) / "openspec" / "changes"
        base.mkdir(parents=True)

        clean = base / CHANGE_ID
        shutil.copytree(CHANGE, clean,
                        ignore=shutil.ignore_patterns("__pycache__"))

        rc, out = run_in(clean)
        if rc != 0:
            failures.append(
                "기준선: 오염하지 않은 사본이 이미 실패한다 — 아래 결과는 "
                "무의미하다\n" + out)
            print("\n".join(failures))
            return 1
        print("기준선 통과 — 오염하지 않은 사본은 exit 0")

        for i, (name, corrupt, want) in enumerate(CASES):
            root = base / f"{CHANGE_ID}-case{i}"
            shutil.copytree(clean, root)
            corrupt(root)
            rc, out = run_in(root)
            if rc == 0:
                failures.append(f"[{name}] 우회가 **통과했다** — 검사가 이것을 "
                                f"안 잡는다\n{out}")
            elif want not in out:
                failures.append(f"[{name}] 실패는 했으나 이유가 다르다 — "
                                f"{want!r}가 출력에 없다\n{out}")
            else:
                print(f"잡았다 — {name}")

    if failures:
        print(f"\n자체 시험 실패 {len(failures)}건\n")
        for i, f in enumerate(failures, 1):
            print(f"  {i}. {f}\n")
        return 1

    print(f"\n자체 시험 통과 — 우회 {len(CASES)}가지 전부 잡힌다")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
