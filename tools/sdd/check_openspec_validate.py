#!/usr/bin/env python3
"""완료 게이트 11단계 — openspec validate 판정 (a190).

판정 = (가) 닫는 change 를 strict 로 + (나) 저장소 전체를 strict 없이(오류만). 둘 다 통과해야 rc 0.
저장소 전체 strict 결과는 판정이 아니라 advisory 한 줄로만 남긴다.

왜 나눴는가: openspec 1.16.0 의 `validate --all --strict` 는 **남의** change 와 main spec 의 경고까지 실패로 쳐서
(2026-11-16: 51/59 실패, 전부 WARNING) 닫는 change 가 strict 로 valid 여도 어떤 change 도 archive 할 수 없었음.

판정은 사람이 읽는 문구가 아니라 `--json` 구조로 읽음. 읽을 수 없거나, 다른 저장소를 봤거나, 아무것도 안 봤으면 FAIL.
사유 앞의 `[V1]`~`[V12]` 는 openspec/changes/a190-…/design.md D3 갈래표의 행 번호임.

사용법: python3 tools/sdd/check_openspec_validate.py <change-id>   (tools/gate.sh 11단계가 부름)
종료 코드: 0 = 통과, 1 = 실패, 2 = 사용법 오류
"""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TIMEOUT_SECONDS = 300
ADVISORY = "advisory — 판정 아님: 저장소 전체 strict"


def run(args: list[str], root: Path, timeout: float):
    """openspec 을 저장소 루트에서 돌려 (rc, 파싱된 JSON) 또는 (None, 사유) 를 돌려줌.

    실행 실패와 시한 초과는 「못 본 것」이라 판정 불가 사유로 바꿈. TimeoutExpired 는 OSError 가 아니므로 따로 받음.
    """
    try:
        proc = subprocess.run(
            ["openspec", *args], cwd=root, capture_output=True, text=True, check=False, timeout=timeout
        )
    except OSError as exc:
        return None, f"[V1] openspec 을 실행할 수 없음: {exc}"
    except subprocess.TimeoutExpired:
        return None, f"[V2] openspec 이 {timeout}초 안에 끝나지 않음"
    try:
        body = json.loads(proc.stdout)
    except ValueError:
        return None, f"[V3] openspec 출력이 JSON 이 아님(rc={proc.returncode}): {proc.stdout[:200]!r} {proc.stderr[:200]!r}"
    if not isinstance(body, dict):
        return None, f"[V3] openspec 출력이 JSON 객체가 아님(rc={proc.returncode})"
    return proc.returncode, body


def shape_faults(body: dict, root: Path) -> list[str]:
    """보고서가 판정할 수 있는 모양인지 봄 — 항목 목록, 오류 status, 검증한 저장소, valid 의 타입."""
    statuses = [s for s in body.get("status", []) if isinstance(s, dict)] if isinstance(body.get("status"), list) else []
    items = body.get("items")
    if not isinstance(items, list):
        said = "; ".join(f"{s.get('code')}: {s.get('message')}" for s in statuses) or "items 없음"
        return [f"[V4] openspec 이 항목을 내지 않음 — {said}"]
    errors = [s for s in statuses if s.get("severity") == "error"]
    if errors:
        return ["[V5] openspec 이 오류 status 를 냄 — " + "; ".join(f"{s.get('code')}: {s.get('message')}" for s in errors)]
    reported = body.get("root", {}).get("path") if isinstance(body.get("root"), dict) else None
    if not isinstance(reported, str):
        return ["[V6] 보고서에 검증한 저장소(root.path)가 없음"]
    # openspec 은 실경로를 냄. 양쪽 다 실경로로 맞춰야 심링크 경로에서 거짓 FAIL 이 안 남.
    if Path(reported).resolve() != root.resolve():
        return [f"[V6] openspec 이 다른 저장소를 검증함: {reported} (이 저장소: {root.resolve()})"]
    bad = [i for i in items if not isinstance(i, dict) or not isinstance(i.get("valid"), bool)]
    if bad:
        return [f"[V7] valid 가 bool 이 아닌 항목 {len(bad)}개: {bad[:3]!r}"]
    return []


def issue_lines(entry: dict) -> str:
    return "; ".join(f"{i.get('level')} {i.get('path')}: {i.get('message')}" for i in entry.get("issues", []) if isinstance(i, dict))


def agreement_faults(rc: int, items: list[dict]) -> list[str]:
    """rc 와 JSON 이 같은 말을 하는지 봄. rc≠0 이면 invalid 를 나열, 둘이 어긋나면 불일치로 FAIL."""
    invalid = [i for i in items if i["valid"] is not True]
    if rc != 0 and invalid:
        return [f"[V11] {i.get('type')} {i.get('id')}: {issue_lines(i)}" for i in invalid]
    if rc != 0:
        return [f"[V12] rc={rc} 인데 JSON 은 전부 valid 라고 함"]
    if invalid:
        return [f"[V12] rc=0 인데 JSON 은 invalid 항목 {len(invalid)}개를 말함: " + ", ".join(str(i.get("id")) for i in invalid)]
    return []


def closing_verdict(change_id: str, root: Path, timeout: float = TIMEOUT_SECONDS):
    """(가) 닫는 change 하나를 strict 로. `--type change` 는 붙이지 않음 — 붙이면 없는 id 가 「delta 없음」으로 바뀌어 못 찾음을 못 가림."""
    rc, body = run(["validate", change_id, "--strict", "--no-interactive", "--json"], root, timeout)
    if rc is None:
        return False, [body]
    faults = shape_faults(body, root)
    if faults:
        if faults[0].startswith("[V4]"):
            faults[0] += (
                " (아카이브된 change 는 openspec 이 검증하지 않는다 — gate 는 archive 전에 돌린다; "
                "같은 이름의 spec 과 겹치면 change 이름을 바꿔야 한다)"
            )
        return False, faults
    items = body["items"]
    if len(items) != 1 or items[0].get("id") != change_id or items[0].get("type") != "change":
        named = [(i.get("type"), i.get("id")) for i in items]
        return False, [f"[V8] 보고서가 정확히 이 change 하나를 말하지 않음: {named}"]
    faults = agreement_faults(rc, items)
    return (not faults), faults


def repository_verdict(change_id: str, root: Path, timeout: float = TIMEOUT_SECONDS):
    """(나) 저장소 전체를 strict 없이 — 오류만 거부. 0 항목이나 닫는 change 를 빠뜨린 검사는 본 것이 아님."""
    rc, body = run(["validate", "--all", "--no-interactive", "--json"], root, timeout)
    if rc is None:
        return False, [body]
    faults = shape_faults(body, root)
    if faults:
        return False, faults
    items = body["items"]
    # 전칭 판정은 표본이 0 이면 자동으로 참이 됨 — 아무것도 안 본 rc 0 을 통과로 치지 않음.
    if not items:
        return False, ["[V9] 저장소 전체 검증이 0 항목을 검사함"]
    if not any(i.get("id") == change_id and i.get("type") == "change" for i in items):
        return False, [f"[V10] 저장소 전체 검증 {len(items)} 항목에 닫는 change {change_id} 가 없음"]
    faults = agreement_faults(rc, items)
    return (not faults), faults


def advisory(root: Path, timeout: float = TIMEOUT_SECONDS) -> str:
    """저장소 전체 strict 합계 한 줄. 무엇이 일어나도 판정에 쓰지 않음."""
    rc, body = run(["validate", "--all", "--strict", "--no-interactive", "--json"], root, timeout)
    try:
        totals = body["summary"]["totals"]
        return f"{ADVISORY} {totals['passed']} passed, {totals['failed']} failed ({totals['items']} items) — 목록: make validate"
    except (TypeError, KeyError):
        return f"{ADVISORY} 합계를 얻지 못함 — rc={rc}, {str(body)[:200]}"


def main(argv: list[str]) -> int:
    if len(argv) != 1 or not argv[0] or argv[0].startswith("-"):
        print("usage: python3 tools/sdd/check_openspec_validate.py <change-id>", file=sys.stderr)
        return 2
    change_id = argv[0]
    # 둘 다 평가해 사유를 전부 내고, 그 뒤 advisory 를 항상 출력함.
    verdicts = (
        ("(가)", f"{change_id} strict valid", closing_verdict(change_id, ROOT)),
        ("(나)", "저장소 전체(strict 없음) 오류 0", repository_verdict(change_id, ROOT)),
    )
    for label, what, (ok, reasons) in verdicts:
        if ok:
            print(f"{label} OK: {what}")
        for reason in reasons:
            print(f"{label} FAIL {reason}", file=sys.stderr)
    print(advisory(ROOT))
    return 0 if all(ok for _, _, (ok, _) in verdicts) else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
