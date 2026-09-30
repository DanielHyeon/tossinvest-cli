# Branch Test Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go` (48-54); **편집 뒤** 측정 — `analysis/harness/branch_coverage.py`, 연결 워크트리 `c6e2e3ac`,
  `./internal/obs` 시험 18개(`Announc|Mode|TestA092`)를 하나씩(`analysis/harness/coverage-post-mode.json`). 편집 전 표는 `analysis/pre-edit/unit2/`.
- 재번호: 편집 전 B1(:50) → 편집 뒤 B1(:49). 편집 전 B2(:54)는 `operatingModeEvent`로 옮겨져 이 함수에 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 49:2 | `n == nil` → nil | `TestAnnouncingWithoutANotifierIsSafe` | 편집 전 기준선(분기 불변) | 블록 49.14-51.3을 시험 1개가 실행, PASS |
