# issues — refine-extended-shadow-bands

이 change가 범위 밖으로 남기는 관찰과, 구현 중 판단한 것.

**I1~I4는 tasks §3이 요구한 기록**이다. **I5~I11은 구현 중에 찾은 것**이고, 그중 일부는
후속 change가 아니라 **경계의 기록**이다 — 고칠 것이 없고 다음 사람이 알아야 할 사실만 있다.
그런 항목에는 시점과 등급을 붙이지 않는다.

---

## I1. `SeenLatePercentileBands`의 하위 절반이 한 칸인 것은 안전한 쪽이다 (§3.1, D14)

`SeenLatePercentileBands = 50/70/80/90/95`는 상위 절반만 분해한다. 하위 절반 전체가
"아무것도 넘지 않음" 한 칸이고, 그것은 `extended`가 이 change 전에 갖고 있던 결함과
**형태가 같다**.

**그런데 방향이 반대다.** `seen_late`는 **높은** 백분위에서 걸린다. 임계가 놓일 구간은
상위 절반이고 그쪽은 이미 다섯 칸으로 나뉘어 있다. `extended`는 후보 임계가 붕괴한 칸
**안**에 있었고, `seen_late`는 붕괴한 칸이 임계가 놓이지 않는 쪽이다.

**이 change에서 건드리지 않는다**(D14). 재확인 조건: `watch` 세션의 `seen_late` 분포가
나오면. 현재 그 분포는 **존재하지 않는다** — I5가 그 이유를 적는다.

## I2. 이 change는 임계를 정하지 않는다 (§3.2)

`extended`는 이 change 이후에도 **veto하지 않는다.** `VetoThresholds.ExtendedGainPct`는
읽지도 쓰지도 않았고, `MeasureExtendedBand`가 그것을 읽기 시작하면
`TestMeasureExtendedBandNeverReadsTheVetoThreshold`가 소스와 동작 양쪽에서 실패한다.
이 change 전에는 읽어도 아무것도 실패하지 않았다.

눈금에 `6`이 있는 것은 승인의 흔적이 아니다. 2의 배수라서 격자 위에 있는 것이고, 그
우연이 무해한 이유는 §0이 실제로 막은 경로다(I6·review.md §0.4).

## I3. `extended` 임계가 승인되면 `BandsFor(VetoExtended)`는 nil이 되어야 한다 (§3.3, D11)

`near_high`에 밴드가 없는 이유가 band.go에 적혀 있다 — *"D18이 그 임계를 승인했으므로
그것은 veto이고, 살아 있는 veto 옆의 그림자 기록은 그림자를 대신 읽으라는 초대다."*

같은 규칙이 `extended`에 적용된다. 임계가 승인되는 순간:

1. `BandsFor(VetoExtended)`가 nil을 돌려주도록 바꾼다
2. `ExtendedGainBands`를 지운다
3. `TestMeasureExtendedBandNeverReadsTheVetoThreshold`를 지우거나, 임계를 읽는 것이
   **의도**가 된 사실을 그 테스트가 반영하게 고친다

**그때까지 임계와 눈금이 같은 숫자로 두 곳에 존재하는 상태는 임시다.** 기록하지 않으면
다음 사람은 `6`이 눈금에 있는 것을 승인의 흔적으로 읽는다.

## I4. `formatDecimal`의 비음수 전제 — 확인 결과 (§3.4)

tasks §3.4는 *"`formatDecimal`이 비음수 입력을 전제로 문서화되어 있다"*고 썼다.
**정확하지 않다. 확인한 사실은 이렇다.**

| 무엇 | 어디 | 사실 |
|---|---|---|
| `formatDecimal` doc | metrics.go:970-976 | *"truncated **towards zero**"* — 부호 전제가 **없다** |
| `truncateAt` 주석 | metrics.go:1016-1018 | *"which for the non-negative values this file produces is the floor"* — 이 문장이 비음수를 전제한다 |
| 음수 사례 | level.go:552-560 | `distancePct`가 **이미** 음수와 그 결과를 명시적으로 문서화한다 |

즉 비음수 전제는 `formatDecimal`이 아니라 `truncateAt`에 있고, 그것은 **정확성 주장이
아니라 "truncate == floor"이라는 동치 주장**이다. 음수에서 그 동치는 깨지고, 깨진 결과는
level.go가 이미 적어 둔 것과 같다 — 렌더링이 정확값보다 마지막 자리 하나만큼 0에 가깝다.

**결과(실측, `TestANegativeValueRendersTheWayTheRestOfThePackageSaysItDoes`)**:

- `3 → 2`(−100/3%)는 `-33.333333333333`로 렌더된다. 부호가 살아 있고 자리수는 12다.
- **교차는 영향받지 않는다.** `bandCrossings`는 정확 rational을 비교하고 문자열을 보지
  않는다. 0을 아슬아슬하게 밑도는 값(`300000000000000000 → 299999999999999999`)도
  밴드 `0`을 넘지 않는다.
- **분위수는 렌더된 값을 되읽는다.** `TallyBands`가 `b.Value`를 파싱하므로 분위수는
  사람이 화면에서 보는 숫자와 같은 숫자다. 이것이 옳은 방향이라고 판단했고 근거를
  band.go에 적었다.

**고칠 것 없음.** `truncateAt`의 문장은 자기 함수에 대해서는 여전히 참이 아니지만
(이 파일은 이제 음수를 만든다), 고치는 것은 `internal/candidate/metrics.go`의 주석
한 줄이고 이 change의 spec delta와 무관하다. 다음 사람이 그 문장을 근거로 재사용하지
않도록 여기에 적는다.

---

**아래는 구현 중 찾은 것.** I5·I6·I7은 **tasks.md 자체의 결함**이고 Manager가 읽어야
한다. I8~I11은 경계의 기록이다.

## I5. §2.7의 실측 결과 — 131개 값은 **전부 정확히 0**이고, 새 눈금도 이 데이터에서는 붕괴한다

tasks §2.7은 *"131개 값이 `[−∞,10)` 안 어디에 있는지"*를 재라고 했고, *"0~2에 몰려 있으면
새 눈금도 같은 붕괴이므로 그 사실을 그대로 보고하라"*고 했다. **몰려 있는 정도가 아니다.**

운영 store(`~/.local/share/tossos/candidates.db`)를 scratchpad로 **복사한 뒤 복사본을
읽기 전용으로** 조회했다(원본 무변경, 2026-07-29).

```
KR: n=131   min=+0.000000   max=+0.000000   exactly 0 = 131 / 131
US: n=230   min=-0.146843   max=+0.083682   exactly 0 = 216 / 230
```

`n=131`은 proposal이 인용한 바로 그 131이다. **그 131개는 전부 정확히 0이다.**

원인은 눈금이 아니다. **데이터에 경과 시간이 없다.**

```
observations의 서로 다른 observed_at: 3개
  2026-07-28T13:23:09Z  KR 230행
  2026-07-28T13:23:31Z  US 200행
  2026-07-28T14:53:19Z  US 200행

first_price_at 분포: 13:23:09 → 131건 · 13:23:31 → 61건 · 14:53:19 → 169건
```

KR은 **한 번만** 스캔됐다. 기준가(`first_price`)는 그 생애의 첫 판독에서 쓰이고, 최신
가격도 같은 판독이다. 그러므로 `gainPct(first, last)`는 **구조적으로 0**이다.
US의 비영값 14건도 시장의 움직임이 아니다 — 전부 같은 instant의 두 원천이 같은 종목에
서로 다른 가격을 보고한 것이고, 크기는 최대 0.15%다.

**두 눈금 아래의 교차 수(실측)**

| 눈금 | KR 131건 | US 230건 |
|---|---|---|
| 옛 눈금 `10·20·30·50·100` | 전부 0 | 전부 0 |
| 새 눈금 `0`만 | **131** | 225 |
| 새 눈금 `2·4·6·8·10·…` | 전부 0 | 전부 0 |

**판정 세 가지, 그대로 적는다.**

1. **KR 데이터에서 새 눈금도 붕괴한다.** 131건이 전부 밴드 `0` 하나만 넘고 나머지 아홉을
   아무도 넘지 않는다. `BandTally.Collapsed()`가 참이고 리포트는 **경보를 낸다** —
   요구사항이 실데이터에서 작동하는 것을 확인한 것이지, 눈금이 문제를 푼 것이 아니다.
2. **US에서는 밴드 `0`이 실제로 갈랐다**(225 vs 5). 이 change가 더한 열 개 중 **부호
   변화 하나만이 이 저장소의 데이터에서 무언가를 분해한다.** `2·4·6·8`은 이 데이터로
   검증도 반증도 되지 않는다.
3. **D13의 "2%p는 잠정"은 여전히 잠정이고, 이 실측은 그것을 좁히지 못했다.** 간격을
   정하려면 `watch` 세션이 필요하다 — 기준가와 최신 판독 사이에 **시간이 흐른** 데이터.
   일회성 `scan`은 그것을 만들 수 없다.

**후속 change 아님. 측정 조건의 문제다.** `watch`를 장중에 돌려 분포가 나온 뒤에
간격을 재검토한다(D13이 이미 그렇게 적었고, 이 항목은 그 조건이 아직 **한 번도**
충족된 적이 없다는 사실을 더한다).

## I6. tasks §0.1이 요구한 것만으로는 §0.4가 RED가 되지 않는다 — **구현이 넓혀야 했다**

**이것이 이 change에서 가장 중요한 정정이다.**

tasks §0.1은 선택자를 두 방향으로 넓히라고 했다: ① 판정 타입 집합에 `Verdict` 추가,
② 결과 타입에서 `*ast.StarExpr`·`ArrayType`·`MapType`도 이름을 꺼낸다.
tasks §0.4는 그 다음에 `if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() }`를
넣으면 §0.1이 **RED여야 한다**고 했다.

**그 둘은 동시에 성립하지 않는다.** §0.1만 적용하면 그 한 줄은 여전히 GREEN이다.

이유는 `bandNames` 집합에 있다. HEAD의 집합은 `ShadowBand`·`BandCrossing`·`BandTally`·
`MeasureSeenLateBand`·`MeasureExtendedBand`·`TallyBands`·`SeenLatePercentileBands`·
`ExtendedGainBands`·`BandsFor` 아홉 개다. 리뷰어의 한 줄이 언급하는 식별자는
`v`·`ExtendedBand`·`Crossed`·`Chase`·`Extended`·`RaisedVeto`이고 **아홉 개 중 어느 것도
아니다.** §0.3이 `MeasureSeenLateBand`/`MeasureExtendedBand`를 `assessOne`에서 빼내는
순간 `assessOne`에는 금지 식별자가 하나도 남지 않고, 그 위에 뒷문 한 줄을 얹어도
검사는 통과한다.

**실측으로 확인했다.** tasks §0.1을 **글자 그대로만** 구현한 복제본을 저장소 밖
(`scratchpad/litecheck`)에 만들어, 뒷문 한 줄이 들어간 상태의 `internal/candidate`에
돌렸다.

```
checked=12 failures=0
RESULT: GREEN — the literal 0.1 widening does NOT catch this state
```

**그러므로 구현은 §0.1을 넘어 두 가지를 더했다.** 근거와 함께 band_test.go에 적었다.

| 더한 것 | 무엇을 막는가 |
|---|---|
| `bandNames`에 `Crossed`·`Crossings` | 밴드를 **읽는** 어휘. band.go가 *"Crossed is the only predicate on this type"*이라고 쓴 바로 그 어휘이고, 필드·지역변수·인자 어느 경로로 받은 밴드든 판정으로 옮기려면 이것을 지나야 한다 |
| `Verdict`의 밴드 필드(`SeenLateBand`·`ExtendedBand`) **읽기** 금지 | 리뷰어가 실제로 쓴 철자. 대입의 좌변은 허용하고 그 밖의 등장은 금지한다 |

**읽기와 쓰기를 가른 것은 취향이 아니라 강제다.** `Verdict`에 밴드 필드가 있고 누군가
채워야 하므로 *"판정 함수는 밴드 필드를 언급할 수 없다"*는 규칙은 어떤 조립도 만족시킬 수
없다. 조립은 대입의 좌변이고, 결정으로 가는 경로는 **읽기**뿐이다.

**Manager가 확인할 것**: spec delta의 시나리오는 *"밴드 식별자를 **언급**하면 검사가
실패한다"*고 썼다. 구현은 *"조립(좌변)은 허용하고 읽기는 금지한다"*이다. 요구사항 문구를
그대로 두면 코드가 요구사항보다 느슨해 보이고, 실제로는 요구사항 문구가 **실현 불가능**하다.
문구 조정이 필요하면 requirement 수준 수정이므로 리뷰 게이트 재실행 대상이다.

## I7. spec delta의 SHALL 하나에 tasks가 없다 — 구현이 채웠다

spec delta(`specs/candidate-discovery/spec.md`)는 요구사항 본문과 **시나리오 하나**로
이렇게 쓴다:

> 측정된 기록이 하나 이상인데 그 전부가 같은 교차 집합을 냈으면, 보고 표면은 그것을
> 정상 수치로 표시하지 않고 **경보로 표시해야 한다**(SHALL).
>
> #### Scenario: 측정값이 전부 한 칸에 들어가면 보고 표면이 경보를 표시한다

**tasks.md에 이 항목이 없다.** §2.1~2.8 어디에도 경보가 없다.

구현하지 않으면 **거짓인 SHALL을 아카이브**한다 — D10이 눈금 순서에 대해 경고한 바로 그
권위 오염이, 같은 change의 다른 요구사항에서 일어난다. 그래서 구현했다:

- `BandTally.Collapsed()` / `CollapsedAlarm()` (`internal/candidate/band.go`)
- 스캔 리포트의 터미널·JSON 양쪽(`cmd/tossctl/candidate.go`), 교차 수치 **위**에 출력
- 테스트: `TestATallyThatResolvedNothingSaysSo`,
  `TestTheReportRaisesTheAlarmForAScaleThatResolvedNothing`

**판단이 필요한 결과 하나**: 스펙이 *"측정 건수가 1 이상"*이라고 썼으므로 **측정 1건도
붕괴로 센다.** 한 건은 분포가 아니라는 점에서 옳지만, 후보가 하나인 스캔은 항상 경보를
낸다. 스펙 문구를 따랐고, 임계를 바꾸려면(예: 2 이상) requirement 수정이다.

**측정 0건은 경보가 아니다.** `measured 0 of N` 줄이 이미 그 사실을 말하고, 거기에 경보를
붙이면 상시 점등이 된다 — proposal이 *"이 저장소가 이번 주기에 세 번 만난 형태"*의 첫
항목으로 든 것이 바로 상시 점등인 `degraded`다.

## I8. 경보와 분위수는 스캔 리포트에만 있고 콘솔 `/signals`에는 없다

`internal/console/signals.go`의 `signalsBandTally`는 `Note`·`Total`·`Measured`·
`Crossed`·`NotMeasured`만 렌더한다. 이 change가 더한 **분위수**와 **붕괴 경보**는 거기에
없다.

`TallyVerdicts`의 doc이 이 방향의 결함을 이미 한 번 기록했다 —
*"a band that is missing from a page renders as a band nobody crossed rather than as one
nobody counted (§5 review P2)"*. 형태는 그것과 같고 대상이 다르다.

**이 change에서 넓히지 않은 이유**: tasks §2.7이 *"리포트에 더한다"*라고 썼고, 붕괴가
읽히고 오독된 표면이 스캔 리포트다. 판단은 데이터가 있는 쪽에서 내리므로 그쪽을 먼저
고쳤다. **넓히는 것이 옳다고 생각하면 후속 change이고 등급은 Small이다** — 표시 전용,
판정 경로 무관.

## I9. 섹션 제목과 경보 문장은 80칼럼을 넘고, 접지 않았다

§2.6의 판단이다. 실측:

| 줄 | 폭 | 처리 |
|---|---|---|
| `shadow bands — extended (…; not a veto)` | **112** | 접지 않음 — HEAD 이전부터 그랬고 이 change가 만들지 않았다 |
| `crossed …` 열 개 항목 | **83** → 접힘 | **이 change가 접는다** |
| `values min … max …` | 74 | 접히지 않음(폭 안) |
| `ALARM extended: …` | **205** | 접지 않음 — `tallyAlarm`의 기존 문장들과 같은 형태 |

**가른 기준**: 산문은 단어 경계에서 소프트랩되어 읽히지만, `·`로 이은 수치 목록은
터미널이 접으면 **0칼럼**으로 돌아가 섹션 라벨 아래에 붙고 새 필드처럼 읽힌다.
접기는 목록에만 필요하다.

`reportWidth`(80)와 `wrapCounts`는 이 change가 만들었다. 산문 줄을 접는 것은 별도
판단이고 기존 `tallyAlarm` 출력 전체에 영향을 주므로 범위 밖으로 남긴다.

## I10. `shadow_acceleration.crossed`는 여전히 map이다 — 지금은 결함이 아니다

§2.8이 함께 볼지 판단하라고 했다. **넓히지 않았다.**

`candidate.ShadowThresholds = {"1.3","1.5","1.8","2.0","2.5"}`는 **전부 같은 자리수**라
`encoding/json`의 문자열 정렬이 수치 순서와 **일치한다.** 지금 이 표면에는 오류가 없다.

`ExtendedGainBands`는 그렇지 않다 — `0`·`2`·`10`·`100`이 섞여 있어 문자열 정렬이
`"0","10","100","2",…`가 된다. 그것이 §2.8이 고친 것이고, 고친 이유는 형식 통일이 아니라
**실제로 틀린 순서**였기 때문이다.

**남는 잠복 결함**: 가속 임계에 자리수가 다른 값(`10`, `0.9`)이 하나라도 들어오면 같은
결함이 조용히 생긴다. 지금 그것을 잡는 것은 없다. 후속 change로 만든다면 두 표면을
같은 `[{…}]` 형태로 통일하는 것이고 등급은 Small이다.

## I11. `wrapCounts`의 접기 표시가 폭 예산에 들어가야 한다 — 구현 중 발견하고 고쳤다

기록으로 남긴다. 첫 구현은 `line + " ·"`를 **폭 검사 뒤에** 붙였다. 그래서 정확히 80칼럼을
채운 줄이 표시를 받아 **82칼럼**이 됐다 — 이 함수가 막으려던 실패가, 그 실패를 고치는
코드에서 재발한 형태다.

`TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth`의 `"counts wide enough to fold twice"`
행이 81칼럼으로 이것을 잡았다. 섹션 렌더 테스트 하나만으로는 잡히지 않았다 — 그 tally는
접힘이 한 번뿐이고 경계에 걸리지 않는다. **속성 테스트가 실물 테스트를 잡았다.**

---

# §5 — 독립 리뷰(§4.4) 처리 기록 (2026-07-29)

I12~I19는 독립 리뷰가 낸 P1 둘과 P2 여섯을 §5에서 처리한 기록이다.
I8은 **닫혔다**(아래 I13). I3은 **한 줄 방어를 넣었다**(아래 I14).

## I12. R1 — 한 홉 건너면 뒷문이 다시 열렸다 (P1, §5.1) — 닫음

`TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`는 **판정을 생산하는 그 함수**에
대한 규칙이라, 판정을 반환하지 않는 헬퍼를 한 단계 끼우면 통과했다.

```go
func worthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }
// assessOne 안
if worthy(v) { v.Chase.Extended = RaisedVeto() }
```

**착수 전 재현(구현자, 2026-07-29)**: 위 두 줄을 넣은 상태에서 그 가드가 `PASS`,
`go vet ./...` 무출력, `go test ./...` **51/51 ok**. Manager·리뷰어 보고와 일치한다.

**고른 형태**: 변환 지점을 국소 규칙으로 금지한다(Manager 제안).
`TestNoFunctionTurnsAShadowRecordIntoSomethingElse`가
**그림자 기록을 뜯어보는 함수는 기록을 돌려주거나, 아무것도 돌려주지 않거나, 기록
타입의 메서드여야 한다**를 강제한다. 자세한 근거와 못 막는 것은 review.md "§5 수정".

**리뷰어의 전이적 검사를 고르지 않은 이유**: `assessOne → shadowBandsFor →
MeasureExtendedBand`가 오탐이 되고, 그 오탐을 피하는 방법은 결국 예외 목록이다.
국소 규칙에는 예외가 **하나도** 필요하지 않았다 — `TallyVerdicts`도 `assessOne`도
`Cycle`도 걸리지 않는다. 기록을 **통째로 나르는 것**과 **뜯어보는 것**을 구조로 가르기
때문이다.

## I13. R2 — 붕괴 경보가 콘솔 `/signals`에 없었다 (P1, §5.2) — 닫음. **I8을 대체한다**

I8은 이것을 "후속 change"로 남겼다. **틀린 판단이었다.** spec delta의 SHALL은 표면을
지정하지 않고 *"보고 표면은 … 경보로 표시해야 한다"*라고 쓴다. 콘솔은 살아 있는 보고
표면이므로, 그대로 아카이브하면 승인된 요구사항이 그 화면에 대해 거짓이 된다.

`signalsBandTally.Alarm` + `signalsBandAlarm` + `templates_signals.go`의
`{{if .Alarm}}<p class="bad">`로 배선했다. 판정은 `candidate.BandTally.Collapsed`이고
**콘솔은 문장만 소유한다** — `signalsTallyAlarm`이 같은 분리를 같은 이유로 한다.

**UI 마찰 없음**: 표시 전용이다. 확인 문구 타이핑·추가 승인·버튼·폼을 넣지 않았다
(`signals_test.go`가 그것을 이미 강제하고 있고 계속 통과한다).

**남은 것 (닫지 않음)**: 분위수(`BandTally.Quantiles`)는 여전히 스캔 리포트에만 있다.
SHALL이 아니고 §5.2도 경보만 지목했으므로 범위 밖으로 둔다. 넓힌다면 후속 change이고
등급은 Small이다.

## I14. R3 — 밴드가 없는 사유에서 `Collapsed()`가 공허하게 참이었다 (P2) — **넣기로 판단**

`TallyBands`는 `Crossed`를 `BandsFor(code)`로만 seed한다. 밴드가 nil이면 맵이 비고,
`Collapsed()`의 루프가 한 번도 돌지 않아 `Measured >= 1`이기만 하면 참이 된다.

**판단: 지금 넣는다.** 근거 셋.

1. **I3이 이것을 의무로 적어 놨다.** 임계가 승인되면 `BandsFor(VetoExtended)`는 nil이
   되어야 한다. 그날 `extended` 경보가 영구 점등된다 — 가정이 아니라 예약된 상태다.
2. **I13이 비용을 올렸다.** 경보가 이제 운영자가 열어 두는 화면에 뜬다. 상시 점등은
   없는 것보다 나쁘다 — 그 블록을 건너뛰는 법을 가르치기 때문이다. 이 요구사항의 근거가
   *"사람이 알아차리는 것에 맡겨서는 안 된다"*인데, 상시 점등이 정확히 그 상태를 만든다.
3. **문장이 거짓이 된다.** *"측정된 N건이 전부 같은 교차 집합을 냈다"*는 교차가 하나도
   없을 때 공허하게 참이다. 눈금이 없는 것은 눈금이 아무것도 못 가른 것과 다르다.

`Collapsed()` 첫 줄을 `t.Measured < 1 || len(t.Crossed) == 0`으로 바꿨다.
`TestATallyWithNoScaleIsNotReportedAsCollapsed`가 RED→GREEN으로 고정한다.
`BandTally.Collapsed`의 FLM B1을 갱신했다.

## I15. R4 — `wrapCounts`의 단위 혼용과 폭 미검사 (P2) — 남긴다

리뷰어 실측: `sep = " · "`와 `mark = " ·"`에 U+00B7이 들어 있어 **바이트 길이가 4와 3**인데
코드가 `len()`(바이트)을 `utf8.RuneCountInString`(룬)과 섞어 비교한다
(`candidatebands.go:134,138`). 방향은 보수적이라 넘치지 않지만, 함수 doc의
*"That marker is two columns and they are budgeted here"*는 **거짓이다** — 두 칼럼이 아니라
세 바이트를 넣는다. 그리고 첫 줄(`label + parts[0]`)과 이어지는 줄(`indent + part`)은
예산 검사를 통과하지 않아, 긴 part 하나면 81·105칼럼이 나온다.

**남기는 이유**: 오늘의 데이터로 도달 불가다(가장 긴 part가 분위수의 약 19칸). 표시 폭
결함이고 판정·저장·주문 어디에도 닿지 않는다. §5의 범위는 P1 둘이며, 여기서 렌더 산술을
다시 만지면 그 자체가 I11이 기록한 재발 형태다. **후속 change, 등급 Small.**
고칠 때는 `len(sep)`·`len(mark)`를 `utf8.RuneCountInString`으로 바꾸고 두 줄 형태를 예산에
넣는 것이 전부이며, `TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth`에 긴 part 행을
더하면 RED가 먼저 나온다.

## I16. R5 — review.md §2.7·issues I5의 US 수치는 두 스캔을 합친 값이다 (P2) — 정정

`US n=230 … 밴드 0이 225 대 5로 갈랐다`는 **어떤 표면도 낸 적 없는 값**이다.
13:23 스캔이 measured 61(밴드 0을 60), 14:53 스캔이 measured 169(밴드 0을 165)이고
61+169=230 · 60+165=225 · 1+4=5다. 인용된 `min=-0.146843`은 앞 스캔, `max=+0.083682`는
뒤 스캔에서 나온다.

**결론은 그대로 산다**: 밴드 `0`만 갈랐고 나머지 아홉은 아무것도 못 갈랐다.
틀린 것은 근거의 형태다. I5 본문은 지우지 않고 이 항목을 정정으로 붙인다 —
두 change에서 반복된 *"결론은 맞고 근거는 틀림"*의 다섯 번째다.

부수 정정: KR의 미측정 사유는 `NO_PRICE 25`이고 `NO_OBSERVATIONS 25`가 아니다
(review.md §2.6 샘플이 후자로 쓰여 있다).

## I17. R6 — §2.6의 렌더 예시는 합성 fixture다 (P2) — 정정

review.md §2.6의 출력은 실측 tally가 아니라 합성 fixture로 렌더한 것인데 실측처럼 읽힌다.
리뷰어 재측정: 실제 KR 행은 접지 않으면 **80칼럼**(84가 아니다)이고, 분위수는 **전부 `0`**
(`median 0.91`·`max 41.07`이 아니다). `candidatebands.go` 주석의 *"92 against the live
tally"*도 이 저장소로는 재현되지 않는다.

**결론은 유지**: 열 개 항목의 `crossed` 줄은 80칼럼 근처이고 접기가 필요하다는 판단은
맞다. 틀린 것은 숫자의 출처 표시다. **주석·문서 정정은 후속 change로 남긴다** — 여기서
`candidatebands.go`를 고치면 §5의 diff에 렌더 코드가 들어오고 FLM 대상이 늘어난다.

## I18. R7 — `orderedCounts`가 호출자 0이 됐다 (P2) — 남긴다

두 호출부가 `wrapCounts`로 옮겨가면서 죽었다. `make lint`는 `go vet ./...`뿐이라 unused
함수를 잡지 않는다. FLM은 *"본문이 orderedCountParts로"*라고만 적고 죽었다는 사실은
적지 않는다.

**남기는 이유**: 지우는 것이 옳지만 `cmd/tossctl/candidate.go`의 FLM 대상이 이미 셋이고,
함수 삭제는 그 파일의 diff hunk를 바꾼다. §5는 P1 둘의 diff로 좁게 유지한다.
**후속 change, 등급 Small** — 삭제 + FLM 갱신이 전부다.

## I19. R8 — `distancePct`의 "the one value in the package that can be negative"가 거짓 (P2)

`level.go:552`. `gainPct`도 음수이고 `ShadowBand.Value`가 그것을 렌더한다. I4가 선례로
인용한 문단 **바로 위**에 있고, I4의 목적이 *"다음 사람이 그 문장을 근거로 재사용하지
않도록"*이었으므로 여기 적는다.

이 change가 만든 문장이 아니고 `internal/candidate/level.go`는 §5의 범위 밖이다.
**후속 change에서 한 줄 고친다.** 그때까지 이 항목이 그 문장의 반례다.

# §6 — 2차 독립 리뷰(§5.5) 처리 기록 (2026-07-29)

I20은 §6.4가 실측으로 다시 썼다. I21~I24는 §6에서 새로 찾은 것이다.

## I20. 두 검사가 **막지 못하는 것** — 2026-07-29 §6.4에서 실측으로 다시 씀

> **이 항목은 §5.1이 쓴 원문을 대체한다.** 2차 독립 리뷰가 원문의 넷 중 **#4를 거짓**,
> **#1을 과소 서술**로 판정했고, **세 계열(void 함수·타입 별칭·임베딩)이 아예 빠져
> 있었다**고 지적했다. §6이 그 셋을 닫았고, 나머지를 **전부 변이로 다시 재어** 아래를
> 썼다. 원문은 남기지 않는다 — 틀린 문장을 근거로 다음 사람이 안심하는 것이 이 항목이
> 막으려던 바로 그 일이기 때문이다.
>
> 실측 표(넣기/명령/출력/되돌리기)는 review.md **"§6 수정 (구현)"** 절에 있다.

이제 방어는 **두 층**이다. 성질이 다르므로 못 막는 것도 다르다.

- **구문 가드 2종** — `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`(어휘)와
  `TestNoFunctionTurnsAShadowRecordIntoSomethingElse`(변환). 위반 **지점을 이름으로**
  가리킨다. `internal/candidate`의 비테스트 소스만 본다.
- **행동 속성 1종** — `TestChangingTheScaleChangesNoVerdict`(그리고 스캔 전체 판). 눈금만
  바꾸고 판정이 달라지는지 본다. **경로를 열거하지 않지만** 위반 지점을 말해 주지 못하고,
  **눈금에 의존하는 읽기만** 덮는다.

### 실측: 무엇을 어느 층이 막는가

`✔`=RED(막음), `–`=PASS(못 막음). 전부 넣고 돌리고 되돌린 결과다.

| 우회 계열 | 구문 | 행동 | 비고 |
|---|---|---|---|
| void 함수 + `Crossed`(S1/B1·B2·B3·B9) | ✔ | ✔ | §6.1이 닫음 |
| 타입 별칭·정의 타입(S2/B4·B9) | ✔ | ✔ | §6.2가 닫음. 행동은 읽기가 `Crossed`일 때만 |
| 임베딩 승격 접근(S2/B5) | ✔ | ✔ | 같음 |
| **기록 메서드가 `*Verdict`를 받는다**(N13) | ✔ | ✔ | §6.4가 새로 찾아 닫음 |
| **기록 메서드가 패키지 var에 쓴다**(N14) | ✔ | ✔ | 같음 |
| **기록을 반환하면서 `*Verdict`도 받는다**(N17) | ✔ | ✔ | 같음 — 결과 면제도 같은 결함이었다 |
| 클로저 캡처(N1) | ✔ | ✔ | |
| `*Verdict`의 void 메서드(N12) | ✔ | ✔ | |
| 인터페이스 지역변수(N15) / 기록 map(N16) | ✔ | ✔ | |
| **리플렉션 → `Crossings`** | – | ✔ | 구문은 못 본다. 행동이 잡는다 |
| **unsafe 오프셋 산술 → `Crossings`** | – | ✔ | 같음. `unsafe.Pointer(&v.ExtendedBand)`는 구문도 잡는다 |
| **리플렉션 → `Value`·`Measured`** | – | – | **아래 1** |
| **다른 패키지**(`internal/console`) | – | – | **아래 2** |
| 테스트 파일 | – | – | **아래 3** |
| 모호한 필드 이름 | 부분 | 눈금 의존이면 ✔ | **아래 4** |
| **패키지 var가 눈금 한 칸을 잡아 둔다**(M5) | – | – | **아래 5 / I26** |
| 패키지 var의 func 값이 **기록을** 읽는다(T3) | ✔ | ✔ | 변환 가드가 `assessOne`을 지목. I27 |

> **행동 열의 `✔`에는 조건이 하나 붙는다 — 2026-07-29 3차 리뷰 C3의 지적, 수용.**
> 위 표의 행동 `✔`는 전부 **판정 시점에 눈금을 읽는 경우**에 대해서만 참이다. 같은 읽기를
> 미리 해 두고 값만 들고 있으면 행동 층은 침묵한다(I26). 리플렉션이든 unsafe든 마찬가지다.
> 원래 표에 이 조건이 없었고, 조건 없는 `✔`는 이 change가 네 번 반복한 형태 — 실제보다
> 넓은 보장을 적어 두는 것 — 이다.

### 남은 것 다섯 — 이번에도 숨기지 않는다

> **5번은 2026-07-29 3차 리뷰가 찾았고 Manager가 재현했다.** 그것이 이 목록에서 가장
> 중요한 항목이다 — 앞의 넷은 전부 *기록*을 우회로 삼지만, 5번은 **기록을 한 번도 읽지
> 않는다.** 전문은 I26에 있다.

1. **눈금에 의존하지 않는 필드를 리플렉션으로 읽는 경로. 두 층이 모두 통과시킨다.**
   `reflect.ValueOf(v).FieldByName("ExtendedBand").FieldByName("Value")`로 분기하면
   `go vet` 무출력, 네 구문 가드 PASS, 행동 속성 PASS, **51개 패키지 전부 초록**이다.
   행동 속성이 못 잡는 이유는 정확하다: `Value`는 후보 자신의 산술이므로 눈금을 바꿔도
   변하지 않는다. 구문 가드가 못 잡는 이유도 정확하다: 어떤 식도 기록 값이 아니다.
   **완화하지 않는다.** 다만 성질은 적어 둔다 — 이 경로가 만드는 것은 "눈금이 판정했다"가
   아니라 **"승인되지 않은 임계를 손으로 박았다"**이고, 그 값은 `Expansion`이 이미
   가지고 있는 gainPct다. 즉 그림자 기록을 지우더라도 같은 방식으로 쓸 수 있는 결함이며,
   그것을 막는 것은 이 change의 요구사항이 아니라 `ExtendedGainPct` 자체의 규율이다.
   `TestMeasureExtendedBandNeverReadsTheVetoThreshold`가 그 규율의 절반을 이미 들고 있다.
2. **다른 패키지. 두 층이 모두 통과시킨다 — Manager의 §6 서술이 여기서 틀렸다.**
   tasks §6.3은 행동 속성이 *"다른 패키지"*를 덮는다고 썼다. **덮지 않는다.** 2026-07-28의
   그 한 줄을 `internal/console/signals.go:652`(`signalsRowFrom`)에 그대로 넣으면
   `go vet` 무출력, 구문 가드 PASS, **행동 속성 PASS**, 51개 패키지 전부 초록이다.
   원인은 단순하다 — 행동 속성은 `internal/candidate`에 있고 `assessOne`·`Assess`가 낸
   판정을 본다. 콘솔은 그 판정을 **받은 뒤에** 고치므로 candidate 쪽에서는 아무것도
   달라지지 않는다. 위험은 §5.1 원문이 쓴 것보다 크다: 허용목록 10개 파일에 방어가
   전무하고, 그중 하나가 이 change가 고친 `signals.go`이며, **그 화면이 임계를 승인할
   사람이 읽는 표면**이다. 닫으려면 소비 표면에서 같은 속성을 재는 테스트가 필요하고,
   그것은 이 change의 §6 범위 밖이다. **후속 change, 등급 Normal.**
3. **테스트 파일.** 두 구문 가드가 `_test.go`를 walk에서 제외한다. 행동 속성에 대해서는
   해당 없다 — 테스트 파일은 프로덕션 판정을 바꾸지 않는다. 남는 위험은 "가드를 통과하는
   예제 코드가 저장소 안에 존재한다"뿐이다.
4. **모호한 필드 이름.** 기록 타입으로 선언된 필드 이름이 이 패키지 어딘가에서 비-기록
   타입으로도 선언돼 있으면 그 이름은 기록 필드 집합에서 빠진다(`Crossings`가 그렇다 —
   `ShadowBand.Crossings []BandCrossing`와 `Acceleration.Crossings []ThresholdCrossing`).
   이것이 가속 계열을 검사에서 빼 주는 장치이면서 동시에 **검사를 무장 해제하는 방법**이다.
   §6에서 `pinnedRecordFields`를 둘에서 **넷**으로 늘렸다(`SeenLateBand`·`ExtendedBand`·
   `Bands`·`Quantiles`) — 리뷰 S7이 지적한 대로 뒤의 둘도 기록을 담는데 방어가 없었다.
   넷 각각에 대해 `type X struct{ <이름> int }`를 넣으면 `Fatalf`임을 확인했다.
   **남는 것**: 그 넷 밖의 이름으로 기록을 담는 새 필드는 여전히 무방비다. 눈금에
   의존하는 읽기라면 행동 속성이 잡는다.

5. **눈금 값을 판정 시점보다 먼저 읽어 두는 경로. 두 층이 모두 통과시킨다.**
   `var zzEdge = ExtendedGainBands[3]`로 패키지 초기화 때 `"6"`을 잡아 두고 `assessOne`에서
   그 값으로 거부하면, 네 구문 가드 PASS·행동 속성 둘 다 PASS·`go vet` 무출력이다.
   **그림자 기록을 한 번도 읽지 않으므로** 구문 층에는 볼 것이 없고, 잡아 둔 값이 눈금
   교체에 반응하지 않으므로 행동 층도 침묵한다. 이 경로가 만드는 것은 앞의 넷과 종류가
   다르다 — "기록이 판정했다"가 아니라 **"눈금 자체가 판정했다"**이며, 눈금은
   `BandsFor`를 통해 바이너리 안 어디서나 읽힌다. 전문·반례·유일한 비열거적 해법은 I26.

**결론.** §5.1 원문이 적은 넷 중 실제로 남은 것은 둘(#2 테스트 파일, #3 모호한 이름)이고,
`#1`(다른 패키지)은 **위험이 과소 서술**이었으며, `#4`(unsafe)는 **거짓**이었다 — unsafe는
오히려 두 형태 다 막힌다(하나는 구문이, 하나는 행동이). 그 자리를 대신 차지한 것이
위의 1번이고, 그것은 §5.1도 2차 리뷰도 적지 않았다.

## I21. 결과 타입 면제도 void 면제와 **같은 결함**이었다 — §6이 직접 찾아 닫음

2차 리뷰의 S1은 `mayReadRecords`의 *"아무것도 반환하지 않으면 허용"*을 지목했다.
§6.1 구현 중에 확인한 것은 **같은 결함이 나머지 두 면제에도 그대로 있었다**는 것이다.
셋 모두 함수의 **표면 하나**를 보고 "그러니 답이 밖으로 못 나간다"고 근사한다.

```go
func (b ShadowBand) decide(v *Verdict)                      // 수신자가 기록 → 면제
var hot bool
func (b ShadowBand) stash()   { hot = b.Crossed("6") }      // 수신자가 기록 → 면제
func hand(s Sighting, v *Verdict) (ShadowBand, ShadowBand)  // 결과가 전부 기록 → 면제
```

**셋 다 넣어 실측했고 세 개 모두 두 가드를 통과했다**(§6.1의 void 수정을 넣은 뒤에도).
`hand`는 특히 조용하다 — 호출부가 `v.SeenLateBand, v.ExtendedBand = hand(…, &v)`이므로
`assessOne`은 아무것도 읽지 않고 아무 금지 이름도 쓰지 않는다.

**그래서 면제를 열거로 고치지 않고 질문을 바꿨다.** `writesOutward`가 "이 함수가 결과가
아닌 통로로 무엇을 내보낼 수 있는가"를 묻는다 — 기록을 이름하지 않는 참조형 인자(포인터·
슬라이스·맵·채널·func·인터페이스와 그 위에 선언된 이름)와 패키지 var 대입 둘이다.
`[]ShadowBand`는 통로가 아니다(원소가 기록이므로 기록을 나르는 것이다) — 그래서
`TallyBands`가 예외 없이 통과한다. **예외 목록은 여전히 0개다.**

## I22. 행동 속성이 덮는 축은 "경로"가 아니라 **"눈금 의존성"**이다 — 범위를 정확히 적는다

tasks §6.3은 이 속성이 *"리플렉션·타입 별칭·포인터 인자·다른 패키지·unsafe를 전부 함께
덮는다"*고 썼다. 실측 결과 **축이 다르다.** 이 속성이 덮는 것은 경로의 종류가 아니라
**읽은 값이 눈금에 의존하는가**이다.

- 눈금에 의존하는 읽기(`Crossed`·`Crossings`)라면 리플렉션이든 unsafe 오프셋 산술이든
  **전부 잡는다** — 경로를 하나도 열거하지 않고. 이것은 tasks의 주장대로다.
- 눈금에 의존하지 않는 읽기(`Value`·`Measured`·`Reason`)는 **경로와 무관하게 못 잡는다.**
  눈금을 어떻게 바꿔도 그 값이 변하지 않으므로 판정도 변하지 않는다.
- 그리고 **다른 패키지는 축과 무관하게 못 잡는다** — 속성이 `internal/candidate`의
  판정을 재기 때문이다(I20 2번).

구문 가드가 정확히 그 반대다: 눈금 의존성과 무관하게 **기록 값에 적용된 셀렉터**를 본다.
그래서 `Value` 읽기는 구문이 잡고 리플렉션은 행동이 잡는다. 스펙이 *"둘은 대체 관계가
아니다"*라고 쓴 것의 구체적인 내용이 이것이며, 두 테스트의 doc 주석에 그대로 적었다.

> **2026-07-29 정정 — 축에 한정어가 하나 빠져 있었다.** 위 문장은 *"읽은 값이 눈금에
> 의존하는가"*라고 썼다. 정확히는 **"판정 시점에 눈금에 의존하는가"**다. 눈금 값을 미리
> 읽어 변수에 담아 두면 그 값은 판정 시점에 더 이상 눈금에 의존하지 않고, 속성은 참으로
> 남는다(I26에서 실측). 한정어 없이 쓴 문장은 이 change가 세 번 반복한 실패 — 실제보다
> 넓은 보장을 적어 두는 것 — 과 같은 형태이므로 여기서 좁힌다.

## I23. §6.3은 **비공허성 단언이 없으면 아무 말도 하지 않는다**

이 속성의 실패 모드는 틀린 답이 아니라 **침묵**이다. 후보가 측정되지 않았거나 눈금
변형이 실제로 다른 기록을 만들지 않으면, "판정이 안 변했다"는 문장은 "입력이 안 변했다"의
동어반복이 된다. 그래서 두 테스트 모두 세 가지를 먼저 강제한다.

1. 모든 fixture의 두 밴드가 `Measured`여야 한다(아니면 `Fatalf`).
2. `ExtendedBand.Value`가 fixture가 주장하는 gain과 같아야 한다.
3. **눈금 변형이 최소 한 쌍에서 기록을 실제로 바꿔야 한다** — 현재 실측은
   `30 of 30 (fixture, variant) pairs recorded a different scale reading`이다.

이 셋이 없으면 §6.3은 통과하는 모습으로 아무것도 검사하지 않는 검사가 되며, 그것이
이 change에서 세 번 일어난 실패의 정확한 형태다.

## I24. 눈금 변수를 테스트가 바꾸는 것은 **직렬 실행 전제**에 기대고 있다

`onScale`은 `ExtendedGainBands`·`SeenLatePercentileBands`를 잠시 갈아끼운다. 그것이
"프로덕션 코드가 실제로 읽는 것을 그대로 읽게 한다"는 이 속성의 힘이지만, 전제가 하나
있다 — `internal/candidate`의 테스트가 **병렬로 돌지 않는다.**

확인: 이 패키지에 `t.Parallel()`은 **0건**이다(`cmd/tossctl`에는 8개 파일에 있으나 다른
패키지이고 다른 프로세스다). `-race`로 세 패키지를 돌려 초록임도 확인했다.

**남는 위험**: 누군가 이 패키지의 테스트에 `t.Parallel()`을 넣는 날 이 두 테스트는
경합한다. 지금 막을 수단은 있다(가드 하나 더)지만, 없는 문제에 검사를 붙이는 것이
이 change에서 여러 번 지적된 형태이므로 **기록만 남긴다.** `-race`가 그날 잡는다.

## I25. `writesOutward`의 참조형 판정은 **이 패키지가 선언한 이름**까지만 닫힌다

`isOutwardType`은 구문 종류(`*T`·`[]T`·`map`·`chan`·`func`·`interface{…}`)와
**이 패키지의 타입 선언**으로 닫은 이름만 통로로 센다. 다른 패키지에서 온 이름은
`*ast.SelectorExpr`로 오고 그 이름이 인터페이스인지 구조체인지 구문만으로는 알 수 없다.

**그래서 오탐이 없다** — `MeasureExtendedBand(e Expansion, c Candidate, at time.Time, …)`의
`time.Time`이 통로로 잡히지 않는 이유가 이것이고, 잡히면 이 패키지의 정당한 기록 독자가
실패한다. **그리고 같은 이유로 미탐이 있다**:

```go
func read(sink otherpkg.Raiser, b ShadowBand) ShadowBand {
    if b.Crossed("6") { sink.Raise() }   // sink는 다른 패키지의 인터페이스
    return b
}
```

`sink`는 통로로 세어지지 않고, 함수는 기록을 반환하므로 면제된다. **닫으려면 타입
정보가 필요하다**(`go/types`) — 구문 검사의 경계이고, 지금 넣으면 이 파일이 타입 검사기가
된다. **행동 속성이 이 경우를 덮는다**(`Crossed`는 눈금 의존 읽기다). 구문 층에서는
잔여물로 남기고 3차 리뷰어에게 먼저 볼 자리로 지목했다(review.md §9.4).

## I26. 눈금 값을 **판정 시점보다 먼저** 읽어 두면 두 층이 모두 통과시킨다 — 3차 리뷰 M5, Manager 재현

**이 항목이 이 change에서 가장 중요한 기록이다.** 앞선 세 번의 돌파는 전부 그림자 *기록*을
우회로 삼았고, 그래서 매번 "기록을 읽는 것을 더 잘 막자"로 닫혔다. 이것은 기록을 **한 번도
읽지 않는다.**

### 넣은 것

```go
// 패키지 수준. 초기화 시점에 평가된다.
var zzEdge = ExtendedGainBands[3]   // "6" — 지금 논의 중인 임계 후보 그 값

// assessOne 안:
if exceeds, ok := v.Expansion.GainExceeds(zzEdge); ok && exceeds {
    v.Chase.Extended = RaisedVeto()
}
```

### 결과 (2026-07-29, 이 worktree)

`go build ./...` 통과. `go vet` 무출력. **네 구문 가드 전부 PASS. 행동 속성 둘 다 PASS.**
`go test ./internal/candidate/` → `ok`. 되돌린 뒤 `sha256sum -c` OK.

승인된 적 없는 숫자가 — 프로덕션 눈금에서 직접 꺼낸 값이 — 거부권을 결정했고, **아무것도
실패하지 않았다.** 이 저장소가 반복해서 만나는 실패 형태 그대로다.

### 순서로는 닫히지 않는다 (실측)

`{"the production scale, run last", ExtendedGainBands, SeenLatePercentileBands}`를 **일곱
번째 변형으로 덧붙여** 프로덕션 눈금을 마지막에 한 번 더 돌렸다. **두 테스트 다 여전히
PASS.** 이유는 단순하다 — `zzEdge`는 패키지 초기화 때 평가되고, 그것은 `onScale`이 어떤
변형을 설치하는 것보다 먼저다. 변형의 순서는 초기화보다 뒤에 있는 모든 것에만 영향을 준다.

`base`를 옮겨서 프로덕션을 첫 자리에서 빼는 방식도 안 된다. 그렇게 하면 스캔 판의
비공허성 단언(`measured`)이 `"[]|[]"` 기준선 때문에 0이 되어 **다른 이유로** 실패한다.
즉 순서 변경은 잡지 못할 뿐 아니라 기존 단언을 망가뜨린다. **넣지 않았다.**

### 반례 — 경계가 정확히 어디인가

> **정정 2026-07-29 (Manager 착오, Teammate가 반박하고 Manager가 재현).** 처음에 반례를
> `GainExceeds(ExtendedGainBands[3])` **인라인**으로 적었다. **그것은 시점 경계를 분리하지
> 못한다** — `assessOne`은 `Verdict`를 반환하므로 어휘 가드의 walk 대상이고,
> `ExtendedGainBands`는 그 가드의 금지 이름에 있다. 재현:
> `band_test.go:371: watch.go:assessOne produces a verdict and mentions ExtendedGainBands`.
> 즉 인라인은 **구문 층이 잡는다.** 앞선 측정에서 이것을 못 본 이유는 빈 눈금 변형에서
> panic이 나면서 테스트 바이너리가 통째로 죽어 다른 테스트의 결과가 출력되지 않았기
> 때문이다. **panic을 실패로만 읽고 어느 층이 잡았는지 확인하지 않은 것이 착오다.**

경계를 실제로 분리하는 반례는 3차 리뷰의 **M5b**이고, Manager가 재현했다 — 같은 값을
**판정 시점에 live로** 읽되 판정 타입을 반환하지 않는 헬퍼를 거친다:

```go
func zzEdgeLive() string { return ExtendedGainBands[3] }
// assessOne 안: if exceeds, ok := v.Expansion.GainExceeds(zzEdgeLive()); ok && exceeds {
```

**구문 가드 4종 PASS(`ok`), 행동 속성 실패.** 헬퍼는 `string`을 반환하므로 어휘 가드의
walk에 들어오지 않고, 기록을 만지지 않으므로 변환 가드에도 안 걸린다. 구문 층은 눈금
변수 경로를 **어떤 형태로도** 보지 못하고, 잡는 것은 행동 층뿐이다.

행동 층이 실패하는 방식도 적어 둔다: 첫 노출은 `no scale at all` 변형에서 읽기 지점의
panic(`watch.go:400`)이고, 값이 있는 변형(`integerScale(-30,130)`에서 `[3]`은 `-27`)에서는
판정 비교로 갈린다. **어느 쪽이든 RED다.**

그러므로 경계는 언어 기능의 목록이 아니라 **읽는 시점**이다.

> **판정 시점에 눈금을 읽으면 경로가 무엇이든 잡힌다. 먼저 읽어 값을 들고 있으면 안 잡힌다.**

이 한 문장이 행동 속성이 실제로 보장하는 것이며, 이 change의 spec delta를 이 수준으로
낮춰 다시 썼다.

### 변수를 감추는 것으로는 좁아지지 않는다

`ExtendedGainBands`·`SeenLatePercentileBands`는 테스트를 빼면 `band.go` 안에서만 참조된다
(`grep -rn --include='*.go'`로 확인). 그래서 unexport가 자연스러운 다음 수로 보이지만,
`BandsFor(code)`가 exported이고 **슬라이스 자체를 돌려준다.** 콘솔이 실제로 그것을 쓴다.
var를 감춰도 `BandsFor("extended")[3]`은 같은 값을 준다. **모양 하나를 지우고 클래스를
그대로 두는 것**이고, 그것이 이 change가 네 번 반복한 실패다. **하지 않는다.**

### 열거가 아닌 유일한 해법 — 하지 않고 기록한다

눈금을 **패키지 상태가 아니라 호출자가 건네는 인자**로 만들면 초기화 시점에 잡을 대상이
존재하지 않는다. `MeasureExtendedBand`·`MeasureSeenLateBand`·`BandsFor`·`assessOne`과
콘솔·스캔 리포트의 소비 지점을 통과하는 실제 리팩터이고, 이 change의 범위가 아니다.

**등급 Normal, 후속 change.** 지금 넣지 않는 이유는 비용이 아니라 순서다 — 임계가 승인되면
`extended`는 그림자가 아니게 되고(I3), 그때 이 눈금이 어떤 모양으로 남아야 하는지가
달라진다. 승인 전에 배관을 바꾸면 두 번 바꾸게 된다.

### 이 결함의 실제 위험도

`ExtendedGainPct`에 대한 비테스트 대입은 **0건**이고(측정), `extended`는 오늘 어떤 후보도
거부하지 않는다. 이 경로는 **누가 새로 쓸 때** 열리는 것이지 지금 열려 있는 것이 아니다.
그래서 랜딩을 막지 않는다. 막는 것은 **이것을 적지 않고 랜딩하는 것**이다.

## I27. 3차 리뷰 T3는 **기록 읽기에 대해서는 거짓**이다 — 재현으로 확인

리뷰 T3는 *"어휘 가드가 판정 반환 `FuncDecl`만 walk하므로 패키지 수준 `var`가 자유롭게
참조한다"*고 썼다. **기록을 읽는 경우에는 그렇지 않다.** 넣어 본 것:

```go
var zzRead = func(b ShadowBand) bool { return b.Crossed("6") }   // GenDecl, FuncDecl 아님
// assessOne 안: if zzRead(v.ExtendedBand) { v.Chase.Extended = RaisedVeto() }
```

**두 층이 다 잡았다.**

- `TestNoFunctionTurnsAShadowRecordIntoSomethingElse` 실패 —
  `watch.go:assessOne takes a shadow record apart (hands one to zzRead) and does not
  hand a record back`. 변환 가드는 패키지 var를 walk할 필요가 없다. **건네주는 자리가
  `assessOne` 안**이고 그 자리를 본다.
- 행동 속성도 실패 — `Crossed`는 판정 시점의 눈금 의존 읽기다(`207940`에서 세 변형이
  다른 판정).

어휘 가드 하나만 침묵한 것은 사실이지만 **하중을 지고 있지 않다.**

**그러므로 T3는 별도 구멍이 아니다.** 패키지 var가 위험해지는 것은 그것이 *기록*이 아니라
*눈금 값*을 담을 때뿐이고, 그것은 I26과 같은 계열이다. 리뷰어가 두 경우를 하나로 묶어
보고했고, 재현해 보니 절반은 이미 막혀 있었다.

**기록해 두는 이유**: 다음 사람이 T3를 근거로 어휘 가드에 패키지 var walk를 추가하면,
이미 막혀 있는 것에 검사를 하나 더 붙이고 **진짜 구멍인 I26은 그대로 두게 된다.**
이 change에서 이미 여러 번 지적된 형태다.
