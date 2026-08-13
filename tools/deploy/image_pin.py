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
선택을 한다. (그 논거가 성립하려면 이 파일이 실제로 테스트에서 돌아야 한다.
`make sdd-test` 와 `compileall` 이 `tools/deploy` 를 포함한다.)

## 세 상태를 구분한다

가드의 입력은 "태그 목록"이 아니라 **관측 상태**다.

- `present` — docker 가 이미지를 보여줬다. RepoTags 가 증거다.
- `absent`  — 그런 이미지가 없다. 첫 빌드다. 잃을 것이 없다.
- `unknown` — docker 를 읽지 못했다. 데몬 정지, 미설치, 소켓 권한, 잘못된
              DOCKER_HOST. **못 본 것은 안전하지 않다** — 거부한다.

초판은 `2>/dev/null || true` 로 셋을 하나로 접었고, 그래서 데몬이 죽어 있을 때
가드가 "잃을 게 없다"고 통과시켰다. 검사가 못 본 것을 안전하다고 보고하면
그것은 가드가 아니다.
"""

import argparse
import json
import os
import re
import sys
from dataclasses import dataclass

# docker 가 받는 tag 문자 집합. 여기서 거부해야 실패가 빌드 *앞*에 온다.
_TAG_COMPONENT = re.compile(r"^[A-Za-z0-9_][A-Za-z0-9._-]*$")

# docker 태그 성분의 상한. 넘으면 빌드가 끝난 뒤에 docker 가 거부한다.
_TAG_MAX = 128

# git 이 없을 때 Makefile 의 COMMIT 이 떨어지는 값. 이 값으로 만든 핀은 두 빌드가
# 같은 이름을 갖게 한다 — 핀이 아니라 덮어쓰기다.
_UNKNOWN_COMMIT = "unknown"

# 빌드가 옮겨 가는 태그. 이것만 달고 있는 이미지는 이 빌드에 이름을 잃는다.
MOVING_TAG = "tossos:local"

# 관측 상태.
PRESENT = "present"
ABSENT = "absent"
UNKNOWN = "unknown"

# CHANGE·COMMIT 은 명령줄이 아니라 환경으로 받는다. `make` 가 값을 recipe 문자열에
# 끼워 넣으면 셸이 먼저 해석하므로, 따옴표가 든 값은 여기 도착하기 전에 이미
# 빠져나간다 — 그때 이 파일의 검증은 아무것도 막지 못한다.
ENV_CHANGE = "TOSSOS_PIN_CHANGE"
ENV_COMMIT = "TOSSOS_PIN_COMMIT"


class PinNameError(ValueError):
    """핀 이름을 만들 수 없다."""


@dataclass(frozen=True)
class Verdict:
    refuse: bool
    reason: str


def orphan_verdict(state, repo_tags, pin_taken: bool = False) -> Verdict:
    """이 빌드를 해도 롤백 대상이 남는가.

    `state` 는 PRESENT/ABSENT/UNKNOWN. `repo_tags` 는 PRESENT 일 때만 증거다.
    `pin_taken` 은 이번에 박을 이름이 이미 쓰이고 있다는 뜻이다.
    """
    if state == UNKNOWN:
        return Verdict(True, "docker 를 볼 수 없다 — 무엇을 잃는지 모르는 채로 빌드하지 않는다")

    if state == ABSENT:
        # 첫 빌드. 잃을 것이 없다. 다만 이름 충돌은 여전히 볼 수 있다.
        if pin_taken:
            return Verdict(True, _pin_taken_reason())
        return Verdict(False, "직전 이미지가 없다 — 이 빌드가 잃을 이름이 없다")

    if state != PRESENT:
        # 호출자가 새 상태를 만들고 여기를 안 고친 경우. 모르는 상태는 안전하지 않다.
        return Verdict(True, f"모르는 관측 상태다: {state!r} — 판단하지 않고 거부한다")

    if pin_taken:
        return Verdict(True, _pin_taken_reason())

    tags = [t.strip() for t in repo_tags if t and t.strip()]
    if not tags:
        return Verdict(
            True,
            "직전 이미지에 이름이 없다 — 이미 되돌릴 수 없는 상태이고 이 빌드가 그것을 굳힌다",
        )

    durable = [t for t in tags if t != MOVING_TAG]
    if durable:
        return Verdict(False, "직전 이미지가 이미 이름을 갖고 있다: " + ", ".join(sorted(durable)))

    return Verdict(
        True,
        f"직전 이미지가 {MOVING_TAG} 하나만 달고 있다 — 이 빌드가 그 이름을 가져가면 "
        "되돌릴 대상이 사라진다",
    )


def _pin_taken_reason() -> str:
    # docker tag 는 이름을 만드는 게 아니라 *옮긴다*. 같은 커밋에서 다시 빌드하면
    # (더러운 트리, --amend, 실패 후 재시도) 방금 "안전하다"고 인정한 그 이름을
    # 직전 이미지에서 떼어 온다. 그 순간 직전 이미지는 이름을 잃는다.
    return (
        "박으려는 이름이 이미 있다 — docker tag 는 이름을 옮기므로 그 이미지가 "
        "이름을 잃는다. 커밋을 새로 만들거나, 그 핀을 의도적으로 버린 뒤 다시 실행한다"
    )


def pin_name(change: str, commit: str) -> str:
    """무엇을 되돌리는 것인지 말하는 이름을 만든다."""
    change = (change or "").strip()
    commit = (commit or "").strip()
    if not change:
        raise PinNameError(f"{ENV_CHANGE} 가 없다 — 핀이 무엇을 되돌리는 것인지 말하지 않는다")
    if not commit:
        raise PinNameError(f"{ENV_COMMIT} 이 없다 — 핀이 어느 빌드인지 말하지 않는다")
    if commit == _UNKNOWN_COMMIT:
        raise PinNameError(
            f"COMMIT 이 {_UNKNOWN_COMMIT} 다 — 두 빌드가 같은 이름을 갖게 되므로 "
            "핀이 아니라 덮어쓰기가 된다"
        )
    tag = f"{change}-{commit}"
    if len(tag) > _TAG_MAX:
        raise PinNameError(
            f"docker 태그 성분은 {_TAG_MAX}자까지다 — {len(tag)}자다. 빌드가 끝난 뒤에 "
            "거부당하지 않도록 여기서 멈춘다"
        )
    if not _TAG_COMPONENT.match(tag):
        raise PinNameError(f"docker 태그로 쓸 수 없는 값이다: {tag!r}")
    return f"tossos:{tag}"


def _name_from_env() -> str:
    return pin_name(os.environ.get(ENV_CHANGE, ""), os.environ.get(ENV_COMMIT, ""))


def _parse_tags(raw: str):
    """`docker image inspect --format '{{json .RepoTags}}'` 의 출력."""
    raw = (raw or "").strip()
    if not raw:
        return []
    try:
        value = json.loads(raw)
    except json.JSONDecodeError:
        return []
    if not isinstance(value, list):
        return []
    return [t for t in value if isinstance(t, str)]


def _main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    guard = sub.add_parser("guard", help="이 빌드가 롤백 대상을 없애는지 본다")
    guard.add_argument("--state", required=True, choices=[PRESENT, ABSENT, UNKNOWN])
    guard.add_argument("--tags", default="", help="RepoTags 의 JSON 배열. PRESENT 일 때만 읽는다")
    guard.add_argument("--pin-taken", default="no", choices=["yes", "no"])

    sub.add_parser("name", help=f"핀 이름을 만든다 ({ENV_CHANGE}/{ENV_COMMIT} 환경변수)")

    args = parser.parse_args(argv)

    if args.command == "guard":
        verdict = orphan_verdict(
            args.state, _parse_tags(args.tags), pin_taken=args.pin_taken == "yes"
        )
        if verdict.refuse:
            print(f"거부: {verdict.reason}", file=sys.stderr)
            return 1
        print(verdict.reason)
        return 0

    try:
        print(_name_from_env())
    except PinNameError as err:
        print(f"거부: {err}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(_main())
