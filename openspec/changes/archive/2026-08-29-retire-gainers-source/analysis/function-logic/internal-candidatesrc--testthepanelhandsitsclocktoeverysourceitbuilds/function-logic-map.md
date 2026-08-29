# Function Logic Map: `TestThePanelHandsItsClockToEverySourceItBuilds`

- Source: `internal/candidatesrc/clock_wiring_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> 이 함수는 테스트다. 그런데도 Function Logic Map을 만드는 이유는 `check_analysis.py`가
> **diff hunk와 줄 범위가 겹치는 기존 함수**를 대상으로 잡기 때문이고, 이 change는 이
> 테스트의 진입 단언을 반드시 고쳐야 했다 — 길이 단언 `len(sources) != 4`가 배선 변경과
> 함께 깨진다(D9). 면제를 주장하지 않고 스크립트 출력을 따랐다.
>
> 바뀐 것은 **진입 단언 하나**(B2·B3)이고, 아래 `t.Run` 루프(B4~B10)와 그것이 무엇을
> 검사하는지는 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `clk` (`clock.NewFake(t0())`) | 테스트가 직접 전진시키는 가짜 시계 | `internal/clock` | 시계를 주입하지 않으면 이 테스트는 아무 것도 말하지 못한다 — 실시간에서는 세 번의 Read가 마이크로초 간격이라 경계가 넘어가지 않는다 |
| `rankings` (`aFullRanking(100)`) | 요청 수와 **같은** 행 수 | `OfficialRanking(..., 100, ...)`의 요청 | 짧은 reading은 memory로 채택되지 않으므로(`rememberRead`) 짧게 주면 시계와 무관하게 통과한다. 그래서 "가득 찬" reading이어야 한다 |
| `popular` (`aFullPopularity(30)`) | 같은 이유로 `WTSPopular(..., 30, ...)`와 같은 크기 | 같음 | 같음 |
| `sources`의 id 집합 | KR에서 **정확히** 거래대금·거래량·WTS 인기 | `candidatesrc.Panel` | 어긋나면 B3가 Fatal. 원천이 사라지는 것과 새로 생기는 것 **둘 다** 실패다 |
| `previousReadingTTL` | `candidate.DefaultStalenessTTL` | `candidatesrc.go` | 이 change는 건드리지 않았다 |

**이 change가 바꾼 불변식**: `sources`에 대한 진입 조건이 "개수가 4"에서 "id 집합이 기대
집합과 정확히 같다"로 바뀌었다. **약화가 아니라 강화**다 — 길이 단언은 원천 하나가 다른
하나로 바뀌어도 통과했고, WTS 세션이 없는 구성에서는 무관한 이유로 깨졌다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range sources` (line 82) — 패널이 만든 원천의 id를 모은다 | 지역 map `got`에 기록 | 없음 | 자기 자신 |
| B2 | `len(got) == 0` (line 85) | 없음 | `t.Fatal` | 자기 자신 — 빈 슬라이스 위의 루프는 아무 것도 단언하지 않고 통과한다 |
| B3 | `!sameSourceSet(got, want)` (line 89) | 없음 | `t.Fatalf` | 변이: `Panel`에 `RankingTopGainers` 복원 → RED |
| B4 | `range sources` (line 97) — 원천마다 subtest | `t.Run`, 각 원천의 `Read`가 자기 memory를 갱신 | 없음 | 자기 자신 |
| B5 | 첫 `Read`의 `err != nil` (line 101) | 없음 | `t.Fatalf` | 자기 자신 |
| B6 | 두 번째 `Read`의 `err != nil` (line 107) | 없음 | `t.Fatalf` | 자기 자신 |
| B7 | `len(fresh.Rows) == 0` (line 110) | 없음 | `t.Fatalf` — 행이 없으면 아래 비교가 공허하다 | 자기 자신 |
| B8 | `knownAnswers(fresh.Rows) != len(fresh.Rows)` (line 114) | 없음 | `t.Fatalf` — 평범한 경우가 동작하지 않으면 시계에 대해 말할 수 없다 | 자기 자신 |
| B9 | 세 번째 `Read`의 `err != nil` (line 122) | 없음 | `t.Fatalf` | 자기 자신 |
| B10 | `knownAnswers(stale.Rows) != 0` (line 125) | 없음 | `t.Errorf` — 경계를 넘긴 뒤에도 답한다면 그 원천은 주입된 시계를 쓰지 않는다 | 이 파일이 존재하는 이유 |

**early return 없음**(ast.json `returns`: null). 종료는 `t.Fatal`/`t.Fatalf`뿐이고 전부
단언 실패 경로다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.NewFake`, `clk.Advance` | 시계를 테스트가 소유한다 | 오류 없음 | ast.json `calls` |
| `aFullRanking`, `aFullPopularity` | 요청 수와 같은 크기의 reading을 만든다 | 오류 없음 | ast.json `calls` |
| `Panel` | 검사 대상 배선 | 오류를 돌려주지 않는다 | ast.json `calls` |
| `sameSourceSet`, `sortedIDs` | 이 change가 더한 집합 비교와 안정적 출력. `retiredsource_test.go`에 있다 — `clock_wiring_test.go`의 편집을 고친 함수 하나로 좁게 유지하기 위해서다 | 오류 없음 | ast.json `calls` |
| `src.Read` | 세 번의 reading | 오류를 돌려주면 B5·B6·B9가 잡는다. 네트워크 없음 — `fakeRankings`/`fakePopular` | ast.json `calls` |
| `knownAnswers` | `NewlyListed`가 known인 행 수 | 오류 없음 | ast.json `calls` |

**live config binding**: 없다. 테스트는 격리된 값만 다루고 파일·네트워크·전역 상태에 닿지
않는다.

## State mutations and fallbacks

- mutation은 지역 map `got`와 각 원천 내부의 `seen` memory뿐이다. 후자는 `Read`의 설계된
  side effect이고 이 change가 건드리지 않았다.
- fallback 없음. 이 테스트가 검출하려는 것이 바로 **암묵적 fallback**이다 — 생성자에 nil
  시계를 넘기면 `clock.System()`으로 조용히 되돌아가는 것.
- 이 change가 만든 fallback은 없다.

## Safety conclusion

- Safe edit boundary: 진입 단언(B2·B3)만. `t.Run` 루프와 세 번의 reading, 경계 검사는
  base와 동일하다.
- High-risk impact: **no.** 테스트 파일이고 production 경로에 배선되지 않는다.
- 회귀 위험: 단언이 약해졌는지가 유일한 질문이고, 답은 아니오다. 변이(리터럴에
  `RankingTopGainers` 복원)에서 새 단언이 RED가 되는 것을 확인했다 — 길이 단언을 `3`으로
  고쳤다면 같은 변이에서 GREEN이었을 것이다.
