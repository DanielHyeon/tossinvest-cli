# a083 · Review

## 2026-08-05 · proposal-freeze

**보이스 구성**: Manager 셀프 적대적 Eng 1인.
**독립 리뷰어 부재 — 이것은 이 change의 알려진 약점이다.** 별도 컨텍스트의 리뷰어로
`codex exec`(gpt-5.6)를 호출했으나 사용량 한도로 거부됐다.

```
ERROR: You've hit your usage limit. … try again at Aug 8th, 2026 12:36 PM.
```

WORKFLOW의 "주문 실행·위험관리·원장·reconciliation을 건드리는 change는 리뷰 보이스에
반드시 적대적 Eng 관점을 포함한다"는 관점으로는 충족했으나, "작성자와 검증자의 분리"는
충족하지 못했다. 배포 전에 별도 세션의 재검증을 받아야 한다.

### 발견 1 (수용, 설계 변경) — credit 소멸이 비교 단위여서 새 구멍을 만든다

최초 설계는 "더 나중 as-of의 관측이 credit을 전부 소멸시킨다"였다. 적대적 검토에서
반례를 구성했다.

```
주기 N     A 불일치, B 불일치  → 둘 다 수렴, 둘 다 credit(T_N)
주기 N+1   A 일치,  B 불일치   → diff가 dirty → 해제 분기 미실행
                                 전부-소멸 규칙이면 A의 credit도 사라진다
주기 N+2   A 일치,  B 일치     → A는 credit이 없다. B만 해제된다
```

A는 다시 credit을 받을 길이 없다 — 일치하는 심볼은 `diff.Quantities`에 들어가지 않아
`ConvergeQuantities`가 건드리지 않는다. **고치려던 결함이 범위만 좁혀 남는다.**

수용. design D2b로 규칙을 심볼 단위로 바꿨다. spec delta에 "무관한 심볼이 불일치한다"
시나리오를 추가했다. tasks 3.11·3.12가 이 반례를 고정한다.

### 발견 2 (기록, 범위 밖) — 해제 분기 자체가 계좌 단위다

`Observe`의 해제는 `!diff.BlocksEntry()`, 즉 비교 **전체**가 일치할 때만 돈다. 어떤
심볼이 계속 불일치하면 다른 심볼의 해제도 그동안 미뤄진다. 차단은 심볼 범위인데 해제
조건은 계좌 범위다.

범위 밖으로 둔다. credit이 보존되므로 한 번이라도 전체가 일치하면 그 순간 전부
해제되고, 영구히 막히는 성질은 사라진다. 심볼 단위 해제는 별도 요구사항 변경이며
`issues.md`에 남긴다.

### 발견 3 (기록, 범위 밖) — missing-order 차단은 credit 발행자가 없다

`blocksFor`는 `diff.MissingOrders`에도 `Cause=QUANTITY_MISMATCH` ·
`Release=adjusted_reconcile` 차단을 만든다. 그런데 `ConvergeQuantities`는
`diff.Quantities`만 처리하므로 그 차단에는 credit이 붙지 않는다 — 상태표가 "adjusted
reconcile"로 자동 해제된다고 적어 두었지만 실제로는 운영자 전용이다.

운영 원장의 활성 8건은 전부 수량 불일치이므로 현재 live 문제는 아니다. `issues.md`에
기록한다.

### 발견 4 (기록) — 같은 심볼의 두 차단이 key를 공유한다

`Block.Key()`는 ScopeSymbol에서 `symbol|account|SYMBOL`이다. 한 심볼에 수량 불일치와
missing-order 차단이 동시에 생기면 map에서 하나만 남는다. 기존 성질이고 이 change가
악화시키지 않는다. `issues.md`에 기록한다.

### 발견 5 (수용, 검증 항목) — fake clock에서 as-of가 고정되면 credit이 영원히 안 쓰인다

as-of가 같으면 credit을 쓰지 않는 것이 설계다(fail-closed). 시계를 진행시키지 않는
테스트는 해제를 관측할 수 없다. 이것은 결함이 아니라 테스트 작성 규율이며, 기존
테스트를 고칠 때 `clk.Advance`와 diff의 as-of를 함께 움직여야 한다.

### 왜 기존 테스트가 결함을 놓쳤나 (근거)

`converge_test.go:113`의 `TestConvergenceMakesTheBlockReleasable`은 "the point of the
whole file"이라 적혀 있는데, 드라이버와 **호출 순서가 반대**다.

```
테스트     Observe(mismatch) → ConvergeQuantities(mismatch) → Observe(clean)
드라이버   ConvergeQuantities(mismatch) → Observe(mismatch) → [다음 주기] Observe(clean)
                                          ~~~~~~~~~~~~~~~~ 테스트에 없는 단계
```

테스트는 불일치를 수렴 **전에** 관측하고 다시 관측하지 않는다. 드라이버는 수렴
**후에** 관측한다. 그 한 단계가 결함 전부다. `mismatch_test.go`의 해제 테스트 4개도
`AdjustmentApplied`를 사이클 밖에서 직접 부른다.

## Pre-Edit Gate

```text
- change id / task id: a083-a-block-releases-itself / 5.1–5.4
- 대상 심볼(패키지.함수):
    reconcile.Tracker.Observe            (mismatch.go:360-491)
    reconcile.Tracker.AdjustmentApplied  (mismatch.go:314)
    reconcile.AdjustmentCrediter         (converge.go:73)
    reconcile.Converger.ConvergeQuantities (converge.go:136-254)
- 기존 동작 파악 근거:
    Function Logic Map 2건 + Branch Test Map 2건 (analysis/function-logic/, AST 20+15 분기 전수)
    CodeGraph callers: Tracker.Observe 14 (13 테스트 + reconcileloop.go)
                       Tracker.AdjustmentApplied 8 (전부 테스트)
                       Converger.ConvergeQuantities 9 (8 테스트 + reconcileloop.go)
    운영 원장 읽기 전용 조사: reconcile_states 활성 8건 release_cause NULL,
                              ADJUSTMENT_APPLIED 해제 이력 0건, position_adjustments에 수렴 기록 존재,
                              최근 사이클 로그 converged:0 blocked:8 released:0
- upstream 상속 테스트 영향: no (internal/reconcile는 TossOS 고유 패키지)
- 실패 테스트 선행 작성: yes (tasks 2.2, 3.1, 3.11, 4.1이 RED 정본)
- 안전 불변식 §0 위반 여부 검토: 통과
    §0.3 청산 즉시성 — 이 함수에 청산 경로 없음. ExitAllowed는 항상 true
    §0.7 운영 토글 flip — 자동 flip 없음. 영구 차단은 여전히 운영자 전용
    §0.9 단방향 안전 — design D3: 어떤 입력에서도 현재보다 먼저 해제되지 않는다
```

## 결정

**SHIP WITH CHANGES.** 발견 1을 설계에 반영한 뒤 구현한다. 발견 2·3·4는 `issues.md`.
배포 전 별도 세션의 독립 재검증이 남아 있다.

## 2026-08-05 · 독립 리뷰 (별도 컨텍스트)

앞의 proposal-freeze 기록은 **Manager 셀프 리뷰**였고, WORKFLOW가 요구하는
"작성자와 검증자의 분리"를 충족하지 못했다. 그 지적을 받고 gstack `/review`를 실행했다.

**보이스 구성**: 별도 컨텍스트 리뷰어 4명 — 적대적 Eng, 보안·안전 불변식, 데이터
마이그레이션, 테스팅/유지보수·성능. 각자 신선한 컨텍스트에서 diff를 읽었고 작성자의
근거를 알지 못한다.
**codex(gpt-5.6)는 여전히 사용량 한도**(2026-08-08 해제)로 교차 모델은 없다.

**결과: 셀프 리뷰가 놓친 blocking 결함을 찾았다.** `issues.md`의 `B` 항목이 정본이며,
여러 건은 실행 가능한 재현으로 확인했고 나는 원본 코드를 직접 읽어 재확인했다.

**결정: 배포 보류.** blocking 항목이 닫히기 전에는 `make gate` 통과 여부와 무관하게
a083 를 배포하지 않는다. 게이트는 테스트가 green이라고 말할 뿐, 여기서 발견된 것들은
테스트가 없어서 green이었다.
