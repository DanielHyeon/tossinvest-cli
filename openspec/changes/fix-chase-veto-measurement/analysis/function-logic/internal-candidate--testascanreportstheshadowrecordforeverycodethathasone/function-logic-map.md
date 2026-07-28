# Function Logic Map: `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne`

- Source: `internal/candidate/watch_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L857–906, 분기 9개)
- Risk scan: `risk-pattern-report.md`

그림자 밴드가 code마다 기록된다는 것. 이 change의 편집은 **마지막 단언의 주석과 문구**다.

전에는 "최초 관측은 후보를 승격한 그 읽기이고 §5가 이제 그 위치를 저장한다"고 적었다.
지금은 그것이 **충분조건이 아니라는 것**까지 말한다 — 소스의 첫 읽기에서 나온 위치는
`NEW_ENTRANT_UNKNOWN`으로 거부되고, 그것이 세션 시작이 패널을 스탬프하는 것을 막는다
(design D3).

이 테스트가 여전히 `Measured == 1`을 얻는 이유는 fixture 소스가 **실제 소스가 보고하는 것**을
보고하기 때문이다: 이 목록의 직전 읽기를 갖고 있었고 도착한 만큼 요청했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixture 소스 | `heldRow`(자격 있는 행) | `wiring_test.go`/`watch_test.go` | 자격이 없으면 밴드가 미측정 |
| 기대 | 가속 crossing 목록, 두 code의 밴드, `seen_late` 밴드 measured=1 | 이 테스트 | `t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Cycle` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `res.Crossings.Total == 0` | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | crossing 목록 길이가 `ShadowThresholds`와 다르다 | 없음 | `t.Errorf` — 아무도 넘지 않은 칸도 있어야 한다 | 자체 실행 |
| B4 | `for _, code := range {seen_late, extended}` | 없음 | — | 자체 실행 |
| B5 | `!ok` — 그 code의 밴드가 없다 | 없음 | `t.Errorf` | 자체 실행 |
| B6 | `tally.Total != 1` | 없음 | `t.Errorf` | 자체 실행 |
| B7 | `for _, n := range tally.NotMeasured` | `sum` 누적 | — | 자체 실행 |
| B8 | `sum != tally.Total` | 없음 | `t.Errorf` — 미측정 합이 전체와 맞아야 한다 | 자체 실행 |
| B9 | `res.Bands[VetoSeenLate].Measured != 1` | 없음 | `t.Errorf` — **문구가 바뀐 단언** | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Cycle(ctx, s, opts)` | 한 turn | `t.Fatal` | ast.json calls |
| `res.Bands[...]` | 밴드 기록 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 주석 6줄 + 실패 메시지 1개. 단언 자체는 무변경.
- High-risk impact: no (테스트 전용).
