#!/usr/bin/env python3
"""task 6.1 부모 — 4.4 의 위조 셋이 **지금 코드에서도** 막히는지 실행으로 잰다.

6.1.2.4 는 셋을 "4.4 가 쓴 스크립트 그대로" 돌렸다고 적었지만 그 스크립트는 저장소에 없다
([[harness-cited-but-never-committed]]). 여기는 같은 세 모양을 `test_check_analysis` 의 픽스처 헬퍼로
다시 만든다 — 헬퍼는 스위트가 매 판 쓰는 것이라 저장소와 같이 늙는다.

모양마다 세 판을 잰다:
    대조군 — 선언 없음 (`check`)
    선언   — 위조 값을 `landed-commit.txt` 로 커밋한 뒤 (`check`)
    기록   — 선언 대신 `--record-landing`(`record_landing`) 이 무엇을 계산하는가

    python3 61_forgeries.py
"""
import hashlib
import json
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))

import check_analysis  # noqa: E402
import test_check_analysis as t  # noqa: E402


def _two_files(raw):
    """R → P(base) → L(`own.go` · `other.go` 둘 다 바뀐다). `(root, marks, own, other)`."""
    root = t._init_fixture(raw)
    own = root / "internal" / "own.go"
    other = root / "internal" / "other.go"
    own.parent.mkdir(parents=True)
    own.write_text("package internal\nfunc Own() int { return 0 }\n")
    other.write_text("package internal\nfunc Other() int { return 0 }\n")
    marks = {"R": t._commit_all(root, "R")}
    own.write_text("package internal\nfunc Own() int { return 1 }\n")
    other.write_text("package internal\nfunc Other() int { return 1 }\n")
    marks["P"] = t._commit_all(root, "P: base")
    change = root / "openspec" / "changes" / "mine"
    change.mkdir(parents=True)
    (change / "base-commit.txt").write_text(marks["P"] + "\n")
    (change / "review.md").write_text("mine\n")
    return root, marks, own, other, change


def build_a():
    """(a) 증거를 **base 상태**로 쓰고 `landed-commit = base` — 스위트의 `_base_shaped_forgery_fixture` 그대로."""
    raw, root, marks = t._base_shaped_forgery_fixture()
    return raw, root, marks["P"]


def build_b():
    """(b) 안 건드린 파일의 **미끼 번들** 하나 + `landed-commit = base`. `other.go` 는 base 뒤로 안 바뀐다."""
    import tempfile
    raw = tempfile.TemporaryDirectory()
    root, marks, own, other, change = _two_files(raw)
    own.write_text("package internal\nfunc Own() int { return 2 }\n")
    marks["L"] = t._commit_all(root, "L: the change's Go work (Own)")
    t._write_evidence(change, package="internal", function="Other", relative="internal/other.go",
                      digest=hashlib.sha256(other.read_bytes()).hexdigest())
    t._commit_all(root, "E: a decoy bundle for an untouched file")
    return raw, root, marks["P"]


def build_c(declare: str = "E"):
    """(c) 정직한 번들 둘 중 하나를 `revision: base` 로 **relabel** — JSON 한 필드."""
    import tempfile
    raw = tempfile.TemporaryDirectory()
    root, marks, own, other, change = _two_files(raw)
    own.write_text("package internal\nfunc Own() int { return 2 }\n")
    other.write_text("package internal\nfunc Other() int { return 2 }\n")
    marks["L"] = t._commit_all(root, "L: the change's Go work (Own · Other)")
    t._write_evidence(change, package="internal", function="Own", relative="internal/own.go",
                      digest=hashlib.sha256(own.read_bytes()).hexdigest())
    bundle = t._write_evidence(change, package="internal", function="Other", relative="internal/other.go",
                               digest=hashlib.sha256(other.read_bytes()).hexdigest())
    value = json.loads((bundle / "ast.json").read_text())
    value["revision"] = "base"
    (bundle / "ast.json").write_text(json.dumps(value))
    marks["E"] = t._commit_all(root, "E: honest evidence, one bundle relabelled base")
    # 선언 값 셋을 잰다: 6.1.2.4 표가 적은 base(P), Go 작업 커밋(L), 그리고 relabel 뒤 남은 번들이 실제로 고정하는
    # 증거 커밋(E). L 행은 6.4 보수(주장정확성 리뷰 P1-F2)가 더했다 — 첫 판은 "(c) 문장이 6.1.2.4 와 다른 것은 원 픽스처를
    # 못 되살린 탓" 이라 적었지만, 같은 픽스처에 L 을 선언하면 6.1.2.4 와 같은 문장이 나온다: 다른 원인은 선언 값이었다.
    return raw, root, marks[declare]


def judge(root: Path) -> str:
    context: dict = {}
    errors = check_analysis.check("mine", root, context)
    required = context.get("required_count", "?")
    landing = str(context.get("landing", ""))[:12] or "worktree"
    return f"{errors or '[]'} · target {landing} · required {required}"


def run(name: str, build) -> None:
    raw, root, declared = build()
    with raw:
        control = judge(root)
        code, lines = check_analysis.record_landing("mine", root)
        recorded = f"rc {code} · {lines}"
        # 기록 명령이 파일을 썼으면 지운다 — 선언 판은 위조 값만 담아야 한다.
        written = root / "openspec" / "changes" / "mine" / check_analysis.LANDING_FILE
        if written.exists():
            written.unlink()
        t._declare_landing(root / "openspec" / "changes" / "mine", declared, "declare the forged landing")
        after = judge(root)
    print(f"## {name} (선언 값 {declared[:12]})\n  대조군 : {control}\n  기록   : {recorded}\n  선언 뒤: {after}\n")


if __name__ == "__main__":
    print(f"HEAD {check_analysis._head_commit(REPO)[:12]}\n")
    run("(a) base 상태 증거 + landed = base", build_a)
    run("(b) 미끼 번들 + landed = base", build_b)
    run("(c) relabel revision: base + landed = base", lambda: build_c("P"))
    run("(c\") relabel revision: base + landed = Go 작업 커밋 L", lambda: build_c("L"))
    run("(c') relabel revision: base + landed = 증거 커밋", build_c)
