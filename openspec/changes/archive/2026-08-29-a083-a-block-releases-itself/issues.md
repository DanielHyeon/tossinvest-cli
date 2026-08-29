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

## B1 — **차단 해제(blocking)**: credit이 자신이 답한 차단에 묶여 있지 않다

2026-08-05 독립 리뷰(별도 컨텍스트 2명)가 각각 실행 가능한 재현으로 확인했다.

`Observe` 끝의 credit 정산은 **분쟁 중이거나 해제가 커밋된 심볼의 credit만** 버린다.

```go
for symbol, comparison := range t.adjusted {
    if !creditUsableBy(comparison, diff.AsOf) { continue }
    if disputed[symbol] || committed[symbol] { delete(t.adjusted, symbol) }
}
```

그 심볼이 **일치하는데 해제가 돌지 않은 경우**(다른 심볼이 `BlocksEntry()`를 참으로
유지) credit은 남는다. 그리고 해제 분기는 `t.adjusted[symbol]`만 보고 **그 차단이
credit이 답한 차단인지 확인하지 않는다.**

재현 A — 재분류로 자기 해제:
```
주기 N     브로커 holdings에서 005930이 일시 누락 → QuantityMismatch(local 10, broker 0)
           → 차단, 수렴이 계좌의 0을 투영에 기록, credit(T_N)
주기 N+1   보유 복귀 → local 0, broker 10 → compare.go:406이 ExternalPos로 분류
           → BlocksEntry()는 ExternalPos를 세지 않는다 → diff "clean"
           → 차단이 ADJUSTMENT_APPLIED로 풀리고 신규 진입이 재개된다
```
관측: `PROBE HIT: the block released while the account holds 10 and the engine holds 0`.
이때 엔진 투영은 0이고 계좌는 10이다 — **틀린 포지션 인식 위에서 매매한다.**

재현 B — 나중에 생긴 다른 차단을 푼다:
6시간 뒤 `position_projection.go`의 oversell이 같은 심볼에 새 `QUANTITY_MISMATCH`
차단을 만들면, 다음 clean 비교가 **6시간 전 credit으로** 그 차단을 푼다.
base `53626032`에서 같은 시나리오는 아무것도 풀지 않는다.

**a083 이전에는 이 차단이 안 풀렸고 운영자가 봤을 것이다. a083이 자기 해제를 만들었다.**

### 왜 설계 리뷰가 놓쳤나

proposal-freeze의 적대적 검토(review.md 발견 1)는 "credit이 **너무 적게** 남는" 반례만
구성했다. 반대 방향 — credit이 **너무 오래** 남는다 — 을 보지 않았다. 셀프 리뷰는
자기가 방금 고른 축을 다시 본다.

### 고칠 방향 (아직 구현 안 함)

credit을 심볼이 아니라 **그것이 답한 차단**에 묶는다. `Block.Key()`나
reconcile_state의 `entered_at`을 함께 기록하고, credit의 비교보다 **나중에 생긴
차단은 그 credit으로 풀 수 없게** 한다. 여기에 명시적 만료(§I6)를 더한다.

## B2 — **차단 해제(blocking)**: 부분 수렴이 credit을 통째로 잃는다

`ConvergeQuantities`는 루프 **안에서** 심볼별 오류로 반환한다.

```go
if errors.Is(err, journal.ErrAdjustmentStale) {
    return report, fmt.Errorf("reconcile: ... re-collect the snapshot: %w", symbol, err)
}
...
if len(credited) > 0 { ... c.Credit.AdjustmentApplied(asOf, credited...) }   // 루프 밖
```

앞서 **이미 커밋된** 심볼들은 credit을 받지 못한다. 그리고 투영이 수렴했으므로 다음
비교는 그 심볼에 **동의**한다 — `diff.Quantities`에 다시 들어오지 않고,
`mismatch.go:84`가 적어 둔 대로 "일치하는 심볼은 다시 credit을 받을 길이 없다".

**a083이 고치려던 그 결함이 오류 경로에서 그대로 재현된다.** 운영자가 손으로 풀기
전까지 영구 차단이다. `reconcileloop.go:432`는 Converger 오류를 `cycle.Err`에 담고
`Observe`로 계속 가므로 도달하기 쉽다.

재현: `converged=1 credited=[] crediterCalls=[]` — AAPL은 계좌의 7로 수렴·커밋됐는데
아무것도 credit되지 않았다.

고칠 방향: 커밋된 심볼은 오류 반환 **전에** credit한다.

## I6 — credit에 만료가 없다

`t.adjusted`에 TTL도 상한도 없다. 다른 심볼이 diff를 계속 dirty하게 두면 어떤 심볼의
credit은 몇 시간이든 살아남고, 마침내 clean해진 비교에 쓰인다. B1의 창을 넓힌다.

## I7 — `Diff.AsOf`는 초 해상도인데 규칙은 strictly-after다

`compare.go:363`이 `Format("2006-01-02T15:04:05Z07:00")`로 초까지만 남긴다.
같은 초 안의 두 비교는 같은 스탬프를 갖고 credit은 영원히 쓰이지 않는다(fail-closed지만
조용하다). 60초 주기가 지금 가리고 있을 뿐 — 재시도·수동 재실행·더 짧은 주기에서 드러난다.
양쪽을 `time.RFC3339Nano`로.

## I8 — `AdjustmentApplied(comparison string, symbols ...string)`은 오용이 컴파일된다

`t.AdjustmentApplied("005930")`이 컴파일된다. 심볼이 `comparison`에 바인딩되고
`symbols`는 비고, 아무것도 기록되지 않으며 조용히 반환한다 — 그 호출자에 대해 a083
규칙이 통째로 꺼진다. 컴파일 오류도 로그도 테스트 실패도 없다.
`symbols []string`이나 작은 struct로 arity trap을 없앤다.

## I9 — `Observe`가 `defer` 없이 journal I/O 너머로 mutex를 들고 간다

`mismatch.go:445` `t.mu.Lock()` … `:577` `t.mu.Unlock()` 사이에 `t.persist`(SQLite)가
있다. 그 사이 panic이면 mutex가 영구히 잠기고, `EntryAllowed` → `Blocks()` →
`t.mu.Lock()`이므로 **이후 모든 진입 판정이 실패가 아니라 정지**한다. 기존 성질이지만
a083이 이 구간을 늘렸다.
