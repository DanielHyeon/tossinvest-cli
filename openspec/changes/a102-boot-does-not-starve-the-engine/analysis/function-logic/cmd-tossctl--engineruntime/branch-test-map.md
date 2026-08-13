# Branch Test Map: `engineRuntime`

Source: `cmd/tossctl/engine.go` (356-443). AST 기준 branches **6** / returns 7.

## 커버리지는 주장이 아니라 측정값이다

`go test ./cmd/tossctl -count=1 -covermode=set -coverprofile`을 **편집 전과 후에 각각** 돌려
블록 카운트를 잘라 읽었다.

| | 통과 | 위치 | 블록 | source SHA-256 |
|---|---|---|---|---|
| 편집 전 | 526건 | `:347-430` | 14개 중 **10개** 실행 | `f13e36b35e08…` |
| **편집 후** | **550건** | **`:356-443`** | **13개 중 10개** 실행 | `8ad1cc88b9e0…` |

블록이 하나 줄었다 — `Recover` 클로저 본문(1판 `401.44,404.4` count=**0**)이 사라졌다.

| Branch | 위치 | 본문 실행 | 근거 블록 (편집 전) | 지는 테스트 (편집 전) |
|---|---|---|---|---|
| B1 | `:359` 체결 감지기 오류 | **no** | `359.16,361.3` count=**0** | 없음 |
| B2 | `:368` 대사 드라이버 오류 | yes | `368.16,370.3` count=**1** | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B3 | `:379` exit 관측자 오류 | yes | `379.16,381.3` count=**1** | 같음 |
| B4 | `:384` 복구 조립 오류 | yes | `384.16,386.3` count=**1** | 같음 |
| B5 | `:388` 전략 진입 supervisor 오류 | **no** | `388.16,390.3` count=**0** | 없음 |
| B6 | `:400` 알림 배출기 오류 | **no** | `400.16,402.3` count=**0** | 없음 |

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
| **(e)** | 실패한 복구에도 `ready()`를 부른다 | `TestAFailedRecoveryNeverPublishesReady` | sha `836582800678` 동일 |
| (j) 자체 | `observe(report)`를 성공 경로 뒤로 옮긴다 (실패 경로 무음) | `TestTheRateLimitedWaitIsReportedOnBothPaths` | sha `836582800678` 동일 |

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 6, returns 8) — `go run ./tools/logic-map`
- 커버리지: `go test ./cmd/tossctl -count=1 -covermode=set -coverprofile` exit 0 ·
  **526건 통과**(편집 전) → **550건 통과**(편집 후)
- Report 소비자 전수: `rg -n 'RateLimitWaits|RateLimitWaited' --glob '!*_test.go'` →
  편집 전 `internal/reconcile/{recovery,ratelimit}.go` 뿐(**cmd 쪽 0건**, A1 F1) →
  편집 후 **`cmd/tossctl/engineready.go` 추가**.
