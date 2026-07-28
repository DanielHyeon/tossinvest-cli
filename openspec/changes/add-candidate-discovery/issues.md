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

## 4. `Collect`는 아직 `NoteFirstRank`를 부르지 않는다 (§5에서 해소, 2026-07-28)

`scan.go`의 `Collect`는 `Promote`와 `NoteSources`만 쓴다. `NoteFirstPrice`도 부르지 않는다 —
§3에서 칼럼을 만들 때부터 그 상태였고, `NoteFirstRank`도 같은 상태로 뒀다. 둘 중 하나만
배선하면 두 veto 중 하나만 살아난다.

**§5(5.1)가 두 write를 함께 배선해야 한다.** 배선 전까지 `extended`는 `NO_BASELINE`,
`seen_late`는 `NO_FIRST_RANK`로 미측정이며, D18상 두 veto 모두 임계가 없어 어차피
`THRESHOLD_ABSENT`이므로 판정에는 차이가 없다. D16(발굴이 원장을 굶길 수 있다) 때문에
스캔당 write를 늘리는 결정은 §5의 예산 안에서 해야 한다.

**해소(§5, `recordFirsts`)**: 둘을 같은 자리에서 배선했다. 예산 결정은 다음과 같다.

- 순진한 배선(모든 readings마다 두 Note를 부르고 저장소가 idempotence를 책임진다)은 정확하지만
  **심볼당 tick당 IMMEDIATE 트랜잭션 2개**다 — DSN의 `_txlock=immediate` 때문에 "쓸 게 없다"를
  알아내려고 write lock을 잡는다. 150심볼이면 스캔당 300회다.
- 그래서 `Store.Summaries`(신규)로 후보 요약 + 두 stored first를 **한 질의**로 읽고, 칼럼이 아직
  NULL인 후보에만 쓴다. 저장소 쪽 once-guard는 그대로 둔다 — 그것이 두 writer 경쟁에서 이른
  기준선을 지키는 장치이고, 이 pass는 그 위의 최적화이지 대체가 아니다.
- 읽기는 promote **뒤**여야 한다. `Promote`가 만료 시 두 칼럼을 지우므로(D1), 앞에서 읽은 집합은
  방금 초기화된 삶에 대해 "이미 있다"고 답하고 새 삶이 죽은 삶의 기준선을 그대로 쓴다.
  `TestANewLifeRecordsItsOwnFirstPriceAndFirstRank`가 그것을 고정하고, 호출을 promote 루프 앞으로
  옮기는 변이로 확인했다.
- 어느 reading을 쓰는가: **panel 순서로 가격을 실은 첫 행**과 **순위를 실은 첫 행**을 따로 고른다.
  하나로 묶으면 가격만 있는 원천(prices)이나 순위만 있는 원천(WTS 인기)에서 다른 한쪽을 잃는다.
- 쓰기 실패는 `Rejected`에 이름과 사유로 남고 스캔은 계속한다. 승격은 이미 됐으므로 그 후보는
  집계되고, "요약은 있는데 기준선이 없다"는 사실이 보이게 된다.

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

## 7. §5가 스펙·설계에서 비어 있던 자리 넷 (safe local, 2026-07-28)

전부 "설계가 방향은 정했고 이름·값은 정하지 않았다"는 종류다. 값을 지어낸 곳은 없다 —
근거 없는 정책 숫자가 필요한 자리(veto 임계 둘)는 D18대로 여전히 비어 있다.

1. **그림자 밴드의 값**. D18은 "§3.5 선례대로 밴드로 기록한다"만 정하고 값을 주지 않았다.
   `seen_late` 50/70/80/90/95 백분위, `extended` 10/20/30/50/100%로 정했다. 근거는 §3.5의
   1.3/1.5/1.8/2.0/2.5와 **같은 등급**이다 — 밴드는 분포를 보여주는 도구이고 D18이 명시적으로
   "출처 없이 정해도 된다"고 쓴 범주다. 전자는 D8이 대조하는 12위/148위가 사는 상위 절반을,
   후자는 하루 범위에서 2배까지를 덮는다. **veto 임계는 여전히 비어 있고 사람 승인 대기다.**
2. **`Backoff` 사다리의 소유자**. tasks 5.4가 30→60→120→300초를 명시했지만 어느 타입이 갖는지,
   회복이 언제 일어나는지는 없었다. `internal/candidate/watch.go`의 `Backoff`가 갖고, 성공한
   읽기(=`ScanResult.Budgets`에 항목이 생긴 원천)가 사다리를 바닥으로 되돌린다. 되돌리지 않으면
   나쁜 1분이 세션 내내 5분 간격을 남긴다.
3. **여유 공간 프로브의 자리**. D16은 "여유 공간 하한을 둔다"만 정했다. `FSProber`가 이미
   `Open`에서 마운트를 판정하므로 같은 인터페이스에 `FreeBytes`/`FreeMeasured`를 가산하고
   `Store`가 prober를 보관해 주기마다 다시 묻는다. 두 값을 한 필드로 합치지 않은 이유는
   `Budget.Reported`와 같다 — 못 잰 것을 "0바이트 남음"으로 렌더하면 가장 놀라운 오독이고,
   "충분함"으로 기본값을 주면 D16이 막으려는 바로 그 방향이다.
4. **`watch` 바깥 tick의 하한**. spec R7은 "간격 하한"을 원천별로만 말한다. 루프 자체의 tick에도
   하한이 필요해서 3초(계약의 가장 빠른 원천 floor)로 뒀고, 0·음수는 **하한이 아니라 기본값**
   15초로 떨어진다 — `VetoThresholds.MaxInputAge`와 같은 규칙이다.

## 8. 스펙이 아직 말하지 않는 것 — Manager 판단이 필요한 두 가지

구현하지 않았고 값을 지어내지도 않았다. 기록만 남긴다.

1. **만료된 후보 요약을 아무도 정리하지 않는다.** §1·§2 리뷰가 `PruneExpiredCandidates`를
   "§4가 칼럼을 추가하기 전에는 위험하다"는 이유로 미뤘고, §4는 끝났다. D16이 계산한 부피는
   `observations`의 것이고 `candidates`는 후보당 1행이라 훨씬 작지만, **영원히 늘어난다**는 성질은
   같다. `Assess`는 만료 후보를 건너뛰므로 화면·집계에는 새지 않는다 — 순수한 용량 문제이고,
   D11의 "후보 요약은 만료까지"라는 문장에 아직 집행자가 없다. 값(만료 후 며칠)이 설계에 없다.
2. **`Assess`의 단위가 후보가 아니라 (후보, 원천) 계열이다.** 가속도 시계열은 D9상 원천별이므로
   `CrossingTally.Total`은 후보 수보다 크다(WTS 인기 원천은 거래대금을 싣지 않아 매번
   `FIGURE_ABSENT`로 계상된다). 정직한 단위이고 화면에 "series"라고 적었지만, T3.2가 임계를
   도출할 때 어느 원천의 계열을 쓸지는 이 change가 정하지 않는다.

## 9. §5 화면(5.5~5.8)이 스펙·설계에서 비어 있던 자리 셋 (2026-07-28, 구현)

### 9-1. 콘솔에 스캔 결과가 도달할 경로가 없다 (분류: ① blocking에 준함 — 기록 후 우회, Manager 판정 요청)

**발견**. tasks 5.7은 "강등을 **빠진 원천 이름**으로 말한다"를 요구하고, 그 이름이 있는 곳은
`ScanResult.Missing` 하나뿐이다(D4가 "스캔 결과에 … 어떤 원천이 왜 빠졌는지"라고 쓴 그 자리).
그런데 `ScanResult`는 **한 사이클 동안 그 사이클을 돈 프로세스의 메모리에만** 있다.
저장소는 후보별 `sources`·`sources_attempted`·`sources_responded`·`degraded`는 남기지만
**빠진 원천의 이름과 사유는 남기지 않는다.** 콘솔은 스캔을 돌지 않으므로 오늘의 배선에서
그 이름을 가질 방법이 없다.

**검토한 세 갈래.**

1. **콘솔이 TTL마다 `Cycle`을 돈다** — 진짜 `Missing`을 얻지만 열린 탭이 **두 번째 발굴자**가
   된다. `tossctl candidate watch`와 같은 미문서화 RANKING 한도를 나눠 쓰게 되고, D14 결정 2에
   따라 429 한 번은 느려지는 것이 아니라 **원천 상실**이다. 게다가 tick마다 심볼당 근사 중복
   관측이 한 줄 더 쌓여 §3 리뷰가 지적한 "중복 행이 사유 없이 가속도를 바꾼다"의 입력이 된다.
   **채택하지 않았다.**
2. **스캔 결과를 저장소에 남긴다(schema v4 `scans` 테이블)** — 콘솔은 계속 읽기 전용이고
   D4의 문장이 실제로 집행된다. 다만 **설계에 없는 스키마 버전**이고 마이그레이션 사다리를
   하나 더 만든다. `design.md`는 구현자가 고칠 수 없으므로 **Manager 판정 대상**이다.
3. **없음을 없음으로 남긴다** — panel을 `seam_unwired`로 렌더하고 이름이 어디 있는지(=
   `tossctl candidate scan` 출력) 문장으로 적는다. **채택.**

**근거**. console-operator-overview D7이 같은 형태를 이미 판정했다 — "미체결 패널이 이
change에서 미측정인 것은 결함이 아니라 규칙의 적용이다 … 그때까지 이 칸은 0이 아니라
`seam_unwired`다." 값을 지어내지 않고, 0으로 렌더하지 않고, 어디를 봐야 하는지 적는다.

**렌더는 완성되어 있다.** `SignalsPanel.Missing`이 채워지면 화면이 즉시 이름·사유·429 여부를
표로 적고, 그 렌더는 테스트로 고정되어 있다(`TestDegradationIsSaidWithTheNamesOfTheMissing
Sources`, 변이 M-E로 확인). 2번을 채택하면 배선 한 줄이 남는다.

**부분 보상**: 후보 행마다 저장소가 기록한 완전성(`4 / 5`)과 `강등` 표식, 그리고 그 후보를
올린 원천 이름을 렌더한다. 어느 원천이 빠졌는지는 못 말하지만 "강등이 있었다"는 후보 단위로
말한다.

### 9-2. 콘솔 미측정 사유 코드가 여덟 번째를 얻었다 (분류: ② safe local)

`discovery_unreadable`. 발굴 저장소를 읽지 못한 경우이고, console-operator-overview의 일곱 중
어느 것도 맞지 않는다 — `journal_unreadable`의 지시는 "엔진을 기동해 마이그레이션하라"인데
발굴 저장소는 다른 파일이고(I-4가 같은 이유로 재사용을 거절했다), `broker_read_failed`는
요청을 보내지도 않았으며, `seam_unwired`는 이 화면이 **따로** 렌더하는 다른 상태다.

I-2 판정문의 논리 그대로다: 자유 문장으로만 존재하는 사유는 셀 수 없고 없음을 테스트할 수도
없다. 다만 **선언은 `internal/console/signals.go`에 두고 `overview.go`의
`unmeasuredSentences` 맵에는 넣지 않았다** — 그 맵은 개요 화면이 스스로 개수를 단언하는
일곱-코드 열거이고(`TestTheUnmeasuredReasonsStayApart`), 한 화면의 코드가 다른 화면의
단언을 움직이면 그 열거는 더 이상 완전하다고 읽을 수 없다. 문장은 `unmeasuredDiscovery`가
합성하며, 코드·문장이 둘 다 나오는 것은 테스트로 고정했다.

spec·design(console-operator-overview)의 `unmeasured_reasons`는 여전히 일곱이다. 이 change의
화면이 하나 더 쓴다는 사실을 그쪽 문서에 반영할지는 Manager 판정이다.

### 9-3. `/signals`에 `readOnly` wrapper를 붙이지 않았다 (분류: ② safe local)

지시는 "GET-only, read-only"였다. 경로는 `c.session0(c.handleSignals)`로 등록했고
`c.readOnly(...)`는 쓰지 않았다 — `console.go`의 `readOnly` 주석이 "이 wrapper는 계좌 동사를
쓰는 유일한 경로(`/orders`)에서만 load-bearing이며, 모든 읽기 경로에 붙이면 그 한 곳이 보이지
않게 된다"고 명시하기 때문이다. `/signals`는 계좌 동사를 하나도 쓰지 않으므로 account-verb
금지와 CSRF 짝 테스트가 이미 덮는다. 화면에는 form도 POST 핸들러도 없다.

필요하면 되돌리는 비용은 한 줄이다.

## 10. §5 리뷰가 만난 계약의 빈자리 셋 (2026-07-28, 분류 ② safe local — Manager 보고)

§5 리뷰의 수정 중 셋은 스펙·설계가 답을 갖고 있지 않은 자리에서 결정을 내렸다. 코드로는
보수적인 방향이고 어느 것도 승인 범위를 넓히지 않지만, `design.md`와 spec delta는 구현자가
고칠 수 없으므로 여기에 남긴다.

### 10-1. "일정상 아무것도 due하지 않은 turn"이 스펙 어디에도 없다

spec Requirement 7은 간격 하한과 백오프 기록을 요구하고, D13은 간격을 원천에 붙인다. 그런데
**tick과 원천 간격이 다를 때 무엇이 일어나는가**는 어느 쪽도 말하지 않는다. 실제로는 두 숫자가
우연히 같아서(둘 다 15초) 문제가 보이지 않았고, `engineYieldFactor`가 한쪽만 두 배로 만들자
바로 갈라졌다.

**내린 결정 둘.**

1. 빈 panel은 오류가 아니라 기록이다(`CycleResult.Quiet`). `CycleResult`의 기존 doc이 이미
   "a retreat is a source we lost, a not-due source is the schedule working"이라고 적었고,
   그 문장이 오류 판정에 도달하지 않았을 뿐이다.
2. 루프의 대기는 `max(운영자 tick, 다음으로 읽을 수 있게 되는 시점)`이다. `--interval`은 이제
   **하한**이고, flag help와 `Long`을 그렇게 고쳤다. 스펙이 "간격 하한"만 규정했으므로 이
   해석은 요구를 좁히지 않고 넓히지도 않는다 — 늘어나는 상한은 panel에서 가장 빠른 원천의
   유효 간격이므로 설정으로 묶여 있다.

Manager 판정이 필요하면: spec Requirement 7에 "tick은 어떤 원천도 읽을 수 없는 시점에는
깨어나지 않는다"를 한 줄 추가하는 것이 정확한 반영이다.

### 10-2. 0행 200은 회복인가 — 스펙에 없고, 아니라고 정했다

§2-2가 "빈 reading은 응답하되 보증하지 않는다"를 정했지만 그것은 **냉각**에 대한 규칙이었다.
백오프 사다리의 회복 신호에 대해서는 아무 문장도 없었고, 구현은 `Budgets`(응답 전체)를 쓰고
있었다 — 즉 이 파일이 다른 어디에서도 증거로 취급하지 않는 판독이 유일하게 회복으로 세어졌다.

**결정: 행을 실은 응답만이 회복이다**(`ScanResult.Vouched`). 근거는 §2-2와 같은 방향이고,
추가로 빈 목록 반환이 rate limit에 걸린 서비스의 흔한 load shedding이라는 것이다. 대가는
작다 — 후퇴 창은 그대로 만료하므로 원천은 계속 읽히고, 달라지는 것은 **다음** 429가 한 칸
위에서 시작한다는 것뿐이다. 방향은 보수적(호출을 덜 한다)이다.

예산 헤더는 응답 전체에 대해 계속 기록한다. D13 결정 2의 측정값이고 body와 무관하게 참이다.

### 10-3. 정적 소비자 가드가 주장하던 범위가 실제보다 넓었다

`TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn`은 §4 리뷰가 §5에 넘긴 경고를 받는
테스트인데, 리뷰가 변이로 확인한 6종이 통과했다. 넷은 값싸게 넓혔고 둘은 남는다 —
결과를 변수에 담는 형태(dataflow가 필요하다)와 pair를 method value로 받는 형태
(`chase.NearHigh` 필드 읽기와 타입 정보 없이 구분할 수 없고, 둘 다 잡으면 **올바른 철자**에서
가드가 울린다).

`isolation_test.go`가 자기 경계를 doc에 적은 선례를 따라 **테스트 자신의 doc 주석에 적었고**
검출기 자체의 표 테스트를 붙였다. 스펙 문장("발굴은 주문 경로에 도달할 수 없어야 한다")과
달리 이 가드는 spec Requirement가 아니라 §4 리뷰의 후속이므로 스펙 변경은 필요 없다.
