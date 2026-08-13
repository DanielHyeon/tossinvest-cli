# Branch Test Map: `engineRuntime`

Source: `cmd/tossctl/engine.go` (384-471). AST 기준 branches **6** / returns 7.

## 커버리지는 주장이 아니라 측정값이다

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile`을 **편집 전과 후에 각각** 돌려
블록 카운트를 잘라 읽었다.

| | 통과 | 위치 | 블록 | source SHA-256 |
|---|---|---|---|---|
| 1판 (편집 전) | 526건 | `:347-430` | 14개 중 **10개** 실행 | `f13e36b35e08…` |
| 2판 (`6cd643ca`) | 550건 | `:356-443` | 13개 중 10개 실행 | `8ad1cc88b9e0…` |
| 3판 (§3.9b) | 564건 | `:372-459` | 13개 중 10개 실행 | `bc5748c552dd…` |
| **4판 (§3.9c)** | **572건** | **`:384-471`** | **13개 중 10개** 실행 | `ee527a6a917a…` |

블록이 하나 줄었다 — `Recover` 클로저 본문(1판 `401.44,404.4` count=**0**)이 사라졌다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 전) | 지는 테스트 (편집 전) |
|---|---|---|---|---|
| B1 | `:387` 체결 감지기 오류 | **no** | `387.16,389.3` count=**0** | 없음 |
| B2 | `:396` 대사 드라이버 오류 | yes | `396.16,398.3` count=**1** | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B3 | `:407` exit 관측자 오류 | yes | `407.16,409.3` count=**1** | 같음 |
| B4 | `:412` 복구 조립 오류 | yes | `412.16,414.3` count=**1** | 같음 |
| B5 | `:416` 전략 진입 supervisor 오류 | **no** | `416.16,418.3` count=**0** | 없음 |
| B6 | `:428` 알림 배출기 오류 | **no** | `428.16,430.3` count=**0** | 없음 |

**`Recover` 클로저 본문은 편집 전 별도 블록이었고 count=0이었다** — `401.44,404.4` count=**0**.
즉 `recovery.Run`을 부르고 Report를 버리는 그 두 줄을 실행하는 테스트가 하나도 없었다.
`TestTheRestartRecoveryRunsBeforeTheLoops`는 실행이 아니라 **소스 문자열**을 본다
(`engine_test.go` — 편집 후에는 `Recover: recoverThenReady(recovery.Run, ready,`를 찾는다).

**편집 후 그 블록은 없다.** 판정이 `recoverThenReady`(100.0%)와
`engineRecoveryObserver`(100.0%)로 나갔고, 둘의 분기는 아래 테스트들이 진다.

> ⚠ **이 측정이 D5의 형태를 정했다.** 판정을 이 클로저 안에 쓰면 그 판정은 어떤 테스트도
> 닿지 못하는 자리에 놓인다(a101이 `runConsole`에서 같은 결론에 도달했다). 그래서 D5는
> 인자를 받는 함수로 나갔고, 클로저는 **없어졌다**.
>
> B1·B5·B6의 공백은 **물려받은 것**이고 편집 후에도 그대로다 — a102는 그 세 줄을 편집하지
> 않는다. **침묵하지 않고 이름을 붙여 남긴다.**

## a102가 이 함수 밖으로 뺀 판정의 커버리지 (측정값)

| 함수 | statements | 지는 테스트 |
|---|---|---|
| `recoverThenReady` | **100.0%** | `TestRecoveryPublishesReadyOnlyWhenItFinished` · `TestAFailedRecoveryNeverPublishesReady` · `TestACancelledRecoveryNeverPublishesReady` · `TestTheRateLimitedWaitIsReportedOnBothPaths` |
| `engineRecoveryObserver` | **100.0%** | `TestARateLimitedRecoveryLeavesOneCountableLine` · `TestARecoveryThatWasNeverThrottledSaysNothing` · `TestTheRecoveryObserverSurvivesANilLogger` |

## 뮤테이션 정산

| 뮤테이션 | 가한 것 | 죽은 테스트 | 원복 |
|---|---|---|---|
| **(e)** | 실패한 복구에도 `ready()`를 부른다 | `TestAFailedRecoveryNeverPublishesReady` | sha `fc5488ccafe2` 동일 |
| (j) 자체 | `observe(report)`를 성공 경로 뒤로 옮긴다 (실패 경로 무음) | `TestTheRateLimitedWaitIsReportedOnBothPaths` | sha `fc5488ccafe2` 동일 |
| **N5** (A2) | `engineRuntimeFactory(…, nil)` — 배선은 남고 seam만 사라진다 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` | sha `ee527a6a917a` 동일 |
| **N5b** (A2 잔여, §3.9b) | **이 함수 본문에서** `ready = nil` 재대입 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` — 조립된 `Recover`가 실행되고 ready가 run ctx를 취소하지 못해 런타임이 진짜 루프를 띄운다. **실패는 그 루프의 패닉으로 나타난다**(typed-nil 브로커) | sha `ee527a6a917a` 동일 |

> ⚠ **N5b의 실패 양식은 패닉이다.** ready가 run 컨텍스트를 취소하는 유일한 주체이므로,
> 그것이 사라지면 런타임이 살아 있는 컨텍스트로 루프를 시작하고 typed-nil official client에
> 닿는다. 스택은 `internal/app/engine.(*ReconcileDriver).stabilise`를 지목하므로 **무엇이
> 깨졌는지는 말하지만**, 이름 있는 `--- FAIL:` 한 줄보다 읽기 어렵고 같은 실행의 다른
> 테스트 결과를 잃는다. 이름 있는 실패로 바꾸려면 취소를 ready 밖으로 빼야 하는데, 그러면
> `recoverThenReady`가 `ctx.Err()`에서 먼저 돌아가 ready 자체가 불리지 않는다 —
> 즉 **이 단언과 양립하지 않는다. 침묵하지 않고 이름을 붙여 남긴다.**

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 6, returns 8) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0 ·
  526건(1판) → 550건(2판) → 564건(3판) → **572건**(4판)
- Report 소비자 전수: `rg -n 'RateLimitWaits|RateLimitWaited' --glob '!*_test.go'` →
  편집 전 `internal/reconcile/{recovery,ratelimit}.go` 뿐(**cmd 쪽 0건**, A1 F1) →
  편집 후 **`cmd/tossctl/engineready.go` 추가**.
