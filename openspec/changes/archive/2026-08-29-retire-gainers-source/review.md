# review — retire-gainers-source

## 구현 기록 (Teammate)

- 날짜: 2026-07-28
- 역할: Teammate(구현). **독립 리뷰는 이 기록에 포함되지 않는다** — §7.4는 미완료다.
- base-commit: `b2c261a6d84c339cd370b8267b8b9f9c6293948b`
- 브랜치: `feat/p0-foundation` (작업 트리에 미커밋 상태로 둠 — Manager 검토용)

### 무엇을 바꿨는가 (파일별)

수정 (`git diff --stat`, PM 등록 2건은 Manager가 이미 넣어 둔 것):

```
cmd/tossctl/candidate.go                   | 15 +++++++++-
internal/candidate/candidate.go            |  9 ++++++
internal/candidate/notdue_test.go          |  8 ++---
internal/candidatesrc/candidatesrc.go      | 44 +++++++++++++++++++++++----
internal/candidatesrc/clock_wiring_test.go | 48 +++++++++++++++++++++++-------
```

| 파일 | 무엇 |
|---|---|
| `internal/candidatesrc/candidatesrc.go` | §1.1 `Panel`의 랭킹 타입 리터럴에서 `RankingTopGainers` 제거 + 왜 내렸는지·재검토 조건 4개를 주석으로 기록. §1.4 const 블록 주석에 "선언되어 있지만 어느 패널에도 없다" 한 줄 |
| `cmd/tossctl/candidate.go` | §1.2 `candidateIntervals()`에서 `SourceOfficialGainers` 항목 제거 + 왜 두 곳을 같이 빼야 하는지와 의무의 방향을 주석에 기록 |
| `internal/candidate/candidate.go` | §1.4 `SourceOfficialGainers` 상수 주석에 패널에 없다는 사실과 id를 남기는 이유 |
| `internal/candidatesrc/clock_wiring_test.go` | §2.1 `len(sources) != 4` → id 집합 단언. §2.2 헤더 주석과 실패 메시지에서 개수 제거 |
| `internal/candidate/notdue_test.go` | §6.1 US 패널을 "exactly three official sources"라고 말하는 주석 2곳을 개수 없는 문장으로 |

신규:

| 파일 | 무엇 |
|---|---|
| `internal/candidatesrc/snapshot_drift_test.go` | §4 스냅샷 표류 테스트 (핵심 산출물) |
| `internal/candidatesrc/retiredsource_test.go` | §3.1 패널에 gainers 없음 + D2가 남기기로 한 상수·매핑이 남아 있음. `sameSourceSet`/`sortedIDs` 헬퍼 |
| `cmd/tossctl/candidateschedule_drift_test.go` | §3.2 일정 키 집합 ⊆ 패널 id 집합, 방향은 한쪽뿐 |
| `internal/candidate/retiredsupporter_cooling_test.go` | §5.2·§5.3 냉각 안전 근거의 두 반쪽 |
| `openspec/changes/retire-gainers-source/issues.md` | §6.3·6.4·6.5 + 판단 하나 추가(I4) |
| `.../analysis/function-logic/` 3건 | §7.2 |

**지우지 않은 것 (§1.3 확인)**: `candidate.SourceOfficialGainers`, `candidatesrc.RankingTopGainers`,
`rankingSourceID[RankingTopGainers]`, `candidate.go`의 "정의상 이미 일어난 움직임의 목록"
논거 — 전부 그대로다. `TestTheRetiredSourceKeepsItsIdentity`가 이제 그 셋을 테스트로 고정한다.

---

## 각 `[T]` task의 변이 확인 결과

RED/GREEN은 **실제로 실행한 결과**다. 변이는 넣고, 돌리고, 되돌렸다.

### §2.3 — 기존 테스트가 실제로 깨지는지 (1.1 적용 후, 2.1 적용 전)

```
=== RUN   TestThePanelHandsItsClockToEverySourceItBuilds
    clock_wiring_test.go:65: the KR panel has 3 sources, want 4 (three rankings and the popularity list)
--- FAIL: TestThePanelHandsItsClockToEverySourceItBuilds
```

같은 실행에서 나머지 49건은 통과했다. 이 단계에서 신규 §3.1·§4 테스트는 이미 GREEN이었다.

### §3.1 — 패널에 gainers가 없다

- **자연 RED** (1.1 적용 전): KR·US 두 subtest 모두 실패.
  `retiredsource_test.go:47: the KR panel builds official_rankings_top_gainers. ...`
- **GREEN** (1.1 적용 후).
- **변이 ① `Panel`에 `RankingTopGainers` 복원 → RED** ✔ (KR·US 둘 다). 되돌린 뒤 GREEN.

### §3.2 — 일정에만 남은 원천이 없다

- **1.1 적용 전 GREEN** (양쪽에 다 있으므로 위반이 아님).
- **변이 ① 일정에만 gainers를 남긴다 → RED** ✔. 두 번 관측했다 — 1.1만 적용하고 1.2 적용
  전의 자연 상태에서 한 번, 1.2 이후 항목만 되살린 변이에서 한 번. 메시지:
  `candidateIntervals sets a cadence for [official_rankings_top_gainers] and no market's panel builds those sources.`
- **변이 ② 일정에서만 빼고 패널에 둔다 → GREEN** ✔. 변이 ①(패널 복원) 상태에서
  `cmd/tossctl` 테스트가 통과하는 것으로 확인했다 — 위반이 아니라는 것이 설계다.
- **변이 ③ 둘 다 뺀다 → GREEN** ✔ (현재 상태).
- 추가로 규칙 자체를 fixture로 양방향 검사하는 `TestTheScheduleGuardIsContainmentAndNotEquality`를
  넣었다. `t.Errorf` 루프 안에만 있는 규칙은 그 자신이 잡히지 않게 되는 것을 검사할 수 없다.

### §3.3 — `TestEveryPanelSourceHasItsOwnID`

수정하지 않았고 계속 통과한다 (`go test ./internal/candidatesrc/` 53건 전부 통과).

### §3.4 — `panelsize_drift_test.go`

수정하지 않았고 계속 통과한다. `declaredPanelSizes`가 읽는 값 집합은 `{100, 30}` 그대로다 —
`OfficialRanking(..., 100, clk)` 호출이 리터럴 순회 안에 있어서 **호출 횟수가 줄어도 AST에
있는 리터럴은 하나**이기 때문이다. issues.md에 기록할 새 결합은 발견되지 않았다.

### §4.4 — 스냅샷 표류 테스트의 변이 세 가지

- **자연 RED** (어떤 배선 변경보다 먼저 작성):

```
snapshot_drift_test.go:322: Panel builds TOP_GAINERS and the reader asks for duration "realtime",
which ../../docs/migration/openapi.latest.json says that type does not accept
(400 unsupported-ranking-duration).
...
Forbidden with "realtime": [TOP_GAINERS TOP_LOSERS]
```

- **변이 ① `Panel`에 `RankingTopGainers` 복원 → RED** ✔ (위와 같은 메시지).
- **변이 ② 스냅샷의 `realtime 미지원` 문장 삭제 → RED** ✔ (빈 금지 집합에서 Fatal):

```
snapshot_drift_test.go:305: ../../docs/migration/openapi.latest.json yielded no forbidden
type/duration combination at all. ... an empty set means the matcher stopped matching rather
than that the API stopped refusing.
```

  스냅샷은 백업본에서 되돌렸고 `git diff docs/migration/openapi.latest.json`이 비어 있음을
  확인했다.

- **변이 ③ duration을 `1d`로 → GREEN** ✔. **두 가지 형태로 확인했다**:
  ⓐ 현재 배선(gainers 없음) + `1d` → GREEN.
  ⓑ **gainers를 패널에 되돌린 상태 + `1d` → GREEN**. ⓑ가 진짜 증명이다 — 금지된 것이
  `TOP_GAINERS`가 아니라 `(TOP_GAINERS, realtime)`이라는 **조합**임을 보인다. D3가
  요구사항 수준에서 말한 것을 테스트가 지킨다.

**금지 집합을 스냅샷의 어디에서 읽었는지 (§4.1)**: `paths./api/v1/rankings.get.parameters`의
`name == "type"` 파라미터 `description`. 같은 사실이 엔드포인트 description과 오류 예시에도
있지만 셋 중 이것만 **타입당 한 줄**이다. 엔드포인트 description은 `` `TOP_GAINERS` /
`TOP_LOSERS` 는 ``처럼 두 타입을 슬래시로 이어 한 문장에 넣으므로 집합을 만들려면 그 결합을
파싱해야 하고, 세 번째 타입이 다른 형태로 붙으면 놓친다. 오류 예시의 기계 판독 부분
(`allowedValues`)은 **허용되는 duration**을 나열할 뿐 어느 타입에 대한 것인지 말하지 않는다 —
필요한 절반이 없다. 근거는 테스트 파일 주석에 표로 적어 두었다.

**어느 패키지에 두었는지 (§4.5)**: `internal/candidatesrc`. 감시 대상 배선이 이 패키지 자신의
파일이므로 `Panel` 편집과 이 테스트의 실패가 같은 `go test ./internal/candidatesrc/`에서
일어나고, 패키지 밖 상대경로가 둘이 아니라 하나면 된다. 파일을 파싱하는 것은 import가
아니므로 `isolation_test.go`의 금지표와 무관하다(선례: `panelsize_drift_test.go`,
`fsguard_drift_test.go`).

추가로 **matcher의 positive control**을 넣었다
(`TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor`). 빈 집합 Fatal만으로는
"두 줄 중 한 줄만 읽는 matcher"를 잡지 못한다.

### §5.2 — 살아 있는 supporter가 있으면 냉각은 정상 동작한다

GREEN.
**변이: `coverageAnswered`에서 `if !heard[id] { continue }` 제거 → RED** ✔

```
retiredsupporter_cooling_test.go:72: cooled = 0, want 1. A supporter that is no longer in the
panel and was not passed over by the schedule is a source that is gone ...
```

되돌린 뒤 `git diff internal/candidate/scan.go`가 비어 있음을 확인했고 패키지 전체 GREEN.

### §5.3 — supporter가 빠진 id 하나뿐인 후보는 냉각되지 않는다

GREEN (현재 동작을 그대로 고정). 위 변이에서는 이 테스트가 **PASS로 남았다** — 두 테스트가
서로 다른 것을 재고 있다는 증거다.

---

## §5.1 측정 결과 — gainers를 sources에 가진 후보 건수

읽기 전용. 운영 store(`/home/daniel/.local/share/tossos/candidates.db`)와 그 `-wal`/`-shm`을
scratchpad로 **복사한 뒤** 복사본을 SQLite로 조회했다(원본 무변경).

```
candidates 총계: 300
sources LIKE '%official_rankings_top_gainers%' 인 후보: 0
observations WHERE source='official_rankings_top_gainers': 0

candidates.sources 분포:
  official_rankings_trading_value,official_rankings_trading_volume   120
  official_rankings_trading_volume                                    75
  official_rankings_trading_value                                     75
  wts_popular                                                         25
  official_rankings_trading_value,...,wts_popular                      5

observations 원천별:
  official_rankings_trading_volume  200
  official_rankings_trading_value   200
  wts_popular                        30
```

**판정: 0건.** proposal의 안전 논증이 요구한 조건이 성립한다 — 이 원천을 유일한 supporter로
가진 후보가 없으므로 §5.3이 고정한 "영영 냉각되지 않는 후보"는 이 저장소에 존재하지 않는다.
관측이 0건이라는 사실은 "한 번도 응답한 적이 없다"를 store 쪽에서 독립적으로 확인해 준다.

---

## §7.3 장중 실측

KR 장은 닫혀 있었다(실행 시각 2026-07-28 23:53 KST). **US 정규장 중**이었으므로
(10:53 EDT, 09:30–16:00) `--market US`로 실행했다. 읽기 전용(`mutating: false`).

```
$ ./bin/tossctl candidate scan --market US

candidate scan — US at 2026-07-28T14:53:19Z
  read-only: no order is placed, amended or cancelled by this command

sources        2 attempted, 2 responded
  reading      official_rankings_trading_value — 100 requested, 100 arrived (whole)
  reading      official_rankings_trading_volume — 100 requested, 100 arrived (whole)

recorded       200 observations, 169 candidates, 0 cooled
  firsts       169 first prices, 0 first ranks newly stored
  held         169 position(s) not stored — the reading that carried them had no previous reading to compare against, and a first rank is written once. A single `scan` can never qualify one; `watch` qualifies from its second turn

chase veto     169 candidate(s) assessed
  unmeasured   169  ← the default reading is "mostly unchecked", not "mostly safe"
  vetoed       0
  passed       0  (structurally 0: seen_late and extended have no approved threshold (design.md D18), so no candidate can have all three vetoes measured and clear. An absent threshold is not a pass)
  seen_late    raised 0, unmeasured 169
  extended     raised 0, unmeasured 169
  near_high    raised 0, unmeasured 169
  reason       THRESHOLD_ABSENT 338
  reason       NO_DAY_HIGH 169

first sightings by source — which readings seen_late is refusing
  official_rankings_trading_value        0 of 100 measured
    refused    NO_FIRST_RANK 100
  official_rankings_trading_volume       0 of 69 measured
    refused    NO_FIRST_RANK 69

shadow crossings — acceleration (records and decides nothing; this is what a threshold will be derived from)
  measured     0 of 200 series
  crossed      1.3 0 · 1.5 0 · 1.8 0 · 2.0 0 · 2.5 0
  not computed WARMING_UP 200

shadow bands — seen_late (records and decides nothing; this is what a threshold will be derived from; not a veto)
  measured     0 of 169
  crossed      50 0 · 70 0 · 80 0 · 90 0 · 95 0
  not measured NO_FIRST_RANK 169

shadow bands — extended (records and decides nothing; this is what a threshold will be derived from; not a veto)
  measured     169 of 169
  crossed      10 0 · 20 0 · 30 0 · 50 0 · 100 0

retention      0 raw row(s) and 0 expired summary/summaries pruned past 48h0m0s; write-ahead log reclaimed
free space     63688163328 bytes on /home/daniel/.local/share/tossos/candidates.db, floor 536870912
```

**판정 (D6의 두 기준, 숫자가 아니다)**

1. `attempted`와 `responded`가 같다 — **2 = 2. 통과.**
2. 그 줄에 `(degraded)`가 없다 — **없다. 통과.**

US 패널이 둘인 것은 `candidatesrc.Panel`이 WTS를 KR에만 주기 때문이고, 이 change 이후
US에 남는 공식 랭킹 둘과 일치한다.

**미측정으로 남기는 것**: "변경 전 같은 명령이 `(degraded)`를 냈다"는 대조 실행은 하지
않았다. 그러려면 이전 배선의 바이너리로 live 호출을 한 번 더 써야 하고, 그 사실은
proposal이 이미 400 응답 전문으로 기록하고 있다. D6의 판정 기준은 사후 상태에 대한
것이므로 완료 조건에는 영향이 없다.

---

## 검증 명령과 실제 결과

| 명령 | 결과 |
|---|---|
| `go test ./...` | **3685 passed in 57 packages** (실패 0) |
| `go vet ./...` | **No issues found** |
| `gofmt -l .` | **출력 없음** (`$(go env GOROOT)/bin/gofmt` — PATH에 `gofmt`가 없다) |
| `go test -race ./internal/candidate/... ./internal/candidatesrc/... ./cmd/tossctl/...` | **662 passed in 3 packages** |
| `python3 tools/pm/generate_master_tracker.py --check` | `hierarchy and generated trackers are current` |

**upstream 상속 테스트 회귀 없음.** 전체 건수는 proposal이 인용한 3676에서 3685로 늘었다 —
이 change가 더한 테스트만큼이고 줄어든 것은 없다.

**§7.5 PM 등록 확인**: `docs/pm/portfolio/_registry.yaml:29`와
`tools/pm/test_generate_master_tracker.py:35` 양쪽에 `"retire-gainers-source"`가 있다.

---

## Function Logic Map

**면제를 주장하지 않았다.** `python3 tools/logic-map/check_analysis.py --change
retire-gainers-source`를 먼저 돌렸고, 출력이 대상 함수를 지목했다.

첫 실행은 **6건**을 잡았다:
`cmd/tossctl/candidate.go:candidateIntervals`, `internal/candidatesrc/candidatesrc.go:Panel`,
그리고 `clock_wiring_test.go`의 `aFullRanking`·`TestThePanelHandsItsClockToEverySourceItBuilds`·
`sameSourceSet`·`sortedIDs`.

design D7은 그 파일에서 대상이 **하나**일 것으로 예측했다. 어긋난 이유는 내가 새 헬퍼 둘을
기존 파일에 넣었기 때문이고, 그것이 인접 함수까지 hunk에 끌어들였다. 헬퍼를 새 파일
(`retiredsource_test.go`)로 옮기자 대상이 **3건**으로 줄어 design의 예측과 일치했다. 이것은
게이트 회피가 아니라 tasks.md가 이미 요구한 배치다 — "새 테스트는 새 파일에 쓴다"이고,
§2의 예외는 **기존 테스트가 깨지기 때문**이지 새 코드를 그 파일에 넣기 위해서가 아니다.

최종 대상과 산출물:

| 함수 | 산출물 |
|---|---|
| `internal/candidatesrc/candidatesrc.go:Panel` | `analysis/function-logic/internal-candidatesrc--panel/` |
| `cmd/tossctl/candidate.go:candidateIntervals` | `analysis/function-logic/cmd-tossctl--candidateintervals/` |
| `internal/candidatesrc/clock_wiring_test.go:TestThePanelHandsItsClockToEverySourceItBuilds` | `analysis/function-logic/internal-candidatesrc--testthepanelhandsitsclocktoeverysourceitbuilds/` |

각 디렉터리에 `ast.json`(현재 리비전, 해시 일치), `function-logic-map.md`,
`branch-test-map.md`, `risk-pattern-report.md`. 최종 확인:

```
[logic-map] retire-gainers-source: evidence complete or diff-proven exempt
```

세 함수 모두 **분기 구조가 base와 동일하다**. `Panel`은 B1~B4 그대로이고 B2가 순회하는 값
집합만 줄었다. `candidateIntervals`는 분기가 아예 없다. 테스트 함수는 진입 단언(B2·B3)만
바뀌었고 `t.Run` 루프(B4~B10)는 그대로다.

---

## §6.2 — 갱신할 문서가 없다

`docs/baseline.md`에는 발굴 원천을 세는 문장이 없다(`candidate`·`발굴`·`discovery`·`원천`
어느 것도 매치하지 않는다). `docs/ROADMAP.md`도 원천 **개수**를 말하지 않는다.
`grep -rn "top_gainers\|TOP_GAINERS\|SourceOfficialGainers"`로 훑은 결과 `.md` 쪽 매치는
전부 아카이브되지 않은 이전 change의 기록(`add-candidate-discovery`,
`console-*`, `fix-chase-veto-measurement`의 FLM 산출물)이고, 그것들은 그 시점의 기록이므로
고치지 않았다. **갱신할 곳이 없다.**

---

## 막힌 것 · 틀렸다고 생각하는 것

### 1. §7.4 독립 리뷰는 하지 않았다 — 그리고 그 때문에 `make gate`가 실패한다

구현자가 자기 구현을 독립 검증할 수 없다. tasks.md의 7.4 체크박스를 비워 두었고,
`tools/gate.sh`의 2단계(미완료 체크박스 0건)가 거기서 멈춘다. **이것은 정상적인 결과이고
숨기지 않는다** — §7.4가 요구하는 것은 "구현과 분리된 컨텍스트"이며, 특히 §4의 변이 세
가지를 **리뷰어가 직접 넣어** 확인하라고 적혀 있다.

### 2. tasks.md §3.4의 근거가 코드와 다르다 (결론은 맞다)

> "AST에서 `OfficialRanking(..., 100)`·`WTSPopular(..., 30)`의 **행 수**를 읽으므로 호출이
> 세 번에서 두 번으로 줄어도 값 집합은 `{100, 30}` 그대로다"

`OfficialRanking`은 소스에서 **한 번** 호출된다 — 리터럴 순회 **안에서**다
(`candidatesrc.go`의 `for _, typ := range ...` 블록). `declaredPanelSizes`는 AST를 읽으므로
런타임 호출 횟수가 아니라 **소스의 호출식 개수**를 보고, 그것은 이 change 전에도 후에도
1이다. 즉 "세 번에서 두 번으로 줄어도"라는 문장이 가리키는 세 번은 소스에 존재한 적이 없다.

결론(`{100, 30}` 유지)은 맞고 테스트도 통과한다. 근거만 틀렸다. 같은 종류의 오류가
proposal에서 `NoteSources`에 대해 한 번 있었으므로 적어 둔다.

### 3. §6.1의 grep이 한 곳을 더 잡았고, 고치지 않기로 판단했다

`internal/candidatesrc/candidatesrc_test.go:290`의 *"The three official rankings shipped
under one source id"*. 이것은 §2 리뷰가 고친 **과거 결함의 서술**이고, `rankingSourceID`가
D2에 따라 세 항목을 유지하므로 상수 쪽에서는 여전히 사실이다. 패널 구성을 말하는 문장이
아니라고 판단했다. 판단 근거와 함께 issues.md I4에 적었다 — 틀렸다면 고칠 곳은 그 한 줄이다.

### 4. §2.2가 지적한 표류의 실제 범위가 더 넓었다

tasks.md는 `clock_wiring_test.go`의 두 곳(헤더, 실패 메시지)과 `notdue_test.go`의 두 곳,
`candidatesrc.go`의 한 곳을 예상했다. 전부 실재했고 전부 고쳤다. 그런데 **그 중 어떤 것도
테스트나 컴파일러가 잡지 않았다** — 잡힌 것은 `want 4` 하나뿐이고 나머지 넷은 조용히
거짓이 됐다. issues.md I2에 그 사실과, 가드를 만든다면 `rowClaim`과 다른 형태여야 하는
이유를 적었다.

### 5. `docs/migration/openapi.latest.json`을 변이로 한 번 수정했다가 되돌렸다

§4.4 변이 ②를 실제로 수행하기 위해서다. 백업본에서 복원했고
`git diff docs/migration/openapi.latest.json`이 비어 있음을 확인했다. 현재 작업 트리에
이 파일의 변경은 없다.

---

## §7.6 — `make sdd-sync` → `make sdd-check` → `make gate`

`.go` 편집을 전부 끝낸 뒤 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph / CodeGraphContext / GBrain) |
| `make sdd-check` | 통과 — `[index-freshness] CodeGraph hard-evidence index matches the worktree`, `[agent-config] ... synchronized`, memory `valid`, `[pm] ... current`, sdd-test 전부 통과 |
| `make gate CHANGE=retire-gainers-source` | **FAIL — 2/8단계에서 미완료 태스크 1건(`7.4 독립 리뷰`)** |

**게이트 실패는 사실 그대로 기록한다.** 실패 지점은 2단계 하나이고 원인은 §7.4가 열려
있다는 것뿐이다. 나머지 단계를 각각 손으로 돌려 확인했다:

```
1/8 tasks.md 확인                     OK
2/8 미완료 태스크 확인                 FAIL — 7.4 독립 리뷰 (구현자가 할 수 없다)
3/8 review.md 존재                    OK (이 파일)
4/8 Function Logic Map                OK — evidence complete or diff-proven exempt
5/8 make sdd-check                    OK
6/8 make test                         OK — 실패 0
7/8 make vet                          OK
8/8 make validate                     OK — 25 passed, 0 failed
```

`openspec validate retire-gainers-source --strict --no-interactive` → `Change
'retire-gainers-source' is valid`.

7.6 체크박스는 **이 세 명령을 실제로 돌렸기 때문에** 채웠다. 게이트가 통과했다는 뜻이
아니다 — 통과하지 못했고 그 이유는 위에 적힌 하나다.

---

## 남은 위험

| 위험 | 상태 |
|---|---|
| 원천이 줄어 후보를 놓친다 | 측정됨 — 이 원천이 올린 후보 0건, 관측 0건 |
| 냉각 불가능한 후보가 생긴다 | 측정됨 — 그런 후보 0건. §5.3이 만약의 동작을 고정 |
| 일정과 패널이 어긋난다 | 테스트가 잡는다(`candidateschedule_drift_test.go`), 변이로 확인 |
| 같은 결함이 다른 원천에서 재발 | `snapshot_drift_test.go`가 **조합**을 금지, 변이 ③으로 확인 |
| 되찾을 수 없는 4xx가 후퇴 사다리에 안 걸린다 | 범위 밖. issues.md I1에 후속 change 등급(High-risk)까지 기록 |
| 원천 개수 주장 주석에 가드가 없다 | 범위 밖. issues.md I2 |
| 독립 리뷰 미완료 | **§7.4 열려 있음.** `make gate` 미통과의 원인 |

---

# 독립 리뷰 (§7.4)

- 날짜: 2026-07-29
- 역할: **독립 리뷰어.** 구현과 분리된 컨텍스트다. 이 절의 RED/GREEN은 전부 리뷰어가
  **직접 변이를 넣고, 돌리고, 되돌린** 결과이며 구현자의 보고를 인용한 것이 아니다.
- base-commit: `b2c261a6d84c339cd370b8267b8b9f9c6293948b` (변경 없음)
- 시작 상태: `git diff --stat` 7개 파일 108 insertions / 22 deletions,
  신규 4개 파일. 리뷰 종료 시 동일함을 확인했다.
- 변이 정책: 모든 변이는 확인 즉시 되돌렸고, 되돌린 뒤 `git diff --numstat`으로 원상복구를
  확인했다. `docs/migration/openapi.latest.json`은 sha256으로 대조했다.

## A. 스냅샷 표류 테스트(§4.4) — 직접 확인

### A.1 `Panel`에 `RankingTopGainers` 복원 → RED ✔

```
$ go test ./internal/candidatesrc/ -run TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
[FAIL] TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
   snapshot_drift_test.go:322: Panel builds TOP_GAINERS and the reader asks for
   duration "realtime", which ../../docs/migration/openapi.latest.json says that type
   does not accept (400 unsupported-ranking-duration).
```

같은 변이에서 `retiredsource_test.go:75`(KR·US 둘 다)와 `clock_wiring_test.go:90`도 RED.
되돌린 뒤 GREEN.

### A.2 스냅샷의 `realtime 미지원` 두 불릿 삭제 → **Fatal** ✔ (조용히 통과하지 않는다)

```
$ go test ./internal/candidatesrc/ -run 'TestThePanelAsksForNothing…|TestTheSnapshotMatcher…'
--- FAIL: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
    snapshot_drift_test.go:305: ../../docs/migration/openapi.latest.json yielded no
    forbidden type/duration combination at all. … an empty set means the matcher stopped
    matching rather than that the API stopped refusing …
--- PASS: TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor
```

`t.Fatalf`이고 line 305이므로 이후 단언에 도달하지 않는다. **장식이 아니다.**

**원상복구 확인**: 백업(scratchpad)에서 복원 후

```
$ sha256sum docs/migration/openapi.latest.json
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49   ← 백업과 동일
$ git diff --stat docs/
 docs/pm/portfolio/_registry.yaml | 4 +++-      ← 이 change의 것뿐. 스냅샷 변경 없음
```

### A.3 duration `"realtime"` → `"1d"` **+ gainers를 패널에 복원** → GREEN ✔

```
$ go test ./internal/candidatesrc/ -run TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused -v
--- PASS
```

두 변이를 **동시에** 넣은 상태에서 통과한다. 금지되는 것이 `TOP_GAINERS`가 아니라
`(TOP_GAINERS, realtime)` **조합**이라는 것이 이것으로 증명된다. D3가 요구사항 수준에서
말한 "형태 금지"를 코드가 지킨다.

### A.4 정규식이 **실파일**에서 무엇을 뽑는가 — fixture가 아니라 실측

`snapshot_drift_test.go:100-101`의 정규식을 그대로 복사한 독립 프로그램을 저장소 밖에서
실행해, 실제 `docs/migration/openapi.latest.json`의 `type` 파라미터 description에 돌렸다.

```
description length: 589 bytes
  MATCH  type="TOP_GAINERS" duration="realtime"  raw="- `TOP_GAINERS`: 급상승 (등락률 상위) — `realtime` 미지원"
  MATCH  type="TOP_LOSERS"  duration="realtime"  raw="- `TOP_LOSERS`: 급하락 (등락률 하위) — `realtime` 미지원"

forbidden map has 1 duration key(s): [realtime]
  "realtime" -> [TOP_GAINERS TOP_LOSERS]

--- same matcher over the entire raw JSON file ---
matches in raw file bytes: 0
```

정확히 `{TOP_GAINERS, TOP_LOSERS}`다. 부수 확인: 원시 바이트에 대해서는 0건이므로
(`\n`이 JSON 이스케이프라 `(?m)^-`가 걸리지 않는다) **JSON 파싱이 실제로 하중을 받는다** —
텍스트 grep으로 대체할 수 없는 형태다.

## B. 냉각 불변식(§5.2/§5.3) — 직접 확인

### B.1 `coverageAnswered`에서 `if !heard[id] { continue }` 제거 → §5.2 RED ✔

```
$ go test ./internal/candidate/
[FAIL] TestALiveSupporterStillCoolsACandidateARetiredSourceAlsoRaised
   retiredsupporter_cooling_test.go:72: cooled = 0, want 1 …
```

§5.3(`TestACandidateWhoseOnlySupporterIsRetiredIsNeverCooled`)은 같은 변이에서 **PASS로
남았다.** 두 테스트가 서로 다른 것을 재고 있다는 증거다. 되돌린 뒤
`git diff internal/candidate/scan.go`는 비었고 패키지 전체 GREEN.

**다만 §5.2가 유일한 가드는 아니다** — 같은 변이에서 기존 테스트 둘도 RED다:
`scan_test.go:687 TestASupporterThatLeftThePanelDoesNotBlockCoolingForever`,
`notdue_test.go:531 TestASourceTheSchedulePassedOverIsNotASourceThatIsGone`.
§5.2는 새 커버리지라기보다 **이 change의 논증을 그 이름으로 고정한 기록**이다. 결함은
아니지만, 누구도 §5.2를 이 불변식의 단독 가드로 세지 않도록 적어 둔다. (F9)

### B.2 §5.3이 고정한 것은 **실제 현재 동작**인가 — `coolAbsent`를 직접 읽었다

꾸민 상황이 아니다. 코드가 그렇게 쓰여 있고, 주석이 그렇게 선언하고 있다.

- `coverageAnswered`([scan.go:704-716])는 마지막이 `return present > 0`이다.
- `coolAbsent`([scan.go:685-687])는 `if !coverageAnswered(...) { continue }`다.
- `coolAbsent`의 doc([scan.go:671-672])이 **명시적으로** 적고 있다:
  *"A candidate whose supporters are *all* gone is still left alone: nothing in this
  scan was in a position to see it."*

§5.3은 그 문장을 실행 가능한 형태로 옮긴 것이고, 테스트가 상황을 꾸며 만든 성질이 아니다.

### B.3 §5.1 저장소 측정 — 리뷰어가 다시 쟀다

운영 store를 scratchpad로 **복사한 뒤** 복사본을 읽기 전용으로 조회했다(원본 무변경).

```
candidates 총계: 386
sources LIKE '%official_rankings_top_gainers%'         0
observations WHERE source='official_rankings_top_gainers'   0
observations by source: trading_value 300 · trading_volume 300 · wts_popular 30
```

총계는 구현자 보고(300 / 200·200·30)에서 움직였다 — 그 뒤로 스캔이 더 돌았다. **하중을
받는 두 숫자(0, 0)는 동일하다.** 안전 논증의 전제가 성립한다. 부수적으로 관측이
`trading_value 300 = trading_volume 300`으로 정확히 같고 gainers가 0인 것은,
§7.3의 "2 attempted, 2 responded"를 store 쪽에서 독립적으로 뒷받침한다.

## C. 일정/패널 방향성(§3.2) — 등호가 아님을 코드와 변이 양쪽에서

**코드**: `scheduledWithNoSource`([candidateschedule_drift_test.go:96-107])는 `intervals`만
순회하며 `buildable[id]`를 본다. 반대 방향은 계산하지 않는다. **포함이지 등호가 아니다.**

**변이 ①(RED여야 함) 일정에만 gainers**:

```
--- FAIL: TestNoIntervalNamesASourceNoPanelBuilds
    candidateschedule_drift_test.go:118: candidateIntervals sets a cadence for
    [official_rankings_top_gainers] and no market's panel builds those sources.
```

**변이 ②(GREEN이어야 함) 패널에만 gainers, 일정에서 제외**:

```
--- PASS: TestNoIntervalNamesASourceNoPanelBuilds
--- PASS: TestTheScheduleGuardIsContainmentAndNotEquality
```

`unconfiguredFloor` 설계는 금지되지 않는다. 근거도 확인했다 —
[source.go:184-196]가 15초를 "계약상 가장 느린 간격"으로 정의하고,
[source.go:246-249]가 미구성 원천에 그것을 준다. 테스트 주석의 서술이 코드와 일치한다.

## D. 구현자가 스스로 제기한 세 가지 — 판정

### D.1 `make gate` — 이 절 끝에 전문

### D.2 tasks.md §3.4 정정은 **정확하다** ✔

`declaredPanelSizes`([panelsize_drift_test.go:44-83])는 AST를 `ast.Inspect`로 훑어
`*ast.CallExpr` 중 `Fun`이 `OfficialRanking`/`WTSPopular`인 것의 INT 리터럴 인자(>1)를
모은다. 소스에 `OfficialRanking(...)`은 **한 번**(리터럴 순회 안) 나오므로 이 값은 change
전후 모두 `{100, 30}`이다. 초안의 *"호출이 세 번에서 두 번으로 줄어도"*가 가리키는 세 번은
소스에 존재한 적이 없다. Manager의 정정 기록(tasks.md:84-90)이 맞다.

### D.3 헬퍼 파일 이동은 **게이트 회피가 아니다** ✔ — 판정

세 가지를 직접 대조했다.

1. **옮겨진 함수는 새 함수다.** `git show HEAD:internal/candidatesrc/clock_wiring_test.go`의
   함수 목록은 `aFullRanking`·`aFullPopularity`·`TestThePanelHandsItsClockToEverySourceItBuilds`·
   `TestAMemoryWithNoInstantOnEitherSideIsNotAnAnswer` 넷뿐이다. `sameSourceSet`·`sortedIDs`는
   base에 **존재하지 않는다.** 기존 함수를 새 파일로 옮겨 diff를 숨긴 것이 아니다.
2. **탈락한 것은 `aFullRanking` 하나이고, 그것은 손대지 않았다.** 현재 diff에
   `aFullRanking` 변경이 없다. 첫 실행에서 6건이 잡힌 것은 새 헬퍼가 그 함수와 hunk 줄
   범위를 공유했기 때문이고(`check_analysis.py:129,141-144`가 hunk와 겹치는 함수를 잡는다),
   내용 변경이 아니라 **인접성**이다.
3. **도구의 규칙과 일치한다.** `changed_existing_functions`는 `--- /dev/null`인 파일에서
   `flush()`가 `if not old_source ... return`으로 빠져나가므로 새 파일 함수는 규칙상 면제다.
   현재 대상은 3건이고 design D7의 예측과 같다:

```
$ python3 -c "…ca.changed_existing_functions(base=…)"
 TARGET: cmd/tossctl/candidate.go :: candidateIntervals
 TARGET: internal/candidatesrc/candidatesrc.go :: Panel
 TARGET: internal/candidatesrc/clock_wiring_test.go :: TestThePanelHandsItsClockToEverySourceItBuilds
count: 3
```

**판정: 회피 아님.** 실제 로직 변경이 새 파일로 숨겨진 사실은 없다.
(단 산출물의 **내용**에는 문제가 있다 — F1 참조.)

## E. `candidatesrc_test.go:290` — 판단은 **맞다** ✔

*"The three official rankings shipped under one source id."* — 동사가 `shipped`이고,
이어지는 문장 전체가 과거형이다(*"the two rankings that answered vouched for the one that
was rate limited"*). §2 리뷰가 고친 id 충돌 결함의 **서술**이지 현재 패널 구성에 대한
주장이 아니다. 그 뒤의 현재형 문장(*"the check is keyed by id"*)은 지금도 참이고,
`rankingSourceID`가 세 항목을 유지하므로 "세 랭킹"이라는 지시 대상도 남아 있다.
**표류가 아니다. I4의 판단을 수용한다.**

---

## 리뷰어가 스스로 찾은 것

### F1 (P1) — `Panel`의 버려진 오류를, 막지 못하는 가드가 막는다고 **세 곳**이 말한다

**파일:줄**
- `internal/candidatesrc/candidatesrc.go:560-563` (이 change가 다시 쓴 문장)
- `analysis/function-logic/internal-candidatesrc--panel/function-logic-map.md:21, 32`
- `analysis/function-logic/internal-candidatesrc--panel/branch-test-map.md:7` (B3 행)

세 곳 모두 *"a failure would be a defect in this file, and `TestEveryPanelSourceHasItsOwnID`
fails if one ever slips"* / B3의 Required test = `TestAnUnknownRankingTypeIsRefused` ·
`TestEveryPanelSourceHasItsOwnID`라고 적는다.

**실패 시나리오(실행함)**: `rankingSourceID` 항목 없이 랭킹 타입 하나를 리터럴에 넣는다 —
스냅샷 enum에 실재하는 값을 골랐다.

```go
const ( … RankingTossAmount = "TOSS_SECURITIES_TRADING_AMOUNT" )   // 매핑 없음
for _, typ := range []string{RankingTradingAmount, RankingTradingVolume, RankingTossAmount} {
```

```
$ go test ./internal/candidatesrc/
ok   github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc   0.008s
```

**패키지 전체가 GREEN이다.** `OfficialRanking`이 오류를 내고 B3가 그 원천을 **조용히
버리므로**, 패널은 그대로 `{trading_value, trading_volume, wts_popular}`이고
`TestEveryPanelSourceHasItsOwnID`(중복 id·비어있음만 본다),
`clock_wiring_test.go`의 새 id 집합 단언(집합이 안 변한다),
`snapshot_drift_test.go`(그 타입은 금지 조합이 아니다) 어느 것도 걸리지 않는다.
`TestAnUnknownRankingTypeIsRefused`는 **생성자를 직접** 부르는 테스트라 `Panel` 경로에
대해 아무 말도 하지 않는다.

결과: 누군가 원천을 하나 더 넣었다고 믿는데 패널은 둘을 읽는다. 주석·FLM·Branch Test
Map은 셋이라고 말하고, 실패하는 것은 없다. **이 change가 없애려는 실패 형태 그 자체가,
이 change가 편집하는 함수 안에, 그것을 막는다고 주장하는 문장 바로 밑에 있다.**

**성격**: 이 주장은 base에도 있었다(문구만 달랐다). 그러나 이 change가 그 문장을 **다시
썼고**, 새 게이트 산출물 두 곳(FLM·Branch Test Map)에 **다시 승인**했다. B3의
`RED observed: no`는 정직하지만, "Test" 칸이 커버하지 않는 테스트를 지목하고 있다.

**최소 수정**: 코드 변경 없이 문장 셋을 사실로 바꾼다 — B3의 오류 분기는 `Panel` 경로에서
도달 불가능하고, 도달하게 만드는 편집(매핑 없는 타입 추가)을 **잡는 테스트는 없다**.
가드를 실제로 만드는 것은 별도 change다(범위 밖). 판정은 Manager에게 맡긴다.

### F2 (P1) — 스냅샷 가드의 positive control이 스냅샷이 아니라 **스냅샷의 사본**을 읽는다

**파일:줄**: `internal/candidatesrc/snapshot_drift_test.go:344-372`
(특히 `asWritten` 하드코딩 fixture, 345-350)

`TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor`는 진짜 matcher를 **자기
fixture**에 돌린다. 실파일에 대한 완전성 검사는 본 가드의 `len(forbidden) == 0` Fatal
하나뿐이다. 그러므로 **금지 집합이 2에서 1로 줄어드는 표류는 어디에도 걸리지 않는다.**

**실패 시나리오(실행함)**: 실파일에서 `TOP_GAINERS` 불릿의 `— \`realtime\` 미지원`만
지우고(`TOP_LOSERS`는 그대로), gainers를 패널에 되돌린다.

```
$ go test ./internal/candidatesrc/ -run 'TestThePanelAsksForNothing…|TestTheSnapshotMatcher…'
ok   github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc   0.007s
```

**통과한다.** `forbidden = {realtime: {TOP_LOSERS}}`로 비어 있지 않아 Fatal이 안 걸리고,
positive control은 자기 fixture를 읽으므로 그대로 초록이며, `Panel`은 `TOP_GAINERS`를
`realtime`으로 부른다. 46일간 아무도 스냅샷을 읽지 않았던 이 change의 원인이, 스냅샷을
읽으라고 만든 테스트 안에서 한 칸 좁아진 형태로 재현된다.

**최소 수정(한 줄)**: 본 가드에서 실파일 결과에 대해 `forbidden["realtime"]`이
`TOP_GAINERS`와 `TOP_LOSERS`를 **둘 다** 담는지 단언한다. 또는 파일에서 읽은 조합 수를
파일의 `미지원` 출현 수와 대조한다. fixture는 그대로 두어도 된다 — 문제는 fixture가
아니라 실파일에 대한 완전성 검사가 "0인가"뿐이라는 것이다.

### F3 (P1, 랜딩 절차) — `_registry.yaml`이 **이 커밋에 없는 change**를 allowlist한다

**파일:줄**: `docs/pm/portfolio/_registry.yaml:29`
(`"refine-extended-shadow-bands"`), `tools/pm/test_generate_master_tracker.py:35`

`generate_master_tracker.py:139-140`은 allowlist에 있는데 `active_changes`에 없는 항목을
`stale bootstrap exception`으로 **실패시킨다**. `active_changes:45-51`은 파일시스템에서
`proposal.md`가 있는 디렉터리를 읽는다. `openspec/changes/refine-extended-shadow-bands/`는
지금 **untracked**이고 이 change의 범위 밖이다.

**실패 시나리오(실행함)**: 저장소 사본에서 그 디렉터리만 빼고 검사 —
이것이 `retire-gainers-source`만 커밋했을 때의 clean checkout 상태다.

```
$ python3 tools/pm/generate_master_tracker.py --root <사본> --check
[pm] invalid:
  - refine-extended-shadow-bands: stale bootstrap exception
```

`make sdd-check`가 이 검사를 돌린다([Makefile:59]). 즉 **이 커밋을 clean checkout하면
`make sdd-check`가 깨진다.** 지금 로컬에서 통과하는 것은 두 디렉터리가 **둘 다 작업
트리에 존재**하기 때문이고, 그 사실은 커밋에 담기지 않는다.
tasks.md §7 자신이 *"이 change가 먼저 랜딩한다 … 두 change는 병행하지 않는다"*라고
적었으므로, 먼저 랜딩하는 커밋이 뒤 change의 등록을 들고 가는 것은 그 문장과도 어긋난다.

**해법 둘 중 하나**: ① 이 커밋에서 `refine-extended-shadow-bands` 두 항목을 뺀다(그
change가 자기 등록을 들고 온다), 또는 ② 두 change 디렉터리를 같은 커밋에 담는다.
①이 §7의 순서 선언과 맞다. **랜딩 전에 결정해야 한다.**

### F4 (P2) — `wiredRankings`는 `Panel` 안의 **무관한 `[]string` 리터럴**에 눈이 먼다

**파일:줄**: `internal/candidatesrc/snapshot_drift_test.go:162-207`
(특히 `len(out) == 0` Fatal, 201)

`Panel` 본문의 **모든** `[]string` 합성 리터럴을 합집합으로 모으고, 합집합이 빌 때만
Fatal한다. 그래서 랭킹 목록이 리터럴 자리를 떠나고 **다른** `[]string` 리터럴이 하나라도
남으면, 공허 방지 Fatal이 무력화된다.

**실패 시나리오(실행함)**: 랭킹 목록을 패키지 수준 `var officialPanelRankings = []string{…,
RankingTopGainers}`로 올리고(D2가 "되돌림은 슬라이스 원소 하나"라고 한 만큼 자연스러운
리팩터다), `Panel` 안에 무관한 `[]string` 리터럴을 하나 남긴다.

```
$ go test ./internal/candidatesrc/ -run TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
ok   (PASS)          ← Panel은 TOP_GAINERS를 realtime으로 부르고 있다
```

같은 상태에서 id 기반 테스트(`TestNoMarketPanelBuildsTheGainersRanking`,
`clock_wiring_test.go`)는 RED이므로 **gainers에 한해서는** 다른 그물이 있다. 눈이 머는
것은 **일반 가드** 쪽 — `TOP_LOSERS`와 앞으로 생길 타입, 즉 D8이 이 테스트를 만든 바로
그 이유다.

반대 방향(시끄러운 쪽)도 있다: `Panel`에 `[]string{candidate.MarketKR}` 같은 리터럴이
생기면 원소가 `*ast.SelectorExpr`이라 `t.Fatalf("Panel builds a ranking from …")`로
**무관한 이유로 죽는다**. 안전한 방향이지만 유지보수 함정이다.

**최소 수정**: 리터럴을 찾는 대신 `range` 대상 식을 따라가거나, 인식한 리터럴이
`rankingSourceID`의 키 집합에 속하는 값만 담는지 확인한다.

### F5 (P2) — 가드는 타입×duration을 **교차곱**으로 짝짓는다. 재검토 시점에 오탐이 된다

`wiredRankings`는 `Panel`에서, `wiredDurations`는 **파일 전체의 모든 `.Rankings(` 호출**에서
읽고, 단언은 두 집합의 곱 위에서 돈다([snapshot_drift_test.go:317-333]). "이 타입은 이
duration으로 부른다"는 짝은 표현되지 않는다.

그래서 proposal의 **재검토 조건 4번**(다른 창에서 잰 순위를 다루는 설계)이 충족되어
gainers가 `1d`로, 나머지가 `realtime`으로 돌아오는 날, 이 가드는 위반이 아닌 것을 위반으로
보고한다. 그리고 그때 그것을 느슨하게 만들 사람은 이 문서를 안 읽은 사람이다.
**지금 근거가 눈앞에 있을 때 적어 둔다.** 이 change에서 고치지 않는다(범위 밖).

### F6 (P2) — 가드가 보는 파일은 하나다. 랭킹 요청을 만드는 곳은 둘이다

`cmd/tossctl/market.go:332, 339-341`도 `client.Rankings(...)`를 부르고, `--type` 도움말이
`TOP_GAINERS|TOP_LOSERS`를 광고한다. **확인 결과 위반은 아니다** — `--duration` 기본값이
`1d`라 기본 조합은 합법이고, 타입·duration이 둘 다 플래그라서 정적으로 대조할 짝이 없다.
다음 사람이 "가드가 저장소 전체를 본다"고 오해하지 않도록 기록한다.

### F7 (P2) — D2의 세 근거 중 하나는 코드가 강제하지 않는다 (결론은 맞다)

`internal/candidate/candidate.go:95-97`: *"The id stays because an observation already
written under it has to read back as something the system recognises"*.
같은 근거가 `retiredsource_test.go:95-97`에도 있다.

`decodeSources`([store.go:2077-2089])는 임의 문자열을 `SourceID(p)`로 만들 뿐 검증하지
않는다. 상수를 지워도 저장된 행이 읽히지 않게 되는 일은 없다. 게다가 §5.1 측정대로 그런
행은 0건이다. D2의 나머지 두 근거(되돌림이 한 편집, 논거가 주석에 산다)는 독립적으로
성립하므로 **결론은 맞다.** 다만 이 change는 같은 형태의 오류를 이미 두 번
정정했다(`NoteSources`, §3.4). 이것이 **세 번째**다.

### F8 (P2) — `clock_wiring_test.go`가, 없어지기로 되어 있는 파일에 의존한다

`sameSourceSet`·`sortedIDs`가 `retiredsource_test.go:36-56`에 있고
`clock_wiring_test.go:88`이 그것을 쓴다. `retiredsource_test.go`의 헤더는 그 파일이
*"pins the one thing `retire-gainers-source` changes"*라고 선언한다 — gainers가 돌아오면
지워질 후보다. 지우면 시계 테스트가 **컴파일되지 않는다**. 시끄러운 실패라 안전하지만
결합 방향이 거꾸로다.

### F9 (P2, 정보) — §5.2는 이 불변식의 단독 가드가 아니다

B.1 참조. 같은 변이에서 기존 테스트 둘이 함께 RED다. 결함이 아니라, 아무도 §5.2를
유일한 그물로 세지 않도록 남기는 기록이다.

---

## 안전 불변식 검토

- 변경 파일 7개 + 신규 테스트 4개. `internal/risk`·`internal/execgw`·`internal/exitpolicy`
  **전부 무변경**. 주문 제출·취소·정정, 손절·익절·사이징, Guardian·kill switch, intent
  journal·원장, reconciliation, retry matrix, 인증·세션, 체결 감지 어느 경로도 닿지 않는다.
- `off.Note`/`Backoff`/`Cycle`의 429 사다리 무변경(D4 준수). `internal/candidate/scan.go`
  무변경.
- 호출량은 단조 감소(순회당 RANKING 3 → 2). 안전 불변식 §0.4에 대해 보수 방향.
- 토글·운영 설정 flip 없음. LIVE 주문 side effect 없음. 리뷰 중 실행한 것은 `go test`,
  `go vet`, 정적 분석, 저장소 **사본** 조회뿐이다.
- **`tossctl candidate scan`은 재실행하지 않았다.** 읽기 전용이지만 운영 store에 관측을
  쓰고 RANKING rate 예산을 쓴다 — 이 change가 아끼기로 한 그 예산이다. 대신 §7.3의 주장을
  store 측 상태로 교차검증했다(B.3): 두 공식 랭킹의 관측이 300/300으로 정확히 같고
  gainers는 0이다. `attempted == responded`·`(degraded)` 없음과 모순되는 흔적이 없다.

## 찾아봤고 **없었던** 것

발견이 없다는 보고가 정보가 되도록, 무엇을 훑었는지 적는다.

- **공허하게 참인 단언**: 신규 4개 파일의 모든 단언 경로에 비어있음 방지가 있다 —
  `retiredsource_test.go:68`, `candidateschedule_drift_test.go:79·113`,
  `snapshot_drift_test.go:201·250·304`, `clock_wiring_test.go:81`. 각각을 읽고
  무엇이 그것을 발동시키는지 확인했다. (`snapshot_drift_test.go:201`은 F4로 우회 가능.)
- **지금 거짓인 새 주석**: F1 하나. 나머지는 검증했다 —
  46일 주장(`ba2cf1a` 2026-06-12 → `5c1ced7` 2026-07-28 = 정확히 46일),
  `unconfiguredFloor` 서술([source.go:184-196, 246-249]과 일치),
  *"mapped to a source id below"*(`rankingSourceID`가 바로 아래, :171),
  `notdue_test.go`의 새 US 문장(US 패널 = 공식 랭킹 둘, 같은 interval — 맞다),
  `clock_wiring_test.go` 헤더의 *"the count … is not written down anywhere in this file"*
  (파일에 개수 주장 없음 확인).
- **표류한 개수 주석 잔여**: `grep -rnE "three official|세 랭킹|three rankings|four
  sources|three sources"`로 전체 훑음. 남은 것은 `candidatesrc_test.go:290`(E — 역사적
  서술, 판단 수용)과 `retiredsource_test.go`의 의도적 인용뿐.
- **§6.2 문서**: `docs/`에 발굴 원천 **개수**를 말하는 문장 없음, `baseline.md`·`ROADMAP.md`
  에 gainers 언급 없음. "갱신할 곳이 없다"는 보고가 맞다.
- **남은 gainers 배선**: 비테스트 코드에서 `SourceOfficialGainers`를 쓰는 곳은
  `rankingSourceID`의 항목 하나뿐이고, 그 항목은 `RankingTopGainers`를
  `OfficialRanking`에 넘기는 호출부가 없으므로 도달하지 않는다. 의도된 상태다(D2).
- **`panelsize_drift_test.go` 결합**: `declaredPanelSizes`를 읽고 D.2에서 확인.
  새 주석이 `숫자+rows/행` 형태를 쓰지 않아 무관하게 걸리지 않는다(전체 스위트 GREEN).
- **FLM 산출물**: 세 디렉터리 모두 TODO 0건, 템플릿이 아닌 실제 내용.
  `check_analysis.py`의 source SHA 대조 통과.

---

## 검증 명령과 실제 결과 (리뷰어 실행)

| 명령 | 결과 |
|---|---|
| `go test ./...` | **3685 passed in 57 packages** |
| `go vet ./...` | No issues found |
| `gofmt -l internal/ cmd/` | 출력 없음 |
| `python3 tools/logic-map/check_analysis.py --change retire-gainers-source` | `evidence complete or diff-proven exempt` |
| `python3 tools/pm/generate_master_tracker.py --check` | `hierarchy and generated trackers are current` (단 F3 참조 — 현재 작업 트리에서만) |

변이를 전부 되돌린 뒤 `git diff --stat`이 리뷰 시작 시점과 **동일**함을 확인했다
(7 files, 108 insertions, 22 deletions). `internal/candidatesrc/candidatesrc.go`의 hunk를
raw diff로 대조했고, `check_analysis.py`의 `Panel` source SHA 대조가 바이트 수준에서
그것을 다시 확인한다.

## `make sdd-sync` → `make sdd-check` → `make gate` (리뷰어 실행, 연속)

`.go` 편집(변이)을 전부 되돌린 뒤 연속 실행했다.

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph `Already up to date`, CodeGraphContext, GBrain) |
| `make sdd-check` | 통과 — agent-config synchronized · memory `valid` · `[index-freshness] CodeGraph hard-evidence index matches the worktree` · `[pm] hierarchy and generated trackers are current` · sdd-test 전부 통과 |
| `make gate CHANGE=retire-gainers-source` | **GATE PASS** (exit 0) |

게이트 전문:

```text
GATE: retire-gainers-source
repo: /mnt/D/project/axipient/TossOS

==> 1/8 tasks.md 확인
OK: openspec/changes/retire-gainers-source/tasks.md

==> 2/8 미완료 태스크 확인
OK: 미완료 태스크 0건

==> 3/8 gstack 리뷰 기록 확인
OK: openspec/changes/retire-gainers-source/review.md

==> 4/8 Function Logic Map 증거 확인
[logic-map] retire-gainers-source: evidence complete or diff-proven exempt
OK: Function Logic Map

==> 5/8 make sdd-check
[agent-config] Claude/Codex safety bootstrap and workflow routing are synchronized
valid
[index-freshness] CodeGraph hard-evidence index matches the worktree
[pm] hierarchy and generated trackers are current
OK: make sdd-check

==> 6/8 make test
OK: make test

==> 7/8 make vet
OK: make vet

==> 8/8 make validate
Totals: 25 passed, 0 failed (25 items)
OK: make validate

GATE PASS: retire-gainers-source
```

구현자 보고대로 실패 지점은 §7.4 하나였고, 리뷰가 그것을 닫자 8단계 전부 통과한다.
**게이트 통과는 랜딩 승인이 아니다.** 게이트는 아래 세 가지 중 어느 것도 검사하지
않는다 — F1은 주석과 산출물의 **내용**이고, F2는 통과하는 가드의 **구멍**이며,
F3은 clean checkout에서만 드러나는 **커밋 구성**이다. 게이트가 초록인 채로 시스템이
뜻을 바꾸는 형태는 이 change가 고치려는 바로 그것이다.

---

## 최종 판정

### **랜딩 불가 — 사유 3건**

이 change의 **핵심 결정은 옳다.** 답할 수 없는 원천을 내린 것, `1d`로 바꾸지 않은 것,
상수를 남긴 것, 요구사항을 타입이 아니라 **형태**로 쓴 것 — 넷 다 근거가 코드와 측정으로
확인된다. §4 변이 세 가지는 리뷰어가 직접 넣어 전부 설계대로 나왔고(변이 ③은 gainers
복원과 `1d`를 **동시에** 적용한 상태에서 GREEN), §5.1은 다시 재도 0/0이며, §3.2의 방향은
등호가 아니라 포함임이 코드와 양방향 변이로 확인된다. 안전 불변식 위반은 없다.

랜딩을 막는 것은 셋이고, 셋 다 값이 싸다.

| # | 사유 | 필요한 것 |
|---|---|---|
| **F3** | `_registry.yaml`이 이 커밋에 없는 change를 allowlist한다. clean checkout에서 `make sdd-check`가 `stale bootstrap exception`으로 깨진다(재현 완료) | 커밋 전 결정: 두 항목을 이 커밋에서 빼거나, 두 change 디렉터리를 같이 커밋한다 |
| **F1** | 게이트 산출물(Branch Test Map B3)과 코드 주석이 **커버하지 않는 테스트를 커버한다고 지목**한다. 재현 완료 — 매핑 없는 랭킹 타입을 패널에 넣으면 패키지 전체 GREEN | 문장 정정. FLM의 "Required test" 칸이 거짓이면 FLM 절차 자체가 형식이 된다 |
| **F2** | 이 change의 **핵심 산출물**이 실파일에 대해 "0인가"만 검사한다. 금지 집합이 2→1로 줄면 가드가 조용히 통과한다(재현 완료) | 한 줄 — 실파일 결과에 두 타입이 다 있는지 단언 |

F4~F9는 랜딩을 막지 않는다. F5·F6은 재검토 시점에 필요한 기록이고, F4·F7·F8은
후속 정리 대상이며, F9는 정보다. issues.md에 남길지는 Manager가 정한다.

**한 문장으로**: 결정은 맞고 배선은 맞다. 남은 것은 **이 change가 남기는 기록이
사실인가**이고, 그것이 이 change의 주제였다.


---

# §8 수정 (구현)

- 날짜: 2026-07-29
- 역할: **Teammate(구현자).** 위 두 절(구현 기록·독립 리뷰)은 지우지 않았다.
- 범위: **§8.1 · 8.2 · 8.3 · 8.5 · 8.6.**
  **§8.4(커밋 절차)와 §8.7(2차 독립 리뷰)은 하지 않았다** — 전자는 Manager가 쥔 커밋
  절차이고 후자는 구현자가 자기 수정을 독립 검증할 수 없다. 커밋·push 하지 않았다.
- 변이 정책: `git checkout --`·`git restore`·`git stash`를 **한 번도 쓰지 않았다.**
  모든 변이는 편집으로 넣고 편집으로 되돌렸으며, 스냅샷만 편집 전 scratchpad 사본에서
  복원했다. 되돌린 뒤 매번 `sha256sum`으로 원본과 대조했다.

## 편집 전 기준 해시

```text
fe6daa185e8ccefdfce412f63e65c711f84e1ebe71fd1d2f0f3f044298da205c  internal/candidatesrc/candidatesrc.go
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49  docs/migration/openapi.latest.json
5481d7100f4d513eed5bdc64fdce3a5da64f0d40b891aaa5e3c47554d61c71d2  internal/candidatesrc/snapshot_drift_test.go
```

---

## 1. F1 재현 — 결함이 실재함을 먼저 보였다 (§8.1)

가드를 쓰기 **전에** 결함을 실행으로 확인했다. 리터럴 원소는 `wiredRankings`가
`*ast.Ident`만 받으므로 상수를 먼저 선언해야 한다(아래 §7의 정밀도 지적 참조).

**넣은 변이**

```go
const (
	RankingTradingAmount = "MARKET_TRADING_AMOUNT"
	RankingTradingVolume = "MARKET_TRADING_VOLUME"
	RankingTopGainers    = "TOP_GAINERS"
	RankingTossAmount    = "TOSS_SECURITIES_TRADING_AMOUNT" // rankingSourceID에 항목 없음
)
...
	for _, typ := range []string{RankingTradingAmount, RankingTradingVolume, RankingTossAmount} {
```

**결과 — 초록이다. 결함 확인.**

```text
$ go test ./internal/candidatesrc/ -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc	0.009s
```

패키지 53건 전부 통과했다. `OfficialRanking`이 오류를 내고 `Panel`이 그것을 버리므로
패널은 `{trading_value, trading_volume, wts_popular}` 그대로이고,
`TestEveryPanelSourceHasItsOwnID`(id 중복·비어있음만 본다),
`clock_wiring_test.go`의 id 집합 단언(집합이 안 바뀐다),
`snapshot_drift_test.go`(그 타입은 금지 조합이 아니다) 어느 것도 걸리지 않는다.
**리뷰 F1이 맞다.**

부수 증거: 같은 변이에서 `TestThePanelHandsItsClockToEverySourceItBuilds`가 통과했다는
사실이 곧 "원천이 조용히 빠졌다"의 증명이다 — 그 테스트가 KR 패널의 id 집합이 **정확히**
셋과 같음을 단언하는데, 넷째 타입이 원천을 만들었다면 그 단언이 깨졌을 것이다.

---

## 2. 새 가드 — 형태와 **왜 그 형태인지** (§8.1)

`TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`
(`internal/candidatesrc/snapshot_drift_test.go`).

tasks §8.1의 권장 형태를 **그대로 채택했다.** 두 집합의 **등호**다.

```text
declared  AST가 candidatesrc.go에서 읽은 Panel의 랭킹 타입 리터럴 값 (wiredRankings)
built     Panel(...)을 실제로 호출해 얻은 원천의 id를 rankingSourceID로 되돌린 타입 집합
```

권장 형태를 바꾸지 않은 이유는 검증했기 때문이다 — 요구된 두 성질이 **둘 다 실측으로**
성립한다(아래 §3). 다만 구현에서 세 가지를 스스로 정했고 근거를 적는다.

**① 두 시장의 합집합으로 `built`를 만든다.** `Panel`의 멤버십 규칙은 시장별이다(그 함수의
doc). 한 시장만 물으면 다른 시장에만 있는 랭킹이 "리터럴에 있는데 안 만들어졌다"로
오탐이 된다. 지금 랭킹 루프는 시장 게이트 밖에 있으므로 KR·US가 같지만, 합집합으로 쓰면
누군가 랭킹을 시장별로 나눠도 이 가드가 그 설계를 금지하지 않는다. `declared` 쪽도
`Panel` 안 모든 리터럴의 합집합이므로 **합집합 대 합집합**으로 대칭이다.

**② `rankingSourceID`의 역매핑에서 id 충돌은 Fatal이다.** 두 타입이 한 id를 쓰는 것은
§2 리뷰의 P0 그 자체다. 역매핑이 조용히 하나만 남기면 이 가드가 **나머지 하나를 "만들어지지
않은 타입"으로 오보**한다. 가드가 자기 입력의 결함을 결론으로 바꾸지 않게 막았다.

**③ 실패 메시지가 두 방향을 구분한다.** `declared`에만 있으면 "원천이 조용히 버려졌다"(F1),
`built`에만 있으면 "AST 리더가 배선을 못 보고 있다"(F4 ①). 같은 등호 위반이지만 고칠 곳이
다르다.

**공허하게 참일 수 없다**: `declared`가 비면 `wiredRankings`가 이미 Fatal하고,
`built`가 비면 첫 루프가 리터럴의 모든 타입에 대해 실패한다. 새로 넣은 비어있음 검사는 없다.

### 어느 파일에 두었는지, 그리고 왜

**`snapshot_drift_test.go`에 두었다** — 새 파일이 아니다. tasks의 "새 테스트는 새 파일에"
규칙은 **base 대비 diff를 좁게 유지**하려는 것인데(D2), 이 파일은 이 change에서 통째로
새 파일이라 그 목적에 걸리지 않는다. `check_analysis.py`도 `--- /dev/null` 파일의 함수를
규칙상 면제하므로 FLM 대상이 늘지 않는다(재실행으로 확인 — 대상 3건 그대로).

두는 편이 나은 이유는 **이 가드가 고정하는 대상이 바로 위의 `wiredRankings`**이기 때문이다.
`wiredRankings`를 읽은 사람이 "그런데 저 AST 읽기가 진짜 배선을 읽는다는 보장은?"에 대한
답을 다른 파일에서 찾게 만들 이유가 없다. 반대로 새 파일로 빼면 그 파일이
`wiredRankings`·`parseWiring`·`sorted`를 가져다 쓰게 되어 **F8이 지적한 결합 방향 문제**를
하나 더 만든다.

---

## 3. 새 가드의 변이 확인 — 요구된 두 성질 (§8.1)

### 3.1 오류로 조용히 빠진 원천이 **실패**로 나타난다

§1의 변이를 그대로 둔 채 가드를 추가했다.

```text
$ go test ./internal/candidatesrc/ -count=1 -run TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds -v
=== RUN   TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
    snapshot_drift_test.go:301: candidatesrc.go names TOSS_SECURITIES_TRADING_AMOUNT in
    Panel's ranking literal and Panel builds no source for it.
    ...
    Named in the literal: [MARKET_TRADING_AMOUNT MARKET_TRADING_VOLUME TOSS_SECURITIES_TRADING_AMOUNT]
    Built by Panel: [MARKET_TRADING_AMOUNT MARKET_TRADING_VOLUME]
--- FAIL: TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds (0.00s)
```

**RED ✔.** 변이를 되돌린 뒤:

```text
$ sha256sum internal/candidatesrc/candidatesrc.go
fe6daa185e8ccefdfce412f63e65c711f84e1ebe71fd1d2f0f3f044298da205c   ← 기준과 동일
$ go test ./internal/candidatesrc/ -count=1
ok  ...  0.007s
```

### 3.2 리터럴이 `Panel` 밖으로 옮겨지면 **실패**로 나타난다 (리뷰 F4 ①)

**넣은 변이** — F4가 서술한 리팩터를 그대로 재현했다:

```go
const ( ... mutationF4Unrelated = "SOMETHING_ELSE" )
var mutationF4Rankings = []string{RankingTradingAmount, RankingTradingVolume}
...
	_ = []string{mutationF4Unrelated}   // 무관한 리터럴을 Panel 안에 남긴다
	for _, typ := range mutationF4Rankings {
```

```text
--- FAIL: TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
    snapshot_drift_test.go:301: candidatesrc.go names SOMETHING_ELSE ... and Panel builds no source for it.
    snapshot_drift_test.go:317: Panel builds a source for MARKET_TRADING_AMOUNT and no ranking literal in Panel names it.
    snapshot_drift_test.go:317: Panel builds a source for MARKET_TRADING_VOLUME and no ranking literal in Panel names it.
--- PASS: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
```

**RED ✔, 그리고 같은 실행에서 스냅샷 가드는 PASS다.** 그것이 F4가 말한 눈멂이고,
새 가드가 정확히 그 상태에서 실패한다. 되돌린 뒤 sha256 기준 일치, 패키지 GREEN.

---

## 4. F2 재현과 수정 (§8.3)

### 4.1 재현 — 결함이 실재한다

**넣은 변이 둘** (스냅샷은 편집 전 `sha256sum` 기록 후 scratchpad에 사본을 떴다):

1. `docs/migration/openapi.latest.json:7938`에서 **`TOP_GAINERS` 불릿의 `— \`realtime\` 미지원`만** 제거.
   `TOP_LOSERS` 불릿은 그대로.
2. `Panel` 리터럴에 `RankingTopGainers` 복원.

```text
$ go test ./internal/candidatesrc/ -count=1 -run 'TestThePanelAsksForNothing…|TestTheSnapshotMatcher…' -v
=== RUN   TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
--- PASS: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused (0.00s)
=== RUN   TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor
--- PASS: TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor (0.00s)
ok  ...
```

**gainers + realtime 배선이 통과한다. 리뷰 F2가 맞다.** 금지 집합이 `{TOP_LOSERS}`로
비어 있지 않아 `len(forbidden) == 0` Fatal이 걸리지 않고, positive control은 자기
하드코딩 fixture를 읽으므로 실파일 표류를 볼 수 없다.

### 4.2 수정과 RED

`TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused` 안, 빈 집합 Fatal **바로 다음**에
실파일 결과에 대한 완전성 단언을 넣었다.

```go
for _, typ := range []string{"TOP_GAINERS", "TOP_LOSERS"} {
	if forbidden["realtime"][typ] {
		continue
	}
	t.Fatalf("%s no longer says %s refuses `realtime`, ...", snapshotPath, typ, ...)
}
```

같은 변이 상태에서:

```text
--- FAIL: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
    snapshot_drift_test.go:468: ../../docs/migration/openapi.latest.json no longer says
    TOP_GAINERS refuses `realtime`, and this guard is built on the fact that it does.
    ...
    Read for "realtime" out of the file: [TOP_LOSERS]
```

**RED ✔.**

**형태의 근거 네 가지를 주석에 적었다.**
① 선례는 `TestTheSourcesStillDeclareTheSizesTheseCommentsClaim`이 `{100, 30}`을 **사본이
아니라 실파일**에 대해 고정하는 것이다.
② `Errorf`가 아니라 `Fatal`이다 — 이후 모든 단언이 `forbidden`에서 계산되므로 불완전한
집합은 답 하나를 틀리게 하는 것이 아니라 **전부를 무증거로** 만든다. 같은 이유로 이 파일의
입력 신뢰성 검사(`rankingsTypeDescription`·`wiredRankings`·`wiredDurations`·빈 집합)는
전부 Fatal이다. 일관된다.
③ **정확한 집합이 아니라 부분집합**을 단언한다. `realtime`을 거부하는 타입이 새로
문서화되는 것은 이 가드가 **행동해야 할 정보**이지 실패할 일이 아니고, 교차 단언이 이미
그것을 처리한다. 막아야 하는 것은 하나가 사라지는 쪽뿐이다.
④ fixture positive control은 **그대로 두었다**(tasks가 허용한 대로). 둘은 다른 일을 한다 —
fixture는 matcher가 **과잉 매치하지 않음**까지 보고, 새 단언은 **파일이 여전히 그렇게
말하는지**를 본다. 어느 쪽도 다른 쪽을 대체하지 않으며, 그 분업을 주석에 적었다.

### 4.3 원상복구 대조

```text
$ cp <scratchpad 사본> docs/migration/openapi.latest.json
$ sha256sum docs/migration/openapi.latest.json
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49   ← 기준과 동일
$ git diff --stat docs/
 docs/pm/portfolio/_registry.yaml | 4 +++-      ← 이 change의 PM 등록뿐. 스냅샷 변경 없음
$ sha256sum internal/candidatesrc/candidatesrc.go
fe6daa185e8ccefdfce412f63e65c711f84e1ebe71fd1d2f0f3f044298da205c   ← 기준과 동일
```

334KB 공식 스냅샷은 **바이트 단위로 원본과 동일**하다.

---

## 5. 고친 세 곳 — before / after (§8.2)

### 5.1 `internal/candidatesrc/candidatesrc.go` — `Panel` 안 주석

**before**

```go
// The errors are discarded here because every type in this literal is a
// compile-time constant of this package with an entry in rankingSourceID —
// a failure would be a defect in this file, and
// TestEveryPanelSourceHasItsOwnID fails if one ever slips.
```

**after** (요지: 왜 버려도 되는지의 조건을 남기고, **거짓인 귀속을 사실로 바꾸고**, 실제
가드를 가리킨다. 무엇이 틀렸었는지도 남긴다 — 이 change의 주제가 그것이므로)

```go
// The errors are discarded here because every type in this literal is a
// compile-time constant of this package with an entry in rankingSourceID —
// a failure would be a defect in this file rather than a condition to handle.
//
// Discarding it is only defensible while something fails when it happens, and
// until this change nothing did. This comment used to name
// TestEveryPanelSourceHasItsOwnID, and that test does not catch it: it walks
// the panel it is handed and checks it for duplicate ids and emptiness, while a
// type with no rankingSourceID entry yields a panel that is one element shorter,
// still unique and still not empty. So the suite stayed green for a panel that
// read two lists while its author believed it read three — the failure shape
// this change exists to remove, sitting under the sentence that claimed to
// prevent it. The guard that does catch it is
// TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds in
// snapshot_drift_test.go, which compares the types this literal names against
// the sources Panel returns when it is called.
```

### 5.2 `analysis/function-logic/internal-candidatesrc--panel/function-logic-map.md`

**두 곳이었다** — 리뷰가 지목한 `:21`(입력 표)과 `:32`(B3 행).

`:21` **before**: *"매핑이 없는 타입은 `OfficialRanking`이 오류를 내고 B3가 그 원천을 버린다.
`TestEveryPanelSourceHasItsOwnID`와 `snapshot_drift_test.go`가 이 리터럴을 감시한다"*
→ 앞의 테스트는 **감시하지 않는다**(뒤의 파일은 감시하지만 스냅샷 금지 조합만 본다).

`:21` **after**: 버림이 **조용하다**는 것을 명시하고, 그 상태를 실패로 만드는 것이
`TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds` **하나뿐**이라고 적었으며,
스냅샷 가드가 무엇만 보는지 구분했다.

`:32`(B3 Required test) **before**: `TestAnUnknownRankingTypeIsRefused`(생성자 쪽),
`TestEveryPanelSourceHasItsOwnID`
`:32` **after**: `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds` — **이 분기를
덮는 유일한 테스트**임을 적고, 나머지 둘이 왜 덮지 않는지(전자는 생성자를 직접 부른다,
후자는 넘겨받은 패널만 본다)를 함께 적었다.

추가로 문서 머리에 **정정 블록**을 넣었다(날짜·사유·재현·무엇을 만들었는지). 그리고
줄 번호를 실제와 맞췄다 — 주석이 길어져 B2 592→604, B3 593→605, B4 600→612,
return 603→615, calls 594/601→606/613으로 이동했다.

### 5.3 `analysis/function-logic/internal-candidatesrc--panel/branch-test-map.md`

**before** (B3 행)

| Branch | Test | RED observed |
|---|---|---|
| B3 | `TestAnUnknownRankingTypeIsRefused` (생성자 쪽) · `TestEveryPanelSourceHasItsOwnID` | no |

**after**

- B3의 "Test" 칸을 `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`로 바꾸고,
  **RED observed를 `no` → `yes`**로 고쳤다(§3.1의 변이에서 실제로 관측했다).
  시나리오 문구도 *"조용히 빠진다 — 그리고 그 상태가 실패로 보고된다"*로 고쳤다.
- **B2′ 행을 새로 넣었다**: *"이 리터럴이 AST가 읽는 그 리터럴이다"*. §3.2의 변이와,
  **같은 변이에서 스냅샷 가드가 PASS였다**는 사실을 근거 칸에 적었다.
- 문서 머리에 정정 블록.

### 5.4 FLM 산출물 재생성

주석 편집으로 `Panel`의 source SHA가 바뀌어 `check_analysis.py`가 **먼저 거절했다**:

```text
[logic-map] internal-candidatesrc--panel: AST source hash is stale: internal/candidatesrc/candidatesrc.go
[logic-map] internal-candidatesrc--panel: AST hash does not match modified function revision
```

`go run ./tools/logic-map --file internal/candidatesrc/candidatesrc.go --func Panel`로
`ast.json`을 재생성했다. **분기 목록은 base와 동일하다**(B1 if / B2 range / B3 if / B4 if,
return 하나) — FLM이 주장하는 "분기 구조 무변경"은 재생성 후에도 성립한다. 이후:

```text
[logic-map] retire-gainers-source: evidence complete or diff-proven exempt
```

FLM 대상은 **3건 그대로**다(새 테스트를 기존 파일이 아닌 새 파일 `snapshot_drift_test.go`에
넣었으므로 늘지 않았다).

---

## 6. issues.md — F4~F9 이관 (§8.5)

| 리뷰 | issues.md | 상태 |
|---|---|---|
| F4 | I5 | **절반 닫힘.** 아래 참조 — 실측으로 확인하고 경계를 적었다 |
| F5 | I6 | 이관. 교차곱 오탐, 재검토 시점 |
| F6 | I7 | 이관. 가드가 보는 파일은 하나 |
| F7 | I8 | 이관. **세 번째 "결론은 맞고 근거는 틀림"** |
| F8 | I9 | 이관. 결합 방향, 지금 옮기지 않는 이유도 |
| F9 | I10 | 이관(정보) |

**F4는 닫혔는지 추측하지 않고 쟀다.**

- **① 눈이 머는 방향 — 닫혔다.** §3.2 실측. 새 가드가 RED, 스냅샷 가드는 PASS.
- **② 시끄러운 방향 — 남았다.** `Panel`에 `[]string{candidate.MarketKR}`를 넣어 확인했다:

```text
--- FAIL: TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
    snapshot_drift_test.go:205: candidatesrc.go: Panel builds a ranking from *ast.SelectorExpr
    rather than a named constant. ...
--- FAIL: TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
    snapshot_drift_test.go:205: (같은 메시지)
```

**둘 다 랭킹과 무관한 이유로 죽는다.** 새 가드도 `wiredRankings`를 재사용하므로 이 함정을
**물려받았다.** 거짓 통과가 아니라 거짓 실패라 랜딩을 막지 않지만, 그때 이것을 느슨하게
만드는 수리가 ①을 되살릴 수 있다는 점을 I5에 적었다. 변이는 되돌렸다(sha256 대조 완료).

**F7이 세 번째라는 사실을 표로 적었다**(I8) — proposal의 `NoteSources`, tasks §3.4의
"호출이 세 번에서 두 번으로", 그리고 D2의 "되읽혀야 한다". 셋 다 결론은 맞고 근거는 틀렸다.
§8.1이 닫은 F1도 같은 뿌리라는 것을 함께 적었다 — 거기서는 근거가 *"이 테스트가 잡는다"*
였고 그 테스트가 잡지 않았다.

---

## 7. 막힌 것 · tasks가 코드와 안 맞는 것

### 7.1 tasks §8.1의 *"이 한 단언이 F1과 F4를 같이 닫는다"*는 **과장이다**

**F4는 절반만 닫힌다.** §6의 실측이 근거다. F4의 서술 자체가 두 방향을 적고 있는데
(*"반대 방향(시끄러운 쪽)도 있다"*), tasks는 그것을 한 문장으로 합치면서 "같이 닫는다"고
썼다. 새 가드는 `wiredRankings`를 **재사용**하므로 그 함수의 인식 한계를 그대로 물려받고,
F4가 제안한 최소 수정(*"`range` 대상 식을 따라가거나 … 키 집합에 속하는지 확인"*)은
`wiredRankings` **자신**을 고치는 일이라 이 가드로 대체되지 않는다.

지금 상태는 안전한 방향(거짓 실패)이므로 **§8에서 고치지 않았다** — 그것은 §8이 받은
범위 밖이고, `wiredRankings`를 고치면 이 change의 핵심 산출물 동작을 바꾸는 편집이 된다.
I5에 경계와 고칠 형태를 적었다. **Manager 판단이 필요하다.**

### 7.2 tasks §8.1의 변이 지시는 **그대로 따르면 엉뚱한 RED가 난다**

*"`rankingSourceID`에 항목이 없는 실제 enum 값(`TOSS_SECURITIES_TRADING_AMOUNT`)을
`Panel` 리터럴에 넣고"* — 문자 그대로 읽어 문자열 리터럴
`[]string{..., "TOSS_SECURITIES_TRADING_AMOUNT"}`를 넣으면 `wiredRankings`가
*"Panel builds a ranking from \*ast.BasicLit rather than a named constant"*로 **Fatal**한다.
초록이 아니라 빨강이 나오고, 그것을 보면 **"결함이 없다"고 잘못 결론 내린다.**

재현하려면 **상수를 먼저 선언**해야 한다(리뷰어의 F1 서술은 그렇게 하고 있다 —
`RankingTossAmount = "TOSS_SECURITIES_TRADING_AMOUNT"`). tasks의 문장에는 그 단계가 없다.
차단 사항은 아니지만, 이 change에서 반복된 형태 — **결론은 맞고 절차가 한 칸 빠짐** — 이라
적어 둔다.

### 7.3 §8.6은 구현자 손에서 **PASS를 낼 수 없다** (결함이 아니라 구조)

`make gate`는 미완료 체크박스 0건을 요구하는데 §8.4(커밋 절차)와 §8.7(2차 독립 리뷰)은
구현자의 것이 아니다. §7.4에서 첫 Teammate가 만난 것과 **정확히 같은 구조**다.
tasks §7.6의 주석이 이미 *"gate PASS는 랜딩 승인이 아니다"*라고 적었으므로 모순은 아니지만,
`[T]`가 붙은 §8.6이 통과를 요구하는 형태로 쓰여 있는 것은 다음 change에서 고칠 값이 있다.

### 7.4 없었던 것 — 확인하고 적는다

- **§8.2의 "세 곳"은 실제로 세 파일 · 네 문장**이었다(`function-logic-map.md`가 `:21`과
  `:32` 둘). 리뷰 F1이 이미 넷을 열거했으므로 tasks의 "세 곳"은 파일 수를 센 것이다.
  누락 없이 전부 고쳤다.
- **`panelsize_drift_test.go` 무관 RED 없음.** 새 주석에 `숫자 + rows/행/개` 형태가 없다
  (숫자 자체가 없다). 전체 스위트 GREEN으로 확인.
- **금지 디렉터리 무변경**: `internal/risk` · `internal/execgw` · `internal/exitpolicy` ·
  `internal/trading` 어느 것도 `git status`에 없다.
- **후퇴 사다리·`duration`·패널 구성원 최종 상태 무변경**: `Backoff.Note` 무변경,
  `Read`의 `"realtime"` 무변경, `Panel` 리터럴은 `{RankingTradingAmount,
  RankingTradingVolume}`로 §8 이전과 동일(변이는 전부 넣었다 되돌렸다).
- **주문 명령 미실행**: `tossctl`을 한 번도 실행하지 않았다. `go test`·`go vet`·`gofmt`·
  `make` 타깃과 정적 분석뿐이다. LIVE 주문 side effect 없음.

---

## 8. 재검증 — 실행한 명령과 실제 결과 (§8.6)

| 명령 | 결과 |
|---|---|
| `go test ./...` | **3686 passed in 57 packages** (실패 0). 리뷰 시점 3685에서 **+1** — 새 가드 하나 |
| `go vet ./...` | **No issues found** |
| `gofmt -l .` | **출력 없음** (`$(go env GOROOT)/bin/gofmt`) |
| `go test -race ./internal/candidate/... ./internal/candidatesrc/... ./cmd/tossctl/...` | **663 passed in 3 packages** (리뷰 시점 662에서 +1) |
| `python3 tools/logic-map/check_analysis.py --change retire-gainers-source` | `evidence complete or diff-proven exempt` |

**upstream 상속 테스트 회귀 없음.** 줄어든 것은 없고 늘어난 하나가 이 §8이 더한 가드다.

### `make sdd-sync` → `make sdd-check` → `make gate` (연속 실행)

**두 번 돌렸고, 두 번째가 최종이다.** 첫 실행은 `.go` 편집을 끝낸 뒤였는데, 그 다음
`issues.md`·FLM 산출물·이 파일을 편집하자 `make sdd-check`가 다시 막았다:

```text
[index-freshness] FAIL: CodeGraph hard-evidence index is missing or stale; run `make sdd-sync`
```

**fingerprint는 `.go`만 세지 않는다** — tasks §7.6의 주의 문구가 *".go 편집이 있으면"*이라고
쓰고 있지만 실제로는 `.md` 편집에도 stale이 된다. 편집을 전부 끝낸 뒤 세 명령을 **다시**
연속 실행했고, 아래는 그 최종 실행 결과다. (부수 관측: `sdd-check`가
`[context-graph] CCG adapter failed; its checkpoint was not advanced`를 냈다. 외부 그래프는
`docs/WORKFLOW.md`가 정한 대로 **관측 전용·비차단**이고 `sdd-check`는 exit 0이다.)

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph `Already up to date` · CodeGraphContext · GBrain) |
| `make sdd-check` | **통과** — `[agent-config] … synchronized` · memory `valid` · `[index-freshness] CodeGraph hard-evidence index matches the worktree` · `[pm] hierarchy and generated trackers are current` · sdd-test 전부 통과 |
| `make gate CHANGE=retire-gainers-source` | **FAIL — 2/8단계, 미완료 태스크 2건** |

**게이트 실패를 사실 그대로 적는다.**

```text
==> 1/8 tasks.md 확인
OK: openspec/changes/retire-gainers-source/tasks.md

==> 2/8 미완료 태스크 확인
미완료 태스크 2 건:
228:- [ ] 8.4 **F3 — 커밋 절차.** …
245:- [ ] 8.7 §8 수정분에 대한 **2차 독립 리뷰**. …
GATE FAIL: retire-gainers-source — 미완료 태스크 2 건이 남아 있습니다
```

**실패 지점은 2단계 하나이고, 원인은 구현자가 할 수 없는 두 task다.** 나머지 단계를 각각
손으로 돌려 확인했다:

```text
3/8 review.md 존재                    OK (이 파일)
4/8 Function Logic Map                OK — evidence complete or diff-proven exempt
5/8 make sdd-check                    OK
6/8 make test                         OK — FAIL 0건
7/8 make vet                          OK
8/8 make validate                     OK — Totals: 25 passed, 0 failed (25 items)
```

§8.1·8.2·8.3·8.5·8.6 체크박스는 **실제로 수행했기 때문에** 채웠다. 게이트가 통과했다는
뜻이 아니다 — 통과하지 못했고 그 이유는 위의 둘이다.

---

## 9. 변경 요약 (§8이 만든 diff)

수정:

```text
internal/candidatesrc/candidatesrc.go   Panel 주석 (§8.2). 배선·리터럴 무변경
```

새 파일이 아닌 신규 내용:

```text
internal/candidatesrc/snapshot_drift_test.go
  + 파일 헤더에 "두 번째 가드가 왜 이 파일에 있는가" 절
  + panelRankingTypes 헬퍼 (§8.1)
  + TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds (§8.1)
  + TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused 안 실파일 완전성 단언 (§8.3)
```

문서:

```text
openspec/changes/retire-gainers-source/issues.md          + I5~I10 (§8.5)
.../analysis/function-logic/internal-candidatesrc--panel/
    ast.json              재생성 (source SHA)
    function-logic-map.md 정정 블록 + :21 · :32 + 줄 번호
    branch-test-map.md    정정 블록 + B3 행 + B2′ 행 신설
openspec/changes/retire-gainers-source/tasks.md           8.1·8.2·8.3·8.5·8.6 체크
openspec/changes/retire-gainers-source/review.md          이 절
```

`git diff --stat` (작업 트리 전체, §8 이전 7파일 108+/22- → 현재):

```text
 cmd/tossctl/candidate.go                   | 15 +++++++-
 docs/pm/portfolio/_registry.yaml           |  4 ++-
 internal/candidate/candidate.go            |  9 +++++
 internal/candidate/notdue_test.go          |  8 ++---
 internal/candidatesrc/candidatesrc.go      | 56 ++++++++++++++++++++++++++----
 internal/candidatesrc/clock_wiring_test.go | 48 +++++++++++++++++++------
 tools/pm/test_generate_master_tracker.py   |  2 ++
 7 files changed, 120 insertions(+), 22 deletions(-)
```

**파일 목록은 §8 이전과 동일하다** — §8은 새 tracked 파일을 만들지 않았고
`candidatesrc.go`의 주석만 늘었다(44 → 56).

---

## 10. 남은 것

| # | 항목 | 누구 |
|---|---|---|
| **8.4** | F3 커밋 절차 — `_registry.yaml`·PM fixture에서 `refine-extended-shadow-bands`를 커밋 시점에만 빼기 | **Manager** |
| **8.7** | §8 diff에 대한 2차 독립 리뷰. 새 가드에 리뷰어가 **직접** 변이를 넣어 확인 | **독립 리뷰어** |
| 7.1 | *"F1과 F4를 같이 닫는다"*는 과장 — F4 ②는 남았다(I5). `wiredRankings`를 고칠지 판단 | **Manager** |


---

# 2차 독립 리뷰 (§8.7)

- 날짜: 2026-07-29
- 역할: **2차 독립 리뷰어.** 구현과도, 1차 리뷰와도 분리된 컨텍스트다. 아래 RED/GREEN은
  전부 **리뷰어가 직접 변이를 넣고, 돌리고, 되돌린** 결과이며 위 두 절의 인용이 아니다.
- 범위: **§8이 만든 diff.** §1~§7의 재검증은 1차 리뷰가 이미 했으므로 반복하지 않았다.
  §8이 §1~§7의 결론을 무효화했는지만 따로 확인했다(→ 무효화 없음, §5.4).
- 변이 정책: `git checkout --`·`git restore`·`git stash`를 **한 번도 쓰지 않았다.**
  모든 변이는 편집으로 넣고 편집으로 되돌렸다(스냅샷도 되돌림 자체는 편집이고,
  scratchpad 사본과는 `cmd`로 **추가** 대조만 했다). 커밋·push 없음.
  `internal/risk`·`internal/execgw`·`internal/exitpolicy`·`internal/trading` 무변경.
  `tossctl` 미실행 — LIVE 주문 side effect 없음.

## 기준 해시 (리뷰 시작 시점 = §8 완료 상태)

```text
cefe12f9a05aa4782a9aaf9c69ee56e16356d601498864bad7d356ca3bf268a5  internal/candidatesrc/candidatesrc.go
12ef78f6bd995daea7517b50ff770e8549f889d94becf766f5632068d26ad4d5  internal/candidatesrc/snapshot_drift_test.go
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49  docs/migration/openapi.latest.json
```

**리뷰 종료 시 세 해시 모두 위와 동일하다**(각 변이 직후 대조했고, 마지막에 다시 대조했다).

---

## 1. 새 가드는 진짜인가 — 변이 6종

### 1.1 변이 A — `rankingSourceID`에 없는 실제 enum 값 → **RED** ✔

지시대로 **먼저 상수를 선언**하고 그 식별자를 리터럴에 넣었다(맨 문자열을 넣으면
`wiredRankings`가 `*ast.BasicLit`에서 Fatal하는 틀린 RED가 난다).

```go
const ( … RankingTossAmountMUT = "TOSS_SECURITIES_TRADING_AMOUNT" )
for _, typ := range []string{RankingTradingAmount, RankingTradingVolume, RankingTossAmountMUT} {
```

```text
$ go test ./internal/candidatesrc/ -run TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds -v
    snapshot_drift_test.go:301: candidatesrc.go names TOSS_SECURITIES_TRADING_AMOUNT in Panel's
    ranking literal and Panel builds no source for it.
    Named in the literal: [MARKET_TRADING_AMOUNT MARKET_TRADING_VOLUME TOSS_SECURITIES_TRADING_AMOUNT]
    Built by Panel: [MARKET_TRADING_AMOUNT MARKET_TRADING_VOLUME]
--- FAIL: TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds (0.00s)
```

### 1.2 변이 A′ — **같은 변이에서 가드만 지우면** → 패키지 전체 GREEN ✔ (결함 실재 확인)

가드 함수명을 `Test…` 밖으로 바꿔(= 실행에서 제거) 변이 A를 그대로 둔 채 돌렸다.

```text
$ go test ./internal/candidatesrc/            → 53 passed (실패 0)
$ go test ./internal/candidate/... ./cmd/tossctl/...  → 609 passed (실패 0)
```

**53건은 54(현재) − 1(새 가드)이다.** 구현 기록의 *"패키지 53건 전부 통과"*와
`branch-test-map.md` B3의 *"가드 도입 전에는 패키지 53건 전부 통과"*가 정확하다.
그리고 이것이 `function-logic-map.md:32`의 *"그 상태를 실패로 만드는 것은 …
하나뿐이다"*에 대한 직접 증거다 — 나머지 53건 중 어느 것도 잡지 못한다.
가드명을 되돌린 뒤 sha256 기준 일치.

### 1.3 변이 B — 리터럴을 패키지 수준 `var`로 (구현자가 B2′에 적은 그 형태) → **RED**, 스냅샷 가드 **PASS** ✔

```go
const ( … mutBUnrelated = "SOMETHING_ELSE" )
var mutBRankings = []string{RankingTradingAmount, RankingTradingVolume}
…
	_ = []string{mutBUnrelated}
	for _, typ := range mutBRankings {
```

```text
$ go test ./internal/candidatesrc/ -run 'TestEveryRankingType…|TestThePanelAsksForNothing…' -v
→ 1 passed, 1 failed
  [FAIL] TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
     :301 candidatesrc.go names SOMETHING_ELSE …
     :317 Panel builds a source for MARKET_TRADING_AMOUNT and no ranking literal in Panel names it.
     :317 Panel builds a source for MARKET_TRADING_VOLUME and no ranking literal in Panel names it.
  [PASS] TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
```

**`branch-test-map.md` B2′ 행의 주장은 사실이다.** 구현자가 적은 그대로 재현된다.

### 1.4 변이 B-sharp — **같은 리팩터에 gainers를 실제로 되살려서** → 진짜 위반이 통과하는지

B2′는 "무관한 타입"으로 눈멂을 보인다. 나는 **실제 금지 조합**을 얹어 F4 ①의 위험을
끝까지 밀었다: `var mutBRankings = []string{…, RankingTopGainers}`.

```text
$ go test ./internal/candidatesrc/     → 46 passed, 5 failed
  [FAIL] TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
     Named in the literal: [SOMETHING_ELSE]
     Built by Panel: [MARKET_TRADING_AMOUNT MARKET_TRADING_VOLUME TOP_GAINERS]
     :317 Panel builds a source for TOP_GAINERS and no ranking literal in Panel names it.
  [FAIL] TestThePanelHandsItsClockToEverySourceItBuilds
  [FAIL] TestNoMarketPanelBuildsTheGainersRanking/KR, /US
  [PASS] TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused   ← 여전히 눈이 멀어 있다
```

**결론**: 스냅샷 가드는 이 상태에서 **진짜 금지 조합을 보고도 통과한다.** 그 상태를 실패로
바꾸는 것이 새 가드이고, 그것이 F4 ①이 닫혔다는 말의 실제 내용이다. `built`에
`TOP_GAINERS`가 실제로 담기고 이름이 지목된다는 것까지 확인했다.

### 1.5 공허하게 참인 방향 — 둘 다 직접 확인 ✔

| 방향 | 만든 상태 | 결과 |
|---|---|---|
| `built`가 빈다 | `rankingSourceID`를 빈 map으로 | `:301`에서 declared의 **두 타입 모두** Errorf. `Built by Panel: []` |
| `declared`가 빈다 | 리터럴을 패키지 수준 `var`로 올리고 `Panel`에 다른 `[]string`을 **남기지 않음** | `wiredRankings`의 `:294`/`:480` Fatal — **두 가드가 함께** *"no ranking type was read out of candidatesrc.go's Panel"* |

테스트 doc의 *"Neither direction can pass vacuously"*는 실측으로 성립한다.

### 1.6 역매핑이 단사가 아니면 — **Fatal** ✔

`rankingSourceID[RankingTopGainers]`를 `SourceOfficialTradingValue`로 바꿨다.

```text
  [FAIL] TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds
     snapshot_drift_test.go:295: rankingSourceID gives MARKET_TRADING_AMOUNT and TOP_GAINERS the same source id "official_rankings_trading_value" …
  [FAIL] TestTheRetiredSourceKeepsItsIdentity   (두 번째 그물이 있다)
```

구현자의 주장대로 **Fatal이며**, 가드가 자기 입력의 결함을 결론으로 바꾸지 않는다.

### 1.7 "두 가드가 함께 눈머는 조합" — 찾았고, **없다**

두 가드를 합치면 이렇게 된다: 새 가드 GREEN ⟹ `declared`(AST) = `built`(런타임).
그러므로 `Panel`이 금지 조합 T를 만들면 `built ∋ T ⟹ declared ∋ T`이고, 스냅샷 가드의
교차 단언이 반드시 T를 잡는다. **거짓 통과가 성립하려면 `wiredRankings`가 조용히
"틀린 비어 있지 않은 집합"을 돌려줘야 하는데**, 그 함수는 `*ast.Ident`가 아닌 원소와
이 파일에 없는 상수에서 전부 Fatal한다. 남은 경우(`[][]string`·타입 생략 리터럴·
`map[string][]string` 값 등 `lit.Type`이 `*ast.ArrayType`이 아닌 형태)는 그 리터럴이
**건너뛰어지므로** `declared`가 비거나 실제와 어긋나고, 둘 다 Fatal 또는 RED다.
`wiredDurations` 쪽도 같다 — 리터럴이 아니면 Fatal, 하나도 없으면 Fatal.
**즉 §8 이후 이 두 가드의 조합에는 false pass 경로가 없다.** 남은 것은 전부 false
**failure** 방향이고 그것이 I5 ②다.

---

## 2. F2 수정은 진짜인가

### 2.1 편집 전 기록

```text
$ sha256sum docs/migration/openapi.latest.json
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49
$ cp docs/migration/openapi.latest.json <scratchpad>/openapi.latest.json.orig
```

### 2.2 변이 — `TOP_GAINERS` 불릿의 `— \`realtime\` 미지원`**만** 제거 (`TOP_LOSERS` 유지)

치환 전 출현 수를 세어 **정확히 1건**임을 확인하고 1건만 바꿨다. 변이 후 실제 파싱 결과:

```text
- `TOP_GAINERS`: 급상승 (등락률 상위)                      ← 절이 사라졌다
- `TOP_LOSERS`: 급하락 (등락률 하위) — `realtime` 미지원   ← 그대로
```

그 상태에서 `Panel`에 `RankingTopGainers`를 되돌렸다.

### 2.3 결과 — **Fatal** ✔

```text
$ go test ./internal/candidatesrc/ -run 'TestThePanelAsksForNothing…|TestTheSnapshotMatcher…' -v
→ 1 passed, 1 failed
  [FAIL] TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused
     snapshot_drift_test.go:468: ../../docs/migration/openapi.latest.json no longer says
     TOP_GAINERS refuses `realtime`, and this guard is built on the fact that it does.
  [PASS] TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor   ← fixture는 그대로 초록
```

1차 리뷰가 **통과**를 재현한 바로 그 상태에서 지금은 Fatal이다. **F2는 닫혔다.**
positive control이 여전히 초록이라는 사실이 "fixture는 실파일 표류를 볼 수 없다"는
F2의 진단이 옳았음을 같은 실행에서 보여준다.

### 2.4 원상복구 대조

```text
$ sha256sum docs/migration/openapi.latest.json
02cfb91237aa0e2fe39a6776c1147afb8b323e0d3ff51ce57a86cb9675d4eb49   ← 기준과 동일
$ cmp docs/migration/openapi.latest.json <scratchpad>/openapi.latest.json.orig
(무출력 — 바이트 동일)
$ git diff --stat docs/
 docs/pm/portfolio/_registry.yaml | 4 +++-      ← _registry.yaml 하나뿐
```

334KB 공식 스냅샷은 **바이트 단위로 원본이다.**

### 2.5 subset(`⊇`)으로 단언한 판단 — **옳다**

정확 일치를 요구하지 않은 것이 맞다. 근거는 두 방향을 다 재봤기 때문이다.

- **새 타입이 `realtime`을 거부하도록 문서화되는 경우**: 집합이 커진다. 정확 일치면
  **위반이 아닌 것에 실패**하고, 그 실패를 만나는 사람은 배선을 안 건드린 사람이다 —
  가드를 느슨하게 만들 동기가 생기는 전형적 자리다. subset이면 조용하고, 그 타입이
  실제로 배선되면 교차 단언이 잡는다.
- **matcher가 과잉 매치하는 경우**: 집합이 커진다. 교차 단언이 살아 있는 배선
  (`MARKET_TRADING_AMOUNT`·`MARKET_TRADING_VOLUME`)에 대해 즉시 RED다 — 시끄러운
  안전 방향이고, fixture control의 음성 단언이 한 겹 더 있다.
- **하나가 사라지는 경우**: subset 단언이 Fatal. 이것이 F2가 재현한 그 방향이다.

세 방향 모두 원하는 결과가 나오므로 **정확 일치로 했어야 할 이유가 없다.**
추가로 재워딩 내성도 확인했다: `미지원: \`realtime\``이나 `\`duration=realtime\``처럼
정규식이 못 읽는 형태로 바뀌면 집합이 줄어 **이 단언이 Fatal한다**(빈 집합 Fatal이 아니라
이 단언이 잡는다). 그것이 F2가 열어 두었던 정확히 그 구멍이다.

---

## 3. FLM 산출물의 정정은 정확한가

### 3.1 `branch-test-map.md`의 "RED observed"를 **직접 재현**

| 행 | 주장 | 리뷰어 재현 |
|---|---|---|
| **B2′** | 목록을 패키지 수준 `var`로 올리고 무관한 `[]string`을 남기면 RED, **같은 변이에서 스냅샷 가드는 PASS** | §1.3 — **둘 다 그대로 관측**. 그럴듯한 서술이 아니라 실제 관측이다 |
| **B3** | 매핑 없는 실제 enum 값을 리터럴에 넣으면 RED. 가드 **도입 전에는 패키지 53건 전부 통과** | §1.1(RED) + §1.2(가드 제거 시 53 passed). **두 절 모두 그대로 관측** |

B1·B4의 `RED observed: no`도 정직하다 — 이 change가 바꾼 조건이 아니다.

### 3.2 `function-logic-map.md`의 두 정정은 현재 코드와 맞는가 — ✔

- **`:32`(입력 표, 랭킹 타입 리터럴 행)**: *"그 상태를 실패로 만드는 것은
  `TestEveryRankingType…` 하나뿐"* → §1.2가 직접 증거다(가드 제거 시 나머지 53건 GREEN).
- **`:43`(B3 행)**: *"이 분기를 덮는 유일한 테스트"* + 나머지 둘이 왜 안 덮는지.
  `OfficialRanking`의 오류 경로를 직접 읽어 확인했다 — 오류 반환은 **정확히 하나**,
  `rankingSourceID`에 없는 타입뿐이다([candidatesrc.go:231-234]). 그러므로 B3의 false
  분기 전체가 새 가드의 첫 루프에 대응하고, 서술에 남은 구멍이 없다.
- **줄 번호**: 문서가 적은 B1 555 / B2 604 / B3 605 / B4 612 / return 615 /
  calls 552·605·606·613을 소스에서 직접 grep해 **전부 일치**를 확인했다.

### 3.3 `ast.json` — `check_analysis.py` 통과 **말고** 분기 구조를 직접 대조 ✔

base commit의 파일을 꺼내 같은 도구로 AST를 새로 뽑아 HEAD와 비교했다.

```text
$ git show b2c261a…:internal/candidatesrc/candidatesrc.go > <scratchpad>/candidatesrc.base.go
$ go run ./tools/logic-map --file <scratchpad>/candidatesrc.base.go --func Panel
BASE branches: [('B1','if'), ('B2','range'), ('B3','if'), ('B4','if')]
BASE returns : 1
BASE calls   : ['strings.ToUpper','strings.TrimSpace','OfficialRanking','append','append','WTSPopular']

HEAD(ast.json) branches: [('B1','if'), ('B2','range'), ('B3','if'), ('B4','if')]
HEAD returns : 1
HEAD calls   : ['strings.ToUpper','strings.TrimSpace','OfficialRanking','append','append','WTSPopular']
HEAD mutations: None
```

**base와 완전히 동일하다.** FLM의 *"분기 구조 무변경"*은 `check_analysis` 통과와 무관하게
독립적으로 참이다. 세 FLM 디렉터리의 `ast.json` `source_sha256`도 현재 파일 해시와
전부 일치함을 따로 계산해 확인했다(`OK` 3/3).

---

## 4. `candidatesrc.go`의 `Panel` 주석

새 문장은 `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds`를 가리킨다.
**그 가드가 실제로 그 분기를 덮는다** — §1.1이 RED, §1.2가 "그것 말고는 아무도 안 잡는다".
F1의 실수(지키지 않는 가드에 귀속)는 **반복되지 않았다.**

두 가지를 더 확인했다.

1. 주석은 *"every type in this literal is a compile-time constant … with an entry in
   `rankingSourceID` — a failure would be a defect in this file"*라고 오류를 버리는 조건을
   말한다. `OfficialRanking`의 오류 경로가 정말 그 하나뿐인지 읽었다 — 그렇다. 조건을
   좁게 적어 놓고 실제로는 다른 오류도 버려지는, F1과 같은 형태의 구멍이 **없다.**
2. 저장소 전체에서 `TestEveryPanelSourceHasItsOwnID`에 대한 **남은 거짓 귀속**을 훑었다.
   이 change의 산출물·코드 안에는 없다. 남은 언급은 (a) 그 테스트가 무엇을 안 잡는지
   **설명하는** 새 문장 두 곳, (b) 그 테스트 자신의 정의, (c) 다른 change
   (`fix-chase-veto-measurement`)의 아카이브 전 산출물 — (c)는 이 change의 범위 밖이다.

---

## 5. 스스로 찾은 것

### 5.1 (P2) `function-logic-map.md:69` — 이 change가 없애기로 한 바로 그 형태의 원천 개수 주장이 게이트 산출물 안에 남아 있다

**파일:줄**: `analysis/function-logic/internal-candidatesrc--panel/function-logic-map.md:69`

> *"남는 구성은 KR 세 원천·US 두 원천이고, '공식 원천만으로 후보가 산출된다'는 요구사항은
> 계속 성립한다."*

**측정(리뷰어가 `Panel`을 세 구성으로 직접 호출)**:

```text
KR with WTS: 3      KR no WTS: 2      US: 2
```

**"KR 세 원천"은 조건부 사실을 무조건으로 적은 것이고, WTS 없는 KR에서 거짓이다.**
그 구성은 가정이 아니라 지원 구성이다 — tasks §7.3 자신이 *"WTS 있는 KR은 3, WTS 없는
KR은 2, US는 2 … **숫자를 고정해서 읽지 않는다**"*라고 쓰고 있다.

**실패 시나리오**: WTS 세션 없이 KR을 도는 배포에서 원천이 하나 줄었는지 판단하려는
사람이 이 문서를 기준선으로 읽는다 → 정상(2)을 "하나 잃었다"로 읽거나, 그 반대로 읽는다.
그리고 이 주장에는 가드가 없다 — issues.md **I2가 정확히 그 사실을 적어 둔 항목**이고,
§2.2와 §6.1은 같은 형태의 문장을 코드 주석에서 **두 곳 지운** task다.

**성격**: 이 문장은 §8이 만든 것이 아니라 §1~§7의 것이고 1차 리뷰가 놓쳤다. 그러나
§8.2의 임무가 *"이 파일의 주장을 사실에 맞게 고친다"*였고 구현자는 이 파일을 열어
네 곳을 고쳤다. **랜딩을 막지 않는다**(문서, 안전 경로 아님). 커밋 전에 한 문장을
개수 없는 표현으로 바꾸는 것이 값이 싸다. 바꾸지 않는다면 I2에 이 줄을 덧붙이는 것으로
족하다.

### 5.2 (P2) `snapshot_drift_test.go:52-59` — "두 방향이 **둘 다 조용히 통과한다**"는 서술 중 하나는 조용하지 않다

**파일:줄**: `internal/candidatesrc/snapshot_drift_test.go:52-59`

> *"it fails in two directions that both pass silently. … A refactor that lifts the
> literal out of Panel leaves the reader looking at whatever other `[]string` is in
> scope, **or at nothing at all**."*
> *"TestEveryRankingType… **closes both**"*

**측정**: 리터럴을 `Panel` 밖으로 올리고 **다른 `[]string`을 남기지 않으면**, 통과하지
않는다 — `wiredRankings`의 기존 `len(out) == 0` Fatal이 **두 가드를 함께** 죽인다
(§1.5 하단, `:294`/`:480`). 그 경우를 닫는 것은 새 가드가 아니라 **§8 이전부터 있던
그 Fatal**이다. 새 가드가 실제로 더한 것은 *"다른 `[]string`이 남은"* 쪽 하나다.

**실패 시나리오**: 결론(가드가 필요하다)은 맞고 근거(그 경우도 조용히 통과한다,
새 가드가 그것도 닫는다)가 틀렸다. 다음 사람이 `wiredRankings`의 빈 집합 Fatal을
"새 가드가 덮으니 중복"이라 보고 지우면, **§1.5의 두 번째 행이 열린다** — `declared`가
비고, 새 가드는 첫 루프를 한 번도 돌지 않아 **공허하게 통과**하며, 스냅샷 가드는 빈
`wired` 위에서 교차곱을 돌아 역시 통과한다. 두 가드가 함께 눈머는 유일한 경로가
**정확히 이 문장이 초대하는 편집**이다.

**이 change에서 네 번째 "결론은 맞고 근거는 틀림"이다**(I8의 표에 셋이 있다:
`NoteSources`, tasks §3.4, D2의 "되읽혀야 한다"). 그리고 F1이 그 뿌리였다.
**랜딩을 막지 않는다** — 지금 코드는 옳다. 고칠 것은 한 문장이고, 고치지 않는다면
`wiredRankings`의 `len(out) == 0` Fatal 위에 *"이 Fatal은 새 가드가 대체하지 않는다"*를
적는 편이 더 안전하다.

### 5.3 (P2, 경계 기록) `panelRankingTypes`의 **두 시장 합집합**이 무엇을 숨기는지는 적혀 있지 않다

**파일:줄**: `internal/candidatesrc/snapshot_drift_test.go:230-233`

주석은 합집합을 **왜** 쓰는지(시장별 멤버십을 금지하지 않으려고)는 적지만, 그 대가로
**무엇을 못 보게 되는지**는 적지 않는다. 이 파일의 다른 모든 한계는 I5~I7에 경계가 있다.

**측정**: 랭킹 루프를 `market == candidate.MarketKR`로 게이트해 **US 패널을 통째로 비웠다.**

```text
[PASS] TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds   ← 합집합이라 초록
[FAIL] TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking  "the US panel is empty"
[FAIL] TestEveryPanelSourceHasItsOwnID                          "US panel is empty"
[FAIL] TestNoMarketPanelBuildsTheGainersRanking/US              "the US panel is empty; an assertion
                                                                 about what it does not contain says…"
[FAIL] cmd/tossctl TestNoIntervalNamesASourceNoPanelBuilds      "the US panel is empty with every reader supplied"
```

**결함이 아니다** — 네 개의 독립적인 비어있음 가드가 잡는다. 기록하는 이유는 하나:
새 가드는 *"한 시장에서만 원천이 조용히 빠지는"* 상태를 **구조적으로 볼 수 없고**,
지금 그것이 도달 불가능한 이유는 `OfficialRanking`이 market을 인자로 받지 않기
때문뿐이다([candidatesrc.go:229]). 그 시그니처가 바뀌는 change가 이 한계를 되살린다.
issues.md에 한 줄 값이 있다(랜딩 차단 아님).

### 5.4 §8이 §1~§7의 결론을 무효화했는가 — **아니다** (확인 목록)

| 확인한 것 | 결과 |
|---|---|
| 패널 리터럴 최종 상태 | `{RankingTradingAmount, RankingTradingVolume}` — §8 이전과 동일 |
| `Read`의 duration | `"realtime"` 무변경 |
| `rankingSourceID` 세 항목 | 무변경(D2 유지) |
| 후퇴 사다리·`scan.go`·`watch.go` | `git status`에 없음 |
| §3.4 (`panelsize_drift_test.go`) | 계속 GREEN — 새 주석에 `숫자+rows/행/개` 형태 없음 |
| §7.3 장중 실측의 전제 | 배선 무변경이므로 유효 |
| FLM 대상 수 | 3건 그대로. 새 가드는 새 파일(`snapshot_drift_test.go`)에 있어 늘지 않았다 |
| `isolation_test.go` | 새 import(`internal/candidate`)는 `candidatesrc` **테스트** 파일이고, 그 검사는 `internal/candidate`에서 바깥으로 걷는다 — 새 엣지 없음 |
| 공유 헬퍼 위치 | `fakeRankings`·`fakePopular`는 `candidatesrc_test.go`(오래 사는 파일)에 있다 — I9(F8)와 같은 결합을 **새로 만들지 않았다** |
| `_registry.yaml`·PM fixture | F3 미해결 상태 그대로 유지(§8.4가 의도한 대로) |

### 5.5 (P2, 형식) issues.md 자신의 규칙을 새 항목이 지키지 않는다

`issues.md:4`가 *"각 항목은 후속 change의 시점과 등급을 함께 적는다"*라고 선언한다.
I1은 둘 다 있다(High-risk / 시점). **I5·I6·I8은 등급이 없다** — I8은 *"등급 판단이 먼저
필요하다"*라고 명시적으로 미루므로 정직하고, I5·I6은 그냥 없다. I9는 시점만 있다.
사소하지만 이 change의 주제가 "기록이 사실인가"이므로 적는다. 랜딩 차단 아님.

### 5.6 나트 (P2 미만, 판정에 영향 없음)

- `snapshot_drift_test.go:288` *"the snapshot guard **above** it"* — 스냅샷 가드는 이
  테스트보다 130줄 **아래**에 있다. 같은 문장의 파일 헤더 쪽 표현
  (*"`wiredRankings` — the function directly above it"*)도 실제로는
  `panelRankingTypes`가 바로 위다.
- `tasks.md`의 `[candidatesrc.go:560]`·`[556-559]`·`[591-594]`·`[566-570]` 등 줄 참조는
  §8의 주석 확장으로 약 12줄씩 밀렸다. tasks는 역사 기록이므로 고칠 필요는 없지만,
  줄 번호로 찾는 사람은 헛다리를 짚는다.

### 5.7 찾아봤고 **없었던** 것

발견 없음이 정보가 되도록 무엇을 훑었는지 적는다.

- **공허하게 참인 새 가드**: 두 방향 모두 실제로 만들어 확인(§1.5). 없다.
- **두 가드가 함께 눈머는 조합**: `wiredRankings`가 조용히 틀린 비어있지 않은 집합을
  돌려줄 수 있는 리터럴 형태를 열거해 확인(§1.7). 지금 코드에는 없다.
- **역매핑 비단사**: Fatal 확인(§1.6). 두 번째 그물(`TestTheRetiredSourceKeepsItsIdentity`)도 있다.
- **F2 수정이 새 false failure를 만드는가**: subset 단언의 세 방향을 다 재봤다(§2.5). 없다.
- **`OfficialRanking`의 숨은 오류 경로**: 하나뿐이다. B3 서술에 구멍 없다(§4).
- **`wiredDurations`의 우회**: 리터럴 아님 → Fatal, 없음 → Fatal. 조용한 우회 없다.
- **I5~I10이 F4~F9를 약화시켰는가**: 여섯 항목을 한 줄씩 대조했다. **약화 없다** —
  I5는 실측 두 건을 **더했고**, I8은 "세 번째"라는 패턴 표를 더했다. I5가 F4의 제안
  *"`rankingSourceID` 키 집합에 속하는 값만 담는지 확인"*을 *"`range` 절에 있는 것만
  본다"*로 **바꾼 것**은 확인했고, 바꾼 쪽이 이 파일의 감시 대상("Panel이 순회하는
  목록")에 더 맞다고 판단한다. 약화가 아니라 정정이다.
- **금지 디렉터리**: `internal/risk`·`internal/execgw`·`internal/exitpolicy`·
  `internal/trading` — `git status`에 하나도 없다.
- **주문 side effect**: `tossctl` 미실행. 실행한 것은 `go test`·`go vet`·`gofmt`·
  `go run ./tools/logic-map`·`make` 타깃·`python3` 정적 스크립트뿐이다.

---

## 6. 재검증 — 리뷰어가 직접 실행한 명령과 결과

| 명령 | 결과 |
|---|---|
| `go test ./...` | **3686 passed in 57 packages** (실패 0) |
| `go vet ./...` | **No issues found** |
| `gofmt -l .` | 출력 없음 |
| `go test -race ./internal/candidate/... ./internal/candidatesrc/... ./cmd/tossctl/...` | **663 passed in 3 packages** |
| `python3 tools/logic-map/check_analysis.py --change retire-gainers-source` | `evidence complete or diff-proven exempt` |
| `git diff --stat` | 7 files, **120 insertions / 22 deletions** — §8 기록과 일치 |
| 기준 해시 3종 재대조 | **전부 동일** (candidatesrc.go / snapshot_drift_test.go / openapi.latest.json) |

### `make sdd-sync` → `make sdd-check` → `make gate` (리뷰어가 연속 실행)

기록(이 절)을 다 쓴 뒤 세 명령을 연속으로 다시 돌렸다 — `.md` 편집도 fingerprint를
무효화하기 때문이다(tasks §7.6 정정 참조).

| 명령 | 결과 |
|---|---|
| `make sdd-sync` | `[sdd-sync] all indexes current` (CodeGraph `Already up to date` · CodeGraphContext · GBrain) |
| `make sdd-check` | **통과** — `[agent-config] … synchronized` · `[index-freshness] CodeGraph hard-evidence index matches the worktree` · `[pm] hierarchy and generated trackers are current` · sdd-test 4묶음 전부 OK |
| `make gate CHANGE=retire-gainers-source` | **FAIL — 2/8단계, 미완료 태스크 2건** |

```text
==> 2/8 미완료 태스크 확인
미완료 태스크 2 건:
251:- [ ] 8.4 **F3 — 커밋 절차.** …
268:- [ ] 8.7 §8 수정분에 대한 **2차 독립 리뷰**. …
GATE FAIL: retire-gainers-source — 미완료 태스크 2 건이 남아 있습니다
```

**걸린 것은 8.4와 8.7 둘뿐이고, 다른 것은 없다.** 게이트가 2단계에서 멈추므로 나머지
단계를 손으로 확인했다:

```text
3/8 review.md 존재       OK (이 파일)
4/8 Function Logic Map   OK — evidence complete or diff-proven exempt
5/8 make sdd-check       OK
6/8 make test            OK — 3686 passed, 실패 0
7/8 make vet             OK — No issues found
8/8 make validate        OK — Totals: 25 passed, 0 failed (25 items)
```

`sdd-check`의 `[context-graph] CCG adapter failed`는 WORKFLOW상 관측 전용·비차단이며
exit 0이다(1차 리뷰·§8과 동일한 관측).

**8.7 체크박스는 리뷰어가 채우지 않았다.** 이 리뷰의 수행 기록은 이 절이고, 체크와
랜딩 승인은 Manager의 것이다. 8.4는 애초에 Manager의 커밋 절차다.

---

## 7. 최종 판정

### **§8 수용 — 랜딩 가능 (8.4 커밋 절차만 남음)**

P1 세 건에 대해:

| # | 요구된 것 | 리뷰어 직접 확인 | 판정 |
|---|---|---|---|
| **F1** | 버려지는 오류를 **실제로** 잡는 가드 + 문장 셋 정정 | 변이 A RED, 가드 제거 시 53건 GREEN(결함 실재), 주석·FLM·Branch Test Map 세 곳 전부 사실로 확인 | **닫혔다** |
| **F2** | 실파일 금지 집합의 완전성 단언 | `TOP_GAINERS` 절만 지운 상태에서 이제 `:468` Fatal. 1차 리뷰가 통과를 재현한 그 상태다. 스냅샷은 sha256·`cmp` 바이트 동일 복구 | **닫혔다** |
| **F3** | 커밋 시점 절차 | 미해결 — **의도된 것**이다(§8.4는 Manager 몫). 작업 트리는 양쪽 등록을 유지한 채 그대로다 | **남음(설계대로)** |

그리고 §8이 더한 가드는 **거짓 통과 경로를 하나도 만들지 않는다**: 두 가드의 결합에
false pass가 없음을 리터럴 인식 형태를 열거해 확인했고(§1.7), 공허성 두 방향과 역매핑
충돌까지 직접 만들어 봤다(§1.5·§1.6). §1~§7의 결론 중 무효화된 것은 없다(§5.4).

**랜딩을 막는 발견은 없다.** 자체 발견 세 건(5.1·5.2·5.3)은 전부 **문서·주석**이고
안전 경로에 닿지 않는다. 다만 **5.2는 값이 싸고 성격이 나쁘다** — 그 문장을 믿고
`wiredRankings`의 빈 집합 Fatal을 "중복"이라며 지우면 두 가드가 함께 눈머는 유일한
경로가 열린다. 커밋 전에 한 문장을 고치거나, 그 Fatal 위에 *"새 가드가 이것을 대체하지
않는다"*를 적기를 권한다. 5.1은 I2에 한 줄을 덧붙이는 것으로 족하다.

**남은 것은 8.4(커밋 절차)와 8.7 체크박스뿐이다.**

**한 문장으로**: §8은 자기가 고치겠다고 한 것을 고쳤고, 이번에는 **가드가 지킨다고
말한 것을 실제로 지킨다** — 리뷰어가 지우고 돌려서 확인했다.
