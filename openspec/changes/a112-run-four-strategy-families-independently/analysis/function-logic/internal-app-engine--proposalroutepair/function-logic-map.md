# Function Logic Map: `proposalRoutePair`

- Source: `internal/app/engine/strategy_proposal_authority_test.go` (84-110)
- Function: `proposalRoutePair` in package `engine`
- Signature: `proposalRoutePair(params=2, results=1)`
- File SHA-256: `4acb8506cc32d2cc5fd4eda1a5366152ba7dcf92e704d078ceced5fe268513ea`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 태스크 5.4.

## Inputs and invariants

KR/US 각 한 종목의 봉인된 경로 항목 쌍을 만드는 테스트 픽스처다. production 코드가 아니다.

## 이 lot 이 무엇을 바꿨는가 — 픽스처에 **승인된 채점 기준**을 붙였다

두 곳이 늘었다(`98:35`, `105:11`).

```go
evidenceDigest, configDigest := arbitrationLineageDigests(t, value, now)
route, err := strategyrouter.ProductionRouteAuthorityForTest(key, ..., evidenceDigest, configDigest, now)
route = strategyrouter.WithArbitrationScoresForTest(route, arbitrationCalibrationForTest, familyScoresForTest(routerMarket))
```

`ProductionRouteAuthorityForTest` 는 보정도 가족 점수도 붙이지 않은 권한을 만든다.
태스크 5.4 이후 그런 권한은 `SealsValid()` 가 거짓이므로, 제안이 하나뿐인 종목도
`ARBITRATION_UNCALIBRATED` 로 거절된다. 이는 승인된 spec 이 요구하는 동작이다 —
"Singleton proposal도 approved score/calibration authority가 없으면 production dispatch에
전달해서는 안 된다 (MUST NOT)".

즉 이 픽스처는 **spec 이 금지하는 상태**를 production-ready 로 흉내 내고 있었다.
픽스처를 고쳐 매니페스트가 싣는 것과 같은 네 가족 점수표를 붙였다. 테스트가 무엇을
재는지는 그대로이고, 재는 대상이 유효한 매니페스트 상태가 됐다.

이 변경을 하지 않으면 `TestStrategyProposalAuthorityLoadsKRUSConcurrently` 와
`TestStrategyProposalAuthorityKeepsMarketFailureLocal` 이 실패한다 — 실제로 실패했고,
그 실패가 곧 spec 요구가 배선됐다는 증거다.

**두 번째 변경 — 증거·설정 다이제스트를 손으로 적지 않는다.** 독립 리뷰(P1-1)가 지적한
대로, 픽스처의 경로 권한은 `"lane-evidence"`/`"lane-config"` 를 쓰고 제안 계보는
`sha256:6…`/`sha256:8…` 를 쓰고 있었다. **생산에서는 그 둘이 같은 값이다** —
`Propose` 가 경로 결정의 두 값을 계보에 그대로 박기 때문이다(`flow.go:67-68`).
픽스처가 생산에서 있을 수 없는 상태를 만들고 있었고, 그래서 중재자에 그 결속 검사를
넣자 테스트가 깨졌다. 값을 손으로 맞추는 대신 `arbitrationLineageDigests` 로
**실제 계보에서 읽는다** — 두 곳에 적으면 언젠가 다시 어긋난다.

## Branches and early returns

- Go 커버리지는 `_test.go` 를 계측하지 않는다. 이 함수의 분기는 커버리지 숫자로 잴 수 없다.
- 행은 "그 분기를 실행한 것으로 관측된 run" 을 적으며, 숫자가 아니다.

Exact AST return positions: 107:3, 109:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 89:3 | US 분기. `TestStrategyProposalAuthorityLoadsKRUSConcurrently` 와 `TestStrategyProposalAuthorityKeepsMarketFailureLocal` 이 두 시장을 모두 만들므로 매 실행 진입한다. |
| B2 | if | 93:3 | `NewOwnerKey` 실패 시 `t.Fatal`. 관측된 진입 없음 — 픽스처 입력이 항상 유효하다. |
| B3 | if | 100:3 | `ProductionRouteAuthorityForTest` 실패 시 `t.Fatal`. 관측된 진입 없음 — 같은 이유. |

관측 근거: `go test -count=1 -tags tossos_testseams -run '^TestStrategyProposalAuthority' ./internal/app/engine/` 가
통과하며, 두 테스트 모두 KR/US 양쪽 스냅샷을 `Ready` 로 읽는다(그 단언이 곧 B1 진입의 증거다).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 85:2 |
| `strategy.ApprovedSnapshotForTest` | 87:15 |
| `string` | 87:48 |
| `strategyrouter.NewOwnerKey` | 92:15 |
| `t.Fatal` | 94:4 |
| `arbitrationLineageDigests` | 98:35 |
| `strategyrouter.ProductionRouteAuthorityForTest` | 99:17 |
| `t.Fatal` | 101:4 |
| `strategyrouter.WithArbitrationScoresForTest` | 105:11 |
| `familyScoresForTest` | 105:93 |
| `string` | 107:238 |
| `market` | 109:57 |
| `market` | 109:97 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.
- 디스크·저널·브로커에 쓰지 않는다. 메모리 안의 값만 만든다.

## Safety conclusion

테스트 전용 픽스처다. production 동작을 만들지 않는다. 이 변경은 픽스처를 승인된
매니페스트 계약에 맞춘 것이며, 픽스처를 느슨하게 해서 테스트를 통과시킨 것이 아니다 —
반대로, 픽스처가 없던 요구(승인된 채점 권한)를 이제 갖춰야만 통과한다.
