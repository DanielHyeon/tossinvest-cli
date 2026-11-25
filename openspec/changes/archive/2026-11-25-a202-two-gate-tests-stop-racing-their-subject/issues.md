# a202 issues

## I1 — 후속 후보: `!inFlight()` 대기 뒤 사후 상태를 단언하는 시험 [후속 · 분석 안 함]

- 범위: `!…inFlight()` 대기 매치 전부(원문 19 매치 — 이 change 가 고친 `TestTheReplacedReaderIsClosed:205` 포함). 그중 대기 직후 단언이 있는 자리는
  design §4 목록(11 자리)이고, 나머지는 시험 끝 정리 대기다.

- **분류**: non-blocking, 이 change 범위 밖(Manager 결정 3, 2026-11-25). **이 change 는 아래 자리를 분석하지 않았다** — 단언이 생산 `attempt` 의
  잠금 밖 효과(닫기 · 부착 보고 로그)에 기대는지, 다른 attachment 타입(a114 · a115)의 `attempt` 가 같은 순서인지 모른다.
- 목록(원문 `analysis/census/inflight-waits.txt`, base `7bc35e5d1`):
  - `cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go:73` → `StrategyRuntimeAbsent`
  - `cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go:183` → (시험 끝 정리 대기)
  - `cmd/tossctl/a109_the_request_path_never_dials_test.go:197` → (정리 대기) · `:215` → 시도 수 · `:254` → `StrategyRuntimeAbsent` · `:288` → 로그 「읽기가 실패했다」 개수
  - `cmd/tossctl/a109_a_cancelled_read_is_not_a_detachment_test.go:152` → 옛 읽기 해제 · `:177` → 로그 「다시 붙었다」 0 개(부재 단언 — 경합이면 거짓 PASS 쪽)
  - `cmd/tossctl/a114_the_console_reattaches_its_lifecycle_test.go:298` · `:312` · `:502` · `:511` · `:822` · `:827` · `:831` → 로그 개수 · 시도 수
  - `cmd/tossctl/a115_the_console_strategy_screen_reattaches_test.go:63` · `:468` · `:523` → 정리 · 경고 개수 · `a115LastTry`
- 수리 방향 후보: 각 자리에서 단언 대상이 잠금 밖에서 쓰이는지 AST 로 본 뒤, 그런 자리만 a202 D2 처럼 그 효과 자체를 기다린다.

## I2 — 다른 활성 change 의 같은 결함 (a166 prototype) [기록 이관]

- `openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype/process_test.go:110` `TestKilledLockOwnerReleases` — 자식 flock 보유 · 타이머 없는
  `select {}`. Manager 결정 2(2026-11-25): 이 change 에서 고치지 않고 a166 `issues.md` 에 한 항목으로 남김(같은 날 추가).

## I3 — 수용된 잔여: D1 의 1분 고아 상한 [수용]

- 부모가 kill 전에 사라지면 자식은 최대 1분 동안 시험 전용 `fixture` 임시 디렉터리의 잠금을 쥔 채 남는다. Manager 결정 3 에서 수용.
