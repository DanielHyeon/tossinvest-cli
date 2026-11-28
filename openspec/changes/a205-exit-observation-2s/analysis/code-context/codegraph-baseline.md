# a205 CodeGraph baseline

- 도구: codegraph **1.6.0**. 색인: primary checkout `/mnt/D/Axipient/workspace/TossOS`(HEAD `e2921fc00` = 이 change 의 base, 2026-11-27 — 파일 3,348 · 노드 57,787 · 엣지 205,610).
  이 worktree(`~/tossos-work/a205`)에는 색인을 만들지 않음 — base 와 같은 커밋이라 primary 색인이 그 커밋의 증거임.
- 질의 원문 출력: `codegraph-callers.txt`.

| 질의 | 결과 |
|---|---|
| `callers DefaultExitObservationInterval` | `ExitObserver.Interval`(exitloop.go:401) 1 |
| `callers ExitObserver.Interval` | `ExitObserver.Run`(exitloop.go:428) 1 — `Tracer.Run` 의 `t.opts.Observer.Interval()` 은 안 잡힘(evidence-reconciliation) |
| `callers lineRefreshInterval` | `lineRefreshSeconds`(line_cadence.go:73) 1 |
| `callers engineRuntimeInterval` | `newPositionPolicyCache`(position_policy_cache.go:201) 1 |
| `callers externalBuyListTimeout` | `readExternalBuyRows`(a197_external_opposite_buy.go:258) 1 — 값 무변경, 주석만 |
| `callers lowerLimitReadTimeout` | `exitLowerLimitBoard.refresh`(exit_lower_limit.go:185) 1 — 무변경(design D2) |
| `affected exitloop.go line_cadence.go position_policy_cache.go a197_external_opposite_buy.go --filter '*_test.go'` | 시험 파일 1,300여 줄 · 100+ 디렉터리 — 패키지 import 사슬 전체라 시험 선택에 못 씀(design §6) |

a166 census(`internal/accountbudgetruntime` 의 shapes.json)는 배선 함수 본문을 고정함 — 이 change 는 `cmd/tossctl` · 배선 함수를 편집하지 않으므로
해당 없음(`git diff --name-only` 에 `cmd/` 0, 구현 뒤 확인).
