# 옛 base(`285c7619`) 기준 FLM 번들 36개 — 역사 자리

1~20판 설계의 분기 근거였던 번들이다. 구현 로트(tasks 25.2, 2026-09-29)가 base를 `721d0338`로 재고정했다. 그러면서 이 번들들의
`source_sha256`이 현재 소스와 맞지 않게 됐다(`check_analysis` stale 29). 21판 이후의 분기 근거는 `analysis/head-ast-21/`이다.

- 게이트(`check_analysis`)는 `analysis/function-logic/`만 읽는다. 그래서 이 자리는 게이트 밖이다.
- 구현 로트가 편집하는 기존 함수의 번들은 새 base에서 `analysis/function-logic/`에 새로 만든다.
- 이 디렉터리를 읽는 것은 a092 자기 문서 도구 셋뿐이다(`tools/check_values.py`의 `FLM_ROOT`, `check_values_selftest.py`, `coverage_gate.py`). 경로는 같은 커밋에서 옮겼다.
- 옛 문서의 `analysis/function-logic/…` 경로 인용은 역사 기록이다. 같은 번들이 이 자리에 있다.
