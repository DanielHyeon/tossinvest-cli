# Function Logic Map: `previousReading.usableAt`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L114–120, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**다(리뷰 F1). 기억된 직전 읽기가 `at` 시각의 질문에 답할 자격이 있는지를 판정한다.

원래 조건은 **존재**뿐이었고 존재는 장애를 살아남는다. 읽기가 실패하면 두 `rememberRead`
모두 swap 이전에 반환하므로, 429 사다리 안에서 한 시간을 보낸 소스는 한 시간 전의 집합을
그대로 들고 돌아온다. 그 집합에 대해 목록을 떠난 적 없는 심볼은 전부 `no`를 받고, `no`는
최초 관측을 **자격 부여하는** 답이다. 같은 시각 냉각 30분 + staleness 10분이 지났으므로
감시목록 전체가 만료되어 재승격되는 중이다 — 말할 자격이 가장 없는 기억이 패널 전체를
다시 스탬프하는 순간에 발화한다. design D3이 끊으려던 경로가 반대쪽에서 열려 있었다.

상한은 여기서 고르지 않는다. `previousReadingTTL = candidate.DefaultStalenessTTL`이고,
그 값은 `BackoffLadder` 마지막 rung의 2배라는 기존 도출을 그대로 물려받는다.

**2026-07-28**: 그 상수 옆의 두 번째 근거가 틀려서 정정했다(issues.md I18). "그 지점을
넘으면 기억이 자격 부여할 대상이 사라진다"는 거짓이다 — `Store.stateAt`은 staleness에
*냉각*시키고 만료는 40분이며 냉각된 후보는 세 first 칼럼을 그대로 갖는다. 참인 두 문장은
(a) 기억보다 먼저 시작된 생명에 대해서는 `nearFirstSighting`이 이미 양쪽(`NoteFirstRank`·
`MeasureFirstSighting`)에서 거부하고, (b) 그것이 덮지 못하는 **회복 읽기 자신이 시작시키는
생명**이 이 bound가 존재하는 이유라는 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.symbols` | 직전 읽기의 심볼 집합, nil이면 기억 없음 | `rememberRead`의 swap | nil이면 false — 미상 |
| `p.at` | 그 읽기를 기록한 instant | `o.now.Now()` / `w.now.Now()` | zero면 false |
| `at` | 지금 묻는 읽기의 instant | 같은 주입 clock | zero면 false |
| `previousReadingTTL` | `candidate.DefaultStalenessTTL` | `candidatesrc.go:81` | `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore`가 도출을 고정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `p.symbols == nil` 또는 `p.at.IsZero()` 또는 `at.IsZero()` | 없음 — 순수 판정 | `false` (미상) | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants`(nil 기억) · `TestAMemoryWithNoInstantOnEitherSideIsNotAnAnswer`(zero instant 네 case + 정상 대조군) |

무분기 꼬리: `age := at.Sub(p.at)` 후 `age >= 0 && age < previousReadingTTL`.
`age >= 0`이 시계 역행을 거부하고(`TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh`),
`< TTL`이 장애를 거부한다. 경계는 **닫힌 쪽이 아래**다 — 정확히 TTL이면 답하지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `at.Sub` | 기억의 나이 | 순수 산술 | ast.json calls |
| `time.Time.IsZero` | instant 부재 판정 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음. value receiver의 순수 판정이며 `p`를 쓰지 않는다.
- fallback 없음 — 자격이 없으면 호출자가 `(nil, false)`를 받고 그것이 `unknown`이 된다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (읽기 전용 판정). 재는 성질은 High-risk 인접이다 — 이 함수가 항상 true를 돌려주면 장애 뒤 재승격 전체가 `seen_late` 측정 가능이 되고, 그것이 D3이 막는 세션 스탬핑이다.
