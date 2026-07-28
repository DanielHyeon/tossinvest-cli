# Review: set-chase-veto-thresholds (proposal-freeze)

두 개의 별도 문맥에서 병렬로 수행. 렌즈를 나눈 이유는 이 change의 위험이 두 종류이기
때문이다 — **숫자가 근거를 갖는가**와 **기계 장치가 제 안전장치를 지키는가**.

- 리뷰 A — 두 값과 그 도출 (2026-07-28)
- 리뷰 B — 메커니즘·가드·누락된 영향 범위 (2026-07-28)

결과: **초안의 핵심 주장이 무너졌다.** 두 값 모두 "인용"이 아니라 사실상 발명이었고,
설계한 안전장치는 배선 경로가 없었다. Manager가 직접 재확인한 항목은 아래에 표시한다.

---

## A. 값 — "인용이지 발명이 아니다"는 주장이 성립하지 않는다

### A-P0-1 패널은 150행이 아니라 100행이다 · **수용(직접 확인)**

design D1이 KR 패널을 150행으로 놓고 `12위/150 = 92.0`을 계산했다. 실제 배선:

| 근거 | 값 |
|---|---|
| `internal/candidatesrc/candidatesrc.go:268` | `OfficialRanking(..., typ, 100)` — KR·US 공통 |
| `internal/candidatesrc/candidatesrc.go:114` | `count > 100 → count = 100` |
| `internal/official/market_reads.go:332` | `max 100 per spec` |
| `internal/candidatesrc/candidatesrc.go:276` | `WTSPopular(wts, 30)` — KR 전용 30행 |

그러면 `12위/100 = 88.0`이고 **임계 90에서 걸리지 않는다.** tasks 3.3이 스스로 세운
합격 조건("예시가 재현되지 않으면 출처가 값을 뒷받침하지 않는다")을 이 값이 통과하지
못한다. 148위는 아예 존재할 수 없다(목록 최대 100행).

`candidatesrc.go:104`가 **이 실수를 하지 말라고 미리 적어 두었다** — "150을 요청하고
100을 받은 호출자는 존재한 적 없는 목록 길이에 대해 백분위를 계산하게 된다."

**150은 이 change의 오타가 아니라 저장소에 퍼져 있는 허구다.** `add-candidate-discovery`
design D8의 예시 자체가 148위를 쓰고, 출하된 코드 주석 다섯 곳이 반복한다 —
`metrics.go:797`, `metrics.go:890`, `store.go:181`, `store.go:1277`, `veto.go:481`.
백분위 정규화의 존재 이유로 적힌 문장("KR 150행과 US 100행에서 12위는 다른 사실")이
**두 시장 다 100행**이라 전제부터 틀렸다.

**판정**: 정정은 이 change의 범위에 넣는다. 허구를 남긴 채 그 위에서 값을 고르면 다음
사람이 같은 계산을 다시 한다.

### A-P0-4 격자 SHALL이 출처 없는 격자에 결정권을 준다 · **수용**

D8은 상한만 준다 — 12위가 걸리려면 임계 < 88이면 되고 {50, 70, 80}이 모두 만족한다.
90을 고른 것은 D8이 아니라 **밴드 격자**였다. 그런데 `band.go:20-25`가 직접 이렇게 쓴다:

> 밴드는 아무것도 결정하지 않기 때문에 출처 없이 골라도 되고, 무언가를 결정하는 순간
> 승인된 적 없는 정책 숫자가 된다 — **D6가 정문에서 막은 것이 뒷문으로 들어온다.**

design D3과 spec delta의 "임계는 격자 위의 값이어야 한다(SHALL)"가 정확히 그 뒷문이었다.

**판정**: SHALL 삭제. 밴드 유지(D3의 나머지 절반)는 옳고 남긴다. 값이 격자에 우연히
얹히면 편의로 적되 요구사항으로 만들지 않는다.

### A-P0-3 `extended`의 두 피연산자가 다른 구간을 잰다 · **수용**

`min_reward_risk`(`internal/risk/contract.go:200`)는 **진입→목표** 구간을 본다.
`extended`(`internal/candidate/level.go:225`)는 **발견가→현재** 구간을 본다. RR 게이트는
그 종목이 발견 후 0% 갔든 40% 갔든 동일하게 통과시키므로, "stop × RR = 우리가 잡으려는
이동폭"은 진입→목표 다리에 대해서만 참이고 `extended`는 그 다리를 재지 않는다.

피연산자 품질도 문제다. `internal/console/settings.go:34` — `defaultStopPct`는
*"운영자가 한 번도 고르지 않았을 때 콘솔이 채워 넣는 값"* 이라고 스스로 말하는 **비공개
UI 초기값**이고, 슬라이더 범위 0.02–0.2에서 같은 도출식이 **4%~40%**를 낸다.

저장소에 진입→목표 다리에 대한 값이 있다 — `internal/exitpolicy/ladder.go:138`
`DefaultLadderPolicy`의 목표 1.5/2.5/4.0/**6.0%**. 다만 `[미검증 — StockOS KOSPI 튜닝값]`.

**판정**: 10은 철회. 6은 재는 구간이 맞고 더 보수적이지만 `미검증`을 물려받는다.
gross vs net 방향은 리뷰가 맞다 — net 기준이면 목표가 더 커야 하므로 gross 사용은
임계를 **타이트하게(보수적으로)** 민다. 어느 쪽이든 이 방향을 문서에 적은 적이 없다.

### A-P0-2 승격에는 임계가 없다 — 세션 시작마다 상위권이 찍힌다 · **수용(직접 확인)**

design D1의 "희소성" 논거가 `Promote`의 **주석**("above the threshold")에 기대고
호출자를 보지 않았다. `internal/candidate/scan.go:340`은 모든 읽기의 모든 행을
`raisedBy`에 넣고 `scan.go:367`이 전부 승격한다. **패널에 있는 것은 전부 후보**다.

그리고 냉각 30분 + staleness 10분 = **40분** 공백이면 감시목록 전체가 만료되고
(`store.go:999`가 `first_rank`를 null로) 다음 스캔이 패널 전체를 동시에 재승격한다.
그때 각 종목이 **현재 순위**를 최초 관측 순위로 기록하므로, 100행 기준 상위 9위가
그날 내내 `seen_late`로 찍힌다 — 우리가 그 종목의 초반을 놓쳐서가 아니라 **우리 프로세스가
시작해서**.

**판정**: 희소성 문단 철회. 다만 리뷰의 "false positive" 규정은 절반만 맞다 — 세션
시작에 이미 거래대금 5위인 종목에 대해 우리가 늦은 것은 사실이다. 문제는 그 표식이
후보 수명 내내 붙어 있어 **6시간을 지켜본 뒤에도 "늦게 봤다"로 남는다**는 것이다.
일괄 승격에서 온 first_rank는 `seen_late`에 쓰지 않고 사유를 명명하는 것이 맞다.

### A-P1-5 얇은 목록에서는 어떤 순위도 걸릴 수 없고, 그것이 **측정된 clear**로 나간다 · **수용**

`percentileOf(1,1) = 0` — 1행 목록의 1위가 최하위로 읽힌다. 임계 90에서 `RankTotal ≤ 10`
이면 **어떤 순위도 raise될 수 없고**, WTS 30행에서는 1–2위만 가능하다.

지금은 `THRESHOLD_ABSENT`라 보이지 않는다. 임계를 넣는 순간 얇은/절단된 읽기가
**측정됨·안전**으로 나와 `Passed()`에 기여한다 — D10이 막으려던 실패가 미측정이 아니라
*측정된 답*을 통해 한 단계 아래에서 재현된다. 이 change가 없으면 드러나지 않았을,
이 change가 만드는 결함이다.

**판정**: 수용. `RankTotal` 하한 미만은 사유를 명명한 미측정.

### A-P1-6 / A-P1-7 · **수용**

- import 허용 목록에 `internal/candidatesrc`가 빠졌다(`candidatesrc.go:35`가 import한다).
- `seen_late` 백분위는 `100/RankTotal` 간격으로 양자화돼 있어 tasks 3.1의 "90.01"은
  `RankTotal = 10000`을 요구한다 — 도달 불가. 실제 경계 핀은 `10/100 = 90.00 clear`,
  `9/100 = 91.00 raised`다. `extended`의 10.00/10.01은 자유 유리수라 문제없다.

### A에서 살아남은 주장

- **부호·방향 정확**. `seen_late`·`extended` 모두 큰 쪽이 위험, `near_high`는 작은 쪽.
  `near_high`가 실제로 겪은 반전 사고가 두 새 행에서 재현되지 않는다.
- **주문 경로 무접촉**. `Chase.Passed()`의 비테스트 호출자는 `signals.go:615` 하나.

---

## B. 메커니즘 — 안전장치가 배선되지 않았다

### B-P0-1 D5에 배선 경로가 없고, 가장 자연스러운 구현이 경보를 삭제한다 · **수용**

D5는 "`VetoThresholds` 세 필드의 `thresholdReason`을 읽어 판단한다"고 썼다. **두 렌더러
어디에도 `VetoThresholds`가 없다.**

| | 근거 |
|---|---|
| `CycleResult`에 `Thresholds` 필드 없음 | `internal/candidate/watch.go:453-506` |
| `CycleOptions.Thresholds`는 `assessInto`에서 소비 후 폐기 | `watch.go:671-677` |
| `SignalsMarket`에 임계 필드 없음 | `internal/console/signals.go:106-127` |
| 콘솔은 임계를 본 적이 없다 — 리터럴은 seam 반대편 | `cmd/tossctl/console.go:898` |

구현자에게 열린 길은 셋이고 tasks.md는 셋 다 승인하지 않았다. 그중 **시그니처도 구조체도
안 건드리고 컴파일되고 전 테스트가 통과하는** 길이 하나 있다 — 문구 함수 안에서
`candidate.DefaultSeenLatePercentilePct`를 직접 읽는 것. 그러면 문구는 *적용된 임계*가
아니라 *상수*의 함수가 되고, `thresholdReason`이 언제나 빈 값이라 **D5 3행이 도달
불가**가 된다. 즉 D5를 만족한 것처럼 보이면서 `passedUnexpected`의 경보를 제거한다.

design D5가 그 결과를 스스로 명명해 두었다 — "그 행이 없으면 이 change는 자기 자신의
안전장치를 제거하는 change가 된다."

**판정**: 배선을 설계에서 정한다(구현 시점에 남기지 않는다). `Cycle`이
`CycleResult.Thresholds = opts.Thresholds`를 채우고, 문구 함수는 `VetoThresholds`를
**인자로** 받으며, `candidate.Default*Pct` 식별자를 참조하지 못하게 AST 테스트로 고정한다.

### B-P0-2 승인 이후에도 작동하는 경보는 tally 항등식이고, 그것은 이미 손에 있다 · **수용(설계보다 우수)**

D5 1행("셋 다 유효 → 정상")이 승인 후의 정상 상태이고, 거기서는 **모든** `Passed` 값이
정상으로 보고된다. 그런데 `THRESHOLD_ABSENT`는 미측정 사유 중 하나일 뿐이다 —
`NO_DAY_HIGH`·`INPUT_TOO_OLD`·`NOT_RANKED`·`BASELINE_TOO_LATE`·`THRESHOLD_UNREADABLE`·
`THRESHOLD_NOT_POSITIVE`가 모두 있다(`veto.go:200-232`). *어떤* 미측정이든 통과로 세는
버그는 1행 아래에서 보이지 않는다.

살아남는 검사는 산술이고 **임계가 전혀 없어도 계산된다**. `TallyVetoes`
(`veto.go:1101-1120`)에서 통과 후보는 모든 code의 `Raised`·`NotMeasured`에 0을
기여하므로, 모든 code에 대해

```
Passed + Raised[code] + NotMeasured[code] <= Total
Reasons[THRESHOLD_ABSENT] > 0 && Passed > 0     ← 직접 모순
```

캔들 예산 때문에 `NotMeasured[near_high]`가 목록의 다수이므로(D13 결정 3) 이 부등식은
느슨하지 않고 **날카롭다**. 그리고 `VetoTally`는 **두 파생 지점 모두에 이미 있다.**

**판정**: 채택. D5는 경보를 렌더러가 갖고 있지 않은 것(임계)에 태우고, 아무것도 필요
없는 것(tally)을 버렸다. 4행을 추가하고 spec의 마지막 시나리오도 tally 기준으로 옮긴다.

### B-P0-3 import 허용 목록이 두 곳에서 열린 채로 실패한다 · **수용**

전제 확인: **역방향 가드는 없다.** `isolation_test.go:255,285,313,345`는 전부 정방향.
task 7.1은 중복이 아니다. 다만 가드의 **형태**가 틀렸다.

1. `internal/candidatesrc/candidatesrc.go:35`가 `internal/candidate`를, `:37`이
   `internal/official`을 import한다. `official`에 `PlaceOrder`가 있다. 즉 허용 목록에
   반드시 올라가야 하는 패키지가, **`Chase.Passed()`와 `PlaceOrder`를 함께 부를 수 있는
   유일한 패키지**다. 목록이 그 줄을 공식 승인하게 된다.
2. `cmd/tossctl`은 **하나의 Go 패키지**이고 이미 `execgw`·`orderintent`·`trading`·
   `app/engine`을 import한다. 새 파일 `discovery_lane.go`가 `chase.Passed()`를 읽고
   `execgw`를 부르면 **import 간선이 0개 늘어난다.** D6이 막겠다고 쓴 문장이 정확히
   이것인데, 제안한 가드는 초록불이다.

**판정**: 패키지 단위가 틀린 단위다. 파일 단위 + 심볼 단위로 간다 — `Chase`/`Passed()`를
명명할 수 있는 **파일 목록**을 고정하고, 그 파일이 `execgw.`/`orderintent.`/`trading.`/
`official.Place`를 함께 명명하지 않음을 교집합으로 검사한다. 선례는
`internal/measure/isolation_static_test.go:216`(역방향 스캔)과 `:260`(심볼 명명).
양성 대조군도 함께 둔다 — 없으면 가드가 조용히 아무것도 금지하지 않는다.

### B-P1-1 밴드는 `>=`, veto는 `>` — 격자 위 임계에서 둘이 어긋난다 · **수용(가장 날카로운 지적)**

| | 비교 |
|---|---|
| `band.go:283` | `value.Cmp(bar) >= 0` — **포함** |
| `veto.go:544` · `level.go:225` | `Cmp(threshold) > 0` — **배제** |

`percentileOf`는 **100행에서 10위, 150행에서 15위에 정확히 90.00**을 낸다. 정수 순위이고
거의 모든 스캔에 존재한다. 그 후보의 그림자 기록은 "90 교차"라고 말하고 veto는
"측정됨·안전"이라고 말한다. **임계를 반증하겠다고 쌓는 분포가 경계 인구만큼 veto를
과다 계상한다.**

**판정**: A-P0-4의 SHALL 삭제로 대부분 해소된다 — 임계가 격자 위에 있을 의무가 없으면
둘이 한 점에서 만날 이유가 없다. 다만 우연히 겹치면 오차는 실재하므로 기록한다:
`seen_late`는 양자화 때문에 **한 순위 버킷 전체**가 어긋나고, `extended`는 자유 유리수라
동률이 사실상 없다. `band.go:33-37`이 `>=`를 "밴드 모서리는 히스토그램 버킷이지 경계가
아니다"로 의도적으로 택한 것이므로 밴드 쪽을 바꾸지 않는다.

### B-P1-2 임계 리터럴이 둘, 묶는 테스트가 없다 · **수용**

`cmd/tossctl/candidate.go:364`와 `cmd/tossctl/console.go:898`이 각각 리터럴을 만든다.
한쪽만 편집되면 `/signals`와 `tossctl candidate scan`이 같은 저장소·같은 시각에 대해
다른 `Passed`를 보고하는데, **D5가 각 표면의 문구를 그 표면의 임계에서 파생시키므로
두 화면 모두 내부적으로 옳게 읽힌다.** 파생이 분기를 은폐한다.

저장소가 이 실패를 이미 한 번 겪고 교훈을 적어 두었다 — `watch.go:686-692`, 네 번째
그림자 밴드가 스캔 출력에만 나타나고 `/signals`에는 안 나타난 §5 리뷰 P2.

**판정**: 단일 출처 + `cmd/tossctl` 생산 파일에 다른 `VetoThresholds{…}` 복합 리터럴이
없음을 AST로 고정. tasks 2.1/2.2는 정반대를 지시하고 있었다.

### B-P1-3 Function Logic Map 대상이 예측보다 한 자릿수 크다 · **수용(일정 위험)**

`tools/logic-map/check_analysis.py:90`의 diff 경로가 `"*.go"`라 **`_test.go`가 포함된다.**
(`add-candidate-discovery`의 106개 대상 중 37개가 `Test*` 함수다.) 그리고 수정된 파일의
경우 새 함수까지 요구된다(`:136-150`) — WORKFLOW의 "새 leaf 함수 면제"는 `required`가
비었을 때만 닿는다(`:308`).

**탈출구가 파일 배치에 있다**: `:115-118`이 `old_source == ""`이면 즉시 반환하므로
**새로 추가된 파일의 함수는 전부 면제**된다.

**판정**: 새 테스트는 **새 `_test.go` 파일**에 넣는다(증거 비용 0, 그리고 task 8.1의
*정정*과 *추가*가 분리된다). 대상 집합은 예측하지 말고
`check_analysis.py --change ...`가 보고하는 것으로 정의한다.

### B-P1-4 깨지는 테스트를 이름으로 지목 · **수용**

`cmd/tossctl/candidate_test.go:188,200,256` 세 개는 **하드 브레이크**(임계가 들어가면
`THRESHOLD_ABSENT`가 `Reasons`에서 사라진다). `cmd/tossctl/consolesignals_test.go:100-106`
은 수치는 살아남고 **메시지가 거짓**이 된다 — task 8.1(b)의 교과서적 사례.
`internal/console/signals_test.go:273`, `candidate_review_test.go:95`도 같은 부류.

tasks가 이름을 하나도 대지 않았고, `templates_signals.go:124-130,140`과
`consolesignals_test.go`는 task 목록 어디에도 없다.

### B-P1-5 다음 fallback이 쓰일 자리는 `inputAge()`가 아니라 `Cycle`의 옵션 블록 · **수용**

D10의 주장 자체는 검증됐다(세 함수 무변경 가능). 다만 유혹적인 자리를 잘못 짚었다.
`watch.go:522-533`이 이미 **네 개**의 옵션 필드를 기본값으로 채운다. 상수가 패키지에
생기는 순간 거기 두 줄을 더하는 것이 파일 자체의 관용구를 따르는 일이 되고, 그러면
모든 호출자의 `THRESHOLD_ABSENT`가 도달 불가가 된다 — task 2.3의 RED가 보는 곳보다
한 단계 위에서.

**판정**: `Cycle` 수준 RED를 추가하고 D10이 지목하는 선례를 교체.

### B-P2 · **수용**

- 거짓이 되는 doc 주석 **여덟 곳**(tasks는 둘만 지목): `candidate.go:14-19`,
  `signals.go:34-44`, `templates_signals.go:124-130`, `veto.go:760-768`, `veto.go:215-217`,
  `watch.go:288-291`, `watch.go:499-501`, `band.go:84-86`.
  특히 `band.go:84-86`은 **"승인된 임계를 가진 veto 옆의 그림자는 그림자를 읽으라는
  초대"** 라고 쓰여 있어, 이 change 이후 코드가 자기 주석을 위반하는 것처럼 읽힌다.
- archive 순서 의존이 산문뿐이고 `openspec validate`도 `gate.sh`도 잡지 않는다.

### B에서 살아남은 주장

- **spec delta 바이트 동일성 검증됨** — 겹치는 구간 `diff -u` 결과 공백. 같은 requirement를
  수정하는 다른 change도 없다.
- `base-commit.txt`(`b268593`)는 정확하고 현재 HEAD다.
- D10의 무-fallback 주장, tasks 3.4의 격자 원소 주장, tasks 9.2의 PM 이중 등록,
  `gate.sh` 조건 1–3 충족 가능성, 주문 경로 무접촉 — 모두 검증됨.

---

## Manager 판정 요약

**초안은 freeze하지 못한다.** 두 값 모두 근거가 없고, 안전장치는 배선이 없었다.

값과 무관하게 참인 수리(어느 방향으로 가든 필요):

1. `150` 허구 정정 — D8 예시 + 코드 주석 5곳
2. `RankTotal` 하한 미만 → 사유 명명 미측정 (A-P1-5)
3. 일괄 승격에서 온 first_rank는 `seen_late`에 쓰지 않음 (A-P0-2)
4. tally 항등식 경보 — 임계 유무와 무관하게 작동 (B-P0-2)
5. 파일+심볼 단위 소비자 가드 (B-P0-3)
6. 임계 단일 출처 + AST 고정 (B-P1-2)
7. 격자 SHALL 삭제 (A-P0-4)
8. 거짓이 되는 주석 8곳 + `band.go:84-86`에 D3 논거 인라인 (B-P2-1)
9. 새 테스트는 새 `_test.go` 파일에 (B-P1-3)

값에 대한 판정:

- **`seen_late`**: 이 change에서 값을 넣지 않는다. 저장소는 상한(< 88)만 주고 값을
  정해주지 않으며, 위 2·3을 고치기 전에는 어떤 값도 무엇을 재는지 알 수 없다.
- **`extended`**: 10 철회. 6(ladder 최종 목표)이 재는 구간이 맞고 더 보수적이지만
  `[미검증]`을 물려받는다. 투입 여부는 사용자 결정 대기.

사용자 승인(2026-07-28, "A")은 **철회된 도출 위에서 받은 것**이므로 그대로 이월되지
않는다. 값 투입 전 재승인 필요.

---

## 구현 후 독립 리뷰 3차 (2026-07-28) — pre-gate

별도 문맥. 모든 지적이 실행 가능한 probe 또는 mutation으로 재현된 상태로 도착했고,
Teammate가 각각을 코드에서 재확인한 뒤 반영했다. 기록은 `issues.md` I16–I19,
tasks.md §10.

| 등급 | 지적 | 판정 |
|---|---|---|
| P0-1 | `whole` 판정이 기억이 버리는 행(빈 심볼) 위에서 내려진다 | **수용** — issues.md I16 |
| P1-2 | `FirstRanksHeld`를 단언하는 테스트가 없다 | **수용** — 아래 |
| P1-3 | `recordFirsts`가 같은 tick의 자격 있는 읽기를 버린다 | **수용** — issues.md I17 |
| P2-1 | 기억 TTL의 두 번째 근거가 거짓 | **수용, 대체 근거는 수정** — I18 |
| P2-2 | `usableAt`의 zero instant 반쪽 미고정 | 수용 |
| P2-3 | `Panel`이 주입 clock을 전달하는지 미고정 | 수용 |
| P2-4 | 술어의 `RankRequested <= 0` 반쪽 미고정 | 수용 |
| P2-5 | `TallySightingSources`의 source 없는 skip 미고정 | 수용 |
| P2-6 | 소비자 가드가 못 보는 형태 | **기록만** — 아래 |
| P2-7 | P0-1 이후의 잔여 | **기록만** — issues.md I19 |

**P1-2**: `result.FirstRanksHeld++`만 지우고 `continue`를 남기면
`./internal/candidate/...`와 `./cmd/tossctl/...`이 전부 green이었다. 보류 *동작*은
`TestASessionStartDoesNotStampThePanelAsSeenLate`가 고정하지만 그 테스트가 카운터를 읽지
않았다. 그 테스트에 두 tick의 `FirstRanksHeld`(3 → 0)와 `FirstRanks`(0 → 3)를 더했다.
확인: mutation을 넣으면 이제 4건이 실패한다.

**P2-1에 대한 Teammate의 이견 (수용하되 근거는 다르게)**: 리뷰가 제안한 대체 근거
— "`nearFirstSighting`이 `first_seen_at`에서 `DefaultStalenessTTL`보다 먼 위치를 거부하므로
그보다 오래된 기억은 어차피 최초 순위를 자격 부여할 수 없다" — 는 **추론으로는 성립하지
않는다.** 기억의 나이는 `T_read - T_mem`이고 그 창이 묶는 것은
`T_read - first_seen_at`이다. 회복 읽기 자신이 새 생명을 시작시키면 `first_seen_at`이 그
instant이므로 창이 열려 있고, 그 경우가 정확히 F1 probe가 재현한 시나리오다. 그래서 두
문장으로 나누어 적었다 — 리뷰의 사실은 절반(기억보다 먼저 시작된 생명)에 대해 참이고,
덮지 못하는 절반이 이 bound가 존재하는 이유다. 같은 숫자, 두 기구가 양 끝을 닫는다.

**P1-2에 대한 부수 사실**: `internal/candidate/firstsighting_source_test.go`는
`base-commit.txt`(b268593) 기준 **신규 파일**이므로 이 편집은 Function Logic Map 대상을
늘리지 않는다(`check_analysis.py`는 `old_source == ""`이면 즉시 반환한다). 리뷰가 예상한
"추가되는 target"은 발생하지 않았다.

### P2-6 (기록만) — 소비자 가드가 구조적으로 못 보는 형태

같은 패키지의 한 줄 helper에서 클라이언트를 받아 `client.CreateConditionalOrder(…)`를
부르면 `cmd/tossctl/candidate.go`에서도 가드가 green이다. **회귀가 아니다** — 이미
`internal/candidate/consumer_guard_test.go:89-108`과 `cmd/tossctl/candidatepanel.go:15-21`이
같은 문장을 적고 있고, 그래서 가드가 금지하는 것이 **생성**(`official.New`·`tossclient.New`)
이며 그 두 줄이 판정을 읽지 않는 파일로 옮겨져 있다. selector 스캔은 값에 붙은 메서드를
볼 수 없고, 그것을 볼 수 있는 검사는 타입 해석(go/types)을 요구하는 더 넓은 change다.
다음 리뷰어가 이것을 새 결함으로 다시 발견하지 않도록 여기에 남긴다.

---

# 구현 후 리뷰 (task 8.4)

구현은 세 라운드로 나뉘었고 각 라운드 뒤에 별도 문맥 리뷰를 돌렸다. 리뷰가 매번 P0을
찾았고, 마지막 라운드에서야 P0이 나오지 않았다.

| 라운드 | 커밋 | 리뷰 | 결과 |
|---|---|---|---|
| §1–§7 구현 | `515c658` | 상태 있는 소스 렌즈 + 스키마·가드·테스트 렌즈 | **P0 3 · P1 6** |
| 수정 F1–F11 | `b49f7a2` | — | — |
| FLM 78 대상 | `f35302a` | 증거 생산 자체가 결함 발견 | **결함 1** |
| 최종 | `e0ada39` | 두 수정 커밋 공격 | **P0 1 · P1 2 · P2 7** |

## 라운드 1 — P0 셋

- **기억한 집합에 유효 조건이 없다.** 읽기 실패 시 `rememberRead` 전에 return하므로
  집합이 무한정 살아남는다. 그런데 프로세스가 살아 있는 동안 감시목록을 만료시키는 것은
  **소스가 답하지 않는 것**(429 백오프)이고, 그때가 정확히 기억이 온전히 보존되는 때다.
  즉 §2의 거부는 프로세스 재시작에만 발동하고 문서화된 트리거에는 발동하지 않았다.
  probe: 200회 연속 429 후 복구 → 만료 후 재승격에서 `measured=true percentile=96`.
  그리고 이것을 옳다고 고정하는 테스트가 근거로 든 전제("소스가 40분간 읽어 왔다")를
  **코드 어디에서도 세우지 않았다.** fixture가 단언할 뿐이었다.
- **절단된 읽기가 "직전 읽기"가 된다.** 3행 degraded 읽기가 100행 기억을 교체하면
  다음 온전한 읽기가 97개를 신규 진입으로 조작해 낸다. 이 change 자신의 문장
  ("절단된 읽기는 목록이 아니다")을 기억 쪽에 적용하지 않아서 생겼다.
- **§3의 절단 사슬에 배선 테스트가 없다.** `scan.go`의 두 링크를 각각 `0`으로 만들어도
  전 스위트가 통과한다. 같은 mutation을 형제 사실에 하면 3개가 깨진다 — §1은 고정돼
  있었고 §3은 아무것도 고정하지 않았다.

P1 여섯은 소비자 가드의 세 구멍(구성된 클라이언트의 메서드, 파일별 import 게이트,
order 쪽 alias 미해결), 미상 절단이 거부되지 않음, 업그레이드 전 후보의 회복 불가,
일회성 `scan`이 자격 없는 위치를 영구히 박음.

## FLM 생산이 찾은 것

이번 change가 만든 경계 거부 **둘에 테스트가 없었고**, 그중 하나는 테스트 표의 부제가
그 거부를 단언하고 있었다(`{"a negative request, which validate refuses at the boundary", …}`).
조용한 실패인 이유: `positive()`가 비양수를 SQL NULL로, `truncationOf`가 `<= 0`을
`TruncationUnknown()`으로 접으므로, 경계를 통과한 음수는 **"아무도 측정하지 않았음"으로
위장돼 저장**되고 write-once라 남은 수명 내내 그 값으로 답한다.

증거 생산자가 자기 증거를 한 번 정정했다 — `assessInto`의 새 줄을 "커버 없음"으로 썼다가
단언 대신 mutation을 돌려 덮는 테스트를 찾아냈다. 커버리지 주장은 실행해서 확인해야 한다.

## 최종 라운드 — P0 하나

**`whole` 판정이 기억이 버리는 행을 센다.** `total := len(raw.Items)`는 빈 심볼 행을
세고 `rows`와 기억 집합은 버린다. 그래서 판정은 거르지 않은 수로, 기억은 거른 집합으로
만들어진다. 3행 중 2행이 빈 심볼이면 다음 온전한 읽기가 2개를 신규 진입으로 보고하고,
전부 빈 읽기면 **빈 non-nil 집합**이 기억이 되어 `usableAt`을 통과하고 다음 읽기의
모든 행이 `yes`가 된다. F2가 닫았다는 결함이 다른 문으로 들어온 것이다.

`issues.md`가 "행 집합과 기억 집합이 같은 조건으로 걸러진다"고 단언했는데, 두 집합은
같게 걸러지고 **`whole` 판정만 안 걸러진다** — 그 틈이었다.

P1 둘: `FirstRanksHeld` 카운터를 아무것도 고정하지 않음(증가 줄을 지워도 전부 통과),
그리고 `recordFirsts`가 패널 순서상 첫 소스의 위치를 잡고 같은 tick의 자격 있는 위치를
버림 — 지속적으로 짧은 소스가 자기가 나열하는 모든 심볼의 `first_rank`를 무기한 막는다.

## 리뷰어가 틀린 것 (Teammate 반박, 수용)

- **F6의 제안(정확 일치 채움)은 도달 불가**다. `recordFirsts`는 이미 first rank가 있는
  후보에 읽기를 제공하지 않고, 제공하더라도 다음 tick의 관측은 다른 순간을 싣는다.
  억지로 만들면 **나중 읽기의 답을 저장된 위치 옆에 쓰는 것**이라 D3이 거부하는 치환이
  된다. 문장을 고치고 회복 경로(수명 종료 시)를 양방향으로 고정했다.
- **P2-1의 대체 논거는 따라오지 않는다.** 기억의 나이는 `T_read − T_mem`이고
  `nearFirstSighting`이 묶는 것은 `T_read − first_seen_at`이다. 회복하는 읽기가 새
  수명을 시작하면 창이 열려 있다 — 그게 이 bound가 추가된 바로 그 시나리오다.
  두 문장으로 나눠 양 끝을 각각 닫았다.

## 남은 것 — 고치지 않고 기록한 것

- **소비자 가드의 잔여 구멍**: 클라이언트를 같은 패키지의 한 줄짜리 헬퍼에서 얻으면
  `client.CreateConditionalOrder(...)`가 선택자 스캔에 보이지 않는다. `consumer_guard_test.go`와
  `candidatepanel.go`가 이미 "텍스트 검사는 이것을 볼 수 없다"고 적어 둔 것이라 F5의
  회귀가 아니다. 다음 리뷰어가 새 발견으로 재발견하지 않도록 여기 적는다.
- **`buildCandidatePanel`·`wtsPopularityReader`에 직접 테스트 0개.** `official.New` —
  주문 가능한 클라이언트를 만드는 줄 — 을 담은 함수다. 이번 change가 옮긴 것은 그 줄이
  어디 사는지이고, 그건 덮여 있다.
- **측정되지 않은 사실 하나가 이 change 전체의 유용성을 결정한다**: 공식 순위
  엔드포인트가 요청한 100행을 실제로 주는가. 주지 않으면 `seen_late`는 거의 측정되지
  않고 후속 임계 change는 고를 분포가 없다. `tossctl candidate scan` 리포트의
  원천별 `requested/arrived` 줄이 한 세션이면 답한다. **사람이 장중에 돌려야 한다.**
