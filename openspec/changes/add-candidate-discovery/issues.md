# Issues: add-candidate-discovery

`docs/WORKFLOW.md` §예외 경로의 기록 파일. 스펙 결함과 구현 중 이탈을 분류해 남긴다.

## 1. `first_rank_at` — 설계가 명시한 3칼럼에 하나를 더했다 (safe local, 4.9)

**분류**: ② safe local — 스펙 의도가 명백한 보완.

design.md D20 "D17의 나머지 절반"과 tasks.md 4.9는 `first_rank`·`first_rank_total`·
`first_rank_source` **세 개**를 지정한다. 구현은 `first_rank_at`을 하나 더 뒀다.

근거는 D17이 `first_price_at`에 대해 이미 쓴 문장 그대로다 — "기준선 세 일 된 것과 20분 된
것은 다른 사건이고, 값만 가진 독자는 어느 쪽인지 알 수 없다". 순위도 같고, 오히려 더하다.
가격은 늦은 기준선이 **과소평가**라는 알려진 방향을 갖지만 순위는 방향이 없다.

instant이 없으면 다음 두 가지를 할 수 없다.

- 저장된 순위가 이번 삶의 것인지 **읽는 쪽에서 검사**할 수 없다. `AssessExtended`가
  `BASELINE_TOO_LATE`/`_TOO_EARLY`로 하는 일이 seen_late에서는 불가능해지고, 저장 시점 가드
  하나만 남는다 — 마이그레이션·fixture·손으로 고친 파일은 그 가드를 통과하지 않는다.
- `Sighting.At`을 채울 수 없다. 그 필드의 기존 주석이 "두 instant이 함께 있는 것이 독자가
  이것이 정말 최초 관측인지 보는 방법 — D17이 `first_price_at`을 저장하는 것과 같은 이유"라고
  말한다. 측정된 sighting의 `At`이 zero가 되면 **absent가 zero로 읽히는** 새 구멍이 하나 는다.

칼럼 하나는 additive-nullable이고 WORKFLOW §0.6의 선호에 맞는다. Manager가 3칼럼을 고집하면
되돌리는 비용은 작지만, 되돌리면 위 두 가지를 잃는다.

## 2. veto 미측정 사유 4종 신설 (safe local)

`THRESHOLD_NOT_POSITIVE`, `BASELINE_TOO_EARLY`, `NO_FIRST_RANK`, `FIRST_RANK_UNDATED`,
`FIRST_RANK_NOT_FIRST`. design.md는 "거부한다"만 정하고 코드 이름을 정하지 않았다.

이 파일의 기존 규칙(대응이 다르면 사유도 다르다)을 따랐다. 특히 `THRESHOLD_ABSENT`와
`THRESHOLD_NOT_POSITIVE`를 나눈 이유: 전자는 아무도 설정하지 않은 계약이고, 후자는 값이
**숫자로** 도착했다는 뜻이라 렌더링하는 쪽(§5)을 봐야 한다. 화면이 둘을 구분하지 못하면
운영자는 다른 파일을 열게 된다.

## 3. `MeasureFirstSighting` 시그니처 변경 (safe local)

`(Candidate, []Observation)` → `(Candidate, FirstRank, []Observation)`.
`MeasureExpansion(Baseline, []Observation)`의 선례를 그대로 따른다. 저장값을 인자로 받지
않으면 "슬라이스가 대신 답한다"가 다시 가능해지고, 그것이 4.9가 고치는 결함이다.

현재 패키지 밖 호출자는 없다(`internal/console`·`cmd/tossctl` 어디에서도 부르지 않는다).
§5가 이 함수를 처음 부를 때 저장값을 함께 읽어야 한다.

## 4. `Collect`는 아직 `NoteFirstRank`를 부르지 않는다 (미완료, §5로 이월)

`scan.go`의 `Collect`는 `Promote`와 `NoteSources`만 쓴다. `NoteFirstPrice`도 부르지 않는다 —
§3에서 칼럼을 만들 때부터 그 상태였고, `NoteFirstRank`도 같은 상태로 뒀다. 둘 중 하나만
배선하면 두 veto 중 하나만 살아난다.

**§5(5.1)가 두 write를 함께 배선해야 한다.** 배선 전까지 `extended`는 `NO_BASELINE`,
`seen_late`는 `NO_FIRST_RANK`로 미측정이며, D18상 두 veto 모두 임계가 없어 어차피
`THRESHOLD_ABSENT`이므로 판정에는 차이가 없다. D16(발굴이 원장을 굶길 수 있다) 때문에
스캔당 write를 늘리는 결정은 §5의 예산 안에서 해야 한다.

## 5. 마이그레이션 백필이 D17의 것보다 좁고, **비대칭**이다 (safe local)

v2 → v3 백필은 `first_seen_at` **이후** 10분 안의 최초 ranked 행만 쓴다. D17의 가격 백필은
살아남은 가장 오래된 priced 행을 **무조건** 쓴다.

같이 갈 수 없다. 가격은 늦은 기준선이 과소평가라는 방향을 갖고 `first_price_at` + 읽는 시점
비교가 그것을 잡는다. 순위는 방향이 없으므로 무조건 백필은 "나중 순간의 위치를 최초 관측으로
영구 저장"이 되고, 아래쪽에서 아무도 구분할 수 없다. 창 밖이면 NULL이고 미측정이다 —
D10이 안전하게 만들어 둔 상태.

**비대칭인 이유가 더 중요하다.** 런타임 `nearFirstSighting`은 D20대로 대칭 ±10분이다 —
스캔이 승격과 관측 instant을 조금 떨어뜨려 찍을 수 있기 때문이다. 그런데 마이그레이션에서
`first_seen_at` **이전** 쪽은 정확히 D20이 지목한 두 형태가 사는 자리다: 간격의 죽은 삶 행,
그리고 아침 내내 5위였다가 −10분에 100위로 밀린 종목. 대칭 창으로 백필하면 **읽는 시점에
고치기로 한 결함을 저장 시점에 구워 넣는다** — 그리고 그 뒤로는 아무도 구분할 수 없다.
`MeasureExpansion`의 주석이 쓴 규칙 그대로다: 그럴듯한 틀린 숫자가 없는 숫자보다 나쁘다.
대칭 술어로 되돌리면 실패하는 테스트가 있다
(`TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows`, 변이로 확인).

SQL의 창 비교는 고정폭 stamp의 앞 19자(`2026-07-28T09:00:00`)를 `strftime('%s', …)`로 읽으므로
소수부가 잘린다. 경계가 1초 이내로 흔들리지만 백필에서만 그렇고, 런타임의
`nearFirstSighting`은 정확하며 읽는 시점에 같은 검사를 다시 한다.

## 6. 리뷰 제안 중 채택하지 않은 것: `NearHighDistancePct` 빈 값의 기본값

§4 리뷰가 "`inputAge()`처럼 빈 `NearHighDistancePct`를 계약값 2.0으로 기본 지정하는 것도
고려하라(선택)"고 했다. **채택하지 않았다.**

- `inputAge()`가 기본값을 갖는 이유는 대안이 **위험**하기 때문이다(상한 없음 = 오래된 캔들을
  수락). 임계의 대안은 **안전**하다(`THRESHOLD_ABSENT` = 미측정, D10상 통과가 아니다).
  기본값을 넣는 비대칭 근거가 여기엔 없다.
- P0을 만든 현실 경로 자체가 빈 문자열을 만들지 않는다. `strconv.FormatFloat`는 `""`를
  출력하지 않는다 — 부재한 키는 `"0"`이고, 그것은 `THRESHOLD_NOT_POSITIVE`로 이미 거부한다.
  빈 문자열은 손으로 만든 `VetoThresholds{}`(테스트·fixture)에서만 오고, 그쪽에서는
  "아무것도 설정하지 않았다"가 정확한 답이다.
- 기존 `TestAnAbsentThresholdIsNotAPassedVeto`가 세 코드 모두에 대해 `THRESHOLD_ABSENT`를
  고정하고 있다. 기본값을 넣으면 그 테스트를 완화해야 하는데, 그것이 D10 + D18을 함께 적은
  테스트다. 조용한 measured 하나를 얻자고 시끄러운 unmeasured 하나를 잃는 거래다.

근거가 틀렸다고 보면 Manager가 결정하면 된다. 되돌리는 비용은 함수 하나다.
