#!/usr/bin/env python3
"""task 7.5.10 영수증 — 원장의 지문이 **서로 다른 두 디스크 상태**에 같은 값을 내는가.

재확인은 같은 경로의 지문을 판정 때와 끝에서 견준다. 두 상태가 같은 지문을 내면 판정 중 그 사이를 오가도 못 본다.
여기서 한 **같은 뿌리** 안에서 상태 A → B 로 바꾸며 지문을 잰다(스크래치 디렉터리 — 저장소는 건드리지 않는다).

- 목록(`_listing_outcome`): A = 파일 `a` · `b`, B = 파일 하나 `a\\tf\\nb` (편집 전 인코딩 `name\\t{d|f}` 을 `\\n` 으로 잇기).
- 순회(`_pattern_outcome`, 편집 전에만 있다): A = `p/a_test.go` · `p/b_test.go`, B = 파일 하나. B 의 경로 **글자**가 A 의 두
  절대경로를 `\\n` 으로 이은 것과 같다 — 디렉터리 `a_test.go\\n` 아래에 뿌리의 절대경로를 다시 세운다.
- 추적 목록(`_tracked_outcome`, 편집 뒤에만 있다): 같은 두 모양을 **한 저장소**에서 차례로 `git add` 한다.

    python3 7510_collide.py [<rev>]     # 기본 1d1e5ca7 (편집 전) — 그리고 워킹트리의 코드를 같이 잰다
"""
import importlib.util
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))


def load(rev: str, scratch: Path):
    if rev == "worktree":
        import check_analysis
        return check_analysis
    path = scratch / f"check_analysis_{rev[:12]}.py"
    path.write_bytes(subprocess.run(["git", "show", f"{rev}:tools/logic-map/check_analysis.py"], cwd=REPO,
                                    capture_output=True, check=True).stdout)
    spec = importlib.util.spec_from_file_location(f"ca_{rev[:12]}", path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


def clear(where: Path) -> None:
    for child in where.iterdir():
        shutil.rmtree(child) if child.is_dir() and not child.is_symlink() else child.unlink()


def main() -> None:
    rev = sys.argv[1] if len(sys.argv) > 1 else "1d1e5ca781d7716e1d5c2afbeb5a127ceb953425"
    with tempfile.TemporaryDirectory() as raw:
        scratch = Path(raw)
        for label in (rev, "worktree"):
            module = load(label, scratch)
            root = scratch / f"root_{label[:8]}"
            root.mkdir()
            print(f"--- {label[:12]}")

            (root / "a").write_text(""); (root / "b").write_text("")
            before = module._listing_outcome(root)[0]
            clear(root)
            (root / "a\tf\nb").write_text("")
            after = module._listing_outcome(root)[0]
            clear(root)
            print(f"listing  A {before[:22]}… B {after[:22]}… {'COLLIDE' if before == after else 'distinct'}")

            if hasattr(module, "_pattern_outcome"):
                (root / "p").mkdir()
                (root / "p" / "a_test.go").write_text(""); (root / "p" / "b_test.go").write_text("")
                before = module._pattern_outcome(root, "*_test.go")[0]
                clear(root)
                deep = root / "p" / "a_test.go\n" / str(root / "p").lstrip("/")
                deep.mkdir(parents=True)
                (deep / "b_test.go").write_text("")
                matched = module._pattern_outcome(root, "*_test.go")[1]
                after = module._pattern_outcome(root, "*_test.go")[0]
                clear(root)
                print(f"pattern  A {before[:22]}… B {after[:22]}… {'COLLIDE' if before == after else 'distinct'} "
                      f"(B matched {len(matched)}: {str(matched[0])!r})")
            else:
                print("pattern  — no `_pattern_outcome` in this revision")

            if hasattr(module, "_tracked_outcome"):
                # 한 저장소의 두 상태 — 뿌리가 같아야 옛 이음의 충돌이 선다(저장소 둘을 견준 첫 판은 뿌리가 달라 공허했다).
                environment = {**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null"}
                repo = root / "repo"
                repo.mkdir()
                subprocess.run(["git", "init", "-q"], cwd=repo, check=True, env=environment)
                (repo / "a_test.go").write_text("package x\n"); (repo / "b_test.go").write_text("package x\n")
                subprocess.run(["git", "add", "-A"], cwd=repo, check=True, env=environment)
                before, listed_before = module._tracked_outcome(repo)
                subprocess.run(["git", "rm", "-q", "--cached", "a_test.go", "b_test.go"], cwd=repo, check=True, env=environment)
                (repo / "a_test.go").unlink(); (repo / "b_test.go").unlink()
                deep = repo / "a_test.go\n" / str(repo).lstrip("/")
                deep.mkdir(parents=True)
                (deep / "b_test.go").write_text("package x\n")
                subprocess.run(["git", "add", "-A"], cwd=repo, check=True, env=environment)
                after, listed_after = module._tracked_outcome(repo)
                same_text = "\n".join(map(str, listed_before)) == "\n".join(map(str, listed_after))
                print(f"tracked  A {before[:22]}… B {after[:22]}… {'COLLIDE' if before == after else 'distinct'} "
                      f"(옛 이음으로는 같은 글자: {same_text})")
            else:
                print("tracked  — no `_tracked_outcome` in this revision")


if __name__ == "__main__":
    main()
