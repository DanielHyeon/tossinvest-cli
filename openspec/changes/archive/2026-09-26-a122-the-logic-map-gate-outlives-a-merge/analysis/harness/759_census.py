#!/usr/bin/env python3
"""task 7.5.9 — 시험 색인 · 인용 해소를 **추적 파일로만** 좁히면 저장소의 어느 번들이 달라지는가.

[[fail-closed-must-name-what-it-rejects]]: 좁히는 가드는 **거부할 정상 입력을 먼저** 센다. 여기 모집단은 저장소의
번들 **전부**(활성 + 아카이브의 `analysis/function-logic/*/`)이고, 각 번들의 `branch-test-map.md` 인용 판정
(`test_citation_errors`)을 세 판으로 돌려 견준다.

- `disk`      : 기준 리비전의 코드 그대로 — 색인과 해소가 디스크 전수(`rglob`)다.
- `simulated` : 같은 코드에 순회(`_globbed`)와 고르기(`_kind`)만 **추적 파일로 거른** 판 — 편집 **전에** 거절을 세는 판이다.
- `new`       : 워킹트리의 코드(편집 뒤). 워킹트리에 `_tracked` 가 없으면 이 열은 건너뛴다.

측정은 순간을 달고 다닌다 ([[a-measurement-carries-its-moment]]): HEAD sha · 추적 `*_test.go` 수 · 디스크 수를 먼저 찍는다.
디스크 수는 이 change 자신의 하네스 사본(`_work/`)이 움직인다 — 그 차이의 경로를 같이 찍는다.

    python3 759_census.py [<before-rev>]    # 기본 HEAD
"""
import importlib.util
import json
import os
import re
import subprocess
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
WORK = Path(__file__).resolve().parent / "_work" / f"759_census.{os.getpid()}"
sys.path.insert(0, str(REPO / "tools" / "logic-map"))


def load(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


def main() -> None:
    before = sys.argv[1] if len(sys.argv) > 1 else "HEAD"
    WORK.mkdir(parents=True, exist_ok=True)
    (WORK / "check_analysis_before.py").write_bytes(subprocess.run(
        ["git", "show", f"{before}:tools/logic-map/check_analysis.py"], cwd=REPO, capture_output=True, check=True,
    ).stdout)
    old = load("ca_before", WORK / "check_analysis_before.py")
    simulated = load("ca_simulated", WORK / "check_analysis_before.py")
    import check_analysis as new
    has_new = hasattr(new, "_tracked")

    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=REPO, capture_output=True, text=True, check=True).stdout.strip()
    tracked_all = {os.fsdecode(raw) for raw in subprocess.run(
        ["git", "ls-files", "-z"], cwd=REPO, capture_output=True, check=True).stdout.split(b"\0") if raw}
    tracked_tests = sorted(path for path in tracked_all if path.endswith("_test.go"))
    disk_tests = sorted(path.relative_to(REPO).as_posix() for path in old._globbed(REPO, "*_test.go")
                        if ".git" not in path.parts)
    extra = sorted(set(disk_tests) - set(tracked_tests))
    print(f"before {before} · HEAD {head[:12]} · tracked *_test.go {len(tracked_tests)} · on disk {len(disk_tests)} "
          f"· disk-only {len(extra)} · tracked-only {len(set(tracked_tests) - set(disk_tests))}")
    for path in extra:
        print(f"  disk-only: {path}")

    # 편집 전 코드에 **추적 거름**만 입힌다 — 순회와 고르기 두 깔때기가 추적 파일만 돌려준다.
    walk, kind = simulated._globbed, simulated._kind

    def tracked_only(path: Path) -> bool:
        try:
            return path.relative_to(REPO).as_posix() in tracked_all
        except ValueError:
            return False

    simulated._globbed = lambda root, pattern: [path for path in walk(root, pattern) if tracked_only(path)]
    simulated._kind = lambda path: kind(path) if tracked_only(path) else ""

    variants = {"disk": old, "simulated": simulated}
    if has_new:
        variants["new"] = new
    indexes = {name: module.test_index(REPO) for name, module in variants.items()}

    changes = REPO / "openspec" / "changes"
    homes = sorted([p for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
                   + [p for p in (changes / "archive").iterdir() if p.is_dir()])
    bundles = differing = 0
    cited = 0
    rows = []
    for home in homes:
        analysis = home / "analysis" / "function-logic"
        if not analysis.is_dir():
            continue
        for target in sorted(p for p in analysis.iterdir() if p.is_dir()):
            ast_path, btm = target / "ast.json", target / "branch-test-map.md"
            if not ast_path.is_file() or not btm.is_file():
                continue
            try:
                value = json.loads(ast_path.read_text(encoding="utf-8"))
                _, relative = old.normalized_source(str(value["file"]), REPO)
            except Exception:
                continue
            text = btm.read_text(encoding="utf-8", errors="replace")
            bundles += 1
            cited += len(set(old.CITED_TEST.findall(text))) + sum(
                len(old.CITED_TEST_LINE.findall(line)) for line in text.splitlines())
            package_dir = (REPO / relative).parent
            # 문장의 꼬리(편집이 "tracked" 를 더한다)는 뺀다 — 여기서 재는 것은 **무엇을** 거절하는가다.
            answers = {name: [re.sub(r"which is not a Go test function.*$", "NOT-A-TEST", error)
                              for error in module.test_citation_errors(
                                  target.name, text, package_dir, REPO, indexes[name])]
                       for name, module in variants.items()}
            if len({json.dumps(answer) for answer in answers.values()}) != 1:
                differing += 1
                rows.append({"bundle": target.relative_to(REPO).as_posix(), **answers})
    print(f"bundles {bundles} · citations {cited} · differing {differing}")
    for row in rows:
        print(json.dumps(row, ensure_ascii=False))


if __name__ == "__main__":
    main()
