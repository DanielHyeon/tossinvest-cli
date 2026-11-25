# a168 Issues

> Manager 결정 2026-11-25(design §10b 「Manager 결정」). 둘 다 이 change 를 막지 않는다 — 기록용 잔여.

## I1 — 만료 모드에서 대상 아닌 CLOSED 행의 `first.status` · `type` 미검사 (F-A1) [후속 change 후보]

- **분류**: non-blocking, 이 change 범위 밖(해독 모호함이 아니라 a164 부터의 판정 의미 공백).
- **내용**: 다른 id 의 `{"status":"EXPIRED","first":{"status":"WATCHING"}}` · `"type":"OCO"` 행이 top-level `EXPIRED` 만으로 통과한다
  (`internal/verifylive/reconcile_expired.go:120-124`, leg 검사는 대상 행만 `:106-109` · `:145-150`). 출처: 리뷰 2단계 1회차 Claude 적대 · Codex.
- **후속 change 후보**: 목록 행에 `Type` 노출 + 모든 CLOSED 행 leg 어휘 검사. **먼저 브로커가 그런 조합을 실제로 내는지 영수증으로 확인**한다.
- 새 change 디렉터리는 만들지 않았다.

## I2 — 병적 본문 비용 (R5 · F-P1) [잔여]

- **분류**: non-blocking 잔여. 깊이 · 크기 상한이나 O(n) 재작성은 넣지 않는다.
- **근거**: 호출자가 `tossctl verify reconcile` 하나뿐이고 엔진 · 주문 경로 영향 0, 정상 100행 53 KB ≈ 5 ms(`analysis/measure/r5_cost-output.txt`).
  병적 입력(깊이 4,000 → 566 ms · 649 MB, 한 객체 키 20,000 → 0.66 s)은 design §9 R5 에 기록.
- **재고 조건**: 대사 읽기가 엔진 경로로 옮겨지거나 비공식 입력을 받게 될 때 — 별도 change.

## I3 — 후속: `TestChildLockAndCrash` 자식 수명 경합 수리 (a169 소유 시험) [후속]

- **분류**: non-blocking, 이 change 범위 밖(a168 은 `internal/shadowsession` 을 건드리지 않음). 완료 게이트 1회차 ⑦ 에서 드러남(`analysis/close/gate-1.txt`).
- **원인(Manager 실측 2026-11-25, main `e24966ac1`)**: 자식 갈래의 `select {}` 가 시험 goroutine 을 막고, `-test.run` 만 받은 자식에는 timeout 타이머가 없어
  `READY` 직후 `fatal error: all goroutines are asleep - deadlock!` 로 exit 2 — 죽으면 flock 이 풀린다. 부하 속에서 자식이 먼저 죽으면 부모 `Begin` 이
  잠금을 얻고 state `active` 를 보고 `ErrInvalid`(`session_linux.go:305-306`, 의도된 fail-closed)를 낸다.
- **수리 방향**: 자식이 교착 검출에 걸리지 않게 살려 둔다(예: 긴 sleep). 기존 시험 함수 편집이라 별도 change.

## I4 — 후속: `TestTheReplacedReaderIsClosed` 대기 조건 경합 수리 (a109 소유 시험) [후속]

- **분류**: non-blocking, 이 change 범위 밖(a168 의 `cmd/tossctl` 변경 0). 완료 게이트 2회차 ⑦ 에서 드러남(`analysis/close/gate-2.txt`).
- **원인(Manager 가 `cmd/tossctl/httpapi_strategy_attach.go` 를 읽어 확인, main `e24966ac1`)**: `strategyRuntimeAttachment.attempt` 는 잠금 안에서
  `a.trying = false`(183행) 뒤 `a.mu.Unlock()`(210행), 그 다음 줄(211행)에서 잠금 밖으로 `closeEvictedStrategyReader(evicted)` — 의도된 순서. 시험은
  `attempts == 1 && !inFlight()` 까지만 기다리고 곧바로 `closes == 1` 을 단언하므로 210~211행 사이 창에서 0 이 보인다.
- **수리 방향**: 시험의 대기 조건에 닫기를 포함한다. 기존 시험 함수 편집이라 별도 change.
