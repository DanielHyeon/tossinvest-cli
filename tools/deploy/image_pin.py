#!/usr/bin/env python3
"""빌드가 지우려는 이름을, 지우기 전에 부른다.

`docker compose build` 는 `tossos:local` 태그를 새 이미지로 옮기고 직전 이미지는
태그를 잃는다. 태그 없는 이미지는 다음 prune 에 사라지고, 그러면 롤백 대상이
사라진다. 2026-08-11 · 08-12 · 08-13 세 번 그렇게 잃었다.

세 번 다 원인은 같았다: 사람이 실행하는 두 명령(`build` → `up -d`) 사이에
문서에만 있는 한 줄(`docker tag …`)이 끼어 있었다. 문서를 한 번 더 쓰는 것은
2026-08-12 에 이미 해 봤고 26시간 뒤에 다시 빠졌다.

그래서 판단을 여기에 둔다. Makefile recipe 안에 쓴 판단은 어떤 테스트도 닿을 수
없는 판단이기 때문이다 — `cmd/tossctl/soakautostart.go:78-81` 이 같은 이유로 같은
선택을 한다.
"""

import argparse
import re
import sys
from dataclasses import dataclass

# docker 가 받는 tag 문자 집합. 여기서 거부해야 실패가 빌드 *앞*에 온다.
_TAG_COMPONENT = re.compile(r"^[A-Za-z0-9_][A-Za-z0-9._-]*$")

# git 이 없을 때 Makefile 의 COMMIT 이 떨어지는 값. 이 값으로 만든 핀은 두 빌드가
# 같은 이름을 갖게 한다 — 핀이 아니라 덮어쓰기다.
_UNKNOWN_COMMIT = "unknown"

# 빌드가 옮겨 가는 태그. 이것만 달고 있는 이미지는 이 빌드에 이름을 잃는다.
MOVING_TAG = "tossos:local"

# docker 가 태그 없는 이미지를 적는 방식.
_DANGLING = "<none>:<none>"


class PinNameError(ValueError):
    """핀 이름을 만들 수 없다."""


@dataclass(frozen=True)
class Verdict:
    would_orphan: bool
    reason: str


def orphan_verdict(repo_tags) -> Verdict:
    """지금 `tossos:local` 인 이미지가 이 빌드에 이름을 잃는가.

    `repo_tags` 는 그 이미지의 RepoTags 다. 빈 목록은 "아직 이미지가 없다"이고
    첫 빌드에는 잃을 것이 없다.
    """
    tags = [t.strip() for t in repo_tags if t and t.strip()]
    if not tags:
        return Verdict(False, "직전 이미지가 없다 — 이 빌드가 잃을 이름이 없다")

    durable = [t for t in tags if t != MOVING_TAG and t != _DANGLING]
    if durable:
        return Verdict(False, "직전 이미지가 이미 이름을 갖고 있다: " + ", ".join(sorted(durable)))

    if tags == [_DANGLING]:
        return Verdict(True, "직전 이미지에 이름이 없다 — 이미 되돌릴 수 없는 상태다")

    return Verdict(
        True,
        f"직전 이미지가 {MOVING_TAG} 하나만 달고 있다 — 이 빌드가 그 이름을 가져가면 "
        "되돌릴 대상이 사라진다",
    )


def pin_name(change: str, commit: str) -> str:
    """무엇을 되돌리는 것인지 말하는 이름을 만든다."""
    change = (change or "").strip()
    commit = (commit or "").strip()
    if not change:
        raise PinNameError("CHANGE 가 없다 — 핀이 무엇을 되돌리는 것인지 말하지 않는다")
    if not commit:
        raise PinNameError("COMMIT 이 없다 — 핀이 어느 빌드인지 말하지 않는다")
    if commit == _UNKNOWN_COMMIT:
        raise PinNameError(
            f"COMMIT 이 {_UNKNOWN_COMMIT} 다 — 두 빌드가 같은 이름을 갖게 되므로 "
            "핀이 아니라 덮어쓰기가 된다"
        )
    tag = f"{change}-{commit}"
    if not _TAG_COMPONENT.match(tag):
        raise PinNameError(f"docker 태그로 쓸 수 없는 값이다: {tag!r}")
    return f"tossos:{tag}"


def _main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    guard = sub.add_parser("guard", help="이 빌드가 직전 이미지의 이름을 지우는지 본다")
    guard.add_argument(
        "--tags",
        default="",
        help="직전 tossos:local 이미지의 RepoTags, 쉼표 구분. 이미지가 없으면 빈 값",
    )

    name = sub.add_parser("name", help="핀 이름을 만든다")
    name.add_argument("--change", required=True)
    name.add_argument("--commit", required=True)

    args = parser.parse_args(argv)

    if args.command == "guard":
        verdict = orphan_verdict(args.tags.split(","))
        if verdict.would_orphan:
            print(f"거부: {verdict.reason}", file=sys.stderr)
            print(
                "먼저 그 이미지에 이름을 준다:\n"
                "  docker tag $(docker image inspect tossos:local --format '{{.Id}}') "
                "tossos:<직전-change>-<직전-commit>",
                file=sys.stderr,
            )
            return 1
        print(verdict.reason)
        return 0

    try:
        print(pin_name(args.change, args.commit))
    except PinNameError as err:
        print(f"거부: {err}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(_main())
