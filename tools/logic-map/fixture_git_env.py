"""시험 픽스처의 git 이 개발자의 설정을 안 읽게 한다 (a122 task 6.4(h)).

시험 모듈 둘(`test_check_analysis.py` · `test_execution_baseline.py`)이 import 때 `isolate()` 를 부른다 — 규칙을 한 곳에 두어
두 모듈이 갈리지 않게 한다. 막는 것은 **설정의 출처** 셋이다:

- 전역 · 시스템 설정 파일 — `GIT_CONFIG_GLOBAL` · `GIT_CONFIG_SYSTEM` 을 `os.devnull` 로.
- 환경 변수로 주는 설정 — `GIT_CONFIG_PARAMETERS`(`git -c` 가 자식에게 넘기는 것) · `GIT_CONFIG_COUNT` · `GIT_CONFIG_KEY_<n>` ·
  `GIT_CONFIG_VALUE_<n>` 를 지운다. 파일 고정은 이것을 못 막는다(독립 적대 리뷰 실측: 전체 스위트 에러 446).

저장소 **지역** 설정은 그대로 읽힌다 — 사용자 설정을 재는 시험은 전부 픽스처 저장소에 `git config` 로 적는다.
"""

from __future__ import annotations

import os

_FROM_THE_ENVIRONMENT = ("GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT")
_NUMBERED = ("GIT_CONFIG_KEY_", "GIT_CONFIG_VALUE_")


def isolate(environment: "os._Environ[str] | dict[str, str]" = os.environ) -> None:
    """`environment`(기본 이 프로세스의 환경)에서 개발자의 git 설정 출처를 끊는다. 자식 프로세스는 이 환경을 물려받는다."""
    environment["GIT_CONFIG_GLOBAL"] = os.devnull
    environment["GIT_CONFIG_SYSTEM"] = os.devnull
    for key in list(environment):
        if key in _FROM_THE_ENVIRONMENT or key.startswith(_NUMBERED):
            del environment[key]
