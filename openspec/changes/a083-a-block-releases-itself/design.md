# a083 · Design

## 문제의 형태

credit은 두 사실을 잇는 증거다.

```
사실 A   조정이 기록되어 투영이 계좌값으로 수렴했다      (ConvergeQuantities)
사실 B   그 뒤에 수집된 비교가 일치한다                  (다음 주기의 Compare)
해제     A ∧ B  →  ADJUSTMENT_APPLIED
```

지금 코드에는 **"그 뒤에"를 판정할 수단이 없다.** credit은 심볼 집합
(`map[string]bool`)일 뿐이고, "다음에 도착하는 관측"을 B로 근사한다. 드라이버에서
다음에 도착하는 관측은 A를 계산한 바로 그 비교이므로 근사가 깨진다.

`Diff.AsOf`가 그 수단이다. 이미 존재하고, `ConvergeQuantities`가 조정의
`BrokerAsOf`로 원장에 기록하고 있으며, 없으면 수렴 자체를 거부한다.

```go
// converge.go — 이미 이렇게 되어 있다
asOf := strings.TrimSpace(diff.AsOf)
if asOf == "" {
    return report, fmt.Errorf("reconcile: the comparison carries no as-of, ...")
}
```

즉 프로덕션에서 credit이 만들어질 때 as-of는 **항상** 있다.

## D1 — 드라이버 순서를 바꾸지 않고 Tracker 안에 불변식을 넣는다

**대안 A: `Observe`를 `ConvergeQuantities` 앞으로 옮긴다.** 두 줄 이동이면 증상은
사라진다. 채택하지 않는다.

- 해제 규칙이 **호출 순서**에 의존하게 된다. 규칙은 `reconciliation` 스펙의 SHALL이고,
  그것을 지키는 책임이 규칙을 모르는 파일(`reconcileloop.go`)로 옮겨간다.
- `reconcileloop.go`의 현재 순서는 우연이 아니라 명시된 설계다(design A6, 파일 주석
  "Steps 5 and 6 are in that order because the design fixes it"). 그 순서 자체는
  옳다 — 트래커는 비교가 만든 diff를 봐야 하고, 조정의 결과는 다음 재조회에 나타난다.
  틀린 것은 credit의 수명이지 단계 순서가 아니다.
- 미래의 호출자가 순서를 되돌리면 결함이 조용히 돌아온다. 테스트는 순서를 검사하지
  않는다.

**채택 B: credit에 그것이 계산된 비교를 각인한다.** 규칙이 그것을 소유한 컴포넌트
안에서 자기 증명된다. 호출 순서가 어떻든 "조정 이후의 재조회"만 credit을 쓴다.
이 저장소의 기존 태도와도 맞는다 — `ExitAllowed`가 "a method rather than an absence
so that §0.3 is something a test can assert"인 것과 같은 이유다.

## D2 — "그 뒤에"는 엄격한 후행이다

```
credit.comparison = C        관측 diff.AsOf = O

O가 C보다 엄격히 뒤     →  이 관측이 credit을 쓴다 (일치면 해제, 불일치면 소멸)
O == C                  →  같은 비교다. 쓰지도 버리지도 않는다        ← 결함의 지점
O가 C보다 앞            →  순서를 신뢰할 수 없다. 쓰지 않는다
C 또는 O를 시각으로 읽을 수 없다 → 쓰지 않는다
```

마지막 두 줄은 fail-closed다. 판정할 수 없으면 차단이 유지된다.

`Snapshot.AsOf`는 수집 시작 시각(`clk.Now()`)이고 주기는 60초, 안정화 대기는 2초이므로
서로 다른 사이클의 as-of가 같은 초에 떨어지는 일은 없다. 그럼에도 문자열 동등이 아니라
**파싱 후 시각 비교**를 쓴다. 형식이 바뀌어도 의미가 유지되고, 판정 불가가 조용한
동등이 아니라 명시적 fail-closed가 된다.

## D3 — §0.9: 이 변경은 어느 방향인가

이 change는 **진입 차단을 푸는** 방향이다. §0.9는 손절·익절·사이징의 단방향 안전을
요구하고, §0.7은 운영 토글 flip을 사람 승인으로 묶는다. 둘 다 이 변경에 직접
적용되지는 않지만, 판단 근거를 남긴다.

1. **새 권한을 만들지 않는다.** 해제 조건은 이미 승인된 SHALL이고 시나리오까지 있다.
   지금 코드는 스펙보다 **엄격**하다 — 승인된 자동 해제를 한 번도 실행하지 않는다.
   이 change는 승인 범위를 넓히는 것이 아니라 미구현을 구현한다.
2. **해제는 주문을 내지 않는다.** 차단이 풀리면 편입 판정이 재개될 뿐이고, 편입은
   그 포지션에 손절선을 **부여한다**. 현재 상태 — 차단된 채 보호선 없이 보유 — 가
   덜 안전한 쪽이다. 042660은 21시간째 손절 없이 있다.
3. **증거 요건은 그대로다.** 조정이 기록되어야 하고(A), 그 뒤 독립 재조회가 일치해야
   한다(B). 이 change는 B의 정의를 느슨하게 하지 않고 **엄격하게** 한다 — 지금은
   "다음에 오는 아무 관측"이고, 바뀐 뒤에는 "조정보다 나중에 수집된 비교"다.
4. **영구 차단은 손대지 않는다.** 3회 연속 실패 승격과 운영자 전용 해제는 그대로다.
5. **판정 불가는 차단 유지다.** as-of가 없거나 읽히지 않으면 credit을 쓰지 않는다.

결론: 어떤 입력에서도 지금보다 **먼저** 해제되지 않는다. 지금 해제되는 입력은 없으므로
(운영 0건), 새로 해제되는 입력은 전부 스펙이 이미 해제하라고 적어 둔 것뿐이다.

## D2b — 소멸은 심볼 단위다 (적대적 리뷰에서 발견)

첫 설계는 "더 나중 as-of의 관측이 credit을 **전부** 소멸시킨다"였다. 그 규칙에는 구멍이
있다.

```
주기 N     A 불일치, B 불일치   →  둘 다 수렴, 둘 다 credit(T_N)
주기 N+1   A 일치,   B 불일치   →  diff가 dirty이므로 해제 분기가 돌지 않는다
                                   전부-소멸 규칙이면 A의 credit도 함께 사라진다
주기 N+2   A 일치,   B 일치     →  A는 credit이 없다. B만 해제된다
```

A는 다시 credit을 받을 길이 없다. 일치하는 심볼은 `diff.Quantities`에 들어가지 않으므로
`ConvergeQuantities`가 건드리지 않는다. **A는 영구히 막힌다** — 고치려던 결함과 같은
모양이, 범위만 좁아져 남는다.

원인은 규칙을 비교 단위로 쓴 것이다. 스펙 문장은 심볼 단위다 — "조정 이벤트가 반영된
뒤의 재조회 **일치**"에서 그 일치는 조정된 그 심볼에 대한 일치다. 무관한 심볼의
불일치는 A의 조정이 통했는지에 대한 답이 아니다.

```
credit (S, C), 관측 비교 O

O가 C보다 엄격히 뒤 아님    →  그대로 둔다 (같은 비교이거나 순서 불명)
O가 뒤 && diff가 S에 불일치 →  소멸. 그 조정은 S를 해결하지 못했다
O가 뒤 && diff가 S에 일치    →  보존. 전체 비교가 일치하는 관측에서 해제된다
해제됨                      →  소멸. 할 일이 끝났다
```

`diff`가 S에 불일치한다는 것은 S가 `diff.Quantities`나 `diff.MissingOrders`에 있다는
뜻이다.

**남는 제약(범위 밖).** 해제 분기 자체는 `!diff.BlocksEntry()` — 비교 전체가 일치할
때만 돈다. 어떤 심볼이 계속 불일치하면 다른 심볼의 해제도 그동안 미뤄진다. 이것은
기존 동작이고 이 change는 넓히지 않는다. 다만 credit이 보존되므로 **한 번이라도 전체가
일치하면 그 순간 전부 해제된다** — 영구히 막히는 성질은 사라진다. 심볼 단위 해제는
별건이며 `issues.md`에 기록한다.

## D4 — persist 실패 시 credit 회계

현재 코드는 clean diff + persist 실패일 때, 아직 활성인 차단의 credit만 남긴다.
각인 후에도 같은 의도를 유지하되 대상이 좁아진다.

D2b의 심볼 단위 규칙 아래에서 관측이 끝날 때 credit은 세 무리로 나뉜다.

```
untouched  O가 C보다 뒤가 아니다 (같은 비교 포함)            → 언제나 보존
answered   O가 뒤이고 diff가 그 심볼에 일치한다               → 해제가 커밋되면 삭제
refuted    O가 뒤이고 diff가 그 심볼에 여전히 불일치한다      → 언제나 삭제
```

`refuted`는 persist 결과와 무관하다 — 원장에 무엇을 썼든 그 조정이 심볼을 해결하지
못했다는 사실은 그대로다. `answered`만 persist 결과를 본다.

```
성공          → answered 삭제 (해제됨), refuted 삭제, untouched 보존
persist 실패  → 해제가 커밋되지 않은 answered는 보존해 다음 관측이 다시 시도하게 한다,
                refuted 삭제, untouched 보존
```

`untouched`가 항상 보존되는 것이 이 change의 핵심이다. 같은 사이클의 관측은 credit을
건드리지 않고 지나간다.

## D5 — 인터페이스 변경의 파급

```go
// before
type AdjustmentCrediter interface{ AdjustmentApplied(symbols ...string) }
func (t *Tracker) AdjustmentApplied(symbols ...string)

// after
type AdjustmentCrediter interface{ AdjustmentApplied(comparison string, symbols ...string) }
func (t *Tracker) AdjustmentApplied(comparison string, symbols ...string)
```

프로덕션 호출자는 `Converger.ConvergeQuantities` 하나뿐이고 그 자리에 `asOf`가 이미
지역 변수로 있다. 나머지 호출자는 전부 테스트다.

**선택적 인자나 두 번째 메서드를 두지 않는다.** `AdjustmentAppliedAt`을 추가하고
기존 메서드를 "각인 없는 credit"으로 남기면, 각인 없는 credit이 무엇을 의미하는지
결정해야 한다. 어느 쪽으로 정하든 나쁘다 — "아무 관측이나 쓸 수 있다"면 결함이 그대로
남는 경로가 생기고, "아무도 쓸 수 없다"면 조용히 죽은 API가 된다. 인자를 필수로 만들면
호출자가 자신이 어느 비교를 근거로 삼는지 말하지 않을 수 없다.

기존 테스트는 각인을 넣도록 수정한다. 그 수정 자체가 개선이다 — 지금 테스트는
드라이버가 실제로 하는 일을 재현하지 않는 것이 결함을 놓친 이유다.

## D6 — RED이 되어야 하는 테스트

단위 테스트만으로는 부족하다. 결함은 `Tracker` 안이 아니라 **`Converger`와 `Tracker`가
한 사이클 안에서 만나는 지점**에 있다. 그래서 드라이버 수준 테스트가 정본이다.

```
주기 1:  계좌 4, 투영 10  →  diff 불일치  →  수렴(10→4) + credit  →  Observe(같은 diff)
         기대: 차단 유지, credit 보존, released = 0
주기 2:  계좌 4, 투영 4    →  diff 일치     →  수렴 없음            →  Observe(새 diff)
         기대: ADJUSTMENT_APPLIED로 해제, reconcile_states.released_at 기록,
               entry gate에서 심볼 차단 해제
```

주기 2가 현재 코드에서 **실패**해야 한다. 이것이 RED의 정본이고, 나머지는 회귀다.

---

## 개정 2 (2026-08-05) — 독립 리뷰가 연 blocking 두 건

`issues.md` B1·B2. 최초 설계는 credit의 **수명**만 규정하고 **적용 범위**를 규정하지
않았다. 그래서 credit이 자기가 답하지 않은 차단을 풀 수 있었다.

### D7 — credit은 자신이 답한 차단만 푼다 (B1)

차단은 `Block.Since`를 갖는다. 이것은 durable `reconcile_states.entered_at`이다
(`mismatch.go:890`이 restore에서 `state.EnteredAt`을 넣는다).

**규칙: credit의 비교 시각보다 나중에 생긴 차단은 그 credit으로 풀 수 없다.**

```
release(block, credit) 가능 ⟺
    creditUsableBy(credit.comparison, diff.AsOf)      # 기존: 엄격히 나중의 재확인
  ∧ !block.Since.After(credit.comparison)             # 신규: credit이 이 차단을 봤다
```

대안으로 `Block.Key()`를 쓰지 않은 이유: `ScopeSymbol`의 key는
`symbol|account|SYMBOL`이라 차단 **인스턴스**를 구별하지 못한다(`issues.md` I3).
시각은 구별한다.

### D8 — 해제는 "막지 않는다"가 아니라 "그 심볼에 동의한다"를 요구한다 (B1 재현 A)

`BlocksEntry()`는 `Quantities`와 `MissingOrders`만 센다. `ExternalPos` — 계좌가 들고
있는데 엔진이 모르는 포지션 — 는 진입을 막지 않는다. 그래서 한 심볼이
`QuantityMismatch(local 10, broker 0)`에서 `ExternalPos(broker 10, local 0)`로
**재분류**되면 비교가 "clean"이 되고 차단이 풀린다. 불일치는 사라지지 않았다.
방향만 뒤집혔다.

**규칙: 해제하려는 심볼이 `ExternalPos`에도 없어야 한다.** 그리고 `symbolsInDispute`가
`ExternalPos`를 포함하도록 넓혀, 그 재분류가 credit을 **반박**하게 한다.

`BlocksEntry()` 자체는 바꾸지 않는다. 그것은 진입 게이트의 계약이고 a083의 범위 밖이다
(`issues.md` I2와 같은 성격). 여기서 좁히는 것은 **해제 조건**뿐이며, 좁히는 방향은
언제나 §0.9가 요구하는 보수 방향이다.

### D9 — 답할 차단이 없는 credit은 버린다 (I6, 만료 대신)

TTL 대신 **참조 무결성**으로 만료를 만든다. 관측 끝 정산에서, 쓸 수 있게 된 credit 중
해당 심볼에 **활성 차단이 하나도 없는** 것은 버린다.

```
credit 폐기 ⟺ usable ∧ ( disputed ∨ committed ∨ 활성 차단 없음 )
```

근거: 차단이 없으면 그 credit은 답할 것이 없다. 남겨 두면 미래의 차단을 풀 수 있을
뿐이고, D7이 이미 그것을 금지한다. 따라서 버리는 것이 항상 안전하고,
`t.adjusted`의 크기가 **차단된 심볼 수**로 묶인다. 벽시계 TTL이 필요 없다.

같은 주기의 관측은 영향받지 않는다 — `!creditUsableBy(...) { continue }`가 앞에 있어
아직 쓸 수 없는 credit은 정산 자체를 건너뛴다. 그것이 a083의 원래 수정이다.

### D10 — 커밋된 수렴은 반환 전에 credit한다 (B2)

`ConvergeQuantities`는 심볼별 오류로 루프 안에서 반환하는데 crediter 호출은 루프 밖에
있어, **이미 커밋된** 심볼이 credit 없이 남는다. 그리고 투영이 수렴했으므로 다음 비교는
그 심볼에 동의하고 — `mismatch.go:84`가 적은 대로 — 다시 credit받을 길이 없다.
영구 차단이며 a083이 고치려던 결함 그 자체다.

**규칙: 오류로 반환하기 전에 그때까지 커밋된 심볼을 credit한다.** `report.Credited`도
같이 채운다(현재는 비어 있어 자기 doc comment와 모순).

부분 credit이 안전한 이유: credit은 해제를 **허가**하지 않는다. D7·D8을 통과한 뒤
엄격히 나중의 비교가 동의할 때만 해제가 일어난다. 커밋된 수렴에 credit을 주는 것은
"이 심볼에 조정이 기록됐다"는 사실 진술이며, 그것은 참이다.
