# Function Logic Map: `TestARankedRowFromTheGapBetweenTwoLivesIsNotThisLifesFirstSighting`

- Source: `internal/candidate/veto_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1496–1571, 분기 15개)
- Risk scan: `risk-pattern-report.md`

두 생명 사이의 gap에 있던 순위 행이 **이 생명의** 최초 관측이 아니라는 것(task 4.9). 이
change의 편집은 **`NoteFirstRank` 호출 2곳의 인자 형태**다.

재는 형태: 첫 생명이 5위에 기록되고, 냉각·만료를 지나며 gap에 148위 행이 남고, 재탄생한
생명이 4위에 기록된다. 죽은 생명의 148위가 새 생명의 최초 관측이 되면 `seen_late`가
clear된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 첫 생명 | 5 of 150 | `storedFirstRank` | 만료 시 `Promote`가 지운다 |
| gap 행 | 148 of 150 | `RecordObservations` | 채택되면 안 된다 |
| 새 생명 | 4 of 150 | `storedFirstRank` | 이것이 최초 관측이어야 한다 |
| 기대 | `Measured`, 4 of 150, `AssessSeenLate` dangerous | 이 테스트 | `t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 첫 생명 관측 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | 첫 승격 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | 냉각 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 | 첫 생명 최초 순위 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | gap 관측 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | 재탄생 관측 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 | 재승격 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 | 새 `first_seen_at`이 재탄생 instant가 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B9 | 새 생명 최초 순위 기록 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B10 | 최초 순위 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B11 | 저장된 위치가 4 of 150이 아니다 | 없음 | `t.Errorf` — 죽은 생명의 5위도 gap의 148위도 아니다 | 자체 실행 |
| B12 | 후보 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B13 | sighting이 미측정이다 | 없음 | `t.Errorf` | 자체 실행 |
| B14 | sighting이 4 of 150이 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B15 | `AssessSeenLate`가 dangerous가 아니다 | 없음 | `t.Errorf` — 4위 진입은 늦게 도착한 것이다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.RecordObservations` / `s.Promote` / `s.Cool` / `s.NoteFirstRank` | 두 생명의 전이 | `t.Fatal` | ast.json calls |
| `MeasureFirstSighting(...)` / `AssessSeenLate(...)` | 측정과 판정 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 2곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — gap 행이 채택되면 `seen_late`가 clear되고, clear된 veto는 통과의 3분의 1이다.
