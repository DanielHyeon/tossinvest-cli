# a083 · Issues

## I1 — missing-order 차단은 credit 발행자가 없어 여전히 운영자 전용이다

`blocksFor`는 `diff.MissingOrders`에도 `Cause=QUANTITY_MISMATCH` ·
`Release=ReleaseOnAdjustedReconcile` 차단을 만든다(`mismatch.go:848`). 상태표는 그
줄을 "auto release: adjusted reconcile"로 적어 두었다.

그런데 credit을 발행하는 유일한 프로덕션 경로인 `ConvergeQuantities`는
`diff.Quantities`만 처리한다. 계좌가 보여주지 않는 주문에는 수렴시킬 수량이 없으므로
조정도 없고, 따라서 credit도 없다. **상태표와 코드가 어긋나 있다** — 이 줄의 실제 해제
경로는 운영자뿐이다.

운영 원장의 활성 8건은 전부 수량 불일치이므로 지금 live 문제는 아니다. 해결 방향은
둘 중 하나다.

1. missing-order 차단에 별도 release cause를 정의하고 상태표를 사실에 맞춘다
2. 주문 계보 해소가 확정한 부재를 credit으로 인정한다

어느 쪽이든 별도 change다.

## I2 — 해제 분기가 계좌 단위여서 한 심볼이 다른 심볼의 해제를 미룬다

`Observe`의 해제는 `!diff.BlocksEntry()`, 즉 비교 **전체**가 일치할 때만 돈다
(`mismatch.go`). 차단은 심볼 범위인데 해제 조건은 계좌 범위다. 어떤 심볼이 계속
불일치하면 이미 수렴하고 일치한 다른 심볼의 해제도 그동안 미뤄진다.

a083은 이것을 넓히지 않는다. 다만 credit이 심볼 단위로 보존되므로(design D2b) **한
번이라도 비교 전체가 일치하면 그 순간 전부 해제되고**, 영구히 막히는 성질은 사라진다.
`TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit`이 그 경로를 고정한다.

심볼 단위 해제는 요구사항 수준의 변경이며 별도 change다.

## I3 — 같은 심볼의 두 차단이 map key를 공유한다

`Block.Key()`는 ScopeSymbol에서 `symbol|account|SYMBOL`이다(`mismatch.go:229`).
한 심볼에 수량 불일치와 missing-order 차단이 동시에 생기면 in-memory map에서 하나만
남는다. 원장은 `(account, symbol)` 단위로 하나의 활성 행만 허용하므로 durable 쪽과는
일관되지만, 두 사유가 하나로 뭉쳐 보이는 것은 사실 손실이다.

기존 성질이고 a083이 악화시키지 않는다. I1과 함께 다뤄야 한다.

## I4 — 차단 evidence 문자열이 첫 관측에 고정되어 화면이 낡은 숫자를 보여준다

`EnterReconcile`은 멱등이고 이미 활성인 scope에 대해 **첫 관측의 evidence를 보존**한다.
그래서 화면의 "the engine believes 2 of 103590, the account says 3"은 2026-08-04
23:29의 값이고, 같은 화면이 바로 아래에 보여주는 원장 수량 4와 다르다.

운영자가 두 숫자를 나란히 보고 어느 쪽이 현재인지 판단할 수 없다. 값을 갱신하려면
멱등성을 깨거나 별도의 관측 기록을 두어야 하므로 a083 범위 밖이다. a085가 화면
표현을 다룰 때 함께 판단한다.

## I5 — 독립 리뷰어 없이 배포 가능 상태에 도달했다

`review.md` 참조. `codex exec`가 사용량 한도로 거부되어 적대적 Eng 리뷰를 작성자와
같은 컨텍스트에서 수행했다. 자기 설계의 구멍 하나(D2b)를 찾아 고쳤지만, 그것이
"작성자와 검증자의 분리"를 대체하지는 않는다. **배포 전 별도 세션의 재검증이 남아
있다.**
