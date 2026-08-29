# issues — retire-gainers-source

이 change가 범위 밖으로 남기는 관찰.

**두 종류가 섞여 있다.** I1~I4는 설계에서 이미 범위 밖으로 정한 것이고 후속 change의
**시점과 등급을 함께 적는다.** I5~I11은 리뷰가 찾아 옮긴 것이고, 그중 일부는 후속 change가
아니라 **경계의 기록**이다 — 고칠 것이 없고 다음 사람이 알아야 할 사실만 있다.
그런 항목에 시점과 등급을 붙이면 있지도 않은 계획을 적는 것이 되므로 붙이지 않는다.

> **정정 2026-07-29 (2차 독립 리뷰 P2-4).** 초판의 위 문장은 *"각 항목은 후속 change의
> 시점과 등급을 함께 적는다"*로 예외 없이 쓰여 있었고, I5·I6·I8이 그것을 지키지 않았다.
> 규칙을 어긴 것이 아니라 **규칙이 과했다** — 파일이 자기 규칙을 위반하는 상태를 두면,
> 다음 사람은 빠진 것을 채우거나 규칙을 무시하는 것 중 하나를 하게 되고 둘 다 나쁘다.

## I1. 되찾을 수 없는 4xx는 후퇴 사다리에 걸리지 않는다 (D4)

`Backoff.Note`는 `missing.RateLimited`일 때만 불린다([watch.go:653] 근방). 400은 429가
아니므로 어떤 4xx도 사다리를 움직이지 않고, `retire-gainers-source` 이전의 gainers 원천은
그래서 영원히 15초마다 재시도되고 있었다.

**고치지 않은 이유는 두 가지이고 둘 다 이 change의 형태 때문이다.**

1. `Missing`이 지닌 분류 비트는 정확히 하나, `RateLimited`뿐이다 —
   [scan.go:59-62]가 *"RateLimited separates a 429 from every other failure"*라고 쓰고 있다.
   그러므로 "되찾을 수 없는 4xx도 백오프시킨다"는 조건 하나를 바꾸는 일이 아니라 **오류를
   새로 분류하는 일**이고, 어떤 오류를 되찾을 수 없다고 부를지는 API에 대한 정책 판단이다.
   틀리면 회복 가능한 원천이 영구 후퇴에 갇히고, 그 실패는 이 change가 고친 실패와 같은
   종류다 — 아무것도 실패하지 않으면서 원천이 조용히 사라진다.
2. `off.Note` 호출은 `Cycle` 안에 있고, 그것이 움직이는 것은 429 후퇴 사다리다.
   `docs/WORKFLOW.md`가 **retry matrix·rate limit**을 High-risk 경로로 명시한다.

**후속 change의 등급: High-risk.** 적대적 Eng 리뷰와 Pre-Edit 선언이 필요하다.

**시점**: 지금 패널에는 사례가 없다. 이 change가 유일한 사례를 내렸기 때문이다. 되찾을 수
없는 4xx를 반복하는 원천이 다시 생기면 그때가 그 change의 시점이다.

## I2. 원천 **개수**를 주장하는 주석에는 가드가 없다

`panelsize_drift_test.go`는 `숫자 + rows/행/개` 형태의 **행 수** 주장만 잡는다. `"four
sources"`는 숫자가 단어이고 단위가 rows가 아니라 잡히지 않으며, 그 파일의
`TestWhatThisGuardCannotCatch`가 이미 *"a number spelled out in words has no digits to
match"*를 첫 항목으로 적어 두었다.

이 change가 그 표류를 실제로 밟았다: `clock_wiring_test.go`의 `want 4`와 두 곳의 *"Panel
builds four of them"*, `notdue_test.go`의 *"exactly three official sources"* 두 곳,
`candidatesrc.go`의 *"the three types are compile-time constants"*가 전부 원천 하나를 뺀
순간 거짓이 됐고, **그 중 컴파일러나 테스트가 잡은 것은 `want 4` 하나뿐**이다.

**이 change에서는 만들지 않는다**(범위 밖). 대신 고치는 방식으로 대응했다 — 개수를 셋으로
줄이는 대신 **개수를 말하지 않는 문장**으로 바꿨다. 다음 원천 변경에서 다시 표류하지 않는
것은 가드가 아니라 문장의 성질이다.

가드를 만든다면 형태는 `rowClaim`과 다르다. 행 수는 `declaredPanelSizes`가 코드에서 읽을 수
있는 값이 있지만, 원천 개수는 시장과 WTS 세션 유무에 따라 달라져서 대조할 단일 값이 없다.
`Panel`을 두 시장 × WTS 유무로 호출해 얻은 집합의 크기들과 대조하는 형태가 되어야 한다.

## I3. `1d` gainers가 장중에 갱신되는지 확인하지 못했다

proposal의 재검토 조건 3번이다. 2026-07-28 실측은 마감 후 관측이라 세 랭킹의 `ranked_at`이
모두 19:59로 같았고, 그것은 "1d가 일 배치"의 증거도 "장중 갱신"의 증거도 되지 못한다.

재검토 시 **장중에** 확인해야 한다. 이 사실을 모르는 채로 `1d`를 넣으면 어떤 창에서 잰
순위인지 모르는 값이 백분위에 들어간다.

## I4. `candidatesrc_test.go:290`의 "three official rankings"는 고치지 않았다

grep이 잡았지만 그 문장은 **과거 결함의 기록**이다 — *"The three official rankings shipped
under one source id"*는 §2 리뷰가 고친 id 충돌에 대한 서술이고, 그때 세 랭킹이 하나의 id를
공유했던 것은 지금도 사실이다. `rankingSourceID`는 D2에 따라 세 항목을 그대로 유지하므로
"세 랭킹"이라는 표현은 상수 쪽에서 여전히 정확하다.

패널 구성을 말하는 문장이 아니므로 표류가 아니라고 판단했다. 판단이 틀렸다면 고칠 곳은
그 한 줄이다.

---

**독립 리뷰(2026-07-29)가 남긴 비차단 발견 — §8.5.**
아래 I5~I10은 `review.md`의 "독립 리뷰 (§7.4)" 절에서 리뷰어가 F4~F9로 낸 것이다.
랜딩을 막지 않으며, P1 세 건(F1·F2·F3)과 달리 이 change에서 닫지 않는다.
**단 F4는 §8.1이 절반을 닫았고, 그 경계를 아래에 실측으로 적는다.**

## I5. `wiredRankings`의 리터럴 인식 — 절반은 §8.1이 닫았고 절반은 남았다 (F4)

`snapshot_drift_test.go`의 `wiredRankings`는 `Panel` 본문의 **모든** `[]string` 합성
리터럴을 합집합으로 모은다. 리뷰는 그 형태가 두 방향으로 잘못될 수 있다고 지적했다.

**① 눈이 머는 방향 — 닫혔다.** 랭킹 목록이 `Panel` 밖(패키지 수준 `var` 등)으로 옮겨지고
`Panel` 안에 다른 `[]string` 리터럴이 남으면, `len(out) == 0` Fatal이 무력화되고 스냅샷
가드는 **보지 못하는 배선에 대해 계속 초록**이 된다.

§8.1의 `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`가 이것을 닫는다 —
AST가 읽은 타입 집합과 `Panel`을 실제로 호출해 얻은 원천의 타입 집합이 다르면 실패한다.
**실측(2026-07-29)**: 목록을 패키지 수준 `var`로 올리고 `Panel`에 무관한 `[]string`
리터럴을 남긴 변이에서

```text
--- FAIL: TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
--- PASS: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
```

스냅샷 가드가 **통과하는 바로 그 상태**에서 새 가드가 실패한다. 그것이 F4가 말한 눈멂이다.

**② 시끄러운 방향 — 남았다.** `Panel`에 `[]string{candidate.MarketKR}` 같은 리터럴이
생기면 원소가 `*ast.SelectorExpr`이라 `wiredRankings`가 *"Panel builds a ranking from
\*ast.SelectorExpr"*로 Fatal한다 — **랭킹과 무관한 이유로 죽는다.**
**실측(2026-07-29)**: 그 리터럴을 넣자 스냅샷 가드와 새 가드가 **둘 다** 그 메시지로 RED.
새 가드도 `wiredRankings`를 쓰므로 같은 함정을 **물려받았다.**

안전한 방향의 실패(거짓 통과가 아니라 거짓 실패)라 랜딩을 막지 않는다. 그러나 다음 사람이
`Panel`에 무관한 문자열 슬라이스를 넣는 순간 두 테스트가 이해할 수 없는 이유로 깨지고,
그때 그것을 느슨하게 만드는 수리가 ①을 되살릴 수 있다.

**고친다면 형태**: 리터럴을 찾는 대신 `range` 대상 식을 따라가거나, `Panel` 안의 `[]string`
중 **`range` 절에 있는 것**만 본다. 후자가 이 파일의 의도에 가깝다 — 감시 대상은 "Panel이
순회하는 목록"이지 "Panel 안의 모든 문자열 슬라이스"가 아니다.

## I6. 가드는 타입×duration을 교차곱으로 짝짓는다 — 재검토 시점에 오탐이 된다 (F5)

`wiredRankings`는 `Panel`에서, `wiredDurations`는 **파일 전체의 모든 `.Rankings(` 호출**에서
읽고, 단언은 두 집합의 **곱** 위에서 돈다. "이 타입은 이 duration으로 부른다"는 짝이
표현되지 않는다.

지금은 랭킹 reader가 하나이고 duration 리터럴도 하나라 곱이 곧 실제 짝이다. 그러나
proposal의 **재검토 조건 4번**이 충족되어 gainers가 `1d`로, 나머지가 `realtime`으로
돌아오는 날 이 가드는 **위반이 아닌 것을 위반으로 보고한다.**

이것을 적어 두는 이유는 그때 이 가드를 느슨하게 만들 사람이 이 문서를 읽지 않은 사람이기
때문이다. **근거가 눈앞에 있는 지금 적는다.** 이 change에서 고치지 않는다(범위 밖) —
지금 고치면 존재하지 않는 배선에 맞춘 추상이 된다.

## I7. 가드가 보는 파일은 하나, 랭킹 요청을 만드는 곳은 둘 (F6)

`cmd/tossctl/market.go:332, 339-341`도 `client.Rankings(...)`를 부르고 `--type` 도움말이
`TOP_GAINERS|TOP_LOSERS`를 광고한다. **위반은 아니다** — `--duration` 기본값이 `1d`라
기본 조합이 합법이고, 타입·duration이 둘 다 런타임 플래그라서 **정적으로 대조할 짝이 없다.**

기록하는 이유는 하나다: 다음 사람이 `snapshot_drift_test.go`를 보고 "이 가드가 저장소
전체를 본다"고 오해하지 않도록. 가드가 보는 것은 `internal/candidatesrc/candidatesrc.go`
하나다. 사용자가 직접 고른 타입·duration 조합에 대한 방어는 정적 가드가 아니라 API의 400
응답이고, 그것으로 충분하다는 것이 이 항목의 판단이다.

## I8. D2의 세 근거 중 하나는 코드가 강제하지 않는다 — **이 change에서 세 번째** (F7)

`internal/candidate/candidate.go:95-97`과 `retiredsource_test.go:95-97`이 같은 근거를 쓴다:
*"관측은 시스템이 알아보는 것으로 되읽혀야 한다."*

**강제되지 않는다.** `decodeSources`([store.go:2077-2089])는 임의 문자열을 `SourceID(p)`로
만들 뿐 아무것도 검증하지 않는다. 상수를 지워도 저장된 행이 읽히지 않게 되는 일은 없다.
게다가 §5.1 측정대로 그런 행은 0건이다.

**결론(D2 — 상수를 남긴다)은 맞다.** 나머지 두 근거가 독립적으로 성립한다: 되돌림이 한
편집이어야 한다는 것, 논거가 주석에 산다는 것. 근거 하나가 빠져도 결정은 서 있다.

**그런데 이것이 이 change에서 세 번째다.**

| # | 어디 | 틀린 근거 | 결론 |
|---|---|---|---|
| 1 | proposal | `NoteSources` | 맞았다 |
| 2 | tasks §3.4 | *"호출이 세 번에서 두 번으로"* — 소스에 세 번이었던 적이 없다 | 맞았다 |
| 3 | D2 / I8 | *"되읽혀야 한다"* — `decodeSources`가 검증하지 않는다 | 맞았다 |

세 번 모두 **결론은 맞고 근거는 틀렸다.** 이 형태가 위험한 것은 결론이 틀려서가 아니라,
**다음 사람이 그 근거를 재사용하기 때문**이다. §8.1이 닫은 F1도 같은 뿌리에서 나왔다 —
거기서는 근거가 "이 테스트가 잡는다"였고, 그 테스트가 잡지 않았다.

**후속 change로 만든다면**: `decodeSources`가 알 수 없는 `SourceID`를 거절하거나 최소한
셈하게 한다. **등급 판단이 먼저 필요하다** — `store.go`의 읽기 경로이고, 거절로 바꾸면
과거 행이 읽히지 않게 되는 쪽의 위험이 생긴다. 셈하는 쪽(관측)이 먼저다.

## I9. `clock_wiring_test.go`가, 없어지기로 되어 있는 파일에 의존한다 (F8)

`sameSourceSet`·`sortedIDs`가 `retiredsource_test.go:36-56`에 있고 `clock_wiring_test.go`가
그것을 쓴다. `retiredsource_test.go`의 헤더는 그 파일이 *"pins the one thing
`retire-gainers-source` changes"*라고 선언한다 — gainers가 돌아오는 날 지워질 후보다.
지우면 시계 테스트가 **컴파일되지 않는다.**

컴파일 실패는 시끄러운 실패라 안전하다(조용히 통과하는 것과 반대다). 결합의 **방향**이
거꾸로일 뿐이다: 오래 사는 파일이 짧게 사는 파일에 의존한다.

헬퍼가 그 파일에 있는 이유는 기록되어 있다 — FLM 대상 범위를 좁게 유지하려고 새 파일에
넣었고, 그것은 tasks.md가 요구한 배치다. **이 change에서 옮기지 않는 이유도 같다**:
지금 옮기면 `clock_wiring_test.go`의 hunk가 다시 넓어져 FLM 대상이 늘어난다.
gainers 재검토 change나 그 다음 정리 change에서 공용 헬퍼 파일로 옮기는 것이 맞다.

## I10. §5.2는 냉각 불변식의 단독 가드가 아니다 (F9, 정보)

`coverageAnswered`에서 `if !heard[id] { continue }`를 제거하는 변이에서
`TestALiveSupporterStillCoolsACandidateARetiredSourceAlsoRaised`(§5.2)만 RED가 되는 것이
아니다. 기존 테스트 둘도 함께 RED다 —
`scan_test.go:687 TestASupporterThatLeftThePanelDoesNotBlockCoolingForever`,
`notdue_test.go:531 TestASourceTheSchedulePassedOverIsNotASourceThatIsGone`.

결함이 아니다. §5.2는 새 커버리지라기보다 **이 change의 안전 논증을 그 이름으로 고정한
기록**이다. 아무도 §5.2를 이 불변식의 유일한 그물로 세지 않도록 남긴다.

## I11. 새 가드는 두 시장의 합집합을 본다 — 시장별 배선은 그 눈에 안 보인다 (2차 리뷰 P2-3)

`TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`는 `declared`와 `built`를 둘 다
**KR·US 합집합**으로 만든다. 그 선택의 *이유*는 파일에 적혀 있다 — 시장으로 갈린 랭킹이
생겼을 때 한쪽 시장에서만 안 만들어졌다고 오탐하지 않기 위해서다.

적혀 있지 않던 것은 **무엇이 안 보이는가**이다. 2차 리뷰 실측: 랭킹 루프를 KR로 게이트해
US 패널을 랭킹 없이 비워도 이 가드는 **통과한다**(합집합에서는 여전히 declared == built).
그 변이를 잡는 것은 이 가드가 아니라 다른 네 가드다.

**지금 도달 불가능하다.** `OfficialRanking`이 market을 인자로 받지 않으므로 랭킹 원천을
시장별로 가르는 배선 자체가 없다. 그러나 그것은 `OfficialRanking`의 시그니처가 지키는
성질이지 이 가드가 지키는 성질이 아니고, 시그니처는 바뀔 수 있다.

**고칠 것 없음 — 경계의 기록이다.** 랭킹이 시장별로 갈리는 change가 오면 그때 이 가드를
시장별 비교로 바꿔야 하고, 이 항목이 그 change가 읽어야 할 문장이다.
