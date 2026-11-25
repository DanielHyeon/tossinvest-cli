# a202 — 완료 게이트를 흔드는 기존 시험 2건이 자기 대상과 경합하지 않는다

## Why

2026-11-25 a168 종결 때 완료 게이트 ⑦ `make test` 가 두 번, 서로 다른 기존 시험으로 떨어졌다
(a168 archive `analysis/close/gate-1.txt:225-226`, `gate-2.txt:137-138`). 둘 다 단독 재실행은 PASS 였고, Manager 가 원인을
기존 시험의 경합으로 판정해 후속으로 넘겼다(a168 archive `issues.md` I3 · I4). 사용자 승인 2026-11-25:
「경합 시험 2건 수리 change 를 착수할까요? — 시험 코드만 고치고 생산 코드는 건드리지 않습니다」 →「승인」.

1. **`internal/shadowsession` `TestChildLockAndCrash`**(a169 소유 시험). 자식 갈래가 `READY` 를 찍고 `select {}` 로 막는다.
   `-test.run` 만 받은 자식 시험 바이너리에는 시한 타이머가 없어 런타임이 곧바로
   `fatal error: all goroutines are asleep - deadlock!`(exit 2)로 프로세스를 끝내고 flock 이 풀린다. 부모의 `Begin` 이 자식의 죽음보다
   늦으면 잠금을 얻고 state `active` 를 보고 `ErrBusy` 대신 `ErrInvalid` 를 받는다 — 시험이 재려던 「자식이 잠금을 쥔 동안 부모는 Busy」가
   아니라 자식의 수명을 재게 된다. 생산 `Begin` 의 그 거절은 의도된 fail-closed 다(고치지 않는다).
   - 편집 전 재현(`analysis/red/shadowsession-child-before.txt`): 자식 모드를 직접 띄우면 `READY` 뒤 **0.10 s 만에 exit 2**, stderr 첫 줄
     `fatal error: all goroutines are asleep - deadlock!`.
   - 편집 전 재현(`analysis/red/shadowsession-old-test-parent-delay.txt`): 사본에서 부모가 `Begin` 전에 300 ms 늦으면 3/3 FAIL
     `lock: shadowsession: invalid or unavailable state` — 게이트 1회차 메시지와 같다.
2. **`cmd/tossctl` `TestTheReplacedReaderIsClosed`**(a109 소유 시험). 시험은 `attempts == 1 && !inFlight()` 까지만 기다리고 곧바로
   `closes == 1` 을 단언한다. 생산 `strategyRuntimeAttachment.attempt`(`cmd/tossctl/httpapi_strategy_attach.go:183` `a.trying = false`,
   `:210` `a.mu.Unlock()`, `:211` `closeEvictedStrategyReader(evicted)`)는 잠금을 푼 **다음 줄**에서 잠금 밖으로 닫는다 — 의도된 순서다
   (같은 파일 `:204` 주석, `:222-224`). 그래서 `:210` 과 `:211` 사이 창에서 시험이 0 을 본다.
   - 편집 전 재현(`analysis/red/tossctl-old-test-prod-sleep.txt`): 사본에서 `:210` 과 `:211` 사이에 50 ms sleep 을 넣으면 5/5 FAIL
     `자리에서 밀려난 reader 를 0번 닫았다, want 1` — 게이트 2회차 메시지와 같다. 무변이 사본은 5/5 PASS
     (`analysis/red/tossctl-old-test-control.txt`).

## What Changes

- `TestChildLockAndCrash` 의 자식 갈래: 타이머 없는 `select {}` 대신 시한이 걸린 대기(`time.Sleep`)로 부모가 죽일 때까지 살아 있게 하고,
  그 시한이 지나면 스스로 끝난다(부모가 사라져도 고아가 오래 남지 않게). 부모 쪽 단언·순서는 그대로다.
- `TestTheReplacedReaderIsClosed`: 「시도 종료」 대기 뒤에 「밀려난 reader 가 닫힘」 자체를 기존 `a109WaitFor`(2 s 기한)로 기다리고,
  기존 단언 `closes == 1` 은 그대로 둔다 — 닫기가 아예 없으면 여전히 FAIL 한다.
- 생산 코드(`*_test.go` 가 아닌 Go · `schemas/`) 변경 0. 새 가드 · 새 헬퍼 · 새 의존성 0.

## Non-goals

- 생산 `Begin`(`internal/shadowsession/session_linux.go`)의 `ErrInvalid` 거절, `strategyRuntimeAttachment.attempt` 의 닫기 순서 변경.
- 시험 시한(`a109WaitFor` 2 s, `go test -timeout`)을 늘리는 것.
- census 에서 찾은 같은 모양의 다른 자리 중 이 change 디렉터리 밖(다른 활성 change 의 분석 디렉터리)의 것 — 목록만 남긴다(`design.md` §4).
- `!inFlight()` 류 대기 뒤 곧바로 사후 상태를 단언하는 다른 시험 — 목록만 남긴다.

## Capabilities

### Modified Capabilities

- 없음 — 이 change 는 생산 동작과 요구사항을 바꾸지 않는다. spec delta 판단은 `design.md` §5.

## Impact

- 위험 등급: Low. 시험 전용. High-risk 경로(주문·손절·익절·사이징·Guardian·원장·대사·인증·체결)와 게이트 판정 로직을 바꾸지 않는다.
- 영향 파일: `internal/shadowsession/session_linux_test.go`, `cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go`.
- 효과: 완료 게이트 ⑦ · ⑧(같은 시험이 태그 스위트에서도 돈다)의 거짓 실패 두 갈래가 사라진다.

## not-applicable / 적용

- FLM/BTM: **적용**. 기존 시험 함수 두 개의 본문을 고친다 → gate ⑤ 가 함수마다 번들을 요구한다(편집 전 작성).
- 변이 원장 · 다중 적대 리뷰: not-applicable — 비례 원칙(시험 전용, High-risk 경로·게이트 판정 불변). 대신 사본 대조 실험(RED 재현 ·
  수리 뒤 같은 지연에서 GREEN · 닫기 호출을 지운 양성 대조 FAIL)이 그 자리를 맡는다.
- gstack 리뷰: 생략 — 사용자 지시 「간단한건 리뷰 할 필요 없음」(시험 전용 · 생산 diff 0). `review.md` 에 사유를 남긴다.
