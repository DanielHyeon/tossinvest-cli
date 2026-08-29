# Function Logic Map: `TestAVetoedCandidateIsStillStoredAndReported`

- Source: `internal/candidate/veto_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1126–1204, 분기 15개)
- Risk scan: `risk-pattern-report.md`

veto가 발화한 후보도 저장되고 보고된다는 것 — D3의 "막힌 것과 놓친 것이 셀 수 있어야
한다". 이 change의 편집은 **`NoteFirstRank` 호출 1곳의 인자 형태**뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| readings | veto를 발화시키는 관측 | 이 테스트 | — |
| `storedFirstRank(5, 150, ...)` | 자격 있는 최초 위치 | `veto_test.go` | — |
| 기대 | `chase.Vetoed()`, `first_seen_at` 보존, 관측 보존, tally 1건 | 이 테스트 | `t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `RecordObservations` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 | `Promote` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 | `NoteFirstPrice` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 | `NoteFirstRank` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B5 | 후보 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B6 | baseline 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 | 최초 순위 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 | 관측 읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B9 | veto가 발화하지 않았다 | 없음 | `t.Errorf` | 자체 실행 |
| B10 | veto 후 후보 읽기 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B11 | `first_seen_at`이 바뀌었다 | 없음 | `t.Errorf` — veto는 기록을 지우지 않는다 | 자체 실행 |
| B12 | 관측 재읽기 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B13 | 관측이 사라졌다 | 없음 | `t.Errorf` | 자체 실행 |
| B14 | tally의 vetoed가 1이 아니다 | 없음 | `t.Errorf` | 자체 실행 |
| B15 | tally의 total이 1이 아니다 | 없음 | `t.Errorf` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.RecordObservations` / `s.Promote` / `s.NoteFirstPrice` / `s.NoteFirstRank` | 저장 | `t.Fatal` | ast.json calls |
| `AssessChase(...)` / `TallyVetoes(...)` | 판정과 집계 | — | ast.json calls |

## State mutations and fallbacks

- 테스트 전용. `t.TempDir()`의 격리된 저장소와 주입 clock만 쓴다.
- 실계좌·실브로커·네트워크 무접촉. 주문 side effect 0.

## Safety conclusion

- Safe edit boundary: 테스트 수정 — 호출 1곳의 인자 형태.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — veto된 후보가 사라지면 D3의 회고가 불가능해진다.
