# Branch Test Map: `runAuxiliaryBody`

Source: `internal/app/engine/auxiliary.go` (130-137).

| Branch | Scenario | Test |
|---|---|---|
| B1 | panic is recovered into an auxiliary stop | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |

Planned A100 RED: a worker panic remains observable and does not silently restart, stop unrelated loops, or leave a circular entry latch.

> **a100 R0 재동결(2026-10-04) — 좌표 이동.** 옛 · 새 AST 의 분기 (id · 종류) 목록이 같아 ast.json 을 현재 소스 추출로 바꾸고 `(start-end)` 범위 · 파일 SHA-256 만 옮겼다(`analysis/harness/r0_refresh_bundle.py shift`). 산문 · 시험 인용은 손대지 않았다. 옛 ast 는 returns/calls 를 싣지 않던 추출기 판본이라 그 둘은 대조 대상이 아니다.
