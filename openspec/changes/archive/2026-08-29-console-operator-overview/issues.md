# Issues: console-operator-overview

WORKFLOW §예외 경로 기록. 구현 중 발견한 스펙·설계 결함과 그 처리.

## I-1 — D8의 seam 시그니처가 기존 안전 가드와 충돌한다 (분류: safe local, 보수 방향으로 해소)

**발견**: design.md D8과 spec delta는 Guardian 한도 seam을 `GateLimits() (config.AutomationGate, error)`로
고정했다. 그런데 `internal/console/engineproc_test.go`의 `TestTheConsoleDecidesNothingAboutTheGate`가
패키지 **전 파일**에서 식별자 `AutomationGate`·`Interlock`·`ProtectionReady`·`automation_gate`를 금지한다.
근거는 "콘솔은 엔진 프로세스에 묻고 그 답을 표시할 뿐, 게이트를 스스로 판정하지 않는다"이다.

즉 D8을 문자 그대로 구현하면 기존 안전 가드가 깨진다. 실제로 깨졌다:

```
--- FAIL: TestTheConsoleDecidesNothingAboutTheGate (0.00s)
    engineproc_test.go:181: overview.go names "AutomationGate"; the console asks the
    engine process and displays its answer, it does not evaluate the gate
```

**처리**: 가드를 완화하지 않고 seam을 **더 좁혔다**. `internal/console`에 값 타입
`console.GateLimits`(float64 5개 + 통화 문자열)를 두고 seam은 `GateLimits() (GateLimits, error)`로 한다.
`cmd/tossctl`의 어댑터가 config를 읽어 변환한다. 결과적으로 콘솔은 config 타입도, 게이트 판정 능력도
쥐지 않으므로 D8의 의도(별개 seam·읽기 전용·좁은 인터페이스)는 그대로 만족하고 표면은 더 작다.

**부수 결정**: 초안 구현에 있던 `게이트 ON/OFF` 배지를 화면에서 뺐다. spec의 ⑤는 "Guardian **한도**"를
요구하고 ON/OFF는 한도가 아니며, 콘솔이 게이트 상태를 스스로 그리는 것은 위 가드가 막으려는 인상
그 자체다. 화면은 대신 "게이트가 열려 있는지는 여기서 판정하지 않는다"를 명시한다.

**Manager 확인 요청**: spec delta의 `Guardian 한도 읽기는 … 별개의 한 메서드 seam이다` 문장은 그대로
유효하지만, design.md D8의 시그니처 예시(`config.AutomationGate`)는 실제와 다르다. archive 전에
design.md D8을 실제 시그니처로 정정할 것.

## I-2 — 다섯 사유로 표현할 수 없는 미측정이 하나 있다 (분류: safe local)

**발견**: spec은 미측정 사유를 다섯(`verify_suspended`·`broker_read_failed`·`journal_unreadable`·
`seam_unwired`·`never_fetched`)으로 고정한다. 그런데 Guardian 한도 seam이 **배선되어 있는데 config를
파싱하지 못한 경우**는 다섯 중 어디에도 맞지 않는다 — seam은 배선돼 있고, 브로커도 원장도 아니며,
캐시 문제도 아니다. 운영자의 대응은 "설정 파일을 고쳐라"다.

**처리**: 다섯 코드는 정확히 다섯으로 유지하고(테스트가 고정한다), 이 경우만 코드 없이
`unmeasuredBecause(<사유 문장>)`으로 렌더한다. 사유 문장은 비어 있을 수 없고
`TestNoUnmeasuredValueIsAReasonlessDash`가 그것을 고정한다. spec의 SHALL은 "다섯을 **구분한다**"이지
"다섯 외의 설명을 금한다"가 아니므로 위반은 아니라고 판단했다. 이견이 있으면 spec 문장을 조정할 것.

## I-3 — 통화가 다른 시장에는 일일 손실 한도 비율을 산출하지 않는다 (분류: safe local)

**발견**: `max_daily_loss_amount`는 `limit_currency` 한 통화의 한 숫자이고, 오늘 실현손익은 시장별
숫자다. spec은 "시장을 가로지르는 합계 금지"와 "실현손익 대 일일 손실 한도 한 축 산출"을 동시에
요구하는데, 둘을 함께 지키려면 통화가 일치하는 시장에서만 비율을 낼 수 있다.

**처리**: 축은 시장별 행으로 렌더하고, 한도 통화와 시장 통화가 다르면 두 숫자를 나란히 보이되
비율은 내지 않으며 그 이유를 행에 적는다. `TestTheDailyLossAxisIsNotComparedAcrossCurrencies`가 고정.

---

## Manager 판정 (2026-07-28)

**I-1 — 승인.** design.md D8을 정정했다. 시그니처 예시를 지우고, 왜 그것이 틀렸는지를 함께
남겼다: **D3이 지적한 것과 같은 부류의 실수를 D8이 저질렀다** — 새 표면을 설계하면서 그 표면에
걸리는 가드를 확인하지 않았다. 라운드 1 리뷰 둘도 이것을 못 봤다.

가드를 완화하지 않고 seam을 더 좁힌 처리가 옳다. 게이트의 자기 타입을 쥔 콘솔은 스스로
"게이트가 괜찮아 보인다"고 판단하는 코드에서 한 편집 거리에 있다.

**부수 결정(게이트 ON/OFF 배지 제거)도 승인.** spec ⑤가 요구하는 것은 한도이고 ON/OFF는
한도가 아니다. spec에 명시 문장을 추가했다.

**I-2 — 처리를 바꾼다. 여섯 번째 코드를 추가한다(`config_unreadable`).**

"다섯을 구분한다"가 "여섯째를 금한다"는 아니라는 읽기는 문법적으로 맞다. 그럼에도 코드를
추가하는 쪽으로 판정하는 이유는 이 열거의 **일**에 있다 — 열거는 "운영자가 무엇을 고쳐야
하는가"를 남김없이 적는 것이다.

자유 문장으로만 존재하는 사유는 셀 수 없고 **없음을 테스트할 수도 없다**. 그러면 다음 사람이
일곱 번째도 문장으로 쓰고, 그 시점에 열거는 표면을 더 이상 기술하지 않는다. 이 저장소가
반복해서 겪은 실패가 정확히 그 형태다 — 규율이 있는데 새 표면에는 없는 상태.

그리고 이 경우의 운영자 대응은 다른 다섯과 확실히 다르다: 고칠 파일이 config다.
spec·design·계약값을 여섯으로 고쳤다. `unmeasuredBecause` 문장 강제는 그대로 두되, 이 사유는
코드를 갖는다.

**I-3 — 승인.** 처리가 옳고, spec이 자기모순으로 읽히지 않도록 문장을 보강했다: 한도 통화와
시장 통화가 다르면 두 숫자를 나란히 보이되 비율을 내지 않고 이유를 표시한다.
**환산해서 비율을 만드는 것은 D6이 금지한 합계를 한 칸 옆에서 다시 만드는 것**이라는 근거를
명시했다.

**보고된 것 중 조치하지 않은 것**: `holdingsSnapshot.Stale()`이 호출자 없는 죽은 코드라는
관찰. 주석 정정은 2.2가 요구한 대로 됐다. 삭제는 이 change의 범위가 아니며, 지금 지우면
`peek` 도입과 얽혀 diff를 읽기 어렵게 만든다. 별도 정리로 남긴다.

---

## 라운드 2 리뷰 반영 (2026-07-28, 구현)

독립 리뷰 둘이 P0 셋·P1 다섯·P2 열을 냈다. 아래는 그중 **스펙·설계와 어긋나는 처리**만 기록한다.
가드 자체의 결함 수정(P0 셋)과 코드 결함 수정은 스펙 편차가 아니므로 여기 적지 않는다.

## I-4 — 미측정 사유에 일곱 번째 코드를 추가했다 (분류: safe local, Manager 확인 요청)

**발견**: I-2 판정으로 `config_unreadable`이 여섯 번째 코드가 됐고 spec·design·계약값이 여섯으로
고쳐졌다. 그런데 **코드 없는 자유 문장으로 남아 있는 미측정이 아직 둘** 있었다.

1. `trade_outcomes.closed_at` / `realized_pnl_after_costs`를 해석하지 못한 경우 — **도달 가능**하다.
   원장은 열렸고 답도 했는데 그 안의 값 하나가 시각이나 숫자가 아니다.
2. 시장의 시간대(`clock.Market.Location()`)를 읽지 못한 경우 — tzdata가 바이너리에 embed되어
   있으므로 **도달 불가**하다.

**처리**: 1번에 일곱 번째 코드 `journal_value_unparsable`을 부여했다. `journal_unreadable`을
재사용하지 않은 이유는 그 사유의 지시가 "엔진을 한 번 기동해 마이그레이션하라"인데, 멀쩡히 열린
원장에 그렇게 해도 아무것도 달라지지 않기 때문이다 — **틀린 지시는 사유 없음보다 나쁘다.**
2번은 `unmeasuredBecause` 자유 문장으로 남겼다: 도달 불가 경로에 코드를 만들면 열거에 검증할 수
없는 항목이 하나 늘어난다.

근거는 I-2 판정문 그대로다 — "자유 문장으로만 존재하는 사유는 셀 수 없고 없음을 테스트할 수도
없다. 그러면 다음 사람이 일곱 번째도 문장으로 쓴다." 판정문이 가리킨 그 일곱 번째가 이미
코드에 있었다.

**Manager 확인 요청**: spec delta의 `사유는 … 여섯을 구분한다`와 design.md D5의 표·계약값
`unmeasured_reasons`는 여섯으로 고정되어 있다. 구현은 일곱이다. I-2에서 승인된 문법적 읽기
("여섯을 구분한다"가 "일곱째를 금한다"는 아니다)에 기대고 있으므로, archive 전에 spec·design을
일곱으로 정정하거나 이 처리를 되돌릴지 판정할 것. **`internal/console`은 이 change의 파일 표면이고
spec·design은 아니므로 구현자가 고치지 않았다.**

## I-5 — `verify_suspended`는 개요 화면에서 더 이상 산출되지 않는다 (분류: safe local)

**발견**: `brokerReadable`의 문서 주석은 "한 번도 채워지지 않은 캐시는 검증이 돌고 있어도
`never_fetched`로 읽는다"고 적혀 있었는데 코드는 반대로 `verify_suspended`를 먼저 반환했다.
그 화면에서 그 문장("검증 중 — 갱신 보류 … 끝나면 다시 읽힌다")은 **양쪽으로 다 거짓**이다.
개요는 계약상 브로커를 부르지 않으므로 보류되는 것이 없고, 검증이 끝나도 그 값은 읽히지 않는다 —
이 캐시를 채우는 것은 `/positions`를 여는 것뿐이다.

**처리**: 코드를 주석에 맞췄다. 개요의 계좌 패널에서 `verify_suspended`는 산출되지 않는다.
사유 열거에서 **지우지는 않았다** — spec이 여섯(일곱) 중 하나로 고정하고 있고, 다른 seam이
배선되면(예: 주문 조회) 다시 산출될 수 있는 어휘다.

대신 주석이 약속하던 "the page says separately"를 실제로 구현했다: 검증이 실행 중이면 계좌 패널
아래에 **값의 사유가 아니라 별도 안내**로 "포지션 화면의 갱신이 검증 종료까지 보류된다"를 적는다.
운영자가 실제로 겪는 것이 그것이기 때문이다 — 개요가 무언가를 못 읽는 것이 아니라, 개요가
가리키는 탈출구가 잠겨 있다.

## I-6 — KR·US 어느 쪽도 아닌 시장의 행을 `기타/미분류`로 모은다 (분류: safe local)

**발견**: 계좌·오늘 패널은 `overviewMarkets`(KR/US) 라벨로 행을 매칭하고, 어느 쪽에도 맞지 않는
보유·왕복은 **행도 표식도 없이 사라졌다**. 같은 콘솔의 `/positions`·`/history`는 그것을 계속
보여준다. 두 화면이 계좌에 대해 다른 말을 하는 상태이고, D10이 금지하는 은폐를 셀이 아니라
심볼 단위로 저지른 것이다.

**처리**: 매칭되지 않은 행을 `기타/미분류` 한 행에 모아 **미측정으로** 렌더하고 심볼 목록과
사유 문장을 적는다. 합산하지 않는 이유는 D6이다 — 그 행들의 통화를 콘솔은 모르고, 모르는 통화를
가정해 만든 평가액은 D6이 금지한 그 합계다.

spec의 `②와 ③은 시장별(KR/US)로 나누어 표시하고`는 그대로 만족한다. 추가되는 것은 "그 둘 중
어느 쪽도 아닌 것이 있으면 있다고 말한다"이며, 같은 요구사항의 `값이 없는 항목을 화면에서
조용히 생략하지 않는다`가 요구하는 바다.

---

## 증거 생산 pass 후속 — `TestNoCapabilityReachesTheConsoleAroundOptions`의 공회전 대조 (2026-07-28)

이 change의 리뷰 지적 P0-3을 막으려고 만든 가드가, 자기 걷기가 비었을 때는 조용히 통과했다.
세 걷기(Options seam 해석 / 패키지 전역 인터페이스 순회 / `*Console` exported 메서드 순회) 중
단언되던 것은 seam 하나뿐이고 인터페이스 계수는 `_ = checkedInterfaces`로 버려졌다. 세부와 변이
증거는 `add-candidate-discovery/issues.md` §11 G-1에 한 벌로 적었다(같은 함수, 같은 수정).

이 change 쪽 조치는 동일하다: 세 걷기에 각자 대조를 붙이고(B21·B22·B23–B25·B26),
`analysis/function-logic/internal-console--testnocapabilityreachestheconsolearoundoptions/`의
ast.json·분기표·논리 지도를 새 본문에 맞춰 다시 만들었다. 이 change의 다른 target 중
`internal/console/static_test.go`에 묶인 `current` revision 전부도 파일 해시가 바뀌어 재생성했다
(본문 무변경이므로 분기표는 그대로다).
