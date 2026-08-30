# Change: verify-reopens-conditional-chain

## Why

KR 조건주문 체인(task 2.5)은 **영구 교착**이다. 도구가 자기 측정 대상을 스스로 파괴했고,
사용자가 검증을 운전하는 콘솔에서는 그것을 되살릴 방법이 없다.

증거 — `~/.local/share/tossos/capability-verify.jsonl` 실측 타임라인:

```
2026-07-26T13:48:58Z  conditional-cancel      skipped   "conditional-register did not pass"
2026-07-28T01:34:36Z  conditional-register    pass      조건주문 grLKqi… 등록 (deliberate=true)
2026-07-28T01:34:37Z  conditional-persist     awaiting-restart
2026-07-28T01:35:31.378Z  cleanup             pass      grLKqi… 취소됨  ← 여기
2026-07-28T01:35:31.686Z  conditional-persist skipped   "이 검증이 만든 살아 있는 조건주문이 없다"
```

정리 prologue가 존속 측정 대상을 **측정 308밀리초 전에** 취소했다.

### 결함 1 — 객체보다 오래된 판정이 그 객체의 운명을 결정한다

현행 규칙은 "조건주문은 그것을 취소하는 단계가 **terminal 판정**을 가진 뒤에만 정리 대상"이다.
`skipped`도 terminal이다. 2026-07-26에는 계좌가 비어 있어 `conditional-cancel`이 skipped로
굳었고, 그 판정은 **07-28에 등록된 조건주문에 대해서는 아무것도 말하지 않는다**. 그런데
`cleanupFrom`은 그것을 근거로 삼았다(`internal/verifylive/cleanup.go:116`).

존재하지 않던 객체에 대해 기록된 판정은 그 객체에 대한 판정이 아니다.

### 결함 2 — 대상이 사라진 체인을 콘솔이 되살릴 수 없다

`conditional-register`는 `pass`다. `RedoableVerdict`는 `pass`를 제외하고(`redo.go:77`),
콘솔은 `RedoSet`만 제시한다(`internal/console/data.go:310`). 그래서:

- `[이어하기]` — persist·modify·cancel은 살아 있는 조건주문이 없어 다시 skipped
- `[재측정]` — register가 집합에 없으므로 대상을 다시 만들지 않는다
- 결과: 이후 세 실행(07-28 12:01, 12:02, 07-29 11:32)이 전부 같은 자리에서 skipped

CLI `tossctl verify run --market KR --redo conditional-register`는 존재하지만, 사용자는
콘솔에서 검증을 운전한다. 콘솔만 쓰는 한 이 체인은 다시는 측정되지 않는다.

### 왜 지금인가

`openspec/changes/verify-execution-capability/tasks.md:3`이 **2c `add-protection-orders`는
task 2.5~2.9 완료 전에 작성하지 않는다**고 못박는다. 2.5는 이 교착 때문에 미완이고,
2c가 없으면 인터록 6/9절이 `ProtectionReady=UNWIRED`를 유지해 `tossctl engine run`은
어느 기계에서도 기동하지 못한다. 이 change는 2c의 선행이지 2c가 아니다.

## What Changes

- **판정은 자기보다 나중에 생긴 객체를 정리할 권한이 없다.** 조건주문은 그것을 취소하는
  단계의 판정이 **그 조건주문이 기록에 등장한 뒤에** 기록됐을 때만 정리 대상이 된다.
  기록은 append-only이므로 기준은 시계가 아니라 **기록 순서**다.
- **대상이 사라진 체인은 콘솔에서 다시 열린다.** 통과한 단계라도 ① 그 단계가 의도적으로
  남긴 조건주문이 더 이상 없고 ② 그것에 의존하는 비-deferred 단계 중 통과하지 못한 것이
  있으면 재측정 집합에 들어간다. 이 세 조건이 모두 참일 때만이다.
- **재측정 표가 재측정 집합과 같은 답을 한다.** 콘솔 표의 "재측정이 건드릴 행" 표시는
  판정만 보고 판단하지 않고 집합과 동일한 근거를 쓴다.

## Non-Goals

- 승인 없이 무언가를 보내는 경로 — 없다. 재측정 집합은 **제안이지 인가가 아니다**.
  모든 요청은 새 nonce로 새 계획·새 배치 승인을 다시 지난다(`runner.go` `approveBatch`).
- `pass` 일반을 재측정 대상으로 여는 것 — 아니다. 유일한 예외는 **통과가 확립한 성질
  자체가 더 이상 참이 아닌** 경우다. `order-cancel`·`sell-boundary` 같은 단계는 대상이
  사라지는 것이 정상 종료이므로 이 규칙에 걸리지 않는다.
- `deferred`·`refused` 재개 — 그대로 제외한다.
- 노출 상한·1주 규칙·승인 창(5분) 변경 — 무변경.
- 조건주문 발동(`conditional-trigger`) 측정 — 그대로 deferred다.
- 2c `add-protection-orders` 스펙 작성 — 금지 유지.

## Capabilities

### Modified Capabilities

- `order-execution`: 정리 대상 판단이 **판정과 객체의 선후**를 따진다; 대상이 사라진
  전제 단계를 재측정 대상으로 인정한다
- `operator-console`: 콘솔이 교착된 조건주문 체인을 다시 열 수 있다

## Impact

- Affected code: `internal/verifylive/cleanup.go`(`cleanupFrom` 조건주문 가드),
  `internal/verifylive/redo.go`(`RedoSet` + 새 술어), `internal/console/pages.go`
  (`redoable` 템플릿 함수), `internal/console/templates.go:451`(재측정 표)
- High-risk 여부: **yes** — 라이브 주문 취소 경로(정리)와 조건주문 등록 경로(재개통)에
  모두 닿는다. 적대적 Eng 리뷰 + Pre-Edit 선언 + Function Logic Map 필수.
- 안전 검토(§0):
  - 결함 1 수정은 **취소를 덜 보내는** 방향이다. 가드가 닫히는 경우만 늘고 열리는 경우는
    늘지 않는다. §0.3(손절 즉시성)과 무관하고 §0.9(보수 방향)에 부합한다.
  - 결함 2 수정은 조건주문 **등록**을 다시 제안할 수 있게 하므로 라이브 side effect가
    늘어날 수 있다. 완화: (a) 등록되는 것은 1주짜리 SINGLE MARKET SELL 손절이고 발동가가
    시장보다 한참 아래라 발동할 수 없다 — 보호를 더하는 방향이다; (b) 집합은 인가가
    아니며 사람이 새 배치 승인을 다시 준다(§0.1·§0.7 충족); (c) 조건이 세 개 모두
    참일 때만이고, 정상 종료한 US 체인은 이 규칙에 걸리지 않는다(실측 확인).
  - §0.4 rate limit: 재측정 1회가 추가하는 호출은 기존 `conditional-register` 단계의
    호출과 동일하다. 새로운 호출 종류는 없다.
