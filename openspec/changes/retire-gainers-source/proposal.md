# 답하지 못하는 원천을 패널에서 내린다

## 무엇이 문제인가

`official_rankings_top_gainers`는 **한 번도 응답한 적이 없다.**

`officialRanking.Read`는 세 랭킹 전부에 `duration=realtime`을 보내는데
([candidatesrc.go:247](../../../internal/candidatesrc/candidatesrc.go)),
`TOP_GAINERS`는 그 값을 받지 않는다. 2026-07-28 실측(KR·US 각 2회):

```
400 {"error":{"code":"unsupported-ranking-duration",
     "message":"지원하지 않는 랭킹 기간입니다.",
     "data":{"field":"duration","allowedValues":["1d","1w","1mo","3mo","6mo","1y"]}}}
```

`MARKET_TRADING_AMOUNT`와 `MARKET_TRADING_VOLUME`은 같은 엔드포인트에서 `realtime`을
정상 수락한다(각 100 요청 / 100 도착). 즉 이것은 엔드포인트 장애도 일시적 강등도 아니라
**이 랭킹 타입이 거절하는 파라미터를 계속 보내고 있는 배선 결함**이다.

## 왜 지금 고치는가 — 이것이 조용한 이유

아무것도 실패하지 않는다. 스캔은 성공하고, 후보는 나오고, 테스트 3676개는 초록이다.
대신 세 가지가 조용히 망가져 있다.

**1. `degraded`가 신호이기를 그만뒀다.** `result.Degraded = len(result.Missing) > 0`
([scan.go:433](../../../internal/candidate/scan.go))이므로 매 스캔 참이고, 그 값이 모든 후보
행에 찍힌다. 운영자는 "gainers가 늘 그렇듯 죽었다"와 "방금 trading_value가 실패했다"를
구분할 수 없다. **상시 점등인 경보는 경보가 아니다.**
스펙이 요구하는 "강등 사실을 후보와 스캔 결과 양쪽에 기록"은 형식적으로 지켜지면서
그 기록이 가리키는 것이 없어졌다.

**2. 후퇴 사다리가 이 실패를 잡지 못한다.** `Backoff.Note`는 `missing.RateLimited`일 때만
불린다([watch.go:651-654](../../../internal/candidate/watch.go)). 400은 429가 아니므로
**영원히 15초마다 재시도**한다([candidate.go:367,372](../../../cmd/tossctl/candidate.go)).
사다리는 되찾을 수 있는 실패를 위해 만들어졌고, 이것은 되찾을 수 없는 실패다.

**3. 되찾을 수 없는 호출이 rate 예산을 쓴다.** 안전 불변식 §0.4는 공식 API 호출을 예산에
계상하라고 요구하는데, RANKING 그룹 호출의 1/3이 성공할 수 없는 요청이다. 그 예산은
`add-candidate-discovery`가 "엔진 > 실계좌 검증 > 발굴" 우선순위로 아끼기로 한 바로 그 예산이다.

## 이 결함은 저장소 안에 이미 문서화되어 있었다

`docs/migration/openapi.latest.json`은 공식 스냅샷이고, 그 안에 이렇게 적혀 있다:

> `TOP_GAINERS` / `TOP_LOSERS` 는 `duration=realtime` 을 지원하지 않습니다
> (400 `unsupported-ranking-duration`).

이 파일은 **2026-06-12**에 들어왔고(`ba2cf1a`), `realtime`을 세 랭킹 전부에 보내는 배선은
**2026-07-28**에 들어왔다(`5c1ced7`). **46일 먼저 있었다.**

그러므로 근본 원인은 "API가 우리를 놀라게 했다"가 아니다. `docs/WORKFLOW.md`의 권위 경계표에서
**"브로커 실제 동작 = 공식 API 응답 fixture"** 계층이 저장소 안에 있었고, 참조되지 않았다.
`realtime`을 고른 근거는 추론이었고, 추론이 fixture를 이겼다.

이 사실이 §2의 표류 테스트를 요구사항이 아니라 **필수**로 만든다 — 사람이 스냅샷을 읽기로
결심하는 것에 의존하면 같은 일이 다시 일어난다.

## 망가지지 않은 것 — 그러나 처음 쓴 이유는 틀렸다

**감시목록은 안전하다.** 그런데 이 문서의 초안이 댄 근거는 코드와 다르다. 리뷰가 잡았고,
확인했다.

`NoteSources`는 **덮어쓰지 않고 합친다.** SQL 문자열은 `SET sources = ?`지만 바인딩되는 값이
합집합이다([store.go:1232-1234](../../../internal/candidate/store.go)):

```go
encodeSources(append(append([]SourceID{}, existing.Sources...), sources...))
```

함수 위 주석이 그렇게 말하고 있다 — *"`sources`는 후보가 이미 지닌 것에 **합쳐지지**
대체되지 않는다"*. 그러니 "죽은 원천은 어느 후보의 `sources`에도 들어가지 않는다"는
`NoteSources`의 성질이 아니다.

안전한 진짜 이유는 **둘**이고, 서로 다른 불변식이다.

1. **이 원천은 한 번도 `raisedBy`에 들어간 적이 없다.** 응답한 적이 없어서다 —
   `scan.go:417-419`는 오류 분기를 통과한 원천의 행에 대해서만 append한다. 이것은
   측정된 사실이다(2026-07-28 스캔: gainers가 올린 후보 0개).
2. **누적된 죽은 원천은, 후보에게 살아 있는 supporter가 하나라도 있으면 냉각을 막지
   않는다.** `heard`는 패널과 `NotAsked`로만 만들어지고 일정 map에서는 오지 않는다
   ([scan.go:261-271](../../../internal/candidate/scan.go)). `coverageAnswered`는 `heard`에
   없는 supporter를 건너뛴다(`scan.go:704-716`). `scan.go:190-199`가 이것을 "it is gone"
   경우로 명시하고 있고, 운영자가 WTS를 빼는 상황을 위해 쓰였다.

**두 번째 불변식에는 조건이 붙는다. 초안은 그 조건을 빼고 썼다.**

`coverageAnswered`는 `heard`에 있는 supporter의 수를 세고 마지막에 `present > 0`을
요구한다. 죽은 원천이 **유일한** supporter인 후보는 `present == 0`이 되어 false가 돌아오고,
`coolAbsent`는 `continue`로 그 후보를 건너뛴다([scan.go:685-687]). 즉 그런 후보는
**이 경로로는 영영 냉각되지 않는다.**

그러므로 두 불변식은 독립적인 두 근거가 아니라 **함께여야 성립하는 하나의 근거**다.
불변식 2가 안전한 것은 "유일한 supporter가 gainers인 후보"가 존재하지 않기 때문이고,
그것을 보장하는 것이 불변식 1이다. 이 change는 그 사실을 **논증이 아니라 측정으로**
확인한다(§5.1) — 저장소에 그런 후보가 한 건이라도 있으면 이 change는 냉각 불가능한
후보를 만들게 되고, 그때는 다른 형태가 필요하다.

**테스트가 이것을 고정해야 한다**(§5.2). 초안은 결론이 맞았다는 이유로 근거를 검증하지
않았고, 검증했더니 근거의 절반이 조건부였다.

이 확인이 이 change의 범위를 정한다 — **긴급 복구가 아니라 정리**다.

## 무엇을 하는가

패널과 일정에서 `SourceOfficialGainers`를 내린다. 그것뿐이다.

- `candidatesrc.Panel`의 랭킹 타입 리터럴에서 `RankingTopGainers`를 뺀다
- `candidateIntervals()`에서 해당 항목을 뺀다
- `SourceOfficialGainers` 상수, `rankingSourceID` 항목, `RankingTopGainers` 상수,
  그리고 "TOP_GAINERS는 정의상 이미 올라간 것들의 목록"이라는
  [candidate.go:10-12](../../../internal/candidate/candidate.go)의 논거는 **전부 남긴다**
- 왜 내렸는지, 무엇이 갖춰지면 다시 볼지를 코드 주석과 이 change에 기록한다

되돌리는 비용은 슬라이스 원소 하나다. 그것이 이 형태를 고른 이유다.

## 무엇을 하지 않는가 — `duration=1d`로 바꾸지 않는다

`1d`는 동작한다(실측 확인, 100행). 그런데도 지금 넣지 않는다. 세 가지가 각각 독립적으로 막는다.

**① 지금 넣으면 걸러지지 않는다.** 2026-07-28 KR 실측:

| | |
|---|---|
| `1d` gainers 100행 중 기존 패널(거래대금∪거래량 131종목)에 없는 종목 | **96** |
| 그 목록의 **최하위** 종목이 오늘 이미 오른 폭 | **+5.26%** |
| 상위 10개 중 상한가(+29.9% 이상) | **9** |

후보군이 131 → 227로 늘고, 늘어난 96종목은 전부 이미 5% 이상 오른 것들이다. 그런데
그것을 걸러야 할 `seen_late`·`extended`는 지금 **둘 다 임계가 없다**(같은 스캔에서
`THRESHOLD_ABSENT 312`). 필터가 꺼진 상태에서 추격 목록을 먼저 넣는 순서다.

**② 싼 형태가 없다.** 두 갈래뿐이고 둘 다 다른 방식으로 나쁘다.

- **1d 순위가 다른 둘과 같은 백분위에 들어간다** — realtime 창에서 잰 순위와 1d 창에서 잰
  순위가 한 백분위에 섞인다. `fix-chase-veto-measurement`가 `seen_late`를 **정의된 것**으로
  만들려고 한 change 전체를 쓴 직후에, 그 정의를 다시 흐린다.
- **1d 순위를 first_rank에서 제외한다** — `qualifiesFirstRank`
  ([scan.go:639](../../../internal/candidate/scan.go))가 막으면, gainers만 올린 96종목은
  `first_rank`를 영원히 갖지 못하고 `seen_late = NO_FIRST_RANK`가 영구화된다. D10의
  "측정하지 못한 veto는 통과가 아니다"가 감시목록 **과반**에 적용된다. 측정 불가능한 후보를
  스캔마다 96개씩 제조하는 셈이다.

**③ 비승격 원천이라는 개념이 없다.** `Collect`는 응답한 모든 원천의 **모든 행**을 `raisedBy`에
넣고([scan.go:403-430](../../../internal/candidate/scan.go)), `raisedBy`의 모든 심볼을
승격한다([scan.go:465-478](../../../internal/candidate/scan.go)). "읽되 후보로 올리지 않는다"는
설정이 아니라 새 기계장치다.

**미측정으로 남기는 것**: `1d`가 장중에 갱신되는지 일 배치인지 구분하지 못했다. 마감 후
관측이라 세 랭킹의 `ranked_at`이 모두 19:59로 같아 판별 근거가 되지 못한다. 재검토 시
장중에 확인해야 한다.

## 재검토 조건 — 이 change가 남기는 것

`1d` gainers를 다시 보는 것은 다음이 **전부** 성립할 때다. 이 문장이 change의 산출물이다.

1. `seen_late`에 실측 기반 승인 임계가 있다 — 그것이 이 목록을 거를 장치다
2. `extended`의 그림자 밴드가 이 목록이 사는 구간(0~10%)을 분해한다
3. 장중 관측으로 `1d`의 갱신 주기를 확인했다
4. 다른 창에서 잰 순위를 백분위에 섞지 않는 방법이 설계되어 있다

## 위험

**등급: Normal.** 주문·손절·사이징·원장·인증·체결에 닿지 않는다. 발굴 패키지는 주문 경로에
도달할 수 없음이 테스트로 강제되어 있다(`isolation_test.go`).

| 위험 | 완화 |
|---|---|
| 원천이 줄어 후보를 놓친다 | 이 원천이 지금 올리는 후보는 **0개**다. 응답한 적이 없다. 손실의 크기가 측정되어 있다 |
| `Panel`이 비는 구성이 생긴다 | KR 3원천·US 2원천이 남는다. "공식 원천만으로 후보가 산출되어야 한다"는 계속 성립 |
| 나중에 되돌리기 어렵다 | 상수·id 매핑·논거를 전부 남긴다. 되돌림은 슬라이스 원소 하나 |
| 같은 결함이 다른 원천에서 재발한다 | spec delta가 요구사항 수준으로 금지하고, 테스트가 배선을 고정한다 |
