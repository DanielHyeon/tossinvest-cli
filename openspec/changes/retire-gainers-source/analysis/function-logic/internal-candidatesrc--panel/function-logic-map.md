# Function Logic Map: `Panel`

- Source: `internal/candidatesrc/candidatesrc.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> 이 change가 `Panel`에서 바꾼 것은 **리터럴 슬라이스의 원소 하나**와 주석이다. 분기·
> early return·mutation·fallback은 하나도 바뀌지 않았고, AST의 분기 목록(B1~B4)은 base와
> 동일하다. 바뀐 것은 B2가 순회하는 값 집합이 `{MARKET_TRADING_AMOUNT,
> MARKET_TRADING_VOLUME, TOP_GAINERS}`에서 `{MARKET_TRADING_AMOUNT,
> MARKET_TRADING_VOLUME}`으로 준 것뿐이다.

> **정정 2026-07-29 (§8.1·8.2, 독립 리뷰 F1).** 이 문서의 초판은 B3(버려지는 오류)를
> `TestEveryPanelSourceHasItsOwnID`가 덮는다고 적었다. **거짓이었다.** 그 테스트는 넘겨받은
> 패널의 id 중복과 비어있음만 보고, 오류로 조용히 빠진 원천은 "원소가 하나 적고 id는 여전히
> 유일하며 비어 있지 않은" 패널을 만든다. 재현: `rankingSourceID`에 항목이 없는 실제 enum 값
> (`TOSS_SECURITIES_TRADING_AMOUNT`)을 리터럴에 넣었더니 패키지 53건 전부 통과했다.
> §8.1이 `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`를 만들어 그 상태를
> 실패로 만들었고(같은 변이에서 RED 관측), 아래 표의 "Required test" 칸과
> `candidatesrc.go`의 주석을 사실에 맞게 고쳤다. 이 문서의 줄 번호는 그 주석 편집으로
> 이동했으며 `ast.json`은 재생성했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market string` | 임의 문자열. `strings.ToUpper(strings.TrimSpace(...))`로 정규화된 뒤 `candidate.MarketKR`과만 비교된다 | `internal/candidate`의 `MarketKR`/`MarketUS` 상수 | 알 수 없는 시장은 오류가 아니다. WTS 원천이 빠진 패널이 나온다 — 시장을 볼 수 없는 원천이 present-and-empty로 있으면 스캔이 그것을 "심볼이 목록에서 빠졌다"는 증거로 읽고 후보를 냉각시킨다 |
| `official RankingReader` | nil 가능 | 호출부 `buildCandidatePanel` — 공식 자격증명이 없으면 그쪽이 먼저 실패한다 | nil이면 랭킹 원천이 하나도 만들어지지 않는다. `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`가 그 결과를 고정한다 |
| `budget BudgetReader` | nil 가능 (선택적 accessor) | `official/ratebudget.go` | nil이면 `officialRanking.rateBudget`이 `Reported=false`를 낸다. 0이 아니라 미보고다 |
| `wts PopularityReader` | nil 가능. typed-nil은 호출부 `wtsPopularityReader`가 막는다 | WTS 세션 존재 여부 | nil이면 WTS 원천 없음. 이것은 정상 경로이고 요구사항이 그렇게 정한다 — 공식 원천만으로 후보가 나와야 한다 |
| `clk clock.Clock` | nil 가능 | 호출부가 `clock.System()`을 넘긴다 | nil은 각 생성자 안에서 `clock.System()`으로 대체된다. `TestThePanelHandsItsClockToEverySourceItBuilds`가 **모든** 생성자에 clk가 전달되는지 본다 |
| 랭킹 타입 리터럴 (B2의 순회 대상) | `rankingSourceID`에 항목이 있는 상수만 | 이 파일의 const 블록 + `rankingSourceID` | 매핑이 없는 타입은 `OfficialRanking`이 오류를 내고 B3가 그 원천을 **조용히** 버린다 — 패널은 원소가 하나 줄 뿐이다. 그 상태를 실패로 만드는 것은 `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`(리터럴의 타입 집합 == `Panel`이 실제로 만든 원천의 타입 집합) 하나뿐이다. `TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused`는 같은 리터럴을 읽지만 스냅샷 금지 조합만 본다 |

**이 change가 건드린 불변식**: 없음. 리터럴의 원소가 줄어 순회 횟수만 준다. 호출량은 단조
감소이며(D10), 새 엔드포인트·새 파라미터·새 원천은 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `official != nil` (line 555) | 없음 — 아래 블록의 실행 여부만 정한다 | 없음 (early return 없음) | `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly` |
| B2 | `range []string{RankingTradingAmount, RankingTradingVolume}` (line 604) | 순회당 `sources`에 최대 하나 append | 없음 | `TestNoMarketPanelBuildsTheGainersRanking`(원소 집합), `TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused`(원소 값 대 스냅샷), `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`(리터럴 집합 == 실제 생성된 원천 집합), `TestEveryPanelSourceHasItsOwnID`(id 유일성) |
| B3 | `err == nil` (line 605) | `sources = append(sources, src)` | 없음. 오류는 버려진다 — 리터럴의 모든 값이 컴파일 타임 상수이고 `rankingSourceID`에 있으므로 오류는 이 파일의 결함이다 | `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds` — **이 분기를 덮는 유일한 테스트다.** `TestAnUnknownRankingTypeIsRefused`는 생성자를 직접 부르므로 `Panel` 경로에 대해 아무 말도 하지 않고, `TestEveryPanelSourceHasItsOwnID`는 넘겨받은 패널의 id 중복·비어있음만 보므로 **버려진 원천을 보지 못한다**(2026-07-29 독립 리뷰 F1이 재현) |
| B4 | `wts != nil && market == candidate.MarketKR` (line 612) | `sources = append(sources, WTSPopular(wts, 30, clk))` | 없음 | `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking`, `TestThePopularityRankingRefusesAMarketItCannotSee` |

**early return 없음.** 함수의 유일한 return은 line 615의 `return sources`다(ast.json `returns`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToUpper` / `strings.TrimSpace` (line 552) | 시장 문자열 정규화. B4의 비교가 이것에 의존한다 | 오류 없음 | ast.json `calls` |
| `OfficialRanking` (line 605) | 랭킹 타입 하나를 `candidate.Source`로 감싼다 | 알 수 없는 타입이면 오류를 돌려주고 B3가 버린다. I/O 없음, 타임아웃 없음, retry 없음 | ast.json `calls`; `candidatesrc.go` `OfficialRanking` |
| `append` (line 606, 613) | 패널 슬라이스 구성 | 오류 없음 | ast.json `calls` |
| `WTSPopular` (line 613) | WTS 인기 목록을 원천으로 감싼다 | 오류를 돌려주지 않는다. size≤0이면 기본값으로 대체한다 | ast.json `calls`; `candidatesrc.go` `WTSPopular` |

**live config binding**: 없다. 이 함수는 설정 파일도 환경변수도 읽지 않는다. 유일한 외부
결합은 인자로 들어온 reader 셋과 시계다. 폴링 간격은 이 함수가 아니라 `cmd/tossctl`의
`candidateIntervals`에 있고, 이 change는 **두 곳을 같이** 바꿨다 —
`candidateschedule_drift_test.go`가 그 결합을 이제 테스트로 고정한다.

## State mutations and fallbacks

- mutation은 지역 변수 `sources`의 append 두 곳뿐이다(ast.json의 `mutations`는 null이고,
  append는 재대입으로만 나타난다). 패키지 전역 상태·파일·네트워크에 아무 것도 쓰지 않는다.
- fallback은 하나이고 이 함수 밖에 있다: 각 생성자 안의 `clk == nil → clock.System()`.
  `Panel` 자신은 fallback을 갖지 않으며 이 change는 그 경로를 건드리지 않았다.
- **이 change가 만든 fallback은 없다.** 원천을 뺀 자리에 대체 원천을 넣지 않았다.
  남는 구성은 공식 랭킹 둘에 **WTS 인기 목록이 KR에서만, 그것도 세션이 있을 때만** 더해진
  것이다 — `Panel`이 `wts != nil && market == KR`에서만 그것을 붙이기 때문이다.
  "공식 원천만으로 후보가 산출된다"는 요구사항은 계속 성립한다.

  > **정정 2026-07-29 (2차 독립 리뷰 P2-1).** 초판은 *"남는 구성은 KR 세 원천·US 두 원천"*
  > 이라고 **무조건으로** 적었다. KR이 셋인 것은 **WTS 세션이 있을 때뿐**이고, 없으면 둘이다.
  > tasks §7.3이 실측 판정에서 *"숫자를 고정해서 읽지 않는다"*고 경고한 바로 그 형태가
  > 게이트 산출물 안에 남아 있었다 — issues.md I2가 "원천 개수 주장에는 가드가 없다"고
  > 적은 것의 실례다.

## Safety conclusion

- Safe edit boundary: 리터럴 슬라이스의 원소 집합. 분기 구조·return·mutation은 base와
  동일하며 AST 분기 목록(B1~B4)도 동일하다.
- High-risk impact: **no.** 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어느
  경로에도 닿지 않는다. `internal/candidate`의 `isolation_test.go`가 발굴 패키지의 의존
  폐포를 강제하고, 이 change는 그 폐포를 넓히지 않는다.
- 호출량: 순회당 RANKING 그룹 호출이 **줄기만 한다**. 안전 불변식 §0.4(rate 예산 계상)에
  대해 이 편집은 언제나 보수 방향이다.
