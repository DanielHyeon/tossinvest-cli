# Branch Test Map: `Runner.mutationSymbol`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:Runner.mutationSymbol`

이 change의 RED는 실제로 관측했고 원문을 아래에 남긴다. 다른 target의 branch-test-map은
이 파일을 참조한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 572: `if step.ActsOnConditional {` — 조건주문 위에서 동작하는 단계 | `TestThePlanNamesTheConditionalItWillModify`, `TestEveryMutatingStepThatActsOnTheLiveConditionalDeclaresIt` | yes (RED 2) | yes |
| B2 | `if` line 573: `if _, symbol, ok := r.liveConditional(); ok && …` — 살아 있는 조건주문이 있다 | `TestThePlanNamesTheConditionalItWillModify`, `TestTheConditionalStepsReachTheBrokerWhenTheProbeSymbolDiffers` | yes (RED 2) | yes |
| B2-false | 같은 줄의 거짓 가지 — 아직 등록되지 않았다 → `holdingSymbol` | `TestAFreshRunPlansTheConditionalStepsAgainstTheHolding` | yes (RED 2) | yes |
| B3 | `if` line 578: `if step.NeedsHolding {` — 무변경 경로 | `TestThePlanCarriesEveryMutationTheCatalogueDeclares`, `TestAUSRunPlansOrdersForTheUSSymbol` | no (동작 무변경) | yes |
| B3-false | 기본 경로 `return r.symbol` — 무변경 | `TestThePlanDigestIsPinnedAcrossBuilds` (digest 불변이 이 경로가 안 움직였다는 증거) | no (동작 무변경) | yes |

## RED 1 — 카탈로그에 선언이 없다 (컴파일 실패)

```
verifylive [build failed]
  internal/verifylive/plan_symbol_test.go:303:48: step.ActsOnConditional undefined (type Step has no field or method Ac...
  internal/verifylive/plan_symbol_test.go:307:13: step.ActsOnConditional undefined (type Step has no field or method Ac...
```

## RED 2 — 필드만 추가하고 해석은 그대로 (5 fail / 2 pass)

```
verifylive (2 passed, 5 failed)
  [FAIL] TestThePlanNamesTheConditionalItWillModify
     plan_symbol_test.go:65: conditional-modify is planned against "005930", but the step acts on the ...
     plan_symbol_test.go:65: conditional-cancel is planned against "005930", but the step acts on the ...
  [FAIL] TestAFreshRunPlansTheConditionalStepsAgainstTheHolding
     plan_symbol_test.go:86: conditional-modify is planned against "005930"; conditional-register will...
     plan_symbol_test.go:86: conditional-cancel is planned against "005930"; conditional-register will...
  [FAIL] TestTheConditionalStepsReachTheBrokerWhenTheProbeSymbolDiffers
     plan_symbol_test.go:107: second invocation: verify: this request is not on the approved list, so ...
     plan_symbol_test.go:113: conditional-modify = "fail", want pass: verify: this request is not on t...
     plan_symbol_test.go:113: conditional-cancel = "", want pass:
     plan_symbol_test.go:117: the verification left [{Kind:conditional-order ID:co-6 Symbol:333430 Cre...
  [FAIL] TestAPlanLineWithoutASymbolAuthorisesNothingWithOne
     plan_symbol_test.go:158: a line that names no symbol authorised a request for 333430 — a list tha...
  [FAIL] TestEveryMutatingStepThatActsOnTheLiveConditionalDeclaresIt
     plan_symbol_test.go:304: conditional-modify sends a live request against the registered condition...
     plan_symbol_test.go:304: conditional-cancel sends a live request against the registered condition...
```

`TestTheConditionalStepsReachTheBrokerWhenTheProbeSymbolDiffers`의 RED는
2026-07-29 실계좌에서 관측된 것과 **같은 실패**다: `conditional-modify = "fail"`,
`ErrOutsidePlan`, `conditional-cancel`은 실행되지도 못했고, 조건주문이 계좌에 남았다.

## RED 3 — 종목 없는 계획 줄 (별도 관측)

harness가 빈 종목을 채워 넣어 처음에는 통과했다. `New`를 직접 불러 계좌가 종목을 줄 수
없는 상태(US 계좌·보유 0)를 만들자:

```
  [FAIL] TestAMutatingStepWithNoNameableTargetIsNotPlanned
     plan_symbol_test.go:154: the approval list carries a line with no symbol: {Ordinal:1 Step:idempot...
     plan_symbol_test.go:154: the approval list carries a line with no symbol: {Ordinal:2 Step:idempot...
     plan_symbol_test.go:154: the approval list carries a line with no symbol: {Ordinal:3 Step:idempot...
     plan_symbol_test.go:154: the approval list carries a line with no symbol: {Ordinal:4 Step:idempot...
```

승인 목록에 **종목을 말하지 않는 라이브 주문 줄이 4개** 있었다.

## GREEN

```
Go test: 210 passed in 1 packages     (internal/verifylive)
Go test: 3749 passed in 57 packages   (전체, 07d4ba0의 3742 + 신규 7)
Go vet: No issues found
```
