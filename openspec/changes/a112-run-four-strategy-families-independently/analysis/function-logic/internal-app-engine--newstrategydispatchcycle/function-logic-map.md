# Function Logic Map: `newStrategyDispatchCycle`

- Source: `internal/app/engine/strategy_dispatch_cycle.go` (48-57)
- Function: `newStrategyDispatchCycle` in package `engine`
- File SHA-256: `c872acdb342ec44f87ab70114e36c7dafd042c90ac0b0c5dcd3668288101a625`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

Exact AST return nodes: `55:2`.

## Inputs and invariants

이 생성자는 공유 dispatch 주기의 필드를 채운다. 태스크 8.8.2 가 인자를 하나
늘렸다: `proposals strategyProposalAuthorityPair`.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `jrn` | nil 가능 | 호출자(`NewPairedStrategyEntryProductionAssembly`) | nil 이면 `dispatch` 첫 관문이 거절한다. 여기서 막지 않는다 |
| `gateway` | nil 가능 | 같음 | 같음 |
| `firstLeg` | nil 가능 | 같음 | 같음 |
| `schedule`·`fx`·`risk` | 값 타입, 영값 가능 | 같은 파도의 권한 | 준비 안 됐으면 `dispatch` 가 거절한다 |
| `proposals` | 값 타입, 영값 가능 | 같은 파도의 제안 권한 | **영값이면 활성화가 영값이고 하한을 아무도 요구하지 않는다** — 그것이 활성화 없는 시장의 값이다 |
| `owner` | nil 이면 새로 만든다 | 호출자 | 유일한 기본값 처리 |

`proposals` 를 값으로 받는 이유: 이 주기는 그 파도의 활성화를 읽어야 하고,
파도마다 새 주기가 만들어지므로 묵은 활성화가 남을 수 없다. 포인터로 들면
다른 파도의 활성화를 보는 창이 생긴다.

## Branches and early returns

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 52:2 | `owner` 가 nil 인 경우의 유일한 기본값 처리. 생산은 언제나 owner 를 준다 |

## Calls and live bindings

| Callee expression | Position |
|---|---|


## State mutations and fallbacks

- 아무것도 바꾸지 않는다. 새 구조체 하나를 만들어 돌려줄 뿐이다.
- 유일한 fallback 은 `owner == nil` 일 때 새 coordinator 를 만드는 것이다.

## Safety conclusion

- Safe edit boundary: 인자 추가는 호출자 한 곳(`NewPairedStrategyEntryProductionAssembly`)과
  시험 픽스처 한 곳(`pairedStrategyDispatchCycleFixture`)에만 닿는다. 컴파일러가
  누락을 잡으므로 조용히 빠질 수 없다.
- High-risk impact: yes — 이 구조체가 주문 경로의 승인 사슬을 들고 있다.
