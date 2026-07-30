#!/usr/bin/env python3
"""Validate the TossOS PM hierarchy and render deterministic tracker views."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PORTFOLIO = ROOT / "docs" / "pm" / "portfolio"
GENERATED = ROOT / "docs" / "pm" / "generated"


def read(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def inspect_objects(
    kind: str,
    root: Path = PORTFOLIO,
) -> tuple[dict[str, dict], list[str]]:
    values: dict[str, dict] = {}
    errors: list[str] = []
    for path in sorted((root / kind).glob("*.yaml")):
        item = read(path)
        item_id = item.get("id")
        if not isinstance(item_id, str) or not item_id:
            errors.append(f"{path}: missing id")
            continue
        if path.stem != item_id:
            errors.append(f"{path}: filename/id mismatch ({item_id})")
        if item_id in values:
            errors.append(f"{kind}: duplicate id {item_id}")
            continue
        values[item_id] = item
    return values, errors


def objects(kind: str, root: Path = PORTFOLIO) -> dict[str, dict]:
    return inspect_objects(kind, root)[0]


def active_changes(repo_root: Path = ROOT) -> set[str]:
    changes = repo_root / "openspec" / "changes"
    return {
        path.name
        for path in changes.iterdir()
        if path.is_dir() and path.name != "archive" and (path / "proposal.md").exists()
    }


def story_contract(story: dict) -> tuple[str | None, str | None]:
    openspec = story.get("openspec")
    if not isinstance(openspec, dict):
        return None, None
    change_id = openspec.get("change_id")
    path = openspec.get("path")
    return (
        change_id if isinstance(change_id, str) and change_id else None,
        path if isinstance(path, str) and path else None,
    )


def resolve_story_change(
    story: dict,
    repo_root: Path = ROOT,
) -> tuple[str | None, Path | None, str | None]:
    change_id, declared = story_contract(story)
    if change_id is None or declared is None:
        return change_id, None, "missing OpenSpec change_id/path"
    relative = Path(declared)
    if relative.is_absolute() or ".." in relative.parts:
        return change_id, None, f"invalid OpenSpec path {declared}"
    target = repo_root / relative
    if not target.is_dir() or not (target / "proposal.md").exists():
        return change_id, None, f"invalid OpenSpec path {declared}"
    active_root = repo_root / "openspec" / "changes"
    archive_root = active_root / "archive"
    if target.parent == active_root and target.name == change_id:
        return change_id, target, None
    if (
        target.parent == archive_root
        and re.sub(r"^\d{4}-\d{2}-\d{2}-", "", target.name) == change_id
    ):
        return change_id, target, None
    return change_id, None, f"invalid OpenSpec path {declared}"


def derive_story_status(story: dict, repo_root: Path = ROOT) -> str:
    _, change_path, error = resolve_story_change(story, repo_root)
    if error or change_path is None:
        return "invalid"
    if change_path.parent.name == "archive":
        return "archived"
    tasks = change_path / "tasks.md"
    if not tasks.exists():
        return "designed"
    checkboxes = re.findall(
        r"(?m)^- \[([ xX])\]\s+",
        tasks.read_text(encoding="utf-8"),
    )
    if not checkboxes:
        return "designed"
    if any(value == " " for value in checkboxes):
        return "in_progress"
    return "implemented"


def validate(repo_root: Path = ROOT) -> list[str]:
    root = repo_root / "docs" / "pm" / "portfolio"
    registry = read(root / "_registry.yaml")
    initiatives, initiative_errors = inspect_objects("initiatives", root)
    epics, epic_errors = inspect_objects("epics", root)
    features, feature_errors = inspect_objects("features", root)
    stories, story_errors = inspect_objects("stories", root)
    errors = initiative_errors + epic_errors + feature_errors + story_errors
    if "bootstrap_change_allowlist" in registry:
        errors.append("registry: bootstrap_change_allowlist is forbidden")

    expected = {
        "initiatives": set(initiatives),
        "epics": set(epics),
        "features": set(features),
        "stories": set(stories),
    }
    for key, values in expected.items():
        registered = registry.get(key, [])
        if len(registered) != len(set(registered)):
            errors.append(f"registry {key} contains duplicates")
        if set(registered) != values:
            errors.append(f"registry {key} mismatch")

    for initiative in initiatives.values():
        linked = initiative.get("epics", [])
        if len(linked) != len(set(linked)):
            errors.append(f"{initiative['id']}: duplicate epic forward link")
        for epic_id in linked:
            child = epics.get(epic_id)
            if child is None or child.get("initiative") != initiative["id"]:
                errors.append(f"{initiative['id']}: invalid epic forward link {epic_id}")
    for epic in epics.values():
        parent = initiatives.get(epic.get("initiative"))
        if parent is None or epic["id"] not in parent.get("epics", []):
            errors.append(f"{epic['id']}: initiative reverse link missing")
        linked = epic.get("features", [])
        if len(linked) != len(set(linked)):
            errors.append(f"{epic['id']}: duplicate feature forward link")
        for feature_id in linked:
            child = features.get(feature_id)
            if child is None or child.get("epic") != epic["id"]:
                errors.append(f"{epic['id']}: invalid feature forward link {feature_id}")
    for feature in features.values():
        parent = epics.get(feature.get("epic"))
        if parent is None or feature["id"] not in parent.get("features", []):
            errors.append(f"{feature['id']}: epic reverse link missing")
        linked = feature.get("stories", [])
        if len(linked) != len(set(linked)):
            errors.append(f"{feature['id']}: duplicate story forward link")
        for story_id in linked:
            child = stories.get(story_id)
            if child is None or child.get("feature") != feature["id"]:
                errors.append(f"{feature['id']}: invalid story forward link {story_id}")
    for story in stories.values():
        parent = features.get(story.get("feature"))
        if parent is None or story["id"] not in parent.get("stories", []):
            errors.append(f"{story['id']}: feature reverse link missing")
        if "status" in story:
            errors.append(
                f"{story['id']}: manual status is forbidden; derive it from OpenSpec"
            )

    by_change: dict[str, list[str]] = {}
    for story in stories.values():
        change, _, contract_error = resolve_story_change(story, repo_root)
        if contract_error:
            errors.append(f"{story['id']}: {contract_error}")
        if change is not None:
            by_change.setdefault(change, []).append(story["id"])
    for change, linked in by_change.items():
        if len(linked) != 1:
            errors.append(f"{change}: expected one story, got {linked}")
    for change in sorted(active_changes(repo_root) - set(by_change)):
        errors.append(f"{change}: active change has no Story")
    return errors


def render(repo_root: Path = ROOT) -> dict[str, str]:
    root = repo_root / "docs" / "pm" / "portfolio"
    initiatives = objects("initiatives", root)
    epics = objects("epics", root)
    features = objects("features", root)
    stories = objects("stories", root)
    master = ["# TossOS master tracker", ""]
    for initiative in initiatives.values():
        master.append(
            f"## {initiative['id']} — {initiative['title']} "
            f"[{initiative.get('intent', 'unspecified')}]"
        )
        for epic_id in initiative.get("epics", []):
            epic = epics[epic_id]
            master.append(
                f"- {epic['id']} — {epic['title']} "
                f"[{epic.get('intent', 'unspecified')}]"
            )
            for feature_id in epic.get("features", []):
                feature = features[feature_id]
                master.append(
                    f"  - {feature['id']} — {feature['title']} "
                    f"[{feature.get('intent', 'unspecified')}]"
                )
                for story_id in feature.get("stories", []):
                    story = stories[story_id]
                    change_id, _ = story_contract(story)
                    master.append(
                        f"    - {story['id']} — {story['title']} "
                        f"[{derive_story_status(story, repo_root)}] "
                        f"→ `{change_id or 'unbound'}`"
                    )
    changes = ["# OpenSpec change map", "", "| Change | Story | Status |", "|---|---|---|"]
    for story in stories.values():
        change_id, _ = story_contract(story)
        if change_id:
            changes.append(
                f"| `{change_id}` | {story['id']} | "
                f"{derive_story_status(story, repo_root)} |"
            )
    readiness = ["# Release readiness", "", "| Story | Acceptance checks |", "|---|---|"]
    for story in stories.values():
        readiness.append(f"| {story['id']} | {len(story.get('acceptance', []))} |")
    return {
        "00-master-tracker.md": "\n".join(master) + "\n",
        "01-active-change-map.md": "\n".join(changes) + "\n",
        "02-release-readiness.md": "\n".join(readiness) + "\n",
    }


def generate(repo_root: Path = ROOT, check: bool = False) -> list[str]:
    errors = validate(repo_root)
    if errors:
        return errors
    output = GENERATED if repo_root == ROOT else repo_root / "docs" / "pm" / "generated"
    expected = render(repo_root)
    if check:
        for name, content in expected.items():
            path = output / name
            if not path.exists() or path.read_text(encoding="utf-8") != content:
                errors.append(f"generated tracker stale: {name}")
        return errors
    output.mkdir(parents=True, exist_ok=True)
    for name, content in expected.items():
        (output / name).write_text(content, encoding="utf-8")
    return []


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default=str(ROOT))
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    errors = generate(Path(args.root), args.check)
    if errors:
        print("[pm] invalid:")
        for error in errors:
            print(f"  - {error}")
        return 1
    print("[pm] hierarchy and generated trackers are current")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
