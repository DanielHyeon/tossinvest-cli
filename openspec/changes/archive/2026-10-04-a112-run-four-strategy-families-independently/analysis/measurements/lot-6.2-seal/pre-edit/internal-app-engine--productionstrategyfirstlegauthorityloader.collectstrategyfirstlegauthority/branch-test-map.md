# Branch Test Map: `collectStrategyFirstLegAuthority`

- Source SHA-256: `1d710710098c03669719779609db137d0660a361a025d070128e8993772ed063`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 211:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | if at 217:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | if at 223:2 | lineage 쪽은 `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` 와 `…ASiblingCampaignOnTheSameSymbol`; **가격 세 축은 각각** `…RewrittenExecutionTermsUnderTheSameLineage` 의 하위 시험 셋 (`internal/app/engine/strategy_first_leg_identity_backstop_test.go`, `tossos_testseams`). stop 출처 셋을 함께 바꾼 경우는 `…RefusesARestatedStopProvenanceAtTheSamePrice`. 나머지 필드 축은 행동이 아니라 **구조 단언 하나 + strategyflow 의 32 필드 표**가 덮는다(아래) | yes | yes |
| B4 | if at 227:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 232:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 239:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 250:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B8 | if at 259:3 | planned targeted RED before any edit; not run by L0 | no | no |
| B9 | if at 263:3 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.

## B3 의 RED/GREEN 실측 (5.5-fix5 · fix6 · fix8 · fix9 · fix10 · fix12 · fix13)

B3 은 위조된 봉투가 나르는 결과를 거절하는 자리다. 5.5-fix5 이전에는 **어떤
시험도 이 분기를 밟지 않았다** — 거절 문구를 검색하면 생산 코드 한 줄만 나왔다.

RED 은 뮤테이션으로 잰다. 저장소 파일은 건드리지 않고 `go test -overlay` 로 넣었고,
원본 sha256 이 `1d710710…` 로 그대로임을 앞뒤로 확인했다.

| 뮤테이션 | 명령 | 결과 |
|---|---|---|
| `223`–`225` 블록 삭제 | `go test -overlay=… -tags tossos_testseams -run TestFirstLegAuthorityRefuses ./internal/app/engine/` | RED — 거절이 `production risk authority scope changed` 로 바뀐다 |
| 술어를 `Lineage.Market` 비교로 약화 | 같은 명령 | RED — `…ASiblingCampaignOnTheSameSymbol` 과 `…RewrittenExecutionTermsUnderTheSameLineage` 가 실패. 앞의 것의 실패값이 `err=<nil>` 이다(같은 종목의 다른 캠페인으로 1차 진입이 나간다) |
| 오른쪽 논리항 삭제(= `Lineage.Identity` 만 비교) | 같은 명령 | RED — `…RewrittenExecutionTermsUnderTheSameLineage` **만** 실패. 앞의 두 시험은 초록이다 |
| `Lineage.CampaignID` 만 비교 | 같은 명령 | RED — 위와 같다 |
| 왼쪽 논리항 삭제(= `ExecutionTerms.Identity()` 만 비교) | 같은 명령 | **GREEN(동등 변이)** — terms identity 해시가 lineage identity 를 품으므로 왼쪽 항은 오른쪽에 포섭된다 |

두 번째가 있어야 하는 이유는 5차 적대 리뷰가 반증했다: 첫 시험만 두면 시장까지
함께 바뀌므로 술어가 identity 대신 market 을 비교해도 초록으로 남는다.

**세 번째가 있어야 하는 이유는 6차가 반증했다.** `:223` 은 논리합 둘이고,
`executionTermsIdentity` 가 `lineageIdentity` 를 해시에 넣으므로
(`internal/strategyflow/types.go:315`) **왼쪽 항은 오른쪽에 포섭된다.** 앞의 두
시험은 campaignID 를 바꿔 lineage 를 움직이므로 두 항이 함께 참이 되고, 그래서
오른쪽 항만 지워도 둘 다 초록이다. 그 상태에서 남는 것은 수량과 세 가격을
지키는 항이 사라진 코드다 — 손절가를 바꾼 제안이 1차 진입을 낸다.

| 이 분기가 지키는 것 | 어느 항이 | 어느 시험이 |
|---|---|---|
| lineage(계좌·시장·종목·캠페인·세대…) | 왼쪽(포섭됨) 또는 오른쪽 | 앞의 두 시험 |
| entry 가격의 **숫자**(`priceMinor`) | 오른쪽만 | `…RewrittenExecutionTermsUnderTheSameLineage/entry 만` |
| **stop(손절) 가격의 숫자** | 오른쪽만 | `…/stop 만` |
| target 가격의 숫자 | 오른쪽만 | `…/target 만` |
| **stop 가격의 출처 세 필드를 함께** 바꾼 경우 | 오른쪽만 | `…RefusesARestatedStopProvenanceAtTheSamePrice` — `digest` 하나만 바꾼 경우는 **이 시험이 못 잡는다**(9차 리뷰가 그 입력으로 주문을 냈다) |
| entry·target 가격의 출처, 주문 수량, 그리고 위 셋 중 **한 필드만** 바꾼 경우 | 오른쪽만 | **행동 시험으로는 덮이지 않음** — 아래 두 시험이 대신 덮는다 |

**행동 시험으로 축을 하나씩 세는 일은 끝나지 않는다.** `executionTermsIdentity` 가
담는 스칼라는 **32 개**이고(계좌·시장·종목·캠페인·leg·수량·정책·lineage 여덟 개에
가격 세 개 × 필드 여덟 개), 대부분은 시험 seam 이 하나만 움직이지도 못한다.
5~9 차가 라운드마다 하나씩 쪼갰고 매번 다음 축이 남았다.

의무를 둘로 나누면 각각 끝난다. 아래 둘이 위 표의 "덮이지 않음"을 덮는다.

| 무엇을 증명하나 | 어디서 | 어떻게 끝나나 |
|---|---|---|
| identity 가 **모든 필드**를 담는다 | `internal/strategyflow/execution_terms_identity_fields_test.go` | 32 개를 하나씩 바꿔 해시가 달라지는지 본다. 개수는 `reflect` 로 타입에서 읽으므로 필드가 늘면 시험이 터진다 |
| 이 가드가 **그 identity 를 그대로** 비교한다 | `internal/app/engine/strategy_first_leg_backstop_shape_test.go` | 단언 하나. `:223` 의 조건이 정확히 두 identity 비교의 논리합인지 AST 로 본다 |

7차 적대 리뷰가 이 표의 앞 판본을 반증했다. "수량과 세 가격을 그 시험이 덮는다"고
적었지만, 그 시험은 stop 과 target 을 **함께** 옮겼으므로 둘 중 하나만 보는 술어에도
걸렸다. 그리고 6.2 가 실제로 할 편집은 논리항 삭제가 아니라 **비교를 필드별로 펼치며
하나를 빠뜨리는 것**이다. 아래 셋이 그 모양을 잰 값이고, 수량은 덮이지 않았다고
적는다 — 침묵한 생략은 금지다.

**필드 하나씩 빠뜨린 확장 셋**(6.2 가 할 편집의 모양)도 같은 방식으로 쟀다.
술어를 `Lineage.Identity || Entry || EffectiveStop || Target || Quantity` 로 펼친 뒤
가격 하나씩을 뺀다.

| 뺀 필드 | 결과 |
|---|---|
| `Entry()` | RED — `…/entry 만` **만** 실패 |
| `EffectiveStop()` | RED — `…/stop 만` **만** 실패 |
| `Target()` | RED — `…/target 만` **만** 실패 |

8차 적대 리뷰가 그 셋도 부족함을 보였다. `PriceProvenance` 는 여덟 필드이고
`executionTermsIdentity` 는 여덟을 다 해시하는데, 시험 seam 이 숫자 말고는 못
움직여서 **비교를 세 숫자로 좁혀도 여섯 하위 시험이 전부 초록**이었다.

| 뮤테이션 | 결과 |
|---|---|
| 세 가격의 `priceMinor` 만 비교 | RED — `…RefusesARestatedStopProvenanceAtTheSamePrice` **만** 실패 |
| 위에 `EffectiveStop().Source()` 를 더한 것 | 행동 시험 **넷(하위 여섯) 전부 초록**(9차) — `stop.digest` 만 바꾼 제안으로 주문이 나갔다 |

**구조 단언은 그 여섯을 전부 잡는다.** 소스를 실제로 바꿔 재야 한다 —
`parser.ParseFile` 은 디스크를 읽으므로 `go test -overlay` 가 안 보인다.
백업 sha256 `1d710710…` 로 복원을 확인하며 하나씩 넣었다.

| `:223` 을 바꾼 것 | `…ComparesTheSealedIdentitiesThemselves` |
|---|---|
| `Lineage.Market` 만 | RED |
| `Lineage.Identity` 만 | RED |
| `Lineage.CampaignID` 만 | RED |
| `ExecutionTerms.Identity()` 만 (**행동상 동등 변이**) | RED |
| 세 `priceMinor` 만 | RED |
| 거기에 `EffectiveStop().Source()` 를 더한 것 | RED |

넷째 줄이 이 단언이 있어야 하는 이유다. 행동으로는 어떤 시험도 그것을 구별할 수
없다(포섭된 항을 지운 것이라 거동이 같다). 구조는 구별한다.

**구조 단언이 못 잡던 것도 적는다.** 10차 적대 리뷰가 조건을 한 글자도 안 바꾸고
**본문 안에** 분기를 넣었다 — ceiling 이 다르면 건너온 값을 그대로 믿고, 아니면
거절. 앞 판본은 본문에서 거절 문구가 **있는지**만 봤으므로 초록이었고, 위조된
수량 9 · 손절 80 짜리 1차 진입이 나갔다(fixture 가 전부 ceiling 8 이라 행동 시험도
그 길을 안 밟았다). 조건은 구조로 보면서 본문은 존재로 본 셈이다.

| 가드 주변을 바꾼 것 | 구조 단언 |
|---|---|
| 거절 **앞에** 분기를 넣어 건너온 값을 믿는 길을 만든다(본문 안) | RED — `본문이 문장 N 개다` |
| 같은 분기를 가드 **한 줄 위로** 옮긴다(본문 밖) | RED — `가드 바로 앞 문장이 …가 아니다` |
| 가드를 바깥 `if` 로 감싼다 | RED — `가드가 함수 최상위 문장이 아니다` |
| 가드 **뒤에서** `result` 를 다시 대입한다 | RED — `result 가 함수 안에서 2 번 대입된다` |

**앞 판본이 여기서 틀린 말을 적었다.** "가드 앞은 다른 가드(B2·B4)와 행동 시험이
맡는다"고 적었는데, 11차 적대 리뷰가 재 보니 **아무도 안 맡고 있었다** — B2(`:217`
개수 관문)도 B4(`:227` 위험 범위)도 행동 시험 넷도 그 편집에서 전부 초록이었고,
위조된 수량 9 · 손절 80 짜리 1차 진입이 나갔다. 구멍을 이름만 적었고 주인은 틀리게
적은 것이다. 이름을 적는 것과 주인을 확인하는 것은 다른 일이다.

이제 **재유도 사슬 전체**를 요구한다. 사슬은 세 고리다.

	proposal, ... := loader.proposals.forMarket(market)   // 뿌리
	proposalAuthority := proposal.entries[0].authority    // 고리 1
	result := proposalAuthority.Proposal()                // 고리 2
	<가드>

본문은 문장 하나(반환 둘, 그 문구의 `errors.New`), 가드는 최상위 문장이고 위 두
고리가 **바로 앞에 이 순서로** 있어야 하며, `proposal`·`proposalAuthority`·`result`
셋이 각각 **한 번만** 대입돼야 한다(`&x` 포인터 넘김도 대입으로 센다).

**고리 1 만 안 잡혀 있을 때를 12차 적대 리뷰가 보였다.** 6.2 가 실제로 쓸 편집
— 네 가족 중 조정자가 고른 항목을 loop 로 찾기 — 을 넣으면 고리 2 와 가드는 한
바이트도 안 바뀌고 모든 시험이 초록이다. 그러면 가드가 **자기 참조**가 된다:
accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와 비교한다.

오늘 그것이 해가 없는 이유는 `:217` 이 항목을 하나로 묶어 두기 때문인데,
**6.2 가 바로 그것을 바꾼다.** 그래서 지금 닫는다.

| 사슬을 바꾼 것 | 구조 단언 |
|---|---|
| 고리 1 뒤에 loop 를 넣어 `proposalAuthority` 를 다시 고른다(6.2 의 모양) | RED — `사슬의 고리 1 이 … 가 아니다` (대입 수로도 걸린다) |
| `proposal` 을 다시 대입한다 | RED — `proposal 가 2 번 대입된다` |
| 가드 뒤에서 `&result` 를 넘긴다 | RED — `result 가 2 번 대입된다` |

**감싸기가 우연히 막히던 것도 적는다.** `refusesWith` 가 재귀하므로 감싼 `if` 도
후보로 잡혀 "가드 2 개"로 터진다. 그것은 의도한 방어가 아니었다 — `refusesWith` 를
직속 문장만 보도록 "정리"하면 사라진다. 그 상태를 만들어 재 보니 최상위 문장 요구가
**독립적으로** 잡았다.

그 축을 밟으려고 `strategyflow.ResultWithRestatedStopProvenanceForTest` 를 새로 뒀다
(`tossos_testseams` 태그, 새 파일). 숫자는 그대로 두고 손절 출처 셋만 다시 적는다.

앞 판본(stop·target 동시 변경)에서는 이 셋이 **전부 초록**이었고, 각각 그 가격만
바꾼 제안으로 1차 진입이 나갔다.

GREEN 은 무뮤테이션 실행이다: 네 시험(하위 셋 포함) 모두 통과(`ok … 1.352s`).

