# Function Logic Map: `Context.Read`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Current source SHA-256: `5f203ad88f4476284006b92099365d42d04a619a2f567524efdd9bb1beb64f65`
- Signature: `Context.Read(params=1, results=2)`
- Source range: `23:1`–`65:2`
- AST evidence: `ast.json`, regenerated from the post-edit worktree; AST 분기 7개 (편집 전 6개).
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 결정 54의 `ConfigDigest`/`BuildDigest` 노출.

## 이 lot 이 무엇을 바꿨는가

1. `if err != nil || supervisor == nil` 한 줄이 **두 분기로 갈라졌다** (B3, B4). 두 조건이 하던 일이
   달라졌기 때문이다 — 오류는 "아무것도 안 붙이고 끝", supervisor 부재는 "붙이고 끝"이다.
2. 그 사이에 직선 한 줄이 들어갔다: 이 프로세스의 config/build digest 를 envelope 에 붙인다.
   **`sha256:` 접두사를 벗기지 않는다** — 이 문자열은 사람이 매니페스트에 그대로 옮겨 적고,
   `scheduler.validateProductionActivationManifest` 는 `body.ConfigVersion != binding.ConfigVersion`
   으로 정확히 비교한다. 벗기면 옮겨 적은 매니페스트가 거절된다(독립 리뷰 P1-1).

## 왜 store 가 아니라 여기인가

`store` 의 내용은 전략 assembly refresh 가 채우는데, 그 refresh 는 **아직 한 번도 안 돌았거나
실패했을 수 있다.** 운영자가 이 숫자를 가장 필요로 하는 순간이 바로 그때다 — 활성화 매니페스트가
없어서 스케줄러가 안 깨어난 상태.

`Read` 는 REST·SSE·콘솔·Unix transport 가 모두 지나는 **단 하나의 출구**다. 여기 붙이면 어느
표면에서 보든 같은 값이고, 표면마다 사본을 두지 않는다.

## 순서가 중요하다 — 붙이고 나서 덮는다

identity 를 붙인 **뒤에** latch 덮어쓰기(B5–B7)가 돈다. `WithMarketFailure` 는 `Clone` 을 거쳐
envelope 를 새로 만들므로, `Clone` 이 이 필드를 안 옮기면 **시장이 latch 된 순간 숫자가 사라진다.**
그 조합이 `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` 이 재는 것이다.

## Inputs and invariants

- 입력은 AST signature 그대로다.
- 불변식: 읽지 못한 것에는 아무것도 붙이지 않는다 (B3).
- 불변식: 이 함수는 활성화·주문·토글 능력이 없다. 관측만 돌려준다.

## Branches and early returns

- Exact AST return nodes: `25:3`, `31:3`, `35:3`, `51:3`, `64:2`.

| Branch | AST kind | Source location | Edited by this lot | Disposition |
|---|---|---|---|---|
| B1 | if | 24:2 | 아니오 | nil receiver/context. 기존 그대로. |
| B2 | if | 30:2 | 아니오 | store 부재. 기존 그대로. |
| B3 | if | 34:2 | **예 — 조건이 좁아짐** | 편집 전 `err != nil || supervisor == nil`. 이제 `err != nil` 만. 오류는 identity 를 붙이기 **전에** 끝난다. |
| B4 | if | 50:2 | **예 — 이 lot 이 추가** | B3 에서 갈라져 나온 `supervisor == nil`. identity 를 붙인 뒤 latch 덮어쓰기 없이 반환한다. |
| B5 | range | 53:2 | 아니오 | KR·US 루프. 기존 그대로. |
| B6 | if | 55:3 | 아니오 | latch 안 된 시장 건너뛰기. 기존 그대로. |
| B7 | if | 59:3 | 아니오 | CURRENT 인 시장만 덮기. 기존 그대로. |

## Calls and live bindings

| Callee expression | Source location |
|---|---|
| errors.New | 25:41 |
| c.strategyProjectionMu.RLock | 27:2 |
| c.strategyProjectionMu.RUnlock | 29:2 |
| errors.New | 31:41 |
| store.Read | 33:19 |
| strategyprojection.WithRuntimeIdentity | 48:13 |
| strategyRuntimeConfigDigest | 49:3 |
| strategyRuntimeBuildDigest | 49:34 |
| supervisor.Snapshot | 54:17 |
| strategyprojection.Market | 58:23 |
| strategyprojection.WithMarketFailure | 60:15 |

## State mutations and fallbacks

- 지역 스냅샷 값만 바꾼다. store 에는 쓰지 않는다. fallback 없음 — digest 를 못 만들면
  `WithRuntimeIdentity` 가 아무것도 붙이지 않고, 화면은 `not_observed` 가 된다.

## Safety conclusion

- 읽기 전용 관측 경로다. 주문·손절·사이징·Guardian·원장·인증 경로에 닿지 않는다.
- 이 lot 전까지 이 함수에는 **어떤 테스트도 없었다.** 이 lot 이 B1·B2·B3·B4 와 B5–B7 의
  latch 경로를 처음으로 실행한다.
- 반증 실측: config/build 인자를 맞바꾸면(뮤테이션 M2) 두 테스트가 실패한다.
