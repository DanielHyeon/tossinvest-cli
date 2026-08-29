# review — refine-extended-shadow-bands

## 구현 기록 (Teammate)

- 날짜: 2026-07-29
- 역할: Teammate(구현). **독립 리뷰는 이 기록에 포함되지 않는다** — §4.4는 미완료다.
- base-commit: `ad16f56d7db04871699e060f120f249518fbee77`
- 브랜치: `feat/p0-foundation` (작업 트리에 미커밋 상태 — Manager 검토용)

### §4.2 base-commit 확인

`cat base-commit.txt` → `ad16f56d7db04871699e060f120f249518fbee77`.
`git log --oneline -1` → `ad16f56 fix(candidate): 응답한 적 없는 gainers 원천을 …
[retire-gainers-source]`. **tasks §4.2가 적은 `ad16f56`과 일치한다.** 재고정하지 않았다.

---

## 무엇을 바꿨는가 (파일별)

```
cmd/tossctl/candidate.go                 |  73 +++++-----
cmd/tossctl/candidate_test.go            |  21 ++-
internal/candidate/band.go               | 213 +++++++++++++++++++++++++-
internal/candidate/band_test.go          |  83 +++++++++---
internal/candidate/watch.go              |  39 +++++-
docs/pm/portfolio/_registry.yaml         |   3 +-   ← Manager가 이미 넣어 둔 것
tools/pm/test_generate_master_tracker.py |   1 +    ← 같음
7 files changed, 375 insertions(+), 58 deletions(-)
```

| 파일 | 무엇 |
|---|---|
| `internal/candidate/band_test.go` | §0.1·0.2 선택자 확대(`Verdict`, 포인터·슬라이스·맵 결과, 읽기 어휘, 밴드 필드 읽기 금지, 검사 함수 수 하한). §0.5 포인터 메서드 집합 |
| `internal/candidate/watch.go` | §0.3 `assessOne`에서 밴드 두 줄 제거 + 형제 함수 `shadowBandsFor` 신규 |
| `internal/candidate/band.go` | §1.1 눈금, §1.2·1.3 주석, §2.7 `BandQuantile`/`BandQuantilePoints`/`bandQuantiles`/`BandTally.Quantiles`·`QuantileBase`, spec 경보(`Collapsed`/`CollapsedAlarm`), `TallyBands` 값 수집 |
| `cmd/tossctl/candidate.go` | §2.8 JSON `crossed` 배열화, 분위수·경보 배선, 접힘 렌더 호출 |
| `cmd/tossctl/candidate_test.go` | §2.8 기존 JSON 테스트를 배열 형태로 + 순서 단언 |

신규:

| 파일 | 무엇 |
|---|---|
| `internal/candidate/bandguard_test.go` | §0.1의 헬퍼(`verdictResultNames`·`assignedTo`·`verdictProducers`) |
| `internal/candidate/bandscale_test.go` | §1.4·2.1~2.5·2.7의 테스트 + 경보·음수 렌더 |
| `cmd/tossctl/candidatebands.go` | §2.7·2.8의 신규 타입·렌더 헬퍼(`bandCount`·`bandQuantile`·`shadowBandReport`·`wrapCounts`·`reportWidth` 등) |
| `cmd/tossctl/candidatebands_test.go` | §2.6·2.7·2.8의 렌더 테스트 |
| `openspec/.../issues.md`, `analysis/function-logic/` 12건 | §3, §4.3 |

**신규 헬퍼를 새 파일에 둔 이유**: 기존 파일에 끼워 넣으면 diff hunk가 인접 함수를 FLM
대상으로 끌고 들어간다. `retire-gainers-source` 독립 리뷰가 같은 배치를 "게이트 회피
아님"으로 판정했고, tasks.md가 신규 테스트에 대해 같은 규칙을 이미 적는다.
**옮긴 것은 전부 이 change가 새로 만든 함수다** — 기존 로직을 새 파일로 옮겨 diff를 숨긴
사실은 없다. `cmd/tossctl/candidate.go`에 남아 FLM 대상이 된 것은 이 change가 실제로
수정한 `buildCandidateReport`·`writeCandidateTable`·`orderedCounts` 셋이다.

---

# ★ §0.4 뒷문 확인 — 이 change의 존재 이유

**결론: 뒷문은 막혔다. 그리고 tasks §0.1이 요구한 것만으로는 막히지 않았다.**

## 0.4-a 변이를 실제로 넣었다

§0(0.1·0.2·0.3·0.5) 전부 GREEN인 상태의 `assessOne`에 리뷰어의 한 줄을 **편집으로** 넣었다.

```go
	v.SeenLateBand, v.ExtendedBand = shadowBandsFor(summary, v.Sighting, v.Expansion, at, th)
	// MUTATION 0.4 — remove after observing RED.
	if v.ExtendedBand.Crossed("6") {
		v.Chase.Extended = RaisedVeto()
	}
```

```
$ go test ./internal/candidate/ -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v
[FAIL] TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
   band_test.go:377: watch.go:assessOne produces a verdict and reads ExtendedBand; assembling a band…
   band_test.go:371: watch.go:assessOne produces a verdict and mentions Crossed; a shadow band is a…
   band_test.go:394: checked 12 verdict-producing functions
$ go vet ./internal/candidate/     → No issues found   (vet은 여전히 통과한다. 통과해야 정상이다)
```

**두 개의 서로 다른 규칙이 각각 잡았다** — 필드 읽기(`ExtendedBand`)와 읽기 어휘(`Crossed`).
하나가 무력화돼도 다른 하나가 남는다.

**복구**: 편집으로 되돌렸다(`git checkout`·`restore`·`stash` 사용 안 함).

```
$ sha256sum internal/candidate/watch.go
1aa356933d8e9ad3ea2b98a4ded5d25972eb163c16897edff4240584ad7b0106   ← 변이 직전 기록과 동일
$ go test ./internal/candidate/   → 318 passed
```

## 0.4-b **tasks §0.1만으로는 GREEN이다** — 실측

tasks §0.1은 ① `verdicts`에 `Verdict` 추가, ② 결과 타입에서 `StarExpr`·`ArrayType`·`MapType`
처리를 요구했다. 그 **둘만** 구현한 복제본을 저장소 밖(`scratchpad/litecheck/`, 자체 go.mod,
표준 라이브러리만)에 만들어, **뒷문 한 줄이 들어간 상태의 `internal/candidate`**에 돌렸다.

```
$ go run . /mnt/D/project/axipient/TossOS/internal/candidate
checked=12 failures=0
RESULT: GREEN — the literal 0.1 widening does NOT catch this state
```

**이유**: HEAD의 `bandNames`는 타입과 생성자 아홉 개뿐이고, 리뷰어의 한 줄이 언급하는
식별자(`ExtendedBand`·`Crossed`·`Chase`·`RaisedVeto`)는 그중 어느 것도 아니다. §0.3이
`MeasureSeenLateBand`/`MeasureExtendedBand`를 `assessOne`에서 빼내는 순간 금지 식별자가
하나도 남지 않고, 그 위에 뒷문을 얹어도 검사는 통과한다.

**그래서 구현이 §0.1을 넘어 두 가지를 더했다.** 근거는 `band_test.go`와 issues.md I6에 있다.

| 더한 것 | 무엇을 막는가 |
|---|---|
| `bandNames`에 `Crossed`·`Crossings` | 밴드를 **읽는** 어휘. band.go가 *"Crossed is the only predicate on this type"*이라고 쓴 그 어휘다 |
| `Verdict`의 `SeenLateBand`·`ExtendedBand` **읽기** 금지(대입 좌변은 허용) | 리뷰어가 실제로 쓴 철자 |

**읽기/쓰기를 가른 것은 취향이 아니라 강제다.** `Verdict`에 밴드 필드가 있고 누군가 채워야
하므로 "판정 함수는 밴드 필드를 언급할 수 없다"는 규칙은 **어떤 조립도 만족시킬 수 없다.**
조립은 대입의 좌변이고, 결정으로 가는 경로는 읽기뿐이다.

**Manager 확인 필요**: spec delta의 시나리오는 *"밴드 식별자를 **언급**하면 실패한다"*이고
구현은 *"조립은 허용, 읽기는 금지"*다. 문구 조정이 필요하면 requirement 수정이므로 리뷰
게이트 재실행 대상이다. 자세히는 issues.md **I6**.

## 0.4-c `shadowBandsFor`가 `*Verdict`를 받지 않는 이유

`func recordBands(v *Verdict, …)` 형태였다면 그 함수 안에 `v.Chase`가 있고, 판정 타입을
반환하지 않으므로 검사 대상이 아니다 — **뒷문을 없앤 것이 아니라 옮긴 것**이 된다.
그래서 시그니처는 `(Summary, Sighting, Expansion, time.Time, VetoThresholds) (ShadowBand, ShadowBand)`이고
Chase가 아예 scope에 없다. 근거를 함수 doc에 적었다.

---

## 각 `[T]` task의 변이 확인

변이는 전부 **편집으로 넣고 편집으로 되돌렸다.** 복구는 sha256으로 대조했다.

| 원본 파일 | 기준 sha256 |
|---|---|
| `internal/candidate/band.go` (HEAD) | `79d205bc…1825b3` |
| `internal/candidate/band.go` (§1.1 적용 후) | `5b9ac7ef…1a28ea` |
| `internal/candidate/watch.go` (§0.3 적용 후) | `1aa35693…7b0106` |
| `internal/candidate/store.go` | `95cae5ba…c51b1` |
| `cmd/tossctl/candidate.go` (§2.7·2.8 적용 후) | `88bd8a2e…b76aac` |

### §0.1 — 선택자 확대

- **자연 RED**(0.3 적용 전): `assessOne`이 `MeasureSeenLateBand`·`MeasureExtendedBand`를
  언급해서 두 건. `checked 12 verdict-producing functions` — HEAD의 10에서 12로 늘었고,
  늘어난 둘이 `Assess`(`[]Verdict`)와 `assessOne`(`Verdict`)이다.
- **GREEN**(0.3 적용 후).

### §0.2 — 검사 함수 수 하한

`verdictProducers = 12`. `checked < verdictProducers`가 `t.Fatalf`다. 값의 근거는
2026-07-29 실측이고 `bandguard_test.go`에 적었다.
**변이 없음** — 하한을 실제로 깨려면 verdict-producing 함수를 지워야 하는데, 그것은 이
change가 만들 수 있는 상태가 아니다. 대신 하한이 **자기 자신을 잡는지**는 위 실측이 보인다:
HEAD의 검사 수가 10이었고 옳은 값은 12였다 — `checked == 0`은 그 차이에 대해 침묵했다.

### §0.4 — 뒷문

위 ★ 절. **RED 확인 + 복구 대조 완료.**

### §0.5 — 포인터 메서드 집합

- **GREEN**(변경 후, 그런 메서드가 없으므로).
- **변이: `func (b *ShadowBand) Dangerous() bool { return b.Measured }` 추가 → RED** ✔

```
band_test.go:291: *candidate.ShadowBand has a Dangerous method; a band is a measuring instrument …
```

  되돌린 뒤 `sha256sum internal/candidate/band.go` = `79d205bc…1825b3`(HEAD와 동일).

### §1.4 — `MeasureExtendedBand`가 임계를 읽으면 실패

- **GREEN**(현재 상태).
- **변이: `th.ExtendedGainPct`를 눈금 앞에 끼워 넣음 → RED** ✔ **두 반쪽이 모두 걸렸다**:

```
bandscale_test.go:230: MeasureExtendedBand mentions ExtendedGainPct … (AST 절반, 2건)
bandscale_test.go:253: the band changed when ExtendedGainPct went from "" to "6": … (동작 절반)
```

  되돌린 뒤 `5b9ac7ef…1a28ea`.

### §2.1 — 하락·보합·소폭 상승이 서로 다른 기록

- **자연 RED**(1.1 적용 전): `-20% / 0% / +9%`가 전부 `[]`. 그리고
  `a candidate exactly at its first price did not cross band 0`.
- **GREEN**(1.1 적용 후).
- **변이: 눈금에서 `"0"` 제거 → RED** ✔ (`-20%`와 `0%`가 같은 기록).
  `+9%`는 여전히 갈리므로 **한 쌍만** 실패한다 — 밴드 `0`이 정확히 무엇을 가르는지 보인다.

### §2.2 — 0~10 구간 분해

- **자연 RED**(1.1 적용 전): `+1/+3`, `+3/+5`, `+5/+7`, `+7/+9` 네 쌍 전부 충돌.
- **GREEN**.
- **변이: 눈금에서 `"4"` 제거 → RED** ✔ `+3% and +5% both recorded [0 2]` — **인접한 한 쌍만**.

### §2.3 — 기존 다섯이 움직이지 않는다

- **GREEN**(1.1 전후 모두 — 접미사가 유지되므로).
- **변이: 눈금에서 `"50"` 제거 → RED** ✔
  `the scale ends [8 10 20 30 100], want it to end [10 20 30 50 100]`.
  tasks §2.3이 적은 대로 **이 change 전에는 `"50"`을 지워도 전체 스위트가 통과했다.**

### §2.4 — 밴드별 열과 건수

- **자연 RED**(1.1 적용 전): 다섯 밴드의 기대 건수가 전부 0으로 나옴(5건 실패).
- **GREEN**. `len(tally.Crossed) == len(BandsFor(code))`도 함께 단언한다(5→10에서 유지).

### §2.5 — 밴드는 저장되지 않는다 (AST)

- **GREEN**. 허용 파일 밖 비테스트 파일 어디에도 `ShadowBand`/`BandTally`/`BandCrossing`이 없다.
- **변이: `internal/candidate/store.go`에 `type storedBand struct{ b ShadowBand }` 추가 → RED** ✔

```
bandscale_test.go:313: internal/candidate/store.go names ShadowBand. A shadow band is recomputed …
```

  되돌린 뒤 `95cae5ba…c51b1`.
- 공허 방지 있음: 어떤 파일도 매치하지 않으면 `t.Fatal`, 그리고 `band.go`·`signals.go`가
  반드시 매치돼야 한다.

### §2.8 — JSON 순서

- **자연 RED**: 기존 `TestTheScanJSONReportsTheCountsAnOperatorActsOn`이
  `json: cannot unmarshal array into Go struct field …crossed of type map[string]int`로 실패.
  배열로 바꾼 뒤 GREEN, 그리고 `BandsFor` 순서 단언을 추가했다.

### §2.6 접힘 — `wrapCounts`

- **변이: `reportWidth` 80 → 200 → RED** ✔
  `nothing folded in a section built to be too wide for one line; the fold is not being exercised
  and the width above proves nothing`. 되돌린 뒤 `88bd8a2e…b76aac`.
- **속성 테스트가 실물 테스트가 못 잡은 결함을 잡았다** — issues.md I11. 첫 구현은 접힘
  표시 `" ·"`를 폭 검사 **뒤에** 붙여서 80칼럼 줄을 82칼럼으로 만들었다.
  `"counts wide enough to fold twice"` 행이 `line 0 is 81 columns`로 잡았고, 표시 두 칸을
  예산에 넣어 고쳤다.

---

## §2.6 — 리포트 출력 전문과 80칼럼 판단

`writeCandidateTable`을 실제 tally로 호출해 렌더한 결과다(네트워크·운영 store 미사용).
실측 tally 모양: `measured 131 of 156`, 낮은 눈금에 세 자리 건수.

```
shadow bands — extended (records and decides nothing; this is what a threshold will be derived from; not a veto)
  measured     131 of 156
  crossed      0 118 · 2 44 · 4 21 · 6 12 · 8 7 · 10 0 · 20 0 · 30 0 · 50 0 ·
               100 0
  values       min -12.35 · p25 -0.42 · median 0.91 · p75 2.68 · max 41.07
  not measured NO_OBSERVATIONS 25
```

붕괴(경보) 상태:

```
shadow bands — extended (records and decides nothing; this is what a threshold will be derived from; not a veto)
  measured     131 of 156
  ALARM extended: all 131 measured record(s) produced the same crossings, so this scale resolved nothing and no threshold may be approved from it. The counts beside it are honest and answer no question
  crossed      0 0 · 2 0 · 4 0 · 6 0 · 8 0 · 10 0 · 20 0 · 30 0 · 50 0 · 100 0
  values       min -12.35 · p25 -0.42 · median 0.91 · p75 2.68 · max 41.07
  not measured NO_OBSERVATIONS 25
```

**판단 (생략하지 않았다)**

| 줄 | 폭 | 처리 |
|---|---|---|
| `crossed` 열 개 항목 | **83** | **접었다.** `reportWidth = 80`, 이어지는 줄은 라벨 폭(15칸)에 정렬하고 접힌 줄 끝에 ` ·` |
| `values` | 74 | 접히지 않음(폭 안). 같은 접힘 경로를 쓰므로 넓어지면 접힌다 |
| 섹션 제목 | **112** | **접지 않았다** — HEAD 이전부터 그랬고 이 change가 만들지 않았다 |
| `ALARM …` | **205** | **접지 않았다** — 기존 `tallyAlarm` 출력과 같은 형태의 산문 |

**가른 기준**: 산문은 단어 경계에서 소프트랩되어 읽히지만 `·`로 이은 수치 목록은 터미널이
접으면 **0칼럼**으로 돌아가 섹션 라벨 아래에 붙고 새 필드처럼 읽힌다. 접기는 목록에만
필요하다. 두 산문 줄의 폭은 issues.md **I9**에 실측으로 남겼다.

리뷰가 실측한 92칼럼과 여기의 83칼럼이 다른 것은 건수 자리수 때문이다. 어느 쪽이든 80을
넘고 접힌다.

---

## ★ §2.7 `Value` 분위수 실측 — **131개는 전부 정확히 0이다**

tasks §2.7은 *"0~2에 몰려 있으면 새 눈금도 같은 붕괴이므로 그 사실을 그대로 보고하라"*고
했다. **몰려 있는 정도가 아니다.**

운영 store를 scratchpad로 **복사한 뒤 복사본을 읽기 전용**으로 조회했다(원본 무변경).

```
KR: n=131   min=+0.000000   max=+0.000000   exactly 0 = 131 / 131
US: n=230   min=-0.146843   max=+0.083682   exactly 0 = 216 / 230
```

`n=131`은 proposal이 인용한 바로 그 131이다.

**원인은 눈금이 아니라 데이터에 경과 시간이 없다는 것이다.**

```
observations의 서로 다른 observed_at: 3개뿐
  2026-07-28T13:23:09Z  KR 230행      ← KR은 한 번만 스캔됐다
  2026-07-28T13:23:31Z  US 200행
  2026-07-28T14:53:19Z  US 200행
first_price_at 분포: 13:23:09 → 131 · 13:23:31 → 61 · 14:53:19 → 169
```

기준가는 그 생애의 첫 판독에서 쓰이고, KR은 판독이 하나뿐이므로 최신 가격이 곧 기준가다.
`gainPct(first, last)`가 **구조적으로 0**이다. US의 비영값 14건도 시장의 움직임이 아니라
**같은 instant에 두 원천이 보고한 가격 차이**이고 크기는 최대 0.15%다.

| 눈금 | KR 131건이 넘는 수 | US 230건 |
|---|---|---|
| 옛 `10·20·30·50·100` | 전부 0 | 전부 0 |
| 새 눈금 중 `0` | **131** | 225 |
| 새 눈금 중 `2·4·6·8·10·20·30·50·100` | 전부 0 | 전부 0 |

**그대로 보고하는 세 가지**

1. **KR 데이터에서 새 눈금도 붕괴한다.** 131건이 밴드 `0` 하나만 넘는다.
   `BandTally.Collapsed()`가 참이고 리포트가 경보를 낸다 — 요구사항이 실데이터에서
   작동하는 것을 확인한 것이지 눈금이 문제를 푼 것이 아니다.
2. **US에서는 밴드 `0`이 실제로 갈랐다**(225 vs 5). 이 change가 더한 열 개 중 **부호 변화
   하나만이** 이 저장소의 데이터에서 무언가를 분해한다.
3. **`2·4·6·8`은 이 데이터로 검증도 반증도 되지 않는다.** D13이 "2%p는 잠정"이라고 쓴 것은
   여전히 잠정이고, 이 실측은 그것을 좁히지 못했다. 간격을 정하려면 `watch` 세션이
   필요하다 — 기준가와 최신 판독 사이에 **시간이 흐른** 데이터. 일회성 `scan`은 그것을
   만들 수 없다. **그 조건은 아직 한 번도 충족된 적이 없다.**

전문은 issues.md **I5**.

---

## §3 기록 — issues.md

- **I1**(§3.1) `seen_late` 하위 절반은 임계가 놓이지 않는 쪽이라 안전하다(D14)
- **I2**(§3.2) 이 change는 임계를 정하지 않는다
- **I3**(§3.3) 임계 승인 시 `BandsFor(VetoExtended)`는 nil이 되어야 한다(D11)
- **I4**(§3.4) `formatDecimal`의 비음수 전제 — **tasks의 서술이 부정확했다.**
  전제는 `formatDecimal`이 아니라 `truncateAt`에 있고, 음수 사례는 `distancePct`가 이미
  문서화하고 있다. 교차는 영향받지 않고(정확 rational 비교) 분위수는 렌더된 값을 읽는다.
  `TestANegativeValueRendersTheWayTheRestOfThePackageSaysItDoes`로 고정했다
- I5~I11: 아래 "막힌 것·틀린 것" 참조

---

## 검증 명령과 실제 결과

| 명령 | 결과 |
|---|---|
| `go test ./...` | **3706 passed in 57 packages** (실패 0) |
| `go vet ./...` | **No issues found** |
| `gofmt -l .` | **출력 없음** (`$(go env GOROOT)/bin/gofmt` — PATH에 `gofmt`가 없다) |
| `go test -race ./internal/candidate/... ./cmd/tossctl/...` | **629 passed in 2 packages** |
| `python3 tools/pm/generate_master_tracker.py --check` | `hierarchy and generated trackers are current` |
| `openspec validate refine-extended-shadow-bands --strict --no-interactive` | 아래 게이트 절 |

**upstream 상속 테스트 회귀 없음.** 전체 건수는 `retire-gainers-source` 랜딩 시점의 3685에서
3706으로 늘었다(+21, 이 change가 더한 테스트). **줄어든 것은 없다.**

**§4.5 PM 등록 확인**: `docs/pm/portfolio/_registry.yaml:30`과
`tools/pm/test_generate_master_tracker.py:36` 양쪽에 `"refine-extended-shadow-bands"`가 있다.
두 항목은 **이 change의 커밋이 실어야 한다**(tasks §4.5의 주의).

---

## Function Logic Map (§4.3)

**면제를 주장하지 않았다.** `check_analysis.py`를 먼저 돌렸고 출력이 대상을 지목했다.
`assessOne`은 tasks의 예상대로 **대상이었고**, 내부를 바꾸기 **전에** 산출물을 만들었다.

최종 대상 **12건**, 전부 `analysis/function-logic/<pkg>--<func>/`에
`ast.json`(현재 리비전, 해시 일치) · `function-logic-map.md` · `branch-test-map.md` ·
`risk-pattern-report.md`.

| 함수 | 성격 |
|---|---|
| `internal/candidate/watch.go:assessOne` | **수정** — 밴드 두 줄이 나감. 분기 구조는 base와 동일(B1~B3) |
| `internal/candidate/watch.go:shadowBandsFor` | 신규(인접) — 분기 없음 |
| `internal/candidate/band.go:TallyBands` | **수정** — B5(값 파싱) 신규, 나머지 8개는 base 그대로 |
| `internal/candidate/band.go:bandQuantiles` | 신규(인접) |
| `internal/candidate/band.go:BandTally.Collapsed` / `.CollapsedAlarm` | 신규(인접) |
| `internal/candidate/band_test.go:TestNoFunctionThatProducesAVerdictCanSeeAShadowBand` | **재작성** — B7·B11 신규 |
| `internal/candidate/band_test.go:TestAShadowBandCannotBeReadAsAVeto` | **수정** — B1(두 메서드 집합) 신규 |
| `cmd/tossctl/candidate.go:buildCandidateReport` | **수정** — B14·B15 신규 |
| `cmd/tossctl/candidate.go:writeCandidateTable` | **수정** — B24·B27 신규, B20·B25·B26은 접힘 루프로 |
| `cmd/tossctl/candidate.go:orderedCounts` | **수정** — 본문이 `orderedCountParts`로 |
| `cmd/tossctl/candidate_test.go:TestTheScanJSONReportsTheCountsAnOperatorActsOn` | **수정** — B12~B14 신규 |

```
$ python3 tools/logic-map/check_analysis.py --change refine-extended-shadow-bands
[logic-map] refine-extended-shadow-bands: evidence complete or diff-proven exempt
```

Branch Test Map의 `RED observed` 열은 **실제로 관측한 것만** `yes`다. base 동작을 그대로
유지하는 분기는 `no`로 두고 현재 그것을 덮는 테스트 이름을 적었다.
`writeCandidateTable` B27은 오늘 **도달 불가**임을 그 줄에 적었다.

---

## 안전 불변식 검토

- 변경 파일 5개(+PM 등록 2) + 신규 4개. `internal/risk`·`internal/execgw`·
  `internal/exitpolicy`·`internal/trading` **전부 무변경**.
- 주문 제출·취소·정정, 손절·익절·사이징, Guardian·kill switch, intent journal·원장,
  reconciliation, retry matrix·rate limit, 인증·세션, 체결 감지 **어느 경로도 닿지 않는다.**
- **임계를 정하지 않았다.** `VetoThresholds.ExtendedGainPct`를 읽지도 쓰지도 않았고,
  이제 읽으면 테스트가 실패한다. `extended`는 여전히 veto하지 않는다.
- `SeenLatePercentileBands` **무변경**(D14).
- 토글·운영 설정 flip 없음. LIVE 주문 side effect 없음.
- **`tossctl candidate scan`을 실행하지 않았다.** 읽기 전용이지만 운영 store에 관측을 쓰고
  RANKING rate 예산을 쓴다. §2.6은 `writeCandidateTable`을 직접 호출해 렌더했고, §2.7은
  운영 store의 **복사본**을 읽었다(원본 무변경).
- 위험 등급: **Normal.**

---

## 막힌 것 · tasks가 코드와 안 맞는 것

### 1. §4.4 독립 리뷰는 하지 않았다 — 그래서 `make gate`가 2/8에서 실패한다

구현자가 자기 구현을 독립 검증할 수 없다. tasks.md의 4.4 체크박스를 비워 두었다.
**정상적인 결과이고 숨기지 않는다.**

**리뷰어가 반드시 직접 할 것**: §0.4의 변이를 **직접 넣고** RED를 보고 되돌릴 것.
그리고 0.4-b의 반례(§0.1만으로는 GREEN)를 직접 재현할 것 — `scratchpad/litecheck/`의
프로그램은 세션 종료와 함께 사라진다.

### 2. **tasks §0.1이 요구한 것만으로는 §0.4가 RED가 되지 않는다** (issues I6)

이 change에서 가장 중요한 정정이다. 위 ★ 0.4-b에 실측이 있다. 구현이 `bandNames`에
읽기 어휘를 더하고 밴드 필드 읽기를 금지해야 했다. **spec delta의 시나리오 문구
("밴드 식별자를 언급하면")는 실현 불가능하다** — 조립이 그 필드를 써야 하기 때문이다.
문구 조정은 requirement 수정이므로 Manager 판단이 필요하다.

### 3. **spec delta의 SHALL 하나에 tasks가 없다** (issues I7)

spec delta는 *"측정된 기록 전부가 같은 교차 집합을 냈으면 보고 표면은 경보로 표시해야
한다(SHALL)"*를 요구사항 본문과 **시나리오**로 쓴다. **tasks.md §2 어디에도 없다.**

구현하지 않으면 **거짓인 SHALL을 아카이브**한다 — D10이 눈금 순서에 대해 경고한 권위
오염이 같은 change의 다른 요구사항에서 일어난다. 그래서 구현했다
(`BandTally.Collapsed`/`CollapsedAlarm` + 두 표면 + 테스트 2건).

**판단이 필요한 결과**: 스펙이 "측정 건수가 1 이상"이라고 썼으므로 **1건도 붕괴로 센다.**
후보가 하나인 스캔은 항상 경보를 낸다. 스펙 문구를 따랐다.

### 4. §2.7의 전제가 데이터로 성립하지 않는다 (issues I5)

131개 값은 "0~10 어딘가"가 아니라 **전부 정확히 0**이고, 원인은 눈금이 아니라 **KR이 한 번만
스캔됐다는 것**이다. 새 눈금 중 이 저장소의 데이터에서 무언가를 가르는 것은 밴드 `0`
하나뿐이다. `2·4·6·8`은 검증도 반증도 되지 않았다.

### 5. tasks §3.4의 서술이 코드와 다르다 (issues I4) — **결론은 맞다**

비음수 전제는 `formatDecimal`이 아니라 `truncateAt`에 있고, 음수 사례는 `distancePct`가
이미 문서화하고 있다. 확인할 가치가 있다는 §3.4의 **판단은 옳았고 위치가 틀렸다.**

> 참고: `retire-gainers-source`에서 **"결론은 맞고 근거는 틀림"이 네 번** 반복됐다.
> 이 change에서 같은 형태는 §3.4(위치가 틀림)와 §0.1(요구가 불충분) 둘이다.

### 6. `wrapCounts`의 첫 구현에 off-by-two가 있었고, 속성 테스트가 잡았다 (issues I11)

접힘 표시 `" ·"`를 폭 검사 뒤에 붙여서 80칼럼 줄이 82칼럼이 됐다 — **막으려던 실패가
그것을 고치는 코드에서 재발한 형태**다. 실물 tally 렌더 테스트로는 잡히지 않았고(접힘이
한 번뿐이라 경계에 안 걸린다) 속성 테스트가 잡았다.

### 7. 경보와 분위수는 스캔 리포트에만 있고 콘솔 `/signals`에는 없다 (issues I8)

`TallyVerdicts`의 doc이 두 화면이 어긋나는 결함을 이미 한 번 기록했다. 형태가 같다.
tasks §2.7이 "리포트에 더한다"라고 썼고 붕괴가 오독된 표면이 스캔 리포트여서 그쪽만
고쳤다. **넓히는 것이 옳다고 판단되면 후속 change이고 등급은 Small이다.**

### 8. `shadow_acceleration.crossed`는 여전히 map이다 — 지금은 결함이 아니다 (issues I10)

`ShadowThresholds`는 전부 같은 자리수라 문자열 정렬이 수치 순서와 일치한다. 범위를
넓히지 않았고 이유를 적었다. 자리수가 다른 임계가 하나라도 들어오면 같은 결함이 조용히
생기며, 그것을 잡는 것은 없다.

---

## 남은 위험

| 위험 | 상태 |
|---|---|
| 눈금이 임계처럼 읽힌다 | 밴드→판정 경로가 실제로 막혔다. 변이로 확인(★ §0.4). **단 §0.1만으로는 안 막혔다** |
| 새 눈금도 붕괴한다 | **실측으로 확인됨**(KR 131건). 경보가 그 상태를 표시한다. 간격 결정은 `watch` 분포 이후 |
| `6`이 승인의 흔적으로 읽힌다 | issues I3이 은퇴 의무를 기록. 주석이 우연임과 그 우연이 무해한 이유를 적는다 |
| 보고가 길어져 읽기 어렵다 | 목록은 접힌다. 산문 두 줄은 접지 않았고 폭을 issues I9에 실측으로 남겼다 |
| 밴드가 늘어 tally가 어긋난다 | `len(Crossed) == len(BandsFor)`와 밴드별 건수를 단언(§2.4) |
| 두 화면이 어긋난다 | **열려 있음** — issues I8 |
| spec 시나리오 문구가 실현 불가능 | **Manager 판단 필요** — issues I6 |
| 독립 리뷰 미완료 | **§4.4 열려 있음.** `make gate` 미통과의 원인 |

---

## §4.6 — `make sdd-sync` → `make sdd-check` → `make gate`

기록(issues.md·review.md·FLM 산출물)까지 **전부 끝낸 뒤** 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain). GBrain은 `b2c261a6..ad16f56d`를 동기화하고 131 chunk 생성 |
| `make sdd-check` | **통과** — `[index-freshness] CodeGraph hard-evidence index matches the worktree`, `[agent-config] … synchronized`, memory `valid`, `[pm] … current`, sdd-test 전부 통과, typedb/neo4j running |
| `make gate CHANGE=refine-extended-shadow-bands` | **FAIL — 2/8단계에서 미완료 태스크 1건(`4.4 독립 리뷰`)** |

```
GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다
미완료 태스크 1 건:
111:- [ ] 4.4 독립 리뷰: 구현과 분리된 컨텍스트가 diff와 테스트를 재검증하고 review.md에
```

**게이트 실패는 사실 그대로 기록한다.** 실패 지점은 2단계 하나이고 원인은 §4.4가 열려
있다는 것뿐이다. 나머지 단계를 각각 손으로 돌려 확인했다:

```
1/8 tasks.md 확인                     OK
2/8 미완료 태스크 확인                 FAIL — 4.4 독립 리뷰 (구현자가 할 수 없다)
3/8 review.md 존재                    OK (이 파일)
4/8 Function Logic Map                OK — evidence complete or diff-proven exempt
5/8 make sdd-check                    OK
6/8 make test                         OK — FAIL 0건
7/8 make vet                          OK
8/8 make validate                     OK — 25 passed, 0 failed
```

`openspec validate refine-extended-shadow-bands --strict --no-interactive`
→ `Change 'refine-extended-shadow-bands' is valid`.

§4.6 체크박스는 **이 세 명령을 실제로 돌렸기 때문에** 채웠다. 게이트가 통과했다는 뜻이
아니다 — 통과하지 못했고 그 이유는 위에 적힌 하나다.

**주의**: `sdd-sync` 이후에는 이 파일의 이 절만 편집했다. 그 편집으로 CodeGraph
fingerprint가 다시 stale이 되었을 수 있으므로, 독립 리뷰 뒤 `make gate`를 다시 돌릴 때는
`make sdd-sync`부터 연속으로 실행할 것.

---

# 독립 리뷰 (§4.4)

- 날짜: 2026-07-29
- 리뷰어: 구현과 **분리된 컨텍스트**. 구현자의 보고를 액면 그대로 받지 않고 A~D를 전부
  직접 재실행했다. 구현자의 `scratchpad/litecheck/`는 사라졌으므로 반례도 다시 만들었다.
- 방법: 변이는 전부 **편집으로 넣고 편집으로 되돌렸다.** `git checkout --`·`git restore`·
  `git stash`는 **한 번도 쓰지 않았다.** 리뷰 시작 시점에 대상 9개 파일의 sha256을
  기록하고, 모든 변이 후 `sha256sum -c`로 대조했다.
- 원시 출력은 `rtk proxy <cmd>`로 받았다(hook이 `go test`를 요약해 버려서 단언 메시지가
  보이지 않는다).

## 복구 대조 — 리뷰 전후 sha256 동일

```
$ sha256sum -c baseline-sha256.txt
cmd/tossctl/candidate.go: OK          cmd/tossctl/candidatebands.go: OK
cmd/tossctl/candidate_test.go: OK     cmd/tossctl/candidatebands_test.go: OK
internal/candidate/band.go: OK        internal/candidate/bandguard_test.go: OK
internal/candidate/band_test.go: OK   internal/candidate/bandscale_test.go: OK
internal/candidate/watch.go: OK
```

리뷰 중 만든 임시 probe 테스트 2개(`internal/candidate/zz_review_probe_test.go`,
`cmd/tossctl/zz_review_probe_test.go`)는 삭제했다. `git status --short`가 리뷰 시작
시점과 동일하다(수정 7 + 신규 4 + change 디렉터리).

운영 store는 **복사본만** 읽었다. `sha256sum ~/.local/share/tossos/candidates.db`
= `5e083940a7ff585a5e50570b51cf32f58fada05c2b9851988f8e3c9ac43feb8e`이고 복사본과
같다(원본 무변경). `tossctl candidate scan`은 실행하지 않았다.

---

## A. §0.4 뒷문 — 직접 확인

### A1. 변이 투입 → RED (구현자 보고와 일치)

`assessOne`에 편집으로 넣었다:

```go
	v.SeenLateBand, v.ExtendedBand = shadowBandsFor(summary, v.Sighting, v.Expansion, at, th)
	// REVIEW MUTATION A1 — remove.
	if v.ExtendedBand.Crossed("6") {
		v.Chase.Extended = RaisedVeto()
	}
```

```
$ rtk proxy go test ./internal/candidate/ -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v
    band_test.go:377: watch.go:assessOne produces a verdict and reads ExtendedBand; …
    band_test.go:371: watch.go:assessOne produces a verdict and mentions Crossed; …
    band_test.go:394: checked 12 verdict-producing functions
--- FAIL
```

두 규칙이 각각 걸린다. 되돌린 뒤 `watch.go` = `1aa35693…7b0106`(기준과 동일).

### A2. 반례 재유도 — **구현자의 주장은 사실이다**

A1의 변이를 **그대로 둔 채** `band_test.go`를 tasks §0.1이 글자 그대로 요구한 상태로
되돌렸다(읽기 어휘 `Crossed`/`Crossings` 제거, `bandFields` 빈 맵). 선택자 확대
(`Verdict` + `StarExpr`/`ArrayType`/`MapType`)만 남긴 상태다.

```
$ rtk proxy go test ./internal/candidate/ -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v
    band_test.go:391: checked 12 verdict-producing functions
--- PASS
```

**`checked=12 failures=0` — 구현자가 보고한 것과 정확히 같다.** Manager의 §0.1은 실제로
불충분했고, 구현이 더한 두 가지(읽기 어휘, 읽기/쓰기 분리)는 **필요한 것이었다.**
필요 없는 것을 더한 것이 아니다. 되돌린 뒤 `band_test.go` = `531a1c47…6f33c92`.

### A3. 우회 시도 — **구멍이 하나 남아 있다**

| # | 변이 | 결과 |
|---|---|---|
| A3-1 | `b := v.ExtendedBand` 후 `b.Crossed("6")` | **RED** — 두 규칙 다 걸림 |
| A3-2 | 판정을 반환하지 않는 헬퍼가 밴드를 읽고 `bool`을 돌려줌 | **GREEN — 구멍** |
| A3-2b | 같은 것의 **최소 형태**(대입 재구성조차 없음) | **GREEN — 구멍** |
| A3-3a | 다른 이름의 지역 구조체에 넣고 `r.e.Crossed("6")` | **RED** — `Crossed`가 잡음 |
| A3-3b | 같은 구조체에 메서드를 달고 `r.hot()` | **GREEN — A3-2와 같은 구멍** |

A3-2b가 가장 심각하다. 밴드 대입 줄을 **전혀 건드리지 않고** 두 줄만 더하면 된다.

```go
// watch.go — 판정 타입을 반환하지 않으므로 검사 대상이 아니다
func chaseWorthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }

// assessOne 안 — 금지 식별자를 하나도 쓰지 않는다
	if chaseWorthy(v) {
		v.Chase.Extended = RaisedVeto()
	}
```

```
$ rtk proxy go test ./internal/candidate/ -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v
    band_test.go:394: checked 12 verdict-producing functions
--- PASS
$ rtk proxy go vet ./...          → 출력 없음
$ rtk proxy go test ./...         → ok 51 / 51, 실패 0
```

**승인된 적 없는 숫자 6이 veto를 결정하고, 51개 패키지 전부가 초록이다.** 2026-07-28
리뷰가 찾은 것과 같은 상태이며, 다른 점은 한 줄이 아니라 두 줄이라는 것뿐이다.
되돌린 뒤 `watch.go` = `1aa35693…7b0106`.

### A4. §0.2 하한은 의미 있는 수다

`checked`의 현재 실측값이 **정확히 12**이고 `verdictProducers = 12`이므로 하한은 빡빡하다.
검증: `verdicts`에서 `Verdict`를 빼면

```
band_test.go:389: checked 10 verdict-producing functions, want at least 12; …
--- FAIL
```

10 → Fatalf. 지금 값보다 낮게 잡혀서 아무것도 못 잡는 상태가 **아니다.**

### A5. §0.5 포인터 메서드 집합 — 직접 재확인

`func (b *ShadowBand) Dangerous() bool { return b.Measured }` 투입 →

```
band_test.go:291: *candidate.ShadowBand has a Dangerous method; …
--- FAIL
```

구현자 보고와 동일. 되돌린 뒤 `band.go` = `88174be2…0dec62`.

---

## B. §2.7 실측 — **패키지 자신의 코드로** 재측정

구현자는 SQL 재구성으로 쟀다. 나는 임시 probe 테스트를 만들어 `Open` → `Assess` →
`TallyVerdicts`를 **실제 코드 경로로** 돌렸다(store 복사본, 각 스캔 instant를 `at`으로).

```
@2026-07-28T13:23:09Z KR extended: total=156 measured=131 collapsed=true
    crossed=map[0:131 2:0 4:0 6:0 8:0 10:0 20:0 30:0 50:0 100:0]
    quantiles=[{min 0} {p25 0} {median 0} {p75 0} {max 0}] base=131
    notMeasured=map[NO_PRICE:25]
    alarm="ALARM extended: all 131 measured record(s) produced the same crossings, …"
@2026-07-28T13:23:31Z US extended: total=230 measured=61  collapsed=false
    crossed=map[0:60 …나머지 전부 0]  quantiles=[{min -0.14684287812} {p25 0} {median 0} {p75 0} {max 0}]
@2026-07-28T14:53:19Z US extended: total=169 measured=169 collapsed=false
    crossed=map[0:165 …나머지 전부 0] quantiles=[{min -0.024308129854} … {max 0.083682008368}]
seen_late: 세 instant · 두 시장 전부 measured=0, collapsed=false, alarm=""
```

저장소 구조도 재확인했다(복사본, 읽기 전용 SQL).

```
서로 다른 observed_at 3개:  13:23:09 KR 230행 · 13:23:31 US 200행 · 14:53:19 US 200행
first_price_at:  NULL 25 · 13:23:09 131 · 13:23:31 61 · 14:53:19 169
```

**확인된 것 (구현자 보고와 일치)**

1. **KR 131건은 전부 정확히 0이고, 새 눈금도 KR에서 붕괴한다.** `Collapsed()`가 참이고
   경보가 뜬다. 분위수 다섯 개가 전부 `0`이다. 원인도 그대로다 — KR은 한 번만 스캔됐고
   기준가와 최신 가격이 같은 판독이다.
2. **이 저장소의 어떤 스캔에서도 밴드 `2` 이상을 넘은 기록이 하나도 없다.** `2·4·6·8`은
   검증도 반증도 되지 않았다. D13의 "잠정"은 여전히 잠정이고, **band.go 주석·review.md·
   issues.md I5 어디에서도 확정된 것처럼 쓰이지 않았다** — 세 곳 모두 잠정이라고 적는다.
   찾아봤고 없다.
3. **이 change가 산 것은 분포가 아니라 경보다.** 그 판단은 review.md와 issues.md I5에
   정직하게 적혀 있다. 숨기지 않았다.
4. 부수적이지만 중요한 확인: **경보는 상시 점등이 아니다.** 같은 저장소에서 KR extended만
   울리고 US extended와 seen_late는 울리지 않는다.

**정정 필요 (P2)** — review.md §2.7과 issues.md I5의 **US 수치는 어떤 표면도 낸 적 없는 값**이다.
`US n=230 … 밴드 0이 225 대 5로 갈랐다`는 **두 개의 다른 스캔을 합친 것**이다:
13:23 스캔이 measured 61(밴드 0을 60), 14:53 스캔이 measured 169(밴드 0을 165)이고,
61+169=230 · 60+165=225 · 1+4=5다. 인용된 `min=-0.146843`은 앞 스캔에서,
`max=+0.083682`는 뒤 스캔에서 나온다. **결론은 그대로 산다**(밴드 0만 갈랐고 나머지 아홉은
아무것도 못 갈랐다). 틀린 것은 근거의 형태다 — `retire-gainers-source`에서 네 번 나온
"결론은 맞고 근거는 틀림"의 다섯 번째다.

부수: KR의 미측정 사유는 `NO_PRICE 25`이고 `NO_OBSERVATIONS 25`가 아니다(§2.6 샘플과 다름).

---

## C. 붕괴 경보

**Scenario와 일치하는가 — 그렇다.** `Collapsed()`는 `Measured`와 `Crossed`만 읽는다.
스펙의 *"집계만으로 계산 가능하며 눈금의 의도를 알 필요가 없다"*를 충족한다.
논리도 건전하다: 밴드마다 건수가 `0` 또는 `Measured`라는 것은 모든 측정 기록의 교차
집합이 동일하다는 것과 **동치**다(건수가 `Measured`인 밴드들이 곧 그 공통 집합이다).

변이 확인: `Collapsed()`를 항상 false로 만들면

```
bandscale_test.go:317: three records that all crossed exactly band 0 are not reported as collapsed …
bandscale_test.go:322: a collapsed tally carries no sentence; …
candidatebands_test.go:221: 131 measured records that all produced the same crossings render as an ordinary row
```

두 표면 모두 RED. 공허하지 않다.

### C1. "측정 1건도 경보" — 내 판단: **스펙 문구를 유지한다**

근거 넷.

1. n=1에서 경보 문장은 **참이다.** 한 건으로는 임계를 승인할 수 없고, 경보가 말하는 것이
   정확히 그것이다.
2. 요구사항이 막는 실패는 "그 줄이 정상으로 보인다"는 것이다. n=1의 줄은 정상으로 보이고
   똑같이 쓸모없다.
3. 하한을 2로 올리면 **근거 없는 숫자를 하나 더 넣는 것**이다. 이 change의 주제가 바로
   근거 없는 숫자를 계측기에 넣지 않는 것이다.
4. 상시 점등 위험은 제한적이다. 진짜 흔한 퇴화 상태인 `measured == 0`은 이미 제외돼
   있고(I7), n=1은 후보가 하나뿐인 스캔에서만 나온다. B의 실측에서 이 저장소는 n=131/61/169다.

다만 **문구 하나는 손볼 값어치가 있다**: n=1일 때 경보는 *"this scale resolved nothing"*이라며
**눈금**을 지목하는데 실제 원인은 **표본 수**다. 사실관계는 틀리지 않았지만 원인을 잘못
가리킨다. 한 줄 분기로 고칠 수 있고 requirement 수정이 아니다. Manager 판단 사항으로 남긴다.

### C2. **P1 — 경보가 콘솔 `/signals`에는 없다. SHALL이 살아 있는 화면에 대해 거짓이다**

`internal/console/signals.go:455`의 `signalsBandTally`에는 `Alarm`도 `Quantiles`도 없고,
`signalsBandTalliesFrom`(:803)은 **같은 `candidate.BandTally`**에서 만든다. 그래서 붕괴한
KR 집계가 그 화면에서는 평범한 행으로 렌더된다 — 요구사항이 금지하는 바로 그 표시다.

구현자는 이것을 I8로 공개하고 tasks §2.7의 문구("리포트에 더한다")를 근거로 범위 밖에
뒀다. 그러나 spec delta의 SHALL은 **표면을 지정하지 않는다**: *"보고 표면은 그것을 정상
수치로 표시하지 않고 경보로 표시해야 한다."* 이대로 아카이브하면 **승인된 요구사항이
살아 있는 화면에 대해 거짓**이 되고, 그것은 D10이 눈금 순서에 대해 이름 붙인 권위
오염과 같은 것이다. 같은 change 안에서 D10을 지키고 D10을 어기는 셈이다.

두 길 중 하나가 필요하다. ① `signalsBandTally`에 경보를 배선한다(표시 전용, 등급 Small,
이 change 안에서 끝난다) ② spec 문구를 *"스캔 리포트"*로 좁히고, 콘솔은 후속 change의
의무로 issues에 남긴다. **①을 권한다** — 요구사항의 근거가 *"사람이 알아차리는 것에
맡겨서는 안 된다"*이고, 운영자가 들여다보는 화면이 콘솔이다.

### C3. P2 — 밴드가 없는 사유에서 경보가 **공허하게 참**이 된다

`TallyBands`는 `Crossed`를 `BandsFor(code)`로만 seed한다(band.go:508-513). 밴드가 nil이면
맵이 비고, `Collapsed()`의 루프가 한 번도 돌지 않아 `Measured >= 1`이기만 하면 **참**이다.

```
$ TallyBands(VetoNearHigh, 2건의 measured 기록)
bands=[] measured=2 crossed=map[] collapsed=true
alarm="ALARM near_high: all 2 measured record(s) produced the same crossings, …"
```

이것은 가상의 상태가 아니다. **issues.md I3이 `extended`에 대해 의무로 적은 바로 그
상태**다(임계 승인 시 1단계 = `BandsFor(VetoExtended)`를 nil로). `TallyVerdicts`는 두
코드의 집계를 무조건 만들고(watch.go:744-747) `writeCandidateTable`은 두 코드를 무조건
렌더한다(candidate.go:918). 즉 **I3을 실행하는 날 extended 경보가 영구 점등된다** —
I7이 `measured == 0`에 대해 피했다고 쓴 바로 그 실패가 다른 문으로 들어온다.

한 줄로 막힌다: `Collapsed()` 첫머리에 `if len(t.Crossed) == 0 { return false }`.

---

## D. 나머지

### §1.4 — RED 확인, 그리고 `extended`는 여전히 veto하지 않는다

`MeasureExtendedBand`가 `th.ExtendedGainPct`를 눈금 앞에 붙이게 변이 →

```
bandscale_test.go:379: MeasureExtendedBand mentions ExtendedGainPct … (band.go:294:8, 295:30)
bandscale_test.go:402: the band changed when ExtendedGainPct went from "" to "6": …
```

**AST 절반과 동작 절반이 모두 걸린다.** 되돌린 뒤 `band.go` = `88174be2…0dec62`.

**veto 여부 직접 확인**: `ExtendedGainPct`는 비테스트 코드 어디에서도 대입되지 않는다
(`grep -rn ExtendedGainPct --include=*.go` → 선언·doc·`AssessExtended`의 두 읽기뿐).
따라서 `AssessExtended`는 `thresholdReason` 분기에서 항상 `UnmeasuredVeto`를 돌려준다.
**`extended`는 이 change 이후에도 veto하지 않는다.**

### §2.1·2.2·2.3 — 눈금 변이 셋 전부 RED

| 변이 | 결과 |
|---|---|
| 눈금에서 `"0"` 제거 | **RED** — `-20% and 0% produced the same record ([])` + `did not cross band 0` |
| 눈금에서 `"4"` 제거 | **RED** — `+3% and +5% both recorded [0 2]` (인접한 한 쌍만) |
| 눈금에서 `"50"` 제거 | **RED** — `the scale ends [8 10 20 30 100], want it to end [10 20 30 50 100]` |

세 번 모두 되돌린 뒤 `band.go` = `88174be2…0dec62`.

추가 변이 두 개로 §2.7 테스트의 공허성도 봤다: `bandQuantiles`의 nearest rank를
ceil→floor로 바꾸면 RED, `sort.Slice`를 빼면 RED. 공허하지 않다.

### §2.5 — 허용 파일 목록은 **넓혀지지 않았다**

`bandIdentityFiles`(bandscale_test.go:411-416)는 tasks §2.5가 적은 넷 그대로다.
`cmd/tossctl/candidatebands.go`는 **목록에 없고**, 파일을 읽어 확인한 결과
`ShadowBand`/`BandTally`/`BandCrossing`을 하나도 이름하지 않는다 — 자기 소문자 타입
(`bandCount`·`bandQuantile`·`shadowBandReport`)만 쓴다. 넓힐 필요가 없어서 안 넓혔다.

### §3.4 — 구현자의 정정이 맞다

직접 읽었다. `formatDecimal`(metrics.go:969-976)의 doc은 *"truncated towards zero"*이고
부호 전제가 **없다.** 전제는 `truncateAt`(metrics.go:1018-1019)의
*"which for the non-negative values this file produces is the floor"*에 있다.
**tasks §3.4가 위치를 틀렸고 구현자의 I4가 옳다.**

**다만 I4가 이웃 하나를 놓쳤다 (P2)**: `distancePct`(level.go:552)의
*"This is **the one value in the package** that can be negative"* — `gainPct`도 음수이고
`ShadowBand.Value`가 그것을 렌더한다. 이 문장은 **거짓**이고, I4가 선례로 인용한 문단
바로 위에 있다. 이 change가 만든 것은 아니지만 I4의 목적이
*"다음 사람이 그 문장을 근거로 재사용하지 않도록"*이었으므로 같은 항목에 적혀야 한다.

### §2.6 `wrapCounts` 80칼럼 경계 — **두 개의 결함, 둘 다 P2**

경계값을 직접 넣었다.

```
[real KR row]                line 0 width=74 "  crossed      0 131 · 2 0 · … · 50 0 ·"
                             line 1 width=20 "               100 0"
[two parts, first line 79]   line 0 width=81  <<< OVER reportWidth
[one oversized part]         line 0 width=105 <<< OVER reportWidth
```

**① 단위 혼용.** `sep = " · "`와 `mark = " ·"`에는 U+00B7이 들어 있어 **바이트 길이가 4와 3**인데,
코드는 `len(sep)`·`len(mark)`(바이트)를 `utf8.RuneCountInString`(룬)과 섞어 비교한다
(candidatebands.go:134, 138). 방향은 보수적이라 이 원인으로 넘치지는 않지만, 함수 doc의
*"That marker is **two columns** and they are budgeted here"*는 **거짓이다 — 두 칼럼이 아니라
세 바이트를 예산에 넣는다.** 실측 결과: 실제 KR 행은 접지 않으면 **정확히 80칼럼**이라
접힐 필요가 없는데 접힌다. issues.md I11이 "표시 두 칸을 예산에 넣어 고쳤다"고 쓴 그 수정이
반대 방향으로 한 칸 어긋나 있다.

**② 첫 줄과 이어지는 줄은 폭 검사를 받지 않는다.** `line = label + parts[0]`(:130)과
`line = indent + part`(:143)는 예산을 통과하지 않는다. 위 실측의 81·105칼럼이 그것이다.
`TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth`의 doc은 *"the properties that make it
safe for **any** tally"*라고 쓰지만 다섯 행 전부 짧은 part만 쓴다. 오늘의 데이터로는 도달
불가(가장 긴 part가 분위수의 `p25 -33.333333333333` ≈ 19칸)여서 P2이지, 실동작 결함은 아니다.

---

## 스스로 찾은 것

| # | 등급 | 위치 | 내용 |
|---|---|---|---|
| R1 | **P1** | `internal/candidate/band_test.go:327-395` | 판정을 반환하지 않는 헬퍼를 한 단계 끼우면 밴드→판정 경로가 다시 열린다(A3-2b). 51개 패키지 전부 초록 |
| R2 | **P1** | `internal/console/signals.go:455,803` | 붕괴 경보가 콘솔 `/signals`에 없다. 승인된 SHALL이 살아 있는 화면에 대해 거짓인 채로 아카이브된다 |
| R3 | P2 | `internal/candidate/band.go:466-476` | 밴드가 없는 사유에서 `Collapsed()`가 공허하게 참. I3을 실행하는 날 extended 경보가 영구 점등 |
| R4 | P2 | `cmd/tossctl/candidatebands.go:122-146` | `wrapCounts` 단위 혼용(바이트 vs 룬) + 첫/이어지는 줄 폭 미검사. doc의 "two columns"가 거짓 |
| R5 | P2 | `review.md` §2.7 · `issues.md` I5 | US 수치가 두 스캔을 합친 값이고 어떤 표면도 낸 적이 없다. 결론은 유효, 근거는 부정확 |
| R6 | P2 | `review.md` §2.6 · `issues.md` I9 · `candidatebands.go:3-26` | 렌더 예시가 합성 fixture인데 실측처럼 읽힌다. 실제 행은 80칼럼(84 아님), 분위수는 전부 `0`(`median 0.91`·`max 41.07` 아님), 코드 주석의 "92 against the live tally"는 이 저장소로 재현 불가 |
| R7 | P2 | `cmd/tossctl/candidate.go:973` | `orderedCounts`가 **호출자 0**이 됐다(두 호출부가 `wrapCounts`로 이동). `make lint`는 `go vet ./...`뿐이라 잡지 않는다. FLM은 "본문이 orderedCountParts로"라고만 적고 죽었다는 사실은 적지 않는다 |
| R8 | P2 | `internal/candidate/level.go:552` | `distancePct`의 "the one value in the package that can be negative"가 거짓. I4가 놓친 이웃 |

### R1 실패 시나리오 (구체)

**입력**: 후속 change가 `internal/candidate`에 다음 두 조각을 넣는다.

```go
func chaseWorthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }   // 판정 타입 아님
// assessOne 안
if chaseWorthy(v) { v.Chase.Extended = RaisedVeto() }
```

**잘못된 결과**: 기준가 대비 6% 이상 오른 모든 후보의 `Chase.Extended`가 raised veto가 되어
스캔 리포트와 콘솔 패널에서 후보가 걸러진다. 근거는 **승인된 적 없는 숫자 6**이다.
`go vet` 통과, `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand` 통과,
`go test ./...` 51/51 통과, `make gate` 통과. **아무것도 실패하지 않으면서 시스템이 뜻을 바꾼다.**

**최소 보강안**: 검사를 **전이적**으로 만든다 — 패키지 안에서 밴드 타입·생성자·읽기 어휘를
쓰거나 밴드 필드를 읽는 함수 집합 `T`를 먼저 고정점까지 모으고(호출 그래프를 AST로 1패스),
판정 생산 함수가 `T`의 원소를 호출하면 실패시킨다. 또는 더 간단히, 판정 생산 함수가
**파라미터·리시버 타입에 `ShadowBand`를 가진 패키지 함수**를 호출하면 실패시킨다.
어느 쪽이든 기존 walk 위에 15줄 안팎이고, A3-2·A3-2b·A3-3b를 전부 RED로 만든다.

### 찾아봤지만 없던 것

- **변이 없이도 통과하는 새 테스트**: 없다. `Collapsed()`(→2개 RED), `bandQuantiles`의
  rank·sort(→각각 RED), 눈금 `0`·`4`·`50`(→각각 RED), `verdictProducers` 하한(→Fatalf),
  §0.5 포인터 메서드(→RED), §1.4 임계 읽기(→양쪽 RED)를 전부 변이로 확인했다.
- **공허 방지 장치**: `TestAShadowBandIsNotPersistedAnywhere`에 빈 walk `t.Fatal` + 필수
  파일 2건, `TestTheShadowBandRowsStayInsideEightyColumns`에 `folded == 0` 실패,
  `TestMeasureExtendedBandNeverReadsTheVetoThreshold`에 `!found` `t.Fatal`이 있다.
  **가드를 검증하는 가드가 스스로 공허한 경우는 §0.2 하한 하나가 후보였는데,
  `checked`가 정확히 12라서 빡빡하다.**
- **`6`이 확정된 것처럼 쓰인 곳**: 없다. band.go 주석·D13·I5·review.md 모두 잠정이라고 적는다.
- **`assessOne`의 FLM이 diff와 어긋나는 곳**: 없다. B1~B3이 base와 동일하고 밴드 두 줄이
  `shadowBandsFor`로 나간 것이 그대로 적혀 있다. "Safe edit boundary" 항목도 실제 가드와 맞다.
- **`seen_late` 눈금 변경**: 없다(D14 준수). `SeenLatePercentileBands` 무변경 확인.
- **High-risk 경로 접촉**: 없다. `internal/risk`·`internal/execgw`·`internal/exitpolicy`·
  `internal/trading` 무변경, 주문·손절·사이징·원장·인증·체결 경로 무접촉.

---

## 검증 명령과 결과 (복구 후 트리)

| 명령 | 결과 |
|---|---|
| `sha256sum -c` (9개 파일) | **전부 OK** — 리뷰 전 상태와 동일 |
| `rtk proxy go test ./...` | **ok 51 / 51**, 실패 0 |
| `rtk proxy go vet ./...` | 출력 없음 |
| `$(go env GOROOT)/bin/gofmt -l .` | 출력 없음 |
| `make lint` | `go vet ./...` — 통과 (unused 린터는 없다 → R7이 잡히지 않는 이유) |
| `cat base-commit.txt` | `ad16f56d7db04871699e060f120f249518fbee77` = `git log -1` |
| PM 등록 | `_registry.yaml:30`, `test_generate_master_tracker.py:36` 양쪽 존재 |
| 미완료 태스크 | `4.4` 한 건뿐 |

---

## §4.6 재실행 — `make sdd-sync` → `make sdd-check` → `make gate` (독립 리뷰 후)

```
$ make sdd-sync
+ codegraph sync .
+ codegraphcontext update . --quiet
+ gbrain sync --source tossos-feat-p0-foundation --strategy code --no-embed --yes
[sdd-sync] all indexes current

$ make sdd-check                                                    exit=0
[agent-config] Claude/Codex safety bootstrap and workflow routing are synchronized
python3 scripts/memory_index.py check
[index-freshness] CodeGraph hard-evidence index matches the worktree
[pm] hierarchy and generated trackers are current
make sdd-test → scripts 15 OK · logic-map 15 OK · sdd 19 OK · sdd-history 15 OK · pm 6 OK
go test ./tools/logic-map → ok

$ make gate CHANGE=refine-extended-shadow-bands
GATE: refine-extended-shadow-bands
repo: /mnt/D/project/axipient/TossOS

==> 1/8 tasks.md 확인
OK: openspec/changes/refine-extended-shadow-bands/tasks.md

==> 2/8 미완료 태스크 확인
미완료 태스크 1 건:
133:- [ ] 4.4 독립 리뷰: 구현과 **분리된 컨텍스트**가 diff와 테스트를 재검증하고 review.md에

GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다
make: *** [Makefile:75: gate] Error 1
```

**§4.4 체크박스는 리뷰어가 채우지 않는다.** 아래 판정이 "랜딩 불가"이므로 열어 두는 것이
정확한 상태다. R1·R2가 처리되고 Manager가 확인한 뒤에 채운다.

게이트가 2/8에서 멈추므로 3~8단계는 **손으로 각각 돌려 확인했다. 걸린 것은 없다.**

```
3/8 review.md 존재                   OK (이 파일)
4/8 Function Logic Map               OK — [logic-map] evidence complete or diff-proven exempt (exit=0)
5/8 make sdd-check                   OK — exit=0
6/8 make test                        OK — exit=0 (go test ./... ok 51/51)
7/8 make vet                         OK — exit=0
8/8 make validate                    OK — Totals: 25 passed, 0 failed (25 items)
openspec validate --strict           Change 'refine-extended-shadow-bands' is valid
```

**미완료로 남은 것은 §4.4 하나뿐이고, 다른 단계에서 걸린 것은 없다.**

> 이 절을 쓰면서 `.md`를 편집했으므로 CodeGraph fingerprint가 다시 stale이다.
> 리뷰 종료 직후 `make sdd-sync`를 한 번 더 돌려 되살렸다. 다음 사람이 `make gate`를
> 돌릴 때는 그 사이에 편집이 없었다면 그대로 쓸 수 있다.

---

## 최종 판정

**랜딩 불가 — 사유 둘.**

1. **R1 (P1)** — 이 change의 존재 이유인 "밴드→판정 경로를 구조로 막는다"가 아직 닫히지
   않았다. 판정을 반환하지 않는 헬퍼를 한 단계 끼운 **두 줄짜리 변이**가 가드·`go vet`·
   51개 패키지를 전부 통과한다(A3-2b, 직접 실측). 2026-07-28에 발견된 상태와 형태가 같고
   길이만 한 줄 늘었다. spec delta는 *"도달 불가는 판정을 조립하는 함수까지 포함해야
   한다(SHALL)"*를 **도달 가능성**으로 쓰는데, 구현은 **직접 읽기 한 홉**만 막는다.
   이대로 아카이브하면 D10이 눈금 순서에 대해 경고한 권위 오염이 §0의 요구사항에서
   그대로 일어난다.
2. **R2 (P1)** — 붕괴 경보가 콘솔 `/signals`에 없다. 같은 `BandTally`에서 만드는 화면이
   붕괴한 KR 집계를 평범한 행으로 렌더한다. 승인된 SHALL이 살아 있는 표면에 대해 거짓인
   채로 아카이브된다.

**이 둘을 빼면 나머지는 좋다.** §0.4 뒷문 차단은 직접 넣은 변이로 RED를 확인했고,
tasks §0.1이 불충분했다는 구현자의 반례는 **재유도해 재현했다**(`checked=12 failures=0`).
§1.4·§2.1·§2.2·§2.3·§0.5의 변이가 전부 RED이고, 새 테스트 중 변이 없이 통과하는 것은
찾지 못했다. §2.5의 허용 파일 목록은 넓혀지지 않았다. §2.7의 핵심 주장(KR 131건 전부 0,
새 눈금도 붕괴, 산 것은 분포가 아니라 경보)은 **패키지 자신의 코드로 재측정해 확인**했고
review.md·issues.md에 정직하게 적혀 있다. `2·4·6·8`은 세 문서 모두에서 잠정으로 남아 있다.

**남은 P2 여섯 건(R3~R8)은 랜딩을 막지 않는다.** 다만 R3은 issues I3이 의무로 적은 후속
작업이 실행되는 순간 경보를 영구 점등시키므로, 한 줄 방어를 이 change에서 같이 넣는 것이
싸다. R5·R6은 문서 정정이고, 이 두 change에서 반복된 *"결론은 맞고 근거는 틀림"*의
다섯·여섯 번째다.

**C1(측정 1건도 경보)에 대한 판단**: 스펙 문구를 유지하는 데 찬성한다. 근거는 위 C1.
경보 문장이 n=1에서 눈금을 지목하는 것만 한 줄로 손보면 좋겠고, 그것은 requirement
수정이 아니다.



---

# §5 수정 (구현) — 독립 리뷰가 낸 P1 둘

- 날짜: 2026-07-29
- 역할: Teammate(구현). 범위는 **§5.1~§5.4**. **§5.5(2차 독립 리뷰)는 이 기록에 포함되지
  않는다** — 구현자가 자기 구현을 독립 검증할 수 없다.
- base-commit: `cat base-commit.txt` → `ad16f56d7db04871699e060f120f249518fbee77`,
  `git log --oneline -1` → `ad16f56`. **일치한다.** 재고정하지 않았다.
- 방법: 변이는 전부 **편집으로 넣고 편집으로 되돌렸다.**
  `git checkout --`·`git restore`·`git stash`는 **한 번도 쓰지 않았다.** 착수 시점에 대상
  9개 파일의 sha256을 기록하고 대조했다.

## §4.4 체크박스에 대하여

독립 리뷰는 **실제로 수행됐고** 그 기록이 이 파일의 "독립 리뷰 (§4.4)" 절이다.
판정은 **"랜딩 불가"**였고 사유가 R1·R2였다. 이 §5가 그 둘을 닫으므로 4.4를 체크한다.
체크가 뜻하는 것은 *"독립 리뷰가 수행되고 기록됐다"*이지 *"통과했다"*가 아니다 —
그 판정과 사유는 위 절에 그대로 남아 있고 지우지 않았다.

---

## R1 재현 — 고치기 전에 그 두 줄이 통과한다

가드를 손대기 **전에** `watch.go`에 편집으로 넣었다.

```go
// MUTATION R1 — remove after observing the result.
func worthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }

// assessOne 안, 밴드 대입 바로 아래
if worthy(v) {
	v.Chase.Extended = RaisedVeto()
}
```

```
$ go test ./internal/candidate/ -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v
    band_test.go:394: checked 12 verdict-producing functions
--- PASS
$ go vet ./...      → 출력 없음
$ go test ./...     → ok 51 / 51
```

**결함 확인.** Manager·리뷰어 보고와 정확히 같다. 승인된 적 없는 숫자 `6`이 veto를
결정하는데 검사·vet·51개 패키지가 전부 초록이다.

이 상태를 **유지한 채** 가드를 고쳤고, 새 가드가 RED가 되는 것을 본 뒤 변이를 되돌렸다.

```
$ go test ./internal/candidate/ -run TestNoFunctionTurnsAShadowRecordIntoSomethingElse -v
    bandguard_test.go:…: watch.go:worthy takes a shadow record apart (reads Crossed out of
    one) and returns something that is not a record. …
--- FAIL
```

되돌린 뒤 `sha256sum internal/candidate/watch.go` = `1aa35693…7b0106`(기준과 동일).

---

## 고른 가드 형태와 그 이유

**Manager 제안(국소 규칙)을 골랐다.** `internal/candidate/bandguard_test.go`의
`TestNoFunctionTurnsAShadowRecordIntoSomethingElse`.

> **그림자 기록을 뜯어보는 함수는 기록(`ShadowBand`·`BandCrossing`·`BandTally`·
> `BandQuantile`)을 돌려주거나, 아무것도 돌려주지 않거나, 기록 타입의 메서드여야 한다.**

"뜯어본다"는 두 가지다.

1. **기록 값 표현식에 selector를 적용하는 것.** `Crossed`만이 아니라 `Measured`·`Value`도
   같은 행위다 — 리뷰어의 우회는 `Crossed`만 아는 검사라면 살아남았을 형태였다.
2. **기록을 받아 기록이 아닌 것을 돌려주는 호출에 기록을 건네는 것.**
   `fmt.Sprintf("%v", band)`와 `identity(band)`는 어떤 필드도 읽지 않으면서 기록을
   기록 아닌 값으로 되돌려 준다.

허용되는 것은 **나르기**다: 기록을 통째로 인자로 넘기거나, 슬라이스에 append하거나,
필드에 담거나, `Verdict`의 밴드 자리에 대입하는 것. 이것이 조립이고 spec이 명시적으로
보호하는 형태다.

### 왜 리뷰어의 전이적 검사를 고르지 않았는가

리뷰어 제안은 밴드 어휘를 쓰는 함수 집합 `T`를 고정점까지 모으고, 판정 생산 함수가 `T`의
원소를 호출하면 실패시키는 것이다. 실제로 만들어 보면 `assessOne → shadowBandsFor →
MeasureExtendedBand`가 곧바로 오탐이다. 세 홉 전부가 정당하고, 이 오탐을 피하는 방법은
**결국 예외 목록**이다 — 그리고 예외 목록은 다음 우회가 들어오는 문이다.

국소 규칙에는 **예외가 하나도 필요하지 않았다.** 검사가 지목한 "기록을 읽는 함수"는
정확히 일곱이고 전부 자연히 허용된다.

```
record readers: [BandTally.Collapsed BandTally.CollapsedAlarm MeasureExtendedBand
                 MeasureSeenLateBand ShadowBand.Crossed ShadowBand.Reason TallyBands]
```

`TallyVerdicts`·`assessOne`·`assessInto`·`Cycle`은 **목록에 없다** — 기록을 통째로 나를 뿐
뜯어보지 않기 때문이다. 허용목록도, 호출 그래프도, `shadowBandsFor`를 위한 특례도 없다.

부수 효과 하나가 중요하다: 규칙이 **어휘가 아니라 구조**이므로, 가속 계열
(`Acceleration.Crossed`·`TallyCrossings`·`Accelerate`)이 같은 단어 `Crossed`/`Crossings`를
쓰는데도 이 검사에 걸리지 않는다. 어휘 기반으로 패키지 전체를 훑었다면 그 넷이 전부
오탐이었을 것이다.

### "기록을 돌려주면 되지 않느냐"가 빠져나갈 길이 아닌 이유

기록을 읽고 **기록을 돌려주는** 헬퍼는 어느 기록을 돌려줄지로 답을 인코딩할 수 있다.
그 답은 여전히 기록이고, 거기서 판정을 꺼내려면 그것을 뜯어봐야 한다 — 그러면 그 함수가
같은 규칙 아래 들어온다. 사슬은 매번 기록으로 끝나거나 여기서 실패한다.
**M5가 그 형태이고 RED다**(아래 표).

### 두 가드는 독립이다

기존 `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`(어휘 + 밴드 필드 읽기)를
**그대로 두었다.** 아래 표에서 보듯 어느 한쪽이 놓치는 변이를 다른 쪽이 잡는 경우가
양방향으로 있다.

---

## 우회 시도 — 넣기 / 결과 / 되돌리기

변이는 `watch.go` 편집(M1·M2·M5)과 임시 probe 파일
(`internal/candidate/zz_bypass_probe.go`, 관찰 후 삭제)로 넣었다.
`판정` 열은 새 가드, `어휘` 열은 기존 가드다.

| # | 변이 | 어휘 가드 | 새 가드 | 출처 |
|---|---|---|---|---|
| M1 | `worthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }` + `assessOne` | **PASS(못 잡음)** | **RED** | tasks §5.1 |
| M2 | `take(v Verdict) ShadowBand { return v.ExtendedBand }` + `assessOne`에서 `take(v).Crossed("6")` | **RED** | **RED** | tasks §5.1 |
| M3 | 기록을 구조체 필드에 담고(`type m3peek struct{ e ShadowBand }`) 그 메서드가 읽음 | PASS | **RED** | tasks §5.1 |
| M4 | 금지 어휘를 **하나도** 쓰지 않음 — `b := v.ExtendedBand; return b.Measured && b.Value >= "6"` | PASS | **RED** | 구현자 |
| M5 | 헬퍼가 **기록을 반환**하고 답을 "어느 기록인가"로 인코딩; `assessOne`이 `hot(v).Measured` | **PASS(못 잡음)** | **RED** | 구현자 |
| M6 | 판정을 아예 건드리지 않고 헬퍼가 자기 밴드를 새로 측정: `MeasureExtendedBand(…).Crossed("6")` | PASS | **RED** | 구현자 |
| M7 | 패키지 수준 변수에 stash: `var m7last ShadowBand` + 다른 함수가 읽음 | PASS | **RED** | 구현자 |
| M8 | **검사 무장 해제** — `type m8decoy struct{ ExtendedBand int }`로 필드 이름을 모호하게 만듦 | PASS | **RED** (`pinnedRecordFields` Fatalf) | 구현자 |
| M9 | 기록을 렌더하고 렌더 결과를 읽음: `fmt.Sprintf("%v", band)` + `strings.Contains` | PASS | **RED** (2차 보강 후) | 구현자 |
| M10 | 빈 인터페이스 세탁, comma-ok: `b, _ := box(v).(ShadowBand)` | PASS | **RED** (2차 보강 후) | 구현자 |
| M10b | 같은 것, 단일값 assertion | PASS | **RED** | 구현자 |
| M11 | 채널 세탁: `b := <-ch` (`ch chan ShadowBand`) | PASS | **RED** | 구현자 |
| M12 | 클로저에 담아 클로저를 반환 | PASS | **RED** | 구현자 |
| M13 | 제네릭 항등 함수 세탁: `m13pass(v.ExtendedBand).Crossed("6")` | PASS | **RED** (2차 보강 후) | 구현자 |
| M14 | 메서드 값: `f := v.ExtendedBand.Crossed; return f("6")` | PASS | **RED** | 구현자 |

**14개 시도 중 어휘 가드가 잡은 것은 하나(M2)뿐이다.** 그것이 R1의 크기다.

**두 번의 보강이 필요했다** — 처음 형태는 M10(comma-ok)과 M9·M13(호출을 통한 세탁)을
놓쳤다. 놓친 것을 보고 규칙을 넓혔다:

- 결과 위치별 기록 여부(`recordResults`)를 재서 `b, _ := x.(ShadowBand)`가 이름을 묶게 했다.
  `TallyVerdicts(…)`의 앞 두 결과는 기록이 아니므로 묶지 않는다 — 오탐을 만들지 않는다.
- 기록을 인자로 받는 호출 중, 결과가 전부 기록이거나 결과가 없는 호출이 **아닌** 것을
  변환으로 센다. `append`·`len`·`cap`·`copy`·`make`·`new`·`delete`는 통과시킨다.

### 복구 대조

```
$ sha256sum internal/candidate/watch.go
1aa356933d8e9ad3ea2b98a4ded5d25972eb163c16897edff4240584ad7b0106   ← 착수 시 기록과 동일
$ git status --short          → probe 파일 없음, §5 시작 시점 + 의도한 변경만
```

`internal/candidate/zz_bypass_probe.go`와 `internal/console/zz_render_probe_test.go`는
**삭제했다.** `band.go`는 §5.3의 한 줄 때문에 의도적으로 달라졌다(아래 R3).

### 못 막는 것 — 숨기지 않는다

전문은 issues.md **I20**. 넷이다.

1. **다른 패키지.** `internal/candidate`만 파싱한다. `RaisedVeto()`는 exported다.
   현재의 방어는 `TestOnlyTheListedFilesCanNameTheChaseVerdict`가 누가 chase를 이름할 수
   있는지를 따로 제한하는 것뿐이다 — 실제로 이번에 새 콘솔 테스트가 그 목록에 걸렸고,
   목록을 넓히는 대신 테스트에서 verdict 참조를 뺐다.
2. **테스트 파일.** 두 가드 모두 `_test.go`를 walk에서 제외한다.
3. **모호한 필드 이름.** 기록 타입 필드 이름이 이 패키지 어딘가에서 비-기록 타입으로도
   선언돼 있으면 그 이름은 빠진다(`Crossings`가 실제로 그렇다). 이것이 가속 계열을 빼
   주는 장치이면서 동시에 무장 해제 수단이다 — `pinnedRecordFields`가
   `SeenLateBand`·`ExtendedBand` 둘만 못박는다(M8이 그것을 확인한다).
4. **어떤 기록 이름도 대지 않고 기록에 닿는 경로** — unsafe 포인터 변환.

**이 가드는 이미 두 번 뚫렸다. 세 번째가 없다고 주장하지 않는다.**

---

## R2 구현 — 콘솔 `/signals`의 붕괴 경보

- `internal/console/signals.go`: `signalsBandTally`에 `Alarm string` 추가,
  `signalsBandAlarm(candidate.BandTally) string` 신규, `signalsBandTalliesFrom`에서 배선.
- `internal/console/templates_signals.go`: `signalsband` 템플릿에
  `{{if .Alarm}}<p class="bad">{{.Alarm}}</p>{{end}}`를 **건수 위**에 추가.
- `internal/console/band_alarm_test.go`(신규): 붕괴/정상 두 상태.

판정은 **`candidate.BandTally.Collapsed`이고 콘솔은 문장만 소유한다.**
`signalsTallyAlarm`이 같은 분리를 같은 이유로 한다 — 규칙을 두 번 구현하면 언젠가
어긋나고, 어긋남은 "평온해 보이는 화면"으로 나타난다.

**UI 마찰 없음.** 표시 전용이다. 확인 문구·추가 승인·버튼·폼을 넣지 않았다.

### 실제 렌더 결과

붕괴 상태(측정 131건이 전부 밴드 `0`만 넘음 — 이 저장소의 KR 실제 모양):

```html
<h3>그림자 밴드 — extended <span class="muted">(veto가 아니다)</span></h3>
<p class="muted">기록만 하고 아무것도 판정하지 않는다. 임계는 여기서 도출된다.</p>
<p class="bad">경보 — 측정된 131건이 전부 같은 교차 집합을 냈다. 이 눈금은 아무것도
분해하지 못했으므로 여기서 임계를 승인할 수 없다. 옆의 건수는 정직하고, 어떤 질문에도
답하지 않는다.</p>
<p class="muted">후보 156개 중 131개 측정. 교차:
0 131 · 2 0 · 4 0 · 6 0 · 8 0 · 10 0 · 20 0 · 30 0 · 50 0 · 100 0</p>
<p class="muted">미측정: <code>NO_OBSERVATIONS</code> 25</p>
```

정상 상태(밴드 `0`이 165 대 4로 가름 — US의 실제 모양):

```html
<h3>그림자 밴드 — extended <span class="muted">(veto가 아니다)</span></h3>
<p class="muted">기록만 하고 아무것도 판정하지 않는다. 임계는 여기서 도출된다.</p>
<p class="muted">후보 194개 중 169개 측정. 교차:
0 165 · 2 4 · 4 0 · 6 0 · 8 0 · 10 0 · 20 0 · 30 0 · 50 0 · 100 0</p>
<p class="muted">미측정: <code>NO_OBSERVATIONS</code> 25</p>
```

**경보는 건수 위에 붙고 건수를 대체하지 않는다.** 경보가 말하는 대상이 그 건수다.
**상시 점등이 아니다** — 같은 코드가 정상 상태에서는 아무것도 내지 않는다.

### R2 변이 확인

| 변이 | 결과 |
|---|---|
| `signalsBandTalliesFrom`에서 `Alarm` 미배선(= 이 change 전 상태) | **RED** — `131 measured records that all produced the same crossings render with no alarm` |
| 템플릿에서 `{{if .Alarm}}` 블록 제거 | **RED** — 같은 단언. 문장이 만들어져도 화면에 닿지 않으면 실패한다 |
| 정상 tally(165 대 4) | 경보 없음 — 음성 대조가 살아 있다 |

되돌린 뒤 두 파일 모두 의도한 최종 상태다.

---

## R3 — `Collapsed()`의 공허한 참: **넣기로 판단했다** (§5.3)

tasks §5.3이 "지금 넣을지 판단하고 판단을 적어라"고 했다. **넣었다.**
`Collapsed()` 첫 줄이 `t.Measured < 1 || len(t.Crossed) == 0`이 됐다.

근거 셋(전문은 issues **I14**):

1. I3이 `BandsFor(VetoExtended) = nil`을 **의무로** 적어 놨다. 그날 `extended` 경보가
   영구 점등된다 — 가정이 아니라 예약된 상태다.
2. **R2가 비용을 올렸다.** 경보가 이제 운영자가 열어 두는 화면에 뜬다. 상시 점등은 없는
   것보다 나쁘다. 이 요구사항의 근거가 *"사람이 알아차리는 것에 맡겨서는 안 된다"*인데
   상시 점등이 정확히 그 상태를 만든다.
3. 문장이 거짓이 된다. 교차가 하나도 없는데 *"전부 같은 교차 집합을 냈다"*는 공허하게
   참이다. **눈금이 없는 것과 눈금이 아무것도 못 가른 것은 다르다.**

RED 확인:

```
$ go test ./internal/candidate/ -run TestATallyWithNoScaleIsNotReportedAsCollapsed -v
    bandscale_test.go:518: a tally with no bands at all reports itself collapsed (crossed = map[])…
    bandscale_test.go:523: and it carries an alarm sentence: "ALARM near_high: all 2 measured…"
--- FAIL
```

한 줄을 넣은 뒤 GREEN. `BandTally.Collapsed`의 FLM B1을 갱신했다.

---

## R4~R8 (P2) — issues.md로 옮겼다 (§5.3)

| # | 처리 | 위치 |
|---|---|---|
| R3 | **고쳤다** (위) | I14 |
| R4 | `wrapCounts` 단위 혼용·폭 미검사 — **남긴다**, 오늘의 데이터로 도달 불가, 후속 Small | I15 |
| R5 | US 수치가 두 스캔 합산 — **정정 기록**. 결론은 유효, 근거의 형태가 틀렸다 | I16 |
| R6 | §2.6 렌더 예시가 합성 fixture인데 실측처럼 읽힌다 — **정정 기록**, 주석 수정은 후속 | I17 |
| R7 | `orderedCounts` 호출자 0 — **남긴다**, 삭제는 후속 Small | I18 |
| R8 | `distancePct`의 "the one value…"가 거짓 — **기록**, `level.go`는 §5 범위 밖 | I19 |

**R4·R6·R7을 여기서 고치지 않은 공통 이유**: 셋 다 `cmd/tossctl/candidatebands.go`·
`candidate.go`의 렌더 코드이고, 손대면 §5의 diff에 렌더 경로가 들어와 FLM 대상이 늘어난다.
§5의 범위는 P1 둘이다. 셋 다 표시 폭·주석·죽은 함수이고 판정·저장·주문 어디에도 닿지 않는다.

---

## FLM 대상 변화 (§4.3)

가드는 테스트 파일이지만 `signals.go`와 `band.go`는 프로덕션이다.
`check_analysis.py`를 돌려 대상을 확인했고 **면제를 주장하지 않았다.**

| 함수 | 성격 | 산출물 |
|---|---|---|
| `internal/console/signals.go:signalsBandTalliesFrom` | **수정** — `Alarm` 필드 배선 1줄. B1~B4는 base 그대로 | 신규 4종 |
| `internal/console/signals.go:signalsBandAlarm` | 신규(인접) — 분기 B1 하나 | 신규 4종 |
| `internal/candidate/band.go:BandTally.Collapsed` | **수정** — B1에 disjunct 추가 | 기존 갱신(ast.json 재생성 + 두 map) |
| `internal/candidate/band.go:bandQuantiles`·`CollapsedAlarm`·`TallyBands` | 무변경. `band.go` 파일 해시가 바뀌어 ast.json만 재생성 | ast.json 갱신 |

`internal/candidate/bandguard_test.go`·`bandscale_test.go`·`internal/console/band_alarm_test.go`는
**신규 파일**이므로 그 안의 함수는 전부 면제다(tasks의 FLM 대상 규칙).
`internal/candidate/watch.go`는 **최종 상태가 무변경**이다(sha256 대조).
`templates_signals.go`는 const 문자열뿐이라 대상 함수가 없다.

```
$ python3 tools/logic-map/check_analysis.py --change refine-extended-shadow-bands
[logic-map] refine-extended-shadow-bands: evidence complete or diff-proven exempt
```

---

## 검증 명령과 실제 결과 (§5.4)

| 명령 | 결과 |
|---|---|
| `go test ./... -count=1 -v` | **3710 passed, 0 failed, 51 packages ok** |
| `go vet ./...` | 출력 없음 |
| `make lint` | `go vet ./...` — 통과 |
| `$(go env GOROOT)/bin/gofmt -l .` | 출력 없음 (`gofmt`가 PATH에 없다) |
| `go test -race ./internal/candidate/... ./cmd/tossctl/... ./internal/console/...` | **3개 패키지 ok** |
| `python3 tools/logic-map/check_analysis.py` | evidence complete or diff-proven exempt |

**upstream 상속 테스트 회귀 없음.** 3706 → **3710**(+4: 새 가드 1, `Collapsed` 공허성 1,
콘솔 경보 2). **줄어든 것은 없다.**

### 임계 확인 — 요구받은 대로 확인하고 보고한다

**`extended`는 이 change 이후에도 veto하지 않는다.**

- `VetoThresholds.ExtendedGainPct`를 §5에서 읽지도 쓰지도 않았다.
  `grep -rn ExtendedGainPct --include=*.go` → 선언·doc·`AssessExtended`의 두 읽기뿐이고
  **비테스트 코드 어디에서도 대입되지 않는다.** 따라서 `AssessExtended`는
  `thresholdReason` 분기에서 항상 `UnmeasuredVeto`를 돌려준다.
- `TestMeasureExtendedBandNeverReadsTheVetoThreshold`가 계속 GREEN이다 — 읽기 시작하면
  AST 절반과 동작 절반이 모두 실패한다.
- `ExtendedGainBands`·`SeenLatePercentileBands` **무변경**(§1에서 확정된 값 그대로).

### 안전 불변식

- `internal/risk`·`internal/execgw`·`internal/exitpolicy`·`internal/trading` **전부 무변경.**
- 주문 제출·취소·정정, 손절·익절·사이징, Guardian·kill switch, intent journal·원장,
  reconciliation, retry matrix·rate limit, 인증·세션, 체결 감지 **어느 경로도 닿지 않는다.**
- 토글·운영 설정 flip 없음. LIVE 주문 side effect 없음. 위험 등급 **Normal.**
- `tossctl candidate scan`을 **실행하지 않았다.** 운영 store를 읽지도 쓰지도 않았다 —
  §5는 렌더 경로를 직접 호출하는 테스트만 쓴다.
- 커밋·push 하지 않았다.

---

## 막힌 것 · tasks가 코드와 안 맞는 것

### 1. §5.5 2차 독립 리뷰는 하지 않았다 — `make gate`가 2/8에서 실패한다

구현자가 자기 구현을 독립 검증할 수 없다. **정상적인 결과이고 숨기지 않는다.**

**2차 리뷰어가 반드시 직접 할 것**: 위 표의 14개 변이를 **직접 다시** 넣고 되돌릴 것.
`zz_bypass_probe.go`는 삭제됐으므로 다시 만들어야 한다. 그리고 **새 우회로를 스스로
찾을 것** — 특히 위 "못 막는 것" 넷 중 1번(다른 패키지)이 실제로 뚫리는지.

### 2. tasks §5.1이 예고한 오탐은 **국소 규칙에서는 일어나지 않는다**

tasks는 리뷰어 제안(전이적 검사)에 대해 *"`assessOne → shadowBandsFor →
MeasureExtendedBand`를 오탐으로 잡을 수 있다. 그 오탐을 어떻게 피하는지가 비용이다"*라고
썼고 **그것은 맞다** — 만들어 보니 정확히 그랬다. 국소 규칙을 고른 이유가 그것이고,
국소 규칙에서는 `shadowBandsFor`가 기록 둘을 반환하므로 애초에 규칙을 만족한다.
**tasks가 틀린 것이 아니라, 두 제안 중 하나에만 있는 비용이었다.**

### 3. tasks §5.1의 두 번째 우회 시도 예상은 **맞았다**

*"`take(v).Crossed("6")` → RED여야 한다(기존 어휘 규칙이 잡는다 — 확인하라)"* —
확인했다. M2가 어휘 가드에서 RED다. **14개 중 어휘 가드가 잡은 유일한 하나**다.

### 4. §5.1의 제안대로 구현하면 **두 번 보강해야 한다** — 제안이 불충분했다

Manager 제안은 *"기록을 읽는 함수는 기록 타입을 반환하거나, 아무것도 반환하지 않거나,
기록 타입 자신의 메서드여야 한다"*이다. 그대로 구현한 첫 형태는 M9·M10·M13을 놓쳤다.
"읽는다"의 정의에 **호출을 통한 세탁**과 **comma-ok 형태**가 빠져 있었기 때문이다.
`retire-gainers-source`와 이 change에서 반복된 *"결론은 맞고 근거는 불충분"*의 다음
사례다 — **형태의 판단은 옳았고 정의가 좁았다.**

### 5. 새 콘솔 테스트가 기존 허용목록에 걸렸다 — 목록을 넓히지 않고 우회했다

`TestOnlyTheListedFilesCanNameTheChaseVerdict`(`consumer_guard_test.go`)가
`internal/console/band_alarm_test.go`를 잡았다. 그 테스트 fixture는 `candidate.Verdict`를
**필요로 하지 않으므로**(밴드 블록은 tally만으로 렌더된다) 목록에 더하는 대신 fixture에서
verdict 참조를 뺐다. `TestOnlyTheListedFilesCanNameTheChaseVerdict`가 지키려는 것이 바로
*"chase 독자가 늘어나는 것은 누군가 내린 결정이어야 한다"*이므로, 필요 없는 파일을
목록에 넣는 것은 그 규칙을 약화하는 것이다.

### 6. `signalsBandTally`에 분위수는 여전히 없다

§5.2는 경보만 지목했고 SHALL도 경보만 요구한다. 분위수를 콘솔에 올리는 것은 후속
change이고 등급 Small이다(issues **I13** 말미).

---

## §5.4 — `make sdd-sync` → `make sdd-check` → `make gate`

기록(issues.md·review.md·FLM 산출물·tasks.md)까지 **전부 끝낸 뒤** 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain) |
| `make sdd-check` | **통과, exit=0** — `[index-freshness] CodeGraph hard-evidence index matches the worktree`, `[agent-config] … synchronized`, memory `valid`, `[pm] … current`, sdd-test 전부 통과, typedb/neo4j running |
| `make gate CHANGE=refine-extended-shadow-bands` | **FAIL — 2/8에서 미완료 태스크 1건(`5.5 2차 독립 리뷰`)** |

```
==> 1/8 tasks.md 확인
OK: openspec/changes/refine-extended-shadow-bands/tasks.md

==> 2/8 미완료 태스크 확인
미완료 태스크 1 건:
203:- [ ] 5.5 §5 수정분에 대한 **2차 독립 리뷰**. 범위는 §5이 만든 diff. 리뷰어는 5.1의

GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다
```

**게이트 실패는 사실 그대로 기록한다.** 실패 지점은 2단계 하나이고 원인은 §5.5가 열려
있다는 것뿐이다. **§4.4는 이제 닫혀 있고**(위 "§4.4 체크박스에 대하여" 참조) 미완료 목록에
없다. 나머지 단계를 각각 손으로 돌려 확인했다.

```
1/8 tasks.md 확인                     OK
2/8 미완료 태스크 확인                 FAIL — 5.5 2차 독립 리뷰 (구현자가 할 수 없다)
3/8 review.md 존재                    OK — 이 파일
4/8 Function Logic Map                OK — evidence complete or diff-proven exempt (exit=0)
5/8 make sdd-check                    OK — exit=0
6/8 make test                         OK — exit=0
7/8 make vet                          OK — exit=0
8/8 make validate                     OK — Totals: 25 passed, 0 failed (25 items)
openspec validate --strict            Change 'refine-extended-shadow-bands' is valid
python3 tools/pm/generate_master_tracker.py --check
                                      [pm] hierarchy and generated trackers are current
```

**미완료로 남은 것은 §5.5 하나뿐이고, 다른 단계에서 걸린 것은 없다.**

> 이 절을 쓰면서 `.md`를 편집했으므로 CodeGraph fingerprint가 다시 stale이다.
> 직후 `make sdd-sync`를 한 번 더 돌려 되살렸고 `make sdd-check`로 확인했다(둘 다 exit=0).
> 2차 독립 리뷰가 `make gate`를 다시 돌릴 때, 그 사이에 편집이 없었다면 그대로 쓸 수 있다.

---

## 최종 상태

**§5.1(R1)과 §5.2(R2)를 닫았다. §5.3의 R3도 판단 끝에 고쳤다.**
남은 것은 **§5.5 2차 독립 리뷰 하나**이고, 그것은 구현자가 할 수 없다.

랜딩 가능 여부에 대해 구현자는 판단하지 않는다. 판단에 필요한 것을 위에 전부 적었고,
특히 **가드가 못 막는 넷**(issues I20)은 막았다고 쓰지 않고 못 막았다고 적었다.

---

## 부록 — §5 착수/종료 sha256 대조 (전체)

착수 시점에 9개 파일의 sha256을 기록하고 종료 시점에 대조했다.
**바뀌면 안 되는 넷은 바이트 단위로 같고**, 바뀐 다섯은 전부 의도한 것이다.

| 파일 | 결과 | 의도 |
|---|---|---|
| `internal/candidate/watch.go` | **OK — 동일** | M1·M2·M5 변이를 넣고 되돌린 파일. **최종 무변경**이어야 한다 |
| `internal/candidate/band_test.go` | **OK — 동일** | 기존 어휘 가드. §5는 손대지 않는다 |
| `cmd/tossctl/candidate.go` | **OK — 동일** | R4·R6·R7을 범위 밖으로 둔 결과 |
| `cmd/tossctl/candidatebands.go` | **OK — 동일** | 같음 |
| `internal/candidate/bandguard_test.go` | 변경 | §5.1 새 가드 |
| `internal/candidate/bandscale_test.go` | 변경 | §5.3 `Collapsed` 공허성 테스트 |
| `internal/candidate/band.go` | 변경 | §5.3 `Collapsed` 한 줄 + doc |
| `internal/console/signals.go` | 변경 | §5.2 |
| `internal/console/templates_signals.go` | 변경 | §5.2 |

신규 파일 `internal/console/band_alarm_test.go` 하나가 늘었다.
임시 probe 둘(`internal/candidate/zz_bypass_probe.go`,
`internal/console/zz_render_probe_test.go`)은 **삭제했고** `git status --short`에 없다.

`ad16f56` 대비 프로덕션 diffstat(§0~§5 누적):

```
cmd/tossctl/candidate.go                 |  73 ++++----
cmd/tossctl/candidate_test.go            |  21 ++-
docs/pm/portfolio/_registry.yaml         |   3 +-
internal/candidate/band.go               | 222 +++++++++++++++++++++++++-
internal/candidate/band_test.go          |  83 ++++++++---
internal/candidate/watch.go              |  39 +++++-
internal/console/signals.go              |  33 ++++-
internal/console/templates_signals.go    |   9 ++
tools/pm/test_generate_master_tracker.py |   1 +
9 files changed, 424 insertions(+), 60 deletions(-)
```

§5가 더한 것은 그중 `signals.go` +31/−2, `templates_signals.go` +9/−0,
`band.go`의 한 줄과 그 doc, 그리고 신규 테스트 파일들이다.

> 이 부록을 쓰면서 다시 `.md`를 편집했으므로, 직후 `make sdd-sync` → `make sdd-check`를
> 한 번 더 연속 실행했다(둘 다 exit=0). 위 게이트 결과는 그 사이에 코드가 바뀌지 않았으므로
> 그대로 유효하다.

---

# 2차 독립 리뷰 (§5.5)

- 날짜: 2026-07-29
- 역할: **2차 독립 리뷰어.** §5가 만든 diff만 본다. 구현자·1차 리뷰어와 분리된 컨텍스트다.
- base-commit: `cat base-commit.txt` → `ad16f56d7db04871699e060f120f249518fbee77`,
  `git log --oneline -1` → `ad16f56`. **일치한다.**
- 방법: 변이는 전부 **편집으로 넣고 편집으로 되돌렸다.**
  `git checkout --`·`git restore`·`git stash`는 **한 번도 쓰지 않았다.**
  착수 시점에 12개 파일의 sha256을 기록하고 종료 시점에 `sha256sum -c`로 대조했다.
  임시 probe는 `internal/candidate/zz_probe.go` 하나이며 관찰 후 삭제했다.
- 커밋·push 없음. `tossctl` 실행 없음. 운영 store를 읽지도 쓰지도 않았다.

> **판정: 랜딩 불가.** 사유는 아래 **S1**이다. 새 가드는 §5.1이 닫으려던 구멍의
> **한 변종만** 닫았고, 같은 형태의 우회가 **일곱 개** 남아 있다. 그중 하나는
> 헬퍼조차 필요 없는 **한 함수**이고, 실제로 veto를 바꾼다는 것을 실행으로 확인했다.

---

## A. 새 가드를 뚫었다

### A0. 기준선

```
$ rtk proxy go test ./internal/candidate/ -run 'TestNoFunctionTurnsAShadowRecordIntoSomethingElse|TestNoFunctionThatProducesAVerdictCanSeeAShadowBand' -v -count=1
    band_test.go:394: checked 12 verdict-producing functions
--- PASS: TestNoFunctionThatProducesAVerdictCanSeeAShadowBand (0.01s)
    bandguard_test.go:671: record readers: [BandTally.Collapsed BandTally.CollapsedAlarm
        MeasureExtendedBand MeasureSeenLateBand ShadowBand.Crossed ShadowBand.Reason TallyBands]
--- PASS: TestNoFunctionTurnsAShadowRecordIntoSomethingElse (0.01s)
```

### A1. 구현자 표의 재현 — 여섯을 임의로 골라 직접 넣었다

`판정` = 새 가드(`TestNoFunctionTurnsAShadowRecordIntoSomethingElse`),
`어휘` = 기존 가드(`TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`).

| # | 변이 | 어휘 | 판정 | 가드가 지목한 이름 |
|---|---|---|---|---|
| M1 | `m1worthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }` | PASS | **RED** | `zz_probe.go:m1worthy … reads Crossed out of one` |
| M4 | `b := v.ExtendedBand; return b.Measured && b.Value >= "6"` | PASS | **RED** | `m4worthy … reads Measured, Value out of one` |
| M5 | 헬퍼가 **기록을 반환**하고 `assessOne`이 `m5hot(v).Measured` | PASS | **RED** | `watch.go:assessOne … reads Measured out of one` (+`m5hot`이 readers에 등록) |
| M7 | 패키지 var stash `var m7last ShadowBand` + 다른 함수가 읽음 | PASS | **RED** | `m7read … reads Crossed out of one` |
| M9 | `strings.Contains(fmt.Sprintf("%v", v.ExtendedBand), "6")` | PASS | **RED** | `m9worthy … hands one to Sprintf` |
| M13 | 제네릭 항등 `m13pass(v.ExtendedBand).Crossed("6")` | PASS | **RED** | `m13worthy … hands one to m13pass` |

**여섯 전부 재현됐다. §5의 14종 표는 서술이 아니라 관측이다.** M5는 구현자 표보다
한 걸음 더 나아가 `assessOne` 자신도 지목했다. 여섯 모두 되돌린 뒤 `watch.go` =
`1aa356933d8e…7b0106`(착수 기록과 동일).

### A2. 새 우회로 — **일곱이 통과한다**

전부 넣고, 돌리고, 되돌렸다. `assessOne`에 넣은 분기는 항상
`v.Chase.Extended = RaisedVeto()` 한 줄이고, 승인된 적 없는 숫자는 `"6"`이다.

| # | 축 | 변이 | 어휘 | 판정 | 결과 |
|---|---|---|---|---|---|
| **B1** | **void 함수** | `func b1annotate(v *Verdict) { if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() } }` + `assessOne`에서 `b1annotate(&v)` | PASS | **PASS** | **뚫림** |
| **B2** | void, 포인터 out-param | `func b2note(v Verdict, hot *bool) { *hot = v.ExtendedBand.Crossed("6") }` | PASS | **PASS** | **뚫림** |
| **B3** | void, 패키지 var | `var b3hot bool; func b3note(v Verdict) { b3hot = v.ExtendedBand.Crossed("6") }` | PASS | **PASS** | **뚫림** |
| **B4** | **타입 별칭** | `type b4band = ShadowBand; func b4take(v Verdict) b4band { return v.ExtendedBand }` + `b4take(v).Measured && b4take(v).Value >= "6"` | PASS | **PASS** | **뚫림** |
| **B5** | **임베딩(승격 필드)** | `type b5box struct{ ShadowBand }; func b5wrap(v Verdict) b5box { return b5box{v.ExtendedBand} }` + `b5wrap(v).Measured && b5wrap(v).Value >= "6"` | PASS | **PASS** | **뚫림** |
| **B6** | **리플렉션** | `reflect.ValueOf(v).FieldByName("ExtendedBand").FieldByName("Value")` | PASS | **PASS** | **뚫림** |
| **B9** | 기록을 `error`로 감싸기 | `type b9note ShadowBand` + `Error() string` + void 함수가 `b9err = b9note(v.ExtendedBand)` | PASS | **PASS** | **뚫림** |
| B7 | `encoding/json` 왕복 | `json.Marshal(v)` → 익명 구조체로 Unmarshal | PASS | **RED** | 막힘 — 다만 **우연히**: 익명 구조체가 `ExtendedBand`를 비-기록 타입으로 선언해서 `pinnedRecordFields`가 `Fatalf` |
| **B8** | **다른 패키지** | 2026-07-28의 그 한 줄을 그대로 `internal/console/signals.go:652`(허용목록에 이미 있는 파일)에 | PASS | **PASS** | **뚫림** |

B1·B2·B3·B9에서 가드는 우회 함수를 **record reader로 인식하고서 통과시킨다**
(`record readers: […] b1annotate`). B4·B5·B6에서는 **인식조차 하지 않는다**
(readers 목록이 기준선 7개 그대로다).

### A3. B1의 전체 출력 — 51개 패키지가 전부 초록이다

```go
// zz_probe.go — 판정 타입을 반환하지 않는다. 아무것도 반환하지 않는다.
func b1annotate(v *Verdict) {
	if v.ExtendedBand.Crossed("6") {
		v.Chase.Extended = RaisedVeto()
	}
}

// watch.go, assessOne 안 — 밴드 대입 줄 바로 아래 한 줄
	b1annotate(&v)
```

```
$ rtk proxy go vet ./...                                    → 출력 없음
$ rtk proxy go test ./internal/candidate/ -run '…' -v
    band_test.go:394: checked 12 verdict-producing functions
--- PASS: TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
    bandguard_test.go:671: record readers: […TallyBands b1annotate]
--- PASS: TestNoFunctionTurnsAShadowRecordIntoSomethingElse
--- PASS: TestAShadowBandCannotBeReadAsAVeto
--- PASS: TestOnlyTheListedFilesCanNameTheChaseVerdict
$ rtk proxy go test ./... -count=1                          → ok 51 / 51, FAIL 0
```

**승인된 적 없는 숫자 `6`이 veto를 결정하고, vet·네 가드·51개 패키지가 전부 초록이다.**
2026-07-28은 한 줄, 2026-07-29는 두 줄이었다. **이번은 다시 한 함수다.**

### A4. 이론이 아니라는 증거 — 실행으로 확인했다

같은 형태에서 밴드만 `"0"`으로 바꿨다. 오늘의 KR 데이터가 실제로 넘는 밴드다.

```
$ rtk proxy go test ./internal/candidate/ -run '…두 가드…' -count=1   → ok (둘 다 PASS)
$ rtk proxy go test ./... -count=1
--- FAIL: TestAScanCountsUnmeasuredVetoesAndNeverCallsAnAbsentThresholdAPass
    watch_test.go:837: unmeasured = 0, want 2 — the screen's default reading has to be 'mostly unchecked'
    watch_test.go:841: THRESHOLD_ABSENT = 2, want 4 (two codes over two candidates)
--- FAIL: TestTheScanJSONReportsTheCountsAnOperatorActsOn
```

**그림자 기록이 실제로 veto를 바꿨다.** 알아차린 것은 건수를 세는 **동작** 테스트
둘뿐이고, 이 change가 그 목적으로 만든 **구조** 가드 둘은 아무 말도 하지 않았다.
그리고 `"6"`— 지금 논의 중인 임계 후보 — 에서는 오늘의 데이터에 넘는 후보가 없으므로
**아무것도 실패하지 않는다.** 조용히 지나가는 것이 정확히 그 값이다.

### S1 (P1 — 랜딩 차단) `mayReadRecords`의 "아무것도 반환하지 않으면 허용"은 Go에서 참이 아니다

`internal/candidate/bandguard_test.go:531-534`

```go
	results := resultLists(fn)
	if len(results) == 0 {
		return true
	}
```

규칙의 전제는 *"기록을 읽고 아무것도 돌려주지 않는 함수는 판정을 옮길 수 없다"*이다.
Go에서 이것은 거짓이다. void 함수는 **포인터 인자·패키지 변수·채널·map 쓰기**로
답을 밖에 내놓는다. 그리고 `assessOne`이 그 함수를 부를 때
`carriesRecordsSafely`(`:502-511`)가 `recordResults[name]`을 빈 슬라이스로 찾아
"결과가 전부 기록"을 **공허하게 참**으로 판정하므로 인자 검사조차 건너뛴다.

가장 나쁜 철자가 B1이다. **헬퍼가 없다. 홉이 없다.** 변환 전체가 void 함수 하나 안에
있고, 그것은 상상할 수 있는 가장 자연스러운 리팩터다 — *"밴드→chase 배선을 assessOne
밖으로 뺀다."* 그리고 이 change의 review.md **0.4-c**가 `shadowBandsFor`에 `*Verdict`를
주지 않은 이유로 적은 것이 바로 이것이다: *"Handing it a \*Verdict would have moved the
one-line transition rather than removed it — and would have moved it somewhere the check
does not look."* **`*Verdict`를 받는 void 함수가 정확히 그 자리이고, 두 가드 모두 거기를
보지 않는다.**

**스펙에 대해 거짓인 지점.** spec delta Scenario *"한 홉 건너서도 도달할 수 없다"*는
*"판정을 생산하지 않는 함수가 그림자 기록을 읽어 **기록이 아닌 값**으로 바꾸고, 판정을
생산하는 함수가 그 값으로 분기하면 → **검사가 실패한다**"*이다. **B2와 B3가 문자 그대로
그 문장이고, 검사는 통과한다.** 지금 랜딩하면 승인된 SHALL이 코드에 대해 거짓인 채
아카이브된다 — D10과 §5.2가 경고한 바로 그 오염이며, 이번에는 §5.1 자신에 대해서다.

### S2 (P1) 타입 별칭과 임베딩이 기록을 **한 홉에** 세탁한다

`bandguard_test.go:166-200`(`typeIdents`/`namesARecord`)은 타입 **철자**로 판정한다.
`type b4band = ShadowBand`는 같은 타입인데 이름이 다르므로 `namesARecord`가 false다.
변환 표현식이 없으므로(별칭 대입은 항등) `carriesRecordsSafely`가 볼 CallExpr도 없다.
**void 함수도 필요 없다** — B4는 기록을 반환하는 것처럼 보이는 헬퍼 하나로 끝난다.

임베딩(B5)은 다른 경로로 같은 곳에 닿는다. `bandShape.fields`는 임베드 필드를
`ShadowBand`라는 이름으로 **올바르게** 기록으로 인식하지만, 승격 접근
`b5wrap(v).Measured`는 셀렉터의 좌변이 `b5box`라서 `isRecordExpr`가 false다.
그리고 `b5box{v.ExtendedBand}`는 `CompositeLit`이라 walk의 두 case(`SelectorExpr`·
`CallExpr`) 어디에도 걸리지 않는다.

### S3 (P2) 리플렉션은 막히지 않는다 — I20 #4의 서술과 반대다

B6은 `reflect.ValueOf(**v**)`로 **Verdict**를 넘긴다. 기록은 인자로 나가지 않고
문자열 `"ExtendedBand"`로 도달한다. issues.md I20 #4가 *"`reflect.ValueOf(band)`처럼
기록을 인자로 건네는 형태는 막는다"*라고 쓴 것은 **그 철자에 대해서만** 맞다.

### A5. 시도했지만 막힌 것 — 전부 나열한다

위 표의 M1·M4·M5·M7·M9·M13(RED, 재현), B7(RED, 다만 우연). 그 밖에 코드를 읽고
**넣기 전에 막힌다고 판단한 것**과 그 근거:

- `b, _ := boxed.(ShadowBand)` (comma-ok): `boundRecords.bind`(`:413-419`)가
  단일 rhs·복수 lhs를 comma-ok로 보고 첫 이름을 묶는다 → 읽기로 잡힌다(구현자 M10).
- `m := map[string]any{…}; m["b"].(ShadowBand)`: `isRecordExpr`의 `*ast.TypeAssertExpr`
  case(`:337`)가 `namesARecord(ShadowBand)`로 잡는다.
- `for _, c := range v.ExtendedBand.Crossings`: `RangeStmt` case(`:452-458`)가 묶는다.
  기록을 반환하는 헬퍼로 한 홉 빼도 `producers`가 그 호출을 기록으로 인식하므로
  최종 독자에서 RED다(M5와 같은 종결).
- `type mb ShadowBand`(별칭이 **아닌** 정의 타입)를 그냥 반환: 변환
  `mb(v.ExtendedBand)`가 `carriesRecordsSafely`에서 unknown callee라 `handed`로 잡힌다.
  **그래서 B9는 그 변환을 void 함수 안으로 옮겼고, 그러자 통과했다** — S1과 같은 원인이다.
- `internal/candidate` 밖의 **새 파일**에서 chase를 이름하기:
  `TestOnlyTheListedFilesCanNameTheChaseVerdict`의 `namesVerdict`가 **비수식**
  `.Chase`·`.Passed` 셀렉터도 잡으므로 허용목록 없이는 RED다. 그래서 B8은
  **이미 목록에 있는 파일**에 넣었고, 그러자 통과했다.

### S4 (P2) `verdictProducers = 12` 하한은 B1류를 세지 않는다

`checked`는 판정 **타입을 반환하는** 함수만 센다. B1·B2·B3·B9의 우회 함수는 어느 것도
`checked`를 늘리지 않으므로, 하한이 아무리 빡빡해도 이 계열에 대해서는 침묵한다.
하한 자체는 유효하다(1차 리뷰 A4에서 확인됨). 지적하는 것은 하한이 **다른 축**을
잰다는 것이다.

---

## B. I20의 "못 막는 넷"이 정직한가 — 넷은 대체로 맞고, **셋이 빠졌고, 하나가 틀렸다**

| I20 | 재현 결과 | 판정 |
|---|---|---|
| #1 다른 패키지 | **B8로 재현. 통과한다.** | **참이나 과소 서술** |
| #2 테스트 파일 | 두 가드 모두 `parser.ParseDir` 필터에서 `_test.go` 제외(`band_test.go:342`, `bandguard_test.go:591`). 코드로 확인 | 참 |
| #3 모호한 필드 이름 | 패키지 전체를 AST로 열거했다(아래) | **참, 정밀함** |
| #4 unsafe만 남는다 | **B6이 반증한다** | **거짓** |

**#3 — Manager의 질문("다른 이름으로 기록을 들고 있는 구조체가 이미 있는가")에 대한 실측.**
`internal/candidate`의 비테스트 코드 전체를 훑어 기록 타입을 담는 필드 이름을 뽑았다.

```
field names that hold a shadow record:
  Bands          watch.go:540    RECOGNISED
  Crossings      band.go:184     DROPPED (also non-record at watch.go:539, metrics.go:296)
  ExtendedBand   watch.go:299    RECOGNISED
  Quantiles      band.go:388     RECOGNISED
  SeenLateBand   watch.go:299    RECOGNISED
```

**I20 #3의 서술이 정확하다.** 오늘 떨어지는 것은 `Crossings` 하나뿐이고 사유도 적힌
그대로다. 덧붙일 잔여물 하나: `Bands`와 `Quantiles`도 기록을 담지만 `pinnedRecordFields`에
**없다** — 같은 이름의 비-기록 필드를 어디든 선언하면 조용히 빠진다. I20 문장에서
따라 나오지만 이름으로 적혀 있지는 않다. **P3.**

**#1 — 과소 서술인 지점.** *"그 두 패키지는 판정을 소비할 뿐 엔진에 되먹이지 않는다"*는
엔진에 대해서는 참이다. 그러나 **목적에 대해서는 참이 아니다.** `verdictReaders`에
이미 올라 있는 **10개 파일에는 방어가 전혀 없고**, 그중 하나가 이 change가 직접 고친
`internal/console/signals.go`다. 그리고 스캔 리포트와 `/signals`는 **사람이 임계를
승인할 때 읽는 바로 그 표면**이다. 거기에 밴드에서 나온 veto가 그려지면, 그것은
"소비된 표시"가 아니라 **승인의 근거로 읽힌다.** 이 change의 존재 이유가 그 승인을
오염에서 지키는 것이므로, #1의 완화 문구는 위험을 실제보다 작게 말한다. **P2.**

**빠진 것 셋 — I20에 한 줄도 없다.**

1. **void·부수효과 함수**(S1). 가드의 규칙 자체가 명시적으로 허용하는 형태이므로,
   "못 막는 것"이 아니라 **"막는다고 쓴 것"**이다. 가장 심각하다.
2. **타입 별칭**(S2 전반). 별칭 선언은 `ShadowBand`를 **이름한다** — I20 #4의
   *"어느 이름도 대지 않고"*에 해당하지 않는다.
3. **임베딩·승격 필드**(S2 후반). 임베드 필드는 `fields`에 **올바르게** 인식되므로
   #3의 "모호한 이름"에도 해당하지 않는다.

**결론: I20은 은폐가 아니다.** 적은 넷 중 셋은 맞고 하나는 철자 수준에서만 맞다.
문제는 정직성이 아니라 **완전성**이며, 빠진 첫 번째가 랜딩 차단 사유다.

---

## C. R2 — 콘솔 붕괴 경보

**spec delta Scenario와 일치한다.** *"측정 건수가 1 이상이고 측정된 기록 전부가 같은
교차 집합"* → `candidate.BandTally.Collapsed`, *"보고 표면은 … 경보로 표시한다"* →
`<p class="bad">`. 판정은 `internal/candidate`에 있고 콘솔은 문장만 갖는다
(`signals.go:566-575`가 `t.Collapsed()`를 **부르고 다시 구현하지 않는다**) — 요구대로다.

**실제로 렌더된다 — 진짜 경로로 확인했다.** `band_alarm_test.go`는 stub이 아니라
`renderSignals` → `h.get(t, "/signals")`로 **인증된 실제 HTTP GET**을 하고 실제 템플릿을
통과한 HTML을 문자열로 검사한다(`signals_test.go:50-59`). 템플릿에서
`{{range .Bands}}{{template "signalsband" .}}{{end}}`는 `signalsmarket`의
`{{if .Read.Known}}` 아래 **후보 행과 나란한 형제**이므로(`templates_signals.go:71`),
행이 0건인 fixture가 프로덕션 경로를 그대로 탄다.

**UI 마찰 없음 — 기계적으로 확인했다.**

```
$ rtk proxy git diff internal/console/templates_signals.go | grep '^+' \
    | grep -Ei "<form|<button|<input|confirm|type=|onclick"
NONE — display-only
```

추가된 것은 `{{if .Alarm}}<p class="bad">{{.Alarm}}</p>{{end}}` 뿐이다. 확인 문구
타이핑·추가 승인 버튼·폼 **없음**. 사용자 금지 사항 위반 없음.

경보는 건수 **위**에 붙고 건수를 대체하지 않으며, 음성 대조
(`TestAShadowScaleThatSeparatedSomethingRaisesNoAlarm`, 165 대 4 → 경보 없음)가 있다.
상시 점등이 아니다.

### C1 (P2) 새 함수가 **기존 함수의 doc 주석 안**에 삽입됐다

`internal/console/signals.go:544-575`. `signalsBandAlarm`이 `signalsTallyAlarm`의
doc 블록 **중간에** 들어가면서, 빈 줄이 없어 그 10줄이 통째로 새 함수의 doc이 됐다.

```
:544  // signalsTallyAlarm is one arithmetic contradiction as this screen words it.
:552  // It is rendered beside the counts rather than instead of them: …
:554  // signalsBandAlarm is the sentence a collapsed shadow tally is reported with here,
      …
:567  func signalsBandAlarm(t candidate.BandTally) string {
:576  func signalsTallyAlarm(a candidate.TallyAnomaly) string {   ← doc 주석이 없다
```

`go doc`은 `signalsBandAlarm`을 *"signalsTallyAlarm is one arithmetic contradiction…"*으로
설명하고, `signalsTallyAlarm`은 **무주석**이 된다. `go vet`도 `gofmt`도 잡지 않는다.
주석이 근거의 일부인 저장소에서 이것은 사소하지 않다. 고치는 것은 빈 줄 하나다.

---

## D. R3 — `Collapsed()`가 빈 `Crossed`에서 false

### D1. 변이로 확인 — 되돌리면 RED다

```
편집: band.go  `if t.Measured < 1 || len(t.Crossed) == 0 {` → `if t.Measured < 1 {`

$ rtk proxy go test ./internal/candidate/ -count=1 -v
    bandscale_test.go:518: a tally with no bands at all reports itself collapsed (crossed = map[]);
        the sentence claims every measured record produced the same crossings when there were
        no crossings for any of them to produce
    bandscale_test.go:523: and it carries an alarm sentence: "ALARM near_high: all 2 measured
        record(s) produced the same crossings, …"
--- FAIL: TestATallyWithNoScaleIsNotReportedAsCollapsed
```

`internal/console`·`cmd/tossctl`는 이 변이에서 초록이므로, 그 테스트가 이 한 줄을 **혼자**
지키고 있다. 되돌린 뒤 `band.go` = `301d1c01…6cff48`(착수 기록과 동일).

### D2. 새 구멍은 만들지 않는다 — 구성으로 증명된다

프로덕션에서 `BandTally`를 만드는 곳은 `band.go:514`의 `TallyBands` **하나뿐**이고
(`grep -rn "BandTally{" --include=*.go | grep -v _test`로 확인: `watch.go:578`·`:744`는
`map[VetoCode]BandTally{}` 컨테이너다), `TallyBands`는 `Crossed`를 `BandsFor(code)`의
밴드마다 하나씩 **반드시 seed**한다(`:520-522`). 따라서

> `len(t.Crossed) == 0` ⟺ `BandsFor(code)`가 비었다 ⟺ **그 사유에는 눈금이 없다**

이고, 눈금이 **있는** 사유에서는 이 disjunct가 결코 발화하지 않는다.
**밴드가 있는 사유에서 정당한 붕괴를 놓치는 경우는 없다.** `watch.go:744`가 담는 두
사유는 `seen_late`·`extended`이고 둘 다 오늘 눈금이 있다.

잔여물(정보): `Measured>0`이면서 `Crossed:nil`인 **손으로 만든** tally는 이제
not-collapsed가 된다. 그런 것을 만드는 곳은 테스트뿐이고, `band_alarm_test.go:44-54`의
`bandTally()`는 `BandsFor`에서 seed하므로 해당 없다.

---

## E. 기존 가드의 fixture를 고쳤다는 건 — **전제가 사실과 다르다**

**기존 파일은 하나도 약화되지 않았다.** `git diff --stat`이 싣는 테스트 파일은 둘뿐이다.

```
cmd/tossctl/candidate_test.go     |  21 ++-      ← §2 (JSON 순서·밴드 열)
internal/candidate/band_test.go   |  83 +++++--- ← §0 (선택자·어휘 확대)
```

`internal/candidate/consumer_guard_test.go`는 **diff에 없다.**
`verdictReaders` 허용목록은 **넓혀지지 않았다.** `internal/console`에서 바뀐 것은
`signals.go`와 `templates_signals.go` 둘뿐이고, 테스트 파일은 **신규**
`band_alarm_test.go` 하나다.

구현자가 말한 "fixture"는 그 **신규 파일 자신의** `marketWithBands`다. 커밋된 적 없는
초안이 `candidate.Verdict`를 참조했다가 가드에 걸렸고, 목록을 넓히는 대신 fixture에서
뺐다. **없앤 것이 아니라 처음부터 넣지 않은 것이다.**

**그리고 그것이 옳은 선택이다.** `TestOnlyTheListedFilesCanNameTheChaseVerdict`가
지키는 것은 *"chase 독자가 느는 것은 누군가 내린 결정이어야 한다"*이므로, chase가
필요 없는 파일을 목록에 넣는 것이 가드를 약화하는 쪽이다.

**지금도 검증하는가 — 그렇다.** 위 C에서 확인한 대로 이 테스트는 실제 HTTP GET으로
실제 템플릿을 통과시키고, 밴드 블록은 `SignalsMarket.Bands`만으로 렌더되므로
(`templates_signals.go:71`) verdict 없는 fixture가 프로덕션 경로를 **그대로** 탄다.
건수(`131`)와 사유명(`extended`)이 남아 있는지도 함께 단언한다.

**판정: 정당한 단순화다. 가드 회피가 아니다.**

---

## F. 회귀 없음

| 확인 | 명령 | 결과 |
|---|---|---|
| `extended`가 여전히 veto하지 않는가 | `grep -rn ExtendedGainPct --include=*.go .` | **비테스트 코드의 대입 0건.** 읽기는 `veto.go:1049`(`thresholdReason`)·`:1069`(`GainExceeds`) 둘뿐이고 선언은 `:884`. 아무도 채우지 않으므로 `AssessExtended`는 항상 `UnmeasuredVeto`다 |
| `ExtendedGainBands` | `band.go:133` | `{"0","2","4","6","8","10","20","30","50","100"}` — §1.1 확정 그대로 |
| `SeenLatePercentileBands` | `band.go:73` | `{"50","70","80","90","95"}` — **무변경**(D14) |
| 안전 경로 무접촉 | `git status --short \| grep -E "internal/(risk\|execgw\|exitpolicy\|trading)"` | **NONE** |
| FLM | `python3 tools/logic-map/check_analysis.py --change refine-extended-shadow-bands` | `evidence complete or diff-proven exempt` |

**콘솔 함수 맵 2개를 코드와 대조했다 — 내용이 맞다.**

- `internal-console--signalsbandalarm`: B1 = `!t.Collapsed()` → `""`. callees
  `t.Collapsed`(signals.go:568)·`strconv.Itoa`(:571) — 줄 번호까지 실제와 일치.
  두 테스트가 두 경로에 각각 붙어 있고 음성 대조가 B1이다.
- `internal-console--signalsbandtalliesfrom`: B1~B4가 실제 함수의 range·`continue`와
  일치하고, *"The `Alarm` assignment is inside B1's body and has no branch of its own"*이
  코드(`Alarm: signalsBandAlarm(tally)`가 복합 리터럴의 필드)와 맞다.
- `internal-candidate--bandtally.collapsed`의 B1이 새 disjunct와 그 사유를 담고 있고
  `TestATallyWithNoScaleIsNotReportedAsCollapsed`를 required test로 지목한다 — D1에서
  실제로 그 테스트가 혼자 RED가 되는 것을 확인했다.

**복구 대조 (전체 12개 파일, `sha256sum -c`)**

```
$ sha256sum -c baseline-sha256.txt
internal/candidate/watch.go: OK          internal/candidate/band.go: OK
internal/candidate/band_test.go: OK      internal/candidate/bandguard_test.go: OK
internal/candidate/bandscale_test.go: OK internal/console/signals.go: OK
internal/console/templates_signals.go: OK internal/console/band_alarm_test.go: OK
cmd/tossctl/candidate.go: OK             cmd/tossctl/candidatebands.go: OK
cmd/tossctl/candidate_test.go: OK        cmd/tossctl/candidatebands_test.go: OK
$ rtk proxy git status --short           → §5 종료 시점과 동일. probe 파일 없음
```

**복구 후 트리 재검증**

```
$ rtk proxy go vet ./...                 → 출력 없음 (exit 0)
$ rtk proxy go test ./... -count=1       → ok 51 / 51, FAIL 0
$ $(go env GOROOT)/bin/gofmt -l .        → 출력 없음
```

---

## 자체 발견 요약

| # | 등급 | 위치 | 내용 |
|---|---|---|---|
| **S1** | **P1 — 랜딩 차단** | `internal/candidate/bandguard_test.go:531-534`, `:502-511` | void 함수 면제. B1·B2·B3·B9가 통과하고, B1은 헬퍼 없는 **한 함수**다. B2·B3는 spec Scenario "한 홉 건너서도 도달할 수 없다"의 **문자 그대로의 사례**다 |
| **S2** | **P1** | `bandguard_test.go:166-200`, `:618-635` | 타입 별칭(B4)·임베딩 승격(B5)이 기록성을 지운다. 별칭은 void 함수조차 필요 없다 |
| **S3** | P2 | `issues.md` I20 #4 | 리플렉션(B6)이 통과한다. I20의 "reflect는 막는다"는 그 철자에만 맞다 |
| **S4** | P2 | `bandguard_test.go:53` | `verdictProducers = 12` 하한이 B1류를 세지 않는다(다른 축을 잰다) |
| **S5** | P2 | `internal/console/signals.go:544-575` | `signalsBandAlarm`이 `signalsTallyAlarm`의 doc 블록 안에 삽입 — 후자가 무주석이 됐다 |
| **S6** | P2 | `issues.md` I20 #1 | 다른 패키지 항의 완화 문구가 위험을 과소 서술. 허용목록 10개 파일에는 방어가 **전혀** 없고 그중 하나가 이 change가 고친 `signals.go`다 |
| **S7** | P3 | `bandguard_test.go:134` | `Bands`·`Quantiles`도 기록을 담지만 `pinnedRecordFields`에 없다 |

### S1의 실패 시나리오 (구체)

1. 누군가 `assessOne`이 길다며 밴드 관련 배선을 `func annotateBands(v *Verdict)`로 뺀다.
   리뷰에서 **자연스럽게 읽힌다** — 함수를 짧게 하는 리팩터이고, 아무것도 반환하지 않으며,
   §0이 만든 "조립과 판정을 분리한다"는 서사와 **모양이 같다.**
2. 그 함수 안에서 `if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() }`가
   한 줄 늘어난다. 어휘 가드는 이 함수를 보지 않고(판정 타입을 반환하지 않는다),
   새 가드는 보고서 통과시킨다(아무것도 반환하지 않는다).
3. `go test ./...` 51개 초록, `go vet` 무출력, `make gate` 통과. **아카이브된다.**
4. 오늘의 데이터로는 `6`을 넘는 후보가 없으므로 스캔 리포트도 `/signals`도 달라지지 않는다.
   승인된 적 없는 임계가 코드에 들어간 사실은 **어떤 표면에도 나타나지 않는다.**
5. 시장이 움직여 후보가 `6`을 넘는 날, `extended`가 veto를 낸다. 그 veto의 근거는
   히스토그램의 눈금이고, 그 눈금은 이 change가 *"아무것도 결정하지 않는다"*는 근거로
   출처 없이 정한 값이다. **spec의 첫 SHALL이 그날 거짓이 된다.**

A4가 그 5단계 중 1~3을 **실행으로** 보였다.

---

## 제안 (구현자가 형태를 고르고 이유를 적는다 — 이 리뷰는 형태를 지시하지 않는다)

S1·S2가 같은 뿌리를 갖는다는 것만 적어 둔다: `mayReadRecords`는 **결과 타입**으로
"이 함수가 답을 밖에 낼 수 있는가"를 근사하는데, Go에서 답이 나가는 통로는 결과만이
아니다. 최소한 다음 셋이 규칙에 들어가야 B1~B9가 닫힌다.

- **결과가 없는 것은 면제가 아니다.** 기록을 읽는 함수는 (a) 포인터·참조 인자에 쓰지
  않고 (b) 패키지 변수에 쓰지 않아야 면제된다. `assessOne`이 `&v`를 넘기는 것 자체를
  변환으로 세는 것이 가장 짧다 — 기록을 담은 값의 **주소**를 넘기는 호출은 나르기가 아니다.
- **기록 타입의 별칭·정의 타입을 기록으로 센다.** `GenDecl`/`TypeSpec` 한 번 훑어
  `shadowRecordTypes`를 닫힘까지 넓히면 B4·B9가 닫힌다.
- **임베드된 기록의 승격 접근을 기록 읽기로 센다.** 임베드 필드는 이미 `fields`에
  있으므로, `isRecordExpr`가 "임베드로 기록을 담은 타입의 값"도 기록으로 보게 하면 된다.
  `CompositeLit`의 원소도 인자와 같게 취급해야 한다.
- 리플렉션(S3)과 다른 패키지(S6)는 **이 change에서 닫으라고 요구하지 않는다.**
  다만 I20 #4의 문장은 사실에 맞게 고쳐야 한다.

**전이적 검사로 돌아가라는 뜻이 아니다.** §5의 판단(국소 규칙, 예외 0개)은 옳다고
본다 — 틀린 것은 "밖으로 나가는 통로"의 정의이고, 그것은 `retire-gainers-source` 이래
반복된 *"형태의 판단은 옳았고 정의가 좁았다"*의 다음 사례다.

---

## `make gate` 결과 전문

기록(이 절)까지 끝낸 뒤 `make sdd-sync` → `make sdd-check` → `make gate`를 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain), exit=0 |
| `make sdd-check` | **exit=0** — logic-map 15, sdd 19, sdd-history 15, pm 6 전부 OK |
| `make gate CHANGE=refine-extended-shadow-bands` | **FAIL — 2/8, 미완료 태스크 1건(`5.5`)** |

```
bash tools/gate.sh refine-extended-shadow-bands
GATE: refine-extended-shadow-bands
repo: /mnt/D/project/axipient/TossOS

==> 1/8 tasks.md 확인
OK: openspec/changes/refine-extended-shadow-bands/tasks.md

==> 2/8 미완료 태스크 확인
미완료 태스크 1 건:
218:- [ ] 5.5 §5 수정분에 대한 **2차 독립 리뷰**. 범위는 §5이 만든 diff. 리뷰어는 5.1의
GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다

make: *** [Makefile:75: gate] Error 1
```

**§5.5만 미완료다. 다른 단계에서 걸린 것은 없다.** 나머지 일곱을 손으로 각각 돌렸다.

```
1/8 tasks.md 확인          OK
2/8 미완료 태스크 확인      FAIL — 5.5 하나뿐
3/8 review.md 존재         OK — 이 파일
4/8 Function Logic Map     OK — evidence complete or diff-proven exempt (exit=0)
5/8 make sdd-check         OK — exit=0
6/8 make test              OK — exit=0
7/8 make vet               OK — exit=0
8/8 make validate          OK — Totals: 25 passed, 0 failed (25 items)
python3 tools/pm/generate_master_tracker.py --check
                           [pm] hierarchy and generated trackers are current
```

**5.5의 체크박스는 리뷰어가 채우지 않는다.** §5가 §4.4에 대해 세운 선례("체크는
*수행되고 기록됐다*는 뜻이지 *통과했다*가 아니다")를 따르면 채울 수도 있지만, 이 리뷰의
판정이 **랜딩 불가**이므로 열어 두는 편이 상태를 정직하게 말한다. 닫는 방법은 §6을
만들어 S1·S2를 처리하고 3차 리뷰를 받는 것이다 — §5가 R1·R2에 대해 그랬던 것과 같다.

> 이 절을 쓰면서 `.md`를 편집했으므로 fingerprint가 다시 stale이다. 직후
> `make sdd-sync` → `make sdd-check`를 한 번 더 연속 실행했다(둘 다 exit=0).
> 위 게이트 결과는 그 사이에 코드가 바뀌지 않았으므로 그대로 유효하다.

---

## 최종 판정

**랜딩 불가.**

**사유 — S1(P1).** 이 change의 존재 이유는 밴드에서 판정으로 가는 경로를 **구조로**
막는 것이고, spec delta는 그것을 *"도달할 수 없을 것"*이라는 SHALL과 네 개의 Scenario로
적었다. §5.1은 그 경로의 한 변종(기록을 읽고 **다른 타입을 반환하는** 함수)을 닫았고,
그것은 실제로 닫혔다 — 여섯을 재현해 확인했다. 그러나 **아무것도 반환하지 않고
부수효과로 답을 내보내는 함수**는 규칙이 명시적으로 면제하고 있으며, 그 형태로
**헬퍼도 홉도 없이 한 함수**가 veto를 결정하는 것을 A3·A4에서 실행으로 보였다.
`checked 12`, `record readers` 정상, `go vet` 무출력, 51개 패키지 초록.

**세 번째가 있었다.** 2026-07-28은 선택자, 2026-07-29는 한 홉, 오늘은 결과 타입이다.
셋 다 같은 성질의 실패다 — 검사가 **함수의 표면 어딘가**로 위험을 근사했고, 위험은
근사가 보지 않는 다른 표면으로 나갔다. 그리고 셋 다 **검사가 통과하는 모습**으로
나타났다. spec이 그 성질을 이미 문장으로 적어 두었다: *"빠진 사실은 검사가 통과하는
모습으로 나타난다."*

**부수 사유 — S2(P1).** 별칭과 임베딩은 void 함수조차 필요 없이 한 홉에 통과한다.

**닫힌 것은 정직하게 적는다.** R2(§5.2)는 spec Scenario와 정확히 일치하고, 실제 HTTP
경로로 렌더되며, **UI 마찰이 전혀 없다.** R3(§5.3)의 판단은 옳고 새 구멍을 만들지
않는다는 것이 구성으로 증명된다. §5.1이 실제로 닫은 열 형태도 재현으로 확인했다.
E의 fixture 변경은 가드 회피가 아니라 정당한 단순화다. 회귀는 없다.

**남은 것은 §5.5 하나가 아니다.** §5.5는 이 절로 수행됐고, 그 결과가 S1·S2다.
tasks에 §6이 필요하다 — §5가 §4.4에 대해 그랬던 것처럼.

---

# §6 수정 (구현)

- 날짜: 2026-07-29
- 역할: **Teammate(구현자).** 범위는 tasks §6.1~§6.6. §6.7(3차 리뷰)은 열어 둔다.
- base-commit: `cat base-commit.txt` → `ad16f56d7db04871699e060f120f249518fbee77`,
  `git log --oneline -1` → `ad16f56`. **일치한다.**
- 방법: 변이는 전부 **편집으로 넣고 편집으로 되돌렸다.**
  `git checkout --`·`git restore`·`git stash`는 **한 번도 쓰지 않았다.**
  착수 시점에 12개 파일의 sha256을 기록하고 종료 시점에 `sha256sum -c`로 대조했다.
  임시 probe는 `internal/candidate/zz_probe.go` 하나이며 매 측정 후 삭제했다.
- 커밋·push 없음. `tossctl` 실행 없음. 주문 명령 없음. 운영 store를 읽지도 쓰지도 않았다.
- **§6이 실제로 바꾼 파일은 둘뿐이다**: `internal/candidate/bandguard_test.go`(§6.1·§6.2·S7),
  `internal/console/signals.go`(§6.5). 새 파일 하나: `internal/candidate/bandbehaviour_test.go`(§6.3).
  `band.go`·`watch.go`는 **바이트 단위로 무변경**이다(아래 sha256 대조).

## 0. 요약 — 무엇이 닫혔고 무엇이 남았는가

| | 결과 |
|---|---|
| §6.1 S1 void 면제 | 닫음. **그리고 같은 결함이 나머지 두 면제에도 있었다**(I21) — 셋 다 닫음 |
| §6.2 S2 별칭·임베딩·CompositeLit | 닫음 |
| §6.3 행동 속성 | 새로 만듦. 두 변이(`annotate(&v)`·`Crossed("0")`)에서 RED 확인 |
| §6.4 I20 | 실측으로 다시 씀. **Manager 서술 하나가 틀렸다**(§7의 1번) |
| §6.5 doc 주석 | 고침. **"빈 줄 하나"로는 고쳐지지 않는다**(§7의 4번) |
| S7(P3) `pinnedRecordFields` | 둘 → 넷. 넷 각각 변이로 확인 |
| 남은 구멍 | 둘. I20의 1번(리플렉션 + 눈금 비의존 필드)과 2번(다른 패키지) |

---

## 1. §6.1·§6.2 — 고치기 전에 변이가 통과한다는 실측

우회 일곱(B1·B2·B3·B4·B5·B6·B9)을 `zz_probe.go` 하나에 전부 넣고 돌렸다.
B4·B5는 리뷰가 쓴 대로 **판정을 반환하는 함수**(`zzB4B5assess`)에서 읽게 했다 —
`assessOne`이 두 가드에 대해 내미는 표면과 같다.

```
$ rtk proxy go vet ./internal/candidate/                            → 출력 없음
$ rtk proxy go test ./internal/candidate/ -run '<네 가드>' -v -count=1
=== RUN   TestAShadowBandCannotBeReadAsAVeto
--- PASS: TestAShadowBandCannotBeReadAsAVeto (0.00s)
=== RUN   TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
    band_test.go:394: checked 13 verdict-producing functions
--- PASS: TestNoFunctionThatProducesAVerdictCanSeeAShadowBand (0.01s)
=== RUN   TestNoFunctionTurnsAShadowRecordIntoSomethingElse
    bandguard_test.go:671: record readers: [BandTally.Collapsed BandTally.CollapsedAlarm
        MeasureExtendedBand MeasureSeenLateBand ShadowBand.Crossed ShadowBand.Reason
        TallyBands zzB1annotate zzB2note zzB3note zzB9stash]
--- PASS: TestNoFunctionTurnsAShadowRecordIntoSomethingElse (0.01s)
=== RUN   TestOnlyTheListedFilesCanNameTheChaseVerdict
--- PASS: TestOnlyTheListedFilesCanNameTheChaseVerdict (0.23s)
PASS
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/candidate	0.252s
```

**네 가드 전부 PASS.** readers 목록이 `zzB1annotate`·`zzB2note`·`zzB3note`·`zzB9stash`를
**독자로 인식하고서 통과시킨다**(void 면제). B4·B5·B6은 인식조차 하지 않는다.
2차 리뷰의 A2 표와 정확히 일치한다.

### 1.1 §6.1이 무엇을 고쳤는가

`internal/candidate/bandguard_test.go`

1. **`mayReadRecords`에서 "결과가 없으면 허용"을 삭제했다.** 완화가 아니라 삭제다 —
   Go에서 답이 나가는 통로를 열거하는 것은 한 층 아래에서 같은 실수를 반복하는 것이고,
   이 저장소는 그 실수를 이미 세 번 했다. 값을 치를 것도 없었다: **이 패키지에는 기록을
   읽으면서 아무것도 반환하지 않는 함수가 하나도 없다.**
2. **`carriesRecordsSafely`의 공허한 참을 없앴다.** 결과 목록이 비면 루프가 한 번도 돌지
   않아 "전부 기록"이 참이 됐다. void callee는 기록을 **나를 곳이 없으므로** 나르기가
   아니다. `sh.voidFuncs`로 판정한다.

### 1.2 §6.2가 무엇을 고쳤는가

3. **`recordTypeClosure`** — 기록 타입 이름을 열거에서 **닫힘**으로 바꿨다. 어떤 타입이
   기록인지를 **철자**로 판정하던 것이 S2의 원인이다. 이제 ① 기록 위에 선언된 별칭·정의
   타입(포인터·슬라이스·맵·채널을 통과해서), ② 기록을 **임베드한** 구조체가 고정점까지
   기록으로 포함된다. `type a = ShadowBand; type b = a`도 잡힌다.
   **명명된 필드로 기록을 갖는 것은 기록이 아니다** — 그렇게 하면 `Verdict`가 기록이 되고
   요구사항이 보호하는 조립 자체가 금지된다.
4. **walk에 `*ast.CompositeLit` case 추가.** `box{v.ExtendedBand}`는 어느 case에도 걸리지
   않았다. 리터럴의 타입이 기록이면 나르기(허용), 아니면 원소가 기록인지 본다.
   **키가 기록 필드면 조립이므로 건너뛴다** — `Verdict{ExtendedBand: b}`가 실패하면 안 된다.

### 1.3 고친 뒤 — 여섯 전부 RED, 각각 이름으로

```
$ rtk proxy go test ./internal/candidate/ -run TestNoFunctionTurnsAShadowRecordIntoSomethingElse -count=1
zz_probe.go:zzB1annotate  takes a shadow record apart (reads Crossed out of one)
zz_probe.go:zzB2note      takes a shadow record apart (reads Crossed out of one)
zz_probe.go:zzB3note      takes a shadow record apart (reads Crossed out of one)
zz_probe.go:zzB4assess    takes a shadow record apart (reads Measured, Value out of one)
zz_probe.go:zzB5assess    takes a shadow record apart (reads Measured, Value out of one)
zz_probe.go:zzB9stash     takes a shadow record apart (hands one to zzB9note)
```

**프로덕션 오탐 0건.** readers 목록은 기준선 일곱 그대로이고 그중 아무도 실패하지 않는다.
예외 목록·허용목록은 **한 줄도 추가하지 않았다.**

### 1.4 내가 새로 만든 우회 일곱 — 그중 셋이 통과했다

리뷰 표를 재현하는 것으로 끝내지 않았다. 여덟을 새로 만들어 넣었다.

| # | 형태 | §6.1·§6.2 직후 | 최종 |
|---|---|---|---|
| N1 | `assessOne` 안의 클로저가 캡처한 지역변수에 쓴다 | RED | RED |
| N12 | `func (v *Verdict) annotate()` — **`*Verdict`의 void 메서드** | RED | RED |
| N15 | `var p interface{ Crossed(string) bool } = v.ExtendedBand` | RED | RED |
| N16 | `map[string]ShadowBand{"e": v.ExtendedBand}["e"].Crossed(…)` | RED | RED |
| **N13** | **`func (b ShadowBand) decide(v *Verdict)`** — 기록 메서드가 verdict를 쓴다 | **PASS** | RED |
| **N14** | **`func (b ShadowBand) stash()`** — 기록 메서드가 패키지 var에 쓴다 | **PASS** | RED |
| **N17** | **`func hand(s Sighting, v *Verdict) (ShadowBand, ShadowBand)`** | **PASS** | RED |

**N13·N14·N17이 이 라운드의 실질적 발견이다**(issues I21). 셋 다 §6.1을 문자 그대로
구현한 뒤에도 통과했다. 이유가 S1과 **같다** — `mayReadRecords`의 면제가 셋인데
(수신자·결과·void) S1이 지목한 것은 그중 하나뿐이고, 나머지 둘도 **함수의 표면 하나를
보고 "그러니 답이 못 나간다"고 근사**한다. N17이 특히 조용하다:

```go
func hand(s Sighting, v *Verdict) (ShadowBand, ShadowBand) {
    a := MeasureSeenLateBand(s)
    if a.Crossed("90") { v.Chase.Extended = RaisedVeto() }
    return a, a
}
// assessOne: v.SeenLateBand, v.ExtendedBand = hand(…, &v)
```
호출부는 **아무것도 읽지 않고 아무 금지 이름도 쓰지 않는다.**

**그래서 면제에 조건을 하나 더 붙이지 않고 질문을 바꿨다.** `writesOutward`가
*"이 함수가 결과가 아닌 통로로 무엇을 내보낼 수 있는가"*를 묻는다 — 기록을 이름하지 않는
참조형 인자(포인터·슬라이스·맵·채널·func·인터페이스, 그리고 그 위에 선언된 이름)와
패키지 var 대입(`=`·`++`·`<-`) 둘이다. `[]ShadowBand`는 통로가 **아니다**(원소가 기록이면
기록을 나르는 것이다) — 그래서 `TallyBands`가 예외 없이 통과한다.
**예외 목록은 여전히 0개다.**

---

## 2. §6.3 — 형태와 그것이 실제로 덮는 것

### 2.1 형태

`internal/candidate/bandbehaviour_test.go`(신규). 테스트 둘.

```
TestChangingTheScaleChangesNoVerdict                 — assessOne 직접, 후보 6 × 눈금 6
TestChangingTheScaleChangesNoVerdictAcrossAWholeScan — Cycle 두 번으로 만든 실제 store에
                                                       Assess, 후보 5 × 눈금 6
```

`onScale`이 `ExtendedGainBands`·`SeenLatePercentileBands`를 잠시 갈아끼우고 되돌린다.
프로덕션 코드가 **실제로 읽는 것을 그대로** 읽게 하는 것이 이 형태의 힘이다 — 어떤 경로로
읽든 상관하지 않는다. 눈금 변형 여섯:

```
the production scale · no scale at all · a single edge ·
a grid that shares no value with the production scale ·
one integer at a time from -30 to 130 · the production scale in reverse
```

프로덕션 눈금의 **모든 값이 최소 한 변형에서 빠지고 최소 한 변형에 들어간다.** 그래서
어떤 숫자로 `Crossed(x)`를 밀어 넣어도 판정이 변형 간에 달라진다.

**비공허성 단언 셋**(issues I23) — 이것이 없으면 이 테스트는 통과하는 모습으로 아무것도
검사하지 않는다: ① 모든 fixture의 두 밴드가 `Measured`, ② `Value`가 fixture가 주장하는
gain과 일치, ③ **눈금 변형이 실제로 기록을 바꾼 쌍이 1 이상**.

```
$ rtk proxy go test ./internal/candidate/ -run TestChangingTheScale -v -count=1
    bandbehaviour_test.go:222: 30 of 30 (fixture, variant) pairs recorded a different scale reading
--- PASS: TestChangingTheScaleChangesNoVerdict (0.00s)
--- PASS: TestChangingTheScaleChangesNoVerdictAcrossAWholeScan (0.05s)
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/candidate	0.050s
```

### 2.2 변이별 RED 확인표 — **전부 넣고, 돌리고, 되돌렸다**

`구문` = 네 구문 가드 중 하나라도 RED인가. `행동` = §6.3 둘 중 하나라도 RED인가.
변이는 언제나 `v.Chase.Extended = RaisedVeto()` 한 줄이고, 승인된 적 없는 숫자다.

| # | 변이 | vet | 구문 | 행동 | 판정 |
|---|---|---|---|---|---|
| **V1** | `func annotate(v *Verdict){ if v.ExtendedBand.Crossed("6") {…} }` + `assessOne: annotate(&v)` | 무출력 | **RED** | **RED** | 요구된 확인 ① |
| **V2** | 같은 변이, 밴드만 `"0"` — 오늘 KR 데이터가 실제로 넘는 값 | 무출력 | **RED** | **RED** | 요구된 확인 ② |
| V3 | 별칭 `type a = ShadowBand` + `take(v).Crossed("6")`를 판정 함수에서 | 무출력 | **RED** | **RED** | §6.2 |
| **V4** | **리플렉션** → `Crossings`: `reflect.ValueOf(*v).FieldByName("ExtendedBand").FieldByName("Crossings")` 순회 | 무출력 | PASS | **RED** | **행동만 잡는다** |
| **V5** | **unsafe 오프셋 산술** → `Crossings`: `unsafe.Add(unsafe.Pointer(v), zzOff)` (오프셋은 패키지 var) | 무출력 | PASS | **RED** | **행동만 잡는다** |
| V6 | unsafe, `unsafe.Pointer(&v.ExtendedBand)` | 무출력 | **RED**(`hands one to Pointer`) | — | 구문이 잡는다 |
| **V7** | **리플렉션** → `Value`: `…FieldByName("Value").String() == "9"` | 무출력 | PASS | **PASS** | **둘 다 못 잡는다 — I20 1번** |
| **V8** | **다른 패키지**: 2026-07-28의 그 한 줄을 `internal/console/signals.go:652`(`signalsRowFrom`)에 | 무출력 | PASS | **PASS** | **둘 다 못 잡는다 — I20 2번** |

V1의 실제 출력(발췌):

```
--- FAIL: TestChangingTheScaleChangesNoVerdict
    bandbehaviour_test.go:201: up seven: the veto verdict changed when the only thing that
        moved was the shadow scale (the production scale → no scale at all).
          verdict on the production scale: {… Extended:raised …}
          verdict on "no scale at all":    {… Extended:unmeasured (THRESHOLD_ABSENT) …}
--- FAIL: TestChangingTheScaleChangesNoVerdictAcrossAWholeScan
    bandbehaviour_test.go:310: 207940: the veto verdict off a real store changed when only
        the shadow scale moved …
```

V7·V8 실측 명령과 결과:

```
V7  $ rtk proxy go test ./internal/candidate/ -count=1        → FAIL 0건 (패키지 전체 초록)
V8  $ rtk proxy go vet ./...                                  → 출력 없음
    $ rtk proxy go test ./... -count=1                        → FAIL 0건, ok 51/51
```

### 2.3 그래서 §6.3이 실제로 덮는 축

tasks §6.3은 이 속성이 *"리플렉션·타입 별칭·포인터 인자·다른 패키지·unsafe를 전부 함께
덮는다"*고 썼다. **축이 다르다**(issues I22). 덮는 것은 경로의 종류가 아니라
**읽은 값이 눈금에 의존하는가**이다.

- 눈금 의존 읽기(`Crossed`·`Crossings`)면 **경로를 하나도 열거하지 않고 전부 잡는다** —
  리플렉션(V4)도 unsafe 오프셋 산술(V5)도 잡힌다. tasks의 주장이 여기서는 맞다.
- 눈금 **비의존** 읽기(`Value`·`Measured`·`Reason`)는 경로와 무관하게 못 잡는다(V7).
  눈금을 어떻게 바꿔도 그 값이 안 변하므로 판정도 안 변한다.
- **다른 패키지는 축과 무관하게 못 잡는다**(V8) — 속성이 `internal/candidate`의 판정을
  재기 때문이다.

구문 가드가 정확히 반대다: 눈금 의존성과 무관하게 **기록 값에 적용된 셀렉터**를 본다.
그래서 `Value` 읽기는 구문이 잡고 리플렉션은 행동이 잡는다. **구문 가드를 지우지 않은
이유가 이 한 문장이고**, 두 테스트의 doc 주석에 그대로 적었다.

---

## 3. §6.4 — I20 재작성의 근거

issues.md I20을 통째로 다시 썼다. **원문은 남기지 않았다** — 틀린 문장을 근거로 다음
사람이 안심하는 것이 그 항목이 막으려던 바로 그 일이기 때문이다.

| 원문 항목 | §6.4 실측 | 처리 |
|---|---|---|
| #1 다른 패키지 | V8로 재현. 통과한다 | **유지하되 위험 상향.** 허용목록 10개 파일에 방어 전무, 그중 하나가 `signals.go`, 그리고 그 화면이 임계 승인의 근거로 읽힌다. 후속 change, 등급 Normal |
| #2 테스트 파일 | 두 가드 모두 `_test.go` 제외를 코드로 확인 | **유지.** 다만 행동 속성에는 해당 없음(테스트 파일은 프로덕션 판정을 안 바꾼다) |
| #3 모호한 필드 이름 | `pinnedRecordFields` 넷 각각에 `type X struct{ <이름> int }` 변이 → 넷 다 `Fatalf` | **유지하되 방어 확대.** 둘 → 넷(S7) |
| #4 unsafe | V5·V6으로 재현 | **삭제.** 거짓이었고, 2차 리뷰가 말한 것보다 더 그렇다 — V6은 구문이, V5는 행동이 잡는다. **unsafe는 이제 넷 중 가장 잘 덮인다** |
| (없었음) void 함수·별칭·임베딩 | §6.1·§6.2가 닫음 | **닫힘으로 기록** |
| (없었음) **리플렉션 + 눈금 비의존 필드** | V7 — 두 층 모두 통과, 51개 패키지 초록 | **새 1번 항목** |

**새 1번의 성질을 완화하지 않되 정확히 적었다**: 이 경로가 만드는 것은 "눈금이 판정했다"가
아니라 **"승인되지 않은 임계를 손으로 박았다"**이고, 읽는 값 `Value`는 `Expansion`이 이미
갖고 있는 gainPct다. 즉 그림자 기록을 통째로 지워도 같은 방식으로 쓸 수 있는 결함이며,
그것을 막는 것은 이 change의 요구사항이 아니라 `ExtendedGainPct` 자체의 규율이다
(`TestMeasureExtendedBandNeverReadsTheVetoThreshold`가 그 규율의 절반을 이미 들고 있다).

---

## 4. §6.5 — doc 주석

`internal/console/signals.go`. `signalsBandAlarm`의 doc이 `signalsTallyAlarm`의 doc 블록
**중간에** 삽입돼 있었다. 고쳤고, 방법은 **빈 줄 하나가 아니다**(§7의 4번).

```
$ rtk proxy go build ./...   → BUILD-OK
```

지금은 각 함수 바로 위에 자기 doc이 있다. 문장은 한 글자도 바꾸지 않고 위치만 옮겼다.

---

## 5. 모든 변이의 넣기 / 결과 / 되돌리기 + sha256 대조

**변이를 넣은 파일은 셋뿐이다.**

| 파일 | 무엇을 넣었나 | 되돌린 방법 | 최종 sha256 |
|---|---|---|---|
| `internal/candidate/zz_probe.go` | 신규 probe 파일(B·N·V 계열 전부) | `rm` | **존재하지 않음** |
| `internal/candidate/watch.go` | `assessOne` 안 한 줄(`zzB1annotate(&v)` / `v = zzAssess(v)`) | 편집으로 삭제 | `1aa35693…7b0106` **= §6 착수값** |
| `internal/console/signals.go` | V8의 세 줄(`signalsRowFrom` 안) | 편집으로 삭제 | §6.5 수정만 남음(아래) |

```
$ sha256sum -c baseline-sha256.txt
internal/candidate/watch.go: OK          internal/candidate/band.go: OK
internal/candidate/band_test.go: OK      internal/candidate/bandscale_test.go: OK
internal/console/templates_signals.go: OK internal/console/band_alarm_test.go: OK
cmd/tossctl/candidate.go: OK             cmd/tossctl/candidatebands.go: OK
cmd/tossctl/candidate_test.go: OK        cmd/tossctl/candidatebands_test.go: OK
internal/candidate/bandguard_test.go: FAILED   ← §6.1·§6.2·S7이 의도적으로 고친 파일
internal/console/signals.go: FAILED            ← §6.5가 의도적으로 고친 파일
```

**§6가 바꾼 것은 정확히 그 둘이고, 나머지 열은 바이트 단위로 착수 시점과 같다.**
특히 `band.go`(눈금 선언)와 `watch.go`(`assessOne`)는 **무변경**이다.

```
$ git status --short     → probe 파일 없음. 신규는 bandbehaviour_test.go 하나
§6 종료 시점 sha256:
  internal/candidate/bandguard_test.go     c95f6dde7c876282efbd104f891bc3c663230c1dc8b37e238cb8be0fdcf115c1
  internal/candidate/bandbehaviour_test.go db6ba4905e9f4f43bdc2972e16c869878d67fc98f4aa995e792c713ef50d16f8
  internal/console/signals.go              de0adcf52d62132b6be3cd493155bd3cd0ce4c334200160511b474eb070d284a
  internal/candidate/band.go               301d1c0167f9618911bceaad6053f6d88ed4eee1d0f55f4a9381d74a446cff48 (무변경)
  internal/candidate/watch.go              1aa356933d8e9ad3ea2b98a4ded5d25972eb163c16897edff4240584ad7b0106 (무변경)
```

---

## 6. 검증 명령과 실제 결과 (§6.6)

| 명령 | 결과 |
|---|---|
| `go test ./... -count=1` | **ok 51 / 51, FAIL 0** |
| `go vet ./...` | 출력 없음, exit 0 |
| `make lint` | `go vet ./...` — 출력 없음, exit 0 |
| `$(go env GOROOT)/bin/gofmt -l .` | 출력 없음 |
| `go test -race ./internal/candidate/... ./cmd/tossctl/... ./internal/console/...` | ok 3/3 (14.6s / 34.4s / 28.8s) |
| `python3 tools/logic-map/check_analysis.py --change refine-extended-shadow-bands` | `evidence complete or diff-proven exempt`, exit 0 |
| `python3 tools/pm/generate_master_tracker.py --check` | `[pm] hierarchy and generated trackers are current` |

upstream 상속 테스트 회귀 없음(51개 패키지 전부 초록, 실패 0).

### 임계 확인 — 요구받은 대로 확인하고 보고한다

```
$ grep -rn "ExtendedGainPct" --include=*.go . | grep -v _test.go
./internal/candidate/band.go:270    (주석)
./internal/candidate/veto.go:882    (주석)
./internal/candidate/veto.go:884    ExtendedGainPct string          ← 선언
./internal/candidate/veto.go:1049   thresholdReason(th.ExtendedGainPct)   ← 읽기
./internal/candidate/veto.go:1069   e.GainExceeds(th.ExtendedGainPct)     ← 읽기
$ grep -rn "ExtendedGainPct *=" --include=*.go . | grep -v _test.go
NONE
```

**비테스트 코드의 대입 0건.** 아무도 채우지 않으므로 `AssessExtended`는 여전히 항상
`THRESHOLD_ABSENT`다. 이 change는 임계를 정하지 않았다.

### 눈금 최종값 — 바꾸지 않았다

```
band.go:73   var SeenLatePercentileBands = []string{"50", "70", "80", "90", "95"}
band.go:133  var ExtendedGainBands = []string{"0","2","4","6","8","10","20","30","50","100"}
```

§6.3의 변형은 테스트 안에서만 살고 `onScale`이 `defer`로 되돌린다. `band.go`의 sha256이
착수값과 같은 것이 그 증거다.

### FLM (§4.3 갱신)

`internal/console/signals.go`를 고쳤으므로 두 함수 맵의 AST 해시가 stale이 됐다.

```
[logic-map] internal-console--signalsbandalarm: AST source hash is stale
[logic-map] internal-console--signalsbandtalliesfrom: AST source hash is stale
```

`go run ./tools/logic-map`로 두 `ast.json`을 재생성하고 `risk_pattern_report.py`를 다시
돌렸다. **분기 구조는 바뀌지 않았다**(옮긴 것은 주석뿐이므로 map 본문은 그대로 유효하다).
재실행 결과 `evidence complete or diff-proven exempt`(exit 0).

`bandguard_test.go`·`bandbehaviour_test.go`는 이 change가 만든 **새 파일**이므로 tasks
서두의 규칙("새 파일의 함수는 전부 면제")대로 대상이 아니다.

### 안전 불변식

| 항목 | 결과 |
|---|---|
| LIVE 주문 side effect | 없음. 주문 명령 미실행 |
| `internal/risk`·`execgw`·`exitpolicy`·`trading` | `git status --short`에 **없음** |
| 손절·익절·사이징 | 무접촉 |
| 운영 토글 flip | 없음 |
| UI 마찰 | 없음. §6.5는 주석 위치만 옮겼고 `templates_signals.go`는 무변경(sha256 OK) |
| 임계 결정 | 없음(위 실측) |
| 커밋·push | 없음 |

---

## 7. 막힌 것 · tasks가 코드와 안 맞는 것

### 1. **tasks §6.3의 "다른 패키지를 덮는다"는 틀렸다** — 열한 번째

tasks §6.3과 착수 지시가 모두 이렇게 쓴다:

> 이 속성은 리플렉션·타입 별칭·포인터 인자·**다른 패키지**·unsafe를 **전부 함께** 덮고

**다른 패키지는 덮지 않는다.** V8로 실측했다 — 2026-07-28의 그 한 줄을
`internal/console/signals.go`의 `signalsRowFrom`에 넣으면 `go vet` 무출력, 네 구문 가드
PASS, **§6.3 두 테스트 모두 PASS**, 51개 패키지 전부 초록이다.

원인은 단순하고 고치기는 단순하지 않다. §6.3은 `internal/candidate`에 있고 `assessOne`과
`Assess`가 **낸** 판정을 잰다. 콘솔은 그 판정을 **받은 뒤에** 고치므로 candidate 쪽에서는
아무것도 달라지지 않는다. 닫으려면 소비 표면에서 같은 속성을 재는 테스트가 필요하고,
그것은 `internal/console`·`cmd/tossctl`에 각각 눈금 변형 하니스를 세우는 일이다.
**§6의 범위 밖이라고 판단했고 issues I20 2번에 등급 Normal 후속 change로 적었다.**
판단 근거: §6의 지시는 "§6.3이 무엇을 덮는지 실측으로 확인하고 못 덮는 것만 남겨라"이지
"전부 덮게 만들어라"가 아니었다.

### 2. **§6.1이 지목한 면제는 셋 중 하나였다**

tasks §6.1은 `bandguard_test.go:531-534`의 void 면제만 지목한다. 문자 그대로 고치면
**N13·N14·N17이 그대로 통과한다** — 수신자 면제와 결과 면제가 같은 결함을 갖기 때문이다.
세 면제 모두 "함수의 표면 하나"를 보고 답의 출구를 근사한다. `writesOutward`로 질문
자체를 바꿔 셋을 함께 닫았다. 예외는 0개다. **이것을 안 했으면 §6도 부분 수정이었다.**

### 3. **`#4(unsafe)는 거짓`이라는 판정은 맞지만 방향이 반대다**

tasks §6.4는 2차 리뷰를 인용해 *"#4(unsafe)는 거짓 — 리플렉션이 통과한다"*고 쓴다.
리플렉션이 통과한다는 것은 맞다(V7). 그러나 그 문장은 **unsafe가 여전히 잔여물**이라는
인상을 준다. 실측은 반대다: `unsafe.Pointer(&v.ExtendedBand)`는 구문 가드가 잡고(V6),
필드를 전혀 이름하지 않는 오프셋 산술은 행동 속성이 잡는다(V5). **unsafe는 I20의 넷 중
이제 가장 잘 덮인 항목이다.** 남은 것은 unsafe가 아니라 *"눈금에 의존하지 않는 필드를
리플렉션으로 읽는 것"*이고, 그것은 §5.1도 2차 리뷰도 적지 않았다.

### 4. **§6.5는 "빈 줄 하나"로 고쳐지지 않는다**

2차 리뷰 C1이 *"고치는 것은 빈 줄 하나다"*라고 썼다. 확인한 결과 **아니다.**
원래 배치는 `[A의 doc][B의 doc][func B][func A]`였다. 두 주석 블록 사이에 빈 줄을 넣으면
`B`는 자기 doc을 얻지만 `A`(아래에 선언됨)는 **여전히 무주석**이고, 앞 블록은 아무 선언에도
붙지 않는 떠 있는 주석이 된다. 실제로 필요한 것은 `A`의 doc 블록을 `func A` 위로
**옮기는 것**이다. 그렇게 했다.

### 5. `zz_probe.go`가 `checked`를 늘린다 — 측정 시 유의사항 (정보)

probe가 판정 반환 함수를 하나 더하면 `checked 12`가 `checked 13`이 된다. 하한
(`verdictProducers = 12`)은 **하한**이므로 실패하지 않는다. 2차 리뷰 S4가 지적한
*"하한이 다른 축을 잰다"*는 여전히 사실이고 §6에서 고치지 않았다 — B1류는 판정 타입을
반환하지 않으므로 어떤 하한으로도 세어지지 않는다. 그 축을 재는 것이 §6.3이다.

### 6. `make gate`는 §6.7이 열려 있으므로 2/8에서 실패한다 (예상됨)

아래 §8에 전문을 붙인다. 다른 단계에서 걸린 것은 없다.

---

## 8. `make sdd-sync` → `make sdd-check` → `make gate` 결과 전문

기록(§6 절·issues.md·tasks.md)까지 전부 끝낸 **뒤에** 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `openspec validate refine-extended-shadow-bands --strict --no-interactive` | `Change 'refine-extended-shadow-bands' is valid` |
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain), exit 0 |
| `make sdd-check` | **exit 0** — CodeGraph worktree fingerprint 일치, PM 최신 |
| `make gate CHANGE=refine-extended-shadow-bands` | **FAIL — 2/8, 미완료 태스크 1건(`6.7`)** |

```
bash tools/gate.sh refine-extended-shadow-bands
GATE: refine-extended-shadow-bands
==> 1/8 tasks.md 확인
OK: openspec/changes/refine-extended-shadow-bands/tasks.md
==> 2/8 미완료 태스크 확인
미완료 태스크 1 건:
282:- [ ] 6.7 **3차 독립 리뷰.** 범위는 §6의 diff. 리뷰어는 §6.3의 행동 속성을 **직접 뚫어라** —
GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다
```

**§6.7만 미완료다.** 나머지 일곱을 손으로 각각 돌렸다.

```
1/8 tasks.md 확인          OK
2/8 미완료 태스크 확인      FAIL — 6.7 하나뿐
3/8 review.md 존재         OK — 이 파일
4/8 Function Logic Map     OK — evidence complete or diff-proven exempt (exit 0)
5/8 make sdd-check         OK — exit 0
6/8 make test              OK — exit 0
7/8 make vet               OK — exit 0
8/8 make validate          OK — Totals: 25 passed, 0 failed (25 items)
```

### §5.5 체크박스를 구현자가 채웠다 — 판단과 근거

2차 리뷰어는 §5.5를 열어 두면서 *"닫는 방법은 §6을 만들어 S1·S2를 처리하고 3차 리뷰를
받는 것"*이라고 적었다. §6.1~§6.6이 그 처리이므로 채웠고, §4.4가 세운 선례와 같은 주석을
달았다 — **체크는 "수행되고 기록됐다"이지 "통과했다"가 아니다.** 판정("랜딩 불가")과
사유(S1·S2)는 그 절에 그대로 남아 있고 한 글자도 지우지 않았다.

열어 두면 게이트가 **같은 사유로 두 항목을 막는다**(§5.5와 §6.7). 랜딩을 막는 조건은
하나이고 그것은 §6.7이다. 이 판단이 틀렸다고 보면 되돌리는 것은 체크박스 하나다.

> 이 절을 쓰면서 `.md`를 편집했으므로 fingerprint가 다시 stale이다. 직후
> `make sdd-sync` → `make sdd-check`를 한 번 더 연속 실행했다(둘 다 exit 0).
> 위 게이트 결과는 그 사이에 **코드가 한 줄도 바뀌지 않았으므로** 그대로 유효하다.

---

## 9. 남은 위험

1. **I20 1번 — 리플렉션 + 눈금 비의존 필드.** 두 층 모두 통과한다(V7 실측). 성질은
   "눈금이 판정했다"가 아니라 "승인되지 않은 임계를 손으로 박았다"이고, 막는 것은
   `ExtendedGainPct`의 규율이지 이 요구사항이 아니다. **후속 판단 필요.**
2. **I20 2번 — 다른 패키지.** 두 층 모두 통과한다(V8 실측). 위험은 §5.1 원문이 쓴 것보다
   크다 — 그 표면이 임계 승인의 근거로 읽히기 때문이다. **후속 change, 등급 Normal.**
3. **I24 — `onScale`은 직렬 실행 전제에 기댄다.** 이 패키지에 `t.Parallel()`은 0건이고
   `-race` 초록을 확인했다. 누가 넣는 날 `-race`가 잡는다.
4. **§6.7 3차 독립 리뷰가 열려 있다.** 리뷰어에게 남기는 것: §6.3을 **직접 뚫어라.**
   V7·V8이 이미 뚫린 둘이므로 세 번째를 찾아야 한다. 특히 `writesOutward`의 참조형 판정이
   `context.Context`처럼 **다른 패키지에서 온 이름의 인터페이스**를 통로로 세지 않는다 —
   이 패키지에 선언된 이름만 닫힘 계산에 들어간다. 그 자리를 먼저 보라.

---

# 3차 독립 리뷰 (§6.7)

- 날짜: 2026-07-29
- 역할: **3차 독립 리뷰어.** 구현(§6)과 분리된 컨텍스트. 범위는 §6이 만든 diff —
  `internal/candidate/bandguard_test.go`(§6.1·§6.2·S7), `internal/candidate/bandbehaviour_test.go`(§6.3, 신규),
  `internal/console/signals.go`(§6.5), 그리고 §6.4·§6.8~§6.11이 고친 기록.
- base-commit: `cat base-commit.txt` → `ad16f56d7db04871699e060f120f249518fbee77`. `git log --oneline -1` → `ad16f56`. **일치.**
- 방법: 변이는 전부 **편집으로 넣고 편집으로 되돌렸다.** `git checkout --`·`git restore`·`git stash`는
  **한 번도 쓰지 않았다.** 착수 시 14개 파일의 sha256을 기록하고 종료 시 `sha256sum -c`로 대조했다(§5 참조).
  임시 probe는 `internal/candidate/zz_r3_probe.go` 하나이며 매 측정 후 `rm`했다.
- 커밋·push 없음. `tossctl` 실행 없음. 주문 명령 없음. 운영 store를 읽지도 쓰지도 않았다.
- **판정: 랜딩 불가.** 사유는 T1·T2·T3(아래). 요약하면 — **네 번째 계열을 찾았고, 그것은
  리플렉션도 unsafe도 다른 패키지도 필요로 하지 않는 평범한 Go다.**

## 0. 기준선

```
$ rtk proxy go test ./internal/candidate/ -run '<네 가드>|TestChangingTheScale' -v -count=1
--- PASS: TestAShadowBandCannotBeReadAsAVeto
    band_test.go:394: checked 12 verdict-producing functions
--- PASS: TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
    bandbehaviour_test.go:222: 30 of 30 (fixture, variant) pairs recorded a different scale reading
--- PASS: TestChangingTheScaleChangesNoVerdict
--- PASS: TestChangingTheScaleChangesNoVerdictAcrossAWholeScan
    bandguard_test.go:1056: record readers: [BandTally.Collapsed BandTally.CollapsedAlarm
        MeasureExtendedBand MeasureSeenLateBand ShadowBand.Crossed ShadowBand.Reason TallyBands]
--- PASS: TestNoFunctionTurnsAShadowRecordIntoSomethingElse
--- PASS: TestOnlyTheListedFilesCanNameTheChaseVerdict
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/candidate	0.347s
```

---

## A. 행동 속성을 뚫었다 — 시도 전부, 성공과 실패 각각

`구문` = 네 구문 가드 중 하나라도 RED. `행동` = §6.3 둘 중 하나라도 RED.
변이는 언제나 `v.Chase.Extended = RaisedVeto()` 한 줄이고 승인된 적 없는 숫자다.
**전부 넣고 · 돌리고 · 되돌렸다.**

| # | 시도 | vet | 구문 | 행동 | 결과 |
|---|---|---|---|---|---|
| M1 | `func zzAnnotate(v *Verdict){ if v.ExtendedBand.Crossed("6") {…} }` + `assessOne: zzAnnotate(&v)` | 무출력 | **RED** | **RED** | 막힘 — **요구된 확인 ①** |
| M2 | 같은 변이, 밴드만 `"0"` | 무출력 | **RED** | **RED** | 막힘 — **요구된 확인 ②** |
| **M3** | **리플렉션 → `Crossings`, 답을 `map[string]bool` 패키지 var에 래치** | 무출력 | PASS | **PASS** | **뚫림 (T1)** |
| M3b | M3에서 래치만 제거(대조군) | 무출력 | PASS | **RED** | 막힘 — 래치가 유일한 차이임을 증명 |
| **M4** | **리플렉션 없음.** `shadowBandsFor`(결과가 전부 기록 → 면제)가 `b.Crossed("6")`을 읽어 **패키지 레벨 func 값**에 넘기고, 그 클로저가 래치한다. `assessOne`은 `map[string]bool` 하나만 읽는다 | 무출력 | PASS | **PASS** | **뚫림 (T1+T2)** |
| M4b | M4에서 래치만 제거(대조군) | 무출력 | PASS | **RED** | 막힘 — **구문 층은 래치 유무와 무관하게 M4를 못 본다** |
| **M5** | **기록을 아예 안 읽는다.** `var zzEdge = ExtendedGainBands[3]`(패키지 레벨) + `assessOne`에서 `v.Expansion.GainExceeds(zzEdge)` | 무출력 | PASS | **PASS** | **뚫림 (T3)** |
| M5b | M5의 눈금을 init 고정 대신 **live**로 읽는다(`zzEdgeLive()`) | 무출력 | PASS | **RED** | 막힘 — 구문은 여전히 못 본다 |
| M6 | `writesOutward` 통로 (b): 기록 메서드가 **패키지 레벨 객체의 메서드**를 부른다(`zzMap.Store(…)`) | 무출력 | PASS | **RED** | 구문 뚫림, 행동이 잡음 |
| M7 | `writesOutward` 통로 (f): 기록 메서드가 **지역 포인터로 패키지 var를 별칭**한다(`s := &zzState; s.Hot = …`) | 무출력 | PASS | **RED** | 구문 뚫림, 행동이 잡음 |
| U1 | `unsafe.Pointer(&v.ExtendedBand)` | 무출력 | **RED**(`hands one to Pointer`) | **RED** | 막힘 |
| U2 | unsafe 오프셋 산술 → `Crossings`(오프셋은 패키지 var) | 무출력 | PASS | **RED** | 막힘(행동) |
| U3 | 리플렉션 → `Value` | 무출력 | PASS | **PASS** | 뚫림 — **I20 1번, 이미 기록됨** |
| V8′ | 2026-07-28의 그 한 줄을 `internal/console/signals.go:652`에 | 무출력 | PASS | **PASS** | 뚫림 — **I20 2번, 이미 기록됨** |
| E1 | `Crossed("100")`(extended 최상단 눈금) | — | — | **RED** | 막힘 |
| E2 | `Crossed("95")`(seen_late 최상단 눈금) | — | — | **RED** | 막힘 |
| E3 | `Crossed("50")`(seen_late 최하단 눈금) | — | — | **RED** | 막힘 |

M3·M4·M5는 각각 `go vet ./...` 무출력, `go test ./... -count=1` **ok 51/51 · FAIL 0**이다.

### A1. 공허성 검사 — **이 테스트는 장식이 아니다**

브리핑이 지목한 위험(§2.7의 KR 131건이 전부 정확히 0이었던 데이터)은 **이 fixture에는 없다.**

- `bandbehaviour_test.go:222`가 `30 of 30 (fixture, variant) pairs recorded a different scale reading`을
  로그한다. 6 fixture × (6−1) 변형 = 30, **전부** 다른 기록을 냈다.
- fixture의 gain은 `−20 · 0 · +3 · +7 · +9 · +120`이고 rank는 `5·30·60·85·92·98 / 100`이다.
  `percentileOf = (total−rank)/total×100`([veto.go:861])이므로 seen_late 백분위는 `95·70·40·15·8·2`다.
  → **`+120` fixture가 `ExtendedGainBands` 열 칸 전부를 넘고, `95` fixture가
  `SeenLatePercentileBands` 다섯 칸 전부를 넘는다.** 그리고 변형 `no scale at all`이 전부를 없앤다.
  따라서 **두 눈금의 모든 프로덕션 밴드 값이 최소 한 fixture에서 넘어지고 최소 한 변형에서 사라진다.**
- 실측으로 확인: 가장 성기게 시험될 세 경계 `Crossed("100")`·`Crossed("95")`·`Crossed("50")`을
  각각 밀어 넣었고 **셋 다 행동 RED**(E1·E2·E3).
- 프로덕션 눈금에 없는 값(`"125"`·`"6.5"` 등)은 `Crossed`가 언제나 false이므로 프로덕션에서도
  죽은 코드다 — 즉 **밴드 값 축에서는 미탐 구간이 없다.**

**결론: `Crossed(x)` 모양의 밀수에 대해 fixture 6 × 변형 6은 충분하다.** 이 축은 문제가 아니다.

### A2. 비교 대상 — **`Chase` 전체다**

`bandbehaviour_test.go:198`·`:307`이 `reflect.DeepEqual(have.chase, want.chase)`로 `Chase` 구조체
**전부**를 비교한다. `Chase = {Key, At, SeenLate, Extended, NearHigh}`([veto.go:417-424])이고 필드를
골라 보지 않는다. 일부 필드로 새는 경로는 **없다.** 이 축도 문제가 아니다.

(경계 하나만 적는다: `Verdict`의 나머지 필드 — `Sighting`·`Expansion`·`Range`·`Accelerations` —
는 비교되지 않는다. 그것들은 거부권 판정이 아니므로 SHALL의 대상이 아니고, 그쪽을 읽는 것은
구문 가드가 잡는다. 실측하지 않았고 결함으로 세지 않는다.)

### A3. **T1 (P1) — 래치가 행동 속성을 무력화한다. 네 번째 계열이다.**

`internal/candidate/bandbehaviour_test.go:161-187`(변형 루프)과 `:189`(`base := scaleVariants()[0].name`).

이 속성은 **평가가 (후보 입력, 눈금)의 순수 함수라는 전제** 위에 서 있다. 호출 사이에
상태를 나르는 코드가 하나라도 있으면 전제가 깨지고, 게다가 루프는 **프로덕션 눈금을 첫 번째로**
돌린다(`scaleVariants()[0]`). 그래서 처음 한 번만 계산하고 기억하는 코드는 **프로덕션 눈금의 답을
래치**하고, 이후 다섯 변형은 눈금이 아니라 메모를 읽는다. 속성은 참이 되고 화면은 이미 거짓이다.

**M3 실측** (`reflect.ValueOf(*v).FieldByName("ExtendedBand").FieldByName("Crossings")` 순회, 답을
`zzLatch[symbol|price]`에 최초 1회만 기록):

```
$ rtk proxy go vet ./internal/candidate/                → 출력 없음
--- PASS: TestAShadowBandCannotBeReadAsAVeto
--- PASS: TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
--- PASS: TestChangingTheScaleChangesNoVerdict
--- PASS: TestChangingTheScaleChangesNoVerdictAcrossAWholeScan
--- PASS: TestNoFunctionTurnsAShadowRecordIntoSomethingElse
--- PASS: TestOnlyTheListedFilesCanNameTheChaseVerdict
$ rtk proxy go test ./... -count=1                      → ok 51/51, FAIL 0
```

**대조군 M3b** — 같은 읽기, 같은 쓰기, **래치만 제거**:

```
--- FAIL: TestChangingTheScaleChangesNoVerdict
    bandbehaviour_test.go:201: up seven: the veto verdict changed when the only thing that
        moved was the shadow scale (the production scale → no scale at all).
--- FAIL: TestChangingTheScaleChangesNoVerdictAcrossAWholeScan
```

읽기는 살아 있고 눈금 의존이다. **숨긴 것은 래치 하나다.**

> 실측 중 얻은 부수적 확인: 첫 대조군 시도(`hot, seen := zzLatch[key]`를 남긴 채 재계산 분기만 연
> 형태)는 **여전히 PASS**였다. 맵에서 꺼낸 값을 시드로 쓰는 것만으로도 래치가 성립하기 때문이다.
> 래치는 `sync.Once`나 명시적 캐시일 필요조차 없다.

**실패 시나리오(구체).** 누군가 "그림자 판정을 후보·가격 단위로 캐시한다"는 지극히 평범한
최적화를 넣는다. 그 순간 §6.3은 *"입력이 안 바뀌면 판정이 안 바뀐다"*는 동어반복이 되고,
그 사실은 **테스트가 통과하는 모습**으로 나타난다 — 이 change가 세 번 겪은 그 모양 그대로다.
`issues.md` I23이 "비공허성 단언이 없으면 아무 말도 하지 않는다"고 적은 것은 맞지만, 세 단언
(`Measured`·`Value` 일치·`moved ≥ 1`)은 **기록이 움직였는지**만 보고 **판정이 기록을 따라
움직일 수 있는 상태인지**는 보지 않는다. M3에서 셋 다 만족하고(30/30) 속성은 여전히 거짓이었다.

### A4. **T2 (P1) — `writesOutward`의 "통로"가 열거이고, 열거가 새고 있다**

`internal/candidate/bandguard_test.go:747-780`.

`mayReadRecords`의 doc(`:872-874`)은 이렇게 쓴다 — *"Enumerating the ways out is the same mistake
one level down … So the exemption is gone rather than qualified."* 그런데 `writesOutward`의 doc
(`:741-742`)은 *"Two channels are recognised"*로 시작한다. **한 층 아래에서 같은 열거를 하고 있고,
셋이 샌다.** 셋 다 실측했다.

| 통로 | 형태 | 왜 안 보이나 | 구문 |
|---|---|---|---|
| (a) 패키지 레벨 **func 값** 호출 | `zzHook(key, b.Crossed("6"))` | `CallExpr`는 `AssignStmt`도 `IncDecStmt`도 `SendStmt`도 아니다([:757-771]) | **PASS(뚫림)** |
| (b) 패키지 레벨 **객체의 메서드** 호출 | `zzMap.Store("hot", b.Crossed("6"))` | 같음 | **PASS(뚫림)** |
| (f) **지역 포인터로 패키지 var 별칭** | `s := &zzState; s.Hot = b.Crossed("6")` | `rootIdent`가 `s`를 돌려주고 `localNames`가 `s`를 지역으로 분류해 제외한다([:773]) | **PASS(뚫림)** |

셋 모두 `mayReadRecords`의 **수신자 면제**(기록 메서드)를 그대로 유지시킨다. 즉 §6.1·§6.10이
닫았다고 한 N13·N14는 *특정 철자*만 닫혔고 **질문은 여전히 열거로 답해지고 있다.**

`internal/candidate` 안에서 M6·M7은 행동 속성이 잡는다(눈금 의존 읽기이므로). **그러나 T1과
합치면 잡히지 않는다** — 그것이 M4다.

**M4 실측 (리플렉션·unsafe·별칭·임베딩·다른 패키지 전부 불필요):**

```go
// zz_r3_probe.go
var zzHook = func(key string, hot bool) {}
var zzHot  = map[string]bool{}
func init() { zzHook = func(key string, hot bool) {
    if _, seen := zzHot[key]; !seen { zzHot[key] = hot }   // ← 래치
} }

// watch.go shadowBandsFor — 결과가 전부 기록이므로 mayReadRecords가 면제한다
a, b := MeasureSeenLateBand(sighting), MeasureExtendedBand(expansion, summary.Candidate, at, th)
zzHook(summary.Symbol+"|"+expansion.LastPrice, b.Crossed("6"))
return a, b

// watch.go assessOne — 금지 어휘 0개, 기록 셀렉터 0개
if zzHot[summary.Symbol+"|"+v.Expansion.LastPrice] { v.Chase.Extended = RaisedVeto() }
```

```
$ rtk proxy go vet ./...                                 → 출력 없음
$ rtk proxy go test ./internal/candidate/ -run '<네 가드>|TestChangingTheScale'
--- PASS ×6   (record readers: […, shadowBandsFor])
$ rtk proxy go test ./... -count=1                       → ok 51/51, FAIL 0
```

`shadowBandsFor`가 **독자로 인식되고서 허용된다.** 그것이 설계다(결과가 전부 기록) —
그리고 그 함수가 결과가 아닌 통로로 답을 내보낼 수 있다는 것을 `writesOutward`가 못 본다.
대조군 M4b(래치만 제거)는 행동 RED, **구문은 여전히 PASS**다.

**실패 시나리오(구체).** "밴드 관측을 훅으로 뽑아 두자"는 리팩터 한 번과 그 훅 안의 캐시 한 줄.
`assessOne`은 `map[string]bool` 하나를 읽을 뿐이고, `shadowBandsFor`는 스펙이 명시적으로 허용한
모양 그대로다. 리뷰 diff에서 이것을 이상하다고 부를 근거가 지금 저장소에 없다.

### A5. **T3 (P1) — 눈금 변수 자체에서 판정으로 가는 길은 어느 층도 보지 않는다**

기록을 **한 번도 읽지 않는다.** 눈금 값만 들어낸다.

```go
// zz_r3_probe.go — 패키지 레벨 var. init 시점에 프로덕션 눈금에서 "6"을 뜬다.
var zzEdge = ExtendedGainBands[3]

// watch.go assessOne
if over, ok := v.Expansion.GainExceeds(zzEdge); ok && over { v.Chase.Extended = RaisedVeto() }
```

```
$ rtk proxy go vet ./internal/candidate/   → 출력 없음
--- PASS ×6  (네 구문 가드 + 행동 둘)
$ rtk proxy go test ./... -count=1         → ok 51/51, FAIL 0
```

두 층이 못 보는 이유는 각각 정확하다.

- **구문**: 어휘 가드는 `ExtendedGainBands`를 금지 이름에 넣어 두었지만([band_test.go:331]),
  검사 대상을 **결과가 판정 타입인 `*ast.FuncDecl`로 한정**한다([band_test.go:351-355]).
  **패키지 레벨 `var` 선언과 판정을 반환하지 않는 헬퍼는 walk에 아예 들어오지 않는다.**
  변환 가드는 기록 값에 적용된 셀렉터만 보는데 여기엔 기록이 없다.
- **행동**: `onScale`은 init **이후에** 변수를 갈아끼운다([bandbehaviour_test.go:105-111]).
  init에서 뜬 `"6"`은 변형을 따라 움직이지 않는다. → T1과 같은 뿌리다.

**대조군 M5b**: 같은 값을 **live로** 읽으면(`func zzEdgeLive() string { return ExtendedGainBands[3] }`)
행동 RED, **구문은 여전히 PASS**. 즉 구문 층은 눈금 변수 경로를 *어떤 형태로도* 보지 않는다.

이 경로의 성질은 I20 1번과 같은 계열("승인되지 않은 임계를 손으로 박았다")이지만 **문장 그대로
눈금을 참조한다.** 스펙이 요구하는 것은 *"눈금에서 판정으로 가는 경로가 없을 것"*이고
([spec.md:15-16]) 이것은 그 경로다. 그리고 I20의 잔여물 넷 중 어디에도 없다.

### A6. 시도했지만 막힌 것 — 전부 나열한다

- M1·M2(요구된 확인 ①②) — 둘 다 구문·행동 모두 RED.
- U1 `unsafe.Pointer(&v.ExtendedBand)` — 구문 RED(`hands one to Pointer`).
- U2 unsafe 오프셋 산술 → `Crossings` — 행동 RED.
- E1·E2·E3 최상·최하단 밴드 경계 — 전부 행동 RED(A1).
- §6.1 계열 넷(아래 B1) — 전부 구문 RED, 각각 이름으로.
- §6.2 계열 넷(아래 B2) — 읽는 지점에서 전부 구문 RED.
- `Verdict`를 기록으로 만들려는 시도 — `recordTypeClosure`가 **명명 필드를 기록으로 세지
  않으므로**([bandguard_test.go:257-273]) `Verdict`는 기록이 아니고, 그 결과 조립이 계속 허용된다.
  주장은 사실이다. 같은 성질의 다른 타입(`CycleResult.Bands`)도 기록이 아니며, 그것을 읽는
  함수는 `Bands`가 `sh.fields`에 있어 정상적으로 잡힌다.
- `Crossings` 이름 모호성으로 무장 해제하기 — `pinnedRecordFields` 넷이 `Fatalf`로 막는다(B3).

---

## B. 구문 가드의 "새 예외 0개" 주장 — **주장 자체는 사실, 그러나 범위가 좁다**

### B1. §6.1 계열 재현 — 넷 전부 RED, 각각 이름으로

`zz_r3_probe.go`에 넷을 한꺼번에 넣고 `assessOne`에서 부른 결과:

```
bandguard_test.go:1035: zz_r3_probe.go:zzP1            (S1  void + *Verdict)          RED
bandguard_test.go:1035: zz_r3_probe.go:ShadowBand.zzP2 (N13 기록 메서드 + *Verdict)   RED
bandguard_test.go:1035: zz_r3_probe.go:ShadowBand.zzP3 (N14 기록 메서드 + 패키지 var) RED
bandguard_test.go:1035: zz_r3_probe.go:zzP4            (N17 결과가 전부 기록)         RED
bandguard_test.go:1035: watch.go:assessOne             (호출부)                        RED
band_test.go:377:       watch.go:assessOne produces a verdict and reads ExtendedBand   RED
```

§6.1·§6.10이 닫았다고 한 세 면제는 **이 세 철자에 대해서는** 실제로 닫혀 있다. 확인했다.

### B2. §6.2 계열 재현 — 별칭·정의 타입·임베딩·CompositeLit

```
zzP8  (zzHolder{v.ExtendedBand}, unkeyed composite)     RED — hands one to zzHolder{…}
zzP9assess (별칭·정의 타입·임베딩을 거친 기록을 읽는다) RED — reads Measured, Value out of one
```

`zzP5`(별칭 반환)·`zzP7`(임베딩 반환)은 개별로는 RED가 아니다 — **기록을 반환하므로 나르기이고,
사슬은 읽는 지점에서 끝난다.** 파일이 적은 논리(`:923-929`)대로 동작한다. 확인했다.

### B3. `pinnedRecordFields` 넷

넷(`SeenLateBand`·`ExtendedBand`·`Bands`·`Quantiles`)이 `sh.fields`에서 계산으로 나오고,
넷 중 하나라도 빠지면 `Fatalf`다([:959-967]). 기준선에서 넷 다 인식된다(테스트 PASS).
S7의 확대는 타당하다.

### B4. **예외 0건 주장 — 사실이다**

`bandguard_test.go` 전체에 함수명 허용목록이 없다. 프로덕션 독자 일곱
(`BandTally.Collapsed`·`BandTally.CollapsedAlarm`·`MeasureExtendedBand`·`MeasureSeenLateBand`·
`ShadowBand.Crossed`·`ShadowBand.Reason`·`TallyBands`)이 규칙만으로 통과한다. **오탐 0건**도
기준선 PASS가 증거다. `pinnedRecordFields`·`pinnedRecordReaders`는 면제가 아니라 **하한**이다.

### B5. **T5 (P3) — keyed composite 예외의 비대칭 (정보)**

`bandguard_test.go:1004-1009`의 keyed 건너뛰기는 **키 이름이 `sh.fields`에 있으면** 무조건
조립으로 본다. `sh.fields`는 패키지에서 계산되므로 **새로 만든 기록 보유 필드 이름이 자동으로
합류**하고, 그 순간 아무 비-기록 구조체에서나 유효한 "조립 키"가 된다.

실측:

```go
type zzHolder struct{ Held ShadowBand }
func zzKeyed(v Verdict)   zzHolder { return zzHolder{Held: v.ExtendedBand} }  // ← 통과. 독자로도 등록 안 됨
func zzUnkeyed(v Verdict) zzHolder { return zzHolder{v.ExtendedBand} }        // ← RED
```

```
bandguard_test.go:1035: zz_r3_probe.go:zzUnkeyed takes a shadow record apart (hands one to zzHolder{…})
record readers: [… TallyBands zzUnkeyed]        ← zzKeyed는 목록에 없다
```

**단독으로는 우회가 아니다** — `zzHolder.Held`를 읽는 쪽이 잡힌다(B2의 `zzP9assess`가 그 증거).
그러나 `pinnedRecordFields`는 **넷의 무장 해제**만 막고 **새 이름이 예외를 넓히는 것**은 막지
않는다. issues.md I20 4번이 "그 넷 밖의 이름으로 기록을 담는 새 필드는 여전히 무방비"라고 쓴 것과
방향이 다르다 — 새 이름은 무방비인 게 아니라 **예외를 자동으로 획득한다.** 기록만 남길 것을 권한다.

### B6. `writesOutward`가 놓치는 수단 — 실측 결과

브리핑이 지목한 목록에 대해:

| 수단 | 결과 |
|---|---|
| 전역 map(`zzHot[key] = …` 형태의 **대입**) | 잡는다 — `rootIdent`가 패키지 var를 찾는다 |
| 전역 map에 **지역 포인터 별칭**으로 쓰기 | **놓친다 (T2-f)** |
| `sync` 타입(`sync.Map.Store`) | **놓친다 (T2-b)** |
| 함수 값 필드/패키지 레벨 func 값 호출 | **놓친다 (T2-a)** |
| 인터페이스 필드 — 다른 패키지의 이름 | 놓친다 — **이미 I25에 기록됨** |
| `context.Value` | 통로 아님(쓸 수 없다). 해당 없음 |
| 채널 송신(`zzCh <- …`) | 잡는다 — `*ast.SendStmt` case가 있다 |

`isOutwardType`이 **이 패키지가 선언한 이름까지만 닫힌다**는 I25의 서술은 정확하고, 그 대가로
`time.Time`에 오탐이 없다는 설명도 맞다. 그러나 I25는 **인자 축**만 적었고, 위 (a)(b)(f)는
**대입 축**의 누락이라 I25로 덮이지 않는다.

---

## C. Manager가 고친 spec 문장이 이제 정직한가 — **스펙은 옳고, 산출물이 그 SHALL을 어긴다**

### C1. spec의 새 SHALL은 옳다

`specs/candidate-discovery/spec.md:29-37`이 ① 행동 속성이 재는 것은 **판정을 만드는 곳이 내놓는
값**이고 ② 두 층 어느 쪽도 *"모든 경로가 막혔다"*고 주장해서는 **안 되며**(SHALL NOT)
③ 각 층의 덮는/못 덮는 범위를 **측정으로 적어야 한다**(SHALL)고 쓴다. §6.8의 정정은 옳고
문장도 정확하다. 콘솔 반례는 내가 독립적으로 재현했다(V8′, 아래 C3).

### C2. **T4 (P1) — `bandbehaviour_test.go`의 헤더가 그 SHALL NOT을 정면으로 어긴다**

`internal/candidate/bandbehaviour_test.go:41-44`:

> *"Reflection, type aliases, embedding, pointer parameters, **package variables, channels,
> closures**, unsafe **and whatever the next reviewer invents are all covered** by the same
> sentence, because the sentence is about the answer rather than about the route."*

이것이 정확히 *"모든 경로가 막혔다"*는 주장이다. 그리고 **거짓이다**:

- `package variables` — M4의 래치는 패키지 var 둘로 이루어져 있고 통과한다.
- `closures` — M4의 훅은 클로저다. 통과한다.
- `Reflection` — M3이 통과한다.
- `whatever the next reviewer invents` — M3·M4·M5가 그 반례다.

그리고 같은 파일 `:45-52`의 *"What it does NOT cover, **stated here so nobody has to infer it**"*는
**한 가지**(눈금 비의존 필드)만 적는다. **§6.8이 스펙 SHALL로 올린 "다른 패키지" 경계가
이 파일에는 한 줄도 없다.** 세 문서(spec·issues.md·테스트 파일) 중 다음 사람이 가장 먼저 여는
것이 테스트 파일이고, 그것이 가장 부정직하다.

### C3. issues.md I20/I22의 실측 재검증 — **대체로 정직, 두 문장이 과다 주장**

내가 직접 다시 잰 것:

| I20/I22의 문장 | 내 실측 | 판정 |
|---|---|---|
| 리플렉션 → `Value`·`Measured`: 두 층 다 통과 | U3 재현. `go test ./internal/candidate/` 전체 초록 | **정직** |
| 다른 패키지(`internal/console`): 두 층 다 통과 | V8′ 재현 — `signals.go:652`에 그 한 줄, `go vet ./...` 무출력, 네 구문 가드 PASS, 행동 둘 PASS, **51/51 초록** | **정직** |
| `unsafe.Pointer(&v.ExtendedBand)`는 구문이 잡는다 | U1 재현 — `hands one to Pointer` | **정직** |
| unsafe 오프셋 산술은 행동이 잡는다 | U2 재현 — 행동 RED, 구문 PASS | **정직** |
| 테스트 파일 제외 / 모호한 필드 이름 | 코드로 확인(`ParseDir` 필터, `alsoHoldsOther`) | **정직** |
| **I22: "눈금 의존 읽기라면 리플렉션이든 unsafe든 _전부 잡는다_"**([issues.md:511-512]) | **M3이 반례.** 래치가 있으면 `Crossings` 읽기도 안 잡힌다 | **거짓 — 조건이 빠졌다** |
| **I20 표: 리플렉션 → `Crossings` = 행동 ✔** ([issues.md:435]) | 비-래치 읽기에 대해서만 참 | **조건부 — 표에 조건이 없다** |

### C4. §6.9의 정정(*"unsafe가 넷 중 가장 잘 덮이고, 진짜 잔여물은 리플렉션이 `Value`를 읽는 것"*)

**맞다.** U1·U2·U3로 셋 다 직접 확인했다(위 표). 방향 정정은 타당하다.

다만 *"진짜 잔여물"*이 **하나**라는 함의는 이제 틀렸다 — 내 실측으로 잔여물은 최소 넷이다:
I20 1번(리플렉션+눈금 비의존 필드), I20 2번(다른 패키지), **T1(래치)**, **T3(눈금 변수 경로)**.
그리고 뒤의 둘은 리플렉션도 unsafe도 필요로 하지 않는다.

---

## D. 회귀·경계

| 항목 | 명령 | 결과 |
|---|---|---|
| `ExtendedGainPct` 비테스트 대입 | `grep -rnE "ExtendedGainPct[[:space:]]*[:]?=" --include=*.go . \| grep -v _test.go` | **NONE.** 읽기 2건(`veto.go:1049`·`:1069`)·선언 1건·주석 2건뿐 → `extended`는 여전히 veto하지 않는다 |
| 눈금 최종값 | `band.go:73`·`:133` | `SeenLatePercentileBands = {"50","70","80","90","95"}`, `ExtendedGainBands = {"0","2","4","6","8","10","20","30","50","100"}` — **§1 확정 그대로** |
| `internal/risk`·`execgw`·`exitpolicy`·`trading` | `git diff --stat ad16f56 \| grep -E "risk\|execgw\|exitpolicy\|trading"` | **NONE — 무접촉** |
| 콘솔 UI 마찰 | `git diff ad16f56 -- internal/console/ \| grep -inE "confirm\|<form\|<input\|<button\|onclick\|타이핑"` | **없음.** 유일한 "승인" 히트는 경보 **문장 안**의 카피(`"여기서 임계를 승인할 수 없다"`). 템플릿 변경은 `<p class="bad">{{.Alarm}}</p>` 하나로 **표시 전용** |
| §6.5 doc 주석 이동이 다른 함수의 문서를 뺏었나 | `go/ast`로 top-level 선언의 `Doc` 부재를 base와 현재에서 각각 산출해 대조 | **뺏지 않았다.** doc 없는 top-level 선언 집합이 `ad16f56`과 **완전히 동일**(`signalsStateText`·`signalsSourcesText`·`signalsVerdictRank`·`RefreshSeconds` + import/type 넷). `signalsBandAlarm`([:544-556])과 `signalsTallyAlarm`([:565-576]) 각각 자기 doc을 갖는다 |
| FLM | `python3 tools/logic-map/check_analysis.py --change refine-extended-shadow-bands` | `evidence complete or diff-proven exempt`, **exit 0** — §6.5로 재생성된 콘솔 함수 맵 2건의 AST 해시가 현재 코드와 일치 |
| `go test ./...` | `-count=1` | **ok 51 / 51, FAIL 0** |
| `go vet ./...` | | 출력 없음 |
| `make lint` | | `go vet ./...` — 출력 없음 |
| `gofmt -l .` | `$(go env GOROOT)/bin/gofmt` | 출력 없음 |
| `-race` | `./internal/candidate/... ./cmd/tossctl/... ./internal/console/...` | **ok 3/3** (12.6s / 34.0s / 26.3s) |
| PM | `python3 tools/pm/generate_master_tracker.py --check` | `[pm] hierarchy and generated trackers are current` |

### 안전 불변식

| 항목 | 결과 |
|---|---|
| LIVE 주문 side effect | 없음. 주문 명령 미실행 |
| 손절·익절·사이징·Guardian·원장·인증·체결 | 무접촉 |
| 운영 토글 flip | 없음 |
| 임계 결정 | 없음(위 실측) |
| 커밋·push | 없음 |
| 운영 store | 읽지도 쓰지도 않음 |

---

## 복구 대조 — sha256 전건 일치

착수 시점에 14개 파일의 sha256을 기록하고, 모든 변이를 되돌린 뒤 대조했다.
치환은 전부 좁은 hunk 단위였고 전역 치환은 쓰지 않았다.

```
$ rtk proxy sha256sum -c r3-baseline.sha256
internal/candidate/band.go: OK            internal/candidate/watch.go: OK
internal/candidate/band_test.go: OK       internal/candidate/bandguard_test.go: OK
internal/candidate/bandbehaviour_test.go: OK  internal/candidate/bandscale_test.go: OK
internal/console/signals.go: OK           internal/console/templates_signals.go: OK
internal/console/band_alarm_test.go: OK   cmd/tossctl/candidate.go: OK
cmd/tossctl/candidatebands.go: OK         cmd/tossctl/candidate_test.go: OK
cmd/tossctl/candidatebands_test.go: OK    internal/candidate/veto.go: OK
```

`git status --short`도 리뷰 착수 시점과 동일하다. `zz_r3_probe.go`는 존재하지 않는다.

---

## 자체 발견 요약

| # | 등급 | 위치 | 내용 |
|---|---|---|---|
| **T1** | **P1** | `internal/candidate/bandbehaviour_test.go:161-189` | 래치(메모이제이션)가 행동 속성을 무력화한다. 변형 루프가 **프로덕션 눈금을 첫 번째로** 돌리므로, 최초 1회만 계산하는 코드는 프로덕션의 답을 래치하고 이후 다섯 변형은 메모를 읽는다. M3·M4로 실측(둘 다 51/51 초록), 대조군으로 래치가 유일한 차이임을 증명 |
| **T2** | **P1** | `internal/candidate/bandguard_test.go:747-780` | `writesOutward`의 통로 열거가 셋을 놓친다 — 패키지 레벨 func 값 호출, 패키지 레벨 객체의 메서드 호출, 지역 포인터로 패키지 var 별칭. 셋 다 기록 메서드/기록 반환 함수의 면제를 유지시킨다. `mayReadRecords`의 doc이 *"열거는 한 층 아래에서 같은 실수"*라고 쓴 바로 그 실수다 |
| **T3** | **P1** | `internal/candidate/band_test.go:351-355` + `bandbehaviour_test.go:105-111` | 기록을 읽지 않고 **눈금 변수에서 직접** 판정으로 가는 경로를 두 층 다 못 본다. 어휘 가드는 판정 반환 `FuncDecl`만 walk하므로 패키지 레벨 `var zzEdge = ExtendedGainBands[3]`이 밖에 있고, init 고정은 `onScale`보다 앞서므로 행동도 못 본다. M5로 실측(51/51 초록) |
| **T4** | **P1** | `internal/candidate/bandbehaviour_test.go:41-44`, `:45-52`; `issues.md:435`·`:511-512` | 산출물이 spec `:35-37`의 SHALL NOT("모든 경로가 막혔다고 주장하지 말 것")과 SHALL("경계를 측정으로 적을 것")을 어긴다. 테스트 헤더는 *"whatever the next reviewer invents are all covered"*라고 쓰고, "못 덮는 것" 절에는 **§6.8이 스펙으로 올린 다른-패키지 경계조차 없다** |
| **T5** | P3 | `internal/candidate/bandguard_test.go:1004-1009` | keyed composite 예외가 비대칭이다. 새 기록 보유 필드 이름이 `sh.fields`에 자동 합류해 아무 비-기록 구조체에서나 조립 키가 된다. 단독 우회는 아니다(읽는 쪽이 잡힌다). 기록 권고 |

**T1·T2·T3는 각각 독립적인 결함이고, 셋 중 둘만 있어도 완전한 우회가 된다**(M4 = T2+T1,
M5 = T3 단독). 어느 것도 리플렉션·unsafe·다른 패키지를 필요로 하지 않는다.

---

## `make gate` 결과

기록(이 절)까지 다 쓴 **뒤에** 연속 실행했다.

```
$ rtk proxy openspec validate refine-extended-shadow-bands --strict --no-interactive
$ rtk proxy make sdd-sync
$ rtk proxy make sdd-check
$ rtk proxy make gate CHANGE=refine-extended-shadow-bands
```

| 명령 | 결과 |
|---|---|
| `openspec validate refine-extended-shadow-bands --strict --no-interactive` | `Change 'refine-extended-shadow-bands' is valid` |
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain), exit 0 |
| `make sdd-check` | **exit 0** — CodeGraph worktree fingerprint 일치, memory/config/PM/tests/doctor 전부 통과 |
| `make gate CHANGE=refine-extended-shadow-bands` | **FAIL — 2/8, 미완료 태스크 1건(`6.7`)** |

```
bash tools/gate.sh refine-extended-shadow-bands
GATE: refine-extended-shadow-bands
repo: /mnt/D/project/axipient/TossOS

==> 1/8 tasks.md 확인
OK: openspec/changes/refine-extended-shadow-bands/tasks.md

==> 2/8 미완료 태스크 확인
미완료 태스크 1 건:
282:- [ ] 6.7 **3차 독립 리뷰.** 범위는 §6의 diff. 리뷰어는 §6.3의 행동 속성을 **직접 뚫어라** —

GATE FAIL: refine-extended-shadow-bands — 미완료 태스크 1 건이 남아 있습니다
```

나머지 일곱 단계를 손으로 각각 돌렸다.

```
1/8 tasks.md 확인          OK
2/8 미완료 태스크 확인      FAIL — 6.7 하나뿐
3/8 review.md 존재         OK — 이 파일
4/8 Function Logic Map     OK — evidence complete or diff-proven exempt (exit 0)
5/8 make sdd-check         OK — exit 0
6/8 make test              OK — ok 51/51, FAIL 0
7/8 make vet               OK — 출력 없음, exit 0
8/8 make validate          OK — Totals: 25 passed, 0 failed (25 items)
```

> 이 절을 쓰면서 `.md`를 편집했으므로 fingerprint가 다시 stale이다. 직후
> `make sdd-sync` → `make sdd-check`를 한 번 더 연속 실행했다(둘 다 exit 0).
> 위 게이트 결과는 그 사이에 **코드가 한 줄도 바뀌지 않았으므로** 그대로 유효하다.

**§6.7만 미완료다.** 체크박스는 **열어 둔다** — 판정이 랜딩 불가이므로 §4.4·§5.5의 선례
(*"수행되고 기록됐다"*)를 적용할 자리가 아니고, 브리핑이 요구한 상태(*"§6.7만 미완료로 남아야
한다"*)와도 일치한다. 닫는 방법은 T1~T4를 처리하고 4차 리뷰를 받는 것이다.

---

## 최종 판정

### **랜딩 불가.**

**사유 (T1·T2·T3·T4).**

1. **네 번째 계열이 있다.** M4는 리플렉션도 unsafe도 별칭도 임베딩도 다른 패키지도 쓰지 않는다.
   스펙이 명시적으로 허용하는 모양(`shadowBandsFor`: 결과가 전부 기록)에서 밴드를 읽고,
   패키지 레벨 func 값 하나로 답을 내보내고, 래치 한 줄로 행동 속성을 침묵시킨다.
   `go vet` 무출력, 네 구문 가드 PASS, 행동 둘 PASS, **51/51 초록.**
   세 번의 돌파와 **형태가 같다** — 나타나는 방식이 "검사가 통과하는 모습"이라는 점까지.
2. **M5는 그보다 더 싸다.** 기록을 한 번도 읽지 않고 눈금 값만 들어낸다. 두 줄이다.
3. **그리고 문서가 그 셋이 불가능하다고 쓴다.** `bandbehaviour_test.go:41-44`의
   *"whatever the next reviewer invents are all covered"*는 §6.8이 스펙 SHALL NOT으로 올린 바로
   그 주장이다. 이대로 아카이브하면 **승인된 SHALL이 이 change의 주 산출물에 대해 거짓인 채로
   남는다** — D10이 경고한 오염이고, 이 두 change에서 반복된 형태다.

**닫는 데 필요한 최소치(제안 — 구현자가 형태를 고르고 이유를 적는다).**

- **T4는 편집만으로 닫힌다.** `bandbehaviour_test.go`의 "all covered" 문장을 지우고,
  "What it does NOT cover" 절에 **측정된 경계 넷**(눈금 비의존 필드 · 다른 패키지 · **래치** ·
  **눈금 변수 직접 참조**)을 적는다. issues.md I20 표의 리플렉션·unsafe 행에 *"래치가 없을 때"*
  조건을 붙이고 I22의 "전부 잡는다"를 고친다. **이것만은 반드시 해야 한다** — 스펙의 SHALL이다.
- **T1**은 속성 자체로는 닫기 어렵다(상태를 가진 코드는 원리적으로 순수성 전제를 깬다).
  값싼 완화 둘: ① 변형 순서를 뒤집거나 프로덕션 눈금을 **마지막에도 한 번 더** 돌려
  *"첫 번째와 마지막의 프로덕션 판정이 같은가"*를 함께 단언한다(래치가 프로덕션 답을
  잡는 것은 못 막지만, 프로덕션→변형→프로덕션에서 래치된 답이 드러난다), ② 각 변형 전에
  패키지 상태가 없음을 보증할 수 없다는 사실을 **경계로 기록**한다. 최소한 ②는 해야 한다.
- **T2**는 `writesOutward`를 "통로 열거"에서 벗어나게 하거나, 벗어날 수 없다면 **놓치는 셋을
  측정으로 기록**한다. (a)(b)는 "패키지 레벨 이름을 뿌리로 하는 **호출**"을 세는 것으로,
  (f)는 `&<패키지 var>`를 취하는 지역 이름을 추적하는 것으로 각각 몇 줄이면 닫힌다 —
  다만 그것도 열거이므로, 고른 이유를 적는 편이 낫다.
- **T3**은 어휘 가드의 대상 선택을 넓히거나(판정 반환 함수 → 패키지 레벨 `var` 초기화식과
  임의의 헬퍼), 닫지 않기로 하면 **왜 닫지 않는지를 측정과 함께** 적는다.

**랜딩 불가로 판단한 근거를 한 문장으로**: 이 가드는 세 번 뚫렸고, 나는 **네 번째를 찾았으며**,
그 네 번째가 불가능하다고 쓴 문장이 이 change가 아카이브할 산출물 안에 있다.

---

# Manager 처분 — 3차 리뷰(§6.7)에 대한 답 (2026-07-29)

리뷰의 판정은 **랜딩 불가**였다. 그 판정을 지우지 않는다. 아래는 각 항목을 **직접 재현한**
결과와, 그 결과로 내린 처분이다. 결론부터: **네 번째 돌파(M5)는 실재하고, 이 계열의 검사로는
닫을 수 없다.** 그래서 검사를 더 강화하지 않고, **보장을 실제 수준으로 낮춰 적고 랜딩한다.**

## 재현 결과

| 리뷰 항목 | Manager 재현 | 처분 |
|---|---|---|
| **M5 — 눈금을 직접 인덱싱해 래치** | 재현됨. 빌드 성공, 구문 가드 4종 PASS, 행동 속성 2종 PASS, `go vet` 무출력 | **수용.** I26에 전문 |
| **T1 — 프로덕션 눈금을 먼저 돌려 래치가 동어반복이 된다** | 결함은 실재. **제안된 해법(순서 변경)은 실측 결과 잡지 못한다** | **수용, 해법 기각.** 7.2 |
| **T4 — 헤더가 spec의 SHALL NOT을 어긴다** | 확인. `bandbehaviour_test.go:41-44` | **수용, 고친다.** 7.7 |
| **T3 — 패키지 var가 자유롭다** | **기록 읽기에 대해서는 거짓.** 변환 가드가 `watch.go:assessOne`을 이름으로 지목하며 실패하고 행동 속성도 실패한다 | **부분 기각.** I27 |
| **T2 — `writesOutward`가 통로를 놓친다** | I25가 이미 담고 있다. 닫으려면 `go/types` | **기록 유지, 닫지 않음** |
| **C3 — I20 표와 I22에 조건이 빠졌다** | 맞다 | **수용.** I22에 *"판정 시점에"* 한정어, I20 표에 조건 주석 |

## 왜 여기서 멈추는가

세 라운드에서 나온 것은 전부 실재하는 결함이었고 매번 다른 계열이었다. 네 번째가 다른 것은
**기록을 한 번도 읽지 않는다**는 점이다. 앞의 셋은 "기록을 읽는 것을 더 잘 막자"로 닫혔지만,
`ExtendedGainBands`는 exported 패키지 변수이고 값을 미리 꺼내 들면 눈금 교체에 반응하지
않는다. 접근자(`BandsFor`)가 슬라이스 자체를 돌려주므로 변수를 감추는 것도 소용없다.

**그러므로 이 계열의 검사는 완전해질 수 없고, 다섯 번째 모양을 찾아도 결론은 같다.**
모양을 하나씩 지우면서 클래스를 그대로 두는 것은 이 change가 막으려는 실패 그 자체다.

랜딩을 막는 것은 M5가 아니라 **M5를 적지 않고 랜딩하는 것**이다. 실제 위험도는 낮다 —
`ExtendedGainPct`에 대한 비테스트 대입은 0건이고 `extended`는 오늘 어떤 후보도 거부하지
않는다. 이 경로는 누가 새로 쓸 때 열리는 것이지 지금 열려 있는 것이 아니다.

## 클래스를 실제로 닫는 두 후속 (둘 다 Normal)

1. **눈금을 패키지 상태가 아니라 호출자가 건네는 인자로 만든다**(I26). 초기화 시점에 잡을
   대상이 없어지므로 래치 계열 전체가 사라진다. 임계가 승인되면 `extended`는 그림자가
   아니게 되므로(I3) 그 전에 배관을 바꾸면 두 번 바꾸게 된다 — 순서상 지금이 아니다.
2. **소비 표면에서 같은 속성을 잰다**(I20 2번). 허용목록 10개 파일에 방어가 전무하고,
   그중 하나가 임계를 승인할 사람이 읽는 화면이다.

## 이 처분으로 바뀐 산출물

- spec delta: 도달불가 SHALL을 **두 범위로 한정**(같은 패키지 안에서 ① 기록→판정,
  ② 판정 시점의 눈금 읽기). 그 밖은 보장되지 않으며 측정으로 적어야 한다(SHALL).
- issues.md: **I26**(래치 전문·반례·유일한 비열거적 해법), **I27**(T3 재현),
  I20 표에 두 행과 조건 주석, I20 목록에 5번, I22에 한정어 정정.
- tasks.md: **§7** 신설. 7.9가 4차 리뷰를 하지 않는 이유를 적는다.
- `bandbehaviour_test.go`: 헤더의 과다 주장 제거와 경계 넷 기재(7.7).
