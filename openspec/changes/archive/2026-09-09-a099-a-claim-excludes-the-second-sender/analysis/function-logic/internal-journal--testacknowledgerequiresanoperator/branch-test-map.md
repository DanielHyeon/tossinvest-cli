# Branch Test Map: `TestAcknowledgeRequiresAnOperator`

`ast.json`의 열거가 정본이다: 분기 1 · 이탈 0 · **`revision: base`**.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:145` 공백 이름으로 확인하면 거절된다 | `TestAcknowledgeRequiresAnOperator` | **no — a099가 이 함수를 안 건드렸다** | **yes** |

## 이 번들이 `base`인 이유

`check_analysis.py`가 요구한 리비전이 `base`다. **본문이 안 바뀌었기 때문이다.**

a099는 바로 다음 함수(`TestMarkingANonPendingAlertIsRefused`) 위에 doc comment
여덟 줄을 더했고, base 좌표에서 그 삽입이 이 함수의 끝 줄에 닿았다.
worktree 쪽에서는 이 함수가 어떤 hunk와도 안 겹치므로 checker가 `base`로 분류한다.

**`git show <base>:internal/journal/outbox_test.go`의 이 함수와 지금 파일의 이
함수는 바이트 단위로 같다.** 그것이 이 표에 RED가 없는 이유다.

## a099가 이 함수를 안 건드린 것이 결정이다

`AcknowledgeAlert`의 시그니처에 토큰을 더했다면 이 테스트가 깨졌을 것이다.
**안 더했다** — 사람의 확인은 임차 위에 있고, 발송자가 쥔 토큰을 운영자가
가지고 있을 이유가 없다.

그 결정을 직접 고정하는 것은 `TestAcknowledgementIgnoresTheLease`
`a099_regression_pins_test.go:101`이고, **이 테스트는 그 결정의 부작용으로
안 깨진 것**이다.

## 덮이지 않은 것을 이름으로 적는다

- **`EnqueueAlert`의 오류를 버린다** (`:143`). 배치가 실패하면 `id = 0`이고
  `AcknowledgeAlert`는 **다른 이유로** 오류를 낸다 — 테스트는 통과하지만
  검사한 것이 달라진다. a099가 만든 문제가 아니다.
- **공백 셋 하나만 본다.** 빈 문자열도, 탭도, 유니코드 공백도 안 본다.
- **확인이 성공하는 경로를 안 본다** — `TestDeliveryAndAcknowledgementAreDistinctStates`
  `outbox_test.go:117`가 본다.
