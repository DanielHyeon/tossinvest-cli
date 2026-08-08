# Branch Test Map: `claimOwed`

측정: `go test -covermode=set -coverprofile ./internal/journal/` — RED 총계 **75.0%**.
블록 단위로 프로파일에서 직접 읽었다.

**a097은 이 함수의 본문을 바꾸지 않는다.** 표의 목적은 proposal R3의 구조 주장을
커버리지와 함께 고정하는 것이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:251` `switch state` (구조 분기) | 아래 B2·B3·B8이 대표한다 | 진입 (셋 다 진입) | 변화 없음이어야 함 |
| B2 | `:252` PENDING — 미완이므로 owed | a096 기존 | 진입 (`252-254 count=1`) | 변화 없음이어야 함 |
| B3 | `:255` settled (DELIVERED/ACKNOWLEDGED) | a096 기존 | 진입 (`255-256 count=1`) | 변화 없음이어야 함 |
| B4 | `:256` `remindAfter <= 0` → 재무장 안 함 | a096 기존 + **a097 2.9 (신규)** | 진입 (`256-258 count=1`) | 변화 없음이어야 함 |
| B5 | `:260` 스탬프를 하나도 못 읽음 → fail-open | 없음 — 원장 손상 주입 (`not-applicable`) | 미진입 (`260-264 count=0`) | 변화 없음이어야 함 |
| B6 | `:266` 미래 스탬프 → fail-open | a096b `TestASettledStampInTheFutureStillOwesDelivery` | 진입 (`266-279 count=1`) | 변화 없음이어야 함 |
| B7 | `:280` 창 안이라 owed 아님 | a096 기존 | 진입 (`280-282 count=1`) | 변화 없음이어야 함 |
| B8 | `:284` 미지 상태 → owed + 재무장 | a096 기존 + **a097 2.8 (신규)** | 진입 (`284-289 count=1`) | 변화 없음이어야 함 |

## 커버리지가 R3을 증명하지 않는다는 것

B4도 B8도 RED에서 이미 진입한다. 그래서 **커버리지로는 P2 ②를 판정할 수 없다.**
`-covermode=set`은 "그 블록이 실행됐다"만 말하고 *어떤 `remindAfter` 값으로* 실행됐는지는
말하지 않는다.

판정한 것은 AST다. `B4@256`이 `B3@255`의 **자식**이고 `B8@284`이 `B3`의 **형제**라는 구조
사실이 "default가 `remindAfter`를 무시한다"를 결함이 아니라 설계로 만든다. 그리고
`.claude/CLAUDE.md`가 요구하는 것이 정확히 이것이다 — 분기를 근거로 삼는 문서는 AST 열거를
먼저 만든다. 손으로 읽었으면 "무시한다"에서 멈췄을 것이다.

a097 2.8·2.9는 그 두 규칙을 각각 **명시적인 `remindAfter=0`**으로 고정한다. 커버리지 값은
바뀌지 않지만 단언은 새로 생긴다. 주석이 지워져도 테스트는 실패한다.

## GREEN에서 이 표가 바뀌면 안 된다

a097은 이 함수를 편집하지 않는다. GREEN 측정에서 어떤 칸이든 값이 바뀌면 그것은
a097이 의도하지 않은 곳을 건드렸다는 신호다. 유일하게 허용되는 변화는 `B5@260`이 새
테스트의 부수 효과로 진입하는 경우이며, 그런 일이 생기면 이유를 review.md에 적는다.
